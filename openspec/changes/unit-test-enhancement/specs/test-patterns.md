# Test Pattern Specification

## Overview

This specification defines the standardized patterns for unit tests in xconfadmin with mock/real DB support.

## 1. Test File Structure Spec

### 1.1 Package-Level Test Initialization

Every package with DB-dependent tests **MUST** have:

```go
// File: <package>/test_utils.go

package <package>

import (
    "testing"
    "github.com/rdkcentral/xconfadmin/adminapi/dcm/mocks"
    "github.com/rdkcentral/xconfwebconfig/db"
    xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

// Package-level variables for mock state
var (
    mockDaoInstance               *mocks.MockCachedSimpleDao
    useMockDatabase               = false
    originalGetCachedSimpleDaoFunc func() db.CachedSimpleDao
)

// InitMockDatabase - Call in TestMain when USE_MOCK_DB=true
func InitMockDatabase() *mocks.MockCachedSimpleDao {
    mockDaoInstance = mocks.NewMockCachedSimpleDao()
    useMockDatabase = true
    
    originalGetCachedSimpleDaoFunc = xwlogupload.GetCachedSimpleDaoFunc
    xwlogupload.GetCachedSimpleDaoFunc = func() db.CachedSimpleDao {
        return mockDaoInstance
    }
    
    return mockDaoInstance
}

// DisableMockDatabase - Call in TestMain cleanup
func DisableMockDatabase() {
    if originalGetCachedSimpleDaoFunc != nil {
        xwlogupload.GetCachedSimpleDaoFunc = originalGetCachedSimpleDaoFunc
    }
    useMockDatabase = false
    mockDaoInstance = nil
}

// IsMockDatabaseEnabled - Check current mode
func IsMockDatabaseEnabled() bool {
    return useMockDatabase
}

// ClearMockDatabase - Clear all mock data between tests
func ClearMockDatabase() {
    if useMockDatabase && mockDaoInstance != nil {
        mockDaoInstance.Clear()
    }
}

// GetMockDaoForTesting - Get mock instance for assertions
func GetMockDaoForTesting() *mocks.MockCachedSimpleDao {
    return mockDaoInstance
}

// SkipIfMockDatabase - Skip integration tests in mock mode
func SkipIfMockDatabase(t *testing.T) {
    if useMockDatabase {
        t.Skip("Skipping integration test in mock mode")
    }
}
```

### 1.2 TestMain Pattern

Every package with DB tests **MUST** have one file with `TestMain`:

```go
// File: <package>/<main_test_file>_test.go

func TestMain(m *testing.M) {
    // 1. Check environment for mode selection
    useMock := os.Getenv("USE_MOCK_DB")
    isMockMode := useMock == "true" || useMock == "1" || useMock == ""
    
    if isMockMode {
        fmt.Println("MODE: Mock Database (fast)")
        
        // 2. Initialize mock DB client FIRST (prevents distributed lock panics)
        mockDbClient := mocks.NewMockDatabaseClient()
        db.SetDatabaseClient(mockDbClient)
        
        // 3. Register required table configurations
        db.RegisterTableConfigSimple(db.TABLE_XXX, xxx.NewXxxInf)
        // ... register all tables used by this package
        
        // 4. Initialize mock DAO
        InitMockDatabase()
        defer DisableMockDatabase()
    } else {
        fmt.Println("MODE: Real Database (integration)")
    }
    
    // 5. Common initialization (config, server, router)
    // ... setup code that works for both modes
    
    // 6. Run all tests
    code := m.Run()
    
    // 7. Cleanup
    // ... teardown code
    
    os.Exit(code)
}
```

## 2. Test Function Pattern Spec

### 2.1 Basic Idempotent Test

