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
package dcm

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	ru "xconfwebconfig/rulesengine"

	queries "xconfadmin/adminapi/queries"
	xcommon "xconfadmin/common"
	xhttp "xconfadmin/http"
	xutil "xconfadmin/util"
	xwcommon "xconfwebconfig/common"
	ds "xconfwebconfig/db"
	xwhttp "xconfwebconfig/http"
	"xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	"xconfwebconfig/shared/logupload"
	xwutil "xconfwebconfig/util"

	"github.com/google/uuid"
)

const (
	cDcmRulePageNumber = "pageNumber"
	cDcmRulePageSize   = "pageSize"
)

func GetDcmFormulaAll() []*logupload.DCMGenericRule {
	dcmformularules := logupload.GetDCMGenericRuleList()
	return dcmformularules
}

func GetDcmFormula(id string) *logupload.DCMGenericRule {
	dcmformula := logupload.GetOneDCMGenericRule(id)
	if dcmformula != nil {
		return dcmformula
	}
	return nil

}

func validateUsageForDcmFormula(id string, appType string) (string, error) {
	dcmformula := GetDcmFormula(id)
	if dcmformula == nil || dcmformula.ApplicationType != appType {
		return fmt.Sprintf("Entity with id  %s does not exist ", id), nil
	}
	return "", nil
}

func DeleteDcmFormulabyId(id string, appType string) *xwhttp.ResponseEntity {
	usage, err := validateUsageForDcmFormula(id, appType)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, err, nil)
	}

	if usage != "" {
		return xwhttp.NewResponseEntity(http.StatusNotFound, errors.New(usage), nil)
	}

	err = DeleteOneDcmFormula(id, appType)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func DeleteOneDcmFormula(id string, appType string) error {
	err := ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_DCM_RULE, id)
	if err != nil {
		return err
	}
	devicesettings := logupload.GetOneDeviceSettings(id)
	if devicesettings != nil {
		err := ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_DEVICE_SETTINGS, id)
		if err != nil {
			return err
		}
	}
	loguploadsettings := logupload.GetOneLogUploadSettings(id)
	if loguploadsettings != nil {
		err := ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_LOG_UPLOAD_SETTINGS, id)
		if err != nil {
			return err
		}
	}
	vodsettings := logupload.GetOneVodSettings(id)
	if vodsettings != nil {
		err := ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_VOD_SETTINGS, id)
		if err != nil {
			return err
		}
	}

	return packPriorities(appType)
}

func packPriorities(appType string) error {
	changedRules := []*logupload.DCMGenericRule{}
	dfrules := GetDcmRulesByApplicationType(appType)
	// sort by ascending priority
	sort.Slice(dfrules, func(i, j int) bool {
		return dfrules[i].Priority < dfrules[j].Priority
	})
	priority := 1
	for _, item := range dfrules {
		if item.Priority != priority {
			item.Priority = priority
			changedRules = append(changedRules, item)
		}
		priority++
	}
	// Now save all updated priorities
	for _, dcmrule := range changedRules {
		if err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, dcmrule.ID, dcmrule); err != nil {
			return err
		}
	}
	return nil
}

func dcmRuleValidate(dfrule *logupload.DCMGenericRule) *xwhttp.ResponseEntity {
	if dfrule == nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("DCM formula Rule should be specified"), nil)
	}

	if xwutil.IsBlank(dfrule.ID) {
		dfrule.ID = uuid.New().String()
	}
	if xwutil.IsBlank(dfrule.ApplicationType) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ApplicationType is empty"), nil)
	}

	if xwutil.IsBlank(dfrule.Name) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Name is empty"), nil)
	}

	if dfrule.GetRule() != nil {
		ru.NormalizeConditions(dfrule.GetRule())
	} else {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Condition is empty"), nil)
	}
	err := queries.ValidateRuleStructure(dfrule.GetRule())
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}
	err = queries.RunGlobalValidation(*dfrule.GetRule(), queries.GetFirmwareRuleAllowedOperations)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}
	err = validatePercentage(dfrule)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}
	dfrules := GetDcmFormulaAll()
	for _, exdfrule := range dfrules {
		if exdfrule.ApplicationType != dfrule.ApplicationType {
			continue
		}
		if exdfrule.ID != dfrule.ID {
			if exdfrule.Name == dfrule.Name {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Formula Name is already used"), nil)
			}
			rule1 := exdfrule.GetRule()
			rule2 := dfrule.GetRule()
			if rule1.Equals(rule2) {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Rule has duplicate: %s", exdfrule.Name), nil)

			}
		}
	}
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, nil)
}

