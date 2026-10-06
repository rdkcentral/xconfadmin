package tag

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	admincommon "github.com/rdkcentral/xconfadmin/common"
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

func tagSyncRequest(method string, path string, headerTenantId string, contextTenantId string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	if headerTenantId != "" {
		req.Header.Set("tenantId", headerTenantId)
	}
	if contextTenantId != "" {
		req = req.WithContext(context.WithValue(req.Context(), xhttp.CTX_KEY_TENANT_ID, contextTenantId))
	}
	return req
}

func TestAbortTagSyncCannotCancelAnotherTenant(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	activeTagSyncMu.Lock()
	activeTagSyncCancel = cancel
	activeTagSyncRunId = "run-a"
	activeTagSyncTenantId = "TENANT-A"
	activeTagSyncMu.Unlock()
	defer func() {
		activeTagSyncMu.Lock()
		activeTagSyncCancel = nil
		activeTagSyncRunId = ""
		activeTagSyncTenantId = ""
		activeTagSyncMu.Unlock()
	}()

	rec := httptest.NewRecorder()
	AbortTagSyncHandler(rec, tagSyncRequest(http.MethodPost, "/taggingService/tags/sync/abort", "TENANT-B", "TENANT-B"))

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.NoError(t, ctx.Err(), "another tenant must not be able to cancel the active run")
}

func TestAbortTagSyncUsesResolvedTenantContext(t *testing.T) {
	oldSatOn := admincommon.SatOn
	admincommon.SatOn = false
	defer func() { admincommon.SatOn = oldSatOn }()

	ctx, cancel := context.WithCancel(context.Background())
	activeTagSyncMu.Lock()
	activeTagSyncCancel = cancel
	activeTagSyncRunId = "run-a"
	activeTagSyncTenantId = "TENANT-A"
	activeTagSyncMu.Unlock()
	defer func() {
		activeTagSyncMu.Lock()
		activeTagSyncCancel = nil
		activeTagSyncRunId = ""
		activeTagSyncTenantId = ""
		activeTagSyncMu.Unlock()
	}()

	rec := httptest.NewRecorder()
	AbortTagSyncHandler(rec, tagSyncRequest(http.MethodPost, "/taggingService/tags/sync/abort", "IGNORED-RAW-HEADER", "TENANT-A"))

	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.ErrorIs(t, ctx.Err(), context.Canceled, "the handler must use the tenant resolved into request context")
}

func TestAbortTagSyncChecksSharedWritePermission(t *testing.T) {
	oldSatOn := admincommon.SatOn
	admincommon.SatOn = true
	defer func() { admincommon.SatOn = oldSatOn }()

	req := tagSyncRequest(http.MethodPost, "/taggingService/tags/sync/abort", "TENANT-A", "TENANT-A")
	req = req.WithContext(context.WithValue(req.Context(), xhttp.CTX_KEY_AUTH_TYPE, xhttp.AUTH_TYPE_SAT_V2))
	rec := httptest.NewRecorder()

	AbortTagSyncHandler(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "No capabilities found in SAT token")
}

func TestTriggerTagSyncChecksPermissionBeforeParsingBody(t *testing.T) {
	oldSatOn := admincommon.SatOn
	admincommon.SatOn = true
	defer func() { admincommon.SatOn = oldSatOn }()

	req := httptest.NewRequest(http.MethodPost, "/taggingService/tags/sync", strings.NewReader("{"))
	ctx := context.WithValue(req.Context(), xhttp.CTX_KEY_TENANT_ID, "TENANT-A")
	ctx = context.WithValue(ctx, xhttp.CTX_KEY_AUTH_TYPE, xhttp.AUTH_TYPE_SAT_V2)
	rec := httptest.NewRecorder()

	TriggerTagSyncHandler(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "No capabilities found in SAT token")
}
