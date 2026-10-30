# XconfAdmin API Documentation

## Overview
The XconfAdmin API provides comprehensive configuration management for RDK-B devices through a RESTful interface. It manages firmware configurations, device settings, telemetry profiles, feature rules, DCM settings, authentication, and various query operations.

**Base URL**: `/xconfAdminService`

## Authentication
Most endpoints require authentication via JWT token or session-based authentication. Each request is validated for appropriate permissions based on the entity type being accessed.

---

# Authentication APIs

## Get Authentication Provider
Get the available authentication provider configuration:

**GET** `http://<host>:<port>/provider`

<details>
<summary><strong>Response Body:</strong> Authentication provider information</summary>

```json
{
  "provider": "xerxes",
  "loginUrl": "/auth/login",
  "enabled": true
}
```
</details>

Response Codes: 200

---

## Get Authentication Info
Get current authentication information:

**GET** `http://<host>:<port>/auth/info`

<details>
<summary><strong>Response Body:</strong> Authentication status</summary>

```json
{
  "authenticated": true,
  "username": "admin",
  "permissions": ["READ", "WRITE"],
  "roles": ["ADMIN"]
}
```
</details>

Response Codes: 200, 401

---

## Basic Authentication
Perform basic authentication with username and password:

**POST** `http://<host>:<port>/auth/basic`

<details>
<summary><strong>Request Body:</strong> Basic authentication credentials</summary>

```json
{
  "username": "admin",
  "password": "password123"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Authentication result</summary>

```json
{
  "success": true,
  "token": "jwt-token-here",
  "expires": "2025-11-06T10:00:00Z"
}
```
</details>

Response Codes: 200, 401

---

# Query APIs

## Get All Environments
Get all available environments:

**GET** `http://<host>:<port>/queries/environments`

<details>
<summary><strong>Response Body:</strong> Array of environment objects</summary>

```json
[
  {
    "id": "env-001",
    "name": "Production",
    "description": "Production environment"
  },
  {
    "id": "env-002", 
    "name": "Staging",
    "description": "Staging environment"
  }
]
```
</details>

Response Codes: 200, 401

---

## Get Environment by ID
Get specific environment by identifier:

**GET** `http://<host>:<port>/queries/environments/{id}`

<details>
<summary><strong>Response Body:</strong> Single environment object</summary>

```json
{
  "id": "env-001",
  "name": "Production",
  "description": "Production environment"
}
```
</details>

Response Codes: 200, 404, 401

---

## Get All Models
Get all device models:

**GET** `http://<host>:<port>/queries/models`

<details>
<summary><strong>Response Body:</strong> Array of model objects</summary>

```json
[
  {
    "id": "model-001",
    "name": "STB_MODEL_X",
    "description": "Set-top box model X"
  }
]
```
</details>

Response Codes: 200, 401

---

## Get Model by ID
Get specific model by identifier:

**GET** `http://<host>:<port>/queries/models/{id}`

<details>
<summary><strong>Response Body:</strong> Single model object</summary>

```json
{
  "id": "model-001",
  "name": "STB_MODEL_X",
  "description": "Set-top box model X"
}
```
</details>

Response Codes: 200, 404, 401

---

## Get IP Address Groups
Get all IP address groups:

**GET** `http://<host>:<port>/queries/ipAddressGroups`

<details>
<summary><strong>Response Body:</strong> Array of IP address group objects</summary>

```json
[
  {
    "id": "ip-group-001",
    "name": "Production IPs",
    "ipAddresses": ["192.168.1.100", "192.168.1.101"]
  }
]
```
</details>

Response Codes: 200, 401

---

## Get IP Address Group by IP
Get IP address groups containing specific IP:

**GET** `http://<host>:<port>/queries/ipAddressGroups/byIp/{ipAddress}`

<details>
<summary><strong>Response Body:</strong> Array of matching IP address groups</summary>

```json
[
  {
    "id": "ip-group-001",
    "name": "Production IPs",
    "ipAddresses": ["192.168.1.100", "192.168.1.101"]
  }
]
```
</details>

Response Codes: 200, 404, 401

---

## Get Firmware Configurations
Get all firmware configurations:

**GET** `http://<host>:<port>/queries/firmwares`

<details>
<summary><strong>Response Body:</strong> Array of firmware configuration objects</summary>

```json
[
  {
    "id": "fw-001",
    "firmwareFilename": "firmware_v1.0.bin",
    "version": "1.0.0",
    "description": "Production firmware v1.0",
    "supportedModelIds": ["model-001"],
    "firmwareDownloadProtocol": "http",
    "firmwareLocation": "http://cdn.example.com/firmware/"
  }
]
```
</details>

Response Codes: 200, 401

---

## Get Firmware Configuration by ID
Get specific firmware configuration by ID:

**GET** `http://<host>:<port>/queries/firmwares/{id}`

<details>
<summary><strong>Response Body:</strong> Single firmware configuration object</summary>

```json
{
  "id": "fw-001",
  "firmwareFilename": "firmware_v1.0.bin",
  "version": "1.0.0",
  "description": "Production firmware v1.0",
  "supportedModelIds": ["model-001"],
  "firmwareDownloadProtocol": "http",
  "firmwareLocation": "http://cdn.example.com/firmware/"
}
```
</details>

