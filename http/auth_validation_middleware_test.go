// Copyright 2025 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	xcommon "github.com/rdkcentral/xconfadmin/common"
	"github.com/rdkcentral/xconfwebconfig/db"
)

// ----------------------------------------------------------------------------
// helpers
// ----------------------------------------------------------------------------

// okHandler is a simple 200 handler used as the "next" in middleware tests.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// serveWithMiddleware routes the request through a mux router so that
// mux.CurrentRoute(r) in logRequestEnds returns a non-nil value.
func serveWithMiddleware(ws *WebconfigServer, next http.Handler, r *http.Request) *httptest.ResponseRecorder {
	router := mux.NewRouter()
	router.Handle(r.URL.Path, ws.AuthValidationMiddleware(next)).Methods(r.Method)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
	return rr
}

// createTenant inserts a tenant into the real DB and registers a cleanup to
// remove it when the test ends.
func createTenant(t *testing.T, id string) {
	t.Helper()
	if err := db.GetDatabaseClient().SetTenant(db.NewTenant(id, id)); err != nil {
		t.Fatalf("failed to create tenant %q: %v", id, err)
	}
	t.Cleanup(func() {
		db.GetDatabaseClient().DeleteTenant(id)
	})
}

// resetOnboardTenantFunc saves testServer.OnboardTenantFunc and restores it when the
// test ends so each test starts from a clean state.
func resetOnboardTenantFunc(t *testing.T) {
	t.Helper()
	original := testServer.OnboardTenantFunc
	t.Cleanup(func() { testServer.OnboardTenantFunc = original })
}

// ----------------------------------------------------------------------------
// tests
// ----------------------------------------------------------------------------

// TestAuthMiddleware_NoToken_SatOff verifies that requests without any token
// are allowed through when SAT is disabled.
func TestAuthMiddleware_NoToken_SatOff(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()

	createTenant(t, db.GetDefaultTenantId())

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := serveWithMiddleware(testServer, okHandler, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with SAT off and no token, got %d", rr.Code)
	}
}

// TestAuthMiddleware_NoToken_SatOn verifies that requests without any token
// are rejected with 401 when SAT is enabled.
func TestAuthMiddleware_NoToken_SatOn(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = true
	defer func() { xcommon.SatOn = oldSatOn }()

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := serveWithMiddleware(testServer, okHandler, r)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with SAT on and no token, got %d", rr.Code)
	}
}

// TestAuthMiddleware_InvalidLoginToken verifies that a malformed login token
// is rejected with 401.
func TestAuthMiddleware_InvalidLoginToken(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set(AUTH_TOKEN, "not-a-valid-jwt")
	rr := serveWithMiddleware(testServer, okHandler, r)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid login token, got %d", rr.Code)
	}
}

// TestAuthMiddleware_TenantNotFound_SatOffDoesNotOnboard verifies that a
// request allowed through the SAT-off bypass cannot onboard a missing tenant.
func TestAuthMiddleware_TenantNotFound_SatOffDoesNotOnboard(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()
	oldTestOnly := testServer.testOnly
	testServer.testOnly = true
	defer func() { testServer.testOnly = oldTestOnly }()

	resetOnboardTenantFunc(t)
	onboardCalled := false
	handlerCalled := false
	testServer.OnboardTenantFunc = func(id, name string) (*db.Tenant, error) {
		onboardCalled = true
		return nil, nil
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("tenantId", uuid.New().String()) // random UUID — not in DB
	rr := serveWithMiddleware(testServer, handler, r)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-SAT-v2 missing tenant, got %d", rr.Code)
	}
	if onboardCalled {
		t.Fatalf("expected non-SAT-v2 request not to onboard tenant")
	}
	if handlerCalled {
		t.Fatalf("expected downstream handler not to execute after tenant validation failure")
	}
}

// TestAuthMiddleware_TenantNotFound_SatOffDoesNotCallFailingOnboardFunc verifies
// that the SAT-off bypass does not call the configured onboarding function.
func TestAuthMiddleware_TenantNotFound_SatOffDoesNotCallFailingOnboardFunc(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()

	resetOnboardTenantFunc(t)
	onboardCalled := false
	testServer.OnboardTenantFunc = func(id, name string) (*db.Tenant, error) {
		onboardCalled = true
		return nil, errors.New("onboard failed")
	}

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("tenantId", uuid.New().String())
	rr := serveWithMiddleware(testServer, okHandler, r)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for SAT-off missing tenant, got %d", rr.Code)
	}
	if onboardCalled {
		t.Fatalf("expected SAT-off request not to onboard tenant")
	}
}

