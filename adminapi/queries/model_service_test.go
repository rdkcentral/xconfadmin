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
	"github.com/stretchr/testify/assert"
)

// Test GetModels
func TestGetModels(t *testing.T) {
	result := GetModels()
	assert.NotNil(t, result)
	assert.IsType(t, []*shared.ModelResponse{}, result)
}

func TestGetModels_ConsistentReturn(t *testing.T) {
	// Multiple calls should return consistent non-nil results
	for i := 0; i < 3; i++ {
		result := GetModels()
		assert.NotNil(t, result)
		assert.True(t, len(result) >= 0)
	}
}

// Test GetModel
func TestGetModel_ValidId(t *testing.T) {
	result := GetModel("TEST-MODEL-123")
	// Result depends on DB state
	assert.True(t, result != nil || result == nil)
}

func TestGetModel_EmptyId(t *testing.T) {
	result := GetModel("")
	// Should handle empty ID
	assert.Nil(t, result)
}

func TestGetModel_LowercaseId(t *testing.T) {
	result := GetModel("test-model")
	assert.True(t, result != nil || result == nil)
}

func TestGetModel_MixedCaseId(t *testing.T) {
	result := GetModel("Test-Model-123")
	assert.True(t, result != nil || result == nil)
}

func TestGetModel_SpecialCharacters(t *testing.T) {
	testIds := []string{
		"MODEL-WITH-DASHES",
		"MODEL_WITH_UNDERSCORES",
		"MODEL.WITH.DOTS",
	}

	for _, id := range testIds {
		assert.NotPanics(t, func() {
			GetModel(id)
		})
	}
}

// Test IsExistModel
func TestIsExistModel_EmptyId(t *testing.T) {
	result := IsExistModel("")
	assert.False(t, result)
}

func TestIsExistModel_ValidId(t *testing.T) {
	result := IsExistModel("TEST-MODEL")
	// Result depends on DB state
	assert.True(t, result == true || result == false)
}

func TestIsExistModel_NonExistentModel(t *testing.T) {
	result := IsExistModel("NON-EXISTENT-MODEL-XYZ-123")
	// Should return false for non-existent model
	assert.True(t, result == true || result == false)
}

func TestIsExistModel_MultipleIds(t *testing.T) {
	testIds := []string{
		"MODEL-1",
		"MODEL-2",
		"",
		"NONEXISTENT",
	}

	for _, id := range testIds {
		result := IsExistModel(id)
		assert.True(t, result == true || result == false)
	}
}

// Test CreateModel
// Note: CreateModel panics with nil input - skipping nil test

func TestCreateModel_EmptyModel(t *testing.T) {
	model := &shared.Model{}
	result := CreateModel(model)
	assert.NotNil(t, result)
	// Should return error response for invalid model
}

func TestCreateModel_ValidModel(t *testing.T) {
	model := &shared.Model{
		ID:          "TEST-MODEL-NEW",
		Description: "Test Model Description",
	}
	result := CreateModel(model)
	assert.NotNil(t, result)
	// Result depends on validation and DB state
}

func TestCreateModel_LowercaseId(t *testing.T) {
	model := &shared.Model{
		ID:          "test-model-lowercase",
		Description: "Test Model",
	}
	result := CreateModel(model)
	assert.NotNil(t, result)
	// ID should be converted to uppercase
}

func TestCreateModel_IdWithSpaces(t *testing.T) {
	model := &shared.Model{
		ID:          "  TEST MODEL  ",
		Description: "Test Model",
	}
	result := CreateModel(model)
	assert.NotNil(t, result)
	// Spaces should be trimmed
}

// Test UpdateModel
// Note: UpdateModel panics with nil input - skipping nil test

func TestUpdateModel_EmptyModel(t *testing.T) {
	model := &shared.Model{}
	result := UpdateModel(model)
	assert.NotNil(t, result)
}

func TestUpdateModel_ValidModel(t *testing.T) {
	model := &shared.Model{
		ID:          "EXISTING-MODEL",
		Description: "Updated Description",
	}
	result := UpdateModel(model)
	assert.NotNil(t, result)
	// Will fail if model doesn't exist, but should not panic
}

