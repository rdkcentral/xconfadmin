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
	"net/url"
	"sort"
	"strconv"
	"strings"

	xutil "github.com/rdkcentral/xconfadmin/util"

	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/rdkcentral/xconfadmin/common"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"

	log "github.com/sirupsen/logrus"
)

const (
	cLogRepoSettingsPageNumber = "pageNumber"
	cLogRepoSettingsPageSize   = "pageSize"
)

func GetLogRepoSettingsList() []*logupload.UploadRepository {
	all := []*logupload.UploadRepository{}
	logRepoSettingsList, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_UPLOAD_REPOSITORY, 0)
	if err != nil {
		log.Warn("no LogRepSettings found")
		return all
	}
	for idx := range logRepoSettingsList {
		if logRepoSettingsList[idx] != nil {
			ds := logRepoSettingsList[idx].(*logupload.UploadRepository)
			all = append(all, ds)
		}
	}
	return all
}

func GetLogRepoSettingsAll() []*logupload.UploadRepository {
	result := []*logupload.UploadRepository{}
	result = GetLogRepoSettingsList()
	return result
}

func GetOneLogRepoSettings(id string) *logupload.UploadRepository {
	var logRepoSettings *logupload.UploadRepository
	logRepoSettingsInst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_UPLOAD_REPOSITORY, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no LogRepoSettings found for Id: %s", id))
		return nil
	}
	logRepoSettings = logRepoSettingsInst.(*logupload.UploadRepository)
	return logRepoSettings
}

func GetLogRepoSettings(id string) *logupload.UploadRepository {
	logRepoSettings := GetOneLogRepoSettings(id)
	if logRepoSettings != nil {
		return logRepoSettings
	}
	return nil

}

func validateUsageForLogRepoSettings(Id string, app string) error {
	lr := GetLogRepoSettings(Id)
	if lr == nil {
		return fmt.Errorf("Entity with id  %s does not exist ", Id)
	}
	if lr.ApplicationType != app {
		return fmt.Errorf("Entity with id  %s ApplicationType doesn't match ", Id)
	}
	return nil
}

func DeleteLogRepoSettingsbyId(id string, app string) *xwhttp.ResponseEntity {
	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_UPLOAD_REPOSITORY, id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf(" %s is not found", id), nil)
	}
	lu := inst.(*logupload.UploadRepository)

	lurules := GetLogUploadSettingsList()
	referred := []string{}
	for _, item := range lurules {
		if item.ApplicationType != lu.ApplicationType {
			continue
		}
		if item.UploadRepositoryID == lu.ID {
			referred = append(referred, item.Name)
		}
	}
	if len(referred) != 0 {
		referredLUs := strings.Join(referred, ", ")
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("%s is used by %d LogUploadSettings (%s)", lu.Name, len(referred), referredLUs), nil)
	}
	err = validateUsageForLogRepoSettings(id, app)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, err, nil)
	}

	err = DeleteOneLogRepoSettings(id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}
func DeleteOneLogRepoSettings(id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(db.TABLE_UPLOAD_REPOSITORY, id)
	if err != nil {
		return err
	}
	return nil
}

func LogRepoSettingsValidate(lr *logupload.UploadRepository) *xwhttp.ResponseEntity {
	if lr == nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Log Repository Settings should be specified"), nil)
	}

	if util.IsBlank(lr.ID) {
		lr.ID = uuid.New().String()
	}
	if util.IsBlank(lr.ApplicationType) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("ApplicationType is empty"), nil)
	}

	if util.IsBlank(lr.Name) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Name is empty"), nil)
	}
	if util.IsBlank(lr.URL) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("URL is empty"), nil)
	}
	if _, err := url.ParseRequestURI(lr.URL); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("URL is InValid"), nil)
	}
	if util.IsBlank(lr.Protocol) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Protocol is empty"), nil)
	}
	if !logupload.IsValidUploadProtocol(lr.Protocol) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("URL is InValid"), nil)
	}

	lrrules := GetLogRepoSettingsAll()
	for _, exlrrule := range lrrules {
		if exlrrule.ApplicationType != lr.ApplicationType {
			continue
		}
		if exlrrule.ID != lr.ID {
			if exlrrule.Name == lr.Name {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Name is alread used"), nil)
			}
		}
	}
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, nil)
}

func CreateLogRepoSettings(lr *logupload.UploadRepository, app string) *xwhttp.ResponseEntity {
	_, err := db.GetCachedSimpleDao().GetOne(db.TABLE_UPLOAD_REPOSITORY, lr.ID)
	if err == nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s already exists", lr.ID)), nil)
	}
	if lr.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s ApplicationType doesn't match", lr.ID)), nil)
	}

	respEntity := LogRepoSettingsValidate(lr)
	if respEntity.Error != nil {
		return respEntity
	}
	lr.Updated = util.GetTimestamp()
	if err = db.GetCachedSimpleDao().SetOne(db.TABLE_UPLOAD_REPOSITORY, lr.ID, lr); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, lr)
}

func UpdateLogRepoSettings(lr *logupload.UploadRepository, app string) *xwhttp.ResponseEntity {
	if util.IsBlank(lr.ID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New(" ID  is empty"), nil)
	}
	inst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_UPLOAD_REPOSITORY, lr.ID)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s does not exists", lr.ID)), nil)
	}
	if lr.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s ApplicationType doesn't match", lr.ID)), nil)
	}
	exlrrule := inst.(*logupload.UploadRepository)
	if exlrrule.ApplicationType != lr.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("ApplicationType can not be changed")), nil)
	}
	respEntity := LogRepoSettingsValidate(lr)
	if respEntity.Error != nil {
		return respEntity
	}

	lr.Updated = util.GetTimestamp()
	if err = db.GetCachedSimpleDao().SetOne(db.TABLE_UPLOAD_REPOSITORY, lr.ID, lr); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, lr)
}

func LogRepoSettingsGeneratePage(list []*logupload.UploadRepository, page int, pageSize int) (result []*logupload.UploadRepository) {
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

func LogRepoSettingsGeneratePageWithContext(lrrules []*logupload.UploadRepository, contextMap map[string]string) (result []*logupload.UploadRepository, err error) {
	sort.Slice(lrrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(lrrules[i].Name), strings.ToLower(lrrules[j].Name)) < 0
	})
	pageNum := 1
	numStr, okval := contextMap[cLogRepoSettingsPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cLogRepoSettingsPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, errors.New("pageNumber and pageSize should both be greater than zero")
	}
	return LogRepoSettingsGeneratePage(lrrules, pageNum, pageSize), nil
}

func LogRepoSettingsFilterByContext(searchContext map[string]string) []*logupload.UploadRepository {
	logRepoSettingsRules := GetLogRepoSettingsList()
	logRepoSettingsRuleList := []*logupload.UploadRepository{}
	for _, lrRule := range logRepoSettingsRules {
		if lrRule == nil {
			continue
		}
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if lrRule.ApplicationType != applicationType && lrRule.ApplicationType != shared.ALL {
				continue
			}
		}
		if name, ok := xutil.FindEntryInContext(searchContext, common.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(lrRule.Name), strings.ToLower(name)) {
				continue
			}
		}
		logRepoSettingsRuleList = append(logRepoSettingsRuleList, lrRule)
	}
	return logRepoSettingsRuleList
}
