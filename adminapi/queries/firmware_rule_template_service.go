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
package queries

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"net/http"

	"github.com/rdkcentral/xconfwebconfig/common"
	ruleutil "github.com/rdkcentral/xconfwebconfig/rulesengine"
	xutil "github.com/rdkcentral/xconfwebconfig/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	xcommon "github.com/rdkcentral/xconfadmin/common"
	core "github.com/rdkcentral/xconfadmin/shared"
	xcorefw "github.com/rdkcentral/xconfadmin/shared/firmware"
	"github.com/rdkcentral/xconfadmin/util"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var firmwareRuleTemplateUpdateMutex sync.Mutex

const (
	cFirmwareRTName                 = xcommon.NAME
	cFirmwareRTKey                  = corefw.KEY
	cFirmwareRTValue                = corefw.VALUE
	cFirmwareRTApplicableActionType = xcommon.APPLICABLE_ACTION_TYPE
	cFirmwareRTPageNumber           = xcommon.PAGE_NUMBER
	cFirmwareRTPageSize             = xcommon.PAGE_SIZE
	cFirmwareRT                     = corefw.RULE_TEMPLATE
	cFirmwareRTBlockingFilter       = corefw.BLOCKING_FILTER_TEMPLATE
	cFirmwareRTDefineProperties     = corefw.DEFINE_PROPERTIES_TEMPLATE
)

func honoredByFirmwareRT(context map[string]string, firmwareRT *corefw.FirmwareRuleTemplate) bool {
	// Objects.nonNull(xRule) && StringUtils.containsIgnoreCase(xRule.getName(), name);
	name, filterByName := util.FindEntryInContext(context, cFirmwareRTName, false)
	if filterByName {
		baseName := strings.ToLower(firmwareRT.GetName())
		givenName := strings.ToLower(name)
		if !strings.Contains(baseName, givenName) {
			return false
		}
	}

	// xRule -> Objects.nonNull(xRule) && RuleUtil.isExistConditionByFreeArgName(xRule.getRule(), key);
	key, filterByKey := util.FindEntryInContext(context, cFirmwareRTKey, false)
	if filterByKey && !re.IsExistConditionByFreeArgName(firmwareRT.Rule, key) {
		return false
	}

	//xRule -> Objects.nonNull(xRule) && RuleUtil.isExistConditionByFixedArgValue(xRule.getRule(), value);
	val, filterByVal := util.FindEntryInContext(context, cFirmwareRTValue, false)
	if filterByVal && !re.IsExistConditionByFixedArgValue(firmwareRT.Rule, val) {
		return false
	}
	return true
}

func filterFirmwareRTsByContext(dbrules []*corefw.FirmwareRuleTemplate, firmwareRTContext map[string]string) (filteredRTs map[string][]*corefw.FirmwareRuleTemplate) {
	filteredRTs = make(map[string][]*corefw.FirmwareRuleTemplate)
	for _, firmwareRT := range dbrules {
		if honoredByFirmwareRT(firmwareRTContext, firmwareRT) {
			filteredRTs[string(firmwareRT.ApplicableAction.ActionType)] = append(filteredRTs[string(firmwareRT.ApplicableAction.ActionType)], firmwareRT)
		}
	}
	return filteredRTs
}

func putSizesOfFirmwareRTsByTypeIntoHeaders2(dbrules []*corefw.FirmwareRuleTemplate) (headers map[string]string) {
	ruleCnt := 0
	blkFilterCnt := 0
	defPropCnt := 0

	for _, firmwareRT := range dbrules {
		if firmwareRT.ApplicableAction.ActionType.CaseIgnoreEquals(cFirmwareRT) {
			ruleCnt++
		} else if firmwareRT.ApplicableAction.ActionType.CaseIgnoreEquals(cFirmwareRTBlockingFilter) {
			blkFilterCnt++
		} else if firmwareRT.ApplicableAction.ActionType.CaseIgnoreEquals(cFirmwareRTDefineProperties) {
			defPropCnt++
		}
	}
	headers = map[string]string{
		string(cFirmwareRT):                 strconv.Itoa(ruleCnt),
		string(cFirmwareRTBlockingFilter):   strconv.Itoa(blkFilterCnt),
		string(cFirmwareRTDefineProperties): strconv.Itoa(defPropCnt),
	}
	return headers
}

