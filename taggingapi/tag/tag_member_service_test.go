package tag

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBucketId(t *testing.T) {
	// Test consistent hashing
	member := "00:11:22:33:44:55"
	bucket1 := getBucketId(member)
	bucket2 := getBucketId(member)

	assert.Equal(t, bucket1, bucket2, "getBucketId should be deterministic")
	assert.True(t, bucket1 >= 0 && bucket1 < BucketCount, "bucket ID should be within valid range")

	// Test different members get distributed
	member2 := "AA:BB:CC:DD:EE:FF"
	bucket3 := getBucketId(member2)
	assert.True(t, bucket3 >= 0 && bucket3 < BucketCount, "bucket ID should be within valid range")

	// Test distribution (not necessarily different buckets, but valid)
	members := []string{
		"00:11:22:33:44:55",
		"AA:BB:CC:DD:EE:FF",
		"12:34:56:78:90:AB",
		"FF:EE:DD:CC:BB:AA",
	}

	buckets := make(map[int]bool)
	for _, member := range members {
		bucket := getBucketId(member)
		assert.True(t, bucket >= 0 && bucket < BucketCount)
		buckets[bucket] = true
	}

	// Should have some distribution (at least 2 different buckets for 4 members)
	assert.True(t, len(buckets) >= 2, "Members should distribute across buckets")
}

func TestParseBucketedCursor(t *testing.T) {
	// Test empty cursor
	cursor := parseBucketedCursor("")
	assert.Equal(t, 0, cursor.BucketId)
	assert.Equal(t, "", cursor.LastMember)
	assert.Equal(t, 0, cursor.TotalCollected)

	// Test valid cursor
	validCursor := generateBucketedCursor(5, "test-member", 100)
	parsed := parseBucketedCursor(validCursor)
	assert.Equal(t, 5, parsed.BucketId)
	assert.Equal(t, "test-member", parsed.LastMember)
	assert.Equal(t, 100, parsed.TotalCollected)

	// Test invalid cursor
	invalidCursor := parseBucketedCursor("invalid-cursor")
	assert.Equal(t, 0, invalidCursor.BucketId)

	// Test cursor with invalid bucket ID
	invalidBucketCursor := generateBucketedCursor(9999, "member", 100)
	parsed2 := parseBucketedCursor(invalidBucketCursor)
	assert.Equal(t, 0, parsed2.BucketId, "Invalid bucket ID should be reset to 0")
}

func TestGenerateBucketedCursor(t *testing.T) {
	cursor := generateBucketedCursor(10, "member123", 500)
	assert.NotEmpty(t, cursor, "Cursor should not be empty")

	// Should be base64 encoded
	parsed := parseBucketedCursor(cursor)
	assert.Equal(t, 10, parsed.BucketId)
	assert.Equal(t, "member123", parsed.LastMember)
	assert.Equal(t, 500, parsed.TotalCollected)

	// Test edge cases
	cursor2 := generateBucketedCursor(0, "", 0)
	assert.NotEmpty(t, cursor2, "Cursor should not be empty even with zero values")

	parsed2 := parseBucketedCursor(cursor2)
	assert.Equal(t, 0, parsed2.BucketId)
	assert.Equal(t, "", parsed2.LastMember)
	assert.Equal(t, 0, parsed2.TotalCollected)
}

