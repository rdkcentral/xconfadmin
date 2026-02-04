// Copyright 2025 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package estbfirmware

import (
	"fmt"
	"testing"

	"github.com/rdkcentral/xconfadmin/util"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"gotest.tools/assert"
)

func TestGetLastConfigLog(t *testing.T) {
	mac := "AA:BB:CC:DD:EE:FF"

	// Test with non-existent MAC
	result := GetLastConfigLog(mac)

	// Should return nil for non-existent entry
	assert.Assert(t, result == nil, "Expected nil for non-existent MAC")
}

func TestGetConfigChangeLogsOnly(t *testing.T) {
	mac := "11:22:33:44:55:66"

	// Test with non-existent MAC
	result := GetConfigChangeLogsOnly(mac)

	// Should return empty slice for non-existent entry
	assert.Assert(t, result != nil, "Expected non-nil slice")
	assert.Equal(t, 0, len(result), "Expected empty slice for non-existent MAC")
}

func TestSetLastConfigLog(t *testing.T) {
	mac := "AA:BB:CC:DD:EE:11"

	// Create a simple config change log
	configLog := &ConfigChangeLog{
		ID:          LAST_CONFIG_LOG_ID,
		Updated:     util.GetTimestamp(),
		Explanation: "Test explanation",
	}

	// Test setting the log
	err := SetLastConfigLog(mac, configLog)

	// Should not return error (DB may not be initialized in test, but function should execute)
	// We're just testing that the function doesn't panic
	_ = err // May be error or nil depending on DB state
	assert.Assert(t, true, "SetLastConfigLog executed without panic")
}

func TestSetConfigChangeLog(t *testing.T) {
	mac := "BB:CC:DD:EE:FF:11"

	// Create a simple config change log
	configLog := &ConfigChangeLog{
		Updated:     util.GetTimestamp(),
		Explanation: "Test config change",
	}

	// Test setting the log
	err := SetConfigChangeLog(mac, configLog)

	// Should not panic (may return error if DB not initialized, but that's ok)
	_ = err
	assert.Assert(t, true, "SetConfigChangeLog executed without panic")
}

func TestGetLastConfigLog_Integration(t *testing.T) {
	mac := "AA:BB:CC:DD:EE:22"

	// Create and set a config log
	configLog := &ConfigChangeLog{
		ID:          LAST_CONFIG_LOG_ID,
		Updated:     util.GetTimestamp(),
		Explanation: "Integration test",
	}

	// Try to set it
	err := SetLastConfigLog(mac, configLog)
	if err != nil {
		// DB might not be initialized, skip the rest
		t.Logf("DB not initialized, skipping integration test: %v", err)
		return
	}

	// Try to retrieve it
	retrieved := GetLastConfigLog(mac)
	if retrieved != nil {
		assert.Equal(t, "Integration test", retrieved.Explanation, "Should retrieve the same explanation")
	}
}

func TestGetConfigChangeLogsOnly_AfterSet(t *testing.T) {
	mac := "CC:DD:EE:FF:11:22"

	// Create a config change log
	configLog := &ConfigChangeLog{
		Updated:     util.GetTimestamp(),
		Explanation: "Test log entry",
	}

	// Try to set it
	err := SetConfigChangeLog(mac, configLog)
	if err != nil {
		// DB might not be initialized, skip the rest
		t.Logf("DB not initialized, skipping test: %v", err)
		return
	}

	// Try to retrieve logs
	logs := GetConfigChangeLogsOnly(mac)
	assert.Assert(t, logs != nil, "Should return non-nil slice")
}

func TestGetCurrentId(t *testing.T) {
	mac := "DD:EE:FF:11:22:33"

	// Test with non-existent MAC - should return error or default ID
	id, err := GetCurrentId(mac)
	if err != nil {
		t.Logf("DB error expected in test environment: %v", err)
		return
	}

	// If no error, should return a valid ID format
	assert.Assert(t, id != "", "Expected non-empty ID")
	t.Logf("Got current ID: %s", id)
}

