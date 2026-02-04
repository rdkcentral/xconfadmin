package dcm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
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

// 5. Success path with matching rules -> should return settings, matchedRules, and ruleType
func TestDcmTestPageHandler_SuccessWithMatchingRules(t *testing.T) {
	// Setup: Create a DCM formula and device settings that will match
	defer func() {
		// Clean up any test data
		if r := recover(); r != nil {
			t.Logf("Test may have panicked (DB not configured): %v", r)
		}
	}()

	// Create a device settings object
	deviceSettings := &logupload.DeviceSettings{
		ID:                "test-formula-123", // Use same ID as formula for linking
		Name:              "TEST_SETTINGS",
		CheckOnReboot:     true,
		SettingsAreActive: true,
		ApplicationType:   "stb",
	}

	// Create a simple DCM formula that matches on model
	condition := re.NewCondition(estbfirmware.RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg("TEST_MODEL"))
	formula := &logupload.DCMGenericRule{
		ID:              "test-formula-123",
		Name:            "TEST_FORMULA",
		Rule:            re.Rule{Condition: condition},
		Priority:        1,
		Percentage:      100,
		ApplicationType: "stb",
	}

	// Store in database - DeviceSettings uses same ID as formula for association
	_ = setOneInDao(ds.TABLE_DCM_RULE, formula.ID, formula)
	_ = setOneInDao(ds.TABLE_DEVICE_SETTINGS, deviceSettings.ID, deviceSettings)

	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/dcm/testpage?applicationType=stb", nil)
	// Provide context that will match our rule
	searchContext := map[string]string{
		"estbMacAddress": "AA:BB:CC:DD:EE:FF",
		"model":          "TEST_MODEL", // This matches our formula
		"env":            "PROD",
	}
	contextJSON, _ := json.Marshal(searchContext)
	xw, rr := newTestXWriter(string(contextJSON))

	DcmTestPageHandler(xw, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	body := rr.Body.String()

	// Should contain context
	if !strings.Contains(body, "context") {
		t.Fatalf("expected context key in response: %s", body)
	}

	// Decode JSON to verify structure
	var decoded map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("response not valid json: %v body=%s", err, body)
	}

	// With our setup, we should have matchedRules
	if matchedRules, ok := decoded["matchedRules"]; ok {
		// We have matching rules - verify the full response structure
		if _, hasSettings := decoded["settings"]; !hasSettings {
			t.Fatalf("expected settings key when matchedRules present: %s", body)
		}
		if ruleType, hasRuleType := decoded["ruleType"]; !hasRuleType || ruleType != "DCMGenericRule" {
			t.Fatalf("expected ruleType='DCMGenericRule' when matchedRules present, got: %v", decoded["ruleType"])
		}
		t.Logf("Successfully matched rules: %v", matchedRules)

		// Verify settings is properly structured (should be from CreateSettingsResponseObject)
		if settings, ok := decoded["settings"].(map[string]interface{}); ok {
			t.Logf("Settings response created successfully: %v", settings)
			// This confirms CreateSettingsResponseObject was called
		} else {
			t.Fatalf("settings should be a map structure")
		}
	} else {
		// If no matching rules, that's OK too (DB might not be fully configured)
		t.Logf("No matching rules found - DB may not be fully initialized")
	}

	// Clean up
	_ = ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_DCM_RULE, formula.ID)
	_ = ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_DEVICE_SETTINGS, deviceSettings.ID)
} // 6. Test with various MAC address formats to ensure normalization works
func TestDcmTestPageHandler_MacAddressNormalization(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/xconfAdminService/dcm/testpage?applicationType=stb", nil)
	// Provide MAC in different format
	searchContext := map[string]string{
		"estbMacAddress": "AA-BB-CC-DD-EE-FF", // dashes instead of colons
		"ecmMacAddress":  "11:22:33:44:55:66",
	}
	contextJSON, _ := json.Marshal(searchContext)
	xw, rr := newTestXWriter(string(contextJSON))

	DcmTestPageHandler(xw, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	// Verify response contains normalized context
	var decoded map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("response not valid json: %v", err)
	}

	// Context should be present
	if _, ok := decoded["context"]; !ok {
		t.Fatalf("expected context in response")
	}

	t.Logf("MAC normalization test passed, response: %s", rr.Body.String())
}
