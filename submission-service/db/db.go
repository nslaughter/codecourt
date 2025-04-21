package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/nslaughter/codecourt/submission-service/config"
	"github.com/nslaughter/codecourt/submission-service/model"
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
	// Create submissions table
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS submissions (
			id UUID PRIMARY KEY,
			problem_id UUID NOT NULL,
			user_id UUID NOT NULL,
			language VARCHAR(50) NOT NULL,
			code TEXT NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create submissions table: %w", err)
	}

	// Create submission_results table
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS submission_results (
			id UUID PRIMARY KEY,
			submission_id UUID NOT NULL REFERENCES submissions(id),
			status VARCHAR(50) NOT NULL,
			execution_time INT,
			memory_usage INT,
			error_message TEXT,
			created_at TIMESTAMP NOT NULL,
			CONSTRAINT fk_submission
				FOREIGN KEY(submission_id)
				REFERENCES submissions(id)
				ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create submission_results table: %w", err)
	}

	// Create test_case_results table
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS test_case_results (
			id UUID PRIMARY KEY,
			submission_result_id UUID NOT NULL REFERENCES submission_results(id),
			test_case_id UUID NOT NULL,
			status VARCHAR(50) NOT NULL,
			execution_time INT,
			memory_usage INT,
			expected_output TEXT,
			actual_output TEXT,
			error_message TEXT,
			created_at TIMESTAMP NOT NULL,
			CONSTRAINT fk_submission_result
				FOREIGN KEY(submission_result_id)
				REFERENCES submission_results(id)
				ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create test_case_results table: %w", err)
	}

	return nil
}

// CreateSubmission creates a new submission in the database
func (db *DB) CreateSubmission(submission *model.Submission) error {
	// Generate a new UUID if not provided
	if submission.ID == "" {
		submission.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	submission.CreatedAt = now
	submission.UpdatedAt = now

	// Insert into database
	_, err := db.conn.Exec(`
		INSERT INTO submissions (id, problem_id, user_id, language, code, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		submission.ID,
		submission.ProblemID,
		submission.UserID,
		submission.Language,
		submission.Code,
		submission.Status,
		submission.CreatedAt,
		submission.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create submission: %w", err)
	}

	return nil
}

// GetSubmission gets a submission by ID
func (db *DB) GetSubmission(id string) (*model.Submission, error) {
	var submission model.Submission

	err := db.conn.QueryRow(`
		SELECT id, problem_id, user_id, language, code, status, created_at, updated_at
		FROM submissions
		WHERE id = $1
	`, id).Scan(
		&submission.ID,
		&submission.ProblemID,
		&submission.UserID,
		&submission.Language,
		&submission.Code,
		&submission.Status,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("submission not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	return &submission, nil
}

// UpdateSubmissionStatus updates the status of a submission
func (db *DB) UpdateSubmissionStatus(id string, status string) error {
	_, err := db.conn.Exec(`
		UPDATE submissions
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	return nil
}

// SaveSubmissionResult saves a submission result to the database
func (db *DB) SaveSubmissionResult(result *model.SubmissionResult) error {
	// Generate a new UUID if not provided
	if result.ID == "" {
		result.ID = uuid.New().String()
	}

	// Set timestamp
	result.CreatedAt = time.Now()

	// Start a transaction
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert submission result
	_, err = tx.Exec(`
		INSERT INTO submission_results (id, submission_id, status, execution_time, memory_usage, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		result.ID,
		result.SubmissionID,
		result.Status,
		result.ExecutionTime,
		result.MemoryUsage,
		result.ErrorMessage,
		result.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save submission result: %w", err)
	}

	// Insert test case results
	for _, testResult := range result.TestCaseResults {
		if testResult.ID == "" {
			testResult.ID = uuid.New().String()
		}
		testResult.CreatedAt = result.CreatedAt

		_, err = tx.Exec(`
			INSERT INTO test_case_results (
				id, submission_result_id, test_case_id, status, execution_time, 
				memory_usage, expected_output, actual_output, error_message, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
			testResult.ID,
			result.ID,
			testResult.TestCaseID,
			testResult.Status,
			testResult.ExecutionTime,
			testResult.MemoryUsage,
			testResult.ExpectedOutput,
			testResult.ActualOutput,
			testResult.ErrorMessage,
			testResult.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to save test case result: %w", err)
		}
	}

	// Update submission status
	_, err = tx.Exec(`
		UPDATE submissions
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, result.Status, result.CreatedAt, result.SubmissionID)
	if err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetSubmissionsByUserID gets all submissions for a user
func (db *DB) GetSubmissionsByUserID(userID string) ([]*model.Submission, error) {
	rows, err := db.conn.Query(`
		SELECT id, problem_id, user_id, language, code, status, created_at, updated_at
		FROM submissions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}
	defer rows.Close()

	var submissions []*model.Submission
	for rows.Next() {
		var submission model.Submission
		err := rows.Scan(
			&submission.ID,
			&submission.ProblemID,
			&submission.UserID,
			&submission.Language,
			&submission.Code,
			&submission.Status,
			&submission.CreatedAt,
			&submission.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}
		submissions = append(submissions, &submission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating submissions: %w", err)
	}

	return submissions, nil
}

// GetSubmissionsByProblemID gets all submissions for a problem
func (db *DB) GetSubmissionsByProblemID(problemID string) ([]*model.Submission, error) {
	rows, err := db.conn.Query(`
		SELECT id, problem_id, user_id, language, code, status, created_at, updated_at
		FROM submissions
		WHERE problem_id = $1
		ORDER BY created_at DESC
	`, problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}
	defer rows.Close()

	var submissions []*model.Submission
	for rows.Next() {
		var submission model.Submission
		err := rows.Scan(
			&submission.ID,
			&submission.ProblemID,
			&submission.UserID,
			&submission.Language,
			&submission.Code,
			&submission.Status,
			&submission.CreatedAt,
			&submission.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}
		submissions = append(submissions, &submission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating submissions: %w", err)
	}

	return submissions, nil
}

// GetSubmissionResult gets a submission result by submission ID
func (db *DB) GetSubmissionResult(submissionID string) (*model.SubmissionResult, error) {
	var result model.SubmissionResult

	// Get submission result
	err := db.conn.QueryRow(`
		SELECT id, submission_id, status, execution_time, memory_usage, error_message, created_at
		FROM submission_results
		WHERE submission_id = $1
	`, submissionID).Scan(
		&result.ID,
		&result.SubmissionID,
		&result.Status,
		&result.ExecutionTime,
		&result.MemoryUsage,
		&result.ErrorMessage,
		&result.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("submission result not found: %s", submissionID)
		}
		return nil, fmt.Errorf("failed to get submission result: %w", err)
	}

	// Get test case results
	rows, err := db.conn.Query(`
		SELECT id, test_case_id, status, execution_time, memory_usage, expected_output, actual_output, error_message, created_at
		FROM test_case_results
		WHERE submission_result_id = $1
	`, result.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test case results: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var testResult model.TestCaseResult
		err := rows.Scan(
			&testResult.ID,
			&testResult.TestCaseID,
			&testResult.Status,
			&testResult.ExecutionTime,
			&testResult.MemoryUsage,
			&testResult.ExpectedOutput,
			&testResult.ActualOutput,
			&testResult.ErrorMessage,
			&testResult.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test case result: %w", err)
		}
		result.TestCaseResults = append(result.TestCaseResults, testResult)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating test case results: %w", err)
	}

	return &result, nil
}
