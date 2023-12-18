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
package queries

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"xconfadmin/common"
	xshared "xconfadmin/shared"
	"xconfadmin/util"
	xcommon "xconfwebconfig/common"
	xwhttp "xconfwebconfig/http"
	re "xconfwebconfig/rulesengine"
	ru "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	coreef "xconfwebconfig/shared/estbfirmware"
	"xconfwebconfig/shared/firmware"
	xutil "xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	PERCENTAGE_FIELD_NAME = "percentage"
	WHITELIST_FIELD_NAME  = "whitelist"
)

// Service APIs for Percent Filter Rule

func GetOnePercentageBeanFromDB(id string) (*coreef.PercentageBean, error) {
	frule, err := firmware.GetFirmwareRuleOneDB(id)
	if err != nil {
		return nil, err
	}

	bean := coreef.ConvertFirmwareRuleToPercentageBean(frule)
	return bean, nil
}

func GetAllGlobalPercentageBeansAsRuleFromDB(applicationType string, sortByName bool) ([]*firmware.FirmwareRule, error) {
	frules, err := firmware.GetFirmwareRuleAllAsListDB()
	if err != nil {
		return nil, err
	}

	var result []*firmware.FirmwareRule

	for _, frule := range frules {
		if frule.ApplicationType == applicationType && frule.Type == firmware.ENV_MODEL_RULE {
			result = append(result, frule)
		}
	}

	if sortByName {
		sort.Slice(result, func(i, j int) bool {
			return strings.Compare(strings.ToLower(result[i].Name), strings.ToLower(result[j].Name)) < 0
		})
	}

	return result, nil
}

func GetAllPercentageBeansFromDB(applicationType string, sortByName bool, convert bool) ([]*coreef.PercentageBean, error) {
	firmwareRules, err := firmware.GetFirmwareRuleAllAsListDB()
	if err != nil {
		return nil, err
	}

	result := []*coreef.PercentageBean{}

	for _, frule := range firmwareRules {
		if frule.ApplicationType == applicationType && frule.Type == firmware.ENV_MODEL_RULE {
			bean := coreef.ConvertFirmwareRuleToPercentageBean(frule)
			if convert {
				replaceFieldsWithFirmwareVersion(bean)
			}
			result = append(result, bean)
		}
	}

	if sortByName {
		sort.Slice(result, func(i, j int) bool {
			return strings.Compare(strings.ToLower(result[i].Name), strings.ToLower(result[j].Name)) < 0
		})
	}

	return result, nil
}

func GetPercentageBeanFilterFieldValues(fieldName string, applicationType string) (map[string][]interface{}, error) {
	fieldValues, err := getPercentageBeanFieldValues(fieldName, applicationType)
	if err != nil {
		return nil, err
	}

	globalFieldValues := getGlobalPercentageFields(fieldName, applicationType)
	for fieldValue := range globalFieldValues {
		fieldValues[fieldValue] = struct{}{}
	}

	resultFieldValues := make([]interface{}, 0)
	for fieldValue := range fieldValues {
		resultFieldValues = append(resultFieldValues, fieldValue)
	}

	result := make(map[string][]interface{})
	result[fieldName] = resultFieldValues
	return result, nil
}

func getGlobalPercentageFields(fieldName string, applicationType string) map[interface{}]struct{} {
	resultFieldValues := make(map[interface{}]struct{})

	globalPercentageId := GetGlobalPercentageIdByApplication(applicationType)
	globalPercentageRule, err := firmware.GetFirmwareRuleOneDB(globalPercentageId)
	if err != nil {
		log.Error(fmt.Sprintf("GetGlobalPercentageFields: %v", err))
		if fieldName == PERCENTAGE_FIELD_NAME {
			resultFieldValues[100] = struct{}{}
		}
		return resultFieldValues
	}

	globalPercentage := coreef.ConvertIntoGlobalPercentageFirmwareRule(globalPercentageRule)
	fieldValues := GetStructFieldValues(fieldName, reflect.ValueOf(*globalPercentage))
	for _, fieldValue := range fieldValues {
		resultFieldValues[fieldValue] = struct{}{}
	}

	return resultFieldValues
}

func getPercentageBeanFieldValues(fieldName string, applicationType string) (map[interface{}]struct{}, error) {
	resultFieldValues := make(map[interface{}]struct{})

	beans, err := GetAllPercentageBeansFromDB(applicationType, false, true)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(fieldName, "distributions") {
		configs := make(map[string]*firmware.ConfigEntry)
		for _, bean := range beans {
			for _, configEntry := range bean.Distributions {
				configs[configEntry.ConfigId] = configEntry
			}
		}
		for _, configEntry := range configs {
			resultFieldValues[configEntry] = struct{}{}
		}
	} else {
		for _, bean := range beans {
			fieldValues := GetStructFieldValues(fieldName, reflect.ValueOf(*bean))
			for _, fieldValue := range fieldValues {
				resultFieldValues[fieldValue] = struct{}{}
			}
		}
	}
	return resultFieldValues, nil
}

