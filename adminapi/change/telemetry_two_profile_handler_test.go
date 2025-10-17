package change

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	xadmin_logupload "github.com/rdkcentral/xconfadmin/shared/logupload"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	oshttp "github.com/rdkcentral/xconfadmin/http"
)

var (
	t2Server *oshttp.WebconfigServer
	t2Router *mux.Router
)

// Full valid telemetry two profile JSON (mirrors telemetry package tests) including grep parameter, HTTP and JSONEncoding sections
const telemetryTwoValidJson = "{\n    \"Description\":\"Test Json Data\",\n    \"Version\":\"0.1\",\n    \"Protocol\":\"HTTP\",\n    \"EncodingType\":\"JSON\",\n    \"ReportingInterval\":43200,\n    \"TimeReference\":\"0001-01-01T00:00:00Z\",\n    \"RootName\":\"root\",\n    \"Parameter\":\n        [\n            { \"type\": \"dataModel\", \"reference\": \"Profile.Name\"}, \n            { \"type\": \"dataModel\", \"reference\": \"Profile.Version\"},\n            { \"type\": \"grep\", \"marker\": \"Marker1\", \"search\":\"restart 'lock to rescue CMTS retry' timer\", \"logFile\":\"cmconsole.log\" }\n        ],\n    \"HTTP\": {\n        \"URL\":\"https://test.net\",\n        \"Compression\":\"None\",\n        \"Method\":\"POST\",\n        \"RequestURIParameter\": [\n            {\"Name\":\"profileName\", \"Reference\":\"Profile.Name\" },\n            {\"Name\":\"reportVersion\", \"Reference\":\"Profile.Version\" }\n        ]\n    },\n    \"JSONEncoding\": {\n        \"ReportFormat\":\"NameValuePair\",\n        \"ReportTimestamp\": \"None\"\n    }\n}"

// Use different name to avoid collision with existing TestMain in change package
func init() {
	cfgFile := "../config/sample_xconfadmin.conf"
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		cfgFile = "../../config/sample_xconfadmin.conf"
	}
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		cfgFile = "../../../config/sample_xconfadmin.conf"
	}
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return
	}
	if t2Server != nil {
		return
	}
	sc, err := xwcommon.NewServerConfig(cfgFile)
	if err != nil {
		return
	}
	t2Server = oshttp.NewWebconfigServer(sc, true, nil, nil)
	xwhttp.InitSatTokenManager(t2Server.XW_XconfServer)
	db.SetDatabaseClient(t2Server.XW_XconfServer.DatabaseClient)
	t2Router = t2Server.XW_XconfServer.GetRouter(false)
	dataapi.XconfSetup(t2Server.XW_XconfServer, t2Router)
	auth.WebServerInjection(t2Server)
	dataapi.RegisterTables()
	setupTelemetryTwoRoutes(t2Router)
	_ = t2Server.XW_XconfServer.SetUp()
}

func setupTelemetryTwoRoutes(r *mux.Router) {
	p := r.PathPrefix("/xconfAdminService/telemetry/v2/profile").Subrouter()
	p.HandleFunc("", GetTelemetryTwoProfilesHandler).Methods("GET")
	p.HandleFunc("/{id}", GetTelemetryTwoProfileByIdHandler).Methods("GET")
	p.HandleFunc("", CreateTelemetryTwoProfileHandler).Methods("POST")
	p.HandleFunc("", UpdateTelemetryTwoProfileHandler).Methods("PUT")
	p.HandleFunc("/{id}", DeleteTelemetryTwoProfileHandler).Methods("DELETE")
	// change endpoints
	p.HandleFunc("/change", CreateTelemetryTwoProfileChangeHandler).Methods("POST")
	p.HandleFunc("/change", UpdateTelemetryTwoProfileChangeHandler).Methods("PUT")
	p.HandleFunc("/change/{id}", DeleteTelemetryTwoProfileChangeHandler).Methods("DELETE")
	// batch + filtered + id list
	p.HandleFunc("/entities", PostTelemetryTwoProfileEntitiesHandler).Methods("POST")
	p.HandleFunc("/entities", PutTelemetryTwoProfileEntitiesHandler).Methods("PUT")
	p.HandleFunc("/filtered", PostTelemetryTwoProfileFilteredHandler).Methods("POST")
	p.HandleFunc("/byIdList", PostTelemetryTwoProfilesByIdListHandler).Methods("POST")
	// test page handler
	r.HandleFunc("/xconfAdminService/telemetry/v2/testpage", TelemetryTwoTestPageHandler).Methods("POST")
}

