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
package change

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/rdkcentral/xconfadmin/shared"
	xshared "github.com/rdkcentral/xconfadmin/shared"
	xutil "github.com/rdkcentral/xconfadmin/util"

	xcommon "github.com/rdkcentral/xconfadmin/common"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xchange "github.com/rdkcentral/xconfadmin/shared/change"
	xlogupload "github.com/rdkcentral/xconfadmin/shared/logupload"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	xwshared "github.com/rdkcentral/xconfwebconfig/shared"
	xwchange "github.com/rdkcentral/xconfwebconfig/shared/change"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
)

func GetTelemetryTwoProfilesByIdList(appType string, idList []string) []xwlogupload.TelemetryTwoProfile {
	telemetryTwoProfiles := []xwlogupload.TelemetryTwoProfile{}
	list := xlogupload.GetAllTelemetryTwoProfileList(appType)
	for _, profile := range list {
		for _, id := range idList {
			if profile.ID == id {
				telemetryTwoProfiles = append(telemetryTwoProfiles, *profile)
			}
		}
	}

	return telemetryTwoProfiles
}

func WriteCreateChangeTelemetryTwoProfile(r *http.Request, profile *xwlogupload.TelemetryTwoProfile) (*xwchange.TelemetryTwoChange, error) {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY, profile.ApplicationType)
	if err != nil {
		return nil, err
	}
	if err := beforeCreatingTelemetryTwoProfile(profile, applicationType); err != nil {
		return nil, err
	}

	if err := beforeSavingTelemetryTwoProfile(profile); err != nil {
		return nil, err
	}

	if err := ValidateTelemetryTwoProfilePendingChanges(profile); err != nil {
		return nil, err
	}

	if _, err := auth.CanWrite(r, auth.CHANGE_ENTITY, profile.ApplicationType); err != nil {
		return nil, err
	}
	change := buildToCreateTelemetryTwoChange(profile, applicationType, auth.GetUserNameOrUnknown(r))
	if err := xchange.CreateOneTelemetryTwoChange(change); err != nil {
		return nil, err
	}
	return change, nil
}

func WriteUpdateChangeOrSaveTelemetryTwoProfile(r *http.Request, newProfile *xwlogupload.TelemetryTwoProfile) (*xwchange.TelemetryTwoChange, error) {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		return nil, err
	}
	if err := beforeUpdatingTelemetryTwoProfile(newProfile, applicationType); err != nil {
		return nil, err
	}
	if err := beforeSavingTelemetryTwoProfile(newProfile); err != nil {
		return nil, err
	}

	var change *xwchange.TelemetryTwoChange

	oldProfile := xlogupload.GetOneTelemetryTwoProfile(newProfile.ID)
	if newProfile.Equals(oldProfile) {
		if err := xlogupload.SetOneTelemetryTwoProfile(newProfile); err != nil {
			return nil, err
		}
	} else {
		if _, err := auth.CanWrite(r, auth.CHANGE_ENTITY); err != nil {
			return nil, err
		}
		change = buildToUpdateTelemetryTwoChange(oldProfile, newProfile, applicationType, auth.GetUserNameOrUnknown(r))
		if err := beforeSavingTelemetryTwoChange(r, change); err != nil {
			return nil, err
		}
		if err := xchange.CreateOneTelemetryTwoChange(change); err != nil {
			return nil, err
		}

	}
	return change, nil
}

func WriteDeleteChangeTelemetryTwoProfile(r *http.Request, id string) (*xwchange.TelemetryTwoChange, error) {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		return nil, err
	}
	deleteProfile, err := beforeRemovingTelemetryTwoProfile(id, applicationType)
	if err != nil {
		return nil, err
	}
	if _, err := auth.CanWrite(r, auth.CHANGE_ENTITY); err != nil {
		return nil, err
	}
	change := buildToDeleteTelemetryTwoChange(deleteProfile, applicationType, auth.GetUserNameOrUnknown(r))
	err = xchange.CreateOneTelemetryTwoChange(change)
	if err != nil {
		return nil, err
	}
	return change, nil
}

