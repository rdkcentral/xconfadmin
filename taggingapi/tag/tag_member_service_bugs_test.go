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

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// mockDbClient injects Cassandra query behavior for service-level tests.
// The embedded interface panics on any method the test does not stub,
// which flags unexpected database usage.
type mockDbClient struct {
	db.DatabaseClient
	queryFunc func(query string, params ...string) ([]map[string]any, error)
}

func (m *mockDbClient) QueryXconfDataRows(query string, params ...string) ([]map[string]any, error) {
	return m.queryFunc(query, params...)
}

func withMockDb(t *testing.T, queryFunc func(query string, params ...string) ([]map[string]any, error)) {
	t.Helper()
	old := db.GetDatabaseClient()
	db.SetDatabaseClient(&mockDbClient{queryFunc: queryFunc})
	t.Cleanup(func() { db.SetDatabaseClient(old) })
}

func isMetadataQuery(query string) bool {
	return strings.Contains(query, "TagBucketMetadata")
}

// BUG-1: a Cassandra failure while listing populated buckets must surface as an
// internal error, not as 404 "tag not found".
func TestGetMembersPaginated_DbErrorIsNot404(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return nil, errors.New("cassandra unavailable")
	})

	resp, err := GetMembersPaginated("some-tag", 10, "")
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

	resp, err := GetMembersPaginated("missing-tag", 10, "")
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, xwcommon.GetXconfErrorStatusCode(err))
}

// BUG-2: a failed bucket read fails the page instead of silently omitting that
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

	resp, err := GetMembersPaginated("some-tag", 10, "")
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket 2")
	assert.Equal(t, http.StatusInternalServerError, xwcommon.GetXconfErrorStatusCode(err))
}

// BUG-2 (non-paginated path): GetTagById must not return a silently incomplete
// member list when a bucket read fails.
func TestGetTagById_BucketFetchErrorFails(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 7}}, nil
		}
		return nil, errors.New("read timeout")
	})

	members, truncated, err := GetTagById("some-tag")
	assert.Error(t, err)
	assert.Nil(t, members)
	assert.False(t, truncated)
}

// BUG-3: a cursor pointing past the last populated bucket ends the enumeration
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
	resp, err := GetMembersPaginated("some-tag", 10, cursor)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Data)
	assert.False(t, resp.HasMore)
	assert.Empty(t, resp.NextCursor)
	assert.Equal(t, int32(0), memberQueries.Load(), "no member fetches expected past the last populated bucket")
}

// BUG-3 (related): an unparseable cursor is a 400, not a silent restart from
// bucket 0.
func TestGetMembersPaginated_InvalidCursorReturns400(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		t.Error("no database query expected for an invalid cursor")
		return nil, nil
	})

	resp, err := GetMembersPaginated("some-tag", 10, "!!!not-a-cursor!!!")
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

	page1, err := GetMembersPaginated("some-tag", 1, "")
	assert.NoError(t, err)
	assert.Equal(t, []string{"m1"}, page1.Data)
	assert.True(t, page1.HasMore)
	assert.NotEmpty(t, page1.NextCursor)

	page2, err := GetMembersPaginated("some-tag", 1, page1.NextCursor)
	assert.NoError(t, err)
	assert.Equal(t, []string{"m2"}, page2.Data)
	assert.False(t, page2.HasMore)
}
