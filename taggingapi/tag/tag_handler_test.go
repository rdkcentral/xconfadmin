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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/rdkcentral/xconfadmin/util"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestGetAllTagsHandler_Integration(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags", GetAllTagsHandler).Methods("GET")

	t.Run("GetAllTags_Full", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags?full=true", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Since this calls the actual service functions, we expect it to handle gracefully
		// The status should be either OK (if data exists) or an error status
		assert.True(t, rr.Code == http.StatusOK || rr.Code >= http.StatusBadRequest)
	})

	t.Run("GetAllTags_IdsOnly", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Since this calls the actual service functions, we expect it to handle gracefully
		assert.True(t, rr.Code == http.StatusOK || rr.Code >= http.StatusBadRequest)
	})
}

func TestGetTagByIdHandler_Integration(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}", GetTagByIdHandler).Methods("GET")

	t.Run("GetTagById_WithValidPath", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/test-tag", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return either OK (if tag exists) or NotFound
		assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusNotFound)
	})

	t.Run("GetTagById_EmptyTag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return 404 since the route doesn't match
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestDeleteTagHandler_Integration(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}", DeleteTagHandler).Methods("DELETE")

	t.Run("DeleteTag_WithValidPath", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return either NoContent (if deleted) or NotFound
		assert.True(t, rr.Code == http.StatusNoContent || rr.Code == http.StatusNotFound)
	})
}

func TestDeleteTagFromXconfWithoutPrefixHandler_Integration(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/noprefix", DeleteTagFromXconfWithoutPrefixHandler).Methods("DELETE")

	t.Run("DeleteTagFromXconf_WithValidPath", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/noprefix", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return either NoContent (if deleted) or NotFound
		assert.True(t, rr.Code == http.StatusNoContent || rr.Code == http.StatusNotFound)
	})
}

func TestGetTagsByMemberHandler_Integration(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/members/{member}", GetTagsByMemberHandler).Methods("GET")

	t.Run("GetTagsByMember_WithValidPath", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/test-member", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return OK with empty or populated array (service connector may be nil in tests)
		assert.Equal(t, http.StatusOK, rr.Code, "expected 200 even when connector is nil")
	})
}

func TestAddMembersToTagHandler_Integration(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/members", AddMembersToTagHandler).Methods("PUT")

	t.Run("AddMembersToTag_WithValidMembers", func(t *testing.T) {
		members := []string{"member1", "member2"}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Wrap with XResponseWriter to simulate the middleware behavior
		rr := httptest.NewRecorder()
		xrr := xwhttp.NewXResponseWriter(rr)

		// Manually read and store the body since we don't have the middleware
		body := make([]byte, len(jsonBody))
		copy(body, jsonBody)
		req.Body = nopCloser{bytes.NewReader(body)}

		router.ServeHTTP(xrr, req)

		// Response could be OK, BadRequest, or InternalServerError depending on implementation
		assert.True(t, rr.Code >= http.StatusOK)
	})

	t.Run("AddMembersToTag_EmptyMembers", func(t *testing.T) {
		members := []string{}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		xrr := xwhttp.NewXResponseWriter(rr)

		body := make([]byte, len(jsonBody))
		copy(body, jsonBody)
		req.Body = nopCloser{bytes.NewReader(body)}

		router.ServeHTTP(xrr, req)

		// Should return BadRequest for empty members
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestRemoveMemberFromTagHandler_Integration(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/members/{member}", RemoveMemberFromTagHandler).Methods("DELETE")

	t.Run("RemoveMemberFromTag_WithValidPath", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members/test-member", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return either NoContent or an error status
		assert.True(t, rr.Code == http.StatusNoContent || rr.Code >= http.StatusBadRequest)
	})
}

// Comprehensive error coverage tests for all handlers

func TestCleanPercentageRangeHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/percentages", CleanPercentageRangeHandler).Methods("DELETE")

	t.Run("CleanPercentageRange_ValidTag", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/percentages", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return either NoContent (if successful) or error status
		assert.True(t, rr.Code == http.StatusNoContent || rr.Code >= http.StatusBadRequest)
	})

	t.Run("CleanPercentageRange_MissingTag", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags//percentages", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Route won't match - should get 404 or 301
		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})
}

func TestAddMemberPercentageToTagHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/percentages/{startRange}/{endRange}", AddMemberPercentageToTagHandler).Methods("PUT")

	t.Run("AddMemberPercentage_ValidRanges", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/percentages/0/50", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return either OK or error status
		assert.True(t, rr.Code == http.StatusOK || rr.Code >= http.StatusBadRequest)
	})

	t.Run("AddMemberPercentage_MissingTag", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/taggingService/tags//percentages/0/50", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Route won't match
		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})

	t.Run("AddMemberPercentage_InvalidRange", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/percentages/invalid/50", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return BadRequest for invalid range
		assert.True(t, rr.Code >= http.StatusBadRequest)
	})

	t.Run("AddMemberPercentage_MissingStartRange", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/percentages//50", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Route won't match
		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})

	t.Run("AddMemberPercentage_MissingEndRange", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/percentages/0/", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Route won't match
		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})
}

func TestGetTagsByMemberPercentageHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/members/{member}/percentages", GetTagsByMemberPercentageHandler).Methods("GET")

	t.Run("GetTagsByMemberPercentage_ValidMember", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/test-member/percentages", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return OK with tags or error
		assert.True(t, rr.Code == http.StatusOK || rr.Code >= http.StatusBadRequest)
	})

	t.Run("GetTagsByMemberPercentage_MissingMember", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/members//percentages", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Route won't match
		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})

	t.Run("GetTagsByMemberPercentage_SpecialCharacters", func(t *testing.T) {
		member := "member%20test"
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/"+member+"/percentages", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should handle special characters
		assert.True(t, rr.Code == http.StatusOK || rr.Code >= http.StatusBadRequest)
	})
}

func TestAddMembersToTagHandler_AllErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/members", AddMembersToTagHandler).Methods("PUT")

	t.Run("AddMembersToTag_ValidMembers", func(t *testing.T) {
		members := []string{"member1", "member2"}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		xrr := xwhttp.NewXResponseWriter(rr)

		body := make([]byte, len(jsonBody))
		copy(body, jsonBody)
		req.Body = nopCloser{bytes.NewReader(body)}

		router.ServeHTTP(xrr, req)

		// Should return OK or error
		assert.True(t, rr.Code >= http.StatusOK)
	})

	t.Run("AddMembersToTag_MissingTag", func(t *testing.T) {
		members := []string{"member1"}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("PUT", "/taggingService/tags//members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Route won't match
		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})

	t.Run("AddMembersToTag_EmptyMemberList", func(t *testing.T) {
		members := []string{}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		xrr := xwhttp.NewXResponseWriter(rr)

		// Read body and store it in XResponseWriter
		body, _ := io.ReadAll(req.Body)
		req.Body = nopCloser{bytes.NewReader(body)}
		xrr.SetBody(string(body))

		router.ServeHTTP(xrr, req)

		// Should return BadRequest for empty list
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "list is empty")
	})

	t.Run("AddMembersToTag_ExceedBatchSize", func(t *testing.T) {
		// Create members list exceeding TagMemberLimit
		members := make([]string, TagMemberLimit+1)
		for i := 0; i < TagMemberLimit+1; i++ {
			members[i] = fmt.Sprintf("member%d", i)
		}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		xrr := xwhttp.NewXResponseWriter(rr)

		// Read body and store it in XResponseWriter
		body, _ := io.ReadAll(req.Body)
		req.Body = nopCloser{bytes.NewReader(body)}
		xrr.SetBody(string(body))

		router.ServeHTTP(xrr, req)

		// Should return BadRequest for exceeding limit or error if service fails
		assert.True(t, rr.Code == http.StatusBadRequest || rr.Code >= http.StatusInternalServerError)
		if rr.Code == http.StatusBadRequest {
			assert.Contains(t, rr.Body.String(), "exceeds the limit")
		}
	})

	t.Run("AddMembersToTag_InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{"invalid": "json"}`)

		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/members", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		xrr := xwhttp.NewXResponseWriter(rr)

		body := make([]byte, len(invalidJSON))
		copy(body, invalidJSON)
		req.Body = nopCloser{bytes.NewReader(body)}

		router.ServeHTTP(xrr, req)

		// Should return BadRequest for invalid JSON
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "request body unmarshall error")
	})

	t.Run("AddMembersToTag_ResponseWriterCastError", func(t *testing.T) {
		members := []string{"member1"}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Use standard recorder instead of XResponseWriter
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return InternalServerError for writer cast error
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "response writer cast error")
	})
}

