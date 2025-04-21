package model

import (
	"time"
)

// Difficulty represents the difficulty level of a problem
type Difficulty string

const (
	// DifficultyEasy represents an easy problem
	DifficultyEasy Difficulty = "EASY"
	// DifficultyMedium represents a medium problem
	DifficultyMedium Difficulty = "MEDIUM"
	// DifficultyHard represents a hard problem
	DifficultyHard Difficulty = "HARD"
)

// Problem represents a coding problem
type Problem struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Difficulty       Difficulty `json:"difficulty"`
	TimeLimit        int        `json:"time_limit"`       // in milliseconds
	MemoryLimit      int        `json:"memory_limit"`     // in megabytes
	FunctionTemplate string     `json:"function_template"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// TestCase represents a test case for a problem
type TestCase struct {
	ID          string    `json:"id"`
	ProblemID   string    `json:"problem_id"`
	Input       string    `json:"input"`
	Output      string    `json:"output"`
	Explanation string    `json:"explanation"`
	IsHidden    bool      `json:"is_hidden"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Category represents a problem category
type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProblemCategory represents a many-to-many relationship between problems and categories
type ProblemCategory struct {
	ProblemID  string    `json:"problem_id"`
	CategoryID string    `json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// Language represents a programming language
type Language string

const (
	// LanguageGo represents the Go programming language
	LanguageGo Language = "go"
	// LanguagePython represents the Python programming language
	LanguagePython Language = "python"
	// LanguageJava represents the Java programming language
	LanguageJava Language = "java"
	// LanguageCPP represents the C++ programming language
	LanguageCPP Language = "cpp"
)

// ProblemTemplate represents a code template for a specific language
type ProblemTemplate struct {
	ID        string    `json:"id"`
	ProblemID string    `json:"problem_id"`
	Language  Language  `json:"language"`
	Template  string    `json:"template"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewProblem creates a new problem
func NewProblem(title, description string, difficulty Difficulty, timeLimit, memoryLimit int, functionTemplate string) *Problem {
	return &Problem{
		Title:            title,
		Description:      description,
		Difficulty:       difficulty,
		TimeLimit:        timeLimit,
		MemoryLimit:      memoryLimit,
		FunctionTemplate: functionTemplate,
	}
}

// NewTestCase creates a new test case
func NewTestCase(problemID, input, output, explanation string, isHidden bool) *TestCase {
	return &TestCase{
		ProblemID:   problemID,
		Input:       input,
		Output:      output,
		Explanation: explanation,
		IsHidden:    isHidden,
	}
}

// NewCategory creates a new category
func NewCategory(name string) *Category {
	return &Category{
		Name: name,
	}
}

// NewProblemTemplate creates a new problem template
func NewProblemTemplate(problemID string, language Language, template string) *ProblemTemplate {
	return &ProblemTemplate{
		ProblemID: problemID,
		Language:  language,
		Template:  template,
	}
}

// ProblemRequest represents a request to create or update a problem
type ProblemRequest struct {
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Difficulty       Difficulty `json:"difficulty"`
	TimeLimit        int        `json:"time_limit"`
	MemoryLimit      int        `json:"memory_limit"`
	FunctionTemplate string     `json:"function_template"`
	Categories       []string   `json:"categories"`
	Templates        []struct {
		Language Language `json:"language"`
		Template string   `json:"template"`
	} `json:"templates"`
	TestCases []struct {
		Input       string `json:"input"`
		Output      string `json:"output"`
		Explanation string `json:"explanation"`
		IsHidden    bool   `json:"is_hidden"`
	} `json:"test_cases"`
}

// ProblemResponse represents a response to a problem request
type ProblemResponse struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Difficulty       Difficulty `json:"difficulty"`
	TimeLimit        int        `json:"time_limit"`
	MemoryLimit      int        `json:"memory_limit"`
	FunctionTemplate string     `json:"function_template"`
	Categories       []Category `json:"categories"`
	Templates        []struct {
		Language Language `json:"language"`
		Template string   `json:"template"`
	} `json:"templates"`
	TestCases []struct {
		ID          string `json:"id"`
		Input       string `json:"input"`
		Output      string `json:"output"`
		Explanation string `json:"explanation"`
		IsHidden    bool   `json:"is_hidden"`
	} `json:"test_cases"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TestCaseRequest represents a request to create or update a test case
type TestCaseRequest struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Explanation string `json:"explanation"`
	IsHidden    bool   `json:"is_hidden"`
}

// CategoryRequest represents a request to create or update a category
type CategoryRequest struct {
	Name string `json:"name"`
}

// ProblemTemplateRequest represents a request to create or update a problem template
type ProblemTemplateRequest struct {
	Language Language `json:"language"`
	Template string   `json:"template"`
}

// ProblemListResponse represents a response to a problem list request
type ProblemListResponse struct {
	Problems []struct {
		ID          string     `json:"id"`
		Title       string     `json:"title"`
		Difficulty  Difficulty `json:"difficulty"`
		Categories  []string   `json:"categories"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
	} `json:"problems"`
}
