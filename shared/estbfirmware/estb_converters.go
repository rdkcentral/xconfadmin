package estbfirmware

import (
	"errors"
	"regexp"
	"strings"
	xutil "xconfadmin/util"

	"github.com/rdkcentral/xconfwebconfig/db"
	rf "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	ru "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/util"
)

func GetNormalizedMacAddresses(macAddresses string) ([]string, error) {
	validMacs := util.Set{}
	if !util.IsBlank(macAddresses) {
		re := regexp.MustCompile("[\\s,]+")
		macs := re.Split(macAddresses, -1)
		for _, mac := range macs {
			if util.IsValidMacAddress(mac) {
				validMacs.Add(mac)
			} else {
				return nil, errors.New("Please enter a valid MAC address or whitespace delimited list of MAC addresses.")
			}
		}
	}

	return validMacs.ToSlice(), nil
}

func ConvertToListOfIpAddressGroups(genericLists []*shared.GenericNamespacedList) []*shared.IpAddressGroup {
	result := make([]*shared.IpAddressGroup, len(genericLists))
	for i, genericList := range genericLists {
		result[i] = shared.ConvertToIpAddressGroup(genericList)
	}
	return result
}

func ConvertGlobalPercentageIntoRule(globalpercentage *coreef.GlobalPercentage, applicationType string) *corefw.FirmwareRule {
	percentage := float64(globalpercentage.Percentage)
	var hundredPercentage float64 = 100.0
	whitelistName := globalpercentage.Whitelist

	globalPercentFirmwareRule := coreef.NewGlobalPercentFilter(ru.NewRuleFactory().NewGlobalPercentFilter(hundredPercentage-percentage, whitelistName))
	globalPercentFirmwareRule.ApplicationType = applicationType
	return globalPercentFirmwareRule
}

func MigrateIntoPercentageBean(envModelPercentage *coreef.EnvModelPercentage, firmwareRule *corefw.FirmwareRule) *coreef.PercentageBean {
	percentageBean := coreef.NewPercentageBean()
	if firmwareRule != nil {
		coreef.ParseEnvModelRule(percentageBean, firmwareRule)
		ruleAction := firmwareRule.ApplicableAction
		if ruleAction != nil {
			if len(ruleAction.ConfigEntries) > 0 {
				percentageBean.Distributions = coreef.ConvertIntoPercentRange(ruleAction.ConfigEntries)
			} else if ruleAction.ConfigId != "" {
				configEntry := corefw.NewConfigEntry(ruleAction.ConfigId, 0.0, float64(envModelPercentage.Percentage))
				percentageBean.Distributions = []*corefw.ConfigEntry{configEntry}
			}
		}
	}

	percentageBean.Whitelist = coreef.GetWhitelistName(envModelPercentage.Whitelist)
	percentageBean.Active = envModelPercentage.Active
	percentageBean.LastKnownGood = envModelPercentage.LastKnownGood
	percentageBean.IntermediateVersion = envModelPercentage.IntermediateVersion
	percentageBean.RebootImmediately = envModelPercentage.RebootImmediately
	percentageBean.FirmwareCheckRequired = envModelPercentage.FirmwareCheckRequired
	percentageBean.FirmwareVersions = envModelPercentage.FirmwareVersions
	percentageBean.ApplicationType = firmwareRule.ApplicationType

	return percentageBean
}

func ConvertMacRuleBeanToFirmwareRule(bean *coreef.MacRuleBean) *corefw.FirmwareRule {
	firmwareRule := *corefw.NewEmptyFirmwareRule()
	firmwareRule.ID = bean.Id
	firmwareRule.Name = bean.Name
	firmwareRule.Type = coreef.MAC_RULE
	if bean.FirmwareConfig != nil {
		ruleAction := corefw.ApplicableAction{
			ConfigId:   bean.FirmwareConfig.ID,
			ActionType: corefw.RULE,
			Type:       corefw.RuleActionClass,
		}
		firmwareRule.ApplicableAction = &ruleAction
		firmwareRule.ApplicationType = bean.FirmwareConfig.ApplicationType
	} else {
		ruleAction := corefw.ApplicableAction{
			ActionType: corefw.RULE,
			Type:       corefw.RuleActionClass,
		}
		firmwareRule.ApplicableAction = &ruleAction
	}
	macRule := coreef.NewMacRule(bean.MacListRef)
	firmwareRule.Rule = macRule
	return &firmwareRule
}

