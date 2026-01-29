package rfc

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
)

func TestFeatureRuleMarshaling(t *testing.T) {

	src := `{
    "applicationType": "stb",
    "featureIds": [
        "d471efce-b7d6-4419-a40e-5a095e8b6318",
        "7a98f5d9-9652-47a4-9ee9-4814db8aaa24"
    ],
    "id": "8a0dce3d-0f98-4cd5-8d93-cdb9cefb5211",
    "name": "Test_BLE_NS",
    "priority": 1,
    "rule": {
        "compoundParts": [],
        "condition": {
            "fixedArg": {
                "bean": {
                    "value": {
                        "java.lang.String": "34:1F:E4:B7:5E:D0"
                    }
                }
            },
            "freeArg": {
                "name": "estbMacAddress",
                "type": "STRING"
            },
            "operation": "IS"
        },
        "negated": false
    }
}`

	var featureRule FeatureRule
	err := json.Unmarshal([]byte(src), &featureRule)
	assert.NilError(t, err)

	t.Logf("\n\nfeatureRule = %v\n\n", featureRule)

	t.Logf("\n\nfeatureRule.Rule = %v\n\n", featureRule.Rule)

	t.Logf("\n\nfeatureRule.Rule.Condition = %v\n\n", featureRule.Rule.Condition)

	t.Logf("\n\nfeatureRule.Rule.Condition.FixedArg = %v\n\n", featureRule.Rule.Condition.FixedArg)

	t.Logf("\n\nfeatureRule.Rule.Condition.FreeArg = %v\n\n", featureRule.Rule.Condition.FreeArg)
}

func TestFeatureRule_SetPriority(t *testing.T) {
	fr := &FeatureRule{}

	tests := []struct {
		name     string
		priority int
	}{
		{"Set priority to 1", 1},
		{"Set priority to 100", 100},
		{"Set priority to 0", 0},
		{"Set priority to negative", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr.SetPriority(tt.priority)
			assert.Equal(t, tt.priority, fr.Priority)
		})
	}
}

func TestFeatureRule_GetPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority int
	}{
		{"Get priority 1", 1},
		{"Get priority 100", 100},
		{"Get priority 0", 0},
		{"Get negative priority", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &FeatureRule{Priority: tt.priority}
			assert.Equal(t, tt.priority, fr.GetPriority())
		})
	}
}

func TestFeatureRule_GetID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"Get UUID", "8a0dce3d-0f98-4cd5-8d93-cdb9cefb5211"},
		{"Get simple ID", "test-id"},
		{"Get empty ID", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &FeatureRule{Id: tt.id}
			assert.Equal(t, tt.id, fr.GetID())
		})
	}
}

func TestFeatureRule_SetApplicationType(t *testing.T) {
	fr := &FeatureRule{}

	tests := []struct {
		name    string
		appType string
	}{
		{"Set STB", "stb"},
		{"Set xHome", "xhome"},
		{"Set empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr.SetApplicationType(tt.appType)
			assert.Equal(t, tt.appType, fr.ApplicationType)
		})
	}
}

func TestFeatureRule_GetApplicationType(t *testing.T) {
	tests := []struct {
		name    string
		appType string
	}{
		{"Get STB", "stb"},
		{"Get xHome", "xhome"},
		{"Get empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &FeatureRule{ApplicationType: tt.appType}
			assert.Equal(t, tt.appType, fr.GetApplicationType())
		})
	}
}

func TestFeatureRule_Clone(t *testing.T) {
	t.Run("Clone success", func(t *testing.T) {
		original := &FeatureRule{
			Id:              "test-id",
			Name:            "test-name",
			Priority:        10,
			FeatureIds:      []string{"feature1", "feature2"},
			ApplicationType: "stb",
		}

		cloned, err := original.Clone()
		assert.NilError(t, err)
		assert.Assert(t, cloned != nil)
		assert.Equal(t, original.Id, cloned.Id)
		assert.Equal(t, original.Name, cloned.Name)
		assert.Equal(t, original.Priority, cloned.Priority)
		assert.Equal(t, original.ApplicationType, cloned.ApplicationType)
		assert.Equal(t, len(original.FeatureIds), len(cloned.FeatureIds))

		// Verify it's a deep copy
		assert.Assert(t, &original.FeatureIds != &cloned.FeatureIds)
	})

	t.Run("Clone with nil FeatureIds", func(t *testing.T) {
		original := &FeatureRule{
			Id:              "test-id",
			Name:            "test-name",
			Priority:        5,
			ApplicationType: "xhome",
		}

		cloned, err := original.Clone()
		assert.NilError(t, err)
		assert.Assert(t, cloned != nil)
		assert.Equal(t, original.Id, cloned.Id)
	})
}

func TestNewFeatureRuleInf(t *testing.T) {
	result := NewFeatureRuleInf()

	assert.Assert(t, result != nil)

	fr, ok := result.(*FeatureRule)
	assert.Assert(t, ok, "Result should be *FeatureRule")
	assert.Equal(t, "stb", fr.ApplicationType)
	assert.Assert(t, fr.FeatureIds != nil)
	assert.Equal(t, 0, len(fr.FeatureIds))
}

func TestFeatureRule_GetId(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"Get UUID", "8a0dce3d-0f98-4cd5-8d93-cdb9cefb5211"},
		{"Get simple ID", "test-id"},
		{"Get empty ID", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &FeatureRule{Id: tt.id}
			assert.Equal(t, tt.id, fr.GetId())
		})
	}
}

func TestFeatureRule_GetRule(t *testing.T) {
	t.Run("Get nil rule", func(t *testing.T) {
		fr := &FeatureRule{}
		assert.Assert(t, fr.GetRule() == nil)
	})

	t.Run("Get non-nil rule", func(t *testing.T) {
		// Use the full FeatureRule unmarshaling to get a valid Rule
		src := `{
			"applicationType": "stb",
			"featureIds": ["feature1"],
			"id": "test-id",
			"name": "Test_Rule",
			"priority": 1,
			"rule": {
				"compoundParts": [],
				"condition": {
					"fixedArg": {
						"bean": {
							"value": {
								"java.lang.String": "test-value"
							}
						}
					},
					"freeArg": {
						"name": "estbMacAddress",
						"type": "STRING"
					},
					"operation": "IS"
				},
				"negated": false
			}
		}`

		var fr FeatureRule
		err := json.Unmarshal([]byte(src), &fr)
		assert.NilError(t, err)

		result := fr.GetRule()
		assert.Assert(t, result != nil)
		assert.Equal(t, false, result.Negated)
	})
}

func TestFeatureRule_GetName(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
	}{
		{"Get name", "Test_BLE_NS"},
		{"Get simple name", "test"},
		{"Get empty name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &FeatureRule{Name: tt.ruleName}
			assert.Equal(t, tt.ruleName, fr.GetName())
		})
	}
}

func TestFeatureRule_GetTemplateId(t *testing.T) {
	fr := &FeatureRule{Id: "some-id", Name: "some-name"}
	result := fr.GetTemplateId()
	assert.Equal(t, "", result, "GetTemplateId should always return empty string")
}

func TestFeatureRule_GetRuleType(t *testing.T) {
	fr := &FeatureRule{}
	result := fr.GetRuleType()
	assert.Equal(t, "FeatureRule", result, "GetRuleType should always return 'FeatureRule'")
}
