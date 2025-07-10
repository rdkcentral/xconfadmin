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
	"math"
	"net/http"
	"strings"

	"xconfadmin/adminapi/auth"
	xhttp "xconfadmin/http"
	"xconfadmin/shared/logupload"
	"xconfadmin/util"
	xwcommon "xconfwebconfig/common"
	xwhttp "xconfwebconfig/http"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func CreateLogFile(w http.ResponseWriter, r *http.Request) {
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
	logFile := logupload.LogFile{}
	err = json.Unmarshal([]byte(body), &logFile)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if logFile.Name == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Log file is empty")
		return
	}
	if !isValidName(logFile) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Name is already used")
		return
	}
	if logFile.ID == "" {
		logFile.ID = uuid.New().String()
		err := logupload.SetLogFile(logFile.ID, &logFile)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		err := logupload.SetLogFile(logFile.ID, &logFile)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		err = updateLogUploadSettingsAndLogFileGroups(&logFile)
		if err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	response, err := util.JSONMarshal(logFile)
	if err != nil {
		log.Error(fmt.Sprintf("json.Marshal featureRuleNew error: %v", err))
	}
	xhttp.WriteXconfResponse(w, http.StatusCreated, response)
}

func isValidName(logFile logupload.LogFile) bool {
	if logFile.Name == "" {
		return false
	}
	lf := getLogFileByName(strings.Trim(logFile.Name, " "))
	if lf != nil && lf.ID != logFile.ID {
		return false
	}
	return true
}

func getLogFileByName(name string) *logupload.LogFile {
	logFileList := logupload.GetLogFileList(0) //logFileList is a list of LogFiles
	for _, logFile := range logFileList {
		if logFile.Name == name {
			return logFile
		}
	}
	return nil
}

func updateLogUploadSettingsAndLogFileGroups(logFile *logupload.LogFile) error {
	listLogUploadSettings, err := logupload.GetAllLogUploadSettings(math.MaxInt32 / 100)
	if err != nil {
		return err
	}
	for _, logUploadSettings := range listLogUploadSettings {
		LogFileList, err := logupload.GetOneLogFileList(logUploadSettings.ID)
		if err != nil {
			log.Warn(fmt.Sprintf("error getting LogFileList for logUploadSettings.Id: %s", logUploadSettings.ID))
			continue
		}
		for _, logFileDB := range LogFileList.Data {
			if logFileDB.ID == logFile.ID {
				logupload.SetOneLogFile(logUploadSettings.ID, logFile)
			}
		}
	}
	listLogFilesGroups, err := logupload.GetLogFileGroupsList(math.MaxInt32 / 100)
	if err != nil {
		return err
	}
	for _, logFilesGroup := range listLogFilesGroups {
		LogFileList, err := logupload.GetOneLogFileList(logFilesGroup.ID)
		if err != nil {
			log.Warn(fmt.Sprintf("error getting LogFileList for logUploadSettings.Id: %s", logFilesGroup.ID))
		}
		for _, logFileDB := range LogFileList.Data {
			if logFileDB.ID == logFile.ID {
				logupload.SetOneLogFile(logFilesGroup.ID, logFile)
			}
		}
	}
	return nil
}
