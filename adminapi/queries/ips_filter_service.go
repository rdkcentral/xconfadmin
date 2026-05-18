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
	"net/http"

	core "github.com/rdkcentral/xconfadmin/shared"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/util"
)

func UpdateIpFilter(applicationType string, ipFilter *coreef.IpFilter) *xwhttp.ResponseEntity {
	if err := firmware.ValidateRuleName(ipFilter.Id, ipFilter.Name, applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if ipFilter.IpAddressGroup != nil && IsChangedIpAddressGroup(ipFilter.IpAddressGroup) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest,
			fmt.Errorf("IP address group denoted by '%s' does not match any existing ipAddressGroup", ipFilter.IpAddressGroup.Name), nil)
	}

	firmwareRule := coreef.ConvertIpFilterToFirmwareRule(ipFilter)
	firmwareRule.ApplicableAction = firmware.NewApplicableActionAndType(firmware.BlockingFilterActionClass, firmware.BLOCKING_FILTER, "")
	if !util.IsBlank(applicationType) {
		firmwareRule.ApplicationType = applicationType
	}

	if err := core.ValidateApplicationType(firmwareRule.ApplicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	err := corefw.CreateFirmwareRuleOneDB(firmwareRule)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if ipFilter.Id == "" {
		ipFilter.Id = firmwareRule.ID
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, ipFilter)
}

func DeleteIpsFilter(name string, applicationType string) *xwhttp.ResponseEntity {
	ipFilter, err := coreef.IpFilterByName(name, applicationType)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if ipFilter != nil {
		err = corefw.DeleteOneFirmwareRule(ipFilter.Id)
		if err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}
