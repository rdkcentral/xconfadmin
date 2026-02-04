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
package dcm

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"gotest.tools/assert"
)

// TestGetVodSettingExportHandler_Success tests successful export of VOD settings
func TestGetVodSettingExportHandler_Success(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header is present
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "", "Content-Disposition header should be present")

	// Verify response body is a valid JSON array
	var vodList []*logupload.VodSettings
	err = json.NewDecoder(res.Body).Decode(&vodList)
	assert.NilError(t, err, "Response should be valid JSON")
	assert.Assert(t, vodList != nil, "Response should not be nil")
}

// TestGetVodSettingExportHandler_EmptyResult tests export with no data
func TestGetVodSettingExportHandler_EmptyResult(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header is present
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "", "Content-Disposition header should be present")

	// Verify response is an empty list
	var vodList []*logupload.VodSettings
	err = json.NewDecoder(res.Body).Decode(&vodList)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(vodList), "Should return empty list when no VOD settings exist")
}

// TestGetVodSettingExportHandler_WithDcmFormulas tests export with DCM formulas
func TestGetVodSettingExportHandler_WithDcmFormulas(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create DCM formulas
	formula1 := &logupload.DCMGenericRule{
		ID:              "formula-1",
		Name:            "Formula 1",
		Description:     "Test Formula 1",
		ApplicationType: "stb",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_DCM_RULE, formula1.ID, formula1)

	formula2 := &logupload.DCMGenericRule{
		ID:              "formula-2",
		Name:            "Formula 2",
		Description:     "Test Formula 2",
		ApplicationType: "stb",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_DCM_RULE, formula2.ID, formula2)

	// Create corresponding VOD settings
	vod1 := &logupload.VodSettings{
		ID:              formula1.ID,
		Name:            "VOD 1",
		LocationsURL:    "http://vod1.com",
		ApplicationType: "stb",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_VOD_SETTINGS, vod1.ID, vod1)

	vod2 := &logupload.VodSettings{
		ID:              formula2.ID,
		Name:            "VOD 2",
		LocationsURL:    "http://vod2.com",
		ApplicationType: "stb",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_VOD_SETTINGS, vod2.ID, vod2)

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var vodList []*logupload.VodSettings
	err = json.NewDecoder(res.Body).Decode(&vodList)
	assert.NilError(t, err)
	assert.Equal(t, 2, len(vodList), "Should return 2 VOD settings")
}

// TestGetVodSettingExportHandler_ApplicationTypeFilter tests that export respects application type
func TestGetVodSettingExportHandler_ApplicationTypeFilter(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create DCM formulas with different application types
	formulaSTB := &logupload.DCMGenericRule{
		ID:              "formula-stb",
		Name:            "Formula STB",
		ApplicationType: "stb",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_DCM_RULE, formulaSTB.ID, formulaSTB)

	formulaXHome := &logupload.DCMGenericRule{
		ID:              "formula-xhome",
		Name:            "Formula XHome",
		ApplicationType: "xhome",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_DCM_RULE, formulaXHome.ID, formulaXHome)

	// Create corresponding VOD settings
	vodSTB := &logupload.VodSettings{
		ID:              formulaSTB.ID,
		Name:            "VOD STB",
		LocationsURL:    "http://vodstb.com",
		ApplicationType: "stb",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_VOD_SETTINGS, vodSTB.ID, vodSTB)

	vodXHome := &logupload.VodSettings{
		ID:              formulaXHome.ID,
		Name:            "VOD XHome",
		LocationsURL:    "http://vodxhome.com",
		ApplicationType: "xhome",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_VOD_SETTINGS, vodXHome.ID, vodXHome)

	// Request export for stb only
	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var vodList []*logupload.VodSettings
	err = json.NewDecoder(res.Body).Decode(&vodList)
	assert.NilError(t, err)
	// Should only return STB formula's VOD settings
	assert.Equal(t, 1, len(vodList), "Should return only 1 VOD setting for stb application type")
}

// TestGetVodSettingExportHandler_MissingVodSettings tests formulas without corresponding VOD settings
func TestGetVodSettingExportHandler_MissingVodSettings(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create DCM formula without corresponding VOD settings
	formula := &logupload.DCMGenericRule{
		ID:              "formula-no-vod",
		Name:            "Formula Without VOD",
		ApplicationType: "stb",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_DCM_RULE, formula.ID, formula)

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var vodList []*logupload.VodSettings
	err = json.NewDecoder(res.Body).Decode(&vodList)
	assert.NilError(t, err)
	// Should return list with nil entry for missing VOD settings
	assert.Equal(t, 1, len(vodList), "Should return 1 entry")
	assert.Assert(t, vodList[0] == nil, "Entry should be nil when VOD settings don't exist")
}

// TestGetVodSettingExportHandler_VerifyHeaders tests that export includes correct headers
func TestGetVodSettingExportHandler_VerifyHeaders(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify Content-Disposition header contains expected filename pattern
	contentDisposition := res.Header.Get("Content-Disposition")
	assert.Assert(t, contentDisposition != "", "Content-Disposition header should be present")
	// The filename should contain "allVodSettings_stb"

	// Verify Content-Type is JSON
	contentType := res.Header.Get("Content-Type")
	assert.Assert(t, contentType != "", "Content-Type header should be present")
}

