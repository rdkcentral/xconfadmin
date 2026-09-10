package tag

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/rdkcentral/xconfadmin/common"
	taggingapi_config "github.com/rdkcentral/xconfadmin/taggingapi/config"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// mockDbClient injects Cassandra query behavior for service-level tests. The
// embedded interface panics on any unstubbed method, flagging unexpected
// database usage; batch and modify operations succeed by default.
type mockDbClient struct {
	db.DatabaseClient
	queryFunc   func(query string, params ...string) ([]map[string]any, error)
	modifyFunc  func(query string, params ...string) error
	execBatchFn func(batch *mockBatch) error
}

type mockBatch struct {
	statements []string
	// Captured alongside statements so tests can assert on the values written,
	// not merely that a statement was issued.
	args [][]any
}

func (b *mockBatch) Query(stmt string, args ...any) {
	b.statements = append(b.statements, stmt)
	b.args = append(b.args, args)
}

func (b *mockBatch) Size() int {
	return len(b.statements)
}

// argsFor returns the arguments of the first occurrence of a statement.
func (b *mockBatch) argsFor(stmt string) []any {
	for i, s := range b.statements {
		if s == stmt {
			return b.args[i]
		}
	}
	return nil
}

func (m *mockDbClient) QueryXconfDataRows(query string, params ...string) ([]map[string]any, error) {
	return m.queryFunc(query, params...)
}

func (m *mockDbClient) ModifyXconfData(query string, params ...string) error {
	if m.modifyFunc != nil {
		return m.modifyFunc(query, params...)
	}
	return nil
}

func (m *mockDbClient) NewBatch(batchType int) db.BatchOperation {
	return &mockBatch{}
}

func (m *mockDbClient) ExecuteBatch(batch db.BatchOperation) error {
	if m.execBatchFn != nil {
		return m.execBatchFn(batch.(*mockBatch))
	}
	return nil
}

func withMockDbClient(t *testing.T, client *mockDbClient) {
	t.Helper()
	old := db.GetDatabaseClient()
	db.SetDatabaseClient(client)
	t.Cleanup(func() { db.SetDatabaseClient(old) })
}

func withMockDb(t *testing.T, queryFunc func(query string, params ...string) ([]map[string]any, error)) {
	t.Helper()
	withMockDbClient(t, &mockDbClient{queryFunc: queryFunc})
}

func isMetadataQuery(query string) bool {
	return strings.Contains(query, "TagBucketMetadata")
}

// withWorkerCount overrides tag_update_worker_count for the duration of a test.
func withWorkerCount(t *testing.T, n int) {
	t.Helper()
	old := GetTagApiConfig()
	cfg := taggingapi_config.TaggingApiConfig{BatchLimit: 5000}
	if old != nil {
		cfg = *old
	}
	cfg.WorkerCount = n
	SetTagApiConfig(&cfg)
	t.Cleanup(func() { SetTagApiConfig(old) })
}

// A non-positive tag_update_worker_count must degrade to a serial write, not
// spawn zero XDAS workers: that reported XdasOk=0/XdasFail=0, skipping the 502
// guard and the Cassandra phase, so the request became a 202 storing nothing.
func TestAddMembersWithXdas_ZeroWorkerCountStillWrites(t *testing.T) {
	setupTestEnvironment()
	withWorkerCount(t, 0)

	var xdasRequests atomic.Int32
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		xdasRequests.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	withMockDbClient(t, &mockDbClient{})

	stats, err := AddMembersWithXdas("some-tag", []string{"AA:BB:CC:DD:EE:01"}, "", TagTypeLegacy)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.XdasOk)
	assert.Equal(t, 1, stats.CassandraOk)
	assert.Equal(t, int32(1), xdasRequests.Load(), "the member write must reach XDAS")
}

// Same zero-worker no-op on the remove path, via removeMembersFromXDAS.
func TestRemoveMembersWithXdas_ZeroWorkerCountStillWrites(t *testing.T) {
	setupTestEnvironment()
	withWorkerCount(t, 0)

	var xdasRequests atomic.Int32
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		xdasRequests.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	withMockDbClient(t, &mockDbClient{})

	stats, err := RemoveMembersWithXdas("some-tag", []string{"AA:BB:CC:DD:EE:01"}, TagTypeLegacy)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.XdasOk)
	assert.Equal(t, 1, stats.CassandraOk)
	assert.Equal(t, int32(1), xdasRequests.Load(), "the member removal must reach XDAS")
}

// Removing a bucket's last member must leave its metadata row in place: the old
// count-then-delete cleanup raced a concurrent add and could delete the row the
// adder just wrote, orphaning members invisible to reads and to tag deletion.
func TestRemoveMembers_EmptiedBucketKeepsMetadataRow(t *testing.T) {
	setupTestEnvironment()

	var batches []*mockBatch
	withMockDbClient(t, &mockDbClient{
		queryFunc: func(query string, params ...string) ([]map[string]any, error) {
			t.Errorf("unexpected query during member removal: %s", query)
			return nil, nil
		},
		modifyFunc: func(query string, params ...string) error {
			t.Errorf("unexpected modify during member removal: %s", query)
			return nil
		},
		execBatchFn: func(b *mockBatch) error {
			batches = append(batches, b)
			return nil
		},
	})

	removed, _, err := RemoveMembers("some-tag", []string{"AA:BB:CC:DD:EE:01"})
	assert.NoError(t, err)
	assert.Equal(t, 1, removed)

	// The member-delete batch is the only database write.
	assert.Len(t, batches, 1)
	for _, stmt := range batches[0].statements {
		assert.Equal(t, QueryRemoveMemberBucketed, stmt)
	}
}

