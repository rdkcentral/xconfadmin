/**
 * Copyright 2023 Comcast Cable Communications Management, LLC
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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	keyRequestSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "sat",
			Subsystem: "server",
			Name:      "publickey_request_seconds",
			Help:      "elapsed time to retrieve and decode the public key",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"status"},
	)
)

// Claims represents all claims as defined in the SAT standard
type Claims struct {
	ID               string           `json:"jti,omitempty"`
	Issuer           string           `json:"iss,omitempty"`
	ExpiresAt        int64            `json:"exp,omitempty"`
	IssuedAt         int64            `json:"iat,omitempty"`
	NotBefore        int64            `json:"nbf,omitempty"`
	Version          string           `json:"version,omitempty"`
	Subject          string           `json:"sub,omitempty"`
	Audience         []string         `json:"aud,omitempty"`
	Capabilities     []string         `json:"capabilities,omitempty"`
	AllowedResources AllowedResources `json:"allowedResources"`
}

// AllowedResources represents resources defined in the SAT
type AllowedResources struct {
	AllowedPartners          []string `json:"allowedPartners,omitempty"`
	AllowedServiceAccountIDs []string `json:"allowedServiceAccountIds,omitempty"`
	AllowedDeviceIDs         []string `json:"allowedDeviceIds,omitempty"`
	AllowedUserIDs           []string `json:"allowedUserIds,omitempty"`
	AllowedTNs               []string `json:"allowedTNs,omitempty"`
}

// ErrInvalidToken ...
type ErrInvalidToken struct {
	Issues []string
}

var _ error = ErrInvalidToken{}

func (e ErrInvalidToken) Error() string {
	return fmt.Sprintf("token was not valid for the following reason(s): %s", strings.Join(e.Issues, ", "))
}

// Valid checks that standard claims and sat claims are good
func (c *Claims) Valid() error {
	var issues []string
	now := time.Now().Unix()

	if c.Issuer == "" {
		issues = append(issues, "issuer is missing")
	}

	if c.ExpiresAt <= now {
		issues = append(issues, "token has already expired")
	}

	if c.NotBefore > now {
		issues = append(issues, "token is not yet valid")
	}

	if c.IssuedAt > now {
		issues = append(issues, "cannot use token before it has been issued")
	}

	if len(c.AllowedResources.AllowedPartners) < 1 {
		issues = append(issues, "at least one partner must be allowed")
	}

	if len(issues) < 1 {
		return nil
	}

	return ErrInvalidToken{issues}
}

// HasCapability Check if Claims has the given capability
func (c *Claims) HasCapability(capability string) bool {
	for _, c := range c.Capabilities {
		if c == capability {
			return true
		}
	}

	return false
}

// HasDevice Check if Claims has the given device id in AllowedDeviceIDs
func (c *Claims) HasDevice(deviceID string) bool {
	for _, d := range c.AllowedResources.AllowedDeviceIDs {
		if d == deviceID {
			return true
		}
	}

	return false
}

// Validator is used to check and parse a string to a valid sat token
type Validator interface {
	Validate(token string) (*Claims, error)
}

// WebValidator implements a Validator using an HTTP client using JWKS
type WebValidator struct {
	Client  *http.Client
	KeysURL string

	// storage for retrieved keys
	Keys map[string]interface{}
}

// ErrNoKIDParameter indicates that the provided JWT is missing the "kid"
// parameter
var ErrNoKIDParameter = errors.New("jwt header missing valid \"kid\" parameter")

func (v *WebValidator) fetchToken(token *jwt.Token) (interface{}, error) {
	kidI, ok := token.Header["kid"]
	if !ok {
		return nil, ErrNoKIDParameter
	}

	kid, ok := kidI.(string)
	if !ok {
		return nil, ErrNoKIDParameter
	}

	// check in local storage
	if key, ok := v.Keys[kid]; ok {
		return key, nil
	}

	var (
		start  = time.Now()
		status = "failure"
	)

	defer func() {
		keyRequestSeconds.WithLabelValues(status).Observe(time.Since(start).Seconds())
	}()

	// retrieve from http
	fetchURL, err := url.Parse(fmt.Sprintf("%s/%s",
		v.KeysURL,
		url.PathEscape(kid),
	))

	if err != nil {
		return nil, fmt.Errorf("failed to build url for sat pub key: %w", err)
	}

	res, err := v.Client.Get(fetchURL.String())

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sat pub key: %w", err)
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if (res.StatusCode / 100) != 2 {
		//nolint
		return nil, fmt.Errorf("attempt to fetch sat pub key failed with non-2xx status: %d", res.StatusCode)
	}

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read sat pub key body: %w", err)
	}

	publicKey, err := convertPublicKeyToPemFormat(data)
	if err != nil {
		return nil, fmt.Errorf("public key is not a valid rsa pub key: %w", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)

	if err != nil {
		return nil, fmt.Errorf("retrieve key is not a valid rsa pub key: %w", err)
	}

	v.Keys[kid] = key

	status = "success"

	return key, nil
}

// Validate parses the token against the configured JWKS and returns the
// extracted SAT claims
func (v *WebValidator) Validate(token string) (*Claims, error) {
	var satClaims Claims

	if _, err := jwt.ParseWithClaims(token, &satClaims, v.fetchToken); err != nil {
		return nil, err
	}

	return &satClaims, nil
}

type PublicKeyResponse struct {
	Kty string
	//xS256		string
	E   string
	Use string
	Kid string
	X5c []string
	N   string
}

func convertPublicKeyToPemFormat(data []byte) ([]byte, error) {
	publicKeyResponse := PublicKeyResponse{}

	err := json.Unmarshal(data, &publicKeyResponse)
	if err != nil {
		return nil, err
	}
	keyPart := [][]byte{[]byte("-----BEGIN RSA PUBLIC KEY-----"),
		[]byte(publicKeyResponse.X5c[0]),
		[]byte("-----END RSA PUBLIC KEY-----")}
	sep := []byte("\n")
	returnBytes := bytes.Join(keyPart, sep)
	return returnBytes, nil
}
