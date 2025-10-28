package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"gotest.tools/assert"
)

type TestStruct struct {
	TestVar1 string `json:"var_one"`
	TestVar2 bool   `json:"var_two"`
}

func TestJsonMarshal(t *testing.T) {
	tmp := TestStruct{
		TestVar1: "test&string 1>0",
		TestVar2: false,
	}
	testdata := []byte(`{"var_one":"test&string 1>0","var_two":false}`)

	adata, err := JSONMarshal(tmp)
	assert.NilError(t, err)

	res := bytes.Compare(adata, testdata)
	assert.Equal(t, res, 1)

	list := []string{}
	adata, err = JSONMarshal(list)
	assert.NilError(t, err)
}

func TestValidateCronDayAndMonth(t *testing.T) {
	err := ValidateCronDayAndMonth("0 0 * * *")
	assert.NilError(t, err)

	err = ValidateCronDayAndMonth("0 0 29 1 *")
	assert.NilError(t, err)

	err = ValidateCronDayAndMonth("0 0 1 0 *")
	assert.NilError(t, err)

	err = ValidateCronDayAndMonth("0 0 31 0 *")
	assert.NilError(t, err)

	err = ValidateCronDayAndMonth("0 0 0 0 *")
	assert.ErrorContains(t, err, "CronExpression has unparseable day or month value:")

	err = ValidateCronDayAndMonth("0 0 32 0 *")
	assert.ErrorContains(t, err, "CronExpression has unparseable day or month value:")

	err = ValidateCronDayAndMonth("0 0 1 12 *")
	assert.ErrorContains(t, err, "CronExpression has unparseable day or month value:")
}

func TestFindEntryInContext(t *testing.T) {
	context := map[string]string{
		"key1":      "value1",
		"KEY2":      "value2",
		"MixedCase": "value3",
	}

	// Exact match
	value, found := FindEntryInContext(context, "key1", true)
	assert.Equal(t, found, true)
	assert.Equal(t, value, "value1")

	// Case-insensitive match (lowercase)
	value, found = FindEntryInContext(context, "KEY1", false)
	assert.Equal(t, found, true)
	assert.Equal(t, value, "value1")

	// Case-insensitive match (uppercase)
	value, found = FindEntryInContext(context, "key2", false)
	assert.Equal(t, found, true)
	assert.Equal(t, value, "value2")

	// Not found
	value, found = FindEntryInContext(context, "nonexistent", false)
	assert.Equal(t, found, false)
	assert.Equal(t, value, "")
}

func TestHelpfulJSONUnmarshalErr(t *testing.T) {
	// Test with JSON syntax error
	invalidJSON := []byte(`{"key": "value"`)
	syntaxErr := &json.SyntaxError{Offset: 10}
	errStr := HelpfulJSONUnmarshalErr(invalidJSON, "TestTag", syntaxErr)
	assert.Assert(t, len(errStr) > 0)
	assert.Assert(t, contains(errStr, "TestTag"))

	// Test with generic error
	validJSON := []byte(`{"key": "value"}`)
	errStr = HelpfulJSONUnmarshalErr(validJSON, "GenericTag", fmt.Errorf("some error"))
	assert.Assert(t, len(errStr) > 0)
	assert.Assert(t, contains(errStr, "GenericTag"))
}

func TestUtcCurrentTimestamp(t *testing.T) {
	ts := UtcCurrentTimestamp()
	assert.Assert(t, !ts.IsZero())
	assert.Equal(t, ts.Location().String(), "UTC")
}

func TestUtcTimeInNano(t *testing.T) {
	nano := UtcTimeInNano()
	assert.Assert(t, nano > 0)
}

func TestUUIDFromTime(t *testing.T) {
	timestamp := int64(1698400000000) // Some timestamp in milliseconds
	node := int64(123456)
	clockSeq := uint32(1000)

	uuid, err := UUIDFromTime(timestamp, node, clockSeq)
	assert.NilError(t, err)
	assert.Assert(t, uuid.String() != "")
	assert.Assert(t, len(uuid.String()) > 0)
}

func TestValidateTimeFormat(t *testing.T) {
	// Valid time format
	err := ValidateTimeFormat("14:30")
	assert.NilError(t, err)

	err = ValidateTimeFormat("00:00")
	assert.NilError(t, err)

	err = ValidateTimeFormat("23:59")
	assert.NilError(t, err)

	// Invalid time format
	err = ValidateTimeFormat("25:00")
	assert.ErrorContains(t, err, "invalid time format")

	err = ValidateTimeFormat("12:60")
	assert.ErrorContains(t, err, "invalid time format")

	err = ValidateTimeFormat("12-30")
	assert.ErrorContains(t, err, "invalid time format")

	err = ValidateTimeFormat("invalid")
	assert.ErrorContains(t, err, "invalid time format")
}

func TestValidateTimezoneList(t *testing.T) {
	// Valid single timezone
	err := ValidateTimezoneList("America/New_York")
	assert.NilError(t, err)

	// Valid multiple timezones
	err = ValidateTimezoneList("America/New_York,Europe/London,Asia/Tokyo")
	assert.NilError(t, err)

	// Valid UTC
	err = ValidateTimezoneList("UTC")
	assert.NilError(t, err)

	// Invalid timezone
	err = ValidateTimezoneList("Invalid/Timezone")
	assert.Assert(t, err != nil)

	// Mixed valid and invalid
	err = ValidateTimezoneList("America/New_York,Invalid/Timezone")
	assert.Assert(t, err != nil)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
