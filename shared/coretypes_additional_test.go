package shared

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestInitializeAndValidateApplicationTypes ensures defaults populated and validation works.
func TestInitializeAndValidateApplicationTypes(t *testing.T) {
	InitializeApplicationTypes()
	if !IsValidApplicationType(STB) || !IsValidApplicationType(RDKCLOUD) {
		t.Fatalf("expected default application types to be valid")
	}
	if IsValidApplicationType("invalid-type") {
		t.Fatalf("did not expect invalid-type to be valid")
	}
	if err := ValidateApplicationType(STB); err != nil {
		t.Fatalf("validate should succeed: %v", err)
	}
	if err := ValidateApplicationType("bad"); err == nil {
		t.Fatalf("expected validation failure for bad type")
	}
}

// TestApplicationTypeEquals with empty defaults mapping to STB
func TestApplicationTypeEquals(t *testing.T) {
	if !ApplicationTypeEquals("", "") { // both default to STB
		t.Fatalf("expected empty types to equal via defaulting")
	}
	if ApplicationTypeEquals("stb", "rdkcloud") {
		t.Fatalf("expected differing types not equal")
	}
}

// TestEnvironmentValidate covers allowed characters and invalid ones.
func TestEnvironmentValidate(t *testing.T) {
	env := NewEnvironment("Prod_1", "desc")
	if err := env.Validate(); err != nil {
		t.Fatalf("expected valid id, got error %v", err)
	}
	envBad := NewEnvironment("Bad#Id", "desc")
	if err := envBad.Validate(); err == nil {
		t.Fatalf("expected validation error for bad id")
	}
}

// TestModelValidate and CreateModelResponse
func TestModelValidate(t *testing.T) {
	m := NewModel("modelA", "desc")
	if err := m.Validate(); err != nil {
		t.Fatalf("model validate should pass: %v", err)
	}
	mBad := NewModel("Bad#Model", "desc")
	if err := mBad.Validate(); err == nil {
		t.Fatalf("expected invalid model id error")
	}
	resp := m.CreateModelResponse()
	if resp.ID != m.ID || resp.Description != m.Description {
		t.Fatalf("response mismatch")
	}
}

// TestNormalizeCommonContext ensures uppercasing and MAC normalization.
func TestNormalizeCommonContext(t *testing.T) {
	ctx := map[string]string{MODEL: "abc", ENVIRONMENT: "prod", PARTNER_ID: "partner", ESTB_MAC: "AA:bb:CC:dd:EE:ff"}
	if err := NormalizeCommonContext(ctx, ESTB_MAC, ECM_MAC); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx[MODEL] != "ABC" || ctx[ENVIRONMENT] != "PROD" || ctx[PARTNER_ID] != "PARTNER" {
		t.Fatalf("expected uppercasing applied")
	}
	if ctx[ESTB_MAC] != "AA:BB:CC:DD:EE:FF" { // normalized MAC
		t.Fatalf("mac not normalized: %s", ctx[ESTB_MAC])
	}
}

// TestGetApplicationFromCookies verifies cookie retrieval.
func TestGetApplicationFromCookies(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	// absent cookie returns empty
	if v := GetApplicationFromCookies(r); v != "" {
		t.Fatalf("expected empty application type with no cookie")
	}
	r.AddCookie(&http.Cookie{Name: APPLICATION_TYPE, Value: STB})
	if v := GetApplicationFromCookies(r); v != STB {
		t.Fatalf("expected %s got %s", STB, v)
	}
}
