package tag

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"

	taggingapi_config "github.com/rdkcentral/xconfadmin/taggingapi/config"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	"github.com/stretchr/testify/assert"
)

func init() {
	// The lock settle pause exists for real Cassandra races; keep tests fast.
	tagSyncLockSettle = time.Millisecond
}

// ---- fakes ----

type fakeTagSyncDao struct {
	mu         sync.Mutex
	runs       map[string]*TagSyncRun
	lock       *TagSyncLock
	lockWrites []time.Time
	getLockErr error
}

func newFakeTagSyncDao() *fakeTagSyncDao {
	return &fakeTagSyncDao{runs: make(map[string]*TagSyncRun)}
}

func (s *fakeTagSyncDao) saveRun(run *TagSyncRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := *run
	s.runs[run.RunId] = &copied
	return nil
}

func (s *fakeTagSyncDao) getRun(runId string) (*TagSyncRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if run, ok := s.runs[runId]; ok {
		copied := *run
		return &copied, nil
	}
	return nil, nil
}

func (s *fakeTagSyncDao) listRuns(limit int) ([]*TagSyncRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	runs := make([]*TagSyncRun, 0, len(s.runs))
	for _, run := range s.runs {
		copied := *run
		runs = append(runs, &copied)
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].RunId > runs[j].RunId })
	if len(runs) > limit {
		runs = runs[:limit]
	}
	return runs, nil
}

func (s *fakeTagSyncDao) pruneRuns(keep int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.runs) <= keep {
		return nil
	}
	ids := make([]string, 0, len(s.runs))
	for id := range s.runs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids[:len(ids)-keep] {
		delete(s.runs, id)
	}
	return nil
}

func (s *fakeTagSyncDao) getLock() (*TagSyncLock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.getLockErr != nil {
		return nil, s.getLockErr
	}
	if s.lock == nil {
		return nil, nil
	}
	copied := *s.lock
	return &copied, nil
}

func (s *fakeTagSyncDao) saveLock(lock *TagSyncLock) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := *lock
	s.lock = &copied
	s.lockWrites = append(s.lockWrites, time.Now())
	return nil
}

func (s *fakeTagSyncDao) lockWriteTimes() []time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]time.Time(nil), s.lockWrites...)
}

type pushRec struct {
	member string
	tag    string
	value  string
}

type fakeXdas struct {
	mu      sync.Mutex
	records map[string]map[string]string
	getErr  map[string]error
	gets    []string
	pushes  []pushRec
}

func newFakeXdas() *fakeXdas {
	return &fakeXdas{
		records: make(map[string]map[string]string),
		getErr:  make(map[string]error),
	}
}

func (x *fakeXdas) getFields(member string) (map[string]string, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.gets = append(x.gets, member)
	if err, ok := x.getErr[member]; ok {
		return nil, err
	}
	fields, ok := x.records[member]
	if !ok {
		return nil, xwcommon.NewRemoteErrorAS(404, "not found")
	}
	copied := make(map[string]string, len(fields))
	for k, v := range fields {
		copied[k] = v
	}
	return copied, nil
}

func (x *fakeXdas) push(member string, tag string, value string) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.pushes = append(x.pushes, pushRec{member: member, tag: tag, value: value})
	return nil
}

func (x *fakeXdas) pushCount() int {
	x.mu.Lock()
	defer x.mu.Unlock()
	return len(x.pushes)
}

func (x *fakeXdas) getCount() int {
	x.mu.Lock()
	defer x.mu.Unlock()
	return len(x.gets)
}

func testSyncConfig() *taggingapi_config.TagSyncConfig {
	return &taggingapi_config.TagSyncConfig{
		RateLimit:               100000,
		WorkerCount:             4,
		ChunkSize:               100,
		CheckpointIntervalSecs:  0, // persist on every chunk
		BreakerWindow:           50,
		BreakerMinSample:        20,
		BreakerErrorRatePercent: 25,
		BreakerMaxConsecErrors:  5,
	}
}

// newTestEnv wires the engine to an in-memory Cassandra view (tagId ->
// members) and the fake XDAS. Bucket layout follows the real getBucketId.
func newTestEnv(cass map[string][]string, xdas *fakeXdas, dao *fakeTagSyncDao) *tagSyncEnv {
	return &tagSyncEnv{
		getAllTagIds: func() ([]string, error) {
			tags := make([]string, 0, len(cass))
			for tagId := range cass {
				tags = append(tags, tagId)
			}
			return tags, nil
		},
		getPopulatedBuckets: func(tagId string) ([]int, error) {
			seen := make(map[int]bool)
			buckets := []int{}
			for _, member := range cass[tagId] {
				b := getBucketId(member)
				if !seen[b] {
					seen[b] = true
					buckets = append(buckets, b)
				}
			}
			return buckets, nil
		},
		getMembersFromBucket: func(tagId string, bucketId int, lastMember string, limit int) ([]string, error) {
			members := []string{}
			for _, member := range cass[tagId] {
				if getBucketId(member) == bucketId && member > lastMember {
					members = append(members, member)
				}
			}
			sort.Strings(members)
			if len(members) > limit {
				members = members[:limit]
			}
			return members, nil
		},
		xdasGetFields: xdas.getFields,
		xdasPush:      xdas.push,
		dao:           dao,
		syncEnabled:   func() bool { return true },
		config:        testSyncConfig(),
	}
}

func members(prefix string, n int) []string {
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = fmt.Sprintf("%s-%04d", prefix, i)
	}
	return out
}

// membersInOneBucket returns n member names that all hash to the same
// bucket, so a test can build one multi-member chunk deterministically.
func membersInOneBucket(prefix string, n int) []string {
	target := -1
	out := make([]string, 0, n)
	for i := 0; len(out) < n; i++ {
		m := fmt.Sprintf("%s-%06d", prefix, i)
		b := getBucketId(m)
		if target == -1 {
			target = b
		}
		if b == target {
			out = append(out, m)
		}
	}
	return out
}

func execute(t *testing.T, opts TagSyncOptions, env *tagSyncEnv) *TagSyncRun {
	t.Helper()
	engine, err := prepareTagSync(opts, env)
	assert.NoError(t, err)
	return engine.Execute(context.Background())
}

// ---- tests ----

