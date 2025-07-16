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
package queries

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"xconfadmin/adminapi/auth"
	"xconfadmin/common"
	xhttp "xconfadmin/http"
	"xconfadmin/shared/logupload"
	"xconfadmin/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
)

// This function is not being referenced in router.go. Should we delete it?
func SaveLogUploadSettings(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.COMMON_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	logUploadSettings := logupload.LogUploadSettings{}
	err = json.Unmarshal([]byte(body), &logUploadSettings)
	if err != nil {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	timezone, found := mux.Vars(r)[common.TIME_ZONE]
	if !found || len(strings.TrimSpace(timezone)) == 0 {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("timezone is blank"))
		return
	}
	scheduleTimezone, found := mux.Vars(r)[common.SCHEDULE_TIME_ZONE]
	if !found || len(strings.TrimSpace(timezone)) == 0 {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("timezone is blank"))
		return
	}
	schedule := logUploadSettings.Schedule
	if schedule == (logupload.Schedule{}) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Schedule is empty")
		return
	}
	if checkDateStrLength(schedule.StartDate) && checkDateStrLength(schedule.EndDate) {
		startValid := isValidDate(schedule.StartDate)
		if !startValid {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Start date is invalid")
			return
		}
		endValid := isValidDate(schedule.EndDate)
		if !endValid {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "End date is invalid")
			return
		}
		if startValid && endValid && !isEnddateAfterStartDate(schedule.StartDate, schedule.EndDate) {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Start date is greater/equal to End date")
			return
		}
	}
	if logUploadSettings.FromDateTime == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Start date is blank")
		return
	}
	if logUploadSettings.ToDateTime == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "End date is blank")
		return
	}
	if checkDateStrLength(logUploadSettings.FromDateTime) && checkDateStrLength(logUploadSettings.ToDateTime) {
		startValid := isValidDate(logUploadSettings.FromDateTime)
		if !startValid {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Start date is invalid")
			return
		}
		endValid := isValidDate(logUploadSettings.ToDateTime)
		if !startValid {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "End date is invalid")
			return
		}
		if startValid && endValid && !isEnddateAfterStartDate(logUploadSettings.FromDateTime, logUploadSettings.ToDateTime) {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Start date is greater/equal to End date")
			return
		}
	}
	if logUploadSettings.ModeToGetLogFiles == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "File mode is empty")
		return
	}
	if len(logUploadSettings.LogFileIds) < 1 && logUploadSettings.ModeToGetLogFiles == logupload.MODE_TO_GET_LOG_FILES_0 {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "At least log file should be specified")
		return
	}
	nameErrorMessage := validateName(&logUploadSettings)
	if nameErrorMessage != "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, nameErrorMessage)
		return
	}
	/*
		twoLogFileList := logupload.GetAllLogFileList(2)
		response, err := util.JSONMarshal(twoLogFileList)
		if err != nil {
			log.Error(fmt.Sprintf("json.Marshal featureRuleNew error: %v", err))
		}
		xhttp.WriteXconfResponse(w, http.StatusCreated, response)
		return */

	if logUploadSettings.ModeToGetLogFiles == logupload.MODE_TO_GET_LOG_FILES_0 {
		ids := logUploadSettings.LogFileIds
		logFiles := getLogFilesByIds(ids)

		oneList, err := logupload.GetOneLogFileList(logUploadSettings.ID)
		for i, logFileInList := range oneList.Data {
			for _, logFile := range logFiles {
				if logFile.ID == logFileInList.ID {
					//remove this logFile from logFileList
					oneList.Data = append(oneList.Data[:i], oneList.Data[i+1:]...)
					break
				}
			}
		}
		oneList.Data = append(oneList.Data, logFiles...)
		logupload.DeleteOneLogFileList(logUploadSettings.ID)

		err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_LOG_FILE_LIST, logUploadSettings.ID, oneList)
		if err != nil {
			log.Warn(fmt.Sprintf("error save logFileList for Id: %s", logUploadSettings.ID))
			xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, "Failed to save logFileList")
		}
	}
	if !checkDateStrLength(logUploadSettings.FromDateTime) || !checkDateStrLength(logUploadSettings.ToDateTime) {
		logUploadSettings.FromDateTime = ""
		logUploadSettings.ToDateTime = ""
	} else if timezone != "" && timezone != "UTC" {
		logUploadSettings.FromDateTime = converterDateTimeToUTC(logUploadSettings.FromDateTime, timezone)
		logUploadSettings.ToDateTime = converterDateTimeToUTC(logUploadSettings.ToDateTime, timezone)
	}
	if !checkDateStrLength(schedule.StartDate) || !checkDateStrLength(schedule.EndDate) {
		schedule.StartDate = ""
		schedule.EndDate = ""
	} else if scheduleTimezone != "" && scheduleTimezone != "UTC" {
		schedule.StartDate = converterDateTimeToUTC(schedule.StartDate, scheduleTimezone)
		schedule.EndDate = converterDateTimeToUTC(schedule.EndDate, scheduleTimezone)
	}
	logUploadSettings.Schedule = schedule
	logupload.SetOneLogUploadSettings(logUploadSettings.ID, &logUploadSettings)
	response, err := util.JSONMarshal(logUploadSettings)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRuleNew error: %v", err))
	}
	xhttp.WriteXconfResponse(w, http.StatusCreated, response)
}

func checkDateStrLength(dateStr string) bool {
	if dateStr != "" && len(dateStr) == 19 {
		return true
	}
	return false
}

func isValidDate(timeStr string) bool {
	layout := "2006-01-02 15:04:05"
	if _, err := time.Parse(layout, timeStr); err != nil {
		return false
	}
	return true
}

func isEnddateAfterStartDate(startDateStr string, endDateStr string) bool {
	// return true when endDate is After startDate
	layout := "2006-01-02 15:04:05"
	startDate, _ := time.Parse(layout, startDateStr)
	endDate, _ := time.Parse(layout, endDateStr)
	return endDate.After(startDate)
}

func converterDateTimeToUTC(timeStr string, sourceTZ string) string {
	layout := "2006-01-02 15:04:05"
	loc, err := time.LoadLocation(sourceTZ)
	if err != nil {
		return timeStr
	}
	t, err := time.ParseInLocation(layout, timeStr, loc)
	if err != nil {
		return timeStr
	}
	return t.UTC().Format(layout)
}

func validateName(logUploadSettings *logupload.LogUploadSettings) string {
	logUploadSettingsList, err := logupload.GetAllLogUploadSettings(0)
	if err != nil {
		return ""
	}
	for _, logUploadSettingsDB := range logUploadSettingsList {
		if logUploadSettingsDB.Name == logUploadSettings.Name && logUploadSettingsDB.ID != logUploadSettings.ID {
			return "Name is already used"
		}
	}
	return ""
}

func getLogFilesByIds(ids []string) []*logupload.LogFile {
	logFiles := []*logupload.LogFile{}
	logFileList := logupload.GetLogFileList(0) //logFileList is a list of LogFiles
	for _, id := range ids {
		for _, logFile := range logFileList {
			if logFile.ID == id {
				logFiles = append(logFiles, logFile)
			}
		}
	}
	return logFiles
}