func validatePercentage(dfrule *logupload.DCMGenericRule) error {
	p := dfrule.Percentage
	p1 := dfrule.PercentageL1
	p2 := dfrule.PercentageL2
	p3 := dfrule.PercentageL3
	if int(p1) < 0 || int(p2) < 0 || int(p3) < 0 {
		err := fmt.Errorf("Percentage must be in range from 0 to 100")
		return err
	}
	psend := int(p1) + int(p2) + int(p3)

	if psend < 0 || psend > 100 || p < 0 || p > 100 {
		err := fmt.Errorf("Total Level percentage sum must be in range from 0 to 100")
		return err
	}
	return nil
}

func getAlteredSubList(itemsList []*logupload.DCMGenericRule, oldPriority int, newPriority int) []*logupload.DCMGenericRule {
	start := int(math.Min(float64(oldPriority), float64(newPriority))) - 1
	end := int(math.Max(float64(oldPriority), float64(newPriority)))
	return itemsList[start:end]
}

func reorganizePriorities(dfrules []*logupload.DCMGenericRule, oldpriority int, newpriority int) []*logupload.DCMGenericRule {
	if newpriority < 1 || newpriority > len(dfrules) {
		newpriority = len(dfrules)
	}
	dfrule := dfrules[oldpriority-1]
	dfrule.Priority = newpriority

	if oldpriority < newpriority {
		for i := oldpriority; i <= newpriority-1; i++ {
			elem := dfrules[i]
			elem.Priority = i
			dfrules[i-1] = elem
		}
	}

	if oldpriority > newpriority {
		for i := oldpriority - 2; i >= newpriority-1; i-- {
			elem := dfrules[i]
			elem.Priority = i + 2
			dfrules[i+1] = elem
		}
	}

	dfrules[newpriority-1] = dfrule

	return getAlteredSubList(dfrules, oldpriority, newpriority)
}

func AddnewItemAndRepriortize(newdfrule *logupload.DCMGenericRule) []*logupload.DCMGenericRule {
	dfrules := GetDcmRulesByApplicationType(newdfrule.ApplicationType)
	// sort by ascending priority
	sort.Slice(dfrules, func(i, j int) bool {
		return dfrules[i].Priority < dfrules[j].Priority
	})
	dfrules = append(dfrules, newdfrule)
	oldpriority := len(dfrules)
	newpriority := newdfrule.Priority
	return reorganizePriorities(dfrules, oldpriority, newpriority)
}

func CreateDcmRule(dfrule *logupload.DCMGenericRule, appType string) *xwhttp.ResponseEntity {
	if existingRule := logupload.GetOneDCMGenericRule(dfrule.ID); existingRule != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s already exists", dfrule.ID), nil)
	}
	if dfrule.ApplicationType != appType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", dfrule.ID), nil)
	}
	if respEntity := dcmRuleValidate(dfrule); respEntity.Error != nil {
		return respEntity
	}

	list := AddnewItemAndRepriortize(dfrule)
	for _, entry := range list {
		entry.Updated = xwutil.GetTimestamp(time.Now().UTC())
		if err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, entry.ID, entry); err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
	}
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, dfrule)
}

func GetDcmRulesByApplicationType(applicationType string) []*logupload.DCMGenericRule {
	list := []*logupload.DCMGenericRule{}
	result := GetDcmFormulaAll()
	for _, DcmRule := range result {
		if DcmRule.ApplicationType == applicationType {
			list = append(list, DcmRule)
		}
	}
	return list
}

