// Copyright 2025 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package shared

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"xconfadmin/common"
	"xconfadmin/util"
	xwcommon "xconfwebconfig/common"
	re "xconfwebconfig/rulesengine"
)

const (
	STB      = "stb"
	RDKCLOUD = "rdkcloud"
	XHOME    = "xhome"
	ALL      = "all"
)

//	type Prioritizable interface {
//		GetPriority() int
//		SetPriority(priority int)
//		GetID() string
//	}

func IsValidApplicationType(at string) bool {
	if len(common.ApplicationTypes) == 0 {
		InitializeApplicationTypes() // Ensuring ApplicationTypes is initialized with default
	}
	return util.Contains(common.ApplicationTypes, at)
}
func InitializeApplicationTypes() {
	common.ApplicationTypes = []string{STB, RDKCLOUD}
}

// Validate whether the ApplicationType is valid if specified
func ValidateApplicationType(applicationType string) error {
	if applicationType == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "ApplicationType is empty")
	}
	if !IsValidApplicationType(applicationType) {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, fmt.Sprintf("ApplicationType %s is not valid", applicationType))
	}
	return nil
}

func ApplicationTypeEquals(type1 string, type2 string) bool {
	if type1 == "" {
		type1 = STB
	}
	if type2 == "" {
		type2 = STB
	}
	return type1 == type2
}

func GetApplicationFromCookies(r *http.Request) string {
	cookie, err := r.Cookie(APPLICATION_TYPE)
	if err != nil || util.IsBlank(cookie.Value) {
		return ""
	}
	return cookie.Value
}

const (
	ESTB_MAC          = "eStbMac"
	ENVIRONMENT       = "env"
	MODEL             = "model"
	FIRMWARE_VERSION  = "firmwareVersion"
	ECM_MAC           = "eCMMac"
	RECEIVER_ID       = "receiverId"
	CONTROLLER_ID     = "controllerId"
	CHANNEL_MAP       = "channelMapId"
	VOD_ID            = "vodId"
	TIME_ZONE         = "timeZone"
	TIME_ZONE_OFFSET  = "timeZoneOffset"
	TIME              = "time"
	IP_ADDRESS        = "ipAddress"
	DOWNLOAD_PROTOCOL = "firmware_download_protocol"
	REBOOT_DECOUPLED  = "rebootDecoupled"
	MATCHED_RULE_TYPE = "matchedRuleType"
	BYPASS_FILTERS    = "bypassFilters"
	FORCE_FILTERS     = "forceFilters"
	CAPABILITIES      = "capabilities"
	PARTNER_ID        = "partnerId"
	ACCOUNT_HASH      = "accountHash"
	ACCOUNT_ID        = "accountId"
	XCONF_HTTP_HEADER = "HA-Haproxy-xconf-http"
)

const (
	ID                         = "id"
	UPDATED                    = "updated"
	DESCRIPTION                = "description"
	SUPPORTED_MODEL_IDS        = "supportedModelIds"
	FIRMWARE_DOWNLOAD_PROTOCOL = "firmwareDownloadProtocol"
	FIRMWARE_FILENAME          = "firmwareFilename"
	FIRMWARE_LOCATION          = "firmwareLocation"
	//FIRMWARE_VERSION = "firmwareVersion"
	IPV6_FIRMWARE_LOCATION = "ipv6FirmwareLocation"
	UPGRADE_DELAY          = "upgradeDelay"
	REBOOT_IMMEDIATELY     = "rebootImmediately"
	PROPERTIES             = "properties"
	APPLICATION_TYPE       = "applicationType"
	MANDATORY_UPDATE       = "mandatoryUpdate"

	// from Java StbContext.FIRMWARE_VERSION
	// also from Java DefinePropertiesAction class
	FIRMWARE_VERSIONS   = "firmwareVersions"
	REGULAR_EXPRESSIONS = "regularExpressions"
)

const (
	TABLE_LOGS_KEY2_FIELD_NAME = "column1"
	LAST_CONFIG_LOG_ID         = "0"
)

const (
	StbContextTime      = "time"
	StbContextModel     = "model"
	MacList             = "MAC_LIST"
	IpList              = "IP_LIST"
	TableGenericNSList  = "GenericXconfNamedList"
	TableFirmwareConfig = "FirmwareConfig"
	TableFirmwareRule   = "FirmwareRule4"
)

