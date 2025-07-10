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
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	ru "xconfwebconfig/rulesengine"

	xcommon "xconfadmin/common"
	xshared "xconfadmin/shared"
	xutil "xconfadmin/util"
	"xconfwebconfig/common"
	xwcommon "xconfwebconfig/common"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	coreef "xconfwebconfig/shared/estbfirmware"
	corefw "xconfwebconfig/shared/firmware"
	util "xconfwebconfig/util"

	uuid "github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	cFirmwareRuleName                 = xwcommon.NAME
	cFirmwareRuleKey                  = corefw.KEY
	cFirmwareRuleValue                = corefw.VALUE
	cFirmwareRuleFirmwareVersion      = "FIRMWARE_VERSION" // xwcommon.FIRMWARE_VERSION
	cFirmwareRuleTemplateId           = "TEMPLATE_ID"      // xwcommon.TEMPLATE_ID
	cFirmwareRuleApplicableActionType = xcommon.APPLICABLE_ACTION_TYPE
	cFirmwareRulePageNumber           = xcommon.PAGE_NUMBER
	cFirmwareRulePageSize             = xcommon.PAGE_SIZE
	cFirmwareRule                     = corefw.RULE
	cFirmwareRuleBlockingFilter       = corefw.BLOCKING_FILTER
	cFirmwareRuleDefineProperties     = corefw.DEFINE_PROPERTIES
)

func isValidFirmwareRuleContext(context map[string]string) error {
	val, ok := context[xwcommon.APPLICATION_TYPE]
	if !ok {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Mandatory param "+xwcommon.APPLICATION_TYPE+" missing")
	}
	if val == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Empty value for Mandatory param "+xwcommon.APPLICATION_TYPE)
	}
	return nil
}

