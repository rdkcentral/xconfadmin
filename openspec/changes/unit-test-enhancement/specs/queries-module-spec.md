# Queries Module Specification

## Module Overview

**Package**: `adminapi/queries/`  
**Priority**: P0 - Critical  
**Status**: Partial mocks, largest module requiring enhancement

The Queries module is the largest module in xconfadmin, handling firmware rules, models, environments, filters, namespaced lists, and various configuration entities.

---

## Architecture

```
adminapi/queries/
├── Core Handlers
│   ├── queries_handler.go           # Main query handlers
│   ├── queries_helper.go            # Query helper functions
│   ├── base_queries_controller.go   # Base controller logic
│   └── queries_test.go              # TestMain, shared setup
│
├── Model Management
│   ├── model_handler.go             # Model CRUD handlers
│   ├── model_handler_test.go        # Model handler tests
│   ├── model_service.go             # Model business logic
│   ├── model_service_test.go        # Model service tests
│   ├── model_test.go                # Model unit tests
│   └── model_query_update_delete_test.go
│
├── Environment Management
│   ├── environment_handler.go       # Environment CRUD
│   ├── environment_handler_test.go
│   ├── environment_service.go
│   └── environment_service_test.go
│
├── Firmware Configuration
│   ├── firmware_config_handler.go
│   ├── firmware_config_handler_test.go
│   ├── firmware_config_service.go
│   ├── firmware_config_service_test.go
│   └── firmware_config_test.go
│
├── Firmware Rules
│   ├── firmware_rule_handler.go
│   ├── firmware_rule_handler_test.go
│   ├── firmware_rule_service.go
│   ├── firmware_rule_service_test.go
│   ├── firmware_rule_test.go
│   ├── firmware_rule_template_handler.go
│   ├── firmware_rule_template_handler_test.go
│   ├── firmware_rule_template_handler_additional_test.go
│   └── firmware_rule_template_service_test.go
│
├── Feature Management
│   ├── feature_handler.go
│   ├── feature_entity_handler_test.go
│   ├── feature_entity_service_test.go
│   ├── feature_rule_handler_test.go
│   ├── feature_rule_service_test.go
│   └── feature_service_test.go
│
├── IP/MAC Address Management
│   ├── ip_address_group_handler.go
│   ├── ip_address_group_service_test.go
│   ├── ipaddressgroup_maclist_handlers_test.go
│   ├── mac_rule_bean_handler.go
│   ├── mac_rule_bean_handler_test.go
│   └── maclist_test.go
│
├── Filters
│   ├── ips_filter_handler.go
│   ├── ips_filter_service_test.go
│   ├── location_filter_handler.go
│   ├── location_filter_service_test.go
│   ├── percent_filter_handler.go
│   ├── percent_filter_service_test.go
│   ├── percentfilter_handler_test.go
│   ├── ri_filter_handler.go
│   ├── ri_filter_service_test.go
│   ├── time_filter_handler.go
│   └── time_filter_service_test.go
│
├── Namespaced Lists
│   ├── namespaced_list_handler.go
│   ├── namespaced_list_handler_test.go
│   ├── namespaced_list_service.go
│   └── namespaced_list_service_test.go
│
├── Percentage Beans
│   ├── percentage_bean_handler.go
│   ├── percentage_bean_service_test.go
│   └── percentagebean_handler_test.go
│
├── AMV (Activation Minimum Version)
│   ├── amv_handler.go
│   ├── amv_handler_test.go
│   ├── amv_service.go
│   ├── amv_service_test.go
│   ├── amv_test.go
│   └── activation_minimum_version_handler_test.go
│
└── Supporting Files
    ├── converter.go
    ├── converter_test.go
    ├── common.go
    ├── common_test.go
    ├── baserule_validator_test.go
    ├── log_controller_test.go
    ├── log_file_handler_test.go
    └── various additional tests...
```

---

## Database Tables Used

