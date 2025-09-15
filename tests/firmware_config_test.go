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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"

	assert "gotest.tools/assert"
)

const (
	FC_API                         = "/xconfAdminService/firmwareconfig"
	jsonFirmwareConfigTestDataLocn = "jsondata/firmwareconfig/"
)

func newFirmwareConfigApiUnitTest(t *testing.T) *apiUnitTest {
	aut := newApiUnitTest(t)
	aut.setupFirmwareConfigApi()
	return aut
}

func TestValidateUsageBeforeRemoving(t *testing.T) {
	DeleteAllEntities()
	percentageBean, err := PreCreatePercentageBean()
	assert.NilError(t, err)
	firmwareConfig, _ := coreef.GetFirmwareConfigOneDB(percentageBean.LastKnownGood)

	url := fmt.Sprintf("/xconfAdminService/delete/firmwares/%v?&applicationType=stb", percentageBean.LastKnownGood)

	r := httptest.NewRequest("DELETE", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusConflict, rr.Code)

	xconfError := unmarshalXconfError(rr.Body.Bytes())

	assert.Equal(t, fmt.Sprintf("FirmwareConfig %v is used by %v rule", firmwareConfig.Description, percentageBean.Name), xconfError.Message)
	DeleteAllEntities()
}

func (aut *apiUnitTest) setupFirmwareConfigApi() {
	if aut.getValOf(FC_API) == "Done" {
		return
	}
	aut.setValOf(FC_API+DATA_LOCN_SUFFIX, jsonFirmwareConfigTestDataLocn)
	aut.setupModelApi()
	testCases := []apiUnitTestCase{
		{MODEL_UAPI, "DPC8888", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{MODEL_UAPI, "DPC8888T", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{MODEL_UAPI, "DPC9999", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{MODEL_UAPI, "DPC9999T", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
	}
	aut.run(testCases)
	aut.setValOf(FC_API, "Done")
}

func (aut *apiUnitTest) cleanupFirmwareConfigApi() {
	if aut.getValOf(FC_API) == "" {
		return
	}
	testCases := []apiUnitTestCase{
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/DPC8888", http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/DPC8888T", http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/DPC9999", http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/DPC9999T", http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
	aut.setValOf(FC_API, "")
}

func (aut *apiUnitTest) firmwareConfigArrayValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, tcase.api == FC_API || tcase.api == FWS_QAPI, true)

	var entries = []coreef.FirmwareConfig{}
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

func (aut *apiUnitTest) firmwareConfigMapValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, tcase.api, FC_API)

	var entries = map[string]coreef.FirmwareConfig{}
	json.Unmarshal(rspBody, &entries)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	aut.assertFetched(kvMap, len(entries))
	_, ok := kvMap["fetched"]
	if ok {
		for k, v := range entries {
			assert.Equal(aut.t, k, v.ID)
		}
	}
	aut.saveFetchedCntIn(kvMap, len(entries))
}

func (aut *apiUnitTest) saveExisted(kvMap map[string][]string, existedCnt int) {
	entry, ok := kvMap["saveExisted"]
	if ok {
		aut.savedMap[entry[0]] = strconv.Itoa(existedCnt)
	}
}

func (aut *apiUnitTest) saveNotExisted(kvMap map[string][]string, notExistedCnt int) {
	entry, ok := kvMap["saveNotExisted"]
	if ok {
		aut.savedMap[entry[0]] = strconv.Itoa(notExistedCnt)
	}
}

func (aut *apiUnitTest) assertExisted(kvMap map[string][]string, existedCnt int) {
	entry, ok := kvMap["fetchExisted"]
	if ok {
		expEntries, _ := strconv.Atoi(entry[0])
		assert.Equal(aut.t, existedCnt, expEntries)
	}
}

func (aut *apiUnitTest) assertNotExisted(kvMap map[string][]string, existedCnt int) {
	entry, ok := kvMap["fetchNotExisted"]
	if ok {
		expEntries, _ := strconv.Atoi(entry[0])
		assert.Equal(aut.t, existedCnt, expEntries)
	}
}

func (aut *apiUnitTest) firmwareVersionMapValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, tcase.api, FC_API)

	versionMap := make(map[string][]string)
	json.Unmarshal(rspBody, &versionMap)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	aut.saveExisted(kvMap, len(versionMap["existedVersions"]))
	aut.saveNotExisted(kvMap, len(versionMap["notExistedVersions"]))
	aut.assertExisted(kvMap, len(versionMap["existedVersions"]))
	aut.assertNotExisted(kvMap, len(versionMap["notExistedVersions"]))
}

