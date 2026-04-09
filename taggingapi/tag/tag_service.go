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
	keyspace := keyspaceForTagType(tagType)
	normalized := NormalizeMember(member, tagType)
	tagsAsHashes, err := GetGroupServiceConnector().GetGroupsMemberBelongsToWithKeyspace(keyspace, normalized)
	if err != nil {
		log.Errorf("xdas error getting members by %s group: %s", normalized, err.Error())
		return []string{}, err
	}
	tagsMap := util.StringMap(tagsAsHashes.GetFields())
	return filterTagEntriesByPrefix(tagsMap.Keys()), err
}

func keyspaceForTagType(tagType string) string {
	if tagType == TagTypeAccount {
		return "ada"
	}
	return "ft"
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

func storeTagMembersInXdas(id string, tagType string, members <-chan string, savedMembers chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	xdasMembers := proto.XdasHashes{
		Fields: map[string]string{id: ""},
	}

	keyspace := keyspaceForTagType(tagType)
	successCount := 0
	failCount := 0

	for member := range members {
		normalized := NormalizeMember(member, tagType)
		err := GetGroupServiceSyncConnector().AddMembersToTagWithKeyspace(keyspace, normalized, &xdasMembers)
		if err != nil {
			failCount++
			log.Errorf("xdas error adding member to %s group: member=%s, error=%s", id, normalized, err.Error())
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

func removeTagMembersFromXdas(id string, tagType string, members <-chan string, removedMembers chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	keyspace := keyspaceForTagType(tagType)
	successCount := 0
	failCount := 0

	for member := range members {
		normalized := NormalizeMember(member, tagType)
		err := GetGroupServiceSyncConnector().RemoveGroupMembersWithKeyspace(keyspace, normalized, id)
		if err != nil {
			failCount++
			log.Errorf("xdas error removing member from %s group: member=%s, error=%s", id, normalized, err.Error())
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
