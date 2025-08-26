package http

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"

	proto2 "github.com/rdkcentral/xconfadmin/taggingapi/proto/generated"
	"github.com/rdkcentral/xconfadmin/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var groupServiceName string

const (
	GetGroupsMembers       = "%s/v2/ft/%s"
	GetAllGroups           = "%s/v2/ft"
	GetAllCanaryGroupsPath = "%s/v2/gs/list/all?format=json"
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

func (c *GroupServiceConnector) GetAllCanaryGroups() ([]string, error) {
	url := fmt.Sprintf(GetAllCanaryGroupsPath, c.GetGroupServiceHost())

	headers := map[string]string{
		"Accept":     "application/json",
		"Connection": "close",
	}

	rbytes, err := c.DoRequest(HttpGet, url, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch canary groups: %v", err)
	}

	if len(rbytes) == 0 {
		return nil, fmt.Errorf("received empty response from canary service")
	}

	var response map[string]interface{}
	err = json.Unmarshal(rbytes, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %v", err)
	}

	meta, ok := response["meta"].(map[string]interface{})
	if !ok {
		return []string{}, nil
	}

	groups := make([]string, 0, len(meta))

	for groupName, groupData := range meta {
		if !c.isTimestampEndedGroup(groupName) {
			continue
		}

		if groupInfo, ok := groupData.(map[string]interface{}); ok {
			if counter, exists := groupInfo["counter"]; exists {
				groups = append(groups, fmt.Sprintf("%s (Size: %s)", groupName, counter))
			} else {
				groups = append(groups, groupName)
			}
		} else {
			groups = append(groups, groupName)
		}
	}
	return groups, nil
}

func (c *GroupServiceConnector) isTimestampEndedGroup(groupName string) bool {

	if len(groupName) < 6 {
		return false
	}

	parts := strings.Split(groupName, "_")
	if len(parts) < 2 {
		return false
	}

	lastPart := parts[len(parts)-1]
	if len(lastPart) >= 10 && c.isNumeric(lastPart) {
		return true
	}

	if len(parts) >= 3 {
		secondLastPart := parts[len(parts)-2]
		return c.isNumeric(lastPart) && c.isNumeric(secondLastPart)
	}

	return false
}

func (c *GroupServiceConnector) isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
