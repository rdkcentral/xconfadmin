package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"xconfadmin/util"
	"xconfwebconfig/db"
	ds "xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	core "xconfwebconfig/shared"
	shared "xconfwebconfig/shared"

	log "github.com/sirupsen/logrus"
)

// http ok response
type HttpResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// http error response
type HttpErrorResponse struct {
	Status    int         `json:"status"`
	ErrorCode int         `json:"error_code,omitempty"`
	Message   string      `json:"message,omitempty"`
	Errors    interface{} `json:"errors,omitempty"`
}

// http error response to match xconf java admin
type HttpAdminErrorResponse struct {
	Status  int    `json:"status"`
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
}

type Version struct {
	CodeGitCommit   string `json:"code_git_commit"`
	BuildTime       string `json:"build_time"`
	BinaryVersion   string `json:"binary_version"`
	BinaryBranch    string `json:"binary_branch"`
	BinaryBuildTime string `json:"binary_build_time"`
}

type InfoVersion struct {
	ProjectName    string `json:"projectName"`
	ProjectVersion string `json:"projectVersion"`
	ServiceName    string `json:"serviceName"`
	ServiceVersion string `json:"serviceVersion"`
	Source         string `json:"source"`
	Rev            string `json:"rev"`
	GitBranch      string `json:"gitBranch"`
	GitBuildTime   string `json:"gitBuildTime"`
	GitCommitId    string `json:"gitCommitId"`
	GitCommitTime  string `json:"gitCommitTime"`
}

type MacIpRuleConfig struct {
	IpMacIsConditionLimit int `json:"ipMacIsConditionLimit"`
}

func SetAppSetting(key string, value interface{}) (*shared.AppSetting, error) {
	setting := shared.AppSetting{
		ID:      key,
		Updated: util.GetTimestamp(time.Now().UTC()),
		Value:   value,
	}

	err := db.GetCachedSimpleDao().SetOne(db.TABLE_APP_SETTINGS, setting.ID, &setting)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func GetBooleanAppSetting(key string, vargs ...bool) bool {
	defaultVal := false
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := ds.GetCachedSimpleDao().GetOne(TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting := inst.(*shared.AppSetting)
	return setting.Value.(bool)
}

type ResponseEntity struct {
	Status int
	Error  error
	Data   interface{}
}

func NewResponseEntity(err error, data interface{}) *ResponseEntity {
	status := GetXconfErrorStatusCode(err)
	return &ResponseEntity{
		Status: status,
		Error:  err,
		Data:   data,
	}
}

// TODO drop this function when we're done converting from NewResponseEntityWithStatus to NewResponseEntity
func NewResponseEntityWithStatus(status int, err error, data interface{}) *ResponseEntity {
	return &ResponseEntity{
		Status: status,
		Error:  err,
		Data:   data,
	}
}

type ApplicationTypeAware interface {
	GetApplicationType() string
	SetApplicationType(appType string)
}

type EntityMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type GenericNamespacedList struct {
	ID       string          `json:"id"`
	TypeName string          `json:"typeName"`
	Data     map[string]bool `json:"data"`
}

// DcmRule DcmRule table
type DCMGenericRule struct {
	re.Rule
	ID              string      `json:"id"`
	Updated         int64       `json:"updated"`
	Name            string      `json:"name,omitempty"`
	Description     string      `json:"description,omitempty"`
	Priority        int         `json:"priority,omitempty"`
	RuleExpression  string      `json:"ruleExpression,omitempty"`
	Percentage      int         `json:"percentage,omitempty"`
	PercentageL1    json.Number `json:"percentageL1,omitempty"`
	PercentageL2    json.Number `json:"percentageL2,omitempty"`
	PercentageL3    json.Number `json:"percentageL3,omitempty"`
	ApplicationType string      `json:"applicationType"`
}

func (obj *DCMGenericRule) GetPriority() int {
	return obj.Priority
}

func (obj *DCMGenericRule) SetPriority(priority int) {
	obj.Priority = priority
}

func (obj *DCMGenericRule) GetID() string {
	return obj.ID
}

func (obj *DCMGenericRule) Clone() (*DCMGenericRule, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*DCMGenericRule), nil
}

func NewDCMGenericRuleInf() interface{} {
	return &DCMGenericRule{
		Percentage:      100,
		ApplicationType: shared.STB,
	}
}

type DCMFormula struct {
	Formula DCMGenericRule `json:"formula"`
}

// GetId XRule interface
func (r *DCMGenericRule) GetId() string {
	return r.ID
}

// GetRule XRule interface
func (r *DCMGenericRule) GetRule() *re.Rule {
	return &r.Rule
}

// GetName XRule interface
func (r *DCMGenericRule) GetName() string {
	return r.Name
}

// GetTemplateId XRule interface
func (r *DCMGenericRule) GetTemplateId() string {
	return ""
}

// GetRuleType XRule interface
func (r *DCMGenericRule) GetRuleType() string {
	return "DCMGenericRule"
}

func (dcm *DCMGenericRule) ToStringOnlyBaseProperties() string {
	if dcm.Rule.IsCompound() {
		var sb strings.Builder
		for _, compoundPart := range dcm.Rule.CompoundParts {
			sb.WriteString(compoundPart.String())
		}
		return sb.String()
	}
	return dcm.Rule.Condition.String()
}

func GetDCMGenericRuleList() []*DCMGenericRule {
	all := []*DCMGenericRule{}
	dmcRuleList, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_DCM_RULE, 0)
	if err != nil {
		log.Warn("no dmcRule found")
		return all
	}
	for idx := range dmcRuleList {
		if dmcRuleList[idx] != nil {
			dmcRule := dmcRuleList[idx].(*DCMGenericRule)
			all = append(all, dmcRule)
		}
	}
	return all
}