func GetGlobalPercentageIdByApplication(applicationType string) string {
	if xshared.ApplicationTypeEquals(applicationType, shared.STB) {
		return firmware.GLOBAL_PERCENT
	}
	return fmt.Sprintf("%s_%s", strings.ToUpper(applicationType), firmware.GLOBAL_PERCENT)
}

func GetStructFieldValues(fieldName string, structValue reflect.Value) []interface{} {
	var resultFieldValues []interface{}

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Type().Field(i)
		value := structValue.Field(i).Interface()
		if strings.EqualFold(fieldName, field.Name) {
			switch field.Type.Kind() {
			case reflect.Bool, reflect.Float32, reflect.Float64, reflect.Ptr:
				resultFieldValues = append(resultFieldValues, value)
			case reflect.String:
				if str, ok := value.(string); ok && str != "" {
					resultFieldValues = append(resultFieldValues, str)
				}
			case reflect.Slice:
				if reflect.TypeOf(value).Elem().Kind() == reflect.String {
					for _, str := range value.([]string) {
						resultFieldValues = append(resultFieldValues, str)
					}
				}
			}
			break
		}
	}

	return resultFieldValues
}

func CreatePercentageBean(bean *coreef.PercentageBean, applicationType string) *xwhttp.ResponseEntity {
	_, err := firmware.GetFirmwareRuleOneDB(bean.ID)
	if err == nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s Already Exist", bean.ID), nil)
	}

	if applicationType != bean.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", bean.ID), nil)
	}

	if err := firmware.ValidateRuleName(bean.ID, bean.Name); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if err := bean.Validate(); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	beans, err := GetAllPercentageBeansFromDB(bean.ApplicationType, false, true)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if err := bean.ValidateAll(beans); err != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, err, nil)
	}

	firmware.SortConfigEntry(bean.Distributions)

	fRule := coreef.ConvertPercentageBeanToFirmwareRule(*bean)
	ru.NormalizeConditions(&fRule.Rule)
	if err := firmware.CreateFirmwareRuleOneDB(fRule); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	newBean := coreef.ConvertFirmwareRuleToPercentageBean(fRule)

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, newBean)
}

func UpdatePercentageBean(bean *coreef.PercentageBean, applicationType string) *xwhttp.ResponseEntity {
	if xutil.IsBlank(bean.ID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Entity id is empty"), nil)
	}

	fRule, err := firmware.GetFirmwareRuleOneDB(bean.ID)
	if fRule == nil || err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Entity with id: %s does not exist", bean.ID), nil)
	}
	if fRule.ApplicationType != applicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id: %s ApplicationType  Mismatch", bean.ID), nil)
	}
	if fRule.ApplicationType != bean.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("ApplicationType cannot be changed: Existing value:"+fRule.ApplicationType+" New Value:"+bean.ApplicationType), nil)
	}

	if err := firmware.ValidateRuleName(bean.ID, bean.Name); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if err := bean.Validate(); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	beans, err := GetAllPercentageBeansFromDB(bean.ApplicationType, false, true)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if err := bean.ValidateAll(beans); err != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, err, nil)
	}

	firmware.SortConfigEntry(bean.Distributions)

	fRule = coreef.ConvertPercentageBeanToFirmwareRule(*bean)
	ru.NormalizeConditions(&fRule.Rule)
	if err := firmware.CreateFirmwareRuleOneDB(fRule); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	newBean := coreef.ConvertFirmwareRuleToPercentageBean(fRule)

	return xwhttp.NewResponseEntity(http.StatusOK, nil, newBean)
}

func DeletePercentageBean(id string, app string) *xwhttp.ResponseEntity {
	fRule, err := firmware.GetFirmwareRuleOneDB(id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("Entity with id: %s does not exist", id), nil)
	}
	if fRule.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("Entity with id: %s ApplicationType doesn't match", id), nil)
	}
	if err = firmware.DeleteOneFirmwareRule(fRule.ID); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func percentageBeanGeneratePage(list []*coreef.PercentageBean, page int, pageSize int) (result []*coreef.PercentageBean) {
	leng := len(list)
	startIndex := page*pageSize - pageSize
	result = make([]*coreef.PercentageBean, 0)
	if page < 1 || startIndex > leng || pageSize < 1 {
		return result
	}
	lastIndex := leng
	if page*pageSize < len(list) {
		lastIndex = page * pageSize
	}

	return list[startIndex:lastIndex]
}

