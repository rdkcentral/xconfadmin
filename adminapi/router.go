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
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	"github.com/rdkcentral/xconfadmin/adminapi/canary"
	"github.com/rdkcentral/xconfadmin/adminapi/change"
	ipmacrule "github.com/rdkcentral/xconfadmin/adminapi/configuration/ip-macrule"
	"github.com/rdkcentral/xconfadmin/adminapi/dcm"
	"github.com/rdkcentral/xconfadmin/adminapi/firmware"
	"github.com/rdkcentral/xconfadmin/adminapi/lockdown"
	"github.com/rdkcentral/xconfadmin/adminapi/queries"
	"github.com/rdkcentral/xconfadmin/adminapi/rfc/feature"
	"github.com/rdkcentral/xconfadmin/adminapi/setting"
	"github.com/rdkcentral/xconfadmin/adminapi/telemetry"
	"github.com/rdkcentral/xconfadmin/adminapi/xcrp"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/taggingapi"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rs/cors"
)

// Xconf setup
func XconfSetup(server *xhttp.WebconfigServer, r *mux.Router) {
	xc := dataapi.GetXconfConfigs(server.XW_XconfServer.ServerConfig.Config)

	WebServerInjection(server, xc)
	db.ConfigInjection(server.XW_XconfServer.ServerConfig.Config)
	dataapi.WebServerInjection(server.XW_XconfServer, xc)
	auth.WebServerInjection(server)
	dataapi.RegisterTables()

	db.GetCacheManager() // Initialize cache manager
	initDB()

	if server.XW_XconfServer.ServerConfig.GetBoolean("xconfwebconfig.xconf.dataservice_enabled") {
		dataapi.XconfSetup(server.XW_XconfServer, r)
	}

	if server.XW_XconfServer.ServerConfig.GetBoolean("xconfwebconfig.xconf.adminservice_enabled") {
		RouteXconfAdminserviceApis(server, r)
	}

	if server.XW_XconfServer.ServerConfig.GetBoolean("xconfwebconfig.xconf.enable_tagging_service_admin") {
		taggingapi.XconfTaggingServiceSetup(server, r)
	}
}

