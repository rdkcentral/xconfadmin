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
package shared

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	xcommon "xconfadmin/common"
	xutil "xconfadmin/util"
	xwcommon "xconfwebconfig/common"
	"xconfwebconfig/db"
	"xconfwebconfig/shared"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

func GetApplicationFromCookies(r *http.Request) string {
	cookie, err := r.Cookie(xwcommon.APPLICATION_TYPE)
	if err != nil || util.IsBlank(cookie.Value) {
		return ""
	}
	return cookie.Value
}

func GetAllModelList() []*shared.Model {
	result := []*shared.Model{}
	list, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_MODEL, 0)
	if err != nil {
		log.Warn("no model found")
		return result
	}
	for _, inst := range list {
		model := inst.(*shared.Model)
		result = append(result, model)
	}
	return result
}

func ApplicationTypeEquals(type1 string, type2 string) bool {
	if type1 == "" {
		type1 = shared.STB
	}
	if type2 == "" {
		type2 = shared.STB
	}
	return type1 == type2
}

func NormalizeCommonContext(contextMap map[string]string, estbMacKey string, ecmMacKey string) (e error) {
	if model := contextMap[xwcommon.MODEL]; model != "" {
		contextMap[xwcommon.MODEL] = strings.ToUpper(model)
	}
	if env := contextMap[xwcommon.ENV]; env != "" {
		contextMap[xwcommon.ENV] = strings.ToUpper(env)
	}
	if partnerId := contextMap[xwcommon.PARTNER_ID]; partnerId != "" {
		contextMap[xwcommon.PARTNER_ID] = strings.ToUpper(partnerId)
	}
	if mac := contextMap[estbMacKey]; mac != "" {
		if normalizedMac, err := xutil.ValidateAndNormalizeMacAddress(mac); err != nil {
			e = err
		} else {
			contextMap[estbMacKey] = normalizedMac
		}
	}
	if mac := contextMap[ecmMacKey]; mac != "" {
		if normalizedMac, err := xutil.ValidateAndNormalizeMacAddress(mac); err != nil {
			e = err
		} else {
			contextMap[ecmMacKey] = normalizedMac
		}
	}
	return e
}

func IsValidApplicationType(at string) bool {
	if at == shared.STB || at == shared.RDKCLOUD {
		return true
	}
	return false
}

// Validate whether the ApplicationType is valid if specified
func ValidateApplicationType(applicationType string) error {
	if applicationType == "" {
		return xcommon.NewXconfError(http.StatusBadRequest, "ApplicationType is empty")
	}
	if !IsValidApplicationType(applicationType) {
		return xcommon.NewXconfError(http.StatusBadRequest, fmt.Sprintf("ApplicationType %s is not valid", applicationType))
	}
	return nil
}

func GetAppSettings() (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	list, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_APP_SETTINGS, 0)
	if err != nil {
		return settings, err
	}
	for _, v := range list {
		p := *v.(*shared.AppSetting)
		settings[p.ID] = p.Value
	}
	return settings, nil
}

func SetAppSetting(key string, value interface{}) (*shared.AppSetting, error) {
	setting := shared.AppSetting{
		ID:      key,
		Updated: util.GetTimestamp(time.Now().UTC()),
		Value:   value,
	}

	err := db.GetCachedSimpleDao().SetOne(db.TABLE_APP_SETTINGS, setting.ID, &setting)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}
