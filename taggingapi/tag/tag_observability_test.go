package tag

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/rdkcentral/xconfadmin/common"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

// Write op: the request-ends audit fields carry per-store outcome counts and a
// first-error message instead of per-member log lines.
func TestAddMembersHandler_RequestEndsAuditFields(t *testing.T) {
	setupTestEnvironment()
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "MFAIL") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	withMockDbClient(t, &mockDbClient{})

	req := httptest.NewRequest("PUT", "/taggingService/tags/audit-tag/members", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "audit-tag"})
	xw := xwhttp.NewXResponseWriter(httptest.NewRecorder())
	xw.SetBody(`["MFAIL", "MOK"]`)

	AddMembersToTagHandler(xw, req)

	fields := xw.Audit()
	assert.Equal(t, OpAddMembers, fields["op"])
	assert.Equal(t, "audit-tag", fields["tag"])
	assert.Equal(t, 2, fields["requested"])
	assert.Equal(t, 1, fields["xdas_ok"])
	assert.Equal(t, 1, fields["xdas_fail"])
	assert.Equal(t, 1, fields["cassandra_ok"])
	assert.Equal(t, 0, fields["cassandra_fail"])
	assert.NotEmpty(t, fields["first_error"])
}

// Read op: pagination parameters and Cassandra cost are visible on the
// request-ends line.
func TestGetTagMembersHandler_PaginatedAuditFields(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 1}, {"bucket_id": 2}}, nil
		}
		return []map[string]any{{"member": "m" + params[2]}}, nil
	})

	req := httptest.NewRequest("GET", "/taggingService/tags/some-tag/members?limit=1", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "some-tag"})
	xw := xwhttp.NewXResponseWriter(httptest.NewRecorder())

	GetTagMembersHandler(xw, req)

	fields := xw.Audit()
	assert.Equal(t, OpGetMembersPage, fields["op"])
	assert.Equal(t, "some-tag", fields["tag"])
	assert.Equal(t, 1, fields["page_limit"])
	assert.Equal(t, false, fields["has_cursor"])
	assert.Equal(t, 2, fields["buckets"])
	assert.Equal(t, 1, fields["num_results"])
	assert.Equal(t, true, fields["has_more"])
	// 1 metadata query + 1 per bucket
	assert.Equal(t, 3, fields["queries"])
}

// Audit fields must be present on error responses too — a failed request
// still reports how far it got.
func TestAddMembersHandler_AuditFieldsOnCassandraFailure(t *testing.T) {
	setupTestEnvironment()
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	withMockDbClient(t, &mockDbClient{
		execBatchFn: func(batch *mockBatch) error {
			return errors.New("cassandra write timeout")
		},
	})

	req := httptest.NewRequest("PUT", "/taggingService/tags/audit-tag/members", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "audit-tag"})
	xw := xwhttp.NewXResponseWriter(httptest.NewRecorder())
	xw.SetBody(`["MOK1", "MOK2"]`)

	AddMembersToTagHandler(xw, req)

	fields := xw.Audit()
	assert.Equal(t, 2, fields["xdas_ok"])
	assert.Equal(t, 0, fields["cassandra_ok"])
	assert.Equal(t, 2, fields["cassandra_fail"])
	assert.Contains(t, fields["first_error"], "cassandra write timeout")
}

// An XDAS outage during a bulk add must not log one line per member, and no
// member value may appear in any log line.
func TestAddMembers_XdasOutageDoesNotFloodLogs(t *testing.T) {
	setupTestEnvironment()
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	withMockDbClient(t, &mockDbClient{})

	hook := logtest.NewGlobal()
	defer hook.Reset()
	oldLevel := log.GetLevel()
	log.SetLevel(log.InfoLevel)
	defer log.SetLevel(oldLevel)

	members := make([]string, 200)
	for i := range members {
		members[i] = fmt.Sprintf("AA:BB:CC:00:%02X:%02X", i/256, i%256)
	}

	stats, err := AddMembersWithXdas("outage-tag", members, "", TagTypeLegacy)
	// A total XDAS outage is an error, not a 202 reporting stored=0 — that made a
	// keyspace rejecting every write look like a healthy API storing nothing.
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadGateway, xwcommon.GetXconfErrorStatusCode(err))
	assert.Equal(t, 200, stats.XdasFail)
	assert.Equal(t, 0, stats.XdasOk)
	assert.NotEmpty(t, stats.FirstError)

	entries := hook.AllEntries()
	assert.Less(t, len(entries), 10, "a 200-member XDAS outage must not emit per-member log lines")
	for _, e := range entries {
		for _, member := range members {
			assert.NotContains(t, e.Message, member, "member values must not be logged")
		}
	}
}

