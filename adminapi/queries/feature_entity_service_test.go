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
	"testing"
	"time"

	xrfc "github.com/rdkcentral/xconfadmin/shared/rfc"

	"github.com/rdkcentral/xconfwebconfig/shared/rfc"

	"github.com/google/uuid"
	"gotest.tools/assert"
)

func TestFeatureGetPostPutDeleteImport(t *testing.T) {
	DeleteAllEntities()

	// test GET ALL
	featureList := GetAllFeatureEntity()
	assert.Equal(t, len(featureList), 0)

	id1 := uuid.New().String()
	feature1 := &rfc.Feature{
		ID:              id1,
		Name:            "name1",
		FeatureName:     "featureName1",
		ApplicationType: "stb",
		ConfigData: map[string]string{
			"key": "value",
		},
	}

	featureEntity1 := feature1.CreateFeatureEntity()

	// test POST
	applicationType := "stb"
	fe, err := PostFeatureEntity(featureEntity1, applicationType)
	// assert feature returned matches feature passed in
	assertFeatureEntity(t, fe, featureEntity1)
	assert.NilError(t, err)
	// assert feature is in db
	featureList = GetAllFeatureEntity()
	assert.Equal(t, len(featureList), 1)
	assertFeatureEntity(t, featureList[0], featureEntity1)

	// test GET FILTERED
	searchContext := map[string]string{
		"applicationType": "rdkcloud",
	}
	featureList = GetFeatureEntityFiltered(searchContext)
	assert.Equal(t, len(featureList), 0)
	searchContext["applicationType"] = "stb"
	featureList = GetFeatureEntityFiltered(searchContext)
	assert.Equal(t, len(featureList), 1)
	assertFeatureEntity(t, featureList[0], featureEntity1)

	// test GET BY ID
	fe = GetFeatureEntityById(featureEntity1.ID)
	assertFeatureEntity(t, fe, featureEntity1)

	// test PUT
	featureEntity1.Name = "newName"
	fe, err = PutFeatureEntity(featureEntity1, applicationType)
	assertFeatureEntity(t, fe, featureEntity1)
	assert.NilError(t, err)

	fe = GetFeatureEntityById(featureEntity1.ID)
	assertFeatureEntity(t, fe, featureEntity1)

	// test IMPORT
	featureEntity1.Name = "name1"
	id2 := uuid.New().String()
	feature2 := &rfc.Feature{
		ID:              id2,
		Name:            "name2",
		FeatureName:     "featureName2",
		ApplicationType: "stb",
		ConfigData: map[string]string{
			"key": "value",
		},
	}
	featureEntity2 := feature2.CreateFeatureEntity()

	featureEntityList := []*rfc.FeatureEntity{featureEntity1, featureEntity2}
	featureImportMap := ImportOrUpdateAllFeatureEntity(featureEntityList, applicationType)
	assert.Equal(t, len(featureImportMap["IMPORTED"]), 2)
	assert.Equal(t, len(featureImportMap["NOT_IMPORTED"]), 0)
	assert.Equal(t, featureImportMap["IMPORTED"][0], featureEntity1.ID)
	assert.Equal(t, featureImportMap["IMPORTED"][1], featureEntity2.ID)

	// use GET to check import
	fe = GetFeatureEntityById(featureEntity1.ID)
	assertFeatureEntity(t, fe, featureEntity1)

	fe = GetFeatureEntityById(featureEntity2.ID)
	assertFeatureEntity(t, fe, featureEntity2)

	// test DELETE
	DeleteFeatureById(featureEntity1.ID)
	time.Sleep(1 * time.Second)
	fe = GetFeatureEntityById(featureEntity1.ID)
	assert.Equal(t, fe == nil, true)

	DeleteFeatureById(featureEntity2.ID)
	time.Sleep(1 * time.Second)
	fe = GetFeatureEntityById(featureEntity2.ID)
	assert.Equal(t, fe == nil, true)
}

