# Mock Infrastructure Specification

## Overview

This document specifies the mock database infrastructure required for dual-mode testing (mock/real DB) across all xconfadmin modules.

---

## Architecture

```
Mock Infrastructure
├── Environment Control
│   └── USE_MOCK_DB environment variable
│
├── Mock DAO Layer
│   ├── MockCachedSimpleDao (exists)
│   ├── MockSimpleDao (NEW - for GetSimpleDao)
│   └── MockListingDao (NEW - for GetListingDao)
│
├── Mock Database Client
│   └── MockDatabaseClient (exists)
│
├── Mock Distributed Lock
│   └── MockDistributedLock (exists)
│
└── Test Utilities
    ├── InitMockDatabase()
    ├── ClearMockDatabase()
    ├── IsMockDatabaseEnabled()
    └── createTestEntity() / cleanup functions
```

---

## DAO Interface Requirements

### 1. CachedSimpleDao Interface (Existing)

**Location**: `adminapi/dcm/mocks/mock_dao.go`

**Used By**: 39 files across all modules

**Required Methods**:

| Method | Signature | Description |
|--------|-----------|-------------|
| GetOne | `(tableName, rowKey string) (interface{}, error)` | Get single entity |
| GetOneFromCacheOnly | `(tableName, rowKey string) (interface{}, error)` | Get from cache only |
| SetOne | `(tableName, rowKey string, entity interface{}) error` | Create/Update entity |
| DeleteOne | `(tableName, rowKey string) error` | Delete entity |
| GetAllAsList | `(tableName string, maxResults int) ([]interface{}, error)` | List all entities |
| GetAllAsMap | `(tableName string) (map[interface{}]interface{}, error)` | Get as map |
| GetAllAsShallowMap | `(tableName string) (map[interface{}]interface{}, error)` | Get shallow map |
| GetAllByKeys | `(tableName string, rowKeys []string) ([]interface{}, error)` | Get by multiple keys |
| RefreshAll | `(tableName string) error` | Refresh cache |
| GetKeys | `(tableName string) ([]string, error)` | Get all keys |

**Implementation Pattern**:
```go
type MockCachedSimpleDao struct {
    data map[string]map[string]interface{} // tableName -> rowKey -> entity
    mu   sync.RWMutex
}

func (m *MockCachedSimpleDao) GetOne(tableName, rowKey string) (interface{}, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if m.data[tableName] == nil {
        return nil, errors.New("not found")
    }
    entity, ok := m.data[tableName][rowKey]
    if !ok {
        return nil, errors.New("not found")
    }
    return entity, nil
}
```

---

### 2. SimpleDao Interface (NEW - Needed)

**Location**: `adminapi/dcm/mocks/mock_simple_dao.go` (to create)

**Used By**: `shared/change/change.go`

**Tables**:
- TABLE_XCONF_CHANGE
- TABLE_XCONF_APPROVED_CHANGE
- TABLE_TELEMETRY_CHANGES
- TABLE_TELEMETRY_APPROVED_CHANGES
- TABLE_TELEMETRY_TWO_CHANGES
- TABLE_TELEMETRY_APPROVED_TWO_CHANGES

**Required Methods**:

| Method | Signature | Description |
|--------|-----------|-------------|
| GetOne | `(tableName, rowKey string) (interface{}, error)` | Get entity |
| SetOne | `(tableName, rowKey string, entity interface{}) error` | Set entity |
| DeleteOne | `(tableName, rowKey string) error` | Delete entity |
| GetAllAsList | `(tableName string, maxResults int) ([]interface{}, error)` | List all |
| GetRange | `(tableName, startKey, endKey string, maxResults int) ([]interface{}, error)` | Range query |

**Implementation Pattern**:
```go
type MockSimpleDao struct {
    data map[string]map[string]interface{}
    mu   sync.RWMutex
}

func NewMockSimpleDao() *MockSimpleDao {
    return &MockSimpleDao{
        data: make(map[string]map[string]interface{}),
    }
}

func (m *MockSimpleDao) GetRange(tableName, startKey, endKey string, maxResults int) ([]interface{}, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    var result []interface{}
    if m.data[tableName] == nil {
        return result, nil
    }
    
    count := 0
    for key, entity := range m.data[tableName] {
        if key >= startKey && key <= endKey {
            result = append(result, entity)
            count++
            if maxResults > 0 && count >= maxResults {
                break
            }
        }
    }
    return result, nil
}
```

---

### 3. ListingDao Interface (NEW - Needed)

**Location**: `adminapi/dcm/mocks/mock_listing_dao.go` (to create)

**Used By**: `shared/estbfirmware/config_change_logs.go`

