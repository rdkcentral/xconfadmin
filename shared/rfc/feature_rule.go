package rfc

import (
	core "github.com/rdkcentral/xconfadmin/shared"
	"github.com/rdkcentral/xconfadmin/util"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	//re "github.com/rdkcentral/xconfwebconfig/rulesengine"
)

// FeatureRule FeatureControlRule2 table
type FeatureRule struct {
	Id              string   `json:"id"`
	Name            string   `json:"name"`
	Rule            *re.Rule `json:"rule"`
	Priority        int      `json:"priority"`
	FeatureIds      []string `json:"featureIds"`
	ApplicationType string   `json:"applicationType"`
}

func (obj *FeatureRule) SetPriority(priority int) {
	obj.Priority = priority
}
func (obj *FeatureRule) GetPriority() int {
	return obj.Priority
}
func (obj *FeatureRule) GetID() string {
	return obj.Id
}

func (obj *FeatureRule) SetApplicationType(appType string) {
	obj.ApplicationType = appType
}

func (obj *FeatureRule) GetApplicationType() string {
	return obj.ApplicationType
}

func (obj *FeatureRule) Clone() (*FeatureRule, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*FeatureRule), nil
}

func NewFeatureRuleInf() interface{} {
	return &FeatureRule{
		ApplicationType: core.STB,
		FeatureIds:      []string{},
	}
}

// GetId XRule interface
func (r *FeatureRule) GetId() string {
	return r.Id
}

// GetRule XRule interface
func (r *FeatureRule) GetRule() *re.Rule {
	return r.Rule
}

// GetName XRule interface
func (r *FeatureRule) GetName() string {
	return r.Name
}

// GetTemplateId XRule interface
func (r *FeatureRule) GetTemplateId() string {
	return ""
}

// GetRuleType XRule interface
func (r *FeatureRule) GetRuleType() string {
	return "FeatureRule"
}
