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
package telemetry

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	xutil "xconfadmin/util"

	xcommon "xconfadmin/common"
	xlogupload "xconfadmin/shared/logupload"
	xwcommon "xconfwebconfig/common"
	xwlogupload "xconfwebconfig/shared/logupload"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"
)

func GetTelemetryRulesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result := []*xwlogupload.TelemetryRule{}
	ruleList := xwlogupload.GetTelemetryRuleList()
	for _, teleRule := range ruleList {
		if teleRule.ApplicationType != applicationType {
			continue
		}
		result = append(result, teleRule)
	}
	response, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	queryParams := r.URL.Query()
	_, ok := queryParams[xcommon.EXPORT]
	if ok {
		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_TELEMETRY_RULES)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, response)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
	}
}

func GetTelemetryRuleByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	teleRule := xlogupload.GetOneTelemetryRule(id)
	if teleRule == nil {
		errorStr := fmt.Sprintf("%v not found", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	if teleRule.ApplicationType != applicationType {
		errorStr := fmt.Sprintf("%v not found", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	response, err := xhttp.ReturnJsonResponse(teleRule, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	queryParams := r.URL.Query()
	_, ok := queryParams[xcommon.EXPORT]
	if ok {
		teleRulelist := []xwlogupload.TelemetryRule{*teleRule}
		exresponse, err := xhttp.ReturnJsonResponse(teleRulelist, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_TELEMETRY_RULE + teleRule.ID + "_" + applicationType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, exresponse)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
	}
}

func DeleteTelmetryRuleByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	respEntity := DeleteTelemetryRulebyId(id, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func CreateTelemetryRuleHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "responsewriter cast error")
		return
	}
	body := xw.Body()
	newtmrule := xwlogupload.TelemetryRule{}
	err = json.Unmarshal([]byte(body), &newtmrule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := CreateTelemetryRule(&newtmrule, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, respEntity.Status, xhttp.ContextTypeHeader(r))

}

func UpdateTelemetryRuleHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, "unable to cast XpcResponseWriter object")
		return
	}
	body := xw.Body()
	newtmrule := xwlogupload.TelemetryRule{}
	err = json.Unmarshal([]byte(body), &newtmrule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdateTelemetryRule(&newtmrule, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, respEntity.Status, xhttp.ContextTypeHeader(r))

}

func PostTelemtryRuleEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	entities := []xwlogupload.TelemetryRule{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract entity from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		respEntity := CreateTelemetryRule(&entity, applicationType)
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

func PutTelemetryRuleEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	entities := []xwlogupload.TelemetryRule{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract entity from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		respEntity := UpdateTelemetryRule(&entity, applicationType)
		if respEntity.Status == http.StatusCreated {
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

func PostTelemetryRuleFilteredWithParamsHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
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
	contextMap := map[string]string{}
	if body != "" {
		if err := json.Unmarshal([]byte(body), &contextMap); err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid Json contents")
			return
		}
	}
	xutil.AddQueryParamsToContextMap(r, contextMap)
	contextMap[xwcommon.APPLICATION_TYPE] = applicationType

	tmrules := TelemetryRuleFilterByContext(contextMap)
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(tmrules))
	tmrules, err = TelemetryRuleGeneratePageWithContext(tmrules, contextMap)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := xhttp.ReturnJsonResponse(tmrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, response)
}
