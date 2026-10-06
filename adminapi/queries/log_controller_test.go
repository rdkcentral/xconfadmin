package queries

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	xcommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestIsMacPresentAndValid(t *testing.T) {
	// no query params
	queryParamsStr := ""
	queryParams, _ := url.ParseQuery(queryParamsStr)
	isValid, mac, errorStr := isMacPresentAndValid(queryParams)
	assert.Equal(t, isValid, false)
	assert.Equal(t, mac, "")
	assert.Equal(t, errorStr, "Required String parameter 'mac' is not present")

	// missing mac query param
	queryParamsStr = "macAddress=1234"
	queryParams, _ = url.ParseQuery(queryParamsStr)
	isValid, mac, errorStr = isMacPresentAndValid(queryParams)
	assert.Equal(t, isValid, false)
	assert.Equal(t, mac, "")
	assert.Equal(t, errorStr, "Required String parameter 'mac' is not present")

	// invalid mac
	queryParamsStr = "macAddress=1234&mac=4321"
	queryParams, _ = url.ParseQuery(queryParamsStr)
	isValid, mac, errorStr = isMacPresentAndValid(queryParams)
	assert.Equal(t, isValid, false)
	assert.Equal(t, mac, "4321")
	assert.Equal(t, errorStr, "Mac is invalid: 4321")

	// valid mac
	queryParamsStr = "macAddress=1234&mac=00:1B:44:11:3A:B7&query=param"
	queryParams, _ = url.ParseQuery(queryParamsStr)
	isValid, mac, errorStr = isMacPresentAndValid(queryParams)
	assert.Equal(t, isValid, true)
	assert.Equal(t, mac, "00:1B:44:11:3A:B7")
	assert.Equal(t, errorStr, "")
}

