package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/problem-service/model"
)

// CreateProblemTemplate creates a new problem template in the database
func (db *DB) CreateProblemTemplate(template *model.ProblemTemplate) error {
	// Generate a new UUID if not provided
	if template.ID == "" {
		template.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	// Insert into database
	_, err := db.conn.Exec(`
		INSERT INTO problem_templates (id, problem_id, language, template, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (problem_id, language) DO UPDATE
		SET template = $4, updated_at = $6
	`,
		template.ID,
		template.ProblemID,
		template.Language,
		template.Template,
		template.CreatedAt,
		template.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create problem template: %w", err)
	}

	return nil
}

// GetProblemTemplate gets a problem template by ID
func (db *DB) GetProblemTemplate(id string) (*model.ProblemTemplate, error) {
	var template model.ProblemTemplate

	err := db.conn.QueryRow(`
		SELECT id, problem_id, language, template, created_at, updated_at
		FROM problem_templates
		WHERE id = $1
	`, id).Scan(
		&template.ID,
		&template.ProblemID,
		&template.Language,
		&template.Template,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem template: %w", err)
	}

	return &template, nil
}

// GetProblemTemplateByLanguage gets a problem template by problem ID and language
func (db *DB) GetProblemTemplateByLanguage(problemID string, language model.Language) (*model.ProblemTemplate, error) {
	var template model.ProblemTemplate

	err := db.conn.QueryRow(`
		SELECT id, problem_id, language, template, created_at, updated_at
		FROM problem_templates
		WHERE problem_id = $1 AND language = $2
	`, problemID, language).Scan(
		&template.ID,
		&template.ProblemID,
		&template.Language,
		&template.Template,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem template by language: %w", err)
	}

	return &template, nil
}

// UpdateProblemTemplate updates a problem template in the database
func (db *DB) UpdateProblemTemplate(template *model.ProblemTemplate) error {
	// Update timestamp
	template.UpdatedAt = time.Now()

	// Update in database
	_, err := db.conn.Exec(`
		UPDATE problem_templates
		SET template = $1, updated_at = $2
		WHERE id = $3
	`,
		template.Template,
		template.UpdatedAt,
		template.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update problem template: %w", err)
	}

	return nil
}

// DeleteProblemTemplate deletes a problem template from the database
func (db *DB) DeleteProblemTemplate(id string) error {
	_, err := db.conn.Exec(`
		DELETE FROM problem_templates
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete problem template: %w", err)
	}

	return nil
}

// ListProblemTemplates lists all templates for a problem
func (db *DB) ListProblemTemplates(problemID string) ([]*model.ProblemTemplate, error) {
	rows, err := db.conn.Query(`
		SELECT id, problem_id, language, template, created_at, updated_at
		FROM problem_templates
		WHERE problem_id = $1
		ORDER BY language ASC
	`, problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list problem templates: %w", err)
	}
	defer rows.Close()

	var templates []*model.ProblemTemplate
	for rows.Next() {
		var template model.ProblemTemplate
		err := rows.Scan(
			&template.ID,
			&template.ProblemID,
			&template.Language,
			&template.Template,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem template: %w", err)
		}
		templates = append(templates, &template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating problem templates: %w", err)
	}

	return templates, nil
}

// Transaction implementation for problem templates

// CreateProblemTemplate creates a new problem template in a transaction
func (tx *Tx) CreateProblemTemplate(template *model.ProblemTemplate) error {
	// Generate a new UUID if not provided
	if template.ID == "" {
		template.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	// Insert into database
	_, err := tx.tx.Exec(`
		INSERT INTO problem_templates (id, problem_id, language, template, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (problem_id, language) DO UPDATE
		SET template = $4, updated_at = $6
	`,
		template.ID,
		template.ProblemID,
		template.Language,
		template.Template,
		template.CreatedAt,
		template.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create problem template in transaction: %w", err)
	}

	return nil
}
