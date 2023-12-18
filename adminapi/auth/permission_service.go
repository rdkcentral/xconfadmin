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
package auth

import (
	"fmt"
	"net/http"
	"strings"

	xcommon "xconfadmin/common"

	xhttp "xconfadmin/http"

	xshared "xconfadmin/shared"
	xwcommon "xconfwebconfig/common"
	xwshared "xconfwebconfig/shared"
	"xconfwebconfig/util"
)

const (
	READ_COMMON  string = "read-common"
	WRITE_COMMON string = "write-common"

	VIEW_TOOLS  string = "view-tools"
	WRITE_TOOLS string = "write-tools"

	READ_DCM_XHOME    string = "read-dcm-xhome"
	READ_DCM_STB      string = "read-dcm-stb"
	READ_DCM_RDLCLOUD string = "read-dcm-rdkcloud"
	READ_DCM_SKY      string = "read-dcm-sky"
	READ_DCM_ALL      string = "read-dcm-*"

	WRITE_DCM_XHOME    string = "write-dcm-xhome"
	WRITE_DCM_STB      string = "write-dcm-stb"
	WRITE_DCM_RDKCLOUD string = "write-dcm-rdkcloud"
	WRITE_DCM_SKY      string = "write-dcm-sky"
	WRITE_DCM_ALL      string = "write-dcm-*"

	READ_FIRMWARE_XHOME    string = "read-firmware-xhome"
	READ_FIRMWARE_STB      string = "read-firmware-stb"
	READ_FIRMWARE_RDKCLOUD string = "read-firmware-rdkcloud"
	READ_FIRMWARE_SKY      string = "read-firmware-sky"
	READ_FIRMWARE_ALL      string = "read-firmware-*"

	WRITE_FIRMWARE_XHOME    string = "write-firmware-xhome"
	WRITE_FIRMWARE_STB      string = "write-firmware-stb"
	WRITE_FIRMWARE_RDKCLOUD string = "write-firmware-rdkcloud"
	WRITE_FIRMWARE_SKY      string = "write-firmware-sky"
	WRITE_FIRMWARE_ALL      string = "write-firmware-*"

	READ_TELEMETRY_XHOME    string = "read-telemetry-xhome"
	READ_TELEMETRY_STB      string = "read-telemetry-stb"
	READ_TELEMETRY_RDKCLOUD string = "read-telemetry-rdkcloud"
	READ_TELEMETRY_SKY      string = "read-telemetry-sky"
	READ_TELEMETRY_ALL      string = "read-telemetry-*"

	WRITE_TELEMETRY_XHOME    string = "write-telemetry-xhome"
	WRITE_TELEMETRY_STB      string = "write-telemetry-stb"
	WRITE_TELEMETRY_RDKCLOUD string = "write-telemetry-rdkcloud"
	WRITE_TELEMETRY_SKY      string = "write-telemetry-sky"
	WRITE_TELEMETRY_ALL      string = "write-telemetry-*"

	READ_CHANGES_XHOME    string = "read-changes-xhome"
	READ_CHANGES_STB      string = "read-changes-stb"
	READ_CHANGES_RDKCLOUD string = "read-changes-rdkcloud"
	READ_CHANGES_SKY      string = "read-changes-sky"
	READ_CHANGES_ALL      string = "read-changes-*"

	WRITE_CHANGES_XHOME    string = "write-changes-xhome"
	WRITE_CHANGES_STB      string = "write-changes-stb"
	WRITE_CHANGES_RDKCLOUD string = "write-changes-rdkcloud"
	WRITE_CHANGES_SKY      string = "write-changes-sky"
	WRITE_CHANGES_ALL      string = "write-changes-*"

	XCONF_ALL           string = "x1:appds:xconf:*"
	XCONF_READ          string = "x1:coast:xconf:read"
	XCONF_READ_MACLIST  string = "x1:coast:xconf:read:maclist"
	XCONF_WRITE         string = "x1:coast:xconf:write"
	XCONF_WRITE_MACLIST string = "x1:coast:xconf:write:maclist"

	IGNORE_GETENV_PROPERTY_NAME    string = "spring.getenv.ignore"
	ACTIVE_PROFILES_PROPERTY_NAME  string = "spring.profiles.active"
	DEFAULT_PROFILES_PROPERTY_NAME string = "spring.profiles.default"
	RESERVED_DEFAULT_PROFILE_NAME  string = "default"
	DEV_PROFILE                    string = "dev"

	COMMON_ENTITY    string = "CommonEntity"
	TOOL_ENTITY      string = "ToolEntity"
	CHANGE_ENTITY    string = "ChangeEntity"
	DCM_ENTITY       string = "DcmEntity"
	FIRMWARE_ENTITY  string = "FirmwareEntity"
	TELEMETRY_ENTITY string = "TelemetryEntity"
)