// TestGetVodSettingExportHandler_MissingAuthCookie tests behavior when auth cookie is missing
func TestGetVodSettingExportHandler_MissingAuthCookie(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	// Not adding applicationType cookie - handler will use default/empty value

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	// Handler still returns 200 but with empty application type filter
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify response is still valid JSON array
	var vodList []*logupload.VodSettings
	err = json.NewDecoder(res.Body).Decode(&vodList)
	assert.NilError(t, err)
}

// TestGetVodSettingExportHandler_DifferentApplicationTypes tests export for different application types
func TestGetVodSettingExportHandler_DifferentApplicationTypes(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create formulas for different application types
	apps := []string{"stb", "xhome", "rdkcloud"}
	for i, app := range apps {
		formula := &logupload.DCMGenericRule{
			ID:              "formula-" + app,
			Name:            "Formula " + app,
			ApplicationType: app,
		}
		db.GetCachedSimpleDao().SetOne(db.TABLE_DCM_RULE, formula.ID, formula)

		if i < 2 { // Create VOD settings for first 2 only
			vod := &logupload.VodSettings{
				ID:              formula.ID,
				Name:            "VOD " + app,
				LocationsURL:    "http://vod" + app + ".com",
				ApplicationType: app,
			}
			db.GetCachedSimpleDao().SetOne(db.TABLE_VOD_SETTINGS, vod.ID, vod)
		}
	}

	// Test for stb
	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var vodListSTB []*logupload.VodSettings
	err = json.NewDecoder(res.Body).Decode(&vodListSTB)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(vodListSTB), "Should return 1 VOD setting for stb")

	// Test for xhome
	req2, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req2.Header.Set("Accept", "application/json")
	req2.AddCookie(&http.Cookie{Name: "applicationType", Value: "xhome"})

	res2 := ExecuteRequest(req2, router).Result()
	defer res2.Body.Close()
	assert.Equal(t, http.StatusOK, res2.StatusCode)

	var vodListXHome []*logupload.VodSettings
	err = json.NewDecoder(res2.Body).Decode(&vodListXHome)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(vodListXHome), "Should return 1 VOD setting for xhome")
}

// TestGetVodSettingExportHandler_MultipleFormulasPartialVodSettings tests mixed scenario
func TestGetVodSettingExportHandler_MultipleFormulasPartialVodSettings(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create 3 formulas but only 2 VOD settings
	for i := 1; i <= 3; i++ {
		formula := &logupload.DCMGenericRule{
			ID:              "formula-" + string(rune('0'+i)),
			Name:            "Formula " + string(rune('0'+i)),
			ApplicationType: "stb",
		}
		db.GetCachedSimpleDao().SetOne(db.TABLE_DCM_RULE, formula.ID, formula)

		// Only create VOD settings for formulas 1 and 2
		if i <= 2 {
			vod := &logupload.VodSettings{
				ID:              formula.ID,
				Name:            "VOD " + string(rune('0'+i)),
				LocationsURL:    "http://vod" + string(rune('0'+i)) + ".com",
				ApplicationType: "stb",
			}
			db.GetCachedSimpleDao().SetOne(db.TABLE_VOD_SETTINGS, vod.ID, vod)
		}
	}

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var vodList []*logupload.VodSettings
	err = json.NewDecoder(res.Body).Decode(&vodList)
	assert.NilError(t, err)
	assert.Equal(t, 3, len(vodList), "Should return 3 entries (2 with data, 1 nil)")

	// Count non-nil entries
	nonNilCount := 0
	for _, vod := range vodList {
		if vod != nil {
			nonNilCount++
		}
	}
	assert.Equal(t, 2, nonNilCount, "Should have 2 non-nil VOD settings")
}

// TestGetVodSettingExportHandler_ValidateResponseStructure tests the structure of the response
func TestGetVodSettingExportHandler_ValidateResponseStructure(t *testing.T) {
	SkipIfMockDatabase(t) // Integration test
	DeleteAllEntities()
	defer DeleteAllEntities()

	// Create a complete VOD setting
	formula := &logupload.DCMGenericRule{
		ID:              "complete-formula",
		Name:            "Complete Formula",
		ApplicationType: "stb",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_DCM_RULE, formula.ID, formula)

	vod := &logupload.VodSettings{
		ID:              formula.ID,
		Name:            "Complete VOD",
		LocationsURL:    "http://complete.com",
		ApplicationType: "stb",
		IPNames:         []string{"ip1", "ip2"},
		IPList:          []string{"192.168.1.1", "192.168.1.2"},
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_VOD_SETTINGS, vod.ID, vod)

	req, err := http.NewRequest("GET", "/xconfAdminService/dcm/vodsettings/export", nil)
	assert.NilError(t, err)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})

	res := ExecuteRequest(req, router).Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var vodList []*logupload.VodSettings
	err = json.NewDecoder(res.Body).Decode(&vodList)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(vodList), "Should return 1 VOD setting")

	// Validate the structure
	returnedVod := vodList[0]
	assert.Assert(t, returnedVod != nil, "VOD setting should not be nil")
	assert.Equal(t, "Complete VOD", returnedVod.Name)
	assert.Equal(t, "http://complete.com", returnedVod.LocationsURL)
	assert.Equal(t, "stb", returnedVod.ApplicationType)
	assert.Equal(t, 2, len(returnedVod.IPNames))
	assert.Equal(t, 2, len(returnedVod.IPList))
}