func putSizesOfFirmwareRTsByTypeIntoHeaders(headers map[string]string, dbRulesMap map[string][]*corefw.FirmwareRuleTemplate) {
	headers[string(cFirmwareRT)] = strconv.Itoa(len(dbRulesMap[string(cFirmwareRT)]))
	headers[string(cFirmwareRTBlockingFilter)] = strconv.Itoa(len(dbRulesMap[string(cFirmwareRTBlockingFilter)]))
	headers[string(cFirmwareRTDefineProperties)] = strconv.Itoa(len(dbRulesMap[string(cFirmwareRTDefineProperties)]))
}

func extractFirmwareRTPage(list []*corefw.FirmwareRuleTemplate, page int, pageSize int) (result []*corefw.FirmwareRuleTemplate) {
	leng := len(list)
	result = make([]*corefw.FirmwareRuleTemplate, 0)
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

func generateFirmwareRTPageByContext(dbrules []*corefw.FirmwareRuleTemplate, contextMap map[string]string) (result []*corefw.FirmwareRuleTemplate, err error) {
	pageNum := 1
	numStr, okval := contextMap[cFirmwareRTPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cFirmwareRTPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "pageNumber and pageSize should both be greater than zero")
	}
	return extractFirmwareRTPage(dbrules, pageNum, pageSize), nil
}

func firmwareRTFilterByActionType(dbrules []*corefw.FirmwareRuleTemplate, actionType string) (result []*corefw.FirmwareRuleTemplate) {
	filteredRules := make([]*corefw.FirmwareRuleTemplate, 0)
	for _, firmwareRT := range dbrules {
		baseName := strings.ToLower(string(firmwareRT.ApplicableAction.ActionType))
		givenName := strings.ToLower(actionType)
		if strings.Contains(baseName, givenName) {
			filteredRules = append(filteredRules, firmwareRT)
		}
	}
	return filteredRules
}

func validateProperties(applicableAction *corefw.TemplateApplicableAction) error {
	if applicableAction.ActionType == corefw.DEFINE_PROPERTIES_TEMPLATE {
		for k := range applicableAction.Properties {
			if xutil.IsBlank(k) {
				return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "properties key is blank")
			}
		}
	}
	return nil
}

func validateRule(fr *re.Rule, action *corefw.TemplateApplicableAction) error {
	if err := ValidateRuleStructure(fr); err != nil {
		return err
	}
	if err := validateRelation(fr); err != nil {
		return err
	}
	if err := checkDuplicateConditions(fr); err != nil {
		return err
	}
	conditions := ruleutil.ToConditions(fr)
	if len(conditions) == 0 {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "FirmwareRuleTemplate "+fr.Id()+" should have a minimum one condition")
	}
	for _, c := range conditions {
		if err := checkOperationName(c, GetFirmwareRuleAllowedOperations); err != nil {
			return err
		}
	}
	return validateProperties(action)
}

func validateOneFirmwareRT(frt corefw.FirmwareRuleTemplate) error {
	if frt.ApplicableAction == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Missing applicable action type ")
	}
	validActionTypes := []corefw.ApplicableActionType{
		cFirmwareRT, cFirmwareRTBlockingFilter, cFirmwareRTDefineProperties,
	}
	found := false
	for _, elem := range validActionTypes {
		if elem == frt.ApplicableAction.ActionType {
			found = true
			break
		}
	}
	if !found {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Invalid action type "+string(frt.ApplicableAction.ActionType)+" in "+frt.GetName())
	}
	return validateRule(frt.GetRule(), frt.ApplicableAction)
}

func validateAgainstFirmwareRTs(frt *corefw.FirmwareRuleTemplate, entities []*corefw.FirmwareRuleTemplate) error {
	for _, rule := range entities {
		if rule.ID == frt.ID {
			continue
		}
		if frt.GetName() == rule.GetName() {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, rule.GetName()+" is already used")
		}
		if ruleutil.EqualComplexRules(frt.GetRule(), rule.GetRule()) {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Rule is duplicate of "+rule.ID)
		}
	}
	return nil
}