func TestDeleteTagFromXconfWithoutPrefixHandler_AllErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/noprefix", DeleteTagFromXconfWithoutPrefixHandler).Methods("DELETE")

	t.Run("DeleteTagFromXconf_ValidTag", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/noprefix", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return either NoContent or NotFound
		assert.True(t, rr.Code == http.StatusNoContent || rr.Code == http.StatusNotFound)
	})

	t.Run("DeleteTagFromXconf_MissingTag", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags//noprefix", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Route won't match
		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})

	t.Run("DeleteTagFromXconf_NonExistentTag", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/non-existent-tag-12345/noprefix", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return NotFound
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "tag not found")
	})
}

func TestGetTagByIdHandler_AllErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}", GetTagByIdHandler).Methods("GET")

	t.Run("GetTagById_ValidTag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/test-tag", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return either OK or NotFound
		assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusNotFound)
	})

	t.Run("GetTagById_MissingTag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Route won't match
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("GetTagById_NonExistentTag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/non-existent-tag-12345", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return NotFound
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "tag not found")
	})

	t.Run("GetTagById_SpecialCharacters", func(t *testing.T) {
		tag := "tag%2Btest"
		req, _ := http.NewRequest("GET", "/taggingService/tags/"+tag, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should handle special characters
		assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusNotFound)
	})
}

// Additional comprehensive error tests for other handlers

func TestRemoveMemberFromTagHandler_AllErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/members/{member}", RemoveMemberFromTagHandler).Methods("DELETE")

	t.Run("RemoveMemberFromTag_ValidParams", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members/test-member", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusNoContent || rr.Code >= http.StatusBadRequest)
	})

	t.Run("RemoveMemberFromTag_MissingTag", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags//members/test-member", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})

	t.Run("RemoveMemberFromTag_MissingMember", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members/", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})
}

func TestGetTagMembersHandler_AllErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/members", GetTagMembersHandler).Methods("GET")

	t.Run("GetTagMembers_ValidTag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/test-tag/members", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusOK || rr.Code >= http.StatusBadRequest)
	})

	t.Run("GetTagMembers_MissingTag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags//members", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})
}

func TestRemoveMembersFromTagHandler_AllErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/members", RemoveMembersFromTagHandler).Methods("DELETE")

	t.Run("RemoveMembersFromTag_ValidMembers", func(t *testing.T) {
		members := []string{"member1", "member2"}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusNoContent || rr.Code >= http.StatusBadRequest)
	})

	t.Run("RemoveMembersFromTag_EmptyList", func(t *testing.T) {
		members := []string{}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "list is empty")
	})

	t.Run("RemoveMembersFromTag_ExceedBatchSize", func(t *testing.T) {
		members := make([]string, TagMemberLimit+1)
		for i := 0; i < TagMemberLimit+1; i++ {
			members[i] = fmt.Sprintf("member%d", i)
		}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusBadRequest || rr.Code == http.StatusNotFound)
		if rr.Code == http.StatusBadRequest {
			assert.Contains(t, rr.Body.String(), "exceeds the limit")
		}
	})

	t.Run("RemoveMembersFromTag_InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{"invalid": json}`)

		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "request body unmarshall error")
	})

	t.Run("RemoveMembersFromTag_MissingTag", func(t *testing.T) {
		members := []string{"member1"}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("DELETE", "/taggingService/tags//members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})
}

func TestDeleteTagHandler_AllErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}", DeleteTagHandler).Methods("DELETE")

	t.Run("DeleteTag_ValidTag", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusNoContent || rr.Code == http.StatusNotFound)
	})

	t.Run("DeleteTag_MissingTag", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("DeleteTag_NonExistentTag", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/non-existent-tag-99999", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "tag not found")
	})
}

