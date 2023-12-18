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
package common

var SatOn bool
var ActiveAuthProfiles string
var DefaultAuthProfiles string
var IpMacIsConditionLimit int
var AllowedNumberOfFeatures int

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
	APPROVE_ID             = "approveId"
	CHANGE_ID              = "changeId"
	PAGE_NUMBER            = "pageNumber"
	PAGE_SIZE              = "pageSize"
	AUTHOR                 = "AUTHOR"
	ENTITY                 = "ENTITY"
	PROFILE_NAME           = "profilename"
	NAME_UPPER             = "NAME"
	EXPORT                 = "export"
	APPLICATION_TYPE       = "applicationType"
	FEATURE_INSTANCE       = "FEATURE_INSTANCE"
	NAME                   = "name"
	FREE_ARG               = "FREE_ARG"
	FIXED_ARG              = "FIXED_ARG"
	TYPE                   = "type"
	PROFILE                = "PROFILE"
	OVERWRITE              = "overwrite"
	NEW_PRIORITY           = "newPriority"
	STB_ESTB_MAC           = "eStbMac"
	MAC_ADDRESS            = "macAddress"
	SCHEDULE_TIME_ZONE     = "scheduleTimezone"
	EXPORTALL              = "exportAll"
	TYPE_UPPER             = "TYPE"
	DATA_UPPER             = "DATA"
	ROW_KEY                = "rowKey"
	RULE_NAME              = "ruleName"
	IP_ADDRESS_GROUP_NAME  = "ipAddressGroupName"
	EDITABLE               = "isEditable"
	APPLICABLE_ACTION_TYPE = "APPLICABLE_ACTION_TYPE"
)

var AllAppSettings = []string{
	READONLY_MODE,
	READONLY_MODE_STARTTIME,
	READONLY_MODE_ENDTIME,
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

const (
	ENTITY_STATUS_SUCCESS = "SUCCESS"
	ENTITY_STATUS_FAILURE = "FAILURE"
)
