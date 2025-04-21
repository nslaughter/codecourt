package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/nslaughter/codecourt/judging-service/config"
)

// Consumer represents a Kafka consumer
type Consumer struct {
	// Exposing the consumer field to allow direct access in the service
	Consumer *kafka.Consumer
	topic    string
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg *config.Config) (*Consumer, error) {
	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       cfg.KafkaBootstrapServers,
		"group.id":                cfg.KafkaGroupID,
		"auto.offset.reset":       cfg.KafkaAutoOffsetReset,
		"session.timeout.ms":      cfg.KafkaSessionTimeoutMs,
		"max.poll.interval.ms":    cfg.KafkaMaxPollIntervalMs,
		"enable.auto.commit":      cfg.KafkaEnableAutoCommit,
		"auto.commit.interval.ms": cfg.KafkaAutoCommitIntervalMs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	if err := kafkaConsumer.SubscribeTopics([]string{cfg.KafkaSubmissionTopic}, nil); err != nil {
		kafkaConsumer.Close()
		return nil, fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	return &Consumer{
		Consumer: kafkaConsumer,
		topic:    cfg.KafkaSubmissionTopic,
	}, nil
}

// Consume consumes a message from Kafka with timeout
func (c *Consumer) Consume(timeout time.Duration) (*kafka.Message, error) {
	msg, err := c.Consumer.ReadMessage(timeout)
	if err != nil {
		// Check if it's just a timeout, which is not a real error
		if err.(kafka.Error).Code() == kafka.ErrTimedOut {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read message: %w", err)
	}
	return msg, nil
}

// Commit commits a message offset
func (c *Consumer) Commit() error {
	_, err := c.Consumer.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit offsets: %w", err)
	}
	return nil
}

// Close closes the consumer
func (c *Consumer) Close() {
	if c.Consumer != nil {
		c.Consumer.Close()
	}
}