Response Codes: 200, 404, 401

---

# Model Management APIs

## Get All Models
Get all device models with pagination support:

**GET** `http://<host>:<port>/model`

<details>
<summary><strong>Response Body:</strong> Array of model objects</summary>

```json
[
  {
    "id": "model-001",
    "name": "STB_MODEL_X",
    "description": "Set-top box model X"
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Model
Create a new device model:

**POST** `http://<host>:<port>/model`

<details>
<summary><strong>Request Body:</strong> Model creation data</summary>

```json
{
  "id": "model-002",
  "name": "STB_MODEL_Y",
  "description": "Set-top box model Y"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created model object</summary>

```json
{
  "id": "model-002",
  "name": "STB_MODEL_Y",
  "description": "Set-top box model Y"
}
```
</details>

Response Codes: 201, 400, 401

---

## Update Model
Update an existing device model:

**PUT** `http://<host>:<port>/model`

<details>
<summary><strong>Request Body:</strong> Model update data</summary>

```json
{
  "id": "model-002",
  "name": "STB_MODEL_Y_Updated",
  "description": "Updated set-top box model Y"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Updated model object</summary>

```json
{
  "id": "model-002",
  "name": "STB_MODEL_Y_Updated",
  "description": "Updated set-top box model Y"
}
```
</details>

Response Codes: 200, 400, 404, 401

---

## Delete Model
Delete a device model by ID:

**DELETE** `http://<host>:<port>/model/{id}`

Response Codes: 204, 404, 401

---

# Environment Management APIs

## Get All Environments
Get all environments with pagination support:

**GET** `http://<host>:<port>/environment`

<details>
<summary><strong>Response Body:</strong> Array of environment objects</summary>

```json
[
  {
    "id": "env-001",
    "name": "Production",
    "description": "Production environment"
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Environment
Create a new environment:

**POST** `http://<host>:<port>/environment`

<details>
<summary><strong>Request Body:</strong> Environment creation data</summary>

```json
{
  "id": "env-003",
  "name": "Development",
  "description": "Development environment"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created environment object</summary>

```json
{
  "id": "env-003",
  "name": "Development",
  "description": "Development environment"
}
```
</details>

Response Codes: 201, 400, 401

---

## Update Environment
Update an existing environment:

**PUT** `http://<host>:<port>/environment`

<details>
<summary><strong>Request Body:</strong> Environment update data</summary>

```json
{
  "id": "env-003",
  "name": "Development_Updated",
  "description": "Updated development environment"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Updated environment object</summary>

```json
{
  "id": "env-003",
  "name": "Development_Updated",
  "description": "Updated development environment"
}
```
</details>

Response Codes: 200, 400, 404, 401

---

## Delete Environment
Delete an environment by ID:

**DELETE** `http://<host>:<port>/environment/{id}`

Response Codes: 204, 404, 401

---

# Firmware Rule Management APIs

## Get All Firmware Rules
Get all firmware rules with filtering:

**GET** `http://<host>:<port>/firmwarerule/filtered?pageNumber=1&pageSize=10`

<details>
<summary><strong>Response Body:</strong> Array of firmware rule objects</summary>

```json
[
  {
    "id": "fw-rule-001",
    "name": "Production STB Rule",
    "type": "MAC_RULE",
    "rule": {
      "condition": {
        "freeArg": "model",
        "operation": "IS",
        "fixedArg": "STB_MODEL_X"
      }
    },
    "applicableAction": {
      "actionType": "RULE",
      "configId": "fw-001"
    },
    "applicationType": "stb"
  }
]
```
</details>

Response Codes: 200, 401

---

## Get Firmware Rule by ID
Get specific firmware rule by ID:

**GET** `http://<host>:<port>/firmwarerule/{id}`

<details>
<summary><strong>Response Body:</strong> Single firmware rule object</summary>

```json
{
  "id": "fw-rule-001",
  "name": "Production STB Rule",
  "type": "MAC_RULE",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS",
      "fixedArg": "STB_MODEL_X"
    }
  },
  "applicableAction": {
    "actionType": "RULE",
    "configId": "fw-001"
  },
  "applicationType": "stb"
}
```
</details>

Response Codes: 200, 404, 401

---

## Create Firmware Rule
Create a new firmware rule:

**POST** `http://<host>:<port>/firmwarerule`

<details>
<summary><strong>Request Body:</strong> Firmware rule creation data</summary>

```json
{
  "name": "New STB Rule",
  "type": "MAC_RULE",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS", 
      "fixedArg": "STB_MODEL_Y"
    }
  },
  "applicableAction": {
    "actionType": "RULE",
    "configId": "fw-002"
  },
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created firmware rule object</summary>

```json
{
  "id": "fw-rule-002",
  "name": "New STB Rule",
  "type": "MAC_RULE",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS",
      "fixedArg": "STB_MODEL_Y"
    }
  },
  "applicableAction": {
    "actionType": "RULE",
    "configId": "fw-002"
  },
  "applicationType": "stb"
}
```
</details>

Response Codes: 201, 400, 401

---

## Update Firmware Rule
Update an existing firmware rule:

**PUT** `http://<host>:<port>/firmwarerule`

