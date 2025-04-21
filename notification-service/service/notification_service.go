package service

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/notification-service/config"
	"github.com/nslaughter/codecourt/notification-service/db"
	"github.com/nslaughter/codecourt/notification-service/model"
	"gopkg.in/gomail.v2"
)

// Common errors
var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrTemplateNotFound     = errors.New("template not found")
	ErrInvalidTemplate      = errors.New("invalid template")
	ErrSendingNotification  = errors.New("error sending notification")
)

// NotificationServiceImpl implements the NotificationService interface
type NotificationServiceImpl struct {
	repo db.NotificationRepository
	cfg  *config.Config
}

// NewNotificationService creates a new notification service
func NewNotificationService(repo db.NotificationRepository, cfg *config.Config) *NotificationServiceImpl {
	return &NotificationServiceImpl{
		repo: repo,
		cfg:  cfg,
	}
}

// SendNotification sends a notification to a user
func (s *NotificationServiceImpl) SendNotification(req *model.NotificationRequest) (*model.NotificationResponse, error) {
	// Create notification
	now := time.Now().UTC()
	notification := &model.Notification{
		ID:          uuid.New(),
		UserID:      req.UserID,
		Type:        req.Type,
		Title:       req.Title,
		Content:     req.Content,
		Status:      model.NotificationStatusPending,
		EventType:   req.EventType,
		EventID:     req.EventID,
		CreatedAt:   now,
		UpdatedAt:   now,
		TemplateID:  req.TemplateID,
		TemplateData: req.TemplateData,
	}

	// If template ID is provided, apply the template
	if req.TemplateID != "" {
		template, err := s.repo.GetTemplateByID(req.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving template: %w", err)
		}
		if template == nil {
			return nil, ErrTemplateNotFound
		}

		// Apply template
		title, content, err := s.applyTemplate(template, req.TemplateData)
		if err != nil {
			return nil, fmt.Errorf("error applying template: %w", err)
		}

		notification.Title = title
		notification.Content = content
	}

	// Save notification to database
	if err := s.repo.CreateNotification(notification); err != nil {
		return nil, fmt.Errorf("error creating notification: %w", err)
	}

	// Send notification based on type
	var err error
	switch notification.Type {
	case model.NotificationTypeEmail:
		err = s.sendEmailNotification(notification)
	case model.NotificationTypeInApp:
		// In-app notifications are just stored in the database
		err = s.repo.UpdateNotificationStatus(notification.ID, model.NotificationStatusSent)
	default:
		err = fmt.Errorf("unsupported notification type: %s", notification.Type)
	}

	if err != nil {
		// Update status to failed
		s.repo.UpdateNotificationStatus(notification.ID, model.NotificationStatusFailed)
		return nil, fmt.Errorf("%w: %v", ErrSendingNotification, err)
	}

	return model.NewNotificationResponse(notification), nil
}

// SendBatchNotifications sends notifications to multiple users
func (s *NotificationServiceImpl) SendBatchNotifications(req *model.BatchNotificationRequest) ([]uuid.UUID, error) {
	var notificationIDs []uuid.UUID

	for _, userID := range req.UserIDs {
		// Create notification request for each user
		notificationReq := &model.NotificationRequest{
			UserID:       userID,
			Type:         req.Type,
			Title:        req.Title,
			Content:      req.Content,
			EventType:    req.EventType,
			EventID:      req.EventID,
			TemplateID:   req.TemplateID,
			TemplateData: req.TemplateData,
		}

		// Send notification
		notification, err := s.SendNotification(notificationReq)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Error sending notification to user %s: %v\n", userID, err)
			continue
		}

		notificationIDs = append(notificationIDs, notification.ID)
	}

	return notificationIDs, nil
}

// GetNotificationByID retrieves a notification by ID
func (s *NotificationServiceImpl) GetNotificationByID(id uuid.UUID) (*model.NotificationResponse, error) {
	notification, err := s.repo.GetNotificationByID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving notification: %w", err)
	}
	if notification == nil {
		return nil, ErrNotificationNotFound
	}

	return model.NewNotificationResponse(notification), nil
}

// GetNotificationsByUserID retrieves notifications for a user
func (s *NotificationServiceImpl) GetNotificationsByUserID(userID uuid.UUID, limit, offset int) ([]*model.NotificationResponse, error) {
	notifications, err := s.repo.GetNotificationsByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error retrieving notifications: %w", err)
	}

	// Convert to response objects
	responses := make([]*model.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = model.NewNotificationResponse(notification)
	}

	return responses, nil
}

