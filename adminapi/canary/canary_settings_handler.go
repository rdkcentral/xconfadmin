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

package canary

import (
	"encoding/json"
	"net/http"

	"xconfadmin/adminapi/auth"
	ccommon "xconfadmin/common"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"
)

func PutCanarySettingsHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.HasWritePermissionForTool(r) {
		xhttp.WriteAdminErrorResponse(w, http.StatusForbidden, "No write permission: tools")
	}

	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "responsewriter cast error")
		return
	}
	body := xw.Body()
	canarySettings := ccommon.CanarySettings{}
	err := json.Unmarshal([]byte(body), &canarySettings)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	respEntity := SetCanarySetting(&canarySettings)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}
	xhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

func GetCanarySettingsHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.HasReadPermissionForTool(r) {
		xhttp.WriteAdminErrorResponse(w, http.StatusUnauthorized, "")
		return
	}

	canarySetting, err := GetCanarySettings()
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := xhttp.ReturnJsonResponse(canarySetting, r)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, res)
}
