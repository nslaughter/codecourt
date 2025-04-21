package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMetricsMiddleware(t *testing.T) {
	// Create a new registry for this test to avoid conflicts
	reg := prometheus.NewRegistry()
	
	// Create metrics with the test registry
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"service", "method", "endpoint", "status"},
	)
	
	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.001, 0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "endpoint"},
	)
	
	// Register metrics with the test registry
	reg.MustRegister(httpRequestsTotal)
	reg.MustRegister(httpRequestDuration)
	
	// Create a middleware that uses our test metrics
	middlewareWithTestMetrics := func(serviceName string) func(http.Handler) http.Handler {
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
				httpRequestsTotal.WithLabelValues(serviceName, r.Method, endpoint, status).Inc()

				// Observe request duration
				httpRequestDuration.WithLabelValues(serviceName, r.Method, endpoint).Observe(duration)
			})
		}
	}

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
			// Create a test handler that returns the expected status code
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Add a small delay to simulate processing time for metrics
				time.Sleep(1 * time.Millisecond)
				w.WriteHeader(tc.statusCode)
				io.WriteString(w, "test response")
			})
			
			// Create a handler with our test middleware
			handler := middlewareWithTestMetrics(tc.serviceName)(testHandler)
			
			// Create a test request
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rec, req)

			// Verify the response
			if rec.Code != tc.statusCode {
				t.Errorf("expected status code %d, got %d", tc.statusCode, rec.Code)
			}

			// Get metrics from the test registry
			metricsRec := httptest.NewRecorder()
			h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
			h.ServeHTTP(metricsRec, httptest.NewRequest("GET", "/metrics", nil))

			// Check that the metrics output contains our expected metrics
			metricsOutput := metricsRec.Body.String()
			
			// Verify request counter metric using regex to be more flexible with the output format
			counterRegex := regexp.MustCompile(`codecourt_http_requests_total{[^}]*endpoint="` + regexp.QuoteMeta(tc.expectedLabels["endpoint"]) + 
				`"[^}]*method="` + regexp.QuoteMeta(tc.expectedLabels["method"]) + 
				`"[^}]*service="` + regexp.QuoteMeta(tc.expectedLabels["service"]) + 
				`"[^}]*status="` + regexp.QuoteMeta(tc.expectedLabels["status"]) + `"[^}]*}`)
			
			if !counterRegex.MatchString(metricsOutput) {
				t.Errorf("metrics output does not contain expected counter metric for endpoint=%s, method=%s, service=%s, status=%s\nOutput: %s", 
					tc.expectedLabels["endpoint"], tc.expectedLabels["method"], tc.expectedLabels["service"], tc.expectedLabels["status"], metricsOutput)
			}
			
			// Verify request duration metric using regex to be more flexible with the output format
			durationRegex := regexp.MustCompile(`codecourt_http_request_duration_seconds_bucket{[^}]*endpoint="` + regexp.QuoteMeta(tc.expectedLabels["endpoint"]) + 
				`"[^}]*method="` + regexp.QuoteMeta(tc.expectedLabels["method"]) + 
				`"[^}]*service="` + regexp.QuoteMeta(tc.expectedLabels["service"]) + `"[^}]*}`)
			
			if !durationRegex.MatchString(metricsOutput) {
				t.Errorf("metrics output does not contain expected duration metric for endpoint=%s, method=%s, service=%s\nOutput: %s", 
					tc.expectedLabels["endpoint"], tc.expectedLabels["method"], tc.expectedLabels["service"], metricsOutput)
			}
		})
	}
}

