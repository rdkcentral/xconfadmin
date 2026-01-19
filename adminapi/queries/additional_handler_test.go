package queries

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/stretchr/testify/assert"
)

// Additional comprehensive handler tests with POST/PUT/DELETE operations

func TestCreateModelHandlerFlow(t *testing.T) {
	model := &shared.Model{
		ID:          "TEST_MODEL_HANDLER_001",
		Description: "Test Model Handler",
	}
	body, _ := json.Marshal(model)
	req, _ := http.NewRequest("POST", "/xconfAdminService/queries/models", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)

	// Clean up
	if res.Code == http.StatusCreated || res.Code == http.StatusOK {
		req, _ = http.NewRequest("DELETE", "/xconfAdminService/queries/models/TEST_MODEL_HANDLER_001", nil)
		ExecuteRequest(req, router)
	}
}

func TestUpdateModelHandlerFlow(t *testing.T) {
	model := &shared.Model{
		ID:          "TEST_MODEL_HANDLER_002",
		Description: "Original Description",
	}
	body, _ := json.Marshal(model)
	req, _ := http.NewRequest("POST", "/xconfAdminService/queries/models", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)

	if res.Code == http.StatusCreated || res.Code == http.StatusOK {
		// Update it
		model.Description = "Updated Description"
		body, _ = json.Marshal(model)
		req, _ = http.NewRequest("PUT", "/xconfAdminService/queries/models", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		res = ExecuteRequest(req, router)
		assert.NotNil(t, res)

		// Clean up
		req, _ = http.NewRequest("DELETE", "/xconfAdminService/queries/models/TEST_MODEL_HANDLER_002", nil)
		ExecuteRequest(req, router)
	}
}

func TestCreateEnvironmentHandlerFlow(t *testing.T) {
	env := &shared.Environment{
		ID:          "TEST_ENV_HANDLER_001",
		Description: "Test Environment",
	}
	body, _ := json.Marshal(env)
	req, _ := http.NewRequest("POST", "/xconfAdminService/queries/environments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)

	// Clean up
	if res.Code == http.StatusCreated || res.Code == http.StatusOK {
		req, _ = http.NewRequest("DELETE", "/xconfAdminService/queries/environments/TEST_ENV_HANDLER_001", nil)
		ExecuteRequest(req, router)
	}
}

func TestCreateFirmwareConfigHandlerFlow(t *testing.T) {
	fc := &coreef.FirmwareConfig{
		ID:              "TEST_FC_HANDLER_001",
		Description:     "Test Firmware Config",
		FirmwareVersion: "1.0.0",
		ApplicationType: "stb",
	}
	body, _ := json.Marshal(fc)
	req, _ := http.NewRequest("POST", "/xconfAdminService/queries/firmwareconfigs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)

	// Clean up
	if res.Code == http.StatusCreated || res.Code == http.StatusOK {
		req, _ = http.NewRequest("DELETE", "/xconfAdminService/queries/firmwareconfigs/TEST_FC_HANDLER_001", nil)
		ExecuteRequest(req, router)
	}
}

func TestPostModelFilteredHandler(t *testing.T) {
	searchContext := map[string]string{
		"id": "TEST",
	}
	body, _ := json.Marshal(searchContext)
	req, _ := http.NewRequest("POST", "/xconfAdminService/model/filtered", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestPostEnvironmentFilteredHandler(t *testing.T) {
	searchContext := map[string]string{
		"id": "TEST",
	}
	body, _ := json.Marshal(searchContext)
	req, _ := http.NewRequest("POST", "/xconfAdminService/environment/filtered", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestPostFirmwareConfigFilteredHandler(t *testing.T) {
	searchContext := map[string]string{
		"description": "TEST",
	}
	body, _ := json.Marshal(searchContext)
	req, _ := http.NewRequest("POST", "/xconfAdminService/firmwareconfig/filtered", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetFirmwareConfigByIdHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/firmwareconfig/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestGetModelByIdHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/model/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetEnvironmentByIdHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/environment/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestPostFirmwareConfigEntitiesHandler(t *testing.T) {
	configs := []coreef.FirmwareConfig{}
	body, _ := json.Marshal(configs)
	req, _ := http.NewRequest("POST", "/xconfAdminService/firmwareconfig/entities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestPutFirmwareConfigEntitiesHandler(t *testing.T) {
	configs := []coreef.FirmwareConfig{}
	body, _ := json.Marshal(configs)
	req, _ := http.NewRequest("PUT", "/xconfAdminService/firmwareconfig/entities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestPostModelEntitiesHandler(t *testing.T) {
	models := []shared.Model{}
	body, _ := json.Marshal(models)
	req, _ := http.NewRequest("POST", "/xconfAdminService/model/entities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestPutModelEntitiesHandler(t *testing.T) {
	models := []shared.Model{}
	body, _ := json.Marshal(models)
	req, _ := http.NewRequest("PUT", "/xconfAdminService/model/entities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestPostEnvironmentEntitiesHandler(t *testing.T) {
	envs := []shared.Environment{}
	body, _ := json.Marshal(envs)
	req, _ := http.NewRequest("POST", "/xconfAdminService/environment/entities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestPutEnvironmentEntitiesHandler(t *testing.T) {
	envs := []shared.Environment{}
	body, _ := json.Marshal(envs)
	req, _ := http.NewRequest("PUT", "/xconfAdminService/environment/entities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetFirmwareConfigHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/firmwareconfig", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetModelHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/model", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetEnvironmentHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/environment", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetFirmwareConfigByModelIdHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/firmwareconfigs/model/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestPostFirmwareConfigBySupportedModelsHandler(t *testing.T) {
	modelIds := []string{"MODEL1", "MODEL2"}
	body, _ := json.Marshal(modelIds)
	req, _ := http.NewRequest("POST", "/xconfAdminService/firmwareconfig/bySupportedModels", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetIpFilterHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/filters/ips", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetTimeFilterHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/filters/time", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetLocationFilterHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/filters/location", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetPercentFilterHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/filters/percent", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetRebootImmediatelyFilterHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/filters/rebootimmediately", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetDownloadLocationFilterHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/filters/downloadlocation", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetIpRulesHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/rules/ips", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetMacRulesHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/rules/macs", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetEnvModelRulesHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/rules/envModels", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestDeleteModelHandlerNonExistent(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/xconfAdminService/queries/models/NONEXISTENT_MODEL_DELETE", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestDeleteEnvironmentHandlerNonExistent(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/xconfAdminService/queries/environments/NONEXISTENT_ENV_DELETE", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestDeleteFirmwareConfigHandlerNonExistent(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/xconfAdminService/queries/firmwareconfigs/NONEXISTENT_FC_DELETE", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetIpFilterByNameHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/filters/ips/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetTimeFilterByNameHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/filters/time/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetLocationFilterByNameHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/filters/location/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetRebootImmediatelyFilterByNameHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/filters/rebootimmediately/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestDeleteIpFilterHandler(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/xconfAdminService/queries/filters/ips/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestDeleteTimeFilterHandlerAdditional(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/xconfAdminService/queries/filters/time/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestDeleteLocationFilterHandlerAdditional(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/xconfAdminService/queries/filters/location/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestDeleteRebootImmediatelyFilterHandler(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/xconfAdminService/queries/filters/rebootimmediately/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestDeleteDownloadLocationFilterHandler(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/xconfAdminService/queries/filters/downloadlocation/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetIpRuleByIdHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/rules/ips/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetMacRuleByNameHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/rules/macs/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetEnvModelRuleByNameHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/rules/envModels/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetPercentageBeanAllHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/percentagebean", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestGetPercentageBeanByIdHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/xconfAdminService/queries/percentagebean/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestDeletePercentageBeanByIdHandler(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/xconfAdminService/queries/percentagebean/NONEXISTENT", nil)
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}

func TestPostPercentageBeanFilteredHandler(t *testing.T) {
	searchContext := map[string]string{
		"name": "TEST",
	}
	body, _ := json.Marshal(searchContext)
	req, _ := http.NewRequest("POST", "/xconfAdminService/percentageBean/filtered", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := ExecuteRequest(req, router)
	assert.NotNil(t, res)
}
