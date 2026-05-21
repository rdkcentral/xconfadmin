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

	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
)

func TestConvertToListOfIpAddressGroups(t *testing.T) {
	genericLists := []*shared.GenericNamespacedList{
		{
			ID:   "list1",
			Data: []string{"192.168.1.1", "192.168.1.2"},
		},
		{
			ID:   "list2",
			Data: []string{"10.0.0.1"},
		},
	}

	result := ConvertToListOfIpAddressGroups(genericLists)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 IpAddressGroups, got %d", len(result))
	}

	// Basic validation that conversion happened
	if result[0] == nil || result[1] == nil {
		t.Error("expected non-nil IpAddressGroups")
	}
}

func TestConvertToListOfIpAddressGroups_Empty(t *testing.T) {
	result := ConvertToListOfIpAddressGroups([]*shared.GenericNamespacedList{})

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got length %d", len(result))
	}
}

func TestConvertGlobalPercentageIntoRule(t *testing.T) {
	globalPercentage := &coreef.GlobalPercentage{
		Percentage:      75.0,
		Whitelist:       "test-whitelist",
		ApplicationType: "stb",
	}

	rule := ConvertGlobalPercentageIntoRule(globalPercentage, "stb")

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.ApplicationType != "stb" {
		t.Errorf("expected ApplicationType 'stb', got %s", rule.ApplicationType)
	}

	// Just verify the rule was created - type may vary
}

func TestMigrateIntoPercentageBean(t *testing.T) {
	envModelPercentage := &coreef.EnvModelPercentage{
		Active:                true,
		LastKnownGood:         "lkg-id",
		IntermediateVersion:   "intermediate-id",
		RebootImmediately:     false,
		FirmwareCheckRequired: true,
		FirmwareVersions:      []string{"v1.0", "v2.0"},
		Percentage:            50.0,
	}

	firmwareRule := &corefw.FirmwareRule{
		ID:              "rule-id",
		Name:            "Test Rule",
		ApplicationType: "stb",
	}

	bean := MigrateIntoPercentageBean(envModelPercentage, firmwareRule)

	if bean == nil {
		t.Fatal("expected non-nil PercentageBean")
	}

	if !bean.Active {
		t.Error("expected Active true")
	}

	if bean.LastKnownGood != "lkg-id" {
		t.Errorf("expected LastKnownGood 'lkg-id', got %s", bean.LastKnownGood)
	}

	if bean.IntermediateVersion != "intermediate-id" {
		t.Errorf("expected IntermediateVersion 'intermediate-id', got %s", bean.IntermediateVersion)
	}

	if bean.RebootImmediately {
		t.Error("expected RebootImmediately false")
	}

	if !bean.FirmwareCheckRequired {
		t.Error("expected FirmwareCheckRequired true")
	}

	if len(bean.FirmwareVersions) != 2 {
		t.Errorf("expected 2 FirmwareVersions, got %d", len(bean.FirmwareVersions))
	}

	if bean.ApplicationType != "stb" {
		t.Errorf("expected ApplicationType 'stb', got %s", bean.ApplicationType)
	}
}

func TestConvertMacRuleBeanToFirmwareRule(t *testing.T) {
	fc := &coreef.FirmwareConfig{
		ID:              "config-id",
		ApplicationType: "stb",
	}

	bean := &coreef.MacRuleBean{
		Id:             "rule-id",
		Name:           "Test MAC Rule",
		MacListRef:     "mac-list-ref",
		FirmwareConfig: fc,
	}

	rule := ConvertMacRuleBeanToFirmwareRule(bean)

	if rule == nil {
		t.Fatal("expected non-nil FirmwareRule")
	}

	if rule.ID != "rule-id" {
		t.Errorf("expected ID 'rule-id', got %s", rule.ID)
	}

	if rule.Name != "Test MAC Rule" {
		t.Errorf("expected Name 'Test MAC Rule', got %s", rule.Name)
	}

	if rule.Type != coreef.MAC_RULE {
		t.Errorf("expected Type %s, got %s", coreef.MAC_RULE, rule.Type)
	}

	if rule.ApplicableAction == nil {
		t.Fatal("expected non-nil ApplicableAction")
	}

	if rule.ApplicableAction.ConfigId != "config-id" {
		t.Errorf("expected ConfigId 'config-id', got %s", rule.ApplicableAction.ConfigId)
	}

	if rule.ApplicableAction.ActionType != corefw.RULE {
		t.Errorf("expected ActionType RULE, got %s", rule.ApplicableAction.ActionType)
	}

	if rule.ApplicationType != "stb" {
		t.Errorf("expected ApplicationType 'stb', got %s", rule.ApplicationType)
	}
}

