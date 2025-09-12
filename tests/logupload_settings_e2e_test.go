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

	"gotest.tools/assert"
)

func ImportLogUploadTableData(data []string, tabletype logupload.LogUploadSettings) error {
	var err error
	for _, row := range data {
		err = json.Unmarshal([]byte(row), &tabletype)
		err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_LOG_UPLOAD_SETTINGS, tabletype.ID, &tabletype)
	}
	return err
}

func TestAllLogUploadSettingsApis(t *testing.T) {

	//GET ALL LOG REPO SETTINGS
	DeleteAllEntities()
	defer DeleteAllEntities()

	var tableData = []string{
		`{"id":"1845ea08-e2c3-4c36-8349-d613d93b78cup2","updated":1592418324468,"name":"dineshcreat2e23","uploadOnReboot":true,"numberOfDays":0,"areSettingsActive":true,"schedule":{"type":"ActNow","expression":"4 7 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"logFileIds":null,"logFilesGroupId":"","modeToGetLogFiles":"","uploadRepositoryId":"f946b0da-619c-4bc8-a876-11f1af2918ca","activeDateTimeRange":false,"fromDateTime":"","toDateTime":"","applicationType":"stb"}`,
	}
	ImportLogUploadTableData(tableData, logupload.LogUploadSettings{})

	urlall := "/xconfAdminService/dcm/logUploadSettings"
	req, err := http.NewRequest("GET", urlall, nil)
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
		var dss = []logupload.LogUploadSettings{}
		json.Unmarshal(body, &dss)
		assert.Equal(t, len(dss) > 0, true)
	}

	//CREATE LOG UPLOAD DATA SETTING

	ludata := []byte(
		`{"id":"1845ea08-e2c3-4c36-8349-d613d93b78ccp2","updated":1592418324568,"name":"dineshcreate23","uploadOnReboot":true,"numberOfDays":0,"areSettingsActive":true,"schedule":{"type":"ActNow","expression":"4 7 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"logFileIds":null,"logFilesGroupId":"","modeToGetLogFiles":"","uploadRepositoryId":"f946b0da-619c-4bc8-a876-11f1af2918ca","activeDateTimeRange":false,"fromDateTime":"","toDateTime":"","applicationType":"stb"}`)

	urlCr := "/xconfAdminService/dcm/logUploadSettings?applicationType=stb"
	req, err = http.NewRequest("POST", urlCr, bytes.NewBuffer(ludata))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	//ERROR CREATING AGAIN SAME ENTRY
	urlCr = "/xconfAdminService/dcm/logUploadSettings?applicationType=stb"
	req, err = http.NewRequest("POST", urlCr, bytes.NewBuffer(ludata))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	//UPDATE EXISTING ENTRY

	ludataup := []byte(
		`{"id":"1845ea08-e2c3-4c36-8349-d613d93b78ccp2","updated":1592418324468,"name":"dineshupdate","uploadOnReboot":true,"numberOfDays":0,"areSettingsActive":true,"schedule":{"type":"ActNow","expression":"4 7 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"logFileIds":null,"logFilesGroupId":"","modeToGetLogFiles":"","uploadRepositoryId":"f946b0da-619c-4bc8-a876-11f1af2918ca","activeDateTimeRange":false,"fromDateTime":"","toDateTime":"","applicationType":"stb"}`)
	urlup := "/xconfAdminService/dcm/logUploadSettings?applicationType=stb"
	req, err = http.NewRequest("PUT", urlup, bytes.NewBuffer(ludataup))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//UPDATE NON EXISTING ENTRY

	ludataer := []byte(
		`{"id":"1845ea08-e2c3-4c36-8349-d613d93b78err","updated":1592418324468,"name":"dineshcreaterr","uploadOnReboot":true,"numberOfDays":0,"areSettingsActive":true,"schedule":{"type":"ActNow","expression":"4 7 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"logFileIds":null,"logFilesGroupId":"","modeToGetLogFiles":"","uploadRepositoryId":"f946b0da-619c-4bc8-a876-11f1af2918ca","activeDateTimeRange":false,"fromDateTime":"","toDateTime":"","applicationType":"stb"}`)
	urlup = "/xconfAdminService/dcm/logUploadSettings?applicationType=stb"
	req, err = http.NewRequest("PUT", urlup, bytes.NewBuffer(ludataer))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	//GET LOGUPLOADSETTINGS BY ID

	urlWithId := "/xconfAdminService/dcm/logUploadSettings/1845ea08-e2c3-4c36-8349-d613d93b78cup2?applicationType=stb"
	req, err = http.NewRequest("GET", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET LOGUPLOADSETTINGS BY SIZE
	urlWithId = "/xconfAdminService/dcm/logUploadSettings/size?applicationType=stb"
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
	//GET LOGUPLOADSETTINGS NAMES

	urlWithId = "/xconfAdminService/dcm/logUploadSettings/names?applicationType=stb"
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
		var dss = []logupload.LogUploadSettings{}
		json.Unmarshal(body, &dss)
		assert.Equal(t, len(dss) > 0, true)
	}

	//GET LOGUPLOAD SETTINGS FILTER NAMES
	urlWithId = "/xconfAdminService/dcm/logUploadSettings/filtered?pageNumber=1&pageSize=50"
	postmapname = []byte(`{"NAME": "dineshcreat2e23"}`)
	req, err = http.NewRequest("POST", urlWithId, bytes.NewBuffer(postmapname))
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
		var dss = []logupload.LogUploadSettings{}
		json.Unmarshal(body, &dss)
		assert.Equal(t, len(dss) > 0, true)
	}

	//DELETE LOGUPLOAD SETTINGS BY ID

	urlWithId = "/xconfAdminService/dcm/logUploadSettings/1845ea08-e2c3-4c36-8349-d613d93b78cup2?applicationType=stb"
	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)

	//DELETE NON EXISTING DEVICE SETTINGS BY ID
	urlWithId = "/xconfAdminService/dcm/logUploadSettings/23069266-45b7-4bf6-a255-e6ee584cd6xxxx?applicationType=stb"

	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
}
