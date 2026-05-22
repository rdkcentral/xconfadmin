# Database Tables Specification

## Overview

Complete mapping of all Cassandra tables used by xconfadmin, their entity types, and which modules access them.

---

## Tables by Category

### DCM (Device Configuration Management)

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_DEVICE_SETTINGS` | device_settings | `logupload.DeviceSettings` | CachedSimpleDao |
| `TABLE_VOD_SETTINGS` | vod_settings | `logupload.VodSettings` | CachedSimpleDao |
| `TABLE_LOG_UPLOAD_SETTINGS` | log_upload_settings | `logupload.LogUploadSettings` | CachedSimpleDao |
| `TABLE_UPLOAD_REPOSITORY` | upload_repository | `logupload.UploadRepository` | CachedSimpleDao |
| `TABLE_DCM_RULES` | dcm_rules | `logupload.DCMGenericRule` | CachedSimpleDao |
| `TABLE_LOG_FILES` | log_files | `logupload.LogFile` | CachedSimpleDao |
| `TABLE_LOG_FILE_GROUPS` | log_file_groups | `logupload.LogFileGroup` | CachedSimpleDao |
| `TABLE_LOG_FILE_LISTS` | log_file_lists | `logupload.LogFileList` | CachedSimpleDao |

### Core Configuration

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_MODELS` | models | `shared.Model` | CachedSimpleDao |
| `TABLE_ENVIRONMENTS` | environments | `shared.Environment` | CachedSimpleDao |
| `TABLE_APP_SETTINGS` | app_settings | `common.ApplicationSetting` | CachedSimpleDao |

### Firmware

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_FIRMWARE_CONFIGS` | firmware_configs | `firmware.FirmwareConfig` | CachedSimpleDao |
| `TABLE_FIRMWARE_RULES` | firmware_rules | `firmware.FirmwareRule` | CachedSimpleDao |
| `TABLE_FIRMWARE_RULE_TEMPLATES` | firmware_rule_templates | `firmware.FirmwareRuleTemplate` | CachedSimpleDao |

### Telemetry

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_TELEMETRY_PROFILES` | telemetry_profiles | `logupload.TelemetryProfile` | CachedSimpleDao |
| `TABLE_TELEMETRY_RULES` | telemetry_rules | `logupload.TelemetryRule` | CachedSimpleDao |
| `TABLE_PERMANENT_TELEMETRY_PROFILES` | permanent_telemetry_profiles | `logupload.PermanentTelemetryProfile` | CachedSimpleDao |
| `TABLE_TELEMETRY_TWO_PROFILES` | telemetry_two_profiles | `logupload.TelemetryTwoProfile` | CachedSimpleDao |
| `TABLE_TELEMETRY_TWO_RULES` | telemetry_two_rules | `logupload.TelemetryTwoRule` | CachedSimpleDao |

### Change Management

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_XCONF_CHANGE` | xconf_change | `change.Change` | **SimpleDao** |
| `TABLE_XCONF_APPROVED_CHANGE` | xconf_approved_change | `change.Change` | **SimpleDao** |
| `TABLE_TELEMETRY_CHANGES` | telemetry_changes | `change.Change` | **SimpleDao** |
| `TABLE_TELEMETRY_APPROVED_CHANGES` | telemetry_approved_changes | `change.Change` | **SimpleDao** |
| `TABLE_TELEMETRY_TWO_CHANGES` | telemetry_two_changes | `change.Change` | **SimpleDao** |
| `TABLE_TELEMETRY_APPROVED_TWO_CHANGES` | telemetry_approved_two_changes | `change.Change` | **SimpleDao** |

### Features/RFC

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_FEATURES` | features | `rfc.Feature` | CachedSimpleDao |
| `TABLE_FEATURE_CONTROL_RULES` | feature_control_rules | `rfc.FeatureRule` | CachedSimpleDao |

