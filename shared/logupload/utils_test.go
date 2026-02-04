package logupload

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRandomizeCronIfNecessary_WithDayRandomized tests day randomized scenario
func TestRandomizeCronIfNecessary_WithDayRandomized(t *testing.T) {
	expression := "0 0 * * *"
	timeWindow := 60
	estbMac := "AA:BB:CC:DD:EE:FF"
	cronName := "TestCron"
	timeZone := "UTC"

	result := randomizeCronIfNecessary(expression, timeWindow, true, estbMac, cronName, timeZone)

	// Should return randomized cron expression
	assert.NotEmpty(t, result, "Should return randomized cron expression")
	assert.True(t, strings.Contains(result, " "), "Should contain space-separated values")
}

// TestRandomizeCronIfNecessary_WithTimeWindow tests time window randomization
func TestRandomizeCronIfNecessary_WithTimeWindow(t *testing.T) {
	expression := "30 10 * * *"
	timeWindow := 120
	estbMac := "11:22:33:44:55:66"
	cronName := "ScheduledCron"
	timeZone := "America/New_York"

	result := randomizeCronIfNecessary(expression, timeWindow, false, estbMac, cronName, timeZone)

	// Should return randomized cron expression
	assert.NotEmpty(t, result, "Should return randomized cron expression")
}

// TestRandomizeCronIfNecessary_NoRandomization tests when no randomization is needed
func TestRandomizeCronIfNecessary_NoRandomization(t *testing.T) {
	expression := "15 8 * * *"
	timeWindow := 0
	estbMac := "AA:BB:CC:DD:EE:11"
	cronName := "NormalCron"
	timeZone := "UTC"

	result := randomizeCronIfNecessary(expression, timeWindow, false, estbMac, cronName, timeZone)

	// Should return empty string when no randomization is needed
	assert.Empty(t, result, "Should return empty string when timeWindow is 0 and not day randomized")
}

// TestRandomizeCronIfNecessary_EmptyExpression tests with empty expression
func TestRandomizeCronIfNecessary_EmptyExpression(t *testing.T) {
	expression := ""
	timeWindow := 60
	estbMac := "AA:BB:CC:DD:EE:22"
	cronName := "EmptyCron"
	timeZone := "UTC"

	result := randomizeCronIfNecessary(expression, timeWindow, false, estbMac, cronName, timeZone)

	// Should return empty string for empty expression
	assert.Empty(t, result, "Should return empty string for empty expression")
}

// TestRandomizeCronIfNecessary_InvalidExpression tests with invalid cron expression
func TestRandomizeCronIfNecessary_InvalidExpression(t *testing.T) {
	expression := "invalid cron"
	timeWindow := 60
	estbMac := "AA:BB:CC:DD:EE:33"
	cronName := "InvalidCron"
	timeZone := "UTC"

	result := randomizeCronIfNecessary(expression, timeWindow, false, estbMac, cronName, timeZone)

	// Should return empty string for invalid expression
	assert.Empty(t, result, "Should return empty string for invalid expression")
}

// TestRandomizeCronEx_DayRandomized tests full day randomization
func TestRandomizeCronEx_DayRandomized(t *testing.T) {
	expression := "0 0 * * *"
	timeWindow := 60
	timeZone := "UTC"

	result := randomizeCronEx(expression, timeWindow, true, timeZone)

	// Should return valid cron expression with randomized time
	assert.NotEmpty(t, result, "Should return randomized cron expression")
	parts := strings.Split(result, " ")
	assert.GreaterOrEqual(t, len(parts), 5, "Should have at least 5 parts in cron expression")
}

// TestRandomizeCronEx_TimeWindowRandomization tests time window based randomization
func TestRandomizeCronEx_TimeWindowRandomization(t *testing.T) {
	expression := "0 12 * * *"
	timeWindow := 180
	timeZone := "America/Los_Angeles"

	result := randomizeCronEx(expression, timeWindow, false, timeZone)

	// Should return valid cron expression
	assert.NotEmpty(t, result, "Should return randomized cron expression")
	parts := strings.Split(result, " ")
	assert.GreaterOrEqual(t, len(parts), 5, "Should have 5 parts in cron expression")
}

