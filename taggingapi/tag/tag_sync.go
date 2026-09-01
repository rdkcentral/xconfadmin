package tag

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	taggingapi_config "github.com/rdkcentral/xconfadmin/taggingapi/config"
	proto "github.com/rdkcentral/xconfadmin/taggingapi/proto/generated"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	log "github.com/sirupsen/logrus"
)

// The tag sync job walks Cassandra (the source of truth) and brings XDAS in
// line with it. Additive-only: it never deletes from either store.
//
//	detect  - read-only census of members present/missing in XDAS
//	repair  - push back only the members missing in XDAS
//	refresh - repair plus re-pushing present members with their observed XDAS
//	          value. Makes the read region authoritative: a member missing
//	          there is pushed blank over any value another region holds.
type TagSyncMode string

const (
	TagSyncModeDetect  TagSyncMode = "detect"
	TagSyncModeRepair  TagSyncMode = "repair"
	TagSyncModeRefresh TagSyncMode = "refresh"
)

const (
	TagSyncStateRunning   = "running"
	TagSyncStateCompleted = "completed"
	TagSyncStateAborted   = "aborted"

	// Internal bounds, deliberately not config: they cap memory and noise.
	tagSyncTopMissingKeep           = 100         // leaderboard entries; the run record is one JSON cell
	tagSyncMaxCountedXdasOnlyFields = 10000       // past this, new fields stop being counted
	tagSyncProgressEvery            = time.Minute // progress log throttle
	tagSyncPushAttempts             = 3           // per-member push attempts; 4xx is not retried
	tagSyncRunHistoryKeep           = 20          // run records kept at finalize
)

var (
	tagSyncLockStaleAfter = 5 * time.Minute
	// Write, wait, re-read: of two racing writers exactly one survives.
	tagSyncLockSettle = 500 * time.Millisecond
)

type TagSyncOptions struct {
	Mode       TagSyncMode `json:"mode"`
	Tags       []string    `json:"tags,omitempty"`
	Rate       int         `json:"rate,omitempty"`
	Workers    int         `json:"workers,omitempty"`
	ChunkSize  int         `json:"chunkSize,omitempty"`
	DryRun     bool        `json:"dryRun,omitempty"`
	MaxMembers int64       `json:"maxMembers,omitempty"`
	Resume     bool        `json:"resume,omitempty"`
}

type TagSyncCounts struct {
	Checked            int64 `json:"checked"`
	Present            int64 `json:"present"`
	MissingField       int64 `json:"missingField"`
	MissingKey         int64 `json:"missingKey"`
	Pushed             int64 `json:"pushed"`
	PushFailed         int64 `json:"pushFailed"`
	WouldPush          int64 `json:"wouldPush,omitempty"`
	XdasErrors         int64 `json:"xdasErrors"`
	XdasOnlyFieldsSeen int64 `json:"xdasOnlyFieldsSeen"`
}

type TagSyncCheckpoint struct {
	TagId      string `json:"tagId,omitempty"`
	BucketId   int    `json:"bucketId,omitempty"`
	LastMember string `json:"lastMember,omitempty"`
}

type TagMissingStat struct {
	TagId   string `json:"tagId"`
	Checked int64  `json:"checked"`
	Missing int64  `json:"missing"`
	Pushed  int64  `json:"pushed"`
}

