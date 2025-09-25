/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
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

package http

import (
	//"net/http"
	"net/http"
	"net/http/httptest"
)

var (
	mockEmptyResponse = []byte(`{}`)
)

// setupMocks sets up mock servers that return the same predefined response for any call to the server
// mock servers are set up for all external services - device, tagging, xconf,
// If a different mock response is desired for a test, use the same template below, but just define a different mockResponse
// An example for a different mock response can be seen in http/supplementary_handler_test.go
func (server *WebconfigServer) setupMocks() {
	server.mockSat()
	server.mockDevice()
	server.mockTagging()
	server.mockAccount()
	server.mockCanaryMgr()
}

func (server *WebconfigServer) mockSat() {
	mockResponse := []byte(`{"access_token":"one_mock_token","expires_in":86400,"scope":"scope1 scope2 scope3","token_type":"Bearer"}`)

	// Sat mock server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(mockResponse)
		}))
	server.XW_XconfServer.SatServiceConnector.SetSatServiceHost(mockServer.URL)

}

func (server *WebconfigServer) mockCanaryMgr() {
	mockScraperStatusResponse := []byte(`{"Status": "COMPLETED"}`)
	mockCreateCanaryGroupResponse := []byte(`{"name": "testCanaryGroupName","estbMacs": ["AA:AA:AA:AA:AA:AA","BB:BB:BB:BB:BB:BB", "CC:CC:CC:CC:CC:CC"]}`)
	mockHealthzResponse := []byte(`{"status": 200,"message": "OK"}`)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/canarygroup", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(mockCreateCanaryGroupResponse)
	})
	mux.HandleFunc("/api/v1/penetrationdata/scraper/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(mockScraperStatusResponse)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(mockHealthzResponse)
	})

	// canarymgr mock server
	mockServer := httptest.NewServer(mux)
	server.SetCanaryMgrHost(mockServer.URL)
}

func (server *WebconfigServer) mockDevice() {
	mockResponse := []byte(`{"status":200,"data":{"account_id":"testAccountId", "cpe_mac":"testCpeMac"}}`)

	// odp mock server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(mockResponse)
		}))
	server.XW_XconfServer.DeviceServiceConnector.SetDeviceServiceHost(mockServer.URL)
}

func (server *WebconfigServer) mockTagging() {
	mockResponse := []byte(`["value1", "value2", "value3"]`)
	// tagging mock server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(mockResponse)
		}))
	server.XW_XconfServer.TaggingConnector.SetTaggingHost(mockServer.URL)
}

func (server *WebconfigServer) mockAccount() {
	// account mock server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockEmptyResponse))
		}))
	server.XW_XconfServer.AccountServiceConnector.SetAccountServiceHost(mockServer.URL)
}

func (server *WebconfigServer) mockXconf() {
	// xconf mock server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockEmptyResponse))
		}))
	server.XconfConnector.SetXconfHost(mockServer.URL)
	//TODO
}
