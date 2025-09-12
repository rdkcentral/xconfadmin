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
	"io"
	"net/http"
	"net/url"
	"testing"

	core "github.com/rdkcentral/xconfwebconfig/shared"

	"gotest.tools/assert"
)

const (
	MACLIST_API             = "/xconfAdminService/genericnamespacedlist"
	jsonMaclistTestDataLocn = "jsondata/maclist/"
)

func TestExpansionContractionOfMacList(t *testing.T) {
	aut := newMaclistApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{MACLIST_API, "large_maclist", "", nil, "POST", "", http.StatusCreated, "saveIdIn=maclist_id_one", aut.maclistResponseValidator},
	}
	aut.run(testCases)
	macid := aut.getValOf("maclist_id_one")

	for i := 0; i < 10; i++ {
		testCases = []apiUnitTestCase{
			{MACLIST_API, "large_maclist", "", nil, "PUT", "", http.StatusOK, "", nil},
			{MACLIST_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + macid, http.StatusOK, NO_POSTERMS, nil},
			{MACLIST_API, "small_maclist", "", nil, "PUT", "", http.StatusOK, "", nil},
			{MACLIST_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + macid, http.StatusOK, NO_POSTERMS, nil},
		}
		aut.run(testCases)
	}

	testCases = []apiUnitTestCase{
		{MACLIST_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + macid, http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}

func newMaclistApiUnitTest(t *testing.T) *apiUnitTest {
	aut := newApiUnitTest(t)
	aut.setupMaclistApi()
	return aut
}

func (aut *apiUnitTest) setupMaclistApi() {
	if aut.getValOf(MACLIST_API) == "Done" {
		return
	}
	aut.setValOf(MACLIST_API+DATA_LOCN_SUFFIX, jsonMaclistTestDataLocn)
	aut.setValOf(MACLIST_API, "Done")
}

func (aut *apiUnitTest) maclistResponseValidator(tcase apiUnitTestCase, genRsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := io.ReadAll(genRsp.Body)
	assert.Equal(aut.t, tcase.api, MACLIST_API)
	var rsp = core.GenericNamespacedList{}
	json.Unmarshal(rspBody, &rsp)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	aut.saveIdIn(kvMap, rsp.ID)
}

func (aut *apiUnitTest) cleanupMaclistApi() {
	if aut.getValOf(MACLIST_API) == "" {
		return
	}
	aut.setValOf(MACLIST_API, "")
}
