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

	"github.com/google/uuid"
	"github.com/rdkcentral/xconfadmin/common"
	owcommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	core "github.com/rdkcentral/xconfadmin/shared"
	"github.com/rdkcentral/xconfadmin/util"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	log "github.com/sirupsen/logrus"
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

type SATv2Domain string

const (
	DOMAIN_CORE    SATv2Domain = "core"
	DOMAIN_TAGGING SATv2Domain = "tagging"
	DOMAIN_SYSTEM  SATv2Domain = "system"
	DOMAIN_METRICS SATv2Domain = "metrics"
)

type RouteDomainMapping struct {
	Prefix string
	Domain SATv2Domain
}

var satV2RouteMappings = []RouteDomainMapping{
	// tagging (own top-level router, must appear before xconfAdminService stripping)
	{Prefix: "/taggingservice", Domain: DOMAIN_TAGGING},

	// metrics
	{Prefix: "/metrics", Domain: DOMAIN_METRICS},

	// system
	{Prefix: "/queries/filters/downloadlocation", Domain: DOMAIN_SYSTEM},
	{Prefix: "/updates/filters/downloadlocation", Domain: DOMAIN_SYSTEM},
	{Prefix: "/roundrobinfilter", Domain: DOMAIN_SYSTEM},
	{Prefix: "/rfc/recooking", Domain: DOMAIN_SYSTEM},
	{Prefix: "/rfc/preprocess", Domain: DOMAIN_SYSTEM},
	{Prefix: "/appsettings", Domain: DOMAIN_SYSTEM},
	{Prefix: "/canarysettings", Domain: DOMAIN_SYSTEM},
	{Prefix: "/lockdownsettings", Domain: DOMAIN_SYSTEM},
	{Prefix: "/wakeuppool", Domain: DOMAIN_SYSTEM},

	// core
	{Prefix: "/dataservice", Domain: DOMAIN_CORE},
	{Prefix: "/estbfirmware", Domain: DOMAIN_CORE},
	{Prefix: "/queries", Domain: DOMAIN_CORE},
	{Prefix: "/updates", Domain: DOMAIN_CORE},
	{Prefix: "/delete", Domain: DOMAIN_CORE},
	{Prefix: "/model", Domain: DOMAIN_CORE},
	{Prefix: "/environment", Domain: DOMAIN_CORE},
	{Prefix: "/genericnamespacedlist", Domain: DOMAIN_CORE},
	{Prefix: "/firmwarerule", Domain: DOMAIN_CORE},
	{Prefix: "/firmwareruletemplate", Domain: DOMAIN_CORE},
	{Prefix: "/firmwareconfig", Domain: DOMAIN_CORE},
	{Prefix: "/percentfilter", Domain: DOMAIN_CORE},
	{Prefix: "/amv", Domain: DOMAIN_CORE},
	{Prefix: "/activationminimumversion", Domain: DOMAIN_CORE},
	{Prefix: "/settings", Domain: DOMAIN_CORE},
	{Prefix: "/setting", Domain: DOMAIN_CORE},
	{Prefix: "/featurerule", Domain: DOMAIN_CORE},
	{Prefix: "/feature", Domain: DOMAIN_CORE},
	{Prefix: "/rfc", Domain: DOMAIN_CORE},
	{Prefix: "/changelog", Domain: DOMAIN_CORE},
	{Prefix: "/log", Domain: DOMAIN_CORE},
	{Prefix: "/reportpage", Domain: DOMAIN_CORE},
	{Prefix: "/stats", Domain: DOMAIN_CORE},
	{Prefix: "/migration", Domain: DOMAIN_CORE},
	{Prefix: "/dcm", Domain: DOMAIN_CORE},
	{Prefix: "/telemetry", Domain: DOMAIN_CORE},
	{Prefix: "/change", Domain: DOMAIN_CORE},
	{Prefix: "/penetrationdata", Domain: DOMAIN_CORE},
}

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
		return &ToolPermissions
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

func hasSATv2ReadCapability(capabilities []string, domain SATv2Domain) bool {
	readCap := "xconf:" + string(domain) + ":readonly"

	// metrics has no readwrite capability; only xconf:metrics:readonly is valid
	if domain == DOMAIN_METRICS {
		return util.Contains(capabilities, readCap)
	}

	writeCap := "xconf:" + string(domain) + ":readwrite"
	return util.Contains(capabilities, readCap) || util.Contains(capabilities, writeCap)
}

func hasSATv2WriteCapability(capabilities []string, domain SATv2Domain) bool {
	// metrics has no write capability
	if domain == DOMAIN_METRICS {
		return false
	}

	writeCap := "xconf:" + string(domain) + ":readwrite"
	return util.Contains(capabilities, writeCap)
}

func getTenantIdForSATv2(r *http.Request) string {
	return xhttp.GetTenantIdFromHeader(r)
}

