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

// Helper functions for feature package tests
package feature_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/rdkcentral/xconfadmin/adminapi"
	oshttp "github.com/rdkcentral/xconfadmin/http"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
)

// Test constants
const (
	defaultModelId           = "X1-1"
	defaultEnvironmentId     = "DEV"
	defaultServiceAccountUri = "defaultServiceAccountUri"
	defaultAccountId         = "defaultAccountId"
	defaultPartnerId         = "defaultpartnerid"
	defaultTimeZone          = "Australia/Brisbane"
	API_VERSION              = "2"
)

// GetTestWebConfigServer returns a configured test server and router
func GetTestWebConfigServer(testConfigFile string) (*oshttp.WebconfigServer, *mux.Router) {
	if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
		testConfigFile = "../../../config/sample_xconfadmin.conf"
		if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
			panic(fmt.Errorf("config file problem %v", err))
		}
	}

	// set env variables
	os.Setenv("SAT_CLIENT_ID", "foo")
	os.Setenv("SAT_CLIENT_SECRET", "bar")
	os.Setenv("IDP_CLIENT_ID", "foo")
	os.Setenv("IDP_CLIENT_SECRET", "bar")

	sc, err := common.NewServerConfig(testConfigFile)
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

// ExecuteRequest executes an HTTP request for testing
func ExecuteRequest(r *http.Request, handler http.Handler) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

// DeleteAllEntities clears all database tables
func DeleteAllEntities() {
	for _, tableInfo := range db.GetAllTableInfo() {
		if err := truncateTable(tableInfo.TableName); err != nil {
			fmt.Printf("failed to truncate table %s\n", tableInfo.TableName)
		}
		if tableInfo.CacheData {
			db.GetCachedSimpleDao().RefreshAll(tableInfo.TableName)
		}
	}
}

func truncateTable(tableName string) error {
	dbClient := db.GetDatabaseClient()
	cassandraClient, ok := dbClient.(*db.CassandraClient)
	if ok {
		return cassandraClient.DeleteAllXconfData(tableName)
	}
	return nil
}

// CreateCondition creates a rule condition
func CreateCondition(freeArg re.FreeArg, operation string, fixedArgValue string) *re.Condition {
	return re.NewCondition(&freeArg, operation, re.NewFixedArg(fixedArgValue))
}

// CreateRuleKeyValue creates a simple key-value rule
func CreateRuleKeyValue(key string, value string) *re.Rule {
	condition := CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, key), re.StandardOperationIs, value)
	return &re.Rule{
		Condition: condition,
	}
}

// CreateExistsRule creates a rule that checks if a tag exists
func CreateExistsRule(tagName string) *re.Rule {
	condition := CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeAny, tagName), re.StandardOperationExists, "")
	rule := &re.Rule{
		Condition: condition,
	}
	return rule
}

// SetupSatServiceMockServerOkResponse creates a mock SAT service that returns success
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

// SetupTaggingMockServerOkResponseDynamic creates a mock tagging server with dynamic response
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

// SetupTaggingMockServer404Response creates a mock tagging server that returns 404
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

// SetupTaggingMockServer500Response creates a mock tagging server that returns 500
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

// SetupAccountServiceMockServerOkResponse creates a mock account service that returns success
func SetupAccountServiceMockServerOkResponse(t *testing.T, server oshttp.WebconfigServer, path string) *httptest.Server {
	mockedAccountResponse := []byte(`[{"data":{"serviceAccountId":"testServiceAccountUri","partner":"testPartnerId"},"id":"testId"}]`)
	accountMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(mockedAccountResponse)
			} else {
				assert.Equal(t, true, false)
			}
		}))
	server.XW_XconfServer.SetAccountServiceHost(accountMockServer.URL)
	targetAccountHost := server.XW_XconfServer.AccountServiceHost()
	assert.Equal(t, accountMockServer.URL, targetAccountHost)
	return accountMockServer
}

// SetupAccountServiceMockServerOkResponseDynamic creates a mock account service with dynamic response
func SetupAccountServiceMockServerOkResponseDynamic(t *testing.T, server oshttp.WebconfigServer, response []byte, path string) *httptest.Server {
	accountMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusOK)
				w.Write(response)
			} else {
				assert.Equal(t, true, false)
			}
		}))
	server.XW_XconfServer.AccountServiceConnector.SetAccountServiceHost(accountMockServer.URL)
	targetAccountHost := server.XW_XconfServer.AccountServiceConnector.AccountServiceHost()
	assert.Equal(t, accountMockServer.URL, targetAccountHost)
	return accountMockServer
}

// SetupAccountServiceMockServerOkResponseDynamicTwoCalls creates a mock account service with two different responses
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

// SetupAccountServiceMockServer404Response creates a mock account service that returns 404
func SetupAccountServiceMockServer404Response(t *testing.T, server oshttp.WebconfigServer, path string) *httptest.Server {
	accountMockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, path) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Error Msg"))
			} else {
				assert.Equal(t, true, false)
			}
		}))
	server.XW_XconfServer.SetAccountServiceHost(accountMockServer.URL)
	targetAccountHost := server.XW_XconfServer.AccountServiceHost()
	assert.Equal(t, accountMockServer.URL, targetAccountHost)
	return accountMockServer
}

// CreateAccountPartnerObject creates an account partner object for testing
func CreateAccountPartnerObject(partnerId string) xwhttp.AccountServiceDevices {
	accountObject := xwhttp.AccountServiceDevices{
		Id: uuid.New().String(),
		DeviceData: xwhttp.DeviceData{
			Partner:           partnerId,
			ServiceAccountUri: defaultServiceAccountUri,
		},
	}
	return accountObject
}

// CreateODPPartnerObject creates an ODP partner object for testing
func CreateODPPartnerObject() xwhttp.DeviceServiceObject {
	odpObject := xwhttp.DeviceServiceObject{
		Status: 200,
		DeviceServiceData: &xwhttp.DeviceServiceData{
			AccountId: defaultServiceAccountUri,
		}}
	return odpObject
}

// CreateODPPartnerObjectWithPartnerAndTimezone creates an ODP partner object with partner and timezone
func CreateODPPartnerObjectWithPartnerAndTimezone() xwhttp.DeviceServiceObject {
	odpObject := xwhttp.DeviceServiceObject{
		Status: 200,
		DeviceServiceData: &xwhttp.DeviceServiceData{
			AccountId: defaultServiceAccountUri,
			PartnerId: defaultPartnerId,
			TimeZone:  defaultTimeZone,
		}}
	return odpObject
}

// CreateODPPartnerObjectWithPartnerAndTimezoneInvalid creates an ODP partner object with invalid timezone
func CreateODPPartnerObjectWithPartnerAndTimezoneInvalid() xwhttp.DeviceServiceObject {
	odpObject := xwhttp.DeviceServiceObject{
		Status: 200,
		DeviceServiceData: &xwhttp.DeviceServiceData{
			AccountId: defaultServiceAccountUri,
			PartnerId: defaultPartnerId,
			TimeZone:  "InvalidTimeZone",
		}}
	return odpObject
}
