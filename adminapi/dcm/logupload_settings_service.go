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
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	xcommon "xconfadmin/common"
	xwcommon "xconfwebconfig/common"
	"xconfwebconfig/shared/logupload"
	"xconfwebconfig/util"

	xutil "xconfadmin/util"

	"github.com/google/uuid"

	ds "xconfwebconfig/db"
	xwhttp "xconfwebconfig/http"
	"xconfwebconfig/shared"

	log "github.com/sirupsen/logrus"
)

const (
	FEBRUARY                     = 2
	LEAPYEARDAYS                 = 29
	cLogUploadSettingsPageNumber = "pageNumber"
	cLogUploadSettingsPageSize   = "pageSize"
)

func GetLogUploadSettingsList() []*logupload.LogUploadSettings {
	all := []*logupload.LogUploadSettings{}
	loguploadSettingsList, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_LOG_UPLOAD_SETTINGS, 0)
	if err != nil {
		log.Warn("no LogUploadSettings found")
		return all
	}
	for idx := range loguploadSettingsList {
		if loguploadSettingsList[idx] != nil {
			ds := loguploadSettingsList[idx].(*logupload.LogUploadSettings)
			all = append(all, ds)
		}
	}
	return all
}

func DeleteLogUploadSettingsbyId(id string, app string) *xwhttp.ResponseEntity {

	lu := logupload.GetOneLogUploadSettings(id)
	if lu == nil {

		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("Entity with id  %s does not exist ", id), nil)
	}
	if lu.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("Entity with id  %s ApplicationType doesn't match ", id), nil)
	}
	err := DeleteOneLogUploadSettings(id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func DeleteOneLogUploadSettings(id string) error {
	err := ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_LOG_UPLOAD_SETTINGS, id)
	if err != nil {
		return err
	}
	return nil
}

func LogUploadSettingsValidate(lu *logupload.LogUploadSettings) *xwhttp.ResponseEntity {
	if lu == nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("LogUploadSettings should be specified"), nil)
	}
	if util.IsBlank(lu.ID) {
		lu.ID = uuid.New().String()
	}
	if util.IsBlank(lu.ApplicationType) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("ApplicationType is empty"), nil)
	}
	if util.IsBlank(lu.Name) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Name is empty"), nil)
	}
	if util.IsBlank(lu.UploadRepositoryID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("UploadRepositoryID is empty"), nil)
	}
	if lu.Schedule == (logupload.Schedule{}) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Schedule is empty"), nil)
	}
	schedule := lu.Schedule
	if util.IsBlank(schedule.TimeZone) || (schedule.TimeZone != logupload.LOCAL_TIME && schedule.TimeZone != logupload.UTC) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("TimeZone must be set to 'Local time' or 'UTC'"), nil)
	}
	if util.IsBlank(schedule.Expression) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Schedule Expression is empty"), nil)
	}
	if util.IsBlank(schedule.Type) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Schedule Type is empty"), nil)
	}
	expressionArray := []string{}
	expressionArray = strings.Split(schedule.Expression, " ")
	if len(expressionArray) < 4 {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Schedule expression Invalid less than required legth"), nil)
	}

	dayofMonth, _ := strconv.Atoi(expressionArray[2])
	month, _ := strconv.Atoi(expressionArray[3])
	dayofMonthstr := expressionArray[2]
	monthstr := expressionArray[3]
	if dayofMonthstr != "*" || monthstr != "*" {
		if month != FEBRUARY && dayofMonth != LEAPYEARDAYS {
			timeStr := monthstr + "-" + dayofMonthstr
			if _, err := time.Parse("1-2", timeStr); err != nil {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Schedule Month and day Invalid"), nil)
			}
		}
	}

	lurules := GetLogUploadSettingsList()
	for _, exlurule := range lurules {
		if exlurule.ApplicationType != lu.ApplicationType {
			continue
		}
		if exlurule.ID != lu.ID {
			if exlurule.Name == lu.Name {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("LogUploadSetting name is already used"), nil)
			}
		}
	}
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, nil)
}

func CreateLogUploadSettings(lu *logupload.LogUploadSettings, app string) *xwhttp.ResponseEntity {
	if existingSettings := logupload.GetOneLogUploadSettings(lu.ID); existingSettings != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s already exists", lu.ID)), nil)
	}
	if lu.ApplicationType == "" {
		lu.ApplicationType = app
	} else if lu.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s ApplicationType mismatch", lu.ID)), nil)
	}
	if respEntity := LogUploadSettingsValidate(lu); respEntity.Error != nil {
		return respEntity
	}

	lu.Updated = util.GetTimestamp()
	if err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_LOG_UPLOAD_SETTINGS, lu.ID, lu); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, lu)
}

func UpdateLogUploadSettings(lu *logupload.LogUploadSettings, app string) *xwhttp.ResponseEntity {
	if util.IsBlank(lu.ID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ID is empty"), nil)
	}
	if lu.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s ApplicationType mismatch", lu.ID)), nil)
	}
	existingSettings := logupload.GetOneLogUploadSettings(lu.ID)
	if existingSettings == nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s does not exists", lu.ID)), nil)
	}
	if existingSettings.ApplicationType != lu.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("ApplicationType can not be changed")), nil)
	}
	if respEntity := LogUploadSettingsValidate(lu); respEntity.Error != nil {
		return respEntity
	}

	lu.Updated = util.GetTimestamp()
	if err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_LOG_UPLOAD_SETTINGS, lu.ID, lu); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, lu)
}

func LogUploadSettingsGeneratePage(list []*logupload.LogUploadSettings, page int, pageSize int) (result []*logupload.LogUploadSettings) {
	leng := len(list)
	startIndex := page*pageSize - pageSize
	if page < 1 || startIndex > leng || pageSize < 1 {
		return result
	}
	lastIndex := leng
	if page*pageSize < len(list) {
		lastIndex = page * pageSize
	}

	return list[startIndex:lastIndex]
}

func LogUploadSettingsGeneratePageWithContext(lurules []*logupload.LogUploadSettings, contextMap map[string]string) (result []*logupload.LogUploadSettings, err error) {
	sort.Slice(lurules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(lurules[i].Name), strings.ToLower(lurules[j].Name)) < 0
	})
	pageNum := 1
	numStr, okval := contextMap[cLogUploadSettingsPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cLogUploadSettingsPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, errors.New("pageNumber and pageSize should both be greater than zero")
	}
	return LogUploadSettingsGeneratePage(lurules, pageNum, pageSize), nil
}

func LogUploadSettingsFilterByContext(searchContext map[string]string) []*logupload.LogUploadSettings {
	logUploadSettingsRules := GetLogUploadSettingsList()
	logUploadSettingsRuleList := []*logupload.LogUploadSettings{}
	for _, luRule := range logUploadSettingsRules {
		if luRule == nil {
			continue
		}
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if luRule.ApplicationType != applicationType && luRule.ApplicationType != shared.ALL {
				continue
			}
		}
		if name, ok := xutil.FindEntryInContext(searchContext, xcommon.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(luRule.Name), strings.ToLower(name)) {
				continue
			}
		}
		logUploadSettingsRuleList = append(logUploadSettingsRuleList, luRule)
	}
	return logUploadSettingsRuleList
}
