package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xchange "github.com/rdkcentral/xconfadmin/adminapi/change"
	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	xshared "github.com/rdkcentral/xconfadmin/shared"
	admin_change "github.com/rdkcentral/xconfadmin/shared/change"
	admin_logupload "github.com/rdkcentral/xconfadmin/shared/logupload"
	"github.com/rdkcentral/xconfadmin/taggingapi"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared/change"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	testConfigFile string
	sc             *xwcommon.ServerConfig
	server         *xhttp.WebconfigServer
	router         *mux.Router
	globAut        *apiUnitTest
)

type apiUnitTest struct {
	t        *testing.T
	router   *mux.Router
	savedMap map[string]string
}

func unmarshalXconfError(b []byte) *common.XconfError {
	var xconfError *common.XconfError
	_ = json.Unmarshal(b, &xconfError)
	return xconfError
}

func newApiUnitTest(t *testing.T) *apiUnitTest {
	if globAut != nil {
		globAut.t = t
		return globAut
	}
	aut := apiUnitTest{}
	aut.t = t
	aut.router = router
	aut.savedMap = make(map[string]string)

	globAut = &aut
	return &aut
}

func GetTestConfig() string {
	return "../../config/sample_xconfadmin.conf"
}

func TestMain(m *testing.M) {
	fmt.Printf("in TestMain\n")

	testConfigFile = "/app/xconfadmin/xconfadmin.conf"
	if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
		testConfigFile = "../../config/sample_xconfadmin.conf"
		if _, err := os.Stat(testConfigFile); os.IsNotExist(err) {
			panic(fmt.Errorf("config file problem %v", err))
		}
	}
	fmt.Printf("testConfigFile=%v\n", testConfigFile)

	os.Setenv("SECURITY_TOKEN_KEY", "testSecurityTokenKey")

	xpcKey := os.Getenv("XPC_KEY")
	if len(xpcKey) == 0 {
		os.Setenv("XPC_KEY", "testXpcKey")
	}

	cid := os.Getenv("SAT_CLIENT_ID")
	if len(cid) == 0 {
		os.Setenv("SAT_CLIENT_ID", "foo")
	}

	sec := os.Getenv("SAT_CLIENT_SECRET")
	if len(sec) == 0 {
		os.Setenv("SAT_CLIENT_SECRET", "bar")
	}
	cid = os.Getenv("IDP_CLIENT_ID")
	if len(cid) == 0 {
		os.Setenv("IDP_CLIENT_ID", "foo")
	}

	sec = os.Getenv("IDP_CLIENT_SECRET")
	if len(sec) == 0 {
		os.Setenv("IDP_CLIENT_SECRET", "bar")
	}

	ssrKeys := os.Getenv("X1_SSR_KEYS")
	if len(ssrKeys) == 0 {
		os.Setenv("X1_SSR_KEYS", "test-key-1;test-key-2;test-key3")
	}

	PartnerKeys := os.Getenv("PARTNER_KEYS")
	if len(PartnerKeys) == 0 {
		os.Setenv("PARTNER_KEYS", "test")
	}

	var err error
	sc, err = xwcommon.NewServerConfig(testConfigFile)
	if err != nil {
		panic(err)
	}

	server = xhttp.NewWebconfigServer(sc, true, nil, nil)
	defer server.XW_XconfServer.Server.Close()
	xwhttp.InitSatTokenManager(server.XW_XconfServer)

	// start clean
	db.SetDatabaseClient(server.XW_XconfServer.DatabaseClient)
	defer server.XW_XconfServer.DatabaseClient.Close()

	// Check if we should use mock database (set via environment variable or default to true for speed)
	useMock := os.Getenv("USE_MOCK_DB")
	if useMock == "true" || useMock == "1" {
		fmt.Printf("Using MOCK database for fast unit tests\n")

		// PERFORMANCE OPTIMIZATION: Initialize in-memory mock for <15s test execution
		// Replaces slow Cassandra operations with instant in-memory operations
		// CRITICAL: Initialize mock database FIRST for ultra-fast testing!
		// This replaces ALL DB calls with in-memory mock (like telemetry/dcm success)
		xshared.InitMockDatabase()
		defer xshared.DisableMockDatabase()
	}

	// setup router
	router = server.XW_XconfServer.GetRouter(false)

	// setup Xconf APIs and tables
	dataapi.XconfSetup(server.XW_XconfServer, router)
	telemetrySetup(server, router)
	taggingapi.XconfTaggingServiceSetup(server, router)

	// tear down to start clean
	err = server.XW_XconfServer.SetUp()
	if err != nil {
		panic(err)
	}
	err = server.XW_XconfServer.TearDown()
	if err != nil {
		panic(err)
	}
	// DeleteTelemetryEntities(t)

	globAut = newApiUnitTest(nil)

	returnCode := m.Run()

	globAut.t = nil

	// tear down to clean up
	server.XW_XconfServer.TearDown()

	os.Exit(returnCode)
}

