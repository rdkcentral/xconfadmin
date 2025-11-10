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
package telemetry

import (
	"testing"

	xcommon "github.com/rdkcentral/xconfadmin/common"
	"github.com/rdkcentral/xconfadmin/shared/logupload"

	"github.com/google/uuid"
	"gotest.tools/assert"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

// Helper function to create a TelemetryTwoRule for testing
func createTestTelemetryTwoRule(name, appType string, boundProfileIDs []string) *xwlogupload.TelemetryTwoRule {
	rule := &xwlogupload.TelemetryTwoRule{
		ID:                uuid.New().String(),
		Name:              name,
		ApplicationType:   appType,
		BoundTelemetryIDs: boundProfileIDs,
		NoOp:              false,
	}
	// Create a simple rule with MODEL condition
	cond := re.NewCondition(coreef.RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg("TEST_MODEL"))
	rule.Rule = re.Rule{Condition: cond}
	return rule
}

// Helper function to create a TelemetryTwoRule with collection fixed arg
func createTestTelemetryTwoRuleWithCollectionFixedArg(name, appType string) *xwlogupload.TelemetryTwoRule {
	rule := &xwlogupload.TelemetryTwoRule{
		ID:                uuid.New().String(),
		Name:              name,
		ApplicationType:   appType,
		BoundTelemetryIDs: []string{},
		NoOp:              true,
	}
	// Create rule with collection fixed arg (using array of strings)
	collectionValues := []string{"value1", "value2", "testvalue"}
	fixedArg := re.NewFixedArg(collectionValues)
	cond := re.NewCondition(coreef.RuleFactoryMODEL, re.StandardOperationIn, fixedArg)
	rule.Rule = re.Rule{Condition: cond}
	return rule
}

// Helper function to create a TelemetryTwoProfile
func createTestTelemetryTwoProfile(name, appType string) *xwlogupload.TelemetryTwoProfile {
	profile := &xwlogupload.TelemetryTwoProfile{
		ID:              uuid.New().String(),
		Name:            name,
		ApplicationType: appType,
	}
	return profile
}

func TestFindByContext_NameFilter(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test rules
	rule1 := createTestTelemetryTwoRule("TestRule1", "stb", []string{})
	rule2 := createTestTelemetryTwoRule("AnotherRule", "stb", []string{})
	rule3 := createTestTelemetryTwoRule("TestRule3", "stb", []string{})

	logupload.SetOneTelemetryTwoRule(rule1.ID, rule1)
	logupload.SetOneTelemetryTwoRule(rule2.ID, rule2)
	logupload.SetOneTelemetryTwoRule(rule3.ID, rule3)

	t.Run("FilterByName_Found", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.NAME_UPPER: "TestRule",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 2, len(results))
		// Verify both TestRule1 and TestRule3 are returned
		foundNames := make(map[string]bool)
		for _, r := range results {
			foundNames[r.Name] = true
		}
		assert.Assert(t, foundNames["TestRule1"])
		assert.Assert(t, foundNames["TestRule3"])
	})

	t.Run("FilterByName_NotFound", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.NAME_UPPER: "NonExistent",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 0, len(results))
	})

	t.Run("FilterByName_EmptyString", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.NAME_UPPER: "",
		}
		results := findByContext(nil, searchContext)
		// Empty string should return all rules
		assert.Equal(t, 3, len(results))
	})

	t.Run("FilterByName_CaseInsensitive", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.NAME_UPPER: "testrule",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 2, len(results))
	})
}

