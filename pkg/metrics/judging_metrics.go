// Package metrics provides standardized Prometheus metrics instrumentation
// for all CodeCourt services.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Judging service specific metrics
var (
	// JudgingTotal counts the total number of judging operations
	JudgingTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "judging",
			Name:      "operations_total",
			Help:      "Total number of judging operations",
		},
		[]string{"language", "status", "problem_id"},
	)

	// JudgingDuration observes the time taken for judging operations
	JudgingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Subsystem: "judging",
			Name:      "duration_seconds",
			Help:      "Time taken for judging operations",
			Buckets:   []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"language", "problem_id"},
	)

	// CodeExecutionMemoryUsage tracks memory usage during code execution
	CodeExecutionMemoryUsage = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Subsystem: "judging",
			Name:      "memory_usage_bytes",
			Help:      "Memory usage during code execution",
			Buckets:   prometheus.ExponentialBuckets(1024*1024, 2, 10), // Start at 1MB with 10 buckets
		},
		[]string{"language", "problem_id"},
	)

	// CodeExecutionCPUUsage tracks CPU usage during code execution
	CodeExecutionCPUUsage = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Subsystem: "judging",
			Name:      "cpu_usage_seconds",
			Help:      "CPU time used during code execution",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
		},
		[]string{"language", "problem_id"},
	)

	// TestCaseResults counts test case results
	TestCaseResults = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "judging",
			Name:      "test_case_results_total",
			Help:      "Total number of test case results",
		},
		[]string{"problem_id", "result"},
	)

	// SecurityViolationsTotal counts security violations during code execution
	SecurityViolationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "judging",
			Name:      "security_violations_total",
			Help:      "Total number of security violations during code execution",
		},
		[]string{"language", "violation_type"},
	)

	// ContainerCreationErrors counts container creation errors
	ContainerCreationErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "judging",
			Name:      "container_creation_errors_total",
			Help:      "Total number of container creation errors",
		},
	)

	// JudgingQueueLength tracks the current length of the judging queue
	JudgingQueueLength = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "codecourt",
			Subsystem: "judging",
			Name:      "queue_length",
			Help:      "Current length of the judging queue",
		},
	)
)

// RecordJudgingOperation records a judging operation
func RecordJudgingOperation(language, status, problemID string) {
	JudgingTotal.WithLabelValues(language, status, problemID).Inc()
}

// ObserveJudgingDuration observes the time taken for a judging operation
func ObserveJudgingDuration(language, problemID string, duration float64) {
	JudgingDuration.WithLabelValues(language, problemID).Observe(duration)
}

// ObserveCodeExecutionMemoryUsage observes memory usage during code execution
func ObserveCodeExecutionMemoryUsage(language, problemID string, memoryBytes float64) {
	CodeExecutionMemoryUsage.WithLabelValues(language, problemID).Observe(memoryBytes)
}

// ObserveCodeExecutionCPUUsage observes CPU usage during code execution
func ObserveCodeExecutionCPUUsage(language, problemID string, cpuSeconds float64) {
	CodeExecutionCPUUsage.WithLabelValues(language, problemID).Observe(cpuSeconds)
}

// RecordTestCaseResult records a test case result
func RecordTestCaseResult(problemID, result string) {
	TestCaseResults.WithLabelValues(problemID, result).Inc()
}

// RecordSecurityViolation records a security violation during code execution
func RecordSecurityViolation(language, violationType string) {
	SecurityViolationsTotal.WithLabelValues(language, violationType).Inc()
}

// IncrementContainerCreationErrors increments the container creation errors counter
func IncrementContainerCreationErrors() {
	ContainerCreationErrors.Inc()
}

// SetJudgingQueueLength sets the current length of the judging queue
func SetJudgingQueueLength(length int) {
	JudgingQueueLength.Set(float64(length))
}
