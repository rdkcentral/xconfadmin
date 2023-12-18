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
package telemetry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	xutil "xconfadmin/util"
	"xconfwebconfig/dataapi/dcm/telemetry"

	xcommon "xconfadmin/common"

	"xconfadmin/shared"
	xlogupload "xconfadmin/shared/logupload"
	xwcommon "xconfwebconfig/common"
	xwlogupload "xconfwebconfig/shared/logupload"
	"xconfwebconfig/util"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	ContextAttributeName   = "contextAttributeName"
	ExpectedValue          = "expectedValue"
	RuleId                 = "ruleId"
	Expires                = "expires"
	TelemetryId            = "telemetryId"
	cTelemetryChannelMapId = "channelMapId"
)

func CreateTelemetryEntryFor(w http.ResponseWriter, r *http.Request) {
	contextAttributeName, found := mux.Vars(r)[ContextAttributeName]
	if !found || contextAttributeName == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing contextAttributeName"))
		return
	}
	expectedValue, found := mux.Vars(r)[ExpectedValue]
	if !found || expectedValue == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing expectedValue"))
		return
	}
	if contextAttributeName != "estbMacAddress" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("only estbMacAddress allowed here"))
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.Error(w, http.StatusInternalServerError, xcommon.NewXconfError(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	telemetryProfile := xwlogupload.TelemetryProfile{}
	err := json.Unmarshal([]byte(body), &telemetryProfile)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	now := time.Now()
	nanos := now.UnixNano()
	millis := nanos / 1000000
	if millis-telemetryProfile.Expires > 0 {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Invalid Expires Timestamp"))
		return
	}
	//timestampedRule := CreateRuleForAttribute(contextAttributeName, expectedValue)
	timestampedRule := CreateTelemetryProfile(contextAttributeName, expectedValue, &telemetryProfile)
	response, err := util.JSONMarshal(timestampedRule)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal timestampedRule error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func DropTelemetryEntryFor(w http.ResponseWriter, r *http.Request) {
	contextAttributeName, found := mux.Vars(r)[ContextAttributeName]
	if !found || contextAttributeName == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing contextAttributeName"))
		return
	}
	expectedValue, found := mux.Vars(r)[ExpectedValue]
	if !found || expectedValue == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing expectedValue"))
		return
	}
	telemetryProfileList := DropTelemetryFor(contextAttributeName, expectedValue)
	response, err := util.JSONMarshal(telemetryProfileList)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal telemetryProfileList error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetDescriptors(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	contextMap := make(map[string]string)
	if len(queryParams) > 0 {
		for k, v := range queryParams {
			contextMap[k] = v[0]
		}
	}
	applicationType, _ := contextMap[xwcommon.APPLICATION_TYPE]
	descriptors := GetAvailableDescriptors(applicationType)
	response, err := util.JSONMarshal(descriptors)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal Descriptors error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetTelemetryDescriptors(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	contextMap := make(map[string]string)
	if len(queryParams) > 0 {
		for k, v := range queryParams {
			contextMap[k] = v[0]
		}
	}
	applicationType, _ := contextMap[xwcommon.APPLICATION_TYPE]
	descriptors := GetAvailableProfileDescriptors(applicationType)
	response, err := util.JSONMarshal(descriptors)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal ProfileDescriptors error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func TempAddToPermanentRule(w http.ResponseWriter, r *http.Request) {
	contextAttributeName, found := mux.Vars(r)[ContextAttributeName]
	if !found || contextAttributeName == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing contextAttributeName"))
		return
	}
	expectedValue, found := mux.Vars(r)[ExpectedValue]
	if !found || expectedValue == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing expectedValue"))
		return
	}
	ruleId, found := mux.Vars(r)[RuleId]
	if !found || ruleId == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing ruleId"))
		return
	}
	expires, found := mux.Vars(r)[Expires]
	if !found || expires == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing expires"))
		return
	}
	if contextAttributeName != "estbMacAddress" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("only estbMacAddress allowed here"))
		return
	}
	expiresInt64, err := strconv.ParseInt(expires, 10, 64)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("expires must be a number"))
		return
	}
	telemetryRule := xlogupload.GetOneTelemetryRule(ruleId) //*TelemetryRule
	if telemetryRule == nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("no rule found for ruleId"))
		return
	}
	profile := xlogupload.GetOnePermanentTelemetryProfile(telemetryRule.BoundTelemetryID) //*PermanentTelemetryProfile
	timedRule := CreateRuleForAttribute(contextAttributeName, expectedValue)
	profile.Expires = expiresInt64
	telemetryRuleBytes, _ := json.Marshal(timedRule)
	xlogupload.SetOneTelemetryProfile(string(telemetryRuleBytes), ConvertPermanentTelemetryProfiletoTelemetryProfile(*profile))

	response, err := util.JSONMarshal(telemetryRule)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal telemetryRule error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func ConvertPermanentTelemetryProfiletoTelemetryProfile(permanentTelemetryProfile xwlogupload.PermanentTelemetryProfile) *xwlogupload.TelemetryProfile {
	telemetryProfile := xwlogupload.TelemetryProfile{}
	telemetryProfile.ID = permanentTelemetryProfile.ID
	telemetryProfile.TelemetryProfile = permanentTelemetryProfile.TelemetryProfile
	telemetryProfile.Schedule = permanentTelemetryProfile.Schedule
	telemetryProfile.Expires = permanentTelemetryProfile.Expires
	telemetryProfile.Name = permanentTelemetryProfile.Name
	telemetryProfile.UploadRepository = permanentTelemetryProfile.UploadRepository
	telemetryProfile.UploadProtocol = permanentTelemetryProfile.UploadProtocol
	telemetryProfile.ApplicationType = permanentTelemetryProfile.ApplicationType
	return &telemetryProfile
}

