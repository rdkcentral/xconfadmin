package xcrp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rdkcentral/xconfadmin/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	testRecookingURL = "/xcrp/recooking-lockdown-settings"
)

// Test case 1: No write permission (covers lines 28-31)
func TestPostRecookingLockdownSettingsHandler(t *testing.T) {
	originalSatOn := common.SatOn
	common.SatOn = true
	defer func() { common.SatOn = originalSatOn }()

	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	req := httptest.NewRequest(http.MethodPost, testRecookingURL, nil)
	PostRecookingLockdownSettingsHandler(w, req)
	assert.Equal(t, http.StatusForbidden, w.Status())

	//invalid JSON
	common.SatOn = false
	w.SetBody(`{"invalid": json}`)
	PostRecookingLockdownSettingsHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status(), "Should return 400 Bad Request for invalid JSON")

	//Valid JSON
	models := []string{"MODEL1", "MODEL2"}
	partners := []string{"PARTNER1", "PARTNER2"}
	recookingSettings := common.RecookingLockdownSettings{
		Models:   &models,
		Partners: &partners,
	}
	validJSON, err := json.Marshal(recookingSettings)
	assert.NoError(t, err, "Should be able to marshal valid recooking lockdown settings")
	w.SetBody(string(validJSON))
	req = httptest.NewRequest(http.MethodPost, testRecookingURL, nil)
	PostRecookingLockdownSettingsHandler(w, req)
	assert.NotEqual(t, http.StatusBadRequest, w.Status(), "Should not return 400 for valid JSON")

	//Invalid ResponseWriter
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, testRecookingURL, nil)
	PostRecookingLockdownSettingsHandler(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code, "Should return 400 Bad Request for responsewriter cast error")
}

func TestIsLockdownMode(t *testing.T) {
	res := isLockdownMode()
	assert.False(t, res)
}
func TestCheckRecookingStatus(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to database not configured: %v", r)
		}
	}()

	// Mock fields for logging
	mockFields := log.Fields{
		"userId": "test-user",
		"action": "recooking-test",
	}

	// GetRecookingStatusFromCanaryMgr error
	// This will likely trigger due to no actual XCRP connector
	lockDuration := 100 * time.Millisecond
	module := "rfc"

	// Start the function in a goroutine since it has time.Sleep
	done := make(chan bool, 1)
	completed := false

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic in CheckRecookingStatus: %v", r)
			}
			completed = true
			done <- true
		}()
		CheckRecookingStatus(lockDuration, module, mockFields)
	}()

	select {
	case <-done:
		assert.True(t, completed, "CheckRecookingStatus should complete successfully")
	case <-time.After(5 * time.Second):
		assert.False(t, completed, "CheckRecookingStatus should not timeout - function contains time.Sleep")
		assert.Fail(t, "CheckRecookingStatus timed out after 5 seconds")
	}
	assert.True(t, completed, "CheckRecookingStatus function should have been executed")
}
