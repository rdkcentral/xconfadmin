package tag

import (
	"fmt"
	"testing"
)

func generateTestMembers(count int) []string {
	members := make([]string, count)
	for i := 0; i < count; i++ {
		// Generate MAC address-like strings
		members[i] = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
			i%256, (i/256)%256, (i/65536)%256,
			(i+1)%256, (i+2)%256, (i+3)%256)
	}
	return members
}

func generateTestMembersSimple(count int) []string {
	members := make([]string, count)
	for i := 0; i < count; i++ {
		members[i] = fmt.Sprintf("test-member-%06d", i)
	}
	return members
}

func BenchmarkGetBucketId(b *testing.B) {
	member := "00:11:22:33:44:55"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		getBucketId(member)
	}
}

func BenchmarkGetBucketIdMACAddresses(b *testing.B) {
	members := generateTestMembers(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, member := range members {
			getBucketId(member)
		}
	}
}

func BenchmarkGenerateBucketedCursor(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		generateBucketedCursor(i%BucketCount, fmt.Sprintf("member-%d", i), i)
	}
}

func BenchmarkParseBucketedCursor(b *testing.B) {
	// Pre-generate cursors
	cursors := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		cursors[i] = generateBucketedCursor(i%BucketCount, fmt.Sprintf("member-%d", i), i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parseBucketedCursor(cursors[i%1000])
	}
}

// Benchmark bucket distribution for different member types
func BenchmarkBucketDistributionMAC(b *testing.B) {
	members := generateTestMembers(10000)
	buckets := make(map[int]int)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, member := range members {
			bucket := getBucketId(member)
			buckets[bucket]++
		}
	}
}

func BenchmarkBucketDistributionSimple(b *testing.B) {
	members := generateTestMembersSimple(10000)
	buckets := make(map[int]int)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, member := range members {
			bucket := getBucketId(member)
			buckets[bucket]++
		}
	}
}

// Benchmarks that would require database setup
func BenchmarkAddMembersV2(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	members := generateTestMembers(1000)
	tagId := "benchmark-tag"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// This would require proper database setup
		// In a real benchmark, you'd set up a test database here
		b.StartTimer()

		// AddMembersV2(tagId, members)
		// Placeholder - actual implementation would call the function
		_ = tagId
		_ = members
	}
}

func BenchmarkAddMembersV2Small(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	members := generateTestMembers(10)
	tagId := "benchmark-tag-small"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// This would require proper database setup
		b.StartTimer()

		// AddMembersV2(tagId, members)
		_ = tagId
		_ = members
	}
}

func BenchmarkAddMembersV2Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	members := generateTestMembers(5000) // Max batch size
	tagId := "benchmark-tag-large"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// This would require proper database setup
		b.StartTimer()

		// AddMembersV2(tagId, members)
		_ = tagId
		_ = members
	}
}

func BenchmarkRemoveMembersV2(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	members := generateTestMembers(1000)
	tagId := "benchmark-tag-remove"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// This would require proper database setup and pre-populated data
		b.StartTimer()

		// RemoveMembersV2(tagId, members)
		_ = tagId
		_ = members
	}
}

func BenchmarkGetMembersV2Paginated(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	tagId := "benchmark-tag-pagination"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// This would require proper database setup and pre-populated data
		b.StartTimer()

		// GetMembersV2Paginated(tagId, 500, "")
		_ = tagId
	}
}

func BenchmarkGetMembersV2PaginatedWithCursor(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	tagId := "benchmark-tag-pagination-cursor"
	cursor := generateBucketedCursor(100, "member-12345", 500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// This would require proper database setup and pre-populated data
		b.StartTimer()

		// GetMembersV2Paginated(tagId, 500, cursor)
		_ = tagId
		_ = cursor
	}
}

// Benchmark memory allocation patterns
func BenchmarkMemberSliceAllocation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		members := make([]string, 0, 5000) // Pre-allocate capacity
		for j := 0; j < 1000; j++ {
			members = append(members, fmt.Sprintf("member-%d", j))
		}
		_ = members
	}
}

func BenchmarkBucketGroupAllocation(b *testing.B) {
	members := generateTestMembers(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bucketGroups := make(map[int][]string)
		for _, member := range members {
			bucketId := getBucketId(member)
			bucketGroups[bucketId] = append(bucketGroups[bucketId], member)
		}
		_ = bucketGroups
	}
}
