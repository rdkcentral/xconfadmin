# Supporting Modules Specification

## Overview

This document covers the remaining modules that require test enhancement:
- Setting Module
- RFC/Feature Module
- Canary Module
- Auth Module
- XCRP Module
- Firmware Module
- Configuration IP-MacRule Module
- Tagging API Module
- Common Package

---

## Setting Module

**Package**: `adminapi/setting/`  
**Priority**: P1 - High  
**Status**: No mocks, no TestMain

### Architecture

```
adminapi/setting/
├── setting_profile_controller.go
├── setting_profile_controller_test.go
├── setting_profile_service.go
├── setting_profile_service_test.go
├── setting_rule_controller.go
├── setting_rule_controller_test.go
├── setting_rule_service.go
└── setting_rule_service_test.go
```

### Database Tables

| Table | Operations | Entity |
|-------|------------|--------|
| TABLE_SETTING_PROFILES | CRUD | `logupload.SettingProfile` |
| TABLE_SETTING_RULES | CRUD | `logupload.SettingRule` |

### Use Cases

| ID | Use Case | Description |
|----|----------|-------------|
| UC-SET-001 | Setting Profile CRUD | Create/Read/Update/Delete profiles |
| UC-SET-002 | Setting Rule CRUD | Create/Read/Update/Delete rules |

### Test Scenarios

| ID | Scenario | Expected | Status |
|----|----------|----------|--------|
| TC-SET-001 | Create setting profile | 201 Created | ⚠️ Needs mock |
| TC-SET-002 | Get all profiles | 200 OK + list | ⚠️ Needs mock |
| TC-SET-003 | Create setting rule | 201 Created | ⚠️ Needs mock |
| TC-SET-004 | Get all rules | 200 OK + list | ⚠️ Needs mock |

### Files to Create

- `adminapi/setting/test_utils.go`

---

## RFC/Feature Module

**Package**: `adminapi/rfc/feature/`  
**Priority**: P1 - High  
**Status**: Has TestMain, needs mock enhancement

### Architecture

```
adminapi/rfc/feature/
├── feature_handler.go
├── feature_handler_test.go           # Has TestMain
├── feature_control_settings.go
├── feature_control_settings_test.go
├── feature_test_helpers.go
└── feature_test_helpers_test.go
```

### Database Tables

| Table | Operations | Entity |
|-------|------------|--------|
| TABLE_FEATURES | CRUD | `rfc.Feature` |
| TABLE_FEATURE_CONTROL_RULES | CRUD | `rfc.FeatureRule` |

### Use Cases

| ID | Use Case | Description |
|----|----------|-------------|
| UC-RFC-001 | Feature CRUD | Manage features |
| UC-RFC-002 | Feature Rule CRUD | Manage feature rules |
| UC-RFC-003 | Feature Control Settings | Configure feature controls |

### Test Scenarios

| ID | Scenario | Expected | Status |
|----|----------|----------|--------|
| TC-RFC-001 | Create feature | 201 Created | ⚠️ Needs refactor |
| TC-RFC-002 | Get all features | 200 OK + list | ⚠️ Needs refactor |
| TC-RFC-003 | Create feature rule | 201 Created | ⚠️ Needs refactor |
| TC-RFC-004 | Get feature rule | 200 OK + entity | ⚠️ Needs refactor |

### Files to Create

- `adminapi/rfc/feature/test_utils.go`

---

## Canary Module

**Package**: `adminapi/canary/`  
**Priority**: P1 - High  
**Status**: No mocks, no TestMain

### Architecture

```
adminapi/canary/
├── canary_settings_handler.go
├── canary_settings_handler_test.go
├── canary_settings_service.go
└── canary_settings_service_test.go
```

### Database Tables

| Table | Operations | Entity |
|-------|------------|--------|
| TABLE_CANARY_SETTINGS | CRUD | Canary settings |

### Use Cases

| ID | Use Case | Description |
|----|----------|-------------|
| UC-CAN-001 | Canary Settings CRUD | Manage canary deployment settings |

### Test Scenarios

| ID | Scenario | Expected | Status |
|----|----------|----------|--------|
| TC-CAN-001 | Create canary setting | 201 Created | ⚠️ Needs mock |
| TC-CAN-002 | Get canary settings | 200 OK + list | ⚠️ Needs mock |
| TC-CAN-003 | Update canary setting | 200 OK | ⚠️ Needs mock |
| TC-CAN-004 | Delete canary setting | 200 OK | ⚠️ Needs mock |

### Files to Create

