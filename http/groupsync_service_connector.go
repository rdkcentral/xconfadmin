package http

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"github.com/rdkcentral/xconfadmin/common"
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
	// Account tags live in a different XDAS keyspace than device (mac) tags.
	// Kept as config-driven templates rather than a hardcoded path so the
	// keyspace can be changed without a code change.
	addAccountGroupMemberTemplate    string
	removeAccountGroupMemberTemplate string
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

	addAccountGroupMemberTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.addAccountGroupMemberTemplate", groupServiceSyncServiceName))

	removeAccountGroupMemberTemplate := conf.GetString(
		fmt.Sprintf("xconfwebconfig.%v.removeAccountGroupMemberTemplate", groupServiceSyncServiceName))

	// Validate arity eagerly. fmt.Sprintf silently produces "%!s(MISSING)" or
	// "%!(EXTRA ...)" on a wrong verb count, which yields an unusable URL and a
	// per-member request failure at runtime rather than a startup failure. The
	// account templates are optional (account tagging may be unconfigured), but
	// if present they must be well-formed.
	mustBeValidTemplate(host+path, "addGroupMemberTemplate", addGroupMemberTemplate, 2)
	mustBeValidTemplate(host+path, "removeGroupMemberTemplate", removeGroupMemberTemplate, 3)
	mustBeValidTemplate(host+path, "addAccountGroupMemberTemplate", addAccountGroupMemberTemplate, 3)
	mustBeValidTemplate(host+path, "removeAccountGroupMemberTemplate", removeAccountGroupMemberTemplate, 3)

	return &GroupServiceSyncConnector{
		BaseURL:                          host + path,
		Client:                           NewHttpClient(conf, groupServiceSyncServiceName, tlsConfig),
		addGroupMemberTemplate:           addGroupMemberTemplate,
		removeGroupMemberTemplate:        removeGroupMemberTemplate,
		addAccountGroupMemberTemplate:    addAccountGroupMemberTemplate,
		removeAccountGroupMemberTemplate: removeAccountGroupMemberTemplate,
	}
}

// mustBeValidTemplate renders a URL template with placeholder arguments and
// panics if the result is not a usable URL. A blank template is skipped — the
// caller decides whether blank is fatal (see templateFor, which fails closed).
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

func (c *GroupServiceSyncConnector) SetAddAccountGroupMemberTemplate(template string) {
	c.addAccountGroupMemberTemplate = template
}

func (c *GroupServiceSyncConnector) SetRemoveAccountGroupMemberTemplate(template string) {
	c.removeAccountGroupMemberTemplate = template
}

// addTemplateFor returns the add-member URL template for a tag type.
//
// It fails closed for account tags: if the account template is unconfigured we
// return an error rather than falling back to the device template. Writing
// account ids into the device keyspace is unrecoverable pollution — the ids are
// indistinguishable from members once stored.
func (c *GroupServiceSyncConnector) addTemplateFor(tagType string) (string, error) {
	if tagType == common.TagTypeAccount {
		if util.IsBlank(c.addAccountGroupMemberTemplate) {
			return "", fmt.Errorf("addAccountGroupMemberTemplate is not configured; account tagging is unavailable")
		}
		return c.addAccountGroupMemberTemplate, nil
	}
	if util.IsBlank(c.addGroupMemberTemplate) {
		return "", fmt.Errorf("addGroupMemberTemplate is not configured")
	}
	return c.addGroupMemberTemplate, nil
}

func (c *GroupServiceSyncConnector) removeTemplateFor(tagType string) (string, error) {
	if tagType == common.TagTypeAccount {
		if util.IsBlank(c.removeAccountGroupMemberTemplate) {
			return "", fmt.Errorf("removeAccountGroupMemberTemplate is not configured; account tagging is unavailable")
		}
		return c.removeAccountGroupMemberTemplate, nil
	}
	if util.IsBlank(c.removeGroupMemberTemplate) {
		return "", fmt.Errorf("removeGroupMemberTemplate is not configured")
	}
	return c.removeGroupMemberTemplate, nil
}

func (c *GroupServiceSyncConnector) DoRequest(method string, url string, headers map[string]string, body []byte) ([]byte, error) {
	rbytes, err := c.Client.DoWithRetries(method, url, headers, body, log.Fields{}, groupServiceSyncServiceName)
	return rbytes, err
}

// AddMembersToTag adds tag fields for a device (mac) member.
//
// Note the parameter naming is historically inverted: groupId is the *member*
// (the XDAS record key) and the fields of members are the tags.
func (c *GroupServiceSyncConnector) AddMembersToTag(groupId string, members *proto2.XdasHashes) error {
	return c.AddMembersToTagOfType(groupId, members, common.TagTypeMac)
}

// AddMembersToTagOfType is AddMembersToTag with an explicit tag type, which
// selects the XDAS keyspace.
func (c *GroupServiceSyncConnector) AddMembersToTagOfType(groupId string, members *proto2.XdasHashes, tagType string) error {
	template, err := c.addTemplateFor(tagType)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(template, c.GetGroupServiceSyncHost(), groupId)
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
	return c.RemoveGroupMembersOfType(groupId, member, common.TagTypeMac)
}

// RemoveGroupMembersOfType is RemoveGroupMembers with an explicit tag type,
// which selects the XDAS keyspace.
func (c *GroupServiceSyncConnector) RemoveGroupMembersOfType(groupId string, member string, tagType string) error {
	template, err := c.removeTemplateFor(tagType)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(template, c.GetGroupServiceSyncHost(), groupId, member)
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
