package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/notification-service/model"
)

// NotificationRepository defines the interface for notification database operations
type NotificationRepository interface {
	// Notification operations
	CreateNotification(notification *model.Notification) error
	GetNotificationByID(id uuid.UUID) (*model.Notification, error)
	GetNotificationsByUserID(userID uuid.UUID, limit, offset int) ([]*model.Notification, error)
	GetUnreadNotificationsByUserID(userID uuid.UUID, limit, offset int) ([]*model.Notification, error)
	UpdateNotificationStatus(id uuid.UUID, status model.NotificationStatus) error
	MarkNotificationAsRead(id uuid.UUID) error
	DeleteNotification(id uuid.UUID) error
	
	// Template operations
	CreateTemplate(template *model.NotificationTemplate) error
	GetTemplateByID(id string) (*model.NotificationTemplate, error)
	GetTemplatesByEventType(eventType model.EventType) ([]*model.NotificationTemplate, error)
	UpdateTemplate(template *model.NotificationTemplate) error
	DeleteTemplate(id string) error
	
	// Preference operations
	CreatePreference(preference *model.NotificationPreference) error
	GetPreferenceByUserIDAndEventType(userID uuid.UUID, eventType model.EventType) (*model.NotificationPreference, error)
	GetPreferencesByUserID(userID uuid.UUID) ([]*model.NotificationPreference, error)
	UpdatePreference(preference *model.NotificationPreference) error
	DeletePreference(id uuid.UUID) error
}

// EnsureNotificationRepository ensures that DB implements NotificationRepository
var _ NotificationRepository = (*DB)(nil)

// CreateNotification creates a new notification in the database
func (db *DB) CreateNotification(notification *model.Notification) error {
	query := `
		INSERT INTO notifications (
			id, user_id, type, title, content, status, event_type, event_id, 
			created_at, updated_at, sent_at, read_at, template_id, template_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	
	templateData, err := json.Marshal(notification.TemplateData)
	if err != nil {
		return err
	}
	
	_, err = db.Exec(
		query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Content,
		notification.Status,
		notification.EventType,
		notification.EventID,
		notification.CreatedAt,
		notification.UpdatedAt,
		notification.SentAt,
		notification.ReadAt,
		notification.TemplateID,
		templateData,
	)
	
	return err
}

// GetNotificationByID retrieves a notification by ID
func (db *DB) GetNotificationByID(id uuid.UUID) (*model.Notification, error) {
	query := `
		SELECT 
			id, user_id, type, title, content, status, event_type, event_id, 
			created_at, updated_at, sent_at, read_at, template_id, template_data
		FROM notifications
		WHERE id = $1
	`
	
	var notification model.Notification
	var templateData []byte
	
	err := db.QueryRow(query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Content,
		&notification.Status,
		&notification.EventType,
		&notification.EventID,
		&notification.CreatedAt,
		&notification.UpdatedAt,
		&notification.SentAt,
		&notification.ReadAt,
		&notification.TemplateID,
		&templateData,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Notification not found
		}
		return nil, err
	}
	
	if len(templateData) > 0 {
		if err := json.Unmarshal(templateData, &notification.TemplateData); err != nil {
			return nil, err
		}
	}
	
	return &notification, nil
}

// GetNotificationsByUserID retrieves notifications for a user
func (db *DB) GetNotificationsByUserID(userID uuid.UUID, limit, offset int) ([]*model.Notification, error) {
	query := `
		SELECT 
			id, user_id, type, title, content, status, event_type, event_id, 
			created_at, updated_at, sent_at, read_at, template_id, template_data
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var notifications []*model.Notification
	for rows.Next() {
		var notification model.Notification
		var templateData []byte
		
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Content,
			&notification.Status,
			&notification.EventType,
			&notification.EventID,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&notification.SentAt,
			&notification.ReadAt,
			&notification.TemplateID,
			&templateData,
		)
		
		if err != nil {
			return nil, err
		}
		
		if len(templateData) > 0 {
			if err := json.Unmarshal(templateData, &notification.TemplateData); err != nil {
				return nil, err
			}
		}
		
		notifications = append(notifications, &notification)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return notifications, nil
}

// GetUnreadNotificationsByUserID retrieves unread notifications for a user
func (db *DB) GetUnreadNotificationsByUserID(userID uuid.UUID, limit, offset int) ([]*model.Notification, error) {
	query := `
		SELECT 
			id, user_id, type, title, content, status, event_type, event_id, 
			created_at, updated_at, sent_at, read_at, template_id, template_data
		FROM notifications
		WHERE user_id = $1 AND read_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var notifications []*model.Notification
	for rows.Next() {
		var notification model.Notification
		var templateData []byte
		
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Content,
			&notification.Status,
			&notification.EventType,
			&notification.EventID,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&notification.SentAt,
			&notification.ReadAt,
			&notification.TemplateID,
			&templateData,
		)
		
		if err != nil {
			return nil, err
		}
		
		if len(templateData) > 0 {
			if err := json.Unmarshal(templateData, &notification.TemplateData); err != nil {
				return nil, err
			}
		}
		
		notifications = append(notifications, &notification)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return notifications, nil
}

// UpdateNotificationStatus updates a notification's status
func (db *DB) UpdateNotificationStatus(id uuid.UUID, status model.NotificationStatus) error {
	query := `
		UPDATE notifications
		SET status = $1, updated_at = $2, sent_at = CASE WHEN $1 = 'sent' THEN $2 ELSE sent_at END
		WHERE id = $3
	`
	
	_, err := db.Exec(query, status, time.Now().UTC(), id)
	return err
}

// MarkNotificationAsRead marks a notification as read
func (db *DB) MarkNotificationAsRead(id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET read_at = $1, updated_at = $1
		WHERE id = $2 AND read_at IS NULL
	`
	
	_, err := db.Exec(query, time.Now().UTC(), id)
	return err
}