// TestAuthMiddleware_TenantExists_NoOnboard verifies that OnboardTenantFunc is NOT
// called when the tenant already exists.
func TestAuthMiddleware_TenantExists_NoOnboard(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()

	createTenant(t, db.GetDefaultTenantId())

	resetOnboardTenantFunc(t)
	onboardCalled := false
	testServer.OnboardTenantFunc = func(id, name string) (*db.Tenant, error) {
		onboardCalled = true
		return nil, nil
	}

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := serveWithMiddleware(testServer, okHandler, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if onboardCalled {
		t.Fatalf("expected OnboardTenantFunc NOT to be called for existing tenant")
	}
}

// TestAuthMiddleware_SatOffTenantHeaderIgnored verifies that requests allowed
// through the SAT-off bypass do not select a tenant from the request header.
func TestAuthMiddleware_SatOffTenantHeaderIgnored(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()

	const headerTenantId = "ACME"
	createTenant(t, headerTenantId)
	createTenant(t, db.GetDefaultTenantId())

	var capturedTenantId string
	captureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantId, _ = r.Context().Value(CTX_KEY_TENANT_ID).(string)
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("tenantId", headerTenantId)
	rr := serveWithMiddleware(testServer, captureHandler, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedTenantId != db.GetDefaultTenantId() {
		t.Fatalf("expected default tenant ID %q in context, got %q", db.GetDefaultTenantId(), capturedTenantId)
	}
}

func TestResolveTenantIDByAuthType(t *testing.T) {
	const (
		headerTenant  = "HEADER_TENANT"
		defaultTenant = "DEFAULT_TENANT"
	)
	tests := []struct {
		name     string
		authType AuthType
		enabled  bool
		header   string
		expected string
	}{
		{name: "SAT v2 uses header", authType: AUTH_TYPE_SAT_V2, header: headerTenant, expected: headerTenant},
		{name: "SAT v2 falls back", authType: AUTH_TYPE_SAT_V2, expected: defaultTenant},
		{name: "legacy SAT uses default", authType: AUTH_TYPE_SAT_LEGACY, enabled: true, header: headerTenant, expected: defaultTenant},
		{name: "login token flag disabled", authType: AUTH_TYPE_LOGIN_TOKEN, header: headerTenant, expected: defaultTenant},
		{name: "login token flag enabled", authType: AUTH_TYPE_LOGIN_TOKEN, enabled: true, header: headerTenant, expected: headerTenant},
		{name: "login token flag enabled without header", authType: AUTH_TYPE_LOGIN_TOKEN, enabled: true, expected: defaultTenant},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := resolveTenantID(test.authType, test.header, defaultTenant, test.enabled); got != test.expected {
				t.Fatalf("expected tenant %q, got %q", test.expected, got)
			}
		})
	}
}

func TestCanAutoCreateTenantRequiresSATV2ReadWrite(t *testing.T) {
	if !canAutoCreateTenant(AUTH_TYPE_SAT_V2, []string{"xconf:system:readwrite"}) {
		t.Fatal("expected SAT v2 system readwrite capability to allow onboarding")
	}
	if canAutoCreateTenant(AUTH_TYPE_SAT_V2, []string{"xconf:system:readonly"}) {
		t.Fatal("expected readonly capability to deny onboarding")
	}
	if canAutoCreateTenant(AUTH_TYPE_LOGIN_TOKEN, []string{"xconf:system:readwrite"}) {
		t.Fatal("expected login-token onboarding to be denied")
	}
	if canAutoCreateTenant(AUTH_TYPE_SAT_LEGACY, []string{"xconf:system:readwrite"}) {
		t.Fatal("expected legacy SAT onboarding to be denied")
	}
}

func TestCanOnboardTenantRequiresAllowedPartnerMatch(t *testing.T) {
	if !canOnboardMissingTenant(AUTH_TYPE_SAT_V2, "COMCAST", []string{"xconf:system:readwrite"}, []string{"comcast"}) {
		t.Fatal("expected allowed partner match to permit SAT v2 tenant onboarding")
	}
	if canOnboardMissingTenant(AUTH_TYPE_SAT_V2, "COMCAST", []string{"xconf:system:readwrite"}, []string{"acme"}) {
		t.Fatal("expected tenant outside allowedPartners to be denied")
	}
	if canOnboardMissingTenant(AUTH_TYPE_SAT_V2, "COMCAST", []string{"xconf:system:readwrite"}, []string{}) {
		t.Fatal("expected empty allowedPartners to deny onboarding")
	}
	if canOnboardMissingTenant(AUTH_TYPE_SAT_V2, "COMCAST", []string{"xconf:system:readonly"}, []string{"comcast"}) {
		t.Fatal("expected system-readonly capability to deny onboarding")
	}
	if canOnboardMissingTenant(AUTH_TYPE_LOGIN_TOKEN, "COMCAST", []string{"xconf:system:readwrite"}, []string{"comcast"}) {
		t.Fatal("expected login-token onboarding to be denied")
	}
}

// TestAuthMiddleware_TenantIdDefaultsWhenHeaderMissing verifies that the
// default tenant ID is used when no tenantId header is present.
func TestAuthMiddleware_TenantIdDefaultsWhenHeaderMissing(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()

	createTenant(t, db.GetDefaultTenantId())

	var capturedTenantId string
	captureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantId, _ = r.Context().Value(CTX_KEY_TENANT_ID).(string)
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	// no tenantId header
	rr := serveWithMiddleware(testServer, captureHandler, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedTenantId != db.GetDefaultTenantId() {
		t.Fatalf("expected default tenant ID %q, got %q", db.GetDefaultTenantId(), capturedTenantId)
	}
}