### Settings

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_SETTING_PROFILES` | setting_profiles | `logupload.SettingProfile` | CachedSimpleDao |
| `TABLE_SETTING_RULES` | setting_rules | `logupload.SettingRule` | CachedSimpleDao |

### Lists

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_NS_LISTS` | namespaced_lists | `shared.NamespacedList` | CachedSimpleDao |
| `TABLE_IP_ADDRESS_GROUPS` | ip_address_groups | `shared.IpAddressGroup` | CachedSimpleDao |
| `TABLE_MAC_LISTS` | mac_lists | `shared.MacList` | CachedSimpleDao |

### Filters

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_SINGLETON_FILTER_VALUE` | singleton_filter_value | `estbfirmware.SingletonFilterValue` | CachedSimpleDao |
| `TABLE_PERCENT_FILTER` | percent_filter | `estbfirmware.PercentFilter` | CachedSimpleDao |
| `TABLE_IP_FILTER` | ip_filter | `estbfirmware.IpFilter` | CachedSimpleDao |

### Logs

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_LOGS` | logs | Log entries | **ListingDao** |

### Tags

| Table Constant | Cassandra Table | Entity Type | DAO Type |
|----------------|-----------------|-------------|----------|
| `TABLE_TAGS` | tags | `tag.Tag` | CachedSimpleDao |
| `TABLE_TAG_MEMBERS` | tag_members | `tag.TagMember` | CachedSimpleDao |

---

## DAO Type Distribution

```
┌─────────────────────────────────────────────────────────────┐
│                      DAO TYPE USAGE                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  CachedSimpleDao                                            │
│  ════════════════════════════════════════════════════════   │
│  39 files use GetCachedSimpleDao()                          │
│  Tables: ALL except change tables and logs                   │
│  Features: Caching, fast reads                               │
│                                                              │
│  SimpleDao                                                   │
│  ════════════════════════════════════════════════════════   │
│  1 file uses GetSimpleDao(): shared/change/change.go        │
│  Tables: All *_CHANGE and *_APPROVED_CHANGE tables          │
│  Features: No caching, direct DB access                      │
│                                                              │
│  ListingDao                                                  │
│  ════════════════════════════════════════════════════════   │
│  1 file uses GetListingDao(): shared/estbfirmware/          │
│  Tables: TABLE_LOGS                                          │
│  Features: Range queries, time-series data                   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Table Usage by Module

### adminapi/dcm/

```go
// Production files
logrepo_settings_service.go:
  db.GetCachedSimpleDao().GetAllAsList(db.TABLE_UPLOAD_REPOSITORY, 0)
  db.GetCachedSimpleDao().GetOne(db.TABLE_UPLOAD_REPOSITORY, id)
  db.GetCachedSimpleDao().SetOne(db.TABLE_UPLOAD_REPOSITORY, id, entity)
  db.GetCachedSimpleDao().DeleteOne(db.TABLE_UPLOAD_REPOSITORY, id)

device_settings_service.go:
  db.GetCachedSimpleDao().GetAllAsList(db.TABLE_DEVICE_SETTINGS, 0)
  db.GetCachedSimpleDao().GetOne(db.TABLE_DEVICE_SETTINGS, id)
  db.GetCachedSimpleDao().SetOne(db.TABLE_DEVICE_SETTINGS, id, entity)
  db.GetCachedSimpleDao().DeleteOne(db.TABLE_DEVICE_SETTINGS, id)

vod_settings_service.go:
  db.GetCachedSimpleDao().GetAllAsList(db.TABLE_VOD_SETTINGS, 0)
  db.GetCachedSimpleDao().GetOne(db.TABLE_VOD_SETTINGS, id)
  db.GetCachedSimpleDao().SetOne(db.TABLE_VOD_SETTINGS, id, entity)
  db.GetCachedSimpleDao().DeleteOne(db.TABLE_VOD_SETTINGS, id)
