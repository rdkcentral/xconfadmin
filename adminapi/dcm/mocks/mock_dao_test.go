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
package mocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testTable = "TEST_TABLE"

// TestNewMockCachedSimpleDao tests creating a new mock DAO
func TestNewMockCachedSimpleDao(t *testing.T) {
	dao := NewMockCachedSimpleDao()
	assert.NotNil(t, dao)
	assert.NotNil(t, dao.data)
}

// TestMockCachedSimpleDao_SetOne tests setting a single entity
func TestMockCachedSimpleDao_SetOne(t *testing.T) {
	dao := NewMockCachedSimpleDao()

	err := dao.SetOne(testTable, "test-id", "test-value")
	assert.Nil(t, err)

	// Verify it was stored
	result, err := dao.GetOne(testTable, "test-id")
	assert.Nil(t, err)
	assert.Equal(t, "test-value", result)
}

// TestMockCachedSimpleDao_GetOne tests getting a single entity
func TestMockCachedSimpleDao_GetOne(t *testing.T) {
	dao := NewMockCachedSimpleDao()

	// Test non-existent key
	_, err := dao.GetOne(testTable, "non-existent")
	assert.NotNil(t, err)

	// Test existing key
	dao.SetOne(testTable, "key1", "value1")
	result, err := dao.GetOne(testTable, "key1")
	assert.Nil(t, err)
	assert.Equal(t, "value1", result)
}

// TestMockCachedSimpleDao_DeleteOne tests deleting a single entity
func TestMockCachedSimpleDao_DeleteOne(t *testing.T) {
	dao := NewMockCachedSimpleDao()

	// Add data
	dao.SetOne(testTable, "key1", "value1")
	dao.SetOne(testTable, "key2", "value2")

	// Delete one
	err := dao.DeleteOne(testTable, "key1")
	assert.Nil(t, err)

	// Verify it was deleted
	_, err = dao.GetOne(testTable, "key1")
	assert.NotNil(t, err)

	// Verify other key still exists
	result, err := dao.GetOne(testTable, "key2")
	assert.Nil(t, err)
	assert.Equal(t, "value2", result)
}

// TestMockCachedSimpleDao_GetKeys tests getting all keys
func TestMockCachedSimpleDao_GetKeys(t *testing.T) {
	dao := NewMockCachedSimpleDao()

	// Test empty dao - should return empty slice, not nil
	keys, err := dao.GetKeys(testTable)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(keys))

	// Add some data
	dao.SetOne(testTable, "key1", "value1")
	dao.SetOne(testTable, "key2", "value2")

	keys, err = dao.GetKeys(testTable)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(keys))
}

// TestMockCachedSimpleDao_GetAllAsMap tests getting all entities as map
func TestMockCachedSimpleDao_GetAllAsMap(t *testing.T) {
	dao := NewMockCachedSimpleDao()

	// Test empty dao
	resultMap, err := dao.GetAllAsMap(testTable)
	assert.Nil(t, err)
	assert.NotNil(t, resultMap)
	assert.Equal(t, 0, len(resultMap))

	// Add some data
	dao.SetOne(testTable, "key1", "value1")
	dao.SetOne(testTable, "key2", "value2")

	resultMap, err = dao.GetAllAsMap(testTable)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resultMap))
	assert.Equal(t, "value1", resultMap["key1"])
	assert.Equal(t, "value2", resultMap["key2"])
}

// TestMockCachedSimpleDao_Clear tests clearing all data
func TestMockCachedSimpleDao_Clear(t *testing.T) {
	dao := NewMockCachedSimpleDao()

	// Add data to multiple tables
	dao.SetOne(testTable, "key1", "value1")
	dao.SetOne(testTable, "key2", "value2")
	dao.SetOne("OTHER_TABLE", "key3", "value3")

	// Clear
	dao.Clear()

	// Verify all data is gone
	_, err := dao.GetOne(testTable, "key1")
	assert.NotNil(t, err)
	_, err = dao.GetOne("OTHER_TABLE", "key3")
	assert.NotNil(t, err)
}

// TestMockCachedSimpleDao_GetOneFromCacheOnly tests cache-only retrieval
func TestMockCachedSimpleDao_GetOneFromCacheOnly(t *testing.T) {
	dao := NewMockCachedSimpleDao()

	// Add data
	dao.SetOne(testTable, "key1", "value1")

	// Get from cache only (same as GetOne in mock)
	result, err := dao.GetOneFromCacheOnly(testTable, "key1")
	assert.Nil(t, err)
	assert.Equal(t, "value1", result)
}
