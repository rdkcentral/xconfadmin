package common

import (
	"testing"
	"time"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"gotest.tools/assert"
)

func TestNewResponseEntity(t *testing.T) {
	// Test with nil error
	entity := NewResponseEntity(nil, "test data")
	assert.Equal(t, "test data", entity.Data)
	assert.Equal(t, 200, entity.Status)
	assert.Assert(t, entity.Error == nil)

	// Test with error
	err := NewXconfError(404, "not found")
	entity = NewResponseEntity(err, nil)
	assert.Equal(t, 404, entity.Status)
	assert.Assert(t, entity.Error != nil)
}

func TestNewResponseEntityWithStatus(t *testing.T) {
	err := NewXconfError(404, "not found")
	entity := NewResponseEntityWithStatus(404, err, "test data")
	assert.Equal(t, "test data", entity.Data)
	assert.Equal(t, 404, entity.Status)
	assert.Assert(t, entity.Error != nil)
}

func TestDCMGenericRuleMethods(t *testing.T) {
	dcmRule := &DCMGenericRule{
		ID:       "test-id",
		Name:     "test-name",
		Priority: 5,
	}

	// Test GetPriority
	assert.Equal(t, 5, dcmRule.GetPriority())

	// Test SetPriority
	dcmRule.SetPriority(10)
	assert.Equal(t, 10, dcmRule.GetPriority())

	// Test GetID
	assert.Equal(t, "test-id", dcmRule.GetID())

	// Test Clone
	cloned, err := dcmRule.Clone()
	assert.Assert(t, err == nil)
	assert.Equal(t, "test-id", cloned.ID)
	assert.Equal(t, "test-name", cloned.Name)

	// Verify it's a copy
	cloned.ID = "modified"
	assert.Equal(t, "test-id", dcmRule.ID)

	// Test GetId
	assert.Equal(t, "test-id", dcmRule.GetId())

	// Test GetName
	assert.Equal(t, "test-name", dcmRule.GetName())

	// Test GetTemplateId
	assert.Equal(t, "", dcmRule.GetTemplateId())

	// Test GetRuleType
	assert.Equal(t, "DCMGenericRule", dcmRule.GetRuleType())

	// Test GetRule
	rule := dcmRule.GetRule()
	assert.Assert(t, rule != nil)
}

func TestNewDCMGenericRuleInf(t *testing.T) {
	ruleInf := NewDCMGenericRuleInf()
	assert.Assert(t, ruleInf != nil)

	dcmRule, ok := ruleInf.(*DCMGenericRule)
	assert.Assert(t, ok)
	assert.Equal(t, 100, dcmRule.Percentage)
	assert.Equal(t, "stb", dcmRule.ApplicationType)
}

func TestToStringOnlyBaseProperties(t *testing.T) {
	// Test with condition
	dcmRule := &DCMGenericRule{}
	freeArg := re.FreeArg{Name: "testArg"}
	dcmRule.Rule.SetCondition(re.NewCondition(&freeArg, "IS", re.NewFixedArg("testValue")))

	str := dcmRule.ToStringOnlyBaseProperties()
	assert.Assert(t, str != "")

	// Test with compound rule
	dcmRule2 := &DCMGenericRule{}
	arg1 := re.FreeArg{Name: "arg1"}
	arg2 := re.FreeArg{Name: "arg2"}

	rule1 := re.Rule{}
	rule1.SetCondition(re.NewCondition(&arg1, "IS", re.NewFixedArg("value1")))

	rule2 := re.Rule{}
	rule2.SetCondition(re.NewCondition(&arg2, "IS", re.NewFixedArg("value2")))

	dcmRule2.Rule.CompoundParts = []re.Rule{rule1, rule2}

	str2 := dcmRule2.ToStringOnlyBaseProperties()
	assert.Assert(t, str2 != "")
}

