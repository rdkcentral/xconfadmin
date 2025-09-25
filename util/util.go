package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
	copy "github.com/mitchellh/copystructure"
)

var (
	TZ, _ = time.LoadLocation("UTC")
)

// UtcCurrentTimestamp - return current timestamp in UTC timezone
func UtcCurrentTimestamp() time.Time {
	return time.Now().In(TZ)
}

// UtcTimeInNano - return current time in nano in UTC timezone
func UtcTimeInNano() int64 {
	return UtcCurrentTimestamp().UnixNano()
}

// GetTimestamp - return current timestamp in Millisecond in UTC timezone or convert specified time to Millisecond
func GetTimestamp(args ...time.Time) int64 {
	var unixNano int64
	if args == nil {
		unixNano = UtcTimeInNano()
	} else {
		unixNano = args[0].UnixNano()
	}
	return unixNano / int64(time.Millisecond)
}

// UtcOffsetTimestamp currect timestamp
func UtcOffsetTimestamp(sec int) time.Time {
	return UtcCurrentTimestamp().Add(time.Duration(sec) * time.Second)
}

// UtcOffsetPriorminTimestamp currect timestamp
func UtcOffsetPriorMinTimestamp(min int) int64 {
	return UtcCurrentTimestamp().Add(time.Duration(-min)*time.Minute).UnixNano() / int64(time.Millisecond)
}

func Copy(obj interface{}) (interface{}, error) {
	// Create a deep copy of the object
	cloneObj, err := copy.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj, nil
}

// UUIDFromTime gocql method implementation
func UUIDFromTime(timestamp int64, node int64, clockSeq uint32) (gocql.UUID, error) {
	microseconds := int64(time.Duration(timestamp) * time.Microsecond)
	intervals := (microseconds * 10) + 0x01b21dd213814000

	timeLow := intervals & 0xffffffff
	timeMid := (intervals >> 32) & 0xffff
	timeHiVersion := (intervals>>48)&0x0fff + 0x1000

	clockSeqLow := clockSeq & 0xff
	clockSeqHiVariant := 0x80 | ((clockSeq >> 8) & 0x3f)

	/*
		Ref: https://tools.ietf.org/html/rfc4122
		     UUID                   = time-low "-" time-mid "-"
		                             time-high-and-version "-"
		                             clock-seq-and-reserved
		                             clock-seq-low "-" node
		    time-low               = 4hexOctet
		    time-mid               = 2hexOctet
		    time-high-and-version  = 2hexOctet
		    clock-seq-and-reserved = hexOctet
		    clock-seq-low          = hexOctet
		    node                   = 6hexOctet
		  hexOctet               = hexDigit hexDigit
	*/
	uuid := fmt.Sprintf("%08x", int64(timeLow)) + "-" +
		fmt.Sprintf("%04x", int64(timeMid)) + "-" +
		fmt.Sprintf("%04x", int64(timeHiVersion)) + "-" +
		fmt.Sprintf("%02x", int64(clockSeqHiVariant)) +
		fmt.Sprintf("%02x", int64(clockSeqLow)) + "-" +
		fmt.Sprintf("%012x", int64(node))
	return gocql.ParseUUID(uuid)
}

// JSONMarshal is used to marshal struct to string Without escaping special character like &, <, >
// Note: JSONMarshal will terminate each value with a newline
func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func XConfJSONMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.Marshal(v)

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}

// HelpfulJSONUnmarshalErr just points near the location of json err when possible
// jsonStr is the string we tried to unmarshal
// tag is just a string identifying who ran into this err, will be part of the return string
// err is the original err got when unmarshalling
// return a string that contains the offset of the err location
// Copied from xap_proxy
func HelpfulJSONUnmarshalErr(jsonBytes []byte, tag string, err error) string {
	jsonStr := string(jsonBytes)
	var errStr string
	if jsonErr, ok := err.(*json.SyntaxError); ok {
		end := jsonErr.Offset + 10
		if end > int64(len(jsonStr)) {
			end = int64(len(jsonStr))
		}
		begin := jsonErr.Offset - 10
		if begin < 0 {
			begin = 0
		}
		problemPart := jsonStr[begin:end]
		errStr = fmt.Sprintf("Fatal Error in unmarshalling for %s, near <%s> (offset %d) %+v", tag, problemPart, jsonErr.Offset, err)
	} else {
		errStr = fmt.Sprintf("Fatal Error in unmarshalling %s results err: %+v, input str: %s", tag, err, jsonStr)
	}
	return errStr
}

func FindEntryInContext(filterContext map[string]string, key string, exact bool) (value string, found bool) {
	value, found = filterContext[key]
	if !(exact || found) {
		value, found = filterContext[strings.ToLower(key)]
		if !found {
			value = filterContext[strings.ToUpper(key)]
		}
	}
	return value, (value != "")
}

func ValidateCronDayAndMonth(cronExpression string) error {
	if IsBlank(cronExpression) {
		return fmt.Errorf("Cron expression is blank")
	}

	cronFields := strings.Split(cronExpression, " ")
	if len(cronFields) < 4 {
		return errors.New("Cron expression invalid")
	}
	if cronFields[2] == "*" || cronFields[3] == "*" {
		return nil
	}

	// Allowed Values for month is 0-11 and day of month 1-31
	var err error
	var dayOfMonth, month int
	if dayOfMonth, err = strconv.Atoi(cronFields[2]); err != nil {
		return errors.New("Cron expression day of month is invalid")
	}
	if month, err = strconv.Atoi(cronFields[3]); err != nil {
		return errors.New("Cron expression month is invalid")
	}
	if month == 1 && dayOfMonth == 29 {
		return nil
	}

	timeStr := fmt.Sprintf("%d-%d", month+1, dayOfMonth)
	if _, err := time.Parse("1-2", timeStr); err != nil {
		return fmt.Errorf("CronExpression has unparseable day or month value: %s", cronExpression)
	}

	return nil
}

func ValidateTimeFormat(timeStr string) error {
	_, err := time.Parse("15:04", timeStr)
	if err != nil {
		return fmt.Errorf("invalid time format: %s", timeStr)
	}
	return nil
}

func ValidateTimezoneList(timezoneList string) error {
	timezones := strings.Split(timezoneList, ",")
	for _, tz := range timezones {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return err // Return the error if an invalid timezone is found
		}
	}
	return nil // All timezones are valid
}
