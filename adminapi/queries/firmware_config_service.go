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
	"fmt"
	"net/http"
	"strconv"
	"strings"

	xshared "github.com/rdkcentral/xconfadmin/shared"
	xutil "github.com/rdkcentral/xconfadmin/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	ru "github.com/rdkcentral/xconfwebconfig/rulesengine"

	xcommon "github.com/rdkcentral/xconfadmin/common"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	cFirmwareConfigApplicableActionType = xcommon.APPLICABLE_ACTION_TYPE
	cFirmwareConfigPageNumber           = xcommon.PAGE_NUMBER
	cFirmwareConfigPageSize             = xcommon.PAGE_SIZE
	cFirmwareConfigExistedVersions      = "existedVersions"
	cFirmwareConfigNotExistedVersions   = "notExistedVersions"
	cFirmwareConfigFirmwareVersion      = "FIRMWARE_VERSION" //xcommon.FIRMWARE_VERSION
	cFirmwareConfigDescription          = xwcommon.DESCRIPTION
	cFirmwareConfigModel                = xwcommon.MODEL
	cFirmwareConfigNumberOfItems        = "numberOfItems"
)

func GetFirmwareConfigs(applicationType string) []*coreef.FirmwareConfigResponse {
	result := []*coreef.FirmwareConfigResponse{}
	list, err := coreef.GetFirmwareConfigAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareConfigs: %v", err))
		return result
	}

	checkApplicationType := len(applicationType) > 0
	for _, fc := range list {
		if checkApplicationType && applicationType != fc.ApplicationType {
			continue
		}
		config := fc.CreateFirmwareConfigResponse()
		result = append(result, config)
	}
	return result
}

func GetFirmwareConfigsAS(applicationType string) []*coreef.FirmwareConfig {
	result := []*coreef.FirmwareConfig{}
	list, err := coreef.GetFirmwareConfigAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareConfigs: %v", err))
		return result
	}

	checkApplicationType := len(applicationType) > 0
	if !checkApplicationType {
		return list
	}
	for _, fc := range list {
		if applicationType != fc.ApplicationType {
			continue
		}
		result = append(result, fc)
	}
	return result
}

func GetFirmwareConfigById(id string) *coreef.FirmwareConfigResponse {
	fc, err := coreef.GetFirmwareConfigOneDB(id)
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareConfigById: %v", err))
		return nil
	}
	return fc.CreateFirmwareConfigResponse()
}

func GetFirmwareConfigByIdAS(id string) *coreef.FirmwareConfig {
	fc, err := coreef.GetFirmwareConfigOneDB(id)
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareConfigById: %v", err))
		return nil
	}
	return fc
}

func GetFirmwareConfigsByModelIdAndApplicationType(modelId string, applicationType string) []*coreef.FirmwareConfigResponse {
	result := []*coreef.FirmwareConfigResponse{}
	list, err := coreef.GetFirmwareConfigAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareConfigsByModelIdAndApplicationType: %v", err))
		return result
	}

	checkApplicationType := len(applicationType) > 0
	for _, fc := range list {
		if checkApplicationType && applicationType != fc.ApplicationType {
			continue
		}
		if !util.Contains(fc.SupportedModelIds, modelId) {
			continue
		}
		config := fc.CreateFirmwareConfigResponse()
		result = append(result, config)
	}
	return result
}

func GetFirmwareConfigsByModelIdAndApplicationTypeAS(modelId string, applicationType string) []*coreef.FirmwareConfig {
	result := []*coreef.FirmwareConfig{}
	list, err := coreef.GetFirmwareConfigAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareConfigsByModelIdAndApplicationType: %v", err))
		return result
	}

	checkApplicationType := len(applicationType) > 0
	for _, fc := range list {
		if checkApplicationType && applicationType != fc.ApplicationType {
			continue
		}
		if !util.Contains(fc.SupportedModelIds, modelId) {
			continue
		}
		result = append(result, fc)
	}
	return result
}