<details>
<summary><strong>Request Body:</strong> Firmware rule update data</summary>

```json
{
  "id": "fw-rule-002",
  "name": "Updated STB Rule",
  "type": "MAC_RULE",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS",
      "fixedArg": "STB_MODEL_Y"
    }
  },
  "applicableAction": {
    "actionType": "RULE",
    "configId": "fw-002"
  },
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Updated firmware rule object</summary>

```json
{
  "id": "fw-rule-002",
  "name": "Updated STB Rule",
  "type": "MAC_RULE",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS",
      "fixedArg": "STB_MODEL_Y"
    }
  },
  "applicableAction": {
    "actionType": "RULE",
    "configId": "fw-002"
  },
  "applicationType": "stb"
}
```
</details>

Response Codes: 200, 400, 404, 401

---

## Delete Firmware Rule
Delete a firmware rule by ID:

**DELETE** `http://<host>:<port>/firmwarerule/{id}`

Response Codes: 204, 404, 401

---

# Firmware Configuration APIs

## Get All Firmware Configurations
Get all firmware configurations:

**GET** `http://<host>:<port>/firmwareconfig`

<details>
<summary><strong>Response Body:</strong> Array of firmware configuration objects</summary>

```json
[
  {
    "id": "fw-001",
    "firmwareFilename": "firmware_v1.0.bin",
    "version": "1.0.0",
    "description": "Production firmware v1.0",
    "supportedModelIds": ["model-001"],
    "firmwareDownloadProtocol": "http",
    "firmwareLocation": "http://cdn.example.com/firmware/"
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Firmware Configuration
Create a new firmware configuration:

**POST** `http://<host>:<port>/firmwareconfig`

<details>
<summary><strong>Request Body:</strong> Firmware configuration creation data</summary>

```json
{
  "firmwareFilename": "firmware_v2.0.bin",
  "version": "2.0.0",
  "description": "New firmware v2.0",
  "supportedModelIds": ["model-001", "model-002"],
  "firmwareDownloadProtocol": "https",
  "firmwareLocation": "https://secure-cdn.example.com/firmware/"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created firmware configuration object</summary>

```json
{
  "id": "fw-002",
  "firmwareFilename": "firmware_v2.0.bin",
  "version": "2.0.0",
  "description": "New firmware v2.0",
  "supportedModelIds": ["model-001", "model-002"],
  "firmwareDownloadProtocol": "https",
  "firmwareLocation": "https://secure-cdn.example.com/firmware/"
}
```
</details>

Response Codes: 201, 400, 401

---

## Update Firmware Configuration
Update an existing firmware configuration:

**PUT** `http://<host>:<port>/firmwareconfig`

<details>
<summary><strong>Request Body:</strong> Firmware configuration update data</summary>

```json
{
  "id": "fw-002",
  "firmwareFilename": "firmware_v2.1.bin",
  "version": "2.1.0",
  "description": "Updated firmware v2.1",
  "supportedModelIds": ["model-001", "model-002"],
  "firmwareDownloadProtocol": "https",
  "firmwareLocation": "https://secure-cdn.example.com/firmware/"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Updated firmware configuration object</summary>

```json
{
  "id": "fw-002",
  "firmwareFilename": "firmware_v2.1.bin",
  "version": "2.1.0",
  "description": "Updated firmware v2.1",
  "supportedModelIds": ["model-001", "model-002"],
  "firmwareDownloadProtocol": "https",
  "firmwareLocation": "https://secure-cdn.example.com/firmware/"
}
```
</details>

Response Codes: 200, 400, 404, 401

---

## Delete Firmware Configuration
Delete a firmware configuration by ID:

**DELETE** `http://<host>:<port>/firmwareconfig/{id}`

Response Codes: 204, 404, 401

---

# DCM (Device Configuration Management) APIs

## Get All DCM Formulas
Get all DCM formulas for authenticated application type:

**GET** `http://<host>:<port>/dcm/formula?applicationType=stb`

<details>
<summary><strong>Response Body:</strong> Array of DCM Formula objects</summary>

```json
[
  {
    "id": "formula-001",
    "name": "STB Configuration Rule",
    "description": "Configuration for STB devices",
    "priority": 1,
    "ruleExpression": "model == 'STB_MODEL_X'",
    "percentage": 100,
    "percentageL1": 50.0,
    "percentageL2": 30.0,
    "percentageL3": 20.0,
    "applicationType": "stb",
    "rule": {
      "condition": {
        "freeArg": "model",
        "operation": "IS",
        "fixedArg": "STB_MODEL_X"
      }
    }
  }
]
```
</details>

Response Codes: 200, 401

---

## Create DCM Formula
Create a new DCM formula:

**POST** `http://<host>:<port>/dcm/formula?applicationType=stb`

<details>
<summary><strong>Request Body:</strong> DCM Formula creation data</summary>

