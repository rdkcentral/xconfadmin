package tag

import (
	"context"
	"errors"
	"fmt"
	"os"
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
// line with it. Three modes share the same walk:
//
//	detect  - read-only census: count members present/missing in XDAS
//	repair  - push back only the members missing in XDAS
//	refresh - re-push every member, preserving the observed value; every
//	          push carries the TTL header, so this resets XDAS TTLs for the
//	          whole population including inactive devices
//
// The job is additive-only: it never deletes anything from either store.
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

	// Internal bounds, deliberately not config: they cap memory and noise,
	// not behavior an operator would tune per deployment.

	// Entries kept on the topMissingTags leaderboard; the run record is one
	// JSON cell re-read by every status poll, so it must stay small. Every
	// tag with missing members still gets an exact log line.
	tagSyncTopMissingKeep = 100
	// Distinct XDAS-only field names remembered for dedup; past the cap new
	// ones stop being counted (undercount, never double-count).
	tagSyncMaxCountedXdasOnlyFields = 10000
	// Progress log line throttle - independent of checkpoint saves.
	tagSyncProgressEvery = time.Minute
	// Ring of members this run recently saw present in XDAS - probes that
	// cannot have rotted, refreshed by the walk itself.
	tagSyncProbePoolSize = 10
	// Probe GETs one confirmation may spend (operator probe first, then
	// pool) before a high missing rate is treated as an outage; also bounds
	// the preflight retries on transient errors.
	tagSyncProbeAttempts = 3
	// Run records surviving pruning at finalize; the status endpoint shows
	// 10, the rest is headroom while keeping the partition scan bounded.
	tagSyncRunHistoryKeep = 20
)

var (
	tagSyncLockStaleAfter = 5 * time.Minute
	// After writing the lock we wait this long and re-read it, so that of two
	// near-simultaneous writers exactly one survives (last write wins in
	// Cassandra; both re-readers then see the same winner).
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
	// ProbeMember overrides the configured probe for this run, so a rotted
	// probe device never requires a config rollout.
	ProbeMember string `json:"probeMember,omitempty"`
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

// TagSyncRun is the persisted record of one run: options, live checkpoint,
// counters, and the final report. It is saved on every checkpoint interval,
// so it doubles as the progress view for the status endpoint.
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
	// MissingRateUnconfirmed marks a detect census that crossed the missing
	// rate threshold with no probe available to tell genuinely missing
	// members from an XDAS outage; the numbers need manual confirmation.
	MissingRateUnconfirmed bool `json:"missingRateUnconfirmed,omitempty"`
}

// tagSyncBusyError is returned when another run holds the lock; the trigger
// endpoint maps it to 409.
type tagSyncBusyError struct {
	RunId string
	Owner string
}

func (e *tagSyncBusyError) Error() string {
	return fmt.Sprintf("tag sync already running: runId=%s owner=%s", e.RunId, e.Owner)
}

// tagSyncAbort carries the reason a run stopped early through the walk's
// error path; errTagSyncLimitReached marks the clean MaxMembers stop.
type tagSyncAbort struct {
	reason string
}

func (e *tagSyncAbort) Error() string { return "tag sync aborted: " + e.reason }

var errTagSyncLimitReached = errors.New("tag sync member limit reached")

// tagSyncEnv is the seam between the engine and the outside world; tests
// swap in fakes, production wires the real primitives below.
type tagSyncEnv struct {
	getAllTagIds         func() ([]string, error)
	getPopulatedBuckets  func(tagId string) ([]int, error)
	getMembersFromBucket func(tagId string, bucketId int, lastMember string, limit int) ([]string, error)
	// xdasGetFields returns every field of the member's XDAS record; the
	// member must already be normalized.
	xdasGetFields func(normalizedMember string) (map[string]string, error)
	xdasPush      func(normalizedMember string, prefixedTag string, value string) error
	dao           tagSyncDao
	syncEnabled   func() bool
	config        *taggingapi_config.TagSyncConfig
}

