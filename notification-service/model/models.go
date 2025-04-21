package model

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

// Notification types
const (
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypeInApp   NotificationType = "in_app"
	NotificationTypeWebhook NotificationType = "webhook"
)

// EventType represents the type of event that triggered a notification
type EventType string

// Event types
const (
	EventTypeSubmissionCreated EventType = "submission_created"
	EventTypeSubmissionJudged  EventType = "submission_judged"
	EventTypeUserRegistered    EventType = "user_registered"
	EventTypeProblemCreated    EventType = "problem_created"
	EventTypeSystemAlert       EventType = "system_alert"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

// Notification statuses
const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// Notification represents a notification in the system
type Notification struct {
	ID          uuid.UUID          `json:"id"`
	UserID      uuid.UUID          `json:"user_id"`
	Type        NotificationType   `json:"type"`
	Title       string             `json:"title"`
	Content     string             `json:"content"`
	Status      NotificationStatus `json:"status"`
	EventType   EventType          `json:"event_type"`
	EventID     string             `json:"event_id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	SentAt      *time.Time         `json:"sent_at,omitempty"`
	ReadAt      *time.Time         `json:"read_at,omitempty"`
	TemplateID  string             `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
}

// NotificationTemplate represents a template for notifications
type NotificationTemplate struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	EventType   EventType        `json:"event_type"`
	Type        NotificationType `json:"type"`
	Subject     string           `json:"subject"`
	Content     string           `json:"content"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// NotificationPreference represents a user's notification preferences
type NotificationPreference struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	EventType EventType        `json:"event_type"`
	Channels  []NotificationType `json:"channels"`
	Enabled   bool             `json:"enabled"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// Event represents an event that can trigger notifications
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	UserID      uuid.UUID             `json:"user_id" validate:"required"`
	Type        NotificationType      `json:"type" validate:"required"`
	Title       string                `json:"title" validate:"required"`
	Content     string                `json:"content" validate:"required"`
	EventType   EventType             `json:"event_type,omitempty"`
	EventID     string                `json:"event_id,omitempty"`
	TemplateID  string                `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
}

// NotificationResponse represents a notification in API responses
type NotificationResponse struct {
	ID        uuid.UUID          `json:"id"`
	UserID    uuid.UUID          `json:"user_id"`
	Type      NotificationType   `json:"type"`
	Title     string             `json:"title"`
	Content   string             `json:"content"`
	Status    NotificationStatus `json:"status"`
	EventType EventType          `json:"event_type,omitempty"`
	EventID   string             `json:"event_id,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	SentAt    *time.Time         `json:"sent_at,omitempty"`
	ReadAt    *time.Time         `json:"read_at,omitempty"`
}

// NewNotificationResponse creates a new NotificationResponse from a Notification
func NewNotificationResponse(notification *Notification) *NotificationResponse {
	return &NotificationResponse{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Title:     notification.Title,
		Content:   notification.Content,
		Status:    notification.Status,
		EventType: notification.EventType,
		EventID:   notification.EventID,
		CreatedAt: notification.CreatedAt,
		SentAt:    notification.SentAt,
		ReadAt:    notification.ReadAt,
	}
}

// BatchNotificationRequest represents a request to send notifications to multiple users
type BatchNotificationRequest struct {
	UserIDs     []uuid.UUID           `json:"user_ids" validate:"required"`
	Type        NotificationType      `json:"type" validate:"required"`
	Title       string                `json:"title" validate:"required"`
	Content     string                `json:"content" validate:"required"`
	EventType   EventType             `json:"event_type,omitempty"`
	EventID     string                `json:"event_id,omitempty"`
	TemplateID  string                `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
}

// NotificationPreferenceRequest represents a request to update notification preferences
type NotificationPreferenceRequest struct {
	EventType EventType          `json:"event_type" validate:"required"`
	Channels  []NotificationType `json:"channels" validate:"required"`
	Enabled   bool               `json:"enabled"`
}