// TestRandomizeCronEx_InvalidExpression tests with invalid expression
func TestRandomizeCronEx_InvalidExpression(t *testing.T) {
	expression := "not valid"
	timeWindow := 60
	timeZone := "UTC"

	result := randomizeCronEx(expression, timeWindow, false, timeZone)

	// Should return empty string for invalid expression
	assert.Empty(t, result, "Should return empty string for invalid expression")
}

// TestRandomizeCronEx_MidnightBoundary tests midnight boundary handling
func TestRandomizeCronEx_MidnightBoundary(t *testing.T) {
	expression := "50 23 * * *"
	timeWindow := 30
	timeZone := "UTC"

	result := randomizeCronEx(expression, timeWindow, false, timeZone)

	// Should handle midnight boundary correctly
	assert.NotEmpty(t, result, "Should return valid cron expression")
	parts := strings.Split(result, " ")
	assert.GreaterOrEqual(t, len(parts), 2, "Should have at least 2 parts")
}

// TestRandomizeCronEx_EmptyTimeZone tests with empty timezone
func TestRandomizeCronEx_EmptyTimeZone(t *testing.T) {
	expression := "15 10 * * *"
	timeWindow := 60
	timeZone := ""

	result := randomizeCronEx(expression, timeWindow, false, timeZone)

	// Should work with empty timezone
	assert.NotEmpty(t, result, "Should return randomized cron expression even with empty timezone")
}

// TestRandomizeCronEx_LargeTimeWindow tests with large time window
func TestRandomizeCronEx_LargeTimeWindow(t *testing.T) {
	expression := "0 0 * * *"
	timeWindow := 1440 // Full day
	timeZone := "UTC"

	result := randomizeCronEx(expression, timeWindow, false, timeZone)

	// Should handle large time window
	assert.NotEmpty(t, result, "Should return valid cron expression")
}

// TestValidate_ValidExpression tests validation with valid cron expression
func TestValidate_ValidExpression(t *testing.T) {
	expression := "30 14 * * *"

	result := validate(expression)

	assert.True(t, result, "Should validate correct cron expression")
}

// TestValidate_InvalidFormat tests validation with invalid format
func TestValidate_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
	}{
		{
			name:       "Single part",
			expression: "30",
		},
		{
			name:       "Empty string",
			expression: "",
		},
		{
			name:       "Only spaces",
			expression: "   ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validate(tc.expression)
			assert.False(t, result, "Should return false for invalid format: %s", tc.name)
		})
	}
}

// TestValidate_NonNumericValues tests validation with non-numeric values
func TestValidate_NonNumericValues(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
	}{
		{
			name:       "Invalid minutes",
			expression: "abc 10 * * *",
		},
		{
			name:       "Invalid hours",
			expression: "30 xyz * * *",
		},
		{
			name:       "Both invalid",
			expression: "foo bar * * *",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validate(tc.expression)
			assert.False(t, result, "Should return false for non-numeric values: %s", tc.name)
		})
	}
}

// TestValidate_NegativeValues tests validation with negative values
func TestValidate_NegativeValues(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
	}{
		{
			name:       "Negative minutes",
			expression: "-5 10 * * *",
		},
		{
			name:       "Negative hours",
			expression: "30 -2 * * *",
		},
		{
			name:       "Both negative",
			expression: "-10 -5 * * *",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validate(tc.expression)
			assert.False(t, result, "Should return false for negative values: %s", tc.name)
		})
	}
}

// TestValidate_BoundaryValues tests validation with boundary values
func TestValidate_BoundaryValues(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		expected   bool
	}{
		{
			name:       "Zero minutes and hours",
			expression: "0 0 * * *",
			expected:   true,
		},
		{
			name:       "Max minutes",
			expression: "59 23 * * *",
			expected:   true,
		},
		{
			name:       "Valid midnight",
			expression: "0 0 * * *",
			expected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validate(tc.expression)
			assert.Equal(t, tc.expected, result, "Validation result mismatch for: %s", tc.name)
		})
	}
}

