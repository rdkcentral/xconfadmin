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
	"math"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"xconfadmin/common"
	xhttp "xconfadmin/http"
	xshared "xconfadmin/shared"
	"xconfadmin/util"
	xcommon "xconfwebconfig/common"
	xwcommon "xconfwebconfig/common"
	xwhttp "xconfwebconfig/http"
	re "xconfwebconfig/rulesengine"
	ru "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	coreef "xconfwebconfig/shared/estbfirmware"
	"xconfwebconfig/shared/firmware"
	xutil "xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	PERCENTAGE_FIELD_NAME = "percentage"
	WHITELIST_FIELD_NAME  = "whitelist"
	PARTNER_ID            = "partnerId"
)

var canaryNameRegex = regexp.MustCompile(`[^-a-zA-Z0-9_.' ]+`)

// Service APIs for Percent Filter Rule

func GetOnePercentageBeanFromDB(id string) (*coreef.PercentageBean, error) {
	frule, err := firmware.GetFirmwareRuleOneDB(id)
	if err != nil {
		return nil, err
	}

	bean := coreef.ConvertFirmwareRuleToPercentageBean(frule)
	return bean, nil
}

func GetAllGlobalPercentageBeansAsRuleFromDB(applicationType string, sortByName bool) ([]*firmware.FirmwareRule, error) {
	frules, err := firmware.GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil {
		return nil, err
	}

	var result []*firmware.FirmwareRule

	for _, frule := range frules {
		if frule.ApplicationType == applicationType && frule.Type == firmware.ENV_MODEL_RULE {
			result = append(result, frule)
		}
	}

	if sortByName {
		sort.Slice(result, func(i, j int) bool {
			return strings.Compare(strings.ToLower(result[i].Name), strings.ToLower(result[j].Name)) < 0
		})
	}

	return result, nil
}

func GetAllPercentageBeansFromDB(applicationType string, sortByName bool, convert bool) ([]*coreef.PercentageBean, error) {
	firmwareRules, err := firmware.GetFirmwareRuleAllAsListDBForAdmin()
	if err != nil {
		return nil, err
	}

	result := []*coreef.PercentageBean{}

	for _, frule := range firmwareRules {
		if frule.ApplicationType == applicationType && frule.Type == firmware.ENV_MODEL_RULE {
			bean := coreef.ConvertFirmwareRuleToPercentageBean(frule)
			if convert {
				replaceFieldsWithFirmwareVersion(bean)
			}
			result = append(result, bean)
		}
	}

	if sortByName {
		sort.Slice(result, func(i, j int) bool {
			return strings.Compare(strings.ToLower(result[i].Name), strings.ToLower(result[j].Name)) < 0
		})
	}

	return result, nil
}

func GetPercentageBeanFilterFieldValues(fieldName string, applicationType string) (map[string][]interface{}, error) {
	fieldValues, err := getPercentageBeanFieldValues(fieldName, applicationType)
	if err != nil {
		return nil, err
	}

	globalFieldValues := getGlobalPercentageFields(fieldName, applicationType)
	for fieldValue := range globalFieldValues {
		fieldValues[fieldValue] = struct{}{}
	}

	resultFieldValues := make([]interface{}, 0)
	for fieldValue := range fieldValues {
		resultFieldValues = append(resultFieldValues, fieldValue)
	}

	result := make(map[string][]interface{})
	result[fieldName] = resultFieldValues
	return result, nil
}

func getGlobalPercentageFields(fieldName string, applicationType string) map[interface{}]struct{} {
	resultFieldValues := make(map[interface{}]struct{})

	globalPercentageId := GetGlobalPercentageIdByApplication(applicationType)
	globalPercentageRule, err := firmware.GetFirmwareRuleOneDB(globalPercentageId)
	if err != nil {
		log.Error(fmt.Sprintf("GetGlobalPercentageFields: %v", err))
		if fieldName == PERCENTAGE_FIELD_NAME {
			resultFieldValues[100] = struct{}{}
		}
		return resultFieldValues
	}

	globalPercentage := coreef.ConvertIntoGlobalPercentageFirmwareRule(globalPercentageRule)
	fieldValues := GetStructFieldValues(fieldName, reflect.ValueOf(*globalPercentage))
	for _, fieldValue := range fieldValues {
		resultFieldValues[fieldValue] = struct{}{}
	}

	return resultFieldValues
}

