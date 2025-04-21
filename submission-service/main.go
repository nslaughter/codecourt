package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/nslaughter/codecourt/submission-service/api"
	"github.com/nslaughter/codecourt/submission-service/config"
	"github.com/nslaughter/codecourt/submission-service/db"
	"github.com/nslaughter/codecourt/submission-service/kafka"
	"github.com/nslaughter/codecourt/submission-service/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create Kafka producer
	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Create Kafka consumer
	consumer, err := kafka.NewConsumer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Create submission service
	submissionService := service.NewSubmissionService(cfg, database, producer, consumer)

	// Create API handler
	handler := api.NewHandler(submissionService)

	// Create router
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start processing judging results
	go submissionService.ProcessJudgingResults(ctx)

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on port %d", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down...", sig)

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Cancel context to stop processing judging results
	cancel()

	log.Println("Shutdown complete")
}
