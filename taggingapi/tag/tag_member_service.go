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
	QueryGetMembersByBucketFirst = `SELECT member FROM "TagMembersBucketed" WHERE tag_id = ? AND bucket_id = ? LIMIT ?`

	QueryGetPopulatedBuckets = `SELECT bucket_id FROM "TagBucketMetadata" WHERE tag_id = ?`
	QueryAddBucketMetadata   = `INSERT INTO "TagBucketMetadata" (tag_id, bucket_id) VALUES (?, ?)`
	QueryGetAllTagIds        = `SELECT tag_id FROM "TagBucketMetadata"`

	// Typed variants, used only when tag_type_column_enabled. Kept separate so the
	// binary can run against a cluster without the ALTER applied: these share an
	// UnloggedBatch with the member inserts, so naming a missing column would fail
	// every add, not just account ones.
	QueryAddBucketMetadataTyped = `INSERT INTO "TagBucketMetadata" (tag_id, bucket_id, tag_type) VALUES (?, ?, ?)`
	QueryGetAllTagIdsTyped      = `SELECT tag_id, tag_type FROM "TagBucketMetadata"`
	QueryGetTagMetadataTyped    = `SELECT bucket_id, tag_type FROM "TagBucketMetadata" WHERE tag_id = ?`
	QueryDeleteBucketMembers    = `DELETE FROM "TagMembersBucketed" WHERE tag_id = ? AND bucket_id = ?`
	QueryDeleteBucketMetadata   = `DELETE FROM "TagBucketMetadata" WHERE tag_id = ? AND bucket_id = ?`

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

type bucketFetchResult struct {
	bucketIndex int
	members     []string
	err         error
}

func getBucketId(member string) int {
	hash := fnv.New32a()
	hash.Write([]byte(member))
	// Modulo in uint32 space — an int conversion first would go negative on
	// 32-bit platforms. The member→bucket mapping is recomputed against stored
	// rows on every read and write, so it must be identical everywhere.
	return int(hash.Sum32() % BucketCount)
}

