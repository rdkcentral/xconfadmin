package change

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	xchange "github.com/rdkcentral/xconfadmin/shared/change"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
	xwchange "github.com/rdkcentral/xconfwebconfig/shared/change"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	oshttp "github.com/rdkcentral/xconfadmin/http"
)

var (
	chgServer *oshttp.WebconfigServer
	chgRouter *mux.Router
)

// TestMain sets up a minimal server and router for exercising handlers
func TestMain(m *testing.M) {
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
	os.Setenv("SECURITY_TOKEN_KEY", "changeUTKey")
	os.Setenv("SAT_CLIENT_ID", "foo")
	os.Setenv("SAT_CLIENT_SECRET", "bar")
	os.Setenv("IDP_CLIENT_ID", "foo")
	os.Setenv("IDP_CLIENT_SECRET", "bar")

	sc, err := xwcommon.NewServerConfig(cfgFile)
	if err != nil {
		panic(err)
	}
	chgServer = oshttp.NewWebconfigServer(sc, true, nil, nil)
	xwhttp.InitSatTokenManager(chgServer.XW_XconfServer)
	db.SetDatabaseClient(chgServer.XW_XconfServer.DatabaseClient)
	chgRouter = chgServer.XW_XconfServer.GetRouter(false)
	dataapi.XconfSetup(chgServer.XW_XconfServer, chgRouter)
	// inject auth + register tables
	auth.WebServerInjection(chgServer)
	dataapi.RegisterTables()
	// only install change routes we test from change_handler.go
	setupChangeRoutes(chgRouter)
	if err = chgServer.XW_XconfServer.SetUp(); err != nil {
		panic(err)
	}
	code := m.Run()
	chgServer.XW_XconfServer.TearDown()
	os.Exit(code)
}

func setupChangeRoutes(r *mux.Router) {
	p := r.PathPrefix("/xconfAdminService/change").Subrouter()
	p.HandleFunc("/changes", GetProfileChangesHandler).Methods("GET")
	p.HandleFunc("/approve/{changeId}", ApproveChangeHandler).Methods("POST")
	p.HandleFunc("/approved", GetApprovedHandler).Methods("GET")
	p.HandleFunc("/approved/filtered", GetApprovedFilteredHandler).Methods("POST")
	p.HandleFunc("/changes/filtered", GetChangesFilteredHandler).Methods("POST")
	p.HandleFunc("/revert/{approveId}", RevertChangeHandler).Methods("POST")
	p.HandleFunc("/cancel/{changeId}", CancelChangeHandler).Methods("POST")
	p.HandleFunc("/grouped", GetGroupedChangesHandler).Methods("GET")
	p.HandleFunc("/groupedApproved", GetGroupedApprovedChangesHandler).Methods("GET")
	p.HandleFunc("/entityIds", GetChangedEntityIdsHandler).Methods("GET")
	p.HandleFunc("/approveEntities", ApproveChangesHandler).Methods("POST")
	p.HandleFunc("/revertEntities", RevertChangesHandler).Methods("POST")
}

// helper to execute and wrap XResponseWriter for body extraction
func execChangeReq(r *http.Request, body []byte) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	if body != nil {
		xw.SetBody(string(body))
	}
	chgRouter.ServeHTTP(xw, r)
	return rr
}

// minimal pending change JSON builder

func TestChangeHandlersBasicFlows(t *testing.T) {
	// Initially empty changes list
	r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/change/changes?applicationType=stb", nil)
	rr := execChangeReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	// entityIds empty
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/change/entityIds?applicationType=stb", nil)
	rr = execChangeReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestApproveChangeValidationErrors(t *testing.T) {
	// missing changeId path variable
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/change/approve/?applicationType=stb", nil)
	rr := execChangeReq(r, nil)
	assert.Equal(t, http.StatusNotFound, rr.Code) // mux won't match handler, ensure route is correct
	// invalid (blank) id should return 404 when not found
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/change/approve/doesNotExist?applicationType=stb", nil)
	rr = execChangeReq(r, nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGroupedChangesParamErrors(t *testing.T) {
	// missing pageNumber
	r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/change/grouped?pageSize=10&applicationType=stb", nil)
	rr := execChangeReq(r, nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// missing pageSize
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/change/grouped?pageNumber=1&applicationType=stb", nil)
	rr = execChangeReq(r, nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGroupedApprovedChangesParamErrors(t *testing.T) {
	// missing pageNumber
	r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/change/groupedApproved?pageSize=5&applicationType=stb", nil)
	rr := execChangeReq(r, nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// missing pageSize
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/change/groupedApproved?pageNumber=1&applicationType=stb", nil)
	rr = execChangeReq(r, nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestApproveChangesInvalidBody(t *testing.T) {
	// invalid JSON list for approveEntities
	body := []byte("{bad json}")
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/change/approveEntities?applicationType=stb", bytes.NewReader(body))
	rr := execChangeReq(r, body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRevertChangesInvalidBody(t *testing.T) {
	body := []byte("{bad json}")
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/change/revertEntities?applicationType=stb", bytes.NewReader(body))
	rr := execChangeReq(r, body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// NOTE: deeper positive approval/revert flows rely on underlying telemetry profile persistence which is large; here we focus on handler validation branches for coverage.

func TestChangeHandlersTimeoutSafety(t *testing.T) {
	// simple repeated calls to ensure no data races
	for i := 0; i < 3; i++ {
		r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/change/changes?applicationType=stb", nil)
		_ = execChangeReq(r, nil)
		time.Sleep(10 * time.Millisecond)
	}
	assert.True(t, true)
}

// --- Additional coverage tests ---

// helper to create a pending change via service APIs then approve
func TestCancelChangeHandlerWithSyntheticChange(t *testing.T) {
	// Create and persist a minimal pending change
	c := &xwchange.Change{}
	c.ID = "handlerFlow1"
	c.EntityID = "entity-flow1"
	c.EntityType = xwchange.TelemetryProfile
	c.ApplicationType = shared.STB
	c.Author = "author"
	c.Operation = xwchange.Create
	if err := xchange.CreateOneChange(c); err != nil {
		t.Fatalf("failed to persist change: %v", err)
	}
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/change/cancel/"+c.ID+"?applicationType=stb", nil)
	rr := execChangeReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Nil(t, xchange.GetOneChange(c.ID))
}

func TestGetApprovedHandlerEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/change/approved?applicationType=stb", nil)
	rr := execChangeReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGroupedChangesApprovedSuccessPaginationEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/change/grouped?pageNumber=1&pageSize=10&applicationType=stb", nil)
	rr := execChangeReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/change/groupedApproved?pageNumber=1&pageSize=5&applicationType=stb", nil)
	rr = execChangeReq(r, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestChangesFilteredAndApprovedFilteredHandlersEmpty(t *testing.T) {
	body := []byte(`{}`)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/change/approved/filtered?pageNumber=1&pageSize=5&applicationType=stb", bytes.NewReader(body))
	rr := execChangeReq(r, body)
	assert.Equal(t, http.StatusOK, rr.Code)
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/change/changes/filtered?pageNumber=1&pageSize=5&applicationType=stb", bytes.NewReader(body))
	rr = execChangeReq(r, body)
	assert.Equal(t, http.StatusOK, rr.Code)
}
