package change

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	xchange "github.com/rdkcentral/xconfadmin/shared/change"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwchange "github.com/rdkcentral/xconfwebconfig/shared/change"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/stretchr/testify/assert"
)

const validTelemetryTwoJSON = "{\n    \"Description\":\"Test Json Data\",\n    \"Version\":\"0.1\",\n    \"Protocol\":\"HTTP\",\n    \"EncodingType\":\"JSON\",\n    \"ReportingInterval\":43200,\n    \"TimeReference\":\"0001-01-01T00:00:00Z\",\n    \"RootName\":\"someNewRootName\",\n    \"Parameter\": [ { \"type\": \"dataModel\", \"reference\": \"Profile.Name\"} ],\n    \"HTTP\": {\n        \"URL\":\"https://test.net\",\n        \"Compression\":\"None\",\n        \"Method\":\"POST\"\n    },\n    \"JSONEncoding\": {\n        \"ReportFormat\":\"NameValuePair\",\n        \"ReportTimestamp\": \"None\"\n    }\n}"

// helper to make a telemetry two profile
func makeT2Profile(name string) *logupload.TelemetryTwoProfile {
	p := &logupload.TelemetryTwoProfile{}
	p.ID = uuid.New().String()
	p.Name = name
	p.Jsonconfig = validTelemetryTwoJSON
	p.ApplicationType = "stb"
	return p
}

// helper to create change objects directly in store
func seedChange(t *testing.T, op xwchange.ChangeOperation, oldP, newP *logupload.TelemetryTwoProfile) *xwchange.TelemetryTwoChange {
	c := xchange.NewEmptyTelemetryTwoChange()
	c.ID = uuid.New().String()
	if oldP != nil {
		c.EntityID = oldP.ID
	} else if newP != nil {
		c.EntityID = newP.ID
	}
	c.EntityType = xchange.TelemetryTwoProfile
	c.OldEntity = oldP
	c.NewEntity = newP
	c.Operation = op
	c.ApplicationType = "stb"
	c.Author = "tester"
	if err := xchange.CreateOneTelemetryTwoChange(c); err != nil {
		t.Fatalf("seed change: %v", err)
	}
	return c
}

// local cleanup to avoid cross-package DeleteAllEntities dependency
func cleanupChangeTest() {
	tables := []string{
		ds.TABLE_XCONF_TELEMETRY_TWO_CHANGE,
		ds.TABLE_XCONF_APPROVED_TELEMETRY_TWO_CHANGE,
		ds.TABLE_TELEMETRY_TWO_PROFILES,
	}
	for _, tbl := range tables {
		list, _ := ds.GetCachedSimpleDao().GetAllAsList(tbl, 0)
		for _, inst := range list {
			// derive key by type assertion
			switch v := inst.(type) {
			case *logupload.TelemetryTwoProfile:
				ds.GetCachedSimpleDao().DeleteOne(tbl, v.ID)
			case *xwchange.TelemetryTwoChange:
				ds.GetCachedSimpleDao().DeleteOne(tbl, v.ID)
			case *xwchange.ApprovedTelemetryTwoChange:
				ds.GetCachedSimpleDao().DeleteOne(tbl, v.ID)
			}
		}
		ds.GetCachedSimpleDao().RefreshAll(tbl)
	}
}

// minimal request with context for auth mocking (permission functions read applicationType query param)
func makeRequest(method, url string) *http.Request {
	return httptest.NewRequest(method, url+"?applicationType=stb", nil)
}

func TestApproveTelemetryTwoChange_CreateFlow(t *testing.T) {
	cleanupChangeTest()
	p := makeT2Profile("createProf")
	c := seedChange(t, xchange.Create, nil, p)
	r := makeRequest("GET", "/x")
	approved, err := ApproveTelemetryTwoChange(r, c.ID)
	assert.NoError(t, err)
	assert.NotNil(t, approved)
	stored := logupload.GetOneTelemetryTwoProfile(p.ID)
	assert.NotNil(t, stored)
}

func TestApproveTelemetryTwoChange_UpdateFlow(t *testing.T) {
	cleanupChangeTest()
	orig := makeT2Profile("orig")
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, orig.ID, orig)
	updated, _ := orig.Clone()
	updated.Jsonconfig = validTelemetryTwoJSON // still valid; change Version text to simulate update
	updated.Name = "orig-upd"
	c := seedChange(t, xchange.Update, orig, updated)
	r := makeRequest("GET", "/x")
	approved, err := ApproveTelemetryTwoChange(r, c.ID)
	assert.NoError(t, err)
	assert.NotNil(t, approved)
	stored := logupload.GetOneTelemetryTwoProfile(orig.ID)
	assert.Equal(t, updated.Jsonconfig, stored.Jsonconfig)
}

func TestApproveTelemetryTwoChange_DeleteFlow(t *testing.T) {
	cleanupChangeTest()
	orig := makeT2Profile("del")
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, orig.ID, orig)
	c := seedChange(t, xchange.Delete, orig, nil)
	r := makeRequest("GET", "/x")
	approved, err := ApproveTelemetryTwoChange(r, c.ID)
	assert.NoError(t, err)
	assert.NotNil(t, approved)
	ds.GetCachedSimpleDao().RefreshAll(ds.TABLE_TELEMETRY_TWO_PROFILES)
	assert.Nil(t, logupload.GetOneTelemetryTwoProfile(orig.ID))
}

