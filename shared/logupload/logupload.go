package logupload

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"

	core "github.com/rdkcentral/xconfadmin/shared"
	util "github.com/rdkcentral/xconfadmin/util"

	ds "github.com/rdkcentral/xconfwebconfig/db"

	log "github.com/sirupsen/logrus"
)

// UploadProtocol enum
type UploadProtocol string

const (
	TFTP  UploadProtocol = "TFTP"
	SFTP  UploadProtocol = "SFTP"
	SCP   UploadProtocol = "SCP"
	HTTP  UploadProtocol = "HTTP"
	HTTPS UploadProtocol = "HTTPS"
	S3    UploadProtocol = "S3"
)

var urlRe = regexp.MustCompile(`^[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`)

func IsValidUploadProtocol(p string) bool {
	str := strings.ToUpper(p)
	if str == string(TFTP) || str == string(SFTP) || str == string(SCP) || str == string(HTTP) || str == string(HTTPS) || str == string(S3) {
		return true
	}
	return false
}

func IsValidUrl(str string) bool {
	u, err := url.ParseRequestURI(str)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	if !IsValidUploadProtocol(u.Scheme) {
		return false
	}
	return urlRe.MatchString(u.Host)
}

const (
	EstbIp            string = "estbIP"
	EstbMacAddress    string = "estbMacAddress"
	EcmMac            string = "ecmMacAddress"
	Env               string = "env"
	Model             string = "model"
	AccountMgmt       string = "accountMgmt"
	SerialNum         string = "serialNum"
	PartnerId         string = "partnerId"
	FirmwareVersion   string = "firmwareVersion"
	ControllerId      string = "controllerId"
	ChannelMapId      string = "channelMapId"
	VodId             string = "vodId"
	UploadImmediately string = "uploadImmediately"
	Timezone          string = "timezone"
	AccountHash       string = "accountHash"
	AccountId         string = "accountId"
	ConfigSetHash     string = "configSetHash"
)

/*
	LogUpload tables
*/

// UploadRepository table
type UploadRepository struct {
	ID              string `json:"id"`
	Updated         int64  `json:"updated"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	URL             string `json:"url"`
	ApplicationType string `json:"applicationType"`
	Protocol        string `json:"protocol"`
}

func (obj *UploadRepository) Clone() (*UploadRepository, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*UploadRepository), nil
}

// NewUploadRepositoryInf constructor
func NewUploadRepositoryInf() interface{} {
	return &UploadRepository{
		ApplicationType: core.STB,
	}
}

// LogFile table
type LogFile struct {
	ID             string `json:"id"`
	Updated        int64  `json:"updated"`
	Name           string `json:"name"`
	DeleteOnUpload bool   `json:"deleteOnUpload"`
}

func (obj *LogFile) Clone() (*LogFile, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*LogFile), nil
}

// NewLogFileInf constructor
func NewLogFileInf() interface{} {
	return &LogFile{}
}

func SetLogFile(id string, logFile *LogFile) error {
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_LOG_FILE, id, logFile)
	if err != nil {
		log.Warn("error saving logFile ")
	}
	return err
}

// LogFilesGroups table
type LogFilesGroups struct {
	ID         string   `json:"id"`
	Updated    int64    `json:"updated"`
	GroupName  string   `json:"groupName"`
	LogFileIDs []string `json:"logFileIds"`
}

func (obj *LogFilesGroups) Clone() (*LogFilesGroups, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*LogFilesGroups), nil
}

// NewLogFilesGroupsInf constructor
func NewLogFilesGroupsInf() interface{} {
	return &LogFilesGroups{}
}

func GetLogFileGroupsList(size int) ([]*LogFilesGroups, error) {
	var logFilesGroupsList []*LogFilesGroups
	logFilesGroupsInst, err := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_LOG_FILES_GROUPS, size)
	if err != nil {
		log.Warn("no logFilesGroups found ")
		return nil, err
	}
	for idx := range logFilesGroupsInst {
		logFilesGroups := logFilesGroupsInst[idx].(*LogFilesGroups)
		logFilesGroupsList = append(logFilesGroupsList, logFilesGroups)
	}
	return logFilesGroupsList, nil
}

// LogFileList LogFileList table
type LogFileList struct {
	Updated int64      `json:"updated"`
	Data    []*LogFile `json:"data"`
}

func (obj *LogFileList) Clone() (*LogFileList, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*LogFileList), nil
}

// NewLogFileListInf constructor
func NewLogFileListInf() interface{} {
	return &LogFileList{}
}

type Schedule struct {
	Type              string      `json:"type,omitempty"`
	Expression        string      `json:"expression,omitempty"`
	TimeZone          string      `json:"timeZone,omitempty"`
	ExpressionL1      string      `json:"expressionL1,omitempty"`
	ExpressionL2      string      `json:"expressionL2,omitempty"`
	ExpressionL3      string      `json:"expressionL3,omitempty"`
	StartDate         string      `json:"startDate,omitempty"`
	EndDate           string      `json:"endDate,omitempty"`
	TimeWindowMinutes json.Number `json:"timeWindowMinutes,omitempty"`
}

type ConfigurationServiceURL struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}
