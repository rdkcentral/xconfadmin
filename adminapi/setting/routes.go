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
package setting

import (
	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/http"
)

// RegisterSettingsRoutes registers the /xconfAdminService/setting prefix routes
func RegisterSettingsRoutes(router *mux.Router, paths []*mux.Router) []*mux.Router {
	// setting/profile
	settingProfilePath := router.PathPrefix("/xconfAdminService/setting/profile").Subrouter()
	settingProfilePath.HandleFunc("", CreateSettingProfileHandler).Methods("POST").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/entities", CreateSettingProfilesPackageHandler).Methods("POST").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("", UpdateSettingProfilesHandler).Methods("PUT").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/entities", UpdateSettingProfilesPackageHandler).Methods("PUT").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("", GetSettingProfilesAllExport).Methods("GET").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/page", http.NotImplementedHandler).Methods("GET").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/{id}", GetSettingProfileOneExport).Methods("GET").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/filtered", GetSettingProfilesFilteredWithPage).Methods("POST").Name("Settings-Profiles")
	settingProfilePath.HandleFunc("/{id}", DeleteOneSettingProfilesHandler).Methods("DELETE").Name("Settings-Profiles")
	paths = append(paths, settingProfilePath)

	// setting/rule
	settingRulePath := router.PathPrefix("/xconfAdminService/setting/rule").Subrouter()
	settingRulePath.HandleFunc("", CreateSettingRuleHandler).Methods("POST").Name("Settings-Rules")
	settingRulePath.HandleFunc("/entities", CreateSettingRulesPackageHandler).Methods("POST").Name("Settings-Rules")
	settingRulePath.HandleFunc("", UpdateSettingRulesHandler).Methods("PUT").Name("Settings-Rules")
	settingRulePath.HandleFunc("/entities", UpdateSettingRulesPackageHandler).Methods("PUT").Name("Settings-Rules")
	settingRulePath.HandleFunc("", GetSettingRulesAllExport).Methods("GET").Name("Settings-Rules")
	settingRulePath.HandleFunc("/page", http.NotImplementedHandler).Methods("GET").Name("Settings-Rules")
	settingRulePath.HandleFunc("/{id}", GetSettingRuleOneExport).Methods("GET").Name("Settings-Rules")
	settingRulePath.HandleFunc("/filtered", GetSettingRulesFilteredWithPage).Methods("POST").Name("Settings-Rules")
	settingRulePath.HandleFunc("/{id}", DeleteOneSettingRulesHandler).Methods("DELETE").Name("Settings-Rules")
	paths = append(paths, settingRulePath)

	// settings/testpage
	settingTestpagePath := router.PathPrefix("/xconfAdminService/settings/testpage").Subrouter()
	settingTestpagePath.HandleFunc("", SettingTestPageHandler).Methods("POST").Name("Settings-TestPage")
	paths = append(paths, settingTestpagePath)

	return paths
}
