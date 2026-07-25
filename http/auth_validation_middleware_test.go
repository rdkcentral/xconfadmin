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

// TestAuthMiddleware_TenantNotFound_OnboardFuncNil verifies that a missing
// OnboardTenantFunc results in 500 when the tenant does not exist yet.
func TestAuthMiddleware_TenantNotFound_OnboardFuncNil(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()

	resetOnboardTenantFunc(t)
	testServer.OnboardTenantFunc = nil

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("tenantId", uuid.New().String()) // random UUID — not in DB
	rr := serveWithMiddleware(testServer, okHandler, r)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when OnboardTenantFunc is nil, got %d", rr.Code)
	}
}

// TestAuthMiddleware_TenantNotFound_OnboardFuncError verifies that a failure in
// OnboardTenantFunc results in 500.
func TestAuthMiddleware_TenantNotFound_OnboardFuncError(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()

	resetOnboardTenantFunc(t)
	testServer.OnboardTenantFunc = func(id, name string) error {
		return errors.New("onboard failed")
	}

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("tenantId", uuid.New().String())
	rr := serveWithMiddleware(testServer, okHandler, r)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when OnboardTenantFunc fails, got %d", rr.Code)
	}
}

// TestAuthMiddleware_TenantNotFound_OnboardFuncSuccess verifies that a new
// tenant is onboarded successfully and the request proceeds.
func TestAuthMiddleware_TenantNotFound_OnboardFuncSuccess(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()

	resetOnboardTenantFunc(t)
	onboardCalled := false
	testServer.OnboardTenantFunc = func(id, name string) error {
		onboardCalled = true
		return nil
	}

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("tenantId", uuid.New().String())
	rr := serveWithMiddleware(testServer, okHandler, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 after successful tenant onboarding, got %d", rr.Code)
	}
	if !onboardCalled {
		t.Fatalf("expected OnboardTenantFunc to be called")
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
	testServer.OnboardTenantFunc = func(id, name string) error {
		onboardCalled = true
		return nil
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

// TestAuthMiddleware_TenantIdFromHeader verifies that when a tenantId header
// is present, it is stored in the request context.
func TestAuthMiddleware_TenantIdFromHeader(t *testing.T) {
	oldSatOn := xcommon.SatOn
	xcommon.SatOn = false
	defer func() { xcommon.SatOn = oldSatOn }()

	const headerTenantId = "ACME"
	createTenant(t, headerTenantId)

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
	if capturedTenantId != headerTenantId {
		t.Fatalf("expected tenant ID %q in context, got %q", headerTenantId, capturedTenantId)
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
