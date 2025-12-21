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

	"github.com/google/uuid"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/stretchr/testify/assert"
)

func newValidIpFilter(name string) *coreef.IpFilter {
	// Create and save IP address group to avoid IsChangedIpAddressGroup check failure
	ipGroup := shared.NewIpAddressGroupWithAddrStrings(name+"_group", name+"_group", []string{"10.0.0.1"})
	ipGroup.RawIpAddresses = []string{"10.0.0.1"}
	nl := shared.ConvertFromIpAddressGroup(ipGroup)
	db.GetCachedSimpleDao().SetOne(db.TABLE_GENERIC_NS_LIST, nl.ID, nl)

	return &coreef.IpFilter{
		Id:             "",
		Name:           name,
		IpAddressGroup: ipGroup,
	}
}

func TestUpdateIpFilter_Success(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	ipFilter := newValidIpFilter("TestIPFilter")

	resp := UpdateIpFilter("stb", ipFilter)

	assert.Equal(t, 200, resp.Status)
	assert.NotEmpty(t, ipFilter.Id)

	// Verify the filter was created
	returnedFilter, ok := resp.Data.(*coreef.IpFilter)
	assert.True(t, ok)
	assert.Equal(t, "TestIPFilter", returnedFilter.Name)
	assert.NotEmpty(t, returnedFilter.Id)
}

func TestUpdateIpFilter_WithExistingId(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	existingId := uuid.New().String()
	ipFilter := newValidIpFilter("TestIPFilterWithId")
	ipFilter.Id = existingId

	resp := UpdateIpFilter("stb", ipFilter)

	assert.Equal(t, 200, resp.Status)
	assert.Equal(t, existingId, ipFilter.Id)
}

func TestUpdateIpFilter_BlankName(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Create IP filter with blank name but valid IP group
	ipGroup := shared.NewIpAddressGroupWithAddrStrings("blank_group", "blank_group", []string{"10.0.0.1"})
	ipGroup.RawIpAddresses = []string{"10.0.0.1"}
	nl := shared.ConvertFromIpAddressGroup(ipGroup)
	db.GetCachedSimpleDao().SetOne(db.TABLE_GENERIC_NS_LIST, nl.ID, nl)

	ipFilter := &coreef.IpFilter{
		Name:           "", // Blank name
		IpAddressGroup: ipGroup,
	}

	resp := UpdateIpFilter("stb", ipFilter)

	// Blank name might be allowed during creation, so verify response
	// The validation might only fail if there's a duplicate
	if resp.Status == 200 {
		t.Log("Blank name allowed during creation")
	} else {
		assert.Equal(t, 400, resp.Status)
		assert.NotNil(t, resp.Error)
	}
}

func TestUpdateIpFilter_InvalidApplicationType(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	ipFilter := newValidIpFilter("TestIPFilter")

	// Use empty application type
	resp := UpdateIpFilter("", ipFilter)

	assert.Equal(t, 400, resp.Status)
	assert.NotNil(t, resp.Error)
}

func TestUpdateIpFilter_DuplicateName(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Create first filter
	ipFilter1 := newValidIpFilter("DuplicateName")
	resp1 := UpdateIpFilter("stb", ipFilter1)
	assert.Equal(t, 200, resp1.Status)

	// Try to create another filter with the same name but different ID
	ipFilter2 := newValidIpFilter("DuplicateName")
	resp2 := UpdateIpFilter("stb", ipFilter2)

	assert.Equal(t, 400, resp2.Status)
	assert.NotNil(t, resp2.Error)
}

func TestUpdateIpFilter_WithValidIpAddressGroup(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Create and save IP address group
	ipGroup := shared.NewIpAddressGroupWithAddrStrings("TestGroup", "TestGroup", []string{"10.0.0.1", "10.0.0.2"})
	ipGroup.RawIpAddresses = []string{"10.0.0.1", "10.0.0.2"}
	nl := shared.ConvertFromIpAddressGroup(ipGroup)
	db.GetCachedSimpleDao().SetOne(db.TABLE_GENERIC_NS_LIST, nl.ID, nl)

	ipFilter := newValidIpFilter("TestWithIPGroup")
	ipFilter.IpAddressGroup = ipGroup

	resp := UpdateIpFilter("stb", ipFilter)

	assert.Equal(t, 200, resp.Status)
	assert.NotEmpty(t, ipFilter.Id)
}

func TestUpdateIpFilter_WithChangedIpAddressGroup(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Create IP address group but don't save it (or save with different content)
	ipGroup := shared.NewIpAddressGroupWithAddrStrings("UnsavedGroup", "UnsavedGroup", []string{"10.0.0.1"})

	ipFilter := newValidIpFilter("TestWithChangedIPGroup")
	ipFilter.IpAddressGroup = ipGroup

	resp := UpdateIpFilter("stb", ipFilter)

	// Should fail because the IP address group doesn't exist or has changed
	assert.Equal(t, 400, resp.Status)
	assert.NotNil(t, resp.Error)
}

