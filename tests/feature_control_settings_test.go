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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"testing"

	oshttp "github.com/rdkcentral/xconfadmin/http"

	"github.com/rdkcentral/xconfwebconfig/dataapi"

	"github.com/rdkcentral/xconfwebconfig/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared/rfc"
	xutils "github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gotest.tools/assert"
)

const (
	testFile                                     = "../config/sample_xconfadmin.conf"
	PARTNER_TAG                                  = "partnerTag"
	MAC_ADDRESS_TAG                              = "macAddressTag"
	ACCOUNT_TAG                                  = "macAddressTag"
	MAC_AND_PARTNER_TAG                          = "macAndPartnerTag"
	ACCOUNT_HASH_TAG                             = "accountHashTag"
	PARTNER                                      = "COMCAST"
	MAC_ADDRESS                                  = "11:22:33:44:55:66"
	URL_TAGS_MAC_ADDRESS                         = "/getTagsForMacAddress/%s"
	URL_TAGS_PARTNER                             = "/getTagsForPartner/%s"
	URL_TAGS_PARTNER_AND_MAC_ADDRESS             = "/getTagsForPartnerAndMacAddress/partner/%s/macaddress/%s"
	URL_TAGS_MAC_ADDRESS_AND_ACCOUNT             = "/getTagsForMacAddressAndAccount/macaddress/%s/account/%s"
	URL_TAGS_ACCOUNT                             = "/getTagsForAccount/%s"
	URL_TAGS_PARTNER_AND_MAC_ADDRESS_AND_ACCOUNT = "/getTagsForPartnerAndMacAddressAndAccount/partner/%s/macaddress/%s/account/%s"
	URL_TAGS_PARTNER_AND_ACCOUNT                 = "/getTagsForPartnerAndAccount/partner/%s/account/%s"
	URL_ODP                                      = "/api/v1/operational/mesh-pod/%s/account"
	URL_ACCOUNT_ESTB                             = "/devices?hostMac=%s&status=Active"
	URL_ACCOUNT_ECM                              = "/devices?ecmMac=%s&status=Active"
)

func TestFeatureSetting(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, PARTNER))
	defer taggingMockServer.Close()

	featureIds := []string{}
	features := []rfc.FeatureResponse{}
	for i := 0; i < 5; i++ {
		feature := createAndSaveFeature()
		featureIds = append(featureIds, feature.ID)
		featureResponse := rfc.CreateFeatureResponseObject(*feature)
		features = append(features, featureResponse)
	}

	createAndSaveFeatureRule(featureIds, createRule(CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "model"), re.StandardOperationIs, "X1-1")), "stb")
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, "?model=X1-1", nil, features)
}

func TestFeatureSettingByApplicationType(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, PARTNER))
	defer taggingMockServer.Close()

	features := createAndSaveFeatures()
	createAndSaveFeatureRules(features)
	featureResponseStb := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*features["stb"]),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, "?model=X1-1", nil, featureResponseStb)
	featureResponseRDK := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*features["rdkcloud"]),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, "/rdkcloud?model=X1-1", nil, featureResponseRDK)
}

// func Test304StatusIfResponseWasNotModified(t *testing.T) {
// 	DeleteAllEntities()
// 	server, router := GetTestWebConfigServer(testFile)

// 	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, PARTNER))
// 	defer taggingMockServer.Close()

// 	feature := createAndSaveFeature()
// 	rule := CreateDefaultEnvModelRule()
// 	featureRule := createFeatureRule([]string{feature.ID}, rule, "stb")
// 	setFeatureRule(featureRule)
// 	featureResponse := []rfc.FeatureResponse{
// 		rfc.CreateFeatureResponseObject(*feature),
// 	}
// 	featureControlRuleBase := featurecontrol.NewFeatureControlRuleBase()
// 	configSetHash := featureControlRuleBase.CalculateHash(featureResponse)
// 	assertNotMofifiedStatus(t, server, router, configSetHash, featureResponse)
// 	assertConfigSetHashChange(t, server, router, configSetHash, featureResponse)
// }

func TestIfFeatureRuleIsAppliedByRangeOperation(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServer404Response(t, *server, fmt.Sprintf(URL_TAGS_MAC_ADDRESS, "B4:F2:E8:15:67:46"))
	defer taggingMockServer.Close()

	feature := createAndSaveFeature()
	createAndSaveFeatureRule([]string{feature.ID}, createPercentRangeRule(), "stb")
	featureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	macFits50To100Range := "B4:F2:E8:15:67:46"
	verifyPercentRangeRuleApplying(t, server, router, macFits50To100Range, featureResponse)
}

