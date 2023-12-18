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
package rfc

import (
	"strings"
	xcommon "xconfadmin/common"
	xwrfc "xconfwebconfig/shared/rfc"
)

func isFeatureValid(feature *xwrfc.Feature, predicates []string, searchContext map[string]string) bool {
	for _, predicate := range predicates {
		switch predicate {
		case xcommon.APPLICATION_TYPE:
			if !isApplicationTypeValid(searchContext[xcommon.APPLICATION_TYPE], feature) {
				return false
			}
		case xcommon.FEATURE_INSTANCE:
			if !isFeatureInstanceValid(searchContext[xcommon.FEATURE_INSTANCE], feature) {
				return false
			}
		case xcommon.NAME:
			if !isNameValid(searchContext[xcommon.NAME], feature) {
				return false
			}
		case xcommon.FREE_ARG:
			if !isFreeArgValid(searchContext[xcommon.FREE_ARG], feature) {
				return false
			}
		case xcommon.FIXED_ARG:
			if !isFixedArgValid(searchContext[xcommon.FIXED_ARG], feature) {
				return false
			}
		}
	}
	return true
}

func isApplicationTypeValid(applicationType string, feature *xwrfc.Feature) bool {
	return feature != nil && (applicationType == "all" || feature.ApplicationType == applicationType)
}

func isFeatureInstanceValid(featureInstance string, feature *xwrfc.Feature) bool {
	return feature != nil && strings.Contains(strings.ToLower(feature.FeatureName), strings.ToLower(featureInstance))
}

func isNameValid(name string, feature *xwrfc.Feature) bool {
	return feature != nil && strings.Contains(strings.ToLower(feature.Name), strings.ToLower(name))
}

func isFreeArgValid(freeArg string, feature *xwrfc.Feature) bool {
	if feature != nil && feature.ConfigData != nil && len(feature.ConfigData) != 0 {
		for configKey := range feature.ConfigData {
			if strings.Contains(strings.ToLower(configKey), strings.ToLower(freeArg)) {
				return true
			}
		}
	}
	return false
}

func isFixedArgValid(fixedArg string, feature *xwrfc.Feature) bool {
	if feature != nil && feature.ConfigData != nil && len(feature.ConfigData) != 0 {
		for _, configValue := range feature.ConfigData {
			if strings.Contains(strings.ToLower(configValue), strings.ToLower(fixedArg)) {
				return true
			}
		}
	}
	return false
}

func getFeaturePredicates(context map[string]string) []string {
	var predicates []string
	if context[xcommon.APPLICATION_TYPE] != "" {
		predicates = append(predicates, xcommon.APPLICATION_TYPE)
	}
	if context[xcommon.FEATURE_INSTANCE] != "" {
		predicates = append(predicates, xcommon.FEATURE_INSTANCE)
	}
	if context[xcommon.NAME] != "" {
		predicates = append(predicates, xcommon.NAME)
	}
	if context[xcommon.FREE_ARG] != "" {
		predicates = append(predicates, xcommon.FREE_ARG)
	}
	if context[xcommon.FIXED_ARG] != "" {
		predicates = append(predicates, xcommon.FIXED_ARG)
	}
	return predicates
}
