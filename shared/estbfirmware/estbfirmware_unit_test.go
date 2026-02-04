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

package estbfirmware

import (
	"encoding/json"
	"testing"
	"time"

	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
)

func TestGetNormalizedMacAddressesValidAndInvalid(t *testing.T) {
	macs, err := GetNormalizedMacAddresses("AA:BB:CC:DD:EE:FF,aa:bb:cc:dd:ee:11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(macs) != 2 {
		t.Fatalf("expected 2 valid macs got %d", len(macs))
	}
	if _, err = GetNormalizedMacAddresses("bad-mac"); err == nil {
		t.Fatalf("expected error for invalid mac")
	}
}

func TestIpFilterIsWarehouse(t *testing.T) {
	f := &IpFilter{Id: "abc"}
	if !f.IsWarehouse() {
		t.Fatalf("expected warehouse for all lowercase letters")
	}
	f2 := &IpFilter{Id: "abc123"}
	if f2.IsWarehouse() {
		t.Fatalf("digits should break warehouse heuristic")
	}
}

func TestIsLetterAndIsLower(t *testing.T) {
	if !IsLetter("abc") || IsLetter("abc1") {
		t.Fatalf("IsLetter logic failure")
	}
	if !IsLower("abc") || IsLower("Abc") {
		t.Fatalf("IsLower logic failure")
	}
}

func TestHasProtocolSuffix(t *testing.T) {
	// The suffix check requires exact suffix per coreef constants (HTTP_SUFFIX/TFTP_SUFFIX)
	if !HasProtocolSuffix("name"+coreef.HTTP_SUFFIX) || !HasProtocolSuffix("other"+coreef.TFTP_SUFFIX) {
		t.Fatalf("expected protocol suffix recognition for %s and %s", coreef.HTTP_SUFFIX, coreef.TFTP_SUFFIX)
	}
	if HasProtocolSuffix("noSuffix") {
		t.Fatalf("did not expect suffix match")
	}
}

func TestSingletonFilterValueMarshalUnmarshalPercent(t *testing.T) {
	payload := `{"id":"PERCENT_FILTER_VALUE","type":"com.comcast.xconf.estbfirmware.PercentFilterValue","percentage":50,"percent":50,"envModelPercentages":{}}`
	var sfv SingletonFilterValue
	if err := json.Unmarshal([]byte(payload), &sfv); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !sfv.IsPercentFilterValue() || sfv.PercentFilterValue == nil || sfv.PercentFilterValue.Percentage != 50 {
		t.Fatalf("percent filter subtype not parsed correctly: %+v", sfv)
	}
	out, err := json.Marshal(&sfv)
	if err != nil || len(out) == 0 {
		t.Fatalf("marshal failed: %v", err)
	}
}

func TestGetRoundRobinIdByApplication(t *testing.T) {
	if GetRoundRobinIdByApplication("stb") != ROUND_ROBIN_FILTER_SINGLETON_ID {
		t.Fatalf("expected base id for stb")
	}
	id := GetRoundRobinIdByApplication("rdkcloud")
	if id == ROUND_ROBIN_FILTER_SINGLETON_ID || id == "" {
		t.Fatalf("expected prefixed id for non-stb app: %s", id)
	}
}

func TestConvertedContextBasicTransformation(t *testing.T) {
	ctx := map[string]string{
		"env":            "prod",
		"model":          "rng150",
		"eStbMac":        "AA:bb:CC:dd:EE:ff",
		"timeZoneOffset": "08:00",
		"time":           "01/02/2024 15:04:05",
		"capabilities":   "RCDL,rebootDecoupled,",
	}
	c := NewConvertedContext(ctx)
	if c.Env != "PROD" || c.Model != "RNG150" {
		t.Fatalf("expected uppercasing of env/model; got %s/%s", c.Env, c.Model)
	}
	// ConvertedContext currently keeps MAC format as provided if valid (no enforced upper-case in this path)
	if c.EstbMac == "" {
		t.Fatalf("expected estb mac to be set")
	}
	if c.TimeZone == nil || c.TimeZone.String() == "" {
		t.Fatalf("expected time zone set")
	}
	if c.Time == nil || c.Time.Format("01/02/2006 15:04:05") != "01/02/2024 15:04:05" {
		t.Fatalf("time parse failed: %v", c.Time)
	}
	// capability interpretation
	if !c.IsRcdl() || !c.IsRebootDecoupled() || c.IsSupportsFullHttpUrl() {
		t.Fatalf("capability flags mismatch RCDL=%v rebootDecoupled=%v fullHttp=%v", c.IsRcdl(), c.IsRebootDecoupled(), c.IsSupportsFullHttpUrl())
	}
	// properties building ensures time present
	props := c.GetProperties()
	if props["timeZone"] == "" || props["time"] == "" {
		t.Fatalf("expected timeZone/time keys in properties")
	}
}

func TestConvertedContextNilTimeFallback(t *testing.T) {
	ctx := map[string]string{"env": "qa"}
	c := NewConvertedContext(ctx)
	if c.Time == nil {
		t.Fatalf("expected time fallback")
	}
	// ensure updating time sets raw context
	now := time.Now()
	c.SetTime(now)
	if c.GetTime() == nil {
		t.Fatalf("expected GetTime after SetTime")
	}
}
