package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	admin_change "github.com/rdkcentral/xconfadmin/shared/change"
	admin_logupload "github.com/rdkcentral/xconfadmin/shared/logupload"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	core_change "github.com/rdkcentral/xconfwebconfig/shared/change"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddTelemetryProfileEntryChangeAndApproveIt(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)

	entry := &logupload.TelemetryElement{uuid.New().String(), "NEW header", "new content", "new type", "10", ""}
	entriesToAdd := []*logupload.TelemetryElement{entry}
	entryByte, _ := json.Marshal(entriesToAdd)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change/entry/add/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(entryByte))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	change := unmarshalChange(rr.Body.Bytes())

	assert.Equal(t, p.ID, change.EntityID)
	assert.Contains(t, change.NewEntity.TelemetryProfile, *entry, "updated profile should contain new telemetry entry")
	assert.NotContains(t, change.OldEntity.TelemetryProfile, *entry, "old profile should not contain new telemetry entry")

	p = logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.NotContains(t, p.TelemetryProfile, *entry, "profile in database should not contain new telemetry entry before approval")

	url = fmt.Sprintf("/xconfAdminService/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	p = logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.Contains(t, p.TelemetryProfile, *entry, "profile in database should contain new telemetry entry after approval")
}

func TestRemoveTelemetryProfileEntryChangeAndApproveIt(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	entry := &logupload.TelemetryElement{uuid.New().String(), "NEW header", "new content", "new type", "10", ""}
	p.TelemetryProfile = append(p.TelemetryProfile, *entry)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)

	entriesToRemove := []*logupload.TelemetryElement{entry}
	entryByte, _ := json.Marshal(entriesToRemove)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change/entry/remove/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(entryByte))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	change := unmarshalChange(rr.Body.Bytes())

	assert.Equal(t, p.ID, change.EntityID)
	assert.NotContains(t, change.NewEntity.TelemetryProfile, *entry, "updated profile should not contain removed telemetry entry")
	assert.Contains(t, change.OldEntity.TelemetryProfile, *entry, "old profile should contain telemetry entry to remove")

	p = logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.Contains(t, p.TelemetryProfile, *entry, "profile in database should contain telemetry entry to remove before approval")

	url = fmt.Sprintf("/xconfAdminService/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	p = logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.NotContains(t, p.TelemetryProfile, *entry, "profile in database should not contain removed telemetry entry after approval")
}

func TestTelemetryProfileCreate(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()

	entryByte, _ := json.Marshal(p)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile?%v", queryParams)

	r := httptest.NewRequest("POST", url, bytes.NewReader(entryByte))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)

	createdProfile := unmarshalProfile(rr.Body.Bytes())

	assert.Equal(t, p, createdProfile)

	dbProfile := logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.Equal(t, p, dbProfile, "profile to create should match created profile in database")
}

func TestTelemetryProfileCreateChangeAndApproveIt(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()

	entryByte, _ := json.Marshal(p)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change?%v", queryParams)

	r := httptest.NewRequest("POST", url, bytes.NewReader(entryByte))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)

	change := unmarshalChange(rr.Body.Bytes())

	assert.Empty(t, change.OldEntity, "old entity in create change should be nil")
	assert.Equal(t, p, change.NewEntity, "new entity should match profile to create")

	dbProfile := logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.Empty(t, dbProfile, "profile before approval should not be present in database")

	url = fmt.Sprintf("/xconfAdminService/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	dbProfile = logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.Equal(t, p, dbProfile, "profile to create should match created profile in database")

	approvedChange := admin_change.GetOneApprovedChange(change.ID)
	assert.NotEmpty(t, approvedChange, "approved telemetry profile change should be created")
	assert.Empty(t, approvedChange.OldEntity, "old entity should not present")
	assert.Equal(t, p, approvedChange.NewEntity, "old entity should not present")
}

func TestTelemetryProfileUpdate(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)

	entry := logupload.TelemetryElement{uuid.New().String(), "newly added header", "newly added content", "newly added type", "10", ""}
	profileToUpdate, _ := p.Clone()
	profileToUpdate.TelemetryProfile = append(profileToUpdate.TelemetryProfile, entry)
	entryByte, _ := json.Marshal(profileToUpdate)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile?%v", queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(entryByte))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	updatedProfile := unmarshalProfile(rr.Body.Bytes())

	assert.Equal(t, profileToUpdate, updatedProfile)

	dbProfile := logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.NotEqual(t, p, dbProfile, "profiles should not match")
	assert.Equal(t, 2, len(dbProfile.TelemetryProfile), "profiles before and after update should not match")
	assert.Contains(t, dbProfile.TelemetryProfile, entry, "profile should contain newly added telemetry entry")

	assert.Equal(t, 0, len(admin_change.GetChangesByEntityId(p.ID)), "no changes should be created")
	assert.Equal(t, 0, len(admin_change.GetApprovedChangeList()), "no approved change should not be created")
}

