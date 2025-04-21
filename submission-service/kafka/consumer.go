package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/nslaughter/codecourt/submission-service/config"
)

// Consumer represents a Kafka consumer
type Consumer struct {
	consumer *kafka.Consumer
	topic    string
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg *config.Config) (*Consumer, error) {
	// Create Kafka consumer configuration
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":  cfg.KafkaBrokers,
		"group.id":           cfg.KafkaGroupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	}

	// Create Kafka consumer
	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	// Subscribe to the topic
	if err := consumer.Subscribe(cfg.KafkaJudgingResultTopic, nil); err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	return &Consumer{
		consumer: consumer,
		topic:    cfg.KafkaJudgingResultTopic,
	}, nil
}

// Consume consumes a message from Kafka with timeout
func (c *Consumer) Consume(timeout time.Duration) (*kafka.Message, error) {
	msg, err := c.consumer.ReadMessage(timeout)
	if err != nil {
		if err.(kafka.Error).Code() == kafka.ErrTimedOut {
			return nil, nil
		}
		return nil, err
	}
	return msg, nil
}

// CommitMessage commits a message
func (c *Consumer) CommitMessage(msg *kafka.Message) error {
	_, err := c.consumer.CommitMessage(msg)
	return err
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.consumer.Close()
}