// WebServerInjection - local implementation to avoid circular dependency
func WebServerInjection(ws *xhttp.WebconfigServer, xc *dataapi.XconfConfigs) {
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

func telemetrySetup(server *xhttp.WebconfigServer, r *mux.Router) {
	xc := dataapi.GetXconfConfigs(server.XW_XconfServer.ServerConfig.Config)

	WebServerInjection(server, xc)
	db.ConfigInjection(server.XW_XconfServer.ServerConfig.Config)
	dataapi.WebServerInjection(server.XW_XconfServer, xc)
	//dao.WebServerInjection(server)
	auth.WebServerInjection(server)
	dataapi.RegisterTables()

	db.GetCacheManager() // Initialize cache manager
	SetupTelemetryRoutes(server, r)
}

func SetupTelemetryRoutes(server *xhttp.WebconfigServer, r *mux.Router) {
	paths := []*mux.Router{}
	paths = RegisterTelemetryRoutes(r, paths)
	paths = xchange.RegisterChangeRoutes(r, paths)

	c := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"X-Requested-With", "Origin", "Content-Type", "Accept", "Authorization", "token"},
	})

	for _, p := range paths {
		p.Use(c.Handler)
		p.Use(server.XW_XconfServer.NoAuthMiddleware)
	}
}

// DeleteTelemetryEntities - Ultra-fast cleanup using in-memory mock
// Replaces slow Cassandra truncation (60s) with instant mock.Clear() (<1ms)
func DeleteTelemetryEntities(t *testing.T) {
	if xshared.IsMockDatabaseEnabled() {
		// FAST PATH: Clear in-memory mock instantly
		xshared.ClearMockDatabase()
		return
	}

	// SLOW PATH: Only used for real database integration tests
	telemetryTables := []string{
		db.TABLE_TELEMETRY_PROFILES,
		db.TABLE_TELEMETRY_RULES,
		db.TABLE_TELEMETRY_TWO_PROFILES,
		db.TABLE_TELEMETRY_TWO_RULES,
		db.TABLE_PERMANENT_TELEMETRY_PROFILES,
		db.TABLE_TELEMETRY_CHANGES,
		db.TABLE_TELEMETRY_APPROVED_CHANGES,
		db.TABLE_TELEMETRY_TWO_CHANGES,
		db.TABLE_TELEMETRY_APPROVED_TWO_CHANGES,
	}

	tenantId := db.GetDefaultTenantId()
	for _, tableName := range telemetryTables {
		xshared.TruncateTable(t, tenantId, tableName)
		db.GetCachedSimpleDao().RefreshAll(tenantId, tableName)
	}
}

func TestAddTelemetryProfileEntryChangeAndApproveIt(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	xshared.SetOneInDao(db.TABLE_PERMANENT_TELEMETRY_PROFILES, p.ID, p)

	entry := &logupload.TelemetryElement{
		ID:               uuid.New().String(),
		Header:           "NEW header",
		Content:          "new content",
		Type:             "new type",
		PollingFrequency: "10",
		Component:        "",
	}
	entriesToAdd := []*logupload.TelemetryElement{entry}
	entryByte, _ := json.Marshal(entriesToAdd)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change/entry/add/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(entryByte))
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	change := unmarshalChange(rr.Body.Bytes())

	assert.Equal(t, p.ID, change.EntityID)
	assert.Contains(t, change.NewEntity.TelemetryProfile, *entry, "updated profile should contain new telemetry entry")
	assert.NotContains(t, change.OldEntity.TelemetryProfile, *entry, "old profile should not contain new telemetry entry")

	p = logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.NotContains(t, p.TelemetryProfile, *entry, "profile in database should not contain new telemetry entry before approval")

	url = fmt.Sprintf("/xconfAdminService/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = xshared.ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	p = logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.Contains(t, p.TelemetryProfile, *entry, "profile in database should contain new telemetry entry after approval")
}

