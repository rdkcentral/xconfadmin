package queries

import (
	"testing"

	admincoreef "github.com/rdkcentral/xconfadmin/shared/estbfirmware"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	shared "github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/stretchr/testify/assert"
)

// helper to build wrapper with single env-model percentage entry
func newWrapper(pct float64) *coreef.PercentFilterWrapper {
	w := admincoreef.NewEmptyPercentFilterWrapper()
	w.Percentage = pct
	return w
}

func TestUpdatePercentFilter_AppTypeAndGlobalRangeValidation(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	w := newWrapper(50)
	assert.Equal(t, 400, UpdatePercentFilter("", w).Status)
	w2 := newWrapper(-1)
	assert.Equal(t, 400, UpdatePercentFilter("stb", w2).Status)
	w3 := newWrapper(101)
	assert.Equal(t, 400, UpdatePercentFilter("stb", w3).Status)
}

func TestUpdatePercentFilter_WhitelistMismatch(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	w := newWrapper(10)
	// provide unsaved ip group -> mismatch
	w.Whitelist = shared.NewIpAddressGroupWithAddrStrings("G_BAD", "G_BAD", []string{"10.0.0.1"})
	assert.Equal(t, 400, UpdatePercentFilter("stb", w).Status)
}

func TestUpdatePercentFilter_EnvModelPercentageValidation(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	w := newWrapper(10)
	// FirmwareCheckRequired true but no FirmwareVersions
	w.EnvModelPercentages = append(w.EnvModelPercentages, coreef.EnvModelPercentage{Name: "P1", FirmwareCheckRequired: true, Percentage: 10})
	assert.Equal(t, 400, UpdatePercentFilter("stb", w).Status)

	// LastKnownGood with percentage=100
	w2 := newWrapper(10)
	w2.EnvModelPercentages = append(w2.EnvModelPercentages, coreef.EnvModelPercentage{Name: "P2", FirmwareCheckRequired: true, FirmwareVersions: []string{"v1"}, Percentage: 100, LastKnownGood: "FWV"})
	assert.Equal(t, 400, UpdatePercentFilter("stb", w2).Status)

	// LastKnownGood when inactive
	w3 := newWrapper(10)
	w3.EnvModelPercentages = append(w3.EnvModelPercentages, coreef.EnvModelPercentage{Name: "P3", FirmwareCheckRequired: true, FirmwareVersions: []string{"v1"}, Percentage: 50, Active: false, LastKnownGood: "FWV"})
	assert.Equal(t, 400, UpdatePercentFilter("stb", w3).Status)

	// IntermediateVersion set but FirmwareCheckRequired false
	w4 := newWrapper(10)
	w4.EnvModelPercentages = append(w4.EnvModelPercentages, coreef.EnvModelPercentage{Name: "P4", FirmwareCheckRequired: false, Percentage: 50, IntermediateVersion: "FWV"})
	assert.Equal(t, 400, UpdatePercentFilter("stb", w4).Status)

	// EnvModelPercentage percentage out of range
	w5 := newWrapper(10)
	w5.EnvModelPercentages = append(w5.EnvModelPercentages, coreef.EnvModelPercentage{Name: "P5", FirmwareCheckRequired: true, FirmwareVersions: []string{"v1"}, Percentage: 150})
	assert.Equal(t, 400, UpdatePercentFilter("stb", w5).Status)
}

func TestUpdatePercentFilter_SuccessMinimal(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	w := newWrapper(25)
	resp := UpdatePercentFilter("stb", w)
	if resp.Status != 200 {
		t.Fatalf("expected 200 got %d", resp.Status)
	}
}

func TestGetPercentFilter_NoRules(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	pf, err := GetPercentFilter("stb")
	assert.NoError(t, err)
	assert.NotNil(t, pf)
	// default percentage may differ; just ensure within [0,100]
	assert.GreaterOrEqual(t, float64(pf.Percentage), 0.0)
	assert.LessOrEqual(t, float64(pf.Percentage), 100.0)
}

func TestGetPercentFilterFieldValues_Empty(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	vals, err := GetPercentFilterFieldValues("Percentage", "stb")
	assert.NoError(t, err)
	assert.NotNil(t, vals)
}

func TestUpdatePercentFilter_LastKnownGoodAndIntermediateVersionNotFound(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	w := newWrapper(10)
	// valid env model percentage to pass earlier checks
	w.EnvModelPercentages = append(w.EnvModelPercentages, coreef.EnvModelPercentage{Name: "EM1", FirmwareCheckRequired: true, FirmwareVersions: []string{"1.0"}, Percentage: 50, Active: true, LastKnownGood: "NO_MATCH"})
	assert.Equal(t, 400, UpdatePercentFilter("stb", w).Status)

	w2 := newWrapper(10)
	w2.EnvModelPercentages = append(w2.EnvModelPercentages, coreef.EnvModelPercentage{Name: "EM2", FirmwareCheckRequired: true, FirmwareVersions: []string{"1.0"}, Percentage: 50, Active: true, IntermediateVersion: "NO_MATCH"})
	assert.Equal(t, 400, UpdatePercentFilter("stb", w2).Status)
}

func TestUpdatePercentFilter_WhitelistValidPath(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	// store whitelist
	ipg := shared.NewIpAddressGroupWithAddrStrings("G_OK_PF", "G_OK_PF", []string{"10.10.0.1"})
	nl := shared.ConvertFromIpAddressGroup(ipg)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, nl)
	ipg.RawIpAddresses = []string{"10.10.0.1"}
	w := newWrapper(40)
	w.Whitelist = ipg
	resp := UpdatePercentFilter("stb", w)
	assert.Equal(t, 200, resp.Status)
}

func TestConvertPercentageBean_SumAndWhitelist(t *testing.T) {
	// prepare a namespaced list
	ipg := shared.NewIpAddressGroupWithAddrStrings("G_PCB", "G_PCB", []string{"192.168.0.1"})
	nl := shared.ConvertFromIpAddressGroup(ipg)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, nl)
	// build bean with distributions (include a nil entry to exercise nil-skip) and whitelist id
	bean := &coreef.PercentageBean{
		Whitelist: nl.ID,
		Distributions: []*corefw.ConfigEntry{
			{Percentage: 10},
			nil,
			{Percentage: 15},
		},
	}
	pct := convertPercentageBean(bean)
	assert.NotNil(t, pct)
	assert.Equal(t, float32(25), pct.Percentage)
	assert.NotNil(t, pct.Whitelist)
}

func TestGetPercentFilterValue_ReturnsEmpty(t *testing.T) {
	v := getPercentFilterValue("stb")
	assert.Empty(t, v.EnvModelPercentages)
}
