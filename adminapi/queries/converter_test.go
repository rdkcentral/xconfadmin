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
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	logupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/stretchr/testify/assert"
)

func TestNullifyUnwantedFieldsPermanentTelemetryProfile_Basic(t *testing.T) {
	profile := &logupload.PermanentTelemetryProfile{
		ApplicationType: "stb",
		TelemetryProfile: []logupload.TelemetryElement{
			{
				ID:        "test-id",
				Component: "test-component",
				Header:    "test-header",
			},
		},
	}

	result := NullifyUnwantedFieldsPermanentTelemetryProfile(profile)

	assert.NotNil(t, result)
	assert.Empty(t, result.ApplicationType)
	if len(result.TelemetryProfile) > 0 {
		assert.Empty(t, result.TelemetryProfile[0].ID)
		assert.Empty(t, result.TelemetryProfile[0].Component)
	}
}

func TestConvertFirmwareConfigToFirmwareConfigResponse_Full(t *testing.T) {
	config := &coreef.FirmwareConfig{
		ID:                       "config-id",
		Updated:                  123456789,
		Description:              "Test Config",
		SupportedModelIds:        []string{"MODEL1"},
		FirmwareFilename:         "firmware.bin",
		FirmwareVersion:          "1.0",
		ApplicationType:          "stb",
		FirmwareDownloadProtocol: "http",
		FirmwareLocation:         "http://example.com/firmware.bin",
		UpgradeDelay:             60,
		RebootImmediately:        true,
		MandatoryUpdate:          false,
		Properties:               map[string]string{"key": "value"},
	}

	response := ConvertFirmwareConfigToFirmwareConfigResponse(config)

	assert.NotNil(t, response)
	assert.Equal(t, config.ID, response.ID)
	assert.Equal(t, config.FirmwareVersion, response.FirmwareVersion)
	assert.Equal(t, config.ApplicationType, response.ApplicationType)
}

func TestConvertIpRuleBeanToIpRuleBeanResponse_WithConfig(t *testing.T) {
	bean := &coreef.IpRuleBean{
		Id:   "rule-id",
		Name: "Test Rule",
		FirmwareConfig: &coreef.FirmwareConfig{
			ID:              "config-id",
			FirmwareVersion: "1.0",
		},
		IpAddressGroup: &shared.IpAddressGroup{
			Id:   "group-id",
			Name: "Test Group",
		},
		EnvironmentId: "PROD",
		ModelId:       "MODEL1",
	}

	response := ConvertIpRuleBeanToIpRuleBeanResponse(bean)

	assert.NotNil(t, response)
	assert.Equal(t, bean.Id, response.Id)
	assert.NotNil(t, response.FirmwareConfig)
	assert.False(t, response.Noop)
}

func TestConvertIpRuleBeanToIpRuleBeanResponse_WithoutConfig(t *testing.T) {
	bean := &coreef.IpRuleBean{
		Id:            "rule-id",
		Name:          "Test Rule",
		EnvironmentId: "PROD",
	}

	response := ConvertIpRuleBeanToIpRuleBeanResponse(bean)

	assert.NotNil(t, response)
	assert.Nil(t, response.FirmwareConfig)
	assert.True(t, response.Noop)
}

func TestConvertMacRuleBeanToMacRuleBeanResponse_WithConfig(t *testing.T) {
	models := []string{"MODEL1"}
	bean := &coreef.MacRuleBean{
		Id:   "mac-rule-id",
		Name: "Mac Rule",
		FirmwareConfig: &coreef.FirmwareConfig{
			ID: "config-id",
		},
		MacAddresses:     "AA:BB:CC:DD:EE:FF",
		TargetedModelIds: &models,
	}

	response := ConvertMacRuleBeanToMacRuleBeanResponse(bean)

	assert.NotNil(t, response)
	assert.NotNil(t, response.FirmwareConfig)
	assert.False(t, response.Noop)
}

func TestConvertMacRuleBeanToMacRuleBeanResponse_WithoutConfig(t *testing.T) {
	bean := &coreef.MacRuleBean{
		Id:           "mac-rule-id",
		Name:         "Mac Rule",
		MacAddresses: "AA:BB:CC:DD:EE:FF",
	}

	response := ConvertMacRuleBeanToMacRuleBeanResponse(bean)

	assert.NotNil(t, response)
	assert.Nil(t, response.FirmwareConfig)
	assert.True(t, response.Noop)
}

func TestConvertEnvModelRuleBeanToEnvModelRuleBeanResponse_Full(t *testing.T) {
	bean := &coreef.EnvModelBean{
		Id:   "env-model-id",
		Name: "Env Model Rule",
		FirmwareConfig: &coreef.FirmwareConfig{
			ID: "config-id",
		},
		EnvironmentId: "PROD",
		ModelId:       "MODEL1",
	}

	response := ConvertEnvModelRuleBeanToEnvModelRuleBeanResponse(bean)

	assert.NotNil(t, response)
	assert.Equal(t, bean.Id, response.Id)
	assert.NotNil(t, response.FirmwareConfig)
}
