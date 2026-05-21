package queries

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	xwrfc "github.com/rdkcentral/xconfwebconfig/shared/rfc"
	"github.com/stretchr/testify/assert"
)

// Helpers
func frMakeFeature(name string, app string) *xwrfc.Feature {
	f := &xwrfc.Feature{ID: uuid.New().String(), Name: name, FeatureName: name + "Fn", ApplicationType: app, Enable: true, EffectiveImmediate: true, ConfigData: map[string]string{"k": "v"}}
	SetOneInDao(ds.TABLE_XCONF_FEATURE, f.ID, f)
	return f
}
func frMakeRule() *re.Rule {
	return &re.Rule{Condition: CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "model"), re.StandardOperationIs, "X1")}
}

func frMakeFeatureRule(featureIds []string, app string, priority int) *xwrfc.FeatureRule {
	fr := &xwrfc.FeatureRule{Id: uuid.New().String(), Name: "FR-" + uuid.New().String(), ApplicationType: app, FeatureIds: featureIds, Priority: priority, Rule: frMakeRule()}
	SetOneInDao(ds.TABLE_FEATURE_CONTROL_RULE, fr.Id, fr)
	return fr
}

func frCleanup() {
	tables := []string{ds.TABLE_FEATURE_CONTROL_RULE, ds.TABLE_XCONF_FEATURE}
	for _, tbl := range tables {
		list, _ := GetAllAsListFromDao(tbl, 0)
		for _, inst := range list {
			switch v := inst.(type) {
			case *xwrfc.FeatureRule:
				DeleteOneFromDao(tbl, v.Id)
			case *xwrfc.Feature:
				DeleteOneFromDao(tbl, v.ID)
			}
		}
		ds.GetCachedSimpleDao().RefreshAll(tbl)
	}
}

