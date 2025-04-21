package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/notification-service/config"
	"github.com/nslaughter/codecourt/notification-service/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNotificationRepository is a mock implementation of the NotificationRepository interface
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) CreateNotification(notification *model.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetNotificationByID(id uuid.UUID) (*model.Notification, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetNotificationsByUserID(userID uuid.UUID, limit, offset int) ([]*model.Notification, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetUnreadNotificationsByUserID(userID uuid.UUID, limit, offset int) ([]*model.Notification, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) UpdateNotificationStatus(id uuid.UUID, status model.NotificationStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkNotificationAsRead(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockNotificationRepository) DeleteNotification(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockNotificationRepository) CreateTemplate(template *model.NotificationTemplate) error {
	args := m.Called(template)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetTemplateByID(id string) (*model.NotificationTemplate, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.NotificationTemplate), args.Error(1)
}

func (m *MockNotificationRepository) GetTemplatesByEventType(eventType model.EventType) ([]*model.NotificationTemplate, error) {
	args := m.Called(eventType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.NotificationTemplate), args.Error(1)
}

func (m *MockNotificationRepository) UpdateTemplate(template *model.NotificationTemplate) error {
	args := m.Called(template)
	return args.Error(0)
}

func (m *MockNotificationRepository) DeleteTemplate(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockNotificationRepository) CreatePreference(preference *model.NotificationPreference) error {
	args := m.Called(preference)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetPreferenceByUserIDAndEventType(userID uuid.UUID, eventType model.EventType) (*model.NotificationPreference, error) {
	args := m.Called(userID, eventType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.NotificationPreference), args.Error(1)
}

func (m *MockNotificationRepository) GetPreferencesByUserID(userID uuid.UUID) ([]*model.NotificationPreference, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.NotificationPreference), args.Error(1)
}

func (m *MockNotificationRepository) UpdatePreference(preference *model.NotificationPreference) error {
	args := m.Called(preference)
	return args.Error(0)
}

func (m *MockNotificationRepository) DeletePreference(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestSendNotification(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		request       *model.NotificationRequest
		setupMock     func(*MockNotificationRepository)
		expectedError error
	}{
		{
			name: "Send in-app notification successfully",
			request: &model.NotificationRequest{
				UserID:  uuid.New(),
				Type:    model.NotificationTypeInApp,
				Title:   "Test Notification",
				Content: "This is a test notification",
			},
			setupMock: func(mockRepo *MockNotificationRepository) {
				mockRepo.On("CreateNotification", mock.AnythingOfType("*model.Notification")).Return(nil)
				mockRepo.On("UpdateNotificationStatus", mock.AnythingOfType("uuid.UUID"), model.NotificationStatusSent).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Send notification with template",
			request: &model.NotificationRequest{
				UserID:      uuid.New(),
				Type:        model.NotificationTypeInApp,
				TemplateID:  "test-template",
				TemplateData: map[string]interface{}{"name": "Test User"},
			},
			setupMock: func(mockRepo *MockNotificationRepository) {
				template := &model.NotificationTemplate{
					ID:      "test-template",
					Name:    "Test Template",
					Type:    model.NotificationTypeInApp,
					Subject: "Hello {{.name}}",
					Content: "Welcome to the system, {{.name}}!",
				}
				mockRepo.On("GetTemplateByID", "test-template").Return(template, nil)
				mockRepo.On("CreateNotification", mock.AnythingOfType("*model.Notification")).Return(nil)
				mockRepo.On("UpdateNotificationStatus", mock.AnythingOfType("uuid.UUID"), model.NotificationStatusSent).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Template not found",
			request: &model.NotificationRequest{
				UserID:     uuid.New(),
				Type:       model.NotificationTypeInApp,
				TemplateID: "non-existent-template",
			},
			setupMock: func(mockRepo *MockNotificationRepository) {
				mockRepo.On("GetTemplateByID", "non-existent-template").Return(nil, nil)
			},
			expectedError: ErrTemplateNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockNotificationRepository)
			
			// Setup mock
			tc.setupMock(mockRepo)
			
			// Create service
			cfg := &config.Config{}
			service := NewNotificationService(mockRepo, cfg)
			
			// Call the method
			notification, err := service.SendNotification(tc.request)
			
			// Check the result
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, notification)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, notification)
				
				// Verify mock expectations
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestGetNotificationByID(t *testing.T) {
	// Create test notification
	notificationID := uuid.New()
	testNotification := &model.Notification{
		ID:        notificationID,
		UserID:    uuid.New(),
		Type:      model.NotificationTypeInApp,
		Title:     "Test Notification",
		Content:   "This is a test notification",
		Status:    model.NotificationStatusSent,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	
	// Test cases
	testCases := []struct {
		name          string
		id            uuid.UUID
		setupMock     func(*MockNotificationRepository)
		expectedError error
	}{
		{
			name: "Get notification successfully",
			id:   notificationID,
			setupMock: func(mockRepo *MockNotificationRepository) {
				mockRepo.On("GetNotificationByID", notificationID).Return(testNotification, nil)
			},
			expectedError: nil,
		},
		{
			name: "Notification not found",
			id:   uuid.New(),
			setupMock: func(mockRepo *MockNotificationRepository) {
				mockRepo.On("GetNotificationByID", mock.AnythingOfType("uuid.UUID")).Return(nil, nil)
			},
			expectedError: ErrNotificationNotFound,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockNotificationRepository)
			
			// Setup mock
			tc.setupMock(mockRepo)
			
			// Create service
			cfg := &config.Config{}
			service := NewNotificationService(mockRepo, cfg)
			
			// Call the method
			notification, err := service.GetNotificationByID(tc.id)
			
			// Check the result
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, notification)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, notification)
				assert.Equal(t, testNotification.ID, notification.ID)
				assert.Equal(t, testNotification.UserID, notification.UserID)
				assert.Equal(t, testNotification.Title, notification.Title)
				assert.Equal(t, testNotification.Content, notification.Content)
				assert.Equal(t, testNotification.Status, notification.Status)
				
				// Verify mock expectations
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestHandleEvent(t *testing.T) {
	// Create test data
	userID := uuid.New()
	eventType := model.EventTypeSubmissionJudged
	
	// Test cases
	testCases := []struct {
		name      string
		event     *model.Event
		setupMock func(*MockNotificationRepository)
	}{
		{
			name: "Handle event with template and preference",
			event: &model.Event{
				ID:   "test-event",
				Type: eventType,
				Data: map[string]interface{}{
					"user_id":        userID.String(),
					"submission_id":  "123",
					"problem_name":   "Test Problem",
					"status":         "Accepted",
					"execution_time": 100,
				},
				Timestamp: time.Now().UTC(),
			},
			setupMock: func(mockRepo *MockNotificationRepository) {
				// Templates for the event
				templates := []*model.NotificationTemplate{
					{
						ID:        "in-app-template",
						Name:      "Submission Judged In-App",
						EventType: eventType,
						Type:      model.NotificationTypeInApp,
						Subject:   "Submission Judged: {{.status}}",
						Content:   "Your submission for {{.problem_name}} has been judged as {{.status}}",
					},
				}
				
				// User preference
				preference := &model.NotificationPreference{
					ID:        uuid.New(),
					UserID:    userID,
					EventType: eventType,
					Channels:  []model.NotificationType{model.NotificationTypeInApp},
					Enabled:   true,
				}
				
				mockRepo.On("GetTemplatesByEventType", eventType).Return(templates, nil)
				mockRepo.On("GetPreferenceByUserIDAndEventType", userID, eventType).Return(preference, nil)
				// Mock GetTemplateByID for each template
				for _, tmpl := range templates {
					mockRepo.On("GetTemplateByID", tmpl.ID).Return(tmpl, nil)
				}
				mockRepo.On("CreateNotification", mock.AnythingOfType("*model.Notification")).Return(nil)
				mockRepo.On("UpdateNotificationStatus", mock.AnythingOfType("uuid.UUID"), model.NotificationStatusSent).Return(nil)
			},
		},
		{
			name: "Handle event with no templates",
			event: &model.Event{
				ID:   "test-event",
				Type: eventType,
				Data: map[string]interface{}{
					"user_id": userID.String(),
				},
				Timestamp: time.Now().UTC(),
			},
			setupMock: func(mockRepo *MockNotificationRepository) {
				mockRepo.On("GetTemplatesByEventType", eventType).Return([]*model.NotificationTemplate{}, nil)
			},
		},
		{
			name: "Handle event with disabled preference",
			event: &model.Event{
				ID:   "test-event",
				Type: eventType,
				Data: map[string]interface{}{
					"user_id": userID.String(),
				},
				Timestamp: time.Now().UTC(),
			},
			setupMock: func(mockRepo *MockNotificationRepository) {
				// Templates for the event
				templates := []*model.NotificationTemplate{
					{
						ID:        "in-app-template",
						Name:      "Submission Judged In-App",
						EventType: eventType,
						Type:      model.NotificationTypeInApp,
						Subject:   "Submission Judged",
						Content:   "Your submission has been judged",
					},
				}
				
				// User preference with notifications disabled
				preference := &model.NotificationPreference{
					ID:        uuid.New(),
					UserID:    userID,
					EventType: eventType,
					Channels:  []model.NotificationType{model.NotificationTypeInApp},
					Enabled:   false,
				}
				
				mockRepo.On("GetTemplatesByEventType", eventType).Return(templates, nil)
				mockRepo.On("GetPreferenceByUserIDAndEventType", userID, eventType).Return(preference, nil)
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockNotificationRepository)
			
			// Setup mock
			tc.setupMock(mockRepo)
			
			// Create service
			cfg := &config.Config{}
			service := NewNotificationService(mockRepo, cfg)
			
			// Call the method
			err := service.HandleEvent(tc.event)
			
			// Check the result
			assert.NoError(t, err)
			
			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestApplyTemplate(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		template       *model.NotificationTemplate
		data           map[string]interface{}
		expectedTitle  string
		expectedContent string
		expectedError  bool
	}{
		{
			name: "Valid template",
			template: &model.NotificationTemplate{
				Subject: "Hello {{.name}}",
				Content: "Welcome to the system, {{.name}}!",
			},
			data: map[string]interface{}{
				"name": "Test User",
			},
			expectedTitle:  "Hello Test User",
			expectedContent: "Welcome to the system, Test User!",
			expectedError:  false,
		},
		{
			name: "Invalid template syntax",
			template: &model.NotificationTemplate{
				Subject: "Hello {{.name",
				Content: "Welcome to the system!",
			},
			data:          map[string]interface{}{},
			expectedError: true,
		},
		{
			name: "Missing template variable",
			template: &model.NotificationTemplate{
				Subject: "Hello {{if .name}}{{.name}}{{else}}User{{end}}",
				Content: "Welcome to the system!",
			},
			data:          map[string]interface{}{},
			expectedTitle: "Hello User",
			expectedContent: "Welcome to the system!",
			expectedError: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create service
			cfg := &config.Config{}
			service := NewNotificationService(nil, cfg)
			
			// Call the method
			title, content, err := service.applyTemplate(tc.template, tc.data)
			
			// Check the result
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTitle, title)
				assert.Equal(t, tc.expectedContent, content)
			}
		})
	}
}
