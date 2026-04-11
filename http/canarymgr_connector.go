package http

import (
	"crypto/tls"
	"encoding/json"
	"fmt"

	"github.com/rdkcentral/xconfadmin/common"
	"github.com/rdkcentral/xconfadmin/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	canarymgrServiceName = "canarymgr"
)

type CanaryMgrConnector struct {
	*HttpClient
	host                      string
	createCanaryPath          string
	createWakeupPoolPath      string
	createWakeupPoolGroupPath string
}

type CanaryRequestBody struct {
	Name                   string   `json:"name"`
	DeviceType             string   `json:"deviceType"`
	Size                   int      `json:"size"`
	DistributionPercentage float64  `json:"distributionPercentage"`
	Partner                string   `json:"partner"`
	Model                  string   `json:"model"`
	FwAppliedRule          string   `json:"fwAppliedRule"`
	TimeZones              []string `json:"timeZones"`
	StartPercentRange      float64  `json:"startPercentRange"`
	EndPercentRange        float64  `json:"endPercentRange"`
}

type WakeupPoolDistribution struct {
	ConfigId          string  `json:"configId"`
	StartPercentRange float64 `json:"startPercentRange"`
	EndPercentRange   float64 `json:"endPercentRange"`
}

type WakeupPoolPercentFilter struct {
	Name          string                   `json:"name"`
	DeviceType    string                   `json:"deviceType"`
	Size          int                      `json:"size"`
	Partner       string                   `json:"partner"`
	Model         string                   `json:"model"`
	TimeZones     []string                 `json:"timeZones"`
	Distributions []WakeupPoolDistribution `json:"distributions"`
}

// Define the request body struct
type WakeupPoolRequestBody struct {
	PercentFilters []WakeupPoolPercentFilter `json:"percentFilters"`
}

func NewCanaryMgrConnector(conf *configuration.Config, tlsConfig *tls.Config) *CanaryMgrConnector {
	confKey := fmt.Sprintf("xconfwebconfig.%v.host", canarymgrServiceName)
	host := conf.GetString(confKey)
	if util.IsBlank(host) {
		panic(fmt.Errorf("%s is required", confKey))
	}

	// Read path configurations with defaults
	createCanaryPath := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.createCanaryPath", canarymgrServiceName))
	createWakeupPoolPath := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.createWakeupPoolPath", canarymgrServiceName))
	createWakeupPoolGroupPath := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.createWakeupPoolGroupPath", canarymgrServiceName))

	return &CanaryMgrConnector{
		HttpClient:                NewHttpClient(conf, canarymgrServiceName, tlsConfig),
		host:                      host,
		createCanaryPath:          createCanaryPath,
		createWakeupPoolPath:      createWakeupPoolPath,
		createWakeupPoolGroupPath: createWakeupPoolGroupPath,
	}
}

func (c *CanaryMgrConnector) GetCanaryMgrHost() string {
	return c.host
}

func (c *CanaryMgrConnector) SetCanaryMgrHost(host string) {
	c.host = host
}

func (c *CanaryMgrConnector) GetCanaryPath() string {
	return c.createCanaryPath
}

func (c *CanaryMgrConnector) SetCanaryPath(path string) {
	c.createCanaryPath = path
}

func (c *CanaryMgrConnector) GetWakeupPoolPath() string {
	return c.createWakeupPoolPath
}

func (c *CanaryMgrConnector) SetWakeupPoolPath(path string) {
	c.createWakeupPoolPath = path
}

func (c *CanaryMgrConnector) GetWakeupPoolGroupPath() string {
	return c.createWakeupPoolGroupPath
}

func (c *CanaryMgrConnector) SetWakeupPoolGroupPath(path string) {
	c.createWakeupPoolGroupPath = path
}

func (c *CanaryMgrConnector) CreateCanary(canaryRequestBody *CanaryRequestBody, isDeepSleepVideoDevice bool, fields log.Fields) error {
	pathTemplate := c.createCanaryPath
	if isDeepSleepVideoDevice {
		pathTemplate = c.createWakeupPoolGroupPath
	}
	url := fmt.Sprintf(pathTemplate, c.GetCanaryMgrHost())
	headers := map[string]string{
		common.HeaderUserAgent: common.HeaderXconfAdminService,
	}

	requestBody, err := json.Marshal(canaryRequestBody)
	if err != nil {
		return err
	}

	_, err = c.DoWithRetries("POST", url, headers, []byte(requestBody), fields, canarymgrServiceName)
	if err != nil {
		return err
	}

	return nil
}

func (c *CanaryMgrConnector) CreateWakeupPool(wakeupPoolRequestBody *WakeupPoolRequestBody, force bool, fields log.Fields) error {
	url := fmt.Sprintf(c.createWakeupPoolPath, c.GetCanaryMgrHost(), force)
	headers := map[string]string{
		common.HeaderUserAgent: common.HeaderXconfAdminService,
	}
	requestBody, err := json.Marshal(wakeupPoolRequestBody)
	if err != nil {
		return err
	}
	_, err = c.DoWithRetries("POST", url, headers, []byte(requestBody), fields, canarymgrServiceName)
	if err != nil {
		return err
	}
	return nil
}
