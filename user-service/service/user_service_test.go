package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/user-service/config"
	"github.com/nslaughter/codecourt/user-service/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(id uuid.UUID, update *model.UserUpdate) (*model.User, error) {
	args := m.Called(id, update)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(id uuid.UUID, passwordHash string) error {
	args := m.Called(id, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) ListUsers() ([]*model.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserRepository) StoreRefreshToken(userID uuid.UUID, token string, expiresAt time.Time) error {
	args := m.Called(userID, token, expiresAt)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserIDByRefreshToken(token string) (uuid.UUID, error) {
	args := m.Called(token)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepository) DeleteRefreshToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteAllRefreshTokens(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockUserRepository)
	
	// Create test config
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpiry:     time.Hour,
		RefreshExpiry: time.Hour * 24,
	}
	
	// Create service
	service := NewUserService(mockRepo, cfg)
	
	// Test data
	registration := &model.UserRegistration{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}
	
	// Test cases
	tests := []struct {
		name          string
		setupMock     func()
		expectedError error
	}{
		{
			name: "Successful registration",
			setupMock: func() {
				mockRepo.On("GetUserByUsername", "testuser").Return(nil, nil)
				mockRepo.On("GetUserByEmail", "test@example.com").Return(nil, nil)
				mockRepo.On("CreateUser", mock.AnythingOfType("*model.User")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Username already exists",
			setupMock: func() {
				existingUser := &model.User{
					ID:       uuid.New(),
					Username: "testuser",
				}
				mockRepo.On("GetUserByUsername", "testuser").Return(existingUser, nil)
			},
			expectedError: ErrUsernameExists,
		},
		{
			name: "Email already exists",
			setupMock: func() {
				mockRepo.On("GetUserByUsername", "testuser").Return(nil, nil)
				existingUser := &model.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
				mockRepo.On("GetUserByEmail", "test@example.com").Return(existingUser, nil)
			},
			expectedError: ErrEmailExists,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock
			mockRepo = new(MockUserRepository)
			service = NewUserService(mockRepo, cfg)
			
			// Setup mock
			tc.setupMock()
			
			// Call the method
			user, err := service.Register(registration)
			
			// Check the result
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, registration.Username, user.Username)
				assert.Equal(t, registration.Email, user.Email)
				assert.Equal(t, registration.FirstName, user.FirstName)
				assert.Equal(t, registration.LastName, user.LastName)
				assert.Equal(t, "user", user.Role)
				
				// Verify mock expectations
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockUserRepository)
	
	// Create test config
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpiry:     time.Hour,
		RefreshExpiry: time.Hour * 24,
	}
	
	// Create service
	service := NewUserService(mockRepo, cfg)
	
	// Create a test user with hashed password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &model.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	
	// Test data
	login := &model.UserLogin{
		Username: "testuser",
		Password: "password123",
	}
	
	// Test cases
	tests := []struct {
		name          string
		setupMock     func()
		expectedError error
	}{
		{
			name: "Successful login",
			setupMock: func() {
				mockRepo.On("GetUserByUsername", "testuser").Return(testUser, nil)
				mockRepo.On("StoreRefreshToken", testUser.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "User not found",
			setupMock: func() {
				mockRepo.On("GetUserByUsername", "testuser").Return(nil, nil)
			},
			expectedError: ErrInvalidCredentials,
		},
		{
			name: "Invalid password",
			setupMock: func() {
				mockRepo.On("GetUserByUsername", "testuser").Return(testUser, nil)
			},
			expectedError: ErrInvalidCredentials,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock
			mockRepo = new(MockUserRepository)
			service = NewUserService(mockRepo, cfg)
			
			// Setup mock
			tc.setupMock()
			
			// Call the method
			var loginData *model.UserLogin
			if tc.name == "Invalid password" {
				// Use incorrect password
				loginData = &model.UserLogin{
					Username: "testuser",
					Password: "wrongpassword",
				}
			} else {
				loginData = login
			}
			
			tokens, err := service.Login(loginData)
			
			// Check the result
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, tokens)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokens)
				assert.NotEmpty(t, tokens.AccessToken)
				assert.NotEmpty(t, tokens.RefreshToken)
				assert.Greater(t, tokens.ExpiresIn, int64(0))
				
				// Verify mock expectations
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockUserRepository)
	
	// Create test config
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		JWTExpiry:     time.Hour,
		RefreshExpiry: time.Hour * 24,
	}
	
	// Create service
	service := NewUserService(mockRepo, cfg)
	
	// Create a test user
	testUser := &model.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}
	
	// Generate a token pair
	mockRepo.On("StoreRefreshToken", testUser.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)
	tokenPair, err := service.generateTokenPair(testUser)
	assert.NoError(t, err)
	assert.NotNil(t, tokenPair)
	
	// Test cases
	tests := []struct {
		name          string
		token         string
		expectedError error
	}{
		{
			name:          "Valid token",
			token:         tokenPair.AccessToken,
			expectedError: nil,
		},
		{
			name:          "Invalid token",
			token:         "invalid-token",
			expectedError: ErrInvalidToken,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call the method
			claims, err := service.ValidateToken(tc.token)
			
			// Check the result
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, testUser.ID, claims.UserID)
				assert.Equal(t, testUser.Username, claims.Username)
				assert.Equal(t, testUser.Role, claims.Role)
			}
		})
	}
}
