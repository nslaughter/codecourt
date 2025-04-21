// Package main implements the API Gateway service for CodeCourt
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nslaughter/codecourt/pkg/metrics"
)

// Version information (would be set during build)
var (
	version    = "0.1.0"
	buildDate  = "2025-04-21"
	commitHash = "development"
)

// Service name
const serviceName = "api-gateway"

func main() {
	// Parse command line flags
	var (
		port = flag.Int("port", 8080, "HTTP server port")
	)
	flag.Parse()

	// Create a new router
	mux := http.NewServeMux()

	// Register service info metrics
	metrics.RegisterServiceInfo(serviceName, version, buildDate, commitHash)

	// Register API routes
	mux.HandleFunc("/api/v1/health", healthCheckHandler)
	mux.HandleFunc("/api/v1/users", forwardToUserService)
	mux.HandleFunc("/api/v1/problems", forwardToProblemService)
	mux.HandleFunc("/api/v1/submissions", forwardToSubmissionService)

	// Set up metrics endpoint
	metrics.SetupMetricsEndpoint(mux)

	// Apply metrics middleware
	handler := metrics.MetricsMiddleware(serviceName)(mux)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting %s server on port %d", serviceName, *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// healthCheckHandler handles health check requests
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","service":"%s","version":"%s"}`, serviceName, version)
}

// forwardToUserService forwards requests to the User Service
func forwardToUserService(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would use a reverse proxy
	// For demonstration purposes, we'll just record metrics and return a placeholder
	
	// Simulate processing
	time.Sleep(10 * time.Millisecond)
	
	// Record database operation metrics (example)
	if r.Method == http.MethodGet {
		metrics.RecordDatabaseOperation(serviceName, "SELECT", "users", "success")
		metrics.ObserveDatabaseOperationDuration(serviceName, "SELECT", "users", 0.005)
	}
	
	// Return a response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Request forwarded to User Service"}`)
}

// forwardToProblemService forwards requests to the Problem Service
func forwardToProblemService(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would use a reverse proxy
	// For demonstration purposes, we'll just record metrics and return a placeholder
	
	// Simulate processing
	time.Sleep(15 * time.Millisecond)
	
	// Record problem access metrics (example)
	if r.Method == http.MethodGet {
		metrics.RecordProblemAccess("example-problem-id", "medium")
	}
	
	// Return a response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Request forwarded to Problem Service"}`)
}

// forwardToSubmissionService forwards requests to the Submission Service
func forwardToSubmissionService(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would use a reverse proxy
	// For demonstration purposes, we'll just record metrics and return a placeholder
	
	// Simulate processing
	time.Sleep(20 * time.Millisecond)
	
	// Record submission metrics (example)
	if r.Method == http.MethodPost {
		metrics.RecordSubmission("go", "pending", "example-problem-id", "example-user-id")
		
		// Record Kafka message metrics
		metrics.RecordKafkaMessage(serviceName, "submission-events", "produce")
		metrics.ObserveKafkaOperationDuration(serviceName, "submission-events", "produce", 0.01)
	}
	
	// Return a response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Request forwarded to Submission Service"}`)
}
