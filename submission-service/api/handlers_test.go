package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nslaughter/codecourt/submission-service/model"
	"github.com/nslaughter/codecourt/submission-service/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubmissionService is a mock implementation of the SubmissionServiceInterface
type MockSubmissionService struct {
	mock.Mock
}

// Ensure MockSubmissionService implements SubmissionServiceInterface
var _ service.SubmissionServiceInterface = (*MockSubmissionService)(nil)

func (m *MockSubmissionService) CreateSubmission(submission *model.Submission) error {
	args := m.Called(submission)
	return args.Error(0)
}

func (m *MockSubmissionService) GetSubmission(id string) (*model.Submission, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Submission), args.Error(1)
}

func (m *MockSubmissionService) GetSubmissionResult(submissionID string) (*model.SubmissionResult, error) {
	args := m.Called(submissionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SubmissionResult), args.Error(1)
}

func (m *MockSubmissionService) GetSubmissionsByUserID(userID string) ([]*model.Submission, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Submission), args.Error(1)
}

func (m *MockSubmissionService) GetSubmissionsByProblemID(problemID string) ([]*model.Submission, error) {
	args := m.Called(problemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Submission), args.Error(1)
}

func TestCreateSubmission(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		requestBody    interface{}
		serviceError   error
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: model.SubmissionRequest{
				ProblemID: uuid.New().String(),
				UserID:    uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
			},
			serviceError:   nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing Required Fields",
			requestBody: model.SubmissionRequest{
				ProblemID: "",
				UserID:    "",
				Language:  model.LanguageGo,
				Code:      "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
			},
			serviceError:   nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			requestBody: model.SubmissionRequest{
				ProblemID: uuid.New().String(),
				UserID:    uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
			},
			serviceError:   fmt.Errorf("service error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid Request Body",
			requestBody:    "invalid",
			serviceError:   nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockSubmissionService)

			// Set up expectations for all cases that call the service
			if tc.expectedStatus == http.StatusCreated || tc.expectedStatus == http.StatusInternalServerError {
				mockService.On("CreateSubmission", mock.AnythingOfType("*model.Submission")).Return(tc.serviceError)
			}

			// Create handler
			handler := NewHandler(mockService)

			// Create request
			var body []byte
			var err error
			if str, ok := tc.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tc.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/api/v1/submissions", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.CreateSubmission(rr, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Verify mock
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetSubmission(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		submissionID   string
		submission     *model.Submission
		serviceError   error
		expectedStatus int
	}{
		{
			name:         "Success",
			submissionID: uuid.New().String(),
			submission: &model.Submission{
				ID:        uuid.New().String(),
				ProblemID: uuid.New().String(),
				UserID:    uuid.New().String(),
				Language:  model.LanguageGo,
				Code:      "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
				Status:    model.SubmissionStatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			serviceError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Not Found",
			submissionID:   uuid.New().String(),
			submission:     nil,
			serviceError:   fmt.Errorf("not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockSubmissionService)

			// Set up expectations
			mockService.On("GetSubmission", tc.submissionID).Return(tc.submission, tc.serviceError)

			// Create handler
			handler := NewHandler(mockService)

			// Create request
			req, err := http.NewRequest("GET", "/api/v1/submissions/"+tc.submissionID, nil)
			assert.NoError(t, err)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create router and add route
			router := mux.NewRouter()
			router.HandleFunc("/api/v1/submissions/{id}", handler.GetSubmission).Methods("GET")

			// Call handler
			router.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Verify mock
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetSubmissionResult(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		submissionID   string
		result         *model.SubmissionResult
		serviceError   error
		expectedStatus int
	}{
		{
			name:         "Success",
			submissionID: uuid.New().String(),
			result: &model.SubmissionResult{
				ID:           uuid.New().String(),
				SubmissionID: uuid.New().String(),
				Status:       model.SubmissionStatusCompleted,
				CreatedAt:    time.Now(),
				TestCaseResults: []model.TestCaseResult{
					{
						ID:         uuid.New().String(),
						TestCaseID: uuid.New().String(),
						Status:     model.TestCaseStatusPassed,
						CreatedAt:  time.Now(),
					},
				},
			},
			serviceError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Not Found",
			submissionID:   uuid.New().String(),
			result:         nil,
			serviceError:   fmt.Errorf("not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockSubmissionService)

			// Set up expectations
			mockService.On("GetSubmissionResult", tc.submissionID).Return(tc.result, tc.serviceError)

			// Create handler
			handler := NewHandler(mockService)

			// Create request
			req, err := http.NewRequest("GET", "/api/v1/submissions/"+tc.submissionID+"/result", nil)
			assert.NoError(t, err)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create router and add route
			router := mux.NewRouter()
			router.HandleFunc("/api/v1/submissions/{id}/result", handler.GetSubmissionResult).Methods("GET")

			// Call handler
			router.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Verify mock
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetSubmissionsByUserID(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		userID         string
		submissions    []*model.Submission
		serviceError   error
		expectedStatus int
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
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			serviceError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Service Error",
			userID:         uuid.New().String(),
			submissions:    nil,
			serviceError:   fmt.Errorf("service error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockSubmissionService)

			// Set up expectations
			mockService.On("GetSubmissionsByUserID", tc.userID).Return(tc.submissions, tc.serviceError)

			// Create handler
			handler := NewHandler(mockService)

			// Create request
			req, err := http.NewRequest("GET", "/api/v1/users/"+tc.userID+"/submissions", nil)
			assert.NoError(t, err)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create router and add route
			router := mux.NewRouter()
			router.HandleFunc("/api/v1/users/{user_id}/submissions", handler.GetSubmissionsByUserID).Methods("GET")

			// Call handler
			router.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Verify mock
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetSubmissionsByProblemID(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		problemID      string
		submissions    []*model.Submission
		serviceError   error
		expectedStatus int
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
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			serviceError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Service Error",
			problemID:      uuid.New().String(),
			submissions:    nil,
			serviceError:   fmt.Errorf("service error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockSubmissionService)

			// Set up expectations
			mockService.On("GetSubmissionsByProblemID", tc.problemID).Return(tc.submissions, tc.serviceError)

			// Create handler
			handler := NewHandler(mockService)

			// Create request
			req, err := http.NewRequest("GET", "/api/v1/problems/"+tc.problemID+"/submissions", nil)
			assert.NoError(t, err)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create router and add route
			router := mux.NewRouter()
			router.HandleFunc("/api/v1/problems/{problem_id}/submissions", handler.GetSubmissionsByProblemID).Methods("GET")

			// Call handler
			router.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Verify mock
			mockService.AssertExpectations(t)
		})
	}
}