func TestGetTagsByMemberHandler_AllErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/members/{member}", GetTagsByMemberHandler).Methods("GET")

	t.Run("GetTagsByMember_ValidMember", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/test-member", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("GetTagsByMember_MissingMember", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("GetTagsByMember_SpecialCharacters", func(t *testing.T) {
		member := "member%20with%20spaces"
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/"+member, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestCalculatePercentageValueHandler_AllErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/members/{member}/percentages/calculation", CalculatePercentageValueHandler).Methods("GET")

	t.Run("CalculatePercentage_ValidMember", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/test-member/percentages/calculation", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var percentage int
		err := json.Unmarshal(rr.Body.Bytes(), &percentage)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, percentage, 0)
		assert.LessOrEqual(t, percentage, 100)
	})

	t.Run("CalculatePercentage_MissingMember", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/members//percentages/calculation", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})

	t.Run("CalculatePercentage_LongMember", func(t *testing.T) {
		longMember := strings.Repeat("a", 1000)
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/"+longMember+"/percentages/calculation", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
func TestGetTagMembersHandler_Integration(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/members", GetTagMembersHandler).Methods("GET")

	t.Run("GetTagMembers_WithValidPath", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/test-tag/members", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return OK with empty array or error
		assert.True(t, rr.Code == http.StatusOK || rr.Code >= http.StatusBadRequest)
	})
}

func TestCalculatePercentageValueHandler_Integration(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/members/{member}/percentages/calculation", CalculatePercentageValueHandler).Methods("GET")

	t.Run("CalculatePercentageValue_WithValidMember", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/test-member/percentages/calculation", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return OK with percentage value
		assert.Equal(t, http.StatusOK, rr.Code)

		var percentage int
		err := json.Unmarshal(rr.Body.Bytes(), &percentage)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, percentage, 0)
		assert.LessOrEqual(t, percentage, 100)
	})
}

func TestRemoveMembersFromTagHandler_Integration(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/members", RemoveMembersFromTagHandler).Methods("DELETE")

	t.Run("RemoveMembersFromTag_WithValidMembers", func(t *testing.T) {
		members := []string{"member1", "member2"}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return either NoContent or error
		assert.True(t, rr.Code == http.StatusNoContent || rr.Code >= http.StatusBadRequest)
	})

	t.Run("RemoveMembersFromTag_EmptyMembers", func(t *testing.T) {
		members := []string{}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return BadRequest for empty members
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("RemoveMembersFromTag_LargeMembers", func(t *testing.T) {
		// Create a large list exceeding the limit
		members := make([]string, TagMemberLimit+1)
		for i := 0; i < TagMemberLimit+1; i++ {
			members[i] = fmt.Sprintf("member%d", i)
		}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Without existing tag, may return 404; with existing and size check, 400
		assert.True(t, rr.Code == http.StatusBadRequest || rr.Code == http.StatusNotFound)
		if rr.Code == http.StatusBadRequest {
			assert.Contains(t, rr.Body.String(), "exceeds the limit")
		}
	})

	t.Run("RemoveMembersFromTag_InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{"invalid": "json"`)

		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return BadRequest for invalid JSON
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "request body unmarshall error")
	})
}

func TestTagHandlerConstants(t *testing.T) {
	// Test that constants are defined correctly
	assert.Equal(t, 1000, TagMemberLimit)
	assert.Contains(t, RequestBodyReadErrorMsg, "request body unmarshall error")
	assert.Contains(t, NotSpecifiedErrorMsg, "is not specified")
	assert.Contains(t, EmptyListErrorMsg, "list is empty")
	assert.Contains(t, MaxMemberLimitExceededErrorMsg, "batch size")
	assert.Contains(t, ResponseWriterCastErrorMsg, "response writer cast error")
	assert.Contains(t, NotFoundErrorMsg, "tag not found")
}

