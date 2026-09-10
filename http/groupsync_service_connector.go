package http

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	proto2 "github.com/rdkcentral/xconfadmin/taggingapi/proto/generated"
	"github.com/rdkcentral/xconfadmin/util"

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
)

type GroupServiceSyncConnector struct {
	BaseURL                   string
	Client                    *HttpClient
	addGroupMemberTemplate    string
	removeGroupMemberTemplate string
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

	// Read path configurations with defaults
	addGroupMemberTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.addGroupMemberTemplate", groupServiceSyncServiceName))

	if util.IsBlank(addGroupMemberTemplate) {
		log.Errorf("addGroupMemberTemplate is required")
	}

	removeGroupMemberTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.removeGroupMemberTemplate", groupServiceSyncServiceName))

	if util.IsBlank(removeGroupMemberTemplate) {
		log.Errorf("removeGroupMemberTemplate is required")
	}

	// Validate arity eagerly. fmt.Sprintf silently produces "%!s(MISSING)" or
	// "%!(EXTRA ...)" on a wrong verb count, which yields an unusable URL and a
	// per-member request failure at runtime rather than a startup failure.
	mustBeValidTemplate(host+path, "addGroupMemberTemplate", addGroupMemberTemplate, 2)
	mustBeValidTemplate(host+path, "removeGroupMemberTemplate", removeGroupMemberTemplate, 3)

	return &GroupServiceSyncConnector{
		BaseURL:                   host + path,
		Client:                    NewHttpClient(conf, groupServiceSyncServiceName, tlsConfig),
		addGroupMemberTemplate:    addGroupMemberTemplate,
		removeGroupMemberTemplate: removeGroupMemberTemplate,
	}
}

// mustBeValidTemplate renders a URL template with placeholder arguments and
// panics if the result is not a usable URL. A blank template is skipped — the
// request paths fail closed on blank templates instead (an unconfigured
// template yields a per-request error, not a bad URL).
//
// This mirrors the treatment of a blank host above: a misconfigured template is
// a deployment error that should stop the process, not produce a service that
// accepts writes and silently stores nothing.
func mustBeValidTemplate(baseURL, name, template string, argCount int) {
	if util.IsBlank(template) {
		return
	}
	args := make([]interface{}, argCount)
	args[0] = baseURL
	for i := 1; i < argCount; i++ {
		args[i] = "placeholder"
	}
	rendered := fmt.Sprintf(template, args...)
	if strings.Contains(rendered, "%!") {
		panic(fmt.Errorf("%s is malformed: expected %d format verbs, rendered as %q", name, argCount, rendered))
	}
	if _, err := url.Parse(rendered); err != nil {
		panic(fmt.Errorf("%s does not produce a valid URL: %w", name, err))
	}
}

func (c *GroupServiceSyncConnector) GetGroupServiceSyncHost() string {
	return c.BaseURL
}

func (c *GroupServiceSyncConnector) SetGroupServiceSyncHost(host string) {
	c.BaseURL = host
}

func (c *GroupServiceSyncConnector) SetAddGroupMemberTemplate(template string) {
	c.addGroupMemberTemplate = template
}

func (c *GroupServiceSyncConnector) SetRemoveGroupMemberTemplate(template string) {
	c.removeGroupMemberTemplate = template
}

func (c *GroupServiceSyncConnector) DoRequest(method string, url string, headers map[string]string, body []byte) ([]byte, error) {
	rbytes, err := c.Client.DoWithRetries(method, url, headers, body, log.Fields{}, groupServiceSyncServiceName)
	return rbytes, err
}

// AddMembersToTag adds tag fields for a member. Device (mac) and account
// members share one XDAS keyspace; the member id format is the caller's
// concern (see NormalizeMember in taggingapi/tag).
//
// Note the parameter naming is historically inverted: groupId is the *member*
// (the XDAS record key) and the fields of members are the tags.
func (c *GroupServiceSyncConnector) AddMembersToTag(groupId string, members *proto2.XdasHashes) error {
	if util.IsBlank(c.addGroupMemberTemplate) {
		return fmt.Errorf("addGroupMemberTemplate is not configured")
	}
	url := fmt.Sprintf(c.addGroupMemberTemplate, c.GetGroupServiceSyncHost(), groupId)
	data, err := proto.Marshal(members)
	if err != nil {
		return err
	}
	headers := protobufHeaders()
	headers[TtlHeader] = OneYearTtl
	rbytes, err := c.DoRequest("POST", url, headers, data)
	if err != nil {
		return err
	}

	// Log response for visibility into XDAS behavior
	if len(rbytes) > 0 {
		log.Debugf("XDAS AddMembersToTag response: groupId=%s, body=%s", groupId, string(rbytes))
	}

	return nil
}

func (c *GroupServiceSyncConnector) RemoveGroupMembers(groupId string, member string) error {
	if util.IsBlank(c.removeGroupMemberTemplate) {
		return fmt.Errorf("removeGroupMemberTemplate is not configured")
	}
	url := fmt.Sprintf(c.removeGroupMemberTemplate, c.GetGroupServiceSyncHost(), groupId, member)
	rbytes, err := c.DoRequest("DELETE", url, protobufHeaders(), nil)
	if err != nil {
		return err
	}

	// Log response for visibility into XDAS behavior
	if len(rbytes) > 0 {
		log.Debugf("XDAS RemoveGroupMembers response: groupId=%s, member=%s, body=%s", groupId, member, string(rbytes))
	}

	return nil
}

func protobufHeaders() map[string]string {
	headers := make(map[string]string)
	headers[Accept] = ApplicationProtobufHeader
	headers[ContentType] = ApplicationProtobufHeader
	return headers
}
