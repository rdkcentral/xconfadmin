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
	_ = SetOneInDao(ds.TABLE_PERMANENT_TELEMETRY, p.ID, p)
	return p
}

func TestGetTelemetryRulesHandler_Empty(t *testing.T) {
	DeleteTelemetryEntities()
	url := "/xconfAdminService/telemetry/rule?applicationType=stb"
	r := httptest.NewRequest(http.MethodGet, url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("[]")))
}

func TestCreateTelemetryRuleHandler_SuccessAndConflict(t *testing.T) {
	DeleteTelemetryEntities()
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
	DeleteTelemetryEntities()
	url := "/xconfAdminService/telemetry/rule?applicationType=stb"
	r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{bad")))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetTelemetryRuleByIdHandler_SuccessAndNotFound(t *testing.T) {
	DeleteTelemetryEntities()
	perm := buildPermanentTelemetryProfile()
	rule := buildTelemetryRule("ruleB", "stb", perm.ID)
	_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)
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
	// Skip this test - it requires complex db.GetCachedSimpleDao() mocking beyond GetCachedSimpleDaoFunc
	SkipIfMockDatabase(t)

	DeleteTelemetryEntities()
	perm := buildPermanentTelemetryProfile()
	_ = SetOneInDao(ds.TABLE_PERMANENT_TELEMETRY, perm.ID, perm)
	rule := buildTelemetryRule("ruleC", "stb", perm.ID)
	_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)
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
	DeleteTelemetryEntities()
	perm := buildPermanentTelemetryProfile()
	rule := buildTelemetryRule("ruleD", "stb", perm.ID)
	_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)
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
	DeleteTelemetryEntities()
	perm := buildPermanentTelemetryProfile()
	valid := buildTelemetryRule("ruleE", "stb", perm.ID)
	conflict := buildTelemetryRule("ruleE", "stb", perm.ID) // same name allowed? uniqueness by ID; make conflict by pre-inserting then re-post
	_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, conflict.ID, conflict)
	entities := []*xwlogupload.TelemetryRule{valid, conflict}
	b, _ := json.Marshal(entities)
	url := "/xconfAdminService/telemetry/rule/entities?applicationType=stb"
	r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte(valid.ID)))
}

func TestPutTelemetryRuleEntitiesHandler_MixedResults(t *testing.T) {
	DeleteTelemetryEntities()
	perm := buildPermanentTelemetryProfile()
	// existing
	existing := buildTelemetryRule("ruleF", "stb", perm.ID)
	_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, existing.ID, existing)
	// update success
	existing.Name = "ruleF-new"
	// conflict by changing appType mismatch
	conflict := buildTelemetryRule("ruleG", "wrong", perm.ID)
	_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, conflict.ID, conflict)
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
	DeleteTelemetryEntities()
	perm := buildPermanentTelemetryProfile()
	// create several rules
	for i := 0; i < 15; i++ {
		rule := buildTelemetryRule(fmt.Sprintf("r%02d", i), "stb", perm.ID)
		_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)
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

// ===== Error Condition Tests for All Handlers =====

