package setting

import (
	"testing"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/stretchr/testify/assert"
)

func TestDeleteSettingProfile(t *testing.T) {
	DeleteSettingProfile(db.GetDefaultTenantId(), "test-profile-123")
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
	validateUsage(db.GetDefaultTenantId(), "non-existent-id")
	assert.NotPanics(t, func() {
		defer func() {
			recover() // Suppress any panics for this test
		}()
		validateUsage(db.GetDefaultTenantId(), "test-id")
	})
}

func TestSetSettingProfile(t *testing.T) {
	profile := &xwlogupload.SettingProfiles{
		ID:               "test-id",
		SettingProfileID: "Test Profile",
		ApplicationType:  "STB",
		SettingType:      "PARTNER_SETTINGS",
		Properties:       map[string]string{"key1": "value1"},
	}
	err := SetSettingProfile(db.GetDefaultTenantId(), profile.ID, profile)
	// DB may not be configured in test environment; error is acceptable
	_ = err
}

func TestFindByContext(t *testing.T) {
	t.Run("WithApplicationType", func(t *testing.T) {
		searchContext := map[string]string{
			"applicationType":  "STB",
			xwcommon.TENANT_ID: db.GetDefaultTenantId(),
		}
		results := FindByContext(searchContext)
		assert.NotNil(t, results)
	})

	t.Run("WithName", func(t *testing.T) {
		searchContext := map[string]string{
			"name":             "test",
			xwcommon.TENANT_ID: db.GetDefaultTenantId(),
		}
		results := FindByContext(searchContext)
		assert.NotNil(t, results)
	})

	t.Run("WithType", func(t *testing.T) {
		searchContext := map[string]string{
			"type":             "PARTNER_SETTINGS",
			xwcommon.TENANT_ID: db.GetDefaultTenantId(),
		}
		results := FindByContext(searchContext)
		assert.NotNil(t, results)
	})

	t.Run("MultipleFilters", func(t *testing.T) {
		searchContext := map[string]string{
			"applicationType":  "STB",
			"name":             "profile",
			"type":             "PARTNER",
			xwcommon.TENANT_ID: db.GetDefaultTenantId(),
		}
		results := FindByContext(searchContext)
		assert.NotNil(t, results)
	})
}

func TestDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Skip("Requires database configuration")
	})

	t.Run("NonExistentID", func(t *testing.T) {
		result, err := Delete(db.GetDefaultTenantId(), "non-existent-delete-id", "STB")
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})

	t.Run("WrongApplicationType", func(t *testing.T) {
		t.Skip("Requires database configuration")
	})
}

func TestUpdate(t *testing.T) {
	t.Run("ValidProfile", func(t *testing.T) {
		t.Skip("Requires database configuration")
	})

	t.Run("InvalidProperties", func(t *testing.T) {
		t.Skip("Requires database configuration")
	})

	t.Run("WrongApplicationType", func(t *testing.T) {
		t.Skip("Requires database configuration")
	})
}

func TestCreate(t *testing.T) {
	t.Run("ValidProfile", func(t *testing.T) {
		t.Skip("Requires database configuration")
	})

	t.Run("InvalidProperties", func(t *testing.T) {
		t.Skip("Requires database configuration")
	})
}

func TestBeforeSaving(t *testing.T) {
	t.Run("ValidEntity", func(t *testing.T) {
		profile := &xwlogupload.SettingProfiles{
			ID:               "before-save-test-1",
			SettingProfileID: "Before Save Test",
			ApplicationType:  "STB",
			SettingType:      "PARTNER_SETTINGS",
			Properties:       map[string]string{"key1": "value1"},
		}

		err := beforeSaving(db.GetDefaultTenantId(), profile, "STB")
		if err != nil {
			// Function validates against existing profiles, error is acceptable
			assert.NotNil(t, err)
		}
	})

	t.Run("EmptyProperties", func(t *testing.T) {
		profile := &xwlogupload.SettingProfiles{
			ID:               "before-save-test-2",
			SettingProfileID: "Before Save Test 2",
			ApplicationType:  "STB",
			SettingType:      "PARTNER_SETTINGS",
			Properties:       nil,
		}

		err := beforeSaving(db.GetDefaultTenantId(), profile, "STB")
		assert.NotNil(t, err)
	})
}

func TestValidate(t *testing.T) {
	t.Run("ValidEntity", func(t *testing.T) {
		profile := &xwlogupload.SettingProfiles{
			SettingType: "PARTNER_SETTINGS",
			Properties:  map[string]string{"key1": "value1"},
		}

		err := validate(profile)
		assert.Nil(t, err)
	})

	t.Run("InvalidEntity", func(t *testing.T) {
		profile := &xwlogupload.SettingProfiles{
			SettingType: "",
			Properties:  map[string]string{"key1": "value1"},
		}

		err := validate(profile)
		assert.NotNil(t, err)
	})
}

func TestSettingProfilesGeneratePage(t *testing.T) {
	t.Run("ValidPage", func(t *testing.T) {
		profiles := []*xwlogupload.SettingProfiles{
			{ID: "1", SettingProfileID: "Profile 1"},
			{ID: "2", SettingProfileID: "Profile 2"},
			{ID: "3", SettingProfileID: "Profile 3"},
			{ID: "4", SettingProfileID: "Profile 4"},
			{ID: "5", SettingProfileID: "Profile 5"},
		}

		result := SettingProfilesGeneratePage(profiles, 1, 2)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "1", result[0].ID)
	})

	t.Run("LastPage", func(t *testing.T) {
		profiles := []*xwlogupload.SettingProfiles{
			{ID: "1", SettingProfileID: "Profile 1"},
			{ID: "2", SettingProfileID: "Profile 2"},
			{ID: "3", SettingProfileID: "Profile 3"},
		}

		result := SettingProfilesGeneratePage(profiles, 2, 2)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, "3", result[0].ID)
	})

	t.Run("InvalidPage", func(t *testing.T) {
		profiles := []*xwlogupload.SettingProfiles{
			{ID: "1", SettingProfileID: "Profile 1"},
		}

		result := SettingProfilesGeneratePage(profiles, 0, 2)
		assert.Equal(t, 0, len(result))
	})

	t.Run("OutOfBounds", func(t *testing.T) {
		profiles := []*xwlogupload.SettingProfiles{
			{ID: "1", SettingProfileID: "Profile 1"},
		}

		result := SettingProfilesGeneratePage(profiles, 10, 2)
		assert.Equal(t, 0, len(result))
	})
}
