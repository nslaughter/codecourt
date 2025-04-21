package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/nslaughter/codecourt/notification-service/config"
)

// DB represents the database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	// Create the connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	// Open the database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// Initialize creates the necessary tables if they don't exist
func (db *DB) Initialize() error {
	// Create notifications table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			type VARCHAR(20) NOT NULL,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			status VARCHAR(20) NOT NULL,
			event_type VARCHAR(50),
			event_id VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			sent_at TIMESTAMP WITH TIME ZONE,
			read_at TIMESTAMP WITH TIME ZONE,
			template_id VARCHAR(50),
			template_data JSONB
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notifications table: %w", err)
	}

	// Create notification_templates table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notification_templates (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			event_type VARCHAR(50) NOT NULL,
			type VARCHAR(20) NOT NULL,
			subject VARCHAR(255),
			content TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notification_templates table: %w", err)
	}

	// Create notification_preferences table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notification_preferences (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			channels JSONB NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			UNIQUE(user_id, event_type)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notification_preferences table: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_event_type ON notifications(event_type)",
		"CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences(user_id)",
	}

	for _, idx := range indexes {
		_, err := db.Exec(idx)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
