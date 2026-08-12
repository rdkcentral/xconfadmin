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
package telemetry

import (
	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/http"
)

// RegisterTelemetryRoutes registers the /xconfAdminService/telemetry prefix routes
func RegisterTelemetryRoutes(router *mux.Router, paths []*mux.Router) []*mux.Router {
	// telemetry
	telemetryPath := router.PathPrefix("/xconfAdminService/telemetry").Subrouter()
	telemetryPath.HandleFunc("/create/{contextAttributeName}/{expectedValue}", CreateTelemetryEntryFor).Methods("POST").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/testpage", TelemetryTestPageHandler).Methods("POST").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/drop/{contextAttributeName}/{expectedValue}", DropTelemetryEntryFor).Methods("POST").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/getAvailableRuleDescriptors", GetDescriptors).Methods("GET").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/getAvailableTelemetryDescriptors", GetTelemetryDescriptors).Methods("GET").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/addTo/{ruleId}/{contextAttributeName}/{expectedValue}/{expires}", TempAddToPermanentRule).Methods("POST").Name("Telemetry1-Uncategorized")
	telemetryPath.HandleFunc("/bindToTelemetry/{telemetryId}/{contextAttributeName}/{expectedValue}/{expires}", BindToTelemetry).Methods("POST").Name("Telemetry1-Uncategorized")
	paths = append(paths, telemetryPath)

	// telemetry/rule
	telemetryRulePath := router.PathPrefix("/xconfAdminService/telemetry/rule").Subrouter()
	telemetryRulePath.HandleFunc("", GetTelemetryRulesHandler).Methods("GET").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("", CreateTelemetryRuleHandler).Methods("POST").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("", UpdateTelemetryRuleHandler).Methods("PUT").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("/entities", PostTelemtryRuleEntitiesHandler).Methods("POST").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("/entities", PutTelemetryRuleEntitiesHandler).Methods("PUT").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("/filtered", PostTelemetryRuleFilteredWithParamsHandler).Methods("POST").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("/{id}", DeleteTelmetryRuleByIdHandler).Methods("DELETE").Name("Telemetry1-Rules")
	telemetryRulePath.HandleFunc("/{id}", GetTelemetryRuleByIdHandler).Methods("GET").Name("Telemetry1-Rules")
	paths = append(paths, telemetryRulePath)

	// telemetry/v2/rule
	telemetryV2RulePath := router.PathPrefix("/xconfAdminService/telemetry/v2/rule").Subrouter()
	telemetryV2RulePath.HandleFunc("", CreateTelemetryTwoRuleHandler).Methods("POST").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/entities", CreateTelemetryTwoRulesPackageHandler).Methods("POST").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("", UpdateTelemetryTwoRuleHandler).Methods("PUT").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/entities", UpdateTelemetryTwoRulesPackageHandler).Methods("PUT").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("", GetTelemetryTwoRulesAllExport).Methods("GET").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/page", http.NotImplementedHandler).Methods("GET").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/{id}", GetTelemetryTwoRuleById).Methods("GET").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/filtered", GetTelemetryTwoRulesFilteredWithPage).Methods("POST").Name("Telemetry2-Rules")
	telemetryV2RulePath.HandleFunc("/{id}", DeleteOneTelemetryTwoRuleHandler).Methods("DELETE").Name("Telemetry2-Rules")
	paths = append(paths, telemetryV2RulePath)

	return paths
}