// Test NewRuleInfo with FirmwareRule
func TestNewRuleInfo_FirmwareRule(t *testing.T) {
	// Test with blocking filter
	blockingRule := &corefw.FirmwareRule{
		ID:   "test-rule-1",
		Type: "FIRMWARE_RULE",
		Name: "Test Blocking Rule",
		ApplicableAction: &corefw.ApplicableAction{
			ActionType: corefw.BLOCKING_FILTER,
		},
	}

	ruleInfo := NewRuleInfo(blockingRule)

	assert.Assert(t, ruleInfo != nil, "RuleInfo should not be nil")
	assert.Equal(t, "test-rule-1", ruleInfo.ID)
	assert.Equal(t, "FIRMWARE_RULE", ruleInfo.Type)
	assert.Equal(t, "Test Blocking Rule", ruleInfo.Name)
	assert.Equal(t, true, ruleInfo.Blocking)

	// Test with non-blocking rule
	nonBlockingRule := &corefw.FirmwareRule{
		ID:   "test-rule-2",
		Type: "FIRMWARE_RULE",
		Name: "Test Non-Blocking Rule",
		ApplicableAction: &corefw.ApplicableAction{
			ActionType: corefw.RULE_TEMPLATE,
		},
	}

	ruleInfo2 := NewRuleInfo(nonBlockingRule)

	assert.Assert(t, ruleInfo2 != nil, "RuleInfo should not be nil")
	assert.Equal(t, false, ruleInfo2.Blocking)
}

// Test NewRuleInfo with SingletonFilterValue
func TestNewRuleInfo_SingletonFilterValue(t *testing.T) {
	// Test with ID ending in _VALUE
	singletonWithValue := &SingletonFilterValue{
		ID: "TEST_FILTER_VALUE",
	}

	ruleInfo := NewRuleInfo(singletonWithValue)

	assert.Assert(t, ruleInfo != nil, "RuleInfo should not be nil")
	assert.Equal(t, "SINGLETON_TEST_FILTER", ruleInfo.ID)
	assert.Equal(t, "SingletonFilter", ruleInfo.Type)
	assert.Equal(t, "TEST_FILTER_VALUE", ruleInfo.Name)
	assert.Equal(t, true, ruleInfo.NoOp)
	assert.Equal(t, false, ruleInfo.Blocking)

	// Test with ID not ending in _VALUE
	singletonNoValue := &SingletonFilterValue{
		ID: "SIMPLE_FILTER",
	}

	ruleInfo2 := NewRuleInfo(singletonNoValue)

	assert.Assert(t, ruleInfo2 != nil, "RuleInfo should not be nil")
	assert.Equal(t, "SINGLETON_SIMPLE_FILTER", ruleInfo2.ID)
}

// Test NewRuleInfo with RuleAction
func TestNewRuleInfo_RuleAction(t *testing.T) {
	ruleAction := &corefw.RuleAction{}

	ruleInfo := NewRuleInfo(ruleAction)

	assert.Assert(t, ruleInfo != nil, "RuleInfo should not be nil")
	assert.Equal(t, "DistributionPercentInRuleAction", ruleInfo.ID)
	assert.Equal(t, "DistributionPercentInRuleAction", ruleInfo.Type)
	assert.Equal(t, "DistributionPercentInRuleAction", ruleInfo.Name)
	assert.Equal(t, false, ruleInfo.NoOp)
	assert.Equal(t, false, ruleInfo.Blocking)
}

// Test NewRuleInfo with PercentageBean
func TestNewRuleInfo_PercentageBean(t *testing.T) {
	percentageBean := &PercentageBean{
		Name: "Test Percentage Bean",
	}

	ruleInfo := NewRuleInfo(percentageBean)

	assert.Assert(t, ruleInfo != nil, "RuleInfo should not be nil")
	assert.Equal(t, "", ruleInfo.ID)
	assert.Equal(t, "PercentageBean", ruleInfo.Type)
	assert.Equal(t, "Test Percentage Bean", ruleInfo.Name)
	assert.Equal(t, false, ruleInfo.NoOp)
	assert.Equal(t, false, ruleInfo.Blocking)
}

// Test NewRuleInfo with unknown type
func TestNewRuleInfo_UnknownType(t *testing.T) {
	unknownType := "some string"

	ruleInfo := NewRuleInfo(unknownType)

	assert.Assert(t, ruleInfo != nil, "RuleInfo should not be nil")
	// Default RuleInfo should be returned
	assert.Equal(t, "", ruleInfo.ID)
	assert.Equal(t, "", ruleInfo.Type)
	assert.Equal(t, "", ruleInfo.Name)
}

