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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	"github.com/rdkcentral/xconfadmin/adminapi/firmware"
	"github.com/rdkcentral/xconfadmin/adminapi/queries"
	"github.com/rdkcentral/xconfadmin/adminapi/rfc/feature"
	"github.com/rdkcentral/xconfadmin/common"
	oshttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/taggingapi"
	"github.com/rdkcentral/xconfadmin/taggingapi/tag"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

const (
	Env_Url                    = "/xconfAdminService/queries"
	Queries_Rules_url          = "/xconfAdminService/queries/rules"
	Queries_Filter_url         = "/xconfAdminService/queries/filters"
	Queries_update_path        = "/xconfAdminService/updates"
	Queries_update_filter_path = "/xconfAdminService/updates/filters"
)

type TableData struct {
	Tablename string
	Tablerow  string
}

var (
	testConfigFile     string
	jsonTestConfigFile string
	sc                 *xwcommon.ServerConfig
	server             *oshttp.WebconfigServer
	router             *mux.Router
	//globAut            *apiUnitTest
)

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
	//SetupTestEnvironment()
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
	queriesSetup(server, router)
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

	//globAut = newApiUnitTest(nil)

	returnCode := m.Run()

	//globAut.t = nil

	// tear down to clean up
	server.XW_XconfServer.TearDown()

	os.Exit(returnCode)
}

func ImportTableData(data []interface{}) error {
	var err error
	for _, row := range data {
		switch row.(TableData).Tablename {
		case "TABLE_ENVIRONMENT":
			var tabletype = shared.Environment{}
			err = json.Unmarshal([]byte(row.(TableData).Tablerow), &tabletype)
			err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_ENVIRONMENT, tabletype.ID, &tabletype)
			break
		case "TABLE_GENERIC_NS_LIST":
			var humptyStrList = []string{
				"Humpty Dumpty sat on a wall",
				"Humpty Dumpty had a great fall",
				"All the king's horses and all the king's men",
				"Couldn't put Humpty together again",
			}

			tabletype := shared.NewGenericNamespacedList(fmt.Sprintf("CDN-TESTING"), "STRING", humptyStrList)
			ipList := []string{
				"127.1.1.1",
				"127.1.1.2",
				"127.1.1.3",
			}

			tabletype.TypeName = "IP_LIST"
			tabletype.Data = ipList
			err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, tabletype.ID, tabletype)
			break
		case "TABLE_FIRMWARE_CONFIG":
			var firmwareConfig = coreef.NewEmptyFirmwareConfig()
			err = json.Unmarshal([]byte(row.(TableData).Tablerow), &firmwareConfig)
			err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_FIRMWARE_CONFIG, firmwareConfig.ID, firmwareConfig)
			break

		case "TABLE_FIRMWARE_RULE":
			var firmwareRule = corefw.NewEmptyFirmwareRule()
			var data_str = row.(TableData).Tablerow
			err = json.Unmarshal([]byte(data_str), &firmwareRule)
			err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_FIRMWARE_RULE, firmwareRule.ID, firmwareRule)
			break

		case "TABLE_SINGLETON_FILTER_VALUE":
			var data_str = row.(TableData).Tablerow
			locationRoundRobinFilter := coreef.NewEmptyDownloadLocationRoundRobinFilterValue()
			err = json.Unmarshal([]byte(data_str), &locationRoundRobinFilter)
			err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_SINGLETON_FILTER_VALUE, locationRoundRobinFilter.ID, locationRoundRobinFilter)
			break
		}

	}

	return err
}

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
	//initAppSettings()             // Initialize Application settings
}

