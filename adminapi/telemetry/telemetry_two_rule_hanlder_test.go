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
package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	core "github.com/rdkcentral/xconfadmin/shared"

	"github.com/rdkcentral/xconfadmin/common"

	"github.com/rdkcentral/xconfadmin/util"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateTelemetryTwoNoopRule(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	telemetryTwoRule := createTelemetryTwoRule(true, []string{})

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/v2/rule?%v", queryParams)

	rBytes, _ := json.Marshal(telemetryTwoRule)
	r := httptest.NewRequest("POST", url, bytes.NewReader(rBytes))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)

	profile := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profile.ID, profile)
}

func TestTelemetryTwoRuleNotCreateInNoOpValidationFails(t *testing.T) {
	tests := []struct {
		name         string
		noOp         bool
		profiles     []string
		expectedCode int
		errMsg       string
	}{
		{
			name:         "NoOp telemetry 2 rule with non empty profiles",
			noOp:         true,
			profiles:     []string{createTelemetryTwoProfile().ID},
			expectedCode: http.StatusBadRequest,
			errMsg:       "NoOp rule: profiles should be empty",
		},
		{
			name:         "Telemetry 2 rule with empty profiles",
			noOp:         false,
			profiles:     []string{},
			expectedCode: http.StatusBadRequest,
			errMsg:       "Profiles are not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteAllEntities()
			defer DeleteAllEntities()

			telemetryTwoRule := createTelemetryTwoRule(tt.noOp, tt.profiles)

			queryParams, _ := util.GetURLQueryParameterString([][]string{
				{"applicationType", "stb"},
			})
			url := fmt.Sprintf("/xconfAdminService/telemetry/v2/rule?%v", queryParams)

			rBytes, _ := json.Marshal(telemetryTwoRule)
			r := httptest.NewRequest("POST", url, bytes.NewReader(rBytes))
			rr := ExecuteRequest(r, router)
			assert.Equal(t, tt.expectedCode, rr.Code)

			var err common.XconfError
			json.Unmarshal(rr.Body.Bytes(), &err)
			assert.Equal(t, tt.errMsg, err.Message)

			savedTelemetryRule, _ := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_RULES, telemetryTwoRule.ID)
			assert.Nil(t, savedTelemetryRule)
		})
	}
}
func createRule(condition *re.Condition) *re.Rule {
	rule := &re.Rule{
		Condition: condition,
	}
	return rule
}

func CreateCondition(freeArg re.FreeArg, operation string, fixedArgValue string) *re.Condition {
	return re.NewCondition(&freeArg, operation, re.NewFixedArg(fixedArgValue))
}
func createTelemetryTwoRule(noOp bool, profiles []string) *xwlogupload.TelemetryTwoRule {
	telemetryRule := &xwlogupload.TelemetryTwoRule{}
	telemetryRule.ID = uuid.NewString()
	telemetryRule.Name = "TestTelemetryTwoRule"
	telemetryRule.ApplicationType = core.STB
	telemetryRule.BoundTelemetryIDs = profiles
	telemetryRule.NoOp = noOp
	telemetryRule.Rule = *createRule(CreateCondition(*estbfirmware.RuleFactoryVERSION, re.StandardOperationIs, "TEST_FIRMWARE_VERSION"))
	return telemetryRule
}

// Additional tests for telemetry_v2_rule_handler.go

func TestGetTelemetryTwoRulesAllExport_EmptyAndHeader(t *testing.T) {
	DeleteAllEntities()
	r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/v2/rule?applicationType=stb", nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "[]")
	// create one rule to test export header path
	prof := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, prof.ID, prof)
	rule := createTelemetryTwoRule(false, []string{prof.ID})
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, rule.ID, rule)
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/v2/rule?applicationType=stb&export=true", nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	cd := rr.Header().Get("Content-Disposition")
	assert.NotEmpty(t, cd)
}

