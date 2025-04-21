package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the configuration for the problem service
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
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Server configuration
	serverPort, err := getEnvInt("SERVER_PORT", 8081)
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}
	cfg.ServerPort = serverPort

	// Database configuration
	cfg.DBHost = getEnvString("DB_HOST", "localhost")
	dbPort, err := getEnvInt("DB_PORT", 5432)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	cfg.DBPort = dbPort
	cfg.DBUser = getEnvString("DB_USER", "postgres")
	cfg.DBPassword = getEnvString("DB_PASSWORD", "postgres")
	cfg.DBName = getEnvString("DB_NAME", "codecourt")
	cfg.DBSSLMode = getEnvString("DB_SSLMODE", "disable")

	return cfg, nil
}

// getEnvString gets an environment variable or returns a default value
func getEnvString(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// getEnvInt gets an environment variable as an integer or returns a default value
func getEnvInt(key string, defaultValue int) (int, error) {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, err
	}
	return value, nil
}
