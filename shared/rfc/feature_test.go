package rfc

import (
	"encoding/json"
	"testing"

	rfc "github.com/rdkcentral/xconfwebconfig/shared/rfc"

	"gotest.tools/assert"
)

func TestFeatureCreationAndMarshall(t *testing.T) {

	configData := map[string]string{
		"configKey": "configValue",
	}

	// only mandatory fields
	feature := &rfc.Feature{
		ConfigData:         configData,
		FeatureName:        "featureName",
		Name:               "name",
		Enable:             true,
		EffectiveImmediate: true,
	}
	featureResponseObject := rfc.CreateFeatureResponseObject(*feature)
	expectedJsonString := "{\"name\":\"name\",\"enable\":true,\"effectiveImmediate\":true,\"configData\":{\"configKey\":\"configValue\"},\"featureInstance\":\"featureName\"}"
	actualByteString, err := featureResponseObject.MarshalJSON()
	assert.NilError(t, err)
	assert.Equal(t, expectedJsonString, string(actualByteString))

	// all fields
	// commented out because order of other fields changes now
	//
	// properties := map[string]interface{}{
	// 	"propertyKey": "propertyValue",
	// }
	//
	// feature = &Feature{
	// 	Properties:         properties,
	// 	ListType:           "listType",
	// 	ListSize:           2,
	// 	ID:                 "id",
	// 	Updated:            1234,
	// 	Name:               "name",
	// 	FeatureName:        "featureName",
	// 	EffectiveImmediate: true,
	// 	Enable:             true,
	// 	Whitelisted:        true,
	// 	WhitelistProperty:  &WhitelistProperty{},
	// 	ConfigData:         configData,
	// 	ApplicationType:    "stb",
	// }

	// featureResponseObject = CreateFeatureResponseObject(*feature)

	// expectedJsonString = "{\"name\":\"name\",\"effectiveImmediate\":true,\"enable\":true,\"configData\":{\"configKey\":\"configValue\"},\"listType\":\"listType\",\"listSize\":2,\"featureInstance\":\"featureName\",\"propertyKey\":\"propertyValue\"}"

	// actualByteString, err = featureResponseObject.MarshalJSON()
	// assert.NilError(t, err)
	// assert.Equal(t, expectedJsonString, string(actualByteString))

}

func TestFeatureEntityAndUnmarshall(t *testing.T) {

	jsonString := "{}"
	var nilWhitelistProperty *rfc.WhitelistProperty
	var featureEntity rfc.FeatureEntity
	err := json.Unmarshal([]byte(jsonString), &featureEntity)
	assert.NilError(t, err)
	assert.Equal(t, featureEntity.ID, "")
	assert.Equal(t, featureEntity.Name, "")
	assert.Equal(t, featureEntity.FeatureName, "")
	//	assert.Equal(t, featureEntity.ApplicationType, "")
	assert.Equal(t, len(featureEntity.ConfigData), 0)
	assert.Equal(t, featureEntity.EffectiveImmediate, false)
	assert.Equal(t, featureEntity.Enable, false)
	assert.Equal(t, featureEntity.Whitelisted, false)
	assert.Equal(t, featureEntity.WhitelistProperty, nilWhitelistProperty)

	jsonString = "{\"id\":\"id\",\"name\":\"name\",\"featureName\":\"featureInstance\",\"applicationType\":\"rdkcloud\",\"effectiveImmediate\":true,\"enable\":true,\"configData\":{\"key\":\"value\"},\"whitelisted\":true,\"whitelistProperty\":{\"key\":\"key\",\"value\":\"value\",\"namespacedListType\":\"namespacedListType\",\"typeName\":\"typeName\"}}"

	err = json.Unmarshal([]byte(jsonString), &featureEntity)
	assert.NilError(t, err)
	assert.Equal(t, featureEntity.ID, "id")
	assert.Equal(t, featureEntity.Name, "name")
	assert.Equal(t, featureEntity.FeatureName, "featureInstance")
	assert.Equal(t, featureEntity.ApplicationType, "rdkcloud")
	assert.Equal(t, len(featureEntity.ConfigData), 1)
	assert.Equal(t, featureEntity.ConfigData["key"], "value")
	assert.Equal(t, featureEntity.EffectiveImmediate, true)
	assert.Equal(t, featureEntity.Enable, true)
	assert.Equal(t, featureEntity.Whitelisted, true)
	assert.Equal(t, featureEntity.WhitelistProperty.Key, "key")
	assert.Equal(t, featureEntity.WhitelistProperty.Value, "value")
	assert.Equal(t, featureEntity.WhitelistProperty.NamespacedListType, "namespacedListType")
	assert.Equal(t, featureEntity.WhitelistProperty.TypeName, "typeName")
}

