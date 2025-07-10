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

	xwhttp "xconfwebconfig/http"
	"xconfwebconfig/shared"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

func GetIpAddressGroups() []*shared.IpAddressGroup {
	result := []*shared.IpAddressGroup{}
	list, err := shared.GetGenericNamedListListsByTypeDB(shared.IP_LIST)
	if err != nil {
		log.Error(fmt.Sprintf("GetIpAddressGroups: %v", err))
		return result
	}
	for _, nl := range list {
		ipGrp := nl.CreateIpAddressGroupResponse()
		result = append(result, ipGrp)
	}
	return result
}

func GetIpAddressGroupByName(name string) *shared.IpAddressGroup {
	nl, err := shared.GetGenericNamedListOneDB(name)
	if err != nil {
		log.Error(fmt.Sprintf("GetIpAddressGroupByName: %v", err))
		return nil
	}

	if nl.TypeName != shared.IP_LIST {
		return nil
	}

	return nl.CreateIpAddressGroupResponse()
}

func GetIpAddressGroupsByIp(ip string) []*shared.IpAddressGroup {
	result := []*shared.IpAddressGroup{}
	list, err := shared.GetGenericNamedListListsByTypeDB(shared.IP_LIST)
	if err != nil {
		log.Error(fmt.Sprintf("GetIpAddressGroupByIp: %v", err))
		return result
	}
	for _, nl := range list {
		ipGrp := shared.NewIpAddressGroupWithAddrStrings(nl.ID, nl.ID, nl.Data)
		if ipGrp.IsInRange(ip) {
			ipGrp.RawIpAddresses = nl.Data // For the response, we need list of ip address as string
			result = append(result, ipGrp)
		}
	}
	return result
}

func CreateIpAddressGroup(ipAddressGroup *shared.IpAddressGroup) *xwhttp.ResponseEntity {
	ipList := shared.ConvertFromIpAddressGroup(ipAddressGroup)
	err := ipList.ValidateForAdminService()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	err = shared.CreateGenericNamedListOneDB(ipList)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}
	resp := ipList.CreateIpAddressGroupResponse()
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, resp)
}

func IsChangedIpAddressGroup(ipAddressGroup *shared.IpAddressGroup) bool {
	if ipAddressGroup != nil && !util.IsBlank(ipAddressGroup.Name) {
		existedIpAddressGroup := getIpAddressGroup(ipAddressGroup.Name)
		if existedIpAddressGroup != nil {
			s1 := []string{}
			for _, addr := range ipAddressGroup.RawIpAddresses {
				s1 = append(s1, addr)
			}
			s2 := []string{}
			for _, addr := range existedIpAddressGroup.RawIpAddresses {
				s2 = append(s2, addr)
			}
			return !util.StringElementsMatch(s1, s2)
		}
	}
	return true
}