```

### adminapi/queries/

```go
// Model operations
model_service.go:
  db.GetCachedSimpleDao().GetAllAsList(db.TABLE_MODELS, 0)
  db.GetCachedSimpleDao().GetOne(db.TABLE_MODELS, id)
  db.GetCachedSimpleDao().SetOne(db.TABLE_MODELS, id, entity)
  db.GetCachedSimpleDao().DeleteOne(db.TABLE_MODELS, id)

// Environment operations
environment_service.go:
  db.GetCachedSimpleDao().GetAllAsList(db.TABLE_ENVIRONMENTS, 0)
  db.GetCachedSimpleDao().GetOne(db.TABLE_ENVIRONMENTS, id)
  db.GetCachedSimpleDao().SetOne(db.TABLE_ENVIRONMENTS, id, entity)
  db.GetCachedSimpleDao().DeleteOne(db.TABLE_ENVIRONMENTS, id)

// Firmware config operations
firmware_config_service.go:
  db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FIRMWARE_CONFIGS, 0)
  db.GetCachedSimpleDao().GetOne(db.TABLE_FIRMWARE_CONFIGS, id)
  db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_CONFIGS, id, entity)
  db.GetCachedSimpleDao().DeleteOne(db.TABLE_FIRMWARE_CONFIGS, id)

// Firmware rule operations
firmware_rule_service.go:
  db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FIRMWARE_RULES, 0)
  db.GetCachedSimpleDao().GetOne(db.TABLE_FIRMWARE_RULES, id)
  db.GetCachedSimpleDao().SetOne(db.TABLE_FIRMWARE_RULES, id, entity)
  db.GetCachedSimpleDao().DeleteOne(db.TABLE_FIRMWARE_RULES, id)
```

### shared/change/

```go
// Uses GetSimpleDao (NOT GetCachedSimpleDao)
change.go:
  db.GetSimpleDao().GetOne(db.TABLE_XCONF_CHANGE, id)
  db.GetSimpleDao().SetOne(db.TABLE_XCONF_CHANGE, id, entity)
  db.GetSimpleDao().DeleteOne(db.TABLE_XCONF_CHANGE, id)
  db.GetSimpleDao().GetAllAsList(db.TABLE_XCONF_CHANGE, 0)
  
  db.GetSimpleDao().GetOne(db.TABLE_XCONF_APPROVED_CHANGE, id)
  db.GetSimpleDao().SetOne(db.TABLE_XCONF_APPROVED_CHANGE, id, entity)
  
  // Telemetry change tables
  db.GetSimpleDao().GetOne(db.TABLE_TELEMETRY_CHANGES, id)
  db.GetSimpleDao().GetOne(db.TABLE_TELEMETRY_APPROVED_CHANGES, id)
  db.GetSimpleDao().GetOne(db.TABLE_TELEMETRY_TWO_CHANGES, id)
  db.GetSimpleDao().GetOne(db.TABLE_TELEMETRY_APPROVED_TWO_CHANGES, id)
```

### shared/estbfirmware/

```go
// Uses GetListingDao for logs
config_change_logs.go:
  db.GetListingDao().SetOne(db.TABLE_LOGS, key, logEntry)
  db.GetListingDao().GetRange(db.TABLE_LOGS, startKey, endKey, maxResults)
```

### common/

```go
struct.go:
  db.GetCachedSimpleDao().GetAllAsList(db.TABLE_DCM_RULES, 0)
  db.GetCachedSimpleDao().GetOne(db.TABLE_DCM_RULES, id)
  db.GetCachedSimpleDao().GetAllAsList(db.TABLE_ENVIRONMENTS, 0)
  db.GetCachedSimpleDao().GetOne(db.TABLE_ENVIRONMENTS, id)
  db.GetCachedSimpleDao().GetAllAsList(db.TABLE_MODELS, 0)
  db.GetCachedSimpleDao().GetOne(db.TABLE_MODELS, id)
  db.GetCachedSimpleDao().SetOne(db.TABLE_APP_SETTINGS, id, entity)
  db.GetCachedSimpleDao().GetOne(db.TABLE_APP_SETTINGS, id)
