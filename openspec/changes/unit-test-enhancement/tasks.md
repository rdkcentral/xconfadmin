# Unit Test Enhancement - Tasks

## Overview

This document tracks all tasks for implementing dual-mode (mock/real DB) testing across xconfadmin modules.

**Verification Loop**: After each task, run:
```bash
# Mock mode
USE_MOCK_DB=true go test -run <TestFunctionName> -cover -coverprofile=mock.out ./path/to/package
go tool cover -func=mock.out | grep <FunctionUnderTest>

# Real mode (if DB available)
USE_MOCK_DB=false go test -run <TestFunctionName> -cover -coverprofile=real.out ./path/to/package
go tool cover -func=real.out | grep <FunctionUnderTest>
```

---

## Phase 1: Infrastructure Setup

### Task 1.1: Standardize Mock DAO Package
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/dcm/mocks/`

**Description**: Ensure mock DAO package is complete and reusable by all modules.

**Subtasks**:
- [ ] Verify `mock_dao.go` implements all `CachedSimpleDao` methods
- [ ] Verify `mock_database_client.go` implements `DatabaseClient` interface
- [ ] Add missing methods if any
- [ ] Add unit tests for mock implementations

**Files**:
- `adminapi/dcm/mocks/mock_dao.go`
- `adminapi/dcm/mocks/mock_dao_test.go`
- `adminapi/dcm/mocks/mock_database_client.go`
- `adminapi/dcm/mocks/mock_database_client_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 2: DCM Module Enhancement

### Task 2.1: Fix DCM TestMain Mock Initialization
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/dcm/`

**Description**: Fix TestMain to properly initialize mock DB before server creation.

**Current Issue**:
```go
// Current: Initializes mock AFTER server tries to connect to Cassandra
// This causes panic when USE_MOCK_DB=true
```

**Subtasks**:
- [ ] Move mock initialization before `NewWebconfigServer()`
- [ ] Ensure mock DB client is set before any DB operations
- [ ] Test with `USE_MOCK_DB=true`
- [ ] Test with `USE_MOCK_DB=false`

**Files**:
- `adminapi/dcm/dcmformula_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 2.2: DCM Device Settings - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/dcm/`

**Description**: Split `TestAllDeviceSettingsApis` into independent tests.

**Current Issue**:
```go
func TestAllDeviceSettingsApis(t *testing.T) {
    // One giant test doing CRUD - not idempotent
}
```

**Subtasks**:
- [ ] Create `TestCreateDeviceSetting`
- [ ] Create `TestGetDeviceSettingById`
- [ ] Create `TestGetAllDeviceSettings`
- [ ] Create `TestUpdateDeviceSetting`
- [ ] Create `TestDeleteDeviceSetting`
- [ ] Each test sets up own data and cleans up only its data
- [ ] Remove original `TestAllDeviceSettingsApis`

**Files**:
- `adminapi/dcm/device_settings_e2e_test.go`
- `adminapi/dcm/device_settings_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 2.3: DCM Vod Settings - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/dcm/`

**Subtasks**:
- [ ] Split e2e test into independent functions
- [ ] Add surgical cleanup for each test
- [ ] Verify mock mode support

**Files**:
- `adminapi/dcm/vod_settings_e2e_test.go`
- `adminapi/dcm/vod_settings_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 2.4: DCM LogRepo Settings - Fix Double Cleanup
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/dcm/`

**Description**: Fix double cleanup pattern in logrepo tests.

**Current Issue**:
```go
DeleteAllEntities()      // ← Remove this
defer DeleteAllEntities() // ← Keep only defer with surgical cleanup
```

**Subtasks**:
- [ ] Remove initial `DeleteAllEntities()` calls
- [ ] Replace `defer DeleteAllEntities()` with surgical cleanup
- [ ] Track inserted IDs and delete only those

**Files**:
- `adminapi/dcm/logrepo_settings_service_test.go`
- `adminapi/dcm/logrepo_settings_e2e_test.go`
- `adminapi/dcm/logrepo_settings_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 2.5: DCM LogUpload Settings - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/dcm/`

**Subtasks**:
- [ ] Split e2e test into independent functions
- [ ] Add surgical cleanup
- [ ] Verify mock mode

**Files**:
- `adminapi/dcm/logupload_settings_e2e_test.go`
- `adminapi/dcm/logupload_settings_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 2.6: DCM Formula Test - Fix Cleanup
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/dcm/`

**Subtasks**:
- [ ] Review all `DeleteAllEntities()` calls
- [ ] Replace with surgical cleanup
- [ ] Ensure idempotent tests

**Files**:
- `adminapi/dcm/dcmformula_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 2.7: DCM Test Page Controller - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P2 - Medium  
**Module**: `adminapi/dcm/`