func TestTelemetryProfileUpdateChangeAndApproveIt(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)

	entry := logupload.TelemetryElement{uuid.New().String(), "newly added header", "newly added content", "newly added type", "10", ""}
	profileToUpdate, _ := p.Clone()
	profileToUpdate.TelemetryProfile = append(profileToUpdate.TelemetryProfile, entry)
	profileBytes, _ := json.Marshal(profileToUpdate)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change?%v", queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(profileBytes))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	change := unmarshalChange(rr.Body.Bytes())

	assert.Equal(t, p, change.OldEntity, "old entity should be equal profile before update")
	assert.Equal(t, profileToUpdate, change.NewEntity, "new entity should match profile to update")

	dbProfile := logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.Equal(t, p, dbProfile, "profile before approval should be equal profile in database")

	url = fmt.Sprintf("/xconfAdminService/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	dbProfile = logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.Equal(t, profileToUpdate, dbProfile, "profile to update should be equal updated profile in database")

	approvedChange := admin_change.GetOneApprovedChange(change.ID)
	assert.NotEmpty(t, approvedChange, "approved telemetry profile change should be created")
	assert.Equal(t, change.ID, approvedChange.ID, "approved change id should be correct")
	assert.Equal(t, p, approvedChange.OldEntity, "old entity should not be present")
	assert.Equal(t, profileToUpdate, approvedChange.NewEntity, "old entity should not be present")
}

func TestTelemetryProfileDelete(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("DELETE", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	ds.GetCachedSimpleDao().RefreshAll(ds.TABLE_PERMANENT_TELEMETRY)
	dbProfile := logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.Empty(t, dbProfile, "telemetry profile should be removed")

	assert.Equal(t, 0, len(admin_change.GetChangesByEntityId(p.ID)), "no changes should be created")
	assert.Equal(t, 0, len(admin_change.GetApprovedChangeList()), "no approved change should not be created")
}

func TestTelemetryProfileDeleteChangeAndApproveIt(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("DELETE", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	change := unmarshalChange(rr.Body.Bytes())

	assert.Equal(t, p, change.OldEntity, "old entity should be equal profile to delete")
	assert.Empty(t, change.NewEntity, "new entity in create change should not exist")

	dbProfile := logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.Equal(t, p, dbProfile, "profile before approval (removing) should be present in database")

	url = fmt.Sprintf("/xconfAdminService/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	ds.GetCachedSimpleDao().RefreshAll(ds.TABLE_PERMANENT_TELEMETRY)

	dbProfile = logupload.GetOnePermanentTelemetryProfile(p.ID)
	assert.Empty(t, dbProfile, "profile should be removed")

	approvedChange := admin_change.GetOneApprovedChange(change.ID)
	assert.NotEmpty(t, approvedChange, "approved telemetry profile change should be created")
	assert.Empty(t, approvedChange.NewEntity, "old entity should not present")
	assert.Equal(t, p, approvedChange.OldEntity, "old entity should be present")
}

func TestTelemetryProfileCreateChangeThrowsExceptionInCaseIfDuplicatedChange(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()

	entryByte, _ := json.Marshal(p)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change?%v", queryParams)

	r := httptest.NewRequest("POST", url, bytes.NewReader(entryByte))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)

	r = httptest.NewRequest("POST", url, bytes.NewReader(entryByte))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusConflict, rr.Code)

	xconfError := unmarshalXconfError(rr.Body.Bytes())
	assert.Equal(t, xconfError.Message, "The same change already exists")
}

func TestTelemetryProfileUpdateChangeThrowsExceptionInCaseIfDuplicatedChange(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)

	entry := logupload.TelemetryElement{uuid.New().String(), "newly added header", "newly added content", "newly added type", "10", ""}
	profileToUpdate, _ := p.Clone()
	profileToUpdate.TelemetryProfile = append(profileToUpdate.TelemetryProfile, entry)
	profileBytes, _ := json.Marshal(profileToUpdate)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change?%v", queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(profileBytes))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	r = httptest.NewRequest("PUT", url, bytes.NewReader(profileBytes))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusConflict, rr.Code)

	xconfError := unmarshalXconfError(rr.Body.Bytes())
	assert.Equal(t, xconfError.Message, "The same change already exists")
}

