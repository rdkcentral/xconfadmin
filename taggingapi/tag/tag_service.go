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

func GetTagsByMember(member string, tagType string) ([]string, error) {
	member, err := NormalizeMember(member, tagType)
	if err != nil {
		return []string{}, err
	}
	tagsAsHashes, err := GetGroupServiceConnector().GetGroupsMemberBelongsToOfType(member, tagType)
	if err != nil {
		log.Errorf("xdas error getting members by %s group: %s", member, err.Error())
		return []string{}, err
	}
	tagsMap := util.StringMap(tagsAsHashes.GetFields())
	return filterTagEntriesByPrefix(tagsMap.Keys()), nil
}

func GetTagsWithValuesByMember(member string, tagType string) (map[string]string, error) {
	member, err := NormalizeMember(member, tagType)
	if err != nil {
		return map[string]string{}, err
	}
	tagsAsHashes, err := GetGroupServiceConnector().GetGroupsMemberBelongsToOfType(member, tagType)
	if err != nil {
		log.Errorf("xdas error getting members by %s group: %s", member, err.Error())
		return map[string]string{}, err
	}
	tagsMap := util.StringMap(tagsAsHashes.GetFields())
	return filterTagEntriesWithValuesByPrefix(tagsMap), nil
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

// Per-member errors go into the aggregator (count + first message) instead of
// being logged one line per member — an XDAS outage during a 5000-member
// batch must not emit 5000 error lines.
func storeTagMembersInXdas(id string, members <-chan string, savedMembers chan<- string, wg *sync.WaitGroup, tagValue string, tagType string, agg *errorAggregator) {
	defer wg.Done()
	// Constructed per worker, not shared across them: proto.Marshal writes to
	// the message's internal state, so hoisting this to a single instance shared
	// by all workers would be a data race.
	xdasMembers := proto.XdasHashes{
		Fields: map[string]string{id: tagValue},
	}

	successCount := 0
	failCount := 0

	for member := range members {
		normalized, err := NormalizeMember(member, tagType)
		if err != nil {
			failCount++
			agg.add(err)
			continue
		}
		err = GetGroupServiceSyncConnector().AddMembersToTagOfType(normalized, &xdasMembers, tagType)
		if err != nil {
			failCount++
			agg.add(err)
		} else {
			successCount++
			savedMembers <- storedMemberForm(member, normalized, tagType)
		}
	}

	log.Debugf("XDAS worker completed for tag %s: success=%d, failed=%d", id, successCount, failCount)
}

func removeTagMembersFromXdas(id string, members <-chan string, removedMembers chan<- string, wg *sync.WaitGroup, tagType string, agg *errorAggregator) {
	defer wg.Done()

	successCount := 0
	failCount := 0

	for member := range members {
		normalized, err := NormalizeMember(member, tagType)
		if err != nil {
			failCount++
			agg.add(err)
			continue
		}
		err = GetGroupServiceSyncConnector().RemoveGroupMembersOfType(normalized, id, tagType)
		if err != nil {
			failCount++
			agg.add(err)
		} else {
			successCount++
			removedMembers <- storedMemberForm(member, normalized, tagType)
		}
	}

	log.Debugf("XDAS remove worker completed for tag %s: success=%d, failed=%d", id, successCount, failCount)
}

// storedMemberForm picks which form of a member is persisted to Cassandra.
//
// Account tags store the normalized id, so that Cassandra and XDAS agree and a
// later delete computes the same FNV bucket. Without this, a member submitted
// with surrounding whitespace would be stored padded, land in a different
// bucket than the clean id, and become permanently undeletable.
//
// Mac and legacy tags keep storing the raw member. This is not an oversight:
// GetEcmMacAddress subtracts 2 and is not idempotent, so storing the normalized
// form would shift every existing member into a different bucket and break every
// subsequent delete against the 43M rows already written.
func storedMemberForm(raw string, normalized string, tagType string) string {
	if IsAccountTag(tagType) {
		return normalized
	}
	return raw
}

func CheckBatchSizeExceeded(batchSize int) error {
	config := GetTagApiConfig()
	if batchSize > config.BatchLimit {
		return fmt.Errorf(MaxMemberLimitExceededErrorMsg, batchSize, config.BatchLimit)
	}
	return nil
}
