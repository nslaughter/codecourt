package kafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaProducer defines the interface for Kafka producer operations
type KafkaProducer interface {
	Produce(key string, value []byte) error
	Close()
}

// KafkaConsumer defines the interface for Kafka consumer operations
type KafkaConsumer interface {
	Consume(timeout time.Duration) (*kafka.Message, error)
	CommitMessage(msg *kafka.Message) error
	Close() error
}
