/*
 * If not stated otherwise in this file or this component's Licenses.txt file the
 * following copyright and licenses apply:
 *
 * Copyright 2018 RDK Management
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
 * Author: cpatel550
 * Created: 06/24/2021
 */

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
