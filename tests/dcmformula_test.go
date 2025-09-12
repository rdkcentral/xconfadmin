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
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"

	xcommon "github.com/rdkcentral/xconfadmin/common"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	core "github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	"gotest.tools/assert"
)

var jsondfCreateData = []byte(
	`{
   "negated":false,
   "condition":{
      "freeArg":{
         "type":"STRING",
         "name":"estbIP"
      },
      "operation":"IS",
      "fixedArg":{
         "bean":{
            "value":{
               "java.lang.String":"3.3.3.3"
            }
         }
      }
   },
   "compoundParts":[
      
   ],
   "id":"33af3261-d74a-40fd-8aa1-884e4f5479a1",
   "name":"dineshtest3",
   "priority":1,
   "percentage":100,
   "percentageL1":10,
   "percentageL2":10,
   "percentageL3":80,
   "applicationType":"stb"
}`)

var jsondfPostCreateData = []byte(
	`{
   "negated":false,
   "compoundParts":[
      {
         "negated":false,
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbIP"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14.14.14.14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      },
      {
         "negated":false,
         "relation":"OR",
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbMacAddress"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14:14:14:14:14:14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      }
   ],
   "id":"3f81ab29-ab8e-40d5-b407-cbc579b46caa",
   "name":"dinesh14",
   "priority":2,
   "percentage":100,
   "percentageL1":10,
   "percentageL2":10,
   "percentageL3":80,
   "applicationType":"stb"
}`)

var jsondfUpdateData = []byte(
	`{
   "negated":false,
   "compoundParts":[
      {
         "negated":false,
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbIP"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14.14.14.14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      },
      {
         "negated":false,
         "relation":"OR",
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbMacAddress"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14:14:14:14:14:14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      }
   ],
   "id":"3f81ab29-ab8e-40d5-b407-cbc579b46caa",
   "name":"dinesh14update",
   "priority":3,
   "percentage":100,
   "percentageL1":20,
   "percentageL2":20,
   "percentageL3":60,
   "applicationType":"stb"
}`)

var jsondfUpdateErrData = []byte(
	`{
   "negated":false,
   "compoundParts":[
      {
         "negated":false,
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbIP"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14.14.14.14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      },
      {
         "negated":false,
         "relation":"OR",
         "condition":{
            "freeArg":{
               "type":"STRING",
               "name":"estbMacAddress"
            },
            "operation":"IS",
            "fixedArg":{
               "bean":{
                  "value":{
                     "java.lang.String":"14:14:14:14:14:14"
                  }
               }
            }
         },
         "compoundParts":[

         ]
      }
   ],
   "id":"3f81ab29-ab8e-40d5-b407-cbc579b46caaer",
   "name":"dinesh14update",
   "priority":3,
   "percentage":100,
   "percentageL1":20,
   "percentageL2":20,
   "percentageL3":60,
   "applicationType":"stb"
}`)

var payload = []byte(`["3f81ab29-ab8e-40d5-b407-cbc579b46caa"]`)
var postmapname = []byte(`{"NAME": "din"}`)
var postmapIPargs = []byte(`{"FIXED_ARG": "3","FREE_ARG": "IP"}`)
var postmapMACargs = []byte(`{"FIXED_ARG": "14","FREE_ARG": "MAC"}`)

const (
	DF_URL = "/xconfAdminService/dcm/formula"
)

