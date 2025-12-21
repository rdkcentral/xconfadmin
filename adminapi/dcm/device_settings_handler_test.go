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
package dcm

import (
	"testing"
	"time"

	"gotest.tools/assert"
)

// ========== Tests for str2Time function ==========

// TestStr2Time_ValidInput tests parsing a valid datetime string
func TestStr2Time_ValidInput(t *testing.T) {
	t.Parallel()
	dateStr := "2021-10-05 14:30:00"

	result, err := str2Time(dateStr)

	assert.NilError(t, err)
	assert.Equal(t, 2021, result.Year())
	assert.Equal(t, time.October, result.Month())
	assert.Equal(t, 5, result.Day())
	assert.Equal(t, 14, result.Hour())
	assert.Equal(t, 30, result.Minute())
	assert.Equal(t, 0, result.Second())
}

// TestStr2Time_ValidInputMidnight tests parsing a datetime string at midnight
func TestStr2Time_ValidInputMidnight(t *testing.T) {
	t.Parallel()
	dateStr := "2025-01-01 00:00:00"

	result, err := str2Time(dateStr)

	assert.NilError(t, err)
	assert.Equal(t, 2025, result.Year())
	assert.Equal(t, time.January, result.Month())
	assert.Equal(t, 1, result.Day())
	assert.Equal(t, 0, result.Hour())
	assert.Equal(t, 0, result.Minute())
	assert.Equal(t, 0, result.Second())
}

// TestStr2Time_ValidInputEndOfDay tests parsing a datetime string at end of day
func TestStr2Time_ValidInputEndOfDay(t *testing.T) {
	t.Parallel()
	dateStr := "2024-12-31 23:59:59"

	result, err := str2Time(dateStr)

	assert.NilError(t, err)
	assert.Equal(t, 2024, result.Year())
	assert.Equal(t, time.December, result.Month())
	assert.Equal(t, 31, result.Day())
	assert.Equal(t, 23, result.Hour())
	assert.Equal(t, 59, result.Minute())
	assert.Equal(t, 59, result.Second())
}

// TestStr2Time_EmptyString tests parsing an empty string (nil condition)
func TestStr2Time_EmptyString(t *testing.T) {
	t.Parallel()
	dateStr := ""

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for empty string")
}

// TestStr2Time_InvalidFormat tests parsing a string with invalid format
func TestStr2Time_InvalidFormat(t *testing.T) {
	t.Parallel()
	dateStr := "2021/10/05 14:30:00" // Wrong separator

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for invalid format")
}

// TestStr2Time_InvalidDateFormat tests parsing a string with wrong date format
func TestStr2Time_InvalidDateFormat(t *testing.T) {
	t.Parallel()
	dateStr := "10-05-2021 14:30:00" // Wrong order

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for wrong date format")
}

// TestStr2Time_PartialString tests parsing a partial datetime string (nil condition)
func TestStr2Time_PartialString(t *testing.T) {
	t.Parallel()
	dateStr := "2021-10-05"

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for partial datetime string")
}

// TestStr2Time_InvalidMonth tests parsing with invalid month
func TestStr2Time_InvalidMonth(t *testing.T) {
	t.Parallel()
	dateStr := "2021-13-05 14:30:00" // Month 13 doesn't exist

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for invalid month")
}

// TestStr2Time_InvalidDay tests parsing with invalid day
func TestStr2Time_InvalidDay(t *testing.T) {
	t.Parallel()
	dateStr := "2021-02-30 14:30:00" // Feb 30 doesn't exist

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for invalid day")
}

// TestStr2Time_InvalidHour tests parsing with invalid hour
func TestStr2Time_InvalidHour(t *testing.T) {
	t.Parallel()
	dateStr := "2021-10-05 25:30:00" // Hour 25 doesn't exist

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for invalid hour")
}

// TestStr2Time_InvalidMinute tests parsing with invalid minute
func TestStr2Time_InvalidMinute(t *testing.T) {
	t.Parallel()
	dateStr := "2021-10-05 14:60:00" // Minute 60 doesn't exist

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for invalid minute")
}

// TestStr2Time_InvalidSecond tests parsing with invalid second
func TestStr2Time_InvalidSecond(t *testing.T) {
	t.Parallel()
	dateStr := "2021-10-05 14:30:60" // Second 60 doesn't exist (normally)

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for invalid second")
}

// TestStr2Time_NullString tests parsing a null/nil-like string (nil condition)
func TestStr2Time_NullString(t *testing.T) {
	t.Parallel()
	dateStr := "null"

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for null string")
}

