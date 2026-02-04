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

	log "github.com/sirupsen/logrus"
)

// Helper function to create a test HTTP client
func newTestHttpClientXconf(server *httptest.Server) *HttpClient {
	return &HttpClient{
		Client:       server.Client(),
		retries:      3,
		retryInMsecs: 100,
	}
}

func TestXconfConnector_Host(t *testing.T) {
	connector := &XconfConnector{
		host: "https://xconf.example.com",
	}

	if connector.Host() != "https://xconf.example.com" {
		t.Errorf("expected 'https://xconf.example.com', got %s", connector.Host())
	}
}

func TestXconfConnector_SetXconfHost(t *testing.T) {
	connector := &XconfConnector{
		host: "https://old.example.com",
	}

	connector.SetXconfHost("https://new.example.com")

	if connector.host != "https://new.example.com" {
		t.Errorf("expected 'https://new.example.com', got %s", connector.host)
	}
}

func TestXconfConnector_ServiceName(t *testing.T) {
	connector := &XconfConnector{
		serviceName: "test-service",
	}

	if connector.ServiceName() != "test-service" {
		t.Errorf("expected 'test-service', got %s", connector.ServiceName())
	}
}

func TestXconfConnector_GetProfiles(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		// Check that the URL contains the expected path
		if r.URL.Path != "/loguploader/getTelemetryProfiles" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Check query parameters
		if r.URL.Query().Get("model") != "RNG150" {
			t.Errorf("expected model=RNG150, got %s", r.URL.Query().Get("model"))
		}

		// Return mock response
		response := `[{"id":"test-profile-1","name":"Profile1"},{"id":"test-profile-2","name":"Profile2"}]`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	httpClient := newTestHttpClientXconf(server)

	connector := &XconfConnector{
		HttpClient:  httpClient,
		host:        server.URL,
		serviceName: "xconf-test",
	}

	result, err := connector.GetProfiles("model=RNG150", log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result) == 0 {
		t.Error("expected non-empty result")
	}

	// Verify response contains expected data
	expectedContent := "test-profile-1"
	if !contains(string(result), expectedContent) {
		t.Errorf("expected result to contain '%s'", expectedContent)
	}
}

func TestXconfConnector_GetProfiles_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	httpClient := newTestHttpClientXconf(server)

	connector := &XconfConnector{
		HttpClient:  httpClient,
		host:        server.URL,
		serviceName: "xconf-test",
	}

	_, err := connector.GetProfiles("model=RNG150", log.Fields{})

	if err == nil {
		t.Fatal("expected error but got none")
	}
}

func TestXconfConnector_GetProfiles_WithDifferentQueryParams(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify different query parameters
		if r.URL.RawQuery == "" {
			t.Error("expected query parameters")
		}

		response := `[]`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	httpClient := newTestHttpClientXconf(server)

	connector := &XconfConnector{
		HttpClient:  httpClient,
		host:        server.URL,
		serviceName: "xconf-test",
	}

	// Test with different URL suffixes
	testCases := []string{
		"model=RNG150",
		"model=RNG150&partner=comcast",
		"firmwareVersion=1.2.3",
	}

	for _, tc := range testCases {
		_, err := connector.GetProfiles(tc, log.Fields{})
		if err != nil {
			t.Errorf("unexpected error for query '%s': %v", tc, err)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
