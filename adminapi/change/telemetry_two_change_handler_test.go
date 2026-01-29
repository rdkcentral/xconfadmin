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
package change

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	xchange "github.com/rdkcentral/xconfadmin/shared/change"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	xwchange "github.com/rdkcentral/xconfwebconfig/shared/change"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/stretchr/testify/assert"
)

const validTelemetryTwoJSONHandler = "{\n    \"Description\":\"Test Json Data\",\n    \"Version\":\"0.1\",\n    \"Protocol\":\"HTTP\",\n    \"EncodingType\":\"JSON\",\n    \"ReportingInterval\":43200,\n    \"TimeReference\":\"0001-01-01T00:00:00Z\",\n    \"RootName\":\"someNewRootName\",\n    \"Parameter\": [ { \"type\": \"dataModel\", \"reference\": \"Profile.Name\"} ],\n    \"HTTP\": {\n        \"URL\":\"https://test.net\",\n        \"Compression\":\"None\",\n        \"Method\":\"POST\"\n    },\n    \"JSONEncoding\": {\n        \"ReportFormat\":\"NameValuePair\",\n        \"ReportTimestamp\": \"None\"\n    }\n}"

// minimal router setup: reuse global test router if available; fallback to direct handler invocation
// here we directly call handlers with crafted requests and XResponseWriter via httptest.ResponseRecorder

func marshal(v interface{}) []byte { b, _ := json.Marshal(v); return b }