func newTagSyncEnv() (*tagSyncEnv, error) {
	if xhttp.WebConfServer == nil || xhttp.WebConfServer.TagSyncConfig == nil {
		return nil, errors.New("tag sync: server not initialized")
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

// tagSyncKillSwitchEnabled reads the TaggingSyncEnabled app setting through
// the shared helper (which tolerates string-typed booleans an operator may
// PUT); absent or unreadable means enabled. Other instances see a flip after
// their cache refresh, so a cross-instance stop takes effect within about a
// minute.
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

	mu                sync.Mutex
	missingStats      []TagMissingStat
	lastSave          time.Time
	lastProgress      time.Time
	ignoreMissingRate bool
	// probePool holds the last few members this run observed present in
	// XDAS. They are the preferred probe candidates: unlike a configured
	// probe device, the pool cannot rot away, because the walk itself
	// refreshes it.
	probePool    []string
	probePoolIdx int
}

// RunTagSync runs a full sync cycle synchronously; it is the entry point for
// a future one-shot binary. HTTP triggering uses PrepareTagSync + Execute so
// the run id can be returned before the walk starts.
func RunTagSync(ctx context.Context, opts TagSyncOptions) (*TagSyncRun, error) {
	engine, err := PrepareTagSync(opts)
	if err != nil {
		return nil, err
	}
	return engine.Execute(ctx), nil
}

// PrepareTagSync validates options, takes the cross-instance lock and saves
// the initial run record. The returned engine must be driven with Execute
// (which releases the lock on every path).
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

func validateTagSyncOptions(opts *TagSyncOptions) error {
	if opts.Mode == "" {
		opts.Mode = TagSyncModeDetect
	}
	switch opts.Mode {
	case TagSyncModeDetect, TagSyncModeRepair, TagSyncModeRefresh:
		return nil
	default:
		return fmt.Errorf("invalid mode %q: must be detect, repair or refresh", opts.Mode)
	}
}

// positiveOr returns v when positive, otherwise fallback, floored to 1: a
// zero must never reach the walk (zero workers would leave the unbuffered
// member channel with no receivers, hanging the run while it holds the lock).
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
		run = resumed
		run.State = TagSyncStateRunning
		run.AbortReason = ""
		run.CompletedAt = nil
		run.Limited = false
		run.Owner = owner
		run.Resumes++
		// Mode and filters stay as recorded; pacing may be overridden.
		run.Options.Rate = opts.Rate
		run.Options.Workers = opts.Workers
		run.Options.ChunkSize = opts.ChunkSize
		run.Options.DryRun = opts.DryRun
		run.Options.MaxMembers = opts.MaxMembers
		if opts.ProbeMember != "" {
			run.Options.ProbeMember = opts.ProbeMember
		}
		opts = run.Options
		opts.Resume = true
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
		return nil, fmt.Errorf("tag sync run save failed: %w", err)
	}

	return &tagSyncEngine{
		env:          env,
		opts:         opts,
		run:          run,
		limiter:      newRateLimiter(opts.Rate),
		breaker:      newSyncBreaker(cfg.BreakerWindow, cfg.BreakerMinSample, cfg.BreakerErrorRatePercent, cfg.BreakerMissingRatePercent, cfg.BreakerMaxConsecErrors),
		xdasOnlySeen: make(map[string]bool),
		missingStats: append([]TagMissingStat(nil), run.TopMissingTags...),
	}, nil
}

func findResumableRun(dao tagSyncDao) (*TagSyncRun, error) {
	runs, err := dao.listRuns(10)
	if err != nil {
		return nil, fmt.Errorf("tag sync run list failed: %w", err)
	}
	for _, run := range runs {
		// A "running" record whose lock went stale is a crashed run; it is
		// resumable because acquireTagSyncLock only succeeds on stale locks.
		// A completed run that stopped at its MaxMembers limit still has a
		// live checkpoint, so the ramp-then-continue workflow can resume it.
		if run.State == TagSyncStateAborted || run.State == TagSyncStateRunning ||
			(run.State == TagSyncStateCompleted && run.Limited) {
			return run, nil
		}
	}
	return nil, errors.New("no aborted tag sync run to resume")
}

