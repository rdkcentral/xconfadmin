# Unit Test Enhancement Proposal - Mock/Real DB Support

## Summary

Enhance the xconfadmin unit test infrastructure to support **dual-mode testing** with both mock database and real Cassandra database, controlled via environment variable. This enables fast isolated unit tests (mock mode) and integration verification (real DB mode) while ensuring **idempotent** tests with proper cleanup strategies.

## Problem Statement

The current test infrastructure has several issues:

1. **Tests require Cassandra** - Tests fail with connection errors when DB is not available
2. **Inconsistent mock patterns** - Some modules have mock support, others don't
3. **Non-idempotent tests** - Some tests call other test functions or have shared state
4. **Improper cleanup** - Tests perform both `DeleteAllEntities()` and `defer DeleteAllEntities()` (redundant double cleanup)
5. **Complete table cleanup** - Some tests wipe entire tables instead of just inserted data
6. **No coverage verification** - No systematic coverage tracking per function

## Goals

1. **Dual-mode support**: Every unit test function supports both mock and real DB
2. **Command control**: `USE_MOCK_DB=true|false` environment variable controls mode
3. **Idempotent tests**: No test function calls or depends on other test functions
4. **Surgical cleanup**: Tests only remove data they inserted
5. **No production changes**: Only test files modified
6. **Coverage verification**: Run `make cover` after each test with both modes
7. **157 test files** across all modules to be enhanced

## Scope

### Modules Using Database (In-Scope)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    MODULES REQUIRING ENHANCEMENT                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │  adminapi/  │  │  adminapi/  │  │  adminapi/  │  │  adminapi/  │    │
│  │    dcm/     │  │  telemetry/ │  │  queries/   │  │   change/   │    │
│  │ (21 tests)  │  │ (15 tests)  │  │ (52 tests)  │  │ (12 tests)  │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │  adminapi/  │  │  adminapi/  │  │  adminapi/  │  │  adminapi/  │    │
│  │  setting/   │  │   canary/   │  │   auth/     │  │   xcrp/     │    │
│  │ (4 tests)   │  │ (2 tests)   │  │ (1 test)    │  │ (2 tests)   │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │  adminapi/  │  │  adminapi/  │  │ taggingapi/ │  │  shared/    │    │
│  │  firmware/  │  │    rfc/     │  │   (all)     │  │   (all)     │    │
│  │ (1 test)    │  │ (5 tests)   │  │ (10 tests)  │  │ (25 tests)  │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    adminapi/configuration/ip-macrule             │   │
│  │                           (1 test)                                │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Non-DB Modules (Out of Scope)

- `util/` - Pure utility functions, no DB access
- `common/` - Configuration and constants
- `http/` - HTTP connectors (already has mocks)

## Current Test Count by Module

