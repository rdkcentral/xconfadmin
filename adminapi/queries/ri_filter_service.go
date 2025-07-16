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
	"strings"

	xshared "xconfadmin/shared"
	xcoreef "xconfadmin/shared/estbfirmware"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
)

func UpdateRebootImmediatelyFilter(applicationType string, rebootFilter *coreef.RebootImmediatelyFilter) *xwhttp.ResponseEntity {
	if util.IsBlank(rebootFilter.Name) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Rule name is empty"), nil)
	}

	if err := xshared.ValidateApplicationType(applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if err := firmware.ValidateRuleName(rebootFilter.Id, rebootFilter.Name, applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if util.IsBlank(rebootFilter.MacAddress) && len(rebootFilter.IpAddressGroup) == 0 &&
		len(rebootFilter.Environments) == 0 && len(rebootFilter.Models) == 0 {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Please specify at least one of filter criteria."), nil)
	}

	modelIds := util.Set{}
	for _, model := range rebootFilter.Models {
		id := strings.ToUpper(model)
		if !IsExistModel(id) {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Model %s is not exist", id), nil)
		}
		modelIds.Add(id)
	}
	rebootFilter.Models = modelIds.ToSlice()

	envIds := util.Set{}
	for _, env := range rebootFilter.Environments {
		id := strings.ToUpper(env)
		if !IsExistEnvironment(id) {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Environment %s is not exist", id), nil)
		}
		envIds.Add(id)
	}
	rebootFilter.Environments = envIds.ToSlice()

	if rebootFilter.IpAddressGroup != nil {
		for _, ipAddressGroup := range rebootFilter.IpAddressGroup {
			if ipAddressGroup != nil && IsChangedIpAddressGroup(ipAddressGroup) {
				return xwhttp.NewResponseEntity(http.StatusBadRequest,
					fmt.Errorf("IP address group denoted by '%s' does not match any existing ipAddressGroup", ipAddressGroup.Name), nil)
			}
		}
	}

	if _, err := xcoreef.GetNormalizedMacAddresses(rebootFilter.MacAddress); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	filterToUpdate, err := xcoreef.RebootImmediatelyFiltersByName(applicationType, rebootFilter.Name)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	status := http.StatusCreated
	if filterToUpdate != nil {
		rebootFilter.Id = filterToUpdate.Id
		status = http.StatusOK
	}

	firmwareRule, err := SaveRebootImmediatelyFilter(rebootFilter, applicationType)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if rebootFilter.Id == "" {
		rebootFilter.Id = firmwareRule.ID
	}

	return xwhttp.NewResponseEntity(status, nil, rebootFilter)
}

func DeleteRebootImmediatelyFilter(name string, applicationType string) *xwhttp.ResponseEntity {
	if err := xshared.ValidateApplicationType(applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	rebootFilter, err := xcoreef.RebootImmediatelyFiltersByName(applicationType, name)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if rebootFilter != nil {
		err = corefw.DeleteOneFirmwareRule(rebootFilter.Id)
		if err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func SaveRebootImmediatelyFilter(filter *coreef.RebootImmediatelyFilter, applicationType string) (*corefw.FirmwareRule, error) {
	firmwareRule, err := xcoreef.ConvertRebootFilterToFirmwareRule(filter)
	if err != nil {
		return nil, err
	}

	if !util.IsBlank(applicationType) {
		firmwareRule.ApplicationType = applicationType
	}

	err = corefw.CreateFirmwareRuleOneDB(firmwareRule)
	if err != nil {
		return nil, err
	}

	return firmwareRule, nil
}
