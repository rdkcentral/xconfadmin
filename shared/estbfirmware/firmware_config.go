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
package estbfirmware

import (
	"xconfwebconfig/db"
	coreef "xconfwebconfig/shared/estbfirmware"
)

func AddExpressionToIpRuleBean(ipRuleBean *coreef.IpRuleBean) {
	expression := coreef.Expression{}
	expression.EnvironmentId = ipRuleBean.EnvironmentId
	expression.IpAddressGroup = ipRuleBean.IpAddressGroup
	expression.ModelId = ipRuleBean.ModelId
	expression.TargetedModelIds = []string{}
	ipRuleBean.Expression = &expression
}

func GetFirmwareConfigAsMapDB(applicationType string) (configMap map[string]coreef.FirmwareConfig, err error) {
	rulelst, ok := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FIRMWARE_CONFIG, 0)
	if ok != nil {
		return nil, err
	}

	configMap = make(map[string]coreef.FirmwareConfig)

	for _, r := range rulelst {
		cfg, ok := r.(*coreef.FirmwareConfig)
		if !ok {
			continue
		}
		if cfg.ApplicationType == applicationType {
			configMap[cfg.ID] = *cfg
		}
	}

	return configMap, nil
}