| Module | Test Files | Tests (Approx) | Has Mock | Has TestMain |
|--------|-----------|----------------|----------|--------------|
| adminapi/dcm | 13 | ~80 | ✓ | ✓ |
| adminapi/queries | 48 | ~200 | ✓ | ✓ |
| adminapi/telemetry | 10 | ~60 | ✓ | ✓ |
| adminapi/change | 8 | ~40 | ✗ | ✓ |
| adminapi/setting | 4 | ~20 | ✗ | ✗ |
| adminapi/canary | 2 | ~10 | ✗ | ✗ |
| adminapi/rfc/feature | 3 | ~15 | ✗ | ✓ |
| adminapi/auth | 1 | ~5 | ✗ | ✗ |
| adminapi/xcrp | 2 | ~10 | ✗ | ✗ |
| adminapi/firmware | 1 | ~5 | ✗ | ✗ |
| adminapi/configuration/ip-macrule | 1 | ~5 | ✗ | ✗ |
| taggingapi/* | 8 | ~30 | ✗ | ✗ |
| shared/* | 25 | ~100 | ✗ | ✗ |
| http/* | 10 | ~20 | ✓ (mocks) | ✓ |

## Non-Goals

- Modifying production code
- Adding new features to the application
- Changing database schema
- Modifying external dependencies

## Technical Approach

### 1. Mock/Real DB Control Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     TEST EXECUTION FLOW                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Environment Check                            │
│                  USE_MOCK_DB=true|false                          │
└─────────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┴───────────────────┐
          ▼                                       ▼
┌─────────────────────┐                 ┌─────────────────────┐
│   Mock Mode         │                 │   Real DB Mode      │
│ USE_MOCK_DB=true    │                 │ USE_MOCK_DB=false   │
├─────────────────────┤                 ├─────────────────────┤
│ • In-memory DAO     │                 │ • Cassandra DAO     │
│ • Ultra fast (<1ms) │                 │ • Integration test  │
│ • No external deps  │                 │ • Full validation   │
│ • Isolated tests    │                 │ • Real transactions │
└─────────────────────┘                 └─────────────────────┘
          │                                       │
          └───────────────────┬───────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Same Test Functions                           │
│              (Both modes use identical test code)                │
└─────────────────────────────────────────────────────────────────┘
```

### 2. Commands to Run Tests

```bash
# Mock mode (fast, no DB required) - DEFAULT
USE_MOCK_DB=true go test ./... -cover

# Real DB mode (integration, requires Cassandra)
USE_MOCK_DB=false go test ./... -cover

# Coverage for specific function (mock)
USE_MOCK_DB=true go test -run TestFunctionName -cover -coverprofile=func_mock.out

# Coverage for specific function (real)
USE_MOCK_DB=false go test -run TestFunctionName -cover -coverprofile=func_real.out
```

### 3. Cleanup Strategy Changes

**CURRENT (Problematic)**:
```go
func TestSomething(t *testing.T) {
    DeleteAllEntities()        // ← Problem: Wipes ALL data
    defer DeleteAllEntities()  // ← Problem: Double cleanup, wipes ALL data
    
    // Create test data
    entity := createTestEntity()
    
    // Test logic...
}
```

**PROPOSED (Surgical)**:
```go
func TestSomething(t *testing.T) {
    // Track what we insert
    var insertedIDs []string
    
    // Create test data
    entity := createTestEntity("test-id-123")
    insertedIDs = append(insertedIDs, entity.ID)
    
    // Cleanup ONLY what we inserted
    defer func() {
        for _, id := range insertedIDs {
            deleteEntity(id)
        }
    }()
    
    // Test logic...
}
```

### 4. Test Independence Pattern

**CURRENT (Non-idempotent)**:
```go
func TestCreate(t *testing.T) {
    entity := createEntity()
    // leaves entity in DB
}

func TestRead(t *testing.T) {
    // Depends on TestCreate having run first!
    entity := getEntity(id)  // ← FAILS if TestCreate didn't run
}
```

**PROPOSED (Idempotent)**:
```go
func TestCreate(t *testing.T) {
    entity := createEntity()
    defer deleteEntity(entity.ID)
    
    // Verify creation
    assert.NotNil(t, entity)
}

func TestRead(t *testing.T) {
    // Setup its own data
    entity := createEntity()
    defer deleteEntity(entity.ID)
    
    // Test read logic
    retrieved := getEntity(entity.ID)
    assert.Equal(t, entity.ID, retrieved.ID)
}
```

## Database Tables Accessed by Module

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      TABLE ACCESS BY MODULE                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DCM Module:                                                             │
│  ├── TABLE_DCM_RULE              ├── TABLE_LOG_FILE                     │
│  ├── TABLE_LOG_FILE_LIST         ├── TABLE_LOG_UPLOAD_SETTINGS          │
│  ├── TABLE_DEVICE_SETTINGS       ├── TABLE_VOD_SETTINGS                 │
│  └── TABLE_UPLOAD_REPOSITORY     └── TABLE_XCONF_CHANGE                 │
│                                                                          │
│  Queries Module:                                                         │
│  ├── TABLE_FIRMWARE_CONFIG       ├── TABLE_FIRMWARE_RULE                │
│  ├── TABLE_FIRMWARE_RULE_TEMPLATE├── TABLE_IP_ADDRESS_GROUP             │
│  ├── TABLE_MAC_LIST              ├── TABLE_MODELS                       │
│  ├── TABLE_ENVIRONMENTS          ├── TABLE_NAMESPACED_LISTS             │
│  └── TABLE_PERCENT_FILTER_VALUE  └── TABLE_FEATURE                      │
│                                                                          │
│  Telemetry Module:                                                       │
│  ├── TABLE_TELEMETRY_RULES       ├── TABLE_TELEMETRY_TWO_RULES          │
│  ├── TABLE_TELEMETRY_TWO_PROFILES├── TABLE_PERMANENT_TELEMETRY_PROFILE  │
│  └── TABLE_TARGETING_RULE                                                │
│                                                                          │
│  RFC Module:                                                             │
│  ├── TABLE_FEATURE               ├── TABLE_FEATURE_RULE                 │
│  └── TABLE_FEATURE_CONTROL_SETTING                                       │
│                                                                          │
│  Change Module:                                                          │
│  ├── TABLE_XCONF_CHANGE          ├── TABLE_TELEMETRY_TWO_CHANGE         │
│  └── TABLE_APPROVED_CHANGES                                              │
│                                                                          │
│  Tagging API:                                                            │
│  ├── TABLE_FIRMWARE_RULE         ├── TABLE_TAGS                         │
│  └── TABLE_TAG_MEMBERS                                                   │
│                                                                          │
│  Setting Module:                                                         │
│  ├── TABLE_SETTING_PROFILE       ├── TABLE_SETTING_RULE                 │
│  └── TABLE_SETTING_TYPE                                                  │
│                                                                          │
│  Canary Module:                                                          │
│  └── TABLE_CANARY_SETTINGS                                               │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Issues Identified for Resolution

### Issue 1: Double Cleanup Pattern
**Files affected**: 15+ test files
**Pattern**:
```go
DeleteAllEntities()
defer DeleteAllEntities()  // Redundant
```
**Fix**: Remove initial `DeleteAllEntities()`, only use targeted cleanup

### Issue 2: Complete Table Wipe
**Files affected**: All modules with `DeleteAllEntities()`
**Pattern**:
```go
func DeleteAllEntities() {
    for _, table := range db.GetAllTableInfo() {
        cassandraClient.DeleteAllXconfData(table)  // Wipes EVERYTHING
    }
}
```
**Fix**: Replace with surgical delete of only inserted data

### Issue 3: Test Function Dependencies
**Files affected**: e2e test files
**Pattern**: `TestAllDeviceSettingsApis` - One giant function testing multiple operations
**Fix**: Split into independent test functions, each with own setup/teardown

### Issue 4: Missing Mock Support
**Modules needing mocks**:
- adminapi/setting
- adminapi/canary
- adminapi/auth
- adminapi/xcrp
- adminapi/firmware
- adminapi/configuration/ip-macrule
- taggingapi/*
- shared/*

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Mock DAO doesn't match real DAO behavior | Tests pass with mock but fail with real | Run both modes in CI; periodic real DB validation |
| Surgical cleanup misses data | Data pollution between tests | Generate unique IDs per test; run with `-race` flag |
| Test isolation breaks shared setup | Tests fail intermittently | Audit all test functions for shared state |
| Coverage differs between modes | False confidence | Track and compare coverage metrics |

## Success Criteria

1. **All tests pass** with `USE_MOCK_DB=true`
2. **All tests pass** with `USE_MOCK_DB=false` (when DB available)
3. **No test function calls another test function**
4. **Each test cleans up only its own data**
5. **Coverage documented** after each test function enhancement
6. **Mock mode completes in < 30 seconds** for full suite

## Open Questions

1. Should we add `SkipIfRealDatabase(t)` for tests that only make sense with mocks?
2. How to handle TestMain collision when multiple test files in same package?
3. Should we create a shared test infrastructure package?
4. How to verify surgical cleanup is complete without full table scan?

## Related Files

- Sample implementation: `sample/xconfadmin/adminapi/dcm/` (reference only, will be removed)
- Mock DAO: `adminapi/dcm/mocks/mock_dao.go`
- Mock DB Client: `adminapi/dcm/mocks/mock_database_client.go`
- Test utils pattern: `adminapi/dcm/test_utils.go`