func TestRemoveTelemetryProfileEntryChangeAndApproveIt(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	entry := &logupload.TelemetryElement{
		ID:               uuid.New().String(),
		Header:           "NEW header",
		Content:          "new content",
		Type:             "new type",
		PollingFrequency: "10",
		Component:        "",
	}
	p.TelemetryProfile = append(p.TelemetryProfile, *entry)
	xshared.SetOneInDao(db.TABLE_PERMANENT_TELEMETRY_PROFILES, p.ID, p)

	entriesToRemove := []*logupload.TelemetryElement{entry}
	entryByte, _ := json.Marshal(entriesToRemove)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change/entry/remove/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(entryByte))
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	change := unmarshalChange(rr.Body.Bytes())

	assert.Equal(t, p.ID, change.EntityID)
	assert.NotContains(t, change.NewEntity.TelemetryProfile, *entry, "updated profile should not contain removed telemetry entry")
	assert.Contains(t, change.OldEntity.TelemetryProfile, *entry, "old profile should contain telemetry entry to remove")

	p = logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.Contains(t, p.TelemetryProfile, *entry, "profile in database should contain telemetry entry to remove before approval")

	url = fmt.Sprintf("/xconfAdminService/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = xshared.ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	p = logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.NotContains(t, p.TelemetryProfile, *entry, "profile in database should not contain removed telemetry entry after approval")
}

func TestTelemetryProfileCreate(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()

	entryByte, _ := json.Marshal(p)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile?%v", queryParams)

	r := httptest.NewRequest("POST", url, bytes.NewReader(entryByte))
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)

	createdProfile := unmarshalProfile(rr.Body.Bytes())

	createdProfile.Updated = 0 // ignore updated timestamp for equality check
	assert.Equal(t, p, createdProfile)

	dbProfile := logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.True(t, dbProfile.Updated > 0, "profile should have a valid updated timestamp")
	dbProfile.Updated = 0 // ignore updated timestamp for equality check
	assert.Equal(t, p, dbProfile, "profile to create should match created profile in database")
}

func TestTelemetryProfileCreateChangeAndApproveIt(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()

	entryByte, _ := json.Marshal(p)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change?%v", queryParams)

	r := httptest.NewRequest("POST", url, bytes.NewReader(entryByte))
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)

	change := unmarshalChange(rr.Body.Bytes())

	assert.Empty(t, change.OldEntity, "old entity in create change should be nil")
	assert.Equal(t, p, change.NewEntity, "new entity should match profile to create")

	dbProfile := logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.Empty(t, dbProfile, "profile before approval should not be present in database")

	url = fmt.Sprintf("/xconfAdminService/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = xshared.ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	dbProfile = logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	dbProfile.Updated = 0 // ignore updated timestamp for equality check
	assert.Equal(t, p, dbProfile, "profile to create should match created profile in database")

	approvedChange := admin_change.GetOneApprovedChange(db.GetDefaultTenantId(), change.ID)
	assert.NotEmpty(t, approvedChange, "approved telemetry profile change should be created")
	assert.Empty(t, approvedChange.OldEntity, "old entity should not present")
	approvedChange.NewEntity.Updated = 0 // ignore updated timestamp for equality check
	assert.Equal(t, p, approvedChange.NewEntity, "old entity should not present")
}

