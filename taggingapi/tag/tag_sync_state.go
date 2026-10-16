package tag

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	ds "github.com/rdkcentral/xconfwebconfig/db"
)

// Tag sync job state lives in its own Cassandra table in the xconf keyspace
// (raw CQL access, same pattern as the tagging tables):
//
//	CREATE TABLE IF NOT EXISTS tag_sync_state
//	    (tenant_id text, key text, column1 text, value text,
//	     PRIMARY KEY ((tenant_id, key), column1));
//
// Every row lives in its tenant's partitions: key='run', column1=<runId>
// holds one JSON record per run (the run id is time-prefixed so rows cluster
// chronologically); key='control', column1='lock' holds the tenant's
// single-run lock with its heartbeat.
const (
	QueryTagSyncStateUpsert = `INSERT INTO tag_sync_state (tenant_id, key, column1, value) VALUES (?, ?, ?, ?)`
	QueryTagSyncStateGet    = `SELECT value FROM tag_sync_state WHERE tenant_id = ? AND key = ? AND column1 = ?`
	QueryTagSyncStateList   = `SELECT column1, value FROM tag_sync_state WHERE tenant_id = ? AND key = ?`

	QueryTagSyncStateListNewest = `SELECT column1, value FROM tag_sync_state WHERE tenant_id = ? AND key = ? ORDER BY column1 DESC LIMIT ?`
	QueryTagSyncStateDelete     = `DELETE FROM tag_sync_state WHERE tenant_id = ? AND key = ? AND column1 = ?`

	tagSyncRunKey     = "run"
	tagSyncControlKey = "control"
	tagSyncLockColumn = "lock"
)

// TagSyncLock is the cross-instance single-run guard. A lock is live while
// Released is false and the heartbeat is fresher than tagSyncLockStaleAfter;
// a crashed pod's lock goes stale on its own and can be taken over.
type TagSyncLock struct {
	Owner       string    `json:"owner"`
	RunId       string    `json:"runId"`
	HeartbeatAt time.Time `json:"heartbeatAt"`
	Released    bool      `json:"released"`
}

type tagSyncDao interface {
	saveRun(run *TagSyncRun) error
	getRun(runId string) (*TagSyncRun, error)
	// listRuns returns the newest limit runs, newest first.
	listRuns(limit int) ([]*TagSyncRun, error)
	// pruneRuns deletes run rows beyond the newest keep, so the history
	// partition the status endpoint scans stays bounded.
	pruneRuns(keep int) error
	// getLock returns nil when no lock row exists yet.
	getLock() (*TagSyncLock, error)
	saveLock(lock *TagSyncLock) error
}

// tagSyncDaoImpl reads and writes the runs and lock of one tenant.
type tagSyncDaoImpl struct {
	tenantId string
}

func newTagSyncDao(tenantId string) tagSyncDao {
	return tagSyncDaoImpl{tenantId: tenantId}
}

func (d tagSyncDaoImpl) saveRun(run *TagSyncRun) error {
	data, err := json.Marshal(run)
	if err != nil {
		return fmt.Errorf("tag sync run marshal failed: %w", err)
	}
	return ds.GetSimpleDao().Modify(QueryTagSyncStateUpsert, d.tenantId, tagSyncRunKey, run.RunId, string(data))
}

func (d tagSyncDaoImpl) getRun(runId string) (*TagSyncRun, error) {
	rows, err := queryRows(QueryTagSyncStateGet, d.tenantId, tagSyncRunKey, runId)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return unmarshalTagSyncRunRow(rows[0])
}

func (d tagSyncDaoImpl) listRuns(limit int) ([]*TagSyncRun, error) {
	rows, err := queryRows(QueryTagSyncStateListNewest, d.tenantId, tagSyncRunKey, strconv.Itoa(limit))
	if err != nil {
		return nil, err
	}
	runs := make([]*TagSyncRun, 0, len(rows))
	for _, row := range rows {
		run, err := unmarshalTagSyncRunRow(row)
		if err != nil || run == nil {
			continue
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// pruneRuns works on raw column1 ids (not unmarshalled records) so corrupt
// rows are pruned too instead of surviving forever.
func (d tagSyncDaoImpl) pruneRuns(keep int) error {
	rows, err := queryRows(QueryTagSyncStateList, d.tenantId, tagSyncRunKey)
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if id, ok := row["column1"].(string); ok && id != "" {
			ids = append(ids, id)
		}
	}
	if keep < 0 {
		keep = 0
	}
	if len(ids) <= keep {
		return nil
	}
	// Run ids are time-prefixed: ascending lexical order is oldest first.
	sort.Strings(ids)
	for _, id := range ids[:len(ids)-keep] {
		if err := ds.GetSimpleDao().Modify(QueryTagSyncStateDelete, d.tenantId, tagSyncRunKey, id); err != nil {
			return err
		}
	}
	return nil
}

func (d tagSyncDaoImpl) getLock() (*TagSyncLock, error) {
	rows, err := queryRows(QueryTagSyncStateGet, d.tenantId, tagSyncControlKey, tagSyncLockColumn)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	value, ok := rowJsonValue(rows[0])
	if !ok {
		return nil, nil
	}
	var lock TagSyncLock
	if err := json.Unmarshal([]byte(value), &lock); err != nil {
		return nil, fmt.Errorf("tag sync lock unmarshal failed: %w", err)
	}
	return &lock, nil
}

func (d tagSyncDaoImpl) saveLock(lock *TagSyncLock) error {
	data, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("tag sync lock marshal failed: %w", err)
	}
	return ds.GetSimpleDao().Modify(QueryTagSyncStateUpsert, d.tenantId, tagSyncControlKey, tagSyncLockColumn, string(data))
}

func unmarshalTagSyncRunRow(row map[string]interface{}) (*TagSyncRun, error) {
	value, ok := rowJsonValue(row)
	if !ok {
		return nil, nil
	}
	var run TagSyncRun
	if err := json.Unmarshal([]byte(value), &run); err != nil {
		return nil, fmt.Errorf("tag sync run unmarshal failed: %w", err)
	}
	return &run, nil
}

// rowJsonValue extracts the JSON payload column shared by every tag_sync_state
// row shape (runs and the lock).
func rowJsonValue(row map[string]interface{}) (string, bool) {
	value, ok := row["value"].(string)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}