func TestDfAllApi(t *testing.T) {
	t.Skip("TODO: cpatel550 - need to move this test under adminapi")
	config := GetTestConfig()
	_, router := GetTestWebConfigServer(config)
	dfrule := logupload.DCMGenericRule{}
	err := json.Unmarshal([]byte(jsondfCreateData), &dfrule)
	assert.NilError(t, err)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, dfrule.ID, &dfrule)

	// get dfrule by id
	urlWithId := fmt.Sprintf("%s/%s", DF_URL, "33af3261-d74a-40fd-8aa1-884e4f5479a1?applicationType=stb")
	req, err := http.NewRequest("GET", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// get dfrule size
	urlWithsize := fmt.Sprintf("%s/%s", DF_URL, "size")
	req, err = http.NewRequest("GET", urlWithsize, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var size string
		json.Unmarshal(body, &size)
		total, _ := strconv.Atoi(size)
		assert.Equal(t, total, 1)
	}

	// get dfrule Names
	urlWithnames := fmt.Sprintf("%s/%s", DF_URL, "names")
	req, err = http.NewRequest("GET", urlWithnames, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		names := []string{}
		json.Unmarshal(body, &names)
		assert.Equal(t, len(names) > 0, true)
	}

	// get dfrule all
	req, err = http.NewRequest("GET", DF_URL, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	if res.StatusCode == http.StatusOK {
		var dfrules = []*logupload.DCMGenericRule{}
		json.Unmarshal(body, &dfrules)
		assert.Equal(t, len(dfrules), 1)
	}

	// import dfrule with settings  for false means create
	urlWithImport := fmt.Sprintf("%s/%s", DF_URL, "import/false")

	impdatacr := []byte(
		`{"formula":{"compoundParts":[{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"SR203"}}}},"negated":false},{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"APPLE123"}}}},"negated":false,"relation":"OR"},{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"SKXI11AIS"}}}},"negated":false,"relation":"OR"}],"negated":false,"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"Dinesh_importgo3_formula","description":"","priority":1,"ruleExpression":"","percentage":100,"percentageL1":60,"percentageL2":20,"percentageL3":20,"applicationType":"stb"},"deviceSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-device","checkOnReboot":true,"configurationServiceURL":{"id":"","name":"","description":"","url":""},"settingsAreActive":true,"schedule":{"type":"CronExpression","expression":"0 8 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"applicationType":"stb"},"logUploadSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-log-upload","uploadOnReboot":true,"numberOfDays":100,"areSettingsActive":true,"schedule":{"type":"CronExpression","expression":"0 10 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"logFileIds":null,"logFilesGroupId":"","modeToGetLogFiles":"","uploadRepositoryId":"d49f4010-eb35-450a-927c-a4be8b68459a","activeDateTimeRange":false,"fromDateTime":"","toDateTime":"","applicationType":"stb"},"vodSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-vod","locationsURL":"https://test.net","ipNames":[],"ipList":[],"srmIPList":{},"applicationType":"stb"}}`)

	req, err = http.NewRequest("POST", urlWithImport+"?applicationType=stb", bytes.NewBuffer(impdatacr))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	// import dfrule with settings  for true means update
	urlWithImportup := fmt.Sprintf("%s/%s", DF_URL, "import/true")

	impdataup := []byte(
		`{"formula":{"compoundParts":[{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"SR203"}}}},"negated":false},{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"APPLE123"}}}},"negated":false,"relation":"OR"},{"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"LIKE","fixedArg":{"bean":{"value":{"java.lang.String":"SKXI11AIS"}}}},"negated":false,"relation":"OR"}],"negated":false,"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"Dinesh_importgo3_formula_update","description":"","priority":1,"ruleExpression":"","percentage":100,"percentageL1":60,"percentageL2":20,"percentageL3":20,"applicationType":"stb"},"deviceSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-device_update","checkOnReboot":true,"configurationServiceURL":{"id":"","name":"","description":"","url":""},"settingsAreActive":true,"schedule":{"type":"CronExpression","expression":"0 8 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"applicationType":"stb"},"logUploadSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-log-upload_update","uploadOnReboot":true,"numberOfDays":100,"areSettingsActive":true,"schedule":{"type":"CronExpression","expression":"0 10 * * *","timeZone":"UTC","expressionL1":"","expressionL2":"","expressionL3":"","startDate":"","endDate":"","timeWindowMinutes":0},"logFileIds":null,"logFilesGroupId":"","modeToGetLogFiles":"","uploadRepositoryId":"d49f4010-eb35-450a-927c-a4be8b68459a","activeDateTimeRange":false,"fromDateTime":"","toDateTime":"","applicationType":"stb"},"vodSettings":{"id":"e0d99b78-e394-45fb-a3b5-178445a3ego3","updated":0,"name":"dinesh_importgo3-vod_update","locationsURL":"https://test.net","ipNames":[],"ipList":[],"srmIPList":{},"applicationType":"stb"}}`)

	req, err = http.NewRequest("POST", urlWithImportup+"?applicationType=stb", bytes.NewBuffer(impdataup))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)

	// POST filtered Name
	urlfiltnames := fmt.Sprintf("%s/%s", DF_URL, "filtered?pageNumber=1&pageSize=50")
	req, err = http.NewRequest("POST", urlfiltnames, bytes.NewBuffer(postmapname))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var dfrules = []*logupload.DCMGenericRule{}
		json.Unmarshal(body, &dfrules)
		assert.Equal(t, len(dfrules), 2)
	}

	// filtered IP Arg
	urlfiltIParg := fmt.Sprintf("%s/%s", DF_URL, "filtered?pageNumber=1&pageSize=50")
	req, err = http.NewRequest("POST", urlfiltIParg, bytes.NewBuffer(postmapIPargs))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var dfrules = []*logupload.DCMGenericRule{}
		json.Unmarshal(body, &dfrules)
		assert.Equal(t, len(dfrules), 1)
	}

	// create entry
	req, err = http.NewRequest("POST", DF_URL+"?applicationType=stb", bytes.NewBuffer(jsondfPostCreateData))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	// change priority
	priourl := "/xconfAdminService/dcm/formula/3f81ab29-ab8e-40d5-b407-cbc579b46caa/priority/1?applicationType=stb"
	req, err = http.NewRequest("POST", priourl, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var dfrules = []*logupload.DCMGenericRule{}
		json.Unmarshal(body, &dfrules)
		assert.Equal(t, len(dfrules) > 0, true)
	}

	//filreerd  MAC Args
	urlfiltMACarg := fmt.Sprintf("%s/%s", DF_URL, "filtered?pageNumber=1&pageSize=50?applicationType=stb")
	req, err = http.NewRequest("POST", urlfiltMACarg, bytes.NewBuffer(postmapMACargs))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		var dfrules = []*logupload.DCMGenericRule{}
		json.Unmarshal(body, &dfrules)
		assert.Equal(t, len(dfrules), 1)
	}

	//settings Availability
	urlWithsetavail := fmt.Sprintf("%s/%s", DF_URL, "settingsAvailability?applicationType=stb")
	req, err = http.NewRequest("POST", urlWithsetavail, bytes.NewBuffer(payload))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		ret := make(map[string]map[string]bool)
		json.Unmarshal(body, &ret)
		assert.Equal(t, len(ret) > 0, true)
	}

	//formulas Availability
	urlWithavail := fmt.Sprintf("%s/%s", DF_URL, "formulasAvailability")
	req, err = http.NewRequest("POST", urlWithavail, bytes.NewBuffer(payload))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	if res.StatusCode == http.StatusOK {
		ret := make(map[string]bool)
		json.Unmarshal(body, &ret)
		assert.Equal(t, len(ret) > 0, true)
	}

	//Error create duplicate Entry
	req, err = http.NewRequest("POST", DF_URL+"?applicationType=stb", bytes.NewBuffer(jsondfPostCreateData))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	//  Update  entry good case
	req, err = http.NewRequest("PUT", DF_URL+"?applicationType=stb", bytes.NewBuffer(jsondfUpdateData))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	//  Update  entry error case
	req, err = http.NewRequest("PUT", DF_URL+"?applicationType=stb", bytes.NewBuffer(jsondfUpdateErrData))
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusConflict)

	// delete dfrule by id
	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNoContent)

	// delete non existing dfrule by id
	req, err = http.NewRequest("DELETE", urlWithId, nil)
	assert.NilError(t, err)
	req.Header.Set("Content-Type", "application/json: charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	res = ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
}

