package service

import "github.com/nslaughter/codecourt/problem-service/model"

// ProblemServiceInterface defines the interface for problem service operations
type ProblemServiceInterface interface {
	// Problem operations
	CreateProblem(req *model.ProblemRequest) (*model.Problem, error)
	GetProblem(id string) (*model.ProblemResponse, error)
	UpdateProblem(id string, req *model.ProblemRequest) (*model.Problem, error)
	DeleteProblem(id string) error
	ListProblems(offset, limit int) ([]*model.Problem, error)
	ListProblemsByCategory(categoryID string, offset, limit int) ([]*model.Problem, error)
	
	// Test case operations
	CreateTestCase(problemID string, req *model.TestCaseRequest) (*model.TestCase, error)
	GetTestCase(id string) (*model.TestCase, error)
	UpdateTestCase(id string, req *model.TestCaseRequest) (*model.TestCase, error)
	DeleteTestCase(id string) error
	ListTestCases(problemID string, includeHidden bool) ([]*model.TestCase, error)
	
	// Category operations
	CreateCategory(req *model.CategoryRequest) (*model.Category, error)
	GetCategory(id string) (*model.Category, error)
	UpdateCategory(id string, req *model.CategoryRequest) (*model.Category, error)
	DeleteCategory(id string) error
	ListCategories() ([]*model.Category, error)
	
	// Problem template operations
	CreateProblemTemplate(problemID string, req *model.ProblemTemplateRequest) (*model.ProblemTemplate, error)
	GetProblemTemplate(id string) (*model.ProblemTemplate, error)
	GetProblemTemplateByLanguage(problemID string, language model.Language) (*model.ProblemTemplate, error)
	UpdateProblemTemplate(id string, req *model.ProblemTemplateRequest) (*model.ProblemTemplate, error)
	DeleteProblemTemplate(id string) error
	ListProblemTemplates(problemID string) ([]*model.ProblemTemplate, error)
}
