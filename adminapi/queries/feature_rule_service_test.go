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
	"testing"

	"github.com/google/uuid"
	xcommon "github.com/rdkcentral/xconfadmin/common"
	xshared "github.com/rdkcentral/xconfadmin/shared"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	xwrfc "github.com/rdkcentral/xconfwebconfig/shared/rfc"
	"github.com/stretchr/testify/assert"
)

// Helper functions
func makeFeatureForService(name string, app string) *xwrfc.Feature {
	f := &xwrfc.Feature{
		ID:                 uuid.New().String(),
		Name:               name,
		FeatureName:        name + "Fn",
		ApplicationType:    app,
		Enable:             true,
		EffectiveImmediate: true,
		ConfigData:         map[string]string{"key": "value"},
	}
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_XCONF_FEATURE, f.ID, f)
	return f
}

func makeRuleForService() *re.Rule {
	return &re.Rule{
		Condition: CreateCondition(
			*re.NewFreeArg(re.StandardFreeArgTypeString, "model"),
			re.StandardOperationIs,
			"X1",
		),
	}
}

func makeRuleWithPercentRange(startRange, endRange string) *re.Rule {
	return &re.Rule{
		Condition: CreateCondition(
			*re.NewFreeArg(re.StandardFreeArgTypeString, "eStbMac"),
			re.StandardOperationRange,
			startRange+"-"+endRange,
		),
	}
}

func makeFeatureRuleForService(featureIds []string, app string, priority int, name string) *xwrfc.FeatureRule {
	if name == "" {
		name = "FR-" + uuid.New().String()
	}
	fr := &xwrfc.FeatureRule{
		Id:              uuid.New().String(),
		Name:            name,
		ApplicationType: app,
		FeatureIds:      featureIds,
		Priority:        priority,
		Rule:            makeRuleForService(),
	}
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr.Id, fr)
	return fr
}

func cleanupServiceTest() {
	tables := []string{ds.TABLE_FEATURE_CONTROL_RULE, ds.TABLE_XCONF_FEATURE}
	for _, tbl := range tables {
		list, _ := ds.GetCachedSimpleDao().GetAllAsList(tbl, 0)
		for _, inst := range list {
			switch v := inst.(type) {
			case *xwrfc.FeatureRule:
				ds.GetCachedSimpleDao().DeleteOne(tbl, v.Id)
			case *xwrfc.Feature:
				ds.GetCachedSimpleDao().DeleteOne(tbl, v.ID)
			}
		}
		ds.GetCachedSimpleDao().RefreshAll(tbl)
	}
}

