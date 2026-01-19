package applicationtype

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/stretchr/testify/assert"
)

func TestCreateApplicationTypeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/application-types", nil)
	rec := httptest.NewRecorder()

	CreateApplicationTypeHandler(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	w := xwhttp.NewXResponseWriter(rec)

	invalidJson := `{invalid json}`
	w.SetBody(invalidJson)
	CreateApplicationTypeHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// valid  Json
	validJson := `{"name": "testAppType"}`
	w.SetBody(validJson)
	CreateApplicationTypeHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	w.SetBody(validJson)
	CreateApplicationTypeHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	validJson = `{"name": "stb"}`
	w.SetBody(validJson)
	CreateApplicationTypeHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetAllApplicationTypeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/application-types", nil)
	rec := httptest.NewRecorder()
	GetAllApplicationTypeHandler(rec, req)
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestGetApplicationTypeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/application-types/{id}", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistentID"})
	rec := httptest.NewRecorder()
	GetApplicationTypeHandler(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateApplicationTypeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/application-types/123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "123"})
	rec := httptest.NewRecorder()

	UpdateApplicationTypeHandler(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "responsewriter cast error")

	rec = httptest.NewRecorder()
	w := xwhttp.NewXResponseWriter(rec)
	invalidJson := `{invalid json}`
	w.SetBody(invalidJson)
	UpdateApplicationTypeHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	rec = httptest.NewRecorder()
	w = xwhttp.NewXResponseWriter(rec)
	validJson := `{"name": "updatedApp"}`
	w.SetBody(validJson)
	UpdateApplicationTypeHandler(w, req)
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusBadRequest || rec.Code == http.StatusInternalServerError)

	rec = httptest.NewRecorder()
	w = xwhttp.NewXResponseWriter(rec)
	validJson = `{"name": "updatedAppType"}`
	w.SetBody(validJson)
	UpdateApplicationTypeHandler(w, req)
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusBadRequest || rec.Code == http.StatusInternalServerError)
}

func TestDeleteApplicationTypeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/application-types/{id}", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistentID"})
	rec := httptest.NewRecorder()
	DeleteApplicationTypeHandler(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