**Files**:
- `adminapi/dcm/test_page_controller_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 3: Queries Module Enhancement

### Task 3.1: Queries TestMain - Add Mock Support
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/queries/`

**Subtasks**:
- [ ] Update TestMain with `USE_MOCK_DB` check
- [ ] Initialize mock before server creation
- [ ] Register all required table configs

**Files**:
- `adminapi/queries/queries_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.2: Queries Model Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/queries/`

**Subtasks**:
- [ ] Fix `model_handler_test.go`
- [ ] Fix `model_service_test.go`
- [ ] Fix `model_test.go`
- [ ] Fix `model_query_update_delete_test.go`
- [ ] Add surgical cleanup

**Files**:
- `adminapi/queries/model_handler_test.go`
- `adminapi/queries/model_service_test.go`
- `adminapi/queries/model_test.go`
- `adminapi/queries/model_query_update_delete_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.3: Queries Firmware Config Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/firmware_config_handler_test.go`
- `adminapi/queries/firmware_config_service_test.go`
- `adminapi/queries/firmware_config_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.4: Queries Firmware Rule Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/firmware_rule_handler_test.go`
- `adminapi/queries/firmware_rule_service_test.go`
- `adminapi/queries/firmware_rule_test.go`
- `adminapi/queries/firmware_rule_template_handler_test.go`
- `adminapi/queries/firmware_rule_template_handler_additional_test.go`
- `adminapi/queries/firmware_rule_template_service_test.go`
- `adminapi/queries/firmware_rule_report_page_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.5: Queries Feature Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/feature_handler.go` (for reference)
- `adminapi/queries/feature_entity_handler_test.go`
- `adminapi/queries/feature_entity_service_test.go`
- `adminapi/queries/feature_rule_handler_test.go`
- `adminapi/queries/feature_rule_service_test.go`
- `adminapi/queries/feature_service_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.6: Queries IP Address/MAC List Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/ip_address_group_service_test.go`
- `adminapi/queries/ipaddressgroup_maclist_handlers_test.go`
- `adminapi/queries/mac_rule_bean_handler_test.go`
- `adminapi/queries/maclist_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.7: Queries Filter Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/ips_filter_service_test.go`
- `adminapi/queries/location_filter_service_test.go`
- `adminapi/queries/percent_filter_service_test.go`
- `adminapi/queries/percentfilter_handler_test.go`
- `adminapi/queries/ri_filter_service_test.go`
- `adminapi/queries/time_filter_service_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.8: Queries Namespaced List Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/namespaced_list_handler_test.go`
- `adminapi/queries/namespaced_list_service_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.9: Queries Environment Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/environment_handler_test.go`
- `adminapi/queries/environment_service_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.10: Queries Percentage Bean Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/percentage_bean_service_test.go`
- `adminapi/queries/percentagebean_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.11: Queries AMV Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/amv_handler_test.go`
- `adminapi/queries/amv_service_test.go`
- `adminapi/queries/amv_test.go`
- `adminapi/queries/activation_minimum_version_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.12: Queries Log/Base/Common Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P2 - Medium  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/log_controller_test.go`
- `adminapi/queries/log_file_handler_test.go`
- `adminapi/queries/base_queries_controller_test.go`
- `adminapi/queries/baserule_validator_test.go`
- `adminapi/queries/common_test.go`
- `adminapi/queries/converter_test.go`
- `adminapi/queries/queries_handler_test.go`
- `adminapi/queries/queries_helper_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 3.13: Queries Additional/Coverage Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P2 - Medium  
**Module**: `adminapi/queries/`

**Files**:
- `adminapi/queries/additional_handler_test.go`
- `adminapi/queries/additional_service_test.go`
- `adminapi/queries/coverage_improvement_test.go`
- `adminapi/queries/simple_service_test.go`
- `adminapi/queries/firmware_simple_test.go`
- `adminapi/queries/firmwares_test.go`
- `adminapi/queries/firstreport_test.go`
- `adminapi/queries/prioritizable_test.go`
- `adminapi/queries/penetration_metrics_client_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 4: Telemetry Module Enhancement

