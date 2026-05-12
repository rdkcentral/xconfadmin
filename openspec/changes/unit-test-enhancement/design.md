# Unit Test Enhancement - Technical Design

## Overview

This design specifies the technical implementation for dual-mode (mock/real DB) unit testing infrastructure across all xconfadmin modules.

## Architecture

### 1. Test Mode Control Flow

```
┌──────────────────────────────────────────────────────────────────────────┐
│                        TEST INITIALIZATION FLOW                          │
└──────────────────────────────────────────────────────────────────────────┘

     ┌─────────────────┐
     │   go test       │
     │   invoked       │
     └────────┬────────┘
              │
              ▼
     ┌─────────────────┐
     │   TestMain()    │──────────────────────────────────────────┐
     │   Entry Point   │                                          │
     └────────┬────────┘                                          │
              │                                                   │
              ▼                                                   │
     ┌─────────────────────────────┐                              │
     │ os.Getenv("USE_MOCK_DB")    │                              │
     └─────────────┬───────────────┘                              │
                   │                                              │
     ┌─────────────┴─────────────┐                                │
     │                           │                                │
     ▼                           ▼                                │
┌─────────────┐          ┌─────────────┐                          │
│ true/1      │          │ false/0     │                          │
│ Mock Mode   │          │ Real Mode   │                          │
└──────┬──────┘          └──────┬──────┘                          │
       │                        │                                 │
       ▼                        ▼                                 │
┌──────────────┐         ┌──────────────┐                         │
│ InitMock     │         │ Connect to   │                         │
│ Database()   │         │ Cassandra    │                         │
└──────┬───────┘         └──────┬───────┘                         │
       │                        │                                 │
       └───────────┬────────────┘                                 │
                   │                                              │
                   ▼                                              │
          ┌────────────────┐                                      │
          │ Override       │                                      │
          │ GetCachedSimple│                                      │
          │ DaoFunc        │                                      │
          └────────┬───────┘                                      │
                   │                                              │
                   ▼                                              │
          ┌────────────────┐                                      │
          │  m.Run()       │◄─────────────────────────────────────┘
          │  Execute Tests │
          └────────────────┘
```

### 2. Mock DAO Interface

The mock DAO must implement the `db.CachedSimpleDao` interface:

```go
type CachedSimpleDao interface {
    GetOne(tenantId, tableName, rowKey string) (interface{}, error)
    GetOneFromCacheOnly(tenantId, tableName, rowKey string) (interface{}, error)
    SetOne(tenantId, tableName, rowKey string, entity interface{}) error
    DeleteOne(tenantId, tableName, rowKey string) error
    GetAllByKeys(tenantId, tableName string, rowKeys []string) ([]interface{}, error)
    GetAllAsList(tenantId, tableName string, maxResults int) ([]interface{}, error)
    GetAllAsMap(tenantId, tableName string) (map[interface{}]interface{}, error)
    GetAllAsShallowMap(tenantId, tableName string) (map[interface{}]interface{}, error)
    GetKeys(tenantId, tableName string) ([]interface{}, error)
    RefreshAll(tenantId, tableName string) error
    RefreshOne(tenantId, tableName, rowKey string) error
}
```

### 3. Test Utility Structure per Module

```
adminapi/
└── <module>/
    ├── test_utils.go           ← Mock/Real DB switching logic
    ├── mocks/                   ← Mock implementations (if needed locally)
    │   ├── mock_dao.go
    │   ├── mock_dao_test.go
    │   ├── mock_database_client.go
    │   └── test_setup.go
    └── *_test.go               ← Test files with idempotent tests
```

### 4. Shared Mock Package Location

```
adminapi/
└── dcm/
    └── mocks/                   ← SHARED mocks used by all modules
        ├── mock_dao.go          ← MockCachedSimpleDao implementation
        ├── mock_database_client.go  ← MockDatabaseClient for locks
        ├── mock_distributed_lock.go ← MockDistributedLock
        └── test_setup.go        ← Global setup helpers
```

