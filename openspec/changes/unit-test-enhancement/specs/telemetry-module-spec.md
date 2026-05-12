# Telemetry Module Specification

## Module Overview

**Package**: `adminapi/telemetry/`  
**Priority**: P0 - Critical  
**Status**: Has mocks, needs idempotent refactoring

The Telemetry module handles telemetry profiles, telemetry rules, and telemetry two (v2) configurations.

---

## Architecture

```
adminapi/telemetry/
├── Telemetry Profiles (v1)
│   ├── telemetry_profile_handler.go
│   ├── telemetry_profile_handler_test.go    # TestMain here
│   ├── telemetry_profile_controller.go
│   ├── telemetry_profile_controller_test.go
│   ├── telemetry_profile_service.go
│   └── telemetry_profile_service_test.go
│
├── Telemetry Rules (v1)
│   ├── telemetry_rule_handler.go
│   ├── telemetry_rule_handler_test.go
│   └── telemetry_v2_rule_service_test.go
│
└── Telemetry Two (v2)
    ├── telemetry_two_loguploader_handler.go
    ├── telemetry_two_loguploader_handler_test.go
    ├── telemetry_two_profile_handler.go
    ├── telemetry_two_profile_handler_test.go
    ├── telemetry_two_rule_handler.go
    ├── telemetry_two_rule_hanlder_test.go
    └── telemetry_two_dao_test.go
```

---

## Database Tables Used

| Table Name | Operations | Entity Type | Version |
|------------|------------|-------------|---------|
| `TABLE_TELEMETRY_PROFILES` | Read/Write | `logupload.TelemetryProfile` | v1 |
| `TABLE_TELEMETRY_RULES` | CRUD | `logupload.TelemetryRule` | v1 |
| `TABLE_PERMANENT_TELEMETRY_PROFILES` | CRUD | `logupload.PermanentTelemetryProfile` | v1 |
| `TABLE_TELEMETRY_TWO_PROFILES` | CRUD | `logupload.TelemetryTwoProfile` | v2 |
| `TABLE_TELEMETRY_TWO_RULES` | CRUD | `logupload.TelemetryTwoRule` | v2 |
| `TABLE_TELEMETRY_CHANGES` | CRUD | `change.Change` | Change mgmt |
| `TABLE_TELEMETRY_APPROVED_CHANGES` | CRUD | `change.Change` | Change mgmt |
| `TABLE_TELEMETRY_TWO_CHANGES` | CRUD | `change.Change` | v2 changes |
| `TABLE_TELEMETRY_APPROVED_TWO_CHANGES` | CRUD | `change.Change` | v2 changes |

---

## Use Cases

### UC-TEL-001: Telemetry Profile Management (v1)

**Description**: CRUD operations for telemetry profiles.

**API Endpoints**:
- GET `/telemetry/profile` - List all profiles
- GET `/telemetry/profile/{id}` - Get profile by ID
- POST `/telemetry/profile` - Create profile
- PUT `/telemetry/profile` - Update profile
- DELETE `/telemetry/profile/{id}` - Delete profile
- GET `/telemetry/profile/export` - Export profiles

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-TEL-001-01 | Create telemetry profile | 201 Created | ⚠️ Needs refactor |
| TC-TEL-001-02 | Create duplicate profile | 409 Conflict | ⚠️ Needs refactor |
| TC-TEL-001-03 | Get all profiles | 200 OK + list | ⚠️ Needs refactor |
| TC-TEL-001-04 | Get profile by ID | 200 OK + entity | ⚠️ Needs refactor |
| TC-TEL-001-05 | Get non-existent profile | 404 Not Found | 🔲 Not tested |
| TC-TEL-001-06 | Update profile | 200 OK | ⚠️ Needs refactor |
| TC-TEL-001-07 | Delete profile | 200 OK | ⚠️ Needs refactor |
| TC-TEL-001-08 | Delete profile with rules | 409 Conflict | 🔲 Not tested |
| TC-TEL-001-09 | Export profiles | 200 OK + data | ⚠️ Needs refactor |
| TC-TEL-001-10 | Export by app type | Filtered data | 🔲 Not tested |

