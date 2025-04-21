package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/problem-service/model"
)

// CreateTestCase creates a new test case in the database
func (db *DB) CreateTestCase(testCase *model.TestCase) error {
	// Generate a new UUID if not provided
	if testCase.ID == "" {
		testCase.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	testCase.CreatedAt = now
	testCase.UpdatedAt = now

	// Insert into database
	_, err := db.conn.Exec(`
		INSERT INTO test_cases (id, problem_id, input, output, explanation, is_hidden, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		testCase.ID,
		testCase.ProblemID,
		testCase.Input,
		testCase.Output,
		testCase.Explanation,
		testCase.IsHidden,
		testCase.CreatedAt,
		testCase.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create test case: %w", err)
	}

	return nil
}

// GetTestCase gets a test case by ID
func (db *DB) GetTestCase(id string) (*model.TestCase, error) {
	var testCase model.TestCase

	err := db.conn.QueryRow(`
		SELECT id, problem_id, input, output, explanation, is_hidden, created_at, updated_at
		FROM test_cases
		WHERE id = $1
	`, id).Scan(
		&testCase.ID,
		&testCase.ProblemID,
		&testCase.Input,
		&testCase.Output,
		&testCase.Explanation,
		&testCase.IsHidden,
		&testCase.CreatedAt,
		&testCase.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get test case: %w", err)
	}

	return &testCase, nil
}

// UpdateTestCase updates a test case in the database
func (db *DB) UpdateTestCase(testCase *model.TestCase) error {
	// Update timestamp
	testCase.UpdatedAt = time.Now()

	// Update in database
	_, err := db.conn.Exec(`
		UPDATE test_cases
		SET input = $1, output = $2, explanation = $3, is_hidden = $4, updated_at = $5
		WHERE id = $6
	`,
		testCase.Input,
		testCase.Output,
		testCase.Explanation,
		testCase.IsHidden,
		testCase.UpdatedAt,
		testCase.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update test case: %w", err)
	}

	return nil
}

// DeleteTestCase deletes a test case from the database
func (db *DB) DeleteTestCase(id string) error {
	_, err := db.conn.Exec(`
		DELETE FROM test_cases
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete test case: %w", err)
	}

	return nil
}

// ListTestCases lists all test cases for a problem
func (db *DB) ListTestCases(problemID string) ([]*model.TestCase, error) {
	rows, err := db.conn.Query(`
		SELECT id, problem_id, input, output, explanation, is_hidden, created_at, updated_at
		FROM test_cases
		WHERE problem_id = $1
		ORDER BY created_at ASC
	`, problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list test cases: %w", err)
	}
	defer rows.Close()

	var testCases []*model.TestCase
	for rows.Next() {
		var testCase model.TestCase
		err := rows.Scan(
			&testCase.ID,
			&testCase.ProblemID,
			&testCase.Input,
			&testCase.Output,
			&testCase.Explanation,
			&testCase.IsHidden,
			&testCase.CreatedAt,
			&testCase.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test case: %w", err)
		}
		testCases = append(testCases, &testCase)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating test cases: %w", err)
	}

	return testCases, nil
}

// Transaction implementation for test cases

// CreateTestCase creates a new test case in a transaction
func (tx *Tx) CreateTestCase(testCase *model.TestCase) error {
	// Generate a new UUID if not provided
	if testCase.ID == "" {
		testCase.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	testCase.CreatedAt = now
	testCase.UpdatedAt = now

	// Insert into database
	_, err := tx.tx.Exec(`
		INSERT INTO test_cases (id, problem_id, input, output, explanation, is_hidden, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		testCase.ID,
		testCase.ProblemID,
		testCase.Input,
		testCase.Output,
		testCase.Explanation,
		testCase.IsHidden,
		testCase.CreatedAt,
		testCase.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create test case in transaction: %w", err)
	}

	return nil
}