**Tables**:
- TABLE_LOGS

**Required Methods**:

| Method | Signature | Description |
|--------|-----------|-------------|
| SetOne | `(tableName, rowKey string, entity interface{}) error` | Store log entry |
| GetRange | `(tableName, startKey, endKey string, maxResults int) ([]interface{}, error)` | Get log range |
| DeleteRange | `(tableName, startKey, endKey string) error` | Delete log range |

**Implementation Pattern**:
```go
type MockListingDao struct {
    data map[string][]interface{} // tableName -> sorted entries
    mu   sync.RWMutex
}

func NewMockListingDao() *MockListingDao {
    return &MockListingDao{
        data: make(map[string][]interface{}),
    }
}
```

---

## Environment Control

### USE_MOCK_DB Variable

```go
// Check if mock mode is enabled
func IsMockDatabaseEnabled() bool {
    return os.Getenv("USE_MOCK_DB") == "true"
}

// Skip test if mock mode (for integration tests)
func SkipIfMockDatabase(t *testing.T) {
    if IsMockDatabaseEnabled() {
        t.Skip("Skipping: requires real database (USE_MOCK_DB=true)")
    }
}

// Skip test if real DB (for mock-only tests)
func SkipIfRealDatabase(t *testing.T) {
    if !IsMockDatabaseEnabled() {
        t.Skip("Skipping: mock-only test (USE_MOCK_DB=false)")
    }
}
```

---

## Mock Initialization

### TestMain Pattern

```go
var (
    mockDao      *mocks.MockCachedSimpleDao
    mockSimpleDao *mocks.MockSimpleDao
    mockListingDao *mocks.MockListingDao
)

func TestMain(m *testing.M) {
    // Load config
    testConfigFile = GetTestConfig()
    sc, _ = xwcommon.NewServerConfig(testConfigFile)
    
    // Initialize mock BEFORE server creation
    if IsMockDatabaseEnabled() {
        InitMockDatabase()
    }
    
    // Create server (will use mock if enabled)
    server = oshttp.NewWebconfigServer(sc, false)
    router = server.GetRouter()
    
    // Run tests
    code := m.Run()
    
    os.Exit(code)
}

func InitMockDatabase() {
    // Create mock instances
    mockDao = mocks.NewMockCachedSimpleDao()
    mockSimpleDao = mocks.NewMockSimpleDao()
    mockListingDao = mocks.NewMockListingDao()
    
    // Register mock DAOs
    db.SetMockCachedSimpleDao(mockDao)
    db.SetMockSimpleDao(mockSimpleDao)
    db.SetMockListingDao(mockListingDao)
    
    // Register table configs
    registerTableConfigs()
}

func registerTableConfigs() {
    // DCM tables
    db.RegisterTableConfigSimple(db.TABLE_DEVICE_SETTINGS, logupload.NewDeviceSettingsInf)
    db.RegisterTableConfigSimple(db.TABLE_VOD_SETTINGS, logupload.NewVodSettingsInf)
    db.RegisterTableConfigSimple(db.TABLE_LOG_UPLOAD_SETTINGS, logupload.NewLogUploadSettingsInf)
    db.RegisterTableConfigSimple(db.TABLE_UPLOAD_REPOSITORY, logupload.NewUploadRepositoryInf)
    
    // Queries tables
    db.RegisterTableConfigSimple(db.TABLE_MODELS, shared.NewModelInf)
    db.RegisterTableConfigSimple(db.TABLE_ENVIRONMENTS, shared.NewEnvironmentInf)
    db.RegisterTableConfigSimple(db.TABLE_FIRMWARE_CONFIGS, firmware.NewFirmwareConfigInf)
    db.RegisterTableConfigSimple(db.TABLE_FIRMWARE_RULES, firmware.NewFirmwareRuleInf)
    
    // ... more tables
}

func ClearMockDatabase() {
    if mockDao != nil {
        mockDao.Clear()
    }
    if mockSimpleDao != nil {
        mockSimpleDao.Clear()
    }
    if mockListingDao != nil {
        mockListingDao.Clear()
    }
}
```

---

## Test Utility Functions

### Entity Creation and Cleanup