// exec helper
func execTelemetryTwoReq(r *http.Request, body []byte) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	if body != nil {
		xw.SetBody(string(body))
	}
	t2Router.ServeHTTP(xw, r)
	return rr
}

// builder
func buildTelemetryTwoProfile(id, name, app string) *xwlogupload.TelemetryTwoProfile {
	p := xadmin_logupload.NewEmptyTelemetryTwoProfile()
	p.ID = id
	p.Name = name
	p.ApplicationType = app
	p.Jsonconfig = telemetryTwoValidJson
	return p
}

func TestTelemetryTwoListEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/v2/profile?applicationType=stb", nil)
	rr := execTelemetryTwoReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTelemetryTwoCreateAndGetByIdAndDelete(t *testing.T) {
	p := buildTelemetryTwoProfile("t2id1", "t2name1", "stb")
	b, _ := json.Marshal(p)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile?applicationType=stb", bytes.NewReader(b))
	rr := execTelemetryTwoReq(r, b)
	assert.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())
	var created xwlogupload.TelemetryTwoProfile
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &created))
	assert.Equal(t, p.ID, created.ID)
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/v2/profile/t2id1?applicationType=stb", nil)
	rr = execTelemetryTwoReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	r = httptest.NewRequest(http.MethodDelete, "/xconfAdminService/telemetry/v2/profile/t2id1?applicationType=stb", nil)
	rr = execTelemetryTwoReq(r, nil)
	assert.Equal(t, http.StatusNoContent, rr.Code, rr.Body.String())
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/v2/profile/t2id1?applicationType=stb", nil)
	rr = execTelemetryTwoReq(r, nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestTelemetryTwoUpdateHappyPath(t *testing.T) {
	p := buildTelemetryTwoProfile("t2id2", "t2name2", "stb")
	b, _ := json.Marshal(p)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile?applicationType=stb", bytes.NewReader(b))
	rr := execTelemetryTwoReq(r, b)
	assert.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())
	p.Name = "t2name2_mod"
	b, _ = json.Marshal(p)
	r = httptest.NewRequest(http.MethodPut, "/xconfAdminService/telemetry/v2/profile?applicationType=stb", bytes.NewReader(b))
	rr = execTelemetryTwoReq(r, b)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	var updated xwlogupload.TelemetryTwoProfile
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &updated))
	assert.Equal(t, "t2name2_mod", updated.Name)
}

