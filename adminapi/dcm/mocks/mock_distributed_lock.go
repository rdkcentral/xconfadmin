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
)

// MockDistributedLock is a no-op distributed lock for testing
// It satisfies the same interface as db.DistributedLock but doesn't require Cassandra
type MockDistributedLock struct {
	name   string
	ttl    int
	mu     sync.Mutex
	locked bool
	owner  string
}

// NewMockDistributedLock creates a new mock distributed lock
func NewMockDistributedLock(name string, ttl int) *MockDistributedLock {
	return &MockDistributedLock{
		name: name,
		ttl:  ttl,
	}
}

// Lock acquires the lock (no-op in mock mode, just tracks state)
func (m *MockDistributedLock) Lock(owner string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// In mock mode, just track that we're locked
	m.locked = true
	m.owner = owner
	return nil
}

// Unlock releases the lock (no-op in mock mode)
func (m *MockDistributedLock) Unlock(owner string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// In mock mode, just track that we're unlocked
	m.locked = false
	m.owner = ""
	return nil
}

// IsLocked returns whether the lock is currently held (for testing)
func (m *MockDistributedLock) IsLocked() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.locked
}

// GetOwner returns the current lock owner (for testing)
func (m *MockDistributedLock) GetOwner() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.owner
}
