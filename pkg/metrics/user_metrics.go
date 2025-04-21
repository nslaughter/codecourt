// Package metrics provides standardized Prometheus metrics instrumentation
// for all CodeCourt services.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// User service specific metrics
var (
	// AuthenticationTotal counts the total number of authentication attempts
	AuthenticationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "user",
			Name:      "authentication_total",
			Help:      "Total number of authentication attempts",
		},
		[]string{"method", "status"},
	)

	// UserOperationsTotal counts the total number of user operations
	UserOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "user",
			Name:      "operations_total",
			Help:      "Total number of user operations",
		},
		[]string{"operation", "status"},
	)

	// TokenOperationsTotal counts the total number of token operations
	TokenOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "user",
			Name:      "token_operations_total",
			Help:      "Total number of token operations",
		},
		[]string{"operation", "status"},
	)

	// ActiveUserCount tracks the current number of active users
	ActiveUserCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "codecourt",
			Subsystem: "user",
			Name:      "active_count",
			Help:      "Current number of active users",
		},
	)

	// UserSessionDuration observes the duration of user sessions
	UserSessionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Subsystem: "user",
			Name:      "session_duration_seconds",
			Help:      "Duration of user sessions in seconds",
			Buckets:   []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400},
		},
	)
)

// RecordAuthentication records an authentication attempt
func RecordAuthentication(method, status string) {
	AuthenticationTotal.WithLabelValues(method, status).Inc()
}

// RecordUserOperation records a user operation
func RecordUserOperation(operation, status string) {
	UserOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordTokenOperation records a token operation
func RecordTokenOperation(operation, status string) {
	TokenOperationsTotal.WithLabelValues(operation, status).Inc()
}

// SetActiveUserCount sets the current number of active users
func SetActiveUserCount(count int) {
	ActiveUserCount.Set(float64(count))
}

// ObserveUserSessionDuration observes the duration of a user session
func ObserveUserSessionDuration(durationSeconds float64) {
	UserSessionDuration.Observe(durationSeconds)
}
