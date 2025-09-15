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
package tests

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