func acquireTagSyncLock(dao tagSyncDao, owner string, runId string) error {
	existing, err := dao.getLock()
	if err != nil {
		return fmt.Errorf("tag sync lock read failed: %w", err)
	}
	if existing != nil && !existing.Released && time.Since(existing.HeartbeatAt) < tagSyncLockStaleAfter {
		return &tagSyncBusyError{RunId: existing.RunId, Owner: existing.Owner}
	}
	mine := &TagSyncLock{Owner: owner, RunId: runId, HeartbeatAt: time.Now().UTC()}
	if err := dao.saveLock(mine); err != nil {
		return fmt.Errorf("tag sync lock write failed: %w", err)
	}
	time.Sleep(tagSyncLockSettle)
	current, err := dao.getLock()
	if err != nil {
		return fmt.Errorf("tag sync lock re-read failed: %w", err)
	}
	if current == nil || current.Owner != owner || current.RunId != runId {
		return &tagSyncBusyError{RunId: lockRunId(current), Owner: lockOwner(current)}
	}
	return nil
}

func lockRunId(l *TagSyncLock) string {
	if l == nil {
		return ""
	}
	return l.RunId
}

func lockOwner(l *TagSyncLock) string {
	if l == nil {
		return ""
	}
	return l.Owner
}

func releaseTagSyncLock(dao tagSyncDao, owner string, runId string) {
	err := dao.saveLock(&TagSyncLock{Owner: owner, RunId: runId, HeartbeatAt: time.Now().UTC(), Released: true})
	if err != nil {
		log.Errorf("tag sync lock release failed (goes stale in %v): %v", tagSyncLockStaleAfter, err)
	}
}

// Execute drives the full walk. It never returns an error: every outcome
// (completed, aborted, breaker trip) is recorded on the returned run.
func (e *tagSyncEngine) Execute(ctx context.Context) *TagSyncRun {
	tagSyncRunningGauge.Set(1)
	start := time.Now()
	hbStop := make(chan struct{})
	hbDone := make(chan struct{})
	go e.heartbeatLoop(hbStop, hbDone)
	defer func() {
		// The heartbeat must be fully stopped before the release write, or an
		// in-flight heartbeat could resurrect the lock as live-unreleased.
		close(hbStop)
		<-hbDone
		releaseTagSyncLock(e.env.dao, e.run.Owner, e.run.RunId)
		tagSyncRunningGauge.Set(0)
		tagSyncRunDurationSeconds.Set(time.Since(start).Seconds())
	}()

	e.logf(log.InfoLevel, "tag sync run started: mode=%s dryRun=%v rate=%d workers=%d chunkSize=%d resume=%v",
		e.opts.Mode, e.opts.DryRun, e.opts.Rate, e.opts.Workers, e.opts.ChunkSize, e.opts.Resume)

	allTags, err := e.env.getAllTagIds()
	if err != nil {
		e.finishAborted("cassandra_error: " + err.Error())
		return e.run
	}
	if len(allTags) == 0 {
		// An empty census is indistinguishable from a swallowed Cassandra
		// failure (the shared query helper reports errors as empty rows), and
		// this system always has tags - abort instead of recording a false
		// all-clear with missingRate=0.
		e.finishAborted("cassandra_suspect_no_tags")
		return e.run
	}
	e.knownTags = make(map[string]bool, len(allTags))
	for _, tagId := range allTags {
		e.knownTags[SetTagPrefix(tagId)] = true
	}

	tags := filterTags(allTags, e.opts.Tags)
	sort.Strings(tags)
	e.run.TagsTotal = len(tags)
	e.run.TagsDone = 0

	if e.opts.Mode != TagSyncModeDetect {
		if err := e.preflightProbe(ctx); err != nil {
			var abort *tagSyncAbort
			if errors.As(err, &abort) {
				e.finishAborted(abort.reason)
			} else {
				e.finishAborted("probe_error: " + err.Error())
			}
			return e.run
		}
	}

	resumeCp := TagSyncCheckpoint{}
	if e.opts.Resume {
		resumeCp = e.run.Checkpoint
	}

	for _, tagId := range tags {
		if resumeCp.TagId != "" && tagId < resumeCp.TagId {
			e.run.TagsDone++
			continue
		}
		perTag := &TagMissingStat{TagId: tagId}
		err := e.walkTag(ctx, tagId, perTag, resumeCp)
		if tagId >= resumeCp.TagId {
			resumeCp = TagSyncCheckpoint{}
		}
		if err != nil {
			if errors.Is(err, errTagSyncLimitReached) {
				e.recordTagResult(perTag)
				e.finishCompleted(true)
				return e.run
			}
			var abort *tagSyncAbort
			if errors.As(err, &abort) {
				e.finishAborted(abort.reason)
			} else {
				e.finishAborted("cassandra_error: " + err.Error())
			}
			return e.run
		}
		e.recordTagResult(perTag)
		e.run.TagsDone++
	}

	e.finishCompleted(false)
	return e.run
}

