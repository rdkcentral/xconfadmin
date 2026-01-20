# XConf Admin REST API Documentation

## Overview

The XConf Admin API provides comprehensive configuration management for RDK devices through a RESTful interface. It manages firmware configurations, device settings, telemetry profiles, feature rules, and various configuration management operations.

**Base URL**: `/xconfAdminService`

**Authentication**: Most endpoints require authentication via SAT token as a Bearer Token in the Authorization header.

---

## API Overview

### Configuration Management
1. [Firmware Config](#firmware-config)
2. [IP Rules](#ip-rules)
3. [Location Filter](#location-filter)
4. [Download Location Filter](#download-location-filter)
5. [Environment Model Rules](#environment-model-rules)
6. [IP Filter](#ip-filter)
7. [Percent Filter](#percent-filter)
8. [Time Filter](#time-filter)
9. [RebootImmediately Filter](#rebootimmediately-filter)

### Entity Management
10. [Environment](#environment)
11. [IP Address Group](#ip-address-group)
12. [Model](#model)
13. [NamespacedList](#namespacedlist)
14. [Mac Rule](#mac-rule)

### Rule and Template Management
15. [FirmwareRuleTemplate](#firmwareruletemplate)
16. [FirmwareRule](#firmwarerule)
17. [Feature](#feature)
18. [Feature Rule](#feature-rule)
19. [Activation Minimum Version](#activation-minimum-version)

### Telemetry Management
20. [Telemetry Profile](#telemetry-profile)
21. [Telemetry Profile 2.0](#telemetry-profile-20)
22. [Telemetry 2.0 Profile Json Schema](#telemetry-20-profile-json-schema)

### Change Management
23. [Change API](#change-api)
24. [Change v2 API](#change-v2-api)

### Device Configuration Management
25. [DCM (Device Configuration Management)](#dcm-device-configuration-management)

---

## Firmware Config

### Retrieve a list of firmware configs

**GET** `http://<host>:<port>/queries/firmwares?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType parameter is not required, default value is stb

**Response:** 200 OK OR 400 BAD REQUEST

**Request Example:**
```
http://localhost:9091/queries/firmwares
```

**JSON Response:**
```json
{
  "id": "firmwareConfigId",
  "description": "FirmwareDescription",
  "supportedModelIds": [
    "MODELA"
  ],
  "firmwareFilename": "FirmwareFilename",
  "firmwareVersion": "FirmwareVersion",
  "properties": {
    "testKey": "testValue"
  } 
}
```

### Retrieve a single firmware config by id

**GET** `http://<host>:<port>/queries/firmwares/<firmwareConfigId>`

**Headers:**
- Accept = application/json

**Response:** 200 OK

**Request Example:**
```
http://localhost:9091/queries/firmwares/b65962b5-1481-4eed-a010-2abfa8c3bbfd
```

**JSON Response:**
```json
{
  "id": "b65962b5-1481-4eed-a010-2abfa8c3bbfd",
  "updated": 1440492963476,
  "description": "_-",
  "supportedModelIds": [
    "YETST"
  ],
  "firmwareDownloadProtocol": "tftp",
  "firmwareFilename": "_-",
  "firmwareVersion": "_-",
  "rebootImmediately": false,
  "properties": {
    "testKey": "testValue"
  } 
}
```

### Retrieve firmware configs by modelId

**GET** `http://<host>:<port>/queries/firmwares/model/{modelId}?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType parameter is not required, default value is stb

**Response:** 200 OK OR 400 BAD REQUEST (if application type has a wrong value)

**Request Example:**
```
http://localhost:9091/queries/firmwares/model/YETST
```

**JSON Response:**
```json
[{
  "id": "b65962b5-1481-4eed-a010-2abfa8c3bbfd",
  "updated": 1440492963476,
  "description": "_-",
  "supportedModelIds": [
    "YETST"
  ],
  "firmwareDownloadProtocol": "tftp",
  "firmwareFilename": "_-",
  "firmwareVersion": "_-",
  "rebootImmediately": false,
  "properties": {
    "testKey": "testValue"
  }  
}]
```

### Create/update a firmware config

If firmware config is missing it will be created, otherwise updated. For update operation id field is not needed.

**POST** `http://<host>:<port>/updates/firmwares?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK and saved object; 400 BAD REQUEST (by validation error); 500 INTERNAL SERVER ERROR

**Restrictions:**
- Description, file name, version, supported model should be not empty

**Request Example:**
```
http://localhost:9091/updates/firmwares
```

**JSON Request:**
```json
{
  "id": "b65962b5-1481-4eed-a010-2abfa8c3bbfd",
  "updated": 1440492963476,
  "description": "_-",
  "supportedModelIds": [
    "YETST"
  ],
  "firmwareDownloadProtocol": "tftp",
  "firmwareFilename": "_-",
  "firmwareVersion": "_-",
  "rebootImmediately": false,
  "properties": {
    "testKey": "testValue"
  }  
}
```

### Delete a firmware config by id

**DELETE** `http://<host>:<port>/delete/firmwares/<firmwareConfigId>`

**Headers:**
- Accept = application/json

**Response:** 204 NO CONTENT and text message: Firmware config successfully deleted OR Config doesn't exist.

**Request Example:**
```
http://localhost:9091/delete/firmwares/b65962b5-1481-4eed-a010-2abfa8c3bbfd
```

---

## IP rules

### Retrieve an ip rule list

**GET** `http://<host>:<port>/queries/rules/ips?applicationType={type}`

**Headers:**
- Accept = application/json
- Default value for applicationType parameter is stb

**Response:** 200 OK OR 400 BAD REQUEST

**Request Example:**
```
http://localhost:9091/queries/rules/ips
```

**JSON Response:**
```json
[
  {
    "id": "ddc07355-d253-4f6b-8b42-296819d0d094",
    "name": "fsd",
    "ipAddressGroup": {
      "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
      "name": "test",
      "ipAddresses": [
        "192.11.11.11"
      ]
    },
    "environmentId": "DEV",
    "modelId": "YETST",
    "noop": true,
    "expression": {
      "targetedModelIds": [],
      "environmentId": "DEV",
      "modelId": "YETST",
      "ipAddressGroup": {
        "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
        "name": "test",
        "ipAddresses": [
          "192.11.11.11"
        ]
      }
    }
  }
]
```

### Retrieve an ip rule by name

**GET** `http://<host>:<port>/queries/rules/ips/{ipRuleName}?applicationType={type}`

**Headers:**
- Accept = application/json
- Default value for applicationType parameter is stb

**Response:** 200 OK OR 400 BAD REQUEST

**Request Example:**
```
http://localhost:9091/queries/rules/ips/fsd
```

**JSON Response:**
```json
{
  "id": "ddc07355-d253-4f6b-8b42-296819d0d094",
  "name": "fsd",
  "ipAddressGroup": {
    "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
    "name": "test",
    "ipAddresses": [
      "192.11.11.11"
    ]
  },
  "environmentId": "DEV",
  "modelId": "YETST",
  "noop": true,
  "expression": {
    "targetedModelIds": [],
    "environmentId": "DEV",
    "modelId": "YETST",
    "ipAddressGroup": {
      "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
      "name": "test",
      "ipAddresses": [
        "192.11.11.11"
      ]
    }
  }
}
```

### Create/update an ip rule

If IpRule is missing it will be created, otherwise updated. For update operation id field is not needed.

**POST** `http://<host>:<port>/updates/rules/ips?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json

**Response:** 200 OK and saved object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Name, environmentId, modelId should be not empty
- IP address group should be specified

**Request Example:**
```
http://localhost:9091/updates/rules/ips
```

**JSON Request:**
```json
{
  "id": "ddc07355-d253-4f6b-8b42-296819d0d094",
  "name": "fsd",
  "ipAddressGroup": {
    "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
    "name": "test",
    "ipAddresses": [
      "192.11.11.11"
    ]
  },
  "environmentId": "DEV",
  "modelId": "YETST",
  "noop": true,
  "expression": {
    "targetedModelIds": [],
    "environmentId": "DEV",
    "modelId": "YETST",
    "ipAddressGroup": {
      "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
      "name": "test",
      "ipAddresses": [
        "192.11.11.11"
      ]
    }
  }
}
```

### Delete an ip rule

**DELETE** `http://<host>:<port>/delete/rules/ips/{ipRuleName}?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType parameter is not required, default value is stb

**Response:** 204 NO CONTENT and message: IpRule successfully deleted OR Rule doesn't exists; 400 BAD REQUEST if applicationType is not valid

**Request Example:**
```
http://localhost:9091/delete/rules/ips/ruleName
```

---

## Location filter

### Retrieve a location filter list

**GET** `http://<host>:<port>/queries/filters/locations?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType parameter is not required, default value is stb

**Response:** 200 OK OR 400 BAD REQUEST if applicationType is not valid

**Request Example:**
```
http://localhost:9091/queries/filters/locations
```

**JSON Response:**
```json
[
  {
    "ipAddressGroup": {
      "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
      "name": "test",
      "ipAddresses": [
        "192.11.11.11"
      ]
    },
    "environments": [],
    "models": [],
    "ipv6FirmwareLocation": "2001:0db8:11a3:09d7:1f34:8a2e:07a0:765d",
    "httpLocation": "http://localhost:8080",
    "forceHttp": true,
    "id": "2ce1279b-bb25-4fda-9a34-fe8466bc2702",
    "name": "name",
    "boundConfigId": "95e75859-ae8f-4d6a-b758-11fefbe647e1",
    "ipv4FirmwareLocation": "10.10.10.10"
  }
]
```

### Retrieve a location filter by name

**GET** `http://<host>:<port>/queries/filters/locations/{locationFilterName}?applicationType={type}`

Or legacy endpoint:
**GET** `http://<host>:<port>/queries/filters/locations/byName/{locationFilterName}`

**Headers:**
- Accept = application/json

**Response:** 200 OK

**Request Example:**
```
http://localhost:9091/queries/filters/locations/name
```

**JSON Response:**
```json
[
  {
    "ipAddressGroup": {
      "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
      "name": "test",
      "ipAddresses": [
        "192.11.11.11"
      ]
    },
    "environments": [],
    "models": [],
    "ipv6FirmwareLocation": "2001:0db8:11a3:09d7:1f34:8a2e:07a0:765d",
    "httpLocation": "http://localhost:8080",
    "forceHttp": true,
    "id": "2ce1279b-bb25-4fda-9a34-fe8466bc2702",
    "name": "name",
    "boundConfigId": "95e75859-ae8f-4d6a-b758-11fefbe647e1",
    "ipv4FirmwareLocation": "10.10.10.10"
  }
]
```

### Create/update location filter

If location filter is missing it will be created, otherwise updated. For update operation id field is not needed.

**POST** `http://<host>:<port>/updates/filters/locations?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType parameter is not required, default value is stb

**Response:** 200 OK and saved object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Condition, models, environments, IPv4, location, any location (HTTP or firmware), IPv4/IPv6 should be valid

**Request Example:**
```
http://localhost:9091/updates/filters/locations
```

**JSON Request:**
```json
{
  "ipAddressGroup": {
    "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
    "name": "test",
    "ipAddresses": [
      "192.11.11.11"
    ]
  },
  "environments": [],
  "models": [],
  "ipv6FirmwareLocation": "2001:0db8:11a3:09d7:1f34:8a2e:07a0:765d",
  "httpLocation": "http://localhost:8080",
  "forceHttp": true,
  "id": "2ce1279b-bb25-4fda-9a34-fe8466bc2702",
  "name": "name",
  "boundConfigId": "95e75859-ae8f-4d6a-b758-11fefbe647e1",
  "ipv4FirmwareLocation": "10.10.10.10"
}
```

### Delete location filter by name

**DELETE** `http://<host>:<port>/delete/filters/locations/{locationFilterName}?applicationType={type}`

**Headers:**
- Accept = application/json

**Response:** 204 NO CONTENT and message: Location filter successfully deleted OR Filter doesn't exist with name: <filterName>; 400 BAD REQUEST if applicationType is not valid

**Request Example:**
```
http://localhost:9091/delete/filters/location/name
```

---

## Download location filter

### Retrieve download location filter

**GET** `http://<host>:<port>/queries/filters/downloadlocation?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST if applicationType is not valid

**Request Example:**
```
http://localhost:9091/queries/filters/downloadlocation
```

**JSON Response:**
```json
{
  "type": "com.comcast.xconf.estbfirmware.DownloadLocationRoundRobinFilterValue",
  "id": "DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
  "locations": [
    {
      "locationIp": "10.10.10.10",
      "percentage": 100.0
    }
  ],
  "ipv6locations": [],
  "rogueModels": [],
  "httpLocation": "lf.com",
  "httpFullUrlLocation": "http://www.localhost.org",
  "neverUseHttp": true,
  "firmwareVersions": "??"
}
```

### Update download location filter

**POST** `http://<host>:<port>/updates/filters/downloadlocation?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK and saved object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Location URL, IPv4/IPv6 should be valid
- Percentage should be positive and within [0, 100]
- Locations should be not duplicated

**Request Example:**
```
http://localhost:9091/updates/filters/downloadlocation
```

**JSON Request:**
```json
{
  "type": "com.comcast.xconf.estbfirmware.DownloadLocationRoundRobinFilterValue",
  "id": "DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE",
  "updated": 1441287139144,
  "locations": [
    {
      "locationIp": "10.10.10.10",
      "percentage": 100.0
    }
  ],
  "ipv6locations": [],
  "rogueModels": [],
  "httpLocation": "lf.com",
  "httpFullUrlLocation": "http://www.localhost.org",
  "neverUseHttp": true,
  "firmwareVersions": "??"
}
```

---

## Environment model rules

### Retrieve an environment model rule list

**GET** `http://<host>:<port>/queries/rules/envModels?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST if applicationType is not valid

**Request Example:**
```
http://localhost:9091/queries/rules/envModels
```

**JSON Response:**
```json
[
  {
    "id": "12b620bd-2e74-4467-91e5-c29657022c05",
    "name": "re",
    "firmwareConfig": {
      "id": "f0b7b35b-4b8e-4a15-9d66-91c4b3d575d1",
      "description": "prav_Firm",
      "supportedModelIds": [
        "PX013ANM",
        "PX013ANC"
      ],
      "firmwareFilename": "PX013AN_2.1s11_VBN_HYBse-signed.bin",
      "firmwareVersion": "PX013AN_2.1s11_VBN_HYBse-signed"
    },
    "environmentId": "TEST",
    "modelId": "PX013ANC"
  }
]
```

### Retrieve an environment model rule by name

**GET** `http://<host>:<port>/queries/rules/envModels/{envModelRuleName}?applicationType={type}`

**Headers:**
- Accept = application/json

**Response:** 200 OK; 400 BAD REQUEST if applicationType is not valid

**Request Example:**
```
http://localhost:9091/queries/rules/envModels/testName
```

**JSON Response:**
```json
{
  "id": "12b620bd-2e74-4467-91e5-c29657022c05",
  "name": "testName",
  "firmwareConfig": {
    "id": "f0b7b35b-4b8e-4a15-9d66-91c4b3d575d1",
    "description": "prav_Firm",
    "supportedModelIds": [
      "PX013ANM",
      "PX013ANC"
    ],
    "firmwareFilename": "PX013AN_2.1s11_VBN_HYBse-signed.bin",
    "firmwareVersion": "PX013AN_2.1s11_VBN_HYBse-signed"
  },
  "environmentId": "TEST",
  "modelId": "PX013ANC"
}
```

### Create/update an environment model rule

If EnvModelRule is missing it will be created, otherwise updated. For update operation id field is not needed.

**POST** `http://<host>:<port>/updates/rules/envModels?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK and saved object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Name, environment, model should be not empty
- Name is used only once
- Environment/model should not overlap each other

**Request Example:**
```
http://localhost:9091/updates/rules/envModels
```

**JSON Request:**
```json
{
  "id": "12b620bd-2e74-4467-91e5-c29657022c05",
  "name": "testName",
  "firmwareConfig": {
    "id": "f0b7b35b-4b8e-4a15-9d66-91c4b3d575d1",
    "description": "prav_Firm",
    "supportedModelIds": [
      "PX013ANM",
      "PX013ANC"
    ],
    "firmwareFilename": "PX013AN_2.1s11_VBN_HYBse-signed.bin",
    "firmwareVersion": "PX013AN_2.1s11_VBN_HYBse-signed"
  },
  "environmentId": "TEST",
  "modelId": "PX013ANC"
}
```

### Delete an environment model rule

**DELETE** `http://<host>:<port>/delete/rules/envModels/{envModelRuleName}?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 204 NO CONTENT and message: Rule successfully deleted OR Rule doesn't exist with name: <ruleName>; 400 if applicationType is not valid

**Request Example:**
```
http://localhost:9091/delete/rules/envModels/testName
```

---

## IP filter

### Retrieve an IP filter list

**GET** `http://<host>:<port>/queries/filters/ips?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST

**Request Example:**
```
http://localhost:9091/queries/filters/ips
```

**JSON Response:**
```json
[
  {
    "id": "8bdb3493-a18b-4230-9b25-fd44df38863b",
    "name": "name",
    "ipAddressGroup": {
      "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
      "name": "test",
      "ipAddresses": [
        "192.11.11.11"
      ]
    },
    "warehouse": false
  }
]
```

### Retrieve an ip filter by name

**GET** `http://<host>:<port>/queries/filters/ips/{ipFilterName}?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST if applicationType is not valid

**Request Example:**
```
http://localhost:9091/queries/filters/ips/namef
```

**JSON Response:**
```json
{
  "id": "f9c5a6e8-d34f-4dc6-ae41-9016b70552ae",
  "name": "namef",
  "ipAddressGroup": {
    "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
    "name": "test",
    "ipAddresses": [
      "192.11.11.11"
    ]
  },
  "warehouse": false
}
```

### Create/update an IP filter

If IpFilter is missing it will be created, otherwise updated. For update operation id field is not needed.

**POST** `http://<host>:<port>/updates/filters/ips?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK and saved object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Name, IP address group should be not empty

**Request Example:**
```
http://localhost:9091/updates/filters/ips
```

**JSON Request:**
```json
{
  "id": "f9c5a6e8-d34f-4dc6-ae41-9016b70552ae",
  "name": "namef",
  "ipAddressGroup": {
    "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
    "name": "test",
    "ipAddresses": [
      "192.11.11.11"
    ]
  },
  "warehouse": false
}
```

### Delete IP filter

**DELETE** `http://<host>:<port>/delete/filters/ips/{ipFilterName}?applicationType={stb}`

**Headers:**
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 204 NO CONTENT and message: IpFilter successfully deleted OR Filter doesn't exist with name: <filterName>; 400 BAD REQUEST if applicationType is not valid

**Request Example:**
```
http://localhost:9091/delete/filters/ips/namef
```

---

## Percent Filter

### Retrieve percent filter

**GET** `http://<host>:<port>/queries/filters/percent?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST if applicationType is not valid

### Retrieve percent filter field values

**GET** `http://<host>:<port>/queries/filters/percent?field=fieldName&applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK if field exists; 404 Not Found if field does not exist

### Update percent filter

**POST** `http://<host>:<port>/updates/filters/percent?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK and saved object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Percentage should be positive and within [0, 100]

### Retrieve EnvModelPercentages

**GET** `http://<host>:<port>/queries/percentageBean?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST

### Retrieve EnvModelPercentage by id

**GET** `http://<host>:<port>/queries/percentageBean/id`

**Headers:**
- Content-Type = application/json
- Accept = application/json

**Response:** 200 OK OR 404 if envModelPercentage is not found

### Create envModelPercentage

**POST** `http://<host>:<port>/updates/percentageBean?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK; 404 NOT FOUND; 409 CONFLICT; 400 BAD REQUEST

**Restrictions:**
- Name should be unique and not blank
- Environment and model should be not empty
- At least one firmware version should be in minCheck list if firmwareCheckRequired=true
- Percentage within [0, 100]
- Distribution firmware version should be in minCheck list if firmwareCheckRequired=true
- Total distribution percentage is within [0, 100]
- Last known good is not empty if total distribution percentage < 100

### Update EnvModelPercentage

**PUT** `http://<host>:<port>/updates/percentageBean?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK; 404 NOT FOUND; 409 CONFLICT; 400 BAD REQUEST

### Delete envModelPercentage

**DELETE** `http://<host>:<port>/delete/percentageBean/id`

**Headers:**
- Content-Type = application/json
- Accept = application/json

**Response:** 204 NO CONTENT OR 404 NOT FOUND

---

## Time Filter

### Retrieve time filter list

**GET** `http://<host>:<port>/queries/filters/time?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST

### Retrieve time filter by name

**GET** `http://<host>:<port>/queries/filters/time/{name}?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST

### Create/update time filter

If time filter is missing it will be created, otherwise updated. For update operation id field is not needed.

**POST** `http://<host>:<port>/updates/filters/time?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType param is not required, default value is stb

**Response:** 200 OK and saved object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Name should be unique

### Delete time filter by name

**DELETE** `http://<host>:<port>/delete/filters/time/{timeFilterName}?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType is not required, default value is stb

**Response:** 204 NO CONTENT and message: Time Filter successfully deleted OR Filter doesn't exist with name: <filterName>

---

## RebootImmediately Filter

### Retrieve an RI filter list

**GET** `http://<host>:<port>/queries/filters/ri?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST

### Retrieve an RI filter by rule name

**GET** `http://<host>:<port>/queries/filters/ri/{ruleName}?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST

### Create/update an RI filter

If RI filter is missing it will be created, otherwise updated. For update operation id field is not needed.

**POST** `http://<host>:<port>/updates/filters/ri?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType is not required, default value is stb

**Response:** 200 OK; 201 CREATED; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Name should be not empty
- At least one of filter criteria should be specified
- MAC addresses should be valid

### Delete RI filter by name

**DELETE** `http://<host>:<port>/delete/filters/ri/{riFilterName}?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType is not required, default value is stb

**Response:** 204 NO CONTENT and message: Filter does't exist OR Successfully deleted; 400 BAD REQUEST

---

## Environment

### Retrieve an list of environments

**GET** `http://<host>:<port>/queries/environments`

**Headers:**
- Accept = application/json

**Response:** 200 OK

**Request Example:**
```
http://localhost:9091/queries/environments
```

**JSON Response:**
```json
[{"id":"DEV","description":"ff"},{"id":"TEST","description":"do not delete"}]
```

### Retrieve environment by id

**GET** `http://<host>:<port>/queries/environments/<environmentId>`

**Headers:**
- Accept = application/json

**Response:** 200 OK; 400 BAD REQUEST

**Request Example:**
```
http://localhost:9091/queries/environments/DEV
```

### Create an environment

**POST** `http://<host>:<port>/updates/environments`

**Headers:**
- Content-Type: application/json
- Accept = application/json

**Response:** 200 OK and saved object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Environment name should be valid by pattern: ^[a-zA-Z0-9]+$
- Name should be unique

**Request Example:**
```
http://localhost:9091/updates/environments
```

**JSON Request:**
```json
{"id":"testName","description":"some description"}
```

### Delete environment by id

**DELETE** `http://<host>:<port>/delete/environments/<environmentId>`

**Headers:**
- Accept = application/json

**Response:** 204 NO CONTENT and message: Environment doesn't exist OR Environment successfully deleted; 400 BAD REQUEST: Environment is used: <usage place>

**Restrictions:**
- Environment should be not used

---

## IP Address Group

### Retrieve an IP address group list

**GET** `http://<host>:<port>/queries/ipAddressGroups`

**Headers:**
- Accept = application/json

**Response:** 200 OK

**Request Example:**
```
http://localhost:9091/queries/ipAddressGroups
```

**JSON Response:**
```json
[
  {
    "id": "2c184325-f9eb-4edc-85c3-5b6466fc3c5c",
    "name": "test",
    "ipAddresses": [
      "192.11.11.11"
    ]
  }
]
```

### Retrieve an IP address group by name

**GET** `http://<host>:<port>/queries/ipAddressGroups/byName/<ipAddressGroupName>/`

**Headers:**
- Accept = application/json

**Response:** 200 OK; 400 BAD REQUEST

### Retrieve an IP address group by IP

**GET** `http://<host>:<port>/queries/ipAddressGroups/byIp/<ipAddressGroupIp>/`

**Headers:**
- Accept = application/json

**Response:** 200 OK; 400 BAD REQUEST

### Create an IP address group

**POST** `http://<host>:<port>/updates/ipAddressGroups`

**Headers:**
- Content-Type: application/json
- Accept = application/json

**Response:** 200 OK and saved object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Name should be not empty and unique

### Add data to IP Address Group

**POST** `http://<host>:<port>/updates/ipAddressGroups/<ipAddressGroup_name>/addData`

**Headers:**
- Content-Type = application/json
- Accept = application/json

**Response:** 200 OK and ipAddressGroup object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- ipAddressGroup with current id should exist

### Delete data from IP Address Group

**POST** `http://<host>:<port>/updates/ipAddressGroups/<ipAddressGroup_name>/removeData`

**Headers:**
- Content-Type = application/json
- Accept = application/json

**Response:** 204 NO CONTENT and ipAddressGroup object; 400 BAD REQUEST

**Restrictions:**
- List contains IPs which should be present in current IP address group
- IP address group should contain at least one IP address

### Delete an IP address group by id

**DELETE** `http://<host>:<port>/delete/ipAddressGroups/<ipAddressGroupId>`

**Headers:**
- Accept = application/json

**Response:** 204 NO CONTENT and message: IpAddressGroup doesn't exist OR IpAddressGroup successfully deleted; 400 BAD REQUEST: IpAddressGroup is used: <usage place>

**Restrictions:**
- IP address group should be not used

---

## Model

### Retrieve a model list

**GET** `http://<host>:<port>/queries/models`

**Headers:**
- Accept = application/json

**Response:** 200 OK

**Request Example:**
```
http://localhost:9091/queries/models
```

**JSON Response:**
```json
[
  {
    "id": "YETST",
    "description": ""
  },
  {
    "id": "PX013ANC",
    "description": "Pace XG1v3 - Cisco Cable Card"
  }
]
```

### Retrieve model by id

**GET** `http://<host>:<port>/queries/models/<modelId>`

**Headers:**
- Accept = application/json

**Response:** 200 OK; 204 NO CONTENT

### Create model

**POST** `http://<host>:<port>/updates/models`

**Headers:**
- Content-Type = application/json
- Accept = application/json

**Response:** 201 CREATED; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Model name should be unique and valid by pattern: ^[a-zA-Z0-9]+$

### Update model description

**PUT** `http://<host>:<port>/updates/models`

**Headers:**
- Content-Type = application/json
- Accept = application/json

**Response:** 200 OK; 400 BAD REQUEST; 404 NOT FOUND; 500 INTERNAL SERVER ERROR

### Delete model by id

**DELETE** `http://<host>:<port>/delete/models/<modelId>`

**Headers:**
- Accept = application/json

**Response:** 204 NO CONTENT and message: Model deleted successfully; 404 NOT found and message "Model doesn't exist"

**Restrictions:**
- Model should be not used in another places

---

## NamespacedList

### Retrieve all NS lists

**GET** `http://<host>:<port>/queries/nsLists`

**Headers:**
- Accept = application/json

**Response:** 200 OK

**Request Example:**
```
http://localhost:9091/queries/nsLists
```

**JSON Response:**
```json
[
  {
    "id": "macs",
    "data": [
      "AA:AA:AA:AA:AA:AA"
    ]
  }
]
```

### Retrieve NS list by id

**GET** `http://<host>:<port>/queries/nsLists/byId/<nsListId>`

**Headers:**
- Accept = application/json

**Response:** 200 OK

### Retrieve NS list by mac part

**GET** `http://<host>:<port>/queries/nsLists/byMacPart/<macAddressPart>`

**Headers:**
- Accept = application/json

**Response:** 200 OK

### Create a NS list

**POST** `http://<host>:<port>/updates/nsLists`

**Headers:**
- Content-Type = application/json
- Accept = application/json

**Response:** 201 CREATED; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Name should be valid by pattern: ^[a-zA-Z0-9]+$
- List data should be not empty and contain valid mac addresses
- MAC address should be used only in one NS list

### Add data to NS list

**POST** `http://<host>:<port>/updates/nsLists/<nsListId>/addData`

Or legacy endpoint:
**POST** `http://<host>:<port>/updates/nslist/<nsListId>/addData`

**Headers:**
- Content-Type = application/json
- Accept = application/json

**Response:** 200 OK and NS list object; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

### Delete data from NS list

**DELETE** `http://<host>:<port>/updates/nsLists/<nsListId>/removeData`

Or legacy endpoint:
**DELETE** `http://<host>:<port>/updates/nslist/<nsListId>/removeData`

**Headers:**
- Content-Type = application/json
- Accept = application/json

**Response:** 204 NO CONTENT and NS list object; 400 BAD REQUEST

**Restrictions:**
- List contains MACs which should be present in current Namespaced list
- Namespaced list should contain at least one MAC address

### Delete an NS list by id

**DELETE** `http://<host>:<port>/delete/nsLists/<nsListId>`

**Headers:**
- Accept = application/json

**Response:** 200 OK and message: NamespacedList doesn't exist OR NamespacedList successfully deleted

**Restrictions:**
- NS list should be not used in another places

---

## Mac Rule

### Retrieve a mac rule list (legacy)

**GET** `http://<host>:<port>/queries/rules/macs?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST

**Note:** With no version parameter or version < 2. Legacy query. For each macrule returned, if it was created with multiple maclists, only the first one is returned in macListRef.

### Retrieve a mac rule list (v2)

**GET** `http://<host>:<port>/queries/rules/macs?version=2&applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType is not required, default value is stb

**Response:** 200 OK; 400 BAD REQUEST

**Note:** Version parameter could be any number >= 2.

### Retrieve mac rule by name (legacy)

**GET** `http://<host>:<port>/queries/rules/macs/<macRuleName>`

**Headers:**
- Accept = application/json

**Response:** 200 OK

### Retrieve mac rule by name (v2)

**GET** `http://<host>:<port>/queries/rules/macs/<macRuleName>?version=2&applicationType={type}`

**Headers:**
- Accept = application/json

**Response:** 200 OK

### Retrieve mac rule by mac address (legacy)

**GET** `http://<host>:<port>/queries/rules/macs/address/{macAddress}?applicationType={type}`

**Headers:**
- Accept = application/json

**Response:** 200 OK; 400 BAD REQUEST

### Retrieve mac rule by mac address (v2)

**GET** `http://<host>:<port>/queries/rules/macs/address/<macAddress>?version=2&applicationType={type}`

**Headers:**
- Accept = application/json

**Response:** 200 OK; 400 BAD REQUEST

### Create/update mac rule

For create operation id field is optional and the system will generate one in that case. If macrule corresponding to 'id' is missing, a new entry will be created. Otherwise existing entry is completely overwritten with the new parameters provided.

**POST** `http://<host>:<port>/updates/rules/macs?applicationType={type}`

**Headers:**
- Content-Type = application/json
- Accept = application/json
- applicationType is not required, default value is stb

**Response:** 200 OK; 201 CREATED; 400 BAD REQUEST; 500 INTERNAL SERVER ERROR

**Restrictions:**
- Name, mac address list, model list, mac list, firmware configuration should be not empty
- MAC address list is never used in another rule
- Model list contain only existed model
- Firmware config should support given models

### Delete mac rule by name

**DELETE** `http://<host>:<port>/delete/rules/macs/{macRuleName}?applicationType={type}`

**Headers:**
- Accept = application/json
- applicationType is not required, default value is stb

**Response:** 204 NO CONTENT and message: MacRule does'n exist OR MacRule deleted successfully; 400 BAD REQUEST

---

## FirmwareRuleTemplate

### Retrieve filtered templates

**GET** `http://<host>:<port>/firmwareruletemplate/filtered?name=MAC_RULE&key=someKey`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Parameters:**
- Without params: Retrieve all firmware rule templates
- `name`: Filter templates by name
- `key`: Filter by rule key
- `value`: Filter by rule value
- Parameters can be combined: `?name=someName&value=testValue`

**Response Codes:** 200

### Import firmware rule templates

**POST** `http://<host>:<port>/firmwareruletemplate/importAll`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Body:** List of firmware rule templates

**Response Codes:** 200, 400, 404, 409

**Response Body:**
```json
{
    "NOT_IMPORTED": [],
    "IMPORTED": []
}
```

### Create firmware rule template

**POST** `http://<host>:<port>/firmwareruletemplate/?applicationType=stb`

**Headers:**
- Accept = application/json
- Content-Type = application/json
- Authorization = Bearer {SAT token}

**Response Status:** 201 Created

### Update firmware rule template

**POST** `http://<host>:<port>/firmwareruletemplate/?applicationType=stb`

**Headers:**
- Accept = application/json
- Content-Type = application/json
- Authorization = Bearer {SAT token}

**Response Status:** 200 OK

### Delete firmware rule template

**POST** `http://<host>:<port>/firmwareruletemplate/testTemplateName`

**Headers:**
- Content-Type = application/json
- Authorization = Bearer {SAT token}

**Response Status:** 204 No Content

---

## FirmwareRule

### Retrieve all firmware rules

**GET** `http://<host>:<port>/firmwarerule`

**Headers:**
- Accept = application/json

**Response Codes:** 200

### Retrieve filtered firmware rules

**GET** `http://<host>:<port>/firmwarerule/filtered?templateId=TEST_ID&key=firmwareVersion`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Parameters:**
- `applicationType` (required): Filter by application type
- `name`: Filter templates by name
- `key`: Filter by rule key
- `value`: Filter by rule value
- `firmwareVersion`: Filter by firmware version
- `templateId`: Filter by template

**Response Codes:** 200

### Import firmware rule

**POST** `http://<host>:<port>/firmwarerule/importAll`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Body:** List of firmware rules

**Response Codes:** 200, 400, 404, 409

**Response Body:**
```json
{
    "NOT_IMPORTED": [],
    "IMPORTED": ["testName"]
}
```

### Create firmware rule

**POST** `http://<host>:<port>/firmwarerule/?applicationType=stb`

**Headers:**
- Accept = application/json
- Content-Type = application/json
- Authorization = Bearer {SAT token}

**Response Status:** 201 Created

### Update firmware rule

**PUT** `http://<host>:<port>/firmwarerule/?applicationType=stb`

**Headers:**
- Accept = application/json
- Content-Type = application/json
- Authorization = Bearer {SAT token}

**Response Status:** 200 OK

### Delete firmware rule

**DELETE** `http://<host>:<port>/firmwarerule/2ea59bab-b080-4593-8539-fb6db5fc8fd5`

**Headers:**
- Accept = application/json
- Content-Type = application/json
- Authorization = Bearer {SAT token}

**Response Status:** 204 No Content

---

## Feature

### Retrieve all features

**GET** `http://<host>:<port>/feature`

**Headers:**
- Accept = application/json

**Response Codes:** 200

### Retrieve filtered features

**GET** `http://<host>:<port>/feature/filtered?`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Parameters:**
- `APPLICATION_TYPE` (required): Filter by application type
- `NAME`: Filter features by name
- `FEATURE_INSTANCE`: Filter by feature instance
- `FREE_ARG`: Filter by property key
- `FIXED_ARG`: Filter by property value

**Response Codes:** 200

### Import feature

**POST** `http://<host>:<port>/feature/importAll`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Body:** List of features

**Response Codes:** 200, 400, 409

### Create feature

**POST** `http://<host>:<port>/feature`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 201, 400, 409

### Update feature

**PUT** `http://<host>:<port>/feature`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 201, 400, 404, 409

### Delete feature

**DELETE** `http://<host>:<port>/feature/{id}`

**Response Codes:** 204, 404, 409

---

## Feature Rule

### Retrieve all feature rules

**GET** `http://<host>:<port>/featurerule`

**Headers:**
- Accept = application/json

**Response Codes:** 200

### Retrieve filtered feature rules

**GET** `http://<host>:<port>/featurerule/filtered?`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Parameters:**
- `APPLICATION_TYPE` (required): Filter by application type
- `NAME`: Filter by rule name
- `FREE_ARG`: Filter by feature rule key
- `FIXED_ARG`: Filter by feature rule value
- `FEATURE`: Filter by feature instance

**Response Codes:** 200

### Import feature rule

If feature rule with provided id does not exist it is imported otherwise updated.

**POST** `http://<host>:<port>/featurerule/importAll`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Body:** List of feature rules

**Response Codes:** 200, 400, 404, 409

### Create feature rule

**POST** `http://<host>:<port>/featurerule`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 201, 400, 404, 409

### Update feature rule

**PUT** `http://<host>:<port>/featurerule`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 201, 400, 404, 409

### Delete feature rule

**DELETE** `http://<host>:<port>/featurerule/{id}`

**Response Codes:** 204, 404, 409

---

## Activation Minimum Version

### Retrieve all activation minimum versions

**GET** `http://<host>:<port>/amv`

**Headers:**
- Accept = application/json

**Response Codes:** 200

### Retrieve filtered activation minimum versions

**GET** `http://<host>:<port>/amv/filtered?`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Parameters:**
- `applicationType` (required): Filter by application type
- `DESCRIPTION`: Filter by description
- `MODEL`: Filter by model
- `PARTNER_ID`: Filter by partner id
- `FIRMWARE_VERSION`: Filter by firmware version
- `REGULAR_EXPRESSION`: Filter by regular expression

**Response Codes:** 200

### Import activation version

If activation minimum version with provided id does not exist it is imported otherwise updated.

**POST** `http://<host>:<port>/amv/importAll`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Body:** List of activation minimum versions

**Response Codes:** 200, 400, 404, 409

### Create activation minimum version

**POST** `http://<host>:<port>/amv`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 201, 400, 404, 409

### Update activation minimum version

**PUT** `http://<host>:<port>/amv`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 201, 400, 404, 409

### Delete activation minimum version

**DELETE** `http://<host>:<port>/amv/{id}`

**Response Codes:** 204, 404, 409

---

## Telemetry Profile

### Retrieve all Telemetry Profiles

**GET** `http://<host>:<port>/telemetry/profile`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200

### Retrieve Telemetry Profile

**GET** `http://<host>:<port>/telemetry/profile/{id}`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200, 404

### Create Telemetry Profile

**POST** `http://<host>:<port>/telemetry/profile`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 201, 400, 409

### Update Telemetry Profile

**PUT** `http://<host>:<port>/telemetry/profile`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200, 400, 404, 409

### Delete Telemetry Profile

**DELETE** `http://<host>:<port>/telemetry/profile/{id}`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 204, 404, 409

### Add Telemetry Profile Entry

**PUT** `http://<host>:<port>/telemetry/profile/entry/add/{id}`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200, 404

### Remove Telemetry Profile entry

**PUT** `http://<host>:<port>/telemetry/profile/entry/remove/{id}`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200, 404

### Create Telemetry Profile through pending changes

**POST** `http://<host>:<port>/telemetry/profile/change`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 201, 400, 409

### Update Telemetry Profile with approval

**PUT** `http://<host>:<port>/telemetry/profile/change`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200, 400, 404, 409

### Delete Telemetry Profile with approval

**DELETE** `http://<host>:<port>/telemetry/profile/change/{id}`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 204, 404, 409

### Add Telemetry Profile Entry with approval

**PUT** `http://<host>:<port>/telemetry/profile/change/entry/add/{id}`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200, 404

### Remove Telemetry Profile entry with approval

**PUT** `http://<host>:<port>/telemetry/profile/change/entry/remove/{id}`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200, 404

---

## Telemetry Profile 2.0

### Retrieve all Telemetry 2.0 Profiles

**GET** `http://<host>:<port>/telemetry/v2/profile`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200

**Response Body:** List of Telemetry 2.0 Profiles

```json
[{
    "id": "8fb459f6-044e-4c64-99ff-e0c7c1b4124b",
    "updated": 1646687418358,
    "name": "test",
    "jsonconfig": "...",
    "applicationType": "rdkcloud"
}]
```

### Retrieve Telemetry 2.0 Profile

**GET** `http://<host>:<port>/telemetry/v2/profile/{id}`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200, 404

### Create Telemetry 2.0 Profile

**POST** `http://<host>:<port>/telemetry/v2/profile`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 201, 400, 409

**Request Body:** Telemetry 2.0 Profile with or without id

### Update Telemetry 2.0 Profile

**PUT** `http://<host>:<port>/telemetry/v2/profile`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 200, 400, 404, 409

### Delete Telemetry 2.0 Profile

**DELETE** `http://<host>:<port>/telemetry/v2/profile/{id}`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Codes:** 204, 404, 409

### Create with approval

**POST** `http://<host>:<port>/telemetry/v2/profile/change`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Code:** 200, 400

**Response Body:** Telemetry 2.0 profile change entity

### Update with approval

**PUT** `http://<host>:<port>/telemetry/v2/profile/change`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Code:** 200, 400

**Response Body:** Telemetry 2.0 profile change entity

### Delete with approval

**DELETE** `http://<host>:<port>/telemetry/v2/profile/change/{profile id}`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Response Status:** 200, 404, 409

---

## Change v2 API

### Approve by change ids (not supported in golang implementation)

All not selected changes by entity id will be canceled. For example, there are two changes, "change A" and "change B", id of "change A" is sent to approve, in result "change A" will be approved and "change B" will be canceled.

**POST** `http://<host>:<port>/change/v2/approve/byChangeIds`

**Headers:**
- Accept = application/json
- Content-Type = application/json

**Request Body:** Array with change ids
```json
["4705486f-2dcc-4ae9-a920-a45b33755993"]
```

**Response Codes:** 200, 404, 409

**Response Body:** Change id - error message object. If change successfully approve an empty object is returned

### Approve by entity id (not supported in golang implementation)

To approve all changes by entity id

**GET** `http://<host>:<port>/change/v2/approve/byEntityId/{id}`

**Response Codes:** 200, 404, 409

**Response Body:** Change id - error message object. If change successfully approve an empty object is returned

### Cancel change

**GET** `http://<host>:<port>/change/v2/cancel/{changeId}`

**Response Status:** 200, 404

### Retrieve all changes

**GET** `http://<host>:<port>/change/v2/all`

**Headers:**
- Accept = application/json

**Response Codes:** 200

**Response Body:** Array with all telemetry changes

---

## Example: Telemetry 2.0 Profile Update with Approval

This is an example of updating a Telemetry 2.0 Profile with approval:

**Request:**
```
PUT http://<host>:<port>/telemetry/v2/profile/change
```

**Response Body Example:**
```json
{
    "id": "c3fee291-5376-40cf-88a3-96aadaa0e28b",
    "updated": 1659727767163,
    "entityId": "8205d716-8e45-4570-a34b-f1ebe0bdc75e",
    "entityType": "TELEMETRY_TWO_PROFILE",
    "newEntity": {
        "@type": "TelemetryTwoProfile",
        "id": "8205d716-8e45-4570-a34b-f1ebe0bdc75e",
        "updated": 1621625846548,
        "name": "Test Telemetry 2.0 Profile name",
        "jsonconfig": "...",
        "applicationType": "stb"
    },
    "oldEntity": {
        "@type": "TelemetryTwoProfile",
        "id": "8205d716-8e45-4570-a34b-f1ebe0bdc75e",
        "updated": 1659727722268,
        "name": "Test Telemetry 2.0 Profile name",
        "jsonconfig": "...",
        "applicationType": "stb"
    },
    "operation": "UPDATE",
    "author": "UNKNOWN_USER"
}
```
