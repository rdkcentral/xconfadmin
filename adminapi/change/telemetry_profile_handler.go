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

	xutil "github.com/rdkcentral/xconfadmin/util"

	"github.com/rdkcentral/xconfwebconfig/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	xcommon "github.com/rdkcentral/xconfadmin/common"
	xlogupload "github.com/rdkcentral/xconfadmin/shared/logupload"

	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xhttp "github.com/rdkcentral/xconfadmin/http"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	ds "github.com/rdkcentral/xconfwebconfig/db"

	"github.com/gorilla/mux"
)

func GetTelemetryProfileByIdHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
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

	profile := xwlogupload.GetOnePermanentTelemetryProfile(id)
	if profile == nil {
		errorStr := fmt.Sprintf("Entity with id %s does not exist", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	if _, ok := r.URL.Query()[xcommon.EXPORT]; ok {
		res, err := xhttp.ReturnJsonResponse([]*xwlogupload.PermanentTelemetryProfile{profile}, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		fileName := fmt.Sprintf("%s%s_%s", xcommon.ExportFileNames_PERMANENT_PROFILE, profile.ID, profile.ApplicationType)
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

func GetTelemetryProfilesHandler(w http.ResponseWriter, r *http.Request) {
	application, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	profiles := xlogupload.GetPermanentTelemetryProfileListByApplicationType(application)

	res, err := xhttp.ReturnJsonResponse(profiles, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	if _, ok := r.URL.Query()[xcommon.EXPORT]; ok {
		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_PERMANENT_PROFILES)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
	}
}

func GeneratePageTelemetryProfiles(list []*xwlogupload.PermanentTelemetryProfile, page int, pageSize int) (result []*xwlogupload.PermanentTelemetryProfile) {
	sort.Slice(list, func(i, j int) bool {
		return strings.Compare(strings.ToLower(list[i].Name), strings.ToLower(list[j].Name)) < 0
	})

	length := len(list)
	startIndex := page*pageSize - pageSize
	if page < 1 || startIndex > length || pageSize < 1 {
		return result
	}
	lastIndex := length
	if page*pageSize < length {
		lastIndex = page * pageSize
	}
	return list[startIndex:lastIndex]
}

func CreateTelemetryProfileChangeHandler(w http.ResponseWriter, r *http.Request) {
	permTelemetryProfile := xlogupload.NewEmptyPermanentTelemetryProfile()
	permTelemetryProfile.ApplicationType = ""
	_, err := auth.ExtractBodyAndCheckPermissions(permTelemetryProfile, w, r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	change, err := WriteCreateChange(r, permTelemetryProfile)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(change, r)
	if err != nil {
		xwhttp.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusCreated, xhttp.ContextTypeHeader(r))
}

func CreateTelemetryProfileHandler(w http.ResponseWriter, r *http.Request) {
	permTelemetryProfile := xlogupload.NewEmptyPermanentTelemetryProfile()
	permTelemetryProfile.ApplicationType = ""
	_, err := auth.ExtractBodyAndCheckPermissions(permTelemetryProfile, w, r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	savedProfile, err := CreatePermanentTelemetryProfile(r, permTelemetryProfile)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(savedProfile, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusCreated, xhttp.ContextTypeHeader(r))
}

func UpdateTelemetryProfileChangeHandler(w http.ResponseWriter, r *http.Request) {
	permTelemetryProfile := xlogupload.NewEmptyPermanentTelemetryProfile()
	permTelemetryProfile.ApplicationType = ""
	_, err := auth.ExtractBodyAndCheckPermissions(permTelemetryProfile, w, r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	change, err := WriteUpdateChangeOrSave(r, permTelemetryProfile)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(change, r)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func UpdateTelemetryProfileHandler(w http.ResponseWriter, r *http.Request) {
	permTelemetryProfile := xlogupload.NewEmptyPermanentTelemetryProfile()
	permTelemetryProfile.ApplicationType = ""
	_, err := auth.ExtractBodyAndCheckPermissions(permTelemetryProfile, w, r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	updatedProfile, err := UpdatePermanentTelemetryProfile(permTelemetryProfile)
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

func DeleteTelemetryProfileChangeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
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

	if util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is empty")
		return
	}

	change, err := WriteDeleteChange(r, id)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(change, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func DeleteTelemetryProfileHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
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

	if util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is empty")
		return
	}

	_, err = DeletePermanentTelemetryProfile(r, id)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

func CreateTelemetryIdsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	respEntity := CreateTelemetryIds()
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

func PostTelemetryProfileEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	entities := []xwlogupload.PermanentTelemetryProfile{}
	err = json.Unmarshal([]byte(body), &entities)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		if util.IsBlank(entity.Type) {
			entity.Type = xlogupload.PermanentTelemetryProfileConst
		}
		if _, err := WriteCreateChange(r, &entity); err != nil {
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

func PutTelemetryProfileEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	entities := []xwlogupload.PermanentTelemetryProfile{}
	err = json.Unmarshal([]byte(body), &entities)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		if _, err := WriteUpdateChangeOrSave(r, &entity); err != nil {
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

func PostTelemetryProfileFilteredHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

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

	profiles := GetTelemetryProfilesByContext(contextMap)
	profilesPerPage := GeneratePageTelemetryProfiles(profiles, pageNumber, pageSize)
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

func AddTelemetryProfileEntryHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
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

	if util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is empty")
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	telemetryElements := []xwlogupload.TelemetryElement{}
	err = json.Unmarshal([]byte(body), &telemetryElements)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	ds.GetCacheManager().ForceSyncChanges()
	profile := xwlogupload.GetOnePermanentTelemetryProfile(id)
	if profile == nil {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("Entity with id: %s does not exist", id)))
		return
	}
	updatedTelemetryEntries := profile.TelemetryProfile
	for _, telemetryElement := range telemetryElements {
		updatedTelemetryEntries, err = AddPermanentTelemetryProfileElement(&telemetryElement, updatedTelemetryEntries)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
	}
	profile.TelemetryProfile = updatedTelemetryEntries
	updatedProfile, err := UpdatePermanentTelemetryProfile(profile)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(updatedProfile, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func AddTelemetryProfileEntryChangeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
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

	if util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is empty")
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	telemetryElements := []xwlogupload.TelemetryElement{}
	err = json.Unmarshal([]byte(body), &telemetryElements)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	profile := xwlogupload.GetOnePermanentTelemetryProfile(id)
	if profile == nil {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("Entity with id: %s does not exist", id)))
		return
	}
	profile, _ = profile.Clone()
	updatedTelemetryEntries := profile.TelemetryProfile
	for _, telemetryElement := range telemetryElements {
		updatedTelemetryEntries, err = AddPermanentTelemetryProfileElement(&telemetryElement, updatedTelemetryEntries)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
	}
	profile.TelemetryProfile = updatedTelemetryEntries
	change, err := WriteUpdateChangeOrSave(r, profile)
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

func RemoveTelemetryProfileEntryHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
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

	if util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is empty")
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	entriesToRemove := []xwlogupload.TelemetryElement{}
	err = json.Unmarshal([]byte(body), &entriesToRemove)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ds.GetCacheManager().ForceSyncChanges()
	profile := xwlogupload.GetOnePermanentTelemetryProfile(id)
	if profile == nil {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("Entity with id: %s does not exist", id)))
		return
	}
	profile, _ = profile.Clone()
	updatedTelemetryEntries := profile.TelemetryProfile
	for _, telemetryElement := range entriesToRemove {
		updatedTelemetryEntries, err = RemovePermanentTelemetryProfileElement(&telemetryElement, updatedTelemetryEntries)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
	}
	profile.TelemetryProfile = updatedTelemetryEntries

	change, err := UpdatePermanentTelemetryProfile(profile)
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

func RemoveTelemetryProfileEntryChangeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
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

	if util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is empty")
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	entriesToRemove := []xwlogupload.TelemetryElement{}
	err = json.Unmarshal([]byte(body), &entriesToRemove)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	profile := xwlogupload.GetOnePermanentTelemetryProfile(id)
	if profile == nil {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("Entity with id: %s does not exist", id)))
		return
	}
	profile, _ = profile.Clone()
	updatedTelemetryEntries := profile.TelemetryProfile
	for _, telemetryElement := range entriesToRemove {
		updatedTelemetryEntries, err = RemovePermanentTelemetryProfileElement(&telemetryElement, updatedTelemetryEntries)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
	}
	profile.TelemetryProfile = updatedTelemetryEntries

	change, err := WriteUpdateChangeOrSave(r, profile)
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