func TestTagSyncDetectClassification(t *testing.T) {
	// tag1: m0 present with value, m1 has a record without the field, m2 has
	// no record at all. Detect must classify all three and push nothing.
	cass := map[string][]string{"tag1": {"M0", "M1", "M2"}}
	xdas := newFakeXdas()
	xdas.records["M0"] = map[string]string{"t_tag1": "gold"}
	xdas.records["M1"] = map[string]string{"t_other": ""}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect}, env)

	assert.Equal(t, TagSyncStateCompleted, run.State)
	assert.Equal(t, int64(3), run.Counts.Checked)
	assert.Equal(t, int64(1), run.Counts.Present)
	assert.Equal(t, int64(1), run.Counts.MissingField)
	assert.Equal(t, int64(1), run.Counts.MissingKey)
	assert.Equal(t, 0, xdas.pushCount())
	assert.InDelta(t, 2.0/3.0, run.MissingRate, 0.0001)
	assert.Equal(t, 1, run.TagsWithMissing)
	if assert.Len(t, run.TopMissingTags, 1) {
		assert.Equal(t, "tag1", run.TopMissingTags[0].TagId)
		assert.Equal(t, int64(2), run.TopMissingTags[0].Missing)
	}
	// t_other is not a Cassandra tag: counted as XDAS-only, never deleted.
	assert.Equal(t, int64(1), run.Counts.XdasOnlyFieldsSeen)
}

func TestTagSyncRefreshPreservesObservedValue(t *testing.T) {
	cass := map[string][]string{"tag1": {"M0", "M1"}}
	xdas := newFakeXdas()
	xdas.records["M0"] = map[string]string{"t_tag1": "gold"}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	run := execute(t, TagSyncOptions{Mode: TagSyncModeRefresh}, env)

	assert.Equal(t, TagSyncStateCompleted, run.State)
	assert.Equal(t, int64(2), run.Counts.Pushed)

	byMember := map[string]pushRec{}
	for _, p := range xdas.pushes {
		byMember[p.member] = p
	}
	// The present member is re-pushed with the exact value read from XDAS
	// (never a blind ""), the missing one restored with an empty value.
	assert.Equal(t, "gold", byMember["M0"].value)
	assert.Equal(t, "t_tag1", byMember["M0"].tag)
	assert.Equal(t, "", byMember["M1"].value)
}

func TestTagSyncRepairPushesOnlyMissing(t *testing.T) {
	cass := map[string][]string{"tag1": {"M0", "M1", "M2"}}
	xdas := newFakeXdas()
	xdas.records["M0"] = map[string]string{"t_tag1": "gold"}
	xdas.records["M1"] = map[string]string{"t_other": ""}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	run := execute(t, TagSyncOptions{Mode: TagSyncModeRepair}, env)

	assert.Equal(t, TagSyncStateCompleted, run.State)
	assert.Equal(t, int64(2), run.Counts.Pushed)
	for _, p := range xdas.pushes {
		assert.NotEqual(t, "M0", p.member, "present member must not be re-pushed in repair mode")
	}
}

func TestTagSyncDryRunPushesNothing(t *testing.T) {
	cass := map[string][]string{"tag1": members("M", 30)}
	xdas := newFakeXdas()

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	run := execute(t, TagSyncOptions{Mode: TagSyncModeRepair, DryRun: true}, env)

	assert.Equal(t, TagSyncStateCompleted, run.State)
	assert.Equal(t, 0, xdas.pushCount())
	assert.Equal(t, int64(30), run.Counts.WouldPush)
	assert.Equal(t, int64(0), run.Counts.Pushed)
}

func TestTagSyncServerErrorsNeverCountAsMissing(t *testing.T) {
	cass := map[string][]string{"tag1": members("M", 40)}
	xdas := newFakeXdas()
	for _, m := range cass["tag1"] {
		xdas.getErr[m] = xwcommon.NewRemoteErrorAS(500, "boom")
	}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect}, env)

	assert.Equal(t, TagSyncStateAborted, run.State)
	assert.Equal(t, "xdas_unhealthy_consecutive_errors", run.AbortReason)
	assert.Equal(t, int64(0), run.Counts.MissingField+run.Counts.MissingKey,
		"5xx must never be classified as missing")
	assert.GreaterOrEqual(t, run.Counts.XdasErrors, int64(5))
	assert.NotEmpty(t, run.Checkpoint.TagId, "abort must leave a checkpoint")
}

func TestTagSyncPushRetriesTransientFailures(t *testing.T) {
	// The checkpoint never revisits this member, so a transient failure would
	// leave it missing while the run still reported completed.
	cass := map[string][]string{"tag1": {"M0"}}
	xdas := newFakeXdas()
	xdas.records["M0"] = map[string]string{"t_other": ""}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	realPush := env.xdasPush
	var mu sync.Mutex
	attempts := 0
	env.xdasPush = func(member string, tag string, value string) error {
		mu.Lock()
		attempts++
		fail := attempts < tagSyncPushAttempts
		mu.Unlock()
		if fail {
			return xwcommon.NewRemoteErrorAS(503, "unavailable")
		}
		return realPush(member, tag, value)
	}
	run := execute(t, TagSyncOptions{Mode: TagSyncModeRepair}, env)

	assert.Equal(t, TagSyncStateCompleted, run.State)
	assert.Equal(t, int64(1), run.Counts.Pushed, "the retry must land the push")
	assert.Equal(t, int64(0), run.Counts.PushFailed)
	assert.Equal(t, tagSyncPushAttempts, attempts)
}

func TestTagSyncPushGivesUpAfterTheAttemptBudget(t *testing.T) {
	// Bounded: a member XDAS keeps refusing is counted failed, not retried
	// forever.
	cass := map[string][]string{"tag1": {"M0"}}
	xdas := newFakeXdas()
	xdas.records["M0"] = map[string]string{"t_other": ""}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	var mu sync.Mutex
	attempts := 0
	env.xdasPush = func(member string, tag string, value string) error {
		mu.Lock()
		attempts++
		mu.Unlock()
		return xwcommon.NewRemoteErrorAS(503, "unavailable")
	}
	run := execute(t, TagSyncOptions{Mode: TagSyncModeRepair}, env)

	assert.Equal(t, int64(1), run.Counts.PushFailed, "one member left unpushed, not one per attempt")
	assert.Equal(t, int64(1), run.Counts.XdasErrors)
	assert.Equal(t, tagSyncPushAttempts, attempts)
}

