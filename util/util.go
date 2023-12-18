/**
 * Copyright 2023 Comcast Cable Communications Management, LLC
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
package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"xconfwebconfig/util"
)

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

// UtcOffsetPriorminTimestamp currect timestamp
func UtcOffsetPriorMinTimestamp(min int) int64 {
	return util.UtcCurrentTimestamp().Add(time.Duration(-min)*time.Minute).UnixNano() / int64(time.Millisecond)
}

func ValidateCronDayAndMonth(cronExpression string) error {
	if util.IsBlank(cronExpression) {
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