func getAlteredFirmwareRTSubList(itemsList []*corefw.FirmwareRuleTemplate, oldPriority int, newPriority int) []*corefw.FirmwareRuleTemplate {
	start := int(math.Min(float64(oldPriority), float64(newPriority)) - float64(1))
	end := int(math.Max(float64(oldPriority), float64(newPriority)))
	return itemsList[start:end]
}

func reorganizeFirmwareRTPriorities(sortedItemsList []*corefw.FirmwareRuleTemplate, oldPriority int, newPriority int) []*corefw.FirmwareRuleTemplate {
	if newPriority < 1 || int(newPriority) > len(sortedItemsList) {
		newPriority = len(sortedItemsList)
	}
	item := sortedItemsList[oldPriority-1]
	item.Priority = int32(newPriority)

	if oldPriority < newPriority {
		for i := oldPriority; i <= newPriority-1; i++ {
			buf := sortedItemsList[i]
			buf.Priority = int32(i)
			sortedItemsList[i-1] = buf
		}
	}

	if oldPriority > newPriority {
		for i := oldPriority - 2; i >= newPriority-1; i-- {
			buf := sortedItemsList[i]
			buf.Priority = int32(i + 2)
			sortedItemsList[i+1] = buf
		}
	}

	sortedItemsList[newPriority-1] = item

	return getAlteredFirmwareRTSubList(sortedItemsList, oldPriority, newPriority)
}

func updateFirmwareRTByPriorityAndReorganize(itemToSave *corefw.FirmwareRuleTemplate, itemsList []*corefw.FirmwareRuleTemplate, newPriority int) (result []*corefw.FirmwareRuleTemplate, err error) {
	sort.Slice(itemsList, func(i, j int) bool {
		return itemsList[i].Priority < itemsList[j].Priority
	})
	oldPriority := len(itemsList)
	if len(itemsList) > 0 {
		for i, item := range itemsList {
			if item.ID == itemToSave.ID {
				oldPriority = int(itemsList[i].Priority)
				itemsList[i] = itemToSave
				break
			}
		}
	} else {
		itemsList = append(itemsList, itemToSave)
	}
	result = reorganizeFirmwareRTPriorities(itemsList, oldPriority, newPriority)
	return result, nil
}

func addNewFirmwareRTAndReorganize(newItem corefw.FirmwareRuleTemplate, itemsList []*corefw.FirmwareRuleTemplate) []*corefw.FirmwareRuleTemplate {
	sort.Slice(itemsList, func(i, j int) bool {
		return itemsList[i].Priority < itemsList[j].Priority
	})
	itemsList = append(itemsList, &newItem)
	return reorganizeFirmwareRTPriorities(itemsList, len(itemsList), int(newItem.Priority))
}

// func saveAllFirmwareRTs(templateList []*corefw.FirmwareRuleTemplate) error {
// 	for _, template := range templateList {
// 		template.Updated = xutil.GetTimestamp(time.Now().UTC())
// 		if err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_FIRMWARE_RULE_TEMPLATE, template.ID, template); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func saveAllTemplates(templateList []core.Prioritizable) error {
	for _, template := range templateList {
		frt := template.(*corefw.FirmwareRuleTemplate)
		if err := frt.Validate(); err != nil {
			return err
		}
		frt.Updated = util.GetTimestamp()
		if err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_FIRMWARE_RULE_TEMPLATE, template.GetID(), template); err != nil {
			return err
		}
	}
	return nil
}
func firmwareRuleTemplatesToPrioritizables(frts []*corefw.FirmwareRuleTemplate) []core.Prioritizable {
	prioritizables := make([]core.Prioritizable, len(frts))
	for i, item := range frts {
		itemCopy := *item
		prioritizables[i] = &itemCopy
	}
	return prioritizables
}

