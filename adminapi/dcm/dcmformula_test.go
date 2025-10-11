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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"

	// Don't import adminapi to avoid circular dependency
	// "github.com/rdkcentral/xconfadmin/adminapi"
	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	queries "github.com/rdkcentral/xconfadmin/adminapi/queries"
	"github.com/rdkcentral/xconfadmin/common"
	oshttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/taggingapi"
	"github.com/rdkcentral/xconfadmin/taggingapi/tag"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	core "github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"
	"gotest.tools/assert"
)

var (
	testConfigFile     string
	jsonTestConfigFile string
	sc                 *xwcommon.ServerConfig
	server             *oshttp.WebconfigServer
	router             *mux.Router
	globAut            *apiUnitTest
)

func Walk(r *mux.Router) {
	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})
	if err != nil {
		panic(err)
	}
}

var jsondfCreateData = []byte(
	`{
   "negated":false,
   "condition":{
      "freeArg":{
         "type":"STRING",
         "name":"estbIP"
      },
      "operation":"IS",
      "fixedArg":{
         "bean":{
            "value":{
               "java.lang.String":"3.3.3.3"
            }
         }
      }
   },
   "compoundParts":[
      
   ],
   "id":"33af3261-d74a-40fd-8aa1-884e4f5479a1",
   "name":"dineshtest3",
   "priority":1,
   "percentage":100,
   "percentageL1":10,
   "percentageL2":10,
   "percentageL3":80,
   "applicationType":"stb"
}`)

var jsondfPostCreateData = []byte(
	`{
   "negated":false,
   "compoundParts":[
      {
         "negated":false,
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbIP"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14.14.14.14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      },
      {
         "negated":false,
         "relation":"OR",
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbMacAddress"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14:14:14:14:14:14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      }
   ],
   "id":"3f81ab29-ab8e-40d5-b407-cbc579b46caa",
   "name":"dinesh14",
   "priority":2,
   "percentage":100,
   "percentageL1":10,
   "percentageL2":10,
   "percentageL3":80,
   "applicationType":"stb"
}`)

var jsondfUpdateData = []byte(
	`{
   "negated":false,
   "compoundParts":[
      {
         "negated":false,
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbIP"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14.14.14.14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      },
      {
         "negated":false,
         "relation":"OR",
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbMacAddress"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14:14:14:14:14:14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      }
   ],
   "id":"3f81ab29-ab8e-40d5-b407-cbc579b46caa",
   "name":"dinesh14update",
   "priority":3,
   "percentage":100,
   "percentageL1":20,
   "percentageL2":20,
   "percentageL3":60,
   "applicationType":"stb"
}`)

var jsondfUpdateErrData = []byte(
	`{
   "negated":false,
   "compoundParts":[
      {
         "negated":false,
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbIP"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14.14.14.14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      },
      {
         "negated":false,
         "relation":"OR",
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbMacAddress"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14:14:14:14:14:14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      }
   ],
   "id":"3f81ab29-ab8e-40d5-b407-cbc579b46caaer",
   "name":"dinesh14update",
   "priority":3,
   "percentage":100,
   "percentageL1":20,
   "percentageL2":20,
   "percentageL3":60,
   "applicationType":"stb"
}`)

var payload = []byte(`["3f81ab29-ab8e-40d5-b407-cbc579b46caa"]`)
var postmapname = []byte(`{"NAME": "din"}`)
var postmapIPargs = []byte(`{"FIXED_ARG": "3","FREE_ARG": "IP"}`)
var postmapMACargs = []byte(`{"FIXED_ARG": "14","FREE_ARG": "MAC"}`)

const (
	DF_URL = "/xconfAdminService/dcm/formula"
)

// WebServerInjection - local implementation to avoid circular dependency
func WebServerInjection(ws *oshttp.WebconfigServer, xc *dataapi.XconfConfigs) {
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
		applicationTypeString := ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.application_types")
		if applicationTypeString == "" {
			applicationTypeString = "stb"
		}
		common.ApplicationTypes = strings.Split(applicationTypeString, ",")
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
}

// initDB - local implementation to avoid circular dependency
func initDB() {
	CreateFirmwareRuleTemplates() // Initialize FirmwareRule templates
	initAppSettings()             // Initialize Application settings
}