type EntityPermission struct {
	ReadAll       string `json:"readAll,omitempty"`
	ReadStb       string `json:"readStb,omitempty"`
	ReadXhome     string `json:"readXhome,omitempty"`
	ReadRdkcloud  string `json:"readRdkcloud,omitempty"`
	ReadSky       string `json:"readSky,omitempty"`
	WriteAll      string `json:"writeAll,omitempty"`
	WriteStb      string `json:"writeStb,omitempty"`
	WriteXhome    string `json:"writeXhome,omitempty"`
	WriteRdkcloud string `json:"writeRdkcloud,omitempty"`
	WriteSky      string `json:"writeSky,omitempty"`
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
	ReadAll:       READ_FIRMWARE_ALL,
	ReadStb:       READ_FIRMWARE_STB,
	ReadXhome:     READ_FIRMWARE_XHOME,
	ReadRdkcloud:  READ_FIRMWARE_RDKCLOUD,
	ReadSky:       READ_FIRMWARE_SKY,
	WriteAll:      WRITE_FIRMWARE_ALL,
	WriteStb:      WRITE_FIRMWARE_STB,
	WriteXhome:    WRITE_FIRMWARE_XHOME,
	WriteRdkcloud: WRITE_FIRMWARE_RDKCLOUD,
	WriteSky:      WRITE_FIRMWARE_SKY,
}

var ChangePermissions = EntityPermission{
	ReadAll:       READ_CHANGES_ALL,
	ReadStb:       READ_CHANGES_STB,
	ReadXhome:     READ_CHANGES_XHOME,
	ReadRdkcloud:  READ_CHANGES_RDKCLOUD,
	ReadSky:       READ_CHANGES_SKY,
	WriteAll:      WRITE_CHANGES_ALL,
	WriteStb:      WRITE_CHANGES_STB,
	WriteXhome:    WRITE_CHANGES_XHOME,
	WriteRdkcloud: WRITE_CHANGES_RDKCLOUD,
	WriteSky:      WRITE_CHANGES_SKY,
}

var DcmPermissions = EntityPermission{
	ReadAll:       READ_DCM_ALL,
	ReadStb:       READ_DCM_STB,
	ReadXhome:     READ_DCM_XHOME,
	ReadRdkcloud:  READ_DCM_RDLCLOUD,
	ReadSky:       READ_DCM_SKY,
	WriteAll:      WRITE_DCM_ALL,
	WriteStb:      WRITE_DCM_STB,
	WriteXhome:    WRITE_DCM_XHOME,
	WriteRdkcloud: WRITE_DCM_RDKCLOUD,
	WriteSky:      WRITE_DCM_SKY,
}

