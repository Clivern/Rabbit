// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/sloth/conf"
	"github.com/clivern/sloth/db"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request sizes in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"method", "path"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response sizes in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"method", "path", "status"},
	)

	httpSlowRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_slow_requests_total",
			Help: "HTTP requests that exceeded the slow-request latency threshold",
		},
		[]string{"method", "path", "status"},
	)

	_ = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_users_total",
			Help: "Current number of users in database.",
		},
		func() float64 {
			conn := db.GetDB()
			if conn == nil {
				return 0
			}

			count, err := db.NewUserRepository(conn).Count()
			if err != nil {
				return 0
			}

			return float64(count)
		},
	)

	_ = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_workspaces_total",
			Help: "Current number of workspaces in database.",
		},
		func() float64 {
			conn := db.GetDB()
			if conn == nil {
				return 0
			}

			count, err := db.NewWorkspaceRepository(conn).CountAll()
			if err != nil {
				return 0
			}

			return float64(count)
		},
	)

	_ = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_user_sessions_total",
			Help: "Current number of non-expired user sessions in database.",
		},
		func() float64 {
			conn := db.GetDB()
			if conn == nil {
				return 0
			}

			count, err := db.NewSessionRepository(conn).Count()
			if err != nil {
				return 0
			}

			return float64(count)
		},
	)

	_ = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_invites_total",
			Help: "Current number of user invites in database.",
		},
		func() float64 {
			conn := db.GetDB()
			if conn == nil {
				return 0
			}

			count, err := db.NewUserInviteRepository(conn).Count()
			if err != nil {
				return 0
			}

			return float64(count)
		},
	)

	_ = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_access_keys_total",
			Help: "Current number of non-expired workspace access keys in database.",
		},
		func() float64 {
			conn := db.GetDB()
			if conn == nil {
				return 0
			}

			count, err := db.NewAccessKeyRepository(conn).Count()
			if err != nil {
				return 0
			}

			return float64(count)
		},
	)

	_ = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_api_keys_total",
			Help: "Current number of non-expired user API keys in database.",
		},
		func() float64 {
			conn := db.GetDB()
			if conn == nil {
				return 0
			}

			count, err := db.NewAPIKeyRepository(conn).Count()
			if err != nil {
				return 0
			}

			return float64(count)
		},
	)

	_ = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "db_async_tasks_total",
			Help:        "Current number of async tasks by status in database.",
			ConstLabels: prometheus.Labels{"status": "pending"},
		},
		func() float64 {
			conn := db.GetDB()
			if conn == nil {
				return 0
			}

			count, err := db.NewAsyncTaskRepository(conn).CountByStatus("pending")
			if err != nil {
				return 0
			}

			return float64(count)
		},
	)

	_ = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "db_async_tasks_total",
			Help:        "Current number of async tasks by status in database.",
			ConstLabels: prometheus.Labels{"status": "running"},
		},
		func() float64 {
			conn := db.GetDB()
			if conn == nil {
				return 0
			}

			count, err := db.NewAsyncTaskRepository(conn).CountByStatus("running")
			if err != nil {
				return 0
			}

			return float64(count)
		},
	)

	_ = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "db_async_tasks_total",
			Help:        "Current number of async tasks by status in database.",
			ConstLabels: prometheus.Labels{"status": "completed"},
		},
		func() float64 {
			conn := db.GetDB()
			if conn == nil {
				return 0
			}

			count, err := db.NewAsyncTaskRepository(conn).CountByStatus("completed")
			if err != nil {
				return 0
			}

			return float64(count)
		},
	)

	_ = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "db_async_tasks_total",
			Help:        "Current number of async tasks by status in database.",
			ConstLabels: prometheus.Labels{"status": "failed"},
		},
		func() float64 {
			conn := db.GetDB()
			if conn == nil {
				return 0
			}

			count, err := db.NewAsyncTaskRepository(conn).CountByStatus("failed")
			if err != nil {
				return 0
			}

			return float64(count)
		},
	)
)

// MetricsResponseWriter wraps the response to capture status and size for metrics.
type MetricsResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	size        int
	wroteHeader bool
}

// WriteHeader captures the HTTP status code for logging.
func (m *MetricsResponseWriter) WriteHeader(code int) {
	if !m.wroteHeader {
		m.statusCode = code
		m.ResponseWriter.WriteHeader(code)
		m.wroteHeader = true
	}
}

// Write captures the response body size for logging.
func (m *MetricsResponseWriter) Write(b []byte) (int, error) {
	n, err := m.ResponseWriter.Write(b)
	m.size += n

	return n, err
}

// Unwrap returns the wrapped response, so http.ResponseController can flush streamed responses through it.
func (m *MetricsResponseWriter) Unwrap() http.ResponseWriter {
	return m.ResponseWriter
}

// PrometheusMiddleware creates a middleware for Prometheus metrics
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/public/_metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now().UTC()
		path := r.URL.Path
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			path = "/assets/*"
		}

		reqSize := float64(r.ContentLength)
		if reqSize > 0 {
			httpRequestSize.WithLabelValues(r.Method, path).Observe(reqSize)
		}

		wrapped := &MetricsResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		elapsed := time.Since(start)
		duration := elapsed.Seconds()
		status := strconv.Itoa(wrapped.statusCode)

		httpRequestsTotal.WithLabelValues(
			r.Method,
			path,
			status,
		).Inc()
		httpRequestDuration.WithLabelValues(
			r.Method,
			path,
			status,
		).Observe(duration)
		httpResponseSize.WithLabelValues(
			r.Method,
			path,
			status,
		).Observe(float64(wrapped.size))

		if elapsed >= conf.SlowRequestThreshold {
			httpSlowRequestsTotal.WithLabelValues(
				r.Method,
				path,
				status,
			).Inc()
		}
	})
}