// DeleteNotification deletes a notification
func (db *DB) DeleteNotification(id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// CreateTemplate creates a new notification template
func (db *DB) CreateTemplate(template *model.NotificationTemplate) error {
	query := `
		INSERT INTO notification_templates (
			id, name, description, event_type, type, subject, content, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	_, err := db.Exec(
		query,
		template.ID,
		template.Name,
		template.Description,
		template.EventType,
		template.Type,
		template.Subject,
		template.Content,
		template.CreatedAt,
		template.UpdatedAt,
	)
	
	return err
}

// GetTemplateByID retrieves a template by ID
func (db *DB) GetTemplateByID(id string) (*model.NotificationTemplate, error) {
	query := `
		SELECT 
			id, name, description, event_type, type, subject, content, created_at, updated_at
		FROM notification_templates
		WHERE id = $1
	`
	
	var template model.NotificationTemplate
	err := db.QueryRow(query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Description,
		&template.EventType,
		&template.Type,
		&template.Subject,
		&template.Content,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Template not found
		}
		return nil, err
	}
	
	return &template, nil
}

// GetTemplatesByEventType retrieves templates by event type
func (db *DB) GetTemplatesByEventType(eventType model.EventType) ([]*model.NotificationTemplate, error) {
	query := `
		SELECT 
			id, name, description, event_type, type, subject, content, created_at, updated_at
		FROM notification_templates
		WHERE event_type = $1
	`
	
	rows, err := db.Query(query, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var templates []*model.NotificationTemplate
	for rows.Next() {
		var template model.NotificationTemplate
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Description,
			&template.EventType,
			&template.Type,
			&template.Subject,
			&template.Content,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		templates = append(templates, &template)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return templates, nil
}

// UpdateTemplate updates a notification template
func (db *DB) UpdateTemplate(template *model.NotificationTemplate) error {
	query := `
		UPDATE notification_templates
		SET 
			name = $1,
			description = $2,
			event_type = $3,
			type = $4,
			subject = $5,
			content = $6,
			updated_at = $7
		WHERE id = $8
	`
	
	_, err := db.Exec(
		query,
		template.Name,
		template.Description,
		template.EventType,
		template.Type,
		template.Subject,
		template.Content,
		time.Now().UTC(),
		template.ID,
	)
	
	return err
}

// DeleteTemplate deletes a notification template
func (db *DB) DeleteTemplate(id string) error {
	query := `DELETE FROM notification_templates WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// CreatePreference creates a new notification preference
func (db *DB) CreatePreference(preference *model.NotificationPreference) error {
	query := `
		INSERT INTO notification_preferences (
			id, user_id, event_type, channels, enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	channels, err := json.Marshal(preference.Channels)
	if err != nil {
		return err
	}
	
	_, err = db.Exec(
		query,
		preference.ID,
		preference.UserID,
		preference.EventType,
		channels,
		preference.Enabled,
		preference.CreatedAt,
		preference.UpdatedAt,
	)
	
	return err
}

// GetPreferenceByUserIDAndEventType retrieves a preference by user ID and event type
func (db *DB) GetPreferenceByUserIDAndEventType(userID uuid.UUID, eventType model.EventType) (*model.NotificationPreference, error) {
	query := `
		SELECT 
			id, user_id, event_type, channels, enabled, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1 AND event_type = $2
	`
	
	var preference model.NotificationPreference
	var channels []byte
	
	err := db.QueryRow(query, userID, eventType).Scan(
		&preference.ID,
		&preference.UserID,
		&preference.EventType,
		&channels,
		&preference.Enabled,
		&preference.CreatedAt,
		&preference.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Preference not found
		}
		return nil, err
	}
	
	if err := json.Unmarshal(channels, &preference.Channels); err != nil {
		return nil, err
	}
	
	return &preference, nil
}

// GetPreferencesByUserID retrieves preferences for a user
func (db *DB) GetPreferencesByUserID(userID uuid.UUID) ([]*model.NotificationPreference, error) {
	query := `
		SELECT 
			id, user_id, event_type, channels, enabled, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`
	
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var preferences []*model.NotificationPreference
	for rows.Next() {
		var preference model.NotificationPreference
		var channels []byte
		
		err := rows.Scan(
			&preference.ID,
			&preference.UserID,
			&preference.EventType,
			&channels,
			&preference.Enabled,
			&preference.CreatedAt,
			&preference.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if err := json.Unmarshal(channels, &preference.Channels); err != nil {
			return nil, err
		}
		
		preferences = append(preferences, &preference)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return preferences, nil
}

// UpdatePreference updates a notification preference
func (db *DB) UpdatePreference(preference *model.NotificationPreference) error {
	query := `
		UPDATE notification_preferences
		SET 
			channels = $1,
			enabled = $2,
			updated_at = $3
		WHERE id = $4
	`
	
	channels, err := json.Marshal(preference.Channels)
	if err != nil {
		return err
	}
	
	_, err = db.Exec(
		query,
		channels,
		preference.Enabled,
		time.Now().UTC(),
		preference.ID,
	)
	
	return err
}

// DeletePreference deletes a notification preference
func (db *DB) DeletePreference(id uuid.UUID) error {
	query := `DELETE FROM notification_preferences WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}