// Test reorganizeFeatureRulePriorities
func TestReorganizeFeatureRulePriorities(t *testing.T) {
	cleanupServiceTest()

	// Create feature rules with different priorities
	f := makeFeatureForService("Feature1", "stb")
	fr1 := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "Rule1")
	fr2 := makeFeatureRuleForService([]string{f.ID}, "stb", 2, "Rule2")
	fr3 := makeFeatureRuleForService([]string{f.ID}, "stb", 3, "Rule3")
	fr4 := makeFeatureRuleForService([]string{f.ID}, "stb", 4, "Rule4")

	itemsList := []*xwrfc.FeatureRule{fr1, fr2, fr3, fr4}

	t.Run("MoveDown", func(t *testing.T) {
		// Move item from priority 2 to 4
		result := reorganizeFeatureRulePriorities(itemsList, 2, 4)
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result)) // Should return altered sublist
		// Verify the moved item has new priority
		for _, item := range itemsList {
			if item.Id == fr2.Id {
				assert.Equal(t, 4, item.Priority)
			}
		}
	})

	t.Run("MoveUp", func(t *testing.T) {
		// Reset
		fr1.Priority = 1
		fr2.Priority = 2
		fr3.Priority = 3
		fr4.Priority = 4
		itemsList = []*xwrfc.FeatureRule{fr1, fr2, fr3, fr4}

		// Move item from priority 4 to 1
		result := reorganizeFeatureRulePriorities(itemsList, 4, 1)
		assert.NotNil(t, result)
		assert.Equal(t, 4, len(result)) // Should return altered sublist
		// Verify the moved item has new priority
		for _, item := range itemsList {
			if item.Id == fr4.Id {
				assert.Equal(t, 1, item.Priority)
			}
		}
	})

	t.Run("NewPriorityTooLow", func(t *testing.T) {
		// Reset
		fr1.Priority = 1
		fr2.Priority = 2
		fr3.Priority = 3
		itemsList = []*xwrfc.FeatureRule{fr1, fr2, fr3}

		// Try to set priority to 0 (should default to length)
		result := reorganizeFeatureRulePriorities(itemsList, 2, 0)
		assert.NotNil(t, result)
		// Item should be moved to last position (priority = length)
		for _, item := range itemsList {
			if item.Id == fr2.Id {
				assert.Equal(t, 3, item.Priority)
			}
		}
	})

	t.Run("NewPriorityTooHigh", func(t *testing.T) {
		// Reset
		fr1.Priority = 1
		fr2.Priority = 2
		fr3.Priority = 3
		itemsList = []*xwrfc.FeatureRule{fr1, fr2, fr3}

		// Try to set priority to 10 (should default to length)
		result := reorganizeFeatureRulePriorities(itemsList, 2, 10)
		assert.NotNil(t, result)
		// Item should be moved to last position
		for _, item := range itemsList {
			if item.Id == fr2.Id {
				assert.Equal(t, 3, item.Priority)
			}
		}
	})

	t.Run("SamePriority", func(t *testing.T) {
		// Reset
		fr1.Priority = 1
		fr2.Priority = 2
		itemsList = []*xwrfc.FeatureRule{fr1, fr2}

		// Keep same priority
		result := reorganizeFeatureRulePriorities(itemsList, 2, 2)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
	})
}

// Test getAlteredFeatureRuleSubList
func TestGetAlteredFeatureRuleSubList(t *testing.T) {
	cleanupServiceTest()

	f := makeFeatureForService("Feature1", "stb")
	fr1 := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "Rule1")
	fr2 := makeFeatureRuleForService([]string{f.ID}, "stb", 2, "Rule2")
	fr3 := makeFeatureRuleForService([]string{f.ID}, "stb", 3, "Rule3")
	fr4 := makeFeatureRuleForService([]string{f.ID}, "stb", 4, "Rule4")
	fr5 := makeFeatureRuleForService([]string{f.ID}, "stb", 5, "Rule5")

	itemsList := []*xwrfc.FeatureRule{fr1, fr2, fr3, fr4, fr5}

	t.Run("MoveDown_Priority2to4", func(t *testing.T) {
		result := getAlteredFeatureRuleSubList(itemsList, 2, 4)
		assert.Equal(t, 3, len(result))
		assert.Equal(t, fr2.Id, result[0].Id)
		assert.Equal(t, fr4.Id, result[2].Id)
	})

	t.Run("MoveUp_Priority4to2", func(t *testing.T) {
		result := getAlteredFeatureRuleSubList(itemsList, 4, 2)
		assert.Equal(t, 3, len(result))
		assert.Equal(t, fr2.Id, result[0].Id)
		assert.Equal(t, fr4.Id, result[2].Id)
	})

	t.Run("SamePriority", func(t *testing.T) {
		result := getAlteredFeatureRuleSubList(itemsList, 3, 3)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, fr3.Id, result[0].Id)
	})

	t.Run("FirstToLast", func(t *testing.T) {
		result := getAlteredFeatureRuleSubList(itemsList, 1, 5)
		assert.Equal(t, 5, len(result))
	})
}

