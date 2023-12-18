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
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"xconfwebconfig/dataapi/featurecontrol"
	ru "xconfwebconfig/rulesengine"

	xcommon "xconfadmin/common"
	xshared "xconfadmin/shared"

	xrfc "xconfadmin/shared/rfc"
	"xconfadmin/util"
	"xconfwebconfig/common"
	"xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	"xconfwebconfig/shared/rfc"
	xutil "xconfwebconfig/util"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func GetAllFeatureRulesByType(applicationType string) []*rfc.FeatureRule {
	ruleList := rfc.GetFeatureRuleList()

	featureRules := []*rfc.FeatureRule{}
	for _, featureRule := range ruleList {
		if featureRule == nil {
			continue
		}
		if featureRule.ApplicationType == applicationType {
			featureRules = append(featureRules, featureRule)
		}
	}
	return featureRules
}

func GetOne(id string) *rfc.FeatureRule {
	return xrfc.GetFeatureRule(id)
}

func FindFeatureRuleByContext(searchContext map[string]string) []*rfc.FeatureRule {
	featureRules := rfc.GetFeatureRuleList()
	sort.Slice(featureRules, func(i, j int) bool {
		if featureRules[i].Priority < featureRules[j].Priority {
			return true
		}
		if featureRules[i].Priority > featureRules[j].Priority {
			return false
		}
		return featureRules[i].Id < featureRules[j].Id
	})
	featureRuleList := []*rfc.FeatureRule{}
	for _, featureRule := range featureRules {
		if featureRule == nil {
			continue
		}
		if applicationType, ok := util.FindEntryInContext(searchContext, common.APPLICATION_TYPE, false); ok {
			if featureRule.ApplicationType != applicationType && featureRule.ApplicationType != shared.ALL {
				continue
			}
		}
		if featureInstance, ok := util.FindEntryInContext(searchContext, xcommon.FEATURE_INSTANCE, false); ok {
			if len(featureRule.FeatureIds) < 1 {
				continue
			}
			featureNameMatch := false
			for _, featureId := range featureRule.FeatureIds {
				feature := rfc.GetOneFeature(featureId)
				if feature != nil && strings.Contains(strings.ToLower(feature.FeatureName), strings.ToLower(featureInstance)) {
					featureNameMatch = true
					break
				}
			}
			if !featureNameMatch {
				continue
			}
		}
		if name, ok := util.FindEntryInContext(searchContext, xcommon.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(featureRule.Name), strings.ToLower(name)) {
				continue
			}
		}
		if key, ok := util.FindEntryInContext(searchContext, xcommon.FREE_ARG, false); ok {
			keyMatch := false
			for _, condition := range ru.ToConditions(featureRule.Rule) {
				if strings.Contains(strings.ToLower(condition.GetFreeArg().Name), strings.ToLower(key)) {
					keyMatch = true
					break
				}
			}
			if !keyMatch {
				continue
			}
		}
		if fixedArgValue, ok := util.FindEntryInContext(searchContext, xcommon.FIXED_ARG, false); ok {
			valueMatch := false
			for _, condition := range ru.ToConditions(featureRule.Rule) {
				if condition.GetFixedArg() != nil && condition.GetFixedArg().IsCollectionValue() {
					fixedArgs := condition.GetFixedArg().GetValue().([]string)
					for _, fixedArg := range fixedArgs {
						if strings.Contains(strings.ToLower(fixedArg), strings.ToLower(fixedArgValue)) {
							valueMatch = true
							break
						}
					}
				}
				if valueMatch {
					break
				}
				if condition.GetOperation() != rulesengine.StandardOperationExists && condition.GetFixedArg() != nil && condition.GetFixedArg().IsStringValue() {
					if strings.Contains(strings.ToLower(condition.FixedArg.Bean.Value.JLString), strings.ToLower(fixedArgValue)) {
						valueMatch = true
						break
					}
				}
			}
			if !valueMatch {
				continue
			}
		}
		featureRuleList = append(featureRuleList, featureRule)
	}
	return featureRuleList
}

func CreateFeatureRule(featureRule *rfc.FeatureRule, applicationType string) error {
	err := beforeCreating(featureRule)
	if err != nil {
		return err
	}
	err = beforeSaving(featureRule, applicationType)
	if err != nil {
		return err
	}
	contextMap := map[string]string{common.APPLICATION_TYPE: featureRule.ApplicationType}
	featureRules := addNewFeatureRuleAndReorganize(featureRule, FindFeatureRuleByContext(contextMap))
	for _, featureRule := range featureRules {
		xrfc.SetFeatureRule(featureRule.Id, featureRule)
	}
	return nil
}

