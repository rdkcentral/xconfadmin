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
package adminapi

import (
	"strings"
	"time"

	"github.com/rdkcentral/xconfwebconfig/dataapi"

	queries "github.com/rdkcentral/xconfadmin/adminapi/queries"
	common "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	xapptype "github.com/rdkcentral/xconfadmin/shared/applicationtype"

	log "github.com/sirupsen/logrus"
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
		common.CacheUpdateWindowSize = 60000
		common.AllowedNumberOfFeatures = 100
		common.ActiveAuthProfiles = "dev"
		common.DefaultAuthProfiles = "prod"
		common.IpMacIsConditionLimit = 20
		common.CanaryCreationEnabled = false
		common.VideoCanaryCreationEnabled = false
		common.AuthProvider = "acl"
		common.ApplicationTypes = []string{"stb"}
		common.WakeupPoolTagName = "t_canary_wakeup"
	} else {
		common.AuthProvider = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.authprovider")
		if len(common.ApplicationTypes) == 0 {
			common.ApplicationTypes = []string{"stb"}
		}
		common.CacheUpdateWindowSize = ws.XW_XconfServer.ServerConfig.GetInt64("xconfwebconfig.xconf.cache_update_window_size")
		common.SatOn = ws.XW_XconfServer.ServerConfig.GetBoolean("xconfwebconfig.sat.SAT_ON")
		common.AllowedNumberOfFeatures = int(ws.XW_XconfServer.ServerConfig.GetInt32("xconfwebconfig.xconf.allowedNumberOfFeatures", 100))
		common.ActiveAuthProfiles = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.authProfilesActive")
		common.DefaultAuthProfiles = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.authProfilesDefault")
		common.IpMacIsConditionLimit = int(ws.XW_XconfServer.ServerConfig.GetInt32("xconfwebconfig.xconf.ipMacIsConditionLimit", 20))
		common.CanaryCreationEnabled = ws.XW_XconfServer.ServerConfig.GetBoolean("xconfwebconfig.xconf.enable_canary_creation")
		common.VideoCanaryCreationEnabled = ws.XW_XconfServer.ServerConfig.GetBoolean("xconfwebconfig.xconf.enable_video_canary_creation")
		common.LockDuration = ws.XW_XconfServer.ServerConfig.GetInt32("xconfwebconfig.xcrp.lock_duration_in_secs", common.DefaultLockDuration)
		if common.CanaryCreationEnabled {
			timezoneStr := ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_time_zone")
			timezone, err := time.LoadLocation(timezoneStr)
			if err != nil {
				log.Errorf("Error loading timezone: %s", timezoneStr)
				panic(err)
			}
			common.CanaryTimezone = timezone
			timezoneListString := ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_timezone_list")
			common.CanaryTimezoneList = strings.Split(timezoneListString, ",")
			common.CanaryStartTime = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_start_time")
			common.CanaryEndTime = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_end_time")
			common.CanaryTimeFormat = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_time_format")
			common.CanaryDefaultPartner = strings.ToLower(ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_default_partner"))
			common.CanarySize = int(ws.XW_XconfServer.ServerConfig.GetInt64("xconfwebconfig.xconf.canary_size"))
			common.CanaryDistributionPercentage = ws.XW_XconfServer.ServerConfig.GetFloat64("xconfwebconfig.xconf.canary_distribution_percentage")
			common.CanaryFwUpgradeStartTime = int(ws.XW_XconfServer.ServerConfig.GetInt64("xconfwebconfig.xconf.canary_firmware_upgrade_start_time"))
			common.CanaryFwUpgradeEndTime = int(ws.XW_XconfServer.ServerConfig.GetInt64("xconfwebconfig.xconf.canary_firmware_upgrade_end_time"))

			percentFilterNameString := strings.ToLower(ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_percent_filter_name"))
			for _, name := range strings.Split(percentFilterNameString, ";") {
				common.CanaryPercentFilterNameSet.Add(name)
			}

			wakeupPercentFilterNameString := strings.ToLower(ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_wakeup_percent_filter_list"))
			for _, name := range strings.Split(wakeupPercentFilterNameString, ",") {
				common.CanaryWakeupPercentFilterNameSet.Add(name)
			}

			videoModelListString := strings.ToUpper(ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_video_model_list"))
			for _, model := range strings.Split(videoModelListString, ",") {
				common.CanaryVideoModelSet.Add(model)
			}

			syndicatePartnerList := strings.ToLower(ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_appsettings_partner_list"))
			for _, name := range strings.Split(syndicatePartnerList, ",") {
				common.CanarySyndicatePartnerSet.Add(name)
				common.AllAppSettings = append(common.AllAppSettings, (common.PROP_CANARY_FW_UPGRADE_STARTTIME + "_" + name))
				common.AllAppSettings = append(common.AllAppSettings, (common.PROP_CANARY_FW_UPGRADE_ENDTIME + "_" + name))
				common.AllAppSettings = append(common.AllAppSettings, (common.PROP_CANARY_TIMEZONE_LIST + "_" + name))
			}
		}
	}
	Xc = xc
}

func initDB() {
	queries.CreateFirmwareRuleTemplates() // Initialize FirmwareRule templates
	initAppSettings()                     // Initialize Application settings
	loadApplicationTypes()                // Load Application Types from DB
}

func loadApplicationTypes() {
	appTypes, err := xapptype.GetAllApplicationTypeAsList()
	if err != nil {
		log.Errorf("Error loading application types from DB: %v", err)
		return
	}
	var at []string
	for _, appName := range appTypes {
		at = append(at, appName.Name)
	}
	common.ApplicationTypes = at
	log.Info("Loaded application types from DB")
}
func initAppSettings() {
	settings, err := common.GetAppSettings()
	if err != nil {
		panic(err)
	}
	if _, ok := settings[common.PROP_LOCKDOWN_ENABLED]; !ok {
		common.SetAppSetting(common.PROP_LOCKDOWN_ENABLED, false)
	}
	if _, ok := settings[common.PROP_CANARY_MAXSIZE]; !ok {
		common.SetAppSetting(common.PROP_CANARY_MAXSIZE, common.CanarySize)
	}
	if _, ok := settings[common.PROP_CANARY_DISTRIBUTION_PERCENTAGE]; !ok {
		common.SetAppSetting(common.PROP_CANARY_DISTRIBUTION_PERCENTAGE, common.CanaryDistributionPercentage)
	}
	if _, ok := settings[common.PROP_CANARY_FW_UPGRADE_STARTTIME]; !ok {
		common.SetAppSetting(common.PROP_CANARY_FW_UPGRADE_STARTTIME, common.CanaryFwUpgradeStartTime)
	}
	if _, ok := settings[common.PROP_CANARY_FW_UPGRADE_ENDTIME]; !ok {
		common.SetAppSetting(common.PROP_CANARY_FW_UPGRADE_ENDTIME, common.CanaryFwUpgradeEndTime)
	}

	if _, ok := settings[common.PROP_LOCKDOWN_STARTTIME]; !ok {
		common.SetAppSetting(common.PROP_LOCKDOWN_STARTTIME, common.DefaultLockdownStartTime)
	}

	if _, ok := settings[common.PROP_LOCKDOWN_ENDTIME]; !ok {
		common.SetAppSetting(common.PROP_LOCKDOWN_ENDTIME, common.DefaultLockdownEndTime)
	}

	if _, ok := settings[common.PROP_LOCKDOWN_MODULES]; !ok {
		common.SetAppSetting(common.PROP_LOCKDOWN_MODULES, common.DefaultLockdownModules)
	}

	if _, ok := settings[common.PROP_PRECOOK_LOCKDOWN_ENABLED]; !ok {
		common.SetAppSetting(common.PROP_LOCKDOWN_ENABLED, common.DefaultPrecookLockdownEnabled)
	}

	if _, ok := settings[common.PROP_CANARY_TIMEZONE_LIST]; !ok {
		common.SetAppSetting(common.PROP_CANARY_TIMEZONE_LIST, common.DefaultCanaryTimezone)
	}

}
