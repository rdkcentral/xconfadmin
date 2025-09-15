package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestValidateTelemetryTwoProfileJson(t *testing.T) {
	jsonData := `{
		"Description": "jw t2 wifi test part 1",
		"Version": "35",
		"Protocol": "HTTP",
		"EncodingType": "JSON",
		"ReportingInterval": 120,
		"TimeReference": "0001-01-01T00:00:00Z",
		"Parameter": [{
			"type": "dataModel",
			"reference": "Profile.Version"
		}],
		"HTTP": {
			"URL": "https://test.url.com/",
			"Compression": "None",
			"Method": "POST",
			"RequestURIParameter": [{
				"Name": "profileName",
				"Reference": "Profile.Name"
			}, {
				"Name": "profileDescription",
				"Reference": "Profile.Description"
			}, {
				"Name": "reportVersion",
				"Reference": "Profile.Version"
			}]
		},
		"JSONEncoding": {
			"ReportFormat": "NameValuePair",
			"ReportTimestamp": "None"
		}
	}`

	err := ValidateTelemetryTwoProfileJson(jsonData)
	assert.NilError(t, err)

	jsonData = `{
		"ActivationTimeout": 600,
		"Protocol": "FTP",
		"Description": "Telemetry 2.0 HSD Gateway WiFi Radio",
		"JSONEncoding": {
		  "ReportTimestamp": "None",
		  "ReportFormat": "NameValuePair"
		},
		"ReportingInterval": 60,
		"Version": "0.1",
		"HTTP": {
		  "URL": "https://test.net/",
		  "RequestURIParameter": [
			{
			  "Name": "profileName",
			  "Reference": "Profile.Name"
			},
			{
			  "Name": "reportVersion",
			  "Reference": "Profile.Version"
			}
		  ],
		  "Compression": "None",
		  "Method": "POST"
		},
		"EncodingType": "JSON",
		"TimeReference": "0001-01-01T00:00:00Z"
	  }`

	err = ValidateTelemetryTwoProfileJson(jsonData)
	assert.ErrorContains(t, err, "Please provide the valid Telemetry 2.0 Profile JSON config data.")
}
