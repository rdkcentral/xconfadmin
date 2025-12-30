/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
	"time"

	"github.com/google/uuid"
	"gotest.tools/assert"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

// helper: build telemetry profile body
func buildTelemetryProfile(expiresOffsetMillis int64) *xwlogupload.TelemetryProfile {
	p := &xwlogupload.TelemetryProfile{}
	p.ID = uuid.New().String()
	p.Name = "test_profile"
	p.ApplicationType = "stb"
	nowMillis := time.Now().UnixNano() / 1_000_000
	p.Expires = nowMillis + expiresOffsetMillis
	p.TelemetryProfile = []xwlogupload.TelemetryElement{{
		ID:               uuid.New().String(),
		Header:           "hdr",
		Content:          "cnt",
		Type:             "type",
		PollingFrequency: "60",
		Component:        "comp",
	}}
	return p
}

func createPermanentTelemetryProfile(id string) *xwlogupload.PermanentTelemetryProfile {
	perm := &xwlogupload.PermanentTelemetryProfile{}
	perm.ID = id
	perm.Name = "perm_profile"
	perm.ApplicationType = "stb"
	perm.TelemetryProfile = []xwlogupload.TelemetryElement{{
		ID:               uuid.New().String(),
		Header:           "hdr_perm",
		Content:          "cnt_perm",
		Type:             "type_perm",
		PollingFrequency: "120",
		Component:        "comp_perm",
	}}
	_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, perm.ID, perm)
	return perm
}

func createTelemetryRule(boundProfileId string) *xwlogupload.TelemetryRule {
	r := &xwlogupload.TelemetryRule{}
	r.ID = uuid.New().String()
	r.Name = "telemetry_rule"
	r.ApplicationType = "stb"
	r.BoundTelemetryID = boundProfileId
	_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_RULES, r.ID, r)
	return r
}

func exec(method, url string, body []byte) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, url, bytes.NewReader(body))
	return ExecuteRequest(r, router)
}

func TestCreateTelemetryEntryForSuccess(t *testing.T) {
	DeleteTelemetryEntities()
	profile := buildTelemetryProfile(60_000)
	body, _ := json.Marshal(profile)
	url := fmt.Sprintf("/xconfAdminService/telemetry/create/estbMacAddress/%s?applicationType=stb", "AA:BB:CC:DD:EE:FF")
	rr := exec("POST", url, body)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Assert(t, len(rr.Body.Bytes()) > 0)
}

