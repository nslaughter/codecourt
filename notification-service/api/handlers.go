package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nslaughter/codecourt/notification-service/model"
	"github.com/nslaughter/codecourt/notification-service/service"
)

// Handler represents the API handler
type Handler struct {
	service service.NotificationService
}

// NewHandler creates a new handler
func NewHandler(service service.NotificationService) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the API routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Notification routes
	router.HandleFunc("/api/v1/notifications", h.SendNotification).Methods("POST")
	router.HandleFunc("/api/v1/notifications/batch", h.SendBatchNotifications).Methods("POST")
	router.HandleFunc("/api/v1/notifications/{id}", h.GetNotification).Methods("GET")
	router.HandleFunc("/api/v1/notifications/{id}", h.DeleteNotification).Methods("DELETE")
	router.HandleFunc("/api/v1/notifications/{id}/read", h.MarkNotificationAsRead).Methods("POST")
	router.HandleFunc("/api/v1/users/{user_id}/notifications", h.GetUserNotifications).Methods("GET")
	router.HandleFunc("/api/v1/users/{user_id}/notifications/unread", h.GetUserUnreadNotifications).Methods("GET")
	
	// Template routes
	router.HandleFunc("/api/v1/templates", h.CreateTemplate).Methods("POST")
	router.HandleFunc("/api/v1/templates/{id}", h.GetTemplate).Methods("GET")
	router.HandleFunc("/api/v1/templates/{id}", h.UpdateTemplate).Methods("PUT")
	router.HandleFunc("/api/v1/templates/{id}", h.DeleteTemplate).Methods("DELETE")
	router.HandleFunc("/api/v1/templates/event/{event_type}", h.GetTemplatesByEventType).Methods("GET")
	
	// Preference routes
	router.HandleFunc("/api/v1/users/{user_id}/preferences", h.SetPreference).Methods("POST")
	router.HandleFunc("/api/v1/users/{user_id}/preferences", h.GetUserPreferences).Methods("GET")
}

// SendNotification handles sending a notification
func (h *Handler) SendNotification(w http.ResponseWriter, r *http.Request) {
	var req model.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	notification, err := h.service.SendNotification(&req)
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			respondWithError(w, http.StatusNotFound, "Template not found")
			return
		}
		if errors.Is(err, service.ErrInvalidTemplate) {
			respondWithError(w, http.StatusBadRequest, "Invalid template")
			return
		}
		if errors.Is(err, service.ErrSendingNotification) {
			respondWithError(w, http.StatusInternalServerError, "Error sending notification")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error processing request")
		return
	}
	
	respondWithJSON(w, http.StatusCreated, notification)
}

// SendBatchNotifications handles sending notifications to multiple users
func (h *Handler) SendBatchNotifications(w http.ResponseWriter, r *http.Request) {
	var req model.BatchNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	notificationIDs, err := h.service.SendBatchNotifications(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error sending batch notifications")
		return
	}
	
	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"notification_ids": notificationIDs,
		"count":           len(notificationIDs),
	})
}

// GetNotification handles retrieving a notification
func (h *Handler) GetNotification(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}
	
	notification, err := h.service.GetNotificationByID(id)
	if err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			respondWithError(w, http.StatusNotFound, "Notification not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving notification")
		return
	}
	
	respondWithJSON(w, http.StatusOK, notification)
}

// DeleteNotification handles deleting a notification
func (h *Handler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}
	
	if err := h.service.DeleteNotification(id); err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			respondWithError(w, http.StatusNotFound, "Notification not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error deleting notification")
		return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Notification deleted successfully"})
}

// MarkNotificationAsRead handles marking a notification as read
func (h *Handler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}
	
	if err := h.service.MarkNotificationAsRead(id); err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			respondWithError(w, http.StatusNotFound, "Notification not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error marking notification as read")
		return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Notification marked as read"})
}

// GetUserNotifications handles retrieving notifications for a user
func (h *Handler) GetUserNotifications(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := uuid.Parse(params["user_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	// Get pagination parameters
	limit, offset := getPaginationParams(r)
	
	notifications, err := h.service.GetNotificationsByUserID(userID, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving notifications")
		return
	}
	
	respondWithJSON(w, http.StatusOK, notifications)
}

// GetUserUnreadNotifications handles retrieving unread notifications for a user
func (h *Handler) GetUserUnreadNotifications(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := uuid.Parse(params["user_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	// Get pagination parameters
	limit, offset := getPaginationParams(r)
	
	notifications, err := h.service.GetUnreadNotificationsByUserID(userID, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving unread notifications")
		return
	}
	
	respondWithJSON(w, http.StatusOK, notifications)
}

// CreateTemplate handles creating a notification template
func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var template model.NotificationTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	if err := h.service.CreateTemplate(&template); err != nil {
		if errors.Is(err, service.ErrInvalidTemplate) {
			respondWithError(w, http.StatusBadRequest, "Invalid template")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error creating template")
		return
	}
	
	respondWithJSON(w, http.StatusCreated, template)
}

// GetTemplate handles retrieving a notification template
func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	
	template, err := h.service.GetTemplateByID(id)
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			respondWithError(w, http.StatusNotFound, "Template not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving template")
		return
	}
	
	respondWithJSON(w, http.StatusOK, template)
}

// UpdateTemplate handles updating a notification template
func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	
	var template model.NotificationTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	// Ensure ID in URL matches ID in payload
	template.ID = id
	
	if err := h.service.UpdateTemplate(&template); err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			respondWithError(w, http.StatusNotFound, "Template not found")
			return
		}
		if errors.Is(err, service.ErrInvalidTemplate) {
			respondWithError(w, http.StatusBadRequest, "Invalid template")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error updating template")
		return
	}
	
	respondWithJSON(w, http.StatusOK, template)
}

// DeleteTemplate handles deleting a notification template
func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	
	if err := h.service.DeleteTemplate(id); err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			respondWithError(w, http.StatusNotFound, "Template not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error deleting template")
		return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Template deleted successfully"})
}

// GetTemplatesByEventType handles retrieving templates by event type
func (h *Handler) GetTemplatesByEventType(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	eventType := model.EventType(params["event_type"])
	
	templates, err := h.service.GetTemplatesByEventType(eventType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving templates")
		return
	}
	
	respondWithJSON(w, http.StatusOK, templates)
}

// SetPreference handles setting a notification preference
func (h *Handler) SetPreference(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := uuid.Parse(params["user_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	var req model.NotificationPreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	if err := h.service.SetPreference(userID, &req); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error setting preference")
		return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Preference set successfully"})
}

// GetUserPreferences handles retrieving preferences for a user
func (h *Handler) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := uuid.Parse(params["user_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	preferences, err := h.service.GetPreferencesByUserID(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving preferences")
		return
	}
	
	respondWithJSON(w, http.StatusOK, preferences)
}

// getPaginationParams extracts pagination parameters from the request
func getPaginationParams(r *http.Request) (int, int) {
	// Default values
	limit := 10
	offset := 0
	
	// Get limit from query parameters
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}
	
	// Get offset from query parameters
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}
	
	return limit, offset
}

// respondWithError responds with an error message
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON responds with a JSON payload
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Error marshalling JSON"}`))
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
