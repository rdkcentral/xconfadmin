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
	"net/http"
	"testing"
)

const (
	jsonFirmwaresTestDataLocn = "jsondata/firmwares/"
)

func newFirmwaresApiUnitTest(t *testing.T) *apiUnitTest {
	aut := newApiUnitTest(t)
	aut.setupFirmwaresApi()
	return aut
}

func (aut *apiUnitTest) setupFirmwaresApi() {
	if aut.getValOf(FWS_QAPI) == "Done" {
		return
	}
	aut.setValOf(FWS_QAPI+DATA_LOCN_SUFFIX, jsonFirmwaresTestDataLocn)
	aut.setValOf(FWS_UAPI+DATA_LOCN_SUFFIX, jsonFirmwaresTestDataLocn)
	aut.setValOf(FWS_DAPI+DATA_LOCN_SUFFIX, jsonFirmwaresTestDataLocn)
	aut.setupModelApi()
	testCases := []apiUnitTestCase{
		{MODEL_UAPI, "FWS_DPC8888", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{MODEL_UAPI, "FWS_DPC8888T", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{MODEL_UAPI, "FWS_DPC9999", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{MODEL_UAPI, "FWS_DPC9999T", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
	}
	aut.run(testCases)
	aut.setValOf(FWS_QAPI, "Done")
}

func (aut *apiUnitTest) cleanupFirmwaresApi() {
	if aut.getValOf(FWS_QAPI) == "" {
		return
	}
	testCases := []apiUnitTestCase{
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/FWS_DPC8888", http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/FWS_DPC8888T", http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/FWS_DPC9999", http.StatusNoContent, NO_POSTERMS, nil},
		{MODEL_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/FWS_DPC9999T", http.StatusNoContent, NO_POSTERMS, nil},
	}
	aut.run(testCases)
	aut.setValOf(FWS_QAPI, "")
}

func TestGetFirmwares(t *testing.T) {
	aut := newFirmwaresApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FWS_UAPI, "firmwares_one", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareConfigResponseValidator},
		{FWS_UAPI, "firmwares_two", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_2", aut.firmwareConfigResponseValidator},
		{FWS_UAPI, "firmwares_three", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_3", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	// Cleanup to make the test idempotent
	testCases = []apiUnitTestCase{
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count+3"), aut.firmwareConfigArrayValidator},
		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_2"), http.StatusNoContent, NO_POSTERMS, nil},
		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_3"), http.StatusNoContent, NO_POSTERMS, nil},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)

}

func TestPostFirmwares(t *testing.T) {
	aut := newFirmwaresApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		// Error cases
		{FWS_UAPI, "missing_application_type", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FWS_UAPI, "missing_description", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FWS_UAPI, "missing_firmware_filename", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FWS_UAPI, "missing_firmware_version", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FWS_UAPI, "model_not_present", NO_PRETERMS, nil, "POST", "", http.StatusBadRequest, NO_POSTERMS, nil},

		//System should generate an ID, if one is not supplied
		{FWS_UAPI, "missing_id", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1&validate=true", aut.firmwareConfigResponseValidator},
		// Create a Firmwares.
		{FWS_UAPI, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_2&validate=true", aut.firmwareConfigResponseValidator},
		// Creating another one with the same id should fail
		{FWS_UAPI, "create", NO_PRETERMS, nil, "POST", "", http.StatusConflict, NO_POSTERMS, nil},
	}
	aut.run(testCases)

	// Cleanup to make the test idempotent
	testCases = []apiUnitTestCase{
		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_2"), http.StatusNoContent, NO_POSTERMS, nil},

		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
}

func TestPutFirmwares(t *testing.T) {
	aut := newFirmwaresApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		// Error cases
		{FWS_UAPI, "missing_application_type", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FWS_UAPI, "missing_description", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FWS_UAPI, "missing_firmware_filename", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FWS_UAPI, "missing_firmware_version", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FWS_UAPI, "model_not_present", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FWS_UAPI, "missing_id", NO_PRETERMS, nil, "PUT", "", http.StatusNotFound, NO_POSTERMS, nil},

		// Create a new Entry
		{FWS_UAPI, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1&validate=true", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	// Update the entry just created, changing only one content at a time
	testCases = []apiUnitTestCase{
		{FWS_UAPI, "create_update_app", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, NO_POSTERMS, nil},
		{FWS_UAPI, "create_update_desc", NO_PRETERMS, nil, "PUT", "", http.StatusOK, "validate=true", aut.firmwareConfigResponseValidator},
		{FWS_UAPI, "create_update_fw_filename", NO_PRETERMS, nil, "PUT", "", http.StatusOK, "validate=true", aut.firmwareConfigResponseValidator},
		{FWS_UAPI, "create_update_fw_version", NO_PRETERMS, nil, "PUT", "", http.StatusOK, "validate=true", aut.firmwareConfigResponseValidator},
		{FWS_UAPI, "create_update_model", NO_PRETERMS, nil, "PUT", "", http.StatusOK, "validate=true", aut.firmwareConfigResponseValidator},
		{FWS_UAPI, "create", NO_PRETERMS, nil, "PUT", "", http.StatusOK, "validate=true", aut.firmwareConfigResponseValidator},
		{FWS_UAPI, "create_partial_update_fw_filename", NO_PRETERMS, nil, "PUT", "", http.StatusBadRequest, "", nil},
	}
	aut.run(testCases)

	// Cleanup to make the test idempotent
	testCases = []apiUnitTestCase{
		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
}

func TestGetFirmwaresById(t *testing.T) {
	aut := newFirmwaresApiUnitTest(t)
	testCases := []apiUnitTestCase{
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/", http.StatusNotFound, "ID=", aut.firmwareConfigSingleValidator},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/firmwares_unit_test_not_exist", http.StatusNotFound, NO_POSTERMS, nil},

		{FWS_UAPI, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
	}
	aut.run(testCases)

	// Cleanup to make the test idempotent
	testCases = []apiUnitTestCase{
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/firmwares_unit_test_1", http.StatusOK, "ID=firmwares_unit_test_1", aut.firmwareConfigSingleValidator},
		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/firmwares_unit_test_1", http.StatusNoContent, NO_POSTERMS, nil},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
}

// func TestDeleteFirmwaresById(t *testing.T) {
// 	aut := newFirmwaresApiUnitTest(t)
// 	percentageBean, err := PreCreatePercentageBean()
// 	assert.NilError(t, err)

// 	testCases := []apiUnitTestCase{
// 		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/" + percentageBean.LastKnownGood, http.StatusOK, "saveIdIn=configId&saveDescIn=configDesc", aut.firmwareConfigResponseValidator},
// 	}
// 	aut.run(testCases)
// 	configId := aut.getValOf("configId")
// 	configDesc := aut.getValOf("configDesc")

// 	testCases = []apiUnitTestCase{
// 		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + configId, http.StatusConflict, "error_message=FirmwareConfig " + configDesc + " is used by " + percentageBean.Name + " rule", aut.ErrorValidator},

// 		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=begin_count", aut.firmwareConfigArrayValidator},

// 		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/", http.StatusNotFound, NO_POSTERMS, nil},
// 		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/firmwares_unit_test_not_exist", http.StatusNotFound, NO_POSTERMS, nil},
// 	}
// 	aut.run(testCases)
// 	testCases = []apiUnitTestCase{
// 		{FWS_UAPI, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
// 		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/firmwares_unit_test_1", http.StatusNoContent, NO_POSTERMS, nil},
// 		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("begin_count"), aut.firmwareConfigArrayValidator},
// 	}
// 	aut.run(testCases)
// }

func TestGetFirmwaresModelByModelId(t *testing.T) {
	aut := newFirmwaresApiUnitTest(t)

	testCases := []apiUnitTestCase{
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "saveFetchedCntIn=countall", aut.firmwareConfigArrayValidator},

		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/FWS_DPC9999T", http.StatusOK, "saveFetchedCntIn=count_9t", aut.firmwareConfigArrayValidator},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/FWS_DPC8888", http.StatusOK, "saveFetchedCntIn=count8", aut.firmwareConfigArrayValidator},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/FWS_DPC8888T", http.StatusOK, "saveFetchedCntIn=count_8t", aut.firmwareConfigArrayValidator},

		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/", http.StatusNotFound, "fetched=0", aut.firmwareConfigArrayValidator},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/non_existant", http.StatusNotFound, "fetched=0", aut.firmwareConfigArrayValidator},

		{FWS_UAPI, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
		{FWS_UAPI, "missing_id", NO_PRETERMS, nil, "POST", "", http.StatusCreated, "saveIdIn=id_1", aut.firmwareConfigResponseValidator},
	}
	aut.run(testCases)

	testCases = []apiUnitTestCase{
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/FWS_DPC9999T", http.StatusOK, "fetched=" + aut.eval("count_9t+1"), aut.firmwareConfigArrayValidator},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/FWS_DPC8888", http.StatusOK, "fetched=" + aut.eval("count8+1"), aut.firmwareConfigArrayValidator},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/FWS_DPC8888T", http.StatusOK, "fetched=" + aut.eval("count_8t+1"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)

	// cleanup to make the test idempotent
	testCases = []apiUnitTestCase{
		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/firmwares_unit_test_1", http.StatusNoContent, NO_POSTERMS, nil},
		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/" + aut.getValOf("id_1"), http.StatusNoContent, NO_POSTERMS, nil},
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, "fetched=" + aut.eval("countall"), aut.firmwareConfigArrayValidator},
	}
	aut.run(testCases)
}

// "/bySupportedModels"
func TestPostFirmwaresBySupportedModels(t *testing.T) {
	aut := newFirmwaresApiUnitTest(t)
	testCases := []apiUnitTestCase{}
	aut.run(testCases)
}

// func TestFirmwaresCRUD(t *testing.T) {
// 	aut := newFirmwaresApiUnitTest(t)
// 	testCases := []apiUnitTestCase{
// 		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/fw_393e2152-9d50-4f30-aab9-c12345678901", http.StatusNotFound, NO_POSTERMS, nil},
// 		{FWS_UAPI, "firmwares_two", NO_PRETERMS, nil, "PUT", "", http.StatusNotFound, NO_POSTERMS, nil},
// 		{FWS_UAPI, "firmwares_two", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},
// 		// Expect still not found because created object has a generated or different ID than the hardcoded one
// 		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/fw_393e2152-9d50-4f30-aab9-c12345678901", http.StatusNotFound, NO_POSTERMS, nil},
// 		// Deleting an ID that was never created should return NotFound throughout
// 		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/fw_393e2152-9d50-4f30-aab9-c12345678901", http.StatusNotFound, NO_POSTERMS, nil},
// 		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/fw_393e2152-9d50-4f30-aab9-c12345678901", http.StatusNotFound, NO_POSTERMS, nil},
// 		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/fw_393e2152-9d50-4f30-aab9-c12345678901", http.StatusNotFound, NO_POSTERMS, nil},
// 	}
// 	aut.run(testCases)
// }

func TestFirmwaresEndPoints(t *testing.T) {
	aut := newFirmwaresApiUnitTest(t)

	testCases := []apiUnitTestCase{
		// 	"" GetFirmwaresHandler "GET"
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "", http.StatusOK, NO_POSTERMS, nil},

		//	"" PostFirmwaresHandler "POST"
		{FWS_UAPI, "create", NO_PRETERMS, nil, "POST", "", http.StatusCreated, NO_POSTERMS, nil},

		//	"" PutFirmwaresHandler "PUT"
		{FWS_UAPI, "create", NO_PRETERMS, nil, "PUT", "", http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" GetFirmwaresByIdHandler "GET"
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/firmwares_unit_test_1", http.StatusOK, NO_POSTERMS, nil},

		//	"/{id}" DeleteFirmwaresByIdHandler "DELETE"
		{FWS_DAPI, NO_INPUT, NO_PRETERMS, nil, "DELETE", "/firmwares_unit_test_1", http.StatusNoContent, NO_POSTERMS, nil},

		//	"/model/{modelId}" GetFirmwaresModelByModelIdHandler "GET"
		{FWS_QAPI, NO_INPUT, NO_PRETERMS, nil, "GET", "/model/dummy_model", http.StatusNotFound, NO_POSTERMS, nil},

		//	"/bySupportedModels" PostFirmwaresBySupportedModelsHandler "POST"
		{FWS_UAPI, NO_INPUT, NO_PRETERMS, nil, "POST", "/bySupportedModels", http.StatusNotFound, NO_POSTERMS, nil},
	}
	aut.run(testCases)
}