func TestIfFeatureRuleIsNotAppliedByRangeOperation(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServer404Response(t, *server, fmt.Sprintf(URL_TAGS_MAC_ADDRESS, "04:02:10:00:00:01"))
	defer taggingMockServer.Close()

	feature := createAndSaveFeature()
	createAndSaveFeatureRule([]string{feature.ID}, createPercentRangeRule(), "stb")
	featureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	macDoesntFit50To100Range := "04:02:10:00:00:01"
	verifyPercentRangeRuleApplying(t, server, router, macDoesntFit50To100Range, featureResponse)
}

func TestFeatureInstanceFieldAddedToRFCResponse(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, PARTNER))
	defer taggingMockServer.Close()

	feature := createAndSaveFeature()
	rule := CreateRuleKeyValue("model", strings.ToUpper(defaultModelId))
	createAndSaveFeatureRule([]string{feature.ID}, rule, "stb")
	performGetSettingsRequestAndVerifyFeatureControlInstanceName(t, server, router, fmt.Sprintf("?version=%s&applicationType=stb&model=%s", API_VERSION, defaultModelId), feature)
}

func TestFeatureIsReturnedForPartnerTag(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, PARTNER))
	defer taggingMockServer.Close()

	feature := createTagFeatureRule(PARTNER_TAG)
	featureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?partnerId=%s", PARTNER), nil, featureResponse)
}

func TestFeatureIsNotReturnedForUnknownPartnerTag(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, PARTNER))
	defer taggingMockServer.Close()

	createTagFeatureRule(PARTNER_TAG)
	emptyFeatureResponse := []rfc.FeatureResponse{}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, "?partnerId=unknown", nil, emptyFeatureResponse)
}

func TestFeatureIsReturnedForMacAddressTag(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, PARTNER))
	defer taggingMockServer.Close()

	SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_MAC_ADDRESS, MAC_ADDRESS))
	feature := createTagFeatureRule(MAC_ADDRESS_TAG)
	featureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?estbMacAddress=%s", MAC_ADDRESS), nil, featureResponse)
}

func TestFeatureIsReturnedForPartnerAndMacAddressTag(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_AND_PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_MAC_ADDRESS, PARTNER, MAC_ADDRESS))
	defer taggingMockServer.Close()

	feature := createTagFeatureRule(MAC_AND_PARTNER_TAG)
	featureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?estbMacAddress=%s&partnerId=%s", MAC_ADDRESS, PARTNER), nil, featureResponse)
}

func TestFeatureIsReturnedForPartnerWhenMacIsInvalid(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, PARTNER))
	defer taggingMockServer.Close()

	feature := createTagFeatureRule(PARTNER_TAG)
	featureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?estbMacAddress=abc&partnerId=%s", PARTNER), nil, featureResponse)
}

func Test200StatusCodeWhenTaggingServiceUnavailableAndEmptyConfigHash(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)

	taggingMockServer := SetupTaggingMockServer500Response(t, *server, fmt.Sprintf(URL_TAGS_PARTNER, PARTNER))
	defer taggingMockServer.Close()

	emptyFeatureResponse := []rfc.FeatureResponse{}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?partnerId=%s", PARTNER), nil, emptyFeatureResponse)
}

func TestGetFeatureSettingByUnknownPartnerId(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	// Xc.
	accountObjectArray := []xwhttp.AccountServiceDevices{
		CreateAccountPartnerObject(PARTNER),
	}
	expectedResponse, _ := json.Marshal(accountObjectArray)
	accountMockServer := SetupAccountServiceMockServerOkResponseDynamic(t, *server, expectedResponse, fmt.Sprintf(URL_ACCOUNT_ESTB, MAC_ADDRESS))
	defer accountMockServer.Close()
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_AND_PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_MAC_ADDRESS, PARTNER, MAC_ADDRESS))
	defer taggingMockServer.Close()
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*getPartnerFeature(PARTNER)),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?partnerId=unknown&estbMacAddress=%s", MAC_ADDRESS), nil, expectedFeatureResponse)
}

func TestGetFeatureByAccountIdAndMacAddressTag(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_MAC_ADDRESS_AND_ACCOUNT, "AA:AA:AA:AA:AA:AA", defaultAccountId))
	defer taggingMockServer.Close()
	feature := createTagFeatureRule(ACCOUNT_TAG)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
		rfc.CreateFeatureResponseObject(*getAccountIdFeature(defaultAccountId)),
	}
	headers := map[string]string{
		"HA-Haproxy-xconf-http": "xconf-https",
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=%s&estbMacAddress=AA:AA:AA:AA:AA:AA", defaultAccountId), headers, expectedFeatureResponse)
}