func getPercentageBeanFieldValues(fieldName string, applicationType string) (map[interface{}]struct{}, error) {
	resultFieldValues := make(map[interface{}]struct{})

	beans, err := GetAllPercentageBeansFromDB(applicationType, false, true)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(fieldName, "distributions") {
		configs := make(map[string]*firmware.ConfigEntry)
		for _, bean := range beans {
			for _, configEntry := range bean.Distributions {
				configs[configEntry.ConfigId] = configEntry
			}
		}
		for _, configEntry := range configs {
			resultFieldValues[configEntry] = struct{}{}
		}
	} else {
		for _, bean := range beans {
			fieldValues := GetStructFieldValues(fieldName, reflect.ValueOf(*bean))
			for _, fieldValue := range fieldValues {
				resultFieldValues[fieldValue] = struct{}{}
			}
		}
	}
	return resultFieldValues, nil
}

func GetGlobalPercentageIdByApplication(applicationType string) string {
	if xshared.ApplicationTypeEquals(applicationType, shared.STB) {
		return firmware.GLOBAL_PERCENT
	}
	return fmt.Sprintf("%s_%s", strings.ToUpper(applicationType), firmware.GLOBAL_PERCENT)
}

func GetStructFieldValues(fieldName string, structValue reflect.Value) []interface{} {
	var resultFieldValues []interface{}

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Type().Field(i)
		value := structValue.Field(i).Interface()
		if strings.EqualFold(fieldName, field.Name) {
			switch field.Type.Kind() {
			case reflect.Bool, reflect.Float32, reflect.Float64, reflect.Ptr:
				resultFieldValues = append(resultFieldValues, value)
			case reflect.String:
				if str, ok := value.(string); ok && str != "" {
					resultFieldValues = append(resultFieldValues, str)
				}
			case reflect.Slice:
				if reflect.TypeOf(value).Elem().Kind() == reflect.String {
					for _, str := range value.([]string) {
						resultFieldValues = append(resultFieldValues, str)
					}
				}
			}
			break
		}
	}

	return resultFieldValues
}

func CreatePercentageBean(bean *coreef.PercentageBean, applicationType string, fields log.Fields) *xwhttp.ResponseEntity {
	_, err := firmware.GetFirmwareRuleOneDB(bean.ID)
	if err == nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s Already Exist", bean.ID), nil)
	}

	if applicationType != bean.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id %s ApplicationType doesn't match", bean.ID), nil)
	}

	if err := firmware.ValidateRuleName(bean.ID, bean.Name, applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if err := bean.ValidateForAS(); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	beans, err := GetAllPercentageBeansFromDB(bean.ApplicationType, false, true)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if err := bean.ValidateAll(beans); err != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, err, nil)
	}

	firmware.SortConfigEntry(bean.Distributions)

	fRule := coreef.ConvertPercentageBeanToFirmwareRule(*bean)
	ru.NormalizeConditions(&fRule.Rule)
	if err := firmware.CreateFirmwareRuleOneDB(fRule); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	newBean := coreef.ConvertFirmwareRuleToPercentageBean(fRule)
	createCanaries(newBean, nil, fields)
	return xwhttp.NewResponseEntity(http.StatusCreated, nil, newBean)
}

func UpdatePercentageBean(bean *coreef.PercentageBean, applicationType string, fields log.Fields) *xwhttp.ResponseEntity {
	if xutil.IsBlank(bean.ID) {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Entity id is empty"), nil)
	}

	fRule, err := firmware.GetFirmwareRuleOneDB(bean.ID)
	if fRule == nil || err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Entity with id: %s does not exist", bean.ID), nil)
	}
	if fRule.ApplicationType != applicationType {
		return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("Entity with id: %s ApplicationType  Mismatch", bean.ID), nil)
	}
	if fRule.ApplicationType != bean.ApplicationType {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("ApplicationType cannot be changed: Existing value:"+fRule.ApplicationType+" New Value:"+bean.ApplicationType), nil)
	}

	if err := firmware.ValidateRuleName(bean.ID, bean.Name, applicationType); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if err := bean.ValidateForAS(); err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	beans, err := GetAllPercentageBeansFromDB(bean.ApplicationType, false, true)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if err := bean.ValidateAll(beans); err != nil {
		return xwhttp.NewResponseEntity(http.StatusConflict, err, nil)
	}

	firmware.SortConfigEntry(bean.Distributions)

	newRule := coreef.ConvertPercentageBeanToFirmwareRule(*bean)
	ru.NormalizeConditions(&newRule.Rule)
	if err := firmware.CreateFirmwareRuleOneDB(newRule); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	newBean := coreef.ConvertFirmwareRuleToPercentageBean(newRule)
	createCanaries(newBean, fRule, fields)
	return xwhttp.NewResponseEntity(http.StatusOK, nil, newBean)
}