```go
func TestEntityCreate(t *testing.T) {
    // 1. Generate unique ID for this test run
    testID := "test-" + uuid.New().String()
    
    // 2. Create test entity
    entity := &Entity{
        ID:   testID,
        Name: "Test Entity",
    }
    
    // 3. Cleanup ONLY this entity (surgical cleanup)
    defer func() {
        if IsMockDatabaseEnabled() {
            GetMockDaoForTesting().DeleteOne(
                db.GetDefaultTenantId(), 
                db.TABLE_ENTITY, 
                testID,
            )
        } else {
            db.GetCachedSimpleDao().DeleteOne(
                db.GetDefaultTenantId(), 
                db.TABLE_ENTITY, 
                testID,
            )
        }
    }()
    
    // 4. Execute operation
    err := CreateEntity(entity)
    
    // 5. Assert results
    assert.NilError(t, err)
    
    // 6. Verify by reading back
    retrieved := GetEntity(db.GetDefaultTenantId(), testID)
    assert.Assert(t, retrieved != nil)
    assert.Equal(t, entity.Name, retrieved.Name)
}
```

### 2.2 Test with Multiple Entities

```go
func TestBatchOperation(t *testing.T) {
    // Track all IDs for cleanup
    var insertedIDs []string
    
    // Cleanup all inserted entities
    defer func() {
        for _, id := range insertedIDs {
            deleteEntityById(id)
        }
    }()
    
    // Create multiple entities
    for i := 0; i < 3; i++ {
        id := fmt.Sprintf("batch-test-%d-%s", i, uuid.New().String())
        entity := &Entity{ID: id, Name: fmt.Sprintf("Entity %d", i)}
        
        err := CreateEntity(entity)
        assert.NilError(t, err)
        
        insertedIDs = append(insertedIDs, id)
    }
    
    // Test batch operation
    results := GetAllEntities()
    assert.Assert(t, len(results) >= 3)
}
```

### 2.3 Test Helper Pattern

```go
// Helper function that returns cleanup function
func createTestEntityWithCleanup(t *testing.T, name string) (*Entity, func()) {
    id := "helper-" + uuid.New().String()
    entity := &Entity{ID: id, Name: name}
    
    err := CreateEntity(entity)
    assert.NilError(t, err)
    
    cleanup := func() {
        deleteEntityById(id)
    }
    
    return entity, cleanup
}

// Usage in test
func TestSomething(t *testing.T) {
    entity, cleanup := createTestEntityWithCleanup(t, "Test")
    defer cleanup()
    
    // Test logic using entity
}
```

## 3. FORBIDDEN Patterns

### 3.1 ❌ NEVER: Double Cleanup

```go
// WRONG - Double cleanup
func TestBad(t *testing.T) {
    DeleteAllEntities()        // ❌ REMOVE THIS
    defer DeleteAllEntities()  // ❌ CHANGE TO SURGICAL
    
    // test code
}
```

### 3.2 ❌ NEVER: Complete Table Wipe

```go
// WRONG - Wipes ALL data
func TestBad(t *testing.T) {
    defer func() {
        for _, table := range db.GetAllTableInfo() {
            cassandra.DeleteAllXconfData(table)  // ❌ FORBIDDEN
        }
    }()
}
```

### 3.3 ❌ NEVER: Test Dependency

```go
// WRONG - Tests depend on each other
func TestCreate(t *testing.T) {
    CreateEntity(entity)  // Leaves in DB
}

func TestRead(t *testing.T) {
    GetEntity(id)  // ❌ Assumes TestCreate ran first
}
```

### 3.4 ❌ NEVER: Hard-coded IDs Without Cleanup

```go
// WRONG - Same ID on every run, no cleanup
func TestBad(t *testing.T) {
    entity := &Entity{ID: "fixed-id-123"}  // ❌ Collision risk
    CreateEntity(entity)
    // No cleanup!
}
```

### 3.5 ❌ NEVER: Calling Other Test Functions

```go
// WRONG - Test calls another test
func TestA(t *testing.T) {
    TestB(t)  // ❌ FORBIDDEN
}
```

## 4. Delete Helper Spec

### 4.1 Surgical Delete Helper

