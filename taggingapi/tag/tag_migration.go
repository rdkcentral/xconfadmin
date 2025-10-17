package tag

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/util"

	log "github.com/sirupsen/logrus"
)

// MigrateV1ToV2Handler handles the migration from V1 to V2 tag storage
func MigrateV1ToV2Handler(w http.ResponseWriter, r *http.Request) {
	log.Info("Starting V1 to V2 tag migration")

	tagIds, err := GetAllTagIds()
	if err != nil {
		log.Errorf("Failed to get tag IDs: %v", err)
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	log.Infof("Starting V1 to V2 tag migration, found %d tags", len(tagIds))

	successCount := 0
	failureCount := 0

	for _, tagId := range tagIds {
		if err := migrateTag(tagId); err != nil {
			log.Errorf("Failed to migrate tag '%s': %v", tagId, err)
			failureCount++
		} else {
			successCount++
		}
	}

	log.Infof("V1 to V2 migration completed, processed %d tags (success: %d, failed: %d)",
		len(tagIds), successCount, failureCount)

	xhttp.WriteXconfResponse(w, http.StatusOK, []byte("Migration completed"))
}

func migrateTag(tagId string) error {
	prefixedTagId := SetTagPrefix(tagId)
	tag := GetOneTag(prefixedTagId)
	if tag == nil {
		log.Warnf("Tag '%s' not found in V1 table, skipping", tagId)
		return nil
	}

	members := tag.Members.ToSlice()
	if len(members) == 0 {
		log.Infof("Tag '%s' has no members, skipping", tagId)
		return nil
	}

	log.Infof("Migrating tag '%s' with %d members", tagId, len(members))

	verifiedMembers, err := verifyMembersInXdas(prefixedTagId, members)
	if err != nil {
		return fmt.Errorf("XDAS verification failed: %w", err)
	}

	if len(verifiedMembers) == 0 {
		log.Warnf("Tag '%s': no members verified in XDAS, skipping", tagId)
		return nil
	}

	if len(verifiedMembers) < len(members) {
		log.Warnf("Tag '%s': only %d/%d members verified in XDAS",
			tagId, len(verifiedMembers), len(members))
	}

	if err := writeMembersToV2InBatches(tagId, verifiedMembers); err != nil {
		return fmt.Errorf("failed to write to V2: %w", err)
	}

	log.Infof("Successfully migrated tag '%s': verified %d/%d members from XDAS, added to V2",
		tagId, len(verifiedMembers), len(members))

	return nil
}

func writeMembersToV2InBatches(tagId string, members []string) error {
	totalMembers := len(members)

	for i := 0; i < totalMembers; i += MaxBatchSizeV2 {
		end := i + MaxBatchSizeV2
		if end > totalMembers {
			end = totalMembers
		}

		batch := members[i:end]
		log.Debugf("Writing batch %d-%d of %d members for tag '%s'", i, end, totalMembers, tagId)

		if err := AddMembersV2(tagId, batch); err != nil {
			return fmt.Errorf("failed to write batch %d-%d: %w", i, end, err)
		}
	}

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