func BindToTelemetry(w http.ResponseWriter, r *http.Request) {
	contextAttributeName, found := mux.Vars(r)[ContextAttributeName]
	if !found || contextAttributeName == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing contextAttributeName"))
		return
	}
	expectedValue, found := mux.Vars(r)[ExpectedValue]
	if !found || expectedValue == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing expectedValue"))
		return
	}
	telemetryId, found := mux.Vars(r)[TelemetryId]
	if !found || telemetryId == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing telemetryId"))
		return
	}
	expires, found := mux.Vars(r)[Expires]
	if !found || expires == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing expires"))
		return
	}
	if contextAttributeName != "estbMacAddress" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("only estbMacAddress allowed here"))
		return
	}
	expiresInt64, err := strconv.ParseInt(expires, 10, 64)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("expires must be a number"))
		return
	}
	profile := xlogupload.GetOnePermanentTelemetryProfile(telemetryId) //*PermanentTelemetryProfile
	if profile == nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("no rule found for ID "+telemetryId+" provided"))
		return
	}
	timedRule := CreateRuleForAttribute(contextAttributeName, expectedValue)
	profile.Expires = expiresInt64
	telemetryRuleBytes, _ := json.Marshal(timedRule)
	xlogupload.SetOneTelemetryProfile(string(telemetryRuleBytes), ConvertPermanentTelemetryProfiletoTelemetryProfile(*profile))

	response, err := util.JSONMarshal(timedRule)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal telemetryRule error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func TelemetryTestPageHandler(w http.ResponseWriter, r *http.Request) {
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Unable to extract body"))
		return
	}

	body := xw.Body()
	contextMap := make(map[string]string)
	var err error
	if body != "" {
		err = json.Unmarshal([]byte(body), &contextMap)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	xutil.AddQueryParamsToContextMap(r, contextMap)

	if err := shared.NormalizeCommonContext(contextMap, xwcommon.ESTB_MAC_ADDRESS, xwcommon.ECM_MAC_ADDRESS); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	applicationType, err := auth.CanRead(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, err.(xcommon.XconfError).StatusCode, err.Error())
		return
	}

	contextMap[xwcommon.APPLICATION_TYPE] = applicationType

	result := make(map[string]interface{})
	result["context"] = contextMap

	telemetryProfileService := telemetry.NewTelemetryProfileService()
	matchedrule := telemetryProfileService.GetTelemetryRuleForContext(contextMap)
	permanentTelemetryProfile := telemetryProfileService.GetPermanentProfileByTelemetryRule(matchedrule)
	if permanentTelemetryProfile != nil {
		result["result"] = map[string]interface{}{permanentTelemetryProfile.Name: []*xwlogupload.TelemetryRule{matchedrule}}
	} else {
		result["result"] = nil
	}

	res, err := xhttp.ReturnJsonResponse(result, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteResponseBytes(w, res, http.StatusOK, xhttp.ContextTypeHeader(r))
}
