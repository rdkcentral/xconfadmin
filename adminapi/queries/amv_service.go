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
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	util "github.com/rdkcentral/xconfadmin/util"
	xutil "github.com/rdkcentral/xconfadmin/util"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	ru "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	xwutil "github.com/rdkcentral/xconfwebconfig/util"
)

const (
	amvModel          = "MODEL"
	amvDescription    = "DESCRIPTION"
	amvPartnerId      = "PARTNER_ID"
	amvPartnerIdAlias = "partnerId"
	amvFwVersion      = "FIRMWARE_VERSION"
	amvRegex          = "REGULAR_EXPRESSION"
	amvPageNumber     = "pageNumber"
	amvPageSize       = "pageSize"
	amvFwVersionAlias = "firmwareVersion"
	amvRegexAlias     = "regularExpression"
)

type ActivationVersionResponse struct {
	ID                 string   `json:"id"`
	ApplicationType    string   `json:"applicationType"`
	Description        string   `json:"description,omitempty"`
	Model              string   `json:"model,omitempty"`
	PartnerId          string   `json:"partnerId,omitempty"`
	RegularExpressions []string `json:"regularExpressions"`
	FirmwareVersions   []string `json:"firmwareVersions"`
}

func CreateActivationVersionResponse(rec *firmware.ActivationVersion) *ActivationVersionResponse {
	resp := ActivationVersionResponse{}
	resp.ID = rec.ID
	resp.ApplicationType = rec.ApplicationType
	resp.Description = rec.Description
	resp.Model = rec.Model
	resp.PartnerId = rec.PartnerId
	resp.RegularExpressions = make([]string, len(rec.RegularExpressions))
	copy(resp.RegularExpressions, rec.RegularExpressions)
	resp.FirmwareVersions = make([]string, len(rec.FirmwareVersions))
	copy(resp.FirmwareVersions, rec.FirmwareVersions)

	return &resp
}

func GetAmvALL(tenantId string) []*ActivationVersionResponse {
	result := []*ActivationVersionResponse{}
	amvs := GetAllAmvList(tenantId)
	for _, amv := range amvs {
		resp := CreateActivationVersionResponse(amv)
		result = append(result, resp)
	}
	return result
}

func GetAmv(tenantId, id string) *ActivationVersionResponse {
	amv := GetOneAmv(tenantId, id)
	if amv != nil {
		return CreateActivationVersionResponse(amv)
	}
	return nil
}

func GetAllAmvList(tenantId string) []*firmware.ActivationVersion {
	result := []*firmware.ActivationVersion{}
	list, err := firmware.GetFirmwareRuleAllAsListDBForAdmin(tenantId)
	if err != nil {
		log.Warn("no amv found")
		return result
	}
	for _, fwRule := range list {
		if fwRule.Type == coreef.ACTIVATION_VERSION {
			amv := coreef.ConvertIntoActivationVersion(fwRule)
			result = append(result, amv)
		}
	}
	return result
}

func GetOneAmv(tenantId string, id string) *firmware.ActivationVersion {
	inst, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_FIRMWARE_RULES, id)
	if err != nil {
		log.Warn("no amv found for " + id)
		return nil
	}
	fwRule := inst.(*firmware.FirmwareRule)
	return coreef.ConvertIntoActivationVersion(fwRule)
}

func validateUsageForAmv(tenantId, amvId string, app string) (string, error) {
	amv := GetOneAmv(tenantId, amvId)
	if amv == nil {
		return fmt.Sprintf("Entity with id  %s does not exist ", amvId), nil
	}
	if amv.ApplicationType != app {
		return fmt.Sprintf("Entity with id  %s does not exist ", amvId), nil
	}
	return "", nil
}

