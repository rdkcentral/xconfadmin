# DCM Module Specification

## Module Overview

**Package**: `adminapi/dcm/`  
**Priority**: P0 - Critical  
**Status**: Has mocks, needs idempotent refactoring

The DCM (Device Configuration Management) module handles device settings, log upload configurations, VOD settings, and upload repository management.

---

## Architecture

```
adminapi/dcm/
├── Production Code
│   ├── device_settings_handler.go      # HTTP handlers for device settings
│   ├── device_settings_service.go      # Business logic for device settings
│   ├── vod_settings_handler.go         # HTTP handlers for VOD settings
│   ├── vod_settings_service.go         # Business logic for VOD settings
│   ├── logupload_settings_handler.go   # HTTP handlers for log upload
│   ├── logupload_settings_service.go   # Business logic for log upload
│   ├── logrepo_settings_handler.go     # HTTP handlers for log repositories
│   ├── logrepo_settings_service.go     # Business logic for log repositories
│   ├── dcmformula_handler.go           # DCM formula management
│   ├── dcmformula_service.go           # DCM formula business logic
│   └── test_page_controller.go         # Test page controller
│
├── Test Code
│   ├── dcmformula_test.go              # TestMain, integration tests
│   ├── device_settings_e2e_test.go     # E2E device settings tests
│   ├── device_settings_handler_test.go # Handler unit tests
│   ├── vod_settings_e2e_test.go        # E2E VOD settings tests
│   ├── vod_settings_handler_test.go    # Handler unit tests
│   ├── logupload_settings_e2e_test.go  # E2E log upload tests
│   ├── logupload_settings_handler_test.go # Handler unit tests
│   ├── logrepo_settings_e2e_test.go    # E2E log repo tests
│   ├── logrepo_settings_handler_test.go # Handler unit tests
│   ├── logrepo_settings_service_test.go # Service unit tests
│   ├── test_page_controller_test.go    # Test page tests
│   └── test_utils.go                   # Test utilities
│
└── Mocks
    ├── mock_dao.go                     # CachedSimpleDao mock
    ├── mock_dao_test.go                # Mock DAO tests
    ├── mock_database_client.go         # DatabaseClient mock
    ├── mock_database_client_test.go    # Mock client tests
    ├── mock_distributed_lock.go        # Distributed lock mock
    ├── mock_distributed_lock_test.go   # Mock lock tests
    └── test_setup.go                   # Test setup utilities
```

---

## Database Tables Used

| Table Name | Operations | Entity Type |
|------------|------------|-------------|
| `TABLE_DEVICE_SETTINGS` | CRUD | `logupload.DeviceSettings` |
| `TABLE_VOD_SETTINGS` | CRUD | `logupload.VodSettings` |
| `TABLE_LOG_UPLOAD_SETTINGS` | CRUD | `logupload.LogUploadSettings` |
| `TABLE_UPLOAD_REPOSITORY` | CRUD | `logupload.UploadRepository` |
| `TABLE_DCM_RULES` | CRUD | `logupload.DCMGenericRule` |

---

## Use Cases

### UC-DCM-001: Device Settings Management

**Description**: CRUD operations for device settings configuration.

**Actors**: Admin User

**Preconditions**:
- User is authenticated
- User has appropriate permissions
- Application type cookie is set

**Main Flow**:
1. **Create**: POST `/xconfAdminService/dcm/deviceSettings`
   - Validate device settings JSON
   - Check for duplicate ID
   - Store in `TABLE_DEVICE_SETTINGS`
   - Return 201 Created

2. **Read All**: GET `/xconfAdminService/dcm/deviceSettings`
   - Filter by application type
   - Return all matching settings

3. **Read One**: GET `/xconfAdminService/dcm/deviceSettings/{id}`
   - Lookup by ID
   - Return 404 if not found

4. **Update**: PUT `/xconfAdminService/dcm/deviceSettings`
   - Validate entity exists
   - Validate changes
   - Update in database
   - Return 200 OK

5. **Delete**: DELETE `/xconfAdminService/dcm/deviceSettings/{id}`
   - Check entity exists
   - Check no dependencies
   - Delete from database
   - Return 200 OK

