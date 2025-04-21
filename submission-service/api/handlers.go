package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nslaughter/codecourt/submission-service/model"
	"github.com/nslaughter/codecourt/submission-service/service"
)

// Handler represents the API handler
type Handler struct {
	service service.SubmissionServiceInterface
}

// NewHandler creates a new API handler
func NewHandler(service service.SubmissionServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the API routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/submissions", h.CreateSubmission).Methods("POST")
	router.HandleFunc("/api/v1/submissions/{id}", h.GetSubmission).Methods("GET")
	router.HandleFunc("/api/v1/submissions/{id}/result", h.GetSubmissionResult).Methods("GET")
	router.HandleFunc("/api/v1/users/{user_id}/submissions", h.GetSubmissionsByUserID).Methods("GET")
	router.HandleFunc("/api/v1/problems/{problem_id}/submissions", h.GetSubmissionsByProblemID).Methods("GET")
}

// CreateSubmission handles the creation of a new submission
func (h *Handler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req model.SubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.ProblemID == "" || req.UserID == "" || req.Code == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create submission
	submission := model.NewSubmission(req.ProblemID, req.UserID, req.Language, req.Code)

	// Save submission
	if err := h.service.CreateSubmission(submission); err != nil {
		log.Printf("Error creating submission: %v", err)
		http.Error(w, "Failed to create submission", http.StatusInternalServerError)
		return
	}

	// Create response
	resp := model.SubmissionResponse{
		ID:        submission.ID,
		ProblemID: submission.ProblemID,
		UserID:    submission.UserID,
		Language:  submission.Language,
		Status:    submission.Status,
		CreatedAt: submission.CreatedAt,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetSubmission handles retrieving a submission by ID
func (h *Handler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	// Get submission ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing submission ID", http.StatusBadRequest)
		return
	}

	// Get submission
	submission, err := h.service.GetSubmission(id)
	if err != nil {
		log.Printf("Error getting submission: %v", err)
		http.Error(w, "Failed to get submission", http.StatusNotFound)
		return
	}

	// Create response
	resp := model.SubmissionResponse{
		ID:        submission.ID,
		ProblemID: submission.ProblemID,
		UserID:    submission.UserID,
		Language:  submission.Language,
		Status:    submission.Status,
		CreatedAt: submission.CreatedAt,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetSubmissionResult handles retrieving a submission result by submission ID
func (h *Handler) GetSubmissionResult(w http.ResponseWriter, r *http.Request) {
	// Get submission ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing submission ID", http.StatusBadRequest)
		return
	}

	// Get submission result
	result, err := h.service.GetSubmissionResult(id)
	if err != nil {
		log.Printf("Error getting submission result: %v", err)
		http.Error(w, "Failed to get submission result", http.StatusNotFound)
		return
	}

	// Create response
	resp := model.SubmissionResultResponse{
		ID:              result.ID,
		SubmissionID:    result.SubmissionID,
		Status:          result.Status,
		ExecutionTime:   result.ExecutionTime,
		MemoryUsage:     result.MemoryUsage,
		ErrorMessage:    result.ErrorMessage,
		TestCaseResults: result.TestCaseResults,
		CreatedAt:       result.CreatedAt,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetSubmissionsByUserID handles retrieving all submissions for a user
func (h *Handler) GetSubmissionsByUserID(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["user_id"]
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Get submissions
	submissions, err := h.service.GetSubmissionsByUserID(userID)
	if err != nil {
		log.Printf("Error getting submissions: %v", err)
		http.Error(w, "Failed to get submissions", http.StatusInternalServerError)
		return
	}

	// Create response
	var resp []model.SubmissionResponse
	for _, submission := range submissions {
		resp = append(resp, model.SubmissionResponse{
			ID:        submission.ID,
			ProblemID: submission.ProblemID,
			UserID:    submission.UserID,
			Language:  submission.Language,
			Status:    submission.Status,
			CreatedAt: submission.CreatedAt,
		})
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetSubmissionsByProblemID handles retrieving all submissions for a problem
func (h *Handler) GetSubmissionsByProblemID(w http.ResponseWriter, r *http.Request) {
	// Get problem ID from URL
	vars := mux.Vars(r)
	problemID := vars["problem_id"]
	if problemID == "" {
		http.Error(w, "Missing problem ID", http.StatusBadRequest)
		return
	}

	// Get submissions
	submissions, err := h.service.GetSubmissionsByProblemID(problemID)
	if err != nil {
		log.Printf("Error getting submissions: %v", err)
		http.Error(w, "Failed to get submissions", http.StatusInternalServerError)
		return
	}

	// Create response
	var resp []model.SubmissionResponse
	for _, submission := range submissions {
		resp = append(resp, model.SubmissionResponse{
			ID:        submission.ID,
			ProblemID: submission.ProblemID,
			UserID:    submission.UserID,
			Language:  submission.Language,
			Status:    submission.Status,
			CreatedAt: submission.CreatedAt,
		})
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