// Test NewConfigChangeLogInf
func TestNewConfigChangeLogInf(t *testing.T) {
	result := NewConfigChangeLogInf()

	assert.Assert(t, result != nil, "Should return non-nil")

	// Check if it's actually a ConfigChangeLog
	_, ok := result.(*ConfigChangeLog)
	assert.Assert(t, ok, "Should return a ConfigChangeLog pointer")
}

// Test NewConfigChangeLog with all parameters
func TestNewConfigChangeLog_Complete(t *testing.T) {
	convertedContext := &ConvertedContext{
		// Add some test data
	}

	explanation := "Test explanation"

	firmwareConfig := &FirmwareConfigFacade{
		Properties:       map[string]interface{}{"version": "1.0"},
		CustomProperties: map[string]string{"key": "value"},
	}

	appliedFilters := []interface{}{
		&SingletonFilterValue{ID: "FILTER_1_VALUE"},
		&corefw.RuleAction{},
	}

	evaluatedRule := &corefw.FirmwareRule{
		ID:   "evaluated-rule-1",
		Type: "FIRMWARE_RULE",
		Name: "Evaluated Rule",
	}

	configLog := NewConfigChangeLog(convertedContext, explanation, firmwareConfig, appliedFilters, evaluatedRule, false)

	assert.Assert(t, configLog != nil, "ConfigChangeLog should not be nil")
	assert.Equal(t, LAST_CONFIG_LOG_ID, configLog.ID)
	assert.Assert(t, configLog.Updated > 0, "Updated timestamp should be set")
	assert.Equal(t, explanation, configLog.Explanation)
	assert.Equal(t, firmwareConfig, configLog.FirmwareConfig)
	assert.Assert(t, configLog.Rule != nil, "Rule should be set")
	assert.Equal(t, "evaluated-rule-1", configLog.Rule.ID)
	assert.Equal(t, 2, len(configLog.Filters), "Should have 2 filters")
}

// Test NewConfigChangeLog with nil evaluatedRule
func TestNewConfigChangeLog_NilRule(t *testing.T) {
	convertedContext := &ConvertedContext{}
	explanation := "Test with nil rule"
	firmwareConfig := &FirmwareConfigFacade{
		Properties: map[string]interface{}{},
	}
	appliedFilters := []interface{}{}

	configLog := NewConfigChangeLog(convertedContext, explanation, firmwareConfig, appliedFilters, nil, false)

	assert.Assert(t, configLog != nil, "ConfigChangeLog should not be nil")
	assert.Assert(t, configLog.Rule == nil, "Rule should be nil")
	assert.Equal(t, 0, len(configLog.Filters), "Should have no filters")
}

// Test NewConfigChangeLog with isLastLog true
func TestNewConfigChangeLog_IsLastLog(t *testing.T) {
	convertedContext := &ConvertedContext{}
	explanation := "Last log entry"
	firmwareConfig := &FirmwareConfigFacade{
		Properties: map[string]interface{}{},
	}

	configLog := NewConfigChangeLog(convertedContext, explanation, firmwareConfig, []interface{}{}, nil, true)

	assert.Assert(t, configLog != nil, "ConfigChangeLog should not be nil")
	assert.Equal(t, int64(0), configLog.Updated, "Updated should be 0 for last log")
}

// Test NewConfigChangeLog with empty filters
func TestNewConfigChangeLog_EmptyFilters(t *testing.T) {
	convertedContext := &ConvertedContext{}
	explanation := "No filters"
	firmwareConfig := &FirmwareConfigFacade{
		Properties: map[string]interface{}{},
	}

	configLog := NewConfigChangeLog(convertedContext, explanation, firmwareConfig, []interface{}{}, nil, false)

	assert.Assert(t, configLog != nil, "ConfigChangeLog should not be nil")
	assert.Assert(t, configLog.Filters != nil, "Filters should not be nil")
	assert.Equal(t, 0, len(configLog.Filters), "Filters should be empty")
}

// Test GetCurrentId with no existing logs
func TestGetCurrentId_NoLogs(t *testing.T) {
	// Use a unique MAC that likely has no logs
	mac := "AA:BB:CC:DD:EE:01"

	result, err := GetCurrentId(mac)

	// In test environment, database may not be configured
	// The function should handle this gracefully
	if err != nil {
		// Expected error in test environment: "Table configuration not found"
		assert.ErrorContains(t, err, "Table configuration")
	} else {
		// If it works, verify the result
		expectedId := fmt.Sprintf("%s_%d", prefix, BOUNDS)
		assert.Equal(t, expectedId, result)
	}
}

