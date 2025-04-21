package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/nslaughter/codecourt/problem-service/config"
)

// DB represents a database connection
type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize the database
	if err := initDB(conn); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initDB initializes the database schema
func initDB(conn *sql.DB) error {
	// Create problems table
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS problems (
			id UUID PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			difficulty VARCHAR(50) NOT NULL,
			time_limit INT NOT NULL,
			memory_limit INT NOT NULL,
			function_template TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create problems table: %w", err)
	}

	// Create test_cases table
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS test_cases (
			id UUID PRIMARY KEY,
			problem_id UUID NOT NULL,
			input TEXT NOT NULL,
			output TEXT NOT NULL,
			explanation TEXT,
			is_hidden BOOLEAN NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			CONSTRAINT fk_problem
				FOREIGN KEY(problem_id)
				REFERENCES problems(id)
				ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create test_cases table: %w", err)
	}

	// Create categories table
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create categories table: %w", err)
	}

	// Create problem_categories table
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS problem_categories (
			problem_id UUID NOT NULL,
			category_id UUID NOT NULL,
			created_at TIMESTAMP NOT NULL,
			PRIMARY KEY (problem_id, category_id),
			CONSTRAINT fk_problem
				FOREIGN KEY(problem_id)
				REFERENCES problems(id)
				ON DELETE CASCADE,
			CONSTRAINT fk_category
				FOREIGN KEY(category_id)
				REFERENCES categories(id)
				ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create problem_categories table: %w", err)
	}

	// Create problem_templates table
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS problem_templates (
			id UUID PRIMARY KEY,
			problem_id UUID NOT NULL,
			language VARCHAR(50) NOT NULL,
			template TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			CONSTRAINT fk_problem
				FOREIGN KEY(problem_id)
				REFERENCES problems(id)
				ON DELETE CASCADE,
			CONSTRAINT unique_problem_language
				UNIQUE (problem_id, language)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create problem_templates table: %w", err)
	}

	return nil
}

// Tx represents a database transaction
type Tx struct {
	tx *sql.Tx
}

// BeginTx begins a transaction
func (db *DB) BeginTx() (Transaction, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &Tx{tx: tx}, nil
}

// Commit commits the transaction
func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

// Rollback rolls back the transaction
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}
