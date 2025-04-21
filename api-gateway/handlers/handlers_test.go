package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/nslaughter/codecourt/api-gateway/config"
	"github.com/nslaughter/codecourt/api-gateway/proxy"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Create a test config
	cfg := &config.Config{}

	// Create a service proxy
	serviceProxy := proxy.NewServiceProxy(cfg)

	// Create a handler
	handler := NewHandler(cfg, serviceProxy)

	// Create a test request
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	// Call the health check handler
	handler.HealthCheck(rr, req)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Parse the response body
	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestRegisterRoutes(t *testing.T) {
	// Create a test config
	cfg := &config.Config{}

	// Create a service proxy
	serviceProxy := proxy.NewServiceProxy(cfg)

	// Create a handler
	handler := NewHandler(cfg, serviceProxy)

	// Create a router
	router := mux.NewRouter()

	// Register routes
	handler.RegisterRoutes(router)

	// Test that routes are registered by checking a few key paths
	testCases := []struct {
		path   string
		method string
	}{
		{"/api/v1/health", "GET"},
		{"/api/v1/problems", "GET"},
		{"/api/v1/problems", "POST"},
		{"/api/v1/problems/123", "GET"},
		{"/api/v1/submissions", "GET"},
		{"/api/v1/auth/login", "POST"},
	}

	for _, tc := range testCases {
		t.Run(tc.path+"-"+tc.method, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(tc.method, tc.path, nil)
			
			// Test that the router matches the route
			var match mux.RouteMatch
			matched := router.Match(req, &match)
			
			// Assert that the route is matched
			assert.True(t, matched, "Route not matched: %s %s", tc.method, tc.path)
		})
	}
}
