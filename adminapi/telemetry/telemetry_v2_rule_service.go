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
package telemetry

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	xcommon "github.com/rdkcentral/xconfadmin/common"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	queries "github.com/rdkcentral/xconfadmin/adminapi/queries"
	xshared "github.com/rdkcentral/xconfadmin/shared"
	"github.com/rdkcentral/xconfadmin/shared/logupload"
	xlogupload "github.com/rdkcentral/xconfadmin/shared/logupload"
	xutil "github.com/rdkcentral/xconfadmin/util"

	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	ru "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func GetAll() []*xwlogupload.TelemetryTwoRule {
	telemetryTwoRules := xwlogupload.GetTelemetryTwoRuleListForAS()
	sort.Slice(telemetryTwoRules, func(i, j int) bool {
		return strings.ToLower(telemetryTwoRules[i].Name) < strings.ToLower(telemetryTwoRules[j].Name)
	})
	return telemetryTwoRules
}

func GetOne(id string) (*xwlogupload.TelemetryTwoRule, error) {
	settingProfile := xlogupload.GetOneTelemetryTwoRule(id)
	if settingProfile == nil {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	return settingProfile, nil
}

func Delete(id string) (*xwlogupload.TelemetryTwoRule, error) {
	entity, err := GetOne(id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	DeleteTelemetryTwoRule(id)
	return entity, nil
}

func DeleteTelemetryTwoRule(id string) {
	err := xlogupload.DeleteTelemetryTwoRule(id)
	if err != nil {
		log.Warn("delete settingProfile failed")
	}
}

func findByContext(r *http.Request, searchContext map[string]string) []*xwlogupload.TelemetryTwoRule {
	telemetryTwoRulesFound := []*xwlogupload.TelemetryTwoRule{}

	telemetryTwoRules := xwlogupload.GetTelemetryTwoRuleListForAS()
	for _, telemetryTwoRule := range telemetryTwoRules {
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if applicationType != "" && applicationType != shared.ALL {
				if telemetryTwoRule.ApplicationType != applicationType {
					continue
				}
			}
		}
		if name, ok := xutil.FindEntryInContext(searchContext, xcommon.NAME_UPPER, false); ok {
			if name != "" {
				if !strings.Contains(strings.ToLower(telemetryTwoRule.Name), strings.ToLower(name)) {
					continue
				}
			}
		}
		if telemetrytwoprofile, ok := xutil.FindEntryInContext(searchContext, xcommon.PROFILE, false); ok {
			if len(telemetryTwoRule.BoundTelemetryIDs) == 0 {
				continue
			}
			telemetryprofileNameMatch := false
			for _, telemetryId := range telemetryTwoRule.BoundTelemetryIDs {
				telemetry := xwlogupload.GetOneTelemetryTwoProfile(telemetryId)
				if telemetry != nil && strings.Contains(strings.ToLower(telemetry.Name), strings.ToLower(telemetrytwoprofile)) {
					telemetryprofileNameMatch = true
					break
				}
			}
			if !telemetryprofileNameMatch {
				continue
			}
		}
		if key, ok := xutil.FindEntryInContext(searchContext, xcommon.FREE_ARG, false); ok {
			keyMatch := false
			for _, condition := range ru.ToConditions(&telemetryTwoRule.Rule) {
				if strings.Contains(strings.ToLower(condition.GetFreeArg().Name), strings.ToLower(key)) {
					keyMatch = true
					break
				}
			}
			if !keyMatch {
				continue
			}
		}
		if fixedArgValue, ok := xutil.FindEntryInContext(searchContext, xcommon.FIXED_ARG, false); ok {
			valueMatch := false
			for _, condition := range ru.ToConditions(&telemetryTwoRule.Rule) {
				if condition.GetFixedArg() != nil && condition.GetFixedArg().IsCollectionValue() {
					fixedArgs := condition.GetFixedArg().GetValue().([]string)
					for _, fixedArg := range fixedArgs {
						if strings.Contains(strings.ToLower(fixedArg), strings.ToLower(fixedArgValue)) {
							valueMatch = true
							break
						}
					}
				}
				if valueMatch {
					break
				}
				if condition.GetOperation() != rulesengine.StandardOperationExists && condition.GetFixedArg() != nil && condition.GetFixedArg().IsStringValue() {
					if strings.Contains(strings.ToLower(*condition.FixedArg.Bean.Value.JLString), strings.ToLower(fixedArgValue)) {
						valueMatch = true
						break
					}
				}
			}
			if !valueMatch {
				continue
			}
		}
		telemetryTwoRulesFound = append(telemetryTwoRulesFound, telemetryTwoRule)
	}
	return telemetryTwoRulesFound
}

func validate(entity *xwlogupload.TelemetryTwoRule) error {
	msg := validateProperties(entity)
	if msg != "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, msg)
	}
	return nil
}

