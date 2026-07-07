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

func GetTagsByMember(member string) ([]string, error) {
	member = ToNormalizedEcm(member)
	tagsAsHashes, err := GetGroupServiceConnector().GetGroupsMemberBelongsTo(member)
	if err != nil {
		log.Errorf("xdas error getting members by %s group: %s", member, err.Error())
		return []string{}, err
	}
	tagsMap := util.StringMap(tagsAsHashes.GetFields())
	return filterTagEntriesByPrefix(tagsMap.Keys()), err
}

func GetTagsWithValuesByMember(member string) (map[string]string, error) {
	member = ToNormalizedEcm(member)
	tagsAsHashes, err := GetGroupServiceConnector().GetGroupsMemberBelongsTo(member)
	if err != nil {
		log.Errorf("xdas error getting members by %s group: %s", member, err.Error())
		return map[string]string{}, err
	}
	tagsMap := util.StringMap(tagsAsHashes.GetFields())
	return filterTagEntriesWithValuesByPrefix(tagsMap), err
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
