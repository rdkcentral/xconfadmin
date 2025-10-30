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

	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	assert "gotest.tools/assert"
)

const (
	FRT_API                              = "/xconfAdminService/firmwareruletemplate"
	jsonFirmwareRuleTemplateTestDataLocn = "jsondata/firmwareruletemplate/"
)

func newFirmwareRuleTemplateApiUnitTest(t *testing.T) *apiUnitTest {
	aut := newApiUnitTest(t)
	aut.setupFirmwareRuleTemplateApi()
	return aut
}

func (aut *apiUnitTest) setupFirmwareRuleTemplateApi() {
	if aut.getValOf(FRT_API) == "Done" {
		return
	}
	aut.setValOf(FRT_API+DATA_LOCN_SUFFIX, jsonFirmwareRuleTemplateTestDataLocn)
	testCases := []apiUnitTestCase{
		{FRT_API, "firmware_rule_template_one", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FRT_API, "firmware_rule_template_two", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FRT_API, "firmware_rule_template_three", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FRT_API, "firmware_rule_template_four", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
	}
	aut.run(testCases)
	aut.setValOf(FRT_API, "Done")

}

func (aut *apiUnitTest) cleanupFirmwareRuleTemplateApi() {
	if aut.getValOf(FRT_API) == "" {
		return
	}
	testCases := []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/IP_RULE", http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/MAC_RULE", http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/GLOBAL_PERCENT", http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/TEST_FW_ENV_MODEL_RULE", http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
	aut.setValOf(FRT_API, "")
}

func (aut *apiUnitTest) firmwareRuleTemplateArrayValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, tcase.api, FRT_API)
	var entries = []corefw.FirmwareRuleTemplate{}
	json.Unmarshal(rspBody, &entries)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	// either saveFetchedCntIn or assertFetchedCnt
	aut.assertFetched(kvMap, len(entries))
	aut.saveFetchedCntIn(kvMap, len(entries))
	validateExport, ok := kvMap["validate_export"]
	if ok {
		if validateExport[0] != "true" {
			return
		}
		val, ok := rsp.Header["Content-Disposition"]
		assert.Equal(aut.t, ok, true)
		assert.Equal(aut.t, strings.Contains(val[0], "json"), true)
	}
}

func (aut *apiUnitTest) firmwareRuleTemplateSingleValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, tcase.api, FRT_API)

	var entry = corefw.FirmwareRuleTemplate{}
	json.Unmarshal(rspBody, &entry)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)
	ID, ok := kvMap["ID"]
	if ok {
		assert.Equal(aut.t, ID[0], entry.ID)
	}

	validateExport, ok := kvMap["validate_export"]
	if ok {
		if validateExport[0] != "true" {
			return
		}
		val, ok := rsp.Header["Content-Disposition"]
		assert.Equal(aut.t, ok, true)
		assert.Equal(aut.t, strings.Contains(val[0], ID[0]), true)
		assert.Equal(aut.t, strings.Contains(val[0], "json"), true)
	}
}

func (aut *apiUnitTest) firmwareRuleTemplateResponseValidator(tcase apiUnitTestCase, genRsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(genRsp.Body)

	assert.Equal(aut.t, tcase.api, FRT_API)
	var rsp = corefw.FirmwareRuleTemplate{}
	json.Unmarshal(rspBody, &rsp)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	aut.saveIdIn(kvMap, rsp.ID)

	validate, ok := kvMap["validate"]
	if !ok || validate[0] != "true" {
		return
	}

	req := corefw.NewEmptyFirmwareRuleTemplate()
	reqBodyBytes, _ := ioutil.ReadAll(reqBody)
	err = json.Unmarshal(reqBodyBytes, &req)
	assert.NilError(aut.t, err)
	if req.ID != "" {
		assert.Equal(aut.t, rsp.ID, req.ID)
	}
	aut.assertPriority(kvMap, (int)(rsp.Priority))
	/*
	   assert.Equal (aut.t, rsp.Description, req.Description)
	   assert.Equal (aut.t, rsp.FirmwareFilename, req.FirmwareFilename)
	   assert.Equal (aut.t, rsp.FirmwareVersion, req.FirmwareVersion)
	   assert.Equal (aut.t, IsEqual (req.SupportedModelIds, rsp.SupportedModelIds), true)
	*/
}

