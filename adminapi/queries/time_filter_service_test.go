package queries

import (
	"testing"

	"github.com/google/uuid"
	admincoreef "github.com/rdkcentral/xconfadmin/shared/estbfirmware"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	ru "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/stretchr/testify/assert"
)

// helper to seed EnvModelRule prerequisite
func seedEnvModelRule(modelId, envId, appType string) *coreef.EnvModelRuleBean {
	CreateAndSaveModel(modelId)
	CreateAndSaveEnvironment(envId)
	// Build rule with actual env/model conditions so lookup logic can match
	factory := ru.NewRuleFactory()
	envModelRule := factory.NewEnvModelRule(envId, modelId)
	fwRule := corefw.NewEmptyFirmwareRule()
	fwRule.ID = uuid.New().String()
	fwRule.Name = "EM_" + modelId
	fwRule.Type = corefw.ENV_MODEL_RULE
	fwRule.Rule = envModelRule
	fwRule.ApplicationType = appType
	corefw.CreateFirmwareRuleOneDB(fwRule)
	return &coreef.EnvModelRuleBean{Id: fwRule.ID, ModelId: modelId, EnvironmentId: envId, Name: fwRule.Name}
}

func newValidTimeFilter(name string) *coreef.TimeFilter {
	return &coreef.TimeFilter{
		Id:               "",
		Name:             name,
		Start:            "00:00",
		End:              "23:59",
		EnvModelRuleBean: coreef.EnvModelRuleBean{Id: "M1_E1", ModelId: "M1", EnvironmentId: "E1", Name: "EM_M1"},
	}
}

// func TestUpdateTimeFilter_SuccessCreatesAndSetsId(t *testing.T) {
// 	truncateTable(ds.TABLE_FIRMWARE_RULE)
// 	seedEnvModelRule("M1", "E1", "stb")
// 	// seed IP whitelist group so IsChangedIpAddressGroup returns false
// 	ipGrp := shared.NewIpAddressGroupWithAddrStrings("G_OK", "G_OK", []string{"10.0.0.1"})
// 	nl := shared.ConvertFromIpAddressGroup(ipGrp)
// 	ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, nl)
// 	// need RawIpAddresses populated to mirror stored list
// 	ipGrp.RawIpAddresses = []string{"10.0.0.1"}
// 	tf := newValidTimeFilter("TF1")
// 	tf.IpWhiteList = ipGrp
// 	resp := UpdateTimeFilter("stb", tf)
// 	if resp.Status != 200 {
// 		t.Fatalf("expected 200 got %d", resp.Status)
// 	}
// 	assert.NotEmpty(t, tf.Id)
// }

func TestUpdateTimeFilter_ValidationFailures(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	seedEnvModelRule("M1", "E1", "stb")
	cases := []struct {
		name string
		tf   *coreef.TimeFilter
		app  string
		want int
	}{
		{"blank-name", &coreef.TimeFilter{}, "stb", 400},
		{"invalid-app", newValidTimeFilter("T1"), "", 400},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { assert.Equal(t, c.want, UpdateTimeFilter(c.app, c.tf).Status) })
	}
}

func TestUpdateTimeFilter_BadTimes(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	seedEnvModelRule("M1", "E1", "stb")
	tf := newValidTimeFilter("BADTIME")
	tf.Start = "25:00" // invalid hour
	assert.Equal(t, 400, UpdateTimeFilter("stb", tf).Status)
	tf.Start = "00:00"
	tf.End = "99:99"
	assert.Equal(t, 400, UpdateTimeFilter("stb", tf).Status)
}

func TestUpdateTimeFilter_InvalidIpGroup(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	seedEnvModelRule("M1", "E1", "stb")
	grp := shared.NewIpAddressGroupWithAddrStrings("G1", "G1", []string{"10.0.0.1"})
	tf := newValidTimeFilter("TFIP")
	tf.IpWhiteList = grp // group not stored so IsChangedIpAddressGroup -> true
	assert.Equal(t, 400, UpdateTimeFilter("stb", tf).Status)
}

func TestUpdateTimeFilter_EnvModelMissing(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	// no seed for env-model
	tf := newValidTimeFilter("TFMISS")
	// add a valid stored IP group to bypass IsChangedIpAddressGroup and avoid nil deref chain
	ipGrp := shared.NewIpAddressGroupWithAddrStrings("G_TMP", "G_TMP", []string{"10.1.1.1"})
	nl := shared.ConvertFromIpAddressGroup(ipGrp)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, nl)
	ipGrp.RawIpAddresses = []string{"10.1.1.1"}
	tf.IpWhiteList = ipGrp
	assert.Equal(t, 400, UpdateTimeFilter("stb", tf).Status)
}

func TestDeleteTimeFilter_Paths(t *testing.T) {
	truncateTable(ds.TABLE_FIRMWARE_RULE)
	seedEnvModelRule("M1", "E1", "stb")
	tf := newValidTimeFilter("DELTF")
	ipGrp := shared.NewIpAddressGroupWithAddrStrings("G_OK2", "G_OK2", []string{"10.0.0.2"})
	nl := shared.ConvertFromIpAddressGroup(ipGrp)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, nl)
	ipGrp.RawIpAddresses = []string{"10.0.0.2"}
	tf.IpWhiteList = ipGrp
	// directly persist a TIME_FILTER firmware rule to exercise delete paths without relying on UpdateTimeFilter validations
	fr := admincoreef.ConvertTimeFilterToFirmwareRule(tf)
	fr.ApplicationType = "stb"
	if fr.ID == "" { // assign id if not set
		fr.ID = uuid.New().String()
		tf.Id = fr.ID
	}
	corefw.CreateFirmwareRuleOneDB(fr)
	// delete existing
	assert.Equal(t, 204, DeleteTimeFilter("DELTF", "stb").Status)
	// delete non-existing
	assert.Equal(t, 204, DeleteTimeFilter("DELTF", "stb").Status)
}