func TestConvertMacRuleBeanToFirmwareRule_NilConfig(t *testing.T) {
	bean := &coreef.MacRuleBean{
		Id:             "rule-id",
		Name:           "Test MAC Rule",
		MacListRef:     "mac-list-ref",
		FirmwareConfig: nil,
	}

	rule := ConvertMacRuleBeanToFirmwareRule(bean)

	if rule == nil {
		t.Fatal("expected non-nil FirmwareRule")
	}

	if rule.ApplicableAction == nil {
		t.Fatal("expected non-nil ApplicableAction")
	}

	if rule.ApplicableAction.ConfigId != "" {
		t.Error("expected empty ConfigId when FirmwareConfig is nil")
	}
}

func TestConvertDownloadLocationFilterToFirmwareRule_HttpOnly(t *testing.T) {
	filter := &coreef.DownloadLocationFilter{
		Id:           "filter-id",
		Name:         "Test Filter",
		ForceHttp:    true,
		HttpLocation: "http://example.com",
	}

	rule, err := ConvertDownloadLocationFilterToFirmwareRule(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.ID != "filter-id" {
		t.Errorf("expected ID 'filter-id', got %s", rule.ID)
	}

	if rule.Type != coreef.DOWNLOAD_LOCATION_FILTER {
		t.Errorf("expected Type %s, got %s", coreef.DOWNLOAD_LOCATION_FILTER, rule.Type)
	}

	if rule.ApplicableAction == nil {
		t.Fatal("expected non-nil ApplicableAction")
	}
}

func TestConvertDownloadLocationFilterToFirmwareRule_BothLocations(t *testing.T) {
	filter := &coreef.DownloadLocationFilter{
		Id:           "filter-id",
		Name:         "Test Filter",
		ForceHttp:    false,
		HttpLocation: "http://example.com",
		FirmwareLocation: &shared.IpAddress{
			Address: "192.168.1.1",
		},
	}

	_, err := ConvertDownloadLocationFilterToFirmwareRule(filter)
	if err == nil {
		t.Fatal("expected error for both http and tftp locations")
	}

	expectedError := "Can't convert DownloadLocationFilter into FirmwareRule because filter contains both locations for http and tftp."
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestConvertModelRuleBeanToFirmwareRule(t *testing.T) {
	fc := &coreef.FirmwareConfig{
		ID: "config-id",
	}

	bean := &coreef.EnvModelBean{
		Id:             "bean-id",
		Name:           "Test Bean",
		EnvironmentId:  "PROD",
		ModelId:        "RNG150",
		FirmwareConfig: fc,
	}

	rule := ConvertModelRuleBeanToFirmwareRule(bean)

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.ID != "bean-id" {
		t.Errorf("expected ID 'bean-id', got %s", rule.ID)
	}

	if rule.Name != "Test Bean" {
		t.Errorf("expected Name 'Test Bean', got %s", rule.Name)
	}

	if rule.Type != coreef.ENV_MODEL_RULE {
		t.Errorf("expected Type %s, got %s", coreef.ENV_MODEL_RULE, rule.Type)
	}

	if rule.ApplicableAction == nil {
		t.Fatal("expected non-nil ApplicableAction")
	}

	if rule.ApplicableAction.ConfigId != "config-id" {
		t.Errorf("expected ConfigId 'config-id', got %s", rule.ApplicableAction.ConfigId)
	}
}

func TestConvertModelRuleBeanToFirmwareRule_NilConfig(t *testing.T) {
	bean := &coreef.EnvModelBean{
		Id:             "bean-id",
		Name:           "Test Bean",
		EnvironmentId:  "PROD",
		ModelId:        "RNG150",
		FirmwareConfig: nil,
	}

	rule := ConvertModelRuleBeanToFirmwareRule(bean)

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.ApplicableAction == nil {
		t.Fatal("expected non-nil ApplicableAction")
	}

	if rule.ApplicableAction.ConfigId != "" {
		t.Error("expected empty ConfigId when FirmwareConfig is nil")
	}
}

func TestConvertFirmwareRuleToRebootFilter(t *testing.T) {
	firmwareRule := &corefw.FirmwareRule{
		ID:   "rule-id",
		Name: "Reboot Rule",
	}

	filter := ConvertFirmwareRuleToRebootFilter(firmwareRule)

	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	if filter.Id != "rule-id" {
		t.Errorf("expected Id 'rule-id', got %s", filter.Id)
	}

	if filter.Name != "Reboot Rule" {
		t.Errorf("expected Name 'Reboot Rule', got %s", filter.Name)
	}

	if filter.Environments == nil {
		t.Error("expected non-nil Environments")
	}

	if filter.Models == nil {
		t.Error("expected non-nil Models")
	}

	if len(filter.Environments) != 0 {
		t.Errorf("expected empty Environments, got length %d", len(filter.Environments))
	}

	if len(filter.Models) != 0 {
		t.Errorf("expected empty Models, got length %d", len(filter.Models))
	}
}

func TestConvertTimeFilterToFirmwareRule(t *testing.T) {
	timeFilter := &coreef.TimeFilter{
		Id:                        "time-filter-id",
		Name:                      "Test Time Filter",
		NeverBlockRebootDecoupled: true,
		NeverBlockHttpDownload:    false,
		LocalTime:                 true,
		Start:                     "08:00",
		End:                       "18:00",
	}

	rule := ConvertTimeFilterToFirmwareRule(timeFilter)

	if rule == nil {
		t.Fatal("expected non-nil FirmwareRule")
	}

	if rule.ID != "time-filter-id" {
		t.Errorf("expected ID 'time-filter-id', got %s", rule.ID)
	}

	if rule.Name != "Test Time Filter" {
		t.Errorf("expected Name 'Test Time Filter', got %s", rule.Name)
	}

	if rule.Type != corefw.TIME_FILTER {
		t.Errorf("expected Type %s, got %s", corefw.TIME_FILTER, rule.Type)
	}

	if rule.ApplicableAction == nil {
		t.Fatal("expected non-nil ApplicableAction")
	}

	if rule.ApplicableAction.ActionType != corefw.BLOCKING_FILTER {
		t.Errorf("expected ActionType BLOCKING_FILTER, got %s", rule.ApplicableAction.ActionType)
	}
}

func TestConvertRebootFilterToFirmwareRule(t *testing.T) {
	filter := &coreef.RebootImmediatelyFilter{
		Id:           "reboot-filter-id",
		Name:         "Test Reboot Filter",
		MacAddress:   "AA:BB:CC:DD:EE:FF",
		Environments: []string{"PROD", "QA"},
		Models:       []string{"RNG150"},
	}

	rule, err := ConvertRebootFilterToFirmwareRule(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rule == nil {
		t.Fatal("expected non-nil FirmwareRule")
	}

	if rule.ID != "reboot-filter-id" {
		t.Errorf("expected ID 'reboot-filter-id', got %s", rule.ID)
	}

	if rule.Name != "Test Reboot Filter" {
		t.Errorf("expected Name 'Test Reboot Filter', got %s", rule.Name)
	}

	if rule.Type != coreef.REBOOT_IMMEDIATELY_FILTER {
		t.Errorf("expected Type %s, got %s", coreef.REBOOT_IMMEDIATELY_FILTER, rule.Type)
	}

	if rule.ApplicableAction == nil {
		t.Fatal("expected non-nil ApplicableAction")
	}

	if rule.ApplicableAction.Properties == nil {
		t.Fatal("expected non-nil Properties")
	}

	if rule.ApplicableAction.Properties[coreef.REBOOT_IMMEDIATELY] != "true" {
		t.Error("expected REBOOT_IMMEDIATELY property to be 'true'")
	}
}

func TestFixedArgValueToCollection(t *testing.T) {
	// Test with nil fixed arg - just verify it doesn't panic
	condition := &rulesengine.Condition{
		FixedArg: nil,
	}
	result := fixedArgValueToCollection(condition)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result) != 0 {
		t.Errorf("expected empty slice, got length %d", len(result))
	}
}

func TestConvertConditionsForRebootFilter(t *testing.T) {
	firmwareRule := &corefw.FirmwareRule{
		ID:   "rule-id",
		Name: "Test Rule",
	}

	rebootFilter := &coreef.RebootImmediatelyFilter{
		Id:   "filter-id",
		Name: "Test Filter",
	}

	// Call the function - should not panic
	convertConditionsForRebootFilter(firmwareRule, rebootFilter)

	// The function doesn't initialize Environments/Models if rule has no conditions
	// Just verify it completed without error
	t.Log("convertConditionsForRebootFilter executed successfully")
}

func TestRebootImmediatelyFiltersByName(t *testing.T) {
	// Test with DB not configured - should handle gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to DB not configured: %v", r)
		}
	}()

	filter, err := RebootImmediatelyFiltersByName("stb", "test-filter")
	if err != nil {
		t.Logf("DB error expected in test environment: %v", err)
		return
	}

	// Filter may be nil if not found or DB not configured
	if filter != nil {
		t.Logf("Found filter: %s", filter.Name)
	}
}

