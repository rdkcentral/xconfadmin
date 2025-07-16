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
package setting

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	//"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	xcommon "xconfadmin/common"
	"xconfadmin/shared"

	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
)

func GetSettingRulesAllExport(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
	}
	all := GetAllSettingRules()
	settingRules := []*logupload.SettingRule{}
	for _, entity := range all {
		if entity.ApplicationType == applicationType {
			settingRules = append(settingRules, entity)
		}
	}
	response, err := util.JSONMarshal(settingRules)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal settingRules error: %v", err))
	}
	_, ok := r.URL.Query()[xcommon.EXPORT]
	if ok {
		headerMap := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_SETTING_RULES)
		xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
	}
}

func GetSettingRuleOneExport(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
	}
	id, found := mux.Vars(r)[xwcommon.ID]
	if !found || len(strings.TrimSpace(id)) == 0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is blank")
		return
	}
	settingRule, _ := GetOneSettingRule(id)
	if settingRule == nil {
		invalid := "Entity with id: " + id + " does not exist"
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, invalid)
		return
	}

	_, ok := r.URL.Query()[xcommon.EXPORT]
	if ok {
		srList := []logupload.SettingRule{*settingRule}
		response, err := util.JSONMarshal(srList)
		if err != nil {
			log.Error(fmt.Sprintf("json.Marshal settingProfile error: %v", err))
		}
		headerMap := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_SETTING_RULE + settingRule.ID)
		xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
	} else {
		response, err := util.JSONMarshal(settingRule)
		if err != nil {
			log.Error(fmt.Sprintf("json.Marshal settingProfile error: %v", err))
		}
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
	}
}

func DeleteOneSettingRulesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found || util.IsBlank(id) {
		xwhttp.WriteXconfResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}
	_, err = DeleteSettingRule(id, applicationType)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

func GetSettingRulesFilteredWithPage(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
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
		xwhttp.Error(w, http.StatusInternalServerError, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
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
	contextMap[xwcommon.APPLICATION_TYPE] = applicationType

	settingRules := FindByContextSettingRule(r, contextMap)
	sort.Slice(settingRules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(settingRules[i].Name), strings.ToLower(settingRules[j].Name)) < 0
	})
	settingRulesList := SettingRulesGeneratePage(settingRules, pageNumber, pageSize)
	response, err := util.JSONMarshal(settingRulesList)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRules error: %v", err))
	}
	headerMap := createNumberOfSettingRulesHttpHeaders(settingRules)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func CreateSettingRuleHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	settingRules := logupload.SettingRule{}
	if body != "" {
		err := json.Unmarshal([]byte(body), &settingRules)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	err = CreateSettingRule(r, &settingRules)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(settingRules)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal settingRules error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusCreated, response)
}

func CreateSettingRulesPackageHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	entities := []logupload.SettingRule{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract SettingRules from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		err := CreateSettingRule(r, &entity)
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

func UpdateSettingRulesHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	settingRules := logupload.SettingRule{}
	if body != "" {
		err := json.Unmarshal([]byte(body), &settingRules)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	err = UpdateSettingRule(r, &settingRules)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(settingRules)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRuleNew error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func UpdateSettingRulesPackageHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Unable to extract Body"))
		return
	}
	entities := []logupload.SettingRule{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract FeatureRules from json file:" + err.Error()
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
		return
	}
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		err := UpdateSettingRule(r, &entity)
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

func createNumberOfSettingRulesHttpHeaders(entities []*logupload.SettingRule) map[string]string {
	headerMap := make(map[string]string, 1)
	if entities == nil {
		headerMap[NumberOfItems] = "0"
	} else {
		headerMap[NumberOfItems] = strconv.Itoa(len(entities))
	}
	return headerMap
}

func SettingTestPageHandler(w http.ResponseWriter, r *http.Request) {
	settingTypes := r.URL.Query()[xwcommon.SETTING_TYPE]
	if len(settingTypes) == 0 {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Define settings type"))
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

	if err := shared.NormalizeCommonContext(contextMap, xwcommon.ESTB_MAC_ADDRESS, xwcommon.ECM_MAC_ADDRESS); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	contextMap[xwcommon.APPLICATION_TYPE] = applicationType

	result := make(map[string]interface{})
	result["result"] = GetSettingRulesWithConfig(settingTypes, contextMap)
	result["context"] = contextMap

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}