func honoredByFirmwareRule(context map[string]string, rule *corefw.FirmwareRule) bool {
	appType, filterByApp := xutil.FindEntryInContext(context, xwcommon.APPLICATION_TYPE, false)
	if filterByApp && rule.ApplicationType != appType {
		return false
	}

	name, filterByName := xutil.FindEntryInContext(context, cFirmwareRuleName, false)
	if filterByName {
		baseName := strings.ToLower(rule.Name)
		givenName := strings.ToLower(name)
		if !strings.Contains(baseName, givenName) {
			return false
		}
	}

	templateId, filterByTemp := xutil.FindEntryInContext(context, cFirmwareRuleTemplateId, false)
	if filterByTemp && rule.GetTemplateId() != templateId {
		return false
	}

	fwVersion, filterByFW := xutil.FindEntryInContext(context, cFirmwareRuleFirmwareVersion, false)
	if filterByFW {
		appAction := rule.ApplicableAction
		if appAction == nil || appAction.ActionType != corefw.RULE {
			return false
		}
		configId := appAction.ConfigId
		ruleConfig, _ := coreef.GetFirmwareConfigOneDB(configId)
		if ruleConfig == nil || !strings.Contains(strings.ToLower(ruleConfig.Description), strings.ToLower(fwVersion)) {
			return false
		}
	}

	key, filterByKey := xutil.FindEntryInContext(context, cFirmwareRuleKey, false)
	if filterByKey && !re.IsExistConditionByFreeArgName(rule.Rule, key) {
		return false
	}

	val, filterByVal := xutil.FindEntryInContext(context, cFirmwareRuleValue, false)
	if filterByVal && !re.IsExistConditionByFixedArgValue(rule.Rule, val) {
		return false
	}

	actionType, filterByActionType := xutil.FindEntryInContext(context, cFirmwareRuleApplicableActionType, false)
	if filterByActionType && !util.IsBlank(actionType) {
		template, err := corefw.GetFirmwareRuleTemplateOneDB(rule.GetTemplateId())
		if err == nil && template.Editable {
			if rule.ApplicableAction != nil {
				baseName := strings.ToLower(string(rule.ApplicableAction.ActionType))
				givenName := strings.ToLower(actionType)
				if !strings.Contains(baseName, givenName) {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

func filterFirmwareRulesByContext(dbrules []*corefw.FirmwareRule, firmwareContext map[string]string) (filteredRules []*corefw.FirmwareRule) {
	for _, rule := range dbrules {
		if honoredByFirmwareRule(firmwareContext, rule) {
			filteredRules = append(filteredRules, rule)
		}
	}
	return filteredRules
}

func putSizesOfFirmwareRulesByTypeIntoHeaders(dbrules []*corefw.FirmwareRule) (headers map[string]string) {
	ruleCnt := 0
	blkFilterCnt := 0
	defPropCnt := 0

	for _, rule := range dbrules {
		template, err := corefw.GetFirmwareRuleTemplateOneDB(rule.GetTemplateId())
		if err == nil && template.Editable {
			if rule.ApplicableAction != nil {
				if rule.ApplicableAction.ActionType.CaseIgnoreEquals(cFirmwareRule) {
					ruleCnt++
				} else if rule.ApplicableAction.ActionType.CaseIgnoreEquals(cFirmwareRuleBlockingFilter) {
					blkFilterCnt++
				} else if rule.ApplicableAction.ActionType.CaseIgnoreEquals(cFirmwareRuleDefineProperties) {
					defPropCnt++
				}
			}
		}
	}
	headers = map[string]string{
		string(cFirmwareRule):                 strconv.Itoa(ruleCnt),
		string(cFirmwareRuleBlockingFilter):   strconv.Itoa(blkFilterCnt),
		string(cFirmwareRuleDefineProperties): strconv.Itoa(defPropCnt),
	}
	return headers
}

func extractFirmwareRulePage(list []*corefw.FirmwareRule, page int, pageSize int) (result []*corefw.FirmwareRule) {
	leng := len(list)
	result = make([]*corefw.FirmwareRule, 0)
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

func generateFirmwareRulePageByContext(dbrules []*corefw.FirmwareRule, contextMap map[string]string) (result []*corefw.FirmwareRule, err error) {
	pageNum := 1
	numStr, okval := contextMap[cFirmwareRulePageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cFirmwareRulePageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "pageNumber and pageSize should both be greater than zero")
	}

	return extractFirmwareRulePage(dbrules, pageNum, pageSize), nil
}

func firmwareRuleFilterByActionType(dbrules []*corefw.FirmwareRule, actionType string) (result []*corefw.FirmwareRule) {
	filteredRules := make([]*corefw.FirmwareRule, 0)
	for _, rule := range dbrules {
		template, err := corefw.GetFirmwareRuleTemplateOneDB(rule.GetTemplateId())
		if err == nil && template.Editable {
			baseName := strings.ToLower(string(rule.ApplicableAction.ActionType))
			givenName := strings.ToLower(actionType)
			if strings.Contains(baseName, givenName) {
				filteredRules = append(filteredRules, rule)
			}
		}
	}
	return filteredRules
}

func importOrUpdateAllFirmwareRules(firmwareRules []*corefw.FirmwareRule, appType string, fields log.Fields) (importResult map[string][]string) {
	result := make(map[string][]string)
	result["IMPORTED"] = []string{}
	result["NOT_IMPORTED"] = []string{}
	for _, entity := range firmwareRules {
		entity := entity
		entityOnDb, err := corefw.GetFirmwareRuleOneDB(entity.ID)
		if err == nil {
			err = checkRuleTypeAndUpdate(*entity, entityOnDb, appType, fields)
		} else {
			err = checkRuleTypeAndCreate(entity, appType, fields)
		}
		if err == nil {
			result["IMPORTED"] = append(result["IMPORTED"], entity.Name)
		} else {
			log.Error(fmt.Sprintf("Error Importing FirmwareRule %v: %v ", entity.Name, err.Error()))
			result["NOT_IMPORTED"] = append(result["NOT_IMPORTED"], entity.Name)
		}
	}
	return result
}

func checkRuleTypeAndCreate(firmwareRule *corefw.FirmwareRule, appType string, fields log.Fields) error {
	if util.IsBlank(firmwareRule.ID) {
		firmwareRule.ID = uuid.New().String()
	}
	if firmwareRule.Type == corefw.ENV_MODEL_RULE {
		ipRuleBean := coreef.ConvertFirmwareRuleToPercentageBean(firmwareRule)
		if ipRuleBean == nil {
			return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Unable to convert FirmwareRule into PercentageBean")
		}
		val := CreatePercentageBean(ipRuleBean, appType, fields)
		if val.Status == http.StatusCreated {
			return nil
		}
		return xwcommon.NewRemoteErrorAS(val.Status, val.Error.Error())
	}
	return createFirmwareRule(*firmwareRule, appType, true)
}

func checkRuleTypeAndUpdate(firmwareRule corefw.FirmwareRule, entityOnDb *corefw.FirmwareRule, appType string, fields log.Fields) error {
	if firmwareRule.Type == corefw.ENV_MODEL_RULE {
		ipRuleBean := coreef.ConvertFirmwareRuleToPercentageBean(&firmwareRule)
		if ipRuleBean == nil {
			return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Unable to convert FirmwareRule into PercentageBean")
		}
		val := UpdatePercentageBean(ipRuleBean, appType, fields)
		// assert (val != nil)
		if val.Status == http.StatusOK {
			return nil
		}
		return xwcommon.NewRemoteErrorAS(val.Status, val.Error.Error())
	}
	if entityOnDb.ApplicationType != firmwareRule.ApplicationType {
		return xwcommon.NewRemoteErrorAS(http.StatusConflict, "ApplicationType cannot be changed. Existing:"+entityOnDb.ApplicationType+" New: "+firmwareRule.ApplicationType)
	}
	return updateFirmwareRule(firmwareRule, appType, true)
}

func createFirmwareRule(entity corefw.FirmwareRule, appType string, validateNameNRule bool) error {
	if err := beforeCreatingFirmwareRule(entity); err != nil {
		return err
	}
	return saveFirmwareRule(entity, appType, validateNameNRule)
}

func updateFirmwareRule(entity corefw.FirmwareRule, appType string, validateNameNRule bool) error {
	if err := beforeUpdatingFirmwareRule(entity); err != nil {
		return err
	}
	return saveFirmwareRule(entity, appType, validateNameNRule)
}

func beforeCreatingFirmwareRule(entity corefw.FirmwareRule) error {
	if _, err := corefw.GetFirmwareRuleOneDB(entity.ID); err == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Entity with id: "+entity.ID+" already exists")

	}
	return nil
}

func beforeUpdatingFirmwareRule(firmwarerule corefw.FirmwareRule) error {
	id := firmwarerule.ID
	if util.IsBlank(id) {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "FirmwareRule id is empty")
	}
	entityOnDb, err := corefw.GetFirmwareRuleOneDB(id)
	if err != nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Entity with id: "+id+" does not exist")

	}
	if entityOnDb.ApplicationType != firmwarerule.ApplicationType {
		return xwcommon.NewRemoteErrorAS(http.StatusConflict, "ApplicationType cannot be changed. Existing:"+entityOnDb.ApplicationType+" New: "+firmwarerule.ApplicationType)
	}
	return nil
}

func saveFirmwareRule(entity corefw.FirmwareRule, appType string, validateNameNRule bool) error {
	if err := beforeSavingFirmwareRule(entity, appType, validateNameNRule); err != nil {
		return err
	}
	return corefw.CreateFirmwareRuleOneDB(&entity)
}

func superBeforeSavingFirmwareRule(entity corefw.FirmwareRule, validateNameNRule bool) error {
	if err := normalizeFirmwareRuleOnSave(entity); err != nil {
		return err
	}
	return validateFirmwareRuleOnSave(entity, validateNameNRule)
}

func beforeSavingFirmwareRule(entity corefw.FirmwareRule, appType string, validateNameNRule bool) error {
	if util.IsBlank(entity.ApplicationType) {
		entity.ApplicationType = appType
		// Can be done only after PermissionService is migrated
		// TODO entity.ApplicationType = getPermissionService().getWriteApplication()
	} else if entity.ApplicationType != appType {
		return xwcommon.NewRemoteErrorAS(http.StatusConflict, "ApplicationType conflict")
	}
	entity.Updated = util.GetTimestamp()
	return superBeforeSavingFirmwareRule(entity, validateNameNRule)
}

func normalizeFirmwareRuleOnSave(firmwareRule corefw.FirmwareRule) error {
	if firmwareRule.GetRule() != nil {
		if err := ru.NormalizeConditions(firmwareRule.GetRule()); err != nil {
			return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, firmwareRule.Name+": "+err.Error())
		}
	}
	return nil
}

func validateFirmwareRuleOnSave(firmwareRule corefw.FirmwareRule, validateNameNRule bool) error {
	err := validateOneFirmewareRule(firmwareRule)
	if err != nil {
		return err
	}
	if !validateNameNRule {
		return nil
	}

	rules, err := corefw.GetFirmwareRuleAllAsListByApplicationType(firmwareRule.ApplicationType)
	if err != nil {
		if err.Error() == common.NotFound.Error() {
			return nil
		}
		return err
	}
	return validateAgainstAllFirmwareRules(firmwareRule, rules)
}

func validateAgainstAllFirmwareRules(ruleToCheck corefw.FirmwareRule, existingRule map[string][]*corefw.FirmwareRule) error {
	ruleToCheckInUpperCase := strings.ToUpper(ruleToCheck.Name)
	for _, rules := range existingRule {
		for _, rule := range rules {
			if rule.ID == ruleToCheck.ID {
				continue
			}
			if ruleToCheckInUpperCase == strings.ToUpper(rule.GetName()) {
				return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.GetName()+": Name is already used")

			}
			if ru.EqualComplexRules(ruleToCheck.GetRule(), rule.GetRule()) {
				return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Rule has duplicate: "+rule.GetName())
			}
		}
	}
	return nil
}