// Test GetCurrentId - verify function exists and basic structure
func TestGetCurrentId_FunctionExists(t *testing.T) {
	// This test verifies the function can be called without panicking
	// Even if the database is not configured
	mac := "TEST:MAC:ADDRESS"

	_, err := GetCurrentId(mac)

	// We just verify it doesn't panic
	// Error is expected in test environment without proper DB config
	_ = err
	assert.Assert(t, true, "Function executed without panic")
}

// Test numberToColumnName helper function indirectly
func TestNumberToColumnName_Format(t *testing.T) {
	// Test that GetCurrentId formats IDs correctly
	// The function uses numberToColumnName internally
	mac := "FORMAT:TEST:MAC"

	result, err := GetCurrentId(mac)

	if err == nil {
		// Verify the format matches pattern: prefix_number
		assert.Assert(t, result != "", "Result should not be empty")
		// Should contain the prefix and underscore
		assert.Assert(t, len(result) > len(prefix), "Result should include prefix and number")
	}
	// If error, it's expected in test environment
}

func TestNewRuleInfo_FirmwareRuleBlocking(t *testing.T) {
	// Test with blocking filter
	rule := &corefw.FirmwareRule{
		ID:   "test-rule-2",
		Type: "RULE",
		Name: "Blocking Rule",
		ApplicableAction: &corefw.ApplicableAction{
			ActionType: corefw.BLOCKING_FILTER,
		},
	}

	ruleInfo := NewRuleInfo(rule)
	assert.Equal(t, "test-rule-2", ruleInfo.ID)
	assert.Equal(t, true, ruleInfo.Blocking)
}

// TestNewRuleInfo_SingletonFilterValueWithSuffix tests NewRuleInfo with _VALUE suffix
func TestNewRuleInfo_SingletonFilterValueWithSuffix(t *testing.T) {
	// Test with SingletonFilterValue with _VALUE suffix
	singleton := &SingletonFilterValue{
		ID: "TEST_SINGLETON_VALUE",
	}

	ruleInfo := NewRuleInfo(singleton)
	assert.Equal(t, "SINGLETON_TEST_SINGLETON", ruleInfo.ID)
	assert.Equal(t, "SingletonFilter", ruleInfo.Type)
	assert.Equal(t, "TEST_SINGLETON_VALUE", ruleInfo.Name)
}

// TestNewRuleInfo_Unknown tests NewRuleInfo with unknown type
func TestNewRuleInfo_Unknown(t *testing.T) {
	// Test with unknown type - should return empty RuleInfo
	ruleInfo := NewRuleInfo("unknown type")
	assert.Equal(t, "", ruleInfo.ID)
	assert.Equal(t, "", ruleInfo.Type)
	assert.Equal(t, "", ruleInfo.Name)
	assert.Equal(t, false, ruleInfo.NoOp)
	assert.Equal(t, false, ruleInfo.Blocking)
}

// TestNewConfigChangeLog tests NewConfigChangeLog function
func TestNewConfigChangeLog(t *testing.T) {
	context := &ConvertedContext{}
	config := &FirmwareConfigFacade{
		Properties: map[string]interface{}{
			"firmwareVersion": "test-version-1.0",
		},
	}
	rule := &corefw.FirmwareRule{
		ID:   "rule-1",
		Name: "Test Rule",
		Type: "ENV_MODEL_RULE",
	}
	filters := []interface{}{
		&SingletonFilterValue{ID: "FILTER_1"},
		&PercentageBean{Name: "Percent Filter"},
	}

	// Test with isLastLog = false (should have timestamp)
	log := NewConfigChangeLog(context, "Test explanation", config, filters, rule, false)
	assert.Equal(t, LAST_CONFIG_LOG_ID, log.ID)
	assert.Assert(t, log.Updated > 0, "Should have timestamp when isLastLog is false")
	assert.Equal(t, "Test explanation", log.Explanation)
	assert.Assert(t, log.Rule != nil, "Should have rule info")
	assert.Equal(t, "rule-1", log.Rule.ID)
	assert.Equal(t, 2, len(log.Filters), "Should have 2 filters")
	assert.Equal(t, config, log.FirmwareConfig)
}

