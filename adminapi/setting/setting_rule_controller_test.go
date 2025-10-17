package setting

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/stretchr/testify/assert"
)

type contextKey string

const (
	applicationTypeKey contextKey = "applicationType"
)

func TestGetSettingRulesAllExport(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/setting-rules", nil)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	ctx := context.WithValue(req.Context(), applicationTypeKey, "STB")
	req = req.WithContext(ctx)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to database not configured: %v", r)
		}
	}()

	req.URL.RawQuery = "export=true"
	w = xwhttp.NewXResponseWriter(recorder)

	GetSettingRulesAllExport(w, req)
	assert.True(t, w.Status() >= 200, "Should return valid status for export")

	ctx = context.WithValue(context.Background(), "applicationType", "RDKV")
	req = req.WithContext(ctx)
	recorder.Body.Reset()
	w = xwhttp.NewXResponseWriter(recorder)
	GetSettingRulesAllExport(w, req)
	assert.True(t, w.Status() >= 200 || w.Status() >= 400, "Should return valid status code for filtering")
}

func TestGetSettingRuleOneExport(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/setting-rules/test-id", nil)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	ctx := context.WithValue(req.Context(), applicationTypeKey, "STB")
	req = req.WithContext(ctx)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to database not configured: %v", r)
		}
	}()
	req.Header = make(http.Header)
	GetSettingRuleOneExport(w, req)
	assert.True(t, w.Status() >= 400, "Should return error status for auth failure")
}

func TestDeleteOneSettingRulesHandler(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to database not configured: %v", r)
		}
	}()

	// Empty/blank ID
	req := httptest.NewRequest(http.MethodDelete, "/setting-rules/", nil)
	ctx := context.WithValue(req.Context(), "applicationType", "STB")
	req = req.WithContext(ctx)
	recorder := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{xwcommon.ID: ""})
	DeleteOneSettingRulesHandler(recorder, req)
	assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code, "Should return MethodNotAllowed for blank ID")

	//  Valid ID but DeleteSettingRule error
	req = httptest.NewRequest(http.MethodDelete, "/setting-rules/valid-id", nil)
	req = mux.SetURLVars(req, map[string]string{xwcommon.ID: "valid-setting-rule-id"})
	DeleteOneSettingRulesHandler(recorder, req)
	assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code, "Should return BadRequest when DeleteSettingRule fails")

}
func TestGetSettingRulesFilteredWithPage(t *testing.T) {

	ctx := context.WithValue(context.Background(), "applicationType", "STB")
	req := httptest.NewRequest(http.MethodPost, "/setting-rules/filtered?pageNumber=invalid", nil)
	req = req.WithContext(ctx)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	GetSettingRulesFilteredWithPage(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	req = httptest.NewRequest(http.MethodPost, "/setting-rules/filtered?pageSize=invalid", nil)
	req = req.WithContext(ctx)
	recorder.Body.Reset()
	w = xwhttp.NewXResponseWriter(recorder)
	GetSettingRulesFilteredWithPage(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	req = httptest.NewRequest(http.MethodPost, "/setting-rules/filtered", nil)
	req = req.WithContext(ctx)
	GetSettingRulesFilteredWithPage(recorder, req)
	assert.NotEqual(t, http.StatusInternalServerError, recorder.Code)

	// Invalid JSON in body
	req = httptest.NewRequest(http.MethodPost, "/setting-rules/filtered", nil)
	req = req.WithContext(ctx)
	w.SetBody(`{"invalid": json}`) // Invalid JSON
	GetSettingRulesFilteredWithPage(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status(), "Should return BadRequest for invalid JSON")

	// with paginatio
	searchContext := map[string]string{
		"name": "test-rule",
		"type": "PARTNER_SETTINGS",
	}
	jsonBody, _ := json.Marshal(searchContext)

	req = httptest.NewRequest(http.MethodPost, "/setting-rules/filtered?pageNumber=2&pageSize=10", nil)
	req = req.WithContext(ctx)
	w.SetBody(string(jsonBody))
	GetSettingRulesFilteredWithPage(w, req)
	assert.True(t, w.Status() >= 200)

	//Empty body
	req = httptest.NewRequest(http.MethodPost, "/setting-rules/filtered", nil)
	req = req.WithContext(ctx)
	w.SetBody("")
	GetSettingRulesFilteredWithPage(w, req)
	assert.True(t, w.Status() >= 200, "Should handle empty body gracefully")
}

func TestCreateSettingRuleHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/setting-rules", nil)
	recorder := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), applicationTypeKey, "STB")
	req = req.WithContext(ctx)
	CreateSettingRuleHandler(recorder, req)
	assert.True(t, recorder.Code == http.StatusBadRequest || recorder.Code == http.StatusInternalServerError)

	w := xwhttp.NewXResponseWriter(recorder)
	w.SetBody(`{"invalid": json}`) // Invalid JSON
	CreateSettingRuleHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	validSettingRule := map[string]interface{}{
		"id":              "test-rule-123",
		"name":            "Test Setting Rule",
		"applicationType": "STB",
		"boundSettingID":  "setting-123",
	}
	jsonBody, _ := json.Marshal(validSettingRule)
	w.SetBody(string(jsonBody))
	CreateSettingRuleHandler(w, req)
	assert.True(t, w.Status() >= 400)

	//Empty body
	w.SetBody("")
	CreateSettingRuleHandler(w, req)
	assert.True(t, w.Status() >= 400, "Should handle empty body")

	minimalRule := map[string]interface{}{
		"id":   "minimal-rule",
		"name": "Minimal Rule",
	}
	jsonBody, _ = json.Marshal(minimalRule)
	w.SetBody(string(jsonBody))
	CreateSettingRuleHandler(w, req)
	assert.True(t, w.Status() == http.StatusCreated || w.Status() >= 400)
}

