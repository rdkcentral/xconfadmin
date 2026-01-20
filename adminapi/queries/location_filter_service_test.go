package queries

import (
	"testing"

	"github.com/rdkcentral/xconfadmin/common"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	shared "github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/stretchr/testify/assert"
)

// helpers
func newLocationFilter(name string) *coreef.DownloadLocationFilter {
	return &coreef.DownloadLocationFilter{Name: name}
}

// reuse global helpers CreateAndSaveModel/CreateAndSaveEnvironment from base test file
func seedEnv(id string) { CreateAndSaveEnvironment(id) }

func TestUpdateLocationFilter_ValidationFailures(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	lfBlank := &coreef.DownloadLocationFilter{}
	assert.Equal(t, 400, UpdateLocationFilter("stb", lfBlank).Status)

	lfAppInvalid := newLocationFilter("LF1")
	assert.Equal(t, 400, UpdateLocationFilter("", lfAppInvalid).Status)
}

func TestUpdateLocationFilter_MissingConditionsBranches(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	// no envs/models/ipgroup -> Condition required
	lf := newLocationFilter("LFCOND")
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf).Status)
	// only models missing envs => Environments required
	lf2 := newLocationFilter("LFCOND2")
	lf2.Models = []string{"M1"}
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf2).Status)
	// only envs missing models => Models required
	lf3 := newLocationFilter("LFCOND3")
	lf3.Environments = []string{"E1"}
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf3).Status)
}

func TestUpdateLocationFilter_ModelEnvExistenceChecks(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	lf := newLocationFilter("LFME")
	lf.Models = []string{"modelx"}
	lf.Environments = []string{"envx"}
	// model doesn't exist => 400
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf).Status)
	// seed model but not env
	seedModel("MODELX")
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf).Status)
	// seed env as well
	seedEnv("ENVX")
	// now fail later due to missing locations (Any location required)
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf).Status)
}

func TestUpdateLocationFilter_LocationValidation(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	seedModel("M1")
	seedEnv("E1")
	lf := newLocationFilter("LFL")
	lf.Models = []string{"m1"}
	lf.Environments = []string{"e1"}
	// ForceHttp true but no HttpLocation -> should 400 (HTTP location required)
	lf.ForceHttp = true
	// ForceHttp true but blank HttpLocation should trigger HTTP location required path
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf).Status)
	// provide HttpLocation but set ipv6 only without ipv4 when not ForceHttp
	lf2 := newLocationFilter("LFL2")
	lf2.Models = []string{"M1"}
	lf2.Environments = []string{"E1"}
	// ipv6 location provided without ipv4 and ForceHttp false
	lf2.Ipv6FirmwareLocation = shared.NewIpAddress("::1")
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf2).Status)
}

func TestUpdateLocationFilter_SuccessAndDelete(t *testing.T) {
	SkipIfMockDatabase(t) // Service test uses ds.GetCachedSimpleDao() directly
	truncateTable(ds.TABLE_FIRMWARE_RULE)

	// Pre-cleanup: remove any models/environments from other tests
	common.DeleteOneModel("M2")
	common.DeleteOneEnvironment("E2")

	// Clean up any existing models and environments to ensure test isolation
	t.Cleanup(func() {
		// Clean up created data
		common.DeleteOneModel("M2")
		common.DeleteOneEnvironment("E2")
	})

	seedModel("M2")
	seedEnv("E2")
	lf := newLocationFilter("LFSUCC")
	lf.Models = []string{"M2"}
	lf.Environments = []string{"E2"}
	lf.HttpLocation = "http://example.com/firmware.bin"
	resp := UpdateLocationFilter("stb", lf)
	if resp.Status != 200 {
		t.Fatalf("expected 200 got %d, error: %v", resp.Status, resp.Error)
	}
	assert.NotEmpty(t, lf.Id)
	// delete existing
	delResp := DeleteLocationFilter("LFSUCC", "stb")
	assert.Equal(t, 204, delResp.Status, "First delete should return 204, error: %v", delResp.Error)
	// delete again (noop)
	delResp2 := DeleteLocationFilter("LFSUCC", "stb")
	assert.Equal(t, 204, delResp2.Status, "Second delete should return 204, error: %v", delResp2.Error)
}

func TestUpdateDownloadLocationRoundRobinFilter(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	// invalid app type
	rr := &coreef.DownloadLocationRoundRobinFilterValue{}
	assert.Equal(t, 400, UpdateDownloadLocationRoundRobinFilter("", rr).Status)
}

func TestUpdateLocationFilter_IpGroupMismatch(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	seedModel("MM1")
	seedEnv("EE1")
	lf := newLocationFilter("LFIPGRP")
	lf.Models = []string{"MM1"}
	lf.Environments = []string{"EE1"}
	// Set IpAddressGroup that is not stored, should trigger mismatch branch
	lf.IpAddressGroup = shared.NewIpAddressGroupWithAddrStrings("NON_EXIST", "NON_EXIST", []string{"10.0.0.10"})
	lf.HttpLocation = "http://example.com/a"
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf).Status)
}

func TestUpdateLocationFilter_FirmwareLocationInvalidVariants(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	seedModel("M3")
	seedEnv("E3")
	// FirmwareLocation IsIpv6 path -> expect Version is invalid
	lf := newLocationFilter("LFIPV6BAD")
	lf.Models = []string{"M3"}
	lf.Environments = []string{"E3"}
	lf.FirmwareLocation = shared.NewIpAddress("::1") // treated as ipv6 => invalid
	lf.HttpLocation = "http://ok"                    // ensure location presence so it reaches branch
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf).Status)

	// FirmwareLocation IsCidrBlock path -> expect IP addresss is invalid
	lf2 := newLocationFilter("LFCIDRBAD")
	lf2.Models = []string{"M3"}
	lf2.Environments = []string{"E3"}
	lf2.FirmwareLocation = shared.NewIpAddress("10.0.0.0/24")
	lf2.HttpLocation = "http://ok"
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf2).Status)
}

func TestUpdateLocationFilter_Ipv6FirmwareLocationInvalidVariants(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	seedModel("M4")
	seedEnv("E4")
	// Ipv6FirmwareLocation IsIpv6 path -> Version is invalid
	lf := newLocationFilter("LFV6BAD")
	lf.Models = []string{"M4"}
	lf.Environments = []string{"E4"}
	lf.HttpLocation = "http://ok"                        // provide http location
	lf.Ipv6FirmwareLocation = shared.NewIpAddress("::1") // triggers IsIpv6()
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf).Status)

	// Ipv6FirmwareLocation IsCidrBlock path -> IP addresss is invalid
	lf2 := newLocationFilter("LFV6CIDR")
	lf2.Models = []string{"M4"}
	lf2.Environments = []string{"E4"}
	lf2.HttpLocation = "http://ok"
	lf2.Ipv6FirmwareLocation = shared.NewIpAddress("2001:db8::/32")
	assert.Equal(t, 400, UpdateLocationFilter("stb", lf2).Status)
}