func TestGetFeatureByUnknownAccountHash(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	accountObjectArray := []xwhttp.AccountServiceDevices{
		CreateAccountPartnerObject(PARTNER),
	}
	expectedResponse, _ := json.Marshal(accountObjectArray)
	accountMockServer := SetupAccountServiceMockServerOkResponseDynamic(t, *server, expectedResponse, fmt.Sprintf(URL_ACCOUNT_ESTB, MAC_ADDRESS))
	defer accountMockServer.Close()
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_MAC_ADDRESS, PARTNER, MAC_ADDRESS))
	defer taggingMockServer.Close()
	calculatedConfigSetHash := xutils.CalculateHash(defaultServiceAccountUri)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*getAccountHashFeature(calculatedConfigSetHash)),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountHash=unknown&estbMacAddress=%s", MAC_ADDRESS), nil, expectedFeatureResponse)
}

func TestGetAccountIdBySecondAccountCall(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	accountObjectArray := []xwhttp.AccountServiceDevices{}
	accountObjectArray2 := []xwhttp.AccountServiceDevices{
		CreateAccountPartnerObject(PARTNER),
	}
	expectedResponse, _ := json.Marshal(accountObjectArray)
	expectedResponse2, _ := json.Marshal(accountObjectArray2)
	estbMac := "AA:AA:AA:AA:AA:AA"
	ecmMac := "BB:BB:BB:BB:BB:BB"
	accountMockServer := SetupAccountServiceMockServerOkResponseDynamicTwoCalls(t, *server, expectedResponse, expectedResponse2, fmt.Sprintf(URL_ACCOUNT_ESTB, estbMac), fmt.Sprintf(URL_ACCOUNT_ECM, ecmMac))
	defer accountMockServer.Close()
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_MAC_ADDRESS_AND_ACCOUNT, accountObjectArray2[0].DeviceData.Partner, estbMac, accountObjectArray2[0].DeviceData.ServiceAccountUri))
	defer taggingMockServer.Close()
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*getAccountIdFeature(defaultServiceAccountUri)),
	}
	var headers = make(map[string]string)
	headers["HA-Haproxy-xconf-http"] = "xconf-https"
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=unknown&estbMacAddress=%s&ecmMacAddress=%s", estbMac, ecmMac), headers, expectedFeatureResponse)
}

// func TestGetAccountIdByOdpCallWithPartnerAndTimezoneKnown(t *testing.T) {
// 	DeleteAllEntities()
// 	server, router := GetTestWebConfigServer(testFile)
// 	serialNum := "P1K648058000"
// 	odpObject := CreateODPPartnerObjectWithPartnerAndTimezone()
// 	expectedResponse, _ := json.Marshal(odpObject)
// 	odpMockServer := SetupDeviceServiceMockServerOkResponseDynamic(t, *server, expectedResponse, fmt.Sprintf(URL_ODP, serialNum))
// 	defer odpMockServer.Close()
// 	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, ACCOUNT_TAG), fmt.Sprintf(URL_TAGS_ACCOUNT, odpObject.DeviceServiceData.AccountId))
// 	defer taggingMockServer.Close()
// 	expectedAccountIdFeatureResponse := rfc.CreateFeatureResponseObject(*getAccountIdFeature(defaultServiceAccountUri))
// 	expectedAccountIdFeatureResponse["accountId"] = defaultServiceAccountUri
// 	expectedAccountIdFeatureResponse["partnerId"] = defaultPartnerId
// 	expectedAccountIdFeatureResponse["timeZone"] = defaultTimeZone
// 	expectedAccountIdFeatureResponse["tzUTCOffset"] = "UTC+10:00"
// 	expectedFeatureResponse := []rfc.FeatureResponse{
// 		expectedAccountIdFeatureResponse,
// 	}
// 	headers := map[string]string{
// 		"HA-Haproxy-xconf-http": "xconf-https",
// 	}
// 	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=unknown&serialNum=%s&accountMgmt=xpc", serialNum), headers, expectedFeatureResponse)
// }

