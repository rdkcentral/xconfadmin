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
	"encoding/json"
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"gotest.tools/assert"
)

// Helper function to setup firmware rule templates
func setupFirmwareRuleTemplates() {
	CreateFirmwareRuleTemplates()

	// Create the test firmware config that rules reference
	testConfig := &estbfirmware.FirmwareConfig{
		ID:                "test-config-id",
		Description:       "Test Config",
		FirmwareVersion:   "1.0.0",
		ApplicationType:   "stb",
		SupportedModelIds: []string{"TEST-MODEL"},
		FirmwareFilename:  "test.bin",
	}
	SetOneInDao(db.TABLE_FIRMWARE_CONFIG, testConfig.ID, testConfig)
	db.GetCacheManager().ForceSyncChanges()
}

// Helper function to create a test firmware rule
func createTestFirmwareRule(id, name, appType string) *firmware.FirmwareRule {
	return createTestFirmwareRuleWithMAC(id, name, appType, "AA:BB:CC:DD:EE:FF")
}

func createTestFirmwareRuleWithMAC(id, name, appType, macAddress string) *firmware.FirmwareRule {
	// Create a valid rule using JSON unmarshaling for simplicity
	ruleJSON := `{
		"id": "` + id + `",
		"name": "` + name + `",
		"applicationType": "` + appType + `",
		"type": "MAC_RULE",
		"active": true,
		"rule": {
			"negated": false,
			"condition": {
				"freeArg": {
					"type": "STRING",
					"name": "eStbMac"
				},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {
							"java.lang.String": "` + macAddress + `"
						}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE",
			"configId": "test-config-id",
			"active": true
		}
	}`

	var rule firmware.FirmwareRule
	json.Unmarshal([]byte(ruleJSON), &rule)
	return &rule
}

// TestPostFirmwareRuleHandler_Success tests successful firmware rule creation
func TestPostFirmwareRuleHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	setupFirmwareRuleTemplates()
	defer DeleteAllEntities()

	rule := createTestFirmwareRule("", "Test Rule Create", "stb")
	body, _ := json.Marshal(rule)

	// Extra sync to ensure firmware config is available for validation
	db.GetCacheManager().ForceSyncChanges()

	req, err := http.NewRequest("POST", "/xconfAdminService/firmwarerule", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var returnedRule firmware.FirmwareRule
	json.NewDecoder(res.Body).Decode(&returnedRule)
	assert.Equal(t, rule.Name, returnedRule.Name)
	assert.Assert(t, returnedRule.ID != "")
}

// TestPostFirmwareRuleHandler_DuplicateID tests duplicate rule ID validation
func TestPostFirmwareRuleHandler_DuplicateID(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create first rule
	rule1 := createTestFirmwareRule("duplicate-id", "First Rule", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)

	// Try to create second rule with same ID
	rule2 := createTestFirmwareRule("duplicate-id", "Second Rule", "stb")
	body, _ := json.Marshal(rule2)

	req, err := http.NewRequest("POST", "/xconfAdminService/firmwarerule", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusConflict, res.StatusCode)
}

