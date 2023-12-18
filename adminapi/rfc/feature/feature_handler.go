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
package feature

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	xwutil "xconfadmin/util"

	xcommon "xconfadmin/common"
	xrfc "xconfadmin/shared/rfc"
	xwcommon "xconfwebconfig/common"
	xwrfc "xconfwebconfig/shared/rfc"
	"xconfwebconfig/util"

	"xconfadmin/adminapi/auth"

	xhttp "xconfadmin/http"

	xwhttp "xconfwebconfig/http"

	"github.com/gorilla/mux"
)

func GetFeaturesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	_, isExport := r.URL.Query()["export"]
	if isExport {
		featureEntityList := GetFeatureEntityListByApplicationTypeSorted(applicationType)
		filename := fmt.Sprintf("%s_%s", xcommon.ExportFileNames_ALL_FEATURES, applicationType)
		header := xhttp.CreateContentDispositionHeader(filename)
		response, _ := util.XConfJSONMarshal(featureEntityList, true)
		xwhttp.WriteXconfResponseWithHeaders(w, header, http.StatusOK, []byte(response))
	} else {
		features := GetFeaturesByApplicationTypeSorted(applicationType)
		response, _ := util.XConfJSONMarshal(features, true)
		xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
	}
}

func GetFeatureByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id := mux.Vars(r)[xwcommon.ID]
	if id == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is blank")
		return
	}
	_, isExport := r.URL.Query()["export"]
	if isExport {
		featureEntity := GetFeatureEntityById(id)
		if featureEntity == nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Entity with id: %s does not exist", id))
			return
		}
		if applicationType != featureEntity.ApplicationType {
			xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Entity with id: %s aplicationType does not match", id))
			return
		}
		filename := fmt.Sprintf("%s%s_%s", xcommon.ExportFileNames_FEATURE, featureEntity.ID, applicationType)
		header := xhttp.CreateContentDispositionHeader(filename)
		featureEntityList := []*xwrfc.FeatureEntity{featureEntity}
		response, _ := util.XConfJSONMarshal(featureEntityList, true)
		xwhttp.WriteXconfResponseWithHeaders(w, header, http.StatusOK, []byte(response))
	} else {
		feature := GetFeatureById(id)
		if feature == nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Entity with id: %s does not exist", id))
			return
		}
		if applicationType != feature.ApplicationType {
			xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Entity with id: %s aplicationType does not match", id))
			return
		}
		response, _ := util.XConfJSONMarshal(feature, true)
		xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
	}
}

func DeleteFeatureByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id := mux.Vars(r)[xwcommon.ID]
	if id == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is blank")
		return
	}
	if !xrfc.DoesFeatureExistWithApplicationType(id, applicationType) {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Entity with id: %s does not exist", id))
		return
	}
	isFeatureUsed, featureName := IsFeatureUsedInFeatureRule(id)
	if isFeatureUsed {
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, fmt.Sprintf("This Feature linked to FeatureRule with name: %s", featureName))
		return
	}
	DeleteFeatureById(id)
	xwhttp.WriteXconfResponse(w, http.StatusNoContent, []byte(""))
}

func PutFeatureEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	var featureEntityList []*xwrfc.FeatureEntity
	err = json.Unmarshal([]byte(body), &featureEntityList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	entitiesMap := ImportFeatureEntities(featureEntityList, true, applicationType)
	response, _ := util.XConfJSONMarshal(entitiesMap, true)
	xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
}

func PostFeatureEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	var featureEntityList []*xwrfc.FeatureEntity
	err = json.Unmarshal([]byte(body), &featureEntityList)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	entitiesMap := ImportFeatureEntities(featureEntityList, false, applicationType)
	response, _ := util.XConfJSONMarshal(entitiesMap, true)
	xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
}

func PostFeatureHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	var featureEntity *xwrfc.FeatureEntity
	err = json.Unmarshal([]byte(body), &featureEntity)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	feature := featureEntity.CreateFeature()

	if xrfc.DoesFeatureExist(feature.ID) {
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, fmt.Sprintf("Entity with id: %s already exists", feature.ID))
		return
	}
	if feature.ApplicationType != applicationType {
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, fmt.Sprintf("Entity with id: %s applicationType doesn't match", feature.ID))
		return
	}
	isValid, errorMsg := xrfc.IsValidFeature(feature)
	if !isValid {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorMsg)
		return
	}
	doesFeatureInstanceExist := xrfc.DoesFeatureNameExistForAnotherIdForApplicationType(feature, applicationType)
	if doesFeatureInstanceExist {
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, fmt.Sprintf("Feature with such featureInstance already exists: %s", feature.FeatureName))
		return
	}
	feature, err = FeaturePost(feature)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
	}
	response, _ := util.XConfJSONMarshal(feature, true)
	xwhttp.WriteXconfResponse(w, http.StatusCreated, []byte(response))
}

func PutFeatureHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY)
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
	var featureEntity *xwrfc.FeatureEntity
	err = json.Unmarshal([]byte(body), &featureEntity)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	feature := featureEntity.CreateFeature()
	if feature.ID == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Entity id is empty")
		return
	}

	if !xrfc.DoesFeatureExistWithApplicationType(feature.ID, applicationType) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Entity with id: %s does not exist", feature.ID))
		return
	}
	isValid, errorMsg := xrfc.IsValidFeature(feature)
	if !isValid {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorMsg)
		return
	}
	doesFeatureInstanceExist := xrfc.DoesFeatureNameExistForAnotherIdForApplicationType(feature, applicationType)
	if doesFeatureInstanceExist {
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, fmt.Sprintf("Feature with such featureInstance already exists: %s", feature.FeatureName))
		return
	}
	feature, err = PutFeature(feature)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
	}
	response, _ := util.XConfJSONMarshal(feature, true)
	xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
}

func GetFeaturesFilteredHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	var pageSize int
	var pageNumber int
	queryParams := map[string]string{}
	xwutil.AddQueryParamsToContextMap(r, queryParams)
	if len(queryParams) <= 0 {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(""))
		return
	}
	pageSize, err1 := strconv.Atoi(queryParams["pageSize"])
	pageNumber, err2 := strconv.Atoi(queryParams["pageNumber"])
	if err1 != nil || err2 != nil || pageSize < 0 {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(""))
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "responsewriter cast error")
		return
	}
	body := xw.Body()
	contextMap := make(map[string]string)
	if body != "" {
		err := json.Unmarshal([]byte(body), &contextMap)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	contextMap[xwcommon.APPLICATION_TYPE] = applicationType

	features := GetFeatureFiltered(contextMap)
	sort.SliceStable(features, func(i, j int) bool {
		return strings.Compare(strings.ToLower(features[i].ID), strings.ToLower(features[j].ID)) < 0
	})
	featuresPerPage := GetFeaturesWithPageNumbers(features, pageNumber, pageSize)
	response, _ := util.XConfJSONMarshal(featuresPerPage, true)
	featureSizeHeader := xhttp.CreateNumberOfItemsHttpHeaders(len(features))
	xwhttp.WriteXconfResponseWithHeaders(w, featureSizeHeader, http.StatusOK, []byte(response))
}

func GetFeaturesByIdListHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, "responsewriter cast error")
		return
	}
	var featureIdList []string
	if err := json.Unmarshal([]byte(xw.Body()), &featureIdList); err != nil {
		response := "Unable to extract featureIds from json file:" + err.Error()
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, response)
		return
	}

	features := GetFeaturesByIdList(featureIdList)
	response, _ := util.JSONMarshal(features)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}