func filterTags(all []string, requested []string) []string {
	if len(requested) == 0 {
		return append([]string(nil), all...)
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

// preflightProbe blocks write modes when the configured known-good member
// cannot be read back from XDAS; without that assurance a broken or empty
// XDAS view would trigger a flood of pointless (or value-clobbering) pushes.
// A clean 404 (the probe rotted out of XDAS) fails immediately; transient
// transport/5xx errors are retried a few times before giving up.
func (e *tagSyncEngine) preflightProbe(ctx context.Context) error {
	probe := e.configuredProbe()
	if probe == "" {
		e.logf(log.WarnLevel, "tag sync: no probeMember passed with the trigger; %s mode runs without the preflight probe check", e.opts.Mode)
		return nil
	}
	normalized := ToNormalizedEcm(probe)
	var lastErr error
	for attempt := 0; attempt < tagSyncProbeAttempts; attempt++ {
		if err := e.limiter.wait(ctx); err != nil {
			return &tagSyncAbort{reason: "cancelled"}
		}
		fields, err := e.env.xdasGetFields(normalized)
		if err == nil {
			if len(fields) == 0 {
				return &tagSyncAbort{reason: "probe_member_not_readable"}
			}
			return nil
		}
		if isXdasNotFound(err) {
			return &tagSyncAbort{reason: "probe_member_not_readable"}
		}
		lastErr = err
	}
	e.logf(log.WarnLevel, "tag sync: preflight probe kept erroring: %v", lastErr)
	return &tagSyncAbort{reason: "probe_member_not_readable"}
}

// heartbeatLoop refreshes the lock heartbeat on a wall-clock cadence,
// independent of chunk pacing: a chunk slower than the staleness window must
// not let another instance treat the lock as stale and start a second run.
func (e *tagSyncEngine) heartbeatLoop(stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(tagSyncLockStaleAfter / 3)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			lock := &TagSyncLock{Owner: e.run.Owner, RunId: e.run.RunId, HeartbeatAt: time.Now().UTC()}
			if err := e.env.dao.saveLock(lock); err != nil {
				e.logf(log.WarnLevel, "tag sync heartbeat save failed: %v", err)
			}
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
		for {
			if err := e.checkAbort(ctx); err != nil {
				e.setCheckpoint(tagId, bucketId, lastMember)
				e.saveRun()
				return err
			}
			chunk, err := e.env.getMembersFromBucket(tagId, bucketId, lastMember, e.opts.ChunkSize)
			if err != nil {
				e.setCheckpoint(tagId, bucketId, lastMember)
				e.saveRun()
				return fmt.Errorf("members of tag %s bucket %d: %w", tagId, bucketId, err)
			}
			if len(chunk) == 0 {
				if lastMember == "" {
					// getPopulatedBuckets just reported members here, so an
					// empty first page smells like a swallowed Cassandra
					// failure reading as end-of-bucket - abort rather than
					// silently skip the bucket. A concurrent tag deletion can
					// also trigger this; a resume then recomputes the bucket
					// list and moves on.
					e.setCheckpoint(tagId, bucketId, "")
					e.saveRun()
					return &tagSyncAbort{reason: "cassandra_suspect_empty_bucket"}
				}
				break
			}

			completed := e.processChunk(ctx, prefixedTag, chunk, perTag)
			if !completed {
				// Part of this chunk was drained unprocessed (breaker trip,
				// cancel, or member limit). The checkpoint stays at the chunk
				// start so a resume re-walks the whole chunk; re-checking a
				// member is idempotent, skipping one is not.
				e.setCheckpoint(tagId, bucketId, lastMember)
				e.saveRun()
				if reason, tripped := e.breaker.tripped(); tripped {
					tagSyncBreakerTrippedTotal.WithLabelValues(reason).Inc()
					return &tagSyncAbort{reason: reason}
				}
				if e.opts.MaxMembers > 0 && e.counts().Checked >= e.opts.MaxMembers {
					return errTagSyncLimitReached
				}
				if err := e.checkAbort(ctx); err != nil {
					return err
				}
				return &tagSyncAbort{reason: "chunk_incomplete"}
			}

			lastMember = chunk[len(chunk)-1]
			e.setCheckpoint(tagId, bucketId, lastMember)
			e.maybeSaveProgress()

			if reason, tripped := e.breaker.tripped(); tripped {
				tagSyncBreakerTrippedTotal.WithLabelValues(reason).Inc()
				e.saveRun()
				return &tagSyncAbort{reason: reason}
			}
			if e.breaker.missingRateHigh() && !e.ignoreMissingRate {
				if err := e.confirmMissingWithProbe(ctx); err != nil {
					e.saveRun()
					return err
				}
			}
			if e.opts.MaxMembers > 0 && e.counts().Checked >= e.opts.MaxMembers {
				e.saveRun()
				return errTagSyncLimitReached
			}
			if len(chunk) < e.opts.ChunkSize {
				break
			}
		}
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
	return nil
}

// configuredProbe returns the probe member passed with this run's trigger,
// if any. There is deliberately no config-file probe: a device pinned in
// deployment config rots silently (its XDAS entry expires like any other),
// while a trigger-time choice is fresh by construction.
func (e *tagSyncEngine) configuredProbe() string {
	return e.opts.ProbeMember
}

// probeCandidates tries the operator's trigger-time probe first - it is an
// explicit "this device must be in XDAS" assertion for this run and must
// never be shadowed by pool entries - then members this run itself recently
// saw present.
func (e *tagSyncEngine) probeCandidates() []string {
	var candidates []string
	if probe := e.configuredProbe(); probe != "" {
		candidates = append(candidates, ToNormalizedEcm(probe))
	}
	e.mu.Lock()
	candidates = append(candidates, e.probePool...)
	e.mu.Unlock()
	return candidates
}

// confirmMissingWithProbe distinguishes genuine mass expiry from an XDAS
// outage that answers 404 for everything: any known-good member still present
// means the missing members are real and the run may continue.
func (e *tagSyncEngine) confirmMissingWithProbe(ctx context.Context) error {
	candidates := e.probeCandidates()
	if len(candidates) == 0 {
		if e.opts.Mode == TagSyncModeDetect {
			// The read-only census has nothing to confirm with, but also
			// nothing to protect: continue and flag the report instead of
			// blocking the census in the exact mass-expiry scenario it
			// exists for. Later chunks retry - the pool may fill and confirm.
			e.mu.Lock()
			alreadyFlagged := e.run.MissingRateUnconfirmed
			e.run.MissingRateUnconfirmed = true
			e.mu.Unlock()
			if !alreadyFlagged {
				e.logf(log.WarnLevel, "tag sync: missing rate exceeded %d%% with no probe available; detect continues with missingRateUnconfirmed",
					e.env.config.BreakerMissingRatePercent)
			}
			return nil
		}
		tagSyncBreakerTrippedTotal.WithLabelValues("missing_rate_no_probe_available").Inc()
		return &tagSyncAbort{reason: "missing_rate_high_no_probe_available"}
	}
	attempts := 0
	for _, candidate := range candidates {
		if attempts >= tagSyncProbeAttempts {
			break
		}
		attempts++
		if err := e.limiter.wait(ctx); err != nil {
			return &tagSyncAbort{reason: "cancelled"}
		}
		fields, err := e.env.xdasGetFields(candidate)
		if err == nil && len(fields) > 0 {
			e.ignoreMissingRate = true
			e.breaker.clearMissingHigh()
			e.mu.Lock()
			e.run.MissingRateUnconfirmed = false
			e.mu.Unlock()
			e.logf(log.WarnLevel, "tag sync: missing rate exceeded %d%% but probe member %s is present in XDAS - the missing members are real, continuing",
				e.env.config.BreakerMissingRatePercent, candidate)
			return nil
		}
	}
	tagSyncBreakerTrippedTotal.WithLabelValues("missing_rate_probe_failed").Inc()
	return &tagSyncAbort{reason: "missing_rate_high_probe_failed"}
}

// processChunk fans the chunk through the worker pool and reports whether
// every member was fully processed; the caller only advances the checkpoint
// past a fully processed chunk.
func (e *tagSyncEngine) processChunk(ctx context.Context, prefixedTag string, chunk []string, perTag *TagMissingStat) bool {
	workers := min(e.opts.Workers, len(chunk))
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
				if e.opts.MaxMembers > 0 && e.counts().Checked >= e.opts.MaxMembers {
					cancel()
				}
			}
		}()
	}
	for _, member := range chunk {
		memberCh <- member
	}
	close(memberCh)
	wg.Wait()
	return processed.Load() == int64(len(chunk))
}