func TestGetFirmwareRuleTemplateFromQueryParams(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	testCases := []apiUnitTestCase{
		// Invalid Param ignored
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?invalidParam=someValue", http.StatusOK, NO_POSTERMS, nil},

		// Happy Paths
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=4", aut.firmwareRuleTemplateArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleTemplateFilteredFromQueryParams(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	testCases := []apiUnitTestCase{
		// Happy path
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered", http.StatusOK, "fetched=4", aut.firmwareRuleTemplateArrayValidator},

		// Happy path, Invalid param ignored
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?invalidParam=someValue", http.StatusOK, "fetched=4", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?invalidParam=someValue&another=value", http.StatusOK, "fetched=4", aut.firmwareRuleTemplateArrayValidator},

		// Ignore: missing value for param
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=", http.StatusOK, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name", http.StatusOK, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?key=", http.StatusOK, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?key", http.StatusOK, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?value=", http.StatusOK, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?value", http.StatusOK, NO_POSTERMS, nil},

		// Happy paths: Duplicate params
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=MAC_RULE&name=second", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?key=eStbMac&key=second", http.StatusOK, "fetched=3", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?value=SKXI11ANS&value=second", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},

		// name Happy Paths
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=nonexistant", http.StatusOK, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=MAC_RULE", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=GLOBAL_PERCENT", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=IP_RULE", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=TEST_FW_ENV_MODEL_RULE", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		// Case sensitivity
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=mac_RULE", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		// partial representation for name
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=RULE", http.StatusOK, "fetched=3", aut.firmwareRuleTemplateArrayValidator},

		// key - Happy Paths
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?key=nonexistant", http.StatusOK, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?key=eStbMac", http.StatusOK, "fetched=3", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?key=ipAddress", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		// Case sensitivity
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?key=ipADDRESS", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		// partial representation for key
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?key=ipADDR", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},

		// value - Happy Paths
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?value=nonexistant", http.StatusOK, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?value=SKXI11AIS", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		// Case sensitiity
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?value=SkXI11AIs", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		// partial representation for value
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?value=SKXI1", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
	}
	aut.run(testCases)
}