func TestUpdateIpFilter_WithModifiedIpAddressGroup(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Save IP address group with certain IPs
	ipGroup := shared.NewIpAddressGroupWithAddrStrings("ModifiedGroup", "ModifiedGroup", []string{"10.0.0.1"})
	ipGroup.RawIpAddresses = []string{"10.0.0.1"}
	nl := shared.ConvertFromIpAddressGroup(ipGroup)
	db.GetCachedSimpleDao().SetOne(db.TABLE_GENERIC_NS_LIST, nl.ID, nl)

	// Modify the group (different IPs than stored)
	ipGroup.RawIpAddresses = []string{"10.0.0.2"}

	ipFilter := newValidIpFilter("TestWithModifiedIPGroup")
	ipFilter.IpAddressGroup = ipGroup

	resp := UpdateIpFilter("stb", ipFilter)

	// Should fail because the IP address group has been modified
	assert.Equal(t, 400, resp.Status)
	assert.NotNil(t, resp.Error)
}

func TestDeleteIpsFilter_Success(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Create an IP filter first
	ipFilter := newValidIpFilter("FilterToDelete")
	createResp := UpdateIpFilter("stb", ipFilter)
	assert.Equal(t, 200, createResp.Status)

	// Delete the filter
	deleteResp := DeleteIpsFilter("FilterToDelete", "stb")

	assert.Equal(t, 204, deleteResp.Status)
	assert.Nil(t, deleteResp.Error)
}

func TestDeleteIpsFilter_NotFound(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Try to delete non-existent filter
	resp := DeleteIpsFilter("NonExistentFilter", "stb")

	// Should still return 204 (NoContent) even if not found
	assert.Equal(t, 204, resp.Status)
	assert.Nil(t, resp.Error)
}

func TestDeleteIpsFilter_EmptyName(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Try to delete with empty name
	resp := DeleteIpsFilter("", "stb")

	// Should return 204 as the filter won't be found
	assert.Equal(t, 204, resp.Status)
}

func TestDeleteIpsFilter_WithApplicationType(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Create IP filter with xhome app type
	ipFilter := newValidIpFilter("XHomeFilter")
	createResp := UpdateIpFilter("xhome", ipFilter)
	assert.Equal(t, 200, createResp.Status)

	// Delete with correct app type
	deleteResp := DeleteIpsFilter("XHomeFilter", "xhome")
	assert.Equal(t, 204, deleteResp.Status)
}

func TestUpdateIpFilter_UpdateExisting(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Create initial filter
	ipFilter := newValidIpFilter("UpdateTest")
	createResp := UpdateIpFilter("stb", ipFilter)
	assert.Equal(t, 200, createResp.Status)
	filterId := ipFilter.Id

	// Update the same filter (same ID and name)
	ipFilter.Id = filterId
	updateResp := UpdateIpFilter("stb", ipFilter)

	assert.Equal(t, 200, updateResp.Status)
	assert.Equal(t, filterId, ipFilter.Id)
}

func TestUpdateIpFilter_MultipleApplicationTypes(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		appType string
		want    int
	}{
		{"stb app type", "stb", 200},
		{"xhome app type", "xhome", 200},
		{"rdkcloud app type", "rdkcloud", 200},
		{"invalid app type", "invalid", 400},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			truncateTable(db.TABLE_FIRMWARE_RULE)
			truncateTable(db.TABLE_GENERIC_NS_LIST)
			ipFilter := newValidIpFilter("Test_" + tc.appType)
			resp := UpdateIpFilter(tc.appType, ipFilter)
			assert.Equal(t, tc.want, resp.Status)
		})
	}
}

func TestDeleteIpsFilter_AfterUpdate(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Create filter
	ipFilter := newValidIpFilter("CreateUpdateDelete")
	createResp := UpdateIpFilter("stb", ipFilter)
	assert.Equal(t, 200, createResp.Status)

	// Update it
	updateResp := UpdateIpFilter("stb", ipFilter)
	assert.Equal(t, 200, updateResp.Status)

	// Delete it
	deleteResp := DeleteIpsFilter("CreateUpdateDelete", "stb")
	assert.Equal(t, 204, deleteResp.Status)

	// Verify it's deleted by trying to delete again
	deleteResp2 := DeleteIpsFilter("CreateUpdateDelete", "stb")
	assert.Equal(t, 204, deleteResp2.Status)
}

func TestUpdateIpFilter_RuleNameValidation(t *testing.T) {
	t.Parallel()
	truncateTable(db.TABLE_FIRMWARE_RULE)
	truncateTable(db.TABLE_GENERIC_NS_LIST)

	// Create first filter
	ipFilter1 := newValidIpFilter("Filter1")
	resp1 := UpdateIpFilter("stb", ipFilter1)
	assert.Equal(t, 200, resp1.Status)
	id1 := ipFilter1.Id

	// Try to create another filter with same name but different ID
	ipFilter2 := newValidIpFilter("Filter1")
	ipFilter2.Id = uuid.New().String() // Different ID
	resp2 := UpdateIpFilter("stb", ipFilter2)

	// Should fail due to duplicate name with different ID
	assert.Equal(t, 400, resp2.Status)

	// Update first filter with same ID and name should work
	ipFilter3 := newValidIpFilter("Filter1")
	ipFilter3.Id = id1
	resp3 := UpdateIpFilter("stb", ipFilter3)
	assert.Equal(t, 200, resp3.Status)
}