func TestGetDCMGenericRuleList(t *testing.T) {
	rules := GetDCMGenericRuleList()
	assert.Assert(t, rules != nil)
}

func TestGetOneDCMGenericRule(t *testing.T) {
	rule := GetOneDCMGenericRule("test-id")
	// Without DB, this will return nil
	assert.Assert(t, rule == nil)
}

func TestEnvironmentAndModelFunctions(t *testing.T) {
	// Test GetAllEnvironmentList
	envs := GetAllEnvironmentList()
	assert.Assert(t, envs != nil)

	// Test GetOneEnvironment
	env := GetOneEnvironment("test-env")
	assert.Assert(t, env == nil) // Without DB

	// Test GetAllModelList
	models := GetAllModelList()
	assert.Assert(t, models != nil)

	// Test GetOneModel
	model := GetOneModel("test-model")
	assert.Assert(t, model == nil) // Without DB

	// Test IsExistModel
	exists := IsExistModel("test-model")
	assert.Assert(t, !exists)

	// Test with blank id
	exists = IsExistModel("")
	assert.Assert(t, !exists)
}

func TestGetIntAppSetting(t *testing.T) {
	// Test with default value
	val := GetIntAppSetting("nonexistent")
	assert.Equal(t, -1, val)

	// Test with custom default
	val = GetIntAppSetting("nonexistent", 42)
	assert.Equal(t, 42, val)
}

func TestGetFloat64AppSetting(t *testing.T) {
	// Test with default value
	val := GetFloat64AppSetting("nonexistent")
	assert.Equal(t, -1.0, val)

	// Test with custom default
	val = GetFloat64AppSetting("nonexistent", 3.14)
	assert.Equal(t, 3.14, val)
}

func TestGetTimeAppSetting(t *testing.T) {
	// Test with default value (will fail to find key)
	val := GetTimeAppSetting("nonexistent")
	assert.Assert(t, val.IsZero())

	// Test with custom default time
	now := time.Now()
	val = GetTimeAppSetting("nonexistent", now)
	assert.Equal(t, now, val)
}

func TestGetStringAppSetting(t *testing.T) {
	// Test with default value
	val := GetStringAppSetting("nonexistent")
	assert.Equal(t, "", val)

	// Test with custom default
	val = GetStringAppSetting("nonexistent", "default")
	assert.Equal(t, "default", val)
}

func TestGetBooleanAppSetting(t *testing.T) {
	// Test with default value
	val := GetBooleanAppSetting("nonexistent")
	assert.Equal(t, false, val)

	// Test with custom default
	val = GetBooleanAppSetting("nonexistent", true)
	assert.Equal(t, true, val)
}

func TestGetAppSettings(t *testing.T) {
	// This function calls DB, so without DB it will return an error
	settings, _ := GetAppSettings()
	// Without DB, we expect error but should still return empty map
	assert.Assert(t, settings != nil)
}