// Test addNewFeatureRuleAndReorganize
func TestAddNewFeatureRuleAndReorganize(t *testing.T) {
	cleanupServiceTest()

	f := makeFeatureForService("Feature1", "stb")
	fr1 := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "Rule1")
	fr2 := makeFeatureRuleForService([]string{f.ID}, "stb", 2, "Rule2")
	fr3 := makeFeatureRuleForService([]string{f.ID}, "stb", 3, "Rule3")

	itemsList := []*xwrfc.FeatureRule{fr1, fr2, fr3}

	t.Run("AddAtEnd", func(t *testing.T) {
		newRule := makeFeatureRuleForService([]string{f.ID}, "stb", 4, "NewRule1")
		result := addNewFeatureRuleAndReorganize(newRule, itemsList)
		assert.Equal(t, 4, len(result))
	})

	t.Run("AddAtBeginning", func(t *testing.T) {
		fr1.Priority = 1
		fr2.Priority = 2
		fr3.Priority = 3
		itemsList = []*xwrfc.FeatureRule{fr1, fr2, fr3}

		newRule := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "NewRule2")
		result := addNewFeatureRuleAndReorganize(newRule, itemsList)
		assert.Equal(t, 4, len(result))
		// All items should be reorganized
		assert.NotNil(t, result)
	})

	t.Run("AddInMiddle", func(t *testing.T) {
		fr1.Priority = 1
		fr2.Priority = 2
		fr3.Priority = 3
		itemsList = []*xwrfc.FeatureRule{fr1, fr2, fr3}

		newRule := makeFeatureRuleForService([]string{f.ID}, "stb", 2, "NewRule3")
		result := addNewFeatureRuleAndReorganize(newRule, itemsList)
		assert.Equal(t, 3, len(result)) // Returns altered sublist
	})
}

