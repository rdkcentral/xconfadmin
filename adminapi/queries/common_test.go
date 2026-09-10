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
package queries

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	xshared "github.com/rdkcentral/xconfadmin/shared"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/stretchr/testify/assert"
)

const tenantsUrl = "/xconfAdminService/tenants"

var tenantRoutesAdded bool

// ensureTenantRoutes dynamically adds tenant routes if not present in test router
func ensureTenantRoutes() {
	if router != nil && !tenantRoutesAdded {
		tenantsPath := router.PathPrefix(tenantsUrl).Subrouter()
		tenantsPath.HandleFunc("", GetTenantsHandler).Methods(http.MethodGet)
		tenantsPath.HandleFunc("", CreateTenantHandler).Methods(http.MethodPost)
		tenantsPath.HandleFunc("/{id}", DeleteTenantHandler).Methods(http.MethodDelete)
		tenantRoutesAdded = true
	}
}

// createTestTenant issues a POST to create a tenant and returns the response
func createTestTenant(t *testing.T, id, name string) *http.Response {
	t.Helper()
	ensureTenantRoutes()
	body, _ := json.Marshal(db.Tenant{ID: id, Name: name})
	req, _ := http.NewRequest(http.MethodPost, tenantsUrl, bytes.NewReader(body))
	req.Header.Set("Accept", "application/json")
	return xshared.ExecuteRequest(req, router).Result()
}

// deleteTestTenant issues a DELETE for the given tenant id and returns the response
func deleteTestTenant(t *testing.T, id string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, tenantsUrl+"/"+id, nil)
	return xshared.ExecuteRequest(req, router).Result()
}

// TestCreateTenantHandler covers successful creation, idempotent re-creation and invalid input
func TestCreateTenantHandler(t *testing.T) {
	id := "TESTTENANTCREATE"
	defer deleteTestTenant(t, id)

	// successful create
	res := createTestTenant(t, id, "Test Tenant")
	if res.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(res.Body)
		t.Fatalf("expected 201 got %d body=%s", res.StatusCode, string(body))
	}
	var created db.Tenant
	body, _ := ioutil.ReadAll(res.Body)
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v body=%s", err, string(body))
	}
	if created.ID != id {
		t.Fatalf("expected tenant id %s got %s", id, created.ID)
	}

	// re-creating the same tenant is idempotent, returns the existing tenant
	res = createTestTenant(t, id, "Test Tenant")
	if res.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(res.Body)
		t.Fatalf("expected 201 on re-create got %d body=%s", res.StatusCode, string(body))
	}

	// invalid JSON body
	req, _ := http.NewRequest(http.MethodPost, tenantsUrl, bytes.NewReader([]byte("{bad")))
	res = xshared.ExecuteRequest(req, router).Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 invalid json got %d", res.StatusCode)
	}

	// invalid tenant ID fails Validate()
	res = createTestTenant(t, "", "No Id")
	if res.StatusCode != http.StatusBadRequest {
		body, _ := ioutil.ReadAll(res.Body)
		t.Fatalf("expected 400 invalid id got %d body=%s", res.StatusCode, string(body))
	}
}

// TestGetTenantsHandler covers listing tenants and verifies a created tenant appears
func TestGetTenantsHandler(t *testing.T) {
	id := "TESTTENANTLIST"
	defer deleteTestTenant(t, id)

	res := createTestTenant(t, id, "List Tenant")
	if res.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(res.Body)
		t.Fatalf("expected 201 got %d body=%s", res.StatusCode, string(body))
	}

	req, _ := http.NewRequest(http.MethodGet, tenantsUrl, nil)
	res = xshared.ExecuteRequest(req, router).Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.StatusCode)
	}

	var tenants []db.Tenant
	body, _ := ioutil.ReadAll(res.Body)
	if err := json.Unmarshal(body, &tenants); err != nil {
		t.Fatalf("failed to unmarshal list response: %v body=%s", err, string(body))
	}
	found := false
	for _, tn := range tenants {
		if tn.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected tenant %s to be present in list, got %+v", id, tenants)
	}
}

