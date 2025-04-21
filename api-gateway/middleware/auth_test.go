package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nslaughter/codecourt/api-gateway/config"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		JWTSecret: "test-secret",
		JWTExpiry: 60,
	}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For public paths, don't expect user in context
		if isPublicPath(r.URL.Path) {
			w.WriteHeader(http.StatusOK)
			return
		}

		// For protected paths, expect user in context
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			t.Error("Expected user in context")
			return
		}
		assert.Equal(t, "test-user", user.UserID)
		assert.Equal(t, "user", user.Role)
		w.WriteHeader(http.StatusOK)
	})

	// Create the auth middleware
	middleware := AuthMiddleware(cfg)

	// Create a valid token
	claims := &UserClaims{
		UserID: "test-user",
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Public path without token",
			path:           "/api/v1/auth/login",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Protected path without token",
			path:           "/api/v1/submissions",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Protected path with invalid token format",
			path:           "/api/v1/submissions",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Protected path with valid token",
			path:           "/api/v1/submissions",
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", tc.path, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Serve the request
			middleware(testHandler).ServeHTTP(rr, req)

			// Check the status code
			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestIsPublicPath(t *testing.T) {
	// Test cases
	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/v1/auth/login", true},
		{"/api/v1/auth/register", true},
		{"/api/v1/health", true},
		{"/api/v1/problems", true},
		{"/api/v1/problems/123", true},
		{"/api/v1/submissions", false},
		{"/api/v1/users", false},
		{"/api/v1/judging/results", false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			result := isPublicPath(tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRequireRole(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create the role middleware
	adminMiddleware := RequireRole("admin")

	// Test cases
	tests := []struct {
		name           string
		role           string
		expectedStatus int
	}{
		{
			name:           "User with admin role",
			role:           "admin",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User with non-admin role",
			role:           "user",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "/api/v1/admin", nil)

			// Add user to context
			claims := &UserClaims{
				UserID: "test-user",
				Role:   tc.role,
			}
			ctx := req.Context()
			ctx = context.WithValue(ctx, "user", claims)
			req = req.WithContext(ctx)

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Serve the request
			adminMiddleware(testHandler).ServeHTTP(rr, req)

			// Check the status code
			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}
