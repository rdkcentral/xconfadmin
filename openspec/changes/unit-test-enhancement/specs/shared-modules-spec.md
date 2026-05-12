# Shared Modules Specification

## Overview

The `shared/` package contains shared business logic, entity types, and utility functions used across multiple modules.

---

## Module Structure

```
shared/
├── Root Level
│   ├── coretypes.go                    # Core type definitions
│   ├── coretypes_test.go
│   ├── coretypes_clone_test.go
│   ├── coretypes_additional_test.go
│   ├── percentage_service.go           # Percentage calculations
│   ├── percentage_service_test.go
│   └── percentage_service_extra_test.go
│
├── estbfirmware/                        # ESTB firmware logic
│   ├── firmware_config.go
│   ├── firmware_config_test.go
│   ├── config_change_logs.go           # Uses GetListingDao!
│   ├── config_change_logs_test.go
│   ├── estb_firmware_context.go
│   ├── estb_firmware_context_test.go
│   ├── estb_converters.go
│   ├── estb_converters_test.go
│   ├── singleton_filter.go
│   ├── singleton_filter_test.go
│   ├── time_filter.go
│   ├── time_filter_test.go
│   ├── percent_filter.go
│   ├── percent_filter_test.go
│   ├── ip_filter.go
│   ├── ip_filter_test.go
│   ├── reboot_immediately_filter.go
│   ├── reboot_immediately_filter_test.go
│   └── estbfirmware_unit_test.go
│
├── firmware/                            # Firmware types
│   ├── firmwarerule.go
│   ├── firmwarerule_test.go
│   └── firmware_unit_test.go
│
├── logupload/                           # Log upload types
│   ├── logupload.go
│   ├── logupload_test.go
│   ├── telemetry_profile.go
│   ├── telemetry_profile_test.go
│   ├── permanent_profile.go
│   ├── permanent_profile_test.go
│   └── utils_test.go
│
├── rfc/                                 # RFC/Feature types
│   ├── feature.go
│   ├── feature_test.go
│   ├── feature_rule.go
│   ├── feature_rule_test.go
│   └── feature_predicate_test.go
│
└── change/                              # Change types
    ├── change.go                        # Uses GetSimpleDao!
    └── change_test.go
```

---

## DAO Usage Summary

| Package | File | DAO Type | Tables |
|---------|------|----------|--------|
| shared/estbfirmware | config_change_logs.go | **GetListingDao** | TABLE_LOGS |
| shared/estbfirmware | firmware_config.go | GetCachedSimpleDao | TABLE_FIRMWARE_CONFIGS |
| shared/estbfirmware | singleton_filter.go | GetCachedSimpleDao | TABLE_SINGLETON_FILTER_VALUE |
| shared/estbfirmware | percent_filter.go | GetCachedSimpleDao | TABLE_PERCENT_FILTER |
| shared/estbfirmware | ip_filter.go | GetCachedSimpleDao | TABLE_IP_FILTER |
| shared/firmware | firmwarerule.go | GetCachedSimpleDao | TABLE_FIRMWARE_RULES |
| shared/logupload | logupload.go | GetCachedSimpleDao | TABLE_LOG_FILES, TABLE_LOG_FILE_GROUPS |
| shared/logupload | telemetry_profile.go | GetCachedSimpleDao | TABLE_TELEMETRY_* |
| shared/rfc | feature.go | GetCachedSimpleDao | TABLE_FEATURES, TABLE_FEATURE_CONTROL_RULES |
| shared/change | change.go | **GetSimpleDao** | TABLE_XCONF_CHANGE, TABLE_TELEMETRY_* |

---

## shared/estbfirmware/ Specification

### config_change_logs.go

**CRITICAL**: Uses `GetListingDao()` - needs MockListingDao!

```go
// Current usage
db.GetListingDao().SetOne(db.TABLE_LOGS, key, logEntry)
db.GetListingDao().GetRange(db.TABLE_LOGS, startKey, endKey, maxResults)
```

**Tables**: TABLE_LOGS

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-ESTB-001 | Write log entry | Success | 🔲 Needs mock |
| TC-ESTB-002 | Get log range | Log entries | 🔲 Needs mock |
| TC-ESTB-003 | Empty range | Empty list | 🔲 Needs mock |

### Filter Types

