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
	"net/http"
	"testing"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"gotest.tools/assert"
)

const (
	Env_Url                    = "/xconfAdminService/queries"
	Queries_Rules_url          = "/xconfAdminService/queries/rules"
	Queries_Filter_url         = "/xconfAdminService/queries/filters"
	Queries_update_path        = "/xconfAdminService/updates"
	Queries_update_filter_path = "/xconfAdminService/updates/filters"
)

type TableData struct {
	Tablename string
	Tablerow  string
}

func ImportTableData(data []interface{}) error {
	var err error
	for _, row := range data {
		switch row.(TableData).Tablename {
		case "TABLE_ENVIRONMENT":
			var tabletype = shared.Environment{}
			err = json.Unmarshal([]byte(row.(TableData).Tablerow), &tabletype)
			err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_ENVIRONMENT, tabletype.ID, &tabletype)
			break
		case "TABLE_GENERIC_NS_LIST":
			var humptyStrList = []string{
				"Humpty Dumpty sat on a wall",
				"Humpty Dumpty had a great fall",
				"All the king's horses and all the king's men",
				"Couldn't put Humpty together again",
			}

			tabletype := shared.NewGenericNamespacedList(fmt.Sprintf("CDN-TESTING"), "STRING", humptyStrList)
			ipList := []string{
				"127.1.1.1",
				"127.1.1.2",
				"127.1.1.3",
			}

			tabletype.TypeName = "IP_LIST"
			tabletype.Data = ipList
			err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, tabletype.ID, tabletype)
			break
		case "TABLE_FIRMWARE_CONFIG":
			var firmwareConfig = coreef.NewEmptyFirmwareConfig()
			err = json.Unmarshal([]byte(row.(TableData).Tablerow), &firmwareConfig)
			err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_FIRMWARE_CONFIG, firmwareConfig.ID, firmwareConfig)
			break

		case "TABLE_FIRMWARE_RULE":
			var firmwareRule = corefw.NewEmptyFirmwareRule()
			var data_str = row.(TableData).Tablerow
			err = json.Unmarshal([]byte(data_str), &firmwareRule)
			err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_FIRMWARE_RULE, firmwareRule.ID, firmwareRule)
			break

		case "TABLE_SINGLETON_FILTER_VALUE":
			var data_str = row.(TableData).Tablerow
			locationRoundRobinFilter := coreef.NewEmptyDownloadLocationRoundRobinFilterValue()
			err = json.Unmarshal([]byte(data_str), &locationRoundRobinFilter)
			err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_SINGLETON_FILTER_VALUE, locationRoundRobinFilter.ID, locationRoundRobinFilter)
			break
		}

	}

	return err
}