```json
{
  "name": "New STB Configuration Rule",
  "description": "New configuration for STB devices",
  "priority": 2,
  "ruleExpression": "model == 'STB_MODEL_Y'",
  "percentage": 75,
  "percentageL1": 40.0,
  "percentageL2": 35.0,
  "percentageL3": 25.0,
  "applicationType": "stb",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS",
      "fixedArg": "STB_MODEL_Y"
    }
  }
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created DCM Formula object</summary>

```json
{
  "id": "formula-002",
  "name": "New STB Configuration Rule",
  "description": "New configuration for STB devices",
  "priority": 2,
  "ruleExpression": "model == 'STB_MODEL_Y'",
  "percentage": 75,
  "percentageL1": 40.0,
  "percentageL2": 35.0,
  "percentageL3": 25.0,
  "applicationType": "stb",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS",
      "fixedArg": "STB_MODEL_Y"
    }
  }
}
```
</details>

Response Codes: 201, 400, 401

---

## Get All Device Settings
Get all device settings:

**GET** `http://<host>:<port>/dcm/deviceSettings`

<details>
<summary><strong>Response Body:</strong> Array of Device Settings objects</summary>

```json
[
  {
    "id": "device-settings-001",
    "name": "STB Device Settings",
    "checkOnReboot": true,
    "settingsAreActive": true,
    "schedule": {
      "type": "CronExpression",
      "expression": "0 2 * * *",
      "timeWindowMinutes": 60,
      "startDate": "2025-01-01",
      "endDate": "2025-12-31"
    },
    "configData": {
      "logLevel": "INFO",
      "uploadOnReboot": true
    }
  }
]
```
</details>

Response Codes: 200, 401

---

## Get All Log Upload Settings  
Get all log upload settings:

**GET** `http://<host>:<port>/dcm/logUploadSettings`

<details>
<summary><strong>Response Body:</strong> Array of Log Upload Settings objects</summary>

```json
[
  {
    "id": "log-upload-001",
    "name": "Production Log Upload",
    "uploadOnReboot": true,
    "numberOfDays": 7,
    "areSettingsActive": true,
    "modeToGetLogFiles": "LogFiles",
    "schedule": {
      "type": "CronExpression",
      "expression": "0 3 * * *",
      "timeWindowMinutes": 120
    },
    "logFiles": [
      {
        "name": "system.log",
        "logFileName": "/var/log/system.log"
      }
    ],
    "logUploadSettings": {
      "uploadRepositoryName": "prod-repo",
      "uploadProtocol": "HTTPS"
    }
  }
]
```
</details>

Response Codes: 200, 401

---

## Get All VOD Settings
Get all Video On Demand settings:

**GET** `http://<host>:<port>/dcm/vodsettings`

<details>
<summary><strong>Response Body:</strong> Array of VOD Settings objects</summary>

```json
[
  {
    "id": "vod-settings-001", 
    "name": "Production VOD Settings",
    "locationsURL": "https://vod.example.com/locations",
    "srmIPList": ["192.168.1.10", "192.168.1.11"],
    "ipNames": ["SRM-1", "SRM-2"],
    "ipList": ["192.168.1.10", "192.168.1.11"]
  }
]
```
</details>

Response Codes: 200, 401

---

## Get All Upload Repositories
Get all upload repository settings:

**GET** `http://<host>:<port>/dcm/uploadRepository`

<details>
<summary><strong>Response Body:</strong> Array of Upload Repository objects</summary>

```json
[
  {
    "id": "repo-001",
    "name": "Production Repository",
    "description": "Production log upload repository",
    "url": "https://logs.example.com/upload",
    "protocol": "HTTPS",
    "applicationType": "stb"
  }
]
```
</details>

Response Codes: 200, 401

---

# RFC (Remote Feature Control) APIs

## Get All Features
Get all feature definitions:

**GET** `http://<host>:<port>/rfc/feature`

<details>
<summary><strong>Response Body:</strong> Array of Feature objects</summary>

```json
[
  {
    "id": "feature-001",
    "featureName": "WiFiDualBand",
    "effectiveImmediate": true,
    "enable": true,
    "configData": {
      "band2g": "enabled",
      "band5g": "enabled"
    },
    "applicationType": "stb"
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Feature
Create a new feature:

**POST** `http://<host>:<port>/rfc/feature`

<details>
<summary><strong>Request Body:</strong> Feature creation data</summary>

```json
{
  "featureName": "NewFeature",
  "effectiveImmediate": false,
  "enable": true,
  "configData": {
    "setting1": "value1",
    "setting2": "value2"
  },
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created Feature object</summary>

```json
{
  "id": "feature-002",
  "featureName": "NewFeature", 
  "effectiveImmediate": false,
  "enable": true,
  "configData": {
    "setting1": "value1",
    "setting2": "value2"
  },
  "applicationType": "stb"
}
```
</details>

Response Codes: 201, 400, 401

---

## Get All Feature Rules
Get all feature rules:

**GET** `http://<host>:<port>/rfc/featurerule`

<details>
<summary><strong>Response Body:</strong> Array of Feature Rule objects</summary>

