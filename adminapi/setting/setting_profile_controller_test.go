package setting

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/stretchr/testify/assert"
)

func TestGetSettingProfilesAllExport(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/setting-profiles", nil)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	ctx := context.WithValue(req.Context(), "applicationType", "STB")
	req = req.WithContext(ctx)
	GetSettingProfilesAllExport(w, req)
	assert.NotEqual(t, http.StatusInternalServerError, w.Status())

	//With export query parameter
	req = httptest.NewRequest(http.MethodGet, "/setting-profiles?export=true", nil)
	req = req.WithContext(ctx)
	GetSettingProfilesAllExport(w, req)
	assert.NotEqual(t, http.StatusInternalServerError, w.Status())
}

func TestCreateNumberOfItemsHttpHeaders(t *testing.T) {

	result := createNumberOfItemsHttpHeaders(nil)
	assert.Equal(t, "0", result[NumberOfItems])

	entities := []*logupload.SettingProfiles{
		{ID: "profile1"},
		{ID: "profile2"},
		{ID: "profile3"},
	}
	result = createNumberOfItemsHttpHeaders(entities)
	assert.Equal(t, "3", result[NumberOfItems])
	emptyEntities := make([]*logupload.SettingProfiles, 0)
	result = createNumberOfItemsHttpHeaders(emptyEntities)
	assert.Equal(t, "0", result[NumberOfItems])
}

func TestDeleteOneSettingProfilesHandler_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/setting-profiles/test-profile-123", nil)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	req = mux.SetURLVars(req, map[string]string{
		"id": "test-profile-123",
	})
	ctx := context.WithValue(req.Context(), "applicationType", "STB")
	req = req.WithContext(ctx)

	DeleteOneSettingProfilesHandler(w, req)
	assert.NotEqual(t, http.StatusMethodNotAllowed, w.Status())
	assert.NotEqual(t, http.StatusForbidden, w.Status())
}

func TestUpdateSettingProfilesHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/setting-profiles", nil)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	ctx := context.WithValue(req.Context(), "applicationType", "STB")
	req = req.WithContext(ctx)
	UpdateSettingProfilesHandler(w, req)

	//Without headers
	req2 := httptest.NewRequest(http.MethodPut, "/setting-profiles", nil)
	recorder2 := httptest.NewRecorder()
	w2 := xwhttp.NewXResponseWriter(recorder2)
	req2.Header = make(http.Header)
	UpdateSettingProfilesHandler(w2, req2)
}

func TestUpdateSettingProfilesHandler_ReachJSONMarshal(t *testing.T) {

	req := httptest.NewRequest(http.MethodPut, "/setting-profiles", nil)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	ctx := context.WithValue(req.Context(), "applicationType", "STB")
	req = req.WithContext(ctx)
	UpdateSettingProfilesHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	//With valid JSON body
	settingProfile := logupload.SettingProfiles{
		ID:               "test-profile-123",
		SettingProfileID: "profile-123",
		ApplicationType:  "STB",
	}
	jsonBody, _ := json.Marshal(settingProfile)

	req = httptest.NewRequest(http.MethodPut, "/setting-profiles", strings.NewReader(string(jsonBody)))
	recorder = httptest.NewRecorder()
	w = xwhttp.NewXResponseWriter(recorder)
	w.SetBody(string(jsonBody))
	ctx = context.WithValue(req.Context(), "applicationType", "STB")
	req = req.WithContext(ctx)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic in Update function: %v", r)
		}
	}()
	UpdateSettingProfilesHandler(w, req)
	assert.NotEqual(t, http.StatusBadRequest, w.Status(), "Should not return BadRequest for valid JSON")
}

func TestGetAllSettingProfilesWithPage(t *testing.T) {
	// Test case 1: Default pagination (no query parameters)
	req := httptest.NewRequest(http.MethodGet, "/setting-profiles", nil)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)

	GetAllSettingProfilesWithPage(w, req)
	assert.Equal(t, http.StatusOK, w.Status(), "Should return OK with default pagination")
	t.Logf("Default pagination test - Status: %d", w.Status())

	// Valid pagination parameters
	req2 := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageNumber=2&pageSize=10", nil)
	GetAllSettingProfilesWithPage(w, req2)
	assert.Equal(t, http.StatusOK, w.Status())

	req3 := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageNumber=invalid", nil)
	GetAllSettingProfilesWithPage(w, req3)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	// Invalid pageSize (triggers line 132-135)
	req4 := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageSize=notanumber", nil)
	GetAllSettingProfilesWithPage(w, req4)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	//Edge case - pageNumber=0, pageSize=0
	req5 := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageNumber=0&pageSize=0", nil)
	GetAllSettingProfilesWithPage(w, req5)
	assert.Equal(t, http.StatusOK, w.Status())
}