func TestDoesFeatureExist(t *testing.T) {
	DeleteAllEntities()

	doesFeatureExist := xrfc.DoesFeatureExist("")
	assert.Equal(t, doesFeatureExist, false)

	id := uuid.New().String()
	doesFeatureExist = xrfc.DoesFeatureExist(id)
	assert.Equal(t, doesFeatureExist, false)

	// feature := createAndSaveFeature()
	// doesFeatureExist = xrfc.DoesFeatureExist(feature.ID)
	// assert.Equal(t, doesFeatureExist, true)
}

func TestDoesFeatureInstanceExist(t *testing.T) {
	DeleteAllEntities()
	applicationType := "stb"
	id1 := uuid.New().String()
	feature1 := &rfc.Feature{
		ID:              id1,
		Name:            "name",
		FeatureName:     "featureName1",
		ApplicationType: applicationType,
		ConfigData: map[string]string{
			"key": "value",
		},
	}

	featureEntity1 := feature1.CreateFeatureEntity()

	id2 := uuid.New().String()
	feature2 := &rfc.Feature{
		ID:              id2,
		Name:            "name",
		FeatureName:     "featureName2",
		ApplicationType: applicationType,
		ConfigData: map[string]string{
			"key": "value",
		},
	}

	featureEntity2 := feature2.CreateFeatureEntity()

	// no features exist in db
	doesFeatureInstanceExist := xrfc.DoesFeatureNameExistForAnotherEntityId(featureEntity1)
	assert.Equal(t, doesFeatureInstanceExist, false)

	// different feature in db
	PostFeatureEntity(featureEntity2, applicationType)
	doesFeatureInstanceExist = xrfc.DoesFeatureNameExistForAnotherEntityId(featureEntity1)
	assert.Equal(t, doesFeatureInstanceExist, false)

	// diff feature with same featureInstance in db
	featureEntity1.FeatureName = "featureName2"
	doesFeatureInstanceExist = xrfc.DoesFeatureNameExistForAnotherEntityId(featureEntity1)
	assert.Equal(t, doesFeatureInstanceExist, true)

	// same exact feature in db
	featureEntity1.FeatureName = "featureName1"
	PostFeatureEntity(featureEntity1, applicationType)
	doesFeatureInstanceExist = xrfc.DoesFeatureNameExistForAnotherEntityId(featureEntity1)
	assert.Equal(t, doesFeatureInstanceExist, false)
}

func assertFeatureEntity(t *testing.T, fe *rfc.FeatureEntity, featureEntity *rfc.FeatureEntity) {
	assert.Equal(t, fe.ID, featureEntity.ID)
	assert.Equal(t, fe.Name, featureEntity.Name)
	assert.Equal(t, fe.FeatureName, featureEntity.FeatureName)
	assert.Equal(t, fe.ApplicationType, featureEntity.ApplicationType)
	assert.Equal(t, len(fe.ConfigData), len(featureEntity.ConfigData))
	for key, value := range fe.ConfigData {
		assert.Equal(t, value, featureEntity.ConfigData[key])
	}
	assert.Equal(t, fe.EffectiveImmediate, featureEntity.EffectiveImmediate)
	assert.Equal(t, fe.Enable, featureEntity.Enable)
	assert.Equal(t, fe.Whitelisted, featureEntity.Whitelisted)
	if fe.WhitelistProperty == nil {
		assert.Equal(t, featureEntity.WhitelistProperty == nil, true)
	} else {
		assert.Equal(t, fe.WhitelistProperty.Key, featureEntity.WhitelistProperty.Key)
		assert.Equal(t, fe.WhitelistProperty.Value, featureEntity.WhitelistProperty.Value)
		assert.Equal(t, fe.WhitelistProperty.NamespacedListType, featureEntity.WhitelistProperty.NamespacedListType)
		assert.Equal(t, fe.WhitelistProperty.TypeName, featureEntity.WhitelistProperty.TypeName)
	}
}