// func TestGetAccountIdByOdpCallWithPartnerAndTimezoneKnownButInvalid(t *testing.T) {
// 	DeleteAllEntities()
// 	server, router := GetTestWebConfigServer(testFile)
// 	serialNum := "P1K648058000"
// 	odpObject := CreateODPPartnerObjectWithPartnerAndTimezoneInvalid()
// 	expectedResponse, _ := json.Marshal(odpObject)
// 	odpMockServer := SetupDeviceServiceMockServerOkResponseDynamic(t, *server, expectedResponse, fmt.Sprintf(URL_ODP, serialNum))
// 	defer odpMockServer.Close()
// 	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, ACCOUNT_TAG), fmt.Sprintf(URL_TAGS_ACCOUNT, odpObject.DeviceServiceData.AccountId))
// 	defer taggingMockServer.Close()
// 	expectedAccountIdFeatureResponse := rfc.CreateFeatureResponseObject(*getAccountIdFeature(defaultServiceAccountUri))
// 	expectedAccountIdFeatureResponse["accountId"] = defaultServiceAccountUri
// 	expectedAccountIdFeatureResponse["partnerId"] = defaultPartnerId
// 	expectedAccountIdFeatureResponse["timeZone"] = "InvalidTimeZone"
// 	expectedAccountIdFeatureResponse["tzUTCOffset"] = "unknown"
// 	expectedFeatureResponse := []rfc.FeatureResponse{
// 		expectedAccountIdFeatureResponse,
// 	}
// 	headers := map[string]string{
// 		"HA-Haproxy-xconf-http": "xconf-https",
// 	}
// 	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=unknown&serialNum=%s&accountMgmt=xpc", serialNum), headers, expectedFeatureResponse)
// }

// func TestGetAccountIdByOdpCallWithPartnerAndTimezoneUnknown(t *testing.T) {
// 	DeleteAllEntities()
// 	server, router := GetTestWebConfigServer(testFile)
// 	serialNum := "P1K648058000"
// 	odpObject := CreateODPPartnerObject()
// 	expectedResponse, _ := json.Marshal(odpObject)
// 	odpMockServer := dataapi.SetupDeviceServiceMockServerOkResponseDynamic(t, *server, expectedResponse, fmt.Sprintf(URL_ODP, serialNum))
// 	defer odpMockServer.Close()
// 	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, ACCOUNT_TAG), fmt.Sprintf(URL_TAGS_ACCOUNT, odpObject.DeviceServiceData.AccountId))
// 	defer taggingMockServer.Close()
// 	expectedAccountIdFeatureResponse := rfc.CreateFeatureResponseObject(*getAccountIdFeature(defaultServiceAccountUri))
// 	expectedAccountIdFeatureResponse["accountId"] = defaultServiceAccountUri
// 	expectedAccountIdFeatureResponse["partnerId"] = "unknown"
// 	expectedAccountIdFeatureResponse["timeZone"] = "unknown"
// 	expectedAccountIdFeatureResponse["tzUTCOffset"] = "unknown"
// 	expectedFeatureResponse := []rfc.FeatureResponse{
// 		expectedAccountIdFeatureResponse,
// 	}
// 	headers := map[string]string{
// 		"HA-Haproxy-xconf-http": "xconf-https",
// 	}
// 	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=unknown&serialNum=%s&accountMgmt=xpc", serialNum), headers, expectedFeatureResponse)
// }

func TestDontCallAccountSecondTimeIfFirstCallSuccessful(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	accountObjectArray := []xwhttp.AccountServiceDevices{
		CreateAccountPartnerObject(PARTNER),
	}
	accountObjectArray2 := []xwhttp.AccountServiceDevices{}
	expectedResponse, _ := json.Marshal(accountObjectArray)
	expectedResponse2, _ := json.Marshal(accountObjectArray2)
	estbMac := "AA:AA:AA:AA:AA:AA"
	ecmMac := "BB:BB:BB:BB:BB:BB"
	accountMockServer := SetupAccountServiceMockServerOkResponseDynamicTwoCalls(t, *server, expectedResponse, expectedResponse2, fmt.Sprintf(URL_ACCOUNT_ESTB, estbMac), fmt.Sprintf(URL_ACCOUNT_ECM, ecmMac))
	defer accountMockServer.Close()
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_MAC_ADDRESS_AND_ACCOUNT, accountObjectArray[0].DeviceData.Partner, estbMac, accountObjectArray[0].DeviceData.ServiceAccountUri))
	defer taggingMockServer.Close()
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*getAccountIdFeature(defaultServiceAccountUri)),
	}
	headers := map[string]string{
		"HA-Haproxy-xconf-http": "xconf-https",
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=unknown&estbMacAddress=%s&ecmMacAddress=%s", estbMac, ecmMac), headers, expectedFeatureResponse)
}

