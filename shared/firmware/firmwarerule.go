package firmware

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rdkcentral/xconfwebconfig/db"

	ru "github.com/rdkcentral/xconfwebconfig/rulesengine"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	log "github.com/sirupsen/logrus"
)

type ApplicableAction struct {
	Type                       string               `json:"type"` // Java class name
	ActionType                 ApplicableActionType `json:"actionType,omitempty" jsonschema:"enum=RULE,enum=DEFINE_PROPERTIES,enum=BLOCKING_FILTER"`
	ConfigId                   string               `json:"configId,omitempty"`
	ConfigEntries              []ConfigEntry        `json:"configEntries"` // RuleAction
	Active                     bool                 `json:"active"`
	UseAccountPercentage       bool                 `json:"useAccountPercentage"`
	FirmwareCheckRequired      bool                 `json:"firmwareCheckRequired"`
	RebootImmediately          bool                 `json:"rebootImmediately"`
	Whitelist                  string               `json:"whitelist,omitempty"`
	IntermediateVersion        string               `json:"intermediateVersion,omitempty"`
	FirmwareVersions           []string             `json:"firmwareVersions,omitempty"`
	Properties                 map[string]string    `json:"properties,omitempty"` // DefinePropertiesAction
	ByPassFilters              []string             `json:"byPassFilters,omitempty"`
	ActivationFirmwareVersions map[string][]string  `json:"activationFirmwareVersions,omitempty"`
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

func GetFirmwareSortedRuleAllAsListDB() ([]*corefw.FirmwareRule, error) {
	log.Debug("GetFirmwareSortedRuleAllAsListDB starts...")
	rulemap, err := db.GetCachedSimpleDao().GetAllAsMap(db.TABLE_FIRMWARE_RULE)
	if err != nil {
		return nil, err
	}

	var rulereflst []*corefw.FirmwareRule

	for _, v := range rulemap {
		rule := v.(*corefw.FirmwareRule)
		rulereflst = append(rulereflst, rule)
	}

	// sort rulereflst based on rule.Name
	sort.Slice(rulereflst, func(i, j int) bool {
		return strings.Compare(strings.ToLower(rulereflst[i].Name), strings.ToLower(rulereflst[j].Name)) < 0
	})

	log.Debug("GetFirmwareSortedRuleAllAsListDB ends...")
	return rulereflst, nil
}
