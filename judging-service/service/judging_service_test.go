package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/judging-service/config"
	"github.com/nslaughter/codecourt/judging-service/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSandbox is a mock implementation of the Sandbox interface
type MockSandbox struct {
	mock.Mock
}

func (m *MockSandbox) Compile(ctx context.Context, language model.Language, code string) (string, error) {
	args := m.Called(ctx, language, code)
	return args.String(0), args.Error(1)
}

func (m *MockSandbox) Execute(ctx context.Context, language model.Language, code string, input string) (string, time.Duration, int64, error) {
	args := m.Called(ctx, language, code, input)
	return args.String(0), args.Get(1).(time.Duration), args.Get(2).(int64), args.Error(3)
}

// MockDB is a mock implementation of the DB interface
type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetTestCases(problemID string) ([]model.TestCase, error) {
	args := m.Called(problemID)
	return args.Get(0).([]model.TestCase), args.Error(1)
}

func (m *MockDB) UpdateSubmissionStatus(submissionID string, status model.Status) error {
	args := m.Called(submissionID, status)
	return args.Error(0)
}

func (m *MockDB) SaveJudgingResult(result *model.JudgingResult) error {
	args := m.Called(result)
	return args.Error(0)
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockKafkaConsumer is a mock implementation of the Kafka Consumer
type MockKafkaConsumer struct {
	mock.Mock
}

func (m *MockKafkaConsumer) Consume(timeout time.Duration) (interface{}, error) {
	args := m.Called(timeout)
	return args.Get(0), args.Error(1)
}

func (m *MockKafkaConsumer) Commit(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockKafkaConsumer) Close() {
	m.Called()
}

// MockKafkaProducer is a mock implementation of the Kafka Producer
type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) Produce(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockKafkaProducer) Close() {
	m.Called()
}

// TestJudgeSubmission tests the judgeSubmission function
func TestJudgeSubmission(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		submission     *model.Submission
		testCases      []model.TestCase
		compileOutput  string
		compileError   error
		executeOutputs []string
		executeTimes   []time.Duration
		executeMemory  []int64
		executeErrors  []error
		expectedStatus model.Status
	}{
		{
			name: "All test cases pass",
			submission: &model.Submission{
				ID:        uuid.New().String(),
				UserID:    uuid.New().String(),
				ProblemID: uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\nfunc main() { println(\"Hello, World!\") }",
				Status:    model.StatusPending,
			},
			testCases: []model.TestCase{
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					Input:     "",
					Output:    "Hello, World!",
				},
			},
			compileOutput:  "",
			compileError:   nil,
			executeOutputs: []string{"Hello, World!"},
			executeTimes:   []time.Duration{100 * time.Millisecond},
			executeMemory:  []int64{1024},
			executeErrors:  []error{nil},
			expectedStatus: model.StatusAccepted,
		},
		{
			name: "Compilation error",
			submission: &model.Submission{
				ID:        uuid.New().String(),
				UserID:    uuid.New().String(),
				ProblemID: uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\nfunc main() { println(\"Hello, World!) }",
				Status:    model.StatusPending,
			},
			testCases: []model.TestCase{
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					Input:     "",
					Output:    "Hello, World!",
				},
			},
			compileOutput:  "syntax error: unexpected }, expecting \"",
			compileError:   assert.AnError,
			executeOutputs: []string{},
			executeTimes:   []time.Duration{},
			executeMemory:  []int64{},
			executeErrors:  []error{},
			expectedStatus: model.StatusCompilationError,
		},
		{
			name: "Test case fails",
			submission: &model.Submission{
				ID:        uuid.New().String(),
				UserID:    uuid.New().String(),
				ProblemID: uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\nfunc main() { println(\"Wrong output\") }",
				Status:    model.StatusPending,
			},
			testCases: []model.TestCase{
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					Input:     "",
					Output:    "Hello, World!",
				},
			},
			compileOutput:  "",
			compileError:   nil,
			executeOutputs: []string{"Wrong output"},
			executeTimes:   []time.Duration{100 * time.Millisecond},
			executeMemory:  []int64{1024},
			executeErrors:  []error{nil},
			expectedStatus: model.StatusRejected,
		},
		{
			name: "Time limit exceeded",
			submission: &model.Submission{
				ID:        uuid.New().String(),
				UserID:    uuid.New().String(),
				ProblemID: uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\nfunc main() { for{} }",
				Status:    model.StatusPending,
			},
			testCases: []model.TestCase{
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					Input:     "",
					Output:    "Hello, World!",
				},
			},
			compileOutput:  "",
			compileError:   nil,
			executeOutputs: []string{""},
			executeTimes:   []time.Duration{11 * time.Second}, // Over the default 10s limit
			executeMemory:  []int64{1024},
			executeErrors:  []error{assert.AnError},
			expectedStatus: model.StatusTimeLimitExceeded,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock sandbox
			mockSandbox := new(MockSandbox)
			
			// Setup mock sandbox expectations
			mockSandbox.On("Compile", mock.Anything, tc.submission.Language, tc.submission.Code).
				Return(tc.compileOutput, tc.compileError)
			
			if tc.compileError == nil {
				for i, testCase := range tc.testCases {
					mockSandbox.On("Execute", mock.Anything, tc.submission.Language, tc.submission.Code, testCase.Input).
						Return(tc.executeOutputs[i], tc.executeTimes[i], tc.executeMemory[i], tc.executeErrors[i])
				}
			}
			
			// Create judging service with mock dependencies
			cfg := &config.Config{
				MaxExecutionTime: 10 * time.Second,
				MaxMemoryUsage:   512 * 1024 * 1024, // 512 MB
			}
			
			service := &JudgingService{
				cfg:     cfg,
				sandbox: mockSandbox,
			}
			
			// Call the function under test
			result, err := service.judgeSubmission(context.Background(), tc.submission, tc.testCases)
			
			// Verify expectations
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, result.Status)
			assert.Equal(t, tc.submission.ID, result.SubmissionID)
			
			// Verify that all test results are included
			if tc.compileError == nil {
				assert.Equal(t, len(tc.testCases), len(result.TestResults))
			}
			
			// Verify mock expectations
			mockSandbox.AssertExpectations(t)
		})
	}
}

