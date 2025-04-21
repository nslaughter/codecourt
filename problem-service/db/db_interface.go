package db

import "github.com/nslaughter/codecourt/problem-service/model"

// Repository defines the interface for database operations
type Repository interface {
	// Problem operations
	CreateProblem(problem *model.Problem) error
	GetProblem(id string) (*model.Problem, error)
	UpdateProblem(problem *model.Problem) error
	DeleteProblem(id string) error
	ListProblems(offset, limit int) ([]*model.Problem, error)
	ListProblemsByCategory(categoryID string, offset, limit int) ([]*model.Problem, error)
	
	// Test case operations
	CreateTestCase(testCase *model.TestCase) error
	GetTestCase(id string) (*model.TestCase, error)
	UpdateTestCase(testCase *model.TestCase) error
	DeleteTestCase(id string) error
	ListTestCases(problemID string) ([]*model.TestCase, error)
	
	// Category operations
	CreateCategory(category *model.Category) error
	GetCategory(id string) (*model.Category, error)
	GetCategoryByName(name string) (*model.Category, error)
	UpdateCategory(category *model.Category) error
	DeleteCategory(id string) error
	ListCategories() ([]*model.Category, error)
	
	// Problem-Category relationship operations
	AddProblemCategory(problemID, categoryID string) error
	RemoveProblemCategory(problemID, categoryID string) error
	ListProblemCategories(problemID string) ([]*model.Category, error)
	
	// Problem template operations
	CreateProblemTemplate(template *model.ProblemTemplate) error
	GetProblemTemplate(id string) (*model.ProblemTemplate, error)
	GetProblemTemplateByLanguage(problemID string, language model.Language) (*model.ProblemTemplate, error)
	UpdateProblemTemplate(template *model.ProblemTemplate) error
	DeleteProblemTemplate(id string) error
	ListProblemTemplates(problemID string) ([]*model.ProblemTemplate, error)
	
	// Transaction support
	BeginTx() (Transaction, error)
	
	// Close the database connection
	Close() error
}

// Transaction defines the interface for database transactions
type Transaction interface {
	// Problem operations
	CreateProblem(problem *model.Problem) error
	
	// Test case operations
	CreateTestCase(testCase *model.TestCase) error
	
	// Category operations
	CreateCategory(category *model.Category) error
	
	// Problem-Category relationship operations
	AddProblemCategory(problemID, categoryID string) error
	
	// Problem template operations
	CreateProblemTemplate(template *model.ProblemTemplate) error
	
	// Transaction control
	Commit() error
	Rollback() error
}