func GetTelemetryTwoProfilesByContext(searchContext map[string]string) []*xwlogupload.TelemetryTwoProfile {
	filteredProfiles := []*xwlogupload.TelemetryTwoProfile{}
	applicationType, ok := xutil.FindEntryInContext(searchContext, shared.APPLICATION_TYPE, false)
	if !ok {
		return filteredProfiles
	}
	profiles := xlogupload.GetAllTelemetryTwoProfileList(applicationType)
	for _, profile := range profiles {
		if name, ok := xutil.FindEntryInContext(searchContext, xcommon.NAME_UPPER, false); ok {
			if !xutil.ContainsIgnoreCase(profile.Name, name) {
				continue
			}
		}
		filteredProfiles = append(filteredProfiles, profile)
	}
	return filteredProfiles
}

func GeneratePageTelemetryTwoProfiles(list []*xwlogupload.TelemetryTwoProfile, page int, pageSize int) (result []*xwlogupload.TelemetryTwoProfile) {
	sort.Slice(list, func(i, j int) bool {
		return strings.Compare(strings.ToLower(list[i].Name), strings.ToLower(list[j].Name)) < 0
	})

	length := len(list)
	startIndex := page*pageSize - pageSize
	if page < 1 || startIndex > length || pageSize < 1 {
		return result
	}
	lastIndex := length
	if page*pageSize < length {
		lastIndex = page * pageSize
	}
	return list[startIndex:lastIndex]
}

func CreateTelemetryTwoProfile(r *http.Request, newProfile *xwlogupload.TelemetryTwoProfile) (*xwlogupload.TelemetryTwoProfile, error) {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY, newProfile.ApplicationType)
	if err != nil {
		return nil, err
	}
	if err := beforeCreatingTelemetryTwoProfile(newProfile, applicationType); err != nil {
		return nil, err
	}
	if err := beforeSavingTelemetryTwoProfile(newProfile); err != nil {
		return nil, err
	}
	if err := xlogupload.SetOneTelemetryTwoProfile(newProfile); err != nil {
		return nil, err
	}
	return newProfile, nil
}

func UpdateTelemetryTwoProfile(r *http.Request, profile *xwlogupload.TelemetryTwoProfile) (*xwlogupload.TelemetryTwoProfile, error) {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY, profile.ApplicationType)
	if err != nil {
		return nil, err
	}
	if err := beforeUpdatingTelemetryTwoProfile(profile, applicationType); err != nil {
		return nil, err
	}
	if err := beforeSavingTelemetryTwoProfile(profile); err != nil {
		return nil, err
	}
	if err := xlogupload.SetOneTelemetryTwoProfile(profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func DeleteTelemetryTwoProfile(r *http.Request, id string) error {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		return err
	}
	if _, err := beforeRemovingTelemetryTwoProfile(id, applicationType); err != nil {
		return err
	}
	if err := xlogupload.DeleteTelemetryTwoProfile(id); err != nil {
		return err
	}
	return nil
}

func beforeCreatingTelemetryTwoProfile(entity *xwlogupload.TelemetryTwoProfile, writeApplication string) error {
	if entity.ID == "" {
		entity.ID = uuid.New().String()
	} else {
		existingEntity := xwlogupload.GetOneTelemetryTwoProfile(entity.ID)
		if existingEntity != nil {
			if !xshared.ApplicationTypeEquals(existingEntity.ApplicationType, entity.ApplicationType) {
				return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Entity with id: %s already exists in %s application", entity.ID, existingEntity.ApplicationType))
			} else if xshared.ApplicationTypeEquals(existingEntity.ApplicationType, writeApplication) {
				return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Entity with id: %s already exists", entity.ID))

			}
		}
	}
	return nil
}

func beforeUpdatingTelemetryTwoProfile(entity *xwlogupload.TelemetryTwoProfile, writeApplication string) error {
	if entity.ID == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity id is empty")
	}
	existingEntity := xwlogupload.GetOneTelemetryTwoProfile(entity.ID)
	if existingEntity == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("Entity with id: %s does not exist", entity.ID))
	} else {
		if !shared.ApplicationTypeEquals(existingEntity.ApplicationType, writeApplication) {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("[%s] - current applicationType [%s] does not match with write applicationType [%s]", entity.ID, existingEntity.ApplicationType, writeApplication))
		}
		if !shared.ApplicationTypeEquals(existingEntity.ApplicationType, entity.ApplicationType) {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Entity with id: %s already exists in %s application", entity.ID, existingEntity.ApplicationType))
		}
	}
	return nil
}