func TestGetFeatureSettingByUnknownAccountId(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*getAccountIdFeature(defaultServiceAccountUri)),
	}
	accountObjectArray := []xwhttp.AccountServiceDevices{
		CreateAccountPartnerObject(PARTNER),
	}
	expectedResponse, _ := json.Marshal(accountObjectArray)
	accountMockServer := SetupAccountServiceMockServerOkResponseDynamic(t, *server, expectedResponse, fmt.Sprintf(URL_ACCOUNT_ESTB, "AA:AA:AA:AA:AA:AA"))
	defer accountMockServer.Close()
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_MAC_ADDRESS_AND_ACCOUNT, accountObjectArray[0].DeviceData.Partner, "AA:AA:AA:AA:AA:AA", accountObjectArray[0].DeviceData.ServiceAccountUri))
	defer taggingMockServer.Close()
	httpsheaders := map[string]string{
		"HA-Haproxy-xconf-http": "xconf-https",
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, "?accountId=unknown&estbMacAddress=AA:AA:AA:AA:AA:AA", httpsheaders, expectedFeatureResponse)

	// with xconf http header (insecure)
	headers := map[string]string{
		"HA-Haproxy-xconf-http": "xconf-http",
	}
	emptyFeatureResponse := []rfc.FeatureResponse{}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=unknown&estbMacAddress=AA:AA:AA:AA:AA:AA"), headers, emptyFeatureResponse)
}

func TestGetFeatureByAccountIdTag(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	ruleFeature := createTagFeatureRule(ACCOUNT_TAG)
	accountIdFeature := getAccountIdFeature(defaultAccountId)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*ruleFeature),
		rfc.CreateFeatureResponseObject(*accountIdFeature),
	}
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_ACCOUNT, defaultAccountId))
	defer taggingMockServer.Close()
	headers := map[string]string{
		"HA-Haproxy-xconf-http": "xconf-https",
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=%s", defaultAccountId), headers, expectedFeatureResponse)
}

func TestGetFeatureByPartnerIdAsFeatureRuleParameter(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*getPartnerFeature(defaultPartnerId)),
	}
	featureFromRule := createAndSaveFeature()
	rule := createRule(CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "partnerId"), re.StandardOperationIs, strings.ToUpper(defaultPartnerId)))
	createAndSaveFeatureRule([]string{featureFromRule.ID}, rule, "stb")
	accountObjectArray := []xwhttp.AccountServiceDevices{
		CreateAccountPartnerObject(PARTNER),
	}
	expectedResponse, _ := json.Marshal(accountObjectArray)
	accountMockServer := SetupAccountServiceMockServerOkResponseDynamic(t, *server, expectedResponse, fmt.Sprintf(URL_ACCOUNT_ESTB, MAC_ADDRESS))
	defer accountMockServer.Close()
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, ACCOUNT_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_MAC_ADDRESS, accountObjectArray[0].DeviceData.Partner, MAC_ADDRESS))
	defer taggingMockServer.Close()
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?partnerId=unknown&estbMacAddress=%s", MAC_ADDRESS), nil, expectedFeatureResponse)
}

func TestGetAccountHashFeatureIfAccountHashIsPassed(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	accountHash := xutils.CalculateHash(defaultServiceAccountUri)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*getAccountHashFeature(accountHash)),
	}
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_MAC_ADDRESS, MAC_ADDRESS))
	defer taggingMockServer.Close()
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountHash=%s&estbMacAddress=%s", accountHash, MAC_ADDRESS), nil, expectedFeatureResponse)
}

func TestGetAccountIdFeatureIfAccountIdIsPassed(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*getAccountIdFeature(defaultServiceAccountUri)),
	}
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_MAC_ADDRESS_AND_ACCOUNT, "AA:AA:AA:AA:AA:AA", defaultServiceAccountUri))
	defer taggingMockServer.Close()
	headers := map[string]string{
		"HA-Haproxy-xconf-http": "xconf-https",
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=%s&estbMacAddress=%s", defaultServiceAccountUri, "AA:AA:AA:AA:AA:AA"), headers, expectedFeatureResponse)
}

func TestGetAccountIdAndHashFeaturesIfSpecificConfigIsEnabled(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	accountId := "serviceAccountUri"
	accountHash := xutils.CalculateHash(accountId)
	accountIdFeature := getAccountIdFeature(accountId)
	accountHashFeature := getAccountHashFeature(accountHash)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*accountIdFeature),
		rfc.CreateFeatureResponseObject(*accountHashFeature),
	}
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_MAC_ADDRESS_AND_ACCOUNT, "AA:AA:AA:AA:AA:AA", accountId))
	defer taggingMockServer.Close()
	headers := map[string]string{
		"HA-Haproxy-xconf-http": "xconf-https",
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=%s&accountHash=%s&estbMacAddress=%s", accountId, accountHash, "AA:AA:AA:AA:AA:AA"), headers, expectedFeatureResponse)
}

