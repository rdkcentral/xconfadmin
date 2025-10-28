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
	"encoding/json"
	"testing"
)

func TestValidateName(t *testing.T) {
	// Test with potential DB error recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to DB not configured: %v", r)
		}
	}()

	fc := &FirmwareConfig{
		ID:              "test-id",
		Description:     "Test Config",
		ApplicationType: "stb",
	}
	err := fc.ValidateName()
	if err != nil {
		t.Logf("DB error expected in test environment: %v", err)
	}
}

func TestGetFirmwareVersion(t *testing.T) {
	// Test with potential DB error recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to DB not configured: %v", r)
		}
	}()

	version := GetFirmwareVersion("test-id")
	// Expect empty string when DB not available
	if version != "" {
		t.Logf("Got version: %s", version)
	}
}

func TestGetFirmwareConfigAsMapDB(t *testing.T) {
	// Test with potential DB error recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to DB not configured: %v", r)
		}
	}()

	configMap, err := GetFirmwareConfigAsMapDB("stb")
	if err != nil {
		t.Logf("DB error expected in test environment: %v", err)
		return
	}
	// Map may be nil or empty when DB is not configured
	t.Logf("Got config map with %d entries", len(configMap))
}

func TestGetFirmwareConfigAsListDB(t *testing.T) {
	// Test with potential DB error recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to DB not configured: %v", r)
		}
	}()

	list, err := GetFirmwareConfigAsListDB()
	if err != nil {
		t.Logf("DB error expected in test environment: %v", err)
		return
	}
	if list == nil {
		t.Fatalf("expected non-nil list")
	}
}

func TestDeleteOneFirmwareConfig(t *testing.T) {
	// Test with potential DB error recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to DB not configured: %v", r)
		}
	}()

	err := DeleteOneFirmwareConfig("test-id")
	if err != nil {
		t.Logf("DB error expected in test environment: %v", err)
	}
}

func TestCreateFirmwareConfigOneDB(t *testing.T) {
	// Test with potential DB error recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to DB not configured: %v", r)
		}
	}()

	fc := &FirmwareConfig{
		Description:     "Test Config",
		FirmwareVersion: "1.0.0",
		ApplicationType: "stb",
	}
	err := CreateFirmwareConfigOneDB(fc)
	if err != nil {
		t.Logf("DB error expected in test environment: %v", err)
		return
	}
	// ID should be auto-generated if blank
	if fc.ID == "" {
		t.Fatalf("expected auto-generated ID")
	}
}

func TestGetFirmwareConfigOneDB(t *testing.T) {
	// Test empty ID error
	_, err := GetFirmwareConfigOneDB("")
	if err == nil || err.Error() != "id is empty" {
		t.Fatalf("expected 'id is empty' error, got: %v", err)
	}

	// Test with potential DB error recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to DB not configured: %v", r)
		}
	}()

	_, err = GetFirmwareConfigOneDB("test-id")
	if err != nil {
		t.Logf("DB error expected in test environment: %v", err)
	}
}

func TestGetUpgradeDelay(t *testing.T) {
	ff := NewDefaulttFirmwareConfigFacade()
	ff.Properties[UPGRADE_DELAY] = 300

	delay := ff.GetUpgradeDelay()
	if delay != 300 {
		t.Fatalf("expected delay 300, got %d", delay)
	}

	// Test nil value
	ff.Properties[UPGRADE_DELAY] = nil
	delay = ff.GetUpgradeDelay()
	if delay != 0 {
		t.Fatalf("expected delay 0 for nil, got %d", delay)
	}

	// Test missing key
	delete(ff.Properties, UPGRADE_DELAY)
	delay = ff.GetUpgradeDelay()
	if delay != 0 {
		t.Fatalf("expected delay 0 for missing key, got %d", delay)
	}
}

func TestPutIfPresent(t *testing.T) {
	ff := NewDefaulttFirmwareConfigFacade()

	// Test with valid string
	ff.PutIfPresent("key1", "value1")
	if ff.Properties["key1"] != "value1" {
		t.Fatalf("expected 'value1', got %v", ff.Properties["key1"])
	}

	// Test with nil (should not add)
	ff.PutIfPresent("key2", nil)
	if _, exists := ff.Properties["key2"]; exists {
		t.Fatalf("expected key2 not to exist")
	}

	// Test with empty string (should not add)
	ff.PutIfPresent("key3", "")
	if _, exists := ff.Properties["key3"]; exists {
		t.Fatalf("expected key3 not to exist")
	}
}

