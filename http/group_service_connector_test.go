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
	"google.golang.org/protobuf/proto"
)

// Helper function to create a test HTTP client
func newTestHttpClient(server *httptest.Server) *HttpClient {
	return &HttpClient{
		Client:       server.Client(),
		retries:      3,
		retryInMsecs: 100,
	}
}

func TestGroupServiceConnector_GetGroupServiceHost(t *testing.T) {
	connector := &GroupServiceConnector{
		BaseURL: "https://group.example.com",
	}

	if connector.GetGroupServiceHost() != "https://group.example.com" {
		t.Errorf("expected 'https://group.example.com', got %s", connector.GetGroupServiceHost())
	}
}

func TestGroupServiceConnector_SetGroupServiceHost(t *testing.T) {
	connector := &GroupServiceConnector{
		BaseURL: "https://old.example.com",
	}

	connector.SetGroupServiceHost("https://new.example.com")

	if connector.BaseURL != "https://new.example.com" {
		t.Errorf("expected 'https://new.example.com', got %s", connector.BaseURL)
	}
}

func TestGroupServiceConnector_GetGroupsMemberBelongsTo(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		// Check that the URL contains the member ID
		if r.URL.Path != "/v2/ft/test-member-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Create a mock response
		groups := &proto2.XdasHashes{
			Fields: map[string]string{
				"group1": "value1",
				"group2": "value2",
				"group3": "value3",
			},
		}

		data, _ := proto.Marshal(groups)
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.Write(data)
	}))
	defer server.Close()

	httpClient := newTestHttpClient(server)

	connector := &GroupServiceConnector{
		BaseURL: server.URL,
		Client:  httpClient,
	}

	result, err := connector.GetGroupsMemberBelongsTo("test-member-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Fields) != 3 {
		t.Errorf("expected 3 groups, got %d", len(result.Fields))
	}

	if result.Fields["group1"] != "value1" {
		t.Errorf("expected group1 value 'value1', got %s", result.Fields["group1"])
	}
}

func TestGroupServiceConnector_GetAllGroups(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		if r.URL.Path != "/v2/ft" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Create a mock response
		groups := &proto2.XdasHashes{
			Fields: map[string]string{
				"all-group1": "val1",
				"all-group2": "val2",
			},
		}

		data, _ := proto.Marshal(groups)
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.Write(data)
	}))
	defer server.Close()

	httpClient := newTestHttpClient(server)

	connector := &GroupServiceConnector{
		BaseURL: server.URL,
		Client:  httpClient,
	}

	result, err := connector.GetAllGroups()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Fields) != 2 {
		t.Errorf("expected 2 groups, got %d", len(result.Fields))
	}
}

func TestGroupServiceConnector_GetGroupsMemberBelongsTo_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	httpClient := newTestHttpClient(server)

	connector := &GroupServiceConnector{
		BaseURL: server.URL,
		Client:  httpClient,
	}

	result, err := connector.GetGroupsMemberBelongsTo("test-member")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestGroupServiceConnector_GetAllGroups_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	httpClient := newTestHttpClient(server)

	connector := &GroupServiceConnector{
		BaseURL: server.URL,
		Client:  httpClient,
	}

	result, err := connector.GetAllGroups()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestUnmarshalXdasHashes_ValidData(t *testing.T) {
	groups := &proto2.XdasHashes{
		Fields: map[string]string{
			"hash1": "value1",
			"hash2": "value2",
		},
	}

	data, err := proto.Marshal(groups)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	result, err := unmarshalXdasHashes(data)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Fields) != 2 {
		t.Errorf("expected 2 hashes, got %d", len(result.Fields))
	}
}

func TestUnmarshalXdasHashes_InvalidData(t *testing.T) {
	invalidData := []byte("not valid protobuf data")

	result, err := unmarshalXdasHashes(invalidData)

	if err == nil {
		t.Fatal("expected error for invalid data")
	}

	if result != nil {
		t.Error("expected nil result for invalid data")
	}
}

func TestGroupServiceConnector_DoRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Content-Type") != "application/x-protobuf" {
			t.Errorf("expected Content-Type header 'application/x-protobuf', got %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	httpClient := newTestHttpClient(server)

	connector := &GroupServiceConnector{
		BaseURL: server.URL,
		Client:  httpClient,
	}

	headers := protobufHeaders()
	result, err := connector.DoRequest("GET", server.URL, headers, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(result) != "success" {
		t.Errorf("expected 'success', got %s", string(result))
	}
}

func TestProtobufHeaders(t *testing.T) {
	headers := protobufHeaders()

	if headers == nil {
		t.Fatal("expected non-nil headers")
	}

	if headers[Accept] != ApplicationProtobufHeader {
		t.Errorf("expected Accept header '%s', got %s", ApplicationProtobufHeader, headers[Accept])
	}

	if headers[ContentType] != ApplicationProtobufHeader {
		t.Errorf("expected ContentType header '%s', got %s", ApplicationProtobufHeader, headers[ContentType])
	}
}