```json
[
  {
    "id": "feature-rule-001",
    "name": "WiFi Feature Rule",
    "rule": {
      "condition": {
        "freeArg": "model",
        "operation": "IS",
        "fixedArg": "STB_MODEL_X"
      }
    },
    "priority": 1,
    "featureIds": ["feature-001"],
    "applicationType": "stb"
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Feature Rule
Create a new feature rule:

**POST** `http://<host>:<port>/rfc/featurerule`

<details>
<summary><strong>Request Body:</strong> Feature Rule creation data</summary>

```json
{
  "name": "New Feature Rule",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS",
      "fixedArg": "STB_MODEL_Y"
    }
  },
  "priority": 2,
  "featureIds": ["feature-002"],
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created Feature Rule object</summary>

```json
{
  "id": "feature-rule-002",
  "name": "New Feature Rule",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS", 
      "fixedArg": "STB_MODEL_Y"
    }
  },
  "priority": 2,
  "featureIds": ["feature-002"],
  "applicationType": "stb"
}
```
</details>

Response Codes: 201, 400, 401

---

# Telemetry APIs

## Get All Telemetry Profiles (v1)
Get all telemetry v1 profiles:

**GET** `http://<host>:<port>/telemetry/profile`

<details>
<summary><strong>Response Body:</strong> Array of Telemetry Profile objects</summary>

```json
[
  {
    "id": "telemetry-profile-001",
    "name": "STB Telemetry Profile",
    "schedule": "0 */15 * * * *",
    "expires": 0,
    "telemetryProfile": [
      {
        "header": "System_Info",
        "content": "Device.DeviceInfo.Manufacturer,Device.DeviceInfo.ModelName",
        "type": "JSON",
        "pollingFrequency": "900"
      }
    ],
    "applicationType": "stb"
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Telemetry Profile (v1)
Create a new telemetry v1 profile:

**POST** `http://<host>:<port>/telemetry/profile`

<details>
<summary><strong>Request Body:</strong> Telemetry Profile creation data</summary>

```json
{
  "name": "New STB Telemetry Profile",
  "schedule": "0 */30 * * * *",
  "expires": 0,
  "telemetryProfile": [
    {
      "header": "Memory_Info",
      "content": "Device.DeviceInfo.MemoryStatus.Total,Device.DeviceInfo.MemoryStatus.Free",
      "type": "JSON",
      "pollingFrequency": "1800"
    }
  ],
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created Telemetry Profile object</summary>

```json
{
  "id": "telemetry-profile-002",
  "name": "New STB Telemetry Profile",
  "schedule": "0 */30 * * * *",
  "expires": 0,
  "telemetryProfile": [
    {
      "header": "Memory_Info",
      "content": "Device.DeviceInfo.MemoryStatus.Total,Device.DeviceInfo.MemoryStatus.Free",
      "type": "JSON",
      "pollingFrequency": "1800"
    }
  ],
  "applicationType": "stb"
}
```
</details>

Response Codes: 201, 400, 401

---

## Get All Telemetry Profiles (v2)
Get all telemetry v2 profiles:

**GET** `http://<host>:<port>/telemetry/v2/profile`

<details>
<summary><strong>Response Body:</strong> Array of Telemetry v2 Profile objects</summary>

```json
[
  {
    "id": "telemetry-v2-profile-001",
    "name": "STB Telemetry v2 Profile",
    "protocol": "HTTP",
    "encodingType": "JSON",
    "reportingInterval": 900,
    "timeReference": "0001-01-01T00:00:00Z",
    "parameter": [
      {
        "name": "Device.WiFi.Radio.1.Stats.BytesSent",
        "reference": "0001-01-01T00:00:00Z",
        "use": "Absolute"
      }
    ],
    "applicationType": "stb"
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Telemetry Profile (v2)
Create a new telemetry v2 profile:

**POST** `http://<host>:<port>/telemetry/v2/profile`

<details>
<summary><strong>Request Body:</strong> Telemetry v2 Profile creation data</summary>

```json
{
  "name": "New STB Telemetry v2 Profile",
  "protocol": "HTTPS",
  "encodingType": "JSON",
  "reportingInterval": 1800,
  "timeReference": "0001-01-01T00:00:00Z",
  "parameter": [
    {
      "name": "Device.WiFi.Radio.2.Stats.BytesSent",
      "reference": "0001-01-01T00:00:00Z", 
      "use": "Absolute"
    }
  ],
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created Telemetry v2 Profile object</summary>

```json
{
  "id": "telemetry-v2-profile-002",
  "name": "New STB Telemetry v2 Profile",
  "protocol": "HTTPS",
  "encodingType": "JSON",
  "reportingInterval": 1800,
  "timeReference": "0001-01-01T00:00:00Z",
  "parameter": [
    {
      "name": "Device.WiFi.Radio.2.Stats.BytesSent",
      "reference": "0001-01-01T00:00:00Z",
      "use": "Absolute"
    }
  ],
  "applicationType": "stb"
}
```
</details>

Response Codes: 201, 400, 401

---

# Settings APIs

## Get All Setting Profiles
Get all setting profiles:

**GET** `http://<host>:<port>/setting/profile`