func DeletePercentageBean(id string, app string) *xwhttp.ResponseEntity {
	fRule, err := firmware.GetFirmwareRuleOneDB(id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("Entity with id: %s does not exist", id), nil)
	}
	if fRule.ApplicationType != app {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("Entity with id: %s ApplicationType doesn't match", id), nil)
	}
	if err = firmware.DeleteOneFirmwareRule(fRule.ID); err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
}

func createCanaries(newBean *coreef.PercentageBean, oldRule *firmware.FirmwareRule, fields log.Fields) {
	fields["canaryPercentFilterName"] = newBean.Name
	tfields := xwcommon.CopyLogFields(fields) // used only for the CreateCanary call
	fields = xwcommon.FilterLogFields(fields)
	canaryCreationTime := time.Now().Unix()
	if !common.CanaryCreationEnabled {
		log.WithFields(fields).Infof("Canary creation flag is disabled, so no canary is created")
		return
	}
	deviceType := "BROADBAND" // default to broadband
	if isVideoDevice(newBean.Model) {
		if !common.VideoCanaryCreationEnabled {
			log.WithFields(fields).Infof("Canary creation flag is disabled for video, so no canary is created")
			return
		}
		deviceType = "VIDEO"
	}
	if !isWithinCanaryWindow(fields) {
		log.WithFields(fields).Infof("Not within canary window, so no canary is created")
		return
	}
	if len(common.CanaryPercentFilterNameSet) > 0 && !common.CanaryPercentFilterNameSet.Contains(strings.ToLower(newBean.Name)) {
		log.WithFields(fields).Infof("PercentFilter name doesn't match list in config: percentFilterName=%s, so no canary is created", newBean.Name)
		return
	}
	// list of canaries to be created
	canaryConfigList := []firmware.ConfigEntry{}
	// list of distribution entries that don't fit into "easy matches"
	specialCasesConfigEntryList := []firmware.ConfigEntry{}
	// list to keep track of which of the old distributions have been matched to a new one (starts all to false) only if there is an old rule (oldRule != nil)
	var oldDistributionsMatched []bool
	if oldRule != nil {
		oldDistributionsMatched = make([]bool, len(oldRule.ApplicableAction.ConfigEntries))
	}
	for _, newConfigEntry := range newBean.Distributions {
		if newConfigEntry.IsCanaryDisabled {
			log.WithFields(fields).Infof("Firmware Config is disabled for canary, not adding distribution to canary list, configId=%s, startPercent=%f, endPercent=%f", newConfigEntry.ConfigId, newConfigEntry.StartPercentRange, newConfigEntry.EndPercentRange)
			continue
		}
		// flag to keep track if we found a match to compare to, starts out false
		hasDistributionMatch := false
		// if configId in distribution matches LKG, we don't want to create a canary
		if newBean.LastKnownGood == newConfigEntry.ConfigId {
			hasDistributionMatch = true
			log.WithFields(fields).Infof("Firmware Config is already the LKG used in Percent Filter, not adding distribution to canary list, configId=%s, startPercent=%f, endPercent=%f", newConfigEntry.ConfigId, newConfigEntry.StartPercentRange, newConfigEntry.EndPercentRange)
			continue
		}
		// flag to keep track of whether the firmwareVersion in current distribution is present in any of the old distributions, starts out false
		existingFirmwareVersion := false
		// if oldRule is nil, this is a new Percent Filter, so every distribution is a new firmware version and can have a canary created
		if oldRule != nil {
			for i, oldConfigEntry := range oldRule.ApplicableAction.ConfigEntries {
				// check if firmwareVersion is the same
				if newConfigEntry.ConfigId == oldConfigEntry.ConfigId {
					existingFirmwareVersion = true
					// only compare to old distributions that haven't matched a new one already
					if !oldDistributionsMatched[i] {
						// check if start AND end percent are the same, in which case don't create canary
						if areFloatsEqualEnough(newConfigEntry.StartPercentRange, oldConfigEntry.StartPercentRange) && areFloatsEqualEnough(newConfigEntry.EndPercentRange, oldConfigEntry.EndPercentRange) {
							oldDistributionsMatched[i] = true
							hasDistributionMatch = true
							log.WithFields(fields).Infof("Firmware Config already used in Percent Filter with same start and end percent, not adding distribution to canary list, configId=%s, startPercent=%f, endPercent=%f", newConfigEntry.ConfigId, newConfigEntry.StartPercentRange, newConfigEntry.EndPercentRange)
							break
						}
						// check if start is the same
						if areFloatsEqualEnough(newConfigEntry.StartPercentRange, oldConfigEntry.StartPercentRange) {
							oldDistributionsMatched[i] = true
							hasDistributionMatch = true
							// check if canary got smaller, in which case don't create canary
							if newConfigEntry.EndPercentRange < oldConfigEntry.EndPercentRange {
								log.WithFields(fields).Infof("Firmware Config already used in Percent Filter with same start and percent range got smaller, not adding distribution to canary list, configId=%s, startPercent=%f, oldEndPercent=%f, newEndPercent=%f", newConfigEntry.ConfigId, newConfigEntry.StartPercentRange, oldConfigEntry.EndPercentRange, newConfigEntry.EndPercentRange)
								break
							}
							// else canary got bigger, create canary for new percent
							canaryConfigEntry := *firmware.NewConfigEntry(newConfigEntry.ConfigId, oldConfigEntry.EndPercentRange, newConfigEntry.EndPercentRange)
							canaryConfigList = append(canaryConfigList, canaryConfigEntry)
							log.WithFields(fields).Infof("Firmware Config already used in Percent Filter with same start but percent range got larger, adding distribution to canary list using partial percent range, configId=%s, canaryStartPercent=%f, [oldEndPercent=%f, canaryEndPercent=%f]", canaryConfigEntry.ConfigId, canaryConfigEntry.StartPercentRange, oldConfigEntry.EndPercentRange, canaryConfigEntry.EndPercentRange)
							break
						}
						// check if end is the same
						if areFloatsEqualEnough(newConfigEntry.EndPercentRange, oldConfigEntry.EndPercentRange) {
							oldDistributionsMatched[i] = true
							hasDistributionMatch = true
							// check if canary got smaller, in which case don't create canary
							if newConfigEntry.StartPercentRange > oldConfigEntry.StartPercentRange {
								log.WithFields(fields).Infof("Firmware Config already used in Percent Filter with same end and percent range got smaller, not creating canary, configId=%s, oldStartPercent=%f, newStartPercent=%f, endPercent=%f", newConfigEntry.ConfigId, oldConfigEntry.StartPercentRange, newConfigEntry.StartPercentRange, newConfigEntry.EndPercentRange)
								break
							}
							// else canary got bigger, create canary for new percent
							canaryConfigEntry := *firmware.NewConfigEntry(newConfigEntry.ConfigId, newConfigEntry.StartPercentRange, oldConfigEntry.StartPercentRange)
							canaryConfigList = append(canaryConfigList, canaryConfigEntry)
							log.WithFields(fields).Infof("Firmware Config already used in Percent Filter with same end but percent range got larger, adding distribution to canary list using partial percent range, configId=%s, [canaryStartPercent=%f, oldStartPercent=%f], canaryEndPercent=%f", canaryConfigEntry.ConfigId, canaryConfigEntry.StartPercentRange, oldConfigEntry.StartPercentRange, canaryConfigEntry.EndPercentRange)
							break
						}
					}
				}
			}
		}
		// check that this distribution didn't already match something above
		if !hasDistributionMatch {
			// after we've checked new config against ALL old configs (or this is a new Percent Filter and there are no old configs), check if the firmwareVersion was already existing in old Percent Filter
			if !existingFirmwareVersion {
				// this is a new firmwareVersion so should create a canary on full range
				canaryConfigEntry := *firmware.NewConfigEntry(newConfigEntry.ConfigId, newConfigEntry.StartPercentRange, newConfigEntry.EndPercentRange)
				canaryConfigList = append(canaryConfigList, canaryConfigEntry)
				log.WithFields(fields).Infof("New Firmware Config detected with new firmwareVersion, creating canary on full percent range, configId=%s, [startPercentRange=%f, endPercentRange=%f]", canaryConfigEntry.ConfigId, canaryConfigEntry.StartPercentRange, canaryConfigEntry.EndPercentRange)
			} else {
				// if we got here, it's an existing firmware that doesn't share start or end with an existing distribution, add to "special cases list" and continue whittling down the rest of the lists
				specialCasesConfigEntryList = append(specialCasesConfigEntryList, *newConfigEntry)
			}
		}

	}
	// we now have a list of distributions that didn't fit the "easy matches", as well as a list that shows which old distributions have/haven't been matched to a new one
	for _, newConfigEntry := range specialCasesConfigEntryList {
		hasDistributionMatch := false
		for i, oldConfigEntry := range oldRule.ApplicableAction.ConfigEntries {
			if newConfigEntry.ConfigId == oldConfigEntry.ConfigId && !oldDistributionsMatched[i] {
				// 4 cases for overlap, only 2 cases for no overlap, so check for no overlap and use ! (no overlap means the start of one of the distributions is greater than the end of the other)
				if !(newConfigEntry.EndPercentRange < oldConfigEntry.StartPercentRange || (oldConfigEntry.EndPercentRange < newConfigEntry.StartPercentRange)) {
					oldDistributionsMatched[i] = true
					hasDistributionMatch = true
					// check if new config fits entirely in old config
					if newConfigEntry.StartPercentRange > oldConfigEntry.StartPercentRange && newConfigEntry.EndPercentRange < oldConfigEntry.EndPercentRange {
						log.WithFields(fields).Infof("Firmware Config already used in Percent Filter and distribution window fits entirely in old window, not adding distribution to canary list, configId=%s, oldStartPercent=%f, newStartPercent=%f, oldEndPercent=%f, newEndPercent=%f", newConfigEntry.ConfigId, oldConfigEntry.StartPercentRange, newConfigEntry.StartPercentRange, oldConfigEntry.EndPercentRange, newConfigEntry.EndPercentRange)
						break
					}
					// check if only the end of the distribution expanded
					// we could have added newConfigEntry.StartPercentRange < oldConfigEntry.EndPercentRange to the below condition for a more clarity
					if newConfigEntry.StartPercentRange > oldConfigEntry.StartPercentRange && newConfigEntry.EndPercentRange > oldConfigEntry.EndPercentRange {
						canaryConfigEntry := *firmware.NewConfigEntry(newConfigEntry.ConfigId, oldConfigEntry.EndPercentRange, newConfigEntry.EndPercentRange)
						canaryConfigList = append(canaryConfigList, canaryConfigEntry)
						log.WithFields(fields).Infof("Firmware Config already used in Percent Filter and distribution window only expanded on the end, adding distribution to canary list using partial percent range, configId=%s, canaryStartPercent=%f, [oldEndPercent=%f, newEndPercent=%f]", canaryConfigEntry.ConfigId, canaryConfigEntry.StartPercentRange, oldConfigEntry.EndPercentRange, canaryConfigEntry.EndPercentRange)
						break
					}
					// check if only the start of the distribution expanded
					// we could have added newConfigEntry.EndPercentRange > oldConfigEntry.StartPercentRange to the below condition for a more clarity
					if newConfigEntry.StartPercentRange < oldConfigEntry.StartPercentRange && newConfigEntry.EndPercentRange < oldConfigEntry.EndPercentRange {
						canaryConfigEntry := *firmware.NewConfigEntry(newConfigEntry.ConfigId, newConfigEntry.StartPercentRange, oldConfigEntry.StartPercentRange)
						canaryConfigList = append(canaryConfigList, canaryConfigEntry)
						log.WithFields(fields).Infof("Firmware Config already used in Percent Filter and distribution window only expanded on the start, adding distribution to canary list using partial percent range, configId=%s, [newStartPercent=%f, oldStartPercent=%f], canaryEndPercent=%f", canaryConfigEntry.ConfigId, canaryConfigEntry.StartPercentRange, oldConfigEntry.StartPercentRange, canaryConfigEntry.EndPercentRange)
						break
					}
					// else must have expanded on both sides of the distribution
					canaryConfigEntry := *firmware.NewConfigEntry(newConfigEntry.ConfigId, newConfigEntry.StartPercentRange, oldConfigEntry.StartPercentRange)
					canaryConfigList = append(canaryConfigList, canaryConfigEntry)
					canaryConfigEntry2 := *firmware.NewConfigEntry(newConfigEntry.ConfigId, oldConfigEntry.EndPercentRange, newConfigEntry.EndPercentRange)
					canaryConfigList = append(canaryConfigList, canaryConfigEntry2)
					log.WithFields(fields).Infof("Firmware Config already used in Percent Filter and distribution window expanded on both sides of distribution, adding two distributions to canary list using partial percent range, configId=%s, [newStartPercent=%f, oldStartPercent=%f], [oldEndPercent=%f, newEndPercent=%f]", canaryConfigEntry.ConfigId, newConfigEntry.StartPercentRange, oldConfigEntry.StartPercentRange, oldConfigEntry.EndPercentRange, newConfigEntry.EndPercentRange)
					break
				}
			}
		}
		if !hasDistributionMatch {
			// if we get here, the distribution doesn't overlap so create a canary over full distribution
			canaryConfigEntry := *firmware.NewConfigEntry(newConfigEntry.ConfigId, newConfigEntry.StartPercentRange, newConfigEntry.EndPercentRange)
			canaryConfigList = append(canaryConfigList, canaryConfigEntry)
			log.WithFields(fields).Infof("New Firmware Config already used in Percent Filter but doesn't overlap with an old distribution, creating canary on full percent range, configId=%s, [newStartPercent=%f, newEndPercent=%f]", newConfigEntry.ConfigId, newConfigEntry.StartPercentRange, newConfigEntry.EndPercentRange)
		}
	}

	// loop through canary list to create canaries in XDAS
	size := common.GetIntAppSetting(common.PROP_CANARY_MAXSIZE, common.CanarySize)
	distPercentage := common.GetFloat64AppSetting(common.PROP_CANARY_DISTRIBUTION_PERCENTAGE, common.CanaryDistributionPercentage)
	for _, canaryConfigEntry := range canaryConfigList {
		go func(canaryConfigEntry firmware.ConfigEntry) {
			partnerId, err := getPartnerOptionalCondition(newBean)
			if err != nil {
				log.WithFields(fields).Errorf("Error getting partnerId from optional condition, err=%+v", err)
			} else {
				firmwareConfig, err := coreef.GetFirmwareConfigOneDB(canaryConfigEntry.ConfigId)
				if err != nil {
					log.WithFields(fields).Errorf("Error looking up firmware config in DB, configId=%s", canaryConfigEntry.ConfigId)
				} else {
					canaryGroupName := fmt.Sprintf("%s_%s_%d_%d_%+v", newBean.Name, firmwareConfig.Description, int(100*canaryConfigEntry.StartPercentRange), int(100*canaryConfigEntry.EndPercentRange), canaryCreationTime)

					timeZoneList := common.CanaryTimezoneList

					if common.CanarySyndicatePartnerSet.Contains(partnerId) {
						partnerTimezoneStr := common.GetStringAppSetting(common.PROP_CANARY_TIMEZONE_LIST + "_" + partnerId)
						if partnerTimezoneStr != "" {
							timeZoneList = strings.Split(partnerTimezoneStr, ",")
						}
					}
					canaryGroupName = canaryNameRegex.ReplaceAllString(canaryGroupName, "")
					canaryRequest := &xhttp.CanaryRequestBody{
						Name:                   canaryGroupName,
						DeviceType:             deviceType,
						Size:                   size,
						DistributionPercentage: distPercentage,
						Partner:                partnerId,
						Model:                  newBean.Model,
						TimeZones:              timeZoneList,
						StartPercentRange:      canaryConfigEntry.StartPercentRange,
						EndPercentRange:        canaryConfigEntry.EndPercentRange,
					}
					if oldRule != nil {
						// specify FwAppliedRule only for an existing rule
						canaryRequest.FwAppliedRule = oldRule.Name
					}
					log.WithFields(fields).Infof("Creating canary, configId=%s, canaryGroupName=%s", canaryConfigEntry.ConfigId, canaryGroupName)
					if err := xhttp.WebConfServer.CanaryMgrConnector.CreateCanary(canaryRequest, tfields); err != nil {
						log.WithFields(fields).Errorf("Error calling canarymgr to create canary, canaryGroupName=%s, err=%+v", canaryGroupName, err)
					} else {
						log.WithFields(fields).Infof("Successfully called canarymgr to create canary, canaryGroupName=%s", canaryGroupName)
					}
				}
			}
		}(canaryConfigEntry)
	}
}

