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

	shared "github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
)

func TestNewEmptyPercentFilterWrapper(t *testing.T) {
	wrapper := NewEmptyPercentFilterWrapper()

	if wrapper == nil {
		t.Fatal("expected non-nil wrapper")
	}

	if wrapper.ID != coreef.PERCENT_FILTER_SINGLETON_ID {
		t.Errorf("expected ID %s, got %s", coreef.PERCENT_FILTER_SINGLETON_ID, wrapper.ID)
	}

	if wrapper.Type != coreef.PercentFilterWrapperClass {
		t.Errorf("expected Type %s, got %s", coreef.PercentFilterWrapperClass, wrapper.Type)
	}

	if wrapper.Percentage != 100.0 {
		t.Errorf("expected Percentage 100.0, got %f", wrapper.Percentage)
	}

	if wrapper.EnvModelPercentages == nil {
		t.Error("expected non-nil EnvModelPercentages")
	}

	if len(wrapper.EnvModelPercentages) != 0 {
		t.Errorf("expected empty EnvModelPercentages, got length %d", len(wrapper.EnvModelPercentages))
	}
}

func TestNewPercentFilterWrapper_BasicConversion(t *testing.T) {
	// Create a basic PercentFilterValue
	percentFilterValue := &coreef.PercentFilterValue{
		ID:         "TEST_PERCENT_FILTER_VALUE",
		Percentage: 75.0,
		Whitelist: &shared.IpAddressGroup{
			Id:   "test-whitelist",
			Name: "Test Whitelist",
		},
		EnvModelPercentages: map[string]coreef.EnvModelPercentage{},
	}

	wrapper := NewPercentFilterWrapper(percentFilterValue, false)

	if wrapper == nil {
		t.Fatal("expected non-nil wrapper")
	}

	if wrapper.ID != percentFilterValue.ID {
		t.Errorf("expected ID %s, got %s", percentFilterValue.ID, wrapper.ID)
	}

	if wrapper.Percentage != 75.0 {
		t.Errorf("expected Percentage 75.0, got %f", wrapper.Percentage)
	}

	if wrapper.Whitelist == nil {
		t.Fatal("expected non-nil Whitelist")
	}

	if wrapper.Whitelist.Id != "test-whitelist" {
		t.Errorf("expected Whitelist Id 'test-whitelist', got %s", wrapper.Whitelist.Id)
	}
}

func TestNewPercentFilterWrapper_WithEnvModelPercentages_NoHumanReadable(t *testing.T) {
	envModelPercentages := map[string]coreef.EnvModelPercentage{
		"PROD-RNG150": {
			Percentage:            50.0,
			Active:                true,
			FirmwareCheckRequired: true,
			RebootImmediately:     false,
			LastKnownGood:         "lkg-config-id",
			IntermediateVersion:   "intermediate-config-id",
			FirmwareVersions:      []string{"v1.0", "v2.0"},
		},
		"QA-MODEL1": {
			Percentage:        25.0,
			Active:            false,
			FirmwareVersions:  []string{"v3.0"},
			RebootImmediately: true,
		},
	}

	percentFilterValue := &coreef.PercentFilterValue{
		ID:                  coreef.PERCENT_FILTER_SINGLETON_ID,
		Percentage:          100.0,
		EnvModelPercentages: envModelPercentages,
	}

	wrapper := NewPercentFilterWrapper(percentFilterValue, false)

	if wrapper == nil {
		t.Fatal("expected non-nil wrapper")
	}

	if len(wrapper.EnvModelPercentages) != 2 {
		t.Fatalf("expected 2 EnvModelPercentages, got %d", len(wrapper.EnvModelPercentages))
	}

	// Check that Name is set correctly
	foundProd := false
	foundQa := false
	for _, emp := range wrapper.EnvModelPercentages {
		if emp.Name == "PROD-RNG150" {
			foundProd = true
			if emp.Percentage != 50.0 {
				t.Errorf("expected Percentage 50.0 for PROD-RNG150, got %f", emp.Percentage)
			}
			if !emp.Active {
				t.Error("expected Active true for PROD-RNG150")
			}
			if !emp.FirmwareCheckRequired {
				t.Error("expected FirmwareCheckRequired true for PROD-RNG150")
			}
			// When toHumanReadableForm is false, versions should remain as IDs
			if emp.LastKnownGood != "lkg-config-id" {
				t.Errorf("expected LastKnownGood 'lkg-config-id', got %s", emp.LastKnownGood)
			}
			if emp.IntermediateVersion != "intermediate-config-id" {
				t.Errorf("expected IntermediateVersion 'intermediate-config-id', got %s", emp.IntermediateVersion)
			}
		} else if emp.Name == "QA-MODEL1" {
			foundQa = true
			if emp.Percentage != 25.0 {
				t.Errorf("expected Percentage 25.0 for QA-MODEL1, got %f", emp.Percentage)
			}
			if emp.Active {
				t.Error("expected Active false for QA-MODEL1")
			}
			if !emp.RebootImmediately {
				t.Error("expected RebootImmediately true for QA-MODEL1")
			}
		}
	}

	if !foundProd {
		t.Error("expected to find PROD-RNG150 in EnvModelPercentages")
	}
	if !foundQa {
		t.Error("expected to find QA-MODEL1 in EnvModelPercentages")
	}
}

