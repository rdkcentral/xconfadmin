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

	"github.com/rdkcentral/xconfwebconfig/shared"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"gotest.tools/assert"
)

const (
	//	FR_API                       = "/xconfAdminService/firmwarerule"
	jsonFirmwareRuleTestDataLocn = "jsondata/firmwarerule/"
)

func newFirmwareRuleApiUnitTest(t *testing.T) *apiUnitTest {
	aut := newApiUnitTest(t)
	aut.setupFirmwareRuleApi()
	return aut
}

func (aut *apiUnitTest) setupFirmwareRuleApi() {
	if aut.getValOf(FR_API) == "Done" {
		return
	}
	aut.setValOf(FR_API+DATA_LOCN_SUFFIX, jsonFirmwareRuleTestDataLocn)

	aut.setupFirmwareConfigApi()
	configTestCases := []apiUnitTestCase{
		{FC_API, "firmware_config_two", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FC_API, "firmware_config_one", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FC_API, "firmware_config_three", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
	}
	aut.run(configTestCases)

	aut.setupFirmwareRuleTemplateApi()
	frtTestCases := []apiUnitTestCase{
		{FRT_API, "firmware_rule_template_iprule", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FRT_API, "firmware_rule_template_ivrule", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
	}
	aut.run(frtTestCases)

	aut.setValOf(FR_API, "Done")
}

func (aut *apiUnitTest) cleanupFirmwareRuleApi() {
	if aut.getValOf(FR_API) == "" {
		return
	}

	frtTestCases := []apiUnitTestCase{
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/IP_RULE", http.StatusNoContent, NO_POSTERMS, nil}, // TODO Should not be able to delete template if there are rules dependent on it
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/IV_RULE", http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(frtTestCases)

	configTestCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/de529a04-3bab-41e3-ad79-f1e583723b47", http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/393e2152-9d50-4f30-aab9-c74977471632", http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/e4b10a02-094b-4941-8aee-6b10a996829d", http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(configTestCases)
	aut.setValOf(FR_API, "")
}

func (aut *apiUnitTest) firmwareRuleResponseValidator(tcase apiUnitTestCase, genRsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(genRsp.Body)

	assert.Equal(aut.t, tcase.api, FR_API)
	var rsp = corefw.FirmwareRule{}
	json.Unmarshal(rspBody, &rsp)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	aut.saveIdIn(kvMap, rsp.ID)

	val, ok := kvMap["validate"]
	if !ok || val[0] != "true" {
		return
	}

	req := corefw.NewEmptyFirmwareRule()
	reqBodyBytes, _ := ioutil.ReadAll(reqBody)
	err = json.Unmarshal(reqBodyBytes, &req)
	assert.NilError(aut.t, err)
	if req.ID != "" {
		assert.Equal(aut.t, rsp.ID, req.ID)
	}
	/*
		assert.Equal (aut.t, rsp.Description, req.Description)
		assert.Equal (aut.t, rsp.FirmwareFilename, req.FirmwareFilename)
		assert.Equal (aut.t, rsp.FirmwareVersion, req.FirmwareVersion)
		assert.Equal (aut.t, IsEqual (req.SupportedModelIds, rsp.SupportedModelIds), true)
	*/
}

func (aut *apiUnitTest) modelArrayValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, tcase.api == MODEL_QAPI || tcase.api == MODEL_WHOLE_API, true)

	var entries = []shared.Model{}
	json.Unmarshal(rspBody, &entries)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	aut.assertFetched(kvMap, len(entries))
	aut.saveFetchedCntIn(kvMap, len(entries))
}

func (aut *apiUnitTest) firmwareRuleArrayValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, tcase.api, FR_API)
	var entries = []corefw.FirmwareRule{}
	json.Unmarshal(rspBody, &entries)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

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

func (aut *apiUnitTest) firmwareRuleSingleValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, tcase.api, FR_API)

	var entry = corefw.FirmwareRule{}
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

