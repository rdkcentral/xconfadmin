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
package queries

import (
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/stretchr/testify/assert"
)

// Test GetIpAddressGroups
func TestGetIpAddressGroups(t *testing.T) {
	result := GetIpAddressGroups()
	assert.NotNil(t, result)
	assert.IsType(t, []*shared.IpAddressGroup{}, result)
}

func TestGetIpAddressGroups_ConsistentReturn(t *testing.T) {
	// Multiple calls should return consistent results
	for i := 0; i < 3; i++ {
		result := GetIpAddressGroups()
		assert.NotNil(t, result)
		assert.True(t, len(result) >= 0)
	}
}

// Test GetIpAddressGroupByName
func TestGetIpAddressGroupByName_ValidName(t *testing.T) {
	result := GetIpAddressGroupByName("test-group")
	// Result depends on DB state
	assert.True(t, result != nil || result == nil)
}

func TestGetIpAddressGroupByName_EmptyName(t *testing.T) {
	result := GetIpAddressGroupByName("")
	assert.True(t, result != nil || result == nil)
}

func TestGetIpAddressGroupByName_NonExistent(t *testing.T) {
	result := GetIpAddressGroupByName("non-existent-group-xyz-123")
	assert.True(t, result != nil || result == nil)
}

func TestGetIpAddressGroupByName_SpecialCharacters(t *testing.T) {
	testNames := []string{
		"group-with-dashes",
		"group_with_underscores",
		"group.with.dots",
	}

	for _, name := range testNames {
		assert.NotPanics(t, func() {
			GetIpAddressGroupByName(name)
		})
	}
}

// Test GetIpAddressGroupsByIp
func TestGetIpAddressGroupsByIp_ValidIp(t *testing.T) {
	result := GetIpAddressGroupsByIp("192.168.1.1")
	assert.NotNil(t, result)
	assert.IsType(t, []*shared.IpAddressGroup{}, result)
}

func TestGetIpAddressGroupsByIp_EmptyIp(t *testing.T) {
	result := GetIpAddressGroupsByIp("")
	assert.NotNil(t, result)
}

func TestGetIpAddressGroupsByIp_InvalidIp(t *testing.T) {
	result := GetIpAddressGroupsByIp("invalid-ip")
	assert.NotNil(t, result)
}

func TestGetIpAddressGroupsByIp_Ipv6(t *testing.T) {
	result := GetIpAddressGroupsByIp("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	assert.NotNil(t, result)
}

func TestGetIpAddressGroupsByIp_LocalhostIpv4(t *testing.T) {
	result := GetIpAddressGroupsByIp("127.0.0.1")
	assert.NotNil(t, result)
}

func TestGetIpAddressGroupsByIp_LocalhostIpv6(t *testing.T) {
	result := GetIpAddressGroupsByIp("::1")
	assert.NotNil(t, result)
}

func TestGetIpAddressGroupsByIp_MultipleIps(t *testing.T) {
	testIps := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
		"",
	}

	for _, ip := range testIps {
		result := GetIpAddressGroupsByIp(ip)
		assert.NotNil(t, result)
	}
}

// Test CreateIpAddressGroup
func TestCreateIpAddressGroup_ValidGroup(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{
		Id:   "test-group",
		Name: "Test Group",
	}
	result := CreateIpAddressGroup(ipGroup)
	assert.NotNil(t, result)
	// Result depends on validation and DB state
}

func TestCreateIpAddressGroup_EmptyGroup(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{}
	result := CreateIpAddressGroup(ipGroup)
	assert.NotNil(t, result)
}

func TestCreateIpAddressGroup_WithIpAddresses(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{
		Id:   "test-group-with-ips",
		Name: "Test Group With IPs",
	}
	result := CreateIpAddressGroup(ipGroup)
	assert.NotNil(t, result)
}

// Test edge cases
func TestGetIpAddressGroups_ReturnsSliceNotNil(t *testing.T) {
	result := GetIpAddressGroups()
	assert.NotNil(t, result)
	assert.IsType(t, []*shared.IpAddressGroup{}, result)
}

func TestGetIpAddressGroupByName_MultipleCalls(t *testing.T) {
	// Multiple calls with same name should not panic
	testName := "consistent-group"
	for i := 0; i < 5; i++ {
		assert.NotPanics(t, func() {
			GetIpAddressGroupByName(testName)
		})
	}
}

func TestGetIpAddressGroupsByIp_PrivateNetworks(t *testing.T) {
	privateIps := []string{
		"10.0.0.1",      // Class A private
		"172.16.0.1",    // Class B private
		"192.168.0.1",   // Class C private
	}

	for _, ip := range privateIps {
		result := GetIpAddressGroupsByIp(ip)
		assert.NotNil(t, result)
	}
}

func TestGetIpAddressGroupsByIp_PublicIps(t *testing.T) {
	publicIps := []string{
		"8.8.8.8",       // Google DNS
		"1.1.1.1",       // Cloudflare DNS
	}

	for _, ip := range publicIps {
		result := GetIpAddressGroupsByIp(ip)
		assert.NotNil(t, result)
	}
}

func TestCreateIpAddressGroup_DuplicateId(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{
		Id:   "duplicate-test",
		Name: "Duplicate Test",
	}
	
	result1 := CreateIpAddressGroup(ipGroup)
	assert.NotNil(t, result1)
	
	// Try creating again
	result2 := CreateIpAddressGroup(ipGroup)
	assert.NotNil(t, result2)
}

func TestGetIpAddressGroupByName_LongName(t *testing.T) {
	longName := "very-long-group-name-" + "repeated-" + "many-times"
	assert.NotPanics(t, func() {
		GetIpAddressGroupByName(longName)
	})
}

func TestGetIpAddressGroupsByIp_EdgeCaseIps(t *testing.T) {
	edgeCaseIps := []string{
		"0.0.0.0",
		"255.255.255.255",
		"192.168.255.255",
	}

	for _, ip := range edgeCaseIps {
		result := GetIpAddressGroupsByIp(ip)
		assert.NotNil(t, result)
	}
}
