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

	"net/http"

	"xconfwebconfig/db"
	"xconfwebconfig/shared"
	corefw "xconfwebconfig/shared/firmware"

	re "xconfwebconfig/rulesengine"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	xwhttp "xconfwebconfig/http"
)

func PostFirmwareRuleReportPageHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.FIRMWARE_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, "Unable to extract ResponseWriter")
		return
	}
	body := xw.Body()
	macRuleIds := []string{}
	err = json.Unmarshal([]byte(body), &macRuleIds)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	header := make(map[string]string)
	header["Content-Disposition"] = "attachment; filename=filename=report.xls"
	header["Content-Type"] = "application/vnd.ms-excel"

	macRules, _ := db.GetSimpleDao().GetAllByKeys(db.TABLE_FIRMWARE_RULE, macRuleIds)

	macIds := getMacAddresses(macRules)
	reportBytes, err := doReport(macIds)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	_, err = w.Write(reportBytes)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xwhttp.WriteXconfResponseWithHeaders(w, header, http.StatusOK, nil)
}

func getMacAddresses(macRuleIds []interface{}) []string {
	resultMap := make(map[string]bool)
	for _, genrule := range macRuleIds {
		rule := genrule.(*corefw.FirmwareRule)
		if rule.GetRule() != nil {
			macListIds := re.GetFixedArgsFromRuleByFreeArgAndOperation(*rule.GetRule(), "eStbMac", re.StandardOperationInList)
			for _, macListId := range macListIds {
				macList, err := shared.GetGenericNamedListOneDB(macListId)
				if err == nil {
					for _, item := range macList.Data {
						resultMap[item] = true
					}
				}
			}
			macListIds = re.GetFixedArgsFromRuleByFreeArgAndOperation(*rule.GetRule(), "eStbMac", re.StandardOperationIs)
			for _, macListId := range macListIds {
				resultMap[macListId] = true
			}
		}
	}

	result := []string{}
	for k := range resultMap {
		result = append(result, k)
	}
	return result
}
