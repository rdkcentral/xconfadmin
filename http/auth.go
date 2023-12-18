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
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

const (
	AUTHORIZATION = "Authorization"
	AUTH_TOKEN    = "token"
	AUTH_SUBJECT  = "X-Auth-Subject"
	UNKNOWN_USER  = "UNKNOWN_USER"

	KeysBaseURL = "https://sat-sample-url.net"
)

type AuthCtxKey string

func (c AuthCtxKey) String() string {
	return string(c)
}

const (
	CTX_KEY_TOKEN        AuthCtxKey = "Token"
	CTX_KEY_PERMISSIONS  AuthCtxKey = "Permissions"
	CTX_KEY_CAPABILITIES AuthCtxKey = "Capabilities"
)

type LoginToken struct {
	Issuer         string
	Subject        string
	Audience       string
	IssuedAt       float64
	ExpirationTime float64
	JwtId          string
	NotValidBefore float64
	LastName       string
	DisplayName    string
	FirstName      string
	PartnerId      string
	Email          string
	Application    []Application
}

type Application struct {
	Id      string
	Role    string
	Partner string
	Rights  []string
}

type AuthResponse struct {
	ServiceName     string   `json:"serviceName,omitempty"`
	Username        string   `json:"username,omitempty"`
	FirstName       string   `json:"firstName,omitempty"`
	LastName        string   `json:"lastName,omitempty"`
	Email           string   `json:"email,omitempty"`
	Permissions     []string `json:"permissions,omitempty"`
	OwnershipGroups []string `json:"ownershipGroups,omitempty"`
	OwnershipAdmin  bool     `json:"ownershipAdmin,omitempty"`
	Groups          []string `json:"groups,omitempty"`
}

func NewAuthResponse(r *http.Request) *AuthResponse {
	LoginToken := GetLoginTokenFromContext(r)
	if LoginToken == nil {
		return nil
	}
	groups := make([]string, len(LoginToken.Application))
	permissions := []string{}
	for i, app := range LoginToken.Application {
		groups[i] = app.Role
		permissions = append(permissions, app.Rights...)
	}
	authResponse := &AuthResponse{
		FirstName:       LoginToken.FirstName,
		LastName:        LoginToken.LastName,
		Username:        LoginToken.Subject,
		Groups:          groups,
		Permissions:     permissions,
		OwnershipGroups: []string{},
	}
	return authResponse
}

func NewErasedAuthTokenCookie() *http.Cookie {
	c := &http.Cookie{
		Name:   AUTH_TOKEN,
		Value:  "",
		Path:   "/",
		MaxAge: 0,
	}
	return c
}

func NewAuthTokenCookie(token string) *http.Cookie {
	c := &http.Cookie{
		Name:   AUTH_TOKEN,
		Value:  token,
		Path:   "/",
		MaxAge: math.MaxInt32,
	}
	return c
}

func GetLoginTokenFromContext(r *http.Request) *LoginToken {
	token := r.Context().Value(CTX_KEY_TOKEN)
	if token == nil {
		log.Debug("Login token not found in context")
		return nil
	}
	return token.(*LoginToken)
}

func GetPermissionsFromContext(r *http.Request) []string {
	permissions := r.Context().Value(CTX_KEY_PERMISSIONS)
	if permissions == nil {
		log.Debug("permissions not found in context")
		return []string{}
	}
	return permissions.([]string)
}

func GetCapabilitiesFromContext(r *http.Request) []string {
	capabilities := r.Context().Value(CTX_KEY_CAPABILITIES)
	if capabilities == nil {
		log.Debug("capabilities not found in context")
		return []string{}
	}
	return capabilities.([]string)
}

func ValidateAndGetLoginToken(authToken string) (*LoginToken, error) {
	if authToken == "" {
		return nil, errors.New("auth token is empty")
	}

	// first parse without validation to get the public key information
	jwtToken, _ := jwt.Parse(authToken, nil)
	if jwtToken == nil {
		return nil, errors.New("error parsing auth token")
	}

	// parse and validate
	token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("xconf"), nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing auth token with public key: %s", err.Error())
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("error getting claims from auth token")
	}
	return NewLoginToken(claims), nil
}