func validateOneFirmewareRule(firmwareRule corefw.FirmwareRule) error {
	if util.IsBlank(firmwareRule.GetName()) {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Name is empty")

	}

	err := superValidate(firmwareRule)
	if err != nil {
		return err
	}
	err = validateTemplateConsistency(firmwareRule)
	if err != nil {
		return err
	}

	err = checkFreeArgExists(firmwareRule)
	if err != nil {
		return err
	}

	err = validateApplicableAction(firmwareRule)
	if err != nil {
		return err
	}

	return xshared.ValidateApplicationType(firmwareRule.ApplicationType)
}

func superValidate(entity corefw.FirmwareRule) error {
	if entity.GetRule() == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, entity.Name+": Rule is empty")
	}
	if err := ValidateRuleStructure(entity.GetRule()); err != nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, entity.Name+": "+err.Error())
	}
	if err := RunGlobalValidation(*entity.GetRule(), GetFirmwareRuleAllowedOperations); err != nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, entity.Name+": "+err.Error())
	}
	return nil
}

func validateTemplateConsistency(rule corefw.FirmwareRule) error {
	templateId := rule.GetTemplateId()
	template, _ := corefw.GetFirmwareRuleTemplateOneDB(templateId) // TODO should I pass false as a parameter

	if template == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, rule.Name+": Can't create rule from non existing template: "+templateId)
	}
	return validateRuleAgainstTemplate(&rule, template)
}

