package tag

import (
	"fmt"
	"strings"
	"sync"

	"github.com/rdkcentral/xconfadmin/http"
	taggingapi_config "github.com/rdkcentral/xconfadmin/taggingapi/config"
	proto "github.com/rdkcentral/xconfadmin/taggingapi/proto/generated"
	"github.com/rdkcentral/xconfadmin/util"
	log "github.com/sirupsen/logrus"
)

func GetGroupServiceSyncConnector() *http.GroupServiceSyncConnector {
	return http.WebConfServer.GroupServiceSyncConnector
}

func GetTagApiConfig() *taggingapi_config.TaggingApiConfig {
	return http.WebConfServer.TaggingApiConfig
}

func SetTagApiConfig(config *taggingapi_config.TaggingApiConfig) {
	http.WebConfServer.TaggingApiConfig = config
}

func GetGroupServiceConnector() *http.GroupServiceConnector {
	return http.WebConfServer.GroupServiceConnector
}

func GetTagsByMember(tenantId string, member string) ([]string, error) {
	member = ToNormalizedEcm(member)
	tagsAsHashes, err := GetGroupServiceConnector().GetGroupsMemberBelongsTo(member)
	if err != nil {
		log.Errorf("xdas error getting members by %s group: %s", member, err.Error())
		return []string{}, err
	}
	tagsMap := util.StringMap(tagsAsHashes.GetFields())
	xdasTags := filterTagEntriesByPrefix(tagsMap.Keys())

	return filterByTenant(tenantId, xdasTags)
}

func GetTagsWithValuesByMember(tenantId string, member string) (map[string]string, error) {
	member = ToNormalizedEcm(member)
	tagsAsHashes, err := GetGroupServiceConnector().GetGroupsMemberBelongsTo(member)
	if err != nil {
		log.Errorf("xdas error getting members by %s group: %s", member, err.Error())
		return map[string]string{}, err
	}
	tagsMap := util.StringMap(tagsAsHashes.GetFields())
	xdasTags := filterTagEntriesWithValuesByPrefix(tagsMap)

	return filterByTenantWithValues(tenantId, xdasTags)
}

func filterTagEntriesByPrefix(ftEntries []string) []string {
	tags := []string{}
	for _, ftEntry := range ftEntries {
		if strings.HasPrefix(ftEntry, Prefix) {
			tags = append(tags, RemovePrefixFromTag(ftEntry))
		}
	}
	return tags
}

func filterTagEntriesWithValuesByPrefix(entries util.StringMap) map[string]string {
	result := map[string]string{}
	for key, value := range entries {
		if strings.HasPrefix(key, Prefix) {
			result[RemovePrefixFromTag(key)] = value
		}
	}
	return result
}

// filterByTenant intersects XDAS tags with tenant-owned tags from Cassandra
func filterByTenant(tenantId string, xdasTags []string) ([]string, error) {
	tenantTags, err := GetAllTagIds(tenantId)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant tags for filtering: %w", err)
	}

	tenantTagSet := make(map[string]bool, len(tenantTags))
	for _, t := range tenantTags {
		tenantTagSet[t] = true
	}

	filtered := make([]string, 0, len(xdasTags))
	for _, tag := range xdasTags {
		if tenantTagSet[tag] {
			filtered = append(filtered, tag)
		}
	}
	return filtered, nil
}

// filterByTenantWithValues intersects XDAS tags (with values) with tenant-owned tags from Cassandra
func filterByTenantWithValues(tenantId string, xdasTags map[string]string) (map[string]string, error) {
	tenantTags, err := GetAllTagIds(tenantId)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant tags for filtering: %w", err)
	}

	tenantTagSet := make(map[string]bool, len(tenantTags))
	for _, t := range tenantTags {
		tenantTagSet[t] = true
	}

	filtered := make(map[string]string)
	for tag, value := range xdasTags {
		if tenantTagSet[tag] {
			filtered[tag] = value
		}
	}
	return filtered, nil
}

func storeTagMembersInXdas(id string, members <-chan string, savedMembers chan<- string, wg *sync.WaitGroup, tagValue string) {
	defer wg.Done()
	xdasMembers := proto.XdasHashes{
		Fields: map[string]string{id: tagValue},
	}

	successCount := 0
	failCount := 0

	for member := range members {
		normalizedEcm := ToNormalizedEcm(member)
		err := GetGroupServiceSyncConnector().AddMembersToTag(normalizedEcm, &xdasMembers)
		if err != nil {
			failCount++
			log.Errorf("xdas error adding member to %s group: ecm=%s, error=%s", id, normalizedEcm, err.Error())
		} else {
			successCount++
			savedMembers <- member
		}
	}

	// Worker summary log (one line per worker)
	if failCount > 0 {
		log.Warnf("XDAS worker completed for tag %s: success=%d, failed=%d", id, successCount, failCount)
	} else {
		log.Debugf("XDAS worker completed for tag %s: success=%d", id, successCount)
	}
}

func removeTagMembersFromXdas(id string, members <-chan string, removedMembers chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	successCount := 0
	failCount := 0

	for member := range members {
		normalizedEcm := ToNormalizedEcm(member)
		err := GetGroupServiceSyncConnector().RemoveGroupMembers(normalizedEcm, id)
		if err != nil {
			failCount++
			log.Errorf("xdas error removing member from %s group: ecm=%s, error=%s", id, normalizedEcm, err.Error())
		} else {
			successCount++
			removedMembers <- member
		}
	}

	// Worker summary log (one line per worker)
	if failCount > 0 {
		log.Warnf("XDAS remove worker completed for tag %s: success=%d, failed=%d", id, successCount, failCount)
	} else {
		log.Debugf("XDAS remove worker completed for tag %s: success=%d", id, successCount)
	}
}

func CheckBatchSizeExceeded(batchSize int) error {
	config := GetTagApiConfig()
	if batchSize > config.BatchLimit {
		return fmt.Errorf(MaxMemberLimitExceededErrorMsg, batchSize, config.BatchLimit)
	}
	return nil
}
