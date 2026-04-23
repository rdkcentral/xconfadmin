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

	xrfc "github.com/rdkcentral/xconfadmin/shared/rfc"

	xwrfc "github.com/rdkcentral/xconfwebconfig/shared/rfc"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	errors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func GetAllFeatureEntity(tenantId string) []*xwrfc.FeatureEntity {
	featureEntityList := xrfc.GetFeatureEntityList(tenantId)
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

func GetFeatureEntityById(tenantId string, id string) *xwrfc.FeatureEntity {
	feature := xwrfc.GetOneFeature(tenantId, id)
	return feature.CreateFeatureEntity()
}

func ImportOrUpdateAllFeatureEntity(tenantId string, featureEntityList []*xwrfc.FeatureEntity, applicationType string) map[string][]string {
	importedList := []string{}
	notImportedList := []string{}
	for _, featureEntity := range featureEntityList {
		featureEntity := featureEntity
		var err error
		var isValid bool
		var doesExist bool
		var errMsg string
		isValid, errMsg = xrfc.IsValidFeatureEntity(tenantId, featureEntity)
		if isValid {
			doesExist = xrfc.DoesFeatureNameExistForAnotherEntityId(tenantId, featureEntity)
			if doesExist {
				errMsg = fmt.Sprintf("Feature with such featureInstance already exists: %s", featureEntity.FeatureName)
			} else {
				if xrfc.DoesFeatureExist(tenantId, featureEntity.ID) {
					// update feature
					_, err = PutFeatureEntity(tenantId, featureEntity, applicationType)
				} else {
					// create feature
					featureEntity, err = PostFeatureEntity(tenantId, featureEntity, applicationType)
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

func PostFeatureEntity(tenantId string, featureEntity *xwrfc.FeatureEntity, applicationType string) (*xwrfc.FeatureEntity, error) {
	feature := featureEntity.CreateFeature()
	if feature.ID == "" {
		feature.ID = uuid.New().String()
	}
	if applicationType != featureEntity.ApplicationType {
		return nil, errors.New("AplicationType cannot be different: : " + applicationType + " New: " + featureEntity.ApplicationType)
	}
	feature, err := xrfc.SetOneFeature(tenantId, feature)
	return feature.CreateFeatureEntity(), err
}

func PutFeatureEntity(tenantId string, featureEntity *xwrfc.FeatureEntity, applicationType string) (*xwrfc.FeatureEntity, error) {
	featureOnDb := xwrfc.GetOneFeature(tenantId, featureEntity.ID)
	if featureOnDb.ApplicationType != featureEntity.ApplicationType {
		return nil, errors.New("AplicationType cannot be different: Old: " + featureOnDb.ApplicationType + " New: " + featureEntity.ApplicationType)
	}
	if applicationType != featureEntity.ApplicationType {
		return nil, errors.New("AplicationType cannot be different: : " + applicationType + " New: " + featureEntity.ApplicationType)
	}
	feature := featureEntity.CreateFeature()
	feature, err := xrfc.SetOneFeature(tenantId, feature)
	return feature.CreateFeatureEntity(), err
}