func NewLoginToken(claims jwt.MapClaims) *LoginToken {
	LoginToken := &LoginToken{}
	if lastName, ok := claims["lastName"].(string); ok {
		LoginToken.LastName = lastName
	}
	if subject, ok := claims["sub"].(string); ok {
		LoginToken.Subject = subject
	}
	if displayName, ok := claims["displayName"].(string); ok {
		LoginToken.DisplayName = displayName
	}
	if issuer, ok := claims["iss"].(string); ok {
		LoginToken.Issuer = issuer
	}
	if firstName, ok := claims["firstName"].(string); ok {
		LoginToken.FirstName = firstName
	}
	if audience, ok := claims["aud"].(string); ok {
		LoginToken.Audience = audience
	} else if audience, ok := claims["aud"].([]interface{}); ok {
		if len(audience) > 0 {
			if a, ok := audience[0].(string); ok {
				LoginToken.Audience = a
			}
		}
	}
	if notValidBefore, ok := claims["nbf"].(float64); ok {
		LoginToken.NotValidBefore = notValidBefore
	}
	if application, ok := claims["application"].(map[string]interface{}); ok {
		var applicationList []Application
		for appKey, appValue := range application {
			if list, ok := appValue.([]interface{}); ok {
				for i := range list {
					if m, ok := list[i].(map[string]interface{}); ok {
						app := Application{}
						app.Id = appKey
						if role, ok := m["role"]; ok {
							if r, ok := role.(string); ok {
								app.Role = r
							}
						}
						if partner, ok := m["partner"]; ok {
							if p, ok := partner.(string); ok {
								app.Partner = p
							}
						}
						if rights, ok := m["rights"]; ok {
							if rightsList, ok := rights.([]interface{}); ok {
								rList := make([]string, len(rightsList))
								for j, right := range rightsList {
									if r, ok := right.(string); ok {
										rList[j] = r
									}
								}
								app.Rights = rList
							}
						}
						applicationList = append(applicationList, app)
					}
				}
			}
		}
		LoginToken.Application = applicationList
	}
	if partnerId, ok := claims["partnerID"].(string); ok {
		LoginToken.PartnerId = partnerId
	}
	if expirationTime, ok := claims["exp"].(float64); ok {
		LoginToken.ExpirationTime = expirationTime
	}
	if issuedAt, ok := claims["iat"].(float64); ok {
		LoginToken.IssuedAt = issuedAt
	}
	if jwtId, ok := claims["jti"].(string); ok {
		LoginToken.JwtId = jwtId
	}
	if email, ok := claims["email"].(string); ok {
		LoginToken.Email = email
	}
	return LoginToken
}

// Get SAT V2 token
func getSatTokenFromRequest(r *http.Request) string {
	return r.Header.Get(AUTHORIZATION)
}

func getLoginTokenFromRequest(r *http.Request) string {
	authToken := r.Header.Get(AUTH_TOKEN)
	if authToken == "" {
		cookie, err := r.Cookie(AUTH_TOKEN)
		if err == nil {
			authToken = cookie.Value
		}
	}
	return authToken
}

func getPermissionsFromLoginToken(LoginToken *LoginToken) []string {
	permissions := []string{}
	for _, app := range LoginToken.Application {
		permissions = append(permissions, app.Rights...)
	}
	return permissions
}

func getWebValidator() WebValidator {
	keysUrl := webConfServer.XW_XconfServer.Config.GetString("xconfwebconfig.sat.host", KeysBaseURL)
	return WebValidator{
		Client:  http.DefaultClient,
		KeysURL: keysUrl,
		Keys:    make(map[string]interface{}),
	}
}

func getSubjectAndCapabilitiesFromSatToken(token string) (string, []string, error) {
	// 1 Extract Sat Token
	fragments := strings.SplitN(token, " ", 2)
	switch len(fragments) {
	case 1:
		token = fragments[0]
	case 2:
		token = fragments[1]
	}
	if strings.TrimSpace(token) == "" {
		return "", nil, errors.New("unable to extract required sat token")
	}
	// 2 Validate Sat Token
	validator := getWebValidator()
	claims, err := validator.Validate(token)
	if err != nil {
		return "", nil, errors.New("unable to extract valid sat token")
	}
	// get capabilities
	capabilities := claims.Capabilities
	if len(capabilities) == 0 {
		return "", nil, errors.New("unable to extract capabilities from sat token")
	}
	return claims.Subject, capabilities, nil
}
