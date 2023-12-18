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
	"context"
	"net/http"
	"strings"

	xcommon "xconfadmin/common"
	"xconfwebconfig/common"
	xhttp "xconfwebconfig/http"

	"xconfwebconfig/db"

	log "github.com/sirupsen/logrus"
)

var (
	webConfServer *WebconfigServer
)

type WebconfigServer struct {
	XW_XconfServer *xhttp.XconfServer
	testOnly       bool
	AppName        string
}

// testOnly=true ==> running unit test
func NewWebconfigServer(sc *common.ServerConfig, testOnly bool, dc db.DatabaseClient) *WebconfigServer {
	conf := sc.Config
	appName := strings.Split(conf.GetString("xconfwebconfig.code_git_commit", "webconfigadmin-xconf"), "-")[0]

	webConfServer = &WebconfigServer{
		XW_XconfServer: xhttp.NewXconfServer(sc, testOnly, dc),
		testOnly:       testOnly,
		AppName:        appName,
	}
	if testOnly {
		webConfServer.XW_XconfServer.SetupMocks()
	}

	return webConfServer
}

func (s *WebconfigServer) AuthValidationMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", AppName())
		w.Header().Set("Xconf-Server", AppName())

		ctx := r.Context()

		// Check for SAT V2 token
		if satToken := getSatTokenFromRequest(r); satToken != "" {
			if subject, capabilities, err := getSubjectAndCapabilitiesFromSatToken(satToken); err != nil {
				log.Error(err.Error())
				http.Error(w, "invalid SAT token", http.StatusUnauthorized)
				return
			} else {
				r.Header.Set(AUTH_SUBJECT, subject)

				// Add capabilities to request context
				ctx = context.WithValue(ctx, CTX_KEY_CAPABILITIES, capabilities)
			}
		} else if authToken := getLoginTokenFromRequest(r); authToken != "" {
			if LoginToken, err := ValidateAndGetLoginToken(authToken); err != nil {
				log.Error(err.Error())
				http.Error(w, "invalid auth token", http.StatusUnauthorized)
				return
			} else {
				r.Header.Set(AUTH_SUBJECT, LoginToken.Subject)

				// Add Login token & permissions to request context
				ctx = context.WithValue(ctx, CTX_KEY_TOKEN, LoginToken)
				permissions := getPermissionsFromLoginToken(LoginToken)
				ctx = context.WithValue(ctx, CTX_KEY_PERMISSIONS, permissions)
			}
		} else if !xcommon.SatOn {
			log.Debug("Skipping validation...")
		} else {
			http.Error(w, "auth token not found", http.StatusUnauthorized)
			return
		}

		newReq := r.WithContext(ctx)
		xw := s.XW_XconfServer.LogRequestStarts(w, newReq)
		defer s.XW_XconfServer.LogRequestEnds(&xw, newReq)

		next.ServeHTTP(&xw, newReq)
	}
	return http.HandlerFunc(fn)
}

func (s *WebconfigServer) TestOnly() bool {
	return s.testOnly
}

// AppName is just a convenience func that returns the AppName, used in metrics
func AppName() string {
	return webConfServer.AppName
}
