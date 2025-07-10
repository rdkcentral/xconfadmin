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
package change

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	xutil "xconfadmin/util"
	"xconfwebconfig/dataapi/dcm/telemetry"

	xcommon "xconfadmin/common"
	xwcommon "xconfwebconfig/common"

	xshared "xconfadmin/shared"
	xlogupload "xconfadmin/shared/logupload"
	"xconfwebconfig/common"
	"xconfwebconfig/shared/logupload"
	xwlogupload "xconfwebconfig/shared/logupload"
	"xconfwebconfig/util"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"

	"github.com/gorilla/mux"
)

func GetTelemetryTwoProfilesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	profiles := xlogupload.GetTelemetryTwoProfileListByApplicationType(applicationType)

	res, err := xhttp.ReturnJsonResponse(profiles, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	if _, ok := r.URL.Query()[xcommon.EXPORT]; ok {
		fileName := xcommon.ExportFileNames_ALL_TELEMETRY_TWO_PROFILES + "_" + applicationType
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
	}
}

func CreateTelemetryTwoProfileChangeHandler(w http.ResponseWriter, r *http.Request) {
	telemetryTwoProfile := xlogupload.NewEmptyTelemetryTwoProfile()
	telemetryTwoProfile.ApplicationType = ""
	_, err := auth.ExtractBodyAndCheckPermissions(telemetryTwoProfile, w, r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	change, err := WriteCreateChangeTelemetryTwoProfile(r, telemetryTwoProfile)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(change, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusCreated, xhttp.ContextTypeHeader(r))
}

func CreateTelemetryTwoProfileHandler(w http.ResponseWriter, r *http.Request) {
	telemetryTwoProfile := xlogupload.NewEmptyTelemetryTwoProfile()
	telemetryTwoProfile.ApplicationType = ""
	_, err := auth.ExtractBodyAndCheckPermissions(telemetryTwoProfile, w, r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	createdProfile, err := CreateTelemetryTwoProfile(r, telemetryTwoProfile)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(createdProfile, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusCreated, xhttp.ContextTypeHeader(r))
}

func UpdateTelemetryTwoProfileChangeHandler(w http.ResponseWriter, r *http.Request) {
	telemetryTwoProfile := xlogupload.NewEmptyTelemetryTwoProfile()
	telemetryTwoProfile.ApplicationType = ""
	_, err := auth.ExtractBodyAndCheckPermissions(telemetryTwoProfile, w, r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	change, err := WriteUpdateChangeOrSaveTelemetryTwoProfile(r, telemetryTwoProfile)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(change, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func UpdateTelemetryTwoProfileHandler(w http.ResponseWriter, r *http.Request) {
	telemetryTwoProfile := xlogupload.NewEmptyTelemetryTwoProfile()
	telemetryTwoProfile.ApplicationType = ""
	_, err := auth.ExtractBodyAndCheckPermissions(telemetryTwoProfile, w, r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	updatedProfile, err := UpdateTelemetryTwoProfile(r, telemetryTwoProfile)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(updatedProfile, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func DeleteTelemetryTwoProfileChangeHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	if util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is empty")
		return
	}

	change, err := WriteDeleteChangeTelemetryTwoProfile(r, id)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(change, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))

	xwhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

func DeleteTelemetryTwoProfileHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	if util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is empty")
		return
	}

	err := DeleteTelemetryTwoProfile(r, id)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

func GetTelemetryTwoProfileByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
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

	profile := xlogupload.GetOneTelemetryTwoProfile(id)
	if profile == nil {
		errorStr := fmt.Sprintf("Entity with id %s does not exist", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	if _, ok := r.URL.Query()[xcommon.EXPORT]; ok {
		profileToExport := []*xwlogupload.TelemetryTwoProfile{profile}
		res, err := xhttp.ReturnJsonResponse(profileToExport, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		fileName := xcommon.ExportFileNames_TELEMETRY_TWO_PROFILE + profile.ID + "_" + applicationType
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		res, err := xhttp.ReturnJsonResponse(profile, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
	}
}

// TODO remove this function when we are sure that we don't need it anymore
func GetTelemetryTwoProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	queryParams := map[string]string{}
	xutil.AddQueryParamsToContextMap(r, queryParams)
	pageNumber, err := strconv.Atoi(queryParams[xcommon.PAGE_NUMBER])
	if err != nil || pageNumber < 1 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageNumber")
		return
	}
	pageSize, err := strconv.Atoi(queryParams[xcommon.PAGE_SIZE])
	if err != nil || pageSize < 1 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageSize")
		return
	}

	profiles := xlogupload.GetAllTelemetryTwoProfileList(applicationType)
	profilesPerPage := GeneratePageTelemetryTwoProfiles(profiles, pageNumber, pageSize)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(profilesPerPage, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(profiles))
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, []byte(res))
}

func PostTelemetryTwoProfilesByIdListHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	idList := []string{}
	err = json.Unmarshal([]byte(body), &idList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	profiles := GetTelemetryTwoProfilesByIdList(applicationType, idList)

	res, err := xhttp.ReturnJsonResponse(profiles, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func PostTelemetryTwoProfileFilteredHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	queryParams := map[string]string{}
	xutil.AddQueryParamsToContextMap(r, queryParams)
	pageNumber, err := strconv.Atoi(queryParams[xcommon.PAGE_NUMBER])
	if err != nil || pageNumber < 1 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageNumber")
		return
	}
	pageSize, err := strconv.Atoi(queryParams[xcommon.PAGE_SIZE])
	if err != nil || pageSize < 1 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageSize")
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()

	contextMap := make(map[string]string)

	if body != "" {
		if err := json.Unmarshal([]byte(body), &contextMap); err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	xutil.AddQueryParamsToContextMap(r, contextMap)
	contextMap[xwcommon.APPLICATION_TYPE] = applicationType

	profiles := GetTelemetryTwoProfilesByContext(contextMap)
	sort.SliceStable(profiles, func(i, j int) bool {
		return strings.Compare(strings.ToLower(profiles[i].ID), strings.ToLower(profiles[j].ID)) < 0
	})
	profilesPerPage := GeneratePageTelemetryTwoProfiles(profiles, pageNumber, pageSize)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(profilesPerPage, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(profiles))
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, []byte(res))
}

func PostTelemetryTwoProfileEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	entities := []logupload.TelemetryTwoProfile{}
	err := json.Unmarshal([]byte(body), &entities)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		if _, err := WriteCreateChangeTelemetryTwoProfile(r, &entity); err != nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
		} else {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: entity.ID,
			}
		}
	}

	res, err := xhttp.ReturnJsonResponse(entitiesMap, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func PutTelemetryTwoProfileEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	entities := []logupload.TelemetryTwoProfile{}
	err := json.Unmarshal([]byte(body), &entities)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		if _, err := WriteUpdateChangeOrSaveTelemetryTwoProfile(r, &entity); err != nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
		} else {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: entity.ID,
			}
		}
	}

	res, err := xhttp.ReturnJsonResponse(entitiesMap, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func TelemetryTwoTestPageHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, err.(xcommon.XconfError).StatusCode, err.Error())
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Unable to extract body"))
		return
	}

	body := xw.Body()
	contextMap := map[string]string{}
	if err := json.Unmarshal([]byte(body), &contextMap); err != nil {
		xutil.AddBodyParamsToContextMap(body, contextMap)
	}
	xutil.AddQueryParamsToContextMap(r, contextMap)

	if err := xshared.NormalizeCommonContext(contextMap, common.ESTB_MAC_ADDRESS, common.ECM_MAC_ADDRESS); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	contextMap[xwcommon.APPLICATION_TYPE] = applicationType

	telemetryProfileService := telemetry.NewTelemetryProfileService()
	telemetryTwoRules := telemetryProfileService.ProcessTelemetryTwoRulesForAS(contextMap)

	result := make(map[string]interface{})
	result["context"] = contextMap
	result["result"] = telemetryTwoRules

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}