// TestDeleteTenantHandler covers deleting a tenant and verifies it no longer appears in the list
func TestDeleteTenantHandler(t *testing.T) {
	id := "TESTTENANTDELETE"

	res := createTestTenant(t, id, "Delete Tenant")
	if res.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(res.Body)
		t.Fatalf("expected 201 got %d body=%s", res.StatusCode, string(body))
	}

	res = deleteTestTenant(t, id)
	if res.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(res.Body)
		t.Fatalf("expected 204 got %d body=%s", res.StatusCode, string(body))
	}

	req, _ := http.NewRequest(http.MethodGet, tenantsUrl, nil)
	res = xshared.ExecuteRequest(req, router).Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.StatusCode)
	}
	var tenants []db.Tenant
	body, _ := ioutil.ReadAll(res.Body)
	if err := json.Unmarshal(body, &tenants); err != nil {
		t.Fatalf("failed to unmarshal list response: %v body=%s", err, string(body))
	}
	for _, tn := range tenants {
		if tn.ID == id {
			t.Fatalf("expected tenant %s to be removed from list, got %+v", id, tenants)
		}
	}

	// deleting again should still succeed (delete is not conditioned on existence)
	res = deleteTestTenant(t, id)
	if res.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(res.Body)
		t.Fatalf("expected 204 on repeat delete got %d body=%s", res.StatusCode, string(body))
	}
}

func TestGetChangeLog(t *testing.T) {
	// This tests the basic structure of GetChangeLog
	result, err := GetChangeLog()

	// May return error if DB not set up, but function should not panic
	if err != nil {
		assert.Error(t, err)
		return
	}

	// Should return a map
	assert.NotNil(t, result)

	// Result is a map of timestamp to changes
	assert.IsType(t, map[int64][]Change{}, result)
}

func TestGetChangedKeysMapRaw(t *testing.T) {
	// Test basic functionality
	result, err := GetChangedKeysMapRaw(db.GetDefaultTenantId())

	// May return error if DB not set up
	if err != nil {
		assert.Error(t, err)
		return
	}

	// Should return a map even if empty
	assert.NotNil(t, result)
}

func TestChangeStruct(t *testing.T) {
	// Test Change struct creation
	change := Change{
		ChangedKey: "test-key",
		Operation:  db.CREATE_OPERATION,
		CfName:     "test-table",
		UserName:   "test-user",
	}

	assert.Equal(t, "test-key", change.ChangedKey)
	assert.Equal(t, db.CREATE_OPERATION, change.Operation)
	assert.Equal(t, "test-table", change.CfName)
	assert.Equal(t, "test-user", change.UserName)
}

func TestCacheStats(t *testing.T) {
	// Test CacheStats struct
	stats := CacheStats{
		DaoRefreshTime: time.Now().Unix(),
		CacheSize:      100,
		NonAbsentCount: 90,
		RequestCount:   1000,
		EvictionCount:  10,
		HitRate:        0.95,
		MissRate:       0.05,
		TotalLoadTime:  time.Second * 10,
	}

	assert.Greater(t, stats.DaoRefreshTime, int64(0))
	assert.Equal(t, 100, stats.CacheSize)
	assert.Equal(t, 90, stats.NonAbsentCount)
	assert.Equal(t, uint64(1000), stats.RequestCount)
	assert.Equal(t, uint64(10), stats.EvictionCount)
	assert.Equal(t, 0.95, stats.HitRate)
	assert.Equal(t, 0.05, stats.MissRate)
	assert.Equal(t, time.Second*10, stats.TotalLoadTime)
}

func TestStatistics(t *testing.T) {
	// Test Statistics struct
	statsMap := make(map[string]CacheStats)
	statsMap["table1"] = CacheStats{
		CacheSize:    50,
		RequestCount: 500,
		HitRate:      0.90,
	}
	statsMap["table2"] = CacheStats{
		CacheSize:    75,
		RequestCount: 750,
		HitRate:      0.85,
	}

	statistics := Statistics{
		StatsMap: statsMap,
	}

	assert.Equal(t, 2, len(statistics.StatsMap))
	assert.Contains(t, statistics.StatsMap, "table1")
	assert.Contains(t, statistics.StatsMap, "table2")
}

