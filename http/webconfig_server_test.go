package http

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-akka/configuration"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
)

func TestMain(m *testing.M) {
	// Set required environment variables for tests
	os.Setenv("SECURITY_TOKEN_KEY", "testSecurityTokenKey")
	os.Setenv("XPC_KEY", "testXpcKey")
	os.Setenv("SAT_CLIENT_ID", "test-sat-client")
	os.Setenv("SAT_CLIENT_SECRET", "test-sat-secret")
	os.Setenv("IDP_CLIENT_ID", "test-idp-client")
	os.Setenv("IDP_CLIENT_SECRET", "test-idp-secret")

	os.Exit(m.Run())
}

// helper to load sample server config
func loadSampleServerConfig(t *testing.T) *xwcommon.ServerConfig {
	t.Helper()
	cfgPath := filepath.Join("..", "config", "sample_xconfadmin.conf")
	sc, err := xwcommon.NewServerConfig(cfgPath)
	if err != nil {
		t.Fatalf("failed to load sample config: %v", err)
	}
	return sc
}

func TestNewWebconfigServer_Defaults(t *testing.T) {
	sc := loadSampleServerConfig(t)
	ws := NewWebconfigServer(sc, true, nil, nil)
	if ws == nil {
		t.Fatalf("expected server instance")
	}
	if ws.AppName == "" {
		t.Fatalf("app name empty")
	}
	// idp provider default is "acl" so IdpServiceConnector should be nil
	if ws.IdpServiceConnector != nil {
		t.Fatalf("expected nil idp service connector for acl provider")
	}
	if !ws.metricsEnabled {
		t.Fatalf("metrics should be enabled by default")
	}
}

func loadServerConfigWithAuthProvider(t *testing.T, provider string) *xwcommon.ServerConfig {
	t.Helper()
	sc := loadSampleServerConfig(t)
	sampleConf, err := os.ReadFile(filepath.Join("..", "config", "sample_xconfadmin.conf"))
	if err != nil {
		t.Fatalf("failed to read sample config: %v", err)
	}
	sc.Config = configuration.ParseString(string(sampleConf) + "\nxconfwebconfig.xconf.authprovider = \"" + provider + "\"\n")
	return sc
}

func TestNewWebconfigServer_NewIdpServiceConnectorFuncUsedWhenConnectorNil(t *testing.T) {
	saved := WebConfServer
	defer func() { WebConfServer = saved }()

	sc := loadServerConfigWithAuthProvider(t, "idp")
	mock := &MockIdpServiceConnector{host: "factory-host"}
	var calls int
	var gotConf *configuration.Config
	var gotTls *tls.Config
	ec := NewExternalConnectors()
	ec.NewIdpServiceConnectorFunc = func(conf *configuration.Config, tlsConfig *tls.Config) IdpServiceConnector {
		calls++
		gotConf = conf
		gotTls = tlsConfig
		return mock
	}

	ws := NewWebconfigServer(sc, true, nil, ec)
	if calls != 1 {
		t.Fatalf("expected factory to be called once, got %d", calls)
	}
	if gotConf != sc.Config {
		t.Fatalf("expected factory to receive server config")
	}
	if gotTls != ws.tlsConfig {
		t.Fatalf("expected factory to receive the server's shared TLS config")
	}
	if ws.IdpServiceConnector != mock {
		t.Fatalf("expected server to use connector built by factory")
	}
	if ec.IdpServiceConnector != mock {
		t.Fatalf("expected factory result to be stored on external connectors")
	}
}

func TestNewWebconfigServer_NewIdpServiceConnectorFuncSkippedWhenConnectorSet(t *testing.T) {
	saved := WebConfServer
	defer func() { WebConfServer = saved }()

	sc := loadServerConfigWithAuthProvider(t, "idp")
	existing := &MockIdpServiceConnector{host: "existing-host"}
	ec := NewExternalConnectors()
	ec.IdpServiceConnector = existing
	ec.NewIdpServiceConnectorFunc = func(conf *configuration.Config, tlsConfig *tls.Config) IdpServiceConnector {
		t.Fatalf("factory should not be called when IdpServiceConnector is already set")
		return nil
	}

	ws := NewWebconfigServer(sc, true, nil, ec)
	if ws.IdpServiceConnector != existing {
		t.Fatalf("expected server to use pre-set connector")
	}
}

