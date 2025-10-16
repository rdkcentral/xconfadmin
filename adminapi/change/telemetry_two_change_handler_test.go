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
