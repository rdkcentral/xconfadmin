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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizationService_Constants(t *testing.T) {
	assert.Equal(t, "t_", Prefix, "Prefix constant should be 't_'")
	assert.Equal(t, "%s%s", Template, "Template constant should be '%s%s'")
}

func TestNormalizationService_ToNormalizedEcm(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "MAC address with colons lowercase",
			input:    "aa:bb:cc:dd:ee:ff",
			expected: "AABBCCDDEEFD", // ECM MAC (subtracts 2)
		},
		{
			name:     "MAC address with colons uppercase",
			input:    "AA:BB:CC:DD:EE:FF",
			expected: "AABBCCDDEEFD", // ECM MAC (subtracts 2)
		},
		{
			name:     "MAC address without colons lowercase",
			input:    "aabbccddeeff",
			expected: "AABBCCDDEEFD", // ECM MAC (subtracts 2)
		},
		{
			name:     "MAC address without colons uppercase",
			input:    "AABBCCDDEEFF",
			expected: "AABBCCDDEEFD", // ECM MAC (subtracts 2)
		},
		{
			name:     "MAC address with dashes",
			input:    "aa-bb-cc-dd-ee-ff",
			expected: "AABBCCDDEEFD", // ECM MAC (subtracts 2)
		},
		{
			name:     "Regular string (not MAC)",
			input:    "regular-string",
			expected: "REGULAR-STRING",
		},
		{
			name:     "String with spaces",
			input:    "  spaced string  ",
			expected: "SPACED STRING",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "String with only spaces",
			input:    "   ",
			expected: "",
		},
		{
			name:     "Mixed case alphanumeric",
			input:    "Test123String",
			expected: "TEST123STRING",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToNormalizedEcm(tc.input)
			assert.Equal(t, tc.expected, result, "ToNormalizedEcm(%q) should return %q", tc.input, tc.expected)
		})
	}
}

func TestNormalizationService_ToNormalized(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase string",
			input:    "lowercase",
			expected: "LOWERCASE",
		},
		{
			name:     "Uppercase string",
			input:    "UPPERCASE",
			expected: "UPPERCASE",
		},
		{
			name:     "Mixed case string",
			input:    "MixedCase",
			expected: "MIXEDCASE",
		},
		{
			name:     "String with spaces",
			input:    "  spaced string  ",
			expected: "SPACED STRING",
		},
		{
			name:     "String with numbers and symbols",
			input:    "test123!@#",
			expected: "TEST123!@#",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "String with only spaces",
			input:    "   ",
			expected: "",
		},
		{
			name:     "String with leading/trailing spaces",
			input:    "  trimmed  ",
			expected: "TRIMMED",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToNormalized(tc.input)
			assert.Equal(t, tc.expected, result, "ToNormalized(%q) should return %q", tc.input, tc.expected)
		})
	}
}

func TestNormalizationService_ToEstbIfMac(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid MAC address with colons lowercase",
			input:    "aa:bb:cc:dd:ee:ff",
			expected: "aa:bb:cc:dd:ee:ff", // Not converted (dashes not supported for ESTB)
		},
		{
			name:     "Valid MAC address with colons uppercase",
			input:    "AA:BB:CC:DD:EE:FF",
			expected: "AA:BB:CC:DD:EE:FF", // Not converted (colons not supported for ESTB)
		},
		{
			name:     "Valid MAC address without colons lowercase",
			input:    "aabbccddeeff",
			expected: "AABBCCDDEF01", // ESTB format (adds 2)
		},
		{
			name:     "Valid MAC address without colons uppercase",
			input:    "AABBCCDDEEFF",
			expected: "AABBCCDDEF01", // ESTB format (adds 2)
		},
		{
			name:     "Valid MAC address with dashes",
			input:    "aa-bb-cc-dd-ee-ff",
			expected: "aa-bb-cc-dd-ee-ff", // Not converted (dashes not supported for ESTB)
		},
		{
			name:     "Non-MAC string",
			input:    "not-a-mac-address",
			expected: "not-a-mac-address", // Unchanged
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Regular alphanumeric string",
			input:    "device123",
			expected: "device123", // Unchanged
		},
		{
			name:     "Invalid MAC format",
			input:    "aa:bb:cc:dd:ee", // Too short
			expected: "aa:bb:cc:dd:ee", // Unchanged
		},
		{
			name:     "Invalid MAC format with wrong characters",
			input:    "gg:hh:ii:jj:kk:ll",
			expected: "gg:hh:ii:jj:kk:ll", // Unchanged
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToEstbIfMac(tc.input)
			assert.Equal(t, tc.expected, result, "ToEstbIfMac(%q) should return %q", tc.input, tc.expected)
		})
	}
}

