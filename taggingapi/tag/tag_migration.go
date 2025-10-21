package tag

import (
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
)

// Skip reason constants
const (
	ReasonTagNotFound         = "tag_not_found_in_v1"
	ReasonNoMembers           = "no_members"
	ReasonTypeConversion      = "type_conversion_error"
	ReasonXdasVerification    = "xdas_verification_failed"
	ReasonDatabaseWriteFailed = "database_write_failed"
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
	TotalTags    int `json:"totalTags"`
	SuccessCount int `json:"successCount"`
	FailureCount int `json:"failureCount"`
	SkippedCount int `json:"skippedCount"`
}

// TagMigrationSkip represents a skipped tag with reason
type TagMigrationSkip struct {
	TagId  string `json:"tagId"`
	Reason string `json:"reason"`
}

// migrationError is a custom error type that includes categorization
type migrationError struct {
	reason  string
	message string
}

func (e *migrationError) Error() string {
	return e.message
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
		performMigration(w, r)
	default:
		// This should not happen due to validation in getCommandParameter, but handle it defensively
		err := xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
			fmt.Sprintf("Invalid migration command '%s'. Allowed values: 'dryRun', 'start'", command))
		xhttp.WriteXconfErrorResponse(w, err)
	}
}

// getCommandParameter extracts and validates the command parameter from the request
func getCommandParameter(r *http.Request) (string, error) {
	command := r.URL.Query().Get("command")

	if command == "" {
		return "", xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
			"Migration command is required. Use 'command=dryRun' to preview or 'command=start' to execute migration")
	}

	if command != CommandDryRun && command != CommandStart {
		return "", xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
			fmt.Sprintf("Invalid migration command '%s'. Allowed values: 'dryRun', 'start'", command))
	}

	return command, nil
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

