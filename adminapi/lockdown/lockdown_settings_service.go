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

package lockdown

import (
	"errors"
	"fmt"
	"net/http"

	"xconfadmin/common"

	log "github.com/sirupsen/logrus"
)

func SetLockdownSetting(lockdownsetting *common.LockdownSettings) *common.ResponseEntity {
	if err := lockdownsetting.Validate(); err != nil {
		return common.NewResponseEntityWithStatus(http.StatusBadRequest, err, nil)
	}

	// Save all lockdown settings
	if lockdownsetting.LockdownEnabled != nil {
		if _, err := common.SetAppSetting(common.PROP_LOCKDOWN_ENABLED, *lockdownsetting.LockdownEnabled); err != nil {
			errorStr := fmt.Sprintf("Unable to save %s: %s", common.PROP_LOCKDOWN_ENABLED, err.Error())
			log.Error(errorStr)
			return common.NewResponseEntityWithStatus(http.StatusInternalServerError, errors.New(errorStr), nil)
		}
	}
	if lockdownsetting.LockdownStartTime != nil {
		if _, err := common.SetAppSetting(common.PROP_LOCKDOWN_STARTTIME, *lockdownsetting.LockdownStartTime); err != nil {
			errorStr := fmt.Sprintf("Unable to save %s: %s", common.PROP_LOCKDOWN_STARTTIME, err.Error())
			log.Error(errorStr)
			return common.NewResponseEntityWithStatus(http.StatusInternalServerError, errors.New(errorStr), nil)
		}
	}
	if lockdownsetting.LockdownEndTime != nil {
		if _, err := common.SetAppSetting(common.PROP_LOCKDOWN_ENDTIME, *lockdownsetting.LockdownEndTime); err != nil {
			errorStr := fmt.Sprintf("Unable to save %s: %s", common.PROP_LOCKDOWN_ENDTIME, err.Error())
			log.Error(errorStr)
			return common.NewResponseEntityWithStatus(http.StatusInternalServerError, errors.New(errorStr), nil)
		}
	}

	if lockdownsetting.LockdownModules != nil {
		if _, err := common.SetAppSetting(common.PROP_LOCKDOWN_MODULES, *lockdownsetting.LockdownModules); err != nil {
			errorStr := fmt.Sprintf("Unable to save %s: %s", common.PROP_LOCKDOWN_MODULES, err.Error())
			log.Error(errorStr)
			return common.NewResponseEntityWithStatus(http.StatusInternalServerError, errors.New(errorStr), nil)
		}
	}

	return common.NewResponseEntityWithStatus(http.StatusNoContent, nil, nil)
}

func GetLockdownSettings() (*common.LockdownSettings, error) {
	settings, err := common.GetAppSettings()
	if err != nil {
		return nil, err
	}

	lockdownsettings := common.LockdownSettings{}
	if v, ok := settings[common.PROP_LOCKDOWN_ENABLED]; ok {
		if value, ok := v.(bool); ok {
			lockdownsettings.LockdownEnabled = &value
		}
	}
	if v, ok := settings[common.PROP_LOCKDOWN_STARTTIME]; ok {
		if value, ok := v.(string); ok {
			lockdownsettings.LockdownStartTime = &value
		}
	}
	if v, ok := settings[common.PROP_LOCKDOWN_ENDTIME]; ok {
		if value, ok := v.(string); ok {
			lockdownsettings.LockdownEndTime = &value
		}
	}

	if v, ok := settings[common.PROP_LOCKDOWN_MODULES]; ok {
		if value, ok := v.(string); ok {
			lockdownsettings.LockdownModules = &value
		}
	}

	return &lockdownsettings, nil
}
