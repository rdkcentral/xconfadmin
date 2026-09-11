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
	ctx = context.WithValue(ctx, xhttp.CTX_KEY_TENANT_ID, strings.ToUpper("comcast"))
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
	ctx = context.WithValue(ctx, xhttp.CTX_KEY_TENANT_ID, strings.ToUpper(tenantId))
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

// 4.3 – SAT_V2 with correct capability and tenantId present, but allowedPartners is an empty slice.
func TestCanReadSATv2FailsWhenAllowedPartnersEmpty(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/dcm",
		http.MethodGet,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readonly"},
		[]string{}, // empty slice, not nil
	)

	_, err := CanRead(r, DCM_ENTITY)
	if err == nil {
		t.Fatalf("expected error for empty allowedPartners slice")
	}
	if !strings.Contains(err.Error(), "missing allowed partners") {
		t.Fatalf("expected missing allowed partners error, got: %v", err)
	}
}

// 4.6 – SAT_V2 with wrong/missing capability and invalid allowedPartners: capability failure
// must be returned before tenant scope failure.
func TestCanReadSATv2CapabilityCheckedBeforeTenantScope(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	// Wrong capability (write-only token on a read path) + empty allowedPartners.
	// If capability is evaluated first, we get a capability error, not a tenant error.
	r := requestWithSATScope(
		"/xconfadminservice/dcm",
		http.MethodGet,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readwrite"}, // write cap; readonly required for CanRead
		[]string{},                       // empty allowedPartners – would be tenant error if reached
	)

	// Force a capability miss: use a path that maps to a different domain so the
	// read capability check fails before tenant scope is evaluated.
	r2 := requestWithSATScope(
		"/xconfadminservice/dcm",
		http.MethodGet,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{}, // no capabilities at all
		[]string{}, // also no allowed partners
	)

	_, err := CanRead(r2, DCM_ENTITY)
	if err == nil {
		t.Fatalf("expected capability error")
	}
	// Error must be about capabilities, not about tenant/allowedPartners.
	if strings.Contains(err.Error(), "allowed partners") || strings.Contains(err.Error(), "tenantId") {
		t.Fatalf("expected capability error before tenant check, got: %v", err)
	}
	_ = r // suppress unused warning
}

// 4.7a – SAT_LEGACY with legacy read capability succeeds without tenantId or allowedPartners.
func TestCanReadLegacySATSucceedsWithoutTenantScope(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := httptest.NewRequest(http.MethodGet, "/xconfadminservice/dcm?applicationType=stb", nil)
	// No tenantId header, no allowedPartners in context.
	ctx := context.WithValue(r.Context(), xhttp.CTX_KEY_AUTH_TYPE, xhttp.AUTH_TYPE_SAT_LEGACY)
	ctx = context.WithValue(ctx, xhttp.CTX_KEY_CAPABILITIES, []string{XCONF_READ})
	r = r.WithContext(ctx)

	if _, err := CanRead(r, DCM_ENTITY); err != nil {
		t.Fatalf("expected legacy SAT read to succeed without tenant scope, got: %v", err)
	}
}

// 4.7b – LOGIN_TOKEN path succeeds with permissions and without tenantId or allowedPartners.
func TestCanReadLoginTokenSucceedsWithoutTenantScope(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	entityPerm := getEntityPermission(DCM_ENTITY)

	r := httptest.NewRequest(http.MethodGet, "/xconfadminservice/dcm?applicationType=stb", nil)
	// No tenantId header, no allowedPartners – login token path uses permissions only.
	ctx := context.WithValue(r.Context(), xhttp.CTX_KEY_AUTH_TYPE, xhttp.AUTH_TYPE_LOGIN_TOKEN)
	ctx = context.WithValue(ctx, xhttp.CTX_KEY_PERMISSIONS, []string{entityPerm.ReadAll})
	r = r.WithContext(ctx)

	if _, err := CanRead(r, DCM_ENTITY); err != nil {
		t.Fatalf("expected login token read to succeed without tenant scope, got: %v", err)
	}
}

// Metrics readonly-only behavior (3.4)

// xconf:metrics:readonly is the only valid metrics capability; read is allowed.
func TestCanReadSATv2MetricsReadonlyCapabilityAllowed(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/metrics",
		http.MethodGet,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:metrics:readonly"},
		[]string{"comcast"},
	)

	if _, err := CanRead(r, COMMON_ENTITY); err != nil {
		t.Fatalf("expected metrics read with xconf:metrics:readonly to be allowed, got: %v", err)
	}
}