// performMigration performs the actual V1 to V2 migration
func performMigration(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Info("Starting V1 to V2 tag migration")

	tagIds, err := GetAllTagIds()
	if err != nil {
		log.Errorf("Failed to get tag IDs: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	log.Infof("Found %d tags to migrate", len(tagIds))

	response := MigrationResponse{
		Summary: MigrationSummary{
			TotalTags: len(tagIds),
		},
		FailedTags:  []string{},
		SkippedTags: []TagMigrationSkip{},
		StartTime:   startTime.Format(time.RFC3339),
	}

	// Migrate each tag
	for _, tagId := range tagIds {
		err := migrateTag(tagId)
		if err != nil {
			// Check if it's a skip reason or actual failure
			var migErr *migrationError
			if errors.As(err, &migErr) {
				switch migErr.reason {
				case ReasonTagNotFound, ReasonNoMembers, ReasonTypeConversion:
					// These are skip reasons
					response.SkippedTags = append(response.SkippedTags, TagMigrationSkip{
						TagId:  tagId,
						Reason: migErr.reason,
					})
					response.Summary.SkippedCount++
				case ReasonXdasVerification, ReasonDatabaseWriteFailed:
					// These are failure reasons
					response.FailedTags = append(response.FailedTags, tagId)
					response.Summary.FailureCount++
					log.Errorf("Failed to migrate tag '%s': %v", tagId, err)
				default:
					// Unknown reason, treat as failure
					response.FailedTags = append(response.FailedTags, tagId)
					response.Summary.FailureCount++
					log.Errorf("Failed to migrate tag '%s': %v", tagId, err)
				}
			} else {
				// Non-migration error, treat as failure
				response.FailedTags = append(response.FailedTags, tagId)
				response.Summary.FailureCount++
				log.Errorf("Failed to migrate tag '%s': %v", tagId, err)
			}
		} else {
			response.Summary.SuccessCount++
		}
	}

	endTime := time.Now()
	response.EndTime = endTime.Format(time.RFC3339)
	response.DurationSeconds = endTime.Sub(startTime).Seconds()

	// Determine overall status
	if response.Summary.FailureCount == 0 && response.Summary.SuccessCount > 0 {
		response.Status = "completed"
	} else if response.Summary.SuccessCount > 0 {
		response.Status = "completed_with_errors"
	} else {
		response.Status = "failed"
	}

	log.Infof("V1 to V2 migration completed in %.1fs: %d successful, %d failed, %d skipped",
		response.DurationSeconds, response.Summary.SuccessCount, response.Summary.FailureCount, response.Summary.SkippedCount)

	// Marshal and write response
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Failed to marshal migration response: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	// Use appropriate HTTP status code
	statusCode := http.StatusOK
	if response.Summary.FailureCount > 0 && response.Summary.SuccessCount > 0 {
		statusCode = http.StatusMultiStatus // 207
	} else if response.Summary.SuccessCount == 0 {
		statusCode = http.StatusInternalServerError // 500
	}

	xhttp.WriteXconfResponse(w, statusCode, responseBytes)
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

func migrateTag(tagId string) error {
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
				return newMigrationError(ReasonTagNotFound, "tag not found")
			case ReasonTypeConversion:
				log.Errorf("[Tag: %s] Type conversion error (expected *tagging.Tag, got incompatible type), skipping", tagId)
				return newMigrationError(ReasonTypeConversion, "type conversion failed")
			}
		}
		// Unexpected error
		log.Errorf("[Tag: %s] Failed to retrieve tag: %v", tagId, err)
		return err
	}

	members := tag.Members.ToSlice()
	if len(members) == 0 {
		log.Warnf("[Tag: %s] has no members, skipping", tagId)
		return newMigrationError(ReasonNoMembers, "no members")
	}

	log.Infof("[Tag: %s] Migrating %d members", tagId, len(members))

	verifiedMembers, err := verifyMembersInXdas(prefixedTagId, members)
	if err != nil {
		log.Errorf("[Tag: %s] XDAS verification failed: %v", tagId, err)
		return newMigrationError(ReasonXdasVerification, fmt.Sprintf("XDAS verification failed: %v", err))
	}

	if len(verifiedMembers) == 0 {
		log.Warnf("[Tag: %s] No members verified in XDAS, skipping", tagId)
		return newMigrationError(ReasonXdasVerification, "no members verified in XDAS")
	}

	if len(verifiedMembers) < len(members) {
		log.Warnf("[Tag: %s] Only %d/%d members verified in XDAS",
			tagId, len(verifiedMembers), len(members))
	}

	if err := writeMembersToV2InBatches(tagId, verifiedMembers); err != nil {
		log.Errorf("[Tag: %s] Failed to write to V2: %v", tagId, err)
		return newMigrationError(ReasonDatabaseWriteFailed, fmt.Sprintf("failed to write to V2: %v", err))
	}

	log.Infof("[Tag: %s] Successfully migrated: verified %d/%d members from XDAS, added to V2",
		tagId, len(verifiedMembers), len(members))

	return nil
}

func writeMembersToV2InBatches(tagId string, members []string) error {
	totalMembers := len(members)
	successCount := 0
	var allErrors []string

	for i := 0; i < totalMembers; i += MaxBatchSizeV2 {
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

func verifyMembersInXdas(tagId string, members []string) ([]string, error) {
	membersChannel := make(chan string, len(members))
	go func() {
		defer close(membersChannel)
		for _, member := range members {
			membersChannel <- member
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
		go verifyMemberInXdasWorker(tagId, membersChannel, verifiedMembersChannel, wg)
	}

	go func() {
		wg.Wait()
		close(verifiedMembersChannel)
	}()

	var verifiedMembers []string
	for member := range verifiedMembersChannel {
		verifiedMembers = append(verifiedMembers, member)
	}

	if len(verifiedMembers) != len(members) {
		log.Warnf("XDAS verification: %d/%d members verified", len(verifiedMembers), len(members))
	}

	return verifiedMembers, nil
}

func verifyMemberInXdasWorker(tagId string, members <-chan string, verifiedMembers chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for member := range members {
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
