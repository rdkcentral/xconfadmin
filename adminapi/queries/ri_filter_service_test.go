package queries

import (
	"testing"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/stretchr/testify/assert"
)

// helper to reset firmware rule table between tests
func resetFirmwareRules() {
	// reuse truncate helper from this package tests if present; otherwise delete directly
	// Firmware rules table name constant resides in ds
	truncateTable(ds.TABLE_FIRMWARE_RULE)
}

func seedModel(id string) {
	CreateAndSaveModel(id)
}

func seedEnvironment(id string) {
	CreateAndSaveEnvironment(id)
}

func seedIpGroup(name string, ips []string) {
	grp := shared.NewIpAddressGroupWithAddrStrings(name, name, ips)
	nl := shared.ConvertFromIpAddressGroup(grp)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, nl)
}

// create minimal valid filter (criteria: one model)
func newValidFilter(name string) *coreef.RebootImmediatelyFilter {
	return &coreef.RebootImmediatelyFilter{
		Id:           "",
		Name:         name,
		Models:       []string{"MODEL1"},
		Environments: []string{"ENV1"},
		MacAddress:   "AA:BB:CC:DD:EE:FF",
	}
}

func TestUpdateRebootImmediatelyFilter_CreateAndUpdatePaths(t *testing.T) {
	resetFirmwareRules()
	seedModel("MODEL1")
	seedEnvironment("ENV1")

	// create (should return 201)
	f := newValidFilter("FILTER_A")
	resp := UpdateRebootImmediatelyFilter("stb", f)
	assert.Equal(t, 201, resp.Status)
	assert.NotEmpty(t, f.Id, "Id should be assigned after save")

	// update same name (should return 200 and keep id)
	originalId := f.Id
	f.MacAddress = "AA:BB:CC:DD:EE:11" // change something to ensure conversion still works
	resp = UpdateRebootImmediatelyFilter("stb", f)
	assert.Equal(t, 200, resp.Status)
	assert.Equal(t, originalId, f.Id)
}

func TestUpdateRebootImmediatelyFilter_ValidationFailures(t *testing.T) {
	resetFirmwareRules()
	seedModel("MODEL1")
	seedEnvironment("ENV1")

	cases := []struct {
		name       string
		filter     *coreef.RebootImmediatelyFilter
		app        string
		wantStatus int
		desc       string
	}{
		{"blank-name", &coreef.RebootImmediatelyFilter{}, "stb", 400, "empty rule name"},
		{"invalid-app", newValidFilter("F1"), "", 400, "invalid application type"},
		{"empty-criteria", &coreef.RebootImmediatelyFilter{Name: "F2"}, "stb", 400, "must have criteria"},
		{"bad-model", &coreef.RebootImmediatelyFilter{Name: "F3", Models: []string{"UNKNOWN"}, Environments: []string{"ENV1"}}, "stb", 400, "model not exist"},
		{"bad-env", &coreef.RebootImmediatelyFilter{Name: "F4", Models: []string{"MODEL1"}, Environments: []string{"BOGUS"}}, "stb", 400, "env not exist"},
		{"bad-mac", &coreef.RebootImmediatelyFilter{Name: "F5", Models: []string{"MODEL1"}, Environments: []string{"ENV1"}, MacAddress: "NOTAMAC"}, "stb", 400, "invalid mac"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := UpdateRebootImmediatelyFilter(tc.app, tc.filter)
			assert.Equal(t, tc.wantStatus, resp.Status, tc.desc)
		})
	}
}

func TestUpdateRebootImmediatelyFilter_IpGroupChanged(t *testing.T) {
	resetFirmwareRules()
	seedModel("MODEL1")
	seedEnvironment("ENV1")
	seedIpGroup("GROUP1", []string{"10.0.0.1"})

	// Provide group with modified content to trigger IsChangedIpAddressGroup true -> 400
	grp := shared.NewIpAddressGroupWithAddrStrings("GROUP1", "GROUP1", []string{"10.0.0.2"})
	f := &coreef.RebootImmediatelyFilter{Name: "FIP", Models: []string{"MODEL1"}, Environments: []string{"ENV1"}, IpAddressGroup: []*shared.IpAddressGroup{grp}}
	resp := UpdateRebootImmediatelyFilter("stb", f)
	assert.Equal(t, 400, resp.Status)
}

func TestDeleteRebootImmediatelyFilter_Paths(t *testing.T) {
	resetFirmwareRules()
	seedModel("MODEL1")
	seedEnvironment("ENV1")
	// create first
	f := newValidFilter("DELME")
	resp := UpdateRebootImmediatelyFilter("stb", f)
	assert.Equal(t, 201, resp.Status)
	// delete existing
	delResp := DeleteRebootImmediatelyFilter("DELME", "stb")
	assert.Equal(t, 204, delResp.Status)
	// delete again (non-existing) should still yield 204
	delResp = DeleteRebootImmediatelyFilter("DELME", "stb")
	assert.Equal(t, 204, delResp.Status)
}

func TestSaveRebootImmediatelyFilter_ErrorPaths(t *testing.T) {
	seedModel("MODEL1")
	seedEnvironment("ENV1")
	// invalid MAC normalization causes ConvertRebootFilterToFirmwareRule to fail
	f := &coreef.RebootImmediatelyFilter{Name: "BAD", Models: []string{"MODEL1"}, Environments: []string{"ENV1"}, MacAddress: "BAD-MAC"}
	_, err := SaveRebootImmediatelyFilter(f, "stb")
	assert.Error(t, err)
}

// Ensure status code when existing rule found sets Id and returns OK
// Redundant with create/update path test; ensure id reuse already covered
// Removed duplicate scenario to simplify suite.

// Additional safety: ensure SaveRebootImmediatelyFilter assigns applicationType
func TestSaveRebootImmediatelyFilter_AssignsAppType(t *testing.T) {
	resetFirmwareRules()
	seedModel("MODEL1")
	seedEnvironment("ENV1")
	f := newValidFilter("APPTYPE")
	fr, err := SaveRebootImmediatelyFilter(f, "stb")
	assert.NoError(t, err)
	assert.Equal(t, "stb", fr.ApplicationType)
}
