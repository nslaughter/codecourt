package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nslaughter/codecourt/notification-service/config"
	"github.com/nslaughter/codecourt/notification-service/model"
	"github.com/nslaughter/codecourt/notification-service/service"
	"github.com/segmentio/kafka-go"
)

// Consumer represents a Kafka consumer
type Consumer struct {
	readers         []*kafka.Reader
	notificationSvc service.NotificationService
	cfg             *config.Config
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(notificationSvc service.NotificationService, cfg *config.Config) *Consumer {
	return &Consumer{
		notificationSvc: notificationSvc,
		cfg:             cfg,
	}
}

// Start starts consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) error {
	// Create readers for each topic
	for _, topic := range c.cfg.KafkaTopics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        c.cfg.KafkaBrokers,
			Topic:          topic,
			GroupID:        c.cfg.KafkaGroupID,
			MinBytes:       10e3,    // 10KB
			MaxBytes:       10e6,    // 10MB
			MaxWait:        1 * time.Second,
			StartOffset:    kafka.FirstOffset,
			RetentionTime:  7 * 24 * time.Hour, // 1 week
			CommitInterval: 1 * time.Second,
		})

		c.readers = append(c.readers, reader)

		// Start consumer goroutine for this topic
		go c.consume(ctx, reader)
	}

	return nil
}

// Stop stops all Kafka consumers
func (c *Consumer) Stop() {
	for _, reader := range c.readers {
		reader.Close()
	}
}

// consume consumes messages from a Kafka topic
func (c *Consumer) consume(ctx context.Context, reader *kafka.Reader) {
	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Read message
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		// Process message
		if err := c.processMessage(msg); err != nil {
			log.Printf("Error processing message: %v", err)
		}
	}
}

// processMessage processes a Kafka message
func (c *Consumer) processMessage(msg kafka.Message) error {
	// Parse event
	var event model.Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("error unmarshalling event: %w", err)
	}

	// Set event type based on topic if not provided
	if event.Type == "" {
		event.Type = model.EventType(msg.Topic)
	}

	// Set event ID if not provided
	if event.ID == "" {
		event.ID = fmt.Sprintf("%s-%d-%d", msg.Topic, msg.Partition, msg.Offset)
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Handle event
	if err := c.notificationSvc.HandleEvent(&event); err != nil {
		return fmt.Errorf("error handling event: %w", err)
	}

	return nil
}