// CreateFirmwareRuleTemplates - local implementation to avoid circular dependency
func CreateFirmwareRuleTemplates() {
	if count, _ := GetFirmwareRuleTemplateCount(); count > 0 {
		return
	}

	log.Info("Creating templates...")

	ruleFactory := coreef.NewRuleFactory()
	templateList := []corefw.FirmwareRuleTemplate{}

	// Rule actions
	rule := coreef.NewMacRule(coreef.EMPTY_NAME)
	templateList = append(templateList, *NewFirmwareRuleTemplate(
		corefw.MAC_RULE, rule, coreef.EMPTY_LIST, 1))

	rule = ruleFactory.NewIpRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	templateList = append(templateList, *NewFirmwareRuleTemplate(
		corefw.IP_RULE, rule, coreef.EMPTY_LIST, 2))

	rule = ruleFactory.NewIntermediateVersionRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	templateList = append(templateList, *NewFirmwareRuleTemplate(
		corefw.IV_RULE, rule, []string{corefw.GLOBAL_PERCENT, corefw.TIME_FILTER}, 3))

	rule = ruleFactory.NewMinVersionCheckRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_LIST)
	templateList = append(templateList, *NewFirmwareRuleTemplate(
		corefw.MIN_CHECK_RULE, rule, []string{corefw.GLOBAL_PERCENT, corefw.TIME_FILTER}, 4))

	rule = ruleFactory.NewEnvModelRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	templ := *NewFirmwareRuleTemplate(corefw.ENV_MODEL_RULE, rule, []string{}, 5)
	templ.Editable = false
	templateList = append(templateList, templ)

	// Blocking filters
	rule = *ruleFactory.NewGlobalPercentFilterTemplate(coreef.DEFAULT_PERCENT, coreef.EMPTY_NAME)
	templ = *NewBlockingFilterTemplate(corefw.GLOBAL_PERCENT, rule, 1)
	templateList = append(templateList, templ)

	rule = *ruleFactory.NewIpFilter(coreef.EMPTY_NAME)
	templateList = append(templateList, *NewBlockingFilterTemplate(
		corefw.IP_FILTER, rule, 2))

	rule = *ruleFactory.NewTimeFilterTemplate(true, true, false, coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_NAME, "01:00", "02:00")
	templateList = append(templateList, *NewBlockingFilterTemplate(
		corefw.TIME_FILTER, rule, 3))

	// Define Properties
	rule = *ruleFactory.NewDownloadLocationFilter(coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	properties := map[string]corefw.PropertyValue{
		coreef.FIRMWARE_DOWNLOAD_PROTOCOL: *corefw.NewPropertyValue("tftp", false, corefw.STRING),
		coreef.FIRMWARE_LOCATION:          *corefw.NewPropertyValue("", false, corefw.STRING),
		coreef.IPV6_FIRMWARE_LOCATION:     *corefw.NewPropertyValue("", true, corefw.STRING),
	}
	templateList = append(templateList, *NewDefinePropertiesTemplate(
		corefw.DOWNLOAD_LOCATION_FILTER, rule, properties, coreef.EMPTY_LIST, 3))

	rule = *ruleFactory.NewRiFilterTemplate()
	properties = map[string]corefw.PropertyValue{
		coreef.REBOOT_IMMEDIATELY: *corefw.NewPropertyValue("true", false, corefw.BOOLEAN),
	}
	templateList = append(templateList, *NewDefinePropertiesTemplate(
		corefw.REBOOT_IMMEDIATELY_FILTER, rule, properties, coreef.EMPTY_LIST, 1))

	rule = ruleFactory.NewMinVersionCheckRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_LIST)
	properties = map[string]corefw.PropertyValue{
		coreef.REBOOT_IMMEDIATELY: *corefw.NewPropertyValue("true", true, corefw.BOOLEAN),
	}
	templateList = append(templateList, *NewDefinePropertiesTemplate(
		corefw.MIN_CHECK_RI, rule, properties, []string{corefw.GLOBAL_PERCENT, corefw.TIME_FILTER}, 2))

	rule = ruleFactory.NewActivationVersionRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	properties = map[string]corefw.PropertyValue{
		coreef.REBOOT_IMMEDIATELY: *corefw.NewPropertyValue("false", false, corefw.BOOLEAN),
	}
	templ = *NewDefinePropertiesTemplate(
		corefw.ACTIVATION_VERSION, rule, properties, coreef.EMPTY_LIST, 4)
	templ.Editable = false
	templateList = append(templateList, templ)

	for _, template := range templateList {
		if err := template.Validate(); err != nil {
			panic(err)
		}
		template.Updated = util.GetTimestamp()
		if jsonData, err := json.Marshal(template); err != nil {
			panic(err)
		} else {
			if err := ds.GetSimpleDao().SetOne(ds.TABLE_FIRMWARE_RULE_TEMPLATE, template.ID, jsonData); err != nil {
				panic(err)
			}
		}
	}
}

