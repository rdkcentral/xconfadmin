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

func ImportLogRepTableData(data []string, tabletype logupload.UploadRepository) error {
	var err error
	for _, row := range data {
		err = json.Unmarshal([]byte(row), &tabletype)
		err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_UPLOAD_REPOSITORY, tabletype.ID, &tabletype)
	}
	return err
}

func TestAllLogRepoSettingsAPIs(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	//GET ALL LOG REPO SETTINGS

	var tableData = []string{
		`{"id":"fbf6c28a-ef6c-4494-8894-f77f03a62ba5","updated":1428932050824,"name":"protocoltest_6","description":"SCP","url":"tftp://pro.net","applicationType":"stb","protocol":"SCP"}`,
		`{"id":"fbf6c28a-ef6c-4494-8894-f77f03a62ca5","updated":1428932050824,"name":"dineshprotocoltest_6","description":"SCP","url":"tftp://pro.net","applicationType":"stb","protocol":"SCP"}`,
	}
	ImportLogRepTableData(tableData, logupload.UploadRepository{})

	urlall := "/xconfAdminService/dcm/uploadRepository"
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
		var dss = []logupload.UploadRepository{}
		json.Unmarshal(body, &dss)
		assert.Equal(t, len(dss) > 0, true)
	}

	//CREATE A NEW ENTRY
	lrdata := []byte(
		`{"id":"60b1e67c-d099-45d7-b163-dae9463dd6cr","updated":1635957735115,"name":"dineshcreate","description":"crtest","url":"http://test.com","applicationType":"stb","protocol":"HTTP"}`)

	urlCr := "/xconfAdminService/dcm/uploadRepository?applicationType=stb"
	req, err = http.NewRequest("POST", urlCr, bytes.NewBuffer(lrdata))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	//ERROR CREATING AGAIN SAME ENTRY
	urlCr = "/xconfAdminService/dcm/uploadRepository?applicationType=stb"
	req, err = http.NewRequest("POST", urlCr, bytes.NewBuffer(lrdata))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	//UPDATE EXISITNG ENTRY

	lrdataup := []byte(
		`{"id":"60b1e67c-d099-45d7-b163-dae9463dd6cr","updated":1635957735115,"name":"dineshupdate","description":"uptest","url":"http://test.com","applicationType":"stb","protocol":"HTTP"}`)
	urlup := "/xconfAdminService/dcm/uploadRepository?applicationType=stb"
	req, err = http.NewRequest("PUT", urlup, bytes.NewBuffer(lrdataup))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//UPDATE NON EXISTING ENTRY

	lrdataer := []byte(
		`{"id":"60b1e67c-d099-45d7-b163-dae9463dd6er","updated":1635957735115,"name":"dineshupdate","description":"uptest","url":"http://test.com","applicationType":"stb","protocol":"HTTP"}`)
	urlup = "/xconfAdminService/dcm/uploadRepository?applicationType=stb"
	req, err = http.NewRequest("PUT", urlup, bytes.NewBuffer(lrdataer))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	//GET ONE LOG REPO SETTINGS

	urlWithId := "/xconfAdminService/dcm/uploadRepository/fbf6c28a-ef6c-4494-8894-f77f03a62ba5?applicationType=stb"
	req, err = http.NewRequest("GET", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET LOG REPO SETTINGS BY SIZE
	urlWithId = "/xconfAdminService/dcm/uploadRepository/size"
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

	//GET LOG REPO SETTINGS BY NAMES
	urlWithId = "/xconfAdminService/dcm/uploadRepository/names"
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
		var dss = []logupload.UploadRepository{}
		json.Unmarshal(body, &dss)
		assert.Equal(t, len(dss) > 0, true)
	}

	//GET LOG REPO SETTINGS WITH FILTERED
	urlWithId = "/xconfAdminService/dcm/uploadRepository/filtered?pageNumber=1&pageSize=50"
	req, err = http.NewRequest("POST", urlWithId, bytes.NewBuffer(postmapname))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var dss = []logupload.UploadRepository{}
		json.Unmarshal(body, &dss)
		assert.Equal(t, len(dss) > 0, true)
	}

	//DELETE LOG REPO SETTINGS BY ID
	urlWithId = "/xconfAdminService/dcm/uploadRepository/fbf6c28a-ef6c-4494-8894-f77f03a62ba5"
	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)

	// DELETE NON EXISTING BY ID

	urlWithId = "/xconfAdminService/dcm/uploadRepository/23069266-45b7-4bf6-a255-e6ee584cd6xxxx"
	// delete non existing device Settings by id
	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
}
