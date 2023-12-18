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
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"xconfadmin/common"
	xcommon "xconfadmin/common"
	xwhttp "xconfwebconfig/http"
	"xconfwebconfig/util"
)

type EntityMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

const (
	TYPE_409 = "EntityConflictException"
	TYPE_400 = "ValidationRuntimeException"
	TYPE_404 = "EntityNotFoundException"
	TYPE_500 = "InternalServerErrorException"
	TYPE_501 = "NotImplementedException"
	TYPE_415 = "UnsupportedMediaTypeException"
)

func AdminError(w http.ResponseWriter, err error) {
	status := xcommon.GetXconfErrorStatusCode(err)
	WriteAdminErrorResponse(w, status, err.Error())
}

// helper function to write a failure json response matching xconf java admin response
func WriteAdminErrorResponse(w http.ResponseWriter, status int, errMsg string) {
	typeMsg := ""
	switch status {
	case 400:
		typeMsg = TYPE_400
	case 409:
		typeMsg = TYPE_409
	case 404:
		typeMsg = TYPE_404
	case 500:
		typeMsg = TYPE_500
	case 501:
		typeMsg = TYPE_501
	case 415:
		typeMsg = TYPE_415
	}
	resp := xcommon.HttpAdminErrorResponse{
		Status:  status,
		Type:    typeMsg,
		Message: errMsg,
	}
	writeByMarshal(w, status, resp)
}

func writeByMarshal(w http.ResponseWriter, status int, o interface{}) {
	if rbytes, err := util.JSONMarshal(o); err == nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(status)
		w.Write(rbytes)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		xwhttp.LogError(w, err)
	}
}

func CreateContentDispositionHeader(fileName string) map[string]string {
	return map[string]string{"Content-Disposition": fmt.Sprintf("attachment; filename=%s.json", escapeXml(fileName))}
}

func CreateNumberOfItemsHttpHeaders(size int) map[string]string {
	return map[string]string{"numberOfItems": strconv.Itoa(size)}
}

func escapeXml(str string) string {
	var buffer bytes.Buffer
	xml.EscapeText(&buffer, []byte(str))
	return buffer.String()
}

// ReturnJsonResponse - return JSON response to api
func ReturnJsonResponse(res interface{}, r *http.Request) ([]byte, error) {
	acceptStr := r.Header.Get("Accept")
	if acceptStr == "" {
		if data, err := util.JSONMarshal(res); err != nil {
			return nil, common.NewXconfError(http.StatusInternalServerError, fmt.Sprintf("JSON marshal error: %v", err))
		} else {
			return data, nil
		}
	}
	acceptTokens := strings.Split(acceptStr, ",")
	for _, acceptVal := range acceptTokens {

		if strings.Contains(strings.ToLower(acceptVal), "*/*") || strings.Contains(strings.ToLower(acceptVal), "application/json") {
			if data, err := util.JSONMarshal(res); err != nil {
				return nil, common.NewXconfError(http.StatusInternalServerError, fmt.Sprintf("JSON marshal error: %v", err))
			} else {
				return data, nil
			}
		}
	}
	return nil, common.NewXconfError(http.StatusNotAcceptable, "At this time only JSON input/output is supported")
}

func ContextTypeHeader(r *http.Request) string {
	return fmt.Sprintf("%s:%s", "application/json", "charset=UTF-8")
}
