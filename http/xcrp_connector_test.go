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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
)

// Helper function to create a test HTTP client
func newTestHttpClientXcrp(server *httptest.Server) *HttpClient {
	return &HttpClient{
		Client:       server.Client(),
		retries:      3,
		retryInMsecs: 100,
	}
}

func TestXcrpConnector_XcrpHosts(t *testing.T) {
	connector := &XcrpConnector{
		hosts: []string{"https://xcrp1.example.com", "https://xcrp2.example.com"},
	}

	hosts := connector.XcrpHosts()
	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(hosts))
	}

	if hosts[0] != "https://xcrp1.example.com" {
		t.Errorf("expected 'https://xcrp1.example.com', got %s", hosts[0])
	}
}

func TestXcrpConnector_SetXcrpHosts(t *testing.T) {
	connector := &XcrpConnector{
		hosts: []string{"https://old.example.com"},
	}

	newHosts := []string{"https://new1.example.com", "https://new2.example.com"}
	connector.SetXcrpHosts(newHosts)

	if len(connector.hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(connector.hosts))
	}

	if connector.hosts[0] != "https://new1.example.com" {
		t.Errorf("expected 'https://new1.example.com', got %s", connector.hosts[0])
	}
}

func TestXcrpConnector_PostRecook_WithModelsAndPartners(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// Check that the URL contains both models and partners
		if !hasSubstringXcrp(r.URL.Path, "/api/v1/precook/rfc") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Check query parameters
		partners := r.URL.Query().Get("partners")
		models := r.URL.Query().Get("models")

		if partners != "comcast,cox" {
			t.Errorf("expected partners=comcast,cox, got %s", partners)
		}

		if models != "RNG150,XB6" {
			t.Errorf("expected models=RNG150,XB6, got %s", models)
		}

		// Check request body
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			t.Error("expected non-empty request body")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := newTestHttpClientXcrp(server)

	connector := &XcrpConnector{
		HttpClient: httpClient,
		hosts:      []string{server.URL},
	}

	models := []string{"RNG150", "XB6"}
	partners := []string{"comcast", "cox"}
	requestBody := []byte(`{"test":"data"}`)

	err := connector.PostRecook(models, partners, requestBody, log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestXcrpConnector_PostRecook_WithModelsOnly(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		models := r.URL.Query().Get("models")
		partners := r.URL.Query().Get("partners")

		if models != "RNG150" {
			t.Errorf("expected models=RNG150, got %s", models)
		}

		if partners != "" {
			t.Errorf("expected no partners parameter, got %s", partners)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := newTestHttpClientXcrp(server)

	connector := &XcrpConnector{
		HttpClient: httpClient,
		hosts:      []string{server.URL},
	}

	models := []string{"RNG150"}
	partners := []string{}
	requestBody := []byte(`{"test":"data"}`)

	err := connector.PostRecook(models, partners, requestBody, log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestXcrpConnector_PostRecook_WithPartnersOnly(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		partners := r.URL.Query().Get("partners")
		models := r.URL.Query().Get("models")

		if partners != "comcast" {
			t.Errorf("expected partners=comcast, got %s", partners)
		}

		if models != "" {
			t.Errorf("expected no models parameter, got %s", models)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := newTestHttpClientXcrp(server)

	connector := &XcrpConnector{
		HttpClient: httpClient,
		hosts:      []string{server.URL},
	}

	models := []string{}
	partners := []string{"comcast"}
	requestBody := []byte(`{"test":"data"}`)

	err := connector.PostRecook(models, partners, requestBody, log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestXcrpConnector_PostRecook_NoParams(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that URL has no query parameters
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query parameters, got %s", r.URL.RawQuery)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := newTestHttpClientXcrp(server)

	connector := &XcrpConnector{
		HttpClient: httpClient,
		hosts:      []string{server.URL},
	}

	models := []string{}
	partners := []string{}
	requestBody := []byte(`{"test":"data"}`)

	err := connector.PostRecook(models, partners, requestBody, log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestXcrpConnector_PostRecook_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	httpClient := newTestHttpClientXcrp(server)

	connector := &XcrpConnector{
		HttpClient: httpClient,
		hosts:      []string{server.URL},
	}

	models := []string{"RNG150"}
	partners := []string{"comcast"}
	requestBody := []byte(`{"test":"data"}`)

	err := connector.PostRecook(models, partners, requestBody, log.Fields{})

	if err == nil {
		t.Fatal("expected error but got none")
	}
}

func TestXcrpConnector_PostRecook_MultipleHosts(t *testing.T) {
	// Track which servers were called
	callCount := 0

	// Create two test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	httpClient := newTestHttpClientXcrp(server1) // Use server1's client

	connector := &XcrpConnector{
		HttpClient: httpClient,
		hosts:      []string{server1.URL, server2.URL},
	}

	models := []string{"RNG150"}
	partners := []string{"comcast"}
	requestBody := []byte(`{"test":"data"}`)

	err := connector.PostRecook(models, partners, requestBody, log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both hosts should have been called
	if callCount != 2 {
		t.Errorf("expected both hosts to be called, got %d calls", callCount)
	}
}

func TestXcrpConnector_GetRecookingStatusFromCanaryMgr_Completed(t *testing.T) {
	// Create a test server that returns completed status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		// Check path contains module name
		if !hasSubstringXcrp(r.URL.Path, "/api/v1/precook/rfc/status") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"status":  200,
			"message": "Success",
			"data": map[string]string{
				"status":      "completed",
				"updatedTime": "2024-01-01T00:00:00Z",
			},
		}

		data, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer server.Close()

	httpClient := newTestHttpClientXcrp(server)

	connector := &XcrpConnector{
		HttpClient: httpClient,
		hosts:      []string{server.URL},
	}

	completed, err := connector.GetRecookingStatusFromCanaryMgr("rfc", log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !completed {
		t.Error("expected status to be completed")
	}
}

func TestXcrpConnector_GetRecookingStatusFromCanaryMgr_Pending(t *testing.T) {
	// Create a test server that returns pending status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":  200,
			"message": "Success",
			"data": map[string]string{
				"status":      "pending",
				"updatedTime": "2024-01-01T00:00:00Z",
			},
		}

		data, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer server.Close()

	httpClient := newTestHttpClientXcrp(server)

	connector := &XcrpConnector{
		HttpClient: httpClient,
		hosts:      []string{server.URL},
	}

	completed, err := connector.GetRecookingStatusFromCanaryMgr("rfc", log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if completed {
		t.Error("expected status to be not completed")
	}
}

func TestXcrpConnector_GetRecookingStatusFromCanaryMgr_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	httpClient := newTestHttpClientXcrp(server)

	connector := &XcrpConnector{
		HttpClient: httpClient,
		hosts:      []string{server.URL},
	}

	_, err := connector.GetRecookingStatusFromCanaryMgr("rfc", log.Fields{})

	if err == nil {
		t.Fatal("expected error but got none")
	}
}

func TestXcrpConnector_GetRecookingStatusFromCanaryMgr_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	httpClient := newTestHttpClientXcrp(server)

	connector := &XcrpConnector{
		HttpClient: httpClient,
		hosts:      []string{server.URL},
	}

	_, err := connector.GetRecookingStatusFromCanaryMgr("rfc", log.Fields{})

	if err == nil {
		t.Fatal("expected error for invalid JSON but got none")
	}
}

// Helper function to check if a string contains a substring
func hasSubstringXcrp(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
