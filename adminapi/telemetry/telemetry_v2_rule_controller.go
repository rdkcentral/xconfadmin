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

	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	xcommon "xconfadmin/common"
	xwcommon "xconfwebconfig/common"

	xlogupload "xconfadmin/shared/logupload"
	"xconfwebconfig/common"
	xwlogupload "xconfwebconfig/shared/logupload"
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

func GetTelemetryTwoRulesAllExport(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	all := GetAll()
	telemetryTwoRules := []*xwlogupload.TelemetryTwoRule{}
	for _, entity := range all {
		if entity.ApplicationType == applicationType {
			telemetryTwoRules = append(telemetryTwoRules, entity)
		}
	}

	response, err := util.JSONMarshal(telemetryTwoRules)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal telemetryTwoRules error: %v", err))
	}
	_, ok := r.URL.Query()["export"]
	if ok {
		fileName := "allTelemetryTwoRules_" + applicationType
		headerMap := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
	}
}

func GetTelemetryTwoRuleById(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	id, found := mux.Vars(r)[common.ID]
	if !found || len(strings.TrimSpace(id)) == 0 {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Id is blank"))
		return
	}
	telemetryTwoRule := xlogupload.GetOneTelemetryTwoRule(id)
	if telemetryTwoRule == nil {
		invalid := "Entity with id: " + id + " does not exist"
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, invalid)
		return
	}

	if _, ok := r.URL.Query()["export"]; ok {
		ruleToExport := []*xwlogupload.TelemetryTwoRule{telemetryTwoRule}
		res, err := xhttp.ReturnJsonResponse(ruleToExport, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		fileName := "telemetryTwoRule_" + telemetryTwoRule.ID + "_" + applicationType
		headerMap := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, res)
	} else {
		res, err := xhttp.ReturnJsonResponse(telemetryTwoRule, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
	}
}

func GetAllTelemetryTwoRulesWithPage(w http.ResponseWriter, r *http.Request) {
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
	telemetryTwoRules := GetAll()
	telemetryTwoRuleList := TelemetryTwoRulesGeneratePage(telemetryTwoRules, pageNumber, pageSize)
	response, err := util.JSONMarshal(telemetryTwoRuleList)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRules error: %v", err))
	}
	headerMap := createNumberOfItemsHttpHeaders(telemetryTwoRules)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func createNumberOfItemsHttpHeaders(entities []*xwlogupload.TelemetryTwoRule) map[string]string {
	headerMap := make(map[string]string, 1)
	if entities == nil {
		headerMap[NumberOfItems] = "0"
	} else {
		headerMap[NumberOfItems] = strconv.Itoa(len(entities))
	}
	return headerMap
}

func DeleteOneTelemetryTwoRuleHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.ID]
	if !found || util.IsBlank(id) {
		xwhttp.WriteXconfResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}
	_, err := Delete(id)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

func GetTelemetryTwoRulesFilteredWithPage(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
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
		if err := json.Unmarshal([]byte(xw.Body()), &contextMap); err != nil {
			response := "Unable to extract searchContext from json file:" + err.Error()
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
			return
		}
	}
	contextMap[xwcommon.APPLICATION_TYPE] = applicationType

	telemetryTwoRules := findByContext(r, contextMap)
	sort.SliceStable(telemetryTwoRules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(telemetryTwoRules[i].Name), strings.ToLower(telemetryTwoRules[j].Name)) < 0
	})
	telemetryTwoRulesList := TelemetryTwoRulesGeneratePage(telemetryTwoRules, pageNumber, pageSize)
	response, err := util.JSONMarshal(telemetryTwoRulesList)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal telemetryTwoRules error: %v", err))
	}
	headerMap := createNumberOfItemsHttpHeaders(telemetryTwoRules)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func CreateTelemetryTwoRuleHandler(w http.ResponseWriter, r *http.Request) {
	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.Error(w, http.StatusInternalServerError, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	telemetry2Rule := xwlogupload.TelemetryTwoRule{}
	err := json.Unmarshal([]byte(body), &telemetry2Rule)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}

	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	err = Create(&telemetry2Rule, applicationType)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(telemetry2Rule)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal telemetry2Rule error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusCreated, response)
}

func CreateTelemetryTwoRulesPackageHandler(w http.ResponseWriter, r *http.Request) {
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Unable to extract Body"))
		return
	}
	entities := []xwlogupload.TelemetryTwoRule{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract TelemetryTwoRules from json file:" + err.Error()
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
		return
	}

	applicationType, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
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
		}
	}
	response, _ := util.JSONMarshal(entitiesMap)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func UpdateTelemetryTwoRuleHandler(w http.ResponseWriter, r *http.Request) {
	writeApplication, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
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
	telemetryTwoRule := xwlogupload.TelemetryTwoRule{}
	err = json.Unmarshal([]byte(body), &telemetryTwoRule)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	err = Update(&telemetryTwoRule, writeApplication)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(telemetryTwoRule)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal telemetryTwoRule error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func UpdateTelemetryTwoRulesPackageHandler(w http.ResponseWriter, r *http.Request) {
	writeApplication, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Unable to extract Body"))
		return
	}
	entities := []xwlogupload.TelemetryTwoRule{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract TelemetryTwoRules from json file:" + err.Error()
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
		return
	}
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		err := Update(&entity, writeApplication)
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
		}
	}
	response, _ := util.JSONMarshal(entitiesMap)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}
