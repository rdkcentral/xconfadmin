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
package dcm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"xconfadmin/common"
	xwcommon "xconfwebconfig/common"
	"xconfwebconfig/shared/logupload"

	xutil "xconfadmin/util"

	xwutil "xconfwebconfig/util"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"
)

func GetVodSettingsHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result := GetVodSettingsAll()
	appRules := []*logupload.VodSettings{}
	for _, rule := range result {
		if applicationType == rule.ApplicationType {
			appRules = append(appRules, rule)
		}
	}
	result = appRules
	response, _ := xwutil.JSONMarshal(result)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetVodSettingsByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(errorStr))
		return
	}
	vodsettings := GetVodSettings(id)
	if vodsettings == nil {
		errorStr := fmt.Sprintf("%v not found", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	if vodsettings.ApplicationType != applicationType {
		errorStr := fmt.Sprintf("%v not found,ApplicationType mismatch", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	response, _ := xwutil.JSONMarshal(vodsettings)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetVodSettingsSizeHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	final := []*logupload.VodSettings{}
	result := GetVodSettingsAll()
	for _, vs := range result {
		if vs.ApplicationType == applicationType {
			final = append(final, vs)
		}
	}
	response, _ := xwutil.JSONMarshal(len(final))
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetVodSettingsNamesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	final := []string{}
	result := GetVodSettingsAll()
	for _, vs := range result {
		if vs.ApplicationType == applicationType {
			final = append(final, vs.Name)
		}
	}
	response, _ := xwutil.JSONMarshal(final)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func DeleteVodSettingsByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	respEntity := DeleteVodSettingsbyId(id, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func CreateVodSettingsHandler(w http.ResponseWriter, r *http.Request) {
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
	newvs := logupload.VodSettings{}
	err = json.Unmarshal([]byte(body), &newvs)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	respEntity := CreateVodSettings(&newvs, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func UpdateVodSettingsHandler(w http.ResponseWriter, r *http.Request) {
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
	newvsrule := logupload.VodSettings{}
	err = json.Unmarshal([]byte(body), &newvsrule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	respEntity := UpdateVodSettings(&newvsrule, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func PostVodSettingsFilteredWithParamsHandler(w http.ResponseWriter, r *http.Request) {
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
	contextMap[xwcommon.APPLICATION_TYPE] = applicationType

	vsrules := VodSettingsFilterByContext(contextMap)
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(vsrules))
	vsrules, err = VodSettingsGeneratePageWithContext(vsrules, contextMap)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	response, _ := xwutil.JSONMarshal(vsrules)
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, response)
}

func GetVodSettingExportHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	allFormulas := GetDcmFormulaAll()
	vodList := []*logupload.VodSettings{}

	for _, DcmRule := range allFormulas {
		if DcmRule.ApplicationType != appType {
			continue
		}
		vodSetting := GetVodSettings(DcmRule.ID)
		vodList = append(vodList, vodSetting)
	}
	response, err := xhttp.ReturnJsonResponse(vodList, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_ALL_VOD_SETTINGS + "_" + appType)
	xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, response)
}