// xconf:metrics:readwrite is not a defined capability; read must be denied.
func TestCanReadSATv2MetricsReadwriteCapabilityDenied(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/metrics",
		http.MethodGet,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:metrics:readwrite"},
		[]string{"comcast"},
	)

	_, err := CanRead(r, COMMON_ENTITY)
	if err == nil {
		t.Fatalf("expected metrics read with xconf:metrics:readwrite to be denied")
	}
	if !strings.Contains(err.Error(), "403") && !strings.Contains(err.Error(), "capability") && !strings.Contains(err.Error(), "permission") {
		t.Fatalf("expected 403 capability denial, got: %v", err)
	}
}

// Write to metrics with xconf:metrics:readonly must be denied.
func TestCanWriteSATv2MetricsReadonlyCapabilityDenied(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/metrics",
		http.MethodPost,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:metrics:readonly"},
		[]string{"comcast"},
	)

	_, err := CanWrite(r, COMMON_ENTITY)
	if err == nil {
		t.Fatalf("expected metrics write with xconf:metrics:readonly to be denied")
	}
	if !strings.Contains(err.Error(), "403") && !strings.Contains(err.Error(), "capability") && !strings.Contains(err.Error(), "permission") {
		t.Fatalf("expected 403 capability denial, got: %v", err)
	}
}

// Write to metrics with xconf:metrics:readwrite must also be denied (capability does not exist).
func TestCanWriteSATv2MetricsReadwriteCapabilityDenied(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/metrics",
		http.MethodPost,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:metrics:readwrite"},
		[]string{"comcast"},
	)

	_, err := CanWrite(r, COMMON_ENTITY)
	if err == nil {
		t.Fatalf("expected metrics write with xconf:metrics:readwrite to be denied")
	}
	if !strings.Contains(err.Error(), "403") && !strings.Contains(err.Error(), "capability") && !strings.Contains(err.Error(), "permission") {
		t.Fatalf("expected 403 capability denial, got: %v", err)
	}
}

// SAT v2 Detection (5.2)

// Capabilities with at least one xconf: prefix are detected as SAT v2 and follow v2 auth logic.
func TestSATv2DetectionWithXconfCapabilities(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSAT(
		"/xconfadminservice/dcm",
		core.STB,
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readonly"},
	)

	// SAT v2 logic: xconf:core:readonly is accepted for read
	if _, err := CanRead(r, DCM_ENTITY); err != nil {
		t.Fatalf("expected SAT v2 with xconf: prefix to use v2 auth logic, got error: %v", err)
	}
}

// Capabilities without any xconf: prefix are treated as legacy SAT and follow legacy auth logic.
func TestLegacySATDetectionWithoutXconfCapabilities(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSAT(
		"/xconfadminservice/dcm",
		"",
		xhttp.AUTH_TYPE_SAT_LEGACY,
		[]string{XCONF_READ}, // legacy capability, no xconf: prefix
	)

	// Legacy SAT logic: XCONF_READ is accepted for read
	if _, err := CanRead(r, DCM_ENTITY); err != nil {
		t.Fatalf("expected legacy SAT without xconf: prefix to use legacy auth logic, got error: %v", err)
	}
}

// Empty capabilities list is not SAT v2; treated as legacy SAT.
func TestLegacySATDetectionWithEmptyCapabilities(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSAT(
		"/xconfadminservice/dcm",
		"",
		xhttp.AUTH_TYPE_SAT_LEGACY,
		[]string{}, // empty capabilities
	)

	// Empty capabilities with legacy SAT type should fail (no XCONF_READ)
	_, err := CanRead(r, DCM_ENTITY)
	if err == nil {
		t.Fatalf("expected empty legacy SAT capabilities to be denied")
	}
	if !strings.Contains(err.Error(), "capabilities") && !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected read permission denial, got: %v", err)
	}
}

// Route-to-Domain Classification (5.3)

// /taggingService/tags should map to tagging domain
func TestClassifySATv2DomainTaggingService(t *testing.T) {
	domain, found := classifySATv2Domain("/taggingService/tags")
	if !found {
		t.Fatalf("expected /taggingService/tags to be classified")
	}
	if domain != owcommon.SATV2DomainTagging {
		t.Fatalf("expected tagging domain for /taggingService/tags, got: %s", domain)
	}
}

// /xconfAdminService/dcm should map to core domain
func TestClassifySATv2DomainDcmCore(t *testing.T) {
	domain, found := classifySATv2Domain("/xconfAdminService/dcm")
	if !found {
		t.Fatalf("expected /xconfAdminService/dcm to be classified")
	}
	if domain != owcommon.SATV2DomainCore {
		t.Fatalf("expected core domain for /xconfAdminService/dcm, got: %s", domain)
	}
}