func UpdateItemAndRepriortize(newdfrule *logupload.DCMGenericRule, oldPriority, newPriority int) (result []*logupload.DCMGenericRule, err error) {
	dfrules := GetDcmRulesByApplicationType(newdfrule.ApplicationType)
	// sort by ascending priority
	sort.Slice(dfrules, func(i, j int) bool {
		return dfrules[i].Priority < dfrules[j].Priority
	})

	if newPriority < 1 || newPriority > len(dfrules)+1 {
		return nil, errors.New("Invalid value for Priority")
	}
	result = reorganizePriorities(dfrules, oldPriority, newPriority)

	if len(result) > (newPriority - 1) {
		result[newPriority-1] = newdfrule
	} else if len(result) == 1 && oldPriority == newPriority {
		result[0] = newdfrule
	}

	return result, nil
}

func UpdateDcmRule(dfrule *logupload.DCMGenericRule, appType string) *xwhttp.ResponseEntity {
	if xwutil.IsBlank(dfrule.ID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ID is empty"), nil)
	}
	if dfrule.ApplicationType != appType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", dfrule.ID), nil)
	}
	existingRule := logupload.GetOneDCMGenericRule(dfrule.ID)
	if existingRule == nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s does not exist", dfrule.ID), nil)
	}
	if existingRule.ApplicationType != dfrule.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("ApplicationType in db %s doesn't match the ApplicationType %s in req", existingRule.ApplicationType, dfrule.ApplicationType), nil)
	}
	respEntity := dcmRuleValidate(dfrule)
	if respEntity.Error != nil {
		return respEntity
	}

	if dfrule.Priority == existingRule.Priority {
		dfrule.Updated = xwutil.GetTimestamp(time.Now().UTC())
		if err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, dfrule.ID, dfrule); err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
	} else {
		list, err := UpdateItemAndRepriortize(dfrule, existingRule.Priority, dfrule.Priority)
		if err != nil {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
		}
		for _, entry := range list {
			entry.Updated = xwutil.GetTimestamp(time.Now().UTC())
			if err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, entry.ID, entry); err != nil {
				return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
			}
		}
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, dfrule)
}

