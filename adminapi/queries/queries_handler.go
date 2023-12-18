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
	"net/http"
	"sort"
	"strconv"
	"strings"

	xshared "xconfadmin/shared"
	xutil "xconfadmin/util"
	"xconfwebconfig/shared"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	daef "xconfwebconfig/dataapi/estbfirmware"

	xcommon "xconfwebconfig/common"

	"xconfadmin/common"
	xcoreef "xconfadmin/shared/estbfirmware"
	coreef "xconfwebconfig/shared/estbfirmware"
	corefw "xconfwebconfig/shared/firmware"
	"xconfwebconfig/util"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xcorefw "xconfadmin/shared/firmware"
	xwhttp "xconfwebconfig/http"
)

func GetQueriesPercentageBean(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	contextMap := make(map[string]string)
	xutil.AddQueryParamsToContextMap(r, contextMap)

	var result interface{}

	fieldName, found := contextMap[xcommon.FIELD]
	if found {
		result, err = GetPercentageBeanFilterFieldValues(fieldName, applicationType)
	} else {
		result, err = GetAllPercentageBeansFromDB(applicationType, true, true)
	}
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	_, ok := contextMap[common.EXPORT]
	if ok {
		percentageBeansToExport := make(map[string]interface{})
		percentageBeansToExport["percentageBeans"] = result
		res, err := xhttp.ReturnJsonResponse(percentageBeansToExport, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_ENV_MODEL_PERCENTAGE_BEANS + "_" + applicationType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
	}
}

func GetQueriesPercentageBeanById(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	bean, err := GetOnePercentageBeanFromDB(id)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "Entity with id: "+id+" does not exist")
		return
	}
	replaceFieldsWithFirmwareVersion(bean)
	if applicationType != bean.ApplicationType {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "ApplicationType doesn't match")
		return
	}
	res, err := xhttp.ReturnJsonResponse(bean, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func CreatePercentageBeanHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	percentageBean := coreef.NewPercentageBean()
	err = json.Unmarshal([]byte(body), &percentageBean)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if percentageBean.ApplicationType == "" {
		percentageBean.ApplicationType = applicationType
	}

	respEntity := CreatePercentageBean(percentageBean, applicationType)
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

func UpdatePercentageBeanHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	percentageBean := coreef.NewPercentageBean()
	err = json.Unmarshal([]byte(body), &percentageBean)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdatePercentageBean(percentageBean, applicationType)
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

func DeletePercentageBeanHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	if util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is empty")
		return
	}

	respEntity := DeletePercentageBean(id, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func GetQueriesEnvironments(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result := shared.GetAllEnvironmentList()
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	queryParams := r.URL.Query()
	_, ok := queryParams[common.EXPORT]
	if ok {
		headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_ALL_ENVIRONMENTS)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
	}
}

func GetQueriesEnvironmentsById(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	id = strings.ToUpper(id)

	env := GetEnvironment(id)
	if env == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "Environment does not exist")
		return
	}

	res, err := xhttp.ReturnJsonResponse(env, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	queryParams := r.URL.Query()
	_, ok := queryParams[common.EXPORT]
	if ok {
		envlist := []shared.Environment{*env}
		exres, err := xhttp.ReturnJsonResponse(envlist, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}

		headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_ENVIRONMENT + env.ID)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, exres)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
	}
}

func CreateEnvironmentHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	newEnv := shared.Environment{}
	err := json.Unmarshal([]byte(body), &newEnv)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := CreateEnvironment(&newEnv)
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

func DeleteEnvironmentHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	id = strings.ToUpper(id)
	respEntity := DeleteEnvironment(id)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func GetQueriesModels(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result := GetModels()
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesModelsById(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	model := GetModel(id)
	if model == nil {
		values, ok := r.URL.Query()[xcommon.VERSION]
		if ok {
			apiVersion := values[0]
			if util.IsVersionGreaterOrEqual(apiVersion, 3.0) {
				xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "Model with id "+id+" does not exist")
				return
			}
			xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(""))
			return
		}
	}

	res, err := xhttp.ReturnJsonResponse(model, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func CreateModelHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	newModel := shared.Model{}
	err := json.Unmarshal([]byte(body), &newModel)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := CreateModel(&newModel)
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

func UpdateModelHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	newModel := shared.Model{}
	err := json.Unmarshal([]byte(body), &newModel)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdateModel(&newModel)
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

func DeleteModelHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	id = strings.ToUpper(id)

	respEntity := DeleteModel(id)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func getQueriesFirmwareConfigsASFlavor(w http.ResponseWriter, r *http.Request, app string) {
	result := GetFirmwareConfigsAS(app)
	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(strings.ToLower(result[i].Description), strings.ToLower(result[j].Description)) < 0
	})
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesFirmwareConfigsById(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	errorStr := fmt.Sprintf("\"FirmwareConfig with id %s does not exist\"", id)
	fc := GetFirmwareConfigByIdAS(id)
	if fc == nil {
		values, ok := r.URL.Query()[xcommon.VERSION]
		if ok {
			apiVersion := values[0]
			if util.IsVersionGreaterOrEqual(apiVersion, 3.0) {
				xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
				return
			}
			xwhttp.WriteXconfResponse(w, http.StatusOK, nil)
			return
		}
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	if fc.ApplicationType != applicationType {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	firmwareConfig := fc.CreateFirmwareConfigResponse()

	res, err := xhttp.ReturnJsonResponse(firmwareConfig, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesFirmwareConfigsByIdASFlavor(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	firmwareConfig := GetFirmwareConfigByIdAS(id)
	if firmwareConfig != nil {
		res, err := xhttp.ReturnJsonResponse(firmwareConfig, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
		return
	}
	if firmwareConfig.ApplicationType != applicationType {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "FirmwareConfig with id "+id+" not found in "+applicationType)
		return
	}

	values, ok := r.URL.Query()[xcommon.VERSION]
	if ok {
		apiVersion := values[0]
		if util.IsVersionGreaterOrEqual(apiVersion, 3.0) {
			errorStr := fmt.Sprintf("FirmwareConfig with id %s does not exist", id)
			xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
			return
		}
	}
	xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "Firmware Config with current id:"+id+" does not exist")
}

func GetQueriesFirmwareConfigsByModelId(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	modelId, found := mux.Vars(r)[xcommon.MODEL_ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.MODEL_ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	model := shared.GetOneModel(modelId)
	if model == nil {
		errorStr := fmt.Sprintf("%v not found", modelId)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	configs := GetFirmwareConfigsByModelIdAndApplicationType(modelId, applicationType)
	res, err := xhttp.ReturnJsonResponse(configs, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesFirmwareConfigsByModelIdASFlavor(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	modelId, found := mux.Vars(r)[xcommon.MODEL_ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.MODEL_ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	model := shared.GetOneModel(modelId)
	if model == nil {
		errorStr := fmt.Sprintf("%v not found", modelId)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}

	configs := GetFirmwareConfigsByModelIdAndApplicationTypeAS(modelId, applicationType)
	res, err := xhttp.ReturnJsonResponse(configs, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func CreateFirmwareConfigHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	firmwareConfig := coreef.NewEmptyFirmwareConfig()
	err = json.Unmarshal([]byte(body), &firmwareConfig)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if firmwareConfig.ApplicationType != applicationType {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "ApplicationType mismatch")
		return
	}

	respEntity := CreateFirmwareConfig(firmwareConfig, applicationType)
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

func UpdateFirmwareConfigHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	firmwareConfig := coreef.NewEmptyFirmwareConfig()
	err = json.Unmarshal([]byte(body), &firmwareConfig)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if firmwareConfig.ApplicationType != applicationType {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "ApplicationType mismatch")
		return
	}

	respEntity := UpdateFirmwareConfig(firmwareConfig, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
}

func DeleteFirmwareConfigHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	if util.IsBlank(id) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Id is empty")
		return
	}

	respEntity := DeleteFirmwareConfig(id, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func DeleteFirmwareConfigHandlerASFlavor(w http.ResponseWriter, r *http.Request) {
	appType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	id, found := mux.Vars(r)[xcommon.ID]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", xcommon.ID)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	respEntity := DeleteFirmwareConfig(id, appType)
	status := respEntity.Status
	err = respEntity.Error

	if err != nil {
		xhttp.WriteAdminErrorResponse(w, status, err.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, status, nil)
}

func GetQueriesRulesIps(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	ipRuleService := daef.IpRuleService{}
	ipRuleBeans := ipRuleService.GetByApplicationType(applicationType)
	ipRuleBeansResponse := []*IpRuleBeanResponse{}
	for _, ipRuleBean := range ipRuleBeans {
		ipRuleBeanResponse := ConvertIpRuleBeanToIpRuleBeanResponse(ipRuleBean)
		ipRuleBeansResponse = append(ipRuleBeansResponse, ipRuleBeanResponse)
	}

	response, err := xhttp.ReturnJsonResponse(ipRuleBeansResponse, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetQueriesRulesMacs(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	var apiVersion string
	values, ok := r.URL.Query()[xcommon.VERSION]
	if ok {
		apiVersion = values[0]
	}
	macRuleService := daef.MacRuleService{}
	macRuleBeans := macRuleService.GetRulesWithMacCondition(applicationType)
	macRuleBeansResponse := []*MacRuleBeanResponse{}
	for _, macRuleBean := range macRuleBeans {
		if macRuleBean != nil {
			macRuleBean = wrap(macRuleBean, apiVersion)
			macRuleBeanResponse := ConvertMacRuleBeanToMacRuleBeanResponse(macRuleBean)
			macRuleBeansResponse = append(macRuleBeansResponse, macRuleBeanResponse)
		}
	}

	response, err := xhttp.ReturnJsonResponse(macRuleBeansResponse, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetQueriesRulesEnvModels(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	emRuleService := daef.EnvModelRuleService{}
	emRuleBeans := emRuleService.GetByApplicationType(applicationType)
	envModelRulesResponse := []*EnvModelRuleBeanResponse{}
	for _, emRuleBean := range emRuleBeans {
		envModelRuleResponse := ConvertEnvModelRuleBeanToEnvModelRuleBeanResponse(emRuleBean)
		envModelRulesResponse = append(envModelRulesResponse, envModelRuleResponse)
	}

	response, err := xhttp.ReturnJsonResponse(envModelRulesResponse, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetQueriesFiltersDownloadLocation(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id := xcoreef.GetRoundRobinIdByApplication(applicationType)

	singletonFilterValue, err := coreef.GetDownloadLocationRoundRobinFilterValOneDB(id)
	if err != nil {
		log.Errorf("unable to get singleton filter value. error: %+v", err)
		xhttp.AdminError(w, err)
		return
	}

	res, err := xhttp.ReturnJsonResponse(singletonFilterValue, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func UpdateDownloadLocationFilterHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	locationRoundRobinFilter := coreef.NewEmptyDownloadLocationRoundRobinFilterValue()
	err = json.Unmarshal([]byte(body), &locationRoundRobinFilter)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if applicationType != locationRoundRobinFilter.ApplicationType {
		xhttp.AdminError(w, common.NewXconfError(http.StatusBadRequest, "ApplicationType Mismatch"))
		return
	}

	respEntity := UpdateDownloadLocationRoundRobinFilter(applicationType, locationRoundRobinFilter)
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

func GetQueriesFiltersIps(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result, err := coreef.IpFiltersByApplicationType(applicationType)
	if err != nil {
		log.Errorf("unable to get ip filter value. error: %+v", err)
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesFiltersIpsByName(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name := mux.Vars(r)[common.NAME]
	if name == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Filter name is empty")
		return
	}

	result, err := coreef.IpFilterByName(name, applicationType)
	if err != nil {
		log.Errorf("unable to get ip filter value. error: %+v", err)
	}

	if result == nil {
		xwhttp.WriteXconfResponse(w, http.StatusOK, []byte{})
		return
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func UpdateIpsFilterHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	ipFilter := coreef.NewEmptyIpFilter()
	err = json.Unmarshal([]byte(body), &ipFilter)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdateIpFilter(applicationType, ipFilter)
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

func DeleteIpsFilterHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name, found := mux.Vars(r)[common.NAME]
	if !found || len(strings.TrimSpace(name)) == 0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is empty")
		return
	}

	respEntity := DeleteIpsFilter(name, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func GetQueriesFiltersTime(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result, err := coreef.TimeFiltersByApplicationType(applicationType)
	if err != nil {
		log.Errorf("unable to get ip filter value. error: %+v", err)
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesFiltersTimeByName(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name := mux.Vars(r)[common.NAME]
	if name == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Filter name is empty")
		return
	}

	result, err := coreef.TimeFilterByName(name, applicationType)
	if err != nil {
		log.Errorf("unable to get ip filter value. error: %+v", err)
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func UpdateTimeFilterHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	timeFilter := xcoreef.NewEmptyTimeFilter()
	err = json.Unmarshal([]byte(body), &timeFilter)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdateTimeFilter(applicationType, timeFilter)
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

func DeleteTimeFilterHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name, found := mux.Vars(r)[common.NAME]
	if !found || len(strings.TrimSpace(name)) == 0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is empty")
		return
	}

	respEntity := DeleteTimeFilter(name, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func GetQueriesFiltersLocation(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result, err := coreef.DownloadLocationFiltersByApplicationType(applicationType)
	if err != nil {
		log.Errorf("unable to get ip filter value. error: %+v", err)
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesFiltersLocationByName(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name := mux.Vars(r)[common.NAME]
	if name == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Filter name is empty")
		return
	}

	result, err := coreef.DownloadLocationFiltersByName(applicationType, name)
	if err != nil {
		log.Errorf("unable to get ip filter value. error: %+v", err)
	}
	if result == nil {
		values, ok := r.URL.Query()[xcommon.VERSION]
		if ok {
			apiVersion := values[0]
			if util.IsVersionGreaterOrEqual(apiVersion, 3.0) {
				errorStr := fmt.Sprintf("LocationFilter with name %s does not exist", name)
				xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
				return
			}
		} else {
			xwhttp.WriteXconfResponse(w, http.StatusOK, []byte{})
			return
		}
	}
	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func UpdateLocationFilterHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	locationFilter := coreef.DownloadLocationFilter{}
	err = json.Unmarshal([]byte(body), &locationFilter)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdateLocationFilter(applicationType, &locationFilter)
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

func DeleteLocationFilterHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name, found := mux.Vars(r)[common.NAME]
	if !found || len(strings.TrimSpace(name)) == 0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is empty")
		return
	}

	respEntity := DeleteLocationFilter(name, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func GetQueriesFiltersPercent(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	contextMap := make(map[string]string)
	xutil.AddQueryParamsToContextMap(r, contextMap)
	var result interface{}

	fieldName, found := contextMap[xcommon.FIELD]
	if found {
		result, err = GetPercentFilterFieldValues(fieldName, applicationType)
	} else {
		percentFilter, err := GetPercentFilter(applicationType)
		if err == nil {
			result = xcoreef.NewPercentFilterWrapper(percentFilter, true)
		}
	}
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	_, ok := contextMap[common.EXPORT]
	if ok {
		headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_PERCENT_FILTER + "_" + applicationType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
	}
}

func UpdatePercentFilterHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	percentFilter := xcoreef.NewEmptyPercentFilterWrapper()
	err = json.Unmarshal([]byte(body), &percentFilter)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdatePercentFilter(applicationType, percentFilter)
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

func GetQueriesFiltersRebootImmediately(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	result, err := coreef.RebootImmediatelyFiltersByApplicationType(applicationType)
	if err != nil {
		log.Errorf("unable to get reboot immediately filter value. error: %+v", err)
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func GetQueriesFiltersRebootImmediatelyByName(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name := mux.Vars(r)[common.NAME]
	if name == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is empty")
		return
	}

	result, err := xcoreef.RebootImmediatelyFiltersByName(applicationType, name)
	if err != nil {
		log.Errorf("unable to get ip filter value. error: %+v", err)
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}

func UpdateRebootImmediatelyHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	rebootFilter := xcoreef.NewEmptyRebootImmediatelyFilter()
	err = json.Unmarshal([]byte(body), &rebootFilter)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdateRebootImmediatelyFilter(applicationType, rebootFilter)
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

func DeleteRebootImmediatelyHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name, found := mux.Vars(r)[common.NAME]
	if !found || len(strings.TrimSpace(name)) == 0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is empty")
		return
	}

	respEntity := DeleteRebootImmediatelyFilter(name, applicationType)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xwhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func GetRoundRobinFilterHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id := xcoreef.GetRoundRobinIdByApplication(applicationType)

	singletonFilterValue, err := coreef.GetDownloadLocationRoundRobinFilterValOneDB(id)
	if err != nil {
		log.Errorf("unable to get singleton filter value. error: %+v", err)
	}

	if singletonFilterValue == nil {
		singletonFilterValue = coreef.NewEmptyDownloadLocationRoundRobinFilterValue()
		singletonFilterValue.ApplicationType = applicationType
	}

	res, err := xhttp.ReturnJsonResponse(singletonFilterValue, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	if _, ok := r.URL.Query()[common.EXPORT]; ok {
		fileName := common.ExportFileNames_ROUND_ROBIN_FILTER + "_" + singletonFilterValue.ApplicationType
		headers := xhttp.CreateContentDispositionHeader(fileName)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteXconfResponse(w, http.StatusOK, res)
	}
}

func GetIpRuleById(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	ruleName, found := mux.Vars(r)[common.RULE_NAME]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.RULE_NAME)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	if ruleName == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is empty")
		return
	}

	var apiVersion string
	values, ok := r.URL.Query()[xcommon.VERSION]
	if ok {
		apiVersion = values[0]
	}
	var ipRuleBean *coreef.IpRuleBean
	ipRuleService := daef.IpRuleService{}
	ipRuleBeans := ipRuleService.GetByApplicationType(applicationType)
	for _, bean := range ipRuleBeans {
		if bean.Name == ruleName {
			ipRuleBean = bean
			break
		}
	}
	if ipRuleBean != nil {
		ipRuleBeanResponse := ConvertIpRuleBeanToIpRuleBeanResponse(ipRuleBean)
		response, err := xhttp.ReturnJsonResponse(ipRuleBeanResponse, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
		return
	}
	if util.IsVersionGreaterOrEqual(apiVersion, 3.0) {
		errorStr := fmt.Sprintf("IpRule with name %s does not exist", ruleName)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, nil)
}

func GetIpRuleByIpAddressGroup(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	ipAddressGroupName, found := mux.Vars(r)[common.IP_ADDRESS_GROUP_NAME]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.IP_ADDRESS_GROUP_NAME)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	if ipAddressGroupName == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "IpAddressGroup id is empty")
		return
	}

	ipRules := []*IpRuleBeanResponse{}
	ipRuleService := daef.IpRuleService{}
	ipRuleBeans := ipRuleService.GetByApplicationType(applicationType)
	for _, bean := range ipRuleBeans {
		if bean.IpAddressGroup != nil && ipAddressGroupName == bean.IpAddressGroup.Name {
			xcoreef.AddExpressionToIpRuleBean(bean)
			ipRuleBeanResponse := ConvertIpRuleBeanToIpRuleBeanResponse(bean)
			ipRules = append(ipRules, ipRuleBeanResponse)
		}
	}

	response, err := xhttp.ReturnJsonResponse(ipRules, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func UpdateIpRule(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	ipRuleBean := coreef.IpRuleBean{}
	err = json.Unmarshal([]byte(body), &ipRuleBean)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ipRuleBean_origin := ipRuleBean
	if ipRuleBean.Name == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is empty")
		return
	}
	if err := corefw.ValidateRuleName(ipRuleBean.Id, ipRuleBean.Name); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if ipRuleBean.EnvironmentId == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Environment id is empty")
		return
	}
	ipRuleBean.EnvironmentId = strings.ToUpper(ipRuleBean.EnvironmentId)
	if !IsExistEnvironment(ipRuleBean.EnvironmentId) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Environment "+ipRuleBean.EnvironmentId+" does not exist")
		return
	}
	if ipRuleBean.ModelId == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Model id is empty")
		return
	}
	ipRuleBean.ModelId = strings.ToUpper(ipRuleBean.ModelId)
	if !IsExistModel(ipRuleBean.ModelId) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Model "+ipRuleBean.ModelId+" does not exist")
		return
	}
	firmwareConfig := ipRuleBean.FirmwareConfig
	if firmwareConfig != nil && firmwareConfig.ID != "" {
		firmwareConfig, err = coreef.GetFirmwareConfigOneDB(firmwareConfig.ID)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "FirmwareConfig with id does not exist")
			return
		}
	}
	if firmwareConfig != nil && applicationType != firmwareConfig.ApplicationType {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "ApplicationType of FirmwareRule and FirmwareConfig does not match")
		return
	}
	if firmwareConfig != nil && !IsValidFirmwareConfigByModelIds(ipRuleBean.ModelId, applicationType, firmwareConfig) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Firmware config does not support this model")
		return
	}
	if ipRuleBean.IpAddressGroup == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Ip address group is not specified")
		return
	}
	if ipRuleBean.IpAddressGroup != nil && IsChangedIpAddressGroup(ipRuleBean.IpAddressGroup) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "IP address group is not matched by existed IP address group")
		return
	}

	ipRuleService := daef.IpRuleService{}
	oldIpRuleBeans := ipRuleService.GetByApplicationType(applicationType)
	for _, oldBean := range oldIpRuleBeans {
		if oldBean.Name == ipRuleBean.Name {
			if ipRuleBean.Id == "" {
				ipRuleBean.Id = oldBean.Id
			} else if ipRuleBean.Id != oldBean.Id {
				xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Ip rule with current name exists")
				return
			}
			break
		}
	}
	firmwareRule := coreef.ConvertIpRuleBeanToFirmwareRule(&ipRuleBean)
	if applicationType != "" {
		firmwareRule.ApplicationType = applicationType
	}
	err = xcorefw.CreateFirmwareRuleOneDBAfterValidate(firmwareRule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("DB error: %v", err))
		return
	}
	if ipRuleBean_origin.Id == "" {
		ipRuleBean_origin.Id = firmwareRule.ID
	}
	if ipRuleBean.FirmwareConfig == nil {
		ipRuleBean_origin.Noop = true
	} else {
		ipRuleBean_origin.Noop = false
	}

	response, err := xhttp.ReturnJsonResponse(ipRuleBean_origin, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetMACRuleByName(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	ruleName, found := mux.Vars(r)[common.RULE_NAME]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.RULE_NAME)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	if ruleName == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is empty")
		return
	}

	var apiVersion string
	values, ok := r.URL.Query()[xcommon.VERSION]
	if ok {
		apiVersion = values[0]
	}
	var macRuleBean *coreef.MacRuleBean
	macRuleService := daef.MacRuleService{}
	macRuleBeans := macRuleService.GetRulesWithMacCondition(applicationType)
	for _, mrBean := range macRuleBeans {
		if mrBean.Name == ruleName {
			macRuleBean = mrBean
		}
	}
	if macRuleBean != nil {
		macRuleBean = wrap(macRuleBean, apiVersion)
		macRuleBeanResponse := ConvertMacRuleBeanToMacRuleBeanResponse(macRuleBean)
		response, err := xhttp.ReturnJsonResponse(macRuleBeanResponse, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		xwhttp.WriteXconfResponse(w, http.StatusOK, response)
		return
	}
	if util.IsVersionGreaterOrEqual(apiVersion, 3.0) {
		errorStr := fmt.Sprintf("MacRule with name %s does not exist", ruleName)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, nil)
}

func GetMACRulesByMAC(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	macAddress, found := mux.Vars(r)[common.MAC_ADDRESS]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.MAC_ADDRESS)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	var apiVersion string
	values, ok := r.URL.Query()[xcommon.VERSION]
	if ok {
		apiVersion = values[0]
	}
	result := []*coreef.MacRuleBeanResponse{}
	if util.IsValidMacAddress(macAddress) {
		macRuleService := daef.MacRuleService{}
		macRuleBeans := macRuleService.SearchMacRules(macAddress, applicationType)
		for _, macRule := range macRuleBeans {
			macRule = wrap(macRule, apiVersion)
			result = append(result, coreef.MacRuleBeanToMacRuleBeanResponse(macRule))
		}
	}

	response, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func wrap(bean *coreef.MacRuleBean, apiVersion string) *coreef.MacRuleBean {
	version := 1.0
	if apiVersion != "" {
		floatVersion, err := strconv.ParseFloat(apiVersion, 64)
		if err == nil {
			version = floatVersion
		}
	}
	if version >= 2.0 {
		if bean.MacListRef != "" {
			macList := GetNamespacedListByIdAndType(bean.MacListRef, shared.MAC_LIST)
			if macList == nil {
				bean.MacList = &[]string{}
			} else {
				bean.MacList = &macList.Data
			}
		}
		return bean
	}
	bean.Id = ""
	bean.MacList = nil
	return bean
}

func SaveMACRule(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	macRule := coreef.MacRuleBean{}
	err = json.Unmarshal([]byte(body), &macRule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if macRule.Name == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Rule name is empty")
		return
	}
	if macRule.MacListRef == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "MAC address list is empty or blank")
		return
	}
	if err := corefw.ValidateRuleName(macRule.Id, macRule.Name); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	nsList := GetNamespacedListByIdAndType(macRule.MacListRef, shared.MAC_LIST)
	if nsList == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Mac list does not exist")
		return
	}
	if len(*macRule.TargetedModelIds) < 1 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Model list is not specified")
		return
	}
	for _, modelId := range *macRule.TargetedModelIds {
		if !IsExistModel(modelId) {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Model "+modelId+" does not exist")
			return
		}
	}
	if macRule.FirmwareConfig == nil || macRule.FirmwareConfig.ID == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Firmware configuration is not specified")
		return
	}
	firmwareConfig, err := coreef.GetFirmwareConfigOneDB(macRule.FirmwareConfig.ID)
	if err != nil || firmwareConfig == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Firmware configuration does not exist")
		return
	}
	if !xshared.ApplicationTypeEquals(applicationType, firmwareConfig.ApplicationType) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "ApplicationType of FirmwareConfig and MacRule does not match")
		return
	}
	if !IsValidFirmwareConfigByModelIdList(macRule.TargetedModelIds, applicationType, firmwareConfig) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Firmware configuration does not support this model")
		return
	}
	var ruleToUpdate *coreef.MacRuleBean
	macRuleService := daef.MacRuleService{}
	macRuleBeans := macRuleService.GetByApplicationType(applicationType)
	for _, rule := range macRuleBeans {
		if macRule.Name == rule.Name {
			ruleToUpdate = rule
			continue
		}
		if macRule.MacListRef == rule.MacListRef {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "MAC addresses list is already used in another rule: "+rule.Name)
			return
		}
	}
	status := http.StatusCreated
	if ruleToUpdate != nil {
		macRule.Id = ruleToUpdate.Id
		status = http.StatusOK
	}
	models := *macRule.TargetedModelIds
	dbModels := shared.GetAllModelList() //[]*shared.Model
	for _, m := range models {
		found := false
		for _, d := range dbModels {
			if d.ID == m {
				found = true
				break
			}
		}
		if !found {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Model list contains not existed models")
			return
		}
	}
	supportedModelIds := make([]string, len(firmwareConfig.SupportedModelIds))
	copy(supportedModelIds, firmwareConfig.SupportedModelIds)
	found := false
	for _, item := range supportedModelIds {
		for _, tgtModel := range *macRule.TargetedModelIds {
			if item == tgtModel {
				found = true
				break
			}
		}
	}
	if !found {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Firmware configuration doesn't support given models")
		return
	}
	macRule.FirmwareConfig = firmwareConfig
	if macRule.Id == "" {
		macRule.Id = uuid.New().String()
	}
	firmwareRule := xcoreef.ConvertMacRuleBeanToFirmwareRule(&macRule)
	if applicationType != "" {
		firmwareRule.ApplicationType = applicationType
	}
	err = xcorefw.CreateFirmwareRuleOneDBAfterValidate(firmwareRule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("DB error: %v", err))
		return
	}

	response, err := xhttp.ReturnJsonResponse(macRule, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, status, response)
}

func DeleteMACRule(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name, found := mux.Vars(r)[common.NAME]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.NAME)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	if name == "" {
		errorStr := "Name is empty"
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}

	var macRuleBean *coreef.MacRuleBean
	macRuleService := daef.MacRuleService{}
	macRuleBeans := macRuleService.GetByApplicationType(applicationType)
	for _, mrBean := range macRuleBeans {
		if mrBean.Name == name {
			macRuleBean = mrBean
		}
	}
	if macRuleBean != nil {
		err := corefw.DeleteOneFirmwareRule(macRuleBean.Id)
		if err != nil {
			xwhttp.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("DB error: %v", err))
			return
		}
	}
	xwhttp.WriteXconfResponse(w, http.StatusNoContent, []byte{})
}

func GetEnvModelRuleByNameHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name, found := mux.Vars(r)[common.NAME]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.NAME)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	if name == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "EnvModelRule name is empty")
		return
	}

	var envModelRule *coreef.EnvModelBean
	emRuleService := daef.EnvModelRuleService{}
	emRuleBeans := emRuleService.GetByApplicationType(applicationType)
	for _, emRuleBean := range emRuleBeans {
		if strings.EqualFold(emRuleBean.Name, name) {
			envModelRule = emRuleBean
			break
		}
	}
	var apiVersion string
	values, ok := r.URL.Query()[xcommon.VERSION]
	if ok {
		apiVersion = values[0]
	}
	if envModelRule == nil && util.IsVersionGreaterOrEqual(apiVersion, 3.0) {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "EnvModelRule with name "+name+" does not exist")
		return
	}
	if envModelRule == nil {
		xwhttp.WriteXconfResponse(w, http.StatusOK, nil)
		return
	}
	envModelRuleResponse := ConvertEnvModelRuleBeanToEnvModelRuleBeanResponse(envModelRule)

	response, err := xhttp.ReturnJsonResponse(envModelRuleResponse, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func UpdateEnvModelRuleHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, common.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	envModelRuleBean := coreef.EnvModelBean{}
	err = json.Unmarshal([]byte(body), &envModelRuleBean)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if envModelRuleBean.Name == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is empty")
		return
	}
	if envModelRuleBean.EnvironmentId == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Environment id is empty")
		return
	}
	if err := corefw.ValidateRuleName(envModelRuleBean.Id, envModelRuleBean.Name); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if !IsExistEnvironment(envModelRuleBean.EnvironmentId) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Environment does not exist")
		return
	}
	if envModelRuleBean.ModelId == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Model is empty")
		return
	}
	if !IsExistModel(envModelRuleBean.ModelId) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Model does not exist")
		return
	}
	envModelRuleBean.EnvironmentId = strings.ToUpper(envModelRuleBean.EnvironmentId)
	firmwareConfig := envModelRuleBean.FirmwareConfig
	if firmwareConfig != nil && firmwareConfig.ID != "" {
		firmwareConfig, err = coreef.GetFirmwareConfigOneDB(firmwareConfig.ID)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "FirmwareConfig with id does not exist")
			return
		}
	}
	if firmwareConfig != nil && !xshared.ApplicationTypeEquals(applicationType, firmwareConfig.ApplicationType) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "ApplicationType of EnvModelRule and FirmwareConfig does not match")
		return
	}
	if firmwareConfig != nil && !IsValidFirmwareConfigByModelIds(envModelRuleBean.ModelId, applicationType, firmwareConfig) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "FirmwareConfig does not support this model")
		return
	}
	if firmwareConfig != nil {
		envModelRuleBean.FirmwareConfig = firmwareConfig
	}
	envModelRuleBean.ModelId = strings.ToUpper(envModelRuleBean.ModelId)
	emRuleService := daef.EnvModelRuleService{}
	emRuleBeans := emRuleService.GetByApplicationType(applicationType)
	for _, emRuleBean := range emRuleBeans {
		if strings.EqualFold(emRuleBean.Name, envModelRuleBean.Name) && !strings.EqualFold(emRuleBean.Id, envModelRuleBean.Id) {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is already used")
			return
		}
		if strings.EqualFold(emRuleBean.EnvironmentId, envModelRuleBean.EnvironmentId) && strings.EqualFold(emRuleBean.ModelId, envModelRuleBean.ModelId) {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Env/Model overlap with rule: "+emRuleBean.Name)
			return
		}
	}
	if envModelRuleBean.Id == "" {
		envModelRuleBean.Id = uuid.New().String()
	}
	firmwareRule := xcoreef.ConvertModelRuleBeanToFirmwareRule(&envModelRuleBean)
	err = xcorefw.CreateFirmwareRuleOneDBAfterValidate(firmwareRule)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("DB error: %v", err))
		return
	}

	response, err := xhttp.ReturnJsonResponse(envModelRuleBean, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func DeleteEnvModelRuleBeanHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name, found := mux.Vars(r)[common.NAME]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.NAME)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	emRuleService := daef.EnvModelRuleService{}
	emRuleBeans := emRuleService.GetByApplicationType(applicationType)
	for _, emRuleBean := range emRuleBeans {
		if strings.EqualFold(emRuleBean.Name, name) {
			err := corefw.DeleteOneFirmwareRule(emRuleBean.Id)
			if err != nil {
				xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("DB error: %v", err))
				return
			}
			break
		}
	}
	xwhttp.WriteXconfResponse(w, http.StatusNoContent, []byte{})
}

func DeleteIpRule(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	name, found := mux.Vars(r)[common.NAME]
	if !found {
		errorStr := fmt.Sprintf("%v is invalid", common.NAME)
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, errorStr)
		return
	}
	if name == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is empty")
		return
	}
	var ipRuleBean *coreef.IpRuleBean
	ipRuleService := daef.IpRuleService{}
	ipRuleBeans := ipRuleService.GetByApplicationType(applicationType)
	for _, bean := range ipRuleBeans {
		if bean.Name == name {
			ipRuleBean = bean
			break
		}
	}
	if ipRuleBean != nil {
		ipRuleService.Delete(ipRuleBean.Id)
	}
	xwhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

// Depricated API so just return 200 status code
func GetMigrationInfoHandler(w http.ResponseWriter, r *http.Request) {
	res, err := xhttp.ReturnJsonResponse([]string{}, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}
