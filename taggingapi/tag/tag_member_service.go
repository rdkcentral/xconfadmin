package tag

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"strings"
	"sync"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	LoggedBatch   = 0
	UnloggedBatch = 1
	CounterBatch  = 2
)

const (
	TableTagMembersBucketed = "TagMembersBucketed"
	BucketCount             = 1000
	DefaultPageSizeV2       = 500
	MaxPageSizeV2           = 5000
	MaxBatchSizeV2          = 5000
	MaxWorkersV2            = 100
	MaxMembersInTagResponse = 100000 // Max members returned in GetTagById
	MemberFetchChunkSize    = 1000   // Chunk size for memory-safe pagination

	QueryAddMemberBucketed       = `INSERT INTO "TagMembersBucketed" (tag_id, bucket_id, member, created) VALUES (?, ?, ?, ?)`
	QueryRemoveMemberBucketed    = `DELETE FROM "TagMembersBucketed" WHERE tag_id = ? AND bucket_id = ? AND member = ?`
	QueryGetMembersByBucket      = `SELECT member FROM "TagMembersBucketed" WHERE tag_id = ? AND bucket_id = ? AND member > ? LIMIT ?`
	QueryGetMembersByBucketFirst = `SELECT member FROM "TagMembersBucketed" WHERE tag_id = ? AND bucket_id = ? LIMIT ?`

	QueryGetPopulatedBuckets  = `SELECT bucket_id FROM "TagBucketMetadata" WHERE tag_id = ?`
	QueryAddBucketMetadata    = `INSERT INTO "TagBucketMetadata" (tag_id, bucket_id) VALUES (?, ?)`
	QueryRemoveBucketMetadata = `DELETE FROM "TagBucketMetadata" WHERE tag_id = ? AND bucket_id = ?`
	QueryGetAllTagIds         = `SELECT tag_id FROM "TagBucketMetadata"`
	QueryDeleteBucketMembers  = `DELETE FROM "TagMembersBucketed" WHERE tag_id = ? AND bucket_id = ?`
	QueryDeleteBucketMetadata = `DELETE FROM "TagBucketMetadata" WHERE tag_id = ? AND bucket_id = ?`
)

type BucketedCursor struct {
	BucketId       int    `json:"bucketId"`
	LastMember     string `json:"lastMember,omitempty"`
	TotalCollected int    `json:"totalCollected"`
}

type PaginatedMembersResponse struct {
	Data       []string `json:"data"`
	NextCursor string   `json:"nextCursor,omitempty"`
	HasMore    bool     `json:"hasMore"`
}

type PaginationParams struct {
	Limit  int    `json:"limit"`
	Cursor string `json:"cursor,omitempty"`
}

func getBucketId(member string) int {
	hash := fnv.New32a()
	hash.Write([]byte(member))
	return int(hash.Sum32()) % BucketCount
}

