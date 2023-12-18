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
	"time"

	"github.com/gorilla/mux"

	xutil "xconfadmin/util"

	xwutil "xconfwebconfig/util"

	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"

	xcommon "xconfadmin/common"
	xwcommon "xconfwebconfig/common"
	"xconfwebconfig/shared/logupload"

	"xconfadmin/adminapi/auth"
)

const (
	// DATE_TIME_FORMAT = "yyyy-MM-dd HH:mm:ss";
	dateLayout = "2006-01-02 15:04:05"
)

// str2Time just parses the given string with the layout "2006-01-02 15:04:05"
// This is NOT the standard time format! e.g. time.RFC822Z or time.RFC3339
// This format is simply "yyyy-mm-dd hh:mm:ss"
func str2Time(d string) (time.Time, error) {
	return time.Parse(dateLayout, d)
}

// changeTZ changes timezone from UTC to the new timezone
// From part is always UTC
// If the original time is say, "2021-10-05 00:00:00", and if the timezone is "MST"
// i.e. Mountain Standard time, then we want the new time to be "2021-10-05 07:00:00"
// i.e The given time is considered to be in MST, and normalized back to UTC
func changeTZ(t time.Time, tz *time.Location) string {
	newT := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, tz)
	return newT.In(xwutil.TZ).Format(dateLayout)
}

func GetDeviceSettingsHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result := GetDeviceSettingsAll()
	appRules := []*logupload.DeviceSettings{}
	for _, rule := range result {
		if appType == rule.ApplicationType {
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

func GetDeviceSettingsByIdHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.DCM_ENTITY)
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
	devicesettings := GetDeviceSettings(id)
	if devicesettings == nil {
		errorStr := fmt.Sprintf("%v not found", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	if devicesettings.ApplicationType != appType {
		errorStr := fmt.Sprintf("%v not found, ApplicationType doesn't match", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, errorStr)
		return
	}
	response, err := xhttp.ReturnJsonResponse(devicesettings, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetDeviceSettingsSizeHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	final := []*logupload.DeviceSettings{}
	result := GetDeviceSettingsAll()
	for _, ds := range result {
		if ds.ApplicationType == appType {
			final = append(final, ds)
		}
	}
	response, err := xhttp.ReturnJsonResponse(len(final), r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetDeviceSettingsNamesHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	final := []string{}
	result := GetDeviceSettingsAll()
	for _, ds := range result {
		if ds.ApplicationType == appType {
			final = append(final, ds.Name)
		}
	}
	response, err := xhttp.ReturnJsonResponse(final, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func DeleteDeviceSettingsByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	respEntity := DeleteDeviceSettingsbyId(id, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func CreateDeviceSettingsHandler(w http.ResponseWriter, r *http.Request) {
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
	newds := logupload.DeviceSettings{}
	err = json.Unmarshal([]byte(body), &newds)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	respEntity := CreateDeviceSettings(&newds, applicationType)
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

func UpdateDeviceSettingsHandler(w http.ResponseWriter, r *http.Request) {
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
	newdsrule := logupload.DeviceSettings{}
	err = json.Unmarshal([]byte(body), &newdsrule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	respEntity := UpdateDeviceSettings(&newdsrule, applicationType)
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

func PostDeviceSettingsFilteredWithParamsHandler(w http.ResponseWriter, r *http.Request) {
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

	dsrules := DeviceSettingsFilterByContext(contextMap)
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(dsrules))
	dsrules, err = DeviceSettingsGeneratePageWithContext(dsrules, contextMap)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := xhttp.ReturnJsonResponse(dsrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, response)
}

func GetDeviceSettingsExportHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	allFormulas := GetDcmFormulaAll()
	dsList := []*logupload.DeviceSettings{}

	for _, DcmRule := range allFormulas {
		if DcmRule.ApplicationType != appType {
			continue
		}
		dsl := GetDeviceSettings(DcmRule.ID)
		dsList = append(dsList, dsl)
	}
	response, err := xhttp.ReturnJsonResponse(dsList, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_DEVICE_SETTINGS + "_" + appType)
	xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, response)
}
