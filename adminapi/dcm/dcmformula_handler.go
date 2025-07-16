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
	"sort"
	"strconv"

	"github.com/gorilla/mux"

	"xconfadmin/adminapi/auth"
	queries "xconfadmin/adminapi/queries"
	"xconfadmin/common"
	xhttp "xconfadmin/http"
	core "xconfadmin/shared"
	requtil "xconfadmin/util"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

func GetDcmFormulaHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	allFormulas := GetDcmFormulaAll()

	queryParams := r.URL.Query()
	_, ok := queryParams[common.EXPORT]
	if ok {
		fwsList := []*logupload.FormulaWithSettings{}
		for _, DcmRule := range allFormulas {
			if DcmRule.ApplicationType != appType {
				continue
			}
			fws := logupload.FormulaWithSettings{}
			fws.Formula = DcmRule
			fws.DeviceSettings = GetDeviceSettings(DcmRule.ID)
			fws.LogUpLoadSettings = logupload.GetOneLogUploadSettings(DcmRule.ID)
			fws.VodSettings = GetVodSettings(DcmRule.ID)
			fwsList = append(fwsList, &fws)
		}
		response, err := xhttp.ReturnJsonResponse(fwsList, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_ALL_FORMULAS + "_" + appType)
		xhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, response)
	} else {
		list := []*logupload.DCMGenericRule{}
		for _, DcmRule := range allFormulas {
			if DcmRule.ApplicationType != appType {
				continue
			}
			list = append(list, DcmRule)
		}
		response, err := xhttp.ReturnJsonResponse(list, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xhttp.WriteXconfResponse(w, http.StatusOK, response)
	}
}

func GetDcmFormulaByIdHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.DCM_ENTITY)
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

	formula := GetDcmFormula(id)
	if formula == nil {
		errorStr := fmt.Sprintf("%v not found", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	if formula.ApplicationType != appType {
		errorStr := fmt.Sprintf("%v not found", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	queryParams := r.URL.Query()
	_, ok := queryParams[common.EXPORT]
	if ok {
		fws := logupload.FormulaWithSettings{}
		fws.Formula = formula
		fws.DeviceSettings = GetDeviceSettings(formula.ID)
		fws.LogUpLoadSettings = logupload.GetOneLogUploadSettings(formula.ID)
		fws.VodSettings = GetVodSettings(formula.ID)
		formulalist := []logupload.FormulaWithSettings{fws}
		exresponse, err := xhttp.ReturnJsonResponse(formulalist, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_FORMULA + formula.ID + "_" + appType)
		xhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, exresponse)
	} else {
		response, err := xhttp.ReturnJsonResponse(formula, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xhttp.WriteXconfResponse(w, http.StatusOK, response)
	}
}

func GetDcmFormulaSizeHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	final := []*logupload.DCMGenericRule{}
	result := GetDcmFormulaAll()
	for _, DcmRule := range result {
		if DcmRule.ApplicationType == appType {
			final = append(final, DcmRule)
		}
	}
	response, err := xhttp.ReturnJsonResponse(strconv.Itoa(len(final)), r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetDcmFormulaNamesHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	final := []string{}
	result := GetDcmFormulaAll()
	for _, DcmRule := range result {
		if DcmRule.ApplicationType == appType {
			final = append(final, DcmRule.Name)
		}
	}
	response, err := xhttp.ReturnJsonResponse(final, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func DeleteDcmFormulaByIdHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[common.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	ds.GetCacheManager().ForceSyncChanges()

	respEntity := DeleteDcmFormulabyId(id, appType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func CreateDcmFormulaHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	newdfrule := logupload.DCMGenericRule{}
	err = json.Unmarshal([]byte(body), &newdfrule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ds.GetCacheManager().ForceSyncChanges()
	respEntity := CreateDcmRule(&newdfrule, appType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func UpdateDcmFormulaHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, "unable to cast XResponseWriter object")
		return
	}
	body := xw.Body()
	newdfrule := logupload.DCMGenericRule{}
	err = json.Unmarshal([]byte(body), &newdfrule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ds.GetCacheManager().ForceSyncChanges()
	respEntity := UpdateDcmRule(&newdfrule, appType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func getsettings(value string, id string) bool {
	switch value {
	case "devicesettings":
		ds := logupload.GetOneDeviceSettings(id)
		return ds != nil
	case "vodsettings":
		vs := logupload.GetOneVodSettings(id)
		return vs != nil
	case "loguploadsettings":
		ls := logupload.GetOneLogUploadSettings(id)
		return ls != nil
	}
	return false
}

func DcmFormulaSettingsAvailabilitygHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, "responsewriter cast error")
		return
	}
	body := xw.Body()
	idlist := []string{}
	err = json.Unmarshal([]byte(body), &idlist)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	dcmmap := make(map[string]map[string]bool)
	for _, id := range idlist {
		data := make(map[string]bool)
		data["vodSettings"] = getsettings("vodsettings", id)
		data["logUploadSettings"] = getsettings("loguploadsettings", id)
		data["deviceSettings"] = getsettings("devicesettings", id)
		dcmmap[id] = data
	}
	res, err := xhttp.ReturnJsonResponse(&dcmmap, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func getiFormulaAvail(id string) bool {
	dfrule := GetDcmFormula(id)
	return dfrule != nil
}

func DcmFormulasAvailabilitygHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, "responsewriter cast error")
		return
	}
	body := xw.Body()
	idlist := []string{}
	err = json.Unmarshal([]byte(body), &idlist)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	data := make(map[string]bool)
	for _, id := range idlist {
		data[id] = getiFormulaAvail(id)
	}
	res, err := xhttp.ReturnJsonResponse(&data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func PostDcmFormulaFilteredWithParamsHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, "responsewriter cast error")
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
	requtil.AddQueryParamsToContextMap(r, contextMap)
	contextMap[core.APPLICATION_TYPE] = applicationType

	dfrules := DcmFormulaFilterByContext(contextMap)
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(dfrules))
	dfrules, err = DcmFormulaRuleGeneratePageWithContext(dfrules, contextMap)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := xhttp.ReturnJsonResponse(dfrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, response)
}

func DcmFormulaChangePriorityHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, ok := mux.Vars(r)[common.ID]
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", common.ID))
		return
	}
	newPriorityStr, ok := mux.Vars(r)[common.NEW_PRIORITY]
	if !ok {
		errorStr := fmt.Sprintf("%v is invalid", common.NEW_PRIORITY)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	ds.GetCacheManager().ForceSyncChanges()

	formulaUpdateMutex.Lock()
	defer formulaUpdateMutex.Unlock()
	formulaToUpdate := logupload.GetOneDCMGenericRule(id)
	if formulaToUpdate == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("unable to find dcm formula  with id  %s", id))
		return
	}

	newPriority, err := strconv.Atoi(newPriorityStr)
	if err != nil || newPriority <= 0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid priority value %s", newPriorityStr))
		return
	}
	if appType != formulaToUpdate.ApplicationType {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "ApplicationType doesn't match")
		return
	}
	formulasByApplicationType := GetDcmRulesByApplicationType(formulaToUpdate.ApplicationType)
	prioritizables := DcmRulesToPrioritizables(formulasByApplicationType)
	reorganizedFormulas := queries.UpdatePrioritizablesPriorities(prioritizables, formulaToUpdate.Priority, newPriority)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("unable to re-organize priorities: %s", err))
		return
	}

	for _, entry := range reorganizedFormulas {
		if err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, entry.GetID(), entry); err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("unable to update dcm rule: %s", err))
			return
		}
	}
	response, err := xhttp.ReturnJsonResponse(reorganizedFormulas, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func ImportDcmFormulaWithOverwriteHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	overwriteStr, ok := mux.Vars(r)[common.OVERWRITE]
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", common.OVERWRITE))
		return
	}
	overwrite := false
	if overwriteStr == "true" {
		overwrite = true
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract body")
		return
	}

	body := xw.Body()
	formulaWithSettings := logupload.FormulaWithSettings{}
	err = json.Unmarshal([]byte(body), &formulaWithSettings)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ds.GetCacheManager().ForceSyncChanges()
	respEntity := importFormula(&formulaWithSettings, overwrite, appType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func ImportDcmFormulasHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	formulaWithSettingsList := []logupload.FormulaWithSettings{}
	err = json.Unmarshal([]byte(body), &formulaWithSettingsList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	sort.Slice(formulaWithSettingsList, func(i, j int) bool {
		return formulaWithSettingsList[i].Formula.Priority < formulaWithSettingsList[j].Formula.Priority
	})

	failedToImport := []string{}
	successfulImportIds := []string{}
	ds.GetCacheManager().ForceSyncChanges()

	for _, formulaWithSettings := range formulaWithSettingsList {
		formulaWithSettings := formulaWithSettings
		formula := formulaWithSettings.Formula
		respEntity := importFormula(&formulaWithSettings, false, appType)
		if respEntity.Error != nil {
			failedToImport = append(failedToImport, respEntity.Error.Error())
		} else {
			successfulImportIds = append(successfulImportIds, formula.ID)
		}
	}

	result := map[string][]string{
		"success": successfulImportIds,
		"failure": failedToImport,
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func PostDcmFormulaListHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	formulaWithSettingsList := []*logupload.FormulaWithSettings{}
	err = json.Unmarshal([]byte(body), &formulaWithSettingsList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ds.GetCacheManager().ForceSyncChanges()
	result := importFormulas(formulaWithSettingsList, appType, false)

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func PutDcmFormulaListHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	formulaWithSettingsList := []*logupload.FormulaWithSettings{}
	err = json.Unmarshal([]byte(body), &formulaWithSettingsList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ds.GetCacheManager().ForceSyncChanges()
	result := importFormulas(formulaWithSettingsList, appType, true)

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, res)
}
