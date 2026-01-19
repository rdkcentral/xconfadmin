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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rdkcentral/xconfadmin/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"github.com/stretchr/testify/assert"
)

func createMockGlobalPercentageRule(applicationType string) *corefw.FirmwareRule {
	rule := &corefw.FirmwareRule{
		ID:              GetGlobalPercentageIdByApplication(applicationType),
		Name:            "GlobalPercentage_" + applicationType,
		Type:            coreef.GLOBAL_PERCENT,
		ApplicationType: applicationType,
		ApplicableAction: &corefw.ApplicableAction{
			Type: string(corefw.RULE_TEMPLATE),
		},
	}
	return rule
}

func TestGetCalculatedHashAndPercentHandler(t *testing.T) {
	tests := []struct {
		name            string
		macParam        string
		expectedStatus  int
		expectedHash    string
		expectedPercent string
	}{
		{
			name:            "Valid MAC address 1",
			macParam:        "00:23:ED:22:E3:BD",
			expectedStatus:  http.StatusOK,
			expectedHash:    "hashValue",
			expectedPercent: "percent",
		},
		{
			name:            "Valid MAC address 2",
			macParam:        "AA:BB:CC:DD:EE:FF",
			expectedStatus:  http.StatusOK,
			expectedHash:    "hashValue",
			expectedPercent: "percent",
		},
		{
			name:           "Missing MAC parameter",
			macParam:       "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid MAC format",
			macParam:       "00:23:ED:22:E3:D",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/xconfAdminService/percentfilter/calculator"
			if tt.macParam != "" {
				url += "?esbMac=" + tt.macParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			GetCalculatedHashAndPercentHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, rr.Body.String(), tt.expectedHash)
				assert.Contains(t, rr.Body.String(), tt.expectedPercent)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			}
		})
	}
}

func TestGetCalculatedHashAndPercent(t *testing.T) {
	tests := []struct {
		name           string
		macParam       string
		expectedStatus int
	}{
		{
			name:           "Valid MAC with esb_mac param",
			macParam:       "AA:BB:CC:DD:EE:11",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing esb_mac parameter",
			macParam:       "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid MAC format with esb_mac",
			macParam:       "INVALID",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/xconfAdminService/percentfilter/calculator2"
			if tt.macParam != "" {
				url += "?esb_mac=" + tt.macParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			GetCalculatedHashAndPercent(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "hashValue")
				assert.Contains(t, response, "percent")
			}
		})
	}
}

func TestUpdatePercentFilterGlobal(t *testing.T) {
	applicationType := "stb"

	t.Run("Create new global percentage", func(t *testing.T) {
		globalPercentage := coreef.NewGlobalPercentage()
		globalPercentage.Percentage = 50.0
		globalPercentage.ApplicationType = applicationType

		respEntity := UpdatePercentFilterGlobal(applicationType, globalPercentage)

		assert.NotNil(t, respEntity)
	})

	t.Run("Update existing global percentage", func(t *testing.T) {
		globalPercentage := coreef.NewGlobalPercentage()
		globalPercentage.Percentage = 30.0
		globalPercentage.ApplicationType = applicationType

		existingRule := createMockGlobalPercentageRule(applicationType)
		SetOneInDao(ds.TABLE_FIRMWARE_RULE, existingRule.ID, existingRule)

		respEntity := UpdatePercentFilterGlobal(applicationType, globalPercentage)

		assert.NotNil(t, respEntity)
	})
}