func (aut *apiUnitTest) firmwareConfigSingleValidator(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, tcase.api == FC_API || tcase.api == FWS_QAPI, true)

	var entry = coreef.FirmwareConfig{}
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

func IsEqual(a1 []string, a2 []string) bool {
	sort.Strings(a1)
	sort.Strings(a2)
	if len(a1) == len(a2) {
		for i, v := range a1 {
			if v != a2[i] {
				return false
			}
		}
	} else {
		return false
	}
	return true
}

func (aut *apiUnitTest) firmwareConfigResponseValidator(tcase apiUnitTestCase, genRsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(genRsp.Body)
	assert.Equal(aut.t, tcase.api == FC_API || tcase.api == FWS_UAPI, true)
	var rsp = coreef.FirmwareConfigResponse{}
	json.Unmarshal(rspBody, &rsp)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	aut.saveIdIn(kvMap, rsp.ID)

	if aut.getValOf("validate") != "true" {
		return
	}

	req := coreef.NewEmptyFirmwareConfig()
	reqBodyBytes, _ := ioutil.ReadAll(reqBody)
	err = json.Unmarshal(reqBodyBytes, &req)
	assert.NilError(aut.t, err)
	if req.ID != "" {
		assert.Equal(aut.t, rsp.ID, req.ID)
	}
	assert.Equal(aut.t, rsp.Description, req.Description)
	assert.Equal(aut.t, rsp.FirmwareFilename, req.FirmwareFilename)
	assert.Equal(aut.t, rsp.FirmwareVersion, req.FirmwareVersion)
	assert.Equal(aut.t, IsEqual(req.SupportedModelIds, rsp.SupportedModelIds), true)
}
func (aut *apiUnitTest) baseEntityCount(t *testing.T, heading string) {
	testCases := []apiUnitTestCase{
		{MODEL_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=MODEL_count", aut.modelArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=FC_count", aut.firmwareConfigArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=FR_count", aut.firmwareRuleArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=FRT_count", aut.firmwareRuleTemplateArrayValidator},
	}
	aut.run(testCases)
	log.Println("At " + heading + t.Name())
	log.Println("Model Count=" + aut.getValOf("MODEL_count"))
	log.Println("FC Count=" + aut.getValOf("FC_count"))
	log.Println("FR Count=" + aut.getValOf("FR_count"))
	log.Println("FRT Count=" + aut.getValOf("FRT_count"))
}

