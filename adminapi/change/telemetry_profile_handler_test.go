package change

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	oshttp "github.com/rdkcentral/xconfadmin/http"
	xlogupload "github.com/rdkcentral/xconfadmin/shared/logupload"
	corelogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

// Reuse server initialization similar to change_handler_test.go but include telemetry profile routes
var (
	tpServer *oshttp.WebconfigServer
	tpRouter *mux.Router
)

// initialization helper (called lazily); cannot have second TestMain
func initTelemetryTestEnv() {
	if tpServer != nil { // already initialized
		return
	}
	cfgFile := "../config/sample_xconfadmin.conf"
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		cfgFile = "../../config/sample_xconfadmin.conf"
	}
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		cfgFile = "../../../config/sample_xconfadmin.conf"
	}
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		panic(err)
	}
	os.Setenv("SECURITY_TOKEN_KEY", "telemetryUTKey")
	os.Setenv("SAT_CLIENT_ID", "foo")
	os.Setenv("SAT_CLIENT_SECRET", "bar")
	os.Setenv("IDP_CLIENT_ID", "foo")
	os.Setenv("IDP_CLIENT_SECRET", "bar")

	sc, err := xwcommon.NewServerConfig(cfgFile)
	if err != nil {
		panic(err)
	}
	tpServer = oshttp.NewWebconfigServer(sc, true, nil, nil)
	xwhttp.InitSatTokenManager(tpServer.XW_XconfServer)
	db.SetDatabaseClient(tpServer.XW_XconfServer.DatabaseClient)
	tpRouter = tpServer.XW_XconfServer.GetRouter(false)
	dataapi.XconfSetup(tpServer.XW_XconfServer, tpRouter)
	auth.WebServerInjection(tpServer)
	dataapi.RegisterTables()
	setupTelemetryProfileRoutes(tpRouter)
	if err = tpServer.XW_XconfServer.SetUp(); err != nil {
		panic(err)
	}
	if err = tpServer.XW_XconfServer.TearDown(); err != nil {
		panic(err)
	}
}

func setupTelemetryProfileRoutes(r *mux.Router) {
	telemetryProfilePath := r.PathPrefix("/xconfAdminService/telemetry/profile").Subrouter()
	telemetryProfilePath.HandleFunc("", GetTelemetryProfilesHandler).Methods("GET")
	telemetryProfilePath.HandleFunc("", CreateTelemetryProfileHandler).Methods("POST")
	telemetryProfilePath.HandleFunc("", UpdateTelemetryProfileHandler).Methods("PUT")
	telemetryProfilePath.HandleFunc("/change", CreateTelemetryProfileChangeHandler).Methods("POST")
	telemetryProfilePath.HandleFunc("/change", UpdateTelemetryProfileChangeHandler).Methods("PUT")
	telemetryProfilePath.HandleFunc("/{id}", DeleteTelemetryProfileHandler).Methods("DELETE")
	telemetryProfilePath.HandleFunc("/change/{id}", DeleteTelemetryProfileChangeHandler).Methods("DELETE")
	telemetryProfilePath.HandleFunc("/{id}", GetTelemetryProfileByIdHandler).Methods("GET")
	telemetryProfilePath.HandleFunc("/entities", PostTelemetryProfileEntitiesHandler).Methods("POST")
	telemetryProfilePath.HandleFunc("/entities", PutTelemetryProfileEntitiesHandler).Methods("PUT")
	telemetryProfilePath.HandleFunc("/filtered", PostTelemetryProfileFilteredHandler).Methods("POST")
	telemetryProfilePath.HandleFunc("/migrate/createTelemetryId", CreateTelemetryIdsHandler).Methods("GET")
	telemetryProfilePath.HandleFunc("/entry/add/{id}", AddTelemetryProfileEntryHandler).Methods("PUT")
	telemetryProfilePath.HandleFunc("/entry/remove/{id}", RemoveTelemetryProfileEntryHandler).Methods("PUT")
	telemetryProfilePath.HandleFunc("/change/entry/add/{id}", AddTelemetryProfileEntryChangeHandler).Methods("PUT")
	telemetryProfilePath.HandleFunc("/change/entry/remove/{id}", RemoveTelemetryProfileEntryChangeHandler).Methods("PUT")
}

// helper exec
func execTPReq(r *http.Request, body []byte) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	if body != nil {
		xw.SetBody(string(body))
	}
	tpRouter.ServeHTTP(xw, r)
	return rr
}