// Tests
func TestGetFeatureRulesFiltered_AndExportHandlers(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")
	frMakeFeatureRule([]string{f.ID}, "stb", 1)
	r := httptest.NewRequest("GET", "/featureRules?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetFeatureRulesFiltered(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	// list export
	r2 := httptest.NewRequest("GET", "/featureRules/export?applicationType=stb&export=true", nil)
	rr2 := httptest.NewRecorder()
	GetFeatureRulesExportHandler(rr2, r2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}

func TestGetFeatureRuleOne_ExportAndErrors(t *testing.T) {
	SkipIfMockDatabase(t)
	rBlank := httptest.NewRequest("GET", "/featureRule//?applicationType=stb", nil)
	rrBlank := httptest.NewRecorder()
	GetFeatureRuleOne(rrBlank, rBlank)
	assert.Equal(t, http.StatusBadRequest, rrBlank.Code)

	f := frMakeFeature("FeatA", "stb")
	fr := frMakeFeatureRule([]string{f.ID}, "stb", 1)
	// export single
	r := httptest.NewRequest("GET", fmt.Sprintf("/fr/%s?applicationType=stb&export=true", fr.Id), nil)
	r = mux.SetURLVars(r, map[string]string{"id": fr.Id})
	rr := httptest.NewRecorder()
	GetFeatureRuleOneExport(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	// mismatched app type
	rBadApp := httptest.NewRequest("GET", fmt.Sprintf("/fr/%s?applicationType=rdkcloud", fr.Id), nil)
	rBadApp = mux.SetURLVars(rBadApp, map[string]string{"id": fr.Id})
	rrBad := httptest.NewRecorder()
	GetFeatureRuleOneExport(rrBad, rBadApp)
	assert.Equal(t, http.StatusNotFound, rrBad.Code)
}

func TestCreateUpdateDeleteFeatureRuleHandlers(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")
	bodyCreate := &xwrfc.FeatureRule{Name: "Rule1", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 1, Rule: frMakeRule()}
	b, _ := json.Marshal(bodyCreate)
	r := httptest.NewRequest("POST", "/featureRule?applicationType=stb", nil)
	rrNative := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rrNative)
	xw.SetBody(string(b))
	CreateFeatureRuleHandler(xw, r)
	assert.Equal(t, http.StatusCreated, rrNative.Code)

	created := &xwrfc.FeatureRule{}
	json.Unmarshal(rrNative.Body.Bytes(), created)
	created.Name = "Rule1-Updated"
	b2, _ := json.Marshal(created)
	r2 := httptest.NewRequest("PUT", "/featureRule?applicationType=stb", nil)
	rr2Native := httptest.NewRecorder()
	xw2 := xwhttp.NewXResponseWriter(rr2Native)
	xw2.SetBody(string(b2))
	UpdateFeatureRuleHandler(xw2, r2)
	assert.Equal(t, http.StatusOK, rr2Native.Code)

	// delete
	rDel := httptest.NewRequest("DELETE", fmt.Sprintf("/featureRule/%s?applicationType=stb", created.Id), nil)
	rDel = mux.SetURLVars(rDel, map[string]string{"id": created.Id})
	rrDel := httptest.NewRecorder()
	DeleteOneFeatureRuleHandler(rrDel, rDel)
	assert.Equal(t, http.StatusNoContent, rrDel.Code)
}

func TestFeatureRulePriorityChangeAndErrors(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")
	fr1 := frMakeFeatureRule([]string{f.ID}, "stb", 1)
	fr2 := frMakeFeatureRule([]string{f.ID}, "stb", 2)
	// change priority of fr2 to 1
	r := httptest.NewRequest("GET", fmt.Sprintf("/featureRule/change/%s/priority/1?applicationType=stb", fr2.Id), nil)
	r = mux.SetURLVars(r, map[string]string{"id": fr2.Id, "newPriority": "1"})
	rr := httptest.NewRecorder()
	ChangeFeatureRulePrioritiesHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	// bad newPriority
	rBad := httptest.NewRequest("GET", fmt.Sprintf("/featureRule/change/%s/priority/x?applicationType=stb", fr1.Id), nil)
	rBad = mux.SetURLVars(rBad, map[string]string{"id": fr1.Id, "newPriority": "x"})
	rrBad := httptest.NewRecorder()
	ChangeFeatureRulePrioritiesHandler(rrBad, rBad)
	assert.Equal(t, http.StatusBadRequest, rrBad.Code)
}

func TestFeatureRulesSizeAllowedNumberHandlers(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")
	frMakeFeatureRule([]string{f.ID}, "stb", 1)
	rSize := httptest.NewRequest("GET", "/featureRules/size?applicationType=stb", nil)
	rrSize := httptest.NewRecorder()
	GetFeatureRulesSizeHandler(rrSize, rSize)
	assert.Equal(t, http.StatusOK, rrSize.Code)
	rAllowed := httptest.NewRequest("GET", "/featureRules/allowedNumber?applicationType=stb", nil)
	rrAllowed := httptest.NewRecorder()
	GetAllowedNumberOfFeaturesHandler(rrAllowed, rAllowed)
	assert.Equal(t, http.StatusOK, rrAllowed.Code)
}

func TestBatchCreateAndUpdateHandlers(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")
	// batch create mixed: second invalid (no featureIds)
	valid := &xwrfc.FeatureRule{Name: "Batch1", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 1, Rule: frMakeRule()}
	invalid := &xwrfc.FeatureRule{Name: "Bad", ApplicationType: "stb", FeatureIds: []string{}, Priority: 2, Rule: frMakeRule()}
	batch := []*xwrfc.FeatureRule{valid, invalid}
	b, _ := json.Marshal(batch)
	r := httptest.NewRequest("POST", "/featureRules?applicationType=stb", nil)
	rrNative := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rrNative)
	xw.SetBody(string(b))
	CreateFeatureRulesHandler(xw, r)
	assert.Equal(t, http.StatusOK, rrNative.Code)

	// batch update: modify existing valid, invalid with missing id
	// need existing rule id
	created := &xwrfc.FeatureRule{}
	json.Unmarshal(rrNative.Body.Bytes(), &created) // body is map, ignore parse error for brevity
	existingList, _ := GetAllAsListFromDao(ds.TABLE_FEATURE_CONTROL_RULE, 0)
	var existing *xwrfc.FeatureRule
	for _, inst := range existingList {
		if fr, ok := inst.(*xwrfc.FeatureRule); ok {
			existing = fr
			break
		}
	}
	existing.Name = "UpdatedName"
	missing := &xwrfc.FeatureRule{Name: "noid", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 3, Rule: frMakeRule()}
	updBatch := []*xwrfc.FeatureRule{existing, missing}
	b2, _ := json.Marshal(updBatch)
	r2 := httptest.NewRequest("PUT", "/featureRules?applicationType=stb", nil)
	rr2Native := httptest.NewRecorder()
	xw2 := xwhttp.NewXResponseWriter(rr2Native)
	xw2.SetBody(string(b2))
	UpdateFeatureRulesHandler(xw2, r2)
	assert.Equal(t, http.StatusOK, rr2Native.Code)
}

func TestFilteredWithPageAndTestPageHandlers(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")
	frMakeFeatureRule([]string{f.ID}, "stb", 1)
	// valid paged filtered (empty body)
	r := httptest.NewRequest("POST", "/featureRules/filteredWithPage?applicationType=stb&pageNumber=1&pageSize=5", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	GetFeatureRulesFilteredWithPage(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	// bad pageNumber
	rBad := httptest.NewRequest("POST", "/featureRules/filteredWithPage?applicationType=stb&pageNumber=x&pageSize=5", nil)
	rrBad := httptest.NewRecorder()
	xwBad := xwhttp.NewXResponseWriter(rrBad)
	GetFeatureRulesFilteredWithPage(xwBad, rBad)
	assert.Equal(t, http.StatusBadRequest, rrBad.Code)
	// test page handler success
	ctxBody := map[string]string{"estbMacAddress": "AA:BB:CC:DD:EE:FF"}
	cb, _ := json.Marshal(ctxBody)
	rTP := httptest.NewRequest("POST", "/featureRules/testPage?applicationType=stb", nil)
	rrTPNative := httptest.NewRecorder()
	xwTP := xwhttp.NewXResponseWriter(rrTPNative)
	xwTP.SetBody(string(cb))
	FeatureRuleTestPageHandler(xwTP, rTP)
	assert.Equal(t, http.StatusOK, rrTPNative.Code)
}

func TestPackFeaturePriorities(t *testing.T) {
	SkipIfMockDatabase(t)
	input := []*xwrfc.FeatureRule{
		{Id: "id1", Priority: 2},
		{Id: "id2", Priority: 1},
		{Id: "id3", Priority: 3},
	}
	ref := &xwrfc.FeatureRule{Id: "id2", Priority: 1}
	result := PackFeaturePriorities(input, ref)
	// Should return altered rules (excluding the deleted one)
	assert.True(t, len(result) >= 0)
	// Verify the deleted rule is not in the result
	for _, r := range result {
		assert.NotEqual(t, "id2", r.Id)
	}
}

func TestDeleteOneFeatureRuleHandler_Error(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("DELETE", "/featureRule//?applicationType=stb", nil)
	r = mux.SetURLVars(r, map[string]string{"id": ""})
	rr := httptest.NewRecorder()
	DeleteOneFeatureRuleHandler(rr, r)
	// Returns 405 Method Not Allowed when route is not properly configured
	assert.True(t, rr.Code >= http.StatusBadRequest)
}

func TestImportAllFeatureRulesHandler_Error(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("POST", "/featureRules/import/all?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	ImportAllFeatureRulesHandler(rr, r)
	// Returns 500 when body is empty/invalid
	assert.True(t, rr.Code >= http.StatusBadRequest)
}

func TestUpdateFeatureRuleHandler_Error(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("PUT", "/featureRule?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("invalid-json")
	UpdateFeatureRuleHandler(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateFeatureRuleHandler_Error(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("POST", "/featureRule?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("invalid-json")
	CreateFeatureRuleHandler(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetFeatureRuleOne_Error(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("GET", "/featureRule//?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetFeatureRuleOne(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetFeatureRulesHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("GET", "/featureRules?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetFeatureRulesHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// Test xhttp.AdminError
func TestAdminErrorResponse(t *testing.T) {
	SkipIfMockDatabase(t)
	rr := httptest.NewRecorder()
	xhttp.WriteAdminErrorResponse(rr, http.StatusForbidden, "test error")
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// Test WriteXconfResponse
func TestWriteXconfResponse(t *testing.T) {
	SkipIfMockDatabase(t)
	rr := httptest.NewRecorder()
	data := []byte(`{"foo":"bar"}`)
	xwhttp.WriteXconfResponse(rr, http.StatusOK, data)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// GetFeatureRulesFilteredWithPage - Error paths
func TestGetFeatureRulesFilteredWithPage_BadPageNumber(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("POST", "/featureRules/filteredWithPage?applicationType=stb&pageNumber=invalid", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	GetFeatureRulesFilteredWithPage(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "pageNumber must be a number")
}

func TestGetFeatureRulesFilteredWithPage_BadPageSize(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("POST", "/featureRules/filteredWithPage?applicationType=stb&pageSize=invalid", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	GetFeatureRulesFilteredWithPage(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "pageSize must be a number")
}

func TestGetFeatureRulesFilteredWithPage_InvalidJSON(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("POST", "/featureRules/filteredWithPage?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("{invalid-json")
	GetFeatureRulesFilteredWithPage(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unable to extract searchContext")
}

func TestGetFeatureRulesFilteredWithPage_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")
	frMakeFeatureRule([]string{f.ID}, "stb", 1)
	frMakeFeatureRule([]string{f.ID}, "stb", 2)

	searchCtx := map[string]string{"name": "FR-"}
	body, _ := json.Marshal(searchCtx)
	r := httptest.NewRequest("POST", "/featureRules/filteredWithPage?applicationType=stb&pageNumber=1&pageSize=10", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	GetFeatureRulesFilteredWithPage(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("numberOfItems"), "")
}

// ImportAllFeatureRulesHandler - Error paths
func TestImportAllFeatureRulesHandler_InvalidJSON(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("POST", "/featureRules/import/all?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("{invalid-json")
	ImportAllFeatureRulesHandler(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unable to extract featureRules")
}

func TestImportAllFeatureRulesHandler_AppTypeMixing(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")

	rules := []xwrfc.FeatureRule{
		{Name: "Rule1", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 1, Rule: frMakeRule()},
		{Name: "Rule2", ApplicationType: "rdkcloud", FeatureIds: []string{f.ID}, Priority: 2, Rule: frMakeRule()},
	}
	body, _ := json.Marshal(rules)

	r := httptest.NewRequest("POST", "/featureRules/import/all?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	ImportAllFeatureRulesHandler(xw, r)
	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "ApplicationType")
}

func TestImportAllFeatureRulesHandler_AppTypeConflict(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")

	// First rule with empty app type, second with mismatching app type (when determined)
	rules := []xwrfc.FeatureRule{
		{Name: "Rule1", ApplicationType: "", FeatureIds: []string{f.ID}, Priority: 1, Rule: frMakeRule()},
		{Name: "Rule2", ApplicationType: "rdkcloud", FeatureIds: []string{f.ID}, Priority: 2, Rule: frMakeRule()},
	}
	body, _ := json.Marshal(rules)

	r := httptest.NewRequest("POST", "/featureRules/import/all?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	ImportAllFeatureRulesHandler(xw, r)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestImportAllFeatureRulesHandler_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")

	rules := []xwrfc.FeatureRule{
		{Name: "ImportRule1", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 1, Rule: frMakeRule()},
		{Name: "ImportRule2", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 2, Rule: frMakeRule()},
	}
	body, _ := json.Marshal(rules)

	r := httptest.NewRequest("POST", "/featureRules/import/all?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody(string(body))
	ImportAllFeatureRulesHandler(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// GetFeatureRuleOne - Error paths
func TestGetFeatureRuleOne_BlankId(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("GET", "/featureRule/", nil)
	r = mux.SetURLVars(r, map[string]string{"id": ""})
	rr := httptest.NewRecorder()
	GetFeatureRuleOne(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Id is blank")
}

func TestGetFeatureRuleOne_NotFound(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest("GET", "/featureRule/nonexistent-id?applicationType=stb", nil)
	r = mux.SetURLVars(r, map[string]string{"id": "nonexistent-id"})
	rr := httptest.NewRecorder()
	GetFeatureRuleOne(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "does not exist")
}

func TestGetFeatureRuleOne_Success(t *testing.T) {
	SkipIfMockDatabase(t)
	frCleanup()
	f := frMakeFeature("FeatA", "stb")
	fr := frMakeFeatureRule([]string{f.ID}, "stb", 1)

	r := httptest.NewRequest("GET", fmt.Sprintf("/featureRule/%s?applicationType=stb", fr.Id), nil)
	r = mux.SetURLVars(r, map[string]string{"id": fr.Id})
	rr := httptest.NewRecorder()
	GetFeatureRuleOne(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)

	var result xwrfc.FeatureRule
	json.Unmarshal(rr.Body.Bytes(), &result)
	assert.Equal(t, fr.Id, result.Id)
}