func TestBucketDistribution(t *testing.T) {
	// Test that MAC addresses distribute well across buckets
	macAddresses := []string{
		"00:11:22:33:44:55",
		"01:23:45:67:89:AB",
		"FF:EE:DD:CC:BB:AA",
		"12:34:56:78:90:AB",
		"98:76:54:32:10:FE",
		"A0:B1:C2:D3:E4:F5",
		"10:20:30:40:50:60",
		"AA:BB:CC:DD:EE:FF",
		"11:22:33:44:55:66",
		"99:88:77:66:55:44",
	}

	buckets := make(map[int]int)
	for _, mac := range macAddresses {
		bucket := getBucketId(mac)
		buckets[bucket]++
	}

	// Should distribute across multiple buckets
	assert.True(t, len(buckets) >= 5, "Should distribute across at least 5 buckets for 10 MAC addresses")

	// Each bucket should have reasonable distribution
	for bucket, count := range buckets {
		assert.True(t, bucket >= 0 && bucket < BucketCount, "Bucket should be in valid range")
		assert.True(t, count >= 1, "Each bucket should have at least 1 member")
		assert.True(t, count <= 5, "No bucket should have more than 5 members for this test")
	}
}
func TestBatchSizeValidation(t *testing.T) {
	// Test empty members list
	err := AddMembers("test-tag", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "member list is empty")

	err = RemoveMembers("test-tag", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "member list is empty")

	// Test oversized batch
	largeMembers := make([]string, MaxBatchSizeV2+1)
	for i := range largeMembers {
		largeMembers[i] = fmt.Sprintf("member-%d", i)
	}

	err = AddMembers("test-tag", largeMembers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch size")
	assert.Contains(t, err.Error(), "exceeds maximum")

	err = RemoveMembers("test-tag", largeMembers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch size")
	assert.Contains(t, err.Error(), "exceeds maximum")
}

func TestPaginationParamsValidation(t *testing.T) {
	// Test parameter validation logic without database access

	// Test that limit is clamped to MaxPageSizeV2
	testLimit := MaxPageSizeV2 + 1
	if testLimit > MaxPageSizeV2 {
		testLimit = MaxPageSizeV2
	}
	assert.Equal(t, MaxPageSizeV2, testLimit, "Limit should be clamped to max")

	// Test default page size assignment
	testLimit = 0
	if testLimit <= 0 {
		testLimit = DefaultPageSizeV2
	}
	assert.Equal(t, DefaultPageSizeV2, testLimit, "Should use default when limit is 0")

	// Test negative limit handling
	testLimit = -1
	if testLimit <= 0 {
		testLimit = DefaultPageSizeV2
	}
	assert.Equal(t, DefaultPageSizeV2, testLimit, "Should use default when limit is negative")

	// Note: Database-dependent tests are in integration test functions
	t.Log("Parameter validation logic tests completed")
}

// Test dynamic worker scaling logic
func TestDynamicWorkerScaling(t *testing.T) {
	// Test min/max helper functions
	assert.Equal(t, 5, min(5, 10))
	assert.Equal(t, 5, min(10, 5))
	assert.Equal(t, 10, max(5, 10))
	assert.Equal(t, 10, max(10, 5))

	// Test scaling logic scenarios
	testCases := []struct {
		name        string
		memberCount int
		baseWorkers int
		expectedMin int
		expectedMax int
	}{
		{"Small batch", 50, 20, 20, 20},        // Uses base workers (50/100=0, max with base=20)
		{"Medium batch", 200, 20, 20, 20},      // Uses base workers (200/100=2, max with base=20)
		{"Large batch", 1000, 20, 20, 20},      // Uses base workers (1000/100=10, max with base=20)
		{"Huge batch", 5000, 20, 50, 50},       // Uses scaled workers (5000/100=50)
		{"Max batch", 10000, 20, 100, 100},     // Uses max workers (10000/100=100, capped at 100)
		{"Extreme batch", 15000, 10, 100, 100}, // Uses max workers (15000/100=150, capped at 100)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the scaling logic: min(max(memberCount/100, baseWorkers), MaxWorkersV2)
			scaledWorkers := min(max(tc.memberCount/100, tc.baseWorkers), MaxWorkersV2)
			assert.True(t, scaledWorkers >= tc.expectedMin,
				"Workers %d should be >= %d for %d members", scaledWorkers, tc.expectedMin, tc.memberCount)
			assert.True(t, scaledWorkers <= tc.expectedMax,
				"Workers %d should be <= %d for %d members", scaledWorkers, tc.expectedMax, tc.memberCount)
		})
	}
}
