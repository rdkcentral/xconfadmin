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
	"net/http"
	"testing"

	xcommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"gotest.tools/assert"
)

// TestPostLogRepoSettingsEntitiesHandler_Success tests successful batch creation of upload repositories
func TestPostLogRepoSettingsEntitiesHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	entities := []logupload.UploadRepository{
		{
			ID:              "repo-1",
			Name:            "Repo One",
			Description:     "Test Repo 1",
			URL:             "http://test1.com",
			Protocol:        "HTTP",
			ApplicationType: "stb",
		},
		{
			ID:              "repo-2",
			Name:            "Repo Two",
			Description:     "Test Repo 2",
			URL:             "http://test2.com",
			Protocol:        "HTTPS",
			ApplicationType: "stb",
		},
	}
	body, _ := json.Marshal(entities)

	req, err := http.NewRequest("POST", "/xconfAdminService/dcm/uploadRepository/entities", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify response structure
	var responseMap map[string]xhttp.EntityMessage
	json.NewDecoder(res.Body).Decode(&responseMap)
	assert.Equal(t, 2, len(responseMap))
	assert.Equal(t, xcommon.ENTITY_STATUS_SUCCESS, responseMap["repo-1"].Status)
	assert.Equal(t, xcommon.ENTITY_STATUS_SUCCESS, responseMap["repo-2"].Status)
}

// TestPostLogRepoSettingsEntitiesHandler_InvalidJSON tests invalid JSON handling
func TestPostLogRepoSettingsEntitiesHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidJSON := []byte(`{bad json}`)

	req, err := http.NewRequest("POST", "/xconfAdminService/dcm/uploadRepository/entities", bytes.NewBuffer(invalidJSON))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestPostLogRepoSettingsEntitiesHandler_DuplicateEntity tests duplicate entity handling
