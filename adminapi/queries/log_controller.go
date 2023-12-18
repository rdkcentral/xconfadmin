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
	"fmt"
	"net/http"

	"xconfadmin/util"
	"xconfwebconfig/shared/estbfirmware"
	xwutil "xconfwebconfig/util"

	xwhttp "xconfwebconfig/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func GetLogs(w http.ResponseWriter, r *http.Request) {
	macStr, found := mux.Vars(r)["macStr"]
	if !found || macStr == "" {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing macStr"))
		return
	}
	macAddress, err := util.ValidateAndNormalizeMacAddress(macStr)
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("invalid mac address: "+macStr))
		return
	}
	result := make(map[string]interface{}, 2)
	last := estbfirmware.GetLastConfigLog(macAddress) //*ConfigChangeLog
	if last != nil {
		configChangeLogList := estbfirmware.GetConfigChangeLogsOnly(macAddress) //[]*ConfigChangeLog
		result["lastConfigLog"] = last
		result["configChangeLog"] = configChangeLogList
	}
	response, err := xwutil.JSONMarshal(result)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal result error: %v", err))
	}
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func getOneConfigChangeLog(macAddress string) *estbfirmware.ConfigChangeLog {
	if macAddress == "" {
		return nil
	}
	configChangeLog1 := estbfirmware.ConfigChangeLog{}
	configChangeLog1.ID = "id1"
	configChangeLog1.Updated = 1636566496
	configChangeLog1.Explanation = "explanation"
	configChangeLog1.HasMinimumFirmware = true
	return &configChangeLog1
}

func getConfigChangeLogList(macAddress string) []*estbfirmware.ConfigChangeLog {
	if macAddress == "" {
		return nil
	}
	list := []*estbfirmware.ConfigChangeLog{}
	configChangeLog1 := estbfirmware.ConfigChangeLog{}
	configChangeLog1.ID = "id1"
	configChangeLog1.Updated = 1636566496
	configChangeLog1.Explanation = "explanation"
	configChangeLog1.HasMinimumFirmware = true
	configChangeLog2 := estbfirmware.ConfigChangeLog{}
	configChangeLog2.ID = "id2"
	configChangeLog2.Updated = 1636566498
	configChangeLog2.Explanation = "explanation"
	configChangeLog2.HasMinimumFirmware = true
	list = append(list, &configChangeLog1)
	list = append(list, &configChangeLog2)
	return list
}
