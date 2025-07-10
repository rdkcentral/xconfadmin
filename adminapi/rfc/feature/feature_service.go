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
package feature

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	xhttp "xconfadmin/http"

	xcommon "xconfadmin/common"
	xrfc "xconfadmin/shared/rfc"
	xwcommon "xconfwebconfig/common"
	xwrfc "xconfwebconfig/shared/rfc"

	"github.com/google/uuid"
)

func GetAllFeature() []*xwrfc.Feature {
	featureList := xwrfc.GetFeatureList()
	if featureList == nil {
		featureList = make([]*xwrfc.Feature, 0)
	}
	return featureList
}

func GetFeatureById(id string) *xwrfc.Feature {
	return xwrfc.GetOneFeature(id)
}

func GetFeatureEntityById(id string) *xwrfc.FeatureEntity {
	feature := xwrfc.GetOneFeature(id)
	return feature.CreateFeatureEntity()
}

func PutFeature(feature *xwrfc.Feature) (*xwrfc.Feature, error) {
	return xrfc.SetOneFeature(feature)
}

func FeaturePost(feature *xwrfc.Feature) (*xwrfc.Feature, error) {
	if feature.ID == "" {
		feature.ID = uuid.New().String()
	}
	return xrfc.SetOneFeature(feature)
}

func GetFeatureFiltered(searchContext map[string]string) []*xwrfc.Feature {
	featureList := xrfc.GetFilteredFeatureList(searchContext)
	if featureList == nil {
		featureList = make([]*xwrfc.Feature, 0)
	}
	return featureList
}

func DeleteFeatureById(id string) {
	xrfc.DeleteOneFeature(id)
}

func IsFeatureUsedInFeatureRule(id string) (bool, string) {
	featureRules := xwrfc.GetFeatureRuleList()
	for _, featureRule := range featureRules {
		for _, featureId := range featureRule.FeatureIds {
			if featureId == id {
				return true, featureRule.Name
			}
		}
	}
	return false, ""
}

func GetFeaturesByApplicationTypeSorted(applicationType string) []*xwrfc.Feature {
	contextMap := map[string]string{xwcommon.APPLICATION_TYPE: applicationType}
	featureList := xrfc.GetFilteredFeatureList(contextMap)
	if featureList == nil {
		featureList = make([]*xwrfc.Feature, 0)
	}
	sort.SliceStable(featureList, func(i, j int) bool {
		return strings.Compare(strings.ToLower(featureList[i].Name), strings.ToLower(featureList[j].Name)) < 0
	})
	return featureList
}

func GetFeatureEntityListByApplicationTypeSorted(applicationType string) []*xwrfc.FeatureEntity {
	contextMap := map[string]string{xwcommon.APPLICATION_TYPE: applicationType}
	featureEntityList := xrfc.GetFilteredFeatureEntityList(contextMap)
	if featureEntityList == nil {
		featureEntityList = make([]*xwrfc.FeatureEntity, 0)
	}
	sort.SliceStable(featureEntityList, func(i, j int) bool {
		return strings.Compare(strings.ToLower(featureEntityList[i].Name), strings.ToLower(featureEntityList[j].Name)) < 0
	})
	return featureEntityList
}

func GetFeaturesByIdList(featureIdList []string) []*xwrfc.Feature {
	features := []*xwrfc.Feature{}
	for _, featureId := range featureIdList {
		feature := GetFeatureById(featureId)
		if feature != nil {
			features = append(features, feature)
		}
	}
	return features
}