// TestDetermineStatus tests the determineStatus function
func TestDetermineStatus(t *testing.T) {
	// Define test cases
	tests := []struct {
		name             string
		testResults      []model.TestResult
		executionTime    time.Duration
		memoryUsed       int64
		maxExecutionTime time.Duration
		maxMemoryUsage   int64
		expectedStatus   model.Status
	}{
		{
			name: "All tests pass",
			testResults: []model.TestResult{
				{TestCaseID: "1", Passed: true},
				{TestCaseID: "2", Passed: true},
			},
			executionTime:    100 * time.Millisecond,
			memoryUsed:       1024,
			maxExecutionTime: 10 * time.Second,
			maxMemoryUsage:   512 * 1024 * 1024,
			expectedStatus:   model.StatusAccepted,
		},
		{
			name: "Some tests fail",
			testResults: []model.TestResult{
				{TestCaseID: "1", Passed: true},
				{TestCaseID: "2", Passed: false},
			},
			executionTime:    100 * time.Millisecond,
			memoryUsed:       1024,
			maxExecutionTime: 10 * time.Second,
			maxMemoryUsage:   512 * 1024 * 1024,
			expectedStatus:   model.StatusRejected,
		},
		{
			name: "Time limit exceeded",
			testResults: []model.TestResult{
				{TestCaseID: "1", Passed: true},
			},
			executionTime:    11 * time.Second,
			memoryUsed:       1024,
			maxExecutionTime: 10 * time.Second,
			maxMemoryUsage:   512 * 1024 * 1024,
			expectedStatus:   model.StatusTimeLimitExceeded,
		},
		{
			name: "Memory limit exceeded",
			testResults: []model.TestResult{
				{TestCaseID: "1", Passed: true},
			},
			executionTime:    100 * time.Millisecond,
			memoryUsed:       513 * 1024 * 1024,
			maxExecutionTime: 10 * time.Second,
			maxMemoryUsage:   512 * 1024 * 1024,
			expectedStatus:   model.StatusMemoryLimitExceeded,
		},
		{
			name: "Runtime error",
			testResults: []model.TestResult{
				{TestCaseID: "1", Passed: false, Error: "runtime error"},
			},
			executionTime:    100 * time.Millisecond,
			memoryUsed:       1024,
			maxExecutionTime: 10 * time.Second,
			maxMemoryUsage:   512 * 1024 * 1024,
			expectedStatus:   model.StatusRuntimeError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status := determineStatus(tc.testResults, tc.executionTime, tc.memoryUsed, tc.maxExecutionTime, tc.maxMemoryUsage)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

// TestCompareOutput tests the compareOutput function
func TestCompareOutput(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		actual   string
		expected string
		result   bool
	}{
		{
			name:     "Exact match",
			actual:   "Hello, World!",
			expected: "Hello, World!",
			result:   true,
		},
		{
			name:     "Different output",
			actual:   "Hello, World!",
			expected: "Hello, Universe!",
			result:   false,
		},
		{
			name:     "Case sensitive",
			actual:   "hello, world!",
			expected: "Hello, World!",
			result:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := compareOutput(tc.actual, tc.expected)
			assert.Equal(t, tc.result, result)
		})
	}
}
