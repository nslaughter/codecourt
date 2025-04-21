package service

import (
	"database/sql"
	"fmt"

	"github.com/nslaughter/codecourt/problem-service/config"
	"github.com/nslaughter/codecourt/problem-service/db"
	"github.com/nslaughter/codecourt/problem-service/model"
)

// ProblemService represents the problem service
type ProblemService struct {
	cfg *config.Config
	db  db.Repository
}

// NewProblemService creates a new problem service
func NewProblemService(cfg *config.Config, repository db.Repository) *ProblemService {
	return &ProblemService{
		cfg: cfg,
		db:  repository,
	}
}

// CreateProblem creates a new problem with test cases, categories, and templates
func (s *ProblemService) CreateProblem(req *model.ProblemRequest) (*model.Problem, error) {
	// Create problem
	problem := model.NewProblem(
		req.Title,
		req.Description,
		req.Difficulty,
		req.TimeLimit,
		req.MemoryLimit,
		req.FunctionTemplate,
	)

	// Begin transaction
	tx, err := s.db.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create problem in transaction
	if err := tx.CreateProblem(problem); err != nil {
		return nil, fmt.Errorf("failed to create problem: %w", err)
	}

	// Create test cases
	for _, tc := range req.TestCases {
		testCase := model.NewTestCase(
			problem.ID,
			tc.Input,
			tc.Output,
			tc.Explanation,
			tc.IsHidden,
		)
		if err := tx.CreateTestCase(testCase); err != nil {
			return nil, fmt.Errorf("failed to create test case: %w", err)
		}
	}

	// Create or get categories and link to problem
	for _, categoryName := range req.Categories {
		// Try to get existing category
		category, err := s.db.GetCategoryByName(categoryName)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, fmt.Errorf("failed to get category: %w", err)
			}
			// Category doesn't exist, create it
			category = model.NewCategory(categoryName)
			if err := tx.CreateCategory(category); err != nil {
				return nil, fmt.Errorf("failed to create category: %w", err)
			}
		}

		// Link category to problem
		if err := tx.AddProblemCategory(problem.ID, category.ID); err != nil {
			return nil, fmt.Errorf("failed to link category to problem: %w", err)
		}
	}

	// Create templates
	for _, tmpl := range req.Templates {
		template := model.NewProblemTemplate(
			problem.ID,
			tmpl.Language,
			tmpl.Template,
		)
		if err := tx.CreateProblemTemplate(template); err != nil {
			return nil, fmt.Errorf("failed to create problem template: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return problem, nil
}

// GetProblem gets a problem by ID with all related data
func (s *ProblemService) GetProblem(id string) (*model.ProblemResponse, error) {
	// Get problem
	problem, err := s.db.GetProblem(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem: %w", err)
	}

	// Get test cases
	testCases, err := s.db.ListTestCases(id)
	if err != nil {
		return nil, fmt.Errorf("failed to list test cases: %w", err)
	}

	// Get categories
	categories, err := s.db.ListProblemCategories(id)
	if err != nil {
		return nil, fmt.Errorf("failed to list problem categories: %w", err)
	}

	// Get templates
	templates, err := s.db.ListProblemTemplates(id)
	if err != nil {
		return nil, fmt.Errorf("failed to list problem templates: %w", err)
	}

	// Create response
	response := &model.ProblemResponse{
		ID:               problem.ID,
		Title:            problem.Title,
		Description:      problem.Description,
		Difficulty:       problem.Difficulty,
		TimeLimit:        problem.TimeLimit,
		MemoryLimit:      problem.MemoryLimit,
		FunctionTemplate: problem.FunctionTemplate,
		Categories:       make([]model.Category, 0, len(categories)),
		Templates:        make([]struct {
			Language model.Language `json:"language"`
			Template string         `json:"template"`
		}, 0, len(templates)),
		TestCases: make([]struct {
			ID          string `json:"id"`
			Input       string `json:"input"`
			Output      string `json:"output"`
			Explanation string `json:"explanation"`
			IsHidden    bool   `json:"is_hidden"`
		}, 0, len(testCases)),
		CreatedAt: problem.CreatedAt,
		UpdatedAt: problem.UpdatedAt,
	}

	// Add categories
	for _, category := range categories {
		response.Categories = append(response.Categories, *category)
	}

	// Add templates
	for _, template := range templates {
		response.Templates = append(response.Templates, struct {
			Language model.Language `json:"language"`
			Template string         `json:"template"`
		}{
			Language: template.Language,
			Template: template.Template,
		})
	}

	// Add test cases
	for _, testCase := range testCases {
		response.TestCases = append(response.TestCases, struct {
			ID          string `json:"id"`
			Input       string `json:"input"`
			Output      string `json:"output"`
			Explanation string `json:"explanation"`
			IsHidden    bool   `json:"is_hidden"`
		}{
			ID:          testCase.ID,
			Input:       testCase.Input,
			Output:      testCase.Output,
			Explanation: testCase.Explanation,
			IsHidden:    testCase.IsHidden,
		})
	}

	return response, nil
}

// UpdateProblem updates a problem
func (s *ProblemService) UpdateProblem(id string, req *model.ProblemRequest) (*model.Problem, error) {
	// Get problem
	problem, err := s.db.GetProblem(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem: %w", err)
	}

	// Update problem fields
	problem.Title = req.Title
	problem.Description = req.Description
	problem.Difficulty = req.Difficulty
	problem.TimeLimit = req.TimeLimit
	problem.MemoryLimit = req.MemoryLimit
	problem.FunctionTemplate = req.FunctionTemplate

	// Update problem in database
	if err := s.db.UpdateProblem(problem); err != nil {
		return nil, fmt.Errorf("failed to update problem: %w", err)
	}

	return problem, nil
}

// DeleteProblem deletes a problem
func (s *ProblemService) DeleteProblem(id string) error {
	if err := s.db.DeleteProblem(id); err != nil {
		return fmt.Errorf("failed to delete problem: %w", err)
	}
	return nil
}

// ListProblems lists all problems with pagination
func (s *ProblemService) ListProblems(offset, limit int) ([]*model.Problem, error) {
	return s.db.ListProblems(offset, limit)
}

// ListProblemsByCategory lists all problems in a category with pagination
func (s *ProblemService) ListProblemsByCategory(categoryID string, offset, limit int) ([]*model.Problem, error) {
	return s.db.ListProblemsByCategory(categoryID, offset, limit)
}

// CreateTestCase creates a new test case for a problem
func (s *ProblemService) CreateTestCase(problemID string, req *model.TestCaseRequest) (*model.TestCase, error) {
	// Create test case
	testCase := model.NewTestCase(
		problemID,
		req.Input,
		req.Output,
		req.Explanation,
		req.IsHidden,
	)

	// Save to database
	if err := s.db.CreateTestCase(testCase); err != nil {
		return nil, fmt.Errorf("failed to create test case: %w", err)
	}

	return testCase, nil
}

// GetTestCase gets a test case by ID
func (s *ProblemService) GetTestCase(id string) (*model.TestCase, error) {
	return s.db.GetTestCase(id)
}

// UpdateTestCase updates a test case
func (s *ProblemService) UpdateTestCase(id string, req *model.TestCaseRequest) (*model.TestCase, error) {
	// Get test case
	testCase, err := s.db.GetTestCase(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get test case: %w", err)
	}

	// Update test case fields
	testCase.Input = req.Input
	testCase.Output = req.Output
	testCase.Explanation = req.Explanation
	testCase.IsHidden = req.IsHidden

	// Update test case in database
	if err := s.db.UpdateTestCase(testCase); err != nil {
		return nil, fmt.Errorf("failed to update test case: %w", err)
	}

	return testCase, nil
}

// DeleteTestCase deletes a test case
func (s *ProblemService) DeleteTestCase(id string) error {
	if err := s.db.DeleteTestCase(id); err != nil {
		return fmt.Errorf("failed to delete test case: %w", err)
	}
	return nil
}

// ListTestCases lists all test cases for a problem
func (s *ProblemService) ListTestCases(problemID string, includeHidden bool) ([]*model.TestCase, error) {
	testCases, err := s.db.ListTestCases(problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list test cases: %w", err)
	}

	// Filter hidden test cases if needed
	if !includeHidden {
		filteredTestCases := make([]*model.TestCase, 0, len(testCases))
		for _, tc := range testCases {
			if !tc.IsHidden {
				filteredTestCases = append(filteredTestCases, tc)
			}
		}
		return filteredTestCases, nil
	}

	return testCases, nil
}

// CreateCategory creates a new category
func (s *ProblemService) CreateCategory(req *model.CategoryRequest) (*model.Category, error) {
	// Create category
	category := model.NewCategory(req.Name)

	// Save to database
	if err := s.db.CreateCategory(category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// GetCategory gets a category by ID
func (s *ProblemService) GetCategory(id string) (*model.Category, error) {
	return s.db.GetCategory(id)
}

// UpdateCategory updates a category
func (s *ProblemService) UpdateCategory(id string, req *model.CategoryRequest) (*model.Category, error) {
	// Get category
	category, err := s.db.GetCategory(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// Update category fields
	category.Name = req.Name

	// Update category in database
	if err := s.db.UpdateCategory(category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

// DeleteCategory deletes a category
func (s *ProblemService) DeleteCategory(id string) error {
	if err := s.db.DeleteCategory(id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}

// ListCategories lists all categories
func (s *ProblemService) ListCategories() ([]*model.Category, error) {
	return s.db.ListCategories()
}

// CreateProblemTemplate creates a new problem template
func (s *ProblemService) CreateProblemTemplate(problemID string, req *model.ProblemTemplateRequest) (*model.ProblemTemplate, error) {
	// Create template
	template := model.NewProblemTemplate(
		problemID,
		req.Language,
		req.Template,
	)

	// Save to database
	if err := s.db.CreateProblemTemplate(template); err != nil {
		return nil, fmt.Errorf("failed to create problem template: %w", err)
	}

	return template, nil
}

// GetProblemTemplate gets a problem template by ID
func (s *ProblemService) GetProblemTemplate(id string) (*model.ProblemTemplate, error) {
	return s.db.GetProblemTemplate(id)
}

// GetProblemTemplateByLanguage gets a problem template by problem ID and language
func (s *ProblemService) GetProblemTemplateByLanguage(problemID string, language model.Language) (*model.ProblemTemplate, error) {
	return s.db.GetProblemTemplateByLanguage(problemID, language)
}

// UpdateProblemTemplate updates a problem template
func (s *ProblemService) UpdateProblemTemplate(id string, req *model.ProblemTemplateRequest) (*model.ProblemTemplate, error) {
	// Get template
	template, err := s.db.GetProblemTemplate(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem template: %w", err)
	}

	// Update template fields
	template.Language = req.Language
	template.Template = req.Template

	// Update template in database
	if err := s.db.UpdateProblemTemplate(template); err != nil {
		return nil, fmt.Errorf("failed to update problem template: %w", err)
	}

	return template, nil
}

// DeleteProblemTemplate deletes a problem template
func (s *ProblemService) DeleteProblemTemplate(id string) error {
	if err := s.db.DeleteProblemTemplate(id); err != nil {
		return fmt.Errorf("failed to delete problem template: %w", err)
	}
	return nil
}

// ListProblemTemplates lists all templates for a problem
func (s *ProblemService) ListProblemTemplates(problemID string) ([]*model.ProblemTemplate, error) {
	return s.db.ListProblemTemplates(problemID)
}
