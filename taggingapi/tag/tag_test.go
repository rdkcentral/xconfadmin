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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rdkcentral/xconfadmin/util"
)

// func TestNewTagInf(t *testing.T) {
// 	tagInf := NewTagInf()
// 	tag, ok := tagInf.(*Tag)

// 	assert.True(t, ok, "NewTagInf should return a *Tag")
// 	assert.NotNil(t, tag, "Returned tag should not be nil")
// 	assert.Equal(t, "", tag.Id, "New tag should have empty Id")
// 	assert.Equal(t, int64(0), tag.Updated, "New tag should have zero Updated timestamp")
// 	assert.NotNil(t, tag.Members, "Members should be initialized")
// 	assert.Equal(t, 0, len(tag.Members), "Members should be empty")
// }

func TestTagClone(t *testing.T) {
	// Create original tag with data
	memberSet := util.Set{}
	memberSet.Add("member1")
	memberSet.Add("member2")
	memberSet.Add("member3")

	originalTag := &Tag{
		Id:      "test-tag",
		Members: memberSet,
		Updated: 1234567890,
	}

	// Clone the tag
	clonedTag, err := originalTag.Clone()

	assert.NoError(t, err, "Clone should not return an error")
	assert.NotNil(t, clonedTag, "Cloned tag should not be nil")

	// Verify the clone has the same data
	assert.Equal(t, originalTag.Id, clonedTag.Id, "Cloned tag should have same Id")
	assert.Equal(t, originalTag.Updated, clonedTag.Updated, "Cloned tag should have same Updated timestamp")
	assert.Equal(t, len(originalTag.Members), len(clonedTag.Members), "Cloned tag should have same number of members")

	// Verify it's a deep copy by checking members
	assert.True(t, clonedTag.Members.Contains("member1"), "Cloned tag should contain member1")
	assert.True(t, clonedTag.Members.Contains("member2"), "Cloned tag should contain member2")
	assert.True(t, clonedTag.Members.Contains("member3"), "Cloned tag should contain member3")

	// Verify it's actually a different object (not same reference)
	assert.NotSame(t, originalTag, clonedTag, "Clone should be a different object instance")

	// Modify clone and ensure original is not affected
	clonedTag.Id = "modified-tag"
	clonedTag.Members.Add("new-member")

	assert.Equal(t, "test-tag", originalTag.Id, "Original tag Id should not be modified")
	assert.False(t, originalTag.Members.Contains("new-member"), "Original tag should not contain new member")
}