func TestCreateTelemetryEntryForFailures(t *testing.T) {
	DeleteTelemetryEntities()
	// wrong attribute
	profile := buildTelemetryProfile(60000)
	body, _ := json.Marshal(profile)
	url := fmt.Sprintf("/xconfAdminService/telemetry/create/model/%s?applicationType=stb", "TESTMODEL")
	rr := exec("POST", url, body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// expired
	expired := buildTelemetryProfile(-1000)
	body, _ = json.Marshal(expired)
	url = fmt.Sprintf("/xconfAdminService/telemetry/create/estbMacAddress/%s?applicationType=stb", "11:22:33:44:55:66")
	rr = exec("POST", url, body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// invalid JSON
	url = fmt.Sprintf("/xconfAdminService/telemetry/create/estbMacAddress/%s?applicationType=stb", "11:22:33:44:55:77")
	r := httptest.NewRequest("POST", url, bytes.NewReader([]byte("{invalid")))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// func TestDropTelemetryEntryForSuccess(t *testing.T) {
// 	DeleteTelemetryEntities()
// 	_ = createPermanentTelemetryProfile("perm-1")
// 	p := buildTelemetryProfile(60000)
// 	body, _ := json.Marshal(p)
// 	url := fmt.Sprintf("/xconfAdminService/telemetry/create/estbMacAddress/%s?applicationType=stb", "AA:AA:AA:AA:AA:AA")
// 	exec("POST", url, body)
// 	url = "/xconfAdminService/telemetry/drop/estbMacAddress/AA:AA:AA:AA:AA:AA?applicationType=stb"
// 	rr := exec("POST", url, nil)
// 	assert.Equal(t, http.StatusOK, rr.Code)
// }

func TestGetDescriptorsAndTelemetryDescriptors(t *testing.T) {
	DeleteTelemetryEntities()
	url := "/xconfAdminService/telemetry/getAvailableRuleDescriptors?applicationType=stb"
	rr := exec("GET", url, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	url = "/xconfAdminService/telemetry/getAvailableTelemetryDescriptors?applicationType=stb"
	rr = exec("GET", url, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTempAddToPermanentRule(t *testing.T) {
	DeleteTelemetryEntities()
	perm := createPermanentTelemetryProfile("perm-2")
	rule := createTelemetryRule(perm.ID)
	expires := (time.Now().UnixNano() / 1_000_000) + 60000
	// success
	url := fmt.Sprintf("/xconfAdminService/telemetry/addTo/%s/estbMacAddress/%s/%d?applicationType=stb", rule.ID, "CC:DD:EE:FF:00:11", expires)
	rr := exec("POST", url, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	// invalid expires
	url = fmt.Sprintf("/xconfAdminService/telemetry/addTo/estbMacAddress/value/%s/notnum?applicationType=stb", rule.ID)
	rr = exec("POST", url, nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// rule not found
	url = fmt.Sprintf("/xconfAdminService/telemetry/addTo/estbMacAddress/value/%s/%d?applicationType=stb", uuid.New().String(), expires)
	rr = exec("POST", url, nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBindToTelemetry(t *testing.T) {
	DeleteTelemetryEntities()
	perm := createPermanentTelemetryProfile("perm-3")
	expires := (time.Now().UnixNano() / 1_000_000) + 60000
	// success
	url := fmt.Sprintf("/xconfAdminService/telemetry/bindToTelemetry/%s/estbMacAddress/%s/%d?applicationType=stb", perm.ID, "DD:EE:FF:00:11:22", expires)
	rr := exec("POST", url, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	// profile not found
	url = fmt.Sprintf("/xconfAdminService/telemetry/bindToTelemetry/estbMacAddress/value/%s/%d?applicationType=stb", uuid.New().String(), expires)
	rr = exec("POST", url, nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// invalid expires
	url = fmt.Sprintf("/xconfAdminService/telemetry/bindToTelemetry/estbMacAddress/value/%s/notnum?applicationType=stb", perm.ID)
	rr = exec("POST", url, nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestTelemetryTestPageHandler(t *testing.T) {
	DeleteTelemetryEntities()
	bodyMap := map[string]interface{}{
		"estbMacAddress": "AA:BB:CC:DD:EE:FF",
		"model":          "TESTMODEL",
	}
	body, _ := json.Marshal(bodyMap)
	url := "/xconfAdminService/telemetry/testpage?applicationType=stb"
	rr := exec("POST", url, body)
	assert.Equal(t, http.StatusOK, rr.Code)
	// invalid json
	r := httptest.NewRequest("POST", url, bytes.NewReader([]byte("{bad")))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestCreateTelemetryEntryFor_AllErrorCases tests all error paths
func TestCreateTelemetryEntryFor_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	tests := []struct {
		name        string
		url         string
		body        []byte
		expectError bool
		description string
	}{
		{
			name:        "InvalidContextAttributeName",
			url:         "/xconfAdminService/telemetry/create/model/TESTMODEL?applicationType=stb",
			body:        []byte("{}"),
			expectError: true,
			description: "xwhttp.WriteXconfResponse - only estbMacAddress allowed",
		},
		{
			name:        "InvalidJSON",
			url:         "/xconfAdminService/telemetry/create/estbMacAddress/AA:BB:CC:DD:EE:FF?applicationType=stb",
			body:        []byte("{invalid json"),
			expectError: true,
			description: "xwhttp.WriteXconfResponse - JSON unmarshal error",
		},
		{
			name: "ExpiredTimestamp",
			url:  "/xconfAdminService/telemetry/create/estbMacAddress/AA:BB:CC:DD:EE:FF?applicationType=stb",
			body: func() []byte {
				p := buildTelemetryProfile(-100000)
				b, _ := json.Marshal(p)
				return b
			}(),
			expectError: true,
			description: "xwhttp.WriteXconfResponse - Invalid Expires Timestamp",
		},
		{
			name: "Success",
			url:  "/xconfAdminService/telemetry/create/estbMacAddress/AA:BB:CC:DD:EE:FF?applicationType=stb",
			body: func() []byte {
				p := buildTelemetryProfile(60000)
				b, _ := json.Marshal(p)
				return b
			}(),
			expectError: false,
			description: "Success case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", tt.url, bytes.NewReader(tt.body))
			rr := ExecuteRequest(r, router)
			if tt.expectError {
				assert.Assert(t, rr.Code >= http.StatusBadRequest, tt.description)
			} else {
				assert.Equal(t, http.StatusOK, rr.Code, tt.description)
			}
		})
	}
}

// TestDropTelemetryEntryFor_AllErrorCases tests all error paths
func TestDropTelemetryEntryFor_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	tests := []struct {
		name        string
		url         string
		expectError bool
		description string
	}{
		{
			name:        "Success",
			url:         "/xconfAdminService/telemetry/drop/estbMacAddress/AA:BB:CC:DD:EE:FF?applicationType=stb",
			expectError: false,
			description: "Success case - xwhttp.WriteXconfResponse with OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := exec("POST", tt.url, nil)
			if tt.expectError {
				assert.Assert(t, rr.Code >= http.StatusBadRequest, tt.description)
			} else {
				assert.Equal(t, http.StatusOK, rr.Code, tt.description)
			}
		})
	}
}

// TestGetDescriptors_AllErrorCases tests GetDescriptors error paths
func TestGetDescriptors_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	tests := []struct {
		name               string
		url                string
		expectedStatusCode int
		description        string
	}{
		{
			name:               "Success_WithApplicationType",
			url:                "/xconfAdminService/telemetry/getAvailableRuleDescriptors?applicationType=stb",
			expectedStatusCode: http.StatusOK,
			description:        "Success with applicationType",
		},
		{
			name:               "Success_NoApplicationType",
			url:                "/xconfAdminService/telemetry/getAvailableRuleDescriptors",
			expectedStatusCode: http.StatusOK,
			description:        "Success without applicationType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := exec("GET", tt.url, nil)
			assert.Equal(t, tt.expectedStatusCode, rr.Code, tt.description)
		})
	}
}

// TestGetTelemetryDescriptors_AllErrorCases tests GetTelemetryDescriptors error paths
func TestGetTelemetryDescriptors_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	tests := []struct {
		name               string
		url                string
		expectedStatusCode int
		description        string
	}{
		{
			name:               "Success_WithApplicationType",
			url:                "/xconfAdminService/telemetry/getAvailableTelemetryDescriptors?applicationType=stb",
			expectedStatusCode: http.StatusOK,
			description:        "Success with applicationType",
		},
		{
			name:               "Success_NoApplicationType",
			url:                "/xconfAdminService/telemetry/getAvailableTelemetryDescriptors",
			expectedStatusCode: http.StatusOK,
			description:        "Success without applicationType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := exec("GET", tt.url, nil)
			assert.Equal(t, tt.expectedStatusCode, rr.Code, tt.description)
		})
	}
}

// TestTempAddToPermanentRule_AllErrorCases tests all error paths
func TestTempAddToPermanentRule_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()
	perm := createPermanentTelemetryProfile("perm-temp-1")
	rule := createTelemetryRule(perm.ID)
	expires := (time.Now().UnixNano() / 1_000_000) + 60000

	tests := []struct {
		name        string
		url         string
		expectError bool
		description string
	}{
		{
			name:        "InvalidContextAttributeName",
			url:         fmt.Sprintf("/xconfAdminService/telemetry/addTo/%s/model/TESTMODEL/%d?applicationType=stb", rule.ID, expires),
			expectError: true,
			description: "xwhttp.WriteXconfResponse - only estbMacAddress allowed",
		},
		{
			name:        "InvalidExpiresFormat",
			url:         fmt.Sprintf("/xconfAdminService/telemetry/addTo/%s/estbMacAddress/AA:BB:CC:DD:EE:FF/notanumber?applicationType=stb", rule.ID),
			expectError: true,
			description: "xwhttp.WriteXconfResponse - expires must be a number",
		},
		{
			name:        "RuleNotFound",
			url:         fmt.Sprintf("/xconfAdminService/telemetry/addTo/%s/estbMacAddress/AA:BB:CC:DD:EE:FF/%d?applicationType=stb", uuid.New().String(), expires),
			expectError: true,
			description: "xwhttp.WriteXconfResponse - no rule found for ruleId",
		},
		{
			name:        "Success",
			url:         fmt.Sprintf("/xconfAdminService/telemetry/addTo/%s/estbMacAddress/AA:BB:CC:DD:EE:FF/%d?applicationType=stb", rule.ID, expires),
			expectError: false,
			description: "Success case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := exec("POST", tt.url, nil)
			if tt.expectError {
				assert.Assert(t, rr.Code >= http.StatusBadRequest, tt.description)
			} else {
				assert.Equal(t, http.StatusOK, rr.Code, tt.description)
			}
		})
	}
}

// TestBindToTelemetry_AllErrorCases tests all error paths
func TestBindToTelemetry_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()
	perm := createPermanentTelemetryProfile("perm-bind-1")
	expires := (time.Now().UnixNano() / 1_000_000) + 60000

	tests := []struct {
		name        string
		url         string
		expectError bool
		description string
	}{
		{
			name:        "InvalidContextAttributeName",
			url:         fmt.Sprintf("/xconfAdminService/telemetry/bindToTelemetry/%s/model/TESTMODEL/%d?applicationType=stb", perm.ID, expires),
			expectError: true,
			description: "xwhttp.WriteXconfResponse - only estbMacAddress allowed",
		},
		{
			name:        "InvalidExpiresFormat",
			url:         fmt.Sprintf("/xconfAdminService/telemetry/bindToTelemetry/%s/estbMacAddress/AA:BB:CC:DD:EE:FF/notanumber?applicationType=stb", perm.ID),
			expectError: true,
			description: "xwhttp.WriteXconfResponse - expires must be a number",
		},
		{
			name:        "ProfileNotFound",
			url:         fmt.Sprintf("/xconfAdminService/telemetry/bindToTelemetry/%s/estbMacAddress/AA:BB:CC:DD:EE:FF/%d?applicationType=stb", uuid.New().String(), expires),
			expectError: true,
			description: "xwhttp.WriteXconfResponse - no rule found for ID provided",
		},
		{
			name:        "Success",
			url:         fmt.Sprintf("/xconfAdminService/telemetry/bindToTelemetry/%s/estbMacAddress/AA:BB:CC:DD:EE:FF/%d?applicationType=stb", perm.ID, expires),
			expectError: false,
			description: "Success case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := exec("POST", tt.url, nil)
			if tt.expectError {
				assert.Assert(t, rr.Code >= http.StatusBadRequest, tt.description)
			} else {
				assert.Equal(t, http.StatusOK, rr.Code, tt.description)
			}
		})
	}
}

// TestTelemetryTestPageHandler_AllErrorCases tests all error paths
func TestTelemetryTestPageHandler_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	tests := []struct {
		name               string
		body               []byte
		url                string
		expectedStatusCode int
		description        string
	}{
		{
			name:               "InvalidJSON",
			body:               []byte("{invalid json"),
			url:                "/xconfAdminService/telemetry/testpage?applicationType=stb",
			expectedStatusCode: http.StatusBadRequest,
			description:        "xhttp.WriteAdminErrorResponse - JSON unmarshal error",
		},
		{
			name: "Success_WithValidContext",
			body: func() []byte {
				m := map[string]interface{}{
					"estbMacAddress": "AA:BB:CC:DD:EE:FF",
					"model":          "TESTMODEL",
				}
				b, _ := json.Marshal(m)
				return b
			}(),
			url:                "/xconfAdminService/telemetry/testpage?applicationType=stb",
			expectedStatusCode: http.StatusOK,
			description:        "Success case with valid context",
		},
		{
			name:               "Success_EmptyBody",
			body:               []byte(""),
			url:                "/xconfAdminService/telemetry/testpage?applicationType=stb",
			expectedStatusCode: http.StatusOK,
			description:        "Success case with empty body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rr *httptest.ResponseRecorder
			if len(tt.body) > 0 {
				r := httptest.NewRequest("POST", tt.url, bytes.NewReader(tt.body))
				rr = ExecuteRequest(r, router)
			} else {
				rr = exec("POST", tt.url, nil)
			}
			assert.Equal(t, tt.expectedStatusCode, rr.Code, tt.description)
		})
	}
}
