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
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	owcommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	core "github.com/rdkcentral/xconfadmin/shared"
)

func requestWithSAT(path string, queryApplicationType string, authType xhttp.AuthType, capabilities []string) *http.Request {
	r := httptest.NewRequest("GET", path, nil)
	r.Header.Set("tenantId", "comcast")
	if queryApplicationType != "" {
		q := r.URL.Query()
		q.Set(core.APPLICATION_TYPE, queryApplicationType)
		r.URL.RawQuery = q.Encode()
	}

	ctx := context.WithValue(r.Context(), xhttp.CTX_KEY_AUTH_TYPE, authType)
	ctx = context.WithValue(ctx, xhttp.CTX_KEY_CAPABILITIES, capabilities)
	ctx = context.WithValue(ctx, xhttp.CTX_KEY_ALLOWED_PARTNERS, []string{"comcast"})
	return r.WithContext(ctx)
}

func requestWithSATScope(path string, method string, tenantId string, authType xhttp.AuthType, capabilities []string, allowedPartners []string) *http.Request {
	r := httptest.NewRequest(method, path, nil)
	if tenantId != "" {
		r.Header.Set("tenantId", tenantId)
	}

	ctx := context.WithValue(r.Context(), xhttp.CTX_KEY_AUTH_TYPE, authType)
	ctx = context.WithValue(ctx, xhttp.CTX_KEY_CAPABILITIES, capabilities)
	ctx = context.WithValue(ctx, xhttp.CTX_KEY_ALLOWED_PARTNERS, allowedPartners)
	return r.WithContext(ctx)
}

func TestValidateReadSATv2MismatchedApplicationType(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSAT(
		"/xconfadminservice/dcm",
		core.RDKCLOUD,
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readonly"},
	)

	err := ValidateRead(r, core.STB, DCM_ENTITY)
	if err == nil {
		t.Fatalf("expected applicationType mismatch error")
	}
	if !strings.Contains(err.Error(), "doesn't match") {
		t.Fatalf("expected mismatch error, got: %v", err)
	}
}

func TestValidateReadSATv2MatchingApplicationType(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSAT(
		"/xconfadminservice/dcm",
		core.STB,
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readonly"},
	)

	if err := ValidateRead(r, core.STB, DCM_ENTITY); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestValidateWriteLegacySATMismatchedApplicationType(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSAT(
		"/xconfadminservice/dcm",
		core.RDKCLOUD,
		xhttp.AUTH_TYPE_SAT_LEGACY,
		[]string{XCONF_WRITE},
	)

	err := ValidateWrite(r, core.STB, DCM_ENTITY)
	if err == nil {
		t.Fatalf("expected applicationType mismatch error")
	}
	if !strings.Contains(err.Error(), "doesn't match") {
		t.Fatalf("expected mismatch error, got: %v", err)
	}
}

func TestValidateWriteLegacySATMatchingApplicationType(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSAT(
		"/xconfadminservice/dcm",
		"",
		xhttp.AUTH_TYPE_SAT_LEGACY,
		[]string{XCONF_WRITE},
	)

	if err := ValidateWrite(r, core.STB, DCM_ENTITY); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestCanReadSATv2CaseInsensitiveTenantMatch(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/dcm",
		http.MethodGet,
		"COX",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readonly"},
		[]string{"cox"},
	)

	if _, err := CanRead(r, DCM_ENTITY); err != nil {
		t.Fatalf("expected success for case-insensitive tenant match, got: %v", err)
	}
}

func TestCanReadSATv2FailsWhenTenantHeaderMissing(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/dcm",
		http.MethodGet,
		"",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readonly"},
		[]string{"comcast"},
	)

	_, err := CanRead(r, DCM_ENTITY)
	if err == nil {
		t.Fatalf("expected error for missing tenantId")
	}
	if !strings.Contains(err.Error(), "Missing tenantId") {
		t.Fatalf("expected missing tenantId error, got: %v", err)
	}
}

func TestCanReadSATv2FailsWhenAllowedPartnersMissing(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/dcm",
		http.MethodGet,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readonly"},
		nil,
	)

	_, err := CanRead(r, DCM_ENTITY)
	if err == nil {
		t.Fatalf("expected error for missing allowed partners")
	}
	if !strings.Contains(err.Error(), "missing allowed partners") {
		t.Fatalf("expected missing allowed partners error, got: %v", err)
	}
}

func TestCanWriteSATv2FailsWhenTenantNotAllowed(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/dcm",
		http.MethodPost,
		"cox",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readwrite"},
		[]string{"comcast"},
	)

	_, err := CanWrite(r, DCM_ENTITY, core.STB)
	if err == nil {
		t.Fatalf("expected error for tenant outside allowedPartners")
	}
	if !strings.Contains(err.Error(), "not allowed for tenant") {
		t.Fatalf("expected tenant scope denial error, got: %v", err)
	}
}