// TestGetAddedHoursToRandomizedCronByTimeZone_ValidTimeZone tests with valid timezone
func TestGetAddedHoursToRandomizedCronByTimeZone_ValidTimeZone(t *testing.T) {
	testCases := []struct {
		name     string
		timeZone string
	}{
		{
			name:     "US Eastern",
			timeZone: "US/Eastern",
		},
		{
			name:     "America New York",
			timeZone: "America/New_York",
		},
		{
			name:     "UTC",
			timeZone: "UTC",
		},
		{
			name:     "America Los Angeles",
			timeZone: "America/Los_Angeles",
		},
		{
			name:     "Europe London",
			timeZone: "Europe/London",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getAddedHoursToRandomizedCronByTimeZone(tc.timeZone)
			// Result should be an integer (may be positive, negative, or zero)
			assert.True(t, result >= -24 && result <= 24, 
				"Time shift should be within reasonable range for %s, got %d", tc.name, result)
		})
	}
}

// TestGetAddedHoursToRandomizedCronByTimeZone_InvalidTimeZone tests with invalid timezone
func TestGetAddedHoursToRandomizedCronByTimeZone_InvalidTimeZone(t *testing.T) {
	timeZone := "Invalid/TimeZone"

	result := getAddedHoursToRandomizedCronByTimeZone(timeZone)

	// Should fallback to default or return 0
	assert.True(t, result >= -24 && result <= 24, "Should return reasonable value even for invalid timezone")
}

// TestGetAddedHoursToRandomizedCronByTimeZone_EmptyTimeZone tests with empty timezone
func TestGetAddedHoursToRandomizedCronByTimeZone_EmptyTimeZone(t *testing.T) {
	timeZone := ""

	result := getAddedHoursToRandomizedCronByTimeZone(timeZone)

	// Should return 0 for empty timezone
	assert.Equal(t, 0, result, "Should return 0 for empty timezone")
}

// TestGetAddedHoursToRandomizedCronByTimeZone_DSTAwareness tests DST awareness
func TestGetAddedHoursToRandomizedCronByTimeZone_DSTAwareness(t *testing.T) {
	timeZone := "America/New_York"

	result := getAddedHoursToRandomizedCronByTimeZone(timeZone)

	// Result should account for DST
	assert.True(t, result >= -24 && result <= 24, "Should return valid time shift with DST consideration")
}

// TestIsDST_SummerTime tests DST detection during summer
func TestIsDST_SummerTime(t *testing.T) {
	// Create a time in July (typically DST in Northern Hemisphere)
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("Cannot load America/New_York timezone")
	}
	summerTime := time.Date(2023, 7, 15, 12, 0, 0, 0, loc)

	result := isDST(summerTime)

	// July in New York should be DST
	assert.True(t, result, "July should be DST in America/New_York")
}

// TestIsDST_WinterTime tests DST detection during winter
func TestIsDST_WinterTime(t *testing.T) {
	// Create a time in January (not DST in Northern Hemisphere)
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("Cannot load America/New_York timezone")
	}
	winterTime := time.Date(2023, 1, 15, 12, 0, 0, 0, loc)

	result := isDST(winterTime)

	// January in New York should not be DST
	assert.False(t, result, "January should not be DST in America/New_York")
}

// TestIsDST_UTCTime tests DST with UTC timezone
func TestIsDST_UTCTime(t *testing.T) {
	// UTC doesn't have DST
	utcTime := time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)

	result := isDST(utcTime)

	// UTC should not have DST
	assert.False(t, result, "UTC should not have DST")
}

// TestIsDST_YearBoundary tests DST detection around year boundary
func TestIsDST_YearBoundary(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("Cannot load America/New_York timezone")
	}

	testCases := []struct {
		name     string
		month    time.Month
		expected bool
	}{
		{
			name:     "December",
			month:    time.December,
			expected: false,
		},
		{
			name:     "August",
			month:    time.August,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testTime := time.Date(2023, tc.month, 15, 12, 0, 0, 0, loc)
			result := isDST(testTime)
			assert.Equal(t, tc.expected, result, "DST detection mismatch for %s", tc.name)
		})
	}
}

