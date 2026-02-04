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

package util

import (
	"net/http"
	"testing"

	"gotest.tools/assert"
)

func TestGrepIpAddressFromXFF(t *testing.T) {
	r := &http.Request{}

	adata := grepIpAddressFromXFF(r)
	assert.Equal(t, adata, "")

	// hdr := map[string][]string{
	// 	"HTTP_X_HEADER": {"test_value"},
	// }
	// r.Header = hdr
	// adata = grepIpAddressFromXFF(r)
	// assert.Equal(t, adata, "")

	// r.Header.Add("HA-Forwarded-For", "276.999.919.800, 10.10.10.1")
	// adata = grepIpAddressFromXFF(r)
	// assert.Equal(t, adata, "")

	// r.Header = nil
	// hdr = map[string][]string{
	// 	"HA-Forwarded-For": {"10.10.10.1, 192.168.1.2"},
	// }
	// r.Header = hdr
	// adata = grepIpAddressFromXFF(r)
	// assert.Equal(t, adata, "")

	r.Header = nil
	hdr := map[string][]string{
		"X-Forwarded-For": {"192.168.1.1, 255.55.44.53"},
	}
	r.Header = hdr

	assert.Equal(t, r.Header.Get("X-Forwarded-For"), "192.168.1.1, 255.55.44.53")
	adata = grepIpAddressFromXFF(r)
	assert.Equal(t, adata, "")
}

func TestFindValidIpAddress(t *testing.T) {
	// Test with valid context IP
	r := &http.Request{}
	ip := FindValidIpAddress(r, "192.168.1.100")
	assert.Equal(t, ip, "192.168.1.100")

	// Test with invalid context IP and valid RemoteAddr
	r.RemoteAddr = "10.0.0.1"
	ip = FindValidIpAddress(r, "invalid-ip")
	assert.Equal(t, ip, "10.0.0.1")

	// Test with X-Forwarded-For header
	r = &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"203.0.113.1"},
		},
	}
	ip = FindValidIpAddress(r, "")
	assert.Equal(t, ip, "203.0.113.1")

	// Test fallback to 0.0.0.0
	r = &http.Request{
		RemoteAddr: "invalid",
	}
	ip = FindValidIpAddress(r, "")
	assert.Equal(t, ip, "0.0.0.0")
}

func TestAddQueryParamsToContextMap(t *testing.T) {
	contextMap := make(map[string]string)

	// Test with query parameters
	r, _ := http.NewRequest("GET", "http://example.com?key1=value1&key2=value2&key3=value%203", nil)
	AddQueryParamsToContextMap(r, contextMap)

	assert.Equal(t, contextMap["key1"], "value1")
	assert.Equal(t, contextMap["key2"], "value2")
	assert.Equal(t, contextMap["key3"], "value 3") // URL decoded

	// Test with no query parameters
	contextMap2 := make(map[string]string)
	r2, _ := http.NewRequest("GET", "http://example.com", nil)
	AddQueryParamsToContextMap(r2, contextMap2)
	assert.Equal(t, len(contextMap2), 0)
}

func TestAddBodyParamsToContextMap(t *testing.T) {
	contextMap := make(map[string]string)

	// Test with body parameters
	body := "param1=value1&param2=value2&param3=value%203"
	AddBodyParamsToContextMap(body, contextMap)

	assert.Equal(t, contextMap["param1"], "value1")
	assert.Equal(t, contextMap["param2"], "value2")
	assert.Equal(t, contextMap["param3"], "value 3") // URL decoded

	// Test with empty body
	contextMap2 := make(map[string]string)
	AddBodyParamsToContextMap("", contextMap2)
	assert.Equal(t, len(contextMap2), 0)

	// Test with malformed parameters (no equals sign)
	contextMap3 := make(map[string]string)
	AddBodyParamsToContextMap("invalidparam", contextMap3)
	assert.Equal(t, len(contextMap3), 0)

	// Test with single parameter
	contextMap4 := make(map[string]string)
	AddBodyParamsToContextMap("single=value", contextMap4)
	assert.Equal(t, contextMap4["single"], "value")
}
