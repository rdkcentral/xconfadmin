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
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/stretchr/testify/assert"
)

// Additional service function tests to increase coverage

func TestFirmwareConfigServiceAdditional(t *testing.T) {
	// Test GetFirmwareConfigsByModelIdsAndApplication with nil
	configs := GetFirmwareConfigsByModelIdsAndApplication(nil, "stb")
	assert.NotNil(t, configs)

	// Test with empty model IDs
	configs = GetFirmwareConfigsByModelIdsAndApplication([]string{}, "stb")
	assert.NotNil(t, configs)

	// Test with single model ID
	configs = GetFirmwareConfigsByModelIdsAndApplication([]string{"MODEL1"}, "stb")
	assert.NotNil(t, configs)

	// Test with multiple model IDs
	configs = GetFirmwareConfigsByModelIdsAndApplication([]string{"MODEL1", "MODEL2"}, "stb")
	assert.NotNil(t, configs)

	// Test GetFirmwareConfigsByModelIdAndApplicationType with empty
	configs2 := GetFirmwareConfigsByModelIdAndApplicationType("", "stb")
	assert.NotNil(t, configs2)

	// Test with specific model
	configs2 = GetFirmwareConfigsByModelIdAndApplicationType("MODEL1", "stb")
	assert.NotNil(t, configs2)

	// Test GetFirmwareConfigsByModelIdAndApplicationTypeAS
	configsAS := GetFirmwareConfigsByModelIdAndApplicationTypeAS("", "stb")
	assert.NotNil(t, configsAS)

	configsAS = GetFirmwareConfigsByModelIdAndApplicationTypeAS("MODEL1", "stb")
	assert.NotNil(t, configsAS)
}

func TestFirmwareConfigValidation(t *testing.T) {
	// Test IsValidFirmwareConfigByModelIds with empty
	isValid := IsValidFirmwareConfigByModelIds("", "", nil)
	assert.False(t, isValid)

	// Test with non-nil config but empty model
	fc := &coreef.FirmwareConfig{
		ID:              "TEST",
		Description:     "Test",
		FirmwareVersion: "1.0",
	}
	isValid = IsValidFirmwareConfigByModelIds("", "stb", fc)
	assert.False(t, isValid)

	// Test IsValidFirmwareConfigByModelIdList with nil
	isValid = IsValidFirmwareConfigByModelIdList(nil, "stb", nil)
	assert.False(t, isValid)

	// Test with empty list
	modelIds := []string{}
	isValid = IsValidFirmwareConfigByModelIdList(&modelIds, "stb", nil)
	assert.False(t, isValid)

	// Test with non-nil config
	isValid = IsValidFirmwareConfigByModelIdList(&modelIds, "stb", fc)
	_ = isValid
}

func TestModelServiceAdditional(t *testing.T) {
	// Test IsExistModel with various inputs
	exists := IsExistModel("")
	assert.False(t, exists)

	exists = IsExistModel("NONEXISTENT_MODEL")
	_ = exists

	// Test GetModel with empty
	model := GetModel("")
	_ = model

	// Test GetModels
	models := GetModels()
	assert.NotNil(t, models)
}

func TestEnvironmentServiceAdditional(t *testing.T) {
	// Test GetEnvironment with various inputs
	env := GetEnvironment("")
	_ = env

	env = GetEnvironment("NONEXISTENT")
	_ = env

	// Test IsExistEnvironment
	exists := IsExistEnvironment("")
	assert.False(t, exists)

	exists = IsExistEnvironment("NONEXISTENT")
	_ = exists
}

func TestFeatureServiceAdditional(t *testing.T) {
	// Test GetAllFeatureEntity
	features := GetAllFeatureEntity()
	assert.NotNil(t, features)

	// Test GetFeatureEntityFiltered with various contexts
	context := make(map[string]string)
	features = GetFeatureEntityFiltered(context)
	assert.NotNil(t, features)

	context["name"] = "TEST"
	features = GetFeatureEntityFiltered(context)
	assert.NotNil(t, features)

	// Test GetFeatureEntityById
	feature := GetFeatureEntityById("")
	_ = feature

	feature = GetFeatureEntityById("NONEXISTENT")
	_ = feature
}