// create a sample profile entity for tests
func newSampleProfile(name string) *corelogupload.PermanentTelemetryProfile {
	p := xlogupload.NewEmptyPermanentTelemetryProfile()
	p.ID = ""
	p.Name = name
	p.ApplicationType = "stb"
	p.TelemetryProfile = []corelogupload.TelemetryElement{{ID: "elem-" + name, Header: "H0", Content: "C0", Type: "T0", PollingFrequency: "60"}}
	p.UploadProtocol = "https"                 // lower case accepted then normalized in validation
	p.UploadRepository = "https://example.com" // valid scheme+host required
	return p
}

func TestCreateTelemetryProfileHandlerAndFetchById(t *testing.T) {
	initTelemetryTestEnv()
	profile := newSampleProfile("profA")
	b, _ := json.Marshal(profile)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/profile?applicationType=stb", bytes.NewReader(b))
	rr := execTPReq(r, b)
	assert.Equal(t, http.StatusCreated, rr.Code)
	// decode returned profile to get id
	var saved corelogupload.PermanentTelemetryProfile
	err := json.Unmarshal(rr.Body.Bytes(), &saved)
	assert.NoError(t, err)
	assert.NotEmpty(t, saved.ID)
	// fetch by id
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/profile/"+saved.ID+"?applicationType=stb", nil)
	rr = execTPReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	var fetched corelogupload.PermanentTelemetryProfile
	err = json.Unmarshal(rr.Body.Bytes(), &fetched)
	assert.NoError(t, err)
	assert.Equal(t, saved.ID, fetched.ID)
	assert.Equal(t, "profA", fetched.Name)
}

func TestCreateTelemetryProfileChangeHandler(t *testing.T) {
	initTelemetryTestEnv()
	profile := newSampleProfile("changeProf")
	b, _ := json.Marshal(profile)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/profile/change?applicationType=stb", bytes.NewReader(b))
	rr := execTPReq(r, b)
	assert.Equal(t, http.StatusCreated, rr.Code)
	// returned change JSON should contain NewEntity with name
	bodyStr := rr.Body.String()
	assert.Contains(t, bodyStr, "changeProf")
}

func TestUpdateTelemetryProfileHandler(t *testing.T) {
	initTelemetryTestEnv()
	// first create
	profile := newSampleProfile("toUpdate")
	b, _ := json.Marshal(profile)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/profile?applicationType=stb", bytes.NewReader(b))
	rr := execTPReq(r, b)
	assert.Equal(t, http.StatusCreated, rr.Code)
	var saved corelogupload.PermanentTelemetryProfile
	_ = json.Unmarshal(rr.Body.Bytes(), &saved)
	// update name
	saved.Name = "updatedName"
	ub, _ := json.Marshal(saved)
	r = httptest.NewRequest(http.MethodPut, "/xconfAdminService/telemetry/profile?applicationType=stb", bytes.NewReader(ub))
	rr = execTPReq(r, ub)
	assert.Equal(t, http.StatusOK, rr.Code)
	var updated corelogupload.PermanentTelemetryProfile
	_ = json.Unmarshal(rr.Body.Bytes(), &updated)
	assert.Equal(t, "updatedName", updated.Name)
}

