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
package dcm

import (
	"encoding/json"
	"fmt"
	"net/http"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	dcmlogupload "github.com/rdkcentral/xconfwebconfig/dataapi/dcm/logupload"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfwebconfig/common"
	logupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/rdkcentral/xconfwebconfig/dataapi"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	log "github.com/sirupsen/logrus"
)

func DcmTestPageHandler(w http.ResponseWriter, r *http.Request) {
	applicationType, err := auth.CanRead(r, auth.DCM_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	searchContext := make(map[string]string)
	if err := json.Unmarshal([]byte(xw.Body()), &searchContext); err != nil {
		response := "Unable to extract searchContext from json file:" + err.Error()
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(response))
		return
	}

	dataapi.NormalizeCommonContext(searchContext, common.ESTB_MAC_ADDRESS, common.ECM_MAC_ADDRESS)

	searchContext[xwcommon.APPLICATION_TYPE] = applicationType
	fields := log.Fields{}
	logUploadRuleBase := dcmlogupload.NewLogUploadRuleBase()
	eval := logUploadRuleBase.Eval(searchContext, fields)

	allSettings := make(map[string]interface{})
	allSettings["context"] = searchContext
	if eval == nil || eval.RuleIDs == nil || len(eval.RuleIDs) == 0 {
		response, err := util.JSONMarshal(allSettings)
		if err != nil {
			log.Error(fmt.Sprintf("json.Marshal allSettings error: %v", err))
		}
		xhttp.WriteXconfResponse(w, http.StatusOK, response)
		return
	}
	evalResponse := logupload.CreateSettingsResponseObject(eval)
	allSettings["settings"] = evalResponse
	allSettings["matchedRules"] = eval.RuleIDs
	allSettings["ruleType"] = "DCMGenericRule"
	response, err := util.JSONMarshal(allSettings)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal allSettings error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}
