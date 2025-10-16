package dcm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
)

// helper to build XResponseWriter with provided raw body JSON
func newTestXWriter(body string) (*xwhttp.XResponseWriter, *httptest.ResponseRecorder) {
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(body)
	return xw, rr
}

// 1. Cast error path: provide a plain ResponseRecorder (not wrapped) so handler fails casting
func TestDcmTestPageHandler_CastError(t *testing.T) {
	// Need applicationType for auth.CanRead; append as query param
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/dcm/testpage?applicationType=stb", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder() // NOT an XResponseWriter -> triggers cast error branch
	DcmTestPageHandler(w, r)
	if w.Code != http.StatusInternalServerError { // AdminError writes 500
		t.Fatalf("expected 500 cast error, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "responsewriter cast error") {
		t.Fatalf("expected cast error message in body, got %s", w.Body.String())
	}
}

// 2. Bad JSON path: XResponseWriter but body not valid JSON
func TestDcmTestPageHandler_BadJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/dcm/testpage?applicationType=stb", nil)
	xw, rr := newTestXWriter("{invalid-json")
	DcmTestPageHandler(xw, r)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad json, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Unable to extract") {
		t.Fatalf("expected extraction error, body=%s", rr.Body.String())
	}
}

// 3. Success path with no matching rules -> should return context only (no settings)
func TestDcmTestPageHandler_SuccessNoRules(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/dcm/testpage?applicationType=stb", nil)
	// Provide minimal empty JSON body
	xw, rr := newTestXWriter("{}")
	DcmTestPageHandler(xw, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	// Body should contain context with applicationType added, but not contain matchedRules/settings keys
	body := rr.Body.String()
	if !strings.Contains(body, "context") {
		t.Fatalf("expected context key in response: %s", body)
	}
	if !strings.Contains(body, xwcommon.APPLICATION_TYPE) {
		t.Fatalf("expected applicationType in context: %s", body)
	}
	if strings.Contains(body, "matchedRules") || strings.Contains(body, "settings") {
		t.Fatalf("did not expect matchedRules/settings for empty eval: %s", body)
	}
	// verify JSON decodes
	var decoded map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("response not valid json: %v body=%s", err, body)
	}
}

// 4. Authentication path when applicationType missing: CanRead should default to stb (dev profile) and still succeed
func TestDcmTestPageHandler_DefaultApplicationType(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/dcm/testpage", nil)
	xw, rr := newTestXWriter("{}")
	DcmTestPageHandler(xw, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "\"applicationType\":\"stb\"") { // default fallback
		t.Fatalf("expected default stb applicationType, body=%s", rr.Body.String())
	}
}
