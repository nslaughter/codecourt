package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/nslaughter/codecourt/judging-service/config"
	"github.com/nslaughter/codecourt/judging-service/model"
)

// DB represents a database connection
type DB struct {
	db *sql.DB
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db: db}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetTestCases retrieves test cases for a problem
func (d *DB) GetTestCases(problemID string) ([]model.TestCase, error) {
	query := `
		SELECT id, problem_id, input, output, is_hidden
		FROM test_cases
		WHERE problem_id = $1
		ORDER BY id
	`

	rows, err := d.db.Query(query, problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query test cases: %w", err)
	}
	defer rows.Close()

	var testCases []model.TestCase
	for rows.Next() {
		var tc model.TestCase
		if err := rows.Scan(&tc.ID, &tc.ProblemID, &tc.Input, &tc.Output, &tc.IsHidden); err != nil {
			return nil, fmt.Errorf("failed to scan test case: %w", err)
		}
		testCases = append(testCases, tc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating test cases: %w", err)
	}

	return testCases, nil
}

// UpdateSubmissionStatus updates the status of a submission
func (d *DB) UpdateSubmissionStatus(submissionID string, status model.Status) error {
	query := `
		UPDATE submissions
		SET status = $1
		WHERE id = $2
	`

	_, err := d.db.Exec(query, status, submissionID)
	if err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	return nil
}

// SaveJudgingResult saves the judging result to the database
func (d *DB) SaveJudgingResult(result *model.JudgingResult) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert judging result
	resultQuery := `
		INSERT INTO judging_results (
			submission_id, status, execution_time, memory_used, 
			compile_output, error, judged_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (submission_id) DO UPDATE SET
			status = EXCLUDED.status,
			execution_time = EXCLUDED.execution_time,
			memory_used = EXCLUDED.memory_used,
			compile_output = EXCLUDED.compile_output,
			error = EXCLUDED.error,
			judged_at = EXCLUDED.judged_at
	`

	_, err = tx.Exec(
		resultQuery,
		result.SubmissionID, result.Status, result.ExecutionTime,
		result.MemoryUsed, result.CompileOutput, result.Error, result.JudgedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert judging result: %w", err)
	}

	// Insert test results
	testResultQuery := `
		INSERT INTO test_results (
			submission_id, test_case_id, passed, actual_output,
			execution_time, memory_used, error
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (submission_id, test_case_id) DO UPDATE SET
			passed = EXCLUDED.passed,
			actual_output = EXCLUDED.actual_output,
			execution_time = EXCLUDED.execution_time,
			memory_used = EXCLUDED.memory_used,
			error = EXCLUDED.error
	`

	for _, tr := range result.TestResults {
		_, err = tx.Exec(
			testResultQuery,
			result.SubmissionID, tr.TestCaseID, tr.Passed, tr.ActualOutput,
			tr.ExecutionTime, tr.MemoryUsed, tr.Error,
		)
		if err != nil {
			return fmt.Errorf("failed to insert test result: %w", err)
		}
	}

	// Update submission status
	statusQuery := `
		UPDATE submissions
		SET status = $1
		WHERE id = $2
	`

	_, err = tx.Exec(statusQuery, result.Status, result.SubmissionID)
	if err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