func TestGetTwoProfileChangesHandler_Empty(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("GET", "/xconfAdminService/telemetry/v2/change/all?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetTwoProfileChangesHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "[]", strings.TrimSpace(rr.Body.String()))
}

func seedCreateChange(t *testing.T, name string) *xwchange.TelemetryTwoChange {
	p := &logupload.TelemetryTwoProfile{ID: uuid.New().String(), Name: name, Jsonconfig: validTelemetryTwoJSONHandler, ApplicationType: "stb"}
	ch := xchange.NewEmptyTelemetryTwoChange()
	ch.ID = uuid.New().String()
	ch.EntityID = p.ID
	ch.NewEntity = p
	ch.Operation = xchange.Create
	ch.EntityType = xchange.TelemetryTwoProfile
	ch.ApplicationType = "stb"
	ch.Author = "tester"
	if err := xchange.CreateOneTelemetryTwoChange(ch); err != nil {
		t.Fatalf("seed err: %v", err)
	}
	return ch
}

func TestApproveAndCancelAndRevertHandlers(t *testing.T) {
	cleanupChangeTest()
	// approve create
	ch := seedCreateChange(t, "ap1")
	r := httptest.NewRequest("GET", fmt.Sprintf("/xconfAdminService/telemetry/v2/change/approve/%s?applicationType=stb", ch.ID), nil)
	r = mux.SetURLVars(r, map[string]string{"changeId": ch.ID})
	rr := httptest.NewRecorder()
	ApproveTwoChangeHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	// cancel non-existing change -> create another and then cancel before approving
	ch2 := seedCreateChange(t, "ap2")
	r2 := httptest.NewRequest("GET", fmt.Sprintf("/xconfAdminService/telemetry/v2/change/cancel/%s?applicationType=stb", ch2.ID), nil)
	r2 = mux.SetURLVars(r2, map[string]string{"changeId": ch2.ID})
	rr2 := httptest.NewRecorder()
	CancelTwoChangeHandler(rr2, r2)
	assert.Equal(t, http.StatusOK, rr2.Code)
	// revert previously approved change
	approved := xchange.GetOneApprovedTelemetryTwoChange(ch.ID)
	r3 := httptest.NewRequest("GET", fmt.Sprintf("/xconfAdminService/telemetry/v2/change/revert/%s?applicationType=stb", approved.ID), nil)
	r3 = mux.SetURLVars(r3, map[string]string{"approveId": approved.ID})
	rr3 := httptest.NewRecorder()
	RevertTwoChangeHandler(rr3, r3)
	assert.Equal(t, http.StatusOK, rr3.Code)
}

func TestGroupedAndFilteredHandlers(t *testing.T) {
	cleanupChangeTest()
	for i := 0; i < 3; i++ {
		seedCreateChange(t, fmt.Sprintf("g%d", i))
	}
	// grouped
	r := httptest.NewRequest("GET", "/xconfAdminService/telemetry/v2/change/changes/grouped/byId?applicationType=stb&pageNumber=1&pageSize=2", nil)
	rr := httptest.NewRecorder()
	GetGroupedTwoChangesHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	// approved grouped (none yet) should still be 200
	r2 := httptest.NewRequest("GET", "/xconfAdminService/telemetry/v2/change/approved/grouped/byId?applicationType=stb&pageNumber=1&pageSize=2", nil)
	rr2 := httptest.NewRecorder()
	GetGroupedApprovedTwoChangesHandler(rr2, r2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}

func TestEntityIdsHandler(t *testing.T) {
	cleanupChangeTest()
	seedCreateChange(t, "e1")
	r := httptest.NewRequest("GET", "/xconfAdminService/telemetry/v2/change/entityIds?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetTwoChangeEntityIdsHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestPagingValidationErrors(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("GET", "/xconfAdminService/telemetry/v2/change/changes/grouped/byId?applicationType=stb&pageNumber=0&pageSize=2", nil)
	rr := httptest.NewRecorder()
	GetGroupedTwoChangesHandler(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	r2 := httptest.NewRequest("GET", "/xconfAdminService/telemetry/v2/change/approved/grouped/byId?applicationType=stb&pageNumber=1&pageSize=0", nil)
	rr2 := httptest.NewRecorder()
	GetGroupedApprovedTwoChangesHandler(rr2, r2)
	assert.Equal(t, http.StatusBadRequest, rr2.Code)
}

// ========== Tests for GetTwoChangesFilteredHandler ==========

func TestGetTwoChangesFilteredHandler_Success(t *testing.T) {
	cleanupChangeTest()
	// Create test changes
	ch1 := seedCreateChange(t, "FilterChange1")
	ch2 := seedCreateChange(t, "FilterChange2")

	// Test filtered request with pagination and empty body
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/changes/filtered?pageNumber=1&pageSize=10&applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("")
	GetTwoChangesFilteredHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	var changes []*xwchange.TelemetryTwoChange
	err := json.Unmarshal(rr.Body.Bytes(), &changes)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(changes), 2)

	// Cleanup
	xchange.DeleteOneTelemetryTwoChange(ch1.ID)
	xchange.DeleteOneTelemetryTwoChange(ch2.ID)
}

func TestGetTwoChangesFilteredHandler_WithContextFilter(t *testing.T) {
	cleanupChangeTest()
	// Create change with specific author
	ch := seedCreateChange(t, "AuthorFilterChange")

	// Filter by author
	filterBody := `{"AUTHOR":"tester"}`
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/changes/filtered?pageNumber=1&pageSize=10&applicationType=stb", strings.NewReader(filterBody))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(filterBody)
	GetTwoChangesFilteredHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	var changes []*xwchange.TelemetryTwoChange
	err := json.Unmarshal(rr.Body.Bytes(), &changes)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(changes), 1)

	// Cleanup
	xchange.DeleteOneTelemetryTwoChange(ch.ID)
}

func TestGetTwoChangesFilteredHandler_MissingPageNumber(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/changes/filtered?pageSize=10&applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetTwoChangesFilteredHandler(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid value for pageNumber")
}

func TestGetTwoChangesFilteredHandler_MissingPageSize(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/changes/filtered?pageNumber=1&applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetTwoChangesFilteredHandler(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid value for pageSize")
}

func TestGetTwoChangesFilteredHandler_InvalidPageNumber(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/changes/filtered?pageNumber=0&pageSize=10&applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetTwoChangesFilteredHandler(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid value for pageNumber")
}

func TestGetTwoChangesFilteredHandler_InvalidPageSize(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/changes/filtered?pageNumber=1&pageSize=0&applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetTwoChangesFilteredHandler(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid value for pageSize")
}

func TestGetTwoChangesFilteredHandler_InvalidJSON(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/changes/filtered?pageNumber=1&pageSize=10&applicationType=stb", strings.NewReader("invalid json"))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("invalid json")
	GetTwoChangesFilteredHandler(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetTwoChangesFilteredHandler_EmptyResult(t *testing.T) {
	cleanupChangeTest()
	defer cleanupChangeTest() // Ensure cleanup even if test fails
	// No changes created in this test, should return empty or valid array
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/changes/filtered?pageNumber=1&pageSize=10&applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("")
	GetTwoChangesFilteredHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify response is a valid JSON array
	var changes []interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &changes)
	assert.NoError(t, err)
}

// ========== Tests for GetApprovedTwoChangesFilteredHandler ==========

func TestGetApprovedTwoChangesFilteredHandler_Success(t *testing.T) {
	cleanupChangeTest()
	// Create and approve a change
	ch := seedCreateChange(t, "ApprovedFilterChange")
	r := httptest.NewRequest("GET", fmt.Sprintf("/approve/%s?applicationType=stb", ch.ID), nil)
	r = mux.SetURLVars(r, map[string]string{"changeId": ch.ID})
	rr := httptest.NewRecorder()
	ApproveTwoChangeHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Now filter approved changes
	r2 := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approved/filtered?pageNumber=1&pageSize=10&applicationType=stb", nil)
	rr2 := httptest.NewRecorder()
	xw2 := xwhttp.NewXResponseWriter(rr2)
	xw2.SetBody("")
	GetApprovedTwoChangesFilteredHandler(xw2, r2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	var approvedChanges []*xwchange.ApprovedTelemetryTwoChange
	err := json.Unmarshal(rr2.Body.Bytes(), &approvedChanges)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(approvedChanges), 1)

	// Cleanup
	for _, ac := range approvedChanges {
		xchange.DeleteOneApprovedTelemetryTwoChange(ac.ID)
	}
}

func TestGetApprovedTwoChangesFilteredHandler_WithFilter(t *testing.T) {
	cleanupChangeTest()
	// Create and approve a change
	ch := seedCreateChange(t, "ApprovedWithAuthor")
	r := httptest.NewRequest("GET", fmt.Sprintf("/approve/%s?applicationType=stb", ch.ID), nil)
	r = mux.SetURLVars(r, map[string]string{"changeId": ch.ID})
	rr := httptest.NewRecorder()
	ApproveTwoChangeHandler(rr, r)

	// Filter by author
	filterBody := `{"AUTHOR":"tester"}`
	r2 := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approved/filtered?pageNumber=1&pageSize=10&applicationType=stb", strings.NewReader(filterBody))
	rr2 := httptest.NewRecorder()
	xw2 := xwhttp.NewXResponseWriter(rr2)
	xw2.SetBody(filterBody)
	GetApprovedTwoChangesFilteredHandler(xw2, r2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	var approvedChanges []*xwchange.ApprovedTelemetryTwoChange
	err := json.Unmarshal(rr2.Body.Bytes(), &approvedChanges)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(approvedChanges), 1)

	// Cleanup
	for _, ac := range approvedChanges {
		xchange.DeleteOneApprovedTelemetryTwoChange(ac.ID)
	}
}

func TestGetApprovedTwoChangesFilteredHandler_MissingPageNumber(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approved/filtered?pageSize=10&applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetApprovedTwoChangesFilteredHandler(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid value for pageNumber")
}

func TestGetApprovedTwoChangesFilteredHandler_MissingPageSize(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approved/filtered?pageNumber=1&applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetApprovedTwoChangesFilteredHandler(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid value for pageSize")
}

func TestGetApprovedTwoChangesFilteredHandler_InvalidJSON(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approved/filtered?pageNumber=1&pageSize=10&applicationType=stb", strings.NewReader("{invalid}"))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("{invalid}")
	GetApprovedTwoChangesFilteredHandler(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetApprovedTwoChangesFilteredHandler_EmptyResult(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approved/filtered?pageNumber=1&pageSize=10&applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("")
	GetApprovedTwoChangesFilteredHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "[]", strings.TrimSpace(rr.Body.String()))
}

// ========== Tests for RevertTwoChangesHandler ==========

func TestRevertTwoChangesHandler_Success(t *testing.T) {
	cleanupChangeTest()
	// Create and approve a change
	ch := seedCreateChange(t, "RevertMultiple1")
	ch2 := seedCreateChange(t, "RevertMultiple2")

	r1 := httptest.NewRequest("GET", fmt.Sprintf("/approve/%s?applicationType=stb", ch.ID), nil)
	r1 = mux.SetURLVars(r1, map[string]string{"changeId": ch.ID})
	rr1 := httptest.NewRecorder()
	ApproveTwoChangeHandler(rr1, r1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	r2 := httptest.NewRequest("GET", fmt.Sprintf("/approve/%s?applicationType=stb", ch2.ID), nil)
	r2 = mux.SetURLVars(r2, map[string]string{"changeId": ch2.ID})
	rr2 := httptest.NewRecorder()
	ApproveTwoChangeHandler(rr2, r2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	// Get approved change IDs
	approved1 := xchange.GetOneApprovedTelemetryTwoChange(ch.ID)
	approved2 := xchange.GetOneApprovedTelemetryTwoChange(ch2.ID)
	assert.NotNil(t, approved1)
	assert.NotNil(t, approved2)

	// Revert multiple changes
	idList := []string{approved1.ID, approved2.ID}
	body := marshal(idList)
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/revert?applicationType=stb", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	RevertTwoChangesHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify response - should be a map of errors (empty if all succeeded)
	var errorMap map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorMap)
	assert.NoError(t, err)
}

func TestRevertTwoChangesHandler_InvalidJSON(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/revert?applicationType=stb", strings.NewReader("not json"))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("not json")
	RevertTwoChangesHandler(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRevertTwoChangesHandler_EmptyList(t *testing.T) {
	cleanupChangeTest()
	body := marshal([]string{})
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/revert?applicationType=stb", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	RevertTwoChangesHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	var errorMap map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorMap)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(errorMap))
}

func TestRevertTwoChangesHandler_NonExistentIds(t *testing.T) {
	cleanupChangeTest()
	// Try to revert non-existent approved changes
	idList := []string{"non-existent-id-1", "non-existent-id-2"}
	body := marshal(idList)
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/revert?applicationType=stb", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	RevertTwoChangesHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Should still return 200 but with empty error map (non-existent changes are skipped)
	var errorMap map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorMap)
	assert.NoError(t, err)
}

// ========== Tests for ApproveTwoChangesHandler ==========

func TestApproveTwoChangesHandler_Success(t *testing.T) {
	cleanupChangeTest()
	// Create multiple changes
	ch1 := seedCreateChange(t, "ApproveMulti1")
	ch2 := seedCreateChange(t, "ApproveMulti2")

	// Approve multiple changes
	idList := []string{ch1.ID, ch2.ID}
	body := marshal(idList)
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approve?applicationType=stb", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	ApproveTwoChangesHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify response - error map should be empty if all succeeded
	var errorMap map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorMap)
	assert.NoError(t, err)

	// Verify changes were approved
	approved1 := xchange.GetOneApprovedTelemetryTwoChange(ch1.ID)
	approved2 := xchange.GetOneApprovedTelemetryTwoChange(ch2.ID)
	assert.NotNil(t, approved1)
	assert.NotNil(t, approved2)

	// Cleanup
	if approved1 != nil {
		xchange.DeleteOneApprovedTelemetryTwoChange(approved1.ID)
	}
	if approved2 != nil {
		xchange.DeleteOneApprovedTelemetryTwoChange(approved2.ID)
	}
}

func TestApproveTwoChangesHandler_InvalidJSON(t *testing.T) {
	cleanupChangeTest()
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approve?applicationType=stb", strings.NewReader("{not valid json"))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("{not valid json")
	ApproveTwoChangesHandler(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestApproveTwoChangesHandler_EmptyList(t *testing.T) {
	cleanupChangeTest()
	body := marshal([]string{})
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approve?applicationType=stb", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	ApproveTwoChangesHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	var errorMap map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorMap)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(errorMap))
}

func TestApproveTwoChangesHandler_NonExistentIds(t *testing.T) {
	cleanupChangeTest()
	// Try to approve non-existent changes
	idList := []string{"non-existent-id-1", "non-existent-id-2"}
	body := marshal(idList)
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approve?applicationType=stb", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	ApproveTwoChangesHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Should return empty error map (non-existent changes are filtered out)
	var errorMap map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorMap)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(errorMap))
}

func TestApproveTwoChangesHandler_MixedValidInvalid(t *testing.T) {
	cleanupChangeTest()
	// Create one valid change and try to approve it along with invalid IDs
	ch := seedCreateChange(t, "ApproveMixed")

	idList := []string{ch.ID, "non-existent-id"}
	body := marshal(idList)
	r := httptest.NewRequest("POST", "/xconfAdminService/telemetry/v2/change/approve?applicationType=stb", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	ApproveTwoChangesHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify the valid one got approved
	approved := xchange.GetOneApprovedTelemetryTwoChange(ch.ID)
	assert.NotNil(t, approved)

	// Cleanup
	if approved != nil {
		xchange.DeleteOneApprovedTelemetryTwoChange(approved.ID)
	}
}

// ========== Tests for GetApprovedTwoChangesHandler ==========

func TestGetApprovedTwoChangesHandler_Success(t *testing.T) {
	cleanupChangeTest()
	// Create and approve changes
	ch1 := seedCreateChange(t, "GetApproved1")
	ch2 := seedCreateChange(t, "GetApproved2")

	r1 := httptest.NewRequest("GET", fmt.Sprintf("/approve/%s?applicationType=stb", ch1.ID), nil)
	r1 = mux.SetURLVars(r1, map[string]string{"changeId": ch1.ID})
	rr1 := httptest.NewRecorder()
	ApproveTwoChangeHandler(rr1, r1)

	r2 := httptest.NewRequest("GET", fmt.Sprintf("/approve/%s?applicationType=stb", ch2.ID), nil)
	r2 = mux.SetURLVars(r2, map[string]string{"changeId": ch2.ID})
	rr2 := httptest.NewRecorder()
	ApproveTwoChangeHandler(rr2, r2)

	// Get all approved changes
	r := httptest.NewRequest("GET", "/xconfAdminService/telemetry/v2/change/approved/all?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetApprovedTwoChangesHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	var approvedChanges []*xwchange.ApprovedTelemetryTwoChange
	err := json.Unmarshal(rr.Body.Bytes(), &approvedChanges)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(approvedChanges), 2)

	// Cleanup
	for _, ac := range approvedChanges {
		xchange.DeleteOneApprovedTelemetryTwoChange(ac.ID)
	}
}

func TestGetApprovedTwoChangesHandler_Empty(t *testing.T) {
	cleanupChangeTest()
	// No approved changes
	r := httptest.NewRequest("GET", "/xconfAdminService/telemetry/v2/change/approved/all?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetApprovedTwoChangesHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "[]", strings.TrimSpace(rr.Body.String()))
}

func TestGetApprovedTwoChangesHandler_ApplicationTypeFilter(t *testing.T) {
	cleanupChangeTest()
	// Create and approve a change
	ch := seedCreateChange(t, "AppTypeTest")
	r1 := httptest.NewRequest("GET", fmt.Sprintf("/approve/%s?applicationType=stb", ch.ID), nil)
	r1 = mux.SetURLVars(r1, map[string]string{"changeId": ch.ID})
	rr1 := httptest.NewRecorder()
	ApproveTwoChangeHandler(rr1, r1)

	// Get approved changes for stb
	r := httptest.NewRequest("GET", "/xconfAdminService/telemetry/v2/change/approved/all?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetApprovedTwoChangesHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	var approvedChanges []*xwchange.ApprovedTelemetryTwoChange
	err := json.Unmarshal(rr.Body.Bytes(), &approvedChanges)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(approvedChanges), 1)

	// Verify all returned changes are for stb application type
	for _, ac := range approvedChanges {
		assert.Equal(t, "stb", ac.ApplicationType)
		xchange.DeleteOneApprovedTelemetryTwoChange(ac.ID)
	}
}
