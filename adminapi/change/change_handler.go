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

	xcommon "xconfadmin/common"
	xshared "xconfadmin/shared"
	xchange "xconfadmin/shared/change"
	xwcommon "xconfwebconfig/common"
	"xconfwebconfig/shared"
	xwchange "xconfwebconfig/shared/change"
	"xconfwebconfig/util"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	PENDING_CHANGE_SIZE  = "pendingChangesSize"
	APPROVED_CHANGE_SIZE = "approvedChangesSize"
)

func GetProfileChangesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	searchContext := make(map[string]string)
	searchContext[xwcommon.APPLICATION_TYPE] = applicationType
	changes := FindByContextForChanges(searchContext)
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

func ApproveChangeHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	changeId, found := mux.Vars(r)[xcommon.CHANGE_ID]
	if !found || changeId == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", xcommon.CHANGE_ID))
		return
	}

	_, err = Approve(r, changeId)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	headerMap := createHeadersMap(applicationType)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, nil)
}

func GetApprovedHandler(w http.ResponseWriter, r *http.Request) {
	approvedChange, err := GetApprovedAll(r)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	response, err := util.JSONMarshal(approvedChange)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal approvedChange error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func RevertChangeHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	approveId, found := mux.Vars(r)[xcommon.APPROVE_ID]
	if !found || approveId == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", xcommon.APPROVE_ID))
		return
	}

	err = Revert(r, approveId)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	headerMap := createHeadersMap(applicationType)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, nil)
}

func CancelChangeHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	changeId, found := mux.Vars(r)[xcommon.CHANGE_ID]
	if !found || changeId == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v is invalid", xcommon.CHANGE_ID))
		return
	}

	err = CancelChange(r, changeId)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	headerMap := createHeadersMap(applicationType)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, nil)
}

func createHeadersMap(applicationType string) map[string]string {
	headerMap := make(map[string]string, 2)
	changeListAll := xchange.GetChangeList()
	approvedChangeListAll := xchange.GetApprovedChangeList()
	var lenChangeList int = len(changeListAll)
	var lenApprovedChangeList int = len(approvedChangeListAll)
	var changeList = []*xwchange.Change{}
	var approvedChangeList = []*xwchange.ApprovedChange{}
	for _, change := range changeListAll {
		if xshared.ApplicationTypeEquals(applicationType, change.ApplicationType) || xshared.ApplicationTypeEquals(applicationType, shared.ALL) {
			changeList = append(changeList, change)
		}
	}
	for _, approvedChange := range approvedChangeListAll {
		if xshared.ApplicationTypeEquals(applicationType, approvedChange.ApplicationType) || xshared.ApplicationTypeEquals(applicationType, shared.ALL) {
			approvedChangeList = append(approvedChangeList, approvedChange)
		}
	}
	if changeList == nil {
		headerMap[PENDING_CHANGE_SIZE] = "0"
	} else {
		headerMap[PENDING_CHANGE_SIZE] = strconv.Itoa(lenChangeList)
	}
	if approvedChangeList == nil {
		headerMap[APPROVED_CHANGE_SIZE] = "0"
	} else {
		headerMap[APPROVED_CHANGE_SIZE] = strconv.Itoa(lenApprovedChangeList)
	}
	return headerMap
}

