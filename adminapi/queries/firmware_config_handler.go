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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	xcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/rdkcentral/xconfadmin/common"
	xutil "github.com/rdkcentral/xconfadmin/util"

	estb "github.com/rdkcentral/xconfadmin/shared/estbfirmware"

	"github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xhttp "github.com/rdkcentral/xconfadmin/http"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Zero api-usage in green splunk for 4 weeks ending 21 Oct 2021
// POST /xconfadminService/ux/api/firmwareconfig
func PostFirmwareConfigHandler(w http.ResponseWriter, r *http.Request) {
	firmwareConfig := estbfirmware.NewEmptyFirmwareConfig()
	firmwareConfig.ApplicationType = ""
	applicationType, err := auth.ExtractBodyAndCheckPermissions(firmwareConfig, w, r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	status := http.StatusCreated
	respEntity := CreateFirmwareConfigAS(firmwareConfig, applicationType, true)
	data := respEntity.Data
	status = respEntity.Status
	err = respEntity.Error

	if err != nil {
		xhttp.WriteAdminErrorResponse(w, status, err.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, status, res)
}

// Zero usages in green splunk for 4 weeks ending 21 Oct 2021
// PUT /xconfadminService/ux/api/firmwareconfig  0
func PutFirmwareConfigHandler(w http.ResponseWriter, r *http.Request) {
	firmwareConfig := estbfirmware.NewEmptyFirmwareConfig()
	firmwareConfig.ApplicationType = ""
	appType, err := auth.ExtractBodyAndCheckPermissions(firmwareConfig, w, r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	status := http.StatusOK
	respEntity := UpdateFirmwareConfigAS(firmwareConfig, appType, true)
	data := respEntity.Data
	status = respEntity.Status
	err = respEntity.Error
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, status, err.Error())
		return
	}
	res, err := xhttp.ReturnJsonResponse(data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, status, res)
}

// Zero usages in green splunk for 4 weeks ending 21 Oct 2021
// DELETE /xconfadminService/ux/api/firmwareconfig/{id}
func DeleteFirmwareConfigByIdHandler(w http.ResponseWriter, r *http.Request) {
	DeleteFirmwareConfigHandlerASFlavor(w, r)
}

func duplicateFcFoundOnDB(entity *estbfirmware.FirmwareConfig, descMap map[string][]*estbfirmware.FirmwareConfig) error {
	items, found := descMap[entity.Description]
	if found {
		for _, item := range items {
			if item.ID == entity.ID || item.ApplicationType != entity.ApplicationType {
				continue
			}
			return errors.New("Description " + entity.Description + " is already used in " + item.ID)
		}
	}
	return nil
}

func findAndDeleteFc(list []*estbfirmware.FirmwareConfig, item *estbfirmware.FirmwareConfig) []*estbfirmware.FirmwareConfig {
	index := 0
	for _, i := range list {
		if i.ID != item.ID {
			list[index] = i
			index++
		}
	}
	return list[:index]
}

// Zero usages in green splunk for 4 weeks ending 21 Oct 2021
// POST /xconfadminService/ux/api/firmwareconfig/entities
func PostFirmwareConfigEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	PutPostFirmwareConfigEntitiesHandler(w, r, false)
}

func PutPostFirmwareConfigEntitiesHandler(w http.ResponseWriter, r *http.Request, isPut bool) {
	appType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	body := xw.Body()
	entities := []estbfirmware.FirmwareConfig{}
	err = json.Unmarshal([]byte(body), &entities)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	descMap := make(map[string][]*estbfirmware.FirmwareConfig)
	list, err := estbfirmware.GetFirmwareConfigAsListDB()
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, item := range list {
		descMap[item.Description] = append(descMap[item.Description], item)
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for i, entity := range entities {
		_, err := estbfirmware.GetFirmwareConfigOneDB(entity.ID)
		if isPut && err != nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  common.ENTITY_STATUS_FAILURE,
				Message: "FirmwareConfig with " + entity.ID + " not present",
			}
			continue
		}
		if !isPut && err == nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  common.ENTITY_STATUS_FAILURE,
				Message: "FirmwareConfig with " + entity.ID + " already present",
			}
			continue
		}

		if entity.ApplicationType != appType {
			if util.IsBlank(entity.ID) {
				entity.ID = uuid.New().String() + entity.Description
			}
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  common.ENTITY_STATUS_FAILURE,
				Message: "ApplicationType conflict. User specified: " + appType + " , Instance applicationType " + entity.ApplicationType,
			}
			continue
		}

		if err = duplicateFcFoundOnDB(&entity, descMap); err != nil {
			if util.IsBlank(entity.ID) {
				entity.ID = uuid.New().String() + entity.Description
			}
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  common.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
			continue
		}
		var err2 *xwhttp.ResponseEntity
		entity := entity
		if isPut {
			err2 = UpdateFirmwareConfigAS(&entity, appType, false)
		} else {
			err2 = CreateFirmwareConfigAS(&entity, appType, false)
		}

		if err2.Error != nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  common.ENTITY_STATUS_FAILURE,
				Message: err2.Error.Error(),
			}
			continue
		}
		descMap[entity.Description] = findAndDeleteFc(descMap[entity.Description], &entity)
		descMap[entity.Description] = append(descMap[entity.Description], &entities[i])
		entitiesMap[entity.ID] = xhttp.EntityMessage{
			Status:  common.ENTITY_STATUS_SUCCESS,
			Message: entity.ID,
		}
	}
	response, err := xhttp.ReturnJsonResponse(entitiesMap, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

// Zero usages in green splunk for 4 weeks ending 21 Oct 2021
// PUT /xconfadminService/ux/api/firmwareconfig/entities
func PutFirmwareConfigEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	PutPostFirmwareConfigEntitiesHandler(w, r, true)
}

