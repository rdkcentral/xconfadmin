package http

import (
	"crypto/tls"
	"fmt"

	"github.com/rdkcentral/xconfadmin/common"
	proto2 "github.com/rdkcentral/xconfadmin/taggingapi/proto/generated"
	"github.com/rdkcentral/xconfadmin/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var groupServiceName string

type GroupServiceConnector struct {
	BaseURL                  string
	Client                   *HttpClient
	getGroupsMembersTemplate string
	getAllGroupsTemplate     string
	// Reverse lookup for account tags reads a different XDAS keyspace than
	// device (mac) tags. Config-driven for the same reason as the write side.
	getAccountGroupsMembersTemplate string
}

func (c *GroupServiceConnector) GetGroupServiceHost() string {
	return c.BaseURL
}

func (c *GroupServiceConnector) SetGroupServiceHost(host string) {
	c.BaseURL = host
}

func (c *GroupServiceConnector) SetGetGroupsMembersTemplate(template string) {
	c.getGroupsMembersTemplate = template
}

func (c *GroupServiceConnector) SetGetAllGroupsTemplate(template string) {
	c.getAllGroupsTemplate = template
}

func (c *GroupServiceConnector) SetGetAccountGroupsMembersTemplate(template string) {
	c.getAccountGroupsMembersTemplate = template
}

// getMembersTemplateFor returns the reverse-lookup template for a tag type.
// Fails closed for account tags — see addTemplateFor in the sync connector.
func (c *GroupServiceConnector) getMembersTemplateFor(tagType string) (string, error) {
	if tagType == common.TagTypeAccount {
		if util.IsBlank(c.getAccountGroupsMembersTemplate) {
			return "", fmt.Errorf("getAccountGroupsMembersTemplate is not configured; account tagging is unavailable")
		}
		return c.getAccountGroupsMembersTemplate, nil
	}
	if util.IsBlank(c.getGroupsMembersTemplate) {
		return "", fmt.Errorf("getGroupsMembersTemplate is not configured")
	}
	return c.getGroupsMembersTemplate, nil
}

func NewGroupServiceConnector(conf *configuration.Config, tlsConfig *tls.Config) *GroupServiceConnector {
	groupServiceName := conf.GetString("xconfwebconfig.xconf.group_service_name")
	confKey := fmt.Sprintf("xconfwebconfig.%v.host", groupServiceName)
	host := conf.GetString(confKey)
	if util.IsBlank(host) {
		panic(fmt.Errorf("%s is required", confKey))
	}

	getGroupsMembersTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.getGroupsMembersTemplate", groupServiceName))

	if util.IsBlank(getGroupsMembersTemplate) {
		log.Error("getGroupsMembersTemplate is required")
	}

	getAllGroupsTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.getAllGroupsTemplate", groupServiceName))

	if util.IsBlank(getAllGroupsTemplate) {
		log.Error("getAllGroupsTemplate is required")
	}

	getAccountGroupsMembersTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.getAccountGroupsMembersTemplate", groupServiceName))

	mustBeValidTemplate(host, "getGroupsMembersTemplate", getGroupsMembersTemplate, 2)
	mustBeValidTemplate(host, "getAllGroupsTemplate", getAllGroupsTemplate, 1)
	mustBeValidTemplate(host, "getAccountGroupsMembersTemplate", getAccountGroupsMembersTemplate, 2)

	return &GroupServiceConnector{
		BaseURL:                         host,
		Client:                          NewHttpClient(conf, groupServiceName, tlsConfig),
		getGroupsMembersTemplate:        getGroupsMembersTemplate,
		getAllGroupsTemplate:            getAllGroupsTemplate,
		getAccountGroupsMembersTemplate: getAccountGroupsMembersTemplate,
	}
}

func (c *GroupServiceConnector) DoRequest(method string, url string, headers map[string]string, body []byte) ([]byte, error) {
	rbytes, err := c.Client.DoWithRetries(method, url, headers, body, log.Fields{}, groupServiceName)
	return rbytes, err
}

func (c *GroupServiceConnector) GetGroupsMemberBelongsTo(memberId string) (*proto2.XdasHashes, error) {
	return c.GetGroupsMemberBelongsToOfType(memberId, common.TagTypeMac)
}

// GetGroupsMemberBelongsToOfType is GetGroupsMemberBelongsTo with an explicit
// tag type, which selects the XDAS keyspace.
func (c *GroupServiceConnector) GetGroupsMemberBelongsToOfType(memberId string, tagType string) (*proto2.XdasHashes, error) {
	template, err := c.getMembersTemplateFor(tagType)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(template, c.GetGroupServiceHost(), memberId)
	rbytes, err := c.DoRequest(HttpGet, url, protobufHeaders(), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalXdasHashes(rbytes)
}

func (c *GroupServiceConnector) GetAllGroups() (*proto2.XdasHashes, error) {
	url := fmt.Sprintf(c.getAllGroupsTemplate, c.GetGroupServiceHost())
	rbytes, err := c.DoRequest(HttpGet, url, protobufHeaders(), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalXdasHashes(rbytes)
}

func unmarshalXdasHashes(bytes []byte) (*proto2.XdasHashes, error) {
	var groups proto2.XdasHashes
	err := proto.Unmarshal(bytes, &groups)
	if err != nil {
		return nil, err
	}
	return &groups, nil
}
