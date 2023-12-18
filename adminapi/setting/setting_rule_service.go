/**
 * Copyright 2023 Comcast Cable Communications Management, LLC
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

	xwcommon "xconfwebconfig/common"
	"xconfwebconfig/dataapi/dcm/settings"

	xcommon "xconfadmin/common"

	"xconfadmin/adminapi/auth"
	"xconfadmin/adminapi/queries"
	"xconfadmin/shared"
	"xconfadmin/util"
	"xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	corefw "xconfwebconfig/shared/firmware"
	"xconfwebconfig/shared/logupload"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func GetAllSettingRules() []*logupload.SettingRule {
	SettingRules := GetSettingRulesList()
	sort.Slice(SettingRules, func(i, j int) bool {
		return strings.ToLower(SettingRules[i].Name) < strings.ToLower(SettingRules[j].Name)
	})
	return SettingRules
}

func GetOneSettingRule(id string) (*logupload.SettingRule, error) {
	settingRule := GetSettingRule(id)
	if settingRule == nil {
		return nil, xcommon.NewXconfError(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	return settingRule, nil
}

func DeleteSettingRule(id string, writeApplication string) (*logupload.SettingRule, error) {
	entity, err := GetOneSettingRule(id)
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

	DeleteSettingRuleOne(id)
	return entity, nil
}

func GetSettingRule(id string) *logupload.SettingRule {
	settingRule, err := db.GetCachedSimpleDao().GetOne(db.TABLE_SETTING_RULES, id)
	if err != nil {
		log.Warn("no settingRule found")
		return nil
	}
	return settingRule.(*logupload.SettingRule)
}

func DeleteSettingRuleOne(id string) {
	err := db.GetCachedSimpleDao().DeleteOne(db.TABLE_SETTING_RULES, id)
	if err != nil {
		log.Warn("delete settingRule failed")
	}
}

func SetSettingRule(id string, settingProfile *logupload.SettingRule) error {
	if err := db.GetCachedSimpleDao().SetOne(db.TABLE_SETTING_RULES, id, settingProfile); err != nil {
		log.Error("cannot save SettingRule to DB")
		return err
	}
	return nil
}

func GetSettingRulesList() []*logupload.SettingRule {
	var settingRules []*logupload.SettingRule
	rulesData, err := db.GetCachedSimpleDao().GetAllAsMap(db.TABLE_SETTING_RULES)
	if err == nil {
		for idx := range rulesData {
			settingRule := rulesData[idx].(*logupload.SettingRule)
			settingRules = append(settingRules, settingRule)
		}
	}
	return settingRules
}

func FindByContextSettingRule(r *http.Request, searchContext map[string]string) []*logupload.SettingRule {
	rules := GetSettingRulesList()
	rulesFound := []*logupload.SettingRule{}
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if applicationType, ok := util.FindEntryInContext(searchContext, xcommon.APPLICATION_TYPE, false); ok {
			if rule.ApplicationType != applicationType {
				continue
			}
		}
		if name, ok := util.FindEntryInContext(searchContext, xwcommon.NAME, false); ok {
			baseName := strings.ToLower(rule.Name)
			givenName := strings.ToLower(name)
			if !strings.Contains(baseName, givenName) {
				continue
			}
		}
		if key, ok := util.FindEntryInContext(searchContext, corefw.KEY, false); ok {
			if !re.IsExistConditionByFreeArgName(rule.Rule, key) {
				continue
			}
		}
		if value, ok := util.FindEntryInContext(searchContext, corefw.VALUE, false); ok {
			if !re.IsExistConditionByFixedArgValue(rule.Rule, value) {
				continue
			}
		}
		rulesFound = append(rulesFound, rule)
	}
	return rulesFound
}

func validateSettingRule(r *http.Request, entity *logupload.SettingRule) error {
	auth.ValidateWrite(r, entity.ApplicationType, auth.DCM_ENTITY)
	msg := validatePropertiesSettingRule(entity)
	if msg != "" {
		return xcommon.NewXconfError(http.StatusBadRequest, msg)
	}
	if entity == nil {
		return xcommon.NewXconfError(http.StatusBadRequest, "SettingRule is empty")
	}
	emptyRule := re.NewEmptyRule()
	if emptyRule.Equals(&entity.Rule) {
		return xcommon.NewXconfError(http.StatusBadRequest, "Rule is empty")
	}
	if err := queries.ValidateRuleStructure(&entity.Rule); err != nil {
		return err
	}
	if err := queries.RunGlobalValidation(entity.Rule, queries.GetAllowedOperations); err != nil {
		return err
	}
	return nil
}

func validatePropertiesSettingRule(entity *logupload.SettingRule) string {
	if entity.Name == "" {
		return "Name is empty"
	}
	if entity.BoundSettingID == "" {
		return "Setting profile is not present"
	}
	return ""
}

func validateAllSettingRule(ruleToCheck *logupload.SettingRule) error {
	existingSettingRules := GetAllSettingRules()
	for _, settingRule := range existingSettingRules {
		if settingRule.ID == ruleToCheck.ID {
			continue
		}
		if settingRule.ApplicationType != ruleToCheck.ApplicationType {
			continue
		}
		if ruleToCheck.GetName() == settingRule.GetName() {
			return xcommon.NewXconfError(http.StatusConflict, "\"Name is already used\"")
		}
		if ruleToCheck.GetRule().Equals(settingRule.GetRule()) {
			return xcommon.NewXconfError(http.StatusConflict, "Rule has duplicate: "+settingRule.GetName())
		}
	}
	return nil
}

func validateUsageSettingRule(id string) error {
	all := GetSettingRulesList()
	for _, rule := range all {
		if rule.BoundSettingID == id {
			return xcommon.NewXconfError(http.StatusConflict, "Can't delete profile as it's used in setting rule: "+rule.Name)
		}
	}
	return nil
}

func SettingRulesGeneratePage(list []*logupload.SettingRule, page int, pageSize int) (result []*logupload.SettingRule) {
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

func beforeCreatingSettingRule(r *http.Request, entity *logupload.SettingRule) error {
	id := entity.ID

	if id == "" {
		entity.ID = uuid.New().String()
	} else {
		existingEntity := GetSettingRule(id)
		writeApplication, err := auth.CanWrite(r, auth.DCM_ENTITY)
		if err != nil {
			return err
		}
		if existingEntity != nil && !shared.ApplicationTypeEquals(existingEntity.ApplicationType, entity.ApplicationType) {
			return xcommon.NewXconfError(http.StatusConflict, "Entity with id: "+id+" already exists in "+existingEntity.ApplicationType+" application")
		} else if existingEntity != nil && shared.ApplicationTypeEquals(existingEntity.ApplicationType, writeApplication) {
			return xcommon.NewXconfError(http.StatusConflict, "Entity with id: "+id+" already exists")
		}
	}
	return nil
}

func beforeUpdatingSettingRule(r *http.Request, entity *logupload.SettingRule) error {
	id := entity.ID
	if id == "" {
		return xcommon.NewXconfError(http.StatusBadRequest, "Entity id is empty")
	}
	existingEntity := GetSettingRule(id)

	writeApplication, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		return err
	}
	if !shared.ApplicationTypeEquals(existingEntity.ApplicationType, writeApplication) {
		return xcommon.NewXconfError(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	if existingEntity == nil {
		return xcommon.NewXconfError(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	return nil
}

func beforeSavingSettingRule(r *http.Request, entity *logupload.SettingRule) error {
	if entity != nil && entity.ApplicationType == "" {
		application, err := auth.CanWrite(r, auth.DCM_ENTITY)
		if err != nil {
			return err
		}
		entity.ApplicationType = application
	}
	if entity != nil && !entity.Rule.Equals(re.NewEmptyRule()) {
		re.NormalizeConditions(&entity.Rule)
	}
	err := validateSettingRule(r, entity)
	if err != nil {
		return err
	}
	err = validateAllSettingRule(entity)
	if err != nil {
		return err
	}
	return nil
}

func CreateSettingRule(r *http.Request, entity *logupload.SettingRule) error {
	err := beforeCreatingSettingRule(r, entity)
	if err != nil {
		return err
	}
	err = beforeSavingSettingRule(r, entity)
	if err != nil {
		return err
	}
	return SetSettingRule(entity.ID, entity)
}

func UpdateSettingRule(r *http.Request, entity *logupload.SettingRule) error {
	err := beforeUpdatingSettingRule(r, entity)
	if err != nil {
		return err
	}
	err = beforeSavingSettingRule(r, entity)
	if err != nil {
		return err
	}
	return SetSettingRule(entity.ID, entity)
}

func GetSettingRulesWithConfig(settingTypes []string, context map[string]string) map[string][]*logupload.SettingRule {
	result := make(map[string][]*logupload.SettingRule)

	for _, settingType := range settingTypes {
		settingRule := settings.GetSettingsRuleByTypeForContext(settingType, context)
		settingProfile := settings.GetSettingProfileBySettingRule(settingRule)
		if settingProfile == nil {
			continue
		}

		profileName := settingProfile.SettingProfileID
		settingRuleList := result[profileName]
		if settingRuleList == nil {
			settingRuleList = make([]*logupload.SettingRule, 0, 1)
		}
		settingRuleList = append(settingRuleList, settingRule)

		result[profileName] = settingRuleList
	}

	return result
}
