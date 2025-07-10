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
	"errors"
	"fmt"
	"net/http"
	"reflect"

	xshared "xconfadmin/shared"
	xcoreef "xconfadmin/shared/estbfirmware"
	xwhttp "xconfwebconfig/http"
	"xconfwebconfig/shared"
	coreef "xconfwebconfig/shared/estbfirmware"
	"xconfwebconfig/shared/firmware"
	corefw "xconfwebconfig/shared/firmware"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

func GetPercentFilterFieldValues(fieldName string, applicationType string) (map[string][]interface{}, error) {
	fieldValues := make(map[interface{}]struct{})

	percentFilter, err := GetPercentFilter(applicationType)
	if err != nil {
		return nil, err
	}
	for _, envModelPercentage := range percentFilter.EnvModelPercentages {
		fValues := GetStructFieldValues(fieldName, reflect.ValueOf(envModelPercentage))
		for _, fieldValue := range fValues {
			fieldValues[fieldValue] = struct{}{}
		}
	}

	var resultFieldValues []interface{}
	for fieldValue := range fieldValues {
		resultFieldValues = append(resultFieldValues, fieldValue)
	}

	result := make(map[string][]interface{})
	result[fieldName] = resultFieldValues
	return result, nil
}

func GetPercentFilter(applicationType string) (*coreef.PercentFilterValue, error) {
	globalPercentageId := GetGlobalPercentageIdByApplication(applicationType)
	globalPercentageRule, err := firmware.GetFirmwareRuleOneDB(globalPercentageId)
	if err != nil {
		log.Warn(fmt.Sprintf("GetPercentFilter %v", err))
	}
	percentFilterValue := coreef.NewEmptyPercentFilterValue()
	if globalPercentageRule != nil {
		globalPercentage := coreef.ConvertIntoGlobalPercentageFirmwareRule(globalPercentageRule)
		percentFilterValue.Percentage = globalPercentage.Percentage
		if !util.IsBlank(globalPercentage.Whitelist) {
			percentFilterValue.Whitelist = getIpAddressGroup(globalPercentage.Whitelist)
		}
	}
	percentFilterValue.EnvModelPercentages = make(map[string]coreef.EnvModelPercentage)

	firmwareRules, err := corefw.GetEnvModelFirmwareRules(applicationType)
	if err != nil {
		log.Error(fmt.Sprintf("GetPercentFilter: %v", err))
		return nil, err
	}
	for _, firmwareRule := range firmwareRules {
		percentageBean := coreef.ConvertFirmwareRuleToPercentageBean(firmwareRule)
		percentFilterValue.EnvModelPercentages[firmwareRule.Name] = *convertPercentageBean(percentageBean)
	}

	return percentFilterValue, nil
}

func UpdatePercentFilter(applicationType string, filter *coreef.PercentFilterWrapper) *xwhttp.ResponseEntity {
	if err := xshared.ValidateApplicationType(applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if filter.Percentage < 0 || filter.Percentage > 100 {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Percentage should be within [0, 100]"), nil)
	}

	if filter.Whitelist != nil && IsChangedIpAddressGroup(filter.Whitelist) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest,
			fmt.Errorf("IP address group denoted by '%s' does not match any existing ipAddressGroup", filter.Whitelist.Name), nil)
	}

	for idx, percentage := range filter.EnvModelPercentages {
		if percentage.FirmwareCheckRequired && len(percentage.FirmwareVersions) == 0 {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("FirmwareVersion is required: %s", percentage.Name), nil)
		}

		if !util.IsBlank(percentage.LastKnownGood) {
			if percentage.Percentage == 100.0 {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Can't set LastKnownGood when percentage=100: %s", percentage.Name), nil)
			}
			if !percentage.Active {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Can't set LastKnownGood when filter is not active: %s", percentage.Name), nil)
			}

			configId := GetFirmwareConfigId(percentage.LastKnownGood, applicationType)
			if configId == "" {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("No version in firmware configs matches LastKnownGood value: %s", percentage.LastKnownGood), nil)
			}
			// Update the original object since percentage is only a copy
			filter.EnvModelPercentages[idx].LastKnownGood = configId
		}

		if !util.IsBlank(percentage.IntermediateVersion) {
			if !percentage.FirmwareCheckRequired {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Can't set IntermediateVersion when firmware check is disabled: %s", percentage.Name), nil)
			}

			configId := GetFirmwareConfigId(percentage.IntermediateVersion, applicationType)
			if configId == "" {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("No version in firmware configs matches IntermediateVersion value: %s", percentage.IntermediateVersion), nil)
			}
			// Update the original object since percentage is only a copy
			filter.EnvModelPercentages[idx].IntermediateVersion = configId
		}

		if percentage.Percentage < 0 || percentage.Percentage > 100 {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Percentage should be within [0, 100]: %s", percentage.Name), nil)
		}
	}

	percentFilterValue := filter.ToPercentFilterValue()

	globalPercentage := coreef.ConvertIntoGlobalPercentage(percentFilterValue, applicationType)
	if globalPercentage != nil {
		err := firmware.CreateFirmwareRuleOneDB(globalPercentage)
		if err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
	}

	firmwareRules, err := corefw.GetEnvModelFirmwareRulesForAS(applicationType)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	for _, firmwareRule := range firmwareRules {
		envModelPercentage := percentFilterValue.GetEnvModelPercentage(firmwareRule.Name)
		if envModelPercentage != nil {
			percentageBean := xcoreef.MigrateIntoPercentageBean(envModelPercentage, firmwareRule)
			convertedRule := coreef.ConvertPercentageBeanToFirmwareRule(*percentageBean)
			convertedRule.ApplicationType = applicationType
			err := corefw.CreateFirmwareRuleOneDB(convertedRule)
			if err != nil {
				return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
			}
		}
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, filter)
}

func getIpAddressGroup(groupId string) *shared.IpAddressGroup {
	if list := GetNamespacedListById(groupId); list != nil {
		return shared.ConvertToIpAddressGroup(list)
	}

	return nil
}

func convertPercentageBean(bean *coreef.PercentageBean) *coreef.EnvModelPercentage {
	percentage := coreef.NewEnvModelPercentage()
	percentage.Active = bean.Active
	percentage.FirmwareCheckRequired = bean.FirmwareCheckRequired
	percentage.LastKnownGood = bean.LastKnownGood
	percentage.FirmwareVersions = bean.FirmwareVersions
	percentage.IntermediateVersion = bean.IntermediateVersion
	percentage.RebootImmediately = bean.RebootImmediately
	if !util.IsBlank(bean.Whitelist) {
		percentage.Whitelist = getIpAddressGroup(bean.Whitelist)
	}
	percentage.Percentage = float32(getPercentageSum(bean.Distributions))

	return percentage
}

func getPercentageSum(distribution []*corefw.ConfigEntry) float64 {
	var total float64
	for _, entry := range distribution {
		if entry != nil {
			total += entry.Percentage
		}
	}
	return total
}

func getPercentFilterValue(applicationType string) coreef.PercentFilterValue {
	return *coreef.NewEmptyPercentFilterValue()
}