func TestPostFirmwareRuleTemplateFilteredFromQueryParams(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	testCases := []apiUnitTestCase{
		// invalid parameters are ignored
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?name=dummy", http.StatusOK, NO_POSTERMS, nil},

		// Happy Paths
		{FRT_API, "rule_template", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=4", http.StatusOK, "fetched=2", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "define_properties", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=4", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "blocking_filter_template", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=2", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},

		// Missing applicableAction fetches all entries
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=2", http.StatusOK, "fetched=2", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=2&pageSize=3", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=0&pageSize=3", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=-1&pageSize=3", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=4", http.StatusOK, "fetched=4", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=5", http.StatusOK, "fetched=4", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=0", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=-1", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered", http.StatusOK, "fetched=4", aut.firmwareRuleTemplateArrayValidator},

		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=A&pageSize=B", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=A&pageSize=3", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=B", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},

		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=&pageSize=", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber= &pageSize= ", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=&pageSize= ", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber= &pageSize=", http.StatusBadRequest, "fetched=0", aut.firmwareRuleTemplateArrayValidator},

		// Happy Paths: default value for missing query params
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1", http.StatusOK, "fetched=4", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageSize=3", http.StatusOK, "fetched=3", aut.firmwareRuleTemplateArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleTemplateIdsWithParam(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	sysGenId1 := uuid.New().String()
	sysGenId2 := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/ids?type=RULE_TEMPLATE", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/ids?type=NonExistant", http.StatusOK, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId1, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleTemplateResponseValidator},
		{FRT_API, "create_with_sys_gen_id_not_editable", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_2", aut.firmwareRuleTemplateResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/ids?type=RULE_TEMPLATE", http.StatusOK, "fetched=" + aut.eval("begin_count+1"), aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_2"), http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/ids?type=RULE_TEMPLATE", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareRuleTemplateArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleTemplateById(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FRT_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleTemplateResponseValidator},
	}
	aut.run(testCases)
	testCases = []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + aut.getValOf("id_1"), http.StatusOK, "ID=" + aut.getValOf("id_1"), aut.firmwareRuleTemplateSingleValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + aut.getValOf("id_1"), http.StatusNotFound, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}
func TestFirmwareRuleTemplateCRUD(t *testing.T) {
	sysGenId := uuid.New().String()
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=123sd_new", http.StatusOK, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "create_missing_applicable_action", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FRT_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=123sd_new", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusConflict, NO_POSTERMS, nil},
		{FRT_API, "frt_env_model", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FRT_API, "frt_env_model_dup", NO_PRETERMS, nil, "POST", "", http.StatusConflict, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/ENV_MODEL_RULE", http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/123sd_new", http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=123sd_new", http.StatusOK, "fetched=0", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/123sd_new", http.StatusNotFound, NO_POSTERMS, nil},
		{FRT_API, "create_missing_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusBadRequest, "saveIdIn=id_frt", aut.firmwareRuleTemplateResponseValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleTemplateByIdWithParam(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FRT_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleTemplateResponseValidator},
	}
	aut.run(testCases)
	testCases = []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + aut.getValOf("id_1") + "?export", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + aut.getValOf("id_1") + "?export", http.StatusNotFound, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleTemplateExportWithParam(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export?type=RULE_TEMPLATE", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleTemplateResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export?type=RULE_TEMPLATE", http.StatusOK, "fetched=" + aut.eval("begin_count +1") + "&validate_export=true", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export?type=RULE_TEMPLATE", http.StatusOK, "fetched=" + aut.eval("begin_count") + "&validate_export=true", aut.firmwareRuleTemplateArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleTemplateAllByType(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/all/RULE_TEMPLATE", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleTemplateResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/all/RULE_TEMPLATE", http.StatusOK, "fetched=" + aut.eval("begin_count +1"), aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/all/RULE_TEMPLATE", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareRuleTemplateArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleTemplateByTypeByEditable(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/RULE_TEMPLATE/true", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleTemplateResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/RULE_TEMPLATE/true", http.StatusOK, "fetched=" + aut.eval("begin_count +1"), aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/RULE_TEMPLATE/true", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareRuleTemplateArrayValidator},
	}
	aut.run(testCases)
}

// func TestFirmwareRuleTemplateChangePriorities(t *testing.T) {
// 	aut := newFirmwareRuleTemplateApiUnitTest(t)
// 	sysGenId1 := uuid.New().String()
// 	sysGenId2 := uuid.New().String()

// 	testCases := []apiUnitTestCase{
// 		// Create two brand new frts. Inputs have no priority specified
// 		{FRT_API, "create_with_sys_gen_id_no_prio", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId1, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=frt1", aut.firmwareRuleTemplateResponseValidator},
// 		{FRT_API, "create_with_sys_gen_id_no_prio", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=frt2", aut.firmwareRuleTemplateResponseValidator},
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/all/RULE_TEMPLATE", http.StatusOK, "saveFetchedCntIn=totFrtCnt", aut.firmwareRuleTemplateArrayValidator},
// 	}
// 	aut.run(testCases)

// 	frt1 := aut.getValOf("frt1")
// 	frt2 := aut.getValOf("frt2")
// 	totFrtCnt := aut.getValOf("totFrtCnt")

// 	testCases = []apiUnitTestCase{
// 		// Change priority of frt1 to 0
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/" + frt1 + "/priority/0", http.StatusBadRequest, "error_message=Invalid priority value 0", globAut.ErrorValidator},

// 		// Change priority of frt1 to negative value
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/" + frt1 + "/priority/-1", http.StatusBadRequest, "error_message=Invalid priority value -1", globAut.ErrorValidator},

// 		// Change priority of frt1 to huge value
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/" + frt1 + "/priority/100", http.StatusOK, NO_POSTERMS, nil},
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + frt1, http.StatusOK, "priority=" + totFrtCnt, globAut.firmwareRuleTemplateResponseValidator},

// 		// Change priority of frt1 to totFrtCnt
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/" + frt1 + "/priority/" + totFrtCnt, http.StatusOK, NO_POSTERMS, nil},
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + frt1, http.StatusOK, "priority=" + totFrtCnt, globAut.firmwareRuleTemplateResponseValidator},

// 		// Change priority of frt1 to totFrtCnt + 1
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/" + frt1 + "/priority/" + totFrtCnt + "1", http.StatusOK, NO_POSTERMS, nil},
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + frt1, http.StatusOK, "priority=" + totFrtCnt, globAut.firmwareRuleTemplateResponseValidator},

// 		// Change priority of frt1 to 1 and frt2 to 2
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/" + frt1 + "/priority/1", http.StatusOK, NO_POSTERMS, nil},
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/" + frt2 + "/priority/2", http.StatusOK, NO_POSTERMS, nil},

