package feature

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	"github.com/rdkcentral/xconfadmin/common"
	oshttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/taggingapi/tag"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	xwrfc "github.com/rdkcentral/xconfwebconfig/shared/rfc"
)

var (
	server *oshttp.WebconfigServer
	router *mux.Router
)

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
	os.Setenv("SECURITY_TOKEN_KEY", "testSecurityTokenKey")
	os.Setenv("SAT_CLIENT_ID", "foo")
	os.Setenv("SAT_CLIENT_SECRET", "bar")
	os.Setenv("IDP_CLIENT_ID", "foo")
	os.Setenv("IDP_CLIENT_SECRET", "bar")

	sc, err := xwcommon.NewServerConfig(cfgFile)
	if err != nil {
		panic(err)
	}
	server = oshttp.NewWebconfigServer(sc, true, nil, nil)
	xwhttp.InitSatTokenManager(server.XW_XconfServer)
	db.SetDatabaseClient(server.XW_XconfServer.DatabaseClient)
	router = server.XW_XconfServer.GetRouter(false)
	dataapi.XconfSetup(server.XW_XconfServer, router)
	featureSetup(server, router)
	if err = server.XW_XconfServer.SetUp(); err != nil {
		panic(err)
	}
	if err = server.XW_XconfServer.TearDown(); err != nil {
		panic(err)
	}

	code := m.Run()
	server.XW_XconfServer.TearDown()
	os.Exit(code)
}
func WebServerInjection(ws *oshttp.WebconfigServer, xc *dataapi.XconfConfigs) {
	if ws == nil {
		common.CacheUpdateWindowSize = 60000
		common.AllowedNumberOfFeatures = 100
		common.ActiveAuthProfiles = "dev"
		common.DefaultAuthProfiles = "prod"
		common.IpMacIsConditionLimit = 20
		common.CanaryCreationEnabled = false
		common.VideoCanaryCreationEnabled = false
		common.AuthProvider = "acl"
		common.ApplicationTypes = []string{"stb"}
		common.WakeupPoolTagName = "t_canary_wakeup"
	} else {
		common.AuthProvider = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.authprovider")
		applicationTypeString := ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.application_types")
		if applicationTypeString == "" {
			applicationTypeString = "stb"
		}
		common.ApplicationTypes = strings.Split(applicationTypeString, ",")
		common.CacheUpdateWindowSize = ws.XW_XconfServer.ServerConfig.GetInt64("xconfwebconfig.xconf.cache_update_window_size")
		common.SatOn = ws.XW_XconfServer.ServerConfig.GetBoolean("xconfwebconfig.sat.SAT_ON")
		common.AllowedNumberOfFeatures = int(ws.XW_XconfServer.ServerConfig.GetInt32("xconfwebconfig.xconf.allowedNumberOfFeatures", 100))
		common.ActiveAuthProfiles = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.authProfilesActive")
		common.DefaultAuthProfiles = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.authProfilesDefault")
		common.IpMacIsConditionLimit = int(ws.XW_XconfServer.ServerConfig.GetInt32("xconfwebconfig.xconf.ipMacIsConditionLimit", 20))
		common.CanaryCreationEnabled = ws.XW_XconfServer.ServerConfig.GetBoolean("xconfwebconfig.xconf.enable_canary_creation")
		common.VideoCanaryCreationEnabled = ws.XW_XconfServer.ServerConfig.GetBoolean("xconfwebconfig.xconf.enable_video_canary_creation")
		common.LockDuration = ws.XW_XconfServer.ServerConfig.GetInt32("xconfwebconfig.xcrp.lock_duration_in_secs", common.DefaultLockDuration)
		if common.CanaryCreationEnabled {
			timezoneStr := ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_time_zone")
			timezone, err := time.LoadLocation(timezoneStr)
			if err != nil {
				log.Errorf("Error loading timezone: %s", timezoneStr)
				panic(err)
			}
			common.CanaryTimezone = timezone
			timezoneListString := ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_timezone_list")
			common.CanaryTimezoneList = strings.Split(timezoneListString, ",")
			common.CanaryStartTime = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_start_time")
			common.CanaryEndTime = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_end_time")
			common.CanaryTimeFormat = ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_time_format")
			common.CanaryDefaultPartner = strings.ToLower(ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_default_partner"))
			common.CanarySize = int(ws.XW_XconfServer.ServerConfig.GetInt64("xconfwebconfig.xconf.canary_size"))
			common.CanaryDistributionPercentage = ws.XW_XconfServer.ServerConfig.GetFloat64("xconfwebconfig.xconf.canary_distribution_percentage")
			common.CanaryFwUpgradeStartTime = int(ws.XW_XconfServer.ServerConfig.GetInt64("xconfwebconfig.xconf.canary_firmware_upgrade_start_time"))
			common.CanaryFwUpgradeEndTime = int(ws.XW_XconfServer.ServerConfig.GetInt64("xconfwebconfig.xconf.canary_firmware_upgrade_end_time"))

			percentFilterNameString := strings.ToLower(ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_percent_filter_name"))
			for _, name := range strings.Split(percentFilterNameString, ";") {
				common.CanaryPercentFilterNameSet.Add(name)
			}

			wakeupPercentFilterNameString := strings.ToLower(ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_wakeup_percent_filter_list"))
			for _, name := range strings.Split(wakeupPercentFilterNameString, ",") {
				common.CanaryWakeupPercentFilterNameSet.Add(name)
			}

			videoModelListString := strings.ToUpper(ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_video_model_list"))
			for _, model := range strings.Split(videoModelListString, ",") {
				common.CanaryVideoModelSet.Add(model)
			}

			syndicatePartnerList := strings.ToLower(ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.canary_appsettings_partner_list"))
			for _, name := range strings.Split(syndicatePartnerList, ",") {
				common.CanarySyndicatePartnerSet.Add(name)
				common.AllAppSettings = append(common.AllAppSettings, (common.PROP_CANARY_FW_UPGRADE_STARTTIME + "_" + name))
				common.AllAppSettings = append(common.AllAppSettings, (common.PROP_CANARY_FW_UPGRADE_ENDTIME + "_" + name))
				common.AllAppSettings = append(common.AllAppSettings, (common.PROP_CANARY_TIMEZONE_LIST + "_" + name))
			}
		}
	}
}
func featureSetup(server *oshttp.WebconfigServer, r *mux.Router) {

	xc := dataapi.GetXconfConfigs(server.XW_XconfServer.ServerConfig.Config)

	WebServerInjection(server, xc)
	db.ConfigInjection(server.XW_XconfServer.ServerConfig.Config)
	dataapi.WebServerInjection(server.XW_XconfServer, xc)
	//dao.WebServerInjection(server)
	auth.WebServerInjection(server)
	dataapi.RegisterTables()

	db.RegisterTableConfigSimple(db.TABLE_TAG, tag.NewTagInf)
	db.GetCacheManager() // Initialize cache manager
	SetupRFCRoutes(server, r)
}

func SetupRFCRoutes(server *oshttp.WebconfigServer, r *mux.Router) {
	// rfc/feature
	rfcFeaturePath := r.PathPrefix("/xconfAdminService/rfc/feature").Subrouter()
	rfcFeaturePath.HandleFunc("", PostFeatureHandler).Methods("POST").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("", PutFeatureHandler).Methods("PUT").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/entities", PostFeatureEntitiesHandler).Methods("POST").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/entities", PutFeatureEntitiesHandler).Methods("PUT").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("", GetFeaturesHandler).Methods("GET").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/{id}", GetFeatureByIdHandler).Methods("GET").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/{id}", DeleteFeatureByIdHandler).Methods("DELETE").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/filtered", GetFeaturesFilteredHandler).Methods("POST").Name("RFC-Feature")
	rfcFeaturePath.HandleFunc("/byIdList", GetFeaturesByIdListHandler).Methods("POST").Name("RFC-Feature")
	// paths variable removed (not needed)

}
func buildFeatureEntity(appType string) *xwrfc.FeatureEntity {
	fe := &xwrfc.FeatureEntity{}
	fe.ID = uuid.NewString()
	fe.ApplicationType = appType
	fe.Name = "Name_" + fe.ID[:8]
	fe.FeatureName = "Feat_" + fe.ID[:8]
	fe.FeatureInstance = "inst" + fe.ID[:4]
	fe.ConfigData = map[string]string{"k": "v"}
	fe.Enable = true
	return fe
}

func TestGetFeaturesEmptyAndExport(t *testing.T) {
	cleanDB()
	r := httptest.NewRequest(http.MethodGet, "/xconfAdminService/rfc/feature?applicationType=stb", nil)
	rr := executeRequest(r)
	assert.Equal(t, http.StatusOK, rr.Code)
	// export empty
	r = httptest.NewRequest(http.MethodGet, "/xconfAdminService/rfc/feature?applicationType=stb&export=true", nil)
	rr = executeRequest(r)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestPostFeatureSuccessAndConflicts(t *testing.T) {
	cleanDB()
	fe := buildFeatureEntity("stb")
	b, _ := json.Marshal(fe)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/rfc/feature?applicationType=stb", bytes.NewReader(b))
	rr := executeRequest(r)
	assert.Equal(t, http.StatusCreated, rr.Code)
	// conflict same id
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/rfc/feature?applicationType=stb", bytes.NewReader(b))
	rr = executeRequest(r)
	assert.Equal(t, http.StatusConflict, rr.Code)
	// applicationType mismatch
	fe.ApplicationType = "wrong"
	b, _ = json.Marshal(fe)
	r = httptest.NewRequest(http.MethodPost, "/xconfAdminService/rfc/feature?applicationType=stb", bytes.NewReader(b))
	rr = executeRequest(r)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestGetFeatureByIdSuccessExportAndNotFound(t *testing.T) {
	cleanDB()
	fe := buildFeatureEntity("stb")
	_, _ = FeaturePost(fe.CreateFeature())
	url := fmt.Sprintf("/xconfAdminService/rfc/feature/%s?applicationType=stb", fe.ID)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	rr := executeRequest(r)
	assert.Equal(t, http.StatusOK, rr.Code)
	// export
	url = fmt.Sprintf("/xconfAdminService/rfc/feature/%s?applicationType=stb&export=true", fe.ID)
	r = httptest.NewRequest(http.MethodGet, url, nil)
	rr = executeRequest(r)
	assert.Equal(t, http.StatusOK, rr.Code)
	// not found
	url = fmt.Sprintf("/xconfAdminService/rfc/feature/%s?applicationType=stb", uuid.NewString())
	r = httptest.NewRequest(http.MethodGet, url, nil)
	rr = executeRequest(r)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestPutFeatureSuccessAndNotFound(t *testing.T) {
	cleanDB()
	fe := buildFeatureEntity("stb")
	_, _ = FeaturePost(fe.CreateFeature())
	fe.ConfigData["extra"] = "123"
	b, _ := json.Marshal(fe)
	r := httptest.NewRequest(http.MethodPut, "/xconfAdminService/rfc/feature?applicationType=stb", bytes.NewReader(b))
	rr := executeRequest(r)
	assert.Equal(t, http.StatusOK, rr.Code)
	// not found different id
	fe2 := buildFeatureEntity("stb")
	b2, _ := json.Marshal(fe2)
	r = httptest.NewRequest(http.MethodPut, "/xconfAdminService/rfc/feature?applicationType=stb", bytes.NewReader(b2))
	rr = executeRequest(r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeleteFeatureByIdSuccessAndNotFound(t *testing.T) {
	cleanDB()
	fe := buildFeatureEntity("stb")
	_, _ = FeaturePost(fe.CreateFeature())
	url := fmt.Sprintf("/xconfAdminService/rfc/feature/%s?applicationType=stb", fe.ID)
	r := httptest.NewRequest(http.MethodDelete, url, nil)
	rr := executeRequest(r)
	assert.Equal(t, http.StatusNoContent, rr.Code)
	url = fmt.Sprintf("/xconfAdminService/rfc/feature/%s?applicationType=stb", fe.ID)
	r = httptest.NewRequest(http.MethodDelete, url, nil)
	rr = executeRequest(r)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetFeaturesFilteredPagingAndInvalid(t *testing.T) {
	cleanDB()
	for i := 0; i < 12; i++ {
		fe := buildFeatureEntity("stb")
		_, _ = FeaturePost(fe.CreateFeature())
	}
	// valid filtered paging request requires pageNumber & pageSize query params
	body := map[string]string{}
	b, _ := json.Marshal(body)
	url := "/xconfAdminService/rfc/feature/filtered?pageNumber=2&pageSize=5&applicationType=stb"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(b))
	GetFeaturesFilteredHandler(xw, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	// header numberOfItems should equal total features (12)
	var numberHeader string
	for k, v := range rr.Header() {
		if strings.EqualFold(k, "numberOfItems") && len(v) > 0 {
			numberHeader = v[0]
			break
		}
	}
	assert.Equal(t, "12", numberHeader)
	// invalid params: omit pageNumber/pageSize to trigger 400
	url = "/xconfAdminService/rfc/feature/filtered?applicationType=stb"
	req = httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rr = httptest.NewRecorder()
	xw = xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(b))
	GetFeaturesFilteredHandler(xw, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestPostAndPutFeatureEntities(t *testing.T) {
	cleanDB()
	// prepare list ensuring unique FeatureName/FeatureInstance across entities
	fe1 := buildFeatureEntity("stb")
	fe2 := buildFeatureEntity("stb")
	fe2.FeatureName = fe2.FeatureName + "_X"
	fe2.FeatureInstance = fe2.FeatureInstance + "_Y"
	list := []*xwrfc.FeatureEntity{fe1, fe2}
	b, _ := json.Marshal(list)
	// direct handler invocation with XResponseWriter to ensure body extraction
	postUrl := "/xconfAdminService/rfc/feature/entities?applicationType=stb"
	postReq := httptest.NewRequest(http.MethodPost, postUrl, bytes.NewReader(b))
	postRR := httptest.NewRecorder()
	postXW := xwhttp.NewXResponseWriter(postRR)
	postXW.SetBody(string(b))
	PostFeatureEntitiesHandler(postXW, postReq)
	assert.Equal(t, http.StatusOK, postRR.Code)
	// update second entity config retains uniqueness
	fe2.ConfigData["k2"] = "v2"
	b, _ = json.Marshal(list)
	putUrl := "/xconfAdminService/rfc/feature/entities?applicationType=stb"
	putReq := httptest.NewRequest(http.MethodPut, putUrl, bytes.NewReader(b))
	putRR := httptest.NewRecorder()
	putXW := xwhttp.NewXResponseWriter(putRR)
	putXW.SetBody(string(b))
	PutFeatureEntitiesHandler(putXW, putReq)
	assert.Equal(t, http.StatusOK, putRR.Code)
}

func TestGetFeaturesByIdList(t *testing.T) {
	cleanDB()
	fe1 := buildFeatureEntity("stb")
	fe2 := buildFeatureEntity("stb")
	_, _ = FeaturePost(fe1.CreateFeature())
	_, _ = FeaturePost(fe2.CreateFeature())
	ids := []string{fe1.ID, fe2.ID}
	b, _ := json.Marshal(ids)
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/rfc/feature/byIdList?applicationType=stb", bytes.NewReader(b))
	rr := executeRequest(r)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// helpers
func executeRequest(r *http.Request) *httptest.ResponseRecorder {
	// Wrap with XResponseWriter so handlers that cast can read drained body
	baseRR := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(baseRR)
	if r.Body != nil {
		// read body bytes to set into XResponseWriter for JSON extract handlers
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
		xw.SetBody(buf.String())
	}
	router.ServeHTTP(xw, r)
	return baseRR
}

func cleanDB() {
	for _, ti := range db.GetAllTableInfo() {
		c := db.GetDatabaseClient().(*db.CassandraClient)
		_ = c.DeleteAllXconfData(ti.TableName)
		if ti.CacheData {
			db.GetCachedSimpleDao().RefreshAll(ti.TableName)
		}
	}
}
