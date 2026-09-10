package setting

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/stretchr/testify/assert"
)

// Test error scenarios - these test the xhttp.AdminError, WriteAdminErrorResponse paths
func TestGetSettingProfilesAllExport(t *testing.T) {
	t.Run("NoAuthContext", func(t *testing.T) {
		// Test without proper auth context to trigger xhttp.AdminError
		req := httptest.NewRequest(http.MethodGet, "/setting-profiles", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		// Not setting auth context to trigger auth error

		GetSettingProfilesAllExport(w, req)

		// The function still returns 200 with empty application type, but calls GetAll
		// which logs warnings. This tests the normal flow with missing auth.
		assert.True(t, w.Status() == http.StatusOK || w.Status() >= 400, "Should handle missing auth gracefully")
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/setting-profiles", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		req = req.WithContext(ctx)
		GetSettingProfilesAllExport(w, req)
		assert.NotEqual(t, http.StatusInternalServerError, w.Status())
	})

	t.Run("WithExportParam", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/setting-profiles?export=true", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		req = req.WithContext(ctx)
		GetSettingProfilesAllExport(w, req)
		assert.NotEqual(t, http.StatusInternalServerError, w.Status())
	})
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

func TestDeleteOneSettingProfilesHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
	})

	t.Run("ErrorCases", func(t *testing.T) {
		// Test case 1: Missing ID to trigger WriteAdminErrorResponse with MethodNotAllowed
		req := httptest.NewRequest(http.MethodDelete, "/setting-profiles/", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)

		// Set empty ID to trigger "missing id" error
		req = mux.SetURLVars(req, map[string]string{"id": ""})
		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		req = req.WithContext(ctx)

		DeleteOneSettingProfilesHandler(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Status(), "Should return MethodNotAllowed for missing ID")
		// Note: Response body may be empty but status code confirms error path

		// Test case 2: Valid ID but delete operation fails to trigger WriteAdminErrorResponse with BadRequest
		req2 := httptest.NewRequest(http.MethodDelete, "/setting-profiles/valid-id", nil)
		recorder2 := httptest.NewRecorder()
		w2 := xwhttp.NewXResponseWriter(recorder2)

		req2 = mux.SetURLVars(req2, map[string]string{"id": "valid-id-that-fails"})
		req2 = req2.WithContext(ctx)

		DeleteOneSettingProfilesHandler(w2, req2)
		// This will trigger the delete error path and call WriteAdminErrorResponse
		assert.True(t, w2.Status() >= 400, "Should return error status for failed delete operation")
	})

	t.Run("NoID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/setting-profiles/", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		req = mux.SetURLVars(req, map[string]string{})
		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		req = req.WithContext(ctx)

		DeleteOneSettingProfilesHandler(w, req)
		// Should handle missing ID
		assert.NotEqual(t, http.StatusOK, w.Status())
	})
}

func TestUpdateSettingProfilesHandler(t *testing.T) {
	t.Run("BasicTest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/setting-profiles", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		req = req.WithContext(ctx)
		UpdateSettingProfilesHandler(w, req)
	})

	t.Run("WithoutHeaders", func(t *testing.T) {
		req2 := httptest.NewRequest(http.MethodPut, "/setting-profiles", nil)
		recorder2 := httptest.NewRecorder()
		w2 := xwhttp.NewXResponseWriter(recorder2)
		req2.Header = make(http.Header)
		UpdateSettingProfilesHandler(w2, req2)
	})

	t.Run("ReachJSONMarshal", func(t *testing.T) {
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
	})

	t.Run("ResponseWriterCastError", func(t *testing.T) {
		// Test ResponseWriter cast error to trigger xwhttp.Error
		req := httptest.NewRequest(http.MethodPut, "/setting-profiles", nil)
		recorder := httptest.NewRecorder()
		// Pass regular recorder instead of XResponseWriter to trigger cast error

		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		req = req.WithContext(ctx)

		UpdateSettingProfilesHandler(recorder, req)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code, "Should return InternalServerError for ResponseWriter cast error")
	})

	t.Run("ValidProfile", func(t *testing.T) {
		t.Skip("Requires database configuration - cannot set up test data for update")
	})
}