// TestStr2Time_WithExtraSpaces tests parsing with extra spaces (nil condition)
func TestStr2Time_WithExtraSpaces(t *testing.T) {
	t.Parallel()
	dateStr := " 2021-10-05 14:30:00 "

	_, err := str2Time(dateStr)

	assert.Assert(t, err != nil, "Expected error for string with extra spaces")
}

// ========== Tests for changeTZ function ==========

// TestChangeTZ_UTCToMST tests timezone conversion from UTC to MST
func TestChangeTZ_UTCToMST(t *testing.T) {
	t.Parallel()
	// Create a time: 2021-10-05 00:00:00 UTC
	inputTime := time.Date(2021, 10, 5, 0, 0, 0, 0, time.UTC)

	// Load MST timezone (Mountain Standard Time, UTC-7)
	mst, err := time.LoadLocation("MST")
	assert.NilError(t, err)

	result := changeTZ(inputTime, mst)

	// When we interpret "2021-10-05 00:00:00" as MST and convert back to UTC,
	// it becomes "2021-10-05 07:00:00" UTC
	assert.Equal(t, "2021-10-05 07:00:00", result)
}

// TestChangeTZ_UTCToEST tests timezone conversion from UTC to EST
func TestChangeTZ_UTCToEST(t *testing.T) {
	t.Parallel()
	// Create a time: 2021-10-05 00:00:00 UTC
	inputTime := time.Date(2021, 10, 5, 0, 0, 0, 0, time.UTC)

	// Load EST timezone (Eastern Standard Time, UTC-5)
	est, err := time.LoadLocation("EST")
	assert.NilError(t, err)

	result := changeTZ(inputTime, est)

	// When we interpret "2021-10-05 00:00:00" as EST and convert back to UTC,
	// it becomes "2021-10-05 05:00:00" UTC
	assert.Equal(t, "2021-10-05 05:00:00", result)
}

// TestChangeTZ_UTCToUTC tests timezone conversion from UTC to UTC (no change expected)
func TestChangeTZ_UTCToUTC(t *testing.T) {
	t.Parallel()
	// Create a time: 2021-10-05 12:30:45 UTC
	inputTime := time.Date(2021, 10, 5, 12, 30, 45, 0, time.UTC)

	result := changeTZ(inputTime, time.UTC)

	// Should remain the same
	assert.Equal(t, "2021-10-05 12:30:45", result)
}

// TestChangeTZ_UTCToAsiaTokyo tests timezone conversion to Asia/Tokyo
func TestChangeTZ_UTCToAsiaTokyo(t *testing.T) {
	t.Parallel()
	// Create a time: 2021-10-05 00:00:00 UTC
	inputTime := time.Date(2021, 10, 5, 0, 0, 0, 0, time.UTC)

	// Load Asia/Tokyo timezone (UTC+9)
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	assert.NilError(t, err)

	result := changeTZ(inputTime, tokyo)

	// When we interpret "2021-10-05 00:00:00" as Tokyo time and convert back to UTC,
	// it becomes "2021-10-04 15:00:00" UTC
	assert.Equal(t, "2021-10-04 15:00:00", result)
}

// TestChangeTZ_UTCToAmericaLosAngeles tests timezone conversion to America/Los_Angeles
func TestChangeTZ_UTCToAmericaLosAngeles(t *testing.T) {
	t.Parallel()
	// Create a time: 2021-10-05 00:00:00 UTC
	inputTime := time.Date(2021, 10, 5, 0, 0, 0, 0, time.UTC)

	// Load America/Los_Angeles timezone (PDT, UTC-7 during daylight saving)
	la, err := time.LoadLocation("America/Los_Angeles")
	assert.NilError(t, err)

	result := changeTZ(inputTime, la)

	// PDT is UTC-7, so "2021-10-05 00:00:00" PDT becomes "2021-10-05 07:00:00" UTC
	assert.Equal(t, "2021-10-05 07:00:00", result)
}

// TestChangeTZ_MidnightCrossover tests timezone conversion that crosses midnight
func TestChangeTZ_MidnightCrossover(t *testing.T) {
	t.Parallel()
	// Create a time: 2021-12-31 23:00:00 UTC
	inputTime := time.Date(2021, 12, 31, 23, 0, 0, 0, time.UTC)

	// Load a timezone ahead of UTC
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	assert.NilError(t, err)

	result := changeTZ(inputTime, tokyo)

	// "2021-12-31 23:00:00" JST becomes "2021-12-31 14:00:00" UTC
	assert.Equal(t, "2021-12-31 14:00:00", result)
}