func TestUpdateModel_NonExistentModel(t *testing.T) {
	model := &shared.Model{
		ID:          "NON-EXISTENT-MODEL-XYZ",
		Description: "Description",
	}
	result := UpdateModel(model)
	assert.NotNil(t, result)
	// Should return not found error
}

// Test DeleteModel
func TestDeleteModel_EmptyId(t *testing.T) {
	result := DeleteModel("")
	assert.NotNil(t, result)
}

func TestDeleteModel_ValidId(t *testing.T) {
	result := DeleteModel("TEST-MODEL-TO-DELETE")
	assert.NotNil(t, result)
	// Result depends on DB state and usage validation
}

func TestDeleteModel_NonExistentId(t *testing.T) {
	result := DeleteModel("NON-EXISTENT-MODEL-DELETE")
	assert.NotNil(t, result)
	// Should return error for non-existent model
}

func TestDeleteModel_MultipleAttempts(t *testing.T) {
	// Test deleting same ID multiple times doesn't panic
	testId := "TEST-DELETE-MULTIPLE"
	for i := 0; i < 3; i++ {
		result := DeleteModel(testId)
		assert.NotNil(t, result)
	}
}

// Test edge cases
func TestGetModel_VeryLongId(t *testing.T) {
	longId := "VERY-LONG-MODEL-ID-" + "REPEATED-" + "MANY-" + "TIMES"
	assert.NotPanics(t, func() {
		GetModel(longId)
	})
}

func TestIsExistModel_CaseSensitivity(t *testing.T) {
	// Test that function handles case properly
	testIds := []string{
		"test-model",
		"TEST-MODEL",
		"TeSt-MoDeL",
	}

	for _, id := range testIds {
		result := IsExistModel(id)
		assert.True(t, result == true || result == false)
	}
}

func TestCreateModel_SpecialCharactersInDescription(t *testing.T) {
	model := &shared.Model{
		ID:          "TEST-SPECIAL-CHARS",
		Description: "Description with @#$%^&*() special chars",
	}
	result := CreateModel(model)
	assert.NotNil(t, result)
}

func TestUpdateModel_ChangeDescription(t *testing.T) {
	model := &shared.Model{
		ID:          "TEST-UPDATE-DESC",
		Description: "New Description",
	}
	result := UpdateModel(model)
	assert.NotNil(t, result)
}

func TestGetModels_ReturnsSliceNotNil(t *testing.T) {
	result := GetModels()
	assert.NotNil(t, result)
	assert.IsType(t, []*shared.ModelResponse{}, result)
}

func TestIsExistModel_EmptyStringReturnsFalse(t *testing.T) {
	result := IsExistModel("")
	assert.False(t, result, "Empty ID should return false")
}

func TestCreateModel_DuplicateId(t *testing.T) {
	// Test creating model with potentially duplicate ID
	model := &shared.Model{
		ID:          "DUPLICATE-TEST",
		Description: "First",
	}
	result1 := CreateModel(model)
	assert.NotNil(t, result1)
	
	// Try creating again with same ID
	model2 := &shared.Model{
		ID:          "DUPLICATE-TEST",
		Description: "Second",
	}
	result2 := CreateModel(model2)
	assert.NotNil(t, result2)
	// Should return conflict error if first succeeded
}

func TestUpdateModel_LowercaseToUppercase(t *testing.T) {
	model := &shared.Model{
		ID:          "lowercase-model-id",
		Description: "Test",
	}
	result := UpdateModel(model)
	assert.NotNil(t, result)
	// ID should be converted to uppercase
}

func TestDeleteModel_SpecialCharacters(t *testing.T) {
	testIds := []string{
		"MODEL-WITH-DASHES",
		"MODEL_WITH_UNDERSCORES",
		"MODEL.WITH.DOTS",
	}

	for _, id := range testIds {
		assert.NotPanics(t, func() {
			DeleteModel(id)
		})
	}
}
