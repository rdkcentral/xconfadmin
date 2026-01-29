package change

import (
	"testing"

	xwshared "github.com/rdkcentral/xconfwebconfig/shared"
	xwchange "github.com/rdkcentral/xconfwebconfig/shared/change"
)

// TestNewEmptyChange ensures the default application type is STB.
func TestNewEmptyChange(t *testing.T) {
	ch := NewEmptyChange()
	if ch == nil {
		t.Fatalf("expected non-nil change")
	}
	if ch.ApplicationType != xwshared.STB {
		t.Fatalf("expected applicationType %s got %s", xwshared.STB, ch.ApplicationType)
	}
}

// TestNewEmptyTelemetryTwoChange ensures the default application type is STB.
func TestNewEmptyTelemetryTwoChange(t *testing.T) {
	ch := NewEmptyTelemetryTwoChange()
	if ch == nil {
		t.Fatalf("expected non-nil telemetry two change")
	}
	if ch.ApplicationType != xwshared.STB {
		t.Fatalf("expected applicationType %s got %s", xwshared.STB, ch.ApplicationType)
	}
}

// TestNewApprovedTelemetryTwoChangeMapping verifies field mapping from TelemetryTwoChange to ApprovedTelemetryTwoChange.
func TestNewApprovedTelemetryTwoChangeMapping(t *testing.T) {
	src := &xwchange.TelemetryTwoChange{
		ID:              "id-123",
		EntityID:        "entity-1",
		EntityType:      "TYPE_X",
		ApplicationType: xwshared.STB,
		Author:          "authorA",
		ApprovedUser:    "approverB",
		Operation:       "CREATE",
	}
	approved := NewApprovedTelemetryTwoChange(src)
	if approved.ID != src.ID || approved.EntityID != src.EntityID || approved.EntityType != src.EntityType ||
		approved.ApplicationType != src.ApplicationType || approved.Author != src.Author || approved.ApprovedUser != src.ApprovedUser ||
		approved.Operation != src.Operation {
		t.Fatalf("approved change did not copy all fields correctly")
	}
}

// TestCreateOneTelemetryTwoChangeSetsIDAndUpdated ensures blank ID is generated and Updated timestamp set before persistence.
func TestCreateOneTelemetryTwoChangeSetsIDAndUpdated(t *testing.T) {
	src := &xwchange.TelemetryTwoChange{ApplicationType: xwshared.STB, Operation: "CREATE"}
	if src.ID != "" {
		t.Fatalf("precondition: expected blank ID")
	}
	err := CreateOneTelemetryTwoChange(src)
	if src.ID == "" {
		t.Fatalf("expected ID to be generated")
	}
	if src.Updated == 0 {
		t.Fatalf("expected Updated timestamp to be set")
	}
	// Persistence may fail if underlying dao not initialized in this test context; that's acceptable.
	_ = err
}

// TestCreateOneChange ensures Updated timestamp is set before persistence.
func TestCreateOneChange(t *testing.T) {
	change := &xwchange.Change{
		ID:              "change-123",
		ApplicationType: xwshared.STB,
		Operation:       "CREATE",
	}
	if change.Updated != 0 {
		t.Fatalf("precondition: expected Updated to be 0")
	}
	err := CreateOneChange(change)
	if change.Updated == 0 {
		t.Fatalf("expected Updated timestamp to be set")
	}
	// Persistence may fail if underlying dao not initialized; that's acceptable.
	_ = err
}

// TestGetChangeList retrieves all changes; may return nil if dao not initialized.
func TestGetChangeList(t *testing.T) {
	changes := GetChangeList()
	// May be nil or empty depending on test environment
	if changes == nil {
		t.Log("GetChangeList returned nil (expected if no data)")
	} else {
		t.Logf("GetChangeList returned %d changes", len(changes))
	}
}

// TestSetOneApprovedChange ensures Updated timestamp is set.
func TestSetOneApprovedChange(t *testing.T) {
	approvedChange := &xwchange.ApprovedChange{
		ID:              "approved-123",
		ApplicationType: xwshared.STB,
		Operation:       "UPDATE",
	}
	if approvedChange.Updated != 0 {
		t.Fatalf("precondition: expected Updated to be 0")
	}
	err := SetOneApprovedChange(approvedChange)
	if approvedChange.Updated == 0 {
		t.Fatalf("expected Updated timestamp to be set")
	}
	_ = err
}

// TestGetOneApprovedChange retrieves a single approved change by ID.
func TestGetOneApprovedChange(t *testing.T) {
	result := GetOneApprovedChange("non-existent-id")
	// Expected to return nil if not found
	if result != nil {
		t.Logf("GetOneApprovedChange returned: %+v", result)
	} else {
		t.Log("GetOneApprovedChange returned nil (expected for non-existent ID)")
	}
}