func ImportFeatureEntities(featureEntityList []*xwrfc.FeatureEntity, overwrite bool, applicationType string) map[string]xhttp.EntityMessage {
	entitiesMap := map[string]xhttp.EntityMessage{}
	var err error
	for _, featureEntity := range featureEntityList {
		feature := featureEntity.CreateFeature()
		if overwrite {
			err = UpdateEntity(feature, applicationType)
		} else {
			err = CreateEntity(feature, applicationType)
		}
		if err != nil {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_FAILURE,
				Message: err.Error(),
			}
			entitiesMap[feature.ID] = entityMessage
		} else {
			entityMessage := xhttp.EntityMessage{
				Status:  xcommon.ENTITY_STATUS_SUCCESS,
				Message: feature.ID,
			}
			entitiesMap[feature.ID] = entityMessage
		}
	}
	return entitiesMap
}

func CreateEntity(feature *xwrfc.Feature, applicationType string) error {
	if feature.ID == "" {
		feature.ID = uuid.New().String()
	} else {
		doesFeatureExist, appType := xrfc.DoesFeatureExistInSomeApplicationType(feature.ID)
		if doesFeatureExist && feature.ApplicationType != appType {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Entity with id: %s already exists in %s", feature.ID, appType))
		}
		if doesFeatureExist && applicationType == appType {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Entity with id: %s already exists", feature.ID))
		}
		if feature.ApplicationType == "" {
			feature.ApplicationType = applicationType
		}
	}
	// TODO add call to permissionService for validateWrite
	isValid, errorMsg := xrfc.IsValidFeature(feature)
	if !isValid {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, errorMsg)
	}
	doesFeatureInstanceExist := xrfc.DoesFeatureNameExistForAnotherIdForApplicationType(feature, applicationType)
	if doesFeatureInstanceExist {
		return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Feature with such featureInstance already exists: %s", feature.FeatureName))
	}
	feature, err := FeaturePost(feature)
	if err != nil {
		return err
	}
	return nil
}

func UpdateEntity(feature *xwrfc.Feature, applicationType string) error {
	if feature.ID == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Entity id is empty")
	}
	doesFeatureExist, appType := xrfc.DoesFeatureExistInSomeApplicationType(feature.ID)
	if !doesFeatureExist || applicationType != appType {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("Entity with id: %s does not exist", feature.ID))
	}
	if doesFeatureExist && feature.ApplicationType != appType {
		return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Entity with id: %s already exists in %s", feature.ID, appType))
	}
	if feature.ApplicationType == "" {
		feature.ApplicationType = applicationType
	}
	// TODO add call to permissionService for validateWrite
	isValid, errorMsg := xrfc.IsValidFeature(feature)
	if !isValid {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, errorMsg)
	}
	doesFeatureInstanceExist := xrfc.DoesFeatureNameExistForAnotherIdForApplicationType(feature, applicationType)
	if doesFeatureInstanceExist {
		return xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Feature with such featureInstance already exists: %s", feature.FeatureName))
	}
	feature, err := PutFeature(feature)
	if err != nil {
		return err
	}
	return nil
}

func GetFeaturesWithPageNumbers(features []*xwrfc.Feature, pageNumber int, pageSize int) []*xwrfc.Feature {
	sort.SliceStable(features, func(i, j int) bool {
		return strings.Compare(strings.ToLower(features[i].Name), strings.ToLower(features[j].Name)) < 0
	})
	featurePageList := make([]*xwrfc.Feature, 0)
	startIndex := pageNumber*pageSize - pageSize
	if pageNumber < 1 || startIndex > len(features) {
		return featurePageList
	}
	lastIndex := len(features)
	if pageNumber*pageSize < len(features) {
		lastIndex = pageNumber * pageSize
	}
	featurePageList = features[startIndex:lastIndex]
	return featurePageList
}

func DoesFeatureInstanceExistForAnotherId(feature *xwrfc.Feature) bool {
	all := GetAllFeature()
	for _, feature := range all {
		if feature.ID != feature.ID && feature.ApplicationType == feature.ApplicationType && feature.FeatureName == feature.FeatureName {
			return true
		}
	}
	return false
}

func DoesFeatureExist(id string) bool {
	if id == "" {
		return false
	}
	feature := xwrfc.GetOneFeature(id)
	return feature != nil
}