// GetFirmwareRuleTemplateCount - local implementation to avoid circular dependency
func GetFirmwareRuleTemplateCount() (int, error) {
	entries, err := db.GetSimpleDao().GetAllAsMapRaw(db.TABLE_FIRMWARE_RULE_TEMPLATE, 0)
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareRuleTemplateCount: %v", err))
		return 0, err
	}
	return len(entries), nil
}

// Helper functions to create templates - local implementations to avoid circular dependency
func NewFirmwareRuleTemplate(id string, rule rulesengine.Rule, byPassFilters []string, priority int) *corefw.FirmwareRuleTemplate {
	action := corefw.NewTemplateApplicableActionAndType(corefw.RuleActionClass, corefw.RULE_TEMPLATE, "")
	return &corefw.FirmwareRuleTemplate{
		ID:               id,
		Priority:         int32(priority),
		Rule:             rule,
		ApplicableAction: action,
		Editable:         true,
		RequiredFields:   []string{},
		ByPassFilters:    byPassFilters,
	}
}

func NewBlockingFilterTemplate(id string, rule rulesengine.Rule, priority int) *corefw.FirmwareRuleTemplate {
	action := corefw.NewTemplateApplicableActionAndType(corefw.BlockingFilterActionClass, corefw.BLOCKING_FILTER_TEMPLATE, "")
	return &corefw.FirmwareRuleTemplate{
		ID:               id,
		Priority:         int32(priority),
		Rule:             rule,
		ApplicableAction: action,
		Editable:         true,
		RequiredFields:   []string{},
		ByPassFilters:    []string{},
	}
}

func NewDefinePropertiesTemplate(id string, rule rulesengine.Rule, properties map[string]corefw.PropertyValue, byPassFilter []string, priority int) *corefw.FirmwareRuleTemplate {
	action := corefw.NewTemplateApplicableActionAndType(corefw.DefinePropertiesActionClass, corefw.DEFINE_PROPERTIES_TEMPLATE, "")
	action.Properties = properties
	return &corefw.FirmwareRuleTemplate{
		ID:               id,
		Priority:         int32(priority),
		Rule:             rule,
		ApplicableAction: action,
		Editable:         true,
		RequiredFields:   []string{},
		ByPassFilters:    byPassFilter,
	}
}

// initAppSettings - local implementation to avoid circular dependency
func initAppSettings() {
	// Initialize application settings if needed
	// This is a simplified version for testing purposes
}

func dcmSetup(server *oshttp.WebconfigServer, r *mux.Router) {

	xc := dataapi.GetXconfConfigs(server.XW_XconfServer.ServerConfig.Config)

	WebServerInjection(server, xc)
	db.ConfigInjection(server.XW_XconfServer.ServerConfig.Config)
	dataapi.WebServerInjection(server.XW_XconfServer, xc)
	//dao.WebServerInjection(server)
	auth.WebServerInjection(server)
	dataapi.RegisterTables()

	db.RegisterTableConfigSimple(db.TABLE_TAG, tag.NewTagInf)
	initDB()
	db.GetCacheManager() // Initialize cache manager
	SetupDCMRoutes(server, r)
}

