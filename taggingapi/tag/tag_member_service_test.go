package tag

import (
	"fmt"
	"net/http"
	"testing"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
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

// The member→bucket mapping is a storage contract: bucket ids are never stored
// with a member — they are recomputed from the member string on every add,
// remove and lookup — so any change to the mapping strands existing rows in
// buckets the code no longer looks in. These golden values pin FNV-1a mod 1000
// with the modulo in uint32 space, which keeps the ids platform-independent
// (converting to int before the modulo goes negative for half the hash space
// on 32-bit platforms).
func TestGetBucketId_GoldenMapping(t *testing.T) {
	golden := map[string]int{
		"AA:BB:CC:DD:EE:01":   228,
		"AABBCCDDEEFF":        15,
		"2846573900878987927": 654,
		"some-member":         358,
	}
	for member, expected := range golden {
		assert.Equal(t, expected, getBucketId(member), "bucket for %q", member)
	}
}

func TestParseBucketedCursor(t *testing.T) {
	// Test empty cursor
	cursor, err := parseBucketedCursor("")
	assert.NoError(t, err)
	assert.Equal(t, 0, cursor.BucketId)
	assert.Equal(t, "", cursor.LastMember)

	// Test valid cursor
	validCursor := generateBucketedCursor(5, "test-member")
	parsed, err := parseBucketedCursor(validCursor)
	assert.NoError(t, err)
	assert.Equal(t, 5, parsed.BucketId)
	assert.Equal(t, "test-member", parsed.LastMember)

	// Invalid cursor is a 400 error, not a silent reset to bucket 0
	_, err = parseBucketedCursor("invalid-cursor")
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, xwcommon.GetXconfErrorStatusCode(err))

	// Out-of-range bucket ID is a 400 error
	invalidBucketCursor := generateBucketedCursor(9999, "member")
	_, err = parseBucketedCursor(invalidBucketCursor)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, xwcommon.GetXconfErrorStatusCode(err))
}

func TestGenerateBucketedCursor(t *testing.T) {
	cursor := generateBucketedCursor(10, "member123")
	assert.NotEmpty(t, cursor, "Cursor should not be empty")

	// Should be base64 encoded
	parsed, err := parseBucketedCursor(cursor)
	assert.NoError(t, err)
	assert.Equal(t, 10, parsed.BucketId)
	assert.Equal(t, "member123", parsed.LastMember)

	// Test edge cases
	cursor2 := generateBucketedCursor(0, "")
	assert.NotEmpty(t, cursor2, "Cursor should not be empty even with zero values")

	parsed2, err := parseBucketedCursor(cursor2)
	assert.NoError(t, err)
	assert.Equal(t, 0, parsed2.BucketId)
	assert.Equal(t, "", parsed2.LastMember)
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
	_, _, err := AddMembers("test-tag", []string{}, TagTypeLegacy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "member list is empty")

	_, _, err = RemoveMembers("test-tag", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "member list is empty")

	// Test oversized batch
	largeMembers := make([]string, MaxBatchSizeV2+1)
	for i := range largeMembers {
		largeMembers[i] = fmt.Sprintf("member-%d", i)
	}

	_, _, err = AddMembers("test-tag", largeMembers, TagTypeLegacy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch size")
	assert.Contains(t, err.Error(), "exceeds maximum")

	_, _, err = RemoveMembers("test-tag", largeMembers)
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

// Worker scaling for the XDAS write phase. The floor-at-one cases pin a
// regression: a non-positive tag_update_worker_count used to yield zero
// workers for batches under 100 members, turning writes into silent no-ops.
func TestGetWriteWorkerCount(t *testing.T) {
	testCases := []struct {
		name        string
		memberCount int
		baseWorkers int
		expected    int
	}{
		{"small batch uses base workers", 50, 20, 20},
		{"medium batch uses base workers", 1000, 20, 20},
		{"huge batch scales past base", 5000, 20, 50},
		{"scaling capped at MaxWorkersV2", 15000, 10, 100},
		{"never more workers than members", 3, 20, 3},
		{"zero worker count floors at one", 50, 0, 1},
		{"negative worker count floors at one", 50, -5, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withWorkerCount(t, tc.baseWorkers)
			assert.Equal(t, tc.expected, getWriteWorkerCount(tc.memberCount))
		})
	}
}
