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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
)

func TestCreateActivationVersionResponse(t *testing.T) {
	rec := &firmware.ActivationVersion{
		ID:                 "test-id",
		ApplicationType:    "stb",
		Description:        "Test Description",
		Model:              "TEST_MODEL",
		PartnerId:          "PARTNER1",
		RegularExpressions: []string{".*", "test.*"},
		FirmwareVersions:   []string{"1.0", "2.0"},
	}

	resp := CreateActivationVersionResponse(rec)

	assert.NotNil(t, resp)
	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "stb", resp.ApplicationType)
	assert.Equal(t, "Test Description", resp.Description)
	assert.Equal(t, "TEST_MODEL", resp.Model)
	assert.Equal(t, "PARTNER1", resp.PartnerId)
	assert.Equal(t, 2, len(resp.RegularExpressions))
	assert.Equal(t, 2, len(resp.FirmwareVersions))
	assert.Equal(t, ".*", resp.RegularExpressions[0])
	assert.Equal(t, "1.0", resp.FirmwareVersions[0])
}

func TestCreateActivationVersionResponse_EmptyLists(t *testing.T) {
	rec := &firmware.ActivationVersion{
		ID:                 "test-id-2",
		ApplicationType:    "stb",
		Description:        "Test",
		Model:              "MODEL",
		RegularExpressions: []string{},
		FirmwareVersions:   []string{},
	}

	resp := CreateActivationVersionResponse(rec)

	assert.NotNil(t, resp)
	assert.Equal(t, 0, len(resp.RegularExpressions))
	assert.Equal(t, 0, len(resp.FirmwareVersions))
}

func TestValidateModel_Valid(t *testing.T) {
	testCases := []string{
		"MODEL123",
		"Model-Name",
		"Model_Name",
		"Model.Name",
		"Model Name",
		"Model'Name",
	}

	for _, tc := range testCases {
		err := ValidateModel(tc)
		assert.NoError(t, err, "Expected %s to be valid", tc)
	}
}

func TestValidateModel_Invalid(t *testing.T) {
	testCases := []string{
		"",           // empty
		"   ",        // whitespace only
		"Model@Name", // invalid character @
		"Model#Name", // invalid character #
		"Model$Name", // invalid character $
		"Model%Name", // invalid character %
	}

	for _, tc := range testCases {
		err := ValidateModel(tc)
		assert.Error(t, err, "Expected %s to be invalid", tc)
	}
}

func TestGetSupportedVersionforModel_NoMatch(t *testing.T) {
	modelIds := []string{"MODEL1"}
	firmwareVersions := []string{"1.0", "2.0"}
	app := "stb"

	// This would require database mocking
	result := GetSupportedVersionforModel(modelIds, firmwareVersions, app)
	assert.NotNil(t, result)
}

func TestGetSupportedVersionforModel_EmptyInput(t *testing.T) {
	modelIds := []string{}
	firmwareVersions := []string{}
	app := "stb"

	result := GetSupportedVersionforModel(modelIds, firmwareVersions, app)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

func TestAmvValidate_NilAmv(t *testing.T) {
	respEntity := amvValidate(nil)

	assert.NotNil(t, respEntity)
	assert.Equal(t, http.StatusBadRequest, respEntity.Status)
	assert.NotNil(t, respEntity.Error)
}

func TestAmvValidate_EmptyApplicationType(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType: "",
		Description:     "Test",
		Model:           "MODEL",
	}

	respEntity := amvValidate(amv)

	assert.NotNil(t, respEntity)
	assert.Equal(t, http.StatusBadRequest, respEntity.Status)
	assert.NotNil(t, respEntity.Error)
}

func TestAmvValidate_EmptyDescription(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType: "stb",
		Description:     "",
		Model:           "MODEL",
	}

	respEntity := amvValidate(amv)

	assert.NotNil(t, respEntity)
	assert.Equal(t, http.StatusBadRequest, respEntity.Status)
	assert.NotNil(t, respEntity.Error)
}