func beforeSavingTelemetryTwoProfile(entity *xwlogupload.TelemetryTwoProfile) error {
	// Type attribute is mandatory and currently there is only one type and it is TelemetryTwoProfile.
	entity.Type = "TelemetryTwoProfile"

	if err := entity.Validate(); err != nil {
		return err
	}
	existingEntities := xlogupload.GetAllTelemetryTwoProfileList(entity.ApplicationType)
	return entity.ValidateAll(existingEntities)
}

func ValidateTelemetryTwoProfilePendingChanges(entity *xwlogupload.TelemetryTwoProfile) error {
	telemetryTwoProfilechanges := xchange.GetAllTelemetryTwoChangeList()
	for _, change := range telemetryTwoProfilechanges {
		if change.ID != entity.ID {
			if change.NewEntity != nil && entity.EqualChangeData(change.NewEntity) {
				return xwcommon.NewRemoteErrorAS(http.StatusConflict, "The same change already exists")
			} else if change.OldEntity != nil && entity.EqualChangeData(change.OldEntity) {
				return xwcommon.NewRemoteErrorAS(http.StatusConflict, "The same change already exists")
			}
		}
	}
	return nil
}

func beforeRemovingTelemetryTwoProfile(id string, writeApplication string) (*xwlogupload.TelemetryTwoProfile, error) {
	entity := xwlogupload.GetOneTelemetryTwoProfile(id)
	if entity == nil || !shared.ApplicationTypeEquals(writeApplication, entity.ApplicationType) {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("Entity with id: %s does not exist", id))
	}
	if err := validateUsageTelemetryTwoProfile(id); err != nil {
		return nil, err
	}
	return entity, nil
}

func validateUsageTelemetryTwoProfile(id string) error {
	all := xwlogupload.GetTelemetryTwoRuleListForAS() // []*TelemetryTwoRule
	for _, rule := range all {
		if util.Contains(rule.BoundTelemetryIDs, id) {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Can't delete profile as it's used in telemetry rule: %s", rule.Name))
		}
	}

	return nil
}

func TelemetryTwoTestPageFilterByContext(searchContext map[string]string) []*xwlogupload.TelemetryTwoRule {
	var keyMatch bool
	var valueMatch bool
	TelemetryRuleList := []*xwlogupload.TelemetryTwoRule{}
	TelemetryRules := xwlogupload.GetTelemetryTwoRuleListForAS()
	for _, tmRule := range TelemetryRules {
		if tmRule == nil {
			continue
		}

		if applicationType, ok := searchContext[xwcommon.APPLICATION_TYPE]; ok {
			if tmRule.ApplicationType != applicationType && tmRule.ApplicationType != xwshared.ALL {
				continue
			}
		}
		for key, fixedArgValue := range searchContext {

			keyMatch = false
			for _, condition := range rulesengine.ToConditions(tmRule.GetRule()) {
				if strings.Contains(strings.ToLower(condition.GetFreeArg().Name), strings.ToLower(key)) {
					keyMatch = true
					break
				}
			}
			if !keyMatch {
				continue
			}
			valueMatch = false
			for _, condition := range rulesengine.ToConditions(tmRule.GetRule()) {
				if condition.GetFixedArg() != nil && condition.GetFixedArg().IsCollectionValue() {
					fixedArgs := condition.GetFixedArg().GetValue().([]string)
					for _, fixedArg := range fixedArgs {
						if strings.EqualFold(strings.ToLower(fixedArg), strings.ToLower(fixedArgValue)) {
							valueMatch = true
							break
						}
					}
				}
				if valueMatch {
					break
				}

				if condition.GetOperation() != rulesengine.StandardOperationExists && condition.GetFixedArg() != nil && condition.GetFixedArg().IsStringValue() {
					if strings.EqualFold(strings.ToLower(*condition.FixedArg.Bean.Value.JLString), strings.ToLower(fixedArgValue)) {
						valueMatch = true
						break
					}
				}

			}
			if !valueMatch {
				continue
			}
			if keyMatch && valueMatch {
				break
			}
		}
		if !(keyMatch && valueMatch) {
			continue
		}
		TelemetryRuleList = append(TelemetryRuleList, tmRule)
	}
	return TelemetryRuleList
}
