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
	"strings"

	"github.com/gorilla/mux"

	xutil "xconfadmin/util"

	xcommon "xconfadmin/common"
	"xconfwebconfig/common"
	"xconfwebconfig/shared"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"
)

func PostModelEntitiesHandler(w http.ResponseWriter, r *http.Request) {
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
	entities := []shared.Model{}
	err := json.Unmarshal([]byte(body), &entities)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		respEntity := CreateModel(&entity)
		if respEntity.Status != http.StatusCreated {
			entitiesMap[entity.ID] = xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: respEntity.Error.Error(),
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

func PutModelEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	entities := []shared.Model{}
	if err := json.Unmarshal([]byte(xw.Body()), &entities); err != nil {
		response := "Unable to extract entity from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}
	entitiesMap := map[string]xhttp.EntityMessage{}
	for _, entity := range entities {
		entity := entity
		respEntity := UpdateModel(&entity)
		if respEntity.Status == http.StatusOK {
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

func ObsoleteGetModelPageHandler(w http.ResponseWriter, r *http.Request) {
	entries := shared.GetAllModelList()
	sort.Slice(entries, func(i, j int) bool {
		return strings.Compare(strings.ToLower(entries[i].ID), strings.ToLower(entries[j].ID)) < 0
	})

	contextMap := map[string]string{}
	xutil.AddQueryParamsToContextMap(r, contextMap)

	var err error
	entries, err = generateModelPageByContext(entries, contextMap)
	allItemsLen := len(entries)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := xhttp.ReturnJsonResponse(entries, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	headerMap := populateHeaderWithNumberOfItems(allItemsLen)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func PostModelFilteredHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// Figure out the pageContext from Query
	pageContext := map[string]string{}
	xutil.AddQueryParamsToContextMap(r, pageContext)

	// Figure out the filterContext from Body
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Unable to extract Body")
		return
	}
	filterContext := make(map[string]string)
	body := xw.Body()
	var err error
	if body != "" {
		err = json.Unmarshal([]byte(body), &filterContext)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Get all entries and sort them
	entries := shared.GetAllModelList()
	sort.Slice(entries, func(i, j int) bool {
		return strings.Compare(strings.ToLower(entries[i].ID), strings.ToLower(entries[j].ID)) < 0
	})

	// Filter entries according to filterContext
	entries, err = filterModelsByContext(entries, filterContext)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	allItemsLen := len(entries)

	// Get the entries from the requested  page as per pageContext
	entries, err = generateModelPageByContext(entries, pageContext)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// return the response
	response, err := xhttp.ReturnJsonResponse(entries, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	headerMap := populateHeaderWithNumberOfItems(allItemsLen)
	xwhttp.WriteXconfResponseWithHeaders(w, headerMap, http.StatusOK, response)
}

func GetModelByIdHandler(w http.ResponseWriter, r *http.Request) {
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
	id = strings.ToUpper(id)
	model := shared.GetOneModel(id)
	if model == nil {
		errorStr := fmt.Sprintf("%v not found", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	queryParams := r.URL.Query()
	_, ok := queryParams[xcommon.EXPORT]
	if ok {
		modelList := []shared.Model{*model}
		res, err := xhttp.ReturnJsonResponse(modelList, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}

		fileName := xcommon.ExportFileNames_MODEL + model.ID
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
		return
	}

	res, err := xhttp.ReturnJsonResponse(model, r)
	if err != nil {
		xhttp.AdminError(w, err)
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetModelHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	models := shared.GetAllModelList()
	sort.Slice(models, func(i, j int) bool {
		return strings.Compare(strings.ToLower(models[i].ID), strings.ToLower(models[j].ID)) < 0
	})
	res, err := xhttp.ReturnJsonResponse(models, r)
	if err != nil {
		xhttp.AdminError(w, err)
	}

	queryParams := r.URL.Query()
	_, ok := queryParams[xcommon.EXPORT]
	if ok {
		headers := xhttp.CreateContentDispositionHeader(xcommon.ExportFileNames_ALL_MODELS)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
		return
	}

	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}
