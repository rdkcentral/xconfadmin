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
	"github.com/stretchr/testify/assert"
)

func TestGetFirmwareConfigs_AllTypes(t *testing.T) {
	// Test get all firmware configs with empty type
	result := GetFirmwareConfigs("")
	assert.NotNil(t, result)
	assert.IsType(t, []*coreef.FirmwareConfigResponse{}, result)

	// Test with specific type
	result = GetFirmwareConfigs("stb")
	assert.NotNil(t, result)
}

func TestGetFirmwareConfigById_NonExistent(t *testing.T) {
	result := GetFirmwareConfigById("NON_EXISTENT_ID")
	// May return nil if not found
	_ = result
}

func TestGetFirmwareConfigsAS_Empty(t *testing.T) {
	result := GetFirmwareConfigsAS("")
	// Accept nil or empty slice when database has no data
	if result != nil {
		assert.IsType(t, []*coreef.FirmwareConfig{}, result)
	}
}

func TestGetFirmwareConfigsAS_WithType(t *testing.T) {
	result := GetFirmwareConfigsAS("stb")
	assert.NotNil(t, result)
}

func TestGetFirmwareConfigByIdAS_NonExistent(t *testing.T) {
	result := GetFirmwareConfigByIdAS("NON_EXISTENT")
	_ = result
}

func TestGetFirmwareConfigsByModelIdAndApplicationType_NonExistent(t *testing.T) {
	result := GetFirmwareConfigsByModelIdAndApplicationType("NON_EXISTENT_MODEL", "stb")
	assert.NotNil(t, result)
}

func TestGetFirmwareConfigsByModelIdAndApplicationTypeAS_NonExistent(t *testing.T) {
	result := GetFirmwareConfigsByModelIdAndApplicationTypeAS("NON_EXISTENT_MODEL", "stb")
	assert.NotNil(t, result)
}

func TestGetGlobalPercentageIdByApplication_STB(t *testing.T) {
	result := GetGlobalPercentageIdByApplication(shared.STB)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "GLOBAL_PERCENT")
}

func TestGetGlobalPercentageIdByApplication_Other(t *testing.T) {
	result := GetGlobalPercentageIdByApplication("xhome")
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "GLOBAL_PERCENT")
	assert.Contains(t, result, "XHOME")
}
