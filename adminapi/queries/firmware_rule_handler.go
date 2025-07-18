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
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xcommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	xfirmware "github.com/rdkcentral/xconfadmin/shared/firmware"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	xutil "github.com/rdkcentral/xconfadmin/util"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
)

func populateContext(w http.ResponseWriter, r *http.Request, isRead bool) (filterContext map[string]string, err error) {
	filterContext = map[string]string{}
	xutil.AddQueryParamsToContextMap(r, filterContext)
	appType, found := filterContext[common.APPLICATION_TYPE]
	if !found || util.IsBlank(appType) {
		if isRead {
			filterContext[common.APPLICATION_TYPE], err = auth.CanRead(r, auth.FIRMWARE_ENTITY)
		} else {
			filterContext[common.APPLICATION_TYPE], err = auth.CanWrite(r, auth.FIRMWARE_ENTITY)
		}
		if err != nil {
			return filterContext, err
		}
	}
	return filterContext, nil
}

// Zero Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
func GetFirmwareRuleFilteredHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	dbrules, err := xfirmware.GetFirmwareSortedRuleAllAsListDB()
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	filterContext, err := populateContext(w, r, true)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	filterContext[common.APPLICATION_TYPE] = appType

	for k, v := range filterContext {
		if strings.ToUpper(k) == "KEY" {
			delete(filterContext, k)
			filterContext[firmware.KEY] = v
		}
		if strings.ToUpper(k) == "VALUE" {
			delete(filterContext, k)
			filterContext[firmware.VALUE] = v
		}
		if strings.ToUpper(k) == "TEMPLATEID" {
			delete(filterContext, k)
			filterContext[cFirmwareRuleTemplateId] = v
		}
	}
	dbrules = filterFirmwareRulesByContext(dbrules, filterContext)

	response, err := xhttp.ReturnJsonResponse(dbrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

// Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
// 10946 POST /xconfadminService/ux/api/firmwarerule/filtered?pageNumber=X&pageSize=Y 10946
func PostFirmwareRuleFilteredHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// Build the pageContext from query params
	pageContext, err := populateContext(w, r, false)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Build the filterContext from Body
	filterContext := make(map[string]string)
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract body")
		return
	}
	body := xw.Body()
	if body != "" {
		err = json.Unmarshal([]byte(body), &filterContext)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	filterContext[common.APPLICATION_TYPE] = applicationType

	// Get all sorted rules
	dbrules, err := firmware.GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	sort.Slice(dbrules, func(i, j int) bool {
		if strings.Compare(strings.ToLower(dbrules[i].Name), strings.ToLower(dbrules[j].Name)) < 0 {
			return true
		}
		if strings.Compare(strings.ToLower(dbrules[i].Name), strings.ToLower(dbrules[j].Name)) > 0 {
			return false
		}
		return strings.Compare(strings.ToLower(dbrules[i].ID), strings.ToLower(dbrules[j].ID)) < 0
	})

	appFilter := map[string]string{xcommon.APPLICABLE_ACTION_TYPE: filterContext[xcommon.APPLICABLE_ACTION_TYPE]}
	delete(filterContext, xcommon.APPLICABLE_ACTION_TYPE)
	// Filter the entries according to filterContext
	dbrules = filterFirmwareRulesByContext(dbrules, filterContext)

	// Populate the headers
	headers := putSizesOfFirmwareRulesByTypeIntoHeaders(dbrules)

	// Filter the entries according to appFilter
	dbrules = filterFirmwareRulesByContext(dbrules, appFilter)

	// Get entries from the requested page according to pageContext
	dbrules, err = generateFirmwareRulePageByContext(dbrules, pageContext)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return the response
	response, err := xhttp.ReturnJsonResponse(dbrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, response)
}

// Zero Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
func PostFirmwareRuleImportAllHandler(w http.ResponseWriter, r *http.Request) {
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}

	fields := xw.Audit()
	const LargeValue = 1024
	var firmwareRules = make([]*firmware.FirmwareRule, LargeValue)
	for i := range firmwareRules {
		firmwareRules[i] = firmware.NewEmptyFirmwareRule()
		firmwareRules[i].ApplicationType = ""
	}

	if err := json.Unmarshal([]byte(xw.Body()), &firmwareRules); err != nil {
		response := "Unable to extract firmwarerule from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	determinedAppType := ""
	for i, _ := range firmwareRules {
		appType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY, firmwareRules[i].ApplicationType)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		if determinedAppType != "" && determinedAppType != appType {
			xhttp.WriteAdminErrorResponse(w, http.StatusConflict, "ApplicationType mixing not allowed")
			return
		}
		if firmwareRules[i].ApplicationType == "" {
			firmwareRules[i].ApplicationType = appType
		} else if firmwareRules[i].ApplicationType != appType {
			xhttp.WriteAdminErrorResponse(w, http.StatusConflict, "ApplicationType Conflict")
			return
		}
		determinedAppType = appType
	}
	result := importOrUpdateAllFirmwareRules(firmwareRules, determinedAppType, fields)
	response, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, response)
}

// Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
//
//	151 POST /xconfadminService/ux/api/firmwarerule/
func PostFirmwareRuleHandler(w http.ResponseWriter, r *http.Request) {
	firmwareRule := firmware.NewEmptyFirmwareRule()
	firmwareRule.ApplicationType = ""
	appType, err := auth.ExtractBodyAndCheckPermissions(firmwareRule, w, r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	if util.IsBlank(firmwareRule.ID) {
		firmwareRule.ID = uuid.New().String()
	} else {
		_, err = firmware.GetFirmwareRuleOneDB(firmwareRule.ID)
		if err == nil {
			response := "firmwareRule already exists for " + firmwareRule.ID
			xhttp.WriteAdminErrorResponse(w, http.StatusConflict, response)
			return
		}
	}
	err = createFirmwareRule(*firmwareRule, appType, true)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	result, _ := firmware.GetFirmwareRuleOneDB(firmwareRule.ID)
	response, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusCreated, response)
}

// Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
//
//	529 PUT /xconfadminService/ux/api/firmwarerule/
func PutFirmwareRuleHandler(w http.ResponseWriter, r *http.Request) {
	firmwareRule := firmware.NewEmptyFirmwareRule()
	firmwareRule.ApplicationType = ""
	appType, err := auth.ExtractBodyAndCheckPermissions(firmwareRule, w, r, auth.FIRMWARE_ENTITY)
	_, err = firmware.GetFirmwareRuleOneDB(firmwareRule.ID)
	if err == nil {
		err = updateFirmwareRule(*firmwareRule, appType, true)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		result, _ := firmware.GetFirmwareRuleOneDB(firmwareRule.ID)
		response, err := xhttp.ReturnJsonResponse(result, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
	} else {
		response := "firmwareRule does not exist for '" + firmwareRule.ID + "'"
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, response)
	}
}

// Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
// 104 DELETE /xconfadminService/ux/api/firmwarerule/9d531fc4-089d-4c05-8fea-d9f32786ef51
func DeleteFirmwareRuleByIdHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
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

	entityOnDb, err := firmware.GetFirmwareRuleOneDB(id)
	if err == nil {
		if entityOnDb.ApplicationType != appType {
			errorStr := fmt.Sprintf("ApplicationType mismatch: %v on db. %v provided", entityOnDb.ApplicationType, appType)
			xhttp.WriteAdminErrorResponse(w, http.StatusConflict, errorStr)
			return
		}
		err = db.GetCachedSimpleDao().DeleteOne(db.TABLE_FIRMWARE_RULE, id)
	}
	if err != nil {
		response := "firmwareRule does not exist for " + id
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, response)
		return
	}

	xwhttp.WriteXconfResponse(w, http.StatusNoContent, []byte(""))
}

// Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
// 3 GET /xconfadminService/ux/api/firmwarerule/MAC_RULE/names
func GetFirmwareRuleByTypeNamesHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	givenType, found := mux.Vars(r)[xcommon.TYPE]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.TYPE)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	nameMap := make(map[string]string)
	dbrules, _ := firmware.GetFirmwareRuleAllAsListDBForAdmin()
	for _, v := range dbrules {
		if v.Type == givenType && appType == v.ApplicationType {
			nameMap[v.ID] = v.Name
		}
	}

	res, err := xhttp.ReturnJsonResponse(nameMap, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

// Zero Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
func GetFirmwareRuleByTemplateNamesHandler(w http.ResponseWriter, r *http.Request) {
	// Zero usage - No urgency to migrate
	xhttp.WriteAdminErrorResponse(w, http.StatusNotImplemented, "")
}

// Zero Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
func GetFirmwareRuleByTemplateByTemplateIdNamesHandler(w http.ResponseWriter, r *http.Request) {
	muxVarTemplateId := "templateId"
	templateId, found := mux.Vars(r)[muxVarTemplateId]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", muxVarTemplateId)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	dbrules, err := firmware.GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	filterContext, err := populateContext(w, r, true)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	filterContext[cFirmwareRuleTemplateId] = templateId
	dbrules = filterFirmwareRulesByContext(dbrules, filterContext)

	namesList := []string{}
	for _, rule := range dbrules {
		namesList = append(namesList, rule.Name)
	}

	response, err := xhttp.ReturnJsonResponse(namesList, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

// Zero Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
// /firmwarerule/export/byType?exportAll&type=applicableActionType
func GetFirmwareRuleExportByTypeHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	_, ok := queryParams[xcommon.EXPORTALL]
	if ok {
		actionTypes, ok2 := queryParams[xcommon.TYPE]
		actionType := ""
		if ok2 {
			actionType = actionTypes[0]
		}
		if util.IsBlank(actionType) {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Missing type param")
			return
		}
		context, err := populateContext(w, r, true)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		appType := context[common.APPLICATION_TYPE]

		frs, _ := firmware.GetFirmwareRuleAllAsListByApplicationTypeForAS(appType)
		dbrules := []*firmware.FirmwareRule{}

		for _, rules := range frs {
			rules = firmwareRuleFilterByActionType(rules, actionType)
			dbrules = append(dbrules, rules...)
		}

		res, err := xhttp.ReturnJsonResponse(dbrules, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		if actionType == "RULE" {
			actionType = actionType + "_ACTION"
		}
		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_FIRMWARE_RULES + "_" + actionType + "_" + appType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
		return
	} else {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Missing exportAll param")
	}
}

// Zero Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
func GetFirmwareRuleExportAllTypesHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	if queryParams[xcommon.EXPORTALL] == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Missing exportAll param")
		return
	}
	dbrules, err := firmware.GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	filterContext, err := populateContext(w, r, true)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	dbrules = filterFirmwareRulesByContext(dbrules, filterContext)

	res, err := xhttp.ReturnJsonResponse(dbrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	appType := filterContext[common.APPLICATION_TYPE]
	headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_FIRMWARE_RULES + "_" + appType)
	xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
}

func duplicateFrFound(entity *firmware.FirmwareRule, nameMap map[string][]*firmware.FirmwareRule, ruleMap map[string][]*firmware.FirmwareRule, estbMap map[string][]*firmware.FirmwareRule) error {
	items, found := nameMap[entity.Name]
	if found {
		for _, item := range items {
			if item.ID == entity.ID || item.ApplicationType != entity.ApplicationType {
				continue
			}
			return errors.New(item.ID + "/" + entity.Name + ": is already used")
		}
	}

	mapKey, estb := convertToMapKey(entity)
	if estb != "" {
		items, found = estbMap[estb]
		if found {
			for _, item := range items {
				if item.ID == entity.ID || item.ApplicationType != entity.ApplicationType {
					continue
				}
				if re.EqualComplexRules(item.GetRule(), entity.GetRule()) {
					return errors.New("Estb Rule " + entity.Name + " is duplicate of rule for " + item.Name)
				}
			}
			return nil
		}
	}

	items, found = ruleMap[mapKey]
	if found {
		for _, item := range items {
			if item.ID == entity.ID || item.ApplicationType != entity.ApplicationType {
				continue
			}
			if re.EqualComplexRules(item.GetRule(), entity.GetRule()) {
				return errors.New("Rule " + entity.Name + " is duplicate of rule for " + item.Name)
			}
		}
	}
	return nil
}

func findAndDeleteFR(list []*firmware.FirmwareRule, item firmware.FirmwareRule) []*firmware.FirmwareRule {
	index := 0
	for _, iter := range list {
		if iter.ID != item.ID {
			list[index] = iter
			index++
		}
	}
	return list[:index]
}

func convertToMapKey(rule *firmware.FirmwareRule) (string, string) {
	keys := []string{}

	flattenedRule := re.FlattenRule(*rule.GetRule())
	for _, elem := range flattenedRule {
		entry := strings.ToLower(elem.Condition.FreeArg.Name)
		if searchList(keys, entry, false) {
			continue
		}
		keys = append(keys, entry)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(strings.ToLower(keys[i]), strings.ToLower(keys[j])) < 0
	})
	result := ""
	for _, key := range keys {
		result = result + key + "&"
	}
	if result == "estbmac&" && len(flattenedRule) == 1 && flattenedRule[0].Relation == re.StandardOperationIs {
		newkey := flattenedRule[0].Condition.GetFixedArg().String()
		newkey = strings.Replace(newkey, "'", "", -1)
		return result, newkey
	}
	return result, ""
}

// Zero Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
func PostFirmwareRuleEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	PostPutFirmwareRuleEntitiesHandler(w, r, false)
}

// Zero Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
func PutFirmwareRuleEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	PostPutFirmwareRuleEntitiesHandler(w, r, true)
}

