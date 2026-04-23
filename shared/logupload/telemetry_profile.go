package logupload

import (
	"encoding/json"
	"fmt"

	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	log "github.com/sirupsen/logrus"
)

const PermanentTelemetryProfileConst = "PermanentTelemetryProfile"

func SetOnePermanentTelemetryProfile(tenantId string, rowKey string, profile *logupload.PermanentTelemetryProfile) error {
	return logupload.GetCachedSimpleDaoFunc().SetOne(tenantId, db.TABLE_PERMANENT_TELEMETRY_PROFILES, rowKey, profile)
}

func DeletePermanentTelemetryProfile(tenantId string, rowKey string) {
	logupload.GetCachedSimpleDaoFunc().DeleteOne(tenantId, db.TABLE_PERMANENT_TELEMETRY_PROFILES, rowKey)
}

func GetPermanentTelemetryProfileListByApplicationType(tenantId string, applicationType string) []*logupload.PermanentTelemetryProfile {
	result := []*logupload.PermanentTelemetryProfile{}
	list := GetPermanentTelemetryProfileList(tenantId)
	for _, profile := range list {
		if profile.ApplicationType == applicationType {
			result = append(result, profile)
		}
	}
	return result
}

func GetPermanentTelemetryProfileList(tenantId string) []*logupload.PermanentTelemetryProfile {
	all := []*logupload.PermanentTelemetryProfile{}
	list, err := logupload.GetCachedSimpleDaoFunc().GetAllAsList(tenantId, db.TABLE_PERMANENT_TELEMETRY_PROFILES, 0)
	if err != nil {
		log.Warn("no TelemetryProfile found")
		return nil
	}
	for idx := range list {
		tProfile := list[idx].(*logupload.PermanentTelemetryProfile)
		all = append(all, tProfile)
	}
	return all
}

func NewEmptyPermanentTelemetryProfile() *logupload.PermanentTelemetryProfile {
	return &logupload.PermanentTelemetryProfile{
		Type:            PermanentTelemetryProfileConst,
		ApplicationType: shared.STB,
	}
}

func GetTelemetryTwoProfileListByApplicationType(tenantId string, applicationType string) []*logupload.TelemetryTwoProfile {
	result := []*logupload.TelemetryTwoProfile{}
	list := GetAllTelemetryTwoProfileList(tenantId, applicationType)
	for _, profile := range list {
		if profile.ApplicationType == applicationType {
			result = append(result, profile)
		}
	}
	return result
}

func GetAllTelemetryTwoProfileList(tenantId string, appType string) []*logupload.TelemetryTwoProfile {
	result := []*logupload.TelemetryTwoProfile{}
	list, err := logupload.GetCachedSimpleDaoFunc().GetAllAsList(tenantId, db.TABLE_TELEMETRY_TWO_PROFILES, 0)
	if err != nil {
		log.Warn("no TelemetryTwoProfile found")
		return nil
	}
	for _, inst := range list {
		twoProfile := inst.(*logupload.TelemetryTwoProfile)
		if twoProfile.ApplicationType != appType {
			continue
		}
		result = append(result, twoProfile)
	}
	return result
}

func NewEmptyTelemetryTwoProfile() *logupload.TelemetryTwoProfile {
	return &logupload.TelemetryTwoProfile{
		Type:            "TelemetryTwoProfile",
		ApplicationType: shared.STB,
	}
}

func GetOneTelemetryTwoProfile(tenantId string, rowKey string) *logupload.TelemetryTwoProfile {
	telemetryInst, err := logupload.GetCachedSimpleDaoFunc().GetOne(tenantId, db.TABLE_TELEMETRY_TWO_PROFILES, rowKey)
	if err != nil {
		log.Warn("no TelemetryTwoProfile found for " + rowKey)
		return nil
	}
	telemetry := telemetryInst.(*logupload.TelemetryTwoProfile)
	return telemetry
}

func SetOneTelemetryTwoProfile(tenantId string, profile *logupload.TelemetryTwoProfile) error {
	return logupload.GetCachedSimpleDaoFunc().SetOne(tenantId, db.TABLE_TELEMETRY_TWO_PROFILES, profile.ID, profile)
}

func DeleteTelemetryTwoProfile(tenantId string, rowKey string) error {
	return logupload.GetCachedSimpleDaoFunc().DeleteOne(tenantId, db.TABLE_TELEMETRY_TWO_PROFILES, rowKey)
}

func SetOneTelemetryProfile(tenantId string, rowKey string, telemetry *logupload.TelemetryProfile) {
	logupload.GetCachedSimpleDaoFunc().SetOne(tenantId, db.TABLE_TELEMETRY_PROFILES, rowKey, telemetry)
}

func GetTimestampedRulesPointer(tenantId string) []*logupload.TimestampedRule {
	timestampedRuleSet, err := logupload.GetCachedSimpleDaoFunc().GetKeys(tenantId, db.TABLE_TELEMETRY_PROFILES)
	if err != nil {
		log.Warn(fmt.Sprintf("no TimestampedRule found"))
		return nil
	}
	rules := []*logupload.TimestampedRule{}
	for idx := range timestampedRuleSet {
		var timestampedRule logupload.TimestampedRule
		timestampedRuleString := timestampedRuleSet[idx].(string)
		json.Unmarshal([]byte(timestampedRuleString), &timestampedRule)
		rules = append(rules, &timestampedRule)
	}
	return rules
}

func GetOneTelemetryRule(tenantId string, id string) *logupload.TelemetryRule {
	tRuleOne, err := logupload.GetCachedSimpleDaoFunc().GetOne(tenantId, db.TABLE_TELEMETRY_RULES, id)
	if err != nil {
		log.Warn("no TelemetryRule found")
		return nil
	}
	tRule := tRuleOne.(*logupload.TelemetryRule)
	return tRule
}

func GetOneTelemetryTwoRule(tenantId string, rowKey string) *logupload.TelemetryTwoRule {
	telemetryInst, err := logupload.GetCachedSimpleDaoFunc().GetOne(tenantId, db.TABLE_TELEMETRY_TWO_RULES, rowKey)
	if err != nil {
		log.Warn("no telemetryProfile found for " + rowKey)
		return nil
	}
	telemetry := telemetryInst.(*logupload.TelemetryTwoRule)
	return telemetry
}

func SetOneTelemetryTwoRule(tenantId string, rowKey string, telemetry *logupload.TelemetryTwoRule) error {
	return logupload.GetCachedSimpleDaoFunc().SetOne(tenantId, db.TABLE_TELEMETRY_TWO_RULES, rowKey, telemetry)
}

func DeleteTelemetryTwoRule(tenantId string, rowKey string) error {
	return logupload.GetCachedSimpleDaoFunc().DeleteOne(tenantId, db.TABLE_TELEMETRY_TWO_RULES, rowKey)
}

func GetOnePermanentTelemetryProfile(tenantId string, rowKey string) *logupload.PermanentTelemetryProfile {
	telemetryInst, err := logupload.GetCachedSimpleDaoFunc().GetOne(tenantId, db.TABLE_PERMANENT_TELEMETRY_PROFILES, rowKey)
	if err != nil {
		log.Warn("no telemetryProfile found for " + rowKey)
		return nil
	}
	telemetry := telemetryInst.(*logupload.PermanentTelemetryProfile)
	return telemetry
}