## Component Designs

### 4.1 test_utils.go Template

Each module requiring DB access needs this pattern:

```go
package <module>

import (
    "testing"
    "github.com/rdkcentral/xconfadmin/adminapi/dcm/mocks"
    "github.com/rdkcentral/xconfwebconfig/db"
    xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

var mockDaoInstance *mocks.MockCachedSimpleDao
var useMockDatabase = false
var originalGetCachedSimpleDaoFunc func() db.CachedSimpleDao

// InitMockDatabase initializes mock mode
func InitMockDatabase() *mocks.MockCachedSimpleDao {
    mockDaoInstance = mocks.NewMockCachedSimpleDao()
    useMockDatabase = true
    
    // CRITICAL: Override global DAO function
    originalGetCachedSimpleDaoFunc = xwlogupload.GetCachedSimpleDaoFunc
    xwlogupload.GetCachedSimpleDaoFunc = func() db.CachedSimpleDao {
        return mockDaoInstance
    }
    
    return mockDaoInstance
}

// DisableMockDatabase restores real mode
func DisableMockDatabase() {
    if originalGetCachedSimpleDaoFunc != nil {
        xwlogupload.GetCachedSimpleDaoFunc = originalGetCachedSimpleDaoFunc
    }
    useMockDatabase = false
    mockDaoInstance = nil
}

// IsMockDatabaseEnabled returns current mode
func IsMockDatabaseEnabled() bool {
    return useMockDatabase
}

// ClearMockDatabase clears all mock data
func ClearMockDatabase() {
    if useMockDatabase && mockDaoInstance != nil {
        mockDaoInstance.Clear()
    }
}

// SkipIfMockDatabase skips integration tests in mock mode
func SkipIfMockDatabase(t *testing.T) {
    if useMockDatabase {
        t.Skip("Skipping integration test in mock mode")
    }
}

// GetMockDaoForTesting returns mock for assertions
func GetMockDaoForTesting() *mocks.MockCachedSimpleDao {
    return mockDaoInstance
}
```

### 4.2 TestMain Template

```go
func TestMain(m *testing.M) {
    // Check mode from environment
    useMock := os.Getenv("USE_MOCK_DB")
    if useMock == "true" || useMock == "1" || useMock == "" {
        // Default to mock mode for speed
        fmt.Println("Using MOCK database")
        
        // Initialize mock DB client (prevents distributed lock panics)
        mockDbClient := mocks.NewMockDatabaseClient()
        db.SetDatabaseClient(mockDbClient)
        
        // Register table configurations
        registerTableConfigs()
        
        // Initialize mock DAO
        InitMockDatabase()
        defer DisableMockDatabase()
    } else {
        fmt.Println("Using REAL database")
        // Real DB initialization follows...
    }
    
    // Common setup (config, router, etc.)
    setupTestEnvironment()
    
    // Run tests
    code := m.Run()
    
    // Cleanup
    teardownTestEnvironment()
    
    os.Exit(code)
}
```

### 4.3 Idempotent Test Pattern

```go
// CORRECT: Idempotent test with surgical cleanup
func TestCreateDeviceSetting(t *testing.T) {
    // Generate unique ID for this test
    testID := "test-device-" + uuid.New().String()
    
    // Setup: Create test data
    setting := &logupload.DeviceSettings{
        ID:              testID,
        Name:            "Test Device Setting",
        ApplicationType: "stb",
    }
    
    // Cleanup: Remove only what we created
    defer func() {
        if IsMockDatabaseEnabled() {
            GetMockDaoForTesting().DeleteOne(db.GetDefaultTenantId(), 
                db.TABLE_DEVICE_SETTINGS, testID)
        } else {
            db.GetCachedSimpleDao().DeleteOne(db.GetDefaultTenantId(), 
                db.TABLE_DEVICE_SETTINGS, testID)
        }
    }()
    
    // Execute
    err := CreateDeviceSetting(setting)
    
    // Assert
    assert.NilError(t, err)
    
    // Verify by reading back
    retrieved := GetDeviceSetting(db.GetDefaultTenantId(), testID)
    assert.Equal(t, setting.Name, retrieved.Name)
}
```