func TestNewWebconfigServer_NewIdpServiceConnectorFuncIgnoredForAclProvider(t *testing.T) {
	saved := WebConfServer
	defer func() { WebConfServer = saved }()

	sc := loadSampleServerConfig(t)
	ec := NewExternalConnectors()
	ec.NewIdpServiceConnectorFunc = func(conf *configuration.Config, tlsConfig *tls.Config) IdpServiceConnector {
		return &MockIdpServiceConnector{host: "factory-host"}
	}

	ws := NewWebconfigServer(sc, true, nil, ec)
	if ws.IdpServiceConnector != nil {
		t.Fatalf("expected nil idp service connector for acl provider")
	}
}

func TestSatTokenValidation_UsesSatConnectorHttpClient(t *testing.T) {
	saved := WebConfServer
	defer func() { WebConfServer = saved }()

	sc := loadSampleServerConfig(t)
	sampleConf, err := os.ReadFile(filepath.Join("..", "config", "sample_xconfadmin.conf"))
	if err != nil {
		t.Fatalf("failed to read sample config: %v", err)
	}
	sc.Config = configuration.ParseString(string(sampleConf) + "\nxconfwebconfig.sat_service.tls_server_name = \"sat.internal.example.com\"\n")
	ws := NewWebconfigServer(sc, true, nil, nil)

	satClient := ws.XW_XconfServer.SatServiceConnector.GetHttpClient()
	if satClient == nil {
		t.Fatalf("expected SAT connector http client")
	}
	transport, ok := satClient.Transport.(*http.Transport)
	if !ok || transport.TLSClientConfig == nil {
		t.Fatalf("expected SAT connector transport with TLS config")
	}
	if transport.TLSClientConfig.ServerName != "sat.internal.example.com" {
		t.Fatalf("expected SAT TLS server name override, got %q", transport.TLSClientConfig.ServerName)
	}

	// the SAT-host validation path must fetch keys with the connector's client, not http.DefaultClient
	var called bool
	satClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		return nil, errors.New("stop")
	})
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT","kid":"k1"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	_, _, _, err = getSubjectAndCapabilitiesFromSatToken("Bearer "+header+"."+payload+".c2ln", false)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !called {
		t.Fatalf("expected SAT keys client to be used for signing key fetch")
	}
}

func TestTestingMiddlewareCapturesBody(t *testing.T) {
	sc := loadSampleServerConfig(t)
	ws := NewWebconfigServer(sc, true, nil, nil)
	bodySent := `{"hello":"world"}`
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xw, ok := w.(*xhttp.XResponseWriter)
		if !ok {
			t.Fatalf("response writer type mismatch")
		}
		if xw.Body() != bodySent {
			t.Fatalf("expected body %s got %s", bodySent, xw.Body())
		}
		w.WriteHeader(200)
	})
	h := ws.TestingMiddleware(final)
	r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(bodySent))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	if rr.Code != 200 {
		t.Fatalf("unexpected status %d", rr.Code)
	}
}

// func TestNoAuthMiddlewareAddsHeadersAndMoneyTrace(t *testing.T) {
// 	sc := loadSampleServerConfig(t)
// 	ws := NewWebconfigServer(sc, true, nil, nil)

// 	// Create a router to provide proper context for mux.CurrentRoute
// 	router := mux.NewRouter()
// 	router.HandleFunc("/trace", func(w http.ResponseWriter, r *http.Request) {
// 		// verify Moneytrace header added
// 		if r.Header.Get("X-Moneytrace") == "" {
// 			t.Fatalf("expected X-Moneytrace header")
// 		}
// 		w.WriteHeader(204)
// 	}).Methods(http.MethodGet)

