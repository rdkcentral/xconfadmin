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
package firmware

import (
	"fmt"
	"net/http"

	xutil "xconfadmin/util"

	ef "github.com/rdkcentral/xconfwebconfig/dataapi/estbfirmware"

	xcommon "xconfadmin/common"
	xshared "xconfadmin/shared"

	xwshared "github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	log "github.com/sirupsen/logrus"
)

type ValueValidator func(string) bool

func writeErrorResponse(w http.ResponseWriter, r *http.Request, errorMsg string, status int, errorType string) {
	errResult := xcommon.HttpErrorResponse{
		Status:  status,
		Message: errorMsg,
		Errors:  errorType,
	}
	response, err := xhttp.ReturnJsonResponse(errResult, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, status, response)
}

func GetFirmwareTestPageHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the search parameters from query params
	context := make(map[string]string)
	xutil.AddQueryParamsToContextMap(r, context)

	if err := xshared.NormalizeCommonContext(context, xwcommon.ESTB_MAC, xwcommon.ECM_MAC); err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// If input has any of these search-paramters, validate their values
	searchValidators := map[string]ValueValidator{
		xwcommon.ENV: func(id string) bool {
			return id != "" && xwshared.GetOneEnvironment(id) != nil
		},
		xwcommon.MODEL: func(id string) bool {
			return id != "" && xwshared.GetOneModel(id) != nil
		},
		xwcommon.IP_ADDRESS: func(val string) bool {
			return xwshared.NewIpAddress(val) != nil
		},
		xwcommon.ESTB_MAC: func(val string) bool {
			ok, _ := util.MACAddressValidator(val)
			return ok
		},
	}
	for k, v := range context {
		if validator, ok := searchValidators[k]; ok {
			if validator != nil && !validator(v) {
				errMsg := "Invalid Value '" + v + "' for " + k
				log.Error(errMsg)
				writeErrorResponse(w, r, errMsg, http.StatusBadRequest, "IllegalArgumentException")
				return
			}
		}
	}

	if _, ok := context[xwcommon.ESTB_MAC]; !ok {
		writeErrorResponse(w, r, xwcommon.ESTB_MAC+" cannot be empty", http.StatusBadRequest, "IllegalArgumentException")
		return
	}
	if _, ok := context[xwcommon.TIME]; !ok {
		context[xwcommon.TIME] = util.UtcCurrentTimestamp().String()
	}

	if _, ok := context[xwcommon.IP_ADDRESS]; !ok {
		context[xwcommon.IP_ADDRESS] = "1.1.1.1"
	}

	// Construct ruleBase
	ruleBase := ef.NewEstbFirmwareRuleBaseDefault()
	convertedContext := coreef.GetContextConverted(context)

	// Evaluate rule
	applicationType, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	eval, err := ruleBase.Eval(context, convertedContext, applicationType, log.Fields{})
	if err != nil {
		errMsg := fmt.Sprintf("Rule Evaluation Error: %v", err)
		writeErrorResponse(w, r, errMsg, http.StatusBadRequest, "IllegalArgumentException")
		log.Error(errMsg)
		return
	}

	resultMap := make(map[string]interface{})
	resultMap["context"] = context
	resultMap["result"] = eval
	response, err := util.JSONMarshal(resultMap)
	if err != nil {
		errMsg := fmt.Sprintf("json.Marshal resultMap error: %v", err)
		writeErrorResponse(w, r, errMsg, http.StatusBadRequest, "IllegalArgumentException")
		log.Error(errMsg)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}
