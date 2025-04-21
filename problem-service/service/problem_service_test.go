package service

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/problem-service/config"
	"github.com/nslaughter/codecourt/problem-service/db"
	"github.com/nslaughter/codecourt/problem-service/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

// Ensure MockRepository implements Repository interface
var _ db.Repository = (*MockRepository)(nil)

// Problem operations
func (m *MockRepository) CreateProblem(problem *model.Problem) error {
	args := m.Called(problem)
	return args.Error(0)
}

func (m *MockRepository) GetProblem(id string) (*model.Problem, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Problem), args.Error(1)
}

func (m *MockRepository) UpdateProblem(problem *model.Problem) error {
	args := m.Called(problem)
	return args.Error(0)
}

func (m *MockRepository) DeleteProblem(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) ListProblems(offset, limit int) ([]*model.Problem, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Problem), args.Error(1)
}

func (m *MockRepository) ListProblemsByCategory(categoryID string, offset, limit int) ([]*model.Problem, error) {
	args := m.Called(categoryID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Problem), args.Error(1)
}

// Test case operations
func (m *MockRepository) CreateTestCase(testCase *model.TestCase) error {
	args := m.Called(testCase)
	return args.Error(0)
}

func (m *MockRepository) GetTestCase(id string) (*model.TestCase, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TestCase), args.Error(1)
}

func (m *MockRepository) UpdateTestCase(testCase *model.TestCase) error {
	args := m.Called(testCase)
	return args.Error(0)
}

func (m *MockRepository) DeleteTestCase(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) ListTestCases(problemID string) ([]*model.TestCase, error) {
	args := m.Called(problemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.TestCase), args.Error(1)
}

// Category operations
func (m *MockRepository) CreateCategory(category *model.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockRepository) GetCategory(id string) (*model.Category, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Category), args.Error(1)
}

func (m *MockRepository) GetCategoryByName(name string) (*model.Category, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Category), args.Error(1)
}

func (m *MockRepository) UpdateCategory(category *model.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockRepository) DeleteCategory(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) ListCategories() ([]*model.Category, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Category), args.Error(1)
}

// Problem-Category relationship operations
func (m *MockRepository) AddProblemCategory(problemID, categoryID string) error {
	args := m.Called(problemID, categoryID)
	return args.Error(0)
}

func (m *MockRepository) RemoveProblemCategory(problemID, categoryID string) error {
	args := m.Called(problemID, categoryID)
	return args.Error(0)
}

func (m *MockRepository) ListProblemCategories(problemID string) ([]*model.Category, error) {
	args := m.Called(problemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Category), args.Error(1)
}

// Problem template operations
func (m *MockRepository) CreateProblemTemplate(template *model.ProblemTemplate) error {
	args := m.Called(template)
	return args.Error(0)
}

func (m *MockRepository) GetProblemTemplate(id string) (*model.ProblemTemplate, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProblemTemplate), args.Error(1)
}

func (m *MockRepository) GetProblemTemplateByLanguage(problemID string, language model.Language) (*model.ProblemTemplate, error) {
	args := m.Called(problemID, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProblemTemplate), args.Error(1)
}

func (m *MockRepository) UpdateProblemTemplate(template *model.ProblemTemplate) error {
	args := m.Called(template)
	return args.Error(0)
}

func (m *MockRepository) DeleteProblemTemplate(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) ListProblemTemplates(problemID string) ([]*model.ProblemTemplate, error) {
	args := m.Called(problemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ProblemTemplate), args.Error(1)
}

// Transaction support
func (m *MockRepository) BeginTx() (db.Transaction, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(db.Transaction), args.Error(1)
}

