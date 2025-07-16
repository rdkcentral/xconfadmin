package logupload

import (
	"fmt"
	"strings"

	"xconfadmin/common"

	core "xconfadmin/shared"
	util "xconfadmin/util"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	log "github.com/sirupsen/logrus"
)

var SettingTypes = [...]string{"PARTNER_SETTINGS", "EPON", "partnersettings", "epon"}

func IsValidSettingType(str string) bool {
	for _, v := range SettingTypes {
		if v == str {
			return true
		}
	}
	return false
}

// Enum for SettingType
const (
	EPON = iota + 1
	PARTNER_SETTINGS
)

func SettingTypeEnum(s string) int {
	switch strings.ToLower(s) {
	case "epon":
		return EPON
	case "partner_settings", "partnersettings":
		return PARTNER_SETTINGS
	}
	return 0
}

type FormulaWithSettings struct {
	Formula           *common.DCMGenericRule `json:"formula"`
	DeviceSettings    *DeviceSettings        `json:"deviceSettings"`
	LogUpLoadSettings *LogUploadSettings     `json:"logUploadSettings"`
	VodSettings       *VodSettings           `json:"vodSettings"`
}

// SettingProfiles table
type SettingProfiles struct {
	ID               string            `json:"id"`
	Updated          int64             `json:"updated"`
	SettingProfileID string            `json:"settingProfileId"`
	SettingType      string            `json:"settingType"`
	Properties       map[string]string `json:"properties"`
	ApplicationType  string            `json:"applicationType"`
}

func (obj *SettingProfiles) Clone() (*SettingProfiles, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*SettingProfiles), nil
}

// NewSettingProfilesInf constructor
func NewSettingProfilesInf() interface{} {
	return &SettingProfiles{
		ApplicationType: core.STB,
	}
}

func GetOneSettingProfile(id string) *logupload.SettingProfiles {
	inst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_SETTING_PROFILES, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no SettingProfile found for %s", id))
		return nil
	}
	telemetry := inst.(*logupload.SettingProfiles)
	return telemetry
}

// VodSettings table
type VodSettings struct {
	ID              string            `json:"id"`
	Updated         int64             `json:"updated"`
	Name            string            `json:"name"`
	LocationsURL    string            `json:"locationsURL"`
	IPNames         []string          `json:"ipNames"`
	IPList          []string          `json:"ipList"`
	SrmIPList       map[string]string `json:"srmIPList"`
	ApplicationType string            `json:"applicationType"`
}

func (obj *VodSettings) Clone() (*VodSettings, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*VodSettings), nil
}

// NewVodSettingsInf constructor
func NewVodSettingsInf() interface{} {
	return &VodSettings{
		ApplicationType: core.STB,
	}
}

// SettingRule SettingRules table
type SettingRule struct {
	ID              string  `json:"id"`
	Updated         int64   `json:"updated"`
	Name            string  `json:"name"`
	Rule            re.Rule `json:"rule"`
	BoundSettingID  string  `json:"boundSettingId"`
	ApplicationType string  `json:"applicationType"`
}

func (obj *SettingRule) Clone() (*SettingRule, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*SettingRule), nil
}

func (r *SettingRule) GetApplicationType() string {
	if len(r.ApplicationType) > 0 {
		return r.ApplicationType
	}
	return core.STB
}

func GetAllSettingRuleList() []*SettingRule {
	list, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_SETTING_RULES, 0)
	if err != nil {
		log.Warn("no SettingRule found")
		return []*SettingRule{}
	}

	result := make([]*SettingRule, len(list))
	for i, v := range list {
		result[i] = v.(*SettingRule)
	}
	return result
}

// GetId XRule interface
func (r *SettingRule) GetId() string {
	return r.ID
}

// GetRule XRule interface
func (r *SettingRule) GetRule() *re.Rule {
	return &r.Rule
}

// GetName XRule interface
func (r *SettingRule) GetName() string {
	return r.Name
}

// GetTemplateId XRule interface
func (r *SettingRule) GetTemplateId() string {
	return ""
}

// GetRuleType XRule interface
func (r *SettingRule) GetRuleType() string {
	return "SettingRule"
}

// NewSettingRulesInf constructor
func NewSettingRulesInf() interface{} {
	return &SettingRule{
		ApplicationType: core.STB,
	}
}

const (
	DEFAULT_LOG_UPLOAD_SETTINGS_MESSAGE = "Don't upload your logs, but check for updates on this schedule."
)