func TestUpdatePercentFilterGlobalHandler(t *testing.T) {
	applicationType := "stb"

	t.Run("Valid update request", func(t *testing.T) {
		globalPercentage := coreef.NewGlobalPercentage()
		globalPercentage.Percentage = 75.0
		globalPercentage.ApplicationType = applicationType

		body, _ := json.Marshal(globalPercentage)
		req := httptest.NewRequest(http.MethodPost, "/xconfAdminService/percentfilter/global", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)

		UpdatePercentFilterGlobalHandler(xw, req)

		assert.NotEqual(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Invalid JSON body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/xconfAdminService/percentfilter/global", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)

		UpdatePercentFilterGlobalHandler(xw, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Non-XResponseWriter cast error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/xconfAdminService/percentfilter/global", nil)
		rr := httptest.NewRecorder()

		UpdatePercentFilterGlobalHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestGetPercentFilterGlobal(t *testing.T) {
	applicationType := "stb"

	t.Run("Get existing global percentage", func(t *testing.T) {
		rule := createMockGlobalPercentageRule(applicationType)
		SetOneInDao(ds.TABLE_FIRMWARE_RULE, rule.ID, rule)

		result, err := GetPercentFilterGlobal(applicationType)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Get non-existing global percentage", func(t *testing.T) {
		result, err := GetPercentFilterGlobal("xhome")

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestGetPercentFilterGlobalHandler(t *testing.T) {
	applicationType := "stb"

	t.Run("Get global percentage without export", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/xconfAdminService/percentfilter/global?applicationType="+applicationType, nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)

		GetPercentFilterGlobalHandler(xw, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Content-Disposition"))
	})

	t.Run("Get global percentage with export", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/xconfAdminService/percentfilter/global?applicationType="+applicationType+"&export=true", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)

		GetPercentFilterGlobalHandler(xw, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Content-Disposition"))
		assert.Contains(t, rr.Header().Get("Content-Disposition"), common.ExportFileNames_PERCENT_FILTER)
	})
}

func TestGetGlobalPercentFilter(t *testing.T) {
	applicationType := "stb"

	t.Run("Get global percent filter VO with existing rule", func(t *testing.T) {
		rule := createMockGlobalPercentageRule(applicationType)
		SetOneInDao(ds.TABLE_FIRMWARE_RULE, rule.ID, rule)

		result, err := GetGlobalPercentFilter(applicationType)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.GlobalPercentage)
		assert.Equal(t, applicationType, result.GlobalPercentage.ApplicationType)
	})

	t.Run("Get global percent filter VO without existing rule", func(t *testing.T) {
		result, err := GetGlobalPercentFilter("xhome")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.GlobalPercentage)
	})
}

func TestGetGlobalPercentFilterHandler(t *testing.T) {
	applicationType := "stb"

	t.Run("Get global percent filter without export", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/xconfAdminService/percentfilter/globalPercent?applicationType="+applicationType, nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)

		GetGlobalPercentFilterHandler(xw, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Content-Disposition"))
	})

	t.Run("Get global percent filter with export", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/xconfAdminService/percentfilter/globalPercent?applicationType="+applicationType+"&export=true", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)

		GetGlobalPercentFilterHandler(xw, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Content-Disposition"))
		assert.Contains(t, rr.Header().Get("Content-Disposition"), common.ExportFileNames_GLOBAL_PERCENT)
	})
}

func TestGetGlobalPercentFilterAsRule(t *testing.T) {
	applicationType := "stb"

	t.Run("Get existing rule", func(t *testing.T) {
		rule := createMockGlobalPercentageRule(applicationType)
		SetOneInDao(ds.TABLE_FIRMWARE_RULE, rule.ID, rule)

		result, err := GetGlobalPercentFilterAsRule(applicationType)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, rule.ID, result.ID)
	})

	t.Run("Get non-existing rule", func(t *testing.T) {
		ClearMockDatabase()

		result, err := GetGlobalPercentFilterAsRule("xhome")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetGlobalPercentFilterAsRuleHandler(t *testing.T) {
	applicationType := "stb"

	t.Run("Get rule without export - existing rule", func(t *testing.T) {
		rule := createMockGlobalPercentageRule(applicationType)
		SetOneInDao(ds.TABLE_FIRMWARE_RULE, rule.ID, rule)

		req := httptest.NewRequest(http.MethodGet, "/xconfAdminService/percentfilter/globalPercentAsRule?applicationType="+applicationType, nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)

		GetGlobalPercentFilterAsRuleHandler(xw, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Content-Disposition"))

		var rules []*corefw.FirmwareRule
		err := json.Unmarshal(rr.Body.Bytes(), &rules)
		assert.NoError(t, err)
		assert.Len(t, rules, 1)
	})

	t.Run("Get rule with export - non-existing rule", func(t *testing.T) {
		ClearMockDatabase()

		req := httptest.NewRequest(http.MethodGet, "/xconfAdminService/percentfilter/globalPercentAsRule?applicationType=xhome&export=true", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)

		GetGlobalPercentFilterAsRuleHandler(xw, req)

		// Handler may return 200 with default rule or 400 if marshaling fails
		assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusBadRequest)
		if rr.Code == http.StatusOK {
			assert.NotEmpty(t, rr.Header().Get("Content-Disposition"))
			assert.Contains(t, rr.Header().Get("Content-Disposition"), common.ExportFileNames_GLOBAL_PERCENT_AS_RULE)
		}
	})

	t.Run("Get rule without export - non-existing rule", func(t *testing.T) {
		ClearMockDatabase()

		req := httptest.NewRequest(http.MethodGet, "/xconfAdminService/percentfilter/globalPercentAsRule?applicationType=sky", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)

		GetGlobalPercentFilterAsRuleHandler(xw, req)

		// Handler may return 200 with default rule or 400 if marshaling fails
		assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusBadRequest)
	})
}

func TestCalculateHashAndPercent(t *testing.T) {
	tests := []struct {
		name       string
		macAddress string
	}{
		{
			name:       "MAC address 1",
			macAddress: `"00:23:ED:22:E3:BD"`,
		},
		{
			name:       "MAC address 2",
			macAddress: `"AA:BB:CC:DD:EE:FF"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashCode, percent := calculateHashAndPercent(tt.macAddress)

			assert.NotZero(t, hashCode)
			assert.GreaterOrEqual(t, percent, 0.0)
			assert.LessOrEqual(t, percent, 100.0)
		})
	}
}
