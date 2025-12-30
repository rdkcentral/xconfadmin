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

func filterTagEntriesByPrefix(ftEntries []string) []string {
	tags := []string{}
	for _, ftEntry := range ftEntries {
		if strings.HasPrefix(ftEntry, Prefix) {
			tags = append(tags, RemovePrefixFromTag(ftEntry))
		}
	}
	return tags
}

func storeTagMembersInXdas(id string, members <-chan string, savedMembers chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	xdasMembers := proto.XdasHashes{
		Fields: map[string]string{id: ""},
	}
	for member := range members {
		normalizedEcm := ToNormalizedEcm(member)
		err := GetGroupServiceSyncConnector().AddMembersToTag(normalizedEcm, &xdasMembers)
		if err != nil {
			log.Errorf("xdas error adding %s member to %s group: %s", id, normalizedEcm, err.Error())
		} else {
			savedMembers <- member
		}
	}
}

func removeTagMembersFromXdas(id string, members <-chan string, removedMembers chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for member := range members {
		normalizedEcm := ToNormalizedEcm(member)
		err := GetGroupServiceSyncConnector().RemoveGroupMembers(normalizedEcm, id)
		if err != nil {
			log.Errorf("xdas error removing %s member from %s group: %s", id, normalizedEcm, err.Error())
		} else {
			removedMembers <- member
		}
	}
}

func CheckBatchSizeExceeded(batchSize int) error {
	config := GetTagApiConfig()
	if batchSize > config.BatchLimit {
		return fmt.Errorf(MaxMemberLimitExceededErrorMsg, batchSize, config.BatchLimit)
	}
	return nil
}