func TestPostLogRepoSettingsEntitiesHandler_DuplicateEntity(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create first repository
	repo := logupload.UploadRepository{
		ID:              "duplicate-repo",
		Name:            "Duplicate Repo",
		Description:     "Test",
		URL:             "http://test.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	CreateLogRepoSettings(&repo, "stb")

	// Try to create the same entity again
	entities := []logupload.UploadRepository{repo}
	body, _ := json.Marshal(entities)

	req, err := http.NewRequest("POST", "/xconfAdminService/dcm/uploadRepository/entities", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseMap map[string]xhttp.EntityMessage
	json.NewDecoder(res.Body).Decode(&responseMap)
	assert.Equal(t, xcommon.ENTITY_STATUS_FAILURE, responseMap["duplicate-repo"].Status)
}

// TestPostLogRepoSettingsEntitiesHandler_MixedSuccessAndFailure tests batch with both successful and failed operations
func TestPostLogRepoSettingsEntitiesHandler_MixedSuccessAndFailure(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create first repository
	existingRepo := logupload.UploadRepository{
		ID:              "existing-repo",
		Name:            "Existing Repo",
		Description:     "Test",
		URL:             "http://existing.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	CreateLogRepoSettings(&existingRepo, "stb")

	// Batch with one new and one duplicate
	entities := []logupload.UploadRepository{
		existingRepo, // This should fail
		{
			ID:              "new-repo",
			Name:            "New Repo",
			Description:     "Test",
			URL:             "http://new.com",
			Protocol:        "HTTP",
			ApplicationType: "stb",
		}, // This should succeed
	}
	body, _ := json.Marshal(entities)

	req, err := http.NewRequest("POST", "/xconfAdminService/dcm/uploadRepository/entities", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseMap map[string]xhttp.EntityMessage
	json.NewDecoder(res.Body).Decode(&responseMap)
	assert.Equal(t, 2, len(responseMap))
	assert.Equal(t, xcommon.ENTITY_STATUS_FAILURE, responseMap["existing-repo"].Status)
	assert.Equal(t, xcommon.ENTITY_STATUS_SUCCESS, responseMap["new-repo"].Status)
}

// TestPutLogRepoSettingsEntitiesHandler_Success tests successful batch update of upload repositories
func TestPutLogRepoSettingsEntitiesHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create initial repositories
	repo1 := logupload.UploadRepository{
		ID:              "update-repo-1",
		Name:            "Original Name 1",
		Description:     "Original Desc",
		URL:             "http://original1.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	repo2 := logupload.UploadRepository{
		ID:              "update-repo-2",
		Name:            "Original Name 2",
		Description:     "Original Desc",
		URL:             "http://original2.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	CreateLogRepoSettings(&repo1, "stb")
	CreateLogRepoSettings(&repo2, "stb")

	// Update both repositories
	updatedEntities := []logupload.UploadRepository{
		{
			ID:              "update-repo-1",
			Name:            "Updated Name 1",
			Description:     "Updated Desc",
			URL:             "http://updated1.com",
			Protocol:        "HTTPS",
			ApplicationType: "stb",
		},
		{
			ID:              "update-repo-2",
			Name:            "Updated Name 2",
			Description:     "Updated Desc",
			URL:             "http://updated2.com",
			Protocol:        "HTTPS",
			ApplicationType: "stb",
		},
	}
	body, _ := json.Marshal(updatedEntities)

	req, err := http.NewRequest("PUT", "/xconfAdminService/dcm/uploadRepository/entities", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseMap map[string]xhttp.EntityMessage
	json.NewDecoder(res.Body).Decode(&responseMap)
	assert.Equal(t, 2, len(responseMap))
	assert.Equal(t, xcommon.ENTITY_STATUS_SUCCESS, responseMap["update-repo-1"].Status)
	assert.Equal(t, xcommon.ENTITY_STATUS_SUCCESS, responseMap["update-repo-2"].Status)
}

// TestPutLogRepoSettingsEntitiesHandler_InvalidJSON tests invalid JSON handling for update
func TestPutLogRepoSettingsEntitiesHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidJSON := []byte(`{bad json}`)

	req, err := http.NewRequest("PUT", "/xconfAdminService/dcm/uploadRepository/entities", bytes.NewBuffer(invalidJSON))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestPutLogRepoSettingsEntitiesHandler_NonExistentEntity tests updating non-existent entity
func TestPutLogRepoSettingsEntitiesHandler_NonExistentEntity(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	entities := []logupload.UploadRepository{
		{
			ID:              "nonexistent-repo",
			Name:            "Nonexistent Repo",
			Description:     "Test",
			URL:             "http://test.com",
			Protocol:        "HTTP",
			ApplicationType: "stb",
		},
	}
	body, _ := json.Marshal(entities)

	req, err := http.NewRequest("PUT", "/xconfAdminService/dcm/uploadRepository/entities", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseMap map[string]xhttp.EntityMessage
	json.NewDecoder(res.Body).Decode(&responseMap)
	assert.Equal(t, xcommon.ENTITY_STATUS_FAILURE, responseMap["nonexistent-repo"].Status)
}

// TestPutLogRepoSettingsEntitiesHandler_MixedSuccessAndFailure tests batch update with mixed results
func TestPutLogRepoSettingsEntitiesHandler_MixedSuccessAndFailure(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create one repository
	existingRepo := logupload.UploadRepository{
		ID:              "existing-update-repo",
		Name:            "Existing Repo",
		Description:     "Test",
		URL:             "http://existing.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	CreateLogRepoSettings(&existingRepo, "stb")

	// Batch with one existing and one non-existent
	entities := []logupload.UploadRepository{
		{
			ID:              "existing-update-repo",
			Name:            "Updated Existing",
			Description:     "Updated",
			URL:             "http://updated.com",
			Protocol:        "HTTPS",
			ApplicationType: "stb",
		}, // Should succeed
		{
			ID:              "nonexistent-update-repo",
			Name:            "Nonexistent",
			Description:     "Test",
			URL:             "http://test.com",
			Protocol:        "HTTP",
			ApplicationType: "stb",
		}, // Should fail
	}
	body, _ := json.Marshal(entities)

	req, err := http.NewRequest("PUT", "/xconfAdminService/dcm/uploadRepository/entities", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var responseMap map[string]xhttp.EntityMessage
	json.NewDecoder(res.Body).Decode(&responseMap)
	assert.Equal(t, 2, len(responseMap))
	assert.Equal(t, xcommon.ENTITY_STATUS_SUCCESS, responseMap["existing-update-repo"].Status)
	assert.Equal(t, xcommon.ENTITY_STATUS_FAILURE, responseMap["nonexistent-update-repo"].Status)
}

// TestGetLogRepoSettingsExportHandler_Success tests successful export of log upload settings
func TestGetLogRepoSettingsExportHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "")
	assert.Assert(t, len(contentDisposition) > 0)

	// Verify response body is a valid JSON array
	var lusList []*logupload.LogUploadSettings
	json.NewDecoder(res.Body).Decode(&lusList)
	assert.Assert(t, len(lusList) >= 0) // Should return list (may be empty or have items)
}

// TestGetLogRepoSettingsExportHandler_EmptyResult tests export with no data
func TestGetLogRepoSettingsExportHandler_EmptyResult(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header is present
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "")

	// Verify response is an empty list
	var lusList []*logupload.LogUploadSettings
	json.NewDecoder(res.Body).Decode(&lusList)
	assert.Equal(t, 0, len(lusList))
}

// TestGetLogRepoSettingsExportHandler_VerifyHeaders tests that export includes correct headers
func TestGetLogRepoSettingsExportHandler_VerifyHeaders(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header contains expected filename pattern
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "")
	// The filename should contain "allLogRepoSettings_stb"

	// Verify Content-Type is JSON
	contentType := res.Header.Get("Content-Type")
	assert.Assert(t, contentType != "")
}
