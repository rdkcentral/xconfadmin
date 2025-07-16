package util

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const (
	HeaderAuthorization        = "Authorization"
	HeaderUserAgent            = "User-Agent"
	HeaderIfNoneMatch          = "If-None-Match"
	HeaderFirmwareVersion      = "X-System-Firmware-Version"
	HeaderSupportedDocs        = "X-System-Supported-Docs"
	HeaderSupplementaryService = "X-System-SupplementaryService-Sync"
	HeaderModelName            = "X-System-Model-Name"
	HeaderProfileVersion       = "X-System-Telemetry-Profile-Version"
	HeaderPartnerID            = "X-System-PartnerID"
	HeaderAccountID            = "X-System-AccountID"
	HeaderXconfDataService     = "XconfDataService"
	HeaderXconfAdminService    = "XconfAdminService"
)

var (
	SupportedPokeDocs = []string{"primary", "telemetry"}
)

var (
	telemetryFields = [][]string{
		{"version", HeaderProfileVersion},
		{"model", HeaderModelName},
		{"partnerId", HeaderPartnerID},
		{"accountId", HeaderAccountID},
		{"firmwareVersion", HeaderFirmwareVersion},
	}

	alnumRe    = regexp.MustCompile("[^a-zA-Z0-9]+")
	validMacRe = regexp.MustCompile(`^([0-9a-fA-F]{12}$)|([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})|([0-9A-Fa-f]{4}[.]){2}([0-9A-Fa-f]{4})$`)
)

func ToAlphaNumericString(str string) string {
	return alnumRe.ReplaceAllString(str, "")
}

func ToColonMac(d string) string {
	return fmt.Sprintf("%v:%v:%v:%v:%v:%v", d[:2], d[2:4], d[4:6], d[6:8], d[8:10], d[10:12])
}

func GetAuditId() string {
	u := uuid.New()
	ustr := u.String()
	uustr := strings.ReplaceAll(ustr, "-", "")
	return uustr
}

func GenerateRandomCpeMac() string {
	u := uuid.New().String()
	return strings.ToUpper(u[len(u)-12:])
}

func GetTelemetryQueryString(header http.Header, mac string) string {
	// build the query parameters in a fixed order
	params := []string{}

	firmwareVersion := header.Get(HeaderFirmwareVersion)
	if strings.Contains(firmwareVersion, "PROD") {
		params = append(params, "env=PROD")
	} else if strings.Contains(firmwareVersion, "DEV") {
		params = append(params, "env=DEV")
	}

	for _, pairs := range telemetryFields {
		params = append(params, fmt.Sprintf("%v=%v", pairs[0], header.Get(pairs[1])))
	}

	estbMacAddress := GetEstbMacAddress(mac)
	params = append(params, fmt.Sprintf("estbMacAddress=%v", estbMacAddress))
	params = append(params, fmt.Sprintf("ecmMacAddress=%v", mac))

	return strings.Join(params, "&")
}

func ValidatePokeQuery(values url.Values) (string, error) {
	// set the default
	queryStr := "primary"

	docQueryParamStrs, ok := values["doc"]
	if ok {
		if len(docQueryParamStrs) > 1 {
			err := fmt.Errorf("multiple doc parameter is not allowed")
			return "", err
		}

		qparams := strings.Split(docQueryParamStrs[0], ",")
		if len(qparams) > 1 {
			err := fmt.Errorf("multiple doc parameter is not allowed")
			return "", err
		}

		queryStr = qparams[0]
		if !Contains(SupportedPokeDocs, queryStr) {
			err := fmt.Errorf("invalid query parameter: %v", queryStr)
			return "", err

		}
	}
	return queryStr, nil
}

func GetEstbMacAddress(mac string) string {
	// if the mac cannot be parsed, then return back the input
	i, err := strconv.ParseInt(mac, 16, 64)
	if err != nil {
		return mac
	}
	return fmt.Sprintf("%012X", i+2)
}

func GetEcmMacAddress(mac string) string {
	// if the mac cannot be parsed, then return back the input
	i, err := strconv.ParseInt(mac, 16, 64)
	if err != nil {
		return mac
	}
	return fmt.Sprintf("%012X", i-2)
}

// REMINDER
// a 2-D slices/arrays of strings are chosen, instead of a map, to keep the params ordering
func GetURLQueryParameterString(kvs [][]string) (string, error) {
	params := []string{}
	for _, kv := range kvs {
		if len(kv) != 2 {
			err := fmt.Errorf("len(kv) != 2")
			return "", err
		}
		params = append(params, fmt.Sprintf("%v=%v", kv[0], kv[1]))
	}
	return strings.Join(params, "&"), nil
}

func IsUnknownValue(param string) bool {
	return strings.EqualFold(param, "unknown") || strings.EqualFold(param, "NoAccount")
}

// MACAddressValidator method is to validate MAC address
// Validate inputs are:
//
//	11-11-11-11-11-11
//	11:11:11:11:11:11
//	1111.1111.1111
//	11111111111
func MACAddressValidator(macAddress string) (bool, error) {
	if validMacRe.MatchString(macAddress) {
		return true, nil
	}

	return false, errors.New("Invalid MAC address")
}

// AlphaNumericMacAddress is converting MAC address to only alphanumeric
func AlphaNumericMacAddress(macAddress string) string {
	macAddress = ToAlphaNumericString(macAddress)
	macAddress = strings.ToUpper(macAddress)
	return macAddress
}

// ValidateAndNormalizeMacAddress is to validate and convert MAC address to XX:XX:XX:XX:XX:XX
func ValidateAndNormalizeMacAddress(macaddr string) (string, error) {
	// 1st validates the mac address
	_, err := MACAddressValidator(macaddr)
	if err != nil {
		return "", err
	}

	// Replace all dash, colon or period from MAC address
	mac := AlphaNumericMacAddress(macaddr)
	return ToColonMac(mac), nil
}

func IsValidMacAddress(macaddr string) bool {
	_, err := MACAddressValidator(macaddr)
	return err == nil
}

// Use this function only if you know the mac address is valid
func NormalizeMacAddress(macAddress string) string {
	macAddress = AlphaNumericMacAddress(macAddress)
	if len(macAddress) != 12 {
		return ""
	}
	return ToColonMac(macAddress)
}

func IsBlank(str string) bool {
	return strings.Trim(str, " ") == ""
}

func RemoveOneElementFromList(ids []string, idToRemove string) []string {
	for i, id := range ids {
		if idToRemove == id {
			return append(ids[:i], ids[i+1:]...)
		}
	}
	return ids
}

func StringSliceContains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