func TestGetTelemetryTwoRuleById_SuccessExportAndNotFound(t *testing.T) {
	DeleteAllEntities()
	prof := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, prof.ID, prof)
	rule := createTelemetryTwoRule(false, []string{prof.ID})
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, rule.ID, rule)
	// success normal
	url := fmt.Sprintf("/xconfAdminService/telemetry/v2/rule/%s?applicationType=stb", rule.ID)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	// export
	url = fmt.Sprintf("/xconfAdminService/telemetry/v2/rule/%s?applicationType=stb&export=true", rule.ID)
	r = httptest.NewRequest(http.MethodGet, url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("Content-Disposition"))
	// not found
	url = fmt.Sprintf("/xconfAdminService/telemetry/v2/rule/%s?applicationType=stb", uuid.NewString())
	r = httptest.NewRequest(http.MethodGet, url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeleteOneTelemetryTwoRuleHandler_SuccessAndNotFound(t *testing.T) {
	DeleteAllEntities()
	prof := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, prof.ID, prof)
	rule := createTelemetryTwoRule(false, []string{prof.ID})
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, rule.ID, rule)
	url := fmt.Sprintf("/xconfAdminService/telemetry/v2/rule/%s?applicationType=stb", rule.ID)
	r := httptest.NewRequest(http.MethodDelete, url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusNoContent, rr.Code)
	// not found
	url = fmt.Sprintf("/xconfAdminService/telemetry/v2/rule/%s?applicationType=stb", uuid.NewString())
	r = httptest.NewRequest(http.MethodDelete, url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateTelemetryTwoRulesPackageHandler_Mixed(t *testing.T) {
	DeleteAllEntities()
	prof := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, prof.ID, prof)
	valid := createTelemetryTwoRule(false, []string{prof.ID})
	invalid := createTelemetryTwoRule(false, []string{}) // no profiles -> validation failure
	entities := []*xwlogupload.TelemetryTwoRule{valid, invalid}
	b, _ := json.Marshal(entities)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/rule/entities?applicationType=stb", bytes.NewReader(b))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, bytes.Contains(rr.Body.Bytes(), []byte(valid.ID)))
}

func TestUpdateTelemetryTwoRuleHandler_SuccessConflict(t *testing.T) {
	DeleteAllEntities()
	prof := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, prof.ID, prof)
	rule := createTelemetryTwoRule(false, []string{prof.ID})
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, rule.ID, rule)
	rule.Name = "UpdatedName"
	b, _ := json.Marshal(rule)
	r := httptest.NewRequest(http.MethodPut, "/xconfAdminService/telemetry/v2/rule?applicationType=stb", bytes.NewReader(b))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	// mismatch application type -> internal server error from service (fmt error path)
	rule.ApplicationType = "wrong"
	b, _ = json.Marshal(rule)
	r = httptest.NewRequest(http.MethodPut, "/xconfAdminService/telemetry/v2/rule?applicationType=stb", bytes.NewReader(b))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUpdateTelemetryTwoRulesPackageHandler_Mixed(t *testing.T) {
	DeleteAllEntities()
	prof := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, prof.ID, prof)
	a := createTelemetryTwoRule(false, []string{prof.ID})
	bRule := createTelemetryTwoRule(false, []string{prof.ID})
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, a.ID, a)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, bRule.ID, bRule)
	a.Name = "AUpdated"             // valid
	bRule.ApplicationType = "wrong" // conflict
	entities := []*xwlogupload.TelemetryTwoRule{a, bRule}
	b, _ := json.Marshal(entities)
	r := httptest.NewRequest(http.MethodPut, "/xconfAdminService/telemetry/v2/rule/entities?applicationType=stb", bytes.NewReader(b))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, bytes.Contains(rr.Body.Bytes(), []byte(a.ID)))
}

func TestGetTelemetryTwoRulesFilteredWithPage_PagingAndInvalid(t *testing.T) {
	DeleteAllEntities()
	prof := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, prof.ID, prof)
	for i := 0; i < 12; i++ {
		rule := createTelemetryTwoRule(false, []string{prof.ID})
		rule.Name = fmt.Sprintf("Rule_%02d", i)
		ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, rule.ID, rule)
	}
	// page 2 size 5
	bodyMap := map[string]string{}
	b, _ := json.Marshal(bodyMap)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/rule/filtered?pageNumber=2&pageSize=5&applicationType=stb", bytes.NewReader(b))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Rule_")
	// invalid pageNumber
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/rule/filtered?pageNumber=Z&pageSize=5&applicationType=stb", bytes.NewReader(b))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// invalid pageSize
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/rule/filtered?pageNumber=1&pageSize=X&applicationType=stb", bytes.NewReader(b))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
