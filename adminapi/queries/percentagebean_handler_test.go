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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"github.com/stretchr/testify/assert"
)

const (
	PB_URL_BASE = "/xconfAdminService/percentfilter/percentageBean"
	PB_URL      = "/xconfAdminService/percentfilter/percentageBean?applicationType=stb"
)
const testconfig = "../config/sample_xconfadmin.conf"

func PBCreateFirmwareConfig(firmwareVersion string, modelId string, firmwareDownloadProtocol string, applicationType string) *coreef.FirmwareConfig {
	firmwareConfig := coreef.NewEmptyFirmwareConfig()
	firmwareConfig.ID = "PB_creste_test"
	firmwareConfig.Description = "FirmwareDescription"
	firmwareConfig.FirmwareFilename = "FirmwareFilename"
	firmwareConfig.FirmwareVersion = firmwareVersion
	firmwareConfig.FirmwareDownloadProtocol = firmwareDownloadProtocol
	firmwareConfig.ApplicationType = applicationType
	supportedModels := make([]string, 1)
	model := CreateAndSaveModel(strings.ToUpper(modelId))
	supportedModels[0] = model.ID
	return firmwareConfig
}

func TestPBAllApi(t *testing.T) {
	DeleteAllEntities()
	//	_, router := GetTestWebConfigServer(testconfig)
	//adminapi.XconfSetup(server, router)

	parameters := map[string]string{}
	configKey := "bindingUrl"
	configValue := "http://test.url.com"
	parameters[configKey] = configValue

	definePropertiesModelId := "DEFINE_PROPERTIES_MODEL_ID"

	firmwareConfig := PBCreateFirmwareConfig(defaultFirmwareVersion, definePropertiesModelId, "http", "stb")
	firmwareConfig.Properties = parameters
	err := SetFirmwareConfig(firmwareConfig)
	assert.Nil(t, err)

	applicableAction := corefw.NewTemplateApplicableActionAndType(corefw.RuleActionClass, corefw.RULE_TEMPLATE, "")
	CreateAndSaveFirmwareRuleTemplate("ENV_MODEL_RULE", CreateDefaultEnvModelRule(), applicableAction)

	percentageBean := CreatePercentageBeanPB("test percentage bean", defaultEnvironmentId, definePropertiesModelId, "", "", defaultFirmwareVersion, "stb")
	percentageBean.LastKnownGood = firmwareConfig.ID
	percentageBean.FirmwareVersions = append(percentageBean.FirmwareVersions, firmwareConfig.FirmwareVersion)
	err = SavePercentageBeanPB(percentageBean)
	assert.Nil(t, err)

	// get PBrule by id
	id := percentageBean.ID
	urlWithId := fmt.Sprintf("%s/%s?applicationType=stb", PB_URL_BASE, id)
	req, err := http.NewRequest("GET", urlWithId, nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// get PBrule all
	req, err = http.NewRequest("GET", PB_URL, nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	if res.StatusCode == http.StatusOK {
		var pbrules = []*coreef.PercentageBean{}
		json.Unmarshal(body, &pbrules)
		assert.Equal(t, len(pbrules), 1)
	}

	// create PB Eentry through API
	pbdata := []byte(
		`{"id":"0f133a83-030c-45b8-846e-a06e75afff8b","name":"!!!!!0000WarrenTest","active":true,"firmwareCheckRequired":true,"rebootImmediately":true,"firmwareVersions":["firmwareVersion","firmwareVersion"],"distributions":[{"configId":"PB_creste_test","percentage":100,"startPercentRange":0,"endPercentRange":100}],"applicationType":"stb","environment":"ENVIRONMENTID2","model":"DEFINE_PROPERTIES_MODEL_ID","useAccountIdPercentage":false}`)
	req, err = http.NewRequest("POST", PB_URL, bytes.NewBuffer(pbdata))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	// Update  PB Eentry through API
	pbdataup := []byte(
		`{"id":"0f133a83-030c-45b8-846e-a06e75afff8b","name":"DineshUpdatePBEntry","active":true,"firmwareCheckRequired":true,"rebootImmediately":true,"firmwareVersions":["firmwareVersion","firmwareVersion"],"distributions":[{"configId":"PB_creste_test","percentage":100,"startPercentRange":0,"endPercentRange":100}],"applicationType":"stb","environment":"ENVIRONMENTID2", "optionalConditions": {"compoundParts": [ {   "condition": { "freeArg": { "type": "STRING", "name": "SomeKey" },  "operation": "IS", "fixedArg": { "bean": { "value": { "java.lang.String": "SomeValue" } } } }, "negated": false } ], "negated": false },"model":"DEFINE_PROPERTIES_MODEL_ID","useAccountIdPercentage":false}`)
	req, err = http.NewRequest("PUT", PB_URL, bytes.NewBuffer(pbdataup))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Update  PB Eentry through API with error
	pbdataerr := []byte(
		`{"id":"0f133a83-030c-45b8-846e-a06e75afferr","name":"DineshUpdatePBEntry","active":true,"firmwareCheckRequired":true,"rebootImmediately":true,"firmwareVersions":["firmwareVersion","firmwareVersion"],"distributions":[{"configId":"PB_creste_test","percentage":100,"startPercentRange":0,"endPercentRange":100}],"applicationType":"stb","environment":"ENVIRONMENTID2","model":"DEFINE_PROPERTIES_MODEL_ID","useAccountIdPercentage":false}`)
	req, err = http.NewRequest("PUT", PB_URL, bytes.NewBuffer(pbdataerr))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)

	//filtered API
	urlfiltnames := fmt.Sprintf("%s/%s", PB_URL_BASE, "filtered?applicationType=stb&pageNumber=1&pageSize=50")
	postmapname2 := []byte(`{"NAME": "DineshUpdatePBEntry"}`)
	req, err = http.NewRequest("POST", urlfiltnames, bytes.NewBuffer(postmapname2))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	if res.StatusCode == http.StatusOK {
		var pbrules = []*coreef.PercentageBean{}
		json.Unmarshal(body, &pbrules)
		assert.Equal(t, len(pbrules), 1)
	}

	//filtered API
	//urlfiltnames := fmt.Sprintf("%s/%s", PB_URL, "filtered?pageNumber=1&pageSize=50")
	var postmapPBargs = []byte(`{"FIXED_ARG": "SomeValue","FREE_ARG": "SomeKey"}`)

	req, err = http.NewRequest("POST", urlfiltnames, bytes.NewBuffer(postmapPBargs))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	if res.StatusCode == http.StatusOK {
		var pbrules = []*coreef.PercentageBean{}
		json.Unmarshal(body, &pbrules)
		assert.Equal(t, len(pbrules) > 0, true)
	}

	// delete PBrule by id
	deleteurl := PB_URL_BASE + "/0f133a83-030c-45b8-846e-a06e75afff8b?applicationType=stb"
	req, err = http.NewRequest("DELETE", deleteurl, nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)

	// POST entities  PB Eentry through API

	urlWithIdent := fmt.Sprintf("%s/%s?applicationType=stb", PB_URL_BASE, "entities")
	pbdataentpost := []byte(
		`[{"id":"0f133a83-030c-45b8-846e-a06e75afff8b","name":"!!!!!0000WarrenTest","active":true,"firmwareCheckRequired":true,"rebootImmediately":true,"firmwareVersions":["firmwareVersion","firmwareVersion"],"distributions":[{"configId":"PB_creste_test","percentage":100,"startPercentRange":0,"endPercentRange":100}],"applicationType":"stb","environment":"ENVIRONMENTID2","model":"DEFINE_PROPERTIES_MODEL_ID","useAccountIdPercentage":false}]`)
	req, err = http.NewRequest("POST", urlWithIdent, bytes.NewBuffer(pbdataentpost))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	if res.StatusCode == http.StatusOK {
		bodyMap := map[string]string{}
		json.Unmarshal(body, &bodyMap)
		assert.Equal(t, len(bodyMap) > 0, true)
	}

	// PUT entities  PB Eentry through API

	pbdataentput := []byte(
		`[{"id":"0f133a83-030c-45b8-846e-a06e75afff8b","name":"!!!!!0000DINESHPUTENT","active":true,"firmwareCheckRequired":true,"rebootImmediately":true,"firmwareVersions":["firmwareVersion","firmwareVersion"],"distributions":[{"configId":"PB_creste_test","percentage":100,"startPercentRange":0,"endPercentRange":100}],"applicationType":"stb","environment":"ENVIRONMENTID2","model":"DEFINE_PROPERTIES_MODEL_ID","useAccountIdPercentage":false}]`)
	req, err = http.NewRequest("PUT", urlWithIdent, bytes.NewBuffer(pbdataentput))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	if res.StatusCode == http.StatusOK {
		bodyMap := map[string]string{}
		json.Unmarshal(body, &bodyMap)
		assert.Equal(t, len(bodyMap) > 0, true)
	}

	// delete non existing PBrule by id
	deleteurlerr := PB_URL_BASE + "/0f133a83-030c-45b8-846e-a06e75afferr?applicationType=stb"
	req, err = http.NewRequest("DELETE", deleteurlerr, nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

	DeleteAllEntities()
}

func TestPercentageBeanAdminUpdateAPI(t *testing.T) {
	DeleteAllEntities()

	percentageBean, err := PreCreatePercentageBean()
	assert.Nil(t, err)

	url := fmt.Sprintf("/xconfAdminService/percentfilter/percentageBean?applicationType=stb")

	percentageBeanBytes, _ := json.Marshal(percentageBean)

	r := httptest.NewRequest("PUT", url, bytes.NewBuffer(percentageBeanBytes))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	percentageBeanResp := unmarshalPercentageBean(rr.Body.Bytes())

	assert.Equal(t, percentageBean, &percentageBeanResp)
	assertPercentageBeanVersionUUIDs(t, percentageBean, &percentageBeanResp)
	assertDistributionUUIDs(t, &percentageBeanResp)
}

func TestPercentageBeanUpdatesAPI(t *testing.T) {
	DeleteAllEntities()
	percentageBean, err := PreCreatePercentageBean()
	assert.Nil(t, err)

	url := fmt.Sprintf("/xconfAdminService/updates/percentageBean?applicationType=stb")

	percentageBeanBytes, _ := json.Marshal(percentageBean)

	r := httptest.NewRequest("PUT", url, bytes.NewBuffer(percentageBeanBytes))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	percentageBeanResp := unmarshalPercentageBean(rr.Body.Bytes())

	assert.Equal(t, percentageBean, &percentageBeanResp)
	assertPercentageBeanVersionUUIDs(t, percentageBean, &percentageBeanResp)
	assertDistributionUUIDs(t, &percentageBeanResp)
}

func TestPercentageBeanExportAllAPI(t *testing.T) {
	DeleteAllEntities()
	percentageBean, err := PreCreatePercentageBean()
	assert.Nil(t, err)

	url := fmt.Sprintf("/xconfAdminService/percentfilter?export&applicationType=stb")

	r := httptest.NewRequest("GET", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	percentFilterExport := unmarshalPercentFilterExport(rr.Body.Bytes())
	percentageBeans := convertPercentageBeans(percentFilterExport["percentageBeans"].([]interface{}))

	assertPercentageBeanVersionUUIDs(t, percentageBean, &percentageBeans[0])

	assertDistributionUUIDs(t, &percentageBeans[0])
}

func TestSearchPercentageBeanByMinCheckVersion(t *testing.T) {
	DeleteAllEntities()
	percentageBean1, err := PreCreatePercentageBean()
	assert.Nil(t, err)

	firmwareVersion2 := "TEST_FIRMWARE_VERSION"
	percentageBean2 := CreatePercentageBeanPB("NEW PERCENTAGE BEAN", "environment2", "model2", "", "", firmwareVersion2, "stb")
	err = SavePercentageBeanPB(percentageBean2)
	assert.Nil(t, err)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
		{"MIN_CHECK_VERSION", "firmwareVersion"},
	})
	url := fmt.Sprintf("/xconfAdminService/percentfilter/percentageBean/filtered?%v", queryParams)

	r := httptest.NewRequest("POST", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	percentageBeans := unmarshalPercentageBeans(rr.Body.Bytes())

	assert.Contains(t, percentageBeans, percentageBean1)

	queryParams, _ = util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
		{"MIN_CHECK_VERSION", "nonExistingVersion"},
	})
	url = fmt.Sprintf("/xconfAdminService/percentfilter/percentageBean/filtered?%v", queryParams)

	r = httptest.NewRequest("POST", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	percentageBeans = unmarshalPercentageBeans(rr.Body.Bytes())

	assert.Empty(t, percentageBeans)

	queryParams, _ = util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
		{"MIN_CHECK_VERSION", "TEST_FIRMWARE_VERSION"},
	})
	url = fmt.Sprintf("/xconfAdminService/percentfilter/percentageBean/filtered?%v", queryParams)

	r = httptest.NewRequest("POST", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	percentageBeans = unmarshalPercentageBeans(rr.Body.Bytes())
	assert.Contains(t, percentageBeans, percentageBean2)

	queryParams, _ = util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
		{"MIN_CHECK_VERSION", "test_firmware_version"},
	})
	url = fmt.Sprintf("/xconfAdminService/percentfilter/percentageBean/filtered?%v", queryParams)

	r = httptest.NewRequest("POST", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	percentageBeans = unmarshalPercentageBeans(rr.Body.Bytes())
	assert.Contains(t, percentageBeans, percentageBean2)

	queryParams, _ = util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
		{"MIN_CHECK_VERSION", "test_firmware_"},
	})
	url = fmt.Sprintf("/xconfAdminService/percentfilter/percentageBean/filtered?%v", queryParams)

	r = httptest.NewRequest("POST", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	percentageBeans = unmarshalPercentageBeans(rr.Body.Bytes())
	assert.Contains(t, percentageBeans, percentageBean2)

	queryParams, _ = util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
		{"MIN_CHECK_VERSION", "version"},
	})
	url = fmt.Sprintf("/xconfAdminService/percentfilter/percentageBean/filtered?%v", queryParams)

	r = httptest.NewRequest("POST", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	percentageBeans = unmarshalPercentageBeans(rr.Body.Bytes())
	assert.Equal(t, 2, len(percentageBeans))
	assert.Contains(t, percentageBeans, percentageBean1)
	assert.Contains(t, percentageBeans, percentageBean2)
}

