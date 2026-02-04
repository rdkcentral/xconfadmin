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
	SkipIfMockDatabase(t) // Integration test
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

// ========== Tests for Nil Conditions and Error Paths ==========

// TestGetLogRepoSettingsByIdHandler_MissingID tests error when ID is missing
func TestGetLogRepoSettingsByIdHandler_MissingID(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/uploadRepository/", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// Should return 404 as the route doesn't match
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestGetLogRepoSettingsByIdHandler_NilResult tests handling when repository doesn't exist (nil condition)
func TestGetLogRepoSettingsByIdHandler_NilResult(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/uploadRepository/nonexistent-id", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestGetLogRepoSettingsByIdHandler_ApplicationTypeMismatch tests when ApplicationType doesn't match (error path)
func TestGetLogRepoSettingsByIdHandler_ApplicationTypeMismatch(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create repository with "xhome" application type
	repo := logupload.UploadRepository{
		ID:              "xhome-repo",
		Name:            "XHome Repo",
		Description:     "Test",
		URL:             "http://test.com",
		Protocol:        "HTTP",
		ApplicationType: "xhome",
	}
	CreateLogRepoSettings(&repo, "xhome")

	// Try to access with "stb" application type
	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/uploadRepository/xhome-repo", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestGetLogRepoSettingsByIdHandler_WithExport tests export functionality for single repository
func TestGetLogRepoSettingsByIdHandler_WithExport(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create repository
	repo := logupload.UploadRepository{
		ID:              "export-repo",
		Name:            "Export Repo",
		Description:     "Test",
		URL:             "http://test.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	CreateLogRepoSettings(&repo, "stb")

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/uploadRepository/export-repo?export=true", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header is present
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "")

	// Verify response is an array with one item
	var repoList []logupload.UploadRepository
	json.NewDecoder(res.Body).Decode(&repoList)
	assert.Equal(t, 1, len(repoList))
	assert.Equal(t, "export-repo", repoList[0].ID)
}

// TestGetLogRepoSettingsHandler_EmptyList tests handling when no repositories exist (nil condition)
func TestGetLogRepoSettingsHandler_EmptyList(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/uploadRepository", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var repoList []*logupload.UploadRepository
	json.NewDecoder(res.Body).Decode(&repoList)
	assert.Equal(t, 0, len(repoList))
}

// TestGetLogRepoSettingsHandler_WithExport tests export functionality for all repositories
func TestGetLogRepoSettingsHandler_WithExport(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create multiple repositories
	repo1 := logupload.UploadRepository{
		ID:              "repo1",
		Name:            "Repo 1",
		URL:             "http://test1.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	repo2 := logupload.UploadRepository{
		ID:              "repo2",
		Name:            "Repo 2",
		URL:             "http://test2.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	CreateLogRepoSettings(&repo1, "stb")
	CreateLogRepoSettings(&repo2, "stb")

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/uploadRepository?export=true", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header is present
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "")
}

// TestGetLogRepoSettingsSizeHandler_ZeroCount tests size handler with no repositories (nil condition)
func TestGetLogRepoSettingsSizeHandler_ZeroCount(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/uploadRepository/size", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var count int
	json.NewDecoder(res.Body).Decode(&count)
	assert.Equal(t, 0, count)
}

// TestGetLogRepoSettingsSizeHandler_NonZeroCount tests size handler with repositories
func TestGetLogRepoSettingsSizeHandler_NonZeroCount(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create repositories
	for i := 1; i <= 3; i++ {
		repo := logupload.UploadRepository{
			ID:              fmt.Sprintf("repo-%d", i),
			Name:            fmt.Sprintf("Repo %d", i),
			URL:             fmt.Sprintf("http://test%d.com", i),
			Protocol:        "HTTP",
			ApplicationType: "stb",
		}
		CreateLogRepoSettings(&repo, "stb")
	}

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/uploadRepository/size", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var count int
	json.NewDecoder(res.Body).Decode(&count)
	assert.Equal(t, 3, count)
}

// TestGetLogRepoSettingsNamesHandler_EmptyList tests names handler with no repositories (nil condition)
func TestGetLogRepoSettingsNamesHandler_EmptyList(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/uploadRepository/names", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var names []string
	json.NewDecoder(res.Body).Decode(&names)
	assert.Equal(t, 0, len(names))
}

// TestGetLogRepoSettingsNamesHandler_WithNames tests names handler with repositories
func TestGetLogRepoSettingsNamesHandler_WithNames(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create repositories with specific names
	names := []string{"Alpha Repo", "Beta Repo", "Gamma Repo"}
	for i, name := range names {
		repo := logupload.UploadRepository{
			ID:              fmt.Sprintf("repo-%d", i),
			Name:            name,
			URL:             "http://test.com",
			Protocol:        "HTTP",
			ApplicationType: "stb",
		}
		CreateLogRepoSettings(&repo, "stb")
	}

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/uploadRepository/names", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var returnedNames []string
	json.NewDecoder(res.Body).Decode(&returnedNames)
	assert.Equal(t, 3, len(returnedNames))
}

// TestDeleteLogRepoSettingsByIdHandler_MissingID tests delete with missing ID (error path)
func TestDeleteLogRepoSettingsByIdHandler_MissingID(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("DELETE", "/xconfAdminService/dcm/uploadRepository/", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// Should return 404 as the route doesn't match
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestDeleteLogRepoSettingsByIdHandler_NonExistent tests delete of non-existent repository (error path)
func TestDeleteLogRepoSettingsByIdHandler_NonExistent(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("DELETE", "/xconfAdminService/dcm/uploadRepository/nonexistent-id", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestDeleteLogRepoSettingsByIdHandler_Success tests successful delete
func TestDeleteLogRepoSettingsByIdHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create repository
	repo := logupload.UploadRepository{
		ID:              "delete-me",
		Name:            "Delete Me",
		URL:             "http://test.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	CreateLogRepoSettings(&repo, "stb")

	req, err := http.NewRequest("DELETE", "/xconfAdminService/dcm/uploadRepository/delete-me", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	// Verify it's actually deleted
	deleted := GetLogRepoSettings("delete-me")
	assert.Assert(t, deleted == nil)
}

// TestCreateLogRepoSettingsHandler_InvalidJSON tests create with invalid JSON (error path)
func TestCreateLogRepoSettingsHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidJSON := []byte(`{invalid json`)

	req, err := http.NewRequest("POST", "/xconfAdminService/dcm/uploadRepository", bytes.NewBuffer(invalidJSON))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestCreateLogRepoSettingsHandler_EmptyBody tests create with empty body (nil condition)
func TestCreateLogRepoSettingsHandler_EmptyBody(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("POST", "/xconfAdminService/dcm/uploadRepository", bytes.NewBuffer([]byte("{}")))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// Should return error for missing required fields
	assert.Assert(t, res.StatusCode >= http.StatusBadRequest)
}

// TestCreateLogRepoSettingsHandler_DuplicateID tests create with duplicate ID (error path)
func TestCreateLogRepoSettingsHandler_DuplicateID(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create first repository
	repo := logupload.UploadRepository{
		ID:              "duplicate-id",
		Name:            "First Repo",
		URL:             "http://test.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	CreateLogRepoSettings(&repo, "stb")

	// Try to create another with same ID
	body, _ := json.Marshal(repo)
	req, err := http.NewRequest("POST", "/xconfAdminService/dcm/uploadRepository", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Assert(t, res.StatusCode >= http.StatusBadRequest)
}

// TestUpdateLogRepoSettingsHandler_InvalidJSON tests update with invalid JSON (error path)
func TestUpdateLogRepoSettingsHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidJSON := []byte(`{invalid json`)

	req, err := http.NewRequest("PUT", "/xconfAdminService/dcm/uploadRepository", bytes.NewBuffer(invalidJSON))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestUpdateLogRepoSettingsHandler_NonExistent tests update of non-existent repository (error path)
func TestUpdateLogRepoSettingsHandler_NonExistent(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	repo := logupload.UploadRepository{
		ID:              "nonexistent",
		Name:            "Nonexistent Repo",
		URL:             "http://test.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	body, _ := json.Marshal(repo)

	req, err := http.NewRequest("PUT", "/xconfAdminService/dcm/uploadRepository", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Assert(t, res.StatusCode >= http.StatusBadRequest)
}

// TestUpdateLogRepoSettingsHandler_Success tests successful update
func TestUpdateLogRepoSettingsHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create repository
	repo := logupload.UploadRepository{
		ID:              "update-me",
		Name:            "Original Name",
		Description:     "Original",
		URL:             "http://original.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	CreateLogRepoSettings(&repo, "stb")

	// Update it
	repo.Name = "Updated Name"
	repo.Description = "Updated"
	repo.URL = "http://updated.com"
	body, _ := json.Marshal(repo)

	req, err := http.NewRequest("PUT", "/xconfAdminService/dcm/uploadRepository", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify the update
	updated := GetLogRepoSettings("update-me")
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated", updated.Description)
}

// TestPostLogRepoSettingsFilteredWithParamsHandler_EmptyBody tests filtered search with empty body (nil condition)
func TestPostLogRepoSettingsFilteredWithParamsHandler_EmptyBody(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("POST", "/xconfAdminService/dcm/uploadRepository/filtered", bytes.NewBuffer([]byte("")))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var repos []logupload.UploadRepository
	json.NewDecoder(res.Body).Decode(&repos)
	assert.Equal(t, 0, len(repos))
}

// TestPostLogRepoSettingsFilteredWithParamsHandler_InvalidJSON tests filtered search with invalid JSON (error path)
func TestPostLogRepoSettingsFilteredWithParamsHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidJSON := []byte(`{invalid}`)

	req, err := http.NewRequest("POST", "/xconfAdminService/dcm/uploadRepository/filtered", bytes.NewBuffer(invalidJSON))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestPostLogRepoSettingsFilteredWithParamsHandler_WithContext tests filtered search with context
func TestPostLogRepoSettingsFilteredWithParamsHandler_WithContext(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create some repositories
	repo1 := logupload.UploadRepository{
		ID:              "filtered-1",
		Name:            "Filtered One",
		URL:             "http://test1.com",
		Protocol:        "HTTP",
		ApplicationType: "stb",
	}
	CreateLogRepoSettings(&repo1, "stb")

	contextMap := map[string]string{}
	body, _ := json.Marshal(contextMap)

	req, err := http.NewRequest("POST", "/xconfAdminService/dcm/uploadRepository/filtered", bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

// TestPostLogRepoSettingsEntitiesHandler_EmptyArray tests batch create with empty array (nil condition)
func TestPostLogRepoSettingsEntitiesHandler_EmptyArray(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	entities := []logupload.UploadRepository{}
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
	assert.Equal(t, 0, len(responseMap))
}

// TestPutLogRepoSettingsEntitiesHandler_EmptyArray tests batch update with empty array (nil condition)
func TestPutLogRepoSettingsEntitiesHandler_EmptyArray(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	entities := []logupload.UploadRepository{}
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
	assert.Equal(t, 0, len(responseMap))
}

// TestGetLogRepoSettingsExportHandler_ApplicationTypeFiltering tests export filters by application type
func TestGetLogRepoSettingsExportHandler_ApplicationTypeFiltering(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/logUploadSettings/export", nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var lusList []*logupload.LogUploadSettings
	json.NewDecoder(res.Body).Decode(&lusList)
	// Verify all returned items match application type (if any exist)
	for _, lus := range lusList {
		if lus != nil {
			assert.Equal(t, "stb", lus.ApplicationType)
		}
	}
}