func authorizeSATv2TenantScope(r *http.Request) error {
	tenantId := getTenantIdForSATv2(r)
	if util.IsBlank(tenantId) {
		return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "Missing tenantId for SAT v2 authorization")
	}

	allowedPartners := xhttp.GetAllowedPartnersFromContext(r)
	if len(allowedPartners) == 0 {
		return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "SAT token is missing allowed partners")
	}
	if !util.CaseInsensitiveContains(allowedPartners, tenantId) {
		return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "SAT token is not allowed for tenant "+tenantId)
	}

	return nil
}

func CanReadSatV2(r *http.Request, capabilities []string, applicationType string) (string, error) {
	domain, ok := classifySATv2Domain(r.URL.Path)
	if !ok {
		return "", xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No SAT v2 read permission for unmapped route")
	}

	if !hasSATv2ReadCapability(capabilities, domain) {
		requiredCap := "xconf:" + string(domain) + ":readonly"
		return "", xwcommon.NewRemoteErrorAS(http.StatusForbidden, fmt.Sprintf("SAT v2 token is missing required capability: %s", requiredCap))
	}
	if err := authorizeSATv2TenantScope(r); err != nil {
		return "", err
	}

	return applicationType, nil
}

func CanWriteSatV2(r *http.Request, capabilities []string, applicationType string) (string, error) {
	domain, ok := classifySATv2Domain(r.URL.Path)
	if !ok {
		return "", xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No SAT v2 write permission for unmapped route")
	}

	if !hasSATv2WriteCapability(capabilities, domain) {
		requiredCap := "xconf:" + string(domain) + ":readwrite"
		return "", xwcommon.NewRemoteErrorAS(http.StatusForbidden, fmt.Sprintf("SAT v2 token is missing required capability: %s", requiredCap))
	}
	if err := authorizeSATv2TenantScope(r); err != nil {
		return "", err
	}

	return applicationType, nil
}

func resolveApplicationType(r *http.Request, entityType string, vargs ...string) (string, error) {
	if entityType == COMMON_ENTITY || entityType == TOOL_ENTITY {
		return "", nil
	}

	applicationType := ""
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
			log.Debugf("applicationType not specified: auth_subject=%s path=%s", r.Header.Get(xhttp.AUTH_SUBJECT), r.URL.Path)
			applicationType = core.STB
		}
	}

	if err := core.ValidateApplicationType(applicationType); err != nil {
		return "", err
	}

	return applicationType, nil
}

func authorizeWrite(r *http.Request, entityType string, applicationType string, authType interface{}) error {
	if !(owcommon.SatOn) {
		return nil
	}

	if authType == xhttp.AUTH_TYPE_SAT_V2 || authType == xhttp.AUTH_TYPE_SAT_LEGACY {
		// get capabilities from SAT token if available, return error if none found
		capabilities := xhttp.GetCapabilitiesFromContext(r)
		if len(capabilities) == 0 {
			return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No capabilities found in SAT token")
		}
		// if SAT is v2, do v2 check
		if authType == xhttp.AUTH_TYPE_SAT_V2 {
			_, err := CanWriteSatV2(r, capabilities, applicationType)
			return err
		}
		// else assume legacy SAT
		if entityType == COMMON_ENTITY && util.Contains(capabilities, XCONF_WRITE_MACLIST) {
			return nil
		}
		if !(util.Contains(capabilities, XCONF_ALL) || util.Contains(capabilities, XCONF_WRITE)) {
			return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No write capabilities")
		}
		return nil
	}

	// check permissions from Login token since SAT token is not available
	permissions := GetPermissionsFunc(r)
	if util.Contains(permissions, getEntityPermission(entityType).WriteAll) {
		return nil
	}
	if util.Contains(common.ApplicationTypes, applicationType) && util.Contains(permissions, getEntityPermission(entityType).Write+applicationType) {
		return nil
	}

	// if we get here, it means user doesn't have required permissions, return error
	if applicationType == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No write permission")
	}
	return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No write permission for ApplicationType "+applicationType)
}

func authorizeRead(r *http.Request, entityType string, applicationType string, authType interface{}) error {
	if !(owcommon.SatOn) {
		return nil
	}

	if authType == xhttp.AUTH_TYPE_SAT_V2 || authType == xhttp.AUTH_TYPE_SAT_LEGACY {
		// get capabilities from SAT token if available, return error if none found
		capabilities := xhttp.GetCapabilitiesFromContext(r)
		if len(capabilities) == 0 {
			return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No capabilities found in SAT token")
		}
		// if SAT is v2, do v2 check
		if authType == xhttp.AUTH_TYPE_SAT_V2 {
			_, err := CanReadSatV2(r, capabilities, applicationType)
			return err
		}
		// else assume legacy SAT
		if entityType == COMMON_ENTITY && util.Contains(capabilities, XCONF_READ_MACLIST) {
			return nil
		}
		if !(util.Contains(capabilities, XCONF_ALL) || util.Contains(capabilities, XCONF_READ)) {
			return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No read capabilities")
		}
		return nil
	}

	// check permissions from Login token since SAT token is not available
	permissions := GetPermissionsFunc(r)
	if util.Contains(permissions, getEntityPermission(entityType).ReadAll) {
		return nil
	}

	if util.Contains(common.ApplicationTypes, applicationType) && util.Contains(permissions, getEntityPermission(entityType).Read+applicationType) {
		return nil
	}

	// if we get here, it means user doesn't have required permissions, return error
	if applicationType == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No read permission")
	}
	return xwcommon.NewRemoteErrorAS(http.StatusForbidden, "No read permission for ApplicationType "+applicationType)
}

