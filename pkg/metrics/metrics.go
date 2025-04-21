// Package metrics provides standardized Prometheus metrics instrumentation
// for all CodeCourt services.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Standard metrics used across all services
var (
	// HTTPRequestsTotal counts the total number of HTTP requests processed
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"service", "method", "endpoint", "status"},
	)

	// HTTPRequestDuration observes the HTTP request duration in seconds
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.001, 0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "endpoint"},
	)

	// DatabaseOperationsTotal counts database operations
	DatabaseOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Name:      "database_operations_total",
			Help:      "Total number of database operations",
		},
		[]string{"service", "operation", "table", "status"},
	)

	// DatabaseOperationDuration observes the database operation duration in seconds
	DatabaseOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Name:      "database_operation_duration_seconds",
			Help:      "Database operation duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2.5},
		},
		[]string{"service", "operation", "table"},
	)

	// KafkaMessagesTotal counts Kafka messages
	KafkaMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Name:      "kafka_messages_total",
			Help:      "Total number of Kafka messages",
		},
		[]string{"service", "topic", "operation"},
	)

	// KafkaOperationDuration observes the Kafka operation duration in seconds
	KafkaOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Name:      "kafka_operation_duration_seconds",
			Help:      "Kafka operation duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"service", "topic", "operation"},
	)

	// ServiceInfoGauge provides service version and build information
	ServiceInfoGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "codecourt",
			Name:      "service_info",
			Help:      "Service version and build information",
		},
		[]string{"service", "version", "build_date", "commit_hash"},
	)
)

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Status returns the HTTP status code
func (rw *responseWriter) Status() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}

// MetricsMiddleware captures HTTP request metrics
func MetricsMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response wrapper to capture status code
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     0,
			}

			// Call the next handler
			next.ServeHTTP(rw, r)

			// Record metrics after the request is processed
			duration := time.Since(start).Seconds()
			endpoint := r.URL.Path
			status := strconv.Itoa(rw.Status())

			// Increment request counter with labels
			HTTPRequestsTotal.WithLabelValues(serviceName, r.Method, endpoint, status).Inc()

			// Observe request duration
			HTTPRequestDuration.WithLabelValues(serviceName, r.Method, endpoint).Observe(duration)
		})
	}
}

// SetupMetricsEndpoint registers the /metrics endpoint
func SetupMetricsEndpoint(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
}

// RegisterServiceInfo registers service information metrics
func RegisterServiceInfo(serviceName, version, buildDate, commitHash string) {
	ServiceInfoGauge.WithLabelValues(serviceName, version, buildDate, commitHash).Set(1)
}

// RecordDatabaseOperation records a database operation metric
func RecordDatabaseOperation(service, operation, table, status string) {
	DatabaseOperationsTotal.WithLabelValues(service, operation, table, status).Inc()
}

// ObserveDatabaseOperationDuration observes the duration of a database operation
func ObserveDatabaseOperationDuration(service, operation, table string, duration float64) {
	DatabaseOperationDuration.WithLabelValues(service, operation, table).Observe(duration)
}

// RecordKafkaMessage records a Kafka message metric
func RecordKafkaMessage(service, topic, operation string) {
	KafkaMessagesTotal.WithLabelValues(service, topic, operation).Inc()
}

// ObserveKafkaOperationDuration observes the duration of a Kafka operation
func ObserveKafkaOperationDuration(service, topic, operation string, duration float64) {
	KafkaOperationDuration.WithLabelValues(service, topic, operation).Observe(duration)
}
