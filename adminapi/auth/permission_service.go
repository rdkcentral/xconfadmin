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
package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"xconfadmin/common"
	owcommon "xconfadmin/common"
	xhttp "xconfadmin/http"
	core "xconfadmin/shared"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"xconfadmin/util"
)

const (
	COMMON_MODULE    string = "common"
	TOOL_MODULE      string = "tools"
	CHANGE_MODULE    string = "changes"
	DCM_MODULE       string = "dcm"
	FIRMWARE_MODULE  string = "firmware"
	RFC_MODULE       string = "rfc"
	TELEMETRY_MODULE string = "telemetry"

	READ_COMMON  string = "read-common"
	WRITE_COMMON string = "write-common"

	VIEW_TOOLS  string = "view-tools"
	WRITE_TOOLS string = "write-tools"

	READ_DCM     string = "read-dcm-"
	READ_DCM_ALL string = "read-dcm-*"

	WRITE_DCM     string = "write-dcm-"
	WRITE_DCM_ALL string = "write-dcm-*"

	READ_FIRMWARE     string = "read-firmware-"
	READ_FIRMWARE_ALL string = "read-firmware-*"

	WRITE_FIRMWARE     string = "write-firmware-"
	WRITE_FIRMWARE_ALL string = "write-firmware-*"

	READ_TELEMETRY     string = "read-telemetry-"
	READ_TELEMETRY_ALL string = "read-telemetry-*"

	WRITE_TELEMETRY     string = "write-telemetry-"
	WRITE_TELEMETRY_ALL string = "write-telemetry-*"

	READ_CHANGES     string = "read-changes-"
	READ_CHANGES_ALL string = "read-changes-*"

	WRITE_CHANGES     string = "write-changes-"
	WRITE_CHANGES_ALL string = "write-changes-*"

	XCONF_ALL           string = "x1:appds:xconf:*"
	XCONF_READ          string = "x1:coast:xconf:read"
	XCONF_READ_MACLIST  string = "x1:coast:xconf:read:maclist"
	XCONF_WRITE         string = "x1:coast:xconf:write"
	XCONF_WRITE_MACLIST string = "x1:coast:xconf:write:maclist"

	DEV_PROFILE string = "dev"

	COMMON_ENTITY    string = "CommonEntity"
	TOOL_ENTITY      string = "ToolEntity"
	CHANGE_ENTITY    string = "ChangeEntity"
	DCM_ENTITY       string = "DcmEntity"
	FIRMWARE_ENTITY  string = "FirmwareEntity"
	TELEMETRY_ENTITY string = "TelemetryEntity"
)

type EntityPermission struct {
	ReadAll  string `json:"readAll,omitempty"`
	Read     string `json:"read,omitempty"`
	WriteAll string `json:"writeAll,omitempty"`
	Write    string `json:"write,omitempty"`
}

var CommonPermissions = EntityPermission{
	ReadAll:  READ_COMMON,
	WriteAll: WRITE_COMMON,
}

var ToolPermissions = EntityPermission{
	ReadAll:  VIEW_TOOLS,
	WriteAll: WRITE_TOOLS,
}

var FirmwarePermissions = EntityPermission{
	ReadAll:  READ_FIRMWARE_ALL,
	Read:     READ_FIRMWARE,
	WriteAll: WRITE_FIRMWARE_ALL,
	Write:    WRITE_FIRMWARE,
}

var ChangePermissions = EntityPermission{
	ReadAll:  READ_CHANGES_ALL,
	Read:     READ_CHANGES,
	WriteAll: WRITE_CHANGES_ALL,
	Write:    WRITE_CHANGES,
}

var DcmPermissions = EntityPermission{
	ReadAll:  READ_DCM_ALL,
	Read:     READ_DCM,
	WriteAll: WRITE_DCM_ALL,
	Write:    WRITE_DCM,
}

var TelemetryPermissions = EntityPermission{
	ReadAll:  READ_TELEMETRY_ALL,
	Read:     READ_TELEMETRY,
	WriteAll: WRITE_TELEMETRY_ALL,
	Write:    WRITE_TELEMETRY,
}

