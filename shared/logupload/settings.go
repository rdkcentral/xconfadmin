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
	"fmt"
	"xconfwebconfig/db"
	"xconfwebconfig/shared/logupload"

	log "github.com/sirupsen/logrus"
)

var SettingTypes = [...]string{"PARTNER_SETTINGS", "EPON", "partnersettings", "epon"}

func GetOneSettingProfile(id string) *logupload.SettingProfiles {
	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_SETTING_PROFILES, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no SettingProfile found for %s", id))
		return nil
	}
	telemetry := inst.(*logupload.SettingProfiles)
	return telemetry
}

func IsValidSettingType(str string) bool {
	for _, v := range SettingTypes {
		if v == str {
			return true
		}
	}
	return false
}

func GetOneLogFileList(id string) (*logupload.LogFileList, error) {
	var logFileList *logupload.LogFileList
	logFileListInst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_LOG_FILE_LIST, id)
	if err != nil {
		logFileList = &logupload.LogFileList{}
	} else {
		logFileList = logFileListInst.(*logupload.LogFileList)
	}
	if logFileList.Data == nil {
		logFileList.Data = []*logupload.LogFile{}
	}
	return logFileList, nil
}

func SetOneLogFile(id string, obj *logupload.LogFile) error {
	oneList, err := GetOneLogFileList(id)
	for i, logFile := range oneList.Data {
		if logFile.ID == obj.ID {
			oneList.Data = append(oneList.Data[:i], oneList.Data[i+1:]...)
			break
		}
	}
	oneList.Data = append(oneList.Data, obj)
	err = db.GetCachedSimpleDao().SetOne(db.TABLE_LOG_FILE_LIST, id, oneList)
	if err != nil {
		log.Warn(fmt.Sprintf("error save logFileList for Id: %s", id))
		return err
	}
	return nil
}

func GetAllLogUploadSettings(size int) ([]*logupload.LogUploadSettings, error) {
	var logUploadSettingsList []*logupload.LogUploadSettings
	logUploadSettingsInst, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_LOG_UPLOAD_SETTINGS, size)
	if err != nil {
		log.Warn("error finding logUploadSettings ")
		return nil, err
	}
	for idx := range logUploadSettingsInst {
		logUploadSettings := logUploadSettingsInst[idx].(*logupload.LogUploadSettings)
		logUploadSettingsList = append(logUploadSettingsList, logUploadSettings)
	}
	return logUploadSettingsList, err
}

func DeleteOneLogFileList(id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(db.TABLE_LOG_FILE_LIST, id)
	return err
}

func SetOneLogUploadSettings(id string, logUploadSettings *logupload.LogUploadSettings) error {
	err := db.GetCachedSimpleDao().SetOne(db.TABLE_LOG_UPLOAD_SETTINGS, id, logUploadSettings)
	if err != nil {
		log.Warn(fmt.Sprintf("error saving logUploadSettings for Id: %s", id))
	}
	return err
}
