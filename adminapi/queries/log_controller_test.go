package queries

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/stretchr/testify/assert"
)

// We will stub estbfirmware functions via simple in-package variable indirection if needed.
// For now, call GetLogs with states that exercise each branch.

func TestGetLogs_MissingMac(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/logs", nil)
	rr := httptest.NewRecorder()
	GetLogs(rr, r) // no mux vars -> missing macStr
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing macStr")
}

func TestGetLogs_InvalidMac(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/logs/bad", nil)
	r = mux.SetURLVars(r, map[string]string{"macStr": "BAD-MAC"})
	rr := httptest.NewRecorder()
	GetLogs(rr, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid mac address")
}

func TestGetLogs_NoLogsForValidMac(t *testing.T) {
	// use a valid mac format but ensure estbfirmware returns nil (assuming empty db) => empty map serialized
	r := httptest.NewRequest(http.MethodGet, "/logs/aa:bb:cc:00:00:01", nil)
	r = mux.SetURLVars(r, map[string]string{"macStr": "AA:BB:CC:00:00:01"})
	rr := httptest.NewRecorder()
	GetLogs(rr, r)
	assert.Equal(t, http.StatusOK, rr.Code)
	// body should be an empty JSON object (map with length 0)
	m := map[string]any{}
	_ = json.Unmarshal(rr.Body.Bytes(), &m)
	assert.Len(t, m, 0)
}

// To cover branch where logs exist we create an XResponseWriter environment and inject a fake last + list by temporarily
// creating them directly via internal helpers if accessible; here we rely on package-level helpers getOneConfigChangeLog and getConfigChangeLogList if exported, else we skip.
// We can't directly set estbfirmware cache without deeper seeding; so current coverage focuses on error and empty-success branches.

func TestGetLogs_ResponseWriterCastNotNeeded(t *testing.T) {
	// Ensure code still works when wrapped writer (not required by this handler but sanity test) and logs empty.
	r := httptest.NewRequest(http.MethodGet, "/logs/aa:bb:cc:00:00:02", nil)
	r = mux.SetURLVars(r, map[string]string{"macStr": "AA:BB:CC:00:00:02"})
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	GetLogs(xw, r)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestLogController_InternalHelpers(t *testing.T) {
	// exercise helper returning nil on empty
	if v := getOneConfigChangeLog(""); v != nil {
		t.Fatalf("expected nil for empty mac")
	}
	if v := getConfigChangeLogList(""); v != nil {
		t.Fatalf("expected nil slice for empty mac")
	}
	// exercise populated paths
	one := getOneConfigChangeLog("AA:BB:CC:00:00:03")
	if one == nil || one.ID != "id1" {
		t.Fatalf("unexpected one %#v", one)
	}
	lst := getConfigChangeLogList("AA:BB:CC:00:00:03")
	if len(lst) != 2 {
		t.Fatalf("expected 2 logs got %d", len(lst))
	}
}