func IsValidFirmwareConfigByModelIds(modelId string, applicationType string, firmwareConfig *coreef.FirmwareConfig) bool {
	list, err := coreef.GetFirmwareConfigAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareConfigsByModelIdAndApplicationType: %v", err))
		return false
	}
	checkApplicationType := len(applicationType) > 0
	for _, fc := range list {
		if fc.ID == firmwareConfig.ID {
			return true
		}
		if checkApplicationType && applicationType != fc.ApplicationType {
			continue
		}
		if !util.Contains(fc.SupportedModelIds, modelId) {
			continue
		}
	}
	return false
}

func IsValidFirmwareConfigByModelIdList(modelIds *[]string, applicationType string, firmwareConfig *coreef.FirmwareConfig) bool {
	list, err := coreef.GetFirmwareConfigAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareConfigsByModelIdAndApplicationType: %v", err))
		return false
	}
	checkApplicationType := len(applicationType) > 0
	for _, fc := range list {
		found := false
		for _, modelId := range *modelIds {
			if util.Contains(fc.SupportedModelIds, modelId) {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		if checkApplicationType && applicationType != fc.ApplicationType {
			continue
		}
		if fc.ID == firmwareConfig.ID {
			return true
		}
	}
	return false
}

func extractFirmwareConfigPage(list []*coreef.FirmwareConfig, page int, pageSize int) (result []*coreef.FirmwareConfig) {
	leng := len(list)
	result = make([]*coreef.FirmwareConfig, 0)
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

func generateFirmwareConfigPageByContext(dbrules []*coreef.FirmwareConfig, contextMap map[string]string) (result []*coreef.FirmwareConfig, err error) {
	pageNum := 1
	numStr, okval := contextMap[cFirmwareConfigPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cFirmwareConfigPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "pageNumber and pageSize should both be greater than zero")

	}
	return extractFirmwareConfigPage(dbrules, pageNum, pageSize), nil
}

func filterFirmwareConfigsByContext(entries []*coreef.FirmwareConfig, searchContext map[string]string) (result []*coreef.FirmwareConfig, err error) {
	for _, config := range entries {
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, true); ok {
			if config.ApplicationType != applicationType && config.ApplicationType != shared.ALL {
				continue
			}
		}
		if model, ok := xutil.FindEntryInContext(searchContext, cFirmwareConfigModel, false); ok {
			if model != "" {
				found := false
				for _, elem := range config.SupportedModelIds {
					if strings.Contains(strings.ToLower(elem), strings.ToLower(model)) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
		}
		if fw_version, ok := xutil.FindEntryInContext(searchContext, cFirmwareConfigFirmwareVersion, false); ok {
			if !strings.Contains(strings.ToLower(config.FirmwareVersion), strings.ToLower(fw_version)) {
				continue
			}
		}
		if description, ok := xutil.FindEntryInContext(searchContext, cFirmwareConfigDescription, false); ok {
			if !strings.Contains(strings.ToLower(config.Description), strings.ToLower(description)) {
				continue
			}
		}
		result = append(result, config)
	}
	return result, nil
}

func beforeCreatingFirmwareConfig(entity *coreef.FirmwareConfig, writeApplication string) error {
	if util.IsBlank(entity.ID) {
		entity.ID = uuid.New().String()
	} else {
		if util.IsBlank(entity.ApplicationType) {
			entity.ApplicationType = writeApplication
		} else if entity.ApplicationType != writeApplication {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "ApplicationType conflict")
		}
		entity.Updated = util.GetTimestamp()
		existingEntity, _ := coreef.GetFirmwareConfigOneDB(entity.ID)

		if existingEntity != nil {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Entity with id: "+entity.ID+" already exists in "+existingEntity.ApplicationType+" application")
		}
	}
	return nil
}

func CreateFirmwareConfigAS(config *coreef.FirmwareConfig, appType string, validateName bool) *xwhttp.ResponseEntity {
	for i, id := range config.SupportedModelIds {
		config.SupportedModelIds[i] = strings.ToUpper(id)
	}

	err := config.Validate()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}
	if validateName {
		err = config.ValidateName()
		if err != xwcommon.NotFound && err != nil {
			return xwhttp.NewResponseEntity(http.StatusConflict, err, nil)
		}
	}
	if err = beforeCreatingFirmwareConfig(config, appType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, err, nil)
	}

	err = coreef.CreateFirmwareConfigOneDB(config)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, config)
}

