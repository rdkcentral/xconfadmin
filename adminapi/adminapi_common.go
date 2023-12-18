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
package adminapi

import (
	"xconfwebconfig/dataapi"

	xcommon "xconfadmin/common"
	xwcommon "xconfwebconfig/common"

	queries "xconfadmin/adminapi/queries"
	xhttp "xconfadmin/http"
	xshared "xconfadmin/shared"
)

// Ws - webserver object
var (
	Ws *xhttp.WebconfigServer
	Xc *dataapi.XconfConfigs
)

// WebServerInjection - dependency injection
func WebServerInjection(ws *xhttp.WebconfigServer, xc *dataapi.XconfConfigs) {
	Ws = ws
	if ws == nil {
		xwcommon.CacheUpdateWindowSize = 60000
		xcommon.AllowedNumberOfFeatures = 100
		xcommon.ActiveAuthProfiles = "dev"
		xcommon.DefaultAuthProfiles = "prod"
		xcommon.SatOn = true
		xcommon.IpMacIsConditionLimit = 20
	} else {
		xwcommon.CacheUpdateWindowSize = ws.XW_XconfServer.ServerConfig.GetInt64("xconfwebconfig.xconf.cache_update_window_size")
		xcommon.AllowedNumberOfFeatures = int(ws.XW_XconfServer.ServerConfig.GetInt32("xconfwebconfig.xconf.allowedNumberOfFeatures", 100))
		xcommon.ActiveAuthProfiles = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.authProfilesActive")
		xcommon.DefaultAuthProfiles = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.authProfilesDefault")
		xcommon.SatOn = ws.XW_XconfServer.ServerConfig.GetBoolean("xconfwebconfig.sat.SAT_ON")
		xcommon.IpMacIsConditionLimit = int(ws.XW_XconfServer.ServerConfig.GetInt32("xconfwebconfig.xconf.ipMacIsConditionLimit", 20))
	}
	if ws.TestOnly() {
		xcommon.SatOn = false
	}
	Xc = xc
}

func initDB() {
	queries.CreateFirmwareRuleTemplates() // Initialize FirmwareRule templates
	initAppSettings()                     // Initialize Application settings
}

func initAppSettings() {
	settings, err := xshared.GetAppSettings()
	if err != nil {
		panic(err)
	}
	if len(settings) == 0 {
		if _, ok := settings[xcommon.READONLY_MODE]; !ok {
			xshared.SetAppSetting(xcommon.READONLY_MODE, false)
		}
	}
}
