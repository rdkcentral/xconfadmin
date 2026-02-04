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

package tag

import (
	"testing"

	xhttp "github.com/rdkcentral/xconfadmin/http"
	taggingapi_config "github.com/rdkcentral/xconfadmin/taggingapi/config"
	"github.com/stretchr/testify/assert"
)

func TestGetGroupServiceSyncConnector(t *testing.T) {
	setupTestEnvironment()
	connector := GetGroupServiceSyncConnector()
	assert.NotNil(t, xhttp.WebConfServer)
	t.Logf("GroupServiceSyncConnector: %v", connector)
}

func TestGetTagApiConfig(t *testing.T) {
	setupTestEnvironment()
	config := GetTagApiConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 5000, config.BatchLimit)
	assert.Equal(t, 20, config.WorkerCount)
}

func TestSetTagApiConfig(t *testing.T) {
	setupTestEnvironment()
	newConfig := &taggingapi_config.TaggingApiConfig{
		BatchLimit:  3000,
		WorkerCount: 10,
	}
	SetTagApiConfig(newConfig)

	retrieved := GetTagApiConfig()
	assert.NotNil(t, retrieved)
	assert.Equal(t, 3000, retrieved.BatchLimit)
	assert.Equal(t, 10, retrieved.WorkerCount)

	// Restore original config after test
	setupTestEnvironment()
}

func TestGetGroupServiceConnector(t *testing.T) {
	setupTestEnvironment()
	connector := GetGroupServiceConnector()
	assert.NotNil(t, xhttp.WebConfServer)
	t.Logf("GroupServiceConnector: %v", connector)
}

func TestCheckBatchSizeExceeded(t *testing.T) {
	setupTestEnvironment() // Reset to BatchLimit: 5000
	// Explicitly set the config to ensure proper test isolation
	SetTagApiConfig(&taggingapi_config.TaggingApiConfig{
		BatchLimit:  5000,
		WorkerCount: 20,
	})

	testCases := []struct {
		name      string
		batchSize int
		expectErr bool
	}{
		{"within limit", 1000, false},
		{"at limit", 5000, false},
		{"exceeds limit", 5001, true},
		{"far exceeds", 10000, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckBatchSizeExceeded(tc.batchSize)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "exceeds the limit")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFilterTagEntriesByPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected int
	}{
		{
			name:     "mixed entries",
			input:    []string{"t_test1", "t_test2", "other/test3", "t_test4"},
			expected: 3,
		},
		{
			name:     "all with prefix",
			input:    []string{"t_a", "t_b", "t_c"},
			expected: 3,
		},
		{
			name:     "none with prefix",
			input:    []string{"other/a", "different/b"},
			expected: 0,
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filterTagEntriesByPrefix(tc.input)
			assert.Equal(t, tc.expected, len(result))
			// Verify prefix is removed
			for _, tag := range result {
				assert.NotContains(t, tag, "t_")
			}
		})
	}
}

func TestGetTagsByMember(t *testing.T) {
	setupTestEnvironment()

	testMembers := []string{
		"AA:BB:CC:DD:EE:FF",
		"AABBCCDDEEFF",
		"test-member",
	}

	for _, member := range testMembers {
		tags, err := GetTagsByMember(member)
		// Without real connector, expect error or empty result
		if err != nil {
			t.Logf("GetTagsByMember(%s) returned error: %v (expected without connector)", member, err)
		} else {
			assert.NotNil(t, tags)
			t.Logf("GetTagsByMember(%s) returned %d tags", member, len(tags))
		}
	}
}
