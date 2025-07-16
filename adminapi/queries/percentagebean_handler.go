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
	"fmt"
	"net/http"

	xcommon "xconfadmin/common"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"github.com/gorilla/mux"

	"xconfadmin/util"

	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
)

const (
	cPercentageBeanPageNumber          = "pageNumber"
	cPercentageBeanPageSize            = "pageSize"
	cPercentageBeanenvironment         = "ENVIRONMENT"
	cPercentageBeanlastknowngood       = "LAST_KNOWN_GOOD"
	cPercentageBeanmincheckversion     = "MIN_CHECK_VERSION"
	cPercentageBeanintermediateversion = "INTERMEDIATE_VERSION"
)

func GetPercentageBeanAllHandler(w http.ResponseWriter, r *http.Request) {
	contextMap := make(map[string]string)

	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	util.AddQueryParamsToContextMap(r, contextMap)

	var result interface{}

	result, err = GetAllPercentageBeansFromDB(applicationType, true, false)

	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	_, ok := contextMap[xcommon.EXPORT]
	if ok {
		//TODO: replace to struct instead of map
		percentageBeansToExport := make(map[string]interface{})
		percentageBeansToExport["percentageBeans"] = result
		res, err := xhttp.ReturnJsonResponse(percentageBeansToExport, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ENV_MODEL_PERCENTAGE_BEANS + "_" + applicationType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
	}
}

func GetPercentageBeanByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[common.ID]
	if !found {
		errorStr := fmt.Sprintf("Required ID parameter '%s' is not present", common.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	bean, err := GetOnePercentageBeanFromDB(id)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "Entity with id: "+id+" does not exist")
		return
	}
	if applicationType != bean.ApplicationType {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "ApplicationType doesn't match")
		return
	}
	res, err := xhttp.ReturnJsonResponse(bean, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	queryParams := r.URL.Query()
	_, ok := queryParams[xcommon.EXPORT]
	if ok {
		beanlist := []coreef.PercentageBean{*bean}
		percentFilterVo := coreef.PercentFilterVo{}
		percentFilterVo.PercentageBeans = beanlist
		exres, err := xhttp.ReturnJsonResponse(percentFilterVo, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}

		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ENV_MODEL_PERCENTAGE_BEAN + bean.ID + "_" + applicationType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, exres)
	} else {
		xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
	}
}

func DeletePercentageBeanByIdHandler(w http.ResponseWriter, r *http.Request) {
	DeletePercentageBeanHandler(w, r)
}

func GetAllPercentageBeanAsRule(w http.ResponseWriter, r *http.Request) {
	contextMap := make(map[string]string)
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	util.AddQueryParamsToContextMap(r, contextMap)

	var result []*firmware.FirmwareRule

	result, err = GetAllGlobalPercentageBeansAsRuleFromDB(applicationType, true)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	_, ok := contextMap[xcommon.EXPORT]
	if ok {
		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ENV_MODEL_PERCENTAGE_AS_RULES + "_" + applicationType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
	}
}

func GetPercentageBeanAsRuleById(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[common.ID]
	if !found {
		errorStr := fmt.Sprintf("Required ID parameter '%s' is not present", common.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	fwRule, err := GetOnePercentageBeanFromDB(id)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "\"<h2>404 NOT FOUND</h2>\"")
		return
	}
	if fwRule.ApplicationType != applicationType {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "ApplicationType mismatch")
		return
	}
	res, err := xhttp.ReturnJsonResponse(fwRule, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	queryParams := r.URL.Query()
	_, ok := queryParams[xcommon.EXPORT]
	if ok {
		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ENV_MODEL_PERCENTAGE_AS_RULE + fwRule.ID + " _" + applicationType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
	}
}

func PostPercentageBeanEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	entities := []coreef.PercentageBean{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract entity from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	fields := xw.Audit()
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		respEntity := CreatePercentageBean(&entity, applicationType, fields)
		if respEntity.Status != http.StatusCreated {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: respEntity.Error.Error(),
			}
		} else {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: entity.ID,
			}
		}
	}
	response, err := xhttp.ReturnJsonResponse(entitiesMap, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func PutPercentageBeanEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	fields := xw.Audit()
	entities := []coreef.PercentageBean{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract entity from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		respEntity := UpdatePercentageBean(&entity, applicationType, fields)
		if respEntity.Status == http.StatusOK {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: entity.ID,
			}
		} else {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: respEntity.Error.Error(),
			}
		}
	}
	response, err := xhttp.ReturnJsonResponse(entitiesMap, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func PostPercentageBeanFilteredWithParamsHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract body")
		return
	}
	body := xw.Body()
	contextMap := make(map[string]string)
	if body != "" {
		if err := json.Unmarshal([]byte(body), &contextMap); err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid Json contents")
			return
		}
	}
	util.AddQueryParamsToContextMap(r, contextMap)
	contextMap[common.APPLICATION_TYPE] = applicationType

	pbrules := PercentageBeanFilterByContext(contextMap, applicationType)
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(pbrules))
	pbrules, err = PercentageBeanRuleGeneratePageWithContext(pbrules, contextMap)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := xhttp.ReturnJsonResponse(pbrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, response)
}