func addNewFeatureRuleAndReorganize(newItem *rfc.FeatureRule, itemsList []*rfc.FeatureRule) []*rfc.FeatureRule {
	sort.Slice(itemsList, func(i, j int) bool {
		return itemsList[i].Priority < itemsList[j].Priority
	})
	itemsList = append(itemsList, newItem)
	return reorganizeFeatureRulePriorities(itemsList, len(itemsList), newItem.Priority)
}

func reorganizeFeatureRulePriorities(sortedItemsList []*rfc.FeatureRule, oldPriority int, newPriority int) []*rfc.FeatureRule {
	if newPriority < 1 || int(newPriority) > len(sortedItemsList) {
		newPriority = len(sortedItemsList)
	}
	item := sortedItemsList[oldPriority-1]
	item.Priority = newPriority
	if oldPriority < newPriority {
		for i := oldPriority; i <= newPriority-1; i++ {
			buf := sortedItemsList[i]
			buf.Priority = i
			sortedItemsList[i-1] = buf
		}
	}
	if oldPriority > newPriority {
		for i := oldPriority - 2; i >= newPriority-1; i-- {
			buf := sortedItemsList[i]
			buf.Priority = i + 2
			sortedItemsList[i+1] = buf
		}
	}
	sortedItemsList[newPriority-1] = item
	return getAlteredFeatureRuleSubList(sortedItemsList, oldPriority, newPriority)
}

func getAlteredFeatureRuleSubList(itemsList []*rfc.FeatureRule, oldPriority int, newPriority int) []*rfc.FeatureRule {
	start := int(math.Min(float64(oldPriority), float64(newPriority)) - float64(1))
	end := int(math.Max(float64(oldPriority), float64(newPriority)))
	return itemsList[start:end]
}

func beforeCreating(entity *rfc.FeatureRule) error {
	id := entity.Id
	if id == "" {
		entity.Id = uuid.New().String()
	} else {
		featureRule := GetOne(id)
		if featureRule != nil {
			return xcommon.NewXconfError(http.StatusConflict, "\"FeatureRule with id: "+id+" already exists\"")
		}
	}
	return nil
}

func beforeSaving(featureRule *rfc.FeatureRule, applicationType string) error {
	if featureRule == nil {
		return xcommon.NewXconfError(http.StatusBadRequest, "FeatureRule is empty")
	}
	if featureRule.ApplicationType == "" {
		featureRule.ApplicationType = applicationType
	} else {
		if err := xshared.ValidateApplicationType(featureRule.ApplicationType); err != nil {
			return err
		}
	}
	if featureRule.Rule != nil {
		ru.NormalizeConditions(featureRule.Rule)
	}
	err := ValidateFeatureRule(featureRule, applicationType)
	if err != nil {
		return err
	}
	err = validateAllFeatureRule(featureRule)
	if err != nil {
		return err
	}
	return nil
}

func ValidateFeatureRule(featureRule *rfc.FeatureRule, applicationType string) error {
	if featureRule == nil {
		return xcommon.NewXconfError(http.StatusBadRequest, "FeatureRule is empty")
	}
	if featureRule.Rule == nil {
		return xcommon.NewXconfError(http.StatusBadRequest, "Rule is empty")
	}
	err := ValidateRuleStructure(featureRule.Rule)
	if err != nil {
		return err
	}
	err = RunGlobalValidation(*featureRule.Rule, GetFeatureRuleAllowedOperations)
	if err != nil {
		return err
	}
	if featureRule.Name == "" {
		return xcommon.NewXconfError(http.StatusBadRequest, "FeatureRule name is blank")
	}
	if len(featureRule.FeatureIds) < 1 {
		return xcommon.NewXconfError(http.StatusBadRequest, "Features should be specified")
	} else if len(featureRule.FeatureIds) > xcommon.AllowedNumberOfFeatures {
		return xcommon.NewXconfError(http.StatusBadRequest, "Number of Features should be up to "+strconv.Itoa(xcommon.AllowedNumberOfFeatures)+" items")
	} else {
		for _, featureId := range featureRule.FeatureIds {
			feature := rfc.GetOneFeature(featureId)
			if feature == nil {
				return xcommon.NewXconfError(http.StatusNotFound, "Feature with id: "+featureId+" does not exist")
			}
			if feature.ApplicationType != featureRule.ApplicationType {
				return xcommon.NewXconfError(http.StatusBadRequest, "Application Mismatch of Feature and Feature Rule:")
			}
		}
	}

	if err := xshared.ValidateApplicationType(featureRule.ApplicationType); err != nil {
		return err
	}

	if !strings.EqualFold(featureRule.ApplicationType, applicationType) {
		return xcommon.NewXconfError(http.StatusBadRequest, "Current application type "+applicationType+" doesn't match with entity application type: "+featureRule.ApplicationType)
	}

	percentRanges, err := getPercentRanges(featureRule.Rule)
	if err != nil {
		return err
	}
	for _, percentRange := range percentRanges {
		//validateStartRange
		if percentRange.StartRange < 0 || percentRange.StartRange >= 100 {
			return xcommon.NewXconfError(http.StatusBadRequest, "Start range "+fmt.Sprint(percentRange.StartRange)+" is not valid")
		}
		//validateEndRange
		if percentRange.EndRange < 0 || percentRange.EndRange > 100 {
			return xcommon.NewXconfError(http.StatusBadRequest, "End range "+fmt.Sprint(percentRange.EndRange)+" is not valid")
		}
		//validateRanges
		if percentRange.StartRange >= percentRange.EndRange {
			return xcommon.NewXconfError(http.StatusBadRequest, "Start range should be less than end range")
		}
		//validateRangesOverlapping
		for _, percentRange_1 := range percentRanges {
			if !percentRange.Equals(&percentRange_1) && percentRange.StartRange <= percentRange_1.StartRange && percentRange_1.StartRange < percentRange.EndRange {
				return xcommon.NewXconfError(http.StatusBadRequest, "Ranges overlap each other")
			}
		}
	}
	return nil
}

