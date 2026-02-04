package util

import (
	"sort"
	"testing"

	"gotest.tools/assert"
)

func TestStringMap_Keys_EmptyMap(t *testing.T) {
	// Test with empty map
	m := StringMap{}
	keys := m.Keys()

	// Empty map returns nil slice (not an empty slice)
	assert.Equal(t, len(keys), 0, "Keys() should return slice with length 0 for empty map")
}

func TestStringMap_Keys_SingleElement(t *testing.T) {
	// Test with single element
	m := StringMap{
		"key1": "value1",
	}
	keys := m.Keys()

	assert.Equal(t, len(keys), 1, "Keys() should return slice with 1 element")
	assert.Equal(t, keys[0], "key1", "Keys() should return correct key")
}

func TestStringMap_Keys_MultipleElements(t *testing.T) {
	// Test with multiple elements
	m := StringMap{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	keys := m.Keys()

	assert.Equal(t, len(keys), 3, "Keys() should return slice with 3 elements")

	// Sort keys for deterministic comparison
	sort.Strings(keys)
	expectedKeys := []string{"key1", "key2", "key3"}
	sort.Strings(expectedKeys)

	assert.DeepEqual(t, keys, expectedKeys)
}

func TestStringMap_Keys_AllKeysPresent(t *testing.T) {
	// Test that all keys are present in the result
	m := StringMap{
		"apple":  "red",
		"banana": "yellow",
		"cherry": "red",
		"date":   "brown",
	}
	keys := m.Keys()

	assert.Equal(t, len(keys), 4, "Keys() should return all keys")

	// Verify all expected keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	assert.Assert(t, keyMap["apple"], "apple should be in keys")
	assert.Assert(t, keyMap["banana"], "banana should be in keys")
	assert.Assert(t, keyMap["cherry"], "cherry should be in keys")
	assert.Assert(t, keyMap["date"], "date should be in keys")
}

func TestStringMap_Keys_NoDuplicates(t *testing.T) {
	// Test that no duplicate keys are returned
	m := StringMap{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	keys := m.Keys()

	// Check for duplicates
	seen := make(map[string]bool)
	for _, key := range keys {
		assert.Assert(t, !seen[key], "Keys() should not return duplicate keys: %s", key)
		seen[key] = true
	}
}

func TestStringMap_Keys_WithEmptyStringKeys(t *testing.T) {
	// Test with empty string as key
	m := StringMap{
		"":     "empty key",
		"key1": "value1",
	}
	keys := m.Keys()

	assert.Equal(t, len(keys), 2, "Keys() should return slice with 2 elements including empty string key")

	// Check that empty string is in keys
	hasEmptyKey := false
	for _, key := range keys {
		if key == "" {
			hasEmptyKey = true
			break
		}
	}
	assert.Assert(t, hasEmptyKey, "Keys() should include empty string key")
}

func TestStringMap_Keys_WithSpecialCharacters(t *testing.T) {
	// Test with special characters in keys
	m := StringMap{
		"key-with-dash":   "value1",
		"key_with_under":  "value2",
		"key.with.dot":    "value3",
		"key/with/slash":  "value4",
		"key with spaces": "value5",
	}
	keys := m.Keys()

	assert.Equal(t, len(keys), 5, "Keys() should return all keys with special characters")

	// Verify specific keys exist
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	assert.Assert(t, keyMap["key-with-dash"], "key-with-dash should be present")
	assert.Assert(t, keyMap["key_with_under"], "key_with_under should be present")
	assert.Assert(t, keyMap["key.with.dot"], "key.with.dot should be present")
	assert.Assert(t, keyMap["key/with/slash"], "key/with/slash should be present")
	assert.Assert(t, keyMap["key with spaces"], "key with spaces should be present")
}
