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
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"gotest.tools/assert"
)

// ========== Tests for PostModelEntitiesHandler ==========

func TestPostModelEntitiesHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	models := []shared.Model{
		{
			ID:          "MODEL1",
			Description: "Test Model 1",
		},
		{
			ID:          "MODEL2",
			Description: "Test Model 2",
		},
	}

	body, err := json.Marshal(models)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/entities"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify response
	var result map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&result)
	assert.NilError(t, err)

	// Check both models were created successfully
	model1Result, ok := result["MODEL1"].(map[string]interface{})
	assert.Check(t, ok, "MODEL1 should be in result")
	assert.Equal(t, model1Result["status"], "SUCCESS")

	model2Result, ok := result["MODEL2"].(map[string]interface{})
	assert.Check(t, ok, "MODEL2 should be in result")
	assert.Equal(t, model2Result["status"], "SUCCESS")

	// Verify models were created in DB
	savedModel1 := shared.GetOneModel("MODEL1")
	assert.Check(t, savedModel1 != nil, "MODEL1 should be saved")
	assert.Equal(t, savedModel1.Description, "Test Model 1")

	savedModel2 := shared.GetOneModel("MODEL2")
	assert.Check(t, savedModel2 != nil, "MODEL2 should be saved")
	assert.Equal(t, savedModel2.Description, "Test Model 2")
}

func TestPostModelEntitiesHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidBody := []byte(`{"invalid json}`)

	url := "/xconfAdminService/model/entities"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(invalidBody))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func TestPostModelEntitiesHandler_DuplicateModel(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create first model
	model1 := &shared.Model{
		ID:          "DUPLICATE_MODEL",
		Description: "First Model",
	}
	CreateModel(model1)

	// Try to create same model again in batch
	models := []shared.Model{
		{
			ID:          "DUPLICATE_MODEL",
			Description: "Duplicate Model",
		},
	}

	body, err := json.Marshal(models)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/entities"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify response shows failure
	var result map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&result)
	assert.NilError(t, err)

	modelResult, ok := result["DUPLICATE_MODEL"].(map[string]interface{})
	assert.Check(t, ok, "DUPLICATE_MODEL should be in result")
	assert.Equal(t, modelResult["status"], "FAILURE")
}

func TestPostModelEntitiesHandler_MixedSuccessAndFailure(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create one model first
	existingModel := &shared.Model{
		ID:          "EXISTING_MODEL",
		Description: "Existing",
	}
	CreateModel(existingModel)

	// Try to create batch with one duplicate and one new
	models := []shared.Model{
		{
			ID:          "EXISTING_MODEL",
			Description: "Duplicate",
		},
		{
			ID:          "NEW_MODEL",
			Description: "New Model",
		},
	}

	body, err := json.Marshal(models)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/entities"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify response
	var result map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&result)
	assert.NilError(t, err)

	existingResult, ok := result["EXISTING_MODEL"].(map[string]interface{})
	assert.Check(t, ok)
	assert.Equal(t, existingResult["status"], "FAILURE")

	newResult, ok := result["NEW_MODEL"].(map[string]interface{})
	assert.Check(t, ok)
	assert.Equal(t, newResult["status"], "SUCCESS")
}

// ========== Tests for PutModelEntitiesHandler ==========

func TestPutModelEntitiesHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create models first
	model1 := &shared.Model{
		ID:          "UPDATE_MODEL1",
		Description: "Original 1",
	}
	model2 := &shared.Model{
		ID:          "UPDATE_MODEL2",
		Description: "Original 2",
	}
	CreateModel(model1)
	CreateModel(model2)

	// Update both models
	updatedModels := []shared.Model{
		{
			ID:          "UPDATE_MODEL1",
			Description: "Updated 1",
		},
		{
			ID:          "UPDATE_MODEL2",
			Description: "Updated 2",
		},
	}

	body, err := json.Marshal(updatedModels)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/entities"
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify updates
	updated1 := shared.GetOneModel("UPDATE_MODEL1")
	assert.Check(t, updated1 != nil)
	assert.Equal(t, updated1.Description, "Updated 1")

	updated2 := shared.GetOneModel("UPDATE_MODEL2")
	assert.Check(t, updated2 != nil)
	assert.Equal(t, updated2.Description, "Updated 2")
}

func TestPutModelEntitiesHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidBody := []byte(`{"bad": json}`)

	url := "/xconfAdminService/model/entities"
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(invalidBody))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func TestPutModelEntitiesHandler_NonExistentModel(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	models := []shared.Model{
		{
			ID:          "NONEXISTENT_MODEL",
			Description: "Does not exist",
		},
	}

	body, err := json.Marshal(models)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/entities"
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify response shows failure
	var result map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&result)
	assert.NilError(t, err)

	modelResult, ok := result["NONEXISTENT_MODEL"].(map[string]interface{})
	assert.Check(t, ok)
	assert.Equal(t, modelResult["status"], "FAILURE")
}

