# Use Cases Specification

## Overview

Complete catalog of all use cases for xconfadmin unit test enhancement, organized by module and functionality.

---

## Use Case Index

```
UC-DCM-001    Device Settings Management
UC-DCM-002    VOD Settings Management
UC-DCM-003    Log Upload Settings Management
UC-DCM-004    Log Repository Settings Management
UC-DCM-005    DCM Formula Management

UC-QRY-001    Model Management
UC-QRY-002    Environment Management
UC-QRY-003    Firmware Configuration
UC-QRY-004    Firmware Rules
UC-QRY-005    Namespaced Lists
UC-QRY-006    Filters Management
UC-QRY-007    Percentage Beans
UC-QRY-008    AMV Rules

UC-TEL-001    Telemetry Profile Management (v1)
UC-TEL-002    Telemetry Rules Management (v1)
UC-TEL-003    Telemetry Two Profile Management (v2)
UC-TEL-004    Telemetry Two Rules Management (v2)
UC-TEL-005    Telemetry Log Uploader

UC-CHG-001    Change Request Management
UC-CHG-002    Change Approval Workflow
UC-CHG-003    Telemetry Change Management

UC-SET-001    Setting Profile Management
UC-SET-002    Setting Rule Management

UC-RFC-001    Feature Management
UC-RFC-002    Feature Rule Management

UC-TAG-001    Tag Management
UC-TAG-002    Tag Member Management

UC-MOCK-001   Mock Database Operations
UC-MOCK-002   Real Database Operations
UC-MOCK-003   Database Mode Switching
```

---

## Detailed Use Cases

### UC-DCM-001: Device Settings Management

**ID**: UC-DCM-001  
**Module**: adminapi/dcm/  
**Priority**: High  
**Complexity**: Medium

#### Description
CRUD operations for managing device settings that control how devices check in with the server.

#### Actors
- Admin User (primary)
- System (automated processes)

#### Preconditions
1. User is authenticated with valid SAT token
2. Application type cookie is set (stb/xhome/etc)
3. User has write permissions for device settings

#### Main Flow

**Create Device Setting (UC-DCM-001-A)**
1. User sends POST request to `/xconfAdminService/dcm/deviceSettings`
2. System validates JSON payload
3. System checks for duplicate ID
4. System checks for duplicate name
5. System stores entity in TABLE_DEVICE_SETTINGS
6. System returns 201 Created with entity

**Get All Device Settings (UC-DCM-001-B)**
1. User sends GET request to `/xconfAdminService/dcm/deviceSettings`
2. System retrieves all settings from cache
3. System filters by application type
4. System returns 200 OK with list

**Get Device Setting by ID (UC-DCM-001-C)**
1. User sends GET request to `/xconfAdminService/dcm/deviceSettings/{id}`
2. System looks up entity by ID
3. System returns 200 OK with entity

**Update Device Setting (UC-DCM-001-D)**
1. User sends PUT request to `/xconfAdminService/dcm/deviceSettings`
2. System validates JSON payload
3. System verifies entity exists
4. System updates entity
5. System returns 200 OK

**Delete Device Setting (UC-DCM-001-E)**
1. User sends DELETE request to `/xconfAdminService/dcm/deviceSettings/{id}`
2. System verifies entity exists
3. System checks for dependencies (DCM rules)
4. System deletes entity
5. System returns 200 OK

#### Alternative Flows

| ID | Condition | Response |
|----|-----------|----------|
| UC-DCM-001-A1 | Duplicate ID | 409 Conflict |
| UC-DCM-001-A2 | Invalid JSON | 400 Bad Request |
| UC-DCM-001-A3 | Missing required field | 400 Bad Request |
| UC-DCM-001-C1 | Entity not found | 404 Not Found |
| UC-DCM-001-D1 | Entity not found | 404 Not Found |
| UC-DCM-001-E1 | Entity not found | 404 Not Found |
| UC-DCM-001-E2 | Has dependencies | 409 Conflict |

#### Test Cases