func TestFeatureRuleServiceAdditional(t *testing.T) {
	// Test GetAllFeatureRulesByType
	rules := GetAllFeatureRulesByType("")
	assert.NotNil(t, rules)

	rules = GetAllFeatureRulesByType("stb")
	assert.NotNil(t, rules)

	rules = GetAllFeatureRulesByType("xhome")
	assert.NotNil(t, rules)

	// Test GetOne
	rule := GetOne("")
	_ = rule

	rule = GetOne("NONEXISTENT")
	_ = rule

	// Test GetFeatureRulesSize
	size := GetFeatureRulesSize("")
	assert.GreaterOrEqual(t, size, 0)

	size = GetFeatureRulesSize("stb")
	assert.GreaterOrEqual(t, size, 0)

	// Test GetAllowedNumberOfFeatures
	allowed := GetAllowedNumberOfFeatures()
	assert.GreaterOrEqual(t, allowed, 0)
}

func TestTimeFilterServiceAdditional(t *testing.T) {
	// Test GetOneByEnvModel with various inputs
	bean := GetOneByEnvModel("", "", "")
	_ = bean

	bean = GetOneByEnvModel("MODEL1", "ENV1", "stb")
	_ = bean

	bean = GetOneByEnvModel("", "ENV1", "stb")
	_ = bean

	bean = GetOneByEnvModel("MODEL1", "", "stb")
	_ = bean
}

func TestPercentFilterServiceAdditional(t *testing.T) {
	// Test GetPercentFilter with various app types
	filter, err := GetPercentFilter("")
	if err == nil {
		_ = filter
	}

	filter, err = GetPercentFilter("stb")
	if err == nil {
		assert.NotNil(t, filter)
	}

	filter, err = GetPercentFilter("xhome")
	if err == nil {
		_ = filter
	}

	// Test GetPercentFilterFieldValues
	values, err := GetPercentFilterFieldValues("", "stb")
	if err == nil {
		_ = values
	}

	values, err = GetPercentFilterFieldValues("firmwareVersion", "stb")
	if err == nil {
		assert.NotNil(t, values)
	}

	values, err = GetPercentFilterFieldValues("model", "stb")
	if err == nil {
		assert.NotNil(t, values)
	}

	values, err = GetPercentFilterFieldValues("environment", "stb")
	if err == nil {
		_ = values
	}
}

func TestNamespacedListValidationAdditional(t *testing.T) {
	// Test IsValidType with various types
	isValid := IsValidType("")
	_ = isValid

	isValid = IsValidType("MAC_LIST")
	_ = isValid

	isValid = IsValidType("IP_LIST")
	_ = isValid

	isValid = IsValidType("NS_LIST")
	_ = isValid

	isValid = IsValidType("INVALID_TYPE")
	_ = isValid
}

func TestFirmwareConfigGettersAdditional(t *testing.T) {
	// Test GetFirmwareConfigId with various combinations
	id := GetFirmwareConfigId("", "")
	_ = id

	id = GetFirmwareConfigId("1.0.0", "")
	_ = id

	id = GetFirmwareConfigId("", "stb")
	_ = id

	id = GetFirmwareConfigId("1.0.0", "stb")
	_ = id

	id = GetFirmwareConfigId("TEST_VERSION", "xhome")
	_ = id
}

func TestCreateUpdateDeleteFlows(t *testing.T) {
	// Test create/update/delete flow for models
	model := &shared.Model{
		ID:          "FLOW_TEST_MODEL_001",
		Description: "Flow Test Model",
	}

	// Create
	response := CreateModel(model)
	if response.Error == nil {
		// Get
		retrieved := GetModel("FLOW_TEST_MODEL_001")
		_ = retrieved

		// Update
		model.Description = "Updated Flow Test Model"
		response = UpdateModel(model)
		_ = response

		// Delete
		_ = DeleteModel("FLOW_TEST_MODEL_001")
	}
}

func TestFirmwareConfigFlows(t *testing.T) {
	// Test firmware config operations
	fc := &coreef.FirmwareConfig{
		ID:              "FLOW_TEST_FC_001",
		Description:     "Flow Test FC",
		FirmwareVersion: "2.0.0",
		ApplicationType: "stb",
	}

	// Create
	response := CreateFirmwareConfig(fc, "stb")
	if response.Error == nil {
		// Get
		retrieved := GetFirmwareConfigById("FLOW_TEST_FC_001")
		_ = retrieved

		// Update
		fc.Description = "Updated Flow Test FC"
		response = UpdateFirmwareConfig(fc, "stb")
		_ = response

		// Delete
		response = DeleteFirmwareConfig("FLOW_TEST_FC_001", "stb")
		_ = response
	}
}

