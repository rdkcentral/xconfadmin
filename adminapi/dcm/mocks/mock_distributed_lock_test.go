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

// TestNewMockDistributedLock tests creating a new mock distributed lock
func TestNewMockDistributedLock(t *testing.T) {
	lock := NewMockDistributedLock("test-key", 60)
	assert.NotNil(t, lock)
	assert.Equal(t, "test-key", lock.name)
	assert.False(t, lock.locked)
}

// TestMockDistributedLock_Lock tests locking
func TestMockDistributedLock_Lock(t *testing.T) {
	lock := NewMockDistributedLock("test-key", 60)

	err := lock.Lock("owner-123")
	assert.Nil(t, err)
	assert.True(t, lock.locked)
	assert.Equal(t, "owner-123", lock.owner)
}

// TestMockDistributedLock_Unlock tests unlocking
func TestMockDistributedLock_Unlock(t *testing.T) {
	lock := NewMockDistributedLock("test-key", 60)

	// Lock first
	lock.Lock("owner-123")
	assert.True(t, lock.locked)

	// Then unlock
	err := lock.Unlock("owner-123")
	assert.Nil(t, err)
	assert.False(t, lock.locked)
}

// TestMockDistributedLock_IsLocked tests checking lock status
func TestMockDistributedLock_IsLocked(t *testing.T) {
	lock := NewMockDistributedLock("test-key", 60)

	// Initially not locked
	assert.False(t, lock.IsLocked())

	// After locking
	lock.Lock("owner-123")
	assert.True(t, lock.IsLocked())

	// After unlocking
	lock.Unlock("owner-123")
	assert.False(t, lock.IsLocked())
}

// TestMockDistributedLock_MultipleLockUnlock tests multiple lock/unlock cycles
func TestMockDistributedLock_MultipleLockUnlock(t *testing.T) {
	lock := NewMockDistributedLock("test-key", 60)

	for i := 0; i < 5; i++ {
		err := lock.Lock("owner")
		assert.Nil(t, err)
		assert.True(t, lock.IsLocked())

		err = lock.Unlock("owner")
		assert.Nil(t, err)
		assert.False(t, lock.IsLocked())
	}
}