func TestTagSyncPushDoesNotRetryRejections(t *testing.T) {
	// A 4xx would be refused again; retrying only burns rate budget.
	cass := map[string][]string{"tag1": {"M0"}}
	xdas := newFakeXdas()
	xdas.records["M0"] = map[string]string{"t_other": ""}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	var mu sync.Mutex
	attempts := 0
	env.xdasPush = func(member string, tag string, value string) error {
		mu.Lock()
		attempts++
		mu.Unlock()
		return xwcommon.NewRemoteErrorAS(400, "bad request")
	}
	run := execute(t, TagSyncOptions{Mode: TagSyncModeRepair}, env)

	assert.Equal(t, int64(1), run.Counts.PushFailed)
	assert.Equal(t, 1, attempts, "a rejected push is not retried")
}

func TestTagSyncMassNotFoundPushesTheWholePopulation(t *testing.T) {
	// XDAS answering "not found" for everything is indistinguishable from
	// mass expiry: a 404 counts as missing, so no breaker fires and repair
	// pushes every member.
	all := membersInOneBucket("M", 1000)
	cass := map[string][]string{"tag1": all}
	xdas := newFakeXdas()

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	env.config.ChunkSize = 1000
	env.xdasGetFields = func(member string) (map[string]string, error) {
		return nil, xwcommon.NewRemoteErrorAS(404, "not found")
	}
	run := execute(t, TagSyncOptions{Mode: TagSyncModeRepair}, env)

	assert.Equal(t, TagSyncStateCompleted, run.State)
	assert.Empty(t, run.AbortReason)
	assert.Equal(t, len(all), xdas.pushCount())
}

func TestGuardBatchSize(t *testing.T) {
	// The guard cannot say anything before minSample members and reads no
	// finer than the window, so the batch is the larger of the two - capped by
	// the Cassandra page it slices.
	assert.Equal(t, 500, guardBatchSize(5000, 200, 500))
	assert.Equal(t, 1000, guardBatchSize(5000, 1000, 500))
	assert.Equal(t, 100, guardBatchSize(100, 200, 500))
	assert.Equal(t, syncBreakerDefaultWindow, guardBatchSize(5000, 0, 0))
}

func TestTagSyncKillSwitchAborts(t *testing.T) {
	cass := map[string][]string{"tag1": members("M", 250)}
	xdas := newFakeXdas()
	for _, m := range cass["tag1"] {
		xdas.records[m] = map[string]string{"t_tag1": ""}
	}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	var mu sync.Mutex
	checks := 0
	env.syncEnabled = func() bool {
		mu.Lock()
		defer mu.Unlock()
		checks++
		return checks <= 1
	}
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect}, env)

	assert.Equal(t, TagSyncStateAborted, run.State)
	assert.Equal(t, "kill_switch", run.AbortReason)
	assert.Less(t, run.Counts.Checked, int64(250))
}

func TestTagSyncMaxMembersStopsClean(t *testing.T) {
	cass := map[string][]string{"tag1": members("M", 250)}
	xdas := newFakeXdas()
	for _, m := range cass["tag1"] {
		xdas.records[m] = map[string]string{"t_tag1": ""}
	}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect, MaxMembers: 150}, env)

	assert.Equal(t, TagSyncStateCompleted, run.State)
	assert.True(t, run.Limited)
	assert.GreaterOrEqual(t, run.Counts.Checked, int64(150))
}

func TestTagSyncResumeSkipsWalkedMembers(t *testing.T) {
	all := members("M", 250)
	cass := map[string][]string{"tag1": all}
	xdas := newFakeXdas()
	for _, m := range all {
		xdas.records[m] = map[string]string{"t_tag1": ""}
	}
	dao := newFakeTagSyncDao()

	// First run dies to the kill switch partway through.
	env1 := newTestEnv(cass, xdas, dao)
	var mu sync.Mutex
	checks := 0
	env1.syncEnabled = func() bool {
		mu.Lock()
		defer mu.Unlock()
		checks++
		return checks <= 1
	}
	run1 := execute(t, TagSyncOptions{Mode: TagSyncModeDetect}, env1)
	assert.Equal(t, TagSyncStateAborted, run1.State)
	firstRunChecked := run1.Counts.Checked
	assert.Greater(t, firstRunChecked, int64(0))
	assert.Less(t, firstRunChecked, int64(250))

	// Second run resumes the same record and finishes the rest.
	xdas.mu.Lock()
	xdas.gets = nil
	xdas.mu.Unlock()

	env2 := newTestEnv(cass, xdas, dao)
	run2 := execute(t, TagSyncOptions{Resume: true}, env2)

	assert.Equal(t, TagSyncStateCompleted, run2.State)
	assert.Equal(t, run1.RunId, run2.RunId, "resume continues the same run record")
	assert.Equal(t, 1, run2.Resumes)
	assert.Equal(t, int64(250), run2.Counts.Checked, "counts accumulate across resume")
	assert.Equal(t, int64(250)-firstRunChecked, int64(xdas.getCount()),
		"resume must not re-check already walked members")
}

func TestTagSyncResumeRestampsTheRecordForTheNewSegment(t *testing.T) {
	// Status polled right after a resume must not show the previous segment's
	// timestamp, nor options claiming this is not a resume.
	stale := time.Now().UTC().Add(-48 * time.Hour)
	dao := newFakeTagSyncDao()
	dao.runs["20260101-000000-aaaa"] = &TagSyncRun{
		RunId:     "20260101-000000-aaaa",
		Mode:      TagSyncModeDetect,
		State:     TagSyncStateAborted,
		Options:   TagSyncOptions{Mode: TagSyncModeDetect},
		StartedAt: stale,
		UpdatedAt: stale,
	}
	env := newTestEnv(map[string][]string{"tag1": {"M0"}}, newFakeXdas(), dao)

	engine, err := prepareTagSync(TagSyncOptions{Resume: true}, env)
	assert.NoError(t, err)
	assert.True(t, engine.opts.Resume)

	saved, err := dao.getRun("20260101-000000-aaaa")
	assert.NoError(t, err)
	assert.True(t, saved.UpdatedAt.After(stale), "the record must be stamped for the segment now running")
	assert.True(t, saved.Options.Resume, "options must describe the segment actually running")
	assert.Equal(t, stale, saved.StartedAt, "startedAt still marks the original run")
	assert.Equal(t, 1, saved.Resumes)
}

