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
package dcm

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/rdkcentral/xconfadmin/adminapi/queries"
	"github.com/rdkcentral/xconfadmin/common"
	xcommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	core "github.com/rdkcentral/xconfadmin/shared"
	"github.com/rdkcentral/xconfadmin/util"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

const (
	cDcmRulePageNumber = "pageNumber"
	cDcmRulePageSize   = "pageSize"
)

var dcmRuleTableMutex sync.Mutex
var dcmRuleTableLock = db.NewDistributedLock(db.TABLE_DCM_RULES, 10)

func GetDcmFormulaAll(tenantId string) []*logupload.DCMGenericRule {
	dcmformularules := logupload.GetDCMGenericRuleListForAS(tenantId)
	return dcmformularules
}

func GetDcmFormula(tenantId string, id string) *logupload.DCMGenericRule {
	dcmformula := logupload.GetOneDCMGenericRule(tenantId, id)
	if dcmformula != nil {
		return dcmformula
	}
	return nil

}

func validateIfExists(tenantId string, id string, appType string) error {
	existingFormula := GetDcmFormula(tenantId, id)
	if existingFormula == nil || existingFormula.ApplicationType != appType {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id "+id+" does not exist ")
	}
	return nil
}

func DeleteDcmFormulabyId(tenantId string, id string, appType string) *xcommon.ResponseEntity {
	err := validateIfExists(tenantId, id, appType)
	if err != nil {
		return xcommon.NewResponseEntityWithStatus(http.StatusNotFound, err, nil)
	}

	err = DeleteOneDcmFormula(tenantId, id, appType)
	if err != nil {
		return xcommon.NewResponseEntityWithStatus(http.StatusInternalServerError, err, nil)
	}

	return xcommon.NewResponseEntityWithStatus(http.StatusNoContent, nil, nil)
}

func SaveDcmRules(tenantId string, itemList []core.Prioritizable) error {
	for _, item := range itemList {
		rule := item.(*logupload.DCMGenericRule)
		if err := db.GetCachedSimpleDao().SetOne(tenantId, db.TABLE_DCM_RULES, rule.GetID(), rule); err != nil {
			return err
		}
	}
	return nil
}

func DeleteOneDcmFormula(tenantId string, id string, appType string) error {
	existingRule := logupload.GetOneDCMGenericRule(tenantId, id)
	if existingRule == nil {
		return fmt.Errorf("Entity with id %s does not exist", id)
	}
	err := db.GetCachedSimpleDao().DeleteOne(tenantId, db.TABLE_DCM_RULES, id)
	if err != nil {
		return err
	}
	devicesettings := logupload.GetOneDeviceSettings(tenantId, id)
	if devicesettings != nil {
		err := db.GetCachedSimpleDao().DeleteOne(tenantId, db.TABLE_DEVICE_SETTINGS, id)
		if err != nil {
			return err
		}
	}
	loguploadsettings := logupload.GetOneLogUploadSettings(tenantId, id)
	if loguploadsettings != nil {
		err := db.GetCachedSimpleDao().DeleteOne(tenantId, db.TABLE_LOG_UPLOAD_SETTINGS, id)
		if err != nil {
			return err
		}
	}
	vodsettings := logupload.GetOneVodSettings(tenantId, id)
	if vodsettings != nil {
		err := db.GetCachedSimpleDao().DeleteOne(tenantId, db.TABLE_VOD_SETTINGS, id)
		if err != nil {
			return err
		}
	}

	dcmRulesByAppType := GetDcmRulesByApplicationType(tenantId, appType)
	prioritizableRules := DcmRulesToPrioritizables(dcmRulesByAppType)
	err = SaveDcmRules(tenantId, queries.PackPriorities(prioritizableRules, existingRule))
	if err != nil {
		return err
	}
	return nil
}

func dcmRuleValidate(tenantId string, dfrule *logupload.DCMGenericRule) *xwhttp.ResponseEntity {
	if dfrule == nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("DCM formula Rule should be specified"), nil)
	}

	if util.IsBlank(dfrule.ID) {
		dfrule.ID = uuid.New().String()
	}
	if util.IsBlank(dfrule.ApplicationType) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ApplicationType is empty"), nil)
	}

	if util.IsBlank(dfrule.Name) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Name is empty"), nil)
	}

	if dfrule.GetRule() != nil {
		rulesengine.NormalizeConditions(dfrule.GetRule())
	} else {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Condition is empty"), nil)
	}
	err := queries.ValidateRuleStructure(dfrule.GetRule())
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}
	err = queries.RunGlobalValidation(tenantId, *dfrule.GetRule(), queries.GetFirmwareRuleAllowedOperations)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}
	err = validatePercentage(dfrule)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}
	dfrules := GetDcmFormulaAll(tenantId)
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
	p1, _ := dfrule.PercentageL1.Int64()
	p2, _ := dfrule.PercentageL2.Int64()
	p3, _ := dfrule.PercentageL3.Int64()
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

