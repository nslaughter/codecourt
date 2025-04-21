package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/submission-service/config"
	"github.com/nslaughter/codecourt/submission-service/db"
	kafkalib "github.com/nslaughter/codecourt/submission-service/kafka"
	"github.com/nslaughter/codecourt/submission-service/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDB is a mock implementation of the Repository interface
type MockDB struct {
	mock.Mock
}

// Ensure MockDB implements Repository interface
var _ db.Repository = (*MockDB)(nil)

func (m *MockDB) CreateSubmission(submission *model.Submission) error {
	args := m.Called(submission)
	return args.Error(0)
}

func (m *MockDB) GetSubmission(id string) (*model.Submission, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Submission), args.Error(1)
}

func (m *MockDB) UpdateSubmissionStatus(id string, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockDB) SaveSubmissionResult(result *model.SubmissionResult) error {
	args := m.Called(result)
	return args.Error(0)
}

func (m *MockDB) GetSubmissionsByUserID(userID string) ([]*model.Submission, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Submission), args.Error(1)
}

func (m *MockDB) GetSubmissionsByProblemID(problemID string) ([]*model.Submission, error) {
	args := m.Called(problemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Submission), args.Error(1)
}

func (m *MockDB) GetSubmissionResult(submissionID string) (*model.SubmissionResult, error) {
	args := m.Called(submissionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SubmissionResult), args.Error(1)
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockProducer is a mock implementation of the KafkaProducer interface
type MockProducer struct {
	mock.Mock
}

// Ensure MockProducer implements KafkaProducer interface
var _ kafkalib.KafkaProducer = (*MockProducer)(nil)

func (m *MockProducer) Produce(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockProducer) Close() {
	m.Called()
}

// MockConsumer is a mock implementation of the KafkaConsumer interface
type MockConsumer struct {
	mock.Mock
}

// Ensure MockConsumer implements KafkaConsumer interface
var _ kafkalib.KafkaConsumer = (*MockConsumer)(nil)

func (m *MockConsumer) Consume(timeout time.Duration) (*kafka.Message, error) {
	args := m.Called(timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kafka.Message), args.Error(1)
}

func (m *MockConsumer) CommitMessage(msg *kafka.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestCreateSubmission(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		submission    *model.Submission
		dbError       error
		produceError  error
		expectedError bool
	}{
		{
			name: "Success",
			submission: &model.Submission{
				ID:        uuid.New().String(),
				ProblemID: uuid.New().String(),
				UserID:    uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
				Status:    model.SubmissionStatusPending,
			},
			dbError:       nil,
			produceError:  nil,
			expectedError: false,
		},
		{
			name: "DB Error",
			submission: &model.Submission{
				ID:        uuid.New().String(),
				ProblemID: uuid.New().String(),
				UserID:    uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
				Status:    model.SubmissionStatusPending,
			},
			dbError:       assert.AnError,
			produceError:  nil,
			expectedError: true,
		},
		{
			name: "Produce Error",
			submission: &model.Submission{
				ID:        uuid.New().String(),
				ProblemID: uuid.New().String(),
				UserID:    uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
				Status:    model.SubmissionStatusPending,
			},
			dbError:       nil,
			produceError:  assert.AnError,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockDB := new(MockDB)
			mockProducer := new(MockProducer)
			mockConsumer := new(MockConsumer)

			// Set up expectations
			mockDB.On("CreateSubmission", tc.submission).Return(tc.dbError)
			if tc.dbError == nil {
				submissionJSON, _ := json.Marshal(tc.submission)
				mockProducer.On("Produce", tc.submission.ID, submissionJSON).Return(tc.produceError)
			}

			// Create service
			service := NewSubmissionService(&config.Config{}, mockDB, mockProducer, mockConsumer)

			// Call method
			err := service.CreateSubmission(tc.submission)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockDB.AssertExpectations(t)
			mockProducer.AssertExpectations(t)
		})
	}
}

