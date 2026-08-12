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
package change

import (
	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/http"
)

// RegisterChangeRoutes registers the /xconfAdminService/telemetry/profile prefix routes
func RegisterChangeRoutes(router *mux.Router, paths []*mux.Router) []*mux.Router {
	// telemetry/profile
	telemetryProfilePath := router.PathPrefix("/xconfAdminService/telemetry/profile").Subrouter()
	telemetryProfilePath.HandleFunc("", GetTelemetryProfilesHandler).Methods("GET").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("", CreateTelemetryProfileHandler).Methods("POST").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("", UpdateTelemetryProfileHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/change", CreateTelemetryProfileChangeHandler).Methods("POST").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/change", UpdateTelemetryProfileChangeHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/{id}", DeleteTelemetryProfileHandler).Methods("DELETE").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/change/{id}", DeleteTelemetryProfileChangeHandler).Methods("DELETE").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/{id}", GetTelemetryProfileByIdHandler).Methods("GET").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/page", http.NotImplementedHandler).Methods("GET").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/entities", PostTelemetryProfileEntitiesHandler).Methods("POST").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/entities", PutTelemetryProfileEntitiesHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/filtered", PostTelemetryProfileFilteredHandler).Methods("POST").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/migrate/createTelemetryId", CreateTelemetryIdsHandler).Methods("GET").Name("Telemetry1-Profiles") //can be removed
	telemetryProfilePath.HandleFunc("/entry/add/{id}", AddTelemetryProfileEntryHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/entry/remove/{id}", RemoveTelemetryProfileEntryHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/change/entry/add/{id}", AddTelemetryProfileEntryChangeHandler).Methods("PUT").Name("Telemetry1-Profiles")
	telemetryProfilePath.HandleFunc("/change/entry/remove/{id}", RemoveTelemetryProfileEntryChangeHandler).Methods("PUT").Name("Telemetry1-Profiles")

	// telemetry/v2/profile
	telemetryV2ProfilePath := router.PathPrefix("/xconfAdminService/telemetry/v2/profile").Subrouter()
	telemetryV2ProfilePath.HandleFunc("", GetTelemetryTwoProfilesHandler).Methods("GET").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("", CreateTelemetryTwoProfileHandler).Methods("POST").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("", UpdateTelemetryTwoProfileHandler).Methods("PUT").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/{id}", DeleteTelemetryTwoProfileHandler).Methods("DELETE").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/change", CreateTelemetryTwoProfileChangeHandler).Methods("POST").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/change", UpdateTelemetryTwoProfileChangeHandler).Methods("PUT").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/change/{id}", DeleteTelemetryTwoProfileChangeHandler).Methods("DELETE").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/{id}", GetTelemetryTwoProfileByIdHandler).Methods("GET").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/page", http.NotImplementedHandler).Methods("GET").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/byIdList", PostTelemetryTwoProfilesByIdListHandler).Methods("POST").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/entities", PostTelemetryTwoProfileEntitiesHandler).Methods("POST").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/entities", PutTelemetryTwoProfileEntitiesHandler).Methods("PUT").Name("Telemetry2-Profiles")
	telemetryV2ProfilePath.HandleFunc("/filtered", PostTelemetryTwoProfileFilteredHandler).Methods("POST").Name("Telemetry2-Profiles")
	paths = append(paths, telemetryV2ProfilePath)

	// telemetry/v2/testpage
	teleV2TestpagePath := router.PathPrefix("/xconfAdminService/telemetry/v2/testpage").Subrouter()
	teleV2TestpagePath.HandleFunc("", TelemetryTwoTestPageHandler).Methods("POST").Name("Telemetry2-Uncategorized")
	paths = append(paths, teleV2TestpagePath)

	// change - these are the same as telemetry/change APIs which are needed
	// to be compatible w/ Java AS; eventually these will be deprecated.
	changePath := router.PathPrefix("/xconfAdminService/change").Subrouter()
	changePath.HandleFunc("/all", GetProfileChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/approved", GetApprovedHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/approve/{changeId}", ApproveChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/revert/{approveId}", RevertChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/cancel/{changeId}", CancelChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/changes/grouped/byId", GetGroupedChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/approved/grouped/byId", GetGroupedApprovedChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/entityIds", GetChangedEntityIdsHandler).Methods("GET").Name("Telemetry1-Changes")
	changePath.HandleFunc("/approveChanges", ApproveChangesHandler).Methods("POST").Name("Telemetry1-Changes") //TODO verify usages
	changePath.HandleFunc("/revertChanges", RevertChangesHandler).Methods("POST").Name("Telemetry1-Changes")
	changePath.HandleFunc("/approved/filtered", GetApprovedFilteredHandler).Methods("POST").Name("Telemetry1-Changes")
	changePath.HandleFunc("/changes/filtered", GetChangesFilteredHandler).Methods("POST").Name("Telemetry1-Changes")
	paths = append(paths, changePath)

	// telemetry/change
	telemetryChangePath := router.PathPrefix("/xconfAdminService/telemetry/change").Subrouter()
	telemetryChangePath.HandleFunc("/all", GetProfileChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/approved", GetApprovedHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/approve/{changeId}", ApproveChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/revert/{approveId}", RevertChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/cancel/{changeId}", CancelChangeHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/changes/grouped/byId", GetGroupedChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/approved/grouped/byId", GetGroupedApprovedChangesHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/entityIds", GetChangedEntityIdsHandler).Methods("GET").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/approveChanges", ApproveChangesHandler).Methods("POST").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/revertChanges", RevertChangesHandler).Methods("POST").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/approved/filtered", GetApprovedFilteredHandler).Methods("POST").Name("Telemetry1-Changes")
	telemetryChangePath.HandleFunc("/changes/filtered", GetChangesFilteredHandler).Methods("POST").Name("Telemetry1-Changes")
	paths = append(paths, telemetryChangePath)

	// telemetry/v2/change
	telemetryTwoChangePath := router.PathPrefix("/xconfAdminService/telemetry/v2/change").Subrouter()
	telemetryTwoChangePath.HandleFunc("/all", GetTwoProfileChangesHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/approved", GetApprovedTwoChangesHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/approve/{changeId}", ApproveTwoChangeHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/revert/{approveId}", RevertTwoChangeHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/cancel/{changeId}", CancelTwoChangeHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/entityIds", GetTwoChangeEntityIdsHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/changes/grouped/byId", GetGroupedTwoChangesHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/approved/grouped/byId", GetGroupedApprovedTwoChangesHandler).Methods("GET").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/approveChanges", ApproveTwoChangesHandler).Methods("POST").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/revertChanges", RevertTwoChangesHandler).Methods("POST").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/approved/filtered", GetApprovedTwoChangesFilteredHandler).Methods("POST").Name("Telemetry2-Changes")
	telemetryTwoChangePath.HandleFunc("/changes/filtered", GetTwoChangesFilteredHandler).Methods("POST").Name("Telemetry2-Changes")
	paths = append(paths, telemetryTwoChangePath)

	return paths
}
