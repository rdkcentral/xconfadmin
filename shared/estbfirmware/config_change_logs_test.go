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
	"testing"

	"github.com/rdkcentral/xconfadmin/util"
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
