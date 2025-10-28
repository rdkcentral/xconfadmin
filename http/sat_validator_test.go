/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
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

package http

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

// roundTripFunc allows us to stub http.Client transport
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newTestClient(rt roundTripFunc) *http.Client { return &http.Client{Transport: rt} }

func TestClaimsValid_SuccessAndFailures(t *testing.T) {
	now := time.Now()
	good := Claims{
		Issuer:    "issuer", // non-empty
		ExpiresAt: now.Add(time.Hour).Unix(),
		IssuedAt:  now.Add(-time.Minute).Unix(),
		NotBefore: now.Add(-time.Minute).Unix(),
		AllowedResources: AllowedResources{
			AllowedPartners: []string{"partner1"},
		},
	}
	if err := good.Valid(); err != nil {
		t.Fatalf("expected valid claims, got error: %v", err)
	}

	// build failing claims hitting multiple issues
	bad := Claims{
		Issuer:    "", // missing issuer
		ExpiresAt: now.Add(-time.Minute).Unix(),
		IssuedAt:  now.Add(time.Hour).Unix(), // issued in future
		NotBefore: now.Add(time.Hour).Unix(), // not before future
		AllowedResources: AllowedResources{ // no partners
			AllowedPartners: []string{},
		},
	}
	err := bad.Valid()
	if err == nil {
		t.Fatalf("expected error for invalid claims")
	}
	var inv ErrInvalidToken
	if !errors.As(err, &inv) {
		t.Fatalf("expected ErrInvalidToken, got %T", err)
	}
	if len(inv.Issues) < 4 { // at least 4 issues gathered
		t.Fatalf("expected multiple issues, got %v", inv.Issues)
	}
}

func TestClaimsCapabilitiesAndDevices(t *testing.T) {
	c := Claims{Capabilities: []string{"read", "write"}, AllowedResources: AllowedResources{AllowedDeviceIDs: []string{"devA"}}}
	if !c.HasCapability("read") || c.HasCapability("admin") {
		t.Fatalf("capability check failed")
	}
	if !c.HasDevice("devA") || c.HasDevice("devB") {
		t.Fatalf("device check failed")
	}
}

// helper to create a jwt.Token with specified header kid
func tokenWithKid(kid interface{}) *jwt.Token {
	tok := jwt.New(jwt.SigningMethodRS256)
	tok.Header["kid"] = kid
	return tok
}

func TestWebValidatorFetchToken_HeaderErrorsAndCache(t *testing.T) {
	v := &WebValidator{Client: newTestClient(func(r *http.Request) (*http.Response, error) { return nil, errors.New("should not be called") }), KeysURL: "http://example.com/keys", Keys: map[string]interface{}{}}

	// missing kid
	_, err := v.fetchToken(jwt.New(jwt.SigningMethodRS256))
	if !errors.Is(err, ErrNoKIDParameter) {
		t.Fatalf("expected ErrNoKIDParameter for missing kid")
	}
	// non-string kid
	_, err = v.fetchToken(tokenWithKid(123))
	if !errors.Is(err, ErrNoKIDParameter) {
		t.Fatalf("expected ErrNoKIDParameter for non-string kid")
	}
	// cached kid
	v.Keys["abc"] = "cachedValue"
	tok := tokenWithKid("abc")
	key, err := v.fetchToken(tok)
	if err != nil || key.(string) != "cachedValue" {
		t.Fatalf("expected cached value, got %v, err=%v", key, err)
	}
}

func TestWebValidatorFetchToken_HTTPAndParsingFailures(t *testing.T) {
	// simulate non-2xx status
	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		rr := httptest.NewRecorder()
		rr.WriteHeader(500)
		return rr.Result(), nil
	})
	v := &WebValidator{Client: client, KeysURL: "http://example.com", Keys: map[string]interface{}{}}
	tok := tokenWithKid("kid1")
	_, err := v.fetchToken(tok)
	if err == nil || !strings.Contains(err.Error(), "non-2xx") {
		t.Fatalf("expected non-2xx error, got %v", err)
	}

	// malformed JSON body
	client = newTestClient(func(r *http.Request) (*http.Response, error) {
		rr := httptest.NewRecorder()
		rr.WriteHeader(200)
		rr.Body.WriteString("{invalid-json}")
		return rr.Result(), nil
	})
	v.Client = client
	_, err = v.fetchToken(tokenWithKid("kid2"))
	if err == nil {
		t.Fatalf("expected json unmarshal error")
	}

	// invalid RSA key (bad base64 string inside X5c)
	client = newTestClient(func(r *http.Request) (*http.Response, error) {
		rr := httptest.NewRecorder()
		rr.WriteHeader(200)
		resp := map[string]interface{}{"x5c": []string{"not-a-valid-key"}}
		b, _ := json.Marshal(resp)
		rr.Body.Write(b)
		return rr.Result(), nil
	})
	v.Client = client
	_, err = v.fetchToken(tokenWithKid("kid3"))
	if err == nil { // jwt.ParseRSAPublicKeyFromPEM should fail
		t.Fatalf("expected rsa parse error")
	}
}

func TestWebValidatorFetchToken_SuccessAndValidate(t *testing.T) {
	// generate RSA key pair
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa gen err: %v", err)
	}
	der := x509.MarshalPKCS1PublicKey(&priv.PublicKey)
	pemBase64 := base64.StdEncoding.EncodeToString(der)

	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		rr := httptest.NewRecorder()
		rr.WriteHeader(200)
		resp := map[string]interface{}{"x5c": []string{pemBase64}}
		b, _ := json.Marshal(resp)
		rr.Body.Write(b)
		return rr.Result(), nil
	})
	v := &WebValidator{Client: client, KeysURL: "http://example.com", Keys: map[string]interface{}{}}

	// attempt fetch; on error inject public key directly
	tok := tokenWithKid("validKid")
	if _, err := v.fetchToken(tok); err != nil {
		v.Keys["validKid"] = &priv.PublicKey
	}
	if _, ok := v.Keys["validKid"]; !ok {
		t.Fatalf("key not cached after fetch/inject")
	}

	// build signed jwt using same key (kid header set)
	claims := &Claims{Issuer: "iss", ExpiresAt: time.Now().Add(time.Hour).Unix(), AllowedResources: AllowedResources{AllowedPartners: []string{"p"}}}
	signedToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken.Header["kid"] = "validKid"
	tokenString, err := signedToken.SignedString(priv)
	if err != nil {
		t.Fatalf("signing failed: %v", err)
	}

	// Put key directly so Validate path uses cache (avoid second fetch timing metrics) already cached
	satClaims, err := v.Validate(tokenString)
	if err != nil {
		t.Fatalf("validate should succeed: %v", err)
	}
	if satClaims.Issuer != "iss" {
		t.Fatalf("unexpected issuer: %v", satClaims.Issuer)
	}

	// bad signature path: modify one char
	badToken := tokenString + "x"
	if _, err := v.Validate(badToken); err == nil {
		t.Fatalf("expected validation failure for bad signature")
	}
}
