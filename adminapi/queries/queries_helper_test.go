/**
 * Copyright 2023 Comcast Cable Communications Management, LLC
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
	"testing"

	"github.com/stretchr/testify/assert"

	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

// TestNullifyUnwantedFieldsPermanentTelemetryProfile_EmptyProfile tests with empty TelemetryProfile array
func TestNullifyUnwantedFieldsPermanentTelemetryProfile_EmptyProfile(t *testing.T) {
	t.Parallel()
	profile := &logupload.PermanentTelemetryProfile{
		ApplicationType:  "stb",
		TelemetryProfile: []logupload.TelemetryElement{},
	}

	result := NullifyUnwantedFieldsPermanentTelemetryProfile(profile)

	assert.NotNil(t, result)
	assert.Equal(t, "", result.ApplicationType)
	assert.Equal(t, 0, len(result.TelemetryProfile))
}

// TestNullifyUnwantedFieldsPermanentTelemetryProfile_WithElements tests with populated TelemetryProfile array
func TestNullifyUnwantedFieldsPermanentTelemetryProfile_WithElements(t *testing.T) {
	t.Parallel()
	profile := &logupload.PermanentTelemetryProfile{
		ApplicationType: "stb",
		TelemetryProfile: []logupload.TelemetryElement{
			{
				ID:               "element-1",
				Component:        "component-1",
				Header:           "header-1",
				Content:          "content-1",
				Type:             "type-1",
				PollingFrequency: "60",
			},
			{
				ID:               "element-2",
				Component:        "component-2",
				Header:           "header-2",
				Content:          "content-2",
				Type:             "type-2",
				PollingFrequency: "120",
			},
		},
	}

	result := NullifyUnwantedFieldsPermanentTelemetryProfile(profile)

	assert.NotNil(t, result)
	assert.Equal(t, "", result.ApplicationType)
	assert.Equal(t, 2, len(result.TelemetryProfile))

	// Verify ID and Component are nullified
	assert.Equal(t, "", result.TelemetryProfile[0].ID)
	assert.Equal(t, "", result.TelemetryProfile[0].Component)
	assert.Equal(t, "", result.TelemetryProfile[1].ID)
	assert.Equal(t, "", result.TelemetryProfile[1].Component)

	// Verify other fields are preserved
	assert.Equal(t, "header-1", result.TelemetryProfile[0].Header)
	assert.Equal(t, "content-1", result.TelemetryProfile[0].Content)
	assert.Equal(t, "type-1", result.TelemetryProfile[0].Type)
	assert.Equal(t, "60", result.TelemetryProfile[0].PollingFrequency)

	assert.Equal(t, "header-2", result.TelemetryProfile[1].Header)
	assert.Equal(t, "content-2", result.TelemetryProfile[1].Content)
	assert.Equal(t, "type-2", result.TelemetryProfile[1].Type)
	assert.Equal(t, "120", result.TelemetryProfile[1].PollingFrequency)
}

// TestNullifyUnwantedFieldsPermanentTelemetryProfile_SingleElement tests with single element
func TestNullifyUnwantedFieldsPermanentTelemetryProfile_SingleElement(t *testing.T) {
	t.Parallel()
	profile := &logupload.PermanentTelemetryProfile{
		ApplicationType: "xhome",
		TelemetryProfile: []logupload.TelemetryElement{
			{
				ID:        "single-id",
				Component: "single-component",
				Header:    "single-header",
			},
		},
	}

	result := NullifyUnwantedFieldsPermanentTelemetryProfile(profile)

	assert.NotNil(t, result)
	assert.Equal(t, "", result.ApplicationType)
	assert.Equal(t, 1, len(result.TelemetryProfile))
	assert.Equal(t, "", result.TelemetryProfile[0].ID)
	assert.Equal(t, "", result.TelemetryProfile[0].Component)
	assert.Equal(t, "single-header", result.TelemetryProfile[0].Header)
}

// TestConvertFirmwareConfigToFirmwareConfigResponse_FullConfig tests with fully populated config
func TestConvertFirmwareConfigToFirmwareConfigResponse_FullConfig(t *testing.T) {
	t.Parallel()
	config := &coreef.FirmwareConfig{
		ID:                       "firmware-id-123",
		Updated:                  int64(1234567890),
		Description:              "Test Firmware Config",
		SupportedModelIds:        []string{"MODEL1", "MODEL2", "MODEL3"},
		FirmwareFilename:         "firmware_v1.0.bin",
		FirmwareVersion:          "1.0.0",
		ApplicationType:          "stb",
		FirmwareDownloadProtocol: "http",
		FirmwareLocation:         "http://example.com/firmware",
		Ipv6FirmwareLocation:     "http://[2001:db8::1]/firmware",
		UpgradeDelay:             300,
		RebootImmediately:        true,
		MandatoryUpdate:          false,
		Properties:               map[string]string{"key1": "value1", "key2": "value2"},
	}

	result := ConvertFirmwareConfigToFirmwareConfigResponse(config)

	assert.NotNil(t, result)
	assert.Equal(t, "firmware-id-123", result.ID)
	assert.Equal(t, int64(1234567890), result.Updated)
	assert.Equal(t, "Test Firmware Config", result.Description)
	assert.Equal(t, []string{"MODEL1", "MODEL2", "MODEL3"}, result.SupportedModelIds)
	assert.Equal(t, "firmware_v1.0.bin", result.FirmwareFilename)
	assert.Equal(t, "1.0.0", result.FirmwareVersion)
	assert.Equal(t, "stb", result.ApplicationType)
	assert.Equal(t, "http", result.FirmwareDownloadProtocol)
	assert.Equal(t, "http://example.com/firmware", result.FirmwareLocation)
	assert.Equal(t, "http://[2001:db8::1]/firmware", result.Ipv6FirmwareLocation)
	assert.Equal(t, int64(300), result.UpgradeDelay)
	assert.Equal(t, true, result.RebootImmediately)
	assert.Equal(t, false, result.MandatoryUpdate)
	assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, result.Properties)
}

// TestConvertFirmwareConfigToFirmwareConfigResponse_MinimalConfig tests with minimal config
func TestConvertFirmwareConfigToFirmwareConfigResponse_MinimalConfig(t *testing.T) {
	t.Parallel()
	config := &coreef.FirmwareConfig{
		ID:              "minimal-id",
		FirmwareVersion: "1.0",
	}

	result := ConvertFirmwareConfigToFirmwareConfigResponse(config)

	assert.NotNil(t, result)
	assert.Equal(t, "minimal-id", result.ID)
	assert.Equal(t, "1.0", result.FirmwareVersion)
	assert.Equal(t, int64(0), result.Updated)
	assert.Equal(t, "", result.Description)
	assert.Nil(t, result.SupportedModelIds)
	assert.Equal(t, "", result.FirmwareFilename)
	assert.Equal(t, "", result.ApplicationType)
	assert.Equal(t, "", result.FirmwareDownloadProtocol)
	assert.Equal(t, "", result.FirmwareLocation)
	assert.Equal(t, "", result.Ipv6FirmwareLocation)
	assert.Equal(t, int64(0), result.UpgradeDelay)
	assert.Equal(t, false, result.RebootImmediately)
	assert.Equal(t, false, result.MandatoryUpdate)
	assert.Nil(t, result.Properties)
}

// TestConvertFirmwareConfigToFirmwareConfigResponse_EmptyStrings tests with empty string values
func TestConvertFirmwareConfigToFirmwareConfigResponse_EmptyStrings(t *testing.T) {
	t.Parallel()
	config := &coreef.FirmwareConfig{
		ID:                       "",
		Description:              "",
		SupportedModelIds:        []string{},
		FirmwareFilename:         "",
		FirmwareVersion:          "",
		ApplicationType:          "",
		FirmwareDownloadProtocol: "",
		FirmwareLocation:         "",
		Ipv6FirmwareLocation:     "",
	}

	result := ConvertFirmwareConfigToFirmwareConfigResponse(config)

	assert.NotNil(t, result)
	assert.Equal(t, "", result.ID)
	assert.Equal(t, "", result.Description)
	assert.Equal(t, []string{}, result.SupportedModelIds)
	assert.Equal(t, "", result.FirmwareFilename)
	assert.Equal(t, "", result.FirmwareVersion)
	assert.Equal(t, "", result.ApplicationType)
	assert.Equal(t, "", result.FirmwareDownloadProtocol)
	assert.Equal(t, "", result.FirmwareLocation)
	assert.Equal(t, "", result.Ipv6FirmwareLocation)
}

// TestConvertFirmwareConfigToFirmwareConfigResponse_NilMaps tests with nil properties map
func TestConvertFirmwareConfigToFirmwareConfigResponse_NilMaps(t *testing.T) {
	t.Parallel()
	config := &coreef.FirmwareConfig{
		ID:         "test-id",
		Properties: nil,
	}

	result := ConvertFirmwareConfigToFirmwareConfigResponse(config)

	assert.NotNil(t, result)
	assert.Equal(t, "test-id", result.ID)
	assert.Nil(t, result.Properties)
}

// TestConvertFirmwareConfigToFirmwareConfigResponse_EmptyMaps tests with empty properties map
func TestConvertFirmwareConfigToFirmwareConfigResponse_EmptyMaps(t *testing.T) {
	t.Parallel()
	config := &coreef.FirmwareConfig{
		ID:         "test-id",
		Properties: map[string]string{},
	}

	result := ConvertFirmwareConfigToFirmwareConfigResponse(config)

	assert.NotNil(t, result)
	assert.Equal(t, "test-id", result.ID)
	assert.Equal(t, map[string]string{}, result.Properties)
}

// TestConvertFirmwareConfigToFirmwareConfigResponse_NilSlices tests with nil SupportedModelIds
func TestConvertFirmwareConfigToFirmwareConfigResponse_NilSlices(t *testing.T) {
	t.Parallel()
	config := &coreef.FirmwareConfig{
		ID:                "test-id",
		SupportedModelIds: nil,
	}

	result := ConvertFirmwareConfigToFirmwareConfigResponse(config)

	assert.NotNil(t, result)
	assert.Equal(t, "test-id", result.ID)
	assert.Nil(t, result.SupportedModelIds)
}

// TestConvertFirmwareConfigToFirmwareConfigResponse_BooleanValues tests boolean field combinations
func TestConvertFirmwareConfigToFirmwareConfigResponse_BooleanValues(t *testing.T) {
	t.Parallel()
	// Test case 1: Both true
	config1 := &coreef.FirmwareConfig{
		ID:                "test-1",
		RebootImmediately: true,
		MandatoryUpdate:   true,
	}
	result1 := ConvertFirmwareConfigToFirmwareConfigResponse(config1)
	assert.True(t, result1.RebootImmediately)
	assert.True(t, result1.MandatoryUpdate)

	// Test case 2: Both false
	config2 := &coreef.FirmwareConfig{
		ID:                "test-2",
		RebootImmediately: false,
		MandatoryUpdate:   false,
	}
	result2 := ConvertFirmwareConfigToFirmwareConfigResponse(config2)
	assert.False(t, result2.RebootImmediately)
	assert.False(t, result2.MandatoryUpdate)

	// Test case 3: Mixed
	config3 := &coreef.FirmwareConfig{
		ID:                "test-3",
		RebootImmediately: true,
		MandatoryUpdate:   false,
	}
	result3 := ConvertFirmwareConfigToFirmwareConfigResponse(config3)
	assert.True(t, result3.RebootImmediately)
	assert.False(t, result3.MandatoryUpdate)
}

// TestConvertFirmwareConfigToFirmwareConfigResponse_LargeValues tests with large numeric values
func TestConvertFirmwareConfigToFirmwareConfigResponse_LargeValues(t *testing.T) {
	t.Parallel()
	config := &coreef.FirmwareConfig{
		ID:           "large-values",
		Updated:      int64(9223372036854775807), // Max int64
		UpgradeDelay: 2147483647,                 // Max int32
	}

	result := ConvertFirmwareConfigToFirmwareConfigResponse(config)

	assert.NotNil(t, result)
	assert.Equal(t, "large-values", result.ID)
	assert.Equal(t, int64(9223372036854775807), result.Updated)
	assert.Equal(t, int64(2147483647), result.UpgradeDelay)
}

// TestConvertFirmwareConfigToFirmwareConfigResponse_SpecialCharacters tests with special characters in strings
func TestConvertFirmwareConfigToFirmwareConfigResponse_SpecialCharacters(t *testing.T) {
	t.Parallel()
	config := &coreef.FirmwareConfig{
		ID:               "special-chars-<>&\"'",
		Description:      "Description with special chars: <>&\"'\n\t",
		FirmwareFilename: "firmware-v1.0_beta@2024.bin",
		FirmwareLocation: "http://example.com/path?param=value&other=123",
	}

	result := ConvertFirmwareConfigToFirmwareConfigResponse(config)

	assert.NotNil(t, result)
	assert.Equal(t, "special-chars-<>&\"'", result.ID)
	assert.Equal(t, "Description with special chars: <>&\"'\n\t", result.Description)
	assert.Equal(t, "firmware-v1.0_beta@2024.bin", result.FirmwareFilename)
	assert.Equal(t, "http://example.com/path?param=value&other=123", result.FirmwareLocation)
}