<details>
<summary><strong>Response Body:</strong> Array of Setting Profile objects</summary>

```json
[
  {
    "id": "setting-profile-001",
    "settingProfileId": "wifi-settings",
    "settingType": "EPON",
    "properties": {
      "wifiEnabled": "true",
      "channel": "auto"
    },
    "applicationType": "stb"
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Setting Profile
Create a new setting profile:

**POST** `http://<host>:<port>/setting/profile`

<details>
<summary><strong>Request Body:</strong> Setting Profile creation data</summary>

```json
{
  "settingProfileId": "network-settings",
  "settingType": "EPON", 
  "properties": {
    "dhcpEnabled": "true",
    "dnsServer": "8.8.8.8"
  },
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created Setting Profile object</summary>

```json
{
  "id": "setting-profile-002",
  "settingProfileId": "network-settings",
  "settingType": "EPON",
  "properties": {
    "dhcpEnabled": "true",
    "dnsServer": "8.8.8.8"
  },
  "applicationType": "stb"
}
```
</details>

Response Codes: 201, 400, 401

---

## Get All Setting Rules
Get all setting rules:

**GET** `http://<host>:<port>/setting/rule`

<details>
<summary><strong>Response Body:</strong> Array of Setting Rule objects</summary>

```json
[
  {
    "id": "setting-rule-001",
    "name": "WiFi Setting Rule",
    "rule": {
      "condition": {
        "freeArg": "model",
        "operation": "IS",
        "fixedArg": "STB_MODEL_X"
      }
    },
    "boundSettingId": "setting-profile-001",
    "applicationType": "stb"
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Setting Rule
Create a new setting rule:

**POST** `http://<host>:<port>/setting/rule`

<details>
<summary><strong>Request Body:</strong> Setting Rule creation data</summary>

```json
{
  "name": "Network Setting Rule",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS",
      "fixedArg": "STB_MODEL_Y"
    }
  },
  "boundSettingId": "setting-profile-002",
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created Setting Rule object</summary>

```json
{
  "id": "setting-rule-002",
  "name": "Network Setting Rule",
  "rule": {
    "condition": {
      "freeArg": "model",
      "operation": "IS",
      "fixedArg": "STB_MODEL_Y"
    }
  },
  "boundSettingId": "setting-profile-002",
  "applicationType": "stb"
}
```
</details>

Response Codes: 201, 400, 401

---

# Percentage Filter APIs

## Get All Percentage Beans
Get all percentage filter beans:

**GET** `http://<host>:<port>/percentfilter/percentageBean`

<details>
<summary><strong>Response Body:</strong> Array of Percentage Bean objects</summary>

```json
[
  {
    "id": "percent-bean-001",
    "name": "Production Rollout",
    "active": true,
    "value": 25.0,
    "firmwareCheckRequired": true,
    "rebootDecoupled": false,
    "applicationType": "stb",
    "lastKnownGood": "fw-001",
    "intermediateVersion": "fw-002",
    "firmwareVersions": ["fw-001", "fw-002"],
    "distributions": [
      {
        "configId": "fw-002",
        "percentage": 25.0
      }
    ]
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Percentage Bean  
Create a new percentage filter bean:

**POST** `http://<host>:<port>/percentfilter/percentageBean`

<details>
<summary><strong>Request Body:</strong> Percentage Bean creation data</summary>

```json
{
  "name": "Beta Rollout",
  "active": true,
  "value": 10.0,
  "firmwareCheckRequired": true,
  "rebootDecoupled": false,
  "applicationType": "stb",
  "lastKnownGood": "fw-001",
  "intermediateVersion": "fw-003",
  "firmwareVersions": ["fw-001", "fw-003"],
  "distributions": [
    {
      "configId": "fw-003",
      "percentage": 10.0
    }
  ]
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created Percentage Bean object</summary>

```json
{
  "id": "percent-bean-002",
  "name": "Beta Rollout",
  "active": true,
  "value": 10.0,
  "firmwareCheckRequired": true,
  "rebootDecoupled": false,
  "applicationType": "stb",
  "lastKnownGood": "fw-001",
  "intermediateVersion": "fw-003",
  "firmwareVersions": ["fw-001", "fw-003"],
  "distributions": [
    {
      "configId": "fw-003",
      "percentage": 10.0
    }
  ]
}
```
</details>

Response Codes: 201, 400, 401

---

# Namespaced List APIs

## Get All Namespaced Lists
Get all generic namespaced lists:

**GET** `http://<host>:<port>/genericnamespacedlist`

<details>
<summary><strong>Response Body:</strong> Array of Namespaced List objects</summary>

```json
[
  {
    "id": "ns-list-001",
    "typeName": "MAC_LIST",
    "data": ["00:11:22:33:44:55", "AA:BB:CC:DD:EE:FF"],
    "applicationType": "stb"
  }
]
```
</details>

Response Codes: 200, 401

---

## Create Namespaced List
Create a new generic namespaced list:

**POST** `http://<host>:<port>/genericnamespacedlist`

<details>
<summary><strong>Request Body:</strong> Namespaced List creation data</summary>