func PostPutFirmwareRuleEntitiesHandler(w http.ResponseWriter, r *http.Request, isPut bool) {
	appType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	body := xw.Body()
	const LargeValue = 1024
	var entities = make([]*firmware.FirmwareRule, LargeValue)
	for i := range entities {
		entities[i] = firmware.NewEmptyFirmwareRule()
		entities[i].ApplicationType = appType
	}
	err = json.Unmarshal([]byte(body), &entities)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	nameMap := make(map[string][]*firmware.FirmwareRule)
	ruleMap := make(map[string][]*firmware.FirmwareRule)
	estbMap := make(map[string][]*firmware.FirmwareRule)

	list, err := firmware.GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i, item := range list {
		nameMap[item.Name] = append(nameMap[item.Name], list[i])
		mapKey, estb := convertToMapKey(item)
		if estb != "" {
			estbMap[estb] = append(estbMap[estb], list[i])
		} else {
			ruleMap[mapKey] = append(ruleMap[mapKey], list[i])
		}
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for i, entity := range entities {
		_, err := firmware.GetFirmwareRuleOneDB(entity.ID)
		if isPut && err != nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: "FirmwareRule with " + entity.ID + " not present",
			}
			continue
		}
		if !isPut && err == nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: "FirmwareRule with " + entity.ID + " already present",
			}
			continue
		}

		if entity.ApplicationType != appType {
			if util.IsBlank(entity.ID) {
				entity.ID = uuid.New().String() + entity.Name
			}
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: "ApplicationType conflict. User specified: " + appType + " , Instance applicationType " + entity.ApplicationType,
			}
			continue
		}
		if err = duplicateFrFound(entity, nameMap, ruleMap, estbMap); err != nil {
			if util.IsBlank(entity.ID) {
				entity.ID = uuid.New().String() + entity.Name
			}
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
			continue
		}

		if isPut {
			err = updateFirmwareRule(*entity, appType, false)
		} else {
			entity.Active = true
			if entity.ApplicableAction != nil {
				entity.ApplicableAction.Active = true
			}
			err = createFirmwareRule(*entity, appType, false)
		}
		if err != nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
			continue
		} else {
			nameMap[entity.Name] = findAndDeleteFR(nameMap[entity.Name], *entity)
			nameMap[entity.Name] = append(nameMap[entity.Name], entities[i])

			mapKey, estb := convertToMapKey(entity)
			if estb != "" {
				estbMap[estb] = findAndDeleteFR(estbMap[estb], *entity)
				estbMap[estb] = append(estbMap[estb], entities[i])
			} else {
				findAndDeleteFR(ruleMap[mapKey], *entity)
				ruleMap[mapKey] = append(ruleMap[mapKey], entities[i])
			}

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

// Zero Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
func ObsoleteGetFirmwareRulePageHandler(w http.ResponseWriter, r *http.Request) {
	// Get all sorted rules
	dbrules, err := firmware.GetFirmwareSortedRuleAllAsListDB()
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Populate the headers
	headers := putSizesOfFirmwareRulesByTypeIntoHeaders(dbrules)

	// Get the entries from the requested page
	filterContext, err := populateContext(w, r, true)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	dbrules, err = generateFirmwareRulePageByContext(dbrules, filterContext)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// return the response
	response, err := xhttp.ReturnJsonResponse(dbrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, response)
}

// Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
// 1  GET /xconfadminService/ux/api/firmwarerule
func GetFirmwareRuleHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	filtRules := []*firmware.FirmwareRule{}
	dbrules, _ := xfirmware.GetFirmwareSortedRuleAllAsListDB()

	for _, rule := range dbrules {
		if appType == rule.ApplicationType {
			filtRules = append(filtRules, rule)
		}
	}
	dbrules = filtRules
	res, err := xhttp.ReturnJsonResponse(dbrules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	queryParams := r.URL.Query()
	_, ok := queryParams[xcommon.EXPORT]
	if ok {
		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_FIRMWARE_RULES)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
		return
	}
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

// Usage pattern from green splunk for 4 weeks ending 23rd Oct 2021
// 1247 GET /xconfadminService/ux/api/firmwarerule/{id}
// 84 GET /xconfadminService/ux/api/firmwarerule/{id}?export
func GetFirmwareRuleByIdHandler(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
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

	fr, _ := firmware.GetFirmwareRuleOneDB(id)
	if fr == nil {
		errorStr := fmt.Sprintf("%v not found", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	if fr.ApplicationType != appType {
		errorStr := fmt.Sprintf("ApplicationType mismatch: %v on db. %v provided", fr.ApplicationType, appType)
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, errorStr)
		return
	}
	queryParams := r.URL.Query()
	_, ok := queryParams[xcommon.EXPORT]
	if ok {
		frList := []firmware.FirmwareRule{*fr}
		res, err := xhttp.ReturnJsonResponse(frList, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}

		fileName := xcommon.ExportFileNames_FIRMWARE_RULE + fr.ID + "_" + appType
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
		return
	}
	res, err := xhttp.ReturnJsonResponse(fr, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}
