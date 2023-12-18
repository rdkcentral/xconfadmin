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
	"net/http"
	"regexp"
	"strconv"
	"strings"

	xcommon "xconfadmin/common"

	xwcommon "xconfwebconfig/common"
	"xconfwebconfig/rulesengine"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	logupload "xconfwebconfig/shared/logupload"
	util "xconfwebconfig/util"
)

var allowedOperations = []string{
	re.StandardOperationIs,
	re.StandardOperationLike,
	re.StandardOperationExists,
	re.StandardOperationPercent,
	re.StandardOperationInList,
}

var firmwareRuleAllowedOperations = []string{
	re.StandardOperationIs,
	re.StandardOperationLike,
	re.StandardOperationExists,
	re.StandardOperationPercent,
	re.StandardOperationInList,
	re.StandardOperationIn,
	re.StandardOperationGte,
	re.StandardOperationLte,
	re.StandardOperationMatch,
}

var featureRuleAllowedOperations = []string{
	re.StandardOperationIs,
	re.StandardOperationLike,
	re.StandardOperationExists,
	re.StandardOperationPercent,
	re.StandardOperationInList,
	re.StandardOperationIn,
	re.StandardOperationGte,
	re.StandardOperationLte,
	re.StandardOperationMatch,
	rulesengine.StandardOperationRange,
}

func equalFreeArgNames(freeArg string, freeArg2 string) bool {
	return freeArg == freeArg2
}

func isNotBlank(str string) bool {
	return !util.IsBlank(str)
}