func TestNormalizationService_SetTagPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Tag without prefix",
			input:    "my-tag",
			expected: "t_my-tag",
		},
		{
			name:     "Tag already with prefix",
			input:    "t_my-tag",
			expected: "t_my-tag", // Should remain unchanged
		},
		{
			name:     "Empty tag",
			input:    "",
			expected: "t_",
		},
		{
			name:     "Tag with only prefix",
			input:    "t_",
			expected: "t_", // Should remain unchanged
		},
		{
			name:     "Complex tag name",
			input:    "complex-tag-name-123",
			expected: "t_complex-tag-name-123",
		},
		{
			name:     "Tag with special characters",
			input:    "tag!@#$%^&*()",
			expected: "t_tag!@#$%^&*()",
		},
		{
			name:     "Tag starting with different prefix",
			input:    "p_percentage-tag",
			expected: "t_p_percentage-tag",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SetTagPrefix(tc.input)
			assert.Equal(t, tc.expected, result, "SetTagPrefix(%q) should return %q", tc.input, tc.expected)
		})
	}
}

func TestNormalizationService_RemovePrefixFromTag(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Tag with prefix",
			input:    "t_my-tag",
			expected: "my-tag",
		},
		{
			name:     "Tag without prefix",
			input:    "my-tag",
			expected: "my-tag", // Should remain unchanged
		},
		{
			name:     "Tag with only prefix",
			input:    "t_",
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Complex tag with prefix",
			input:    "t_complex-tag-name-123",
			expected: "complex-tag-name-123",
		},
		{
			name:     "Tag with multiple prefixes",
			input:    "t_t_double-prefix",
			expected: "t_double-prefix", // Only removes first occurrence
		},
		{
			name:     "Tag with prefix-like substring",
			input:    "my-t_-tag",
			expected: "my-t_-tag", // Should remain unchanged as prefix is not at start
		},
		{
			name:     "Tag starting with different prefix",
			input:    "p_percentage-tag",
			expected: "p_percentage-tag", // Should remain unchanged
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := RemovePrefixFromTag(tc.input)
			assert.Equal(t, tc.expected, result, "RemovePrefixFromTag(%q) should return %q", tc.input, tc.expected)
		})
	}
}

func TestNormalizationService_RemovePrefixFromTags(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Mixed tags with and without prefixes",
			input:    []string{"t_tag1", "t_tag2", "tag3", "t_tag4"},
			expected: []string{"tag1", "tag2", "tag3", "tag4"},
		},
		{
			name:     "All tags with prefixes",
			input:    []string{"t_tag1", "t_tag2", "t_tag3"},
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name:     "All tags without prefixes",
			input:    []string{"tag1", "tag2", "tag3"},
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Single tag with prefix",
			input:    []string{"t_single-tag"},
			expected: []string{"single-tag"},
		},
		{
			name:     "Single tag without prefix",
			input:    []string{"single-tag"},
			expected: []string{"single-tag"},
		},
		{
			name:     "Tags with empty strings",
			input:    []string{"t_tag1", "", "t_", "tag2"},
			expected: []string{"tag1", "", "", "tag2"},
		},
		{
			name:     "Tags with special characters",
			input:    []string{"t_tag!@#", "t_tag$%^", "tag&*()"},
			expected: []string{"tag!@#", "tag$%^", "tag&*()"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Make a copy to avoid modifying the input slice
			inputCopy := make([]string, len(tc.input))
			copy(inputCopy, tc.input)

			result := removePrefixFromTags(inputCopy)
			assert.Equal(t, tc.expected, result, "removePrefixFromTags(%v) should return %v", tc.input, tc.expected)
		})
	}
}

func TestNormalizationService_SetTagPrefixAndRemoveRoundtrip(t *testing.T) {
	// Test that adding and removing prefix works correctly together
	testTags := []string{
		"simple-tag",
		"complex-tag-with-dashes",
		"tag123with456numbers",
		"tag!@#with$%^special&*()chars",
		"",
	}

	for _, tag := range testTags {
		t.Run(fmt.Sprintf("Roundtrip_%s", tag), func(t *testing.T) {
			// Add prefix then remove it
			withPrefix := SetTagPrefix(tag)
			withoutPrefix := RemovePrefixFromTag(withPrefix)

			assert.Equal(t, tag, withoutPrefix, "Roundtrip should preserve original tag")

			// Verify prefix was actually added
			if tag != "" && !strings.HasPrefix(tag, Prefix) {
				assert.True(t, strings.HasPrefix(withPrefix, Prefix), "Prefix should be added")
			}
		})
	}
}

func TestNormalizationService_TemplateUsage(t *testing.T) {
	// Test that the Template constant is used correctly
	testTag := "test-tag"
	expected := fmt.Sprintf(Template, Prefix, testTag)
	result := SetTagPrefix(testTag)

	assert.Equal(t, expected, result, "SetTagPrefix should use Template constant correctly")
	assert.Equal(t, "t_test-tag", result, "Result should match expected format")
}
