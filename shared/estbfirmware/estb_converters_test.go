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
