package service

import (
	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/notification-service/model"
)

// NotificationService defines the interface for notification service operations
type NotificationService interface {
	// Notification operations
	SendNotification(req *model.NotificationRequest) (*model.NotificationResponse, error)
	SendBatchNotifications(req *model.BatchNotificationRequest) ([]uuid.UUID, error)
	GetNotificationByID(id uuid.UUID) (*model.NotificationResponse, error)
	GetNotificationsByUserID(userID uuid.UUID, limit, offset int) ([]*model.NotificationResponse, error)
	GetUnreadNotificationsByUserID(userID uuid.UUID, limit, offset int) ([]*model.NotificationResponse, error)
	MarkNotificationAsRead(id uuid.UUID) error
	DeleteNotification(id uuid.UUID) error
	
	// Template operations
	CreateTemplate(template *model.NotificationTemplate) error
	GetTemplateByID(id string) (*model.NotificationTemplate, error)
	GetTemplatesByEventType(eventType model.EventType) ([]*model.NotificationTemplate, error)
	UpdateTemplate(template *model.NotificationTemplate) error
	DeleteTemplate(id string) error
	
	// Preference operations
	SetPreference(userID uuid.UUID, req *model.NotificationPreferenceRequest) error
	GetPreferencesByUserID(userID uuid.UUID) ([]*model.NotificationPreference, error)
	
	// Event handling
	HandleEvent(event *model.Event) error
}
