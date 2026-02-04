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
func newTestHttpClientCanary(server *httptest.Server) *HttpClient {
	return &HttpClient{
		Client:       server.Client(),
		retries:      3,
		retryInMsecs: 100,
	}
}

func TestCanaryMgrConnector_GetCanaryMgrHost(t *testing.T) {
	connector := &CanaryMgrConnector{
		host: "https://canarymgr.example.com",
	}

	if connector.GetCanaryMgrHost() != "https://canarymgr.example.com" {
		t.Errorf("expected 'https://canarymgr.example.com', got %s", connector.GetCanaryMgrHost())
	}
}

func TestCanaryMgrConnector_SetCanaryMgrHost(t *testing.T) {
	connector := &CanaryMgrConnector{
		host: "https://old.example.com",
	}

	connector.SetCanaryMgrHost("https://new.example.com")

	if connector.host != "https://new.example.com" {
		t.Errorf("expected 'https://new.example.com', got %s", connector.host)
	}
}

func TestCanaryMgrConnector_CreateCanary(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// Verify path
		if r.URL.Path != "/api/v1/canarygroup" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify User-Agent header
		if r.Header.Get("User-Agent") == "" {
			t.Error("expected User-Agent header")
		}

		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var requestBody CanaryRequestBody
		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if requestBody.Name != "test-canary" {
			t.Errorf("expected Name 'test-canary', got %s", requestBody.Name)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	httpClient := newTestHttpClientCanary(server)

	connector := &CanaryMgrConnector{
		HttpClient: httpClient,
		host:       server.URL,
	}

	canaryRequest := &CanaryRequestBody{
		Name:                   "test-canary",
		DeviceType:             "stb",
		Size:                   100,
		DistributionPercentage: 10.0,
		Partner:                "test-partner",
		Model:                  "RNG150",
		FwAppliedRule:          "test-rule",
		TimeZones:              []string{"UTC", "EST"},
		StartPercentRange:      0.0,
		EndPercentRange:        10.0,
	}

	err := connector.CreateCanary(canaryRequest, false, log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCanaryMgrConnector_CreateCanary_DeepSleep(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// Verify path for deep sleep
		if r.URL.Path != "/api/v1/canarygroup/deepsleep" {
			t.Errorf("expected path '/api/v1/canarygroup/deepsleep', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	httpClient := newTestHttpClientCanary(server)

	connector := &CanaryMgrConnector{
		HttpClient: httpClient,
		host:       server.URL,
	}

	canaryRequest := &CanaryRequestBody{
		Name:       "test-canary-deepsleep",
		DeviceType: "video",
	}

	err := connector.CreateCanary(canaryRequest, true, log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCanaryMgrConnector_CreateCanary_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	httpClient := newTestHttpClientCanary(server)

	connector := &CanaryMgrConnector{
		HttpClient: httpClient,
		host:       server.URL,
	}

	canaryRequest := &CanaryRequestBody{
		Name: "test-canary",
	}

	err := connector.CreateCanary(canaryRequest, false, log.Fields{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCanaryMgrConnector_CreateWakeupPool(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// Verify path
		if r.URL.Path != "/api/v1/wakeuppool" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify force query parameter
		forceParam := r.URL.Query().Get("force")
		if forceParam != "true" {
			t.Errorf("expected force parameter 'true', got %s", forceParam)
		}

		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var requestBody WakeupPoolRequestBody
		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if len(requestBody.PercentFilters) != 1 {
			t.Errorf("expected 1 PercentFilter, got %d", len(requestBody.PercentFilters))
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	httpClient := newTestHttpClientCanary(server)

	connector := &CanaryMgrConnector{
		HttpClient: httpClient,
		host:       server.URL,
	}

	wakeupPoolRequest := &WakeupPoolRequestBody{
		PercentFilters: []WakeupPoolPercentFilter{
			{
				Name:       "test-filter",
				DeviceType: "video",
				Size:       50,
				Partner:    "test-partner",
				Model:      "MODEL1",
				TimeZones:  []string{"UTC"},
				Distributions: []WakeupPoolDistribution{
					{
						ConfigId:          "config-1",
						StartPercentRange: 0.0,
						EndPercentRange:   50.0,
					},
				},
			},
		},
	}

	err := connector.CreateWakeupPool(wakeupPoolRequest, true, log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCanaryMgrConnector_CreateWakeupPool_ForcefalseParameter(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify force query parameter is false
		forceParam := r.URL.Query().Get("force")
		if forceParam != "false" {
			t.Errorf("expected force parameter 'false', got %s", forceParam)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	httpClient := newTestHttpClientCanary(server)

	connector := &CanaryMgrConnector{
		HttpClient: httpClient,
		host:       server.URL,
	}

	wakeupPoolRequest := &WakeupPoolRequestBody{
		PercentFilters: []WakeupPoolPercentFilter{},
	}

	err := connector.CreateWakeupPool(wakeupPoolRequest, false, log.Fields{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCanaryMgrConnector_CreateWakeupPool_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	httpClient := newTestHttpClientCanary(server)

	connector := &CanaryMgrConnector{
		HttpClient: httpClient,
		host:       server.URL,
	}

	wakeupPoolRequest := &WakeupPoolRequestBody{
		PercentFilters: []WakeupPoolPercentFilter{},
	}

	err := connector.CreateWakeupPool(wakeupPoolRequest, false, log.Fields{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