// CanWrite returns the applicationType the user has write permission for non-common entityType,
// otherwise returns error if applicationType is not specified in query parameter or cookie
func CanWrite(r *http.Request, entityType string, vargs ...string) (applicationType string, err error) {
	authType := r.Context().Value(xhttp.CTX_KEY_AUTH_TYPE)

	applicationType, err = resolveApplicationType(r, entityType, vargs...)
	if err != nil {
		return "", err
	}

	if err = authorizeWrite(r, entityType, applicationType, authType); err != nil {
		return "", err
	}

	// Lockdown check runs after authorization: unauthorized callers should receive
	// 401/403, not a 423 that reveals operational system state.
	tenantId := xhttp.GetTenantId(r)
	if isLockdownMode(tenantId) {
		lockdownModules := strings.Split(common.GetStringAppSetting(tenantId, common.PROP_LOCKDOWN_MODULES), ",")
		if len(lockdownModules) != 0 {
			if util.CaseInsensitiveContains(lockdownModules, getCurrentModule(r, entityType)) || strings.ToUpper(lockdownModules[0]) == common.DefaultLockdownModules {
				return "", xwcommon.NewRemoteErrorAS(http.StatusLocked, "Modification not allowed in Lockdown mode")
			}
		}
	}

	return applicationType, nil
}

// CanRead returns the applicationType the user has read permission for non-common entityType,
// otherwise returns error if applicationType is not specified in query parameter or cookie
func CanRead(r *http.Request, entityType string, vargs ...string) (applicationType string, err error) {
	authType := r.Context().Value(xhttp.CTX_KEY_AUTH_TYPE)

	applicationType, err = resolveApplicationType(r, entityType, vargs...)
	if err != nil {
		return "", err
	}

	if err = authorizeRead(r, entityType, applicationType, authType); err != nil {
		return "", err
	}

	return applicationType, nil
}

func classifySATv2Domain(path string) (SATv2Domain, bool) {
	path = strings.ToLower(strings.TrimSuffix(path, "/"))

	// tagging paths are not under xconfAdminService; check registry before stripping prefix
	for _, m := range satV2RouteMappings {
		if strings.HasPrefix(path, m.Prefix) {
			return m.Domain, true
		}
	}

	// strip the xconfAdminService prefix and re-check for admin routes
	adminPath := strings.TrimPrefix(path, "/xconfadminservice")
	if adminPath == path {
		// no prefix was stripped; no match found above
		return "", false
	}

	for _, m := range satV2RouteMappings {
		if strings.HasPrefix(adminPath, m.Prefix) {
			return m.Domain, true
		}
	}

	return "", false
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

func isLockdownMode(tenantId string) bool {
	if owcommon.GetBooleanAppSetting(tenantId, owcommon.PROP_LOCKDOWN_ENABLED, false) {
		startTime := owcommon.GetStringAppSetting(tenantId, owcommon.PROP_LOCKDOWN_STARTTIME)
		endTime := owcommon.GetStringAppSetting(tenantId, owcommon.PROP_LOCKDOWN_ENDTIME)

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

func GetDistributedLockOwner(r *http.Request) (owner string) {
	owner = r.Header.Get(xhttp.AUTH_SUBJECT)
	if owner == "" {
		owner = uuid.New().String()
		log.Warnf("Unknown user; setting lock owner to a random UUID: %s", owner)
	}
	return
}

func ExtractBodyAndCheckPermissions(obj owcommon.ApplicationTypeAware, w http.ResponseWriter, r *http.Request, entityType string) (applicationType string, err error) {
	applicationType, err = CanWrite(r, entityType, obj.GetApplicationType())
	if err != nil {
		return "", err
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		return "", xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "responsewriter cast error")
	}
	body := xw.Body()
	err = json.Unmarshal([]byte(body), &obj)
	if err != nil {
		return "", xwcommon.NewRemoteErrorAS(http.StatusBadRequest, err.Error())
	}

	if obj.GetApplicationType() == "" {
		obj.SetApplicationType(applicationType)
	} else if obj.GetApplicationType() != applicationType {
		return "", xwcommon.NewRemoteErrorAS(http.StatusConflict, "ApplicationType Conflict")
	}
	return applicationType, nil
}

func isReadonlyMode(tenantId string) bool {
	return owcommon.GetBooleanAppSetting(tenantId, owcommon.READONLY_MODE, false)
}
