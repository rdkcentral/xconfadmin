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
	"time"

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
	BucketCount             = 1000
	DefaultPageSizeV2       = 500
	MaxPageSizeV2           = 200000
	MaxBatchSizeV2          = 5000
	MaxWorkersV2            = 100
	MaxMembersInTagResponse = 100000 // Max members returned in GetTagById
	MemberFetchChunkSize    = 1000   // Chunk size for memory-safe pagination

	QueryAddMemberBucketed       = `INSERT INTO "TagMembersBucketed" (tag_id, bucket_id, member, created) VALUES (?, ?, ?, ?)`
	QueryRemoveMemberBucketed    = `DELETE FROM "TagMembersBucketed" WHERE tag_id = ? AND bucket_id = ? AND member = ?`
	QueryGetMembersByBucket      = `SELECT member FROM "TagMembersBucketed" WHERE tag_id = ? AND bucket_id = ? AND member > ? LIMIT ?`
	QueryGetMembersCountByBucket = `SELECT count(*) FROM "TagMembersBucketed" WHERE tag_id = ? and bucket_id = ?`
	QueryGetMembersByBucketFirst = `SELECT member FROM "TagMembersBucketed" WHERE tag_id = ? AND bucket_id = ? LIMIT ?`

	QueryGetPopulatedBuckets  = `SELECT bucket_id FROM "TagBucketMetadata" WHERE tag_id = ?`
	QueryAddBucketMetadata    = `INSERT INTO "TagBucketMetadata" (tag_id, bucket_id) VALUES (?, ?)`
	QueryGetAllTagIds         = `SELECT tag_id FROM "TagBucketMetadata"`
	QueryDeleteBucketMembers  = `DELETE FROM "TagMembersBucketed" WHERE tag_id = ? AND bucket_id = ?`
	QueryDeleteBucketMetadata = `DELETE FROM "TagBucketMetadata" WHERE tag_id = ? AND bucket_id = ?`

	CountMembersCassandraResp = "count"
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

// bucketFetchResult holds the result of fetching members from a single bucket
type bucketFetchResult struct {
	bucketIndex int
	members     []string
	err         error
}

func getBucketId(member string) int {
	hash := fnv.New32a()
	hash.Write([]byte(member))
	return int(hash.Sum32()) % BucketCount
}

