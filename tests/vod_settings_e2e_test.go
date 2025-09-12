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

func ImportVodSettingsTableData(data []string, tabletype logupload.VodSettings) error {
	var err error
	for _, row := range data {
		err = json.Unmarshal([]byte(row), &tabletype)
		err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_VOD_SETTINGS, tabletype.ID, &tabletype)
	}
	return err
}

func TestAllVodSettingsApis(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	//GET ALL VOD SETTINGS
	var tableData = []string{
		`{"id":"07f05421-8e6e-4f93-8918-46fc247a61d3","updated":1572462347409,"ttlMap":{},"name":"wsmithDCM6VOD","locationsURL":"http://www.dcmTest.com","ipNames":[],"ipList":[],"srmIPList":{},"applicationType":"stb"}`,
		`{"id":"07f05421-8e6e-4f93-8918-46fc247a61d3id","updated":1572462347409,"ttlMap":{},"name":"wsmithDCM6VOD","locationsURL":"http://www.dcmTest.com","ipNames":[],"ipList":[],"srmIPList":{},"applicationType":"stb"}`,
		`{"id":"07f05421-8e6e-4f93-8918-46fc247a61d3sz","updated":1572462347409,"ttlMap":{},"name":"wsmithDCM6VOD","locationsURL":"http://www.dcmTest.com","ipNames":[],"ipList":[],"srmIPList":{},"applicationType":"stb"}`,
		`{"id":"07f05421-8e6e-4f93-8918-46fc247a61d3nsz","updated":1572462347409,"ttlMap":{},"name":"wsmithDCM3VOD","locationsURL":"http://www.dcmTest.com","ipNames":[],"ipList":[],"srmIPList":{},"applicationType":"stb"}`,
		`{"id":"07f05421-8e6e-4f93-8918-46fc247a61d3fz","updated":1572462347409,"ttlMap":{},"name":"dineshfiltVOD","locationsURL":"http://www.dcmTest.com","ipNames":[],"ipList":[],"srmIPList":{},"applicationType":"stb"}`,
		`{"id":"07f05421-8e6e-4f93-8918-46fc247a61d3dl","updated":1572462347409,"ttlMap":{},"name":"wsmithDCM6VOD","locationsURL":"http://www.dcmTest.com","ipNames":[],"ipList":[],"srmIPList":{},"applicationType":"stb"}`,
	}
	ImportVodSettingsTableData(tableData, logupload.VodSettings{})

	urlall := "/xconfAdminService/dcm/vodsettings"
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
		var dss = []logupload.VodSettings{}
		json.Unmarshal(body, &dss)
		assert.Equal(t, len(dss) > 0, true)
	}

	//CREATE VOD SETTING
	vsdata := []byte(
		`{"id":"33af3261-d74a-40fd-8aa1-884e4f5479a1","updated":1635290206352,"name":"testvod","locationsURL":"http://test.com","ipNames":["ip1","ip2"],"ipList":["1.1.1.1","2.2.2.2"], "applicationType":"stb"}`)

	urlCr := "/xconfAdminService/dcm/vodsettings?applicationType=stb"
	req, err = http.NewRequest("POST", urlCr, bytes.NewBuffer(vsdata))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	//ERROR CREATING AGAIN SAME ENTRY
	urlCr = "/xconfAdminService/dcm/vodsettings?applicationType=stb"
	req, err = http.NewRequest("POST", urlCr, bytes.NewBuffer(vsdata))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	//UPDATE EXISING ENTRY
	vsdataup := []byte(
		`{"id":"33af3261-d74a-40fd-8aa1-884e4f5479a1","updated":1635290206352,"name":"testdata","locationsURL":"http://test.com","ipNames":["ip1","ip2"],"ipList":["14.14.14.1","2.2.2.2"],"applicationType":"stb"}`)

	urlup := "/xconfAdminService/dcm/vodsettings?applicationType=stb"
	req, err = http.NewRequest("PUT", urlup, bytes.NewBuffer(vsdataup))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//UPDATE NON EXISTING ENTRY
	vsdataerr := []byte(
		`{"id":"33af3261-d74a-40fd-8aa1-884e4f5479a1err","updated":1635290206352,"name":"testdata","locationsURL":"http://test.com","ipNames":["ip1","ip2"],"ipList":["14.14.14.1","2.2.2.2"],"applicationType":"stb"}`)

	urlup = "/xconfAdminService/dcm/vodsettings?applicationType=stb"
	req, err = http.NewRequest("PUT", urlup, bytes.NewBuffer(vsdataerr))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	//GET VOD SETTING BY ID
	urlWithId := "/xconfAdminService/dcm/vodsettings/07f05421-8e6e-4f93-8918-46fc247a61d3id?applicationType=stb"
	req, err = http.NewRequest("GET", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	//GET VOD SETTING BY SIZE

	urlWithId = "/xconfAdminService/dcm/vodsettings/size?applicationType=stb"
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

	//GET VOD SETTING BY NAMES
	urlWithId = "/xconfAdminService/dcm/vodsettings/names?applicationType=stb"
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
		var vss = []logupload.VodSettings{}
		json.Unmarshal(body, &vss)
		assert.Equal(t, len(vss) > 0, true)
	}

	//GET VOD RULES BY FILTERED NAMES
	urlWithfilt := "/xconfAdminService/dcm/vodsettings/filtered?pageNumber=1&pageSize=50"
	postmapname1 := []byte(`{"NAME": "testdata"}`)
	req, err = http.NewRequest("POST", urlWithfilt, bytes.NewBuffer(postmapname1))
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
		var vss = []logupload.VodSettings{}
		json.Unmarshal(body, &vss)
		assert.Equal(t, len(vss) > 0, true)
	}

	//DELETE VOD SETTINGS BY ID
	urlWithId = "/xconfAdminService/dcm/vodsettings/07f05421-8e6e-4f93-8918-46fc247a61d3dl?applicationType=stb"
	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)

	//DELETE NON EXISTING VOD SETTINGS BY ID
	urlWithId = "/xconfAdminService/dcm/vodsettings/23069266-45b7-4bf6-a255-e6ee584cd6xxxx?applicationType=stb"

	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

}