// Zero usages in green splunk for 4 weeks ending 21 Oct 2021
// GET /xconfadminService/ux/api/firmwareconfig/page
func ObsoleteGetFirmwareConfigPageHandler(w http.ResponseWriter, r *http.Request) {
	dbrules, _ := estbfirmware.GetFirmwareConfigAsListDB()
	sort.Slice(dbrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(dbrules[i].Description), strings.ToLower(dbrules[j].Description)) < 0
	})

	contextMap := map[string]string{}
	xutil.AddQueryParamsToContextMap(r, contextMap)

	var err error
	dbrules, err = generateFirmwareConfigPageByContext(dbrules, contextMap)
	allItemsLen := len(dbrules)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := xhttp.ReturnJsonResponse(dbrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	headerMap := populateHeaderWithNumberOfItems(allItemsLen)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func hasCommonEntries(list1 []string, list2 []string) bool {
	for _, v1 := range list1 {
		for _, v2 := range list2 {
			if v1 == v2 {
				return true
			}
		}
	}
	return false
}

// 7 api-usage in green splunk for 4 weeks ending 21 Oct 2021
// POST /xconfadminService/ux/api/firmwareconfig/bySupportedModels
func PostFirmwareConfigBySupportedModelsHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}

	body := xw.Body()
	modelIds := []string{}
	if err := json.Unmarshal([]byte(body), &modelIds); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract modelIds fromm Body: ")
		return
	}

	result := GetFirmwareConfigsByModelIdsAndApplication(modelIds, appType)
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

//	1447 usages in green splunk for 4 weeks ending 21 Oct 2021
//
// GET /xconfadminService/ux/api/firmwareconfig/firmwareConfigMap
func GetFirmwareConfigFirmwareConfigMapHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	configMap, err := estb.GetFirmwareConfigAsMapDB(appType)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	res, err := xhttp.ReturnJsonResponse(configMap, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

type FirmwareConfigData struct {
	Versions []string `json:"firmwareVersions"`
	ModelSet []string `json:"models"`
}

// 413 usages in green splunk for 4 weeks ending 21 Oct 2021
// POST /xconfadminService/ux/api/firmwareconfig/getSortedFirmwareVersionsIfExistOrNot
func PostFirmwareConfigGetSortedFirmwareVersionsIfExistOrNotHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}

	body := xw.Body()
	fcData := FirmwareConfigData{}

	if err := json.Unmarshal([]byte(body), &fcData); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to unmarshall body: ")
		return
	}

	result := GetSortedFirmwareVersionsIfDoesExistOrNot(fcData, appType)

	response, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

//	112 usages in green splunk for 4 weeks ending 21 Oct 2021
//
// GET /xconfadminService/ux/api/firmwareconfig/model/{modelId}
func GetFirmwareConfigModelByModelIdHandler(w http.ResponseWriter, r *http.Request) {
	GetQueriesFirmwareConfigsByModelIdASFlavor(w, r)
}

func searchList(stringList []string, k string, caseSensitive bool) bool {
	for _, v := range stringList {
		if caseSensitive {
			if k == v {
				return true
			}
		} else {
			if strings.EqualFold(v, k) {
				return true
			}
		}

	}
	return false
}

