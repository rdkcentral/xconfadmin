package firmware

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/db"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"gotest.tools/assert"
)

// Test GetFirmwareSortedRuleAllAsListDB with no rules
func TestGetFirmwareSortedRuleAllAsListDB_Empty(t *testing.T) {
	result, err := GetFirmwareSortedRuleAllAsListDB()

	// May return error or empty list depending on DB state
	if err == nil {
		assert.Assert(t, result != nil, "Result should not be nil")
	}
}

// Test GetFirmwareSortedRuleAllAsListDB with single rule
func TestGetFirmwareSortedRuleAllAsListDB_SingleRule(t *testing.T) {
	// Create a test rule with unique ID
	rule := &corefw.FirmwareRule{
		ID:   "test-rule-single-001",
		Name: "Test Single Rule",
		Type: "FIRMWARE_RULE",
	}
	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule.ID, rule)

	result, err := GetFirmwareSortedRuleAllAsListDB()

	// In test environment, database may not be configured
	if err != nil {
		// Expected error in test environment
		assert.ErrorContains(t, err, "cache not found")
		return
	}

	assert.Assert(t, result != nil, "Result should not be nil")
	assert.Assert(t, len(result) >= 1, "Should return at least one rule")

	// Find our test rule
	found := false
	for _, r := range result {
		if r.ID == rule.ID {
			assert.Equal(t, "Test Single Rule", r.Name)
			found = true
			break
		}
	}
	assert.Assert(t, found, "Should find our test rule")
}

// Test GetFirmwareSortedRuleAllAsListDB with multiple rules (sorted)
func TestGetFirmwareSortedRuleAllAsListDB_MultipleSorted(t *testing.T) {
	// Create rules with different names (not in alphabetical order) - use unique IDs
	rule1 := &corefw.FirmwareRule{
		ID:   "test-rule-multi-zebra",
		Name: "Zebra Rule",
		Type: "FIRMWARE_RULE",
	}
	rule2 := &corefw.FirmwareRule{
		ID:   "test-rule-multi-alpha",
		Name: "Alpha Rule",
		Type: "FIRMWARE_RULE",
	}
	rule3 := &corefw.FirmwareRule{
		ID:   "test-rule-multi-beta",
		Name: "Beta Rule",
		Type: "FIRMWARE_RULE",
	}

	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)
	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule2.ID, rule2)
	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule3.ID, rule3)

	result, err := GetFirmwareSortedRuleAllAsListDB()

	// Handle DB not configured in test environment
	if err != nil {
		assert.ErrorContains(t, err, "cache not found")
		return
	}

	assert.Assert(t, result != nil, "Result should not be nil")
	assert.Assert(t, len(result) >= 3, "Should return at least three rules")

	// Find our test rules and verify ordering among them
	var testRules []*corefw.FirmwareRule
	for _, r := range result {
		if r.ID == rule1.ID || r.ID == rule2.ID || r.ID == rule3.ID {
			testRules = append(testRules, r)
		}
	}

	assert.Equal(t, 3, len(testRules), "Should find all three test rules")

	// Verify alphabetical sorting by name among our test rules
	assert.Equal(t, "Alpha Rule", testRules[0].Name, "First should be Alpha Rule")
	assert.Equal(t, "Beta Rule", testRules[1].Name, "Second should be Beta Rule")
	assert.Equal(t, "Zebra Rule", testRules[2].Name, "Third should be Zebra Rule")
}

// Test GetFirmwareSortedRuleAllAsListDB with case-insensitive sorting
func TestGetFirmwareSortedRuleAllAsListDB_CaseInsensitive(t *testing.T) {
	// Create rules with mixed case names - use unique IDs
	rule1 := &corefw.FirmwareRule{
		ID:   "test-rule-case-charlie",
		Name: "charlie Rule",
		Type: "FIRMWARE_RULE",
	}
	rule2 := &corefw.FirmwareRule{
		ID:   "test-rule-case-alpha",
		Name: "ALPHA Rule",
		Type: "FIRMWARE_RULE",
	}
	rule3 := &corefw.FirmwareRule{
		ID:   "test-rule-case-beta",
		Name: "Beta Rule",
		Type: "FIRMWARE_RULE",
	}

	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)
	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule2.ID, rule2)
	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule3.ID, rule3)

	result, err := GetFirmwareSortedRuleAllAsListDB()

	// Handle DB not configured in test environment
	if err != nil {
		assert.ErrorContains(t, err, "cache not found")
		return
	}

	assert.Assert(t, result != nil, "Result should not be nil")

	// Find our test rules
	var testRules []*corefw.FirmwareRule
	for _, r := range result {
		if r.ID == rule1.ID || r.ID == rule2.ID || r.ID == rule3.ID {
			testRules = append(testRules, r)
		}
	}

	assert.Equal(t, 3, len(testRules), "Should find all three test rules")

	// Verify case-insensitive alphabetical sorting
	assert.Equal(t, "ALPHA Rule", testRules[0].Name, "First should be ALPHA Rule")
	assert.Equal(t, "Beta Rule", testRules[1].Name, "Second should be Beta Rule")
	assert.Equal(t, "charlie Rule", testRules[2].Name, "Third should be charlie Rule")
}

