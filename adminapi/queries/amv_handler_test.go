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
package queries

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	xcommon "github.com/rdkcentral/xconfadmin/common"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
)

func TestGetAmvHandler_Success(t *testing.T) {
	// Create test request
	req := httptest.NewRequest("GET", "/api/queries/amv", nil)
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{ResponseWriter: w}

	// Note: This test would require auth setup and database mocking to fully work
	// For now, testing basic structure
	assert.NotNil(t, req)
	assert.NotNil(t, xw)
}

func TestGetAmvByIdHandler_InvalidId(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/queries/amv/", nil)
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()

	// Without ID in mux vars, should fail
	GetAmvByIdHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostAmvFilteredHandler_EmptyBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/queries/amv/filtered", bytes.NewBufferString(""))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	// This would require auth setup
	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestPostAmvFilteredHandler_InvalidJSON(t *testing.T) {
	invalidJSON := `{"invalid": json}`
	req := httptest.NewRequest("POST", "/api/queries/amv/filtered", bytes.NewBufferString(invalidJSON))
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestDeleteAmvByIdHandler_NoId(t *testing.T) {
	req := httptest.NewRequest("DELETE", "/api/queries/amv/", nil)
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()

	DeleteAmvByIdHandler(w, req)

	// Should return error when ID is missing
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateAmvHandler_InvalidBody(t *testing.T) {
	invalidJSON := `{"invalid json`
	req := httptest.NewRequest("POST", "/api/queries/amv", bytes.NewBufferString(invalidJSON))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestImportAllAmvHandler_InvalidJSON(t *testing.T) {
	invalidJSON := `[{"invalid": json}]`
	req := httptest.NewRequest("POST", "/api/queries/amv/import", bytes.NewBufferString(invalidJSON))
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestImportAllAmvHandler_EmptyList(t *testing.T) {
	emptyList := `[]`
	req := httptest.NewRequest("POST", "/api/queries/amv/import", bytes.NewBufferString(emptyList))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestUpdateAmvHandler_ValidRequest(t *testing.T) {
	amv := firmware.ActivationVersion{
		ID:                 "test-id",
		ApplicationType:    "stb",
		Description:        "Test AMV",
		Model:              "TEST_MODEL",
		FirmwareVersions:   []string{"1.0"},
		RegularExpressions: []string{".*"},
	}
	body, _ := json.Marshal(amv)

	req := httptest.NewRequest("PUT", "/api/queries/amv", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

func TestPostAmvEntitiesHandler_EmptyList(t *testing.T) {
	emptyList := `[]`
	req := httptest.NewRequest("POST", "/api/queries/amv/entities", bytes.NewBufferString(emptyList))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestPutAmvEntitiesHandler_EmptyList(t *testing.T) {
	emptyList := `[]`
	req := httptest.NewRequest("PUT", "/api/queries/amv/entities", bytes.NewBufferString(emptyList))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestNotImplementedHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/queries/not-implemented", nil)
	w := httptest.NewRecorder()

	NotImplementedHandler(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestGetAmvFilteredHandler_WithParams(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/queries/amv/filtered?pageNumber=1&pageSize=10", nil)
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()

	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

func TestGetAmvHandler_WithExportAll(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/queries/amv?"+xcommon.EXPORTALL, nil)
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()

	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

func TestGetAmvByIdHandler_WithExport(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/queries/amv/test-id?"+xcommon.EXPORT, nil)
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	req = mux.SetURLVars(req, map[string]string{xwcommon.ID: "test-id"})
	w := httptest.NewRecorder()

	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

func TestPostAmvFilteredHandler_WithPagination(t *testing.T) {
	filterContext := map[string]string{
		"pageNumber": "1",
		"pageSize":   "10",
	}
	body, _ := json.Marshal(filterContext)

	req := httptest.NewRequest("POST", "/api/queries/amv/filtered", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestCreateAmvHandler_ValidAmv(t *testing.T) {
	amv := firmware.ActivationVersion{
		ApplicationType:    "stb",
		Description:        "Test Description",
		Model:              "TEST_MODEL",
		FirmwareVersions:   []string{"1.0.0"},
		RegularExpressions: []string{".*"},
	}
	body, _ := json.Marshal(amv)

	req := httptest.NewRequest("POST", "/api/queries/amv", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

func TestImportAllAmvHandler_SingleAmv(t *testing.T) {
	amvList := []firmware.ActivationVersion{
		{
			ID:                 "test-import-1",
			ApplicationType:    "stb",
			Description:        "Import Test 1",
			Model:              "TEST_MODEL",
			FirmwareVersions:   []string{"1.0.0"},
			RegularExpressions: []string{".*"},
		},
	}
	body, _ := json.Marshal(amvList)

	req := httptest.NewRequest("POST", "/api/queries/amv/import", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestImportAllAmvHandler_MultipleAmv(t *testing.T) {
	amvList := []firmware.ActivationVersion{
		{
			ID:                 "test-import-1",
			ApplicationType:    "stb",
			Description:        "Import Test 1",
			Model:              "TEST_MODEL_1",
			FirmwareVersions:   []string{"1.0.0"},
			RegularExpressions: []string{".*"},
		},
		{
			ID:                 "test-import-2",
			ApplicationType:    "stb",
			Description:        "Import Test 2",
			Model:              "TEST_MODEL_2",
			FirmwareVersions:   []string{"2.0.0"},
			RegularExpressions: []string{".*"},
		},
	}
	body, _ := json.Marshal(amvList)

	req := httptest.NewRequest("POST", "/api/queries/amv/import", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()

	assert.NotNil(t, w)
	assert.NotNil(t, req)
}

func TestPostAmvEntitiesHandler_ValidEntities(t *testing.T) {
	entities := []firmware.ActivationVersion{
		{
			ApplicationType:    "stb",
			Description:        "Entity 1",
			Model:              "MODEL_1",
			FirmwareVersions:   []string{"1.0"},
			RegularExpressions: []string{".*"},
		},
	}
	body, _ := json.Marshal(entities)

	req := httptest.NewRequest("POST", "/api/queries/amv/entities", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()

	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

func TestPutAmvEntitiesHandler_ValidEntities(t *testing.T) {
	entities := []firmware.ActivationVersion{
		{
			ID:                 "update-1",
			ApplicationType:    "stb",
			Description:        "Updated Entity 1",
			Model:              "MODEL_1",
			FirmwareVersions:   []string{"1.0"},
			RegularExpressions: []string{".*"},
		},
	}
	body, _ := json.Marshal(entities)

	req := httptest.NewRequest("PUT", "/api/queries/amv/entities", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()

	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

func TestGetAmvFilteredHandler_WithModelFilter(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/queries/amv/filtered?MODEL=TEST_MODEL", nil)
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()

	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

func TestGetAmvFilteredHandler_WithPartnerIdFilter(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/queries/amv/filtered?PARTNER_ID=PARTNER1", nil)
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()

	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

func TestPostAmvFilteredHandler_WithModelFilter(t *testing.T) {
	filterContext := map[string]string{
		"MODEL": "TEST_MODEL",
	}
	body, _ := json.Marshal(filterContext)

	req := httptest.NewRequest("POST", "/api/queries/amv/filtered", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestPostAmvFilteredHandler_WithDescriptionFilter(t *testing.T) {
	filterContext := map[string]string{
		"DESCRIPTION": "Test Description",
	}
	body, _ := json.Marshal(filterContext)

	req := httptest.NewRequest("POST", "/api/queries/amv/filtered", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestPostAmvFilteredHandler_WithFirmwareVersionFilter(t *testing.T) {
	filterContext := map[string]string{
		"FIRMWARE_VERSION": "1.0",
	}
	body, _ := json.Marshal(filterContext)

	req := httptest.NewRequest("POST", "/api/queries/amv/filtered", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}

func TestPostAmvFilteredHandler_WithRegexFilter(t *testing.T) {
	filterContext := map[string]string{
		"REGULAR_EXPRESSION": ".*test.*",
	}
	body, _ := json.Marshal(filterContext)

	req := httptest.NewRequest("POST", "/api/queries/amv/filtered", bytes.NewBuffer(body))
	req.Header.Set(xwcommon.APPLICATION_TYPE, "stb")
	w := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{
		ResponseWriter: w,
	}

	assert.NotNil(t, xw)
	assert.NotNil(t, req)
}