func dcmRuleGeneratePage(list []*logupload.DCMGenericRule, page int, pageSize int) (result []*logupload.DCMGenericRule) {
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

func DcmFormulaRuleGeneratePageWithContext(dfrules []*logupload.DCMGenericRule, contextMap map[string]string) (result []*logupload.DCMGenericRule, err error) {
	sort.Slice(dfrules, func(i, j int) bool {
		return dfrules[i].Priority < dfrules[j].Priority
	})
	pageNum := 1
	numStr, okval := contextMap[cDcmRulePageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cDcmRulePageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, errors.New("pageNumber and pageSize should both be greater than zero")
	}
	return dcmRuleGeneratePage(dfrules, pageNum, pageSize), nil
}

func DcmFormulaFilterByContext(searchContext map[string]string) []*logupload.DCMGenericRule {
	dcmFormulaRules := logupload.GetDCMGenericRuleList()
	dcmFormulaRuleList := []*logupload.DCMGenericRule{}
	for _, dcmRule := range dcmFormulaRules {
		if dcmRule == nil {
			continue
		}
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if dcmRule.ApplicationType != applicationType && dcmRule.ApplicationType != shared.ALL {
				continue
			}
		}
		if name, ok := xutil.FindEntryInContext(searchContext, xcommon.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(dcmRule.Name), strings.ToLower(name)) {
				continue
			}
		}
		if key, ok := xutil.FindEntryInContext(searchContext, xcommon.FREE_ARG, false); ok {
			keyMatch := false
			for _, condition := range ru.ToConditions(dcmRule.GetRule()) {
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
			for _, condition := range ru.ToConditions(dcmRule.GetRule()) {
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
		dcmFormulaRuleList = append(dcmFormulaRuleList, dcmRule)
	}
	return dcmFormulaRuleList
}

func importFormula(formulaWithSettings *logupload.FormulaWithSettings, overwrite bool, appType string) *xwhttp.ResponseEntity {
	formula := formulaWithSettings.Formula
	deviceSettings := formulaWithSettings.DeviceSettings
	logUploadSettings := formulaWithSettings.LogUpLoadSettings
	vodSettings := formulaWithSettings.VodSettings

	if xwutil.IsBlank(formula.ApplicationType) {
		formula.ApplicationType = appType
	}
	if deviceSettings != nil {
		if xwutil.IsBlank(deviceSettings.ApplicationType) {
			deviceSettings.ApplicationType = appType
		}
		if formula.ApplicationType != deviceSettings.ApplicationType {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("DeviceSettings ApplicationType mismatch"), nil)
		}
		if xwutil.IsBlank(deviceSettings.Schedule.TimeZone) {
			if logUploadSettings != nil {
				logUploadSettings.Schedule.TimeZone = logupload.UTC
			}
		}
		if respEntity := DeviceSettingsValidate(deviceSettings); respEntity.Error != nil {
			return respEntity
		}
	}
	if logUploadSettings != nil {
		if xwutil.IsBlank(logUploadSettings.ApplicationType) {
			logUploadSettings.ApplicationType = appType
		}
		if formula.ApplicationType != logUploadSettings.ApplicationType {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("logUploadSettings ApplicationType mismatch"), nil)
		}
		if xwutil.IsBlank(logUploadSettings.Schedule.TimeZone) {
			logUploadSettings.Schedule.TimeZone = logupload.UTC
		}
		if respEntity := LogUploadSettingsValidate(logUploadSettings); respEntity.Error != nil {
			return respEntity
		}
	}
	if vodSettings != nil {
		if xwutil.IsBlank(vodSettings.ApplicationType) {
			vodSettings.ApplicationType = appType
		}
		if formula.ApplicationType != vodSettings.ApplicationType {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("vodSettings ApplicationType mismatch"), nil)
		}
		if respEntity := VodSettingsValidate(vodSettings); respEntity.Error != nil {
			return respEntity
		}
	}

	if overwrite {
		if respEntity := UpdateDcmRule(formula, appType); respEntity.Error != nil {
			return respEntity
		}
		if deviceSettings != nil {
			if respEntity := UpdateDeviceSettings(deviceSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
		if logUploadSettings != nil {
			if respEntity := UpdateLogUploadSettings(logUploadSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
		if vodSettings != nil {
			if respEntity := UpdateVodSettings(vodSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
	} else {
		if respEntity := CreateDcmRule(formula, appType); respEntity.Error != nil {
			return respEntity
		}
		if deviceSettings != nil {
			if respEntity := CreateDeviceSettings(deviceSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
		if logUploadSettings != nil {
			if respEntity := CreateLogUploadSettings(logUploadSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
		if vodSettings != nil {
			if respEntity := CreateVodSettings(vodSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, formulaWithSettings)
}

func importFormulas(formulaWithSettingsList []*logupload.FormulaWithSettings, appType string, overwrite bool) map[string]xhttp.EntityMessage {
	entitiesMap := map[string]xhttp.EntityMessage{}

	sort.Slice(formulaWithSettingsList, func(i, j int) bool {
		return formulaWithSettingsList[i].Formula.Priority < formulaWithSettingsList[j].Formula.Priority
	})

	for _, formulaWithSettings := range formulaWithSettingsList {
		formula := formulaWithSettings.Formula
		respEntity := importFormula(formulaWithSettings, overwrite, appType)
		if respEntity.Error != nil {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: respEntity.Error.Error(),
			}
			entitiesMap[formula.ID] = entityMessage
		} else {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: formula.ID,
			}
			entitiesMap[formula.ID] = entityMessage
		}
	}

	return entitiesMap
}
