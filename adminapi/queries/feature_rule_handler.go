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
package queries

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"xconfadmin/shared"
	xshared "xconfadmin/shared"
	"xconfwebconfig/db"
	xwhttp "xconfwebconfig/http"

	"xconfadmin/adminapi/auth"
	xcommon "xconfadmin/common"
	xhttp "xconfadmin/http"
	xrfc "xconfadmin/shared/rfc"
	"xconfwebconfig/common"
	xwcommon "xconfwebconfig/common"
	ds "xconfwebconfig/db"
	"xconfwebconfig/shared/rfc"
	"xconfwebconfig/util"
)

const (
	NewPriority   = "newPriority"
	PageNumber    = "pageNumber"
	PageSize      = "pageSize"
	NumberOfItems = "numberOfItems"
)

func GetFeatureRulesFiltered(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	queryParams := r.URL.Query()
	contextMap := make(map[string]string)
	if len(queryParams) > 0 {
		for k, v := range queryParams {
			contextMap[k] = v[0]
		}
	}
	contextMap[common.APPLICATION_TYPE] = applicationType

	featureRules := FindFeatureRuleByContext(contextMap)
	response, err := util.JSONMarshal(featureRules)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRules error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetFeatureRulesFilteredWithPage(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
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
		xwhttp.Error(w, http.StatusInternalServerError, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	contextMap := make(map[string]string)
	bodyStr := xw.Body()
	if bodyStr != "" {
		if err := json.Unmarshal([]byte(xw.Body()), &contextMap); err != nil {
			response := "Unable to extract searchContext from json file:" + err.Error()
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
			return
		}
	}
	contextMap[common.APPLICATION_TYPE] = applicationType

	featureRules := FindFeatureRuleByContext(contextMap)
	featureRuleList := FeatureRulesGeneratePage(featureRules, pageNumber, pageSize)
	response, err := util.JSONMarshal(featureRuleList)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRules error: %v", err))
	}
	headerMap := createNumberOfItemsHttpHeaders(featureRules)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func FeatureRulesGeneratePage(list []*rfc.FeatureRule, page int, pageSize int) []*rfc.FeatureRule {
	result := []*rfc.FeatureRule{}
	leng := len(list)
	startIndex := page*pageSize - pageSize
	if page < 1 || startIndex > leng || pageSize < 1 {
		return result
	}
	lastIndex := leng
	if page*pageSize < len(list) {
		lastIndex = page * pageSize
	}
	return list[startIndex:lastIndex]
}

func createNumberOfItemsHttpHeaders(entities []*rfc.FeatureRule) map[string]string {
	headerMap := make(map[string]string, 1)
	if entities == nil {
		headerMap[NumberOfItems] = "0"
	} else {
		headerMap[NumberOfItems] = strconv.Itoa(len(entities))
	}
	return headerMap
}

func GetFeatureRulesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	featureRules := GetAllFeatureRulesByType(applicationType)
	response, err := util.JSONMarshal(featureRules)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRules error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetFeatureRulesExportHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	featureRules := GetAllFeatureRulesByType(applicationType)
	sort.Slice(featureRules, func(i, j int) bool {
		return featureRules[j].Priority > featureRules[i].Priority
	})

	response, err := util.JSONMarshal(featureRules)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRules error: %v", err))
	}
	_, ok := r.URL.Query()[xcommon.EXPORT]
	if ok {
		headerMap := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_FEATURE_RUlES)
		xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
	}
}

func GetFeatureRuleOneExport(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	id, found := mux.Vars(r)[common.ID]
	if !found || len(strings.TrimSpace(id)) == 0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is blank")
		return
	}
	featureRule := GetOne(id)
	if featureRule == nil {
		invalid := "Entity with id: " + id + " does not exist"
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, invalid)
		return
	}
	if featureRule.ApplicationType != applicationType {
		invalid := "Non existing Entity in application with id: " + id
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, invalid)
		return
	}
	_, ok := r.URL.Query()[xcommon.EXPORT]
	if ok {
		featureList := []*rfc.FeatureRule{featureRule}
		response, err := util.JSONMarshal(featureList)
		if err != nil {
			log.Error(fmt.Sprintf("json.Marshal featureRule error: %v", err))
		}
		headerMap := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_FEATURE_RULE + featureRule.Id)
		xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
	} else {
		response, err := util.JSONMarshal(featureRule)
		if err != nil {
			log.Error(fmt.Sprintf("json.Marshal featureRule error: %v", err))
		}
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
	}
}

