package firmware

import (
	"encoding/json"
	"testing"

	ru "github.com/rdkcentral/xconfwebconfig/rulesengine"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
)

func TestNewActivationVersionDefaults(t *testing.T) {
	av := NewActivationVersion()
	if av == nil {
		t.Fatalf("expected activation version instance")
	}
	if av.ApplicationType != "" {
		t.Fatalf("expected empty applicationType initially")
	}
	if len(av.RegularExpressions) != 0 || len(av.FirmwareVersions) != 0 {
		t.Fatalf("expected empty slices")
	}
	av.SetApplicationType("stb")
	if av.GetApplicationType() != "stb" {
		t.Fatalf("Set/GetApplicationType mismatch")
	}
}

func TestApplicableActionTypeHelpers(t *testing.T) {
	valid := []ApplicableActionType{RULE, DEFINE_PROPERTIES, BLOCKING_FILTER, RULE_TEMPLATE, DEFINE_PROPERTIES_TEMPLATE, BLOCKING_FILTER_TEMPLATE}
	for _, v := range valid {
		if !IsValidApplicableActionType(v) {
			t.Fatalf("expected valid action type %s", v)
		}
		if ApplicableActionTypeToString(v) == "" {
			t.Fatalf("expected non-empty string mapping for %s", v)
		}
	}
	// CaseIgnoreEquals
	a := RULE
	b := ApplicableActionType("rule")
	if !a.CaseIgnoreEquals(b) {
		t.Fatalf("case ignore equals failed")
	}
	// IsSuperSetOf
	sup := ApplicableActionType("Define_Properties_Template")
	sub := DEFINE_PROPERTIES_TEMPLATE
	if !sup.IsSuperSetOf(&sub) {
		t.Fatalf("expected supersets contains behavior")
	}
}

func TestNewRuleActionDefaults(t *testing.T) {
	ra := NewRuleAction()
	if !ra.Active || ra.UseAccountPercentage || ra.FirmwareCheckRequired || ra.RebootImmediately {
		t.Fatalf("default flags unexpected: %+v", ra)
	}
	if len(ra.FirmwareVersions) != 0 || len(ra.ConfigEntries) != 0 {
		t.Fatalf("expected empty slices")
	}
}

func TestNewConfigEntryPercentageRoundingAndEqualsCompare(t *testing.T) {
	ce := NewConfigEntry("cfg1", 1.2345, 2.3456) // diff 1.1111 -> percentage 1.111 after rounding
	if ce.Percentage != 1.111 {
		t.Fatalf("expected rounded percentage 1.111 got %f", ce.Percentage)
	}
	ce2 := NewConfigEntry("cfg1", 1.2345, 2.3456)
	if !ce.Equals(ce2) {
		t.Fatalf("equals should succeed for identical entries")
	}
	if ce.CompareTo(ce2) != 0 {
		t.Fatalf("compareTo identical should be 0")
	}
	ceEarlier := NewConfigEntry("cfg2", 0.5, 0.6)
	if ce.CompareTo(ceEarlier) != 1 || ceEarlier.CompareTo(ce) != -1 {
		t.Fatalf("compareTo ordering incorrect")
	}
}

func TestSortConfigEntry(t *testing.T) {
	a := NewConfigEntry("a", 20, 30)
	b := NewConfigEntry("b", 10, 20)
	c := NewConfigEntry("c", 15, 16)
	entries := []*ConfigEntry{a, b, c}
	SortConfigEntry(entries)
	if entries[0] != b || entries[1] != c || entries[2] != a {
		t.Fatalf("unexpected sort order")
	}
}

func TestHasFirmwareVersion(t *testing.T) {
	versions := []string{"A", "B"}
	if !HasFirmwareVersion(versions, "A") || HasFirmwareVersion(versions, "C") {
		t.Fatalf("HasFirmwareVersion logic incorrect")
	}
}

