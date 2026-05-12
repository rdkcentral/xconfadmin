# Unit Test Enhancement - Module Inventory

## Overview

Complete inventory of all modules and files requiring enhancement for mock/real DB support.

---

## Modules Summary

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         XCONFADMIN MODULE INVENTORY                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ADMINAPI MODULES (Primary Focus)                                            │
│  ═══════════════════════════════════════════════════════════════════════     │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ adminapi/dcm/           │ 21 files │ Has mocks │ Has TestMain │ P0  │    │
│  │ adminapi/queries/       │ 52 files │ Partial   │ Has TestMain │ P0  │    │
│  │ adminapi/telemetry/     │ 15 files │ Has mocks │ Has TestMain │ P0  │    │
│  │ adminapi/change/        │ 12 files │ No mocks  │ Has TestMain │ P1  │    │
│  │ adminapi/setting/       │  4 files │ No mocks  │ No TestMain  │ P1  │    │
│  │ adminapi/canary/        │  2 files │ No mocks  │ No TestMain  │ P1  │    │
│  │ adminapi/rfc/feature/   │  5 files │ No mocks  │ Has TestMain │ P1  │    │
│  │ adminapi/auth/          │  1 file  │ No mocks  │ No TestMain  │ P1  │    │
│  │ adminapi/xcrp/          │  2 files │ No mocks  │ No TestMain  │ P2  │    │
│  │ adminapi/firmware/      │  1 file  │ No mocks  │ No TestMain  │ P2  │    │
│  │ adminapi/configuration/ │  1 file  │ No mocks  │ No TestMain  │ P2  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  TAGGING API MODULES                                                         │
│  ═══════════════════════════════════════════════════════════════════════     │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ taggingapi/tag/         │  6 files │ No mocks  │ No TestMain  │ P1  │    │
│  │ taggingapi/config/      │  1 file  │ No mocks  │ No TestMain  │ P2  │    │
│  │ taggingapi/percentage/  │  1 file  │ No mocks  │ No TestMain  │ P2  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  SHARED MODULES                                                              │
│  ═══════════════════════════════════════════════════════════════════════     │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ shared/estbfirmware/    │ 11 files │ No mocks  │ No TestMain  │ P1  │    │
│  │ shared/firmware/        │  2 files │ No mocks  │ No TestMain  │ P2  │    │
│  │ shared/logupload/       │  4 files │ No mocks  │ No TestMain  │ P2  │    │
│  │ shared/rfc/             │  3 files │ No mocks  │ No TestMain  │ P2  │    │
│  │ shared/change/          │  1 file  │ No mocks  │ No TestMain  │ P2  │    │
│  │ shared/ (root)          │  5 files │ No mocks  │ No TestMain  │ P2  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  HTTP MODULE (Already Has Mocks)                                             │
│  ═══════════════════════════════════════════════════════════════════════     │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ http/                   │ 10 files │ Has mocks │ Has TestMain │ P3  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  UTILITY MODULES (No DB Access - Out of Scope)                               │
│  ═══════════════════════════════════════════════════════════════════════     │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ util/                   │  9 files │ N/A       │ N/A          │ --  │    │
│  │ common/                 │  4 files │ N/A       │ N/A          │ --  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Detailed File Inventory

### adminapi/dcm/ (21 test files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `dcmformula_test.go` | Unit/E2E | ✓ | ⚠️ Partial | ⚠️ Double |
| `device_settings_e2e_test.go` | E2E | ✓ | ❌ Single function | ⚠️ DeleteAll |
| `device_settings_handler_test.go` | Handler | ✓ | ⚠️ Partial | ⚠️ Partial |
| `vod_settings_e2e_test.go` | E2E | ✓ | ❌ Single function | ⚠️ DeleteAll |
| `vod_settings_handler_test.go` | Handler | ✓ | ⚠️ Partial | ⚠️ Partial |
| `logrepo_settings_e2e_test.go` | E2E | ✓ | ⚠️ Partial | ⚠️ Double |
| `logrepo_settings_handler_test.go` | Handler | ✓ | ⚠️ Partial | ⚠️ Partial |
| `logrepo_settings_service_test.go` | Service | ✓ | ⚠️ Partial | ⚠️ Double |
| `logupload_settings_e2e_test.go` | E2E | ✓ | ⚠️ Partial | ⚠️ DeleteAll |
| `logupload_settings_handler_test.go` | Handler | ✓ | ⚠️ Partial | ⚠️ Partial |
| `test_page_controller_test.go` | Controller | ✓ | ⚠️ Partial | ⚠️ Partial |
| `test_utils.go` | Utils | ✓ | N/A | N/A |
| `mocks/mock_dao.go` | Mock | ✓ | N/A | N/A |
| `mocks/mock_dao_test.go` | Test | ✓ | ✓ | ✓ |
| `mocks/mock_database_client.go` | Mock | ✓ | N/A | N/A |
| `mocks/mock_database_client_test.go` | Test | ✓ | ✓ | ✓ |
| `mocks/mock_distributed_lock.go` | Mock | ✓ | N/A | N/A |
| `mocks/mock_distributed_lock_test.go` | Test | ✓ | ✓ | ✓ |
| `mocks/test_setup.go` | Utils | ✓ | N/A | N/A |