func ConvertDownloadLocationFilterToFirmwareRule(filter *coreef.DownloadLocationFilter) (*corefw.FirmwareRule, error) {
	var ipList string
	if filter.IpAddressGroup != nil {
		ipList = filter.IpAddressGroup.Name
	}
	ipv4Location := filter.FirmwareLocation
	ipv6Location := filter.Ipv6FirmwareLocation
	isTftpEmpty := (ipv4Location == nil)
	isHttpEmpty := util.IsBlank(filter.HttpLocation)

	var rule *rf.Rule
	var action *corefw.ApplicableAction

	if !filter.ForceHttp && !isTftpEmpty && !isHttpEmpty {
		return nil, errors.New("Can't convert DownloadLocationFilter into FirmwareRule because filter contains both locations for http and tftp.")
	} else if filter.ForceHttp || isTftpEmpty {
		downloadProtocol := ""
		if HasProtocolSuffix(filter.Name) {
			downloadProtocol = shared.Http
		}
		rule = ru.NewRuleFactory().NewDownloadLocationFilter(ipList, downloadProtocol)
		action = newHttpAction(filter.HttpLocation)
	} else {
		downloadProtocol := ""
		if HasProtocolSuffix(filter.Name) {
			downloadProtocol = shared.Tftp
		}
		rule = ru.NewRuleFactory().NewDownloadLocationFilter(ipList, downloadProtocol)
		action = newTftpAction(ipv4Location, ipv6Location)
	}

	firmwareRule := corefw.NewEmptyFirmwareRule()
	firmwareRule.ID = filter.Id
	firmwareRule.Name = filter.Name
	firmwareRule.Type = coreef.DOWNLOAD_LOCATION_FILTER
	firmwareRule.Rule = *rule
	firmwareRule.ApplicableAction = action

	return firmwareRule, nil
}

func newHttpAction(location string) *corefw.ApplicableAction {
	properties := make(map[string]string)
	properties[coreef.FIRMWARE_LOCATION] = location
	properties[coreef.IPV6_FIRMWARE_LOCATION] = ""
	properties[coreef.FIRMWARE_DOWNLOAD_PROTOCOL] = shared.Http

	return newDefinePropertiesAction(properties)
}

func newTftpAction(firmwareLocation *shared.IpAddress, ipv6FirmwareLocation *shared.IpAddress) *corefw.ApplicableAction {
	location := firmwareLocation.Address
	var ipv6Location string
	if ipv6FirmwareLocation != nil {
		ipv6Location = ipv6FirmwareLocation.Address
	}

	properties := make(map[string]string)
	properties[coreef.FIRMWARE_LOCATION] = location
	properties[coreef.IPV6_FIRMWARE_LOCATION] = ipv6Location
	properties[coreef.FIRMWARE_DOWNLOAD_PROTOCOL] = shared.Http

	return newDefinePropertiesAction(properties)
}

func HasProtocolSuffix(name string) bool {
	return strings.HasSuffix(name, coreef.HTTP_SUFFIX) || strings.HasSuffix(name, coreef.TFTP_SUFFIX)
}

func ConvertModelRuleBeanToFirmwareRule(bean *coreef.EnvModelBean) *corefw.FirmwareRule {
	firmwareRule := *corefw.NewEmptyFirmwareRule()
	envModelRule := coreef.NewRuleFactory().NewEnvModelRule(bean.EnvironmentId, bean.ModelId)
	firmwareRule.Rule = envModelRule
	firmwareRule.Type = coreef.ENV_MODEL_RULE
	firmwareRule.Name = bean.Name
	firmwareRule.ID = bean.Id
	if bean.FirmwareConfig != nil {
		ruleAction := corefw.ApplicableAction{
			ConfigId:   bean.FirmwareConfig.ID,
			ActionType: corefw.RULE,
			Type:       corefw.RuleActionClass,
		}
		firmwareRule.ApplicableAction = &ruleAction
	} else {
		ruleAction := corefw.ApplicableAction{
			ActionType: corefw.RULE,
			Type:       corefw.RuleActionClass,
		}
		firmwareRule.ApplicableAction = &ruleAction
	}
	return &firmwareRule
}

func RebootImmediatelyFiltersByName(applicationType string, name string) (*coreef.RebootImmediatelyFilter, error) {
	rulelst, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FIRMWARE_RULE, 0)
	if err != nil {
		return nil, err
	}

	for _, rule := range rulelst {
		frule := rule.(*corefw.FirmwareRule)
		if frule.ApplicationType != applicationType {
			continue
		}
		if frule.GetTemplateId() != coreef.REBOOT_IMMEDIATELY_FILTER {
			continue
		}
		if frule.Name == name {
			filter := ConvertFirmwareRuleToRebootFilter(frule)
			return filter, nil
		}
	}

	return nil, nil
}

func ConvertFirmwareRuleToRebootFilter(firmwareRule *corefw.FirmwareRule) *coreef.RebootImmediatelyFilter {
	filter := &coreef.RebootImmediatelyFilter{
		Id:   firmwareRule.ID,
		Name: firmwareRule.Name,
	}

	convertConditionsForRebootFilter(firmwareRule, filter)

	if filter.Environments == nil {
		filter.Environments = make([]string, 0)
	}

	if filter.Models == nil {
		filter.Models = make([]string, 0)
	}

	return filter
}

