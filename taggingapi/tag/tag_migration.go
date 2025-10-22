package tag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/util"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwtagging "github.com/rdkcentral/xconfwebconfig/tagging"

	log "github.com/sirupsen/logrus"
)

// Migration command constants
const (
	CommandDryRun = "dryRun"
	CommandStart  = "start"
	CommandCancel = "cancel"
)

// Skip reason constants
const (
	ReasonTagNotFound         = "tag_not_found_in_v1"
	ReasonNoMembers           = "no_members"
	ReasonTypeConversion      = "type_conversion_error"
	ReasonXdasVerification    = "xdas_verification_failed"
	ReasonDatabaseWriteFailed = "database_write_failed"
)

// Migration job state constants
type MigrationJobState string

const (
	StateIdle      MigrationJobState = "idle"
	StateRunning   MigrationJobState = "running"
	StateCompleted MigrationJobState = "completed"
	StateFailed    MigrationJobState = "failed"
	StateCancelled MigrationJobState = "cancelled"
)

// DryRunResponse represents the response for a dry run migration
type DryRunResponse struct {
	Summary  DryRunSummary  `json:"summary"`
	Tags     map[string]int `json:"tags"`
	Warnings []string       `json:"warnings,omitempty"`
}

// DryRunSummary provides overall statistics for dry run
type DryRunSummary struct {
	TotalTags           int `json:"totalTags"`
	TotalMembersInXconf int `json:"totalMembersInXconf"`
}

// MigrationResponse represents the response for an actual migration
type MigrationResponse struct {
	Status          string             `json:"status"`
	Summary         MigrationSummary   `json:"summary"`
	FailedTags      []string           `json:"failedTags,omitempty"`
	SkippedTags     []TagMigrationSkip `json:"skippedTags,omitempty"`
	StartTime       string             `json:"startTime"`
	EndTime         string             `json:"endTime"`
	DurationSeconds float64            `json:"durationSeconds"`
}

// MigrationSummary provides overall statistics for migration
type MigrationSummary struct {
	TotalTags                 int `json:"totalTags"`
	SuccessCount              int `json:"successCount"`
	FailureCount              int `json:"failureCount"`
	SkippedCount              int `json:"skippedCount"`
	TotalMembers              int `json:"totalMembers"`
	TotalMembersWritten       int `json:"totalMembersWritten"`
	TotalMembersMissingInXdas int `json:"totalMembersMissingInXdas"`
}

// TagMigrationSkip represents a skipped tag with reason
type TagMigrationSkip struct {
	TagId  string `json:"tagId"`
	Reason string `json:"reason"`
}

// TagMemberStats represents member statistics for a tag migration
type TagMemberStats struct {
	TotalMembers         int
	MembersWritten       int
	MembersMissingInXdas int
}

// MigrationProgress tracks real-time progress of migration
type MigrationProgress struct {
	TotalTags                      int    `json:"totalTags"`
	ProcessedTags                  int    `json:"processedTags"`
	CurrentTag                     string `json:"currentTag,omitempty"`
	SuccessCount                   int    `json:"successCount"`
	FailureCount                   int    `json:"failureCount"`
	SkippedCount                   int    `json:"skippedCount"`
	CurrentTagTotalMembers         int    `json:"currentTagTotalMembers,omitempty"`
	CurrentTagMembersWritten       int    `json:"currentTagMembersWritten,omitempty"`
	CurrentTagMembersMissingInXdas int    `json:"currentTagMembersMissingInXdas,omitempty"`
}

// MigrationJobStatus represents the current state of migration job
type MigrationJobStatus struct {
	State           MigrationJobState  `json:"state"`
	Command         string             `json:"command,omitempty"`
	StartTime       *time.Time         `json:"startTime,omitempty"`
	EndTime         *time.Time         `json:"endTime,omitempty"`
	DurationSeconds float64            `json:"durationSeconds,omitempty"`
	Progress        MigrationProgress  `json:"progress"`
	Result          *MigrationResponse `json:"result,omitempty"`
	Error           string             `json:"error,omitempty"`
}

// migrationError is a custom error type that includes categorization
type migrationError struct {
	reason  string
	message string
}

func (e *migrationError) Error() string {
	return e.message
}

// MigrationJobManager manages the migration job lifecycle
type MigrationJobManager struct {
	mu         sync.RWMutex
	status     MigrationJobStatus
	cancelFunc context.CancelFunc
	ctx        context.Context
}

