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
	"fmt"
	"net/http"
	"net/url"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/util"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfwebconfig/common"
	estbfirmware "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	log "github.com/sirupsen/logrus"
)

func GetLogs(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	macStr, found := mux.Vars(r)["macStr"]
	if !found || macStr == "" {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("missing macStr"))
		return
	}
	macAddress, err := util.ValidateAndNormalizeMacAddress(macStr)
	if err != nil {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("invalid mac address: "+macStr))
		return
	}

	result := make(map[string]interface{}, 2)
	tenantId := xhttp.GetTenantId(r)
	last := estbfirmware.GetLastConfigLog(tenantId, macAddress) //*ConfigChangeLog
	if last != nil {
		configChangeLogList := estbfirmware.GetConfigChangeLogsOnly(tenantId, macAddress) //[]*ConfigChangeLog
		result["lastConfigLog"] = last
		result["configChangeLog"] = configChangeLogList
	}
	response, err := util.JSONMarshal(result)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal result error: %v", err))
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetEstbLastlogPath(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	isValid, mac, errStr := isMacPresentAndValid(r.URL.Query())
	if !isValid {
		xhttp.WriteXconfResponseAsText(w, 400, []byte(errStr))
	} else {
		mac := util.NormalizeMacAddress(mac)
		tenantId := xhttp.GetTenantId(r)
		lastConfigLog := estbfirmware.GetLastConfigLog(tenantId, mac)
		if lastConfigLog != nil {
			logPreDisplayCleanup(lastConfigLog)
			response, _ := util.JSONMarshal(*lastConfigLog)
			xhttp.WriteXconfResponse(w, 200, response)
		} else {
			log.Debugf("Last log is not found for mac %s", mac)
			xhttp.WriteXconfResponse(w, 200, []byte(""))
		}
	}
}

func GetEstbChangelogsPath(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	isValid, mac, errStr := isMacPresentAndValid(r.URL.Query())
	if !isValid {
		xhttp.WriteXconfResponseAsText(w, 400, []byte(errStr))
	} else {
		mac := util.NormalizeMacAddress(mac)
		tenantId := xhttp.GetTenantId(r)
		configChangeLogs := estbfirmware.GetConfigChangeLogsOnly(tenantId, mac)
		if len(configChangeLogs) > 0 {
			for _, log := range configChangeLogs {
				logPreDisplayCleanup(log)
			}
		} else {
			log.Debugf("Last log is not found for mac %s", mac)
		}
		response, _ := util.JSONMarshal(configChangeLogs)
		xhttp.WriteXconfResponse(w, 200, response)
	}
}

func logPreDisplayCleanup(lastConfigLog *estbfirmware.ConfigChangeLog) {
	if lastConfigLog != nil {
		lastConfigLog.ID = ""
		lastConfigLog.Updated = 0
	}
}

func isMacPresentAndValid(queryParams url.Values) (bool, string, string) {
	var mac string
	var errorStr string
	if len(queryParams) > 0 {
		for k, v := range queryParams {
			if k == common.MAC {
				mac = v[0]
			}
		}
	}
	if mac == "" {
		errorStr = fmt.Sprintf("Required String parameter '%s' is not present", common.MAC)
		return false, mac, errorStr
	}
	if !util.IsValidMacAddress(mac) {
		errorStr = fmt.Sprintf("Mac is invalid: %s", mac)
		return false, mac, errorStr
	}
	return true, mac, errorStr
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