func SetupDCMRoutes(server *oshttp.WebconfigServer, r *mux.Router) {
	paths := []*mux.Router{}
	//authPaths := []*mux.Router{} // Do not required auth token validation middleware
	// Register DCM formula routes
	dcmFormulaPath := r.PathPrefix("/xconfAdminService/dcm/formula").Subrouter()
	dcmFormulaPath.HandleFunc("", GetDcmFormulaHandler).Methods("GET")
	dcmFormulaPath.HandleFunc("", CreateDcmFormulaHandler).Methods("POST")
	dcmFormulaPath.HandleFunc("", UpdateDcmFormulaHandler).Methods("PUT")
	dcmFormulaPath.HandleFunc("/filtered", PostDcmFormulaFilteredWithParamsHandler).Methods("POST")
	dcmFormulaPath.HandleFunc("/size", GetDcmFormulaSizeHandler).Methods("GET")
	dcmFormulaPath.HandleFunc("/names", GetDcmFormulaNamesHandler).Methods("GET")
	dcmFormulaPath.HandleFunc("/formulasAvailability", DcmFormulasAvailabilitygHandler).Methods("POST")
	dcmFormulaPath.HandleFunc("/settingsAvailability", DcmFormulaSettingsAvailabilitygHandler).Methods("POST")
	dcmFormulaPath.HandleFunc("/import/{overwrite}", ImportDcmFormulaWithOverwriteHandler).Methods("POST")

	// URL with var has to be placed last otherwise, it gets confused with url with defined paths
	dcmFormulaPath.HandleFunc("/{id}", GetDcmFormulaByIdHandler).Methods("GET")
	dcmFormulaPath.HandleFunc("/{id}", DeleteDcmFormulaByIdHandler).Methods("DELETE")
	dcmFormulaPath.HandleFunc("/{id}/priority/{newPriority}", DcmFormulaChangePriorityHandler).Methods("POST")
	paths = append(paths, dcmFormulaPath)

	dcmDeviceSettingsPath := r.PathPrefix("/xconfAdminService/dcm/deviceSettings").Subrouter()
	dcmDeviceSettingsPath.HandleFunc("", GetDeviceSettingsHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("", CreateDeviceSettingsHandler).Methods("POST").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("", UpdateDeviceSettingsHandler).Methods("PUT").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/size", GetDeviceSettingsSizeHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/names", GetDeviceSettingsNamesHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/filtered", PostDeviceSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/export", GetDeviceSettingsExportHandler).Methods("GET")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	dcmDeviceSettingsPath.HandleFunc("/{id}", DeleteDeviceSettingsByIdHandler).Methods("DELETE").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/{id}", GetDeviceSettingsByIdHandler).Methods("GET").Name("DCM-DeviceSettings")
	paths = append(paths, dcmDeviceSettingsPath)

	// dcm/vodsettings
	dcmVodSettingsPath := r.PathPrefix("/xconfAdminService/dcm/vodsettings").Subrouter()
	dcmVodSettingsPath.HandleFunc("", GetVodSettingsHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("", CreateVodSettingsHandler).Methods("POST").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("", UpdateVodSettingsHandler).Methods("PUT").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/size", GetVodSettingsSizeHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/names", GetVodSettingsNamesHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/filtered", PostVodSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/export", GetVodSettingExportHandler).Methods("GET").Name("DCM-VODSettings")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	dcmVodSettingsPath.HandleFunc("/{id}", DeleteVodSettingsByIdHandler).Methods("DELETE").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/{id}", GetVodSettingsByIdHandler).Methods("GET").Name("DCM-VODSettings")
	paths = append(paths, dcmVodSettingsPath)

	// dcm/uploadRepository
	dcmUploadRepositoryPath := r.PathPrefix("/xconfAdminService/dcm/uploadRepository").Subrouter()
	dcmUploadRepositoryPath.HandleFunc("", GetLogRepoSettingsHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("", CreateLogRepoSettingsHandler).Methods("POST").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("", UpdateLogRepoSettingsHandler).Methods("PUT").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/entities", PostLogRepoSettingsEntitiesHandler).Methods("POST").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/entities", PutLogRepoSettingsEntitiesHandler).Methods("PUT").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/size", GetLogRepoSettingsSizeHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/names", GetLogRepoSettingsNamesHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/filtered", PostLogRepoSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/{id}", DeleteLogRepoSettingsByIdHandler).Methods("DELETE").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/{id}", GetLogRepoSettingsByIdHandler).Methods("GET").Name("DCM-UploadRepository")
	paths = append(paths, dcmUploadRepositoryPath)

	// dcm/logUploadSettings
	dcmLogUploadSettingsPath := r.PathPrefix("/xconfAdminService/dcm/logUploadSettings").Subrouter()
	dcmLogUploadSettingsPath.HandleFunc("", GetLogUploadSettingsHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("", CreateLogUploadSettingsHandler).Methods("POST").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("", UpdateLogUploadSettingsHandler).Methods("PUT").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/size", GetLogUploadSettingsSizeHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/names", GetLogUploadSettingsNamesHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/filtered", PostLogUploadSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/export", GetLogRepoSettingsExportHandler).Methods("GET").Name("DCM-LogUploadSettings")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths)
	dcmLogUploadSettingsPath.HandleFunc("/{id}", DeleteLogUploadSettingsByIdHandler).Methods("DELETE").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/{id}", GetLogUploadSettingsByIdHandler).Methods("GET").Name("DCM-LogUploadSettings")
	paths = append(paths, dcmLogUploadSettingsPath)
	c := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"X-Requested-With", "Origin", "Content-Type", "Accept", "Authorization", "token"},
	})

	for _, p := range paths {
		p.Use(c.Handler)
		p.Use(server.XW_XconfServer.NoAuthMiddleware)
	}
}

type apiUnitTest struct {
	t        *testing.T
	router   *mux.Router
	savedMap map[string]string
}