### 4.4 Test Helper Functions

```go
// Helper to create entity with cleanup tracking
func createTestEntity(t *testing.T, entity interface{}, id string, tableName string) func() {
    // Insert
    err := setOneInDao(tableName, id, entity)
    assert.NilError(t, err)
    
    // Return cleanup function
    return func() {
        deleteOneFromDao(tableName, id)
    }
}

// Usage:
func TestSomething(t *testing.T) {
    entity := &MyEntity{ID: "test-123"}
    cleanup := createTestEntity(t, entity, entity.ID, db.TABLE_MY_ENTITY)
    defer cleanup()
    
    // Test logic...
}
```

## Coverage Tracking Design

### Coverage Verification Loop

```
┌──────────────────────────────────────────────────────────────────────┐
│                    COVERAGE VERIFICATION WORKFLOW                     │
└──────────────────────────────────────────────────────────────────────┘

For each test function enhancement:

┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ 1. Modify   │───▶│ 2. Run Mock │───▶│ 3. Run Real │───▶│ 4. Document │
│    Test     │    │    Tests    │    │    Tests    │    │   Coverage  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                         │                  │                   │
                         ▼                  ▼                   ▼
                   ┌───────────┐      ┌───────────┐      ┌───────────┐
                   │ Verify    │      │ Verify    │      │ Record in │
                   │ Pass + %  │      │ Pass + %  │      │ tasks.md  │
                   └───────────┘      └───────────┘      └───────────┘
```

### Coverage Commands

```bash
# Single function coverage (mock)
USE_MOCK_DB=true go test -run TestFunctionName \
    -coverprofile=coverage_mock.out \
    -coverpkg=./...

# Single function coverage (real)
USE_MOCK_DB=false go test -run TestFunctionName \
    -coverprofile=coverage_real.out \
    -coverpkg=./...

# View coverage
go tool cover -func=coverage_mock.out | grep FunctionName

# HTML report
go tool cover -html=coverage_mock.out -o coverage.html
```

### Coverage Tracking Format

In `tasks.md`, each task records:

```markdown
### Task X: TestFunctionName enhancement

**Status**: ✅ Completed

**Coverage Results**:
| Mode | Pass | Coverage | Function Coverage |
|------|------|----------|-------------------|
| Mock | ✅   | 85.2%    | 100% (X/X lines)  |
| Real | ✅   | 85.2%    | 100% (X/X lines)  |

**Changes Made**:
- Added surgical cleanup
- Made idempotent
- Added mock support
```

## Module-Specific Designs

### DCM Module (Already Has Mocks)

**Current State**: Has mock infrastructure, needs cleanup improvements
**Changes Required**:
- Fix double cleanup pattern
- Make e2e tests idempotent
- Add coverage tracking

### Queries Module (Partially Has Mocks)

**Current State**: Has test_utils.go, needs standardization
**Changes Required**:
- Ensure all tests use mock/real pattern
- Fix surgical cleanup
- Remove test dependencies

### Telemetry Module (Has Mocks)

**Current State**: Has mock support
**Changes Required**:
- Standardize with other modules
- Fix cleanup patterns
- Add coverage tracking

### Setting Module (No Mocks)

**Current State**: No mock infrastructure
**Changes Required**:
- Add test_utils.go
- Add TestMain with mode switching
- Convert all tests to idempotent pattern

### Canary Module (No Mocks)

**Current State**: No mock infrastructure
**Changes Required**:
- Add test_utils.go
- Add TestMain with mode switching
- Convert tests to idempotent pattern

### RFC Module (No Mocks)

