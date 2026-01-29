package canary

import (
	"net/http"
	"testing"

	common "github.com/rdkcentral/xconfadmin/common"
	"github.com/stretchr/testify/assert"
)

func TestSetCanarySettingValidationPass(t *testing.T) {
	distributionPercentage := 15.0
	maxSize := 5000
	startTime := 1800
	endTime := 3600

	validSettings := &common.CanarySettings{
		CanaryDistributionPercentage: &distributionPercentage,
		CanaryMaxSize:                &maxSize,
		CanaryFwUpgradeStartTime:     &startTime,
		CanaryFwUpgradeEndTime:       &endTime,
	}

	result := SetCanarySetting(validSettings)
	assert.NotEqual(t, http.StatusBadRequest, result.Status)

	validSettings.CanaryMaxSize = nil
	result = SetCanarySetting(validSettings)
	assert.NotEqual(t, http.StatusBadRequest, result.Status)

	validSettings.CanaryDistributionPercentage = nil
	result = SetCanarySetting(validSettings)
	assert.NotEqual(t, http.StatusBadRequest, result.Status)

	validSettings.CanaryFwUpgradeStartTime = nil
	result = SetCanarySetting(validSettings)
	assert.NotEqual(t, http.StatusBadRequest, result.Status)

}

func TestGetCanarySettings(t *testing.T) {
	_, err := GetCanarySettings()
	assert.Error(t, err, "Should return error when app settings are not set")
}

func TestProcessCanarySettings(t *testing.T) {
	settings := map[string]interface{}{
		common.PROP_CANARY_DISTRIBUTION_PERCENTAGE: 20.0,
		common.PROP_CANARY_MAXSIZE:                 10000,
		common.PROP_CANARY_FW_UPGRADE_STARTTIME:    1200,
		common.PROP_CANARY_FW_UPGRADE_ENDTIME:      2400,
	}

	canarySettings, err := ProcessCanarySettings(settings)
	assert.NoError(t, err, "Should not return error for valid settings map")
	assert.NotNil(t, canarySettings, "CanarySettings should not be nil")
}
