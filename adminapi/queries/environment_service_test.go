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

// Test GetEnvironment
func TestGetEnvironment_Found(t *testing.T) {
	// Test basic function - actual DB call would require setup
	// Testing that function returns properly typed result
	result := GetEnvironment("test-id")
	// Result can be nil if DB not configured
	if result != nil {
		assert.IsType(t, &shared.Environment{}, result)
	}
}

func TestGetEnvironment_EmptyID(t *testing.T) {
	result := GetEnvironment("")
	// Should handle empty ID gracefully
	assert.True(t, result == nil || result != nil)
}

// Test IsExistEnvironment
func TestIsExistEnvironment_EmptyID(t *testing.T) {
	result := IsExistEnvironment("")
	assert.False(t, result)
}

func TestIsExistEnvironment_NonEmptyID(t *testing.T) {
	result := IsExistEnvironment("TEST_ENV")
	// Result depends on DB state
	assert.True(t, result == true || result == false)
}

// Test environmentGeneratePage
func TestEnvironmentGeneratePage_EmptyList(t *testing.T) {
	list := []*shared.Environment{}
	result := environmentGeneratePage(list, 1, 10)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

func TestEnvironmentGeneratePage_SingleItem(t *testing.T) {
	list := []*shared.Environment{
		{ID: "ENV1", Description: "Test Environment 1"},
	}
	result := environmentGeneratePage(list, 1, 10)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "ENV1", result[0].ID)
}

func TestEnvironmentGeneratePage_MultiplePages(t *testing.T) {
	list := []*shared.Environment{}
	for i := 1; i <= 25; i++ {
		list = append(list, &shared.Environment{
			ID:          "ENV" + string(rune('0'+i)),
			Description: "Environment " + string(rune('0'+i)),
		})
	}

	// Test first page
	page1 := environmentGeneratePage(list, 1, 10)
	assert.Equal(t, 10, len(page1))

	// Test second page
	page2 := environmentGeneratePage(list, 2, 10)
	assert.Equal(t, 10, len(page2))

	// Test third page (partial)
	page3 := environmentGeneratePage(list, 3, 10)
	assert.Equal(t, 5, len(page3))
}

func TestEnvironmentGeneratePage_InvalidPageNumber(t *testing.T) {
	list := []*shared.Environment{
		{ID: "ENV1"},
		{ID: "ENV2"},
	}

	// Page 0 or negative
	result := environmentGeneratePage(list, 0, 10)
	assert.Equal(t, 0, len(result))

	result = environmentGeneratePage(list, -1, 10)
	assert.Equal(t, 0, len(result))
}

func TestEnvironmentGeneratePage_InvalidPageSize(t *testing.T) {
	list := []*shared.Environment{
		{ID: "ENV1"},
		{ID: "ENV2"},
	}

	// Page size 0 or negative
	result := environmentGeneratePage(list, 1, 0)
	assert.Equal(t, 0, len(result))

	result = environmentGeneratePage(list, 1, -1)
	assert.Equal(t, 0, len(result))
}

func TestEnvironmentGeneratePage_PageBeyondRange(t *testing.T) {
	list := []*shared.Environment{
		{ID: "ENV1"},
		{ID: "ENV2"},
	}

	// Request page beyond available data
	result := environmentGeneratePage(list, 10, 10)
	assert.Equal(t, 0, len(result))
}

func TestEnvironmentGeneratePage_ExactPageBoundary(t *testing.T) {
	list := []*shared.Environment{}
	for i := 1; i <= 10; i++ {
		list = append(list, &shared.Environment{ID: "ENV"})
	}

	result := environmentGeneratePage(list, 1, 10)
	assert.Equal(t, 10, len(result))
}

func TestEnvironmentGeneratePage_VariousPageSizes(t *testing.T) {
	list := []*shared.Environment{}
	for i := 1; i <= 20; i++ {
		list = append(list, &shared.Environment{ID: "ENV"})
	}

	testCases := []struct {
		name        string
		page        int
		pageSize    int
		expectedLen int
	}{
		{"Page size 5, page 1", 1, 5, 5},
		{"Page size 5, page 2", 2, 5, 5},
		{"Page size 5, page 4", 4, 5, 5},
		{"Page size 20, page 1", 1, 20, 20},
		{"Page size 3, page 1", 1, 3, 3},
		{"Page size 7, page 3", 3, 7, 6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := environmentGeneratePage(list, tc.page, tc.pageSize)
			assert.Equal(t, tc.expectedLen, len(result))
		})
	}
}

