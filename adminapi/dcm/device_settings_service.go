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
package dcm

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"xconfwebconfig/shared/logupload"
	"xconfwebconfig/util"

	"github.com/google/uuid"

	xcommon "xconfadmin/common"
	xutil "xconfadmin/util"
	xwcommon "xconfwebconfig/common"
	"xconfwebconfig/db"
	xwhttp "xconfwebconfig/http"
	"xconfwebconfig/shared"

	log "github.com/sirupsen/logrus"
)

const (
	cDeviceSettingsPageNumber = "pageNumber"
	cDeviceSettingsPageSize   = "pageSize"
)

func GetDeviceSettingsList() []*logupload.DeviceSettings {
	all := []*logupload.DeviceSettings{}
	deviceSettingsList, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_DEVICE_SETTINGS, 0)
	if err != nil {
		log.Warn("no DeviceSettings found")
		return all
	}
	for idx := range deviceSettingsList {
		if deviceSettingsList[idx] != nil {
			ds := deviceSettingsList[idx].(*logupload.DeviceSettings)
			all = append(all, ds)
		}
	}
	return all
}

func GetDeviceSettingsAll() []*logupload.DeviceSettings {
	result := []*logupload.DeviceSettings{}
	result = GetDeviceSettingsList()
	return result
}

func GetDeviceSettings(id string) *logupload.DeviceSettings {
	devicesettings := logupload.GetOneDeviceSettings(id)
	if devicesettings != nil {
		return devicesettings
	}
	return nil

}

func validateUsageForDeviceSettings(Id string, app string) (string, error) {
	ds := GetDeviceSettings(Id)
	if ds == nil {
		return fmt.Sprintf("Entity with id  %s does not exist ", Id), nil
	}
	if ds.ApplicationType != app {
		return fmt.Sprintf("Entity with id  %s ApplicationType sodent match ", Id), nil
	}
	return "", nil
}

func DeleteDeviceSettingsbyId(id string, app string) *xwhttp.ResponseEntity {
	usage, err := validateUsageForDeviceSettings(id, app)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, err, nil)
	}

	if usage != "" {
		return xwhttp.NewResponseEntity(http.StatusNotFound, errors.New(usage), nil)
	}

	err = DeleteOneDeviceSettings(id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func DeleteOneDeviceSettings(id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(db.TABLE_DEVICE_SETTINGS, id)
	if err != nil {
		return err
	}
	return nil
}

func DeviceSettingsValidate(ds *logupload.DeviceSettings) *xwhttp.ResponseEntity {
	if ds == nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("DeviceSettings should be specified"), nil)
	}
	if util.IsBlank(ds.ID) {
		ds.ID = uuid.New().String()
	}
	if util.IsBlank(ds.ApplicationType) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ApplicationType is empty"), nil)
	}
	if util.IsBlank(ds.Name) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Name is empty"), nil)
	}
	if ds.Schedule == (logupload.Schedule{}) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Schedule is empty"), nil)
	}

	schedule := ds.Schedule
	if util.IsBlank(schedule.TimeZone) || (schedule.TimeZone != logupload.LOCAL_TIME && schedule.TimeZone != logupload.UTC) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("TimeZone must be set to 'Local time' or 'UTC'"), nil)
	}
	if util.IsBlank(schedule.Expression) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Schedule Expression is empty"), nil)
	}
	if util.IsBlank(schedule.Type) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Schedule Type is empty"), nil)
	}

	twm := schedule.TimeWindowMinutes
	if twm < 0 {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Schedule TimeWindowMinutes is invalid"), nil)
	}

	if err := xutil.ValidateCronDayAndMonth(schedule.Expression); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	dsrules := GetDeviceSettingsList()
	for _, exdsrule := range dsrules {
		if exdsrule.ApplicationType != ds.ApplicationType {
			continue
		}
		if exdsrule.ID != ds.ID {
			if exdsrule.Name == ds.Name {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("DeviceSettings name is already used"), nil)
			}
		}
	}

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, nil)
}

func CreateDeviceSettings(dset *logupload.DeviceSettings, app string) *xwhttp.ResponseEntity {
	if existingSettings := logupload.GetOneDeviceSettings(dset.ID); existingSettings != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s already exists", dset.ID), nil)
	}
	if dset.ApplicationType == "" {
		dset.ApplicationType = app
	} else if dset.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", dset.ID), nil)
	}
	if respEntity := DeviceSettingsValidate(dset); respEntity.Error != nil {
		return respEntity
	}

	dset.Updated = util.GetTimestamp(time.Now().UTC())
	if err := db.GetCachedSimpleDao().SetOne(db.TABLE_DEVICE_SETTINGS, dset.ID, dset); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, dset)
}

func UpdateDeviceSettings(dset *logupload.DeviceSettings, app string) *xwhttp.ResponseEntity {
	if util.IsBlank(dset.ID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ID is empty"), nil)
	}
	existingSettings := logupload.GetOneDeviceSettings(dset.ID)
	if existingSettings == nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s does not exists", dset.ID), nil)
	}
	if dset.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", dset.ID), nil)
	}
	if existingSettings.ApplicationType != dset.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New("ApplicationType can not be changed"), nil)
	}
	if respEntity := DeviceSettingsValidate(dset); respEntity.Error != nil {
		return respEntity
	}

	dset.Updated = util.GetTimestamp(time.Now().UTC())
	if err := db.GetCachedSimpleDao().SetOne(db.TABLE_DEVICE_SETTINGS, dset.ID, dset); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}
	return xwhttp.NewResponseEntity(http.StatusOK, nil, dset)
}

func DeviceSettingsGeneratePage(list []*logupload.DeviceSettings, page int, pageSize int) (result []*logupload.DeviceSettings) {
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

func DeviceSettingsGeneratePageWithContext(dsrules []*logupload.DeviceSettings, contextMap map[string]string) (result []*logupload.DeviceSettings, err error) {
	sort.Slice(dsrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(dsrules[i].Name), strings.ToLower(dsrules[j].Name)) < 0
	})
	pageNum := 1
	numStr, okval := contextMap[cDeviceSettingsPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cDeviceSettingsPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, errors.New("pageNumber and pageSize should both be greater than zero")
	}
	return DeviceSettingsGeneratePage(dsrules, pageNum, pageSize), nil
}

func DeviceSettingsFilterByContext(searchContext map[string]string) []*logupload.DeviceSettings {
	deviceSettingsRules := GetDeviceSettingsList()
	deviceSettingsRuleList := []*logupload.DeviceSettings{}
	for _, dsRule := range deviceSettingsRules {
		if dsRule == nil {
			continue
		}
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if dsRule.ApplicationType != applicationType && dsRule.ApplicationType != shared.ALL {
				continue
			}
		}
		if name, ok := xutil.FindEntryInContext(searchContext, xcommon.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(dsRule.Name), strings.ToLower(name)) {
				continue
			}
		}
		deviceSettingsRuleList = append(deviceSettingsRuleList, dsRule)
	}
	return deviceSettingsRuleList
}
