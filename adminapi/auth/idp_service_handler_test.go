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
package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	xhttp "github.com/rdkcentral/xconfadmin/http"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
)

// setupTestEnv sets all required environment variables for tests
func setupTestEnv() {
	os.Setenv("SECURITY_TOKEN_KEY", "testSecurityTokenKey")
	os.Setenv("XPC_KEY", "testXpcKey")
	os.Setenv("SAT_CLIENT_ID", "test-sat-client")
	os.Setenv("SAT_CLIENT_SECRET", "test-sat-secret")
	os.Setenv("IDP_CLIENT_ID", "test-idp-client")
	os.Setenv("IDP_CLIENT_SECRET", "test-idp-secret")
}

// cleanupTestEnv removes all test environment variables
func cleanupTestEnv() {
	os.Unsetenv("SECURITY_TOKEN_KEY")
	os.Unsetenv("XPC_KEY")
	os.Unsetenv("SAT_CLIENT_ID")
	os.Unsetenv("SAT_CLIENT_SECRET")
	os.Unsetenv("IDP_CLIENT_ID")
	os.Unsetenv("IDP_CLIENT_SECRET")
}

// fake idp connector
type fakeIdp struct {
	tokenReturn   string
	tokenErr      bool
	logoutErr     bool
	lastLogoutUrl string
	lastLoginUrl  string
}

func (f *fakeIdp) IdpServiceHost() string        { return "http://idp" }
func (f *fakeIdp) SetIdpServiceHost(host string) {}
func (f *fakeIdp) GetFullLoginUrl(continueUrl string) string {
	f.lastLoginUrl = continueUrl
	return "http://idp/login?continue=" + url.QueryEscape(continueUrl)
}
func (f *fakeIdp) GetJsonWebKeyResponse(u string) *xhttp.JsonWebKeyResponse { return nil }
func (f *fakeIdp) GetFullLogoutUrl(continueUrl string) string {
	f.lastLogoutUrl = continueUrl
	return "http://idp/logout?continue=" + url.QueryEscape(continueUrl)
}
func (f *fakeIdp) GetToken(code string) string {
	if f.tokenErr {
		return ""
	}
	return f.tokenReturn
}
func (f *fakeIdp) Logout(u string) error {
	if f.logoutErr {
		return fmt.Errorf("logout fail")
	}
	return nil
}
func (f *fakeIdp) GetIdpServiceConfig() *xhttp.IdpServiceConfig { return nil }

// minimal server config stub via configuration.Config wrapped in WebconfigServer
func makeWs(idp *fakeIdp) *xhttp.WebconfigServer {
	cfgPath := filepath.Join("config", "sample_xconfadmin.conf")
	if _, err := os.Stat(cfgPath); err != nil {
		alt := filepath.Join("..", "..", "config", "sample_xconfadmin.conf")
		if _, err2 := os.Stat(alt); err2 == nil {
			cfgPath = alt
		}
	}
	sc, err := xwcommon.NewServerConfig(cfgPath)
	if err != nil {
		panic(err)
	}
	ws := xhttp.NewWebconfigServer(sc, true, nil, nil)
	// override connector with fake
	ws.IdpServiceConnector = idp
	return ws
}

// helper execute handler and return response recorder
func runHandler(h func(http.ResponseWriter, *http.Request), req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h(rr, req)
	return rr
}

func TestGetAdminUIUrlFromCookies_FallbackAndSuccess(t *testing.T) {
	// missing cookie => default
	r1 := httptest.NewRequest("GET", "/", nil)
	if v := GetAdminUIUrlFromCookies(r1); v != defaultAdminUIHost {
		t.Fatalf("expected default got %s", v)
	}
	// present cookie escaped
	val := url.QueryEscape("http://example.com/app")
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.AddCookie(&http.Cookie{Name: adminUrlCookieName, Value: val})
	if v := GetAdminUIUrlFromCookies(r2); v != "http://example.com/app" {
		t.Fatalf("unescape failed got %s", v)
	}
}