func PercentageBeanRuleGeneratePageWithContext(pbrules []*coreef.PercentageBean, contextMap map[string]string) (result []*coreef.PercentageBean, err error) {
	sort.Slice(pbrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(pbrules[i].Name), strings.ToLower(pbrules[j].Name)) < 0
	})
	pageNum := 1
	numStr, okval := contextMap[cPercentageBeanPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cPercentageBeanPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, errors.New("pageNumber and pageSize should both be greater than zero")
	}
	return percentageBeanGeneratePage(pbrules, pageNum, pageSize), nil
}

func PercentageBeanFilterByContext(searchContext map[string]string, applicationType string) []*coreef.PercentageBean {
	percentageBeansSearchResult := []*coreef.PercentageBean{}
	percentageBeans, err := GetAllPercentageBeansFromDB(applicationType, true, false)
	if err != nil {
		return percentageBeansSearchResult
	}
	for _, pbRule := range percentageBeans {
		if pbRule == nil {
			continue
		}
		if pbRule.ApplicationType != applicationType {
			continue
		}
		if name, ok := util.FindEntryInContext(searchContext, common.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(pbRule.Name), strings.ToLower(name)) {
				continue
			}
		}
		if env, ok := util.FindEntryInContext(searchContext, cPercentageBeanenvironment, false); ok {
			if !strings.Contains(strings.ToLower(pbRule.Environment), strings.ToLower(env)) {
				continue
			}
		}
		if lkg, ok := util.FindEntryInContext(searchContext, cPercentageBeanlastknowngood, false); ok {
			fc, err := coreef.GetFirmwareConfigOneDB(pbRule.LastKnownGood)
			if err != nil {
				continue
			}

			if !strings.Contains(strings.ToLower(fc.FirmwareVersion), strings.ToLower(lkg)) {
				continue
			}
		}
		if intver, ok := util.FindEntryInContext(searchContext, cPercentageBeanintermediateversion, false); ok {
			fc, err := coreef.GetFirmwareConfigOneDB(pbRule.IntermediateVersion)
			if err != nil {
				continue
			}
			if !strings.Contains(strings.ToLower(fc.FirmwareVersion), strings.ToLower(intver)) {
				continue
			}
		}
		if minCheckVersion, ok := util.FindEntryInContext(searchContext, cPercentageBeanmincheckversion, false); ok {
			if !containsMinCheckVersion(minCheckVersion, pbRule.FirmwareVersions) {
				continue
			}
		}

		if model, ok := util.FindEntryInContext(searchContext, xcommon.MODEL, false); ok {
			if !strings.Contains(strings.ToLower(pbRule.Model), strings.ToLower(model)) {
				continue
			}
		}

		if key, ok := util.FindEntryInContext(searchContext, common.FREE_ARG, false); ok {
			if pbRule.OptionalConditions == nil {
				continue
			}
			if !re.IsExistConditionByFreeArgName(*pbRule.OptionalConditions, key) {
				continue
			}
		}
		val, ok := util.FindEntryInContext(searchContext, common.FIXED_ARG, false)
		if ok {
			if pbRule.OptionalConditions == nil {
				continue
			}
			if !re.IsExistConditionByFixedArgValue(*pbRule.OptionalConditions, val) {
				continue
			}
		}
		percentageBeansSearchResult = append(percentageBeansSearchResult, pbRule)
	}
	return percentageBeansSearchResult
}

func containsMinCheckVersion(versionToSearch string, firmwareVersions []string) bool {
	if len(firmwareVersions) > 0 {
		for _, firmwareVersion := range firmwareVersions {
			if strings.Contains(strings.ToLower(firmwareVersion), strings.ToLower(versionToSearch)) {
				return true
			}
		}
	}
	return false
}

func replaceFieldsWithFirmwareVersion(bean *coreef.PercentageBean) *coreef.PercentageBean {
	if bean.LastKnownGood != "" {
		firmwareVersion := coreef.GetFirmwareVersion(bean.LastKnownGood)
		bean.LastKnownGood = firmwareVersion
	}

	if bean.IntermediateVersion != "" {
		firmwareVersion := coreef.GetFirmwareVersion(bean.IntermediateVersion)
		bean.IntermediateVersion = firmwareVersion
	}

	if bean.Distributions != nil && len(bean.Distributions) > 0 {
		firmwareVersionDistributions := make([]*firmware.ConfigEntry, 0)
		for _, dist := range bean.Distributions {
			if dist.ConfigId != "" {
				firmwareVersion := coreef.GetFirmwareVersion(dist.ConfigId)
				if firmwareVersion != "" {
					firmwareVersionDistributions = append(firmwareVersionDistributions, firmware.NewConfigEntry(firmwareVersion, dist.StartPercentRange, dist.EndPercentRange))
				} else {
					firmwareVersionDistributions = append(firmwareVersionDistributions, dist)
				}
			}
		}
		bean.Distributions = firmwareVersionDistributions
	}

	return bean
}
