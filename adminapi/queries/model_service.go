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
	"strconv"
	"strings"

	xcommon "github.com/rdkcentral/xconfadmin/common"
	"github.com/rdkcentral/xconfadmin/util"
	"github.com/rdkcentral/xconfwebconfig/common"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	ru "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	xwutil "github.com/rdkcentral/xconfwebconfig/util"
)

const (
	cModelPageNumber           = xcommon.PAGE_NUMBER
	cModelPageSize             = xcommon.PAGE_SIZE
	cModelApplicableActionType = xcommon.APPLICABLE_ACTION_TYPE
	cModelDescription          = xwcommon.DESCRIPTION
	cModelID                   = xwcommon.ID
)

func GetModels(tenantId string) []*shared.ModelResponse {
	result := []*shared.ModelResponse{}
	models := shared.GetAllModelList(tenantId)
	for _, model := range models {
		resp := model.CreateModelResponse()
		result = append(result, resp)
	}
	return result
}

func GetModel(tenantId string, id string) *shared.ModelResponse {
	model := shared.GetOneModel(tenantId, id)
	if model != nil {
		return model.CreateModelResponse()
	}
	return nil
}

func IsExistModel(tenantId string, id string) bool {
	return id != "" && shared.GetOneModel(tenantId, id) != nil
}

func CreateModel(tenantId string, model *shared.Model) *xwhttp.ResponseEntity {
	// Model's ID (name) is stored in uppercase
	model.ID = strings.ToUpper(strings.TrimSpace(model.ID))

	err := model.Validate()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, model.ID)
	}

	existingModel := shared.GetOneModel(tenantId, model.ID)
	if existingModel != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New("\"Model with current name already exists\""), model.ID)

	}

	model.Updated = xwutil.GetTimestamp()
	env, err := shared.SetOneModel(tenantId, model)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, model)
	}

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, env)
}

func UpdateModel(tenantId string, model *shared.Model) *xwhttp.ResponseEntity {
	// Model's ID (name) is stored in uppercase
	model.ID = strings.ToUpper(strings.TrimSpace(model.ID))

	err := model.Validate()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, model)
	}

	existingModel := shared.GetOneModel(tenantId, model.ID)
	if existingModel == nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, errors.New(model.ID+" model does not exist"), model)
	}

	model.Updated = xwutil.GetTimestamp()
	env, err := shared.SetOneModel(tenantId, model)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, model)
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, env)
}

func DeleteModel(tenantId string, id string) *xcommon.ResponseEntity {
	err := validateUsageForModel(tenantId, id)
	if err != nil {
		return xcommon.NewResponseEntity(err, id)
	}
	//TODO - Use NewResponseEntity at either one place
	existingModel := shared.GetOneModel(tenantId, id)
	if existingModel == nil {
		return xcommon.NewResponseEntityWithStatus(http.StatusNotFound, errors.New("Entity with id: "+id+" does not exist"), id)
	}

	err = shared.DeleteOneModel(tenantId, id)
	if err != nil {
		return xcommon.NewResponseEntityWithStatus(http.StatusInternalServerError, err, id)
	}

	return xcommon.NewResponseEntityWithStatus(http.StatusNoContent, nil, id)
}

// Return usage info if Model is used by a rule, empty string otherwise
func validateUsageForModel(tenantId string, modelId string) error {
	// Check for usage in all Rules
	ruleTables := []string{
		db.TABLE_DCM_RULES,
		db.TABLE_FIRMWARE_RULE_TEMPLATES,
		db.TABLE_TELEMETRY_RULES,
		db.TABLE_TELEMETRY_TWO_RULES,
		db.TABLE_FEATURE_CONTROL_RULES,
		db.TABLE_SETTING_RULES,
		db.TABLE_FIRMWARE_RULES,
	}

	for _, tableName := range ruleTables {
		resultMap, err := db.GetCachedSimpleDao().GetAllAsMap(tenantId, tableName)
		if err != nil {
			return err
		}

		for _, v := range resultMap {
			xrule, ok := v.(ru.XRule)
			if !ok {
				return xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, fmt.Sprintf("Failed to assert %s as XRule type", tableName))
			}
			if ru.IsExistConditionByFreeArgAndFixedArg(xrule.GetRule(), coreef.RuleFactoryMODEL.GetName(), modelId) {
				return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Model %s is used by %s %s(%s)", modelId, xrule.GetRuleType(), xrule.GetName(), tableName))
			}
		}
	}

	// Check for usage in FirmwareConfig
	list, err := coreef.GetFirmwareConfigAsListDB(tenantId)
	if err != nil && err.Error() != common.NotFound.Error() {
		return xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, err.Error())
	}

	for _, config := range list {
		if config != nil {
			if xwutil.Contains(config.SupportedModelIds, modelId) {
				return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Model %s is used by FirmwareConfig %s", modelId, config.Description))
			}
		}
	}

	return nil
}

func extractModelPage(list []*shared.Model, page int, pageSize int) (result []*shared.Model) {
	leng := len(list)
	result = make([]*shared.Model, 0)
	startIndex := page*pageSize - pageSize
	if page < 1 || startIndex > leng || pageSize < 1 {
		return result
	}
	lastIndex := leng
	if page*pageSize < len(list) {
		lastIndex = page * pageSize
	}

	return list[startIndex:lastIndex]
}

func generateModelPageByContext(dbrules []*shared.Model, contextMap map[string]string) (result []*shared.Model, err error) {
	/*
		validContexts := []string{cModelPageNumber, cModelPageSize}
		for k := range contextMap {
			if searchList(validContexts, k, false) {
				continue
			}
			return nil, xwcommon.NewRemoteErrorAS (http.StatusBadRequest, "Inapplicable parameter: " + k)

		}
	*/
	pageNum := 1
	numStr, okval := contextMap[cModelPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cModelPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "pageNumber and pageSize should both be greater than zero")
	}
	return extractModelPage(dbrules, pageNum, pageSize), nil
}

func filterModelsByContext(entries []*shared.Model, searchContext map[string]string) (result []*shared.Model, err error) {
	/*
		validFilters := []string{xcommon.ID, shared.DESCRIPTION}
		for k := range searchContext {
			if searchList(validFilters, k, false) {
				continue
			}
			return nil, xwcommon.NewRemoteErrorAS (http.StatusBadRequest, "Invalid param " + k + ". Valid Params are: " + strings.Join(validFilters[:], ","))
		}
	*/

	for _, entry := range entries {
		if id, ok := util.FindEntryInContext(searchContext, xwcommon.ID, false); ok {
			if !strings.Contains(strings.ToLower(entry.ID), strings.ToLower(id)) {
				continue
			}
		}
		if description, ok := util.FindEntryInContext(searchContext, xwcommon.DESCRIPTION, false); ok {
			if !strings.Contains(strings.ToLower(entry.Description), strings.ToLower(description)) {
				continue
			}
		}
		result = append(result, entry)
	}
	return result, nil
}
