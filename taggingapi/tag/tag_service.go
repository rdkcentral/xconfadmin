package tag

import (
	"errors"
	"fmt"
	http2 "net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/rdkcentral/xconfadmin/common"
	"github.com/rdkcentral/xconfadmin/http"
	taggingapi_config "github.com/rdkcentral/xconfadmin/taggingapi/config"
	percentageutils "github.com/rdkcentral/xconfadmin/taggingapi/percentage"
	proto "github.com/rdkcentral/xconfadmin/taggingapi/proto/generated"
	"github.com/rdkcentral/xconfadmin/util"
	taggingds "github.com/rdkcentral/xconfwebconfig/tag"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	log "github.com/sirupsen/logrus"
)

const (
	percentageTag            = "p:%v"
	StringToIntConversionErr = "error converting string %s value to int: %s"
	IncorrectRangeErr        = "start range should be greater then end range"
	MinStartPercentage       = 0
	MaxEndPercentage         = 100
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

func GetTagById(id string) *taggingds.Tag {
	tag := GetOneTag(SetTagPrefix(id))
	if tag != nil {
		tag.Id = RemovePrefixFromTag(tag.Id)
	}
	return tag
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

func GetTagMembers(id string) ([]string, error) {
	id = SetTagPrefix(id)
	tag := GetOneTag(id)
	if tag == nil {
		return []string{}, xwcommon.NewRemoteErrorAS(http2.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, id))
	}
	converted := make([]string, len(tag.Members))
	for i, member := range tag.Members.ToSlice() {
		converted[i] = ToEstbIfMac(member)
	}
	return converted, nil
}

func AddMembersToTag(id string, members []string) (int, error) {
	id = SetTagPrefix(id)

	membersChannel := make(chan string, len(members))
	go func() {
		defer close(membersChannel)
		for _, member := range members {
			membersChannel <- member
		}
	}()

	wg := &sync.WaitGroup{}

	savedMembersChannel := make(chan string, len(members))
	config := GetTagApiConfig()
	numOfWorkers := config.WorkerCount
	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go storeTagMembersInXdas(id, membersChannel, savedMembersChannel, wg)
	}

	go func() {
		wg.Wait()
		close(savedMembersChannel)
	}()

	var savedMembers []string
	for savedMember := range savedMembersChannel {
		savedMembers = append(savedMembers, savedMember)
	}

	updatedTag := AddMembersToXconfTag(id, savedMembers)
	err := SaveTag(updatedTag)
	if err != nil {
		return 0, err
	}
	return len(updatedTag.Members), nil
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

func RemoveMemberFromTag(id string, member string) (*taggingds.Tag, error) {
	id = SetTagPrefix(id)
	normalizedEcm := ToNormalizedEcm(member)
	err := GetGroupServiceSyncConnector().RemoveGroupMembers(normalizedEcm, id)
	if err != nil {
		log.Errorf("xdas error removing %s member from %s group: %s", id, normalizedEcm, err.Error())
		return nil, err
	}

	tag := GetOneTag(id)
	if tag == nil {
		return nil, xwcommon.NewRemoteErrorAS(http2.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, id))
	}
	tag = removeMembersFromXconfTag(tag, []string{ToNormalized(member)})
	err = saveOrRemove(tag)

	if err != nil {
		return nil, err
	}
	return tag, nil
}

func RemoveMembersFromTag(id string, members []string) (int, error) {
	id = SetTagPrefix(id)

	membersChannel := make(chan string, len(members))
	go func() {
		defer close(membersChannel)
		for _, member := range members {
			membersChannel <- member
		}
	}()

	wg := &sync.WaitGroup{}
	removedMembersChannel := make(chan string, len(members))
	config := GetTagApiConfig()
	numOfWorkers := config.WorkerCount
	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go removeTagMembersFromXdas(id, membersChannel, removedMembersChannel, wg)
	}

	go func() {
		wg.Wait()
		close(removedMembersChannel)
	}()

	var removedMembers []string
	for member := range removedMembersChannel {
		removedMembers = append(removedMembers, member)
	}

	tag := GetOneTag(id)
	if tag == nil {
		return 0, xwcommon.NewRemoteErrorAS(http2.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, id))
	}
	tag = removeMembersFromXconfTag(tag, removedMembers)
	err := saveOrRemove(tag)
	if err != nil {
		return 0, err
	}
	return len(tag.Members), nil
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

