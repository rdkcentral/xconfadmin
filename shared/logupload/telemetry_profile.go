package logupload

import (
	"encoding/json"
	"fmt"
	"xconfwebconfig/db"
	"xconfwebconfig/shared"
	"xconfwebconfig/shared/logupload"

	log "github.com/sirupsen/logrus"
)

const PermanentTelemetryProfileConst = "PermanentTelemetryProfile"

func SetOnePermanentTelemetryProfile(rowKey string, profile *logupload.PermanentTelemetryProfile) error {
	return logupload.GetCachedSimpleDaoFunc().SetOne(db.TABLE_PERMANENT_TELEMETRY, rowKey, profile)
}

func DeletePermanentTelemetryProfile(rowKey string) {
	logupload.GetCachedSimpleDaoFunc().DeleteOne(db.TABLE_PERMANENT_TELEMETRY, rowKey)
}

func GetPermanentTelemetryProfileListByApplicationType(applicationType string) []*logupload.PermanentTelemetryProfile {
	result := []*logupload.PermanentTelemetryProfile{}
	list := GetPermanentTelemetryProfileList()
	for _, profile := range list {
		if profile.ApplicationType == applicationType {
			result = append(result, profile)
		}
	}
	return result
}

func GetPermanentTelemetryProfileList() []*logupload.PermanentTelemetryProfile {
	all := []*logupload.PermanentTelemetryProfile{}
	list, err := logupload.GetCachedSimpleDaoFunc().GetAllAsList(db.TABLE_PERMANENT_TELEMETRY, 0)
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

func GetTelemetryTwoProfileListByApplicationType(applicationType string) []*logupload.TelemetryTwoProfile {
	result := []*logupload.TelemetryTwoProfile{}
	list := GetAllTelemetryTwoProfileList(applicationType)
	for _, profile := range list {
		if profile.ApplicationType == applicationType {
			result = append(result, profile)
		}
	}
	return result
}

func GetAllTelemetryTwoProfileList(appType string) []*logupload.TelemetryTwoProfile {
	result := []*logupload.TelemetryTwoProfile{}
	list, err := logupload.GetCachedSimpleDaoFunc().GetAllAsList(db.TABLE_TELEMETRY_TWO_PROFILES, 0)
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

func GetOneTelemetryTwoProfile(rowKey string) *logupload.TelemetryTwoProfile {
	telemetryInst, err := logupload.GetCachedSimpleDaoFunc().GetOne(db.TABLE_TELEMETRY_TWO_PROFILES, rowKey)
	if err != nil {
		log.Warn(fmt.Sprintf("no TelemetryTwoProfile found for " + rowKey))
		return nil
	}
	telemetry := telemetryInst.(*logupload.TelemetryTwoProfile)
	return telemetry
}

func SetOneTelemetryTwoProfile(profile *logupload.TelemetryTwoProfile) error {
	return logupload.GetCachedSimpleDaoFunc().SetOne(db.TABLE_TELEMETRY_TWO_PROFILES, profile.ID, profile)
}

func DeleteTelemetryTwoProfile(rowKey string) error {
	return logupload.GetCachedSimpleDaoFunc().DeleteOne(db.TABLE_TELEMETRY_TWO_PROFILES, rowKey)
}

func SetOneTelemetryProfile(rowKey string, telemetry *logupload.TelemetryProfile) {
	logupload.GetCachedSimpleDaoFunc().SetOne(db.TABLE_TELEMETRY, rowKey, telemetry)
}

func GetTimestampedRulesPointer() []*logupload.TimestampedRule {
	timestampedRuleSet, err := logupload.GetCachedSimpleDaoFunc().GetKeys(db.TABLE_TELEMETRY)
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

func GetOneTelemetryRule(id string) *logupload.TelemetryRule {
	tRuleOne, err := logupload.GetCachedSimpleDaoFunc().GetOne(db.TABLE_TELEMETRY_RULES, id)
	if err != nil {
		log.Warn("no TelemetryRule found")
		return nil
	}
	tRule := tRuleOne.(*logupload.TelemetryRule)
	return tRule
}

func GetOneTelemetryTwoRule(rowKey string) *logupload.TelemetryTwoRule {
	telemetryInst, err := logupload.GetCachedSimpleDaoFunc().GetOne(db.TABLE_TELEMETRY_TWO_RULES, rowKey)
	if err != nil {
		log.Warn(fmt.Sprintf("no telemetryProfile found for " + rowKey))
		return nil
	}
	telemetry := telemetryInst.(*logupload.TelemetryTwoRule)
	return telemetry
}

func SetOneTelemetryTwoRule(rowKey string, telemetry *logupload.TelemetryTwoRule) error {
	return logupload.GetCachedSimpleDaoFunc().SetOne(db.TABLE_TELEMETRY_TWO_RULES, rowKey, telemetry)
}

func DeleteTelemetryTwoRule(rowKey string) error {
	return logupload.GetCachedSimpleDaoFunc().DeleteOne(db.TABLE_TELEMETRY_TWO_RULES, rowKey)
}

func GetOnePermanentTelemetryProfile(rowKey string) *logupload.PermanentTelemetryProfile {
	telemetryInst, err := logupload.GetCachedSimpleDaoFunc().GetOne(db.TABLE_PERMANENT_TELEMETRY, rowKey)
	if err != nil {
		log.Warn(fmt.Sprintf("no telemetryProfile found for " + rowKey))
		return nil
	}
	telemetry := telemetryInst.(*logupload.PermanentTelemetryProfile)
	return telemetry
}
