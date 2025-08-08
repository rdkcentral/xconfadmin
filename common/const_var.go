package common

import (
	"time"

	"github.com/rdkcentral/xconfadmin/util"
)

var SatOn bool
var ActiveAuthProfiles string
var DefaultAuthProfiles string
var IpMacIsConditionLimit int
var AllowedNumberOfFeatures int
var CacheUpdateWindowSize int64
var LockDuration int32
var VideoCanaryCreationEnabled bool
var CanaryCreationEnabled bool
var CanaryStartTime string
var CanaryEndTime string
var CanaryTimeFormat string
var CanaryTimezone *time.Location
var CanaryTimezoneList []string
var CanaryDefaultPartner string
var CanarySize int
var CanaryDistributionPercentage float64
var CanaryFwUpgradeStartTime int
var CanaryFwUpgradeEndTime int
var CanaryPercentFilterNameSet = util.Set{}
var CanaryVideoModelSet = util.Set{}
var CanarySyndicatePartnerSet = util.Set{}
var CanaryWakeupPercentFilterNameSet = util.Set{}
var Wakeuppool_tag_name string
var AuthProvider string
var ApplicationTypes []string

const (
	DATE_TIME_FORMATTER = "1/2/2006 15:04"
	LAST_CONFIG_LOG_ID  = "0"
)

const (
	ENTITY_STATUS_SUCCESS = "SUCCESS"
	ENTITY_STATUS_FAILURE = "FAILURE"
)

const (
	XCONF_HTTP_HEADER       = "HA-Haproxy-xconf-http"
	XCONF_HTTP_VALUE        = "xconf-http"
	X_FORWARDED_FOR_HEADER  = "X-Forwarded-For"
	HA_FORWARDED_FOR_HEADER = "HA-Forwarded-For"
)

const (
	HOST_MAC_PARAM = "hostMac"
	ECM_MAC_PARAM  = "ecmMac"
)

const (
	DEFAULT_AUTH_PROVIDER = "acl"
)
const (
	READ_COMMON         string = "read-common"
	WRITE_COMMON        string = "write-common"
	READ_FIRMWARE_ALL   string = "read-firmware-*"
	WRITE_FIRMWARE_ALL  string = "write-firmware-*"
	WRITE_DCM_ALL       string = "write-dcm-*"
	READ_DCM_ALL        string = "read-dcm-*"
	READ_TELEMETRY_ALL  string = "read-telemetry-*"
	WRITE_TELEMETRY_ALL string = "write-telemetry-*"
	WRITE_CHANGES_ALL   string = "write-changes-*"
	READ_CHANGES_ALL    string = "read-changes-*"
)

