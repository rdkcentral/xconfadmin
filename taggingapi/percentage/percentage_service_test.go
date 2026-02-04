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

package percentage

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPercentageService_Constants(t *testing.T) {
	assert.Equal(t, uint64(506097522914230528), SipHashKey0, "SipHashKey0 constant should be correct")
	assert.Equal(t, uint64(1084818905618843912), SipHashKey1, "SipHashKey1 constant should be correct")
}

func TestCalculateHashAndPercent(t *testing.T) {
	testCases := []struct {
		name                 string
		input                string
		expectedHashRange    [2]float64 // min, max range for hash
		expectedPercentRange [2]float64 // min, max range for percent
	}{
		{
			name:                 "Empty string",
			input:                "",
			expectedHashRange:    [2]float64{0, float64(math.MaxInt64*2 + 1)},
			expectedPercentRange: [2]float64{0, 100},
		},
		{
			name:                 "Simple string",
			input:                "test",
			expectedHashRange:    [2]float64{0, float64(math.MaxInt64*2 + 1)},
			expectedPercentRange: [2]float64{0, 100},
		},
		{
			name:                 "MAC address",
			input:                "AA:BB:CC:DD:EE:FF",
			expectedHashRange:    [2]float64{0, float64(math.MaxInt64*2 + 1)},
			expectedPercentRange: [2]float64{0, 100},
		},
		{
			name:                 "Device ID",
			input:                "device-12345",
			expectedHashRange:    [2]float64{0, float64(math.MaxInt64*2 + 1)},
			expectedPercentRange: [2]float64{0, 100},
		},
		{
			name:                 "Long string",
			input:                "this-is-a-very-long-string-for-testing-hash-calculation",
			expectedHashRange:    [2]float64{0, float64(math.MaxInt64*2 + 1)},
			expectedPercentRange: [2]float64{0, 100},
		},
		{
			name:                 "Special characters",
			input:                "!@#$%^&*()_+-=[]{}|;:,.<>?",
			expectedHashRange:    [2]float64{0, float64(math.MaxInt64*2 + 1)},
			expectedPercentRange: [2]float64{0, 100},
		},
		{
			name:                 "Unicode characters",
			input:                "测试字符串",
			expectedHashRange:    [2]float64{0, float64(math.MaxInt64*2 + 1)},
			expectedPercentRange: [2]float64{0, 100},
		},
		{
			name:                 "Numbers only",
			input:                "1234567890",
			expectedHashRange:    [2]float64{0, float64(math.MaxInt64*2 + 1)},
			expectedPercentRange: [2]float64{0, 100},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, percent := CalculateHashAndPercent(tc.input)

			// Test hash value is within expected range
			assert.GreaterOrEqual(t, hash, tc.expectedHashRange[0], "Hash should be >= 0")
			assert.LessOrEqual(t, hash, tc.expectedHashRange[1], "Hash should be <= max range")

			// Test percent value is within expected range
			assert.GreaterOrEqual(t, percent, tc.expectedPercentRange[0], "Percent should be >= 0")
			assert.LessOrEqual(t, percent, tc.expectedPercentRange[1], "Percent should be <= 100")

			// Test that hash and percent are consistent
			vrange := float64(math.MaxInt64*2 + 1)
			expectedPercent := (hash / vrange) * 100
			assert.InDelta(t, expectedPercent, percent, 0.0001, "Percent calculation should be correct")
		})
	}
}

func TestCalculatePercent(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		minVal int
		maxVal int
	}{
		{
			name:   "Empty string",
			input:  "",
			minVal: 0,
			maxVal: 100,
		},
		{
			name:   "Simple string",
			input:  "test",
			minVal: 0,
			maxVal: 100,
		},
		{
			name:   "MAC address",
			input:  "AA:BB:CC:DD:EE:FF",
			minVal: 0,
			maxVal: 100,
		},
		{
			name:   "Device ID",
			input:  "device-12345",
			minVal: 0,
			maxVal: 100,
		},
		{
			name:   "Long string",
			input:  "this-is-a-very-long-string-for-testing",
			minVal: 0,
			maxVal: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculatePercent(tc.input)

			// Test result is within expected range
			assert.GreaterOrEqual(t, result, tc.minVal, "Percent should be >= 0")
			assert.LessOrEqual(t, result, tc.maxVal, "Percent should be <= 100")

			// Test that result is an integer
			assert.IsType(t, int(0), result, "Result should be an integer")

			// Test consistency with CalculateHashAndPercent
			_, floatPercent := CalculateHashAndPercent(tc.input)
			expectedInt := int(math.Round(floatPercent))
			assert.Equal(t, expectedInt, result, "CalculatePercent should match rounded CalculateHashAndPercent")
		})
	}
}

