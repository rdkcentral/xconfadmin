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
package dcm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	xutil "xconfadmin/util"

	"xconfwebconfig/common"
	"xconfwebconfig/shared/logupload"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"
)

func GetLogUploadSettingsHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result := GetLogUploadSettingsList()
	appRules := []*logupload.LogUploadSettings{}
	for _, rule := range result {
		if applicationType == rule.ApplicationType {
			appRules = append(appRules, rule)
		}
	}
	result = appRules
	response, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetLogUploadSettingsByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[common.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	loguploadsettings := logupload.GetOneLogUploadSettings(id)
	if loguploadsettings == nil {
		errorStr := fmt.Sprintf("%v not found", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	if loguploadsettings.ApplicationType != applicationType {
		errorStr := fmt.Sprintf("%v not found,ApplicationType doesn't match", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	response, err := xhttp.ReturnJsonResponse(loguploadsettings, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetLogUploadSettingsSizeHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	final := []*logupload.LogUploadSettings{}
	result := GetLogUploadSettingsList()
	for _, lu := range result {
		if lu.ApplicationType == applicationType {
			final = append(final, lu)
		}
	}
	response, err := xhttp.ReturnJsonResponse(len(final), r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetLogUploadSettingsNamesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	final := []string{}
	result := GetLogUploadSettingsList()
	for _, lu := range result {
		if lu.ApplicationType == applicationType {
			final = append(final, lu.Name)
		}
	}
	response, err := xhttp.ReturnJsonResponse(final, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func DeleteLogUploadSettingsByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[common.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	respEntity := DeleteLogUploadSettingsbyId(id, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func CreateLogUploadSettingsHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	newlu := logupload.LogUploadSettings{}
	err = json.Unmarshal([]byte(body), &newlu)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	respEntity := CreateLogUploadSettings(&newlu, applicationType)
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

func UpdateLogUploadSettingsHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	newlurule := logupload.LogUploadSettings{}
	err = json.Unmarshal([]byte(body), &newlurule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	respEntity := UpdateLogUploadSettings(&newlurule, applicationType)
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

func PostLogUploadSettingsFilteredWithParamsHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
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
	contextMap[common.APPLICATION_TYPE] = applicationType

	lurules := LogUploadSettingsFilterByContext(contextMap)
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(lurules))
	lurules, err = LogUploadSettingsGeneratePageWithContext(lurules, contextMap)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	response, err := xhttp.ReturnJsonResponse(lurules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, response)
}
