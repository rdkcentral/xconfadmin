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
package auth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"xconfadmin/common"
	xhttp "xconfadmin/http"
	"xconfadmin/util"
	xwhttp "xconfwebconfig/http"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

// Ws - webserver object
var (
	Ws *xhttp.WebconfigServer
)

// WebServerInjection - dependency injection
func WebServerInjection(ws *xhttp.WebconfigServer) {
	Ws = ws
}

func AuthInfoHandler(w http.ResponseWriter, r *http.Request) {
	// No permission check needed
	authResponse := xhttp.NewAuthResponse(r)
	if authResponse == nil {
		xhttp.WriteXconfResponse(w, http.StatusUnauthorized, []byte(""))
		return
	}
	response, _ := util.JSONMarshal(&authResponse)
	headers := map[string]string{
		"authProvider": getAuthProvider(),
	}
	xhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, response)
}

func AuthProvider(w http.ResponseWriter, r *http.Request) {
	responseRaw := map[string]string{
		"name": getAuthProvider(),
	}
	response, _ := util.JSONMarshal(&responseRaw)
	xhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func getAuthProvider() string {
	return common.AuthProvider
}

func BasicAuthHandler(w http.ResponseWriter, r *http.Request) {

	type AuthRequest struct {
		Username string `json:"login"`
		Password string `json:"password"`
	}

	var authRequest AuthRequest
	err := json.NewDecoder(r.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//Login and password are hardcoded now for testing purposes only.
	if authRequest.Username == "admin" && authRequest.Password == "admin" {
		claims := jwt.NewWithClaims(jwt.SigningMethodHS256,
			jwt.MapClaims{
				"username":    authRequest.Username,
				"lastName":    "",
				"displayName": authRequest.Username,
				"firstName":   authRequest.Username,
				"application": map[string]interface{}{
					"xconf_AS": []map[string]interface{}{
						{
							"role": "xconf_admin",
							"rights": []string{
								"read-common",
								"read-dcm-*",
								"read-firmware-*",
								"read-firmware-rule-templates",
								"read-telemetry-*",
								"read-changes-*",
								"write-changes-*",
								"view-tools",
								"write-common",
								"write-dcm-*",
								"write-firmware-*",
								"write-telemetry-*",
								"write-firmware-rule-templates",
								"write-tools",
							},
						},
					},
					"exp": time.Now().Add(time.Hour * 24).Unix(),
				}})
		token, err := claims.SignedString([]byte("xconf"))

		if err != nil {
			log.Error("Authentication Error : ", err)
			http.Error(w, "Authentication Error", http.StatusUnauthorized)
		}
		// Add the cookie to the response
		w.Header()[xhttp.AUTH_TOKEN] = []string{token}
		http.SetCookie(w, xhttp.NewAuthTokenCookie(token))
		headers := map[string]string{
			"Location": GetAdminUIUrlFromCookies(r),
		}
		xwhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusFound, []byte(""))
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

func getAdminUIUrlFromCookies(r *http.Request) string {
	cookie, err := r.Cookie(adminUrlCookieName)
	if err != nil {
		log.Errorf("%s: %s", adminUrlCookieName, err.Error())
		return defaultAdminUIHost
	}
	adminServiceUrl, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		log.Errorf("error unescaping %s cookie value %s: %s", adminUrlCookieName, cookie.Value, err.Error())
		return defaultAdminUIHost
	}
	return adminServiceUrl
}
