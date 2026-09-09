package tag

import (
	"net/http"
	"sync"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	log "github.com/sirupsen/logrus"
)

// Operation names used in the "op" audit field and metric labels.
// Stable values — log queries and dashboards depend on them.
const (
	OpAddMembers          = "add_members"
	OpRemoveMembers       = "remove_members"
	OpRemoveMember        = "remove_member"
	OpGetMembersPage      = "get_members_page"
	OpGetMembersFull      = "get_members_full"
	OpGetTag              = "get_tag"
	OpGetAllTags          = "get_all_tags"
	OpDeleteTag           = "delete_tag"
	OpReverseLookup       = "reverse_lookup"
	OpReverseLookupValues = "reverse_lookup_values"
)

// WriteStats summarizes a write operation across both stores. Returned by the
// *WithXdas functions so handlers can attach the counts to the request log.
type WriteStats struct {
	Requested     int
	XdasOk        int
	XdasFail      int
	CassandraOk   int
	CassandraFail int
	Buckets       int
	FirstError    string
}

// ReadStats summarizes the Cassandra cost of a read operation.
type ReadStats struct {
	Buckets int
	Queries int
}

// opAudit attaches tagging fields to the framework "request ends" log line via
// XResponseWriter.SetAuditData. Audit field names are defined here and nowhere
// else. All methods are safe to call when the writer is not an XResponseWriter
// (plain recorders in tests): they become no-ops for logging but still update
// metrics.
type opAudit struct {
	xw *xwhttp.XResponseWriter
	op string
}

func newOpAudit(w http.ResponseWriter, op string, tenantId string) opAudit {
	a := opAudit{op: op}
	if xw, ok := w.(*xwhttp.XResponseWriter); ok {
		a.xw = xw
	}
	a.set("op", op)
	a.set("tenant", tenantId)
	return a
}

func (a opAudit) set(key string, value interface{}) {
	if a.xw != nil {
		a.xw.SetAuditData(key, value)
	}
}

func (a opAudit) setTag(tagId string) {
	a.set("tag", tagId)
}

func (a opAudit) setWriteStats(s WriteStats) {
	a.set("requested", s.Requested)
	a.set("xdas_ok", s.XdasOk)
	a.set("xdas_fail", s.XdasFail)
	a.set("cassandra_ok", s.CassandraOk)
	a.set("cassandra_fail", s.CassandraFail)
	a.set("buckets", s.Buckets)
	if s.FirstError != "" {
		a.set("first_error", s.FirstError)
	}
	membersProcessed.WithLabelValues(a.op).Add(float64(s.CassandraOk))
	if s.CassandraFail > 0 {
		cassandraFailures.WithLabelValues(a.op).Add(float64(s.CassandraFail))
	}
}

func (a opAudit) setReadStats(s ReadStats) {
	a.set("buckets", s.Buckets)
	a.set("queries", s.Queries)
	readQueriesPerRequest.WithLabelValues(a.op).Observe(float64(s.Queries))
}

// maxFirstErrorLen bounds the stored first-error message. Cassandra and XDAS
// errors can embed long diagnostics, and the value flows into log fields and
// API error responses.
const maxFirstErrorLen = 256

// errorAggregator aggregates errors from concurrent operations (per-member
// XDAS calls, per-bucket Cassandra batches) into a total count plus the first
// error message as a representative sample. Replaces per-item error log
// lines, which flooded the logs during outages.
type errorAggregator struct {
	mu         sync.Mutex
	total      int
	firstError string
}

func (a *errorAggregator) add(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.total++
	if a.firstError == "" {
		msg := err.Error()
		if len(msg) > maxFirstErrorLen {
			msg = msg[:maxFirstErrorLen] + "...(truncated)"
		}
		a.firstError = msg
	}
}

// summary returns the error count and the first error message.
func (a *errorAggregator) summary() (int, string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.total, a.firstError
}

// logDivergence marks the stores diverging: XDAS accepted an operation but the
// Cassandra side failed. These lines are the input for the topic-4
// reconciliation work — keep the message text stable.
func logDivergence(op string, tenantId string, tagId string, err error) {
	divergenceEvents.Inc()
	log.WithFields(log.Fields{
		"op":     op,
		"tenant": tenantId,
		"tag":    tagId,
	}).Errorf("Critical: XDAS succeeded but Cassandra failed: %v", err)
}