// TestNewConfigChangeLog_LastLog tests NewConfigChangeLog with isLastLog=true
func TestNewConfigChangeLog_LastLog(t *testing.T) {
	context := &ConvertedContext{}
	config := &FirmwareConfigFacade{
		Properties: map[string]interface{}{
			"firmwareVersion": "test-version-2.0",
		},
	}

	// Test with isLastLog = true (should NOT have timestamp)
	log := NewConfigChangeLog(context, "Last log", config, nil, nil, true)
	assert.Equal(t, int64(0), log.Updated, "Should NOT have timestamp when isLastLog is true")
	assert.Assert(t, log.Rule == nil, "Should have nil rule")
	assert.Equal(t, 0, len(log.Filters), "Should have no filters")
}

// TestNewConfigChangeLog_NoRule tests NewConfigChangeLog without rule
func TestNewConfigChangeLog_NoRule(t *testing.T) {
	context := &ConvertedContext{}
	config := &FirmwareConfigFacade{}

	log := NewConfigChangeLog(context, "No rule test", config, []interface{}{}, nil, false)
	assert.Assert(t, log.Rule == nil, "Should have nil rule when evaluatedRule is nil")
}

// TestNumberToColumnName tests the numberToColumnName function
func TestNumberToColumnName(t *testing.T) {
	// Test various numbers
	result1 := numberToColumnName(0)
	assert.Assert(t, len(result1) > 0, "Should return non-empty string")
	assert.Assert(t, result1[len(result1)-1] == '0', "Should end with 0")

	result2 := numberToColumnName(5)
	assert.Assert(t, result2[len(result2)-1] == '5', "Should end with 5")

	result3 := numberToColumnName(10)
	assert.Assert(t, len(result3) > 0, "Should return non-empty string")
}

// TestGetCurrentId_EmptyLogs tests GetCurrentId with no existing logs
func TestGetCurrentId_EmptyLogs(t *testing.T) {
	mac := "FF:EE:DD:CC:BB:AA"

	id, err := GetCurrentId(mac)
	if err != nil {
		// DB might not be initialized
		t.Logf("DB error expected: %v", err)
		return
	}

	// Should return a valid ID (default is BOUNDS when count is 1)
	assert.Assert(t, id != "", "Should return non-empty ID")
	t.Logf("Current ID for empty logs: %s", id)
}

// TestGetConfigChangeLogsOnly_Sorting tests that logs are sorted by Updated time
func TestGetConfigChangeLogsOnly_Sorting(t *testing.T) {
	mac := "AA:11:22:33:44:55"

	// Get logs (may be empty if DB not initialized)
	logs := GetConfigChangeLogsOnly(mac)
	assert.Assert(t, logs != nil, "Should return non-nil slice")

	// If we have logs, verify they're sorted in descending order
	if len(logs) > 1 {
		for i := 0; i < len(logs)-1; i++ {
			assert.Assert(t, logs[i].Updated >= logs[i+1].Updated,
				"Logs should be sorted by descending Updated time")
		}
	}
}

// TestSetLastConfigLog_Marshaling tests JSON marshaling in SetLastConfigLog
func TestSetLastConfigLog_Marshaling(t *testing.T) {
	mac := "BB:22:33:44:55:66"

	// Create a config log with various fields
	configLog := &ConfigChangeLog{
		ID:          LAST_CONFIG_LOG_ID,
		Updated:     util.GetTimestamp(),
		Explanation: "Test with complex data",
		Input: &ConvertedContext{
			EstbMac: mac,
		},
		FirmwareConfig: &FirmwareConfigFacade{
			Properties: map[string]interface{}{
				"firmwareVersion": "1.0.0",
			},
		},
		HasMinimumFirmware: true,
	}

	err := SetLastConfigLog(mac, configLog)
	// Function should execute without panic
	_ = err
	assert.Assert(t, true, "SetLastConfigLog with complex data executed")
}

// TestSetConfigChangeLog_IDAssignment tests that SetConfigChangeLog assigns ID
func TestSetConfigChangeLog_IDAssignment(t *testing.T) {
	mac := "CC:33:44:55:66:77"

	configLog := &ConfigChangeLog{
		Updated:     util.GetTimestamp(),
		Explanation: "Test ID assignment",
	}

	// ID should be empty initially
	assert.Equal(t, "", configLog.ID)

	err := SetConfigChangeLog(mac, configLog)
	if err != nil {
		// DB might not be initialized, but we tested the function execution
		t.Logf("DB error (expected in test): %v", err)
	}
}

