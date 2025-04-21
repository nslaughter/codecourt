package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/user-service/config"
	"github.com/nslaughter/codecourt/user-service/db"
	"github.com/nslaughter/codecourt/user-service/model"
	"golang.org/x/crypto/bcrypt"
)

// Common errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameExists    = errors.New("username already exists")
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("token has expired")
)

// UserServiceImpl implements the UserService interface
type UserServiceImpl struct {
	repo db.UserRepository
	cfg  *config.Config
}

// NewUserService creates a new user service
func NewUserService(repo db.UserRepository, cfg *config.Config) *UserServiceImpl {
	return &UserServiceImpl{
		repo: repo,
		cfg:  cfg,
	}
}

// Register registers a new user
func (s *UserServiceImpl) Register(reg *model.UserRegistration) (*model.UserResponse, error) {
	// Check if username already exists
	existingUser, err := s.repo.GetUserByUsername(reg.Username)
	if err != nil {
		return nil, fmt.Errorf("error checking username: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUsernameExists
	}

	// Check if email already exists
	existingUser, err = s.repo.GetUserByEmail(reg.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create the user
	now := time.Now().UTC()
	user := &model.User{
		ID:           uuid.New(),
		Username:     reg.Username,
		Email:        reg.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    reg.FirstName,
		LastName:     reg.LastName,
		Role:         "user", // Default role
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Save the user to the database
	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return model.NewUserResponse(user), nil
}

// GetUserByID retrieves a user by ID
func (s *UserServiceImpl) GetUserByID(id uuid.UUID) (*model.UserResponse, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return model.NewUserResponse(user), nil
}

// GetUserByUsername retrieves a user by username
func (s *UserServiceImpl) GetUserByUsername(username string) (*model.UserResponse, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return model.NewUserResponse(user), nil
}

// UpdateUser updates a user's information
func (s *UserServiceImpl) UpdateUser(id uuid.UUID, update *model.UserUpdate) (*model.UserResponse, error) {
	// Check if user exists
	existingUser, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	if existingUser == nil {
		return nil, ErrUserNotFound
	}

	// Check if email already exists (if updating email)
	if update.Email != "" && update.Email != existingUser.Email {
		userWithEmail, err := s.repo.GetUserByEmail(update.Email)
		if err != nil {
			return nil, fmt.Errorf("error checking email: %w", err)
		}
		if userWithEmail != nil {
			return nil, ErrEmailExists
		}
	}

	// Update the user
	updatedUser, err := s.repo.UpdateUser(id, update)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return model.NewUserResponse(updatedUser), nil
}

// ChangePassword changes a user's password
func (s *UserServiceImpl) ChangePassword(id uuid.UUID, change *model.PasswordChange) error {
	// Get the user
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("error retrieving user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(change.CurrentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(change.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	// Update the password
	if err := s.repo.UpdatePassword(id, string(hashedPassword)); err != nil {
		return fmt.Errorf("error updating password: %w", err)
	}

	return nil
}

// DeleteUser deletes a user
func (s *UserServiceImpl) DeleteUser(id uuid.UUID) error {
	// Check if user exists
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("error retrieving user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Delete all refresh tokens for the user
	if err := s.repo.DeleteAllRefreshTokens(id); err != nil {
		return fmt.Errorf("error deleting refresh tokens: %w", err)
	}

	// Delete the user
	if err := s.repo.DeleteUser(id); err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}

// ListUsers retrieves all users
func (s *UserServiceImpl) ListUsers() ([]*model.UserResponse, error) {
	users, err := s.repo.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	// Convert to user responses
	userResponses := make([]*model.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = model.NewUserResponse(user)
	}

	return userResponses, nil
}

// Login authenticates a user and returns a token pair
func (s *UserServiceImpl) Login(login *model.UserLogin) (*model.TokenPair, error) {
	// Get the user
	user, err := s.repo.GetUserByUsername(login.Username)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(login.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate token pair
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("error generating tokens: %w", err)
	}

	return tokenPair, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *UserServiceImpl) RefreshToken(refreshToken string) (*model.TokenPair, error) {
	// Get user ID from refresh token
	userID, err := s.repo.GetUserIDByRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("error retrieving refresh token: %w", err)
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidToken
	}

	// Get the user
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Delete the old refresh token
	if err := s.repo.DeleteRefreshToken(refreshToken); err != nil {
		return nil, fmt.Errorf("error deleting refresh token: %w", err)
	}

	// Generate new token pair
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("error generating tokens: %w", err)
	}

	return tokenPair, nil
}

// Logout invalidates a refresh token
func (s *UserServiceImpl) Logout(refreshToken string) error {
	return s.repo.DeleteRefreshToken(refreshToken)
}

// LogoutAll invalidates all refresh tokens for a user
func (s *UserServiceImpl) LogoutAll(userID uuid.UUID) error {
	return s.repo.DeleteAllRefreshTokens(userID)
}

// ValidateToken validates a JWT token and returns the claims
func (s *UserServiceImpl) ValidateToken(tokenString string) (*TokenClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		if err.Error() == "token is expired" {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// Validate the token
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Extract user ID
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Extract username
	username, ok := claims["username"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Extract role
	role, ok := claims["role"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Extract expiry
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}
	expiresAt := time.Unix(int64(exp), 0)

	return &TokenClaims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: expiresAt,
	}, nil
}

// generateTokenPair generates an access token and refresh token
func (s *UserServiceImpl) generateTokenPair(user *model.User) (*model.TokenPair, error) {
	// Generate access token
	accessTokenExpiry := time.Now().Add(s.cfg.JWTExpiry)
	accessTokenClaims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"role":     user.Role,
		"exp":      accessTokenExpiry.Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshTokenExpiry := time.Now().Add(s.cfg.RefreshExpiry)
	refreshToken := uuid.NewString()

	// Store refresh token
	if err := s.repo.StoreRefreshToken(user.ID, refreshToken, refreshTokenExpiry); err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.JWTExpiry.Seconds()),
	}, nil
}