```json
{
  "id": "ns-list-002",
  "typeName": "IP_LIST",
  "data": ["192.168.1.100", "192.168.1.101"],
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Created Namespaced List object</summary>

```json
{
  "id": "ns-list-002",
  "typeName": "IP_LIST",
  "data": ["192.168.1.100", "192.168.1.101"],
  "applicationType": "stb"
}
```
</details>

Response Codes: 201, 400, 401

---

# Test Page APIs

## Firmware Test Page
Test firmware rule evaluation:

**POST** `http://<host>:<port>/firmwarerule/testpage`

<details>
<summary><strong>Request Body:</strong> Test parameters</summary>

```json
{
  "parameters": {
    "mac": "00:11:22:33:44:55",
    "model": "STB_MODEL_X",
    "env": "PROD"
  },
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Test result</summary>

```json
{
  "result": "RULE_MATCH",
  "firmwareConfig": {
    "id": "fw-001",
    "firmwareFilename": "firmware_v1.0.bin",
    "version": "1.0.0"
  },
  "explanation": "Device matches firmware rule fw-rule-001"
}
```
</details>

Response Codes: 200, 400, 401

---

## DCM Test Page
Test DCM rule evaluation:

**POST** `http://<host>:<port>/dcm/testpage`

<details>
<summary><strong>Request Body:</strong> Test parameters</summary>

```json
{
  "parameters": {
    "mac": "00:11:22:33:44:55",
    "model": "STB_MODEL_X",
    "env": "PROD"
  },
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> DCM test result</summary>

```json
{
  "result": "FORMULA_MATCH",
  "dcmSettings": {
    "logUploadSettings": {
      "name": "Production Log Upload",
      "uploadOnReboot": true
    }
  },
  "explanation": "Device matches DCM formula formula-001"
}
```
</details>

Response Codes: 200, 400, 401

---

## Settings Test Page
Test setting rule evaluation:

**POST** `http://<host>:<port>/settings/testpage`

<details>
<summary><strong>Request Body:</strong> Test parameters</summary>

```json
{
  "parameters": {
    "mac": "00:11:22:33:44:55",
    "model": "STB_MODEL_X",
    "env": "PROD"
  },
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Settings test result</summary>

```json
{
  "result": "RULE_MATCH",
  "settingsProfile": {
    "id": "setting-profile-001",
    "settingProfileId": "wifi-settings",
    "properties": {
      "wifiEnabled": "true"
    }
  },
  "explanation": "Device matches setting rule setting-rule-001"
}
```
</details>

Response Codes: 200, 400, 401

---

## RFC Test Page
Test RFC feature rule evaluation:

**POST** `http://<host>:<port>/rfc/test`

<details>
<summary><strong>Request Body:</strong> Test parameters</summary>

```json
{
  "parameters": {
    "mac": "00:11:22:33:44:55",
    "model": "STB_MODEL_X",
    "env": "PROD"
  },
  "applicationType": "stb"
}
```
</details>

<details>
<summary><strong>Response Body:</strong> RFC test result</summary>

```json
{
  "result": "RULE_MATCH",
  "features": [
    {
      "id": "feature-001",
      "featureName": "WiFiDualBand",
      "enable": true,
      "configData": {
        "band2g": "enabled",
        "band5g": "enabled"
      }
    }
  ],
  "explanation": "Device matches feature rule feature-rule-001"
}
```
</details>

Response Codes: 200, 400, 401

---

# System Information APIs

## Get System Statistics
Get system statistics and information:

**GET** `http://<host>:<port>/stats`

<details>
<summary><strong>Response Body:</strong> System statistics</summary>

```json
{
  "totalFirmwareRules": 150,
  "totalFirmwareConfigs": 75,
  "totalFeatures": 25,
  "totalFeatureRules": 50,
  "totalDcmFormulas": 10,
  "totalDeviceSettings": 5,
  "cacheStatus": "ACTIVE",
  "lastRefreshTime": "2025-11-05T10:00:00Z"
}
```
</details>

Response Codes: 200, 401

---

## Refresh All Caches
Refresh all system caches:

**GET** `http://<host>:<port>/info/refreshAll`

<details>
<summary><strong>Response Body:</strong> Cache refresh result</summary>

```json
{
  "status": "SUCCESS",
  "message": "All caches refreshed successfully",
  "refreshedTables": ["firmwareconfig", "firmwarerule", "feature", "featurerule", "dcmrule"],
  "timestamp": "2025-11-05T10:05:00Z"
}
```
</details>

Response Codes: 200, 401

---

## Get Table Information
Get information about specific database table:

**GET** `http://<host>:<port>/info/tables/{tableName}`

<details>
<summary><strong>Response Body:</strong> Table information</summary>

```json
{
  "tableName": "firmwareconfig",
  "rowCount": 75,
  "lastModified": "2025-11-05T09:30:00Z",
  "cacheStatus": "LOADED",
  "sizeBytes": 524288
}
```
</details>

Response Codes: 200, 404, 401

---

# Change Management APIs

## Get All Changes
Get all pending changes:

**GET** `http://<host>:<port>/change/all`

<details>
<summary><strong>Response Body:</strong> Array of change objects</summary>