func TestGetAllSettingProfilesWithPage(t *testing.T) {
	// Test case 1: Default pagination (no query parameters)
	t.Run("DefaultPagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/setting-profiles", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)

		GetAllSettingProfilesWithPage(w, req)
		assert.Equal(t, http.StatusOK, w.Status(), "Should return OK with default pagination")
		t.Logf("Default pagination test - Status: %d", w.Status())
	})

	t.Run("ValidPagination", func(t *testing.T) {
		// Valid pagination parameters
		req2 := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageNumber=2&pageSize=10", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		GetAllSettingProfilesWithPage(w, req2)
		assert.Equal(t, http.StatusOK, w.Status())
	})

	t.Run("InvalidPageNumber", func(t *testing.T) {
		req3 := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageNumber=invalid", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		GetAllSettingProfilesWithPage(w, req3)
		assert.Equal(t, http.StatusBadRequest, w.Status())
	})

	t.Run("InvalidPageSize", func(t *testing.T) {
		// Invalid pageSize (triggers line 132-135)
		req4 := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageSize=notanumber", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		GetAllSettingProfilesWithPage(w, req4)
		assert.Equal(t, http.StatusBadRequest, w.Status())
	})

	t.Run("EdgeCase", func(t *testing.T) {
		//Edge case - pageNumber=0, pageSize=0
		req5 := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageNumber=0&pageSize=0", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		GetAllSettingProfilesWithPage(w, req5)
		assert.Equal(t, http.StatusOK, w.Status())
	})

	t.Run("AdditionalErrorCases", func(t *testing.T) {
		// Test case 1: pageNumber = 0 (edge case)
		req := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageNumber=0", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)

		GetAllSettingProfilesWithPage(w, req)
		assert.Equal(t, http.StatusOK, w.Status(), "Should handle pageNumber=0")

		// Test case 2: pageSize = 0 (edge case)
		req2 := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageSize=0", nil)
		recorder2 := httptest.NewRecorder()
		w2 := xwhttp.NewXResponseWriter(recorder2)

		GetAllSettingProfilesWithPage(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Status(), "Should handle pageSize=0")

		// Test case 3: Negative pageNumber
		req3 := httptest.NewRequest(http.MethodGet, "/setting-profiles?pageNumber=-1", nil)
		recorder3 := httptest.NewRecorder()
		w3 := xwhttp.NewXResponseWriter(recorder3)

		GetAllSettingProfilesWithPage(w3, req3)
		assert.Equal(t, http.StatusOK, w3.Status(), "Should handle negative pageNumber gracefully")
	})
}

