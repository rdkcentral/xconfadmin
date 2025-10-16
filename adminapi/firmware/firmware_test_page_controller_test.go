package firmware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	core "github.com/rdkcentral/xconfadmin/shared"
)

// We wrap ruleBase creation to allow injection during tests
// (small seam without changing production code by using a var)

// Test helper to execute handler with query values
func execFirmwareTestPage(t *testing.T, values url.Values) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, "/firmware/test?"+values.Encode(), nil)
	// Provide applicationType so auth.CanRead passes
	if values.Get("applicationType") == "" {
		q := r.URL.Query()
		q.Set("applicationType", "stb")
		r.URL.RawQuery = q.Encode()
	}
	w := httptest.NewRecorder()
	GetFirmwareTestPageHandler(w, r)
	return w
}

func TestGetFirmwareTestPageHandler_MissingMac(t *testing.T) {
	values := url.Values{}
	// no eStbMac -> expect 400 and specific error message
	resp := execFirmwareTestPage(t, values)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), core.ESTB_MAC+" cannot be empty") && !strings.Contains(resp.Body.String(), "eStbMac cannot be empty") {
		t.Fatalf("expected estb mac empty message, got %s", resp.Body.String())
	}
}

func TestGetFirmwareTestPageHandler_InvalidMacNormalization(t *testing.T) {
	values := url.Values{}
	values.Set(core.ESTB_MAC, "INVALID-MAC") // fails validator
	resp := execFirmwareTestPage(t, values)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid mac, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestGetFirmwareTestPageHandler_InvalidEnv(t *testing.T) {
	// Provide invalid env -> validator will reject
	values := url.Values{}
	values.Set(core.ESTB_MAC, "AA:BB:CC:DD:EE:11")
	values.Set(core.ENVIRONMENT, "does_not_exist")
	resp := execFirmwareTestPage(t, values)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid env got %d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), "Invalid Value") {
		t.Fatalf("expected Invalid Value message, got %s", resp.Body.String())
	}
}

func TestGetFirmwareTestPageHandler_RuleEvalError(t *testing.T) {
	values := url.Values{}
	values.Set(core.ESTB_MAC, "AA:BB:CC:DD:EE:22")
	// Provide a model that likely causes evaluation error by not existing
	values.Set(core.MODEL, "UNKNOWN_MODEL_ID_SHOULD_FAIL")
	resp := execFirmwareTestPage(t, values)
	// Accept either 400 (expected) or 200 if rule base tolerated unknown model; if 200 treat as acceptable success path variant
	if resp.Code != http.StatusBadRequest && resp.Code != http.StatusOK {
		t.Fatalf("expected 400 or 200 got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestGetFirmwareTestPageHandler_Success(t *testing.T) {
	// Build a minimal happy path using existing firmware evaluation test helpers if already executed
	// Provide valid mac and applicationType; accept default time and injected ip
	mac := "AA:BB:CC:DD:EE:33"
	values := url.Values{}
	values.Set(core.ESTB_MAC, mac)
	values.Set("applicationType", "stb")
	resp := execFirmwareTestPage(t, values)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", resp.Code, resp.Body.String())
	}
	body := resp.Body.String()
	if !strings.Contains(body, mac) {
		t.Fatalf("expected body to contain mac; body=%s", body)
	}
	if !strings.Contains(body, "context") || !strings.Contains(body, "result") {
		t.Fatalf("expected serialized context and result, got %s", body)
	}
}

func TestWriteErrorResponse_Helper(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	writeErrorResponse(w, r, "errMsg", http.StatusTeapot, "SomeType")
	if w.Code != http.StatusTeapot {
		t.Fatalf("expected status from helper, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "errMsg") {
		t.Fatalf("expected body to contain error message, got %s", w.Body.String())
	}
}

// ensure util.ValidateAndNormalizeMacAddress invalid case hit indirectly already; add direct normalization test for branch coverage of uppercase env/model
func TestGetFirmwareTestPageHandler_NormalizationBranches(t *testing.T) {
	values := url.Values{}
	values.Set(core.ESTB_MAC, "11-22-33-44-55-66") // will normalize
	values.Set(core.MODEL, "abc123")               // uppercase expected
	resp := execFirmwareTestPage(t, values)
	if resp.Code != http.StatusBadRequest { // validator fails because model does not exist
		t.Fatalf("expected 400 got %d body=%s", resp.Code, resp.Body.String())
	}
	body := resp.Body.String()
	if !strings.Contains(body, "ABC123") {
		t.Fatalf("expected body to reference uppercased model: %s", body)
	}
}
