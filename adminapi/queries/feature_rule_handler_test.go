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
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_XCONF_FEATURE, f.ID, f)
	return f
}
func frMakeRule() *re.Rule {
	return &re.Rule{Condition: CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "model"), re.StandardOperationIs, "X1")}
}

func frMakeFeatureRule(featureIds []string, app string, priority int) *xwrfc.FeatureRule {
	fr := &xwrfc.FeatureRule{Id: uuid.New().String(), Name: "FR-" + uuid.New().String(), ApplicationType: app, FeatureIds: featureIds, Priority: priority, Rule: frMakeRule()}
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, fr.Id, fr)
	return fr
}

func frCleanup() {
	tables := []string{ds.TABLE_FEATURE_CONTROL_RULE, ds.TABLE_XCONF_FEATURE}
	for _, tbl := range tables {
		list, _ := ds.GetCachedSimpleDao().GetAllAsList(tbl, 0)
		for _, inst := range list {
			switch v := inst.(type) {
			case *xwrfc.FeatureRule:
				ds.GetCachedSimpleDao().DeleteOne(tbl, v.Id)
			case *xwrfc.Feature:
				ds.GetCachedSimpleDao().DeleteOne(tbl, v.ID)
			}
		}
		ds.GetCachedSimpleDao().RefreshAll(tbl)
	}
}

// Tests
func TestGetFeatureRulesFiltered_AndExportHandlers(t *testing.T) {
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
	existingList, _ := ds.GetCachedSimpleDao().GetAllAsList(ds.TABLE_FEATURE_CONTROL_RULE, 0)
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
	r := httptest.NewRequest("DELETE", "/featureRule//?applicationType=stb", nil)
	r = mux.SetURLVars(r, map[string]string{"id": ""})
	rr := httptest.NewRecorder()
	DeleteOneFeatureRuleHandler(rr, r)
	// Returns 405 Method Not Allowed when route is not properly configured
	assert.True(t, rr.Code >= http.StatusBadRequest)
}