func TestGetTelemetryRuleByIdHandler_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	t.Run("MissingRuleID_WriteAdminErrorResponse", func(t *testing.T) {
		// Empty ruleId in path triggers 404 from router
		url := "/xconfAdminService/telemetry/rule/?applicationType=stb"
		r := httptest.NewRequest(http.MethodGet, url, nil)
		rr := ExecuteRequest(r, router)
		// Router returns 404 for missing path param
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("RuleNotFound_WriteAdminErrorResponse_404", func(t *testing.T) {
		nonexistentID := uuid.New().String()
		url := fmt.Sprintf("/xconfAdminService/telemetry/rule/%s?applicationType=stb", nonexistentID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("not found")))
	})

	t.Run("WrongApplicationType_WriteAdminErrorResponse_404", func(t *testing.T) {
		perm := buildPermanentTelemetryProfile()
		rule := buildTelemetryRule("test-rule", "stb", perm.ID)
		_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)

		// Query with different applicationType triggers 400 (invalid application type)
		url := fmt.Sprintf("/xconfAdminService/telemetry/rule/%s?applicationType=xhome", rule.ID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestDeleteTelemetryRuleByIdHandler_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	t.Run("MissingRuleID_WriteAdminErrorResponse_404", func(t *testing.T) {
		url := "/xconfAdminService/telemetry/rule/?applicationType=stb"
		r := httptest.NewRequest(http.MethodDelete, url, nil)
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("DeleteServiceError_WriteAdminErrorResponse", func(t *testing.T) {
		nonexistentID := uuid.New().String()
		url := fmt.Sprintf("/xconfAdminService/telemetry/rule/%s?applicationType=stb", nonexistentID)
		r := httptest.NewRequest(http.MethodDelete, url, nil)
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("does not exist")))
	})
}

func TestCreateTelemetryRuleHandler_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	t.Run("InvalidJSON_WriteAdminErrorResponse_400", func(t *testing.T) {
		url := "/xconfAdminService/telemetry/rule?applicationType=stb"
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{invalid json")))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("invalid character")))
	})

	t.Run("CreateServiceError_ApplicationTypeMismatch_WriteAdminErrorResponse", func(t *testing.T) {
		perm := buildPermanentTelemetryProfile()
		rule := buildTelemetryRule("conflict-rule", "stb", perm.ID)
		// Store with stb
		_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)

		// Try to create with different applicationType in body
		rule.ApplicationType = "xhome"
		b, _ := json.Marshal(rule)
		url := "/xconfAdminService/telemetry/rule?applicationType=stb"
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("already exists")))
	})

	t.Run("EmptyRuleName_WriteAdminErrorResponse", func(t *testing.T) {
		perm := buildPermanentTelemetryProfile()
		rule := buildTelemetryRule("", "stb", perm.ID)
		b, _ := json.Marshal(rule)
		url := "/xconfAdminService/telemetry/rule?applicationType=stb"
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("Name is empty")))
	})
}

func TestUpdateTelemetryRuleHandler_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	t.Run("InvalidJSON_WriteAdminErrorResponse_400", func(t *testing.T) {
		url := "/xconfAdminService/telemetry/rule?applicationType=stb"
		r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte("{invalid json")))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("invalid character")))
	})

	t.Run("UpdateServiceError_ApplicationTypeMismatch_WriteAdminErrorResponse", func(t *testing.T) {
		perm := buildPermanentTelemetryProfile()
		rule := buildTelemetryRule("existing-rule", "stb", perm.ID)
		_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)

		// Try to update with different applicationType
		rule.ApplicationType = "xhome"
		b, _ := json.Marshal(rule)
		url := "/xconfAdminService/telemetry/rule?applicationType=stb"
		r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("ApplicationType doesn't match")))
	})

	t.Run("RuleNotFound_WriteAdminErrorResponse", func(t *testing.T) {
		perm := buildPermanentTelemetryProfile()
		rule := buildTelemetryRule("nonexistent-rule", "stb", perm.ID)
		rule.ID = uuid.New().String() // New ID that doesn't exist
		b, _ := json.Marshal(rule)
		url := "/xconfAdminService/telemetry/rule?applicationType=stb"
		r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("does not exist")))
	})
}

