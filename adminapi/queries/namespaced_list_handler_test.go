package queries

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
)

// helper to wrap recorder for drained body handlers
func makeXW(body any) (*httptest.ResponseRecorder, *xwhttp.XResponseWriter) {
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	if body != nil {
		b, _ := json.Marshal(body)
		xw.SetBody(string(b))
	}
	return rr, xw
}

func sampleIPGroup(id string, ips []string) *shared.IpAddressGroup {
	return shared.NewIpAddressGroupWithAddrStrings(id, id, ips)
}

func TestIpAddressGroupHandlers_BasicFlow(t *testing.T) {
	// create ip group
	grp := sampleIPGroup("G1", []string{"10.0.0.1"})
	rrCreate, xw := makeXW(grp)
	rCreate := httptest.NewRequest(http.MethodPost, "/ip/address/group", nil)
	CreateIpAddressGroupHandler(xw, rCreate)
	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", rrCreate.Code, rrCreate.Body.String())
	}

	// list groups
	rList := httptest.NewRequest(http.MethodGet, "/ip/address/group", nil)
	rrList := httptest.NewRecorder()
	GetQueriesIpAddressGroups(rrList, rList)
	if rrList.Code != http.StatusOK {
		t.Fatalf("list groups expected 200 got %d", rrList.Code)
	}

	// by name
	rByName := httptest.NewRequest(http.MethodGet, "/ip/address/group/name/G1", nil)
	rByName = mux.SetURLVars(rByName, map[string]string{"name": "G1"})
	rrByName := httptest.NewRecorder()
	GetQueriesIpAddressGroupsByName(rrByName, rByName)
	if rrByName.Code != http.StatusOK {
		t.Fatalf("get by name expected 200 got %d body=%s", rrByName.Code, rrByName.Body.String())
	}

	// by ip
	rByIp := httptest.NewRequest(http.MethodGet, "/ip/address/group/ip/10.0.0.1", nil)
	rByIp = mux.SetURLVars(rByIp, map[string]string{"ipAddress": "10.0.0.1"})
	rrByIp := httptest.NewRecorder()
	GetQueriesIpAddressGroupsByIp(rrByIp, rByIp)
	if rrByIp.Code != http.StatusOK {
		t.Fatalf("get by ip expected 200 got %d", rrByIp.Code)
	}
}

func TestIpAddressGroupHandlers_ErrorBranches(t *testing.T) {
	// create with bad json (simulate cast success but unmarshal fail)
	rrBad, xwBad := makeXW("not-json")
	rBad := httptest.NewRequest(http.MethodPost, "/ip/address/group", nil)
	CreateIpAddressGroupHandler(xwBad, rBad)
	if rrBad.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rrBad.Code)
	}

	// invalid IP parameter format
	rBadIp := httptest.NewRequest(http.MethodGet, "/ip/address/group/ip/zzzz", nil)
	rBadIp = mux.SetURLVars(rBadIp, map[string]string{"ipAddress": "zzzz"})
	rrBadIp := httptest.NewRecorder()
	GetQueriesIpAddressGroupsByIp(rrBadIp, rBadIp)
	if rrBadIp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad ip got %d", rrBadIp.Code)
	}

	// missing name var
	rNoName := httptest.NewRequest(http.MethodGet, "/ip/address/group/name/", nil)
	rrNoName := httptest.NewRecorder()
	GetQueriesIpAddressGroupsByName(rrNoName, rNoName)
	if rrNoName.Code != http.StatusNotFound {
		t.Fatalf("expected 404 missing name got %d", rrNoName.Code)
	}
}