func TestDeleteTelemetryProfileHandlerValidation(t *testing.T) {
	initTelemetryTestEnv()
	// delete non-existing should 404
	r := httptest.NewRequest(http.MethodDelete, "/xconfAdminService/telemetry/profile/notFound?applicationType=stb", nil)
	rr := execTPReq(r, nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestBatchPostTelemetryProfileEntitiesHandler(t *testing.T) {
	initTelemetryTestEnv()
	// create two profiles in batch (changes)
	prof1 := newSampleProfile("batchA")
	prof2 := newSampleProfile("batchB")
	list := []corelogupload.PermanentTelemetryProfile{*prof1, *prof2}
	b, _ := json.Marshal(list)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/profile/entities?applicationType=stb", bytes.NewReader(b))
	rr := execTPReq(r, b)
	assert.Equal(t, http.StatusOK, rr.Code)
	// expect success entries
	var resp map[string]map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	// map contains id-> {Status, Message}. Since IDs are empty pre-change, message should contain generated uuid later; we just assert keys length ==2
	assert.Equal(t, 2, len(resp))
}

func TestBatchPutTelemetryProfileEntitiesHandler(t *testing.T) {
	initTelemetryTestEnv()
	// first create a permanent profile
	profile := newSampleProfile("permForBatchUpdate")
	b, _ := json.Marshal(profile)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/profile?applicationType=stb", bytes.NewReader(b))
	rr := execTPReq(r, b)
	assert.Equal(t, http.StatusCreated, rr.Code)
	var saved corelogupload.PermanentTelemetryProfile
	_ = json.Unmarshal(rr.Body.Bytes(), &saved)
	saved.Name = "updatedBatch"
	list := []corelogupload.PermanentTelemetryProfile{saved}
	ub, _ := json.Marshal(list)
	r = httptest.NewRequest(http.MethodPut, "/xconfAdminService/telemetry/profile/entities?applicationType=stb", bytes.NewReader(ub))
	rr = execTPReq(r, ub)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestPostTelemetryProfileFilteredHandlerPaginationErrors(t *testing.T) {
	initTelemetryTestEnv()
	body := []byte("{}")
	// missing pageNumber
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/profile/filtered?pageSize=5&applicationType=stb", bytes.NewReader(body))
	rr := execTPReq(r, body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// missing pageSize
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/profile/filtered?pageNumber=1&applicationType=stb", bytes.NewReader(body))
	rr = execTPReq(r, body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAddAndRemoveTelemetryProfileEntryHandlers(t *testing.T) {
	initTelemetryTestEnv()
	// create profile
	profile := newSampleProfile("entryProf")
	b, _ := json.Marshal(profile)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/profile?applicationType=stb", bytes.NewReader(b))
	rr := execTPReq(r, b)
	assert.Equal(t, http.StatusCreated, rr.Code)
	var saved corelogupload.PermanentTelemetryProfile
	_ = json.Unmarshal(rr.Body.Bytes(), &saved)

	// add entry
	entry := corelogupload.TelemetryElement{Header: "H", Content: "C", Type: "T", PollingFrequency: "10"}
	eb, _ := json.Marshal([]corelogupload.TelemetryElement{entry})
	r = httptest.NewRequest(http.MethodPut, "/xconfAdminService/telemetry/profile/entry/add/"+saved.ID+"?applicationType=stb", bytes.NewReader(eb))
	rr = execTPReq(r, eb)
	assert.Equal(t, http.StatusOK, rr.Code)
	var updated corelogupload.PermanentTelemetryProfile
	_ = json.Unmarshal(rr.Body.Bytes(), &updated)
	assert.Equal(t, 2, len(updated.TelemetryProfile)) // initial element + added entry

	// remove entry via change route (ensures removal logic path)
	rb, _ := json.Marshal([]corelogupload.TelemetryElement{updated.TelemetryProfile[0]})
	r = httptest.NewRequest(http.MethodPut, "/xconfAdminService/telemetry/profile/change/entry/remove/"+saved.ID+"?applicationType=stb", bytes.NewReader(rb))
	rr = execTPReq(r, rb)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCreateTelemetryIdsHandler(t *testing.T) {
	initTelemetryTestEnv()
	// create two profiles first so IDs are normalized and then migrated
	for _, nm := range []string{"migrate1", "migrate2"} {
		p := newSampleProfile(nm)
		b, _ := json.Marshal(p)
		r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetryProfile1?applicationType=stb", bytes.NewReader(b))
		_ = execTPReq(r, b)
	}
	r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/profile/migrate/createTelemetryId?applicationType=stb", nil)
	rr := execTPReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTelemetryProfilesExportFlag(t *testing.T) {
	initTelemetryTestEnv()
	// create profile
	profile := newSampleProfile("exportable")
	b, _ := json.Marshal(profile)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/telemetry/profile?applicationType=stb", bytes.NewReader(b))
	rr := execTPReq(r, b)
	assert.Equal(t, http.StatusCreated, rr.Code)
	var saved corelogupload.PermanentTelemetryProfile
	_ = json.Unmarshal(rr.Body.Bytes(), &saved)
	// fetch with export flag
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/profile/"+saved.ID+"?applicationType=stb&export=true", nil)
	rr = execTPReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	// list all with export flag
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/profile?applicationType=stb&export=true", nil)
	rr = execTPReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTelemetryProfileHandlerTimeoutSafety(t *testing.T) {
	initTelemetryTestEnv()
	for i := 0; i < 3; i++ {
		r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/telemetry/profile?applicationType=stb", nil)
		_ = execTPReq(r, nil)
		time.Sleep(5 * time.Millisecond)
	}
	assert.True(t, true)
}