var TelemetryPermissions = EntityPermission{
	ReadAll:       READ_TELEMETRY_ALL,
	ReadStb:       READ_TELEMETRY_STB,
	ReadXhome:     READ_TELEMETRY_XHOME,
	ReadRdkcloud:  READ_TELEMETRY_RDKCLOUD,
	ReadSky:       READ_TELEMETRY_SKY,
	WriteAll:      WRITE_TELEMETRY_ALL,
	WriteStb:      WRITE_TELEMETRY_STB,
	WriteXhome:    WRITE_TELEMETRY_XHOME,
	WriteRdkcloud: WRITE_TELEMETRY_RDKCLOUD,
	WriteSky:      WRITE_TELEMETRY_SKY,
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

func HasReadPermissionForTool(r *http.Request) bool {
	if !(xcommon.SatOn) {
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
	if !(xcommon.SatOn) {
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
// otherwise returns error if applicationType is not specified in query parameter, cookie, or vargs param
func CanWrite(r *http.Request, entityType string, vargs ...string) (applicationType string, err error) {
	if isReadonlyMode() {
		return "", xcommon.NewXconfError(http.StatusForbidden, "Modification not allowed in read-only mode")
	}

	if entityType != COMMON_ENTITY && entityType != TOOL_ENTITY {
		if values, ok := r.URL.Query()[xwcommon.APPLICATION_TYPE]; ok {
			applicationType = values[0]
		}
		if util.IsBlank(applicationType) {
			applicationType = xshared.GetApplicationFromCookies(r)
		}
		if util.IsBlank(applicationType) {
			if len(vargs) > 0 {
				applicationType = vargs[0]
			}
		}
		if err := xshared.ValidateApplicationType(applicationType); err != nil {
			return "", err
		}
	}

	if !(xcommon.SatOn) {
		return applicationType, nil
	}

	// checked capabilities from SAT token if available
	if capabilities := xhttp.GetCapabilitiesFromContext(r); len(capabilities) > 0 {
		if !(util.Contains(capabilities, XCONF_ALL) || util.Contains(capabilities, XCONF_WRITE)) {
			return "", xcommon.NewXconfError(http.StatusForbidden, "No write capabilities")
		}
		return applicationType, nil
	} else {
		// checked permissions from Login token
		permissions := GetPermissionsFunc(r)
		if util.Contains(permissions, getEntityPermission(entityType).WriteAll) {
			return applicationType, nil
		}
		if xwshared.STB == applicationType && util.Contains(permissions, getEntityPermission(entityType).WriteStb) {
			return applicationType, nil
		}
	}

	if applicationType == "" {
		return "", xcommon.NewXconfError(http.StatusForbidden, "No write permission")
	} else {
		return "", xcommon.NewXconfError(http.StatusForbidden, "No write permission for ApplicationType "+applicationType)
	}
}

// CanRead returns the applicationType the user has read permission for non-common entityType,
// otherwise returns error if applicationType is not specified in query parameter, cookie, or vargs param.
func CanRead(r *http.Request, entityType string, vargs ...string) (applicationType string, err error) {
	if entityType != COMMON_ENTITY && entityType != TOOL_ENTITY {
		if values, ok := r.URL.Query()[xwcommon.APPLICATION_TYPE]; ok {
			applicationType = values[0]
		}
		if util.IsBlank(applicationType) {
			applicationType = xshared.GetApplicationFromCookies(r)
		}
		if util.IsBlank(applicationType) {
			if len(vargs) > 0 {
				applicationType = vargs[0]
			}
		}
		if err := xshared.ValidateApplicationType(applicationType); err != nil {
			return "", err
		}
	}

	if !(xcommon.SatOn) {
		return applicationType, nil
	}

	// checked capabilities from SAT token if available
	if capabilities := xhttp.GetCapabilitiesFromContext(r); len(capabilities) > 0 {
		if !(util.Contains(capabilities, XCONF_ALL) || util.Contains(capabilities, XCONF_READ)) {
			return "", xcommon.NewXconfError(http.StatusForbidden, "No read capabilities")
		}
		return applicationType, nil
	} else {
		// checked permissions from Login token
		permissions := GetPermissionsFunc(r)
		if util.Contains(permissions, getEntityPermission(entityType).ReadAll) {
			return applicationType, nil
		}
		if xwshared.STB == applicationType && util.Contains(permissions, getEntityPermission(entityType).ReadStb) {
			return applicationType, nil
		}
	}

	if applicationType == "" {
		return "", xcommon.NewXconfError(http.StatusForbidden, "No read permission")
	} else {
		return "", xcommon.NewXconfError(http.StatusForbidden, "No read permission for ApplicationType "+applicationType)
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
	activeProfiles := strings.Split(strings.TrimSpace(xcommon.ActiveAuthProfiles), ",")
	if len(activeProfiles) > 0 {
		return DEV_PROFILE == activeProfiles[0]
	}
	defaultProfiles := strings.Split(strings.TrimSpace(xcommon.DefaultAuthProfiles), ",")
	return DEV_PROFILE == defaultProfiles[0]
}

func ValidateRead(r *http.Request, entityApplicationType string, entityType string) error {
	if err := xshared.ValidateApplicationType(entityApplicationType); err != nil {
		return err
	}
	applicationType, err := CanRead(r, entityType)
	if err != nil {
		return err
	}
	if applicationType != entityApplicationType {
		return xcommon.NewXconfError(http.StatusForbidden,
			fmt.Sprintf("Current ApplicationType %s doesn't match with entity's ApplicationType: %s", applicationType, entityApplicationType))
	}
	return nil
}

func ValidateWrite(r *http.Request, entityApplicationType string, entityType string) error {
	if err := xshared.ValidateApplicationType(entityApplicationType); err != nil {
		return err
	}
	applicationType, err := CanWrite(r, entityType)
	if err != nil {
		return err
	}
	if applicationType != entityApplicationType {
		return xcommon.NewXconfError(http.StatusForbidden,
			fmt.Sprintf("Current ApplicationType %s doesn't match with entity's ApplicationType: %s", applicationType, entityApplicationType))
	}
	return nil
}

func isReadonlyMode() bool {
	return xcommon.GetBooleanAppSetting(xcommon.READONLY_MODE, false)
}

func GetUserNameOrUnknown(r *http.Request) string {
	if userName := r.Header.Get(xhttp.AUTH_SUBJECT); userName == "" {
		return xhttp.UNKNOWN_USER
	} else {
		return userName
	}
}
