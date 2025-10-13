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
	"net/http/httptest"
	"strings"

	"github.com/rdkcentral/xconfadmin/common"

	estb "github.com/rdkcentral/xconfwebconfig/dataapi/estbfirmware"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/http"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"github.com/google/uuid"
)

// code is based
// Java com.comcast.xconf.queries.controllers.BaseQueriesControllerTest
const (
	defaultModelId                   = "modelId"
	defaultEnvironmentId             = "environmentId"
	defaultEnvModelId                = "envModelId"
	defaultIpFilterId                = "ipFilterId"
	defaultTimeFilterId              = "timeFilterId"
	defaultRebootImmediatelyFilterId = "rebootImmediatelyFilterId"
	defaultFirmwareVersion           = "firmwareVersion"
	contextFirmwareVersion           = "contextFirmwareVersion"
	defaultIpRuleId                  = "ipRuleId"
	defaultMacRuleId                 = "macRuleId"
	defaultDownloadLocationFilterId  = "dowloadLocationFilterId"
	defaultIpListId                  = "ipListId"
	defaultMacListId                 = "macListId"
	defaultIpAddress                 = "1.1.1.1"
	defaultIpv6Address               = "::1"
	defaultMacAddress                = "11:11:11:11:11:11"
	defaultHttpLocation              = "httpLocation.com"
	defaultHttpFullUrlLocation       = "http://fullUrlLocation.com"
	defaultHttpsFullUrlLocation      = "https://fullUrlLocation.com"
	defaultFormulaId                 = "defaultFormulaObject"
	defaultFirmwareConfigId          = "firmwareConfigId"
	defaultPartnerId                 = "defaultpartnerid"
	defaultTimeZone                  = "Australia/Brisbane"
	defaultServiceAccountUri         = "defaultServiceAccountUri"
	defaultAccountId                 = "defaultAccountId"
	defaultFirmwareDownloadProtocol  = "http"
	defaultDeviceSettingName         = "deviceSettingsName"
	defaultLogUploadSettingName      = "logUploadSettingsName"

	API_VERSION = "2"
	//APPLICATION_XML_UTF8 = new MediaType(MediaType.APPLICATION_XML.getType(), MediaType.APPLICATION_XML.getSubtype(), Charsets.UTF_8)
	APPLICATION_TYPE_PARAM = "applicationType"
	WRONG_APPLICATION      = "wrongVersion"
)

func CreateGenericNamespacedList(name string, ttype string, data string) *shared.GenericNamespacedList {
	namespacedList := shared.NewGenericNamespacedList(name, ttype, strings.Split(data, ","))
	return namespacedList
}

func CreateCondition(freeArg re.FreeArg, operation string, fixedArgValue string) *re.Condition {
	return re.NewCondition(&freeArg, operation, re.NewFixedArg(fixedArgValue))
}

func CreateRule(relation string, freeArg re.FreeArg, operation string, fixedArgValue string) *re.Rule {
	rule := re.Rule{}
	rule.SetRelation(relation)
	rule.SetCondition(CreateCondition(freeArg, operation, fixedArgValue))
	return &rule
}

func CreateRuleKeyValue(key string, value string) *re.Rule {
	condition := CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, key), re.StandardOperationIs, value)
	return &re.Rule{
		Condition: condition,
	}
}

func CreateAndSaveFirmwareRule(id string, templateId string, applicationType string, action *corefw.ApplicableAction, rule *re.Rule) *corefw.FirmwareRule {
	firmwareRule := CreateFirmwareRule(id, templateId, applicationType, action, rule)
	corefw.CreateFirmwareRuleOneDB(firmwareRule)
	return firmwareRule
}

func CreateFirmwareRule(id string, templateId string, applicationType string, action *corefw.ApplicableAction, rule *re.Rule) *corefw.FirmwareRule {
	firmwareRule := &corefw.FirmwareRule{
		ID:               id,
		Name:             id,
		Active:           true,
		ApplicableAction: action,
		ApplicationType:  applicationType,
		Type:             templateId,
		Rule:             *rule,
	}
	return firmwareRule
}

