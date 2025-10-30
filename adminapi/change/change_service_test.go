package change

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	xchange "github.com/rdkcentral/xconfadmin/shared/change"
	"github.com/rdkcentral/xconfwebconfig/shared"
	xwchange "github.com/rdkcentral/xconfwebconfig/shared/change"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

// helper builders
func buildPermTelemetryProfile(id, name, app string) *xwlogupload.PermanentTelemetryProfile {
	p := &xwlogupload.PermanentTelemetryProfile{}
	p.ID = id
	p.Name = name
	p.ApplicationType = app
	p.Type = xwlogupload.PermanentTelemetryProfileConst
	p.UploadProtocol = "https"
	p.UploadRepository = "https://example.com"
	p.TelemetryProfile = []xwlogupload.TelemetryElement{{Header: "BH", Content: "BC", Type: "BT", PollingFrequency: "30"}}
	return p
}

func buildChange(id string, op xwchange.ChangeOperation, oldEnt, newEnt *xwlogupload.PermanentTelemetryProfile, app string, author string) *xwchange.Change {
	c := xchange.NewEmptyChange()
	c.ID = id
	if oldEnt != nil {
		c.EntityID = oldEnt.ID
	} else if newEnt != nil {
		c.EntityID = newEnt.ID
	}
	c.EntityType = xwchange.TelemetryProfile
	c.ApplicationType = app
	c.Author = author
	c.Operation = op
	if oldEnt != nil {
		c.OldEntity = oldEnt
	}
	if newEnt != nil {
		c.NewEntity = newEnt
	}
	return c
}

// minimal http request for auth; tests that don't hit auth paths pass nil
func dummyRequest() *http.Request { r := httptest.NewRequest(http.MethodGet, "/", nil); return r }

func TestValidateChangeErrors(t *testing.T) {
	// empty
	if err := validateChange(nil); err == nil {
		t.Fatalf("expected error for nil change")
	}
	// blank id
	c := buildChange("", xwchange.Create, nil, buildPermTelemetryProfile("n1", "n1", shared.STB), shared.STB, "author")
	if err := validateChange(c); err == nil {
		t.Fatalf("expected error for blank id")
	}
	// missing author
	c.ID = "chg1"
	c.Author = ""
	if err := validateChange(c); err == nil {
		t.Fatalf("expected error for missing author")
	}
	// missing entity id
	c.Author = "auth"
	c.EntityID = ""
	if err := validateChange(c); err == nil {
		t.Fatalf("expected error for missing entity id")
	}
	// missing operation
	c.EntityID = c.NewEntity.ID
	c.Operation = ""
	if err := validateChange(c); err == nil {
		t.Fatalf("expected error for missing operation")
	}
	// empty new entity for create
	c.Operation = xwchange.Create
	c.NewEntity = &xwlogupload.PermanentTelemetryProfile{} // empty
	if err := validateChange(c); err == nil {
		t.Fatalf("expected error for empty new entity")
	}
	// restore valid new entity but empty old entity for delete
	c.Operation = xwchange.Delete
	c.NewEntity = nil
	c.OldEntity = &xwlogupload.PermanentTelemetryProfile{} // empty
	if err := validateChange(c); err == nil {
		t.Fatalf("expected error for empty old entity on delete")
	}
}

