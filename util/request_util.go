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

package util

import (
	//"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

/**
 * First we check 'X-Forwarded-For' header, then 'HA-Forwarded-For' if it exists and contains valid ip address.
 * Usually format of header is 'X-Forwarded-For: client, proxy1, proxy2' so we split string by "[,]" and take first part.
 * @param req http request info
 * @return valid ip address or empty ""
 */
func grepIpAddressFromXFF(r *http.Request) string {
	XffHeaders := make([]string, 0)
	XffHeaders = append(XffHeaders, "HA-Forwarded-For")
	XffHeaders = append(XffHeaders, "X-Forwarded-For")
	for _, headerName := range XffHeaders {
		header := r.Header.Get(headerName)
		if len(header) > 0 {
			splited := strings.Split(header, "[,]")
			if len(splited) > 0 && len(splited[0]) > 0 && net.ParseIP(splited[0]) != nil {
				return strings.TrimSpace(splited[0])
			}
		}
	}
	return ""
}

/**
 * Most important is IP from 'X-Forwarded-For' or 'HA-Forwarded-For' header. If it's valid we use it.
 * If not then check value from context. If valid - use it.
 * If not then read remote address from request info. If valid - use it.
 * At edge case when nothing above is a correct IP address then we fallback to '0.0.0.0'
 * @param contextIpAddress ip address from request context
 * @param req http request meta info
 */
func FindValidIpAddress(req *http.Request, contextIpAddress string) string {
	if net.ParseIP(contextIpAddress) != nil {
		log.Debug("supplied valid IP in context: " + contextIpAddress)
		return contextIpAddress
	}
	ipFromHeader := grepIpAddressFromXFF(req)
	if (net.ParseIP(ipFromHeader)) != nil {
		log.Debug("supplied valid IP address in XFF header: " + ipFromHeader)
		return ipFromHeader
	} else if net.ParseIP(req.RemoteAddr) != nil {
		if len(ipFromHeader) > 0 {
			log.Warn("invalid IP is specified in XFF header: " + ipFromHeader)
		}
		if len(contextIpAddress) > 0 {
			log.Warn("invalid IP is specified in context: " + contextIpAddress)
		}
		log.Debug("using IP from request remote address: " + req.RemoteAddr)
		return req.RemoteAddr
	} else {
		log.Warn("using 0.0.0.0 because IP was invalid in XFF, context, request remote address")
		return "0.0.0.0"
	}
}

func AddQueryParamsToContextMap(r *http.Request, contextMap map[string]string) {
	// check query params for data, these can override body data if they both exist
	queryParams := r.URL.Query()
	for k, v := range queryParams {
		key, _ := url.PathUnescape(k)
		value, _ := url.PathUnescape(v[0])
		contextMap[key] = value
	}
}

func AddBodyParamsToContextMap(body string, contextMap map[string]string) {
	if len(body) > 0 {
		// check body for data
		bodyList := strings.Split(body, "&")
		for _, item := range bodyList {
			index := strings.Index(item, "=")
			if index > 0 {
				k, _ := url.PathUnescape(item[:index])
				v, _ := url.PathUnescape(item[index+1:])
				contextMap[k] = v
			}
		}
	}
}