**Alternative Flows**:
- A1: Duplicate ID → Return 409 Conflict
- A2: Entity not found → Return 404 Not Found
- A3: Invalid JSON → Return 400 Bad Request
- A4: Missing auth cookie → Return 401 Unauthorized

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-DCM-001-01 | Create valid device setting | 201 Created | ⚠️ In monolithic test |
| TC-DCM-001-02 | Create duplicate device setting | 409 Conflict | ⚠️ In monolithic test |
| TC-DCM-001-03 | Get all device settings | 200 OK + list | ⚠️ In monolithic test |
| TC-DCM-001-04 | Get device setting by ID | 200 OK + entity | ⚠️ In monolithic test |
| TC-DCM-001-05 | Get non-existent setting | 404 Not Found | 🔲 Not tested |
| TC-DCM-001-06 | Update valid device setting | 200 OK | ⚠️ In monolithic test |
| TC-DCM-001-07 | Update non-existent setting | 404 Not Found | ⚠️ In monolithic test |
| TC-DCM-001-08 | Delete device setting | 200 OK | ⚠️ In monolithic test |
| TC-DCM-001-09 | Delete non-existent setting | 404 Not Found | 🔲 Not tested |
| TC-DCM-001-10 | Missing auth cookie | 401 Unauthorized | 🔲 Not tested |

---

### UC-DCM-002: VOD Settings Management

**Description**: CRUD operations for Video On Demand settings.

**Actors**: Admin User

**Main Flow**:
1. **Create**: POST `/xconfAdminService/dcm/vodSettings`
2. **Read All**: GET `/xconfAdminService/dcm/vodSettings`
3. **Read One**: GET `/xconfAdminService/dcm/vodSettings/{id}`
4. **Update**: PUT `/xconfAdminService/dcm/vodSettings`
5. **Delete**: DELETE `/xconfAdminService/dcm/vodSettings/{id}`
6. **Export**: GET `/xconfAdminService/dcm/vodSettings/export`

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-DCM-002-01 | Create valid VOD setting | 201 Created | ⚠️ In monolithic test |
| TC-DCM-002-02 | Get VOD settings export | 200 OK + data | ✅ Individual test |
| TC-DCM-002-03 | Export empty result | 200 OK + empty | ✅ Individual test |
| TC-DCM-002-04 | Export with DCM formulas | 200 OK + formulas | ✅ Individual test |
| TC-DCM-002-05 | Export filter by app type | 200 OK + filtered | ✅ Individual test |
| TC-DCM-002-06 | Export missing VOD settings | 200 OK + partial | ✅ Individual test |
| TC-DCM-002-07 | Export verify headers | Correct headers | ✅ Individual test |
| TC-DCM-002-08 | Export missing auth cookie | 401 Unauthorized | ✅ Individual test |
| TC-DCM-002-09 | Export different app types | Correct filtering | ✅ Individual test |
| TC-DCM-002-10 | Export multiple formulas | All included | ✅ Individual test |
| TC-DCM-002-11 | Export validate structure | Valid JSON | ✅ Individual test |

---

### UC-DCM-003: Log Upload Settings Management

**Description**: CRUD operations for log upload configuration.

**Actors**: Admin User

**Main Flow**:
1. **Create**: POST `/xconfAdminService/dcm/logUploadSettings`
2. **Read All**: GET `/xconfAdminService/dcm/logUploadSettings`
3. **Read One**: GET `/xconfAdminService/dcm/logUploadSettings/{id}`
4. **Update**: PUT `/xconfAdminService/dcm/logUploadSettings`
5. **Delete**: DELETE `/xconfAdminService/dcm/logUploadSettings/{id}`

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-DCM-003-01 | Create log upload setting | 201 Created | ⚠️ In monolithic test |
| TC-DCM-003-02 | Get all log upload settings | 200 OK + list | ⚠️ In monolithic test |
| TC-DCM-003-03 | Get log upload setting by ID | 200 OK + entity | ⚠️ In monolithic test |
| TC-DCM-003-04 | Update log upload setting | 200 OK | ⚠️ In monolithic test |
| TC-DCM-003-05 | Delete log upload setting | 200 OK | ⚠️ In monolithic test |

---

### UC-DCM-004: Log Repository Settings Management

**Description**: CRUD operations for upload repository configuration.

**Actors**: Admin User

