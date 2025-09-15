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