func getPercentRanges(rule *rulesengine.Rule) ([]rfc.PercentRange, error) {
	percentRanges := []rfc.PercentRange{}
	for _, condition := range ru.ToConditions(rule) {
		if rulesengine.StandardOperationRange == condition.GetOperation() && condition.GetFixedArg() != nil && condition.GetFixedArg().IsStringValue() {
			percentRangeString := condition.FixedArg.Bean.Value.JLString
			percentRange, err := parsePercentRange(percentRangeString)
			if err != nil {
				return nil, err
			}
			percentRanges = append(percentRanges, *percentRange)
		}
	}
	sort.Slice(percentRanges, func(i, j int) bool {
		return percentRanges[i].StartRange < percentRanges[j].StartRange
	})
	return percentRanges, nil
}

func parsePercentRange(percentRange string) (*rfc.PercentRange, error) {
	splitRange := strings.Split(strings.Trim(percentRange, " "), "-")
	if len(splitRange) < 2 {
		return nil, xcommon.NewXconfError(http.StatusBadRequest, "Range format exception "+percentRange+", format pattern is: startRange-endRange")
	}
	convertedRange := rfc.PercentRange{}
	startRange, err := strconv.ParseFloat(splitRange[0], 8)
	if err != nil {
		return nil, xcommon.NewXconfError(http.StatusBadRequest, "Percent range "+percentRange+" is not valid")
	}
	endRange, err := strconv.ParseFloat(splitRange[1], 8)
	if err != nil {
		return nil, xcommon.NewXconfError(http.StatusBadRequest, "Percent range "+percentRange+" is not valid")
	}
	convertedRange.StartRange = startRange
	convertedRange.EndRange = endRange
	return &convertedRange, nil
}

func validateAllFeatureRule(ruleToCheck *rfc.FeatureRule) error {
	existingFeatureRules := rfc.GetFeatureRuleList()
	for _, featureRule := range existingFeatureRules {
		if featureRule.Id == ruleToCheck.Id {
			continue
		}
		if ruleToCheck.ApplicationType != featureRule.ApplicationType {
			continue
		}
		if ruleToCheck.GetName() == featureRule.GetName() {
			return xcommon.NewXconfError(http.StatusConflict, "\"Name is already used\"")
		}
		if ruleToCheck.GetRule().Equals(featureRule.GetRule()) {
			return xcommon.NewXconfError(http.StatusConflict, "Rule has duplicate:"+featureRule.GetName())
		}
	}
	return nil
}

func UpdateFeatureRule(featureRule *rfc.FeatureRule, applicationType string) error {
	if featureRule.Id == "" {
		return xcommon.NewXconfError(http.StatusBadRequest, "FeatureRule id is empty")
	}
	if err := beforeSaving(featureRule, applicationType); err != nil {
		return err
	}

	featureRuleToUpdate := GetOne(featureRule.Id)
	if featureRuleToUpdate == nil {
		return xcommon.NewXconfError(http.StatusNotFound, "FeatureRule with id: "+featureRule.Id+" does not exist")
	}
	if featureRuleToUpdate.ApplicationType != featureRule.ApplicationType {
		return xcommon.NewXconfError(http.StatusConflict, "ApplicationType cannot be changed: Existing value:"+featureRuleToUpdate.ApplicationType+" New Value:"+featureRule.ApplicationType)
	}

	contextMap := map[string]string{common.APPLICATION_TYPE: featureRule.ApplicationType}
	featureRules := updateFeatureRuleByPriorityAndReorganize(featureRule, FindFeatureRuleByContext(contextMap), featureRuleToUpdate.Priority)
	for _, featureRule := range featureRules {
		xrfc.SetFeatureRule(featureRule.Id, featureRule)
	}
	return nil
}

