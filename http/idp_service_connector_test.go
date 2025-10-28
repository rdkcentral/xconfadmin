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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// Helper function to create a test HTTP client
func newTestHttpClientIdp(server *httptest.Server) *HttpClient {
	return &HttpClient{
		Client:       server.Client(),
		retries:      3,
		retryInMsecs: 100,
	}
}

type MockIdpService struct {
	host   string
	config *IdpServiceConfig
}

func (m *MockIdpService) IdpServiceHost() string {
	return m.host
}

func (m *MockIdpService) SetIdpServiceHost(host string) {
	m.host = host
}

func (m *MockIdpService) GetFullLoginUrl(continueUrl string) string {
	return fmt.Sprintf(fullLoginUrl, m.host, continueUrl, m.config.ClientId)
}

func (m *MockIdpService) GetJsonWebKeyResponse(url string) *JsonWebKeyResponse {
	return &JsonWebKeyResponse{
		Keys: []JsonWebKey{
			{
				KeyType: "RSA",
				E:       "AQAB",
				Use:     "sig",
				Kid:     "test-kid",
				Alg:     "RS256",
				N:       "test-n-value",
			},
		},
	}
}

func (m *MockIdpService) GetFullLogoutUrl(continueUrl string) string {
	return fmt.Sprintf(fullLogoutUrl, m.host, continueUrl, m.config.ClientId)
}

func (m *MockIdpService) GetToken(code string) string {
	return `{"access_token":"test-token","token_type":"Bearer"}`
}

func (m *MockIdpService) Logout(url string) error {
	return nil
}

func (m *MockIdpService) GetIdpServiceConfig() *IdpServiceConfig {
	return m.config
}

func TestDefaultIdpService_IdpServiceHost(t *testing.T) {
	service := &DefaultIdpService{
		host: "https://idp.example.com",
	}

	if service.IdpServiceHost() != "https://idp.example.com" {
		t.Errorf("expected 'https://idp.example.com', got %s", service.IdpServiceHost())
	}
}

func TestDefaultIdpService_SetIdpServiceHost(t *testing.T) {
	service := &DefaultIdpService{
		host: "https://old.example.com",
	}

	service.SetIdpServiceHost("https://new.example.com")

	if service.host != "https://new.example.com" {
		t.Errorf("expected 'https://new.example.com', got %s", service.host)
	}
}

func TestDefaultIdpService_GetFullLoginUrl(t *testing.T) {
	config := &IdpServiceConfig{
		ClientId: "test-client-id",
	}
	service := &DefaultIdpService{
		host:             "https://idp.example.com",
		IdpServiceConfig: config,
	}

	continueUrl := "https://app.example.com/callback"
	loginUrl := service.GetFullLoginUrl(continueUrl)

	expected := fmt.Sprintf(fullLoginUrl, "https://idp.example.com", continueUrl, "test-client-id")
	if loginUrl != expected {
		t.Errorf("expected '%s', got '%s'", expected, loginUrl)
	}
}

func TestDefaultIdpService_GetFullLogoutUrl(t *testing.T) {
	config := &IdpServiceConfig{
		ClientId: "test-client-id",
	}
	service := &DefaultIdpService{
		host:             "https://idp.example.com",
		IdpServiceConfig: config,
	}

	continueUrl := "https://app.example.com"
	logoutUrl := service.GetFullLogoutUrl(continueUrl)

	expected := fmt.Sprintf(fullLogoutUrl, "https://idp.example.com", continueUrl, "test-client-id")
	if logoutUrl != expected {
		t.Errorf("expected '%s', got '%s'", expected, logoutUrl)
	}
}

func TestDefaultIdpService_GetIdpServiceConfig(t *testing.T) {
	config := &IdpServiceConfig{
		ClientId:        "test-client",
		ClientSecret:    "test-secret",
		KidMap:          sync.Map{},
		AuthHeaderValue: "Basic dGVzdA==",
	}
	service := &DefaultIdpService{
		IdpServiceConfig: config,
	}

	result := service.GetIdpServiceConfig()

	if result == nil {
		t.Fatal("expected non-nil config")
	}

	if result.ClientId != "test-client" {
		t.Errorf("expected ClientId 'test-client', got %s", result.ClientId)
	}

	if result.ClientSecret != "test-secret" {
		t.Errorf("expected ClientSecret 'test-secret', got %s", result.ClientSecret)
	}
}