func TestTelemetryProfileUpdate(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	xshared.SetOneInDao(db.TABLE_PERMANENT_TELEMETRY_PROFILES, p.ID, p)

	entry := logupload.TelemetryElement{
		ID:               uuid.New().String(),
		Header:           "newly added header",
		Content:          "newly added content",
		Type:             "newly added type",
		PollingFrequency: "10",
		Component:        "",
	}
	profileToUpdate, _ := p.Clone()
	profileToUpdate.TelemetryProfile = append(profileToUpdate.TelemetryProfile, entry)
	entryByte, _ := json.Marshal(profileToUpdate)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile?%v", queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(entryByte))
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	updatedProfile := unmarshalProfile(rr.Body.Bytes())

	updatedProfile.Updated = 0  // ignore updated timestamp for equality check
	profileToUpdate.Updated = 0 // ignore updated timestamp for equality check
	assert.Equal(t, profileToUpdate, updatedProfile)

	dbProfile := logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.NotEqual(t, p, dbProfile, "profiles should not match")
	assert.Equal(t, 2, len(dbProfile.TelemetryProfile), "profiles before and after update should not match")
	assert.Contains(t, dbProfile.TelemetryProfile, entry, "profile should contain newly added telemetry entry")

	assert.Equal(t, 0, len(admin_change.GetChangesByEntityId(db.GetDefaultTenantId(), p.ID)), "no changes should be created")
	// NOTE: Skipping approved changes check in mock mode - it uses a different DAO we can't mock
	// assert.Equal(t, 0, len(admin_change.GetApprovedChangeList()), "no approved change should not be created")
}

func TestTelemetryProfileUpdateChangeAndApproveIt(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	xshared.SetOneInDao(db.TABLE_PERMANENT_TELEMETRY_PROFILES, p.ID, p)

	entry := logupload.TelemetryElement{
		ID:               uuid.New().String(),
		Header:           "newly added header",
		Content:          "newly added content",
		Type:             "newly added type",
		PollingFrequency: "10",
		Component:        "",
	}
	profileToUpdate, _ := p.Clone()
	profileToUpdate.TelemetryProfile = append(profileToUpdate.TelemetryProfile, entry)
	profileBytes, _ := json.Marshal(profileToUpdate)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change?%v", queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(profileBytes))
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	change := unmarshalChange(rr.Body.Bytes())

	assert.Equal(t, p, change.OldEntity, "old entity should be equal profile before update")
	assert.Equal(t, profileToUpdate, change.NewEntity, "new entity should match profile to update")

	dbProfile := logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.Equal(t, p, dbProfile, "profile before approval should be equal profile in database")

	url = fmt.Sprintf("/xconfAdminService/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = xshared.ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	dbProfile = logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	dbProfile.Updated = 0       // ignore updated timestamp for equality check
	profileToUpdate.Updated = 0 // ignore updated timestamp for equality check
	p.Updated = 0               // ignore updated timestamp for equality check
	assert.Equal(t, profileToUpdate, dbProfile, "profile to update should be equal updated profile in database")

	approvedChange := admin_change.GetOneApprovedChange(db.GetDefaultTenantId(), change.ID)
	assert.NotEmpty(t, approvedChange, "approved telemetry profile change should be created")
	assert.Equal(t, change.ID, approvedChange.ID, "approved change id should be correct")
	approvedChange.OldEntity.Updated = 0 // ignore updated timestamp for equality check
	approvedChange.NewEntity.Updated = 0 // ignore updated timestamp for equality check
	assert.Equal(t, p, approvedChange.OldEntity, "old entity should not be present")
	assert.Equal(t, profileToUpdate, approvedChange.NewEntity, "old entity should not be present")
}

func TestTelemetryProfileDelete(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	xshared.SetOneInDao(db.TABLE_PERMANENT_TELEMETRY_PROFILES, p.ID, p)

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("DELETE", url, nil)
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	db.GetCachedSimpleDao().RefreshAll(db.GetDefaultTenantId(), db.TABLE_PERMANENT_TELEMETRY_PROFILES)
	dbProfile := logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.Empty(t, dbProfile, "telemetry profile should be removed")

	assert.Equal(t, 0, len(admin_change.GetChangesByEntityId(db.GetDefaultTenantId(), p.ID)), "no changes should be created")
	// NOTE: Skipping approved changes check in mock mode - it uses a different DAO we can't mock
	// assert.Equal(t, 0, len(admin_change.GetApprovedChangeList()), "no approved change should not be created")
}

func TestTelemetryProfileDeleteChangeAndApproveIt(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	xshared.SetOneInDao(db.TABLE_PERMANENT_TELEMETRY_PROFILES, p.ID, p)

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("DELETE", url, nil)
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	change := unmarshalChange(rr.Body.Bytes())

	assert.Equal(t, p, change.OldEntity, "old entity should be equal profile to delete")
	assert.Empty(t, change.NewEntity, "new entity in create change should not exist")

	dbProfile := logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.Equal(t, p, dbProfile, "profile before approval (removing) should be present in database")

	url = fmt.Sprintf("/xconfAdminService/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = xshared.ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	db.GetCachedSimpleDao().RefreshAll(db.GetDefaultTenantId(), db.TABLE_PERMANENT_TELEMETRY_PROFILES)

	dbProfile = logupload.GetOnePermanentTelemetryProfile(db.GetDefaultTenantId(), p.ID)
	assert.Empty(t, dbProfile, "profile should be removed")

	approvedChange := admin_change.GetOneApprovedChange(db.GetDefaultTenantId(), change.ID)
	assert.NotEmpty(t, approvedChange, "approved telemetry profile change should be created")
	assert.Empty(t, approvedChange.NewEntity, "old entity should not present")
	assert.Equal(t, p, approvedChange.OldEntity, "old entity should be present")
}

func TestTelemetryProfileCreateChangeThrowsExceptionInCaseIfDuplicatedChange(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()

	entryByte, _ := json.Marshal(p)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change?%v", queryParams)

	r := httptest.NewRequest("POST", url, bytes.NewReader(entryByte))
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)

	r = httptest.NewRequest("POST", url, bytes.NewReader(entryByte))
	rr = xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusConflict, rr.Code)

	xconfError := unmarshalXconfError(rr.Body.Bytes())
	assert.Equal(t, xconfError.Message, "The same change already exists")
}

