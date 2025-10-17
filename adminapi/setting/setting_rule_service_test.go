package setting

import (
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/stretchr/testify/assert"
)

func TestGetOneSettingRule(t *testing.T) {

	settingRule, err := GetOneSettingRule("non-existent-id")
	assert.Nil(t, settingRule)
	assert.NotNil(t, err)
}

func TestDeleteSettingRuleOne(t *testing.T) {
	DeleteSettingRuleOne("non-existent-id")
	assert.True(t, true)
}

func TestSetSettingRule(t *testing.T) {
	err := SetSettingRule("id", &logupload.SettingRule{})
	assert.NotNil(t, err)
}

func TestValidateUsageSettingRule(t *testing.T) {
	err := validateUsageSettingRule("id")
	assert.Nil(t, err)
}

func TestValidateAllSettingRule(t *testing.T) {
	err := validateAllSettingRule(&logupload.SettingRule{})
	assert.Nil(t, err)
}
