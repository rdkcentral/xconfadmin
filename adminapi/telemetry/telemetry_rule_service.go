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
package telemetry

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	ru "xconfwebconfig/rulesengine"

	queries "xconfadmin/adminapi/queries"
	xcommon "xconfadmin/common"
	xlogupload "xconfadmin/shared/logupload"
	xutil "xconfadmin/util"
	xwcommon "xconfwebconfig/common"
	"xconfwebconfig/db"
	xwhttp "xconfwebconfig/http"
	"xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	xwlogupload "xconfwebconfig/shared/logupload"
	xwutil "xconfwebconfig/util"

	"github.com/google/uuid"
)

func validateUsageForTelemetryRule(Id string, app string) (string, error) {
	tmRule := xlogupload.GetOneTelemetryRule(Id)
	if tmRule == nil {
		return fmt.Sprintf("Entity with id  %s does not exist ", Id), nil
	}
	if tmRule.ApplicationType != app {
		return fmt.Sprintf("Entity with id  %s does not exist ", Id), nil
	}
	return "", nil
}

func DeleteTelemetryRulebyId(id string, app string) *xwhttp.ResponseEntity {
	usage, err := validateUsageForTelemetryRule(id, app)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, err, nil)
	}

	if usage != "" {
		return xwhttp.NewResponseEntity(http.StatusNotFound, errors.New(usage), nil)
	}

	err = DeleteOneTelemetryRule(id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func DeleteOneTelemetryRule(id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(db.TABLE_TELEMETRY_RULES, id)
	if err != nil {
		return err
	}
	return nil
}

func telemetryRuleValidate(tmrule *xwlogupload.TelemetryRule) *xwhttp.ResponseEntity {
	if tmrule == nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("DCM formula Rule should be specified"), nil)
	}

	if xwutil.IsBlank(tmrule.ApplicationType) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ApplicationType is empty"), nil)
	}

	if xwutil.IsBlank(tmrule.Name) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Name is empty"), nil)
	}

	if xwutil.IsBlank(tmrule.BoundTelemetryID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("BoundTelemetryID is empty"), nil)
	} else {
		profile := xwlogupload.GetOnePermanentTelemetryProfile(tmrule.BoundTelemetryID)
		if profile == nil {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("BoundTelemetryID does not exist"), nil)
		}
	}
	if tmrule.GetRule() != nil {
		ru.NormalizeConditions(tmrule.GetRule())
	} else {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Condition is empty"), nil)
	}
	err := queries.ValidateRuleStructure(tmrule.GetRule())
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}
	err = queries.RunGlobalValidation(*tmrule.GetRule(), queries.GetAllowedOperations)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}
	tmrules := xwlogupload.GetTelemetryRuleList()
	for _, extmrule := range tmrules {
		if extmrule.ApplicationType != tmrule.ApplicationType {
			continue
		}
		if extmrule.ID != tmrule.ID {
			if extmrule.ApplicationType != tmrule.ApplicationType {
				continue
			}
			if extmrule.Name == tmrule.Name {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Name is alread used"), nil)
			}
			rule1 := extmrule.GetRule()
			rule2 := tmrule.GetRule()
			if rule1.Equals(rule2) {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Rule has duplicates: %s", extmrule.Name), nil)
			}
		}
	}
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, nil)
}

func CreateTelemetryRule(tmrule *xwlogupload.TelemetryRule, app string) *xwhttp.ResponseEntity {
	if xwutil.IsBlank(tmrule.ID) {
		tmrule.ID = uuid.New().String()
	} else {
		existingRule := xlogupload.GetOneTelemetryRule(tmrule.ID)
		if existingRule != nil {
			return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s already exists", tmrule.ID), nil)
		}
	}

	if tmrule.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", tmrule.ID), nil)
	}
	respEntity := telemetryRuleValidate(tmrule)
	if respEntity.Error != nil {
		return respEntity
	}
	if xwutil.IsBlank(tmrule.ApplicationType) {
		tmrule.ApplicationType = app
	}

	tmrule.Updated = xwutil.GetTimestamp(time.Now().UTC())
	if err := db.GetCachedSimpleDao().SetOne(db.TABLE_TELEMETRY_RULES, tmrule.ID, tmrule); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, tmrule)
}

