package tag

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/stretchr/testify/assert"
)

// errReader stands in for a body that dies mid-read: a reset connection or a
// truncated upload.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("connection reset by peer") }

func TestTagSyncTriggerRejectsUnreadableBody(t *testing.T) {
	// A body that fails to read used to look the same as no body at all, so
	// a caller asking for repair got a 202 for a default detect run that
	// held the cluster lock against the run they meant to start.
	req := httptest.NewRequest(http.MethodPost, "/taggingService/tags/sync", errReader{})
	rec := httptest.NewRecorder()

	TriggerTagSyncHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "request body read error")
	assert.NotContains(t, rec.Body.String(), "runId", "no run may have been started")
}

func TestTagSyncRunStatusRejectsMissingRunId(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/taggingService/tags/sync/status/", nil)
	req = mux.SetURLVars(req, map[string]string{common.RunId: ""})
	rec := httptest.NewRecorder()

	TagSyncRunStatusHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), common.RunId)
}

func TestTagSyncAbortOnlyCancelsOwnTenant(t *testing.T) {
	ctxA, cancelA := context.WithCancel(context.Background())
	defer cancelA()
	activeTagSyncMu.Lock()
	activeTagSyncs["TENANT_A"] = activeTagSync{cancel: cancelA, runId: "20260101-000000-a"}
	activeTagSyncMu.Unlock()
	t.Cleanup(func() {
		activeTagSyncMu.Lock()
		delete(activeTagSyncs, "TENANT_A")
		activeTagSyncMu.Unlock()
	})

	abort := func(tenantId string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/taggingService/tags/sync/abort", nil)
		req = req.WithContext(context.WithValue(req.Context(), xhttp.CTX_KEY_TENANT_ID, tenantId))
		rec := httptest.NewRecorder()
		AbortTagSyncHandler(rec, req)
		return rec
	}

	rec := abort("TENANT_B")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.NoError(t, ctxA.Err(), "another tenant's run must keep running")

	rec = abort("TENANT_A")
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Contains(t, rec.Body.String(), "20260101-000000-a")
	assert.Error(t, ctxA.Err())
}
