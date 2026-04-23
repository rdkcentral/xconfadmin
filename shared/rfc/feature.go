package rfc

import (
	"fmt"

	xshared "github.com/rdkcentral/xconfadmin/shared"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwshared "github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/rfc"
	xwrfc "github.com/rdkcentral/xconfwebconfig/shared/rfc"

	log "github.com/sirupsen/logrus"
)

func DoesFeatureExistWithApplicationType(tenantId string, id string, applicationType string) bool {
	if id == "" {
		return false
	}
	feature := xwrfc.GetOneFeature(tenantId, id)
	return feature != nil && applicationType == feature.ApplicationType
}

func DoesFeatureExist(tenantId string, id string) bool {
	if id == "" {
		return false
	}
	feature := xwrfc.GetOneFeature(tenantId, id)
	return feature != nil
}

func IsValidFeature(tenantId string, feature *xwrfc.Feature) (bool, string) {
	errorMsg := ""
	if feature == nil || feature.ApplicationType == "" {
		errorMsg = "Application type is empty"
		return false, errorMsg
	}
	if !xshared.IsValidApplicationType(feature.ApplicationType) {
		errorMsg = fmt.Sprintf("ApplicationType %s is not valid", feature.ApplicationType)
		return false, errorMsg
	}
	if feature.Name == "" {
		errorMsg = "Name is blank"
		return false, errorMsg
	}
	if feature.FeatureName == "" {
		errorMsg = "Feature Name is blank"
		return false, errorMsg
	}
	if feature.ConfigData != nil && len(feature.ConfigData) > 0 {
		for key, value := range feature.ConfigData {
			if key == "" {
				errorMsg = "Key is blank"
				return false, errorMsg
			}
			if value == "" {
				errorMsg = fmt.Sprintf("Value is blank for key: %s", key)
				return false, errorMsg
			}
		}
	}
	if feature.Whitelisted {
		if feature.WhitelistProperty == nil || feature.WhitelistProperty.Key == "" {
			errorMsg = "Key is required"
			return false, errorMsg
		}
		if feature.WhitelistProperty.Value == "" {
			errorMsg = "Value is required"
			return false, errorMsg
		}
		result, _ := xwshared.GetGenericNamedListOneDB(tenantId, feature.WhitelistProperty.Value)
		if result == nil || result.TypeName != feature.WhitelistProperty.NamespacedListType {
			errorMsg = fmt.Sprintf("%s with id %s does not exist", feature.WhitelistProperty.NamespacedListType, feature.WhitelistProperty.Value)
			return false, errorMsg
		}
		if feature.WhitelistProperty.NamespacedListType == "" {
			errorMsg = "NamespacedList type is required"
			return false, errorMsg
		}
		if feature.WhitelistProperty.TypeName == "" {
			errorMsg = "NamespacedList type name is required"
			return false, errorMsg
		}
	}
	return true, errorMsg
}

func DoesFeatureNameExistForAnotherIdForApplicationType(tenantId string, feature *xwrfc.Feature, applicationType string) bool {
	contextMap := map[string]string{xwcommon.APPLICATION_TYPE: applicationType, xwcommon.TENANT_ID: tenantId}
	featureList := GetFilteredFeatureList(contextMap)
	return DoesFeatureNameExistForAnotherIdInList(feature, featureList)
}

func DoesFeatureNameExistForAnotherIdInList(feature *xwrfc.Feature, featureList []*xwrfc.Feature) bool {
	for _, f := range featureList {
		if f.ID != feature.ID && f.ApplicationType == feature.ApplicationType && f.FeatureName == feature.FeatureName {
			return true
		}
	}
	return false
}

func GetFilteredFeatureList(searchContext map[string]string) []*xwrfc.Feature {
	var featureList []*xwrfc.Feature
	tenantId := searchContext[xwcommon.TENANT_ID]
	features, err := db.GetCachedSimpleDao().GetAllAsList(tenantId, db.TABLE_FEATURES, 0)
	if err != nil {
		log.Warn(fmt.Sprintf("no feature found"))
		return nil
	}
	predicates := getFeaturePredicates(searchContext)
	for idx := range features {
		feature := features[idx].(*xwrfc.Feature)
		if isFeatureValid(feature, predicates, searchContext) {
			featureList = append(featureList, feature)
		}
	}
	return featureList
}

