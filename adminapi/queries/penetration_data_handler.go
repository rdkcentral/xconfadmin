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

	xcommon "xconfadmin/common"
	"xconfwebconfig/db"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"

	"github.com/gorilla/mux"
)

func GetPenetrationMetricsByEstbMac(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	macAddress := mux.Vars(r)[xcommon.MAC_ADDRESS]
	pr, err := db.GetDatabaseClient().GetPenetrationMetrics(macAddress)
	if err != nil {
		errorStr := fmt.Sprintf("%v not found", macAddress)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	res, err := json.Marshal(pr)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, res)
}
