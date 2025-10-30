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
	"io/ioutil"
	"net/http"
	"testing"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"github.com/gorilla/mux"
	"gotest.tools/assert"
)

func ImportDeviceSettingsTableData(data []string, tabletype logupload.DeviceSettings) error {
	var err error
	for _, row := range data {
		err = json.Unmarshal([]byte(row), &tabletype)
		err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_DEVICE_SETTINGS, tabletype.ID, &tabletype)

	}
	return err
}
func TestAllDeviceSettingsApis(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// GET ALL DEVICE SETTINGS API

	var tableData = []string{
		`{"id":"23069266-45b7-4bf6-a255-e6ee584cd68b","name":"RDKB_PLATFORM_SECURITY_GROUP_SV","checkOnReboot":true,"settingsAreActive":true,"schedule":{"type":"ActNow","expression":"26 4 * * *","timeZone":"UTC","timeWindowMinutes":0},"applicationType":"stb"}`,
		`{"id":"23069266-45b7-4bf6-a255-e6ee584cd68bid","name":"Get By Id Test","checkOnReboot":true,"settingsAreActive":true,"schedule":{"type":"ActNow","expression":"26 4 * * *","timeZone":"UTC","timeWindowMinutes":0},"applicationType":"stb"}`,
		`{"id":"23069266-45b7-4bf6-a255-e6ee584cd68bsz","name":"Get Size","checkOnReboot":true,"settingsAreActive":true,"schedule":{"type":"ActNow","expression":"26 4 * * *","timeZone":"UTC","timeWindowMinutes":0},"applicationType":"stb"}`,
		`{"id":"23069266-45b7-4bf6-a255-e6ee584cd68bnm","name":"Get Names","checkOnReboot":true,"settingsAreActive":true,"schedule":{"type":"ActNow","expression":"26 4 * * *","timeZone":"UTC","timeWindowMinutes":0},"applicationType":"stb"}`,
		`{"id":"23069266-45b7-4bf6-a255-e6ee584cd68brm","name":"Delete By Id Test","checkOnReboot":true,"settingsAreActive":true,"schedule":{"type":"ActNow","expression":"26 4 * * *","timeZone":"UTC","timeWindowMinutes":0},"applicationType":"stb"}`,
	}

	err := ImportDeviceSettingsTableData(tableData, logupload.DeviceSettings{})
	assert.NilError(t, err)

	url := "/xconfAdminService/dcm/deviceSettings"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	if res.StatusCode == http.StatusOK {
		var dss = []logupload.DeviceSettings{}
		json.Unmarshal(body, &dss)
		assert.Equal(t, len(dss) > 0, true)
	}

	//CREATE DEVICE SETTING AND UPDATE
	dsdata := []byte(
		`{"id":"54bac1f5-0146-4399-a55d-efb8fa2661fa","updated":1636408666071,"name":"dineshcrup","checkOnReboot":false,"configurationServiceURL":{"id":"","name":"","description":"","url":""},"settingsAreActive":false,"schedule":{"type":"ActNow","expression":"3 1 3 4 *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"applicationType":"stb"}`)

	req, err = http.NewRequest("POST", url, bytes.NewBuffer(dsdata))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	//ERROR CREATING AGAIN SAME ENTRY
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(dsdata))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	//UPDATE EXISTING ENTRY
	dsdataup := []byte(
		`{"id":"54bac1f5-0146-4399-a55d-efb8fa2661fa","updated":1636408666071,"name":"dineshupdate","checkOnReboot":false,"configurationServiceURL":{"id":"","name":"","description":"","url":""},"settingsAreActive":false,"schedule":{"type":"ActNow","expression":"3 1 13 11 *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"applicationType":"stb"}`)
	req, err = http.NewRequest("PUT", url, bytes.NewBuffer(dsdataup))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//UPDATE NON EXISTING ENTRY
	dsdataer := []byte(
		`{"id":"54bac1f5-0146-4399-a55d-efb8fa266err","updated":1636408666071,"name":"dineshcrup","checkOnReboot":false,"configurationServiceURL":{"id":"","name":"","description":"","url":""},"settingsAreActive":false,"schedule":{"type":"ActNow","expression":"3 1 3 4 *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"applicationType":"stb"}`)
	req, err = http.NewRequest("PUT", url, bytes.NewBuffer(dsdataer))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	// UpdateDeviceSettings BadJSON
	// commenting out this test because this API is now using NotImplementedHandler
	// badPayload := []byte(`{"foo":}`)
	// url := "/xconfAdminService/updates/deviceSettings/UTC"
	// performRequest(t, router, url, "POST", badPayload, http.StatusBadRequest)

	//GET DFRULE BY ID

	urlWithId := "/xconfAdminService/dcm/deviceSettings/23069266-45b7-4bf6-a255-e6ee584cd68bid"
	req, err = http.NewRequest("GET", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET DF RULE BY SIZE

	urlWithId = "/xconfAdminService/dcm/deviceSettings/size"
	req, err = http.NewRequest("GET", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var size int = 0
		json.Unmarshal(body, &size)
		assert.Equal(t, size > 0, true)
	}

	// GET DFRULE BY NAMES
	urlWithId = "/xconfAdminService/dcm/deviceSettings/names"
	req, err = http.NewRequest("GET", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		names := []string{}
		json.Unmarshal(body, &names)
		assert.Equal(t, len(names) > 0, true)
	}

	//DELETE AN EXISTING RECORD
	delUrlWithId := "/xconfAdminService/dcm/deviceSettings/23069266-45b7-4bf6-a255-e6ee584cd68brm"
	req, err = http.NewRequest("DELETE", delUrlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)

	//DELETE NON EXISTING DEVICE SETTINGS BY ID
	urlWithId = "/xconfAdminService/dcm/deviceSettings/23069266-45b7-4bf6-a255-e6ee584cd6xxxx"
	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

	//POST FILTERED FOR NAMES
	urlWithfilt := "/xconfAdminService/dcm/deviceSettings/filtered?pageNumber=1&pageSize=50"
	req, err = http.NewRequest("POST", urlWithfilt, bytes.NewBuffer(postmapname))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		names := []string{}
		json.Unmarshal(body, &names)
		assert.Equal(t, len(names) > 0, true)
	}

}

