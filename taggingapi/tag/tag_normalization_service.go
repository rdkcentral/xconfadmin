package tag

import (
	"fmt"
	"strings"
	"xconfadmin/util"

	log "github.com/sirupsen/logrus"
)

const (
	Prefix   = "t_"
	Template = "%s%s"
)

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
