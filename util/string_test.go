package util

import (
	"net/http"
	"net/url"
	"testing"

	"gotest.tools/assert"
)

func TestString(t *testing.T) {
	s := "112233445566"
	c := ToColonMac(s)
	expected := "11:22:33:44:55:66"
	assert.Equal(t, c, expected)
}

func TestAlphaNumericMacAddress(t *testing.T) {
	expected := "6AF6F8B65794"
	mac := "6a-F6.F8 b6:57-94"
	assert.Equal(t, AlphaNumericMacAddress(mac), expected)
}

func TestGetAuditId(t *testing.T) {
	auditId := GetAuditId()
	assert.Equal(t, len(auditId), 32)
}

func TestTelemetryQuery(t *testing.T) {
	header := http.Header{}
	header.Set(HeaderProfileVersion, "2.0")
	header.Set(HeaderModelName, "TG1682G")
	header.Set(HeaderPartnerID, "abc")
	header.Set(HeaderAccountID, "1234567890")
	header.Set(HeaderFirmwareVersion, "TG1682_3.14p9s6_PROD_sey")
	mac := "567890ABCDEF"
	qstr := GetTelemetryQueryString(header, mac)

	expected := "env=PROD&version=2.0&model=TG1682G&partnerId=abc&accountId=1234567890&firmwareVersion=TG1682_3.14p9s6_PROD_sey&estbMacAddress=567890ABCDF1&ecmMacAddress=567890ABCDEF"
	assert.Equal(t, qstr, expected)
}

func TestGetQueryParameters(t *testing.T) {
	// ==== normal ====
	kvs := [][]string{
		{"env", "PROD"},
		{"version", "2.0"},
		{"model", "CGM4140COM"},
		{"partnerId", "abcd"},
		{"accountId", "1234567890"},
		{"firmwareVersion", "testfirmwareVersion"},
		{"estbMacAddress", "112233445565"},
		{"ecmMacAddress", "112233445567"},
	}
	expected := "env=PROD&version=2.0&model=CGM4140COM&partnerId=abcd&accountId=1234567890&firmwareVersion=testfirmwareVersion&estbMacAddress=112233445565&ecmMacAddress=112233445567"
	queryParams, err := GetURLQueryParameterString(kvs)
	assert.NilError(t, err)
	assert.Equal(t, expected, queryParams)

	// ==== ill formatted ====
	kvs = [][]string{
		{"env", "PROD"},
		{"version", "2.0"},
		{"model", "CGM4140COM"},
		{"partnerId", "abcd", "abcde"},
		{"accountId", "1234567890"},
		{"firmwareVersion", "testfirmwareVersion"},
		{"estbMacAddress", "112233445565"},
		{"ecmMacAddress", "112233445567"},
	}
	_, err = GetURLQueryParameterString(kvs)
	assert.Assert(t, err != nil)
}

func TestIsUnknownValue(t *testing.T) {
	isUnknown := IsUnknownValue("hello")
	assert.Equal(t, isUnknown, false)

	isUnknown = IsUnknownValue("")
	assert.Equal(t, isUnknown, false)

	isUnknown = IsUnknownValue("UNKNOWN")
	assert.Equal(t, isUnknown, true)

	isUnknown = IsUnknownValue("noaccount")
	assert.Equal(t, isUnknown, true)
}

func TestMACAddressValidator(t *testing.T) {
	// Positive scenarios
	validMac, err := MACAddressValidator("142536ABAC23")
	assert.Equal(t, validMac, true)
	assert.NilError(t, err)

	validMac, err = MACAddressValidator("14:68:36:AB:DD:23")
	assert.Equal(t, validMac, true)

	validMac, err = MACAddressValidator("14-25-36-AB-AC-23")
	assert.Equal(t, validMac, true)

	validMac, err = MACAddressValidator("bd-c5-9a-7e-fd-23")
	assert.Equal(t, validMac, true)

	validMac, err = MACAddressValidator("bdc5.9a7e.fd23")
	assert.Equal(t, validMac, true)

	// Negative scenarios
	validMac, err = MACAddressValidator("14-25-36-LP-AT-23")
	assert.Equal(t, validMac, false)
	assert.Error(t, err, "Invalid MAC address")

	validMac, err = MACAddressValidator("14253 6LPAT:23")
	assert.Equal(t, validMac, false)

	validMac, err = MACAddressValidator("14-25-36AC-23")
	assert.Equal(t, validMac, false)

	validMac, err = MACAddressValidator("14-25-36AC-23:aa 66")
	assert.Equal(t, validMac, false)

	validMac, err = MACAddressValidator("MAC:142536HBAC23")
	assert.Equal(t, validMac, false)

	validMac, err = MACAddressValidator("AA BB CC DD EE FF")
	assert.Equal(t, validMac, false)

	validMac, err = MACAddressValidator("00112233445Z")
	assert.Equal(t, validMac, false)
}

func TestIsValidMacAddress(t *testing.T) {
	isValidMacAddress := IsValidMacAddress("142536ABAC23")
	assert.Equal(t, isValidMacAddress, true)

	isValidMacAddress = IsValidMacAddress("14:25:36:ab:ac:23")
	assert.Equal(t, isValidMacAddress, true)

	isValidMacAddress = IsValidMacAddress("helloworld")
	assert.Equal(t, isValidMacAddress, false)

	isValidMacAddress = IsValidMacAddress("")
	assert.Equal(t, isValidMacAddress, false)
}