func TestUpdatePriorityAndRuleInFormula_RuleIsUpdatedAndPrioritiesAreReorganized(t *testing.T) {
	DeleteAllEntities()
	numberOfFormulas := 10
	formulas := preCreateFormulas(numberOfFormulas, "TEST_MODEL", t)

	formulaToChangeIndex := 7
	var formulaToUpdate *xcommon.DCMGenericRule
	b, _ := json.Marshal(formulas[formulaToChangeIndex])
	json.Unmarshal(b, &formulaToUpdate)
	newPriority := 10
	formulaToUpdate.Priority = newPriority
	formulaToUpdate.Rule = *CreateRule(rulesengine.RelationAnd, *coreef.RuleFactoryIP, rulesengine.StandardOperationIs, "10.10.10.10")

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/dcm/formula?%v", queryParams)

	formulaJson, _ := json.Marshal(formulaToUpdate)
	r := httptest.NewRequest("PUT", url, bytes.NewReader(formulaJson))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	receivedFormula := unmarshalFormula(rr.Body.Bytes())
	assert.Equal(t, newPriority, receivedFormula.Priority)
	assert.Equal(t, "10.10.10.10", receivedFormula.Rule.Condition.FixedArg.GetValue().(string))

	url = fmt.Sprintf("/xconfAdminService/dcm/formula/%s?%v", receivedFormula.ID, queryParams)
	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	receivedFormula = unmarshalFormula(rr.Body.Bytes())
	assert.Equal(t, formulaToUpdate.ID, receivedFormula.ID)
	assert.Equal(t, newPriority, receivedFormula.Priority)
	assert.Equal(t, "10.10.10.10", receivedFormula.Rule.Condition.FixedArg.GetValue().(string))

	url = fmt.Sprintf("/xconfAdminService/dcm/formula?%v", queryParams)
	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	receivedFormulas := unmarshalFormulas(rr.Body.Bytes())
	assert.Equal(t, numberOfFormulas, len(receivedFormulas))

	sort.Slice(receivedFormulas, func(i, j int) bool {
		return receivedFormulas[i].Priority < receivedFormulas[j].Priority
	})

	for i, formula := range receivedFormulas {
		assert.Equal(t, i+1, formula.Priority)
	}
}