func (m *MockRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockTransaction is a mock implementation of the Transaction interface
type MockTransaction struct {
	mock.Mock
}

// Ensure MockTransaction implements Transaction interface
var _ db.Transaction = (*MockTransaction)(nil)

// Problem operations
func (m *MockTransaction) CreateProblem(problem *model.Problem) error {
	args := m.Called(problem)
	return args.Error(0)
}

// Test case operations
func (m *MockTransaction) CreateTestCase(testCase *model.TestCase) error {
	args := m.Called(testCase)
	return args.Error(0)
}

// Category operations
func (m *MockTransaction) CreateCategory(category *model.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

// Problem-Category relationship operations
func (m *MockTransaction) AddProblemCategory(problemID, categoryID string) error {
	args := m.Called(problemID, categoryID)
	return args.Error(0)
}

// Problem template operations
func (m *MockTransaction) CreateProblemTemplate(template *model.ProblemTemplate) error {
	args := m.Called(template)
	return args.Error(0)
}

// Transaction control
func (m *MockTransaction) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func TestGetProblem(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		id            string
		problem       *model.Problem
		testCases     []*model.TestCase
		categories    []*model.Category
		templates     []*model.ProblemTemplate
		dbError       error
		expectedError bool
	}{
		{
			name: "Success",
			id:   uuid.New().String(),
			problem: &model.Problem{
				ID:          uuid.New().String(),
				Title:       "Test Problem",
				Description: "Test Description",
				Difficulty:  model.DifficultyMedium,
				TimeLimit:   1000,
				MemoryLimit: 128,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			testCases: []*model.TestCase{
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					Input:     "1 2",
					Output:    "3",
					IsHidden:  false,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			categories: []*model.Category{
				{
					ID:        uuid.New().String(),
					Name:      "Test Category",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			templates: []*model.ProblemTemplate{
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					Language:  model.LanguageGo,
					Template:  "func solution(a, b int) int {\n\treturn a + b\n}",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			dbError:       nil,
			expectedError: false,
		},
		{
			name:          "DB Error",
			id:            uuid.New().String(),
			problem:       nil,
			testCases:     nil,
			categories:    nil,
			templates:     nil,
			dbError:       sql.ErrNoRows,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockRepository)

			// Set up expectations
			mockRepo.On("GetProblem", tc.id).Return(tc.problem, tc.dbError)
			if tc.problem != nil {
				// Use mock.Anything to avoid UUID comparison issues
				mockRepo.On("ListTestCases", mock.Anything).Return(tc.testCases, nil)
				mockRepo.On("ListProblemCategories", mock.Anything).Return(tc.categories, nil)
				mockRepo.On("ListProblemTemplates", mock.Anything).Return(tc.templates, nil)
			}

			// Create service
			service := NewProblemService(&config.Config{}, mockRepo)

			// Call method
			problem, err := service.GetProblem(tc.id)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, problem)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, problem)
				assert.Equal(t, tc.problem.ID, problem.ID)
				assert.Equal(t, tc.problem.Title, problem.Title)
				assert.Equal(t, tc.problem.Description, problem.Description)
				assert.Equal(t, tc.problem.Difficulty, problem.Difficulty)
				assert.Equal(t, len(tc.testCases), len(problem.TestCases))
				assert.Equal(t, len(tc.categories), len(problem.Categories))
				assert.Equal(t, len(tc.templates), len(problem.Templates))
			}

			// Verify expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateProblem(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		request        *model.ProblemRequest
		txError        error
		categoryExists bool
		expectedError  bool
	}{
		{
			name: "Success",
			request: &model.ProblemRequest{
				Title:       "Test Problem",
				Description: "Test Description",
				Difficulty:  model.DifficultyMedium,
				TimeLimit:   1000,
				MemoryLimit: 128,
				Categories:  []string{"Test Category"},
				TestCases: []struct {
					Input       string `json:"input"`
					Output      string `json:"output"`
					Explanation string `json:"explanation"`
					IsHidden    bool   `json:"is_hidden"`
				}{
					{
						Input:    "1 2",
						Output:   "3",
						IsHidden: false,
					},
				},
				Templates: []struct {
					Language model.Language `json:"language"`
					Template string         `json:"template"`
				}{
					{
						Language: model.LanguageGo,
						Template: "func solution(a, b int) int {\n\treturn a + b\n}",
					},
				},
			},
			txError:        nil,
			categoryExists: false,
			expectedError:  false,
		},
		{
			name: "Transaction Error",
			request: &model.ProblemRequest{
				Title:       "Test Problem",
				Description: "Test Description",
				Difficulty:  model.DifficultyMedium,
				TimeLimit:   1000,
				MemoryLimit: 128,
			},
			txError:        assert.AnError,
			categoryExists: false,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository and transaction
			mockRepo := new(MockRepository)
			mockTx := new(MockTransaction)

			// Set up expectations
			mockRepo.On("BeginTx").Return(mockTx, tc.txError)
			
			if tc.txError == nil {
				// Transaction methods
				mockTx.On("CreateProblem", mock.AnythingOfType("*model.Problem")).Return(nil)
				
				// Test cases
				for range tc.request.TestCases {
					mockTx.On("CreateTestCase", mock.AnythingOfType("*model.TestCase")).Return(nil)
				}
				
				// Categories
				for _, category := range tc.request.Categories {
					if tc.categoryExists {
						mockRepo.On("GetCategoryByName", category).Return(&model.Category{
							ID:   uuid.New().String(),
							Name: category,
						}, nil)
					} else {
						mockRepo.On("GetCategoryByName", category).Return(nil, sql.ErrNoRows)
						mockTx.On("CreateCategory", mock.AnythingOfType("*model.Category")).Return(nil)
					}
					mockTx.On("AddProblemCategory", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
				}
				
				// Templates
				for range tc.request.Templates {
					mockTx.On("CreateProblemTemplate", mock.AnythingOfType("*model.ProblemTemplate")).Return(nil)
				}
				
				mockTx.On("Commit").Return(nil)
				mockTx.On("Rollback").Return(nil)
			}

			// Create service
			service := NewProblemService(&config.Config{}, mockRepo)

			// Call method
			problem, err := service.CreateProblem(tc.request)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, problem)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, problem)
				assert.Equal(t, tc.request.Title, problem.Title)
				assert.Equal(t, tc.request.Description, problem.Description)
				assert.Equal(t, tc.request.Difficulty, problem.Difficulty)
				assert.Equal(t, tc.request.TimeLimit, problem.TimeLimit)
				assert.Equal(t, tc.request.MemoryLimit, problem.MemoryLimit)
			}

			// Verify expectations
			mockRepo.AssertExpectations(t)
			if tc.txError == nil {
				mockTx.AssertExpectations(t)
			}
		})
	}
}

