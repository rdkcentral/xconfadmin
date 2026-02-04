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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	oscommon "github.com/rdkcentral/xconfadmin/common"

	"github.com/rdkcentral/xconfwebconfig/shared/rfc"

	"gotest.tools/assert"
)

func TestImportFeatureSecondTimeWithDiffAppType(t *testing.T) {
	SkipIfMockDatabase(t)
	DeleteAllEntities()

	featureDiffAppType := &rfc.FeatureEntity{
		Name:            "nameAppType",
		FeatureName:     "featureInstanceAppType",
		ID:              "idAppType",
		ApplicationType: "stb",
	}

	featureEntityList := []*rfc.FeatureEntity{featureDiffAppType}
	jsonByte, _ := json.Marshal(featureEntityList)
	url := "/xconfAdminService/feature/importAll?applicationType=stb"
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(jsonByte)))
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// Importing existing feature with different application-type should fail
	featureDiffAppType.ApplicationType = "different_application"
	featureEntityList = []*rfc.FeatureEntity{featureDiffAppType}
	jsonByte, _ = json.Marshal(featureEntityList)
	req, _ = http.NewRequest("POST", url, strings.NewReader(string(jsonByte)))
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusConflict)
}
func TestAllFeatureHandlers(t *testing.T) {
	SkipIfMockDatabase(t)

	featureEntity1 := &rfc.FeatureEntity{
		Name:        "name1",
		FeatureName: "featureInstance1",
		ID:          "id1",
		ConfigData: map[string]string{
			"key": "value",
		},
	}

	featureEntity2 := &rfc.FeatureEntity{
		Name:            "name2",
		FeatureName:     "featureInstance2",
		ID:              "id2",
		ApplicationType: "fakeApplicationType",
	}

	featureEntity3 := &rfc.FeatureEntity{
		Name:        "name3",
		FeatureName: "featureInstance3",
		ID:          "id3",
	}

	featureEntity4 := &rfc.FeatureEntity{
		Name:            "name4",
		FeatureName:     "featureInstance4",
		ID:              "id4",
		ApplicationType: "stb",
	}

	DeleteAllEntities()

	// no data, GET empty 200 response
	url := "/xconfAdminService/feature?applicationType=stb"
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, string(body), "[]")
	res.Body.Close()

	// no body, POST 400 bad request
	req, err = http.NewRequest("POST", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)

	// invalid applicationType, POST 400 bad request
	jsonByte, err := json.Marshal(featureEntity2)
	req, err = http.NewRequest("POST", url, strings.NewReader(string(jsonByte)))
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	// good request POST 201 created (no applicationType specified, should default on stb)
	jsonByte, err = json.Marshal(featureEntity1)
	req, err = http.NewRequest("POST", url, strings.NewReader(string(jsonByte)))
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusCreated)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	var featureEntityResponse1 *rfc.FeatureEntity
	err = json.Unmarshal(body, &featureEntityResponse1)
	assert.NilError(t, err)
	// add stb to featureEntity1 for comparison
	featureEntity1.ApplicationType = "stb"
	compareFeatureEntityObjects(t, featureEntityResponse1, featureEntity1)
	res.Body.Close()

	// bad request, feature already exists, POST 409
	req, err = http.NewRequest("POST", url, strings.NewReader(string(jsonByte)))
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusConflict)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, string(body), "\"Entity with id: id1 already exists\"")
	res.Body.Close()

	// good request GET 200 with response
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	var featureEntityList []*rfc.FeatureEntity
	err = json.Unmarshal([]byte(body), &featureEntityList)
	assert.NilError(t, err)
	assert.Equal(t, len(featureEntityList), 1)
	compareFeatureEntityObjects(t, featureEntity1, featureEntityList[0])
	res.Body.Close()

	// data that doesn't match filter, GET /filtered 200 empty response
	url = "/xconfAdminService/feature/filtered?applicationType=rdkcloud"
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, string(body), "[]")
	res.Body.Close()

	// good request for 2nd feature, POST 201 created
	featureEntity2.ApplicationType = "rdkcloud"
	jsonByte, err = json.Marshal(featureEntity2)
	url = "/xconfAdminService/feature?applicationType=rdkcloud"
	req, err = http.NewRequest("POST", url, strings.NewReader(string(jsonByte)))
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusCreated)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	var featureEntityResponse2 *rfc.FeatureEntity
	err = json.Unmarshal(body, &featureEntityResponse2)
	assert.NilError(t, err)
	compareFeatureEntityObjects(t, featureEntityResponse2, featureEntity2)
	res.Body.Close()

	// good request for 3rd feature, POST 201 created
	featureEntity3.ApplicationType = "stb"
	jsonByte, err = json.Marshal(featureEntity3)
	url = "/xconfAdminService/feature?applicationType=stb"
	req, err = http.NewRequest("POST", url, strings.NewReader(string(jsonByte)))
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusCreated)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	var featureEntityResponse3 *rfc.FeatureEntity
	err = json.Unmarshal(body, &featureEntityResponse3)
	assert.NilError(t, err)
	compareFeatureEntityObjects(t, featureEntityResponse3, featureEntity3)
	res.Body.Close()

	// data where one matches filter, GET /filter 200 with response
	url = "/xconfAdminService/feature/filtered?applicationType=rdkcloud"
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	err = json.Unmarshal([]byte(body), &featureEntityList)
	assert.NilError(t, err)
	assert.Equal(t, len(featureEntityList), 1)
	compareFeatureEntityObjects(t, featureEntity2, featureEntityList[0])
	res.Body.Close()

	// data where match on multiple filters, GET /filter 200 with response
	url = "/xconfAdminService/feature/filtered?applicationType=stb&FREE_ARG=key&FIXED_ARG=value"
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	err = json.Unmarshal([]byte(body), &featureEntityList)
	assert.NilError(t, err)
	assert.Equal(t, len(featureEntityList), 1)
	compareFeatureEntityObjects(t, featureEntity1, featureEntityList[0])
	res.Body.Close()

	// feature does not exists, GET /{id} 404 not found
	url = "/xconfAdminService/feature/fakeFeatureId?applicationType=stb"
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
	res.Body.Close()

	// feature exists, GET /{id} 200 with response
	url = fmt.Sprintf("/xconfAdminService/feature/%s"+"?applicationType=stb", featureEntity1.ID)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	var featureEntity *rfc.FeatureEntity
	err = json.Unmarshal([]byte(body), &featureEntity)
	assert.NilError(t, err)
	compareFeatureEntityObjects(t, featureEntity1, featureEntity)
	res.Body.Close()

	// no body, PUT 400 bad request
	url = "/xconfAdminService/feature?applicationType=stb"
	req, err = http.NewRequest("PUT", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)

	// no featureId, PUT 400 bad requst
	featureEntity3.ID = ""
	jsonByte, err = json.Marshal(featureEntity3)
	url = "/xconfAdminService/feature?applicationType=stb"
	req, err = http.NewRequest("PUT", url, strings.NewReader(string(jsonByte)))
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, string(body), "\"Entity id is empty\"")
	res.Body.Close()
	featureEntity3.ID = "id3"

	// feature doesn't exist, PUT 400 bad request
	featureEntity4.ID = "id4"
	jsonByte, err = json.Marshal(featureEntity4)
	url = "/xconfAdminService/feature?applicationType=stb"
	req, err = http.NewRequest("PUT", url, strings.NewReader(string(jsonByte)))
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, string(body), "\"Entity with id: id4 does not exist\"")
	res.Body.Close()

	// no feature name, PUT 400 bad request
	featureEntity3.Name = ""
	jsonByte, err = json.Marshal(featureEntity3)
	url = "/xconfAdminService/feature?applicationType=stb"
	req, err = http.NewRequest("PUT", url, strings.NewReader(string(jsonByte)))
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, string(body), "\"Name is blank\"")
	res.Body.Close()

	// featureInstance already exists on another feature, PUT 409 conflict
	featureEntity3.Name = "name3"
	featureEntity3.FeatureName = "featureInstance1"
	jsonByte, err = json.Marshal(featureEntity3)
	url = "/xconfAdminService/feature?applicationType=stb"
	req, err = http.NewRequest("PUT", url, strings.NewReader(string(jsonByte)))
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	// assert.Equal(t, res.StatusCode, http.StatusConflict)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, string(body), "\"Feature with such featureInstance already exists: featureInstance1\"")
	res.Body.Close()

	// good request, PUT 200 OK
	featureEntity2.FeatureName = "featureInstance2"
	featureEntity2.ConfigData = map[string]string{
		"key": "value",
	}
	jsonByte, err = json.Marshal(featureEntity2)
	url = "/xconfAdminService/feature?applicationType=rdkcloud"
	req, err = http.NewRequest("PUT", url, strings.NewReader(string(jsonByte)))
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	err = json.Unmarshal(body, &featureEntityResponse2)
	assert.NilError(t, err)
	compareFeatureEntityObjects(t, featureEntityResponse2, featureEntity2)
	res.Body.Close()

	// import: one good PUT (featureEntity1), one good POST (featureEntity4),
	// one invalid feature (feature5, invalid applicationType), one featureInstance already exists (feature6),
	featureEntity1.Name = "newName1"
	featureEntity6 := &rfc.FeatureEntity{
		Name:            "name6",
		FeatureName:     "featureInstance1",
		ID:              "id6",
		ApplicationType: "stb",
	}

	featureEntityList = []*rfc.FeatureEntity{featureEntity1, featureEntity4, featureEntity6}
	jsonByte, err = json.Marshal(featureEntityList)
	url = "/xconfAdminService/feature/importAll?applicationType=stb"
	req, err = http.NewRequest("POST", url, strings.NewReader(string(jsonByte)))
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	res.Body.Close()

	// check which features made it into DB
	url = fmt.Sprintf("/xconfAdminService/feature/%s"+"?applicationType=stb", featureEntity1.ID)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	err = json.Unmarshal([]byte(body), &featureEntity)
	assert.NilError(t, err)
	compareFeatureEntityObjects(t, featureEntity1, featureEntity)
	res.Body.Close()

	url = fmt.Sprintf("/xconfAdminService/feature/%s"+"?applicationType=stb", featureEntity4.ID)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	err = json.Unmarshal([]byte(body), &featureEntity)
	assert.NilError(t, err)
	compareFeatureEntityObjects(t, featureEntity4, featureEntity)
	res.Body.Close()

	url = fmt.Sprintf("/xconfAdminService/feature/%s"+"?applicationType=stb", featureEntity6.ID)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

	// feature doesn't exist, GET /{id} 404 not found
	url = "/xconfAdminService/feature/someFakeId?applicationType=stb"
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	assert.Equal(t, string(body), "\"Entity with id: someFakeId does not exist\"")
	res.Body.Close()

	// feature doesn't exist, DELETE /{id} 404 not found
	url = "/xconfAdminService/feature/someFakeId?applicationType=stb"
	req, err = http.NewRequest("DELETE", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	errRsp := oscommon.HttpAdminErrorResponse{}
	err = json.Unmarshal(body, &errRsp)
	assert.Equal(t, errRsp.Message, "Entity with id: someFakeId does not exist")
	res.Body.Close()

	// feature does exist, DELETE /{id} 204 accepted, delete features
	url = "/xconfAdminService/feature/id1?applicationType=stb"
	req, err = http.NewRequest("DELETE", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)
	res.Body.Close()

	url = "/xconfAdminService/feature/id2?applicationType=rdkcloud"
	req, err = http.NewRequest("DELETE", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)
	res.Body.Close()

	url = "/xconfAdminService/feature/id3?applicationType=stb"
	req, err = http.NewRequest("DELETE", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)
	res.Body.Close()

	url = "/xconfAdminService/feature/id4?applicationType=stb"
	req, err = http.NewRequest("DELETE", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)
	res.Body.Close()

	url = "/xconfAdminService/feature/id5?applicationType=stb"
	req, err = http.NewRequest("DELETE", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
	res.Body.Close()

	url = "/xconfAdminService/feature/id6?applicationType=stb"
	req, err = http.NewRequest("DELETE", url, nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
	res.Body.Close()

}

func compareFeatureEntityObjects(t *testing.T, featureEntity1 *rfc.FeatureEntity, featureEntity2 *rfc.FeatureEntity) {
	assert.Equal(t, featureEntity1.ID, featureEntity2.ID)
	assert.Equal(t, featureEntity1.Name, featureEntity2.Name)
	assert.Equal(t, featureEntity1.FeatureName, featureEntity2.FeatureName)
	assert.Equal(t, featureEntity1.ApplicationType, featureEntity2.ApplicationType)
	assert.Equal(t, len(featureEntity1.ConfigData), len(featureEntity2.ConfigData))
	for key, value := range featureEntity1.ConfigData {
		assert.Equal(t, value, featureEntity2.ConfigData[key])
	}
	assert.Equal(t, featureEntity1.EffectiveImmediate, featureEntity2.EffectiveImmediate)
	assert.Equal(t, featureEntity1.Enable, featureEntity2.Enable)
	assert.Equal(t, featureEntity1.Whitelisted, featureEntity2.Whitelisted)
	if featureEntity1.WhitelistProperty == nil {
		assert.Equal(t, featureEntity2.WhitelistProperty == nil, true)
	} else {
		assert.Equal(t, featureEntity1.WhitelistProperty.Key, featureEntity2.WhitelistProperty.Key)
		assert.Equal(t, featureEntity1.WhitelistProperty.Value, featureEntity2.WhitelistProperty.Value)
		assert.Equal(t, featureEntity1.WhitelistProperty.NamespacedListType, featureEntity2.WhitelistProperty.NamespacedListType)
		assert.Equal(t, featureEntity1.WhitelistProperty.TypeName, featureEntity2.WhitelistProperty.TypeName)
	}
}
