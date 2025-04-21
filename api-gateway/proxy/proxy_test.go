package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nslaughter/codecourt/api-gateway/config"
	"github.com/stretchr/testify/assert"
)

func TestGetTargetURL(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		ProblemServiceURL:    "http://problem-service:8081",
		SubmissionServiceURL: "http://submission-service:8082",
		JudgingServiceURL:    "http://judging-service:8083",
		AuthServiceURL:       "http://auth-service:8084",
	}

	// Create a service proxy
	proxy := NewServiceProxy(cfg)

	// Test cases
	tests := []struct {
		path           string
		expectedTarget string
	}{
		{"/api/v1/problems", "http://problem-service:8081"},
		{"/api/v1/problems/123", "http://problem-service:8081"},
		{"/api/v1/submissions", "http://submission-service:8082"},
		{"/api/v1/submissions/123", "http://submission-service:8082"},
		{"/api/v1/judging/results", "http://judging-service:8083"},
		{"/api/v1/judging/status/123", "http://judging-service:8083"},
		{"/api/v1/auth/login", "http://auth-service:8084"},
		{"/api/v1/auth/register", "http://auth-service:8084"},
		{"/api/v1/unknown", "http://problem-service:8081"}, // Default
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			// Get the target URL
			targetURL, err := proxy.getTargetURL(tc.path)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedTarget, targetURL.String())
		})
	}
}

func TestProxyRequest(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		ProblemServiceURL: "http://localhost:9999", // Non-existent service
	}

	// Create a service proxy
	proxy := NewServiceProxy(cfg)

	// Create a test request
	req := httptest.NewRequest("GET", "/api/v1/problems", nil)
	rr := httptest.NewRecorder()

	// This will fail since the target service doesn't exist,
	// but we can still test that the proxy is working correctly
	// by checking that it attempts to forward the request
	proxy.ProxyRequest(rr, req)

	// The response should indicate a gateway error
	assert.Equal(t, http.StatusBadGateway, rr.Code)
}