func validateProperties(entity *xwlogupload.TelemetryTwoRule) string {
	if entity.Name == "" {
		return "Name is empty"
	}
	if !entity.NoOp && len(entity.BoundTelemetryIDs) == 0 {
		return "Profiles are not set"
	}
	if entity.NoOp && len(entity.BoundTelemetryIDs) > 0 {
		return "NoOp rule: profiles should be empty"
	}
	for _, boundTelemetryId := range entity.BoundTelemetryIDs {
		if boundTelemetryId == "" {
			continue
		}
		if logupload.GetOneTelemetryTwoProfile(boundTelemetryId) == nil {
			return "Telemetry 2.0 profile with id: " + boundTelemetryId + " does not exist"
		}
	}
	return ""
}

func validateAll(entity *xwlogupload.TelemetryTwoRule, existingEntities []*xwlogupload.TelemetryTwoRule) error {
	for _, rule := range existingEntities {
		if rule.ID == entity.ID {
			continue
		}
		if rule.Name == entity.Name {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Name is already used")
		}
		if ru.EqualComplexRules(&rule.Rule, &entity.Rule) {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Rule has duplicate: "+rule.Name)
		}
	}
	return nil
}

func TelemetryTwoRulesGeneratePage(list []*xwlogupload.TelemetryTwoRule, page int, pageSize int) (result []*xwlogupload.TelemetryTwoRule) {
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

func beforeCreating(entity *xwlogupload.TelemetryTwoRule, writeApplication string) error {
	id := entity.ID
	if id == "" {
		entity.ID = uuid.New().String()
	} else {
		existingEntity := xlogupload.GetOneTelemetryTwoRule(id)
		if existingEntity != nil && !xshared.ApplicationTypeEquals(existingEntity.ApplicationType, entity.ApplicationType) {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Entity with id: "+id+" already exists in "+existingEntity.ApplicationType+" application")
		} else if existingEntity != nil && xshared.ApplicationTypeEquals(existingEntity.ApplicationType, writeApplication) {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Entity with id: "+id+" already exists")
		}
	}
	return nil
}

func beforeUpdating(entity *xwlogupload.TelemetryTwoRule, writeApplication string) error {
	id := entity.ID
	if id == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Entity id is empty")
	}
	existingEntity := xlogupload.GetOneTelemetryTwoRule(id)
	if !xshared.ApplicationTypeEquals(existingEntity.ApplicationType, writeApplication) {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	if existingEntity == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	return nil
}

func beforeSaving(entity *xwlogupload.TelemetryTwoRule, writeApplication string) error {
	if entity != nil && entity.ApplicationType == "" {
		entity.ApplicationType = writeApplication
	}
	if entity != nil && !entity.Rule.Equals(rulesengine.NewEmptyRule()) {
		ru.NormalizeConditions(&entity.Rule)
	}

	if entity.ApplicationType != writeApplication {
		return fmt.Errorf("Current ApplicationType %s doesn't match with entity's ApplicationType: %s", writeApplication, entity.ApplicationType)
	}

	err := validate(entity)
	if err != nil {
		return err
	}
	all := xwlogupload.GetTelemetryTwoRuleListForAS()
	err = validateAll(entity, all)
	if err != nil {
		return err
	}
	return nil
}

func Create(entity *xwlogupload.TelemetryTwoRule, writeApplication string) error {
	err := beforeCreating(entity, writeApplication)
	if err != nil {
		return err
	}
	err = beforeSaving(entity, writeApplication)
	if err != nil {
		return err
	}
	err = queries.RunGlobalValidation(*entity.GetRule(), queries.GetAllowedOperations)
	if err != nil {
		return err
	}
	return xlogupload.SetOneTelemetryTwoRule(entity.ID, entity)
}

func Update(entity *xwlogupload.TelemetryTwoRule, writeApplication string) error {
	err := beforeUpdating(entity, writeApplication)
	if err != nil {
		return err
	}
	err = beforeSaving(entity, writeApplication)
	if err != nil {
		return err
	}
	err = queries.RunGlobalValidation(*entity.GetRule(), queries.GetAllowedOperations)
	if err != nil {
		return err
	}
	return xlogupload.SetOneTelemetryTwoRule(entity.ID, entity)
}