func updateFirmwareRT(templateToUpdate corefw.FirmwareRuleTemplate, frtOnDb *corefw.FirmwareRuleTemplate) error {
	err := validateOneFirmwareRT(templateToUpdate)
	if err != nil {
		return err
	}

	//TODO
	firmwareRuleTemplateUpdateMutex.Lock()
	defer firmwareRuleTemplateUpdateMutex.Unlock()
	existingTemplate, err := corefw.GetFirmwareRuleTemplateOneDB(templateToUpdate.ID)
	if err != nil {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "FirmwareRuleTemplate does not exist for "+templateToUpdate.ID)
	}
	templatesByActionType, err := corefw.GetFirmwareRuleTemplateAllAsListDBForAS(templateToUpdate.ApplicableAction.ActionType)
	if err != nil {
		return err
	}
	err = validateAgainstFirmwareRTs(&templateToUpdate, templatesByActionType)
	if err != nil {
		return err
	}

	templatesByActionTypeCopy := firmwareRuleTemplatesToPrioritizables(templatesByActionType) //TODO
	list := UpdatePrioritizablePriorityAndReorganize(&templateToUpdate, templatesByActionTypeCopy, int(existingTemplate.Priority))
	if err = saveAllTemplates(list); err != nil {
		return err
	}
	return nil
}

func createFirmwareRT(template corefw.FirmwareRuleTemplate) (templ *corefw.FirmwareRuleTemplate, err error) {
	err = validateOneFirmwareRT(template)
	if err != nil {
		return nil, err
	}

	firmwareRuleTemplateUpdateMutex.Lock()
	defer firmwareRuleTemplateUpdateMutex.Unlock()
	templatesOfCurrentType, err := corefw.GetFirmwareRuleTemplateAllAsListDBForAS(template.ApplicableAction.ActionType)
	if err != nil {
		if err.Error() != common.NotFound.Error() {
			return nil, err
		}
	}
	err = validateAgainstFirmwareRTs(&template, templatesOfCurrentType)
	if err != nil {
		return nil, err
	}
	templatesOfCurrentTypeCopy := firmwareRuleTemplatesToPrioritizables(templatesOfCurrentType)
	reorganizedTemplates := AddNewPrioritizableAndReorganizePriorities(&template, templatesOfCurrentTypeCopy)
	if err = saveAllTemplates(reorganizedTemplates); err != nil {
		return nil, err
	}
	templ = &template

	return templ, nil
}

func importOrUpdateAllFirmwareRTs(entities []corefw.FirmwareRuleTemplate, successTag string, failedTag string) map[string][]string {
	result := make(map[string][]string)
	result[successTag] = []string{}
	result[failedTag] = []string{}

	for _, entity := range entities {
		if entity.GetName() == "" {
			result[failedTag] = append(result[failedTag], entity.GetName())
			continue
		}
		if entity.ID == "" {
			entity.ID = uuid.New().String()
		}
		entityOnDb, err := corefw.GetFirmwareRuleTemplateOneDBWithId(entity.ID)
		if err != nil {
			_, err = createFirmwareRT(entity)
		} else {
			err = updateFirmwareRT(entity, entityOnDb)
		}
		if err == nil {
			result[successTag] = append(result[successTag], entity.ID)
		} else {
			result[failedTag] = append(result[failedTag], entity.ID)
		}
	}
	return result
}

func GetFirmwareRuleTemplateById(id string) *corefw.FirmwareRuleTemplate {
	frt, err := corefw.GetFirmwareRuleTemplateOneDB(id)
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareRuleTemplateById: %v", err))
		return nil
	}
	return frt
}

func getFirmwareRuleTemplateExportName(all bool) string {
	if all {
		return "allFirmwareRuleTemplates"
	}
	return "firmwareRuleTemplate_"
}