func TestFindByContext_ProfileFilter(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test profiles
	profile1 := createTestTelemetryTwoProfile("Profile1", "stb")
	profile2 := createTestTelemetryTwoProfile("TestProfile", "stb")

	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profile1.ID, profile1)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profile2.ID, profile2)

	// Create rules with different profile bindings
	rule1 := createTestTelemetryTwoRule("Rule1", "stb", []string{profile1.ID})
	rule2 := createTestTelemetryTwoRule("Rule2", "stb", []string{profile2.ID})
	rule3 := createTestTelemetryTwoRule("Rule3", "stb", []string{}) // No profiles

	logupload.SetOneTelemetryTwoRule(rule1.ID, rule1)
	logupload.SetOneTelemetryTwoRule(rule2.ID, rule2)
	logupload.SetOneTelemetryTwoRule(rule3.ID, rule3)

	t.Run("FilterByProfile_Found", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.PROFILE: "Profile1",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Rule1", results[0].Name)
	})

	t.Run("FilterByProfile_NotFound", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.PROFILE: "NonExistentProfile",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 0, len(results))
	})

	t.Run("FilterByProfile_RuleWithNoProfiles", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.PROFILE: "Profile1",
		}
		results := findByContext(nil, searchContext)
		// Rule3 with no profiles should not be included
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Rule1", results[0].Name)
	})

	t.Run("FilterByProfile_CaseInsensitive", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.PROFILE: "testprofile",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Rule2", results[0].Name)
	})
}

func TestFindByContext_FreeArgFilter(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create rules with different free args
	rule1 := createTestTelemetryTwoRule("Rule1", "stb", []string{})
	// rule1 already has MODEL as free arg from createTestTelemetryTwoRule

	rule2 := createTestTelemetryTwoRule("Rule2", "stb", []string{})
	// Add a different free arg condition
	cond2 := re.NewCondition(coreef.RuleFactoryMAC, re.StandardOperationIs, re.NewFixedArg("AA:BB:CC:DD:EE:FF"))
	rule2.Rule = re.Rule{Condition: cond2}

	logupload.SetOneTelemetryTwoRule(rule1.ID, rule1)
	logupload.SetOneTelemetryTwoRule(rule2.ID, rule2)

	t.Run("FilterByFreeArg_Found", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FREE_ARG: "model",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Rule1", results[0].Name)
	})

	t.Run("FilterByFreeArg_NotFound", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FREE_ARG: "nonexistent",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 0, len(results))
	})

	t.Run("FilterByFreeArg_CaseInsensitive", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FREE_ARG: "MAC",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Rule2", results[0].Name)
	})
}

func TestFindByContext_FixedArgFilter_CollectionValue(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create rule with collection fixed arg
	rule1 := createTestTelemetryTwoRuleWithCollectionFixedArg("Rule1", "stb")
	logupload.SetOneTelemetryTwoRule(rule1.ID, rule1)

	t.Run("FilterByFixedArg_CollectionValue_Found", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FIXED_ARG: "testvalue",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Rule1", results[0].Name)
	})

	t.Run("FilterByFixedArg_CollectionValue_NotFound", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FIXED_ARG: "notinlist",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 0, len(results))
	})

	t.Run("FilterByFixedArg_CollectionValue_CaseInsensitive", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FIXED_ARG: "VALUE1",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Rule1", results[0].Name)
	})
}

func TestFindByContext_FixedArgFilter_StringValue(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create rule with string fixed arg
	rule1 := createTestTelemetryTwoRule("Rule1", "stb", []string{})
	// rule1 already has string fixed arg "TEST_MODEL"

	logupload.SetOneTelemetryTwoRule(rule1.ID, rule1)

	t.Run("FilterByFixedArg_StringValue_Found", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FIXED_ARG: "TEST_MODEL",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Rule1", results[0].Name)
	})

	t.Run("FilterByFixedArg_StringValue_PartialMatch", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FIXED_ARG: "MODEL",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
	})

	t.Run("FilterByFixedArg_StringValue_NotFound", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FIXED_ARG: "NONEXISTENT",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 0, len(results))
	})

	t.Run("FilterByFixedArg_StringValue_CaseInsensitive", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FIXED_ARG: "test_model",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
	})
}

func TestFindByContext_FixedArgFilter_ExistsOperation(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create rule with EXISTS operation (should be skipped for string value check)
	rule1 := createTestTelemetryTwoRule("Rule1", "stb", []string{})
	cond := re.NewCondition(coreef.RuleFactoryMODEL, re.StandardOperationExists, nil)
	rule1.Rule = re.Rule{Condition: cond}
	logupload.SetOneTelemetryTwoRule(rule1.ID, rule1)

	t.Run("FilterByFixedArg_ExistsOperation_Skipped", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.FIXED_ARG: "anything",
		}
		results := findByContext(nil, searchContext)
		// Should not match because EXISTS operation doesn't have a string value to compare
		assert.Equal(t, 0, len(results))
	})
}