func validateRuleAgainstTemplate(rule *corefw.FirmwareRule, template *corefw.FirmwareRuleTemplate) error {

	ruleFreeArgs := getFreeArgList(ru.ToConditions(&rule.Rule))
	templateFreeArgs := getFreeArgList(ru.ToConditions(template.GetRule()))

	for _, v := range templateFreeArgs {
		ruleFreeArgs = remove(ruleFreeArgs, *v)
	}
	if len(ruleFreeArgs) != 0 {
		msg := ""
		for _, v := range getFreeArgNames(ruleFreeArgs) {
			msg += v + ","
		}
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": "+msg+" do(es) not belong to template "+template.ID)
	}
	return nil
}

func checkFreeArgExists(firmwareRule corefw.FirmwareRule) error {
	typef := firmwareRule.Type
	rule := firmwareRule.GetRule()
	if rule == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, firmwareRule.Name+" has empty Rule")
	}

	conditions := ru.ToConditions(rule)
	if len(conditions) == 0 {
		return nil
	}
	conditionInfos := ru.GetConditionInfos(conditions)

	if equalTypes(corefw.MAC_RULE, typef) {
		return ru.CheckFreeArgExists2(conditionInfos, *coreef.RuleFactoryMAC)
	} else if equalTypes(corefw.IP_RULE, typef) {
		err := ru.CheckFreeArgExists2(conditionInfos, *coreef.RuleFactoryIP)
		if err != nil {
			return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, firmwareRule.Name+": mismatches its template, "+err.Error())
		}
		err = ru.CheckFreeArgExists2(conditionInfos, *coreef.RuleFactoryENV)
		if err != nil {
			return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, firmwareRule.Name+": mismatches its template, "+err.Error())
		}
		err = ru.CheckFreeArgExists2(conditionInfos, *coreef.RuleFactoryMODEL)
		if err != nil {
			return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, firmwareRule.Name+": mismatches its template, "+err.Error())
		}
		return nil
	} else if equalTypes(corefw.ENV_MODEL_RULE, typef) {
		err := ru.CheckFreeArgExists2(conditionInfos, *coreef.RuleFactoryENV)
		if err != nil {
			return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, firmwareRule.Name+": mismatches its template, "+err.Error())
		}
		err = ru.CheckFreeArgExists2(conditionInfos, *coreef.RuleFactoryMODEL)
		if err != nil {
			return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, firmwareRule.Name+": mismatches its template, "+err.Error())
		}
		return nil
	} else if equalTypes(corefw.TIME_FILTER, typef) {
		err := ru.CheckFreeArgExists3(conditionInfos, *coreef.RuleFactoryLOCAL_TIME, re.StandardOperationGte)
		if err != nil {
			return err
		}
		return ru.CheckFreeArgExists3(conditionInfos, *coreef.RuleFactoryLOCAL_TIME, re.StandardOperationLte)
	} else if equalTypes(corefw.REBOOT_IMMEDIATELY_FILTER, typef) {
		return checkRebootImmediatelyFilter(conditionInfos, &firmwareRule)
	} else if equalTypes(corefw.GLOBAL_PERCENT, typef) {
		return ru.CheckFreeArgExists3(conditionInfos, *coreef.RuleFactoryMAC, re.StandardOperationPercent)
	} else if equalTypes(corefw.IP_FILTER, typef) {
		return ru.CheckFreeArgExists2(conditionInfos, *coreef.RuleFactoryIP)
	}
	return nil
}