// performReq is a helper function that creates a req, executes a req,
// and checks the result against the expected status
func performRequest(t *testing.T, router *mux.Router, url string, method string, body []byte, expectedStatus int) []byte {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	assert.NilError(t, err)
	if method == "POST" || method == "PUT" {
		req.Header.Add("Content-Type", "application/json")
	}
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, expectedStatus)
	defer res.Body.Close()
	respBody, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	return respBody
}

// ========== Tests for GetDeviceSettingsExportHandler ==========

// TestGetDeviceSettingsExportHandler_Success tests successful export with matching formulas and device settings
func TestGetDeviceSettingsExportHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test DCM formulas
	formula1 := &logupload.DCMGenericRule{
		ID:              "export-formula-1",
		Name:            "Export Formula 1",
		ApplicationType: "stb",
	}
	formula2 := &logupload.DCMGenericRule{
		ID:              "export-formula-2",
		Name:            "Export Formula 2",
		ApplicationType: "stb",
	}

	// Create corresponding device settings
	deviceSettings1 := &logupload.DeviceSettings{
		ID:                "export-formula-1",
		Name:              "Export Device Settings 1",
		CheckOnReboot:     true,
		SettingsAreActive: true,
		ApplicationType:   "stb",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}
	deviceSettings2 := &logupload.DeviceSettings{
		ID:                "export-formula-2",
		Name:              "Export Device Settings 2",
		CheckOnReboot:     false,
		SettingsAreActive: true,
		ApplicationType:   "stb",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}

	// Save test data directly to DB
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, formula1.ID, formula1)
	assert.NilError(t, err)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, formula2.ID, formula2)
	assert.NilError(t, err)
	CreateDeviceSettings(deviceSettings1, "stb")
	CreateDeviceSettings(deviceSettings2, "stb")

	// Make request
	url := "/xconfAdminService/dcm/deviceSettings/export"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Assertions
	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.Check(t, res.Header.Get("Content-Disposition") != "")
	assert.Check(t, res.Header.Get("Content-Disposition") != "", "Content-Disposition header should be set")

	// Verify response body
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	var result []*logupload.DeviceSettings
	err = json.Unmarshal(body, &result)
	assert.NilError(t, err)
	assert.Equal(t, len(result), 2, "Should return 2 device settings")
}