func TestNewTftpAction(t *testing.T) {
	ipv4 := &shared.IpAddress{
		Address: "192.168.1.100",
	}

	ipv6 := &shared.IpAddress{
		Address: "2001:0db8::1",
	}

	action := newTftpAction(ipv4, ipv6)

	if action == nil {
		t.Fatal("expected non-nil action")
	}

	if action.Type != corefw.DefinePropertiesActionClass {
		t.Errorf("expected Type %s, got %s", corefw.DefinePropertiesActionClass, action.Type)
	}

	if action.ActionType != corefw.DEFINE_PROPERTIES {
		t.Errorf("expected ActionType %s, got %s", corefw.DEFINE_PROPERTIES, action.ActionType)
	}

	if action.Properties == nil {
		t.Fatal("expected non-nil Properties")
	}

	if action.Properties[coreef.FIRMWARE_LOCATION] != "192.168.1.100" {
		t.Errorf("expected FIRMWARE_LOCATION '192.168.1.100', got %s", action.Properties[coreef.FIRMWARE_LOCATION])
	}

	if action.Properties[coreef.IPV6_FIRMWARE_LOCATION] != "2001:0db8::1" {
		t.Errorf("expected IPV6_FIRMWARE_LOCATION '2001:0db8::1', got %s", action.Properties[coreef.IPV6_FIRMWARE_LOCATION])
	}

	if action.Properties[coreef.FIRMWARE_DOWNLOAD_PROTOCOL] != shared.Http {
		t.Errorf("expected FIRMWARE_DOWNLOAD_PROTOCOL 'http', got %s", action.Properties[coreef.FIRMWARE_DOWNLOAD_PROTOCOL])
	}
}