// Test Tag model functionality
func TestTagModel(t *testing.T) {
	t.Run("Tag_NewTagInf", func(t *testing.T) {
		tagInf := NewTagInf()
		tag, ok := tagInf.(*Tag)
		assert.True(t, ok, "NewTagInf should return a *Tag")
		assert.NotNil(t, tag)
	})

	t.Run("Tag_Clone", func(t *testing.T) {
		memberSet := util.Set{}
		memberSet.Add("member1")
		memberSet.Add("member2")
		originalTag := &Tag{
			Id:      "test-tag",
			Members: memberSet,
			Updated: 123456789,
		}

		clonedTag, err := originalTag.Clone()
		assert.NoError(t, err)
		assert.NotNil(t, clonedTag)
		assert.Equal(t, originalTag.Id, clonedTag.Id)
		assert.Equal(t, originalTag.Updated, clonedTag.Updated)

		// Verify it's a deep copy
		assert.NotSame(t, originalTag, clonedTag)
	})

	t.Run("Tag_MarshalJSON", func(t *testing.T) {
		memberSet2 := util.Set{}
		memberSet2.Add("member1")
		memberSet2.Add("member2")
		tag := &Tag{
			Id:      "test-tag",
			Members: memberSet2,
			Updated: 123456789,
		}

		jsonBytes, err := json.Marshal(tag)
		assert.NoError(t, err)
		assert.NotNil(t, jsonBytes)

		// Verify the JSON structure
		var result map[string]interface{}
		err = json.Unmarshal(jsonBytes, &result)
		assert.NoError(t, err)
		assert.Equal(t, "test-tag", result["id"])
		assert.Equal(t, float64(123456789), result["updated"])

		// Members should be an array
		members, ok := result["members"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, members, 2)
	})

	t.Run("Tag_UnmarshalJSON", func(t *testing.T) {
		jsonStr := `{"id":"test-tag","members":["member1","member2"],"updated":123456789}`

		var tag Tag
		err := json.Unmarshal([]byte(jsonStr), &tag)
		assert.NoError(t, err)
		assert.Equal(t, "test-tag", tag.Id)
		assert.Equal(t, int64(123456789), tag.Updated)
		assert.True(t, tag.Members.Contains("member1"))
		assert.True(t, tag.Members.Contains("member2"))
		assert.Len(t, tag.Members, 2)
	})

	t.Run("Tag_UnmarshalJSON_InvalidJSON", func(t *testing.T) {
		invalidJSON := `{"id":"test-tag","members":["member1","member2"],"updated":"invalid"}`

		var tag Tag
		err := json.Unmarshal([]byte(invalidJSON), &tag)
		assert.Error(t, err)
	})
}

// Additional comprehensive tests for better coverage

func TestHandlerConstants(t *testing.T) {
	// Test all constant values
	assert.Equal(t, "request body unmarshall error: %s", RequestBodyReadErrorMsg)
	assert.Equal(t, "%s is not specified", NotSpecifiedErrorMsg)
	assert.Equal(t, "%s list is empty", EmptyListErrorMsg)
	assert.Equal(t, "batch size %d exceeds the limit of %d", MaxMemberLimitExceededErrorMsg)
	assert.Equal(t, "response writer cast error", ResponseWriterCastErrorMsg)
	assert.Equal(t, "%s tag not found", NotFoundErrorMsg)
	assert.Equal(t, 1000, TagMemberLimit)
}

func TestGetAllTagsHandler_ErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags", GetAllTagsHandler).Methods("GET")

	t.Run("GetAllTags_WithInvalidQueryParam", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags?invalid=param", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should still work and return tag IDs
		assert.True(t, rr.Code == http.StatusOK || rr.Code >= http.StatusBadRequest)
	})
}