func CreateFirmwareRuleTemplates() {
	if count, _ := xcorefw.GetFirmwareRuleTemplateCount(); count > 0 {
		return
	}

	log.Info("Creating templates...")

	ruleFactory := coreef.NewRuleFactory()
	templateList := []corefw.FirmwareRuleTemplate{}

	// Rule actions
	rule := coreef.NewMacRule(coreef.EMPTY_NAME)
	templateList = append(templateList, *xcorefw.NewFirmwareRuleTemplate(
		corefw.MAC_RULE, rule, coreef.EMPTY_LIST, 1))

	rule = ruleFactory.NewIpRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	templateList = append(templateList, *xcorefw.NewFirmwareRuleTemplate(
		corefw.IP_RULE, rule, coreef.EMPTY_LIST, 2))

	rule = ruleFactory.NewIntermediateVersionRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	templateList = append(templateList, *xcorefw.NewFirmwareRuleTemplate(
		corefw.IV_RULE, rule, []string{corefw.GLOBAL_PERCENT, corefw.TIME_FILTER}, 3))

	rule = ruleFactory.NewMinVersionCheckRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_LIST)
	templateList = append(templateList, *xcorefw.NewFirmwareRuleTemplate(
		corefw.MIN_CHECK_RULE, rule, []string{corefw.GLOBAL_PERCENT, corefw.TIME_FILTER}, 4))

	rule = ruleFactory.NewEnvModelRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	templ := *xcorefw.NewFirmwareRuleTemplate(corefw.ENV_MODEL_RULE, rule, []string{}, 5)
	templ.Editable = false
	templateList = append(templateList, templ)

	// Blocking filters
	rule = *ruleFactory.NewGlobalPercentFilterTemplate(coreef.DEFAULT_PERCENT, coreef.EMPTY_NAME)
	templ = *xcorefw.NewBlockingFilterTemplate(corefw.GLOBAL_PERCENT, rule, 1)
	templateList = append(templateList, templ)

	rule = *ruleFactory.NewIpFilter(coreef.EMPTY_NAME)
	templateList = append(templateList, *xcorefw.NewBlockingFilterTemplate(
		corefw.IP_FILTER, rule, 2))

	rule = *ruleFactory.NewTimeFilterTemplate(true, true, false, coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_NAME, "01:00", "02:00")
	templateList = append(templateList, *xcorefw.NewBlockingFilterTemplate(
		corefw.TIME_FILTER, rule, 3))

	// Define Properties
	rule = *ruleFactory.NewDownloadLocationFilter(coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	properties := map[string]corefw.PropertyValue{
		coreef.FIRMWARE_DOWNLOAD_PROTOCOL: *corefw.NewPropertyValue("tftp", false, corefw.STRING),
		coreef.FIRMWARE_LOCATION:          *corefw.NewPropertyValue("", false, corefw.STRING),
		coreef.IPV6_FIRMWARE_LOCATION:     *corefw.NewPropertyValue("", true, corefw.STRING),
	}
	templateList = append(templateList, *xcorefw.NewDefinePropertiesTemplate(
		corefw.DOWNLOAD_LOCATION_FILTER, rule, properties, coreef.EMPTY_LIST, 3))

	rule = *ruleFactory.NewRiFilterTemplate()
	properties = map[string]corefw.PropertyValue{
		coreef.REBOOT_IMMEDIATELY: *corefw.NewPropertyValue("true", false, corefw.BOOLEAN),
	}
	templateList = append(templateList, *xcorefw.NewDefinePropertiesTemplate(
		corefw.REBOOT_IMMEDIATELY_FILTER, rule, properties, coreef.EMPTY_LIST, 1))

	rule = ruleFactory.NewMinVersionCheckRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_LIST)
	properties = map[string]corefw.PropertyValue{
		coreef.REBOOT_IMMEDIATELY: *corefw.NewPropertyValue("true", true, corefw.BOOLEAN),
	}
	templateList = append(templateList, *xcorefw.NewDefinePropertiesTemplate(
		corefw.MIN_CHECK_RI, rule, properties, []string{corefw.GLOBAL_PERCENT, corefw.TIME_FILTER}, 2))

	rule = ruleFactory.NewActivationVersionRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	properties = map[string]corefw.PropertyValue{
		coreef.REBOOT_IMMEDIATELY: *corefw.NewPropertyValue("false", false, corefw.BOOLEAN),
	}
	templ = *xcorefw.NewDefinePropertiesTemplate(
		corefw.ACTIVATION_VERSION, rule, properties, coreef.EMPTY_LIST, 4)
	templ.Editable = false
	templateList = append(templateList, templ)

	for _, template := range templateList {
		if err := template.Validate(); err != nil {
			panic(err)
		}
		template.Updated = util.GetTimestamp()
		if jsonData, err := json.Marshal(template); err != nil {
			panic(err)
		} else {
			if err := ds.GetSimpleDao().SetOne(ds.TABLE_FIRMWARE_RULE_TEMPLATE, template.ID, jsonData); err != nil {
				panic(err)
			}
		}
	}
}