func TestTagSyncResumeNothingToResume(t *testing.T) {
	env := newTestEnv(map[string][]string{}, newFakeXdas(), newFakeTagSyncDao())
	_, err := prepareTagSync(TagSyncOptions{Resume: true}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no resumable tag sync run to resume")
	// Nothing to resume is the caller's state, not a server fault.
	assert.Equal(t, http.StatusNotFound, xwcommon.GetXconfErrorStatusCode(err))
}

func TestTagSyncStoreErrorDoesNotLeakDriverDetail(t *testing.T) {
	dao := newFakeTagSyncDao()
	dao.getLockErr = errors.New("gocql: no hosts available in the pool: 10.0.0.7:9042 keyspace ApplicationsDiscoveryDataService")
	env := newTestEnv(map[string][]string{"tag1": {"M0"}}, newFakeXdas(), dao)

	_, err := prepareTagSync(TagSyncOptions{Mode: TagSyncModeDetect}, env)

	if assert.Error(t, err) {
		assert.Equal(t, "tag sync state store unavailable", err.Error())
		assert.NotContains(t, err.Error(), "10.0.0.7")
		assert.NotContains(t, err.Error(), "gocql")
		assert.Equal(t, http.StatusInternalServerError, xwcommon.GetXconfErrorStatusCode(err))
	}
}

func TestTagSyncResumeBehindManyFinishedRuns(t *testing.T) {
	// The aborted run is older than a full status page of completed runs, but
	// still inside the retained history, so resume must reach it.
	dao := newFakeTagSyncDao()
	dao.saveRun(&TagSyncRun{RunId: "20250101-000000", State: TagSyncStateAborted})
	for i := 1; i < tagSyncRunHistoryKeep; i++ {
		dao.saveRun(&TagSyncRun{RunId: fmt.Sprintf("20250101-%06d", i), State: TagSyncStateCompleted})
	}

	resumed, err := findResumableRun(dao)
	assert.NoError(t, err)
	if assert.NotNil(t, resumed) {
		assert.Equal(t, "20250101-000000", resumed.RunId)
	}
}

func TestTagSyncTagFilter(t *testing.T) {
	cass := map[string][]string{
		"tag1": members("A", 10),
		"tag2": members("B", 10),
	}
	xdas := newFakeXdas()
	for _, m := range append(cass["tag1"], cass["tag2"]...) {
		xdas.records[m] = map[string]string{"t_tag1": "", "t_tag2": ""}
	}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect, Tags: []string{"tag2"}}, env)

	assert.Equal(t, TagSyncStateCompleted, run.State)
	assert.Equal(t, 1, run.TagsTotal)
	assert.Equal(t, int64(10), run.Counts.Checked)
	for _, m := range xdas.gets {
		assert.NotContains(t, m, "A-", "tag1 members must not be walked")
	}
}

func TestTagSyncNormalizesRawMemberOnce(t *testing.T) {
	raw := "aa:bb:cc:dd:ee:01"
	normalized := ToNormalizedEcm(raw)
	cass := map[string][]string{"tag1": {raw}}
	xdas := newFakeXdas()

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	run := execute(t, TagSyncOptions{Mode: TagSyncModeRepair}, env)

	assert.Equal(t, TagSyncStateCompleted, run.State)
	if assert.Equal(t, 1, xdas.getCount()) {
		assert.Equal(t, normalized, xdas.gets[0], "GET must use the normalized form of the raw Cassandra member")
	}
	if assert.Equal(t, 1, xdas.pushCount()) {
		assert.Equal(t, normalized, xdas.pushes[0].member, "push must reuse the same normalized form")
	}
}

func TestTagSyncLockBusy(t *testing.T) {
	dao := newFakeTagSyncDao()
	dao.lock = &TagSyncLock{Owner: "other-pod", RunId: "run-1", HeartbeatAt: time.Now().UTC()}

	env := newTestEnv(map[string][]string{}, newFakeXdas(), dao)
	_, err := prepareTagSync(TagSyncOptions{Mode: TagSyncModeDetect}, env)

	var busy *tagSyncBusyError
	if assert.ErrorAs(t, err, &busy) {
		assert.Equal(t, "run-1", busy.RunId)
		assert.Equal(t, "other-pod", busy.Owner)
	}
}

func TestTagSyncStaleLockIsTakenOver(t *testing.T) {
	dao := newFakeTagSyncDao()
	dao.lock = &TagSyncLock{Owner: "crashed-pod", RunId: "run-1", HeartbeatAt: time.Now().UTC().Add(-time.Hour)}

	env := newTestEnv(map[string][]string{"tag1": {"M0"}}, newFakeXdas(), dao)
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect}, env)
	assert.Equal(t, TagSyncStateCompleted, run.State)

	lock, _ := dao.getLock()
	assert.True(t, lock.Released, "lock is released after the run")
	assert.Equal(t, run.RunId, lock.RunId)
}

func TestTagSyncRunPersistedWithFinalState(t *testing.T) {
	dao := newFakeTagSyncDao()
	cass := map[string][]string{"tag1": {"M0", "M1"}}
	xdas := newFakeXdas()
	xdas.records["M0"] = map[string]string{"t_tag1": ""}

	env := newTestEnv(cass, xdas, dao)
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect}, env)

	saved, err := dao.getRun(run.RunId)
	assert.NoError(t, err)
	if assert.NotNil(t, saved) {
		assert.Equal(t, TagSyncStateCompleted, saved.State)
		assert.Equal(t, int64(2), saved.Counts.Checked)
		assert.NotNil(t, saved.CompletedAt)
	}
}

func TestTagSyncStatusShowsLiveMissingRate(t *testing.T) {
	// The rate used to be computed only when the run finished, so status
	// polled during a long walk reported missingRate=0 for the whole run.
	dao := newFakeTagSyncDao()
	env := newTestEnv(map[string][]string{"tag1": {"M0"}}, newFakeXdas(), dao)
	engine, err := prepareTagSync(TagSyncOptions{Mode: TagSyncModeDetect}, env)
	assert.NoError(t, err)

	engine.run.Counts.Checked = 4
	engine.run.Counts.Present = 1
	engine.run.Counts.MissingField = 2
	engine.run.Counts.MissingKey = 1
	engine.saveRun()

	saved, err := dao.getRun(engine.run.RunId)
	assert.NoError(t, err)
	assert.Nil(t, saved.CompletedAt, "the run is still going")
	assert.InDelta(t, 0.75, saved.MissingRate, 0.0001)
}