**Current State**: Has TestMain but no mock switching
**Changes Required**:
- Add mock infrastructure to test_utils.go
- Update TestMain for mode switching
- Convert tests to idempotent pattern

### Tagging API Module (No Mocks)

**Current State**: No mock infrastructure
**Changes Required**:
- Add test_utils.go to taggingapi/tag/
- Add test_utils.go to taggingapi/config/
- Add test_utils.go to taggingapi/percentage/
- Add TestMain with mode switching
- Convert all tests to idempotent pattern

### Shared Module (No Mocks)

**Current State**: Tests may use real DB directly
**Changes Required**:
- Audit each submodule (estbfirmware, firmware, logupload, rfc, change)
- Add mock support where DB is used
- Convert tests to idempotent pattern

### Auth Module (No Mocks)

**Current State**: Panics without DB connection
**Changes Required**:
- Add test_utils.go
- Add mock for IDP service
- Make WebconfigServer initialization conditional

### XCRP Module (No Mocks)

**Current State**: No mock infrastructure
**Changes Required**:
- Add test_utils.go
- Add TestMain with mode switching
- Convert tests to idempotent pattern

### Configuration IP-MacRule Module (No Mocks)

**Current State**: No mock infrastructure
**Changes Required**:
- Add test_utils.go
- Add TestMain with mode switching
- Convert tests to idempotent pattern

## File Changes Summary

### New Files to Create

| File | Purpose |
|------|---------|
| `adminapi/setting/test_utils.go` | Mock/real switching for setting module |
| `adminapi/canary/test_utils.go` | Mock/real switching for canary module |
| `adminapi/auth/test_utils.go` | Mock/real switching for auth module |
| `adminapi/xcrp/test_utils.go` | Mock/real switching for xcrp module |
| `adminapi/firmware/test_utils.go` | Mock/real switching for firmware module |
| `adminapi/configuration/ip-macrule/test_utils.go` | Mock/real switching |
| `taggingapi/tag/test_utils.go` | Mock/real switching for tag module |
| `taggingapi/config/test_utils.go` | Mock/real switching for config module |
| `taggingapi/percentage/test_utils.go` | Mock/real switching for percentage module |
| `shared/estbfirmware/test_utils.go` | Mock/real switching |
| `shared/firmware/test_utils.go` | Mock/real switching |
| `shared/logupload/test_utils.go` | Mock/real switching |
| `shared/rfc/test_utils.go` | Mock/real switching |
| `shared/change/test_utils.go` | Mock/real switching |

### Files to Modify

| File | Changes |
|------|---------|
| All `*_test.go` files | Idempotent tests, surgical cleanup |
| Existing `test_utils.go` files | Standardize mock infrastructure |
| Files with `TestMain` | Add mode switching |
| Files with `DeleteAllEntities` | Replace with surgical cleanup |

## Testing Strategy

### Unit Test Execution Order

1. **Mock Mode First**: Always run mock mode first (fast feedback)
2. **Real Mode Second**: Run real mode for integration validation
3. **Compare Coverage**: Ensure both modes achieve same coverage

### CI/CD Integration

```yaml
# GitHub Actions example
jobs:
  test-mock:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Mock Tests
        run: USE_MOCK_DB=true make cover
        
  test-real:
    runs-on: ubuntu-latest
    services:
      cassandra:
        image: cassandra:4.1
    steps:
      - uses: actions/checkout@v3
      - name: Run Real Tests
        run: USE_MOCK_DB=false make cover
```

## Acceptance Criteria

1. **Mode Switching**: `USE_MOCK_DB=true|false` controls database mode
2. **All Tests Pass**: Both modes pass all tests
3. **Idempotent**: No test depends on another test's execution
4. **Surgical Cleanup**: Tests only delete data they create
5. **Coverage Parity**: Mock and real modes achieve same coverage %
6. **Performance**: Mock mode completes in < 30 seconds
7. **No Production Changes**: Zero modifications to non-test files
