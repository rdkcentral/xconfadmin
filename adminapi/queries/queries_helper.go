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
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	logupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

func NullifyUnwantedFieldsPermanentTelemetryProfile(profile *logupload.PermanentTelemetryProfile) *logupload.PermanentTelemetryProfile {
	if len(profile.TelemetryProfile) > 0 {
		for index := range profile.TelemetryProfile {
			profile.TelemetryProfile[index].ID = ""
			profile.TelemetryProfile[index].Component = ""
		}
	}

	profile.ApplicationType = ""
	return profile
}

type FirmwareConfigResponse struct {
	ID                       string            `json:"id"`
	Updated                  int64             `json:"updated,omitempty"`
	Description              string            `json:"description,omitempty"`
	SupportedModelIds        []string          `json:"supportedModelIds,omitempty"`
	FirmwareFilename         string            `json:"firmwareFilename,omitempty"`
	FirmwareVersion          string            `json:"firmwareVersion,omitempty"`
	ApplicationType          string            `json:"applicationType,omitempty"`
	FirmwareDownloadProtocol string            `json:"firmwareDownloadProtocol,omitempty"`
	FirmwareLocation         string            `json:"firmwareLocation,omitempty"`
	Ipv6FirmwareLocation     string            `json:"ipv6FirmwareLocation,omitempty"`
	UpgradeDelay             int64             `json:"upgradeDelay,omitempty"`
	RebootImmediately        bool              `json:"rebootImmediately,omitempty"`
	MandatoryUpdate          bool              `json:"mandatoryUpdate,omitempty"`
	Properties               map[string]string `json:"properties,omitempty"`
}

func ConvertFirmwareConfigToFirmwareConfigResponse(config *coreef.FirmwareConfig) *FirmwareConfigResponse {
	return &FirmwareConfigResponse{
		ID:                       config.ID,
		Updated:                  config.Updated,
		Description:              config.Description,
		SupportedModelIds:        config.SupportedModelIds,
		FirmwareFilename:         config.FirmwareFilename,
		FirmwareVersion:          config.FirmwareVersion,
		ApplicationType:          config.ApplicationType,
		FirmwareDownloadProtocol: config.FirmwareDownloadProtocol,
		FirmwareLocation:         config.FirmwareLocation,
		Ipv6FirmwareLocation:     config.Ipv6FirmwareLocation,
		UpgradeDelay:             config.UpgradeDelay,
		RebootImmediately:        config.RebootImmediately,
		MandatoryUpdate:          config.MandatoryUpdate,
		Properties:               config.Properties,
	}
}

type IpRuleBeanResponse struct {
	Id             string                  `json:"id,omitempty"`
	FirmwareConfig *FirmwareConfigResponse `json:"firmwareConfig,omitempty"`
	Name           string                  `json:"name,omitempty"`
	IpAddressGroup *shared.IpAddressGroup  `json:"ipAddressGroup,omitempty"`
	EnvironmentId  string                  `json:"environmentId,omitempty"`
	ModelId        string                  `json:"modelId,omitempty"`
	Expression     *coreef.Expression      `json:"expression,omitempty"`
	Noop           bool                    `json:"noop"`
}

func ConvertIpRuleBeanToIpRuleBeanResponse(bean *coreef.IpRuleBean) *IpRuleBeanResponse {
	noop := true
	var firmwareConfigResponse *FirmwareConfigResponse
	if bean.FirmwareConfig != nil {
		noop = false
		firmwareConfigResponse = ConvertFirmwareConfigToFirmwareConfigResponse(bean.FirmwareConfig)
	}
	var expression *coreef.Expression
	if bean.Expression != nil {
		expression = bean.Expression
	} else {
		expression = &coreef.Expression{
			TargetedModelIds: []string{},
			EnvironmentId:    bean.EnvironmentId,
			ModelId:          bean.ModelId,
			IpAddressGroup:   bean.IpAddressGroup,
		}
	}
	return &IpRuleBeanResponse{
		Id:             bean.Id,
		FirmwareConfig: firmwareConfigResponse,
		Name:           bean.Name,
		IpAddressGroup: bean.IpAddressGroup,
		EnvironmentId:  bean.EnvironmentId,
		ModelId:        bean.ModelId,
		Expression:     expression,
		Noop:           noop,
	}
}

type MacRuleBeanResponse struct {
	Id               string                  `json:"id,omitempty"`
	Name             string                  `json:"name,omitempty"`
	MacAddresses     string                  `json:"macAddresses,omitempty"`
	MacListRef       string                  `json:"macListRef,omitempty"`
	FirmwareConfig   *FirmwareConfigResponse `json:"firmwareConfig,omitempty"`
	TargetedModelIds *[]string               `json:"targetedModelIds,omitempty"`
	MacList          *[]string               `json:"macList,omitempty"`
	Noop             bool                    `json:"-"`
}

func ConvertMacRuleBeanToMacRuleBeanResponse(bean *coreef.MacRuleBean) *MacRuleBeanResponse {
	noop := false
	if bean.FirmwareConfig == nil && (bean.TargetedModelIds == nil || len(*bean.TargetedModelIds) == 0) {
		noop = true
	}

	var firmwareConfigResponse *FirmwareConfigResponse
	if bean.FirmwareConfig != nil {
		firmwareConfigResponse = ConvertFirmwareConfigToFirmwareConfigResponse(bean.FirmwareConfig)
	}
	return &MacRuleBeanResponse{
		Id:               bean.Id,
		FirmwareConfig:   firmwareConfigResponse,
		Name:             bean.Name,
		MacAddresses:     bean.MacAddresses,
		MacListRef:       bean.MacListRef,
		TargetedModelIds: bean.TargetedModelIds,
		MacList:          bean.MacList,
		Noop:             noop,
	}
}

type EnvModelRuleBeanResponse struct {
	Id             string                  `json:"id,omitempty"`
	Name           string                  `json:"name,omitempty"`
	EnvironmentId  string                  `json:"environmentId,omitempty"`
	ModelId        string                  `json:"modelId,omitempty"`
	FirmwareConfig *FirmwareConfigResponse `json:"firmwareConfig,omitempty"`
	Noop           bool                    `json:"-"`
}

func ConvertEnvModelRuleBeanToEnvModelRuleBeanResponse(bean *coreef.EnvModelBean) *EnvModelRuleBeanResponse {
	var firmwareConfigResponse *FirmwareConfigResponse
	if bean.FirmwareConfig != nil {
		firmwareConfigResponse = ConvertFirmwareConfigToFirmwareConfigResponse(bean.FirmwareConfig)
	}
	return &EnvModelRuleBeanResponse{
		Id:             bean.Id,
		FirmwareConfig: firmwareConfigResponse,
		Name:           bean.Name,
		EnvironmentId:  bean.EnvironmentId,
		ModelId:        bean.ModelId,
	}
}