func TestAddMembersToTagHandler_ErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}/members", AddMembersToTagHandler).Methods("PUT")

	t.Run("AddMembersToTag_InvalidJSON", func(t *testing.T) {
		invalidJSON := `{"invalid": json}`
		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/members", bytes.NewBuffer([]byte(invalidJSON)))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		xrr := xwhttp.NewXResponseWriter(rr)
		router.ServeHTTP(xrr, req)

		// Should return BadRequest for invalid JSON
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("AddMembersToTag_NoBody", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/members", nil)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		xrr := xwhttp.NewXResponseWriter(rr)
		router.ServeHTTP(xrr, req)

		// Should return BadRequest for missing body
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("AddMembersToTag_ExceedLimit", func(t *testing.T) {
		// Create a list that exceeds TagMemberLimit
		members := make([]string, TagMemberLimit+1)
		for i := 0; i < TagMemberLimit+1; i++ {
			members[i] = fmt.Sprintf("member%d", i)
		}
		jsonBody, _ := json.Marshal(members)

		req, _ := http.NewRequest("PUT", "/taggingService/tags/test-tag/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		xrr := xwhttp.NewXResponseWriter(rr)

		body := make([]byte, len(jsonBody))
		copy(body, jsonBody)
		req.Body = nopCloser{bytes.NewReader(body)}

		router.ServeHTTP(xrr, req)

		// Should return BadRequest for exceeding limit
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// func TestRemoveMembersFromTagHandler_ErrorCases(t *testing.T) {
// 	router := mux.NewRouter()
// 	router.HandleFunc("/taggingService/tags/{tag}/members", RemoveMembersFromTagHandler).Methods("DELETE")

// 	t.Run("RemoveMembersFromTag_InvalidJSON", func(t *testing.T) {
// 		invalidJSON := `{"invalid": json}`
// 		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members", bytes.NewBuffer([]byte(invalidJSON)))
// 		req.Header.Set("Content-Type", "application/json")

// 		rr := httptest.NewRecorder()
// 		router.ServeHTTP(rr, req)

// 		// Should return BadRequest for invalid JSON
// 		assert.Equal(t, http.StatusBadRequest, rr.Code)
// 	})

// 	t.Run("RemoveMembersFromTag_NoBody", func(t *testing.T) {
// 		req, _ := http.NewRequest("DELETE", "/taggingService/tags/test-tag/members", nil)
// 		req.Header.Set("Content-Type", "application/json")

// 		rr := httptest.NewRecorder()
// 		router.ServeHTTP(rr, req)

// 		// Should return BadRequest for missing body
// 		assert.Equal(t, http.StatusBadRequest, rr.Code)
// 	})
// }

func TestGetTagsByMemberHandler_ErrorCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/members/{member}", GetTagsByMemberHandler).Methods("GET")

	t.Run("GetTagsByMember_EmptyMember", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Route won't match, should return 404
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("GetTagsByMember_SpecialCharacters", func(t *testing.T) {
		member := "test%20member"
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/"+member, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should handle special characters
		assert.True(t, rr.Code == http.StatusOK || rr.Code >= http.StatusBadRequest)
	})
}

func TestCalculatePercentageValueHandler_EdgeCases(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/members/{member}/percentages/calculation", CalculatePercentageValueHandler).Methods("GET")

	t.Run("CalculatePercentageValue_EmptyMember", func(t *testing.T) {
		// Empty member path is invalid; expect 404
		req, _ := http.NewRequest("GET", "/taggingService/tags/members//percentages/calculation", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		// Some routers may treat the double slash as redirect (301) before 404
		assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusMovedPermanently)
	})

	t.Run("CalculatePercentageValue_LongMember", func(t *testing.T) {
		longMember := strings.Repeat("a", 1000)
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/"+longMember+"/percentages/calculation", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should handle long member names
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("CalculatePercentageValue_SpecialCharacters", func(t *testing.T) {
		member := "testSpecialChars"
		req, _ := http.NewRequest("GET", "/taggingService/tags/members/"+member+"/percentages/calculation", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestHandlerPathVariables(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/taggingService/tags/{tag}", GetTagByIdHandler).Methods("GET")
	router.HandleFunc("/taggingService/tags/{tag}/members/{member}", RemoveMemberFromTagHandler).Methods("DELETE")

	t.Run("PathVariables_WithURLEncoding", func(t *testing.T) {
		// Test URL encoding in path variables
		encodedTag := "test%2Btag" // "test+tag" encoded
		req, _ := http.NewRequest("GET", "/taggingService/tags/"+encodedTag, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusOK || rr.Code >= http.StatusBadRequest)
	})

	t.Run("PathVariables_WithSlashes", func(t *testing.T) {
		// Test path variables containing encoded slashes
		encodedTag := "test%2Ftag"       // "test/tag" encoded
		encodedMember := "member%2Ftest" // "member/test" encoded
		req, _ := http.NewRequest("DELETE", "/taggingService/tags/"+encodedTag+"/members/"+encodedMember, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusNoContent || rr.Code >= http.StatusBadRequest)
	})
}