func getEntityPermission(entityType string) *EntityPermission {
	if entityType == COMMON_ENTITY {
		return &CommonPermissions
	}
	if entityType == TOOL_ENTITY {
		return &CommonPermissions
	}
	if entityType == CHANGE_ENTITY {
		return &ChangePermissions
	}
	if entityType == DCM_ENTITY {
		return &DcmPermissions
	}
	if entityType == FIRMWARE_ENTITY {
		return &FirmwarePermissions
	}
	if entityType == TELEMETRY_ENTITY {
		return &TelemetryPermissions
	}
	return nil
}

func getCurrentModule(r *http.Request, entityType string) string {
	if entityType == COMMON_ENTITY {
		return COMMON_MODULE
	}
	if entityType == TOOL_ENTITY {
		return TOOL_MODULE
	}
	if entityType == CHANGE_ENTITY {
		return CHANGE_MODULE
	}
	if entityType == DCM_ENTITY {
		rfcpaths := []string{"/rfc", "/feature", "/featurerule"}
		if util.StringArrayContains(rfcpaths, r.URL.Path) {
			return RFC_MODULE
		}
		return DCM_MODULE
	}
	if entityType == FIRMWARE_ENTITY {
		rfcpaths := []string{"/rfc", "/feature", "/featurerule"}
		if util.StringArrayContains(rfcpaths, r.URL.Path) {
			return RFC_MODULE
		}
		return FIRMWARE_MODULE
	}
	if entityType == TELEMETRY_ENTITY {
		return TELEMETRY_MODULE
	}
	return ""
}

func HasReadPermissionForTool(r *http.Request) bool {
	if !(owcommon.SatOn) {
		return true
	}

	// checked capabilities from SAT token if available
	if capabilities := xhttp.GetCapabilitiesFromContext(r); len(capabilities) > 0 {
		if util.Contains(capabilities, XCONF_ALL) || util.Contains(capabilities, XCONF_READ) {
			return true
		}
	} else {
		// checked permissions from Login token
		permissions := GetPermissionsFunc(r)
		if util.Contains(permissions, getEntityPermission(TOOL_ENTITY).ReadAll) {
			return true
		}
	}
	return false
}

func HasWritePermissionForTool(r *http.Request) bool {
	if !(owcommon.SatOn) {
		return true
	}

	// checked capabilities from SAT token if available
	if capabilities := xhttp.GetCapabilitiesFromContext(r); len(capabilities) > 0 {
		if util.Contains(capabilities, XCONF_ALL) || util.Contains(capabilities, XCONF_WRITE) {
			return true
		}
	} else {
		// checked permissions from Login token
		permissions := GetPermissionsFunc(r)
		if util.Contains(permissions, getEntityPermission(TOOL_ENTITY).WriteAll) {
			return true
		}
	}
	return false
}

// CanWrite returns the applicationType the user has write permission for non-common entityType,
// otherwise returns error if applicationType is not specified in query parameter or cookie
func CanWrite(r *http.Request, entityType string, vargs ...string) (applicationType string, err error) {
	if isLockdownMode() {
		lockdownModules := strings.Split(common.GetStringAppSetting(common.PROP_LOCKDOWN_MODULES), ",")
		if len(lockdownModules) != 0 {
			if util.CaseInsensitiveContains(lockdownModules, getCurrentModule(r, entityType)) || strings.ToUpper(lockdownModules[0]) == common.DefaultLockdownModules {
				return "", xwcommon.NewRemoteErrorAS(http.StatusLocked, "Modification not allowed in Lockdown mode")
			}
		}
	}

	if entityType != COMMON_ENTITY && entityType != TOOL_ENTITY {
		if values, ok := r.URL.Query()[core.APPLICATION_TYPE]; ok {
			applicationType = values[0]
		}
		if util.IsBlank(applicationType) {
			applicationType = core.GetApplicationFromCookies(r)
		}
		if util.IsBlank(applicationType) {
			if len(vargs) > 0 && vargs[0] != "" {
				applicationType = vargs[0]
			} else {
				// work-around for backward compatibility
				log.Infof("applicationType not specified: auth_subject=%s path=%s", r.Header.Get(xhttp.AUTH_SUBJECT), r.URL.Path)
				applicationType = core.STB
			}
		}
		if err := core.ValidateApplicationType(applicationType); err != nil {
			return "", err
		}
	}

	//TODO
	if !(owcommon.SatOn) {
		return applicationType, nil
	}

	// checked capabilities from SAT token if available
	if capabilities := xhttp.GetCapabilitiesFromContext(r); len(capabilities) > 0 {
		if entityType == COMMON_ENTITY && util.Contains(capabilities, XCONF_WRITE_MACLIST) {
			return applicationType, nil
		}
		if !(util.Contains(capabilities, XCONF_ALL) || util.Contains(capabilities, XCONF_WRITE)) {
			return "", xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No write capabilities")
		}
		return applicationType, nil
	} else {
		// checked permissions from Login token
		permissions := GetPermissionsFunc(r)
		if util.Contains(permissions, getEntityPermission(entityType).WriteAll) {
			return applicationType, nil
		}
		if util.Contains(common.ApplicationTypes, applicationType) && util.Contains(permissions, getEntityPermission(entityType).Write+applicationType) {
			return applicationType, nil
		}
	}

	if applicationType == "" {
		return "", xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No write permission")
	} else {
		return "", xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No write permission for ApplicationType "+applicationType)
	}
}