func TestFindByContext_ApplicationTypeFilter(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	rule1 := createTestTelemetryTwoRule("Rule1", "stb", []string{})
	rule2 := createTestTelemetryTwoRule("Rule2", "xhome", []string{})

	logupload.SetOneTelemetryTwoRule(rule1.ID, rule1)
	logupload.SetOneTelemetryTwoRule(rule2.ID, rule2)

	t.Run("FilterByApplicationType_STB", func(t *testing.T) {
		searchContext := map[string]string{
			xwcommon.APPLICATION_TYPE: "stb",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Rule1", results[0].Name)
	})

	t.Run("FilterByApplicationType_ALL", func(t *testing.T) {
		searchContext := map[string]string{
			xwcommon.APPLICATION_TYPE: shared.ALL,
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 2, len(results))
	})

	t.Run("FilterByApplicationType_Empty", func(t *testing.T) {
		searchContext := map[string]string{
			xwcommon.APPLICATION_TYPE: "",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 2, len(results))
	})
}

func TestFindByContext_CombinedFilters(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	profile1 := createTestTelemetryTwoProfile("TestProfile", "stb")
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profile1.ID, profile1)

	rule1 := createTestTelemetryTwoRule("TestRule1", "stb", []string{profile1.ID})
	rule2 := createTestTelemetryTwoRule("TestRule2", "stb", []string{})
	rule3 := createTestTelemetryTwoRule("OtherRule", "xhome", []string{})

	logupload.SetOneTelemetryTwoRule(rule1.ID, rule1)
	logupload.SetOneTelemetryTwoRule(rule2.ID, rule2)
	logupload.SetOneTelemetryTwoRule(rule3.ID, rule3)

	t.Run("CombinedFilters_NameAndApplicationType", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.NAME_UPPER:        "TestRule",
			xwcommon.APPLICATION_TYPE: "stb",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 2, len(results))
	})

	t.Run("CombinedFilters_NameAndProfile", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.NAME_UPPER: "TestRule",
			xcommon.PROFILE:    "TestProfile",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "TestRule1", results[0].Name)
	})

	t.Run("CombinedFilters_AllFilters", func(t *testing.T) {
		searchContext := map[string]string{
			xcommon.NAME_UPPER:        "TestRule1",
			xwcommon.APPLICATION_TYPE: "stb",
			xcommon.PROFILE:           "TestProfile",
			xcommon.FREE_ARG:          "model",
			xcommon.FIXED_ARG:         "TEST_MODEL",
		}
		results := findByContext(nil, searchContext)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "TestRule1", results[0].Name)
	})
}

func TestGetOne_ErrorCondition(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	t.Run("GetOne_NotFound_ReturnsRemoteError", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		result, err := GetOne(nonExistentID)

		assert.Assert(t, result == nil)
		assert.Assert(t, err != nil)
		assert.Assert(t, err.Error() != "")
	})

	t.Run("GetOne_Success", func(t *testing.T) {
		rule := createTestTelemetryTwoRule("TestRule", "stb", []string{})
		logupload.SetOneTelemetryTwoRule(rule.ID, rule)

		result, err := GetOne(rule.ID)
		assert.Assert(t, err == nil)
		assert.Assert(t, result != nil)
		assert.Equal(t, rule.ID, result.ID)
		assert.Equal(t, "TestRule", result.Name)
	})
}

func TestDelete_ErrorCondition(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	t.Run("Delete_NotFound_ReturnsRemoteError", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		result, err := Delete(nonExistentID)

		assert.Assert(t, result == nil)
		assert.Assert(t, err != nil)
		assert.Assert(t, err.Error() != "")
	})

	t.Run("Delete_Success", func(t *testing.T) {
		rule := createTestTelemetryTwoRule("TestRule", "stb", []string{})
		logupload.SetOneTelemetryTwoRule(rule.ID, rule)

		result, err := Delete(rule.ID)
		assert.Assert(t, err == nil)
		assert.Assert(t, result != nil)
		assert.Equal(t, rule.ID, result.ID)

		// Verify it's deleted
		//deletedRule, _ := GetOne(rule.ID)
		//assert.Assert(t, deletedRule == nil)
	})
}
