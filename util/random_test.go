package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestRandomDouble(t *testing.T) {
	// Test that RandomDouble returns a value between 0.0 and 1.0
	for i := 0; i < 100; i++ {
		result := RandomDouble()
		assert.Assert(t, result >= 0.0, "RandomDouble should return value >= 0.0, got %f", result)
		assert.Assert(t, result < 1.0, "RandomDouble should return value < 1.0, got %f", result)
	}
}

func TestRandomPercentage(t *testing.T) {
	// Test that RandomPercentage returns a value between 0 and 99
	for i := 0; i < 100; i++ {
		result := RandomPercentage()
		assert.Assert(t, result >= 0, "RandomPercentage should return value >= 0, got %d", result)
		assert.Assert(t, result < 100, "RandomPercentage should return value < 100, got %d", result)
	}
}

func TestRandomDoubleDistribution(t *testing.T) {
	// Test that RandomDouble produces different values (basic randomness check)
	values := make(map[float64]bool)
	for i := 0; i < 50; i++ {
		values[RandomDouble()] = true
	}
	// With 50 calls, we should have many unique values (at least 40)
	assert.Assert(t, len(values) > 40, "RandomDouble should produce diverse values, got %d unique values out of 50", len(values))
}

func TestRandomPercentageDistribution(t *testing.T) {
	// Test that RandomPercentage produces different values (basic randomness check)
	values := make(map[int]bool)
	for i := 0; i < 100; i++ {
		values[RandomPercentage()] = true
	}
	// With 100 calls, we should have many unique values (at least 30)
	assert.Assert(t, len(values) > 30, "RandomPercentage should produce diverse values, got %d unique values out of 100", len(values))
}