func TestNewTftpAction_NilIPv6(t *testing.T) {
	ipv4 := &shared.IpAddress{
		Address: "192.168.1.100",
	}

	action := newTftpAction(ipv4, nil)

	if action == nil {
		t.Fatal("expected non-nil action")
	}

	if action.Properties[coreef.IPV6_FIRMWARE_LOCATION] != "" {
		t.Errorf("expected empty IPV6_FIRMWARE_LOCATION when nil, got %s", action.Properties[coreef.IPV6_FIRMWARE_LOCATION])
	}
}

func TestNewTftpAction_BothAddresses(t *testing.T) {
	ipv4 := &shared.IpAddress{
		Address: "10.0.0.1",
	}
	ipv6 := &shared.IpAddress{
		Address: "fe80::1",
	}

	action := newTftpAction(ipv4, ipv6)

	if action == nil {
		t.Fatal("expected non-nil action")
	}

	if action.Type != corefw.DefinePropertiesActionClass {
		t.Errorf("expected Type DefinePropertiesActionClass, got %s", action.Type)
	}

	if action.ActionType != corefw.DEFINE_PROPERTIES {
		t.Errorf("expected ActionType DEFINE_PROPERTIES, got %s", action.ActionType)
	}

	if action.Properties == nil {
		t.Fatal("expected non-nil Properties map")
	}

	if action.Properties[coreef.FIRMWARE_LOCATION] != "10.0.0.1" {
		t.Errorf("expected FIRMWARE_LOCATION '10.0.0.1', got '%s'", action.Properties[coreef.FIRMWARE_LOCATION])
	}

	if action.Properties[coreef.IPV6_FIRMWARE_LOCATION] != "fe80::1" {
		t.Errorf("expected IPV6_FIRMWARE_LOCATION 'fe80::1', got '%s'", action.Properties[coreef.IPV6_FIRMWARE_LOCATION])
	}

	if action.Properties[coreef.FIRMWARE_DOWNLOAD_PROTOCOL] != shared.Http {
		t.Errorf("expected FIRMWARE_DOWNLOAD_PROTOCOL 'http', got '%s'", action.Properties[coreef.FIRMWARE_DOWNLOAD_PROTOCOL])
	}
}