func removeMembersFromXdasTag(id string, members []string) ([]string, error) {
	var removeFromXconf []string
	for _, member := range members {
		normalizedMember := ToNormalizedEcm(member)
		err := GetGroupServiceSyncConnector().RemoveGroupMembers(normalizedMember, id)
		if err != nil {
			if common.GetXconfErrorStatusCode(err) == http2.StatusNotFound { //ignoring NOT FOUND error if tag was removed by any other way, to not block entry removal from xconf
				log.Warnf("%s member was not found in %s group", id, normalizedMember)
				removeFromXconf = append(removeFromXconf, member)
				continue
			}
			return removeFromXconf, err
		}
		removeFromXconf = append(removeFromXconf, normalizedMember)
	}
	return removeFromXconf, nil
}

func saveOrRemove(tag *taggingds.Tag) error {
	if len(tag.Members) > 0 {
		return SaveTag(tag)
	} else {
		return DeleteOneTag(tag.Id)
	}
}

func DeleteTag(id string) error {
	id = SetTagPrefix(id)
	tag := GetOneTag(id)
	if tag == nil {
		return xwcommon.NewRemoteErrorAS(http2.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, id))
	}
	tag, err := deleteTagFromXdas(tag)
	if err != nil && len(tag.Members) > 0 {
		if saveErr := SaveTag(tag); saveErr != nil {
			return errors.Join(err, saveErr)
		}
		return err
	}

	return DeleteOneTag(id)
}

func deleteTagFromXdas(tag *taggingds.Tag) (*taggingds.Tag, error) {
	var removedMembers []string
	var err error
	for _, member := range tag.Members.ToSlice() {
		normalizedMember := ToNormalizedEcm(member)
		xdasErr := GetGroupServiceSyncConnector().RemoveGroupMembers(normalizedMember, tag.Id)
		if xdasErr != nil {
			log.Errorf("xdas error removing %s member from %s group: %s", tag.Id, normalizedMember, xdasErr.Error())
			if common.GetXconfErrorStatusCode(xdasErr) == http2.StatusNotFound { //ignoring NOT FOUND error if tag was removed by any other way, to not block entry removal from xconf
				removedMembers = append(removedMembers, member)
				continue
			}
			err = xdasErr
			break
		}
		removedMembers = append(removedMembers, member)
	}
	tag = removeMembersFromXconfTag(tag, removedMembers)
	return tag, err
}

func AddAccountRangeToTag(id string, startRangeStr string, endRangeStr string) error {
	startRange, err := strconv.Atoi(startRangeStr)
	if err != nil {
		return xwcommon.NewRemoteErrorAS(http2.StatusBadRequest, fmt.Sprintf(StringToIntConversionErr, startRangeStr, err.Error()))
	}
	endRange, err := strconv.Atoi(endRangeStr)
	if err != nil {
		return xwcommon.NewRemoteErrorAS(http2.StatusBadRequest, fmt.Sprintf(StringToIntConversionErr, endRangeStr, err.Error()))
	}
	if startRange >= endRange {
		return xwcommon.NewRemoteErrorAS(http2.StatusBadRequest, IncorrectRangeErr)
	}
	if err := CleanPercentageRange(id); err != nil {
		return err
	}
	accountPercentages := buildPercentageRangeMembers(startRange, endRange)
	_, err = AddMembersToTag(id, accountPercentages)
	if err != nil {
		return err
	}
	return nil
}

func CleanPercentageRange(id string) error {
	id = SetTagPrefix(id)
	tag := GetOneTag(id)
	if tag == nil {
		return xwcommon.NewRemoteErrorAS(http2.StatusNotFound, fmt.Sprintf(NotFoundErrorMsg, id))
	}
	accountPercentages := buildPercentageRangeMembers(MinStartPercentage, MaxEndPercentage)
	removedXdasGroups, xdasErr := removeMembersFromXdasTag(id, accountPercentages)
	tag = removeMembersFromXconfTag(tag, removedXdasGroups)
	var xconfErr error
	if len(tag.Members) > 0 {
		if saveErr := SaveTag(tag); saveErr != nil {
			xconfErr = saveErr
		}
	} else {
		if deleteErr := DeleteOneTag(id); deleteErr != nil {
			xconfErr = deleteErr
		}
	}
	return errors.Join(xdasErr, xconfErr)
}

func GetTagsByMemberPercentage(member string) ([]string, error) {
	memberPercentage := percentageutils.CalculatePercent(member)
	tagMember := fmt.Sprintf(percentageTag, memberPercentage)
	return GetTagsByMember(tagMember)
}

func buildPercentageRangeMembers(startRange int, endRange int) []string {
	var members []string
	for start := startRange; start <= endRange; start++ {
		accountPercentage := fmt.Sprintf(percentageTag, start)
		members = append(members, accountPercentage)
	}
	return members
}

func CheckBatchSizeExceeded(batchSize int) error {
	config := GetTagApiConfig()
	if batchSize > config.BatchLimit {
		return fmt.Errorf(MaxMemberLimitExceededErrorMsg, batchSize, config.BatchLimit)
	}
	return nil
}
