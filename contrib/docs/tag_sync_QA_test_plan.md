# Tag Sync Job: API and QA Test Plan

The tag sync job brings XDAS (east and west) back in line with Cassandra. This page covers the API, the recommended options and the QA test plan.

## How it works

- The job walks the tag members in Cassandra (the source of truth) and reads each one from XDAS through **xdas**. It only sees one region, east or west: whichever the instance's `group_service.host` points at (the **read region**).
- Writes go through **xdassync** (Kafka) to both regions, after a delay. Every push sets a 1-year TTL.
- Modes:
  - `detect`: counts missing members, writes nothing.
  - `repair`: pushes the members missing in the read region.
  - `refresh`: pushes every member. This resets the TTL and fixes members missing only in the other region, which `detect` and `repair` can't see.
- The job never deletes anything. Only one run can be active at a time.

## Prerequisites

- Create the state table once, in the XConf keyspace:
  ```sql
  CREATE TABLE IF NOT EXISTS "TagSyncState"
      (key text, column1 text, value text, PRIMARY KEY ((key), column1));
  ```
- The `TaggingSyncEnabled` app setting is on (the default).
- In QA, **always pass `tags`**. Without it, a write run pushes every member in the environment.

---

## API

All endpoints take the same SAT token as the other tagging APIs.

| Endpoint | Purpose |
|---|---|
| `POST /taggingService/tags/sync` | Start or resume a run |
| `GET /taggingService/tags/sync/status` | Active run and recent runs, with their counts |
| `POST /taggingService/tags/sync/abort` | Abort the run on this instance |
| `PUT /xconfAdminService/appsettings` | Kill switch (`TaggingSyncEnabled`) |
| `GET /taggingService/tags/{tag}/members` | The tag's members in Cassandra (the expected state) |

### Start a run: `POST /taggingService/tags/sync`

```bash
curl -X POST 'http://<xconf-admin-url>/taggingService/tags/sync' -H 'Authorization: Bearer <SAT token>' -H 'Content-Type: application/json' -d '{"mode": "detect", "tags": ["qa:sync:1"]}'
```

The body is JSON and every field is optional. An empty body runs `detect` over all tags.

| Field | Default | Description |
|---|---|---|
| `mode` | `detect` | `detect`, `repair` or `refresh` |
| `tags` | all tags | Tag ids to walk. Unknown ids are ignored |
| `dryRun` | `false` | Read and count, but don't push. Skipped pushes are counted in `wouldPush` |
| `maxMembers` | no limit | Stop after checking this many members. The run ends `completed` with `limited: true` and can be resumed. Counted per segment. Use 500 or more: a smaller value can stop a resume from moving forward |
| `rate` | 100 | Maximum XDAS calls per second for the whole run. Reads and pushes both count, so `refresh` handles about `rate / 2` members per second |
| `workers` | 20 | Members processed in parallel |
| `chunkSize` | 5000 | Members read from Cassandra per page |
| `resume` | `false` | Continue the newest aborted, crashed or limited run from its checkpoint. `mode` and `tags` come from that run, and sending different ones gets a 400. The other fields are **not** carried over, so send them again |

The `rate`, `workers` and `chunkSize` defaults come from the service config.

Response `202`:
```json
{"runId": "20260925-101500-1a2b3c4d", "mode": "detect", "state": "running", "dryRun": false}
```

| Error | When |
|---|---|
| 400 | Malformed JSON, an unknown `mode`, or a resume with a different `mode` or `tags` |
| 404 | `resume: true`, but there is no run to resume |
| 409 | A run is already active (the body has its `runId`), or the kill switch is off |
| 500 | Run state can't be read or written. Check the `TagSyncState` table and Cassandra |
| 503 | The instance is still starting. Retry |

### Run status: `GET /taggingService/tags/sync/status`

Returns `active` (the running run, or `null`), `history` (the last 10 runs, newest first) and `enabled` (the kill switch state). Main fields of a run:

| Field | Meaning |
|---|---|
| `runId` | Run id. It also appears as `audit_id` in the logs |
| `mode`, `options` | The mode and request options in effect |
| `state` | `running`, `completed` or `aborted` |
| `abortReason` | Why an aborted run stopped (see below) |
| `counts.checked` | Members checked |
| `counts.present` | Members found in XDAS |
| `counts.missingKey` | Members with no XDAS record at all (typically TTL expiry) |
| `counts.missingField` | Members whose XDAS record doesn't have this tag |
| `counts.pushed` / `counts.pushFailed` | Members pushed / members left unpushed after all attempts |
| `counts.wouldPush` | Pushes skipped because of `dryRun` |
| `counts.xdasErrors` | XDAS 5xx errors or timeouts. These are never counted as missing |
| `missingRate` | `(missingKey + missingField) / checked` |
| `tagsDone` / `tagsTotal` | Progress by tag |
| `emptyBuckets` | Buckets listed in Cassandra's bucket metadata but empty when read. They are skipped and logged with their bucket id |
| `checkpoint` | The position a resume continues from |
| `limited` | `true` when the run stopped at `maxMembers` |
| `updatedAt` | Advances about every 30 s while the run is alive |