func validateApplicableAction(rule corefw.FirmwareRule) error {
	action := rule.ApplicableAction
	if action == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": Applicable action must not be null")

	}

	if action.ActionType == corefw.RULE {
		return validateRuleAction(rule, *action)
	} else if action.ActionType == corefw.DEFINE_PROPERTIES {
		return validateDefinePropertiesApplicableAction(*action, rule.Type, &rule)
	}
	return nil
}

func validateRuleAction(firmwareRule corefw.FirmwareRule, action corefw.ApplicableAction) error {
	if util.IsBlank(action.ConfigId) {
		return nil // noop rule
	}

	err := validateConfigId(firmwareRule, action.ConfigId)
	if err != nil {
		return err
	}

	configEntries := action.ConfigEntries
	configList := []string{}
	if configEntries != nil {
		totalPercentage := 0
		for _, entry := range configEntries {

			configId := entry.ConfigId
			err = validateConfigId(firmwareRule, configId)
			if err != nil {
				return err
			}
			for _, v := range configList {
				if v == configId {
					return xwcommon.NewRemoteErrorAS(http.StatusConflict, firmwareRule.Name+": Distribution contains duplicate firmware configs")

				}
			}
			configList = append(configList, configId)

			percentage := entry.Percentage
			validatePercentageRange(percentage, "Percentage", &firmwareRule)
			totalPercentage += int(percentage)

			startPercentRange := entry.StartPercentRange
			validatePercentageRange(startPercentRange, "StartPercentRange", &firmwareRule)
			endPercentRange := entry.EndPercentRange
			validatePercentageRange(endPercentRange, "EndPercentRange", &firmwareRule)
		}
		if totalPercentage > 100 {
			return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, firmwareRule.Name+": Total percent sum should not be bigger than 100")
		}
	}
	return nil
}

