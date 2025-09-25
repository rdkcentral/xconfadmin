package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// AppMetrics just collects all the needed metrics
type AppMetrics struct {
	counter                  *prometheus.CounterVec
	duration                 *prometheus.HistogramVec
	extAPIDuration           *prometheus.HistogramVec
	extAPICounts             *prometheus.CounterVec
	inFlight                 prometheus.Gauge
	responseSize             *prometheus.HistogramVec
	requestSize              *prometheus.HistogramVec
	reqsByModCounter         *prometheus.CounterVec
	reqsByOriginCounter      *prometheus.CounterVec
	cacheSyncCounter         *prometheus.CounterVec
	cacheRefreshAllCounter   *prometheus.CounterVec
	cacheRefreshTableCounter *prometheus.CounterVec
}

var metrics *AppMetrics

// NewMetrics creates all the metrics needed for xconfadmin
func NewMetrics() *AppMetrics {
	metrics = &AppMetrics{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_requests_total",
				Help: "A counter for total number of requests.",
			},
			[]string{"app", "code", "method", "path", "app_type"}, // app name, status code, http method, request URL, applictionType
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_request_duration_seconds",
				Help:    "A histogram of latencies for requests.",
				Buckets: []float64{.01, .02, .05, 0.1, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"app", "code", "method", "path"},
		),
		extAPICounts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "external_api_count",
				Help: "A counter for external API calls",
			},
			[]string{"app", "code", "method", "service"}, // app name, status code, http method, extService
		),
		extAPIDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "external_api_request_duration_seconds",
				Help:    "A histogram of latencies for requests.",
				Buckets: []float64{.01, .02, .05, 0.1, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"app", "code", "method", "service"},
		),
		inFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "in_flight_requests",
				Help: "A gauge of requests currently being served.",
			},
		),
		requestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "request_size_bytes",
				Help:    "A histogram of request sizes for requests.",
				Buckets: []float64{200, 500, 1000, 10000, 100000},
			},
			[]string{"app"},
		),
		responseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "response_size_bytes",
				Help:    "A histogram of response sizes for requests.",
				Buckets: []float64{200, 500, 1000, 10000, 100000},
			},
			[]string{"app"},
		),
		reqsByModCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "requests_by_module",
				Help: "A counter for total number of requests group by module",
			},
			[]string{"app", "module", "code"}, // app, module, status code
		),
		reqsByOriginCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "requests_by_origin",
				Help: "A counter for total number of requests group by origin",
			},
			[]string{"app", "origin", "code"}, // app, origin, status code
		),
		cacheSyncCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_sync_total",
				Help: "A counter for total number of cache sync status",
			},
			[]string{"app", "status"}, // app name, status (success or failure)
		),
		cacheRefreshAllCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_refresh_all_total",
				Help: "A counter for total number of cache refresh all status",
			},
			[]string{"app", "status"}, // app name, status (success or failure)
		),
		cacheRefreshTableCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_refresh_table_total",
				Help: "A counter for total number of table cache refresh status",
			},
			[]string{"app", "table", "status"}, // app name, table name, status (success or failure)
		),
	}
	prometheus.MustRegister(metrics.inFlight, metrics.counter, metrics.duration,
		metrics.extAPICounts, metrics.extAPIDuration,
		metrics.responseSize, metrics.requestSize,
		metrics.reqsByModCounter, metrics.reqsByOriginCounter,
		metrics.cacheSyncCounter, metrics.cacheRefreshAllCounter, metrics.cacheRefreshTableCounter)
	return metrics
}

// WebMetrics updates infligh, reqSize and respSize metrics
func (s *WebconfigServer) WebMetrics(m *AppMetrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		promhttp.InstrumentHandlerInFlight(m.inFlight,
			promhttp.InstrumentHandlerRequestSize(m.requestSize.MustCurryWith(prometheus.Labels{"app": s.AppName}),
				promhttp.InstrumentHandlerResponseSize(m.responseSize.MustCurryWith(prometheus.Labels{"app": s.AppName}), next),
			),
		).ServeHTTP(w, r)
	})
}

