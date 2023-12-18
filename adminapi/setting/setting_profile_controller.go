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
package setting

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"strconv"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	xwcommon "xconfwebconfig/common"

	xcommon "xconfadmin/common"
	"xconfwebconfig/shared/logupload"
	"xconfwebconfig/util"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"
)

const (
	NumberOfItems = "numberOfItems"
	PageNumber    = "pageNumber"
	PageSize      = "pageSize"
)

func GetSettingProfilesAllExport(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	all := GetAll()
	settingProfiles := []*logupload.SettingProfiles{}
	for _, entity := range all {
		if entity.ApplicationType == applicationType {
			settingProfiles = append(settingProfiles, entity)
		}
	}
	response, err := util.JSONMarshal(settingProfiles)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal settingProfiles error: %v", err))
	}
	_, ok := r.URL.Query()["export"]
	if ok {
		headerMap := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_SETTING_PROFILES)
		xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
	}
}

func GetSettingProfileOneExport(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	id, found := mux.Vars(r)[xwcommon.ID]
	if !found || len(strings.TrimSpace(id)) == 0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is blank")
		return
	}
	settingProfile, _ := GetOne(id)
	if settingProfile == nil {
		invalid := "Entity with id: " + id + " does not exist"
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, invalid)
		return
	}

	if _, ok := r.URL.Query()[xcommon.EXPORT]; ok {
		res, err := xhttp.ReturnJsonResponse([]*logupload.SettingProfiles{settingProfile}, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		fileName := fmt.Sprintf("%s%s_%s", xcommon.ExportFileNames_SETTING_PROFILE, settingProfile.ID, settingProfile.ApplicationType)
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		res, err := xhttp.ReturnJsonResponse(settingProfile, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
	}
}

func GetAllSettingProfilesWithPage(w http.ResponseWriter, r *http.Request) {
	var pageNumberStr, pageSizeStr string
	pageNumber := 1
	pageSize := 50
	var err error
	if values, ok := r.URL.Query()[PageNumber]; ok {
		pageNumberStr = values[0]
		pageNumber, err = strconv.Atoi(pageNumberStr)
		if err != nil {
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("pageNumber must be a number"))
			return
		}
	}
	if values, ok := r.URL.Query()[PageSize]; ok {
		pageSizeStr = values[0]
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("pageSize must be a number"))
			return
		}
	}
	settingProfiles := GetAll()
	featureRuleList := SettingProfilesGeneratePage(settingProfiles, pageNumber, pageSize)
	response, err := util.JSONMarshal(featureRuleList)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRules error: %v", err))
	}
	headerMap := createNumberOfItemsHttpHeaders(settingProfiles)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func createNumberOfItemsHttpHeaders(entities []*logupload.SettingProfiles) map[string]string {
	headerMap := make(map[string]string, 1)
	if entities == nil {
		headerMap[NumberOfItems] = "0"
	} else {
		headerMap[NumberOfItems] = strconv.Itoa(len(entities))
	}
	return headerMap
}

func DeleteOneSettingProfilesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found || util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusMethodNotAllowed, "missing id")
		return
	}
	_, err = Delete(id, applicationType)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

func GetSettingProfilesFilteredWithPage(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	var pageNumberStr, pageSizeStr string
	pageNumber := 1
	pageSize := 50
	if values, ok := r.URL.Query()[PageNumber]; ok {
		pageNumberStr = values[0]
		pageNumber, err = strconv.Atoi(pageNumberStr)
		if err != nil {
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("pageNumber must be a number"))
			return
		}
	}
	if values, ok := r.URL.Query()[PageSize]; ok {
		pageSizeStr = values[0]
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("pageSize must be a number"))
			return
		}
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.Error(w, http.StatusInternalServerError, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	contextMap := make(map[string]string)
	body := xw.Body()
	if body != "" {
		if err := json.Unmarshal([]byte(body), &contextMap); err != nil {
			response := "Unable to extract searchContext from json file:" + err.Error()
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
			return
		}
	}
	contextMap[xcommon.APPLICATION_TYPE] = applicationType

	settingProfiles := FindByContext(contextMap)
	sort.Slice(settingProfiles, func(i, j int) bool {
		return strings.Compare(strings.ToLower(settingProfiles[i].SettingProfileID), strings.ToLower(settingProfiles[j].SettingProfileID)) < 0
	})
	settingProfilesList := SettingProfilesGeneratePage(settingProfiles, pageNumber, pageSize)
	response, err := util.JSONMarshal(settingProfilesList)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRules error: %v", err))
	}
	headerMap := createNumberOfItemsHttpHeaders(settingProfiles)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func CreateSettingProfileHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.Error(w, http.StatusInternalServerError, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	settingProfiles := logupload.SettingProfiles{}
	if body != "" {
		err := json.Unmarshal([]byte(body), &settingProfiles)
		if err != nil {
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
			return
		}
	}

	err = Create(&settingProfiles, applicationType)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(settingProfiles)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal ettingProfiles error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusCreated, response)
}

func CreateSettingProfilesPackageHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Unable to extract Body"))
		return
	}
	entities := []logupload.SettingProfiles{}
	body := xw.Body()
	if body != "" {
		if err := json.Unmarshal([]byte(body), &entities); err != nil {
			response := "Unable to extract SettingProfiles from json file:" + err.Error()
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
			return
		}
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		err := Create(&entity, applicationType)
		if err == nil {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: entity.ID,
			}
			entitiesMap[entity.ID] = entityMessage
		} else {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
			entitiesMap[entity.ID] = entityMessage
			break
		}
	}
	response, _ := util.JSONMarshal(entitiesMap)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func UpdateSettingProfilesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.Error(w, http.StatusInternalServerError, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	settingProfiles := logupload.SettingProfiles{}
	if body != "" {
		err := json.Unmarshal([]byte(body), &settingProfiles)
		if err != nil {
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
			return
		}
	}

	err = Update(&settingProfiles, applicationType)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(settingProfiles)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRuleNew error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func UpdateSettingProfilesPackageHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Unable to extract Body"))
		return
	}
	entities := []logupload.SettingProfiles{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract SettingProfiles from json file:" + err.Error()
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
		return
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		err := Update(&entity, applicationType)
		if err == nil {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: entity.ID,
			}
			entitiesMap[entity.ID] = entityMessage
		} else {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
			entitiesMap[entity.ID] = entityMessage
			break
		}
	}
	response, _ := util.JSONMarshal(entitiesMap)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}
