package http

import (
	"crypto/tls"
	"fmt"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	defaultXconfHost = "http://test.net:8080"
)

type XconfConnector struct {
	*HttpClient
	host             string
	serviceName      string
	xconfUrlTemplate string
}

func NewXconfConnector(conf *configuration.Config, serviceName string, tlsConfig *tls.Config) *XconfConnector {
	confKey := fmt.Sprintf("xconfwebconfig.%v.dataservice_host", serviceName)
	host := conf.GetString(confKey, defaultXconfHost)

	// Read path configuration with default
	telemetryProfilesPath := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.xconfUrlTemplate", serviceName))

	return &XconfConnector{
		HttpClient:       NewHttpClient(conf, serviceName, tlsConfig),
		host:             host,
		serviceName:      serviceName,
		xconfUrlTemplate: telemetryProfilesPath,
	}
}

func (c *XconfConnector) Host() string {
	return c.host
}

func (c *XconfConnector) SetXconfHost(host string) {
	c.host = host
}

func (c *XconfConnector) ServiceName() string {
	return c.serviceName
}

func (c *XconfConnector) GetProfiles(urlSuffix string, fields log.Fields) ([]byte, error) {
	url := fmt.Sprintf(c.xconfUrlTemplate, c.Host(), urlSuffix)
	rbytes, err := c.DoWithRetries("GET", url, nil, nil, fields, c.ServiceName())
	if err != nil {
		return rbytes, err
	}
	return rbytes, nil
}