func TestApproveTelemetryTwoChange_NotFound(t *testing.T) {
	cleanupChangeTest()
	r := makeRequest("GET", "/x")
	approved, err := ApproveTelemetryTwoChange(r, uuid.New().String())
	assert.Nil(t, approved)
	assert.Error(t, err)
}

func TestApproveTelemetryTwoChanges_MixedBatch(t *testing.T) {
	cleanupChangeTest()
	// create
	p1 := makeT2Profile("p1")
	c1 := seedChange(t, xchange.Create, nil, p1)
	// update ok
	base := makeT2Profile("base")
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, base.ID, base)
	upd, _ := base.Clone()
	upd.Jsonconfig = validTelemetryTwoJSON
	upd.Name = "base-upd"
	c2 := seedChange(t, xchange.Update, base, upd)
	// delete missing entity -> will cause error when approving (entity not present?) we remove before approve
	toDelete := makeT2Profile("gone")
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, toDelete.ID, toDelete)
	c3 := seedChange(t, xchange.Delete, toDelete, nil)
	// corrupt store: remove entity so delete will fail (simulate conflict) by deleting manually so Delete telemetry profile returns not found -> error path
	ds.GetCachedSimpleDao().DeleteOne(ds.TABLE_TELEMETRY_TWO_PROFILES, toDelete.ID)

	r := makeRequest("GET", "/x")
	errs := ApproveTelemetryTwoChanges(r, []string{c1.ID, c2.ID, c3.ID})
	// expect one error for delete
	assert.Len(t, errs, 1)
	// profiles for c1 and c2 should exist and updated
	assert.NotNil(t, logupload.GetOneTelemetryTwoProfile(p1.ID))
	assert.Equal(t, upd.Jsonconfig, logupload.GetOneTelemetryTwoProfile(base.ID).Jsonconfig)
}

func TestRevertTelemetryTwoChange_CreateAndDeleteFlows(t *testing.T) {
	cleanupChangeTest()
	// create approval then revert (operation=create) should delete entity
	p := makeT2Profile("revCreate")
	createChange := seedChange(t, xchange.Create, nil, p)
	r := makeRequest("GET", "/x")
	approvedCreate, _ := ApproveTelemetryTwoChange(r, createChange.ID)
	resp := RevertTelemetryTwoChange(r, approvedCreate.ID)
	assert.Equal(t, http.StatusOK, resp.Status)
	assert.Nil(t, logupload.GetOneTelemetryTwoProfile(p.ID))

	// delete approval then revert (operation=delete) should recreate entity
	p2 := makeT2Profile("revDelete")
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, p2.ID, p2)
	delChange := seedChange(t, xchange.Delete, p2, nil)
	approvedDel, _ := ApproveTelemetryTwoChange(r, delChange.ID)
	resp2 := RevertTelemetryTwoChange(r, approvedDel.ID)
	assert.Equal(t, http.StatusOK, resp2.Status)
	assert.NotNil(t, logupload.GetOneTelemetryTwoProfile(p2.ID))
}

func TestRevertTelemetryTwoChanges_Batch(t *testing.T) {
	cleanupChangeTest()
	p1 := makeT2Profile("b1")
	c1 := seedChange(t, xchange.Create, nil, p1)
	p2 := makeT2Profile("b2")
	c2 := seedChange(t, xchange.Create, nil, p2)
	r := makeRequest("GET", "/x")
	a1, _ := ApproveTelemetryTwoChange(r, c1.ID)
	a2, _ := ApproveTelemetryTwoChange(r, c2.ID)
	errs := RevertTelemetryTwoChanges(r, []string{a1.ID, a2.ID})
	assert.Empty(t, errs)
	assert.Nil(t, logupload.GetOneTelemetryTwoProfile(p1.ID))
	assert.Nil(t, logupload.GetOneTelemetryTwoProfile(p2.ID))
}

func TestPagingAndGroupingHelpers(t *testing.T) {
	cleanupChangeTest()
	// create multiple changes
	for i := 0; i < 5; i++ {
		p := makeT2Profile("pg" + uuid.New().String())
		seedChange(t, xchange.Create, nil, p)
	}
	all := xchange.GetAllTelemetryTwoChangeList()
	pg := GeneratePageTelemetryTwoChanges(all, 1, 2)
	assert.Equal(t, 2, len(pg))
	groups := GroupTelemetryTwoChanges(all)
	assert.True(t, len(groups) >= 5)
}

func TestApplyUpdateTelemetryTwoChange_Merge(t *testing.T) {
	cleanupChangeTest()
	orig := makeT2Profile("merge")
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, orig.ID, orig)
	upd, _ := orig.Clone()
	upd.Jsonconfig = validTelemetryTwoJSON
	upd.Name = "merge2"
	change := seedChange(t, xchange.Update, orig, upd)
	// first merge (nil existing)
	mr, err := applyUpdateTelemetryTwoChange(nil, change)
	assert.NoError(t, err)
	assert.Equal(t, upd.Jsonconfig, mr.Jsonconfig)
	// second merge change name again
	upd2, _ := upd.Clone()
	upd2.Name = "merge3"
	change.NewEntity = upd2
	mr2, err2 := applyUpdateTelemetryTwoChange(mr, change)
	assert.NoError(t, err2)
	assert.Equal(t, "merge3", mr2.Name)
}
