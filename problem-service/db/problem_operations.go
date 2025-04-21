package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/problem-service/model"
)

// CreateProblem creates a new problem in the database
func (db *DB) CreateProblem(problem *model.Problem) error {
	// Generate a new UUID if not provided
	if problem.ID == "" {
		problem.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	problem.CreatedAt = now
	problem.UpdatedAt = now

	// Insert into database
	_, err := db.conn.Exec(`
		INSERT INTO problems (id, title, description, difficulty, time_limit, memory_limit, function_template, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		problem.ID,
		problem.Title,
		problem.Description,
		problem.Difficulty,
		problem.TimeLimit,
		problem.MemoryLimit,
		problem.FunctionTemplate,
		problem.CreatedAt,
		problem.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create problem: %w", err)
	}

	return nil
}

// GetProblem gets a problem by ID
func (db *DB) GetProblem(id string) (*model.Problem, error) {
	var problem model.Problem

	err := db.conn.QueryRow(`
		SELECT id, title, description, difficulty, time_limit, memory_limit, function_template, created_at, updated_at
		FROM problems
		WHERE id = $1
	`, id).Scan(
		&problem.ID,
		&problem.Title,
		&problem.Description,
		&problem.Difficulty,
		&problem.TimeLimit,
		&problem.MemoryLimit,
		&problem.FunctionTemplate,
		&problem.CreatedAt,
		&problem.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem: %w", err)
	}

	return &problem, nil
}

// UpdateProblem updates a problem in the database
func (db *DB) UpdateProblem(problem *model.Problem) error {
	// Update timestamp
	problem.UpdatedAt = time.Now()

	// Update in database
	_, err := db.conn.Exec(`
		UPDATE problems
		SET title = $1, description = $2, difficulty = $3, time_limit = $4, memory_limit = $5, function_template = $6, updated_at = $7
		WHERE id = $8
	`,
		problem.Title,
		problem.Description,
		problem.Difficulty,
		problem.TimeLimit,
		problem.MemoryLimit,
		problem.FunctionTemplate,
		problem.UpdatedAt,
		problem.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update problem: %w", err)
	}

	return nil
}

// DeleteProblem deletes a problem from the database
func (db *DB) DeleteProblem(id string) error {
	_, err := db.conn.Exec(`
		DELETE FROM problems
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete problem: %w", err)
	}

	return nil
}

// ListProblems lists all problems with pagination
func (db *DB) ListProblems(offset, limit int) ([]*model.Problem, error) {
	rows, err := db.conn.Query(`
		SELECT id, title, description, difficulty, time_limit, memory_limit, function_template, created_at, updated_at
		FROM problems
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list problems: %w", err)
	}
	defer rows.Close()

	var problems []*model.Problem
	for rows.Next() {
		var problem model.Problem
		err := rows.Scan(
			&problem.ID,
			&problem.Title,
			&problem.Description,
			&problem.Difficulty,
			&problem.TimeLimit,
			&problem.MemoryLimit,
			&problem.FunctionTemplate,
			&problem.CreatedAt,
			&problem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}
		problems = append(problems, &problem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating problems: %w", err)
	}

	return problems, nil
}

// ListProblemsByCategory lists all problems in a category with pagination
func (db *DB) ListProblemsByCategory(categoryID string, offset, limit int) ([]*model.Problem, error) {
	rows, err := db.conn.Query(`
		SELECT p.id, p.title, p.description, p.difficulty, p.time_limit, p.memory_limit, p.function_template, p.created_at, p.updated_at
		FROM problems p
		JOIN problem_categories pc ON p.id = pc.problem_id
		WHERE pc.category_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list problems by category: %w", err)
	}
	defer rows.Close()

	var problems []*model.Problem
	for rows.Next() {
		var problem model.Problem
		err := rows.Scan(
			&problem.ID,
			&problem.Title,
			&problem.Description,
			&problem.Difficulty,
			&problem.TimeLimit,
			&problem.MemoryLimit,
			&problem.FunctionTemplate,
			&problem.CreatedAt,
			&problem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}
		problems = append(problems, &problem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating problems: %w", err)
	}

	return problems, nil
}

// Transaction implementation for problems

// CreateProblem creates a new problem in a transaction
func (tx *Tx) CreateProblem(problem *model.Problem) error {
	// Generate a new UUID if not provided
	if problem.ID == "" {
		problem.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	problem.CreatedAt = now
	problem.UpdatedAt = now

	// Insert into database
	_, err := tx.tx.Exec(`
		INSERT INTO problems (id, title, description, difficulty, time_limit, memory_limit, function_template, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		problem.ID,
		problem.Title,
		problem.Description,
		problem.Difficulty,
		problem.TimeLimit,
		problem.MemoryLimit,
		problem.FunctionTemplate,
		problem.CreatedAt,
		problem.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create problem in transaction: %w", err)
	}

	return nil
}
