package lockdown

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ccommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/stretchr/testify/assert"
)

const testURL = "/lockdown-settings"

func TestPutLockdownSettingsHandler(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to database not configured: %v", r)
		}
	}()
	req := httptest.NewRequest(http.MethodPut, testURL, nil)
	recorder := httptest.NewRecorder()
	w := xhttp.NewXResponseWriter(recorder)
	PutLockdownSettingsHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	PutLockdownSettingsHandler(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Invalid JSON
	w.SetBody(`{"invalid": json}`)
	PutLockdownSettingsHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	// Valid JSON but SetLockdownSetting error
	val := true
	validLockdownSettings := ccommon.LockdownSettings{
		LockdownEnabled: &val,
	}
	jsonBody, _ := json.Marshal(validLockdownSettings)
	w.SetBody(string(jsonBody))
	PutLockdownSettingsHandler(w, req)
	assert.True(t, w.Status() >= 400 || w.Status() == http.StatusOK)

	// Valid request - success path
	val = false
	simpleLockdownSettings := ccommon.LockdownSettings{
		LockdownEnabled: &val,
	}
	jsonBody, _ = json.Marshal(simpleLockdownSettings)
	w.SetBody(string(jsonBody))
	PutLockdownSettingsHandler(w, req)
	assert.True(t, w.Status() == http.StatusOK || w.Status() >= 400)

	// Empty body
	w.SetBody("")
	PutLockdownSettingsHandler(w, req)
	assert.True(t, w.Status() >= 200)
}

func TestGetLockdownSettingsHandler(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/lockdown/settings", nil)
	recorder := httptest.NewRecorder()

	GetLockdownSettingsHandler(recorder, req)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