// GetUnreadNotificationsByUserID retrieves unread notifications for a user
func (s *NotificationServiceImpl) GetUnreadNotificationsByUserID(userID uuid.UUID, limit, offset int) ([]*model.NotificationResponse, error) {
	notifications, err := s.repo.GetUnreadNotificationsByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error retrieving unread notifications: %w", err)
	}

	// Convert to response objects
	responses := make([]*model.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = model.NewNotificationResponse(notification)
	}

	return responses, nil
}

// MarkNotificationAsRead marks a notification as read
func (s *NotificationServiceImpl) MarkNotificationAsRead(id uuid.UUID) error {
	// Check if notification exists
	notification, err := s.repo.GetNotificationByID(id)
	if err != nil {
		return fmt.Errorf("error retrieving notification: %w", err)
	}
	if notification == nil {
		return ErrNotificationNotFound
	}

	// Mark as read
	if err := s.repo.MarkNotificationAsRead(id); err != nil {
		return fmt.Errorf("error marking notification as read: %w", err)
	}

	return nil
}

// DeleteNotification deletes a notification
func (s *NotificationServiceImpl) DeleteNotification(id uuid.UUID) error {
	// Check if notification exists
	notification, err := s.repo.GetNotificationByID(id)
	if err != nil {
		return fmt.Errorf("error retrieving notification: %w", err)
	}
	if notification == nil {
		return ErrNotificationNotFound
	}

	// Delete notification
	if err := s.repo.DeleteNotification(id); err != nil {
		return fmt.Errorf("error deleting notification: %w", err)
	}

	return nil
}

// CreateTemplate creates a new notification template
func (s *NotificationServiceImpl) CreateTemplate(template *model.NotificationTemplate) error {
	// Set created and updated timestamps
	now := time.Now().UTC()
	template.CreatedAt = now
	template.UpdatedAt = now

	// Validate template
	if _, _, err := s.applyTemplate(template, map[string]interface{}{}); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidTemplate, err)
	}

	// Save template
	if err := s.repo.CreateTemplate(template); err != nil {
		return fmt.Errorf("error creating template: %w", err)
	}

	return nil
}

// GetTemplateByID retrieves a template by ID
func (s *NotificationServiceImpl) GetTemplateByID(id string) (*model.NotificationTemplate, error) {
	template, err := s.repo.GetTemplateByID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving template: %w", err)
	}
	if template == nil {
		return nil, ErrTemplateNotFound
	}

	return template, nil
}

// GetTemplatesByEventType retrieves templates by event type
func (s *NotificationServiceImpl) GetTemplatesByEventType(eventType model.EventType) ([]*model.NotificationTemplate, error) {
	templates, err := s.repo.GetTemplatesByEventType(eventType)
	if err != nil {
		return nil, fmt.Errorf("error retrieving templates: %w", err)
	}

	return templates, nil
}

// UpdateTemplate updates a notification template
func (s *NotificationServiceImpl) UpdateTemplate(template *model.NotificationTemplate) error {
	// Check if template exists
	existingTemplate, err := s.repo.GetTemplateByID(template.ID)
	if err != nil {
		return fmt.Errorf("error retrieving template: %w", err)
	}
	if existingTemplate == nil {
		return ErrTemplateNotFound
	}

	// Validate template
	if _, _, err := s.applyTemplate(template, map[string]interface{}{}); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidTemplate, err)
	}

	// Update template
	template.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateTemplate(template); err != nil {
		return fmt.Errorf("error updating template: %w", err)
	}

	return nil
}

// DeleteTemplate deletes a notification template
func (s *NotificationServiceImpl) DeleteTemplate(id string) error {
	// Check if template exists
	template, err := s.repo.GetTemplateByID(id)
	if err != nil {
		return fmt.Errorf("error retrieving template: %w", err)
	}
	if template == nil {
		return ErrTemplateNotFound
	}

	// Delete template
	if err := s.repo.DeleteTemplate(id); err != nil {
		return fmt.Errorf("error deleting template: %w", err)
	}

	return nil
}

// SetPreference sets a notification preference for a user
func (s *NotificationServiceImpl) SetPreference(userID uuid.UUID, req *model.NotificationPreferenceRequest) error {
	// Check if preference exists
	preference, err := s.repo.GetPreferenceByUserIDAndEventType(userID, req.EventType)
	if err != nil {
		return fmt.Errorf("error retrieving preference: %w", err)
	}

	now := time.Now().UTC()
	if preference == nil {
		// Create new preference
		preference = &model.NotificationPreference{
			ID:        uuid.New(),
			UserID:    userID,
			EventType: req.EventType,
			Channels:  req.Channels,
			Enabled:   req.Enabled,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := s.repo.CreatePreference(preference); err != nil {
			return fmt.Errorf("error creating preference: %w", err)
		}
	} else {
		// Update existing preference
		preference.Channels = req.Channels
		preference.Enabled = req.Enabled
		preference.UpdatedAt = now

		if err := s.repo.UpdatePreference(preference); err != nil {
			return fmt.Errorf("error updating preference: %w", err)
		}
	}

	return nil
}

// GetPreferencesByUserID retrieves preferences for a user
func (s *NotificationServiceImpl) GetPreferencesByUserID(userID uuid.UUID) ([]*model.NotificationPreference, error) {
	preferences, err := s.repo.GetPreferencesByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving preferences: %w", err)
	}

	return preferences, nil
}