| Filter | Table | Entity |
|--------|-------|--------|
| SingletonFilter | TABLE_SINGLETON_FILTER_VALUE | SingletonFilterValue |
| PercentFilter | TABLE_PERCENT_FILTER | PercentFilter |
| IpFilter | TABLE_IP_FILTER | IpFilter |
| TimeFilter | TABLE_TIME_FILTER | TimeFilter |

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-FLT-001 | Apply singleton filter | Filtered result | ⚠️ Needs refactor |
| TC-FLT-002 | Apply percent filter | Percent match | ⚠️ Needs refactor |
| TC-FLT-003 | Apply IP filter | IP match | ⚠️ Needs refactor |
| TC-FLT-004 | Apply time filter | Time match | ⚠️ Needs refactor |

---

## shared/firmware/ Specification

### firmwarerule.go

**Tables**: TABLE_FIRMWARE_RULES

**Functions**:
- `GetFirmwareRule(id string)` - Get rule by ID
- `GetAllFirmwareRules()` - Get all rules
- `SetFirmwareRule(rule)` - Create/update rule
- `DeleteFirmwareRule(id)` - Delete rule

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-FW-001 | Get firmware rule | Rule entity | ⚠️ Needs mock |
| TC-FW-002 | Get all rules | Rule list | ⚠️ Needs mock |
| TC-FW-003 | Set firmware rule | Success | ⚠️ Needs mock |
| TC-FW-004 | Delete rule | Success | ⚠️ Needs mock |

---

## shared/logupload/ Specification

### logupload.go

**Tables**: 
- TABLE_LOG_FILES
- TABLE_LOG_FILE_GROUPS

**Functions**:
- `SetLogFile(id, logFile)` - Store log file
- `GetAllLogFileGroups(size)` - Get log file groups

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-LOG-001 | Set log file | Success | ⚠️ Needs mock |
| TC-LOG-002 | Get log file groups | Group list | ⚠️ Needs mock |

### telemetry_profile.go

**Tables**:
- TABLE_PERMANENT_TELEMETRY_PROFILES
- TABLE_TELEMETRY_TWO_PROFILES
- TABLE_TELEMETRY_PROFILES
- TABLE_TELEMETRY_RULES
- TABLE_TELEMETRY_TWO_RULES

**Functions**:
- `SetPermanentTelemetryProfile(id, profile)` - Store profile
- `DeletePermanentTelemetryProfile(id)` - Delete profile
- `GetAllPermanentTelemetryProfiles()` - Get all profiles
- `GetAllTelemetryTwoProfiles()` - Get v2 profiles
- `GetOneTelemetryTwoProfile(id)` - Get v2 profile by ID
- `SetOneTelemetryTwoProfile(profile)` - Store v2 profile
- `DeleteOneTelemetryTwoProfile(id)` - Delete v2 profile

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-TEL-001 | Set permanent profile | Success | ⚠️ Needs mock |
| TC-TEL-002 | Get all permanent | Profile list | ⚠️ Needs mock |
| TC-TEL-003 | Set v2 profile | Success | ⚠️ Needs mock |
| TC-TEL-004 | Get v2 profile | Profile entity | ⚠️ Needs mock |
| TC-TEL-005 | Delete v2 profile | Success | ⚠️ Needs mock |

---

## shared/rfc/ Specification

### feature.go

**Tables**:
- TABLE_FEATURES
- TABLE_FEATURE_CONTROL_RULES

**Functions**:
- `GetAllFeatures()` - Get all features
- `DeleteFeature(id)` - Delete feature
- `SetFeature(feature)` - Create/update feature
- `GetFeatureRule(id)` - Get feature rule
- `SetFeatureRule(id, rule)` - Set feature rule
- `DeleteFeatureRule(id)` - Delete feature rule

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-RFC-001 | Get all features | Feature list | ⚠️ Needs mock |
| TC-RFC-002 | Set feature | Success | ⚠️ Needs mock |
| TC-RFC-003 | Delete feature | Success | ⚠️ Needs mock |
| TC-RFC-004 | Get feature rule | Rule entity | ⚠️ Needs mock |
| TC-RFC-005 | Set feature rule | Success | ⚠️ Needs mock |
| TC-RFC-006 | Delete feature rule | Success | ⚠️ Needs mock |

---

## shared/change/ Specification

### change.go

**CRITICAL**: Uses `GetSimpleDao()` NOT `GetCachedSimpleDao()`!

**Tables**:
- TABLE_XCONF_CHANGE
- TABLE_XCONF_APPROVED_CHANGE
- TABLE_TELEMETRY_CHANGES
- TABLE_TELEMETRY_APPROVED_CHANGES
- TABLE_TELEMETRY_TWO_CHANGES
- TABLE_TELEMETRY_APPROVED_TWO_CHANGES

