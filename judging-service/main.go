package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nslaughter/codecourt/judging-service/config"
	kafkalib "github.com/nslaughter/codecourt/judging-service/kafka"
	"github.com/nslaughter/codecourt/judging-service/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create judging service
	judgingService, err := service.NewJudgingService(cfg)
	if err != nil {
		log.Fatalf("Failed to create judging service: %v", err)
	}
	defer judgingService.Close()

	// Create Kafka consumer
	consumer, err := kafkalib.NewConsumer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Create Kafka producer
	producer, err := kafkalib.NewProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Create context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start processing submissions
	go judgingService.ProcessSubmissions(ctx, consumer, producer)

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down...", sig)
}
