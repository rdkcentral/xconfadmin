package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rdkcentral/xconfadmin/util"

	"github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	shared "github.com/rdkcentral/xconfwebconfig/shared"
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

func SetAppSetting(tenantId string, key string, value interface{}) (*shared.AppSetting, error) {
	setting := shared.AppSetting{
		ID:      key,
		Updated: util.GetTimestamp(),
		Value:   value,
	}

	err := db.GetCachedSimpleDao().SetOne(tenantId, db.TABLE_APP_SETTINGS, setting.ID, &setting)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func GetBooleanAppSetting(tenantId string, key string, vargs ...bool) bool {
	defaultVal := false
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := db.GetCachedSimpleDao().GetOne(tenantId, TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting, ok := inst.(*shared.AppSetting)
	if !ok {
		log.Warn(fmt.Sprintf("AppSetting %s has an unexpected record type; using default %v", key, defaultVal))
		return defaultVal
	}
	return coerceBoolSetting(key, setting.Value, defaultVal)
}

// coerceBoolSetting tolerates the JSON types operators actually send for a
// boolean setting: a real bool, a string like "true"/"false", or the numbers
// 1 and 0. Anything else falls back to the default instead of panicking or
// being ignored.
func coerceBoolSetting(key string, value interface{}, defaultVal bool) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		if parsed, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			return parsed
		}
	case float64:
		// encoding/json decodes every JSON number into float64 when the target
		// is an interface{}, so an operator PUTing 1 or 0 arrives here. Only
		// those two carry a boolean meaning; any other number is a typo rather
		// than an intent to flip the setting.
		if v == 0 || v == 1 {
			return v == 1
		}
	}
	log.Warn(fmt.Sprintf("AppSetting %s has a non-boolean value %v; using default %v", key, value, defaultVal))
	return defaultVal
}

func InitAppSettings(tenantId string) error {
	settings, err := GetAppSettings(tenantId)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"tenantId": tenantId}).Infof("Initializing AppSettings...")

	if _, ok := settings[PROP_LOCKDOWN_ENABLED]; !ok {
		SetAppSetting(tenantId, PROP_LOCKDOWN_ENABLED, false)
	}
	if _, ok := settings[PROP_CANARY_MAXSIZE]; !ok {
		SetAppSetting(tenantId, PROP_CANARY_MAXSIZE, CanarySize)
	}
	if _, ok := settings[PROP_CANARY_DISTRIBUTION_PERCENTAGE]; !ok {
		SetAppSetting(tenantId, PROP_CANARY_DISTRIBUTION_PERCENTAGE, CanaryDistributionPercentage)
	}
	if _, ok := settings[PROP_CANARY_FW_UPGRADE_STARTTIME]; !ok {
		SetAppSetting(tenantId, PROP_CANARY_FW_UPGRADE_STARTTIME, CanaryFwUpgradeStartTime)
	}
	if _, ok := settings[PROP_CANARY_FW_UPGRADE_ENDTIME]; !ok {
		SetAppSetting(tenantId, PROP_CANARY_FW_UPGRADE_ENDTIME, CanaryFwUpgradeEndTime)
	}
	if _, ok := settings[PROP_LOCKDOWN_STARTTIME]; !ok {
		SetAppSetting(tenantId, PROP_LOCKDOWN_STARTTIME, DefaultLockdownStartTime)
	}
	if _, ok := settings[PROP_LOCKDOWN_ENDTIME]; !ok {
		SetAppSetting(tenantId, PROP_LOCKDOWN_ENDTIME, DefaultLockdownEndTime)
	}
	if _, ok := settings[PROP_LOCKDOWN_MODULES]; !ok {
		SetAppSetting(tenantId, PROP_LOCKDOWN_MODULES, DefaultLockdownModules)
	}
	if _, ok := settings[PROP_PRECOOK_LOCKDOWN_ENABLED]; !ok {
		SetAppSetting(tenantId, PROP_PRECOOK_LOCKDOWN_ENABLED, DefaultPrecookLockdownEnabled)
	}
	if _, ok := settings[PROP_CANARY_TIMEZONE_LIST]; !ok {
		SetAppSetting(tenantId, PROP_CANARY_TIMEZONE_LIST, DefaultCanaryTimezone)
	}

	return nil
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

func GetDCMGenericRuleList(tenantId string) []*DCMGenericRule {
	all := []*DCMGenericRule{}
	dmcRuleList, err := db.GetCachedSimpleDao().GetAllAsList(tenantId, db.TABLE_DCM_RULES, 0)
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

func GetOneDCMGenericRule(tenantId string, id string) *DCMGenericRule {
	dmcRuleInst, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_DCM_RULES, id)
	if err != nil {
		log.Warn("no dmcRule found for " + id)
		return nil
	}
	dmcRule := dmcRuleInst.(*DCMGenericRule)
	return dmcRule
}

