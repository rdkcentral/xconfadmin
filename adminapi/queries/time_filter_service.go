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
	"net/http"
	"strings"
	"time"

	coreef "xconfadmin/shared/estbfirmware"
	xcoreef "xconfwebconfig/shared/estbfirmware"
	"xconfwebconfig/util"

	daef "xconfwebconfig/dataapi/estbfirmware"

	xshared "xconfadmin/shared"
	xwhttp "xconfwebconfig/http"
	"xconfwebconfig/shared/firmware"
	corefw "xconfwebconfig/shared/firmware"
)

func UpdateTimeFilter(applicationType string, timeFilter *xcoreef.TimeFilter) *xwhttp.ResponseEntity {
	if util.IsBlank(timeFilter.Name) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Name is blank"), nil)
	}

	if err := xshared.ValidateApplicationType(applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if err := firmware.ValidateRuleName(timeFilter.Id, timeFilter.Name); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if _, err := time.Parse("15:04", timeFilter.Start); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if _, err := time.Parse("15:04", timeFilter.End); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if IsChangedIpAddressGroup(timeFilter.IpWhiteList) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("IP address group is not matched by existed IP address group"), nil)
	}

	if !IsExistEnvModelRule(timeFilter.EnvModelRuleBean, applicationType) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Env/Model does not match with existed Env/Model"), nil)
	}

	timeFilter.EnvModelRuleBean.EnvironmentId = strings.ToUpper(timeFilter.EnvModelRuleBean.EnvironmentId)
	timeFilter.EnvModelRuleBean.ModelId = strings.ToUpper(timeFilter.EnvModelRuleBean.ModelId)

	firmwareRule := coreef.ConvertTimeFilterToFirmwareRule(timeFilter)

	if !util.IsBlank(applicationType) {
		firmwareRule.ApplicationType = applicationType
	}

	if err := xshared.ValidateApplicationType(firmwareRule.ApplicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	err := corefw.CreateFirmwareRuleOneDB(firmwareRule)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if timeFilter.Id == "" {
		timeFilter.Id = firmwareRule.ID
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, timeFilter)
}

func DeleteTimeFilter(name string, applicationType string) *xwhttp.ResponseEntity {
	timeFilter, err := xcoreef.TimeFilterByName(name, applicationType)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if timeFilter != nil {
		err = corefw.DeleteOneFirmwareRule(timeFilter.Id)
		if err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func IsExistEnvModelRule(envModelRule xcoreef.EnvModelRuleBean, applicationType string) bool {
	if envModelRule.Id != "" && envModelRule.ModelId != "" {
		bean := GetOneByEnvModel(envModelRule.ModelId, envModelRule.EnvironmentId, applicationType)
		return bean != nil
	}
	return false
}

func GetOneByEnvModel(model string, environment string, applicationType string) *xcoreef.EnvModelBean {
	emRuleService := daef.EnvModelRuleService{}
	emRuleBeans := emRuleService.GetByApplicationType(applicationType)
	for _, emRuleBean := range emRuleBeans {
		if strings.EqualFold(emRuleBean.ModelId, model) && strings.EqualFold(emRuleBean.EnvironmentId, environment) {
			return emRuleBean
		}
	}
	return nil
}