| Table Name | Operations | Entity Type |
|------------|------------|-------------|
| `TABLE_MODELS` | CRUD | `shared.Model` |
| `TABLE_ENVIRONMENTS` | CRUD | `shared.Environment` |
| `TABLE_FIRMWARE_CONFIGS` | CRUD | `firmware.FirmwareConfig` |
| `TABLE_FIRMWARE_RULES` | CRUD | `firmware.FirmwareRule` |
| `TABLE_FIRMWARE_RULE_TEMPLATES` | CRUD | `firmware.FirmwareRuleTemplate` |
| `TABLE_FEATURES` | CRUD | `rfc.Feature` |
| `TABLE_FEATURE_CONTROL_RULES` | CRUD | `rfc.FeatureRule` |
| `TABLE_IP_ADDRESS_GROUPS` | CRUD | `shared.IpAddressGroup` |
| `TABLE_MAC_LISTS` | CRUD | `shared.MacList` |
| `TABLE_NS_LISTS` | CRUD | `shared.NamespacedList` |
| `TABLE_SINGLETON_FILTER_VALUE` | CRUD | Filter values |
| `TABLE_PERCENT_FILTER` | CRUD | `estbfirmware.PercentFilter` |

---

## Use Cases

### UC-QUERIES-001: Model Management

**Description**: CRUD operations for device models.

**API Endpoints**:
- GET `/queries/models` - List all models
- GET `/queries/models/{id}` - Get model by ID
- POST `/queries/models` - Create model
- PUT `/queries/models` - Update model
- DELETE `/queries/models/{id}` - Delete model

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-QRY-001-01 | Create valid model | 201 Created | ⚠️ Needs refactor |
| TC-QRY-001-02 | Create duplicate model | 409 Conflict | ⚠️ Needs refactor |
| TC-QRY-001-03 | Get all models | 200 OK + list | ⚠️ Needs refactor |
| TC-QRY-001-04 | Get model by ID | 200 OK + entity | ⚠️ Needs refactor |
| TC-QRY-001-05 | Get non-existent model | 404 Not Found | 🔲 Not tested |
| TC-QRY-001-06 | Update model | 200 OK | ⚠️ Needs refactor |
| TC-QRY-001-07 | Delete model | 200 OK | ⚠️ Needs refactor |
| TC-QRY-001-08 | Delete model with references | 409 Conflict | 🔲 Not tested |

---

### UC-QUERIES-002: Environment Management

**Description**: CRUD operations for environments.

**API Endpoints**:
- GET `/queries/environments` - List all environments
- GET `/queries/environments/{id}` - Get by ID
- POST `/queries/environments` - Create
- PUT `/queries/environments` - Update
- DELETE `/queries/environments/{id}` - Delete

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-QRY-002-01 | Create environment | 201 Created | ⚠️ Needs refactor |
| TC-QRY-002-02 | Get all environments | 200 OK + list | ⚠️ Needs refactor |
| TC-QRY-002-03 | Get environment by ID | 200 OK + entity | ⚠️ Needs refactor |
| TC-QRY-002-04 | Update environment | 200 OK | ⚠️ Needs refactor |
| TC-QRY-002-05 | Delete environment | 200 OK | ⚠️ Needs refactor |
| TC-QRY-002-06 | Delete with references | 409 Conflict | 🔲 Not tested |

---

### UC-QUERIES-003: Firmware Configuration

**Description**: CRUD operations for firmware configurations.

**API Endpoints**:
- GET `/queries/firmwareConfigs` - List all
- GET `/queries/firmwareConfigs/{id}` - Get by ID
- POST `/queries/firmwareConfigs` - Create
- PUT `/queries/firmwareConfigs` - Update
- DELETE `/queries/firmwareConfigs/{id}` - Delete

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-QRY-003-01 | Create firmware config | 201 Created | ⚠️ Needs refactor |
| TC-QRY-003-02 | Create duplicate | 409 Conflict | 🔲 Not tested |
| TC-QRY-003-03 | Get all configs | 200 OK + list | ⚠️ Needs refactor |
| TC-QRY-003-04 | Get config by ID | 200 OK + entity | ⚠️ Needs refactor |
| TC-QRY-003-05 | Update config | 200 OK | ⚠️ Needs refactor |
| TC-QRY-003-06 | Delete config | 200 OK | ⚠️ Needs refactor |
| TC-QRY-003-07 | Delete with rules | 409 Conflict | 🔲 Not tested |

