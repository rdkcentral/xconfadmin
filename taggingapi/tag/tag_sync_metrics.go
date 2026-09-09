package tag

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	tagSyncMetricsOnce sync.Once

	tagSyncCheckedTotal        prometheus.Counter
	tagSyncPushedTotal         *prometheus.CounterVec
	tagSyncXdasErrorsTotal     *prometheus.CounterVec
	tagSyncBreakerTrippedTotal *prometheus.CounterVec
	tagSyncMissingRate         prometheus.Gauge
	tagSyncRunningGauge        prometheus.Gauge
	tagSyncRunDurationSeconds  prometheus.Gauge
)

// initTagSyncMetrics registers the tag sync metrics on the default registry
// (exposed by the existing /metrics endpoint). Tag ids are deliberately not
// used as labels; per-tag numbers go to logs and the run report instead.
func initTagSyncMetrics() {
	tagSyncMetricsOnce.Do(func() {
		tagSyncCheckedTotal = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "tagging_sync_members_checked_total",
			Help: "Members checked against XDAS by the tag sync job",
		})
		tagSyncPushedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tagging_sync_pushed_total",
			Help: "Members pushed to XDAS by the tag sync job, by reason",
		}, []string{"reason"})
		tagSyncXdasErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tagging_sync_xdas_errors_total",
			Help: "XDAS errors seen by the tag sync job, by operation",
		}, []string{"op"})
		tagSyncBreakerTrippedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tagging_sync_breaker_tripped_total",
			Help: "Times the tag sync circuit breaker aborted a run, by reason",
		}, []string{"reason"})
		tagSyncMissingRate = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tagging_sync_missing_rate",
			Help: "Missing/checked ratio observed by the most recent tag sync run",
		})
		tagSyncRunningGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tagging_sync_running",
			Help: "1 while a tag sync run is in progress on this instance",
		})
		tagSyncRunDurationSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tagging_sync_run_duration_seconds",
			Help: "Duration of the most recent tag sync run on this instance",
		})
		prometheus.MustRegister(
			tagSyncCheckedTotal,
			tagSyncPushedTotal,
			tagSyncXdasErrorsTotal,
			tagSyncBreakerTrippedTotal,
			tagSyncMissingRate,
			tagSyncRunningGauge,
			tagSyncRunDurationSeconds,
		)
	})
}