```go
// CreateTestEntity creates an entity and returns cleanup function
func createTestEntity(t *testing.T, entity interface{}, id, tableName string) func() {
    t.Helper()
    
    err := db.GetCachedSimpleDao().SetOne(tableName, id, entity)
    if err != nil {
        t.Fatalf("Failed to create test entity: %v", err)
    }
    
    // Return cleanup function
    return func() {
        _ = db.GetCachedSimpleDao().DeleteOne(tableName, id)
    }
}

// CreateMultipleTestEntities creates multiple entities
func createMultipleTestEntities(t *testing.T, entities map[string]interface{}, tableName string) func() {
    t.Helper()
    
    for id, entity := range entities {
        err := db.GetCachedSimpleDao().SetOne(tableName, id, entity)
        if err != nil {
            t.Fatalf("Failed to create test entity %s: %v", id, err)
        }
    }
    
    return func() {
        for id := range entities {
            _ = db.GetCachedSimpleDao().DeleteOne(tableName, id)
        }
    }
}
```

### HTTP Test Helpers

```go
// ExecuteRequest executes HTTP request and returns response
func ExecuteRequest(r *http.Request, handler http.Handler) *httptest.ResponseRecorder {
    recorder := httptest.NewRecorder()
    handler.ServeHTTP(recorder, r)
    return recorder
}

// PerformRequest helper for common test patterns
func PerformRequest(t *testing.T, method, url string, body []byte, expectedStatus int) *httptest.ResponseRecorder {
    t.Helper()
    
    var req *http.Request
    var err error
    
    if body != nil {
        req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
    } else {
        req, err = http.NewRequest(method, url, nil)
    }
    
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.AddCookie(&http.Cookie{Name: "applicationType", Value: "stb"})
    
    res := ExecuteRequest(req, router)
    
    if res.Code != expectedStatus {
        t.Errorf("Expected status %d, got %d. Body: %s", expectedStatus, res.Code, res.Body.String())
    }
    
    return res
}
```

---

## Table Configuration Map

### Tables by Module

| Module | Tables | DAO Type |
|--------|--------|----------|
| DCM | DEVICE_SETTINGS, VOD_SETTINGS, LOG_UPLOAD_SETTINGS, UPLOAD_REPOSITORY, DCM_RULES | CachedSimpleDao |
| Queries | MODELS, ENVIRONMENTS, FIRMWARE_CONFIGS, FIRMWARE_RULES, NS_LISTS, IP_ADDRESS_GROUPS | CachedSimpleDao |
| Telemetry | TELEMETRY_PROFILES, TELEMETRY_RULES, TELEMETRY_TWO_PROFILES, TELEMETRY_TWO_RULES | CachedSimpleDao |
| Change | XCONF_CHANGE, XCONF_APPROVED_CHANGE, TELEMETRY_CHANGES, TELEMETRY_APPROVED_CHANGES | SimpleDao |
| Setting | SETTING_PROFILES, SETTING_RULES | CachedSimpleDao |
| RFC | FEATURES, FEATURE_CONTROL_RULES | CachedSimpleDao |
| Shared/estbfirmware | LOGS | ListingDao |
| Common | APP_SETTINGS, DCM_RULES, ENVIRONMENTS, MODELS | CachedSimpleDao |

---

## Thread Safety

All mock implementations MUST be thread-safe:

```go
type MockCachedSimpleDao struct {
    data map[string]map[string]interface{}
    mu   sync.RWMutex  // Read/Write mutex for thread safety
}

// Read operations use RLock
func (m *MockCachedSimpleDao) GetOne(tableName, rowKey string) (interface{}, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    // ...
}

// Write operations use Lock
func (m *MockCachedSimpleDao) SetOne(tableName, rowKey string, entity interface{}) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    // ...
}
```

---

## Performance Characteristics

| Operation | Mock Time | Real DB Time | Speedup |
|-----------|-----------|--------------|---------|
| GetOne | ~1µs | ~10ms | 10,000x |
| SetOne | ~1µs | ~15ms | 15,000x |
| DeleteOne | ~1µs | ~10ms | 10,000x |
| GetAllAsList (100 items) | ~10µs | ~50ms | 5,000x |
| Full test suite | ~30s | ~15min | 30x |

---

## Files to Create

| File | Purpose | Priority |
|------|---------|----------|
| `adminapi/dcm/mocks/mock_simple_dao.go` | SimpleDao mock implementation | P0 |
| `adminapi/dcm/mocks/mock_simple_dao_test.go` | SimpleDao mock tests | P0 |
| `adminapi/dcm/mocks/mock_listing_dao.go` | ListingDao mock implementation | P1 |
| `adminapi/dcm/mocks/mock_listing_dao_test.go` | ListingDao mock tests | P1 |

---

## Verification Commands

```bash
# Test mock infrastructure
USE_MOCK_DB=true go test -v ./adminapi/dcm/mocks/... -count=1

# Verify mock works across modules
USE_MOCK_DB=true go test -v ./adminapi/dcm/... -run "Mock" -count=1

# Full suite with mock
USE_MOCK_DB=true go test ./... -count=1 -timeout=5m
```
