# XConf Tagging Service API

## Table of Contents

1. [Performance Information](#performance-information)
   - [Add Members API](#add-members-api)
   - [Delete Members API](#delete-members-api)
2. [API Endpoints](#api-endpoints)
   - [SAT Token Requirements](#sat-token-requirements)
   - [Get Tag by ID](#get-tag-by-id)
   - [Delete Tag by ID (Asynchronous)](#delete-tag-by-id-asynchronous)
   - [Add Members to Tag](#add-members-to-tag)
   - [Remove Members from Tag](#remove-members-from-tag)
   - [Remove Member from Tag](#remove-member-from-tag)
   - [Get Tag Members](#get-tag-members)
3. [Tag Sync Job](#tag-sync-job)
   - [Trigger Tag Sync](#trigger-tag-sync)
   - [Tag Sync Status](#tag-sync-status)
   - [Abort Tag Sync](#abort-tag-sync)
   - [Kill Switch](#kill-switch)
   - [Abort Reasons](#abort-reasons)
   - [Configuration Defaults](#configuration-defaults)
4. [XConf Rule Configuration with Tags](#xconf-rule-configuration-with-tags)

---

## Performance Information

### Add Members API

The Add Members API processes member additions in batches, with each batch supporting up to 2,000 members.

### Delete Members API

The Delete Members API also operates in batches of up to 2,000 members per batch.

---

## API Endpoints

### SAT Token Requirements

Client should have following SAT capabilities:
- `"x1:coast:cmtagds:assign"`
- `"x1:coast:cmtagds:read"`
- `"x1:coast:cmtagds:unassign"`
- `"x1:coast:xconf:read"`
- `"x1:coast:xconf:read:maclist"`
- `"x1:coast:xconf:write"`
- `"x1:coast:xconf:write:maclist"`

---

### Get Tag by ID

Returns representation of XConf tag by provided tag id

**Endpoint:**
```
GET /taggingService/tags/{id}
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

**Response Status Codes:**
- `200 OK`
- `404 NOT FOUND`

**Response Body:**
```json
{
    "id": "test:tag:demotag",
    "description": "",
    "members": [
        "A2:A2:A2:A2:B2:B2"
    ],
    "updated": 1711651165855
}
```

---

### Delete Tag by ID (Asynchronous)

Deletes a tag and all its members asynchronously. The API returns immediately after validation, and the actual deletion is processed in the background.

**Endpoint:**
```
DELETE /taggingService/tags/{id}
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

**Success Response (202 Accepted):**
The tag deletion request has been accepted and queued for processing.

**Status Code:** `202 Accepted`

**Response Body:**
```json
{
    "status": "accepted",
    "message": "Tag 'my-tag' deletion has been queued for processing",
    "tag": "my-tag"
}
```

#### Behavior

- **Immediate Response:** API returns 202 Accepted immediately after validating that the tag exists
- **Background Processing:** Tag deletion (including all members and buckets) happens asynchronously
- **No Status Tracking:** Currently no endpoint to check deletion progress (work is pending)
- **Error Handling:** Any errors during background deletion are logged server-side

#### Notes

- The 202 Accepted status indicates the request was valid and accepted, not that deletion is complete
- For large tags with many members, deletion may take several minutes
- Once accepted, the deletion cannot be cancelled

---

### Add Members to Tag

Adds new members to the tag. If tag does not exist – new tag is created in XConf. By default XConf does tag member normalization: whitespaces are trimmed, string data is set to upper case.

**Endpoint:**
```
PUT /taggingService/tags/{tag}/members
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

**Request Body - list of members:**
```json
["A1:A1:A1:A1:B1:B1", "A2:A2:A2:A2:B2:B2"]
```

**Response Status Code:** `202 Accepted`

**Response Body - XConf tag entity with added members:**
```json
{
    "id": "test:tag:demotag",
    "description": "",
    "members": [
        "A1:A1:A1:A1:B1:B1",
        "A2:A2:A2:A2:B2:B2"
    ],
    "updated": 1711651165855
}
```

---

### Remove Members from Tag

Removes members from the tag. If all members are removed, the tag is automatically deleted.

**Endpoint:**
```
DELETE /taggingService/tags/{tag}/members
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

**Request Body - list of members:**
```json
["A1:A1:A1:A1:B1:B1", "A2:A2:A2:A2:B2:B2"]
```

**Response Status Codes:**
- `404 NOT FOUND`
- `204 NO CONTENT`

---

### Remove Member from Tag

Removes member record from XDAS first, in case of success removes tag member from XConf. Remove API takes non-normalized data, normalization is done by XConf.

**Endpoint:**
```
DELETE /taggingService/tags/{tag}/members/{member}
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

**Response Status Code:** `204 NO CONTENT`

---

### Get Tag Members

Retrieves all members of a specified tag. Supports both non-paginated (V1 compatible) and paginated responses.

**Endpoint:**
```
GET /taggingService/tags/{tag}/members
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

#### Query Parameters (Optional - for pagination)

| Parameter | Type | Required | Default | Maximum | Description |
|-----------|------|----------|---------|---------|-------------|
| `limit` | integer | No | 500 | 5000 | Number of members to return per page. Must be a positive integer. If exceeds maximum, returns 400 Bad Request |
| `cursor` | string | No | - | - | Pagination cursor for retrieving the next page of results. Obtained from nextCursor field in the previous response |

**Note:** If either limit or cursor is provided, the endpoint returns a paginated response. Otherwise, it returns a non-paginated response.

#### Response Status Codes

- `200 OK`: Successfully retrieved members
- `206 Partial Content`: Response contains only first 100,000 members (tag has more than 100k members)
- `400 Bad Request`: Invalid tag or query parameters
- `404 Not Found`: Tag does not exist

#### Non-Paginated Response Body

```json
[
    "A2:A2:A2:A2:B2:B2"
]
```

**Important:** In non-paginated mode, if a tag has more than 100,000 members, the response will be truncated to the first 100,000 members and the status code will be 206 Partial Content. To retrieve all members of large tags, use paginated mode.

#### Paginated Mode

Used when limit and/or cursor query parameters are provided.

**Response Status Codes:**
- `200 OK`: Successfully retrieved page of members
- `400 Bad Request`: Invalid query parameters or tag parameter
- `404 Not Found`: Tag does not exist

**Response Body:**
```json
{
    "data": [
        "A2:A2:A2:A2:B2:B2",
        "C3:C3:C3:C3:D3:D3",
        "E4:E4:E4:E4:F4:F4"
    ],
    "nextCursor": "eyJidWNrZXQiOjEyLCJsYXN0S2V5IjoiQTI6QTI6QTI6QTI6QjI6QjIifQ==",
    "hasMore": true
}
```

**Response Fields:**
- `data` (array of strings): List of member identifiers in the current page
- `nextCursor` (string, optional): Cursor for the next page. Omitted if there are no more results.
- `hasMore` (boolean): Indicates whether more results are available
  - `true`: More members available, use nextCursor to retrieve next page
  - `false`: No more members to retrieve

##### Example Requests

**Non-Paginated (all members, up to 100k):**
```
GET /taggingService/tags/my-tag-123/members
```

**Paginated (first page with custom limit):**
```
GET /taggingService/tags/my-tag-123/members?limit=1000
```

**Paginated (subsequent page):**
```
GET /taggingService/tags/my-tag-123/members?limit=1000&cursor=eyJidWNrZXQiOjEyLCJsYXN0S2V5IjoiQTI6QTI6QTI6QTI6QjI6QjIifQ==
```

##### Pagination Workflow

1. Make initial request with optional limit parameter
2. Process the data array containing members
3. Check hasMore field:
   - If `true`: Use the nextCursor value as the cursor parameter in the next request
   - If `false`: All members have been retrieved
4. Repeat until hasMore is false

---

## Tag Sync Job

The tag sync job walks Cassandra (the source of truth for tag membership) and brings XDAS in line
with it. It exists because XDAS entries carry a server-side TTL: members on inactive devices expire
out of XDAS while Cassandra still holds them, and rule evaluation stops seeing their tags.

The job has three modes sharing the same walk:

| Mode | Writes | What it does |
|------|--------|--------------|
| `detect` | none | Read-only census: classifies every member as present / missing in XDAS and reports the missing-member numbers |
| `repair` | pushes missing members only | Restores members that expired out of XDAS |
| `refresh` | re-pushes every member | Every push carries the TTL header, so this resets XDAS TTLs for the whole population, including inactive devices. Present members are re-pushed with their **observed** XDAS value (values are never overwritten with blanks) |

The job is **additive-only**: it never deletes anything from either store. XDAS server errors (5xx,
transport failures) are never counted as missing — only a clean "not found" is. A circuit breaker
aborts the run if XDAS looks unhealthy, and a suspiciously high missing rate must be confirmed by a
probe (a known-good member that still reads back) before the run continues.

Only **one run** can be active across the whole cluster at a time (a Cassandra lock with heartbeat
enforces this). Progress is checkpointed continuously, so an aborted or crashed run can be resumed
without re-walking what was already covered.

---

### Trigger Tag Sync

Starts a run in the background and returns immediately with the run id.

**Endpoint:**
```
POST /taggingService/tags/sync
```

**Headers:**
```
Accept = application/json
Content-Type = application/json
Authorization = Bearer {SAT token}
```

Requires the write capability (`x1:coast:xconf:write` or `x1:appds:xconf:*`).

**Request Body (JSON, all fields optional — an empty body runs `detect` with defaults):**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mode` | string | `"detect"` | `detect`, `repair` or `refresh` (see table above) |
| `tags` | array of strings | all tags | Restrict the walk to these tag ids; unknown ids are ignored |
| `rate` | integer | config (100) | Maximum XDAS calls per second for the entire run. Both reads and pushes take a slot, so in `refresh` mode the effective member throughput is about `rate / 2` |
| `workers` | integer | config (20) | Concurrent workers processing members within a chunk |
| `chunkSize` | integer | config (5000) | Members fetched per Cassandra page; also the checkpoint granularity |
| `dryRun` | boolean | `false` | Classify and count, but never push; pushes that would have happened are reported as `wouldPush` |
| `maxMembers` | integer | unlimited | Stop cleanly after checking this many members. The run finishes as `completed` with `limited: true` and **can be resumed** — use this to ramp up (e.g. 100k first, review, then resume) |
| `resume` | boolean | `false` | Continue the most recent resumable run (aborted, crashed, or completed-limited) from its checkpoint. `mode` and `tags` are taken from the resumed run; `rate`, `workers`, `chunkSize`, `dryRun` and `maxMembers` may be set anew — note they do **not** inherit from the original run: omitted values fall back to the config defaults (so restate `dryRun: true` when resuming a dry run, or the continuation pushes for real). `probeMember` is the exception: it carries over from the original run unless overridden |
| `probeMember` | string | none | A known-good member (one that must currently be present in XDAS). Used to tell genuine mass expiry from an XDAS outage that answers "not found" for everything, and as a preflight check in write modes. Passed per run deliberately — a probe pinned in config would itself rot away via TTL expiry |

**Example — read-only census over everything:**
```bash
curl --location --request POST 'http://<xconf-admin-url>/taggingService/tags/sync' \
  --header 'Authorization: Bearer <SAT token>' \
  --header 'Content-Type: application/json' \
  --data '{"mode": "detect", "probeMember": "AA:BB:CC:DD:EE:FF"}'
```

**Example — ramped repair of two tags, dry run first:**
```bash
curl --location --request POST 'http://<xconf-admin-url>/taggingService/tags/sync' \
  --header 'Authorization: Bearer <SAT token>' \
  --header 'Content-Type: application/json' \
  --data '{"mode": "repair", "tags": ["tag-a", "tag-b"], "dryRun": true, "maxMembers": 100000, "rate": 300}'
```

**Example — resume the previous run:**
```bash
curl --location --request POST 'http://<xconf-admin-url>/taggingService/tags/sync' \
  --header 'Authorization: Bearer <SAT token>' \
  --header 'Content-Type: application/json' \
  --data '{"resume": true}'
```

**Response Status Codes:**
- `202 Accepted`: run started in the background
- `400 Bad Request`: malformed body or invalid `mode`
- `403 Forbidden`: token lacks the tools write capability
- `409 Conflict`: a run is already active (response includes its `runId` and `owner`), or the kill switch is off

**Response Body (202):**
```json
{
    "runId": "20260817-153012-1a2b3c4d",
    "mode": "detect",
    "state": "running",
    "dryRun": false
}
```

---

### Tag Sync Status

Reports the currently active run (on any instance) and recent run history. Read-level access.

**Endpoint:**
```
GET /taggingService/tags/sync/status
```

**Response Status Code:** `200 OK`

**Response Body:**
```json
{
    "active": { "runId": "20260817-153012-1a2b3c4d", "state": "running", "...": "..." },
    "history": [ { "runId": "...", "state": "completed", "...": "..." } ],
    "enabled": true
}
```

- `active` — the run currently holding the cluster lock, or `null`
- `history` — up to 10 most recent runs, newest first (older records are pruned automatically)
- `enabled` — current kill switch state

**Run record fields:**

| Field | Description |
|-------|-------------|
| `runId` | Time-prefixed unique id; also the `audit_id` on every log line of the run |
| `mode`, `options`, `owner` | What was requested and which instance is running it |
| `state` | `running`, `completed` or `aborted` |
| `startedAt`, `updatedAt`, `completedAt` | Timestamps; `updatedAt` advances with every checkpoint save |
| `checkpoint` | Walk position (`tagId`, `bucketId`, `lastMember`) — where a resume would continue |
| `counts.checked` | Members checked against XDAS so far |
| `counts.present` | Members whose tag field was found in XDAS |
| `counts.missingField` | XDAS record exists but the tag field is gone |
| `counts.missingKey` | No XDAS record for the member at all (typically TTL expiry) |
| `counts.pushed` / `counts.pushFailed` | Write-mode push outcomes |
| `counts.wouldPush` | Pushes suppressed by `dryRun` |
| `counts.xdasErrors` | 5xx/transport errors — never counted as missing |
| `counts.xdasOnlyFieldsSeen` | Distinct XDAS tag fields with no Cassandra counterpart (reported only, never deleted) |
| `tagsTotal`, `tagsDone`, `tagsWithMissing` | Walk progress by tag |
| `topMissingTags` | Up to 100 tags with the most missing members, with per-tag checked/missing/pushed counts; refreshed on every checkpoint save |
| `missingRate` | `(missingField + missingKey) / checked` |
| `abortReason` | Why an aborted run stopped (see [Abort Reasons](#abort-reasons)) |
| `limited` | Run stopped at `maxMembers`; resumable |
| `resumes` | How many times this record has been resumed |
| `missingRateUnconfirmed` | A `detect` census crossed the missing-rate threshold with no probe available; the numbers need manual confirmation |

---

### Abort Tag Sync

Cancels the run owned by **this instance**. The run checkpoints and finishes as `aborted`, so it can
be resumed later.

**Endpoint:**
```
POST /taggingService/tags/sync/abort
```

Requires the write capability (`x1:coast:xconf:write` or `x1:appds:xconf:*`).

**Response Status Codes:**
- `202 Accepted`: abort requested; body contains the `runId`
- `403 Forbidden`: token lacks the tools write capability
- `404 Not Found`: no active run on the instance that received the request — to stop a run on
  another instance, use the [kill switch](#kill-switch)

---

### Kill Switch

The `TaggingSyncEnabled` app setting gates the whole job cluster-wide:

```bash
curl --location --request PUT 'http://<xconf-admin-url>/xconfAdminService/appsettings' \
  --header 'Authorization: Bearer <SAT token>' \
  --header 'Content-Type: application/json' \
  --data '{"TaggingSyncEnabled": false}'
```

- `false` rejects new triggers (`409`) and makes any running run abort at its next chunk boundary
  (`abortReason: "kill_switch"`), on whichever instance it runs — takes effect within about a minute
  (app settings cache refresh)
- The setting is seeded as `true` at startup; absent or unreadable also means enabled
- Both JSON `false` and the string `"false"` are honored

---

### Abort Reasons

| `abortReason` | Meaning | What to do |
|---------------|---------|------------|
| `cancelled` | Abort endpoint or shutdown | Resume when ready |
| `kill_switch` | `TaggingSyncEnabled` was set to false | Re-enable, then resume |
| `xdas_unhealthy_consecutive_errors` | Too many XDAS errors in a row | Check XDAS health, then resume |
| `xdas_unhealthy_error_rate` | XDAS error rate over the window threshold | Check XDAS health, then resume |
| `missing_rate_high_probe_failed` | Missing rate crossed the threshold and the probe (or recently-present members) could not be read back — looks like an XDAS outage, not genuine expiry | Verify XDAS; do not trust the run's missing counts |
| `missing_rate_high_no_probe_available` | Write mode crossed the missing-rate threshold with no probe to confirm | Re-trigger with a `probeMember` |
| `probe_member_not_readable` | Write-mode preflight: the supplied probe is not present in XDAS | Pick a probe device that is verifiably in XDAS |
| `cassandra_suspect_no_tags` | The tag census came back empty — indistinguishable from a Cassandra failure | Check Cassandra; re-trigger |
| `cassandra_suspect_empty_bucket` | A bucket reported as populated returned no members — suspected swallowed Cassandra error (or a concurrent tag deletion) | Resume; it self-heals if the tag was genuinely deleted |
| `cassandra_error: ...` | Explicit Cassandra error | Check Cassandra, then resume |

---

### Configuration Defaults

Server-side defaults for the trigger options and the safety guards, set in the service config
(see `config/sample_xconfadmin.conf`):

| Key | Default | Description |
|-----|---------|-------------|
| `tag_sync_rate_limit` | 100 | XDAS calls/sec cap when the trigger does not pass `rate` |
| `tag_sync_worker_count` | 20 | Worker pool size when the trigger does not pass `workers` |
| `tag_sync_chunk_size` | 5000 | Cassandra page size when the trigger does not pass `chunkSize` |
| `tag_sync_checkpoint_interval_secs` | 30 | How often run progress is persisted |
| `tag_sync_breaker_window` | 200 | Sliding window (members) for the error/missing rates |
| `tag_sync_breaker_min_sample` | 500 | Members that must be seen before the rate thresholds arm |
| `tag_sync_breaker_error_rate_percent` | 25 | Error rate over the window that aborts the run |
| `tag_sync_breaker_missing_rate_percent` | 90 | Missing rate that triggers probe confirmation |
| `tag_sync_breaker_max_consec_errors` | 10 | Consecutive XDAS errors that abort the run |

---

## XConf Rule Configuration with Tags

### Steps to Configure Rules with Tags

1. **Create New Firmware Rule with the tag as the condition using EXISTS operation.**

2. **Add needed MAC address or any other parameters to the tag using "Add member to tag" API:**

```bash
curl --location --request PUT 'http://<xconf-admin-url>/taggingService/tags/xconf:tag:usage:demo/members' \
  --header 'Authorization: Bearer <SAT token>' \
  --header 'Accept: application/json' \
  --header 'Content-Type: application/json' \
  --data '["BB:BB:BB:BB:BB:BB"]'
```

3. **Trigger /swu/xconf/ API to evaluate the rules, make sure that tag member from step 2 is present as in the request parameters of /swu/xconf/ query:**

```bash
curl --location 'http://<xconf-url>/xconf/swu/stb?model=TESTMODEL&eStbMac=BB%3ABB%3ABB%3ABB%3ABB%3ABB&firmwareVersion=TEST_VERSION'
```

**Example Response:**
```json
{
    "firmwareDownloadProtocol": "tftp",
    "firmwareFilename": "filename.t",
    "firmwareVersion": "TEST_VERSION_TAGGING_USAGE",
    "mandatoryUpdate": false,
    "rebootImmediately": false
}
```

---

*Document Version: 1.1*
*Last Updated: August 2026*
