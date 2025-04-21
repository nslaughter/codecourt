package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMetricsMiddleware(t *testing.T) {
	// Reset the registry to avoid conflicts with other tests
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	
	// Re-register our metrics with the new registry
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"service", "method", "endpoint", "status"},
	)
	
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.001, 0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "endpoint"},
	)

	// Define test cases using table-driven style
	tests := []struct {
		name           string
		method         string
		path           string
		statusCode     int
		serviceName    string
		expectedLabels map[string]string
	}{
		{
			name:       "GET request with 200 status",
			method:     "GET",
			path:       "/api/v1/problems",
			statusCode: http.StatusOK,
			serviceName: "test-service",
			expectedLabels: map[string]string{
				"service":  "test-service",
				"method":   "GET",
				"endpoint": "/api/v1/problems",
				"status":   "200",
			},
		},
		{
			name:       "POST request with 201 status",
			method:     "POST",
			path:       "/api/v1/submissions",
			statusCode: http.StatusCreated,
			serviceName: "test-service",
			expectedLabels: map[string]string{
				"service":  "test-service",
				"method":   "POST",
				"endpoint": "/api/v1/submissions",
				"status":   "201",
			},
		},
		{
			name:       "PUT request with 400 status",
			method:     "PUT",
			path:       "/api/v1/users/1",
			statusCode: http.StatusBadRequest,
			serviceName: "test-service",
			expectedLabels: map[string]string{
				"service":  "test-service",
				"method":   "PUT",
				"endpoint": "/api/v1/users/1",
				"status":   "400",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test handler that returns the specified status code
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Add a small delay to simulate processing time for metrics
				time.Sleep(1 * time.Millisecond)
				w.WriteHeader(tc.statusCode)
				io.WriteString(w, "test response")
			})

			// Apply the metrics middleware
			handler := MetricsMiddleware(tc.serviceName)(testHandler)

			// Create a test request
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rec, req)

			// Verify the response
			if rec.Code != tc.statusCode {
				t.Errorf("expected status code %d, got %d", tc.statusCode, rec.Code)
			}

			// Set up a metrics endpoint
			metricsRec := httptest.NewRecorder()
			promhttp.Handler().ServeHTTP(metricsRec, httptest.NewRequest("GET", "/metrics", nil))

			// Check that the metrics output contains our expected metrics
			metricsOutput := metricsRec.Body.String()
			
			// Verify request counter metric
			expectedMetricName := `codecourt_http_requests_total{endpoint="` + tc.expectedLabels["endpoint"] + 
				`",method="` + tc.expectedLabels["method"] + 
				`",service="` + tc.expectedLabels["service"] + 
				`",status="` + tc.expectedLabels["status"] + `"}`
			
			if !strings.Contains(metricsOutput, expectedMetricName) {
				t.Errorf("metrics output does not contain expected counter metric: %s", expectedMetricName)
			}
			
			// Verify request duration metric (just check that it exists, not the value)
			expectedDurationMetric := `codecourt_http_request_duration_seconds_bucket{endpoint="` + tc.expectedLabels["endpoint"] + 
				`",method="` + tc.expectedLabels["method"] + 
				`",service="` + tc.expectedLabels["service"] + `"`
			
			if !strings.Contains(metricsOutput, expectedDurationMetric) {
				t.Errorf("metrics output does not contain expected duration metric: %s", expectedDurationMetric)
			}
		})
	}
}

func TestServiceInfoRegistration(t *testing.T) {
	// Reset the registry to avoid conflicts with other tests
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	
	// Re-register our metrics with the new registry
	ServiceInfoGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "codecourt",
			Name:      "service_info",
			Help:      "Service version and build information",
		},
		[]string{"service", "version", "build_date", "commit_hash"},
	)

	// Define test cases using table-driven style
	tests := []struct {
		name       string
		service    string
		version    string
		buildDate  string
		commitHash string
	}{
		{
			name:       "API Gateway service info",
			service:    "api-gateway",
			version:    "1.0.0",
			buildDate:  "2025-04-21",
			commitHash: "abc123",
		},
		{
			name:       "User service info",
			service:    "user-service",
			version:    "0.9.0",
			buildDate:  "2025-04-20",
			commitHash: "def456",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Register service info
			RegisterServiceInfo(tc.service, tc.version, tc.buildDate, tc.commitHash)

			// Set up a metrics endpoint
			metricsRec := httptest.NewRecorder()
			promhttp.Handler().ServeHTTP(metricsRec, httptest.NewRequest("GET", "/metrics", nil))

			// Check that the metrics output contains our expected metrics
			metricsOutput := metricsRec.Body.String()
			
			// Verify service info metric
			expectedMetricName := `codecourt_service_info{build_date="` + tc.buildDate + 
				`",commit_hash="` + tc.commitHash + 
				`",service="` + tc.service + 
				`",version="` + tc.version + `"}`
			
			if !strings.Contains(metricsOutput, expectedMetricName) {
				t.Errorf("metrics output does not contain expected service info metric: %s", expectedMetricName)
			}
		})
	}
}

