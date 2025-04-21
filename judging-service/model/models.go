package model

import (
	"time"
)

// Language represents a programming language
type Language string

// Supported programming languages
const (
	LanguageGo     Language = "go"
	LanguagePython Language = "python"
	LanguageJava   Language = "java"
	LanguageC      Language = "c"
	LanguageCPP    Language = "cpp"
)

// Status represents the status of a submission
type Status string

// Submission statuses
const (
	StatusPending    Status = "pending"
	StatusRunning    Status = "running"
	StatusAccepted   Status = "accepted"
	StatusRejected   Status = "rejected"
	StatusError      Status = "error"
	StatusTimeLimitExceeded Status = "time_limit_exceeded"
	StatusMemoryLimitExceeded Status = "memory_limit_exceeded"
	StatusCompilationError Status = "compilation_error"
	StatusRuntimeError Status = "runtime_error"
)

// Submission represents a code submission
type Submission struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ProblemID   string    `json:"problem_id"`
	Language    Language  `json:"language"`
	Code        string    `json:"code"`
	Status      Status    `json:"status"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// TestCase represents a test case for a problem
type TestCase struct {
	ID        string `json:"id"`
	ProblemID string `json:"problem_id"`
	Input     string `json:"input"`
	Output    string `json:"output"`
	IsHidden  bool   `json:"is_hidden"`
}

// TestResult represents the result of a test case execution
type TestResult struct {
	TestCaseID string `json:"test_case_id"`
	Passed     bool   `json:"passed"`
	ActualOutput string `json:"actual_output"`
	ExecutionTime time.Duration `json:"execution_time"`
	MemoryUsed int64 `json:"memory_used"`
	Error      string `json:"error,omitempty"`
}

// JudgingResult represents the result of judging a submission
type JudgingResult struct {
	SubmissionID  string       `json:"submission_id"`
	Status        Status       `json:"status"`
	TestResults   []TestResult `json:"test_results"`
	ExecutionTime time.Duration `json:"execution_time"`
	MemoryUsed    int64        `json:"memory_used"`
	CompileOutput string       `json:"compile_output,omitempty"`
	Error         string       `json:"error,omitempty"`
	JudgedAt      time.Time    `json:"judged_at"`
}
