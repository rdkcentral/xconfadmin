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
	// TagType is the type the write actually executed as: for untyped requests
	// the stored type adopted by effectiveWriteTagType, which can differ from the
	// route's. Logged so account traffic on an untyped route stays visible.
	TagType string
}

// ReadStats summarizes the Cassandra cost of a read operation.
type ReadStats struct {
	Buckets int
	Queries int
}

// opAudit attaches tagging fields to the framework "request ends" log line via
// XResponseWriter.SetAuditData; audit field names live here and nowhere else.
// With a plain recorder (tests) the methods are logging no-ops but still update
// metrics.
type opAudit struct {
	xw *xwhttp.XResponseWriter
	op string
}

func newOpAudit(w http.ResponseWriter, op string) opAudit {
	a := opAudit{op: op}
	if xw, ok := w.(*xwhttp.XResponseWriter); ok {
		a.xw = xw
	}
	a.set("op", op)
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

// setTagType records the tag type as a log field only. Deliberately not a
// Prometheus label: the vectors are declared []string{"op"}, so adding a label
// would reset every existing series and per-type op values would leave existing
// panels silently under-reporting.
func (a opAudit) setTagType(tagType string) {
	if tagType != TagTypeLegacy {
		a.set("tag_type", tagType)
	}
}

func (a opAudit) setWriteStats(s WriteStats) {
	// Overwrites the handler's requested type; skipped for legacy, so an error
	// before type resolution keeps the handler's value.
	a.setTagType(s.TagType)
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

// errorAggregator collapses errors from concurrent operations (per-member XDAS
// calls, per-bucket Cassandra batches) into a count plus the first message,
// replacing per-item log lines that flooded the logs during outages.
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

// logDivergence marks the stores diverging: XDAS accepted an operation but
// Cassandra failed. Reconciliation tooling consumes these lines — keep the
// message text stable.
func logDivergence(op string, tagId string, err error) {
	divergenceEvents.Inc()
	log.WithFields(log.Fields{
		"op":  op,
		"tag": tagId,
	}).Errorf("Critical: XDAS succeeded but Cassandra failed: %v", err)
}
