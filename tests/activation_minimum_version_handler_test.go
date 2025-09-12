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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	//ashttp "xconfas/http"

	"github.com/rdkcentral/xconfadmin/adminapi/queries"

	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	AMV_URL_BASE = "/xconfAdminService/amv%v?%v"

	TEST_MODEL_ID         = "TEST_MODEL_ID"
	TEST_FIRMWARE_VERSION = "TEST_FIRMWARE_VERSION"
	TEST_REGEX            = "test regex"
)

func TestGetAllAmvs(t *testing.T) {
	DeleteAllEntities()
	amv := perCreateActivationVersion(strings.ToUpper(TEST_MODEL_ID), TEST_FIRMWARE_VERSION, TEST_REGEX)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf(AMV_URL_BASE, "", queryParams)

	r := httptest.NewRequest("GET", url, nil)

	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	amvResponse, err := unmarshalActivationVersion(rr.Body.Bytes())
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, []*firmware.ActivationVersion{amv}, amvResponse)
}

func TestGetFilteredAmvHasEmptyRegExFieldIfNoValuesSet(t *testing.T) {
	DeleteAllEntities()
	amv := perCreateActivationVersion(strings.ToUpper(TEST_MODEL_ID), TEST_FIRMWARE_VERSION, "")

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf(AMV_URL_BASE, "/filtered", queryParams)

	r := httptest.NewRequest("GET", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	amvResponse, err := unmarshalActivationVersion(rr.Body.Bytes())
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, []*firmware.ActivationVersion{amv}, amvResponse)
	assert.Equal(t, []string{}, amvResponse[0].RegularExpressions)
}

func TestGetFilteredAmvHasEmptyFirmwareVersionsFieldIfNoValuesSet(t *testing.T) {
	DeleteAllEntities()
	amv := perCreateActivationVersion(strings.ToUpper(TEST_MODEL_ID), "", "test regex")

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf(AMV_URL_BASE, "/filtered", queryParams)

	r := httptest.NewRequest("GET", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	amvResponse, err := unmarshalActivationVersion(rr.Body.Bytes())
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, []*firmware.ActivationVersion{amv}, amvResponse)
	assert.Equal(t, []string{}, amvResponse[0].FirmwareVersions)
}

func perCreateActivationVersion(modelId string, firmwareVersion string, regex string) *firmware.ActivationVersion {
	fc := CreateAndSaveFirmwareConfig(firmwareVersion, modelId, "tftp", "stb")

	amv := firmware.NewActivationVersion()
	amv.ID = uuid.New().String()
	amv.Description = "Test Activation Version"
	amv.ApplicationType = "stb"
	amv.Model = modelId
	if firmwareVersion != "" {
		amv.FirmwareVersions = []string{fc.FirmwareVersion}
	}
	if regex != "" {
		amv.RegularExpressions = []string{regex}
	}
	amv.PartnerId = "TEST_PARTNER_ID"

	queries.CreateAmv(amv, amv.ApplicationType)

	return amv
}

func unmarshalActivationVersion(b []byte) ([]*firmware.ActivationVersion, error) {
	var amvs []*firmware.ActivationVersion
	err := json.Unmarshal(b, &amvs)
	if err != nil {
		return nil, err
	}
	return amvs, nil
}