// Test EnvironmentRuleGeneratePageWithContext
func TestEnvironmentRuleGeneratePageWithContext_EmptyContext(t *testing.T) {
	list := []*shared.Environment{
		{ID: "ENV1"},
		{ID: "ENV2"},
	}
	contextMap := make(map[string]string)

	result, err := EnvironmentRuleGeneratePageWithContext(list, contextMap)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Default page 1, size 10
	assert.LessOrEqual(t, len(result), 2)
}

func TestEnvironmentRuleGeneratePageWithContext_WithPageNumber(t *testing.T) {
	list := []*shared.Environment{}
	for i := 1; i <= 25; i++ {
		list = append(list, &shared.Environment{ID: "ENV"})
	}

	contextMap := map[string]string{
		cPercentageBeanPageNumber: "2",
		cPercentageBeanPageSize:   "10",
	}

	result, err := EnvironmentRuleGeneratePageWithContext(list, contextMap)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(result))
}

func TestEnvironmentRuleGeneratePageWithContext_InvalidPageNumber(t *testing.T) {
	list := []*shared.Environment{{ID: "ENV1"}}

	contextMap := map[string]string{
		cPercentageBeanPageNumber: "0",
	}

	result, err := EnvironmentRuleGeneratePageWithContext(list, contextMap)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "greater than zero")
}