func TestTagMarshalJSON(t *testing.T) {
	// Create tag with data
	memberSet := util.Set{}
	memberSet.Add("member1")
	memberSet.Add("member2")

	tag := &Tag{
		Id:      "test-tag",
		Members: memberSet,
		Updated: 1234567890,
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(tag)

	assert.NoError(t, err, "Marshal should not return an error")
	assert.NotNil(t, jsonBytes, "JSON bytes should not be nil")

	// Unmarshal to verify structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NoError(t, err, "Should be able to unmarshal JSON")

	// Verify JSON structure
	assert.Equal(t, "test-tag", result["id"], "JSON should contain correct id")
	assert.Equal(t, float64(1234567890), result["updated"], "JSON should contain correct updated timestamp")

	// Verify members is an array
	members, ok := result["members"].([]interface{})
	assert.True(t, ok, "Members should be an array in JSON")
	assert.Equal(t, 2, len(members), "Members array should contain 2 items")

	// Convert to strings and verify content
	memberStrings := make([]string, len(members))
	for i, member := range members {
		memberStrings[i] = member.(string)
	}
	assert.Contains(t, memberStrings, "member1", "Members should contain member1")
	assert.Contains(t, memberStrings, "member2", "Members should contain member2")
}

func TestTagMarshalJSON_EmptyMembers(t *testing.T) {
	tag := &Tag{
		Id:      "empty-tag",
		Members: util.Set{},
		Updated: 9876543210,
	}

	jsonBytes, err := json.Marshal(tag)

	assert.NoError(t, err, "Marshal should not return an error")

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NoError(t, err, "Should be able to unmarshal JSON")

	assert.Equal(t, "empty-tag", result["id"])
	assert.Equal(t, float64(9876543210), result["updated"])

	// Members could be null or an empty array - both are valid
	members := result["members"]
	if members != nil {
		membersArray, ok := members.([]interface{})
		assert.True(t, ok, "Members should be an array if not null")
		assert.Equal(t, 0, len(membersArray), "Members array should be empty")
	}
}

func TestTagUnmarshalJSON(t *testing.T) {
	jsonStr := `{
		"id": "test-tag",
		"members": ["member1", "member2", "member3"],
		"updated": 1234567890
	}`

	var tag Tag
	err := json.Unmarshal([]byte(jsonStr), &tag)

	assert.NoError(t, err, "Unmarshal should not return an error")
	assert.Equal(t, "test-tag", tag.Id, "Tag should have correct Id")
	assert.Equal(t, int64(1234567890), tag.Updated, "Tag should have correct Updated timestamp")

	// Verify members
	assert.Equal(t, 3, len(tag.Members), "Tag should have 3 members")
	assert.True(t, tag.Members.Contains("member1"), "Tag should contain member1")
	assert.True(t, tag.Members.Contains("member2"), "Tag should contain member2")
	assert.True(t, tag.Members.Contains("member3"), "Tag should contain member3")
}

func TestTagUnmarshalJSON_EmptyMembers(t *testing.T) {
	jsonStr := `{
		"id": "empty-tag",
		"members": [],
		"updated": 9876543210
	}`

	var tag Tag
	err := json.Unmarshal([]byte(jsonStr), &tag)

	assert.NoError(t, err, "Unmarshal should not return an error")
	assert.Equal(t, "empty-tag", tag.Id, "Tag should have correct Id")
	assert.Equal(t, int64(9876543210), tag.Updated, "Tag should have correct Updated timestamp")
	assert.Equal(t, 0, len(tag.Members), "Tag should have no members")
}

func TestTagUnmarshalJSON_NullMembers(t *testing.T) {
	jsonStr := `{
		"id": "null-members-tag",
		"members": null,
		"updated": 5555555555
	}`

	var tag Tag
	err := json.Unmarshal([]byte(jsonStr), &tag)

	assert.NoError(t, err, "Unmarshal should not return an error")
	assert.Equal(t, "null-members-tag", tag.Id, "Tag should have correct Id")
	assert.Equal(t, int64(5555555555), tag.Updated, "Tag should have correct Updated timestamp")
	assert.Equal(t, 0, len(tag.Members), "Tag with null members should have empty set")
}

func TestTagUnmarshalJSON_InvalidJSON(t *testing.T) {
	invalidJSONs := []string{
		`{"id": "test", "members": ["member1"], "updated": "invalid"}`, // Invalid updated type
		`{"id": "test", "members": "invalid", "updated": 123}`,         // Invalid members type
		`{invalid json}`, // Malformed JSON
		`{"id": 123, "members": [], "updated": 456}`, // Invalid id type
	}

	for i, invalidJSON := range invalidJSONs {
		t.Run(fmt.Sprintf("InvalidJSON_%d", i), func(t *testing.T) {
			var tag Tag
			err := json.Unmarshal([]byte(invalidJSON), &tag)
			assert.Error(t, err, "Unmarshal should return an error for invalid JSON")
		})
	}
}

func TestTagMarshalUnmarshalRoundtrip(t *testing.T) {
	// Create original tag
	memberSet := util.Set{}
	memberSet.Add("member1")
	memberSet.Add("member2")
	memberSet.Add("special-chars-!@#$%")
	memberSet.Add("unicode-тест")

	originalTag := &Tag{
		Id:      "roundtrip-test",
		Members: memberSet,
		Updated: 1577836800, // January 1, 2020
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(originalTag)
	assert.NoError(t, err, "Marshal should not return an error")

	// Unmarshal back to Tag
	var unmarshaledTag Tag
	err = json.Unmarshal(jsonBytes, &unmarshaledTag)
	assert.NoError(t, err, "Unmarshal should not return an error")

	// Verify roundtrip preserved all data
	assert.Equal(t, originalTag.Id, unmarshaledTag.Id, "Id should be preserved in roundtrip")
	assert.Equal(t, originalTag.Updated, unmarshaledTag.Updated, "Updated should be preserved in roundtrip")
	assert.Equal(t, len(originalTag.Members), len(unmarshaledTag.Members), "Member count should be preserved")

	// Check all members are preserved
	for _, member := range originalTag.Members.ToSlice() {
		assert.True(t, unmarshaledTag.Members.Contains(member), "Member %s should be preserved in roundtrip", member)
	}
}

func TestTagJSONWithSpecialCharacters(t *testing.T) {
	// Test with special characters and edge cases
	memberSet := util.Set{}
	memberSet.Add("member with spaces")
	memberSet.Add("member-with-dashes")
	memberSet.Add("member_with_underscores")
	memberSet.Add("member.with.dots")
	memberSet.Add("member/with/slashes")
	memberSet.Add("member\\with\\backslashes")
	memberSet.Add("member\"with\"quotes")
	memberSet.Add("member'with'apostrophes")
	memberSet.Add("") // empty member

	tag := &Tag{
		Id:      "special-chars-test",
		Members: memberSet,
		Updated: 1234567890,
	}

	// Should handle special characters correctly
	jsonBytes, err := json.Marshal(tag)
	assert.NoError(t, err, "Should handle special characters in marshal")

	var unmarshaledTag Tag
	err = json.Unmarshal(jsonBytes, &unmarshaledTag)
	assert.NoError(t, err, "Should handle special characters in unmarshal")

	// Verify all special characters are preserved
	assert.Equal(t, len(tag.Members), len(unmarshaledTag.Members), "All members should be preserved")
	for _, member := range tag.Members.ToSlice() {
		assert.True(t, unmarshaledTag.Members.Contains(member), "Special character member should be preserved: %s", member)
	}
}
