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
package common

import (
	"fmt"
	"time"
	"xconfwebconfig/db"
	ds "xconfwebconfig/db"
	"xconfwebconfig/shared"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

// http error response to match xconf java admin
type HttpAdminErrorResponse struct {
	Status  int    `json:"status"`
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
}

func GetBooleanAppSetting(key string, vargs ...bool) bool {
	defaultVal := false
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := ds.GetCachedSimpleDao().GetOne(TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting := inst.(*shared.AppSetting)
	return setting.Value.(bool)
}

type MacIpRuleConfig struct {
	IpMacIsConditionLimit int `json:"ipMacIsConditionLimit"`
}

// http error response
type HttpErrorResponse struct {
	Status    int         `json:"status"`
	ErrorCode int         `json:"error_code,omitempty"`
	Message   string      `json:"message,omitempty"`
	Errors    interface{} `json:"errors,omitempty"`
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

func IsValidAppSetting(key string) bool {
	return util.Contains(AllAppSettings, key)
}
