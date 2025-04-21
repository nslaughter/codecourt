package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/problem-service/model"
)

// CreateCategory creates a new category in the database
func (db *DB) CreateCategory(category *model.Category) error {
	// Generate a new UUID if not provided
	if category.ID == "" {
		category.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	// Insert into database
	_, err := db.conn.Exec(`
		INSERT INTO categories (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`,
		category.ID,
		category.Name,
		category.CreatedAt,
		category.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

// GetCategory gets a category by ID
func (db *DB) GetCategory(id string) (*model.Category, error) {
	var category model.Category

	err := db.conn.QueryRow(`
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE id = $1
	`, id).Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// GetCategoryByName gets a category by name
func (db *DB) GetCategoryByName(name string) (*model.Category, error) {
	var category model.Category

	err := db.conn.QueryRow(`
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE name = $1
	`, name).Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get category by name: %w", err)
	}

	return &category, nil
}

// UpdateCategory updates a category in the database
func (db *DB) UpdateCategory(category *model.Category) error {
	// Update timestamp
	category.UpdatedAt = time.Now()

	// Update in database
	_, err := db.conn.Exec(`
		UPDATE categories
		SET name = $1, updated_at = $2
		WHERE id = $3
	`,
		category.Name,
		category.UpdatedAt,
		category.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// DeleteCategory deletes a category from the database
func (db *DB) DeleteCategory(id string) error {
	_, err := db.conn.Exec(`
		DELETE FROM categories
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// ListCategories lists all categories
func (db *DB) ListCategories() ([]*model.Category, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, created_at, updated_at
		FROM categories
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []*model.Category
	for rows.Next() {
		var category model.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// AddProblemCategory adds a problem-category relationship
func (db *DB) AddProblemCategory(problemID, categoryID string) error {
	now := time.Now()

	_, err := db.conn.Exec(`
		INSERT INTO problem_categories (problem_id, category_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (problem_id, category_id) DO NOTHING
	`,
		problemID,
		categoryID,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to add problem category: %w", err)
	}

	return nil
}

// RemoveProblemCategory removes a problem-category relationship
func (db *DB) RemoveProblemCategory(problemID, categoryID string) error {
	_, err := db.conn.Exec(`
		DELETE FROM problem_categories
		WHERE problem_id = $1 AND category_id = $2
	`,
		problemID,
		categoryID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove problem category: %w", err)
	}

	return nil
}

// ListProblemCategories lists all categories for a problem
func (db *DB) ListProblemCategories(problemID string) ([]*model.Category, error) {
	rows, err := db.conn.Query(`
		SELECT c.id, c.name, c.created_at, c.updated_at
		FROM categories c
		JOIN problem_categories pc ON c.id = pc.category_id
		WHERE pc.problem_id = $1
		ORDER BY c.name ASC
	`, problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list problem categories: %w", err)
	}
	defer rows.Close()

	var categories []*model.Category
	for rows.Next() {
		var category model.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// Transaction implementation for categories

// CreateCategory creates a new category in a transaction
func (tx *Tx) CreateCategory(category *model.Category) error {
	// Generate a new UUID if not provided
	if category.ID == "" {
		category.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	// Insert into database
	_, err := tx.tx.Exec(`
		INSERT INTO categories (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO NOTHING
	`,
		category.ID,
		category.Name,
		category.CreatedAt,
		category.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create category in transaction: %w", err)
	}

	return nil
}

// AddProblemCategory adds a problem-category relationship in a transaction
func (tx *Tx) AddProblemCategory(problemID, categoryID string) error {
	now := time.Now()

	_, err := tx.tx.Exec(`
		INSERT INTO problem_categories (problem_id, category_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (problem_id, category_id) DO NOTHING
	`,
		problemID,
		categoryID,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to add problem category in transaction: %w", err)
	}

	return nil
}