func TestGetFirmwareRuleFromQueryParams(t *testing.T) {
	aut := newFirmwareRuleApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleArrayValidator},

		// Ignore invalid Param
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?invalidParam=someValue", http.StatusOK, NO_POSTERMS, nil},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		// Happy Paths
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.getValOf("begin_count"), aut.firmwareRuleArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleFilteredFromQueryParams(t *testing.T) {

	aut := newFirmwareRuleApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FR_API, "firmware_rule_one", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "firmware_rule_two", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "firmware_rule_four", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "firmware_rule_three", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleArrayValidator},
	}
	aut.run(testCases)

	stPt := aut.getValOf("begin_count")

	testCases = []apiUnitTestCase{
		// Errors: missing mandatory param. Currently fallback for applicationType is stb. So the below 3 will not fail
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType", http.StatusOK, NO_POSTERMS, nil},

		// Invalid param are ignored. So no error
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?invalidParam=someValue", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&invalidParam=someValue", http.StatusOK, "fetched=" + stPt, aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&invalidParam=someValue&another=value", http.StatusOK, "fetched=" + stPt, aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&invalidParam=someValue", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&invalidParam=someValue&another=value", http.StatusOK, NO_POSTERMS, nil},

		// Errors: missing value for param
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&name=", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&name", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&key=", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&key", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&value=", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&value", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&templateId=", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&templateId", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&FIRMWARE_VERSION=", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&FIRMWARE_VERSION", http.StatusOK, NO_POSTERMS, nil},

		// Happy paths: Duplicate params (second value is ignored)
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&applicationType=stb&applicationType=json", http.StatusOK, "fetched=" + stPt, aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=1-3939&applicationType=stb&name=second", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?key=eStbMac&applicationType=stb&key=second", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?value=1717_LED_ABCD&applicationType=stb&value=second", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?templateId=IP_RULE_1&applicationType=stb&templateId=second", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?FIRMWARE_VERSION=unit&applicationType=stb&FIRMWARE_VERSION=second", http.StatusOK, "fetched=4", aut.firmwareRuleArrayValidator},

		// applicationType - Happy Paths
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=nonexistant", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb", http.StatusOK, "fetched=" + stPt, aut.firmwareRuleArrayValidator},
		// Change Case
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=STB", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},

		// name - Happy Paths
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&name=nonexistant", http.StatusOK, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&name=1-3939", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&name=1717_LED_ABC23", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&name=000ipPerformanceTestRule", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		// Case sensitivity
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&name=000ipPERFORMANCETESTRULE", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		// partial representation for name
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&name=000ipPERFORMANCETEST", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},

		// key - Happy Paths
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&key=nonexistant", http.StatusOK, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&key=eStbMac", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&key=ipAddress", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		// Case sensitivity
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&key=ipADDRESS", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		// partial representation for key
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&key=ipADDR", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},

		// value - Happy Paths
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&value=nonexistant", http.StatusOK, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&value=1717_LED_ABCD", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		// Case sensitiity
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&value=1717_LED_ABCD", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		// partial representation for value
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&value=1717_LED", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},

		// templateId - Happy Paths
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&templateId=nonexistant", http.StatusOK, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&templateId=IV_RULE_1", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&templateId=IP_RULE_1", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&templateId=MAC_RULE", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},

		// FIRMWARE_VERSION - Happy Paths
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&FIRMWARE_VERSION=nonexistant", http.StatusOK, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&FIRMWARE_VERSION=firmware_config_unit", http.StatusOK, "fetched=4", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=stb&FIRMWARE_VERSION=firmware_config_unit_test_1", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},

		// Happy paths- order of params reversed
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?name=1-3939&applicationType=stb", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?key=eStbMac&applicationType=stb", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?value=1717_LED_ABCD&applicationType=stb", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?templateId=IP_RULE_1&applicationType=stb", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?FIRMWARE_VERSION=firmware_config_unit&applicationType=stb", http.StatusOK, "fetched=4", aut.firmwareRuleArrayValidator},
	}
	aut.run(testCases)

	frTestCases := []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/64a19e12-21d0-4a72-9f0e-346fa53c3c68", http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/e05a5b92-8605-4309-bfe5-25646e888137", http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/64a19e12-21d0-4a72-9f0e-346fa53c3c67", http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/aa534186-ef60-4516-8c47-c254f9066c22", http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(frTestCases)
}

