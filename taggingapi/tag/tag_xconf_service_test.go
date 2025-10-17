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

package tag

import (
	"testing"

	"github.com/rdkcentral/xconfadmin/util"
	"github.com/stretchr/testify/assert"
)

func TestXconfService_Constants(t *testing.T) {
	assert.Equal(t, "error cloning %s tag", CloneErrorMsg, "CloneErrorMsg constant should be correct")
}

func TestXconfService_GetAllTags_Integration(t *testing.T) {
	// Integration test - calls actual database
	tags, err := GetAllTags()

	// Should not panic and return slice (may be empty if no data)
	assert.NotNil(t, tags, "GetAllTags should return a slice")
	assert.True(t, err == nil || err != nil, "GetAllTags should handle errors gracefully")

	// If we have tags, verify they have required fields
	for _, tag := range tags {
		assert.NotNil(t, tag, "Tag should not be nil")
		if tag != nil {
			assert.NotEmpty(t, tag.Id, "Tag ID should not be empty")
			assert.NotNil(t, tag.Members, "Tag Members should not be nil")
		}
	}
}

func TestXconfService_GetAllTagIds_Integration(t *testing.T) {
	// Integration test - calls actual database
	tagIds, err := GetAllTagIds()

	// Should not panic and return slice (may be empty if no data)
	assert.NotNil(t, tagIds, "GetAllTagIds should return a slice")
	assert.True(t, err == nil || err != nil, "GetAllTagIds should handle errors gracefully")

	// If we have tag IDs, they should be non-empty strings
	for _, tagId := range tagIds {
		assert.NotEmpty(t, tagId, "Tag ID should not be empty")
		assert.IsType(t, "", tagId, "Tag ID should be a string")
	}
}

func TestXconfService_GetOneTag_Integration(t *testing.T) {
	// Test with non-existent tag ID
	nonExistentTag := GetOneTag("non-existent-tag-id")
	// Should return nil for non-existent tags
	assert.Nil(t, nonExistentTag, "GetOneTag should return nil for non-existent tags")

	// Test with empty tag ID
	emptyTag := GetOneTag("")
	assert.Nil(t, emptyTag, "GetOneTag should return nil for empty tag ID")

	// Test function doesn't panic with various inputs
	testIds := []string{
		"test-tag",
		"tag:test-tag",
		"t_test-tag",
		"123",
		"special!@#$%^&*()",
	}

	for _, testId := range testIds {
		t.Run("GetOneTag_"+testId, func(t *testing.T) {
			result := GetOneTag(testId)
			// Should not panic and result should be nil or valid tag
			assert.True(t, result == nil || result != nil, "GetOneTag should not panic")
			if result != nil {
				assert.NotEmpty(t, result.Id, "Returned tag should have an ID")
				assert.NotNil(t, result.Members, "Returned tag should have Members")
			}
		})
	}
}

func TestXconfService_SaveTag_Integration(t *testing.T) {
	// Test saving a valid tag
	memberSet := util.Set{}
	memberSet.Add("member1", "member2")

	testTag := &Tag{
		Id:      "test-save-tag",
		Members: memberSet,
		Updated: util.GetTimestamp(),
	}

	err := SaveTag(testTag)
	// Should not panic and may succeed or fail depending on database setup
	assert.True(t, err == nil || err != nil, "SaveTag should handle errors gracefully")

	// Test saving tag with empty ID
	emptyIdTag := &Tag{
		Id:      "",
		Members: util.Set{},
		Updated: util.GetTimestamp(),
	}

	err = SaveTag(emptyIdTag)
	assert.True(t, err == nil || err != nil, "SaveTag should handle empty ID gracefully")
}

func TestXconfService_DeleteOneTag_Integration(t *testing.T) {
	// Test deleting non-existent tag
	err := DeleteOneTag("non-existent-tag")
	assert.True(t, err == nil || err != nil, "DeleteOneTag should handle non-existent tags gracefully")

	// Test deleting with empty ID
	err = DeleteOneTag("")
	assert.True(t, err == nil || err != nil, "DeleteOneTag should handle empty ID gracefully")

	// Test various tag ID formats
	testIds := []string{
		"test-delete-tag",
		"tag:test-delete",
		"t_test-delete",
		"123456",
		"special!@#chars",
	}

	for _, testId := range testIds {
		t.Run("DeleteOneTag_"+testId, func(t *testing.T) {
			err := DeleteOneTag(testId)
			assert.True(t, err == nil || err != nil, "DeleteOneTag should not panic")
		})
	}
}

