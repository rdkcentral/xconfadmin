package tag

import (
	"fmt"
	"math"

	ds "github.com/rdkcentral/xconfwebconfig/db"

	"github.com/rdkcentral/xconfadmin/util"

	log "github.com/sirupsen/logrus"
)

const (
	CloneErrorMsg = "error cloning %s tag"
)

func GetAllTags() ([]*Tag, error) {
	insts, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_TAG, math.MaxInt)
	tags := make([]*Tag, len(insts))
	for i, inst := range insts {
		tag := inst.(*Tag)
		cloned, _ := tag.Clone()
		cloned.Id = RemovePrefixFromTag(cloned.Id)
		tags[i] = cloned
	}
	return tags, err
}

func GetAllTagIds() ([]string, error) {
	tagIds, err := ds.GetCachedSimpleDao().GetKeys(ds.TABLE_TAG)
	if err != nil {
		return []string{}, err
	}
	tags := []string{}
	for _, tag := range tagIds {
		tags = append(tags, RemovePrefixFromTag(tag.(string)))
	}
	return tags, nil
}

func GetOneTag(id string) *Tag {
	inst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TAG, id)
	if err != nil {
		log.Warn(fmt.Sprintf(NotFoundErrorMsg, id))
		return nil
	}
	tag := inst.(*Tag)
	clone, err := tag.Clone()
	if err != nil {
		log.Error(fmt.Sprintf(CloneErrorMsg, id))
		return nil
	}
	return clone
}

func SaveTag(tag *Tag) error {
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_TAG, tag.Id, tag)
	if err != nil {
		return err
	}
	return nil
}

func DeleteOneTag(id string) error {
	err := ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_TAG, id)
	if err != nil {
		return err
	}
	return nil
}

func AddMembersToXconfTag(id string, members []string) *Tag {
	tag := GetOneTag(id)
	if tag == nil {
		memberSet := util.Set{}
		memberSet.Add(members...)
		return &Tag{Id: id, Members: memberSet, Updated: util.GetTimestamp()}
	}
	tag.Members.Add(members...)
	return tag
}

func removeMembersFromXconfTag(tag *Tag, members []string) *Tag {
	for _, member := range members {
		tag.Members.Remove(member)
	}
	return tag
}