func DcmRulesToPrioritizables(dcmRules []*logupload.DCMGenericRule) []core.Prioritizable {
	prioritizables := make([]core.Prioritizable, len(dcmRules))
	for i, item := range dcmRules {
		itemCopy := *item
		prioritizables[i] = &itemCopy
	}
	return prioritizables
}

func CreateDcmRule(tenantId string, dfrule *logupload.DCMGenericRule, appType string) *xwhttp.ResponseEntity {
	if existingRule := logupload.GetOneDCMGenericRule(tenantId, dfrule.ID); existingRule != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s already exists", dfrule.ID), nil)
	}
	if dfrule.ApplicationType != appType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", dfrule.ID), nil)
	}
	if respEntity := dcmRuleValidate(tenantId, dfrule); respEntity.Error != nil {
		return respEntity
	}

	dcmRulesByAppType := GetDcmRulesByApplicationType(tenantId, dfrule.ApplicationType)
	changedDcmRules := queries.AddNewPrioritizableAndReorganizePriorities(dfrule, DcmRulesToPrioritizables(dcmRulesByAppType))
	for _, entry := range changedDcmRules {
		entry.(*logupload.DCMGenericRule).Updated = util.GetTimestamp()
		if err := db.GetCachedSimpleDao().SetOne(tenantId, db.TABLE_DCM_RULES, entry.GetID(), entry); err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
	}
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, dfrule)
}

func GetDcmRulesByApplicationType(tenantId string, applicationType string) []*logupload.DCMGenericRule {
	list := []*logupload.DCMGenericRule{}
	result := GetDcmFormulaAll(tenantId)
	for _, DcmRule := range result {
		if DcmRule.ApplicationType == applicationType {
			list = append(list, DcmRule)
		}
	}
	return list
}

func UpdateDcmRule(tenantId string, incomingFormula *logupload.DCMGenericRule, appType string) *xwhttp.ResponseEntity {
	if util.IsBlank(incomingFormula.ID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ID is empty"), nil)
	}
	if incomingFormula.ApplicationType != appType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", incomingFormula.ID), nil)
	}

	existingFormula := logupload.GetOneDCMGenericRule(tenantId, incomingFormula.ID)
	if existingFormula == nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s does not exist", incomingFormula.ID), nil)
	}
	if existingFormula.ApplicationType != incomingFormula.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("ApplicationType in db %s doesn't match the ApplicationType %s in req", existingFormula.ApplicationType, incomingFormula.ApplicationType), nil)
	}
	respEntity := dcmRuleValidate(tenantId, incomingFormula)
	if respEntity.Error != nil {
		return respEntity
	}

	if incomingFormula.Priority == existingFormula.Priority {
		incomingFormula.Updated = util.GetTimestamp()
		if err := db.GetCachedSimpleDao().SetOne(tenantId, db.TABLE_DCM_RULES, incomingFormula.ID, incomingFormula); err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
	} else {
		formulasByApplicationType := GetDcmRulesByApplicationType(tenantId, incomingFormula.ApplicationType)
		changedFormulae := queries.UpdatePrioritizablePriorityAndReorganize(incomingFormula, DcmRulesToPrioritizables(formulasByApplicationType), existingFormula.Priority)
		for _, entry := range changedFormulae {
			entry.(*logupload.DCMGenericRule).Updated = util.GetTimestamp()
			if err := db.GetCachedSimpleDao().SetOne(tenantId, db.TABLE_DCM_RULES, entry.GetID(), entry); err != nil {
				return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
			}
		}
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, incomingFormula)
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
	dcmFormulaRules := logupload.GetDCMGenericRuleListForAS(searchContext[xwcommon.TENANT_ID])
	dcmFormulaRuleList := []*logupload.DCMGenericRule{}
	for _, dcmRule := range dcmFormulaRules {
		if dcmRule == nil {
			continue
		}
		if applicationType, ok := util.FindEntryInContext(searchContext, core.APPLICATION_TYPE, false); ok {
			if dcmRule.ApplicationType != applicationType && dcmRule.ApplicationType != core.ALL {
				continue
			}
		}
		if name, ok := util.FindEntryInContext(searchContext, xcommon.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(dcmRule.Name), strings.ToLower(name)) {
				continue
			}
		}
		if key, ok := util.FindEntryInContext(searchContext, xcommon.FREE_ARG, false); ok {
			if !re.IsExistConditionByFreeArgName(*dcmRule.GetRule(), key) {
				continue
			}
		}
		if fixedArgValue, ok := util.FindEntryInContext(searchContext, xcommon.FIXED_ARG, false); ok {
			if !re.IsExistConditionByFixedArgValue(*dcmRule.GetRule(), fixedArgValue) {
				continue
			}
		}
		dcmFormulaRuleList = append(dcmFormulaRuleList, dcmRule)
	}
	return dcmFormulaRuleList
}