// createRuleActionn return *corefw.RuleAction
// but due to FirmwaereRule and FirmwareRuleTemplate has only corefw.ApplicableAction
// OR TemplateApplicableAction
// so We have no change it as two methods
func CreateRuleAction(typ string, actiontyp corefw.ApplicableActionType, firmwareConfigId string) *corefw.ApplicableAction {
	ruleAction := corefw.NewApplicableActionAndType(typ, actiontyp, firmwareConfigId)
	//ruleAction.ApplicableAction = corefw.ApplicableAction{Type: ttype,}
	//ruleAction.ConfigId = firmwareConfigId
	//todo why the tuleAction has id
	//ruleAction.ID = uuid.New().String()
	return ruleAction
}

func CreateTemplateRuleAction(typ string, actiontyp corefw.ApplicableActionType, firmwareConfigId string) *corefw.TemplateApplicableAction {
	ruleAction := corefw.NewTemplateApplicableActionAndType(typ, actiontyp, firmwareConfigId)
	//ruleAction.ApplicableAction = corefw.ApplicableAction{Type: ttype,}
	//ruleAction.ConfigId = firmwareConfigId
	//todo why the tuleAction has id
	//ruleAction.ID = uuid.New().String()
	return ruleAction
}

func CreateDefaultEnvModelRule() *re.Rule {
	envModelRule := re.NewEmptyRule()
	envModelRule.AddCompoundPart(*CreateRule("", *coreef.RuleFactoryENV, re.StandardOperationIs, strings.ToUpper(defaultEnvironmentId)))
	envModelRule.AddCompoundPart(*CreateRule(re.RelationAnd, *coreef.RuleFactoryMODEL, re.StandardOperationIs, strings.ToUpper(defaultModelId)))
	return envModelRule
}

func CreateEnvModelRule(envId string, modelId string, namespacedListId string) *re.Rule {
	envModelRule := re.NewEmptyRule()
	envModelRule.AddCompoundPart(*CreateRule("", *coreef.RuleFactoryENV, re.StandardOperationIs, envId))
	envModelRule.AddCompoundPart(*CreateRule(re.RelationAnd, *coreef.RuleFactoryMODEL, re.StandardOperationIs, modelId))
	envModelRule.AddCompoundPart(*CreateRule(re.RelationAnd, *coreef.RuleFactoryMAC, *&coreef.RuleFactoryIN_LIST, namespacedListId))

	return envModelRule
}

func CreateExistsRule(tagName string) *re.Rule {
	condition := CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeAny, tagName), re.StandardOperationExists, "")
	rule := &re.Rule{
		Condition: condition,
	}
	return rule
}

func CreateAccountPartnerObject(partnerId string) http.AccountServiceDevices {
	accountObject := http.AccountServiceDevices{
		Id: uuid.New().String(),
		DeviceData: http.DeviceData{
			Partner:           partnerId,
			ServiceAccountUri: defaultServiceAccountUri,
		},
	}
	return accountObject
}

func CreateODPPartnerObject() http.DeviceServiceObject {
	odpObject := http.DeviceServiceObject{
		Status: 200,
		DeviceServiceData: &http.DeviceServiceData{
			AccountId: defaultServiceAccountUri,
		}}
	return odpObject
}

func CreateODPPartnerObjectWithPartnerAndTimezone() http.DeviceServiceObject {
	odpObject := http.DeviceServiceObject{
		Status: 200,
		DeviceServiceData: &http.DeviceServiceData{
			AccountId: defaultServiceAccountUri,
			PartnerId: defaultPartnerId,
			TimeZone:  defaultTimeZone,
		}}
	return odpObject
}

func CreateODPPartnerObjectWithPartnerAndTimezoneInvalid() http.DeviceServiceObject {
	odpObject := http.DeviceServiceObject{
		Status: 200,
		DeviceServiceData: &http.DeviceServiceData{
			AccountId: defaultServiceAccountUri,
			PartnerId: defaultPartnerId,
			TimeZone:  "InvalidTimeZone",
		}}
	return odpObject
}