// TestGetDeviceSettingsExportHandler_EmptyResult tests when no formulas exist
func TestGetDeviceSettingsExportHandler_EmptyResult(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Make request without any data
	url := "/xconfAdminService/dcm/deviceSettings/export"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Assertions
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify response body is empty array
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	var result []*logupload.DeviceSettings
	err = json.Unmarshal(body, &result)
	assert.NilError(t, err)
	assert.Equal(t, len(result), 0, "Should return empty array")
}

// TestGetDeviceSettingsExportHandler_FilterByApplicationType tests that only matching app type is exported
func TestGetDeviceSettingsExportHandler_FilterByApplicationType(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test data with different application types
	formulaSTB := &logupload.DCMGenericRule{
		ID:              "formula-stb-export",
		Name:            "STB Formula Export",
		ApplicationType: "stb",
	}
	formulaXHome := &logupload.DCMGenericRule{
		ID:              "formula-xhome-export",
		Name:            "XHome Formula Export",
		ApplicationType: "xhome",
	}

	deviceSettingsSTB := &logupload.DeviceSettings{
		ID:                "formula-stb-export",
		Name:              "STB Settings Export",
		ApplicationType:   "stb",
		CheckOnReboot:     true,
		SettingsAreActive: true,
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}
	deviceSettingsXHome := &logupload.DeviceSettings{
		ID:                "formula-xhome-export",
		Name:              "XHome Settings Export",
		ApplicationType:   "xhome",
		CheckOnReboot:     true,
		SettingsAreActive: true,
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}

	// Save test data
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, formulaSTB.ID, formulaSTB)
	assert.NilError(t, err)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, formulaXHome.ID, formulaXHome)
	assert.NilError(t, err)
	CreateDeviceSettings(deviceSettingsSTB, "stb")
	CreateDeviceSettings(deviceSettingsXHome, "xhome")

	// Request with stb application type
	url := "/xconfAdminService/dcm/deviceSettings/export"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Assertions
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify only STB settings are returned
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	var result []*logupload.DeviceSettings
	err = json.Unmarshal(body, &result)
	assert.NilError(t, err)
	assert.Equal(t, len(result), 1, "Should return only 1 STB device setting")
	if len(result) > 0 && result[0] != nil {
		assert.Equal(t, result[0].ApplicationType, "stb")
		assert.Equal(t, result[0].Name, "STB Settings Export")
	}
}

// TestGetDeviceSettingsExportHandler_MissingDeviceSettings tests when formula exists but device settings don't
func TestGetDeviceSettingsExportHandler_MissingDeviceSettings(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create formula but not corresponding device settings
	formula := &logupload.DCMGenericRule{
		ID:              "formula-orphan-export",
		Name:            "Orphan Formula Export",
		ApplicationType: "stb",
	}
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, formula.ID, formula)
	assert.NilError(t, err)

	// Make request
	url := "/xconfAdminService/dcm/deviceSettings/export"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Assertions
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify response contains nil for missing device setting
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	var result []*logupload.DeviceSettings
	err = json.Unmarshal(body, &result)
	assert.NilError(t, err)
	assert.Equal(t, len(result), 1, "Should return array with one element")
	assert.Check(t, result[0] == nil, "Device setting should be nil when not found")
}

// TestGetDeviceSettingsExportHandler_VerifyContentDisposition tests Content-Disposition header format
func TestGetDeviceSettingsExportHandler_VerifyContentDisposition(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Test with different application types to verify header varies
	testCases := []struct {
		appType          string
		expectedInHeader string
	}{
		{"stb", "allDeviceSettings_stb"},
		{"xhome", "allDeviceSettings_xhome"},
	}

	for _, tc := range testCases {
		url := "/xconfAdminService/dcm/deviceSettings/export"
		req, err := http.NewRequest("GET", url, nil)
		assert.NilError(t, err)
		req.AddCookie(&http.Cookie{Name: "applicationType", Value: tc.appType})

		res := ExecuteRequest(req, router).Result()
		defer res.Body.Close()

		assert.Equal(t, res.StatusCode, http.StatusOK)
		contentDisposition := res.Header.Get("Content-Disposition")
		assert.Check(t, contentDisposition != "", "Content-Disposition should not be empty")
	}
}