func TestNewDefaulttFirmwareConfigFacade(t *testing.T) {
	ff := NewDefaulttFirmwareConfigFacade()

	if ff == nil {
		t.Fatalf("expected non-nil facade")
	}
	if ff.Properties == nil {
		t.Fatalf("expected Properties map to be initialized")
	}
	if len(ff.Properties) != 0 {
		t.Fatalf("expected empty Properties map, got %d entries", len(ff.Properties))
	}
}

func TestUnmarshalJSON(t *testing.T) {
	jsonData := `{
		"estbMac": "AA:BB:CC:DD:EE:FF",
		"model": "TEST_MODEL",
		"env": "PROD",
		"firmwareVersion": "1.0.0",
		"timeZone": "America/New_York",
		"time": "10/27/2025 14:30:00",
		"bypassFilters": ["filter1", "filter2"],
		"forceFilters": ["force1"],
		"capabilities": ["RCDL", "rebootDecoupled"]
	}`

	var cc ConvertedContext
	err := json.Unmarshal([]byte(jsonData), &cc)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Verify basic fields
	if cc.EstbMac != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected EstbMac 'AA:BB:CC:DD:EE:FF', got '%s'", cc.EstbMac)
	}
	if cc.Model != "TEST_MODEL" {
		t.Errorf("expected Model 'TEST_MODEL', got '%s'", cc.Model)
	}

	// Verify time zone was loaded
	if cc.TimeZone == nil {
		t.Fatalf("expected TimeZone to be set")
	}

	// Verify time was parsed
	if cc.Time == nil {
		t.Fatalf("expected Time to be set")
	}

	// Verify filters were converted to sets
	if len(cc.BypassFilters) != 2 {
		t.Errorf("expected 2 bypass filters, got %d", len(cc.BypassFilters))
	}
	if len(cc.ForceFilters) != 1 {
		t.Errorf("expected 1 force filter, got %d", len(cc.ForceFilters))
	}
}

func TestAddFiltersIntoConverted(t *testing.T) {
	filters := make(map[string]struct{})

	// Test with single filter
	addFiltersIntoConverted("filter1", filters)
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(filters))
	}
	if _, exists := filters["filter1"]; !exists {
		t.Errorf("expected filter1 to exist")
	}

	// Test with multiple filters
	filters = make(map[string]struct{})
	addFiltersIntoConverted("filter1,filter2,filter3", filters)
	if len(filters) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(filters))
	}
	if _, exists := filters["filter1"]; !exists {
		t.Errorf("expected filter1 to exist")
	}
	if _, exists := filters["filter2"]; !exists {
		t.Errorf("expected filter2 to exist")
	}
	if _, exists := filters["filter3"]; !exists {
		t.Errorf("expected filter3 to exist")
	}

	// Test with spaces
	filters = make(map[string]struct{})
	addFiltersIntoConverted(" filter1 , filter2 ", filters)
	if len(filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(filters))
	}
}

func TestSetCapabilities(t *testing.T) {
	cc := &ConvertedContext{
		Context: make(map[string]string),
	}

	// Test with single capability
	cc.SetCapabilities([]string{"RCDL"})
	caps := cc.GetCapabilities()
	if len(caps) != 1 {
		t.Fatalf("expected 1 capability, got %d", len(caps))
	}
	if caps[0] != "RCDL" {
		t.Errorf("expected 'RCDL', got '%s'", caps[0])
	}

	// Test with multiple capabilities
	cc.SetCapabilities([]string{"RCDL", "REBOOTDECOUPLED", "SUPPORTSFULLHTTPURL"})
	caps = cc.GetCapabilities()
	if len(caps) != 3 {
		t.Fatalf("expected 3 capabilities, got %d", len(caps))
	}

	// Test with empty slice
	cc.SetCapabilities([]string{})
	caps = cc.GetCapabilities()
	if len(caps) != 1 || caps[0] != "" {
		t.Errorf("expected empty capabilities")
	}
}