### Task 4.1: Telemetry TestMain - Fix Mock Support
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/telemetry/`

**Subtasks**:
- [ ] Update TestMain in `telemetry_profile_handler_test.go`
- [ ] Initialize mock before server creation
- [ ] Verify all tests pass with mock

**Files**:
- `adminapi/telemetry/telemetry_profile_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 4.2: Telemetry Profile Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/telemetry/`

**Files**:
- `adminapi/telemetry/telemetry_profile_controller_test.go`
- `adminapi/telemetry/telemetry_profile_handler_test.go`
- `adminapi/telemetry/telemetry_profile_service_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 4.3: Telemetry Rule Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/telemetry/`

**Files**:
- `adminapi/telemetry/telemetry_rule_handler_test.go`
- `adminapi/telemetry/telemetry_v2_rule_service_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 4.4: Telemetry Two Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/telemetry/`

**Files**:
- `adminapi/telemetry/telemetry_two_dao_test.go`
- `adminapi/telemetry/telemetry_two_loguploader_handler_test.go`
- `adminapi/telemetry/telemetry_two_profile_handler_test.go`
- `adminapi/telemetry/telemetry_two_rule_hanlder_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 5: Change Module Enhancement

### Task 5.1: Change Module - Add test_utils.go
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/change/`

**Subtasks**:
- [ ] Create `test_utils.go` with mock/real switching
- [ ] Update TestMain in `change_handler_test.go`
- [ ] Verify tests pass with mock

**Files**:
- `adminapi/change/test_utils.go` (NEW)
- `adminapi/change/change_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 5.2: Change Handler/Service Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/change/`

**Files**:
- `adminapi/change/change_handler_test.go`
- `adminapi/change/change_service_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 5.3: Change Telemetry Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/change/`

**Files**:
- `adminapi/change/telemetry_profile_handler_test.go`
- `adminapi/change/telemetry_two_change_handler_test.go`
- `adminapi/change/telemetry_two_change_service_test.go`
- `adminapi/change/telemetry_two_profile_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 6: Setting Module Enhancement

### Task 6.1: Setting Module - Add test_utils.go and TestMain
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/setting/`

**Subtasks**:
- [ ] Create `test_utils.go`
- [ ] Add TestMain to one test file
- [ ] Verify mock mode works

**Files**:
- `adminapi/setting/test_utils.go` (NEW)
- `adminapi/setting/setting_profile_controller_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 6.2: Setting Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/setting/`

**Files**:
- `adminapi/setting/setting_profile_controller_test.go`
- `adminapi/setting/setting_profile_service_test.go`
- `adminapi/setting/setting_rule_controller_test.go`
- `adminapi/setting/setting_rule_service_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 7: Canary Module Enhancement

### Task 7.1: Canary Module - Add test_utils.go and TestMain
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/canary/`

**Files**:
- `adminapi/canary/test_utils.go` (NEW)
- `adminapi/canary/canary_settings_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 7.2: Canary Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/canary/`

**Files**:
- `adminapi/canary/canary_settings_handler_test.go`
- `adminapi/canary/canary_settings_service_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 8: RFC Module Enhancement

### Task 8.1: RFC Feature Module - Add test_utils.go
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/rfc/feature/`

**Files**:
- `adminapi/rfc/feature/test_utils.go` (NEW)
- `adminapi/rfc/feature/feature_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 8.2: RFC Feature Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/rfc/feature/`

**Files**:
- `adminapi/rfc/feature/feature_handler_test.go`
- `adminapi/rfc/feature/feature_control_settings_test.go`
- `adminapi/rfc/feature/feature_test_helpers_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 9: Auth Module Enhancement