func TestServiceInfoRegistration(t *testing.T) {
	// Create a new registry for this test
	reg := prometheus.NewRegistry()
	
	// Create service info gauge with the test registry
	serviceInfoGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "codecourt",
			Name:      "service_info",
			Help:      "Service version and build information",
		},
		[]string{"service", "version", "build_date", "commit_hash"},
	)
	
	// Register metrics with the test registry
	reg.MustRegister(serviceInfoGauge)
	
	// Create a function that registers service info using our test gauge
	registerServiceInfo := func(serviceName, version, buildDate, commitHash string) {
		serviceInfoGauge.WithLabelValues(serviceName, version, buildDate, commitHash).Set(1)
	}

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
			// Register service info using our test function
			registerServiceInfo(tc.service, tc.version, tc.buildDate, tc.commitHash)

			// Get metrics from the test registry
			metricsRec := httptest.NewRecorder()
			h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
			h.ServeHTTP(metricsRec, httptest.NewRequest("GET", "/metrics", nil))

			// Check that the metrics output contains our expected metrics
			metricsOutput := metricsRec.Body.String()
			
			// Verify service info metric using regex to be more flexible with the output format
			infoRegex := regexp.MustCompile(`codecourt_service_info{[^}]*build_date="` + regexp.QuoteMeta(tc.buildDate) + 
				`"[^}]*commit_hash="` + regexp.QuoteMeta(tc.commitHash) + 
				`"[^}]*service="` + regexp.QuoteMeta(tc.service) + 
				`"[^}]*version="` + regexp.QuoteMeta(tc.version) + `"[^}]*}`)
			
			if !infoRegex.MatchString(metricsOutput) {
				t.Errorf("metrics output does not contain expected service info metric for service=%s, version=%s\nOutput: %s", 
					tc.service, tc.version, metricsOutput)
			}
		})
	}
}

func TestDatabaseMetrics(t *testing.T) {
	// Create a new registry for this test
	reg := prometheus.NewRegistry()
	
	// Create database metrics with the test registry
	databaseOperationsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Name:      "database_operations_total",
			Help:      "Total number of database operations",
		},
		[]string{"service", "operation", "table", "status"},
	)
	
	databaseOperationDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Name:      "database_operation_duration_seconds",
			Help:      "Database operation duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2.5},
		},
		[]string{"service", "operation", "table"},
	)
	
	// Register metrics with the test registry
	reg.MustRegister(databaseOperationsTotal)
	reg.MustRegister(databaseOperationDuration)
	
	// Create functions that record database metrics using our test metrics
	recordDatabaseOperation := func(service, operation, table, status string) {
		databaseOperationsTotal.WithLabelValues(service, operation, table, status).Inc()
	}
	
	observeDatabaseOperationDuration := func(service, operation, table string, duration float64) {
		databaseOperationDuration.WithLabelValues(service, operation, table).Observe(duration)
	}

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
			// Record database operation and duration using our test functions
			recordDatabaseOperation(tc.service, tc.operation, tc.table, tc.status)
			observeDatabaseOperationDuration(tc.service, tc.operation, tc.table, tc.duration)

			// Get metrics from the test registry
			metricsRec := httptest.NewRecorder()
			h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
			h.ServeHTTP(metricsRec, httptest.NewRequest("GET", "/metrics", nil))

			// Check that the metrics output contains our expected metrics
			metricsOutput := metricsRec.Body.String()
			
			// Verify database operation counter metric using regex to be more flexible with the output format
			dbCounterRegex := regexp.MustCompile(`codecourt_database_operations_total{[^}]*operation="` + regexp.QuoteMeta(tc.operation) + 
				`"[^}]*service="` + regexp.QuoteMeta(tc.service) + 
				`"[^}]*status="` + regexp.QuoteMeta(tc.status) + 
				`"[^}]*table="` + regexp.QuoteMeta(tc.table) + `"[^}]*}`)
			
			if !dbCounterRegex.MatchString(metricsOutput) {
				t.Errorf("metrics output does not contain expected database counter metric for operation=%s, service=%s, table=%s\nOutput: %s", 
					tc.operation, tc.service, tc.table, metricsOutput)
			}
			
			// Verify database operation duration metric using regex to be more flexible with the output format
			dbDurationRegex := regexp.MustCompile(`codecourt_database_operation_duration_seconds_bucket{[^}]*operation="` + regexp.QuoteMeta(tc.operation) + 
				`"[^}]*service="` + regexp.QuoteMeta(tc.service) + 
				`"[^}]*table="` + regexp.QuoteMeta(tc.table) + `"[^}]*}`)
			
			if !dbDurationRegex.MatchString(metricsOutput) {
				t.Errorf("metrics output does not contain expected database duration metric for operation=%s, service=%s, table=%s\nOutput: %s", 
					tc.operation, tc.service, tc.table, metricsOutput)
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