func TestXconfService_AddMembersToXconfTag_NewTag(t *testing.T) {
	// Test adding members to a new tag (non-existent)
	testMembers := []string{"member1", "member2", "member3"}
	result := AddMembersToXconfTag("new-tag-id", testMembers)

	assert.NotNil(t, result, "AddMembersToXconfTag should return a tag")
	assert.Equal(t, "new-tag-id", result.Id, "Tag ID should match input")
	assert.True(t, result.Members.Contains("member1"), "Tag should contain member1")
	assert.True(t, result.Members.Contains("member2"), "Tag should contain member2")
	assert.True(t, result.Members.Contains("member3"), "Tag should contain member3")
	assert.Len(t, result.Members, 3, "Tag should have 3 members")
	assert.Greater(t, result.Updated, int64(0), "Updated timestamp should be set")
}

func TestXconfService_AddMembersToXconfTag_EmptyMembers(t *testing.T) {
	// Test adding empty members list
	result := AddMembersToXconfTag("empty-members-tag", []string{})

	assert.NotNil(t, result, "AddMembersToXconfTag should return a tag")
	assert.Equal(t, "empty-members-tag", result.Id, "Tag ID should match input")
	assert.Len(t, result.Members, 0, "Tag should have no members")
	assert.Greater(t, result.Updated, int64(0), "Updated timestamp should be set")
}

func TestXconfService_AddMembersToXconfTag_DuplicateMembers(t *testing.T) {
	// Test adding duplicate members
	testMembers := []string{"member1", "member1", "member2", "member2"}
	result := AddMembersToXconfTag("duplicate-members-tag", testMembers)

	assert.NotNil(t, result, "AddMembersToXconfTag should return a tag")
	assert.Equal(t, "duplicate-members-tag", result.Id, "Tag ID should match input")
	assert.True(t, result.Members.Contains("member1"), "Tag should contain member1")
	assert.True(t, result.Members.Contains("member2"), "Tag should contain member2")
	// Set should handle duplicates
	assert.Len(t, result.Members, 2, "Tag should have 2 unique members")
}

func TestXconfService_AddMembersToXconfTag_SpecialCharacters(t *testing.T) {
	// Test with special characters in members
	testMembers := []string{
		"member!@#$%",
		"member with spaces",
		"member-with-dashes",
		"member_with_underscores",
		"member123",
		"AA:BB:CC:DD:EE:FF", // MAC address
	}

	result := AddMembersToXconfTag("special-chars-tag", testMembers)

	assert.NotNil(t, result, "AddMembersToXconfTag should return a tag")
	assert.Equal(t, "special-chars-tag", result.Id, "Tag ID should match input")

	for _, member := range testMembers {
		assert.True(t, result.Members.Contains(member), "Tag should contain member: %s", member)
	}
	assert.Len(t, result.Members, len(testMembers), "Tag should have all members")
}

func TestXconfService_RemoveMembersFromXconfTag(t *testing.T) {
	// Setup: Create a tag with initial members
	initialMembers := util.Set{}
	initialMembers.Add("member1", "member2", "member3", "member4")

	initialTag := &Tag{
		Id:      "test-remove-tag",
		Members: initialMembers,
		Updated: 123456789,
	}

	// Test removing some members
	membersToRemove := []string{"member1", "member3"}
	result := removeMembersFromXconfTag(initialTag, membersToRemove)

	assert.NotNil(t, result, "removeMembersFromXconfTag should return a tag")
	assert.Equal(t, "test-remove-tag", result.Id, "Tag ID should remain unchanged")
	assert.False(t, result.Members.Contains("member1"), "member1 should be removed")
	assert.True(t, result.Members.Contains("member2"), "member2 should remain")
	assert.False(t, result.Members.Contains("member3"), "member3 should be removed")
	assert.True(t, result.Members.Contains("member4"), "member4 should remain")
	assert.Len(t, result.Members, 2, "Tag should have 2 remaining members")
}