func TestValidateApprovedChange(t *testing.T) {
	newEnt := buildPermTelemetryProfile("p1", "p1", shared.STB)
	c := buildChange("c1", xwchange.Create, nil, newEnt, shared.STB, "author")
	c.ApprovedUser = ""
	if err := validateApprovedChange(c); err == nil {
		t.Fatalf("expected error for missing approved user")
	}
	c.ApprovedUser = "approver"
	if err := validateApprovedChange(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGroupChanges(t *testing.T) {
	p1 := buildPermTelemetryProfile("p1", "p1", shared.STB)
	p2 := buildPermTelemetryProfile("p2", "p2", shared.STB)
	c1 := buildChange("c1", xwchange.Create, nil, p1, shared.STB, "a1")
	c2 := buildChange("c2", xwchange.Update, p1, p1, shared.STB, "a2")
	c3 := buildChange("c3", xwchange.Create, nil, p2, shared.STB, "a3")
	grouped := GroupChanges([]*xwchange.Change{c1, c2, c3})
	if len(grouped) != 2 {
		t.Fatalf("expected 2 entity groups")
	}
	if len(grouped["p1"]) != 2 {
		t.Fatalf("expected p1 group size 2")
	}
}

func TestGroupApprovedChanges(t *testing.T) {
	p1 := buildPermTelemetryProfile("p1", "p1", shared.STB)
	c := buildChange("c1", xwchange.Create, nil, p1, shared.STB, "a1")
	c.ApprovedUser = "approver"
	ac := xwchange.ApprovedChange(*c)
	grouped := GroupApprovedChanges([]*xwchange.ApprovedChange{&ac})
	if len(grouped) != 1 {
		t.Fatalf("expected 1 group")
	}
}

func TestFindByContextForChanges(t *testing.T) {
	// create some pending changes directly in DB via CreateOneChange
	p1 := buildPermTelemetryProfile("p1", "telemetry-alpha", shared.STB)
	p2 := buildPermTelemetryProfile("p2", "telemetry-beta", shared.STB)
	c1 := buildChange("ch1", xwchange.Create, nil, p1, shared.STB, "alice")
	c2 := buildChange("ch2", xwchange.Create, nil, p2, shared.STB, "bob")
	if err := xchange.CreateOneChange(c1); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := xchange.CreateOneChange(c2); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// filter by author substring
	res := FindByContextForChanges(map[string]string{"author": "ali"})
	if len(res) != 1 || res[0].Author != "alice" {
		t.Fatalf("expected filter by author matched alice only")
	}
	// filter by profile name substring
	res = FindByContextForChanges(map[string]string{"entity": "beta"})
	if len(res) != 1 || res[0].NewEntity.Name != "telemetry-beta" {
		t.Fatalf("expected beta profile filter")
	}
}

func TestValidateAllChangesConflict(t *testing.T) {
	p := buildPermTelemetryProfile("pp", "pp", shared.STB)
	c1 := buildChange("dup1", xwchange.Create, nil, p, shared.STB, "alice")
	c2 := buildChange("dup2", xwchange.Create, nil, p, shared.STB, "alice")
	if err := xchange.CreateOneChange(c1); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := validateAllChanges(c2); err == nil {
		t.Fatalf("expected conflict error for duplicate change data")
	}
}

func TestBeforeDeleteErrors(t *testing.T) {
	if err := beforeDelete(""); err == nil {
		t.Fatalf("expected blank id error")
	}
	if err := beforeDelete("nope"); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestGetChangedEntityIds(t *testing.T) {
	p := buildPermTelemetryProfile("ent1", "ent1", shared.STB)
	c := buildChange("cidx", xwchange.Create, nil, p, shared.STB, "author")
	if err := xchange.CreateOneChange(c); err != nil {
		t.Fatalf("setup: %v", err)
	}
	ids := GetChangedEntityIds()
	if ids == nil || len(*ids) == 0 {
		t.Fatalf("expected at least one changed entity id")
	}
}

func TestApproveAndCancelSiblingChanges(t *testing.T) {
	// create entity and two pending changes (update + delete) same entity id
	p := buildPermTelemetryProfile("ap1", "ap1", shared.STB)
	cCreate := buildChange("cCreate", xwchange.Create, nil, p, shared.STB, "author")
	if err := xchange.CreateOneChange(cCreate); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// approve create should move to approved and remove pending
	_, err := Approve(dummyRequest(), cCreate.ID)
	if err != nil {
		t.Fatalf("approve error: %v", err)
	}
	still := xchange.GetOneChange(cCreate.ID)
	if still != nil {
		t.Fatalf("expected pending change removed after approve")
	}
	approved := xchange.GetOneApprovedChange(cCreate.ID)
	if approved == nil {
		t.Fatalf("expected approved change present")
	}
	// now create another pending change update on same entity; approving should cancel siblings (none here) but keep approved
	pUpdated := buildPermTelemetryProfile("ap1", "ap1-new", shared.STB)
	cUpdate := buildChange("cUpdate", xwchange.Update, p, pUpdated, shared.STB, "author")
	if err := xchange.CreateOneChange(cUpdate); err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err = Approve(dummyRequest(), cUpdate.ID)
	if err != nil {
		t.Fatalf("approve update: %v", err)
	}
	if xchange.GetOneChange(cUpdate.ID) != nil {
		t.Fatalf("expected update pending removed")
	}
}

func TestApproveChangesBatch(t *testing.T) {
	p1 := buildPermTelemetryProfile("bp1", "bp1", shared.STB)
	p2 := buildPermTelemetryProfile("bp2", "bp2", shared.STB)
	c1 := buildChange("bc1", xwchange.Create, nil, p1, shared.STB, "author")
	c2 := buildChange("bc2", xwchange.Create, nil, p2, shared.STB, "author")
	if err := xchange.CreateOneChange(c1); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := xchange.CreateOneChange(c2); err != nil {
		t.Fatalf("setup: %v", err)
	}
	ids := []string{c1.ID, c2.ID}
	m, err := ApproveChanges(dummyRequest(), &ids)
	if err != nil {
		t.Fatalf("batch approve error: %v", err)
	}
	if len(m) != 0 {
		t.Fatalf("expected no error messages")
	}
	if xchange.GetOneApprovedChange(c1.ID) == nil || xchange.GetOneApprovedChange(c2.ID) == nil {
		t.Fatalf("expected both approved")
	}
}

func TestRevertChangeCreate(t *testing.T) {
	// create and approve then revert a create
	p := buildPermTelemetryProfile("rp1", "rp1", shared.STB)
	c := buildChange("rc1", xwchange.Create, nil, p, shared.STB, "author")
	if err := xchange.CreateOneChange(c); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := Approve(dummyRequest(), c.ID); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if err := Revert(dummyRequest(), c.ID); err != nil {
		t.Fatalf("revert: %v", err)
	}
	if xchange.GetOneApprovedChange(c.ID) != nil {
		t.Fatalf("expected approved change deleted after revert")
	}
}

func TestFindByContextForApprovedChanges(t *testing.T) {
	// create several approved changes
	targets := []struct{ id, name, author string }{
		{"apf1", "name-filter", "authorOne"},
		{"apf2", "other-name", "authorTwo"},
	}
	for _, tg := range targets {
		p := buildPermTelemetryProfile(tg.id, tg.name, shared.STB)
		c := buildChange("chg-"+tg.id, xwchange.Create, nil, p, shared.STB, tg.author)
		if err := xchange.CreateOneChange(c); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if _, err := Approve(dummyRequest(), c.ID); err != nil {
			t.Fatalf("approve: %v", err)
		}
	}
	// filter should use keys AUTHOR and PROFILE_NAME to fetch single match
	res := FindByContextForApprovedChanges(dummyRequest(), map[string]string{"author": "authorOne", "profileName": "name-filter"})
	if len(res) != 1 {
		t.Fatalf("expected one approved filtered result, got %d", len(res))
	}
	if res[0].NewEntity.Name != "name-filter" {
		t.Fatalf("unexpected entity name %s", res[0].NewEntity.Name)
	}
}

func TestSaveToApprovedAndCleanUpChange(t *testing.T) {
	p := buildPermTelemetryProfile("sacc1", "sacc1", shared.STB)
	c := buildChange("saccCh", xwchange.Create, nil, p, shared.STB, "auth")
	if err := xchange.CreateOneChange(c); err != nil {
		t.Fatalf("setup: %v", err)
	}
	ac, err := SaveToApprovedAndCleanUpChange(dummyRequest(), c)
	if err != nil {
		t.Fatalf("save approved: %v", err)
	}
	if ac == nil || xchange.GetOneChange(c.ID) != nil || xchange.GetOneApprovedChange(c.ID) == nil {
		t.Fatalf("expected cleanup & approved presence")
	}
}

func TestApproveNotFound(t *testing.T) {
	if _, err := Approve(dummyRequest(), "no-such"); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestRevertErrors(t *testing.T) {
	if err := Revert(dummyRequest(), ""); err == nil {
		t.Fatalf("expected blank id error")
	}
	if err := Revert(dummyRequest(), "no-id"); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestApproveChangesErrors(t *testing.T) {
	ids := []string{"missing"}
	if _, err := ApproveChanges(dummyRequest(), &ids); err == nil {
		t.Fatalf("expected missing change error")
	}
}

func TestRevertChangesErrors(t *testing.T) {
	ids := []string{"missing"}
	if _, err := RevertChanges(dummyRequest(), &ids); err == nil {
		t.Fatalf("expected missing approved change error")
	}
}

func TestJSONMarshallingApprovedChange(t *testing.T) {
	p := buildPermTelemetryProfile("json1", "json1", shared.STB)
	c := buildChange("jsonchg", xwchange.Create, nil, p, shared.STB, "author")
	c.ApprovedUser = "approver"
	ac := xwchange.ApprovedChange(*c)
	b, err := json.Marshal(ac)
	if err != nil || len(b) == 0 {
		t.Fatalf("expected json marshal success")
	}
}

// ============================================================================
// Additional Coverage Tests for change_service.go
// ============================================================================

func TestGetApprovedAll_EmptyResult(t *testing.T) {
	// Clean all approved changes first
	approvedChanges := xchange.GetApprovedChangeList()
	for _, ac := range approvedChanges {
		xchange.DeleteOneApprovedChange(ac.ID)
	}

	r := httptest.NewRequest(http.MethodGet, "/?applicationType=stb", nil)
	result, err := GetApprovedAll(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
}

func TestGetApprovedAll_WithResults(t *testing.T) {
	defer func() {
		approvedChanges := xchange.GetApprovedChangeList()
		for _, ac := range approvedChanges {
			xchange.DeleteOneApprovedChange(ac.ID)
		}
	}()

	// Create test approved changes
	p1 := buildPermTelemetryProfile("gaa1", "gaa1", shared.STB)
	c1 := buildChange("gaac1", xwchange.Create, nil, p1, shared.STB, "author1")
	ac1 := xwchange.ApprovedChange(*c1)
	if err := xchange.SetOneApprovedChange(&ac1); err != nil {
		t.Fatalf("failed to create approved change: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/?applicationType=stb", nil)
	result, err := GetApprovedAll(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatalf("expected results")
	}
}

func TestFindByContextForChanges_EmptyContext(t *testing.T) {
	context := make(map[string]string)
	result := FindByContextForChanges(context)
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
}

func TestFindByContextForChanges_WithApplicationType(t *testing.T) {
	defer func() {
		changes := xchange.GetChangeList()
		for _, c := range changes {
			xchange.DeleteOneChange(c.ID)
		}
	}()

	p := buildPermTelemetryProfile("fbc1", "fbc1", shared.STB)
	c := buildChange("fbcc1", xwchange.Create, nil, p, shared.STB, "testauthor")
	if err := xchange.CreateOneChange(c); err != nil {
		t.Fatalf("failed to create change: %v", err)
	}

	context := map[string]string{"applicationType": shared.STB}
	result := FindByContextForChanges(context)
	if len(result) == 0 {
		t.Fatalf("expected results")
	}
}

func TestFindByContextForApprovedChanges_EmptyContext(t *testing.T) {
	context := make(map[string]string)
	result := FindByContextForApprovedChanges(dummyRequest(), context)
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
}

func TestFindByContextForApprovedChanges_WithApplicationType(t *testing.T) {
	defer func() {
		approvedChanges := xchange.GetApprovedChangeList()
		for _, ac := range approvedChanges {
			xchange.DeleteOneApprovedChange(ac.ID)
		}
	}()

	p := buildPermTelemetryProfile("fbac1", "fbac1", shared.STB)
	c := buildChange("fbacc1", xwchange.Create, nil, p, shared.STB, "testauthor")
	ac := xwchange.ApprovedChange(*c)
	if err := xchange.SetOneApprovedChange(&ac); err != nil {
		t.Fatalf("failed to create approved change: %v", err)
	}

	context := map[string]string{"applicationType": shared.STB}
	result := FindByContextForApprovedChanges(dummyRequest(), context)
	if len(result) == 0 {
		t.Fatalf("expected results")
	}
}

func TestGetChangesByEntityIds_EmptyList(t *testing.T) {
	ids := []string{}
	result, err := GetChangesByEntityIds(&ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
}

func TestGetChangesByEntityIds_NonExistent(t *testing.T) {
	ids := []string{"nonexistent1", "nonexistent2"}
	_, err := GetChangesByEntityIds(&ids)
	if err == nil {
		t.Fatalf("expected error for nonexistent entities")
	}
}

func TestCancelApprovedChangesByEntityId_EmptyList(t *testing.T) {
	err := CancelApprovedChangesByEntityId(dummyRequest(), []string{}, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCancelApprovedChangesByEntityId_NonExistent(t *testing.T) {
	err := CancelApprovedChangesByEntityId(dummyRequest(), []string{"nonexistent"}, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRevertChanges_EmptyList(t *testing.T) {
	ids := []string{}
	result, err := RevertChanges(dummyRequest(), &ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
}

func TestApproveChanges_EmptyList(t *testing.T) {
	ids := []string{}
	result, err := ApproveChanges(dummyRequest(), &ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
}

// Tests for low coverage functions
func TestLogAndCollectChangeException(t *testing.T) {
	change := buildChange("change1", xwchange.Create, nil, buildPermTelemetryProfile("p1", "P1", "stb"), "stb", "admin")
	errorMessages := make(map[string]string)
	testErr := http.ErrAbortHandler

	logAndCollectChangeException(change, testErr, errorMessages)

	if _, exists := errorMessages[change.ID]; !exists {
		t.Fatalf("error message should have been collected")
	}
	if errorMessages[change.ID] == "" {
		t.Fatalf("error message should not be empty")
	}
}

func TestBeforeSavingChange_MissingID(t *testing.T) {
	change := buildChange("", xwchange.Create, nil, buildPermTelemetryProfile("p1", "P1", "stb"), "stb", "admin")

	err := beforeSavingChange(dummyRequest(), change)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if change.ID == "" {
		t.Fatalf("ID should have been generated")
	}
}

func TestBeforeSavingApprovedChange_Validation(t *testing.T) {
	change := buildChange("c1", xwchange.Create, nil, buildPermTelemetryProfile("p1", "P1", "stb"), "stb", "admin")
	change.ApprovedUser = "approver"

	err := beforeSavingApprovedChange(dummyRequest(), change)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCancelApprovedChangesByEntityId_EmptyEntityList(t *testing.T) {
	entityIds := []string{}
	excludeIds := []string{}

	err := CancelApprovedChangesByEntityId(dummyRequest(), entityIds, excludeIds)
	if err != nil {
		t.Fatalf("unexpected error for empty list: %v", err)
	}
}

func TestRevertChanges_NonExistent(t *testing.T) {
	ids := []string{"nonexistent1"}
	_, err := RevertChanges(dummyRequest(), &ids)
	if err == nil {
		t.Fatalf("expected error for non-existent approved change")
	}
}

func TestApproveChanges_NonExistent(t *testing.T) {
	ids := []string{"nonexistent1"}
	_, err := ApproveChanges(dummyRequest(), &ids)
	if err == nil {
		t.Fatalf("expected error for non-existent change")
	}
}

func TestGroupApprovedChange_SingleChange(t *testing.T) {
	p1 := buildPermTelemetryProfile("p1", "P1", "stb")
	c1 := buildChange("c1", xwchange.Create, nil, p1, "stb", "admin")
	c1.ApprovedUser = "approver"
	ac1 := xwchange.ApprovedChange(*c1)

	result := make(map[string][]*xwchange.ApprovedChange)
	groupApprovedChange(&ac1, result)

	if len(result) != 1 || len(result["p1"]) != 1 {
		t.Fatalf("expected single group with single change")
	}
}

func TestGetChangeIds_MultipleChanges(t *testing.T) {
	p1 := buildPermTelemetryProfile("p1", "P1", "stb")
	p2 := buildPermTelemetryProfile("p2", "P2", "stb")
	c1 := buildChange("c1", xwchange.Create, nil, p1, "stb", "admin")
	c2 := buildChange("c2", xwchange.Create, nil, p2, "stb", "admin")

	changes := []*xwchange.Change{c1, c2}
	entityIds := getChangeIds(changes)

	if len(entityIds) != 2 {
		t.Fatalf("expected 2 entity IDs, got %d", len(entityIds))
	}
	if entityIds[0] != "p1" || entityIds[1] != "p2" {
		t.Fatalf("unexpected entity IDs: %v", entityIds)
	}
}
