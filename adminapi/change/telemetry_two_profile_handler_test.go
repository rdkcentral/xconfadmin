package change

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