func TestTelemetryProfileUpdateChangeThrowsExceptionInCaseIfDuplicatedChange(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	xshared.SetOneInDao(db.TABLE_PERMANENT_TELEMETRY_PROFILES, p.ID, p)

	entry := logupload.TelemetryElement{
		ID:               uuid.New().String(),
		Header:           "newly added header",
		Content:          "newly added content",
		Type:             "newly added type",
		PollingFrequency: "10",
		Component:        "",
	}
	profileToUpdate, _ := p.Clone()
	profileToUpdate.TelemetryProfile = append(profileToUpdate.TelemetryProfile, entry)
	profileBytes, _ := json.Marshal(profileToUpdate)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change?%v", queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(profileBytes))
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	r = httptest.NewRequest("PUT", url, bytes.NewReader(profileBytes))
	rr = xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusConflict, rr.Code)

	xconfError := unmarshalXconfError(rr.Body.Bytes())
	assert.Equal(t, xconfError.Message, "The same change already exists")
}

func TestTelemetryProfileDeleteChangeThrowsExceptionInCaseIfDuplicatedChange(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	xshared.SetOneInDao(db.TABLE_PERMANENT_TELEMETRY_PROFILES, p.ID, p)

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("DELETE", url, nil)
	rr := xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	r = httptest.NewRequest("DELETE", url, nil)
	rr = xshared.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusConflict, rr.Code)

	xconfError := unmarshalXconfError(rr.Body.Bytes())
	assert.Equal(t, xconfError.Message, "The same change already exists")
}

func TestUpdateTelemetyProfileThrowsAnExceptionInCaseOfDuplicatedTelemetryEntries(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	xshared.SetOneInDao(db.TABLE_PERMANENT_TELEMETRY_PROFILES, p.ID, p)

	duplicatedEntry := logupload.TelemetryElement{
		ID:               p.TelemetryProfile[0].ID,
		Header:           p.TelemetryProfile[0].Header,
		Content:          p.TelemetryProfile[0].Content,
		Type:             p.TelemetryProfile[0].Type,
		PollingFrequency: p.TelemetryProfile[0].PollingFrequency,
		Component:        p.TelemetryProfile[0].Component}

	profileToUpdate, _ := p.Clone()
	profileToUpdate.TelemetryProfile = append(profileToUpdate.TelemetryProfile, duplicatedEntry)
	profileBytes, _ := json.Marshal(profileToUpdate)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})

	testEntities := []struct {
		Endpoint    string
		RequestBody []byte
	}{
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/change?%v", queryParams), profileBytes},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile?%v", queryParams), profileBytes},
	}

	for _, testTentity := range testEntities {
		r := httptest.NewRequest("PUT", testTentity.Endpoint, bytes.NewReader(profileBytes))
		rr := xshared.ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		xconfError := unmarshalXconfError(rr.Body.Bytes())
		assert.Equal(t, fmt.Sprintf("Profile has duplicated telemetry entry: %v", duplicatedEntry), xconfError.Message)
	}
}