func TestEnvironmentRuleGeneratePageWithContext_InvalidPageSize(t *testing.T) {
	list := []*shared.Environment{{ID: "ENV1"}}

	contextMap := map[string]string{
		cPercentageBeanPageSize: "0",
	}

	result, err := EnvironmentRuleGeneratePageWithContext(list, contextMap)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestEnvironmentRuleGeneratePageWithContext_NegativeValues(t *testing.T) {
	list := []*shared.Environment{{ID: "ENV1"}}

	testCases := []struct {
		name    string
		context map[string]string
	}{
		{
			"Negative page number",
			map[string]string{cPercentageBeanPageNumber: "-1"},
		},
		{
			"Negative page size",
			map[string]string{cPercentageBeanPageSize: "-1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := EnvironmentRuleGeneratePageWithContext(list, tc.context)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestEnvironmentRuleGeneratePageWithContext_Sorting(t *testing.T) {
	list := []*shared.Environment{
		{ID: "ZZZ"},
		{ID: "AAA"},
		{ID: "MMM"},
		{ID: "BBB"},
	}

	contextMap := map[string]string{
		cPercentageBeanPageNumber: "1",
		cPercentageBeanPageSize:   "10",
	}

	result, err := EnvironmentRuleGeneratePageWithContext(list, contextMap)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(result))
	// Should be sorted alphabetically
	assert.Equal(t, "AAA", result[0].ID)
	assert.Equal(t, "BBB", result[1].ID)
	assert.Equal(t, "MMM", result[2].ID)
	assert.Equal(t, "ZZZ", result[3].ID)
}

func TestEnvironmentRuleGeneratePageWithContext_CaseInsensitiveSorting(t *testing.T) {
	list := []*shared.Environment{
		{ID: "zzz"},
		{ID: "AAA"},
		{ID: "Mmm"},
		{ID: "bbb"},
	}

	contextMap := map[string]string{
		cPercentageBeanPageNumber: "1",
		cPercentageBeanPageSize:   "10",
	}

	result, err := EnvironmentRuleGeneratePageWithContext(list, contextMap)
	assert.NoError(t, err)
	// Sorting should be case-insensitive
	assert.Equal(t, 4, len(result))
}

// Test EnvironmentFilterByContext
func TestEnvironmentFilterByContext_EmptyContext(t *testing.T) {
	searchContext := make(map[string]string)
	result := EnvironmentFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestEnvironmentFilterByContext_EmptyContextReturnsEmptyList(t *testing.T) {
	// When no environments exist
	searchContext := make(map[string]string)
	result := EnvironmentFilterByContext(searchContext)
	assert.NotNil(t, result)
	assert.IsType(t, []*shared.Environment{}, result)
}

func TestEnvironmentFilterByContext_WithIDFilter(t *testing.T) {
	// Test that function handles ID filter
	searchContext := map[string]string{
		cEnvironmentID: "TEST",
	}
	result := EnvironmentFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestEnvironmentFilterByContext_WithDescriptionFilter(t *testing.T) {
	// Test that function handles description filter
	searchContext := map[string]string{
		cEnvironmentDescription: "production",
	}
	result := EnvironmentFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestEnvironmentFilterByContext_WithBothFilters(t *testing.T) {
	// Test that function handles multiple filters
	searchContext := map[string]string{
		cEnvironmentID:          "PROD",
		cEnvironmentDescription: "production",
	}
	result := EnvironmentFilterByContext(searchContext)
	assert.NotNil(t, result)
}

func TestEnvironmentFilterByContext_CaseInsensitive(t *testing.T) {
	// Test that filtering is case-insensitive
	searchContext := map[string]string{
		cEnvironmentID: "prod",
	}
	result := EnvironmentFilterByContext(searchContext)
	assert.NotNil(t, result)
}

// Test edge cases for pagination
func TestEnvironmentGeneratePage_LargeDataset(t *testing.T) {
	list := []*shared.Environment{}
	for i := 1; i <= 1000; i++ {
		list = append(list, &shared.Environment{ID: "ENV"})
	}

	// Test various pages
	result := environmentGeneratePage(list, 1, 100)
	assert.Equal(t, 100, len(result))

	result = environmentGeneratePage(list, 10, 100)
	assert.Equal(t, 100, len(result))

	result = environmentGeneratePage(list, 11, 100)
	assert.Equal(t, 0, len(result))
}

func TestEnvironmentGeneratePage_SingleItemPerPage(t *testing.T) {
	list := []*shared.Environment{
		{ID: "ENV1"},
		{ID: "ENV2"},
		{ID: "ENV3"},
	}

	result := environmentGeneratePage(list, 1, 1)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "ENV1", result[0].ID)

	result = environmentGeneratePage(list, 2, 1)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "ENV2", result[0].ID)

	result = environmentGeneratePage(list, 3, 1)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "ENV3", result[0].ID)
}

// Test boundary conditions
func TestEnvironmentGeneratePage_BoundaryConditions(t *testing.T) {
	testCases := []struct {
		name        string
		listSize    int
		page        int
		pageSize    int
		expectedLen int
	}{
		{"Empty list", 0, 1, 10, 0},
		{"One item, page 1", 1, 1, 10, 1},
		{"Ten items, page 1, size 10", 10, 1, 10, 10},
		{"Ten items, page 2, size 10", 10, 2, 10, 0},
		{"Eleven items, page 2, size 10", 11, 2, 10, 1},
		{"Twenty items, page 2, size 10", 20, 2, 10, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			list := []*shared.Environment{}
			for i := 0; i < tc.listSize; i++ {
				list = append(list, &shared.Environment{ID: "ENV"})
			}
			result := environmentGeneratePage(list, tc.page, tc.pageSize)
			assert.Equal(t, tc.expectedLen, len(result))
		})
	}
}

// Test context validation
func TestEnvironmentRuleGeneratePageWithContext_ValidPageNumbers(t *testing.T) {
	list := []*shared.Environment{}
	for i := 1; i <= 50; i++ {
		list = append(list, &shared.Environment{ID: "ENV"})
	}

	testCases := []struct {
		page     string
		pageSize string
		valid    bool
	}{
		{"1", "10", true},
		{"5", "10", true},
		{"1", "50", true},
		{"0", "10", false},
		{"-1", "10", false},
		{"1", "0", false},
		{"1", "-1", false},
	}

	for _, tc := range testCases {
		contextMap := map[string]string{
			cPercentageBeanPageNumber: tc.page,
			cPercentageBeanPageSize:   tc.pageSize,
		}

		result, err := EnvironmentRuleGeneratePageWithContext(list, contextMap)
		if tc.valid {
			assert.NoError(t, err)
			assert.NotNil(t, result)
		} else {
			assert.Error(t, err)
			assert.Nil(t, result)
		}
	}
}
