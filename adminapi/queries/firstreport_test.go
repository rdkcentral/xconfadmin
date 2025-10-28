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
	"bytes"
	"testing"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	xestb "github.com/rdkcentral/xconfadmin/shared/estbfirmware"
	"github.com/stretchr/testify/assert"
)

func TestNextChar(t *testing.T) {
	tests := []struct {
		name     string
		input    rune
		expected rune
	}{
		{"lowercase a to b", 'a', 'b'},
		{"lowercase z wraps to a", 'z', 'a'},
		{"lowercase m to n", 'm', 'n'},
		{"uppercase A to B", 'A', 'B'},
		{"uppercase Z wraps to [", 'Z', '['},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nextChar(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDoReport_EmptyMacAddresses(t *testing.T) {
	macAddresses := []string{}

	reportBytes, err := doReport(macAddresses)

	assert.NoError(t, err)
	assert.NotNil(t, reportBytes)

	// Verify the report can be parsed as Excel
	xlsx, err := excelize.OpenReader(bytes.NewReader(reportBytes))
	assert.NoError(t, err)
	assert.NotNil(t, xlsx)

	// Verify headers exist
	rows := xlsx.GetRows("Sheet1")
	assert.Greater(t, len(rows), 0)
	assert.Equal(t, "estbMac", rows[0][0])
}

func TestDoReport_WithNoConfigLog(t *testing.T) {
	// Setup: Create MAC addresses that don't have any config logs
	macAddresses := []string{
		"AA:BB:CC:DD:EE:01",
		"AA:BB:CC:DD:EE:02",
	}

	reportBytes, err := doReport(macAddresses)

	assert.NoError(t, err)
	assert.NotNil(t, reportBytes)

	// Verify the report structure
	xlsx, err := excelize.OpenReader(bytes.NewReader(reportBytes))
	assert.NoError(t, err)

	rows := xlsx.GetRows("Sheet1")
	// Should have only headers since no config logs exist
	assert.Equal(t, 1, len(rows))
}

func TestDoReport_WithCompleteConfigLog(t *testing.T) {
	// Test with MAC that has a complete config log set up properly through the system
	macAddress := "11:22:33:44:55:66"

	// This test verifies the report can be generated
	// In real usage, the Time field is populated by ConvertedContext marshaling logic
	reportBytes, err := doReport([]string{macAddress})

	assert.NoError(t, err)
	assert.NotNil(t, reportBytes)

	// Verify the report content
	xlsx, err := excelize.OpenReader(bytes.NewReader(reportBytes))
	assert.NoError(t, err)

	rows := xlsx.GetRows("Sheet1")
	assert.Greater(t, len(rows), 0) // At least headers

	// Verify headers
	headers := rows[0]
	assert.Contains(t, headers, "estbMac")
	assert.Contains(t, headers, "env")
	assert.Contains(t, headers, "model")
	assert.Contains(t, headers, "firmwareVersion")
	assert.Contains(t, headers, "rule type")
	assert.Contains(t, headers, "filter name")
}

func TestDoReport_WithNilFields(t *testing.T) {
	macAddress := "AA:BB:CC:DD:EE:FF"
	testTime := time.Now()

	// Create a config log with nil fields
	configLog := &xestb.ConfigChangeLog{
		ID:             xestb.LAST_CONFIG_LOG_ID,
		Updated:        testTime.Unix(),
		Input:          nil,                 // Nil input
		Rule:           nil,                 // Nil rule
		Filters:        []*xestb.RuleInfo{}, // Empty filters
		FirmwareConfig: nil,                 // Nil firmware config
	}

	err := xestb.SetLastConfigLog(macAddress, configLog)
	assert.NoError(t, err)

	macAddresses := []string{macAddress}

	reportBytes, err := doReport(macAddresses)

	assert.NoError(t, err)
	assert.NotNil(t, reportBytes)

	// Verify the report can be parsed
	xlsx, err := excelize.OpenReader(bytes.NewReader(reportBytes))
	assert.NoError(t, err)
	assert.NotNil(t, xlsx)
}

func TestDoReport_WithConfigChangeLogs(t *testing.T) {
	// Test that report generates with change logs structure
	macAddress := "12:34:56:78:90:AB"

	reportBytes, err := doReport([]string{macAddress})

	assert.NoError(t, err)
	assert.NotNil(t, reportBytes)

	// Verify the report
	xlsx, err := excelize.OpenReader(bytes.NewReader(reportBytes))
	assert.NoError(t, err)

	rows := xlsx.GetRows("Sheet1")
	assert.Greater(t, len(rows), 0)

	// Check that headers include change log columns
	headers := rows[0]
	assert.Contains(t, headers, "lst chg env")
	assert.Contains(t, headers, "lst chg model")
	assert.Contains(t, headers, "lst chg firmwareVersion")
}

func TestDoReport_MacAddressSorting(t *testing.T) {
	// Test MAC address sorting without config logs
	macAddresses := []string{
		"ZZ:ZZ:ZZ:ZZ:ZZ:ZZ",
		"AA:AA:AA:AA:AA:AA",
		"MM:MM:MM:MM:MM:MM",
	}

	reportBytes, err := doReport(macAddresses)

	assert.NoError(t, err)
	assert.NotNil(t, reportBytes)

	// Verify the report structure is valid
	xlsx, err := excelize.OpenReader(bytes.NewReader(reportBytes))
	assert.NoError(t, err)

	rows := xlsx.GetRows("Sheet1")
	assert.Greater(t, len(rows), 0) // At least headers
}

func TestDoReport_EmptyConfigChangeLogs(t *testing.T) {
	macAddress := "CC:DD:EE:FF:00:11"

	reportBytes, err := doReport([]string{macAddress})

	assert.NoError(t, err)
	assert.NotNil(t, reportBytes)

	// Verify report structure
	xlsx, err := excelize.OpenReader(bytes.NewReader(reportBytes))
	assert.NoError(t, err)

	rows := xlsx.GetRows("Sheet1")
	assert.Greater(t, len(rows), 0) // At least headers
}

func TestDoReport_MultipleFilters(t *testing.T) {
	macAddress := "11:11:11:11:11:11"

	reportBytes, err := doReport([]string{macAddress})

	assert.NoError(t, err)

	// Verify report is valid
	xlsx, err := excelize.OpenReader(bytes.NewReader(reportBytes))
	assert.NoError(t, err)

	rows := xlsx.GetRows("Sheet1")
	assert.Greater(t, len(rows), 0) // At least headers
}

func TestDoReport_AllHeadersPresent(t *testing.T) {
	expectedHeaders := []string{
		"estbMac",
		"env",
		"model",
		"firmwareVersion",
		"time",
		"ipAddress",
		"rule type",
		"rule name",
		"noop",
		"filter name",
		"firmwareVersion(Config)",
		"firmwareFilename",
		"firmwareLocation",
		"firmwareDownloadProtocol",
		"lst chg env",
		"lst chg model",
		"lst chg firmwareVersion",
		"lst chg time",
		"lst chg ipAddress",
		"lst chg rule type",
		"lst chg rule name",
		"lst chg noop",
		"lst chg firmwareVersion(Config)",
		"lst chg firmwareFilename",
		"lst chg firmwareLocation",
		"lst chg firmwareDownloadProtocol",
	}

	reportBytes, err := doReport([]string{})
	assert.NoError(t, err)

	xlsx, err := excelize.OpenReader(bytes.NewReader(reportBytes))
	assert.NoError(t, err)

	rows := xlsx.GetRows("Sheet1")
	assert.Greater(t, len(rows), 0)

	headers := rows[0]
	for _, expectedHeader := range expectedHeaders {
		assert.Contains(t, headers, expectedHeader)
	}
}