func TestPostFirmwareRuleFilteredFromQueryParams(t *testing.T) {

	aut := newFirmwareRuleApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FR_API, "firmware_rule_one", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "firmware_rule_two", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "firmware_rule_four", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "firmware_rule_three", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		// invalid query params are ignored
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/filtered?name=dummy", http.StatusOK, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/filtered?pageNum=1", http.StatusOK, NO_POSTERMS, nil},

		// Happy Paths
		{FR_API, "rule", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=4", http.StatusOK, "fetched=4", aut.firmwareRuleArrayValidator},
		{FR_API, "define_properties", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=2", http.StatusOK, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, "blocking_filter", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=2", http.StatusOK, "fetched=0", aut.firmwareRuleArrayValidator},

		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=2", http.StatusOK, "fetched=2", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=2&pageSize=3", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=0&pageSize=3", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=-1&pageSize=3", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=4", http.StatusOK, "fetched=4", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=5", http.StatusOK, "fetched=4", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=0", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=-1", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered", http.StatusOK, "fetched=4", aut.firmwareRuleArrayValidator},

		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=A&pageSize=B", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=A&pageSize=1", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=B", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},

		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=&pageSize=", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber= &pageSize= ", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=&pageSize= ", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber= &pageSize=", http.StatusBadRequest, "fetched=0", aut.firmwareRuleArrayValidator},

		// Happy Paths: default value for missing query params
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1", http.StatusOK, "fetched=4", aut.firmwareRuleArrayValidator},
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageSize=3", http.StatusOK, "fetched=3", aut.firmwareRuleArrayValidator},
	}
	aut.run(testCases)

	frTestCases := []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/64a19e12-21d0-4a72-9f0e-346fa53c3c68", http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/e05a5b92-8605-4309-bfe5-25646e888137", http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/64a19e12-21d0-4a72-9f0e-346fa53c3c67", http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/aa534186-ef60-4516-8c47-c254f9066c22", http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(frTestCases)
}

func TestGetFirmwareRuleById(t *testing.T) {
	aut := newFirmwareRuleApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FR_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleResponseValidator},
	}
	aut.run(testCases)
	testCases = []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + aut.getValOf("id_1"), http.StatusOK, "ID=" + aut.getValOf("id_1"), aut.firmwareRuleSingleValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + aut.getValOf("id_1"), http.StatusNotFound, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}