func UpdateTelemetryRule(tmrule *xwlogupload.TelemetryRule, app string) *xwhttp.ResponseEntity {
	if xwutil.IsBlank(tmrule.ID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New(" ID  is empty"), nil)
	}
	if tmrule.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", tmrule.ID), nil)
	}
	tmruleex, err := db.GetCachedSimpleDao().GetOne(db.TABLE_TELEMETRY_RULES, tmrule.ID)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s does not exist", tmrule.ID), nil)
	}
	tmruleDB := tmruleex.(*xwlogupload.TelemetryRule)
	if tmruleDB.ApplicationType != tmrule.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("ApplicationType in db %s doesn't match the ApplicationType %s in req", tmruleDB.ApplicationType, tmrule.ApplicationType), nil)
	}
	respEntity := telemetryRuleValidate(tmrule)
	if respEntity.Error != nil {
		return respEntity
	}

	tmrule.Updated = xwutil.GetTimestamp(time.Now().UTC())
	if err = db.GetCachedSimpleDao().SetOne(db.TABLE_TELEMETRY_RULES, tmrule.ID, tmrule); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, tmrule)
}

func telemetryRuleGeneratePage(list []*xwlogupload.TelemetryRule, page int, pageSize int) (result []*xwlogupload.TelemetryRule) {
	leng := len(list)
	startIndex := page*pageSize - pageSize
	result = make([]*xwlogupload.TelemetryRule, 0)
	if page < 1 || startIndex > leng || pageSize < 1 {
		return result
	}
	lastIndex := leng
	if page*pageSize < len(list) {
		lastIndex = page * pageSize
	}

	return list[startIndex:lastIndex]
}

func TelemetryRuleGeneratePageWithContext(tmrules []*xwlogupload.TelemetryRule, contextMap map[string]string) (result []*xwlogupload.TelemetryRule, err error) {
	sort.Slice(tmrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(tmrules[i].Name), strings.ToLower(tmrules[j].Name)) < 0
	})
	pageNum := 1
	numStr, okval := contextMap[xcommon.PAGE_NUMBER]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[xcommon.PAGE_SIZE]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, errors.New("pageNumber and pageSize should both be greater than zero")
	}
	return telemetryRuleGeneratePage(tmrules, pageNum, pageSize), nil
}

func TelemetryRuleFilterByContext(searchContext map[string]string) []*xwlogupload.TelemetryRule {
	tmRules := xwlogupload.GetTelemetryRuleList()
	tmRuleList := []*xwlogupload.TelemetryRule{}
	for _, tmRule := range tmRules {
		if tmRule == nil {
			continue
		}
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if tmRule.ApplicationType != applicationType && tmRule.ApplicationType != shared.ALL {
				continue
			}
		}

		if name, ok := xutil.FindEntryInContext(searchContext, xcommon.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(tmRule.Name), strings.ToLower(name)) {
				continue
			}
		}
		if key, ok := xutil.FindEntryInContext(searchContext, xcommon.FREE_ARG, false); ok {
			keyMatch := false
			for _, condition := range ru.ToConditions(tmRule.GetRule()) {
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
			for _, condition := range ru.ToConditions(tmRule.GetRule()) {
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
					if strings.Contains(strings.ToLower(condition.FixedArg.Bean.Value.JLString), strings.ToLower(fixedArgValue)) {
						valueMatch = true
						break
					}
				}

			}
			if !valueMatch {
				continue
			}
		}
		if telemetryProfile, ok := xutil.FindEntryInContext(searchContext, xcommon.PROFILE, false); ok {

			telemetry := xwlogupload.GetOnePermanentTelemetryProfile(tmRule.BoundTelemetryID)
			if telemetry != nil && !strings.Contains(strings.ToLower(telemetry.Name), strings.ToLower(telemetryProfile)) {
				continue
			}
		}
		tmRuleList = append(tmRuleList, tmRule)
	}
	return tmRuleList
}