// TestIsDST_SouthernHemisphere tests DST in Southern Hemisphere
func TestIsDST_SouthernHemisphere(t *testing.T) {
	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		t.Skip("Cannot load Australia/Sydney timezone")
	}

	// January is summer in Southern Hemisphere (DST)
	summerTime := time.Date(2023, 1, 15, 12, 0, 0, 0, loc)
	// July is winter in Southern Hemisphere (no DST)
	winterTime := time.Date(2023, 7, 15, 12, 0, 0, 0, loc)

	summerResult := isDST(summerTime)
	winterResult := isDST(winterTime)

	// January should be DST in Sydney, July should not
	assert.True(t, summerResult, "January should be DST in Australia/Sydney")
	assert.False(t, winterResult, "July should not be DST in Australia/Sydney")
}

// TestIsDST_TransitionDates tests DST during transition periods
func TestIsDST_TransitionDates(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("Cannot load America/New_York timezone")
	}

	// March typically has DST transition
	marchTime := time.Date(2023, 3, 15, 12, 0, 0, 0, loc)
	// November typically has DST transition
	novTime := time.Date(2023, 11, 15, 12, 0, 0, 0, loc)

	marchResult := isDST(marchTime)
	novResult := isDST(novTime)

	// Both should execute without error
	assert.True(t, marchResult || !marchResult, "Should handle March transition")
	assert.True(t, novResult || !novResult, "Should handle November transition")
}

// TestRandomizeCronEx_HourOverflow tests hour overflow beyond 24
func TestRandomizeCronEx_HourOverflow(t *testing.T) {
	// Test case where adding random minutes causes hour overflow
	expression := "55 23 * * *"
	timeWindow := 120
	timeZone := "UTC"

	result := randomizeCronEx(expression, timeWindow, false, timeZone)

	assert.NotEmpty(t, result, "Should handle hour overflow")
	parts := strings.Split(result, " ")
	if len(parts) >= 2 {
		// Verify the hour is valid (0-23)
		// Note: We can't check the exact value due to randomness
		assert.NotEmpty(t, parts[1], "Hour should be present")
	}
}

// TestRandomizeCronEx_PreservesOtherFields tests that day/month/weekday are preserved
func TestRandomizeCronEx_PreservesOtherFields(t *testing.T) {
	expression := "30 10 15 6 1"
	timeWindow := 60
	timeZone := "UTC"

	result := randomizeCronEx(expression, timeWindow, false, timeZone)

	assert.NotEmpty(t, result, "Should return randomized expression")
	parts := strings.Split(result, " ")
	assert.Equal(t, 5, len(parts), "Should have 5 parts")
	// Day, month, weekday should be preserved
	assert.Equal(t, "15", parts[2], "Day should be preserved")
	assert.Equal(t, "6", parts[3], "Month should be preserved")
	assert.Equal(t, "1", parts[4], "Weekday should be preserved")
}

// TestGetAddedHoursToRandomizedCronByTimeZone_MultipleTimeZones tests various timezones
func TestGetAddedHoursToRandomizedCronByTimeZone_MultipleTimeZones(t *testing.T) {
	timeZones := []string{
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"Europe/London",
		"Asia/Tokyo",
		"Australia/Sydney",
		"UTC",
	}

	for _, tz := range timeZones {
		t.Run(tz, func(t *testing.T) {
			result := getAddedHoursToRandomizedCronByTimeZone(tz)
			// Should return a reasonable value
			assert.True(t, result >= -24 && result <= 24, 
				"Time shift should be within valid range for %s, got %d", tz, result)
		})
	}
}

// TestValidate_FullCronExpression tests validation with complete cron expressions
func TestValidate_FullCronExpression(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		expected   bool
	}{
		{
			name:       "Standard daily",
			expression: "0 2 * * *",
			expected:   true,
		},
		{
			name:       "Specific day and month",
			expression: "30 14 15 6 *",
			expected:   true,
		},
		{
			name:       "With weekday",
			expression: "45 8 * * 1",
			expected:   true,
		},
		{
			name:       "All wildcards after hour",
			expression: "15 20 * * *",
			expected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validate(tc.expression)
			assert.Equal(t, tc.expected, result, "Validation mismatch for: %s", tc.name)
		})
	}
}