func GetAllEnvironmentList(tenantId string) []*shared.Environment {
	result := []*shared.Environment{}
	list, err := db.GetCachedSimpleDao().GetAllAsList(tenantId, db.TABLE_ENVIRONMENTS, 0)
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

func GetOneEnvironment(tenantId string, id string) *shared.Environment {
	inst, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_ENVIRONMENTS, id)
	if err != nil {
		log.Warn("no environment found for " + id)
		return nil
	}
	return inst.(*shared.Environment)
}

func GetAllModelList(tenantId string) []*shared.Model {
	result := []*shared.Model{}
	list, err := db.GetCachedSimpleDao().GetAllAsList(tenantId, db.TABLE_MODELS, 0)
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

func GetOneModel(tenantId string, id string) *shared.Model {
	inst, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_MODELS, id)
	if err != nil {
		log.Warn("no model found for " + id)
		return nil
	}
	return inst.(*shared.Model)
}

func SetOneEnvironment(tenantId string, env *shared.Environment) (*shared.Environment, error) {
	env.Updated = util.GetTimestamp()
	err := db.GetCachedSimpleDao().SetOne(tenantId, db.TABLE_ENVIRONMENTS, env.ID, env)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func DeleteOneEnvironment(tenantId string, id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(tenantId, db.TABLE_ENVIRONMENTS, id)
	if err != nil {
		return err
	}
	return nil
}

func SetOneModel(tenantId string, model *core.Model) (*core.Model, error) {
	model.Updated = util.GetTimestamp()
	err := db.GetCachedSimpleDao().SetOne(tenantId, db.TABLE_MODELS, model.ID, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func DeleteOneModel(tenantId string, id string) error {
	err := db.GetCachedSimpleDao().DeleteOne(tenantId, db.TABLE_MODELS, id)
	if err != nil {
		return err
	}
	return nil
}

func IsExistModel(tenantId string, id string) bool {
	if !util.IsBlank(id) {
		inst, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_MODELS, id)
		if inst != nil && err == nil {
			return true
		}
	}
	return false
}

func GetIntAppSetting(tenantId string, key string, vargs ...int) int {
	defaultVal := -1
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_APP_SETTINGS, key)
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

func GetFloat64AppSetting(tenantId string, key string, vargs ...float64) float64 {
	defaultVal := -1.0
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn(fmt.Sprintf("no AppSetting found for %s", key))
		return defaultVal
	}

	setting := inst.(*shared.AppSetting)
	return setting.Value.(float64)
}

func GetTimeAppSetting(tenantId string, key string, vargs ...time.Time) time.Time {
	var defaultVal time.Time
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_APP_SETTINGS, key)
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

func GetStringAppSetting(tenantId string, key string, vargs ...string) string {
	defaultVal := ""
	if len(vargs) > 0 {
		defaultVal = vargs[0]
	}

	inst, err := db.GetCachedSimpleDao().GetOne(tenantId, db.TABLE_APP_SETTINGS, key)
	if err != nil {
		log.Warn("no AppSetting found for " + key)
		return defaultVal
	}

	setting := inst.(*shared.AppSetting)
	return setting.Value.(string)
}

func GetAppSettings(tenantId string) (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	list, err := db.GetCachedSimpleDao().GetAllAsList(tenantId, db.TABLE_APP_SETTINGS, 0)
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

func DeleteTenant(tenantId string) error {
	if tenantId == "" {
		return fmt.Errorf("tenantId cannot be empty")
	}

	dbClient := db.GetDatabaseClient()

	var errs []error
	for _, tableInfo := range db.GetAllTableInfo() {
		// Only delete data for tables that are sharded, i.e. partitioned by tenant ID
		if !tableInfo.Unsharded {
			if err := dbClient.DeleteAllXconfData(tenantId, tableInfo.TableName); err != nil {
				errs = append(errs, fmt.Errorf("failed to delete data for table %s: %v", tableInfo.TableName, err))
			}
		}
	}

	if err := dbClient.DeleteTenant(tenantId); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete tenant %s: %v", tenantId, err))
	}

	db.GetCacheManager().DeleteTenantCache(tenantId)

	err := errors.Join(errs...)
	if err != nil {
		log.WithFields(log.Fields{"tenantId": tenantId}).Errorf("Errors occurred while deleting tenant: %v", err)
	}

	return err
}