func TestTagSyncInvalidMode(t *testing.T) {
	opts := TagSyncOptions{Mode: "bogus"}
	err := validateTagSyncOptions(&opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mode")

	opts = TagSyncOptions{Resume: true, Mode: "bogus"}
	assert.Error(t, validateTagSyncOptions(&opts), "a resume still rejects a value that is not a mode")
}

func TestTagSyncValidateDefaultsModeOnlyForFreshRuns(t *testing.T) {
	opts := TagSyncOptions{}
	assert.NoError(t, validateTagSyncOptions(&opts))
	assert.Equal(t, TagSyncModeDetect, opts.Mode)

	opts = TagSyncOptions{Resume: true}
	assert.NoError(t, validateTagSyncOptions(&opts))
	assert.Equal(t, TagSyncMode(""), opts.Mode, "the recorded run supplies the mode")
}

func TestTagSyncResumeRejectsAModeChange(t *testing.T) {
	aborted := func() *fakeTagSyncDao {
		dao := newFakeTagSyncDao()
		dao.runs["20260101-000000-aaaa"] = &TagSyncRun{
			RunId:   "20260101-000000-aaaa",
			Mode:    TagSyncModeDetect,
			State:   TagSyncStateAborted,
			Options: TagSyncOptions{Mode: TagSyncModeDetect},
		}
		return dao
	}
	cass := map[string][]string{"tag1": {"M0"}}

	dao := aborted()
	_, err := prepareTagSync(TagSyncOptions{Resume: true, Mode: TagSyncModeRepair}, newTestEnv(cass, newFakeXdas(), dao))
	var remoteErr xwcommon.RemoteHttpErrorAS
	if assert.ErrorAs(t, err, &remoteErr) {
		assert.Equal(t, 400, remoteErr.StatusCode)
		assert.Contains(t, err.Error(), "mode cannot change on resume")
	}
	assert.Nil(t, dao.lock, "a rejected resume must not take the lock")

	for _, mode := range []TagSyncMode{"", TagSyncModeDetect} {
		engine, err := prepareTagSync(TagSyncOptions{Resume: true, Mode: mode}, newTestEnv(cass, newFakeXdas(), aborted()))
		if assert.NoError(t, err, "mode %q", mode) {
			assert.Equal(t, TagSyncModeDetect, engine.opts.Mode)
		}
	}
}

func TestTagSyncResumeRejectsATagsFilterChange(t *testing.T) {
	aborted := func(tags []string) *fakeTagSyncDao {
		dao := newFakeTagSyncDao()
		dao.runs["20260101-000000-aaaa"] = &TagSyncRun{
			RunId:   "20260101-000000-aaaa",
			Mode:    TagSyncModeDetect,
			State:   TagSyncStateAborted,
			Options: TagSyncOptions{Mode: TagSyncModeDetect, Tags: tags},
		}
		return dao
	}
	cass := map[string][]string{"tag1": {"M0"}}

	// A filter the recorded run does not have, and a different one.
	for _, recorded := range [][]string{nil, {"tag2"}} {
		dao := aborted(recorded)
		_, err := prepareTagSync(TagSyncOptions{Resume: true, Tags: []string{"tag1"}}, newTestEnv(cass, newFakeXdas(), dao))
		var remoteErr xwcommon.RemoteHttpErrorAS
		if assert.ErrorAs(t, err, &remoteErr, "recorded %v", recorded) {
			assert.Equal(t, 400, remoteErr.StatusCode)
			assert.Contains(t, err.Error(), "tags filter cannot change on resume")
		}
		assert.Nil(t, dao.lock, "a rejected resume must not take the lock")
	}

	// No filter, or the recorded one restated — same order or not, with
	// duplicates or not — resumes fine: the filter is a set.
	for _, tags := range [][]string{nil, {"tag1", "tag2"}, {"tag2", "tag1", "tag2"}} {
		engine, err := prepareTagSync(TagSyncOptions{Resume: true, Tags: tags}, newTestEnv(cass, newFakeXdas(), aborted([]string{"tag1", "tag2"})))
		if assert.NoError(t, err, "tags %v", tags) {
			assert.Equal(t, []string{"tag1", "tag2"}, engine.opts.Tags, "the recorded filter stays as recorded")
		}
	}
}

func TestRateLimiterPaces(t *testing.T) {
	limiter := newRateLimiter(100) // 10ms apart
	start := time.Now()
	for i := 0; i < 5; i++ {
		assert.NoError(t, limiter.wait(context.Background()))
	}
	// First call is immediate, the other four are spaced 10ms.
	assert.GreaterOrEqual(t, time.Since(start), 35*time.Millisecond)
}

func TestRateLimiterCancel(t *testing.T) {
	limiter := newRateLimiter(1) // 1/s, second wait would sleep ~1s
	ctx, cancel := context.WithCancel(context.Background())
	assert.NoError(t, limiter.wait(ctx))
	cancel()
	err := limiter.wait(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestSyncBreakerConsecutiveErrors(t *testing.T) {
	b := newSyncBreaker(50, 20, 25, 5)
	for i := 0; i < 4; i++ {
		b.record(syncOutcomeError)
	}
	_, tripped := b.tripped()
	assert.False(t, tripped)

	b.record(syncOutcomeOk) // resets the consecutive counter
	for i := 0; i < 5; i++ {
		b.record(syncOutcomeError)
	}
	reason, tripped := b.tripped()
	assert.True(t, tripped)
	assert.Equal(t, "xdas_unhealthy_consecutive_errors", reason)
}

func TestSyncBreakerErrorRate(t *testing.T) {
	b := newSyncBreaker(50, 20, 25, 1000)
	// 30% errors interleaved with successes: over the 25% threshold once the
	// minimum sample is reached, without ever being consecutive.
	for i := 0; i < 30; i++ {
		if i%3 == 0 {
			b.record(syncOutcomeError)
		} else {
			b.record(syncOutcomeOk)
		}
	}
	reason, tripped := b.tripped()
	assert.True(t, tripped)
	assert.Equal(t, "xdas_unhealthy_error_rate", reason)
}

func TestSyncBreakerRatesArmWithMinSampleLargerThanWindow(t *testing.T) {
	// Production default shape: min_sample (500) is larger than the window
	// (200). The rate thresholds must arm on total members seen, not on the
	// window fill level, which caps at the window size.
	b := newSyncBreaker(200, 500, 25, 100000)
	for i := 0; i < 499; i++ {
		b.record(syncOutcomeError)
	}
	_, tripped := b.tripped()
	assert.False(t, tripped, "rates must not arm before min_sample members total")

	b.record(syncOutcomeError)
	reason, tripped := b.tripped()
	assert.True(t, tripped, "rates must arm once min_sample members were seen in total")
	assert.Equal(t, "xdas_unhealthy_error_rate", reason)
}

func TestTagSyncMidChunkAbortKeepsCheckpointAtChunkStart(t *testing.T) {
	// The breaker trips partway through a chunk and the rest of the chunk
	// drains unprocessed: the checkpoint must stay at the chunk start so a
	// resume re-walks the skipped members instead of losing them forever.
	all := membersInOneBucket("M", 30)
	cass := map[string][]string{"tag1": all}
	xdas := newFakeXdas()
	for _, m := range all {
		xdas.getErr[m] = xwcommon.NewRemoteErrorAS(500, "boom")
	}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect, Workers: 1}, env)

	assert.Equal(t, TagSyncStateAborted, run.State)
	assert.Equal(t, "xdas_unhealthy_consecutive_errors", run.AbortReason)
	assert.Equal(t, "tag1", run.Checkpoint.TagId)
	assert.Equal(t, "", run.Checkpoint.LastMember,
		"checkpoint must not advance past members the abort left unprocessed")
	assert.Less(t, run.Counts.Checked, int64(30))
}

func TestTagSyncZeroConfigValuesAreFloored(t *testing.T) {
	// A zero in config (or the trigger body) must never reach the walk: zero
	// workers would leave the unbuffered member channel with no receivers and
	// hang the run forever while it holds the lock.
	env := newTestEnv(map[string][]string{"tag1": {"M0"}}, newFakeXdas(), newFakeTagSyncDao())
	env.config.RateLimit = 0
	env.config.WorkerCount = 0
	env.config.ChunkSize = 0
	engine, err := prepareTagSync(TagSyncOptions{Mode: TagSyncModeDetect}, env)
	assert.NoError(t, err)
	assert.Equal(t, 1, engine.opts.Rate)
	assert.Equal(t, 1, engine.opts.Workers)
	assert.Equal(t, 1, engine.opts.ChunkSize)

	// And a run configured with zero workers completes instead of hanging.
	cass := map[string][]string{"tag1": {"M0", "M1"}}
	xdas := newFakeXdas()
	for _, m := range cass["tag1"] {
		xdas.records[m] = map[string]string{"t_tag1": ""}
	}
	env2 := newTestEnv(cass, xdas, newFakeTagSyncDao())
	env2.config.WorkerCount = 0
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect}, env2)
	assert.Equal(t, TagSyncStateCompleted, run.State)
	assert.Equal(t, int64(2), run.Counts.Checked)
}

