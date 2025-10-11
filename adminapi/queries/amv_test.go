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
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"gotest.tools/assert"
)

var jsonAmvCreateData = []byte(
	`{
    "id": "42670af7-6ea2-485f-9aee-1fa5895d655b",
    "applicationType": "stb",
    "description": "APItest1DineshTuesdayiFINAL",
    "regularExpressions": [
       "apiTestExp"
    ],
    "model": "00",
    "firmwareVersions": [],
    "partnerId": "apiTest1"
}`)

var jsonAmvImportData = []byte(`[ 
{
     "id": "42670af7-6ea2-485f-9aee-1fa5895d6ws1",
        "applicationType": "stb",
        "description": "APItest3",
        "regularExpressions": [
       "apiTestExp"
    ],
        "model": "12",
        "firmwareVersions": [
            "a"
        ],
        "partnerId": "apiTest3"
}
] `)

var jsonAmvImporterrData = []byte(`[
{    
     "id": "42670af7-6ea2-485f-9aee-1fa5895d6wx1",
        "description": "APItest3",
        "regularExpressions": [
       "apiTestExp"],
        "model": "12",
        "firmwareVersions": [
            "a"
        ],
        "partnerId": "apiTest3"
}
] `)
var jsonAmvImportupdateErrData = []byte(`[
{
     "id": "42670af7-6ea2-485f-9aee-1fa5895d6ws1",
        "applicationType": "json",
        "description": "APItest3update",
        "regularExpressions": [
       "apiTestExp"
    ],
        "model": "12",
        "firmwareVersions": [
            "a"
        ],
        "partnerId": "apiTest3"
}
] `)
var jsonAmvImportupdateData = []byte(`[
{    
     "id": "42670af7-6ea2-485f-9aee-1fa5895d6ws1",
        "applicationType": "stb",
        "description": "APItest3update",
       "regularExpressions": [
       "apiTestExp"
    ],
        "model": "12",
        "firmwareVersions": [
            "a"
        ],
        "partnerId": "apiTest3"
}
] `)

var jsonAmvupdateData = []byte(
	`{
     "id": "42670af7-6ea2-485f-9aee-1fa5895d6ws1",
        "applicationType": "stb",
        "description": "APItest3Update",
        "regularExpressions": [
       "apiTestExp"
    ],
        "model": "12",
        "firmwareVersions": [
            "a"
        ],
        "partnerId": "apiTest3"
}`)

var jsonAmvupdateerrData = []byte(
	`{
     "id": "42670af7-6ea2-485f-9aee-1fa5895d6wx1",
        "description": "APItest3",
        "regularExpressions": [],
        "model": "12",
        "firmwareVersions": [
            "a"
        ],
        "partnerId": "apiTest3"
}`)

const (
	AMV_URL = "/xconfAdminService/amv"
)

func TestAmvAllApi(t *testing.T) {
	//t.Skip("TODO:need to move this to adminapi")
	//	config := GetTestConfig()
	//	_, router := GetTestWebConfigServer(config)

	//Badrequest
	req, err := http.NewRequest("POST", AMV_URL, bytes.NewBuffer(jsonAmvCreateData))
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)

	// with Model good case
	newModel := shared.Model{}
	newModel.ID = "00"
	_, err1 := shared.SetOneModel(&newModel)
	assert.NilError(t, err1)

	req, err = http.NewRequest("POST", AMV_URL, bytes.NewBuffer(jsonAmvCreateData))
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	// get amv by id
	urlWithId := fmt.Sprintf("%s/%s", AMV_URL, "42670af7-6ea2-485f-9aee-1fa5895d655b")
	req, err = http.NewRequest("GET", urlWithId, nil)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// get amv all
	req, err = http.NewRequest("GET", AMV_URL, nil)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	if res.StatusCode == http.StatusOK {
		var amvrules = []corefw.ActivationVersion{}
		json.Unmarshal(body, &amvrules)
		assert.Equal(t, len(amvrules), 1)
	}

	// filtered
	urlfiltered := fmt.Sprintf("%s/%s", AMV_URL, "filtered?applicationType=stb&MODEL=00")
	req, err = http.NewRequest("GET", urlfiltered, nil)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// filtered invalid path
	urlfilterederr := fmt.Sprintf("%s/%s", AMV_URL, "filtered?applicationType=stb&MODEL=00")
	req, err = http.NewRequest("GET", urlfilterederr, nil)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//importAll good case
	impnewModel := shared.Model{}
	impnewModel.ID = "12"
	_, err2 := shared.SetOneModel(&impnewModel)
	assert.NilError(t, err2)

	urlimport := fmt.Sprintf("%s/%s", AMV_URL, "importAll")
	req, err = http.NewRequest("POST", urlimport, bytes.NewBuffer(jsonAmvImportData))
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		bodyMap := map[string][]string{}
		json.Unmarshal(body, &bodyMap)
		assert.Equal(t, len(bodyMap["IMPORTED"]) > 0, true)
	}
	// err not imported
	req, err = http.NewRequest("POST", urlimport, bytes.NewBuffer(jsonAmvImporterrData))
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		bodyMap := map[string][]string{}
		json.Unmarshal(body, &bodyMap)
		assert.Equal(t, len(bodyMap["NOT_IMPORTED"]) > 0, true)
	}

	//update ImportALL error
	req, err = http.NewRequest("POST", urlimport, bytes.NewBuffer(jsonAmvImportupdateErrData))
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)

	//update ImportALL
	req, err = http.NewRequest("POST", urlimport, bytes.NewBuffer(jsonAmvImportupdateData))
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		bodyMap := map[string][]string{}
		json.Unmarshal(body, &bodyMap)
		assert.Equal(t, len(bodyMap["IMPORTED"]) > 0, true)
	}

	// update good case
	req, err = http.NewRequest("PUT", AMV_URL, bytes.NewBuffer(jsonAmvupdateData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// update error case
	// req, err = http.NewRequest("PUT", AMV_URL, bytes.NewBuffer(jsonAmvupdateerrData))
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Accept", "application/json")
	// assert.NilError(t, err)
	// res = ExecuteRequest(req, router).Result()
	// defer res.Body.Close()
	// assert.Equal(t, res.StatusCode, http.StatusBadRequest)

	// delete amv by id
	req, err = http.NewRequest("DELETE", urlWithId, nil)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)

	// delete non existing amv by id
	// TODO:commenting this to make sure there is no issue else where...
	// req, err = http.NewRequest("DELETE", urlWithId, nil)
	// req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	// req.Header.Set("Accept", "application/json")
	// assert.NilError(t, err)
	// res = ExecuteRequest(req, router).Result()
	// defer res.Body.Close()
	// assert.Equal(t, res.StatusCode, http.StatusNotFound)
}