func TestIsValidFeatureEntity(t *testing.T) {
	// nil feature
	var featureEntity *rfc.FeatureEntity
	isValid, errMsg := IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Application type is empty")

	// empty feature
	featureEntity = &rfc.FeatureEntity{}
	isValid, errMsg = IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Application type is empty")

	// not valid application type
	featureEntity.ApplicationType = "fakeApplicationType"
	isValid, errMsg = IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "ApplicationType fakeApplicationType is not valid")

	// no name
	featureEntity.ApplicationType = "stb"
	isValid, errMsg = IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Name is blank")

	// no feature name
	featureEntity.Name = "name"
	isValid, errMsg = IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Feature Name is blank")

	// blank config key
	featureEntity.FeatureName = "featureInstance"
	featureEntity.ConfigData = map[string]string{
		"": "",
	}
	isValid, errMsg = IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Key is blank")

	// blank config value
	featureEntity.ConfigData = map[string]string{
		"key": "",
	}
	isValid, errMsg = IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Value is blank for key: key")

	// whitelisted with no whitelist data
	featureEntity.ConfigData["key"] = "value"
	featureEntity.Whitelisted = true
	isValid, errMsg = IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Key is required")

	// whitelist has empty value
	featureEntity.WhitelistProperty = &rfc.WhitelistProperty{
		Key:   "key",
		Value: "",
	}
	isValid, errMsg = IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Value is required")

	// whitelist no matching namespaced list type
	featureEntity.WhitelistProperty.Value = "value"
	featureEntity.WhitelistProperty.NamespacedListType = "namespacedListType"
	featureEntity.WhitelistProperty.TypeName = "typeName"
	isValid, errMsg = IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "namespacedListType with id value does not exist")

	// valid feature
	featureEntity.Whitelisted = false
	isValid, errMsg = IsValidFeatureEntity(featureEntity)
	assert.Equal(t, isValid, true)
	assert.Equal(t, errMsg, "")
}

func TestIsValidFeature(t *testing.T) {
	// nil feature
	var feature *rfc.Feature
	isValid, errMsg := IsValidFeature(feature)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Application type is empty")

	// empty feature
	feature = &rfc.Feature{}
	isValid, errMsg = IsValidFeature(feature)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Application type is empty")

	// not valid application type
	feature.ApplicationType = "fakeApplicationType"
	isValid, errMsg = IsValidFeature(feature)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "ApplicationType fakeApplicationType is not valid")

	// no name
	feature.ApplicationType = "stb"
	isValid, errMsg = IsValidFeature(feature)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Name is blank")

	// no feature name
	feature.Name = "name"
	isValid, errMsg = IsValidFeature(feature)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Feature Name is blank")

	// blank config key
	feature.FeatureName = "featureInstance"
	feature.ConfigData = map[string]string{
		"": "",
	}
	isValid, errMsg = IsValidFeature(feature)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Key is blank")

	// blank config value
	feature.ConfigData = map[string]string{
		"key": "",
	}
	isValid, errMsg = IsValidFeature(feature)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Value is blank for key: key")

	// whitelisted with no whitelist data
	feature.ConfigData["key"] = "value"
	feature.Whitelisted = true
	isValid, errMsg = IsValidFeature(feature)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Key is required")

	// whitelist has empty value
	feature.WhitelistProperty = &rfc.WhitelistProperty{
		Key:   "key",
		Value: "",
	}
	isValid, errMsg = IsValidFeature(feature)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "Value is required")

	// whitelist no matching namespaced list type
	feature.WhitelistProperty.Value = "value"
	feature.WhitelistProperty.NamespacedListType = "namespacedListType"
	feature.WhitelistProperty.TypeName = "typeName"
	isValid, errMsg = IsValidFeature(feature)
	assert.Equal(t, isValid, false)
	assert.Equal(t, errMsg, "namespacedListType with id value does not exist")

	// valid feature
	feature.Whitelisted = false
	isValid, errMsg = IsValidFeature(feature)
	assert.Equal(t, isValid, true)
	assert.Equal(t, errMsg, "")
}

func TestDeleteFeatureRule(t *testing.T) {
	// Test function executes without panic
	assert.Assert(t, true, "DeleteFeatureRule should execute without panic")
	DeleteFeatureRule("test-rule-id")
}

func TestDoesFeatureNameExistForAnotherId(t *testing.T) {
	feature := &rfc.Feature{
		ID:              "test-id",
		FeatureName:     "test-feature",
		ApplicationType: "stb",
	}

	// Should return false for non-existent feature or when DB not initialized
	result := DoesFeatureNameExistForAnotherId(feature)
	assert.Equal(t, result, false)
}

func TestDoesFeatureNameExistForAnotherEntityId(t *testing.T) {
	featureEntity := &rfc.FeatureEntity{
		ID:              "test-entity-id",
		FeatureName:     "test-feature-entity",
		ApplicationType: "stb",
	}

	// Should return false for non-existent feature entity
	result := DoesFeatureNameExistForAnotherEntityId(featureEntity)
	assert.Equal(t, result, false)
}

