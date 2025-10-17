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
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/rdkcentral/xconfwebconfig/dataapi"

	"github.com/rdkcentral/xconfadmin/adminapi"

	oshttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/taggingapi"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
)

var (
	testConfigFile     string
	jsonTestConfigFile string
	sc                 *common.ServerConfig
	server             *oshttp.WebconfigServer
	router             *mux.Router
	globAut            *apiUnitTest
)
var (
	//used /app/xconfadmin... config
	testConfig = "/app/xconfadmin/xconfadmin.conf"
)

/*
Code is:
Copyright (c) 2023 The Gorilla Authors. All rights reserved.
Licensed under the BSD-3 License
*/
func Walk(r *mux.Router) {
	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	fmt.Printf("in TestMain\n")

	testConfigFile = "/app/xconfadmin/xconfadmin.conf"
	if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
		testConfigFile = "../config/sample_xconfadmin.conf"
		if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
			panic(fmt.Errorf("config file problem %v", err))
		}
	}
	fmt.Printf("testConfigFile=%v\n", testConfigFile)

	os.Setenv("SECURITY_TOKEN_KEY", "testSecurityTokenKey")

	xpcKey := os.Getenv("XPC_KEY")
	if len(xpcKey) == 0 {
		os.Setenv("XPC_KEY", "testXpcKey")
	}

	cid := os.Getenv("SAT_CLIENT_ID")
	if len(cid) == 0 {
		os.Setenv("SAT_CLIENT_ID", "foo")
	}

	sec := os.Getenv("SAT_CLIENT_SECRET")
	if len(sec) == 0 {
		os.Setenv("SAT_CLIENT_SECRET", "bar")
	}
	cid = os.Getenv("IDP_CLIENT_ID")
	if len(cid) == 0 {
		os.Setenv("IDP_CLIENT_ID", "foo")
	}

	sec = os.Getenv("IDP_CLIENT_SECRET")
	if len(sec) == 0 {
		os.Setenv("IDP_CLIENT_SECRET", "bar")
	}

	ssrKeys := os.Getenv("X1_SSR_KEYS")
	if len(ssrKeys) == 0 {
		os.Setenv("X1_SSR_KEYS", "test-key-1;test-key-2;test-key3")
	}

	PartnerKeys := os.Getenv("PARTNER_KEYS")
	if len(PartnerKeys) == 0 {
		os.Setenv("PARTNER_KEYS", "test")
	}

	var err error
	sc, err = common.NewServerConfig(testConfigFile)
	if err != nil {
		panic(err)
	}

	server = oshttp.NewWebconfigServer(sc, true, nil, nil)
	defer server.XW_XconfServer.Server.Close()
	xwhttp.InitSatTokenManager(server.XW_XconfServer)

	// start clean
	db.SetDatabaseClient(server.XW_XconfServer.DatabaseClient)
	defer server.XW_XconfServer.DatabaseClient.Close()

	// setup router
	router = server.XW_XconfServer.GetRouter(false)

	// setup Xconf APIs and tables
	dataapi.XconfSetup(server.XW_XconfServer, router)
	adminapi.XconfSetup(server, router)
	taggingapi.XconfTaggingServiceSetup(server, router)

	// tear down to start clean
	err = server.XW_XconfServer.SetUp()
	if err != nil {
		panic(err)
	}
	err = server.XW_XconfServer.TearDown()
	if err != nil {
		panic(err)
	}
	// DeleteAllEntities()

	globAut = newApiUnitTest(nil)

	returnCode := m.Run()

	globAut.t = nil

	// tear down to clean up
	server.XW_XconfServer.TearDown()

	os.Exit(returnCode)
}

type apiUnitTest struct {
	t        *testing.T
	router   *mux.Router
	savedMap map[string]string
}

func newApiUnitTest(t *testing.T) *apiUnitTest {
	if globAut != nil {
		globAut.t = t
		return globAut
	}
	aut := apiUnitTest{}
	aut.t = t
	aut.router = router
	aut.savedMap = make(map[string]string)

	globAut = &aut
	return &aut
}

func ExecuteRequest(r *http.Request, handler http.Handler) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func GetTestConfig() string {
	return testConfig
}

func GetTestWebConfigServer(testConfigFile string) (*oshttp.WebconfigServer, *mux.Router) {
	if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
		testConfigFile = "../../config/sample_xconfadmin.conf"
		if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
			panic(fmt.Errorf("config file problem %v", err))
		}
	}

	// set env variables
	os.Setenv("SAT_CLIENT_ID", "foo")
	os.Setenv("SAT_CLIENT_SECRET", "bar")

	os.Setenv("IDP_CLIENT_ID", "foo")
	os.Setenv("IDP_CLIENT_SECRET", "bar")

	var err error
	sc, err = common.NewServerConfig(testConfigFile)
	if err != nil {
		panic(err)
	}
	server := oshttp.NewWebconfigServer(sc, true, nil, nil)
	xwhttp.InitSatTokenManager(server.XW_XconfServer)
	router := server.XW_XconfServer.GetRouter(true)
	dataapi.XconfSetup(server.XW_XconfServer, router)
	adminapi.XconfSetup(server, router)
	return server, router
}

