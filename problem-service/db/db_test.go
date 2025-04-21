package db

import (
	"testing"

	"github.com/nslaughter/codecourt/problem-service/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// This is a simple test to ensure the package compiles
	// In a real environment, we would use a test database
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "postgres",
		DBPassword: "postgres",
		DBName:     "codecourt_test",
		DBSSLMode:  "disable",
	}

	// We're not actually connecting to a database here
	// Just testing that the function doesn't panic
	_, err := New(cfg)
	assert.Error(t, err, "Expected error when connecting to non-existent database")
}
