package tag

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"

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

	QueryAddMemberBucketed       = `INSERT INTO tag_members_bucketed (tenant_id, tag_id, bucket_id, member, value, updated) VALUES (?, ?, ?, ?, ?, ?)`
	QueryRemoveMemberBucketed    = `DELETE FROM tag_members_bucketed WHERE tenant_id = ? AND tag_id = ? AND bucket_id = ? AND member = ?`
	QueryGetMembersByBucket      = `SELECT member FROM tag_members_bucketed WHERE tenant_id = ? AND tag_id = ? AND bucket_id = ? AND member > ? LIMIT ?`
	QueryGetMembersCountByBucket = `SELECT count(*) FROM tag_members_bucketed WHERE tenant_id = ? AND tag_id = ? AND bucket_id = ?`
	QueryGetMembersByBucketFirst = `SELECT member FROM tag_members_bucketed WHERE tenant_id = ? AND tag_id = ? AND bucket_id = ? LIMIT ?`

	QueryGetPopulatedBuckets  = `SELECT bucket_id FROM tag_member_metadata WHERE tenant_id = ? AND shard_id = ? AND tag_id = ?`
	QueryAddBucketMetadata    = `INSERT INTO tag_member_metadata (tenant_id, shard_id, tag_id, bucket_id) VALUES (?, ?, ?, ?)`
	QueryGetAllTagIdsByShard  = `SELECT tag_id FROM tag_member_metadata WHERE tenant_id = ? AND shard_id = ?`
	QueryDeleteBucketMembers  = `DELETE FROM tag_members_bucketed WHERE tenant_id = ? AND tag_id = ? AND bucket_id = ?`
	QueryDeleteBucketMetadata = `DELETE FROM tag_member_metadata WHERE tenant_id = ? AND shard_id = ? AND tag_id = ? AND bucket_id = ?`

	CountMembersCassandraResp = "count"

	InvalidCursorErrorMsg = "invalid pagination cursor"
)