func GetFeatureRuleOne(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.ID]
	if !found || len(strings.TrimSpace(id)) == 0 {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Id is blank"))
		return
	}
	featureRule := GetOne(id)
	if featureRule == nil {
		invalid := "Entity with id: " + id + " does not exist"
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, invalid)
		return
	}

	if err := auth.ValidateRead(r, featureRule.ApplicationType, auth.FIRMWARE_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return

	}

	response, err := util.JSONMarshal(featureRule)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRule error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func CreateFeatureRuleHandler(w http.ResponseWriter, r *http.Request) {
	featureRule := rfc.FeatureRule{}
	applicationType, err := auth.ExtractBodyAndCheckPermissions(&featureRule, w, r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	db.GetCacheManager().ForceSyncChanges()
	createdFeatureRule, err := CreateFeatureRule(featureRule, applicationType)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(createdFeatureRule)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRuleNew error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusCreated, response)
}

func UpdateFeatureRuleHandler(w http.ResponseWriter, r *http.Request) {
	featureRule := rfc.FeatureRule{}
	applicationType, err := auth.ExtractBodyAndCheckPermissions(&featureRule, w, r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	ds.GetCacheManager().ForceSyncChanges()
	updatedFeatureRule, err := UpdateFeatureRule(featureRule, applicationType)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(updatedFeatureRule)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRuleNew error: %v", err))
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func ImportAllFeatureRulesHandler(w http.ResponseWriter, r *http.Request) {
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.Error(w, http.StatusInternalServerError, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	var featureRules []rfc.FeatureRule
	if err := json.Unmarshal([]byte(xw.Body()), &featureRules); err != nil {
		response := "Unable to extract featureRules from json file: " + err.Error()
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
		return
	}

	ds.GetCacheManager().ForceSyncChanges()

	determinedAppType := ""
	for i, _ := range featureRules {
		applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY, featureRules[i].ApplicationType)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		if determinedAppType != "" && determinedAppType != applicationType {
			xhttp.WriteAdminErrorResponse(w, http.StatusConflict, "ApplicationType mixing not allowed")
			return
		}
		if featureRules[i].ApplicationType == "" {
			featureRules[i].ApplicationType = applicationType
		} else if featureRules[i].ApplicationType != applicationType {
			xhttp.WriteAdminErrorResponse(w, http.StatusConflict, "ApplicationType Conflict")
			return
		}
		determinedAppType = applicationType
	}

	sort.Slice(featureRules, func(i, j int) bool {
		return featureRules[i].Priority < featureRules[j].Priority
	})

	importResult := ImportOrUpdateAllFeatureRule(featureRules, determinedAppType)
	response, err := util.JSONMarshal(importResult)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRuleNew error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func DeleteOneFeatureRuleHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.ID]
	if !found || util.IsBlank(id) {
		xwhttp.WriteXconfResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}
	ds.GetCacheManager().ForceSyncChanges()
	featureRuleUpdateMutex.Lock()
	defer featureRuleUpdateMutex.Unlock()
	featureRuleToDelete := GetOne(id)
	if featureRuleToDelete == nil {
		xwhttp.WriteXconfResponse(w, http.StatusNotFound, []byte("\"Entity with id: "+id+" does not exist\""))
		return
	}

	if err := auth.ValidateWrite(r, featureRuleToDelete.ApplicationType, auth.FIRMWARE_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xrfc.DeleteFeatureRule(id)

	context := map[string]string{shared.APPLICATION_TYPE: featureRuleToDelete.ApplicationType}
	prioritizableRules := FeatureRulesToPrioritizables(FindFeatureRuleByContext(context))
	err := SaveFeatureRules(PackPriorities(prioritizableRules, featureRuleToDelete))
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusNoContent, []byte(""))
}

func PackFeaturePriorities(allFeatures []*rfc.FeatureRule, featureToDelete *rfc.FeatureRule) []*rfc.FeatureRule {
	altered := []*rfc.FeatureRule{}
	// sort by ascending priority
	sort.Slice(allFeatures, func(i, j int) bool {
		return allFeatures[i].Priority < allFeatures[j].Priority
	})
	priority := 1
	for _, item := range allFeatures {
		if item.Id == featureToDelete.Id {
			continue
		}
		if item.ApplicationType != featureToDelete.ApplicationType {
			continue
		}
		oldpriority := item.Priority
		item.Priority = priority
		priority++
		if item.Priority != oldpriority {
			altered = append(altered, item)
		}
	}
	return altered
}

func ChangeFeatureRulePrioritiesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[common.ID]
	if !found || len(strings.TrimSpace(id)) == 0 {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Id is blank"))
		return
	}
	newPriority, found := mux.Vars(r)[NewPriority]
	if !found || len(strings.TrimSpace(NewPriority)) == 0 {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Id is blank"))
		return
	}
	newPriorityInt, err := strconv.Atoi(newPriority)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("newPriority must be a number"))
		return
	}
	db.GetCacheManager().ForceSyncChanges()

	featureRuleToUpdate := GetOne(id)
	if featureRuleToUpdate == nil {
		xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte("FeatureRule with id: "+id+" does not exist"))
	}

	featureRuleUpdateMutex.Lock()
	defer featureRuleUpdateMutex.Unlock()
	featureRules, err := ChangePrioritizablePriorities(featureRuleToUpdate, newPriorityInt, applicationType)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(featureRules)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRules error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetFeatureRulesSizeHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	size := GetFeatureRulesSize(applicationType)
	sizeString := strconv.Itoa(size)
	response, err := util.JSONMarshal(sizeString)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal FeatureRulesSize error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetAllowedNumberOfFeaturesHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	allowedNumber := GetAllowedNumberOfFeatures()
	response, err := util.JSONMarshal(allowedNumber)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal AllowedNumberOfFeatures error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func UpdateFeatureRulesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Unable to extract Body"))
		return
	}
	entities := []rfc.FeatureRule{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract FeatureRules from json file:" + err.Error()
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
		return
	}

	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Priority < entities[j].Priority
	})
	db.GetCacheManager().ForceSyncChanges()
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		_, err := UpdateFeatureRule(entity, applicationType)
		if err == nil {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: entity.Id,
			}
			entitiesMap[entity.Id] = entityMessage
		} else {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
			entitiesMap[entity.Id] = entityMessage
		}
	}
	response, _ := util.JSONMarshal(entitiesMap)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func CreateFeatureRulesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Unable to extract Body"))
		return
	}
	entities := []rfc.FeatureRule{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract FeatureRules from json file:" + err.Error()
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
		return
	}

	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Priority < entities[j].Priority
	})
	db.GetCacheManager().ForceSyncChanges()
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		featureRule, err := CreateFeatureRule(entity, applicationType)
		if err == nil {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: featureRule.Id,
			}
			entitiesMap[entity.Id] = entityMessage
		} else {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
			entitiesMap[entity.Id] = entityMessage
		}
	}
	response, _ := util.JSONMarshal(entitiesMap)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func FeatureRuleTestPageHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
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
	fields := xw.Audit()

	var contextMap map[string]string
	if err := json.Unmarshal([]byte(body), &contextMap); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := xshared.NormalizeCommonContext(contextMap, common.ESTB_MAC_ADDRESS, common.ECM_MAC_ADDRESS); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	contextMap[common.APPLICATION_TYPE] = applicationType

	result := ProcessFeatureRules(contextMap, fields)

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}