// HandleEvent handles an event and sends notifications
func (s *NotificationServiceImpl) HandleEvent(event *model.Event) error {
	// Get templates for this event type
	templates, err := s.repo.GetTemplatesByEventType(event.Type)
	if err != nil {
		return fmt.Errorf("error retrieving templates: %w", err)
	}

	if len(templates) == 0 {
		// No templates for this event type
		return nil
	}

	// Extract user ID from event data
	userIDStr, ok := event.Data["user_id"].(string)
	if !ok {
		return fmt.Errorf("event data missing user_id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user_id in event data: %w", err)
	}

	// Check user preferences
	preference, err := s.repo.GetPreferenceByUserIDAndEventType(userID, event.Type)
	if err != nil {
		return fmt.Errorf("error retrieving preference: %w", err)
	}

	// If no preference exists or notifications are disabled, use default channels
	var channels []model.NotificationType
	if preference == nil || preference.Enabled {
		if preference != nil {
			channels = preference.Channels
		} else {
			// Default to in-app notifications
			channels = []model.NotificationType{model.NotificationTypeInApp}
		}
	} else {
		// Notifications disabled for this event type
		return nil
	}

	// Send notifications for each template and channel
	for _, tmpl := range templates {
		for _, channel := range channels {
			// Skip if template type doesn't match channel
			if tmpl.Type != channel {
				continue
			}

			// Apply template
			title, content, err := s.applyTemplate(tmpl, event.Data)
			if err != nil {
				fmt.Printf("Error applying template %s: %v\n", tmpl.ID, err)
				continue
			}

			// Create notification request
			req := &model.NotificationRequest{
				UserID:      userID,
				Type:        channel,
				Title:       title,
				Content:     content,
				EventType:   event.Type,
				EventID:     event.ID,
				TemplateID:  tmpl.ID,
				TemplateData: event.Data,
			}

			// Send notification
			_, err = s.SendNotification(req)
			if err != nil {
				fmt.Printf("Error sending notification for event %s: %v\n", event.ID, err)
				continue
			}
		}
	}

	return nil
}

// applyTemplate applies a template with data
func (s *NotificationServiceImpl) applyTemplate(tmpl *model.NotificationTemplate, data map[string]interface{}) (string, string, error) {
	// Parse title template
	titleTmpl, err := template.New("title").Parse(tmpl.Subject)
	if err != nil {
		return "", "", fmt.Errorf("error parsing title template: %w", err)
	}

	// Parse content template
	contentTmpl, err := template.New("content").Parse(tmpl.Content)
	if err != nil {
		return "", "", fmt.Errorf("error parsing content template: %w", err)
	}

	// Execute title template
	var titleBuf bytes.Buffer
	if err := titleTmpl.Execute(&titleBuf, data); err != nil {
		return "", "", fmt.Errorf("error executing title template: %w", err)
	}

	// Execute content template
	var contentBuf bytes.Buffer
	if err := contentTmpl.Execute(&contentBuf, data); err != nil {
		return "", "", fmt.Errorf("error executing content template: %w", err)
	}

	return titleBuf.String(), contentBuf.String(), nil
}

// sendEmailNotification sends an email notification
func (s *NotificationServiceImpl) sendEmailNotification(notification *model.Notification) error {
	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.SMTPFrom)
	m.SetHeader("To", notification.UserID.String()+"@example.com") // In a real system, we would look up the user's email
	m.SetHeader("Subject", notification.Title)
	m.SetBody("text/html", notification.Content)

	// Create dialer
	d := gomail.NewDialer(s.cfg.SMTPHost, s.cfg.SMTPPort, s.cfg.SMTPUsername, s.cfg.SMTPPassword)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	// Update notification status
	now := time.Now().UTC()
	notification.Status = model.NotificationStatusSent
	notification.SentAt = &now
	notification.UpdatedAt = now

	if err := s.repo.UpdateNotificationStatus(notification.ID, model.NotificationStatusSent); err != nil {
		return fmt.Errorf("error updating notification status: %w", err)
	}

	return nil
}
