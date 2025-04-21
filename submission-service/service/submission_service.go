package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/nslaughter/codecourt/submission-service/config"
	"github.com/nslaughter/codecourt/submission-service/db"
	kafkalib "github.com/nslaughter/codecourt/submission-service/kafka"
	"github.com/nslaughter/codecourt/submission-service/model"
)

// SubmissionService represents the submission service
type SubmissionService struct {
	cfg      *config.Config
	db       db.Repository
	producer kafkalib.KafkaProducer
	consumer kafkalib.KafkaConsumer
}

// NewSubmissionService creates a new submission service
func NewSubmissionService(cfg *config.Config, database db.Repository, producer kafkalib.KafkaProducer, consumer kafkalib.KafkaConsumer) *SubmissionService {
	return &SubmissionService{
		cfg:      cfg,
		db:       database,
		producer: producer,
		consumer: consumer,
	}
}

// CreateSubmission creates a new submission
func (s *SubmissionService) CreateSubmission(submission *model.Submission) error {
	// Save submission to database
	if err := s.db.CreateSubmission(submission); err != nil {
		return fmt.Errorf("failed to create submission: %w", err)
	}

	// Send submission to Kafka
	submissionJSON, err := json.Marshal(submission)
	if err != nil {
		return fmt.Errorf("failed to marshal submission: %w", err)
	}

	if err := s.producer.Produce(submission.ID, submissionJSON); err != nil {
		return fmt.Errorf("failed to produce submission to Kafka: %w", err)
	}

	return nil
}

// GetSubmission gets a submission by ID
func (s *SubmissionService) GetSubmission(id string) (*model.Submission, error) {
	return s.db.GetSubmission(id)
}

// GetSubmissionResult gets a submission result by submission ID
func (s *SubmissionService) GetSubmissionResult(submissionID string) (*model.SubmissionResult, error) {
	return s.db.GetSubmissionResult(submissionID)
}

// GetSubmissionsByUserID gets all submissions for a user
func (s *SubmissionService) GetSubmissionsByUserID(userID string) ([]*model.Submission, error) {
	return s.db.GetSubmissionsByUserID(userID)
}

// GetSubmissionsByProblemID gets all submissions for a problem
func (s *SubmissionService) GetSubmissionsByProblemID(problemID string) ([]*model.Submission, error) {
	return s.db.GetSubmissionsByProblemID(problemID)
}

// ProcessJudgingResults processes judging results from Kafka
func (s *SubmissionService) ProcessJudgingResults(ctx context.Context) {
	log.Println("Starting to process judging results...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Context canceled, stopping judging results processing")
			return
		default:
			// Try to consume a message with a 100ms timeout
			msg, err := s.consumer.Consume(100 * time.Millisecond)
			if err != nil {
				log.Printf("Error consuming message: %v", err)
				continue
			}

			// No message received, continue
			if msg == nil {
				continue
			}

			// Process the message
			if err := s.processJudgingResult(msg); err != nil {
				log.Printf("Error processing judging result: %v", err)
			}

			// Commit the message
			if err := s.consumer.CommitMessage(msg); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

// processJudgingResult processes a single judging result
func (s *SubmissionService) processJudgingResult(msg *kafka.Message) error {
	// Parse the judging result
	var result model.SubmissionResult
	if err := json.Unmarshal(msg.Value, &result); err != nil {
		return fmt.Errorf("failed to unmarshal judging result: %w", err)
	}

	// Save the result to the database
	if err := s.db.SaveSubmissionResult(&result); err != nil {
		return fmt.Errorf("failed to save judging result: %w", err)
	}

	// Update the submission status
	if err := s.db.UpdateSubmissionStatus(result.SubmissionID, string(result.Status)); err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	log.Printf("Processed judging result for submission %s with status %s", result.SubmissionID, result.Status)
	return nil
}

// Close closes the service
func (s *SubmissionService) Close() {
	// Nothing to close in the service itself
}
