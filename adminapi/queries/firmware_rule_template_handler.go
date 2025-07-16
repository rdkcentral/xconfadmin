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

	"xconfadmin/common"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	xutil "github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"xconfadmin/adminapi/auth"
	xcommon "xconfadmin/common"
	xhttp "xconfadmin/http"
	"xconfadmin/util"

	"github.com/rdkcentral/xconfwebconfig/db"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
)

func GetFirmwareRuleTemplateFilteredHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	filterContext := make(map[string]string)
	util.AddQueryParamsToContextMap(r, filterContext)

	var err error
	allTemplates, err := corefw.GetFirmwareRuleTemplateAllAsListDBForAS("")
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	sort.Slice(allTemplates, func(i, j int) bool {
		return strings.Compare(strings.ToLower(allTemplates[i].ID), strings.ToLower(allTemplates[j].ID)) < 0
	})
	for k, v := range filterContext {
		if strings.ToUpper(k) == "KEY" {
			delete(filterContext, k)
			filterContext[firmware.KEY] = v
		}
		if strings.ToUpper(k) == "VALUE" {
			delete(filterContext, k)
			filterContext[firmware.VALUE] = v
		}
	}
	filteredRTsByAction := filterFirmwareRTsByContext(allTemplates, filterContext)
	allFilteredTemplates := []*corefw.FirmwareRuleTemplate{}
	for _, lst := range filteredRTsByAction {
		allFilteredTemplates = append(allFilteredTemplates, lst...)
	}
	response, err := xhttp.ReturnJsonResponse(allFilteredTemplates, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

// Usage on green splunk for 4 weeks ending 23rd Oc 2021
// /xconfadminService/ux/api/firmwareruletemplate/filtered?pageNumber=X&pageSize=Y 83
func PostFirmwareRuleTemplateFilteredHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	// Build the pageContext from query params
	pageContext := make(map[string]string)
	util.AddQueryParamsToContextMap(r, pageContext)

	// Build the filterContext from Body params
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	filterContext := map[string]string{}
	body := xw.Body()
	var err error
	if body != "" {
		err = json.Unmarshal([]byte(body), &filterContext)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	allTemplates, _ := corefw.GetFirmwareRuleTemplateAllAsListDBForAS("")
	sort.Slice(allTemplates, func(i, j int) bool {
		return strings.Compare(strings.ToLower(allTemplates[i].ID), strings.ToLower(allTemplates[j].ID)) < 0
	})

	actionContext := make(map[string]string)
	actionContext[cFirmwareRTApplicableActionType] = filterContext[cFirmwareRTApplicableActionType]
	templatesByAction := filterFirmwareRTsByContext(allTemplates, actionContext)[filterContext[cFirmwareRTApplicableActionType]]
	headers := make(map[string]string)
	headers["templateSizeByType"] = strconv.Itoa(len(templatesByAction))

	filteredTemplatesByType := filterFirmwareRTsByContext(allTemplates, filterContext)
	putSizesOfFirmwareRTsByTypeIntoHeaders(headers, filteredTemplatesByType)
	actionType, ok := util.FindEntryInContext(filterContext, cFirmwareRTApplicableActionType, true)
	filteredTemplates := []*corefw.FirmwareRuleTemplate{}
	if ok {
		filteredTemplates = filteredTemplatesByType[actionType]
	} else {
		for _, v := range filteredTemplatesByType {
			filteredTemplates = append(filteredTemplates, v...)
		}
	}

	sort.Slice(filteredTemplates, func(i, j int) bool {
		if filteredTemplates[i].Priority < filteredTemplates[j].Priority {
			return true
		}
		if filteredTemplates[i].Priority > filteredTemplates[j].Priority {
			return false
		}
		return filteredTemplates[i].ID < filteredTemplates[j].ID
	})

	filteredTemplates, err = generateFirmwareRTPageByContext(filteredTemplates, pageContext)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := xhttp.ReturnJsonResponse(filteredTemplates, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, response)
}

func PostFirmwareRuleTemplateImportAllHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	successTag := "IMPORTED"
	failedTag := "NOT_IMPORTED"

	/* If a JSON value is not appropriate for a given target type,
	 * or if a JSON number overflows the target type,
	 * Unmarshal skips that field and completes the unmarshalling as best it can.
	 * If no more serious errors are encountered, Unmarshal returns an UnmarshalTypeError describing the earliest such error.
	 */
	var firmwareRTs []corefw.FirmwareRuleTemplate
	if err := json.Unmarshal([]byte(xw.Body()), &firmwareRTs); err != nil {
		response := "Unable to extract firmwareruletemplate from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	for _, entity := range firmwareRTs {
		entity := entity
		if entity.Rule.Condition == nil || entity.Rule.Condition.FixedArg == nil {
			continue
		}
		if !entity.Rule.Condition.FixedArg.IsValid() {
			response := "Missing FixedArg:" + entity.ID
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
			return
		}
	}
	ds.GetCacheManager().ForceSyncChanges()
	result := importOrUpdateAllFirmwareRTs(firmwareRTs, successTag, failedTag)
	response, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func PostFirmwareRuleTemplateImportHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	successTag := "success"
	failedTag := "failure"
	result := make(map[string][]string)
	result[successTag] = []string{}
	result[failedTag] = []string{}

	type wrappedFrt struct {
		Entity    corefw.FirmwareRuleTemplate `json:"entity"`
		Overwrite bool                        `json:"overwrite"`
	}
	var wrappedFrts []wrappedFrt
	if err := json.Unmarshal([]byte(xw.Body()), &wrappedFrts); err != nil {
		response := "Unable to extract firmwareruletemplate from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	sort.Slice(wrappedFrts, func(i, j int) bool {
		if wrappedFrts[i].Entity.Priority < wrappedFrts[j].Entity.Priority {
			return true
		}
		if wrappedFrts[i].Entity.Priority > wrappedFrts[j].Entity.Priority {
			return false
		}
		return wrappedFrts[i].Entity.ID < wrappedFrts[j].Entity.ID
	})

	for _, wrapped := range wrappedFrts {
		entity := wrapped.Entity
		if entity.ID == "" {
			entity.ID = uuid.New().String()
		}
		entityOnDb, err := corefw.GetFirmwareRuleTemplateOneDB(entity.ID)
		if wrapped.Overwrite {
			if err != nil {
				result[failedTag] = append(result[failedTag], "FirmwareRuleTemplate with id '"+entity.ID+"' does not exist")
				continue
			}
			if err := updateFirmwareRT(entity, entityOnDb); err != nil {
				result[failedTag] = append(result[failedTag], "failed to import FirmwareRuleTemplate with id ="+entity.ID+", Error = "+err.Error())
			} else {
				result[successTag] = append(result[successTag], entity.ID)
			}
		} else {
			if err == nil {
				result[failedTag] = append(result[failedTag], "FirmwareRuleTemplate with id '"+entity.ID+"' already exists")
				continue
			}
			if _, err := createFirmwareRT(entity); err != nil {
				result[failedTag] = append(result[failedTag], "failed to import FirmwareRuleTemplate with id ="+entity.ID+", Error = "+err.Error())
			} else {
				result[successTag] = append(result[successTag], entity.ID)
			}
		}
	}

	response, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}
func PostChangePriorityHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	templateId, ok := mux.Vars(r)[common.ID]
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", common.ID))
		return
	}
	newPrioVar, ok := mux.Vars(r)[common.NEW_PRIORITY]
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", common.NEW_PRIORITY))
		return
	}

	frt, err := corefw.GetFirmwareRuleTemplateOneDB(templateId)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("unable to find template with id  %s", templateId))
		return
	}

	newPriority, err := strconv.Atoi(newPrioVar)
	if err != nil || newPriority <= 0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid priority value %s", newPrioVar))
		return
	}
	ds.GetCacheManager().ForceSyncChanges()
	firmwareRuleTemplateUpdateMutex.Lock()
	defer firmwareRuleTemplateUpdateMutex.Unlock()

	//TODO: basically this is the same action get all and filtered by action type
	allTemplates, _ := corefw.GetFirmwareRuleTemplateAllAsListDBForAS("")
	templatesOfCurrentType := firmwareRTFilterByActionType(allTemplates, string(frt.ApplicableAction.ActionType))
	if len(templatesOfCurrentType) == 0 {
		xhttp.WriteXconfResponse(w, http.StatusOK, nil)
		return
	}
	templatesOfCurrentTypeCopy := firmwareRuleTemplatesToPrioritizables(templatesOfCurrentType)
	reorganizedTemplates := UpdatePrioritizablesPriorities(templatesOfCurrentTypeCopy, int(frt.Priority), newPriority)

	if err = saveAllTemplates(reorganizedTemplates); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("unable to re-organize priorities: %s", err))
		return
	}
	res, err := xhttp.ReturnJsonResponse(reorganizedTemplates, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, res)
}
func PostFirmwareRuleTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}

	var firmwareRT corefw.FirmwareRuleTemplate

	if err := json.Unmarshal([]byte(xw.Body()), &firmwareRT); err != nil {
		response := "Unable to extract firmwareruletemplate from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}

	if xutil.IsBlank(firmwareRT.ID) {
		// ID is the name of the template so error if it is not specified
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "ID is required")
		return
	}

	_, err := corefw.GetFirmwareRuleTemplateOneDB(firmwareRT.ID)
	if err == nil {
		response := "firmwareRuleTemplate already exists for " + firmwareRT.ID
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, response)
		return
	}
	ds.GetCacheManager().ForceSyncChanges()
	if _, err = createFirmwareRT(firmwareRT); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	result, _ := corefw.GetFirmwareRuleTemplateOneDB(firmwareRT.ID)
	response, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusCreated, response)
}

func PutFirmwareRuleTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}

	var firmwareRT corefw.FirmwareRuleTemplate

	if err := json.Unmarshal([]byte(xw.Body()), &firmwareRT); err != nil {
		response := "Unable to extract firmwareruletemplate from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	ds.GetCacheManager().ForceSyncChanges()
	entityOnDb, err := corefw.GetFirmwareRuleTemplateOneDB(firmwareRT.ID)
	if err == nil {
		err = updateFirmwareRT(firmwareRT, entityOnDb)
	} else {
		response := "firmwareRuleTemplate does not exist for " + firmwareRT.ID
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	response := []byte{}
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func DeleteFirmwareRuleTemplateByIdHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	id, found := mux.Vars(r)[common.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	// Check for usage in FirmwareRule
	rules, err := corefw.GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil && err != common.NotFound {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "Unable to get Rules that use Config with id "+id)
		return
	}

	usedByRules := []string{}
	for _, rule := range rules {
		if rule.GetTemplateId() == id {
			usedByRules = append(usedByRules, rule.ApplicationType+"/"+rule.Name)
		}
	}
	if len(usedByRules) != 0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, "FirmwareRuleTemplate "+id+" is used by Rule(s): "+strings.Join(usedByRules[:], ","))
		return
	}
	ds.GetCacheManager().ForceSyncChanges()
	firmwareRuleTemplateUpdateMutex.Lock()
	defer firmwareRuleTemplateUpdateMutex.Unlock()
	templateToDelete, err := corefw.GetFirmwareRuleTemplateOneDBWithId(id)
	if err == nil {
		err = db.GetCachedSimpleDao().DeleteOne(db.TABLE_FIRMWARE_RULE_TEMPLATE, id)
	}
	if err != nil {
		response := "firmwareRuletemplate does not exist for " + id
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, response)
		return
	}

	allFrts, _ := corefw.GetFirmwareRuleTemplateAllAsListDBForAS("")
	actionContext := make(map[string]string)
	actionType := string(templateToDelete.ApplicableAction.ActionType)
	actionContext[cFirmwareRTApplicableActionType] = actionType
	templatesByAction := filterFirmwareRTsByContext(allFrts, actionContext)[actionType]

	templatesByActionCopy := firmwareRuleTemplatesToPrioritizables(templatesByAction)
	err = saveAllTemplates(PackPriorities(templatesByActionCopy, templateToDelete))
	if err != nil {
		response := "Failed to save firmwarerule templates after priority reorganization"
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, response)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusNoContent, []byte(""))
}

func PackFrtPriorities(allFrts []*corefw.FirmwareRuleTemplate, templateToDelete *corefw.FirmwareRuleTemplate) []*corefw.FirmwareRuleTemplate {
	alteredFrts := []*corefw.FirmwareRuleTemplate{}
	// sort by ascending priority
	sort.Slice(allFrts, func(i, j int) bool {
		return allFrts[i].Priority < allFrts[j].Priority
	})
	priority := 1
	for _, item := range allFrts {
		if item.ID == templateToDelete.ID {
			continue
		}
		oldpriority := item.Priority
		item.Priority = int32(priority)
		priority++
		if item.Priority != oldpriority {
			alteredFrts = append(alteredFrts, item)
		}
	}
	return alteredFrts
}

