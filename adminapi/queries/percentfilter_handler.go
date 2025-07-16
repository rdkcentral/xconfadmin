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
	"math"
	"net/http"

	"xconfadmin/adminapi/auth"
	"xconfadmin/common"
	xhttp "xconfadmin/http"
	xshared "xconfadmin/shared"
	xcoreef "xconfadmin/shared/estbfirmware"
	xutil "xconfadmin/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/dchest/siphash"
	log "github.com/sirupsen/logrus"
)

func UpdatePercentFilterGlobal(applicationType string, globalPercentage *coreef.GlobalPercentage) *xwhttp.ResponseEntity {
	globalFwRule := xcoreef.ConvertGlobalPercentageIntoRule(globalPercentage, applicationType)
	globalFwRule.ID = GetGlobalPercentageIdByApplication(applicationType)
	ruleDb, err := firmware.GetFirmwareRuleOneDB(globalFwRule.ID)
	if err == nil || ruleDb != nil {
		err = updateFirmwareRule(*globalFwRule, applicationType, false)
	} else {
		err = createFirmwareRule(*globalFwRule, applicationType, false)
	}
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, globalPercentage)
}

func GetCalculatedHashAndPercentHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	macAddress := r.FormValue("esbMac")
	if macAddress == "" {
		http.Error(w, "Missing 'esbMac' parameter", http.StatusBadRequest)
		return
	}
	_, err = util.MACAddressValidator(macAddress)
	if err != nil {
		http.Error(w, "Invalid Estb Mac", http.StatusBadRequest)
		return
	}
	source := fmt.Sprintf(`"%v"`, macAddress)
	hashCode, percent := xshared.CalculateHashAndPercent(source)
	response := map[string]interface{}{
		"hashValue": hashCode,
		"percent":   percent,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func UpdatePercentFilterGlobalHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanWrite(r, auth.FIRMWARE_ENTITY)
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
	globalPercentage := coreef.NewGlobalPercentage()
	err = json.Unmarshal([]byte(body), &globalPercentage)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := UpdatePercentFilterGlobal(applicationType, globalPercentage)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(respEntity.Data, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, respEntity.Status, xhttp.ContextTypeHeader(r))
}

func GetPercentFilterGlobal(applicationType string) (*coreef.GlobalPercentage, error) {
	globalPercentageId := GetGlobalPercentageIdByApplication(applicationType)
	globalPercentageRule, err := firmware.GetFirmwareRuleOneDB(globalPercentageId)
	if err != nil {
		log.Warn(fmt.Sprintf("GetPercentFilter %v", err))
	}
	globalPercentage := coreef.NewGlobalPercentage()
	if globalPercentageRule != nil {
		globalPercentage = coreef.ConvertIntoGlobalPercentageFirmwareRule(globalPercentageRule)
	}
	return globalPercentage, nil
}

func GetPercentFilterGlobalHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	contextMap := make(map[string]string)
	xutil.AddQueryParamsToContextMap(r, contextMap)

	globalpercent, err := GetPercentFilterGlobal(applicationType)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("unable to get globalpercent reponse. error: %v", err))
		return
	}

	res, err := xhttp.ReturnJsonResponse(globalpercent, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	_, ok := contextMap[common.EXPORT]
	if ok {
		//TODO: rework with struct avoiding map type below
		percentageBeans, err := GetAllPercentageBeansFromDB(applicationType, true, false)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		percentFiltersToExport := make(map[string]interface{})
		percentFiltersToExport["percentageBeans"] = percentageBeans
		percentFiltersToExport["globalPercentage"] = globalpercent
		res, err := xhttp.ReturnJsonResponse(percentFiltersToExport, r)
		if err != nil {
			xhttp.AdminError(w, err)
			return
		}
		headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_PERCENT_FILTER + "_" + applicationType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
	}
}

func GetGlobalPercentFilter(applicationType string) (*coreef.PercentFilterVo, error) {
	globalPercentageId := GetGlobalPercentageIdByApplication(applicationType)
	globalPercentageRule, err := firmware.GetFirmwareRuleOneDB(globalPercentageId)
	if err != nil {
		log.Warn(fmt.Sprintf("GetPercentFilter %v", err))
	}
	globalPercentage := coreef.NewGlobalPercentage()
	if globalPercentageRule != nil {
		globalPercentage = coreef.ConvertIntoGlobalPercentageFirmwareRule(globalPercentageRule)
	}
	PercentfilterVo := coreef.NewDefaultPercentFilterVo()
	PercentfilterVo.GlobalPercentage = globalPercentage
	PercentfilterVo.GlobalPercentage.ApplicationType = applicationType
	return PercentfilterVo, nil
}

func GetGlobalPercentFilterHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	contextMap := make(map[string]string)
	xutil.AddQueryParamsToContextMap(r, contextMap)

	globalpercent, err := GetGlobalPercentFilter(applicationType)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("unable to get globalpercent reponse. error: %v", err))
		return
	}

	res, err := xhttp.ReturnJsonResponse(globalpercent, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	_, ok := contextMap[common.EXPORT]
	if ok {

		headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_GLOBAL_PERCENT + "_" + applicationType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
	}
}

const (
	SipHashKey0 = uint64(506097522914230528)
	SipHashKey1 = uint64(1084818905618843912)
)

func calculateHashAndPercent(macAddress string) (float64, float64) {
	voffset := float64(math.MaxInt64 + 1)
	vrange := float64(math.MaxInt64*2 + 1)
	bbytes := []byte(macAddress)
	hashCode := float64(int64(siphash.Hash(SipHashKey0, SipHashKey1, bbytes))) + voffset
	percent := (hashCode / vrange) * 100
	return hashCode, percent
}

func GetCalculatedHashAndPercent(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	macAddress := r.FormValue("esb_mac")
	if macAddress == "" {
		http.Error(w, "Missing 'esb_mac' parameter", http.StatusBadRequest)
		return
	}
	_, err = util.MACAddressValidator(macAddress)
	if err != nil {
		http.Error(w, "Invalid Estb Mac", http.StatusBadRequest)
		return
	}
	hashCode, percent := calculateHashAndPercent(macAddress)
	response := map[string]interface{}{
		"hashValue": hashCode,
		"percent":   percent,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
func GetGlobalPercentFilterAsRule(applicationType string) (*corefw.FirmwareRule, error) {
	globalPercentageId := GetGlobalPercentageIdByApplication(applicationType)
	globalPercentageRule, err := firmware.GetFirmwareRuleOneDB(globalPercentageId)
	if err != nil {
		log.Warn(fmt.Sprintf("GetPercentFilter %v", err))
		return nil, err
	}
	return globalPercentageRule, nil
}

func GetGlobalPercentFilterAsRuleHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	contextMap := make(map[string]string)
	xutil.AddQueryParamsToContextMap(r, contextMap)

	globalpercentasrule, err := GetGlobalPercentFilterAsRule(applicationType)
	if err != nil {
		globalPercentage := coreef.NewGlobalPercentage()
		globalpercentasrule = xcoreef.ConvertGlobalPercentageIntoRule(globalPercentage, applicationType)
	}
	resArray := []*corefw.FirmwareRule{globalpercentasrule}

	res, err := xhttp.ReturnJsonResponse(resArray, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	_, ok := contextMap[common.EXPORT]
	if ok {

		headers := xhttp.CreateContentDispositionHeader(common.ExportFileNames_GLOBAL_PERCENT_AS_RULE + "_" + applicationType)
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, res)
	} else {
		xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
	}
}