// TODO: group constants by their usage, for now everithing is in one list: search related, rule evaluation, pagination, etc
const (
	ID                     = "id"
	IP_ADDRESS             = "ipAddress"
	ESTB_IP                = "estbIP"
	ESTB_MAC_ADDRESS       = "estbMacAddress"
	STB_ESTB_MAC           = "eStbMac"
	ECM_MAC_ADDRESS        = "ecmMacAddress"
	STB_ECM_MAC            = "eCMMac"
	ENV                    = "env"
	MODEL                  = "model"
	MODEL_ID               = "modelId"
	ACCOUNT_MGMT           = "accountMgmt"
	SERIAL_NUM             = "serialNum"
	PARTNER_ID             = "partnerId"
	PASSED_PARTNER_ID      = "passedPartnerId"
	FIRMWARE_VERSION       = "firmwareVersion"
	RECEIVER_ID            = "receiverId"
	CONTROLLER_ID          = "controllerId"
	CHANNEL_MAP_ID         = "channelMapId"
	VOD_ID                 = "vodId"
	BYPASS_FILTERS         = "bypassFilters"
	FORCE_FILTERS          = "forceFilters"
	UPLOAD_IMMEDIATELY     = "uploadImmediately"
	TIME_ZONE              = "timezone"
	TIME_ZONE_OFFSET       = "timeZoneOffset"
	SCHEDULE_TIME_ZONE     = "scheduleTimezone"
	TIME                   = "time"
	ACCOUNT_ID             = "accountId"
	ACCOUNT_HASH           = "accountHash"
	CONFIG_SET_HASH        = "configSetHash"
	SYNDICATION_PARTNER    = "SyndicationPartner"
	MAC                    = "mac"
	CHECK_NOW              = "checkNow"
	VERSION                = "version"
	SETTING_TYPE           = "settingType"
	TABLE_NAME             = "tableName"
	ROW_KEY                = "rowKey"
	FIELD                  = "field"
	NAME                   = "name"
	LIST_ID                = "listId"
	RULE_NAME              = "ruleName"
	MAC_ADDRESS            = "macAddress"
	IP_ADDRESS_GROUP_NAME  = "ipAddressGroupName"
	FEATURE_INSTANCE       = "FEATURE_INSTANCE"
	FREE_ARG               = "FREE_ARG"
	FIXED_ARG              = "FIXED_ARG"
	NAME_UPPER             = "NAME"
	EXPORT                 = "export"
	EXPORTALL              = "exportAll"
	TYPE                   = "type"
	TYPE_UPPER             = "TYPE"
	DATA_UPPER             = "DATA"
	EDITABLE               = "isEditable"
	OVERWRITE              = "overwrite"
	NEW_PRIORITY           = "newPriority"
	APPLICABLE_ACTION_TYPE = "APPLICABLE_ACTION_TYPE"
	TEMPLATE_ID            = "templateId"
	CHANGE_ID              = "changeId"
	APPROVE_ID             = "approveId"
	PAGE_NUMBER            = "pageNumber"
	PAGE_SIZE              = "pageSize"
	DESCRIPTION            = "description"
	ENTITY                 = "ENTITY"
	AUTHOR                 = "AUTHOR"
	PROFILE                = "PROFILE"
	PROFILE_NAME           = "profilename"
	FULL                   = "full"
)
const (
	TR181_DEVICE_TYPE_PARTNER_ID   = "tr181.Device.DeviceInfo.X_RDKCENTRAL-COM_Syndication.PartnerId"
	TR181_DEVICE_TYPE_ACCOUNT_ID   = "tr181.Device.DeviceInfo.X_RDKCENTRAL-COM_RFC.Feature.AccountInfo.AccountID"
	TR181_DEVICE_TYPE_ACCOUNT_HASH = "tr181.Device.DeviceInfo.X_RDKCENTRAL-COM_RFC.Feature.MD5AccountHash"
)

const (
	GenericNamespacedListTypes_STRING      = "STRING"
	GenericNamespacedListTypes_MAC_LIST    = "MAC_LIST"
	GenericNamespacedListTypes_IP_LIST     = "IP_LIST"
	GenericNamespacedListTypes_RI_MAC_LIST = "RI_MAC_LIST"
)

const (
	DefaultTimeDateFormatLayout   = "2006-01-02 15:04"
	DefaultDateFormatLayout       = "2006-01-02"
	DefaultTimeFormatLayout       = "15:04"
	DefaultLockdownStartTime      = "19:00" //EST Timezone
	DefaultLockdownEndTime        = "07:00" //EST Timezone
	DefaultLockdownTimezone       = "America/New_York"
	DefaultCanaryTimezone         = "America/New_York"
	DefaultLockdownModules        = "ALL"
	DefaultLockDuration           = 1800
	DefaultPrecookLockdownEnabled = false
)

const (
	PROP_LOCKDOWN_ENABLED               = "LockdownEnabled"
	PROP_LOCKDOWN_MODULES               = "LockdownModules"
	PROP_CANARY_MAXSIZE                 = "CanaryMaxSize"
	PROP_CANARY_DISTRIBUTION_PERCENTAGE = "CanaryDistributionPercentage"
	PROP_CANARY_FW_UPGRADE_STARTTIME    = "CanaryFwUpgradeStartTime"
	PROP_CANARY_FW_UPGRADE_ENDTIME      = "CanaryFwUpgradeEndTime"
	PROP_READONLY_PERMISSION            = "ReadOnlyPermission"
	PROP_LOCKDOWN_STARTTIME             = "LockdownStartTime"
	PROP_LOCKDOWN_ENDTIME               = "LockdownEndTime"
	PROP_PRECOOK_LOCKDOWN_ENABLED       = "PrecookLockdownEnabled"
	PROP_CANARY_TIMEZONE_LIST           = "CanaryTimezoneList"
)

const (
	Member     = "member"
	Tag        = "tag"
	StartRange = "startRange"
	EndRange   = "endRange"
)

const (
	READONLY_MODE           = "ReadonlyMode"
	READONLY_MODE_STARTTIME = "ReadonlyModeStartTime"
	READONLY_MODE_ENDTIME   = "ReadonlyModeEndTime"
)