// TagSyncRun is the persisted record of one run, saved every checkpoint
// interval so it doubles as the status endpoint's progress view.
type TagSyncRun struct {
	RunId           string            `json:"runId"`
	Mode            TagSyncMode       `json:"mode"`
	State           string            `json:"state"`
	Options         TagSyncOptions    `json:"options"`
	Owner           string            `json:"owner"`
	StartedAt       time.Time         `json:"startedAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	CompletedAt     *time.Time        `json:"completedAt,omitempty"`
	Checkpoint      TagSyncCheckpoint `json:"checkpoint"`
	Counts          TagSyncCounts     `json:"counts"`
	TagsTotal       int               `json:"tagsTotal"`
	TagsDone        int               `json:"tagsDone"`
	TagsWithMissing int               `json:"tagsWithMissing"`
	TopMissingTags  []TagMissingStat  `json:"topMissingTags,omitempty"`
	MissingRate     float64           `json:"missingRate"`
	AbortReason     string            `json:"abortReason,omitempty"`
	Limited         bool              `json:"limited,omitempty"`
	Resumes         int               `json:"resumes,omitempty"`
}

// tagSyncBusyError maps to 409 at the trigger endpoint.
type tagSyncBusyError struct {
	RunId string
	Owner string
}

func (e *tagSyncBusyError) Error() string {
	return fmt.Sprintf("tag sync already running: runId=%s owner=%s", e.RunId, e.Owner)
}

// tagSyncAbort carries the reason a run stopped early;
// errTagSyncLimitReached marks the clean MaxMembers stop.
type tagSyncAbort struct {
	reason string
}

func (e *tagSyncAbort) Error() string { return "tag sync aborted: " + e.reason }

var errTagSyncLimitReached = errors.New("tag sync member limit reached")

// tagSyncStoreError keeps driver detail (hosts, keyspace, query fragments) in
// the log and hands the caller a generic message. The status stays 500: an
// unclassified store error may be a down cluster or a missing table, and only
// one of those is worth retrying.
func tagSyncStoreError(op string, err error) error {
	log.Errorf("tag sync state store %s failed: %v", op, err)
	return xwcommon.NewRemoteErrorAS(http.StatusInternalServerError,
		"tag sync state store unavailable")
}

// tagSyncEnv is the seam tests swap fakes into.
type tagSyncEnv struct {
	getAllTagIds         func() ([]string, error)
	getPopulatedBuckets  func(tagId string) ([]int, error)
	getMembersFromBucket func(tagId string, bucketId int, lastMember string, limit int) ([]string, error)
	// Both take an already-normalized member.
	xdasGetFields func(normalizedMember string) (map[string]string, error)
	xdasPush      func(normalizedMember string, prefixedTag string, value string) error
	dao           tagSyncDao
	syncEnabled   func() bool
	config        *taggingapi_config.TagSyncConfig
}

func newTagSyncEnv() (*tagSyncEnv, error) {
	if xhttp.WebConfServer == nil || xhttp.WebConfServer.TagSyncConfig == nil {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusServiceUnavailable,
			"tag sync: server not initialized")
	}
	return &tagSyncEnv{
		getAllTagIds:         GetAllTagIds,
		getPopulatedBuckets:  getPopulatedBuckets,
		getMembersFromBucket: getMembersFromBucket,
		xdasGetFields: func(normalizedMember string) (map[string]string, error) {
			hashes, err := GetGroupServiceConnector().GetGroupsMemberBelongsTo(normalizedMember)
			if err != nil {
				return nil, err
			}
			return hashes.GetFields(), nil
		},
		xdasPush: func(normalizedMember string, prefixedTag string, value string) error {
			xdasMembers := proto.XdasHashes{
				Fields: map[string]string{prefixedTag: value},
			}
			return GetGroupServiceSyncConnector().AddMembersToTag(normalizedMember, &xdasMembers)
		},
		dao:         newTagSyncDao(),
		syncEnabled: tagSyncKillSwitchEnabled,
		config:      xhttp.WebConfServer.TagSyncConfig,
	}, nil
}

// Absent or unreadable means enabled. Other instances see a flip only after
// their cache refresh, so a cross-instance stop takes about a minute.
func tagSyncKillSwitchEnabled() bool {
	return common.GetBooleanAppSetting(common.PROP_TAGGING_SYNC_ENABLED, true)
}

type tagSyncEngine struct {
	env     *tagSyncEnv
	opts    TagSyncOptions
	run     *TagSyncRun
	limiter *rateLimiter
	breaker *syncBreaker

	knownTags    map[string]bool
	xdasOnlySeen map[string]bool

	// Members processed between outage-guard evaluations.
	guardBatch int

	// Checked at segment start, so MaxMembers is per segment.
	checkedAtStart int64

	mu              sync.Mutex
	missingStats    []TagMissingStat
	lastSave        time.Time
	lastProgress    time.Time
	lastHeartbeatOk time.Time
}

// PrepareTagSync validates, locks and saves the run record so the trigger can
// answer with the run id. Drive the returned engine with Execute, which
// releases the lock on every path.
func PrepareTagSync(opts TagSyncOptions) (*tagSyncEngine, error) {
	if err := validateTagSyncOptions(&opts); err != nil {
		return nil, err
	}
	env, err := newTagSyncEnv()
	if err != nil {
		return nil, err
	}
	return prepareTagSync(opts, env)
}

// A resume leaves an omitted mode empty for the recorded run to supply.
func validateTagSyncOptions(opts *TagSyncOptions) error {
	if opts.Mode == "" {
		if !opts.Resume {
			opts.Mode = TagSyncModeDetect
		}
		return nil
	}
	switch opts.Mode {
	case TagSyncModeDetect, TagSyncModeRepair, TagSyncModeRefresh:
		return nil
	default:
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
			fmt.Sprintf("invalid mode %q: must be detect, repair or refresh", opts.Mode))
	}
}

// The breakers cannot arm before minSample and read no finer than the window,
// so batching at the larger of the two evaluates them as early as they can say
// anything - capping a bad XDAS at one batch, not a whole Cassandra page.
func guardBatchSize(chunkSize, window, minSample int) int {
	if window <= 0 {
		window = syncBreakerDefaultWindow
	}
	return min(chunkSize, max(window, minSample))
}

// Floored to 1: zero workers would leave the unbuffered member channel with
// no receivers, hanging the run while it holds the lock.
func positiveOr(v, fallback int) int {
	if v > 0 {
		return v
	}
	if fallback > 0 {
		return fallback
	}
	return 1
}

func prepareTagSync(opts TagSyncOptions, env *tagSyncEnv) (*tagSyncEngine, error) {
	initTagSyncMetrics()
	cfg := env.config
	opts.Rate = positiveOr(opts.Rate, cfg.RateLimit)
	opts.Workers = positiveOr(opts.Workers, cfg.WorkerCount)
	opts.ChunkSize = positiveOr(opts.ChunkSize, cfg.ChunkSize)

	owner, _ := os.Hostname()
	if owner == "" {
		owner = "unknown"
	}

	var run *TagSyncRun
	if opts.Resume {
		resumed, err := findResumableRun(env.dao)
		if err != nil {
			return nil, err
		}
		// Silently keeping the recorded mode would let an operator believe a
		// detect census had been resumed as a repair.
		if opts.Mode != "" && opts.Mode != resumed.Mode {
			return nil, xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
				fmt.Sprintf("mode cannot change on resume: run %s is %s", resumed.RunId, resumed.Mode))
		}
		// Same for the tag filter: the recorded one wins on resume.
		if len(opts.Tags) > 0 && !sameTagFilter(opts.Tags, resumed.Options.Tags) {
			return nil, xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
				fmt.Sprintf("tags filter cannot change on resume: run %s recorded %v", resumed.RunId, resumed.Options.Tags))
		}
		run = resumed
		run.State = TagSyncStateRunning
		run.AbortReason = ""
		run.CompletedAt = nil
		run.Limited = false
		run.Owner = owner
		// Otherwise status shows a running run that looks dead until the first
		// batch save.
		run.UpdatedAt = time.Now().UTC()
		run.Resumes++
		// Mode and filters stay as recorded; pacing may be overridden.
		run.Options.Rate = opts.Rate
		run.Options.Workers = opts.Workers
		run.Options.ChunkSize = opts.ChunkSize
		run.Options.DryRun = opts.DryRun
		run.Options.MaxMembers = opts.MaxMembers
		run.Options.Resume = true
		opts = run.Options
	} else {
		now := time.Now().UTC()
		run = &TagSyncRun{
			RunId:     fmt.Sprintf("%s-%s", now.Format("20060102-150405"), uuid.New().String()[:8]),
			Mode:      opts.Mode,
			State:     TagSyncStateRunning,
			Options:   opts,
			Owner:     owner,
			StartedAt: now,
			UpdatedAt: now,
		}
	}

	if err := acquireTagSyncLock(env.dao, owner, run.RunId); err != nil {
		return nil, err
	}
	if err := env.dao.saveRun(run); err != nil {
		releaseTagSyncLock(env.dao, owner, run.RunId)
		return nil, tagSyncStoreError("run save", err)
	}

	return &tagSyncEngine{
		env:          env,
		opts:         opts,
		run:          run,
		limiter:      newRateLimiter(opts.Rate),
		breaker:      newSyncBreaker(cfg.BreakerWindow, cfg.BreakerMinSample, cfg.BreakerErrorRatePercent, cfg.BreakerMaxConsecErrors),
		xdasOnlySeen: make(map[string]bool),
		missingStats: append([]TagMissingStat(nil), run.TopMissingTags...),
		// The lock was just acquired, so start the staleness clock now.
		lastHeartbeatOk: time.Now(),
		checkedAtStart:  run.Counts.Checked,
		guardBatch:      guardBatchSize(opts.ChunkSize, cfg.BreakerWindow, cfg.BreakerMinSample),
	}, nil
}

func findResumableRun(dao tagSyncDao) (*TagSyncRun, error) {
	// A resumable run stays resumable while its record survives pruning.
	runs, err := dao.listRuns(tagSyncRunHistoryKeep)
	if err != nil {
		return nil, tagSyncStoreError("run list", err)
	}
	for _, run := range runs {
		// "running" with a stale lock is a crashed run (acquireTagSyncLock
		// only succeeds on stale locks); completed-limited still has a live
		// checkpoint, for the ramp-then-continue workflow.
		if run.State == TagSyncStateAborted || run.State == TagSyncStateRunning ||
			(run.State == TagSyncStateCompleted && run.Limited) {
			return run, nil
		}
	}
	return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound,
		"no aborted tag sync run to resume")
}

func acquireTagSyncLock(dao tagSyncDao, owner string, runId string) error {
	existing, err := dao.getLock()
	if err != nil {
		return tagSyncStoreError("lock read", err)
	}
	if existing != nil && !existing.Released && time.Since(existing.HeartbeatAt) < tagSyncLockStaleAfter {
		return &tagSyncBusyError{RunId: existing.RunId, Owner: existing.Owner}
	}
	mine := &TagSyncLock{Owner: owner, RunId: runId, HeartbeatAt: time.Now().UTC()}
	if err := dao.saveLock(mine); err != nil {
		return tagSyncStoreError("lock write", err)
	}
	time.Sleep(tagSyncLockSettle)
	current, err := dao.getLock()
	if err != nil {
		return tagSyncStoreError("lock re-read", err)
	}
	if current == nil {
		return errors.New("tag sync lock vanished after write")
	}
	if current.Owner != owner || current.RunId != runId {
		return &tagSyncBusyError{RunId: current.RunId, Owner: current.Owner}
	}
	return nil
}

// Releases only a lock this run still holds: the row is one cell every
// instance overwrites, so releasing after a takeover would hand a third run a
// free pass over a walk still in progress.
func releaseTagSyncLock(dao tagSyncDao, owner string, runId string) {
	current, err := dao.getLock()
	if err != nil {
		log.Errorf("tag sync lock read before release failed (goes stale in %v): %v", tagSyncLockStaleAfter, err)
		return
	}
	if current != nil && current.RunId != runId {
		log.Warnf("tag sync lock is held by runId=%s owner=%s; runId=%s leaves it alone", current.RunId, current.Owner, runId)
		return
	}
	if err := dao.saveLock(&TagSyncLock{Owner: owner, RunId: runId, HeartbeatAt: time.Now().UTC(), Released: true}); err != nil {
		log.Errorf("tag sync lock release failed (goes stale in %v): %v", tagSyncLockStaleAfter, err)
	}
}

// Execute never returns an error: every outcome is recorded on the run.
func (e *tagSyncEngine) Execute(ctx context.Context) *TagSyncRun {
	tagSyncRunningGauge.Set(1)
	start := time.Now()
	hbStop := make(chan struct{})
	hbDone := make(chan struct{})
	go e.heartbeatLoop(hbStop, hbDone)
	defer func() {
		// Stop the heartbeat before the release write, or an in-flight one
		// resurrects the lock as live-unreleased.
		close(hbStop)
		<-hbDone
		releaseTagSyncLock(e.env.dao, e.run.Owner, e.run.RunId)
		tagSyncRunningGauge.Set(0)
		tagSyncRunDurationSeconds.Set(time.Since(start).Seconds())
	}()

	e.logf(log.InfoLevel, "tag sync run started: mode=%s dryRun=%v rate=%d workers=%d chunkSize=%d resume=%v",
		e.opts.Mode, e.opts.DryRun, e.opts.Rate, e.opts.Workers, e.opts.ChunkSize, e.opts.Resume)

	e.finish(e.walk(ctx))
	return e.run
}

// walk visits every selected tag from the run's checkpoint, empty on a fresh
// run.
func (e *tagSyncEngine) walk(ctx context.Context) error {
	allTags, err := e.env.getAllTagIds()
	if err != nil {
		return err
	}
	if len(allTags) == 0 {
		// The shared query helper reports errors as empty rows, and this
		// system always has tags - so this is a swallowed Cassandra failure,
		// not a real all-clear with missingRate=0.
		return &tagSyncAbort{reason: "cassandra_suspect_no_tags"}
	}
	e.knownTags = make(map[string]bool, len(allTags))
	for _, tagId := range allTags {
		e.knownTags[SetTagPrefix(tagId)] = true
	}

	tags := filterTags(allTags, e.opts.Tags)
	sort.Strings(tags)
	e.run.TagsTotal = len(tags)
	e.run.TagsDone = 0

	// walkTag only consults the checkpoint on the tag it names.
	resumeCp := e.run.Checkpoint
	for _, tagId := range tags {
		if tagId < resumeCp.TagId {
			e.run.TagsDone++
			continue
		}
		perTag := &TagMissingStat{TagId: tagId}
		err := e.walkTag(ctx, tagId, perTag, resumeCp)
		e.recordTagResult(perTag)
		if err != nil {
			return err
		}
		e.run.TagsDone++
	}
	return nil
}

func (e *tagSyncEngine) finish(err error) {
	var abort *tagSyncAbort
	switch {
	case err == nil:
		e.finishCompleted(false)
	case errors.Is(err, errTagSyncLimitReached):
		e.finishCompleted(true)
	case errors.As(err, &abort):
		e.finishAborted(abort.reason)
	default:
		e.finishAborted("cassandra_error: " + err.Error())
	}
}

func filterTags(all []string, requested []string) []string {
	if len(requested) == 0 {
		return all
	}
	allSet := make(map[string]bool, len(all))
	for _, tagId := range all {
		allSet[tagId] = true
	}
	filtered := make([]string, 0, len(requested))
	seen := make(map[string]bool, len(requested))
	for _, tagId := range requested {
		if allSet[tagId] && !seen[tagId] {
			filtered = append(filtered, tagId)
			seen[tagId] = true
		}
	}
	return filtered
}

// Compared as sets: the walk sorts and filterTags dedupes.
func sameTagFilter(a, b []string) bool {
	canon := func(tags []string) []string {
		return slices.Compact(slices.Sorted(slices.Values(tags)))
	}
	return slices.Equal(canon(a), canon(b))
}

// Wall-clock cadence, independent of chunk pacing: a chunk slower than the
// staleness window must not let another instance start a second run. Stops
// the moment the lock belongs to someone else.
func (e *tagSyncEngine) heartbeatLoop(stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(tagSyncLockStaleAfter / 3)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			current, err := e.env.dao.getLock()
			if err != nil {
				e.logf(log.WarnLevel, "tag sync heartbeat lock read failed: %v", err)
				continue
			}
			if current != nil && current.RunId != "" && current.RunId != e.run.RunId {
				e.logf(log.WarnLevel, "tag sync heartbeat: lock taken over by runId=%s owner=%s; this run stops",
					current.RunId, current.Owner)
				e.mu.Lock()
				e.lastHeartbeatOk = time.Time{}
				e.mu.Unlock()
				return
			}
			lock := &TagSyncLock{Owner: e.run.Owner, RunId: e.run.RunId, HeartbeatAt: time.Now().UTC()}
			if err := e.env.dao.saveLock(lock); err != nil {
				// checkAbort stops the walk once these span a full window.
				e.logf(log.WarnLevel, "tag sync heartbeat save failed: %v", err)
				continue
			}
			e.mu.Lock()
			e.lastHeartbeatOk = time.Now()
			e.mu.Unlock()
		}
	}
}

func (e *tagSyncEngine) walkTag(ctx context.Context, tagId string, perTag *TagMissingStat, resume TagSyncCheckpoint) error {
	prefixedTag := SetTagPrefix(tagId)
	buckets, err := e.env.getPopulatedBuckets(tagId)
	if err != nil {
		e.setCheckpoint(tagId, 0, "")
		return fmt.Errorf("buckets for tag %s: %w", tagId, err)
	}
	sort.Ints(buckets)

	for _, bucketId := range buckets {
		lastMember := ""
		if resume.TagId == tagId {
			if bucketId < resume.BucketId {
				continue
			}
			if bucketId == resume.BucketId {
				lastMember = resume.LastMember
			}
		}
		// The checkpoint names the last fully processed member; every early
		// return below saves the run with it.
		e.setCheckpoint(tagId, bucketId, lastMember)
		for {
			chunk, err := e.env.getMembersFromBucket(tagId, bucketId, lastMember, e.opts.ChunkSize)
			if err != nil {
				e.saveRun()
				return fmt.Errorf("members of tag %s bucket %d: %w", tagId, bucketId, err)
			}
			if len(chunk) == 0 {
				if lastMember == "" {
					// getPopulatedBuckets just reported members here, so an
					// empty first page is a swallowed Cassandra failure
					// reading as end-of-bucket. A concurrent tag deletion
					// also lands here; a resume recomputes and moves on.
					e.saveRun()
					return &tagSyncAbort{reason: "cassandra_suspect_empty_bucket"}
				}
				break
			}
			if err := e.checkAbort(ctx); err != nil {
				e.saveRun()
				return err
			}

			// Guards run between batches, so a bad XDAS costs at most one
			// batch of pushes instead of the whole page.
			for start := 0; start < len(chunk); start += e.guardBatch {
				batch := chunk[start:min(start+e.guardBatch, len(chunk))]
				// A resume re-walks a partial batch: re-checking a member is
				// idempotent, skipping one is not.
				complete := e.processBatch(ctx, prefixedTag, batch, perTag)
				if complete {
					lastMember = batch[len(batch)-1]
					e.setCheckpoint(tagId, bucketId, lastMember)
				}
				if err := e.batchGuards(ctx, complete); err != nil {
					e.saveRun()
					return err
				}
				e.maybeSaveProgress()
			}

			if len(chunk) < e.opts.ChunkSize {
				break
			}
		}
	}
	return nil
}

// batchGuards decides whether the walk may continue past a batch: breaker
// trip, member budget, then the abort signals. An incomplete batch can only
// mean one of those fired.
func (e *tagSyncEngine) batchGuards(ctx context.Context, complete bool) error {
	if reason, tripped := e.breaker.tripped(); tripped {
		tagSyncBreakerTrippedTotal.WithLabelValues(reason).Inc()
		return &tagSyncAbort{reason: reason}
	}
	if e.memberLimitReached() {
		return errTagSyncLimitReached
	}
	if err := e.checkAbort(ctx); err != nil {
		return err
	}
	if !complete {
		return &tagSyncAbort{reason: "chunk_incomplete"}
	}
	return nil
}

func (e *tagSyncEngine) checkAbort(ctx context.Context) error {
	if ctx.Err() != nil {
		return &tagSyncAbort{reason: "cancelled"}
	}
	if !e.env.syncEnabled() {
		return &tagSyncAbort{reason: "kill_switch"}
	}
	// Past the staleness window another instance may hold this lock, and two
	// walks would put twice the configured rate on XDAS.
	if e.heartbeatStale() {
		return &tagSyncAbort{reason: "lock_heartbeat_stale"}
	}
	return nil
}

func (e *tagSyncEngine) heartbeatStale() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return time.Since(e.lastHeartbeatOk) >= tagSyncLockStaleAfter
}

// Per segment, not per run record: a resume carries earlier Checked forward,
// so the raw total would stop a restated MaxMembers after one member.
func (e *tagSyncEngine) memberLimitReached() bool {
	if e.opts.MaxMembers <= 0 {
		return false
	}
	return e.counts().Checked-e.checkedAtStart >= e.opts.MaxMembers
}

// Reports whether every member was fully processed; the caller only advances
// the checkpoint past a fully processed batch.
func (e *tagSyncEngine) processBatch(ctx context.Context, prefixedTag string, batch []string, perTag *TagMissingStat) bool {
	workers := min(e.opts.Workers, len(batch))
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var processed atomic.Int64
	memberCh := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for member := range memberCh {
				if cctx.Err() != nil {
					continue // drain remaining members without processing
				}
				if _, tripped := e.breaker.tripped(); tripped {
					cancel()
					continue
				}
				if err := e.limiter.wait(cctx); err != nil {
					continue
				}
				if e.stepMember(cctx, prefixedTag, member, perTag) {
					processed.Add(1)
				}
				if e.memberLimitReached() {
					cancel()
				}
			}
		}()
	}
	for _, member := range batch {
		memberCh <- member
	}
	close(memberCh)
	wg.Wait()
	return processed.Load() == int64(len(batch))
}

// One XDAS read, classification, and outside detect mode at most one push.
// The raw Cassandra member is normalized exactly once here. Read and push
// take a rate-limiter slot each, so the rate caps XDAS calls, not members.
//
// Returns false only when a due push was skipped by cancellation, so the
// caller keeps the earlier checkpoint and a resume retries the member.
func (e *tagSyncEngine) stepMember(ctx context.Context, prefixedTag string, rawMember string, perTag *TagMissingStat) bool {
	normalized := ToNormalizedEcm(rawMember)

	fields, err := e.env.xdasGetFields(normalized)

	const (
		classPresent = iota
		classMissingField
		classMissingKey
		classError
	)
	class := classError
	observedValue := ""
	if err == nil {
		if value, ok := fields[prefixedTag]; ok {
			class = classPresent
			observedValue = value
		} else {
			class = classMissingField
		}
		e.noteXdasOnlyFields(fields)
	} else if isXdasNotFound(err) {
		class = classMissingKey
	}

	e.mu.Lock()
	e.run.Counts.Checked++
	perTag.Checked++
	switch class {
	case classPresent:
		e.run.Counts.Present++
		e.breaker.record(syncOutcomeOk)
	case classMissingField:
		e.run.Counts.MissingField++
		perTag.Missing++
		e.breaker.record(syncOutcomeMissing)
	case classMissingKey:
		e.run.Counts.MissingKey++
		perTag.Missing++
		e.breaker.record(syncOutcomeMissing)
	case classError:
		e.run.Counts.XdasErrors++
		tagSyncXdasErrorsTotal.WithLabelValues("get").Inc()
		e.breaker.record(syncOutcomeError)
	}
	e.mu.Unlock()
	tagSyncCheckedTotal.Inc()
	if class == classError {
		return true
	}

	pushValue := ""
	pushReason := ""
	switch {
	case class == classPresent && e.opts.Mode == TagSyncModeRefresh:
		pushValue = observedValue
		pushReason = "refresh"
	case class == classMissingField && e.opts.Mode != TagSyncModeDetect:
		pushReason = "missing_field"
	case class == classMissingKey && e.opts.Mode != TagSyncModeDetect:
		pushReason = "missing_key"
	default:
		return true
	}

	if e.opts.DryRun {
		e.mu.Lock()
		e.run.Counts.WouldPush++
		e.mu.Unlock()
		return true
	}
	// The checkpoint advances either way, so a push abandoned here is drift
	// this run found, failed to fix, and never revisits.
	var pushErr error
	for attempt := 0; attempt < tagSyncPushAttempts; attempt++ {
		if err := e.limiter.wait(ctx); err != nil {
			// Cancelled waiting for the push slot: report unprocessed so the
			// resume re-checks this member.
			return false
		}
		pushErr = e.env.xdasPush(normalized, prefixedTag, pushValue)
		if pushErr == nil || !isRetryablePushError(pushErr) {
			break
		}
	}
	if pushErr != nil {
		// One outcome per member, not per attempt: the window counts members.
		e.mu.Lock()
		e.run.Counts.PushFailed++
		e.run.Counts.XdasErrors++
		e.mu.Unlock()
		tagSyncXdasErrorsTotal.WithLabelValues("push").Inc()
		e.breaker.record(syncOutcomeError)
		return true
	}
	e.mu.Lock()
	e.run.Counts.Pushed++
	perTag.Pushed++
	e.mu.Unlock()
	tagSyncPushedTotal.WithLabelValues(pushReason).Inc()
	return true
}

// A 4xx is XDAS refusing the request itself and would be refused again.
func isRetryablePushError(err error) bool {
	var remoteErr xwcommon.RemoteHttpErrorAS
	if errors.As(err, &remoteErr) {
		return remoteErr.StatusCode < 400 || remoteErr.StatusCode >= 500
	}
	return true
}

func isXdasNotFound(err error) bool {
	var remoteErr xwcommon.RemoteHttpErrorAS
	return errors.As(err, &remoteErr) && remoteErr.StatusCode == 404
}

// Counts distinct tag fields that exist only in XDAS, with no counterpart in
// Cassandra. Detection only - the job never deletes them.
func (e *tagSyncEngine) noteXdasOnlyFields(fields map[string]string) {
	// Filter before locking: knownTags is read-only during the walk, so the
	// common case never takes the engine lock.
	var xdasOnly []string
	for field := range fields {
		if !strings.HasPrefix(field, Prefix) || e.knownTags[field] {
			continue
		}
		xdasOnly = append(xdasOnly, field)
	}
	if len(xdasOnly) == 0 {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	for _, field := range xdasOnly {
		if !e.xdasOnlySeen[field] && len(e.xdasOnlySeen) < tagSyncMaxCountedXdasOnlyFields {
			e.xdasOnlySeen[field] = true
			e.run.Counts.XdasOnlyFieldsSeen++
		}
	}
}

func (e *tagSyncEngine) counts() TagSyncCounts {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.run.Counts
}

func (e *tagSyncEngine) setCheckpoint(tagId string, bucketId int, lastMember string) {
	e.mu.Lock()
	e.run.Checkpoint = TagSyncCheckpoint{TagId: tagId, BucketId: bucketId, LastMember: lastMember}
	e.mu.Unlock()
}

// Merges on TagId: a resume records the tag it stopped inside a second time,
// and appending would count one tag as two. Only a tag with missing members
// earns an entry, but once listed it keeps accumulating.
func (e *tagSyncEngine) recordTagResult(perTag *TagMissingStat) {
	e.mu.Lock()
	if existing := e.findMissingStatLocked(perTag.TagId); existing != nil {
		existing.Checked += perTag.Checked
		existing.Missing += perTag.Missing
		existing.Pushed += perTag.Pushed
	} else if perTag.Missing > 0 {
		e.run.TagsWithMissing++
		e.missingStats = append(e.missingStats, *perTag)
		if len(e.missingStats) > 2*tagSyncTopMissingKeep {
			e.missingStats = topMissing(e.missingStats, e.run.Checkpoint.TagId)
		}
	}
	e.mu.Unlock()

	if perTag.Missing == 0 {
		return
	}
	// This segment's numbers, not the merged total in the run record.
	e.logf(log.InfoLevel, "tag sync: tag %s has missing members: checked=%d missing=%d pushed=%d",
		perTag.TagId, perTag.Checked, perTag.Missing, perTag.Pushed)
}

// Hold e.mu, and do not append to e.missingStats while using the pointer.
func (e *tagSyncEngine) findMissingStatLocked(tagId string) *TagMissingStat {
	for i := range e.missingStats {
		if e.missingStats[i].TagId == tagId {
			return &e.missingStats[i]
		}
	}
	return nil
}

// Saves at most once per checkpoint interval, logs at most once per minute.
// The lock heartbeat is deliberately not on this cadence - heartbeatLoop owns
// it on wall clock, so a slow chunk cannot let the lock go stale.
func (e *tagSyncEngine) maybeSaveProgress() {
	interval := time.Duration(e.env.config.CheckpointIntervalSecs) * time.Second
	now := time.Now()

	e.mu.Lock()
	shouldSave := now.Sub(e.lastSave) >= interval
	if shouldSave {
		e.lastSave = now
	}
	shouldLog := now.Sub(e.lastProgress) >= tagSyncProgressEvery
	if shouldLog {
		e.lastProgress = now
	}
	e.mu.Unlock()

	if shouldSave {
		e.saveRun()
	}
	if shouldLog {
		counts := e.counts()
		e.mu.Lock()
		cp := e.run.Checkpoint
		tagsDone, tagsTotal := e.run.TagsDone, e.run.TagsTotal
		e.mu.Unlock()
		e.logf(log.InfoLevel, "tag sync progress: tags=%d/%d position=%s/bucket=%d checked=%d present=%d missing=%d pushed=%d xdasErrors=%d",
			tagsDone, tagsTotal, cp.TagId, cp.BucketId, counts.Checked, counts.Present,
			counts.MissingField+counts.MissingKey, counts.Pushed, counts.XdasErrors)
	}
}

func (e *tagSyncEngine) saveRun() TagSyncRun {
	e.mu.Lock()
	e.run.UpdatedAt = time.Now().UTC()
	// Refreshed on every save so mid-run status shows live numbers.
	e.run.TopMissingTags = topMissing(e.missingStats, e.run.Checkpoint.TagId)
	e.run.MissingRate = missingRate(e.run.Counts)
	snapshot := *e.run
	e.mu.Unlock()

	tagSyncMissingRate.Set(snapshot.MissingRate)
	if err := e.env.dao.saveRun(&snapshot); err != nil {
		e.logf(log.WarnLevel, "tag sync run save failed: %v", err)
	}
	return snapshot
}

func missingRate(c TagSyncCounts) float64 {
	if c.Checked == 0 {
		return 0
	}
	return float64(c.MissingField+c.MissingKey) / float64(c.Checked)
}

// Worst first, input not mutated. The tag the checkpoint sits in is kept
// whatever its rank: a resume seeds from this list, and a tag trimmed off it
// comes back as a second entry, counted as a second tag with missing members.
func topMissing(missingStats []TagMissingStat, keepTagId string) []TagMissingStat {
	out := append([]TagMissingStat(nil), missingStats...)
	sort.Slice(out, func(i, j int) bool { return out[i].Missing > out[j].Missing })
	if len(out) <= tagSyncTopMissingKeep {
		return out
	}
	kept := out[:tagSyncTopMissingKeep:tagSyncTopMissingKeep]
	if keepTagId == "" || containsTagStat(kept, keepTagId) {
		return kept
	}
	for _, stat := range out[tagSyncTopMissingKeep:] {
		if stat.TagId == keepTagId {
			return append(kept, stat)
		}
	}
	return kept
}

func containsTagStat(stats []TagMissingStat, tagId string) bool {
	for _, stat := range stats {
		if stat.TagId == tagId {
			return true
		}
	}
	return false
}

func (e *tagSyncEngine) finishCompleted(limited bool) {
	run := e.finalize(TagSyncStateCompleted, "", limited)
	counts := run.Counts
	e.logf(log.InfoLevel, "tag sync run completed: limited=%v checked=%d present=%d missingField=%d missingKey=%d pushed=%d pushFailed=%d xdasErrors=%d xdasOnlyFields=%d missingRate=%.4f tagsWithMissing=%d",
		limited, counts.Checked, counts.Present, counts.MissingField, counts.MissingKey,
		counts.Pushed, counts.PushFailed, counts.XdasErrors, counts.XdasOnlyFieldsSeen,
		run.MissingRate, run.TagsWithMissing)
	if counts.PushFailed > 0 {
		// completed is not proof the drift is closed.
		e.logf(log.WarnLevel, "tag sync: %d member(s) stayed unpushed after %d attempts each; the run completed but did not close their drift - re-run the same mode over the affected tags to retry them",
			counts.PushFailed, tagSyncPushAttempts)
	}
}

func (e *tagSyncEngine) finishAborted(reason string) {
	run := e.finalize(TagSyncStateAborted, reason, false)
	counts := run.Counts
	e.logf(log.WarnLevel, "tag sync run aborted: reason=%s checked=%d missing=%d pushed=%d checkpoint=%s/%d/%s",
		reason, counts.Checked, counts.MissingField+counts.MissingKey, counts.Pushed,
		run.Checkpoint.TagId, run.Checkpoint.BucketId, run.Checkpoint.LastMember)
}

// Returns the record as persisted, so the caller can log it without reading
// e.run outside the mutex.
func (e *tagSyncEngine) finalize(state string, abortReason string, limited bool) TagSyncRun {
	now := time.Now().UTC()
	e.mu.Lock()
	e.run.State = state
	e.run.AbortReason = abortReason
	e.run.Limited = limited
	e.run.CompletedAt = &now
	e.mu.Unlock()

	snapshot := e.saveRun()
	// Run ids are time-prefixed, so lexical pruning keeps the newest.
	if err := e.env.dao.pruneRuns(tagSyncRunHistoryKeep); err != nil {
		e.logf(log.WarnLevel, "tag sync run history prune failed: %v", err)
	}
	return snapshot
}

func (e *tagSyncEngine) logf(level log.Level, format string, args ...interface{}) {
	entry := log.WithFields(log.Fields{
		"audit_id": e.run.RunId,
		"job":      "tag_sync",
		"mode":     string(e.opts.Mode),
	})
	switch level {
	case log.WarnLevel:
		entry.Warnf(format, args...)
	case log.ErrorLevel:
		entry.Errorf(format, args...)
	default:
		entry.Infof(format, args...)
	}
}