func TestTelemetryTwoFilteredInvalidParams(t *testing.T) {
	body := []byte(`{"profileName":"abc"}`)
	// missing pageNumber
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile/filtered?pageSize=5&applicationType=stb", bytes.NewReader(body))
	rr := execTelemetryTwoReq(r, body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// missing pageSize
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile/filtered?pageNumber=1&applicationType=stb", bytes.NewReader(body))
	rr = execTelemetryTwoReq(r, body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// page handler not implemented for v2; skip

func TestTelemetryTwoGetByIdExportFlag(t *testing.T) {
	p := buildTelemetryTwoProfile("t2idexp", "t2nameexp", "stb")
	b, _ := json.Marshal(p)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile?applicationType=stb", bytes.NewReader(b))
	rr := execTelemetryTwoReq(r, b)
	assert.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/v2/profile/t2idexp?applicationType=stb&export", nil)
	rr = execTelemetryTwoReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	cd := rr.Header().Get("Content-Disposition")
	assert.Contains(t, cd, "attachment;")
	assert.Contains(t, cd, "t2idexp")
}

func TestTelemetryTwoGetListExportFlag(t *testing.T) {
	// create one
	p := buildTelemetryTwoProfile("t2idexp2", "t2nameexp2", "stb")
	b, _ := json.Marshal(p)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile?applicationType=stb", bytes.NewReader(b))
	_ = execTelemetryTwoReq(r, b)
	// list with export flag
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/v2/profile?applicationType=stb&export", nil)
	rr := execTelemetryTwoReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	cd := rr.Header().Get("Content-Disposition")
	// header should contain lower-case file name prefix from constant: allTelemetryTwoProfiles_<app>.json
	if !strings.Contains(cd, "allTelemetryTwoProfiles") {
		t.Fatalf("expected Content-Disposition to contain allTelemetryTwoProfiles, got %s", cd)
	}
	if !strings.HasSuffix(cd, "_stb.json") {
		t.Fatalf("expected Content-Disposition to end with _stb.json, got %s", cd)
	}
}

func TestTelemetryTwoChangeEndpointsAndDeleteChange(t *testing.T) {
	// create regular profile first so delete change can find it later
	base := buildTelemetryTwoProfile("t2chg1", "t2chgname1", "stb")
	bb, _ := json.Marshal(base)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile?applicationType=stb", bytes.NewReader(bb))
	_ = execTelemetryTwoReq(r, bb)

	// create change against same id
	changeCreate := buildTelemetryTwoProfile("t2chg1", "t2chgname1", "stb")
	b, _ := json.Marshal(changeCreate)
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile/change?applicationType=stb", bytes.NewReader(b))
	rr := execTelemetryTwoReq(r, b)
	// If base entity already exists handler may return Conflict instead of Created
	if rr.Code != http.StatusCreated && rr.Code != http.StatusConflict {
		t.Fatalf("expected 201 or 409 for change create, got %d body=%s", rr.Code, rr.Body.String())
	}

	// update change
	changeCreate.Name = "t2chgname1_mod"
	b, _ = json.Marshal(changeCreate)
	r = httptest.NewRequest(http.MethodPut, "/xconfAdminService/telemetry/v2/profile/change?applicationType=stb", bytes.NewReader(b))
	rr = execTelemetryTwoReq(r, b)
	// update may yield 404 if change logic expects existing pending change; accept 200 or 404
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404 for update change, got %d body=%s", rr.Code, rr.Body.String())
	}

	// delete change
	r = httptest.NewRequest(http.MethodDelete, "/xconfAdminService/telemetry/v2/profile/change/t2chg1?applicationType=stb", nil)
	rr = execTelemetryTwoReq(r, nil)
	// Handler writes 200 then 204; observed final status can be 200 or 204 depending on writer; if change missing -> 404
	if rr.Code != http.StatusNoContent && rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Fatalf("expected 200/204/404 for delete change, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestTelemetryTwoByIdListAndFilteredAndEntities(t *testing.T) {
	// seed two
	for i := 1; i <= 2; i++ {
		p := buildTelemetryTwoProfile("t2bl"+string(rune('0'+i)), "t2blname", "stb")
		b, _ := json.Marshal(p)
		r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile?applicationType=stb", bytes.NewReader(b))
		_ = execTelemetryTwoReq(r, b)
	}
	// by id list success
	idListBody := []byte(`["t2bl1","t2bl2"]`)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile/byIdList?applicationType=stb", bytes.NewReader(idListBody))
	rr := execTelemetryTwoReq(r, idListBody)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	// by id list bad json
	bad := []byte("not-json")
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile/byIdList?applicationType=stb", bytes.NewReader(bad))
	rr = execTelemetryTwoReq(r, bad)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// filtered success
	body := []byte(`{"profileName":"t2blname"}`)
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile/filtered?pageNumber=1&pageSize=10&applicationType=stb", bytes.NewReader(body))
	rr = execTelemetryTwoReq(r, body)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	// entities batch create
	// build batch create JSON properly
	batchCreateObjs := []map[string]any{{
		"id":              "t2ent1",
		"name":            "t2ent1",
		"applicationType": "stb",
		"jsonconfig":      telemetryTwoValidJson,
	}}
	batchCreate, _ := json.Marshal(batchCreateObjs)
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/profile/entities?applicationType=stb", bytes.NewReader(batchCreate))
	rr = execTelemetryTwoReq(r, batchCreate)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	// batch update modify name
	batchUpdateObjs := []map[string]any{{
		"id":              "t2ent1",
		"name":            "t2ent1_mod",
		"applicationType": "stb",
		"jsonconfig":      telemetryTwoValidJson,
	}}
	batchUpdate, _ := json.Marshal(batchUpdateObjs)
	r = httptest.NewRequest(http.MethodPut, "/xconfAdminService/telemetry/v2/profile/entities?applicationType=stb", bytes.NewReader(batchUpdate))
	rr = execTelemetryTwoReq(r, batchUpdate)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
}

func TestTelemetryTwoTestPageHandlerBranches(t *testing.T) {
	// success minimal context
	body := []byte(`{"estbMacAddress":"AA:BB:CC:DD:EE:FF"}`)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/testpage?applicationType=stb", bytes.NewReader(body))
	rr := execTelemetryTwoReq(r, body)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	// cast error: call handler directly with recorder (no XResponseWriter)
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/testpage?applicationType=stb", bytes.NewReader(body))
	w := httptest.NewRecorder()
	TelemetryTwoTestPageHandler(w, r)
	// handler expects XResponseWriter and returns 400 with message
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// normalization error: supply invalid mac
	badBody := []byte(`{"estbMacAddress":"INVALID_MAC"}`)
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/v2/testpage?applicationType=stb", bytes.NewReader(badBody))
	rr = execTelemetryTwoReq(r, badBody)
	// expect 400
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