### Task 9.1: Auth Module - Add test_utils.go and Fix TestMain
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/auth/`

**Description**: Auth tests panic without DB. Need mock support.

**Files**:
- `adminapi/auth/test_utils.go` (NEW)
- `adminapi/auth/idp_service_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 10: XCRP Module Enhancement

### Task 10.1: XCRP Module - Add test_utils.go and TestMain
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/xcrp/`

**Files**:
- `adminapi/xcrp/test_utils.go` (NEW)
- `adminapi/xcrp/recooking_lockdown_settings_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 10.2: XCRP Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/xcrp/`

**Files**:
- `adminapi/xcrp/recooking_lockdown_settings_handler_test.go`
- `adminapi/xcrp/recooking_status_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 11: Firmware Module Enhancement

### Task 11.1: Firmware Module - Add test_utils.go and TestMain
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/firmware/`

**Files**:
- `adminapi/firmware/test_utils.go` (NEW)
- `adminapi/firmware/firmware_test_page_controller_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 12: Configuration IP-MacRule Module Enhancement

### Task 12.1: IP-MacRule Module - Add test_utils.go and TestMain
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/configuration/ip-macrule/`

**Files**:
- `adminapi/configuration/ip-macrule/test_utils.go` (NEW)
- `adminapi/configuration/ip-macrule/ip_mac_ruleconfig_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 13: Tagging API Module Enhancement

### Task 13.1: Tagging Tag Module - Add test_utils.go and TestMain
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `taggingapi/tag/`

**Files**:
- `taggingapi/tag/test_utils.go` (NEW)
- `taggingapi/tag/tag_handler_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 13.2: Tag Tests - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `taggingapi/tag/`

**Files**:
- `taggingapi/tag/tag_handler_test.go`
- `taggingapi/tag/tag_member_service_test.go`
- `taggingapi/tag/tag_normalization_service_test.go`
- `taggingapi/tag/tag_service_test.go`
- `taggingapi/tag/tag_member_benchmark_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 13.3: Tagging Config Module - Add test_utils.go
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `taggingapi/config/`

**Files**:
- `taggingapi/config/test_utils.go` (NEW)
- `taggingapi/config/tag_config_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 13.4: Tagging Percentage Module - Add test_utils.go
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `taggingapi/percentage/`

**Files**:
- `taggingapi/percentage/test_utils.go` (NEW)
- `taggingapi/percentage/percentage_service_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 14: Shared Module Enhancement

### Task 14.1: Shared estbfirmware - Add test_utils.go
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `shared/estbfirmware/`

**Files**:
- `shared/estbfirmware/test_utils.go` (NEW)
- `shared/estbfirmware/estb_firmware_context_test.go`
- `shared/estbfirmware/config_change_logs_test.go`
- `shared/estbfirmware/singleton_filter_test.go`
- `shared/estbfirmware/time_filter_test.go`
- `shared/estbfirmware/estbfirmware_unit_test.go`
- `shared/estbfirmware/percent_filter_test.go`
- `shared/estbfirmware/estb_converters_test.go`
- `shared/estbfirmware/ip_filter_test.go`
- `shared/estbfirmware/reboot_immediately_filter_test.go`
- `shared/estbfirmware/firmware_config_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 14.2: Shared firmware - Add test_utils.go
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `shared/firmware/`

**Files**:
- `shared/firmware/test_utils.go` (NEW)
- `shared/firmware/firmwarerule_test.go`
- `shared/firmware/firmware_unit_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 14.3: Shared logupload - Add test_utils.go
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `shared/logupload/`

**Files**:
- `shared/logupload/test_utils.go` (NEW)
- `shared/logupload/utils_test.go`
- `shared/logupload/permanent_profile_test.go`
- `shared/logupload/logupload_test.go`
- `shared/logupload/telemetry_profile_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 14.4: Shared rfc - Add test_utils.go
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `shared/rfc/`

**Files**:
- `shared/rfc/test_utils.go` (NEW)
- `shared/rfc/feature_rule_test.go`
- `shared/rfc/feature_test.go`
- `shared/rfc/feature_predicate_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 14.5: Shared change - Add test_utils.go
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `shared/change/`