type BucketedCursor struct {
	BucketId   int    `json:"bucketId"`
	LastMember string `json:"lastMember,omitempty"`
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

// AddMembers writes members to the bucketed Cassandra tables. Returns the
// number of members stored and the number of buckets touched. The value is
// per-request (one PUT carries one value) and is stored alongside every member
// so Cassandra, the source of truth, can restore it if XDAS loses the entry.
func AddMembers(tenantId string, tagId string, members []string, value string) (int, int, error) {
	if len(members) > MaxBatchSizeV2 {
		return 0, 0, fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	if len(members) == 0 {
		return 0, 0, fmt.Errorf("member list is empty")
	}

	// Group by bucket for efficient batching
	bucketGroups := make(map[int][]string)
	for _, member := range members {
		bucketId := getBucketId(member)
		bucketGroups[bucketId] = append(bucketGroups[bucketId], member)
	}

	updated := time.Now()
	shardId := strconv.Itoa(ds.GetShardId(tagId))
	agg := &errorAggregator{}
	successCount := 0

	for bucketId, bucketMembers := range bucketGroups {
		if err := addMembersToBucket(tenantId, shardId, tagId, bucketId, bucketMembers, value, updated); err != nil {
			agg.add(fmt.Errorf("bucket %d: %w", bucketId, err))
		} else {
			successCount += len(bucketMembers)
			log.Debugf("Successfully added %d members to bucket %d for tag %s",
				len(bucketMembers), bucketId, tagId)
		}
	}

	// One aggregate ERROR line per call — a batch can spread over up to 1000
	// buckets, and a Cassandra outage must not emit one error line per bucket.
	if failedBuckets, firstError := agg.summary(); failedBuckets > 0 {
		log.Errorf("Failed to add members to %d/%d buckets for tag %s (first error: %s)",
			failedBuckets, len(bucketGroups), tagId, firstError)
		return successCount, len(bucketGroups), fmt.Errorf("failed to add %d/%d members (%d/%d buckets failed; first error: %s)",
			len(members)-successCount, len(members), failedBuckets, len(bucketGroups), firstError)
	}

	return successCount, len(bucketGroups), nil
}

func addMembersToBucket(tenantId string, shardId string, tagId string, bucketId int, members []string, value string, updated time.Time) error {
	batch := ds.GetSimpleDao().NewBatch(UnloggedBatch)

	// Add member records
	for _, member := range members {
		batch.Query(QueryAddMemberBucketed, tenantId, tagId, strconv.Itoa(bucketId), member, value, updated)
	}

	// Add metadata record for this bucket (will be ignored if already exists)
	batch.Query(QueryAddBucketMetadata, tenantId, shardId, tagId, strconv.Itoa(bucketId))

	return ds.GetSimpleDao().ExecuteBatch(batch)
}

// RemoveMembers deletes members from the bucketed Cassandra tables. Returns
// the number of members removed and the number of buckets touched.
func RemoveMembers(tenantId string, tagId string, members []string) (int, int, error) {
	if len(members) > MaxBatchSizeV2 {
		return 0, 0, fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	if len(members) == 0 {
		return 0, 0, fmt.Errorf("member list is empty")
	}

	// Group by bucket for efficient batching
	bucketGroups := make(map[int][]string)
	for _, member := range members {
		bucketId := getBucketId(member)
		bucketGroups[bucketId] = append(bucketGroups[bucketId], member)
	}

	shardId := strconv.Itoa(ds.GetShardId(tagId))
	agg := &errorAggregator{}
	successCount := 0

	for bucketId, bucketMembers := range bucketGroups {
		if err := removeMembersFromBucket(tenantId, tagId, bucketId, bucketMembers); err != nil {
			agg.add(fmt.Errorf("bucket %d: %w", bucketId, err))
			// The delete failed, so the bucket cannot have become empty —
			// skip the metadata cleanup check
			continue
		}
		successCount += len(bucketMembers)
		log.Debugf("Successfully removed %d members from bucket %d for tag %s",
			len(bucketMembers), bucketId, tagId)
		// Clean up bucket metadata if bucket is now empty
		membersCount, err := getMembersCountOfBucket(tenantId, tagId, bucketId)
		if err != nil {
			log.Warnf("Failed to check bucket %d count for tag %s: %v (skipping cleanup)", bucketId, tagId, err)
			continue
		}
		if membersCount == 0 {
			err = ds.GetSimpleDao().Modify(QueryDeleteBucketMetadata, tenantId, shardId, tagId, strconv.Itoa(bucketId))
			if err != nil {
				log.Warnf("Failed to delete empty bucket %d metadata for tag %s: %v", bucketId, tagId, err)
			} else {
				log.Debugf("Deleted empty bucket %d metadata for tag %s", bucketId, tagId)
			}
		}
	}

	// One aggregate ERROR line per call — see AddMembers.
	if failedBuckets, firstError := agg.summary(); failedBuckets > 0 {
		log.Errorf("Failed to remove members from %d/%d buckets for tag %s (first error: %s)",
			failedBuckets, len(bucketGroups), tagId, firstError)
		return successCount, len(bucketGroups), fmt.Errorf("failed to remove %d/%d members (%d/%d buckets failed; first error: %s)",
			len(members)-successCount, len(members), failedBuckets, len(bucketGroups), firstError)
	}

	return successCount, len(bucketGroups), nil
}

func getMembersCountOfBucket(tenantId string, tagId string, bucketId int) (int, error) {
	rows, err := ds.GetSimpleDao().Query(QueryGetMembersCountByBucket, tenantId, tagId, strconv.Itoa(bucketId))
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

func removeMembersFromBucket(tenantId string, tagId string, bucketId int, members []string) error {
	batch := ds.GetSimpleDao().NewBatch(UnloggedBatch)

	for _, member := range members {
		batch.Query(QueryRemoveMemberBucketed, tenantId, tagId, strconv.Itoa(bucketId), member)
	}

	return ds.GetSimpleDao().ExecuteBatch(batch)
}

func getPopulatedBuckets(tenantId string, tagId string) ([]int, error) {
	shardId := strconv.Itoa(ds.GetShardId(tagId))
	rows, err := ds.GetSimpleDao().Query(QueryGetPopulatedBuckets, tenantId, shardId, tagId)
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

// GetMembersPaginated returns one page of members plus the Cassandra cost of
// producing it (for the request log).
func GetMembersPaginated(tenantId string, tagId string, limit int, cursor string) (*PaginatedMembersResponse, ReadStats, error) {
	if limit > MaxPageSizeV2 {
		limit = MaxPageSizeV2
	}
	if limit <= 0 {
		limit = DefaultPageSizeV2
	}

	stats := ReadStats{}
	queries := &atomic.Int64{}

	state, err := parseBucketedCursor(cursor)
	if err != nil {
		return nil, stats, err
	}

	populatedBuckets, err := getPopulatedBuckets(tenantId, tagId)
	queries.Add(1)
	stats.Queries = int(queries.Load())
	if err != nil {
		log.Errorf("Error getting populated buckets for tag %s: %v", tagId, err)
		return nil, stats, fmt.Errorf("failed to get populated buckets: %w", err)
	}

	if len(populatedBuckets) == 0 {
		return nil, stats, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, tagId))
	}
	stats.Buckets = len(populatedBuckets)

	var allMembers []string

	// Find the first populated bucket at or after the cursor position. If none
	// exists (buckets were emptied since the previous page), the enumeration is
	// complete — do not wrap around to the beginning.
	startIndex := -1
	for i, bucketId := range populatedBuckets {
		if bucketId >= state.BucketId {
			startIndex = i
			break
		}
	}
	if startIndex == -1 {
		return &PaginatedMembersResponse{
			Data:    []string{},
			HasMore: false,
		}, stats, nil
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

	orderedResults := fetchBucketsConcurrent(tenantId, tagId, workItems, workers, queries)
	stats.Queries = int(queries.Load())

	// Merge in bucket order, building cursor at the truncation point.
	// A failed bucket fails the whole page: returning partial data would let the
	// cursor advance past the failed bucket and silently omit its members from
	// the enumeration. The caller retries with the same cursor instead.
	lastProcessedBucketIndex := startIndex - 1
	for idx, result := range orderedResults {
		if result.err != nil {
			return nil, stats, fmt.Errorf("failed to fetch members from bucket %d: %w", remainingBuckets[idx], result.err)
		}
		if len(result.members) == 0 {
			lastProcessedBucketIndex = startIndex + idx
			continue
		}

		currentBucketId := remainingBuckets[idx]
		needed := limit - len(allMembers)

		if len(result.members) > needed {
			allMembers = append(allMembers, result.members[:needed]...)
			nextCursor := generateBucketedCursor(currentBucketId, result.members[needed-1])
			log.Debugf("Returning %d members for tag %s with more data in bucket %d",
				len(allMembers), tagId, currentBucketId)
			return &PaginatedMembersResponse{
				Data:       allMembers,
				NextCursor: nextCursor,
				HasMore:    true,
			}, stats, nil
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
		nextCursor = generateBucketedCursor(nextBucketId, "")
	}

	log.Debugf("Returning %d members for tag %s, hasMore: %v", len(allMembers), tagId, hasMore)
	return &PaginatedMembersResponse{
		Data:       allMembers,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, stats, nil
}

func getMembersFromBucket(tenantId string, tagId string, bucketId int, lastMember string, limit int) ([]string, error) {
	var query string
	var args []string

	if lastMember == "" {
		query = QueryGetMembersByBucketFirst
		args = []string{tenantId, tagId, strconv.Itoa(bucketId), strconv.Itoa(limit)}
	} else {
		query = QueryGetMembersByBucket
		args = []string{tenantId, tagId, strconv.Itoa(bucketId), lastMember, strconv.Itoa(limit)}
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
func generateBucketedCursor(bucketId int, lastMember string) string {
	cursor := BucketedCursor{
		BucketId:   bucketId,
		LastMember: lastMember,
	}

	data, err := json.Marshal(cursor)
	if err != nil {
		log.Errorf("Error marshaling cursor: %v", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(data)
}

func parseBucketedCursor(cursor string) (BucketedCursor, error) {
	if cursor == "" {
		return BucketedCursor{BucketId: 0}, nil
	}

	data, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return BucketedCursor{}, xwcommon.NewRemoteErrorAS(http.StatusBadRequest, InvalidCursorErrorMsg)
	}

	var state BucketedCursor
	if err := json.Unmarshal(data, &state); err != nil {
		return BucketedCursor{}, xwcommon.NewRemoteErrorAS(http.StatusBadRequest, InvalidCursorErrorMsg)
	}

	if state.BucketId < 0 || state.BucketId >= BucketCount {
		return BucketedCursor{}, xwcommon.NewRemoteErrorAS(http.StatusBadRequest, InvalidCursorErrorMsg)
	}

	return state, nil
}

// getReadWorkerCount returns the worker count for concurrent read operations
func getReadWorkerCount() int {
	config := GetTagApiConfig()
	if config != nil && config.WorkerCount > 0 {
		return min(config.WorkerCount, MaxWorkersV2)
	}
	return 1
}

// fetchBucketMembersWithLimit fetches all members from a single bucket in
// chunks, counting issued queries into the shared counter.
func fetchBucketMembersWithLimit(tenantId string, tagId string, bucketId int, lastMember string, limit int, queries *atomic.Int64) ([]string, error) {
	collected := make([]string, 0)

	for {
		remainingCapacity := limit - len(collected)
		if remainingCapacity <= 0 {
			break
		}

		chunkLimit := min(MemberFetchChunkSize, remainingCapacity)
		queries.Add(1)
		chunk, err := getMembersFromBucket(tenantId, tagId, bucketId, lastMember, chunkLimit)
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
func fetchBucketsConcurrent(tenantId string, tagId string, workItems []bucketWorkItem, workers int, queries *atomic.Int64) []bucketFetchResult {
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
				members, err := fetchBucketMembersWithLimit(tenantId, tagId, work.bucketId, work.lastMember, work.limit, queries)
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
func fetchMembersFromBucketsConcurrent(tenantId string, tagId string, bucketIds []int, totalLimit int, workers int, queries *atomic.Int64) ([]string, bool, error) {
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

	orderedResults := fetchBucketsConcurrent(tenantId, tagId, workItems, workers, queries)

	// Merge in bucket order, stop at totalLimit.
	// A failed bucket fails the whole read — a partial result would silently
	// under-report the tag's membership.
	collected := make([]string, 0)
	for idx, result := range orderedResults {
		space := totalLimit - len(collected)
		if space <= 0 {
			// Response already full — errors in later buckets are irrelevant
			return collected, true, nil
		}
		if result.err != nil {
			return nil, false, fmt.Errorf("failed to fetch members from bucket %d: %w", bucketIds[idx], result.err)
		}
		if len(result.members) == 0 {
			continue
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

// AddMembersWithXdas adds members to both XDAS and Cassandra (XDAS-first
// approach). The returned WriteStats carry per-store outcome counts for the
// request log — populated on error paths too. The former per-request summary
// Infof lines are gone: the counts land on the framework "request ends" line.
func AddMembersWithXdas(tenantId string, tagId string, members []string, tagValue string) (WriteStats, error) {
	stats := WriteStats{Requested: len(members)}

	if len(members) == 0 {
		return stats, fmt.Errorf("member list is empty")
	}

	if len(members) > MaxBatchSizeV2 {
		return stats, fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	savedToXdasMembers, xdasFail, firstError := addMembersToXdas(tagId, members, tagValue)
	stats.XdasOk = len(savedToXdasMembers)
	stats.XdasFail = xdasFail
	stats.FirstError = firstError

	if stats.XdasOk > 0 {
		stored, buckets, err := AddMembers(tenantId, tagId, savedToXdasMembers, tagValue)
		stats.CassandraOk = stored
		stats.CassandraFail = stats.XdasOk - stored
		stats.Buckets = buckets
		if err != nil {
			if stats.FirstError == "" {
				stats.FirstError = err.Error()
			}
			logDivergence(OpAddMembers, tenantId, tagId, err)
			return stats, fmt.Errorf("cassandra V2 storage failed after XDAS success: %w", err)
		}
	}

	return stats, nil
}

// RemoveMembersWithXdas removes members from both XDAS and Cassandra
// (XDAS-first approach). See AddMembersWithXdas for the stats contract.
func RemoveMembersWithXdas(tenantId string, tagId string, members []string) (WriteStats, error) {
	stats := WriteStats{Requested: len(members)}

	if len(members) == 0 {
		return stats, fmt.Errorf("member list is empty")
	}

	if len(members) > MaxBatchSizeV2 {
		return stats, fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	successfulRemovals, xdasFail, firstError := removeMembersFromXDAS(tagId, members)
	stats.XdasOk = len(successfulRemovals)
	stats.XdasFail = xdasFail
	stats.FirstError = firstError

	if stats.XdasOk > 0 {
		removed, buckets, err := RemoveMembers(tenantId, tagId, successfulRemovals)
		stats.CassandraOk = removed
		stats.CassandraFail = stats.XdasOk - removed
		stats.Buckets = buckets
		if err != nil {
			if stats.FirstError == "" {
				stats.FirstError = err.Error()
			}
			logDivergence(OpRemoveMembers, tenantId, tagId, err)
			return stats, fmt.Errorf("cassandra V2 removal failed after XDAS success: %w", err)
		}
	}

	return stats, nil
}

// RemoveMemberWithXdas removes a single member from both XDAS and Cassandra V2
func RemoveMemberWithXdas(tenantId string, tagId string, member string) (WriteStats, error) {
	return RemoveMembersWithXdas(tenantId, tagId, []string{member})
}

// addMembersToXdas adds members to Xdas using concurrent workers (similar to
// V1 pattern). Returns the saved members, the failure count, and // first error message. Partial failure is reported through the counts, not an
// error — the caller decides how to surface it.
func addMembersToXdas(tagId string, members []string, tagValue string) ([]string, int, string) {
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
	errAgg := &errorAggregator{}

	config := GetTagApiConfig()
	numOfWorkers := 1
	if config != nil {
		baseWorkers := config.WorkerCount
		scaledWorkers := min(max(len(members)/100, baseWorkers), MaxWorkersV2)
		numOfWorkers = min(scaledWorkers, len(members)) // Never spawn more workers than members
	}
	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go storeTagMembersInXdas(tagId, membersChannel, savedMembersChannel, wg, tagValue, errAgg)
	}

	go func() {
		wg.Wait()
		close(savedMembersChannel)
	}()

	var savedMembers []string
	for savedMember := range savedMembersChannel {
		savedMembers = append(savedMembers, savedMember)
	}

	failCount, firstError := errAgg.summary()
	return savedMembers, failCount, firstError
}

// removeMembersFromXDAS removes members from XDAS using concurrent workers.
// See addMembersToXdas for the return contract.
func removeMembersFromXDAS(tagId string, members []string) ([]string, int, string) {
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
	agg := &errorAggregator{}

	config := GetTagApiConfig()
	numOfWorkers := 1
	if config != nil {
		baseWorkers := config.WorkerCount
		scaledWorkers := min(max(len(members)/100, baseWorkers), MaxWorkersV2)
		numOfWorkers = min(scaledWorkers, len(members)) // Never spawn more workers than members
	}
	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go removeTagMembersFromXdas(tagId, membersChannel, removedMembersChannel, wg, agg)
	}

	go func() {
		wg.Wait()
		close(removedMembersChannel)
	}()

	var removedMembers []string
	for member := range removedMembersChannel {
		removedMembers = append(removedMembers, member)
	}

	failCount, firstError := agg.summary()
	return removedMembers, failCount, firstError
}

// GetAllTagIds returns all tag IDs from V2 tables for a specific tenant
func GetAllTagIds(tenantId string) ([]string, error) {
	tagIdSet := make(map[string]bool)

	// Query each shard individually since SimpleDao.Query doesn't support IN clause with slices
	for _, shardId := range ds.GetShardIds() {
		rows, err := ds.GetSimpleDao().Query(QueryGetAllTagIdsByShard, tenantId, strconv.Itoa(shardId))
		if err != nil {
			return nil, fmt.Errorf("failed to query tag IDs for shard %d: %w", shardId, err)
		}
		for _, row := range rows {
			if tagId, ok := row["tag_id"].(string); ok {
				cleanTagId := RemovePrefixFromTag(tagId)
				tagIdSet[cleanTagId] = true
			}
		}
	}

	tagIds := make([]string, 0, len(tagIdSet))
	for tagId := range tagIdSet {
		tagIds = append(tagIds, tagId)
	}

	log.Debugf("Retrieved %d unique tag IDs from V2 storage for tenant %s", len(tagIds), tenantId)
	return tagIds, nil
}

// GetTagById retrieves a tag with up to MaxMembersInTagResponse members
func GetTagById(tenantId string, tagId string) ([]string, bool, ReadStats, error) {
	stats := ReadStats{}
	queries := &atomic.Int64{}

	populatedBuckets, err := getPopulatedBuckets(tenantId, tagId)
	queries.Add(1)
	stats.Queries = int(queries.Load())
	if err != nil {
		return nil, false, stats, fmt.Errorf("failed to get populated buckets: %w", err)
	}

	if len(populatedBuckets) == 0 {
		return nil, false, stats, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, tagId))
	}
	stats.Buckets = len(populatedBuckets)

	workers := getReadWorkerCount()
	collected, wasTruncated, err := fetchMembersFromBucketsConcurrent(
		tenantId, tagId, populatedBuckets, MaxMembersInTagResponse, workers, queries)
	stats.Queries = int(queries.Load())
	if err != nil {
		return nil, false, stats, err
	}

	log.Debugf("Tag '%s': retrieved %d members, truncated=%v", tagId, len(collected), wasTruncated)
	return collected, wasTruncated, stats, nil
}

// DeleteTag deletes a tag completely from V2 storage (XDAS and Cassandra)
// Uses memory-safe chunked deletion to handle tags with millions of members.
// Runs in a background goroutine, so it logs its own START/PROGRESS/END lines
// carrying the audit_id of the request that queued it.
func DeleteTag(tenantId string, tagId string, auditId string) error {
	fields := log.Fields{
		"audit_id": auditId,
		"op":       OpDeleteTag,
		"tenant":   tenantId,
		"tag":      tagId,
	}

	populatedBuckets, err := getPopulatedBuckets(tenantId, tagId)
	if err != nil {
		return fmt.Errorf("failed to get populated buckets: %w", err)
	}

	if len(populatedBuckets) == 0 {
		return fmt.Errorf("tag not found")
	}

	startTime := time.Now()
	log.WithFields(fields).Infof("tag deletion started: %d buckets", len(populatedBuckets))

	deletedBuckets := 0
	totalMembersDeleted := 0

	// Process each bucket: fetch members in chunks, delete from XDAS, then delete from Cassandra
	for _, bucketId := range populatedBuckets {
		membersDeleted, err := deleteBucketMembers(tenantId, tagId, bucketId)
		totalMembersDeleted += membersDeleted
		if err != nil {
			// Return error with partial progress saved
			return fmt.Errorf("partial deletion: %d/%d buckets deleted, %d members removed: %w",
				deletedBuckets, len(populatedBuckets), totalMembersDeleted, err)
		}

		deletedBuckets++
		if deletedBuckets%50 == 0 && deletedBuckets < len(populatedBuckets) {
			log.WithFields(fields).Infof("tag deletion progress: %d/%d buckets, %d members deleted",
				deletedBuckets, len(populatedBuckets), totalMembersDeleted)
		}
	}

	membersProcessed.WithLabelValues(OpDeleteTag).Add(float64(totalMembersDeleted))
	log.WithFields(fields).Infof("tag deletion completed: %d members removed from %d buckets in %v",
		totalMembersDeleted, deletedBuckets, time.Since(startTime).Round(time.Millisecond))
	return nil
}

// deleteBucketMembers deletes all members from a single bucket (XDAS first, then Cassandra)
// Returns number of members deleted
func deleteBucketMembers(tenantId string, tagId string, bucketId int) (int, error) {
	totalDeleted := 0
	lastMember := ""

	for {
		chunk, err := getMembersFromBucket(tenantId, tagId, bucketId, lastMember, MaxBatchSizeV2)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to fetch members from bucket: %w", err)
		}

		if len(chunk) == 0 {
			break
		}

		log.Debugf("Fetched %d members from bucket %d for tag '%s' (total deleted so far: %d)",
			len(chunk), bucketId, tagId, totalDeleted)

		removedFromXdas, _, firstError := removeMembersFromXDAS(tagId, chunk)

		if len(removedFromXdas) > 0 {
			// Delete successfully removed members from Cassandra
			removed, _, err := RemoveMembers(tenantId, tagId, removedFromXdas)
			totalDeleted += removed
			if err != nil {
				logDivergence(OpDeleteTag, tenantId, tagId, err)
				return totalDeleted, fmt.Errorf("cassandra deletion failed after XDAS success: %w", err)
			}
		}

		if len(removedFromXdas) < len(chunk) {
			// Partial XDAS failure must fail the bucket: returning success here
			// would let DeleteTag report a completed deletion while leftover
			// members and bucket metadata remain in both stores.
			return totalDeleted, fmt.Errorf("partial XDAS deletion in bucket %d: %d/%d members removed (first error: %s)",
				bucketId, len(removedFromXdas), len(chunk), firstError)
		}

		if len(chunk) < MaxBatchSizeV2 {
			break
		}

		lastMember = chunk[len(chunk)-1]
	}

	// All members deleted from this bucket, now delete bucket metadata
	if err := deleteBucketFromCassandra(tenantId, tagId, bucketId); err != nil {
		return totalDeleted, fmt.Errorf("failed to delete bucket metadata: %w", err)
	}

	return totalDeleted, nil
}

// deleteBucketFromCassandra deletes a bucket's metadata from Cassandra
func deleteBucketFromCassandra(tenantId string, tagId string, bucketId int) error {
	shardId := strconv.Itoa(ds.GetShardId(tagId))
	batch := ds.GetSimpleDao().NewBatch(UnloggedBatch)

	batch.Query(QueryDeleteBucketMembers, tenantId, tagId, strconv.Itoa(bucketId))
	batch.Query(QueryDeleteBucketMetadata, tenantId, shardId, tagId, strconv.Itoa(bucketId))

	if err := ds.GetSimpleDao().ExecuteBatch(batch); err != nil {
		return fmt.Errorf("batch execution failed: %w", err)
	}

	log.Debugf("Deleted bucket %d metadata for tag '%s'", bucketId, tagId)
	return nil
}

// GetMembersNonPaginated retrieves tag members for non-paginated response (V1 compatibility)
// Returns up to MaxMembersInTagResponse (100k) members as a plain array
func GetMembersNonPaginated(tenantId string, tagId string) ([]string, bool, ReadStats, error) {
	stats := ReadStats{}
	queries := &atomic.Int64{}

	populatedBuckets, err := getPopulatedBuckets(tenantId, tagId)
	queries.Add(1)
	stats.Queries = int(queries.Load())
	if err != nil {
		return nil, false, stats, fmt.Errorf("failed to get populated buckets: %w", err)
	}

	if len(populatedBuckets) == 0 {
		return nil, false, stats, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, tagId))
	}
	stats.Buckets = len(populatedBuckets)

	workers := getReadWorkerCount()
	collected, wasTruncated, err := fetchMembersFromBucketsConcurrent(
		tenantId, tagId, populatedBuckets, MaxMembersInTagResponse, workers, queries)
	stats.Queries = int(queries.Load())
	if err != nil {
		return nil, false, stats, err
	}

	log.Debugf("Tag '%s': retrieved %d members (non-paginated), truncated=%v", tagId, len(collected), wasTruncated)
	return collected, wasTruncated, stats, nil
}