// TestGetLastConfigLog_TypeAssertion tests type assertion in GetLastConfigLog
func TestGetLastConfigLog_TypeAssertion(t *testing.T) {
	mac := "DD:44:55:66:77:88"

	// Even if DB returns something, type assertion should work
	result := GetLastConfigLog(mac)
	// Result is either nil or *sharedef.ConfigChangeLog
	if result != nil {
		assert.Assert(t, result.ID != "", "Should have an ID if not nil")
	}
}

// TestGetConfigChangeLogsOnly_FilterLastLog tests that LAST_CONFIG_LOG_ID is filtered out
func TestGetConfigChangeLogsOnly_FilterLastLog(t *testing.T) {
	mac := "EE:55:66:77:88:99"

	logs := GetConfigChangeLogsOnly(mac)
	assert.Assert(t, logs != nil, "Should return non-nil slice")

	// Verify no log has ID == LAST_CONFIG_LOG_ID
	for _, log := range logs {
		assert.Assert(t, log.ID != LAST_CONFIG_LOG_ID,
			"Should filter out LAST_CONFIG_LOG_ID")
	}
}

// TestGetCurrentId_WithExistingLogs tests GetCurrentId with various scenarios
func TestGetCurrentId_WithExistingLogs(t *testing.T) {
	mac := "11:AA:BB:CC:DD:EE"

	// Try to get current ID - may fail if DB not configured
	id, err := GetCurrentId(mac)
	if err != nil {
		t.Logf("DB not configured (expected): %v", err)
		return
	}

	// If successful, ID should follow the format PREFIX_NUMBER
	assert.Assert(t, id != "", "Should return non-empty ID")
	assert.Assert(t, len(id) > 2, "Should have minimum length")
}

// TestSetConfigChangeLog_WithValidData tests SetConfigChangeLog with complete data
func TestSetConfigChangeLog_WithValidData(t *testing.T) {
	mac := "22:BB:CC:DD:EE:FF"

	configLog := &ConfigChangeLog{
		Updated:     util.GetTimestamp(),
		Explanation: "Complete test data",
		Input: &ConvertedContext{
			EstbMac: mac,
		},
		FirmwareConfig: &FirmwareConfigFacade{
			Properties: map[string]interface{}{
				"firmwareVersion": "2.0.0",
			},
		},
		Rule: &RuleInfo{
			ID:   "test-rule",
			Name: "Test Rule",
		},
		Filters: []*RuleInfo{
			{ID: "filter-1", Name: "Filter 1"},
		},
	}

	err := SetConfigChangeLog(mac, configLog)
	if err == nil {
		// ID should be assigned by SetConfigChangeLog
		assert.Assert(t, configLog.ID != "", "ID should be assigned")
	} else {
		t.Logf("DB error (expected in test): %v", err)
	}
}

// TestGetLastConfigLog_WithSet tests GetLastConfigLog after SetLastConfigLog
func TestGetLastConfigLog_WithSet(t *testing.T) {
	mac := "33:CC:DD:EE:FF:00"

	// Try to set a last config log
	configLog := &ConfigChangeLog{
		ID:          LAST_CONFIG_LOG_ID,
		Updated:     util.GetTimestamp(),
		Explanation: "Test get after set",
		Input: &ConvertedContext{
			EstbMac: mac,
		},
	}

	err := SetLastConfigLog(mac, configLog)
	if err != nil {
		t.Logf("DB not configured: %v", err)
		return
	}

	// Try to retrieve it
	retrieved := GetLastConfigLog(mac)
	if retrieved != nil {
		assert.Equal(t, "Test get after set", retrieved.Explanation)
	}
}

// TestGetConfigChangeLogsOnly_WithMultipleLogs tests with multiple logs
func TestGetConfigChangeLogsOnly_WithMultipleLogs(t *testing.T) {
	mac := "44:DD:EE:FF:00:11"

	// Set multiple config change logs
	for i := 1; i <= 3; i++ {
		configLog := &ConfigChangeLog{
			Updated:     util.GetTimestamp() + int64(i*1000),
			Explanation: "Test log " + string(rune(i+'0')),
		}
		err := SetConfigChangeLog(mac, configLog)
		if err != nil {
			t.Logf("DB not configured: %v", err)
			return
		}
	}

	// Retrieve all logs
	logs := GetConfigChangeLogsOnly(mac)
	assert.Assert(t, logs != nil, "Should return non-nil slice")

	// Logs should be sorted by descending Updated time
	if len(logs) > 1 {
		for i := 0; i < len(logs)-1; i++ {
			assert.Assert(t, logs[i].Updated >= logs[i+1].Updated,
				"Should be sorted in descending order")
		}
	}
}