func TestGetEstbLastlogPath(t *testing.T) {
	setSATDisabledForLogHandlerTest(t)

	tests := []struct {
		name       string
		url        string
		statusCode int
	}{
		{name: "missing mac", url: "/xconfAdminService/estbfirmware/lastlog", statusCode: http.StatusBadRequest},
		{name: "invalid mac", url: "/xconfAdminService/estbfirmware/lastlog?mac=invalid", statusCode: http.StatusBadRequest},
		{name: "empty result", url: "/xconfAdminService/estbfirmware/lastlog?mac=AA:BB:CC:00:00:11", statusCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rr := httptest.NewRecorder()

			GetEstbLastlogPath(rr, r)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

func TestGetEstbChangelogsPath(t *testing.T) {
	setSATDisabledForLogHandlerTest(t)

	tests := []struct {
		name       string
		url        string
		statusCode int
	}{
		{name: "missing mac", url: "/xconfAdminService/estbfirmware/changelogs", statusCode: http.StatusBadRequest},
		{name: "invalid mac", url: "/xconfAdminService/estbfirmware/changelogs?mac=invalid", statusCode: http.StatusBadRequest},
		{name: "empty result", url: "/xconfAdminService/estbfirmware/changelogs?mac=AA:BB:CC:00:00:12", statusCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rr := httptest.NewRecorder()

			GetEstbChangelogsPath(rr, r)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

func TestGetEstbLastlogPath_TenantIDMismatch(t *testing.T) {
	setSATDisabledForLogHandlerTest(t)
	const (
		macAddress = "AA:BB:CC:00:00:21"
		tenantID   = "tenant-a"
	)
	stubLogGetters(t, func(string, string) *estbfirmware.ConfigChangeLog {
		return &estbfirmware.ConfigChangeLog{TenantId: "tenant-b"}
	}, nil)

	r := logRequestWithTenant("/xconfAdminService/estbfirmware/lastlog?mac="+macAddress, tenantID)
	rr := httptest.NewRecorder()
	GetEstbLastlogPath(rr, r)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Tenant ID mismatch")
}

func TestGetEstbChangelogsPath_FiltersTenantLogs(t *testing.T) {
	setSATDisabledForLogHandlerTest(t)
	const (
		macAddress = "AA:BB:CC:00:00:22"
		tenantID   = "tenant-a"
	)
	stubLogGetters(t, nil, func(string, string) []*estbfirmware.ConfigChangeLog {
		return []*estbfirmware.ConfigChangeLog{
			{ID: "tenant-a-log", Updated: 1, TenantId: tenantID},
			{ID: "tenant-b-log", Updated: 2, TenantId: "tenant-b"},
		}
	})

	r := logRequestWithTenant("/xconfAdminService/estbfirmware/changelogs?mac="+macAddress, tenantID)
	rr := httptest.NewRecorder()
	GetEstbChangelogsPath(rr, r)

	require.Equal(t, http.StatusOK, rr.Code)
	var logs []*estbfirmware.ConfigChangeLog
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &logs))
	require.Len(t, logs, 1)
	assert.Equal(t, tenantID, logs[0].TenantId)
	assert.Empty(t, logs[0].ID)
	assert.Zero(t, logs[0].Updated)
}

func TestGetEstbChangelogsPath_TenantIDMismatch(t *testing.T) {
	setSATDisabledForLogHandlerTest(t)
	const (
		macAddress = "AA:BB:CC:00:00:23"
		tenantID   = "tenant-a"
	)
	stubLogGetters(t, nil, func(string, string) []*estbfirmware.ConfigChangeLog {
		return []*estbfirmware.ConfigChangeLog{{TenantId: "tenant-b"}}
	})

	r := logRequestWithTenant("/xconfAdminService/estbfirmware/changelogs?mac="+macAddress, tenantID)
	rr := httptest.NewRecorder()
	GetEstbChangelogsPath(rr, r)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Tenant ID mismatch")
}

func logRequestWithTenant(target, tenantID string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, target, nil)
	return r.WithContext(context.WithValue(r.Context(), xhttp.CTX_KEY_TENANT_ID, tenantID))
}

func stubLogGetters(t *testing.T, lastConfigLog func(string, string) *estbfirmware.ConfigChangeLog, configChangeLogs func(string, string) []*estbfirmware.ConfigChangeLog) {
	t.Helper()
	originalLastConfigLog := getLastConfigLog
	originalConfigChangeLogs := getConfigChangeLogsOnly
	if lastConfigLog != nil {
		getLastConfigLog = lastConfigLog
	}
	if configChangeLogs != nil {
		getConfigChangeLogsOnly = configChangeLogs
	}
	t.Cleanup(func() {
		getLastConfigLog = originalLastConfigLog
		getConfigChangeLogsOnly = originalConfigChangeLogs
	})
}

func setSATDisabledForLogHandlerTest(t *testing.T) {
	t.Helper()
	previousSatOn := xcommon.SatOn
	xcommon.SatOn = false
	t.Cleanup(func() {
		xcommon.SatOn = previousSatOn
	})
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

func TestLogPreDisplayCleanup(t *testing.T) {
	tests := []struct {
		name           string
		lastConfigLog  *estbfirmware.ConfigChangeLog
		expectedID     string
		expectedUpdate int64
	}{
		{
			name: "Clean up non-nil log",
			lastConfigLog: &estbfirmware.ConfigChangeLog{
				ID:      "test-id-123",
				Updated: 1234567890,
			},
			expectedID:     "",
			expectedUpdate: 0,
		},
		{
			name:           "Nil log does nothing",
			lastConfigLog:  nil,
			expectedID:     "",
			expectedUpdate: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logPreDisplayCleanup(tt.lastConfigLog)

			if tt.lastConfigLog != nil {
				assert.Equal(t, tt.expectedID, tt.lastConfigLog.ID)
				assert.Equal(t, tt.expectedUpdate, tt.lastConfigLog.Updated)
			}
		})
	}
}
