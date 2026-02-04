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
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
)

func TestSingletonFilterValue_Clone(t *testing.T) {
	original := &SingletonFilterValue{
		ID: "PERCENT_FILTER_VALUE",
		PercentFilterValue: &PercentFilterValue{
			ID:         "PERCENT_FILTER_VALUE",
			Percentage: 75.0,
		},
	}

	cloned, err := original.Clone()
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	if cloned == nil {
		t.Fatal("expected non-nil cloned value")
	}

	if cloned.ID != original.ID {
		t.Error("ID mismatch in clone")
	}

	// Verify it's a deep copy
	cloned.ID = "MODIFIED"
	if original.ID == "MODIFIED" {
		t.Error("Clone is not independent - modifying clone affected original")
	}
}

func TestNewSingletonFilterValueInf(t *testing.T) {
	result := NewSingletonFilterValueInf()

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	sfv, ok := result.(*SingletonFilterValue)
	if !ok {
		t.Fatal("expected *SingletonFilterValue type")
	}

	if sfv == nil {
		t.Fatal("expected non-nil SingletonFilterValue")
	}
}

func TestSingletonFilterValue_IsDownloadLocationRoundRobinFilterValue(t *testing.T) {
	tests := []struct {
		id       string
		expected bool
	}{
		{ROUND_ROBIN_FILTER_SINGLETON_ID, true},
		{"XHOME_" + ROUND_ROBIN_FILTER_SINGLETON_ID, true},
		{PERCENT_FILTER_SINGLETON_ID, false},
		{"SOME_OTHER_ID", false},
		{"", false},
	}

	for _, test := range tests {
		sfv := &SingletonFilterValue{ID: test.id}
		result := sfv.IsDownloadLocationRoundRobinFilterValue()
		if result != test.expected {
			t.Errorf("IsDownloadLocationRoundRobinFilterValue(%s): expected %v, got %v", test.id, test.expected, result)
		}
	}
}

func TestSingletonFilterValue_UnmarshalJSON_PercentFilter(t *testing.T) {
	jsonData := `{
		"id": "PERCENT_FILTER_VALUE",
		"type": "com.comcast.xconf.estbfirmware.PercentFilterValue",
		"percentage": 80.0,
		"envModelPercentages": {}
	}`

	var sfv SingletonFilterValue
	err := json.Unmarshal([]byte(jsonData), &sfv)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if sfv.ID != PERCENT_FILTER_SINGLETON_ID {
		t.Errorf("expected ID %s, got %s", PERCENT_FILTER_SINGLETON_ID, sfv.ID)
	}

	if !sfv.IsPercentFilterValue() {
		t.Error("expected IsPercentFilterValue to be true")
	}

	if sfv.PercentFilterValue == nil {
		t.Fatal("expected non-nil PercentFilterValue")
	}

	if sfv.PercentFilterValue.Percentage != 80.0 {
		t.Errorf("expected Percentage 80.0, got %f", sfv.PercentFilterValue.Percentage)
	}
}

func TestSingletonFilterValue_UnmarshalJSON_RoundRobinFilter(t *testing.T) {
	jsonData := `{
		"id": "DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
		"type": "com.comcast.xconf.estbfirmware.DownloadLocationRoundRobinFilterValue"
	}`

	var sfv SingletonFilterValue
	err := json.Unmarshal([]byte(jsonData), &sfv)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if sfv.ID != ROUND_ROBIN_FILTER_SINGLETON_ID {
		t.Errorf("expected ID %s, got %s", ROUND_ROBIN_FILTER_SINGLETON_ID, sfv.ID)
	}

	if !sfv.IsDownloadLocationRoundRobinFilterValue() {
		t.Error("expected IsDownloadLocationRoundRobinFilterValue to be true")
	}

	if sfv.DownloadLocationRoundRobinFilterValue == nil {
		t.Fatal("expected non-nil DownloadLocationRoundRobinFilterValue")
	}
}

func TestSingletonFilterValue_UnmarshalJSON_InvalidID(t *testing.T) {
	jsonData := `{
		"id": "INVALID_FILTER_VALUE",
		"type": "com.comcast.xconf.estbfirmware.SomeFilterValue"
	}`

	var sfv SingletonFilterValue
	err := json.Unmarshal([]byte(jsonData), &sfv)
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}

	if err.Error() != "Invalid ID for SingletonFilterValue: "+jsonData {
		t.Logf("Got error: %v", err)
	}
}

func TestSingletonFilterValue_MarshalJSON_PercentFilter(t *testing.T) {
	sfv := &SingletonFilterValue{
		ID: PERCENT_FILTER_SINGLETON_ID,
		PercentFilterValue: &PercentFilterValue{
			ID:         PERCENT_FILTER_SINGLETON_ID,
			Percentage: 90.0,
		},
	}

	data, err := json.Marshal(sfv)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}

	// Verify it's valid JSON and contains the PercentFilterValue data
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	if result["id"] != PERCENT_FILTER_SINGLETON_ID {
		t.Error("ID not properly marshaled")
	}
}

func TestSingletonFilterValue_MarshalJSON_RoundRobinFilter(t *testing.T) {
	sfv := &SingletonFilterValue{
		ID: ROUND_ROBIN_FILTER_SINGLETON_ID,
		DownloadLocationRoundRobinFilterValue: &coreef.DownloadLocationRoundRobinFilterValue{
			ID: ROUND_ROBIN_FILTER_SINGLETON_ID,
		},
	}

	data, err := json.Marshal(sfv)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestSingletonFilterValue_MarshalJSON_Invalid(t *testing.T) {
	// Neither subtype is set
	sfv := &SingletonFilterValue{
		ID: "SOME_ID",
	}

	_, err := json.Marshal(sfv)
	if err == nil {
		t.Fatal("expected error for invalid SingletonFilterValue")
	}
}

func TestGetRoundRobinIdByApplication_STB(t *testing.T) {
	id := GetRoundRobinIdByApplication(core.STB)

	if id != ROUND_ROBIN_FILTER_SINGLETON_ID {
		t.Errorf("expected %s for STB, got %s", ROUND_ROBIN_FILTER_SINGLETON_ID, id)
	}
}

func TestGetRoundRobinIdByApplication_NonSTB(t *testing.T) {
	tests := []struct {
		appType  string
		expected string
	}{
		{"xhome", "XHOME_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"},
		{"rdkcloud", "RDKCLOUD_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"},
		{"sky", "SKY_DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"},
	}

	for _, test := range tests {
		result := GetRoundRobinIdByApplication(test.appType)
		if result != test.expected {
			t.Errorf("GetRoundRobinIdByApplication(%s): expected %s, got %s", test.appType, test.expected, result)
		}
	}
}