func TestChangeFormulaPriorityWithNotValidValue_ExceptionIsThrown(t *testing.T) {
	DeleteAllEntities()
	formula := createFormula("MODEL_ID", 0)
	saveFormula(formula, t)
	newPriority := 0
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/dcm/formula/%s/priority/%v?%v", formula.ID, newPriority, queryParams)

	r := httptest.NewRequest("POST", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	xconfError := unmarshalXconfError(rr.Body.Bytes())
	assert.Equal(t, fmt.Sprintf("Invalid priority value %v", newPriority), xconfError.Message)
}

func preCreateFormulas(numberOfFormulas int, modelId string, t *testing.T) []*logupload.DCMGenericRule {
	createdFormulas := []*logupload.DCMGenericRule{}
	for i := 0; i < numberOfFormulas; i++ {
		formula := createFormula(modelId, i)
		saveFormula(formula, t)
		createdFormulas = append(createdFormulas, formula)
	}
	return createdFormulas
}

func createFormula(modelId string, testIndex int) *logupload.DCMGenericRule {
	model := CreateAndSaveModel(strings.ToUpper(fmt.Sprintf(modelId+"%v", testIndex)))
	formula := logupload.DCMGenericRule{}
	formula.ID = uuid.New().String()
	formula.Name = fmt.Sprintf("TEST_FORMULA_%v", testIndex)
	formula.Description = fmt.Sprintf("TEST_DESCRIPTION_%v", testIndex)
	formula.ApplicationType = core.STB
	formula.Rule = *CreateRule(rulesengine.RelationAnd, *coreef.RuleFactoryMODEL, rulesengine.StandardOperationIs, model.ID)
	formula.Priority = testIndex + 1
	formula.Percentage = 100
	return &formula
}

func saveFormula(formula *logupload.DCMGenericRule, t *testing.T) {
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/dcm/formula?%v", queryParams)

	formulaJson, _ := json.Marshal(formula)
	r := httptest.NewRequest("POST", url, bytes.NewReader(formulaJson))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func unmarshalFormula(b []byte) *logupload.DCMGenericRule {
	var formula logupload.DCMGenericRule
	err := json.Unmarshal(b, &formula)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling formula: %v", err))
	}
	return &formula
}

func unmarshalFormulas(b []byte) []*logupload.DCMGenericRule {
	var formulas []*logupload.DCMGenericRule
	err := json.Unmarshal(b, &formulas)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling formulas: %v", err))
	}
	return formulas
}