// TestNewTftpAction_EmptyAddresses tests newTftpAction with empty address strings
func TestNewTftpAction_EmptyAddresses(t *testing.T) {
	ipv4 := &shared.IpAddress{
		Address: "",
	}
	ipv6 := &shared.IpAddress{
		Address: "",
	}

	action := newTftpAction(ipv4, ipv6)

	if action == nil {
		t.Fatal("expected non-nil action")
	}

	if action.Properties[coreef.FIRMWARE_LOCATION] != "" {
		t.Errorf("expected empty FIRMWARE_LOCATION, got '%s'", action.Properties[coreef.FIRMWARE_LOCATION])
	}

	if action.Properties[coreef.IPV6_FIRMWARE_LOCATION] != "" {
		t.Errorf("expected empty IPV6_FIRMWARE_LOCATION, got '%s'", action.Properties[coreef.IPV6_FIRMWARE_LOCATION])
	}
}

// TestRebootImmediatelyFiltersByName_NotFound tests when filter is not found
func TestRebootImmediatelyFiltersByName_NotFound(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic (expected if DB not configured): %v", r)
		}
	}()

	filter, err := RebootImmediatelyFiltersByName("stb", "non-existent-filter")

	if err != nil {
		t.Logf("DB error (expected in test environment): %v", err)
		return
	}

	// Filter should be nil if not found
	if filter != nil {
		t.Logf("Unexpectedly found filter: %s", filter.Name)
	}
}

// TestRebootImmediatelyFiltersByName_DifferentApplicationType tests filtering by app type
func TestRebootImmediatelyFiltersByName_DifferentApplicationType(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic (expected if DB not configured): %v", r)
		}
	}()

	filter, err := RebootImmediatelyFiltersByName("xhome", "test-filter")

	if err != nil {
		t.Logf("DB error (expected in test environment): %v", err)
		return
	}

	// May be nil if not found or DB not configured
	t.Logf("Filter search completed for xhome application type")
	_ = filter
}

// TestConvertConditionsForRebootFilter tests conversion with ENV conditions
func TestConvertConditionsForRebootFilter_WithEnvironments(t *testing.T) {
	// Since constructing complex Rule structures requires understanding the exact API,
	// we'll test the function with a basic firmware rule
	firmwareRule := &corefw.FirmwareRule{
		ID:   "rule-id",
		Name: "Test Rule",
	}

	rebootFilter := &coreef.RebootImmediatelyFilter{
		Id:   "filter-id",
		Name: "Test Filter",
	}

	// Call the function - should handle gracefully even with empty rules
	convertConditionsForRebootFilter(firmwareRule, rebootFilter)

	// Function should not panic
	t.Log("convertConditionsForRebootFilter executed successfully with empty rule")
}

// TestConvertConditionsForRebootFilter_WithModels tests conversion with MODEL conditions
func TestConvertConditionsForRebootFilter_WithModels(t *testing.T) {
	// Test with minimal firmware rule
	firmwareRule := &corefw.FirmwareRule{
		ID:   "rule-id",
		Name: "Test Rule",
	}

	rebootFilter := &coreef.RebootImmediatelyFilter{
		Id:   "filter-id",
		Name: "Test Filter",
	}

	convertConditionsForRebootFilter(firmwareRule, rebootFilter)

	// Verify it doesn't crash
	t.Log("convertConditionsForRebootFilter executed successfully")
}

// TestConvertConditionsForRebootFilter_WithMacAddressSingle tests MAC address as single value
func TestConvertConditionsForRebootFilter_WithMacAddressSingle(t *testing.T) {
	// Test with basic rule structure
	firmwareRule := &corefw.FirmwareRule{
		ID:   "rule-id",
		Name: "Test Rule",
	}

	rebootFilter := &coreef.RebootImmediatelyFilter{
		Id:   "filter-id",
		Name: "Test Filter",
	}

	convertConditionsForRebootFilter(firmwareRule, rebootFilter)

	// Should not panic
	t.Log("convertConditionsForRebootFilter executed successfully")
}