func CreateFirmwareConfig(config *coreef.FirmwareConfig, appType string) *xwhttp.ResponseEntity {
	if err := CreateFirmwareConfigAS(config, appType, true); err != nil {
		return err
	}
	resp := config.CreateFirmwareConfigResponse()
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, resp)
}

func beforeUpdatingFirmwareConfig(entity *coreef.FirmwareConfig, writeApplication string) error {
	if util.IsBlank(entity.ID) {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Entity id is empty: ")

	}
	entity.Updated = util.GetTimestamp()
	if util.IsBlank(entity.ApplicationType) {
		if !util.IsBlank(writeApplication) {
			entity.ApplicationType = writeApplication
		}
	}
	existingEntity, _ := coreef.GetFirmwareConfigOneDB(entity.ID)

	if existingEntity == nil || existingEntity.ApplicationType != entity.ApplicationType {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id: "+entity.ID+" does not exist in "+existingEntity.ApplicationType+" application")
	}
	return nil
}

func UpdateFirmwareConfigAS(config *coreef.FirmwareConfig, appType string, validateName bool) *xwhttp.ResponseEntity {
	for i, id := range config.SupportedModelIds {
		config.SupportedModelIds[i] = strings.ToUpper(id)
	}

	err := config.Validate()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if GetFirmwareConfigById(config.ID) == nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("\"FirmwareConfig with current id: %s does not exist\"", config.ID), nil)
	}

	if validateName {
		err = config.ValidateName()
		if err != nil {
			return xwhttp.NewResponseEntity(http.StatusConflict, err, nil)
		}
	}

	if err = beforeUpdatingFirmwareConfig(config, appType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, err, nil)
	}

	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_FIRMWARE_CONFIG, config.ID, config)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}
	return xwhttp.NewResponseEntity(http.StatusOK, nil, config)
}

func UpdateFirmwareConfig(config *coreef.FirmwareConfig, appType string) *xwhttp.ResponseEntity {
	if err := UpdateFirmwareConfigAS(config, appType, true); err != nil {
		return err
	}
	resp := config.CreateFirmwareConfigResponse()
	return xwhttp.NewResponseEntity(http.StatusOK, nil, resp)
}

func beforeDeletingFirmwareConfig(id string, appType string) *xwhttp.ResponseEntity {
	entity, _ := coreef.GetFirmwareConfigOneDB(id)
	if entity == nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("Entity with id: %s does not exist", id), nil)
	}

	err := beforeUpdatingFirmwareConfig(entity, appType)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, err, nil)
	}

	// Check for usage in FirmwareRule
	amvs := GetAllAmvList()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, fmt.Errorf("ActivationVersion check Referential Integrity while deleting %s failed", id), nil)
	}

	for _, amv := range amvs {
		if searchList(amv.FirmwareVersions, entity.FirmwareVersion, true) {
			return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("FirmwareConfig %s is used by AMV %s", entity.Description, amv.Description), nil)

		}
	}

	// Check for usage in FirmwareRule
	rules, err := corefw.GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil && err.Error() != xcommon.NotFound.Error() {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, fmt.Errorf("Get FirmwareRules to check Referential Integrity while deleting %s failed", id), nil)
	}
	for _, rule := range rules {
		if rule.ConfigId() == entity.ID {
			return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("FirmwareConfig %s is used by %s rule", entity.Description, rule.Name), nil)
		}
		if id == rule.ApplicableAction.IntermediateVersion || searchList(rule.ApplicableAction.GetFirmwareVersions(), entity.FirmwareVersion, true) {
			return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("FirmwareConfig %s is used by %s rule", entity.Description, rule.Name), nil)
		}
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, entity)
}

