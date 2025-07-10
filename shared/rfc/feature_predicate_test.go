/*
 * If not stated otherwise in this file or this component's Licenses.txt file the
 * following copyright and licenses apply:
 *
 * Copyright 2018 RDK Management
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
 * Author: kloder
 * Created: 11/3/2021
 */
package rfc

import (
	"testing"
	"xconfwebconfig/shared/rfc"

	"gotest.tools/assert"
)

func TestGetFeaturePredicates(t *testing.T) {
	// empty context
	var context map[string]string
	predicates := getFeaturePredicates(context)
	assert.Equal(t, len(predicates), 0)

	// non-empty context with no predicates
	context = map[string]string{
		"key": "value",
	}
	predicates = getFeaturePredicates(context)
	assert.Equal(t, len(predicates), 0)

	// contains some predicates
	context["applicationType"] = "applicationType"
	context["FEATURE_INSTANCE"] = "featureName"
	predicates = getFeaturePredicates(context)
	assert.Equal(t, len(predicates), 2)
	assert.Equal(t, predicates[0], "applicationType")
	assert.Equal(t, predicates[1], "FEATURE_INSTANCE")

	// all predicates
	context["NAME"] = "name"
	context["FREE_ARG"] = "freeArg"
	context["FIXED_ARG"] = "fixedArg"
	predicates = getFeaturePredicates(context)
	assert.Equal(t, len(predicates), 5)
	assert.Equal(t, predicates[0], "applicationType")
	assert.Equal(t, predicates[1], "FEATURE_INSTANCE")
	assert.Equal(t, predicates[2], "NAME")
	assert.Equal(t, predicates[3], "FREE_ARG")
	assert.Equal(t, predicates[4], "FIXED_ARG")
}

func TestIsFeatureValid(t *testing.T) {
	// all empty values
	var predicates []string
	var searchContext map[string]string
	var feature *rfc.Feature
	isValid := isFeatureValid(feature, predicates, searchContext)
	assert.Equal(t, isValid, true)

	// not empty, invalid feature
	searchContext = map[string]string{
		"APPLICATION_NAME": "stb",
		"FEATURE_INSTANCE": "featureInstance",
		"NAME":             "name",
		"FREE_ARG":         "freeArg",
		"FIXED_ARG":        "fixedArg",
	}
	predicates = []string{"APPLICATION_NAME", "FEATURE_INSTANCE", "NAME", "FREE_ARG", "FIXED_ARG"}
	isValid = isFeatureValid(feature, predicates, searchContext)
	assert.Equal(t, isValid, false)

	// feature has some attributes, but not all
	feature = &rfc.Feature{
		ApplicationType: "stb",
		FeatureName:     "featureInstance",
		Name:            "name",
	}
	isValid = isFeatureValid(feature, predicates, searchContext)
	assert.Equal(t, isValid, false)

	// valid feature
	feature = &rfc.Feature{
		ApplicationType: "stb",
		FeatureName:     "featureInstance",
		Name:            "name",
		ConfigData: map[string]string{
			"test":    "test",
			"freeArg": "fixedArg",
		},
	}
	isValid = isFeatureValid(feature, predicates, searchContext)
	assert.Equal(t, isValid, true)
}

func TestIsApplicationTypeValid(t *testing.T) {
	// empty values
	var feature *rfc.Feature
	isValid := isApplicationTypeValid("", feature)
	assert.Equal(t, isValid, false)

	// not valid
	isValid = isApplicationTypeValid("applicationType", feature)
	assert.Equal(t, isValid, false)

	// valid
	feature = &rfc.Feature{
		ApplicationType: "applicationType",
	}
	isValid = isApplicationTypeValid("applicationType", feature)
	assert.Equal(t, isValid, true)

	// valid
	isValid = isApplicationTypeValid("all", feature)
	assert.Equal(t, isValid, true)
}

func TestIsFeatureInstanceValid(t *testing.T) {
	// empty values
	var feature *rfc.Feature
	isValid := isFeatureInstanceValid("", feature)
	assert.Equal(t, isValid, false)

	// not valid
	isValid = isFeatureInstanceValid("featureInstance", feature)
	assert.Equal(t, isValid, false)

	// valid
	feature = &rfc.Feature{
		FeatureName: "featureInstance",
	}
	isValid = isFeatureInstanceValid("featureInstance", feature)
	assert.Equal(t, isValid, true)
}

func TestIsNameValid(t *testing.T) {
	// empty values
	var feature *rfc.Feature
	isValid := isNameValid("", feature)
	assert.Equal(t, isValid, false)

	// not valid
	isValid = isNameValid("name", feature)
	assert.Equal(t, isValid, false)

	// valid
	feature = &rfc.Feature{
		Name: "name",
	}
	isValid = isNameValid("name", feature)
	assert.Equal(t, isValid, true)
}

func TestIsFreeArgValid(t *testing.T) {
	// empty values
	var feature *rfc.Feature
	isValid := isFreeArgValid("", feature)
	assert.Equal(t, isValid, false)

	// not valid
	isValid = isFreeArgValid("key", feature)
	assert.Equal(t, isValid, false)

	// not valid
	feature = &rfc.Feature{
		ConfigData: map[string]string{
			"test": "test",
		},
	}
	isValid = isFreeArgValid("key", feature)
	assert.Equal(t, isValid, false)

	// valid
	feature = &rfc.Feature{
		ConfigData: map[string]string{
			"test": "test",
			"key":  "value",
		},
	}
	isValid = isFreeArgValid("key", feature)
	assert.Equal(t, isValid, true)
}

func TestIsFixedArgValid(t *testing.T) {
	// empty values
	var feature *rfc.Feature
	isValid := isFixedArgValid("", feature)
	assert.Equal(t, isValid, false)

	// not valid
	isValid = isFixedArgValid("key", feature)
	assert.Equal(t, isValid, false)

	// not valid
	feature = &rfc.Feature{
		ConfigData: map[string]string{
			"test": "test",
		},
	}
	isValid = isFixedArgValid("key", feature)
	assert.Equal(t, isValid, false)

	// valid
	feature = &rfc.Feature{
		ConfigData: map[string]string{
			"test": "test",
			"key":  "value",
		},
	}
	isValid = isFixedArgValid("value", feature)
	assert.Equal(t, isValid, true)
}