// 1802 usages in green splunk for 4 weeks ending 21 Oct 2021
// POST /xconfadminService/ux/api/firmwareconfig/filtered?pageSize=X&pageNumber=Y
func PostFirmwareConfigFilteredHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// Figure out the pageContext from Query
	pageContext := map[string]string{}
	xutil.AddQueryParamsToContextMap(r, pageContext)

	// Figure out the filterContext from Body
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}

	filterContext := make(map[string]string)
	body := xw.Body()
	if body != "" {
		err = json.Unmarshal([]byte(body), &filterContext)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	filterContext[xcommon.APPLICATION_TYPE] = appType

	// Get all entries and sort them
	entries, _ := estbfirmware.GetFirmwareConfigAsListDB()
	sort.Slice(entries, func(i, j int) bool {
		return strings.Compare(strings.ToLower(entries[i].Description), strings.ToLower(entries[j].Description)) < 0
	})

	// Filter entries according to filterContext
	entries, err = filterFirmwareConfigsByContext(entries, filterContext)
	allItemsLen := len(entries)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the entries from the requested  page as per pageContext
	entries, err = generateFirmwareConfigPageByContext(entries, pageContext)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// return the response
	response, err := xhttp.ReturnJsonResponse(entries, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	headerMap := populateHeaderWithNumberOfItems(allItemsLen)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

// 663 usagess in green splunk for 4 weeks ending 21 Oct 2021
// GET /xconfadminService/ux/api/firmwareconfig/{id}
func GetFirmwareConfigByIdHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%s is invalid.", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	fc, _ := estbfirmware.GetFirmwareConfigOneDB(id)
	if fc == nil {
		errorStr := fmt.Sprintf("Entity with id: %s does not exist", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	if fc.ApplicationType != appType {
		errorStr := fmt.Sprintf("ApplicationType mismatch: %s on db. %s provided", fc.ApplicationType, appType)
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, errorStr)
		return
	}
	fcList := []estbfirmware.FirmwareConfig{*fc}
	res, err := xhttp.ReturnJsonResponse(fcList, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	queryParams := r.URL.Query()
	_, ok := queryParams[common.EXPORT]
	if ok {
		fileName := common.ExportFileNames_FIRMWARE_CONFIG + fc.Description + "_" + appType
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
		return
	}
	GetQueriesFirmwareConfigsByIdASFlavor(w, r)

}

// 4687 usages in green splunk for 4 weeks ending 21 Oct 2021
// GET /xconfadminService/ux/api/firmwareconfig
func GetFirmwareConfigHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	queryParams := r.URL.Query()
	_, ok1 := queryParams[common.EXPORT]
	_, ok2 := queryParams[common.EXPORTALL]

	if ok1 || ok2 {
		entries := GetFirmwareConfigsAS(appType)

		res, err := xhttp.ReturnJsonResponse(entries, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}

		headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_ALL_FIRMWARE_CONFIGS + "_" + appType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
		return
	}
	getQueriesFirmwareConfigsASFlavor(w, r, appType)
}

// Zero usages in green splunk for 4 weeks ending 21 Oct 2021
// GET  /xconfadminService/ux/api/firmwareconfig/supportedConfigsByEnvModelRuleName/{ruleName}
func GetSupportedConfigsByEnvModelRuleName(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	ruleName, found := mux.Vars(r)[common.RULE_NAME]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.RULE_NAME)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	fwConfig := getSupportedConfigsByEnvModelRuleName(ruleName, appType)
	if len(fwConfig) == 0 {
		errorStr := fmt.Sprintf("%s not found", ruleName)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	res, err := xhttp.ReturnJsonResponse(fwConfig, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

// Zero usages in green splunk for 4 weeks ending 21 Oct 2021
// GET      /xconfadminService/ux/api/firmwareconfig/byEnvModelRuleName/{ruleName}
func GetFirmwareConfigByEnvModelRuleNameByRuleNameHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	entry, found := mux.Vars(r)[common.RULE_NAME]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.RULE_NAME)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	fwConfig := getFirmwareConfigByEnvModelRuleName(entry)
	if fwConfig != nil && fwConfig.ApplicationType != appType {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Entity with id: %s aplicationType does not match", fwConfig.ID))
		return
	}
	res, err := xhttp.ReturnJsonResponse(fwConfig, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func populateHeaderWithNumberOfItems(len int) map[string]string {
	headerMap := make(map[string]string, 1)
	headerMap[cFirmwareConfigNumberOfItems] = strconv.Itoa(len)
	return headerMap
}