func TestMain(m *testing.M) {
	fmt.Printf("in TestMain\n")

	testConfigFile = "/app/xconfadmin/xconfadmin.conf"
	if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
		testConfigFile = "../config/sample_xconfadmin.conf"
		if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
			panic(fmt.Errorf("config file problem %v", err))
		}
	}
	fmt.Printf("testConfigFile=%v\n", testConfigFile)

	os.Setenv("SECURITY_TOKEN_KEY", "testSecurityTokenKey")

	xpcKey := os.Getenv("XPC_KEY")
	if len(xpcKey) == 0 {
		os.Setenv("XPC_KEY", "testXpcKey")
	}

	cid := os.Getenv("SAT_CLIENT_ID")
	if len(cid) == 0 {
		os.Setenv("SAT_CLIENT_ID", "foo")
	}

	sec := os.Getenv("SAT_CLIENT_SECRET")
	if len(sec) == 0 {
		os.Setenv("SAT_CLIENT_SECRET", "bar")
	}
	cid = os.Getenv("IDP_CLIENT_ID")
	if len(cid) == 0 {
		os.Setenv("IDP_CLIENT_ID", "foo")
	}

	sec = os.Getenv("IDP_CLIENT_SECRET")
	if len(sec) == 0 {
		os.Setenv("IDP_CLIENT_SECRET", "bar")
	}

	ssrKeys := os.Getenv("X1_SSR_KEYS")
	if len(ssrKeys) == 0 {
		os.Setenv("X1_SSR_KEYS", "test-key-1;test-key-2;test-key3")
	}

	PartnerKeys := os.Getenv("PARTNER_KEYS")
	if len(PartnerKeys) == 0 {
		os.Setenv("PARTNER_KEYS", "test")
	}

	var err error
	sc, err = xwcommon.NewServerConfig(testConfigFile)
	if err != nil {
		panic(err)
	}

	server = oshttp.NewWebconfigServer(sc, true, nil, nil)
	defer server.XW_XconfServer.Server.Close()
	xwhttp.InitSatTokenManager(server.XW_XconfServer)

	// start clean
	db.SetDatabaseClient(server.XW_XconfServer.DatabaseClient)
	defer server.XW_XconfServer.DatabaseClient.Close()

	// setup router
	router = server.XW_XconfServer.GetRouter(false)

	// setup Xconf APIs and tables
	dataapi.XconfSetup(server.XW_XconfServer, router)
	dcmSetup(server, router)
	taggingapi.XconfTaggingServiceSetup(server, router)

	// tear down to start clean
	err = server.XW_XconfServer.SetUp()
	if err != nil {
		panic(err)
	}
	err = server.XW_XconfServer.TearDown()
	if err != nil {
		panic(err)
	}
	// DeleteAllEntities()

	globAut = newApiUnitTest(nil)

	returnCode := m.Run()

	globAut.t = nil

	// tear down to clean up
	server.XW_XconfServer.TearDown()

	os.Exit(returnCode)
}

func newApiUnitTest(t *testing.T) *apiUnitTest {
	if globAut != nil {
		globAut.t = t
		return globAut
	}
	aut := apiUnitTest{}
	aut.t = t
	aut.router = router
	aut.savedMap = make(map[string]string)

	globAut = &aut
	return &aut
}

func GetTestConfig() string {
	return "../../config/sample_xconfadmin.conf"
}

func GetTestWebConfigServer(testConfigFile string) (*oshttp.WebconfigServer, *mux.Router) {
	if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
		testConfigFile = "../../config/sample_xconfadmin.conf"
		if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
			panic(fmt.Errorf("config file problem %v", err))
		}
	}
	os.Setenv("SAT_CLIENT_ID", "foo")
	os.Setenv("SAT_CLIENT_SECRET", "bar")
	os.Setenv("IDP_CLIENT_ID", "foo")
	os.Setenv("IDP_CLIENT_SECRET", "bar")
	sc, err := xwcommon.NewServerConfig(testConfigFile)
	if err != nil {
		panic(err)
	}
	server := oshttp.NewWebconfigServer(sc, true, nil, nil)
	xwhttp.InitSatTokenManager(server.XW_XconfServer)
	router := server.XW_XconfServer.GetRouter(true)

	// Set up core database and data APIs
	dataapi.XconfSetup(server.XW_XconfServer, router)

	// Set up tagging service without going through adminapi
	taggingapi.XconfTaggingServiceSetup(server, router)

	// Don't call adminapi.XconfSetup to avoid circular dependency
	// We'll register just the routes we need in getDCMTestRouter

	return server, router
}

