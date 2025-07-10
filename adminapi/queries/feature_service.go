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
package queries

import (
	"fmt"

	xrfc "xconfadmin/shared/rfc"
	xwrfc "xconfwebconfig/shared/rfc"
	"xconfwebconfig/util"

	"github.com/google/uuid"
	errors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func GetAllFeatureEntity() []*xwrfc.FeatureEntity {
	featureEntityList := xrfc.GetFeatureEntityList()
	if featureEntityList == nil {
		featureEntityList = make([]*xwrfc.FeatureEntity, 0)
	}
	return featureEntityList
}

func GetFeatureEntityFiltered(searchContext map[string]string) []*xwrfc.FeatureEntity {
	featureEntityList := xrfc.GetFilteredFeatureEntityList(searchContext)
	if featureEntityList == nil {
		featureEntityList = make([]*xwrfc.FeatureEntity, 0)
	}
	return featureEntityList
}

func GetFeatureEntityById(id string) *xwrfc.FeatureEntity {
	feature := xwrfc.GetOneFeature(id)
	return feature.CreateFeatureEntity()
}

func DeleteFeatureById(id string) {
	xrfc.DeleteOneFeature(id)
}

func ImportOrUpdateAllFeatureEntity(featureEntityList []*xwrfc.FeatureEntity, applicationType string) map[string][]string {
	importedList := []string{}
	notImportedList := []string{}
	for _, featureEntity := range featureEntityList {
		featureEntity := featureEntity
		var err error
		var isValid bool
		var doesExist bool
		var errMsg string
		isValid, errMsg = xrfc.IsValidFeatureEntity(featureEntity)
		if isValid {
			doesExist = xrfc.DoesFeatureNameExistForAnotherEntityId(featureEntity)
			if doesExist {
				errMsg = fmt.Sprintf("Feature with such featureInstance already exists: %s", featureEntity.FeatureName)
			} else {
				if xrfc.DoesFeatureExist(featureEntity.ID) {
					// update feature
					_, err = PutFeatureEntity(featureEntity, applicationType)
				} else {
					// create feature
					featureEntity, err = PostFeatureEntity(featureEntity, applicationType)
				}
			}
			if err != nil {
				errMsg = err.Error()
			}
		}
		if errMsg != "" {
			json, _ := util.JSONMarshal(featureEntity)
			log.Errorf("Exception: %s, with feature: %s", errMsg, json)
			notImportedList = append(notImportedList, featureEntity.ID)
		} else {
			importedList = append(importedList, featureEntity.ID)
		}
	}
	return map[string][]string{
		IMPORTED:     importedList,
		NOT_IMPORTED: notImportedList,
	}
}

func PostFeatureEntity(featureEntity *xwrfc.FeatureEntity, applicationType string) (*xwrfc.FeatureEntity, error) {
	feature := featureEntity.CreateFeature()
	if feature.ID == "" {
		feature.ID = uuid.New().String()
	}
	if applicationType != featureEntity.ApplicationType {
		return nil, errors.New("AplicationType cannot be different: : " + applicationType + " New: " + featureEntity.ApplicationType)
	}
	feature, err := xrfc.SetOneFeature(feature)
	return feature.CreateFeatureEntity(), err
}

func PutFeatureEntity(featureEntity *xwrfc.FeatureEntity, applicationType string) (*xwrfc.FeatureEntity, error) {
	featureOnDb := xwrfc.GetOneFeature(featureEntity.ID)
	if featureOnDb.ApplicationType != featureEntity.ApplicationType {
		return nil, errors.New("AplicationType cannot be different: Old: " + featureOnDb.ApplicationType + " New: " + featureEntity.ApplicationType)
	}
	if applicationType != featureEntity.ApplicationType {
		return nil, errors.New("AplicationType cannot be different: : " + applicationType + " New: " + featureEntity.ApplicationType)
	}
	feature := featureEntity.CreateFeature()
	feature, err := xrfc.SetOneFeature(feature)
	return feature.CreateFeatureEntity(), err
}