func TestValidateAndNormalizeMacAddress(t *testing.T) {
	// Positive scenarios
	validMac, err := ValidateAndNormalizeMacAddress("142536ABAC23")
	assert.NilError(t, err)
	assert.Equal(t, validMac, "14:25:36:AB:AC:23")

	validMac, err = ValidateAndNormalizeMacAddress("AA:bb:CC:dd:ee:FF")
	assert.NilError(t, err)
	assert.Equal(t, validMac, "AA:BB:CC:DD:EE:FF")

	// Negative scenarios
	validMac, err = ValidateAndNormalizeMacAddress("11 25 R6 AB AC 23")
	assert.Error(t, err, "Invalid MAC address")
	assert.Equal(t, validMac, "")

	_, err = ValidateAndNormalizeMacAddress("AA:bb:CC:dd:ee;FF")
	assert.Error(t, err, "Invalid MAC address")
}

func TestNormalizeMacAddress(t *testing.T) {
	normalizedMac := NormalizeMacAddress("142536abAc23")
	assert.Equal(t, normalizedMac, "14:25:36:AB:AC:23")

	normalizedMac = NormalizeMacAddress("14:25:36:ab:AC:23")
	assert.Equal(t, normalizedMac, "14:25:36:AB:AC:23")

	normalizedMac = NormalizeMacAddress("14-25-36-ab-AC-23")
	assert.Equal(t, normalizedMac, "14:25:36:AB:AC:23")

	normalizedMac = NormalizeMacAddress("142536ab")
	assert.Equal(t, normalizedMac, "")
}

func TestContainsIgnoreCase(t *testing.T) {
	// Positive scenarios
	containsIgnoreCase := ContainsIgnoreCase("hello", "HELLO")
	assert.Equal(t, containsIgnoreCase, true)

	containsIgnoreCase = ContainsIgnoreCase("HELLO, WORLD", "hello")
	assert.Equal(t, containsIgnoreCase, true)

	containsIgnoreCase = ContainsIgnoreCase("Goodbye, Hello", "HELLO")
	assert.Equal(t, containsIgnoreCase, true)

	containsIgnoreCase = ContainsIgnoreCase("Goodbye, Hello again", "hello")
	assert.Equal(t, containsIgnoreCase, true)

	// Nagetive scenarios
	containsIgnoreCase = ContainsIgnoreCase("", "WORLD")
	assert.Equal(t, containsIgnoreCase, false)

	containsIgnoreCase = ContainsIgnoreCase("hello", "WORLD")
	assert.Equal(t, containsIgnoreCase, false)

	containsIgnoreCase = ContainsIgnoreCase("hello", "helloo")
	assert.Equal(t, containsIgnoreCase, false)

	containsIgnoreCase = ContainsIgnoreCase("Hella Hot Hot Sauce", "hello")
	assert.Equal(t, containsIgnoreCase, false)
}

func TestGenerateRandomCpeMac(t *testing.T) {
	mac := GenerateRandomCpeMac()
	assert.Equal(t, len(mac), 12)
	// Check all characters are uppercase hex digits
	for _, c := range mac {
		assert.Assert(t, (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F'))
	}
}

func TestValidatePokeQuery(t *testing.T) {
	// Valid query with "doc" parameter
	values := url.Values{}
	values.Set("doc", "telemetry")
	doc, err := ValidatePokeQuery(values)
	assert.NilError(t, err)
	assert.Equal(t, doc, "telemetry")

	// No "doc" parameter returns default "primary"
	emptyValues := url.Values{}
	doc, err = ValidatePokeQuery(emptyValues)
	assert.NilError(t, err)
	assert.Equal(t, doc, "primary")
}

func TestGetEcmMacAddress(t *testing.T) {
	// Valid MAC address (no colons) - subtracts 2 from hex value
	ecmMac := GetEcmMacAddress("AABBCCDDEEFF")
	assert.Equal(t, ecmMac, "AABBCCDDEEFD")

	// Another valid MAC
	ecmMac = GetEcmMacAddress("112233445567")
	assert.Equal(t, ecmMac, "112233445565")
}

func TestRemoveOneElementFromList(t *testing.T) {
	list := []string{"apple", "banana", "cherry"}
	result := RemoveOneElementFromList(list, "banana")
	assert.Equal(t, len(result), 2)
	assert.Equal(t, result[0], "apple")
	assert.Equal(t, result[1], "cherry")

	// Element not in list
	result = RemoveOneElementFromList(list, "grape")
	assert.Equal(t, len(result), 3)
}

func TestStringSliceContains(t *testing.T) {
	// Sorted slice
	sortedList := []string{"apple", "banana", "cherry", "date"}
	assert.Equal(t, StringSliceContains(sortedList, "banana"), true)
	assert.Equal(t, StringSliceContains(sortedList, "grape"), false)

	// Empty slice
	emptyList := []string{}
	assert.Equal(t, StringSliceContains(emptyList, "apple"), false)
}