func TestCalculateHashAndPercent_Deterministic(t *testing.T) {
	// Test that the function is deterministic (same input -> same output)
	testInputs := []string{
		"test-input-1",
		"test-input-2",
		"AA:BB:CC:DD:EE:FF",
		"device-123",
		"",
	}

	for _, input := range testInputs {
		t.Run("Deterministic_"+input, func(t *testing.T) {
			hash1, percent1 := CalculateHashAndPercent(input)
			hash2, percent2 := CalculateHashAndPercent(input)

			assert.Equal(t, hash1, hash2, "Hash should be deterministic")
			assert.Equal(t, percent1, percent2, "Percent should be deterministic")
		})
	}
}

func TestCalculatePercent_Deterministic(t *testing.T) {
	// Test that the function is deterministic (same input -> same output)
	testInputs := []string{
		"test-input-1",
		"test-input-2",
		"AA:BB:CC:DD:EE:FF",
		"device-123",
		"",
	}

	for _, input := range testInputs {
		t.Run("Deterministic_"+input, func(t *testing.T) {
			result1 := CalculatePercent(input)
			result2 := CalculatePercent(input)

			assert.Equal(t, result1, result2, "CalculatePercent should be deterministic")
		})
	}
}

func TestCalculateHashAndPercent_Distribution(t *testing.T) {
	// Test that different inputs produce different results (good distribution)
	inputs := []string{
		"input1", "input2", "input3", "input4", "input5",
		"device-1", "device-2", "device-3", "device-4", "device-5",
		"AA:BB:CC:DD:EE:F1", "AA:BB:CC:DD:EE:F2", "AA:BB:CC:DD:EE:F3",
	}

	results := make(map[float64]string)
	percentResults := make(map[float64]string)

	for _, input := range inputs {
		hash, percent := CalculateHashAndPercent(input)

		// Check if we've seen this hash before
		if prevInput, exists := results[hash]; exists {
			t.Errorf("Hash collision: inputs '%s' and '%s' produced same hash %f", input, prevInput, hash)
		} else {
			results[hash] = input
		}

		// Store percent results for analysis
		percentResults[percent] = input

		// Verify hash and percent are valid numbers
		assert.False(t, math.IsNaN(hash), "Hash should not be NaN for input: %s", input)
		assert.False(t, math.IsInf(hash, 0), "Hash should not be Inf for input: %s", input)
		assert.False(t, math.IsNaN(percent), "Percent should not be NaN for input: %s", input)
		assert.False(t, math.IsInf(percent, 0), "Percent should not be Inf for input: %s", input)
	}

	// We should have unique results for different inputs
	assert.Equal(t, len(inputs), len(results), "All inputs should produce unique hash values")
}

func TestCalculatePercent_EdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Single character",
			input: "a",
		},
		{
			name:  "Repeated character",
			input: "aaaaaaaaaa",
		},
		{
			name:  "Newline character",
			input: "\n",
		},
		{
			name:  "Tab character",
			input: "\t",
		},
		{
			name:  "Space character",
			input: " ",
		},
		{
			name:  "Multiple spaces",
			input: "   ",
		},
		{
			name:  "Binary data simulation",
			input: "\x00\x01\x02\x03\xFF\xFE\xFD",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculatePercent(tc.input)

			// Test basic constraints
			assert.GreaterOrEqual(t, result, 0, "Percent should be >= 0")
			assert.LessOrEqual(t, result, 100, "Percent should be <= 100")

			// Test that it doesn't panic or return invalid values
			assert.NotPanics(t, func() {
				CalculatePercent(tc.input)
			}, "CalculatePercent should not panic")
		})
	}
}

func TestCalculateHashAndPercent_MathematicalProperties(t *testing.T) {
	// Test mathematical properties of the hash function
	testInput := "test-mathematical-properties"

	hash, percent := CalculateHashAndPercent(testInput)

	// Test offset calculation
	voffset := float64(math.MaxInt64 + 1)
	vrange := float64(math.MaxInt64*2 + 1)

	// Hash should be properly offset
	assert.GreaterOrEqual(t, hash, 0.0, "Hash should be >= 0 after offset")
	assert.LessOrEqual(t, hash, vrange, "Hash should be <= vrange")

	// Percent calculation verification
	expectedPercent := (hash / vrange) * 100
	assert.InDelta(t, expectedPercent, percent, 0.0001, "Percent calculation should be mathematically correct")

	// Verify the range calculations are correct
	assert.Equal(t, float64(math.MaxInt64)+1, voffset, "Offset calculation should be correct")
	assert.Equal(t, float64(math.MaxInt64*2)+1, vrange, "Range calculation should be correct")
}

func BenchmarkCalculateHashAndPercent(b *testing.B) {
	testInput := "benchmark-test-input-string"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateHashAndPercent(testInput)
	}
}

func BenchmarkCalculatePercent(b *testing.B) {
	testInput := "benchmark-test-input-string"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculatePercent(testInput)
	}
}