func percentageBeanGeneratePage(list []*coreef.PercentageBean, page int, pageSize int) (result []*coreef.PercentageBean) {
	leng := len(list)
	startIndex := page*pageSize - pageSize
	result = make([]*coreef.PercentageBean, 0)
	if page < 1 || startIndex > leng || pageSize < 1 {
		return result
	}
	lastIndex := leng
	if page*pageSize < len(list) {
		lastIndex = page * pageSize
	}

	return list[startIndex:lastIndex]
}

func areFloatsEqualEnough(a float64, b float64) bool {
	return math.Abs(a-b) <= 0.001
}

func getPartnerOptionalCondition(newBean *coreef.PercentageBean) (string, error) {
	foundInvalidPartner := false
	if newBean.OptionalConditions != nil {
		if newBean.OptionalConditions.Condition != nil {
			//if it has one partnerid in optional condition it will be considered for creating canary
			if isPartnerConditionExists(*newBean.OptionalConditions.Condition) {
				currentPartnerId := strings.ToLower(*newBean.OptionalConditions.Condition.FixedArg.Bean.Value.JLString)
				if common.CanarySyndicatePartnerSet.Contains(currentPartnerId) || common.CanaryDefaultPartner == currentPartnerId {
					return currentPartnerId, nil
				} else {
					//if the partnerId mentioned in optionalcondition are not valid for canary
					return "", errors.New("PartnerId is invalid: not default or not found in Syndicate Partner List")
				}
			}
		} else {
			for _, conditions := range newBean.OptionalConditions.CompoundParts {
				if isPartnerConditionExists(*conditions.Condition) {
					currentPartnerId := strings.ToLower(*conditions.Condition.FixedArg.Bean.Value.JLString)
					//if the First partnerId matches the list or Default partner value will be considered for canary
					if common.CanarySyndicatePartnerSet.Contains(currentPartnerId) || common.CanaryDefaultPartner == currentPartnerId {
						return currentPartnerId, nil
					}
					// keep checking since there could be multiple optional conditions with partnerId and other conditions could be valid
					foundInvalidPartner = true
				}
			}
			// partnerId mentioned in optionalcondition are not valid for canary
			if foundInvalidPartner {
				return "", errors.New("PartnerId is invalid: not default or not found in Syndicate Partner List")
			}
		}
	}
	// if Optional Condition is not present fetch default from config
	log.Debugf("Default partner %s  will be used to create the canary", common.CanaryDefaultPartner)
	return common.CanaryDefaultPartner, nil
}