func TestTagSyncResumedTagIsNotCountedTwice(t *testing.T) {
	// A limited run records the tag it stopped inside; the resume re-walks and
	// records it again. Without merging the report counts one tag as two.
	all := members("M", 250)
	cass := map[string][]string{"tag1": all}
	xdas := newFakeXdas()
	for i, m := range all {
		if i%4 != 0 { // 25% missing: under the breaker's missing-rate threshold
			xdas.records[m] = map[string]string{"t_tag1": ""}
		}
	}
	dao := newFakeTagSyncDao()

	env1 := newTestEnv(cass, xdas, dao)
	run1 := execute(t, TagSyncOptions{Mode: TagSyncModeDetect, MaxMembers: 150}, env1)
	assert.True(t, run1.Limited, "the first run must stop inside tag1 for this to be a resume")
	assert.Equal(t, "tag1", run1.Checkpoint.TagId)
	assert.Equal(t, 1, run1.TagsWithMissing)

	env2 := newTestEnv(cass, xdas, dao)
	run2 := execute(t, TagSyncOptions{Resume: true}, env2)

	assert.Equal(t, TagSyncStateCompleted, run2.State)
	assert.Equal(t, 1, run2.TagsWithMissing, "one tag with missing members, not one per run segment")
	if assert.Len(t, run2.TopMissingTags, 1, "the resumed tag must merge into its existing entry") {
		assert.Equal(t, "tag1", run2.TopMissingTags[0].TagId)
		assert.Equal(t, run2.Counts.MissingField+run2.Counts.MissingKey, run2.TopMissingTags[0].Missing,
			"the merged entry must account for every missing member both segments saw")
	}
}

func TestTagSyncTagStatsMergeAcrossSegments(t *testing.T) {
	// A listed tag whose later segment finds nothing missing must still fold
	// that segment's checked/pushed into its entry.
	e := &tagSyncEngine{run: &TagSyncRun{}}
	e.recordTagResult(&TagMissingStat{TagId: "tag1", Checked: 100, Missing: 5, Pushed: 5})
	e.recordTagResult(&TagMissingStat{TagId: "tag1", Checked: 50, Pushed: 3})
	e.recordTagResult(&TagMissingStat{TagId: "tag2", Checked: 70})

	assert.Equal(t, 1, e.run.TagsWithMissing, "a tag with nothing missing is never listed")
	if assert.Len(t, e.missingStats, 1) {
		assert.Equal(t, "tag1", e.missingStats[0].TagId)
		assert.Equal(t, int64(150), e.missingStats[0].Checked)
		assert.Equal(t, int64(5), e.missingStats[0].Missing)
		assert.Equal(t, int64(8), e.missingStats[0].Pushed)
	}
}

func TestTagSyncResumeSpendsAFreshMemberBudget(t *testing.T) {
	// Restating the same maxMembers on a resume must walk the next page. Against
	// the run's lifetime Checked the budget is spent and it walks one member.
	all := members("M", 250)
	cass := map[string][]string{"tag1": all}
	xdas := newFakeXdas()
	for _, m := range all {
		xdas.records[m] = map[string]string{"t_tag1": ""}
	}
	dao := newFakeTagSyncDao()

	env1 := newTestEnv(cass, xdas, dao)
	run1 := execute(t, TagSyncOptions{Mode: TagSyncModeDetect, MaxMembers: 150}, env1)
	assert.True(t, run1.Limited)
	assert.Equal(t, int64(150), run1.Counts.Checked)

	env2 := newTestEnv(cass, xdas, dao)
	run2 := execute(t, TagSyncOptions{Resume: true, MaxMembers: 150}, env2)

	assert.Greater(t, run2.Counts.Checked-run1.Counts.Checked, int64(50),
		"the resume must spend its own maxMembers budget, not the run's lifetime total")
	assert.Equal(t, TagSyncStateCompleted, run2.State)
}