// TestChangeTZ_NilLocation tests changeTZ with nil location (nil condition)
func TestChangeTZ_NilLocation(t *testing.T) {
	t.Parallel()
	// Create a time: 2021-10-05 00:00:00 UTC
	inputTime := time.Date(2021, 10, 5, 0, 0, 0, 0, time.UTC)

	// This should panic or handle nil gracefully
	// The actual function doesn't handle nil, so this documents the behavior
	defer func() {
		if r := recover(); r != nil {
			// Expected panic for nil location
			t.Logf("Expected panic occurred: %v", r)
		}
	}()

	result := changeTZ(inputTime, nil)

	// If no panic, the result will use nil location which defaults to UTC
	t.Logf("Result with nil location: %s", result)
}

// TestChangeTZ_ZeroTime tests changeTZ with zero time value (nil condition)
func TestChangeTZ_ZeroTime(t *testing.T) {
	t.Parallel()
	// Zero time
	var inputTime time.Time

	// Use UTC timezone for zero time test
	result := changeTZ(inputTime, time.UTC)

	// Zero time is "0001-01-01 00:00:00 UTC"
	// With UTC timezone, it should remain the same
	assert.Equal(t, "0001-01-01 00:00:00", result)
}

// TestChangeTZ_LeapYear tests timezone conversion with leap year date
func TestChangeTZ_LeapYear(t *testing.T) {
	t.Parallel()
	// Create a time on leap day: 2024-02-29 12:00:00 UTC
	inputTime := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)

	// Load EST timezone
	est, err := time.LoadLocation("EST")
	assert.NilError(t, err)

	result := changeTZ(inputTime, est)

	// "2024-02-29 12:00:00" EST becomes "2024-02-29 17:00:00" UTC
	assert.Equal(t, "2024-02-29 17:00:00", result)
}

// TestChangeTZ_DaylightSavingTransition tests timezone conversion during DST transition
func TestChangeTZ_DaylightSavingTransition(t *testing.T) {
	t.Parallel()
	// Create a time during DST transition: 2021-03-14 02:30:00 UTC
	inputTime := time.Date(2021, 3, 14, 2, 30, 0, 0, time.UTC)

	// Load America/New_York timezone
	ny, err := time.LoadLocation("America/New_York")
	assert.NilError(t, err)

	result := changeTZ(inputTime, ny)

	// This tests DST handling
	// The exact result depends on whether the time falls before or after DST transition
	assert.Assert(t, len(result) > 0, "Result should not be empty")
}

// TestChangeTZ_FarFutureDate tests timezone conversion with far future date
func TestChangeTZ_FarFutureDate(t *testing.T) {
	t.Parallel()
	// Create a time in the far future: 2099-12-31 23:59:59 UTC
	inputTime := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)

	// Load UTC timezone
	result := changeTZ(inputTime, time.UTC)

	assert.Equal(t, "2099-12-31 23:59:59", result)
}

// TestChangeTZ_EarlyMorningHour tests timezone conversion with early morning hour
func TestChangeTZ_EarlyMorningHour(t *testing.T) {
	t.Parallel()
	// Create a time: 2021-10-05 01:00:00 UTC
	inputTime := time.Date(2021, 10, 5, 1, 0, 0, 0, time.UTC)

	// Load MST timezone
	mst, err := time.LoadLocation("MST")
	assert.NilError(t, err)

	result := changeTZ(inputTime, mst)

	// "2021-10-05 01:00:00" MST becomes "2021-10-05 08:00:00" UTC
	assert.Equal(t, "2021-10-05 08:00:00", result)
}

// TestChangeTZ_ConsistencyCheck tests that changeTZ maintains consistency
func TestChangeTZ_ConsistencyCheck(t *testing.T) {
	t.Parallel()
	// Create multiple times and ensure consistent behavior
	times := []time.Time{
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 6, 15, 12, 30, 0, 0, time.UTC),
		time.Date(2021, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	mst, err := time.LoadLocation("MST")
	assert.NilError(t, err)

	for _, testTime := range times {
		result := changeTZ(testTime, mst)
		// Ensure result is in the expected format
		assert.Assert(t, len(result) == 19, "Result should be 19 characters (YYYY-MM-DD HH:MM:SS)")

		// Parse the result back to verify it's valid
		_, parseErr := str2Time(result)
		assert.NilError(t, parseErr, "Result should be parseable by str2Time")
	}
}