func isPartnerConditionExists(Condition re.Condition) bool {
	if Condition.FreeArg != nil && Condition.FreeArg.GetName() == PARTNER_ID && Condition.FixedArg != nil {
		return true
	}
	return false
}

func isWithinCanaryWindow(fields log.Fields) bool {
	t := time.Now().In(common.CanaryTimezone).Format(common.CanaryTimeFormat)
	if t < common.CanaryStartTime {
		log.WithFields(fields).Infof("Current time is before the canary start time, not creating canary. Current time=%s, startTime=%s", t, common.CanaryStartTime)
		return false
	}
	if t > common.CanaryEndTime {
		log.WithFields(fields).Infof("Current time is after the canary end time, not creating canary. Current time=%s, endTime=%s", t, common.CanaryEndTime)
		return false
	}
	return true
}

func isVideoDevice(model string) bool {
	return common.CanaryVideoModelSet.Contains(strings.ToUpper(model))
}

func PercentageBeanRuleGeneratePageWithContext(pbrules []*coreef.PercentageBean, contextMap map[string]string) (result []*coreef.PercentageBean, err error) {
	sort.Slice(pbrules, func(i, j int) bool {
		return strings.Compare(strings.ToLower(pbrules[i].Name), strings.ToLower(pbrules[j].Name)) < 0
	})
	pageNum := 1
	numStr, okval := contextMap[cPercentageBeanPageNumber]
	if okval {
		pageNum, _ = strconv.Atoi(numStr)
	}
	pageSize := 10
	szStr, okSz := contextMap[cPercentageBeanPageSize]
	if okSz {
		pageSize, _ = strconv.Atoi(szStr)
	}
	if pageNum < 1 || pageSize < 1 {
		return nil, errors.New("pageNumber and pageSize should both be greater than zero")
	}
	return percentageBeanGeneratePage(pbrules, pageNum, pageSize), nil
}