// Test FindFeatureRuleByContext
func TestFindFeatureRuleByContext(t *testing.T) {
	cleanupServiceTest()

	f1 := makeFeatureForService("SearchFeature1", "stb")
	f2 := makeFeatureForService("SearchFeature2", "rdkcloud")
	_ = makeFeatureRuleForService([]string{f1.ID}, "stb", 1, "SearchRule1")
	_ = makeFeatureRuleForService([]string{f1.ID}, "stb", 2, "SearchRule2")
	_ = makeFeatureRuleForService([]string{f2.ID}, "rdkcloud", 1, "CloudRule1")

	// Add a rule with collection fixed arg
	freeArg := re.NewFreeArg(re.StandardFreeArgTypeString, "partnerId")
	fixedArg := re.NewFixedArg([]string{"partner1", "partner2"})
	cond := re.NewCondition(freeArg, re.StandardOperationIn, fixedArg)
	ruleWithCollection := &re.Rule{Condition: cond}

	fr4 := &xwrfc.FeatureRule{
		Id:              uuid.New().String(),
		Name:            "RuleWithCollection",
		ApplicationType: "stb",
		FeatureIds:      []string{f1.ID},
		Priority:        3,
		Rule:            ruleWithCollection,
	}
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr4.Id, fr4)

	t.Run("FilterByApplicationType_STB", func(t *testing.T) {
		context := map[string]string{xshared.APPLICATION_TYPE: "stb"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 2)
		for _, rule := range result {
			assert.True(t, rule.ApplicationType == "stb" || rule.ApplicationType == shared.ALL)
		}
	})

	t.Run("FilterByApplicationType_RdkCloud", func(t *testing.T) {
		context := map[string]string{xshared.APPLICATION_TYPE: "rdkcloud"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
		for _, rule := range result {
			assert.True(t, rule.ApplicationType == "rdkcloud" || rule.ApplicationType == shared.ALL)
		}
	})

	t.Run("FilterByFeatureInstance", func(t *testing.T) {
		context := map[string]string{xcommon.FEATURE_INSTANCE: "SearchFeature1"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("FilterByFeatureInstance_CaseInsensitive", func(t *testing.T) {
		context := map[string]string{xcommon.FEATURE_INSTANCE: "searchfeature1"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("FilterByFeatureInstance_NoMatch", func(t *testing.T) {
		context := map[string]string{xcommon.FEATURE_INSTANCE: "NonExistentFeature"}
		result := FindFeatureRuleByContext(context)
		assert.Equal(t, 0, len(result))
	})

	t.Run("FilterByName", func(t *testing.T) {
		context := map[string]string{xcommon.NAME_UPPER: "SearchRule1"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("FilterByName_PartialMatch", func(t *testing.T) {
		context := map[string]string{xcommon.NAME_UPPER: "search"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("FilterByName_CaseInsensitive", func(t *testing.T) {
		context := map[string]string{xcommon.NAME_UPPER: "searchrule"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("FilterByFreeArg", func(t *testing.T) {
		context := map[string]string{xcommon.FREE_ARG: "model"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("FilterByFreeArg_CaseInsensitive", func(t *testing.T) {
		context := map[string]string{xcommon.FREE_ARG: "MODEL"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("FilterByFreeArg_NoMatch", func(t *testing.T) {
		context := map[string]string{xcommon.FREE_ARG: "nonexistentkey"}
		result := FindFeatureRuleByContext(context)
		assert.Equal(t, 0, len(result))
	})

	t.Run("FilterByFixedArg_StringValue", func(t *testing.T) {
		context := map[string]string{xcommon.FIXED_ARG: "X1"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("FilterByFixedArg_CollectionValue", func(t *testing.T) {
		context := map[string]string{xcommon.FIXED_ARG: "partner1"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("FilterByFixedArg_CaseInsensitive", func(t *testing.T) {
		context := map[string]string{xcommon.FIXED_ARG: "x1"}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("CombinedFilters", func(t *testing.T) {
		context := map[string]string{
			xshared.APPLICATION_TYPE: "stb",
			xcommon.NAME_UPPER:       "SearchRule1",
		}
		result := FindFeatureRuleByContext(context)
		assert.True(t, len(result) >= 1)
	})

	t.Run("EmptyContext", func(t *testing.T) {
		context := map[string]string{}
		result := FindFeatureRuleByContext(context)
		// Should return all rules sorted by priority
		assert.True(t, len(result) >= 4)
	})

	t.Run("NilFeatureRule_Skipped", func(t *testing.T) {
		// This tests the nil check in the function
		context := map[string]string{}
		result := FindFeatureRuleByContext(context)
		assert.NotNil(t, result)
	})
}

// Test ValidateFeatureRule
func TestValidateFeatureRule(t *testing.T) {
	cleanupServiceTest()

	f := makeFeatureForService("ValidateFeature", "stb")

	t.Run("NilFeatureRule", func(t *testing.T) {
		err := ValidateFeatureRule(nil, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "FeatureRule is empty")
	})

	t.Run("NilRule", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Rule:            nil,
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Rule is empty")
	})

	t.Run("EmptyName", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Rule:            makeRuleForService(),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "FeatureRule name is blank")
	})

	t.Run("NoFeatures", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{},
			Rule:            makeRuleForService(),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Features should be specified")
	})

	t.Run("TooManyFeatures", func(t *testing.T) {
		// Create more features than allowed
		featureIds := make([]string, xcommon.AllowedNumberOfFeatures+1)
		for i := 0; i < len(featureIds); i++ {
			tmpF := makeFeatureForService("Feature"+string(rune(i)), "stb")
			featureIds[i] = tmpF.ID
		}
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      featureIds,
			Rule:            makeRuleForService(),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Number of Features should be up to")
	})

	t.Run("NonExistentFeature", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{"nonexistent-id"},
			Rule:            makeRuleForService(),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("FeatureApplicationTypeMismatch", func(t *testing.T) {
		rdkFeature := makeFeatureForService("RdkFeature", "rdkcloud")
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{rdkFeature.ID},
			Rule:            makeRuleForService(),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Application Mismatch")
	})

	t.Run("InvalidApplicationType", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "invalid",
			FeatureIds:      []string{f.ID},
			Rule:            makeRuleForService(),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
	})

	t.Run("ApplicationTypeMismatchWithParam", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "rdkcloud",
			FeatureIds:      []string{f.ID},
			Rule:            makeRuleForService(),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "doesn't match")
	})

	t.Run("InvalidPercentRange_StartTooLow", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Rule:            makeRuleWithPercentRange("-1", "50"),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Start range")
	})

	t.Run("InvalidPercentRange_StartTooHigh", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Rule:            makeRuleWithPercentRange("100", "101"),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Start range")
	})

	t.Run("InvalidPercentRange_EndTooLow", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Rule:            makeRuleWithPercentRange("0", "-1"),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "End range")
	})

	t.Run("InvalidPercentRange_EndTooHigh", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Rule:            makeRuleWithPercentRange("0", "101"),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "End range")
	})

	t.Run("InvalidPercentRange_StartGreaterThanEnd", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Rule:            makeRuleWithPercentRange("60", "40"),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Start range should be less than end range")
	})

	t.Run("OverlappingPercentRanges", func(t *testing.T) {
		// Create a rule with two overlapping ranges
		compound := re.NewEmptyRule()
		compound.AddCompoundPart(*CreateRule("", *re.NewFreeArg(re.StandardFreeArgTypeString, "eStbMac"), re.StandardOperationRange, "0-50"))
		compound.AddCompoundPart(*CreateRule(re.RelationAnd, *re.NewFreeArg(re.StandardFreeArgTypeString, "eStbMac"), re.StandardOperationRange, "40-80"))

		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Rule:            compound,
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Ranges overlap")
	})

	t.Run("ValidPercentRange", func(t *testing.T) {
		fr := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "TestRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Rule:            makeRuleWithPercentRange("0", "50"),
		}
		err := ValidateFeatureRule(fr, "stb")
		assert.Nil(t, err)
	})

	t.Run("ValidRule_Success", func(t *testing.T) {
		fr := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "ValidRule")
		err := ValidateFeatureRule(fr, "stb")
		assert.Nil(t, err)
	})
}

// Test parsePercentRange
func TestParsePercentRange(t *testing.T) {
	t.Run("ValidRange", func(t *testing.T) {
		result, err := parsePercentRange("10-50")
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, float64(10), result.StartRange)
		assert.Equal(t, float64(50), result.EndRange)
	})

	t.Run("ValidRange_WithSpaces", func(t *testing.T) {
		result, err := parsePercentRange("  20-60  ")
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, float64(20), result.StartRange)
		assert.Equal(t, float64(60), result.EndRange)
	})

	t.Run("InvalidFormat_NoDash", func(t *testing.T) {
		result, err := parsePercentRange("50")
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Range format exception")
	})

	t.Run("InvalidFormat_InvalidStartRange", func(t *testing.T) {
		result, err := parsePercentRange("abc-50")
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not valid")
	})

	t.Run("InvalidFormat_InvalidEndRange", func(t *testing.T) {
		result, err := parsePercentRange("10-xyz")
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not valid")
	})

	t.Run("ValidRange_Decimals", func(t *testing.T) {
		result, err := parsePercentRange("10.5-50.7")
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 10.5, result.StartRange)
		assert.Equal(t, 50.7, result.EndRange)
	})
}

// Test validateAllFeatureRule
func TestValidateAllFeatureRule(t *testing.T) {
	cleanupServiceTest()

	f := makeFeatureForService("Feature1", "stb")
	existingRule := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "ExistingRule")

	t.Run("DuplicateName_SameApplicationType", func(t *testing.T) {
		newRule := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "ExistingRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        2,
			Rule:            makeRuleForService(),
		}
		err := validateAllFeatureRule(newRule)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Name is already used")
	})

	t.Run("DuplicateName_DifferentApplicationType", func(t *testing.T) {
		f2 := makeFeatureForService("Feature2", "rdkcloud")
		newRule := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "ExistingRule",
			ApplicationType: "rdkcloud",
			FeatureIds:      []string{f2.ID},
			Priority:        1,
			Rule:            makeRuleForService(),
		}
		err := validateAllFeatureRule(newRule)
		assert.Nil(t, err) // Different app type, should be OK
	})

	t.Run("DuplicateRule_SameApplicationType", func(t *testing.T) {
		newRule := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "DifferentName",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        2,
			Rule:            existingRule.Rule,
		}
		err := validateAllFeatureRule(newRule)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Rule has duplicate")
	})

	t.Run("SameId_Skipped", func(t *testing.T) {
		// Same ID means it's an update, not a duplicate
		ruleUpdate := &xwrfc.FeatureRule{
			Id:              existingRule.Id,
			Name:            "UpdatedName",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        1,
			Rule:            makeRuleForService(),
		}
		err := validateAllFeatureRule(ruleUpdate)
		assert.Nil(t, err)
	})

	t.Run("UniqueName_UniqueRule_Success", func(t *testing.T) {
		newRule := &xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "CompletelyNewRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        3,
			Rule: &re.Rule{
				Condition: CreateCondition(
					*re.NewFreeArg(re.StandardFreeArgTypeString, "firmwareVersion"),
					re.StandardOperationIs,
					"1.2.3",
				),
			},
		}
		err := validateAllFeatureRule(newRule)
		assert.Nil(t, err)
	})
}

// Test getPercentRanges
func TestGetPercentRanges(t *testing.T) {
	t.Run("SingleRange", func(t *testing.T) {
		rule := makeRuleWithPercentRange("10", "50")
		ranges, err := getPercentRanges(rule)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(ranges))
		assert.Equal(t, float64(10), ranges[0].StartRange)
		assert.Equal(t, float64(50), ranges[0].EndRange)
	})

	t.Run("MultipleRanges", func(t *testing.T) {
		compound := re.NewEmptyRule()
		compound.AddCompoundPart(*CreateRule("", *re.NewFreeArg(re.StandardFreeArgTypeString, "eStbMac"), re.StandardOperationRange, "0-25"))
		compound.AddCompoundPart(*CreateRule(re.RelationAnd, *re.NewFreeArg(re.StandardFreeArgTypeString, "eStbMac"), re.StandardOperationRange, "50-75"))
		compound.AddCompoundPart(*CreateRule(re.RelationAnd, *re.NewFreeArg(re.StandardFreeArgTypeString, "eStbMac"), re.StandardOperationRange, "25-50"))

		ranges, err := getPercentRanges(compound)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(ranges))
		// Should be sorted by start range
		assert.Equal(t, float64(0), ranges[0].StartRange)
		assert.Equal(t, float64(25), ranges[1].StartRange)
		assert.Equal(t, float64(50), ranges[2].StartRange)
	})

	t.Run("NoRangeConditions", func(t *testing.T) {
		rule := makeRuleForService()
		ranges, err := getPercentRanges(rule)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(ranges))
	})

	t.Run("InvalidRangeFormat", func(t *testing.T) {
		rule := &re.Rule{
			Condition: CreateCondition(
				*re.NewFreeArg(re.StandardFreeArgTypeString, "eStbMac"),
				re.StandardOperationRange,
				"invalid",
			),
		}
		ranges, err := getPercentRanges(rule)
		assert.NotNil(t, err)
		assert.Nil(t, ranges)
	})

	t.Run("MixedConditions", func(t *testing.T) {
		compound := re.NewEmptyRule()
		compound.AddCompoundPart(*CreateRule("", *re.NewFreeArg(re.StandardFreeArgTypeString, "model"), re.StandardOperationIs, "X1"))
		compound.AddCompoundPart(*CreateRule(re.RelationAnd, *re.NewFreeArg(re.StandardFreeArgTypeString, "eStbMac"), re.StandardOperationRange, "10-50"))

		ranges, err := getPercentRanges(compound)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(ranges))
	})
}

// Test UpdateFeatureRule
func TestUpdateFeatureRule(t *testing.T) {
	cleanupServiceTest()

	f := makeFeatureForService("UpdateFeature", "stb")
	existingRule := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "UpdateRule")

	t.Run("EmptyId", func(t *testing.T) {
		fr := xwrfc.FeatureRule{
			Id:              "",
			Name:            "Test",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Rule:            makeRuleForService(),
		}
		result, err := UpdateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "id is empty")
	})

	t.Run("NonExistentRule", func(t *testing.T) {
		fr := xwrfc.FeatureRule{
			Id:              "nonexistent-id",
			Name:            "Test",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        1,
			Rule:            makeRuleForService(),
		}
		result, err := UpdateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("ChangeApplicationType", func(t *testing.T) {
		fr := xwrfc.FeatureRule{
			Id:              existingRule.Id,
			Name:            existingRule.Name,
			ApplicationType: "rdkcloud",
			FeatureIds:      []string{f.ID},
			Priority:        1,
			Rule:            makeRuleForService(),
		}
		result, err := UpdateFeatureRule(fr, "rdkcloud")
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ApplicationType cannot be changed")
	})

	t.Run("UpdateWithSamePriority", func(t *testing.T) {
		fr := xwrfc.FeatureRule{
			Id:              existingRule.Id,
			Name:            "UpdatedName",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        existingRule.Priority,
			Rule:            makeRuleForService(),
		}
		result, err := UpdateFeatureRule(fr, "stb")
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "UpdatedName", result.Name)
	})

	t.Run("UpdateWithDifferentPriority", func(t *testing.T) {
		// Create additional rules for priority testing
		fr2 := makeFeatureRuleForService([]string{f.ID}, "stb", 2, "Rule2")
		fr3 := makeFeatureRuleForService([]string{f.ID}, "stb", 3, "Rule3")

		// Update existingRule to priority 3
		fr := xwrfc.FeatureRule{
			Id:              existingRule.Id,
			Name:            existingRule.Name,
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        3,
			Rule:            makeRuleForService(),
		}
		result, err := UpdateFeatureRule(fr, "stb")
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, result.Priority)

		// Cleanup
		ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_FEATURE_CONTROL_RULE, fr2.Id)
		ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_FEATURE_CONTROL_RULE, fr3.Id)
	})

	t.Run("ValidationError", func(t *testing.T) {
		fr := xwrfc.FeatureRule{
			Id:              existingRule.Id,
			Name:            "", // Empty name should fail validation
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        1,
			Rule:            makeRuleForService(),
		}
		result, err := UpdateFeatureRule(fr, "stb")
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})
}

