package common

import (
	"testing"

	"gotest.tools/assert"
)

func TestIsValidAppSetting(t *testing.T) {
	// Test valid app settings
	assert.Assert(t, IsValidAppSetting(READONLY_MODE))
	assert.Assert(t, IsValidAppSetting(READONLY_MODE_STARTTIME))
	assert.Assert(t, IsValidAppSetting(READONLY_MODE_ENDTIME))
	assert.Assert(t, IsValidAppSetting(PROP_LOCKDOWN_ENABLED))

	// Test invalid app setting
	assert.Assert(t, !IsValidAppSetting("invalid-setting"))
	assert.Assert(t, !IsValidAppSetting(""))
}

func TestIsValidType(t *testing.T) {
	// Test valid types
	assert.Assert(t, isValidType(GenericNamespacedListTypes_STRING))
	assert.Assert(t, isValidType(GenericNamespacedListTypes_MAC_LIST))
	assert.Assert(t, isValidType(GenericNamespacedListTypes_IP_LIST))
	assert.Assert(t, isValidType(GenericNamespacedListTypes_RI_MAC_LIST))

	// Test invalid types
	assert.Assert(t, !isValidType("invalid-type"))
	assert.Assert(t, !isValidType(""))
}