func PercentageBeanFilterByContext(searchContext map[string]string, applicationType string) []*coreef.PercentageBean {
	percentageBeansSearchResult := []*coreef.PercentageBean{}
	percentageBeans, err := GetAllPercentageBeansFromDB(applicationType, true, false)
	if err != nil {
		return percentageBeansSearchResult
	}
	for _, pbRule := range percentageBeans {
		if pbRule == nil {
			continue
		}
		if pbRule.ApplicationType != applicationType {
			continue
		}
		if name, ok := util.FindEntryInContext(searchContext, common.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(pbRule.Name), strings.ToLower(name)) {
				continue
			}
		}
		if env, ok := util.FindEntryInContext(searchContext, cPercentageBeanenvironment, false); ok {
			if !strings.Contains(strings.ToLower(pbRule.Environment), strings.ToLower(env)) {
				continue
			}
		}
		if lkg, ok := util.FindEntryInContext(searchContext, cPercentageBeanlastknowngood, false); ok {
			fc, err := coreef.GetFirmwareConfigOneDB(pbRule.LastKnownGood)
			if err != nil {
				continue
			}

			if !strings.Contains(strings.ToLower(fc.FirmwareVersion), strings.ToLower(lkg)) {
				continue
			}
		}
		if intver, ok := util.FindEntryInContext(searchContext, cPercentageBeanintermediateversion, false); ok {
			fc, err := coreef.GetFirmwareConfigOneDB(pbRule.IntermediateVersion)
			if err != nil {
				continue
			}
			if !strings.Contains(strings.ToLower(fc.FirmwareVersion), strings.ToLower(intver)) {
				continue
			}
		}
		if minCheckVersion, ok := util.FindEntryInContext(searchContext, cPercentageBeanmincheckversion, false); ok {
			if !containsMinCheckVersion(minCheckVersion, pbRule.FirmwareVersions) {
				continue
			}
		}

		if model, ok := util.FindEntryInContext(searchContext, xcommon.MODEL, false); ok {
			if !strings.Contains(strings.ToLower(pbRule.Model), strings.ToLower(model)) {
				continue
			}
		}

		if key, ok := util.FindEntryInContext(searchContext, common.FREE_ARG, false); ok {
			if pbRule.OptionalConditions == nil {
				continue
			}
			if !re.IsExistConditionByFreeArgName(*pbRule.OptionalConditions, key) {
				continue
			}
		}
		val, ok := util.FindEntryInContext(searchContext, common.FIXED_ARG, false)
		if ok {
			if pbRule.OptionalConditions == nil {
				continue
			}
			if !re.IsExistConditionByFixedArgValue(*pbRule.OptionalConditions, val) {
				continue
			}
		}
		percentageBeansSearchResult = append(percentageBeansSearchResult, pbRule)
	}
	return percentageBeansSearchResult
}