func validateConfigId(firmwareRule corefw.FirmwareRule, configId string) error {
	if util.IsBlank(configId) {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, firmwareRule.Name+": ConfigId could not be blank")

	}

	firmwareConfig, _ := coreef.GetFirmwareConfigOneDB(configId)

	if firmwareConfig == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, firmwareRule.Name+": config '"+configId+"' doesn't exist")

	}

	if firmwareRule.ApplicationType != firmwareConfig.ApplicationType {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "ApplicationType ("+firmwareRule.ApplicationType+") of "+firmwareRule.Name+" does not match with applicationType ("+firmwareConfig.ApplicationType+") of its FirmwareConfig "+firmwareConfig.ID+"("+firmwareConfig.Description+")")

	}
	return nil
}

func validateDefinePropertiesApplicableAction(action corefw.ApplicableAction, templateType string, rule *corefw.FirmwareRule) error {
	if !util.IsBlank(templateType) {
		properties := action.Properties
		err := validateApplicableActionPropertiesGeneric(templateType, properties, rule)
		if err != nil {
			return err
		}
		err = validateApplicableActionPropertiesSpecific(templateType, properties, rule)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateApplicableActionPropertiesGeneric(templateType string, propertiesFromRule map[string]string, rule *corefw.FirmwareRule) error {
	template, _ := corefw.GetFirmwareRuleTemplateOneDBWithId(templateType)
	if template != nil {
		templateAction := template.ApplicableAction
		if templateAction.ActionType == corefw.DEFINE_PROPERTIES_TEMPLATE {
			templateProperties := templateAction.Properties
			for templatePropertyKey, templatePropertyValue := range templateProperties {
				err := validateCorrespondentPropertyFromRule(templatePropertyKey, templatePropertyValue, propertiesFromRule, rule)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateCorrespondentPropertyFromRule(templatePropertyKey string,
	templatePropertyValue corefw.PropertyValue,
	propertiesFromRule map[string]string,
	rule *corefw.FirmwareRule) error {
	correspondentPropertyFromRule := propertiesFromRule[templatePropertyKey]
	if util.IsBlank(correspondentPropertyFromRule) && !templatePropertyValue.Optional {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Property "+templatePropertyKey+" is required")

	} else if !util.IsBlank(correspondentPropertyFromRule) {
		validationTypes := templatePropertyValue.ValidationTypes
		if len(validationTypes) > 0 {
			return validatePropertyType(correspondentPropertyFromRule, validationTypes, rule)
		}
	}
	return nil
}

func validatePropertyType(propertyToValidate string, validationTypes []corefw.ValidationType, rule *corefw.FirmwareRule) error {
	for _, v := range validationTypes {
		if v == corefw.STRING {
			return nil
		}
	}
	if canBeNumber(validationTypes) && isNumber(propertyToValidate) {
		return nil
	}
	if canBeBoolean(validationTypes) && isBoolean(propertyToValidate) {
		return nil
	}
	if canBePercent(validationTypes) && isPercent(propertyToValidate) {
		return nil
	}
	if canBePort(validationTypes) && isPort(propertyToValidate) {
		return nil
	}
	if canBeUrl(validationTypes) && isUrl(propertyToValidate) {
		return nil
	}
	if canBeIpv4(validationTypes) && isIpv4(propertyToValidate) {
		return nil
	}
	if canBeIpv6(validationTypes) && isIpv6(propertyToValidate) {
		return nil
	}
	vstr := ""
	for _, v := range validationTypes {
		vstr = vstr + string(v) + ","
	}
	return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": Property must be one of "+vstr)
}

func validateApplicableActionPropertiesSpecific(templateType string, properties map[string]string, rule *corefw.FirmwareRule) error {
	switch templateType {
	case corefw.DOWNLOAD_LOCATION_FILTER:
		return validateDownloadLocationFilterApplicableActionProperties(properties, rule)
	case corefw.REBOOT_IMMEDIATELY_FILTER:
		return validateRebootImmediatelyFilterApplicableActionProperties(properties, rule)
	case corefw.MIN_CHECK_RI:
		return validateMinVersionCheckApplicableActionProperties(properties, rule)
	default:
		//do nothing
	}
	return nil
}

func validateDownloadLocationFilterApplicableActionProperties(properties map[string]string, rule *corefw.FirmwareRule) error {
	firmwareDownloadProtocol := properties[common.FIRMWARE_DOWNLOAD_PROTOCOL]
	firmwareLocation := properties[common.FIRMWARE_LOCATION]
	ipv6FirmwareLocation := properties[common.IPV6_FIRMWARE_LOCATION]

	err := validateFirmwareDowloadProtocol(firmwareDownloadProtocol, rule)
	if err != nil {
		return err
	}
	if firmwareDownloadProtocol == shared.Tftp {
		return validateTftpLocation(firmwareLocation, ipv6FirmwareLocation, rule)
	} else if firmwareDownloadProtocol == shared.Http {
		return validateHttpLocation(firmwareLocation, ipv6FirmwareLocation, rule)
	}
	return nil
}

func validateMinVersionCheckApplicableActionProperties(properties map[string]string, rule *corefw.FirmwareRule) error {
	rebootImmediately := properties[common.REBOOT_IMMEDIATELY]
	if !util.IsBlank(rebootImmediately) {
		return validateRebootImmediately(rebootImmediately, rule)
	}
	return nil
}

func validatePercentageRange(value float64, name string, rule *corefw.FirmwareRule) error {
	if value < 0 || value > 100 {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": "+name+" can not be negative or 0 or bigger that 100")
	}
	return nil
}

func canBeNumber(validationTypes []corefw.ValidationType) bool {
	for _, v := range validationTypes {
		if v == corefw.NUMBER {
			return true
		}
	}
	return false
	// Cant use util.Contains because it works only on collection of string, float or int. Not ValidationType
	// return util.Contains(validationTypes, corefw.NUMBER)
}

func canBeBoolean(validationTypes []corefw.ValidationType) bool {
	for _, v := range validationTypes {
		if v == corefw.BOOLEAN {
			return true
		}
	}
	return false
	// return util.Contains(validationTypes, corefw.BOOLEAN)
}

func canBePercent(validationTypes []corefw.ValidationType) bool {
	for _, v := range validationTypes {
		if v == corefw.PERCENT {
			return true
		}
	}
	return false
	// return util.Contains(validationTypes, corefw.PERCENT)
}

func canBePort(validationTypes []corefw.ValidationType) bool {
	for _, v := range validationTypes {
		if v == corefw.PORT {
			return true
		}
	}
	return false
	// return util.Contains(validationTypes, corefw.PORT)
}

func canBeUrl(validationTypes []corefw.ValidationType) bool {
	for _, v := range validationTypes {
		if v == corefw.URL {
			return true
		}
	}
	return false
	// return util.Contains(validationTypes, corefw.URL)
}

func canBeIpv4(validationTypes []corefw.ValidationType) bool {
	for _, v := range validationTypes {
		if v == corefw.IPV4 {
			return true
		}
	}
	return false
	// return util.Contains(validationTypes, corefw.IPV4)
}

func canBeIpv6(validationTypes []corefw.ValidationType) bool {
	for _, v := range validationTypes {
		if v == corefw.IPV6 {
			return true
		}
	}
	return false
	// return util.Contains(validationTypes, corefw.IPV6)
}

func isNumber(value string) bool {
	if _, err := strconv.Atoi(value); err == nil {
		return true
	}
	return false
}

func isBoolean(value string) bool {
	val := util.ToInt(value)
	return val == 1 || val == 0
}

func isPercent(value string) bool {
	val := util.ToInt(value)
	return val >= 0 && val <= 100
}

func isPort(value string) bool {
	val := util.ToInt(value)
	return val >= 1 && val <= 65535
}

func isUrl(value string) bool {
	// Ref: https://pkg.go.dev/net/url#ParseRequestURI
	_, err := url.Parse(value)
	return err != nil
}

func isIpv4(value string) bool {
	// Ref: https://pkg.go.dev/net#ParseIP
	return net.ParseIP(value) != nil
}

func isIpv6(value string) bool {
	// Ref: https://pkg.go.dev/net#ParseIP
	return net.ParseIP(value) != nil
}

func validateRebootImmediatelyFilterApplicableActionProperties(properties map[string]string, rule *corefw.FirmwareRule) error {
	return validateRebootImmediately(properties[common.REBOOT_IMMEDIATELY], rule)
}

func validateRebootImmediately(rebootImmediately string, rule *corefw.FirmwareRule) error {
	boolVal := util.ToInt(rebootImmediately)

	if boolVal == 0 || boolVal == 1 {
		return nil
	}
	return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": Reboot immediately must be boolean")
}

func validateFirmwareDowloadProtocol(firmwareDownloadProtocol string, rule *corefw.FirmwareRule) error {
	if firmwareDownloadProtocol != shared.Tftp && firmwareDownloadProtocol != shared.Http {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": FirmwareDownloadProtocol must be 'http' or 'tftp'")
	}
	return nil
}

func validateTftpLocation(firmwareLocation string, ipv6FirmwareLocation string, rule *corefw.FirmwareRule) error {
	ipAdd := shared.NewIpAddress(firmwareLocation)
	if ipAdd == nil || ipAdd.IsIpv6() {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": FirmwareLocation must be valid ipv4 address")
	}
	ipAdd = shared.NewIpAddress(ipv6FirmwareLocation)
	if ipAdd == nil || !ipAdd.IsIpv6() {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": Ipv6FirmwareLocation must be valid ipv6 address")
	}

	return nil
}