func TestListTestCases(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		problemID     string
		includeHidden bool
		testCases     []*model.TestCase
		dbError       error
		expectedCount int
		expectedError bool
	}{
		{
			name:          "Success - Include Hidden",
			problemID:     uuid.New().String(),
			includeHidden: true,
			testCases: []*model.TestCase{
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					Input:     "1 2",
					Output:    "3",
					IsHidden:  false,
				},
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					Input:     "3 4",
					Output:    "7",
					IsHidden:  true,
				},
			},
			dbError:       nil,
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:          "Success - Exclude Hidden",
			problemID:     uuid.New().String(),
			includeHidden: false,
			testCases: []*model.TestCase{
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					Input:     "1 2",
					Output:    "3",
					IsHidden:  false,
				},
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					Input:     "3 4",
					Output:    "7",
					IsHidden:  true,
				},
			},
			dbError:       nil,
			expectedCount: 1,
			expectedError: false,
		},
		{
			name:          "DB Error",
			problemID:     uuid.New().String(),
			includeHidden: true,
			testCases:     nil,
			dbError:       assert.AnError,
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockRepository)

			// Set up expectations
			mockRepo.On("ListTestCases", tc.problemID).Return(tc.testCases, tc.dbError)

			// Create service
			service := NewProblemService(&config.Config{}, mockRepo)

			// Call method
			testCases, err := service.ListTestCases(tc.problemID, tc.includeHidden)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, testCases)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, testCases)
				assert.Equal(t, tc.expectedCount, len(testCases))
				
				// Verify hidden test cases are filtered correctly
				if !tc.includeHidden {
					for _, tc := range testCases {
						assert.False(t, tc.IsHidden)
					}
				}
			}

			// Verify expectations
			mockRepo.AssertExpectations(t)
		})
	}
}
