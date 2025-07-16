/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package setting

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	xcommon "xconfadmin/common"

	"xconfadmin/shared"
	xlogupload "xconfadmin/shared/logupload"
	"xconfadmin/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func GetAll() []*xwlogupload.SettingProfiles {
	SettingProfiles := GetSettingProfileList()
	sort.Slice(SettingProfiles, func(i, j int) bool {
		return strings.ToLower(SettingProfiles[i].SettingProfileID) < strings.ToLower(SettingProfiles[j].SettingProfileID)
	})
	return SettingProfiles
}

func GetOne(id string) (*xwlogupload.SettingProfiles, error) {
	settingProfile := xlogupload.GetOneSettingProfile(id)
	if settingProfile == nil {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	return settingProfile, nil
}

func Delete(id string, writeApplication string) (*xwlogupload.SettingProfiles, error) {
	entity, err := GetOne(id)
	if err != nil {
		return nil, err
	}
	err = validateUsage(id)
	if err != nil {
		return nil, err
	}
	if entity.ApplicationType != writeApplication {
		return nil, fmt.Errorf("Entity with id %s ApplicationType doesn't match", id)
	}
	DeleteSettingProfile(id)
	return entity, nil
}

func GetSettingProfileList() []*xwlogupload.SettingProfiles {
	all := []*xwlogupload.SettingProfiles{}
	settingProfileList, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_SETTING_PROFILES, 0)
	if err != nil {
		log.Warn("no SettingProfiles found")
		return nil
	}
	for idx := range settingProfileList {
		settingProfile := settingProfileList[idx].(*xwlogupload.SettingProfiles)
		all = append(all, settingProfile)
	}
	return all
}

func DeleteSettingProfile(id string) {
	err := ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_SETTING_PROFILES, id)
	if err != nil {
		log.Warn("delete settingProfile failed")
	}
}

func SetSettingProfile(id string, settingProfile *xwlogupload.SettingProfiles) error {
	if err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_SETTING_PROFILES, id, settingProfile); err != nil {
		log.Error("cannot save settingProfile to DB")
		return err
	}
	return nil
}

func FindByContext(searchContext map[string]string) []*xwlogupload.SettingProfiles {
	profiles := GetSettingProfileList()
	profilesFound := []*xwlogupload.SettingProfiles{}
	for _, profile := range profiles {
		if applicationType, ok := util.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if profile.ApplicationType != applicationType {
				continue
			}
		}
		if name, ok := util.FindEntryInContext(searchContext, xwcommon.NAME, false); ok {
			if name != "" {
				baseName := strings.ToLower(profile.SettingProfileID)
				givenName := strings.ToLower(name)
				if !strings.Contains(baseName, givenName) {
					continue
				}
			}
		}
		if typeName, ok := util.FindEntryInContext(searchContext, xcommon.TYPE, false); ok {
			if typeName != "" {
				baseName := strings.ToLower(profile.SettingType)
				givenName := strings.ToLower(typeName)
				if !strings.Contains(baseName, givenName) {
					continue
				}
			}
		}
		profilesFound = append(profilesFound, profile)
	}
	return profilesFound
}

func validate(entity *xwlogupload.SettingProfiles) error {
	msg := validateProperties(entity)
	if msg != "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, msg)
	}
	return nil
}

func validateProperties(entity *xwlogupload.SettingProfiles) string {
	if entity.SettingType == "" {
		return "Setting type is empty"
	}
	if !xlogupload.IsValidSettingType(entity.SettingType) {
		return entity.SettingType + " not one of declared Enum instance names: [PARTNER_SETTINGS, EPON]"
	}
	if entity.Properties == nil || len(entity.Properties) == 0 {
		return "Property map is empty"
	}
	for key, value := range entity.Properties {
		if key == "" {
			return "Key is blank"
		}
		if value == "" {
			return "Value is blank for key: " + key
		}
	}
	return ""
}

func validateAll(entity *xwlogupload.SettingProfiles, existingEntities []*xwlogupload.SettingProfiles) error {
	for _, profile := range existingEntities {
		if profile.ID != entity.ID && profile.SettingProfileID == entity.SettingProfileID {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "SettingProfile with such settingProfileId exists: "+entity.SettingProfileID)
		}
	}
	return nil
}

func validateUsage(id string) error {
	all := GetSettingRulesList()
	for _, rule := range all {
		if rule.BoundSettingID == id {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Can't delete profile as it's used in setting rule: "+rule.Name)
		}
	}
	return nil
}

func SettingProfilesGeneratePage(list []*xwlogupload.SettingProfiles, page int, pageSize int) (result []*xwlogupload.SettingProfiles) {
	leng := len(list)
	startIndex := page*pageSize - pageSize
	if page < 1 || startIndex > leng || pageSize < 1 {
		return result
	}
	lastIndex := leng
	if page*pageSize < len(list) {
		lastIndex = page * pageSize
	}
	return list[startIndex:lastIndex]
}

func beforeCreating(entity *xwlogupload.SettingProfiles, writeApplication string) error {
	id := entity.ID
	if id == "" {
		entity.ID = uuid.New().String()
	} else {
		existingEntity := xlogupload.GetOneSettingProfile(id)
		if existingEntity != nil && !shared.ApplicationTypeEquals(existingEntity.ApplicationType, entity.ApplicationType) {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Entity with id: "+id+" already exists in "+existingEntity.ApplicationType+" application")
		} else if existingEntity != nil && shared.ApplicationTypeEquals(existingEntity.ApplicationType, writeApplication) {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Entity with id: "+id+" already exists")
		}
	}
	return nil
}

func beforeUpdating(entity *xwlogupload.SettingProfiles, writeApplication string) error {
	id := entity.ID
	if id == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Entity id is empty")
	}
	existingEntity := xlogupload.GetOneSettingProfile(id)
	if !shared.ApplicationTypeEquals(existingEntity.ApplicationType, writeApplication) {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	if existingEntity == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	return nil
}

func beforeSaving(entity *xwlogupload.SettingProfiles, writeApplication string) error {
	if entity != nil && entity.ApplicationType == "" {
		entity.ApplicationType = writeApplication
	}
	err := validate(entity)
	if err != nil {
		return err
	}
	all := GetSettingProfileList()
	err = validateAll(entity, all)
	if err != nil {
		return err
	}
	return nil
}

func Create(entity *xwlogupload.SettingProfiles, applicationType string) error {
	err := beforeCreating(entity, applicationType)
	if err != nil {
		return err
	}
	err = beforeSaving(entity, applicationType)
	if err != nil {
		return err
	}
	return SetSettingProfile(entity.ID, entity)
}

func Update(entity *xwlogupload.SettingProfiles, applicationType string) error {
	err := beforeUpdating(entity, applicationType)
	if err != nil {
		return err
	}
	err = beforeSaving(entity, applicationType)
	if err != nil {
		return err
	}
	return SetSettingProfile(entity.ID, entity)
}
