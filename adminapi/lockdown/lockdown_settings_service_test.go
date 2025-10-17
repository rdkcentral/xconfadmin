package lockdown

import (
	"net/http"
	"testing"

	common "github.com/rdkcentral/xconfadmin/common"
	"github.com/stretchr/testify/assert"
)

func TestSetLockdownSettings(t *testing.T) {
	enabled := true
	startTime := "10:00"
	endTime := "18:00"
	modules := "all"

	validSettings := &common.LockdownSettings{
		LockdownEnabled:   &enabled,
		LockdownStartTime: &startTime,
		LockdownEndTime:   &endTime,
		LockdownModules:   &modules,
	}
	result := SetLockdownSetting(validSettings)
	assert.NotEqual(t, http.StatusBadRequest, result.Status, "Validation should pass for valid lockdown settings")

	invalidStartTime := "invalid-time-format"
	validSettings.LockdownStartTime = &invalidStartTime
	result = SetLockdownSetting(validSettings)
	assert.Equal(t, http.StatusBadRequest, result.Status, "Validation should fail for invalid start time format")
}

func TestGetLockdownSettings(t *testing.T) {
	_, err := GetLockdownSettings()
	assert.Error(t, err, "Should return error when app settings are not set")
}

func TestProcessLockdownSettings(t *testing.T) {
	settings := map[string]interface{}{
		common.PROP_LOCKDOWN_ENABLED:   true,
		common.PROP_LOCKDOWN_STARTTIME: "1200",
		common.PROP_LOCKDOWN_ENDTIME:   "2400",
		common.PROP_LOCKDOWN_MODULES:   "module1,module2",
	}

	lockdownSettings, err := ProcessLockdownSettings(settings)
	assert.NoError(t, err, "Should not return error for valid settings map")
	assert.NotNil(t, lockdownSettings, "LockdownSettings should not be nil")
}
