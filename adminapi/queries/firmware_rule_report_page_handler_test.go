package queries

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/stretchr/testify/assert"
)

// helper for XResponseWriter body
func makeFirmwareReportXW(obj any) (*httptest.ResponseRecorder, *xwhttp.XResponseWriter) {
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	if obj != nil {
		b, _ := json.Marshal(obj)
		xw.SetBody(string(b))
	}
	return rr, xw
}

func TestPostFirmwareRuleReportPageHandler_ResponseWriterCastError(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest(http.MethodPost, "/firmware/report", nil)
	rr := httptest.NewRecorder()
	PostFirmwareRuleReportPageHandler(rr, r)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestPostFirmwareRuleReportPageHandler_BadJSON(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest(http.MethodPost, "/firmware/report", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("[not json")
	PostFirmwareRuleReportPageHandler(xw, r)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetMacAddresses(t *testing.T) {
	t.Parallel()
	listId := "macList1"
	macA := "AA:BB:CC:DD:EE:01"
	macB := "AA:BB:CC:DD:EE:02"
	macSingle := "AA:BB:CC:DD:EE:FF"
	// Persist list
	namedList := shared.NewGenericNamespacedList(listId, shared.MacList, []string{macA, macB})
	_ = shared.CreateGenericNamedListOneDB(namedList)

	// Build firmware rule JSON with two compound parts: one IN_LIST (listId) and one IS (macSingle)
	ruleJSON := `{
            "id": "rule-1",
            "name": "mac rule",
            "rule": {
                "negated": false,
                "condition": {
                    "freeArg": {"type": "STRING", "name": "eStbMac"},
                    "operation": "IN_LIST",
                    "fixedArg": {"bean": {"value": {"java.lang.String": "` + listId + `"}}}
                },
                "compoundParts": [
                    {"negated": false,
                        "relation": "AND",
                        "condition": {"freeArg": {"type": "STRING", "name": "eStbMac"}, "operation": "IS", "fixedArg": {"bean": {"value": {"java.lang.String": "` + macSingle + `"}}}},
                        "compoundParts": []}
                ]
            },
            "applicableAction": {"type": ".RuleAction", "actionType": "RULE", "configId": "cfg1", "configEntries": [], "active": true, "useAccountPercentage": false, "firmwareCheckRequired": false, "rebootImmediately": false},
            "type": "IV_RULE",
            "active": true,
            "applicationType": "stb"
        }`
	fr := &corefw.FirmwareRule{}
	_ = json.Unmarshal([]byte(ruleJSON), fr)

	macs := getMacAddresses([]interface{}{fr})
	assert.Len(t, macs, 3)
}

func TestPostFirmwareRuleReportPageHandler_SuccessEmptyRules(t *testing.T) {
	t.Parallel()
	// empty list -> should still 200 with headers after writing empty report
	rr, xw := makeFirmwareReportXW([]string{})
	r := httptest.NewRequest(http.MethodPost, "/firmware/report", nil)
	PostFirmwareRuleReportPageHandler(xw, r)
	// expect OK
	assert.Equal(t, http.StatusOK, rr.Code)
	// check header presence
	assert.Equal(t, "attachment; filename=filename=report.xls", rr.Header().Get("Content-Disposition"))
	assert.Equal(t, "application/vnd.ms-excel", rr.Header().Get("Content-Type"))
}