// Test updateFeatureRuleByPriorityAndReorganize
func TestUpdateFeatureRuleByPriorityAndReorganize(t *testing.T) {
	cleanupServiceTest()

	f := makeFeatureForService("Feature1", "stb")
	fr1 := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "Rule1")
	fr2 := makeFeatureRuleForService([]string{f.ID}, "stb", 2, "Rule2")
	fr3 := makeFeatureRuleForService([]string{f.ID}, "stb", 3, "Rule3")

	itemsList := []*xwrfc.FeatureRule{fr1, fr2, fr3}

	t.Run("UpdateExistingItem", func(t *testing.T) {
		updatedFr2 := &xwrfc.FeatureRule{
			Id:              fr2.Id,
			Name:            "UpdatedRule2",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        3,
			Rule:            makeRuleForService(),
		}
		result := updateFeatureRuleByPriorityAndReorganize(updatedFr2, itemsList, 2)
		assert.NotNil(t, result)
		// Verify the item was updated in the list
		found := false
		for _, item := range itemsList {
			if item.Id == updatedFr2.Id {
				assert.Equal(t, "UpdatedRule2", item.Name)
				found = true
			}
		}
		assert.True(t, found)
	})

	t.Run("AddNewItemToEmptyList", func(t *testing.T) {
		emptyList := []*xwrfc.FeatureRule{}
		newItem := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "NewRule")
		result := updateFeatureRuleByPriorityAndReorganize(newItem, emptyList, 1)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(emptyList))
	})

	t.Run("UpdateAndChangePriority", func(t *testing.T) {
		fr1.Priority = 1
		fr2.Priority = 2
		fr3.Priority = 3
		itemsList = []*xwrfc.FeatureRule{fr1, fr2, fr3}

		updatedFr1 := &xwrfc.FeatureRule{
			Id:              fr1.Id,
			Name:            fr1.Name,
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        3,
			Rule:            makeRuleForService(),
		}
		result := updateFeatureRuleByPriorityAndReorganize(updatedFr1, itemsList, 1)
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result))
	})
}