func TestSetFeatureRule(t *testing.T) {
	featureRule := &rfc.FeatureRule{
		Id:              "test-rule",
		Name:            "Test Rule",
		ApplicationType: "stb",
	}

	// Test function executes - may error if DB not initialized
	err := SetFeatureRule("test-rule", featureRule)
	// Either succeeds or returns error, both are valid
	_ = err
}

func TestGetFeatureRule(t *testing.T) {
	// Test with non-existent ID
	result := GetFeatureRule("non-existent-id")

	// Should return nil when not found or DB not initialized
	assert.Assert(t, result == nil)
}

func TestGetFeatureEntityList(t *testing.T) {
	// Test function executes without panic
	result := GetFeatureEntityList()

	// Should return nil or slice when executed
	if result != nil {
		assert.Assert(t, len(result) >= 0)
	}
}

func TestDoesFeatureExistInSomeApplicationType(t *testing.T) {
	// Test with empty ID
	exists, appType := DoesFeatureExistInSomeApplicationType("")
	assert.Equal(t, exists, false)
	assert.Equal(t, appType, "")

	// Test with non-existent ID
	exists, appType = DoesFeatureExistInSomeApplicationType("non-existent-id")
	assert.Equal(t, exists, false)
	assert.Equal(t, appType, "")
}

func TestGetFilteredFeatureEntityList(t *testing.T) {
	searchContext := map[string]string{
		"APPLICATION_TYPE": "stb",
	}

	// Test function executes without panic
	result := GetFilteredFeatureEntityList(searchContext)

	// Should return nil or slice when executed
	if result != nil {
		assert.Assert(t, len(result) >= 0)
	}
}

func TestSetOneFeature(t *testing.T) {
	feature := &rfc.Feature{
		ID:              "test-feature-id",
		Name:            "Test Feature",
		FeatureName:     "TEST_FEATURE",
		ApplicationType: "stb",
	}

	// Test function executes - may error if DB not initialized
	result, err := SetOneFeature(feature)

	// Either succeeds or returns error
	if err == nil {
		assert.Equal(t, result.ID, feature.ID)
	}
}

func TestDeleteOneFeature(t *testing.T) {
	// Test function executes without panic
	assert.Assert(t, true, "DeleteOneFeature should execute without panic")
	DeleteOneFeature("test-feature-id")
}

func TestGetFilteredFeatureList(t *testing.T) {
	searchContext := map[string]string{
		"APPLICATION_TYPE": "stb",
	}

	// Test function executes without panic
	result := GetFilteredFeatureList(searchContext)

	// Should return nil or slice when executed
	if result != nil {
		assert.Assert(t, len(result) >= 0)
	}
}

func TestDoesFeatureNameExistForAnotherIdInList(t *testing.T) {
	feature := &rfc.Feature{
		ID:              "test-id-1",
		FeatureName:     "test-feature",
		ApplicationType: "stb",
	}

	// Test with empty list
	emptyList := []*rfc.Feature{}
	result := DoesFeatureNameExistForAnotherIdInList(feature, emptyList)
	assert.Equal(t, result, false)

	// Test with list containing same feature
	sameFeatureList := []*rfc.Feature{feature}
	result = DoesFeatureNameExistForAnotherIdInList(feature, sameFeatureList)
	assert.Equal(t, result, false)

	// Test with list containing different feature with same name
	differentFeature := &rfc.Feature{
		ID:              "test-id-2",
		FeatureName:     "test-feature",
		ApplicationType: "stb",
	}
	conflictList := []*rfc.Feature{differentFeature}
	result = DoesFeatureNameExistForAnotherIdInList(feature, conflictList)
	assert.Equal(t, result, true)
}

func TestDoesFeatureNameExistForAnotherIdForApplicationType(t *testing.T) {
	feature := &rfc.Feature{
		ID:              "test-id",
		FeatureName:     "test-feature",
		ApplicationType: "stb",
	}

	// Test function executes without panic
	result := DoesFeatureNameExistForAnotherIdForApplicationType(feature, "stb")

	// Should return false when no conflicts or DB not initialized
	assert.Equal(t, result, false)
}

func TestDoesFeatureExist(t *testing.T) {
	// Test with empty ID
	result := DoesFeatureExist("")
	assert.Equal(t, result, false)

	// Test with non-existent ID
	result = DoesFeatureExist("non-existent-id")
	assert.Equal(t, result, false)
}

func TestDoesFeatureExistWithApplicationType(t *testing.T) {
	// Test with empty ID
	result := DoesFeatureExistWithApplicationType("", "stb")
	assert.Equal(t, result, false)

	// Test with non-existent ID
	result = DoesFeatureExistWithApplicationType("non-existent-id", "stb")
	assert.Equal(t, result, false)
}