func TestAllQueriesApis(t *testing.T) {
	DeleteAllEntities()

	table_data := []interface{}{
		TableData{Tablename: "TABLE_ENVIRONMENT", Tablerow: `{"id":"AX061AEI","updated":1591604177484,"description":"RT1319"}`},
		TableData{Tablename: "TABLE_GENERIC_NS_LIST", Tablerow: ``},
		TableData{Tablename: "TABLE_FIRMWARE_CONFIG", Tablerow: `{"id":"207dc5a5-d324-4e2e-9daf-5017ed98f8f3","updated":1558520642121,"description":"CPEAUTO_FW_AA:AA:AA:AA:AA:AA","supportedModelIds":["XCONFTESTMODEL"],"firmwareDownloadProtocol":"http","firmwareFilename":"DPC3941_3.3p17s1_DEV_sey-signed.bin","firmwareVersion":"DPC3941_3.3p17s1_DEV_sey-signed","rebootImmediately":false,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"437afab9-cbe3-4e4d-b175-220865e0f720","name":" Cisco Arris XG1","rule":{"negated":false,"compoundParts":[{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"ipAddress"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":""}}}}},{"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"env"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"VBN"}}}}},{"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"MX011ANC"}}}}}]},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configId":"e675358b-506d-48f8-86c5-c8c8e3bb6254","active":true,"firmwareCheckRequired":false,"rebootImmediately":false},"type":"IP_RULE","active":true}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"c4681132-c518-459a-99fb-9b93a1f42f37","name":"CDN-TESTING","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CDN-TESTING"}}}}},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configId":"dff46b03-be65-4f0c-804d-542d5ffec8ec","active":true,"firmwareCheckRequired":false,"rebootImmediately":false},"type":"MAC_RULE","active":true,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"67333656-9e8e-46a3-9a87-2f42644a35c9","name":"Arris_XG1v1_VBN_Moto-DEV","rule":{"negated":false,"compoundParts":[{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"env"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"VBN"}}}}},{"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"MX011ANM"}}}}},{"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"partnerId"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"testDEV"}}}}}]},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configEntries":[{"configId":"5de4a2df-2673-4be3-ae67-4e09648a929b","percentage":100.0,"startPercentRange":0.0,"endPercentRange":100.0}],"active":true,"firmwareCheckRequired":true,"rebootImmediately":true,"firmwareVersions":["MX011AN_3.8p3s1_VBN_sey","MX011AN_3.1p1s3_VBN_sey","MX011AN_3.2p6s1_VBN_sey-signed"]},"type":"ENV_MODEL_RULE","active":true,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"c4681132-c518-459a-99fb-9b93a1f41gf37","name":"Test_Ip_filter_device","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CDN-TESTING"}}}}},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configId":"dff46b03-be65-4f0c-804d-542d5ffec8ec","active":true,"firmwareCheckRequired":false,"rebootImmediately":false},"type":"IP_FILTER","active":true,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"c4681132-c518-459a-99fb-9b93a1f63534","name":"Test_Time_filter_device","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CDN-TESTING"}}}}},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configId":"dff46b03-be65-4f0c-804d-542d5ffec8ec","active":true,"firmwareCheckRequired":false,"rebootImmediately":false},"type":"TIME_FILTER","active":true,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"67f595ae-3e1d-418d-9b86-22b3e46816e4","name":"CPEAUTO_LF_80:f5:03:34:11:fd","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"ipAddress"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CPEAUTOIPGRP80f5033411fd"}}}}},"applicableAction":{"type":".DefinePropertiesAction","ttlMap":{},"actionType":"DEFINE_PROPERTIES","properties":{"firmwareLocation":"http://ssr.ccp.xcal.tv/cgi-bin/x1-sign-redirect.pl?K=10&F=stb_cdl","firmwareDownloadProtocol":"http","ipv6FirmwareLocation":""},"activationFirmwareVersions":{}},"type":"DOWNLOAD_LOCATION_FILTER","active":true,"applicationType":"stb"}`},
		TableData{Tablename: "TABLE_SINGLETON_FILTER_VALUE", Tablerow: `{"type":"com.comcast.xconf.estbfirmware.DownloadLocationRoundRobinFilterValue","id":"DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE","updated":1616699042493,"applicationType":"stb","locations":[{"locationIp":"96.114.220.246","percentage":100.0},{"locationIp":"69.252.106.162","percentage":0.0}],"ipv6locations":[{"locationIp":"2600:1f18:227b:c01:b161:3d17:7a86:fe36","percentage":100.0},{"locationIp":"2001:558:1020:1:250:56ff:fe94:646f","percentage":0.0}],"httpLocation":"test.com","httpFullUrlLocation":"https://test.com/Images"}`},
		TableData{Tablename: "TABLE_FIRMWARE_RULE", Tablerow: `{"id":"e313bc81-8a02-4087-8c91-1da6db4b3159","name":"CDL-ARRISXG1V4-QA","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CDL-ARRISXG1V4-QA"}}}}},"applicableAction":{"type":".DefinePropertiesAction","ttlMap":{},"actionType":"DEFINE_PROPERTIES","properties":{"rebootImmediately":"true"},"byPassFilters":[]},"type":"REBOOT_IMMEDIATELY_FILTER","active":true}`},
	}
	err := ImportTableData(table_data)
	assert.NilError(t, err)
	//GET ENVIRONMENTS
	url := fmt.Sprintf("%s/%s", Env_Url, "environments")
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Get ENVIRONMENTS BY ID
	urlWithId := fmt.Sprintf("%s/%s/%s", Env_Url, "environments", "AX061AEI")
	req, err = http.NewRequest("GET", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET IPADDRESSGROUPS
	url = fmt.Sprintf("%s/%s", Env_Url, "ipAddressGroups")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET IPADDRESSGROUPS BY IP
	url = fmt.Sprintf("%s/%s", Env_Url, "ipAddressGroups/byIp/127.1.1.1")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET NSLISTS
	url = fmt.Sprintf("%s/%s", Env_Url, "nsLists")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET NSLISTS BY ID
	url = fmt.Sprintf("%s/%s", Env_Url, "nsLists/byId/"+"wweii2900292ii39")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET FIRMWARES BY MODEL ID

	url = fmt.Sprintf("%s/%s", Env_Url, "firmwares/model/"+"XCONFTESTMODEL?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

	//GET FIRMWARES BY SUPORTEDMODELS
	var postData = []byte(
		`["XCONFTESTMODEL"]`,
	)
	url = fmt.Sprintf("%s/%s", Env_Url, "firmwares/bySupportedModels?applicationType=stb")
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(postData))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//-----------------------------------------------------
	//QUERIES RULES API'S
	//-------------------------------------------------------

	//GET IPS RULES
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, "ips?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 0, true)
	}

	//GET IPS RULES BY NAME
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, `ips/ Cisco Arris XG1?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 0, true)
	}

	//GET MAC RULES
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, "macs?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 0, true)
	}

	//GET MAC RULES BY RULE NAME
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, `macs/CDN-TESTING?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET ENV MODELS
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, "envModels?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET ENV MODELS WITH NAME
	url = fmt.Sprintf("%s/%s", Queries_Rules_url, "envModels/Arris_XG1v1_VBN_Moto-DEV?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//-----------------------------------------------------
	//QUERIES FILTERS API'S
	//-------------------------------------------------------

	//GET IPS FILTER
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, "ips?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET IPS RULES BY NAME
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, `ips/Test_Ip_filter_device?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET TIME FILTER
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, "time?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET TIME FILTER BY NAME
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, `time/Test_Time_filter_device?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET LOCATION FILTER
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, "locations?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET LOCATION FILTER BY NAME
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, `locations/CPEAUTO_LF_80:f5:03:34:11:fd?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET DOWNLOAD LOCATION
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, "downloadlocation?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET REEBOOT IMMEDIATELY FILTER
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, "ri?applicationType=stb")
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//GET REEBOOT IMMEDIATELY FILTER BY NAME
	url = fmt.Sprintf("%s/%s", Queries_Filter_url, `ri/CDL-ARRISXG1V4-QA?applicationType=stb`)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//-----------------------------------------------------
	//QUERIES UPDATES API'S
	//-----------------------------------------------------
	var body_data = []byte(`{"id":"AX061AE2","updated":1541604177484,"description":"TESTRT1319"}`)
	url = fmt.Sprintf("%s/%s", Queries_update_path, "environments")
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(body_data))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)
	body, err = ioutil.ReadAll(res.Body)

	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

	//-----------------------------------------------------
	//QUERIES UPDATES FILTERS API'S
	//-----------------------------------------------------

	//POST IPS FILTER

	body_data = []byte(`{"id":"c4681132-c518-459a-99fb-9b93a1f41gf37","name":"Test_Ip_filter_device","rule":{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"eStbMac"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"CDN-TESTING"}}}}},"applicableAction":{"type":".RuleAction","ttlMap":{},"actionType":"RULE","configId":"dff46b03-be65-4f0c-804d-542d5ffec8ec","active":true,"firmwareCheckRequired":false,"rebootImmediately":false},"type":"IP_FILTER","IpAddressGroup":{"Id":"CDN-TESTING","Name":"CDN-TESTING","IpAddresses":["127.1.1.1","127.1.1.2","127.1.1.3"],"RawIpAddresses":["127.1.1.1","127.1.1.2","127.1.1.3"]},"active":true,"applicationType":"stb"}`)
	url = fmt.Sprintf("%s/%s", Queries_update_filter_path, "ips?applicationType=stb")
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(body_data))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		assert.Equal(t, len(body) > 5, true)
	}

}