func validateHttpLocation(firmwareLocation string, ipv6FirmwareLocation string, rule *corefw.FirmwareRule) error {
	if util.IsBlank(firmwareLocation) {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": FirmwareLocation must not be empty")
	}
	if !util.IsBlank(ipv6FirmwareLocation) {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": Ipv6FirmwareLocation must be empty")
	}
	return nil
}

func checkRebootImmediatelyFilter(conditionInfos []ru.ConditionInfo, rule *corefw.FirmwareRule) error {
	ipExists := ru.FreeArgExists(conditionInfos, *coreef.RuleFactoryIP)
	macExists := ru.FreeArgExists(conditionInfos, *coreef.RuleFactoryMAC)
	envExists := ru.FreeArgExists(conditionInfos, *coreef.RuleFactoryENV)
	modelExists := ru.FreeArgExists(conditionInfos, *coreef.RuleFactoryMODEL)

	isValid := (ipExists || macExists) || (envExists && modelExists)

	if !isValid {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, rule.Name+": Need to set "+xwcommon.IP_ADDRESS+" OR "+xwcommon.ESTB_MAC_ADDRESS+
			" OR "+xwcommon.ENV+" AND "+xwcommon.MODEL)
	}
	return nil
}

func remove(items []*re.FreeArg, item re.FreeArg) []*re.FreeArg {
	newitems := []*re.FreeArg{}

	for _, i := range items {
		if *i != item {
			newitems = append(newitems, i)
		}
	}

	return newitems
}

func getFreeArgList(conditions []*re.Condition) (result []*re.FreeArg) {
	for _, cond := range conditions {
		result = append(result, cond.GetFreeArg())
	}
	return result
}

func getFreeArgNames(freeArgs []*re.FreeArg) (result []string) {
	for _, fa := range freeArgs {
		result = append(result, fa.Name)
	}
	return result
}

func GetFirmwareRuleById(id string) *corefw.FirmwareRule {
	fr, err := corefw.GetFirmwareRuleOneDB(id)
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareRuleById: %v", err))
		return nil
	}
	return fr
}
