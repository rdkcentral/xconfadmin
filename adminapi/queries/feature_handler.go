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

	xhttp "github.com/rdkcentral/xconfadmin/http"
	xrfc "github.com/rdkcentral/xconfadmin/shared/rfc"
	requtil "github.com/rdkcentral/xconfadmin/util"

	"github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared/rfc"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"

	"github.com/gorilla/mux"
)

func GetFeatureEntityHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	featureEntityList := []*rfc.FeatureEntity{}
	features := GetAllFeatureEntity()
	for _, rule := range features {
		if applicationType == rule.ApplicationType {
			featureEntityList = append(featureEntityList, rule)
		}
	}
	response, _ := util.XConfJSONMarshal(featureEntityList, true)
	xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
}

func GetFeatureEntityFilteredHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	contextMap := map[string]string{}
	requtil.AddQueryParamsToContextMap(r, contextMap)
	contextMap[common.APPLICATION_TYPE] = applicationType

	featureList := GetFeatureEntityFiltered(contextMap)
	response, _ := util.XConfJSONMarshal(featureList, true)
	xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
}

func GetFeatureEntityByIdHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id := mux.Vars(r)[common.ID]
	if id == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("\"Id is blank\""))
		return
	}
	featureEntity := GetFeatureEntityById(id)
	if featureEntity == nil {
		xwhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf("\"Entity with id: %s does not exist\"", id)))
		return
	}
	if featureEntity.ApplicationType != applicationType {
		errorStr := fmt.Sprintf("%v not found", id)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	response, _ := util.XConfJSONMarshal(featureEntity, true)
	xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
}

func PostFeatureEntityImportAllHandler(w http.ResponseWriter, r *http.Request) {
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte("responsewriter cast error"))
		return
	}
	body := xw.Body()
	var featureEntityList []*rfc.FeatureEntity
	err := json.Unmarshal([]byte(body), &featureEntityList)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf("\"%s\"", err.Error())))
		return
	}
	determinedAppType := ""
	for i, _ := range featureEntityList {
		applicationType, err := auth.CanWrite(r, auth.DCM_ENTITY, featureEntityList[i].ApplicationType)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		if determinedAppType != "" && determinedAppType != applicationType {
			xhttp.WriteAdminErrorResponse(w, http.StatusConflict, "ApplicationType mixing not allowed")
			return
		}
		if featureEntityList[i].ApplicationType == "" {
			featureEntityList[i].ApplicationType = applicationType
		} else if featureEntityList[i].ApplicationType != applicationType {
			xhttp.WriteAdminErrorResponse(w, http.StatusConflict, "ApplicationType Conflict")
			return
		}
		determinedAppType = applicationType
	}

	featureEntityMap := ImportOrUpdateAllFeatureEntity(featureEntityList, determinedAppType)
	response, _ := util.XConfJSONMarshal(featureEntityMap, true)
	xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
}

func PostFeatureEntityHandler(w http.ResponseWriter, r *http.Request) {
	var featureEntity rfc.FeatureEntity
	applicationType, err := auth.ExtractBodyAndCheckPermissions(&featureEntity, w, r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	if xrfc.DoesFeatureExist(featureEntity.ID) {
		xwhttp.WriteXconfResponse(w, http.StatusConflict, []byte(fmt.Sprintf("\"Entity with id: %s already exists\"", featureEntity.ID)))
		return
	}
	isValid, errorMsg := xrfc.IsValidFeatureEntity(&featureEntity)
	if !isValid {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf("\"%s\"", errorMsg)))
		return
	}
	doesFeatureInstanceExist := xrfc.DoesFeatureNameExistForAnotherEntityId(&featureEntity)
	if doesFeatureInstanceExist {
		xwhttp.WriteXconfResponse(w, http.StatusConflict, []byte(fmt.Sprintf("\"Feature with such featureInstance already exists: %s\"", featureEntity.FeatureName)))
		return
	}
	createdFeature, err := PostFeatureEntity(&featureEntity, applicationType)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, _ := util.XConfJSONMarshal(createdFeature, true)
	xwhttp.WriteXconfResponse(w, http.StatusCreated, []byte(response))
}

func PutFeatureEntityHandler(w http.ResponseWriter, r *http.Request) {
	var featureEntity rfc.FeatureEntity
	applicationType, err := auth.ExtractBodyAndCheckPermissions(&featureEntity, w, r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	if featureEntity.ID == "" {
		xwhttp.WriteXconfResponse(w, http.StatusNotFound, []byte("\"Entity id is empty\""))
		return
	}
	if !xrfc.DoesFeatureExist(featureEntity.ID) {
		xwhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf("\"Entity with id: %s does not exist\"", featureEntity.ID)))
		return
	}
	isValid, errorMsg := xrfc.IsValidFeatureEntity(&featureEntity)
	if !isValid {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf("\"%s\"", errorMsg)))
		return
	}
	doesFeatureInstanceExist := xrfc.DoesFeatureNameExistForAnotherEntityId(&featureEntity)
	if doesFeatureInstanceExist {
		xwhttp.WriteXconfResponse(w, http.StatusConflict, []byte(fmt.Sprintf("\"Feature with such featureInstance already exists: %s\"", featureEntity.FeatureName)))
		return
	}
	updatedFeature, err := PutFeatureEntity(&featureEntity, applicationType)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, _ := util.XConfJSONMarshal(updatedFeature, true)
	xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
}
