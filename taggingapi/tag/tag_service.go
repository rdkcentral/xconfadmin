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

// GetTagsByMember returns the member's tags filtered to the tenant, plus the
// number of tags XDAS reported before filtering (for the request log).
func GetTagsByMember(tenantId string, member string) ([]string, int, error) {
	member = ToNormalizedEcm(member)
	tagsAsHashes, err := GetGroupServiceConnector().GetGroupsMemberBelongsTo(member)
	if err != nil {
		log.Errorf("xdas error getting members by %s group: %s", member, err.Error())
		return []string{}, 0, err
	}
	tagsMap := util.StringMap(tagsAsHashes.GetFields())
	xdasTags := filterTagEntriesByPrefix(tagsMap.Keys())

	filtered, err := filterByTenant(tenantId, xdasTags)
	return filtered, len(xdasTags), err
}

// GetTagsWithValuesByMember returns the member's tags with values filtered to
// the tenant, plus the number of tags XDAS reported before filtering.
func GetTagsWithValuesByMember(tenantId string, member string) (map[string]string, int, error) {
	member = ToNormalizedEcm(member)
	tagsAsHashes, err := GetGroupServiceConnector().GetGroupsMemberBelongsTo(member)
	if err != nil {
		log.Errorf("xdas error getting members by %s group: %s", member, err.Error())
		return map[string]string{}, 0, err
	}
	tagsMap := util.StringMap(tagsAsHashes.GetFields())
	xdasTags := filterTagEntriesWithValuesByPrefix(tagsMap)

	filtered, err := filterByTenantWithValues(tenantId, xdasTags)
	return filtered, len(xdasTags), err
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

// Per-member errors go into the aggregator (count + first message) instead of
// being logged one line per member — an XDAS outage during a 5000-member
// batch must not emit 5000 error lines.
func storeTagMembersInXdas(id string, members <-chan string, savedMembers chan<- string, wg *sync.WaitGroup, tagValue string, agg *errorAggregator) {
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
			agg.add(err)
		} else {
			successCount++
			savedMembers <- member
		}
	}

	log.Debugf("XDAS worker completed for tag %s: success=%d, failed=%d", id, successCount, failCount)
}

func removeTagMembersFromXdas(id string, members <-chan string, removedMembers chan<- string, wg *sync.WaitGroup, agg *errorAggregator) {
	defer wg.Done()

	successCount := 0
	failCount := 0

	for member := range members {
		normalizedEcm := ToNormalizedEcm(member)
		err := GetGroupServiceSyncConnector().RemoveGroupMembers(normalizedEcm, id)
		if err != nil {
			failCount++
			agg.add(err)
		} else {
			successCount++
			removedMembers <- member
		}
	}

	log.Debugf("XDAS remove worker completed for tag %s: success=%d, failed=%d", id, successCount, failCount)
}

func CheckBatchSizeExceeded(batchSize int) error {
	config := GetTagApiConfig()
	if batchSize > config.BatchLimit {
		return fmt.Errorf(MaxMemberLimitExceededErrorMsg, batchSize, config.BatchLimit)
	}
	return nil
}