---

### UC-TEL-002: Telemetry Rules Management (v1)

**Description**: CRUD operations for telemetry rules.

**API Endpoints**:
- GET `/telemetry/rule` - List all rules
- GET `/telemetry/rule/{id}` - Get rule by ID
- POST `/telemetry/rule` - Create rule
- PUT `/telemetry/rule` - Update rule
- DELETE `/telemetry/rule/{id}` - Delete rule

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-TEL-002-01 | Create telemetry rule | 201 Created | ⚠️ Needs refactor |
| TC-TEL-002-02 | Create rule with invalid profile | 400 Bad Request | 🔲 Not tested |
| TC-TEL-002-03 | Get all rules | 200 OK + list | ⚠️ Needs refactor |
| TC-TEL-002-04 | Get rule by ID | 200 OK + entity | ⚠️ Needs refactor |
| TC-TEL-002-05 | Update rule | 200 OK | ⚠️ Needs refactor |
| TC-TEL-002-06 | Delete rule | 200 OK | ⚠️ Needs refactor |

---

### UC-TEL-003: Telemetry Two Profile Management (v2)

**Description**: CRUD operations for Telemetry 2.0 profiles.

**API Endpoints**:
- GET `/telemetry/v2/profile` - List all v2 profiles
- GET `/telemetry/v2/profile/{id}` - Get v2 profile by ID
- POST `/telemetry/v2/profile` - Create v2 profile
- PUT `/telemetry/v2/profile` - Update v2 profile
- DELETE `/telemetry/v2/profile/{id}` - Delete v2 profile

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-TEL-003-01 | Create v2 profile | 201 Created | ⚠️ Needs refactor |
| TC-TEL-003-02 | Create duplicate v2 profile | 409 Conflict | 🔲 Not tested |
| TC-TEL-003-03 | Get all v2 profiles | 200 OK + list | ⚠️ Needs refactor |
| TC-TEL-003-04 | Get v2 profile by ID | 200 OK + entity | ⚠️ Needs refactor |
| TC-TEL-003-05 | Update v2 profile | 200 OK | ⚠️ Needs refactor |
| TC-TEL-003-06 | Delete v2 profile | 200 OK | ⚠️ Needs refactor |

---

### UC-TEL-004: Telemetry Two Rules Management (v2)

**Description**: CRUD operations for Telemetry 2.0 rules.

**API Endpoints**:
- GET `/telemetry/v2/rule` - List all v2 rules
- GET `/telemetry/v2/rule/{id}` - Get v2 rule by ID
- POST `/telemetry/v2/rule` - Create v2 rule
- PUT `/telemetry/v2/rule` - Update v2 rule
- DELETE `/telemetry/v2/rule/{id}` - Delete v2 rule

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-TEL-004-01 | Create v2 rule | 201 Created | ⚠️ Needs refactor |
| TC-TEL-004-02 | Create v2 rule invalid profile | 400 Bad Request | 🔲 Not tested |
| TC-TEL-004-03 | Get all v2 rules | 200 OK + list | ⚠️ Needs refactor |
| TC-TEL-004-04 | Get v2 rule by ID | 200 OK + entity | ⚠️ Needs refactor |
| TC-TEL-004-05 | Update v2 rule | 200 OK | ⚠️ Needs refactor |
| TC-TEL-004-06 | Delete v2 rule | 200 OK | ⚠️ Needs refactor |

---

### UC-TEL-005: Telemetry Log Uploader (v2)

**Description**: Log uploader configuration management.

**API Endpoints**:
- GET `/telemetry/v2/logUploader` - Get log uploader config
- POST `/telemetry/v2/logUploader` - Create/Update config

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-TEL-005-01 | Get log uploader config | 200 OK | ⚠️ Needs refactor |
| TC-TEL-005-02 | Create log uploader config | 201 Created | ⚠️ Needs refactor |
| TC-TEL-005-03 | Update log uploader config | 200 OK | ⚠️ Needs refactor |

---

## Current Issues

### Issue TEL-001: TestMain Mock Timing

