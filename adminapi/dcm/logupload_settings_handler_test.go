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
package dcm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"gotest.tools/assert"
)

// ========== Tests for GetLogUploadSettingsByIdHandler - nil and error conditions ==========

// TestGetLogUploadSettingsByIdHandler_MissingID tests error when ID is missing
func TestGetLogUploadSettingsByIdHandler_MissingID(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req := httptest.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/", nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// Should return 404 as the route doesn't match
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestGetLogUploadSettingsByIdHandler_NilResult tests handling when settings don't exist (nil condition)
func TestGetLogUploadSettingsByIdHandler_NilResult(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req := httptest.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/nonexistent-id", nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestGetLogUploadSettingsByIdHandler_ApplicationTypeMismatch tests when ApplicationType doesn't match (error path)
func TestGetLogUploadSettingsByIdHandler_ApplicationTypeMismatch(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create a formula first
	formula := createFormula("TEST_MODEL_MISMATCH", 1)
	saveFormula(formula, t)

	// Create settings with "xhome" application type
	settings := &logupload.LogUploadSettings{
		ID:                 formula.ID,
		Name:               "XHome Settings",
		UploadRepositoryID: "test-repo",
		ApplicationType:    "xhome",
		Schedule: logupload.Schedule{
			Type:       "CronExpression",
			Expression: "0 0 * * *",
			TimeZone:   "UTC",
		},
	}
	CreateLogUploadSettings(settings, "xhome")

	// Try to access with "stb" application type
	req := httptest.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/"+formula.ID, nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestGetLogUploadSettingsByIdHandler_Success tests successful retrieval
func TestGetLogUploadSettingsByIdHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create a formula first
	formula := createFormula("TEST_MODEL_SUCCESS", 1)
	saveFormula(formula, t)

	// Create settings
	settings := &logupload.LogUploadSettings{
		ID:                 formula.ID,
		Name:               "Test Settings",
		UploadRepositoryID: "test-repo",
		ApplicationType:    "stb",
		Schedule: logupload.Schedule{
			Type:       "CronExpression",
			Expression: "0 0 * * *",
			TimeZone:   "UTC",
		},
	}
	CreateLogUploadSettings(settings, "stb")

	req := httptest.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/"+formula.ID, nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var result logupload.LogUploadSettings
	json.NewDecoder(res.Body).Decode(&result)
	assert.Equal(t, formula.ID, result.ID)
	assert.Equal(t, "Test Settings", result.Name)
}

// ========== Tests for GetLogUploadSettingsHandler - nil conditions ==========

// TestGetLogUploadSettingsHandler_EmptyList tests handling when no settings exist (nil condition)
func TestGetLogUploadSettingsHandler_EmptyList(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req := httptest.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings", nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var settingsList []*logupload.LogUploadSettings
	json.NewDecoder(res.Body).Decode(&settingsList)
	assert.Equal(t, 0, len(settingsList))
}

// TestGetLogUploadSettingsHandler_FilterByApplicationType tests filtering by application type
func TestGetLogUploadSettingsHandler_FilterByApplicationType(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create formulas for different application types
	formulaStb := createFormula("TEST_MODEL_STB", 1)
	saveFormula(formulaStb, t)

	settingsStb := &logupload.LogUploadSettings{
		ID:                 formulaStb.ID,
		Name:               "STB Settings",
		UploadRepositoryID: "test-repo",
		ApplicationType:    "stb",
		Schedule: logupload.Schedule{
			Type:       "CronExpression",
			Expression: "0 0 * * *",
			TimeZone:   "UTC",
		},
	}
	CreateLogUploadSettings(settingsStb, "stb")

	req := httptest.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings", nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var settingsList []*logupload.LogUploadSettings
	json.NewDecoder(res.Body).Decode(&settingsList)

	// Verify only stb settings are returned
	for _, settings := range settingsList {
		assert.Equal(t, "stb", settings.ApplicationType)
	}
}

// ========== Tests for GetLogUploadSettingsSizeHandler - nil conditions ==========

// TestGetLogUploadSettingsSizeHandler_ZeroCount tests size handler with no settings (nil condition)
func TestGetLogUploadSettingsSizeHandler_ZeroCount(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req := httptest.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/size", nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var count int
	json.NewDecoder(res.Body).Decode(&count)
	assert.Equal(t, 0, count)
}

// TestGetLogUploadSettingsSizeHandler_NonZeroCount tests size handler with settings
func TestGetLogUploadSettingsSizeHandler_NonZeroCount(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create multiple settings
	for i := 1; i <= 3; i++ {
		formula := createFormula(fmt.Sprintf("TEST_MODEL_SIZE_%d", i), i)
		saveFormula(formula, t)

		settings := &logupload.LogUploadSettings{
			ID:                 formula.ID,
			Name:               fmt.Sprintf("Settings %d", i),
			UploadRepositoryID: "test-repo",
			ApplicationType:    "stb",
			Schedule: logupload.Schedule{
				Type:       "CronExpression",
				Expression: "0 0 * * *",
				TimeZone:   "UTC",
			},
		}
		CreateLogUploadSettings(settings, "stb")
	}

	req := httptest.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/size", nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var count int
	json.NewDecoder(res.Body).Decode(&count)
	assert.Equal(t, 3, count)
}

// ========== Tests for GetLogUploadSettingsNamesHandler - nil conditions ==========

// TestGetLogUploadSettingsNamesHandler_EmptyList tests names handler with no settings (nil condition)
func TestGetLogUploadSettingsNamesHandler_EmptyList(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req := httptest.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/names", nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var names []string
	json.NewDecoder(res.Body).Decode(&names)
	assert.Equal(t, 0, len(names))
}

// TestGetLogUploadSettingsNamesHandler_WithNames tests names handler with settings
func TestGetLogUploadSettingsNamesHandler_WithNames(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create settings with specific names
	names := []string{"Alpha Settings", "Beta Settings", "Gamma Settings"}
	for i, name := range names {
		formula := createFormula(fmt.Sprintf("TEST_MODEL_NAMES_%d", i), i+1)
		saveFormula(formula, t)

		settings := &logupload.LogUploadSettings{
			ID:                 formula.ID,
			Name:               name,
			UploadRepositoryID: "test-repo",
			ApplicationType:    "stb",
			Schedule: logupload.Schedule{
				Type:       "CronExpression",
				Expression: "0 0 * * *",
				TimeZone:   "UTC",
			},
		}
		CreateLogUploadSettings(settings, "stb")
	}

	req := httptest.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/names", nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var returnedNames []string
	json.NewDecoder(res.Body).Decode(&returnedNames)
	assert.Equal(t, 3, len(returnedNames))
}

// ========== Tests for DeleteLogUploadSettingsByIdHandler - error paths ==========

// TestDeleteLogUploadSettingsByIdHandler_MissingID tests delete with missing ID (error path)
func TestDeleteLogUploadSettingsByIdHandler_MissingID(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req := httptest.NewRequest("DELETE", "/xconfAdminService/dcm/logUploadSettings/", nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// Should return 404 as the route doesn't match
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestDeleteLogUploadSettingsByIdHandler_NonExistent tests delete of non-existent settings (error path)
func TestDeleteLogUploadSettingsByIdHandler_NonExistent(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req := httptest.NewRequest("DELETE", "/xconfAdminService/dcm/logUploadSettings/nonexistent-id", nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestDeleteLogUploadSettingsByIdHandler_Success tests successful delete
func TestDeleteLogUploadSettingsByIdHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create a formula and settings
	formula := createFormula("TEST_MODEL_DELETE", 1)
	saveFormula(formula, t)

	settings := &logupload.LogUploadSettings{
		ID:                 formula.ID,
		Name:               "Delete Me",
		UploadRepositoryID: "test-repo",
		ApplicationType:    "stb",
		Schedule: logupload.Schedule{
			Type:       "CronExpression",
			Expression: "0 0 * * *",
			TimeZone:   "UTC",
		},
	}
	CreateLogUploadSettings(settings, "stb")

	req := httptest.NewRequest("DELETE", "/xconfAdminService/dcm/logUploadSettings/"+formula.ID, nil)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	// Verify it's actually deleted
	deleted := logupload.GetOneLogUploadSettings(formula.ID)
	assert.Assert(t, deleted == nil)
}

// ========== Tests for CreateLogUploadSettingsHandler - error paths ==========

// TestCreateLogUploadSettingsHandler_InvalidJSON tests create with invalid JSON (error path)
func TestCreateLogUploadSettingsHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidJSON := []byte(`{invalid json`)

	req := httptest.NewRequest("POST", "/xconfAdminService/dcm/logUploadSettings", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestCreateLogUploadSettingsHandler_EmptyBody tests create with empty body (nil condition)
func TestCreateLogUploadSettingsHandler_EmptyBody(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req := httptest.NewRequest("POST", "/xconfAdminService/dcm/logUploadSettings", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// Should return error for missing required fields
	assert.Assert(t, res.StatusCode >= http.StatusBadRequest)
}

// TestCreateLogUploadSettingsHandler_DuplicateID tests create with duplicate ID (error path)
func TestCreateLogUploadSettingsHandler_DuplicateID(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create a formula and settings
	formula := createFormula("TEST_MODEL_DUP", 1)
	saveFormula(formula, t)

	settings := &logupload.LogUploadSettings{
		ID:                 formula.ID,
		Name:               "First Settings",
		UploadRepositoryID: "test-repo",
		ApplicationType:    "stb",
		Schedule: logupload.Schedule{
			Type:       "CronExpression",
			Expression: "0 0 * * *",
			TimeZone:   "UTC",
		},
	}
	CreateLogUploadSettings(settings, "stb")

	// Try to create another with same ID
	body, _ := json.Marshal(settings)
	req := httptest.NewRequest("POST", "/xconfAdminService/dcm/logUploadSettings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Assert(t, res.StatusCode >= http.StatusBadRequest)
}

// TestCreateLogUploadSettingsHandler_Success tests successful creation
func TestCreateLogUploadSettingsHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create a formula first
	formula := createFormula("TEST_MODEL_CREATE", 1)
	saveFormula(formula, t)

	settings := &logupload.LogUploadSettings{
		ID:                 formula.ID,
		Name:               "Test Settings",
		UploadRepositoryID: "test-repo",
		ApplicationType:    "stb",
		Schedule: logupload.Schedule{
			Type:       "CronExpression",
			Expression: "0 0 * * *",
			TimeZone:   "UTC",
		},
	}
	body, _ := json.Marshal(settings)

	req := httptest.NewRequest("POST", "/xconfAdminService/dcm/logUploadSettings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusCreated, res.StatusCode)
}

// ========== Tests for UpdateLogUploadSettingsHandler - error paths ==========

// TestUpdateLogUploadSettingsHandler_InvalidJSON tests update with invalid JSON (error path)
func TestUpdateLogUploadSettingsHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidJSON := []byte(`{invalid json`)

	req := httptest.NewRequest("PUT", "/xconfAdminService/dcm/logUploadSettings", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestUpdateLogUploadSettingsHandler_NonExistent tests update of non-existent settings (error path)
func TestUpdateLogUploadSettingsHandler_NonExistent(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create a formula but don't create settings
	formula := createFormula("TEST_MODEL_NONEXIST", 1)
	saveFormula(formula, t)

	settings := &logupload.LogUploadSettings{
		ID:                 formula.ID,
		Name:               "Nonexistent Settings",
		UploadRepositoryID: "test-repo",
		ApplicationType:    "stb",
	}
	body, _ := json.Marshal(settings)

	req := httptest.NewRequest("PUT", "/xconfAdminService/dcm/logUploadSettings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Assert(t, res.StatusCode >= http.StatusBadRequest)
}

// TestUpdateLogUploadSettingsHandler_Success tests successful update
func TestUpdateLogUploadSettingsHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create a formula and settings
	formula := createFormula("TEST_MODEL_UPDATE", 1)
	saveFormula(formula, t)

	settings := &logupload.LogUploadSettings{
		ID:                 formula.ID,
		Name:               "Original Name",
		UploadRepositoryID: "test-repo",
		ApplicationType:    "stb",
		Schedule: logupload.Schedule{
			Type:       "CronExpression",
			Expression: "0 0 * * *",
			TimeZone:   "UTC",
		},
	}
	CreateLogUploadSettings(settings, "stb")

	// Update it
	settings.Name = "Updated Name"
	body, _ := json.Marshal(settings)

	req := httptest.NewRequest("PUT", "/xconfAdminService/dcm/logUploadSettings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify the update
	updated := logupload.GetOneLogUploadSettings(formula.ID)
	assert.Equal(t, "Updated Name", updated.Name)
}

// ========== Tests for PostLogUploadSettingsFilteredWithParamsHandler - error paths ==========

// TestPostLogUploadSettingsFilteredWithParamsHandler_EmptyBody tests filtered search with empty body (nil condition)
func TestPostLogUploadSettingsFilteredWithParamsHandler_EmptyBody(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req := httptest.NewRequest("POST", "/xconfAdminService/dcm/logUploadSettings/filtered", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var settings []logupload.LogUploadSettings
	json.NewDecoder(res.Body).Decode(&settings)
	assert.Equal(t, 0, len(settings))
}

// TestPostLogUploadSettingsFilteredWithParamsHandler_InvalidJSON tests filtered search with invalid JSON (error path)
func TestPostLogUploadSettingsFilteredWithParamsHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidJSON := []byte(`{invalid}`)

	req := httptest.NewRequest("POST", "/xconfAdminService/dcm/logUploadSettings/filtered", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestPostLogUploadSettingsFilteredWithParamsHandler_WithContext tests filtered search with context
func TestPostLogUploadSettingsFilteredWithParamsHandler_WithContext(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create some settings
	formula := createFormula("TEST_MODEL_FILTER", 1)
	saveFormula(formula, t)

	settings := &logupload.LogUploadSettings{
		ID:                 formula.ID,
		Name:               "Filtered Settings",
		UploadRepositoryID: "test-repo",
		ApplicationType:    "stb",
		Schedule: logupload.Schedule{
			Type:       "CronExpression",
			Expression: "0 0 * * *",
			TimeZone:   "UTC",
		},
	}
	CreateLogUploadSettings(settings, "stb")

	contextMap := map[string]string{}
	body, _ := json.Marshal(contextMap)

	req := httptest.NewRequest("POST", "/xconfAdminService/dcm/logUploadSettings/filtered", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify X-Number-Of-Items header is present
	//numberHeader := res.Header.Get("X-Number-Of-Items")
	//assert.Assert(t, numberHeader != "")
}

// TestPostLogUploadSettingsFilteredWithParamsHandler_InvalidPagination tests filtered search with invalid pagination
func TestPostLogUploadSettingsFilteredWithParamsHandler_InvalidPagination(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	contextMap := map[string]string{
		"pageNumber": "0", // Invalid page number
		"pageSize":   "10",
	}
	body, _ := json.Marshal(contextMap)

	req := httptest.NewRequest("POST", "/xconfAdminService/dcm/logUploadSettings/filtered", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