// A bucket emptied of members but still holding its metadata row reads as an
// empty tag, not an error — the tag stays addressable until explicitly deleted.
func TestGetTagById_EmptyBucketsReadAsEmptyTag(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 1}, {"bucket_id": 2}}, nil
		}
		return []map[string]any{}, nil
	})

	members, truncated, _, err := GetTagById("some-tag", TagTypeLegacy)
	assert.NoError(t, err)
	assert.Empty(t, members)
	assert.False(t, truncated)
}

// The non-paginated read must stop fetching once the response is full. Fetching
// every bucket eagerly buffered the whole tag before the merge truncated it —
// gigabytes of heap for a tag spread thinly across many buckets.
func TestFetchMembersFromBuckets_StopsDispatchingWhenFull(t *testing.T) {
	setupTestEnvironment()

	var memberQueries atomic.Int32
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		memberQueries.Add(1)
		bucket := params[1]
		if bucket != "1" && bucket != "2" {
			t.Errorf("bucket %s fetched after the response was already full", bucket)
		}
		rows := make([]map[string]any, 0, 5)
		for i := 0; i < 5; i++ {
			rows = append(rows, map[string]any{"member": fmt.Sprintf("b%s-m%d", bucket, i)})
		}
		return rows, nil
	})

	// Window size = workers = 2: bucket 1 alone fills the limit of 5, so only
	// the first window [1 2] may be fetched; buckets 3-6 must never be queried.
	members, truncated, err := fetchMembersFromBucketsConcurrent(
		"some-tag", []int{1, 2, 3, 4, 5, 6}, 5, 2, &atomic.Int64{})
	assert.NoError(t, err)
	assert.True(t, truncated)
	assert.Equal(t, []string{"b1-m0", "b1-m1", "b1-m2", "b1-m3", "b1-m4"}, members)
	assert.Equal(t, int32(2), memberQueries.Load(), "only the first window of buckets should be fetched")
}

// A read that does not fill the limit still visits every bucket across window
// boundaries and returns members in bucket order.
func TestFetchMembersFromBuckets_MultiWindowReturnsAllInOrder(t *testing.T) {
	setupTestEnvironment()

	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return []map[string]any{{"member": "m" + params[1]}}, nil
	})

	members, truncated, err := fetchMembersFromBucketsConcurrent(
		"some-tag", []int{1, 2, 3, 4, 5}, 100, 2, &atomic.Int64{})
	assert.NoError(t, err)
	assert.False(t, truncated)
	assert.Equal(t, []string{"m1", "m2", "m3", "m4", "m5"}, members)
}

// A page satisfied by the first buckets must not query the ones after them:
// fetching every bucket from the cursor onward cost ~N queries per page.
func TestGetMembersPaginated_StopsDispatchingWhenPageFull(t *testing.T) {
	setupTestEnvironment()
	withWorkerCount(t, 2) // window size = read worker count

	var memberQueries atomic.Int32
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{
				{"bucket_id": 1}, {"bucket_id": 2}, {"bucket_id": 3},
				{"bucket_id": 4}, {"bucket_id": 5}, {"bucket_id": 6},
			}, nil
		}
		memberQueries.Add(1)
		bucket := params[1]
		if bucket != "1" && bucket != "2" {
			t.Errorf("bucket %s fetched after the page was already full", bucket)
		}
		rows := make([]map[string]any, 0, 3)
		for i := 0; i < 3; i++ {
			rows = append(rows, map[string]any{"member": fmt.Sprintf("b%s-m%d", bucket, i)})
		}
		return rows, nil
	})

	// Bucket 1 alone overfills the page of 2, so only the first window [1 2]
	// may be fetched and the cursor must point into bucket 1.
	resp, _, err := GetMembersPaginated("some-tag", 2, "", TagTypeLegacy)
	assert.NoError(t, err)
	assert.Equal(t, []string{"b1-m0", "b1-m1"}, resp.Data)
	assert.True(t, resp.HasMore)
	cur, err := parseBucketedCursor(resp.NextCursor)
	assert.NoError(t, err)
	assert.Equal(t, 1, cur.BucketId)
	assert.Equal(t, "b1-m1", cur.LastMember)
	assert.Equal(t, int32(2), memberQueries.Load(), "only the first window of buckets should be fetched")
}