const (
	Tftp  = "tftp"
	Http  = "http"
	Https = "https"
)

// XRule is ...
type XRule interface {
	GetId() string
	GetRule() *re.Rule
	GetName() string
	GetTemplateId() string
	GetRuleType() string
}

// XEnvModel is ...
type XEnvModel interface {
	GetId() string
	GetDescription() string
}

// Environment table object
type Environment struct {
	ID          string `json:"id"`
	Updated     int64  `json:"updated"`
	Description string `json:"description"`
}

func (obj *Environment) Clone() (*Environment, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*Environment), nil
}

func (obj *Environment) Validate() error {
	if len(obj.ID) > 0 {
		match, _ := regexp.MatchString("^[-a-zA-Z0-9_.' ]+$", obj.ID)
		if match {
			return nil
		}
	}

	return errors.New("Id is invalid")
}

// NewEnvironmentInf constructor
func NewEnvironmentInf() interface{} {
	return &Environment{}
}

// NewEnvironment ...
func NewEnvironment(id string, description string) *Environment {
	if id != "" {
		id = strings.ToUpper(strings.TrimSpace(id))
	}

	return &Environment{
		ID:          id,
		Description: description,
	}
}

// Model table object
type Model struct {
	ID          string `json:"id"`
	Updated     int64  `json:"updated"`
	Description string `json:"description,omitempty"`
}

func (obj *Model) Clone() (*Model, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*Model), nil
}

func (obj *Model) Validate() error {
	if len(obj.ID) > 0 {
		match, _ := regexp.MatchString("^[-a-zA-Z0-9_.' ]+$", obj.ID)
		if match {
			return nil
		}
	}

	return errors.New("Id is invalid. Valid Characters: alphanumeric _ . -")
}

// NewModelInf constructor
func NewModelInf() interface{} {
	return &Model{}
}

// NewModel ...
func NewModel(id string, description string) *Model {
	return &Model{
		ID:          strings.ToUpper(id),
		Description: description,
	}
}

type ModelResponse struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
}

func (m *Model) CreateModelResponse() *ModelResponse {
	return &ModelResponse{
		ID:          m.ID,
		Description: m.Description,
	}
}

// AppSettings table object
type AppSetting struct {
	ID      string      `json:"id"`
	Updated int64       `json:"updated"`
	Value   interface{} `json:"value"`
}

func (obj *AppSetting) Clone() (*AppSetting, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*AppSetting), nil
}

// NewAppSettingInf constructor
func NewAppSettingInf() interface{} {
	return &AppSetting{}
}

// ConditionInfo is ...
type ConditionInfo struct {
	FreeArg   re.FreeArg
	Operation string
}

// NewConditionInfo create a new instance
func NewConditionInfo(freeArg re.FreeArg, operation string) *ConditionInfo {
	return &ConditionInfo{
		FreeArg:   freeArg,
		Operation: operation,
	}
}

// Prioritizable is ...
type Prioritizable interface {
	GetPriority() int
	SetPriority(priority int)
	GetID() string
}

// StringListWrapper ...
type StringListWrapper struct {
	List []string `json:"list"`
}

func NewStringListWrapper(list []string) *StringListWrapper {
	return &StringListWrapper{List: list}
}

func NormalizeCommonContext(contextMap map[string]string, estbMacKey string, ecmMacKey string) (e error) {
	if model := contextMap[MODEL]; model != "" {
		contextMap[MODEL] = strings.ToUpper(model)
	}
	if env := contextMap[ENVIRONMENT]; env != "" {
		contextMap[ENVIRONMENT] = strings.ToUpper(env)
	}
	if partnerId := contextMap[PARTNER_ID]; partnerId != "" {
		contextMap[PARTNER_ID] = strings.ToUpper(partnerId)
	}
	if mac := contextMap[estbMacKey]; mac != "" {
		if normalizedMac, err := util.ValidateAndNormalizeMacAddress(mac); err != nil {
			e = err
		} else {
			contextMap[estbMacKey] = normalizedMac
		}
	}
	if mac := contextMap[ecmMacKey]; mac != "" {
		if normalizedMac, err := util.ValidateAndNormalizeMacAddress(mac); err != nil {
			e = err
		} else {
			contextMap[ecmMacKey] = normalizedMac
		}
	}
	return e
}
