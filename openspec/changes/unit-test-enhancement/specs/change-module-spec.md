# Change Module Specification

## Module Overview

**Package**: `adminapi/change/`  
**Priority**: P1 - High  
**Status**: No mocks, needs test_utils.go

The Change module handles change management workflow - creating, approving, and rejecting configuration changes that require review.

---

## Architecture

```
adminapi/change/
├── Core Change Management
│   ├── change_handler.go               # Change CRUD handlers
│   ├── change_handler_test.go          # TestMain, handler tests
│   ├── change_service.go               # Business logic
│   └── change_service_test.go          # Service tests
│
└── Telemetry Changes
    ├── telemetry_profile_handler.go
    ├── telemetry_profile_handler_test.go
    ├── telemetry_two_change_handler.go
    ├── telemetry_two_change_handler_test.go
    ├── telemetry_two_change_service.go
    └── telemetry_two_change_service_test.go
```

---

## Database Tables Used

| Table Name | Operations | Entity Type | DAO Type |
|------------|------------|-------------|----------|
| `TABLE_XCONF_CHANGE` | CRUD | `change.Change` | **SimpleDao** |
| `TABLE_XCONF_APPROVED_CHANGE` | CRUD | `change.Change` | **SimpleDao** |
| `TABLE_TELEMETRY_CHANGES` | CRUD | `change.Change` | **SimpleDao** |
| `TABLE_TELEMETRY_APPROVED_CHANGES` | CRUD | `change.Change` | **SimpleDao** |
| `TABLE_TELEMETRY_TWO_CHANGES` | CRUD | `change.Change` | **SimpleDao** |
| `TABLE_TELEMETRY_APPROVED_TWO_CHANGES` | CRUD | `change.Change` | **SimpleDao** |

**CRITICAL**: This module uses `db.GetSimpleDao()` NOT `db.GetCachedSimpleDao()`!

---

## Use Cases

### UC-CHG-001: Change Request Creation

**Description**: Create change requests for configuration modifications.

**API Endpoints**:
- POST `/change/create` - Create change request
- GET `/change/pending` - List pending changes
- GET `/change/{id}` - Get change by ID

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-CHG-001-01 | Create valid change | 201 Created | ⚠️ Needs mock |
| TC-CHG-001-02 | Get all pending | 200 OK + list | ⚠️ Needs mock |
| TC-CHG-001-03 | Get change by ID | 200 OK + entity | ⚠️ Needs mock |
| TC-CHG-001-04 | Get non-existent | 404 Not Found | 🔲 Not tested |

---

### UC-CHG-002: Change Approval Workflow

**Description**: Approve or reject pending changes.

**API Endpoints**:
- POST `/change/approve/{id}` - Approve change
- POST `/change/reject/{id}` - Reject change
- GET `/change/approved` - List approved changes

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-CHG-002-01 | Approve valid change | 200 OK | ⚠️ Needs mock |
| TC-CHG-002-02 | Approve non-existent | 404 Not Found | 🔲 Not tested |
| TC-CHG-002-03 | Reject valid change | 200 OK | ⚠️ Needs mock |
| TC-CHG-002-04 | Get approved changes | 200 OK + list | ⚠️ Needs mock |

---

### UC-CHG-003: Telemetry Two Changes

**Description**: Change management for Telemetry 2.0 profiles.

**API Endpoints**:
- POST `/change/telemetry/v2/pending` - List pending
- POST `/change/telemetry/v2/approve/{id}` - Approve
- POST `/change/telemetry/v2/reject/{id}` - Reject

**Test Scenarios**:

| ID | Scenario | Expected | Current Status |
|----|----------|----------|----------------|
| TC-CHG-003-01 | Get telemetry v2 pending | 200 OK + list | ⚠️ Needs mock |
| TC-CHG-003-02 | Approve telemetry v2 | 200 OK | ⚠️ Needs mock |
| TC-CHG-003-03 | Reject telemetry v2 | 200 OK | ⚠️ Needs mock |

---

## Current Issues

### Issue CHG-001: Uses SimpleDao Instead of CachedSimpleDao

**File**: `shared/change/change.go`

**Current Code**:
```go
// Uses GetSimpleDao() - different DAO type!
db.GetSimpleDao().GetOne(db.TABLE_XCONF_CHANGE, id)
db.GetSimpleDao().SetOne(db.TABLE_XCONF_CHANGE, id, entity)
```

**Impact**: Need separate MockSimpleDao implementation.

**Solution**:
1. Create `mock_simple_dao.go`
2. Implement SimpleDao interface
3. Register mock in test setup

---

### Issue CHG-002: No test_utils.go

**Solution**: Create test_utils.go with mock support.

```go
// adminapi/change/test_utils.go
package change

import (
    "os"
    "testing"
    "github.com/rdkcentral/xconfadmin/adminapi/dcm/mocks"
)

var mockSimpleDao *mocks.MockSimpleDao

func IsMockDatabaseEnabled() bool {
    return os.Getenv("USE_MOCK_DB") == "true"
}

func InitMockDatabase() {
    mockSimpleDao = mocks.NewMockSimpleDao()
    db.SetMockSimpleDao(mockSimpleDao)
}

func ClearMockDatabase() {
    if mockSimpleDao != nil {
        mockSimpleDao.Clear()
    }
}
```

---

## Test Data Fixtures

```go
// Change fixture
func NewTestChange(id, entityId, entityType, operation string) *change.Change {
    return &change.Change{
        ID:         id,
        EntityId:   entityId,
        EntityType: entityType,
        Operation:  operation,
        Author:     "test-user",
        Updated:    time.Now().UnixMilli(),
    }
}

// Create change fixture
func NewTestCreateChange(id string, entity interface{}) *change.Change {
    return &change.Change{
        ID:         id,
        EntityId:   uuid.New().String(),
        EntityType: "FirmwareRule",
        Operation:  "CREATE",
        OldEntity:  nil,
        NewEntity:  entity,
        Author:     "test-user",
        Updated:    time.Now().UnixMilli(),
    }
}

// Update change fixture
func NewTestUpdateChange(id string, oldEntity, newEntity interface{}) *change.Change {
    return &change.Change{
        ID:         id,
        EntityId:   uuid.New().String(),
        EntityType: "FirmwareRule",
        Operation:  "UPDATE",
        OldEntity:  oldEntity,
        NewEntity:  newEntity,
        Author:     "test-user",
        Updated:    time.Now().UnixMilli(),
    }
}

// Delete change fixture
func NewTestDeleteChange(id string, entity interface{}) *change.Change {
    return &change.Change{
        ID:         id,
        EntityId:   uuid.New().String(),
        EntityType: "FirmwareRule",
        Operation:  "DELETE",
        OldEntity:  entity,
        NewEntity:  nil,
        Author:     "test-user",
        Updated:    time.Now().UnixMilli(),
    }
}
```

---

## Coverage Goals

| File | Current | Target | Focus Areas |
|------|---------|--------|-------------|
| change_handler.go | ~50% | 85% | All endpoints |
| change_service.go | ~55% | 85% | Approval flow |
| telemetry_two_change_handler.go | ~45% | 85% | v2 changes |
| telemetry_two_change_service.go | ~50% | 85% | v2 service |

---

## Test Execution Commands

```bash
# Run change module tests with mock
USE_MOCK_DB=true go test -v ./adminapi/change/... -count=1

# Run with coverage
USE_MOCK_DB=true go test ./adminapi/change/... -coverprofile=change.out
go tool cover -func=change.out
```
