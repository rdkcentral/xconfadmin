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
	"net/http"
	"net/http/httptest"
	"testing"

	admin_corefw "github.com/rdkcentral/xconfadmin/shared/firmware"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetMacRuleBeansWithoutVersionParam(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()

	macList := createAndSaveMacList()
	mrt := createAndSaveMacRuleTemplate(macList.ID)
	createAndSaveFirmwareMacRule(mrt.ID, &mrt.Rule, t)

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})

	url := fmt.Sprintf("/xconfAdminService/queries/rules/macs?%v", queryParams)
	r := httptest.NewRequest("GET", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	mrbs := unmarshallMacRuleBeans(t, rr)

	assert.NotEmpty(t, mrbs)
	assert.Empty(t, mrbs[0].MacList)

}

func TestGetMacRuleBeansWithVersionParams(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()

	macList := createAndSaveMacList()
	macRuleTemplate := createAndSaveMacRuleTemplate(macList.ID)
	createAndSaveFirmwareMacRule(macRuleTemplate.ID, &macRuleTemplate.Rule, t)

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})

	url := fmt.Sprintf("/xconfAdminService/firmwarerule?%v", queryParams)
	queryParams, _ = util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
		{"version", "3"},
	})

	url = fmt.Sprintf("/xconfAdminService/queries/rules/macs?%v", queryParams)
	r := httptest.NewRequest("GET", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	macRuleBeans := unmarshallMacRuleBeans(t, rr)

	assert.NotEmpty(t, macRuleBeans)
	assert.NotEmpty(t, macRuleBeans[0].MacList)
	assert.Contains(t, *macRuleBeans[0].MacList, macList.Data[0])
	assert.Contains(t, *macRuleBeans[0].MacList, macList.Data[1])

	url = fmt.Sprintf("/xconfAdminService/firmwarerule?%v", queryParams)
	queryParams, _ = util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
		{"version", "1"},
	})

	url = fmt.Sprintf("/xconfAdminService/queries/rules/macs?%v", queryParams)
	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	macRuleBeans = unmarshallMacRuleBeans(t, rr)

	assert.NotEmpty(t, macRuleBeans)
	assert.Empty(t, macRuleBeans[0].MacList)
}

func createAndSaveMacList() *shared.GenericNamespacedList {
	macList := shared.NewMacList()
	macList.ID = "TEST_MAC_LIST"
	macList.Data = []string{"AA:AA:AA:AA:AA:AA", "BB:BB:BB:BB:BB:BB"}
	SetOneInDao(ds.TABLE_GENERIC_NS_LIST, macList.ID, macList)
	return macList
}

func createAndSaveMacRuleTemplate(macListId string) *corefw.FirmwareRuleTemplate {
	macRule := estbfirmware.NewMacRule(macListId)
	mrt := admin_corefw.NewFirmwareRuleTemplate(corefw.MAC_RULE, macRule, []string{}, 1)
	SetOneInDao(ds.TABLE_FIRMWARE_RULE_TEMPLATE, mrt.ID, mrt)
	return mrt
}

func createAndSaveFirmwareMacRule(templateId string, macRule *re.Rule, t *testing.T) *corefw.FirmwareRule {
	ruleAction := corefw.NewApplicableAction(corefw.RuleActionClass, "")
	ruleAction.ActionType = corefw.RULE
	firmwareRule := corefw.NewFirmwareRule(uuid.New().String(), "TEST MAC RULE", templateId, macRule, ruleAction, true)

	firmwareRuleBytes, _ := json.Marshal(firmwareRule)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/firmwarerule?%v", queryParams)

	r := httptest.NewRequest("POST", url, bytes.NewReader(firmwareRuleBytes))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)

	return firmwareRule
}

func unmarshallMacRuleBeans(t *testing.T, rr *httptest.ResponseRecorder) []estbfirmware.MacRuleBeanResponse {
	var macRuleBeans []estbfirmware.MacRuleBeanResponse
	err := json.Unmarshal(rr.Body.Bytes(), &macRuleBeans)
	assert.NoError(t, err)
	return macRuleBeans
}