func RunGlobalValidation(rule re.Rule, fp func() []string) error {
	conditions := re.ToConditions(&rule)
	if len(conditions) == 0 {
		return xcommon.NewXconfError(http.StatusBadRequest, "Rule is empty")
	}
	err := validateRelation(&rule)
	if err != nil {
		return err
	}
	err = checkDuplicateConditions(&rule)
	if err != nil {
		return xcommon.NewXconfError(http.StatusBadRequest, "Duplicate Conditions present")
	}
	for _, condition := range conditions {
		err := checkConditionNullsOrBlanks(*condition)
		if err != nil {
			return err
		}
		err = checkDuplicateFixedArgListItems(*condition)
		if err != nil {
			return err
		}
		err = checkOperationName(condition, fp)
		if err != nil {
			return err
		}
		err = checkOperationName(condition, fp)
		if err != nil {
			return err
		}
		freeArg := condition.GetFreeArg()
		if equalFreeArgNames(xcommon.STB_ESTB_MAC, freeArg.GetName()) ||
			equalFreeArgNames(logupload.EstbMacAddress, freeArg.Name) ||
			equalFreeArgNames(logupload.EcmMac, freeArg.Name) {
			err = checkFixedArgValue(*condition, util.IsValidMacAddress)
			if err != nil {
				return err
			}
		} else if equalFreeArgNames(xwcommon.IP_ADDRESS, freeArg.GetName()) ||
			equalFreeArgNames(logupload.EstbIp, freeArg.Name) {
			err = checkFixedArgValue(*condition, shared.IsValidIpAddress)
			if err != nil {
				return err
			}
		} else {
			err = checkFixedArgValue(*condition, isNotBlank)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func checkConditionNullsOrBlanks(condition re.Condition) error {
	freeArg := condition.GetFreeArg()
	if freeArg == nil || util.IsBlank(freeArg.GetName()) {
		return xcommon.NewXconfError(http.StatusBadRequest, "FreeArg is empty")
	}

	operation := condition.GetOperation()
	if util.IsBlank(operation) {
		return xcommon.NewXconfError(http.StatusBadRequest, "Operation is null")
	}

	fixedArg := condition.GetFixedArg()
	if re.StandardOperationExists != operation {
		if fixedArg == nil || fixedArg.GetValue() == nil {
			return xcommon.NewXconfError(http.StatusBadRequest, "FixedArg is null")
		}

		if re.StandardOperationIn == operation {
			if !fixedArg.IsCollectionValue() {
				return xcommon.NewXconfError(http.StatusBadRequest, freeArg.GetName()+" is not collection")
			}
			if len(fixedArg.Collection.Value) == 0 {
				return xcommon.NewXconfError(http.StatusBadRequest, freeArg.GetName()+" is empty")
			}
		} else {
			if fixedArg.GetValue() == "" {
				return xcommon.NewXconfError(http.StatusBadRequest, freeArg.GetName()+" is empty")
			}
		}
	}
	return nil
}

func checkDuplicateFixedArgListItems(condition re.Condition) error {
	fixedArg := condition.GetFixedArg()
	if fixedArg != nil && fixedArg.GetValue() != nil {
		duplicateFixedArgListItems := re.GetDuplicateFixedArgListItems(*fixedArg)
		if len(duplicateFixedArgListItems) > 0 {
			return xcommon.NewXconfError(http.StatusBadRequest, "FixedArg of condition contains duplicate items: ")
		}
	}
	return nil
}

func checkFixedArgValue(condition re.Condition, fp func(string) bool) error {
	operation := condition.GetOperation()
	if re.StandardOperationIn == operation {
		fixedArgValues := condition.GetFixedArg().Collection.Value
		for _, val := range fixedArgValues {
			if !fp(val) {
				return xcommon.NewXconfError(http.StatusBadRequest, "Incorrect Collection Value")
			}
		}
	} else if re.StandardOperationIs == operation {
		fixedArgValue := condition.GetFixedArg().GetValue().(string)
		//fixedArgValue := coreef.trimSingleQuote (condition.GetFixedArg().String())
		if !fp(fixedArgValue) {
			return xcommon.NewXconfError(http.StatusBadRequest, condition.FreeArg.GetName()+" is invalid: "+fixedArgValue)
		}
	} else if re.StandardOperationPercent == operation {
		ret, err := checkPercentOperation(condition)
		if err != nil {
			return err
		}
		if !ret {
			return xcommon.NewXconfError(http.StatusBadRequest, "Invalid value for operations")
		}
		return nil
	} else if re.StandardOperationLike == operation {
		return checkLikeOperation(condition)
	}
	return nil
}

func checkPercentOperation(condition re.Condition) (bool, error) {
	if condition.GetFixedArg().IsDoubleValue() {
		fixedArgDouble := condition.GetFixedArg().Bean.Value.JLDouble
		if fixedArgDouble >= 0.0 && fixedArgDouble <= 100.0 {
			return true, nil
		}
	} else if condition.GetFixedArg().IsStringValue() {
		const bitSize = 64
		fixedArgDouble, _ := strconv.ParseFloat(condition.GetFixedArg().Bean.Value.JLString, bitSize)
		if fixedArgDouble >= 0.0 && fixedArgDouble <= 100.0 {
			return true, nil
		}
	}
	return false, xcommon.NewXconfError(http.StatusBadRequest, "Invalid value for percent "+condition.GetFixedArg().String())
}

func checkLikeOperation(condition re.Condition) error {
	_, err := regexp.Compile(condition.GetFixedArg().GetValue().(string))
	return err
}

func equalTypes(type1 string, type2 string) bool {
	return strings.EqualFold(type1, type2)
}

func assertDuplicateConditions(duplicateConditions []re.Condition) error {
	if len(duplicateConditions) > 0 {
		return xcommon.NewXconfError(http.StatusBadRequest, ": Duplicate conditions present ")
	}
	return nil
}

func checkDuplicateConditions(rule *re.Rule) error {
	err := assertDuplicateConditions(re.GetDuplicateConditionsFromRule(*rule))
	if err != nil {
		return err
	}
	return assertDuplicateConditions(re.GetDuplicateConditionsBetweenOR(*rule))
}

func equalOperations(op1 string, op2 string) bool {
	return strings.ToUpper(op1) == strings.ToUpper(op2)
}

func checkOperationName(cond *re.Condition, fp func() []string) error {
	isExists := false
	ruleOperation := cond.GetOperation()
	for _, operation := range fp() {
		if equalOperations(operation, ruleOperation) {
			isExists = true
			break
		}
	}

	if !isExists {
		return xcommon.NewXconfError(http.StatusBadRequest, "Operation is not valid: "+string(ruleOperation))
	}
	return nil
}

func GetAllowedOperations() []string {
	return allowedOperations
}

func GetFirmwareRuleAllowedOperations() []string {
	return firmwareRuleAllowedOperations
}

func GetFeatureRuleAllowedOperations() []string {
	return featureRuleAllowedOperations
}

func validateCompoundPartsTree(rule *re.Rule) error {
	if len(rule.GetCompoundParts()) == 0 {
		return nil
	}
	for _, compoundPart := range rule.GetCompoundParts() {
		if len(compoundPart.GetCompoundParts()) > 0 {
			return xcommon.NewXconfError(http.StatusBadRequest, "CompoundPart rule should not have one more compoundParts")
		}
	}
	return nil
}

func ValidateRuleStructure(rule *re.Rule) error {
	if !rule.IsCompound() && len(rule.GetCompoundParts()) > 0 {
		return xcommon.NewXconfError(http.StatusBadRequest, "rule should have only condition or compoundParts field")
	}
	return validateCompoundPartsTree(rule)
}

func validateRelation(rule *re.Rule) error {
	if rule.IsCompound() && len(rule.GetCompoundParts()) > 0 {
		for i, compoundPart := range rule.GetCompoundParts() {
			if i == 0 {
				continue
			}
			if compoundPart.Relation == "" {
				return xcommon.NewXconfError(http.StatusBadRequest, "Relation of "+compoundPart.Condition.FreeArg.Name+" is empty")
			}
		}
	}
	return nil
}