// A page that needs more than one window still walks all buckets in order and
// terminates correctly.
func TestGetMembersPaginated_PageSpansWindows(t *testing.T) {
	setupTestEnvironment()
	withWorkerCount(t, 2)

	var memberQueries atomic.Int32
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{
				{"bucket_id": 1}, {"bucket_id": 2}, {"bucket_id": 3},
				{"bucket_id": 4}, {"bucket_id": 5},
			}, nil
		}
		memberQueries.Add(1)
		return []map[string]any{{"member": "m" + params[1]}}, nil
	})

	resp, _, err := GetMembersPaginated("some-tag", 10, "", TagTypeLegacy)
	assert.NoError(t, err)
	assert.Equal(t, []string{"m1", "m2", "m3", "m4", "m5"}, resp.Data)
	assert.False(t, resp.HasMore)
	assert.Empty(t, resp.NextCursor)
	assert.Equal(t, int32(5), memberQueries.Load(), "all five buckets fetched across three windows")
}

// A Cassandra failure while listing populated buckets must surface as an
// internal error, not as 404 "tag not found".
func TestGetMembersPaginated_DbErrorIsNot404(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return nil, errors.New("cassandra unavailable")
	})

	resp, _, err := GetMembersPaginated("some-tag", 10, "", TagTypeLegacy)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, xwcommon.GetXconfErrorStatusCode(err))
}

// Unknown tag (no populated buckets, no DB error) is still a 404.
func TestGetMembersPaginated_UnknownTagIs404(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return []map[string]any{}, nil
	})

	resp, _, err := GetMembersPaginated("missing-tag", 10, "", TagTypeLegacy)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, xwcommon.GetXconfErrorStatusCode(err))
}

// A failed bucket read fails the page instead of silently omitting that
// bucket's members and advancing the cursor past them.
func TestGetMembersPaginated_BucketFetchErrorFailsPage(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 1}, {"bucket_id": 2}}, nil
		}
		if params[1] == "2" {
			return nil, errors.New("read timeout")
		}
		return []map[string]any{{"member": "AA:BB:CC:DD:EE:01"}}, nil
	})

	resp, _, err := GetMembersPaginated("some-tag", 10, "", TagTypeLegacy)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket 2")
	assert.Equal(t, http.StatusInternalServerError, xwcommon.GetXconfErrorStatusCode(err))
}

// GetTagById must not return a silently incomplete member list when a bucket
// read fails.
func TestGetTagById_BucketFetchErrorFails(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 7}}, nil
		}
		return nil, errors.New("read timeout")
	})

	members, truncated, _, err := GetTagById("some-tag", TagTypeLegacy)
	assert.Error(t, err)
	assert.Nil(t, members)
	assert.False(t, truncated)
}

// A cursor pointing past the last populated bucket ends the enumeration
// instead of wrapping around to bucket 0 and re-delivering everything.
func TestGetMembersPaginated_CursorBeyondLastBucketEndsEnumeration(t *testing.T) {
	setupTestEnvironment()
	var memberQueries atomic.Int32
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 1}, {"bucket_id": 2}}, nil
		}
		memberQueries.Add(1)
		return []map[string]any{{"member": "m1"}}, nil
	})

	cursor := generateBucketedCursor(500, "")
	resp, _, err := GetMembersPaginated("some-tag", 10, cursor, TagTypeLegacy)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Data)
	assert.False(t, resp.HasMore)
	assert.Empty(t, resp.NextCursor)
	assert.Equal(t, int32(0), memberQueries.Load(), "no member fetches expected past the last populated bucket")
}

// An unparseable cursor is a 400, not a silent restart from bucket 0.
func TestGetMembersPaginated_InvalidCursorReturns400(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		t.Error("no database query expected for an invalid cursor")
		return nil, nil
	})

	resp, _, err := GetMembersPaginated("some-tag", 10, "!!!not-a-cursor!!!", TagTypeLegacy)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, xwcommon.GetXconfErrorStatusCode(err))
}

func TestGetTagMembersHandler_InvalidCursorReturns400(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return nil, errors.New("unexpected db call")
	})

	req := httptest.NewRequest("GET", "/taggingService/tags/some-tag/members?cursor=%21bad%21", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "some-tag"})
	w := httptest.NewRecorder()
	GetTagMembersHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), InvalidCursorErrorMsg)
}

// Regression: two-page walk over two buckets returns each member exactly once
// and terminates.
func TestGetMembersPaginated_TwoPageWalk(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 1}, {"bucket_id": 2}}, nil
		}
		switch params[1] {
		case "1":
			return []map[string]any{{"member": "m1"}}, nil
		case "2":
			return []map[string]any{{"member": "m2"}}, nil
		}
		return nil, fmt.Errorf("unexpected bucket %s", params[1])
	})

	page1, _, err := GetMembersPaginated("some-tag", 1, "", TagTypeLegacy)
	assert.NoError(t, err)
	assert.Equal(t, []string{"m1"}, page1.Data)
	assert.True(t, page1.HasMore)
	assert.NotEmpty(t, page1.NextCursor)

	page2, _, err := GetMembersPaginated("some-tag", 1, page1.NextCursor, TagTypeLegacy)
	assert.NoError(t, err)
	assert.Equal(t, []string{"m2"}, page2.Data)
	assert.False(t, page2.HasMore)
}