func CreateAndSaveModel(id string) *shared.Model {
	model := shared.NewModel(id, "ModelDescription")
	//jsonData, _ := json.Marshal(model)

	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_MODEL, model.ID, model)
	if err != nil {
		return nil
	}

	return model
}

func CreateAndSaveEnvironment(id string) *shared.Environment {
	env := shared.NewEnvironment(id, "ENV_MODEL_RULE_ENVIRONMENT_ID")
	//jsonData, _ := json.Marshal(env)

	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_ENVIRONMENT, env.ID, env)
	if err != nil {
		return nil
	}

	return env
}

func CreateAndSaveGenericNamespacedList(name string, ttype string, data string) *shared.GenericNamespacedList {
	namespacedList := CreateGenericNamespacedList(name, ttype, data)
	//jsonData, _ := json.Marshal(namespacedList)

	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, namespacedList.ID, namespacedList)
	if err != nil {
		return nil
	}
	return namespacedList
}

func CreateFirmwareConfigfw(firmwareVersion string, modelId string, firmwareDownloadProtocol string, applicationType string) *coreef.FirmwareConfig {
	firmwareConfig := coreef.NewEmptyFirmwareConfig()
	firmwareConfig.ID = uuid.New().String()
	firmwareConfig.Description = "FirmwareDescription"
	firmwareConfig.FirmwareFilename = "FirmwareFilename"
	firmwareConfig.FirmwareVersion = firmwareVersion
	firmwareConfig.FirmwareDownloadProtocol = firmwareDownloadProtocol
	firmwareConfig.ApplicationType = applicationType
	supportedModels := make([]string, 1)
	model := CreateAndSaveModel(strings.ToUpper(modelId))
	supportedModels[0] = model.ID
	firmwareConfig.SupportedModelIds = supportedModels
	return firmwareConfig
}

func CreateAndSaveFirmwareConfig(firmwareVersion string, modelId string, firmwareDownloadProtocol string, applicationType string) *coreef.FirmwareConfig {
	firmwareConfig := CreateFirmwareConfigfw(firmwareVersion, modelId, firmwareDownloadProtocol, applicationType)
	err := SetFirmwareConfig(firmwareConfig)
	if err != nil {
		return nil
	}
	return firmwareConfig
}

func SetFirmwareConfig(firmwareConfig *coreef.FirmwareConfig) error {
	err := coreef.CreateFirmwareConfigOneDB(firmwareConfig)
	if err != nil {
		return err
	}
	return nil
}

func CreatePercentageBeanPB(name string, envId string, modelId string, whitelistId string, whitelistData string, firmwareVersion string, applicationType string) *coreef.PercentageBean {
	var whitelist string
	if whitelistId != "" {
		whitelist = CreateAndSaveGenericNamespacedList(whitelistId, "IP_LIST", whitelistData).ID
	}
	firmwareConfig := CreateAndSaveFirmwareConfig(firmwareVersion, modelId, "http", applicationType)
	configEntry := corefw.NewConfigEntry(firmwareConfig.ID, 0.0, 66.0)
	percentageBean := &coreef.PercentageBean{
		ID:                    uuid.New().String(),
		Name:                  name,
		Whitelist:             whitelist,
		Active:                true,
		Environment:           CreateAndSaveEnvironment(envId).ID,
		Model:                 CreateAndSaveModel(modelId).ID,
		FirmwareCheckRequired: true,
		ApplicationType:       applicationType,
		FirmwareVersions:      []string{firmwareConfig.FirmwareVersion},
		LastKnownGood:         firmwareConfig.ID,
		Distributions:         []*corefw.ConfigEntry{configEntry},
		IntermediateVersion:   firmwareConfig.ID,
	}
	return percentageBean
}

func CreateAndSaveFirmwareRuleTemplate(id string, rule *re.Rule, applicableAction *corefw.TemplateApplicableAction) *corefw.FirmwareRuleTemplate {
	template := CreateFirmwareRuleTemplate(id, rule, applicableAction)
	if err := corefw.CreateFirmwareRuleTemplateOneDB(template); err != nil {
		panic(err)
	}
	return template
}

