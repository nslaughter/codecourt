package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds the configuration for the Notification Service
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

	// Kafka configuration
	KafkaBrokers []string
	KafkaGroupID string
	KafkaTopics  []string

	// Email configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Load server configuration
	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8083"))
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
	cfg.DBName = getEnv("DB_NAME", "notification_service")
	cfg.DBSSLMode = getEnv("DB_SSLMODE", "disable")

	// Load Kafka configuration
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	cfg.KafkaBrokers = strings.Split(kafkaBrokers, ",")
	cfg.KafkaGroupID = getEnv("KAFKA_GROUP_ID", "notification-service")
	
	kafkaTopics := getEnv("KAFKA_TOPICS", "submission-created,submission-judged,user-registered")
	cfg.KafkaTopics = strings.Split(kafkaTopics, ",")

	// Load email configuration
	cfg.SMTPHost = getEnv("SMTP_HOST", "smtp.example.com")
	
	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %v", err)
	}
	cfg.SMTPPort = smtpPort
	
	cfg.SMTPUsername = getEnv("SMTP_USERNAME", "")
	cfg.SMTPPassword = getEnv("SMTP_PASSWORD", "")
	cfg.SMTPFrom = getEnv("SMTP_FROM", "noreply@codecourt.com")

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