### adminapi/queries/ (52 test files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `model_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `model_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `model_test.go` | Unit | ⚠️ | ⚠️ | ⚠️ |
| `model_query_update_delete_test.go` | Mixed | ⚠️ | ❌ | ⚠️ |
| `firmware_config_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `firmware_config_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `firmware_config_test.go` | Unit | ⚠️ | ⚠️ | ⚠️ |
| `firmware_rule_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `firmware_rule_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `firmware_rule_test.go` | Unit | ⚠️ | ⚠️ | ⚠️ |
| `firmware_rule_template_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `firmware_rule_template_handler_additional_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `firmware_rule_template_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `firmware_rule_report_page_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `feature_entity_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `feature_entity_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `feature_rule_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `feature_rule_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `feature_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `ip_address_group_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `ipaddressgroup_maclist_handlers_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `mac_rule_bean_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `maclist_test.go` | Unit | ⚠️ | ⚠️ | ⚠️ |
| `ips_filter_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `location_filter_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `percent_filter_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `percentfilter_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `ri_filter_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `time_filter_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `namespaced_list_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `namespaced_list_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `environment_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `environment_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `percentage_bean_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `percentagebean_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `amv_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `amv_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `amv_test.go` | Unit | ⚠️ | ⚠️ | ⚠️ |
| `activation_minimum_version_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `log_controller_test.go` | Controller | ⚠️ | ⚠️ | ⚠️ |
| `log_file_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `base_queries_controller_test.go` | Controller | ⚠️ | ⚠️ | ⚠️ |
| `baserule_validator_test.go` | Validator | ⚠️ | ⚠️ | ⚠️ |
| `common_test.go` | Common | ⚠️ | ⚠️ | ⚠️ |
| `converter_test.go` | Converter | ⚠️ | ⚠️ | ⚠️ |
| `queries_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `queries_helper_test.go` | Helper | ⚠️ | ⚠️ | ⚠️ |
| `queries_test.go` | Unit | ⚠️ | ⚠️ | ⚠️ |
| `additional_handler_test.go` | Handler | ⚠️ | ⚠️ | ⚠️ |
| `additional_service_test.go` | Service | ⚠️ | ⚠️ | ⚠️ |
| `coverage_improvement_test.go` | Coverage | ⚠️ | ⚠️ | ⚠️ |
| `test_utils.go` | Utils | ✓ | N/A | N/A |

### adminapi/telemetry/ (15 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `telemetry_profile_controller_test.go` | Controller | ✓ | ⚠️ | ⚠️ |
| `telemetry_profile_handler_test.go` | Handler | ✓ | ⚠️ | ⚠️ |
| `telemetry_profile_service_test.go` | Service | ✓ | ⚠️ | ⚠️ |
| `telemetry_rule_handler_test.go` | Handler | ✓ | ⚠️ | ⚠️ |
| `telemetry_v2_rule_service_test.go` | Service | ✓ | ⚠️ | ⚠️ |
| `telemetry_two_dao_test.go` | DAO | ✓ | ⚠️ | ⚠️ |
| `telemetry_two_loguploader_handler_test.go` | Handler | ✓ | ⚠️ | ⚠️ |
| `telemetry_two_profile_handler_test.go` | Handler | ✓ | ⚠️ | ⚠️ |
| `telemetry_two_rule_hanlder_test.go` | Handler | ✓ | ⚠️ | ⚠️ |
| `test_utils.go` | Utils | ✓ | N/A | N/A |

### adminapi/change/ (12 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `change_handler_test.go` | Handler | ❌ | ⚠️ | ⚠️ |
| `change_service_test.go` | Service | ❌ | ⚠️ | ⚠️ |
| `telemetry_profile_handler_test.go` | Handler | ❌ | ⚠️ | ⚠️ |
| `telemetry_two_change_handler_test.go` | Handler | ❌ | ⚠️ | ⚠️ |
| `telemetry_two_change_service_test.go` | Service | ❌ | ⚠️ | ⚠️ |
| `telemetry_two_profile_handler_test.go` | Handler | ❌ | ⚠️ | ⚠️ |

### adminapi/setting/ (4 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `setting_profile_controller_test.go` | Controller | ❌ | ⚠️ | ⚠️ |
| `setting_profile_service_test.go` | Service | ❌ | ⚠️ | ⚠️ |
| `setting_rule_controller_test.go` | Controller | ❌ | ⚠️ | ⚠️ |
| `setting_rule_service_test.go` | Service | ❌ | ⚠️ | ⚠️ |

### adminapi/canary/ (2 files) - NO DIRECT DB

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `canary_settings_handler_test.go` | Handler | N/A | ⚠️ | ⚠️ |
| `canary_settings_service_test.go` | Service | N/A | ⚠️ | ⚠️ |

**Note**: No direct DB calls. May need TestMain if depends on other DB-using packages.

### adminapi/rfc/feature/ (5 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `feature_handler_test.go` | Handler | ❌ | ⚠️ | ⚠️ |
| `feature_control_settings_test.go` | Settings | ❌ | ⚠️ | ⚠️ |
| `feature_test_helpers_test.go` | Helper | ❌ | ⚠️ | ⚠️ |

### adminapi/auth/ (1 file)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `idp_service_handler_test.go` | Handler | ❌ | ⚠️ | ⚠️ |

### adminapi/xcrp/ (2 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `recooking_lockdown_settings_handler_test.go` | Handler | ❌ | ⚠️ | ⚠️ |
| `recooking_status_handler_test.go` | Handler | ❌ | ⚠️ | ⚠️ |

### adminapi/firmware/ (1 file)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `firmware_test_page_controller_test.go` | Controller | ❌ | ⚠️ | ⚠️ |

### adminapi/configuration/ip-macrule/ (1 file)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `ip_mac_ruleconfig_handler_test.go` | Handler | ❌ | ⚠️ | ⚠️ |

### taggingapi/tag/ (6 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `tag_handler_test.go` | Handler | ❌ | ⚠️ | ⚠️ |
| `tag_member_service_test.go` | Service | ❌ | ⚠️ | ⚠️ |
| `tag_normalization_service_test.go` | Service | ❌ | ⚠️ | ⚠️ |
| `tag_service_test.go` | Service | ❌ | ⚠️ | ⚠️ |
| `tag_member_benchmark_test.go` | Benchmark | ❌ | ⚠️ | ⚠️ |

### taggingapi/config/ (1 file)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `tag_config_test.go` | Config | ❌ | ⚠️ | ⚠️ |

### taggingapi/percentage/ (1 file)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `percentage_service_test.go` | Service | ❌ | ⚠️ | ⚠️ |

### shared/estbfirmware/ (11 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `estb_firmware_context_test.go` | Context | ❌ | ⚠️ | ⚠️ |
| `config_change_logs_test.go` | Logs | ❌ | ⚠️ | ⚠️ |
| `singleton_filter_test.go` | Filter | ❌ | ⚠️ | ⚠️ |
| `time_filter_test.go` | Filter | ❌ | ⚠️ | ⚠️ |
| `estbfirmware_unit_test.go` | Unit | ❌ | ⚠️ | ⚠️ |
| `percent_filter_test.go` | Filter | ❌ | ⚠️ | ⚠️ |
| `estb_converters_test.go` | Converter | ❌ | ⚠️ | ⚠️ |
| `ip_filter_test.go` | Filter | ❌ | ⚠️ | ⚠️ |
| `reboot_immediately_filter_test.go` | Filter | ❌ | ⚠️ | ⚠️ |
| `firmware_config_test.go` | Config | ❌ | ⚠️ | ⚠️ |

### shared/firmware/ (2 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `firmwarerule_test.go` | Rule | ❌ | ⚠️ | ⚠️ |
| `firmware_unit_test.go` | Unit | ❌ | ⚠️ | ⚠️ |

### shared/logupload/ (4 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `utils_test.go` | Utils | ❌ | ⚠️ | ⚠️ |
| `permanent_profile_test.go` | Profile | ❌ | ⚠️ | ⚠️ |
| `logupload_test.go` | Unit | ❌ | ⚠️ | ⚠️ |
| `telemetry_profile_test.go` | Profile | ❌ | ⚠️ | ⚠️ |

### shared/rfc/ (3 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `feature_rule_test.go` | Rule | ❌ | ⚠️ | ⚠️ |
| `feature_test.go` | Feature | ❌ | ⚠️ | ⚠️ |
| `feature_predicate_test.go` | Predicate | ❌ | ⚠️ | ⚠️ |

### shared/change/ (1 file)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `change_test.go` | Change | ❌ | ⚠️ | ⚠️ |

### shared/ (root - 5 files)

| File | Type | Has Mock | Idempotent | Cleanup |
|------|------|----------|------------|---------|
| `coretypes_test.go` | CoreTypes | ❌ | ⚠️ | ⚠️ |
| `coretypes_clone_test.go` | Clone | ❌ | ⚠️ | ⚠️ |
| `coretypes_additional_test.go` | Additional | ❌ | ⚠️ | ⚠️ |
| `percentage_service_test.go` | Service | ❌ | ⚠️ | ⚠️ |
| `percentage_service_extra_test.go` | Service | ❌ | ⚠️ | ⚠️ |

---

## Legend

| Symbol | Meaning |
|--------|---------|
| ✓ | Complete/Present |
| ❌ | Missing/Not Present |
| ⚠️ | Partial/Needs Review |
| N/A | Not Applicable |

---

## Statistics

| Category | Count |
|----------|-------|
| Total Test Files | 157 |
| Files with Mock Support | ~25 |
| Files Needing Mock Infrastructure | ~132 |
| New test_utils.go Files Needed | ~15 |
| Files with Double Cleanup Issue | ~20+ |
| Files with DeleteAllEntities | ~30+ |
