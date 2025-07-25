package util

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

const (
	TelemetryTwoProfileJSONSchema = `{
  "$schema": "http://json-schema.org/draft-06/schema#",
  "$id": "https://github.com/cfry002/telemetry2/schemas/t2_reportProfileSchema.schema.json",
  "title":"Telemetry 2.0 Report Profile Description",
  "version": "2.0.10",
  "type": "object",
  "description": "<b>Telemetry 2.0 Report Profile Description</b><p>A Telemetry 2.0 Report Profile is a configuration, authored in JSON, that can be sent to any RDK device which supports Telemetry 2.0.  A Report Profile contains properties that are interpreted by the CPE in order to generate and upload a telemetry report. These properties define the details of a generated report, including: <br/>&bull;&nbsp;&nbsp;Scheduling (how often the report should be generated) <br/>&bull;&nbsp;&nbsp;Parameters (what key/value pairs should be in the report) <br/>&bull;&nbsp;&nbsp;Encoding (the format of the generated report) <br/>&bull;&nbsp;&nbsp;Protocol (protocol to use to send generated report)</p>",

  "definitions": {
    "parmUse": {
        "type":  "string",
        "enum": [ "count", "absolute", "csv", "accumulate" ],
        "default": "absolute",
        "example":"\"use\":\"count\""
    },
    "dataModelMethod": {
        "type":  "string",
        "enum": [ "get", "subscribe" ],
        "default": "get",
        "example":"\"method\":\"subscribe\""
    },
    "parmReportTimestamp":{
        "type":"string",
        "enum": ["Unix-Epoch","None"],
        "default": "None",
        "example":"\"reportTimestamp\":\"Unix-Epoch\""
        
    },
    "parmDefinitions": {
        "title":"Definitions for the different Parameter types",
        "type":"object",
        "properties":{
            "grep": {
                "title":"\"grep\" Parameter",
                "type": "object",
                "properties": {
                    "type":    { "type": "string", "const": "grep", "description": "Defines a grep parameter"  },
                    "marker":  { "type": "string", "description": "The key name to be used for this data in the generated report." },
                    "search":  { "type": "string", "description": "The string for which to search within the log file." },
                    "logFile": { "type": "string", "description": "The name of the log file to be searched."},
                    "use":     { "$ref":    "#/definitions/parmUse", "description": "This property indicates how the data for this parameter should be gathered and reported.<br/>&bull;&nbsp;&nbsp;\"count\": Indicates that the value to report for this parameter is the number of times it has occurred during the reporting interval..<br/>&bull;&nbsp;&nbsp;\"absolute\": Indicates that the value to report for this parameter is the last actual value received, in the case of events, or found in the log file, in the case of greps.<br/>&bull;&nbsp;&nbsp;\"csv\": Indicates that the value to report for this parameter is a comma separated list of all the actual values received, in the case of events, or found in the log file, in the case of greps. <b>NOTE:</b> \"csv\" is not currently supported in Telemetry 2.0."},
                    "reportEmpty": { "type": "boolean", "default":"false", "description": "Should this marker name be included in the generated report even if the search string was not found in the log file?"},
                    "firstSeekFromEOF": { "type": "integer", "description": "An offset, in bytes, backward from logFile EOF to where the grep for search must begin for the first report this profile generates.  See documentation."}
                },
                "additionalProperties": false,
                "required": ["type", "marker", "search", "logFile"],
                "description": "A grep parameter defines report data that comes from searching a log file for a particular string.",
                "example":"{ \"type\": \"grep\", \"marker\": \"T2_btime_dsLock_split\", \"search\": \"Downstream Lock Success=\", \"logFile\": \"BootTime.log\", \"use\": \"absolute\"}"
            },
            "event": {
                "title":"\"event\" Parameter",
                "type": "object",
                "properties": {
                    "type":      { "type": "string", "const": "event", "description": "Defines an event parameter.  This data comes from a component that has been instrumented to send its telemetry data via Telemetry 2 APIs."  },
                    "name":      { "type": "string", "description": "Optional: The key name to be used for this data in the generated report." },
                    "eventName":  { "type": "string", "description": "The event name by which the component will report this data to Telemetry 2" },
                    "component":  { "type": "string", "description": "The name of the component from which this data should be expected.  Telemetry 2 will use this name to register its interest with the component." },
                    "use":        { "$ref": "#/definitions/parmUse", "description": "This property indicates how the data for this parameter should be gathered and reported.<br/>&bull;&nbsp;&nbsp;\"count\": Indicates that the value to report for this parameter is the number of times it has occurred during the reporting interval..<br/>&bull;&nbsp;&nbsp;\"absolute\": Indicates that the value to report for this parameter is the last actual value received, in the case of events, or found in the log file, in the case of greps.<br/>&bull;&nbsp;&nbsp;\"csv\": Indicates that the value to report for this parameter is a comma separated list of all the actual values received, in the case of events, or found in the log file, in the case of greps. <b>NOTE:</b> \"csv\" is not currently supported in Telemetry 2.0."},
                    "reportEmpty": { "type": "boolean", "default":"false", "description": "Should this marker name be included in the generated report even if the search string was not found in the log file?"},
                    "reportTimestamp": { "$ref": "#/definitions/parmReportTimestamp", "description": "This property indicates whether or not a timestamp should be encoded to indicate the time at which this parameter data was received."}
                },
                "additionalProperties": false,
                "required": ["type", "eventName", "component"],
                "description": "An event parameter defines data that will come from a component event.",
                "example": "{ \"type\": \"event\", \"name\": \"XH_RSSI_1_split\", \"eventName\":\"xh_rssi_3_split\",\"component\": \"ccsp-wifi-agent\", \"use\":\"absolute\" }, "
            },
            "dataModel": {
                "title":"\"dataModel\" Parameter",
                "type": "object",
                "properties": {
                    "type":      { "type": "string", "const": "dataModel", "description": "Defines a data model parameter, e.g., TR-181 data."  },
                    "name":      { "type": "string", "description": "Optional: The key name to be used for this data in the generated report." },
                    "reference":  { "type": "string", "description": "The data model object or property name whose value is to be in the generated report, e.g., \"Device.DeviceInfo.HardwareVersion\"" },
                    "reportEmpty": { "type": "boolean", "default":"false", "description": "Should this marker name be included in the generated report even if the search string was not found in the log file?"},
                    "method":     { "$ref": "#/definitions/dataModelMethod", "description": "How should the value for this parameter be retrieved?  \"get\" will do a parameter GET; \"subscribe\" will subscribe for changes on the event indicated by \"reference\"." },
                    "use":        { "$ref": "#/definitions/parmUse", "description": "This property indicates how the data for this parameter should be gathered and reported.<br/>&bull;&nbsp;&nbsp;\"count\": Indicates that the value to report for this parameter is the number of times it has occurred during the reporting interval..<br/>&bull;&nbsp;&nbsp;\"absolute\": Indicates that the value to report for this parameter is the last actual value received.<br/>&bull;&nbsp;&nbsp;\"csv\": Indicates that the value to report for this parameter is a comma separated list of all the actual values received. <b>NOTE:</b> \"csv\" is not currently supported in Telemetry 2.0."},
                    "reportTimestamp": { "$ref": "#/definitions/parmReportTimestamp", "description": "This property indicates whether or not a timestamp should be encoded to indicate the time at which this parameter data was received."}
                },
                "additionalProperties": false,
                "description":"A dataModel parameter defines data that will come from the CPE data model, e.g., TR-181",
                "example":"",
                "required": ["type", "reference"],
                "anyOf": [
                    {
                        "properties": {
                           "reportTimestamp": { "const": "Unix-Epoch" },
                           "method":{ "const":"subscribe" }
                        },
                        "required": ["method"]
                    },
                    {
                        "properties": {
                           "reportTimestamp": { "const": "None" }
                        }
                    }                  
                ]
            }   
        }     
    },
    "protocolDefinitions": {
        "title":"Definitions for the supported Protocols",
        "type":"object",
        "properties": {
            "HTTP": { 
                "title":"HTTP Definition",
                "type": "object",
                "properties": {
                    "URL": {"type":"string", "description": "The URL to which the generated report should be uploaded."},
                    "Compression": {"type":"string", "enum": ["None"], "description": "Compression scheme to be used in the generated report. <b>NOTE:</b> Only \"None\" is currently supported in Telemetry 2.0."},
                    "Method": {"type":"string", "enum":["POST", "PUT"], "description": "HTTP method to be used to upload the generated report. <b>NOTE:</b> Only \"POST\" is currently supported in Telemetry 2.0."},
                    "RequestURIParameter": {
                        "title":"RequestURIParameter",
                        "type":"array",
                        "items": {
                            "type":"object",
                            "properties": {
                                "Name": {"type":"string", "description": "Value to be used as the Name in the query parameter name/value pair."} ,
                                "Reference": {"type":"string", "description": "Value to be used as the Value in the query parameter name/value pair.  Must be a data model reference."}
                            },
                            "required": ["Name", "Reference"]
                        }, 
                        "description": "Optional: Query parameters to be included in the report's upload HTTP URL."
                    }
                },
                "additionalProperties": false,
                "required": ["URL", "Compression", "Method"],
                "description": "HTTP Protocol details that will be used when Protocol=\"HTTP\"."
            },
            "RBUS_METHOD": {
                "title": "RBUS_METHOD Definition",
                "type": "object",
                "properties": {
                    "Method": {"type":"string", "description": "The name of the method to invoke via rbus. This is the method name that the provider component registers with rbus."},
                    "Parameters": {
                        "title":"Parameters to send to the provider method",
                        "type":"array",
                        "items": {
                            "type":"object",
                            "properties": {
                                "name": {"type":"string", "description": "Value to be used as the name in the method input parameter name/value pair."} ,
                                "value": {"type":"string", "description": "Value to be used as the value in the method input parameter name/value pair."}
                            },
                            "required": ["name", "value"]

                        }
                    }
                },
                "additionalProperties": false,
                "required": ["Method", "Parameters"],
                "description": "RBUS_METHOD Protocol details that will be used when Protocol=\"RBUS_METHOD\".  These method and parameters to pass to RBUS for transport of generated report."
            }
        }
    },
    "encodingDefinitions": {
        "title":"Definitions for the supported Encoding types",
        "type":"object",
        "properties": {
            "JSONEncodingDefinition": { 
                "title":"JSONEncoding Definition",
                "type": "object",
                "properties": {
                    "ReportFormat": {"type":"string", "enum": ["NameValuePair"], "description":"JSON Format to be used for JSON encoding in the generated report. <b>NOTE:</b> Only \"NameValuePair\" is currently supported in Telemetry 2.0."},
                    "ReportTimestamp": {"type":"string","enum": ["None"], "description":"Timestamp format to be used in generated report. <b>NOTE:</b> Only \"None\" is currently supported in Telemetry 2.0."}
                },
                "required": ["ReportFormat", "ReportTimestamp"],
                "description": "JSON Encoding details that will be used when EncodingType=\"JSON\"."
            }
        }
    },
    "conditionDefinitions": {
      "type":"object",
      "properties":{
          "dataModel": {
              "type": "object",
              "properties": {
                  "type":      { "type": "string", "enum": [ "dataModel" ]  },
                  "operator":  { "type": "string", "enum": [ "any", "lt", "gt", "eq" ]  },
                  "threshold": {"type":"integer"},
                  "minThresholdDuration": {"type":"integer"},
                  "reference":  { "type": "string" },
                  "report":     { "type":"boolean", "default":"true", "description": "Should this dataModel event be included in the generated report if this condition caused report generation?"}
              },
              "additionalProperties": false,
              "required": ["type", "operator", "reference"]
              }   
          }     
    },
    "ReportingAdjustments": {
        "type":"object",
        "properties":{
            "ReportOnUpdate": {"title":"ReportOnUpdate","type":"boolean", "default": false, "description": "Indicates if a report should be generated and sent before a new version of this profile is activated."},
            "FirstReportingInterval": {"title":"FirstReportingInterval","type":"integer", "default": 0, "description": "The number of seconds to wait after profile activation before generating a report once."},
            "MaxUploadLatency": {"title":"ReportOnUpdate","type":"integer", "default": 0, "description": "If present, this value is used to randomize the upload time of a generated report. Only valid when TimeReference is not equal to the default value."}
        },
        "additionalProperties": false,
        "minProperties": 1
    }     
    },




    "properties": {
        "Description": {
            "type":"string", 
            "title":"Description",
            "description":"Text describing the purpose of this Report Profile."
        },
        "Version":  { 
            "title":"Version",
            "type":"string", 
            "description":"Version of the profile. This value is opaque to the Telemetry 2 component, but can be used by server processing to indicate specifics about data available in the generated report."
        },
        "Protocol": { 
            "title":"Protocol",
            "type":"string", 
            "enum":["HTTP", "RBUS_METHOD"], 
            "description":"The protocol to be used for the upload of report generated by this profile." 
        },
        "EncodingType": { "title":"EncodingType","type":"string", "enum":["JSON"], "description": "The encoding type to be used in the report generated by this profile." },
        "ReportingInterval": {"title":"ReportingInterval","type":"integer", "description": "The interval, in seconds, at which this profile shall cause a report to be generated."},
        "ActivationTimeOut": {"title":"ActivationTimeOut","type":"integer", "description": "The amount of time, in seconds, that this profile shall remain active on the device.  This is the amount of time from which the profile is received until the CPE will consider the profile to be disabled. After this time, no further reports will be generated for this report."},
        "DeleteOnTimeOut": {"title":"DeleteOnTimeout","type":"boolean", "default":false, "description": "Indicates whether this profile should be removed from memory when the ActivationTimeOut is reached."},
        "TimeReference": {"title":"TimeReference", "type":"string", "default":"0001-01-01T00:00:00Z", "description": "An absolute time reference in UTC that indicates when a report shall be sent."},
        "GenerateNow": {"title":"GenerateNow","type":"boolean", "default": false, "description": "When true, indicates that the report for this Report Profile should be generated immediately upon receipt of the profile."},
        "Parameter": { 
            "title":"Parameter",
            "type":"array",
            "maxItems":800, 
            "items": {
                "type":"object",
                "title":"Parameter items",
                "oneOf": [
                    { "$ref": "#/definitions/parmDefinitions/properties/grep", "title": "grep" },
                    { "$ref": "#/definitions/parmDefinitions/properties/event", "title": "event" },
                    { "$ref": "#/definitions/parmDefinitions/properties/dataModel", "title": "dataModel" }
                    ]
            },
            "description": "An array of objects which defines the data to be included in the generated report. Each object defines the type of data, the source of the data and an optional name to be used as the name (marker) for this data in the generated report. "
        },
        "TriggerCondition": {
            "type":"array",
            "maxItems":50,
            "items": 
                    { "$ref": "#/definitions/conditionDefinitions/properties/dataModel"}
        },
        "HTTP": { "$ref": "#/definitions/protocolDefinitions/properties/HTTP"},
        "RBUS_METHOD": { "$ref": "#/definitions/protocolDefinitions/properties/RBUS_METHOD"},
        "JSONEncoding": {"$ref": "#/definitions/encodingDefinitions/properties/JSONEncodingDefinition"},
        "RootName": {  "title":"RootName","type":"string","default":"Report","description": "The name to be used for the root of the JSON report, e.g., \"searchResult\"." },
        "ReportingAdjustments": { "$ref":"#/definitions/ReportingAdjustments", "title": "ReportingAdjustments"}
    },
    "required": ["Protocol", "EncodingType","Parameter"],
    "additionalProperties": false,
    "anyOf": [
        {
            "properties": {
            "Protocol": { "const": "HTTP" }
            },
            "required": ["HTTP"]
        },
        {
            "properties": {
            "Protocol": { "const": "RBUS_METHOD" }
            },
            "required": ["RBUS_METHOD"]
        }
    ],
    "dependencies": {
        "DeleteOnTimeOut": { "required": ["ActivationTimeOut"] }
      }
}`
)