func AddMembersV2(tagId string, members []string) error {
	if len(members) > MaxBatchSizeV2 {
		return fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	if len(members) == 0 {
		return fmt.Errorf("member list is empty")
	}

	// Group by bucket for efficient batching
	bucketGroups := make(map[int][]string)
	for _, member := range members {
		bucketId := getBucketId(member)
		bucketGroups[bucketId] = append(bucketGroups[bucketId], member)
	}

	created := strconv.FormatInt(util.GetTimestamp(), 10)
	var allErrors []string
	successCount := 0

	for bucketId, bucketMembers := range bucketGroups {
		if err := addMembersToBucket(tagId, bucketId, bucketMembers, created); err != nil {
			allErrors = append(allErrors, fmt.Sprintf("bucket %d: %v", bucketId, err))
			log.Errorf("Failed to add %d members to bucket %d for tag %s: %v",
				len(bucketMembers), bucketId, tagId, err)
		} else {
			successCount += len(bucketMembers)
			log.Debugf("Successfully added %d members to bucket %d for tag %s",
				len(bucketMembers), bucketId, tagId)
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("failed to add %d/%d members: %s",
			len(members)-successCount, len(members), strings.Join(allErrors, "; "))
	}

	log.Infof("Successfully added %d members to tag %s across %d buckets",
		successCount, tagId, len(bucketGroups))
	return nil
}

func addMembersToBucket(tagId string, bucketId int, members []string, created string) error {
	batch := ds.GetSimpleDao().NewBatch(UnloggedBatch)

	// Add member records
	for _, member := range members {
		batch.Query(QueryAddMemberBucketed, tagId, strconv.Itoa(bucketId), member, created)
	}

	// Add metadata record for this bucket (will be ignored if already exists)
	batch.Query(QueryAddBucketMetadata, tagId, strconv.Itoa(bucketId))

	return ds.GetSimpleDao().ExecuteBatch(batch)
}

func RemoveMembersV2(tagId string, members []string) error {
	if len(members) > MaxBatchSizeV2 {
		return fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	if len(members) == 0 {
		return fmt.Errorf("member list is empty")
	}

	// Group by bucket for efficient batching
	bucketGroups := make(map[int][]string)
	for _, member := range members {
		bucketId := getBucketId(member)
		bucketGroups[bucketId] = append(bucketGroups[bucketId], member)
	}

	var allErrors []string
	successCount := 0

	for bucketId, bucketMembers := range bucketGroups {
		if err := removeMembersFromBucket(tagId, bucketId, bucketMembers); err != nil {
			allErrors = append(allErrors, fmt.Sprintf("bucket %d: %v", bucketId, err))
			log.Errorf("Failed to remove %d members from bucket %d for tag %s: %v",
				len(bucketMembers), bucketId, tagId, err)
		} else {
			successCount += len(bucketMembers)
			log.Debugf("Successfully removed %d members from bucket %d for tag %s",
				len(bucketMembers), bucketId, tagId)
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("failed to remove %d/%d members: %s",
			len(members)-successCount, len(members), strings.Join(allErrors, "; "))
	}

	log.Infof("Successfully removed %d members from tag %s across %d buckets",
		successCount, tagId, len(bucketGroups))
	return nil
}

func removeMembersFromBucket(tagId string, bucketId int, members []string) error {
	batch := ds.GetSimpleDao().NewBatch(UnloggedBatch)

	for _, member := range members {
		batch.Query(QueryRemoveMemberBucketed, tagId, strconv.Itoa(bucketId), member)
	}

	return ds.GetSimpleDao().ExecuteBatch(batch)
}

func getPopulatedBuckets(tagId string) ([]int, error) {
	rows, err := ds.GetSimpleDao().Query(QueryGetPopulatedBuckets, tagId)
	if err != nil {
		return nil, err
	}

	buckets := make([]int, 0, len(rows))
	for _, row := range rows {
		if bucketId, ok := row["bucket_id"].(int); ok {
			buckets = append(buckets, bucketId)
		}
	}

	return buckets, nil
}

func GetMembersV2Paginated(tagId string, limit int, cursor string) (*PaginatedMembersResponse, error) {
	if limit > MaxPageSizeV2 {
		limit = MaxPageSizeV2
	}
	if limit <= 0 {
		limit = DefaultPageSizeV2
	}

	log.Debugf("Getting paginated members for tag %s, limit %d, cursor %s", tagId, limit, cursor)

	populatedBuckets, err := getPopulatedBuckets(tagId)
	if err != nil {
		log.Errorf("Error getting populated buckets for tag %s: %v", tagId, err)
	}

	if len(populatedBuckets) == 0 {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, tagId))
	}

	log.Debugf("Found %d populated buckets for tag %s", len(populatedBuckets), tagId)

	state := parseBucketedCursor(cursor)
	var allMembers []string

	startIndex := 0
	for i, bucketId := range populatedBuckets {
		if bucketId >= state.BucketId {
			startIndex = i
			break
		}
	}

	lastProcessedIndex := startIndex - 1 // Track the last bucket we processed

	for i := startIndex; i < len(populatedBuckets) && len(allMembers) < limit; i++ {
		lastProcessedIndex = i // Update as we process each bucket
		bucketId := populatedBuckets[i]

		lastMember := ""
		if bucketId == state.BucketId {
			lastMember = state.LastMember
		}

		bucketMembers, err := getMembersFromBucket(tagId, bucketId, lastMember, limit-len(allMembers)+1)
		if err != nil {
			log.Errorf("Error getting members from bucket %d for tag %s: %v", bucketId, tagId, err)
			continue
		}

		if len(bucketMembers) == 0 {
			continue
		}

		needed := limit - len(allMembers)
		if len(bucketMembers) > needed {
			allMembers = append(allMembers, bucketMembers[:needed]...)
			nextCursor := generateBucketedCursor(bucketId, bucketMembers[needed-1], len(allMembers))
			log.Debugf("Returning %d members for tag %s with more data in bucket %d",
				len(allMembers), tagId, bucketId)
			return &PaginatedMembersResponse{
				Data:       allMembers,
				NextCursor: nextCursor,
				HasMore:    true,
			}, nil
		}

		allMembers = append(allMembers, bucketMembers...)
	}

	// Check if we have more populated buckets to process
	// hasMore is true only if there are more buckets after the last one we processed
	hasMore := lastProcessedIndex+1 < len(populatedBuckets)
	var nextCursor string
	if hasMore {
		nextBucketId := populatedBuckets[lastProcessedIndex+1]
		nextCursor = generateBucketedCursor(nextBucketId, "", 0)
	}

	log.Debugf("Returning %d members for tag %s, hasMore: %v", len(allMembers), tagId, hasMore)
	return &PaginatedMembersResponse{
		Data:       allMembers,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func getMembersFromBucket(tagId string, bucketId int, lastMember string, limit int) ([]string, error) {
	var query string
	var args []string

	if lastMember == "" {
		query = QueryGetMembersByBucketFirst
		args = []string{tagId, strconv.Itoa(bucketId), strconv.Itoa(limit)}
	} else {
		query = QueryGetMembersByBucket
		args = []string{tagId, strconv.Itoa(bucketId), lastMember, strconv.Itoa(limit)}
	}

	rows, err := ds.GetSimpleDao().Query(query, args...)
	if err != nil {
		return nil, err
	}

	members := make([]string, 0, len(rows))
	for _, row := range rows {
		if member, ok := row["member"].(string); ok {
			members = append(members, member)
		}
	}

	return members, nil
}

// Cursor management functions
func generateBucketedCursor(bucketId int, lastMember string, totalCollected int) string {
	cursor := BucketedCursor{
		BucketId:       bucketId,
		LastMember:     lastMember,
		TotalCollected: totalCollected,
	}

	data, err := json.Marshal(cursor)
	if err != nil {
		log.Errorf("Error marshaling cursor: %v", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(data)
}

func parseBucketedCursor(cursor string) BucketedCursor {
	if cursor == "" {
		return BucketedCursor{BucketId: 0}
	}

	data, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		log.Errorf("Error decoding cursor: %v", err)
		return BucketedCursor{BucketId: 0}
	}

	var state BucketedCursor
	if err := json.Unmarshal(data, &state); err != nil {
		log.Errorf("Error unmarshaling cursor: %v", err)
		return BucketedCursor{BucketId: 0}
	}

	// Validate cursor values
	if state.BucketId < 0 || state.BucketId >= BucketCount {
		log.Warnf("Invalid bucket ID in cursor: %d, resetting to 0", state.BucketId)
		return BucketedCursor{BucketId: 0}
	}

	return state
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// AddMembersWithXdas adds members to both XDAS and Cassandra (XDAS-first approach)
func AddMembersWithXdas(tagId string, members []string) error {
	if len(members) == 0 {
		return fmt.Errorf("member list is empty")
	}

	if len(members) > MaxBatchSizeV2 {
		return fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	savedToXdasMembers, err := addMembersToXdas(tagId, members)
	if err != nil {
		return fmt.Errorf("XDAS operation failed: %w", err)
	}

	if len(savedToXdasMembers) > 0 {
		if err := AddMembersV2(tagId, savedToXdasMembers); err != nil {
			// Log error but don't remove from XDAS to maintain consistency
			log.Errorf("Critical: XDAS succeeded but Cassandra V2 failed for tag %s: %v", tagId, err)
			return fmt.Errorf("cassandra V2 storage failed after XDAS success: %w", err)
		}
	}

	log.Infof("Successfully added %d members to tag %s (V2+XDAS)", len(savedToXdasMembers), tagId)
	return nil
}

// RemoveMembersV2WithXdas removes members from both XDAS and Cassandra (XDAS-first approach)
func RemoveMembersV2WithXdas(tagId string, members []string) error {
	if len(members) == 0 {
		return fmt.Errorf("member list is empty")
	}

	if len(members) > MaxBatchSizeV2 {
		return fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	successfulRemovals, err := removeMembersFromXDAS(tagId, members)
	if err != nil {
		return fmt.Errorf("XDAS removal failed: %w", err)
	}

	if len(successfulRemovals) > 0 {
		if err := RemoveMembersV2(tagId, successfulRemovals); err != nil {
			log.Errorf("Critical: XDAS removal succeeded but Cassandra V2 failed for tag %s: %v", tagId, err)
			return fmt.Errorf("cassandra V2 removal failed after XDAS success: %w", err)
		}
	}

	log.Infof("Successfully removed %d members from tag %s (V2+XDAS)", len(successfulRemovals), tagId)
	return nil
}

// RemoveMemberV2WithXdas removes a single member from both XDAS and Cassandra V2
func RemoveMemberV2WithXdas(tagId string, member string) error {
	return RemoveMembersV2WithXdas(tagId, []string{member})
}

// addMembersToXdas adds members to Xdas using concurrent workers (similar to V1 pattern)
func addMembersToXdas(tagId string, members []string) ([]string, error) {
	tagId = SetTagPrefix(tagId)

	membersChannel := make(chan string, len(members))
	go func() {
		defer close(membersChannel)
		for _, member := range members {
			membersChannel <- member
		}
	}()

	wg := &sync.WaitGroup{}
	savedMembersChannel := make(chan string, len(members))

	config := GetTagApiConfig()
	numOfWorkers := 1
	if config != nil {
		baseWorkers := config.WorkerCount
		scaledWorkers := min(max(len(members)/100, baseWorkers), MaxWorkersV2)
		numOfWorkers = scaledWorkers
	}
	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go storeTagMembersInXdas(tagId, membersChannel, savedMembersChannel, wg)
	}

	go func() {
		wg.Wait()
		close(savedMembersChannel)
	}()

	var savedMembers []string
	for savedMember := range savedMembersChannel {
		savedMembers = append(savedMembers, savedMember)
	}

	if len(savedMembers) != len(members) {
		log.Warnf("XDAS: %d/%d members successfully added to tag %s", len(savedMembers), len(members), tagId)
	}

	return savedMembers, nil
}

// removeMembersFromXDAS removes members from XDAS using concurrent workers
func removeMembersFromXDAS(tagId string, members []string) ([]string, error) {
	tagId = SetTagPrefix(tagId)

	membersChannel := make(chan string, len(members))
	go func() {
		defer close(membersChannel)
		for _, member := range members {
			membersChannel <- member
		}
	}()

	wg := &sync.WaitGroup{}
	removedMembersChannel := make(chan string, len(members))

	config := GetTagApiConfig()
	numOfWorkers := 1
	if config != nil {
		baseWorkers := config.WorkerCount
		scaledWorkers := min(max(len(members)/100, baseWorkers), MaxWorkersV2)
		numOfWorkers = scaledWorkers
	}
	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go removeTagMembersFromXdas(tagId, membersChannel, removedMembersChannel, wg)
	}

	go func() {
		wg.Wait()
		close(removedMembersChannel)
	}()

	var removedMembers []string
	for member := range removedMembersChannel {
		removedMembers = append(removedMembers, member)
	}

	if len(removedMembers) != len(members) {
		log.Warnf("XDAS: %d/%d members successfully removed from tag %s", len(removedMembers), len(members), tagId)
	}

	return removedMembers, nil
}

// GetAllTagIdsV2 returns all tag IDs from V2 tables
func GetAllTagIdsV2() ([]string, error) {
	rows, err := ds.GetSimpleDao().Query(QueryGetAllTagIds)
	if err != nil {
		return nil, fmt.Errorf("failed to query tag IDs: %w", err)
	}

	log.Debugf("Scanned %d rows from TagBucketMetadata", len(rows))

	tagIdSet := make(map[string]bool)
	for _, row := range rows {
		if tagId, ok := row["tag_id"].(string); ok {
			cleanTagId := RemovePrefixFromTag(tagId)
			tagIdSet[cleanTagId] = true
		}
	}

	tagIds := make([]string, 0, len(tagIdSet))
	for tagId := range tagIdSet {
		tagIds = append(tagIds, tagId)
	}

	log.Infof("Retrieved %d unique tag IDs from V2 storage", len(tagIds))
	return tagIds, nil
}

// GetTagByIdV2 retrieves a tag with up to MaxMembersInTagResponse members
func GetTagByIdV2(tagId string) ([]string, bool, error) {
	populatedBuckets, err := getPopulatedBuckets(tagId)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get populated buckets: %w", err)
	}

	if len(populatedBuckets) == 0 {
		return nil, false, fmt.Errorf("tag not found")
	}

	log.Infof("Fetching tag '%s' with %d populated buckets", tagId, len(populatedBuckets))

	collected := make([]string, 0, MaxMembersInTagResponse)

	for _, bucketId := range populatedBuckets {
		lastMember := ""

		for {
			space := MaxMembersInTagResponse - len(collected)
			if space <= 0 {
				log.Infof("Tag '%s': reached %d member limit, truncating", tagId, MaxMembersInTagResponse)
				return collected, true, nil
			}

			chunkLimit := min(MemberFetchChunkSize, space)
			chunk, err := getMembersFromBucket(tagId, bucketId, lastMember, chunkLimit)
			if err != nil {
				log.Errorf("Error fetching members from bucket %d for tag %s: %v", bucketId, tagId, err)
				break
			}

			if len(chunk) == 0 {
				break
			}

			collected = append(collected, chunk...)
			log.Debugf("Tag '%s': collected %d members from bucket %d (total: %d)",
				tagId, len(chunk), bucketId, len(collected))

			if len(chunk) < chunkLimit {
				break
			}

			lastMember = chunk[len(chunk)-1]
		}

		if len(collected) >= MaxMembersInTagResponse {
			log.Infof("Tag '%s': reached %d member limit after bucket %d, truncating",
				tagId, MaxMembersInTagResponse, bucketId)
			return collected, true, nil
		}
	}

	log.Infof("Tag '%s': retrieved all %d members", tagId, len(collected))
	return collected, false, nil
}

// DeleteTagV2 deletes a tag completely from V2 storage (XDAS and Cassandra)
// Uses memory-safe chunked deletion to handle tags with millions of members
func DeleteTagV2(tagId string) error {
	populatedBuckets, err := getPopulatedBuckets(tagId)
	if err != nil {
		return fmt.Errorf("failed to get populated buckets: %w", err)
	}

	if len(populatedBuckets) == 0 {
		return fmt.Errorf("tag not found")
	}

	log.Infof("Deleting tag '%s' with %d populated buckets", tagId, len(populatedBuckets))

	deletedBuckets := []int{}
	totalMembersDeleted := 0

	// Process each bucket: fetch members in chunks, delete from XDAS, then delete from Cassandra
	for _, bucketId := range populatedBuckets {
		log.Debugf("Processing bucket %d for tag '%s'", bucketId, tagId)

		membersDeleted, err := deleteBucketMembers(tagId, bucketId)
		if err != nil {
			log.Errorf("Failed to delete bucket %d for tag '%s': %v", bucketId, tagId, err)
			// Return error with partial progress saved
			return fmt.Errorf("partial deletion: %d/%d buckets deleted, %d members removed: %w",
				len(deletedBuckets), len(populatedBuckets), totalMembersDeleted, err)
		}

		totalMembersDeleted += membersDeleted
		deletedBuckets = append(deletedBuckets, bucketId)
		log.Debugf("Successfully deleted bucket %d for tag '%s' (%d members)",
			bucketId, tagId, membersDeleted)
	}

	log.Infof("Successfully deleted tag '%s': %d members removed from %d buckets",
		tagId, totalMembersDeleted, len(deletedBuckets))
	return nil
}

// deleteBucketMembers deletes all members from a single bucket (XDAS first, then Cassandra)
// Returns number of members deleted
func deleteBucketMembers(tagId string, bucketId int) (int, error) {
	totalDeleted := 0
	lastMember := ""

	for {
		chunk, err := getMembersFromBucket(tagId, bucketId, lastMember, MaxBatchSizeV2)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to fetch members from bucket: %w", err)
		}

		if len(chunk) == 0 {
			break
		}

		log.Debugf("Fetched %d members from bucket %d for tag '%s' (total deleted so far: %d)",
			len(chunk), bucketId, tagId, totalDeleted)

		removedFromXdas, err := removeMembersFromXDAS(tagId, chunk)
		if err != nil {
			return totalDeleted, fmt.Errorf("XDAS deletion failed: %w", err)
		}

		if len(removedFromXdas) > 0 {
			// Delete successfully removed members from Cassandra
			if err := RemoveMembersV2(tagId, removedFromXdas); err != nil {
				log.Errorf("Critical: XDAS deletion succeeded but Cassandra V2 deletion failed for tag %s: %v", tagId, err)
				return totalDeleted, fmt.Errorf("cassandra deletion failed after XDAS success: %w", err)
			}
			totalDeleted += len(removedFromXdas)
		}

		if len(removedFromXdas) < len(chunk) {
			log.Warnf("partial XDAS deletion: %d/%d members removed", len(removedFromXdas), len(chunk))
			return totalDeleted, nil
		}

		if len(chunk) < MaxBatchSizeV2 {
			break
		}

		lastMember = chunk[len(chunk)-1]
	}

	// All members deleted from this bucket, now delete bucket metadata
	if err := deleteBucketFromCassandra(tagId, bucketId); err != nil {
		return totalDeleted, fmt.Errorf("failed to delete bucket metadata: %w", err)
	}

	return totalDeleted, nil
}

// deleteBucketFromCassandra deletes a bucket's metadata from Cassandra
func deleteBucketFromCassandra(tagId string, bucketId int) error {
	batch := ds.GetSimpleDao().NewBatch(UnloggedBatch)

	batch.Query(QueryDeleteBucketMembers, tagId, strconv.Itoa(bucketId))
	batch.Query(QueryDeleteBucketMetadata, tagId, strconv.Itoa(bucketId))

	if err := ds.GetSimpleDao().ExecuteBatch(batch); err != nil {
		return fmt.Errorf("batch execution failed: %w", err)
	}

	log.Debugf("Deleted bucket %d metadata for tag '%s'", bucketId, tagId)
	return nil
}

// GetMembersV2NonPaginated retrieves tag members for non-paginated response (V1 compatibility)
// Returns up to MaxMembersInTagResponse (100k) members as a plain array
func GetMembersV2NonPaginated(tagId string) ([]string, bool, error) {
	populatedBuckets, err := getPopulatedBuckets(tagId)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get populated buckets: %w", err)
	}

	if len(populatedBuckets) == 0 {
		return nil, false, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, tagId))
	}

	log.Infof("Fetching tag members for '%s' (non-paginated) with %d populated buckets", tagId, len(populatedBuckets))

	collected := make([]string, 0, MaxMembersInTagResponse)

	for _, bucketId := range populatedBuckets {
		lastMember := ""

		for {
			space := MaxMembersInTagResponse - len(collected)
			if space <= 0 {
				log.Infof("Tag '%s': reached %d member limit, truncating (non-paginated)", tagId, MaxMembersInTagResponse)
				return collected, true, nil
			}

			chunkLimit := min(MemberFetchChunkSize, space)
			chunk, err := getMembersFromBucket(tagId, bucketId, lastMember, chunkLimit)
			if err != nil {
				log.Errorf("Error fetching members from bucket %d for tag %s: %v", bucketId, tagId, err)
				break
			}

			if len(chunk) == 0 {
				break
			}

			collected = append(collected, chunk...)
			log.Debugf("Tag '%s': collected %d members from bucket %d (total: %d)",
				tagId, len(chunk), bucketId, len(collected))

			if len(chunk) < chunkLimit {
				break
			}

			lastMember = chunk[len(chunk)-1]
		}

		if len(collected) >= MaxMembersInTagResponse {
			log.Infof("Tag '%s': reached %d member limit after bucket %d, truncating (non-paginated)",
				tagId, MaxMembersInTagResponse, bucketId)
			return collected, true, nil
		}
	}

	log.Infof("Tag '%s': retrieved all %d members (non-paginated)", tagId, len(collected))
	return collected, false, nil
}