// Usage on green splunk for 4 weeks ending 23rd Oc 2021
// /xconfadminService/ux/api/firmwareruletemplate/{id} 579
func GetFirmwareRuleTemplateByIdHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	id, found := mux.Vars(r)[common.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	frt := GetFirmwareRuleTemplateById(id)
	if frt == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, fmt.Sprintf("unable to find FirmwareRuleTemplate with id : %v", id))
		return
	}
	queryParams := r.URL.Query()
	_, ok := queryParams[xcommon.EXPORT]
	if ok {
		frtList := []corefw.FirmwareRuleTemplate{*frt}
		res, err := xhttp.ReturnJsonResponse(frtList, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		fileName := xcommon.ExportFileNames_FIRMWARE_RULE_TEMPLATE + frt.ID
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
		return
	}
	res, err := xhttp.ReturnJsonResponse(frt, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func PostFirmwareRuleTemplateEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "Unable to extract body")
		return
	}
	body := xw.Body()
	entities := []corefw.FirmwareRuleTemplate{}
	err := json.Unmarshal([]byte(body), &entities)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].ApplicableAction.ActionType < entities[j].ApplicableAction.ActionType {
			return true
		}
		if entities[i].ApplicableAction.ActionType > entities[j].ApplicableAction.ActionType {
			return false
		}
		return entities[i].Priority < entities[j].Priority
	})
	ds.GetCacheManager().ForceSyncChanges()
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		_, err := corefw.GetFirmwareRuleTemplateOneDB(entity.ID)
		if err == nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: "firmwareRuleTemplate " + entity.ID + " already present.",
			}
			continue
		}
		if _, err = createFirmwareRT(entity); err != nil {
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
	response, err := xhttp.ReturnJsonResponse(entitiesMap, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func PutFirmwareRuleTemplateEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract body")
		return
	}
	body := xw.Body()
	entities := []corefw.FirmwareRuleTemplate{}
	err := json.Unmarshal([]byte(body), &entities)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].ApplicableAction.ActionType < entities[j].ApplicableAction.ActionType {
			return true
		}
		if entities[i].ApplicableAction.ActionType > entities[j].ApplicableAction.ActionType {
			return false
		}
		return entities[i].Priority < entities[j].Priority
	})
	ds.GetCacheManager().ForceSyncChanges()
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entityOnDb, err := corefw.GetFirmwareRuleTemplateOneDB(entity.ID)
		if err != nil || entityOnDb == nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
			continue
		}
		if err := updateFirmwareRT(entity, entityOnDb); err != nil {
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
	response, err := xhttp.ReturnJsonResponse(entitiesMap, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func ObsoleteGetFirmwareRuleTemplatePageHandler(w http.ResponseWriter, r *http.Request) {
	pageContext := map[string]string{}
	util.AddQueryParamsToContextMap(r, pageContext)

	dbrules, _ := corefw.GetFirmwareRuleTemplateAllAsListDBForAS("")
	sort.Slice(dbrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(dbrules[i].ID), strings.ToLower(dbrules[j].ID)) < 0
	})
	headers := putSizesOfFirmwareRTsByTypeIntoHeaders2(dbrules)
	var err error
	dbrules, err = generateFirmwareRTPageByContext(dbrules, pageContext)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := xhttp.ReturnJsonResponse(dbrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, response)
}

func GetFirmwareRuleTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	dbrules, _ := corefw.GetFirmwareRuleTemplateAllAsListDBForAS("")
	sort.Slice(dbrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(dbrules[i].ID), strings.ToLower(dbrules[j].ID)) < 0
	})
	res, err := xhttp.ReturnJsonResponse(dbrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	queryParams := r.URL.Query()
	_, ok1 := queryParams[xcommon.EXPORT]
	_, ok2 := queryParams[xcommon.EXPORTALL]
	if ok1 || ok2 {
		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_FIRMWARE_RULE_TEMPLATES)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

// Usage on green splunk for 4 weeks ending 23rd Oc 2021
// /xconfadminService/ux/api/firmwareruletemplate/all/{type} 133
func GetFirmwareRuleTemplateAllByTypeHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	applicableActionType, ok := mux.Vars(r)[xcommon.TYPE]
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Unable to decipher %s", xcommon.TYPE))
		return
	}
	dbrules, _ := corefw.GetFirmwareRuleTemplateAllAsListDBForAS("")
	tempIds := []corefw.FirmwareRuleTemplate{}
	for _, v := range dbrules {
		if string(v.ApplicableAction.ActionType) == applicableActionType {
			tempIds = append(tempIds, *v)
		}
	}
	res, err := xhttp.ReturnJsonResponse(tempIds, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

// Usage on green splunk for 4 weeks ending 23rd Oc 2021
// /xconfadminService/ux/api/firmwareruletemplate/ids?type=applicationType 2817
// Non zero usage - Has to be migrated soon
func GetFirmwareRuleTemplateIdsHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	queryParams := r.URL.Query()
	applicableActionTypes, ok := queryParams[xcommon.TYPE]
	if ok {
		applicableActionType := applicableActionTypes[0]
		dbrules, _ := corefw.GetFirmwareRuleTemplateAllAsListDBForAS("")
		tempIds := []string{}
		for _, v := range dbrules {
			if string(v.ApplicableAction.ActionType) == applicableActionType && v.Editable {
				tempIds = append(tempIds, v.ID)
			}
		}
		sort.Strings(tempIds)
		res, err := xhttp.ReturnJsonResponse(tempIds, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
		return
	}
	// xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "type query param not found")
	// Java returns NotFound and so are we, though BadRequest would have been better
	xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "type query param not found")
}

