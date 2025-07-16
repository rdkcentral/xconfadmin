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
	"sort"
	"strconv"
	"strings"

	"xconfadmin/util"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	ru "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
)

func GetEnvironment(id string) *shared.Environment {
	environment := shared.GetOneEnvironment(id)
	if environment != nil {
		return environment
	}
	return nil
}

func IsExistEnvironment(envId string) bool {
	if envId != "" {
		environment := shared.GetOneEnvironment(envId)
		return environment != nil
	}
	return false
}

func CreateEnvironment(environment *shared.Environment) *xwhttp.ResponseEntity {
	// Environment's ID (name) is stored in uppercase
	environment.ID = strings.ToUpper(strings.TrimSpace(environment.ID))

	err := environment.Validate()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	existedEnv := shared.GetOneEnvironment(environment.ID)
	if existedEnv != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New("Environment with "+environment.ID+" already exists"), nil)
	}

	env, err := shared.SetOneEnvironment(environment)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, env)
}

func UpdateEnvironment(environment *shared.Environment) *xwhttp.ResponseEntity {
	// Environment's ID (name) is stored in uppercase
	environment.ID = strings.ToUpper(strings.TrimSpace(environment.ID))

	err := environment.Validate()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	existedEnv := shared.GetOneEnvironment(environment.ID)
	if existedEnv == nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New("Environment with "+environment.ID+" does not exist"), nil)
	}

	env, err := shared.SetOneEnvironment(environment)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, env)
}

func DeleteEnvironment(id string) *xwhttp.ResponseEntity {
	usage, err := validateUsageForEnvironment(id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if usage != "" {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(usage), nil)
	}

	err = shared.DeleteOneEnvironment(id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

// Return usage info if Environment is used by a rule, empty string otherwise
func validateUsageForEnvironment(id string) (string, error) {
	ruleTables := []string{
		ds.TABLE_DCM_RULE,
		ds.TABLE_FIRMWARE_RULE,
		ds.TABLE_FIRMWARE_RULE_TEMPLATE,
		ds.TABLE_TELEMETRY_RULES,
		ds.TABLE_TELEMETRY_TWO_RULES,
		ds.TABLE_FEATURE_CONTROL_RULE,
		ds.TABLE_SETTING_RULES,
	}

	for _, tableName := range ruleTables {
		resultMap, err := ds.GetCachedSimpleDao().GetAllAsMap(tableName)
		if err != nil {
			return "", err
		}

		for _, v := range resultMap {
			xrule, ok := v.(ru.XRule)
			if !ok {
				return "", fmt.Errorf("Failed to assert %s as XRule type", tableName)
			}
			if ru.IsExistConditionByFreeArgAndFixedArg(xrule.GetRule(), coreef.RuleFactoryENV.GetName(), id) {
				return fmt.Sprintf("Environment %s is used by %s %s", id, xrule.GetRuleType(), xrule.GetName()), nil
			}
		}
	}

	return "", nil
}

func environmentGeneratePage(list []*shared.Environment, page int, pageSize int) (result []*shared.Environment) {
	leng := len(list)
	startIndex := page*pageSize - pageSize
	result = make([]*shared.Environment, 0)
	if page < 1 || startIndex > leng || pageSize < 1 {
		return result
	}
	lastIndex := leng
	if page*pageSize < len(list) {
		lastIndex = page * pageSize
	}

	return list[startIndex:lastIndex]
}

func EnvironmentRuleGeneratePageWithContext(evrules []*shared.Environment, contextMap map[string]string) (result []*shared.Environment, err error) {
	sort.Slice(evrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(evrules[i].ID), strings.ToLower(evrules[j].ID)) < 0
	})
	pageNum := 1
	numStr, okval := contextMap[cPercentageBeanPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cPercentageBeanPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, errors.New("pageNumber and pageSize should both be greater than zero")
	}
	return environmentGeneratePage(evrules, pageNum, pageSize), nil
}

func EnvironmentFilterByContext(searchContext map[string]string) []*shared.Environment {
	EnvironmentRuleList := []*shared.Environment{}
	environments := shared.GetAllEnvironmentList()
	if (len(environments)) == 0 {
		return EnvironmentRuleList
	}
	for _, env := range environments {
		if env == nil {
			continue
		}
		if id, ok := util.FindEntryInContext(searchContext, cEnvironmentID, false); ok {
			if !strings.Contains(strings.ToLower(env.ID), strings.ToLower(id)) {
				continue
			}
		}
		if dsc, ok := util.FindEntryInContext(searchContext, cEnvironmentDescription, false); ok {
			if !strings.Contains(strings.ToLower(env.Description), strings.ToLower(dsc)) {
				continue
			}
		}
		EnvironmentRuleList = append(EnvironmentRuleList, env)
	}
	return EnvironmentRuleList
}