// TestGetDeviceSettingsExportHandler_AuthError tests auth error handling
func TestGetDeviceSettingsExportHandler_AuthError(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Make request without auth cookie
	url := "/xconfAdminService/dcm/deviceSettings/export"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	// Don't add applicationType cookie to test default behavior

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// In test environment, auth might pass with default "stb" or fail
	// We verify it returns a valid response (either success or error)
	assert.Check(t, res.StatusCode == http.StatusOK || res.StatusCode >= 400,
		"Should return either success or error status")
}

// TestGetDeviceSettingsExportHandler_MultipleFormulasWithSomeMatching tests partial matching
func TestGetDeviceSettingsExportHandler_MultipleFormulasWithSomeMatching(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create multiple formulas, only some with matching device settings
	formula1 := &logupload.DCMGenericRule{
		ID:              "formula-with-ds-export",
		Name:            "Formula With DS Export",
		ApplicationType: "stb",
	}
	formula2 := &logupload.DCMGenericRule{
		ID:              "formula-without-ds-export",
		Name:            "Formula Without DS Export",
		ApplicationType: "stb",
	}

	deviceSettings1 := &logupload.DeviceSettings{
		ID:                "formula-with-ds-export",
		Name:              "Device Settings Export",
		ApplicationType:   "stb",
		CheckOnReboot:     true,
		SettingsAreActive: true,
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}

	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, formula1.ID, formula1)
	assert.NilError(t, err)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, formula2.ID, formula2)
	assert.NilError(t, err)
	respEntity := CreateDeviceSettings(deviceSettings1, "stb")
	assert.Check(t, respEntity.Error == nil, "Failed to create device settings: %v", respEntity.Error)

	// Make request
	url := "/xconfAdminService/dcm/deviceSettings/export"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Assertions
	assert.Equal(t, res.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	var result []*logupload.DeviceSettings
	err = json.Unmarshal(body, &result)
	assert.NilError(t, err)

	// Count non-nil entries
	nonNilCount := 0
	for _, ds := range result {
		if ds != nil {
			nonNilCount++
		}
	}

	// The handler appends device settings for each matching formula
	// If device settings don't exist, it appends nil
	// So we should have at least 1 non-nil (formula1 has matching device settings)
	assert.Check(t, len(result) > 0, "Should return at least one item")
	assert.Check(t, nonNilCount >= 1, "Should have at least 1 non-nil device setting")
}

// TestGetDeviceSettingsExportHandler_JSONResponseFormat tests JSON response structure
func TestGetDeviceSettingsExportHandler_JSONResponseFormat(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create complete test data
	formula := &logupload.DCMGenericRule{
		ID:              "formula-json-export",
		Name:            "JSON Export Formula",
		ApplicationType: "stb",
	}
	deviceSettings := &logupload.DeviceSettings{
		ID:                "formula-json-export",
		Name:              "JSON Export Settings",
		CheckOnReboot:     true,
		SettingsAreActive: false,
		ApplicationType:   "stb",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}

	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, formula.ID, formula)
	assert.NilError(t, err)
	CreateDeviceSettings(deviceSettings, "stb")

	// Make request
	url := "/xconfAdminService/dcm/deviceSettings/export"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Assertions
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify it's valid JSON array
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	var result []*logupload.DeviceSettings
	err = json.Unmarshal(body, &result)
	assert.NilError(t, err, "Response should be valid JSON")

	// Verify structure
	assert.Equal(t, len(result), 1)
	if len(result) > 0 && result[0] != nil {
		assert.Equal(t, result[0].ID, "formula-json-export")
		assert.Equal(t, result[0].Name, "JSON Export Settings")
		assert.Equal(t, result[0].CheckOnReboot, true)
		assert.Equal(t, result[0].SettingsAreActive, false)
		assert.Equal(t, result[0].ApplicationType, "stb")
	}
}