---

### UC-QUERIES-004: Firmware Rules

**Description**: CRUD operations for firmware rules and templates.

**API Endpoints**:
- GET `/queries/rules` - List all rules
- GET `/queries/rules/{id}` - Get rule by ID
- POST `/queries/rules` - Create rule
- PUT `/queries/rules` - Update rule
- DELETE `/queries/rules/{id}` - Delete rule
- GET `/queries/rules/templates` - List templates

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-QRY-004-01 | Create firmware rule | 201 Created | ⚠️ Needs refactor |
| TC-QRY-004-02 | Create rule with invalid model | 400 Bad Request | 🔲 Not tested |
| TC-QRY-004-03 | Get all rules | 200 OK + list | ⚠️ Needs refactor |
| TC-QRY-004-04 | Get rule by ID | 200 OK + entity | ⚠️ Needs refactor |
| TC-QRY-004-05 | Update rule | 200 OK | ⚠️ Needs refactor |
| TC-QRY-004-06 | Delete rule | 200 OK | ⚠️ Needs refactor |
| TC-QRY-004-07 | Create rule template | 201 Created | ⚠️ Needs refactor |
| TC-QRY-004-08 | Apply template | 200 OK | 🔲 Not tested |

---

### UC-QUERIES-005: Namespaced Lists

**Description**: CRUD operations for namespaced lists (IP lists, MAC lists, etc.).

**API Endpoints**:
- GET `/queries/namespacedLists` - List all
- GET `/queries/namespacedLists/{id}` - Get by ID
- POST `/queries/namespacedLists` - Create
- PUT `/queries/namespacedLists` - Update
- DELETE `/queries/namespacedLists/{id}` - Delete

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-QRY-005-01 | Create namespaced list | 201 Created | ⚠️ Needs refactor |
| TC-QRY-005-02 | Get all lists | 200 OK + list | ⚠️ Needs refactor |
| TC-QRY-005-03 | Get list by ID | 200 OK + entity | ⚠️ Needs refactor |
| TC-QRY-005-04 | Update list | 200 OK | ⚠️ Needs refactor |
| TC-QRY-005-05 | Delete list | 200 OK | ⚠️ Needs refactor |
| TC-QRY-005-06 | Delete with references | 409 Conflict | 🔲 Not tested |

---

### UC-QUERIES-006: Filters

**Description**: Manage various filter types (IP, location, time, percent, RI).

**API Endpoints**:
- GET/POST/PUT/DELETE `/queries/filters/ips`
- GET/POST/PUT/DELETE `/queries/filters/location`
- GET/POST/PUT/DELETE `/queries/filters/time`
- GET/POST/PUT/DELETE `/queries/filters/percent`
- GET/POST/PUT/DELETE `/queries/filters/ri`

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-QRY-006-01 | Create IP filter | 201 Created | ⚠️ Needs refactor |
| TC-QRY-006-02 | Create location filter | 201 Created | ⚠️ Needs refactor |
| TC-QRY-006-03 | Create time filter | 201 Created | ⚠️ Needs refactor |
| TC-QRY-006-04 | Create percent filter | 201 Created | ⚠️ Needs refactor |
| TC-QRY-006-05 | Create RI filter | 201 Created | ⚠️ Needs refactor |

---

### UC-QUERIES-007: Percentage Beans

**Description**: Manage percentage distribution configurations.

**API Endpoints**:
- GET `/queries/percentageBeans` - List all
- POST `/queries/percentageBeans` - Create
- PUT `/queries/percentageBeans` - Update
- DELETE `/queries/percentageBeans/{id}` - Delete

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-QRY-007-01 | Create percentage bean | 201 Created | ⚠️ Needs refactor |
| TC-QRY-007-02 | Get all beans | 200 OK + list | ⚠️ Needs refactor |
| TC-QRY-007-03 | Update bean | 200 OK | ⚠️ Needs refactor |
| TC-QRY-007-04 | Delete bean | 200 OK | ⚠️ Needs refactor |
| TC-QRY-007-05 | Validate percentages sum | 400 if >100% | 🔲 Not tested |

