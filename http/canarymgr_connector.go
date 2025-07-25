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
	createCanaryPath     = "%s/api/v1/canarygroup"
)

type CanaryMgrConnector struct {
	*HttpClient
	host string
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

func NewCanaryMgrConnector(conf *configuration.Config, tlsConfig *tls.Config) *CanaryMgrConnector {
	confKey := fmt.Sprintf("xconfwebconfig.%v.host", canarymgrServiceName)
	host := conf.GetString(confKey)
	if util.IsBlank(host) {
		panic(fmt.Errorf("%s is required", confKey))
	}

	return &CanaryMgrConnector{
		HttpClient: NewHttpClient(conf, canarymgrServiceName, tlsConfig),
		host:       host,
	}
}

func (c *CanaryMgrConnector) GetCanaryMgrHost() string {
	return c.host
}

func (c *CanaryMgrConnector) SetCanaryMgrHost(host string) {
	c.host = host
}

func (c *CanaryMgrConnector) CreateCanary(canaryRequestBody *CanaryRequestBody, fields log.Fields) error {
	url := fmt.Sprintf(createCanaryPath, c.GetCanaryMgrHost())
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