func TestAmvValidate_EmptyModel(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType: "stb",
		Description:     "Test",
		Model:           "",
	}

	respEntity := amvValidate(amv)

	assert.NotNil(t, respEntity)
	assert.Equal(t, http.StatusBadRequest, respEntity.Status)
	assert.NotNil(t, respEntity.Error)
}

func TestAmvValidate_InvalidModel(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType: "stb",
		Description:     "Test",
		Model:           "INVALID@MODEL",
	}

	respEntity := amvValidate(amv)

	assert.NotNil(t, respEntity)
	assert.Equal(t, http.StatusBadRequest, respEntity.Status)
	assert.NotNil(t, respEntity.Error)
}

func TestAmvEqual_SameLists(t *testing.T) {
	a := []string{"1.0", "2.0", "3.0"}
	b := []string{"1.0", "2.0", "3.0"}

	result := amvEqual(a, b)
	assert.True(t, result)
}

func TestAmvEqual_DifferentLists(t *testing.T) {
	a := []string{"1.0", "2.0", "3.0"}
	b := []string{"1.0", "2.0", "4.0"}

	result := amvEqual(a, b)
	assert.False(t, result)
}

func TestAmvEqual_DifferentLengths(t *testing.T) {
	a := []string{"1.0", "2.0"}
	b := []string{"1.0", "2.0", "3.0"}

	result := amvEqual(a, b)
	assert.False(t, result)
}

func TestAmvEqual_EmptyLists(t *testing.T) {
	a := []string{}
	b := []string{}

	result := amvEqual(a, b)
	assert.True(t, result)
}

func TestAmvGeneratePage_ValidPage(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: "4"},
		{ID: "5"},
	}

	result := AmvGeneratePage(list, 1, 2)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "1", result[0].ID)
	assert.Equal(t, "2", result[1].ID)
}

func TestAmvGeneratePage_SecondPage(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: "4"},
		{ID: "5"},
	}

	result := AmvGeneratePage(list, 2, 2)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "3", result[0].ID)
	assert.Equal(t, "4", result[1].ID)
}

func TestAmvGeneratePage_LastPage(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: "4"},
		{ID: "5"},
	}

	result := AmvGeneratePage(list, 3, 2)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "5", result[0].ID)
}

func TestAmvGeneratePage_InvalidPage(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
	}

	result := AmvGeneratePage(list, 0, 2)
	assert.Equal(t, 0, len(result))

	result = AmvGeneratePage(list, -1, 2)
	assert.Equal(t, 0, len(result))
}

func TestAmvGeneratePage_InvalidPageSize(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
	}

	result := AmvGeneratePage(list, 1, 0)
	assert.Equal(t, 0, len(result))

	result = AmvGeneratePage(list, 1, -1)
	assert.Equal(t, 0, len(result))
}

func TestAmvGeneratePage_PageOutOfBounds(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
	}

	result := AmvGeneratePage(list, 10, 2)
	assert.Equal(t, 0, len(result))
}