var TelemetryTwoProfileSchema *gojsonschema.Schema

func init() {
	var err error

	schemaLoader := gojsonschema.NewSchemaLoader()
	schemaLoader.Draft = gojsonschema.Draft6
	schemaLoader.Validate = true

	// TODO - figure out how to get the schema from the file system that will work with unit tests
	// refLoader := gojsonschema.NewReferenceLoader("file://xconf/schema/telemetrytwoprofile-schema.json")
	refLoader := gojsonschema.NewStringLoader(TelemetryTwoProfileJSONSchema)

	TelemetryTwoProfileSchema, err = schemaLoader.Compile(refLoader)
	if err != nil {
		panic(fmt.Errorf("fatal error loads and compiles JSON schema: %+v", err))
	}
}

// ValidateTelemetryTwoProfileJson validates JSON against the schema
func ValidateTelemetryTwoProfileJson(json string) error {
	jsonLoader := gojsonschema.NewStringLoader(json)
	result, err := TelemetryTwoProfileSchema.Validate(jsonLoader)
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	} else {
		var errList []string
		for _, err := range result.Errors() {
			errList = append(errList, err.String())
		}
		log.Errorf("Invalid Telemetry 2.0 Profile JSON config data: %s", errors.New(strings.Join(errList, ". ")))

		return xwcommon.NewRemoteError(http.StatusBadRequest, "Please provide the valid Telemetry 2.0 Profile JSON config data.")
	}
}