func TestTagSyncReleaseLeavesAnotherRunsLockAlone(t *testing.T) {
	// A run that lost its lock to a takeover must not mark the new owner's lock
	// released on the way out - that would let a third run start.
	dao := newFakeTagSyncDao()
	assert.NoError(t, dao.saveLock(&TagSyncLock{
		Owner: "podA", RunId: "runA", HeartbeatAt: time.Now().UTC().Add(-10 * tagSyncLockStaleAfter),
	}))
	assert.NoError(t, acquireTagSyncLock(dao, "podB", "runB"))

	releaseTagSyncLock(dao, "podA", "runA")

	lock, err := dao.getLock()
	assert.NoError(t, err)
	if assert.NotNil(t, lock) {
		assert.Equal(t, "runB", lock.RunId, "B still owns the lock")
		assert.False(t, lock.Released, "B is still walking, so its lock must not read as released")
	}
	assert.Error(t, acquireTagSyncLock(dao, "podC", "runC"), "a third run must still be refused")

	// B's own release works normally.
	releaseTagSyncLock(dao, "podB", "runB")
	assert.NoError(t, acquireTagSyncLock(dao, "podC", "runC"))
}

func TestTagSyncStaleHeartbeatStopsTheWalk(t *testing.T) {
	// Unrefreshed for a full staleness window, the lock may already be taken
	// over, so this walk must stop.
	env := newTestEnv(map[string][]string{"tag1": {"M0"}}, newFakeXdas(), newFakeTagSyncDao())
	engine, err := prepareTagSync(TagSyncOptions{Mode: TagSyncModeDetect}, env)
	assert.NoError(t, err)

	assert.NoError(t, engine.checkAbort(context.Background()), "a fresh run holds a fresh heartbeat")

	engine.mu.Lock()
	engine.lastHeartbeatOk = time.Now().Add(-2 * tagSyncLockStaleAfter)
	engine.mu.Unlock()

	err = engine.checkAbort(context.Background())
	var abort *tagSyncAbort
	if assert.ErrorAs(t, err, &abort) {
		assert.Equal(t, "lock_heartbeat_stale", abort.reason)
	}
}

func TestTagSyncHeartbeatStopsAfterLockTakeover(t *testing.T) {
	// Heartbeats stalled long enough for another instance to take the lock
	// over. Once they recover, writing ours back would flap the row and leave
	// both walks running: the loop must stand down and let the walk abort.
	oldStale := tagSyncLockStaleAfter
	tagSyncLockStaleAfter = 60 * time.Millisecond
	defer func() { tagSyncLockStaleAfter = oldStale }()

	dao := newFakeTagSyncDao()
	env := newTestEnv(map[string][]string{"tag1": {"M0"}}, newFakeXdas(), dao)
	engine, err := prepareTagSync(TagSyncOptions{Mode: TagSyncModeDetect}, env)
	assert.NoError(t, err)

	assert.NoError(t, dao.saveLock(&TagSyncLock{Owner: "podB", RunId: "runB", HeartbeatAt: time.Now().UTC()}))

	stop := make(chan struct{})
	defer close(stop)
	done := make(chan struct{})
	go engine.heartbeatLoop(stop, done)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("the heartbeat loop kept running after the takeover")
	}

	lock, err := dao.getLock()
	assert.NoError(t, err)
	if assert.NotNil(t, lock) {
		assert.Equal(t, "runB", lock.RunId, "the new owner's lock must not be overwritten")
	}
	var abort *tagSyncAbort
	if assert.ErrorAs(t, engine.checkAbort(context.Background()), &abort) {
		assert.Equal(t, "lock_heartbeat_stale", abort.reason, "the walk must stop, not keep going beside the new owner")
	}
}

func TestTagSyncCheckpointTagSurvivesLeaderboardTrim(t *testing.T) {
	// With more tags missing than the leaderboard keeps, the tag the run
	// stopped inside must still be on it: a resume seeds from the record and
	// merges into that entry instead of counting the tag a second time.
	env := newTestEnv(map[string][]string{"tag1": {"M0"}}, newFakeXdas(), newFakeTagSyncDao())
	engine, err := prepareTagSync(TagSyncOptions{Mode: TagSyncModeDetect}, env)
	assert.NoError(t, err)

	for i := 0; i < 2*tagSyncTopMissingKeep+5; i++ {
		engine.setCheckpoint(fmt.Sprintf("busy%03d", i), 0, "M9")
		engine.recordTagResult(&TagMissingStat{TagId: fmt.Sprintf("busy%03d", i), Checked: 1000, Missing: 1000})
	}
	// The run stops inside a tag with far fewer missing members than the rest.
	engine.setCheckpoint("tail", 0, "M9")
	engine.recordTagResult(&TagMissingStat{TagId: "tail", Checked: 10, Missing: 1})
	saved := engine.saveRun()

	assert.LessOrEqual(t, len(saved.TopMissingTags), tagSyncTopMissingKeep+1)
	assert.True(t, containsTagStat(saved.TopMissingTags, "tail"), "the tag a resume re-walks must stay on the leaderboard")

	resumed := &tagSyncEngine{
		run:          &TagSyncRun{TagsWithMissing: saved.TagsWithMissing},
		missingStats: append([]TagMissingStat(nil), saved.TopMissingTags...),
	}
	resumed.recordTagResult(&TagMissingStat{TagId: "tail", Checked: 5, Missing: 2})

	assert.Equal(t, saved.TagsWithMissing, resumed.run.TagsWithMissing, "the re-walked tag must merge, not count as a second tag")
	if assert.True(t, containsTagStat(resumed.missingStats, "tail")) {
		merged := TagMissingStat{}
		for _, stat := range resumed.missingStats {
			if stat.TagId == "tail" {
				merged = stat
			}
		}
		assert.Equal(t, int64(15), merged.Checked)
		assert.Equal(t, int64(3), merged.Missing)
	}
}