func AddMembers(tagId string, members []string) error {
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

func RemoveMembers(tagId string, members []string) error {
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
		// Clean up bucket metadata if bucket is now empty
		membersCount, err := getMembersCountOfBucket(tagId, bucketId)
		if err != nil {
			log.Warnf("Failed to check bucket %d count for tag %s: %v (skipping cleanup)", bucketId, tagId, err)
			continue
		}
		if membersCount == 0 {
			err = ds.GetSimpleDao().Modify(QueryDeleteBucketMetadata, tagId, strconv.Itoa(bucketId))
			if err != nil {
				log.Warnf("Failed to delete empty bucket %d metadata for tag %s: %v", bucketId, tagId, err)
			} else {
				log.Infof("Deleted empty bucket %d metadata for tag %s", bucketId, tagId)
			}
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

func getMembersCountOfBucket(tagId string, bucketId int) (int, error) {
	rows, err := ds.GetSimpleDao().Query(QueryGetMembersCountByBucket, tagId, strconv.Itoa(bucketId))
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	countVal, exists := rows[0][CountMembersCassandraResp]
	if !exists || countVal == nil {
		log.Errorf("Count result missing for bucket %d, tag %s", bucketId, tagId)
		return 0, fmt.Errorf("count result missing")
	}
	count, ok := countVal.(int64)
	if !ok {
		log.Errorf("Failed to parse count for bucket %d, tag %s: unexpected type %T", bucketId, tagId, countVal)
		return 0, fmt.Errorf("failed to parse count result")
	}
	return int(count), nil
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

func GetMembersPaginated(tagId string, limit int, cursor string) (*PaginatedMembersResponse, error) {
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

	// Build work items for remaining buckets (apply cursor's lastMember to first bucket only)
	workers := getReadWorkerCount()
	remainingBuckets := populatedBuckets[startIndex:]

	workItems := make([]bucketWorkItem, len(remainingBuckets))
	for idx, bucketId := range remainingBuckets {
		lm := ""
		if idx == 0 && bucketId == state.BucketId {
			lm = state.LastMember
		}
		workItems[idx] = bucketWorkItem{
			bucketId:   bucketId,
			lastMember: lm,
			limit:      limit + 1,
		}
	}

	orderedResults := fetchBucketsConcurrent(tagId, workItems, workers)

	// Merge in bucket order, building cursor at the truncation point
	lastProcessedBucketIndex := startIndex - 1
	for idx, result := range orderedResults {
		if result.err != nil || len(result.members) == 0 {
			lastProcessedBucketIndex = startIndex + idx
			continue
		}

		currentBucketId := remainingBuckets[idx]
		needed := limit - len(allMembers)

		if len(result.members) > needed {
			allMembers = append(allMembers, result.members[:needed]...)
			nextCursor := generateBucketedCursor(currentBucketId, result.members[needed-1], len(allMembers))
			log.Debugf("Returning %d members for tag %s with more data in bucket %d",
				len(allMembers), tagId, currentBucketId)
			return &PaginatedMembersResponse{
				Data:       allMembers,
				NextCursor: nextCursor,
				HasMore:    true,
			}, nil
		}

		allMembers = append(allMembers, result.members...)
		lastProcessedBucketIndex = startIndex + idx

		if len(allMembers) >= limit {
			break
		}
	}

	// Check if we have more populated buckets to process
	hasMore := lastProcessedBucketIndex+1 < len(populatedBuckets)
	var nextCursor string
	if hasMore {
		nextBucketId := populatedBuckets[lastProcessedBucketIndex+1]
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

// getReadWorkerCount returns the worker count for concurrent read operations
func getReadWorkerCount() int {
	config := GetTagApiConfig()
	if config != nil && config.WorkerCount > 0 {
		return min(config.WorkerCount, MaxWorkersV2)
	}
	return 1
}

// fetchBucketMembersWithLimit fetches all members from a single bucket in chunks
func fetchBucketMembersWithLimit(tagId string, bucketId int, lastMember string, limit int) ([]string, error) {
	collected := make([]string, 0)

	for {
		remainingCapacity := limit - len(collected)
		if remainingCapacity <= 0 {
			break
		}

		chunkLimit := min(MemberFetchChunkSize, remainingCapacity)
		chunk, err := getMembersFromBucket(tagId, bucketId, lastMember, chunkLimit)
		if err != nil {
			return collected, err
		}

		if len(chunk) == 0 {
			break
		}

		collected = append(collected, chunk...)

		if len(chunk) < chunkLimit {
			break
		}

		lastMember = chunk[len(chunk)-1]
	}

	return collected, nil
}

// bucketWorkItem represents a single bucket fetch task
type bucketWorkItem struct {
	bucketId   int
	lastMember string
	limit      int
}

// fetchBucketsConcurrent fetches members from multiple buckets using a worker pool
// Returns ordered results (one per bucket) without merging
func fetchBucketsConcurrent(tagId string, workItems []bucketWorkItem, workers int) []bucketFetchResult {
	if len(workItems) == 0 {
		return nil
	}

	numWorkers := min(workers, len(workItems))
	workChan := make(chan int, len(workItems))
	for idx := range workItems {
		workChan <- idx
	}
	close(workChan)

	resultsChan := make(chan bucketFetchResult, len(workItems))
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range workChan {
				work := workItems[idx]
				members, err := fetchBucketMembersWithLimit(tagId, work.bucketId, work.lastMember, work.limit)
				resultsChan <- bucketFetchResult{
					bucketIndex: idx,
					members:     members,
					err:         err,
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	orderedResults := make([]bucketFetchResult, len(workItems))
	for result := range resultsChan {
		if result.err != nil {
			log.Errorf("Error fetching members from bucket %d for tag %s: %v",
				workItems[result.bucketIndex].bucketId, tagId, result.err)
		}
		orderedResults[result.bucketIndex] = result
	}

	return orderedResults
}

// fetchMembersFromBucketsConcurrent fetches members from multiple buckets concurrently
// and returns a merged, truncated result
func fetchMembersFromBucketsConcurrent(tagId string, bucketIds []int, totalLimit int, workers int) ([]string, bool, error) {
	if len(bucketIds) == 0 {
		return nil, false, nil
	}

	// Build work items (all with empty lastMember for fresh fetch)
	workItems := make([]bucketWorkItem, len(bucketIds))
	for idx, bucketId := range bucketIds {
		workItems[idx] = bucketWorkItem{
			bucketId:   bucketId,
			lastMember: "",
			limit:      totalLimit,
		}
	}

	orderedResults := fetchBucketsConcurrent(tagId, workItems, workers)

	// Merge in bucket order, stop at totalLimit
	collected := make([]string, 0)
	for _, result := range orderedResults {
		if result.err != nil || len(result.members) == 0 {
			continue
		}
		space := totalLimit - len(collected)
		if space <= 0 {
			return collected, true, nil
		}
		if len(result.members) > space {
			collected = append(collected, result.members[:space]...)
			return collected, true, nil
		}
		collected = append(collected, result.members...)
	}

	wasTruncated := len(collected) >= totalLimit
	return collected, wasTruncated, nil
}

// AddMembersWithXdas adds members to both XDAS and Cassandra (XDAS-first approach)
// Returns the count of members actually stored to Cassandra.
func AddMembersWithXdas(tagId string, members []string) (int, error) {
	startTime := time.Now()

	if len(members) == 0 {
		return 0, fmt.Errorf("member list is empty")
	}

	if len(members) > MaxBatchSizeV2 {
		return 0, fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	savedToXdasMembers, err := addMembersToXdas(tagId, members)
	if err != nil {
		return 0, fmt.Errorf("XDAS operation failed: %w", err)
	}

	xdasAccepted := len(savedToXdasMembers)
	cassandraStored := 0

	if xdasAccepted > 0 {
		if err := AddMembers(tagId, savedToXdasMembers); err != nil {
			duration := time.Since(startTime)
			log.Errorf("Critical: XDAS succeeded but Cassandra V2 failed for tag %s: %v", tagId, err)
			log.Infof("AddMembers summary for tag '%s': requested=%d, xdasAccepted=%d, cassandraStored=%d, duration=%v", tagId, len(members), xdasAccepted, cassandraStored, duration)
			return cassandraStored, fmt.Errorf("cassandra V2 storage failed after XDAS success: %w", err)
		}
		cassandraStored = xdasAccepted
	}

	duration := time.Since(startTime)
	log.Infof("AddMembers summary for tag '%s': requested=%d, xdasAccepted=%d, cassandraStored=%d, duration=%v", tagId, len(members), xdasAccepted, cassandraStored, duration)
	return cassandraStored, nil
}

// RemoveMembersWithXdas removes members from both XDAS and Cassandra (XDAS-first approach)
// Returns the count of members actually removed from Cassandra.
func RemoveMembersWithXdas(tagId string, members []string) (int, error) {
	startTime := time.Now()

	if len(members) == 0 {
		return 0, fmt.Errorf("member list is empty")
	}

	if len(members) > MaxBatchSizeV2 {
		return 0, fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	successfulRemovals, err := removeMembersFromXDAS(tagId, members)
	if err != nil {
		return 0, fmt.Errorf("XDAS removal failed: %w", err)
	}

	xdasRemoved := len(successfulRemovals)
	cassandraRemoved := 0

	if xdasRemoved > 0 {
		if err := RemoveMembers(tagId, successfulRemovals); err != nil {
			duration := time.Since(startTime)
			log.Errorf("Critical: XDAS removal succeeded but Cassandra V2 failed for tag %s: %v", tagId, err)
			log.Infof("RemoveMembers summary for tag '%s': requested=%d, xdasRemoved=%d, cassandraRemoved=%d, duration=%v", tagId, len(members), xdasRemoved, cassandraRemoved, duration)
			return cassandraRemoved, fmt.Errorf("cassandra V2 removal failed after XDAS success: %w", err)
		}
		cassandraRemoved = xdasRemoved
	}

	duration := time.Since(startTime)
	log.Infof("RemoveMembers summary for tag '%s': requested=%d, xdasRemoved=%d, cassandraRemoved=%d, duration=%v", tagId, len(members), xdasRemoved, cassandraRemoved, duration)
	return cassandraRemoved, nil
}

// RemoveMemberWithXdas removes a single member from both XDAS and Cassandra V2
func RemoveMemberWithXdas(tagId string, member string) error {
	_, err := RemoveMembersWithXdas(tagId, []string{member})
	return err
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

// GetAllTagIds returns all tag IDs from V2 tables
func GetAllTagIds() ([]string, error) {
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

// GetTagById retrieves a tag with up to MaxMembersInTagResponse members
func GetTagById(tagId string) ([]string, bool, error) {
	populatedBuckets, err := getPopulatedBuckets(tagId)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get populated buckets: %w", err)
	}

	if len(populatedBuckets) == 0 {
		return nil, false, fmt.Errorf("tag not found")
	}

	log.Infof("Fetching tag '%s' with %d populated buckets", tagId, len(populatedBuckets))

	workers := getReadWorkerCount()
	collected, wasTruncated, err := fetchMembersFromBucketsConcurrent(
		tagId, populatedBuckets, MaxMembersInTagResponse, workers)
	if err != nil {
		return nil, false, err
	}

	log.Infof("Tag '%s': retrieved %d members, truncated=%v", tagId, len(collected), wasTruncated)
	return collected, wasTruncated, nil
}

// DeleteTag deletes a tag completely from V2 storage (XDAS and Cassandra)
// Uses memory-safe chunked deletion to handle tags with millions of members
func DeleteTag(tagId string) error {
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
			if err := RemoveMembers(tagId, removedFromXdas); err != nil {
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

// GetMembersNonPaginated retrieves tag members for non-paginated response (V1 compatibility)
// Returns up to MaxMembersInTagResponse (100k) members as a plain array
func GetMembersNonPaginated(tagId string) ([]string, bool, error) {
	populatedBuckets, err := getPopulatedBuckets(tagId)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get populated buckets: %w", err)
	}

	if len(populatedBuckets) == 0 {
		return nil, false, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, tagId))
	}

	log.Infof("Fetching tag members for '%s' (non-paginated) with %d populated buckets", tagId, len(populatedBuckets))

	workers := getReadWorkerCount()
	collected, wasTruncated, err := fetchMembersFromBucketsConcurrent(
		tagId, populatedBuckets, MaxMembersInTagResponse, workers)
	if err != nil {
		return nil, false, err
	}

	log.Infof("Tag '%s': retrieved %d members (non-paginated), truncated=%v", tagId, len(collected), wasTruncated)
	return collected, wasTruncated, nil
}