func TestNewPercentFilterWrapper_EmptyEnvModelPercentages(t *testing.T) {
	percentFilterValue := &coreef.PercentFilterValue{
		ID:                  "TEST_ID",
		Percentage:          90.0,
		EnvModelPercentages: map[string]coreef.EnvModelPercentage{},
	}

	wrapper := NewPercentFilterWrapper(percentFilterValue, true)

	if wrapper == nil {
		t.Fatal("expected non-nil wrapper")
	}

	if len(wrapper.EnvModelPercentages) != 0 {
		t.Errorf("expected empty EnvModelPercentages, got %d items", len(wrapper.EnvModelPercentages))
	}
}

func TestNewPercentFilterWrapper_NilWhitelist(t *testing.T) {
	percentFilterValue := &coreef.PercentFilterValue{
		ID:                  "TEST_ID",
		Percentage:          50.0,
		Whitelist:           nil,
		EnvModelPercentages: map[string]coreef.EnvModelPercentage{},
	}

	wrapper := NewPercentFilterWrapper(percentFilterValue, false)

	if wrapper == nil {
		t.Fatal("expected non-nil wrapper")
	}

	if wrapper.Whitelist != nil {
		t.Error("expected nil Whitelist")
	}
}

func TestNewPercentFilterWrapper_EnvModelWithEmptyVersions(t *testing.T) {
	envModelPercentages := map[string]coreef.EnvModelPercentage{
		"ENV-MODEL": {
			Percentage:          30.0,
			LastKnownGood:       "",
			IntermediateVersion: "",
		},
	}

	percentFilterValue := &coreef.PercentFilterValue{
		ID:                  "TEST_ID",
		Percentage:          100.0,
		EnvModelPercentages: envModelPercentages,
	}

	// Test with toHumanReadableForm = true, but versions are empty
	wrapper := NewPercentFilterWrapper(percentFilterValue, true)

	if wrapper == nil {
		t.Fatal("expected non-nil wrapper")
	}

	if len(wrapper.EnvModelPercentages) != 1 {
		t.Fatalf("expected 1 EnvModelPercentage, got %d", len(wrapper.EnvModelPercentages))
	}

	emp := wrapper.EnvModelPercentages[0]
	if emp.Name != "ENV-MODEL" {
		t.Errorf("expected Name 'ENV-MODEL', got %s", emp.Name)
	}

	// Empty strings should remain empty even with toHumanReadableForm = true
	if emp.LastKnownGood != "" {
		t.Errorf("expected empty LastKnownGood, got %s", emp.LastKnownGood)
	}
	if emp.IntermediateVersion != "" {
		t.Errorf("expected empty IntermediateVersion, got %s", emp.IntermediateVersion)
	}
}
