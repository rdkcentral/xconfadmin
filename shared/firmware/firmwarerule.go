package firmware

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/rdkcentral/xconfadmin/util"
	"github.com/rdkcentral/xconfwebconfig/db"
	ru "github.com/rdkcentral/xconfwebconfig/rulesengine"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
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

func GetFirmwareRuleTemplateCount(tenantId string) (int, error) {
	entries, err := db.GetSimpleDao().GetAllAsMapRaw(tenantId, db.TABLE_FIRMWARE_RULE_TEMPLATES, 0)
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

func GetFirmwareSortedRuleAllAsListDB(tenantId string) ([]*corefw.FirmwareRule, error) {
	log.Debug("GetFirmwareSortedRuleAllAsListDB starts...")
	rulemap, err := db.GetCachedSimpleDao().GetAllAsMap(tenantId, db.TABLE_FIRMWARE_RULES)
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

func CreateFirmwareRuleTemplates(tenantId string) (e error) {
	if count, _ := GetFirmwareRuleTemplateCount(tenantId); count > 0 {
		return
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Infof("Creating default FirmwareRuleTemplate...")

	ruleFactory := coreef.NewRuleFactory()
	templateList := []corefw.FirmwareRuleTemplate{}

	// Rule actions
	rule := coreef.NewMacRule(coreef.EMPTY_NAME)
	templateList = append(templateList, *NewFirmwareRuleTemplate(
		corefw.MAC_RULE, rule, coreef.EMPTY_LIST, 1))

	rule = ruleFactory.NewIpRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	templateList = append(templateList, *NewFirmwareRuleTemplate(
		corefw.IP_RULE, rule, coreef.EMPTY_LIST, 2))

	rule = ruleFactory.NewIntermediateVersionRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	templateList = append(templateList, *NewFirmwareRuleTemplate(
		corefw.IV_RULE, rule, []string{corefw.GLOBAL_PERCENT, corefw.TIME_FILTER}, 3))

	rule = ruleFactory.NewMinVersionCheckRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_LIST)
	templateList = append(templateList, *NewFirmwareRuleTemplate(
		corefw.MIN_CHECK_RULE, rule, []string{corefw.GLOBAL_PERCENT, corefw.TIME_FILTER}, 4))

	rule = ruleFactory.NewEnvModelRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	templ := *NewFirmwareRuleTemplate(corefw.ENV_MODEL_RULE, rule, []string{}, 5)
	templ.Editable = false
	templateList = append(templateList, templ)

	// Blocking filters
	rule = *ruleFactory.NewGlobalPercentFilterTemplate(coreef.DEFAULT_PERCENT, coreef.EMPTY_NAME)
	templ = *NewBlockingFilterTemplate(corefw.GLOBAL_PERCENT, rule, 1)
	templateList = append(templateList, templ)

	rule = *ruleFactory.NewIpFilter(coreef.EMPTY_NAME)
	templateList = append(templateList, *NewBlockingFilterTemplate(
		corefw.IP_FILTER, rule, 2))

	rule = *ruleFactory.NewTimeFilterTemplate(true, true, false, coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_NAME, "01:00", "02:00")
	templateList = append(templateList, *NewBlockingFilterTemplate(
		corefw.TIME_FILTER, rule, 3))

	// Define Properties
	rule = *ruleFactory.NewDownloadLocationFilter(coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	properties := map[string]corefw.PropertyValue{
		coreef.FIRMWARE_DOWNLOAD_PROTOCOL: *corefw.NewPropertyValue("tftp", false, corefw.STRING),
		coreef.FIRMWARE_LOCATION:          *corefw.NewPropertyValue("", false, corefw.STRING),
		coreef.IPV6_FIRMWARE_LOCATION:     *corefw.NewPropertyValue("", true, corefw.STRING),
	}
	templateList = append(templateList, *NewDefinePropertiesTemplate(
		corefw.DOWNLOAD_LOCATION_FILTER, rule, properties, coreef.EMPTY_LIST, 3))

	rule = *ruleFactory.NewRiFilterTemplate()
	properties = map[string]corefw.PropertyValue{
		coreef.REBOOT_IMMEDIATELY: *corefw.NewPropertyValue("true", false, corefw.BOOLEAN),
	}
	templateList = append(templateList, *NewDefinePropertiesTemplate(
		corefw.REBOOT_IMMEDIATELY_FILTER, rule, properties, coreef.EMPTY_LIST, 1))

	rule = ruleFactory.NewMinVersionCheckRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME, coreef.EMPTY_LIST)
	properties = map[string]corefw.PropertyValue{
		coreef.REBOOT_IMMEDIATELY: *corefw.NewPropertyValue("true", true, corefw.BOOLEAN),
	}
	templateList = append(templateList, *NewDefinePropertiesTemplate(
		corefw.MIN_CHECK_RI, rule, properties, []string{corefw.GLOBAL_PERCENT, corefw.TIME_FILTER}, 2))

	rule = ruleFactory.NewActivationVersionRule(coreef.EMPTY_NAME, coreef.EMPTY_NAME)
	properties = map[string]corefw.PropertyValue{
		coreef.REBOOT_IMMEDIATELY: *corefw.NewPropertyValue("false", false, corefw.BOOLEAN),
	}
	templ = *NewDefinePropertiesTemplate(
		corefw.ACTIVATION_VERSION, rule, properties, coreef.EMPTY_LIST, 4)
	templ.Editable = false
	templateList = append(templateList, templ)

	for _, template := range templateList {
		if err := template.Validate(); err != nil {
			e = errors.Join(e, err)
		}
		template.Updated = util.GetTimestamp()
		if jsonData, err := json.Marshal(template); err != nil {
			e = errors.Join(e, err)
		} else {
			if err := db.GetSimpleDao().SetOne(tenantId, db.TABLE_FIRMWARE_RULE_TEMPLATES, template.ID, jsonData, template.Updated); err != nil {
				e = errors.Join(e, err)
			}
		}
	}
	return e
}