// Global job manager instance
var globalJobManager = &MigrationJobManager{
	status: MigrationJobStatus{State: StateIdle},
}

// GetStatus returns the current job status (thread-safe)
func (m *MigrationJobManager) GetStatus() MigrationJobStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// StartJob initiates a new migration job
func (m *MigrationJobManager) StartJob(command string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status.State == StateRunning {
		return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Migration is already running")
	}

	// Create new context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	m.ctx = ctx
	m.cancelFunc = cancel

	// Initialize job status
	startTime := time.Now()
	m.status = MigrationJobStatus{
		State:     StateRunning,
		Command:   command,
		StartTime: &startTime,
		Progress:  MigrationProgress{},
	}

	return nil
}

// CancelJob cancels the running migration job
func (m *MigrationJobManager) CancelJob() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status.State != StateRunning {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "No migration is currently running")
	}

	if m.cancelFunc != nil {
		m.cancelFunc()
		log.Info("Migration cancellation requested")
	}

	return nil
}

// UpdateProgress updates the migration progress (thread-safe)
func (m *MigrationJobManager) UpdateProgress(updateFunc func(*MigrationProgress)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	updateFunc(&m.status.Progress)
}

// CompleteJob marks the job as completed with results
func (m *MigrationJobManager) CompleteJob(result *MigrationResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	endTime := time.Now()
	m.status.State = StateCompleted
	m.status.EndTime = &endTime
	if m.status.StartTime != nil {
		m.status.DurationSeconds = endTime.Sub(*m.status.StartTime).Seconds()
	}
	m.status.Result = result

	// Clean up context
	if m.cancelFunc != nil {
		m.cancelFunc()
		m.cancelFunc = nil
	}
	m.ctx = nil
}

// FailJob marks the job as failed
func (m *MigrationJobManager) FailJob(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	endTime := time.Now()
	m.status.State = StateFailed
	m.status.EndTime = &endTime
	if m.status.StartTime != nil {
		m.status.DurationSeconds = endTime.Sub(*m.status.StartTime).Seconds()
	}
	m.status.Error = err.Error()

	// Clean up context
	if m.cancelFunc != nil {
		m.cancelFunc()
		m.cancelFunc = nil
	}
	m.ctx = nil
}

// CancelJobComplete marks the job as cancelled with partial results
func (m *MigrationJobManager) CancelJobComplete(result *MigrationResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	endTime := time.Now()
	m.status.State = StateCancelled
	m.status.EndTime = &endTime
	if m.status.StartTime != nil {
		m.status.DurationSeconds = endTime.Sub(*m.status.StartTime).Seconds()
	}
	m.status.Result = result

	// Clean up context
	if m.cancelFunc != nil {
		m.cancelFunc()
		m.cancelFunc = nil
	}
	m.ctx = nil
}