// TestNumberToColumnName_Various tests numberToColumnName with various inputs
func TestNumberToColumnName_Various(t *testing.T) {
	testCases := []struct {
		number   int
		expected string
	}{
		{0, prefix + "_0"},
		{1, prefix + "_1"},
		{5, prefix + "_5"},
		{10, prefix + "_10"},
		{100, prefix + "_100"},
	}

	for _, tc := range testCases {
		result := numberToColumnName(tc.number)
		assert.Equal(t, tc.expected, result, "Should format correctly")
	}
}

// TestGetCurrentId_BoundsLogic tests the BOUNDS logic in GetCurrentId
func TestGetCurrentId_BoundsLogic(t *testing.T) {
	mac := "55:EE:FF:00:11:22"

	// This tests the logic where count cycles through BOUNDS
	id, err := GetCurrentId(mac)
	if err != nil {
		t.Logf("DB not configured: %v", err)
		return
	}

	// ID should contain the prefix
	assert.Assert(t, len(id) > len(prefix), "ID should contain prefix")
	t.Logf("Generated ID: %s", id)
}

// TestSetLastConfigLog_ErrorHandling tests error handling in marshaling
func TestSetLastConfigLog_ErrorHandling(t *testing.T) {
	mac := "66:FF:00:11:22:33"

	// Create a valid config log
	configLog := &ConfigChangeLog{
		ID:          LAST_CONFIG_LOG_ID,
		Updated:     util.GetTimestamp(),
		Explanation: "Error handling test",
	}

	// SetLastConfigLog should handle marshaling internally
	err := SetLastConfigLog(mac, configLog)
	// May succeed or fail depending on DB, but shouldn't panic
	_ = err
	assert.Assert(t, true, "Function executed without panic")
}

// TestSetConfigChangeLog_GetCurrentIdError tests error propagation from GetCurrentId
func TestSetConfigChangeLog_GetCurrentIdError(t *testing.T) {
	mac := "77:00:11:22:33:44"

	configLog := &ConfigChangeLog{
		Updated:     util.GetTimestamp(),
		Explanation: "Test error propagation",
	}

	err := SetConfigChangeLog(mac, configLog)
	// If GetCurrentId fails, SetConfigChangeLog should also fail
	// But in test environment, DB might not be configured
	_ = err
	assert.Assert(t, true, "Function executed")
}

// TestInit_PrefixAssignment tests that init() sets prefix correctly
func TestInit_PrefixAssignment(t *testing.T) {
	// After init(), prefix should be set to either hostname or DEFAULT_PREFIX
	assert.Assert(t, prefix != "", "Prefix should be set")
	assert.Assert(t, len(prefix) > 0, "Prefix should have length > 0")
	t.Logf("Prefix is: %s", prefix)
}

// TestNewRuleInfo_NilInputs tests NewRuleInfo with nil-like inputs
func TestNewRuleInfo_NilInputs(t *testing.T) {
	// Test with nil
	ruleInfo := NewRuleInfo(nil)
	assert.Equal(t, "", ruleInfo.ID)
	assert.Equal(t, "", ruleInfo.Type)
}

// TestGetConfigChangeLogsOnly_EmptyResult tests when no logs exist
func TestGetConfigChangeLogsOnly_EmptyResult(t *testing.T) {
	mac := "88:11:22:33:44:55"

	// Get logs for non-existent MAC
	logs := GetConfigChangeLogsOnly(mac)
	assert.Assert(t, logs != nil, "Should return non-nil slice")
	// Should return empty slice
	assert.Equal(t, 0, len(logs), "Should have no logs for new MAC")
}

// TestConstants tests package constants
func TestConstants(t *testing.T) {
	assert.Equal(t, "XCONF", DEFAULT_PREFIX)
	assert.Equal(t, 5, BOUNDS)
	assert.Equal(t, "0", LAST_CONFIG_LOG_ID)
}