func ExecuteRequest(r *http.Request, handler http.Handler) *httptest.ResponseRecorder { // restored local version
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func DeleteAllEntities() {
	for _, tableInfo := range db.GetAllTableInfo() {
		if err := truncateTable(tableInfo.TableName); err != nil {
			fmt.Printf("failed to truncate table %s\n", tableInfo.TableName)
		}
		if tableInfo.CacheData {
			db.GetCachedSimpleDao().RefreshAll(tableInfo.TableName)
		}
	}
}

func truncateTable(tableName string) error {
	dbClient := db.GetDatabaseClient()
	cassandraClient, ok := dbClient.(*db.CassandraClient)
	if ok {
		return cassandraClient.DeleteAllXconfData(tableName)
	}
	return nil
}

func CreateAndSaveModel(id string) *core.Model {
	model := core.NewModel(id, "ModelDescription")
	err := db.GetCachedSimpleDao().SetOne(db.TABLE_MODEL, model.ID, model)
	if err != nil {
		return nil
	}

	return model
}

func CreateRule(relation string, freeArg rulesengine.FreeArg, operation string, fixedArgValue string) *rulesengine.Rule {
	rule := rulesengine.Rule{}
	rule.SetRelation(relation)
	rule.SetCondition(rulesengine.NewCondition(&freeArg, operation, rulesengine.NewFixedArg(fixedArgValue)))
	return &rule
}

func unmarshalXconfError(b []byte) *common.XconfError {
	var xconfError *common.XconfError
	_ = json.Unmarshal(b, &xconfError)
	return xconfError
}

func TestDfAllApi(t *testing.T) {
	//t.Skip("TODO: cpatel550 - need to move this test under adminapi")
	//config := GetTestConfig()
	//_, router := GetTestWebConfigServer(config)
	dfrule := logupload.DCMGenericRule{}
	err := json.Unmarshal([]byte(jsondfCreateData), &dfrule)
	assert.NilError(t, err)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, dfrule.ID, &dfrule)

	// create entry
	url := fmt.Sprintf("%s", DF_URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsondfPostCreateData))
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	// get dfrule by id
	urlWithId := fmt.Sprintf("%s/%s", DF_URL, "33af3261-d74a-40fd-8aa1-884e4f5479a1?applicationType=stb")
	req, err = http.NewRequest("GET", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// get dfrule size
	urlWithsize := fmt.Sprintf("%s/%s", DF_URL, "size")
	req, err = http.NewRequest("GET", urlWithsize, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var size string
		json.Unmarshal(body, &size)
		total, _ := strconv.Atoi(size)
		assert.Equal(t, total, 2)
	}

	// get dfrule Names
	urlWithnames := fmt.Sprintf("%s/%s", DF_URL, "names")
	req, err = http.NewRequest("GET", urlWithnames, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		names := []string{}
		json.Unmarshal(body, &names)
		assert.Equal(t, len(names) > 0, true)
	}

	// import dfrule with settings  for false means create
	urlWithImport := fmt.Sprintf("%s/%s", DF_URL, "import/false")

	impdatacr := []byte(
		`{"formula":{"compoundParts":[{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"SR203"}}}},"negated":false},{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"APPLE123"}}}},"negated":false,"relation":"OR"},{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"SKXI11AIS"}}}},"negated":false,"relation":"OR"}],"negated":false,"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"Dinesh_importgo3_formula","description":"","priority":1,"ruleExpression":"","percentage":100,"percentageL1":60,"percentageL2":20,"percentageL3":20,"applicationType":"stb"},"deviceSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-device","checkOnReboot":true,"configurationServiceURL":{"id":"","name":"","description":"","url":""},"settingsAreActive":true,"schedule":{"type":"CronExpression","expression":"0 8 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"applicationType":"stb"},"logUploadSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-log-upload","uploadOnReboot":true,"numberOfDays":100,"areSettingsActive":true,"schedule":{"type":"CronExpression","expression":"0 10 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"logFileIds":null,"logFilesGroupId":"","modeToGetLogFiles":"","uploadRepositoryId":"d49f4010-eb35-450a-927c-a4be8b68459a","activeDateTimeRange":false,"fromDateTime":"","toDateTime":"","applicationType":"stb"},"vodSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-vod","locationsURL":"https://test.net","ipNames":[],"ipList":[],"srmIPList":{},"applicationType":"stb"}}`)

	req, err = http.NewRequest("POST", urlWithImport+"?applicationType=stb", bytes.NewBuffer(impdatacr))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	// import dfrule with settings  for true means update
	urlWithImportup := fmt.Sprintf("%s/%s", DF_URL, "import/true")

	impdataup := []byte(
		`{"formula":{"compoundParts":[{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"SR203"}}}},"negated":false},{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"APPLE123"}}}},"negated":false,"relation":"OR"},{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"SKXI11AIS"}}}},"negated":false,"relation":"OR"}],"negated":false,"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"Dinesh_importgo3_formula_update","description":"","priority":1,"ruleExpression":"","percentage":100,"percentageL1":60,"percentageL2":20,"percentageL3":20,"applicationType":"stb"},"deviceSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-device_update","checkOnReboot":true,"configurationServiceURL":{"id":"","name":"","description":"","url":""},"settingsAreActive":true,"schedule":{"type":"CronExpression","expression":"0 8 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"applicationType":"stb"},"logUploadSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-log-upload_update","uploadOnReboot":true,"numberOfDays":100,"areSettingsActive":true,"schedule":{"type":"CronExpression","expression":"0 10 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"logFileIds":null,"logFilesGroupId":"","modeToGetLogFiles":"","uploadRepositoryId":"d49f4010-eb35-450a-927c-a4be8b68459a","activeDateTimeRange":false,"fromDateTime":"","toDateTime":"","applicationType":"stb"},"vodSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-vod_update","locationsURL":"https://test.net","ipNames":[],"ipList":[],"srmIPList":{},"applicationType":"stb"}}`)

	req, err = http.NewRequest("POST", urlWithImportup+"?applicationType=stb", bytes.NewBuffer(impdataup))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	// POST filtered Name
	urlfiltnames := fmt.Sprintf("%s/%s", DF_URL, "filtered?pageNumber=1&pageSize=50")
	req, err = http.NewRequest("POST", urlfiltnames, bytes.NewBuffer(postmapname))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var dfrules = []*logupload.DCMGenericRule{}
		json.Unmarshal(body, &dfrules)
		assert.Equal(t, len(dfrules), 3)
	}
	// get dfrule all
	req, err = http.NewRequest("GET", DF_URL, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	if res.StatusCode == http.StatusOK {
		var dfrules = []*logupload.DCMGenericRule{}
		json.Unmarshal(body, &dfrules)
		assert.Equal(t, len(dfrules), 3)
	}
	// filtered IP Arg
	urlfiltIParg := fmt.Sprintf("%s/%s", DF_URL, "filtered?pageNumber=1&pageSize=50")
	req, err = http.NewRequest("POST", urlfiltIParg, bytes.NewBuffer(postmapIPargs))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var dfrules = []*logupload.DCMGenericRule{}
		json.Unmarshal(body, &dfrules)
		assert.Equal(t, len(dfrules), 1)
	}

	// change priority
	priourl := "/xconfAdminService/dcm/formula/3f81ab29-ab8e-40d5-b407-cbc579b46caa/priority/1?applicationType=stb"
	req, err = http.NewRequest("POST", priourl, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var dfrules = []*logupload.DCMGenericRule{}
		json.Unmarshal(body, &dfrules)
		assert.Equal(t, len(dfrules) > 0, true)
	}

	//filreerd  MAC Args
	urlfiltMACarg := fmt.Sprintf("%s/%s", DF_URL, "filtered?pageNumber=1&pageSize=50?applicationType=stb")
	req, err = http.NewRequest("POST", urlfiltMACarg, bytes.NewBuffer(postmapMACargs))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var dfrules = []*logupload.DCMGenericRule{}
		json.Unmarshal(body, &dfrules)
		assert.Equal(t, len(dfrules), 1)
	}

	//settings Availability
	urlWithsetavail := fmt.Sprintf("%s/%s", DF_URL, "settingsAvailability?applicationType=stb")
	req, err = http.NewRequest("POST", urlWithsetavail, bytes.NewBuffer(payload))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		ret := make(map[string]map[string]bool)
		json.Unmarshal(body, &ret)
		assert.Equal(t, len(ret) > 0, true)
	}

	//formulas Availability
	urlWithavail := fmt.Sprintf("%s/%s", DF_URL, "formulasAvailability")
	req, err = http.NewRequest("POST", urlWithavail, bytes.NewBuffer(payload))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		ret := make(map[string]bool)
		json.Unmarshal(body, &ret)
		assert.Equal(t, len(ret) > 0, true)
	}

	//Error create duplicate Entry
	req, err = http.NewRequest("POST", DF_URL+"?applicationType=stb", bytes.NewBuffer(jsondfPostCreateData))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	//  Update  entry good case
	req, err = http.NewRequest("PUT", DF_URL+"?applicationType=stb", bytes.NewBuffer(jsondfUpdateData))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//  Update  entry error case
	req, err = http.NewRequest("PUT", DF_URL+"?applicationType=stb", bytes.NewBuffer(jsondfUpdateErrData))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	// delete dfrule by id
	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)

	// delete non existing dfrule by id
	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
}