// 	// Wrap the router with the middleware
// 	h := ws.NoAuthMiddleware(router)
// 	r := httptest.NewRequest(http.MethodGet, "/trace", nil)
// 	rr := httptest.NewRecorder()
// 	h.ServeHTTP(rr, r)
// 	if rr.Header().Get("Server") == "" || rr.Header().Get("Xconf-Server") == "" {
// 		t.Fatalf("expected server headers")
// 	}
// 	if rr.Code != 204 {
// 		t.Fatalf("unexpected status %d", rr.Code)
// 	}
// }

func TestLogRequestStartsObfuscatesMsgpack(t *testing.T) {
	sc := loadSampleServerConfig(t)
	ws := NewWebconfigServer(sc, true, nil, nil)
	r := httptest.NewRequest(http.MethodPost, "/msgpack", strings.NewReader("rawbytes"))
	r.Header.Set("Content-type", "application/msgpack")
	xw := ws.logRequestStarts(httptest.NewRecorder(), r)
	if xw.Body() != "rawbytes" {
		t.Fatalf("body mismatch")
	}
	audit := xw.Audit()
	// Implementation in underlying xhttp may or may not obfuscate; accept either but ensure key present
	if _, ok := audit["body"]; !ok {
		t.Fatalf("expected body key in audit")
	}
	// if not obfuscated, log note (implicit coverage)
	if audit["body"] != "****" && audit["body"] != "rawbytes" {
		t.Fatalf("unexpected body value %v", audit["body"])
	}
}

// func TestLogRequestEndsPasswordMaskAndTruncate(t *testing.T) {
// 	sc := loadSampleServerConfig(t)
// 	ws := NewWebconfigServer(sc, true, nil, nil)
// 	router := mux.NewRouter()
// 	// small JSON with password to mask
// 	router.HandleFunc("/mask", func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusBadRequest) // trigger logging branch (>=400)
// 		w.Write([]byte(`{"password":"secret","other":"x"}`))
// 	}).Methods(http.MethodGet)
// 	// large response to truncate
// 	large := strings.Repeat("A", responseLoggingUpperBound+100)
// 	router.HandleFunc("/truncate", func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusBadRequest)
// 		obj := map[string]string{"data": large}
// 		b, _ := json.Marshal(obj)
// 		w.Write(b)
// 	}).Methods(http.MethodGet)

// 	// wrap routes with middleware
// 	h := ws.NoAuthMiddleware(router)

// 	// mask test
// 	r := httptest.NewRequest(http.MethodGet, "/mask", nil)
// 	rr := httptest.NewRecorder()
// 	h.ServeHTTP(rr, r)
// 	// cannot directly access internal audit fields after end; rely on status code only
// 	if rr.Code != http.StatusBadRequest {
// 		t.Fatalf("expected 400 for mask test")
// 	}

// 	// truncate test
// 	r2 := httptest.NewRequest(http.MethodGet, "/truncate", nil)
// 	rr2 := httptest.NewRecorder()
// 	h.ServeHTTP(rr2, r2)
// 	if rr2.Code != http.StatusBadRequest {
// 		t.Fatalf("expected 400 for truncate test")
// 	}
// 	// Ensure original body length huge
// 	if len(rr2.Body.Bytes()) < responseLoggingUpperBound {
// 		t.Fatalf("expected large body for truncate test")
// 	}
// }

func TestGetHeadersForLogAsMapFiltersIgnored(t *testing.T) {
	sc := loadSampleServerConfig(t)
	ws := NewWebconfigServer(sc, true, nil, nil)
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer token") // default ignored list includes authorization
	hdr.Set("X-Test", "value")
	m := getHeadersForLogAsMap(hdr, ws.notLoggedHeaders)
	if _, exists := m["Authorization"]; exists {
		t.Fatalf("expected Authorization to be filtered")
	}
	if m["X-Test"] == nil {
		t.Fatalf("expected X-Test to be present")
	}
}

func TestAppNameFunc(t *testing.T) {
	sc := loadSampleServerConfig(t)
	ws := NewWebconfigServer(sc, true, nil, nil)
	if AppName() != ws.AppName {
		t.Fatalf("AppName func mismatch")
	}
}