func DeleteAmvbyId(tenantId, id string, app string) *xwhttp.ResponseEntity {
	usage, err := validateUsageForAmv(tenantId, id, app)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, err, nil)
	}

	if usage != "" {
		return xwhttp.NewResponseEntity(http.StatusNotFound, errors.New(usage), nil)
	}

	err = DeleteOneAmv(tenantId, id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func DeleteOneAmv(tenantId, id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(tenantId, db.TABLE_FIRMWARE_RULES, id)
	if err != nil {
		return err
	}
	return nil
}

func ValidateModel(Id string) error {
	if len(strings.TrimSpace(Id)) > 0 {
		if match, _ := regexp.MatchString("^[-a-zA-Z0-9_.' ]+$", Id); match {
			return nil
		}
	}
	return errors.New("Model is invalid")
}

func GetSupportedVersionforModel(tenantId string, modelids []string, FirmwareVersions []string, app string) []string {
	supportedFwList := []string{}
	existedFwList := []string{}
	configs := GetFirmwareConfigsByModelIdsAndApplication(tenantId, modelids, app)
	for _, config := range configs {
		existedFwList = append(existedFwList, config.FirmwareVersion)
	}
	m := make(map[string]uint8)
	for _, k := range existedFwList {
		if m[k] == 0 {
			m[k] += 1
		}
	}
	for _, k := range FirmwareVersions {
		m[k] += 1
	}

	for k, v := range m {
		if v == 2 {
			supportedFwList = append(supportedFwList, k)
		}
	}

	return supportedFwList
}

func amvValidate(tenantId string, newamv *firmware.ActivationVersion) *xwhttp.ResponseEntity {
	if newamv == nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Activation minimum version should be specified"), nil)
	}
	if xwutil.IsBlank(newamv.ApplicationType) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("ApplicationType  is empty"), nil)
	}

	if xwutil.IsBlank(newamv.Description) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Description  is empty"), nil)
	}
	if xwutil.IsBlank(newamv.Model) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New(" Model is empty"), nil)
	}
	if err1 := ValidateModel(newamv.Model); err1 != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err1, nil)
	}
	newamv.Model = strings.ToUpper(newamv.Model)
	if existedModel := shared.GetOneModel(tenantId, newamv.Model); existedModel == nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New(" Model does not exist "), nil)
	}
	modelIds := []string{}
	modelIds = append(modelIds, newamv.Model)
	newamv.FirmwareVersions = GetSupportedVersionforModel(tenantId, modelIds, newamv.FirmwareVersions, newamv.ApplicationType)
	if xwutil.IsBlank(newamv.ID) {
		newamv.ID = uuid.New().String()
	}

	if !xwutil.IsBlank(newamv.PartnerId) {
		newamv.PartnerId = strings.TrimSpace(newamv.PartnerId)
		newamv.PartnerId = strings.ToUpper(newamv.PartnerId)
	}

	if len(newamv.RegularExpressions) == 0 && len(newamv.FirmwareVersions) == 0 {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("regex and firmwareversions both can't be empty Or Given firmware version is not supported for this model"), nil)
	}

	amvs := GetAllAmvList(tenantId)
	for _, examv := range amvs {
		if newamv.ID != examv.ID && newamv.Model == examv.Model && newamv.PartnerId == examv.PartnerId {
			return xwhttp.NewResponseEntity(http.StatusConflict, errors.New("ActivationVersion with the following model/partnerId already exists"), nil)
		}

		if newamv.ID != examv.ID && newamv.ApplicationType == examv.ApplicationType && newamv.Description == examv.Description {
			return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("activation versions with description already %s exists", examv.Description), nil)
		}
	}

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, nil)
}

func CreateAmv(tenantId string, amv *firmware.ActivationVersion, app string) *xwhttp.ResponseEntity {
	_, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_FIRMWARE_RULES, amv.ID)
	if err == nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s already exists", amv.ID), nil)
	}
	if xwutil.IsBlank(amv.ApplicationType) {
		amv.ApplicationType = app
	}

	if amv.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", amv.ID), nil)

	}
	respEntity := amvValidate(tenantId, amv)
	if respEntity.Error != nil {
		return respEntity
	}

	fwRule := coreef.ConvertIntoRule(amv)
	ru.NormalizeConditions(&fwRule.Rule)
	if err = firmware.CreateFirmwareRuleOneDB(tenantId, fwRule); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, amv)
}

func amvEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func UpdateAmvImport(tenantId string, amvToImport *firmware.ActivationVersion, amvinDB *firmware.ActivationVersion) *xwhttp.ResponseEntity {
	respEntity := amvValidate(tenantId, amvToImport)
	if respEntity.Error != nil {
		return respEntity
	}
	if amvToImport.ApplicationType != amvinDB.ApplicationType {
		amvinDB.ApplicationType = amvToImport.ApplicationType
	}
	if amvToImport.Description != amvinDB.Description {
		amvinDB.Description = amvToImport.Description
	}
	if amvToImport.Model != amvinDB.Model {
		amvinDB.Model = amvToImport.Model
	}
	if amvToImport.PartnerId != amvinDB.PartnerId {
		amvinDB.PartnerId = amvToImport.PartnerId
	}
	if len(amvToImport.RegularExpressions) == 0 && len(amvToImport.FirmwareVersions) == 0 {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("regex and firmware versions both can't be  empty"), nil)
	}
	if !amvEqual(amvToImport.RegularExpressions, amvinDB.RegularExpressions) {
		amvinDB.RegularExpressions = amvToImport.RegularExpressions
	}
	if !amvEqual(amvToImport.FirmwareVersions, amvinDB.FirmwareVersions) {
		amvinDB.FirmwareVersions = amvToImport.FirmwareVersions
	}
	fwRule := coreef.ConvertIntoRule(amvinDB)
	ru.NormalizeConditions(&fwRule.Rule)
	if err := firmware.CreateFirmwareRuleOneDB(tenantId, fwRule); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, amvToImport)
}

func importOrUpdateAllAmvs(tenantId string, entities []firmware.ActivationVersion, app string) (map[string][]string, error) {
	result := make(map[string][]string)
	result["NOT_IMPORTED"] = []string{}
	result["IMPORTED"] = []string{}

	for _, entity := range entities {
		entity := entity
		entityOnDb, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_FIRMWARE_RULES, entity.ID)
		if err != nil {
			respCreate := CreateAmv(tenantId, &entity, app)
			err = respCreate.Error
		} else {
			fwRule := entityOnDb.(*firmware.FirmwareRule)
			amvinDB := coreef.ConvertIntoActivationVersion(fwRule)
			if entity.ApplicationType != amvinDB.ApplicationType || amvinDB.ApplicationType != app {
				result["NOT_IMPORTED"] = append(result["NOT_IMPORTED"], entity.ID)
				continue
			}
			respUpdate := UpdateAmvImport(tenantId, &entity, amvinDB)
			err = respUpdate.Error
		}
		if err == nil {
			result["IMPORTED"] = append(result["IMPORTED"], entity.ID)
		} else {
			result["NOT_IMPORTED"] = append(result["NOT_IMPORTED"], entity.ID)
		}
	}
	return result, nil
}

func UpdateAmv(tenantId string, amv *firmware.ActivationVersion, app string) *xwhttp.ResponseEntity {
	if xwutil.IsBlank(amv.ID) {
		return xwhttp.NewResponseEntity(http.StatusNotFound, errors.New(" ID  is empty"), nil)
	}
	fwRule, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_FIRMWARE_RULES, amv.ID)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("Entity with id %s does not exist", amv.ID), nil)
	}
	fwRuleDB := fwRule.(*firmware.FirmwareRule)

	amvinDB := coreef.ConvertIntoActivationVersion(fwRuleDB)
	if amvinDB.ApplicationType != amv.ApplicationType || amvinDB.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("ApplicationType in db %s doesn't match the ApplicationType %s in req", amvinDB.ApplicationType, amv.ApplicationType), nil)
	}
	if respEntity := UpdateAmvImport(tenantId, amv, amvinDB); respEntity.Error != nil {
		return respEntity
	}
	return xwhttp.NewResponseEntity(http.StatusOK, nil, amv)
}

