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
	"errors"
	"sync"
)

// MockCachedSimpleDao is an in-memory mock implementation of the CachedSimpleDao interface
// This mock stores data in memory for fast unit testing without requiring a real database
type MockCachedSimpleDao struct {
	data map[string]map[string]interface{} // tableName -> rowKey -> entity
	mu   sync.RWMutex
}

// NewMockCachedSimpleDao creates a new instance of the mock DAO
func NewMockCachedSimpleDao() *MockCachedSimpleDao {
	return &MockCachedSimpleDao{
		data: make(map[string]map[string]interface{}),
	}
}

// GetOne retrieves a single entity by table name and row key
func (m *MockCachedSimpleDao) GetOne(tableName string, rowKey string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data[tableName] == nil {
		return nil, errors.New("not found")
	}

	entity, ok := m.data[tableName][rowKey]
	if !ok {
		return nil, errors.New("not found")
	}

	return entity, nil
}

// GetOneFromCacheOnly retrieves a single entity from cache (same as GetOne in mock)
func (m *MockCachedSimpleDao) GetOneFromCacheOnly(tableName string, rowKey string) (interface{}, error) {
	return m.GetOne(tableName, rowKey)
}

// SetOne stores a single entity in the specified table with the given row key
func (m *MockCachedSimpleDao) SetOne(tableName string, rowKey string, entity interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data[tableName] == nil {
		m.data[tableName] = make(map[string]interface{})
	}

	m.data[tableName][rowKey] = entity
	return nil
}

// DeleteOne removes a single entity from the specified table
func (m *MockCachedSimpleDao) DeleteOne(tableName string, rowKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data[tableName] != nil {
		delete(m.data[tableName], rowKey)
	}

	return nil
}

// GetAllByKeys retrieves multiple entities by their keys from a table
func (m *MockCachedSimpleDao) GetAllByKeys(tableName string, rowKeys []string) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []interface{}

	if m.data[tableName] == nil {
		return result, nil
	}

	for _, key := range rowKeys {
		if entity, ok := m.data[tableName][key]; ok {
			result = append(result, entity)
		}
	}

	return result, nil
}

// GetAllAsList retrieves all entities from a table as a list
func (m *MockCachedSimpleDao) GetAllAsList(tableName string, maxResults int) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []interface{}

	if m.data[tableName] == nil {
		return result, nil
	}

	count := 0
	for _, entity := range m.data[tableName] {
		result = append(result, entity)
		count++
		if maxResults > 0 && count >= maxResults {
			break
		}
	}

	return result, nil
}

// GetAllAsMap retrieves all entities from a table as a map
func (m *MockCachedSimpleDao) GetAllAsMap(tableName string) (map[interface{}]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[interface{}]interface{})

	if m.data[tableName] == nil {
		return result, nil
	}

	for key, entity := range m.data[tableName] {
		result[key] = entity
	}

	return result, nil
}

// GetAllAsShallowMap retrieves all entities from a table as a shallow map (same as GetAllAsMap in mock)
func (m *MockCachedSimpleDao) GetAllAsShallowMap(tableName string) (map[interface{}]interface{}, error) {
	return m.GetAllAsMap(tableName)
}

// GetKeys retrieves all keys from a table
func (m *MockCachedSimpleDao) GetKeys(tableName string) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []interface{}

	if m.data[tableName] == nil {
		return result, nil
	}

	for key := range m.data[tableName] {
		result = append(result, key)
	}

	return result, nil
}

// RefreshAll refreshes all cached data for a table (no-op in mock)
func (m *MockCachedSimpleDao) RefreshAll(tableName string) error {
	// No-op for in-memory mock - data is always "fresh"
	return nil
}

// RefreshOne refreshes cached data for a single entity (no-op in mock)
func (m *MockCachedSimpleDao) RefreshOne(tableName string, rowKey string) error {
	// No-op for in-memory mock - data is always "fresh"
	return nil
}

// Clear removes all data from all tables (useful for test cleanup)
func (m *MockCachedSimpleDao) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]map[string]interface{})
}

// ClearTable removes all data from a specific table
func (m *MockCachedSimpleDao) ClearTable(tableName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, tableName)
}

// GetTableData returns a copy of all data in a table (for testing/debugging)
func (m *MockCachedSimpleDao) GetTableData(tableName string) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data[tableName] == nil {
		return make(map[string]interface{})
	}

	// Return a copy to prevent external modifications
	result := make(map[string]interface{})
	for k, v := range m.data[tableName] {
		result[k] = v
	}

	return result
}

// CountEntries returns the number of entries in a table
func (m *MockCachedSimpleDao) CountEntries(tableName string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data[tableName] == nil {
		return 0
	}

	return len(m.data[tableName])
}
