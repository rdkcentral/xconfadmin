package http

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rdkcentral/xconfadmin/common"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	xcrpServiceName = "xcrp"
)

type XcrpConnector struct {
	*HttpClient
	hosts                      []string
	recookPathTemplate         string
	precookPathTemplate        string
	precookModelPathTemplate   string
	precookPartnerPathTemplate string
	precookStatusPathTemplate  string
}

func NewXcrpConnector(conf *configuration.Config, tlsConfig *tls.Config) *XcrpConnector {
	confKey := fmt.Sprintf("xconfwebconfig.%v.canarymgr_host", xcrpServiceName)
	var hosts []string
	hosts = conf.GetStringList(confKey)
	if hosts == nil || len(hosts) == 0 {
		panic(fmt.Errorf("%s is required", confKey))
	}

	// Read path configurations with defaults
	precookPathTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.precookPathTemplate", xcrpServiceName))
	recookPathTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.recookPathTemplate", xcrpServiceName))

	precookModelPathTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.precookModelPathTemplate", xcrpServiceName))
	precookPartnerPathTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.precookPartnerPathTemplate", xcrpServiceName))
	precookStatusPathTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.recookStatusPathTemplate", xcrpServiceName))

	return &XcrpConnector{
		HttpClient:                 NewHttpClient(conf, xcrpServiceName, tlsConfig),
		hosts:                      hosts,
		precookPathTemplate:        precookPathTemplate,
		recookPathTemplate:         recookPathTemplate,
		precookModelPathTemplate:   precookModelPathTemplate,
		precookPartnerPathTemplate: precookPartnerPathTemplate,
		precookStatusPathTemplate:  precookStatusPathTemplate,
	}
}

func (c *XcrpConnector) XcrpHosts() []string {
	return c.hosts
}

func (c *XcrpConnector) SetXcrpHosts(hosts []string) {
	c.hosts = hosts
}

func (c *XcrpConnector) PostRecook(m, p []string, bbytes []byte, fields log.Fields) error {
	models := strings.Join(m, ",")
	partners := strings.Join(p, ",")
	var url string
	for _, host := range c.XcrpHosts() {
		if len(models) == 0 && len(partners) == 0 {
			url = fmt.Sprintf(c.precookPathTemplate, host)
		} else if len(models) != 0 && len(partners) != 0 {
			url = fmt.Sprintf(c.recookPathTemplate, host, partners, models)
		} else if len(models) != 0 { // input empty string to xcrp will have issues. corner cases handled here for now
			url = fmt.Sprintf(c.precookModelPathTemplate, host, models)
		} else if len(partners) != 0 {
			url = fmt.Sprintf(c.precookPartnerPathTemplate, host, partners)
		}
		headers := map[string]string{
			common.HeaderUserAgent: common.HeaderXconfAdminService,
		}

		_, err := c.DoWithRetries("POST", url, headers, bbytes, fields, xcrpServiceName)
		log.Infof("PostRecook url: %s", url)
		if err != nil {
			return common.NewError(err)
		}

	}

	return nil
}

func (c *XcrpConnector) GetRecookingStatusFromCanaryMgr(module string, fields log.Fields) (bool, error) {
	var url string
	for _, host := range c.XcrpHosts() {
		url = fmt.Sprintf(c.precookStatusPathTemplate, host, module)
		headers := map[string]string{
			common.HeaderUserAgent: common.HeaderXconfAdminService,
		}
		response, err := c.DoWithRetries("GET", url, headers, nil, nil, xcrpServiceName)
		if err != nil {
			return false, err
		}

		var result struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
			Data    struct {
				Status      string `json:"status"`
				UpdatedTime string `json:"updatedTime"`
			} `json:"data"`
		}

		err = json.Unmarshal(response, &result)
		if err != nil {
			return false, err
		}

		if result.Data.Status == "completed" {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, nil
}