func AmvGeneratePage(list []*firmware.ActivationVersion, page int, pageSize int) (result []*firmware.ActivationVersion) {
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

func AmvGeneratePageWithContext(amvrules []*firmware.ActivationVersion, contextMap map[string]string) (result []*firmware.ActivationVersion, err error) {
	sort.Slice(amvrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(amvrules[i].Description), strings.ToLower(amvrules[j].Description)) < 0
	})
	pageNum := 1
	numStr, okval := contextMap[amvPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[amvPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, errors.New("pageNumber and pageSize should both be greater than zero")
	}
	return AmvGeneratePage(amvrules, pageNum, pageSize), nil
}

func AmvFilterByContext(searchContext map[string]string) []*firmware.ActivationVersion {
	var found bool
	amvRules := GetAllAmvList(searchContext[xwcommon.TENANT_ID])
	amvRuleList := []*firmware.ActivationVersion{}
	for _, amvRule := range amvRules {
		if amvRule == nil {
			continue
		}
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if amvRule.ApplicationType != applicationType && amvRule.ApplicationType != shared.ALL {
				continue
			}
		}
		if model, ok := xutil.FindEntryInContext(searchContext, amvModel, false); ok {
			if !strings.Contains(strings.ToLower(amvRule.Model), strings.ToLower(model)) {
				continue
			}
		}
		if partnerid, ok := xutil.FindEntryInContext(searchContext, amvPartnerId, false); ok {
			if !strings.Contains(strings.ToLower(amvRule.PartnerId), strings.ToLower(partnerid)) {
				continue
			}
		}
		if partnerid, ok := util.FindEntryInContext(searchContext, amvPartnerIdAlias, false); ok {
			if !strings.Contains(strings.ToLower(amvRule.PartnerId), strings.ToLower(partnerid)) {
				continue
			}
		}

		if desc, ok := xutil.FindEntryInContext(searchContext, amvDescription, false); ok {
			if !strings.Contains(strings.ToLower(amvRule.Description), strings.ToLower(desc)) {
				continue
			}
		}
		found = false
		if ver, ok := xutil.FindEntryInContext(searchContext, amvFwVersion, false); ok {
			filtver := strings.ToLower(ver)
			amvversions := amvRule.FirmwareVersions
			for _, version := range amvversions {
				version = strings.ToLower(version)
				if strings.Contains(version, filtver) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		found = false
		if ver, ok := xutil.FindEntryInContext(searchContext, amvFwVersionAlias, false); ok {
			filtver := strings.ToLower(ver)
			amvversions := amvRule.FirmwareVersions
			for _, version := range amvversions {
				version = strings.ToLower(version)
				if strings.Contains(version, filtver) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		found = false
		if regex, ok := xutil.FindEntryInContext(searchContext, amvRegex, false); ok {
			amvregexs := amvRule.RegularExpressions
			filtregx := strings.ToLower(regex)
			for _, regex := range amvregexs {
				regex = strings.ToLower(regex)
				if strings.Contains(regex, filtregx) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		found = false
		if regex, ok := xutil.FindEntryInContext(searchContext, amvRegexAlias, false); ok {
			amvregexs := amvRule.RegularExpressions
			filtregx := strings.ToLower(regex)
			for _, regex := range amvregexs {
				regex = strings.ToLower(regex)
				if strings.Contains(regex, filtregx) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		sort.Slice(amvRule.FirmwareVersions, func(i, j int) bool {
			return strings.Compare(
				strings.ToLower(amvRule.FirmwareVersions[i]),
				strings.ToLower(amvRule.FirmwareVersions[j])) < 0
		})

		amvRuleList = append(amvRuleList, amvRule)
	}
	return amvRuleList
}
