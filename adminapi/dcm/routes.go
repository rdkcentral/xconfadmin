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
	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/http"
)

// RegisterDCMRoutes registers the /xconfAdminService/dcm prefix routes
func RegisterDCMRoutes(router *mux.Router, paths []*mux.Router) []*mux.Router {
	// dcm/formula
	dcmFormulaPath := router.PathPrefix("/xconfAdminService/dcm/formula").Subrouter()
	dcmFormulaPath.HandleFunc("", GetDcmFormulaHandler).Methods("GET").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("", CreateDcmFormulaHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("", UpdateDcmFormulaHandler).Methods("PUT").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/page", http.NotImplementedHandler).Methods("GET").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/entities", PostDcmFormulaListHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/entities", PutDcmFormulaListHandler).Methods("PUT").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/list", PostDcmFormulaListHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/list", PutDcmFormulaListHandler).Methods("PUT").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/settingsAvailability", DcmFormulaSettingsAvailabilitygHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/import/{overwrite}", ImportDcmFormulaWithOverwriteHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/import", ImportDcmFormulasHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/formulasAvailability", DcmFormulasAvailabilitygHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/size", GetDcmFormulaSizeHandler).Methods("GET").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/names", GetDcmFormulaNamesHandler).Methods("GET").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/filtered", PostDcmFormulaFilteredWithParamsHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/{id}/priority/{newPriority}", DcmFormulaChangePriorityHandler).Methods("POST").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/{id}", DeleteDcmFormulaByIdHandler).Methods("DELETE").Name("DCM-Formulas")
	dcmFormulaPath.HandleFunc("/{id}", GetDcmFormulaByIdHandler).Methods("GET").Name("DCM-Formulas")
	paths = append(paths, dcmFormulaPath)

	// dcm/deviceSettings
	dcmDeviceSettingsPath := router.PathPrefix("/xconfAdminService/dcm/deviceSettings").Subrouter()
	dcmDeviceSettingsPath.HandleFunc("", GetDeviceSettingsHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("", CreateDeviceSettingsHandler).Methods("POST").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("", UpdateDeviceSettingsHandler).Methods("PUT").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/page", http.NotImplementedHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/size", GetDeviceSettingsSizeHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/names", GetDeviceSettingsNamesHandler).Methods("GET").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/filtered", PostDeviceSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/export", GetDeviceSettingsExportHandler).Methods("GET")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	dcmDeviceSettingsPath.HandleFunc("/{id}", DeleteDeviceSettingsByIdHandler).Methods("DELETE").Name("DCM-DeviceSettings")
	dcmDeviceSettingsPath.HandleFunc("/{id}", GetDeviceSettingsByIdHandler).Methods("GET").Name("DCM-DeviceSettings")
	paths = append(paths, dcmDeviceSettingsPath)

	// dcm/vodsettings
	dcmVodSettingsPath := router.PathPrefix("/xconfAdminService/dcm/vodsettings").Subrouter()
	dcmVodSettingsPath.HandleFunc("", GetVodSettingsHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("", CreateVodSettingsHandler).Methods("POST").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("", UpdateVodSettingsHandler).Methods("PUT").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/page", http.NotImplementedHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/size", GetVodSettingsSizeHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/names", GetVodSettingsNamesHandler).Methods("GET").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/filtered", PostVodSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/export", GetVodSettingExportHandler).Methods("GET").Name("DCM-VODSettings")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths
	dcmVodSettingsPath.HandleFunc("/{id}", DeleteVodSettingsByIdHandler).Methods("DELETE").Name("DCM-VODSettings")
	dcmVodSettingsPath.HandleFunc("/{id}", GetVodSettingsByIdHandler).Methods("GET").Name("DCM-VODSettings")
	paths = append(paths, dcmVodSettingsPath)

	// dcm/uploadRepository
	dcmUploadRepositoryPath := router.PathPrefix("/xconfAdminService/dcm/uploadRepository").Subrouter()
	dcmUploadRepositoryPath.HandleFunc("", GetLogRepoSettingsHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("", CreateLogRepoSettingsHandler).Methods("POST").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("", UpdateLogRepoSettingsHandler).Methods("PUT").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/page", http.NotImplementedHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/entities", PostLogRepoSettingsEntitiesHandler).Methods("POST").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/entities", PutLogRepoSettingsEntitiesHandler).Methods("PUT").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/size", GetLogRepoSettingsSizeHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/names", GetLogRepoSettingsNamesHandler).Methods("GET").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/filtered", PostLogRepoSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/{id}", DeleteLogRepoSettingsByIdHandler).Methods("DELETE").Name("DCM-UploadRepository")
	dcmUploadRepositoryPath.HandleFunc("/{id}", GetLogRepoSettingsByIdHandler).Methods("GET").Name("DCM-UploadRepository")
	paths = append(paths, dcmUploadRepositoryPath)

	// dcm/logUploadSettings
	dcmLogUploadSettingsPath := router.PathPrefix("/xconfAdminService/dcm/logUploadSettings").Subrouter()
	dcmLogUploadSettingsPath.HandleFunc("", GetLogUploadSettingsHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("", CreateLogUploadSettingsHandler).Methods("POST").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("", UpdateLogUploadSettingsHandler).Methods("PUT").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/page", http.NotImplementedHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/size", GetLogUploadSettingsSizeHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/names", GetLogUploadSettingsNamesHandler).Methods("GET").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/filtered", PostLogUploadSettingsFilteredWithParamsHandler).Methods("POST").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/export", GetLogRepoSettingsExportHandler).Methods("GET").Name("DCM-LogUploadSettings")
	// url with var has to be placed last otherwise, it gets confused with url with defined paths)
	dcmLogUploadSettingsPath.HandleFunc("/{id}", DeleteLogUploadSettingsByIdHandler).Methods("DELETE").Name("DCM-LogUploadSettings")
	dcmLogUploadSettingsPath.HandleFunc("/{id}", GetLogUploadSettingsByIdHandler).Methods("GET").Name("DCM-LogUploadSettings")
	paths = append(paths, dcmLogUploadSettingsPath)

	// dcm/testpage
	dcmTestpagePath := router.PathPrefix("/xconfAdminService/dcm/testpage").Subrouter()
	dcmTestpagePath.HandleFunc("", DcmTestPageHandler).Methods("POST").Name("DCM-TestPage")
	paths = append(paths, dcmTestpagePath)

	return paths
}