**Functions**:
- `GetChange(id)` - Get change by ID
- `GetAllPendingChanges()` - Get all pending
- `CreateChange(change)` - Create change
- `ApproveChange(id)` - Approve change
- `RejectChange(id)` - Reject change

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-CHG-001 | Get change | Change entity | 🔲 Needs SimpleDao mock |
| TC-CHG-002 | Get all pending | Change list | 🔲 Needs SimpleDao mock |
| TC-CHG-003 | Create change | Success | 🔲 Needs SimpleDao mock |
| TC-CHG-004 | Approve change | Success | 🔲 Needs SimpleDao mock |
| TC-CHG-005 | Reject change | Success | 🔲 Needs SimpleDao mock |

---

## Files to Create

### shared/estbfirmware/test_utils.go

```go
package estbfirmware

import (
    "os"
    "testing"
    "github.com/rdkcentral/xconfadmin/adminapi/dcm/mocks"
)

var mockDao *mocks.MockCachedSimpleDao
var mockListingDao *mocks.MockListingDao

func IsMockDatabaseEnabled() bool {
    return os.Getenv("USE_MOCK_DB") == "true"
}

func InitMockDatabase() {
    mockDao = mocks.NewMockCachedSimpleDao()
    mockListingDao = mocks.NewMockListingDao()
    db.SetMockCachedSimpleDao(mockDao)
    db.SetMockListingDao(mockListingDao)
}

func ClearMockDatabase() {
    if mockDao != nil {
        mockDao.Clear()
    }
    if mockListingDao != nil {
        mockListingDao.Clear()
    }
}
```

### shared/firmware/test_utils.go

```go
package firmware

// Same pattern as above
```

### shared/logupload/test_utils.go

```go
package logupload

// Same pattern as above
```

### shared/rfc/test_utils.go

```go
package rfc

// Same pattern as above
```

### shared/change/test_utils.go

```go
package change

import (
    "os"
    "testing"
    "github.com/rdkcentral/xconfadmin/adminapi/dcm/mocks"
)

var mockSimpleDao *mocks.MockSimpleDao

func IsMockDatabaseEnabled() bool {
    return os.Getenv("USE_MOCK_DB") == "true"
}

func InitMockDatabase() {
    mockSimpleDao = mocks.NewMockSimpleDao()
    db.SetMockSimpleDao(mockSimpleDao)
}

func ClearMockDatabase() {
    if mockSimpleDao != nil {
        mockSimpleDao.Clear()
    }
}
```

---

## Test Data Fixtures

```go
// Firmware Rule fixture
func NewTestFirmwareRule(id string) *firmware.FirmwareRule {
    return &firmware.FirmwareRule{
        ID:              id,
        Name:            "Test_Rule_" + id[:8],
        Type:            "MODEL_RULE",
        ApplicationType: "stb",
    }
}

// Feature fixture
func NewTestFeature(id string) *rfc.Feature {
    return &rfc.Feature{
        ID:          id,
        Name:        "Test_Feature_" + id[:8],
        FeatureName: "test.feature." + id[:8],
        Effective:   true,
    }
}

// Feature Rule fixture
func NewTestFeatureRule(id string, featureIds []string) *rfc.FeatureRule {
    return &rfc.FeatureRule{
        Id:              id,
        Name:            "Test_FeatureRule_" + id[:8],
        FeatureIds:      featureIds,
        ApplicationType: "stb",
    }
}

// Log Entry fixture
func NewTestLogEntry(key string) *estbfirmware.ConfigChangeLog {
    return &estbfirmware.ConfigChangeLog{
        Key:       key,
        Timestamp: time.Now().UnixMilli(),
        Message:   "Test log entry",
    }
}
```

---

## Coverage Goals

| Package | Current | Target | Priority |
|---------|---------|--------|----------|
| shared/estbfirmware | ~55% | 85% | P1 |
| shared/firmware | ~50% | 85% | P2 |
| shared/logupload | ~60% | 85% | P2 |
| shared/rfc | ~55% | 85% | P2 |
| shared/change | ~45% | 85% | P2 |
| shared/ (root) | ~65% | 85% | P2 |

---

## Test Execution Commands

```bash
# Run all shared module tests
USE_MOCK_DB=true go test -v ./shared/... -count=1

# Run specific package
USE_MOCK_DB=true go test -v ./shared/estbfirmware/... -count=1
USE_MOCK_DB=true go test -v ./shared/firmware/... -count=1
USE_MOCK_DB=true go test -v ./shared/logupload/... -count=1
USE_MOCK_DB=true go test -v ./shared/rfc/... -count=1
USE_MOCK_DB=true go test -v ./shared/change/... -count=1

# Run with coverage
USE_MOCK_DB=true go test ./shared/... -coverprofile=shared.out
go tool cover -func=shared.out
```
