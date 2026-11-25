package tag

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/db"

	"github.com/stretchr/testify/assert"
)

// fakeTagStore serves one tag's buckets and records which were read.
type fakeTagStore struct {
	buckets     map[int][]string // members sorted within a bucket
	failBuckets map[int]bool

	mu          sync.Mutex
	queries     int
	bucketsRead map[int]bool
}

func newFakeTagStore(buckets map[int][]string) *fakeTagStore {
	return &fakeTagStore{buckets: buckets, bucketsRead: map[int]bool{}}
}

// uniformTagStore builds bucketCount buckets of perBucket members each.
func uniformTagStore(bucketCount int, perBucket int) *fakeTagStore {
	buckets := make(map[int][]string, bucketCount)
	for b := 0; b < bucketCount; b++ {
		members := make([]string, perBucket)
		for i := range members {
			members[i] = fmt.Sprintf("b%03d-m%05d", b, i)
		}
		buckets[b] = members
	}
	return newFakeTagStore(buckets)
}

func (s *fakeTagStore) install(t *testing.T) {
	t.Helper()
	setupTestEnvironment()
	withMockDb(t, s.query)
}

func (s *fakeTagStore) query(query string, params ...string) ([]map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queries++

	if isMetadataQuery(query) {
		ids := make([]int, 0, len(s.buckets))
		for id := range s.buckets {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		rows := make([]map[string]any, len(ids))
		for i, id := range ids {
			rows[i] = map[string]any{"bucket_id": id}
		}
		return rows, nil
	}

	bucketId, _ := strconv.Atoi(params[2])
	s.bucketsRead[bucketId] = true
	if s.failBuckets[bucketId] {
		return nil, errors.New("read timeout")
	}

	members := s.buckets[bucketId]
	if query == QueryGetMembersByBucket {
		after := params[3]
		members = members[sort.SearchStrings(members, after+"\x00"):]
	}
	limit, _ := strconv.Atoi(params[len(params)-1])
	if len(members) > limit {
		members = members[:limit]
	}
	rows := make([]map[string]any, len(members))
	for i, member := range members {
		rows[i] = map[string]any{"member": member}
	}
	return rows, nil
}

// expected returns every member in the order a full walk delivers them.
func (s *fakeTagStore) expected() []string {
	ids := make([]int, 0, len(s.buckets))
	for id := range s.buckets {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	var all []string
	for _, id := range ids {
		all = append(all, s.buckets[id]...)
	}
	return all
}

func (s *fakeTagStore) readCount() (queries int, buckets int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queries, len(s.bucketsRead)
}

func walkAllPages(t *testing.T, limit int) []string {
	t.Helper()
	var all []string
	cursor := ""
	for page := 0; ; page++ {
		if page > 100000 {
			t.Fatal("pagination did not terminate")
		}
		resp, _, err := GetMembersPaginated(db.GetDefaultTenantId(), "some-tag", limit, cursor)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		assert.LessOrEqual(t, len(resp.Data), limit)
		all = append(all, resp.Data...)
		if !resp.HasMore {
			return all
		}
		cursor = resp.NextCursor
	}
}

// A page reads about as many buckets as it needs, not the rest of the tag.
func TestGetMembersPaginated_ReadsOnlyBucketsThePageNeeds(t *testing.T) {
	store := uniformTagStore(BucketCount, 100)
	store.install(t)

	resp, stats, err := GetMembersPaginated(db.GetDefaultTenantId(), "some-tag", 1000, "")
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 1000)
	assert.True(t, resp.HasMore)

	queries, bucketsRead := store.readCount()
	neededBuckets := 10
	assert.LessOrEqual(t, bucketsRead, neededBuckets+getReadWorkerCount())
	assert.Equal(t, queries, stats.Queries)
}

// A small page of a dense tag is served from its first bucket alone.
func TestGetMembersPaginated_SmallPageReadsOneBucket(t *testing.T) {
	store := uniformTagStore(BucketCount, 100)
	store.install(t)

	resp, _, err := GetMembersPaginated(db.GetDefaultTenantId(), "some-tag", 50, "")
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 50)

	_, bucketsRead := store.readCount()
	assert.Equal(t, 1, bucketsRead)
}

func TestGetMembersPaginated_FullWalkReturnsEachMemberOnce(t *testing.T) {
	uneven := map[int][]string{
		0:   {"a1", "a2", "a3"},
		3:   {},
		4:   {"b1"},
		17:  {"c1", "c2", "c3", "c4", "c5", "c6", "c7"},
		18:  {},
		500: {"d1", "d2"},
		999: {},
	}
	shapes := map[string]func() *fakeTagStore{
		"uneven with empty buckets": func() *fakeTagStore { return newFakeTagStore(uneven) },
		"uniform":                   func() *fakeTagStore { return uniformTagStore(60, 25) },
		"sparse":                    func() *fakeTagStore { return uniformTagStore(200, 1) },
	}

	for name, build := range shapes {
		for _, limit := range []int{1, 2, 7, 25, 50, 333, 100000} {
			t.Run(fmt.Sprintf("%s limit %d", name, limit), func(t *testing.T) {
				store := build()
				store.install(t)
				assert.Equal(t, store.expected(), walkAllPages(t, limit))
			})
		}
	}
}

// A failing bucket the page does not reach must not fail the page.
func TestGetMembersPaginated_FailureBeyondThePageIsNotRead(t *testing.T) {
	store := uniformTagStore(BucketCount, 100)
	store.failBuckets = map[int]bool{BucketCount - 1: true}
	store.install(t)

	resp, _, err := GetMembersPaginated(db.GetDefaultTenantId(), "some-tag", 1000, "")
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 1000)
}

func TestGetMembersPaginated_FailureInsideThePageFailsIt(t *testing.T) {
	store := uniformTagStore(BucketCount, 100)
	store.failBuckets = map[int]bool{5: true}
	store.install(t)

	resp, _, err := GetMembersPaginated(db.GetDefaultTenantId(), "some-tag", 1000, "")
	assert.Nil(t, resp)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bucket 5")
	}
}

// The non-paginated reads stop at the response cap.
func TestGetMembersNonPaginated_StopsAtCap(t *testing.T) {
	store := uniformTagStore(BucketCount, 200)
	store.install(t)

	members, truncated, _, err := GetMembersNonPaginated(db.GetDefaultTenantId(), "some-tag")
	assert.NoError(t, err)
	assert.True(t, truncated)
	assert.Equal(t, store.expected()[:MaxMembersInTagResponse], members)

	_, bucketsRead := store.readCount()
	neededBuckets := MaxMembersInTagResponse / 200
	assert.LessOrEqual(t, bucketsRead, neededBuckets+getReadWorkerCount())
}

func TestGetTagById_UnderCapReturnsAllMembers(t *testing.T) {
	store := uniformTagStore(300, 10)
	store.install(t)

	members, truncated, _, err := GetTagById(db.GetDefaultTenantId(), "some-tag")
	assert.NoError(t, err)
	assert.False(t, truncated)
	assert.Equal(t, store.expected(), members)
}