func updateFeatureRuleByPriorityAndReorganize(newItem *rfc.FeatureRule, itemsList []*rfc.FeatureRule, priority int) []*rfc.FeatureRule {
	sort.Slice(itemsList, func(i, j int) bool {
		return itemsList[i].Priority < itemsList[j].Priority
	})
	if len(itemsList) > 0 {
		for i, item := range itemsList {
			if item.Id == newItem.Id {
				itemsList[i] = newItem
				break
			}
		}
	} else {
		itemsList = append(itemsList, newItem)
	}
	return reorganizeFeatureRulePriorities(itemsList, priority, newItem.Priority)
}

func ImportOrUpdateAllFeatureRule(featureRuleList []rfc.FeatureRule, applicationType string) map[string][]string {
	importResult := make(map[string][]string, 2)
	imported := []string{}
	notImported := []string{}
	var err error
	for _, featureRule := range featureRuleList {
		if featureRule.Id != "" {
			err = CreateFeatureRule(&featureRule, applicationType)
		} else {
			if featureRuleDB := GetOne(featureRule.Id); featureRuleDB != nil {
				err = UpdateFeatureRule(&featureRule, applicationType)
			} else {
				err = CreateFeatureRule(&featureRule, applicationType)
			}
		}
		if err == nil {
			imported = append(imported, featureRule.Id)
		} else {
			b, err := xutil.JSONMarshal(featureRule)
			if err != nil {
				log.Error(fmt.Println(err))
			} else {
				log.Error("Exception to import: " + string(b))
			}
			notImported = append(notImported, featureRule.Id)
		}
	}
	importResult[IMPORTED] = imported
	importResult[NOT_IMPORTED] = notImported
	return importResult
}

func ChangeFeatureRulePriorities(featureRuleId string, newPriority int, applicationType string) ([]*rfc.FeatureRule, error) {
	featureRuleToUpdate := GetOne(featureRuleId)
	if featureRuleToUpdate == nil {
		return nil, xcommon.NewXconfError(http.StatusNotFound, "FeatureRule with id: "+featureRuleId+" does not exist")
	}
	oldPriority := featureRuleToUpdate.Priority
	featureRuleList := rfc.GetFeatureRuleList()
	featureRuleListForApplicationType := []*rfc.FeatureRule{}
	if applicationType != "" {
		for _, featureRule := range featureRuleList {
			if featureRule.ApplicationType == applicationType {
				featureRuleListForApplicationType = append(featureRuleListForApplicationType, featureRule)
			}
		}
	} else {
		featureRuleListForApplicationType = featureRuleList
	}
	reorganizedFeatureRules := UpdateFeatureRulePriorities(featureRuleListForApplicationType, oldPriority, newPriority)
	for _, featureRule := range reorganizedFeatureRules {
		xrfc.SetFeatureRule(featureRule.Id, featureRule)
	}
	log.Info("Priority of FeatureRule " + featureRuleId + " has been changed, oldPriority=" + strconv.Itoa(oldPriority) + ", newPriority=" + strconv.Itoa(newPriority))
	return reorganizedFeatureRules, nil
}

func UpdateFeatureRulePriorities(itemsList []*rfc.FeatureRule, oldPriority int, newPriority int) []*rfc.FeatureRule {
	sort.Slice(itemsList, func(i, j int) bool {
		return itemsList[i].Priority < itemsList[j].Priority
	})
	return reorganizeFeatureRulePriorities(itemsList, oldPriority, newPriority)
}

func GetAllowedNumberOfFeatures() int {
	return xcommon.AllowedNumberOfFeatures
}

func GetFeatureRulesSize(appType string) int {
	featureRuleList := rfc.GetFeatureRuleList()
	cnt := 0
	for _, entry := range featureRuleList {
		if entry.ApplicationType == appType {
			cnt++
		}
	}
	return cnt
}

func ProcessFeatureRules(context map[string]string, fields log.Fields) map[string]interface{} {
	result := make(map[string]interface{})

	featureControlRuleBase := featurecontrol.NewFeatureControlRuleBase()
	matchedRules := featureControlRuleBase.ProcessFeatureRules(context, context[common.APPLICATION_TYPE])
	if len(matchedRules) > 0 {
		result["result"] = map[string]interface{}{"": matchedRules}
	} else {
		result["result"] = nil
	}
	featureControl := featureControlRuleBase.Eval(context, context[common.APPLICATION_TYPE], fields)
	result["featureControl"] = featureControl
	result["context"] = context

	return result
}
