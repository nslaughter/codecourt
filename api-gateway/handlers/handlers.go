package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nslaughter/codecourt/api-gateway/config"
	"github.com/nslaughter/codecourt/api-gateway/proxy"
)

// Handler represents the API Gateway handler
type Handler struct {
	cfg   *config.Config
	proxy *proxy.ServiceProxy
}

// NewHandler creates a new handler
func NewHandler(cfg *config.Config, proxy *proxy.ServiceProxy) *Handler {
	return &Handler{
		cfg:   cfg,
		proxy: proxy,
	}
}

// RegisterRoutes registers the API routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Health check endpoint
	router.HandleFunc("/api/v1/health", h.HealthCheck).Methods("GET")

	// Create a subrouter for API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Register routes for each service
	h.registerProblemRoutes(apiRouter)
	h.registerSubmissionRoutes(apiRouter)
	h.registerJudgingRoutes(apiRouter)
	h.registerAuthRoutes(apiRouter)

	// Catch-all route for proxying requests
	router.PathPrefix("/api/v1/").HandlerFunc(h.proxy.ProxyRequest)
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// registerProblemRoutes registers routes for the Problem Service
func (h *Handler) registerProblemRoutes(router *mux.Router) {
	// Problems
	router.HandleFunc("/problems", h.proxy.ProxyRequest).Methods("GET", "POST")
	router.HandleFunc("/problems/{id}", h.proxy.ProxyRequest).Methods("GET", "PUT", "DELETE")
	
	// Test cases
	router.HandleFunc("/problems/{id}/testcases", h.proxy.ProxyRequest).Methods("GET", "POST")
	router.HandleFunc("/testcases/{id}", h.proxy.ProxyRequest).Methods("GET", "PUT", "DELETE")
	
	// Categories
	router.HandleFunc("/categories", h.proxy.ProxyRequest).Methods("GET", "POST")
	router.HandleFunc("/categories/{id}", h.proxy.ProxyRequest).Methods("GET", "PUT", "DELETE")
	
	// Templates
	router.HandleFunc("/problems/{id}/templates", h.proxy.ProxyRequest).Methods("GET", "POST")
	router.HandleFunc("/templates/{id}", h.proxy.ProxyRequest).Methods("GET", "PUT", "DELETE")
}

// registerSubmissionRoutes registers routes for the Submission Service
func (h *Handler) registerSubmissionRoutes(router *mux.Router) {
	// Submissions
	router.HandleFunc("/submissions", h.proxy.ProxyRequest).Methods("GET", "POST")
	router.HandleFunc("/submissions/{id}", h.proxy.ProxyRequest).Methods("GET")
	router.HandleFunc("/users/{id}/submissions", h.proxy.ProxyRequest).Methods("GET")
	router.HandleFunc("/problems/{id}/submissions", h.proxy.ProxyRequest).Methods("GET")
}

// registerJudgingRoutes registers routes for the Judging Service
func (h *Handler) registerJudgingRoutes(router *mux.Router) {
	// Judging results
	router.HandleFunc("/judging/results", h.proxy.ProxyRequest).Methods("GET")
	router.HandleFunc("/judging/results/{id}", h.proxy.ProxyRequest).Methods("GET")
	router.HandleFunc("/judging/status/{id}", h.proxy.ProxyRequest).Methods("GET")
}

// registerAuthRoutes registers routes for the Auth Service
func (h *Handler) registerAuthRoutes(router *mux.Router) {
	// Authentication
	router.HandleFunc("/auth/login", h.proxy.ProxyRequest).Methods("POST")
	router.HandleFunc("/auth/register", h.proxy.ProxyRequest).Methods("POST")
	router.HandleFunc("/auth/refresh", h.proxy.ProxyRequest).Methods("POST")
	router.HandleFunc("/auth/logout", h.proxy.ProxyRequest).Methods("POST")
	
	// User management
	router.HandleFunc("/users", h.proxy.ProxyRequest).Methods("GET")
	router.HandleFunc("/users/{id}", h.proxy.ProxyRequest).Methods("GET", "PUT", "DELETE")
	router.HandleFunc("/users/me", h.proxy.ProxyRequest).Methods("GET")
}
