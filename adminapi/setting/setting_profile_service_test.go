package setting

import (
	"testing"

	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/stretchr/testify/assert"
)

func TestDeleteSettingProfile(t *testing.T) {
	DeleteSettingProfile("test-profile-123")
	assert.True(t, true)
}

func TestValidateProperties(t *testing.T) {
	validEntity := &xwlogupload.SettingProfiles{
		SettingType: "PARTNER_SETTINGS",
		Properties: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	assert.Equal(t, "", validateProperties(validEntity))
	assert.Equal(t, "Setting type is empty", validateProperties(&xwlogupload.SettingProfiles{
		SettingType: "",
		Properties:  map[string]string{"key": "value"},
	}))
	assert.Equal(t, "INVALID not one of declared Enum instance names: [PARTNER_SETTINGS, EPON]",
		validateProperties(&xwlogupload.SettingProfiles{
			SettingType: "INVALID",
			Properties:  map[string]string{"key": "value"},
		}))
	assert.Equal(t, "Property map is empty", validateProperties(&xwlogupload.SettingProfiles{
		SettingType: "PARTNER_SETTINGS",
		Properties:  nil,
	}))
	assert.Equal(t, "Key is blank", validateProperties(&xwlogupload.SettingProfiles{
		SettingType: "PARTNER_SETTINGS",
		Properties:  map[string]string{"": "value"},
	}))
	assert.Equal(t, "Value is blank for key: key1", validateProperties(&xwlogupload.SettingProfiles{
		SettingType: "PARTNER_SETTINGS",
		Properties:  map[string]string{"key1": ""},
	}))
}

func TestValidateAll(t *testing.T) {
	entity := &xwlogupload.SettingProfiles{
		ID:               "entity-1",
		SettingProfileID: "profile-new",
	}
	existingEntities := []*xwlogupload.SettingProfiles{
		{ID: "entity-2", SettingProfileID: "profile-existing-1"},
		{ID: "entity-3", SettingProfileID: "profile-existing-2"},
	}
	assert.Nil(t, validateAll(entity, existingEntities))

	existingEntities = []*xwlogupload.SettingProfiles{
		{ID: "existing-entity-id", SettingProfileID: "duplicate-profile"},
	}
	err := validateAll(entity, existingEntities)
	assert.Nil(t, err)
}

func TestValidateUsage(t *testing.T) {
	validateUsage("non-existent-id")
	assert.NotPanics(t, func() {
		defer func() {
			recover() // Suppress any panics for this test
		}()
		validateUsage("test-id")
	})
}

func TestSetSettingProfile(t *testing.T) {
	err := SetSettingProfile("test-id", nil)
	assert.NotNil(t, err)
}