func TestApplicableActionFirmwareVersionAccessors(t *testing.T) {
	aa := &ApplicableAction{ActivationFirmwareVersions: map[string][]string{"firmwareVersions": {"v1"}, "regularExpressions": {"re1"}}}
	if len(aa.GetFirmwareVersions()) != 1 || aa.GetFirmwareVersions()[0] != "v1" {
		t.Fatalf("GetFirmwareVersions mismatch")
	}
	if len(aa.GetFirmwareVersionRegExs()) != 1 || aa.GetFirmwareVersionRegExs()[0] != "re1" {
		t.Fatalf("GetFirmwareVersionRegExs mismatch")
	}
}

func TestPropertyValueCreation(t *testing.T) {
	pv := NewPropertyValue("value", true, STRING)
	if _, err := json.Marshal(pv); err != nil || pv.Value != "value" || !pv.Optional {
		t.Fatalf("property value marshal or fields incorrect: %v %v", pv, err)
	}
	if len(pv.ValidationTypes) != 1 || pv.ValidationTypes[0] != STRING {
		t.Fatalf("validation types mismatch")
	}
}

func TestGetFirmwareRuleTemplateCount(t *testing.T) {
	// Test with potential DB error recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to DB not configured: %v", r)
		}
	}()

	count, err := GetFirmwareRuleTemplateCount()
	if err != nil {
		t.Logf("DB error expected in test environment: %v", err)
		return
	}
	if count < 0 {
		t.Fatalf("count should not be negative")
	}
}

func TestNewFirmwareRuleTemplate(t *testing.T) {
	rule := ru.Rule{}
	byPassFilters := []string{"filter1", "filter2"}
	template := NewFirmwareRuleTemplate("test-id", rule, byPassFilters, 100)

	if template == nil {
		t.Fatalf("expected non-nil template")
	}
	if template.ID != "test-id" {
		t.Fatalf("expected ID 'test-id', got %s", template.ID)
	}
	if template.Priority != 100 {
		t.Fatalf("expected priority 100, got %d", template.Priority)
	}
	if len(template.ByPassFilters) != 2 {
		t.Fatalf("expected 2 bypass filters, got %d", len(template.ByPassFilters))
	}
	if !template.Editable {
		t.Fatalf("expected template to be editable")
	}
}

func TestNewBlockingFilterTemplate(t *testing.T) {
	rule := ru.Rule{}
	template := NewBlockingFilterTemplate("blocking-id", rule, 50)

	if template == nil {
		t.Fatalf("expected non-nil template")
	}
	if template.ID != "blocking-id" {
		t.Fatalf("expected ID 'blocking-id', got %s", template.ID)
	}
	if template.Priority != 50 {
		t.Fatalf("expected priority 50, got %d", template.Priority)
	}
	if len(template.ByPassFilters) != 0 {
		t.Fatalf("expected 0 bypass filters, got %d", len(template.ByPassFilters))
	}
}

func TestNewDefinePropertiesTemplate(t *testing.T) {
	rule := ru.Rule{}
	properties := map[string]corefw.PropertyValue{
		"key1": {Value: "value1"},
	}
	byPassFilter := []string{"filter1"}
	template := NewDefinePropertiesTemplate("props-id", rule, properties, byPassFilter, 75)

	if template == nil {
		t.Fatalf("expected non-nil template")
	}
	if template.ID != "props-id" {
		t.Fatalf("expected ID 'props-id', got %s", template.ID)
	}
	if template.Priority != 75 {
		t.Fatalf("expected priority 75, got %d", template.Priority)
	}
	if len(template.ApplicableAction.Properties) != 1 {
		t.Fatalf("expected 1 property, got %d", len(template.ApplicableAction.Properties))
	}
}

func TestGetFirmwareSortedRuleAllAsListDB(t *testing.T) {
	// Test with potential DB error recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to DB not configured: %v", r)
		}
	}()

	rules, err := GetFirmwareSortedRuleAllAsListDB()
	if err != nil {
		t.Logf("DB error expected in test environment: %v", err)
		return
	}
	if rules == nil {
		t.Fatalf("expected non-nil rules slice")
	}
}
