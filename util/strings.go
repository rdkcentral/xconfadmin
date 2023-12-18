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
	"sort"
	"strings"
	"xconfwebconfig/util"
)

func StringSliceContains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ValidateAndNormalizeMacAddress is to validate and convert MAC address to XX:XX:XX:XX:XX:XX
func ValidateAndNormalizeMacAddress(macaddr string) (string, error) {
	// 1st validates the mac address
	_, err := util.MACAddressValidator(macaddr)
	if err != nil {
		return "", err
	}

	// Replace all dash, colon or period from MAC address
	mac := util.AlphaNumericMacAddress(macaddr)
	return util.ToColonMac(mac), nil
}
