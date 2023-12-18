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
package change

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	xchange "xconfadmin/shared/change"

	xutil "xconfadmin/util"

	xcommon "xconfadmin/common"
	xwcommon "xconfwebconfig/common"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func GetTwoProfileChangesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	searchContext := make(map[string]string)
	searchContext[xwcommon.APPLICATION_TYPE] = applicationType

	changes := GetTelemetryTwoChangesByContext(searchContext)
	sort.Slice(changes, func(i, j int) bool {
		return changes[j].Updated < changes[i].Updated
	})

	res, err := xhttp.ReturnJsonResponse(changes, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func GetApprovedTwoChangesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	changes := xchange.GetApprovedTelemetryTwoChangesByApplicationType(applicationType)
	res, err := xhttp.ReturnJsonResponse(changes, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func GetTwoChangeEntityIdsHandler(w http.ResponseWriter, r *http.Request) {
	entityIds := GetTelemetryTwoChangeEntityIds()
	res, err := xhttp.ReturnJsonResponse(entityIds, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func ApproveTwoChangeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	changeId, found := mux.Vars(r)[xcommon.CHANGE_ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.CHANGE_ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	approvedChange, err := ApproveTelemetryTwoChange(r, changeId)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(approvedChange, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func ApproveTwoChangesHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	idList := []string{}
	err = json.Unmarshal([]byte(body), &idList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	errorMessages := ApproveTelemetryTwoChanges(r, idList)

	res, err := xhttp.ReturnJsonResponse(errorMessages, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func RevertTwoChangeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	approveId, found := mux.Vars(r)[xcommon.APPROVE_ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.APPROVE_ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	respEntity := RevertTelemetryTwoChange(r, approveId)
	if respEntity.Error != nil {
		xwhttp.WriteXconfResponse(w, respEntity.Status, []byte(respEntity.Error.Error()))
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, respEntity.Status, xhttp.ContextTypeHeader(r))
}

func RevertTwoChangesHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	idList := []string{}
	err = json.Unmarshal([]byte(body), &idList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	errorMessages := RevertTelemetryTwoChanges(r, idList)

	res, err := xhttp.ReturnJsonResponse(errorMessages, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func CancelTwoChangeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	changeId, found := mux.Vars(r)[xcommon.CHANGE_ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.CHANGE_ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	if err := DeleteTelemetryTwoChange(changeId); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	userName := auth.GetUserNameOrUnknown(r)
	log.Info(fmt.Sprintf("Change has been canceled by %s: %s", userName, changeId))

	xwhttp.WriteResponseBytes(w, []byte{}, http.StatusOK, xhttp.ContextTypeHeader(r))
}

func GetGroupedTwoChangesHandler(w http.ResponseWriter, r *http.Request) {
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

	changes := xchange.GetAllTelemetryTwoChangeList()
	changesPerPage := GeneratePageTelemetryTwoChanges(changes, pageNumber, pageSize)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	groupedChanges := GroupTelemetryTwoChanges(changesPerPage)
	res, err := xhttp.ReturnJsonResponse(groupedChanges, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(changes))
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, []byte(res))
}

func GetGroupedApprovedTwoChangesHandler(w http.ResponseWriter, r *http.Request) {
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

	changes := xchange.GetAllApprovedTelemetryTwoChangeList()
	changesPerPage := GeneratePageApprovedTelemetryTwoChanges(changes, pageNumber, pageSize)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	groupedChanges := GroupApprovedTelemetryTwoChanges(changesPerPage)
	res, err := xhttp.ReturnJsonResponse(groupedChanges, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(changes))
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, []byte(res))
}

func GetApprovedTwoChangesFilteredHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.CHANGE_ENTITY)
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
		xhttp.AdminError(w, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
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

	approvedChanges := GetApprovedTelemetryTwoChangesByContext(contextMap)
	approvedChangesPerPage := GeneratePageApprovedTelemetryTwoChanges(approvedChanges, pageNumber, pageSize)
	changes := GetTelemetryTwoChangesByContext(contextMap)

	res, err := xhttp.ReturnJsonResponse(approvedChangesPerPage, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	sizeHeader := createHeadersWithEntitySize(len(changes), len(approvedChanges))
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, []byte(res))
}

func GetTwoChangesFilteredHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.CHANGE_ENTITY)
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
		xhttp.AdminError(w, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()

	contextMap := map[string]string{}
	if body != "" {
		if err := json.Unmarshal([]byte(body), &contextMap); err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		xutil.AddQueryParamsToContextMap(r, contextMap)
	}
	contextMap[xwcommon.APPLICATION_TYPE] = applicationType

	changes := GetTelemetryTwoChangesByContext(contextMap)
	changesPerPage := GeneratePageTelemetryTwoChanges(changes, pageNumber, pageSize)
	approvedChanges := GetApprovedTelemetryTwoChangesByContext(contextMap)

	res, err := xhttp.ReturnJsonResponse(changesPerPage, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	sizeHeader := createHeadersWithEntitySize(len(changes), len(approvedChanges))
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, []byte(res))
}
