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
	"sync"
	"time"

	"github.com/rdkcentral/xconfwebconfig/db"
)

// MockDatabaseClient is a minimal implementation of db.DatabaseClient interface
// This is only used to prevent panics when distributed locks are created in test mode
// It implements lock-related methods and basic data storage for testing
type MockDatabaseClient struct {
	locks map[string]string            // lockName -> lockedBy
	data  map[string]map[string][]byte // tableName -> rowKey -> data
	mu    sync.Mutex
}

// NewMockDatabaseClient creates a new mock database client
func NewMockDatabaseClient() *MockDatabaseClient {
	return &MockDatabaseClient{
		locks: make(map[string]string),
		data:  make(map[string]map[string][]byte),
	}
}

// AcquireLock implements the lock acquisition (no-op for tests)
func (m *MockDatabaseClient) AcquireLock(lockName string, lockedBy string, ttlSeconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.locks[lockName] = lockedBy
	return nil
}

// ReleaseLock implements the lock release (no-op for tests)
func (m *MockDatabaseClient) ReleaseLock(lockName string, lockedBy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.locks, lockName)
	return nil
}

// All other methods are stubbed out with minimal implementations

func (m *MockDatabaseClient) SetUp() error    { return nil }
func (m *MockDatabaseClient) TearDown() error { return nil }
func (m *MockDatabaseClient) Close() error    { return nil }
func (m *MockDatabaseClient) Sleep()          {}
func (m *MockDatabaseClient) SetXconfData(tableName string, rowKey string, value []byte, ttl int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data[tableName] == nil {
		m.data[tableName] = make(map[string][]byte)
	}
	m.data[tableName][rowKey] = value
	return nil
}
func (m *MockDatabaseClient) GetXconfData(tableName string, rowKey string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data[tableName] == nil {
		return nil, nil
	}
	return m.data[tableName][rowKey], nil
}
func (m *MockDatabaseClient) GetAllXconfDataByKeys(tableName string, rowKeys []string) [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data[tableName] == nil {
		return nil
	}
	var result [][]byte
	for _, key := range rowKeys {
		if data, ok := m.data[tableName][key]; ok {
			result = append(result, data)
		}
	}
	return result
}
func (m *MockDatabaseClient) GetAllXconfKeys(tableName string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data[tableName] == nil {
		return nil
	}
	var keys []string
	for key := range m.data[tableName] {
		keys = append(keys, key)
	}
	return keys
}
func (m *MockDatabaseClient) GetAllXconfDataAsList(tableName string, maxResults int) [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data[tableName] == nil {
		return nil
	}
	var result [][]byte
	count := 0
	for _, data := range m.data[tableName] {
		result = append(result, data)
		count++
		if maxResults > 0 && count >= maxResults {
			break
		}
	}
	return result
}
func (m *MockDatabaseClient) GetAllXconfDataAsMap(tableName string, maxResults int) map[string][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data[tableName] == nil {
		return nil
	}
	result := make(map[string][]byte)
	count := 0
	for key, data := range m.data[tableName] {
		result[key] = data
		count++
		if maxResults > 0 && count >= maxResults {
			break
		}
	}
	return result
}
func (m *MockDatabaseClient) DeleteXconfData(tableName string, rowKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data[tableName] != nil {
		delete(m.data[tableName], rowKey)
	}
	return nil
}
func (m *MockDatabaseClient) DeleteAllXconfData(tableName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, tableName)
	return nil
}
func (m *MockDatabaseClient) GetAllXconfData(tableName string, rowKey string) [][]byte {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfDataTwoKeysRange(tableName string, rowKey interface{}, key2FieldName string, rangeInfo *db.RangeInfo) [][]byte {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfDataTwoKeysAsMap(tableName string, rowKey string, key2FieldName string, key2List []interface{}) map[interface{}][]byte {
	return nil
}
func (m *MockDatabaseClient) SetXconfDataTwoKeys(tableName string, rowKey interface{}, key2FieldName string, key2 interface{}, value []byte, ttl int) error {
	return nil
}
func (m *MockDatabaseClient) GetXconfDataTwoKeys(tableName string, rowKey string, key2FieldName string, key2 interface{}) ([]byte, error) {
	return nil, nil
}
func (m *MockDatabaseClient) DeleteXconfDataTwoKeys(tableName string, rowKey string, key2FieldName string, key2 interface{}) error {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfTwoKeys(tableName string, key2FieldName string) []db.TwoKeys {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfKey2s(tableName string, rowKey string, key2FieldName string) []interface{} {
	return nil
}
func (m *MockDatabaseClient) SetXconfCompressedData(tableName string, rowKey string, values [][]byte, ttl int) error {
	return nil
}
func (m *MockDatabaseClient) GetXconfCompressedData(tableName string, rowKey string) ([]byte, error) {
	return nil, nil
}
func (m *MockDatabaseClient) GetAllXconfCompressedDataAsMap(tableName string) map[string][]byte {
	return nil
}
func (m *MockDatabaseClient) GetEcmMacFromPodTable(s string) (string, error) {
	return "", nil
}
func (m *MockDatabaseClient) IsDbNotFound(error) bool {
	return false
}
func (m *MockDatabaseClient) GetPenetrationMetrics(macAddress string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseClient) SetPenetrationMetrics(penetrationmetrics *db.PenetrationMetrics) error {
	return nil
}
func (m *MockDatabaseClient) SetFwPenetrationMetrics(metrics *db.FwPenetrationMetrics) error {
	return nil
}
func (m *MockDatabaseClient) GetFwPenetrationMetrics(s string) (*db.FwPenetrationMetrics, error) {
	return nil, nil
}
func (m *MockDatabaseClient) SetRfcPenetrationMetrics(pMetrics *db.RfcPenetrationMetrics, is304FromPrecook bool) error {
	return nil
}
func (m *MockDatabaseClient) GetRfcPenetrationMetrics(s string) (*db.RfcPenetrationMetrics, error) {
	return nil, nil
}
func (m *MockDatabaseClient) UpdateFwPenetrationMetrics(m2 map[string]string) error {
	return nil
}
func (m *MockDatabaseClient) GetEstbIp(s string) (string, error) {
	return "", nil
}
func (m *MockDatabaseClient) SetRecookingStatus(module string, partitionId string, state int) error {
	return nil
}
func (m *MockDatabaseClient) GetRecookingStatus(module string, partitionId string) (int, time.Time, error) {
	return 0, time.Time{}, nil
}
func (m *MockDatabaseClient) CheckFinalRecookingStatus(module string) (bool, time.Time, error) {
	return false, time.Time{}, nil
}
func (m *MockDatabaseClient) SetPrecookDataInXPC(RfcPrecookHash string, RfcPrecookPayload []byte) error {
	return nil
}
func (m *MockDatabaseClient) GetPrecookDataFromXPC(RfcPrecookHash string) ([]byte, string, error) {
	return nil, "", nil
}
func (m *MockDatabaseClient) GetLockInfo(lockName string) (map[string]interface{}, error) {
	return nil, nil
}

// ExecuteBatch executes a batch of operations (stub for tests)
func (m *MockDatabaseClient) ExecuteBatch(operation db.BatchOperation) error {
	return nil
}

// ModifyXconfData modifies existing data (stub for tests)
func (m *MockDatabaseClient) ModifyXconfData(tableName string, rowKeys ...string) error {
	return nil
}

// NewBatch creates a new batch operation (stub for tests)
func (m *MockDatabaseClient) NewBatch(size int) db.BatchOperation {
	return nil
}

// QueryXconfDataRows queries data rows (stub for tests)
func (m *MockDatabaseClient) QueryXconfDataRows(tableName string, rowKeys ...string) ([]map[string]interface{}, error) {
	return nil, nil
}

// Clear removes all stored data (useful for test cleanup)
func (m *MockDatabaseClient) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]map[string][]byte)
	m.locks = make(map[string]string)
}