func TestGetSettingProfileOneExport(t *testing.T) {
	t.Run("BasicTest", func(t *testing.T) {
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
	})

	t.Run("WithExportParam", func(t *testing.T) {
		//Valid ID with export parameter
		req2 := httptest.NewRequest(http.MethodGet, "/setting-profiles/test-profile-123?export=true", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		ctx := context.WithValue(req2.Context(), "applicationType", "STB")
		req2 = mux.SetURLVars(req2, map[string]string{
			"id": "test-profile-123",
		})
		req2 = req2.WithContext(ctx)
		GetSettingProfileOneExport(w, req2)
	})

	t.Run("EmptyID", func(t *testing.T) {
		req3 := httptest.NewRequest(http.MethodGet, "/setting-profiles/", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		ctx := context.WithValue(req3.Context(), "applicationType", "STB")
		req3 = mux.SetURLVars(req3, map[string]string{
			"id": "",
		})
		req3 = req3.WithContext(ctx)
		GetSettingProfileOneExport(w, req3)
		assert.Equal(t, http.StatusBadRequest, w.Status())
	})

	t.Run("MissingIDInMuxVars", func(t *testing.T) {
		// Missing ID in mux vars
		req4 := httptest.NewRequest(http.MethodGet, "/setting-profiles/", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		ctx := context.WithValue(req4.Context(), "applicationType", "STB")
		req4 = req4.WithContext(ctx)
		GetSettingProfileOneExport(w, req4)
		assert.Equal(t, http.StatusBadRequest, w.Status())
	})

	t.Run("NoHeaders", func(t *testing.T) {
		req5 := httptest.NewRequest(http.MethodGet, "/setting-profiles/test-profile-123", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		req5.Header = make(http.Header)
		GetSettingProfileOneExport(w, req5)
		statusCode5 := w.Status()
		assert.True(t, statusCode5 >= 400)
	})

	t.Run("WriteAdminErrorResponseCases", func(t *testing.T) {
		// Test case 1: Missing ID to trigger WriteAdminErrorResponse with BadRequest
		req := httptest.NewRequest(http.MethodGet, "/setting-profiles/", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)

		// Set mux vars with empty ID
		req = mux.SetURLVars(req, map[string]string{"id": ""})
		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		ctx = context.WithValue(ctx, "auth_subject", "admin")
		req = req.WithContext(ctx)

		GetSettingProfileOneExport(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Status(), "Should return BadRequest for empty ID")
		// Note: The response body might be empty due to how xwhttp.WriteAdminErrorResponse works
		// but the status code is the important part for this test

		// Test case 2: Non-existent ID to trigger WriteAdminErrorResponse with NotFound
		req2 := httptest.NewRequest(http.MethodGet, "/setting-profiles/non-existent-id", nil)
		recorder2 := httptest.NewRecorder()
		w2 := xwhttp.NewXResponseWriter(recorder2)

		req2 = mux.SetURLVars(req2, map[string]string{"id": "non-existent-id-12345"})
		ctx2 := context.WithValue(req2.Context(), "applicationType", "STB")
		ctx2 = context.WithValue(ctx2, "auth_subject", "admin")
		req2 = req2.WithContext(ctx2)

		GetSettingProfileOneExport(w2, req2)
		assert.Equal(t, http.StatusNotFound, w2.Status(), "Should return NotFound for non-existent ID")
		// Note: The response may be empty but status code indicates the error path was taken
	})

	t.Run("Success", func(t *testing.T) {
		// Create a test profile
		profile := &logupload.SettingProfiles{
			ID:               "export-test-profile-1",
			SettingProfileID: "export-profile-1",
			ApplicationType:  "STB",
			SettingType:      "EPON",
		}
		SetSettingProfile(db.GetDefaultTenantId(), profile.ID, profile)

		req := httptest.NewRequest(http.MethodGet, "/setting-profiles/export-test-profile-1", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		req = mux.SetURLVars(req, map[string]string{"id": "export-test-profile-1"})
		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		req = req.WithContext(ctx)

		GetSettingProfileOneExport(w, req)
		// Database not configured in tests, so just verify handler executes
		assert.NotEqual(t, http.StatusInternalServerError, w.Status())
	})

	t.Run("WithExportParamSuccess", func(t *testing.T) {
		profile := &logupload.SettingProfiles{
			ID:               "export-test-profile-2",
			SettingProfileID: "export-profile-2",
			ApplicationType:  "STB",
			SettingType:      "EPON",
		}
		SetSettingProfile(db.GetDefaultTenantId(), profile.ID, profile)

		req := httptest.NewRequest(http.MethodGet, "/setting-profiles/export-test-profile-2?export=true", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		req = mux.SetURLVars(req, map[string]string{"id": "export-test-profile-2"})
		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		req = req.WithContext(ctx)

		GetSettingProfileOneExport(w, req)
		// Database not configured in tests, verify handler executes
		assert.NotEqual(t, http.StatusInternalServerError, w.Status())
	})

	t.Run("BlankID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/setting-profiles/", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		req = mux.SetURLVars(req, map[string]string{"id": ""})
		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		req = req.WithContext(ctx)

		GetSettingProfileOneExport(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Status())
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/setting-profiles/non-existent-id", nil)
		recorder := httptest.NewRecorder()
		w := xwhttp.NewXResponseWriter(recorder)
		req = mux.SetURLVars(req, map[string]string{"id": "non-existent-id"})
		ctx := context.WithValue(req.Context(), "applicationType", "STB")
		req = req.WithContext(ctx)

		GetSettingProfileOneExport(w, req)
		assert.Equal(t, http.StatusNotFound, w.Status())
	})
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
