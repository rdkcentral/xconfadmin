/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
 * SPDX-License-Identifier: Apache-2.0
 */
package telemetry

import (
	"testing"

	"github.com/rdkcentral/xconfadmin/adminapi/dcm/mocks"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

// mockDaoInstance holds the global mock DAO for testing
var mockDaoInstance *mocks.MockCachedSimpleDao

// useMockDatabase determines if we're using mock or real database
var useMockDatabase = false

// originalGetCachedSimpleDaoFunc stores the original function to restore later
var originalGetCachedSimpleDaoFunc func() db.CachedSimpleDao

// InitMockDatabase initializes the mock database for testing
// Call this in TestMain to enable mock mode for <15s test execution
// This GLOBALLY replaces the DAO so all service calls use the mock!
func InitMockDatabase() *mocks.MockCachedSimpleDao {
	mockDaoInstance = mocks.NewMockCachedSimpleDao()
	useMockDatabase = true

	// CRITICAL: Override the global GetCachedSimpleDaoFunc so ALL code uses our mock
	// This includes handlers, services, and shared/logupload functions
	originalGetCachedSimpleDaoFunc = xwlogupload.GetCachedSimpleDaoFunc
	xwlogupload.GetCachedSimpleDaoFunc = func() db.CachedSimpleDao {
		return mockDaoInstance
	}

	return mockDaoInstance
}

// RestoreRealDatabase restores the real DAO (call in cleanup/teardown)
func RestoreRealDatabase() {
	if originalGetCachedSimpleDaoFunc != nil {
		xwlogupload.GetCachedSimpleDaoFunc = originalGetCachedSimpleDaoFunc
	}
	useMockDatabase = false
	mockDaoInstance = nil
}

// GetMockDaoForTesting returns the mock DAO instance for test assertions
func GetMockDaoForTesting() *mocks.MockCachedSimpleDao {
	return mockDaoInstance
}

// ClearMockDatabase clears all mock data - ultra fast cleanup
func ClearMockDatabase() {
	if useMockDatabase && mockDaoInstance != nil {
		mockDaoInstance.Clear()
	}
}

// DisableMockDatabase disables mock mode (for real integration tests)
func DisableMockDatabase() {
	RestoreRealDatabase()
}

// IsMockDatabaseEnabled returns true if mock database is enabled
func IsMockDatabaseEnabled() bool {
	return useMockDatabase
}

// Helper functions to abstract DAO operations for mock/real database

// GetOneFromDao retrieves a single entity - works with both mock and real DAO
func GetOneFromDao(tableName string, rowKey string) (interface{}, error) {
	if useMockDatabase && mockDaoInstance != nil {
		return mockDaoInstance.GetOne(tableName, rowKey)
	}
	return db.GetCachedSimpleDao().GetOne(tableName, rowKey)
}

// SetOneInDao stores a single entity - works with both mock and real DAO
func SetOneInDao(tableName string, rowKey string, entity interface{}) error {
	if useMockDatabase && mockDaoInstance != nil {
		return mockDaoInstance.SetOne(tableName, rowKey, entity)
	}
	return db.GetCachedSimpleDao().SetOne(tableName, rowKey, entity)
}

// DeleteOneFromDao removes a single entity - works with both mock and real DAO
func DeleteOneFromDao(tableName string, rowKey string) error {
	if useMockDatabase && mockDaoInstance != nil {
		return mockDaoInstance.DeleteOne(tableName, rowKey)
	}
	return db.GetCachedSimpleDao().DeleteOne(tableName, rowKey)
}

// GetAllAsListFromDao retrieves all entities as a list - works with both mock and real DAO
func GetAllAsListFromDao(tableName string, maxResults int) ([]interface{}, error) {
	if useMockDatabase && mockDaoInstance != nil {
		return mockDaoInstance.GetAllAsList(tableName, maxResults)
	}
	return db.GetCachedSimpleDao().GetAllAsList(tableName, maxResults)
}

// GetAllAsMapFromDao retrieves all entities as a map - works with both mock and real DAO
func GetAllAsMapFromDao(tableName string) (map[interface{}]interface{}, error) {
	if useMockDatabase && mockDaoInstance != nil {
		return mockDaoInstance.GetAllAsMap(tableName)
	}
	return db.GetCachedSimpleDao().GetAllAsMap(tableName)
}

// RefreshAllInDao refreshes cache for a table - no-op for mock
func RefreshAllInDao(tableName string) error {
	if useMockDatabase && mockDaoInstance != nil {
		return mockDaoInstance.RefreshAll(tableName)
	}
	return db.GetCachedSimpleDao().RefreshAll(tableName)
}

// SkipIfMockDatabase marks integration tests to skip in mock mode
// Use this for integration tests that require real database operations
func SkipIfMockDatabase(t *testing.T) {
	if useMockDatabase {
		t.Skip("Skipping integration test in mock mode (requires real database)")
	}
}