func TestUpdatePriorityAndRuleInFormula_RuleIsUpdatedAndPrioritiesAreReorganized(t *testing.T) {
	DeleteAllEntities()
	numberOfFormulas := 10
	formulas := preCreateFormulas(numberOfFormulas, "TEST_MODEL", t)

	formulaToChangeIndex := 7
	var formulaToUpdate *logupload.DCMGenericRule
	b, _ := json.Marshal(formulas[formulaToChangeIndex])
	json.Unmarshal(b, &formulaToUpdate)
	newPriority := 8
	formulaToUpdate.Priority = newPriority
	formulaToUpdate.Rule = *CreateRule(rulesengine.RelationAnd, *coreef.RuleFactoryIP, rulesengine.StandardOperationIs, "10.10.10.10")

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/dcm/formula?%v", queryParams)

	formulaJson, _ := json.Marshal(formulaToUpdate)
	r := httptest.NewRequest("PUT", url, bytes.NewReader(formulaJson))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	receivedFormula := unmarshalFormula(rr.Body.Bytes())
	assert.Equal(t, newPriority, receivedFormula.Priority)
	assert.Equal(t, "10.10.10.10", receivedFormula.Rule.Condition.FixedArg.GetValue().(string))

	url = fmt.Sprintf("/xconfAdminService/dcm/formula/%s?%v", receivedFormula.ID, queryParams)
	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	receivedFormula = unmarshalFormula(rr.Body.Bytes())
	assert.Equal(t, formulaToUpdate.ID, receivedFormula.ID)
	assert.Equal(t, newPriority, receivedFormula.Priority)
	assert.Equal(t, "10.10.10.10", receivedFormula.Rule.Condition.FixedArg.GetValue().(string))

	url = fmt.Sprintf("/xconfAdminService/dcm/formula?%v", queryParams)
	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	receivedFormulas := unmarshalFormulas(rr.Body.Bytes())
	assert.Equal(t, numberOfFormulas, len(receivedFormulas))

	sort.Slice(receivedFormulas, func(i, j int) bool {
		return receivedFormulas[i].Priority < receivedFormulas[j].Priority
	})

	for i, formula := range receivedFormulas {
		assert.Equal(t, i+1, formula.Priority)
	}
}

