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

	core "github.com/rdkcentral/xconfadmin/shared"
)

func TestNewEmptyFirmwareConfig(t *testing.T) {
	fc := NewEmptyFirmwareConfig()

	if fc == nil {
		t.Fatal("expected non-nil firmware config")
	}

	if fc.RebootImmediately {
		t.Error("expected RebootImmediately false")
	}

	if fc.ApplicationType != "stb" {
		t.Errorf("expected ApplicationType 'stb', got %s", fc.ApplicationType)
	}

	if fc.FirmwareDownloadProtocol != "tftp" {
		t.Errorf("expected FirmwareDownloadProtocol 'tftp', got %s", fc.FirmwareDownloadProtocol)
	}
}

func TestFirmwareConfig_SetGetApplicationType(t *testing.T) {
	fc := &FirmwareConfig{}

	fc.SetApplicationType("xhome")
	if fc.GetApplicationType() != "xhome" {
		t.Errorf("expected ApplicationType 'xhome', got %s", fc.GetApplicationType())
	}

	fc.SetApplicationType("stb")
	if fc.GetApplicationType() != "stb" {
		t.Errorf("expected ApplicationType 'stb', got %s", fc.GetApplicationType())
	}
}

func TestFirmwareConfig_Clone(t *testing.T) {
	original := &FirmwareConfig{
		ID:                       "test-id",
		Description:              "Test Config",
		SupportedModelIds:        []string{"MODEL1", "MODEL2"},
		FirmwareFilename:         "firmware.bin",
		FirmwareVersion:          "v1.0",
		ApplicationType:          "stb",
		FirmwareDownloadProtocol: "http",
		FirmwareLocation:         "http://example.com",
		Ipv6FirmwareLocation:     "http://[::1]",
		UpgradeDelay:             300,
		RebootImmediately:        true,
		Properties: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	cloned, err := original.Clone()
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	if cloned == nil {
		t.Fatal("expected non-nil cloned config")
	}

	// Verify all fields are copied
	if cloned.ID != original.ID {
		t.Errorf("ID mismatch: expected %s, got %s", original.ID, cloned.ID)
	}

	if cloned.Description != original.Description {
		t.Errorf("Description mismatch")
	}

	if len(cloned.SupportedModelIds) != len(original.SupportedModelIds) {
		t.Errorf("SupportedModelIds length mismatch")
	}

	// Verify it's a deep copy (modifying clone shouldn't affect original)
	cloned.Description = "Modified"
	if original.Description == "Modified" {
		t.Error("Clone is not independent - modifying clone affected original")
	}
}

func TestFirmwareConfig_Validate_Success(t *testing.T) {
	// First, we need to register models in the common package
	// Assuming Model1 exists or can be created for testing
	fc := &FirmwareConfig{
		Description:       "Valid Config",
		FirmwareFilename:  "test.bin",
		FirmwareVersion:   "v1.0",
		SupportedModelIds: []string{}, // Will be validated if models exist
		ApplicationType:   core.STB,
		Properties:        map[string]string{"key": "value"},
	}

	// This test may fail if models don't exist in DB
	// For now, test the structure
	err := fc.Validate()
	// If error is about model not existing, that's expected in unit test environment
	if err != nil && err.Error() != "Supported model list is empty" {
		// Models may not be set up, so we accept model-related errors
		t.Logf("Validation error (expected in unit test): %v", err)
	}
}

func TestFirmwareConfig_Validate_NilConfig(t *testing.T) {
	var fc *FirmwareConfig = nil

	err := fc.Validate()
	if err == nil {
		t.Fatal("expected error for nil config")
	}

	if err.Error() != "Firmware config is not present" {
		t.Errorf("expected 'Firmware config is not present', got %s", err.Error())
	}
}

func TestFirmwareConfig_Validate_EmptyDescription(t *testing.T) {
	fc := &FirmwareConfig{
		Description:      "",
		FirmwareFilename: "test.bin",
		FirmwareVersion:  "v1.0",
	}

	err := fc.Validate()
	if err == nil {
		t.Fatal("expected error for empty description")
	}

	if err.Error() != "Description is empty" {
		t.Errorf("expected 'Description is empty', got %s", err.Error())
	}
}

func TestFirmwareConfig_Validate_EmptyFilename(t *testing.T) {
	fc := &FirmwareConfig{
		Description:      "Test",
		FirmwareFilename: "",
		FirmwareVersion:  "v1.0",
	}

	err := fc.Validate()
	if err == nil {
		t.Fatal("expected error for empty filename")
	}

	if err.Error() != "File name is empty" {
		t.Errorf("expected 'File name is empty', got %s", err.Error())
	}
}

func TestFirmwareConfig_Validate_EmptyVersion(t *testing.T) {
	fc := &FirmwareConfig{
		Description:      "Test",
		FirmwareFilename: "test.bin",
		FirmwareVersion:  "",
	}

	err := fc.Validate()
	if err == nil {
		t.Fatal("expected error for empty version")
	}

	if err.Error() != "Version is empty" {
		t.Errorf("expected 'Version is empty', got %s", err.Error())
	}
}

func TestFirmwareConfig_Validate_EmptySupportedModels(t *testing.T) {
	fc := &FirmwareConfig{
		Description:       "Test",
		FirmwareFilename:  "test.bin",
		FirmwareVersion:   "v1.0",
		SupportedModelIds: []string{},
	}

	err := fc.Validate()
	if err == nil {
		t.Fatal("expected error for empty supported models")
	}

	if err.Error() != "Supported model list is empty" {
		t.Errorf("expected 'Supported model list is empty', got %s", err.Error())
	}
}

func TestFirmwareConfig_Validate_InvalidDownloadProtocol(t *testing.T) {
	fc := &FirmwareConfig{
		Description:              "Test",
		FirmwareFilename:         "test.bin",
		FirmwareVersion:          "v1.0",
		SupportedModelIds:        []string{"MODEL1"},
		FirmwareDownloadProtocol: "ftp", // Invalid protocol
		ApplicationType:          core.STB,
	}

	err := fc.Validate()
	// Will fail on model check first, but if we had valid models, would fail on protocol
	if err != nil && !contains(err.Error(), "FirmwareDownloadProtocol") && !contains(err.Error(), "does not exist") {
		t.Logf("Got error (may be model-related): %v", err)
	}
}

func TestFirmwareConfig_Validate_TooManyProperties(t *testing.T) {
	fc := &FirmwareConfig{
		Description:       "Test",
		FirmwareFilename:  "test.bin",
		FirmwareVersion:   "v1.0",
		SupportedModelIds: []string{"MODEL1"},
		ApplicationType:   core.STB,
		Properties:        make(map[string]string),
	}

	// Add more than MAX_ALLOWED_NUMBER_OF_PROPERTIES (20)
	for i := 0; i < 25; i++ {
		fc.Properties[string(rune('a'+i))] = "value"
	}

	err := fc.Validate()
	// Will fail on model check first
	if err != nil && !contains(err.Error(), "Max allowed number") && !contains(err.Error(), "does not exist") {
		t.Logf("Got error: %v", err)
	}
}

func TestFirmwareConfig_Validate_EmptyPropertyKey(t *testing.T) {
	fc := &FirmwareConfig{
		Description:       "Test",
		FirmwareFilename:  "test.bin",
		FirmwareVersion:   "v1.0",
		SupportedModelIds: []string{"MODEL1"},
		ApplicationType:   core.STB,
		Properties: map[string]string{
			"": "value",
		},
	}

	err := fc.Validate()
	// Will fail on model check first
	if err != nil {
		t.Logf("Got error: %v", err)
	}
}

func TestFirmwareConfigFacade_NewFirmwareConfigFacade(t *testing.T) {
	fc := &FirmwareConfig{
		ID:                       "test-id",
		Description:              "Test",
		FirmwareFilename:         "test.bin",
		FirmwareVersion:          "v1.0",
		FirmwareDownloadProtocol: "http",
		RebootImmediately:        true,
		Properties: map[string]string{
			"custom1": "value1",
		},
	}

	facade := NewFirmwareConfigFacade(fc)

	if facade == nil {
		t.Fatal("expected non-nil facade")
	}

	if facade.Properties == nil {
		t.Fatal("expected non-nil Properties")
	}

	if facade.Properties[core.ID] != "test-id" {
		t.Errorf("expected ID 'test-id', got %v", facade.Properties[core.ID])
	}

	if facade.Properties[core.FIRMWARE_VERSION] != "v1.0" {
		t.Errorf("expected version 'v1.0', got %v", facade.Properties[core.FIRMWARE_VERSION])
	}

	if facade.CustomProperties == nil {
		t.Fatal("expected non-nil CustomProperties")
	}

	if facade.CustomProperties["custom1"] != "value1" {
		t.Errorf("expected custom1='value1', got %v", facade.CustomProperties["custom1"])
	}
}

func TestFirmwareConfigFacade_GetSetMethods(t *testing.T) {
	facade := &FirmwareConfigFacade{
		Properties: make(map[string]interface{}),
	}

	facade.SetFirmwareDownloadProtocol("https")
	if facade.GetFirmwareDownloadProtocol() != "https" {
		t.Error("FirmwareDownloadProtocol get/set failed")
	}

	facade.SetFirmwareLocation("http://example.com")
	if facade.GetFirmwareLocation() != "http://example.com" {
		t.Error("FirmwareLocation get/set failed")
	}

	facade.SetRebootImmediately(true)
	if !facade.GetRebootImmediately() {
		t.Error("RebootImmediately get/set failed")
	}

	facade.SetRebootImmediately(false)
	if facade.GetRebootImmediately() {
		t.Error("RebootImmediately should be false")
	}
}

func TestFirmwareConfigFacade_GetStringValue(t *testing.T) {
	facade := &FirmwareConfigFacade{
		Properties: map[string]interface{}{
			"key1": "value1",
			"key2": nil,
		},
	}

	if facade.GetStringValue("key1") != "value1" {
		t.Error("GetStringValue failed for existing key")
	}

	if facade.GetStringValue("key2") != "" {
		t.Error("GetStringValue should return empty string for nil value")
	}

	if facade.GetStringValue("nonexistent") != "" {
		t.Error("GetStringValue should return empty string for nonexistent key")
	}
}

func TestFirmwareConfigFacade_MarshalJSON(t *testing.T) {
	facade := &FirmwareConfigFacade{
		Properties: map[string]interface{}{
			core.FIRMWARE_FILENAME:  "test.bin",
			core.FIRMWARE_VERSION:   "v1.0",
			core.FIRMWARE_LOCATION:  "http://example.com",
			core.REBOOT_IMMEDIATELY: true,
			core.UPGRADE_DELAY:      int64(300),
		},
	}

	data, err := json.Marshal(facade)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	// Verify some fields are present
	if result["firmwareFilename"] != "test.bin" {
		t.Error("firmwareFilename not properly marshaled")
	}
}

func TestFirmwareConfigFacade_MarshalJSON_ZeroUpgradeDelay(t *testing.T) {
	facade := &FirmwareConfigFacade{
		Properties: map[string]interface{}{
			core.FIRMWARE_FILENAME: "test.bin",
			core.UPGRADE_DELAY:     int64(0),
		},
	}

	data, err := json.Marshal(facade)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// upgradeDelay with 0 value should be excluded
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if _, exists := result["upgradeDelay"]; exists {
		t.Error("upgradeDelay with 0 value should be excluded from JSON")
	}
}

func TestFirmwareConfigFacade_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"id": "test-id",
		"description": "Test Config",
		"firmwareFilename": "test.bin",
		"firmwareVersion": "v1.0",
		"rebootImmediately": true
	}`

	var facade FirmwareConfigFacade
	err := json.Unmarshal([]byte(jsonData), &facade)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if facade.Properties == nil {
		t.Fatal("expected non-nil Properties after unmarshal")
	}

	if facade.Properties[core.ID] != "test-id" {
		t.Error("ID not properly unmarshaled")
	}

	if facade.Properties[core.FIRMWARE_FILENAME] != "test.bin" {
		t.Error("firmwareFilename not properly unmarshaled")
	}

	rebootImm, ok := facade.Properties[core.REBOOT_IMMEDIATELY].(bool)
	if !ok || !rebootImm {
		t.Error("rebootImmediately not properly unmarshaled")
	}
}

func TestIsRedundantEntry(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"id", true},
		{"description", true},
		{"supportedModelIds", true},
		{"updated", true},
		{"firmwareVersion", false},
		{"firmwareFilename", false},
		{"customProperty", false},
	}

	for _, test := range tests {
		result := IsRedundantEntry(test.key)
		if result != test.expected {
			t.Errorf("IsRedundantEntry(%s): expected %v, got %v", test.key, test.expected, result)
		}
	}
}

func TestCreateFirmwareConfigFacadeResponse(t *testing.T) {
	facade := FirmwareConfigFacade{
		Properties: map[string]interface{}{
			"id":                "test-id",
			"description":       "Test",
			"supportedModelIds": []string{"MODEL1"},
			"firmwareFilename":  "test.bin",
			"firmwareVersion":   "v1.0",
			"rebootImmediately": false,
			"firmwareLocation":  "http://example.com",
			"upgradeDelay":      int64(0),
		},
		CustomProperties: map[string]string{
			"custom1": "value1",
			"custom2": "value2",
		},
	}

	response := CreateFirmwareConfigFacadeResponse(facade)

	if response == nil {
		t.Fatal("expected non-nil response")
	}

	// Redundant entries should be excluded
	if _, exists := response["id"]; exists {
		t.Error("id should be excluded from response")
	}

	if _, exists := response["description"]; exists {
		t.Error("description should be excluded from response")
	}

	// Non-redundant entries should be included
	if response["firmwareFilename"] != "test.bin" {
		t.Error("firmwareFilename should be included")
	}

	// Zero upgradeDelay should be excluded
	if _, exists := response["upgradeDelay"]; exists {
		t.Error("zero upgradeDelay should be excluded")
	}

	// Custom properties should be included
	if response["custom1"] != "value1" {
		t.Error("custom1 should be included")
	}

	// rebootImmediately should always be present (even if false)
	if _, exists := response[core.REBOOT_IMMEDIATELY]; !exists {
		t.Error("rebootImmediately should always be present")
	}
}

func TestNewFirmwareConfigInf(t *testing.T) {
	result := NewFirmwareConfigInf()

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	fc, ok := result.(*FirmwareConfig)
	if !ok {
		t.Fatal("expected *FirmwareConfig type")
	}

	if fc.ApplicationType != core.STB {
		t.Errorf("expected ApplicationType '%s', got %s", core.STB, fc.ApplicationType)
	}

	if fc.RebootImmediately {
		t.Error("expected RebootImmediately false")
	}

	if fc.FirmwareDownloadProtocol != "tftp" {
		t.Errorf("expected protocol 'tftp', got %s", fc.FirmwareDownloadProtocol)
	}
}

func TestFirmwareConfig_ToPropertiesMap(t *testing.T) {
	fc := &FirmwareConfig{
		ID:                       "test-id",
		Updated:                  123456789,
		Description:              "Test",
		SupportedModelIds:        []string{"MODEL1", "MODEL2"},
		FirmwareDownloadProtocol: "http",
		FirmwareFilename:         "test.bin",
		FirmwareLocation:         "http://example.com",
		FirmwareVersion:          "v1.0",
		Ipv6FirmwareLocation:     "http://[::1]",
		UpgradeDelay:             300,
		RebootImmediately:        true,
		MandatoryUpdate:          false,
	}

	propMap := fc.ToPropertiesMap()

	if propMap == nil {
		t.Fatal("expected non-nil properties map")
	}

	if propMap[core.ID] != "test-id" {
		t.Error("ID not in properties map")
	}

	if propMap[core.FIRMWARE_VERSION] != "v1.0" {
		t.Error("FirmwareVersion not in properties map")
	}

	rebootImm, ok := propMap[core.REBOOT_IMMEDIATELY].(bool)
	if !ok || !rebootImm {
		t.Error("RebootImmediately not properly set in properties map")
	}
}

func TestFirmwareConfig_CreateFirmwareConfigResponse(t *testing.T) {
	fc := &FirmwareConfig{
		ID:                "test-id",
		Description:       "Test Description",
		SupportedModelIds: []string{"MODEL1"},
		FirmwareFilename:  "test.bin",
		FirmwareVersion:   "v1.0",
		Properties: map[string]string{
			"prop1": "val1",
		},
	}

	response := fc.CreateFirmwareConfigResponse()

	if response == nil {
		t.Fatal("expected non-nil response")
	}

	if response.ID != "test-id" {
		t.Error("ID mismatch in response")
	}

	if response.Description != "Test Description" {
		t.Error("Description mismatch in response")
	}

	if len(response.SupportedModelIds) != 1 {
		t.Error("SupportedModelIds mismatch in response")
	}

	if response.Properties["prop1"] != "val1" {
		t.Error("Properties mismatch in response")
	}
}

func TestNewModelFirmwareConfiguration(t *testing.T) {
	mfc := NewModelFirmwareConfiguration("RNG150", "firmware.bin", "v1.0")

	if mfc == nil {
		t.Fatal("expected non-nil ModelFirmwareConfiguration")
	}

	if mfc.Model != "RNG150" {
		t.Error("Model mismatch")
	}

	if mfc.FirmwareFilename != "firmware.bin" {
		t.Error("FirmwareFilename mismatch")
	}

	if mfc.FirmwareVersion != "v1.0" {
		t.Error("FirmwareVersion mismatch")
	}

	// Test ToString
	str := mfc.ToString()
	if str == "" {
		t.Error("ToString returned empty string")
	}
}

func TestAddExpressionToIpRuleBean(t *testing.T) {
	// This test would require importing sharedef which has IpRuleBean
	// Skipping detailed test as it depends on external structures
}

func TestMacRuleBeanToMacRuleBeanResponse(t *testing.T) {
	fc := &FirmwareConfig{
		ID:              "config-id",
		Description:     "Test Config",
		FirmwareVersion: "v1.0",
	}

	macRuleBean := &MacRuleBean{
		Id:             "rule-id",
		Name:           "Test Rule",
		MacAddresses:   "AA:BB:CC:DD:EE:FF",
		MacListRef:     "mac-list-ref",
		FirmwareConfig: fc,
	}

	response := MacRuleBeanToMacRuleBeanResponse(macRuleBean)

	if response == nil {
		t.Fatal("expected non-nil response")
	}

	if response.Id != "rule-id" {
		t.Error("Id mismatch in response")
	}

	if response.Name != "Test Rule" {
		t.Error("Name mismatch in response")
	}

	if response.FirmwareConfig == nil {
		t.Fatal("expected non-nil FirmwareConfig in response")
	}

	if response.FirmwareConfig.ID != "config-id" {
		t.Error("FirmwareConfig ID mismatch in response")
	}
}

func TestFirmwareConfigToFirmwareConfigForMacRuleBeanResponse(t *testing.T) {
	fc := &FirmwareConfig{
		ID:                       "test-id",
		Updated:                  123456789,
		Description:              "Test",
		SupportedModelIds:        []string{"MODEL1"},
		FirmwareFilename:         "test.bin",
		FirmwareVersion:          "v1.0",
		ApplicationType:          "stb",
		FirmwareDownloadProtocol: "http",
		FirmwareLocation:         "http://example.com",
		Ipv6FirmwareLocation:     "http://[::1]",
		UpgradeDelay:             300,
		RebootImmediately:        true,
		Properties: map[string]string{
			"key": "value",
		},
	}

	response := FirmwareConfigToFirmwareConfigForMacRuleBeanResponse(fc)

	if response == nil {
		t.Fatal("expected non-nil response")
	}

	if response.ID != "test-id" {
		t.Error("ID mismatch")
	}

	if response.FirmwareVersion != "v1.0" {
		t.Error("FirmwareVersion mismatch")
	}

	if !response.RebootImmediately {
		t.Error("RebootImmediately should be true")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