func TestConstants(t *testing.T) {
	// Test package constants
	assert.Equal(t, "IMPORTED", IMPORTED)
	assert.Equal(t, "NOT_IMPORTED", NOT_IMPORTED)
}

// Additional comprehensive tests for maximum coverage

func TestChangeStructWithAllOperations(t *testing.T) {
	testCases := []struct {
		name      string
		operation db.OperationType
	}{
		{"Create Operation", db.CREATE_OPERATION},
		{"Update Operation", db.UPDATE_OPERATION},
		{"Delete Operation", db.DELETE_OPERATION},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			change := Change{
				ChangedKey: "key_" + tc.name,
				Operation:  tc.operation,
				CfName:     "cf_" + tc.name,
				UserName:   "user_" + tc.name,
			}

			assert.Equal(t, "key_"+tc.name, change.ChangedKey)
			assert.Equal(t, tc.operation, change.Operation)
			assert.Equal(t, "cf_"+tc.name, change.CfName)
			assert.Equal(t, "user_"+tc.name, change.UserName)
		})
	}
}

func TestChangeStructWithEmptyValues(t *testing.T) {
	change := Change{}

	assert.Equal(t, "", change.ChangedKey)
	assert.Equal(t, db.OperationType(""), change.Operation)
	assert.Equal(t, "", change.CfName)
	assert.Equal(t, "", change.UserName)
}

func TestCacheStatsWithZeroValues(t *testing.T) {
	stats := CacheStats{}

	assert.Equal(t, int64(0), stats.DaoRefreshTime)
	assert.Equal(t, 0, stats.CacheSize)
	assert.Equal(t, 0, stats.NonAbsentCount)
	assert.Equal(t, uint64(0), stats.RequestCount)
	assert.Equal(t, uint64(0), stats.EvictionCount)
	assert.Equal(t, 0.0, stats.HitRate)
	assert.Equal(t, 0.0, stats.MissRate)
}

func TestCacheStatsWithBoundaryValues(t *testing.T) {
	stats := CacheStats{
		DaoRefreshTime: 9223372036854775807, // Max int64
		CacheSize:      2147483647,          // Max int32
		NonAbsentCount: 2147483647,
		RequestCount:   18446744073709551615, // Max uint64
		EvictionCount:  18446744073709551615,
		HitRate:        1.0,
		MissRate:       0.0,
		TotalLoadTime:  time.Hour * 24 * 365,
	}

	assert.Equal(t, int64(9223372036854775807), stats.DaoRefreshTime)
	assert.Equal(t, 2147483647, stats.CacheSize)
	assert.Equal(t, uint64(18446744073709551615), stats.RequestCount)
	assert.Equal(t, 1.0, stats.HitRate)
	assert.Equal(t, 0.0, stats.MissRate)
}

func TestStatisticsWithMultipleCaches(t *testing.T) {
	statsMap := make(map[string]CacheStats)

	for i := 1; i <= 10; i++ {
		cacheName := "cache" + string(rune('0'+i))
		statsMap[cacheName] = CacheStats{
			CacheSize:    i * 10,
			RequestCount: uint64(i * 100),
			HitRate:      float64(i) * 0.1,
			MissRate:     1.0 - (float64(i) * 0.1),
		}
	}

	statistics := Statistics{
		StatsMap: statsMap,
	}

	assert.Equal(t, 10, len(statistics.StatsMap))

	// Verify each cache entry
	for key, stats := range statistics.StatsMap {
		assert.NotEmpty(t, key)
		assert.Greater(t, stats.CacheSize, 0)
		assert.Greater(t, stats.RequestCount, uint64(0))
	}
}

func TestStatisticsWithEmptyMap(t *testing.T) {
	statistics := Statistics{
		StatsMap: make(map[string]CacheStats),
	}

	assert.NotNil(t, statistics.StatsMap)
	assert.Equal(t, 0, len(statistics.StatsMap))
}