func queriesSetup(server *oshttp.WebconfigServer, r *mux.Router) {
	xc := dataapi.GetXconfConfigs(server.XW_XconfServer.ServerConfig.Config)

	WebServerInjection(server, xc)
	db.ConfigInjection(server.XW_XconfServer.ServerConfig.Config)
	dataapi.WebServerInjection(server.XW_XconfServer, xc)
	//dao.WebServerInjection(server)
	auth.WebServerInjection(server)
	dataapi.RegisterTables()
	setupRoutes(server, router)
	db.RegisterTableConfigSimple(db.TABLE_TAG, tag.NewTagInf)
	initDB()
	db.GetCacheManager() // Initialize cache manager
}

func setupRoutes(server *oshttp.WebconfigServer, r *mux.Router) {
	// Register DCM formula routes
	paths := []*mux.Router{}
	//dcmFormulaPath := r.PathPrefix("/xconfAdminService/dcm/formula").Subrouter()

	// Note: We cannot import the dcm package here due to circular dependency
	// Instead, the DCM tests will need to set up these routes themselves
	// or we need to use function injection/callbacks
	queriesPath := r.PathPrefix("/xconfAdminService/queries").Subrouter()
	queriesPath.HandleFunc("/environments", GetQueriesEnvironments).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/environments/{id}", GetQueriesEnvironmentsById).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/models", GetModelHandler).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/models/{id}", GetModelByIdHandler).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/ipAddressGroups", GetQueriesIpAddressGroups).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/ipAddressGroups/byIp/{ipAddress}", GetQueriesIpAddressGroupsByIp).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/ipAddressGroups/byName/{name}", GetQueriesIpAddressGroupsByName).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/ipAddressGroups", GetQueriesIpAddressGroupsV2).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/ipAddressGroups/byIp/{ipAddress}", GetQueriesIpAddressGroupsByIpV2).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/ipAddressGroups/byName/{id}", GetQueriesIpAddressGroupsByNameV2).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/nsLists", GetQueriesMacLists).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/nsLists/byId/{id}", GetQueriesMacListsById).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/nsLists/byMacPart/{mac}", GetQueriesMacListsByMacPart).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/nsLists", GetQueriesMacLists).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/nsLists/byId/{id}", GetQueriesMacListsByIdV2).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/nsLists/byMacPart/{mac}", GetQueriesMacListsByMacPart).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/firmwares", GetFirmwareConfigHandler).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/firmwares/{id}", GetFirmwareConfigByIdHandler).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/firmwares/model/{modelId}", GetQueriesFirmwareConfigsByModelId).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/firmwares/bySupportedModels", PostFirmwareConfigBySupportedModelsHandler).Methods("POST").Name("Queries")
	queriesPath.HandleFunc("/percentageBean", GetQueriesPercentageBean).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/percentageBean/{id}", GetQueriesPercentageBeanById).Methods("GET").Name("Queries")
	paths = append(paths, queriesPath)

	queriesRulesPath := r.PathPrefix("/xconfAdminService/queries/rules").Subrouter()
	queriesRulesPath.HandleFunc("/ips", GetQueriesRulesIps).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/ips/{ruleName}", GetIpRuleById).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/ips/byIpAddressGroup/{ipAddressGroupName}", GetIpRuleByIpAddressGroup).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/macs", GetQueriesRulesMacs).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/macs/{ruleName}", GetMACRuleByName).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/macs/address/{macAddress}", GetMACRulesByMAC).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/envModels", GetQueriesRulesEnvModels).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/envModels/{name}", GetEnvModelRuleByNameHandler).Methods("GET").Name("QueriesRules")
	paths = append(paths, queriesRulesPath)

	queriesFiltersPath := r.PathPrefix("/xconfAdminService/queries/filters").Subrouter()
	queriesFiltersPath.HandleFunc("/ips", GetQueriesFiltersIps).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/ips/{name}", GetQueriesFiltersIpsByName).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/time", GetQueriesFiltersTime).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/time/{name}", GetQueriesFiltersTimeByName).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/locations", GetQueriesFiltersLocation).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/locations/{name}", GetQueriesFiltersLocationByName).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/locations/byName/{name}", GetQueriesFiltersLocationByName).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/downloadlocation", GetQueriesFiltersDownloadLocation).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/percent", GetQueriesFiltersPercent).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/ri", GetQueriesFiltersRebootImmediately).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/ri/{name}", GetQueriesFiltersRebootImmediatelyByName).Methods("GET").Name("QueriesFilters")
	paths = append(paths, queriesFiltersPath)

	updatePath := r.PathPrefix("/xconfAdminService/updates").Subrouter()
	updatePath.HandleFunc("/environments", CreateEnvironmentHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/models", CreateModelHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/models", UpdateModelHandler).Methods("PUT").Name("Updates")
	updatePath.HandleFunc("/rules/ips", UpdateIpRule).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/rules/macs", SaveMACRule).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/rules/envModels", UpdateEnvModelRuleHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/ipAddressGroups", CreateIpAddressGroupHandler).Methods("POST", "PUT").Name("Updates")
	updatePath.HandleFunc("/ipAddressGroups/{listId}/addData", AddDataIpAddressGroupHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/ipAddressGroups/{listId}/removeData", RemoveDataIpAddressGroupHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/v2/ipAddressGroups", CreateIpAddressGroupHandlerV2).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/v2/ipAddressGroups", UpdateIpAddressGroupHandlerV2).Methods("PUT").Name("Updates")
	updatePath.HandleFunc("/nsLists", SaveMacListHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/nsLists/{listId}/addData", AddDataMacListHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/nsLists/{listId}/removeData", RemoveDataMacListHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/v2/nsLists", CreateMacListHandlerV2).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/v2/nsLists", UpdateMacListHandlerV2).Methods("PUT").Name("Updates")
	updatePath.HandleFunc("/v2/nsLists/{listId}/addData", AddDataMacListHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/v2/nsLists/{listId}/removeData", RemoveDataMacListHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/firmwares", PostFirmwareConfigHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/firmwares", PutFirmwareConfigHandler).Methods("PUT").Name("Updates")
	updatePath.HandleFunc("/percentageBean", CreatePercentageBeanHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/percentageBean", UpdatePercentageBeanHandler).Methods("PUT").Name("Updates")
	updatePath.HandleFunc("/logFile", CreateLogFile).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/logUploadSettings/{timezone}/{scheduleTimezone}", NotImplementedHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/deviceSettings", NotImplementedHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/deviceSettings/{scheduleTimeZone}", NotImplementedHandler).Methods("POST").Name("Updates")
	paths = append(paths, updatePath)

	updateFilterPath := r.PathPrefix("/xconfAdminService/updates/filters").Subrouter()
	updateFilterPath.HandleFunc("/ips", UpdateIpsFilterHandler).Methods("POST").Name("UpdatesFilters")
	updateFilterPath.HandleFunc("/time", UpdateTimeFilterHandler).Methods("POST").Name("UpdatesFilters")
	updateFilterPath.HandleFunc("/locations", UpdateLocationFilterHandler).Methods("POST").Name("UpdatesFilters")
	updateFilterPath.HandleFunc("/downloadlocation", UpdateDownloadLocationFilterHandler).Methods("POST").Name("UpdatesFilters")
	updateFilterPath.HandleFunc("/percent", UpdatePercentFilterHandler).Methods("POST").Name("UpdatesFilters")
	updateFilterPath.HandleFunc("/ri", UpdateRebootImmediatelyHandler).Methods("POST").Name("UpdatesFilters")
	paths = append(paths, updateFilterPath)

	amvPath := r.PathPrefix("/xconfAdminService/amv").Subrouter()
	amvPath.HandleFunc("", GetAmvHandler).Methods("GET").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("", CreateAmvHandler).Methods("POST").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("", UpdateAmvHandler).Methods("PUT").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("/page", NotImplementedHandler).Methods("GET").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("/filtered", GetAmvFilteredHandler).Methods("GET").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("/importAll", ImportAllAmvHandler).Methods("POST").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("/{id}", DeleteAmvByIdHandler).Methods("DELETE").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("/{id}", GetAmvByIdHandler).Methods("GET").Name("Firmware-ActivationVersion")
	paths = append(paths, amvPath)

	// featurerule
	featureRulePath := r.PathPrefix("/xconfAdminService/featurerule").Subrouter()
	featureRulePath.HandleFunc("", GetFeatureRulesHandler).Methods("GET").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("/filtered", GetFeatureRulesFiltered).Methods("GET").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("/{id}", GetFeatureRuleOne).Methods("GET").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("", CreateFeatureRuleHandler).Methods("POST").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("", UpdateFeatureRuleHandler).Methods("PUT").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("/importAll", ImportAllFeatureRulesHandler).Methods("POST").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("/{id}", DeleteOneFeatureRuleHandler).Methods("DELETE").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("/", DeleteOneFeatureRuleHandler).Methods("DELETE").Name("RFC-FeatureRules")
	paths = append(paths, featureRulePath)

	// feature
	featurePath := r.PathPrefix("/xconfAdminService/feature").Subrouter()
	featurePath.HandleFunc("", GetFeatureEntityHandler).Methods("GET").Name("RFC-Feature")
	featurePath.HandleFunc("/filtered", GetFeatureEntityFilteredHandler).Methods("GET").Name("RFC-Feature")
	featurePath.HandleFunc("/{id}", GetFeatureEntityByIdHandler).Methods("GET").Name("RFC-Feature")
	featurePath.HandleFunc("/{id}", feature.DeleteFeatureByIdHandler).Methods("DELETE").Name("RFC-Feature")
	featurePath.HandleFunc("", PostFeatureEntityHandler).Methods("POST").Name("RFC-Feature")
	featurePath.HandleFunc("", PutFeatureEntityHandler).Methods("PUT").Name("RFC-Feature")
	featurePath.HandleFunc("/importAll", PostFeatureEntityImportAllHandler).Methods("POST").Name("RFC-Feature")
	paths = append(paths, featurePath)

	// model
	modelPath := r.PathPrefix("/xconfAdminService/model").Subrouter()
	modelPath.HandleFunc("", GetModelHandler).Methods("GET").Name("Models")
	modelPath.HandleFunc("", CreateModelHandler).Methods("POST").Name("Models")
	modelPath.HandleFunc("", UpdateModelHandler).Methods("PUT").Name("Models")
	modelPath.HandleFunc("/entities", PostModelEntitiesHandler).Methods("POST").Name("Models")
	modelPath.HandleFunc("/entities", PutModelEntitiesHandler).Methods("PUT").Name("Models")
	modelPath.HandleFunc("/filtered", PostModelFilteredHandler).Methods("POST").Name("Models")
	modelPath.HandleFunc("/page", NotImplementedHandler).Methods("GET").Name("Models")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	modelPath.HandleFunc("/{id}", DeleteModelHandler).Methods("DELETE").Name("Models")
	modelPath.HandleFunc("/{id}", GetModelByIdHandler).Methods("GET").Name("Models")
	paths = append(paths, modelPath)

	// firmwarerule
	firmwareRulePath := r.PathPrefix("/xconfAdminService/firmwarerule").Subrouter()
	firmwareRulePath.HandleFunc("/filtered", GetFirmwareRuleFilteredHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/importAll", PostFirmwareRuleImportAllHandler).Methods("POST").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/{type}/names", GetFirmwareRuleByTypeNamesHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/byTemplate/{templateId}/names", GetFirmwareRuleByTemplateByTemplateIdNamesHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/export/byType", GetFirmwareRuleExportByTypeHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/export/allTypes", GetFirmwareRuleExportAllTypesHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/testpage", firmware.GetFirmwareTestPageHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("", GetFirmwareRuleHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("", PostFirmwareRuleHandler).Methods("POST").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("", PutFirmwareRuleHandler).Methods("PUT").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/entities", PostFirmwareRuleEntitiesHandler).Methods("POST").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/entities", PutFirmwareRuleEntitiesHandler).Methods("PUT").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/filtered", PostFirmwareRuleFilteredHandler).Methods("POST").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/page", NotImplementedHandler).Methods("GET").Name("Firmware-Rules")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	firmwareRulePath.HandleFunc("/{id}", DeleteFirmwareRuleByIdHandler).Methods("DELETE").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/{id}", GetFirmwareRuleByIdHandler).Methods("GET").Name("Firmware-Rules")
	paths = append(paths, firmwareRulePath)

	// firmwareruletemplate
	firmwareRuleTempPath := r.PathPrefix("/xconfAdminService/firmwareruletemplate").Subrouter()
	firmwareRuleTempPath.HandleFunc("/filtered", GetFirmwareRuleTemplateFilteredHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/importAll", PostFirmwareRuleTemplateImportAllHandler).Methods("POST").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/all/{type}", GetFirmwareRuleTemplateAllByTypeHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/ids", GetFirmwareRuleTemplateIdsHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/{id}/priority/{newPriority}", PostChangePriorityHandler).Methods("POST").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/export", GetFirmwareRuleTemplateExportHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/{type}/{isEditable}", GetFirmwareRuleTemplateWithVarWithVarHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("", GetFirmwareRuleTemplateHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("", PostFirmwareRuleTemplateHandler).Methods("POST").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("", PutFirmwareRuleTemplateHandler).Methods("PUT").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/entities", PostFirmwareRuleTemplateEntitiesHandler).Methods("POST").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/entities", PutFirmwareRuleTemplateEntitiesHandler).Methods("PUT").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/filtered", PostFirmwareRuleTemplateFilteredHandler).Methods("POST").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/page", NotImplementedHandler).Methods("GET").Name("Firmware-Templates")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	firmwareRuleTempPath.HandleFunc("/{id}", DeleteFirmwareRuleTemplateByIdHandler).Methods("DELETE").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/{id}", GetFirmwareRuleTemplateByIdHandler).Methods("GET").Name("Firmware-Templates")
	paths = append(paths, firmwareRuleTempPath)

	// penetration data report
	penetrationPath := r.PathPrefix("/xconfAdminService/penetrationdata").Subrouter()
	penetrationPath.HandleFunc("/{macAddress}", GetPenetrationDataByMacHandler).Methods("GET").Name("PenetrationData")
	paths = append(paths, penetrationPath)

	// percentfilter/percentageBean
	percentageBeanPath := r.PathPrefix("/xconfAdminService/percentfilter/percentageBean").Subrouter()
	percentageBeanPath.HandleFunc("", GetPercentageBeanAllHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("", CreatePercentageBeanHandler).Methods("POST").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("", UpdatePercentageBeanHandler).Methods("PUT").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/page", NotImplementedHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/filtered", PostPercentageBeanFilteredWithParamsHandler).Methods("POST").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/entities", PostPercentageBeanEntitiesHandler).Methods("POST").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/entities", PutPercentageBeanEntitiesHandler).Methods("PUT").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/allAsRules", GetAllPercentageBeanAsRule).Methods("GET").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/asRule/{id}", GetPercentageBeanAsRuleById).Methods("GET").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/{id}", GetPercentageBeanByIdHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/{id}", DeletePercentageBeanByIdHandler).Methods("DELETE").Name("Firmware-PercentFilter")
	paths = append(paths, percentageBeanPath)
	// percentfilter
	percentageFilterPath := r.PathPrefix("/xconfAdminService/percentfilter").Subrouter()
	percentageFilterPath.HandleFunc("", GetPercentFilterGlobalHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageFilterPath.HandleFunc("", UpdatePercentFilterGlobalHandler).Methods("POST").Name("Firmware-PercentFilter")
	percentageFilterPath.HandleFunc("/globalPercentage", GetGlobalPercentFilterHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageFilterPath.HandleFunc("/calculator", GetCalculatedHashAndPercentHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageFilterPath.HandleFunc("/globalPercentage/asRule", GetGlobalPercentFilterAsRuleHandler).Methods("GET").Name("Firmware-PercentFilter")
	paths = append(paths, percentageFilterPath)
	deletePath := r.PathPrefix("/xconfAdminService/delete").Subrouter()
	deletePath.HandleFunc("/environments/{id}", DeleteEnvironmentHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/models/{id}", DeleteModelHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/rules/ips/{name}", DeleteIpRule).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/rules/macs/{name}", DeleteMACRule).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/rules/envModels/{name}", DeleteEnvModelRuleBeanHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/ipAddressGroups/{id}", DeleteIpAddressGroupHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/v2/ipAddressGroups/{id}", DeleteIpAddressGroupHandlerV2).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/nsLists/{id}", DeleteMacListHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/v2/nsLists/{id}", DeleteMacListHandlerV2).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/firmwares/{id}", DeleteFirmwareConfigHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/percentageBean/{id}", DeletePercentageBeanHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/filters/ips/{name}", DeleteIpsFilterHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/filters/time/{name}", DeleteTimeFilterHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/filters/locations/{name}", DeleteLocationFilterHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/filters/ri/{name}", DeleteRebootImmediatelyHandler).Methods("DELETE").Name("Delete")
	paths = append(paths, deletePath)

	// percentfilter/percentageBean
	percentageBeanPath := r.PathPrefix("/xconfAdminService/percentfilter/percentageBean").Subrouter()
	percentageBeanPath.HandleFunc("", queries.GetPercentageBeanAllHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("", queries.CreatePercentageBeanHandler).Methods("POST").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("", queries.UpdatePercentageBeanHandler).Methods("PUT").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/filtered", queries.PostPercentageBeanFilteredWithParamsHandler).Methods("POST").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/entities", queries.PostPercentageBeanEntitiesHandler).Methods("POST").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/entities", queries.PutPercentageBeanEntitiesHandler).Methods("PUT").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/allAsRules", queries.GetAllPercentageBeanAsRule).Methods("GET").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/asRule/{id}", queries.GetPercentageBeanAsRuleById).Methods("GET").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/{id}", queries.GetPercentageBeanByIdHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageBeanPath.HandleFunc("/{id}", queries.DeletePercentageBeanByIdHandler).Methods("DELETE").Name("Firmware-PercentFilter")
	paths = append(paths, percentageBeanPath)

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