func importFormula(tenantId string, formulaWithSettings *logupload.FormulaWithSettings, overwrite bool, appType string) *xwhttp.ResponseEntity {
	formula := formulaWithSettings.Formula
	deviceSettings := formulaWithSettings.DeviceSettings
	logUploadSettings := formulaWithSettings.LogUpLoadSettings
	vodSettings := formulaWithSettings.VodSettings

	if util.IsBlank(formula.ApplicationType) {
		formula.ApplicationType = appType
	}
	if deviceSettings != nil {
		if util.IsBlank(deviceSettings.ApplicationType) {
			deviceSettings.ApplicationType = appType
		}
		if formula.ApplicationType != deviceSettings.ApplicationType {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("DeviceSettings ApplicationType mismatch"), nil)
		}
		if util.IsBlank(deviceSettings.Schedule.TimeZone) {
			if logUploadSettings != nil {
				logUploadSettings.Schedule.TimeZone = logupload.UTC
			}
		}
		if respEntity := DeviceSettingsValidate(tenantId, deviceSettings); respEntity.Error != nil {
			return respEntity
		}
	}
	if logUploadSettings != nil {
		if util.IsBlank(logUploadSettings.ApplicationType) {
			logUploadSettings.ApplicationType = appType
		}
		if formula.ApplicationType != logUploadSettings.ApplicationType {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("logUploadSettings ApplicationType mismatch"), nil)
		}
		if util.IsBlank(logUploadSettings.Schedule.TimeZone) {
			logUploadSettings.Schedule.TimeZone = logupload.UTC
		}
		if respEntity := LogUploadSettingsValidate(tenantId, logUploadSettings); respEntity.Error != nil {
			return respEntity
		}
	}
	if vodSettings != nil {
		if util.IsBlank(vodSettings.ApplicationType) {
			vodSettings.ApplicationType = appType
		}
		if formula.ApplicationType != vodSettings.ApplicationType {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("vodSettings ApplicationType mismatch"), nil)
		}
		if respEntity := VodSettingsValidate(tenantId, vodSettings); respEntity.Error != nil {
			return respEntity
		}
	}

	if overwrite {
		if respEntity := UpdateDcmRule(tenantId, formula, appType); respEntity.Error != nil {
			return respEntity
		}
		if deviceSettings != nil {
			if respEntity := UpdateDeviceSettings(tenantId, deviceSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
		if logUploadSettings != nil {
			if respEntity := UpdateLogUploadSettings(tenantId, logUploadSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
		if vodSettings != nil {
			if respEntity := UpdateVodSettings(tenantId, vodSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
	} else {
		if respEntity := CreateDcmRule(tenantId, formula, appType); respEntity.Error != nil {
			return respEntity
		}
		if deviceSettings != nil {
			if respEntity := CreateDeviceSettings(tenantId, deviceSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
		if logUploadSettings != nil {
			if respEntity := CreateLogUploadSettings(tenantId, logUploadSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
		if vodSettings != nil {
			if respEntity := CreateVodSettings(tenantId, vodSettings, appType); respEntity.Error != nil {
				return respEntity
			}
		}
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, formulaWithSettings)
}

func importFormulas(tenantId string, formulaWithSettingsList []*logupload.FormulaWithSettings, appType string, overwrite bool) map[string]xhttp.EntityMessage {
	entitiesMap := map[string]xhttp.EntityMessage{}

	sort.Slice(formulaWithSettingsList, func(i, j int) bool {
		return formulaWithSettingsList[i].Formula.Priority < formulaWithSettingsList[j].Formula.Priority
	})

	for _, formulaWithSettings := range formulaWithSettingsList {
		formula := formulaWithSettings.Formula
		respEntity := importFormula(tenantId, formulaWithSettings, overwrite, appType)
		if respEntity.Error != nil {
			entityMessage := xhttp.EntityMessage{
				Status:  common.ENTITY_STATUS_FAILURE,
				Message: respEntity.Error.Error(),
			}
			entitiesMap[formula.ID] = entityMessage
		} else {
			entityMessage := xhttp.EntityMessage{
				Status:  common.ENTITY_STATUS_SUCCESS,
				Message: formula.ID,
			}
			entitiesMap[formula.ID] = entityMessage
		}
	}

	return entitiesMap
}