// db
const (
	TABLE_APP_SETTINGS = "AppSettings"
)
const (
	HeaderAuthorization        = "Authorization"
	HeaderUserAgent            = "User-Agent"
	HeaderIfNoneMatch          = "If-None-Match"
	HeaderFirmwareVersion      = "X-System-Firmware-Version"
	HeaderSupportedDocs        = "X-System-Supported-Docs"
	HeaderSupplementaryService = "X-System-SupplementaryService-Sync"
	HeaderModelName            = "X-System-Model-Name"
	HeaderProfileVersion       = "X-System-Telemetry-Profile-Version"
	HeaderPartnerID            = "X-System-PartnerID"
	HeaderAccountID            = "X-System-AccountID"
	HeaderXconfDataService     = "XconfDataService"
	HeaderXconfAdminService    = "XconfAdminService"
)

// const (
// 	APPROVE_ID             = "approveId"
// 	CHANGE_ID              = "changeId"
// 	PAGE_NUMBER            = "pageNumber"
// 	PAGE_SIZE              = "pageSize"
// 	AUTHOR                 = "AUTHOR"
// 	ENTITY                 = "ENTITY"
// 	PROFILE_NAME           = "profilename"
// 	NAME_UPPER             = "NAME"
// 	EXPORT                 = "export"
// 	APPLICATION_TYPE       = "applicationType"
// 	FEATURE_INSTANCE       = "FEATURE_INSTANCE"
// 	NAME                   = "name"
// 	FREE_ARG               = "FREE_ARG"
// 	FIXED_ARG              = "FIXED_ARG"
// 	TYPE                   = "type"
// 	PROFILE                = "PROFILE"
// 	OVERWRITE              = "overwrite"
// 	NEW_PRIORITY           = "newPriority"
// 	STB_ESTB_MAC           = "eStbMac"
// 	MAC_ADDRESS            = "macAddress"
// 	SCHEDULE_TIME_ZONE     = "scheduleTimezone"
// 	EXPORTALL              = "exportAll"
// 	TYPE_UPPER             = "TYPE"
// 	DATA_UPPER             = "DATA"
// 	ROW_KEY                = "rowKey"
// 	RULE_NAME              = "ruleName"
// 	IP_ADDRESS_GROUP_NAME  = "ipAddressGroupName"
// 	EDITABLE               = "isEditable"
// 	APPLICABLE_ACTION_TYPE = "APPLICABLE_ACTION_TYPE"
// )

var AllAppSettings = []string{
	READONLY_MODE,
	READONLY_MODE_STARTTIME,
	READONLY_MODE_ENDTIME,
	PROP_LOCKDOWN_ENABLED,
	PROP_LOCKDOWN_MODULES,
	PROP_CANARY_MAXSIZE,
	PROP_CANARY_DISTRIBUTION_PERCENTAGE,
	PROP_CANARY_FW_UPGRADE_STARTTIME,
	PROP_CANARY_FW_UPGRADE_ENDTIME,
	PROP_LOCKDOWN_STARTTIME,
	PROP_LOCKDOWN_ENDTIME,
	PROP_PRECOOK_LOCKDOWN_ENABLED,
	PROP_CANARY_TIMEZONE_LIST,
}

