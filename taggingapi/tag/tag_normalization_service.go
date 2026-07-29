package tag

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rdkcentral/xconfadmin/common"
	"github.com/rdkcentral/xconfadmin/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	log "github.com/sirupsen/logrus"
)

const (
	Prefix   = "t_"
	Template = "%s%s"
)

// Tag types. Aliased from common so this package reads naturally while the
// canonical values stay importable by the http connectors, which select an XDAS
// keyspace from the type and cannot import this package (taggingapi/tag already
// depends on http).
const (
	TagTypeLegacy  = common.TagTypeLegacy
	TagTypeMac     = common.TagTypeMac
	TagTypeAccount = common.TagTypeAccount
)

// ValidateTagType rejects anything that is not a known tag type. The empty
// string is valid and means "legacy / untyped route".
func ValidateTagType(tagType string) error {
	switch tagType {
	case TagTypeLegacy, TagTypeMac, TagTypeAccount:
		return nil
	default:
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
			fmt.Sprintf("unknown tag type %q, expected %q or %q", tagType, TagTypeMac, TagTypeAccount))
	}
}

func IsAccountTag(tagType string) bool {
	return tagType == TagTypeAccount
}

// NormalizeMember converts a member to its canonical form for the given tag type.
//
// It returns an error rather than a bare string because account normalization
// can fail: swallowing a malformed account id would let it through to the
// account keyspace, where it is indistinguishable from a real one afterwards.
//
// For mac and legacy tags the result is byte-identical to ToNormalizedEcm, which
// is what keeps the untyped routes behaving exactly as they do today.
func NormalizeMember(member string, tagType string) (string, error) {
	if IsAccountTag(tagType) {
		normalized := strings.TrimSpace(member)
		if _, err := util.AccountIdValidator(normalized); err != nil {
			return "", xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
				fmt.Sprintf("invalid account id %q: account ids must be numeric", member))
		}
		return normalized, nil
	}
	return ToNormalizedEcm(member), nil
}

func ToNormalizedEcm(member string) string {
	member = strings.TrimSpace(member)
	if valid, _ := util.MACAddressValidator(member); valid {
		member = util.ToAlphaNumericString(member)
		member = util.GetEcmMacAddress(member)
	}
	return strings.ToUpper(member)
}

func ToNormalized(member string) string {
	member = strings.TrimSpace(member)
	return strings.ToUpper(member)
}

func ToEstbIfMac(member string) string {
	if valid, _ := util.MACAddressValidator(member); valid {
		return util.GetEstbMacAddress(member)
	}
	return member
}

func SetTagPrefix(tagId string) string {
	if strings.HasPrefix(tagId, Prefix) {
		log.Warn(fmt.Sprintf("%s tag already has prefix", tagId))
		return tagId
	}
	return fmt.Sprintf(Template, Prefix, tagId)
}

func RemovePrefixFromTag(tagId string) string {
	after, _ := strings.CutPrefix(tagId, Prefix)
	return after
}

func removePrefixFromTags(tags []string) []string {
	for i := 0; i < len(tags); i++ {
		tags[i] = RemovePrefixFromTag(tags[i])
	}
	return tags
}