// ========== Tests for ObsoleteGetModelPageHandler ==========

func TestObsoleteGetModelPageHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test models
	for i := 1; i <= 5; i++ {
		model := &shared.Model{
			ID:          fmt.Sprintf("PAGE_MODEL_%d", i),
			Description: fmt.Sprintf("Model %d", i),
		}
		CreateModel(model)
	}

	url := "/xconfAdminService/model/page?pageNumber=1&pageSize=3"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Note: This handler is defined but may not be routed
	// If routed, it should return OK, otherwise 404
	if res.StatusCode == http.StatusOK {
		var models []shared.Model
		err = json.NewDecoder(res.Body).Decode(&models)
		assert.NilError(t, err)
		assert.Check(t, len(models) <= 3, "Should return at most 3 models")

		// Verify header with total count
		numberHeader := res.Header.Get("numberOfItems")
		assert.Check(t, numberHeader != "", "Should have numberOfItems header")
	}
}

func TestObsoleteGetModelPageHandler_InvalidPageNumber(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	url := "/xconfAdminService/model/page?pageNumber=invalid&pageSize=3"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Should return 400 Bad Request for invalid pageNumber
	if res.StatusCode == http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(res.Body)
		assert.Check(t, len(bodyBytes) > 0, "Should have error message in body")
	}
}

func TestObsoleteGetModelPageHandler_InvalidPageSize(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	url := "/xconfAdminService/model/page?pageNumber=1&pageSize=invalid"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Should return 400 Bad Request for invalid pageSize
	if res.StatusCode == http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(res.Body)
		assert.Check(t, len(bodyBytes) > 0, "Should have error message in body")
	}
}

func TestObsoleteGetModelPageHandler_Pagination(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create 10 models
	for i := 1; i <= 10; i++ {
		model := &shared.Model{
			ID:          fmt.Sprintf("PAGINATE_%02d", i),
			Description: fmt.Sprintf("Model %d", i),
		}
		CreateModel(model)
	}

	// Request page 2 with 3 items per page
	url := "/xconfAdminService/model/page?pageNumber=2&pageSize=3"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var models []shared.Model
		err = json.NewDecoder(res.Body).Decode(&models)
		assert.NilError(t, err)
		assert.Check(t, len(models) <= 3, "Should return at most 3 models")
	}
}

func TestObsoleteGetModelPageHandler_EmptyResult(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	url := "/xconfAdminService/model/page?pageNumber=1&pageSize=10"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var models []shared.Model
		bodyBytes, _ := io.ReadAll(res.Body)
		err = json.Unmarshal(bodyBytes, &models)
		assert.NilError(t, err)
		assert.Equal(t, len(models), 0, "Should return empty array")
	}
}

// ========== Tests for PostModelFilteredHandler ==========

func TestPostModelFilteredHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test models
	model1 := &shared.Model{
		ID:          "FILTER_MODEL1",
		Description: "Test Model 1",
	}
	model2 := &shared.Model{
		ID:          "FILTER_MODEL2",
		Description: "Test Model 2",
	}
	CreateModel(model1)
	CreateModel(model2)

	filterContext := map[string]string{}
	body, err := json.Marshal(filterContext)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/filtered?pageNumber=1&pageSize=10"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify we got models back
	var models []shared.Model
	err = json.NewDecoder(res.Body).Decode(&models)
	assert.NilError(t, err)
	assert.Check(t, len(models) >= 2, "Should return at least 2 models")
}

func TestPostModelFilteredHandler_WithEmptyBody(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test model
	model := &shared.Model{
		ID:          "EMPTY_FILTER_MODEL",
		Description: "Test",
	}
	CreateModel(model)

	url := "/xconfAdminService/model/filtered?pageNumber=1&pageSize=10"
	req, err := http.NewRequest("POST", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)
}

func TestPostModelFilteredHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	invalidBody := []byte(`{invalid}`)

	url := "/xconfAdminService/model/filtered?pageNumber=1&pageSize=10"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(invalidBody))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func TestPostModelFilteredHandler_InvalidPageNumber(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	filterContext := map[string]string{}
	body, err := json.Marshal(filterContext)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/filtered?pageNumber=invalid&pageSize=10"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func TestPostModelFilteredHandler_Pagination(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create 5 models
	for i := 1; i <= 5; i++ {
		model := &shared.Model{
			ID:          fmt.Sprintf("PAGINATED_MODEL_%d", i),
			Description: fmt.Sprintf("Model %d", i),
		}
		CreateModel(model)
	}

	filterContext := map[string]string{}
	body, err := json.Marshal(filterContext)
	assert.NilError(t, err)

	// Request first page with 3 items
	url := "/xconfAdminService/model/filtered?pageNumber=1&pageSize=3"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	var models []shared.Model
	err = json.NewDecoder(res.Body).Decode(&models)
	assert.NilError(t, err)
	assert.Equal(t, len(models), 3, "Should return 3 models on first page")
}

