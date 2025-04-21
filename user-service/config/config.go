package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for the User Service
type Config struct {
	// Server configuration
	ServerPort int
	
	// Database configuration
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	
	// JWT configuration
	JWTSecret     string
	JWTExpiry     time.Duration // in minutes
	RefreshExpiry time.Duration // in hours
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	
	// Load server configuration
	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %v", err)
	}
	cfg.ServerPort = serverPort
	
	// Load database configuration
	cfg.DBHost = getEnv("DB_HOST", "localhost")
	
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}
	cfg.DBPort = dbPort
	
	cfg.DBUser = getEnv("DB_USER", "postgres")
	cfg.DBPassword = getEnv("DB_PASSWORD", "postgres")
	cfg.DBName = getEnv("DB_NAME", "user_service")
	cfg.DBSSLMode = getEnv("DB_SSLMODE", "disable")
	
	// Load JWT configuration
	cfg.JWTSecret = getEnv("JWT_SECRET", "your-secret-key")
	
	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY", "60"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY: %v", err)
	}
	cfg.JWTExpiry = time.Duration(jwtExpiry) * time.Minute
	
	refreshExpiry, err := strconv.Atoi(getEnv("REFRESH_EXPIRY", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_EXPIRY: %v", err)
	}
	cfg.RefreshExpiry = time.Duration(refreshExpiry) * time.Hour
	
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