```

---

## Entity Type Definitions

### Key Entities

```go
// logupload.DeviceSettings
type DeviceSettings struct {
    ID                    string   `json:"id"`
    Name                  string   `json:"name"`
    CheckOnReboot         bool     `json:"checkOnReboot"`
    SettingsAreActive     bool     `json:"settingsAreActive"`
    Schedule              Schedule `json:"schedule"`
    ApplicationType       string   `json:"applicationType"`
}

// shared.Model
type Model struct {
    ID          string `json:"id"`
    Description string `json:"description"`
}

// shared.Environment
type Environment struct {
    ID          string `json:"id"`
    Description string `json:"description"`
}

// firmware.FirmwareConfig
type FirmwareConfig struct {
    ID               string `json:"id"`
    Description      string `json:"description"`
    FirmwareFilename string `json:"firmwareFilename"`
    FirmwareVersion  string `json:"firmwareVersion"`
    ApplicationType  string `json:"applicationType"`
}

// firmware.FirmwareRule
type FirmwareRule struct {
    ID              string `json:"id"`
    Name            string `json:"name"`
    Type            string `json:"type"`
    Rule            Rule   `json:"rule"`
    ApplicationType string `json:"applicationType"`
}

// change.Change
type Change struct {
    ID             string      `json:"id"`
    EntityId       string      `json:"entityId"`
    EntityType     string      `json:"entityType"`
    Operation      string      `json:"operation"`
    OldEntity      interface{} `json:"oldEntity"`
    NewEntity      interface{} `json:"newEntity"`
    Author         string      `json:"author"`
    ApprovedUser   string      `json:"approvedUser"`
    Updated        int64       `json:"updated"`
}
```

---

## Mock Data Requirements

### Per-Table Test Data

| Table | Min Test Records | Fixture Function |
|-------|-----------------|------------------|
| DEVICE_SETTINGS | 5 | `NewTestDeviceSetting(id)` |
| VOD_SETTINGS | 3 | `NewTestVodSettings(id)` |
| LOG_UPLOAD_SETTINGS | 3 | `NewTestLogUploadSettings(id)` |
| UPLOAD_REPOSITORY | 2 | `NewTestUploadRepository(id)` |
| MODELS | 5 | `NewTestModel(id)` |
| ENVIRONMENTS | 3 | `NewTestEnvironment(id)` |
| FIRMWARE_CONFIGS | 5 | `NewTestFirmwareConfig(id)` |
| FIRMWARE_RULES | 10 | `NewTestFirmwareRule(id)` |
| TELEMETRY_PROFILES | 3 | `NewTestTelemetryProfile(id)` |
| TELEMETRY_RULES | 5 | `NewTestTelemetryRule(id)` |
| FEATURES | 5 | `NewTestFeature(id)` |
| FEATURE_CONTROL_RULES | 5 | `NewTestFeatureRule(id)` |
| XCONF_CHANGE | 10 | `NewTestChange(id)` |
| NS_LISTS | 5 | `NewTestNamespacedList(id)` |

---

## Cache Refresh Requirements

Tables that require cache refresh after modifications:

```go
// After SetOne or DeleteOne, refresh cache
func refreshTableCache(tableName string) error {
    return db.GetCachedSimpleDao().RefreshAll(tableName)
}

// Tables requiring refresh
var cachedTables = []string{
    db.TABLE_DEVICE_SETTINGS,
    db.TABLE_VOD_SETTINGS,
    db.TABLE_LOG_UPLOAD_SETTINGS,
    db.TABLE_UPLOAD_REPOSITORY,
    db.TABLE_MODELS,
    db.TABLE_ENVIRONMENTS,
    db.TABLE_FIRMWARE_CONFIGS,
    db.TABLE_FIRMWARE_RULES,
    db.TABLE_FEATURES,
    db.TABLE_FEATURE_CONTROL_RULES,
    // ... etc
}
```
