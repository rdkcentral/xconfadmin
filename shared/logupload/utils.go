package logupload

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	WHOLE_DAY_RANDOMIZED        = "Whole Day Randomized"
	UTC                  string = "UTC"
	LOCAL_TIME           string = "Local time"
)

func randomizeCronIfNecessary(expression string, timeWindow int, isDayRandomized bool, estbMac string, cronName string, timeZone string) string {
	var randomCronExp = ""
	if isDayRandomized || (len(expression) > 0 && timeWindow > 0) {
		randomCronExp = randomizeCronEx(expression, timeWindow, isDayRandomized, timeZone)
		if len(randomCronExp) < 1 {
			//log.Error("Invalid %s=%s for estbMac=%s", cronName, expression, estbMac)
			log.Error("Invalid {}={} for estbMac={}", cronName, expression, estbMac)
		} else {
			currentTime := time.Now().Format("2021-03-23 10:11:12")
			log.Debugf("SettingsUtil original {%s}={%s} randomized {%s}={%s} for estbMac={%s} at dcmTime={%s}", cronName, expression, cronName, randomCronExp, estbMac, currentTime)
		}
	}
	return randomCronExp
}

/**
 * Randomize the cron expression between the cron expression and upper bound as timeWindow.
 * Also depending on type random range is fixed.
 * @param expression   cron expression.
 * @param timeWindow   upper bound.
 * @param isDayRandomized DayRandomized/cron expression.
 * @return  String randomized cron expression.
 */
func randomizeCronEx(expression string, timeWindow int, isDayRandomized bool, timeZone string) string {
	expressionArray := []string{"0", "0", "*", "*", "*"}
	var lowerMinutes int = 0
	var lowerHour int = 0
	var randomNumber int
	if isDayRandomized {
		randomNumber = rand.Intn(1440)
	} else {
		if !validate(expression) {
			return ""
		}
		expressionArray = strings.Split(expression, " ")
		lowerMinutes, _ = strconv.Atoi(expressionArray[0])
		lowerHour, _ = strconv.Atoi(expressionArray[1])
		randomNumber = rand.Intn(timeWindow)
	}
	//Get next random hour and random minute
	newMin := lowerMinutes + randomNumber
	//To tackle midnight boundaries.
	//If Minutes>= 60 extract out hour and add it to new hour value.
	//Being division and mod operators it will take care of while conditions,and remainder value shall
	//always be less than 60 for minutes and less than 24 or 0 for hours.
	newHr := newMin / 60
	newMin = newMin % 60
	//If new hour value is >=24 i.e.  at 00 am or more then convert to AM values i.e. 0,1 etc
	newHr = lowerHour + newHr
	newHr = newHr + getAddedHoursToRandomizedCronByTimeZone(timeZone)
	newHr = newHr % 24
	//As per ticket only hour and day need to be considered.
	var sb strings.Builder
	sb.WriteString(strconv.Itoa(newMin) + " " + strconv.Itoa(newHr))
	for i := 2; i < len(expressionArray); i++ {
		sb.WriteString(" " + expressionArray[i])
	}
	return sb.String()
}

/**
 * Validates hours and minutes section of  the cron expression.Ideally at the time of entering these details by the user it should be validated.
 *
 * @param expression  Cron expression.
 * @return  boolean for validation.
 */
func validate(expression string) bool {
	split := strings.Split(expression, " ")
	if len(split) < 2 {
		return false
	}
	minutes, err := strconv.Atoi(split[0])
	if err != nil {
		log.Error("Invalid cron expression:" + expression)
		return false
	}
	hour, err := strconv.Atoi(split[1])
	if err != nil {
		log.Error("Invalid cron expression:" + expression)
		return false
	}
	if minutes < 0 || hour < 0 {
		return false
	}
	return true
}

const (
	DEFAULT_TIME_ZONE  = "US/Eastern"
	ONE_HOUR_SECONDS   = 3600
	DEFAULT_OFFSET_ROW = -5
)

func getAddedHoursToRandomizedCronByTimeZone(timeZoneStr string) int {
	if len(timeZoneStr) < 1 {
		return 0
	}
	loc, err := time.LoadLocation(timeZoneStr)
	if err != nil {
		log.Errorf("unknown time zone(%s): %v", timeZoneStr, err)
		loc, err = time.LoadLocation(DEFAULT_TIME_ZONE)
		if err != nil {
			return 0
		}
	}
	log.Debugf("success find time zone: %s", timeZoneStr)

	now := time.Now().In(loc)
	_, offset := now.Zone() // offset in seconds east of UTC for specified TZ
	timeShift := DEFAULT_OFFSET_ROW - offset/ONE_HOUR_SECONDS

	if isDST(now) {
		// Get the raw value that is not affected by daylight saving time
		timeShift++
	}

	log.Debug("SettingsUtil incomingTimeZone=" + timeZoneStr + " matchedTimeZone=" + loc.String() + " timeShift=" + strconv.Itoa(timeShift))
	return timeShift
}

// IsDST returns true if the time given is in DST, false if not
// DST is defined as when the offset from UTC is increased
func isDST(t time.Time) bool {
	// t
	_, tOffset := t.Zone()

	// January 1
	janYear := t.Year()
	if t.Month() > 6 {
		janYear = janYear + 1
	}
	jan1Location := time.Date(janYear, 1, 1, 0, 0, 0, 0, t.Location())
	_, janOffset := jan1Location.Zone()

	// July 1
	jul1Location := time.Date(t.Year(), 7, 1, 0, 0, 0, 0, t.Location())
	_, julOffset := jul1Location.Zone()

	if tOffset == janOffset {
		return janOffset > julOffset
	}
	return julOffset > janOffset
}