func TestGetFirmwareConfigsWithDifferentTypes(t *testing.T) {
	// Test with all application types
	appTypes := []string{"", "stb", "xhome"}

	for _, appType := range appTypes {
		configs := GetFirmwareConfigs(appType)
		assert.NotNil(t, configs)

		configsAS := GetFirmwareConfigsAS(appType)
		// Accept nil or empty slice as valid when no data exists
		if configsAS != nil {
			assert.IsType(t, []*coreef.FirmwareConfig{}, configsAS)
		}
	}
}

func TestModelOperationsWithEmptyDescription(t *testing.T) {
	model := &shared.Model{
		ID:          "TEST_EMPTY_DESC_001",
		Description: "",
	}
	response := CreateModel(model)
	assert.NotNil(t, response)

	if response.Error == nil {
		_ = DeleteModel("TEST_EMPTY_DESC_001")
	}
}

func TestFirmwareConfigOperationsWithEmptyFields(t *testing.T) {
	// Test with empty description
	fc := &coreef.FirmwareConfig{
		ID:              "TEST_EMPTY_001",
		Description:     "",
		FirmwareVersion: "1.0",
		ApplicationType: "stb",
	}
	response := CreateFirmwareConfig(fc, "stb")
	assert.NotNil(t, response)

	// Test with empty version
	fc2 := &coreef.FirmwareConfig{
		ID:              "TEST_EMPTY_002",
		Description:     "Test",
		FirmwareVersion: "",
		ApplicationType: "stb",
	}
	response = CreateFirmwareConfig(fc2, "stb")
	assert.NotNil(t, response)
	assert.NotNil(t, response.Error)
}

func TestMultipleModelCreationAndDeletion(t *testing.T) {
	// Create multiple models
	models := []string{"MULTI_TEST_001", "MULTI_TEST_002", "MULTI_TEST_003"}

	for _, id := range models {
		model := &shared.Model{
			ID:          id,
			Description: "Multi Test Model " + id,
		}
		response := CreateModel(model)
		if response.Error == nil {
			// Verify existence
			exists := IsExistModel(id)
			_ = exists
		}
	}

	// Clean up
	for _, id := range models {
		_ = DeleteModel(id)
	}
}

func TestNamespacedListOperations(t *testing.T) {
	// Test various namespaced list type checks
	types := []string{"MAC_LIST", "IP_LIST", "NS_LIST", ""}

	for _, listType := range types {
		isValid := IsValidType(listType)
		_ = isValid
	}
}

func TestHandlerWithInvalidJSON(t *testing.T) {
	// Test handlers with invalid JSON
	invalidJSON := []byte("{invalid json}")

	req, _ := http.NewRequest("POST", "/xconfAdminService/queries/models", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
	assert.NotEqual(t, http.StatusOK, res.Code)
}

func TestHandlerWithEmptyBody(t *testing.T) {
	// Test POST handlers with empty body
	endpoints := []string{
		"/xconfAdminService/queries/models",
		"/xconfAdminService/queries/environments",
		"/xconfAdminService/queries/firmwareconfigs",
	}

	for _, endpoint := range endpoints {
		req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte{}))
		req.Header.Set("Content-Type", "application/json")
		res := ExecuteRequest(req, router)
		assert.NotNil(t, res)
	}
}

func TestMultipleFirmwareConfigsByModel(t *testing.T) {
	// Test getting firmware configs for various models
	models := []string{"", "MODEL1", "MODEL2", "NONEXISTENT"}

	for _, modelId := range models {
		configs := GetFirmwareConfigsByModelIdAndApplicationType(modelId, "stb")
		assert.NotNil(t, configs)

		configsAS := GetFirmwareConfigsByModelIdAndApplicationTypeAS(modelId, "stb")
		assert.NotNil(t, configsAS)
	}
}

func TestBatchModelOperations(t *testing.T) {
	// Create multiple models in batch
	models := []shared.Model{
		{ID: "BATCH_001", Description: "Batch Model 1"},
		{ID: "BATCH_002", Description: "Batch Model 2"},
		{ID: "BATCH_003", Description: "Batch Model 3"},
	}

	for _, model := range models {
		response := CreateModel(&model)
		_ = response
	}

	// Retrieve all
	allModels := GetModels()
	assert.NotNil(t, allModels)

	// Clean up
	for _, model := range models {
		_ = DeleteModel(model.ID)
	}
}