type Settings struct {
	//java: Set<String> ruleIDs
	RuleIDs                           map[string]string
	SchedulerType                     string
	GroupName                         string
	CheckOnReboot                     bool
	ConfigurationServiceURL           string
	ScheduleCron                      string
	ScheduleDurationMinutes           int
	ScheduleStartDate                 string
	ScheduleEndDate                   string
	LusMessage                        string
	LusName                           string
	LusNumberOfDay                    int
	LusUploadRepositoryName           string
	LusUploadRepositoryURLNew         string
	LusUploadRepositoryUploadProtocol string
	LusUploadRepositoryURL            string
	LusUploadOnReboot                 bool
	UploadImmediately                 bool
	//Upload flag to indicate if allowed to upload logs or not.
	Upload               bool
	LusLogFiles          []*LogFile
	LusLogFilesStartDate string
	LusLogFilesEndDate   string
	//For level one logging
	LusScheduleCron            string
	LusScheduleCronL1          string
	LusScheduleCronL2          string
	LusScheduleCronL3          string
	LusScheduleDurationMinutes int
	LusScheduleStartDate       string
	LusScheduleEndDate         string
	VodSettingsName            string
	LocationUrl                string
	TelemetryProfile           *logupload.PermanentTelemetryProfile
	SrmIPList                  map[string]string
	EponSettings               map[string]string
	PartnerSettings            map[string]string
}

func NewSettings(logFileLenth int) *Settings {
	var newSettings *Settings
	newSettings = new(Settings)
	newSettings.RuleIDs = make(map[string]string)
	newSettings.SrmIPList = make(map[string]string)
	newSettings.EponSettings = make(map[string]string)
	newSettings.PartnerSettings = make(map[string]string)
	newSettings.LusLogFiles = make([]*LogFile, logFileLenth)
	return newSettings
}

func (s *Settings) CopyDeviceSettings(settings *Settings) {
	s.GroupName = settings.GroupName
	s.CheckOnReboot = settings.CheckOnReboot
	s.ConfigurationServiceURL = settings.ConfigurationServiceURL
	s.ScheduleCron = settings.ScheduleCron
	s.ScheduleDurationMinutes = settings.ScheduleDurationMinutes
	s.ScheduleStartDate = settings.ScheduleStartDate
	s.ScheduleEndDate = settings.ScheduleEndDate
}

func (s *Settings) CopyLusSetting(settings *Settings, setLUSSettings bool) {
	if setLUSSettings {
		s.LusMessage = ""
		s.LusName = settings.LusName
		s.LusNumberOfDay = settings.LusNumberOfDay
		s.LusUploadRepositoryName = settings.LusUploadRepositoryName
		s.LusUploadRepositoryURL = settings.LusUploadRepositoryURL
		s.LusUploadRepositoryURLNew = settings.LusUploadRepositoryURLNew
		s.LusUploadRepositoryUploadProtocol = settings.LusUploadRepositoryUploadProtocol
		s.LusUploadOnReboot = settings.LusUploadOnReboot
		s.LusLogFiles = settings.LusLogFiles
		s.LusLogFilesStartDate = settings.LusLogFilesStartDate
		s.LusLogFilesEndDate = settings.LusLogFilesEndDate
		s.LusScheduleDurationMinutes = settings.LusScheduleDurationMinutes
		s.LusScheduleStartDate = settings.LusScheduleStartDate
		s.LusScheduleEndDate = settings.LusScheduleEndDate
		s.Upload = true
	} else {
		s.LusMessage = DEFAULT_LOG_UPLOAD_SETTINGS_MESSAGE
		s.LusName = ""
		s.LusNumberOfDay = 0
		s.LusUploadRepositoryName = ""
		s.LusUploadRepositoryURL = ""
		s.LusUploadRepositoryURLNew = ""
		s.LusUploadRepositoryUploadProtocol = ""
		s.LusUploadOnReboot = false
		s.LusLogFiles = nil
		s.LusLogFilesStartDate = ""
		s.LusLogFilesEndDate = ""
		s.LusScheduleDurationMinutes = 0
		s.LusScheduleStartDate = ""
		s.LusScheduleEndDate = ""
		s.Upload = false
	}
}

func (s *Settings) CopyVodSettings(settings *Settings) {
	s.VodSettingsName = settings.VodSettingsName
	s.LocationUrl = settings.LocationUrl
	s.SrmIPList = settings.SrmIPList
}

