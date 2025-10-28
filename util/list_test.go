package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestContains(t *testing.T) {
	days := []string{"mon", "tue", "wed", "thu"}
	c1 := Contains(days, "wed")
	assert.Assert(t, c1)
	c2 := Contains(days, "fri")
	assert.Assert(t, !c2)

	assert.Assert(t, Contains([]int{1, 2, 3, 4}, 3))
	assert.Assert(t, !Contains([]int{1, 2, 3, 4}, 9))
	assert.Assert(t, Contains([]string{"red", "orange", "yellow", "green", "blue"}, "orange"))
	assert.Assert(t, !Contains([]string{"red", "orange", "yellow", "green", "blue"}, "violet"))
	assert.Assert(t, Contains([]float64{1.1, 2.2, 3.3, 4.4}, 3.3))
	assert.Assert(t, !Contains([]float64{1.1, 2.2, 3.3, 4.4}, 9.2))

	type unsupported string
	wed := unsupported("wed")
	collection := []unsupported{"mon", "tue", "wed"}
	ok := Contains(collection, wed)
	assert.Assert(t, !ok)

	tue := "tue"
	ok = Contains(collection, tue)
	assert.Assert(t, !ok)

	ok = Contains([]string{"mon", "tue", "wed"}, wed)
	assert.Assert(t, !ok)
}

func TestContainsInt(t *testing.T) {
	values := []int{1, 2, 3, 4}
	c1 := ContainsInt(values, 3)
	assert.Assert(t, c1)
	c2 := ContainsInt(values, 5)
	assert.Assert(t, !c2)
}

func TestCaseInsensitiveContains(t *testing.T) {
	days := []string{"lon", "tue", "Wed", "thu"}
	c1 := CaseInsensitiveContains(days, "weD")
	assert.Assert(t, c1)
	c2 := CaseInsensitiveContains(days, "fri")
	assert.Assert(t, !c2)
}

func TestContainsAny(t *testing.T) {
	c1 := []string{"dog", "cat", "hamster", "fish"}
	c2 := []string{"dog", "cat"}
	found := ContainsAny(c1, c2)
	assert.Assert(t, found)

	c3 := []string{"bird", "squirrel"}
	found = ContainsAny(c1, c3)
	assert.Assert(t, !found)
}

func TestIsEqualStringSlice(t *testing.T) {
	c1 := []string{"dog", "cat", "hamster", "fish"}
	c2 := []string{"dog", "cat"}
	assert.Assert(t, !StringElementsMatch(c1, c2))

	c2 = []string{"fish", "dog", "hamster", "cat"}
	assert.Assert(t, StringElementsMatch(c1, c2))

	c2 = []string{"dog", "hamster", "cat", "fishy"}
	assert.Assert(t, !StringElementsMatch(c1, c2))

	c2 = []string{"Dog", "hamster", "cat", "fish"}
	assert.Assert(t, !StringElementsMatch(c1, c2))

	c2 = nil
	assert.Assert(t, !StringElementsMatch(c1, c2))
}

func TestStringCopySlice(t *testing.T) {
	c1 := []string{"dog", "cat", "hamster", "fish"}
	c2 := StringCopySlice(c1)
	assert.Assert(t, StringElementsMatch(c1, c2))
}

func TestPutIfValuePresent(t *testing.T) {
	m := make(map[string]interface{})

	// Test with non-empty string
	PutIfValuePresent(m, "key1", "value1")
	assert.Equal(t, m["key1"], "value1")

	// Test with empty string - should not be added
	PutIfValuePresent(m, "key2", "")
	_, exists := m["key2"]
	assert.Assert(t, !exists)

	// Test with nil value - should not be added
	PutIfValuePresent(m, "key3", nil)
	_, exists = m["key3"]
	assert.Assert(t, !exists)

	// Test with non-empty slice
	PutIfValuePresent(m, "key4", []string{"a", "b"})
	assert.Equal(t, len(m["key4"].([]string)), 2)

	// Test with empty slice - should not be added
	PutIfValuePresent(m, "key5", []string{})
	_, exists = m["key5"]
	assert.Assert(t, !exists)

	// Test with integer
	PutIfValuePresent(m, "key6", 42)
	assert.Equal(t, m["key6"], 42)
}

func TestStringArrayContains(t *testing.T) {
	collection := []string{"apple", "banana", "cherry"}

	// Test with value containing element
	assert.Assert(t, StringArrayContains(collection, "I like bananas"))

	// Test with value not containing any element
	assert.Assert(t, !StringArrayContains(collection, "I like oranges"))

	// Test with exact match
	assert.Assert(t, StringArrayContains(collection, "apple"))

	// Test with empty collection
	assert.Assert(t, !StringArrayContains([]string{}, "test"))
}

func TestNewStringSet(t *testing.T) {
	// Test with normal collection
	collection := []string{"a", "b", "c", "a"}
	set := NewStringSet(collection)
	assert.Equal(t, len(set), 3) // "a" is duplicated
	_, exists := set["a"]
	assert.Assert(t, exists)
	_, exists = set["b"]
	assert.Assert(t, exists)
	_, exists = set["c"]
	assert.Assert(t, exists)

	// Test with nil collection
	nilSet := NewStringSet(nil)
	assert.Assert(t, nilSet == nil)

	// Test with empty collection
	emptySet := NewStringSet([]string{})
	assert.Equal(t, len(emptySet), 0)
}
