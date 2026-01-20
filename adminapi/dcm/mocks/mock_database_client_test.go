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

// TestNewMockDatabaseClient tests creating a new mock database client
func TestNewMockDatabaseClient(t *testing.T) {
	client := NewMockDatabaseClient()
	assert.NotNil(t, client)
}

// TestMockDatabaseClient_AcquireLock tests acquiring lock
func TestMockDatabaseClient_AcquireLock(t *testing.T) {
	client := NewMockDatabaseClient()
	err := client.AcquireLock("test-lock", "owner-123", 60)
	assert.Nil(t, err)
}

// TestMockDatabaseClient_ReleaseLock tests releasing lock
func TestMockDatabaseClient_ReleaseLock(t *testing.T) {
	client := NewMockDatabaseClient()
	err := client.ReleaseLock("test-cf", "test-key")
	assert.Nil(t, err)
}

// TestMockDatabaseClient_Close tests closing client
func TestMockDatabaseClient_Close(t *testing.T) {
	client := NewMockDatabaseClient()
	client.Close()
	// No error expected, just verify no panic
	assert.True(t, true)
}
