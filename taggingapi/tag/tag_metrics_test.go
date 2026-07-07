package tag

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// The /metrics endpoint (main.go: promhttp.Handler()) serves the default
// Prometheus registry. This test proves the tagging metrics land in that
// registry and are gatherable — i.e. they will appear on /metrics without
// any further wiring.
func TestTaggingMetricsGatherableFromDefaultRegistry(t *testing.T) {
	RegisterTaggingMetrics()
	RegisterTaggingMetrics() // idempotent — must not panic on double registration

	divergenceEvents.Inc()
	membersProcessed.WithLabelValues(OpAddMembers).Add(5)
	cassandraFailures.WithLabelValues(OpRemoveMembers).Inc()
	readQueriesPerRequest.WithLabelValues(OpGetMembersPage).Observe(3)

	families, err := prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)

	found := map[string]bool{}
	for _, mf := range families {
		found[mf.GetName()] = true
	}
	for _, name := range []string{
		"tagging_divergence_events_total",
		"tagging_cassandra_failures_total",
		"tagging_members_processed_total",
		"tagging_read_queries_per_request",
		"xconf_db_concurrent_queries_in_use",
	} {
		assert.True(t, found[name], "expected %s on the default registry (served by /metrics)", name)
	}
}
