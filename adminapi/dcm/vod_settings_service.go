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
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"

	xcommon "github.com/rdkcentral/xconfadmin/common"
	xutil "github.com/rdkcentral/xconfadmin/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	xwutil "github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	cVodSettingsPageNumber = "pageNumber"
	cVodSettingsPageSize   = "pageSize"
)

func GetVodSettingsList() []*logupload.VodSettings {
	all := []*logupload.VodSettings{}
	vodSettingsList, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_VOD_SETTINGS, 0)
	if err != nil {
		log.Warn("no VodSettings found")
		return all
	}
	for idx := range vodSettingsList {
		if vodSettingsList[idx] != nil {
			ds := vodSettingsList[idx].(*logupload.VodSettings)
			all = append(all, ds)
		}
	}
	return all
}

func GetVodSettingsAll() []*logupload.VodSettings {
	result := []*logupload.VodSettings{}
	result = GetVodSettingsList()
	return result
}

func GetVodSettings(id string) *logupload.VodSettings {
	vodsettings := logupload.GetOneVodSettings(id)
	if vodsettings != nil {
		return vodsettings
	}
	return nil
}

func validateUsageForVodSettings(Id string, app string) (string, error) {
	vs := GetVodSettings(Id)
	if vs == nil {
		return fmt.Sprintf("Entity with id  %s does not exist ", Id), nil
	}
	if vs.ApplicationType != app {
		return fmt.Sprintf("Entity with id  %s does not exist ,ApplicationType mismatch", Id), nil
	}
	return "", nil
}

func DeleteVodSettingsbyId(id string, app string) *xwhttp.ResponseEntity {
	usage, err := validateUsageForVodSettings(id, app)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, err, nil)
	}

	if usage != "" {
		return xwhttp.NewResponseEntity(http.StatusNotFound, errors.New(usage), nil)
	}

	err = DeleteOneVodSettings(id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func DeleteOneVodSettings(id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(db.TABLE_VOD_SETTINGS, id)
	if err != nil {
		return err
	}
	return nil
}

func VodSettingsValidate(vs *logupload.VodSettings) *xwhttp.ResponseEntity {
	if vs == nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("VodSettings should be specified"), nil)
	}
	if xwutil.IsBlank(vs.ID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ID is empty"), nil)
	}
	if xwutil.IsBlank(vs.ApplicationType) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ApplicationType is empty"), nil)
	}
	if xwutil.IsBlank(vs.Name) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Name is empty"), nil)
	}
	if xwutil.IsBlank(vs.LocationsURL) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("URL is empty"), nil)
	}
	if !logupload.IsValidUrl(vs.LocationsURL) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("URL is InValid"), nil)
	}
	if len(vs.IPNames) != len(vs.IPList) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("IP names and IP address doesn't match"), nil)
	}
	if len(vs.IPNames) == 0 && len(vs.IPList) == 0 {
		vs.SrmIPList = map[string]string{}
	} else {
		for _, ip := range vs.IPList {
			if net.ParseIP(ip) == nil {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("IP Address Invalid"), nil)
			}
		}
		//build now hashmap
		IPListMap := make(map[string]string)
		vs.SrmIPList = IPListMap
		for idx, ip := range vs.IPList {
			vs.SrmIPList[vs.IPNames[idx]] = ip
		}
	}

	vsrules := GetVodSettingsAll()
	for _, exvsrule := range vsrules {
		if exvsrule.ApplicationType != vs.ApplicationType {
			continue
		}
		if exvsrule.ID != vs.ID {
			if exvsrule.Name == vs.Name {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("VodSettings name is already used"), nil)
			}
		}
	}

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, nil)
}

func CreateVodSettings(vs *logupload.VodSettings, app string) *xwhttp.ResponseEntity {
	if existingSettings := logupload.GetOneVodSettings(vs.ID); existingSettings != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s already exists", vs.ID)), nil)
	}
	if vs.ApplicationType == "" {
		vs.ApplicationType = app
	} else if vs.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s ApplicationType mismatch", vs.ID)), nil)
	}
	respEntity := VodSettingsValidate(vs)
	if respEntity.Error != nil {
		return respEntity
	}

	vs.Updated = xutil.GetTimestamp()
	if err := db.GetCachedSimpleDao().SetOne(db.TABLE_VOD_SETTINGS, vs.ID, vs); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, vs)
}

func UpdateVodSettings(vs *logupload.VodSettings, app string) *xwhttp.ResponseEntity {
	if xwutil.IsBlank(vs.ID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ID is empty"), nil)
	}
	existingSettings := logupload.GetOneVodSettings(vs.ID)
	if existingSettings == nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("Entity with id %s does not exists", vs.ID)), nil)
	}
	if existingSettings.ApplicationType != vs.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(fmt.Sprintf("ApplicationType can not be changed")), nil)
	}
	if respEntity := VodSettingsValidate(vs); respEntity.Error != nil {
		return respEntity
	}

	vs.Updated = xwutil.GetTimestamp()
	if err := db.GetCachedSimpleDao().SetOne(db.TABLE_VOD_SETTINGS, vs.ID, vs); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, vs)
}

func VodSettingsGeneratePage(list []*logupload.VodSettings, page int, pageSize int) (result []*logupload.VodSettings) {
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

func VodSettingsGeneratePageWithContext(vsrules []*logupload.VodSettings, contextMap map[string]string) (result []*logupload.VodSettings, err error) {
	sort.Slice(vsrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(vsrules[i].Name), strings.ToLower(vsrules[j].Name)) < 0
	})
	pageNum := 1
	numStr, okval := contextMap[cVodSettingsPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cVodSettingsPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, errors.New("pageNumber and pageSize should both be greater than zero")
	}
	return VodSettingsGeneratePage(vsrules, pageNum, pageSize), nil
}

func VodSettingsFilterByContext(searchContext map[string]string) []*logupload.VodSettings {
	vodSettingsRules := GetVodSettingsList()
	vodSettingsRuleList := []*logupload.VodSettings{}
	for _, vsRule := range vodSettingsRules {
		if vsRule == nil {
			continue
		}
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if vsRule.ApplicationType != applicationType && vsRule.ApplicationType != shared.ALL {
				continue
			}
		}
		if name, ok := xutil.FindEntryInContext(searchContext, xcommon.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(vsRule.Name), strings.ToLower(name)) {
				continue
			}
		}
		vodSettingsRuleList = append(vodSettingsRuleList, vsRule)
	}
	return vodSettingsRuleList
}
