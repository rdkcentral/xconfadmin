package tag

import (
	"sync"

	ds "github.com/rdkcentral/xconfwebconfig/db"

	"github.com/prometheus/client_golang/prometheus"
)

// Tagging-specific metrics. Request counts/durations per endpoint and XDAS
// call outcomes are NOT duplicated here — they are already covered by the
// app-wide api_requests_total / api_request_duration_seconds middleware and
// by external_api_count fed from the HTTP client for every XDAS call.
var (
	divergenceEvents = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tagging_divergence_events_total",
		Help: "Times XDAS accepted an operation but the Cassandra side failed (stores diverged).",
	})
	cassandraFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tagging_cassandra_failures_total",
		Help: "Tag members that failed to be written to or removed from Cassandra.",
	}, []string{"op"})
	membersProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tagging_members_processed_total",
		Help: "Tag members successfully written or removed per operation.",
	}, []string{"op"})
	readQueriesPerRequest = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tagging_read_queries_per_request",
		Help:    "Cassandra queries issued per tagging read request (fan-out cost).",
		Buckets: []float64{1, 2, 5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"op"})
	concurrentQueriesInUse = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "xconf_db_concurrent_queries_in_use",
		Help: "Occupancy of the Cassandra client's ConcurrentQueries semaphore.",
	}, func() float64 {
		if c, ok := ds.GetDatabaseClient().(*ds.CassandraClient); ok && c != nil {
			return float64(len(c.ConcurrentQueries))
		}
		return 0
	})
)

var registerMetricsOnce sync.Once

// RegisterTaggingMetrics registers the tagging metrics with the default
// Prometheus registry (the same one the app-wide metrics use). Called from
// the tagging router setup; safe to call more than once.
func RegisterTaggingMetrics() {
	registerMetricsOnce.Do(func() {
		prometheus.MustRegister(divergenceEvents, cassandraFailures,
			membersProcessed, readQueriesPerRequest, concurrentQueriesInUse)
	})
}