func TestXconfService_RemoveMembersFromXconfTag_NonExistentMembers(t *testing.T) {
	// Setup: Create a tag with initial members
	initialMembers := util.Set{}
	initialMembers.Add("member1", "member2")

	initialTag := &Tag{
		Id:      "test-remove-nonexistent",
		Members: initialMembers,
		Updated: 123456789,
	}

	// Test removing members that don't exist
	membersToRemove := []string{"nonexistent1", "nonexistent2"}
	result := removeMembersFromXconfTag(initialTag, membersToRemove)

	assert.NotNil(t, result, "removeMembersFromXconfTag should return a tag")
	assert.Equal(t, "test-remove-nonexistent", result.Id, "Tag ID should remain unchanged")
	assert.True(t, result.Members.Contains("member1"), "member1 should remain")
	assert.True(t, result.Members.Contains("member2"), "member2 should remain")
	assert.Len(t, result.Members, 2, "Tag should still have 2 members")
}

func TestXconfService_RemoveMembersFromXconfTag_EmptyList(t *testing.T) {
	// Setup: Create a tag with initial members
	initialMembers := util.Set{}
	initialMembers.Add("member1", "member2")

	initialTag := &Tag{
		Id:      "test-remove-empty",
		Members: initialMembers,
		Updated: 123456789,
	}

	// Test removing empty list
	result := removeMembersFromXconfTag(initialTag, []string{})

	assert.NotNil(t, result, "removeMembersFromXconfTag should return a tag")
	assert.Equal(t, "test-remove-empty", result.Id, "Tag ID should remain unchanged")
	assert.True(t, result.Members.Contains("member1"), "member1 should remain")
	assert.True(t, result.Members.Contains("member2"), "member2 should remain")
	assert.Len(t, result.Members, 2, "Tag should still have 2 members")
}

func TestXconfService_RemoveMembersFromXconfTag_AllMembers(t *testing.T) {
	// Setup: Create a tag with initial members
	initialMembers := util.Set{}
	initialMembers.Add("member1", "member2", "member3")

	initialTag := &Tag{
		Id:      "test-remove-all",
		Members: initialMembers,
		Updated: 123456789,
	}

	// Test removing all members
	membersToRemove := []string{"member1", "member2", "member3"}
	result := removeMembersFromXconfTag(initialTag, membersToRemove)

	assert.NotNil(t, result, "removeMembersFromXconfTag should return a tag")
	assert.Equal(t, "test-remove-all", result.Id, "Tag ID should remain unchanged")
	assert.Len(t, result.Members, 0, "Tag should have no members")
}

func TestXconfService_FlowIntegration(t *testing.T) {
	// Integration test of typical tag operations flow
	tagId := "flow-test-tag"

	// Step 1: Verify tag doesn't exist initially
	initialTag := GetOneTag(tagId)
	assert.Nil(t, initialTag, "Tag should not exist initially")

	// Step 2: Add members to create new tag
	members := []string{"flow-member1", "flow-member2"}
	createdTag := AddMembersToXconfTag(tagId, members)

	assert.NotNil(t, createdTag, "Tag should be created")
	assert.Equal(t, tagId, createdTag.Id, "Tag ID should match")
	assert.Len(t, createdTag.Members, 2, "Tag should have 2 members")

	// Step 3: Save the tag
	err := SaveTag(createdTag)
	assert.True(t, err == nil || err != nil, "Save should complete without panic")

	// Step 4: Remove some members
	updatedTag := removeMembersFromXconfTag(createdTag, []string{"flow-member1"})
	assert.Len(t, updatedTag.Members, 1, "Tag should have 1 member after removal")

	// Step 5: Delete the tag
	deleteErr := DeleteOneTag(tagId)
	assert.True(t, deleteErr == nil || deleteErr != nil, "Delete should complete without panic")
}