func TestGetFeaturesByAccountIdAndMacAddress(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	accountId := "accountId"
	feature := createTagFeatureRule(MAC_ADDRESS_TAG)
	accountFeature := getAccountIdFeature(accountId)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
		rfc.CreateFeatureResponseObject(*accountFeature),
	}
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_MAC_ADDRESS_AND_ACCOUNT, "AA:AA:AA:AA:AA:AA", accountId))
	defer taggingMockServer.Close()
	headers := map[string]string{
		"HA-Haproxy-xconf-http": "xconf-https",
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=%s&estbMacAddress=%s", accountId, "AA:AA:AA:AA:AA:AA"), headers, expectedFeatureResponse)
}

func TestGetFeaturesByAccountV2(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	dataapi.Xc.ReturnAccountId = false
	accountId := "accountId"
	featureRule := createTagFeatureRule(ACCOUNT_TAG)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*featureRule),
	}
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, ACCOUNT_TAG), fmt.Sprintf(URL_TAGS_ACCOUNT, accountId))
	defer taggingMockServer.Close()
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=%s", accountId), nil, expectedFeatureResponse)
}

func TestGetFeatureByPartnerAndAccountIdV2(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	dataapi.Xc.ReturnAccountId = false
	accountId := "accountId"
	feature := createTagFeatureRule(ACCOUNT_TAG)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, ACCOUNT_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_ACCOUNT, PARTNER, accountId))
	defer taggingMockServer.Close()
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=%s&partnerId=%s", accountId, PARTNER), nil, expectedFeatureResponse)
}

func TestGetFeatureByPartnerMacAddressAndAccountIdV2(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	dataapi.Xc.ReturnAccountId = false

	accountId := "accountId"
	feature := createTagFeatureRule(ACCOUNT_TAG)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, ACCOUNT_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_MAC_ADDRESS_AND_ACCOUNT, PARTNER, "AA:AA:AA:AA:AA:AA", accountId))
	defer taggingMockServer.Close()
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=%s&partnerId=%s&estbMacAddress=%s", accountId, PARTNER, "AA:AA:AA:AA:AA:AA"), nil, expectedFeatureResponse)
}

func TestGetFeatureByAccountIdAsFeatureRuleParameterAndAccountIdFeatureIsNotReturned(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testFile)
	dataapi.Xc.ReturnAccountId = false
	accountHashFeature := getAccountHashFeature(xutils.CalculateHash(defaultServiceAccountUri))
	featureFromRule := createAndSaveFeature()
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*accountHashFeature),
		rfc.CreateFeatureResponseObject(*featureFromRule),
	}
	rule := createRule(CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "accountId"), re.StandardOperationIs, defaultServiceAccountUri))
	createAndSaveFeatureRule([]string{featureFromRule.ID}, rule, "stb")
	accountObjectArray := []xwhttp.AccountServiceDevices{
		CreateAccountPartnerObject(PARTNER),
	}
	expectedResponse, _ := json.Marshal(accountObjectArray)
	accountMockServer := SetupAccountServiceMockServerOkResponseDynamic(t, *server, expectedResponse, fmt.Sprintf(URL_ACCOUNT_ESTB, MAC_ADDRESS))
	defer accountMockServer.Close()
	taggingMockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, ACCOUNT_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_MAC_ADDRESS_AND_ACCOUNT, accountObjectArray[0].DeviceData.Partner, MAC_ADDRESS, accountObjectArray[0].DeviceData.ServiceAccountUri))
	defer taggingMockServer.Close()
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=unknown&accountHash=unknown&estbMacAddress=%s", MAC_ADDRESS), nil, expectedFeatureResponse)
}

func verifyPercentRangeRuleApplying(t *testing.T, server *oshttp.WebconfigServer, router *mux.Router, macAddress string, expectedFeatures []rfc.FeatureResponse) {
	satMockServer := SetupSatServiceMockServerOkResponse(t, *server)
	defer satMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings?estbMacAddress=%s", macAddress)
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	compareFeatureControlResponses(t, res, expectedFeatures)
}

func assertConfigSetHashChange(t *testing.T, server *oshttp.WebconfigServer, router *mux.Router, configSetHash string, expectedFeatures []rfc.FeatureResponse) {
	satMockServer := SetupSatServiceMockServerOkResponse(t, *server)
	defer satMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings?model=%s&env=%s", strings.ToUpper(defaultModelId), strings.ToUpper(defaultEnvironmentId))
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("configSetHash", "")
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.Equal(t, res.Header["configSetHash"][0], configSetHash)
	compareFeatureControlResponses(t, res, expectedFeatures)
}