func TestDatabaseMetrics(t *testing.T) {
	// Reset the registry to avoid conflicts with other tests
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	
	// Re-register our metrics with the new registry
	DatabaseOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Name:      "database_operations_total",
			Help:      "Total number of database operations",
		},
		[]string{"service", "operation", "table", "status"},
	)
	
	DatabaseOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Name:      "database_operation_duration_seconds",
			Help:      "Database operation duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2.5},
		},
		[]string{"service", "operation", "table"},
	)

	// Define test cases using table-driven style
	tests := []struct {
		name      string
		service   string
		operation string
		table     string
		status    string
		duration  float64
	}{
		{
			name:      "SELECT operation on users table",
			service:   "user-service",
			operation: "SELECT",
			table:     "users",
			status:    "success",
			duration:  0.015,
		},
		{
			name:      "INSERT operation on problems table",
			service:   "problem-service",
			operation: "INSERT",
			table:     "problems",
			status:    "success",
			duration:  0.025,
		},
		{
			name:      "UPDATE operation on submissions table with error",
			service:   "submission-service",
			operation: "UPDATE",
			table:     "submissions",
			status:    "error",
			duration:  0.005,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Record database operation
			RecordDatabaseOperation(tc.service, tc.operation, tc.table, tc.status)
			
			// Observe database operation duration
			ObserveDatabaseOperationDuration(tc.service, tc.operation, tc.table, tc.duration)

			// Set up a metrics endpoint
			metricsRec := httptest.NewRecorder()
			promhttp.Handler().ServeHTTP(metricsRec, httptest.NewRequest("GET", "/metrics", nil))

			// Check that the metrics output contains our expected metrics
			metricsOutput := metricsRec.Body.String()
			
			// Verify database operation counter metric
			expectedCounterMetric := `codecourt_database_operations_total{operation="` + tc.operation + 
				`",service="` + tc.service + 
				`",status="` + tc.status + 
				`",table="` + tc.table + `"}`
			
			if !strings.Contains(metricsOutput, expectedCounterMetric) {
				t.Errorf("metrics output does not contain expected database counter metric: %s", expectedCounterMetric)
			}
			
			// Verify database duration metric (just check that it exists, not the value)
			expectedDurationMetric := `codecourt_database_operation_duration_seconds_bucket{operation="` + tc.operation + 
				`",service="` + tc.service + 
				`",table="` + tc.table + `"`
			
			if !strings.Contains(metricsOutput, expectedDurationMetric) {
				t.Errorf("metrics output does not contain expected database duration metric: %s", expectedDurationMetric)
			}
		})
	}
}

func TestResponseWriter(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "Status OK",
			statusCode: http.StatusOK,
		},
		{
			name:       "Status Not Found",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Status Internal Server Error",
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test response recorder
			rec := httptest.NewRecorder()
			
			// Create our responseWriter wrapper
			rw := &responseWriter{
				ResponseWriter: rec,
				statusCode:     0,
			}
			
			// Write the status code
			rw.WriteHeader(tc.statusCode)
			
			// Check that the status code was captured correctly
			if rw.Status() != tc.statusCode {
				t.Errorf("expected status code %d, got %d", tc.statusCode, rw.Status())
			}
			
			// Check that the status code was passed to the underlying ResponseWriter
			if rec.Code != tc.statusCode {
				t.Errorf("expected recorder status code %d, got %d", tc.statusCode, rec.Code)
			}
		})
	}
	
	// Test default status code (200 OK)
	t.Run("Default Status Code", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: rec,
			statusCode:     0,
		}
		
		// Don't call WriteHeader, should default to 200 OK
		if rw.Status() != http.StatusOK {
			t.Errorf("expected default status code %d, got %d", http.StatusOK, rw.Status())
		}
	})
}