func (s *Settings) AreFull() bool {
	if s.GroupName != "" && s.LusName != "" && s.VodSettingsName != "" {
		return true
	}
	return false
}

func (s *Settings) SetSettingProfiles(settingProfiles []SettingProfiles) {
	if len(settingProfiles) < 1 {
		return
	}
	for _, settingProfile := range settingProfiles {
		properties := settingProfile.Properties
		switch SettingTypeEnum(settingProfile.SettingType) {
		case PARTNER_SETTINGS:
			s.PartnerSettings = properties
		case EPON:
			s.EponSettings = properties
		}

	}
}

type SettingsResponse struct {
	GroupName                         interface{}                          `json:"urn:settings:GroupName"`
	CheckOnReboot                     bool                                 `json:"urn:settings:CheckOnReboot"`
	ScheduleCron                      interface{}                          `json:"urn:settings:CheckSchedule:cron"`
	ScheduleDurationMinutes           int                                  `json:"urn:settings:CheckSchedule:DurationMinutes"`
	LusMessage                        interface{}                          `json:"urn:settings:LogUploadSettings:Message"`
	LusName                           interface{}                          `json:"urn:settings:LogUploadSettings:Name"`
	LusNumberOfDay                    int                                  `json:"urn:settings:LogUploadSettings:NumberOfDays"`
	LusUploadRepositoryName           interface{}                          `json:"urn:settings:LogUploadSettings:UploadRepositoryName"`
	LusUploadRepositoryURLNew         string                               `json:"urn:settings:LogUploadSettings:UploadRepository:URL,omitempty"`
	LusUploadRepositoryUploadProtocol string                               `json:"urn:settings:LogUploadSettings:UploadRepository:uploadProtocol,omitempty"`
	LusUploadRepositoryURL            string                               `json:"urn:settings:LogUploadSettings:RepositoryURL,omitempty"`
	LusUploadOnReboot                 bool                                 `json:"urn:settings:LogUploadSettings:UploadOnReboot"`
	UploadImmediately                 bool                                 `json:"urn:settings:LogUploadSettings:UploadImmediately"`
	Upload                            bool                                 `json:"urn:settings:LogUploadSettings:upload"`
	LusScheduleCron                   interface{}                          `json:"urn:settings:LogUploadSettings:UploadSchedule:cron"`
	LusScheduleCronL1                 interface{}                          `json:"urn:settings:LogUploadSettings:UploadSchedule:levelone:cron"`
	LusScheduleCronL2                 interface{}                          `json:"urn:settings:LogUploadSettings:UploadSchedule:leveltwo:cron"`
	LusScheduleCronL3                 interface{}                          `json:"urn:settings:LogUploadSettings:UploadSchedule:levelthree:cron"`
	LusScheduleDurationMinutes        int                                  `json:"urn:settings:LogUploadSettings:UploadSchedule:DurationMinutes"`
	VodSettingsName                   interface{}                          `json:"urn:settings:VODSettings:Name"`
	LocationUrl                       interface{}                          `json:"urn:settings:VODSettings:LocationsURL"`
	SrmIPList                         interface{}                          `json:"urn:settings:VODSettings:SRMIPList"`
	EponSettings                      map[string]string                    `json:"urn:settings:SettingType:epon,omitempty"`
	TelemetryProfile                  *logupload.PermanentTelemetryProfile `json:"urn:settings:TelemetryProfile,omitempty"`
	PartnerSettings                   map[string]string                    `json:"urn:settings:SettingType:partnersettings,omitempty"`
}