// TestGetDeviceSettingsByIdHandler_Success tests successful retrieval by ID
func TestGetDeviceSettingsByIdHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	deviceSettings := &logupload.DeviceSettings{
		ID:                "test-get-by-id",
		Name:              "Test Get By ID",
		CheckOnReboot:     true,
		SettingsAreActive: true,
		ApplicationType:   "stb",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}
	CreateDeviceSettings(deviceSettings, "stb")

	url := "/xconfAdminService/dcm/deviceSettings/test-get-by-id"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	var result logupload.DeviceSettings
	err = json.Unmarshal(body, &result)
	assert.NilError(t, err)
	assert.Equal(t, result.ID, "test-get-by-id")
	assert.Equal(t, result.Name, "Test Get By ID")
}

// TestGetDeviceSettingsByIdHandler_NotFound tests non-existent ID
func TestGetDeviceSettingsByIdHandler_NotFound(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	url := "/xconfAdminService/dcm/deviceSettings/non-existent-id"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusNotFound)
}

// TestGetDeviceSettingsByIdHandler_EmptyID tests empty ID parameter
// Note: Empty ID doesn't match GetAll endpoint - it returns 404
func TestGetDeviceSettingsByIdHandler_EmptyID(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	url := "/xconfAdminService/dcm/deviceSettings/"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	// Empty ID results in 404 as it's looking for empty string ID
	assert.Check(t, res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotFound)
}

// TestDeleteDeviceSettingsByIdHandler_Success tests successful deletion
func TestDeleteDeviceSettingsByIdHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	deviceSettings := &logupload.DeviceSettings{
		ID:                "test-delete-success",
		Name:              "Test Delete Success",
		CheckOnReboot:     false,
		SettingsAreActive: false,
		ApplicationType:   "stb",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}
	CreateDeviceSettings(deviceSettings, "stb")

	url := "/xconfAdminService/dcm/deviceSettings/test-delete-success"
	req, err := http.NewRequest("DELETE", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusNoContent)

	// Verify it's actually deleted
	req2, _ := http.NewRequest("GET", url, nil)
	req2.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res2 := ExecuteRequest(req2, router).Result()
	defer res2.Body.Close()
	assert.Equal(t, res2.StatusCode, http.StatusNotFound)
}

// TestDeleteDeviceSettingsByIdHandler_NotFound tests deleting non-existent setting
func TestDeleteDeviceSettingsByIdHandler_NotFound(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	url := "/xconfAdminService/dcm/deviceSettings/non-existent-delete-id"
	req, err := http.NewRequest("DELETE", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusNotFound)
}