func containsMinCheckVersion(versionToSearch string, firmwareVersions []string) bool {
	if len(firmwareVersions) > 0 {
		for _, firmwareVersion := range firmwareVersions {
			if strings.Contains(strings.ToLower(firmwareVersion), strings.ToLower(versionToSearch)) {
				return true
			}
		}
	}
	return false
}

func replaceFieldsWithFirmwareVersion(bean *coreef.PercentageBean) *coreef.PercentageBean {
	if bean.LastKnownGood != "" {
		firmwareVersion := coreef.GetFirmwareVersion(bean.LastKnownGood)
		bean.LastKnownGood = firmwareVersion
	}

	if bean.IntermediateVersion != "" {
		firmwareVersion := coreef.GetFirmwareVersion(bean.IntermediateVersion)
		bean.IntermediateVersion = firmwareVersion
	}

	if bean.Distributions != nil && len(bean.Distributions) > 0 {
		firmwareVersionDistributions := make([]*firmware.ConfigEntry, 0)
		for _, dist := range bean.Distributions {
			if dist.ConfigId != "" {
				firmwareVersion := coreef.GetFirmwareVersion(dist.ConfigId)
				if firmwareVersion != "" {
					firmwareconfigentry := firmware.NewConfigEntry(firmwareVersion, dist.StartPercentRange, dist.EndPercentRange)
					firmwareconfigentry.IsCanaryDisabled = dist.IsCanaryDisabled
					firmwareconfigentry.IsPaused = dist.IsPaused
					firmwareVersionDistributions = append(firmwareVersionDistributions, firmwareconfigentry)
				} else {
					firmwareVersionDistributions = append(firmwareVersionDistributions, dist)
				}
			}
		}
		bean.Distributions = firmwareVersionDistributions
	}

	return bean
}
