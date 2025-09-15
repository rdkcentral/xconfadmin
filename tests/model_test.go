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
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

const (
	MODEL_WHOLE_API            = "/xconfAdminService/model"
	jsonModelWholeTestDataLocn = "jsondata/model/"
)

func newModelWholeApiUnitTest(t *testing.T) *apiUnitTest {
	aut := newApiUnitTest(t)
	aut.setValOf(MODEL_WHOLE_API+DATA_LOCN_SUFFIX, jsonModelWholeTestDataLocn)
	aut.setupModelWholeApi()
	return aut
}

func (aut *apiUnitTest) setupModelWholeApi() {
	if aut.getValOf(MODEL_WHOLE_API) == "Done" {
		return
	}
	aut.setValOf(MODEL_WHOLE_API+DATA_LOCN_SUFFIX, jsonModelWholeTestDataLocn)
	aut.setValOf(MODEL_WHOLE_API, "Done")
}

func (aut *apiUnitTest) cleanupModelWholeApi() {
	if aut.getValOf(MODEL_WHOLE_API) == "" {
		return
	}
	aut.setValOf(MODEL_WHOLE_API, "")
}

func TestModelWholeCRUD(t *testing.T) {
	aut := newModelWholeApiUnitTest(t)
	sysGenId1 := uuid.New().String()
	sysGenId2 := uuid.New().String()

	testCases := []apiUnitTestCase{
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=model_count", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId1, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_one&validate=true", aut.modelResponseValidator},
		{MODEL_WHOLE_API, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_two&validate=true", aut.modelResponseValidator},
		{MODEL_WHOLE_API, "create_missing_id", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
	}
	aut.run(testCases)

	m1 := aut.getValOf("model_id_one")
	m2 := aut.getValOf("model_id_two")

	testCases = []apiUnitTestCase{
		{MODEL_WHOLE_API, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + aut.getValOf("model_id_one"), aut.replaceKeysByValues, "PUT", "", http.StatusOK, "validate=true", aut.modelResponseValidator},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("model_count+2"), aut.modelArrayValidator},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + m1, http.StatusOK, "ID=" + m1, aut.modelSingleValidator},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + m2, http.StatusOK, "ID=" + m2, aut.modelSingleValidator},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + m1, http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + m2, http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + m1, http.StatusNotFound, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + m2, http.StatusNotFound, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + m1, http.StatusNotFound, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + m2, http.StatusNotFound, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("model_count"), aut.modelArrayValidator},
	}
	aut.run(testCases)
}

func TestModelWholeEndPoints(t *testing.T) {
	aut := newModelWholeApiUnitTest(t)
	sysGenId := strings.ToUpper(uuid.New().String())
	sysGenId2 := strings.ToUpper(uuid.New().String())

	testCases := []apiUnitTestCase{
		//	"" PostModelWholeHandler "POST"
		{MODEL_WHOLE_API, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_1&validate=true", aut.modelResponseValidator},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + sysGenId, http.StatusOK, NO_POSTERMS, nil},

		// "/entities" PostModelWholeEntitiesHandler "POST"
		{MODEL_WHOLE_API, "[create_unique_model]", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "POST", "/entities", http.StatusOK, "saveIdIn=model_id_2", aut.modelEntitiesMapValidator},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + sysGenId2, http.StatusOK, NO_POSTERMS, nil},

		//	"/entities" PutModelWholeEntitiesHandler "PUT"
		{MODEL_WHOLE_API, "[create_unique_model]", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "PUT", "/entities", http.StatusOK, NO_POSTERMS, nil},
	}
	aut.run(testCases)
	idCreated1 := aut.getValOf("model_id_1")

	testCases = []apiUnitTestCase{
		//	"" PutModelWholeHandler "PUT"
		{MODEL_WHOLE_API, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + idCreated1, aut.replaceKeysByValues, "PUT", "", http.StatusOK, NO_POSTERMS, nil},

		// 	"" GetModelWholeHandler "GET"
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, NO_POSTERMS, nil},

		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + sysGenId2, http.StatusNoContent, NO_POSTERMS, nil},

		//	"/page" GetModelWholePageHandler "GET"
		// {MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/page?pageNumber=1&pageSize=10", http.StatusOK, NO_POSTERMS, nil},

		//	"" GetModelWholeWithParamHandler "GET"
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?export", http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" GetModelWholeByIdHandler "GET"
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + idCreated1, http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" GetModelWholeByIdWithParamHandler "GET"
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + idCreated1 + "?export", http.StatusOK, NO_POSTERMS, nil},

		// No registered handler
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + idCreated1 + "?unknown", http.StatusOK, NO_POSTERMS, nil},

		//	"/filtered" PostModelWholeFilteredWithParamsHandler "POST"
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=10", http.StatusOK, NO_POSTERMS, nil},

		// No registered handler
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "GET", "?unknown", http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" DeleteModelWholeByIdHandler "DELETE"
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + idCreated1, http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}

func TestPostModelFilteredWithParams(t *testing.T) {
	aut := newModelWholeApiUnitTest(t)
	sysGenId1 := uuid.New().String()
	sysGenId2 := uuid.New().String()
	sysGenId3 := uuid.New().String()
	sysGenId4 := uuid.New().String()

	testCases := []apiUnitTestCase{
		// invalid query params are ignored
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?name=dummy", http.StatusOK, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNum=1", http.StatusOK, NO_POSTERMS, nil},

		// Success
		{MODEL_WHOLE_API, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId1, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_1", aut.modelResponseValidator},
		{MODEL_WHOLE_API, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId2, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_2", aut.modelResponseValidator},
		{MODEL_WHOLE_API, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId3, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_3", aut.modelResponseValidator},
		{MODEL_WHOLE_API, "create_unique_model", "SYSTEM_GENERATED_UNIQUE_IDENTIFIER=" + sysGenId4, aut.replaceKeysByValues, "POST", "", http.StatusCreated, "saveIdIn=model_id_4", aut.modelResponseValidator},
		// Happy Paths
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=4", http.StatusOK, "fetched=4", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=2", http.StatusOK, "fetched=2", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=2&pageSize=3", http.StatusOK, "fetched=1", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=0&pageSize=3", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=-1&pageSize=3", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=4", http.StatusOK, "fetched=4", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=5", http.StatusOK, "fetched=4", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=0", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1&pageSize=-1", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered", http.StatusOK, "fetched=4", aut.modelArrayValidator},

		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=A&pageSize=B", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=A&pageSize=3", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=3&pageSize=B", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},

		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=&pageSize=", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber= &pageSize= ", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=&pageSize= ", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber= &pageSize=", http.StatusBadRequest, "fetched=0", aut.modelArrayValidator},

		// Happy Paths: default value for missing query params
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageNumber=1", http.StatusOK, "fetched=4", aut.modelArrayValidator},
		{MODEL_WHOLE_API, "empty", NO_PRETERMS, nil, "POST", "/filtered?pageSize=3", http.StatusOK, "fetched=3", aut.modelArrayValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("model_id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("model_id_2"), http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("model_id_3"), http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_WHOLE_API, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("model_id_4"), http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}
