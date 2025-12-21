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
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	core "github.com/rdkcentral/xconfwebconfig/shared"

	assert "gotest.tools/assert"
)

const (
	MODEL_QAPI = "/xconfAdminService/queries/models"
	MODEL_UAPI = "/xconfAdminService/updates/models"
	MODEL_DAPI = "/xconfAdminService/delete/models"

	jsonModelTestDataLocn = "jsondata/model/"
)

func newModelApiUnitTest(t *testing.T) *apiUnitTest {
	aut := newApiUnitTest(t)
	aut.setValOf(MODEL_QAPI+DATA_LOCN_SUFFIX, jsonModelTestDataLocn)
	aut.setValOf(MODEL_UAPI+DATA_LOCN_SUFFIX, jsonModelTestDataLocn)
	aut.setValOf(MODEL_DAPI+DATA_LOCN_SUFFIX, jsonModelTestDataLocn)
	aut.setupModelApi()
	return aut
}

// func (aut *apiUnitTest) setupModelApi() {
// 	if aut.getValOf(MODEL_QAPI) == "Done" {
// 		return
// 	}
// 	aut.setValOf(MODEL_QAPI+DATA_LOCN_SUFFIX, jsonModelTestDataLocn)
// 	aut.setValOf(MODEL_UAPI+DATA_LOCN_SUFFIX, jsonModelTestDataLocn)
// 	aut.setValOf(MODEL_DAPI+DATA_LOCN_SUFFIX, jsonModelTestDataLocn)

// 	aut.setValOf(MODEL_QAPI, "Done")
// }

func (aut *apiUnitTest) cleanupModelApi() {
	if aut.getValOf(MODEL_QAPI) == "" {
		return
	}
	aut.setValOf(MODEL_QAPI, "")
}

// func (aut *apiUnitTest) modelArrayValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
// 	rspBody, _ := ioutil.ReadAll(rsp.Body)
// 	assert.Equal(aut.t, tcase.api == MODEL_QAPI || tcase.api == MODEL_WHOLE_API, true)

// 	var entries = []core.Model{}
// 	json.Unmarshal(rspBody, &entries)

// 	kvMap, err := url.ParseQuery(tcase.postTerms)
// 	assert.NilError(aut.t, err)

// 	aut.assertFetched(kvMap, len(entries))
// 	aut.saveFetchedCntIn(kvMap, len(entries))
// }

func (aut *apiUnitTest) modelSingleValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, tcase.api == MODEL_QAPI || tcase.api == MODEL_WHOLE_API, true)

	var entry = core.Model{}
	json.Unmarshal(rspBody, &entry)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)
	ID, ok := kvMap["ID"]
	if ok {
		assert.Equal(aut.t, ID[0], entry.ID)
	}
}

func (aut *apiUnitTest) modelResponseValidator(tcase apiUnitTestCase, genRsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(genRsp.Body)
	assert.Equal(aut.t, tcase.api == MODEL_UAPI || tcase.api == MODEL_WHOLE_API, true)
	var rsp = core.ModelResponse{}
	json.Unmarshal(rspBody, &rsp)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	aut.saveIdIn(kvMap, rsp.ID)

	validate, ok := kvMap["validate"]
	if !ok || validate[0] != "true" {
		return
	}

	req := core.NewModel("", "")
	reqBodyBytes, _ := ioutil.ReadAll(reqBody)
	err = json.Unmarshal(reqBodyBytes, &req)
	assert.NilError(aut.t, err)
	if req.ID != "" {
		assert.Equal(aut.t, rsp.ID, strings.ToUpper(req.ID))
	}
	assert.Equal(aut.t, rsp.Description, req.Description)
}

func (aut *apiUnitTest) modelEntitiesMapValidator(tcase apiUnitTestCase, genRsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(genRsp.Body)
	assert.Equal(aut.t, tcase.api == MODEL_UAPI || tcase.api == MODEL_WHOLE_API, true)
	var entitiesMap = make(map[string]string)
	json.Unmarshal(rspBody, &entitiesMap)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	for k, v := range entitiesMap {
		if strings.Contains(v, "SUCCESS") {
			aut.saveIdIn(kvMap, k)
		}
	}
}

func TestModelsCRUD(t *testing.T) {
	t.Parallel()
	aut := newModelApiUnitTest(t)
	sysGenId1 := uuid.New().String()
	sysGenId2 := uuid.New().String()

	testCases := []apiUnitTestCase{
		{MODEL_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=model_count", aut.modelArrayValidator},
		{MODEL_UAPI, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId1, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_one&validate=true", aut.modelResponseValidator},
		{MODEL_UAPI, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_two&validate=true", aut.modelResponseValidator},
	}
	aut.run(testCases)

	m1 := aut.getValOf("model_id_one")
	m2 := aut.getValOf("model_id_two")

	testCases = []apiUnitTestCase{
		{MODEL_UAPI, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + aut.getValOf("model_id_one"), aut.replaceKeysByValues, "PUT", "", http.StatusOK, "validate=true", aut.modelResponseValidator},

		{MODEL_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("model_count+2"), aut.modelArrayValidator},
		{MODEL_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + m1, http.StatusOK, "ID=" + m1, aut.modelSingleValidator},
		{MODEL_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + m2, http.StatusOK, "ID=" + m2, aut.modelSingleValidator},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + m1, http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + m2, http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("model_count"), aut.modelArrayValidator},
	}
	aut.run(testCases)
}