func DeleteOneFeature(tenantId string, featureId string) {
	err := db.GetCachedSimpleDao().DeleteOne(tenantId, db.TABLE_FEATURES, featureId)
	if err != nil {
		log.Warn(fmt.Sprintf("no feature found for featureId: %s", featureId))
	}
}

func SetOneFeature(tenantId string, feature *xwrfc.Feature) (*xwrfc.Feature, error) {
	err := db.GetCachedSimpleDao().SetOne(tenantId, db.TABLE_FEATURES, feature.ID, feature)
	if err != nil {
		log.Warn(fmt.Sprintf("error creating feature with featureId: %s", feature.ID))
	}
	return feature, err
}

func GetFilteredFeatureEntityList(searchContext map[string]string) []*xwrfc.FeatureEntity {
	var featureEntityList []*xwrfc.FeatureEntity
	tenantId := searchContext[xwcommon.TENANT_ID]
	features, err := db.GetCachedSimpleDao().GetAllAsList(tenantId, db.TABLE_FEATURES, 0)
	if err != nil {
		log.Warn(fmt.Sprintf("no feature found"))
		return nil
	}
	predicates := getFeaturePredicates(searchContext)
	for idx := range features {
		feature := features[idx].(*xwrfc.Feature)
		if isFeatureValid(feature, predicates, searchContext) {
			featureEntityList = append(featureEntityList, feature.CreateFeatureEntity())
		}
	}
	return featureEntityList
}

func DoesFeatureExistInSomeApplicationType(tenantId string, id string) (bool, string) {
	if id == "" {
		return false, ""
	}
	feature := xwrfc.GetOneFeature(tenantId, id)
	if feature == nil {
		return false, ""
	}
	return true, feature.ApplicationType
}

func GetFeatureEntityList(tenantId string) []*rfc.FeatureEntity {
	var featureEntityList []*rfc.FeatureEntity
	features, err := db.GetCachedSimpleDao().GetAllAsList(tenantId, db.TABLE_FEATURES, 0)
	if err != nil {
		log.Warn(fmt.Sprintf("no feature found"))
		return nil
	}
	for idx := range features {
		featureEntity := (features[idx].(*rfc.Feature)).CreateFeatureEntity()
		featureEntityList = append(featureEntityList, featureEntity)
	}
	return featureEntityList
}

func GetFeatureRule(tenantId string, id string) *rfc.FeatureRule {
	featureRule, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_FEATURE_CONTROL_RULES, id)
	if err != nil {
		log.Warn("no featureRule found")
		return nil
	}
	return featureRule.(*rfc.FeatureRule)
}

func SetFeatureRule(tenantId string, id string, featureRule *rfc.FeatureRule) error {
	if err := db.GetCachedSimpleDao().SetOne(tenantId, db.TABLE_FEATURE_CONTROL_RULES, id, featureRule); err != nil {
		log.Error("cannot save featureRule to DB")
		return err
	}
	return nil
}

func IsValidFeatureEntity(tenantId string, featureEntity *rfc.FeatureEntity) (bool, string) {
	feature := featureEntity.CreateFeature()
	return IsValidFeature(tenantId, feature)
}

func DoesFeatureNameExistForAnotherEntityId(tenantId string, featureEntity *rfc.FeatureEntity) bool {
	feature := featureEntity.CreateFeature()
	return DoesFeatureNameExistForAnotherId(tenantId, feature)
}

func DoesFeatureNameExistForAnotherId(tenantId string, feature *rfc.Feature) bool {
	featureList := rfc.GetFeatureList(tenantId)
	return DoesFeatureNameExistForAnotherIdInList(feature, featureList)
}

func DeleteFeatureRule(tenantId string, id string) {
	err := db.GetCachedSimpleDao().DeleteOne(tenantId, db.TABLE_FEATURE_CONTROL_RULES, id)
	if err != nil {
		log.Warn("delete featureRule failed")
	}
}
