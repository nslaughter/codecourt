// Package metrics provides standardized Prometheus metrics instrumentation
// for all CodeCourt services.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Problem service specific metrics
var (
	// ProblemOperationsTotal counts the total number of problem operations
	ProblemOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "problem",
			Name:      "operations_total",
			Help:      "Total number of problem operations",
		},
		[]string{"operation", "status"},
	)

	// ProblemAccessTotal counts the total number of problem accesses
	ProblemAccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "problem",
			Name:      "access_total",
			Help:      "Total number of problem accesses",
		},
		[]string{"problem_id", "difficulty"},
	)

	// ProblemCount tracks the current number of problems by difficulty
	ProblemCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "codecourt",
			Subsystem: "problem",
			Name:      "count",
			Help:      "Current number of problems by difficulty",
		},
		[]string{"difficulty", "category"},
	)

	// TestCaseCount tracks the current number of test cases
	TestCaseCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "codecourt",
			Subsystem: "problem",
			Name:      "test_case_count",
			Help:      "Current number of test cases",
		},
		[]string{"problem_id"},
	)

	// ProblemSuccessRate tracks the success rate for problems
	ProblemSuccessRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "codecourt",
			Subsystem: "problem",
			Name:      "success_rate",
			Help:      "Success rate for problems (percentage)",
		},
		[]string{"problem_id", "difficulty"},
	)
)

// RecordProblemOperation records a problem operation
func RecordProblemOperation(operation, status string) {
	ProblemOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordProblemAccess records a problem access
func RecordProblemAccess(problemID, difficulty string) {
	ProblemAccessTotal.WithLabelValues(problemID, difficulty).Inc()
}

// SetProblemCount sets the current number of problems by difficulty
func SetProblemCount(difficulty, category string, count int) {
	ProblemCount.WithLabelValues(difficulty, category).Set(float64(count))
}

// SetTestCaseCount sets the current number of test cases for a problem
func SetTestCaseCount(problemID string, count int) {
	TestCaseCount.WithLabelValues(problemID).Set(float64(count))
}

// SetProblemSuccessRate sets the success rate for a problem
func SetProblemSuccessRate(problemID, difficulty string, rate float64) {
	ProblemSuccessRate.WithLabelValues(problemID, difficulty).Set(rate)
}