func TrailingSlashRemover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}
func RouteXconfAdminserviceApis(s *xhttp.WebconfigServer, r *mux.Router) {
	paths := []*mux.Router{}
	authPaths := []*mux.Router{} // Do not required auth token validation middleware

	// Auth APIs
	providerPath := r.PathPrefix("/xconfAdminService/provider").Subrouter()
	providerPath.HandleFunc("", auth.AuthProvider).Methods("GET").Name("Auth-Uncategorized")
	authPaths = append(authPaths, providerPath)

	authInfoPath := r.PathPrefix("/xconfAdminService/auth/info").Subrouter()
	authInfoPath.HandleFunc("", auth.AuthInfoHandler).Methods("GET").Name("Auth-Uncategorized")
	paths = append(paths, authInfoPath)

	basicAuthpath := r.PathPrefix("/xconfAdminService/auth/basic").Subrouter()
	basicAuthpath.HandleFunc("", auth.BasicAuthHandler).Methods("POST").Name("Auth-Basic")
	paths = append(paths, authInfoPath)

	loginPath := r.Path("/xconfAdminService" + s.IdpLoginPath).Subrouter()
	loginPath.HandleFunc("", auth.LoginUrlHandler).Methods("GET").Name("Auth-Xerxes")
	authPaths = append(authPaths, loginPath)

	logoutPath := r.Path("/xconfAdminService" + s.IdpLogoutPath).Subrouter()
	logoutPath.HandleFunc("", auth.LogoutHandler).Methods("GET").Name("Auth-Xerxes")
	authPaths = append(authPaths, logoutPath)

	logoutAfterPath := r.Path("/xconfAdminService" + s.IdpLogoutAfterPath).Subrouter()
	logoutAfterPath.HandleFunc("", auth.LogoutAfterHandler).Methods("GET").Name("Auth-Xerxes")
	authPaths = append(authPaths, logoutAfterPath)

	codePath := r.Path("/xconfAdminService" + s.IdpCodePath).Subrouter()
	codePath.HandleFunc("", auth.CodeHandler).Methods("GET").Name("Auth-Xerxes")
	authPaths = append(authPaths, codePath)

	urlPath := r.Path("/xconfAdminService" + s.IdpUrlPath).Subrouter()
	urlPath.HandleFunc("", auth.LoginUrlHandler).Methods("GET").Name("Auth-Xerxes")
	authPaths = append(authPaths, urlPath)

	// DataService bypass APIs
	dsBypassPathPrefix := r.PathPrefix("/xconfAdminService/dataService").Subrouter()
	dsBypassPathPrefix.HandleFunc("/xconf/swu/{applicationType}", dataapi.GetEstbFirmwareSwuHandler).Methods("GET").Name("DataServiceByPass")
	dsBypassPathPrefix.HandleFunc("/estbfirmware/lastlog", dataapi.GetEstbLastlogPath).Methods("GET").Name("DataServiceByPass")
	dsBypassPathPrefix.HandleFunc("/estbfirmware/changelogs", dataapi.GetEstbChangelogsPath).Methods("GET").Name("DataServiceByPass")
	dsBypassPathPrefix.HandleFunc("/queries/filters/percent", queries.GetQueriesFiltersPercent).Methods("GET").Name("DataServiceByPass")
	dsBypassPathPrefix.HandleFunc("/firmwarerule/filtered", queries.GetFirmwareRuleFilteredHandler).Methods("GET").Name("DataServiceByPass")
	paths = append(paths, dsBypassPathPrefix)

	// APIs moved from DataService to AdminService
	getEstbLastlogPath := r.Path("/xconfAdminService/estbfirmware/lastlog").Subrouter()
	getEstbLastlogPath.HandleFunc("", dataapi.GetEstbLastlogPath).Name("Firmware-Logs")
	paths = append(paths, getEstbLastlogPath)

	getEstbChangelogsPath := r.Path("/xconfAdminService/estbfirmware/changelogs").Subrouter()
	getEstbChangelogsPath.HandleFunc("", dataapi.GetEstbChangelogsPath).Name("Firmware-Logs")
	paths = append(paths, getEstbChangelogsPath)

	getInfoPathPrefix := r.PathPrefix("/xconfAdminService/info").Subrouter()
	getInfoPathPrefix.HandleFunc("/refreshAll", queries.GetInfoRefreshAllHandler).Methods("GET").Name("General-Uncategorized")
	getInfoPathPrefix.HandleFunc("/refresh/{tableName}", queries.GetInfoRefreshHandler).Methods("GET").Name("General-Uncategorized")
	getInfoPathPrefix.HandleFunc("/statistics", queries.GetInfoStatistics).Methods("GET").Name("General-Uncategorized")
	getInfoPathPrefix.HandleFunc("/tables", queries.GetInfoTableNames).Methods("GET").Name("General-Uncategorized")
	getInfoPathPrefix.HandleFunc("/tables/{tableName}", queries.GetInfoTable).Methods("GET").Name("General-Uncategorized")
	getInfoPathPrefix.HandleFunc("/tables/{tableName}/{rowKey}", queries.GetInfoTableRowKey).Methods("GET").Name("General-Uncategorized")
	getInfoPathPrefix.HandleFunc("/tables/{tableName}/{rowKey}", queries.UpdateInfoTableRowKey).Methods("PUT").Name("General-Uncategorized")
	paths = append(paths, getInfoPathPrefix)

	queriesPath := r.PathPrefix("/xconfAdminService/queries").Subrouter()
	queriesPath.HandleFunc("/environments", queries.GetQueriesEnvironments).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/environments/{id}", queries.GetQueriesEnvironmentsById).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/models", queries.GetModelHandler).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/models/{id}", queries.GetModelByIdHandler).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/ipAddressGroups", queries.GetQueriesIpAddressGroups).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/ipAddressGroups/byIp/{ipAddress}", queries.GetQueriesIpAddressGroupsByIp).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/ipAddressGroups/byName/{name}", queries.GetQueriesIpAddressGroupsByName).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/ipAddressGroups", queries.GetQueriesIpAddressGroupsV2).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/ipAddressGroups/byIp/{ipAddress}", queries.GetQueriesIpAddressGroupsByIpV2).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/ipAddressGroups/byName/{id}", queries.GetQueriesIpAddressGroupsByNameV2).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/nsLists", queries.GetQueriesMacLists).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/nsLists/byId/{id}", queries.GetQueriesMacListsById).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/nsLists/byMacPart/{mac}", queries.GetQueriesMacListsByMacPart).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/nsLists", queries.GetQueriesMacLists).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/nsLists/byId/{id}", queries.GetQueriesMacListsByIdV2).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/v2/nsLists/byMacPart/{mac}", queries.GetQueriesMacListsByMacPart).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/firmwares", queries.GetFirmwareConfigHandler).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/firmwares/{id}", queries.GetFirmwareConfigByIdHandler).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/firmwares/model/{modelId}", queries.GetQueriesFirmwareConfigsByModelId).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/firmwares/bySupportedModels", queries.PostFirmwareConfigBySupportedModelsHandler).Methods("POST").Name("Queries")
	queriesPath.HandleFunc("/percentageBean", queries.GetQueriesPercentageBean).Methods("GET").Name("Queries")
	queriesPath.HandleFunc("/percentageBean/{id}", queries.GetQueriesPercentageBeanById).Methods("GET").Name("Queries")
	paths = append(paths, queriesPath)

	queriesRulesPath := r.PathPrefix("/xconfAdminService/queries/rules").Subrouter()
	queriesRulesPath.HandleFunc("/ips", queries.GetQueriesRulesIps).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/ips/{ruleName}", queries.GetIpRuleById).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/ips/byIpAddressGroup/{ipAddressGroupName}", queries.GetIpRuleByIpAddressGroup).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/macs", queries.GetQueriesRulesMacs).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/macs/{ruleName}", queries.GetMACRuleByName).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/macs/address/{macAddress}", queries.GetMACRulesByMAC).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/envModels", queries.GetQueriesRulesEnvModels).Methods("GET").Name("QueriesRules")
	queriesRulesPath.HandleFunc("/envModels/{name}", queries.GetEnvModelRuleByNameHandler).Methods("GET").Name("QueriesRules")
	paths = append(paths, queriesRulesPath)

	queriesFiltersPath := r.PathPrefix("/xconfAdminService/queries/filters").Subrouter()
	queriesFiltersPath.HandleFunc("/ips", queries.GetQueriesFiltersIps).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/ips/{name}", queries.GetQueriesFiltersIpsByName).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/time", queries.GetQueriesFiltersTime).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/time/{name}", queries.GetQueriesFiltersTimeByName).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/locations", queries.GetQueriesFiltersLocation).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/locations/{name}", queries.GetQueriesFiltersLocationByName).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/locations/byName/{name}", queries.GetQueriesFiltersLocationByName).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/downloadlocation", queries.GetQueriesFiltersDownloadLocation).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/percent", queries.GetQueriesFiltersPercent).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/ri", queries.GetQueriesFiltersRebootImmediately).Methods("GET").Name("QueriesFilters")
	queriesFiltersPath.HandleFunc("/ri/{name}", queries.GetQueriesFiltersRebootImmediatelyByName).Methods("GET").Name("QueriesFilters")
	paths = append(paths, queriesFiltersPath)

	updatePath := r.PathPrefix("/xconfAdminService/updates").Subrouter()
	updatePath.HandleFunc("/environments", queries.CreateEnvironmentHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/models", queries.CreateModelHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/models", queries.UpdateModelHandler).Methods("PUT").Name("Updates")
	updatePath.HandleFunc("/rules/ips", queries.UpdateIpRule).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/rules/macs", queries.SaveMACRule).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/rules/envModels", queries.UpdateEnvModelRuleHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/ipAddressGroups", queries.CreateIpAddressGroupHandler).Methods("POST", "PUT").Name("Updates")
	updatePath.HandleFunc("/ipAddressGroups/{listId}/addData", queries.AddDataIpAddressGroupHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/ipAddressGroups/{listId}/removeData", queries.RemoveDataIpAddressGroupHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/v2/ipAddressGroups", queries.CreateIpAddressGroupHandlerV2).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/v2/ipAddressGroups", queries.UpdateIpAddressGroupHandlerV2).Methods("PUT").Name("Updates")
	updatePath.HandleFunc("/nsLists", queries.SaveMacListHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/nsLists/{listId}/addData", queries.AddDataMacListHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/nsLists/{listId}/removeData", queries.RemoveDataMacListHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/v2/nsLists", queries.CreateMacListHandlerV2).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/v2/nsLists", queries.UpdateMacListHandlerV2).Methods("PUT").Name("Updates")
	updatePath.HandleFunc("/v2/nsLists/{listId}/addData", queries.AddDataMacListHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/v2/nsLists/{listId}/removeData", queries.RemoveDataMacListHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/firmwares", queries.PostFirmwareConfigHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/firmwares", queries.PutFirmwareConfigHandler).Methods("PUT").Name("Updates")
	updatePath.HandleFunc("/percentageBean", queries.CreatePercentageBeanHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/percentageBean", queries.UpdatePercentageBeanHandler).Methods("PUT").Name("Updates")
	updatePath.HandleFunc("/logFile", queries.CreateLogFile).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/logUploadSettings/{timezone}/{scheduleTimezone}", queries.NotImplementedHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/deviceSettings", queries.NotImplementedHandler).Methods("POST").Name("Updates")
	updatePath.HandleFunc("/deviceSettings/{scheduleTimeZone}", queries.NotImplementedHandler).Methods("POST").Name("Updates")
	paths = append(paths, updatePath)

	updateFilterPath := r.PathPrefix("/xconfAdminService/updates/filters").Subrouter()
	updateFilterPath.HandleFunc("/ips", queries.UpdateIpsFilterHandler).Methods("POST").Name("UpdatesFilters")
	updateFilterPath.HandleFunc("/time", queries.UpdateTimeFilterHandler).Methods("POST").Name("UpdatesFilters")
	updateFilterPath.HandleFunc("/locations", queries.UpdateLocationFilterHandler).Methods("POST").Name("UpdatesFilters")
	updateFilterPath.HandleFunc("/downloadlocation", queries.UpdateDownloadLocationFilterHandler).Methods("POST").Name("UpdatesFilters")
	updateFilterPath.HandleFunc("/percent", queries.UpdatePercentFilterHandler).Methods("POST").Name("UpdatesFilters")
	updateFilterPath.HandleFunc("/ri", queries.UpdateRebootImmediatelyHandler).Methods("POST").Name("UpdatesFilters")
	paths = append(paths, updateFilterPath)

	deletePath := r.PathPrefix("/xconfAdminService/delete").Subrouter()
	deletePath.HandleFunc("/environments/{id}", queries.DeleteEnvironmentHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/models/{id}", queries.DeleteModelHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/rules/ips/{name}", queries.DeleteIpRule).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/rules/macs/{name}", queries.DeleteMACRule).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/rules/envModels/{name}", queries.DeleteEnvModelRuleBeanHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/ipAddressGroups/{id}", queries.DeleteIpAddressGroupHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/v2/ipAddressGroups/{id}", queries.DeleteIpAddressGroupHandlerV2).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/nsLists/{id}", queries.DeleteMacListHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/v2/nsLists/{id}", queries.DeleteMacListHandlerV2).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/firmwares/{id}", queries.DeleteFirmwareConfigHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/percentageBean/{id}", queries.DeletePercentageBeanHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/filters/ips/{name}", queries.DeleteIpsFilterHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/filters/time/{name}", queries.DeleteTimeFilterHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/filters/locations/{name}", queries.DeleteLocationFilterHandler).Methods("DELETE").Name("Delete")
	deletePath.HandleFunc("/filters/ri/{name}", queries.DeleteRebootImmediatelyHandler).Methods("DELETE").Name("Delete")
	paths = append(paths, deletePath)

	// model
	modelPath := r.PathPrefix("/xconfAdminService/model").Subrouter()
	modelPath.HandleFunc("", queries.GetModelHandler).Methods("GET").Name("Models")
	modelPath.HandleFunc("", queries.CreateModelHandler).Methods("POST").Name("Models")
	modelPath.HandleFunc("", queries.UpdateModelHandler).Methods("PUT").Name("Models")
	modelPath.HandleFunc("/entities", queries.PostModelEntitiesHandler).Methods("POST").Name("Models")
	modelPath.HandleFunc("/entities", queries.PutModelEntitiesHandler).Methods("PUT").Name("Models")
	modelPath.HandleFunc("/filtered", queries.PostModelFilteredHandler).Methods("POST").Name("Models")
	modelPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Models")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	modelPath.HandleFunc("/{id}", queries.DeleteModelHandler).Methods("DELETE").Name("Models")
	modelPath.HandleFunc("/{id}", queries.GetModelByIdHandler).Methods("GET").Name("Models")
	paths = append(paths, modelPath)

	// environment
	environmentPath := r.PathPrefix("/xconfAdminService/environment").Subrouter()
	environmentPath.HandleFunc("", queries.GetQueriesEnvironments).Methods("GET").Name("Environments")
	environmentPath.HandleFunc("", queries.CreateEnvironmentHandler).Methods("POST").Name("Environments")
	environmentPath.HandleFunc("", queries.UpdateEnvironmentHandler).Methods("PUT").Name("Environments")
	environmentPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Environments")
	environmentPath.HandleFunc("/filtered", queries.PostEnvironmentFilteredHandler).Methods("POST").Name("Environments")
	environmentPath.HandleFunc("/entities", queries.PostEnvironmentEntitiesHandler).Methods("POST").Name("Environments")
	environmentPath.HandleFunc("/entities", queries.PutEnvironmentEntitiesHandler).Methods("PUT").Name("Environments")
	environmentPath.HandleFunc("/{id}", queries.GetQueriesEnvironmentsById).Methods("GET").Name("Environments")
	environmentPath.HandleFunc("/{id}", queries.DeleteEnvironmentHandler).Methods("DELETE").Name("Environments")
	paths = append(paths, environmentPath)

	// genericnamespacedlist
	nameSpacedListPath := r.PathPrefix("/xconfAdminService/genericnamespacedlist").Subrouter()
	nameSpacedListPath.HandleFunc("", queries.GetNamespacedListsHandler).Methods("GET").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("", queries.CreateNamespacedListHandler).Methods("POST").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("", queries.UpdateNamespacedListHandler).Methods("PUT").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/ids", queries.GetNamespacedListIdsHandler).Methods("GET").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/ipAddressGroups", queries.GetIpAddressGroupsHandler).Methods("GET").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/filtered", queries.PostNamespacedListFilteredHandler).Methods("POST").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/entities", queries.PostNamespacedListEntitiesHandler).Methods("POST").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/entities", queries.PutNamespacedListEntitiesHandler).Methods("PUT").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/{id}", queries.GetNamespacedListHandler).Methods("GET").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/{id}", queries.RenameNamespacedListHandler).Methods("PUT").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/{id}", queries.DeleteNamespacedListHandler).Methods("DELETE").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/{type}/ids", queries.GetNamespacedListIdsByTypeHandler).Methods("GET").Name("NameSpaced-Lists")
	nameSpacedListPath.HandleFunc("/all/{type}", queries.GetNamespacedListsByTypeHandler).Methods("GET").Name("NameSpaced-Lists")
	paths = append(paths, nameSpacedListPath)

	// firmwarerule
	firmwareRulePath := r.PathPrefix("/xconfAdminService/firmwarerule").Subrouter()
	firmwareRulePath.HandleFunc("/filtered", queries.GetFirmwareRuleFilteredHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/importAll", queries.PostFirmwareRuleImportAllHandler).Methods("POST").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/{type}/names", queries.GetFirmwareRuleByTypeNamesHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/byTemplate/{templateId}/names", queries.GetFirmwareRuleByTemplateByTemplateIdNamesHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/export/byType", queries.GetFirmwareRuleExportByTypeHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/export/allTypes", queries.GetFirmwareRuleExportAllTypesHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/testpage", firmware.GetFirmwareTestPageHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("", queries.GetFirmwareRuleHandler).Methods("GET").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("", queries.PostFirmwareRuleHandler).Methods("POST").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("", queries.PutFirmwareRuleHandler).Methods("PUT").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/entities", queries.PostFirmwareRuleEntitiesHandler).Methods("POST").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/entities", queries.PutFirmwareRuleEntitiesHandler).Methods("PUT").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/filtered", queries.PostFirmwareRuleFilteredHandler).Methods("POST").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Firmware-Rules")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	firmwareRulePath.HandleFunc("/{id}", queries.DeleteFirmwareRuleByIdHandler).Methods("DELETE").Name("Firmware-Rules")
	firmwareRulePath.HandleFunc("/{id}", queries.GetFirmwareRuleByIdHandler).Methods("GET").Name("Firmware-Rules")
	paths = append(paths, firmwareRulePath)

	// firmwareruletemplate
	firmwareRuleTempPath := r.PathPrefix("/xconfAdminService/firmwareruletemplate").Subrouter()
	firmwareRuleTempPath.HandleFunc("/filtered", queries.GetFirmwareRuleTemplateFilteredHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/importAll", queries.PostFirmwareRuleTemplateImportAllHandler).Methods("POST").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/all/{type}", queries.GetFirmwareRuleTemplateAllByTypeHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/ids", queries.GetFirmwareRuleTemplateIdsHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/{id}/priority/{newPriority}", queries.PostChangePriorityHandler).Methods("POST").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/export", queries.GetFirmwareRuleTemplateExportHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/{type}/{isEditable}", queries.GetFirmwareRuleTemplateWithVarWithVarHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("", queries.GetFirmwareRuleTemplateHandler).Methods("GET").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("", queries.PostFirmwareRuleTemplateHandler).Methods("POST").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("", queries.PutFirmwareRuleTemplateHandler).Methods("PUT").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/entities", queries.PostFirmwareRuleTemplateEntitiesHandler).Methods("POST").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/entities", queries.PutFirmwareRuleTemplateEntitiesHandler).Methods("PUT").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/filtered", queries.PostFirmwareRuleTemplateFilteredHandler).Methods("POST").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Firmware-Templates")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	firmwareRuleTempPath.HandleFunc("/{id}", queries.DeleteFirmwareRuleTemplateByIdHandler).Methods("DELETE").Name("Firmware-Templates")
	firmwareRuleTempPath.HandleFunc("/{id}", queries.GetFirmwareRuleTemplateByIdHandler).Methods("GET").Name("Firmware-Templates")
	paths = append(paths, firmwareRuleTempPath)

	// firmwareconfig
	firmwareConfigPath := r.PathPrefix("/xconfAdminService/firmwareconfig").Subrouter()
	firmwareConfigPath.HandleFunc("/firmwareConfigMap", queries.GetFirmwareConfigFirmwareConfigMapHandler).Methods("GET").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("/getSortedFirmwareVersionsIfExistOrNot", queries.PostFirmwareConfigGetSortedFirmwareVersionsIfExistOrNotHandler).Methods("POST").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("/model/{modelId}", queries.GetFirmwareConfigModelByModelIdHandler).Methods("GET").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("/supportedConfigsByEnvModelRuleName/{ruleName}", queries.GetSupportedConfigsByEnvModelRuleName).Methods("GET").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("/byEnvModelRuleName/{ruleName}", queries.GetFirmwareConfigByEnvModelRuleNameByRuleNameHandler).Methods("GET").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("", queries.GetFirmwareConfigHandler).Methods("GET").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("", queries.PostFirmwareConfigHandler).Methods("POST").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("", queries.PutFirmwareConfigHandler).Methods("PUT").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("/bySupportedModels", queries.PostFirmwareConfigBySupportedModelsHandler).Methods("POST").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("/entities", queries.PostFirmwareConfigEntitiesHandler).Methods("POST").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("/entities", queries.PutFirmwareConfigEntitiesHandler).Methods("PUT").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("/filtered", queries.PostFirmwareConfigFilteredHandler).Methods("POST").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Firmware-Configs")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	firmwareConfigPath.HandleFunc("/{id}", queries.DeleteFirmwareConfigByIdHandler).Methods("DELETE").Name("Firmware-Configs")
	firmwareConfigPath.HandleFunc("/{id}", queries.GetFirmwareConfigByIdHandler).Methods("GET").Name("Firmware-Configs")
	paths = append(paths, firmwareConfigPath)

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

	// percentfilter
	percentageFilterPath := r.PathPrefix("/xconfAdminService/percentfilter").Subrouter()
	percentageFilterPath.HandleFunc("", queries.GetPercentFilterGlobalHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageFilterPath.HandleFunc("", queries.UpdatePercentFilterGlobalHandler).Methods("POST").Name("Firmware-PercentFilter")
	percentageFilterPath.HandleFunc("/globalPercentage", queries.GetGlobalPercentFilterHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageFilterPath.HandleFunc("/calculator", queries.GetCalculatedHashAndPercentHandler).Methods("GET").Name("Firmware-PercentFilter")
	percentageFilterPath.HandleFunc("/globalPercentage/asRule", queries.GetGlobalPercentFilterAsRuleHandler).Methods("GET").Name("Firmware-PercentFilter")
	paths = append(paths, percentageFilterPath)

	// roundrobinfilter
	roundrobinFilterPath := r.PathPrefix("/xconfAdminService/roundrobinfilter").Subrouter()
	roundrobinFilterPath.HandleFunc("", queries.UpdateDownloadLocationFilterHandler).Methods("POST").Name("Firmware-DownloadLocationFilter")
	roundrobinFilterPath.HandleFunc("/{applicationType}", queries.GetRoundRobinFilterHandler).Methods("GET").Name("Firmware-DownloadLocationFilter")
	paths = append(paths, roundrobinFilterPath)

	// amv
	amvPath := r.PathPrefix("/xconfAdminService/amv").Subrouter()
	amvPath.HandleFunc("", queries.GetAmvHandler).Methods("GET").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("", queries.CreateAmvHandler).Methods("POST").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("", queries.UpdateAmvHandler).Methods("PUT").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("/filtered", queries.GetAmvFilteredHandler).Methods("GET").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("/importAll", queries.ImportAllAmvHandler).Methods("POST").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("/{id}", queries.DeleteAmvByIdHandler).Methods("DELETE").Name("Firmware-ActivationVersion")
	amvPath.HandleFunc("/{id}", queries.GetAmvByIdHandler).Methods("GET").Name("Firmware-ActivationVersion")
	paths = append(paths, amvPath)

	// activationMinimumVersion - code is same as amv only export
	actMinVerPath := r.PathPrefix("/xconfAdminService/activationMinimumVersion").Subrouter()
	actMinVerPath.HandleFunc("", queries.GetAmvHandler).Methods("GET").Name("Firmware-ActivationVersion")
	actMinVerPath.HandleFunc("", queries.CreateAmvHandler).Methods("POST").Name("Firmware-ActivationVersion")
	actMinVerPath.HandleFunc("", queries.UpdateAmvHandler).Methods("PUT").Name("Firmware-ActivationVersion")
	actMinVerPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Firmware-ActivationVersion")
	actMinVerPath.HandleFunc("/filtered", queries.PostAmvFilteredHandler).Methods("POST").Name("Firmware-ActivationVersion")
	actMinVerPath.HandleFunc("/entities", queries.PostAmvEntitiesHandler).Methods("POST").Name("Firmware-ActivationVersion")
	actMinVerPath.HandleFunc("/entities", queries.PutAmvEntitiesHandler).Methods("PUT").Name("Firmware-ActivationVersion")
	actMinVerPath.HandleFunc("/importAll", queries.ImportAllAmvHandler).Methods("POST").Name("Firmware-ActivationVersion")
	actMinVerPath.HandleFunc("/{id}", queries.DeleteAmvByIdHandler).Methods("DELETE").Name("Firmware-ActivationVersion")
	actMinVerPath.HandleFunc("/{id}", queries.GetAmvByIdHandler).Methods("GET").Name("Firmware-ActivationVersion")
	paths = append(paths, actMinVerPath)

	// setting/profile
	settingProfilePath := r.PathPrefix("/xconfAdminService/setting/profile").Subrouter()
	settingProfilePath.HandleFunc("", setting.CreateSettingProfileHandler).Methods("POST").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/entities", setting.CreateSettingProfilesPackageHandler).Methods("POST").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("", setting.UpdateSettingProfilesHandler).Methods("PUT").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/entities", setting.UpdateSettingProfilesPackageHandler).Methods("PUT").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("", setting.GetSettingProfilesAllExport).Methods("GET").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/{id}", setting.GetSettingProfileOneExport).Methods("GET").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/filtered", setting.GetSettingProfilesFilteredWithPage).Methods("POST").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/{id}", setting.DeleteOneSettingProfilesHandler).Methods("DELETE").Name("Settings-Profiles")
	paths = append(paths, settingProfilePath)

	// setting/rule
	settingRulePath := r.PathPrefix("/xconfAdminService/setting/rule").Subrouter()
	settingRulePath.HandleFunc("", setting.CreateSettingRuleHandler).Methods("POST").Name("Settings-Rules")
	settingRulePath.HandleFunc("/entities", setting.CreateSettingRulesPackageHandler).Methods("POST").Name("Settings-Rules")
	settingRulePath.HandleFunc("", setting.UpdateSettingRulesHandler).Methods("PUT").Name("Settings-Rules")
	settingRulePath.HandleFunc("/entities", setting.UpdateSettingRulesPackageHandler).Methods("PUT").Name("Settings-Rules")
	settingRulePath.HandleFunc("", setting.GetSettingRulesAllExport).Methods("GET").Name("Settings-Rules")
	settingRulePath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Settings-Rules")
	settingRulePath.HandleFunc("/{id}", setting.GetSettingRuleOneExport).Methods("GET").Name("Settings-Rules")
	settingRulePath.HandleFunc("/filtered", setting.GetSettingRulesFilteredWithPage).Methods("POST").Name("Settings-Rules")
	settingRulePath.HandleFunc("/{id}", setting.DeleteOneSettingRulesHandler).Methods("DELETE").Name("Settings-Rules")
	paths = append(paths, settingRulePath)

	// settings/testpage
	settingTestpagePath := r.PathPrefix("/xconfAdminService/settings/testpage").Subrouter()
	settingTestpagePath.HandleFunc("", setting.SettingTestPageHandler).Methods("POST").Name("Settings-TestPage")
	paths = append(paths, settingTestpagePath)

	// featurerule
	featureRulePath := r.PathPrefix("/xconfAdminService/featurerule").Subrouter()
	featureRulePath.HandleFunc("", queries.GetFeatureRulesHandler).Methods("GET").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("/filtered", queries.GetFeatureRulesFiltered).Methods("GET").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("/{id}", queries.GetFeatureRuleOne).Methods("GET").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("", queries.CreateFeatureRuleHandler).Methods("POST").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("", queries.UpdateFeatureRuleHandler).Methods("PUT").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("/importAll", queries.ImportAllFeatureRulesHandler).Methods("POST").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("/{id}", queries.DeleteOneFeatureRuleHandler).Methods("DELETE").Name("RFC-FeatureRules")
	featureRulePath.HandleFunc("/", queries.DeleteOneFeatureRuleHandler).Methods("DELETE").Name("RFC-FeatureRules")
	paths = append(paths, featureRulePath)

	// feature
	featurePath := r.PathPrefix("/xconfAdminService/feature").Subrouter()
	featurePath.HandleFunc("", queries.GetFeatureEntityHandler).Methods("GET").Name("RFC-Feature")
	featurePath.HandleFunc("/filtered", queries.GetFeatureEntityFilteredHandler).Methods("GET").Name("RFC-Feature")
	featurePath.HandleFunc("/{id}", queries.GetFeatureEntityByIdHandler).Methods("GET").Name("RFC-Feature")
	featurePath.HandleFunc("/{id}", feature.DeleteFeatureByIdHandler).Methods("DELETE").Name("RFC-Feature")
	featurePath.HandleFunc("", queries.PostFeatureEntityHandler).Methods("POST").Name("RFC-Feature")
	featurePath.HandleFunc("", queries.PutFeatureEntityHandler).Methods("PUT").Name("RFC-Feature")
	featurePath.HandleFunc("/importAll", queries.PostFeatureEntityImportAllHandler).Methods("POST").Name("RFC-Feature")
	paths = append(paths, featurePath)

	// rfc
	rfcFeaturerulePath := r.PathPrefix("/xconfAdminService/rfc").Subrouter()
	rfcFeaturerulePath.HandleFunc("/featurerule/{id}/priority/{newPriority}", queries.ChangeFeatureRulePrioritiesHandler).Methods("POST").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule/size", queries.GetFeatureRulesSizeHandler).Methods("GET").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule/allowedNumberOfFeatures", queries.GetAllowedNumberOfFeaturesHandler).Methods("GET").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule", queries.CreateFeatureRuleHandler).Methods("POST").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule/entities", queries.CreateFeatureRulesHandler).Methods("POST").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule", queries.UpdateFeatureRuleHandler).Methods("PUT").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule/entities", queries.UpdateFeatureRulesHandler).Methods("PUT").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule", queries.GetFeatureRulesExportHandler).Methods("GET").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule/page", queries.NotImplementedHandler).Methods("GET").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule/{id}", queries.GetFeatureRuleOneExport).Methods("GET").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule/filtered", queries.GetFeatureRulesFilteredWithPage).Methods("POST").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/featurerule/{id}", queries.DeleteOneFeatureRuleHandler).Methods("DELETE").Name("RFC-FeatureRules")
	rfcFeaturerulePath.HandleFunc("/test", queries.FeatureRuleTestPageHandler).Methods("POST").Name("RFC-FeatureRules")
	paths = append(paths, rfcFeaturerulePath)

	// rfc/feature
	rfcFeaturePath := r.PathPrefix("/xconfAdminService/rfc/feature").Subrouter()
	rfcFeaturePath.HandleFunc("", feature.PostFeatureHandler).Methods("POST").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("", feature.PutFeatureHandler).Methods("PUT").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/entities", feature.PostFeatureEntitiesHandler).Methods("POST").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/entities", feature.PutFeatureEntitiesHandler).Methods("PUT").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("", feature.GetFeaturesHandler).Methods("GET").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/{id}", feature.GetFeatureByIdHandler).Methods("GET").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/{id}", feature.DeleteFeatureByIdHandler).Methods("DELETE").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/filtered", feature.GetFeaturesFilteredHandler).Methods("POST").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/byIdList", feature.GetFeaturesByIdListHandler).Methods("POST").Name("RFC-Feature")
	paths = append(paths, rfcFeaturePath)

	// dcm/formula
	dcmFormulaPath := r.PathPrefix("/xconfAdminService/dcm/formula").Subrouter()
	dcmFormulaPath.HandleFunc("", dcm.GetDcmFormulaHandler).Methods("GET").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("", dcm.CreateDcmFormulaHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("", dcm.UpdateDcmFormulaHandler).Methods("PUT").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/entities", dcm.PostDcmFormulaListHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/entities", dcm.PutDcmFormulaListHandler).Methods("PUT").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/list", dcm.PostDcmFormulaListHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/list", dcm.PutDcmFormulaListHandler).Methods("PUT").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/settingsAvailability", dcm.DcmFormulaSettingsAvailabilitygHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/import/{overwrite}", dcm.ImportDcmFormulaWithOverwriteHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/import", dcm.ImportDcmFormulasHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/formulasAvailability", dcm.DcmFormulasAvailabilitygHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/size", dcm.GetDcmFormulaSizeHandler).Methods("GET").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/names", dcm.GetDcmFormulaNamesHandler).Methods("GET").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/filtered", dcm.PostDcmFormulaFilteredWithParamsHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/{id}/priority/{newPriority}", dcm.DcmFormulaChangePriorityHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/{id}", dcm.DeleteDcmFormulaByIdHandler).Methods("DELETE").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/{id}", dcm.GetDcmFormulaByIdHandler).Methods("GET").Name("DCM-Formulas")
	paths = append(paths, dcmFormulaPath)

	// dcm/deviceSettings
	dcmDeviceSettingsPath := r.PathPrefix("/xconfAdminService/dcm/deviceSettings").Subrouter()
	dcmDeviceSettingsPath.HandleFunc("", dcm.GetDeviceSettingsHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("", dcm.CreateDeviceSettingsHandler).Methods("POST").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("", dcm.UpdateDeviceSettingsHandler).Methods("PUT").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/size", dcm.GetDeviceSettingsSizeHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/names", dcm.GetDeviceSettingsNamesHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/filtered", dcm.PostDeviceSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/export", dcm.GetDeviceSettingsExportHandler).Methods("GET")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	dcmDeviceSettingsPath.HandleFunc("/{id}", dcm.DeleteDeviceSettingsByIdHandler).Methods("DELETE").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/{id}", dcm.GetDeviceSettingsByIdHandler).Methods("GET").Name("DCM-DeviceSettings")
	paths = append(paths, dcmDeviceSettingsPath)

	// dcm/vodsettings
	dcmVodSettingsPath := r.PathPrefix("/xconfAdminService/dcm/vodsettings").Subrouter()
	dcmVodSettingsPath.HandleFunc("", dcm.GetVodSettingsHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("", dcm.CreateVodSettingsHandler).Methods("POST").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("", dcm.UpdateVodSettingsHandler).Methods("PUT").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/size", dcm.GetVodSettingsSizeHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/names", dcm.GetVodSettingsNamesHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/filtered", dcm.PostVodSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/export", dcm.GetVodSettingExportHandler).Methods("GET").Name("DCM-VODSettings")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	dcmVodSettingsPath.HandleFunc("/{id}", dcm.DeleteVodSettingsByIdHandler).Methods("DELETE").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/{id}", dcm.GetVodSettingsByIdHandler).Methods("GET").Name("DCM-VODSettings")
	paths = append(paths, dcmVodSettingsPath)

	// dcm/uploadRepository
	dcmUploadRepositoryPath := r.PathPrefix("/xconfAdminService/dcm/uploadRepository").Subrouter()
	dcmUploadRepositoryPath.HandleFunc("", dcm.GetLogRepoSettingsHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("", dcm.CreateLogRepoSettingsHandler).Methods("POST").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("", dcm.UpdateLogRepoSettingsHandler).Methods("PUT").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/entities", dcm.PostLogRepoSettingsEntitiesHandler).Methods("POST").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/entities", dcm.PutLogRepoSettingsEntitiesHandler).Methods("PUT").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/size", dcm.GetLogRepoSettingsSizeHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/names", dcm.GetLogRepoSettingsNamesHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/filtered", dcm.PostLogRepoSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/{id}", dcm.DeleteLogRepoSettingsByIdHandler).Methods("DELETE").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/{id}", dcm.GetLogRepoSettingsByIdHandler).Methods("GET").Name("DCM-UploadRepository")
	paths = append(paths, dcmUploadRepositoryPath)

	// dcm/logUploadSettings
	dcmLogUploadSettingsPath := r.PathPrefix("/xconfAdminService/dcm/logUploadSettings").Subrouter()
	dcmLogUploadSettingsPath.HandleFunc("", dcm.GetLogUploadSettingsHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("", dcm.CreateLogUploadSettingsHandler).Methods("POST").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("", dcm.UpdateLogUploadSettingsHandler).Methods("PUT").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/size", dcm.GetLogUploadSettingsSizeHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/names", dcm.GetLogUploadSettingsNamesHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/filtered", dcm.PostLogUploadSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/export", dcm.GetLogRepoSettingsExportHandler).Methods("GET").Name("DCM-LogUploadSettings")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths)
	dcmLogUploadSettingsPath.HandleFunc("/{id}", dcm.DeleteLogUploadSettingsByIdHandler).Methods("DELETE").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/{id}", dcm.GetLogUploadSettingsByIdHandler).Methods("GET").Name("DCM-LogUploadSettings")
	paths = append(paths, dcmLogUploadSettingsPath)

	// dcm/testpage
	dcmTestpagePath := r.PathPrefix("/xconfAdminService/dcm/testpage").Subrouter()
	dcmTestpagePath.HandleFunc("", dcm.DcmTestPageHandler).Methods("POST").Name("DCM-TestPage")
	paths = append(paths, dcmTestpagePath)

	// telemetry
	telemetryPath := r.PathPrefix("/xconfAdminService/telemetry").Subrouter()
	telemetryPath.HandleFunc("/create/{contextAttributeName}/{expectedValue}", telemetry.CreateTelemetryEntryFor).Methods("POST").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/testpage", telemetry.TelemetryTestPageHandler).Methods("POST").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/drop/{contextAttributeName}/{expectedValue}", telemetry.DropTelemetryEntryFor).Methods("POST").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/getAvailableRuleDescriptors", telemetry.GetDescriptors).Methods("GET").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/getAvailableTelemetryDescriptors", telemetry.GetTelemetryDescriptors).Methods("GET").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/addTo/{ruleId}/{contextAttributeName}/{expectedValue}/{expires}", telemetry.TempAddToPermanentRule).Methods("POST").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/bindToTelemetry/{telemetryId}/{contextAttributeName}/{expectedValue}/{expires}", telemetry.BindToTelemetry).Methods("POST").Name("Telemetry1-Uncategorized")
	paths = append(paths, telemetryPath)

	// telemetry/profile
	telemetryProfilePath := r.PathPrefix("/xconfAdminService/telemetry/profile").Subrouter()
	telemetryProfilePath.HandleFunc("", change.GetTelemetryProfilesHandler).Methods("GET").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("", change.CreateTelemetryProfileHandler).Methods("POST").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("", change.UpdateTelemetryProfileHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/change", change.CreateTelemetryProfileChangeHandler).Methods("POST").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/change", change.UpdateTelemetryProfileChangeHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/{id}", change.DeleteTelemetryProfileHandler).Methods("DELETE").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/change/{id}", change.DeleteTelemetryProfileChangeHandler).Methods("DELETE").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/{id}", change.GetTelemetryProfileByIdHandler).Methods("GET").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/entities", change.PostTelemetryProfileEntitiesHandler).Methods("POST").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/entities", change.PutTelemetryProfileEntitiesHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/filtered", change.PostTelemetryProfileFilteredHandler).Methods("POST").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/migrate/createTelemetryId", change.CreateTelemetryIdsHandler).Methods("GET").Name("Telemetry1-Profiles") //can be removed
	telemetryProfilePath.HandleFunc("/entry/add/{id}", change.AddTelemetryProfileEntryHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/entry/remove/{id}", change.RemoveTelemetryProfileEntryHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/change/entry/add/{id}", change.AddTelemetryProfileEntryChangeHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/change/entry/remove/{id}", change.RemoveTelemetryProfileEntryChangeHandler).Methods("PUT").Name("Telemetry1-Profiles")

	paths = append(paths, telemetryProfilePath)

	// telemetry/rule
	telemetryRulePath := r.PathPrefix("/xconfAdminService/telemetry/rule").Subrouter()
	telemetryRulePath.HandleFunc("", telemetry.GetTelemetryRulesHandler).Methods("GET").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("", telemetry.CreateTelemetryRuleHandler).Methods("POST").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("", telemetry.UpdateTelemetryRuleHandler).Methods("PUT").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("/entities", telemetry.PostTelemtryRuleEntitiesHandler).Methods("POST").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("/entities", telemetry.PutTelemetryRuleEntitiesHandler).Methods("PUT").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("/filtered", telemetry.PostTelemetryRuleFilteredWithParamsHandler).Methods("POST").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("/{id}", telemetry.DeleteTelmetryRuleByIdHandler).Methods("DELETE").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("/{id}", telemetry.GetTelemetryRuleByIdHandler).Methods("GET").Name("Telemetry1-Rules")
	paths = append(paths, telemetryRulePath)

	// telemetry/v2/profile
	telemetryV2ProfilePath := r.PathPrefix("/xconfAdminService/telemetry/v2/profile").Subrouter()
	telemetryV2ProfilePath.HandleFunc("", change.GetTelemetryTwoProfilesHandler).Methods("GET").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("", change.CreateTelemetryTwoProfileHandler).Methods("POST").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("", change.UpdateTelemetryTwoProfileHandler).Methods("PUT").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/{id}", change.DeleteTelemetryTwoProfileHandler).Methods("DELETE").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/change", change.CreateTelemetryTwoProfileChangeHandler).Methods("POST").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/change", change.UpdateTelemetryTwoProfileChangeHandler).Methods("PUT").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/change/{id}", change.DeleteTelemetryTwoProfileChangeHandler).Methods("DELETE").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/{id}", change.GetTelemetryTwoProfileByIdHandler).Methods("GET").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/byIdList", change.PostTelemetryTwoProfilesByIdListHandler).Methods("POST").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/entities", change.PostTelemetryTwoProfileEntitiesHandler).Methods("POST").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/entities", change.PutTelemetryTwoProfileEntitiesHandler).Methods("PUT").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/filtered", change.PostTelemetryTwoProfileFilteredHandler).Methods("POST").Name("Telemetry2-Profiles")
	paths = append(paths, telemetryV2ProfilePath)

	// telemetry/v2/rule
	telemetryV2RulePath := r.PathPrefix("/xconfAdminService/telemetry/v2/rule").Subrouter()
	telemetryV2RulePath.HandleFunc("", telemetry.CreateTelemetryTwoRuleHandler).Methods("POST").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/entities", telemetry.CreateTelemetryTwoRulesPackageHandler).Methods("POST").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("", telemetry.UpdateTelemetryTwoRuleHandler).Methods("PUT").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/entities", telemetry.UpdateTelemetryTwoRulesPackageHandler).Methods("PUT").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("", telemetry.GetTelemetryTwoRulesAllExport).Methods("GET").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/page", queries.NotImplementedHandler).Methods("GET").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/{id}", telemetry.GetTelemetryTwoRuleById).Methods("GET").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/filtered", telemetry.GetTelemetryTwoRulesFilteredWithPage).Methods("POST").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/{id}", telemetry.DeleteOneTelemetryTwoRuleHandler).Methods("DELETE").Name("Telemetry2-Rules")
	paths = append(paths, telemetryV2RulePath)

	// telemetry/v2/testpage
	teleV2TestpagePath := r.PathPrefix("/xconfAdminService/telemetry/v2/testpage").Subrouter()
	teleV2TestpagePath.HandleFunc("", change.TelemetryTwoTestPageHandler).Methods("POST").Name("Telemetry2-Uncategorized")
	paths = append(paths, teleV2TestpagePath)

	// change - these are the same as telemetry/change APIs which are needed
	// to be compatible w/ Java AS; eventually these will be deprecated.
	changePath := r.PathPrefix("/xconfAdminService/change").Subrouter()
	changePath.HandleFunc("/all", change.GetProfileChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/approved", change.GetApprovedHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/approve/{changeId}", change.ApproveChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/revert/{approveId}", change.RevertChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/cancel/{changeId}", change.CancelChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/changes/grouped/byId", change.GetGroupedChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/approved/grouped/byId", change.GetGroupedApprovedChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/entityIds", change.GetChangedEntityIdsHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/approveChanges", change.ApproveChangesHandler).Methods("POST").Name("Telemetry1-Changes") //TODO verify usages
	changePath.HandleFunc("/revertChanges", change.RevertChangesHandler).Methods("POST").Name("Telemetry1-Changes")
	changePath.HandleFunc("/approved/filtered", change.GetApprovedFilteredHandler).Methods("POST").Name("Telemetry1-Changes")
	changePath.HandleFunc("/changes/filtered", change.GetChangesFilteredHandler).Methods("POST").Name("Telemetry1-Changes")
	paths = append(paths, changePath)

	// telemetry/change
	telemetryChangePath := r.PathPrefix("/xconfAdminService/telemetry/change").Subrouter()
	telemetryChangePath.HandleFunc("/all", change.GetProfileChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/approved", change.GetApprovedHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/approve/{changeId}", change.ApproveChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/revert/{approveId}", change.RevertChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/cancel/{changeId}", change.CancelChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/changes/grouped/byId", change.GetGroupedChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/approved/grouped/byId", change.GetGroupedApprovedChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/entityIds", change.GetChangedEntityIdsHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/approveChanges", change.ApproveChangesHandler).Methods("POST").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/revertChanges", change.RevertChangesHandler).Methods("POST").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/approved/filtered", change.GetApprovedFilteredHandler).Methods("POST").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/changes/filtered", change.GetChangesFilteredHandler).Methods("POST").Name("Telemetry1-Changes")
	paths = append(paths, telemetryChangePath)

	// telemetry/v2/change
	telemetryTwoChangePath := r.PathPrefix("/xconfAdminService/telemetry/v2/change").Subrouter()
	telemetryTwoChangePath.HandleFunc("/all", change.GetTwoProfileChangesHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/approved", change.GetApprovedTwoChangesHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/approve/{changeId}", change.ApproveTwoChangeHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/revert/{approveId}", change.RevertTwoChangeHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/cancel/{changeId}", change.CancelTwoChangeHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/entityIds", change.GetTwoChangeEntityIdsHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/changes/grouped/byId", change.GetGroupedTwoChangesHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/approved/grouped/byId", change.GetGroupedApprovedTwoChangesHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/approveChanges", change.ApproveTwoChangesHandler).Methods("POST").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/revertChanges", change.RevertTwoChangesHandler).Methods("POST").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/approved/filtered", change.GetApprovedTwoChangesFilteredHandler).Methods("POST").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/changes/filtered", change.GetTwoChangesFilteredHandler).Methods("POST").Name("Telemetry2-Changes")
	paths = append(paths, telemetryTwoChangePath)

	// changelog
	changelogPath := r.PathPrefix("/xconfAdminService/changelog").Subrouter()
	changelogPath.HandleFunc("", queries.GetChangeLogForTheDay).Methods("GET").Name("General-Uncategorized")
	paths = append(paths, changelogPath)

	// log
	logPath := r.PathPrefix("/xconfAdminService/log").Subrouter()
	logPath.HandleFunc("/{macStr}", queries.GetLogs).Methods("GET").Name("Firmware-Logs")
	paths = append(paths, logPath)

	// reportpage
	reportpagePath := r.PathPrefix("/xconfAdminService/reportpage").Subrouter()
	reportpagePath.HandleFunc("", queries.PostFirmwareRuleReportPageHandler).Methods("POST").Name("Firmware-ReportPage")
	paths = append(paths, reportpagePath)

	// stats
	statsPath := r.PathPrefix("/xconfAdminService/stats").Subrouter()
	statsPath.HandleFunc("", queries.GetStats).Methods("GET").Name("Statistics")
	statsPath.HandleFunc("/cache/reloadAll", queries.GetInfoRefreshAllHandler).Methods("GET").Name("Statistics")
	statsPath.HandleFunc("/cache/{tableName}/reload", queries.GetInfoRefreshHandler).Methods("GET").Name("Statistics")
	paths = append(paths, statsPath)

	// migration
	migrationPath := r.PathPrefix("/xconfAdminService/migration/info").Subrouter()
	migrationPath.HandleFunc("", queries.GetMigrationInfoHandler).Methods("GET").Name("Migration")
	paths = append(paths, migrationPath)

	// appsettings
	appsettingsPath := r.PathPrefix("/xconfAdminService/appsettings").Subrouter()
	appsettingsPath.HandleFunc("", queries.GetAppSettings).Methods("GET").Name("AppSettings")
	appsettingsPath.HandleFunc("", queries.UpdateAppSettings).Methods("PUT").Name("AppSettings")
	paths = append(paths, appsettingsPath)

	// lockdownSettings
	lockdownsettingsPath := r.PathPrefix("/xconfAdminService/lockdownsettings").Subrouter()
	lockdownsettingsPath.HandleFunc("", lockdown.GetLockdownSettingsHandler).Methods("GET").Name("LockdownSettings")
	lockdownsettingsPath.HandleFunc("", lockdown.PutLockdownSettingsHandler).Methods("PUT").Name("LockdownSettings")
	paths = append(paths, lockdownsettingsPath)

	//lockdown during recooking
	xcrpRecookingPath := r.PathPrefix("/xconfAdminService/rfc/recooking").Subrouter()
	xcrpRecookingPath.HandleFunc("", xcrp.PostRecookingLockdownSettingsHandler).Methods("POST").Name("Recooking")
	paths = append(paths, xcrpRecookingPath)

	//get recooking status. Disabled for now
	//xcrpRecookingStatusPath := r.PathPrefix("/xconfAdminService/rfc/recooking/status").Subrouter()
	//xcrpRecookingStatusPath.HandleFunc("", xcrp.GetRecookingStatusHandler).Methods("GET").Name("RecookingStatus")
	//xcrpRecookingStatusPath.HandleFunc("/details", xcrp.GetRecookingStatusDetailsHandler).Methods("GET").Name("RecookingStatusDetails")
	//paths = append(paths, xcrpRecookingStatusPath)

	// canarysettings
	canarysettingsPath := r.PathPrefix("/xconfAdminService/canarysettings").Subrouter()
	canarysettingsPath.HandleFunc("", canary.GetCanarySettingsHandler).Methods("GET").Name("CanarySettings")
	canarysettingsPath.HandleFunc("", canary.PutCanarySettingsHandler).Methods("PUT").Name("CanarySettings")
	paths = append(paths, canarysettingsPath)

	// penetration data report
	penetrationPath := r.PathPrefix("/xconfAdminService/penetrationdata").Subrouter()
	penetrationPath.HandleFunc("/{macAddress}", queries.GetPenetrationDataByMacHandler).Methods("GET").Name("PenetrationData")
	paths = append(paths, penetrationPath)

	// rules configuration
	macipruleconfigPath := r.PathPrefix("/xconfAdminService/config").Subrouter()
	macipruleconfigPath.HandleFunc("/maciprule", ipmacrule.GetIpMacRuleConfigurationHandler).Methods("GET").Name("Mac-Ip-RuleConfig")
	paths = append(paths, macipruleconfigPath)

	// api to create wakeuppool manually
	wakeupPoolCreationPath := r.PathPrefix("/xconfAdminService/wakeuppool").Subrouter()
	wakeupPoolCreationPath.HandleFunc("", queries.CreateWakeupPoolHandler).Methods("POST").Name("createwakeuppool")
	authPaths = append(authPaths, wakeupPoolCreationPath)

	// CORS
	c := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"X-Requested-With", "Origin", "Content-Type", "Accept", "Authorization", "token"},
	})

	for _, p := range authPaths {
		p.Use(c.Handler)
		p.Use(s.XW_XconfServer.SpanMiddleware)
		p.Use(s.XW_XconfServer.NoAuthMiddleware)
	}

	for _, p := range paths {
		p.Use(c.Handler)
		if !s.TestOnly() {
			p.Use(s.XW_XconfServer.SpanMiddleware)
			p.Use(s.AuthValidationMiddleware)
		} else {
			p.Use(s.XW_XconfServer.NoAuthMiddleware)
		}
	}
}