func SetupSatServiceMockServerOkResponse(t *testing.T, server oshttp.WebconfigServer) *httptest.Server {
	mockedSatResponse := []byte(`{"access_token":"one_mock_token","expires_in":86400,"scope":"scope1 scope2 scope3","token_type":"Bearer"}`)
	satServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(mockedSatResponse)
		}))
	server.XW_XconfServer.SatServiceConnector.SetSatServiceHost(satServiceMockServer.URL)
	targetSatHost := server.XW_XconfServer.SatServiceConnector.SatServiceHost()
	assert.Equal(t, satServiceMockServer.URL, targetSatHost)
	return satServiceMockServer
}

func SetupTaggingMockServerOkResponse(t *testing.T, server oshttp.WebconfigServer, path string) *httptest.Server {
	mockedTaggingResponse := []byte(`["value1", "value2", "value3"]`)
	taggingMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedTaggingResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.XW_XconfServer.SetTaggingHost(taggingMockServer.URL)
	targetTaggingHost := server.XW_XconfServer.TaggingHost()
	assert.Equal(t, taggingMockServer.URL, targetTaggingHost)
	return taggingMockServer
}

func SetupTaggingMockServerOkResponseDynamic(t *testing.T, server oshttp.WebconfigServer, response string, path string) *httptest.Server {
	mockedTaggingResponse := []byte(response)
	taggingMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedTaggingResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))

	server.XW_XconfServer.TaggingConnector.SetTaggingHost(taggingMockServer.URL)
	targetTaggingHost := server.XW_XconfServer.TaggingConnector.TaggingHost()
	assert.Equal(t, taggingMockServer.URL, targetTaggingHost)
	return taggingMockServer
}

func SetupAccountServiceMockServerOkResponse(t *testing.T, server oshttp.WebconfigServer, path string) *httptest.Server {
	mockedAccountResponse := []byte(`[{"data":{"serviceAccountId":"testServiceAccountUri","partner":"testPartnerId"},"id":"testId"}]`)
	accountMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedAccountResponse)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.XW_XconfServer.SetAccountServiceHost(accountMockServer.URL)
	targetAccountHost := server.XW_XconfServer.AccountServiceConnector.AccountServiceHost()
	assert.Equal(t, accountMockServer.URL, targetAccountHost)
	return accountMockServer
}

func SetupAccountServiceMockServerOkResponseDynamic(t *testing.T, server oshttp.WebconfigServer, response []byte, path string) *httptest.Server {
	accountMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(response)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.XW_XconfServer.AccountServiceConnector.SetAccountServiceHost(accountMockServer.URL)
	targetAccountHost := server.XW_XconfServer.AccountServiceConnector.AccountServiceHost()
	assert.Equal(t, accountMockServer.URL, targetAccountHost)
	return accountMockServer
}

func SetupAccountServiceMockServerOkResponseDynamicTwoCalls(t *testing.T, server oshttp.WebconfigServer, response []byte, response2 []byte, path string, path2 string) *httptest.Server {
	accountMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(response)
			} else if strings.Contains(r.RequestURI, path2) {
				w.WriteHeader(http.StatusOK)
				w.Write(response2)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.XW_XconfServer.SetAccountServiceHost(accountMockServer.URL)
	targetAccountHost := server.XW_XconfServer.AccountServiceHost()
	assert.Equal(t, accountMockServer.URL, targetAccountHost)
	return accountMockServer
}

func SetupDeviceServiceMockServerOkResponseDynamic(t *testing.T, server oshttp.WebconfigServer, response []byte, path string) *httptest.Server {
	deviceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(response)
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.XW_XconfServer.SetDeviceServiceHost(deviceMockServer.URL)
	targetOdpHost := server.XW_XconfServer.DeviceServiceHost()
	assert.Equal(t, deviceMockServer.URL, targetOdpHost)
	return deviceMockServer
}
func SetupTaggingMockServer500Response(t *testing.T, server oshttp.WebconfigServer, path string) *httptest.Server {
	taggingMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error Msg"))
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.XW_XconfServer.SetTaggingHost(taggingMockServer.URL)
	targetTaggingHost := server.XW_XconfServer.TaggingHost()
	assert.Equal(t, taggingMockServer.URL, targetTaggingHost)
	return taggingMockServer
}

func SetupTaggingMockServer404Response(t *testing.T, server oshttp.WebconfigServer, path string) *httptest.Server {
	taggingMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Error Msg"))
			} else {
				// fail because request was not matched
				assert.Equal(t, true, false)
			}
		}))
	server.XW_XconfServer.SetTaggingHost(taggingMockServer.URL)
	targetTaggingHost := server.XW_XconfServer.TaggingHost()
	assert.Equal(t, taggingMockServer.URL, targetTaggingHost)
	return taggingMockServer
}

func SetupSatServiceMockServerErrorResponse(t *testing.T, server oshttp.WebconfigServer) *httptest.Server {
	satServiceMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
	server.XW_XconfServer.SetSatServiceHost(satServiceMockServer.URL)
	targetSatServiceHost := server.XW_XconfServer.SatServiceHost()
	assert.Equal(t, satServiceMockServer.URL, targetSatServiceHost)
	return satServiceMockServer
}