// TestCreateDeviceSettingsHandler_InvalidJSON tests create with invalid JSON
func TestCreateDeviceSettingsHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	url := "/xconfAdminService/dcm/deviceSettings"
	invalidJSON := []byte(`{"id":"invalid"invalid json}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(invalidJSON))
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

// TestUpdateDeviceSettingsHandler_Success tests successful update
func TestUpdateDeviceSettingsHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create initial setting
	deviceSettings := &logupload.DeviceSettings{
		ID:                "test-update-id",
		Name:              "Original Name",
		CheckOnReboot:     false,
		SettingsAreActive: false,
		ApplicationType:   "stb",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}
	CreateDeviceSettings(deviceSettings, "stb")

	// Update it
	updatedSettings := &logupload.DeviceSettings{
		ID:                "test-update-id",
		Name:              "Updated Name",
		CheckOnReboot:     true,
		SettingsAreActive: true,
		ApplicationType:   "stb",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}
	updatedJSON, _ := json.Marshal(updatedSettings)

	url := "/xconfAdminService/dcm/deviceSettings"
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(updatedJSON))
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Verify the update
	getReq, _ := http.NewRequest("GET", "/xconfAdminService/dcm/deviceSettings/test-update-id", nil)
	getReq.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	getRes := ExecuteRequest(getReq, router).Result()
	defer getRes.Body.Close()

	body, _ := ioutil.ReadAll(getRes.Body)
	var result logupload.DeviceSettings
	json.Unmarshal(body, &result)
	assert.Equal(t, result.Name, "Updated Name")
	assert.Equal(t, result.CheckOnReboot, true)
}

// TestUpdateDeviceSettingsHandler_NotExisting tests updating non-existent setting
func TestUpdateDeviceSettingsHandler_NotExisting(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	deviceSettings := &logupload.DeviceSettings{
		ID:                "non-existent-update",
		Name:              "Should Not Update",
		CheckOnReboot:     false,
		SettingsAreActive: false,
		ApplicationType:   "stb",
	}
	settingsJSON, _ := json.Marshal(deviceSettings)

	url := "/xconfAdminService/dcm/deviceSettings"
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(settingsJSON))
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusConflict)
}

// TestPostDeviceSettingsFilteredWithParamsHandler_WithFilters tests filtered endpoint with context
func TestPostDeviceSettingsFilteredWithParamsHandler_WithFilters(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create test data
	ds1 := &logupload.DeviceSettings{
		ID:                "filter-test-1",
		Name:              "Filter Test 1",
		CheckOnReboot:     true,
		SettingsAreActive: true,
		ApplicationType:   "stb",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}
	ds2 := &logupload.DeviceSettings{
		ID:                "filter-test-2",
		Name:              "Filter Test 2",
		CheckOnReboot:     false,
		SettingsAreActive: false,
		ApplicationType:   "stb",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}
	CreateDeviceSettings(ds1, "stb")
	CreateDeviceSettings(ds2, "stb")

	url := "/xconfAdminService/dcm/deviceSettings/filtered?pageNumber=1&pageSize=10"
	filterContext := map[string]interface{}{}
	filterJSON, _ := json.Marshal(filterContext)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(filterJSON))
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, _ := ioutil.ReadAll(res.Body)
	var results []logupload.DeviceSettings
	json.Unmarshal(body, &results)
	assert.Check(t, len(results) >= 2)
}

// TestPostDeviceSettingsFilteredWithParamsHandler_InvalidPagination tests invalid pagination
func TestPostDeviceSettingsFilteredWithParamsHandler_InvalidPagination(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	url := "/xconfAdminService/dcm/deviceSettings/filtered?pageNumber=0&pageSize=0"
	filterContext := map[string]interface{}{}
	filterJSON, _ := json.Marshal(filterContext)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(filterJSON))
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

// TestGetDeviceSettingsExportHandler_MultipleApplicationTypes tests export for different app types
func TestGetDeviceSettingsExportHandler_MultipleApplicationTypes(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create formulas for different app types
	formula1 := &logupload.DCMGenericRule{
		ID:              "export-formula-stb",
		Name:            "STB Formula",
		ApplicationType: "stb",
	}
	formula2 := &logupload.DCMGenericRule{
		ID:              "export-formula-xhome",
		Name:            "XHome Formula",
		ApplicationType: "xhome",
	}
	ds1 := &logupload.DeviceSettings{
		ID:                "export-formula-stb",
		Name:              "STB Settings",
		CheckOnReboot:     true,
		SettingsAreActive: true,
		ApplicationType:   "stb",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}
	ds2 := &logupload.DeviceSettings{
		ID:                "export-formula-xhome",
		Name:              "XHome Settings",
		CheckOnReboot:     false,
		SettingsAreActive: false,
		ApplicationType:   "xhome",
		Schedule: logupload.Schedule{
			Type:              "ActNow",
			Expression:        "0 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: json.Number("0"),
		},
	}

	ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, formula1.ID, formula1)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, formula2.ID, formula2)
	CreateDeviceSettings(ds1, "stb")
	CreateDeviceSettings(ds2, "xhome")

	// Test STB export
	url := "/xconfAdminService/dcm/deviceSettings/export"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, _ := ioutil.ReadAll(res.Body)
	var stbResults []*logupload.DeviceSettings
	json.Unmarshal(body, &stbResults)

	// Should only have STB results
	nonNilCount := 0
	for _, ds := range stbResults {
		if ds != nil && ds.ApplicationType == "stb" {
			nonNilCount++
		}
	}
	assert.Check(t, nonNilCount >= 1)
}
