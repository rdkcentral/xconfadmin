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
	"strings"
	"testing"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	xhttp "github.com/rdkcentral/xconfadmin/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gotest.tools/assert"
)

func TestFirmwareConfigParametersAreReturned(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testConfigFile)

	parameters := map[string]string{}
	configKey := "bindingUrl"
	configValue := "http://test.url.com"
	parameters[configKey] = configValue
	firmwareConfig := CreateFirmwareConfig(defaultFirmwareVersion, defaultModelId, "http", "stb")
	firmwareConfig.Properties = parameters
	err := SetFirmwareConfig(firmwareConfig)
	assert.NilError(t, err)
	applicableAction := corefw.NewTemplateApplicableActionAndType(corefw.RuleActionClass, corefw.RULE_TEMPLATE, "")
	rt := CreateAndSaveFirmwareRuleTemplate("ENV_MODEL_RULE", CreateDefaultEnvModelRule(), applicableAction)
	assert.Assert(t, rt != nil)
	bean, err := createAndSaveUseAccountPercentageBean(firmwareConfig)
	assert.Assert(t, bean != nil)
	assert.NilError(t, err)

	context := CreateContext(defaultFirmwareVersion, defaultModelId, defaultEnvironmentId, defaultIpAddress, defaultMacAddress)
	expectedResponse := map[string]interface{}{
		"bindingUrl":               "http://test.url.com",
		"firmwareDownloadProtocol": "http",
		"firmwareFilename":         "FirmwareFilename",
		"firmwareVersion":          defaultFirmwareVersion,
		"rebootImmediately":        false,
	}

	taggingMockServer := SetupTaggingMockServerOkResponse(t, *server, fmt.Sprintf(URL_TAGS_MAC_ADDRESS, defaultMacAddress))
	defer taggingMockServer.Close()

	performPostSwuRequestAndValidateBody(t, server, router, map[string]string{}, context, expectedResponse)
}

func TestFirmwareConfigParametersCanNotBeOverriddenByDefinePropertiesRule(t *testing.T) {
	DeleteAllEntities()
	server, router := GetTestWebConfigServer(testConfigFile)

	parameters := map[string]string{}
	configKey := "bindingUrl"
	configValue := "http://test.url.com"
	parameters[configKey] = configValue

	definePropertiesModelId := "DEFINE_PROPERTIES_MODEL_ID"

	firmwareConfig := CreateFirmwareConfig(defaultFirmwareVersion, definePropertiesModelId, "http", "stb")
	firmwareConfig.Properties = parameters
	err := SetFirmwareConfig(firmwareConfig)
	assert.NilError(t, err)

	applicableAction := corefw.NewTemplateApplicableActionAndType(corefw.RuleActionClass, corefw.RULE_TEMPLATE, "")
	rt := CreateAndSaveFirmwareRuleTemplate("ENV_MODEL_RULE", CreateDefaultEnvModelRule(), applicableAction)
	assert.Assert(t, rt != nil)
	percentageBean := CreatePercentageBean("test percentage bean", defaultEnvironmentId, definePropertiesModelId, "", "", defaultFirmwareVersion, "stb")
	percentageBean.LastKnownGood = firmwareConfig.ID
	percentageBean.FirmwareVersions = append(percentageBean.FirmwareVersions, firmwareConfig.FirmwareVersion)
	err = SavePercentageBean(percentageBean)
	assert.NilError(t, err)

	defineProperties := map[string]string{}
	defineProperties[configKey] = "CHANGED VALUE BY DEFINE PROPERTY RULE"
	defineProperties["definePropertyKey"] = "definePropertyValue"
	modelRule := CreateRule("", *coreef.RuleFactoryMODEL, re.StandardOperationIs, definePropertiesModelId)
	definePropertiesApplicableAction := corefw.NewApplicableActionAndType(corefw.DefinePropertiesActionClass, corefw.DEFINE_PROPERTIES_TEMPLATE, "")
	definePropertiesApplicableAction.Properties = defineProperties

	definePropertiesTemplateAction := corefw.NewTemplateApplicableActionAndType(corefw.DefinePropertiesTemplateActionClass, corefw.DEFINE_PROPERTIES_TEMPLATE, "")
	definePropertiesTemplateAction.Properties = buildDefinePropertyTemplateAction(defineProperties, false)
	definePropertiesTemplate := CreateAndSaveFirmwareRuleTemplate("OVERRIDE_FIRMWARE_CONFIG_PARAMETERS", modelRule, definePropertiesTemplateAction)

	fr := CreateAndSaveFirmwareRule(uuid.New().String(), definePropertiesTemplate.ID, "stb", definePropertiesApplicableAction, &definePropertiesTemplate.Rule)
	assert.Assert(t, fr != nil)
	context := CreateContext(defaultFirmwareVersion, definePropertiesModelId, defaultEnvironmentId, defaultIpAddress, defaultMacAddress)
	expectedResponse := map[string]interface{}{
		"bindingUrl":               "http://test.url.com",
		"firmwareDownloadProtocol": "http",
		"firmwareFilename":         "FirmwareFilename",
		"firmwareVersion":          defaultFirmwareVersion,
		"rebootImmediately":        false,
		"definePropertyKey":        "definePropertyValue",
	}

	taggingMockServer := SetupTaggingMockServerOkResponse(t, *server, fmt.Sprintf(URL_TAGS_MAC_ADDRESS, defaultMacAddress))
	defer taggingMockServer.Close()

	performPostSwuRequestAndValidateBody(t, server, router, map[string]string{}, context, expectedResponse)
}

