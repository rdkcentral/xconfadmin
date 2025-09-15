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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rdkcentral/xconfadmin/adminapi"

	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/stretchr/testify/assert"
)

func TestCalculateHashAndPercent(t *testing.T) {
	_, router := GetTestWebConfigServer(testconfig)
	adminapi.XconfSetup(server, router)
	testCases := []struct {
		queryParams     [][]string
		expectedCode    int
		expectedHash    string
		expectedPercent string
	}{
		{
			queryParams: [][]string{
				{"applicationType", "stb"},
				{"esbMac", "00:23:ED:22:E3:BD"},
			},
			expectedCode:    http.StatusOK,
			expectedHash:    "12320340683479030000",
			expectedPercent: "66.78870067394755",
		},
		{
			queryParams: [][]string{
				{"applicationType", "stb"},
				{"esbMac", "AA:BB:CC:DD:EE:FF"},
			},
			expectedCode:    http.StatusOK,
			expectedHash:    "12349223593569946000",
			expectedPercent: "66.94527524328892",
		},
	}
	for _, testCase := range testCases {
		queryString, _ := util.GetURLQueryParameterString(testCase.queryParams)
		url := fmt.Sprintf("/xconfAdminService/percentfilter/calculator?%v", queryString)
		r := httptest.NewRequest("GET", url, nil)
		rr := ExecuteRequest(r, router)
		responseBody := rr.Body.String()
		assert.Equal(t, testCase.expectedCode, rr.Code)
		assert.Equal(t, strings.Contains(string(responseBody), testCase.expectedHash), true)
		assert.Equal(t, strings.Contains(string(responseBody), testCase.expectedPercent), true)

		//passing invalid estb mac to check the validation
		queryParams, _ := util.GetURLQueryParameterString([][]string{
			{"applicationType", "stb"},
			{"esbMac", "00:23:ED:22:E3:D"},
		})
		url = fmt.Sprintf("/xconfAdminService/percentfilter/calculator?%v", queryParams)
		r = httptest.NewRequest("GET", url, nil)
		rr = ExecuteRequest(r, router)
		responseBody = rr.Body.String()
		assert.Equal(t, 400, rr.Code)
		assert.Equal(t, strings.Contains(string(responseBody), "Invalid Estb Mac"), true)

	}
}
