package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nslaughter/codecourt/problem-service/model"
	"github.com/nslaughter/codecourt/problem-service/service"
)

// Handler represents the API handler
type Handler struct {
	service service.ProblemServiceInterface
}

// NewHandler creates a new API handler
func NewHandler(service service.ProblemServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the API routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Problem routes
	router.HandleFunc("/api/v1/problems", h.CreateProblem).Methods("POST")
	router.HandleFunc("/api/v1/problems", h.ListProblems).Methods("GET")
	router.HandleFunc("/api/v1/problems/{id}", h.GetProblem).Methods("GET")
	router.HandleFunc("/api/v1/problems/{id}", h.UpdateProblem).Methods("PUT")
	router.HandleFunc("/api/v1/problems/{id}", h.DeleteProblem).Methods("DELETE")

	// Test case routes
	router.HandleFunc("/api/v1/problems/{problem_id}/test-cases", h.CreateTestCase).Methods("POST")
	router.HandleFunc("/api/v1/problems/{problem_id}/test-cases", h.ListTestCases).Methods("GET")
	router.HandleFunc("/api/v1/test-cases/{id}", h.GetTestCase).Methods("GET")
	router.HandleFunc("/api/v1/test-cases/{id}", h.UpdateTestCase).Methods("PUT")
	router.HandleFunc("/api/v1/test-cases/{id}", h.DeleteTestCase).Methods("DELETE")

	// Category routes
	router.HandleFunc("/api/v1/categories", h.CreateCategory).Methods("POST")
	router.HandleFunc("/api/v1/categories", h.ListCategories).Methods("GET")
	router.HandleFunc("/api/v1/categories/{id}", h.GetCategory).Methods("GET")
	router.HandleFunc("/api/v1/categories/{id}", h.UpdateCategory).Methods("PUT")
	router.HandleFunc("/api/v1/categories/{id}", h.DeleteCategory).Methods("DELETE")
	router.HandleFunc("/api/v1/categories/{id}/problems", h.ListProblemsByCategory).Methods("GET")

	// Problem template routes
	router.HandleFunc("/api/v1/problems/{problem_id}/templates", h.CreateProblemTemplate).Methods("POST")
	router.HandleFunc("/api/v1/problems/{problem_id}/templates", h.ListProblemTemplates).Methods("GET")
	router.HandleFunc("/api/v1/problems/{problem_id}/templates/{language}", h.GetProblemTemplateByLanguage).Methods("GET")
	router.HandleFunc("/api/v1/templates/{id}", h.GetProblemTemplate).Methods("GET")
	router.HandleFunc("/api/v1/templates/{id}", h.UpdateProblemTemplate).Methods("PUT")
	router.HandleFunc("/api/v1/templates/{id}", h.DeleteProblemTemplate).Methods("DELETE")
}