func CreateFirmwareRuleTemplate(id string, rule *re.Rule, applicableAction *corefw.TemplateApplicableAction) *corefw.FirmwareRuleTemplate {
	template := corefw.NewEmptyFirmwareRuleTemplate()
	template.ID = id
	template.Rule = *rule
	template.ApplicableAction = applicableAction
	return template
}

func CreateAndSaveEnvModelFirmwareRule(name string, firmwareConfigId string, envId string, modelId string, macListId string) *corefw.FirmwareRule {
	envModelRule := corefw.NewEmptyFirmwareRule()
	envModelRule.ID = uuid.New().String()
	envModelRule.Name = name
	ruleAct := CreateRuleAction(corefw.RuleActionClass, corefw.RULE, firmwareConfigId)
	envModelRule.ApplicableAction = ruleAct
	envModelRule.Type = "ENV_MODEL_RULE"
	envModelRule.Rule = *CreateEnvModelRule(envId, modelId, macListId)
	//jsonData, _ := json.Marshal(envModelRule)
	err := corefw.CreateFirmwareRuleOneDB(envModelRule)
	if err != nil {
		return nil
	}
	return envModelRule
}

func CreateIpAddressGroupExtended(stringIpAddresses []string) *shared.IpAddressGroup {
	return CreateIpAddressGroupExtendedWithName(uuid.New().String(), stringIpAddresses)
}

func CreateIpAddressGroupExtendedWithName(name string, stringIpAddresses []string) *shared.IpAddressGroup {
	return shared.NewIpAddressGroupWithAddrStrings(name, name, stringIpAddresses)
}

func CreateAndSavePercentFilter(
	envModelRuleName string,
	percentage float64,
	lastKnownGood string,
	intermediateVersion string,
	envModelPercent float64,
	firmwareVersions []string,
	isActive bool,
	isFirmwareCheckRequired bool,
	rebootImmediately bool,
	applicationType string) *coreef.PercentFilterValue {

	percentFilter := coreef.NewEmptyPercentFilterValue()

	whitelist := CreateIpAddressGroupExtended([]string{"127.1.1.1", "127.1.1.2"})

	envModelPercentage := coreef.NewEnvModelPercentage()
	envModelPercentage.Whitelist = whitelist
	envModelPercentage.LastKnownGood = lastKnownGood
	envModelPercentage.IntermediateVersion = intermediateVersion
	envModelPercentage.FirmwareVersions = firmwareVersions
	envModelPercentage.Percentage = float32(envModelPercent)
	envModelPercentage.Active = isActive
	envModelPercentage.FirmwareCheckRequired = isFirmwareCheckRequired
	envModelPercentage.RebootImmediately = rebootImmediately

	percentFilter.Percentage = float32(percentage)
	percentFilter.Whitelist = whitelist
	mapEnvModes := make(map[string]coreef.EnvModelPercentage)
	mapEnvModes[envModelRuleName] = *envModelPercentage
	percentFilter.EnvModelPercentages = mapEnvModes

	percentFilterService := estb.NewPercentFilterService()
	percentFilterService.Save(percentFilter, applicationType)

	return percentFilter
}

func CreateContext(firmwareVersion string, modelId string, environmentId string, ipAddress string, eStbMac string) *coreef.ConvertedContext {
	contextMap := map[string]string{
		"firmwareVersion": firmwareVersion,
		"model":           modelId,
		"env":             environmentId,
		"ipAddress":       ipAddress,
		"eStbMac":         eStbMac,
	}
	context := coreef.GetContextConverted(contextMap)
	return context
}

func unmarshalXconfError(b []byte) *common.XconfError {
	var xconfError *common.XconfError
	err := json.Unmarshal(b, &xconfError)
	if err != nil {
		(fmt.Errorf("error unmarshaling xconf error"))
	}
	return xconfError
}

func SendRequest(url string, method string, entity interface{}) *httptest.ResponseRecorder {
	entityJson, _ := json.Marshal(entity)
	r := httptest.NewRequest(method, url, bytes.NewReader(entityJson))
	rr := ExecuteRequest(r, router)
	return rr
}