| TC ID | Flow | Input | Expected Output | Mock | Real |
|-------|------|-------|-----------------|------|------|
| TC-001-01 | A | Valid device setting | 201 Created | ✅ | ✅ |
| TC-001-02 | A1 | Duplicate ID | 409 Conflict | ✅ | ✅ |
| TC-001-03 | A2 | Malformed JSON | 400 Bad Request | ✅ | ✅ |
| TC-001-04 | B | No filters | 200 + list | ✅ | ✅ |
| TC-001-05 | B | App type filter | 200 + filtered | ✅ | ✅ |
| TC-001-06 | C | Valid ID | 200 + entity | ✅ | ✅ |
| TC-001-07 | C1 | Invalid ID | 404 Not Found | ✅ | ✅ |
| TC-001-08 | D | Valid update | 200 OK | ✅ | ✅ |
| TC-001-09 | D1 | Non-existent ID | 404 Not Found | ✅ | ✅ |
| TC-001-10 | E | Valid delete | 200 OK | ✅ | ✅ |
| TC-001-11 | E1 | Non-existent ID | 404 Not Found | ✅ | ✅ |
| TC-001-12 | E2 | Has DCM rule ref | 409 Conflict | ✅ | ✅ |

---

### UC-QRY-004: Firmware Rules

**ID**: UC-QRY-004  
**Module**: adminapi/queries/  
**Priority**: High  
**Complexity**: High

#### Description
CRUD operations for firmware rules that determine which firmware version a device should receive.

#### Actors
- Admin User
- Firmware Release Manager

#### Preconditions
1. User is authenticated
2. Referenced models exist
3. Referenced firmware configs exist

#### Main Flow

**Create Firmware Rule (UC-QRY-004-A)**
1. User sends POST request to `/xconfAdminService/queries/rules`
2. System validates rule structure
3. System validates referenced entities (model, environment, config)
4. System checks rule condition validity
5. System stores rule in TABLE_FIRMWARE_RULES
6. System returns 201 Created

**Get Firmware Rules (UC-QRY-004-B)**
1. User sends GET request to `/xconfAdminService/queries/rules`
2. System retrieves rules from cache
3. System applies pagination if requested
4. System returns 200 OK with rules

**Apply Firmware Rule Template (UC-QRY-004-F)**
1. User sends POST to `/xconfAdminService/queries/rules/template/apply`
2. System loads template
3. System substitutes parameters
4. System creates rule from template
5. System returns 201 Created

#### Rule Validation Logic

```go
type FirmwareRuleValidation struct {
    // Required fields
    ID              string // Must be unique
    Name            string // Must be unique per app type
    Type            string // Must be valid rule type
    ApplicationType string // Must be valid app type
    
    // Rule condition validation
    Rule struct {
        // Condition must reference valid entities
        // Model ID must exist in TABLE_MODELS
        // Environment ID must exist in TABLE_ENVIRONMENTS
        // FirmwareConfig ID must exist in TABLE_FIRMWARE_CONFIGS
    }
}

// Valid rule types
var ValidRuleTypes = []string{
    "MODEL_RULE",
    "MAC_RULE", 
    "IP_RULE",
    "ENV_MODEL_RULE",
    "TIME_FILTER",
    "GLOBAL_PERCENT",
}
```

#### Test Cases

| TC ID | Flow | Input | Expected Output | Notes |
|-------|------|-------|-----------------|-------|
| TC-004-01 | A | Valid MODEL_RULE | 201 Created | Model must exist |
| TC-004-02 | A | Valid MAC_RULE | 201 Created | MAC list must exist |
| TC-004-03 | A | Valid IP_RULE | 201 Created | IP group must exist |
| TC-004-04 | A | Invalid model ref | 400 Bad Request | Non-existent model |
| TC-004-05 | A | Invalid rule type | 400 Bad Request | Unknown type |
| TC-004-06 | B | No filters | 200 + list | Returns all rules |
| TC-004-07 | B | By rule type | 200 + filtered | Filter works |
| TC-004-08 | F | Valid template | 201 Created | Template applied |
| TC-004-09 | F | Invalid template | 400 Bad Request | Missing params |

---

### UC-CHG-001: Change Request Management