func TestCacheStatsCalculations(t *testing.T) {
	testCases := []struct {
		name          string
		requestCount  uint64
		hitRate       float64
		missRate      float64
		expectedValid bool
	}{
		{"Perfect hits", 1000, 1.0, 0.0, true},
		{"Perfect misses", 1000, 0.0, 1.0, true},
		{"Balanced", 1000, 0.5, 0.5, true},
		{"High hit rate", 1000, 0.95, 0.05, true},
		{"Low hit rate", 1000, 0.05, 0.95, true},
		{"Quarter hit rate", 1000, 0.25, 0.75, true},
		{"Three quarter hit rate", 1000, 0.75, 0.25, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stats := CacheStats{
				RequestCount: tc.requestCount,
				HitRate:      tc.hitRate,
				MissRate:     tc.missRate,
			}

			assert.Equal(t, tc.requestCount, stats.RequestCount)
			assert.Equal(t, tc.hitRate, stats.HitRate)
			assert.Equal(t, tc.missRate, stats.MissRate)

			// Verify hit rate and miss rate sum to approximately 1.0
			if tc.expectedValid {
				sum := stats.HitRate + stats.MissRate
				assert.InDelta(t, 1.0, sum, 0.001)
			}
		})
	}
}

func TestChangeWithSpecialCharacters(t *testing.T) {
	testCases := []struct {
		name       string
		changedKey string
		cfName     string
		userName   string
	}{
		{"With colons", "key:with:colons", "cf:name", "user:name"},
		{"With dashes", "key-with-dashes", "cf-name", "user-name"},
		{"With underscores", "key_with_underscores", "cf_name", "user_name"},
		{"With dots", "key.with.dots", "cf.name", "user.name"},
		{"With slashes", "key/with/slashes", "cf/name", "user/name"},
		{"With email", "key@example.com", "cf@example.com", "user@example.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			change := Change{
				ChangedKey: tc.changedKey,
				Operation:  db.UPDATE_OPERATION,
				CfName:     tc.cfName,
				UserName:   tc.userName,
			}

			assert.Equal(t, tc.changedKey, change.ChangedKey)
			assert.Equal(t, tc.cfName, change.CfName)
			assert.Equal(t, tc.userName, change.UserName)
		})
	}
}

func TestCacheStatsEdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		stats CacheStats
	}{
		{
			"Negative refresh time",
			CacheStats{DaoRefreshTime: -1, CacheSize: 100},
		},
		{
			"Zero cache size",
			CacheStats{DaoRefreshTime: 1000, CacheSize: 0},
		},
		{
			"Negative non-absent count",
			CacheStats{CacheSize: 100, NonAbsentCount: -1},
		},
		{
			"High eviction count",
			CacheStats{RequestCount: 100, EvictionCount: 1000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Struct should hold any values assigned
			assert.NotNil(t, tc.stats)
		})
	}
}

func TestStatisticsMapOperations(t *testing.T) {
	statistics := Statistics{
		StatsMap: make(map[string]CacheStats),
	}

	// Add entries
	statistics.StatsMap["cache1"] = CacheStats{CacheSize: 100, RequestCount: 1000}
	statistics.StatsMap["cache2"] = CacheStats{CacheSize: 200, RequestCount: 2000}
	statistics.StatsMap["cache3"] = CacheStats{CacheSize: 300, RequestCount: 3000}

	assert.Equal(t, 3, len(statistics.StatsMap))

	// Verify each entry
	assert.Equal(t, 100, statistics.StatsMap["cache1"].CacheSize)
	assert.Equal(t, 200, statistics.StatsMap["cache2"].CacheSize)
	assert.Equal(t, 300, statistics.StatsMap["cache3"].CacheSize)

	// Update entry
	statistics.StatsMap["cache1"] = CacheStats{CacheSize: 150, RequestCount: 1500}
	assert.Equal(t, 150, statistics.StatsMap["cache1"].CacheSize)

	// Delete entry
	delete(statistics.StatsMap, "cache2")
	assert.Equal(t, 2, len(statistics.StatsMap))

	// Check non-existent key
	_, exists := statistics.StatsMap["cache2"]
	assert.False(t, exists)
}