func TestFirmwareRuleCRUD(t *testing.T) {
	aut := newFirmwareRuleApiUnitTest(t)
	sysGenId := uuid.New().String()
	testCases := []apiUnitTestCase{
		{FR_API, "missing_free_arg", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},

		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=rdkcloud", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleArrayValidator},

		{FR_API, "define_props", NO_PRETERMS, nil, "PUT", "", http.StatusNotFound, NO_POSTERMS, nil},
		//applicationType=rdkcloud
		{FR_API, "define_props", NO_PRETERMS, nil, "POST", "?applicationType=rdkcloud", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleResponseValidator},
		{FR_API, "define_props", NO_PRETERMS, nil, "POST", "?applicationType=rdkcloud", http.StatusConflict, NO_POSTERMS, nil},
		{FR_API, "define_props", NO_PRETERMS, nil, "PUT", "?applicationType=rdkcloud", http.StatusOK, NO_POSTERMS, nil},
		// applicationType=stb
		{FR_API, "create_to_change_app_type", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_2", aut.firmwareRuleResponseValidator},
		// applicationType=stb
		{FR_API, "duplicate", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_3", aut.firmwareRuleResponseValidator},

		//applicationType=json
		{FR_API, "update_to_change_app_type", NO_PRETERMS, nil, "PUT", "", http.StatusConflict, NO_POSTERMS, nil},
		{FR_API, "missing_free_arg", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FR_API, "unwanted_trailing_comma", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FR_API, "create_missing_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_5", aut.firmwareRuleResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=rdkcloud", http.StatusOK, "fetched=" + aut.eval("begin_count+1"), aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1") + "?applicationType=rdkcloud", http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_2"), http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_3"), http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNotFound, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_2"), http.StatusNotFound, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_3"), http.StatusNotFound, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_5"), http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=rdkcloud", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareRuleArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleByIdWithExportParam(t *testing.T) {
	aut := newFirmwareRuleApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FR_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleResponseValidator},
	}
	aut.run(testCases)
	testCases = []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + aut.getValOf("id_1") + "?export", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + aut.getValOf("id_1") + "?export", http.StatusNotFound, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleWithParam(t *testing.T) {
	aut := newFirmwareRuleApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleArrayValidator},
		{FR_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, "fetched=" + aut.eval("begin_count +1") + "&validate_export=true", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, "fetched=" + aut.eval("begin_count") + "&validate_export=true", aut.firmwareRuleArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleExportAllTypesWithParam(t *testing.T) {
	aut := newFirmwareRuleApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export/allTypes?exportAll", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleArrayValidator},
		{FR_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export/allTypes?exportAll", http.StatusOK, "fetched=" + aut.eval("begin_count +1") + "&validate_export=true", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export/allTypes?exportAll", http.StatusOK, "fetched=" + aut.eval("begin_count") + "&validate_export=true", aut.firmwareRuleArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleExportByTypeWithParam(t *testing.T) {
	aut := newFirmwareRuleApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export/byType?exportAll", http.StatusBadRequest, "", nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export/byType", http.StatusBadRequest, "", nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export/byType?exportAll&type=RULE", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleArrayValidator},
		{FR_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export/byType?exportAll&type=RULE", http.StatusOK, "fetched=" + aut.eval("begin_count +1") + "&validate_export=true", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export/byType?exportAll&type=RULE", http.StatusOK, "fetched=" + aut.eval("begin_count") + "&validate_export=true", aut.firmwareRuleArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleByTypeNames(t *testing.T) {
	aut := newFirmwareRuleApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/MAC_RULE/names", http.StatusOK, "saveFetchedCntIn=begin_count", aut.apiNameMapValidator},
		{FR_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/MAC_RULE/names", http.StatusOK, "fetched=" + aut.eval("begin_count +1"), aut.apiNameMapValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/MAC_RULE/names", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.apiNameMapValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwareRuleByTemplateByTemplateIdNames(t *testing.T) {
	aut := newFirmwareRuleApiUnitTest(t)
	sysGenId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/byTemplate/MAC_RULE/names", http.StatusOK, "saveFetchedCntIn=begin_count", aut.apiNameListValidator},
		{FR_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareRuleResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/byTemplate/MAC_RULE/names", http.StatusOK, "fetched=" + aut.eval("begin_count +1"), aut.apiNameListValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/byTemplate/MAC_RULE/names", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.apiNameListValidator},
	}
	aut.run(testCases)
}

func TestFirmwareRuleEndPoints(t *testing.T) {

	aut := newFirmwareRuleApiUnitTest(t)
	sysGenId := uuid.New().String()
	sysGenId2 := uuid.New().String()

	testCases := []apiUnitTestCase{
		// "" PostFirmwareRuleHandler "POST"
		{FR_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=fr_id_1", aut.firmwareRuleResponseValidator},
	}
	aut.run(testCases)

	idCreated := aut.getValOf("fr_id_1")
	testCases = []apiUnitTestCase{
		// "" PutFirmwareRuleHandler "PUT"
		{FR_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + idCreated, aut.replaceKeysByValues, "PUT", "", http.StatusOK, NO_POSTERMS, nil},

		//	"" GetFirmwareRuleHandler "GET"
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, NO_POSTERMS, nil},

		// "/entities" PostFirmwareRuleEntitiesHandler "POST"
		{FR_API, "[create_with_sys_gen_id]", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "POST", "/entities", http.StatusOK, NO_POSTERMS, nil},

		//	"/entities" PutFirmwareRuleEntitiesHandler "PUT"
		{FR_API, "[create_with_sys_gen_id]", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "PUT", "/entities", http.StatusOK, NO_POSTERMS, nil},

		// "",  GetFirmwareRuleWithParamHandler "GET"
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, NO_POSTERMS, nil},

		// "/{id}" GetFirmwareRuleByIdWithParamHandler "GET"
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + idCreated + "?export", http.StatusOK, NO_POSTERMS, nil},

		// "/{id}" GetFirmwareRuleByIdHandler "GET"
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + idCreated, http.StatusOK, "ID=" + idCreated, aut.firmwareRuleSingleValidator},

		// "/filtered" GetFirmwareRuleFilteredWithParamsHandler "GET"
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=rdkcloud&name=somenewname", http.StatusOK, NO_POSTERMS, nil},

		// No registered handler
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/somenewname?unknown", http.StatusNotFound, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/import/", http.StatusNotFound, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "POST", "/import", http.StatusMethodNotAllowed, NO_POSTERMS, nil}, // Should be StatusNotFound as per java

		//  "/filtered", queries.PostFirmwareRuleFilteredWithParamsHandler "POST"
		{FR_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=10", http.StatusOK, NO_POSTERMS, nil},

		// "/{type}/names" GetFirmwareRuleByTypeNamesHandler "GET"
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/MAC_RULE/names", http.StatusOK, NO_POSTERMS, nil},

		// "/byTemplate/{templateId}/names" GetFirmwareRuleByTemplateByTemplateIdNamesHandler "GET"
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/byTemplate/MAC_RULE/names", http.StatusOK, NO_POSTERMS, nil},

		// "/export/byType" GetFirmwareRuleExportByTypeWithParamHandler "GET"
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export/byType?exportAll&type=RULE", http.StatusOK, NO_POSTERMS, nil},

		// "/export/allTypes" GetFirmwareRuleExportAllTypesWithParamHandler "GET"
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/export/allTypes?exportAll", http.StatusOK, NO_POSTERMS, nil},

		// "/importAll" PostFirmwareRuleImportAllHandler "POST"
		{FR_API, "[create_with_sys_gen_id]", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + idCreated, aut.replaceKeysByValues, "POST", "/importAll", http.StatusOK, NO_POSTERMS, nil},

		// "/{id}" DeleteFirmwareRuleByIdHandler "DELETE"
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + idCreated, http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}

func TestFirmwareRuleCRUDInLoop(t *testing.T) {

	aut := newFirmwareRuleApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FR_API, "define_props", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=rdkcloud&name=somenewname", http.StatusOK, "fetched=1", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/36be74c7-f3fc-4fb9-ac98-980810033372", http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/filtered?applicationType=rdkcloud&name=somenewname", http.StatusOK, "fetched=0", aut.firmwareRuleArrayValidator},
	}
	numTimes := 1
	for i := 1; i < numTimes; i++ {
		aut.run(testCases)
	}
}

func TestPostFirmwareRuleImportAllFromBodyParams(t *testing.T) {

	aut := newFirmwareRuleApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FR_API, "[missing_free_arg define_props]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusBadRequest, NO_POSTERMS, nil},
		{FR_API, "[create_to_change_app_type]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=1&not_imported=0", aut.apiImportValidator},
		{FR_API, "[update_to_change_app_type]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=0&not_imported=1", aut.apiImportValidator},
		{FR_API, "[missing_free_arg]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=0&not_imported=1", aut.apiImportValidator},
		{FR_API, "[missing_fixed_arg]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=0&not_imported=1", aut.apiImportValidator},
		{FR_API, "[define_props]", NO_PRETERMS, nil, "POST", "/importAll?applicationType=rdkcloud", http.StatusOK, "imported=1&not_imported=0", aut.apiImportValidator},
		{FR_API, "[update]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=1&not_imported=0", aut.apiImportValidator},
		{FR_API, "[create]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=1&not_imported=0", aut.apiImportValidator},
		{FR_API, "[update]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=1&not_imported=0", aut.apiImportValidator},
		{FR_API, "[create]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=1&not_imported=0", aut.apiImportValidator},
		{FR_API, "[duplicate]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusOK, "imported=0&not_imported=1", aut.apiImportValidator},
		{FR_API, "[unwanted_trailing_comma]", NO_PRETERMS, nil, "POST", "/importAll", http.StatusBadRequest, NO_POSTERMS, nil},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/36be74c7-f3fc-4fb9-ac98-980810044472", http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/NEW_RULE_WITH_NEW_NAME", http.StatusNoContent, NO_POSTERMS, nil},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/CREATE_TO_CHANGE_APP_TYPE", http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}

func TestApplicationType(t *testing.T) {

	sysGenId1 := uuid.New().String()
	sysGenId2 := uuid.New().String()
	sysGenId3 := uuid.New().String()
	sysGenId4 := uuid.New().String()
	sysGenId5 := uuid.New().String()
	sysGenId6 := uuid.New().String()

	aut := newFirmwareRuleApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?applicationType=stb", http.StatusOK, "saveFetchedCntIn=stb_count", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?applicationType=rdkcloud", http.StatusOK, "saveFetchedCntIn=rdkcloud_count", aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?applicationType=rdkcloud", http.StatusOK, "saveFetchedCntIn=rdkcloud_count", aut.firmwareRuleArrayValidator},

		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetchedCnt=" + aut.getValOf("stb_count"), aut.firmwareRuleArrayValidator},

		{FR_API, "create_with_sys_gen_id_for_app_type", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2 + "&SUPPLIED_APPLICATION_TYPE=stb", aut.replaceKeysByValues, "POST", "?applicationType=stb", http.StatusConflict, NO_POSTERMS, nil},
		{FR_API, "create_with_sys_gen_id_for_app_type", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId1 + "&SUPPLIED_APPLICATION_TYPE=stb", aut.replaceKeysByValues, "POST", "?applicationType=stb", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "create_with_sys_gen_id_for_app_type", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId3 + "&SUPPLIED_APPLICATION_TYPE=rdkcloud", aut.replaceKeysByValues, "POST", "?applicationType=stb", http.StatusConflict, NO_POSTERMS, nil},
		{FR_API, "create_with_sys_gen_id_for_app_type", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId4 + "&SUPPLIED_APPLICATION_TYPE=rdkcloud", aut.replaceKeysByValues, "POST", "?applicationType=stb", http.StatusConflict, NO_POSTERMS, nil},
		{FR_API, "create_with_sys_gen_id_for_app_type", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId5 + "&SUPPLIED_APPLICATION_TYPE=", aut.replaceKeysByValues, "POST", "?applicationType=stb", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "create_with_sys_gen_id_for_app_type", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId5 + "&SUPPLIED_APPLICATION_TYPE=stb", aut.replaceKeysByValues, "POST", "?applicationType=rdkcloud", http.StatusConflict, NO_POSTERMS, nil},
		// applictionTypes match between user and object but not with the assoicated firmwareconfig
		{FR_API, "create_with_sys_gen_id_for_app_type", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId6 + "&SUPPLIED_APPLICATION_TYPE=stb", aut.replaceKeysByValues, "POST", "?applicationType=stb", http.StatusBadRequest, NO_POSTERMS, nil},

		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?applicationType=stb", http.StatusOK, "fetchedCnt=" + aut.getValOf("stb_count+2"), aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?applicationType=stb", http.StatusOK, "fetchedCnt=" + aut.getValOf("stb_count+2"), aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?applicationType=rdkcloud", http.StatusOK, "fetchedCnt=" + aut.getValOf("rdkcloud_count+1"), aut.firmwareRuleArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?applicationType=rdkcloud", http.StatusOK, "fetchedCnt=" + aut.getValOf("rdkcloud_count+1"), aut.firmwareRuleArrayValidator},
	}
	aut.run(testCases)
}

func TestOrderDifferentButEqualConditionsInFRCreation(t *testing.T) {

	sysGenId1 := uuid.New().String()
	sysGenId2 := uuid.New().String()

	aut := newFirmwareRuleApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FRT_API, "RI_MACLIST", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "complex_rule_one", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId1, aut.replaceKeysByValues, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "complex_rule_two", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "POST", "", http.StatusConflict, NO_POSTERMS, nil},
	}
	aut.run(testCases)
	testCases = []apiUnitTestCase{
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + sysGenId1, http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}