// CanRead returns the applicationType the user has read permission for non-common entityType,
// otherwise returns error if applicationType is not specified in query parameter or cookie
func CanRead(r *http.Request, entityType string, vargs ...string) (applicationType string, err error) {
	if entityType != COMMON_ENTITY && entityType != TOOL_ENTITY {
		if values, ok := r.URL.Query()[core.APPLICATION_TYPE]; ok {
			applicationType = values[0]
		}
		if util.IsBlank(applicationType) {
			applicationType = core.GetApplicationFromCookies(r)
		}
		if util.IsBlank(applicationType) {
			if len(vargs) > 0 && vargs[0] != "" {
				applicationType = vargs[0]
			} else {
				// work-around for backward compatibility
				log.Infof("applicationType not specified: auth_subject=%s path=%s", r.Header.Get(xhttp.AUTH_SUBJECT), r.URL.Path)
				applicationType = core.STB
			}
		}
		if err := core.ValidateApplicationType(applicationType); err != nil {
			return "", err
		}
	}

	if !(owcommon.SatOn) {
		return applicationType, nil
	}

	// checked capabilities from SAT token if available
	if capabilities := xhttp.GetCapabilitiesFromContext(r); len(capabilities) > 0 {
		if entityType == COMMON_ENTITY && util.Contains(capabilities, XCONF_READ_MACLIST) {
			return applicationType, nil
		}
		if !(util.Contains(capabilities, XCONF_ALL) || util.Contains(capabilities, XCONF_READ)) {
			return "", xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No read capabilities")
		}
		return applicationType, nil
	} else {
		// checked permissions from Login token
		permissions := GetPermissionsFunc(r)
		if util.Contains(permissions, getEntityPermission(entityType).ReadAll) {
			return applicationType, nil
		}

		if util.Contains(common.ApplicationTypes, applicationType) && util.Contains(permissions, getEntityPermission(entityType).Read+applicationType) {
			return applicationType, nil
		}
	}

	if applicationType == "" {
		return "", xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No read permission")
	} else {
		return "", xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No read permission for ApplicationType "+applicationType)
	}
}

var GetPermissionsFunc = getPermissions

func getPermissions(r *http.Request) (permissions []string) {
	if IsDevProfile() {
		permissions = []string{
			WRITE_COMMON, READ_COMMON,
			WRITE_FIRMWARE_ALL, READ_FIRMWARE_ALL,
			WRITE_DCM_ALL, READ_DCM_ALL,
			WRITE_TELEMETRY_ALL, READ_TELEMETRY_ALL,
			READ_CHANGES_ALL, WRITE_CHANGES_ALL}
	} else {
		permissions = xhttp.GetPermissionsFromContext(r)
	}
	return permissions
}

func IsDevProfile() bool {
	activeProfiles := strings.Split(strings.TrimSpace(owcommon.ActiveAuthProfiles), ",")
	if len(activeProfiles) > 0 {
		return DEV_PROFILE == activeProfiles[0]
	}
	defaultProfiles := strings.Split(strings.TrimSpace(owcommon.DefaultAuthProfiles), ",")
	return DEV_PROFILE == defaultProfiles[0]
}