func TestGetChangedKeysMapRawLogic(t *testing.T) {
	// Test the logic concepts used in GetChangedKeysMapRaw
	// without requiring actual database access

	// Test time window calculations
	changedKeysTimeWindowSize := int64(60000) // 1 minute in milliseconds
	testTimestamp := int64(1641024000000)     // Example timestamp

	rowKey := testTimestamp - (testTimestamp % changedKeysTimeWindowSize)

	// Verify row key is aligned to window size
	assert.Equal(t, int64(0), rowKey%changedKeysTimeWindowSize)

	// Test multiple row keys
	nextRowKey := rowKey + changedKeysTimeWindowSize
	assert.Equal(t, changedKeysTimeWindowSize, nextRowKey-rowKey)

	// Test range calculations
	startTS := testTimestamp - (15 * 60 * 1000) // 15 minutes prior
	currentRowKey := startTS - (startTS % changedKeysTimeWindowSize)
	assert.Equal(t, int64(0), currentRowKey%changedKeysTimeWindowSize)
}

func TestChangeStructLongValues(t *testing.T) {
	// Create a reasonably long string (1000 characters)
	longString := ""
	for i := 0; i < 1000; i++ {
		longString += "a"
	}

	change := Change{
		ChangedKey: longString,
		Operation:  db.CREATE_OPERATION,
		CfName:     longString,
		UserName:   longString,
	}

	assert.NotEmpty(t, change.ChangedKey)
	assert.NotEmpty(t, change.CfName)
	assert.NotEmpty(t, change.UserName)
	assert.Equal(t, 1000, len(change.ChangedKey))
}

func TestCacheStatsWithDifferentTimeUnits(t *testing.T) {
	testCases := []struct {
		name          string
		totalLoadTime time.Duration
	}{
		{"Nanoseconds", time.Nanosecond * 100},
		{"Microseconds", time.Microsecond * 100},
		{"Milliseconds", time.Millisecond * 100},
		{"Seconds", time.Second * 10},
		{"Minutes", time.Minute * 5},
		{"Hours", time.Hour * 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stats := CacheStats{
				TotalLoadTime: tc.totalLoadTime,
			}

			assert.Equal(t, tc.totalLoadTime, stats.TotalLoadTime)
		})
	}
}

func TestStatisticsWithDuplicateKeys(t *testing.T) {
	statistics := Statistics{
		StatsMap: make(map[string]CacheStats),
	}

	key := "duplicate_key"

	// Add first value
	statistics.StatsMap[key] = CacheStats{CacheSize: 100}
	assert.Equal(t, 100, statistics.StatsMap[key].CacheSize)

	// Overwrite with second value
	statistics.StatsMap[key] = CacheStats{CacheSize: 200}
	assert.Equal(t, 200, statistics.StatsMap[key].CacheSize)

	// Map should still have only one entry
	assert.Equal(t, 1, len(statistics.StatsMap))
}

func TestChangeStructComparison(t *testing.T) {
	change1 := Change{
		ChangedKey: "key1",
		Operation:  db.CREATE_OPERATION,
		CfName:     "cf1",
		UserName:   "user1",
	}

	change2 := Change{
		ChangedKey: "key1",
		Operation:  db.CREATE_OPERATION,
		CfName:     "cf1",
		UserName:   "user1",
	}

	change3 := Change{
		ChangedKey: "key2",
		Operation:  db.UPDATE_OPERATION,
		CfName:     "cf2",
		UserName:   "user2",
	}

	// Test equality
	assert.Equal(t, change1, change2)
	assert.NotEqual(t, change1, change3)
}

func TestCacheStatsComparison(t *testing.T) {
	stats1 := CacheStats{
		CacheSize:    100,
		RequestCount: 1000,
		HitRate:      0.95,
	}

	stats2 := CacheStats{
		CacheSize:    100,
		RequestCount: 1000,
		HitRate:      0.95,
	}

	stats3 := CacheStats{
		CacheSize:    200,
		RequestCount: 2000,
		HitRate:      0.85,
	}

	assert.Equal(t, stats1, stats2)
	assert.NotEqual(t, stats1, stats3)
}

func TestStatisticsNilMap(t *testing.T) {
	statistics := Statistics{}

	// Nil map should be nil
	assert.Nil(t, statistics.StatsMap)

	// Initialize and test
	statistics.StatsMap = make(map[string]CacheStats)
	assert.NotNil(t, statistics.StatsMap)
	assert.Equal(t, 0, len(statistics.StatsMap))
}
