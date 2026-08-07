package tag

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// withMockXdasSync points the GroupServiceSyncConnector at an httptest server
// for the duration of the test.
func withMockXdasSync(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	connector := xhttp.WebConfServer.GroupServiceSyncConnector
	oldHost := connector.GetGroupServiceSyncHost()
	oldClient := connector.Client
	connector.SetGroupServiceSyncHost(server.URL)
	connector.SetAddGroupMemberTemplate("%s/ft/%s")
	connector.SetRemoveGroupMemberTemplate("%s/ft/%s?field=%s")
	connector.Client = &xhttp.HttpClient{Client: server.Client()}
	t.Cleanup(func() {
		connector.SetGroupServiceSyncHost(oldHost)
		connector.Client = oldClient
	})
}

// When XDAS removes only part of a bucket's members, DeleteTag must report an
// error instead of claiming a completed deletion.
func TestDeleteTag_PartialXdasFailureReturnsError(t *testing.T) {
	setupTestEnvironment()
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "MFAIL") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	var executedBatches []string
	mock := &mockDbClient{
		queryFunc: func(query string, params ...string) ([]map[string]any, error) {
			if isMetadataQuery(query) {
				return []map[string]any{{"bucket_id": 1}}, nil
			}
			if strings.Contains(query, "count(*)") {
				return []map[string]any{{"count": int64(1)}}, nil
			}
			return []map[string]any{{"member": "MFAIL"}, {"member": "MOK"}}, nil
		},
		execBatchFn: func(batch *mockBatch) error {
			executedBatches = append(executedBatches, batch.statements...)
			return nil
		},
	}
	withMockDbClient(t, mock)

	err := DeleteTag("some-tag", "test-audit-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "partial XDAS deletion")

	// Bucket metadata must survive so a retry can resume the deletion
	for _, stmt := range executedBatches {
		assert.NotContains(t, stmt, "TagBucketMetadata",
			"bucket metadata must not be deleted after a partial XDAS failure")
	}
}

// A second DELETE for a tag whose deletion is already running returns 202
// without spawning another background deleter.
func TestDeleteTagHandler_DeduplicatesConcurrentDeletions(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 1}}, nil
		}
		t.Error("no member fetches expected while deletion is already in flight")
		return nil, nil
	})

	deletionKey := "dup-tag"
	inFlightTagDeletions.Store(deletionKey, true)
	t.Cleanup(func() { inFlightTagDeletions.Delete(deletionKey) })

	req := httptest.NewRequest("DELETE", "/taggingService/tags/dup-tag", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "dup-tag"})
	rec := httptest.NewRecorder()
	DeleteTagHandler(xwhttp.NewXResponseWriter(rec), req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Contains(t, rec.Body.String(), "already in progress")
}

// The dedup guard is released once the background deletion finishes, so a
// later DELETE for the same tag is accepted again.
func TestDeleteTagHandler_ReleasesGuardAfterCompletion(t *testing.T) {
	setupTestEnvironment()
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 1}}, nil
		}
		// Empty bucket: deletion completes immediately
		return []map[string]any{}, nil
	})

	req := httptest.NewRequest("DELETE", "/taggingService/tags/guard-tag", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "guard-tag"})
	rec := httptest.NewRecorder()
	DeleteTagHandler(xwhttp.NewXResponseWriter(rec), req)
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Contains(t, rec.Body.String(), "queued for processing")

	// Wait for the background deletion to finish and release the guard.
	// This must complete before the test ends so the mock DB is not restored
	// under a still-running goroutine.
	deletionKey := "guard-tag"
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, running := inFlightTagDeletions.Load(deletionKey); !running {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("background deletion did not release the in-flight guard")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Unknown tag on GET /tags/{tag} is a typed 404, not a string-matched error.
func TestGetTagByIdHandler_UnknownTagReturns404(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return []map[string]any{}, nil
	})

	req := httptest.NewRequest("GET", "/taggingService/tags/missing-tag", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "missing-tag"})
	w := httptest.NewRecorder()
	GetTagByIdHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "missing-tag tag not found")
}

// A database failure on GET /tags/{tag} is a 500, not a 404.
func TestGetTagByIdHandler_DbErrorReturns500(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return nil, errors.New("cassandra unavailable")
	})

	req := httptest.NewRequest("GET", "/taggingService/tags/some-tag", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "some-tag"})
	w := httptest.NewRecorder()
	GetTagByIdHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
