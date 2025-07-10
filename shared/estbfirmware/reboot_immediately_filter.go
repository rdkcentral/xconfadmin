/*
 * If not stated otherwise in this file or this component's Licenses.txt file the
 * following copyright and licenses apply:
 *
 * Copyright 2018 RDK Management
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
 * Author: cpatel550
 * Created: 07/28/2020
 */

package estbfirmware

import (
	coreef "xconfwebconfig/shared/estbfirmware"
)

func NewEmptyRebootImmediatelyFilter() *coreef.RebootImmediatelyFilter {
	return &coreef.RebootImmediatelyFilter{
		Environments: []string{},
		Models:       []string{},
	}
}

// type RebootImmediatelyFilter struct {
// 	IpAddressGroups []*core.IpAddressGroup `json:"ipAddressGroups,omitempty" xml:"ipAddressGroups"`
// 	Environments    []string               `json:"environments" xml:"environments"`
// 	Models          []string               `json:"models" xml:"models"`
// 	MacAddresses    string                 `json:"macAddresses,omitempty" xml:"macAddresses"`
// 	ID              string                 `json:"id" xml:"id"`
// 	Name            string                 `json:"name" xml:"name"`
// }

// func NewEmptyRebootImmediatelyFilter() *coreef.RebootImmediatelyFilter {
// 	return &coreef.RebootImmediatelyFilter{
// 		Environments: []string{},
// 		Models:       []string{},
// 	}
// }

// func RebootImmediatelyFiltersByApplicationType(applicationType string) ([]*RebootImmediatelyFilter, error) {
// 	rulelst, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_FIRMWARE_RULE, 0)
// 	if err != nil {
// 		return nil, err
// 	}

// 	filterRules := make([]*RebootImmediatelyFilter, 0)
// 	for _, rule := range rulelst {
// 		frule := rule.(*corefw.FirmwareRule)
// 		if frule.ApplicationType != applicationType {
// 			continue
// 		}
// 		if frule.GetTemplateId() != REBOOT_IMMEDIATELY_FILTER {
// 			continue
// 		}
// 		filter := ConvertFirmwareRuleToRebootFilter(frule)
// 		filterRules = append(filterRules, filter)
// 	}

// 	return filterRules, nil
// }

// func RebootImmediatelyFiltersByName(applicationType string, name string) (*RebootImmediatelyFilter, error) {
// 	rulelst, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_FIRMWARE_RULE, 0)
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, rule := range rulelst {
// 		frule := rule.(*corefw.FirmwareRule)
// 		if frule.ApplicationType != applicationType {
// 			continue
// 		}
// 		if frule.GetTemplateId() != REBOOT_IMMEDIATELY_FILTER {
// 			continue
// 		}
// 		if frule.Name == name {
// 			filter := ConvertFirmwareRuleToRebootFilter(frule)
// 			return filter, nil
// 		}
// 	}

// 	return nil, nil
// }
