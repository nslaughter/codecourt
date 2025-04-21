package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/user-service/model"
)

// UserRepository defines the interface for user database operations
type UserRepository interface {
	CreateUser(user *model.User) error
	GetUserByID(id uuid.UUID) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	UpdateUser(id uuid.UUID, update *model.UserUpdate) (*model.User, error)
	UpdatePassword(id uuid.UUID, passwordHash string) error
	DeleteUser(id uuid.UUID) error
	ListUsers() ([]*model.User, error)
	
	// Token operations
	StoreRefreshToken(userID uuid.UUID, token string, expiresAt time.Time) error
	GetUserIDByRefreshToken(token string) (uuid.UUID, error)
	DeleteRefreshToken(token string) error
	DeleteAllRefreshTokens(userID uuid.UUID) error
}

// EnsureUserRepository ensures that DB implements UserRepository
var _ UserRepository = (*DB)(nil)

// CreateUser creates a new user in the database
func (db *DB) CreateUser(user *model.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	_, err := db.Exec(
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	
	if err != nil {
		return err
	}
	
	return nil
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	
	var user model.User
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}
	
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (db *DB) GetUserByUsername(username string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	
	var user model.User
	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}
	
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	
	var user model.User
	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}
	
	return &user, nil
}

// UpdateUser updates a user's information
func (db *DB) UpdateUser(id uuid.UUID, update *model.UserUpdate) (*model.User, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	
	// Update only the provided fields
	query := `
		UPDATE users
		SET 
			email = COALESCE($1, email),
			first_name = COALESCE($2, first_name),
			last_name = COALESCE($3, last_name),
			role = COALESCE($4, role),
			updated_at = $5
		WHERE id = $6
	`
	
	now := time.Now().UTC()
	_, err = tx.Exec(
		query,
		nullableString(update.Email),
		nullableString(update.FirstName),
		nullableString(update.LastName),
		nullableString(update.Role),
		now,
		id,
	)
	
	if err != nil {
		return nil, err
	}
	
	// Get the updated user
	var user model.User
	query = `
		SELECT id, username, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	
	err = tx.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	
	return &user, nil
}

// UpdatePassword updates a user's password
func (db *DB) UpdatePassword(id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`
	
	_, err := db.Exec(query, passwordHash, time.Now().UTC(), id)
	return err
}

// DeleteUser deletes a user
func (db *DB) DeleteUser(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// ListUsers retrieves all users
func (db *DB) ListUsers() ([]*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		users = append(users, &user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return users, nil
}

// StoreRefreshToken stores a refresh token
func (db *DB) StoreRefreshToken(userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`
	
	_, err := db.Exec(query, token, userID, expiresAt)
	return err
}

// GetUserIDByRefreshToken retrieves a user ID by refresh token
func (db *DB) GetUserIDByRefreshToken(token string) (uuid.UUID, error) {
	query := `
		SELECT user_id
		FROM refresh_tokens
		WHERE token = $1 AND expires_at > $2
	`
	
	var userID uuid.UUID
	err := db.QueryRow(query, token, time.Now().UTC()).Scan(&userID)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, nil // Token not found or expired
		}
		return uuid.Nil, err
	}
	
	return userID, nil
}

// DeleteRefreshToken deletes a refresh token
func (db *DB) DeleteRefreshToken(token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := db.Exec(query, token)
	return err
}

// DeleteAllRefreshTokens deletes all refresh tokens for a user
func (db *DB) DeleteAllRefreshTokens(userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := db.Exec(query, userID)
	return err
}

// Helper function to handle nullable strings in SQL queries
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