| `abortReason` | Next step |
|---|---|
| `cancelled` | Aborted through the API or by an instance shutdown. Resume when ready |
| `kill_switch` | Turn `TaggingSyncEnabled` back on, then resume |
| `xdas_unhealthy_consecutive_errors`, `xdas_unhealthy_error_rate` | Too many XDAS errors. Check XDAS, then resume |
| `lock_heartbeat_stale` | The run couldn't refresh its lock in Cassandra. Check Cassandra, then resume |
| `cassandra_error: …` | A Cassandra read failed; the reason includes the driver error. Check Cassandra, then resume |
| `cassandra_suspect_no_tags` | Cassandra returned no tags at all. Check Cassandra, then resume |

### Abort a run: `POST /taggingService/tags/sync/abort`

Stops the run on the instance that receives the call. The run saves its checkpoint and ends `aborted` with `abortReason: cancelled`, so it can be resumed. Returns `202` with the `runId`, or `404` if this instance isn't running one.

### Kill switch: `PUT /xconfAdminService/appsettings`

`{"TaggingSyncEnabled": false}` stops a running job within about 1 min, on any instance, and rejects new runs with 409. Set it back to `true` to allow runs again.

---

## Recommended options

| Goal | Request body |
|---|---|
| Measure drift | `{"mode": "detect"}` |
| Restore expired members (first step) | `{"mode": "repair", "maxMembers": 100000}` |
| Reset TTLs and fix one-region gaps (first step) | `{"mode": "refresh", "maxMembers": 100000}` |
| Next step | `{"resume": true, "mode": "refresh", "maxMembers": 100000}` |
| Rest of the run | `{"resume": true, "mode": "refresh"}` |

- **Measure first.** `detect` is read-only. Prefer it to `dryRun`: a `repair` dry run reports the same counts, and a dry run resumed without `dryRun: true` writes for real.
- **Write in steps.** After each `maxMembers` step, check `status` (`pushFailed`, `xdasErrors`, no `abortReason`) and a few members in both regions before resuming.
- **Always send `mode` when resuming,** plus `tags` if the run used them. A resume continues the newest aborted or limited run. If that isn't the run you mean, the call fails with 400 instead of resuming the wrong one.
- **Keep the default `rate` of 100** unless XDAS and xdassync have headroom. Every push goes through the Kafka queue, so a faster write run also increases the sync lag to east and west. At 100 calls per second, 1M members take about 3 h to `detect` and about 6 h to `refresh`.
- Leave `workers` and `chunkSize` at their defaults.
- When a run finishes, wait for the sync lag and run `detect` again. Missing should be close to 0, apart from `pushFailed`.

---

## Checking both regions

The job's writes reach east and west through xdassync, the same path every normal tag write takes. Its reads only cover the read region.

- **Read region:** `detect` checks every member, so no manual checks are needed, whatever the tag size.
- **Both regions:** after a write run, wait for the xdassync lag (poll for up to about 10 min). Then check a few known members directly in the east and west xdas. A small sample is enough; checking every member of a large tag would mostly re-test xdassync.
- **Optional, whole tag in the other region:** temporarily point `group_service.host` at the other region's xdas (config change and restart), run `detect`, then switch back.

## Test data

Create tag `qa:sync:1` with 15 test MACs. Then delete XDAS records as shown below. xconfadmin always writes to both stores, so this needs the XDAS team or Redis access.

| Group | XDAS record deleted in |
|---|---|
| G0 (5) | none |
| G1 (5) | both regions (simulates TTL expiry) |
| G2 (5) | one region only, east or west |

## Test cases

Run in order. Tests 1–5 use `"tags": ["qa:sync:1"]` in every request body.

| # | Steps | Expected |
|---|---|---|
| 1 | `detect` | `missingKey` is 5 (G1), or 10 if G2 was deleted in the read region. Nothing is written |
| 2 | `repair` with `dryRun: true` | `wouldPush` equals the missing count from step 1. Nothing is written |
| 3 | `repair`, wait for the lag, then `detect` | `pushed` equals the missing count from step 1, and the `detect` shows 0 missing. G1 is back in both regions. G2 is back only if it was deleted in the read region |
| 4 | `refresh`, then wait for the lag | `pushed` is 15. G2 is present in both east and west |
| 5 | With a firmware rule on `qa:sync:1` EXISTS, call `/xconf/swu/stb` for a G1 MAC before and after step 3 | Before: no match. After: match |
| 6 | Large tag: `refresh`, wait for the lag, then `detect` | The run ends `completed`, `checked` equals the tag's member count and `pushFailed` is 0. The `detect` shows 0 missing. A sample of about 20 members is present in both regions |
| 7 | Start a slow run (`"rate": 1`), `abort` it, then send `{"resume": true}` | The run ends `aborted`. The resume keeps the same `runId`, continues and ends `completed` |
| 8 | Trigger a run while one is active | 409 with the active `runId` |

## Notes

- Missing members are pushed with a blank value, because Cassandra doesn't store tag values.
- A `completed` run with `pushFailed > 0` still left some drift. Run the same mode again.
- If XDAS answers "not found" for everything (an outage that looks like mass expiry), a write run pushes the whole scope. Check that the `detect` numbers look plausible before any write run.