func TestDefaultIdpService_GetToken(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("expected Authorization header")
		}

		// Return a mock token response
		response := map[string]string{
			"access_token": "mock-access-token",
			"token_type":   "Bearer",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &IdpServiceConfig{
		ClientId:        "test-client",
		ClientSecret:    "test-secret",
		AuthHeaderValue: "Basic " + base64.StdEncoding.EncodeToString([]byte("test-client:test-secret")),
	}

	httpClient := newTestHttpClientIdp(server)

	service := &DefaultIdpService{
		host:             server.URL,
		HttpClient:       httpClient,
		IdpServiceConfig: config,
	}

	token := service.GetToken("test-code")

	if token == "" {
		t.Error("expected non-empty token")
	}

	// Verify the token contains expected structure
	var tokenData map[string]interface{}
	err := json.Unmarshal([]byte(token), &tokenData)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if tokenData["access_token"] != "mock-access-token" {
		t.Errorf("expected access_token 'mock-access-token', got %v", tokenData["access_token"])
	}
}

func TestDefaultIdpService_GetJsonWebKeyResponse(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		response := JsonWebKeyResponse{
			Keys: []JsonWebKey{
				{
					KeyType: "RSA",
					E:       "AQAB",
					Use:     "sig",
					Kid:     "key-1",
					Alg:     "RS256",
					N:       "modulus-value",
				},
				{
					KeyType: "RSA",
					E:       "AQAB",
					Use:     "sig",
					Kid:     "key-2",
					Alg:     "RS256",
					N:       "another-modulus",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	httpClient := newTestHttpClientIdp(server)

	service := &DefaultIdpService{
		HttpClient: httpClient,
	}

	result := service.GetJsonWebKeyResponse(server.URL)

	if result == nil {
		t.Fatal("expected non-nil JsonWebKeyResponse")
	}

	if len(result.Keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(result.Keys))
	}

	if result.Keys[0].Kid != "key-1" {
		t.Errorf("expected first key Kid 'key-1', got %s", result.Keys[0].Kid)
	}

	if result.Keys[1].Kid != "key-2" {
		t.Errorf("expected second key Kid 'key-2', got %s", result.Keys[1].Kid)
	}
}

func TestDefaultIdpService_GetJsonWebKeyResponse_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	httpClient := newTestHttpClientIdp(server)

	service := &DefaultIdpService{
		HttpClient: httpClient,
	}

	result := service.GetJsonWebKeyResponse(server.URL)

	if result != nil {
		t.Error("expected nil result for invalid JSON")
	}
}

func TestDefaultIdpService_Logout(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := newTestHttpClientIdp(server)

	service := &DefaultIdpService{
		HttpClient: httpClient,
	}

	err := service.Logout(server.URL + "/logout")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMockIdpService_AllMethods(t *testing.T) {
	config := &IdpServiceConfig{
		ClientId:     "mock-client-id",
		ClientSecret: "mock-secret",
	}

	mock := &MockIdpService{
		host:   "https://mock-idp.example.com",
		config: config,
	}

	// Test IdpServiceHost
	if mock.IdpServiceHost() != "https://mock-idp.example.com" {
		t.Error("IdpServiceHost failed")
	}

	// Test SetIdpServiceHost
	mock.SetIdpServiceHost("https://new-mock-idp.example.com")
	if mock.host != "https://new-mock-idp.example.com" {
		t.Error("SetIdpServiceHost failed")
	}

	// Test GetFullLoginUrl
	loginUrl := mock.GetFullLoginUrl("https://continue.url")
	if loginUrl == "" {
		t.Error("GetFullLoginUrl returned empty string")
	}

	// Test GetFullLogoutUrl
	logoutUrl := mock.GetFullLogoutUrl("https://continue.url")
	if logoutUrl == "" {
		t.Error("GetFullLogoutUrl returned empty string")
	}

	// Test GetToken
	token := mock.GetToken("test-code")
	if token == "" {
		t.Error("GetToken returned empty string")
	}

	// Test GetJsonWebKeyResponse
	jwkResponse := mock.GetJsonWebKeyResponse("https://jwks.url")
	if jwkResponse == nil || len(jwkResponse.Keys) == 0 {
		t.Error("GetJsonWebKeyResponse failed")
	}

	// Test Logout
	err := mock.Logout("https://logout.url")
	if err != nil {
		t.Errorf("Logout failed: %v", err)
	}

	// Test GetIdpServiceConfig
	cfg := mock.GetIdpServiceConfig()
	if cfg == nil || cfg.ClientId != "mock-client-id" {
		t.Error("GetIdpServiceConfig failed")
	}
}