```go
// deleteEntityById - works with both mock and real DB
func deleteEntityById(id string) error {
    if IsMockDatabaseEnabled() {
        return GetMockDaoForTesting().DeleteOne(
            db.GetDefaultTenantId(),
            db.TABLE_ENTITY,
            id,
        )
    }
    return db.GetCachedSimpleDao().DeleteOne(
        db.GetDefaultTenantId(),
        db.TABLE_ENTITY,
        id,
    )
}
```

### 4.2 Multi-Table Cleanup Helper

```go
// cleanupTestData - for tests that insert into multiple tables
func cleanupTestData(entityID string, relatedIDs map[string][]string) {
    // Delete main entity
    deleteEntityById(entityID)
    
    // Delete related records
    for tableName, ids := range relatedIDs {
        for _, id := range ids {
            deleteFromTable(tableName, id)
        }
    }
}
```

## 5. Mock vs Real DB Behavior Spec

### 5.1 Operations That Should Work Identically

| Operation | Mock | Real | Notes |
|-----------|------|------|-------|
| GetOne | ✓ | ✓ | Returns nil if not found |
| SetOne | ✓ | ✓ | Creates or updates |
| DeleteOne | ✓ | ✓ | No error if not exists |
| GetAllAsList | ✓ | ✓ | Empty list if none |
| GetAllAsMap | ✓ | ✓ | Empty map if none |

### 5.2 Operations That May Differ

| Operation | Mock Behavior | Real Behavior | Handling |
|-----------|--------------|---------------|----------|
| Distributed Lock | No-op | Actually locks | Mock doesn't test lock contention |
| Transaction | Not supported | Depends on schema | Skip transactional tests in mock |
| Concurrent writes | In-memory sync | Cassandra consistency | Test with `-race` flag |

### 5.3 Integration-Only Tests

Tests that **MUST** use real DB:

```go
func TestDistributedLockBehavior(t *testing.T) {
    SkipIfMockDatabase(t)  // Only run with real DB
    
    // Test actual lock behavior
}

func TestCassandraConsistency(t *testing.T) {
    SkipIfMockDatabase(t)
    
    // Test consistency behavior
}
```

## 6. Coverage Verification Spec

### 6.1 Per-Function Verification

After each test function modification:

```bash
# Mock mode
USE_MOCK_DB=true go test -v -run TestFunctionName \
    -coverprofile=func_mock.out \
    ./path/to/package/...

# Check specific function coverage
go tool cover -func=func_mock.out | grep "function_under_test"

# Real mode (if DB available)
USE_MOCK_DB=false go test -v -run TestFunctionName \
    -coverprofile=func_real.out \
    ./path/to/package/...

go tool cover -func=func_real.out | grep "function_under_test"
```

### 6.2 Required Coverage Documentation

In tasks.md, after completing each task:

```markdown
**Coverage Results**:
| Mode | Pass | Coverage | Function Lines |
|------|------|----------|----------------|
| Mock | ✅   | 85.2%    | 42/50 covered  |
| Real | ✅   | 85.2%    | 42/50 covered  |
```

## 7. Naming Convention Spec

### 7.1 Test Function Names

```
Test<Entity><Operation>[_<Scenario>]

Examples:
- TestDeviceSettingCreate
- TestDeviceSettingCreate_EmptyName
- TestDeviceSettingCreate_DuplicateId
- TestDeviceSettingGetById
- TestDeviceSettingGetById_NotFound
- TestDeviceSettingUpdate
- TestDeviceSettingDelete
```

### 7.2 Test ID Prefixes

```
<module>-<operation>-<uuid>

Examples:
- dcm-create-a1b2c3d4-...
- queries-model-e5f6g7h8-...
- telemetry-rule-i9j0k1l2-...
```

## 8. Table Registration Spec

### 8.1 Required Table Registrations by Module