func TestGetSubmission(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		id            string
		submission    *model.Submission
		dbError       error
		expectedError bool
	}{
		{
			name: "Success",
			id:   uuid.New().String(),
			submission: &model.Submission{
				ID:        uuid.New().String(),
				ProblemID: uuid.New().String(),
				UserID:    uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
				Status:    model.SubmissionStatusPending,
			},
			dbError:       nil,
			expectedError: false,
		},
		{
			name:          "DB Error",
			id:            uuid.New().String(),
			submission:    nil,
			dbError:       assert.AnError,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockDB := new(MockDB)
			mockProducer := new(MockProducer)
			mockConsumer := new(MockConsumer)

			// Set up expectations
			mockDB.On("GetSubmission", tc.id).Return(tc.submission, tc.dbError)

			// Create service
			service := NewSubmissionService(&config.Config{}, mockDB, mockProducer, mockConsumer)

			// Call method
			submission, err := service.GetSubmission(tc.id)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, submission)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.submission, submission)
			}

			// Verify mocks
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetSubmissionResult(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		submissionID  string
		result        *model.SubmissionResult
		dbError       error
		expectedError bool
	}{
		{
			name:         "Success",
			submissionID: uuid.New().String(),
			result: &model.SubmissionResult{
				ID:           uuid.New().String(),
				SubmissionID: uuid.New().String(),
				Status:       model.SubmissionStatusCompleted,
				TestCaseResults: []model.TestCaseResult{
					{
						ID:         uuid.New().String(),
						TestCaseID: uuid.New().String(),
						Status:     model.TestCaseStatusPassed,
					},
				},
			},
			dbError:       nil,
			expectedError: false,
		},
		{
			name:          "DB Error",
			submissionID:  uuid.New().String(),
			result:        nil,
			dbError:       assert.AnError,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockDB := new(MockDB)
			mockProducer := new(MockProducer)
			mockConsumer := new(MockConsumer)

			// Set up expectations
			mockDB.On("GetSubmissionResult", tc.submissionID).Return(tc.result, tc.dbError)

			// Create service
			service := NewSubmissionService(&config.Config{}, mockDB, mockProducer, mockConsumer)

			// Call method
			result, err := service.GetSubmissionResult(tc.submissionID)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.result, result)
			}

			// Verify mocks
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetSubmissionsByUserID(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		userID        string
		submissions   []*model.Submission
		dbError       error
		expectedError bool
	}{
		{
			name:   "Success",
			userID: uuid.New().String(),
			submissions: []*model.Submission{
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					UserID:    uuid.New().String(),
					Language:  model.LanguageGo,
					Code:      "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
					Status:    model.SubmissionStatusPending,
				},
			},
			dbError:       nil,
			expectedError: false,
		},
		{
			name:          "DB Error",
			userID:        uuid.New().String(),
			submissions:   nil,
			dbError:       assert.AnError,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockDB := new(MockDB)
			mockProducer := new(MockProducer)
			mockConsumer := new(MockConsumer)

			// Set up expectations
			mockDB.On("GetSubmissionsByUserID", tc.userID).Return(tc.submissions, tc.dbError)

			// Create service
			service := NewSubmissionService(&config.Config{}, mockDB, mockProducer, mockConsumer)

			// Call method
			submissions, err := service.GetSubmissionsByUserID(tc.userID)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, submissions)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.submissions, submissions)
			}

			// Verify mocks
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetSubmissionsByProblemID(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		problemID     string
		submissions   []*model.Submission
		dbError       error
		expectedError bool
	}{
		{
			name:      "Success",
			problemID: uuid.New().String(),
			submissions: []*model.Submission{
				{
					ID:        uuid.New().String(),
					ProblemID: uuid.New().String(),
					UserID:    uuid.New().String(),
					Language:  model.LanguageGo,
					Code:      "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
					Status:    model.SubmissionStatusPending,
				},
			},
			dbError:       nil,
			expectedError: false,
		},
		{
			name:          "DB Error",
			problemID:     uuid.New().String(),
			submissions:   nil,
			dbError:       assert.AnError,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockDB := new(MockDB)
			mockProducer := new(MockProducer)
			mockConsumer := new(MockConsumer)

			// Set up expectations
			mockDB.On("GetSubmissionsByProblemID", tc.problemID).Return(tc.submissions, tc.dbError)

			// Create service
			service := NewSubmissionService(&config.Config{}, mockDB, mockProducer, mockConsumer)

			// Call method
			submissions, err := service.GetSubmissionsByProblemID(tc.problemID)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, submissions)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.submissions, submissions)
			}

			// Verify mocks
			mockDB.AssertExpectations(t)
		})
	}
}