// stepMember is the per-member unit of work: one XDAS read, classification,
// and (outside detect mode) at most one push. The member arrives in its raw
// Cassandra form and is normalized exactly once here. The caller pays the
// rate-limiter slot for the read; the push takes its own slot below, so the
// configured rate caps total XDAS calls, not members.
//
// The return value reports whether the member's work fully happened: false
// only when a due push was skipped by cancellation, so the caller keeps the
// checkpoint before this chunk and a resume retries the member.
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
		if len(e.probePool) < tagSyncProbePoolSize {
			e.probePool = append(e.probePool, normalized)
		} else {
			e.probePool[e.probePoolIdx%tagSyncProbePoolSize] = normalized
		}
		e.probePoolIdx++
	case classMissingField:
		e.run.Counts.MissingField++
		perTag.Missing++
	case classMissingKey:
		e.run.Counts.MissingKey++
		perTag.Missing++
	case classError:
		e.run.Counts.XdasErrors++
	}
	e.mu.Unlock()

	tagSyncCheckedTotal.Inc()
	switch class {
	case classPresent:
		e.breaker.record(syncOutcomeOk)
	case classMissingField, classMissingKey:
		e.breaker.record(syncOutcomeMissing)
	case classError:
		tagSyncXdasErrorsTotal.WithLabelValues("get").Inc()
		e.breaker.record(syncOutcomeError)
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
	if err := e.limiter.wait(ctx); err != nil {
		// Cancelled while waiting for the push slot: report the member as
		// unprocessed so the chunk does not count as complete and the resume
		// re-checks it.
		return false
	}
	if err := e.env.xdasPush(normalized, prefixedTag, pushValue); err != nil {
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

func isXdasNotFound(err error) bool {
	var remoteErr xwcommon.RemoteHttpErrorAS
	return errors.As(err, &remoteErr) && remoteErr.StatusCode == 404
}

// noteXdasOnlyFields counts distinct tag fields that exist only in XDAS,
// with no counterpart tag in Cassandra (detection only - the job never
// deletes them). The distinct count is capped by the tracking set: beyond
// tagSyncMaxCountedXdasOnlyFields new fields stop being counted rather than
// silently double-counted.
func (e *tagSyncEngine) noteXdasOnlyFields(fields map[string]string) {
	for field := range fields {
		if !strings.HasPrefix(field, Prefix) || e.knownTags[field] {
			continue
		}
		e.mu.Lock()
		if !e.xdasOnlySeen[field] && len(e.xdasOnlySeen) < tagSyncMaxCountedXdasOnlyFields {
			e.xdasOnlySeen[field] = true
			e.run.Counts.XdasOnlyFieldsSeen++
		}
		e.mu.Unlock()
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

func (e *tagSyncEngine) recordTagResult(perTag *TagMissingStat) {
	if perTag.Missing == 0 {
		return
	}
	e.mu.Lock()
	e.run.TagsWithMissing++
	e.missingStats = append(e.missingStats, *perTag)
	if len(e.missingStats) > 2*tagSyncTopMissingKeep {
		e.trimMissingLocked()
	}
	e.mu.Unlock()
	e.logf(log.InfoLevel, "tag sync: tag %s has missing members: checked=%d missing=%d pushed=%d",
		perTag.TagId, perTag.Checked, perTag.Missing, perTag.Pushed)
}

func (e *tagSyncEngine) trimMissingLocked() {
	sort.Slice(e.missingStats, func(i, j int) bool { return e.missingStats[i].Missing > e.missingStats[j].Missing })
	if len(e.missingStats) > tagSyncTopMissingKeep {
		e.missingStats = e.missingStats[:tagSyncTopMissingKeep]
	}
}

// maybeSaveProgress persists the run record at most once per checkpoint
// interval and emits a progress line at most once per minute. The lock
// heartbeat is NOT tied to this cadence - heartbeatLoop owns it on wall
// clock, so a slow chunk cannot let the lock go stale.
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

func (e *tagSyncEngine) saveRun() {
	e.mu.Lock()
	e.run.UpdatedAt = time.Now().UTC()
	// Refresh the missing-members leaderboard on every save so mid-run status
	// shows live numbers and a crash loses at most one checkpoint interval.
	e.run.TopMissingTags = topMissing(e.missingStats)
	snapshot := *e.run
	e.mu.Unlock()
	if err := e.env.dao.saveRun(&snapshot); err != nil {
		e.logf(log.WarnLevel, "tag sync run save failed: %v", err)
	}
}

// topMissing returns the tags with the most missing members, worst first,
// without mutating the input.
func topMissing(missingStats []TagMissingStat) []TagMissingStat {
	out := append([]TagMissingStat(nil), missingStats...)
	sort.Slice(out, func(i, j int) bool { return out[i].Missing > out[j].Missing })
	if len(out) > tagSyncTopMissingKeep {
		out = out[:tagSyncTopMissingKeep]
	}
	return out
}

func (e *tagSyncEngine) finishCompleted(limited bool) {
	e.finalize(TagSyncStateCompleted, "", limited)
	counts := e.counts()
	e.logf(log.InfoLevel, "tag sync run completed: limited=%v checked=%d present=%d missingField=%d missingKey=%d pushed=%d pushFailed=%d xdasErrors=%d xdasOnlyFields=%d missingRate=%.4f tagsWithMissing=%d",
		limited, counts.Checked, counts.Present, counts.MissingField, counts.MissingKey,
		counts.Pushed, counts.PushFailed, counts.XdasErrors, counts.XdasOnlyFieldsSeen,
		e.run.MissingRate, e.run.TagsWithMissing)
}

func (e *tagSyncEngine) finishAborted(reason string) {
	e.finalize(TagSyncStateAborted, reason, false)
	counts := e.counts()
	e.logf(log.WarnLevel, "tag sync run aborted: reason=%s checked=%d missing=%d pushed=%d checkpoint=%s/%d/%s",
		reason, counts.Checked, counts.MissingField+counts.MissingKey, counts.Pushed,
		e.run.Checkpoint.TagId, e.run.Checkpoint.BucketId, e.run.Checkpoint.LastMember)
}

func (e *tagSyncEngine) finalize(state string, abortReason string, limited bool) {
	now := time.Now().UTC()
	e.mu.Lock()
	e.run.State = state
	e.run.AbortReason = abortReason
	e.run.Limited = limited
	e.run.CompletedAt = &now
	if e.run.Counts.Checked > 0 {
		e.run.MissingRate = float64(e.run.Counts.MissingField+e.run.Counts.MissingKey) / float64(e.run.Counts.Checked)
	}
	e.mu.Unlock()

	tagSyncMissingRate.Set(e.run.MissingRate)
	e.saveRun()
	// Run ids are time-prefixed, so pruning by lexical order keeps the newest
	// records and stops the history partition from growing without bound.
	if err := e.env.dao.pruneRuns(tagSyncRunHistoryKeep); err != nil {
		e.logf(log.WarnLevel, "tag sync run history prune failed: %v", err)
	}
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