func GetOneDCMGenericRule(id string) *DCMGenericRule {
	dmcRuleInst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_DCM_RULE, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no dmcRule found for " + id))
		return nil
	}
	dmcRule := dmcRuleInst.(*DCMGenericRule)
	return dmcRule
}

func GetAllEnvironmentList() []*shared.Environment {
	result := []*shared.Environment{}
	list, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_ENVIRONMENT, 0)
	if err != nil {
		log.Warn("no environment found")
		return result
	}
	for _, inst := range list {
		env := inst.(*shared.Environment)
		result = append(result, env)
	}
	return result
}

func GetOneEnvironment(id string) *shared.Environment {
	inst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_ENVIRONMENT, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no environment found for " + id))
		return nil
	}
	return inst.(*shared.Environment)
}

func GetAllModelList() []*shared.Model {
	result := []*shared.Model{}
	list, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_MODEL, 0)
	if err != nil {
		log.Warn("no model found")
		return result
	}
	for _, inst := range list {
		model := inst.(*shared.Model)
		result = append(result, model)
	}
	return result
}

func GetOneModel(id string) *shared.Model {
	inst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_MODEL, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no model found for " + id))
		return nil
	}
	return inst.(*shared.Model)
}

func SetOneEnvironment(env *shared.Environment) (*shared.Environment, error) {
	env.Updated = util.GetTimestamp()
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_ENVIRONMENT, env.ID, env)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func DeleteOneEnvironment(id string) error {
	err := ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_ENVIRONMENT, id)
	if err != nil {
		return err
	}
	return nil
}

func SetOneModel(model *core.Model) (*core.Model, error) {
	model.Updated = util.GetTimestamp()
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_MODEL, model.ID, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func DeleteOneModel(id string) error {
	err := ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_MODEL, id)
	if err != nil {
		return err
	}
	return nil
}

func IsExistModel(id string) bool {
	if !util.IsBlank(id) {
		inst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_MODEL, id)
		if inst != nil && err == nil {
			return true
		}
	}
	return false
}

func GetIntAppSetting(key string, vargs ...int) int {
	defaultVal := -1
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting := inst.(*shared.AppSetting)

	// Note: json.Unmarshal numbers into float64 when target type is of type interface{}
	if val, ok := setting.Value.(float64); ok {
		return int(val)
	} else {
		return setting.Value.(int)
	}
}

func GetFloat64AppSetting(key string, vargs ...float64) float64 {
	defaultVal := -1.0
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting := inst.(*shared.AppSetting)
	return setting.Value.(float64)
}

func GetTimeAppSetting(key string, vargs ...time.Time) time.Time {
	var defaultVal time.Time
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting := inst.(*shared.AppSetting)
	timeStr := setting.Value.(string)
	time, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		log.Error(fmt.Sprintf("error getting AppSetting for %s: %s ", key, err.Error()))
	}

	return time
}