// ""
func TestGetFirmwareConfig(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)
	aut.baseEntityCount(t, " begin:")
	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FC_API, "firmware_config_one", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareConfigResponseValidator},
		{FC_API, "firmware_config_two", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_2", aut.firmwareConfigResponseValidator},
		{FC_API, "firmware_config_three", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_3", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count +3"), aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_2"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_3"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// ""
func TestPostFirmwareConfig(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		// Error cases
		{FC_API, "missing_application_type", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "missing_description", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "missing_firmware_filename", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "missing_firmware_version", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "model_not_present", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},

		//System should generate an ID, if one is not supplied
		{FC_API, "missing_id", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1&validate=true", aut.firmwareConfigResponseValidator},
		// Create a FirmwareConfig.
		{FC_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_2&validate=true", aut.firmwareConfigResponseValidator},
		// Creating another one with the same id should fail
		{FC_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusConflict, NO_POSTERMS, nil},
	}
	aut.run(testCases)

	// Cleanup to make the test idempotent
	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_2"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// ""
func TestPutFirmwareConfig(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		// Error cases
		{FC_API, "missing_application_type", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "missing_description", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "missing_firmware_filename", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "missing_firmware_version", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "model_not_present", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "missing_id", NO_PRETERMS, nil, "PUT", "", http.StatusNotFound, NO_POSTERMS, nil},

		// Create a new FirmwareConfig Entry
		{FC_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1&validate=true", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	// Update the FirmwareConfig just created, changing only one content at a time
	testCases = []apiUnitTestCase{
		{FC_API, "create_update_app", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "create_update_desc", NO_PRETERMS, nil, "PUT", "", http.StatusOK, "validate=true", aut.firmwareConfigResponseValidator},
		{FC_API, "create_update_fw_filename", NO_PRETERMS, nil, "PUT", "", http.StatusOK, "validate=true", aut.firmwareConfigResponseValidator},
		{FC_API, "create_update_fw_version", NO_PRETERMS, nil, "PUT", "", http.StatusOK, "validate=true", aut.firmwareConfigResponseValidator},
		{FC_API, "create_update_model", NO_PRETERMS, nil, "PUT", "", http.StatusOK, "validate=true", aut.firmwareConfigResponseValidator},
		{FC_API, "create", NO_PRETERMS, nil, "PUT", "", http.StatusOK, "validate=true", aut.firmwareConfigResponseValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "/entities"
func TestPostFirmwareConfigEntities(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// TODO "/entities"
func TestPutFirmwareConfigEntities(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "?export"
func TestGetFirmwareConfigWithParamExport(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FC_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1&validate=true", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, "fetched=" + aut.eval("begin_count +1") + "&validate_export=true", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, "fetched=" + aut.eval("begin_count") + "&validate_export=true", aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "?exportAll"
func TestGetFirmwareConfigWithParamExportAll(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?exportAll", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FC_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1&validate=true", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?exportAll", http.StatusOK, "fetched=" + aut.eval("begin_count +1") + "&validate_export=true", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?exportAll", http.StatusOK, "fetched=" + aut.eval("begin_count") + "&validate_export=true", aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "/{id}"
func TestGetFirmwareConfigById(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/", http.StatusNotFound, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/firmware_config_unit_test_not_exist", http.StatusNotFound, NO_POSTERMS, nil},

		{FC_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/firmware_config_unit_test_1", http.StatusOK, "ID=firmware_config_unit_test_1", aut.firmwareConfigSingleValidator},

		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/firmware_config_unit_test_1", http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "/{id}?export"
func TestGetFirmwareConfigByIdWithParam(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FC_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1&validate=true", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)
	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + aut.getValOf("id_1") + "?export", http.StatusOK, "fetched=1", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + aut.getValOf("id_1") + "?export", http.StatusNotFound, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "/{id}"
func TestDeleteFirmwareConfigById(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/", http.StatusNotFound, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/firmware_config_unit_test_not_exist", http.StatusNotFound, NO_POSTERMS, nil},

		{FC_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/firmware_config_unit_test_1", http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "/filtered"
func TestPostFirmwareConfigFilteredWithParams(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)
	sysGenId := uuid.New().String()
	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	log.Print("BEGIN_COUNT=" + aut.eval("begin_count"))
	testCases = []apiUnitTestCase{
		// invalid query params are ignored
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?name=dummy", http.StatusOK, NO_POSTERMS, nil},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNum=1", http.StatusOK, NO_POSTERMS, nil},

		{FC_API, "firmware_config_one", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareConfigResponseValidator},
		{FC_API, "firmware_config_two", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_2", aut.firmwareConfigResponseValidator},
		{FC_API, "firmware_config_three", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_3", aut.firmwareConfigResponseValidator},
		{FC_API, "firmware_config_four", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_4", aut.firmwareConfigResponseValidator},

		// Happy Paths
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=4", http.StatusOK, "fetched=4", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=2", http.StatusOK, "fetched=2", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=2&pageSize=3", http.StatusOK, "fetched=1", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=0&pageSize=3", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=-1&pageSize=3", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=4", http.StatusOK, "fetched=4", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=5", http.StatusOK, "fetched=4", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=0", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=-1", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered", http.StatusOK, "fetched=4", aut.firmwareConfigArrayValidator},

		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=A&pageSize=B", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=A&pageSize=1", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=B", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},

		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=&pageSize=", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber= &pageSize= ", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber= &pageSize=", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=&pageSize= ", http.StatusBadRequest, "fetched=0", aut.firmwareConfigArrayValidator},

		// Happy Paths: default value for missing query params
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1", http.StatusOK, "fetched=4", aut.firmwareConfigArrayValidator},
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageSize=3", http.StatusOK, "fetched=3", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count + 4"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_2"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_3"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_4"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "/model/{modelId}"
func TestGetFirmwareConfigModelByModelId(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/", http.StatusNotFound, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/non_existant", http.StatusNotFound, "fetched=0", aut.firmwareConfigArrayValidator},

		{FC_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/DPC9999T", http.StatusOK, "fetched=1", aut.firmwareConfigArrayValidator},
		{FC_API, "missing_id", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/DPC9999T", http.StatusOK, "fetched=1", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/", http.StatusNotFound, "fetched=0", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/DPC", http.StatusNotFound, "fetched=0", aut.firmwareConfigArrayValidator},

		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/firmware_config_unit_test_1", http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// TODO "/bySupportedModels"
func TestPostFirmwareConfigBySupportedModels(t *testing.T) {

	aut := newFirmwareConfigApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// supportedConfigsByEnvModelRuleName/{ruleName}
func IgnoreTestGetFirmwareConfigSupportedConfigsByEnvModelRuleNameByRuleName(t *testing.T) {

	aut := newFirmwareRuleApiUnitTest(t)
	sysGenFRId := uuid.New().String()
	sysGenFCId := uuid.New().String()
	sysGenModelId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{MODEL_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=MODEL_begin_count", aut.modelArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=FC_begin_count", aut.firmwareConfigArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=FR_begin_count", aut.firmwareRuleArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=FRT_begin_count", aut.firmwareRuleTemplateArrayValidator},

		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/supportedConfigsByEnvModelRuleName/aawrule2", http.StatusOK, "saveFetchedCntIn=FC_API_begin_count", aut.firmwareConfigArrayValidator},

		{MODEL_UAPI, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenModelId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_one&validate=true", aut.modelResponseValidator},
		{FRT_API, "create_env_model", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=templ_id_1", aut.firmwareRuleTemplateResponseValidator},
		{FC_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenFCId + "&SYSTEM_GENERATED_UNIQUE_MODEL_ID=" + sysGenModelId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FR_API, "create_with_sys_gen_id_for_config", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenFRId + "&SYSTEM_GENERATED_UNIQUE_CONFIG_ID=" + sysGenFCId + "&SYSTEM_GENERATED_UNIQUE_MODEL_ID=" + sysGenModelId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/supportedConfigsByEnvModelRuleName/aawrule2", http.StatusOK, "fetched=" + aut.eval("FC_API_begin_count + 1"), aut.firmwareConfigArrayValidator},

		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + sysGenFRId, http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + sysGenFCId, http.StatusNoContent, NO_POSTERMS, nil},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("templ_id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("model_id_one"), http.StatusNoContent, NO_POSTERMS, nil},

		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/supportedConfigsByEnvModelRuleName/aawrule2", http.StatusOK, "fetched=" + aut.eval("FC_API_begin_count"), aut.firmwareConfigArrayValidator},

		{MODEL_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("MODEL_begin_count"), aut.modelArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("FC_begin_count"), aut.firmwareConfigArrayValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("FR_begin_count"), aut.firmwareRuleArrayValidator},
		{FRT_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("FRT_begin_count"), aut.firmwareRuleTemplateArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "/getSortedFirmwareVersionsIfExistOrNot"

func TestPostFirmwareConfigGetSortedFirmwareVersionsIfExistOrNot(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)
	sysGenConfigId := uuid.New().String()
	sysGenModelId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{MODEL_UAPI, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenModelId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_one&validate=true", aut.modelResponseValidator},
		{FC_API, "firmware_config_data", NO_PRETERMS, nil, "POST", "/getSortedFirmwareVersionsIfExistOrNot", http.StatusOK, "saveExisted=begin_existed&saveNotExisted=begin_not_exist", aut.firmwareVersionMapValidator},
		{FC_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenConfigId + "&SYSTEM_GENERATED_UNIQUE_MODEL_ID=" + sysGenModelId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=config_id_1", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, "firmware_config_data", NO_PRETERMS, nil, "POST", "/getSortedFirmwareVersionsIfExistOrNot", http.StatusOK, "fetchExisted=" + aut.eval("begin_existed+1"), aut.firmwareVersionMapValidator},
		{FC_API, "firmware_config_data", NO_PRETERMS, nil, "POST", "/getSortedFirmwareVersionsIfExistOrNot", http.StatusOK, "fetchNotExisted=" + aut.eval("begin_not_exist-1"), aut.firmwareVersionMapValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("config_id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("model_id_one"), http.StatusNoContent, NO_POSTERMS, nil},

		{FC_API, "firmware_config_data", NO_PRETERMS, nil, "POST", "/getSortedFirmwareVersionsIfExistOrNot", http.StatusOK, "fetchExisted=" + aut.eval("begin_existed"), aut.firmwareVersionMapValidator},
		{FC_API, "firmware_config_data", NO_PRETERMS, nil, "POST", "/getSortedFirmwareVersionsIfExistOrNot", http.StatusOK, "fetchNotExisted=" + aut.eval("begin_not_exist"), aut.firmwareVersionMapValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

func TestGetFirmwareConfigBySupportedModels(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)
	sysGenConfigId := uuid.New().String()
	sysGenModelId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FC_API, "model_ids", NO_PRETERMS, nil, "POST", "/bySupportedModels", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{MODEL_UAPI, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenModelId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_one&validate=true", aut.modelResponseValidator},
		{FC_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenConfigId + "&SYSTEM_GENERATED_UNIQUE_MODEL_ID=" + sysGenModelId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=config_id_1", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, "model_ids", NO_PRETERMS, nil, "POST", "/bySupportedModels", http.StatusOK, "fetched=" + aut.eval("begin_count +1"), aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("config_id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, "model_ids", NO_PRETERMS, nil, "POST", "/bySupportedModels", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("model_id_one"), http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "/firmwareConfigMap"
func TestGetFirmwareConfigFirmwareConfigMap(t *testing.T) {
	aut := newFirmwareConfigApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/firmwareConfigMap", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigMapValidator},
		{FC_API, "firmware_config_one", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareConfigResponseValidator},
		{FC_API, "firmware_config_two", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_2", aut.firmwareConfigResponseValidator},
		{FC_API, "firmware_config_three", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_3", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/firmwareConfigMap", http.StatusOK, "fetched=" + aut.eval("begin_count+3"), aut.firmwareConfigMapValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_2"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_3"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_val"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

// "/byEnvModelRuleName/dummy_ruleName"

func EnvModelRuleCreationNotReadyYetTestGetFirmwareConfigByEnvModelRuleNameByRuleName(t *testing.T) {

	newFirmwareRuleApiUnitTest(t)
	newFirmwareRuleTemplateApiUnitTest(t)
	aut := newFirmwareConfigApiUnitTest(t)
	sysGenId := uuid.New().String()
	sysGenConfigId := uuid.New().String()
	sysGenModelId := uuid.New().String()

	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/byEnvModelRuleName/aawrule2", http.StatusOK, "ID=", aut.firmwareConfigSingleValidator},
		{FRT_API, "create_env_model", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{MODEL_UAPI, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenModelId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_one&validate=true", aut.modelResponseValidator},
		{FC_API, "create_with_sys_gen_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenConfigId + "&SYSTEM_GENERATED_UNIQUE_MODEL_ID=" + sysGenModelId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=config_id_1", aut.firmwareConfigResponseValidator},
		{FR_API, "create_with_sys_gen_id_for_config", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId + "&SYSTEM_GENERATED_UNIQUE_CONFIG_ID=" + sysGenConfigId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/byEnvModelRuleName/aawrule2", http.StatusOK, "ID=" + sysGenConfigId, aut.firmwareConfigSingleValidator},
		{FR_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + sysGenId, http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + sysGenConfigId, http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + sysGenModelId, http.StatusNoContent, NO_POSTERMS, nil},
		//{FRT_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf ("templ_id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/byEnvModelRuleName/aawrule2", http.StatusOK, "ID=", aut.firmwareConfigSingleValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

func TestFirmwareConfigCRUD(t *testing.T) {

	aut := newFirmwareConfigApiUnitTest(t)
	sysGenId := uuid.New().String()
	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/crud_393e2152-9d50-4f30-aab9-c74977471632", http.StatusNotFound, NO_POSTERMS, nil},
		{FC_API, "firmware_config_crud", NO_PRETERMS, nil, "PUT", "", http.StatusNotFound, NO_POSTERMS, nil},
		{FC_API, "firmware_config_crud", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/DPC8888", http.StatusConflict, NO_POSTERMS, nil},
		{FC_API, "firmware_config_crud_dup", NO_PRETERMS, nil, "POST", "", http.StatusConflict, NO_POSTERMS, nil},
		{FC_API, "firmware_config_crud", NO_PRETERMS, nil, "PUT", "", http.StatusOK, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/crud_393e2152-9d50-4f30-aab9-c74977471632", http.StatusOK, "ID=crud_393e2152-9d50-4f30-aab9-c74977471632", aut.firmwareConfigSingleValidator},
		{FC_API, "firmware_config_crud", NO_PRETERMS, nil, "POST", "", http.StatusConflict, NO_POSTERMS, nil},
		{FC_API, "firmware_config_crud", NO_PRETERMS, nil, "PUT", "", http.StatusOK, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/crud_393e2152-9d50-4f30-aab9-c74977471632", http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/crud_393e2152-9d50-4f30-aab9-c74977471632", http.StatusNotFound, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/crud_393e2152-9d50-4f30-aab9-c74977471632", http.StatusNotFound, NO_POSTERMS, nil},
		{FC_API, "create_invalid_fw_download", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FC_API, "create_missing_fw_download", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/firmware_config_unit_test_1", http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, "create_missing_id", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=id_fc", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)
	testCases = []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_fc"), http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}

func TestFirmwareConfigEndPoints(t *testing.T) {

	aut := newFirmwareConfigApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		//	"" PostFirmwareConfigHandler "POST"
		{FC_API, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=fc_id_1", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)
	idCreated := aut.getValOf("fc_id_1")

	testCases = []apiUnitTestCase{
		//	"" PutFirmwareConfigHandler "PUT"
		{FC_API, "create", NO_PRETERMS, nil, "PUT", "", http.StatusOK, NO_POSTERMS, nil},

		// 	"" GetFirmwareConfigHandler "GET"
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, NO_POSTERMS, nil},

		// "/entities" PostFirmwareConfigEntitiesHandler "POST"
		{FC_API, "[create]", NO_PRETERMS, nil, "POST", "/entities", http.StatusOK, NO_POSTERMS, nil},

		//	"/entities" PutFirmwareConfigEntitiesHandler "PUT"
		{FC_API, "[create]", NO_PRETERMS, nil, "PUT", "/entities", http.StatusOK, NO_POSTERMS, nil},

		//	"/page" GetFirmwareConfigPageHandler "GET"
		// {FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/page?pageNumber=1&pageSize=10", http.StatusOK, NO_POSTERMS, nil},

		//	"" GetFirmwareConfigWithParamHandler "GET"
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?exportAll", http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" GetFirmwareConfigByIdHandler "GET"
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + idCreated, http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" GetFirmwareConfigByIdWithParamHandler "GET"
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + idCreated + "?export", http.StatusOK, NO_POSTERMS, nil},

		// No registered handler
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + idCreated + "?unknown", http.StatusOK, NO_POSTERMS, nil},

		//	"/filtered" PostFirmwareConfigFilteredWithParamsHandler "POST"
		{FC_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=10", http.StatusOK, NO_POSTERMS, nil},

		//	"/model/{modelId}" GetFirmwareConfigModelByModelIdHandler "GET"
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/dummy_model", http.StatusNotFound, NO_POSTERMS, nil},

		//	"/bySupportedModels" PostFirmwareConfigBySupportedModelsHandler "POST"
		{FC_API, "model_ids", NO_PRETERMS, nil, "POST", "/bySupportedModels", http.StatusOK, NO_POSTERMS, nil},

		//	"/supportedConfigsByEnvModelRuleName/{ruleName}" GetFirmwareConfigSupportedConfigsByEnvModelRuleNameByRuleNameHandler "GET"
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/supportedConfigsByEnvModelRuleName/dummy_ruleName", http.StatusNotFound, NO_POSTERMS, nil},

		//	"/getSortedFirmwareVersionsIfExistOrNot" PostFirmwareConfigGetSortedFirmwareVersionsIfExistOrNotHandler "POST"
		{FC_API, "firmware_config_data", NO_PRETERMS, nil, "POST", "/getSortedFirmwareVersionsIfExistOrNot", http.StatusOK, NO_POSTERMS, nil},

		//	"/byEnvModelRuleName/{ruleName}" GetFirmwareConfigByEnvModelRuleNameByRuleNameHandler "GET"
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/byEnvModelRuleName/dummy_ruleName", http.StatusOK, NO_POSTERMS, nil},

		//	"/firmwareConfigMap" GetFirmwareConfigFirmwareConfigMapHandler "GET"
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/firmwareConfigMap", http.StatusOK, NO_POSTERMS, nil},

		// No registered handler
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?unknown", http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" DeleteFirmwareConfigByIdHandler "DELETE"
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + idCreated, http.StatusNoContent, NO_POSTERMS, nil},
		{FC_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
	aut.baseEntityCount(t, "end:")

}