// Test importOrUpdateAllFeatureRule
func TestImportOrUpdateAllFeatureRule(t *testing.T) {
	cleanupServiceTest()

	f := makeFeatureForService("ImportFeature", "stb")
	existingRule := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "ExistingImportRule")

	t.Run("ImportNewRules", func(t *testing.T) {
		newRule1 := xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "NewImportRule1",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        2,
			Rule:            makeRuleForService(),
		}
		newRule2 := xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "NewImportRule2",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        3,
			Rule:            makeRuleForService(),
		}

		rules := []xwrfc.FeatureRule{newRule1, newRule2}
		result := importOrUpdateAllFeatureRule(rules, "stb")

		assert.Equal(t, 2, len(result[IMPORTED]))
		assert.Equal(t, 0, len(result[NOT_IMPORTED]))
	})

	t.Run("UpdateExistingRule", func(t *testing.T) {
		updatedRule := xwrfc.FeatureRule{
			Id:              existingRule.Id,
			Name:            "UpdatedImportRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        1,
			Rule:            makeRuleForService(),
		}

		rules := []xwrfc.FeatureRule{updatedRule}
		result := importOrUpdateAllFeatureRule(rules, "stb")

		assert.Equal(t, 1, len(result[IMPORTED]))
		assert.Equal(t, 0, len(result[NOT_IMPORTED]))
	})

	t.Run("MixedImport_SuccessAndFailure", func(t *testing.T) {
		validRule := xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "ValidImportRule",
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        5,
			Rule:            makeRuleForService(),
		}

		invalidRule := xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "", // Empty name will fail validation
			ApplicationType: "stb",
			FeatureIds:      []string{f.ID},
			Priority:        6,
			Rule:            makeRuleForService(),
		}

		rules := []xwrfc.FeatureRule{validRule, invalidRule}
		result := importOrUpdateAllFeatureRule(rules, "stb")

		assert.Equal(t, 1, len(result[IMPORTED]))
		assert.Equal(t, 1, len(result[NOT_IMPORTED]))
	})

	t.Run("ImportWithInvalidFeature", func(t *testing.T) {
		invalidRule := xwrfc.FeatureRule{
			Id:              uuid.New().String(),
			Name:            "InvalidFeatureRule",
			ApplicationType: "stb",
			FeatureIds:      []string{"nonexistent-feature-id"},
			Priority:        7,
			Rule:            makeRuleForService(),
		}

		rules := []xwrfc.FeatureRule{invalidRule}
		result := importOrUpdateAllFeatureRule(rules, "stb")

		assert.Equal(t, 0, len(result[IMPORTED]))
		assert.Equal(t, 1, len(result[NOT_IMPORTED]))
	})

	t.Run("ImportEmptyList", func(t *testing.T) {
		rules := []xwrfc.FeatureRule{}
		result := importOrUpdateAllFeatureRule(rules, "stb")

		assert.Equal(t, 0, len(result[IMPORTED]))
		assert.Equal(t, 0, len(result[NOT_IMPORTED]))
	})
}

