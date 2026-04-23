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
// It implements only the lock-related methods; all other methods return nil/empty values
type MockDatabaseClient struct {
	locks map[string]string // lockName -> lockedBy
	mu    sync.Mutex
}

// NewMockDatabaseClient creates a new mock database client
func NewMockDatabaseClient() *MockDatabaseClient {
	return &MockDatabaseClient{
		locks: make(map[string]string),
	}
}

// AcquireLock implements the lock acquisition (no-op for tests)
func (m *MockDatabaseClient) AcquireLock(tenantId string, lockName string, lockedBy string, ttlSeconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.locks[lockName] = lockedBy
	return nil
}

// ReleaseLock implements the lock release (no-op for tests)
func (m *MockDatabaseClient) ReleaseLock(tenantId string, lockName string, lockedBy string) error {
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
func (m *MockDatabaseClient) SetXconfData(tenantId string, tableName string, rowKey string, value []byte, ttl int) error {
	return nil
}
func (m *MockDatabaseClient) GetXconfData(tenantId string, tableName string, rowKey string) ([]byte, error) {
	return nil, nil
}
func (m *MockDatabaseClient) GetAllXconfDataByKeys(tenantId string, tableName string, rowKeys []string) [][]byte {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfKeys(tenantId string, tableName string) []string {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfDataAsList(tenantId string, tableName string, maxResults int) [][]byte {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfDataAsMap(tenantId string, tableName string, maxResults int) map[string][]byte {
	return nil
}
func (m *MockDatabaseClient) DeleteXconfData(tenantId string, tableName string, rowKey string) error {
	return nil
}
func (m *MockDatabaseClient) DeleteAllXconfData(tenantId string, tableName string) error {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfData(tenantId string, tableName string, rowKey string) [][]byte {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfDataTwoKeysRange(tenantId string, tableName string, rowKey interface{}, key2FieldName string, rangeInfo *db.RangeInfo) [][]byte {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfDataTwoKeysAsMap(tenantId string, tableName string, rowKey string, key2FieldName string, key2List []interface{}) map[interface{}][]byte {
	return nil
}
func (m *MockDatabaseClient) SetXconfDataTwoKeys(tenantId string, tableName string, rowKey interface{}, key2FieldName string, key2 interface{}, value []byte, ttl int) error {
	return nil
}
func (m *MockDatabaseClient) GetXconfDataTwoKeys(tenantId string, tableName string, rowKey string, key2FieldName string, key2 interface{}) ([]byte, error) {
	return nil, nil
}
func (m *MockDatabaseClient) DeleteXconfDataTwoKeys(tenantId string, tableName string, rowKey string, key2FieldName string, key2 interface{}) error {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfTwoKeys(tenantId string, tableName string, key2FieldName string) []db.TwoKeys {
	return nil
}
func (m *MockDatabaseClient) GetAllXconfKey2s(tenantId string, tableName string, rowKey string, key2FieldName string) []interface{} {
	return nil
}
func (m *MockDatabaseClient) SetXconfCompressedData(tenantId string, tableName string, rowKey string, values [][]byte, ttl int) error {
	return nil
}
func (m *MockDatabaseClient) GetXconfCompressedData(tenantId string, tableName string, rowKey string) ([]byte, error) {
	return nil, nil
}
func (m *MockDatabaseClient) GetAllXconfCompressedDataAsMap(tenantId string, tableName string) map[string][]byte {
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
func (m *MockDatabaseClient) GetLockInfo(tenantId string, lockName string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseClient) GetAllTenants() []*db.Tenant {
	return nil
}
func (m *MockDatabaseClient) SetTenant(tenant *db.Tenant) error {
	return nil
}
func (m *MockDatabaseClient) DeleteTenant(tenantId string) error {
	return nil
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