func ValidateRead(r *http.Request, entityApplicationType string, entityType string) error {
	if err := core.ValidateApplicationType(entityApplicationType); err != nil {
		return err
	}
	applicationType, err := CanRead(r, entityType)
	if err != nil {
		return err
	}
	if applicationType != entityApplicationType {
		return xwcommon.NewRemoteErrorAS(http.StatusForbidden,
			fmt.Sprintf("Current ApplicationType %s doesn't match with entity's ApplicationType: %s", applicationType, entityApplicationType))
	}
	return nil
}

func ValidateWrite(r *http.Request, entityApplicationType string, entityType string) error {
	if err := core.ValidateApplicationType(entityApplicationType); err != nil {
		return err
	}
	applicationType, err := CanWrite(r, entityType, entityApplicationType)
	if err != nil {
		return err
	}
	if applicationType != entityApplicationType {
		return xwcommon.NewRemoteErrorAS(http.StatusForbidden,
			fmt.Sprintf("Current ApplicationType %s doesn't match with entity's ApplicationType: %s", applicationType, entityApplicationType))
	}
	return nil
}

func isLockdownMode() bool {
	if owcommon.GetBooleanAppSetting(owcommon.PROP_LOCKDOWN_ENABLED, false) {
		startTime := owcommon.GetStringAppSetting(owcommon.PROP_LOCKDOWN_STARTTIME)
		endTime := owcommon.GetStringAppSetting(owcommon.PROP_LOCKDOWN_ENDTIME)

		timezone, err := time.LoadLocation(owcommon.DefaultLockdownTimezone)
		if err != nil {
			log.Errorf("Error loading timezone: %s", owcommon.DefaultLockdownTimezone)
			return false
		}

		t := time.Now().In(timezone).Format(owcommon.DefaultTimeDateFormatLayout)
		CurrentDate := time.Now().In(timezone).Format(owcommon.DefaultDateFormatLayout)

		Currenttime, err := time.Parse(owcommon.DefaultTimeDateFormatLayout, t)

		if err != nil {
			log.Errorf("Unable to Parse currenttime: %s", Currenttime)
			return false
		}
		LockdownStartTime, err := time.Parse(owcommon.DefaultTimeDateFormatLayout, CurrentDate+" "+startTime)
		if err != nil {
			log.Errorf("Unable to Parse LockdownStartTime: %s", LockdownStartTime)
			return false
		}
		LockdownEndTime, err := time.Parse(owcommon.DefaultTimeDateFormatLayout, CurrentDate+" "+endTime)
		if err != nil {
			log.Errorf("Unable to Parse LockdownEndTime: %s", LockdownEndTime)
			return false
		}

		if LockdownStartTime.After(LockdownEndTime) || LockdownStartTime.Equal(LockdownEndTime) {
			LockdownStartTime = LockdownStartTime.AddDate(0, 0, -1)
		}

		if (Currenttime.Equal(LockdownStartTime) || Currenttime.After(LockdownStartTime)) && Currenttime.Before(LockdownEndTime) {
			log.Infof("Lockdown Mode is Scheduled Now. Current time=%s, Lockdown StartTime=%s, Lockdown EndTime=%s", t, startTime, endTime)
			return true
		}
		return false
	}
	return false
}

func GetUserNameOrUnknown(r *http.Request) string {
	if userName := r.Header.Get(xhttp.AUTH_SUBJECT); userName == "" {
		return xhttp.UNKNOWN_USER
	} else {
		return userName
	}
}

func ExtractBodyAndCheckPermissions(obj owcommon.ApplicationTypeAware, w http.ResponseWriter, r *http.Request, entityType string) (applicationType string, err error) {
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		return "", xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "responsewriter cast error")
	}
	body := xw.Body()
	err = json.Unmarshal([]byte(body), &obj)
	if err != nil {
		return "", xwcommon.NewRemoteErrorAS(http.StatusBadRequest, err.Error())
	}

	applicationType, err = CanWrite(r, entityType, obj.GetApplicationType())
	if err != nil {
		return "", err
	}

	if obj.GetApplicationType() == "" {
		obj.SetApplicationType(applicationType)
	} else if obj.GetApplicationType() != applicationType {
		return "", xwcommon.NewRemoteErrorAS(http.StatusConflict, "ApplicationType Conflict")
	}
	return applicationType, nil
}

func isReadonlyMode() bool {
	return owcommon.GetBooleanAppSetting(owcommon.READONLY_MODE, false)
}