func TestSyncBreakerZeroConfigDoesNotTripOnHealthyMembers(t *testing.T) {
	// Thresholds are >= comparisons, so a zero from config would abort on the
	// first healthy member. Non-positive knobs fall back to defaults.
	b := newSyncBreaker(0, 0, 0, 0)
	for i := 0; i < 1000; i++ {
		b.record(syncOutcomeOk)
	}
	reason, tripped := b.tripped()
	assert.False(t, tripped, "healthy traffic must not trip a zero-configured breaker: %s", reason)

	// The defaults are genuinely in force: 10 consecutive errors still trip.
	for i := 0; i < syncBreakerDefaultMaxConsecErrors; i++ {
		b.record(syncOutcomeError)
	}
	reason, tripped = b.tripped()
	assert.True(t, tripped)
	assert.Equal(t, "xdas_unhealthy_consecutive_errors", reason)
}

func TestTagSyncLimitedRunIsResumable(t *testing.T) {
	// Ramp workflow: run with MaxMembers, review the numbers, then resume to
	// cover the rest without re-walking what was already checked.
	all := members("M", 250)
	cass := map[string][]string{"tag1": all}
	xdas := newFakeXdas()
	for _, m := range all {
		xdas.records[m] = map[string]string{"t_tag1": ""}
	}
	dao := newFakeTagSyncDao()

	env1 := newTestEnv(cass, xdas, dao)
	run1 := execute(t, TagSyncOptions{Mode: TagSyncModeDetect, MaxMembers: 150}, env1)
	assert.Equal(t, TagSyncStateCompleted, run1.State)
	assert.True(t, run1.Limited)

	env2 := newTestEnv(cass, xdas, dao)
	run2 := execute(t, TagSyncOptions{Resume: true}, env2)

	assert.Equal(t, TagSyncStateCompleted, run2.State)
	assert.False(t, run2.Limited)
	assert.Equal(t, run1.RunId, run2.RunId, "resume continues the same run record")
	assert.GreaterOrEqual(t, run2.Counts.Checked, int64(250))

	// The real invariant: between the two runs, every member was checked.
	xdas.mu.Lock()
	seen := make(map[string]bool, len(xdas.gets))
	for _, g := range xdas.gets {
		seen[g] = true
	}
	xdas.mu.Unlock()
	for _, m := range all {
		assert.True(t, seen[m], "member %s was never checked", m)
	}
}

func TestTagSyncRunHistoryPruned(t *testing.T) {
	dao := newFakeTagSyncDao()
	for i := 0; i < 30; i++ {
		dao.saveRun(&TagSyncRun{RunId: fmt.Sprintf("20250101-%06d", i), State: TagSyncStateCompleted})
	}
	cass := map[string][]string{"tag1": {"M0"}}
	xdas := newFakeXdas()
	xdas.records["M0"] = map[string]string{"t_tag1": ""}

	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect}, newTestEnv(cass, xdas, dao))
	assert.Equal(t, TagSyncStateCompleted, run.State)

	runs, err := dao.listRuns(2 * tagSyncRunHistoryKeep)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(runs), tagSyncRunHistoryKeep, "old runs are pruned at finalize")
	saved, _ := dao.getRun(run.RunId)
	assert.NotNil(t, saved, "the just-finished run survives pruning")
}

func TestTagSyncNoTagsAborts(t *testing.T) {
	// An empty tag census is indistinguishable from a swallowed Cassandra
	// failure: abort loudly instead of recording a missingRate=0 all-clear.
	env := newTestEnv(map[string][]string{}, newFakeXdas(), newFakeTagSyncDao())
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect}, env)

	assert.Equal(t, TagSyncStateAborted, run.State)
	assert.Equal(t, "cassandra_suspect_no_tags", run.AbortReason)
}

func TestTagSyncEmptyPopulatedBucketAborts(t *testing.T) {
	// getPopulatedBuckets reported members but the first page comes back
	// empty: that smells like a swallowed Cassandra error reading as
	// end-of-bucket, so the run aborts instead of skipping the bucket.
	cass := map[string][]string{"tag1": {"M0"}}
	xdas := newFakeXdas()
	xdas.records["M0"] = map[string]string{"t_tag1": ""}

	env := newTestEnv(cass, xdas, newFakeTagSyncDao())
	env.getMembersFromBucket = func(tagId string, bucketId int, lastMember string, limit int) ([]string, error) {
		return nil, nil
	}
	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect}, env)

	assert.Equal(t, TagSyncStateAborted, run.State)
	assert.Equal(t, "cassandra_suspect_empty_bucket", run.AbortReason)
}

func TestTagSyncHeartbeatIndependentOfChunkPace(t *testing.T) {
	// One slow chunk, longer than the whole staleness window: the lock must
	// still be refreshed on wall-clock cadence so no other instance can
	// treat it as stale mid-chunk and start a concurrent run.
	oldStale := tagSyncLockStaleAfter
	tagSyncLockStaleAfter = 300 * time.Millisecond
	defer func() { tagSyncLockStaleAfter = oldStale }()

	all := membersInOneBucket("H", 4)
	cass := map[string][]string{"tag1": all}
	xdas := newFakeXdas()
	for _, m := range all {
		xdas.records[m] = map[string]string{"t_tag1": ""}
	}
	dao := newFakeTagSyncDao()
	env := newTestEnv(cass, xdas, dao)
	realGet := env.xdasGetFields
	env.xdasGetFields = func(member string) (map[string]string, error) {
		time.Sleep(150 * time.Millisecond)
		return realGet(member)
	}

	run := execute(t, TagSyncOptions{Mode: TagSyncModeDetect, Workers: 1}, env)
	assert.Equal(t, TagSyncStateCompleted, run.State)

	writes := dao.lockWriteTimes()
	assert.GreaterOrEqual(t, len(writes), 3, "heartbeats must fire during the slow chunk")
	for i := 1; i < len(writes); i++ {
		assert.Less(t, writes[i].Sub(writes[i-1]), tagSyncLockStaleAfter,
			"the lock must never look stale while the run is alive")
	}
}

func TestSyncBreakerMissingNeverTrips(t *testing.T) {
	// A missing member is a healthy XDAS answer, so even an all-missing
	// window must leave the breaker closed.
	b := newSyncBreaker(50, 20, 25, 5)
	for i := 0; i < 30; i++ {
		b.record(syncOutcomeMissing)
	}
	_, tripped := b.tripped()
	assert.False(t, tripped, "missing members must never trip the breaker")
}