**ID**: UC-CHG-001  
**Module**: adminapi/change/  
**Priority**: High  
**Complexity**: High

#### Description
Create and manage change requests for configuration modifications that require approval.

#### Actors
- Change Author (creates change)
- Change Approver (approves/rejects)
- System (tracks changes)

#### Preconditions
1. User is authenticated
2. Entity being changed exists
3. User has permission to create change requests

#### Main Flow

**Create Change Request (UC-CHG-001-A)**
1. User modifies entity via normal API
2. System detects change requires approval
3. System creates Change record in TABLE_XCONF_CHANGE
4. System stores old and new entity versions
5. System returns change ID to user

**Get Pending Changes (UC-CHG-001-B)**
1. User sends GET to `/xconfAdminService/change/pending`
2. System retrieves changes from TABLE_XCONF_CHANGE
3. System returns list of pending changes

**Approve Change (UC-CHG-001-C)**
1. Approver sends POST to `/xconfAdminService/change/approve/{id}`
2. System verifies change exists
3. System applies change to production table
4. System moves change to TABLE_XCONF_APPROVED_CHANGE
5. System returns 200 OK

**Reject Change (UC-CHG-001-D)**
1. Approver sends POST to `/xconfAdminService/change/reject/{id}`
2. System verifies change exists
3. System deletes change from TABLE_XCONF_CHANGE
4. System returns 200 OK

#### Change Entity Structure

```go
type Change struct {
    ID           string      `json:"id"`           // UUID
    EntityId     string      `json:"entityId"`     // ID of changed entity
    EntityType   string      `json:"entityType"`   // Type (e.g., "FirmwareRule")
    Operation    string      `json:"operation"`    // CREATE/UPDATE/DELETE
    OldEntity    interface{} `json:"oldEntity"`    // Previous state (null for CREATE)
    NewEntity    interface{} `json:"newEntity"`    // New state (null for DELETE)
    Author       string      `json:"author"`       // Who made the change
    ApprovedUser string      `json:"approvedUser"` // Who approved (if approved)
    Updated      int64       `json:"updated"`      // Timestamp
}
```

#### Test Cases

| TC ID | Flow | Input | Expected Output | Notes |
|-------|------|-------|-----------------|-------|
| TC-CHG-01 | A | Valid CREATE change | Change created | New entity |
| TC-CHG-02 | A | Valid UPDATE change | Change created | Old + new entity |
| TC-CHG-03 | A | Valid DELETE change | Change created | Old entity only |
| TC-CHG-04 | B | No pending | 200 + empty | No changes |
| TC-CHG-05 | B | Has pending | 200 + list | Returns changes |
| TC-CHG-06 | C | Valid approve | 200 OK | Entity updated |
| TC-CHG-07 | C | Non-existent | 404 Not Found | Invalid ID |
| TC-CHG-08 | D | Valid reject | 200 OK | Change deleted |

---

### UC-MOCK-001: Mock Database Operations

**ID**: UC-MOCK-001  
**Module**: adminapi/dcm/mocks/  
**Priority**: Critical  
**Complexity**: Medium

#### Description
Verify mock database implementation correctly simulates real database behavior.

#### Actors
- Test Framework (automated)
- Developer (runs tests)

#### Preconditions
1. USE_MOCK_DB=true environment variable set
2. Mock DAO initialized before tests

#### Main Flow

**Mock GetOne (UC-MOCK-001-A)**
1. Test creates entity via MockCachedSimpleDao.SetOne()
2. Test retrieves entity via MockCachedSimpleDao.GetOne()
3. Mock returns stored entity
4. Test verifies entity matches

**Mock SetOne (UC-MOCK-001-B)**
1. Test calls MockCachedSimpleDao.SetOne() with entity
2. Mock stores entity in memory map
3. Test verifies no error returned
4. Test verifies entity retrievable

**Mock DeleteOne (UC-MOCK-001-C)**
1. Test creates entity via SetOne()
2. Test deletes entity via DeleteOne()
3. Mock removes from memory map
4. Test verifies GetOne returns "not found"

