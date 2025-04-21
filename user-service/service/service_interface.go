package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/user-service/model"
)

// UserService defines the interface for user service operations
type UserService interface {
	// User management
	Register(reg *model.UserRegistration) (*model.UserResponse, error)
	GetUserByID(id uuid.UUID) (*model.UserResponse, error)
	GetUserByUsername(username string) (*model.UserResponse, error)
	UpdateUser(id uuid.UUID, update *model.UserUpdate) (*model.UserResponse, error)
	ChangePassword(id uuid.UUID, change *model.PasswordChange) error
	DeleteUser(id uuid.UUID) error
	ListUsers() ([]*model.UserResponse, error)
	
	// Authentication
	Login(login *model.UserLogin) (*model.TokenPair, error)
	RefreshToken(refreshToken string) (*model.TokenPair, error)
	Logout(refreshToken string) error
	LogoutAll(userID uuid.UUID) error
	
	// Token validation
	ValidateToken(token string) (*TokenClaims, error)
}

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	ExpiresAt time.Time `json:"exp"`
}
