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
package logupload

import (
	"xconfwebconfig/db"
	"xconfwebconfig/shared/logupload"

	log "github.com/sirupsen/logrus"
)

func SetLogFile(id string, logFile *logupload.LogFile) error {
	err := db.GetCachedSimpleDao().SetOne(db.TABLE_LOG_FILE, id, logFile)
	if err != nil {
		log.Warn("error saving logFile ")
	}
	return err
}

func GetLogFileGroupsList(size int) ([]*logupload.LogFilesGroups, error) {
	var logFilesGroupsList []*logupload.LogFilesGroups
	logFilesGroupsInst, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_LOG_FILES_GROUPS, size)
	if err != nil {
		log.Warn("no logFilesGroups found ")
		return nil, err
	}
	for idx := range logFilesGroupsInst {
		logFilesGroups := logFilesGroupsInst[idx].(*logupload.LogFilesGroups)
		logFilesGroupsList = append(logFilesGroupsList, logFilesGroups)
	}
	return logFilesGroupsList, nil
}