func assertNotMofifiedStatus(t *testing.T, server *oshttp.WebconfigServer, router *mux.Router, configSetHash string, expectedFeatures []rfc.FeatureResponse) {
	satMockServer := SetupSatServiceMockServerOkResponse(t, *server)
	defer satMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings?model=%s&env=%s", strings.ToUpper(defaultModelId), strings.ToUpper(defaultEnvironmentId))
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("configSetHash", configSetHash)
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotModified)
	assert.Equal(t, res.Header["configSetHash"][0], configSetHash)
	compareFeatureControlResponses(t, res, expectedFeatures)
}

func performGetSettingsRequestAndVerifyFeatureControlInstanceName(t *testing.T, server *oshttp.WebconfigServer, router *mux.Router, extraUrl string, expectedFeature *rfc.Feature) {
	satMockServer := SetupSatServiceMockServerOkResponse(t, *server)
	defer satMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings%s", extraUrl)
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	actualResponse := map[string]rfc.FeatureControl{}
	err = json.Unmarshal(body, &actualResponse)
	assert.NilError(t, err)
	assert.Equal(t, actualResponse["featureControl"].FeatureResponses[0]["featureInstance"], expectedFeature.FeatureName)
	res.Body.Close()
}

func performGetSettingsRequestAndVerifyFeatureControl(t *testing.T, server *oshttp.WebconfigServer, router *mux.Router, extraUrl string, headers map[string]string, expectedFeatures []rfc.FeatureResponse) {
	satMockServer := SetupSatServiceMockServerOkResponse(t, *server)
	defer satMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings%s", extraUrl)
	req, err := http.NewRequest("GET", url, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	compareFeatureControlResponses(t, res, expectedFeatures)
}

func performGetSettingsRequestAndVerify500ErrorWithNonEmptyConfigSetHash(t *testing.T, server *oshttp.WebconfigServer, router *mux.Router, extraUrl string) {
	satMockServer := SetupSatServiceMockServerOkResponse(t, *server)
	defer satMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings%s", extraUrl)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("configSetHash", "nonEmptyValue")
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, res.StatusCode, http.StatusInternalServerError)
	assert.Equal(t, strings.Contains(string(body), "Error Msg"), true)
	res.Body.Close()
}

func compareFeatureControlResponses(t *testing.T, res *http.Response, expectedFeatures []rfc.FeatureResponse) {
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	actualResponse := map[string]rfc.FeatureControl{}
	err = json.Unmarshal(body, &actualResponse)
	assert.NilError(t, err)
	actualFeatureControl, ok := actualResponse["featureControl"]
	assert.Equal(t, ok, true)
	actualFeatures := actualFeatureControl.FeatureResponses
	assert.Equal(t, actualFeatures != nil, true)
	sortFeatures(actualFeatures)
	sortFeatures(expectedFeatures)
	for i := range expectedFeatures {
		assert.Equal(t, len(expectedFeatures[i]), len(actualFeatures[i]))
		for key, value := range expectedFeatures[i] {
			switch v := value.(type) {
			case int:
				assert.Equal(t, value, actualFeatures[i][key].(int))
			case string:
				assert.Equal(t, value, actualFeatures[i][key].(string))
			case bool:
				assert.Equal(t, value, actualFeatures[i][key].(bool))
			case map[string]string:
				for mapK, mapV := range v {
					assert.Equal(t, mapV, actualFeatures[i][key].(map[string]interface{})[mapK].(string))
				}
			// fail if not one of above types so we don't accidentally miss one
			default:
				assert.Equal(t, true, false)
			}
		}
	}
	res.Body.Close()
}

func sortFeatures(features []rfc.FeatureResponse) {
	sort.SliceStable(features, func(i, j int) bool {
		return fmt.Sprintf("%s", features[i]["name"]) < fmt.Sprintf("%s", features[j]["name"])
	})
}

func getPartnerFeature(partnerId string) *rfc.Feature {
	partnerFeature := &rfc.Feature{
		Name:               common.SYNDICATION_PARTNER,
		FeatureName:        common.SYNDICATION_PARTNER,
		EffectiveImmediate: true,
		Enable:             true,
		ConfigData: map[string]string{
			common.TR181_DEVICE_TYPE_PARTNER_ID: strings.ToLower(PARTNER),
		},
	}
	return partnerFeature
}

func getAccountIdFeature(accountId string) *rfc.Feature {
	accountIdFeature := rfc.Feature{
		Name:               "AccountId",
		FeatureName:        "AccountId",
		EffectiveImmediate: true,
		Enable:             true,
		ConfigData: map[string]string{
			common.TR181_DEVICE_TYPE_ACCOUNT_ID: accountId,
		},
	}
	return &accountIdFeature
}

