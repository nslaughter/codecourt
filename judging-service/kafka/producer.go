package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/nslaughter/codecourt/judging-service/config"
)

// Producer represents a Kafka producer
type Producer struct {
	// Exposing the producer field to allow direct access in the service
	Producer *kafka.Producer
	topic    string
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg *config.Config) (*Producer, error) {
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBootstrapServers,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &Producer{
		Producer: kafkaProducer,
		topic:    cfg.KafkaResultTopic,
	}, nil
}

// Produce produces a message to Kafka
func (p *Producer) Produce(key string, value []byte) error {
	if err := p.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: value,
	}, nil); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Flush to ensure delivery
	remaining := p.Producer.Flush(5000) // 5 seconds timeout
	if remaining > 0 {
		return fmt.Errorf("failed to deliver %d messages", remaining)
	}

	return nil
}

// Close closes the producer
func (p *Producer) Close() {
	if p.Producer != nil {
		p.Producer.Flush(15 * 1000) // 15 seconds timeout
		p.Producer.Close()
	}
}