**Mock GetAllAsList (UC-MOCK-001-D)**
1. Test creates multiple entities via SetOne()
2. Test calls GetAllAsList()
3. Mock returns all entities in table
4. Test verifies count and contents

#### Test Cases

| TC ID | Operation | Input | Expected Output | Notes |
|-------|-----------|-------|-----------------|-------|
| TC-MOCK-01 | SetOne | Valid entity | nil error | Stored |
| TC-MOCK-02 | GetOne | Existing key | Entity, nil | Found |
| TC-MOCK-03 | GetOne | Missing key | nil, error | Not found |
| TC-MOCK-04 | DeleteOne | Existing key | nil error | Deleted |
| TC-MOCK-05 | DeleteOne | Missing key | nil error | Idempotent |
| TC-MOCK-06 | GetAllAsList | 5 entities | List of 5 | All returned |
| TC-MOCK-07 | GetAllAsList | Empty table | Empty list | No error |
| TC-MOCK-08 | GetAllAsMap | 3 entities | Map of 3 | All mapped |
| TC-MOCK-09 | RefreshAll | Any table | nil error | No-op in mock |

---

### UC-MOCK-003: Database Mode Switching

**ID**: UC-MOCK-003  
**Module**: All modules  
**Priority**: Critical  
**Complexity**: Low

#### Description
Verify tests can switch between mock and real database modes seamlessly.

#### Actors
- CI/CD Pipeline
- Developer

#### Main Flow

**Run with Mock (UC-MOCK-003-A)**
1. Set USE_MOCK_DB=true
2. Run tests
3. Tests use MockCachedSimpleDao
4. Tests complete in <30 seconds

**Run with Real DB (UC-MOCK-003-B)**
1. Set USE_MOCK_DB=false
2. Ensure Cassandra is running
3. Run tests
4. Tests use real CachedSimpleDao
5. Tests complete (may take minutes)

**Skip Integration Tests in Mock Mode (UC-MOCK-003-C)**
1. Test calls SkipIfMockDatabase(t)
2. If USE_MOCK_DB=true, test skipped
3. If USE_MOCK_DB=false, test runs

#### Test Cases

| TC ID | Mode | Test Type | Expected |
|-------|------|-----------|----------|
| TC-MODE-01 | Mock | Unit test | Runs |
| TC-MODE-02 | Mock | Integration | Skipped |
| TC-MODE-03 | Real | Unit test | Runs |
| TC-MODE-04 | Real | Integration | Runs |
| TC-MODE-05 | Mock | Full suite | <1 min |
| TC-MODE-06 | Real | Full suite | <15 min |

---

## Use Case Dependencies

```
UC-MOCK-001 ─────────────────────────────────────┐
                                                  │
UC-DCM-001 ──► UC-DCM-005 (DCM rules reference   ├─► All Module Tests
              device settings)                    │
                                                  │
UC-QRY-001 ──► UC-QRY-004 (Firmware rules        │
              reference models)                   │
                                                  │
UC-QRY-004 ──► UC-CHG-001 (Changes reference     │
              firmware rules)                    ─┘
```

---

## Test Coverage Matrix

| Use Case | Unit Tests | Integration | Mock Support | Real DB Support |
|----------|------------|-------------|--------------|-----------------|
| UC-DCM-001 | ⚠️ Partial | ⚠️ Monolithic | ✅ | ✅ |
| UC-DCM-002 | ✅ Good | ⚠️ Monolithic | ✅ | ✅ |
| UC-DCM-003 | ⚠️ Partial | ⚠️ Monolithic | ✅ | ✅ |
| UC-DCM-004 | ⚠️ Partial | ⚠️ Monolithic | ✅ | ✅ |
| UC-QRY-001 | ⚠️ Partial | ⚠️ Shared | ✅ | ✅ |
| UC-QRY-004 | ⚠️ Partial | ⚠️ Shared | ✅ | ✅ |
| UC-TEL-001 | ⚠️ Partial | ⚠️ Shared | ✅ | ✅ |
| UC-CHG-001 | ⚠️ Partial | ❌ Missing | 🔲 | ✅ |
| UC-MOCK-001 | ✅ Good | N/A | ✅ | N/A |
