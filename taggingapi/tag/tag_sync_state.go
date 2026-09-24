package tag

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	ds "github.com/rdkcentral/xconfwebconfig/db"
)

// Tag sync job state lives in its own Cassandra table (raw CQL access, same
// pattern as the tagging tables; no DAO registration needed):
//
//	CREATE TABLE IF NOT EXISTS tag_sync_state
//	    (tenant_id text, key text, column1 text, value text,
//	     PRIMARY KEY ((tenant_id, key), column1));
//
// Run rows are partitioned by tenant. The lock uses a reserved tenant so it
// remains global: XDAS and the configured call budget are shared by tenants.
const (
	QueryTagSyncStateUpsert = `INSERT INTO tag_sync_state (tenant_id, key, column1, value) VALUES (?, ?, ?, ?)`
	QueryTagSyncStateGet    = `SELECT value FROM tag_sync_state WHERE tenant_id = ? AND key = ? AND column1 = ?`
	QueryTagSyncStateList   = `SELECT column1, value FROM tag_sync_state WHERE tenant_id = ? AND key = ?`

	QueryTagSyncStateListNewest = `SELECT column1, value FROM tag_sync_state WHERE tenant_id = ? AND key = ? ORDER BY column1 DESC LIMIT ?`
	QueryTagSyncStateDelete     = `DELETE FROM tag_sync_state WHERE tenant_id = ? AND key = ? AND column1 = ?`

	tagSyncGlobalTenantId = "__global__"
	tagSyncRunKey         = "run"
	tagSyncControlKey     = "control"
	tagSyncLockColumn     = "lock"
)

// TagSyncLock is the cross-instance single-run guard. A lock is live while
// Released is false and the heartbeat is fresher than tagSyncLockStaleAfter;
// a crashed pod's lock goes stale on its own and can be taken over.
type TagSyncLock struct {
	TenantId    string    `json:"tenantId"`
	Owner       string    `json:"owner"`
	RunId       string    `json:"runId"`
	HeartbeatAt time.Time `json:"heartbeatAt"`
	Released    bool      `json:"released"`
}

type tagSyncDao interface {
	saveRun(tenantId string, run *TagSyncRun) error
	getRun(tenantId string, runId string) (*TagSyncRun, error)
	// listRuns returns the newest limit runs, newest first.
	listRuns(tenantId string, limit int) ([]*TagSyncRun, error)
	// pruneRuns deletes run rows beyond the newest keep, so the history
	// partition the status endpoint scans stays bounded.
	pruneRuns(tenantId string, keep int) error
	// getLock returns nil when no lock row exists yet.
	getLock() (*TagSyncLock, error)
	saveLock(lock *TagSyncLock) error
}

type tagSyncDaoImpl struct{}

func newTagSyncDao() tagSyncDao {
	return tagSyncDaoImpl{}
}

func (tagSyncDaoImpl) saveRun(tenantId string, run *TagSyncRun) error {
	data, err := json.Marshal(run)
	if err != nil {
		return fmt.Errorf("tag sync run marshal failed: %w", err)
	}
	return ds.GetSimpleDao().Modify(QueryTagSyncStateUpsert, tenantId, tagSyncRunKey, run.RunId, string(data))
}

func (tagSyncDaoImpl) getRun(tenantId string, runId string) (*TagSyncRun, error) {
	rows, err := ds.GetSimpleDao().Query(QueryTagSyncStateGet, tenantId, tagSyncRunKey, runId)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return unmarshalTagSyncRunRow(rows[0])
}

func (tagSyncDaoImpl) listRuns(tenantId string, limit int) ([]*TagSyncRun, error) {
	rows, err := ds.GetSimpleDao().Query(QueryTagSyncStateListNewest, tenantId, tagSyncRunKey, strconv.Itoa(limit))
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
func (tagSyncDaoImpl) pruneRuns(tenantId string, keep int) error {
	rows, err := ds.GetSimpleDao().Query(QueryTagSyncStateList, tenantId, tagSyncRunKey)
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
		if err := ds.GetSimpleDao().Modify(QueryTagSyncStateDelete, tenantId, tagSyncRunKey, id); err != nil {
			return err
		}
	}
	return nil
}

func (tagSyncDaoImpl) getLock() (*TagSyncLock, error) {
	rows, err := ds.GetSimpleDao().Query(QueryTagSyncStateGet, tagSyncGlobalTenantId, tagSyncControlKey, tagSyncLockColumn)
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

func (tagSyncDaoImpl) saveLock(lock *TagSyncLock) error {
	data, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("tag sync lock marshal failed: %w", err)
	}
	return ds.GetSimpleDao().Modify(QueryTagSyncStateUpsert, tagSyncGlobalTenantId, tagSyncControlKey, tagSyncLockColumn, string(data))
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