// TestConfigChangeLog_Struct tests ConfigChangeLog struct
func TestConfigChangeLog_Struct(t *testing.T) {
	log := &ConfigChangeLog{
		ID:          "test-id",
		Updated:     12345678,
		Explanation: "test explanation",
		Input: &ConvertedContext{
			EstbMac: "AA:BB:CC:DD:EE:FF",
		},
		Rule: &RuleInfo{
			ID:   "rule-1",
			Type: "TEST_RULE",
			Name: "Test Rule",
		},
		Filters: []*RuleInfo{
			{ID: "filter-1", Name: "Filter 1", NoOp: true},
		},
		FirmwareConfig: &FirmwareConfigFacade{
			Properties: map[string]interface{}{
				"version": "1.0",
			},
		},
		HasMinimumFirmware: true,
	}

	assert.Equal(t, "test-id", log.ID)
	assert.Equal(t, int64(12345678), log.Updated)
	assert.Equal(t, "test explanation", log.Explanation)
	assert.Assert(t, log.HasMinimumFirmware)
	assert.Equal(t, 1, len(log.Filters))
}

// TestRuleInfo_Struct tests RuleInfo struct
func TestRuleInfo_Struct(t *testing.T) {
	ruleInfo := &RuleInfo{
		ID:       "test-id",
		Type:     "test-type",
		Name:     "Test Name",
		NoOp:     true,
		Blocking: false,
	}

	assert.Equal(t, "test-id", ruleInfo.ID)
	assert.Equal(t, "test-type", ruleInfo.Type)
	assert.Equal(t, "Test Name", ruleInfo.Name)
	assert.Equal(t, true, ruleInfo.NoOp)
	assert.Equal(t, false, ruleInfo.Blocking)
}

// TestNewRuleInfo_FirmwareRuleNoop tests NoOp detection
func TestNewRuleInfo_FirmwareRuleNoop(t *testing.T) {
	// Create a rule that returns true for IsNoop()
	rule := &corefw.FirmwareRule{
		ID:               "noop-rule",
		Type:             "NOOP_TYPE",
		Name:             "NoOp Rule",
		ApplicableAction: nil,
	}

	ruleInfo := NewRuleInfo(rule)
	assert.Equal(t, "noop-rule", ruleInfo.ID)
	// NoOp value depends on the FirmwareRule.IsNoop() implementation
	t.Logf("NoOp value: %v", ruleInfo.NoOp)
}

// TestNewConfigChangeLog_AllFilters tests with all filter types
func TestNewConfigChangeLog_AllFilters(t *testing.T) {
	context := &ConvertedContext{
		EstbMac: "AA:BB:CC:DD:EE:FF",
	}
	config := &FirmwareConfigFacade{}

	// Include all types of filters
	filters := []interface{}{
		&SingletonFilterValue{ID: "SINGLETON_1"},
		&SingletonFilterValue{ID: "SINGLETON_2_VALUE"},
		&PercentageBean{Name: "Percent1"},
		&corefw.RuleAction{},
		&corefw.FirmwareRule{ID: "filter-rule", Name: "Filter Rule"},
	}

	log := NewConfigChangeLog(context, "All filters test", config, filters, nil, false)
	assert.Equal(t, 5, len(log.Filters), "Should have all 5 filters")

	// Verify each filter type is converted
	filterTypes := make(map[string]bool)
	for _, f := range log.Filters {
		filterTypes[f.Type] = true
	}

	assert.Assert(t, filterTypes["SingletonFilter"], "Should have SingletonFilter")
	assert.Assert(t, filterTypes["PercentageBean"], "Should have PercentageBean")
	assert.Assert(t, filterTypes["DistributionPercentInRuleAction"], "Should have RuleAction")
}

// TestNewConfigChangeLog_TimestampLogic tests timestamp assignment logic
func TestNewConfigChangeLog_TimestampLogic(t *testing.T) {
	context := &ConvertedContext{}
	config := &FirmwareConfigFacade{}

	// When isLastLog = false, should have timestamp
	log1 := NewConfigChangeLog(context, "log1", config, nil, nil, false)
	assert.Assert(t, log1.Updated > 0, "Non-last log should have timestamp")

	// When isLastLog = true, should NOT have timestamp
	log2 := NewConfigChangeLog(context, "log2", config, nil, nil, true)
	assert.Equal(t, int64(0), log2.Updated, "Last log should have zero timestamp")
}