func TestAllQueriesApis(t *testing.T) {
	//server, _ := SetupTestEnvironment()
	DeleteAllEntities()

	table_data := []interface{}{
		TableData{Tablename: "TABLE_ENVIRONMENT", Tablerow: `{"id":"AX061AEI","updated":1591604177484,"description":"RT1319"}`},
		TableData{Tablename: "TABLE_GENERIC_NS_LIST", Tablerow: ``},
		TableData{Tablename: "TABLE_FIRMWARE_CONFIG", Tablerow: `{"id":"207dc5a5-d324-4e2e-9daf-5017ed98f8f3","updated":1558520642121,"description":"CPEAUTO_FW_AA:AA:AA:AA:AA:AA","supportedModelIds":["XCONFTESTMODEL"],"firmwareDownloadProtocol":"http","firmwareFilename":"DPC3941_3.3p17s1_DEV_sey-test","firmwareVersion":"DPC3941_3.3p17s1_DEV_sey-test","rebootImmediately":false,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"437afab9-cbe3-4e4d-b175-220865e0f720","name":" Cisco Arris XG1","rule":{"negated":false,"compoundParts":[{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"ipAddress"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":""}}}}},{"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"env"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"VBN"}}}}},{"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"MX011ANC"}}}}}]},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configId":"e675358b-506d-48f8-86c5-c8c8e3bb6254","active":true,"firmwareCheckRequired":false,"rebootImmediately":false},"type":"IP_RULE","active":true}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"c4681132-c518-459a-99fb-9b93a1f42f37","name":"CDN-TESTING","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CDN-TESTING"}}}}},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configId":"dff46b03-be65-4f0c-804d-542d5ffec8ec","active":true,"firmwareCheckRequired":false,"rebootImmediately":false},"type":"MAC_RULE","active":true,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"67333656-9e8e-46a3-9a87-2f42644a35c9","name":"Arris_XG1v1_VBN_Moto-DEV","rule":{"negated":false,"compoundParts":[{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"env"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"VBN"}}}}},{"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"MX011ANM"}}}}},{"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"partnerId"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"testDEV"}}}}}]},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configEntries":[{"configId":"5de4a2df-2673-4be3-ae67-4e09648a929b","percentage":100.0,"startPercentRange":0.0,"endPercentRange":100.0}],"active":true,"firmwareCheckRequired":true,"rebootImmediately":true,"firmwareVersions":["MX011AN_3.8p3s1_VBN_sey","MX011AN_3.1p1s3_VBN_sey","MX011AN_3.2p6s1_VBN_sey-test"]},"type":"ENV_MODEL_RULE","active":true,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"c4681132-c518-459a-99fb-9b93a1f41gf37","name":"Test_Ip_filter_device","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CDN-TESTING"}}}}},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configId":"dff46b03-be65-4f0c-804d-542d5ffec8ec","active":true,"firmwareCheckRequired":false,"rebootImmediately":false},"type":"IP_FILTER","active":true,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"c4681132-c518-459a-99fb-9b93a1f63534","name":"Test_Time_filter_device","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CDN-TESTING"}}}}},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configId":"dff46b03-be65-4f0c-804d-542d5ffec8ec","active":true,"firmwareCheckRequired":false,"rebootImmediately":false},"type":"TIME_FILTER","active":true,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"67f595ae-3e1d-418d-9b86-22b3e46816e4","name":"CPEAUTO_LF_80:f5:03:34:11:fd","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"ipAddress"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CPEAUTOIPGRP80f5033411fd"}}}}},"applicableAction":{"type":".DefinePropertiesAction","ttlMap":{},"actionType":"DEFINE_PROPERTIES","properties":{"firmwareLocation":"http://ssr.ccp.xcal.tv/cgi-bin/x1-sign-redirect.pl?K=10&F=stb_cdl","firmwareDownloadProtocol":"http","ipv6FirmwareLocation":""},"activationFirmwareVersions":{}},"type":"DOWNLOAD_LOCATION_FILTER","active":true,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_SINGLETON_FILTER_VALUE", Tablerow: `{"type":"com.comcast.xconf.estbfirmware.DownloadLocationRoundRobinFilterValue","id":"DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE","updated":1616699042493,"applicationType":"stb","locations":[{"locationIp":"96.114.220.246","percentage":100.0},{"locationIp":"69.252.106.162","percentage":0.0}],"ipv6locations":[{"locationIp":"2600:1f18:227b:c01:b161:3d17:7a86:fe36","percentage":100.0},{"locationIp":"2001:558:1020:1:250:56ff:fe94:646f","percentage":0.0}],"httpLocation":"test.com","httpFullUrlLocation":"https://test.com/Images"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"e313bc81-8a02-4087-8c91-1da6db4b3159","name":"CDL-ARRISXG1V4-QA","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CDL-ARRISXG1V4-QA"}}}}},"applicableAction":{"type":".DefinePropertiesAction","ttlMap":{},"actionType":"DEFINE_PROPERTIES","properties":{"rebootImmediately":"true"},"byPassFilters":[]},"type":"REBOOT_IMMEDIATELY_FILTER","active":true}`},
	}
	err := ImportTableData(table_data)
	assert.NilError(t, err)
	//GET ENVIRONMENTS
	url := fmt.Sprintf("%s/%s", Env_Url, "environments")
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Get ENVIRONMENTS BY ID
	urlWithId := fmt.Sprintf("%s/%s/%s", Env_Url, "environments", "AX061AEI")
	req, err = http.NewRequest("GET", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET IPADDRESSGROUPS
	url = fmt.Sprintf("%s/%s", Env_Url, "ipAddressGroups")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET IPADDRESSGROUPS BY IP
	url = fmt.Sprintf("%s/%s", Env_Url, "ipAddressGroups/byIp/127.1.1.1")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET NSLISTS
	url = fmt.Sprintf("%s/%s", Env_Url, "nsLists")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET NSLISTS BY ID
	url = fmt.Sprintf("%s/%s", Env_Url, "nsLists/byId/"+"wweii2900292ii39")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET FIRMWARES BY MODEL ID

	url = fmt.Sprintf("%s/%s", Env_Url, "firmwares/model/"+"XCONFTESTMODEL?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

	//GET FIRMWARES BY SUPORTEDMODELS
	var postData = []byte(
		`["XCONFTESTMODEL"]`,
	)
	url = fmt.Sprintf("%s/%s", Env_Url, "firmwares/bySupportedModels?applicationType=stb")
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(postData))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//-----------------------------------------------------
	//QUERIES RULES API'S
	//-------------------------------------------------------

	//GET IPS RULES
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, "ips?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 0, true)
	}

	//GET IPS RULES BY NAME
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, `ips/ Cisco Arris XG1?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 0, true)
	}

	//GET MAC RULES
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, "macs?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 0, true)
	}

	//GET MAC RULES BY RULE NAME
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, `macs/CDN-TESTING?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET ENV MODELS
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, "envModels?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET ENV MODELS WITH NAME
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, "envModels/Arris_XG1v1_VBN_Moto-DEV?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//-----------------------------------------------------
	//QUERIES FILTERS API'S
	//-------------------------------------------------------

	//GET IPS FILTER
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, "ips?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET IPS RULES BY NAME
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, `ips/Test_Ip_filter_device?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET TIME FILTER
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, "time?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET TIME FILTER BY NAME
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, `time/Test_Time_filter_device?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET LOCATION FILTER
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, "locations?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET LOCATION FILTER BY NAME
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, `locations/CPEAUTO_LF_80:f5:03:34:11:fd?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET DOWNLOAD LOCATION
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, "downloadlocation?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET REEBOOT IMMEDIATELY FILTER
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, "ri?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET REEBOOT IMMEDIATELY FILTER BY NAME
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, `ri/CDL-ARRISXG1V4-QA?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//-----------------------------------------------------
	//QUERIES UPDATES API'S
	//-----------------------------------------------------
	var body_data = []byte(`{"id":"AX061AE2","updated":1541604177484,"description":"TESTRT1319"}`)
	url = fmt.Sprintf("%s/%s", Queries_update_path, "environments")
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(body_data))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)
	body, err = ioutil.ReadAll(res.Body)

	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//-----------------------------------------------------
	//QUERIES UPDATES FILTERS API'S
	//-----------------------------------------------------

	//POST IPS FILTER

	body_data = []byte(`{"id":"c4681132-c518-459a-99fb-9b93a1f41gf37","name":"Test_Ip_filter_device","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CDN-TESTING"}}}}},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configId":"dff46b03-be65-4f0c-804d-542d5ffec8ec","active":true,"firmwareCheckRequired":false,"rebootImmediately":false},"type":"IP_FILTER","IpAddressGroup":{"Id":"CDN-TESTING","Name":"CDN-TESTING","IpAddresses":["127.1.1.1","127.1.1.2","127.1.1.3"],"RawIpAddresses":["127.1.1.1","127.1.1.2","127.1.1.3"]},"active":true,"applicationType":"stb"}`)
	url = fmt.Sprintf("%s/%s", Queries_update_filter_path, "ips?applicationType=stb")
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(body_data))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

}