func CreateSettingsResponseObject(settings *Settings) *SettingsResponse {
	settingsResponse := &SettingsResponse{
		CheckOnReboot:                     settings.CheckOnReboot,
		ScheduleDurationMinutes:           settings.ScheduleDurationMinutes,
		LusNumberOfDay:                    settings.LusNumberOfDay,
		LusUploadRepositoryURLNew:         settings.LusUploadRepositoryURLNew,
		LusUploadRepositoryUploadProtocol: settings.LusUploadRepositoryUploadProtocol,
		LusUploadRepositoryURL:            settings.LusUploadRepositoryURL,
		LusUploadOnReboot:                 settings.LusUploadOnReboot,
		UploadImmediately:                 settings.UploadImmediately,
		Upload:                            settings.Upload,
		LusScheduleDurationMinutes:        settings.LusScheduleDurationMinutes,
		EponSettings:                      settings.EponSettings,
		TelemetryProfile:                  settings.TelemetryProfile,
		PartnerSettings:                   settings.PartnerSettings,
	}

	if settings.GroupName != "" {
		settingsResponse.GroupName = settings.GroupName
	} else {
		settingsResponse.GroupName = nil
	}
	if settings.ScheduleCron != "" {
		settingsResponse.ScheduleCron = settings.ScheduleCron
	} else {
		settingsResponse.ScheduleCron = nil
	}
	if settings.LusMessage != "" {
		settingsResponse.LusMessage = settings.LusMessage
	} else {
		settingsResponse.LusMessage = nil
	}
	if settings.LusName != "" {
		settingsResponse.LusName = settings.LusName
	} else {
		settingsResponse.LusName = nil
	}
	if settings.LusUploadRepositoryName != "" {
		settingsResponse.LusUploadRepositoryName = settings.LusUploadRepositoryName
	} else {
		settingsResponse.LusUploadRepositoryName = nil
	}
	if settings.LusScheduleCron != "" {
		settingsResponse.LusScheduleCron = settings.LusScheduleCron
	} else {
		settingsResponse.LusScheduleCron = nil
	}
	if settings.LusScheduleCronL1 != "" {
		settingsResponse.LusScheduleCronL1 = settings.LusScheduleCronL1
	} else {
		settingsResponse.LusScheduleCronL1 = nil
	}
	if settings.LusScheduleCronL2 != "" {
		settingsResponse.LusScheduleCronL2 = settings.LusScheduleCronL2
	} else {
		settingsResponse.LusScheduleCronL2 = nil
	}
	if settings.LusScheduleCronL3 != "" {
		settingsResponse.LusScheduleCronL3 = settings.LusScheduleCronL3
	} else {
		settingsResponse.LusScheduleCronL3 = nil
	}
	if settings.VodSettingsName != "" {
		settingsResponse.VodSettingsName = settings.VodSettingsName
	} else {
		settingsResponse.VodSettingsName = nil
	}
	if settings.LocationUrl != "" {
		settingsResponse.LocationUrl = settings.LocationUrl
	} else {
		settingsResponse.LocationUrl = nil
	}
	if len(settings.SrmIPList) > 0 {
		settingsResponse.SrmIPList = settings.SrmIPList
	} else {
		settingsResponse.SrmIPList = nil
	}
	return settingsResponse
}

// DeviceSettings DeviceSettings2 table
type DeviceSettings struct {
	ID                      string                   `json:"id"`
	Updated                 int64                    `json:"updated"`
	Name                    string                   `json:"name"`
	CheckOnReboot           bool                     `json:"checkOnReboot"`
	ConfigurationServiceURL *ConfigurationServiceURL `json:"configurationServiceURL,omitempty"`
	SettingsAreActive       bool                     `json:"settingsAreActive"`
	Schedule                Schedule                 `json:"schedule"`
	ApplicationType         string                   `json:"applicationType"`
}

func (obj *DeviceSettings) Clone() (*DeviceSettings, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*DeviceSettings), nil
}

// NewDeviceSettingsInf constructor
func NewDeviceSettingsInf() interface{} {
	return &DeviceSettings{
		ApplicationType: core.STB,
	}
}

const (
	MODE_TO_GET_LOG_FILES_0 = "LogFiles"
	MODE_TO_GET_LOG_FILES_1 = "LogFilesGroup"
	MODE_TO_GET_LOG_FILES_2 = "AllLogFiles"
)

// LogUploadSettings LogUploadSettings2 table
type LogUploadSettings struct {
	ID                  string   `json:"id"`
	Updated             int64    `json:"updated"`
	Name                string   `json:"name"`
	UploadOnReboot      bool     `json:"uploadOnReboot"`
	NumberOfDays        int      `json:"numberOfDays"`
	AreSettingsActive   bool     `json:"areSettingsActive"`
	Schedule            Schedule `json:"schedule"`
	LogFileIds          []string `json:"logFileIds"`
	LogFilesGroupID     string   `json:"logFilesGroupId"`
	ModeToGetLogFiles   string   `json:"modeToGetLogFiles"`
	UploadRepositoryID  string   `json:"uploadRepositoryId"`
	ActiveDateTimeRange bool     `json:"activeDateTimeRange"`
	FromDateTime        string   `json:"fromDateTime"`
	ToDateTime          string   `json:"toDateTime"`
	ApplicationType     string   `json:"applicationType"`
}