func assertPercentageBeanVersionUUIDs(t *testing.T, expectedPB *coreef.PercentageBean, actualPB *coreef.PercentageBean) {
	lkgId, err := uuid.Parse(actualPB.LastKnownGood)
	assert.Nil(t, err)
	assert.Equal(t, expectedPB.LastKnownGood, lkgId.String())

	ivId, err := uuid.Parse(actualPB.IntermediateVersion)
	assert.Nil(t, err)
	assert.Equal(t, expectedPB.IntermediateVersion, ivId.String())
}

func assertDistributionUUIDs(t *testing.T, pb *coreef.PercentageBean) {
	if pb.Distributions != nil && len(pb.Distributions) > 0 {
		for _, distribution := range pb.Distributions {
			distributionId, err := uuid.Parse(distribution.ConfigId)
			assert.Nil(t, err)
			assert.Equal(t, distribution.ConfigId, distributionId.String())
		}
	}
}

func unmarshalPercentageBean(b []byte) coreef.PercentageBean {
	var percentageBean coreef.PercentageBean
	err := json.Unmarshal(b, &percentageBean)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling percentage bean"))
	}
	return percentageBean
}

func unmarshalPercentageBeans(b []byte) []*coreef.PercentageBean {
	var percentageBeans = make([]*coreef.PercentageBean, 0)
	err := json.Unmarshal(b, &percentageBeans)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling percentage bean"))
	}
	return percentageBeans
}

func convertPercentageBeans(pbis []interface{}) []coreef.PercentageBean {
	var percentageBeans []coreef.PercentageBean
	for _, pbi := range pbis {
		var percentageBean coreef.PercentageBean
		mapedPb := pbi.(map[string]interface{})
		b, _ := json.Marshal(mapedPb)

		percentageBean = unmarshalPercentageBean(b)
		percentageBeans = append(percentageBeans, percentageBean)
	}

	return percentageBeans
}

func unmarshalPercentFilterExport(b []byte) map[string]interface{} {
	var percentFilter map[string]interface{}
	err := json.Unmarshal(b, &percentFilter)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling percent filter"))
	}
	return percentFilter
}