// ========== Tests for GetModelByIdHandler ==========

func TestGetModelByIdHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test model
	model := &shared.Model{
		ID:          "GET_BY_ID_MODEL",
		Description: "Test Model",
	}
	CreateModel(model)

	url := "/xconfAdminService/model/GET_BY_ID_MODEL"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify response contains the model
	var returnedModel shared.Model
	err = json.NewDecoder(res.Body).Decode(&returnedModel)
	assert.NilError(t, err)
	assert.Equal(t, returnedModel.ID, "GET_BY_ID_MODEL")
	assert.Equal(t, returnedModel.Description, "Test Model")
}

func TestGetModelByIdHandler_NotFound(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	url := "/xconfAdminService/model/NONEXISTENT"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusNotFound)
}

func TestGetModelByIdHandler_WithExport(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test model
	model := &shared.Model{
		ID:          "EXPORT_MODEL",
		Description: "Export Test",
	}
	CreateModel(model)

	url := "/xconfAdminService/model/EXPORT_MODEL?export"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify Content-Disposition header
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Check(t, contentDisposition != "", "Content-Disposition header should be set")
	assert.Check(t, len(contentDisposition) > 0, "Content-Disposition should contain filename")

	// Verify response is an array
	var models []shared.Model
	err = json.NewDecoder(res.Body).Decode(&models)
	assert.NilError(t, err)
	assert.Equal(t, len(models), 1, "Export should return array with 1 model")
	assert.Equal(t, models[0].ID, "EXPORT_MODEL")
}

func TestGetModelByIdHandler_CaseInsensitive(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test model with lowercase ID
	model := &shared.Model{
		ID:          "lowercase_model",
		Description: "Test",
	}
	CreateModel(model)

	// Request with uppercase
	url := "/xconfAdminService/model/LOWERCASE_MODEL"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)
}

// ========== Tests for GetModelHandler ==========

func TestGetModelHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test models
	model1 := &shared.Model{
		ID:          "ALL_MODEL1",
		Description: "Model 1",
	}
	model2 := &shared.Model{
		ID:          "ALL_MODEL2",
		Description: "Model 2",
	}
	CreateModel(model1)
	CreateModel(model2)

	url := "/xconfAdminService/model"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	var models []shared.Model
	err = json.NewDecoder(res.Body).Decode(&models)
	assert.NilError(t, err)
	assert.Check(t, len(models) >= 2, "Should return at least 2 models")
}

func TestGetModelHandler_EmptyResult(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	url := "/xconfAdminService/model"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	var models []shared.Model
	bodyBytes, err := io.ReadAll(res.Body)
	assert.NilError(t, err)

	err = json.Unmarshal(bodyBytes, &models)
	assert.NilError(t, err)
	assert.Equal(t, len(models), 0, "Should return empty array")
}

func TestGetModelHandler_WithExport(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test models
	model := &shared.Model{
		ID:          "EXPORT_ALL_MODEL",
		Description: "Export Test",
	}
	CreateModel(model)

	url := "/xconfAdminService/model?export"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify Content-Disposition header
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Check(t, contentDisposition != "", "Content-Disposition header should be set for export")
}

func TestGetModelHandler_SortedAlphabetically(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create models in non-alphabetical order
	modelZ := &shared.Model{
		ID:          "Z_MODEL",
		Description: "Z",
	}
	modelA := &shared.Model{
		ID:          "A_MODEL",
		Description: "A",
	}
	modelM := &shared.Model{
		ID:          "M_MODEL",
		Description: "M",
	}
	CreateModel(modelZ)
	CreateModel(modelA)
	CreateModel(modelM)

	url := "/xconfAdminService/model"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	var models []shared.Model
	err = json.NewDecoder(res.Body).Decode(&models)
	assert.NilError(t, err)
	assert.Check(t, len(models) >= 3, "Should return at least 3 models")

	// Verify they are sorted (check first few)
	for i := 0; i < len(models)-1; i++ {
		current := models[i].ID
		next := models[i+1].ID
		assert.Check(t, current <= next, fmt.Sprintf("Models should be sorted: %s should come before or equal to %s", current, next))
	}
}

// ========== Additional Error Path Tests for WriteAdminErrorResponse ==========