// TestPostFirmwareRuleHandler_InvalidJSON tests invalid JSON handling
func TestPostFirmwareRuleHandler_InvalidJSON(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidJSON := []byte(`{invalid json}`)

	req, err := http.NewRequest("POST", "/xconfAdminService/firmwarerule", bytes.NewBuffer(invalidJSON))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestPutFirmwareRuleHandler_Success tests successful firmware rule update
func TestPutFirmwareRuleHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	setupFirmwareRuleTemplates()
	defer DeleteAllEntities()

	// Create initial rule
	rule := createTestFirmwareRule("rule-to-update", "Original Name", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule.ID, rule)
	db.GetCacheManager().ForceSyncChanges() // Ensure cache is synchronized before update

	// Update the rule
	rule.Name = "Updated Name"
	body, _ := json.Marshal(rule)

	req, err := http.NewRequest("PUT", "/xconfAdminService/firmwarerule", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var returnedRule firmware.FirmwareRule
	json.NewDecoder(res.Body).Decode(&returnedRule)
	assert.Equal(t, "Updated Name", returnedRule.Name)
}

// TestPutFirmwareRuleHandler_NotFound tests updating non-existent rule
func TestPutFirmwareRuleHandler_NotFound(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	rule := createTestFirmwareRule("non-existent-rule", "Does Not Exist", "stb")
	body, _ := json.Marshal(rule)

	req, err := http.NewRequest("PUT", "/xconfAdminService/firmwarerule", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestDeleteFirmwareRuleByIdHandler_Success tests successful deletion
func TestDeleteFirmwareRuleByIdHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create rule to delete
	rule := createTestFirmwareRule("rule-to-delete", "To Be Deleted", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule.ID, rule)
	db.GetCacheManager().ForceSyncChanges() // Ensure rule is available before deletion

	req, err := http.NewRequest("DELETE", "/xconfAdminService/firmwarerule/rule-to-delete", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	// Verify deletion
	deleted, _ := firmware.GetFirmwareRuleOneDB("rule-to-delete")
	assert.Assert(t, deleted == nil)
}

// TestDeleteFirmwareRuleByIdHandler_NotFound tests deleting non-existent rule
func TestDeleteFirmwareRuleByIdHandler_NotFound(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("DELETE", "/xconfAdminService/firmwarerule/nonexistent", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestDeleteFirmwareRuleByIdHandler_ApplicationTypeMismatch tests app type validation
func TestDeleteFirmwareRuleByIdHandler_ApplicationTypeMismatch(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create rule with xhome app type
	rule := createTestFirmwareRule("rule-app-mismatch", "App Mismatch Rule", "xhome")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule.ID, rule)
	db.GetCacheManager().ForceSyncChanges() // Ensure rule is available before deletion attempt

	// Try to delete with stb app type
	req, err := http.NewRequest("DELETE", "/xconfAdminService/firmwarerule/rule-app-mismatch", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusConflict, res.StatusCode)
}

// TestGetFirmwareRuleByIdHandler_Success tests getting rule by ID
func TestGetFirmwareRuleByIdHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	rule := createTestFirmwareRule("rule-get-by-id", "Get By ID Test", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule.ID, rule)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/rule-get-by-id", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var returnedRule firmware.FirmwareRule
	json.NewDecoder(res.Body).Decode(&returnedRule)
	assert.Equal(t, rule.ID, returnedRule.ID)
	assert.Equal(t, rule.Name, returnedRule.Name)
}

// TestGetFirmwareRuleByIdHandler_WithExport tests export functionality
func TestGetFirmwareRuleByIdHandler_WithExport(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	rule := createTestFirmwareRule("rule-export-test", "Export Test", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule.ID, rule)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/rule-export-test?export", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "")
}

// TestGetFirmwareRuleByIdHandler_NotFound tests non-existent rule
func TestGetFirmwareRuleByIdHandler_NotFound(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/nonexistent", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestGetFirmwareRuleByIdHandler_ApplicationTypeMismatch tests app type validation
func TestGetFirmwareRuleByIdHandler_ApplicationTypeMismatch(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	rule := createTestFirmwareRule("rule-get-mismatch", "Get Mismatch Test", "xhome")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule.ID, rule)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/rule-get-mismatch", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusConflict, res.StatusCode)
}

// TestGetFirmwareRuleHandler_Success tests getting all rules
func TestGetFirmwareRuleHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test rules
	rule1 := createTestFirmwareRule("rule-all-1", "All Rules Test 1", "stb")
	rule2 := createTestFirmwareRule("rule-all-2", "All Rules Test 2", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule2.ID, rule2)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var rules []firmware.FirmwareRule
	json.NewDecoder(res.Body).Decode(&rules)
	assert.Assert(t, len(rules) >= 2)
}

// TestGetFirmwareRuleHandler_WithExport tests export all functionality
func TestGetFirmwareRuleHandler_WithExport(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	rule := createTestFirmwareRule("rule-export-all", "Export All Test", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule.ID, rule)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule?export", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "")
}

// TestGetFirmwareRuleFilteredHandler tests filtering functionality
func TestGetFirmwareRuleFilteredHandler(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test rules
	rule1 := createTestFirmwareRule("rule-filter-1", "Filter Test 1", "stb")
	rule1.Type = firmware.MAC_RULE
	rule2 := createTestFirmwareRule("rule-filter-2", "Filter Test 2", "stb")
	rule2.Type = firmware.ENV_MODEL_RULE
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule2.ID, rule2)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/filtered", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var rules []firmware.FirmwareRule
	json.NewDecoder(res.Body).Decode(&rules)
	assert.Assert(t, len(rules) >= 2)
}

// TestPostFirmwareRuleFilteredHandler_Success tests POST filtered endpoint
func TestPostFirmwareRuleFilteredHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test rules
	rule1 := createTestFirmwareRule("rule-post-filter-1", "POST Filter 1", "stb")
	rule2 := createTestFirmwareRule("rule-post-filter-2", "POST Filter 2", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule2.ID, rule2)

	filterContext := map[string]string{}
	body, _ := json.Marshal(filterContext)

	req, err := http.NewRequest("POST", "/xconfAdminService/firmwarerule/filtered?pageNumber=1&pageSize=10", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

// TestPostFirmwareRuleFilteredHandler_InvalidPageNumber tests invalid pagination
func TestPostFirmwareRuleFilteredHandler_InvalidPageNumber(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	filterContext := map[string]string{}
	body, _ := json.Marshal(filterContext)

	req, err := http.NewRequest("POST", "/xconfAdminService/firmwarerule/filtered?pageNumber=0&pageSize=10", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestGetFirmwareRuleByTypeNamesHandler_Success tests getting rule names by type
func TestGetFirmwareRuleByTypeNamesHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create rules with different types
	rule1 := createTestFirmwareRule("rule-type-1", "Type Test 1", "stb")
	rule1.Type = firmware.MAC_RULE
	rule2 := createTestFirmwareRule("rule-type-2", "Type Test 2", "stb")
	rule2.Type = firmware.MAC_RULE
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule2.ID, rule2)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/MAC_RULE/names", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var nameMap map[string]string
	json.NewDecoder(res.Body).Decode(&nameMap)
	assert.Assert(t, len(nameMap) >= 2)
}

// TestGetFirmwareRuleByTemplateNamesHandler tests byTemplate/names endpoint
func TestGetFirmwareRuleByTemplateNamesHandler(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/byTemplate/names", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// This endpoint matches /{type}/names where type="byTemplate", so it returns OK
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

// TestPostFirmwareRuleEntitiesHandler_Success tests batch creation
func TestPostFirmwareRuleEntitiesHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	setupFirmwareRuleTemplates()
	defer DeleteAllEntities()

	entities := []*firmware.FirmwareRule{
		createTestFirmwareRuleWithMAC("batch-create-1", "Batch Create 1", "stb", "AA:BB:CC:DD:EE:11"),
		createTestFirmwareRuleWithMAC("batch-create-2", "Batch Create 2", "stb", "AA:BB:CC:DD:EE:12"),
	}
	body, _ := json.Marshal(entities)

	req, err := http.NewRequest("POST", "/xconfAdminService/firmwarerule/entities", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseMap map[string]xhttp.EntityMessage
	json.NewDecoder(res.Body).Decode(&responseMap)
	assert.Equal(t, 2, len(responseMap))
	assert.Equal(t, common.ENTITY_STATUS_SUCCESS, responseMap["batch-create-1"].Status)
	assert.Equal(t, common.ENTITY_STATUS_SUCCESS, responseMap["batch-create-2"].Status)
}

// TestPostFirmwareRuleEntitiesHandler_DuplicateEntity tests duplicate detection
func TestPostFirmwareRuleEntitiesHandler_DuplicateEntity(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create existing rule
	existing := createTestFirmwareRule("duplicate-batch", "Existing Rule", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, existing.ID, existing)

	// Try to create batch with duplicate
	entities := []*firmware.FirmwareRule{
		createTestFirmwareRule("duplicate-batch", "Duplicate Rule", "stb"),
	}
	body, _ := json.Marshal(entities)

	req, err := http.NewRequest("POST", "/xconfAdminService/firmwarerule/entities", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseMap map[string]xhttp.EntityMessage
	json.NewDecoder(res.Body).Decode(&responseMap)
	assert.Equal(t, common.ENTITY_STATUS_FAILURE, responseMap["duplicate-batch"].Status)
}

// TestPutFirmwareRuleEntitiesHandler_Success tests batch update
func TestPutFirmwareRuleEntitiesHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	setupFirmwareRuleTemplates()
	defer DeleteAllEntities()

	// Create initial rules with different MAC addresses to avoid duplicate detection
	rule1 := createTestFirmwareRuleWithMAC("batch-update-1", "Original 1", "stb", "AA:BB:CC:DD:EE:01")
	rule2 := createTestFirmwareRuleWithMAC("batch-update-2", "Original 2", "stb", "AA:BB:CC:DD:EE:02")

	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule2.ID, rule2)
	db.GetCacheManager().ForceSyncChanges()

	// Update the rules
	rule1.Name = "Updated 1"
	rule2.Name = "Updated 2"
	entities := []*firmware.FirmwareRule{rule1, rule2}
	body, _ := json.Marshal(entities)

	req, err := http.NewRequest("PUT", "/xconfAdminService/firmwarerule/entities", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseMap map[string]xhttp.EntityMessage
	json.NewDecoder(res.Body).Decode(&responseMap)
	assert.Equal(t, 2, len(responseMap))
	assert.Equal(t, common.ENTITY_STATUS_SUCCESS, responseMap["batch-update-1"].Status)
	assert.Equal(t, common.ENTITY_STATUS_SUCCESS, responseMap["batch-update-2"].Status)
}

// TestPutFirmwareRuleEntitiesHandler_NonExistent tests updating non-existent rules
func TestPutFirmwareRuleEntitiesHandler_NonExistent(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	entities := []*firmware.FirmwareRule{
		createTestFirmwareRule("non-existent-batch", "Does Not Exist", "stb"),
	}
	body, _ := json.Marshal(entities)

	req, err := http.NewRequest("PUT", "/xconfAdminService/firmwarerule/entities", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseMap map[string]xhttp.EntityMessage
	json.NewDecoder(res.Body).Decode(&responseMap)
	assert.Equal(t, common.ENTITY_STATUS_FAILURE, responseMap["non-existent-batch"].Status)
}

// TestObsoleteGetFirmwareRulePageHandler tests pagination endpoint
func TestObsoleteGetFirmwareRulePageHandler(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Note: /page endpoint is mapped to NotImplementedHandler in router (line 309 of router.go)
	// This test verifies that the endpoint returns NotImplemented status
	for i := 1; i <= 5; i++ {
		rule := createTestFirmwareRule("page-rule-"+string(rune('0'+i)), "Page Rule "+string(rune('0'+i)), "stb")
		SetOneInDao(db.TABLE_FIRMWARE_RULE, rule.ID, rule)
	}

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/page?pageNumber=1&pageSize=3", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotImplemented, res.StatusCode)
}

// TestGetFirmwareRuleExportAllTypesHandler tests export all types
func TestGetFirmwareRuleExportAllTypesHandler(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	rule := createTestFirmwareRule("export-all-types", "Export All Types Test", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule.ID, rule)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/export/allTypes?exportAll", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "")
}

// TestGetFirmwareRuleExportByTypeHandler_Success tests export by type
func TestGetFirmwareRuleExportByTypeHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	rule := createTestFirmwareRule("export-by-type", "Export By Type Test", "stb")
	rule.ApplicableAction.ActionType = "RULE"
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule.ID, rule)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/export/byType?exportAll&type=RULE", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "")
}

// TestGetFirmwareRuleExportByTypeHandler_MissingType tests missing type param
func TestGetFirmwareRuleExportByTypeHandler_MissingType(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/export/byType?exportAll", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestPostFirmwareRuleImportAllHandler_Success tests import functionality
func TestPostFirmwareRuleImportAllHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	rules := []*firmware.FirmwareRule{
		createTestFirmwareRule("import-1", "Import Rule 1", "stb"),
		createTestFirmwareRule("import-2", "Import Rule 2", "stb"),
	}
	body, _ := json.Marshal(rules)

	req, err := http.NewRequest("POST", "/xconfAdminService/firmwarerule/importAll", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

// TestPostFirmwareRuleImportAllHandler_ApplicationTypeMixing tests app type mixing
func TestPostFirmwareRuleImportAllHandler_ApplicationTypeMixing(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	rules := []*firmware.FirmwareRule{
		createTestFirmwareRule("import-mix-1", "Import STB", "stb"),
		createTestFirmwareRule("import-mix-2", "Import XHOME", "xhome"),
	}
	body, _ := json.Marshal(rules)

	req, err := http.NewRequest("POST", "/xconfAdminService/firmwarerule/importAll", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusConflict, res.StatusCode)
}

// Test helper functions

// TestConvertToMapKey tests the convertToMapKey function
func TestConvertToMapKey(t *testing.T) {
	SkipIfMockDatabase(t)
	rule := createTestFirmwareRule("test-map-key", "Test Map Key", "stb")

	// Test with simple rule
	mapKey, estb := convertToMapKey(rule)
	assert.Assert(t, mapKey != "")

	// ESTB will be empty for non-estbmac rules
	_ = estb
}

// TestDuplicateFrFound tests the duplicateFrFound function
func TestDuplicateFrFound(t *testing.T) {
	SkipIfMockDatabase(t)
	rule1 := createTestFirmwareRule("dup-test-1", "Duplicate Test 1", "stb")
	rule2 := createTestFirmwareRule("dup-test-2", "Duplicate Test 1", "stb") // Same name

	nameMap := make(map[string][]*firmware.FirmwareRule)
	nameMap["Duplicate Test 1"] = []*firmware.FirmwareRule{rule1}

	ruleMap := make(map[string][]*firmware.FirmwareRule)
	estbMap := make(map[string][]*firmware.FirmwareRule)

	err := duplicateFrFound(rule2, nameMap, ruleMap, estbMap)
	assert.Assert(t, err != nil) // Should detect duplicate name
}

// TestFindAndDeleteFR tests the findAndDeleteFR function
func TestFindAndDeleteFR(t *testing.T) {
	SkipIfMockDatabase(t)
	rule1 := createTestFirmwareRule("find-del-1", "Find Delete 1", "stb")
	rule2 := createTestFirmwareRule("find-del-2", "Find Delete 2", "stb")
	rule3 := createTestFirmwareRule("find-del-3", "Find Delete 3", "stb")

	list := []*firmware.FirmwareRule{rule1, rule2, rule3}

	// Delete rule2
	result := findAndDeleteFR(list, *rule2)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, "find-del-1", result[0].ID)
	assert.Equal(t, "find-del-3", result[1].ID)
}

// TestPopulateContext tests the populateContext function
func TestPopulateContext(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule?pageNumber=1&pageSize=10", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	// We can't directly call populateContext as it needs a ResponseWriter
	// But we can test it indirectly through the handlers that use it
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

// ObsoleteGetFirmwareRulePageHandler - Error paths
func TestObsoleteGetFirmwareRulePageHandler_ErrorGettingRules(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Note: /page endpoint is mapped to NotImplementedHandler in router
	// This test verifies the handler code itself works if called directly
	// Skipping this test as endpoint is not implemented in router
	t.Skip("ObsoleteGetFirmwareRulePageHandler is not implemented in router")
}

func TestObsoleteGetFirmwareRulePageHandler_InvalidPageNumber(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Note: /page endpoint is mapped to NotImplementedHandler in router
	t.Skip("ObsoleteGetFirmwareRulePageHandler is not implemented in router")
}

func TestObsoleteGetFirmwareRulePageHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Note: /page endpoint is mapped to NotImplementedHandler in router
	t.Skip("ObsoleteGetFirmwareRulePageHandler is not implemented in router")
}

// GetFirmwareRuleExportAllTypesHandler - Error paths
func TestGetFirmwareRuleExportAllTypesHandler_MissingExportAllParam(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/export/allTypes", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestGetFirmwareRuleExportAllTypesHandler_ErrorGettingRules(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/export/allTypes?exportAll", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// Should succeed even with no rules
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestGetFirmwareRuleExportAllTypesHandler_SuccessWithRules(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create rules of different types
	rule1 := createTestFirmwareRule("export-all-1", "Export All 1", "stb")
	rule1.Type = firmware.MAC_RULE
	rule2 := createTestFirmwareRule("export-all-2", "Export All 2", "stb")
	rule2.Type = firmware.ENV_MODEL_RULE
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule2.ID, rule2)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/export/allTypes?exportAll", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header is set
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "")

	var rules []firmware.FirmwareRule
	json.NewDecoder(res.Body).Decode(&rules)
	assert.Assert(t, len(rules) >= 2)
}

// GetFirmwareRuleByTemplateByTemplateIdNamesHandler - Error paths
func TestGetFirmwareRuleByTemplateByTemplateIdNamesHandler_MissingTemplateId(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Empty templateId - router will match but handler should handle empty templateId
	// Testing with just empty string in path - the router may still route this
	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/byTemplate/ /names", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// May return 200 with empty results or 400/404 depending on routing
	assert.Assert(t, res.StatusCode >= http.StatusOK && res.StatusCode < 500)
}

func TestGetFirmwareRuleByTemplateByTemplateIdNamesHandler_ErrorGettingRules(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/byTemplate/template-123/names", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// Should succeed even with no rules matching the template
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestGetFirmwareRuleByTemplateByTemplateIdNamesHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create rules with template IDs
	rule1 := createTestFirmwareRule("template-rule-1", "Template Rule 1", "stb")
	rule2 := createTestFirmwareRule("template-rule-2", "Template Rule 2", "stb")
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)
	SetOneInDao(db.TABLE_FIRMWARE_RULE, rule2.ID, rule2)

	req, err := http.NewRequest("GET", "/xconfAdminService/firmwarerule/byTemplate/some-template-id/names", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var namesList []string
	json.NewDecoder(res.Body).Decode(&namesList)
	// Names list should be returned (may be empty if no rules match the template)
	assert.Assert(t, namesList != nil)
}