**DCM Module**:
```go
db.RegisterTableConfigSimple(db.TABLE_DCM_RULE, logupload.NewDCMGenericRuleInf)
db.RegisterTableConfigSimple(db.TABLE_LOG_FILE, logupload.NewLogFileInf)
db.RegisterTableConfigSimple(db.TABLE_LOG_FILE_LIST, logupload.NewLogFileListInf)
db.RegisterTableConfigSimple(db.TABLE_LOG_UPLOAD_SETTINGS, logupload.NewLogUploadSettingsInf)
db.RegisterTableConfigSimple(db.TABLE_DEVICE_SETTINGS, logupload.NewDeviceSettingsInf)
db.RegisterTableConfigSimple(db.TABLE_VOD_SETTINGS, logupload.NewVodSettingsInf)
db.RegisterTableConfigSimple(db.TABLE_UPLOAD_REPOSITORY, logupload.NewUploadRepositoryInf)
db.RegisterTableConfigSimple(db.TABLE_XCONF_CHANGE, db.NewChangedDataInf)
```

**Queries Module**:
```go
db.RegisterTableConfigSimple(db.TABLE_FIRMWARE_CONFIG, firmware.NewFirmwareConfigInf)
db.RegisterTableConfigSimple(db.TABLE_FIRMWARE_RULE, firmware.NewFirmwareRuleInf)
db.RegisterTableConfigSimple(db.TABLE_FIRMWARE_RULE_TEMPLATE, firmware.NewFirmwareRuleTemplateInf)
db.RegisterTableConfigSimple(db.TABLE_IP_ADDRESS_GROUP, core.NewIpAddressGroupInf)
db.RegisterTableConfigSimple(db.TABLE_MAC_LIST, core.NewGenericNamespacedListInf)
db.RegisterTableConfigSimple(db.TABLE_MODELS, core.NewModelInf)
db.RegisterTableConfigSimple(db.TABLE_ENVIRONMENTS, core.NewEnvironmentInf)
// ... additional tables
```

**Telemetry Module**:
```go
db.RegisterTableConfigSimple(db.TABLE_TELEMETRY_RULES, logupload.NewTelemetryRuleInf)
db.RegisterTableConfigSimple(db.TABLE_TELEMETRY_TWO_RULES, logupload.NewTelemetryTwoRuleInf)
db.RegisterTableConfigSimple(db.TABLE_TELEMETRY_TWO_PROFILES, logupload.NewTelemetryTwoProfileInf)
db.RegisterTableConfigSimple(db.TABLE_PERMANENT_TELEMETRY_PROFILE, logupload.NewPermanentTelemetryProfileInf)
```

## 9. Error Handling Spec

### 9.1 Test Setup Errors

```go
func TestSomething(t *testing.T) {
    entity, err := setupTestData()
    if err != nil {
        t.Fatalf("Setup failed: %v", err)  // Fatal stops test
    }
    defer cleanup()
    
    // Test code
}
```

### 9.2 Cleanup Errors

```go
func TestSomething(t *testing.T) {
    defer func() {
        if err := cleanup(); err != nil {
            t.Logf("Warning: cleanup failed: %v", err)  // Log but don't fail
        }
    }()
    
    // Test code
}
```

## 10. Concurrency Spec

### 10.1 Race Detection

Always run tests with race detection:

```bash
USE_MOCK_DB=true go test -race ./...
```

### 10.2 Parallel Tests

For truly independent tests:

```go
func TestA(t *testing.T) {
    t.Parallel()  // Can run concurrently with other parallel tests
    
    // Each parallel test MUST use unique IDs
    id := "parallel-a-" + uuid.New().String()
    // ...
}

func TestB(t *testing.T) {
    t.Parallel()
    
    id := "parallel-b-" + uuid.New().String()
    // ...
}
```

### 10.3 Non-Parallel Tests

Tests sharing state should NOT be parallel:

```go
func TestSequentialA(t *testing.T) {
    // No t.Parallel() - runs sequentially
    // Can use shared test fixtures
}
```