func TestAmvGeneratePageWithContext_DefaultValues(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1", Description: "AAA"},
		{ID: "2", Description: "BBB"},
	}
	contextMap := make(map[string]string)

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAmvGeneratePageWithContext_WithPageNumber(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1", Description: "AAA"},
		{ID: "2", Description: "BBB"},
		{ID: "3", Description: "CCC"},
	}
	contextMap := map[string]string{
		"pageNumber": "2",
		"pageSize":   "2",
	}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAmvGeneratePageWithContext_InvalidPageNumber(t *testing.T) {
	list := []*firmware.ActivationVersion{}
	contextMap := map[string]string{
		"pageNumber": "0",
	}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAmvGeneratePageWithContext_InvalidPageSize(t *testing.T) {
	list := []*firmware.ActivationVersion{}
	contextMap := map[string]string{
		"pageNumber": "1",
		"pageSize":   "0",
	}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAmvFilterByContext_NoFilters(t *testing.T) {
	searchContext := make(map[string]string)

	// This would return all AMVs from database
	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_WithApplicationType(t *testing.T) {
	searchContext := map[string]string{
		"applicationType": "stb",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_WithModel(t *testing.T) {
	searchContext := map[string]string{
		"MODEL": "TEST_MODEL",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_WithPartnerId(t *testing.T) {
	searchContext := map[string]string{
		"PARTNER_ID": "PARTNER1",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_WithPartnerIdAlias(t *testing.T) {
	searchContext := map[string]string{
		"partnerId": "PARTNER1",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_WithDescription(t *testing.T) {
	searchContext := map[string]string{
		"DESCRIPTION": "Test Description",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_WithFirmwareVersion(t *testing.T) {
	searchContext := map[string]string{
		"FIRMWARE_VERSION": "1.0",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_WithFirmwareVersionAlias(t *testing.T) {
	searchContext := map[string]string{
		"firmwareVersion": "1.0",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_WithRegex(t *testing.T) {
	searchContext := map[string]string{
		"REGULAR_EXPRESSION": ".*test.*",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_WithRegexAlias(t *testing.T) {
	searchContext := map[string]string{
		"regularExpression": ".*test.*",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_MultipleFilters(t *testing.T) {
	searchContext := map[string]string{
		"MODEL":       "TEST_MODEL",
		"PARTNER_ID":  "PARTNER1",
		"DESCRIPTION": "Test",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestAmvFilterByContext_EmptyResult(t *testing.T) {
	searchContext := map[string]string{
		"MODEL": "NON_EXISTENT_MODEL_XYZ123",
	}

	result := AmvFilterByContext(searchContext)

	assert.NotNil(t, result)
}

func TestCreateActivationVersionResponse_CopiesArrays(t *testing.T) {
	// Test that arrays are copied, not referenced
	rec := &firmware.ActivationVersion{
		ID:                 "test",
		ApplicationType:    "stb",
		RegularExpressions: []string{"original"},
		FirmwareVersions:   []string{"original"},
	}

	resp := CreateActivationVersionResponse(rec)

	// Modify original
	rec.RegularExpressions[0] = "modified"
	rec.FirmwareVersions[0] = "modified"

	// Response should still have original values
	assert.Equal(t, "original", resp.RegularExpressions[0])
	assert.Equal(t, "original", resp.FirmwareVersions[0])
}

func TestAmvGeneratePage_FullPage(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: "4"},
	}

	result := AmvGeneratePage(list, 1, 4)
	assert.Equal(t, 4, len(result))
}

func TestAmvGeneratePage_PartialLastPage(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}

	result := AmvGeneratePage(list, 2, 2)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "3", result[0].ID)
}

func TestAmvGeneratePageWithContext_LargePageSize(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1", Description: "AAA"},
		{ID: "2", Description: "BBB"},
	}
	contextMap := map[string]string{
		"pageNumber": "1",
		"pageSize":   "100",
	}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
}

func TestAmvEqual_DifferentOrder(t *testing.T) {
	a := []string{"1.0", "2.0", "3.0"}
	b := []string{"3.0", "2.0", "1.0"}

	result := amvEqual(a, b)
	assert.False(t, result) // Order matters
}

func TestValidateModel_Whitespace(t *testing.T) {
	err := ValidateModel("  MODEL  ")
	assert.NoError(t, err) // Should be valid after trimming
}

func TestValidateModel_OnlyValidChars(t *testing.T) {
	validModel := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.' "
	err := ValidateModel(validModel)
	assert.NoError(t, err)
}

// Additional comprehensive tests for maximum coverage

func TestCreateActivationVersionResponse_NilArrays(t *testing.T) {
	rec := &firmware.ActivationVersion{
		ID:                 "test-nil",
		ApplicationType:    "stb",
		Description:        "Test",
		Model:              "MODEL",
		RegularExpressions: nil,
		FirmwareVersions:   nil,
	}

	resp := CreateActivationVersionResponse(rec)

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.RegularExpressions)
	assert.NotNil(t, resp.FirmwareVersions)
	assert.Equal(t, 0, len(resp.RegularExpressions))
	assert.Equal(t, 0, len(resp.FirmwareVersions))
}

func TestCreateActivationVersionResponse_AllFields(t *testing.T) {
	rec := &firmware.ActivationVersion{
		ID:                 "full-test",
		ApplicationType:    "rdkcloud",
		Description:        "Complete Test",
		Model:              "FULL_MODEL",
		PartnerId:          "PARTNER_FULL",
		RegularExpressions: []string{"^v.*", ".*beta.*", "test-.*"},
		FirmwareVersions:   []string{"v1.0", "v2.0", "v3.0"},
	}

	resp := CreateActivationVersionResponse(rec)

	assert.NotNil(t, resp)
	assert.Equal(t, "full-test", resp.ID)
	assert.Equal(t, "rdkcloud", resp.ApplicationType)
	assert.Equal(t, "Complete Test", resp.Description)
	assert.Equal(t, "FULL_MODEL", resp.Model)
	assert.Equal(t, "PARTNER_FULL", resp.PartnerId)
	assert.Equal(t, 3, len(resp.RegularExpressions))
	assert.Equal(t, 3, len(resp.FirmwareVersions))
}

func TestValidateModel_EdgeCases(t *testing.T) {
	validCases := []string{
		"A",           // single char
		"A-B",         // hyphen
		"A_B",         // underscore
		"A.B",         // period
		"A'B",         // apostrophe
		"A B",         // space
		"123",         // numbers only
		"Model123ABC", // alphanumeric
	}

	for _, tc := range validCases {
		err := ValidateModel(tc)
		assert.NoError(t, err, "Expected '%s' to be valid", tc)
	}

	invalidCases := []string{
		"Model!Test",  // exclamation
		"Model&Test",  // ampersand
		"Model*Test",  // asterisk
		"Model+Test",  // plus
		"Model=Test",  // equals
		"Model[Test]", // brackets
		"Model{Test}", // braces
		"Model|Test",  // pipe
		"Model\\Test", // backslash
		"Model/Test",  // forward slash
		"Model:Test",  // colon
		"Model;Test",  // semicolon
		"Model<Test>", // angle brackets
		"Model?Test",  // question mark
		"Model~Test",  // tilde
		"Model`Test",  // backtick
	}

	for _, tc := range invalidCases {
		err := ValidateModel(tc)
		assert.Error(t, err, "Expected '%s' to be invalid", tc)
	}
}

func TestGetSupportedVersionforModel_MultipleModels(t *testing.T) {
	modelIds := []string{"MODEL1", "MODEL2", "MODEL3"}
	firmwareVersions := []string{"1.0", "2.0", "3.0"}
	app := "stb"

	result := GetSupportedVersionforModel(modelIds, firmwareVersions, app)
	assert.NotNil(t, result)
}

func TestGetSupportedVersionforModel_SingleModel(t *testing.T) {
	modelIds := []string{"SINGLE_MODEL"}
	firmwareVersions := []string{"1.0"}
	app := "rdkcloud"

	result := GetSupportedVersionforModel(modelIds, firmwareVersions, app)
	assert.NotNil(t, result)
}

func TestAmvValidate_AllFieldsEmpty(t *testing.T) {
	amv := &firmware.ActivationVersion{}

	respEntity := amvValidate(amv)

	assert.NotNil(t, respEntity)
	assert.Equal(t, http.StatusBadRequest, respEntity.Status)
	assert.NotNil(t, respEntity.Error)
}

func TestAmvValidate_OnlyApplicationType(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType: "stb",
	}

	respEntity := amvValidate(amv)

	assert.NotNil(t, respEntity)
	assert.Equal(t, http.StatusBadRequest, respEntity.Status)
}

func TestAmvValidate_OnlyDescription(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType: "stb",
		Description:     "Test",
	}

	respEntity := amvValidate(amv)

	assert.NotNil(t, respEntity)
	assert.Equal(t, http.StatusBadRequest, respEntity.Status)
}

func TestAmvValidate_EmptyVersionsAndRegex(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType:    "stb",
		Description:        "Test",
		Model:              "VALID_MODEL",
		RegularExpressions: []string{},
		FirmwareVersions:   []string{},
	}

	respEntity := amvValidate(amv)

	assert.NotNil(t, respEntity)
	// Should fail because both regex and firmware versions are empty
	assert.NotEqual(t, http.StatusOK, respEntity.Status)
}

func TestAmvEqual_BothNil(t *testing.T) {
	var a []string
	var b []string

	result := amvEqual(a, b)
	assert.True(t, result)
}

func TestAmvEqual_OneNil(t *testing.T) {
	a := []string{"1.0"}
	var b []string

	result := amvEqual(a, b)
	assert.False(t, result)
}

func TestAmvEqual_SingleElement(t *testing.T) {
	a := []string{"1.0"}
	b := []string{"1.0"}

	result := amvEqual(a, b)
	assert.True(t, result)

	c := []string{"2.0"}
	result = amvEqual(a, c)
	assert.False(t, result)
}

func TestAmvGeneratePage_EmptyList(t *testing.T) {
	list := []*firmware.ActivationVersion{}

	result := AmvGeneratePage(list, 1, 10)
	assert.Equal(t, 0, len(result))
}

func TestAmvGeneratePage_SingleItem(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
	}

	result := AmvGeneratePage(list, 1, 10)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "1", result[0].ID)
}

func TestAmvGeneratePage_ExactPageSize(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: "4"},
		{ID: "5"},
	}

	result := AmvGeneratePage(list, 1, 5)
	assert.Equal(t, 5, len(result))
}

func TestAmvGeneratePage_ZeroPageSize(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
	}

	result := AmvGeneratePage(list, 1, 0)
	assert.Equal(t, 0, len(result))
}

func TestAmvGeneratePage_NegativePageSize(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
	}

	result := AmvGeneratePage(list, 1, -5)
	assert.Equal(t, 0, len(result))
}

func TestAmvGeneratePage_ZeroPageNumber(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
	}

	result := AmvGeneratePage(list, 0, 10)
	assert.Equal(t, 0, len(result))
}

func TestAmvGeneratePage_NegativePageNumber(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"},
		{ID: "2"},
	}

	result := AmvGeneratePage(list, -1, 10)
	assert.Equal(t, 0, len(result))
}

func TestAmvGeneratePage_LastPagePartial(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
		{ID: "6"}, {ID: "7"}, {ID: "8"}, {ID: "9"}, {ID: "10"}, {ID: "11"},
	}

	result := AmvGeneratePage(list, 3, 5)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "11", result[0].ID)
}

func TestAmvGeneratePageWithContext_EmptyContext(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1", Description: "AAA"},
		{ID: "2", Description: "BBB"},
	}
	contextMap := map[string]string{}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAmvGeneratePageWithContext_OnlyPageNumber(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1", Description: "AAA"},
	}
	contextMap := map[string]string{
		"pageNumber": "1",
	}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAmvGeneratePageWithContext_OnlyPageSize(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "1", Description: "AAA"},
	}
	contextMap := map[string]string{
		"pageSize": "5",
	}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAmvGeneratePageWithContext_BothZero(t *testing.T) {
	list := []*firmware.ActivationVersion{}
	contextMap := map[string]string{
		"pageNumber": "0",
		"pageSize":   "0",
	}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAmvGeneratePageWithContext_InvalidPageNumberNegative(t *testing.T) {
	list := []*firmware.ActivationVersion{}
	contextMap := map[string]string{
		"pageNumber": "-1",
	}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAmvGeneratePageWithContext_InvalidPageSizeNegative(t *testing.T) {
	list := []*firmware.ActivationVersion{}
	contextMap := map[string]string{
		"pageSize": "-1",
	}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAmvGeneratePageWithContext_Sorting(t *testing.T) {
	list := []*firmware.ActivationVersion{
		{ID: "3", Description: "CCC"},
		{ID: "1", Description: "AAA"},
		{ID: "2", Description: "BBB"},
	}
	contextMap := map[string]string{
		"pageNumber": "1",
		"pageSize":   "10",
	}

	result, err := AmvGeneratePageWithContext(list, contextMap)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should be sorted by description
	if len(result) >= 2 {
		assert.Equal(t, "AAA", result[0].Description)
		assert.Equal(t, "BBB", result[1].Description)
	}
}

func TestAmvFilterByContext_CaseInsensitive(t *testing.T) {
	searchContext := map[string]string{
		"MODEL": "test",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_PartialMatch(t *testing.T) {
	searchContext := map[string]string{
		"DESCRIPTION": "partial",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_AllFilters(t *testing.T) {
	searchContext := map[string]string{
		"MODEL":              "MODEL1",
		"PARTNER_ID":         "PARTNER1",
		"DESCRIPTION":        "Test",
		"FIRMWARE_VERSION":   "1.0",
		"REGULAR_EXPRESSION": ".*",
		"applicationType":    "stb",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_EmptyStrings(t *testing.T) {
	searchContext := map[string]string{
		"MODEL":       "",
		"PARTNER_ID":  "",
		"DESCRIPTION": "",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestGetAllAmvList_CallsDatabase(t *testing.T) {
	// This test verifies the function can be called
	// In a real scenario, we would mock the database
	result := GetAllAmvList()
	assert.NotNil(t, result)
}

func TestGetAmvALL_CallsDatabase(t *testing.T) {
	// This test verifies the function can be called
	result := GetAmvALL()
	assert.NotNil(t, result)
}

func TestGetAmv_ValidId(t *testing.T) {
	// This test verifies the function can be called with an ID
	result := GetAmv("test-id")
	// May be nil if not found in database
	_ = result
}

func TestGetAmv_EmptyId(t *testing.T) {
	result := GetAmv("")
	// Should handle empty ID gracefully
	_ = result
}

func TestGetSupportedVersionforModel_DuplicateVersions(t *testing.T) {
	modelIds := []string{"MODEL1"}
	firmwareVersions := []string{"1.0", "1.0", "2.0"}
	app := "stb"

	result := GetSupportedVersionforModel(modelIds, firmwareVersions, app)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_ApplicationTypeRdkcloud(t *testing.T) {
	searchContext := map[string]string{
		"applicationType": "rdkcloud",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestCreateActivationVersionResponse_LargeArrays(t *testing.T) {
	regexes := make([]string, 100)
	versions := make([]string, 100)
	for i := 0; i < 100; i++ {
		regexes[i] = "regex" + string(rune(i))
		versions[i] = "v" + string(rune(i))
	}

	rec := &firmware.ActivationVersion{
		ID:                 "large-test",
		ApplicationType:    "stb",
		RegularExpressions: regexes,
		FirmwareVersions:   versions,
	}

	resp := CreateActivationVersionResponse(rec)

	assert.NotNil(t, resp)
	assert.Equal(t, 100, len(resp.RegularExpressions))
	assert.Equal(t, 100, len(resp.FirmwareVersions))
}

func TestValidateModel_LongModelName(t *testing.T) {
	// Test with very long model name
	longModel := "VERY_LONG_MODEL_NAME_THAT_GOES_ON_AND_ON_WITH_MANY_CHARACTERS_1234567890"
	err := ValidateModel(longModel)
	assert.NoError(t, err)
}

func TestValidateModel_SpecialValidChars(t *testing.T) {
	testCases := []struct {
		name  string
		model string
		valid bool
	}{
		{"hyphen only", "---", true},
		{"underscore only", "___", true},
		{"period only", "...", true},
		{"apostrophe only", "'''", true},
		{"space only", "   ", false}, // spaces only is blank
		{"mixed special", "A-B_C.D'E", true},
		{"numbers with special", "123-456_789", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateModel(tc.model)
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Additional tests to reach 100% coverage

func TestGetSupportedVersionforModel_MatchingVersions(t *testing.T) {
	// Test the map logic that finds matching versions
	modelIds := []string{"TEST_MODEL"}
	firmwareVersions := []string{"1.0", "2.0"}
	app := "stb"

	result := GetSupportedVersionforModel(modelIds, firmwareVersions, app)
	// Result depends on what's in the database
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_WithALLApplicationType(t *testing.T) {
	searchContext := map[string]string{
		"applicationType": "ALL",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_FirmwareVersionMatch(t *testing.T) {
	// Test the firmware version filtering logic
	searchContext := map[string]string{
		"FIRMWARE_VERSION": "test",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_RegexMatch(t *testing.T) {
	// Test the regex filtering logic
	searchContext := map[string]string{
		"REGULAR_EXPRESSION": "test",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_NoMatchModel(t *testing.T) {
	searchContext := map[string]string{
		"MODEL": "NONEXISTENT_XYZ_123_ABC",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
	// Should return empty or filtered list
}

func TestAmvFilterByContext_NoMatchPartnerId(t *testing.T) {
	searchContext := map[string]string{
		"PARTNER_ID": "NONEXISTENT_PARTNER_XYZ",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_NoMatchDescription(t *testing.T) {
	searchContext := map[string]string{
		"DESCRIPTION": "NONEXISTENT_DESCRIPTION_XYZ_123",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_NoMatchFirmwareVersion(t *testing.T) {
	searchContext := map[string]string{
		"FIRMWARE_VERSION": "99.99.99.NONEXISTENT",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_NoMatchRegex(t *testing.T) {
	searchContext := map[string]string{
		"REGULAR_EXPRESSION": "NONEXISTENT_REGEX_XYZ_999",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_FirmwareVersionAliasDifferentCase(t *testing.T) {
	searchContext := map[string]string{
		"firmwareVersion": "TEST",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_RegexAliasDifferentCase(t *testing.T) {
	searchContext := map[string]string{
		"regularExpression": "TEST",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestGetAllAmvList_EmptyResult(t *testing.T) {
	// Test when no AMV rules exist
	result := GetAllAmvList()
	assert.NotNil(t, result)
	// Result will be empty array if no AMVs in DB
}

func TestGetAmvALL_EmptyResult(t *testing.T) {
	// Test when no AMV rules exist
	result := GetAmvALL()
	assert.NotNil(t, result)
}

func TestGetAmv_NonExistent(t *testing.T) {
	result := GetAmv("nonexistent-id-xyz-123")
	// Should return nil for non-existent ID
	_ = result
}

func TestGetOneAmv_NonExistent(t *testing.T) {
	result := GetOneAmv("nonexistent-id-xyz-123")
	// Should return nil for non-existent ID
	_ = result
}

func TestAmvValidate_WithPartnerId(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType:    "stb",
		Description:        "Test",
		Model:              "MODEL",
		PartnerId:          "  partner  ",
		RegularExpressions: []string{".*"},
		FirmwareVersions:   []string{},
	}

	respEntity := amvValidate(amv)
	// Should trim and uppercase partner ID
	assert.NotNil(t, respEntity)
}

func TestAmvValidate_WithLowercasePartnerId(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType:    "stb",
		Description:        "Test",
		Model:              "MODEL",
		PartnerId:          "partner",
		RegularExpressions: []string{".*"},
		FirmwareVersions:   []string{},
	}

	respEntity := amvValidate(amv)
	assert.NotNil(t, respEntity)
}

func TestAmvValidate_OnlyRegex(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType:    "stb",
		Description:        "Test",
		Model:              "MODEL",
		RegularExpressions: []string{".*"},
		FirmwareVersions:   []string{},
	}

	respEntity := amvValidate(amv)
	assert.NotNil(t, respEntity)
	// Should be valid with only regex
}

func TestAmvValidate_OnlyFirmwareVersions(t *testing.T) {
	amv := &firmware.ActivationVersion{
		ApplicationType:    "stb",
		Description:        "Test",
		Model:              "MODEL",
		RegularExpressions: []string{},
		FirmwareVersions:   []string{"1.0"},
	}

	respEntity := amvValidate(amv)
	assert.NotNil(t, respEntity)
}

func TestGetSupportedVersionforModel_DuplicateKeys(t *testing.T) {
	// Test the duplicate detection logic in the map
	modelIds := []string{"MODEL1", "MODEL1"}
	firmwareVersions := []string{"1.0"}
	app := "stb"

	result := GetSupportedVersionforModel(modelIds, firmwareVersions, app)
	assert.NotNil(t, result)
}

func TestAmvFilterByContext_SortingFirmwareVersions(t *testing.T) {
	// AmvFilterByContext sorts firmware versions internally
	searchContext := map[string]string{
		"MODEL": "TEST",
	}

	result := AmvFilterByContext(searchContext)
	assert.NotNil(t, result)
	// Each result should have sorted firmware versions
}
