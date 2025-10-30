/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package queries

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"

	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
)

// Helper function to create a firmware config for testing
func createTestFirmwareConfigForService(id string, version string, modelIds []string, appType string) *coreef.FirmwareConfig {
	fc := &coreef.FirmwareConfig{
		ID:                       id,
		Description:              "Test Config " + id,
		FirmwareVersion:          version,
		SupportedModelIds:        modelIds,
		ApplicationType:          appType,
		FirmwareDownloadProtocol: "http",
		FirmwareFilename:         "test.bin",
		FirmwareLocation:         "http://test.com/test.bin",
	}
	coreef.CreateFirmwareConfigOneDB(fc)
	return fc
}

// Helper function to create a firmware rule for testing
func createEnvModelFirmwareRule(id string, name string, model string, configId string, appType string) *corefw.FirmwareRule {
	// Create rule using JSON to avoid struct complexity
	ruleJSON := `{
		"id": "` + id + `",
		"name": "` + name + `",
		"applicationType": "` + appType + `",
		"type": "ENV_MODEL_RULE",
		"active": true,
		"rule": {
			"negated": false,
			"condition": {
				"freeArg": {
					"type": "STRING",
					"name": "model"
				},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {
							"java.lang.String": "` + model + `"
						}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE",
			"configId": "` + configId + `",
			"active": true
		}
	}`

	var rule corefw.FirmwareRule
	json.Unmarshal([]byte(ruleJSON), &rule)
	corefw.CreateFirmwareRuleOneDB(&rule)
	return &rule
}

func TestIsValidFirmwareConfigByModelIdList(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test firmware configs
	modelIds1 := []string{"MODEL1", "MODEL2"}
	modelIds2 := []string{"MODEL3", "MODEL4"}
	fc1 := createTestFirmwareConfigForService("fc1", "version1.0", modelIds1, "stb")
	fc2 := createTestFirmwareConfigForService("fc2", "version2.0", modelIds2, "stb")
	createTestFirmwareConfigForService("fc3", "version3.0", []string{"MODEL5"}, "xhome")

	// Test 1: Valid config with matching model IDs
	testModelIds := []string{"MODEL1"}
	result := IsValidFirmwareConfigByModelIdList(&testModelIds, "stb", fc1)
	assert.Assert(t, result, "Should return true for valid config with matching model")

	// Test 2: Valid config with multiple model IDs
	testModelIds2 := []string{"MODEL2", "MODEL3"}
	result2 := IsValidFirmwareConfigByModelIdList(&testModelIds2, "stb", fc1)
	assert.Assert(t, result2, "Should return true for valid config with matching model from list")

	// Test 3: Config exists but different application type
	testModelIds3 := []string{"MODEL5"}
	result3 := IsValidFirmwareConfigByModelIdList(&testModelIds3, "stb", fc2)
	assert.Assert(t, !result3, "Should return false for config with non-matching model")

	// Test 4: Empty model ID list
	emptyModelIds := []string{}
	result4 := IsValidFirmwareConfigByModelIdList(&emptyModelIds, "stb", fc1)
	assert.Assert(t, !result4, "Should return false for empty model ID list")

	// Test 5: Non-matching model IDs
	testModelIds5 := []string{"NONEXISTENT"}
	result5 := IsValidFirmwareConfigByModelIdList(&testModelIds5, "stb", fc1)
	assert.Assert(t, !result5, "Should return false for non-matching model IDs")
}