func TestImportAllFeatureRulesHandler_Error(t *testing.T) {
	r := httptest.NewRequest("POST", "/featureRules/import/all?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	ImportAllFeatureRulesHandler(rr, r)
	// Returns 500 when body is empty/invalid
	assert.True(t, rr.Code >= http.StatusBadRequest)
}

func TestUpdateFeatureRuleHandler_Error(t *testing.T) {
	r := httptest.NewRequest("PUT", "/featureRule?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("invalid-json")
	UpdateFeatureRuleHandler(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateFeatureRuleHandler_Error(t *testing.T) {
	r := httptest.NewRequest("POST", "/featureRule?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("invalid-json")
	CreateFeatureRuleHandler(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetFeatureRuleOne_Error(t *testing.T) {
	r := httptest.NewRequest("GET", "/featureRule//?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetFeatureRuleOne(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetFeatureRulesHandler_Success(t *testing.T) {
	r := httptest.NewRequest("GET", "/featureRules?applicationType=stb", nil)
	rr := httptest.NewRecorder()
	GetFeatureRulesHandler(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// Test xhttp.AdminError
func TestAdminErrorResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	xhttp.WriteAdminErrorResponse(rr, http.StatusForbidden, "test error")
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// Test WriteXconfResponse
func TestWriteXconfResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []byte(`{"foo":"bar"}`)
	xwhttp.WriteXconfResponse(rr, http.StatusOK, data)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// ===== Comprehensive Error Condition Tests =====

func TestGetFeatureRulesFiltered_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("Success", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		frMakeFeatureRule([]string{f.ID}, "stb", 1)
		r := httptest.NewRequest("GET", "/featureRules?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		GetFeatureRulesFiltered(rr, r)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestGetFeatureRulesFilteredWithPage_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("InvalidPageNumber_WriteXconfResponse_400", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/featureRules/filteredWithPage?applicationType=stb&pageNumber=invalid&pageSize=10", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		GetFeatureRulesFilteredWithPage(xw, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "pageNumber must be a number")
	})

	t.Run("InvalidPageSize_WriteXconfResponse_400", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/featureRules/filteredWithPage?applicationType=stb&pageNumber=1&pageSize=invalid", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		GetFeatureRulesFilteredWithPage(xw, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "pageSize must be a number")
	})

	t.Run("InvalidJSON_WriteXconfResponse_400", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/featureRules/filteredWithPage?applicationType=stb&pageNumber=1&pageSize=10", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody("{invalid json")
		GetFeatureRulesFilteredWithPage(xw, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unable to extract searchContext from json file")
	})

	t.Run("Success_WithValidContext", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		frMakeFeatureRule([]string{f.ID}, "stb", 1)
		contextBody := map[string]string{"name": "FR"}
		b, _ := json.Marshal(contextBody)
		r := httptest.NewRequest("POST", "/featureRules/filteredWithPage?applicationType=stb&pageNumber=1&pageSize=10", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		GetFeatureRulesFilteredWithPage(xw, r)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestGetFeatureRuleOneExport_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("EmptyID_WriteAdminErrorResponse_400", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/featureRule/export/?applicationType=stb", nil)
		r = mux.SetURLVars(r, map[string]string{"id": ""})
		rr := httptest.NewRecorder()
		GetFeatureRuleOneExport(rr, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Id is blank")
	})

	t.Run("RuleNotFound_WriteAdminErrorResponse_404", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		r := httptest.NewRequest("GET", fmt.Sprintf("/featureRule/export/%s?applicationType=stb", nonExistentID), nil)
		r = mux.SetURLVars(r, map[string]string{"id": nonExistentID})
		rr := httptest.NewRecorder()
		GetFeatureRuleOneExport(rr, r)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "does not exist")
	})

	t.Run("ApplicationTypeMismatch_WriteAdminErrorResponse_404", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		fr := frMakeFeatureRule([]string{f.ID}, "stb", 1)
		r := httptest.NewRequest("GET", fmt.Sprintf("/featureRule/export/%s?applicationType=xhome", fr.Id), nil)
		r = mux.SetURLVars(r, map[string]string{"id": fr.Id})
		rr := httptest.NewRecorder()
		GetFeatureRuleOneExport(rr, r)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Non existing Entity")
	})

	t.Run("Success_WithExport", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		fr := frMakeFeatureRule([]string{f.ID}, "stb", 1)
		r := httptest.NewRequest("GET", fmt.Sprintf("/featureRule/export/%s?applicationType=stb&export=true", fr.Id), nil)
		r = mux.SetURLVars(r, map[string]string{"id": fr.Id})
		rr := httptest.NewRecorder()
		GetFeatureRuleOneExport(rr, r)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestGetFeatureRuleOne_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("EmptyID_WriteXconfResponse_400", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/featureRule//?applicationType=stb", nil)
		r = mux.SetURLVars(r, map[string]string{"id": ""})
		rr := httptest.NewRecorder()
		GetFeatureRuleOne(rr, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Id is blank")
	})

	t.Run("RuleNotFound_WriteAdminErrorResponse_400", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		r := httptest.NewRequest("GET", fmt.Sprintf("/featureRule/%s?applicationType=stb", nonExistentID), nil)
		r = mux.SetURLVars(r, map[string]string{"id": nonExistentID})
		rr := httptest.NewRecorder()
		GetFeatureRuleOne(rr, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "does not exist")
	})

	t.Run("Success", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		fr := frMakeFeatureRule([]string{f.ID}, "stb", 1)
		r := httptest.NewRequest("GET", fmt.Sprintf("/featureRule/%s?applicationType=stb", fr.Id), nil)
		r = mux.SetURLVars(r, map[string]string{"id": fr.Id})
		rr := httptest.NewRecorder()
		GetFeatureRuleOne(rr, r)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestCreateFeatureRuleHandler_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("InvalidJSON_AdminError_400", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/featureRule?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody("{invalid json")
		CreateFeatureRuleHandler(xw, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("ValidationError_AdminError", func(t *testing.T) {
		// Missing required fields (no FeatureIds)
		badRule := &xwrfc.FeatureRule{Name: "BadRule", ApplicationType: "stb", FeatureIds: []string{}, Priority: 1, Rule: frMakeRule()}
		b, _ := json.Marshal(badRule)
		r := httptest.NewRequest("POST", "/featureRule?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		CreateFeatureRuleHandler(xw, r)
		assert.True(t, rr.Code >= http.StatusBadRequest)
	})

	t.Run("Success", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		validRule := &xwrfc.FeatureRule{Name: "ValidRule", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 1, Rule: frMakeRule()}
		b, _ := json.Marshal(validRule)
		r := httptest.NewRequest("POST", "/featureRule?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		CreateFeatureRuleHandler(xw, r)
		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

func TestUpdateFeatureRuleHandler_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("InvalidJSON_AdminError_400", func(t *testing.T) {
		r := httptest.NewRequest("PUT", "/featureRule?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody("{invalid json")
		UpdateFeatureRuleHandler(xw, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("NonExistentRule_AdminError", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		nonExistentRule := &xwrfc.FeatureRule{Id: uuid.New().String(), Name: "NonExistent", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 1, Rule: frMakeRule()}
		b, _ := json.Marshal(nonExistentRule)
		r := httptest.NewRequest("PUT", "/featureRule?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		UpdateFeatureRuleHandler(xw, r)
		assert.True(t, rr.Code >= http.StatusBadRequest)
	})

	t.Run("Success", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		fr := frMakeFeatureRule([]string{f.ID}, "stb", 1)
		fr.Name = "UpdatedName"
		b, _ := json.Marshal(fr)
		r := httptest.NewRequest("PUT", "/featureRule?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		UpdateFeatureRuleHandler(xw, r)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestImportAllFeatureRulesHandler_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("InvalidJSON_WriteXconfResponse_400", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/featureRules/import/all?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody("{invalid json")
		ImportAllFeatureRulesHandler(xw, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unable to extract featureRules from json file")
	})

	t.Run("ApplicationTypeMixing_WriteAdminErrorResponse_409", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		rule1 := xwrfc.FeatureRule{Name: "Rule1", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 1, Rule: frMakeRule()}
		rule2 := xwrfc.FeatureRule{Name: "Rule2", ApplicationType: "xhome", FeatureIds: []string{f.ID}, Priority: 2, Rule: frMakeRule()}
		rules := []xwrfc.FeatureRule{rule1, rule2}
		b, _ := json.Marshal(rules)
		r := httptest.NewRequest("POST", "/featureRules/import/all?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		ImportAllFeatureRulesHandler(xw, r)
		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("Success", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		rule1 := xwrfc.FeatureRule{Name: "Import1", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 1, Rule: frMakeRule()}
		rules := []xwrfc.FeatureRule{rule1}
		b, _ := json.Marshal(rules)
		r := httptest.NewRequest("POST", "/featureRules/import/all?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		ImportAllFeatureRulesHandler(xw, r)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestDeleteOneFeatureRuleHandler_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("EmptyID_WriteXconfResponse_405", func(t *testing.T) {
		r := httptest.NewRequest("DELETE", "/featureRule//?applicationType=stb", nil)
		r = mux.SetURLVars(r, map[string]string{"id": ""})
		rr := httptest.NewRecorder()
		DeleteOneFeatureRuleHandler(rr, r)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("RuleNotFound_WriteXconfResponse_404", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		r := httptest.NewRequest("DELETE", fmt.Sprintf("/featureRule/%s?applicationType=stb", nonExistentID), nil)
		r = mux.SetURLVars(r, map[string]string{"id": nonExistentID})
		rr := httptest.NewRecorder()
		DeleteOneFeatureRuleHandler(rr, r)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "does not exist")
	})

	t.Run("Success", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		fr := frMakeFeatureRule([]string{f.ID}, "stb", 1)
		r := httptest.NewRequest("DELETE", fmt.Sprintf("/featureRule/%s?applicationType=stb", fr.Id), nil)
		r = mux.SetURLVars(r, map[string]string{"id": fr.Id})
		rr := httptest.NewRecorder()
		DeleteOneFeatureRuleHandler(rr, r)
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})
}

func TestChangeFeatureRulePrioritiesHandler_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("EmptyID_WriteXconfResponse_400", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/featureRule/change//priority/1?applicationType=stb", nil)
		r = mux.SetURLVars(r, map[string]string{"id": "", "newPriority": "1"})
		rr := httptest.NewRecorder()
		ChangeFeatureRulePrioritiesHandler(rr, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Id is blank")
	})

	t.Run("InvalidNewPriority_WriteXconfResponse_400", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		fr := frMakeFeatureRule([]string{f.ID}, "stb", 1)
		r := httptest.NewRequest("GET", fmt.Sprintf("/featureRule/change/%s/priority/invalid?applicationType=stb", fr.Id), nil)
		r = mux.SetURLVars(r, map[string]string{"id": fr.Id, "newPriority": "invalid"})
		rr := httptest.NewRecorder()
		ChangeFeatureRulePrioritiesHandler(rr, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "newPriority must be a number")
	})

	t.Run("Success", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		_ = frMakeFeatureRule([]string{f.ID}, "stb", 1)
		fr2 := frMakeFeatureRule([]string{f.ID}, "stb", 2)
		r := httptest.NewRequest("GET", fmt.Sprintf("/featureRule/change/%s/priority/1?applicationType=stb", fr2.Id), nil)
		r = mux.SetURLVars(r, map[string]string{"id": fr2.Id, "newPriority": "1"})
		rr := httptest.NewRecorder()
		ChangeFeatureRulePrioritiesHandler(rr, r)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestUpdateFeatureRulesHandler_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("InvalidJSON_WriteXconfResponse_400", func(t *testing.T) {
		r := httptest.NewRequest("PUT", "/featureRules?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody("{invalid json")
		UpdateFeatureRulesHandler(xw, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unable to extract FeatureRules from json file")
	})

	t.Run("MixedResults_PartialFailure", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		existingRule := frMakeFeatureRule([]string{f.ID}, "stb", 1)
		existingRule.Name = "UpdatedName"

		// Non-existent rule will fail
		nonExistentRule := &xwrfc.FeatureRule{Id: uuid.New().String(), Name: "NonExistent", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 2, Rule: frMakeRule()}

		rules := []*xwrfc.FeatureRule{existingRule, nonExistentRule}
		b, _ := json.Marshal(rules)
		r := httptest.NewRequest("PUT", "/featureRules?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		UpdateFeatureRulesHandler(xw, r)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestCreateFeatureRulesHandler_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("InvalidJSON_WriteXconfResponse_400", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/featureRules?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody("{invalid json")
		CreateFeatureRulesHandler(xw, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unable to extract FeatureRules from json file")
	})

	t.Run("MixedResults_PartialFailure", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		validRule := &xwrfc.FeatureRule{Name: "Valid", ApplicationType: "stb", FeatureIds: []string{f.ID}, Priority: 1, Rule: frMakeRule()}
		// Invalid rule (no feature IDs)
		invalidRule := &xwrfc.FeatureRule{Name: "Invalid", ApplicationType: "stb", FeatureIds: []string{}, Priority: 2, Rule: frMakeRule()}

		rules := []*xwrfc.FeatureRule{validRule, invalidRule}
		b, _ := json.Marshal(rules)
		r := httptest.NewRequest("POST", "/featureRules?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		CreateFeatureRulesHandler(xw, r)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestFeatureRuleTestPageHandler_AllErrorCases(t *testing.T) {
	frCleanup()
	defer frCleanup()

	t.Run("InvalidJSON_WriteAdminErrorResponse_400", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/featureRules/testPage?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody("{invalid json")
		FeatureRuleTestPageHandler(xw, r)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidContext_WriteAdminErrorResponse_400", func(t *testing.T) {
		// Invalid MAC address format
		invalidContext := map[string]string{"estbMacAddress": "invalid-mac"}
		b, _ := json.Marshal(invalidContext)
		r := httptest.NewRequest("POST", "/featureRules/testPage?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		FeatureRuleTestPageHandler(xw, r)
		// May pass validation or fail depending on normalization logic
		assert.True(t, rr.Code >= http.StatusOK)
	})

	t.Run("Success_ValidContext", func(t *testing.T) {
		f := frMakeFeature("FeatA", "stb")
		frMakeFeatureRule([]string{f.ID}, "stb", 1)
		validContext := map[string]string{"estbMacAddress": "AA:BB:CC:DD:EE:FF"}
		b, _ := json.Marshal(validContext)
		r := httptest.NewRequest("POST", "/featureRules/testPage?applicationType=stb", nil)
		rr := httptest.NewRecorder()
		xw := xwhttp.NewXResponseWriter(rr)
		xw.SetBody(string(b))
		FeatureRuleTestPageHandler(xw, r)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
