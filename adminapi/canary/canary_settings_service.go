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

package canary

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rdkcentral/xconfadmin/common"

	log "github.com/sirupsen/logrus"
)

func SetCanarySetting(canarySettings *common.CanarySettings) *common.ResponseEntity {
	if err := canarySettings.Validate(); err != nil {
		return common.NewResponseEntityWithStatus(http.StatusBadRequest, err, nil)
	}

	// Save all canary settings
	if canarySettings.CanaryMaxSize != nil {
		if _, err := common.SetAppSetting(common.PROP_CANARY_MAXSIZE, *canarySettings.CanaryMaxSize); err != nil {
			errorStr := fmt.Sprintf("Unable to save %s: %s", common.PROP_CANARY_MAXSIZE, err.Error())
			log.Error(errorStr)
			return common.NewResponseEntityWithStatus(http.StatusInternalServerError, errors.New(errorStr), nil)
		}
	}
	if canarySettings.CanaryDistributionPercentage != nil {
		if _, err := common.SetAppSetting(common.PROP_CANARY_DISTRIBUTION_PERCENTAGE, *canarySettings.CanaryDistributionPercentage); err != nil {
			errorStr := fmt.Sprintf("Unable to save %s: %s", common.PROP_CANARY_DISTRIBUTION_PERCENTAGE, err.Error())
			log.Error(errorStr)
			return common.NewResponseEntityWithStatus(http.StatusInternalServerError, errors.New(errorStr), nil)
		}
	}
	if canarySettings.CanaryFwUpgradeStartTime != nil {
		if _, err := common.SetAppSetting(common.PROP_CANARY_FW_UPGRADE_STARTTIME, *canarySettings.CanaryFwUpgradeStartTime); err != nil {
			errorStr := fmt.Sprintf("Unable to save %s: %s", common.PROP_CANARY_FW_UPGRADE_STARTTIME, err.Error())
			log.Error(errorStr)
			return common.NewResponseEntityWithStatus(http.StatusInternalServerError, errors.New(errorStr), nil)
		}
	}
	if canarySettings.CanaryFwUpgradeEndTime != nil {
		if _, err := common.SetAppSetting(common.PROP_CANARY_FW_UPGRADE_ENDTIME, *canarySettings.CanaryFwUpgradeEndTime); err != nil {
			errorStr := fmt.Sprintf("Unable to save %s: %s", common.PROP_CANARY_FW_UPGRADE_ENDTIME, err.Error())
			log.Error(errorStr)
			return common.NewResponseEntityWithStatus(http.StatusInternalServerError, errors.New(errorStr), nil)
		}
	}
	return common.NewResponseEntityWithStatus(http.StatusNoContent, nil, nil)
}

func GetCanarySettings() (*common.CanarySettings, error) {
	settings, err := common.GetAppSettings()
	if err != nil {
		return nil, err
	}

	// Note: json.Unmarshal numbers into float64 when target type is of type interface{}
	intValue := func(v interface{}) *int {
		var value int
		if val, ok := v.(float64); ok {
			value = int(val)
		} else {
			value = v.(int)
		}
		return &value
	}

	canarySettings := common.CanarySettings{}
	if v, ok := settings[common.PROP_CANARY_DISTRIBUTION_PERCENTAGE]; ok {
		if value, ok := v.(float64); ok {
			canarySettings.CanaryDistributionPercentage = &value
		}
	}
	if v, ok := settings[common.PROP_CANARY_MAXSIZE]; ok {
		canarySettings.CanaryMaxSize = intValue(v)
	}
	if v, ok := settings[common.PROP_CANARY_FW_UPGRADE_STARTTIME]; ok {
		canarySettings.CanaryFwUpgradeStartTime = intValue(v)
	}
	if v, ok := settings[common.PROP_CANARY_FW_UPGRADE_ENDTIME]; ok {
		canarySettings.CanaryFwUpgradeEndTime = intValue(v)
	}

	return &canarySettings, nil
}