// AddMembers writes members to the bucketed Cassandra tables. Returns the
// number of members stored and the number of buckets touched.
func AddMembers(tagId string, members []string, tagType string) (int, int, error) {
	if len(members) > MaxBatchSizeV2 {
		return 0, 0, fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	if len(members) == 0 {
		return 0, 0, fmt.Errorf("member list is empty")
	}

	bucketGroups := make(map[int][]string)
	for _, member := range members {
		bucketId := getBucketId(member)
		bucketGroups[bucketId] = append(bucketGroups[bucketId], member)
	}

	created := strconv.FormatInt(util.GetTimestamp(), 10)
	agg := &errorAggregator{}
	successCount := 0

	for bucketId, bucketMembers := range bucketGroups {
		if err := addMembersToBucket(tagId, bucketId, bucketMembers, created, tagType); err != nil {
			agg.add(fmt.Errorf("bucket %d: %w", bucketId, err))
		} else {
			successCount += len(bucketMembers)
			log.Debugf("Successfully added %d members to bucket %d for tag %s",
				len(bucketMembers), bucketId, tagId)
		}
	}

	// One aggregate ERROR line per call — a batch can span 1000 buckets, and an
	// outage must not emit an error line per bucket.
	if failedBuckets, firstError := agg.summary(); failedBuckets > 0 {
		log.Errorf("Failed to add members to %d/%d buckets for tag %s (first error: %s)",
			failedBuckets, len(bucketGroups), tagId, firstError)
		return successCount, len(bucketGroups), fmt.Errorf("failed to add %d/%d members (%d/%d buckets failed; first error: %s)",
			len(members)-successCount, len(members), failedBuckets, len(bucketGroups), firstError)
	}

	return successCount, len(bucketGroups), nil
}

func addMembersToBucket(tagId string, bucketId int, members []string, created string, tagType string) error {
	batch := ds.GetSimpleDao().NewBatch(UnloggedBatch)

	// Deliberately no tag_type on member rows: a text cell per member on the
	// hottest write path with no reader. The type belongs to the metadata row.
	for _, member := range members {
		batch.Query(QueryAddMemberBucketed, tagId, strconv.Itoa(bucketId), member, created)
	}

	if tagTypeColumnEnabled() {
		// Stored type domain is {"", "account"}: mac and legacy are one
		// equivalence class everywhere, so storing "mac" would only mint a third
		// value that every reader collapses anyway.
		if tagType == TagTypeMac {
			tagType = TagTypeLegacy
		}
		batch.Query(QueryAddBucketMetadataTyped, tagId, strconv.Itoa(bucketId), tagType)
	} else {
		batch.Query(QueryAddBucketMetadata, tagId, strconv.Itoa(bucketId))
	}

	return ds.GetSimpleDao().ExecuteBatch(batch)
}

// RemoveMembers deletes members from the bucketed Cassandra tables. Returns
// the number of members removed and the number of buckets touched.
func RemoveMembers(tagId string, members []string) (int, int, error) {
	if len(members) > MaxBatchSizeV2 {
		return 0, 0, fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	if len(members) == 0 {
		return 0, 0, fmt.Errorf("member list is empty")
	}

	bucketGroups := make(map[int][]string)
	for _, member := range members {
		bucketId := getBucketId(member)
		bucketGroups[bucketId] = append(bucketGroups[bucketId], member)
	}

	agg := &errorAggregator{}
	successCount := 0

	// Deliberately no empty-bucket metadata cleanup: deleting a metadata row on a
	// zero count races a concurrent add (which writes member + metadata in one
	// batch), orphaning members that no metadata row points to — invisible to
	// reads and to DeleteTag. Empty metadata rows are harmless: reads skip them
	// and DeleteTag reaps them.
	for bucketId, bucketMembers := range bucketGroups {
		if err := removeMembersFromBucket(tagId, bucketId, bucketMembers); err != nil {
			agg.add(fmt.Errorf("bucket %d: %w", bucketId, err))
			continue
		}
		successCount += len(bucketMembers)
		log.Debugf("Successfully removed %d members from bucket %d for tag %s",
			len(bucketMembers), bucketId, tagId)
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

func removeMembersFromBucket(tagId string, bucketId int, members []string) error {
	batch := ds.GetSimpleDao().NewBatch(UnloggedBatch)

	for _, member := range members {
		batch.Query(QueryRemoveMemberBucketed, tagId, strconv.Itoa(bucketId), member)
	}

	return ds.GetSimpleDao().ExecuteBatch(batch)
}

// tagTypeColumnEnabled reports whether the tag_type column may be referenced.
// False when config is absent, so tests and partially-migrated clusters take
// the untyped path.
func tagTypeColumnEnabled() bool {
	config := GetTagApiConfig()
	return config != nil && config.TagTypeColumnEnabled
}

// getPopulatedBuckets resolves a tag's populated buckets, scoped to the
// requested type. A type mismatch reports no buckets, which callers surface as
// 404 — the id is absent from that namespace, matching the listing filter. An
// untyped requestedType matches everything.
func getPopulatedBuckets(tagId string, requestedType string) ([]int, error) {
	buckets, storedType, err := getTagMeta(tagId)
	if err != nil {
		return nil, err
	}
	if !matchesTagTypeFilter(storedType, requestedType) {
		return nil, nil
	}
	return buckets, nil
}

// getTagMeta reads a tag's metadata partition once, returning both its populated
// buckets and its resolved type, so callers needing the type pay no second round
// trip. A tag with no metadata rows does not exist yet: empty buckets and
// TagTypeLegacy, which callers treat as "unclaimed".
func getTagMeta(tagId string) ([]int, string, error) {
	typed := tagTypeColumnEnabled()
	query := QueryGetPopulatedBuckets
	if typed {
		query = QueryGetTagMetadataTyped
	}

	rows, err := ds.GetSimpleDao().Query(query, tagId)
	if err != nil {
		return nil, TagTypeLegacy, err
	}

	buckets := make([]int, 0, len(rows))
	rowTypes := make([]string, 0, len(rows))
	for _, row := range rows {
		if bucketId, ok := row["bucket_id"].(int); ok {
			buckets = append(buckets, bucketId)
		}
		if typed {
			rowTypes = append(rowTypes, tagTypeFromRow(row))
		}
	}

	return buckets, resolveTagType(rowTypes), nil
}

// tagTypeFromRow reads tag_type from a Cassandra row. A NULL text column comes
// back as a missing key or an empty string; both mean legacy.
func tagTypeFromRow(row map[string]interface{}) string {
	if value, ok := row["tag_type"].(string); ok {
		return value
	}
	return TagTypeLegacy
}

// resolveTagType collapses one tag's per-bucket tag_type values into a single
// answer. Rows can disagree when two concurrent first-ever adds of the same id
// claim different types; account wins, so the tag reads consistently and the
// next mac write gets a clean conflict instead of flapping. Anything
// non-account reads as legacy, including a stored "mac" from rows written
// before addMembersToBucket normalized it away.
func resolveTagType(rowTypes []string) string {
	for _, rowType := range rowTypes {
		if rowType == TagTypeAccount {
			return TagTypeAccount
		}
	}
	return TagTypeLegacy
}

// ensureTagTypeCompatible rejects writes that would give one tag id two types.
// tag_id is the Cassandra partition key, so a mac and an account tag sharing an
// id would share partitions, and reads, pagination and deletes would each
// operate on a mixture of MACs and account ids. Legacy requests never conflict:
// effectiveWriteTagType resolves them to the stored type before calling this.
func ensureTagTypeCompatible(tagId string, requestedType string) error {
	if requestedType == TagTypeLegacy || !tagTypeColumnEnabled() {
		return nil
	}

	buckets, existing, err := getTagMeta(tagId)
	if err != nil {
		// Fail closed: an unreadable type is not permission to overwrite it.
		return err
	}

	// No metadata rows means the tag does not exist, so any write may claim the
	// id. Decided on the bucket count, not the type: a stored legacy type and an
	// absent tag are the same empty string, and conflating them rejects every
	// first write of a new account tag as a conflict with a tag never there.
	if len(buckets) == 0 {
		return nil
	}

	return checkTagTypeCompatible(tagId, existing, requestedType)
}

// checkTagTypeCompatible is the comparison half of ensureTagTypeCompatible,
// split out so callers that already resolved the stored type (the delete
// handler) do not re-read the same partition. existing must be a resolveTagType
// result for a tag that exists.
func checkTagTypeCompatible(tagId string, existing string, requestedType string) error {
	if existing == requestedType {
		return nil
	}
	// Mac and legacy are the same equivalence class, so a mac write adopts a
	// legacy tag.
	if existing == TagTypeLegacy && requestedType == TagTypeMac {
		return nil
	}
	// Untyped rows on an existing tag mean a mac tag predating the type column,
	// so its members are MACs.
	if existing == TagTypeLegacy && requestedType == TagTypeAccount {
		return xwcommon.NewRemoteErrorAS(http.StatusConflict,
			fmt.Sprintf("tag '%s' already exists as a mac tag and cannot be reused as an account tag", tagId))
	}
	return xwcommon.NewRemoteErrorAS(http.StatusConflict,
		fmt.Sprintf("tag '%s' already exists with type '%s'", tagId, existing))
}

// effectiveWriteTagType decides which type a member write executes as: the
// requested type for typed routes (after the compatibility check), the stored
// type for untyped ones.
//
// Adopting the stored type keeps an untyped write to an account tag from
// splitting it across stores — normalizing its members as MACs would send them
// to the device keyspace while their Cassandra rows land in the account tag's
// partitions, and DeleteTag would then wedge on the mixture. Untyped writes to
// mac/legacy tags and to tags that do not exist resolve to legacy, so untyped
// behavior is unchanged; the cost is one metadata read per untyped batch.
func effectiveWriteTagType(tagId string, requestedType string) (string, error) {
	if !tagTypeColumnEnabled() {
		return requestedType, nil
	}
	if requestedType != TagTypeLegacy {
		return requestedType, ensureTagTypeCompatible(tagId, requestedType)
	}
	_, stored, err := getTagMeta(tagId)
	if err != nil {
		// Fail closed: an unreadable type must not default to the device keyspace.
		return requestedType, err
	}
	return stored, nil
}

// GetMembersPaginated returns one page of members plus the Cassandra cost of
// producing it (for the request log).
func GetMembersPaginated(tagId string, limit int, cursor string, tagType string) (*PaginatedMembersResponse, ReadStats, error) {
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

	populatedBuckets, err := getPopulatedBuckets(tagId, tagType)
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

	// First populated bucket at or after the cursor. If none exists (buckets
	// emptied since the previous page), enumeration is complete — never wrap.
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

	workers := getReadWorkerCount()
	remainingBuckets := populatedBuckets[startIndex:]

	// Fetch in windows of `workers` buckets, in order, stopping once the page is
	// full — building work items for every remaining bucket up front cost ~N
	// queries per page even when the first bucket satisfied it. The cursor's
	// lastMember applies only to the bucket the cursor points into.
	lastProcessedBucketIndex := startIndex - 1
	for windowStart := 0; windowStart < len(remainingBuckets) && len(allMembers) < limit; windowStart += workers {
		window := remainingBuckets[windowStart:min(windowStart+workers, len(remainingBuckets))]
		workItems := make([]bucketWorkItem, len(window))
		for idx, bucketId := range window {
			lm := ""
			if windowStart == 0 && idx == 0 && bucketId == state.BucketId {
				lm = state.LastMember
			}
			workItems[idx] = bucketWorkItem{
				bucketId:   bucketId,
				lastMember: lm,
				// One over the remaining space, so an overfull bucket is
				// detectable for mid-bucket cursor placement.
				limit: limit - len(allMembers) + 1,
			}
		}

		orderedResults := fetchBucketsConcurrent(tagId, workItems, workers, queries)
		stats.Queries = int(queries.Load())

		// Merge in bucket order, building the cursor at the truncation point. A
		// failed bucket fails the whole page: partial data would let the cursor
		// advance past it and silently drop its members from the enumeration.
		for idx, result := range orderedResults {
			if result.err != nil {
				return nil, stats, fmt.Errorf("failed to fetch members from bucket %d: %w", window[idx], result.err)
			}
			if len(result.members) == 0 {
				lastProcessedBucketIndex = startIndex + windowStart + idx
				continue
			}

			currentBucketId := window[idx]
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
			lastProcessedBucketIndex = startIndex + windowStart + idx

			if len(allMembers) >= limit {
				break
			}
		}
	}

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

func getReadWorkerCount() int {
	config := GetTagApiConfig()
	if config != nil && config.WorkerCount > 0 {
		return min(config.WorkerCount, MaxWorkersV2)
	}
	return 1
}

// getWriteWorkerCount sizes the concurrent XDAS write phase: the configured
// count, scaled up for large batches (one worker per 100 members), capped at
// MaxWorkersV2 and the batch size, floored at one. The floor is load-bearing —
// with zero workers nothing drains the member channel, so the write reports
// XdasOk=0/XdasFail=0 and the request becomes a silent no-op 202.
func getWriteWorkerCount(memberCount int) int {
	config := GetTagApiConfig()
	if config == nil {
		return 1
	}
	scaledWorkers := min(max(memberCount/100, config.WorkerCount), MaxWorkersV2)
	return max(1, min(scaledWorkers, memberCount))
}

// fetchBucketMembersWithLimit fetches a single bucket's members in chunks,
// counting issued queries into the shared counter.
func fetchBucketMembersWithLimit(tagId string, bucketId int, lastMember string, limit int, queries *atomic.Int64) ([]string, error) {
	collected := make([]string, 0)

	for {
		remainingCapacity := limit - len(collected)
		if remainingCapacity <= 0 {
			break
		}

		chunkLimit := min(MemberFetchChunkSize, remainingCapacity)
		queries.Add(1)
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

type bucketWorkItem struct {
	bucketId   int
	lastMember string
	limit      int
}

// fetchBucketsConcurrent fetches multiple buckets through a worker pool and
// returns ordered results (one per bucket) without merging.
func fetchBucketsConcurrent(tagId string, workItems []bucketWorkItem, workers int, queries *atomic.Int64) []bucketFetchResult {
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
				members, err := fetchBucketMembersWithLimit(tagId, work.bucketId, work.lastMember, work.limit, queries)
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

// fetchMembersFromBucketsConcurrent fetches buckets in windows of `workers`, in
// bucket order, and returns a merged, truncated result. The windowing is
// load-bearing: fetching every populated bucket eagerly buffered a whole tag in
// memory — gigabytes of heap per request — before the merge truncated it.
// Dispatch stops once the response is full, over-fetching at most one window.
func fetchMembersFromBucketsConcurrent(tagId string, bucketIds []int, totalLimit int, workers int, queries *atomic.Int64) ([]string, bool, error) {
	if len(bucketIds) == 0 {
		return nil, false, nil
	}
	if workers < 1 {
		// The window advances by `workers`; a non-positive count would loop forever.
		workers = 1
	}

	// Merge in bucket order, stopping at totalLimit. A failed bucket fails the
	// whole read — a partial result would silently under-report membership.
	// Errors past the fill point are irrelevant, the response is already full.
	collected := make([]string, 0)
	for windowStart := 0; windowStart < len(bucketIds); windowStart += workers {
		remaining := totalLimit - len(collected)
		if remaining <= 0 {
			return collected, true, nil
		}

		window := bucketIds[windowStart:min(windowStart+workers, len(bucketIds))]
		workItems := make([]bucketWorkItem, len(window))
		for idx, bucketId := range window {
			// Any one bucket may have to supply the whole rest of the response.
			workItems[idx] = bucketWorkItem{
				bucketId:   bucketId,
				lastMember: "",
				limit:      remaining,
			}
		}

		orderedResults := fetchBucketsConcurrent(tagId, workItems, workers, queries)

		for idx, result := range orderedResults {
			space := totalLimit - len(collected)
			if space <= 0 {
				return collected, true, nil
			}
			if result.err != nil {
				return nil, false, fmt.Errorf("failed to fetch members from bucket %d: %w", window[idx], result.err)
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
	}

	wasTruncated := len(collected) >= totalLimit
	return collected, wasTruncated, nil
}

// AddMembersWithXdas adds members to XDAS first, then Cassandra. The returned
// WriteStats carry per-store outcome counts for the request log, on error paths
// too; they land on the framework "request ends" line.
func AddMembersWithXdas(tagId string, members []string, tagValue string, tagType string) (WriteStats, error) {
	stats := WriteStats{Requested: len(members)}

	if len(members) == 0 {
		return stats, fmt.Errorf("member list is empty")
	}

	if len(members) > MaxBatchSizeV2 {
		return stats, fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	// Must run before the XDAS phase. A conflict detected afterwards would leave
	// members in a keyspace with no Cassandra record pointing at them, and XDAS
	// cannot be enumerated to find them again.
	tagType, err := effectiveWriteTagType(tagId, tagType)
	if err != nil {
		return stats, err
	}
	stats.TagType = tagType

	savedToXdasMembers, xdasFail, firstError := addMembersToXdas(tagId, members, tagValue, tagType)
	stats.XdasOk = len(savedToXdasMembers)
	stats.XdasFail = xdasFail
	stats.FirstError = firstError

	// Every XDAS write failed. Falling through to a 202 with stored=0 would make
	// a misconfigured keyspace look like a healthy API that stored nothing.
	if stats.XdasOk == 0 && stats.XdasFail > 0 {
		return stats, xwcommon.NewRemoteErrorAS(http.StatusBadGateway,
			fmt.Sprintf("all %d XDAS writes failed: %s", stats.XdasFail, firstError))
	}

	if stats.XdasOk > 0 {
		stored, buckets, err := AddMembers(tagId, savedToXdasMembers, tagType)
		stats.CassandraOk = stored
		stats.CassandraFail = stats.XdasOk - stored
		stats.Buckets = buckets
		if err != nil {
			if stats.FirstError == "" {
				stats.FirstError = err.Error()
			}
			logDivergence(OpAddMembers, tagId, err)
			return stats, fmt.Errorf("cassandra V2 storage failed after XDAS success: %w", err)
		}
	}

	return stats, nil
}

// RemoveMembersWithXdas removes members from XDAS first, then Cassandra. See
// AddMembersWithXdas for the stats contract.
func RemoveMembersWithXdas(tagId string, members []string, tagType string) (WriteStats, error) {
	stats := WriteStats{Requested: len(members)}

	if len(members) == 0 {
		return stats, fmt.Errorf("member list is empty")
	}

	if len(members) > MaxBatchSizeV2 {
		return stats, fmt.Errorf("batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)
	}

	// Without this, DELETE /tags/account/{macTag}/members would delete from the
	// account keyspace for a tag whose members live in the device one.
	tagType, err := effectiveWriteTagType(tagId, tagType)
	if err != nil {
		return stats, err
	}
	stats.TagType = tagType

	successfulRemovals, xdasFail, firstError := removeMembersFromXDAS(tagId, members, tagType)
	stats.XdasOk = len(successfulRemovals)
	stats.XdasFail = xdasFail
	stats.FirstError = firstError

	if stats.XdasOk == 0 && stats.XdasFail > 0 {
		return stats, xwcommon.NewRemoteErrorAS(http.StatusBadGateway,
			fmt.Sprintf("all %d XDAS removals failed: %s", stats.XdasFail, firstError))
	}

	if stats.XdasOk > 0 {
		removed, buckets, err := RemoveMembers(tagId, successfulRemovals)
		stats.CassandraOk = removed
		stats.CassandraFail = stats.XdasOk - removed
		stats.Buckets = buckets
		if err != nil {
			if stats.FirstError == "" {
				stats.FirstError = err.Error()
			}
			logDivergence(OpRemoveMembers, tagId, err)
			return stats, fmt.Errorf("cassandra V2 removal failed after XDAS success: %w", err)
		}
	}

	return stats, nil
}

func RemoveMemberWithXdas(tagId string, member string, tagType string) (WriteStats, error) {
	return RemoveMembersWithXdas(tagId, []string{member}, tagType)
}

// addMembersToXdas adds members through concurrent workers, returning the saved
// members, the failure count and the first error message. Partial failure is
// reported through the counts, not an error — the caller decides how to surface it.
func addMembersToXdas(tagId string, members []string, tagValue string, tagType string) ([]string, int, string) {
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

	numOfWorkers := getWriteWorkerCount(len(members))
	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go storeTagMembersInXdas(tagId, membersChannel, savedMembersChannel, wg, tagValue, tagType, errAgg)
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
func removeMembersFromXDAS(tagId string, members []string, tagType string) ([]string, int, string) {
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

	numOfWorkers := getWriteWorkerCount(len(members))
	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go removeTagMembersFromXdas(tagId, membersChannel, removedMembersChannel, wg, tagType, agg)
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

// GetAllTagIds lists distinct tag ids; an empty tagTypeFilter returns every tag.
// Filtering happens in Go over the existing full-table scan: tag_type has two
// distinct values across ~105K rows, so a secondary index would hotspot two
// partitions and ALLOW FILTERING would be worse than the scan we already pay.
func GetAllTagIds(tagTypeFilter string) ([]string, error) {
	typed := tagTypeColumnEnabled()
	query := QueryGetAllTagIds
	if typed {
		query = QueryGetAllTagIdsTyped
	}

	rows, err := ds.GetSimpleDao().Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tag IDs: %w", err)
	}

	// One row per (tag_id, bucket_id), so a tag appears up to BucketCount times
	// and its per-row types must be collapsed before filtering.
	rowTypes := make(map[string][]string)
	for _, row := range rows {
		tagId, ok := row["tag_id"].(string)
		if !ok {
			continue
		}
		if typed {
			rowTypes[tagId] = append(rowTypes[tagId], tagTypeFromRow(row))
		} else if _, seen := rowTypes[tagId]; !seen {
			rowTypes[tagId] = nil
		}
	}

	tagIds := make([]string, 0, len(rowTypes))
	for tagId, types := range rowTypes {
		if matchesTagTypeFilter(resolveTagType(types), tagTypeFilter) {
			tagIds = append(tagIds, tagId)
		}
	}

	log.Debugf("Retrieved %d unique tag IDs from V2 storage (filter=%q)", len(tagIds), tagTypeFilter)
	return tagIds, nil
}

// matchesTagTypeFilter reports whether a resolved tag type satisfies a filter.
// A mac filter also matches legacy rows: legacy tags are mac tags that predate
// the type column.
func matchesTagTypeFilter(resolved string, filter string) bool {
	switch filter {
	case TagTypeLegacy:
		return true
	case TagTypeMac:
		return resolved == TagTypeLegacy || resolved == TagTypeMac
	default:
		return resolved == filter
	}
}

// GetTagById retrieves a tag with up to MaxMembersInTagResponse members
func GetTagById(tagId string, tagType string) ([]string, bool, ReadStats, error) {
	stats := ReadStats{}
	queries := &atomic.Int64{}

	populatedBuckets, err := getPopulatedBuckets(tagId, tagType)
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
		tagId, populatedBuckets, MaxMembersInTagResponse, workers, queries)
	stats.Queries = int(queries.Load())
	if err != nil {
		return nil, false, stats, err
	}

	log.Debugf("Tag '%s': retrieved %d members, truncated=%v", tagId, len(collected), wasTruncated)
	return collected, wasTruncated, stats, nil
}

// DeleteTag deletes a tag from XDAS and Cassandra in memory-safe chunks, for
// tags with millions of members. Runs in a background goroutine, so it logs its
// own start/progress/end lines carrying the queuing request's audit_id.
func DeleteTag(tagId string, auditId string) error {
	fields := log.Fields{
		"audit_id": auditId,
		"op":       OpDeleteTag,
		"tag":      tagId,
	}

	// Type resolved from storage, never from the request: deleting an account tag
	// through the untyped route must still hit the account keyspace, or the
	// Cassandra rows go and the XDAS entries are orphaned beyond recovery.
	populatedBuckets, tagType, err := getTagMeta(tagId)
	if err != nil {
		return fmt.Errorf("failed to get populated buckets: %w", err)
	}

	if len(populatedBuckets) == 0 {
		return fmt.Errorf("tag not found")
	}
	if tagType != TagTypeLegacy {
		fields["tag_type"] = tagType
	}

	startTime := time.Now()
	log.WithFields(fields).Infof("tag deletion started: %d buckets", len(populatedBuckets))

	deletedBuckets := 0
	totalMembersDeleted := 0

	for _, bucketId := range populatedBuckets {
		membersDeleted, err := deleteBucketMembers(tagId, bucketId, tagType)
		totalMembersDeleted += membersDeleted
		if err != nil {
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

// deleteBucketMembers deletes a bucket's members, XDAS first, and returns how
// many were deleted.
func deleteBucketMembers(tagId string, bucketId int, tagType string) (int, error) {
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

		removedFromXdas, _, firstError := removeMembersFromXDAS(tagId, chunk, tagType)

		if len(removedFromXdas) > 0 {
			removed, _, err := RemoveMembers(tagId, removedFromXdas)
			totalDeleted += removed
			if err != nil {
				logDivergence(OpDeleteTag, tagId, err)
				return totalDeleted, fmt.Errorf("cassandra deletion failed after XDAS success: %w", err)
			}
		}

		if len(removedFromXdas) < len(chunk) {
			// Partial XDAS failure fails the bucket: reporting success would let
			// DeleteTag claim completion while members remain in both stores.
			return totalDeleted, fmt.Errorf("partial XDAS deletion in bucket %d: %d/%d members removed (first error: %s)",
				bucketId, len(removedFromXdas), len(chunk), firstError)
		}

		if len(chunk) < MaxBatchSizeV2 {
			break
		}

		lastMember = chunk[len(chunk)-1]
	}

	if err := deleteBucketFromCassandra(tagId, bucketId); err != nil {
		return totalDeleted, fmt.Errorf("failed to delete bucket metadata: %w", err)
	}

	return totalDeleted, nil
}

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

// GetMembersNonPaginated returns up to MaxMembersInTagResponse members as a
// plain array (V1 compatibility).
func GetMembersNonPaginated(tagId string, tagType string) ([]string, bool, ReadStats, error) {
	stats := ReadStats{}
	queries := &atomic.Int64{}

	populatedBuckets, err := getPopulatedBuckets(tagId, tagType)
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
		tagId, populatedBuckets, MaxMembersInTagResponse, workers, queries)
	stats.Queries = int(queries.Load())
	if err != nil {
		return nil, false, stats, err
	}

	log.Debugf("Tag '%s': retrieved %d members (non-paginated), truncated=%v", tagId, len(collected), wasTruncated)
	return collected, wasTruncated, stats, nil
}
