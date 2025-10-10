package canary

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	common "github.com/rdkcentral/xconfadmin/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/stretchr/testify/assert"
)

const testURL = "/canary-settings"

func TestPutCanarySettingsHandler(t *testing.T) {
	originalSatOn := common.SatOn
	common.SatOn = false
	defer func() { common.SatOn = originalSatOn }()

	distributionPercentage := 15.0
	maxSize := 1000
	startTime := 3600
	endTime := 5400

	validCanarySettings := common.CanarySettings{
		CanaryDistributionPercentage: &distributionPercentage,
		CanaryMaxSize:                &maxSize,
		CanaryFwUpgradeStartTime:     &startTime,
		CanaryFwUpgradeEndTime:       &endTime,
	}

	//Valid JSON
	validJSON, err := json.Marshal(validCanarySettings)
	assert.NoError(t, err, "Should be able to marshal valid canary settings")

	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	w.SetBody(string(validJSON))

	req := httptest.NewRequest(http.MethodPut, testURL, nil)
	PutCanarySettingsHandler(w, req)

	assert.NotEqual(t, http.StatusForbidden, w.Status(), "Should not return forbidden when SAT is disabled")

	//Invalid Json
	w.SetBody(`{"invalid": json}`)
	req = httptest.NewRequest(http.MethodPut, testURL, nil)
	PutCanarySettingsHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	//Invalid Auth
	common.SatOn = true
	w.SetBody(`{"canaryDistributionPercentage": 15}`)
	req = httptest.NewRequest(http.MethodPut, testURL, nil)
	PutCanarySettingsHandler(w, req)
}

func TestGetCanarySettingsHandler(t *testing.T) {
	originalSatOn := common.SatOn
	common.SatOn = false
	defer func() { common.SatOn = originalSatOn }()

	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	req := httptest.NewRequest(http.MethodGet, testURL, nil)
	GetCanarySettingsHandler(w, req)
	assert.NotEqual(t, http.StatusUnauthorized, w.Status(), "Should not return unauthorized when SAT is disabled")

	//Invalid Auth
	common.SatOn = true
	req = httptest.NewRequest(http.MethodGet, testURL, nil)
	GetCanarySettingsHandler(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Status())
}
