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

package tag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	taggingapi_config "github.com/rdkcentral/xconfadmin/taggingapi/config"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/stretchr/testify/assert"
)

func setupTestEnvironment() {
	if xhttp.WebConfServer == nil {
		xhttp.WebConfServer = &xhttp.WebconfigServer{}
	}
	if xhttp.WebConfServer.TaggingApiConfig == nil {
		xhttp.WebConfServer.TaggingApiConfig = &taggingapi_config.TaggingApiConfig{
			BatchLimit:  5000,
			WorkerCount: 20,
		}
	}
	if xhttp.WebConfServer.GroupServiceConnector == nil {
		xhttp.WebConfServer.GroupServiceConnector = &xhttp.GroupServiceConnector{
			BaseURL: "http://localhost:9999",
			Client: &xhttp.HttpClient{
				Client: &http.Client{}, // Create a proper http.Client
			},
		}
	}
	if xhttp.WebConfServer.GroupServiceSyncConnector == nil {
		xhttp.WebConfServer.GroupServiceSyncConnector = &xhttp.GroupServiceSyncConnector{}
	}
}

func TestGetTagsByMemberHandler_MissingMember(t *testing.T) {
	setupTestEnvironment()
	req := httptest.NewRequest("GET", "/tags/by-member/", nil)
	w := httptest.NewRecorder()
	GetTagsByMemberHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "member is not specified")
}

func TestGetTagMembersHandler_MissingTag(t *testing.T) {
	setupTestEnvironment()
	req := httptest.NewRequest("GET", "/tags//members", nil)
	w := httptest.NewRecorder()
	GetTagMembersHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "tag is not specified")
}

func TestAddMembersToTagHandler_MissingTag(t *testing.T) {
	setupTestEnvironment()
	req := httptest.NewRequest("POST", "/tags//members", nil)
	w := httptest.NewRecorder()
	AddMembersToTagHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "tag is not specified")
}

func TestAddMembersToTagHandler_EmptyList(t *testing.T) {
	setupTestEnvironment()
	members := []string{}
	body, _ := json.Marshal(members)
	req := httptest.NewRequest("POST", "/tags/test-tag/members", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{common.Tag: "test-tag"})
	recorder := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{ResponseWriter: recorder}
	xw.SetBody(string(body))
	AddMembersToTagHandler(xw, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "member list is empty")
}

func TestAddMembersToTagHandler_ExceedsBatchSize(t *testing.T) {
	setupTestEnvironment()
	members := make([]string, MaxBatchSizeV2+1)
	for i := range members {
		members[i] = fmt.Sprintf("member%d", i)
	}
	body, _ := json.Marshal(members)
	req := httptest.NewRequest("POST", "/tags/test-tag/members", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{common.Tag: "test-tag"})
	recorder := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{ResponseWriter: recorder}
	xw.SetBody(string(body))
	AddMembersToTagHandler(xw, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "exceeds maximum")
}

func TestRemoveMembersFromTagHandler_MissingTag(t *testing.T) {
	setupTestEnvironment()
	req := httptest.NewRequest("DELETE", "/tags//members", nil)
	w := httptest.NewRecorder()
	RemoveMembersFromTagHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "tag is not specified")
}

func TestRemoveMembersFromTagHandler_EmptyList(t *testing.T) {
	setupTestEnvironment()
	members := []string{}
	body, _ := json.Marshal(members)
	req := httptest.NewRequest("DELETE", "/tags/test-tag/members", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{common.Tag: "test-tag"})
	w := httptest.NewRecorder()
	RemoveMembersFromTagHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "member list is empty")
}

func TestRemoveMemberFromTagHandler_MissingTag(t *testing.T) {
	setupTestEnvironment()
	req := httptest.NewRequest("DELETE", "/tags//members/member1", nil)
	w := httptest.NewRecorder()
	RemoveMemberFromTagHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "tag is not specified")
}

func TestRemoveMemberFromTagHandler_MissingMember(t *testing.T) {
	setupTestEnvironment()
	req := httptest.NewRequest("DELETE", "/tags/test-tag/members/", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "test-tag"})
	w := httptest.NewRecorder()
	RemoveMemberFromTagHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "member is not specified")
}

func TestGetTagByIdHandler_MissingTag(t *testing.T) {
	setupTestEnvironment()
	req := httptest.NewRequest("GET", "/tags/", nil)
	w := httptest.NewRecorder()
	GetTagByIdHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "tag is not specified")
}

func TestDeleteTagHandler_MissingTag(t *testing.T) {
	setupTestEnvironment()
	req := httptest.NewRequest("DELETE", "/tags/", nil)
	w := httptest.NewRecorder()
	DeleteTagHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "tag is not specified")
}

func TestParsePaginationParams_Default(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	params, err := parsePaginationParams(req)
	assert.NoError(t, err)
	assert.Equal(t, DefaultPageSizeV2, params.Limit)
	assert.Equal(t, "", params.Cursor)
}

func TestParsePaginationParams_ExceedsMax(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/test?limit=%d", MaxPageSizeV2+1), nil)
	params, err := parsePaginationParams(req)
	assert.Error(t, err)
	assert.Nil(t, params)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

func TestParsePaginationParams_InvalidLimit(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?limit=invalid", nil)
	params, err := parsePaginationParams(req)
	assert.Error(t, err)
	assert.Nil(t, params)
	assert.Contains(t, err.Error(), "invalid limit parameter")
}

func TestParsePaginationParams_NegativeLimit(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?limit=-1", nil)
	params, err := parsePaginationParams(req)
	assert.Error(t, err)
	assert.Nil(t, params)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestTagHandlerConstants(t *testing.T) {
	assert.Equal(t, 1000, TagMemberLimit)
	assert.Contains(t, RequestBodyReadErrorMsg, "request body unmarshall error")
	assert.Contains(t, NotSpecifiedErrorMsg, "is not specified")
	assert.Contains(t, EmptyListErrorMsg, "list is empty")
}

// Test GetTagsByMemberHandler success cases
func TestGetTagsByMemberHandler_WithValidMember(t *testing.T) {
	setupTestEnvironment()
	req := httptest.NewRequest("GET", "/tags/by-member/test-member", nil)
	req = mux.SetURLVars(req, map[string]string{common.Member: "test-member"})
	w := httptest.NewRecorder()
	GetTagsByMemberHandler(w, req)

	// Should return OK with empty or populated array
	assert.Equal(t, http.StatusOK, w.Code)
	var tags []string
	err := json.Unmarshal(w.Body.Bytes(), &tags)
	assert.NoError(t, err)
}