**Files**:
- `shared/change/test_utils.go` (NEW)
- `shared/change/change_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 14.6: Shared coretypes/percentage - Make Idempotent
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `shared/`

**Files**:
- `shared/coretypes_test.go`
- `shared/coretypes_clone_test.go`
- `shared/coretypes_additional_test.go`
- `shared/percentage_service_test.go`
- `shared/percentage_service_extra_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 15: HTTP Module Enhancement (If Needed)

### Task 15.1: HTTP Module - Verify Mock Support
**Status**: 🔲 Not Started  
**Priority**: P2 - Medium  
**Module**: `http/`

**Description**: HTTP module already has mock patterns. Verify and standardize if needed.

**Files**:
- `http/webconfig_server_test.go`
- `http/sat_validator_test.go`
- `http/xconf_connector_test.go`
- `http/groupsync_service_connector_test.go`
- `http/canarymgr_connector_test.go`
- `http/response_test.go`
- `http/xcrp_connector_test.go`
- `http/response_writer_test.go`
- `http/group_service_connector_test.go`
- `http/idp_service_connector_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 16: Final Validation

### Task 16.1: Full Suite Mock Mode Validation
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  

**Command**:
```bash
USE_MOCK_DB=true go test ./... -cover -count=1 -timeout=5m
```

**Expected**: All tests pass in < 30 seconds

---

### Task 16.2: Full Suite Real DB Validation
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  

**Command**:
```bash
USE_MOCK_DB=false go test ./... -cover -count=1 -timeout=45m
```

**Expected**: All tests pass with Cassandra running

---

### Task 16.3: Coverage Report Generation
**Status**: 🔲 Not Started  
**Priority**: P1 - High  

**Commands**:
```bash
# Mock mode coverage
USE_MOCK_DB=true go test ./... -coverprofile=coverage_mock.out -timeout=5m
go tool cover -html=coverage_mock.out -o coverage_mock.html