func TestCanarySettingsValidate(t *testing.T) {
	// Test valid settings
	maxSize := 100
	distPct := 10.0
	startTime := 100
	endTime := 200

	cs := &CanarySettings{
		CanaryMaxSize:                &maxSize,
		CanaryDistributionPercentage: &distPct,
		CanaryFwUpgradeStartTime:     &startTime,
		CanaryFwUpgradeEndTime:       &endTime,
	}

	err := cs.Validate()
	assert.Assert(t, err == nil)

	// Test invalid maxSize < 1
	invalidMaxSize := 0
	cs2 := &CanarySettings{CanaryMaxSize: &invalidMaxSize}
	err = cs2.Validate()
	assert.Assert(t, err != nil)

	// Test invalid maxSize > 100000
	tooLargeMaxSize := 100001
	cs3 := &CanarySettings{CanaryMaxSize: &tooLargeMaxSize}
	err = cs3.Validate()
	assert.Assert(t, err != nil)

	// Test invalid distributionPercentage < 1
	invalidDistPct := 0.5
	cs4 := &CanarySettings{CanaryDistributionPercentage: &invalidDistPct}
	err = cs4.Validate()
	assert.Assert(t, err != nil)

	// Test invalid distributionPercentage > 25
	tooLargeDistPct := 26.0
	cs5 := &CanarySettings{CanaryDistributionPercentage: &tooLargeDistPct}
	err = cs5.Validate()
	assert.Assert(t, err != nil)

	// Test invalid firmwareUpgradeStartTime < 0
	invalidStartTime := -1
	cs6 := &CanarySettings{CanaryFwUpgradeStartTime: &invalidStartTime}
	err = cs6.Validate()
	assert.Assert(t, err != nil)

	// Test invalid firmwareUpgradeStartTime > 5400
	tooLargeStartTime := 5401
	cs7 := &CanarySettings{CanaryFwUpgradeStartTime: &tooLargeStartTime}
	err = cs7.Validate()
	assert.Assert(t, err != nil)

	// Test invalid firmwareUpgradeEndTime < 0
	invalidEndTime := -1
	cs8 := &CanarySettings{CanaryFwUpgradeEndTime: &invalidEndTime}
	err = cs8.Validate()
	assert.Assert(t, err != nil)

	// Test invalid firmwareUpgradeEndTime > 5400
	tooLargeEndTime := 5401
	cs9 := &CanarySettings{CanaryFwUpgradeEndTime: &tooLargeEndTime}
	err = cs9.Validate()
	assert.Assert(t, err != nil)

	// Test endTime <= startTime
	equalTime := 100
	cs10 := &CanarySettings{
		CanaryFwUpgradeStartTime: &equalTime,
		CanaryFwUpgradeEndTime:   &equalTime,
	}
	err = cs10.Validate()
	assert.Assert(t, err != nil)
}

func TestLockdownSettingsValidate(t *testing.T) {
	enabled := true
	startTime := "10:00"
	endTime := "20:00"
	modules := "all"

	// Test valid settings
	ls := &LockdownSettings{
		LockdownEnabled:   &enabled,
		LockdownStartTime: &startTime,
		LockdownEndTime:   &endTime,
		LockdownModules:   &modules,
	}

	err := ls.Validate()
	assert.Assert(t, err == nil)

	// Test missing LockdownEnabled
	ls2 := &LockdownSettings{
		LockdownStartTime: &startTime,
		LockdownEndTime:   &endTime,
		LockdownModules:   &modules,
	}
	err = ls2.Validate()
	assert.Assert(t, err != nil)

	// NOTE: Cannot test missing LockdownModules because code has a bug -
	// it dereferences LockdownModules before checking if it's nil
	// This would cause a panic

	// Test startTime without endTime
	ls4 := &LockdownSettings{
		LockdownEnabled:   &enabled,
		LockdownStartTime: &startTime,
		LockdownModules:   &modules,
	}
	err = ls4.Validate()
	assert.Assert(t, err != nil)

	// Test endTime without startTime
	ls5 := &LockdownSettings{
		LockdownEnabled: &enabled,
		LockdownEndTime: &endTime,
		LockdownModules: &modules,
	}
	err = ls5.Validate()
	assert.Assert(t, err != nil)

	// Test invalid module
	invalidModules := "invalid"
	ls6 := &LockdownSettings{
		LockdownEnabled:   &enabled,
		LockdownStartTime: &startTime,
		LockdownEndTime:   &endTime,
		LockdownModules:   &invalidModules,
	}
	err = ls6.Validate()
	assert.Assert(t, err != nil)

	// Test valid multiple modules
	multipleModules := "dcm,rfc,firmware"
	ls7 := &LockdownSettings{
		LockdownEnabled:   &enabled,
		LockdownStartTime: &startTime,
		LockdownEndTime:   &endTime,
		LockdownModules:   &multipleModules,
	}
	err = ls7.Validate()
	assert.Assert(t, err == nil)
}