func GetGroupedChangesHandler(w http.ResponseWriter, r *http.Request) {
	var pageNumberStr, pageSizeStr string
	var pageNumber, pageSize int
	var err error
	if values, ok := r.URL.Query()[xcommon.PAGE_NUMBER]; ok {
		pageNumberStr = values[0]
		pageNumber, err = strconv.Atoi(pageNumberStr)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageNumber")
			return
		}
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Required parameter 'pageNumber' is not present"))
		return
	}
	if values, ok := r.URL.Query()[xcommon.PAGE_SIZE]; ok {
		pageSizeStr = values[0]
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageSize")
			return
		}
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Required parameter 'pageSize' is not present"))
		return
	}
	applicationType, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	changeList := xchange.GetChangeList()
	sort.Slice(changeList, func(i, j int) bool {
		return changeList[i].Updated < changeList[j].Updated
	})
	changesPerPage := ChangesGeneratePage(changeList, pageNumber, pageSize)
	changeMap := GroupChanges(changesPerPage)
	response, err := util.JSONMarshal(changeMap)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal changeMap error: %v", err))
	}
	headerMap := createHeadersMap(applicationType)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func GetGroupedApprovedChangesHandler(w http.ResponseWriter, r *http.Request) {
	var pageNumberStr, pageSizeStr string
	var pageNumber, pageSize int
	var err error
	if values, ok := r.URL.Query()[xcommon.PAGE_NUMBER]; ok {
		pageNumberStr = values[0]
		pageNumber, err = strconv.Atoi(pageNumberStr)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageNumber")
			return
		}
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Required parameter 'pageNumber' is not present"))
		return
	}
	if values, ok := r.URL.Query()[xcommon.PAGE_SIZE]; ok {
		pageSizeStr = values[0]
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageSize")
			return
		}
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Required parameter 'pageSize' is not present"))
		return
	}
	applicationType, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	changeList := xchange.GetApprovedChangeList()
	sort.Slice(changeList, func(i, j int) bool {
		return changeList[j].Updated < changeList[i].Updated
	})
	changesPerPage := ApprovedChangesGeneratePage(changeList, pageNumber, pageSize)
	changeMap := GroupApprovedChanges(changesPerPage)
	ApprovedChangesMap := make(map[string]map[string][]*xwchange.ApprovedChange, 1)
	ApprovedChangesMap["changesPerPage"] = changeMap
	response, err := util.JSONMarshal(ApprovedChangesMap)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal ApprovedChangesMap error: %v", err))
	}
	headerMap := createHeadersMap(applicationType)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func ChangesGeneratePage(list []*xwchange.Change, page int, pageSize int) (result []*xwchange.Change) {
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

func ApprovedChangesGeneratePage(list []*xwchange.ApprovedChange, page int, pageSize int) (result []*xwchange.ApprovedChange) {
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

func GetChangedEntityIdsHandler(w http.ResponseWriter, r *http.Request) {
	entityIds := GetChangedEntityIds()
	response, err := util.JSONMarshal(entityIds)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal entityIds error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func ApproveChangesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	var changeIds []string
	if err := json.Unmarshal([]byte(xw.Body()), &changeIds); err != nil {
		response := "Unable to extract changeIds from json file:" + err.Error()
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
		return
	}

	errorMessages, err := ApproveChanges(r, &changeIds)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(errorMessages)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal ApprovedChangesMap error: %v", err))
	}
	headerMap := createHeadersMap(applicationType)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func RevertChangesHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	var changeIds []string
	if err := json.Unmarshal([]byte(xw.Body()), &changeIds); err != nil {
		response := "Unable to extract changeIds from json file:" + err.Error()
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
		return
	}
	errorMessages, err := RevertChanges(r, &changeIds)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, err := util.JSONMarshal(errorMessages)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal ApprovedChangesMap error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetApprovedFilteredHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	var pageNumberStr, pageSizeStr string
	pageNumber := 1
	pageSize := 50
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	if values, ok := r.URL.Query()[xcommon.PAGE_NUMBER]; ok {
		pageNumberStr = values[0]
		pageNumber, err = strconv.Atoi(pageNumberStr)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageNumber")
			return
		}
	}
	if values, ok := r.URL.Query()[xcommon.PAGE_SIZE]; ok {
		pageSizeStr = values[0]
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageSize")
			return
		}
	}

	searchContext := make(map[string]string)

	if xw.Body() != "" {
		if err := json.Unmarshal([]byte(xw.Body()), &searchContext); err != nil {
			response := "Unable to extract searchContext from json file:" + err.Error()
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
			return
		}
	}
	searchContext[xwcommon.APPLICATION_TYPE] = applicationType

	approvedChangeList := FindByContextForApprovedChanges(r, searchContext)
	sort.Slice(approvedChangeList, func(i, j int) bool {
		return approvedChangeList[j].Updated < approvedChangeList[i].Updated
	})
	changesPerPage := ApprovedChangesGeneratePage(approvedChangeList, pageNumber, pageSize)
	response, err := util.JSONMarshal(changesPerPage)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal ApprovedChangesMap error: %v", err))
	}
	changeList := FindByContextForChanges(searchContext)
	headerMap := createHeadersWithEntitySize(len(changeList), len(approvedChangeList))
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func GetChangesFilteredHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	var pageNumberStr, pageSizeStr string
	pageNumber := 1
	pageSize := 50
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	if values, ok := r.URL.Query()[xcommon.PAGE_NUMBER]; ok {
		pageNumberStr = values[0]
		pageNumber, err = strconv.Atoi(pageNumberStr)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageNumber")
			return
		}
	}
	if values, ok := r.URL.Query()[xcommon.PAGE_SIZE]; ok {
		pageSizeStr = values[0]
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Invalid value for pageSize")
			return
		}
	}
	searchContext := make(map[string]string)
	bodyStr := xw.Body()
	if bodyStr != "" {
		if err := json.Unmarshal([]byte(xw.Body()), &searchContext); err != nil {
			response := "Unable to extract searchContext from json file:" + err.Error()
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
			return
		}
	}
	searchContext[xwcommon.APPLICATION_TYPE] = applicationType

	changeList := FindByContextForChanges(searchContext)
	sort.Slice(changeList, func(i, j int) bool {
		return changeList[j].Updated < changeList[i].Updated
	})
	changesPerPage := ChangesGeneratePage(changeList, pageNumber, pageSize)
	response, err := util.JSONMarshal(changesPerPage)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal changeMap error: %v", err))
	}
	approvedChangeList := FindByContextForApprovedChanges(r, searchContext)
	headerMap := createHeadersWithEntitySize(len(changeList), len(approvedChangeList))
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func createHeadersWithEntitySize(pendingChangesSize int, approvedChangesSize int) map[string]string {
	headerMap := make(map[string]string, 2)
	headerMap[PENDING_CHANGE_SIZE] = strconv.Itoa(pendingChangesSize)
	headerMap[APPROVED_CHANGE_SIZE] = strconv.Itoa(approvedChangesSize)
	return headerMap
}