func GetStringAppSetting(key string, vargs ...string) string {
	defaultVal := ""
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for " + key))
		return defaultVal
	}

	setting := inst.(*shared.AppSetting)
	return setting.Value.(string)
}

func GetAppSettings() (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	list, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_APP_SETTINGS, 0)
	if err != nil {
		return settings, err
	}
	for _, v := range list {
		p := *v.(*shared.AppSetting)
		settings[p.ID] = p.Value
	}
	return settings, nil
}

// CanarySettings settings for canary deployment
type CanarySettings struct {
	CanaryDistributionPercentage *float64 `json:"distributionPercentage,omitempty"`
	CanaryMaxSize                *int     `json:"maxSize,omitempty"`
	CanaryFwUpgradeStartTime     *int     `json:"firmwareUpgradeStartTime,omitempty"`
	CanaryFwUpgradeEndTime       *int     `json:"firmwareUpgradeEndTime,omitempty"`
}

func (obj *CanarySettings) Validate() error {
	if obj.CanaryMaxSize != nil && *obj.CanaryMaxSize < 1 {
		return errors.New("maxSize must be greater than 0")
	}
	if obj.CanaryMaxSize != nil && *obj.CanaryMaxSize > 100000 {
		return errors.New("maxSize should not be greater than 100k")
	}
	if obj.CanaryDistributionPercentage != nil && (*obj.CanaryDistributionPercentage < 1 || *obj.CanaryDistributionPercentage > 25) {
		return errors.New("distributionPercentage must be in range from 1 to 25")
	}
	if obj.CanaryFwUpgradeStartTime != nil && (*obj.CanaryFwUpgradeStartTime < 0 || *obj.CanaryFwUpgradeStartTime > 5400) {
		return errors.New("firmwareUpgradeStartTime must be in range from 0 to 5400")
	}
	if obj.CanaryFwUpgradeEndTime != nil && (*obj.CanaryFwUpgradeEndTime < 0 || *obj.CanaryFwUpgradeEndTime > 5400) {
		return errors.New("firmwareUpgradeEndTime must be in range from 0 to 5400")
	}
	if obj.CanaryFwUpgradeStartTime != nil && obj.CanaryFwUpgradeEndTime != nil && *obj.CanaryFwUpgradeEndTime <= *obj.CanaryFwUpgradeStartTime {
		return errors.New("firmwareUpgradeEndTime must be greater than firmwareUpgradeStartTime")
	}
	return nil
}

type LockdownSettings struct {
	LockdownEnabled   *bool   `json:"lockdownEnabled,omitempty"`
	LockdownStartTime *string `json:"lockdownStartTime,omitempty"`
	LockdownEndTime   *string `json:"lockdownEndTime,omitempty"`
	LockdownModules   *string `json:"lockdownModules,omitempty"`
}

// recooking_lockdown_settings struct
type RecookingLockdownSettings struct {
	LockdownStartTime *string   `json:"lockdownStartTime,omitempty"`
	Models            *[]string `json:"models,omitempty"`
	Partners          *[]string `json:"partners,omitempty"`
}

func (obj *LockdownSettings) Validate() error {

	if obj.LockdownStartTime != nil && obj.LockdownEndTime == nil {
		return errors.New("LockdownEndTime is required when LockdownStartTime is provided")
	}

	if obj.LockdownEndTime != nil && obj.LockdownStartTime == nil {
		return errors.New("LockdownStartTime is required when LockdownEndTime is provided")
	}

	if obj.LockdownStartTime != nil {
		if err := util.ValidateTimeFormat(*obj.LockdownStartTime); err != nil {
			return err
		}
	}

	if obj.LockdownEndTime != nil {
		if err := util.ValidateTimeFormat(*obj.LockdownEndTime); err != nil {
			return err
		}
	}

	if obj.LockdownEnabled == nil {
		return errors.New("LockdownEnabled is required to be set")
	}

	avaliableModules := []string{"all", "dcm", "rfc", "firmware", "changes", "tools", "common", "telemetry"}

	lockedmodules := strings.Split(strings.ToLower(*obj.LockdownModules), ",")

	for _, module := range lockedmodules {
		if !util.Contains(avaliableModules, module) {
			return errors.New("LockdownModules must be one of: all, dcm, rfc, firmware, changes, tools, common, telemetry")
		}
	}

	if obj.LockdownModules == nil {
		return errors.New("LockdownModules is required to be set")
	}

	return nil
}