- `adminapi/canary/test_utils.go`

---

## Auth Module

**Package**: `adminapi/auth/`  
**Priority**: P1 - High  
**Status**: No mocks, panics without DB

### Architecture

```
adminapi/auth/
├── idp_service_handler.go
└── idp_service_handler_test.go
```

### Current Issue

Tests panic without real DB because WebconfigServer initialization requires DB connection.

### Test Scenarios

| ID | Scenario | Expected | Status |
|----|----------|----------|--------|
| TC-AUTH-001 | IDP service call | Valid response | ⚠️ Needs mock |
| TC-AUTH-002 | Invalid token | 401 Unauthorized | ⚠️ Needs mock |

### Files to Create

- `adminapi/auth/test_utils.go`

---

## XCRP Module

**Package**: `adminapi/xcrp/`  
**Priority**: P2 - Medium  
**Status**: No mocks, no TestMain

### Architecture

```
adminapi/xcrp/
├── recooking_lockdown_settings_handler.go
├── recooking_lockdown_settings_handler_test.go
├── recooking_status_handler.go
└── recooking_status_handler_test.go
```

### Use Cases

| ID | Use Case | Description |
|----|----------|-------------|
| UC-XCRP-001 | Recooking Lockdown | Manage lockdown settings |
| UC-XCRP-002 | Recooking Status | Check recooking status |

### Test Scenarios

| ID | Scenario | Expected | Status |
|----|----------|----------|--------|
| TC-XCRP-001 | Get lockdown settings | 200 OK | ⚠️ Needs mock |
| TC-XCRP-002 | Set lockdown | 200 OK | ⚠️ Needs mock |
| TC-XCRP-003 | Get recooking status | 200 OK | ⚠️ Needs mock |

### Files to Create

- `adminapi/xcrp/test_utils.go`

---

## Firmware Module

**Package**: `adminapi/firmware/`  
**Priority**: P2 - Medium  
**Status**: No mocks, no TestMain

### Architecture

```
adminapi/firmware/
├── firmware_test_page_controller.go
└── firmware_test_page_controller_test.go
```

### Use Cases

| ID | Use Case | Description |
|----|----------|-------------|
| UC-FW-001 | Firmware Test Page | Test firmware configuration |

### Test Scenarios

| ID | Scenario | Expected | Status |
|----|----------|----------|--------|
| TC-FW-001 | Load test page | 200 OK | ⚠️ Needs mock |
| TC-FW-002 | Test firmware config | Valid result | ⚠️ Needs mock |

### Files to Create

- `adminapi/firmware/test_utils.go`

---

## Configuration IP-MacRule Module

**Package**: `adminapi/configuration/ip-macrule/`  
**Priority**: P2 - Medium  
**Status**: No mocks, no TestMain

### Architecture

```
adminapi/configuration/ip-macrule/
├── ip_mac_ruleconfig_handler.go
└── ip_mac_ruleconfig_handler_test.go
```

### Use Cases

| ID | Use Case | Description |
|----|----------|-------------|
| UC-CFG-001 | IP-MAC Rule Config | Configure IP/MAC rules |

### Test Scenarios

| ID | Scenario | Expected | Status |
|----|----------|----------|--------|
| TC-CFG-001 | Get IP-MAC rule config | 200 OK | ⚠️ Needs mock |
| TC-CFG-002 | Set IP-MAC rule config | 200 OK | ⚠️ Needs mock |

### Files to Create

- `adminapi/configuration/ip-macrule/test_utils.go`

---

## Tagging API Module

**Package**: `taggingapi/`  
**Priority**: P1 - High  
**Status**: No mocks, no TestMain

### Architecture

```
taggingapi/
├── router.go
├── config/
│   ├── tag_config.go
│   └── tag_config_test.go
├── percentage/
│   ├── percentage_service.go
│   └── percentage_service_test.go
└── tag/
    ├── tag_handler.go
    ├── tag_handler_test.go
    ├── tag_service.go
    ├── tag_service_test.go
    ├── tag_member_service.go
    ├── tag_member_service_test.go
    ├── tag_normalization_service.go
    ├── tag_normalization_service_test.go
    └── tag_member_benchmark_test.go
```

### Database Tables

| Table | Operations | Entity |
|-------|------------|--------|
| TABLE_TAGS | CRUD | `tag.Tag` |
| TABLE_TAG_MEMBERS | CRUD | `tag.TagMember` |

### Use Cases

