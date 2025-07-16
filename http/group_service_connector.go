package http

import (
	"crypto/tls"
	"fmt"

	proto2 "github.com/rdkcentral/xconfadmin/taggingapi/proto/generated"
	"github.com/rdkcentral/xconfadmin/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var groupServiceName string

const (
	GetGroupsMembers = "%s/v2/ft/%s"
	GetAllGroups     = "%s/v2/ft"
)

type GroupServiceConnector struct {
	BaseURL string
	Client  *HttpClient
}

func (c *GroupServiceConnector) GetGroupServiceHost() string {
	return c.BaseURL
}

func (c *GroupServiceConnector) SetGroupServiceHost(host string) {
	c.BaseURL = host
}

func NewGroupServiceConnector(conf *configuration.Config, tlsConfig *tls.Config) *GroupServiceConnector {
	groupServiceName := conf.GetString("xconfwebconfig.xconf.group_service_name")
	confKey := fmt.Sprintf("xconfwebconfig.%v.host", groupServiceName)
	host := conf.GetString(confKey)
	if util.IsBlank(host) {
		panic(fmt.Errorf("%s is required", confKey))
	}

	return &GroupServiceConnector{
		BaseURL: host,
		Client:  NewHttpClient(conf, groupServiceName, tlsConfig),
	}
}

func (c *GroupServiceConnector) DoRequest(method string, url string, headers map[string]string, body []byte) ([]byte, error) {
	rbytes, err := c.Client.DoWithRetries(method, url, headers, body, log.Fields{}, groupServiceName)
	return rbytes, err
}

func (c *GroupServiceConnector) GetGroupsMemberBelongsTo(memberId string) (*proto2.XdasHashes, error) {
	url := fmt.Sprintf(GetGroupsMembers, c.GetGroupServiceHost(), memberId)
	rbytes, err := c.DoRequest(HttpGet, url, protobufHeaders(), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalXdasHashes(rbytes)
}

func (c *GroupServiceConnector) GetAllGroups() (*proto2.XdasHashes, error) {
	url := fmt.Sprintf(GetAllGroups, c.GetGroupServiceHost())
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