**Main Flow**:
1. **Create**: POST `/xconfAdminService/dcm/logRepoSettings`
2. **Read All**: GET `/xconfAdminService/dcm/logRepoSettings`
3. **Read One**: GET `/xconfAdminService/dcm/logRepoSettings/{id}`
4. **Update**: PUT `/xconfAdminService/dcm/logRepoSettings`
5. **Delete**: DELETE `/xconfAdminService/dcm/logRepoSettings/{id}`

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-DCM-004-01 | Create log repo setting | 201 Created | ⚠️ In monolithic test |
| TC-DCM-004-02 | Get all log repo settings | 200 OK + list | ⚠️ In monolithic test |
| TC-DCM-004-03 | Get log repo by ID | 200 OK + entity | ⚠️ In monolithic test |
| TC-DCM-004-04 | Update log repo setting | 200 OK | ⚠️ In monolithic test |
| TC-DCM-004-05 | Delete log repo setting | 200 OK | ⚠️ In monolithic test |

---

### UC-DCM-005: DCM Formula Management

**Description**: DCM rule/formula CRUD operations.

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-DCM-005-01 | Create DCM formula | 201 Created | 🔲 Needs test |
| TC-DCM-005-02 | Get all DCM formulas | 200 OK + list | 🔲 Needs test |
| TC-DCM-005-03 | Get DCM formula by ID | 200 OK + entity | 🔲 Needs test |
| TC-DCM-005-04 | Update DCM formula | 200 OK | 🔲 Needs test |
| TC-DCM-005-05 | Delete DCM formula | 200 OK | 🔲 Needs test |

---

## Current Issues

### Issue DCM-001: Monolithic E2E Tests

**File**: `device_settings_e2e_test.go`

**Current Code**:
```go
func TestAllDeviceSettingsApis(t *testing.T) {
    SkipIfMockDatabase(t)
    DeleteAllEntities()          // ❌ Clears ALL tables
    defer DeleteAllEntities()     // ❌ Clears ALL tables
    
    // Tests GET ALL, CREATE, CREATE DUPLICATE, UPDATE, GET BY ID, DELETE
    // All in one function - not idempotent
}
```

**Problem**:
- Single test function does CRUD - cannot run individual tests
- `DeleteAllEntities()` affects other parallel tests
- Not idempotent - depends on previous operations

**Solution**:
```go
// Split into independent tests
func TestCreateDeviceSetting(t *testing.T) {
    id := uuid.New().String()
    cleanup := createTestDeviceSetting(t, id)
    defer cleanup()
    
    // Test create logic
}

func TestGetDeviceSettingById(t *testing.T) {
    id := uuid.New().String()
    cleanup := createTestDeviceSetting(t, id)
    defer cleanup()
    
    // Test get by ID logic
}
```

---

### Issue DCM-002: Double Cleanup Pattern

**File**: `logrepo_settings_service_test.go`

**Current Code**:
```go
func TestSomeFunction(t *testing.T) {
    DeleteAllEntities()           // ❌ Unnecessary
    defer DeleteAllEntities()      // ❌ Should be surgical
    
    // Test logic
}
```

**Solution**:
```go
func TestSomeFunction(t *testing.T) {
    id := uuid.New().String()
    cleanup := createTestEntity(t, id, db.TABLE_UPLOAD_REPOSITORY)
    defer cleanup()  // ✅ Only deletes what was created
    
    // Test logic
}
```

---

### Issue DCM-003: TestMain Mock Timing

**File**: `dcmformula_test.go`

**Current Code**:
```go
func TestMain(m *testing.M) {
    // Config loaded
    // Server created (tries to connect to Cassandra!)
    // Then mock initialized - TOO LATE
}
```

**Solution**:
```go
func TestMain(m *testing.M) {
    // Load config
    
    if IsMockDatabaseEnabled() {
        InitMockDatabase()  // ✅ Initialize mock FIRST
    }
    
    // Now create server (will use mock if enabled)
    server = oshttp.NewWebconfigServer(...)
}
```

---

## Mock Requirements

### MockCachedSimpleDao Methods Needed

| Method | Parameters | Return | Used By |
|--------|------------|--------|---------|
| `GetOne` | tableName, rowKey | interface{}, error | All CRUD |
| `SetOne` | tableName, rowKey, entity | error | Create, Update |
| `DeleteOne` | tableName, rowKey | error | Delete |
| `GetAllAsList` | tableName, maxResults | []interface{}, error | List all |
| `GetAllAsMap` | tableName | map[interface{}]interface{}, error | List as map |
| `RefreshAll` | tableName | error | Cache refresh |

### Test Data Fixtures

