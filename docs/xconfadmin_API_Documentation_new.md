# XConf Admin REST API Documentation

## Admin API Calling Changes

Please note the below examples earlier were towards XConf DataService, but as of Oct 6th, 2021, these XConf APIs are considered Admin APIs, and are not available on the XConf DataService endpoint. These APIs are available on the XConf Admin Service.

Also note that to communicate with XConf Admin API service you will need a SAT token as a Bearer Token in your RESTful API call. XConf Admin Service API endpoint is protected by Comcast firewall, so the endpoint is not visible to Internet, only client from Comcast ENTERPRISE networks can access this XConf admin service endpoint. If your application is not in the Comcast network, then it is required that the application be configured to communicate with XConf Admin service via CodeBig2. Request for xconf admin service credentials on CodeBig2 portal.

Use the following prefix for the below APIs instead of http://<host>:<port>/:
```
https://<admin-service-endpoint>:9443/xconfAdminService/
```

If communicating via CodeBig2, then use:
```
https://<codebig-admin-service-endpoint>/xconfAdminService/
```

**Important Notes:**
- The http://<host>:<port>/queries/foo.json method of calling is not supported and will not be supported
- In the next major revision of XConf Admin service, all Admin API responses will default to "application/json" response format
- Application types "text/xml" or "application/xml" for XConf API is no longer supported

---

## API Endpoints Overview

### Firmware Config
- Retrieve a list of firmware configs
- Retrieve a single firmware config by id
- Retrieve firmware configs by modelId
- Create/update a firmware config
- Delete a firmware config by id

### IP rules
- Retrieve an ip rule list
- Retrieve an ip rule by name
- Create/update an ip rule
- Delete an ip rule

### Location filter
- Retrieve a location filter list
- Retrieve a location filter by name
- Create/update location filter
- Delete location filter by name

### Download location filter
- Retrieve download location filter
- Update download location filter

### Environment model rules
- Retrieve an environment model rule list
- Retrieve an environment model rule by name
- Create/update an environment model rule
- Delete an environment model rule

### IP filter
- Retrieve an IP filter list
- Retrieve an ip filter by name
- Create/update an IP filter
- Delete IP filter

### Percent filter
- Retrieve percent filter
- Retrieve percent filter field values
- Update percent filter
- Retrieve EnvModelPercentages
- Retrieve EnvModelPercentage by id
- Create envModelPercentage
- Update EnvModelPercentage
- Delete envModelPercentage

### Time filter
- Retrieve time filter list
- Retrieve time filter by name
- Create/update time filter
- Delete time filter by name

### Environment
- Retrieve an list of environments
- Retrieve environment by id
- Create an environment
- Delete environment by id

### IP address group
- Retrieve an IP address group list
- Retrieve an IP address group by name
- Retrieve an IP address group by IP
- Create an IP address group
- Add data to IP Address Group (dev in progress)
- Delete data from IP Address Group (dev in progress)
- Delete an IP address group by id

### Mac rule
- Retrieve a mac rule list (legacy)
- Retrieve a mac rule list (v2)
- Retrieve mac rule by name (legacy)
- Retrieve mac rule by name (v2)
- Retrieve mac rule by mac address (legacy)
- Retrieve mac rule by mac address (v2)
- Create/update mac rule
- Delete mac rule by name

### Model
- Retrieve a model list
- Retrieve model by id
- Create model
- Update model description
- Delete model by id

### NamespacedList
- Retrieve all NS lists
- Retrieve NS list by mac part
- Create a NS list
- Add data to NS list
- Delete data from NS list
- Delete an NS list by id

### RebootImmediately filter
- Retrieve an RI filter list
- Retrieve and RI filter by rule name
- Create/update an RI filter
- Delete RI filter by name

### FirmwareRuleTemplate
- Retrieve filtered templates
- Import firmware rule templates
- Create firmware rule template
- Updating firmware rule template
- Deleting Firmware rule template

### FirmwareRule
- Retrieve all firmware rules
- Retrieve filtered firmware rules
- Import firmware rule
- Create firmware rule
- Update firmware Rule
- Delete firmware Rule

### Feature
- Retrieve all features
- Retrieve filtered features
- Import feature
- Create feature
- Update feature
- Delete feature

### Feature Rule
- Retrieve all feature rules
- Retrieve filtered feature rules
- Import feature rule
- Create feature rule
- Update feature rule
- Delete feature rule

### Activation Minimum Version
- Retrieve all activation minimum versions
- Retrieve filtered activation minimum versions
- Import activation version
- Create activation minimum version
- Update activation minimum version
- Delete activation minimum version

### Telemetry Profile
- Retrieve all Telemetry Profiles
- Retrieve Telemetry Profile
- Create Telemetry Profile
- Update Telemetry Profile
- Delete Telemetry Profile
- Add Telemetry Profile Entry
- Remove Telemetry Profile entry
- Create Telemetry Profile through pending changes
- Update Telemetry Profile with approval
- Delete Telemetry Profile with approval
- Add Telemetry Profile Entry with approval
- Remove Telemetry Profile entry with approval

### Telemetry Profile 2.0
- Retrieve all Telemetry 2.0 Profiles
- Retrieve Telemetry 2.0 Profile
- Create Telemetry 2.0 Profile
- Update Telemetry 2.0 Profile
- Delete Telemetry 2.0 Profile
- Create with approval
- Update with approval
- Delete with approval
- Telemetry 2.0 Profile Json Schema

### Change API
- Approve by change ids (not supported in golang implementation)
- Approve by entity id (not supported in golang implementation)
- Cancel change
- Retrieve all changes

### Change v2 API
- Approve by change ids (not supported in golang implementation)
- Approve by entity id (not supported in golang implementation)
- Cancel change
- Retrieve all changes

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