// /xconfAdminService/info/statistics should map to system domain
func TestClassifySATv2DomainInfoSystem(t *testing.T) {
	domain, found := classifySATv2Domain("/xconfAdminService/info/statistics")
	if !found {
		t.Fatalf("expected /xconfAdminService/info/statistics to be classified")
	}
	if domain != owcommon.SATV2DomainSystem {
		t.Fatalf("expected system domain for /xconfAdminService/info/statistics, got: %s", domain)
	}
}

// /xconfAdminService/lockdownsettings should map to system domain
func TestClassifySATv2DomainLockdownSystem(t *testing.T) {
	domain, found := classifySATv2Domain("/xconfAdminService/lockdownsettings")
	if !found {
		t.Fatalf("expected /xconfAdminService/lockdownsettings to be classified")
	}
	if domain != owcommon.SATV2DomainSystem {
		t.Fatalf("expected system domain for /xconfAdminService/lockdownsettings, got: %s", domain)
	}
}

// /metrics should map to metrics domain
func TestClassifySATv2DomainMetrics(t *testing.T) {
	domain, found := classifySATv2Domain("/metrics")
	if !found {
		t.Fatalf("expected /metrics to be classified")
	}
	if domain != owcommon.SATV2DomainMetrics {
		t.Fatalf("expected metrics domain for /metrics, got: %s", domain)
	}
}

// First-match-wins: /xconfAdminService/rfc/recooking should match system before /rfc matches core
func TestClassifySATv2DomainFirstMatchWinsRfcRecooking(t *testing.T) {
	domain, found := classifySATv2Domain("/xconfAdminService/rfc/recooking")
	if !found {
		t.Fatalf("expected /xconfAdminService/rfc/recooking to be classified")
	}
	if domain != owcommon.SATV2DomainSystem {
		t.Fatalf("expected system domain for /xconfAdminService/rfc/recooking (first-match /rfc/recooking), got: %s", domain)
	}
}

// First-match-wins: /xconfAdminService/queries/filters/downloadlocation should match system before /queries matches core
func TestClassifySATv2DomainFirstMatchWinsQueriesFilters(t *testing.T) {
	domain, found := classifySATv2Domain("/xconfAdminService/queries/filters/downloadlocation")
	if !found {
		t.Fatalf("expected /xconfAdminService/queries/filters/downloadlocation to be classified")
	}
	if domain != owcommon.SATV2DomainSystem {
		t.Fatalf("expected system domain for /xconfAdminService/queries/filters/downloadlocation (first-match /queries/filters/downloadlocation), got: %s", domain)
	}
}

// Path normalization: case insensitivity and trailing slash handling
func TestClassifySATv2DomainCaseInsensitive(t *testing.T) {
	domain, found := classifySATv2Domain("/XconfAdminService/DCM")
	if !found {
		t.Fatalf("expected /XconfAdminService/DCM (mixed case) to be classified")
	}
	if domain != owcommon.SATV2DomainCore {
		t.Fatalf("expected core domain for mixed case path, got: %s", domain)
	}
}

// Path normalization: trailing slash should be stripped
func TestClassifySATv2DomainTrailingSlashStripped(t *testing.T) {
	domain, found := classifySATv2Domain("/xconfadminservice/dcm/")
	if !found {
		t.Fatalf("expected /xconfadminservice/dcm/ (with trailing slash) to be classified")
	}
	if domain != owcommon.SATV2DomainCore {
		t.Fatalf("expected core domain after stripping trailing slash, got: %s", domain)
	}
}

// Deny-by-default on Unclassified Routes (5.5)

// SAT v2 with valid capability on unmapped route should be denied (403).
func TestCanReadSATv2UnmappedRouteDeniedByDefault(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/unknown-api",
		http.MethodGet,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readonly"}, // valid SAT v2 capability
		[]string{"comcast"},             // tenant allowed
	)

	_, err := CanRead(r, COMMON_ENTITY)
	if err == nil {
		t.Fatalf("expected unmapped route to be denied by default (deny-by-default)")
	}
	if !strings.Contains(err.Error(), "permission") && !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected authorization denial for unmapped route, got: %v", err)
	}
}

// SAT v2 with valid capability on unmapped route CanWrite should also be denied (403).
func TestCanWriteSATv2UnmappedRouteDeniedByDefault(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	r := requestWithSATScope(
		"/xconfadminservice/unknown-api",
		http.MethodPost,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readwrite"}, // valid SAT v2 capability for write
		[]string{"comcast"},              // tenant allowed
	)

	_, err := CanWrite(r, COMMON_ENTITY)
	if err == nil {
		t.Fatalf("expected unmapped route to be denied by default (deny-by-default)")
	}
	if !strings.Contains(err.Error(), "permission") && !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected authorization denial for unmapped route, got: %v", err)
	}
}