func TestTelemetryProfileDeleteChangeThrowsExceptionInCaseIfDuplicatedChange(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/profile/change/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("DELETE", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	r = httptest.NewRequest("DELETE", url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusConflict, rr.Code)

	xconfError := unmarshalXconfError(rr.Body.Bytes())
	assert.Equal(t, xconfError.Message, "The same change already exists")
}

func TestUpdateTelemetyProfileThrowsAnExceptionInCaseOfDuplicatedTelemetryEntries(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)

	duplicatedEntry := logupload.TelemetryElement{
		ID:               p.TelemetryProfile[0].ID,
		Header:           p.TelemetryProfile[0].Header,
		Content:          p.TelemetryProfile[0].Content,
		Type:             p.TelemetryProfile[0].Type,
		PollingFrequency: p.TelemetryProfile[0].PollingFrequency,
		Component:        p.TelemetryProfile[0].Component}

	profileToUpdate, _ := p.Clone()
	profileToUpdate.TelemetryProfile = append(profileToUpdate.TelemetryProfile, duplicatedEntry)
	profileBytes, _ := json.Marshal(profileToUpdate)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})

	testEntities := []struct {
		Endpoint    string
		RequestBody []byte
	}{
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/change?%v", queryParams), profileBytes},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile?%v", queryParams), profileBytes},
	}

	for _, testTentity := range testEntities {
		r := httptest.NewRequest("PUT", testTentity.Endpoint, bytes.NewReader(profileBytes))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		xconfError := unmarshalXconfError(rr.Body.Bytes())
		assert.Equal(t, fmt.Sprintf("Profile has duplicated telemetry entry: %v", duplicatedEntry), xconfError.Message)
	}
}

func TestAddTelemetryThrowsAnExceptionInCaseOfDuplicate(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)

	duplicatedEntry := logupload.TelemetryElement{
		ID:               p.TelemetryProfile[0].ID,
		Header:           p.TelemetryProfile[0].Header,
		Content:          p.TelemetryProfile[0].Content,
		Type:             p.TelemetryProfile[0].Type,
		PollingFrequency: p.TelemetryProfile[0].PollingFrequency,
		Component:        p.TelemetryProfile[0].Component}

	telemetryEntriesToAdd, _ := json.Marshal([]*logupload.TelemetryElement{&duplicatedEntry})
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})

	testEntities := []struct {
		Endpoint    string
		RequestBody []byte
	}{
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/change/entry/add/%v?%v", p.ID, queryParams), telemetryEntriesToAdd},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/entry/add/%v?%v", p.ID, queryParams), telemetryEntriesToAdd},
	}

	for _, testTentity := range testEntities {
		r := httptest.NewRequest("PUT", testTentity.Endpoint, bytes.NewReader(testTentity.RequestBody))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusConflict, rr.Code)

		xconfError := unmarshalXconfError(rr.Body.Bytes())
		assert.Equal(t, fmt.Sprintf("Telemetry Profile entry already exists: %v", duplicatedEntry), xconfError.Message)
	}
}

func IgnoreTestApplicationTypeIsMandatory(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryProfile()
	profileBytes, _ := json.Marshal(p)
	entryBytes, _ := json.Marshal(p.TelemetryProfile)

	endpoints := []struct {
		Endpoint       string
		Method         string
		RequestBody    []byte
		ResponseStatus int
	}{
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/{%s}", p.ID), "GET", nil, 400},
		{"/xconfAdminService/telemetry/profile", "POST", profileBytes, 400},
		{"/xconfAdminService/telemetry/profile", "PUT", profileBytes, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/{%s}", p.ID), "DELETE", nil, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/entry/add/{%s}", p.ID), "PUT", entryBytes, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/entry/remove/{%s}", p.ID), "PUT", entryBytes, 400},
		{"/xconfAdminService/telemetry/profile/change", "POST", profileBytes, 400},
		{"/xconfAdminService/telemetry/profile/change", "PUT", profileBytes, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/change/{%s}", p.ID), "DELETE", nil, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/change/entry/add/{%s}", p.ID), "PUT", entryBytes, 400},
		{fmt.Sprintf("/xconfAdminService/telemetry/profile/change/entry/remove/{%s}", p.ID), "PUT", entryBytes, 400},
	}

	for _, entry := range endpoints {
		r := httptest.NewRequest(entry.Method, entry.Endpoint, bytes.NewReader(entry.RequestBody))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, entry.ResponseStatus, rr.Code)

		xconfError := unmarshalXconfError(rr.Body.Bytes())
		assert.Equal(t, xconfError.Message, "ApplicationType is empty")
	}
}

func createTelemetryProfile() *logupload.PermanentTelemetryProfile {
	p := admin_logupload.NewEmptyPermanentTelemetryProfile()
	p.ID = uuid.New().String()
	p.Name = "Test Telemetry Profile"
	p.Schedule = "1 1 1 1 1"
	p.UploadRepository = "http://test.comcast.com"
	p.UploadProtocol = logupload.HTTP
	p.TelemetryProfile = []logupload.TelemetryElement{{uuid.New().String(), "test header", "test content", "str", "10", ""}}
	p.ApplicationType = "stb"
	return p
}

func unmarshalChange(b []byte) core_change.Change {
	var change core_change.Change
	err := json.Unmarshal(b, &change)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling telemetry profile change"))
	}
	return change
}

func unmarshalProfile(b []byte) *logupload.PermanentTelemetryProfile {
	var profile logupload.PermanentTelemetryProfile
	err := json.Unmarshal(b, &profile)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling telemetry profile change"))
	}
	return &profile
}
