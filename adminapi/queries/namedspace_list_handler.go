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
package queries

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"xconfwebconfig/shared"

	"github.com/gorilla/mux"

	xcommon "xconfadmin/common"
	covt "xconfadmin/shared/estbfirmware"
	"xconfadmin/util"
	xutil "xconfwebconfig/util"

	xwcommon "xconfwebconfig/common"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xshared "xconfadmin/shared"
	xwhttp "xconfwebconfig/http"
)

func GetQueriesIpAddressGroups(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result := GetIpAddressGroups()
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesIpAddressGroupsByIp(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	ipAddress, found := mux.Vars(r)[xwcommon.IP_ADDRESS]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.IP_ADDRESS)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	if net.ParseIP(ipAddress) == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "IpAddress is invalid.")
		return
	}

	result := GetIpAddressGroupsByIp(ipAddress)
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesIpAddressGroupsByName(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name, found := mux.Vars(r)[xwcommon.NAME]
	if !found || len(strings.TrimSpace(name)) == 0 {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.NAME)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	result := []*shared.IpAddressGroup{}

	ipAddrGrp := GetIpAddressGroupByName(name)
	if ipAddrGrp == nil {
		values, ok := r.URL.Query()[xwcommon.VERSION]
		if ok {
			apiVersion := values[0]
			if xutil.IsVersionGreaterOrEqual(apiVersion, 3.0) {
				errorStr := fmt.Sprintf("IpAddressGroup with name %s does not exist", name)
				xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
				return
			}
		}
	} else {
		result = append(result, ipAddrGrp)
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func CreateIpAddressGroupHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
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
	newIpAddressGroup := shared.IpAddressGroup{}
	err := json.Unmarshal([]byte(body), &newIpAddressGroup)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := CreateIpAddressGroup(&newIpAddressGroup)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func AddDataIpAddressGroupHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	listId, found := mux.Vars(r)[xwcommon.LIST_ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.LIST_ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	stringListWrapper := shared.StringListWrapper{}
	err := json.Unmarshal([]byte(body), &stringListWrapper)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := AddNamespacedListData(shared.IP_LIST, listId, &stringListWrapper)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func RemoveDataIpAddressGroupHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	listId, found := mux.Vars(r)[xwcommon.LIST_ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.LIST_ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	stringListWrapper := shared.StringListWrapper{}
	err := json.Unmarshal([]byte(body), &stringListWrapper)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := RemoveNamespacedListData(shared.IP_LIST, listId, &stringListWrapper)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
}

func DeleteIpAddressGroupHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	respEntity := DeleteNamespacedList(shared.IP_LIST, id)
	if respEntity.Error != nil {
		if respEntity.Status == http.StatusNotFound {
			respEntity.Status = http.StatusNoContent // Ignored not found
		} else {
			xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
			return
		}
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func GetQueriesIpAddressGroupsV2(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result := GetNamespacedListsByType(shared.IP_LIST)
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesIpAddressGroupsByIpV2(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	ipAddress, found := mux.Vars(r)[xwcommon.IP_ADDRESS]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.IP_ADDRESS)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	if net.ParseIP(ipAddress) == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "IpAddress is invalid.")
		return
	}

	result := GetNamespacedListsByIp(ipAddress)
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesIpAddressGroupsByNameV2(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	ipAddrGrp := GetNamespacedListByIdAndType(id, shared.IP_LIST)
	if ipAddrGrp == nil {
		errorStr := fmt.Sprintf("IpAddressGroup with name %s does not exist", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	res, err := xhttp.ReturnJsonResponse(ipAddrGrp, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func CreateIpAddressGroupHandlerV2(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
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

	newIpList := shared.NewIpList()
	err := json.Unmarshal([]byte(body), &newIpList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := CreateNamespacedList(newIpList, false)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func UpdateIpAddressGroupHandlerV2(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
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

	newIpList := shared.NewIpList()
	err := json.Unmarshal([]byte(body), &newIpList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdateNamespacedList(newIpList, "")
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func DeleteIpAddressGroupHandlerV2(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	respEntity := DeleteNamespacedList(shared.IP_LIST, id)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, []byte(fmt.Sprintf("Successfully deleted %s", id)))
}

func GetQueriesMacLists(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result := GetNamespacedListsByType(shared.MAC_LIST)
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesMacListsById(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	macList := GetNamespacedListByIdAndType(id, shared.MAC_LIST)
	if macList == nil {
		values, ok := r.URL.Query()[xwcommon.VERSION]
		if ok {
			apiVersion := values[0]
			if xutil.IsVersionGreaterOrEqual(apiVersion, 3.0) {
				errorStr := fmt.Sprintf("MacList with id %s does not exist", id)
				xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
				return
			}
		} else {
			xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(""))
			return
		}
	}
	res, err := xhttp.ReturnJsonResponse(macList, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesMacListsByMacPart(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	mac := mux.Vars(r)[xwcommon.MAC]
	result := GetMacListsByMacPart(mac)
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func SaveMacListHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
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
	newMacList := shared.NewMacList()
	err := json.Unmarshal([]byte(body), &newMacList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create the new MacList or update an existing one
	respEntity := CreateNamespacedList(newMacList, true)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func CreateMacListHandlerV2(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
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
	newMacList := shared.NewMacList()
	err := json.Unmarshal([]byte(body), &newMacList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := CreateNamespacedList(newMacList, false)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func UpdateMacListHandlerV2(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
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
	newMacList := shared.NewMacList()
	err := json.Unmarshal([]byte(body), &newMacList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdateNamespacedList(newMacList, "")
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func AddDataMacListHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	listId, found := mux.Vars(r)[xwcommon.LIST_ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.LIST_ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	stringListWrapper := shared.StringListWrapper{}
	err := json.Unmarshal([]byte(body), &stringListWrapper)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := AddNamespacedListData(shared.MAC_LIST, listId, &stringListWrapper)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func RemoveDataMacListHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	listId, found := mux.Vars(r)[xwcommon.LIST_ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.LIST_ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	stringListWrapper := shared.StringListWrapper{}
	err := json.Unmarshal([]byte(body), &stringListWrapper)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := RemoveNamespacedListData(shared.MAC_LIST, listId, &stringListWrapper)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func DeleteMacListHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	respEntity := DeleteNamespacedList(shared.MAC_LIST, id)
	if respEntity.Error != nil {
		if respEntity.Status == http.StatusNotFound {
			respEntity.Status = http.StatusNoContent // Ignored not found
		} else {
			xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
			return
		}
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func GetQueriesMacListsByIdV2(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	macList := GetNamespacedListByIdAndType(id, shared.MAC_LIST)
	if macList == nil {
		errorStr := fmt.Sprintf("MacList with id %s does not exist", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	res, err := xhttp.ReturnJsonResponse(macList, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func DeleteMacListHandlerV2(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	respEntity := DeleteNamespacedList(shared.MAC_LIST, id)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, []byte(fmt.Sprintf("Successfully deleted %s", id)))
}

func GetNamespacedListHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	nsList, err := xshared.GetGenericNamedListOneNonCached(id)
	if err != nil {
		errorStr := fmt.Sprintf("List with id %s does not exist", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	if _, ok := r.URL.Query()[xcommon.EXPORT]; ok {
		res, err := xhttp.ReturnJsonResponse([]*shared.GenericNamespacedList{nsList}, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		fileName := xcommon.ExportFileNames_NAMESPACEDLIST + id
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		res, err := xhttp.ReturnJsonResponse(nsList, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
	}
}

func GetNamespacedListIdsHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	ids := GetNamespacedListIdsByType("")

	res, err := xhttp.ReturnJsonResponse(ids, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetNamespacedListIdsByTypeHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	typeName, found := mux.Vars(r)[xcommon.TYPE]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.TYPE)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	ids := GetNamespacedListIdsByType(typeName)
	sortedById := func(i, j int) bool {
		return strings.ToLower(ids[i]) < strings.ToLower(ids[j])
	}
	sort.Slice(ids, sortedById)

	res, err := xhttp.ReturnJsonResponse(ids, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetNamespacedListsHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result := GetNamespacedListsByType("")
	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(strings.ToLower(result[i].ID), strings.ToLower(result[j].ID)) < 0
	})
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	if _, ok := r.URL.Query()[xcommon.EXPORT]; ok {
		fileName := xcommon.ExportFileNames_ALL_NAMESPACEDLISTS
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
	}
}

func GetNamespacedListsByTypeHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	typeName, found := mux.Vars(r)[xcommon.TYPE]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.TYPE)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	result := GetNamespacedListsByType(typeName)
	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(strings.ToLower(result[i].ID), strings.ToLower(result[j].ID)) < 0
	})
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	if _, ok := r.URL.Query()[xcommon.EXPORT]; ok {
		fileName := xcommon.ExportFileNames_ALL + typeName + "S"
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
	}
}

func GetIpAddressGroupsHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	nsLists := GetNamespacedListsByType(shared.IP_LIST)
	result := covt.ConvertToListOfIpAddressGroups(nsLists)
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func CreateNamespacedListHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
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

	newNamespacedListList := shared.NewEmptyGenericNamespacedList()
	err := json.Unmarshal([]byte(body), &newNamespacedListList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := CreateNamespacedList(newNamespacedListList, false)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func UpdateNamespacedListHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
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

	namespacedListList := shared.NewEmptyGenericNamespacedList()
	err := json.Unmarshal([]byte(body), &namespacedListList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdateNamespacedList(namespacedListList, "")
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func RenameNamespacedListHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()

	namespacedListList := shared.NewEmptyGenericNamespacedList()
	err := json.Unmarshal([]byte(body), &namespacedListList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdateNamespacedList(namespacedListList, id)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, res)
}

func DeleteNamespacedListHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xwcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xwcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	respEntity := DeleteNamespacedList("", id)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func PostNamespacedListFilteredHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	queryParams := map[string]string{}
	util.AddQueryParamsToContextMap(r, queryParams)
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
	util.AddQueryParamsToContextMap(r, contextMap)

	nsLists := GetNamespacedListsByContext(contextMap)
	sort.Slice(nsLists, func(i, j int) bool {
		return strings.Compare(strings.ToLower(nsLists[i].ID), strings.ToLower(nsLists[j].ID)) < 0
	})
	nsListsPerPage := GeneratePageNamespacedLists(nsLists, pageNumber, pageSize)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(nsListsPerPage, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	sizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(nsLists))
	xwhttp.WriteXconfResponseWithHeaders(w, sizeHeader, http.StatusOK, []byte(res))
}

func PostNamespacedListEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	entities := []shared.GenericNamespacedList{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract entity from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		respEntity := CreateNamespacedList(&entity, false)
		if respEntity.Error == nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: entity.ID,
			}
		} else {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: respEntity.Error.Error(),
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

func PutNamespacedListEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	entities := []shared.GenericNamespacedList{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract entity from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}

	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		respEntity := UpdateNamespacedList(&entity, "")
		if respEntity.Error == nil {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: entity.ID,
			}
		} else {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: respEntity.Error.Error(),
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
