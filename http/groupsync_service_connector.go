package http

import (
	"crypto/tls"
	"fmt"

	proto2 "xconfadmin/taggingapi/proto/generated"
	"xconfadmin/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var groupServiceSyncServiceName string

const (
	Accept                    = "Content-Type"
	ContentType               = "Content-Type"
	ApplicationProtobufHeader = "application/x-protobuf"
	TtlHeader                 = "Xttl"
	OneYearTtl                = "31536000"

	AddGroupMember    = "%s/v2/ft/%s"
	RemoveGroupMember = "%s/v2/ft/%s?field=%s"
)

type GroupServiceSyncConnector struct {
	BaseURL string
	Client  *HttpClient
}

func NewGroupServiceSyncConnector(conf *configuration.Config, tlsConfig *tls.Config) *GroupServiceSyncConnector {
	groupServiceSyncServiceName = conf.GetString("xconfwebconfig.xconf.group_sync_service_name")
	confKey := fmt.Sprintf("xconfwebconfig.%v.host", groupServiceSyncServiceName)
	host := conf.GetString(confKey)
	if util.IsBlank(host) {
		panic(fmt.Errorf("%s is required", confKey))
	}
	confKey = fmt.Sprintf("xconfwebconfig.%v.path", groupServiceSyncServiceName)
	path := conf.GetString(confKey, "")
	return &GroupServiceSyncConnector{
		BaseURL: host + path,
		Client:  NewHttpClient(conf, groupServiceSyncServiceName, tlsConfig),
	}
}

func (c *GroupServiceSyncConnector) GetGroupServiceSyncUrl() string {
	return c.BaseURL
}

func (c *GroupServiceSyncConnector) DoRequest(method string, url string, headers map[string]string, body []byte) ([]byte, error) {
	rbytes, err := c.Client.DoWithRetries(method, url, headers, body, log.Fields{}, groupServiceSyncServiceName)
	return rbytes, err
}

func (c *GroupServiceSyncConnector) AddMembersToTag(groupId string, members *proto2.XdasHashes) error {
	url := fmt.Sprintf(AddGroupMember, c.GetGroupServiceSyncUrl(), groupId)
	data, err := proto.Marshal(members)
	if err != nil {
		return err
	}
	headers := protobufHeaders()
	headers[TtlHeader] = OneYearTtl
	_, err = c.DoRequest("POST", url, headers, data)
	if err != nil {
		return err
	}
	return nil
}

func (c *GroupServiceSyncConnector) RemoveGroupMembers(groupId string, member string) error {
	url := fmt.Sprintf(RemoveGroupMember, c.GetGroupServiceSyncUrl(), groupId, member)
	_, err := c.DoRequest("DELETE", url, protobufHeaders(), nil)
	if err != nil {
		return err
	}
	return nil
}

func protobufHeaders() map[string]string {
	headers := make(map[string]string)
	headers[Accept] = ApplicationProtobufHeader
	headers[ContentType] = ApplicationProtobufHeader
	return headers
}