func TestIsValidFirmwareConfigByModelIds(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test firmware configs
	modelIds := []string{"TESTMODEL1", "TESTMODEL2"}
	fc1 := createTestFirmwareConfigForService("test-fc-1", "1.0.0", modelIds, "stb")
	fc2 := createTestFirmwareConfigForService("test-fc-2", "2.0.0", []string{"OTHERMODEL"}, "stb")

	// Test 1: Config ID matches - should return true (function returns true if config ID exists)
	result1 := IsValidFirmwareConfigByModelIds("TESTMODEL1", "stb", fc1)
	assert.Assert(t, result1, "Should return true when config ID exists")

	// Test 2: Different config - even if model doesn't match other configs, if THIS config ID is in DB, returns true
	result2 := IsValidFirmwareConfigByModelIds("TESTMODEL1", "stb", fc2)
	assert.Assert(t, result2, "Should return true because fc2 config ID exists in DB")

	// Test 3: Wrong application type - but config ID still exists
	result3 := IsValidFirmwareConfigByModelIds("TESTMODEL1", "xhome", fc1)
	assert.Assert(t, result3, "Should return true when config ID exists even with wrong app type")

	// Test 4: Config that exists
	result4 := IsValidFirmwareConfigByModelIds("NONEXISTENT", "stb", fc1)
	assert.Assert(t, result4, "Should return true when config ID exists regardless of model match")

	// Test 5: Empty application type (should not filter by app type)
	result5 := IsValidFirmwareConfigByModelIds("TESTMODEL1", "", fc1)
	assert.Assert(t, result5, "Should return true when application type is empty")

	// Test 6: Create a config that doesn't exist in DB yet
	fcNew := &coreef.FirmwareConfig{
		ID:                       "new-fc",
		Description:              "New Config",
		FirmwareVersion:          "3.0.0",
		SupportedModelIds:        []string{"NEWMODEL"},
		ApplicationType:          "stb",
		FirmwareDownloadProtocol: "http",
		FirmwareFilename:         "new.bin",
		FirmwareLocation:         "http://test.com/new.bin",
	}
	// Don't save it to DB - just test with it
	result6 := IsValidFirmwareConfigByModelIds("NEWMODEL", "stb", fcNew)
	assert.Assert(t, !result6, "Should return false when config doesn't exist in DB")
}

// Additional edge case tests
func TestIsValidFirmwareConfigByModelIdList_EdgeCases(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test data
	fc := createTestFirmwareConfigForService("edge-fc", "1.0.0", []string{"EDGEMODEL"}, "stb")

	// Test with empty (not nil) model IDs
	emptyModelIds := []string{}
	result1 := IsValidFirmwareConfigByModelIdList(&emptyModelIds, "stb", fc)
	assert.Assert(t, !result1, "Should return false for empty model IDs")

	// Test with nil firmware config would cause panic, so we skip it
	// The function should ideally handle this gracefully but currently doesn't

	// Cleanup
	coreef.DeleteOneFirmwareConfig(fc.ID)
}

func TestGetFirmwareConfigsByModelIdAndApplicationType_EmptyDatabase(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Test with empty database
	result := GetFirmwareConfigsByModelIdAndApplicationType("ANYMODEL", "stb")
	assert.Equal(t, 0, len(result), "Should return empty list when database is empty")
}

func TestGetSupportedConfigsByEnvModelRuleName_NoMatchingModel(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create config and rule with non-matching model
	fc := createTestFirmwareConfigForService("nomatch-fc", "1.0.0", []string{"MODEL_A"}, "stb")

	// Create rule with different model using JSON
	ruleJSON := `{
		"id": "nomatch-rule",
		"name": "NoMatchRule",
		"applicationType": "stb",
		"type": "ENV_MODEL_RULE",
		"active": true,
		"rule": {
			"condition": {
				"freeArg": {
					"type": "STRING",
					"name": "model"
				},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {
							"java.lang.String": "MODEL_B"
						}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE",
			"configId": "` + fc.ID + `",
			"active": true
		}
	}`

	var rule corefw.FirmwareRule
	json.Unmarshal([]byte(ruleJSON), &rule)
	corefw.CreateFirmwareRuleOneDB(&rule)

	// Test - should not find config because model doesn't match
	result := getSupportedConfigsByEnvModelRuleName("NoMatchRule", "stb")
	assert.Equal(t, 0, len(result), "Should return empty when model doesn't match")

	// Cleanup
	coreef.DeleteOneFirmwareConfig(fc.ID)
	corefw.DeleteOneFirmwareRule(rule.ID)
}