func convertConditionsForRebootFilter(firmwareRule *corefw.FirmwareRule, rebootFilter *coreef.RebootImmediatelyFilter) {
	conditions := rf.ToConditions(&firmwareRule.Rule)
	for _, condition := range conditions {
		isLegacyIpFreeArg := coreef.IsLegacyIpFreeArg(condition.FreeArg)
		if isLegacyIpFreeArg || ru.RuleFactoryIP.Equals(condition.FreeArg) {
			if rebootFilter.IpAddressGroup == nil {
				rebootFilter.IpAddressGroup = make([]*shared.IpAddressGroup, 0)
			}
			ipAddressGroup := coreef.GetIpAddressGroup(condition)
			rebootFilter.IpAddressGroup = append(rebootFilter.IpAddressGroup, ipAddressGroup)
		} else if isLegacyIpFreeArg || ru.RuleFactoryMAC.Equals(condition.FreeArg) {
			if condition.FixedArg.IsCollectionValue() {
				macAddresses := fixedArgValueToCollection(condition)
				rebootFilter.MacAddress = strings.Join(macAddresses, "\n")
			} else {
				rebootFilter.MacAddress = condition.FixedArg.GetValue().(string)
			}
		} else if ru.RuleFactoryENV.Equals(condition.FreeArg) {
			rebootFilter.Environments = fixedArgValueToCollection(condition)
		} else if ru.RuleFactoryMODEL.Equals(condition.FreeArg) {
			rebootFilter.Models = fixedArgValueToCollection(condition)
		}
	}
}

func fixedArgValueToCollection(condition *rf.Condition) []string {
	if condition.FixedArg != nil && condition.FixedArg.IsCollectionValue() {
		return xutil.StringCopySlice(condition.FixedArg.Collection.Value)
	}
	return []string{}
}

func ConvertRebootFilterToFirmwareRule(filter *coreef.RebootImmediatelyFilter) (*corefw.FirmwareRule, error) {
	macAddresses, err := GetNormalizedMacAddresses(filter.MacAddress)
	if err != nil {
		return nil, err
	}

	ipAddressGroups := make([]string, 0)
	if filter.IpAddressGroup != nil {
		for _, ipAddressGroup := range filter.IpAddressGroup {
			if ipAddressGroup != nil {
				ipAddressGroups = append(ipAddressGroups, ipAddressGroup.Name)
			}
		}
	}

	rule := ru.NewRuleFactory().NewRiFilter(ipAddressGroups, macAddresses, filter.Environments, filter.Models)

	properties := make(map[string]string)
	properties[coreef.REBOOT_IMMEDIATELY] = "true"
	action := newDefinePropertiesAction(properties)

	firmwareRule := corefw.NewEmptyFirmwareRule()
	firmwareRule.ID = filter.Id
	firmwareRule.Name = filter.Name
	firmwareRule.Type = coreef.REBOOT_IMMEDIATELY_FILTER
	firmwareRule.Rule = *rule
	firmwareRule.ApplicableAction = action

	return firmwareRule, nil
}

func newDefinePropertiesAction(properties map[string]string) *corefw.ApplicableAction {
	return &corefw.ApplicableAction{
		Type:       corefw.DefinePropertiesActionClass,
		ActionType: corefw.DEFINE_PROPERTIES,
		Properties: properties,
	}
}

func ConvertTimeFilterToFirmwareRule(timeFilter *coreef.TimeFilter) *corefw.FirmwareRule {
	var ipList string
	if timeFilter.IpWhiteList != nil {
		ipList = timeFilter.IpWhiteList.Name
	}

	rule := ru.NewRuleFactory().NewTimeFilter(
		timeFilter.NeverBlockRebootDecoupled,
		timeFilter.NeverBlockHttpDownload,
		timeFilter.LocalTime,
		timeFilter.EnvModelRuleBean.EnvironmentId,
		timeFilter.EnvModelRuleBean.ModelId,
		ipList,
		timeFilter.Start,
		timeFilter.End)

	action := &corefw.ApplicableAction{
		Type:       corefw.BlockingFilterActionClass,
		ActionType: corefw.BLOCKING_FILTER,
	}

	firmwareRule := corefw.NewEmptyFirmwareRule()
	firmwareRule.ID = timeFilter.Id
	firmwareRule.Name = timeFilter.Name
	firmwareRule.Rule = *rule
	firmwareRule.Type = corefw.TIME_FILTER
	firmwareRule.ApplicableAction = action

	return firmwareRule
}