func (m *AppMetrics) MetricsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		promhttp.InstrumentHandlerInFlight(m.inFlight,
			promhttp.InstrumentHandlerRequestSize(m.requestSize.MustCurryWith(prometheus.Labels{"app": WebConfServer.AppName}),
				promhttp.InstrumentHandlerResponseSize(m.responseSize.MustCurryWith(prometheus.Labels{"app": WebConfServer.AppName}), next),
			),
		).ServeHTTP(w, r)
	})
}

// updateMetrics updates api_req_total, number of API calls
func (s *AppMetrics) UpdateAPIMetrics(r *http.Request, status int, startTime time.Time) {
	if metrics == nil {
		// Metrics may not be initialized in tests, or disabled by a config flag
		return
	}

	route := mux.CurrentRoute(r)
	if route == nil {
		// Paranoia, the code should never come here
		return
	}

	statusCode := strconv.Itoa(status)

	var path string
	var err error
	// Piggyback on mux's regex matching
	if path, err = route.GetPathTemplate(); err != nil {
		log.Debug(fmt.Sprintf("mux GetPathTemplate err in metrics %+v", err.Error()))
		path = "path extraction error"
	}
	queryParams := r.URL.Query()
	qp := queryParams["applicationType"]
	appType := ""
	if len(qp) != 0 {
		appType = qp[0]
	}
	vals := prometheus.Labels{"app": WebConfServer.AppName, "code": statusCode, "method": r.Method, "path": path, "app_type": appType}
	metrics.counter.With(vals).Inc()

	vals = prometheus.Labels{"app": WebConfServer.AppName, "code": statusCode, "method": r.Method, "path": path}
	metrics.duration.With(vals).Observe(time.Since(startTime).Seconds())

	module := route.GetName()
	vals = prometheus.Labels{"app": WebConfServer.AppName, "module": module, "code": statusCode}
	metrics.reqsByModCounter.With(vals).Inc()

	origin := r.Header.Get("Origin")
	vals = prometheus.Labels{"app": WebConfServer.AppName, "origin": origin, "code": statusCode}
	metrics.reqsByOriginCounter.With(vals).Inc()
}

// updateExternalAPIMetrics updates duration and counts for external API calls to titan, sat etc.
func (s *AppMetrics) UpdateExternalAPIMetrics(service string, method string, statusCode int, startTime time.Time) {
	if metrics == nil {
		// Metrics may not be initialized in tests, or disabled by a config flag
		return
	}
	statusStr := strconv.Itoa(statusCode)
	vals := prometheus.Labels{"app": AppName(), "code": statusStr, "method": method, "service": service}
	metrics.extAPICounts.With(vals).Inc()

	externalCallDuration := time.Since(startTime).Seconds()
	metrics.extAPIDuration.With(vals).Observe(externalCallDuration)
}

// func (s *WebconfigServer) UpdateCacheSyncMetrics(success bool) {
// 	if metrics == nil {
// 		// Metrics may not be initialized in tests, or disabled by a config flag
// 		return
// 	}
// 	var statusStr string
// 	if success {
// 		statusStr = "success"
// 	} else {
// 		statusStr = "failure"
// 	}
// 	vals := prometheus.Labels{"app": s.AppName, "status": statusStr}
// 	metrics.cacheSyncCounter.With(vals).Inc()
// }

// func (s *WebconfigServer) UpdateCacheRefreshAllMetrics(success bool) {
// 	if metrics == nil {
// 		// Metrics may not be initialized in tests, or disabled by a config flag
// 		return
// 	}
// 	var statusStr string
// 	if success {
// 		statusStr = "success"
// 	} else {
// 		statusStr = "failure"
// 	}
// 	vals := prometheus.Labels{"app": s.AppName, "status": statusStr}
// 	metrics.cacheRefreshAllCounter.With(vals).Inc()
// }

// func (s *WebconfigServer) UpdateCacheRefreshTableMetrics(tableName string, success bool) {
// 	if metrics == nil {
// 		// Metrics may not be initialized in tests, or disabled by a config flag
// 		return
// 	}
// 	var statusStr string
// 	if success {
// 		statusStr = "success"
// 	} else {
// 		statusStr = "failure"
// 	}
// 	vals := prometheus.Labels{"app": s.AppName, "table": tableName, "status": statusStr}
// 	metrics.cacheRefreshTableCounter.With(vals).Inc()
// }