```json
[
  {
    "id": "change-001",
    "entityType": "TELEMETRY_PROFILE",
    "entityId": "telemetry-profile-001",
    "operation": "UPDATE",
    "oldEntity": {...},
    "newEntity": {...},
    "author": "admin",
    "created": "2025-11-05T09:00:00Z"
  }
]
```
</details>

Response Codes: 200, 401

---

## Approve Change
Approve a pending change:

**GET** `http://<host>:<port>/change/approve/{changeId}`

<details>
<summary><strong>Response Body:</strong> Approval result</summary>

```json
{
  "approved": true,
  "changeId": "change-001",
  "approvedBy": "admin",
  "approvedAt": "2025-11-05T10:00:00Z"
}
```
</details>

Response Codes: 200, 404, 401

---

## Cancel Change
Cancel a pending change:

**GET** `http://<host>:<port>/change/cancel/{changeId}`

<details>
<summary><strong>Response Body:</strong> Cancellation result</summary>

```json
{
  "cancelled": true,
  "changeId": "change-001",
  "cancelledBy": "admin",
  "cancelledAt": "2025-11-05T10:00:00Z"
}
```
</details>

Response Codes: 200, 404, 401

---

# Application Settings APIs

## Get Application Settings
Get current application settings:

**GET** `http://<host>:<port>/appsettings`

<details>
<summary><strong>Response Body:</strong> Application settings object</summary>

```json
{
  "applicationName": "XconfAdmin",
  "version": "1.0.0",
  "maxFirmwareRuleSize": 1000,
  "enableChangeApproval": true,
  "defaultApplicationType": "stb",
  "supportedApplicationTypes": ["stb", "rdkv"]
}
```
</details>

Response Codes: 200, 401

---

## Update Application Settings
Update application settings:

**PUT** `http://<host>:<port>/appsettings`

<details>
<summary><strong>Request Body:</strong> Application settings update data</summary>

```json
{
  "applicationName": "XconfAdmin",
  "version": "1.1.0",
  "maxFirmwareRuleSize": 1200,
  "enableChangeApproval": true,
  "defaultApplicationType": "stb",
  "supportedApplicationTypes": ["stb", "rdkv", "rdkc"]
}
```
</details>

<details>
<summary><strong>Response Body:</strong> Updated application settings</summary>

```json
{
  "applicationName": "XconfAdmin",
  "version": "1.1.0", 
  "maxFirmwareRuleSize": 1200,
  "enableChangeApproval": true,
  "defaultApplicationType": "stb",
  "supportedApplicationTypes": ["stb", "rdkv", "rdkc"]
}
```
</details>

Response Codes: 200, 400, 401

---

# Error Handling

## Common Error Responses

### 400 Bad Request
```json
{
  "status": 400,
  "error": "Bad Request",
  "message": "Invalid request parameters",
  "timestamp": "2025-11-05T10:00:00Z"
}
```

### 401 Unauthorized
```json
{
  "status": 401,
  "error": "Unauthorized", 
  "message": "Authentication required",
  "timestamp": "2025-11-05T10:00:00Z"
}
```

### 403 Forbidden
```json
{
  "status": 403,
  "error": "Forbidden",
  "message": "Insufficient permissions",
  "timestamp": "2025-11-05T10:00:00Z"
}
```

### 404 Not Found
```json
{
  "status": 404,
  "error": "Not Found",
  "message": "Resource not found",
  "timestamp": "2025-11-05T10:00:00Z"
}
```

### 500 Internal Server Error
```json
{
  "status": 500,
  "error": "Internal Server Error",
  "message": "An unexpected error occurred",
  "timestamp": "2025-11-05T10:00:00Z"
}
```

---

# Request/Response Headers

## Common Request Headers
- `Content-Type: application/json`
- `Accept: application/json`
- `Authorization: Bearer <jwt-token>` (for authenticated endpoints)

## Common Response Headers
- `Content-Type: application/json`
- `Cache-Control: no-cache`
- `X-Request-ID: <unique-request-id>`

---

# Rate Limiting
All endpoints are subject to rate limiting:
- **Rate Limit**: 100 requests per minute per IP
- **Burst Limit**: 20 requests per second
- **Headers**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

---

# Pagination
Many list endpoints support pagination via query parameters:
- `pageNumber`: Page number (starting from 1)
- `pageSize`: Number of items per page (default: 50, max: 200)
- `sort`: Sort field (optional)
- `sortOrder`: Sort direction (`ASC` or `DESC`)

Example: `GET /model?pageNumber=1&pageSize=10&sort=name&sortOrder=ASC`

---

# Filtering
List endpoints support filtering via POST to `/filtered` sub-endpoints:

<details>
<summary><strong>Example Filter Request:</strong> Generic filter structure</summary>

```json
{
  "searchContext": {
    "applicationType": "stb"
  },
  "pageNumber": 1,
  "pageSize": 10,
  "sortOrder": "ASC"
}
```
</details>

---

*This documentation covers the core XconfAdmin API endpoints. For additional details on specific data structures and validation rules, refer to the service implementation or contact the development team.*