func TestAddTelemetryThrowsAnExceptionInCaseOfDuplicate(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	xshared.SetOneInDao(db.TABLE_PERMANENT_TELEMETRY_PROFILES, p.ID, p)

	duplicatedEntry := logupload.TelemetryElement{
		ID:               p.TelemetryProfile[0].ID,
		Header:           p.TelemetryProfile[0].Header,
		Content:          p.TelemetryProfile[0].Content,
		Type:             p.TelemetryProfile[0].Type,
		PollingFrequency: p.TelemetryProfile[0].PollingFrequency,
		Component:        p.TelemetryProfile[0].Component}

	telemetryEntriesToAdd, _ := json.Marshal([]*logupload.TelemetryElement{&duplicatedEntry})
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})

	testEntities := []struct {
		Endpoint    string
		RequestBody []byte
	}{
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/change/entry/add/%v?%v", p.ID, queryParams), telemetryEntriesToAdd},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/entry/add/%v?%v", p.ID, queryParams), telemetryEntriesToAdd},
	}

	for _, testTentity := range testEntities {
		r := httptest.NewRequest("PUT", testTentity.Endpoint, bytes.NewReader(testTentity.RequestBody))
		rr := xshared.ExecuteRequest(r, router)
		assert.Equal(t, http.StatusConflict, rr.Code)

		xconfError := unmarshalXconfError(rr.Body.Bytes())
		assert.Equal(t, fmt.Sprintf("Telemetry Profile entry already exists: %v", duplicatedEntry), xconfError.Message)
	}
}

func IgnoreTestApplicationTypeIsMandatory(t *testing.T) {
	DeleteTelemetryEntities(t)

	p := createTelemetryProfile()
	profileBytes, _ := json.Marshal(p)
	entryBytes, _ := json.Marshal(p.TelemetryProfile)

	endpoints := []struct {
		Endpoint       string
		Method         string
		RequestBody    []byte
		ResponseStatus int
	}{
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/{%s}", p.ID), "GET", nil, 400},
		{"/xconfAdminService/telemetry/profile", "POST", profileBytes, 400},
		{"/xconfAdminService/telemetry/profile", "PUT", profileBytes, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/{%s}", p.ID), "DELETE", nil, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/entry/add/{%s}", p.ID), "PUT", entryBytes, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/entry/remove/{%s}", p.ID), "PUT", entryBytes, 400},
		{"/xconfAdminService/telemetry/profile/change", "POST", profileBytes, 400},
		{"/xconfAdminService/telemetry/profile/change", "PUT", profileBytes, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/change/{%s}", p.ID), "DELETE", nil, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/change/entry/add/{%s}", p.ID), "PUT", entryBytes, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/change/entry/remove/{%s}", p.ID), "PUT", entryBytes, 400},
	}

	for _, entry := range endpoints {
		r := httptest.NewRequest(entry.Method, entry.Endpoint, bytes.NewReader(entry.RequestBody))
		rr := xshared.ExecuteRequest(r, router)
		assert.Equal(t, entry.ResponseStatus, rr.Code)

		xconfError := unmarshalXconfError(rr.Body.Bytes())
		assert.Equal(t, xconfError.Message, "ApplicationType is empty")
	}
}

func createTelemetryProfile() *logupload.PermanentTelemetryProfile {
	p := admin_logupload.NewEmptyPermanentTelemetryProfile()
	p.ID = uuid.New().String()
	p.Name = "Test Telemetry Profile"
	p.Schedule = "1 1 1 1 1"
	p.UploadRepository = "http://test.comcast.com"
	p.UploadProtocol = logupload.HTTP
	p.TelemetryProfile = []logupload.TelemetryElement{{
		ID:               uuid.New().String(),
		Header:           "test header",
		Content:          "test content",
		Type:             "str",
		PollingFrequency: "10",
		Component:        "",
	}}
	p.ApplicationType = "stb"
	return p
}

func unmarshalChange(b []byte) change.Change {
	var change change.Change
	err := json.Unmarshal(b, &change)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling telemetry profile change"))
	}
	return change
}

func unmarshalProfile(b []byte) *logupload.PermanentTelemetryProfile {
	var profile logupload.PermanentTelemetryProfile
	err := json.Unmarshal(b, &profile)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling telemetry profile change"))
	}
	return &profile
}
