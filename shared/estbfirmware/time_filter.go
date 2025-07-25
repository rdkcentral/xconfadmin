// Copyright 2025 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package estbfirmware

import (
	corefw "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
)

// type TimeFilter struct {
// 	Id                        string               `json:"id" xml:"id"`
// 	Name                      string               `json:"name" xml:"name"`
// 	IpWhiteList               *core.IpAddressGroup `json:"ipWhitelist" xml:"ipWhitelist"`
// 	EnvModelWhitelist         EnvModelRuleBean     `json:"envModelWhitelist" xml:"envModelWhitelist"`
// 	NeverBlockRebootDecoupled bool                 `json:"neverBlockRebootDecoupled" xml:"neverBlockRebootDecoupled"`
// 	NeverBlockHttpDownload    bool                 `json:"neverBlockHttpDownload" xml:"neverBlockHttpDownload"`
// 	Start                     string               `json:"startTime" xml:"startTime"`
// 	End                       string               `json:"endTime" xml:"endTime"`
// 	LocalTime                 bool                 `json:"localTime" xml:"localTime"`
// }

// type EnvModelRuleBean struct {
// 	Id             string          `json:"id,omitempty" xml:"id,omitempty"`
// 	Name           string          `json:"name,omitempty" xml:"name,omitempty"`
// 	EnvironmentId  string          `json:"environmentId,omitempty" xml:"environmentId,omitempty"`
// 	ModelId        string          `json:"modelId,omitempty" xml:"modelId,omitempty"`
// 	FirmwareConfig *FirmwareConfig `json:"firmwareConfig,omitempty" xml:"firmwareConfig,omitempty"`
// 	Noop           bool            `json:"-"`
// }

func NewEmptyTimeFilter() *corefw.TimeFilter {
	return &corefw.TimeFilter{}
}

// func TimeFiltersByApplicationType(applicationType string) ([]*TimeFilter, error) {
// 	rulelst, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_FIRMWARE_RULE, 0)
// 	if err != nil {
// 		return nil, err
// 	}

// 	filterRules := make([]*TimeFilter, 0)
// 	for _, rule := range rulelst {
// 		frule := rule.(*corefw.FirmwareRule)
// 		if frule.ApplicationType != applicationType {
// 			continue
// 		}
// 		if frule.GetTemplateId() != TIME_FILTER {
// 			continue
// 		}
// 		filter := &TimeFilter{
// 			Id:                        frule.ID,
// 			Name:                      frule.Name,
// 			NeverBlockRebootDecoupled: false,
// 			NeverBlockHttpDownload:    false,
// 		}
// 		convertConditionsForTimeFilter(frule, filter)
// 		filterRules = append(filterRules, filter)
// 	}
// 	return filterRules, nil
// }

// func TimeFilterByName(name string, applicationType string) (*TimeFilter, error) {
// 	rules, _ := TimeFiltersByApplicationType(applicationType)
// 	for _, rule := range rules {
// 		if rule.Name == name {
// 			return rule, nil
// 		}
// 	}
// 	return nil, nil
// }

// func convertConditionsForTimeFilter(rule *corefw.FirmwareRule, timefilter *TimeFilter) {
// 	for _, r := range rule.Rule.CompoundParts {
// 		cond := r.Condition
// 		fAName := cond.GetFreeArg().Name
// 		operation := cond.GetOperation()

// 		if cond != nil {
// 			if RuleFactoryREBOOT_DECOUPLED.Name == fAName {
// 				if re.StandardOperationExists == operation {
// 					timefilter.NeverBlockRebootDecoupled = true
// 				}
// 			} else if RuleFactoryFIRMWARE_DOWNLOAD_PROTOCOL.Name == fAName {
// 				if re.StandardOperationIs == operation {
// 					timefilter.NeverBlockHttpDownload = true
// 				}
// 			} else if IsLegacyIpFreeArg(cond.GetFreeArg()) || RuleFactoryIP.Name == fAName {
// 				timefilter.IpWhiteList = GetIpAddressGroup(cond)
// 			} else if RuleFactoryMODEL.Name == fAName {
// 				timefilter.EnvModelWhitelist.ModelId = trimSingleQuote(cond.GetFixedArg().String())
// 			} else if RuleFactoryENV.Name == fAName {
// 				timefilter.EnvModelWhitelist.EnvironmentId = trimSingleQuote(cond.GetFixedArg().String())
// 			} else if IsLegacyLocalTimeFreeArg(*cond.GetFreeArg()) || RuleFactoryLOCAL_TIME.Name == fAName {
// 				if re.StandardOperationGte == operation {
// 					timefilter.Start = parseStringTime(cond.GetFixedArg().String())
// 				} else if re.StandardOperationLte == operation {
// 					timefilter.End = parseStringTime(cond.GetFixedArg().String())
// 				}
// 			} else if RuleFactoryTIME_ZONE.Name == fAName {
// 				timefilter.LocalTime = r.IsNegated()
// 			}
// 		}
// 	}
// }

// func parseStringTime(t string) string {
// 	tmp := strings.ReplaceAll(t, "'", "")
// 	sTime := strings.Split(tmp, ":")
// 	return fmt.Sprintf("%s:%s", sTime[0], sTime[1])
// }

// func trimSingleQuote(str string) string {
// 	return strings.ReplaceAll(str, "'", "")
// }
