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

	xcoreef "xconfadmin/shared/estbfirmware"
	xwhttp "xconfwebconfig/http"
	"xconfwebconfig/shared/estbfirmware"
	coreef "xconfwebconfig/shared/estbfirmware"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"

	xshared "xconfadmin/shared"
	"xconfwebconfig/shared/firmware"
	corefw "xconfwebconfig/shared/firmware"
)

func UpdateLocationFilter(applicationType string, locationFilter *coreef.DownloadLocationFilter) *xwhttp.ResponseEntity {
	if util.IsBlank(locationFilter.Name) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Name is blank"), nil)
	}

	if err := xshared.ValidateApplicationType(applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if err := firmware.ValidateRuleName(locationFilter.Id, locationFilter.Name, applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if locationFilter.IpAddressGroup == nil {
		if len(locationFilter.Environments) == 0 {
			if len(locationFilter.Models) == 0 {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Condition is required"), nil)
			}
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Environments are required"), nil)
		}
		if len(locationFilter.Models) == 0 {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Models are required"), nil)
		}
	}

	modelIds := util.Set{}
	for _, model := range locationFilter.Models {
		id := strings.ToUpper(model)
		if !IsExistModel(id) {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Model %s is not exist", id), nil)
		}
		modelIds.Add(id)
	}
	locationFilter.Models = modelIds.ToSlice()

	envIds := util.Set{}
	for _, env := range locationFilter.Environments {
		id := strings.ToUpper(env)
		if !IsExistEnvironment(id) {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Environment %s is not exist", id), nil)
		}
		envIds.Add(id)
	}
	locationFilter.Environments = envIds.ToSlice()

	if locationFilter.IpAddressGroup != nil && IsChangedIpAddressGroup(locationFilter.IpAddressGroup) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest,
			fmt.Errorf("IP address group denoted by '%s' does not match any existing ipAddressGroup", locationFilter.IpAddressGroup.Name), nil)
	}

	if !locationFilter.ForceHttp && locationFilter.Ipv6FirmwareLocation != nil && locationFilter.FirmwareLocation == nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("If you are not forcing HTTP, you can't use IPv6 without IPv4 location"), nil)
	}

	if util.IsBlank(locationFilter.HttpLocation) {
		if locationFilter.FirmwareLocation == nil {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Any location are required"), nil)
		}
		if locationFilter.ForceHttp {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("HTTP location is required"), nil)
		}
	}

	if locationFilter.FirmwareLocation != nil {
		if locationFilter.FirmwareLocation.IsIpv6() {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Version is invalid"), nil)
		} else if locationFilter.FirmwareLocation.IsCidrBlock() {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("IP addresss is invalid"), nil)
		}
	}

	if locationFilter.Ipv6FirmwareLocation != nil {
		if locationFilter.Ipv6FirmwareLocation.IsIpv6() {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Version is invalid"), nil)
		} else if locationFilter.Ipv6FirmwareLocation.IsCidrBlock() {
			return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("IP addresss is invalid"), nil)
		}
	}

	firmwareRule, err := SaveDownloadLocationFilter(locationFilter, applicationType)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if locationFilter.Id == "" {
		locationFilter.Id = firmwareRule.ID
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, locationFilter)
}

func DeleteLocationFilter(name string, applicationType string) *xwhttp.ResponseEntity {
	if err := xshared.ValidateApplicationType(applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	locationFilter, err := coreef.DownloadLocationFiltersByName(applicationType, name)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if locationFilter != nil {
		err = corefw.DeleteOneFirmwareRule(locationFilter.Id)
		if err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func SaveDownloadLocationFilter(filter *coreef.DownloadLocationFilter, applicationType string) (*corefw.FirmwareRule, error) {
	firmwareRule, err := xcoreef.ConvertDownloadLocationFilterToFirmwareRule(filter)
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

func UpdateDownloadLocationRoundRobinFilter(applicationType string, filter *coreef.DownloadLocationRoundRobinFilterValue) *xwhttp.ResponseEntity {
	if err := xshared.ValidateApplicationType(applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}
	filter.ApplicationType = applicationType

	filter.ID = xcoreef.GetRoundRobinIdByApplication(applicationType)

	if err := filter.Validate(); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	err := estbfirmware.CreateDownloadLocationRoundRobinFilterValOneDB(filter)
	if err != nil {
		errorStr := fmt.Sprintf("Unable to save DownloadLocationRoundRobin: %s", err.Error())
		log.Error(errorStr)
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, errors.New(errorStr), nil)
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, filter)
}