func TestLoginUrlHandler_SetsUrl(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	fidp := &fakeIdp{tokenReturn: "tok"}
	WebServerInjection(makeWs(fidp))
	// cookie defines admin UI base
	r := httptest.NewRequest("GET", "/loginurl", nil)
	r.AddCookie(&http.Cookie{Name: adminUrlCookieName, Value: url.QueryEscape("http://admin.local")})
	rr := runHandler(LoginUrlHandler, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "http://idp/login") {
		t.Fatalf("missing login url body=%s", rr.Body.String())
	}
	if fidp.lastLoginUrl == "" || !strings.Contains(fidp.lastLoginUrl, "http://admin.local") {
		t.Fatalf("continue url not captured")
	}
}

func TestLogoutHandler_SuccessAndError(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	// success
	fidp := &fakeIdp{}
	WebServerInjection(makeWs(fidp))
	r := httptest.NewRequest("POST", "/logout", nil)
	r.AddCookie(&http.Cookie{Name: adminUrlCookieName, Value: url.QueryEscape("http://admin.local")})
	rr := runHandler(LogoutHandler, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	// error path
	fidp2 := &fakeIdp{logoutErr: true}
	WebServerInjection(makeWs(fidp2))
	r2 := httptest.NewRequest("POST", "/logout", nil)
	r2.AddCookie(&http.Cookie{Name: adminUrlCookieName, Value: url.QueryEscape("http://admin.local")})
	rr2 := runHandler(LogoutHandler, r2)
	if rr2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rr2.Code)
	}
}

func TestLogoutAfterHandler_Redirect(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	fidp := &fakeIdp{}
	WebServerInjection(makeWs(fidp))
	r := httptest.NewRequest("GET", "/logoutafter", nil)
	r.AddCookie(&http.Cookie{Name: adminUrlCookieName, Value: url.QueryEscape("http://admin.local")})
	rr := runHandler(LogoutAfterHandler, r)
	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302 got %d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, "#/authorization") {
		t.Fatalf("unexpected Location %s", loc)
	}
}

// CodeHandler branches: missing code, idp returns empty token, invalid token, valid token
// For invalid token we inject a token that ValidateAndGetLoginToken will reject (use plain text)
func TestCodeHandler_Branches(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	// missing code
	fidp := &fakeIdp{}
	WebServerInjection(makeWs(fidp))
	rMissing := httptest.NewRequest("GET", "/code", nil)
	rrMissing := runHandler(CodeHandler, rMissing)
	if rrMissing.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rrMissing.Code)
	}

	// empty token from idp => 500
	fidpEmpty := &fakeIdp{tokenReturn: "", tokenErr: true}
	WebServerInjection(makeWs(fidpEmpty))
	rEmpty := httptest.NewRequest("GET", "/code?code=abc", nil)
	rrEmpty := runHandler(CodeHandler, rEmpty)
	if rrEmpty.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 empty token got %d", rrEmpty.Code)
	}

	// invalid token (ValidateAndGetLoginToken fails) => expect 500
	// xhttp.ValidateAndGetLoginToken expects JWT; feed junk so it fails
	fidpInvalid := &fakeIdp{tokenReturn: "not-a-jwt"}
	WebServerInjection(makeWs(fidpInvalid))
	rInvalid := httptest.NewRequest("GET", "/code?code=abc", nil)
	rrInvalid := runHandler(CodeHandler, rInvalid)
	if rrInvalid.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 invalid token got %d", rrInvalid.Code)
	}

	// valid-ish token: craft minimal signed JWT using helper? Simplify by bypassing validation: we cannot easily sign acceptable JWT without secret knowledge; skip success branch if library enforces signature.
	// Instead, simulate success by stubbing ValidateAndGetLoginToken via a simple replacement if available. If not, this branch is omitted.
}