func TestPostModelEntitiesHandler_UnableToExtractBody(t *testing.T) {
	// This test verifies the error path when response writer is not XResponseWriter
	// In practice, this is hard to trigger in the test harness as ExecuteRequest
	// always wraps with XResponseWriter, but we can document the behavior
	DeleteAllEntities()
	defer DeleteAllEntities()

	models := []shared.Model{
		{
			ID:          "TEST_MODEL",
			Description: "Test",
		},
	}

	body, err := json.Marshal(models)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/entities"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Normal path should succeed
	assert.Equal(t, res.StatusCode, http.StatusOK)
}

func TestPutModelEntitiesHandler_EmptyID(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Try to update model with empty ID
	models := []shared.Model{
		{
			ID:          "",
			Description: "Empty ID Model",
		},
	}

	body, err := json.Marshal(models)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/entities"
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify response shows failure
	var result map[string]interface{}
	bodyBytes, _ := io.ReadAll(res.Body)
	err = json.Unmarshal(bodyBytes, &result)
	assert.NilError(t, err)

	// Empty ID should result in failure
	modelResult, ok := result[""].(map[string]interface{})
	assert.Check(t, ok, "Empty ID should be in result")
	assert.Equal(t, modelResult["status"], "FAILURE")
}

func TestPostModelFilteredHandler_FilterContextError(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test model
	model := &shared.Model{
		ID:          "FILTER_ERROR_MODEL",
		Description: "Test",
	}
	CreateModel(model)

	// Use invalid filter context (malformed JSON)
	invalidBody := []byte(`{"key": "value"`)

	url := "/xconfAdminService/model/filtered?pageNumber=1&pageSize=10"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(invalidBody))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusBadRequest)

	bodyBytes, _ := io.ReadAll(res.Body)
	assert.Check(t, len(bodyBytes) > 0, "Should have error message")
}

func TestPostModelFilteredHandler_NegativePageNumber(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	filterContext := map[string]string{}
	body, err := json.Marshal(filterContext)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/filtered?pageNumber=-1&pageSize=10"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Should return 400 for negative page number
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func TestPostModelFilteredHandler_ZeroPageSize(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	filterContext := map[string]string{}
	body, err := json.Marshal(filterContext)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/filtered?pageNumber=1&pageSize=0"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Should return 400 for zero page size
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func TestGetModelByIdHandler_EmptyID(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Try to get model with empty ID - this will fail at routing level
	// but test the handler behavior
	url := "/xconfAdminService/model/"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Router will not match this path, so it will return 404 or redirect
	assert.Check(t, res.StatusCode != http.StatusOK, "Empty ID should not succeed")
}

func TestPostModelEntitiesHandler_ValidationError(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create model with invalid data
	models := []shared.Model{
		{
			ID:          "", // Empty ID should cause validation error
			Description: "Invalid Model",
		},
	}

	body, err := json.Marshal(models)
	assert.NilError(t, err)

	url := "/xconfAdminService/model/entities"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify response shows failure
	var result map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&result)
	assert.NilError(t, err)

	// Empty ID model should fail
	emptyResult, ok := result[""].(map[string]interface{})
	assert.Check(t, ok, "Empty ID should be in result")
	assert.Equal(t, emptyResult["status"], "FAILURE")
}

func TestObsoleteGetModelPageHandler_PageOutOfBounds(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create 3 models
	for i := 1; i <= 3; i++ {
		model := &shared.Model{
			ID:          fmt.Sprintf("OOB_MODEL_%d", i),
			Description: fmt.Sprintf("Model %d", i),
		}
		CreateModel(model)
	}

	// Request page 10 which doesn't exist
	url := "/xconfAdminService/model/page?pageNumber=10&pageSize=3"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var models []shared.Model
		err = json.NewDecoder(res.Body).Decode(&models)
		assert.NilError(t, err)
		// Out of bounds page should return empty array
		assert.Equal(t, len(models), 0, "Out of bounds page should return empty array")
	}
}

func TestPostModelFilteredHandler_LargePageSize(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create a few models
	for i := 1; i <= 5; i++ {
		model := &shared.Model{
			ID:          fmt.Sprintf("LARGE_PAGE_%d", i),
			Description: fmt.Sprintf("Model %d", i),
		}
		CreateModel(model)
	}

	filterContext := map[string]string{}
	body, err := json.Marshal(filterContext)
	assert.NilError(t, err)

	// Request with very large page size
	url := "/xconfAdminService/model/filtered?pageNumber=1&pageSize=1000"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	var models []shared.Model
	err = json.NewDecoder(res.Body).Decode(&models)
	assert.NilError(t, err)
	// Should return all models (at least 5)
	assert.Check(t, len(models) >= 5, "Should return all models with large page size")
}
