package ipmacrule

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/stretchr/testify/assert"
)

// TestGetIpMacRuleConfigurationHandler_Success verifies a 200 response and JSON body contents
func TestGetIpMacRuleConfigurationHandler_Success(t *testing.T) {
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	r := httptest.NewRequest(http.MethodGet, "/ipmac/config", nil)

	GetIpMacRuleConfigurationHandler(xw, r)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	// Response should contain ipMacIsConditionLimit field (case sensitive per struct tag)
	var body map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	assert.NoError(t, err)
	_, has := body["ipMacIsConditionLimit"]
	assert.True(t, has, "expected ipMacIsConditionLimit field in response")
}
