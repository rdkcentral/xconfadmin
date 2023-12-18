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
	coreef "xconfwebconfig/shared/estbfirmware"
)

func NewPercentFilterWrapper(percentFilterValue *coreef.PercentFilterValue, toHumanReadableForm bool) *coreef.PercentFilterWrapper {
	wrapper := coreef.PercentFilterWrapper{
		ID:         percentFilterValue.ID,
		Type:       coreef.PercentFilterWrapperClass,
		Whitelist:  percentFilterValue.Whitelist,
		Percentage: float64(percentFilterValue.Percentage),
	}
	for k, v := range percentFilterValue.EnvModelPercentages {
		v.Name = k
		if toHumanReadableForm {
			if v.LastKnownGood != "" {
				v.LastKnownGood = coreef.GetFirmwareVersion(v.LastKnownGood)
			}
			if v.IntermediateVersion != "" {
				v.IntermediateVersion = coreef.GetFirmwareVersion(v.IntermediateVersion)
			}
		}
		wrapper.EnvModelPercentages = append(wrapper.EnvModelPercentages, v)
	}

	return &wrapper
}

func NewEmptyPercentFilterWrapper() *coreef.PercentFilterWrapper {
	return &coreef.PercentFilterWrapper{
		ID:                  coreef.PERCENT_FILTER_SINGLETON_ID,
		Type:                coreef.PercentFilterWrapperClass,
		Percentage:          100.0,
		EnvModelPercentages: []coreef.EnvModelPercentage{},
	}
}
