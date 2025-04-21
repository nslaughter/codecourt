package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for the judging service
type Config struct {
	// Kafka configuration
	KafkaBootstrapServers    string
	KafkaSubmissionTopic     string
	KafkaResultTopic         string
	KafkaGroupID             string
	KafkaAutoOffsetReset     string
	KafkaSessionTimeoutMs    int
	KafkaMaxPollIntervalMs   int
	KafkaEnableAutoCommit    bool
	KafkaAutoCommitIntervalMs int

	// Database configuration
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Judging configuration
	MaxExecutionTime time.Duration
	MaxMemoryUsage   int64 // in bytes
	SandboxEnabled   bool
	WorkDir          string
	ConcurrentJudges int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		// Kafka defaults
		KafkaBootstrapServers:    getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		KafkaSubmissionTopic:     getEnv("KAFKA_SUBMISSION_TOPIC", "code-submissions"),
		KafkaResultTopic:         getEnv("KAFKA_RESULT_TOPIC", "judge-results"),
		KafkaGroupID:             getEnv("KAFKA_GROUP_ID", "judging-service"),
		KafkaAutoOffsetReset:     getEnv("KAFKA_AUTO_OFFSET_RESET", "earliest"),
		KafkaSessionTimeoutMs:    getEnvAsInt("KAFKA_SESSION_TIMEOUT_MS", 10000),
		KafkaMaxPollIntervalMs:   getEnvAsInt("KAFKA_MAX_POLL_INTERVAL_MS", 300000),
		KafkaEnableAutoCommit:    getEnvAsBool("KAFKA_ENABLE_AUTO_COMMIT", true),
		KafkaAutoCommitIntervalMs: getEnvAsInt("KAFKA_AUTO_COMMIT_INTERVAL_MS", 5000),

		// Database defaults
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "codecourt"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Judging defaults
		MaxExecutionTime: getEnvAsDuration("MAX_EXECUTION_TIME", 10*time.Second),
		MaxMemoryUsage:   getEnvAsInt64("MAX_MEMORY_USAGE", 512*1024*1024), // 512 MB
		SandboxEnabled:   getEnvAsBool("SANDBOX_ENABLED", true),
		WorkDir:          getEnv("WORK_DIR", "/tmp/codecourt"),
		ConcurrentJudges: getEnvAsInt("CONCURRENT_JUDGES", 4),
	}

	// Create work directory if it doesn't exist
	if err := os.MkdirAll(cfg.WorkDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	return cfg, nil
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