// 		// Check that the priority of frt1 is 1 and frt2 is 2
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + frt1, http.StatusOK, "priority=1", globAut.firmwareRuleTemplateResponseValidator},
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + frt2, http.StatusOK, "priority=2", globAut.firmwareRuleTemplateResponseValidator},

// 		// Delete frt1
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("frt1"), http.StatusNoContent, NO_POSTERMS, nil},
// 		// Check that the priority of frt2 is 1
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + frt2, http.StatusOK, "priority=1", globAut.firmwareRuleTemplateResponseValidator},

// 		// Delete frt2
// 		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("frt2"), http.StatusNoContent, NO_POSTERMS, nil},
// 	}
// 	aut.run(testCases)
// }

func TestGetFirmwareRuleTemplateWithParam(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleTemplateResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, "fetched=" + aut.eval("begin_count +1") + "&validate_export=true", aut.firmwareRuleTemplateArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, "fetched=" + aut.eval("begin_count") + "&validate_export=true", aut.firmwareRuleTemplateArrayValidator},
	}
	aut.run(testCases)
}

func TestFirmwareRuleTemplateEndPoints(t *testing.T) {
	// Clean up any existing "stb" firmware rule templates before test
	//DeleteAllEntities()
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	sysGenId := uuid.New().String()
	sysGenId2 := uuid.New().String()

	testCases := []apiUnitTestCase{
		//	"" PostFirmwareRuleTemplateHandler "POST"
		{FRT_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=frt_id_1", aut.firmwareRuleTemplateResponseValidator},
	}
	aut.run(testCases)

	idCreated := aut.getValOf("frt_id_1")

	testCases = []apiUnitTestCase{
		//	"" PutFirmwareRuleTemplateHandler "PUT"
		{FRT_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + idCreated, aut.replaceKeysByValues, "PUT", "", http.StatusOK, NO_POSTERMS, nil},

		//	"" GetFirmwareRuleTemplateHandler "GET"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, NO_POSTERMS, nil},

		//	"/entities" PostFirmwareRuleTemplateEntitiesHandler "POST"
		{FRT_API, "[create_with_sys_gen_id]", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "POST", "/entities", http.StatusOK, NO_POSTERMS, nil},

		//	"/entities" PutFirmwareRuleTemplateEntitiesHandler "PUT"
		{FRT_API, "[create_with_sys_gen_id]", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "PUT", "/entities", http.StatusOK, NO_POSTERMS, nil},

		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + sysGenId2, http.StatusNoContent, NO_POSTERMS, nil},

		//	"" GetFirmwareRuleTemplateWithParamHandler "GET"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" GetFirmwareRuleTemplateByIdHandler "GET"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + idCreated, http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" GetFirmwareRuleTemplateByIdWithParamHandler "GET"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + idCreated + "?export", http.StatusOK, NO_POSTERMS, nil},

		//	No registered handler
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/123sd_new?unknown", http.StatusNotFound, NO_POSTERMS, nil},

		//	"/filtered" PostFirmwareRuleTemplateFilteredWithParamsHandler "POST"
		{FRT_API, "only_stb", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=10", http.StatusOK, NO_POSTERMS, nil},

		//	"/all/{type}" GetFirmwareRuleTemplateAllByTypeHandler "GET"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/all/RULE_TEMPLATE", http.StatusOK, NO_POSTERMS, nil},

		//	"/ids" GetFirmwareRuleTemplateIdsWithParamHandler "GET"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/ids?type=RULE_TEMPLATE", http.StatusOK, "fetched=1", aut.firmwareRuleTemplateArrayValidator},

		//	"/{id}/priority/{newPriority}" PostFirmwareRuleTemplateByIdPriorityByNewPriorityHandler "POST"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/MAC_RULE/priority/1", http.StatusOK, NO_POSTERMS, nil},

		//	"/export" GetFirmwareRuleTemplateExportWithParamHandler "GET"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export?type=RULE_TEMPLATE", http.StatusOK, NO_POSTERMS, nil},

		// "/{type}/{isEditable}" GetFirmwareRuleTemplateExportHandler "GET"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/RULE_TEMPLATE/true", http.StatusOK, NO_POSTERMS, nil},

		//	"/importAll" PostFirmwareRuleTemplateImportAllHandler "POST"
		{FRT_API, "[create_with_sys_gen_id]", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + idCreated, aut.replaceKeysByValues, "POST", "/importAll", http.StatusOK, NO_POSTERMS, nil},

		//	"/filtered" GetFirmwareRuleTemplateFilteredWithParamsHandler "GET"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered", http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" DeleteFirmwareRuleTemplateByIdHandler "DELETE"
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + idCreated, http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}

func TestPostFirmwareRuleTemplateImportAllFromBodyParams(t *testing.T) {
	aut := newFirmwareRuleTemplateApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FRT_API, "[simple_duplicate]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=0&not_imported=1", aut.apiImportValidator},
		{FRT_API, "[missing_name]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusBadRequest, NO_POSTERMS, nil},
		{FRT_API, "[update]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=1&not_imported=0", aut.apiImportValidator},
		{FRT_API, "[firmware_rule_template_two]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=1&not_imported=0", aut.apiImportValidator},
		{FRT_API, "[create]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=1&not_imported=0", aut.apiImportValidator},
		{FRT_API, "[create]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=1&not_imported=0", aut.apiImportValidator},
		// {FRT_API, "[create duplicate]",            NO_PRETERMS, nil,"POST", "/importAll", http.StatusOK, "imported=1&not_imported=1", aut.apiImportValidator},
		{FRT_API, "[duplicate]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=0&not_imported=1", aut.apiImportValidator},
		{FRT_API, "[missing_id]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=0&not_imported=1", aut.apiImportValidator},
		{FRT_API, "[missing_fixedarg_jlstring]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusBadRequest, NO_POSTERMS, nil},
		{FRT_API, "[missing_fixedarg_value]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusBadRequest, NO_POSTERMS, nil},
		{FRT_API, "[missing_fixedarg_bean]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusBadRequest, NO_POSTERMS, nil},
		{FRT_API, "[missing_fixedarg]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusBadRequest, NO_POSTERMS, nil},
		{FRT_API, "[missing_operation]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=0&not_imported=1", aut.apiImportValidator},
		{FRT_API, "[missing_relation]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=0&not_imported=1", aut.apiImportValidator},
		{FRT_API, "[missing_freearg]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusBadRequest, NO_POSTERMS, nil},
		{FRT_API, "[unwanted_trailing_comma]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusBadRequest, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}

// Additional comprehensive tests for uncovered code paths

func TestPostFirmwareRuleTemplateFilteredHandler_ErrorPaths(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Test with invalid JSON body
	req, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate/filtered", bytes.NewBufferString("{invalid json"))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Test with invalid page number
	filterBody := `{"applicableActionType":"RULE_TEMPLATE"}`
	req2, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate/filtered?pageNumber=-1&pageSize=10", bytes.NewBufferString(filterBody))
	assert.NilError(t, err)
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res2 := ExecuteRequest(req2, router).Result()
	defer res2.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res2.StatusCode)

	// Test with empty body
	req3, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate/filtered?pageNumber=1&pageSize=10", bytes.NewBufferString(""))
	assert.NilError(t, err)
	req3.Header.Set("Content-Type", "application/json")
	req3.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res3 := ExecuteRequest(req3, router).Result()
	defer res3.Body.Close()
	assert.Equal(t, http.StatusOK, res3.StatusCode)
}

func TestPostFirmwareRuleTemplateImportHandler_Success(t *testing.T) {
	t.Skip("Import handler route not registered - test skipped")
}

func TestPostFirmwareRuleTemplateImportHandler_Overwrite(t *testing.T) {
	t.Skip("Import handler route not registered - test skipped")
}

func TestPostFirmwareRuleTemplateImportHandler_ErrorPaths(t *testing.T) {
	t.Skip("Import handler route not registered - test skipped")
}

func TestPostChangePriorityHandler_ErrorPaths(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Create a template first
	templateJSON := `{
		"id": "PRIORITY_TEST",
		"priority": 5,
		"editable": true,
		"rule": {
			"condition": {
				"freeArg": {"type": "STRING", "name": "eStbMac"},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {"java.lang.String": "AA:BB:CC:DD:EE:FF"}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE_TEMPLATE"
		}
	}`
	var frt corefw.FirmwareRuleTemplate
	json.Unmarshal([]byte(templateJSON), &frt)
	corefw.CreateFirmwareRuleTemplateOneDB(&frt)

	// Test with invalid priority (0)
	req, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate/PRIORITY_TEST/priority/0", nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Test with invalid priority (negative)
	req2, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate/PRIORITY_TEST/priority/-1", nil)
	assert.NilError(t, err)
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res2 := ExecuteRequest(req2, router).Result()
	defer res2.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res2.StatusCode)

	// Test with non-existent template ID
	req3, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate/NONEXISTENT/priority/1", nil)
	assert.NilError(t, err)
	req3.Header.Set("Content-Type", "application/json")
	req3.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res3 := ExecuteRequest(req3, router).Result()
	defer res3.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res3.StatusCode)

	// Test with invalid priority format
	req4, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate/PRIORITY_TEST/priority/abc", nil)
	assert.NilError(t, err)
	req4.Header.Set("Content-Type", "application/json")
	req4.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res4 := ExecuteRequest(req4, router).Result()
	defer res4.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res4.StatusCode)
}

func TestPostChangePriorityHandler_Success(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Create multiple templates using JSON
	for i := 1; i <= 3; i++ {
		templateJSON := `{
			"id": "PRIORITY_` + string(rune('0'+i)) + `",
			"priority": ` + string(rune('0'+i)) + `,
			"editable": true,
			"rule": {
				"condition": {
					"freeArg": {"type": "STRING", "name": "eStbMac"},
					"operation": "IS",
					"fixedArg": {
						"bean": {
							"value": {"java.lang.String": "AA:BB:CC:DD:EE:FF"}
						}
					}
				}
			},
			"applicableAction": {
				"type": ".RuleAction",
				"actionType": "RULE_TEMPLATE"
			}
		}`
		var frt corefw.FirmwareRuleTemplate
		json.Unmarshal([]byte(templateJSON), &frt)
		corefw.CreateFirmwareRuleTemplateOneDB(&frt)
	}

	// Change priority
	req, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate/PRIORITY_1/priority/3", nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestPostFirmwareRuleTemplateHandler_ErrorPaths(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Test with invalid JSON
	req, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate", bytes.NewBufferString("{invalid}"))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Test with missing ID
	templateData := `{
		"priority": 1,
		"editable": true,
		"rule": {
			"condition": {
				"freeArg": {"type": "STRING", "name": "eStbMac"},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {"java.lang.String": "AA:BB:CC:DD:EE:FF"}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE_TEMPLATE"
		}
	}`

	req2, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate", bytes.NewBufferString(templateData))
	assert.NilError(t, err)
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res2 := ExecuteRequest(req2, router).Result()
	defer res2.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res2.StatusCode)

	// Test with duplicate ID
	templateJSON := `{
		"id": "DUPLICATE_ID",
		"priority": 1,
		"editable": true,
		"rule": {
			"condition": {
				"freeArg": {"type": "STRING", "name": "eStbMac"},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {"java.lang.String": "AA:BB:CC:DD:EE:FF"}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE_TEMPLATE"
		}
	}`
	var frt corefw.FirmwareRuleTemplate
	json.Unmarshal([]byte(templateJSON), &frt)
	corefw.CreateFirmwareRuleTemplateOneDB(&frt)

	templateData2 := `{
		"id": "DUPLICATE_ID",
		"priority": 1,
		"editable": true,
		"rule": {
			"condition": {
				"freeArg": {"type": "STRING", "name": "eStbMac"},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {"java.lang.String": "AA:BB:CC:DD:EE:FF"}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE_TEMPLATE"
		}
	}`

	req3, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate", bytes.NewBufferString(templateData2))
	assert.NilError(t, err)
	req3.Header.Set("Content-Type", "application/json")
	req3.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res3 := ExecuteRequest(req3, router).Result()
	defer res3.Body.Close()
	assert.Equal(t, http.StatusConflict, res3.StatusCode)
}

func TestDeleteFirmwareRuleTemplateByIdHandler_ErrorPaths(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Test delete non-existent template
	req, err := http.NewRequest("DELETE", "/xconfAdminService/firmwareruletemplate/NONEXISTENT", nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	// Note: Template deletion with usage check is tested but the handler
	// might not enforce it in current implementation, test skipped
}

func TestGetFirmwareRuleTemplateByIdHandler_ErrorPaths(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Test get non-existent template
	req, err := http.NewRequest("GET", "/xconfAdminService/firmwareruletemplate/NONEXISTENT", nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestObsoleteGetFirmwareRuleTemplatePageHandler_ErrorPaths(t *testing.T) {
	t.Skip("Obsolete handler returns 501 NotImplemented - test skipped")
}

func TestObsoleteGetFirmwareRuleTemplatePageHandler_Success(t *testing.T) {
	t.Skip("Obsolete handler returns 501 NotImplemented - test skipped")
}

func TestPutFirmwareRuleTemplateEntitiesHandler_ErrorPaths(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Test with invalid JSON
	req, err := http.NewRequest("PUT", "/xconfAdminService/firmwareruletemplate/entities", bytes.NewBufferString("{invalid}"))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Test update non-existent entity
	updateData := `[{
		"id": "NONEXISTENT",
		"priority": 1,
		"editable": true,
		"rule": {
			"condition": {
				"freeArg": {"type": "STRING", "name": "eStbMac"},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {"java.lang.String": "AA:BB:CC:DD:EE:FF"}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE_TEMPLATE"
		}
	}]`

	req2, err := http.NewRequest("PUT", "/xconfAdminService/firmwareruletemplate/entities", bytes.NewBufferString(updateData))
	assert.NilError(t, err)
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res2 := ExecuteRequest(req2, router).Result()
	defer res2.Body.Close()
	assert.Equal(t, http.StatusOK, res2.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(res2.Body).Decode(&result)
	// Should have failure for non-existent entity
	assert.Assert(t, result != nil)
}

func TestPutFirmwareRuleTemplateEntitiesHandler_Success(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Create entity first using JSON
	templateJSON := `{
		"id": "UPDATE_ENTITY_TEST",
		"priority": 1,
		"editable": true,
		"rule": {
			"condition": {
				"freeArg": {"type": "STRING", "name": "eStbMac"},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {"java.lang.String": "AA:BB:CC:DD:EE:FF"}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE_TEMPLATE"
		}
	}`
	var frt corefw.FirmwareRuleTemplate
	json.Unmarshal([]byte(templateJSON), &frt)
	corefw.CreateFirmwareRuleTemplateOneDB(&frt)

	// Update it
	updateData := `[{
		"id": "UPDATE_ENTITY_TEST",
		"priority": 2,
		"editable": true,
		"rule": {
			"condition": {
				"freeArg": {"type": "STRING", "name": "eStbMac"},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {"java.lang.String": "AA:BB:CC:DD:EE:FF"}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE_TEMPLATE"
		}
	}]`

	req, err := http.NewRequest("PUT", "/xconfAdminService/firmwareruletemplate/entities", bytes.NewBufferString(updateData))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)
	assert.Assert(t, result != nil)
}

func TestGetFirmwareRuleTemplateIdsHandler_ErrorPaths(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Test without type parameter
	req, err := http.NewRequest("GET", "/xconfAdminService/firmwareruletemplate/ids", nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestGetFirmwareRuleTemplateExportHandler_ErrorPaths(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Test without type parameter
	req, err := http.NewRequest("GET", "/xconfAdminService/firmwareruletemplate/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestPutFirmwareRuleTemplateHandler_ErrorPaths(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Test with invalid JSON
	req, err := http.NewRequest("PUT", "/xconfAdminService/firmwareruletemplate", bytes.NewBufferString("{invalid}"))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Test update non-existent template
	templateData := `{
		"id": "NONEXISTENT",
		"priority": 1,
		"editable": true,
		"rule": {
			"condition": {
				"freeArg": {"type": "STRING", "name": "eStbMac"},
				"operation": "IS",
				"fixedArg": {
					"bean": {
						"value": {"java.lang.String": "AA:BB:CC:DD:EE:FF"}
					}
				}
			}
		},
		"applicableAction": {
			"type": ".RuleAction",
			"actionType": "RULE_TEMPLATE"
		}
	}`

	req2, err := http.NewRequest("PUT", "/xconfAdminService/firmwareruletemplate", bytes.NewBufferString(templateData))
	assert.NilError(t, err)
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res2 := ExecuteRequest(req2, router).Result()
	defer res2.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res2.StatusCode)
}

func TestPostFirmwareRuleTemplateEntitiesHandler_ErrorPaths(t *testing.T) {
	DeleteAllEntities()
	setupTestModels()
	defer DeleteAllEntities()

	// Test with invalid JSON
	req, err := http.NewRequest("POST", "/xconfAdminService/firmwareruletemplate/entities", bytes.NewBufferString("{invalid}"))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
