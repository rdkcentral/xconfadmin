package queries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/stretchr/testify/assert"
)

// helper to create a request and execute using the provided handler
func execReq(t *testing.T, method, url string, body []byte) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	assert.NoError(t, err)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// minimal fake auth: override CanRead/CanWrite via build tags would be ideal, but for quick coverage we rely on default no-auth middleware path in tests using direct handler invocation (auth already bypassed in tests setup in queries_test.go). Here we assume auth passes.

func TestGetQueriesIpAddressGroupsByName_Failure_InvalidName(t *testing.T) {
	rr := execReq(t, http.MethodGet, "/xconfAdminService/queries/ipAddressGroups/byName/ ", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetQueriesIpAddressGroupsByName_NotFound_Version3(t *testing.T) {
	rr := execReq(t, http.MethodGet, "/xconfAdminService/queries/ipAddressGroups/byName/doesNotExist?version=3.0", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCreateIpAddressGroupHandler_Failure_BadJSON(t *testing.T) {
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups", []byte("{"))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateIpAddressGroupHandler_Success(t *testing.T) {
	grp := shared.NewIpAddressGroupWithAddrStrings("grp1", "grp1", []string{"127.0.0.1"})
	b, _ := json.Marshal(grp)
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups", b)
	assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, rr.Code)
}

func TestAddDataIpAddressGroupHandler_Failure_MissingListId(t *testing.T) {
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups//addData", []byte("{}"))
	// Gorilla/mux collapses duplicate slashes and may redirect (301); treat 301 or 404 as acceptable failure modes
	assert.Contains(t, []int{http.StatusNotFound, http.StatusMovedPermanently}, rr.Code)
}

func TestAddDataIpAddressGroupHandler_Failure_BadJSON(t *testing.T) {
	// create base group first
	grp := shared.NewIpAddressGroupWithAddrStrings("list1", "list1", []string{})
	b, _ := json.Marshal(grp)
	_ = execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups", b)
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups/list1/addData", []byte("{"))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAddDataIpAddressGroupHandler_Success(t *testing.T) {
	grp := shared.NewIpAddressGroupWithAddrStrings("listAdd", "listAdd", []string{"10.0.0.2"}) // seed with one IP so list exists
	b, _ := json.Marshal(grp)
	_ = execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups", b)
	wrapper := &shared.StringListWrapper{List: []string{"10.0.0.1"}}
	wb, _ := json.Marshal(wrapper)
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups/listAdd/addData", wb)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRemoveDataIpAddressGroupHandler_Failure_BadJSON(t *testing.T) {
	grp := shared.NewIpAddressGroupWithAddrStrings("listRemBad", "listRemBad", []string{"10.0.0.1"})
	b, _ := json.Marshal(grp)
	_ = execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups", b)
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups/listRemBad/removeData", []byte("{"))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRemoveDataIpAddressGroupHandler_Success(t *testing.T) {
	grp := shared.NewIpAddressGroupWithAddrStrings("listRem", "listRem", []string{"10.0.0.1", "10.0.0.2"})
	b, _ := json.Marshal(grp)
	_ = execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups", b)
	wrapper := &shared.StringListWrapper{List: []string{"10.0.0.2"}} // removing one leaves at least one entry
	wb, _ := json.Marshal(wrapper)
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/ipAddressGroups/listRem/removeData", wb)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCreateIpAddressGroupHandlerV2_Failure_BadJSON(t *testing.T) {
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/v2/ipAddressGroups", []byte("{"))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateIpAddressGroupHandlerV2_Success(t *testing.T) {
	grp := shared.NewGenericNamespacedList("grpV2", shared.IP_LIST, []string{"192.168.0.1"})
	b, _ := json.Marshal(grp)
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/v2/ipAddressGroups", b)
	assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, rr.Code)
}

func TestUpdateIpAddressGroupHandlerV2_Failure_BadJSON(t *testing.T) {
	rr := execReq(t, http.MethodPut, "/xconfAdminService/updates/v2/ipAddressGroups", []byte("{"))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateIpAddressGroupHandlerV2_Success(t *testing.T) {
	grp := shared.NewGenericNamespacedList("grpV2Upd", shared.IP_LIST, []string{"172.16.0.5"})
	b, _ := json.Marshal(grp)
	_ = execReq(t, http.MethodPost, "/xconfAdminService/updates/v2/ipAddressGroups", b)
	time.Sleep(10 * time.Millisecond)
	grp.Data = []string{"172.16.0.6"}
	b2, _ := json.Marshal(grp)
	rr := execReq(t, http.MethodPut, "/xconfAdminService/updates/v2/ipAddressGroups", b2)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetQueriesIpAddressGroupsByNameV2_Failure_NoID(t *testing.T) {
	rr := execReq(t, http.MethodGet, "/xconfAdminService/queries/v2/ipAddressGroups/byName/", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code) // route mismatch
}

func TestGetQueriesIpAddressGroupsByNameV2_NotFound(t *testing.T) {
	rr := execReq(t, http.MethodGet, "/xconfAdminService/queries/v2/ipAddressGroups/byName/doesnotexist", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetQueriesIpAddressGroupsByNameV2_Success(t *testing.T) {
	grp := shared.NewGenericNamespacedList("grpLookup", shared.IP_LIST, []string{"8.8.8.8"})
	b, _ := json.Marshal(grp)
	_ = execReq(t, http.MethodPost, "/xconfAdminService/updates/v2/ipAddressGroups", b)
	rr := execReq(t, http.MethodGet, "/xconfAdminService/queries/v2/ipAddressGroups/byName/grpLookup", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetQueriesIpAddressGroupsByIpV2_Failure_InvalidIP(t *testing.T) {
	rr := execReq(t, http.MethodGet, "/xconfAdminService/queries/v2/ipAddressGroups/byIp/notanip", nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetQueriesIpAddressGroupsByIpV2_Success(t *testing.T) {
	grp := shared.NewGenericNamespacedList("grpByIp", shared.IP_LIST, []string{"203.0.113.1"})
	b, _ := json.Marshal(grp)
	_ = execReq(t, http.MethodPost, "/xconfAdminService/updates/v2/ipAddressGroups", b)
	rr := execReq(t, http.MethodGet, "/xconfAdminService/queries/v2/ipAddressGroups/byIp/203.0.113.1", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestDeleteIpAddressGroupHandlerV2_NotFound(t *testing.T) {
	rr := execReq(t, http.MethodDelete, "/xconfAdminService/delete/v2/ipAddressGroups/doesnotexist", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDeleteIpAddressGroupHandlerV2_Success(t *testing.T) {
	grp := shared.NewGenericNamespacedList("grpDelete", shared.IP_LIST, []string{"10.10.10.10"})
	b, _ := json.Marshal(grp)
	_ = execReq(t, http.MethodPost, "/xconfAdminService/updates/v2/ipAddressGroups", b)
	rr := execReq(t, http.MethodDelete, "/xconfAdminService/delete/v2/ipAddressGroups/grpDelete", nil)
	// Delete returns 200 with body or could be 204 based on service logic; accept both
	assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, rr.Code)
}

func TestSaveMacListHandler_Failure_BadJSON(t *testing.T) {
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/nsLists", []byte("{"))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSaveMacListHandler_Success(t *testing.T) {
	ml := shared.NewMacList()
	ml.ID = "mac1"
	ml.Data = []string{"AA:BB:CC:DD:EE:FF"}
	b, _ := json.Marshal(ml)
	rr := execReq(t, http.MethodPost, "/xconfAdminService/updates/nsLists", b)
	assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, rr.Code, fmt.Sprintf("unexpected status %d", rr.Code))
}