**File**: `telemetry_profile_handler_test.go`

**Current Code**:
```go
func TestMain(m *testing.M) {
    // Config loaded
    // Server created (connects to Cassandra!)
    // Mock not initialized first
}
```

**Solution**:
```go
func TestMain(m *testing.M) {
    // Load config
    
    if IsMockDatabaseEnabled() {
        InitMockDatabase()  // Initialize mock FIRST
    }
    
    // Create server
    server = oshttp.NewWebconfigServer(...)
}
```

---

### Issue TEL-002: Shared Test Data

**Problem**: Tests share telemetry profiles/rules, causing test pollution.

**Solution**: Each test creates unique entities with UUIDs.

```go
func TestCreateTelemetryProfile(t *testing.T) {
    id := uuid.New().String()
    profile := NewTestTelemetryProfile(id)
    
    cleanup := createTestEntity(t, profile, db.TABLE_TELEMETRY_TWO_PROFILES)
    defer cleanup()
    
    // Test logic
}
```

---

## Test Data Fixtures

```go
// Telemetry Profile v1 fixture
func NewTestTelemetryProfile(id string) *logupload.PermanentTelemetryProfile {
    return &logupload.PermanentTelemetryProfile{
        ID:              id,
        Name:            "Test_Profile_" + id[:8],
        ApplicationType: "stb",
        TelemetryProfile: []logupload.TelemetryElement{
            {
                Header: "Test_Header",
                Content: "Test_Content",
                Type:    "Test_Type",
            },
        },
    }
}

// Telemetry Rule v1 fixture
func NewTestTelemetryRule(id, profileId string) *logupload.TelemetryRule {
    return &logupload.TelemetryRule{
        ID:              id,
        Name:            "Test_Rule_" + id[:8],
        BoundTelemetryId: profileId,
        ApplicationType: "stb",
    }
}

// Telemetry Two Profile v2 fixture
func NewTestTelemetryTwoProfile(id string) *logupload.TelemetryTwoProfile {
    return &logupload.TelemetryTwoProfile{
        ID:              id,
        Name:            "Test_V2_Profile_" + id[:8],
        ApplicationType: "stb",
    }
}

// Telemetry Two Rule v2 fixture
func NewTestTelemetryTwoRule(id, profileId string) *logupload.TelemetryTwoRule {
    return &logupload.TelemetryTwoRule{
        ID:                id,
        Name:              "Test_V2_Rule_" + id[:8],
        BoundTelemetryIds: []string{profileId},
        ApplicationType:   "stb",
    }
}
```

---

## Dependencies

### Internal Dependencies
- `adminapi/change/` - Change management integration
- `shared/logupload/` - Shared telemetry types
- `shared/change/` - Change entity types

### External Dependencies
- `github.com/rdkcentral/xconfwebconfig/db`
- `github.com/rdkcentral/xconfwebconfig/shared/logupload`
- `github.com/rdkcentral/xconfwebconfig/shared/change`

---

## Coverage Goals

| File | Current | Target | Focus Areas |
|------|---------|--------|-------------|
| telemetry_profile_handler.go | ~55% | 85% | Error paths |
| telemetry_profile_service.go | ~60% | 85% | Business logic |
| telemetry_rule_handler.go | ~50% | 85% | CRUD paths |
| telemetry_two_profile_handler.go | ~55% | 85% | v2 CRUD |
| telemetry_two_rule_handler.go | ~50% | 85% | v2 rules |
| telemetry_two_loguploader_handler.go | ~45% | 80% | Config mgmt |

---

## Test Execution Commands

```bash
# Run all telemetry tests with mock
USE_MOCK_DB=true go test -v ./adminapi/telemetry/... -count=1

# Run specific test pattern
USE_MOCK_DB=true go test -v ./adminapi/telemetry/... -run "TelemetryTwo" -count=1

# Run with coverage
USE_MOCK_DB=true go test ./adminapi/telemetry/... -coverprofile=telemetry.out
go tool cover -func=telemetry.out | grep -E "handler|service"

# Generate HTML report
go tool cover -html=telemetry.out -o telemetry_coverage.html
```
