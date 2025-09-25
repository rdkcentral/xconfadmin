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
package telemetry

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/rdkcentral/xconfadmin/shared"
	xlogupload "github.com/rdkcentral/xconfadmin/shared/logupload"

	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

func CreateTelemetryProfile(contextAttribute string, expectedValue string, telemetry *xwlogupload.TelemetryProfile) *xwlogupload.TimestampedRule {
	telemetryRule := CreateRuleForAttribute(contextAttribute, expectedValue)
	telemetryRuleBytes, _ := json.Marshal(telemetryRule)
	xlogupload.SetOneTelemetryProfile(string(telemetryRuleBytes), telemetry)
	return telemetryRule
}

func CreateRuleForAttribute(contextAttribute string, expectedValue string) *xwlogupload.TimestampedRule {
	freeArg := rulesengine.FreeArg{
		Type: "STRING",
		Name: contextAttribute,
	}
	fixedArg := rulesengine.NewFixedArg(expectedValue)
	condition := rulesengine.NewCondition(&freeArg, rulesengine.StandardOperationIs, fixedArg)
	rule := rulesengine.NewEmptyRule()
	rule.Condition = condition
	timestampedRule := xwlogupload.NewTimestampedRule()
	timestampedRule.Rule = *rule
	now := time.Now()
	nanos := now.UnixNano()
	millis := nanos / 1000000
	timestampedRule.Timestamp = millis
	return timestampedRule
}

func DropTelemetryFor(contextAttribute string, expectedValue string) []*xwlogupload.TelemetryProfile {
	context := map[string]string{
		contextAttribute: expectedValue,
	}
	matchedRules := getMatchedRules(context)
	telemetryProfileList := []*xwlogupload.TelemetryProfile{}
	for _, timestampedRule := range matchedRules {
		timestampedRuleBytes, _ := json.Marshal(timestampedRule)
		telemetryProfile := xwlogupload.GetOneTelemetryProfile(string(timestampedRuleBytes))
		if telemetryProfile != nil {
			telemetryProfileList = append(telemetryProfileList, telemetryProfile)
			log.Debug(fmt.Sprintf("removing temporary rule: : %v", telemetryProfile))
			xwlogupload.DeleteTelemetryProfile(string(timestampedRuleBytes))
		}
	}
	return telemetryProfileList
}

func getMatchedRules(context map[string]string) []*xwlogupload.TimestampedRule {
	timestampedRuleList := xlogupload.GetTimestampedRulesPointer()
	matched := []*xwlogupload.TimestampedRule{}
	ruleProcessor := rulesengine.NewRuleProcessor()
	for _, timestampedRule := range timestampedRuleList {
		if ruleProcessor.Evaluate(&timestampedRule.Rule, context, log.Fields{}) {
			matched = append(matched, timestampedRule)
		}
	}
	return matched
}

func GetAvailableDescriptors(applicationType string) []*xwlogupload.PermanentTelemetryRuleDescriptor {
	descriptors := []*xwlogupload.PermanentTelemetryRuleDescriptor{}
	telemetryRuleList := xwlogupload.GetTelemetryRuleListForAs() //[]*TelemetryRule
	for _, telemetryRule := range telemetryRuleList {
		if telemetryRule != nil && shared.ApplicationTypeEquals(telemetryRule.ApplicationType, applicationType) {
			ruleDescriptor := xwlogupload.NewPermanentTelemetryRuleDescriptor()
			ruleDescriptor.RuleId = telemetryRule.ID
			ruleDescriptor.RuleName = telemetryRule.Name
			descriptors = append(descriptors, ruleDescriptor)
		}
	}
	return descriptors
}

func GetAvailableProfileDescriptors(applicationType string) []*xwlogupload.TelemetryProfileDescriptor {
	descriptors := []*xwlogupload.TelemetryProfileDescriptor{}
	permanentTelemetryProfileList := xwlogupload.GetPermanentTelemetryProfileList() //[]*PermanentTelemetryProfile
	for _, telemetry := range permanentTelemetryProfileList {
		if telemetry != nil && shared.ApplicationTypeEquals(telemetry.ApplicationType, applicationType) {
			profileDescriptor := xwlogupload.NewTelemetryProfileDescriptor()
			profileDescriptor.ID = telemetry.ID
			profileDescriptor.Name = telemetry.Name
			descriptors = append(descriptors, profileDescriptor)
		}
	}
	return descriptors
}
