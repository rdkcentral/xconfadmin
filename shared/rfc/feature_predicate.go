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
package rfc

import (
	"strings"

	xwrfc "github.com/rdkcentral/xconfwebconfig/shared/rfc"
)

const (
	APPLICATION_TYPE = "applicationType"
	FEATURE_INSTANCE = "FEATURE_INSTANCE"
	NAME             = "NAME"
	FREE_ARG         = "FREE_ARG"
	FIXED_ARG        = "FIXED_ARG"
)

func isFeatureValid(feature *xwrfc.Feature, predicates []string, searchContext map[string]string) bool {
	for _, predicate := range predicates {
		switch predicate {
		case APPLICATION_TYPE:
			if !isApplicationTypeValid(searchContext[APPLICATION_TYPE], feature) {
				return false
			}
		case FEATURE_INSTANCE:
			if !isFeatureInstanceValid(searchContext[FEATURE_INSTANCE], feature) {
				return false
			}
		case NAME:
			if !isNameValid(searchContext[NAME], feature) {
				return false
			}
		case FREE_ARG:
			if !isFreeArgValid(searchContext[FREE_ARG], feature) {
				return false
			}
		case FIXED_ARG:
			if !isFixedArgValid(searchContext[FIXED_ARG], feature) {
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
	if context[APPLICATION_TYPE] != "" {
		predicates = append(predicates, APPLICATION_TYPE)
	}
	if context[FEATURE_INSTANCE] != "" {
		predicates = append(predicates, FEATURE_INSTANCE)
	}
	if context[NAME] != "" {
		predicates = append(predicates, NAME)
	}
	if context[FREE_ARG] != "" {
		predicates = append(predicates, FREE_ARG)
	}
	if context[FIXED_ARG] != "" {
		predicates = append(predicates, FIXED_ARG)
	}
	return predicates
}