// Test ChangeFeatureRulePriorities
func TestChangeFeatureRulePriorities(t *testing.T) {
	cleanupServiceTest()

	f := makeFeatureForService("PriorityFeature", "stb")
	fr1 := makeFeatureRuleForService([]string{f.ID}, "stb", 1, "PriorityRule1")
	fr2 := makeFeatureRuleForService([]string{f.ID}, "stb", 2, "PriorityRule2")
	fr3 := makeFeatureRuleForService([]string{f.ID}, "stb", 3, "PriorityRule3")

	t.Run("NonExistentRule", func(t *testing.T) {
		result, err := ChangeFeatureRulePriorities("nonexistent-id", 1, "stb")
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("ChangePriority_MoveDown", func(t *testing.T) {
		result, err := ChangeFeatureRulePriorities(fr1.Id, 3, "stb")
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.True(t, len(result) > 0)
	})

	t.Run("ChangePriority_MoveUp", func(t *testing.T) {
		// Reset priorities
		fr1.Priority = 1
		fr2.Priority = 2
		fr3.Priority = 3
		ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr1.Id, fr1)
		ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr2.Id, fr2)
		ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr3.Id, fr3)

		result, err := ChangeFeatureRulePriorities(fr3.Id, 1, "stb")
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.True(t, len(result) > 0)
	})

	t.Run("ChangePriority_WithApplicationType", func(t *testing.T) {
		// Reset
		fr1.Priority = 1
		fr2.Priority = 2
		fr3.Priority = 3
		ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr1.Id, fr1)
		ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr2.Id, fr2)
		ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr3.Id, fr3)

		result, err := ChangeFeatureRulePriorities(fr2.Id, 1, "stb")
		assert.Nil(t, err)
		assert.NotNil(t, result)
	})

	t.Run("ChangePriority_EmptyApplicationType", func(t *testing.T) {
		// Reset
		fr1.Priority = 1
		fr2.Priority = 2
		fr3.Priority = 3
		ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr1.Id, fr1)
		ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr2.Id, fr2)
		ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr3.Id, fr3)

		result, err := ChangeFeatureRulePriorities(fr2.Id, 3, "")
		assert.Nil(t, err)
		assert.NotNil(t, result)
	})
}