func TestChangeFormulaPriorityWithNotValidValue_ExceptionIsThrown(t *testing.T) {
	DeleteAllEntities()
	formula := createFormula("MODEL_ID", 0)
	saveFormula(formula, t)
	newPriority := 0
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/dcm/formula/%s/priority/%v?%v", formula.ID, newPriority, queryParams)

	r := httptest.NewRequest("POST", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	xconfError := unmarshalXconfError(rr.Body.Bytes())
	assert.Equal(t, fmt.Sprintf("Invalid priority value %v", newPriority), xconfError.Message)
}

func preCreateFormulas(numberOfFormulas int, modelId string, t *testing.T) []*logupload.DCMGenericRule {
	createdFormulas := []*logupload.DCMGenericRule{}
	for i := 0; i < numberOfFormulas; i++ {
		formula := createFormula(modelId, i)
		saveFormula(formula, t)
		createdFormulas = append(createdFormulas, formula)
	}
	return createdFormulas
}

func createFormula(modelId string, testIndex int) *logupload.DCMGenericRule {
	model := CreateAndSaveModel(strings.ToUpper(fmt.Sprintf(modelId+"%v", testIndex)))
	formula := logupload.DCMGenericRule{}
	formula.ID = uuid.New().String()
	formula.Name = fmt.Sprintf("TEST_FORMULA_%v", testIndex)
	formula.Description = fmt.Sprintf("TEST_DESCRIPTION_%v", testIndex)
	formula.ApplicationType = core.STB
	formula.Rule = *CreateRule(rulesengine.RelationAnd, *coreef.RuleFactoryMODEL, rulesengine.StandardOperationIs, model.ID)
	formula.Priority = testIndex + 1
	formula.Percentage = 100
	return &formula
}

func saveFormula(formula *logupload.DCMGenericRule, t *testing.T) {
	queryParams, _ := util.GetURLQueryParameterString([][]string{{"applicationType", "stb"}})
	url := fmt.Sprintf("/xconfAdminService/dcm/formula?%v", queryParams)

	formulaJson, _ := json.Marshal(formula)
	r := httptest.NewRequest("POST", url, bytes.NewReader(formulaJson))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func unmarshalFormula(b []byte) *logupload.DCMGenericRule {
	var formula logupload.DCMGenericRule
	err := json.Unmarshal(b, &formula)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling formula: %v", err))
	}
	return &formula
}

func unmarshalFormulas(b []byte) []*logupload.DCMGenericRule {
	var formulas []*logupload.DCMGenericRule
	err := json.Unmarshal(b, &formulas)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling formulas: %v", err))
	}
	return formulas
}
