// Copyright 2025 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	proto2 "github.com/rdkcentral/xconfadmin/taggingapi/proto/generated"
)

// Helper function to create a test HTTP client
func newTestHttpClientSync(server *httptest.Server) *HttpClient {
	return &HttpClient{
		Client:       server.Client(),
		retries:      3,
		retryInMsecs: 100,
	}
}

func TestGroupServiceSyncConnector_GetGroupServiceSyncHost(t *testing.T) {
	connector := &GroupServiceSyncConnector{
		BaseURL: "https://groupsync.example.com",
	}

	if connector.GetGroupServiceSyncHost() != "https://groupsync.example.com" {
		t.Errorf("expected 'https://groupsync.example.com', got %s", connector.GetGroupServiceSyncHost())
	}
}

func TestGroupServiceSyncConnector_SetGroupServiceSyncHost(t *testing.T) {
	connector := &GroupServiceSyncConnector{
		BaseURL: "https://old.example.com",
	}

	connector.SetGroupServiceSyncHost("https://new.example.com")

	if connector.BaseURL != "https://new.example.com" {
		t.Errorf("expected 'https://new.example.com', got %s", connector.BaseURL)
	}
}

func TestGroupServiceSyncConnector_AddMembersToTag(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// Verify path
		if r.URL.Path != "/v2/ft/test-group-id" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Content-Type") != ApplicationProtobufHeader {
			t.Errorf("expected Content-Type '%s', got %s", ApplicationProtobufHeader, r.Header.Get("Content-Type"))
		}

		if r.Header.Get(TtlHeader) != OneYearTtl {
			t.Errorf("expected Xttl '%s', got %s", OneYearTtl, r.Header.Get(TtlHeader))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := newTestHttpClientSync(server)

	connector := &GroupServiceSyncConnector{
		BaseURL: server.URL,
		Client:  httpClient,
	}

	members := &proto2.XdasHashes{
		Fields: map[string]string{
			"member1": "value1",
			"member2": "value2",
		},
	}

	err := connector.AddMembersToTag("test-group-id", members)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGroupServiceSyncConnector_AddMembersToTag_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	httpClient := newTestHttpClientSync(server)

	connector := &GroupServiceSyncConnector{
		BaseURL: server.URL,
		Client:  httpClient,
	}

	members := &proto2.XdasHashes{
		Fields: map[string]string{
			"member1": "value1",
		},
	}

	err := connector.AddMembersToTag("test-group-id", members)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGroupServiceSyncConnector_RemoveGroupMembers(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}

		// Verify path contains group ID and member
		expectedPath := "/v2/ft/test-group-id"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		// Verify query parameter
		if r.URL.Query().Get("field") != "test-member-id" {
			t.Errorf("expected field parameter 'test-member-id', got '%s'", r.URL.Query().Get("field"))
		}

		// Verify headers
		if r.Header.Get("Content-Type") != ApplicationProtobufHeader {
			t.Errorf("expected Content-Type '%s', got %s", ApplicationProtobufHeader, r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := newTestHttpClientSync(server)

	connector := &GroupServiceSyncConnector{
		BaseURL: server.URL,
		Client:  httpClient,
	}

	err := connector.RemoveGroupMembers("test-group-id", "test-member-id")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGroupServiceSyncConnector_RemoveGroupMembers_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	httpClient := newTestHttpClientSync(server)

	connector := &GroupServiceSyncConnector{
		BaseURL: server.URL,
		Client:  httpClient,
	}

	err := connector.RemoveGroupMembers("test-group-id", "test-member-id")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGroupServiceSyncConnector_DoRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	httpClient := newTestHttpClientSync(server)

	connector := &GroupServiceSyncConnector{
		BaseURL: server.URL,
		Client:  httpClient,
	}

	result, err := connector.DoRequest("GET", server.URL, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(result) != "success" {
		t.Errorf("expected 'success', got %s", string(result))
	}
}

func TestProtobufHeaders_Groupsync(t *testing.T) {
	headers := protobufHeaders()

	if headers == nil {
		t.Fatal("expected non-nil headers")
	}

	if len(headers) != 1 {
		t.Errorf("expected 2 headers, got %d", len(headers))
	}

	if headers[Accept] != ApplicationProtobufHeader {
		t.Errorf("expected Accept '%s', got %s", ApplicationProtobufHeader, headers[Accept])
	}

	if headers[ContentType] != ApplicationProtobufHeader {
		t.Errorf("expected ContentType '%s', got %s", ApplicationProtobufHeader, headers[ContentType])
	}
}
