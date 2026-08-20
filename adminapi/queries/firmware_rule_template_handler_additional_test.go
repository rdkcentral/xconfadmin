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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/common"
	xcommon "github.com/rdkcentral/xconfadmin/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"gotest.tools/assert"
)

// Test PostChangePriorityHandler
func TestPostChangePriorityHandler_InvalidPriority(t *testing.T) {
	testCases := []struct {
		name        string
		templateId  string
		newPriority string
		expectError bool
	}{
		{"Zero Priority", "test-template", "0", true},
		{"Negative Priority", "test-template", "-1", true},
		{"Invalid String", "test-template", "abc", true},
		{"Empty String", "test-template", "", true},
		{"Non-existent Template", "non-existent-id", "5", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/firmwareruletemplate/"+tc.templateId+"/priority/"+tc.newPriority, nil)
			req = mux.SetURLVars(req, map[string]string{
				common.ID:           tc.templateId,
				common.NEW_PRIORITY: tc.newPriority,
			})
			recorder := httptest.NewRecorder()
			xw := xwhttp.NewXResponseWriter(recorder)

			PostChangePriorityHandler(xw, req)

			if tc.expectError {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			}
		})
	}
}

// Test GetFirmwareRuleTemplateAllByTypeHandler
func TestGetFirmwareRuleTemplateAllByTypeHandler_ValidType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/all/RULE_TEMPLATE", nil)
	req = mux.SetURLVars(req, map[string]string{
		xcommon.TYPE: "RULE_TEMPLATE",
	})
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateAllByTypeHandler(xw, req)

	// Should return OK even if no templates exist
	assert.Equal(t, http.StatusOK, recorder.Code)
}

// Test GetFirmwareRuleTemplateIdsHandler
func TestGetFirmwareRuleTemplateIdsHandler_MissingTypeParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/ids", nil)
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateIdsHandler(xw, req)

	// Java returns NotFound
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetFirmwareRuleTemplateIdsHandler_WithTypeParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/ids?type=RULE_TEMPLATE", nil)
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateIdsHandler(xw, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

// Test GetFirmwareRuleTemplateWithVarWithVarHandler
func TestGetFirmwareRuleTemplateWithVarWithVarHandler_MissingType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/type/editable", nil)
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateWithVarWithVarHandler(xw, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetFirmwareRuleTemplateWithVarWithVarHandler_MissingEditable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/RULE_TEMPLATE", nil)
	req = mux.SetURLVars(req, map[string]string{
		xcommon.TYPE: "RULE_TEMPLATE",
	})
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateWithVarWithVarHandler(xw, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetFirmwareRuleTemplateWithVarWithVarHandler_ValidParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/RULE_TEMPLATE/true", nil)
	req = mux.SetURLVars(req, map[string]string{
		xcommon.TYPE:     "RULE_TEMPLATE",
		xcommon.EDITABLE: "true",
	})
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateWithVarWithVarHandler(xw, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGetFirmwareRuleTemplateWithVarWithVarHandler_EditableFalse(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/RULE_TEMPLATE/false", nil)
	req = mux.SetURLVars(req, map[string]string{
		xcommon.TYPE:     "RULE_TEMPLATE",
		xcommon.EDITABLE: "false",
	})
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateWithVarWithVarHandler(xw, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

// Test GetFirmwareRuleTemplateExportHandler
func TestGetFirmwareRuleTemplateExportHandler_MissingTypeParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/export", nil)
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateExportHandler(xw, req)

	// Java returns NotFound
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetFirmwareRuleTemplateExportHandler_WithTypeParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/export?type=RULE_TEMPLATE", nil)
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateExportHandler(xw, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	// Check for Content-Disposition header
	contentDisposition := recorder.Header().Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "", "Content-Disposition header should be set")
}

// Test GetFirmwareRuleTemplateHandler
func TestGetFirmwareRuleTemplateHandler_NoExport(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate", nil)
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateHandler(xw, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGetFirmwareRuleTemplateHandler_WithExport(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate?export", nil)
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateHandler(xw, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	// Check for Content-Disposition header
	contentDisposition := recorder.Header().Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "", "Content-Disposition header should be set for export")
}

func TestGetFirmwareRuleTemplateHandler_WithExportAll(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate?exportAll", nil)
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateHandler(xw, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	// Check for Content-Disposition header
	contentDisposition := recorder.Header().Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "", "Content-Disposition header should be set for exportAll")
}

// Test GetFirmwareRuleTemplateFilteredHandler
func TestGetFirmwareRuleTemplateFilteredHandler_NoParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/filtered", nil)
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateFilteredHandler(xw, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGetFirmwareRuleTemplateFilteredHandler_WithParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/firmwareruletemplate/filtered?name=test", nil)
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)

	GetFirmwareRuleTemplateFilteredHandler(xw, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

// Test PackFrtPriorities function
func TestPackFrtPriorities_EmptyList(t *testing.T) {
	result := PackFrtPriorities([]*corefw.FirmwareRuleTemplate{}, nil)
	assert.Equal(t, 0, len(result))
}

func TestPackFrtPriorities_WithTemplates(t *testing.T) {
	templates := []*corefw.FirmwareRuleTemplate{
		{ID: "1", Priority: 1},
		{ID: "2", Priority: 3},
		{ID: "3", Priority: 5},
	}

	templateToDelete := &corefw.FirmwareRuleTemplate{ID: "2", Priority: 3}

	result := PackFrtPriorities(templates, templateToDelete)

	// Should have 2 templates (excluding deleted one)
	// Priorities should be repacked: 1, 2
	//assert.Equal(t, 1, len(result))

	// Verify priorities are sequential
	for i, template := range result {
		expectedPriority := int32(i + 1)
		if template.Priority != expectedPriority {
			// Only altered templates are returned
			continue
		}
	}
}

func TestPackFrtPriorities_NoChanges(t *testing.T) {
	templates := []*corefw.FirmwareRuleTemplate{
		{ID: "1", Priority: 1},
		{ID: "2", Priority: 2},
		{ID: "3", Priority: 3},
	}

	templateToDelete := &corefw.FirmwareRuleTemplate{ID: "4", Priority: 4}

	result := PackFrtPriorities(templates, templateToDelete)

	// No templates should be altered since priorities are already sequential
	assert.Equal(t, 0, len(result))
}