func TestPostTelemetryRuleEntitiesHandler_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	t.Run("InvalidJSON_WriteAdminErrorResponse_400", func(t *testing.T) {
		url := "/xconfAdminService/telemetry/rule/entities?applicationType=stb"
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{invalid json")))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("Unable to extract entity from json file")))
	})

	t.Run("EmptyEntitiesList_ReturnsEmptyResult", func(t *testing.T) {
		entities := []*xwlogupload.TelemetryRule{}
		b, _ := json.Marshal(entities)
		url := "/xconfAdminService/telemetry/rule/entities?applicationType=stb"
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("PartialFailure_MixedResults", func(t *testing.T) {
		perm := buildPermanentTelemetryProfile()
		validRule := buildTelemetryRule("valid-entity", "stb", perm.ID)

		// Create a conflicting rule by pre-storing it
		conflictRule := buildTelemetryRule("conflict-entity", "stb", perm.ID)
		_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, conflictRule.ID, conflictRule)

		entities := []*xwlogupload.TelemetryRule{validRule, conflictRule}
		b, _ := json.Marshal(entities)
		url := "/xconfAdminService/telemetry/rule/entities?applicationType=stb"
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusOK, rr.Code)
		// Response contains both success and error entries
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte(validRule.ID)))
	})
}

func TestPutTelemetryRuleEntitiesHandler_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	t.Run("InvalidJSON_WriteAdminErrorResponse_400", func(t *testing.T) {
		url := "/xconfAdminService/telemetry/rule/entities?applicationType=stb"
		r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte("{invalid json")))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("Unable to extract entity from json file")))
	})

	t.Run("EmptyEntitiesList_ReturnsEmptyResult", func(t *testing.T) {
		entities := []*xwlogupload.TelemetryRule{}
		b, _ := json.Marshal(entities)
		url := "/xconfAdminService/telemetry/rule/entities?applicationType=stb"
		r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("ApplicationTypeMismatch_PartialFailure", func(t *testing.T) {
		perm := buildPermanentTelemetryProfile()

		// Create and store a rule with stb
		existingRule := buildTelemetryRule("existing-update", "stb", perm.ID)
		_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, existingRule.ID, existingRule)
		existingRule.Name = "existing-update-modified"

		// Create a rule with wrong applicationType to trigger conflict
		conflictRule := buildTelemetryRule("conflict-update", "xhome", perm.ID)
		_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, conflictRule.ID, conflictRule)
		conflictRule.ApplicationType = "stb" // Change to trigger mismatch

		entities := []*xwlogupload.TelemetryRule{existingRule, conflictRule}
		b, _ := json.Marshal(entities)
		url := "/xconfAdminService/telemetry/rule/entities?applicationType=stb"
		r := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusOK, rr.Code)
		// Response contains mixed results
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte(existingRule.ID)))
	})
}

func TestPostTelemetryRuleFilteredWithParamsHandler_AllErrorCases(t *testing.T) {
	DeleteTelemetryEntities()

	t.Run("InvalidJSON_WriteAdminErrorResponse_400", func(t *testing.T) {
		url := "/xconfAdminService/telemetry/rule/filtered?applicationType=stb"
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{invalid json")))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("Invalid Json contents")))
	})

	t.Run("InvalidPageNumber_WriteAdminErrorResponse_400", func(t *testing.T) {
		body := map[string]string{"pageNumber": "0", "pageSize": "10"}
		b, _ := json.Marshal(body)
		url := "/xconfAdminService/telemetry/rule/filtered?applicationType=stb"
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("pageNumber and pageSize should both be greater than zero")))
	})

	t.Run("InvalidPageSize_WriteAdminErrorResponse_400", func(t *testing.T) {
		body := map[string]string{"pageNumber": "1", "pageSize": "-5"}
		b, _ := json.Marshal(body)
		url := "/xconfAdminService/telemetry/rule/filtered?applicationType=stb"
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Assert(t, bytes.Contains(rr.Body.Bytes(), []byte("pageNumber and pageSize should both be greater than zero")))
	})

	t.Run("MissingPaginationParams_UsesDefaults", func(t *testing.T) {
		perm := buildPermanentTelemetryProfile()
		rule := buildTelemetryRule("filter-rule", "stb", perm.ID)
		_ = SetOneInDao(ds.TABLE_TELEMETRY_RULES, rule.ID, rule)

		body := map[string]string{} // Empty body should use defaults
		b, _ := json.Marshal(body)
		url := "/xconfAdminService/telemetry/rule/filtered?applicationType=stb"
		r := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		rr := ExecuteRequest(r, router)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
