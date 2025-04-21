// Package metrics provides standardized Prometheus metrics instrumentation
// for all CodeCourt services.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Submission service specific metrics
var (
	// SubmissionsTotal counts the total number of code submissions
	SubmissionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "submission",
			Name:      "submissions_total",
			Help:      "Total number of code submissions",
		},
		[]string{"language", "status", "problem_id"},
	)

	// SubmissionProcessingTime observes the time taken to process submissions
	SubmissionProcessingTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Subsystem: "submission",
			Name:      "processing_seconds",
			Help:      "Time taken to process submissions",
			Buckets:   []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"language", "problem_id"},
	)

	// SubmissionQueueLength tracks the current length of the submission queue
	SubmissionQueueLength = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "codecourt",
			Subsystem: "submission",
			Name:      "queue_length",
			Help:      "Current length of the submission queue",
		},
	)

	// SubmissionsByUser counts submissions by user
	SubmissionsByUser = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "submission",
			Name:      "by_user_total",
			Help:      "Total number of submissions by user",
		},
		[]string{"user_id"},
	)
)

// RecordSubmission records a new submission
func RecordSubmission(language, status, problemID, userID string) {
	SubmissionsTotal.WithLabelValues(language, status, problemID).Inc()
	SubmissionsByUser.WithLabelValues(userID).Inc()
}

// ObserveSubmissionProcessingTime observes the time taken to process a submission
func ObserveSubmissionProcessingTime(language, problemID string, duration float64) {
	SubmissionProcessingTime.WithLabelValues(language, problemID).Observe(duration)
}

// SetSubmissionQueueLength sets the current length of the submission queue
func SetSubmissionQueueLength(length int) {
	SubmissionQueueLength.Set(float64(length))
}