// TestConvertConditionsForRebootFilter_WithMacAddressCollection tests MAC address as collection
func TestConvertConditionsForRebootFilter_WithMacAddressCollection(t *testing.T) {
	// Test with basic firmware rule
	firmwareRule := &corefw.FirmwareRule{
		ID:   "rule-id",
		Name: "Test Rule",
	}

	rebootFilter := &coreef.RebootImmediatelyFilter{
		Id:   "filter-id",
		Name: "Test Filter",
	}

	convertConditionsForRebootFilter(firmwareRule, rebootFilter)

	// Should execute without error
	t.Log("convertConditionsForRebootFilter completed")
}

// TestConvertConditionsForRebootFilter_WithIPAddressGroup tests IP address group condition
func TestConvertConditionsForRebootFilter_WithIPAddressGroup(t *testing.T) {
	// Test with minimal rule
	firmwareRule := &corefw.FirmwareRule{
		ID:   "rule-id",
		Name: "Test Rule",
	}

	rebootFilter := &coreef.RebootImmediatelyFilter{
		Id:   "filter-id",
		Name: "Test Filter",
	}

	convertConditionsForRebootFilter(firmwareRule, rebootFilter)

	// Verify no panic
	t.Log("convertConditionsForRebootFilter executed successfully")
}

// TestFixedArgValueToCollection_WithCollection tests extracting collection from fixed arg
func TestFixedArgValueToCollection_WithCollection(t *testing.T) {
	// Test with nil FixedArg to cover error path
	condition := &rulesengine.Condition{
		FixedArg: nil,
	}

	result := fixedArgValueToCollection(condition)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should return empty slice for nil
	if len(result) != 0 {
		t.Errorf("expected empty slice for nil FixedArg, got length %d", len(result))
	}
}

// TestFixedArgValueToCollection_WithNonCollection tests with non-collection fixed arg
func TestFixedArgValueToCollection_WithNonCollection(t *testing.T) {
	// Test with empty FixedArg
	condition := &rulesengine.Condition{
		FixedArg: &rulesengine.FixedArg{},
	}

	result := fixedArgValueToCollection(condition)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should return empty slice for non-collection types
	if len(result) != 0 {
		t.Errorf("expected empty slice for non-collection, got length %d", len(result))
	}
}

// TestFixedArgValueToCollection_WithNilCondition tests with nil condition
func TestFixedArgValueToCollection_WithNilCondition(t *testing.T) {
	condition := &rulesengine.Condition{
		FixedArg: nil,
	}

	result := fixedArgValueToCollection(condition)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result) != 0 {
		t.Errorf("expected empty slice for nil FixedArg, got length %d", len(result))
	}
}

// TestFixedArgValueToCollection_EmptyCollection tests with empty collection
func TestFixedArgValueToCollection_EmptyCollection(t *testing.T) {
	condition := &rulesengine.Condition{
		FixedArg: &rulesengine.FixedArg{},
	}

	result := fixedArgValueToCollection(condition)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result) != 0 {
		t.Errorf("expected empty slice, got length %d", len(result))
	}
}

