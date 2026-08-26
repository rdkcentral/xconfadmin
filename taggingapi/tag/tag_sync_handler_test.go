package tag

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	xcommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"

	"github.com/stretchr/testify/assert"
)

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