const (
	ExportFileNames_ALL                             = "all"
	ExportFileNames_FIRMWARE_CONFIG                 = "firmwareConfig_"
	ExportFileNames_ALL_FIRMWARE_CONFIGS            = "allFirmwareConfigs"
	ExportFileNames_FIRMWARE_RULE                   = "firmwareRule_"
	ExportFileNames_ALL_FIRMWARE_RULES              = "allFirmwareRules"
	ExportFileNames_FIRMWARE_RULE_TEMPLATE          = "firmwareRuleTemplate_"
	ExportFileNames_ALL_FIRMWARE_RULE_TEMPLATES     = "allFirmwareRuleTemplates"
	ExportFileNames_ALL_PERMANENT_PROFILES          = "allPermanentProfiles"
	ExportFileNames_PERMANENT_PROFILE               = "permanentProfile_"
	ExportFileNames_ALL_TELEMETRY_RULES             = "allTelemetryRules"
	ExportFileNames_ALL_TELEMETRY_TWO_RULES         = "allTelemetryTwoRules"
	ExportFileNames_TELEMETRY_RULE                  = "telemetryRule_"
	ExportFileNames_TELEMETRY_TWO_RULE              = "telemetryTwoRule_"
	ExportFileNames_TELEMETRY_TWO_PROFILE           = "telemetryTwoProfile_"
	ExportFileNames_ALL_TELEMETRY_TWO_PROFILES      = "allTelemetryTwoProfiles"
	ExportFileNames_ALL_SETTING_PROFILES            = "allSettingProfiles"
	ExportFileNames_SETTING_PROFILE                 = "settingProfile_"
	ExportFileNames_ALL_SETTING_RULES               = "allSettingRules"
	ExportFileNames_SETTING_RULE                    = "settingRule_"
	ExportFileNames_ALL_FORMULAS                    = "allFormulas"
	ExportFileNames_FORMULA                         = "formula_"
	ExportFileNames_ALL_ENVIRONMENTS                = "allEnvironments"
	ExportFileNames_ENVIRONMENT                     = "environment_"
	ExportFileNames_ALL_MODELS                      = "allModels"
	ExportFileNames_MODEL                           = "model_"
	ExportFileNames_UPLOAD_REPOSITORY               = "uploadRepository_"
	ExportFileNames_ALL_UPLOAD_REPOSITORIES         = "allUploadRepositories"
	ExportFileNames_ROUND_ROBIN_FILTER              = "roundRobinFilter"
	ExportFileNames_GLOBAL_PERCENT                  = "globalPercent"
	ExportFileNames_GLOBAL_PERCENT_AS_RULE          = "globalPercentAsRule"
	ExportFileNames_ENV_MODEL_PERCENTAGE_BEANS      = "envModelPercentageBeans"
	ExportFileNames_ENV_MODEL_PERCENTAGE_BEAN       = "envModelPercentageBean_"
	ExportFileNames_ENV_MODEL_PERCENTAGE_AS_RULES   = "envModelPercentageAsRules"
	ExportFileNames_ENV_MODEL_PERCENTAGE_AS_RULE    = "envModelPercentageAsRule_"
	ExportFileNames_PERCENT_FILTER                  = "percentFilter"
	ExportFileNames_ALL_NAMESPACEDLISTS             = "allNamespacedLists"
	ExportFileNames_NAMESPACEDLIST                  = "namespacedList_"
	ExportFileNames_ALL_FEATURES                    = "allFeatures"
	ExportFileNames_FEATURE                         = "feature_"
	ExportFileNames_ALL_FEATURE_SETS                = "allFeatureSets"
	ExportFileNames_FEATURE_SET                     = "featureSet_"
	ExportFileNames_ALL_FEATURE_RUlES               = "allFeatureRules"
	ExportFileNames_FEATURE_RULE                    = "featureRule_"
	ExportFileNames_ACTIVATION_MINIMUM_VERSION      = "activationMinimumVersion_"
	ExportFileNames_ALL_ACTIVATION_MINIMUM_VERSIONS = "allActivationMinimumVersions"
	ExportFileNames_ALL_DEVICE_SETTINGS             = "allDeviceSettings"
	ExportFileNames_ALL_VOD_SETTINGS                = "allVodSettings"
	ExportFileNames_ALL_LOGREPO_SETTINGS            = "allLogRepoSettings"
)

var (
	SupportedPokeDocs = []string{"primary", "telemetry"}
)

var (
	BinaryVersion   = ""
	BinaryBranch    = ""
	BinaryBuildTime = ""

	DefaultIgnoredHeaders = []string{
		"Accept",
		"User-Agent",
		"Authorization",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-B3-Sampled",
		"X-B3-Spanid",
		"X-B3-Traceid",
		"X-Envoy-Decorator-Operation",
		"X-Envoy-External-Address",
		"X-Envoy-Peer-Metadata",
		"X-Envoy-Peer-Metadata-Id",
		"X-Forwarded-Proto",
		"Token",
		"Cookie",
		"Set-Cookie",
	}
)

func IsValidAppSetting(key string) bool {
	return util.Contains(AllAppSettings, key)
}

func isValidType(namespacedListType string) bool {
	if namespacedListType == GenericNamespacedListTypes_STRING ||
		namespacedListType == GenericNamespacedListTypes_MAC_LIST ||
		namespacedListType == GenericNamespacedListTypes_IP_LIST ||
		namespacedListType == GenericNamespacedListTypes_RI_MAC_LIST {
		return true
	}
	return false
}