func DeleteFirmwareConfig(id string, appType string) *xwhttp.ResponseEntity {
	err := beforeDeletingFirmwareConfig(id, appType)
	if err.Error != nil {
		return err
	}
	err2 := coreef.DeleteOneFirmwareConfig(id)
	if err2 != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err2, nil)
	}
	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func GetFirmwareConfigId(version string, applicationType string) string {
	list, err := coreef.GetFirmwareConfigAsListDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareConfigId: %v", err))
		return ""
	}

	for _, fc := range list {
		if version == fc.FirmwareVersion && xshared.ApplicationTypeEquals(applicationType, fc.ApplicationType) {
			return fc.ID
		}
	}

	return ""
}

func GetFirmwareConfigsByModelIdsAndApplication(modelIds []string, applicationType string) []coreef.FirmwareConfig {
	result := []coreef.FirmwareConfig{}
	entries, _ := coreef.GetFirmwareConfigAsListDB()
	for _, entry := range entries {
		if applicationType == entry.ApplicationType && hasCommonEntries(modelIds, entry.SupportedModelIds) {
			result = append(result, *entry)
		}
	}
	return result
}

func containsVersion(configs []coreef.FirmwareConfig, version string) bool {
	for _, config := range configs {
		if config.FirmwareVersion == version {
			return true
		}
	}
	return false
}

func GetSortedFirmwareVersionsIfDoesExistOrNot(firmwareConfigData FirmwareConfigData, applicationType string) map[string][]string {
	firmwareVersionMap := make(map[string][]string)

	if len(firmwareConfigData.Versions) == 0 || len(firmwareConfigData.ModelSet) == 0 {
		return firmwareVersionMap
	}
	firmwareConfigsByModel := GetFirmwareConfigsByModelIdsAndApplication(firmwareConfigData.ModelSet, applicationType)
	existedVersions := []string{}
	notExistedVersions := []string{}
	for _, firmwareVersion := range firmwareConfigData.Versions {
		if searchList(existedVersions, firmwareVersion, false) || searchList(notExistedVersions, firmwareVersion, false) {
			continue
		}
		if containsVersion(firmwareConfigsByModel, firmwareVersion) {
			existedVersions = append(existedVersions, firmwareVersion)
		} else {
			notExistedVersions = append(notExistedVersions, firmwareVersion)
		}
	}
	firmwareVersionMap[cFirmwareConfigExistedVersions] = existedVersions
	firmwareVersionMap[cFirmwareConfigNotExistedVersions] = notExistedVersions
	return firmwareVersionMap
}

func getSupportedConfigsByEnvModelRuleName(envModelName string, appType string) []coreef.FirmwareConfig {
	versionSet := []coreef.FirmwareConfig{}
	model := ""
	firmwareRules, _ := corefw.GetFirmwareRuleAllAsListDBForAdmin()
	for _, rule := range firmwareRules {
		if rule.Type == corefw.ENV_MODEL_RULE && rule.Name == envModelName && rule.ApplicationType == appType {
			model = extractModel(*rule)
			break
		}
	}
	if model == "" {
		return versionSet
	}

	configs, _ := coreef.GetFirmwareConfigAsListDB()
	for _, config := range configs {
		if config.ApplicationType != appType {
			continue
		}
		supportedModels := config.SupportedModelIds
		if len(supportedModels) == 0 {
			continue
		}
		if searchList(supportedModels, model, true) {
			versionSet = append(versionSet, *config)
		}
	}
	return versionSet
}

func getFirmwareConfigByEnvModelRuleName(envModelRuleName string) *coreef.FirmwareConfig {
	firmwareRules, _ := corefw.GetFirmwareRuleAllAsListDBForAdmin()
	for _, rule := range firmwareRules {
		if rule.Type == corefw.ENV_MODEL_RULE && rule.Name == envModelRuleName && rule.ApplicableAction.Type == ".RuleAction" { // TODO Not sure what instanceof means in GO
			fc, err := coreef.GetFirmwareConfigOneDB(rule.ConfigId())
			if err == nil {
				return fc
			}
		}
	}
	return nil
}

func extractModel(rule corefw.FirmwareRule) string {
	for _, condition := range ru.ToConditions(&rule.Rule) {
		if condition.FreeArg.Equals(coreef.RuleFactoryMODEL) {
			return strings.Trim(condition.FixedArg.String(), "'")
		}
	}
	return ""
}
