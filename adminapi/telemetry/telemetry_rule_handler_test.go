package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"gotest.tools/assert"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	xwlogupload "github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

// helper to build a minimal TelemetryRule with rule content
func buildTelemetryRule(name string, appType string, profileId string) *xwlogupload.TelemetryRule {
	r := &xwlogupload.TelemetryRule{}
	r.ID = uuid.New().String()
	r.Name = name
	r.ApplicationType = appType
	r.BoundTelemetryID = profileId
	// Provide a minimal valid rule with single condition: model IS TESTMODEL
	cond := re.NewCondition(coreef.RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg("TESTMODEL"))
	r.Rule = re.Rule{Condition: cond}
	return r
}

func buildPermanentTelemetryProfile() *xwlogupload.PermanentTelemetryProfile {
	p := &xwlogupload.PermanentTelemetryProfile{}
	p.ID = uuid.New().String()
	p.Name = "perm-profile" + uuid.New().String()[0:8]
	p.ApplicationType = "stb"
	p.TelemetryProfile = []xwlogupload.TelemetryElement{{
		ID:               uuid.New().String(),
		Header:           "hdr",
		Content:          "content",
		Type:             "type",
		PollingFrequency: "30",
		Component:        "comp",
	}}
	_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)
	return p
}

func TestGetTelemetryRulesHandler_Empty(t *testing.T) {
	DeleteAllEntities()
	url := "/xconfAdminService/telemetry/rule?applicationType=stb"
	r := httptest.NewRequest(http.MethodGet, url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("[]")))
}

func TestCreateTelemetryRuleHandler_SuccessAndConflict(t *testing.T) {
	DeleteAllEntities()
	perm := buildPermanentTelemetryProfile()
	// success create
	rule := buildTelemetryRule("ruleA", "stb", perm.ID)
	b, _ := json.Marshal(rule)
	url := "/xconfAdminService/telemetry/rule?applicationType=stb"
	r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)
	// conflict: reuse same ID with different applicationType in body triggering ApplicationType mismatch
	rule.ApplicationType = "wrong"
	b, _ = json.Marshal(rule)
	r = httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestCreateTelemetryRuleHandler_InvalidJSON(t *testing.T) {
	DeleteAllEntities()
	url := "/xconfAdminService/telemetry/rule?applicationType=stb"
	r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{bad")))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetTelemetryRuleByIdHandler_SuccessAndNotFound(t *testing.T) {
	DeleteAllEntities()
	perm := buildPermanentTelemetryProfile()
	rule := buildTelemetryRule("ruleB", "stb", perm.ID)
	_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)
	url := fmt.Sprintf("/xconfAdminService/telemetry/rule/%s?applicationType=stb", rule.ID)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	// not found
	url = fmt.Sprintf("/xconfAdminService/telemetry/rule/%s?applicationType=stb", uuid.New().String())
	r = httptest.NewRequest(http.MethodGet, url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateTelemetryRuleHandler_SuccessAndConflict(t *testing.T) {
	DeleteAllEntities()
	perm := buildPermanentTelemetryProfile()
	rule := buildTelemetryRule("ruleC", "stb", perm.ID)
	_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)
	// success update
	rule.Name = "ruleC-updated"
	b, _ := json.Marshal(rule)
	url := "/xconfAdminService/telemetry/rule?applicationType=stb"
	r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	// conflict: change ApplicationType mismatch with stored value
	rule.ApplicationType = "wrong"
	b, _ = json.Marshal(rule)
	r = httptest.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestDeleteTelemetryRuleHandler_SuccessAndNotFound(t *testing.T) {
	DeleteAllEntities()
	perm := buildPermanentTelemetryProfile()
	rule := buildTelemetryRule("ruleD", "stb", perm.ID)
	_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)
	url := fmt.Sprintf("/xconfAdminService/telemetry/rule/%s?applicationType=stb", rule.ID)
	r := httptest.NewRequest(http.MethodDelete, url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusNoContent, rr.Code)
	// not found
	url = fmt.Sprintf("/xconfAdminService/telemetry/rule/%s?applicationType=stb", uuid.New().String())
	r = httptest.NewRequest(http.MethodDelete, url, nil)
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestPostTelemetryRuleEntitiesHandler_MixedResults(t *testing.T) {
	DeleteAllEntities()
	perm := buildPermanentTelemetryProfile()
	valid := buildTelemetryRule("ruleE", "stb", perm.ID)
	conflict := buildTelemetryRule("ruleE", "stb", perm.ID) // same name allowed? uniqueness by ID; make conflict by pre-inserting then re-post
	_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_RULES, conflict.ID, conflict)
	entities := []*xwlogupload.TelemetryRule{valid, conflict}
	b, _ := json.Marshal(entities)
	url := "/xconfAdminService/telemetry/rule/entities?applicationType=stb"
	r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte(valid.ID)))
}

func TestPutTelemetryRuleEntitiesHandler_MixedResults(t *testing.T) {
	DeleteAllEntities()
	perm := buildPermanentTelemetryProfile()
	// existing
	existing := buildTelemetryRule("ruleF", "stb", perm.ID)
	_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_RULES, existing.ID, existing)
	// update success
	existing.Name = "ruleF-new"
	// conflict by changing appType mismatch
	conflict := buildTelemetryRule("ruleG", "wrong", perm.ID)
	_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_RULES, conflict.ID, conflict)
	conflict.ApplicationType = "stb" // will not conflict if mismatch? Need mismatch with stored value: stored wrong, send stb -> conflict
	entities := []*xwlogupload.TelemetryRule{existing, conflict}
	b, _ := json.Marshal(entities)
	url := "/xconfAdminService/telemetry/rule/entities?applicationType=stb"
	r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte(existing.ID)))
}

func TestPostTelemetryRuleFilteredWithParamsHandler_PagingAndFilters(t *testing.T) {
	DeleteAllEntities()
	perm := buildPermanentTelemetryProfile()
	// create several rules
	for i := 0; i < 15; i++ {
		rule := buildTelemetryRule(fmt.Sprintf("r%02d", i), "stb", perm.ID)
		_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)
	}
	// page 2 size 5
	body := map[string]string{"pageNumber": "2", "pageSize": "5"}
	b, _ := json.Marshal(body)
	url := "/xconfAdminService/telemetry/rule/filtered?applicationType=stb"
	r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	// ensure contains r05 or r06 in page 2 results
	assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("r05")) || bytes.Contains(rr.Body.Bytes(), []byte("r06")))
	// invalid json
	r = httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{bad")))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// invalid paging params
	body = map[string]string{"pageNumber": "0", "pageSize": "5"}
	b, _ = json.Marshal(body)
	r = httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rr = ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
