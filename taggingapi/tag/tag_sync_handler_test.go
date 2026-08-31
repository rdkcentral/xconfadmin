package tag

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	xcommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"

	"github.com/stretchr/testify/assert"
)

// errReader stands in for a body that dies mid-read: a reset connection or a
// truncated upload.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("connection reset by peer") }

func TestTagSyncStatusRequiresToolReadPermission(t *testing.T) {
	// The status payload is operational detail, not public tagging data: an
	// authenticated caller with no tools capability must not read it.
	prev := xcommon.SatOn
	xcommon.SatOn = true
	t.Cleanup(func() { xcommon.SatOn = prev })

	req := httptest.NewRequest(http.MethodGet, "/taggingService/tags/sync/status", nil)
	req = req.WithContext(context.WithValue(req.Context(), xhttp.CTX_KEY_CAPABILITIES, []string{"x1:coast:someotherservice:read"}))
	rec := httptest.NewRecorder()

	TagSyncStatusHandler(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.NotContains(t, rec.Body.String(), "history", "no run data may leak with the rejection")
}

func TestTagSyncTriggerRejectsUnreadableBody(t *testing.T) {
	// A body that fails to read used to look the same as no body at all, so
	// a caller asking for repair got a 202 for a default detect run that
	// held the cluster lock against the run they meant to start.
	prev := xcommon.SatOn
	xcommon.SatOn = false // permission is not what this test is about
	t.Cleanup(func() { xcommon.SatOn = prev })

	req := httptest.NewRequest(http.MethodPost, "/taggingService/tags/sync", errReader{})
	rec := httptest.NewRecorder()

	TriggerTagSyncHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "request body read error")
	assert.NotContains(t, rec.Body.String(), "runId", "no run may have been started")
}
