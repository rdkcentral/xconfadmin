package tag

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
