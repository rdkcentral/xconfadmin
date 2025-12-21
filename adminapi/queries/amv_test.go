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

	"github.com/google/uuid"
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
	t.Parallel()
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
		//	assert.Equal(t, len(amvrules), 1)
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

// Additional tests for comprehensive coverage of amv_handler and amv_service
func TestAmv_GetById_NotFound(t *testing.T) {
	t.Parallel()
	// create request with non-existent id
	urlWithId := fmt.Sprintf("%s/%s", AMV_URL, uuid.New().String())
	req, err := http.NewRequest("GET", urlWithId, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
}

func TestAmv_GetById_Export(t *testing.T) {
	t.Parallel()
	// prepare model and create an amv
	newModel := shared.Model{ID: "EXPORT00"}
	_, err1 := shared.SetOneModel(&newModel)
	assert.NilError(t, err1)
	amvID := uuid.New().String()
	body := fmt.Sprintf(`{"id":"%s","applicationType":"stb","description":"descExp","regularExpressions":["re"],"model":"EXPORT00","firmwareVersions":[],"partnerId":"p"}`, amvID)
	req, err := http.NewRequest("POST", AMV_URL, bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	// export by id
	urlExport := fmt.Sprintf("%s/%s?export", AMV_URL, amvID)
	req, err = http.NewRequest("GET", urlExport, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.Assert(t, res.Header.Get("Content-Disposition") != "")
}

func TestAmv_GetAll_ExportAll(t *testing.T) {
	t.Parallel()
	// ensure at least one amv present per applicationType
	newModel := shared.Model{ID: "EXPALL00"}
	_, err1 := shared.SetOneModel(&newModel)
	assert.NilError(t, err1)
	amvID := uuid.New().String()
	body := fmt.Sprintf(`{"id":"%s","applicationType":"stb","description":"descAll","regularExpressions":["re"],"model":"EXPALL00","firmwareVersions":[],"partnerId":"p"}`, amvID)
	req, err := http.NewRequest("POST", AMV_URL, bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	req, err = http.NewRequest("GET", AMV_URL+"?exportAll", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.Assert(t, res.Header.Get("Content-Disposition") != "")
}

func TestAmv_Create_ApplicationTypeMismatch(t *testing.T) {
	t.Parallel()
	// model exists
	newModel := shared.Model{ID: "MIS00"}
	_, err1 := shared.SetOneModel(&newModel)
	assert.NilError(t, err1)
	// send different applicationType cookie than body to force conflict in CreateAmv
	amvID := uuid.New().String()
	body := fmt.Sprintf(`{"id":"%s","applicationType":"wrong","description":"mismatch","regularExpressions":["re"],"model":"MIS00","firmwareVersions":[],"partnerId":"p"}`, amvID)
	req, err := http.NewRequest("POST", AMV_URL, bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)
}

func TestAmv_Update_NotFound(t *testing.T) {
	t.Parallel()
	// attempt update with unknown id
	body := fmt.Sprintf(`{"id":"%s","applicationType":"stb","description":"desc","regularExpressions":["re"],"model":"UNKNOWN","firmwareVersions":[],"partnerId":"p"}`, uuid.New().String())
	req, err := http.NewRequest("PUT", AMV_URL, bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// model UNKNOWN not set; validation will fail with model does not exist -> BadRequest OR NotFound due to missing in DB after validation path differences
	assert.Assert(t, res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusNotFound)
}

func TestAmv_Filtered_Post_InvalidJSON(t *testing.T) {
	t.Parallel()
	// correct POST filtered endpoint lives under activationMinimumVersion
	req, err := http.NewRequest("POST", "/xconfAdminService/activationMinimumVersion/filtered?pageNumber=1&pageSize=10", bytes.NewBuffer([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func TestAmv_Filtered_Post_PaginationErrors(t *testing.T) {
	t.Parallel()
	// endpoints under activationMinimumVersion
	req, err := http.NewRequest("POST", "/xconfAdminService/activationMinimumVersion/filtered?pageNumber=0&pageSize=1", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
	// pageSize=0
	req, err = http.NewRequest("POST", "/xconfAdminService/activationMinimumVersion/filtered?pageNumber=1&pageSize=0", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func TestAmv_BatchCreateAndUpdate(t *testing.T) {
	t.Parallel()
	// create model
	newModel := shared.Model{ID: "BATCH00"}
	_, err := shared.SetOneModel(&newModel)
	assert.NilError(t, err)
	id1 := uuid.New().String()
	id2 := uuid.New().String()
	// batch create
	bodyCreate := fmt.Sprintf(`[{"id":"%s","applicationType":"stb","description":"d1","regularExpressions":["r1"],"model":"BATCH00","firmwareVersions":[],"partnerId":"p"},{"id":"%s","applicationType":"stb","description":"d2","regularExpressions":["r2"],"model":"BATCH00","firmwareVersions":[],"partnerId":"p"}]`, id1, id2)
	req, err := http.NewRequest("POST", "/xconfAdminService/activationMinimumVersion/entities", bytes.NewBuffer([]byte(bodyCreate)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// expect OK after batch create
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// batch update (modify description of one)
	bodyUpdate := fmt.Sprintf(`[{"id":"%s","applicationType":"stb","description":"d1u","regularExpressions":["r1"],"model":"BATCH00","firmwareVersions":[],"partnerId":"p"},{"id":"%s","applicationType":"stb","description":"d2u","regularExpressions":["r2"],"model":"BATCH00","firmwareVersions":[],"partnerId":"p"}]`, id1, id2)
	req, err = http.NewRequest("PUT", "/xconfAdminService/activationMinimumVersion/entities", bytes.NewBuffer([]byte(bodyUpdate)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
}

func TestAmv_ImportAll_MixingApplicationTypes(t *testing.T) {
	t.Parallel()
	newModel := shared.Model{ID: "MIX00"}
	_, err := shared.SetOneModel(&newModel)
	assert.NilError(t, err)
	amvID1 := uuid.New().String()
	amvID2 := uuid.New().String()
	body := fmt.Sprintf(`[{"id":"%s","applicationType":"stb","description":"d1","regularExpressions":["r"],"model":"MIX00","firmwareVersions":[],"partnerId":"p"},{"id":"%s","applicationType":"wrong","description":"d2","regularExpressions":["r"],"model":"MIX00","firmwareVersions":[],"partnerId":"p"}]`, amvID1, amvID2)
	req, err := http.NewRequest("POST", AMV_URL+"/importAll", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// observed status is 400 due to validation of applicationType wrong
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}