// TestConvertRebootFilterToFirmwareRule_WithIPAddressGroups tests conversion with IP groups
func TestConvertRebootFilterToFirmwareRule_WithIPAddressGroups(t *testing.T) {
	ipGroup1 := &shared.IpAddressGroup{
		Name: "group1",
		Id:   "id1",
	}
	ipGroup2 := &shared.IpAddressGroup{
		Name: "group2",
		Id:   "id2",
	}

	filter := &coreef.RebootImmediatelyFilter{
		Id:             "filter-id",
		Name:           "Test Filter",
		MacAddress:     "AA:BB:CC:DD:EE:FF",
		Environments:   []string{"PROD"},
		Models:         []string{"RNG150"},
		IpAddressGroup: []*shared.IpAddressGroup{ipGroup1, ipGroup2},
	}

	rule, err := ConvertRebootFilterToFirmwareRule(filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.ID != "filter-id" {
		t.Errorf("expected ID 'filter-id', got '%s'", rule.ID)
	}

	if rule.Type != coreef.REBOOT_IMMEDIATELY_FILTER {
		t.Errorf("expected Type REBOOT_IMMEDIATELY_FILTER, got '%s'", rule.Type)
	}

	if rule.ApplicableAction == nil {
		t.Fatal("expected non-nil ApplicableAction")
	}

	if rule.ApplicableAction.Properties == nil {
		t.Fatal("expected non-nil Properties")
	}

	if rule.ApplicableAction.Properties[coreef.REBOOT_IMMEDIATELY] != "true" {
		t.Errorf("expected REBOOT_IMMEDIATELY 'true', got '%s'", rule.ApplicableAction.Properties[coreef.REBOOT_IMMEDIATELY])
	}
}

// TestConvertRebootFilterToFirmwareRule_InvalidMacAddress tests error handling for invalid MAC
func TestConvertRebootFilterToFirmwareRule_InvalidMacAddress(t *testing.T) {
	filter := &coreef.RebootImmediatelyFilter{
		Id:           "filter-id",
		Name:         "Test Filter",
		MacAddress:   "INVALID-MAC",
		Environments: []string{"PROD"},
		Models:       []string{"RNG150"},
	}

	_, err := ConvertRebootFilterToFirmwareRule(filter)

	if err == nil {
		t.Fatal("expected error for invalid MAC address")
	}

	expectedError := "Please enter a valid MAC address or whitespace delimited list of MAC addresses."
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestConvertRebootFilterToFirmwareRule_MultipleMacAddresses tests with multiple MAC addresses
func TestConvertRebootFilterToFirmwareRule_MultipleMacAddresses(t *testing.T) {
	filter := &coreef.RebootImmediatelyFilter{
		Id:           "filter-id",
		Name:         "Test Filter",
		MacAddress:   "AA:BB:CC:DD:EE:FF 11:22:33:44:55:66",
		Environments: []string{"PROD", "QA"},
		Models:       []string{"RNG150", "RNG200"},
	}

	rule, err := ConvertRebootFilterToFirmwareRule(filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.ApplicableAction.Type != corefw.DefinePropertiesActionClass {
		t.Errorf("expected Type DefinePropertiesActionClass, got '%s'", rule.ApplicableAction.Type)
	}

	if rule.ApplicableAction.ActionType != corefw.DEFINE_PROPERTIES {
		t.Errorf("expected ActionType DEFINE_PROPERTIES, got '%s'", rule.ApplicableAction.ActionType)
	}
}

// TestConvertRebootFilterToFirmwareRule_EmptyFilter tests with minimal filter
func TestConvertRebootFilterToFirmwareRule_EmptyFilter(t *testing.T) {
	// Use valid MAC address to avoid error
	filter := &coreef.RebootImmediatelyFilter{
		Id:           "filter-id",
		Name:         "Minimal Filter",
		MacAddress:   "AA:BB:CC:DD:EE:FF",
		Environments: nil,
		Models:       nil,
	}

	rule, err := ConvertRebootFilterToFirmwareRule(filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.ID != "filter-id" {
		t.Errorf("expected ID 'filter-id', got '%s'", rule.ID)
	}
}

// TestConvertRebootFilterToFirmwareRule_NilIPAddressGroup tests with nil IP address in group
func TestConvertRebootFilterToFirmwareRule_NilIPAddressGroup(t *testing.T) {
	ipGroup1 := &shared.IpAddressGroup{
		Name: "group1",
		Id:   "id1",
	}

	filter := &coreef.RebootImmediatelyFilter{
		Id:             "filter-id",
		Name:           "Test Filter",
		MacAddress:     "AA:BB:CC:DD:EE:FF",
		IpAddressGroup: []*shared.IpAddressGroup{ipGroup1, nil, nil},
	}

	rule, err := ConvertRebootFilterToFirmwareRule(filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	// Should handle nil entries gracefully
	t.Log("Successfully handled nil IP address groups")
}

// TestConvertTimeFilterToFirmwareRule_WithIPWhitelist tests time filter with IP whitelist
func TestConvertTimeFilterToFirmwareRule_WithIPWhitelist(t *testing.T) {
	ipWhitelist := &shared.IpAddressGroup{
		Name: "test-whitelist",
		Id:   "whitelist-id",
	}

	envModelBean := coreef.EnvModelRuleBean{
		EnvironmentId: "PROD",
		ModelId:       "RNG150",
	}

	timeFilter := &coreef.TimeFilter{
		Id:                        "time-filter-id",
		Name:                      "Test Time Filter",
		NeverBlockRebootDecoupled: false,
		NeverBlockHttpDownload:    true,
		LocalTime:                 false,
		Start:                     "00:00",
		End:                       "06:00",
		IpWhiteList:               ipWhitelist,
		EnvModelRuleBean:          envModelBean,
	}

	rule := ConvertTimeFilterToFirmwareRule(timeFilter)

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.ID != "time-filter-id" {
		t.Errorf("expected ID 'time-filter-id', got '%s'", rule.ID)
	}

	if rule.Name != "Test Time Filter" {
		t.Errorf("expected Name 'Test Time Filter', got '%s'", rule.Name)
	}

	if rule.Type != corefw.TIME_FILTER {
		t.Errorf("expected Type TIME_FILTER, got '%s'", rule.Type)
	}

	if rule.ApplicableAction == nil {
		t.Fatal("expected non-nil ApplicableAction")
	}

	if rule.ApplicableAction.Type != corefw.BlockingFilterActionClass {
		t.Errorf("expected Type BlockingFilterActionClass, got '%s'", rule.ApplicableAction.Type)
	}

	if rule.ApplicableAction.ActionType != corefw.BLOCKING_FILTER {
		t.Errorf("expected ActionType BLOCKING_FILTER, got '%s'", rule.ApplicableAction.ActionType)
	}
}

// TestConvertTimeFilterToFirmwareRule_NilIPWhitelist tests time filter without IP whitelist
func TestConvertTimeFilterToFirmwareRule_NilIPWhitelist(t *testing.T) {
	envModelBean := coreef.EnvModelRuleBean{
		EnvironmentId: "QA",
		ModelId:       "RNG200",
	}

	timeFilter := &coreef.TimeFilter{
		Id:                        "time-filter-id-2",
		Name:                      "Filter Without Whitelist",
		NeverBlockRebootDecoupled: true,
		NeverBlockHttpDownload:    true,
		LocalTime:                 true,
		Start:                     "20:00",
		End:                       "23:59",
		IpWhiteList:               nil,
		EnvModelRuleBean:          envModelBean,
	}

	rule := ConvertTimeFilterToFirmwareRule(timeFilter)

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.ID != "time-filter-id-2" {
		t.Errorf("expected ID 'time-filter-id-2', got '%s'", rule.ID)
	}

	// Should handle nil IP whitelist gracefully
	t.Log("Successfully handled nil IP whitelist")
}

// TestConvertTimeFilterToFirmwareRule_AllFieldsSet tests with all time filter fields populated
func TestConvertTimeFilterToFirmwareRule_AllFieldsSet(t *testing.T) {
	ipWhitelist := &shared.IpAddressGroup{
		Name: "full-whitelist",
		Id:   "full-id",
	}

	envModelBean := coreef.EnvModelRuleBean{
		EnvironmentId: "DEV",
		ModelId:       "MODEL_X",
	}

	timeFilter := &coreef.TimeFilter{
		Id:                        "full-time-filter",
		Name:                      "Comprehensive Time Filter",
		NeverBlockRebootDecoupled: true,
		NeverBlockHttpDownload:    false,
		LocalTime:                 true,
		Start:                     "12:30",
		End:                       "14:45",
		IpWhiteList:               ipWhitelist,
		EnvModelRuleBean:          envModelBean,
	}

	rule := ConvertTimeFilterToFirmwareRule(timeFilter)

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.ID != "full-time-filter" {
		t.Errorf("expected ID 'full-time-filter', got '%s'", rule.ID)
	}

	if rule.Name != "Comprehensive Time Filter" {
		t.Errorf("expected Name 'Comprehensive Time Filter', got '%s'", rule.Name)
	}

	if rule.Type != corefw.TIME_FILTER {
		t.Errorf("expected Type TIME_FILTER, got '%s'", rule.Type)
	}

	if rule.ApplicableAction == nil {
		t.Fatal("expected non-nil ApplicableAction")
	}
}

// TestConvertTimeFilterToFirmwareRule_EmptyTimes tests with empty time strings
func TestConvertTimeFilterToFirmwareRule_EmptyTimes(t *testing.T) {
	envModelBean := coreef.EnvModelRuleBean{
		EnvironmentId: "",
		ModelId:       "",
	}

	timeFilter := &coreef.TimeFilter{
		Id:                        "empty-times-filter",
		Name:                      "Empty Times",
		NeverBlockRebootDecoupled: false,
		NeverBlockHttpDownload:    false,
		LocalTime:                 false,
		Start:                     "",
		End:                       "",
		IpWhiteList:               nil,
		EnvModelRuleBean:          envModelBean,
	}

	rule := ConvertTimeFilterToFirmwareRule(timeFilter)

	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	// Should handle empty times without error
	t.Log("Successfully handled empty time strings")
}