# Real mode coverage  
USE_MOCK_DB=false go test ./... -coverprofile=coverage_real.out -timeout=45m
go tool cover -html=coverage_real.out -o coverage_real.html
```

---

---

## Phase 17: Common Package Enhancement

### Task 17.1: Common Package - Add test_utils.go
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `common/`

**Description**: common/struct.go uses GetCachedSimpleDao extensively for:
- TABLE_APP_SETTINGS
- TABLE_DCM_RULE  
- TABLE_ENVIRONMENT
- TABLE_MODEL

**Files**:
- `common/test_utils.go` (NEW)
- `common/struct_test.go`

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Phase 18: Additional Infrastructure (GetSimpleDao & GetListingDao)

### Task 18.1: Mock GetSimpleDao Interface
**Status**: 🔲 Not Started  
**Priority**: P0 - Critical  
**Module**: `adminapi/dcm/mocks/`

**Description**: shared/change/change.go uses `db.GetSimpleDao()` (not GetCachedSimpleDao).
Need to add mock for SimpleDao interface.

**Tables Used**:
- TABLE_XCONF_CHANGE
- TABLE_XCONF_APPROVED_CHANGE
- TABLE_TELEMETRY_CHANGES
- TABLE_TELEMETRY_APPROVED_CHANGES
- TABLE_TELEMETRY_TWO_CHANGES
- TABLE_TELEMETRY_APPROVED_TWO_CHANGES

**Files**:
- `adminapi/dcm/mocks/mock_simple_dao.go` (NEW)
- `adminapi/dcm/mocks/mock_simple_dao_test.go` (NEW)

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

### Task 18.2: Mock GetListingDao Interface  
**Status**: 🔲 Not Started  
**Priority**: P1 - High  
**Module**: `adminapi/dcm/mocks/`

**Description**: shared/estbfirmware/config_change_logs.go uses `db.GetListingDao()`.
Need to add mock for ListingDao interface.

**Tables Used**:
- TABLE_LOGS (for config change logs)

**Files**:
- `adminapi/dcm/mocks/mock_listing_dao.go` (NEW)
- `adminapi/dcm/mocks/mock_listing_dao_test.go` (NEW)

**Coverage After**:
| Mode | Pass | Coverage |
|------|------|----------|
| Mock | 🔲   | --       |
| Real | 🔲   | --       |

---

## Appendix A: Packages WITHOUT Direct DB Access

These packages have tests but do NOT directly use database operations.
They may still need TestMain updates if they depend on other packages that use DB.

| Package | Files | Notes |
|---------|-------|-------|
| `adminapi/canary/` | 2 test files | No direct DB calls |
| `adminapi/auth/` | 1 test file | No direct DB calls, but creates WebconfigServer |
| `adminapi/lockdown/` | 2 test files | No direct DB calls |
| `taggingapi/config/` | 1 test file | No direct DB calls |
| `taggingapi/percentage/` | 1 test file | No direct DB calls |
| `util/` | 9 test files | Pure utility, no DB |

---

## Appendix B: Complete DB Usage Map

### Production Files Using DB

| File | DAO Type | Tables |
|------|----------|--------|
| `common/struct.go` | GetCachedSimpleDao | APP_SETTINGS, DCM_RULE, ENVIRONMENT, MODEL |
| `shared/change/change.go` | GetSimpleDao | XCONF_CHANGE, TELEMETRY_* |
| `shared/estbfirmware/firmware_config.go` | GetCachedSimpleDao | FIRMWARE_CONFIGS |
| `shared/estbfirmware/config_change_logs.go` | GetListingDao | LOGS |
| `shared/estbfirmware/*.go` (filters) | GetCachedSimpleDao | Various |
| `shared/firmware/firmwarerule.go` | GetCachedSimpleDao | FIRMWARE_RULES |
| `shared/logupload/logupload.go` | GetCachedSimpleDao | LOG_FILES, LOG_FILE_GROUPS |
| `shared/logupload/telemetry_profile.go` | GetCachedSimpleDao | TELEMETRY_* |
| `shared/rfc/feature.go` | GetCachedSimpleDao | XCONF_FEATURE |
| `taggingapi/tag/tag_member_service.go` | GetCachedSimpleDao | TAGS, TAG_MEMBERS |
| `adminapi/dcm/*.go` | GetCachedSimpleDao | DCM_RULE, DEVICE_SETTINGS, etc |
| `adminapi/queries/*.go` | GetCachedSimpleDao | Multiple tables |
| `adminapi/setting/*.go` | GetCachedSimpleDao | SETTING_PROFILE, SETTING_RULE |
| `adminapi/telemetry/*.go` | GetCachedSimpleDao | TELEMETRY_* |
| `adminapi/xcrp/*.go` | GetCachedSimpleDao | XCRP tables |

### Test Files Requiring Mock Support

**Total: 52 test files with direct DB imports**

```
adminapi/change/          - 4 files
adminapi/dcm/             - 14 files  
adminapi/queries/         - 24 files
adminapi/rfc/feature/     - 3 files
adminapi/setting/         - 4 files (indirect)
adminapi/telemetry/       - 9 files
shared/firmware/          - 1 file
shared/logupload/         - 1 file
```

---

## Summary Statistics

| Phase | Module | Tasks | Critical | High | Medium |
|-------|--------|-------|----------|------|--------|
| 1 | Infrastructure | 1 | 1 | 0 | 0 |
| 2 | DCM | 7 | 1 | 5 | 1 |
| 3 | Queries | 13 | 1 | 10 | 2 |
| 4 | Telemetry | 4 | 1 | 3 | 0 |
| 5 | Change | 3 | 1 | 2 | 0 |
| 6 | Setting | 2 | 1 | 1 | 0 |
| 7 | Canary | 2 | 1 | 1 | 0 |
| 8 | RFC | 2 | 1 | 1 | 0 |
| 9 | Auth | 1 | 1 | 0 | 0 |
| 10 | XCRP | 2 | 1 | 1 | 0 |
| 11 | Firmware | 1 | 1 | 0 | 0 |
| 12 | IP-MacRule | 1 | 1 | 0 | 0 |
| 13 | Tagging API | 4 | 3 | 1 | 0 |
| 14 | Shared | 6 | 0 | 6 | 0 |
| 15 | HTTP | 1 | 0 | 0 | 1 |
| 16 | Validation | 3 | 2 | 1 | 0 |
| 17 | Common | 1 | 0 | 1 | 0 |
| 18 | Additional DAO Mocks | 2 | 1 | 1 | 0 |
| **Total** | | **56** | **18** | **35** | **4** |