func (obj *LogUploadSettings) Clone() (*LogUploadSettings, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*LogUploadSettings), nil
}

// NewLogUploadSettingsInf constructor
func NewLogUploadSettingsInf() interface{} {
	return &LogUploadSettings{
		ApplicationType: core.STB,
	}
}

func GetOneDeviceSettings(id string) *DeviceSettings {
	var deviceSettings *DeviceSettings
	deviceSettingsInst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_DEVICE_SETTINGS, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no deviceSettings found for Id: %s", id))
		return nil
	}
	deviceSettings = deviceSettingsInst.(*DeviceSettings)
	return deviceSettings
}

func GetOneLogUploadSettings(id string) *LogUploadSettings {
	var logUploadSettings *LogUploadSettings
	logUploadSettingsInst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_LOG_UPLOAD_SETTINGS, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no logUploadSettings found for Id: %s", id))
		return nil
	}
	logUploadSettings = logUploadSettingsInst.(*LogUploadSettings)
	return logUploadSettings
}

func SetOneLogUploadSettings(id string, logUploadSettings *LogUploadSettings) error {
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_LOG_UPLOAD_SETTINGS, id, logUploadSettings)
	if err != nil {
		log.Warn(fmt.Sprintf("error saving logUploadSettings for Id: %s", id))
	}
	return err
}

func GetAllLogUploadSettings(size int) ([]*LogUploadSettings, error) {
	var logUploadSettingsList []*LogUploadSettings
	logUploadSettingsInst, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_LOG_UPLOAD_SETTINGS, size)
	if err != nil {
		log.Warn("error finding logUploadSettings ")
		return nil, err
	}
	for idx := range logUploadSettingsInst {
		logUploadSettings := logUploadSettingsInst[idx].(*LogUploadSettings)
		logUploadSettingsList = append(logUploadSettingsList, logUploadSettings)
	}
	return logUploadSettingsList, err
}

func GetOneUploadRepository(id string) *UploadRepository {
	var uploadRepository *UploadRepository
	uploadRepositoryInst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_UPLOAD_REPOSITORY, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no uploadRepository found for Id: %s", id))
		return nil
	}
	uploadRepository = uploadRepositoryInst.(*UploadRepository)
	return uploadRepository
}

func GetLogFileList(size int) []*LogFile {
	var logFiles []*LogFile
	logFileListInst, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_LOG_FILE, size)
	if err != nil {
		log.Warn("no logFiles found ")
		return nil
	}
	for idx := range logFileListInst {
		logFile := logFileListInst[idx].(*LogFile)
		logFiles = append(logFiles, logFile)
	}
	return logFiles
}

func GetAllLogFileList(size int) []*LogFileList {
	var logFileLists []*LogFileList
	logFileListInst, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_LOG_FILE_LIST, size)
	if err != nil {
		log.Warn("no logFileLists found ")
		return nil
	}
	for idx := range logFileListInst {
		logFileList := logFileListInst[idx].(*LogFileList)
		logFileLists = append(logFileLists, logFileList)
	}
	return logFileLists
}

func GetOneVodSettings(id string) *VodSettings {
	var vodSettings *VodSettings
	vodSettingsInst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_VOD_SETTINGS, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no vodSettings found for Id: %s", id))
		return nil
	}
	vodSettings = vodSettingsInst.(*VodSettings)
	return vodSettings
}

func GetOneLogFileList(id string) (*LogFileList, error) {
	var logFileList *LogFileList
	logFileListInst, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_LOG_FILE_LIST, id)
	if err != nil {
		logFileList = &LogFileList{}
	} else {
		logFileList = logFileListInst.(*LogFileList)
	}
	if logFileList.Data == nil {
		logFileList.Data = []*LogFile{}
	}
	return logFileList, nil
}

func SetOneLogFile(id string, obj *LogFile) error {
	oneList, err := GetOneLogFileList(id)
	for i, logFile := range oneList.Data {
		if logFile.ID == obj.ID {
			oneList.Data = append(oneList.Data[:i], oneList.Data[i+1:]...)
			break
		}
	}
	oneList.Data = append(oneList.Data, obj)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_LOG_FILE_LIST, id, oneList)
	if err != nil {
		log.Warn(fmt.Sprintf("error save logFileList for Id: %s", id))
		return err
	}
	return nil
}

func DeleteOneLogFileList(id string) error {
	err := ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_LOG_FILE_LIST, id)
	return err
}