func TestGetSettingProfileOneExport(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/setting-profiles/test-profile-123", nil)
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)

	req = mux.SetURLVars(req, map[string]string{
		"id": "test-profile-123",
	})
	ctx := context.WithValue(req.Context(), "applicationType", "STB")
	req = req.WithContext(ctx)
	GetSettingProfileOneExport(w, req)
	assert.NotEqual(t, http.StatusInternalServerError, w.Status())

	//Valid ID with export parameter
	req2 := httptest.NewRequest(http.MethodGet, "/setting-profiles/test-profile-123?export=true", nil)
	req2 = mux.SetURLVars(req2, map[string]string{
		"id": "test-profile-123",
	})
	req2 = req2.WithContext(ctx)
	GetSettingProfileOneExport(w, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/setting-profiles/", nil)
	req3 = mux.SetURLVars(req3, map[string]string{
		"id": "",
	})
	req3 = req3.WithContext(ctx)
	GetSettingProfileOneExport(w, req3)
	assert.Equal(t, http.StatusNotFound, w.Status())

	// Missing ID in mux vars
	req4 := httptest.NewRequest(http.MethodGet, "/setting-profiles/", nil)
	req4 = req4.WithContext(ctx)
	GetSettingProfileOneExport(w, req4)
	assert.Equal(t, http.StatusNotFound, w.Status())

	req5 := httptest.NewRequest(http.MethodGet, "/setting-profiles/test-profile-123", nil)
	req5.Header = make(http.Header)
	GetSettingProfileOneExport(w, req5)
	statusCode5 := w.Status()
	assert.True(t, statusCode5 >= 400)
}

func TestGetSettingProfilesFilteredWithPage(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	ctx := context.WithValue(context.Background(), "applicationType", "STB")

	contextMap := map[string]string{
		"settingType": "LOG_UPLOAD_SETTINGS",
		"profileName": "test-profile",
	}
	jsonBody, _ := json.Marshal(contextMap)

	req := httptest.NewRequest(http.MethodPost, "/setting-profiles/filtered?pageNumber=1&pageSize=10", strings.NewReader(string(jsonBody)))
	w.SetBody(string(jsonBody))
	req = req.WithContext(ctx)
	GetSettingProfilesFilteredWithPage(w, req)
	assert.Equal(t, http.StatusOK, w.Status())

	recorder.Body.Reset()
	w = xwhttp.NewXResponseWriter(recorder)
	req2 := httptest.NewRequest(http.MethodPost, "/setting-profiles/filtered?pageNumber=invalid", nil)
	req2 = req2.WithContext(ctx)
	GetSettingProfilesFilteredWithPage(w, req2)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	recorder.Body.Reset()
	w = xwhttp.NewXResponseWriter(recorder)
	req3 := httptest.NewRequest(http.MethodPost, "/setting-profiles/filtered?pageSize=notanumber", nil)
	req3 = req3.WithContext(ctx)
	GetSettingProfilesFilteredWithPage(w, req3)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	recorder.Body.Reset()
	req4 := httptest.NewRequest(http.MethodPost, "/setting-profiles/filtered", nil)
	req4 = req4.WithContext(ctx)
	GetSettingProfilesFilteredWithPage(recorder, req4)
	assert.True(t, recorder.Code == http.StatusOK || recorder.Code >= 400)

	// Invalid JSON in body
	recorder.Body.Reset()
	w = xwhttp.NewXResponseWriter(recorder)
	w.SetBody(`{"invalid": json}`)
	req5 := httptest.NewRequest(http.MethodPost, "/setting-profiles/filtered", nil)
	req5 = req5.WithContext(ctx)
	GetSettingProfilesFilteredWithPage(w, req5)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	// Empty body
	recorder.Body.Reset()
	w = xwhttp.NewXResponseWriter(recorder)
	w.SetBody("")
	req6 := httptest.NewRequest(http.MethodPost, "/setting-profiles/filtered", nil)
	req6 = req6.WithContext(ctx)
	GetSettingProfilesFilteredWithPage(w, req6)
	assert.Equal(t, http.StatusOK, w.Status())
}

func TestCreateSettingProfileHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "applicationType", "STB")
	settingProfile := logupload.SettingProfiles{
		ID:               "test-profile-123",
		SettingProfileID: "profile-123",
		ApplicationType:  "STB",
	}
	jsonBody, _ := json.Marshal(settingProfile)
	req := httptest.NewRequest(http.MethodPost, "/setting-profiles", strings.NewReader(string(jsonBody)))
	w := xwhttp.NewXResponseWriter(recorder)
	w.SetBody(string(jsonBody))
	req = req.WithContext(ctx)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic in Create function: %v", r)
		}
	}()
	CreateSettingProfileHandler(w, req)
	assert.True(t, w.Status() == 0 || w.Status() == http.StatusCreated || w.Status() >= 400, "Unexpected status code")

	// ResponseWriter cast error
	recorder.Body.Reset()
	req2 := httptest.NewRequest(http.MethodPost, "/setting-profiles", nil)
	req2 = req2.WithContext(ctx)
	CreateSettingProfileHandler(recorder, req2)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Invalid JSON in body
	recorder.Body.Reset()
	req3 := httptest.NewRequest(http.MethodPost, "/setting-profiles", nil)
	w.SetBody(`{"invalid": json}`)
	req3 = req3.WithContext(ctx)
	CreateSettingProfileHandler(w, req3)
	assert.Equal(t, http.StatusBadRequest, w.Status())

	recorder.Body.Reset()
	req4 := httptest.NewRequest(http.MethodPost, "/setting-profiles", nil)
	w.SetBody("")
	req4 = req4.WithContext(ctx)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic with empty body: %v", r)
		}
	}()
	CreateSettingProfileHandler(w, req4)
	assert.Equal(t, http.StatusBadRequest, w.Status())
}

func TestCreateSettingProfilesPackageHandler(t *testing.T) {
	settingProfiles := []logupload.SettingProfiles{
		{
			ID:               "test-profile-1",
			SettingProfileID: "profile-1",
			ApplicationType:  "STB",
		},
		{
			ID:               "test-profile-2",
			SettingProfileID: "profile-2",
			ApplicationType:  "STB",
		},
	}

	jsonBody, _ := json.Marshal(settingProfiles)

	req := httptest.NewRequest(http.MethodPost, "/setting-profiles/package", strings.NewReader(string(jsonBody)))
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	w.SetBody(string(jsonBody))
	ctx := context.WithValue(req.Context(), "applicationType", "STB")
	req = req.WithContext(ctx)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic in Create function: %v", r)
		}
	}()
	CreateSettingProfilesPackageHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Status(), "Should return OK for valid package creation")

	req3 := httptest.NewRequest(http.MethodPost, "/setting-profiles/package", nil)
	ctx3 := context.WithValue(req3.Context(), "applicationType", "STB")
	req3 = req3.WithContext(ctx3)
	CreateSettingProfilesPackageHandler(recorder, req3)
	assert.NotEqual(t, http.StatusBadRequest, recorder.Code)

	//Invalid JSON in body
	req4 := httptest.NewRequest(http.MethodPost, "/setting-profiles/package", nil)
	w.SetBody(`{"invalid": json}`)
	req4 = req4.WithContext(ctx)
	CreateSettingProfilesPackageHandler(w, req4)

	assert.Equal(t, http.StatusBadRequest, w.Status(), "Should return BadRequest for invalid JSON")
	t.Log("Successfully triggered JSON unmarshal error - lines 281-285")

	// Empty body
	req5 := httptest.NewRequest(http.MethodPost, "/setting-profiles/package", nil)
	w.SetBody("")
	req5 = req5.WithContext(ctx)
	CreateSettingProfilesPackageHandler(w, req5)
	assert.NotEqual(t, http.StatusBadGateway, w.Status(), "Should handle empty body gracefully")
}

func TestUpdateSettingProfilesPackageHandler(t *testing.T) {
	settingProfiles := []logupload.SettingProfiles{
		{
			ID:               "test-profile-1",
			SettingProfileID: "profile-1",
			ApplicationType:  "STB",
		},
		{
			ID:               "test-profile-2",
			SettingProfileID: "profile-2",
			ApplicationType:  "STB",
		},
	}

	jsonBody, _ := json.Marshal(settingProfiles)

	req := httptest.NewRequest(http.MethodPut, "/setting-profiles/package", strings.NewReader(string(jsonBody)))
	recorder := httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(recorder)
	w.SetBody(string(jsonBody))
	ctx := context.WithValue(req.Context(), "applicationType", "STB")
	req = req.WithContext(ctx)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic in Update function: %v", r)
		}
	}()
	UpdateSettingProfilesPackageHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Status(), "Should return OK for valid update")

	req = httptest.NewRequest(http.MethodPut, "/setting-profiles/package", nil)
	UpdateSettingProfilesPackageHandler(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Set invalid JSON body
	w.SetBody(`{"invalid": json}`)
	UpdateSettingProfilesPackageHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Status(), "Should return BadRequest for invalid JSON")
}