func GetFirmwareRuleTemplateWithVarWithVarHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	applicableActionType, ok := mux.Vars(r)[xcommon.TYPE]
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", xcommon.TYPE))
		return
	}
	editVar, ok := mux.Vars(r)[xcommon.EDITABLE]
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", xcommon.EDITABLE))
		return
	}
	editable := false
	if editVar == "true" {
		editable = true
	}
	dbrules, _ := corefw.GetFirmwareRuleTemplateAllAsListDBForAS("")
	tempIds := []corefw.FirmwareRuleTemplate{}
	for _, v := range dbrules {
		if string(v.ApplicableAction.ActionType) == applicableActionType && editable == v.Editable {
			tempIds = append(tempIds, *v)
		}
	}
	sort.Slice(tempIds, func(i, j int) bool {
		return strings.Compare(strings.ToLower(tempIds[i].ID), strings.ToLower(tempIds[j].ID)) < 0
	})
	res, err := xhttp.ReturnJsonResponse(tempIds, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetFirmwareRuleTemplateExportHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	queryParams := r.URL.Query()
	actionTypes, ok := queryParams[xcommon.TYPE]
	if ok {
		actionType := actionTypes[0]
		entities, _ := corefw.GetFirmwareRuleTemplateAllAsListDBForAS("")
		dbrules := []corefw.FirmwareRuleTemplate{}
		for _, v := range entities {
			if string(v.ApplicableAction.ActionType) == actionType {
				dbrules = append(dbrules, *v)
			}
		}
		res, err := xhttp.ReturnJsonResponse(dbrules, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		if actionType == string(corefw.RULE_TEMPLATE) {
			actionType = strings.Replace(actionType, "_TEMPLATE", "_ACTION_TEMPLATE", 1)
		}
		fileName := xcommon.ExportFileNames_ALL_FIRMWARE_RULE_TEMPLATES + "_" + actionType
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
		return
	}
	// Java returns NotFound and so are we, though BadRequest would have been better
	xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "type query param not found")
}

// // Usage on green splunk for 4 weeks ending 23rd Oc 2021
// // /xconfadminService/ux/api/firmwareruletemplate/{id}/priority/9 1
// func PostFirmwareRuleTemplateByIdPriorityByNewPriorityHandler(w http.ResponseWriter, r *http.Request) {
// 	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
// 		xhttp.AdminError(w, err)
// 		return
// 	}
// 	templateId, ok := mux.Vars(r)[common.ID]
// 	if !ok {
// 		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", common.ID))
// 		return
// 	}
// 	newPrioVar, ok := mux.Vars(r)[xcommon.NEW_PRIORITY]
// 	if !ok {
// 		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", xcommon.NEW_PRIORITY))
// 		return
// 	}

// 	frt, err := corefw.GetFirmwareRuleTemplateOneDB(templateId)
// 	if err != nil {
// 		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("unable to find template with id  %s", templateId))
// 		return
// 	}

// 	newPriority, err := strconv.Atoi(newPrioVar)
// 	if err != nil {
// 		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Incorrect priority value  for %s", newPrioVar))
// 		return
// 	}

// 	dbrules, _ := corefw.GetFirmwareRuleTemplateAllAsListDB("")
// 	templatesOfCurrentType := firmwareRTFilterByActionType(dbrules, string(frt.ApplicableAction.ActionType))
// 	if len(templatesOfCurrentType) == 0 {
// 		xwhttp.WriteXconfResponse(w, http.StatusOK, nil)
// 		return
// 	}
// 	reorganizedTemplates, err := updateFirmwareRTByPriorityAndReorganize(frt, templatesOfCurrentType, newPriority)
// 	if err != nil {
// 		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("unable to re-organize priorities: %s", err))
// 		return
// 	}
// 	if err = saveAllTemplates(reorganizedTemplates); err != nil {
// 		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("unable to re-organize priorities: %s", err))
// 		return
// 	}
// 	res, err := xhttp.ReturnJsonResponse(reorganizedTemplates, r)
// 	if err != nil {
// 		xhttp.AdminError(w, err)
// 		return
// 	}
// 	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
// }