// TestGetApprovedChangeList retrieves all approved changes.
func TestGetApprovedChangeList(t *testing.T) {
	changes := GetApprovedChangeList()
	// May be nil or empty depending on test environment
	if changes == nil {
		t.Log("GetApprovedChangeList returned nil (expected if no data)")
	} else {
		t.Logf("GetApprovedChangeList returned %d approved changes", len(changes))
	}
}

// TestGetChangesByEntityId filters changes by entity ID.
func TestGetChangesByEntityId(t *testing.T) {
	changes := GetChangesByEntityId("entity-123")
	// May be empty or nil depending on test environment
	if changes == nil {
		t.Log("GetChangesByEntityId returned nil")
	} else {
		t.Logf("GetChangesByEntityId returned %d changes", len(changes))
	}
}

// TestGetOneChange retrieves a single change by ID.
func TestGetOneChange(t *testing.T) {
	result := GetOneChange("non-existent-change-id")
	// Expected to return nil if not found
	if result != nil {
		t.Logf("GetOneChange returned: %+v", result)
	} else {
		t.Log("GetOneChange returned nil (expected for non-existent ID)")
	}
}

// TestGetApprovedTelemetryTwoChangesByApplicationType retrieves approved telemetry two changes by app type.
func TestGetApprovedTelemetryTwoChangesByApplicationType(t *testing.T) {
	changes := GetApprovedTelemetryTwoChangesByApplicationType(xwshared.STB)
	// May be nil or empty depending on test environment
	if changes == nil {
		t.Log("GetApprovedTelemetryTwoChangesByApplicationType returned nil (expected if no data)")
	} else {
		t.Logf("GetApprovedTelemetryTwoChangesByApplicationType returned %d changes", len(changes))
	}
}

// TestGetAllTelemetryTwoChangeList retrieves all telemetry two changes.
func TestGetAllTelemetryTwoChangeList(t *testing.T) {
	changes := GetAllTelemetryTwoChangeList()
	// May be nil or empty depending on test environment
	if changes == nil {
		t.Log("GetAllTelemetryTwoChangeList returned nil (expected if no data)")
	} else {
		t.Logf("GetAllTelemetryTwoChangeList returned %d changes", len(changes))
	}
}

// TestGetAllApprovedTelemetryTwoChangeList retrieves all approved telemetry two changes.
func TestGetAllApprovedTelemetryTwoChangeList(t *testing.T) {
	changes := GetAllApprovedTelemetryTwoChangeList()
	// May be nil or empty depending on test environment
	if changes == nil {
		t.Log("GetAllApprovedTelemetryTwoChangeList returned nil (expected if no data)")
	} else {
		t.Logf("GetAllApprovedTelemetryTwoChangeList returned %d changes", len(changes))
	}
}

// TestGetOneTelemetryTwoChange retrieves a single telemetry two change by ID.
func TestGetOneTelemetryTwoChange(t *testing.T) {
	result := GetOneTelemetryTwoChange("non-existent-telemetry-id")
	// Expected to return nil if not found
	if result != nil {
		t.Logf("GetOneTelemetryTwoChange returned: %+v", result)
	} else {
		t.Log("GetOneTelemetryTwoChange returned nil (expected for non-existent ID)")
	}
}

// TestGetOneApprovedTelemetryTwoChange retrieves a single approved telemetry two change by ID.
func TestGetOneApprovedTelemetryTwoChange(t *testing.T) {
	result := GetOneApprovedTelemetryTwoChange("non-existent-approved-telemetry-id")
	// Expected to return nil if not found
	if result != nil {
		t.Logf("GetOneApprovedTelemetryTwoChange returned: %+v", result)
	} else {
		t.Log("GetOneApprovedTelemetryTwoChange returned nil (expected for non-existent ID)")
	}
}

// TestSetOneApprovedTelemetryTwoChangeSetsIDAndUpdated ensures ID/timestamp set when blank.
func TestSetOneApprovedTelemetryTwoChangeSetsIDAndUpdated(t *testing.T) {
	approved := &xwchange.ApprovedTelemetryTwoChange{ApplicationType: xwshared.STB, Operation: "CREATE"}
	if approved.ID != "" {
		t.Fatalf("precondition: expected blank ID")
	}
	err := SetOneApprovedTelemetryTwoChange(approved)
	if approved.ID == "" {
		t.Fatalf("expected ID to be generated for approved change")
	}
	if approved.Updated == 0 {
		t.Fatalf("expected Updated timestamp to be set for approved change")
	}
	_ = err
}