// CreateProblem handles the creation of a new problem
func (h *Handler) CreateProblem(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req model.ProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Title == "" || req.Description == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create problem
	problem, err := h.service.CreateProblem(&req)
	if err != nil {
		log.Printf("Error creating problem: %v", err)
		http.Error(w, "Failed to create problem", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(problem)
}

// GetProblem handles retrieving a problem by ID
func (h *Handler) GetProblem(w http.ResponseWriter, r *http.Request) {
	// Get problem ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing problem ID", http.StatusBadRequest)
		return
	}

	// Get problem
	problem, err := h.service.GetProblem(id)
	if err != nil {
		log.Printf("Error getting problem: %v", err)
		http.Error(w, "Failed to get problem", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problem)
}

// UpdateProblem handles updating a problem
func (h *Handler) UpdateProblem(w http.ResponseWriter, r *http.Request) {
	// Get problem ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing problem ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req model.ProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Title == "" || req.Description == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Update problem
	problem, err := h.service.UpdateProblem(id, &req)
	if err != nil {
		log.Printf("Error updating problem: %v", err)
		http.Error(w, "Failed to update problem", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(problem)
}

// DeleteProblem handles deleting a problem
func (h *Handler) DeleteProblem(w http.ResponseWriter, r *http.Request) {
	// Get problem ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing problem ID", http.StatusBadRequest)
		return
	}

	// Delete problem
	if err := h.service.DeleteProblem(id); err != nil {
		log.Printf("Error deleting problem: %v", err)
		http.Error(w, "Failed to delete problem", http.StatusInternalServerError)
		return
	}

	// Return response
	w.WriteHeader(http.StatusNoContent)
}

// ListProblems handles listing all problems with pagination
func (h *Handler) ListProblems(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	offset, limit := getPaginationParams(r)

	// List problems
	problems, err := h.service.ListProblems(offset, limit)
	if err != nil {
		log.Printf("Error listing problems: %v", err)
		http.Error(w, "Failed to list problems", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"problems": problems,
	})
}

// CreateTestCase handles the creation of a new test case
func (h *Handler) CreateTestCase(w http.ResponseWriter, r *http.Request) {
	// Get problem ID from URL
	vars := mux.Vars(r)
	problemID := vars["problem_id"]
	if problemID == "" {
		http.Error(w, "Missing problem ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req model.TestCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Input == "" || req.Output == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create test case
	testCase, err := h.service.CreateTestCase(problemID, &req)
	if err != nil {
		log.Printf("Error creating test case: %v", err)
		http.Error(w, "Failed to create test case", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(testCase)
}

// GetTestCase handles retrieving a test case by ID
func (h *Handler) GetTestCase(w http.ResponseWriter, r *http.Request) {
	// Get test case ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing test case ID", http.StatusBadRequest)
		return
	}

	// Get test case
	testCase, err := h.service.GetTestCase(id)
	if err != nil {
		log.Printf("Error getting test case: %v", err)
		http.Error(w, "Failed to get test case", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testCase)
}

// UpdateTestCase handles updating a test case
func (h *Handler) UpdateTestCase(w http.ResponseWriter, r *http.Request) {
	// Get test case ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing test case ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req model.TestCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Input == "" || req.Output == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Update test case
	testCase, err := h.service.UpdateTestCase(id, &req)
	if err != nil {
		log.Printf("Error updating test case: %v", err)
		http.Error(w, "Failed to update test case", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testCase)
}

// DeleteTestCase handles deleting a test case
func (h *Handler) DeleteTestCase(w http.ResponseWriter, r *http.Request) {
	// Get test case ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing test case ID", http.StatusBadRequest)
		return
	}

	// Delete test case
	if err := h.service.DeleteTestCase(id); err != nil {
		log.Printf("Error deleting test case: %v", err)
		http.Error(w, "Failed to delete test case", http.StatusInternalServerError)
		return
	}

	// Return response
	w.WriteHeader(http.StatusNoContent)
}

// ListTestCases handles listing all test cases for a problem
func (h *Handler) ListTestCases(w http.ResponseWriter, r *http.Request) {
	// Get problem ID from URL
	vars := mux.Vars(r)
	problemID := vars["problem_id"]
	if problemID == "" {
		http.Error(w, "Missing problem ID", http.StatusBadRequest)
		return
	}

	// Check if hidden test cases should be included
	includeHidden := r.URL.Query().Get("include_hidden") == "true"

	// List test cases
	testCases, err := h.service.ListTestCases(problemID, includeHidden)
	if err != nil {
		log.Printf("Error listing test cases: %v", err)
		http.Error(w, "Failed to list test cases", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"test_cases": testCases,
	})
}

// CreateCategory handles the creation of a new category
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req model.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create category
	category, err := h.service.CreateCategory(&req)
	if err != nil {
		log.Printf("Error creating category: %v", err)
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// GetCategory handles retrieving a category by ID
func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	// Get category ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing category ID", http.StatusBadRequest)
		return
	}

	// Get category
	category, err := h.service.GetCategory(id)
	if err != nil {
		log.Printf("Error getting category: %v", err)
		http.Error(w, "Failed to get category", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// UpdateCategory handles updating a category
func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	// Get category ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing category ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req model.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Update category
	category, err := h.service.UpdateCategory(id, &req)
	if err != nil {
		log.Printf("Error updating category: %v", err)
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// DeleteCategory handles deleting a category
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	// Get category ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing category ID", http.StatusBadRequest)
		return
	}

	// Delete category
	if err := h.service.DeleteCategory(id); err != nil {
		log.Printf("Error deleting category: %v", err)
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}

	// Return response
	w.WriteHeader(http.StatusNoContent)
}

// ListCategories handles listing all categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	// List categories
	categories, err := h.service.ListCategories()
	if err != nil {
		log.Printf("Error listing categories: %v", err)
		http.Error(w, "Failed to list categories", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
	})
}

// ListProblemsByCategory handles listing all problems in a category
func (h *Handler) ListProblemsByCategory(w http.ResponseWriter, r *http.Request) {
	// Get category ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing category ID", http.StatusBadRequest)
		return
	}

	// Get pagination parameters
	offset, limit := getPaginationParams(r)

	// List problems
	problems, err := h.service.ListProblemsByCategory(id, offset, limit)
	if err != nil {
		log.Printf("Error listing problems by category: %v", err)
		http.Error(w, "Failed to list problems", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"problems": problems,
	})
}

// CreateProblemTemplate handles the creation of a new problem template
func (h *Handler) CreateProblemTemplate(w http.ResponseWriter, r *http.Request) {
	// Get problem ID from URL
	vars := mux.Vars(r)
	problemID := vars["problem_id"]
	if problemID == "" {
		http.Error(w, "Missing problem ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req model.ProblemTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Template == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create template
	template, err := h.service.CreateProblemTemplate(problemID, &req)
	if err != nil {
		log.Printf("Error creating problem template: %v", err)
		http.Error(w, "Failed to create problem template", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

// GetProblemTemplate handles retrieving a problem template by ID
func (h *Handler) GetProblemTemplate(w http.ResponseWriter, r *http.Request) {
	// Get template ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing template ID", http.StatusBadRequest)
		return
	}

	// Get template
	template, err := h.service.GetProblemTemplate(id)
	if err != nil {
		log.Printf("Error getting problem template: %v", err)
		http.Error(w, "Failed to get problem template", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// GetProblemTemplateByLanguage handles retrieving a problem template by language
func (h *Handler) GetProblemTemplateByLanguage(w http.ResponseWriter, r *http.Request) {
	// Get problem ID and language from URL
	vars := mux.Vars(r)
	problemID := vars["problem_id"]
	language := vars["language"]
	if problemID == "" || language == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Get template
	template, err := h.service.GetProblemTemplateByLanguage(problemID, model.Language(language))
	if err != nil {
		log.Printf("Error getting problem template by language: %v", err)
		http.Error(w, "Failed to get problem template", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// UpdateProblemTemplate handles updating a problem template
func (h *Handler) UpdateProblemTemplate(w http.ResponseWriter, r *http.Request) {
	// Get template ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing template ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req model.ProblemTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Template == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Update template
	template, err := h.service.UpdateProblemTemplate(id, &req)
	if err != nil {
		log.Printf("Error updating problem template: %v", err)
		http.Error(w, "Failed to update problem template", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// DeleteProblemTemplate handles deleting a problem template
func (h *Handler) DeleteProblemTemplate(w http.ResponseWriter, r *http.Request) {
	// Get template ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing template ID", http.StatusBadRequest)
		return
	}

	// Delete template
	if err := h.service.DeleteProblemTemplate(id); err != nil {
		log.Printf("Error deleting problem template: %v", err)
		http.Error(w, "Failed to delete problem template", http.StatusInternalServerError)
		return
	}

	// Return response
	w.WriteHeader(http.StatusNoContent)
}

// ListProblemTemplates handles listing all templates for a problem
func (h *Handler) ListProblemTemplates(w http.ResponseWriter, r *http.Request) {
	// Get problem ID from URL
	vars := mux.Vars(r)
	problemID := vars["problem_id"]
	if problemID == "" {
		http.Error(w, "Missing problem ID", http.StatusBadRequest)
		return
	}

	// List templates
	templates, err := h.service.ListProblemTemplates(problemID)
	if err != nil {
		log.Printf("Error listing problem templates: %v", err)
		http.Error(w, "Failed to list problem templates", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
	})
}

// getPaginationParams gets pagination parameters from the request
func getPaginationParams(r *http.Request) (int, int) {
	// Get offset parameter
	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		offsetInt, err := strconv.Atoi(offsetStr)
		if err == nil && offsetInt >= 0 {
			offset = offsetInt
		}
	}

	// Get limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err == nil && limitInt > 0 {
			limit = limitInt
		}
	}

	return offset, limit
}