func getAccountHashFeature(accountHash string) *rfc.Feature {
	accountHashFeature := rfc.Feature{
		Name:               "AccountHash",
		FeatureName:        "AccountHash",
		EffectiveImmediate: true,
		Enable:             true,
		ConfigData: map[string]string{
			common.TR181_DEVICE_TYPE_ACCOUNT_HASH: accountHash,
		},
	}
	return &accountHashFeature
}

func createTagFeatureRule(tagNameForRule string) *rfc.Feature {
	feature := createAndSaveFeature()
	createAndSaveFeatureRule([]string{feature.ID}, CreateExistsRule(tagNameForRule), "stb")
	return feature
}

func setFeatureRule(featureRule *rfc.FeatureRule) {
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, featureRule.Id, featureRule)
}

func setFeature(feature *rfc.Feature) {
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_XCONF_FEATURE, feature.ID, feature)
}

func createAndSaveFeature() *rfc.Feature {
	feature := createFeature()
	setFeature(feature)
	return feature
}

func createFeature() *rfc.Feature {
	id := uuid.New().String()
	configData := map[string]string{}
	configData[fmt.Sprintf("%s-key", id)] = fmt.Sprintf("%s-value", id)

	feature := &rfc.Feature{
		ID:                 id,
		Name:               fmt.Sprintf("%s-name", id),
		EffectiveImmediate: false,
		Enable:             false,
		ConfigData:         configData,
	}
	return feature
}

func createAndSaveFeatureWithApplicationTypeAndConfigData(applicationType string) *rfc.Feature {
	feature := createFeatureWithApplicationTypeAndConfigData(applicationType)
	setFeature(feature)
	return feature
}

func createFeatureWithApplicationTypeAndConfigData(applicationType string) *rfc.Feature {
	id := uuid.New().String()
	configData := map[string]string{}
	configData["key"] = "value"

	feature := &rfc.Feature{
		ID:                 id,
		ApplicationType:    applicationType,
		Name:               fmt.Sprintf("%s-name", id),
		EffectiveImmediate: false,
		Enable:             false,
		ConfigData:         configData,
	}
	return feature
}

func createAndSaveFeatureRule(featureIds []string, rule *re.Rule, applicationType string) *rfc.FeatureRule {
	featureRule := createFeatureRule(featureIds, rule, applicationType)
	setFeatureRule(featureRule)
	return featureRule
}

func createFeatureRule(featureIds []string, rule *re.Rule, applicationType string) *rfc.FeatureRule {
	id := uuid.New().String()
	configData := map[string]string{}
	configData[fmt.Sprintf("%s-key", id)] = fmt.Sprintf("%s-value", id)

	featureRule := &rfc.FeatureRule{
		Id:              id,
		Name:            fmt.Sprintf("%s-name", id),
		ApplicationType: applicationType,
		FeatureIds:      featureIds,
		Rule:            rule,
	}
	return featureRule
}

func createRule(condition *re.Condition) *re.Rule {
	rule := &re.Rule{
		Condition: condition,
	}
	return rule
}

func createPercentRangeRule() *re.Rule {
	condition := CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "estbMacAddress"), re.StandardOperationRange, "50-100")
	return createRule(condition)
}

func createAndSaveFeatureRules(features map[string]*rfc.Feature) map[string]*rfc.FeatureRule {
	stbFeatureIdList := []string{features["stb"].ID}
	stbFeatureRule := createFeatureRule(stbFeatureIdList, createRule(CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "model"), re.StandardOperationIs, "X1-1")), "stb")
	setFeatureRule(stbFeatureRule)
	RdkFeatureIdList := []string{features["rdkcloud"].ID}
	RdkFeatureRule := createFeatureRule(RdkFeatureIdList, createRule(CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "model"), re.StandardOperationIs, "X1-1")), "rdkcloud")
	setFeatureRule(RdkFeatureRule)
	featureRules := map[string]*rfc.FeatureRule{
		"stb":      stbFeatureRule,
		"rdkcloud": RdkFeatureRule,
	}
	return featureRules
}

func createAndSaveFeatures() map[string]*rfc.Feature {
	stbFeature := createFeature()
	stbFeature.ApplicationType = "stb"
	setFeature(stbFeature)

	RdkFeature := createFeature()
	RdkFeature.ApplicationType = "rdkcloud"
	setFeature(RdkFeature)

	features := map[string]*rfc.Feature{
		"stb":      stbFeature,
		"rdkcloud": RdkFeature,
	}
	return features
}