```go
// Device Settings fixture
func NewTestDeviceSetting(id string) *logupload.DeviceSettings {
    return &logupload.DeviceSettings{
        ID:                id,
        Name:              "Test_" + id[:8],
        CheckOnReboot:     true,
        SettingsAreActive: true,
        Schedule: logupload.Schedule{
            Type:       "ActNow",
            Expression: "26 4 * * *",
            TimeZone:   "UTC",
        },
        ApplicationType: "stb",
    }
}

// VOD Settings fixture
func NewTestVodSettings(id string) *logupload.VodSettings {
    return &logupload.VodSettings{
        ID:              id,
        Name:            "Test_" + id[:8],
        ApplicationType: "stb",
    }
}

// Log Upload Settings fixture
func NewTestLogUploadSettings(id string) *logupload.LogUploadSettings {
    return &logupload.LogUploadSettings{
        ID:              id,
        Name:            "Test_" + id[:8],
        ApplicationType: "stb",
    }
}

// Upload Repository fixture
func NewTestUploadRepository(id string) *logupload.UploadRepository {
    return &logupload.UploadRepository{
        ID:       id,
        Name:     "Test_" + id[:8],
        Protocol: "HTTP",
        URL:      "http://test.example.com",
    }
}
```

---

## API Endpoints Reference

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/dcm/deviceSettings` | GetAllDeviceSettings | List all device settings |
| GET | `/dcm/deviceSettings/{id}` | GetDeviceSettingById | Get device setting by ID |
| POST | `/dcm/deviceSettings` | CreateDeviceSetting | Create device setting |
| PUT | `/dcm/deviceSettings` | UpdateDeviceSetting | Update device setting |
| DELETE | `/dcm/deviceSettings/{id}` | DeleteDeviceSetting | Delete device setting |
| GET | `/dcm/deviceSettings/size` | GetDeviceSettingsSize | Get count |
| GET | `/dcm/deviceSettings/names` | GetDeviceSettingsNames | Get names list |
| GET | `/dcm/vodSettings` | GetAllVodSettings | List all VOD settings |
| GET | `/dcm/vodSettings/{id}` | GetVodSettingById | Get VOD setting by ID |
| POST | `/dcm/vodSettings` | CreateVodSetting | Create VOD setting |
| PUT | `/dcm/vodSettings` | UpdateVodSetting | Update VOD setting |
| DELETE | `/dcm/vodSettings/{id}` | DeleteVodSetting | Delete VOD setting |
| GET | `/dcm/vodSettings/export` | GetVodSettingsExport | Export VOD settings |
| GET | `/dcm/logUploadSettings` | GetAllLogUploadSettings | List all log upload settings |
| GET | `/dcm/logRepoSettings` | GetAllLogRepoSettings | List all log repo settings |

---

## Coverage Goals

| File | Current | Target | Focus Areas |
|------|---------|--------|-------------|
| device_settings_handler.go | ~60% | 85% | Error paths, edge cases |
| device_settings_service.go | ~55% | 85% | Validation, business logic |
| vod_settings_handler.go | ~70% | 90% | Already well tested |
| vod_settings_service.go | ~60% | 85% | Validation |
| logupload_settings_handler.go | ~50% | 85% | All CRUD paths |
| logrepo_settings_handler.go | ~55% | 85% | All CRUD paths |
| logrepo_settings_service.go | ~65% | 90% | Business logic |
| dcmformula_handler.go | ~40% | 80% | CRUD operations |

---

## Dependencies

### Internal Dependencies
- `common/` - Common structs and utilities
- `shared/logupload/` - LogUpload types and functions
- `http/` - WebconfigServer

### External Dependencies
- `github.com/rdkcentral/xconfwebconfig/db` - Database access
- `github.com/rdkcentral/xconfwebconfig/shared/logupload` - Core types
- `github.com/gorilla/mux` - HTTP routing

---

## Test Execution Commands

```bash
# Run all DCM tests with mock
USE_MOCK_DB=true go test -v ./adminapi/dcm/... -count=1

# Run specific test
USE_MOCK_DB=true go test -v ./adminapi/dcm/... -run TestGetVodSettingExportHandler_Success

# Run with coverage
USE_MOCK_DB=true go test ./adminapi/dcm/... -coverprofile=dcm.out
go tool cover -func=dcm.out | grep -E "handler|service"

# Run with real DB
USE_MOCK_DB=false go test -v ./adminapi/dcm/... -count=1 -timeout=5m
```