// HTTP Status Semantics: 401 vs 403 (5.7)

// 403 Forbidden: Valid SAT v2 auth but insufficient capability
// 403 Forbidden: Valid SAT v2 auth but insufficient capability (wrong domain)
func TestSATv2InsufficientCapabilityReturns403(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	// Valid SAT v2 auth but with tagging capability trying to read core resource
	r := requestWithSATScope(
		"/xconfadminservice/dcm",
		http.MethodGet, // read operation
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:tagging:readonly"}, // wrong domain capability
		[]string{"comcast"},
	)

	_, err := CanRead(r, DCM_ENTITY)
	if err == nil {
		t.Fatalf("expected insufficient capability to be denied")
	}

	// Error should indicate authorization failure (403), not authentication failure (401)
	errMsg := err.Error()
	if !strings.Contains(errMsg, "403") && !strings.Contains(errMsg, "capability") && !strings.Contains(errMsg, "permission") {
		t.Fatalf("expected 403 authorization denial for insufficient capability, got: %v", err)
	}
}

// 403 Forbidden: Valid SAT v2 auth but unclassified route
func TestSATv2UnclassifiedRouteReturns403(t *testing.T) {
	oldSatOn := owcommon.SatOn
	owcommon.SatOn = true
	defer func() { owcommon.SatOn = oldSatOn }()

	// Valid SAT v2 auth on unmapped route
	r := requestWithSATScope(
		"/xconfadminservice/unknown-endpoint",
		http.MethodGet,
		"comcast",
		xhttp.AUTH_TYPE_SAT_V2,
		[]string{"xconf:core:readonly"},
		[]string{"comcast"},
	)

	_, err := CanRead(r, COMMON_ENTITY)
	if err == nil {
		t.Fatalf("expected unclassified route to be denied")
	}

	// Error should indicate authorization failure (403), not authentication failure (401)
	errMsg := err.Error()
	if !strings.Contains(errMsg, "403") && !strings.Contains(errMsg, "permission") {
		t.Fatalf("expected 403 authorization denial for unclassified route, got: %v", err)
	}
}

// Dev profile permissions include VIEW_TOOLS and WRITE_TOOLS so that TOOL_ENTITY
// endpoints (e.g. penetration metrics) pass CanRead/CanWrite checks without a SAT token.
func TestGetPermissionsDevProfileIncludesToolPermissions(t *testing.T) {
	oldActive := owcommon.ActiveAuthProfiles
	oldDefault := owcommon.DefaultAuthProfiles
	owcommon.ActiveAuthProfiles = "dev"
	owcommon.DefaultAuthProfiles = "prod"
	defer func() {
		owcommon.ActiveAuthProfiles = oldActive
		owcommon.DefaultAuthProfiles = oldDefault
	}()

	r := httptest.NewRequest("GET", "/xconfAdminService/penetrationdata/AA:BB:CC:DD:EE:FF", nil)
	permissions := GetPermissionsFunc(r)

	for _, required := range []string{VIEW_TOOLS, WRITE_TOOLS} {
		found := false
		for _, p := range permissions {
			if p == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("dev profile permissions missing %q; got: %v", required, permissions)
		}
	}
}

// TOOL_ENTITY CanRead succeeds under dev profile without a SAT token when SAT_ON=true.
func TestCanReadToolEntitySucceedsInDevProfileWithSATOn(t *testing.T) {
	oldSatOn := owcommon.SatOn
	oldActive := owcommon.ActiveAuthProfiles
	oldDefault := owcommon.DefaultAuthProfiles
	owcommon.SatOn = true
	owcommon.ActiveAuthProfiles = "dev"
	owcommon.DefaultAuthProfiles = "prod"
	defer func() {
		owcommon.SatOn = oldSatOn
		owcommon.ActiveAuthProfiles = oldActive
		owcommon.DefaultAuthProfiles = oldDefault
	}()

	// No auth context set — simulates NoAuthMiddleware (test router behaviour)
	r := httptest.NewRequest("GET", "/xconfAdminService/penetrationdata/AA:BB:CC:DD:EE:FF", nil)

	_, err := CanRead(r, TOOL_ENTITY)
	if err != nil {
		t.Fatalf("expected CanRead to succeed for TOOL_ENTITY in dev profile with SAT_ON=true, got: %v", err)
	}
}
