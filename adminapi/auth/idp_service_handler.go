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
package auth

import (
	"fmt"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"

	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/util"
)

const (
	adminUrlCookieName = "admin-ui-location"
	defaultAdminUIHost = "http://localhost:8081"
)

func GetAdminUIUrlFromCookies(r *http.Request) string {
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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, xhttp.NewErasedAuthTokenCookie())

	idpAuthServer := Ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.idp_service_name")
	continueUrl := GetAdminUIUrlFromCookies(r) + Ws.XW_XconfServer.ServerConfig.GetString(fmt.Sprintf("xconfwebconfig.%v.idp_logout_after_path", idpAuthServer), "/"+getAuthProvider()+"/logout/after")

	logoutUrl := Ws.IdpServiceConnector.GetFullLogoutUrl(continueUrl)
	if err := Ws.IdpServiceConnector.Logout(logoutUrl); err != nil {
		msg := fmt.Sprintf("error logging out of SSO [%s]: %s", logoutUrl, err.Error())
		xhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(msg))
	} else {
		xhttp.WriteXconfResponse(w, http.StatusOK, nil)
	}
}

func LogoutAfterHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, xhttp.NewErasedAuthTokenCookie())

	adminUrl := fmt.Sprintf("%s/#/authorization", GetAdminUIUrlFromCookies(r))
	headers := map[string]string{
		"Location": adminUrl,
	}
	xhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusFound, []byte(""))
}

func CodeHandler(w http.ResponseWriter, r *http.Request) {
	codeList, ok := r.URL.Query()["code"]
	if !ok {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(""))
		return
	}
	code := codeList[0]
	log.Debugf("getting login token for code=%s", code)
	token := Ws.IdpServiceConnector.GetToken(code)
	if token == "" {
		xhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte("Idp service error"))
		return
	}
	if loginToken, err := xhttp.ValidateAndGetLoginToken(token); err != nil {
		log.Error(err.Error())
		xhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte("error validating Login token"))
		return
	} else {
		log.Debugf("received and parsed token: %+v", loginToken)
		addTokenToResponse(token, w)
	}

	headers := map[string]string{
		"Location": GetAdminUIUrlFromCookies(r),
	}
	xhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusFound, []byte(""))
}

func LoginUrlHandler(w http.ResponseWriter, r *http.Request) {
	idpAuthServer := Ws.XW_XconfServer.ServerConfig.GetString("xconfwebconfig.xconf.idp_service_name")
	continueUrl := GetAdminUIUrlFromCookies(r) + Ws.XW_XconfServer.ServerConfig.GetString(fmt.Sprintf("xconfwebconfig.%v.idp_code_path", idpAuthServer), "/"+getAuthProvider()+"/code")
	loginUrl := Ws.IdpServiceConnector.GetFullLoginUrl(continueUrl)
	responseMap := map[string]string{
		"url": loginUrl,
	}
	response, _ := util.JSONMarshal(&responseMap)
	xhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
}

func addTokenToResponse(token string, response http.ResponseWriter) {
	response.Header()[xhttp.AUTH_TOKEN] = []string{token}
	http.SetCookie(response, xhttp.NewAuthTokenCookie(token))
}
