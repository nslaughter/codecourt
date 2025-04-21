package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/nslaughter/codecourt/submission-service/config"
)

// Producer represents a Kafka producer
type Producer struct {
	producer *kafka.Producer
	topic    string
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg *config.Config) (*Producer, error) {
	// Create Kafka producer configuration
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokers,
	}

	// Create Kafka producer
	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &Producer{
		producer: producer,
		topic:    cfg.KafkaSubmissionTopic,
	}, nil
}

// Produce produces a message to Kafka
func (p *Producer) Produce(key string, value []byte) error {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: value,
	}

	// Produce the message
	if err := p.producer.Produce(message, nil); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Wait for message delivery
	p.producer.Flush(15 * 1000) // Wait up to 15 seconds

	return nil
}

// Close closes the producer
func (p *Producer) Close() {
	p.producer.Close()
}