| ID | Use Case | Description |
|----|----------|-------------|
| UC-TAG-001 | Tag Management | Create/manage tags |
| UC-TAG-002 | Tag Member Management | Manage tag members |
| UC-TAG-003 | Tag Normalization | Normalize tag names |
| UC-TAG-004 | Percentage Service | Calculate percentages |

### Test Scenarios

| ID | Scenario | Expected | Status |
|----|----------|----------|--------|
| TC-TAG-001 | Create tag | 201 Created | ⚠️ Needs mock |
| TC-TAG-002 | Get all tags | 200 OK + list | ⚠️ Needs mock |
| TC-TAG-003 | Add tag member | 200 OK | ⚠️ Needs mock |
| TC-TAG-004 | Get tag members | 200 OK + list | ⚠️ Needs mock |
| TC-TAG-005 | Normalize tag | Normalized name | ⚠️ Needs mock |
| TC-TAG-006 | Calculate percentage | Valid percentage | ⚠️ Needs mock |

### Files to Create

- `taggingapi/tag/test_utils.go`
- `taggingapi/config/test_utils.go`
- `taggingapi/percentage/test_utils.go`

---

## Common Package

**Package**: `common/`  
**Priority**: P1 - High  
**Status**: Uses DB extensively, no test_utils.go

### Architecture

```
common/
├── const_var.go
├── const_var_test.go
├── error.go
├── error_test.go
├── server_config.go
├── server_config_test.go
├── struct.go                  # Uses GetCachedSimpleDao!
└── struct_test.go
```

### Database Tables Used in struct.go

| Table | Operations | Entity |
|-------|------------|--------|
| TABLE_APP_SETTINGS | CRUD | `common.ApplicationSetting` |
| TABLE_DCM_RULES | Read | `logupload.DCMGenericRule` |
| TABLE_ENVIRONMENTS | CRUD | `shared.Environment` |
| TABLE_MODELS | CRUD | `shared.Model` |

### Use Cases

| ID | Use Case | Description |
|----|----------|-------------|
| UC-CMN-001 | App Settings | Manage application settings |
| UC-CMN-002 | Get DCM Rules | Retrieve DCM rules |
| UC-CMN-003 | Environment CRUD | Manage environments |
| UC-CMN-004 | Model CRUD | Manage models |

### Test Scenarios

| ID | Scenario | Expected | Status |
|----|----------|----------|--------|
| TC-CMN-001 | Set app setting | Success | ⚠️ Needs mock |
| TC-CMN-002 | Get app setting | Setting value | ⚠️ Needs mock |
| TC-CMN-003 | Get all DCM rules | Rule list | ⚠️ Needs mock |
| TC-CMN-004 | Get environment | Environment | ⚠️ Needs mock |
| TC-CMN-005 | Set environment | Success | ⚠️ Needs mock |
| TC-CMN-006 | Get model | Model | ⚠️ Needs mock |
| TC-CMN-007 | Set model | Success | ⚠️ Needs mock |

### Files to Create

- `common/test_utils.go`

---

## Summary of Files to Create

| Module | File | Priority |
|--------|------|----------|
| adminapi/setting | test_utils.go | P1 |
| adminapi/rfc/feature | test_utils.go | P1 |
| adminapi/canary | test_utils.go | P1 |
| adminapi/auth | test_utils.go | P1 |
| adminapi/xcrp | test_utils.go | P2 |
| adminapi/firmware | test_utils.go | P2 |
| adminapi/configuration/ip-macrule | test_utils.go | P2 |
| taggingapi/tag | test_utils.go | P1 |
| taggingapi/config | test_utils.go | P2 |
| taggingapi/percentage | test_utils.go | P2 |
| common | test_utils.go | P1 |

**Total**: 11 test_utils.go files to create

---

## Test Execution Commands

```bash
# Setting module
USE_MOCK_DB=true go test -v ./adminapi/setting/... -count=1

# RFC module
USE_MOCK_DB=true go test -v ./adminapi/rfc/... -count=1

# Canary module
USE_MOCK_DB=true go test -v ./adminapi/canary/... -count=1

# Auth module
USE_MOCK_DB=true go test -v ./adminapi/auth/... -count=1

# XCRP module
USE_MOCK_DB=true go test -v ./adminapi/xcrp/... -count=1

# Firmware module
USE_MOCK_DB=true go test -v ./adminapi/firmware/... -count=1

# Configuration module
USE_MOCK_DB=true go test -v ./adminapi/configuration/... -count=1

# Tagging API
USE_MOCK_DB=true go test -v ./taggingapi/... -count=1

# Common package
USE_MOCK_DB=true go test -v ./common/... -count=1
```