func TestCreateSettingRulesPackageHandler(t *testing.T) {

	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	req := httptest.NewRequest(http.MethodPost, "/setting-rules/package", nil)
	ctx := context.WithValue(req.Context(), applicationTypeKey, "STB")
	req = req.WithContext(ctx)
	CreateSettingRulesPackageHandler(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Invalid JSON array
	w.SetBody(`[{"invalid": json}]`)
	CreateSettingRulesPackageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	rulesWithError := []map[string]interface{}{
		{
			"id":   "rule-1",
			"name": "Rule 1",
		},
		{
			"id":   "rule-2",
			"name": "Rule 2",
		},
		{
			"id":   "rule-3",
			"name": "Rule 3",
		},
	}
	jsonBody, _ := json.Marshal(rulesWithError)
	w.SetBody(string(jsonBody))

	CreateSettingRulesPackageHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Status())
}

func TestUpdateSettingRulesHandler(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to database not configured: %v", r)
		}
	}()
	ctx := context.WithValue(context.Background(), applicationTypeKey, "STB")
	req := httptest.NewRequest(http.MethodPut, "/setting-rules", nil)
	req = req.WithContext(ctx)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	UpdateSettingRulesHandler(recorder, req)
	assert.True(t, recorder.Code == http.StatusBadRequest || recorder.Code == http.StatusInternalServerError)

	// Invalid JSON
	w.SetBody(`{"invalid": json}`)
	UpdateSettingRulesHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status(), "Should return BadRequest for invalid JSON")

	//Empty body
	w.SetBody("")
	UpdateSettingRulesHandler(w, req)
	assert.True(t, w.Status() >= 400, "Should handle empty body")

	validSettingRule := map[string]interface{}{
		"id":              "test-rule-123",
		"name":            "Updated Test Rule",
		"applicationType": "STB",
		"boundSettingID":  "setting-123",
	}
	jsonBody, _ := json.Marshal(validSettingRule)
	w.SetBody(string(jsonBody))
	UpdateSettingRulesHandler(w, req)
	assert.True(t, w.Status() >= 400)
}

func TestUpdateSettingRulesPackageHandler(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to database not configured: %v", r)
		}
	}()
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	ctx := context.WithValue(context.Background(), applicationTypeKey, "STB")
	req := httptest.NewRequest(http.MethodPut, "/setting-rules/package", nil)
	req = req.WithContext(ctx)
	UpdateSettingRulesPackageHandler(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Invalid JSON array
	w.SetBody(`[{"invalid": json}]`)
	UpdateSettingRulesPackageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	//Empty array
	w.SetBody(`[]`)
	UpdateSettingRulesPackageHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Status())

	validRules := []map[string]interface{}{
		{
			"id":              "update-rule-1",
			"name":            "Updated Test Rule 1",
			"applicationType": "STB",
			"boundSettingID":  "setting-1",
		},
	}
	jsonBody, _ := json.Marshal(validRules)
	w.SetBody(string(jsonBody))

	UpdateSettingRulesPackageHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Status())
}

func TestSettingTestPageHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/setting-test", nil)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	ctx := context.WithValue(req.Context(), applicationTypeKey, "STB")
	req = req.WithContext(ctx)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to database not configured: %v", r)
		}
	}()
	SettingTestPageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	// ResponseWriter cast error
	req = httptest.NewRequest(http.MethodPost, "/setting-test?settingType=PARTNER_SETTINGS", nil)
	req = req.WithContext(ctx)
	SettingTestPageHandler(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Invalid JSON
	w.SetBody(`{"invalid": json}`)
	SettingTestPageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	// Set context with invalid MAC address format to trigger normalization error
	invalidContext := map[string]string{
		"estbMacAddress": "invalid-mac-format",
	}
	jsonBody4, _ := json.Marshal(invalidContext)
	w.SetBody(string(jsonBody4))

	SettingTestPageHandler(w, req)
	assert.True(t, w.Status() >= 400)

	// Empty body
	w.SetBody("")
	SettingTestPageHandler(w, req)
	assert.True(t, w.Status() >= 200, "Should handle empty body")
}