// Background deletion START/END lines carry the audit_id of the request that
// queued the deletion.
func TestDeleteTag_LogsCarryAuditId(t *testing.T) {
	setupTestEnvironment()
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 1}}, nil
		}
		return []map[string]any{}, nil
	})

	hook := logtest.NewGlobal()
	defer hook.Reset()

	err := DeleteTag("audited-tag", "audit-123")
	assert.NoError(t, err)

	var started, completed bool
	for _, e := range hook.AllEntries() {
		if strings.HasPrefix(e.Message, "tag deletion started") {
			started = true
			assert.Equal(t, "audit-123", e.Data["audit_id"])
		}
		if strings.HasPrefix(e.Message, "tag deletion completed") {
			completed = true
			assert.Equal(t, "audit-123", e.Data["audit_id"])
		}
	}
	assert.True(t, started, "expected a 'tag deletion started' log line")
	assert.True(t, completed, "expected a 'tag deletion completed' log line")
}

// A Cassandra outage during a bulk write must not log one line per bucket: the
// members spread over many buckets and every batch fails, but only the aggregate
// error line (plus the divergence line) may be emitted.
func TestAddMembers_CassandraOutageDoesNotFloodLogs(t *testing.T) {
	setupTestEnvironment()
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	withMockDbClient(t, &mockDbClient{
		execBatchFn: func(batch *mockBatch) error {
			return errors.New("cassandra unavailable")
		},
	})

	hook := logtest.NewGlobal()
	defer hook.Reset()
	oldLevel := log.GetLevel()
	log.SetLevel(log.InfoLevel)
	defer log.SetLevel(oldLevel)

	members := make([]string, 200)
	for i := range members {
		members[i] = fmt.Sprintf("AA:BB:CC:01:%02X:%02X", i/256, i%256)
	}

	stats, err := AddMembersWithXdas("cass-outage-tag", members, "", TagTypeLegacy)
	assert.Error(t, err)
	assert.Equal(t, 200, stats.XdasOk)
	assert.Equal(t, 200, stats.CassandraFail)
	assert.Contains(t, stats.FirstError, "cassandra unavailable")

	entries := hook.AllEntries()
	assert.Less(t, len(entries), 10, "a Cassandra outage must not emit one error line per bucket")
}

// The stored first error is capped so oversized diagnostics cannot blow up
// log fields or API error responses.
func TestErrorAggregator_TruncatesFirstError(t *testing.T) {
	agg := &errorAggregator{}
	agg.add(errors.New(strings.Repeat("x", 10*maxFirstErrorLen)))

	_, firstError := agg.summary()
	assert.LessOrEqual(t, len(firstError), maxFirstErrorLen+len("...(truncated)"))
	assert.Contains(t, firstError, "...(truncated)")
}

func TestErrorAggregator_Concurrent(t *testing.T) {
	agg := &errorAggregator{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			agg.add(fmt.Errorf("error %d", n))
		}(i)
	}
	wg.Wait()

	count, firstError := agg.summary()
	assert.Equal(t, 100, count)
	assert.Contains(t, firstError, "error ")
}

// Reverse lookup: the op is stamped on the request-ends line even when the
// lookup itself fails.
func TestGetTagsByMemberHandler_AuditFields(t *testing.T) {
	setupTestEnvironment()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return []map[string]any{}, nil
	})

	req := httptest.NewRequest("GET", "/taggingService/tags/members/AA:BB:CC:DD:EE:01", nil)
	req = mux.SetURLVars(req, map[string]string{common.Member: "AA:BB:CC:DD:EE:01"})
	xw := xwhttp.NewXResponseWriter(httptest.NewRecorder())

	GetTagsByMemberHandler(xw, req)

	fields := xw.Audit()
	// The group service connector is not mocked here, so the handler may fail
	// before setting result counts — but op must be set regardless.
	assert.Equal(t, OpReverseLookup, fields["op"])
}
