package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nslaughter/codecourt/user-service/model"
	"github.com/nslaughter/codecourt/user-service/service"
)

// Handler represents the API handler
type Handler struct {
	service service.UserService
}

// NewHandler creates a new handler
func NewHandler(service service.UserService) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the API routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Authentication routes
	router.HandleFunc("/api/v1/auth/register", h.Register).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", h.Login).Methods("POST")
	router.HandleFunc("/api/v1/auth/refresh", h.RefreshToken).Methods("POST")
	router.HandleFunc("/api/v1/auth/logout", h.Logout).Methods("POST")
	
	// User routes
	router.HandleFunc("/api/v1/users", h.ListUsers).Methods("GET")
	router.HandleFunc("/api/v1/users/{id}", h.GetUser).Methods("GET")
	router.HandleFunc("/api/v1/users/{id}", h.UpdateUser).Methods("PUT")
	router.HandleFunc("/api/v1/users/{id}", h.DeleteUser).Methods("DELETE")
	router.HandleFunc("/api/v1/users/{id}/password", h.ChangePassword).Methods("PUT")
	router.HandleFunc("/api/v1/users/me", h.GetCurrentUser).Methods("GET")
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	user, err := h.service.Register(&req)
	if err != nil {
		if errors.Is(err, service.ErrUsernameExists) || errors.Is(err, service.ErrEmailExists) {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error registering user")
		return
	}
	
	respondWithJSON(w, http.StatusCreated, user)
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	tokens, err := h.service.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error logging in")
		return
	}
	
	respondWithJSON(w, http.StatusOK, tokens)
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	tokens, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error refreshing token")
		return
	}
	
	respondWithJSON(w, http.StatusOK, tokens)
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	if err := h.service.Logout(req.RefreshToken); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error logging out")
		return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// GetUser retrieves a user by ID
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	user, err := h.service.GetUserByID(id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving user")
		return
	}
	
	respondWithJSON(w, http.StatusOK, user)
}

// UpdateUser updates a user
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	var req model.UserUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	user, err := h.service.UpdateUser(id, &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		if errors.Is(err, service.ErrEmailExists) {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error updating user")
		return
	}
	
	respondWithJSON(w, http.StatusOK, user)
}

// DeleteUser deletes a user
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	if err := h.service.DeleteUser(id); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error deleting user")
		return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// ChangePassword changes a user's password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	var req model.PasswordChange
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	
	if err := h.service.ChangePassword(id, &req); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondWithError(w, http.StatusUnauthorized, "Invalid current password")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error changing password")
		return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

// ListUsers retrieves all users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving users")
		return
	}
	
	respondWithJSON(w, http.StatusOK, users)
}

// GetCurrentUser retrieves the current user based on the JWT token
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Authorization header required")
		return
	}
	
	// Check if the header has the correct format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}
	
	// Validate the token
	claims, err := h.service.ValidateToken(parts[1])
	if err != nil {
		if errors.Is(err, service.ErrExpiredToken) {
			respondWithError(w, http.StatusUnauthorized, "Token expired")
			return
		}
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}
	
	// Get the user
	user, err := h.service.GetUserByID(claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving user")
		return
	}
	
	respondWithJSON(w, http.StatusOK, user)
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
