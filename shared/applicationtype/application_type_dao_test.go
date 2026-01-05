package applicationtype

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetOneApplicationType(t *testing.T) {
	appType := &ApplicationType{
		ID:   "testAppTypeID",
		Name: "testAppType",
	}
	err := SetOneApplicationType(appType)
	if err != nil && strings.Contains(err.Error(), "Table configuration not found") {
		t.Skip("Skipping test: database not configured")
		return
	}
	assert.NoError(t, err)
}

func TestToApplicationType(t *testing.T) {
	appType := &ApplicationType{
		ID:   "test123",
		Name: "testApp",
	}
	result, err := toApplicationType(appType)
	assert.NoError(t, err)
	assert.Equal(t, "test123", result.ID)
	assert.Equal(t, "testApp", result.Name)
	assert.Equal(t, appType, result)

	inputMap := map[string]interface{}{
		"id":   "test456",
		"name": "testApp2",
	}
	result, err = toApplicationType(inputMap)
	assert.NoError(t, err)
	assert.Equal(t, "test456", result.ID)
	assert.Equal(t, "testApp2", result.Name)

	invalidInput := make(chan int)
	result, err = toApplicationType(invalidInput)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetOneApplicationType(t *testing.T) {
	result, err := GetOneApplicationType("7a3rf34ff1d9fa9qa")
	if err != nil && (strings.Contains(err.Error(), "Table configuration not found") ||
		strings.Contains(err.Error(), "cache not found or configured")) {
		t.Skip("Skipping test: database not configured")
		return
	}
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetApplicationTypeByName(t *testing.T) {
	exists, err := GetApplicationTypeByName("test")
	if err != nil && (strings.Contains(err.Error(), "Table configuration not found") ||
		strings.Contains(err.Error(), "cache not found or configured")) {
		t.Skip("Skipping test: database not configured")
		return
	}
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestGetAllApplicationTypeAsList(t *testing.T) {
	result, err := GetAllApplicationTypeAsList()
	if err != nil && (strings.Contains(err.Error(), "Table configuration not found") ||
		strings.Contains(err.Error(), "cache not found or configured")) {
		t.Skip("Skipping test: database not configured")
		return
	}
	assert.NoError(t, err)
	assert.IsType(t, []*ApplicationType{}, result)
}
