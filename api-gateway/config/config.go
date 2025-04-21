package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config represents the API Gateway configuration
type Config struct {
	// Server configuration
	ServerPort int

	// Service URLs
	ProblemServiceURL   string
	SubmissionServiceURL string
	JudgingServiceURL    string
	AuthServiceURL       string

	// JWT configuration
	JWTSecret string
	JWTExpiry int // in minutes
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Load server configuration
	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}
	cfg.ServerPort = serverPort

	// Load service URLs
	cfg.ProblemServiceURL = getEnv("PROBLEM_SERVICE_URL", "http://localhost:8081")
	cfg.SubmissionServiceURL = getEnv("SUBMISSION_SERVICE_URL", "http://localhost:8082")
	cfg.JudgingServiceURL = getEnv("JUDGING_SERVICE_URL", "http://localhost:8083")
	cfg.AuthServiceURL = getEnv("AUTH_SERVICE_URL", "http://localhost:8084")

	// Load JWT configuration
	cfg.JWTSecret = getEnv("JWT_SECRET", "your-secret-key")
	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY", "60"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY: %w", err)
	}
	cfg.JWTExpiry = jwtExpiry

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