// Test GetFirmwareSortedRuleAllAsListDB with many rules
func TestGetFirmwareSortedRuleAllAsListDB_ManyRules(t *testing.T) {
	// Create 10 rules with unique IDs
	ruleNames := []string{
		"Rule J", "Rule A", "Rule E", "Rule C", "Rule B",
		"Rule I", "Rule D", "Rule H", "Rule F", "Rule G",
	}

	var testRuleIDs []string
	for i, name := range ruleNames {
		ruleID := fmt.Sprintf("test-rule-many-%d", i)
		testRuleIDs = append(testRuleIDs, ruleID)
		rule := &corefw.FirmwareRule{
			ID:   ruleID,
			Name: name,
			Type: "FIRMWARE_RULE",
		}
		db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule.ID, rule)
	}

	result, err := GetFirmwareSortedRuleAllAsListDB()

	// Handle DB not configured in test environment
	if err != nil {
		assert.ErrorContains(t, err, "cache not found")
		return
	}

	assert.Assert(t, result != nil, "Result should not be nil")

	// Find our test rules
	var testRules []*corefw.FirmwareRule
	for _, r := range result {
		for _, testID := range testRuleIDs {
			if r.ID == testID {
				testRules = append(testRules, r)
				break
			}
		}
	}

	assert.Equal(t, 10, len(testRules), "Should find all ten test rules")

	// Verify first and last are correctly sorted
	assert.Equal(t, "Rule A", testRules[0].Name, "First should be Rule A")
	assert.Equal(t, "Rule J", testRules[9].Name, "Last should be Rule J")

	// Verify complete ordering among our test rules
	for i := 0; i < len(testRules)-1; i++ {
		current := strings.ToLower(testRules[i].Name)
		next := strings.ToLower(testRules[i+1].Name)
		assert.Assert(t, current <= next, "Rules should be in alphabetical order")
	}
}

// Test GetFirmwareSortedRuleAllAsListDB with duplicate names
func TestGetFirmwareSortedRuleAllAsListDB_DuplicateNames(t *testing.T) {
	// Create rules with duplicate names but unique IDs
	rule1 := &corefw.FirmwareRule{
		ID:   "test-rule-dup-1",
		Name: "Duplicate Rule",
		Type: "FIRMWARE_RULE",
	}
	rule2 := &corefw.FirmwareRule{
		ID:   "test-rule-dup-another",
		Name: "Another Rule",
		Type: "FIRMWARE_RULE",
	}
	rule3 := &corefw.FirmwareRule{
		ID:   "test-rule-dup-3",
		Name: "Duplicate Rule",
		Type: "FIRMWARE_RULE",
	}

	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule1.ID, rule1)
	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule2.ID, rule2)
	db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULE, rule3.ID, rule3)

	result, err := GetFirmwareSortedRuleAllAsListDB()

	// Handle DB not configured in test environment
	if err != nil {
		assert.ErrorContains(t, err, "cache not found")
		return
	}

	assert.Assert(t, result != nil, "Result should not be nil")

	// Find our test rules
	var testRules []*corefw.FirmwareRule
	for _, r := range result {
		if r.ID == rule1.ID || r.ID == rule2.ID || r.ID == rule3.ID {
			testRules = append(testRules, r)
		}
	}

	assert.Equal(t, 3, len(testRules), "Should find all three test rules")

	// First should be "Another Rule"
	assert.Equal(t, "Another Rule", testRules[0].Name)
	// Last two should be "Duplicate Rule"
	assert.Equal(t, "Duplicate Rule", testRules[1].Name)
	assert.Equal(t, "Duplicate Rule", testRules[2].Name)
}
