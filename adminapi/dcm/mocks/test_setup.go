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

// Global variable to hold the mock DAO instance for test access
var globalMockDao *MockCachedSimpleDao

// SetupMockDatabase initializes the mock database infrastructure for tests
// This should be called in TestMain before running tests
// Returns the mock DAO instance that can be used in tests
func SetupMockDatabase() *MockCachedSimpleDao {
	mockDao := NewMockCachedSimpleDao()
	globalMockDao = mockDao
	return mockDao
}

// GetMockDao returns the global mock DAO instance
func GetMockDao() *MockCachedSimpleDao {
	return globalMockDao
}

// CleanupMockDatabase clears all mock data
// This should be called between tests or in test cleanup
func CleanupMockDatabase() {
	if globalMockDao != nil {
		globalMockDao.Clear()
	}
}