// GetContext returns the current job context
func (m *MigrationJobManager) GetContext() context.Context {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

func newMigrationError(reason, message string) *migrationError {
	return &migrationError{reason: reason, message: message}
}

// MigrateV1ToV2Handler handles the migration from V1 to V2 tag storage
func MigrateV1ToV2Handler(w http.ResponseWriter, r *http.Request) {
	// Extract and validate command parameter
	command, err := getCommandParameter(r)
	if err != nil {
		log.Warnf("Migration request missing or invalid command parameter: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	// Route to appropriate handler based on command
	switch command {
	case CommandDryRun:
		performDryRun(w, r)
	case CommandStart:
		performMigrationAsync(w, r)
	case CommandCancel:
		performCancelMigration(w, r)
	default:
		// This should not happen due to validation in getCommandParameter, but handle it defensively
		err := xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
			fmt.Sprintf("Invalid migration command '%s'. Allowed values: 'dryRun', 'start', 'cancel'", command))
		xhttp.WriteXconfErrorResponse(w, err)
	}
}

// MigrationStatusHandler returns the current migration job status
func MigrationStatusHandler(w http.ResponseWriter, r *http.Request) {
	status := globalJobManager.GetStatus()

	responseBytes, err := json.Marshal(status)
	if err != nil {
		log.Errorf("Failed to marshal migration status: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusOK, responseBytes)
}

// getCommandParameter extracts and validates the command parameter from the request
func getCommandParameter(r *http.Request) (string, error) {
	command := r.URL.Query().Get("command")

	if command == "" {
		return "", xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
			"Migration command is required. Use 'command=dryRun' to preview, 'command=start' to execute migration, or 'command=cancel' to cancel running migration")
	}

	if command != CommandDryRun && command != CommandStart && command != CommandCancel {
		return "", xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
			fmt.Sprintf("Invalid migration command '%s'. Allowed values: 'dryRun', 'start', 'cancel'", command))
	}

	return command, nil
}

// performCancelMigration cancels the running migration
func performCancelMigration(w http.ResponseWriter, r *http.Request) {
	log.Info("Cancellation requested for V1 to V2 tag migration")

	err := globalJobManager.CancelJob()
	if err != nil {
		log.Warnf("Failed to cancel migration: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	// Return current status
	status := globalJobManager.GetStatus()
	responseBytes, err := json.Marshal(status)
	if err != nil {
		log.Errorf("Failed to marshal status response: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusOK, responseBytes)
}

// performMigrationAsync starts the migration asynchronously
func performMigrationAsync(w http.ResponseWriter, r *http.Request) {
	log.Info("Starting async V1 to V2 tag migration")

	// Try to start the job
	err := globalJobManager.StartJob(CommandStart)
	if err != nil {
		// Job already running - return 409 Conflict
		log.Warnf("Failed to start migration: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	go runMigrationJob()

	// Return 202 Accepted with initial status
	status := globalJobManager.GetStatus()
	responseBytes, err := json.Marshal(status)
	if err != nil {
		log.Errorf("Failed to marshal status response: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusAccepted, responseBytes)
}

// performDryRun performs a dry run migration analysis
func performDryRun(w http.ResponseWriter, r *http.Request) {
	log.Info("Starting V1 to V2 tag migration dry run")

	tagIds, err := GetAllTagIds()
	if err != nil {
		log.Errorf("Failed to get tag IDs: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	log.Infof("Found %d tags to analyze", len(tagIds))

	response := DryRunResponse{
		Summary: DryRunSummary{
			TotalTags: len(tagIds),
		},
		Tags:     make(map[string]int),
		Warnings: []string{},
	}

	// Analyze each tag
	for _, tagId := range tagIds {
		prefixedTagId := SetTagPrefix(tagId)

		// Use safe type conversion
		tag, err := safeGetOneTag(prefixedTagId)
		if err != nil {
			var migErr *migrationError
			if errors.As(err, &migErr) {
				switch migErr.reason {
				case ReasonTagNotFound:
					log.Warnf("Tag '%s' not found in V1 table, will be skipped during migration", tagId)
					response.Warnings = append(response.Warnings,
						fmt.Sprintf("Tag '%s' not found in V1 table, will be skipped during migration", tagId))
				case ReasonTypeConversion:
					log.Errorf("Tag '%s' has type conversion error (expected *tagging.Tag, got incompatible type), will be skipped during migration", tagId)
					response.Warnings = append(response.Warnings,
						fmt.Sprintf("Tag '%s' has type conversion error, will be skipped during migration", tagId))
				}
			}
			continue
		}

		members := tag.Members.ToSlice()
		if len(members) == 0 {
			log.Warnf("Tag '%s' has 0 members, will be skipped during migration", tagId)
			response.Warnings = append(response.Warnings,
				fmt.Sprintf("Tag '%s' has no members, will be skipped during migration", tagId))
			continue
		}

		// Add to response
		response.Tags[tagId] = len(members)
		response.Summary.TotalMembersInXconf += len(members)
	}

	log.Infof("Dry run completed: %d tags can be migrated (%d members in Xconf), %d warnings",
		len(response.Tags), response.Summary.TotalMembersInXconf, len(response.Warnings))

	// Marshal and write response
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Failed to marshal dry run response: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusOK, responseBytes)
}

// runMigrationJob executes the migration in the background
func runMigrationJob() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Migration job panicked: %v", r)
			globalJobManager.FailJob(fmt.Errorf("migration panicked: %v", r))
		}
	}()

	startTime := time.Now()
	ctx := globalJobManager.GetContext()

	log.Info("Background migration job started")

	tagIds, err := GetAllTagIds()
	if err != nil {
		log.Errorf("Failed to get tag IDs: %v", err)
		globalJobManager.FailJob(err)
		return
	}

	log.Infof("Found %d tags to migrate", len(tagIds))

	// Initialize progress
	globalJobManager.UpdateProgress(func(p *MigrationProgress) {
		p.TotalTags = len(tagIds)
	})

	response := MigrationResponse{
		Summary: MigrationSummary{
			TotalTags: len(tagIds),
		},
		FailedTags:  []string{},
		SkippedTags: []TagMigrationSkip{},
		StartTime:   startTime.Format(time.RFC3339),
	}

	cancelled := false

	// Migrate each tag
	for _, tagId := range tagIds {
		select {
		case <-ctx.Done():
			processedTags := globalJobManager.GetStatus().Progress.ProcessedTags
			log.Warnf("Migration cancelled after processing %d/%d tags", processedTags, len(tagIds))
			cancelled = true
			break
		default:
		}

		if cancelled {
			break
		}

		// Update current tag in progress and reset member counters for new tag
		globalJobManager.UpdateProgress(func(p *MigrationProgress) {
			p.CurrentTag = tagId
			p.CurrentTagTotalMembers = 0
			p.CurrentTagMembersWritten = 0
			p.CurrentTagMembersMissingInXdas = 0
		})

		stats, err := migrateTagWithContext(ctx, tagId)

		// Update progress with stats if available
		if stats != nil {
			globalJobManager.UpdateProgress(func(p *MigrationProgress) {
				p.CurrentTagTotalMembers = stats.TotalMembers
				p.CurrentTagMembersWritten = stats.MembersWritten
				p.CurrentTagMembersMissingInXdas = stats.MembersMissingInXdas
			})

			// Aggregate to overall summary
			response.Summary.TotalMembers += stats.TotalMembers
			response.Summary.TotalMembersWritten += stats.MembersWritten
			response.Summary.TotalMembersMissingInXdas += stats.MembersMissingInXdas
		}

		if err != nil {
			// Check if it's a skip reason or actual failure
			var migErr *migrationError
			if errors.As(err, &migErr) {
				switch migErr.reason {
				case ReasonTagNotFound, ReasonNoMembers, ReasonTypeConversion:
					response.SkippedTags = append(response.SkippedTags, TagMigrationSkip{
						TagId:  tagId,
						Reason: migErr.reason,
					})
					response.Summary.SkippedCount++
					globalJobManager.UpdateProgress(func(p *MigrationProgress) {
						p.SkippedCount++
					})
				case ReasonXdasVerification, ReasonDatabaseWriteFailed:
					response.FailedTags = append(response.FailedTags, tagId)
					response.Summary.FailureCount++
					globalJobManager.UpdateProgress(func(p *MigrationProgress) {
						p.FailureCount++
					})
					log.Errorf("Failed to migrate tag '%s': %v", tagId, err)
				default:
					response.FailedTags = append(response.FailedTags, tagId)
					response.Summary.FailureCount++
					globalJobManager.UpdateProgress(func(p *MigrationProgress) {
						p.FailureCount++
					})
					log.Errorf("Failed to migrate tag '%s': %v", tagId, err)
				}
			} else {
				// Non-migration error, treat as failure
				response.FailedTags = append(response.FailedTags, tagId)
				response.Summary.FailureCount++
				globalJobManager.UpdateProgress(func(p *MigrationProgress) {
					p.FailureCount++
				})
				log.Errorf("Failed to migrate tag '%s': %v", tagId, err)
			}
		} else {
			response.Summary.SuccessCount++
			globalJobManager.UpdateProgress(func(p *MigrationProgress) {
				p.SuccessCount++
			})
		}

		// Update processed count
		globalJobManager.UpdateProgress(func(p *MigrationProgress) {
			p.ProcessedTags++
		})
	}

	endTime := time.Now()
	response.EndTime = endTime.Format(time.RFC3339)
	response.DurationSeconds = endTime.Sub(startTime).Seconds()

	// Determine overall status
	if cancelled {
		response.Status = "cancelled"
		log.Infof("V1 to V2 migration cancelled in %.1fs: %d successful, %d failed, %d skipped",
			response.DurationSeconds, response.Summary.SuccessCount, response.Summary.FailureCount, response.Summary.SkippedCount)
		globalJobManager.CancelJobComplete(&response)
	} else {
		if response.Summary.FailureCount == 0 && response.Summary.SuccessCount > 0 {
			response.Status = "completed"
		} else if response.Summary.SuccessCount > 0 {
			response.Status = "completed_with_errors"
		} else {
			response.Status = "failed"
		}

		log.Infof("V1 to V2 migration completed in %.1fs: %d successful, %d failed, %d skipped",
			response.DurationSeconds, response.Summary.SuccessCount, response.Summary.FailureCount, response.Summary.SkippedCount)
		globalJobManager.CompleteJob(&response)
	}
}

// safeGetOneTag safely retrieves a tag with type conversion protection
func safeGetOneTag(tagId string) (*xwtagging.Tag, error) {
	inst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TAG, tagId)
	if err != nil {
		log.Debugf("Tag '%s' not found in V1 table: %v", tagId, err)
		return nil, newMigrationError(ReasonTagNotFound, fmt.Sprintf("tag not found: %v", err))
	}

	// Attempt type assertion with safety check
	tag, ok := inst.(*xwtagging.Tag)
	if !ok {
		actualType := reflect.TypeOf(inst)
		log.Errorf("Type conversion error for tag '%s': expected *tagging.Tag, got %v", tagId, actualType)
		return nil, newMigrationError(ReasonTypeConversion,
			fmt.Sprintf("type conversion failed: expected *tagging.Tag, got %v", actualType))
	}

	// Clone the tag to avoid mutations
	clone, err := tag.Clone()
	if err != nil {
		log.Errorf("Failed to clone tag '%s': %v", tagId, err)
		return nil, newMigrationError(ReasonTypeConversion, fmt.Sprintf("failed to clone tag: %v", err))
	}

	return clone, nil
}

func migrateTagWithContext(ctx context.Context, tagId string) (*TagMemberStats, error) {
	prefixedTagId := SetTagPrefix(tagId)

	// Use safe type conversion to prevent panics
	tag, err := safeGetOneTag(prefixedTagId)
	if err != nil {
		// Check a specific reason of an error
		var migErr *migrationError
		if errors.As(err, &migErr) {
			switch migErr.reason {
			case ReasonTagNotFound:
				log.Warnf("[Tag: %s] Tag not found in V1 table, skipping", tagId)
				return nil, newMigrationError(ReasonTagNotFound, "tag not found")
			case ReasonTypeConversion:
				log.Errorf("[Tag: %s] Type conversion error (expected *tagging.Tag, got incompatible type), skipping", tagId)
				return nil, newMigrationError(ReasonTypeConversion, "type conversion failed")
			}
		}
		// Unexpected error
		log.Errorf("[Tag: %s] Failed to retrieve tag: %v", tagId, err)
		return nil, err
	}

	members := tag.Members.ToSlice()
	if len(members) == 0 {
		log.Warnf("[Tag: %s] has no members, skipping", tagId)
		return nil, newMigrationError(ReasonNoMembers, "no members")
	}

	log.Infof("[Tag: %s] Migrating %d members", tagId, len(members))

	stats := &TagMemberStats{
		TotalMembers: len(members),
	}

	// Update progress with total members for this tag
	globalJobManager.UpdateProgress(func(p *MigrationProgress) {
		p.CurrentTagTotalMembers = len(members)
	})

	verifiedMembers, missingInXdas, err := verifyMembersInXdasWithContext(ctx, prefixedTagId, members)
	if err != nil {
		log.Errorf("[Tag: %s] XDAS verification failed: %v", tagId, err)
		return nil, newMigrationError(ReasonXdasVerification, fmt.Sprintf("XDAS verification failed: %v", err))
	}

	stats.MembersMissingInXdas = missingInXdas

	// Update progress with missing members count
	globalJobManager.UpdateProgress(func(p *MigrationProgress) {
		p.CurrentTagMembersMissingInXdas = missingInXdas
	})

	if len(verifiedMembers) == 0 {
		log.Warnf("[Tag: %s] No members verified in XDAS, skipping", tagId)
		return stats, newMigrationError(ReasonXdasVerification, "no members verified in XDAS")
	}

	if len(verifiedMembers) < len(members) {
		log.Warnf("[Tag: %s] Only %d/%d members verified in XDAS",
			tagId, len(verifiedMembers), len(members))
	}

	if err := writeMembersToV2InBatchesWithContext(ctx, tagId, verifiedMembers); err != nil {
		log.Errorf("[Tag: %s] Failed to write to V2: %v", tagId, err)
		return stats, newMigrationError(ReasonDatabaseWriteFailed, fmt.Sprintf("failed to write to V2: %v", err))
	}

	stats.MembersWritten = len(verifiedMembers)

	log.Infof("[Tag: %s] Successfully migrated: verified %d/%d members from XDAS, added to V2",
		tagId, len(verifiedMembers), len(members))

	return stats, nil
}

func writeMembersToV2InBatchesWithContext(ctx context.Context, tagId string, members []string) error {
	totalMembers := len(members)
	successCount := 0
	var allErrors []string

	for i := 0; i < totalMembers; i += MaxBatchSizeV2 {
		// Check for cancellation
		select {
		case <-ctx.Done():
			log.Warnf("[Tag: %s] Batch write cancelled at %d/%d members", tagId, i, totalMembers)
			return fmt.Errorf("batch write cancelled: %w", ctx.Err())
		default:
		}

		end := i + MaxBatchSizeV2
		if end > totalMembers {
			end = totalMembers
		}

		batch := members[i:end]
		log.Debugf("[Tag: %s] Writing batch %d-%d of %d members", tagId, i, end, totalMembers)

		if err := AddMembersV2(tagId, batch); err != nil {
			// Log the batch failure but continue with remaining batches
			log.Errorf("[Tag: %s] Failed to write batch %d-%d: %v", tagId, i, end, err)
			allErrors = append(allErrors, fmt.Sprintf("batch %d-%d: %v", i, end, err))
			// Note: AddMembersV2 already logs individual member failures at the bucket level
		} else {
			successCount += len(batch)
			// Update progress for current tag after each successful batch
			globalJobManager.UpdateProgress(func(p *MigrationProgress) {
				p.CurrentTagMembersWritten += len(batch)
			})
		}
	}

	if successCount == 0 {
		return fmt.Errorf("failed to write any members: %s", strings.Join(allErrors, "; "))
	}

	if len(allErrors) > 0 {
		log.Warnf("[Tag: %s] Partial migration: %d/%d members written, some batches failed: %s",
			tagId, successCount, totalMembers, strings.Join(allErrors, "; "))
		return fmt.Errorf("partial migration: %d/%d members written", successCount, totalMembers)
	}

	log.Debugf("[Tag: %s] Successfully wrote all %d members to V2", tagId, successCount)
	return nil
}

func verifyMembersInXdasWithContext(ctx context.Context, tagId string, members []string) ([]string, int, error) {
	membersChannel := make(chan string, len(members))
	go func() {
		defer close(membersChannel)
		for _, member := range members {
			select {
			case <-ctx.Done():
				return
			case membersChannel <- member:
			}
		}
	}()

	wg := &sync.WaitGroup{}
	verifiedMembersChannel := make(chan string, len(members))

	config := GetTagApiConfig()
	numOfWorkers := 1
	if config != nil {
		baseWorkers := config.WorkerCount
		scaledWorkers := min(max(len(members)/100, baseWorkers), MaxWorkersV2)
		numOfWorkers = scaledWorkers
	}

	log.Debugf("Using %d workers for XDAS verification of %d members", numOfWorkers, len(members))

	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go verifyMemberInXdasWorkerWithContext(ctx, tagId, membersChannel, verifiedMembersChannel, wg)
	}

	go func() {
		wg.Wait()
		close(verifiedMembersChannel)
	}()

	var verifiedMembers []string
	for member := range verifiedMembersChannel {
		verifiedMembers = append(verifiedMembers, member)
	}

	missingInXdas := len(members) - len(verifiedMembers)
	if missingInXdas > 0 {
		log.Warnf("XDAS verification: %d/%d members verified, %d missing in XDAS",
			len(verifiedMembers), len(members), missingInXdas)
	}

	return verifiedMembers, missingInXdas, nil
}

func verifyMemberInXdasWorkerWithContext(ctx context.Context, tagId string, members <-chan string, verifiedMembers chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for member := range members {
		// Check for cancellation
		select {
		case <-ctx.Done():
			log.Debugf("Worker cancelled during XDAS verification for tag '%s'", tagId)
			return
		default:
		}

		normalizedMember := ToNormalizedEcm(member)

		tagsResponse, err := GetGroupServiceConnector().GetGroupsMemberBelongsTo(normalizedMember)
		if err != nil {
			log.Errorf("XDAS error verifying member '%s' for tag '%s': %v", normalizedMember, tagId, err)
			continue
		}

		if tagsResponse != nil && tagsResponse.Fields != nil {
			tagsMap := util.StringMap(tagsResponse.GetFields())
			tagKeys := tagsMap.Keys()

			found := false
			for _, returnedTag := range tagKeys {
				if strings.EqualFold(returnedTag, tagId) {
					found = true
					break
				}
			}

			if found {
				verifiedMembers <- member
			} else {
				log.Warnf("Member '%s' does not belong to tag '%s' in XDAS, skipping", normalizedMember, tagId)
			}
		} else {
			log.Warnf("Member '%s' has no tags in XDAS, skipping for tag '%s'", normalizedMember, tagId)
		}
	}
}