func createAndSaveUseAccountPercentageBean(lkgConfig *coreef.FirmwareConfig) (*coreef.PercentageBean, error) {
	useAccountBean := CreatePercentageBean("useAccountName", defaultEnvironmentId, defaultModelId, "", "", defaultFirmwareVersion, "stb")
	useAccountBean.UseAccountIdPercentage = true
	useAccountBean.LastKnownGood = lkgConfig.ID
	firmwareVersions := useAccountBean.FirmwareVersions
	firmwareVersions = append(firmwareVersions, lkgConfig.FirmwareVersion)
	useAccountBean.FirmwareVersions = firmwareVersions
	err := SavePercentageBean(useAccountBean)
	return useAccountBean, err
}

func buildDefinePropertyTemplateAction(parameters map[string]string, requiredAll bool) map[string]corefw.PropertyValue {
	propertyValues := map[string]corefw.PropertyValue{}
	for k, v := range parameters {
		propertyValue := corefw.PropertyValue{
			Value:           v,
			Optional:        requiredAll,
			ValidationTypes: []corefw.ValidationType{"STRING"},
		}
		propertyValues[k] = propertyValue
	}
	return propertyValues
}

func SavePercentageBean(percentageBean *coreef.PercentageBean) error {
	firmwareRule := coreef.ConvertPercentageBeanToFirmwareRule(*percentageBean)
	return corefw.CreateFirmwareRuleOneDB(firmwareRule)
}

func performPostSwuRequestAndValidateBody(t *testing.T, server *xhttp.WebconfigServer, router *mux.Router, headers map[string]string, context *coreef.ConvertedContext, expectedResponse coreef.FirmwareConfigFacadeResponse) {
	satMockServer := SetupSatServiceMockServerOkResponse(t, *server)
	defer satMockServer.Close()

	url := postContext("/xconf/swu/stb", context)
	req, err := http.NewRequest("POST", url, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	actualResponse := map[string]interface{}{}
	err = json.Unmarshal(body, &actualResponse)
	assert.NilError(t, err)
	for k, v := range expectedResponse {
		switch v.(type) {
		case string:
			assert.Equal(t, v, actualResponse[k].(string))
		case bool:
			assert.Equal(t, v, actualResponse[k].(bool))
		// fail if not one of above types so we don't accidentally miss one
		default:
			assert.Equal(t, true, false)
		}
	}
}

func postContext(url string, context *coreef.ConvertedContext) string {
	contextMap := context.Context
	if len(contextMap) == 0 {
		return url
	}
	var sb strings.Builder
	for k, v := range contextMap {
		sb.Write([]byte(fmt.Sprintf("%s=%s&", k, v)))
	}
	queryParamString := sb.String()
	return fmt.Sprintf("%s?%s", url, queryParamString[0:len(queryParamString)-1])
}
