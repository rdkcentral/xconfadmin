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
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rdkcentral/xconfwebconfig/db"

	"gotest.tools/assert"
)

func TestGetPenetrationMetrics(t *testing.T) {
	truncateTable("PenetrationMetrics")
	err := createPenetrationSampleData()
	assert.NilError(t, err)

	//When EstbMac not present in the PenetrationMetics Table (Response 404)
	url := "/xconfAdminService/penetrationdata/11:22:33:44:65:66"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, strings.Contains(string(body), "11:22:33:44:65:66 not found"), true)
	res.Body.Close()

	url = "/xconfAdminService/penetrationdata/AA:BB:CC:DD:ee"
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, strings.Contains(string(body), "Invalid MAC address"), true)
	res.Body.Close()

	//When Estmac Present in PenetrationTable (Response 200)
	url = "/xconfAdminService/penetrationdata/AA:10:AA:31:AA:35"
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	url = "/xconfAdminService/penetrationdata/aa10aa31aa35"
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
}

func createPenetrationSampleData() error {
	dbClient := db.GetDatabaseClient()
	cassandraClient, ok := dbClient.(*db.CassandraClient)
	if ok {
		penetrationdata := &db.PenetrationMetrics{
			EstbMac:                 "AA:10:AA:31:AA:35",
			Partner:                 "COMCAST",
			Model:                   "TG1682G",
			FwVersion:               "test.12p24s1_PROD_sey",
			FwReportedVersion:       "test.12p24s1_PROD_sey",
			FwAdditionalVersionInfo: "test.12p",
			FwAppliedRule:           "testrule",
			FwTs:                    time.Now(),
			RfcAppliedRules:         "Rule1",
			RfcFeatures:             "Feature1",
			RfcTs:                   time.Now(),
		}
		return cassandraClient.SetPenetrationMetrics(penetrationdata)
	}
	return nil
}
