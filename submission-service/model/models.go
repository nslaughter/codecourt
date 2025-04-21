package model

import (
	"time"
)

// SubmissionStatus represents the status of a submission
type SubmissionStatus string

const (
	// SubmissionStatusPending indicates the submission is pending processing
	SubmissionStatusPending SubmissionStatus = "PENDING"
	// SubmissionStatusProcessing indicates the submission is being processed
	SubmissionStatusProcessing SubmissionStatus = "PROCESSING"
	// SubmissionStatusCompleted indicates the submission has been processed
	SubmissionStatusCompleted SubmissionStatus = "COMPLETED"
	// SubmissionStatusFailed indicates the submission processing failed
	SubmissionStatusFailed SubmissionStatus = "FAILED"
)

// TestCaseStatus represents the status of a test case
type TestCaseStatus string

const (
	// TestCaseStatusPassed indicates the test case passed
	TestCaseStatusPassed TestCaseStatus = "PASSED"
	// TestCaseStatusFailed indicates the test case failed
	TestCaseStatusFailed TestCaseStatus = "FAILED"
	// TestCaseStatusError indicates there was an error running the test case
	TestCaseStatusError TestCaseStatus = "ERROR"
	// TestCaseStatusTimeLimitExceeded indicates the test case exceeded the time limit
	TestCaseStatusTimeLimitExceeded TestCaseStatus = "TIME_LIMIT_EXCEEDED"
	// TestCaseStatusMemoryLimitExceeded indicates the test case exceeded the memory limit
	TestCaseStatusMemoryLimitExceeded TestCaseStatus = "MEMORY_LIMIT_EXCEEDED"
)

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

// Submission represents a code submission
type Submission struct {
	ID        string          `json:"id"`
	ProblemID string          `json:"problem_id"`
	UserID    string          `json:"user_id"`
	Language  Language        `json:"language"`
	Code      string          `json:"code"`
	Status    SubmissionStatus `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// SubmissionResult represents the result of a submission
type SubmissionResult struct {
	ID              string           `json:"id"`
	SubmissionID    string           `json:"submission_id"`
	Status          SubmissionStatus `json:"status"`
	ExecutionTime   int              `json:"execution_time"`
	MemoryUsage     int              `json:"memory_usage"`
	ErrorMessage    string           `json:"error_message"`
	TestCaseResults []TestCaseResult `json:"test_case_results"`
	CreatedAt       time.Time        `json:"created_at"`
}

// TestCaseResult represents the result of a test case
type TestCaseResult struct {
	ID              string        `json:"id"`
	TestCaseID      string        `json:"test_case_id"`
	Status          TestCaseStatus `json:"status"`
	ExecutionTime   int           `json:"execution_time"`
	MemoryUsage     int           `json:"memory_usage"`
	ExpectedOutput  string        `json:"expected_output"`
	ActualOutput    string        `json:"actual_output"`
	ErrorMessage    string        `json:"error_message"`
	CreatedAt       time.Time     `json:"created_at"`
}

// NewSubmission creates a new submission
func NewSubmission(problemID, userID string, language Language, code string) *Submission {
	return &Submission{
		ProblemID: problemID,
		UserID:    userID,
		Language:  language,
		Code:      code,
		Status:    SubmissionStatusPending,
	}
}

// SubmissionRequest represents a request to create a submission
type SubmissionRequest struct {
	ProblemID string   `json:"problem_id"`
	UserID    string   `json:"user_id"`
	Language  Language `json:"language"`
	Code      string   `json:"code"`
}

// SubmissionResponse represents a response to a submission request
type SubmissionResponse struct {
	ID        string          `json:"id"`
	ProblemID string          `json:"problem_id"`
	UserID    string          `json:"user_id"`
	Language  Language        `json:"language"`
	Status    SubmissionStatus `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}

// SubmissionResultResponse represents a response to a submission result request
type SubmissionResultResponse struct {
	ID              string           `json:"id"`
	SubmissionID    string           `json:"submission_id"`
	Status          SubmissionStatus `json:"status"`
	ExecutionTime   int              `json:"execution_time"`
	MemoryUsage     int              `json:"memory_usage"`
	ErrorMessage    string           `json:"error_message"`
	TestCaseResults []TestCaseResult `json:"test_case_results"`
	CreatedAt       time.Time        `json:"created_at"`
}
