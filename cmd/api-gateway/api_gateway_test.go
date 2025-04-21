package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nslaughter/codecourt/pkg/metrics"
)

func TestHealthCheckHandler(t *testing.T) {
	// Create a test request
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	healthCheckHandler(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Check the response body contains expected fields
	body := rec.Body.String()
	if !strings.Contains(body, "status") || !strings.Contains(body, "ok") {
		t.Errorf("response body does not contain expected fields: %s", body)
	}
	if !strings.Contains(body, serviceName) {
		t.Errorf("response body does not contain service name: %s", body)
	}
	if !strings.Contains(body, version) {
		t.Errorf("response body does not contain version: %s", body)
	}
}

func TestForwardHandlers(t *testing.T) {
	// Define test cases using table-driven style
	tests := []struct {
		name     string
		path     string
		method   string
		handler  func(http.ResponseWriter, *http.Request)
		expected string
	}{
		{
			name:     "Forward to User Service",
			path:     "/api/v1/users",
			method:   "GET",
			handler:  forwardToUserService,
			expected: "User Service",
		},
		{
			name:     "Forward to Problem Service",
			path:     "/api/v1/problems",
			method:   "GET",
			handler:  forwardToProblemService,
			expected: "Problem Service",
		},
		{
			name:     "Forward to Submission Service",
			path:     "/api/v1/submissions",
			method:   "POST",
			handler:  forwardToSubmissionService,
			expected: "Submission Service",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			// Call the handler
			tc.handler(rec, req)

			// Check the response
			if rec.Code != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
			}

			// Check the response body contains expected service name
			body := rec.Body.String()
			if !strings.Contains(body, tc.expected) {
				t.Errorf("response body does not contain expected service name: %s", body)
			}
		})
	}
}

func TestMetricsMiddlewareIntegration(t *testing.T) {
	// Create a new router
	mux := http.NewServeMux()

	// Register API routes
	mux.HandleFunc("/api/v1/health", healthCheckHandler)
	mux.HandleFunc("/api/v1/users", forwardToUserService)
	mux.HandleFunc("/api/v1/problems", forwardToProblemService)
	mux.HandleFunc("/api/v1/submissions", forwardToSubmissionService)

	// Set up metrics endpoint
	metrics.SetupMetricsEndpoint(mux)

	// Apply metrics middleware
	handler := metrics.MetricsMiddleware(serviceName)(mux)

	// Define test cases using table-driven style
	tests := []struct {
		name   string
		path   string
		method string
	}{
		{
			name:   "Health Check",
			path:   "/api/v1/health",
			method: "GET",
		},
		{
			name:   "User Service",
			path:   "/api/v1/users",
			method: "GET",
		},
		{
			name:   "Problem Service",
			path:   "/api/v1/problems",
			method: "GET",
		},
		{
			name:   "Submission Service",
			path:   "/api/v1/submissions",
			method: "POST",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rec, req)

			// Check the response
			if rec.Code != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
			}
		})
	}

	// Test metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check that metrics endpoint returns 200 OK
	if rec.Code != http.StatusOK {
		t.Errorf("expected metrics endpoint to return status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Check that metrics output contains expected metrics
	body := rec.Body.String()
	if !strings.Contains(body, "codecourt_http_requests_total") {
		t.Errorf("metrics output does not contain expected HTTP request metrics")
	}
}
