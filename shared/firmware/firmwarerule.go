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
package firmware

import (
	"fmt"
	"time"
	"xconfwebconfig/db"
	"xconfwebconfig/util"

	ru "xconfwebconfig/rulesengine"
	corefw "xconfwebconfig/shared/firmware"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func CreateFirmwareRuleOneDBAfterValidate(fr *corefw.FirmwareRule) error {
	if util.IsBlank(fr.ID) {
		fr.ID = uuid.New().String()
	}
	fr.Updated = util.GetTimestamp(time.Now().UTC())
	return db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, fr.ID, fr)
}

func GetFirmwareRuleTemplateCount() (int, error) {
	entries, err := db.GetSimpleDao().GetAllAsMapRaw(db.TABLE_FIRMWARE_RULE_TEMPLATE, 0)
	if err != nil {
		log.Error(fmt.Sprintf("GetFirmwareRuleTemplateCount: %v", err))
		return 0, err
	}
	return len(entries), nil
}

func NewFirmwareRuleTemplate(id string, rule ru.Rule, byPassFilters []string, priority int) *corefw.FirmwareRuleTemplate {
	action := corefw.NewTemplateApplicableActionAndType(corefw.RuleActionClass, corefw.RULE_TEMPLATE, "")
	return &corefw.FirmwareRuleTemplate{
		ID:               id,
		Priority:         int32(priority),
		Rule:             rule,
		ApplicableAction: action,
		Editable:         true,
		RequiredFields:   []string{},
		ByPassFilters:    byPassFilters,
	}
}

func NewBlockingFilterTemplate(id string, rule ru.Rule, priority int) *corefw.FirmwareRuleTemplate {
	action := corefw.NewTemplateApplicableActionAndType(corefw.BlockingFilterActionClass, corefw.BLOCKING_FILTER_TEMPLATE, "")
	return &corefw.FirmwareRuleTemplate{
		ID:               id,
		Priority:         int32(priority),
		Rule:             rule,
		ApplicableAction: action,
		Editable:         true,
		RequiredFields:   []string{},
		ByPassFilters:    []string{},
	}
}

func NewDefinePropertiesTemplate(id string, rule ru.Rule, properties map[string]corefw.PropertyValue, byPassFilter []string, priority int) *corefw.FirmwareRuleTemplate {
	action := corefw.NewTemplateApplicableActionAndType(corefw.DefinePropertiesTemplateActionClass, corefw.DEFINE_PROPERTIES_TEMPLATE, "")
	action.Properties = properties
	return &corefw.FirmwareRuleTemplate{
		ID:               id,
		Priority:         int32(priority),
		Rule:             rule,
		ApplicableAction: action,
		Editable:         true,
		RequiredFields:   []string{},
		ByPassFilters:    byPassFilter,
	}
}