---

### UC-QUERIES-008: AMV (Activation Minimum Version)

**Description**: Manage activation minimum version rules.

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-QRY-008-01 | Create AMV rule | 201 Created | ⚠️ Needs refactor |
| TC-QRY-008-02 | Get all AMV rules | 200 OK + list | ⚠️ Needs refactor |
| TC-QRY-008-03 | Update AMV rule | 200 OK | ⚠️ Needs refactor |
| TC-QRY-008-04 | Delete AMV rule | 200 OK | ⚠️ Needs refactor |

---

## Current Issues

### Issue QUERIES-001: TestMain Shared State

**File**: `queries_test.go`

**Problem**: TestMain creates global server instance used by all tests.

**Solution**:
```go
func TestMain(m *testing.M) {
    if IsMockDatabaseEnabled() {
        InitMockDatabase()  // Initialize mock before server
    }
    
    // Create server (uses mock if enabled)
    server = oshttp.NewWebconfigServer(...)
    
    code := m.Run()
    os.Exit(code)
}
```

---

### Issue QUERIES-002: Large Test Files

**Problem**: Some test files have 50+ test functions with shared setup.

**Solution**: 
- Group related tests in subtests
- Use table-driven tests
- Each test creates and cleans up own data

---

## Test Data Fixtures

```go
// Model fixture
func NewTestModel(id string) *shared.Model {
    return &shared.Model{
        ID:          id,
        Description: "Test Model " + id[:8],
    }
}

// Environment fixture
func NewTestEnvironment(id string) *shared.Environment {
    return &shared.Environment{
        ID:          id,
        Description: "Test Environment " + id[:8],
    }
}

// Firmware Config fixture
func NewTestFirmwareConfig(id string) *firmware.FirmwareConfig {
    return &firmware.FirmwareConfig{
        ID:               id,
        Description:      "Test Config " + id[:8],
        FirmwareFilename: "test.bin",
        FirmwareVersion:  "1.0.0",
    }
}

// Firmware Rule fixture
func NewTestFirmwareRule(id, modelId string) *firmware.FirmwareRule {
    return &firmware.FirmwareRule{
        ID:              id,
        Name:            "Test Rule " + id[:8],
        Type:            "MODEL_RULE",
        ApplicationType: "stb",
        Rule: shared.Rule{
            Condition: shared.Condition{
                FreeArg: shared.FreeArg{
                    Type: "STRING",
                    Name: "model",
                },
                Operation: "IS",
                FixedArg: shared.FixedArg{
                    Bean: shared.Bean{
                        Value: shared.Value{
                            Java_class: "java.lang.String",
                            Value:      modelId,
                        },
                    },
                },
            },
        },
    }
}

// Namespaced List fixture
func NewTestNamespacedList(id string) *shared.NamespacedList {
    return &shared.NamespacedList{
        ID:   id,
        Type: "MAC_LIST",
        Data: []string{"AA:BB:CC:DD:EE:FF"},
    }
}
```

---

## Coverage Goals

| Area | Files | Current | Target |
|------|-------|---------|--------|
| Model | 3 | ~60% | 85% |
| Environment | 2 | ~55% | 85% |
| Firmware Config | 3 | ~50% | 85% |
| Firmware Rules | 6 | ~45% | 80% |
| Features | 5 | ~55% | 85% |
| Filters | 6 | ~50% | 80% |
| Namespaced Lists | 2 | ~60% | 85% |
| Percentage Beans | 2 | ~55% | 85% |
| AMV | 4 | ~50% | 80% |

---

## Test Execution Commands

```bash
# Run all queries tests with mock
USE_MOCK_DB=true go test -v ./adminapi/queries/... -count=1 -timeout=5m

# Run specific test file
USE_MOCK_DB=true go test -v ./adminapi/queries/... -run "Model" -count=1

# Run with coverage
USE_MOCK_DB=true go test ./adminapi/queries/... -coverprofile=queries.out
go tool cover -func=queries.out | head -50

# Generate HTML coverage report
go tool cover -html=queries.out -o queries_coverage.html
```
