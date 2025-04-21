package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/nslaughter/codecourt/judging-service/config"
	"github.com/nslaughter/codecourt/judging-service/db"
	kafkalib "github.com/nslaughter/codecourt/judging-service/kafka"
	"github.com/nslaughter/codecourt/judging-service/model"
	"github.com/nslaughter/codecourt/judging-service/sandbox"
)

// JudgingService handles the judging of code submissions
type JudgingService struct {
	cfg     *config.Config
	db      *db.DB
	sandbox sandbox.Sandbox
	workers chan struct{}
}

// NewJudgingService creates a new judging service
func NewJudgingService(cfg *config.Config) (*JudgingService, error) {
	// Initialize database connection
	database, err := db.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize sandbox
	var sb sandbox.Sandbox
	if cfg.SandboxEnabled {
		sb = sandbox.NewSecureSandbox(cfg.WorkDir, cfg.MaxExecutionTime, cfg.MaxMemoryUsage)
	} else {
		sb = sandbox.NewLocalSandbox(cfg.WorkDir, cfg.MaxExecutionTime, cfg.MaxMemoryUsage)
	}

	return &JudgingService{
		cfg:     cfg,
		db:      database,
		sandbox: sb,
		workers: make(chan struct{}, cfg.ConcurrentJudges),
	}, nil
}

// Close closes the judging service
func (s *JudgingService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// ProcessSubmissions processes code submissions from Kafka
func (s *JudgingService) ProcessSubmissions(ctx context.Context, consumer *kafkalib.Consumer, producer *kafkalib.Producer) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Context canceled, stopping submission processing")
			return
		default:
			// Try to consume a message with a 100ms timeout
			msg, err := consumer.Consume(100 * time.Millisecond)
			if err != nil {
				log.Printf("Error consuming message: %v", err)
				continue
			}

			// No message received, continue
			if msg == nil {
				continue
			}

			// Process the message
			go func(msg *kafka.Message) {
				s.processSubmission(ctx, msg, consumer, producer)
			}(msg)
		}
	}
}

// processSubmission processes a single submission
func (s *JudgingService) processSubmission(ctx context.Context, msg *kafka.Message, consumer *kafkalib.Consumer, producer *kafkalib.Producer) {
	// Acquire a worker slot
	s.workers <- struct{}{}
	defer func() {
		// Release the worker slot
		<-s.workers
	}()

	// Parse the submission
	var submission model.Submission
	if err := json.Unmarshal(msg.Value, &submission); err != nil {
		log.Printf("Error unmarshaling submission: %v", err)
		consumer.Commit()
		return
	}

	log.Printf("Processing submission %s for problem %s", submission.ID, submission.ProblemID)

	// Update submission status to running
	if err := s.db.UpdateSubmissionStatus(submission.ID, model.StatusRunning); err != nil {
		log.Printf("Error updating submission status: %v", err)
		consumer.Commit()
		return
	}

	// Get test cases for the problem
	testCases, err := s.db.GetTestCases(submission.ProblemID)
	if err != nil {
		log.Printf("Error getting test cases: %v", err)
		s.handleError(submission.ID, err, producer)
		consumer.Commit()
		return
	}

	if len(testCases) == 0 {
		err := fmt.Errorf("no test cases found for problem %s", submission.ProblemID)
		log.Printf("%v", err)
		s.handleError(submission.ID, err, producer)
		consumer.Commit()
		return
	}

	// Judge the submission
	result, err := s.judgeSubmission(ctx, &submission, testCases)
	if err != nil {
		log.Printf("Error judging submission: %v", err)
		s.handleError(submission.ID, err, producer)
		consumer.Commit()
		return
	}

	// Save the judging result
	if err := s.db.SaveJudgingResult(result); err != nil {
		log.Printf("Error saving judging result: %v", err)
		s.handleError(submission.ID, err, producer)
		consumer.Commit()
		return
	}

	// Send the result to Kafka
	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling judging result: %v", err)
		consumer.Commit()
		return
	}

	// Produce the result message
	if err := producer.Produce(submission.ID, resultBytes); err != nil {
		log.Printf("Error producing judging result: %v", err)
		consumer.Commit()
		return
	}

	log.Printf("Successfully judged submission %s with status %s", submission.ID, result.Status)
	consumer.Commit()
}

// judgeSubmission judges a submission against test cases
func (s *JudgingService) judgeSubmission(ctx context.Context, submission *model.Submission, testCases []model.TestCase) (*model.JudgingResult, error) {
	// Create a result with the submission ID
	result := &model.JudgingResult{
		SubmissionID: submission.ID,
		Status:       model.StatusPending,
		JudgedAt:     time.Now(),
	}

	// Compile the code if needed
	compileOutput, err := s.sandbox.Compile(ctx, submission.Language, submission.Code)
	if err != nil {
		result.Status = model.StatusCompilationError
		result.CompileOutput = compileOutput
		result.Error = err.Error()
		return result, nil
	}

	result.CompileOutput = compileOutput

	// Run test cases
	var wg sync.WaitGroup
	testResults := make([]model.TestResult, len(testCases))
	var mu sync.Mutex
	var maxExecutionTime time.Duration
	var maxMemoryUsed int64

	for i, tc := range testCases {
		wg.Add(1)
		go func(i int, tc model.TestCase) {
			defer wg.Done()

			// Run the test case
			output, executionTime, memoryUsed, err := s.sandbox.Execute(ctx, submission.Language, submission.Code, tc.Input)
			
			// Create test result
			testResult := model.TestResult{
				TestCaseID:    tc.ID,
				ActualOutput:  output,
				ExecutionTime: executionTime,
				MemoryUsed:    memoryUsed,
			}

			// Check for errors
			if err != nil {
				testResult.Passed = false
				testResult.Error = err.Error()
				
				// Determine error type
				if executionTime >= s.cfg.MaxExecutionTime {
					testResult.Error = "Time limit exceeded"
				} else if memoryUsed >= s.cfg.MaxMemoryUsage {
					testResult.Error = "Memory limit exceeded"
				}
			} else {
				// Compare output with expected output
				testResult.Passed = compareOutput(output, tc.Output)
			}

			// Update test results and track max resource usage
			mu.Lock()
			testResults[i] = testResult
			if executionTime > maxExecutionTime {
				maxExecutionTime = executionTime
			}
			if memoryUsed > maxMemoryUsed {
				maxMemoryUsed = memoryUsed
			}
			mu.Unlock()
		}(i, tc)
	}

	// Wait for all test cases to complete
	wg.Wait()

	// Set resource usage
	result.ExecutionTime = maxExecutionTime
	result.MemoryUsed = maxMemoryUsed
	result.TestResults = testResults

	// Determine overall status
	result.Status = determineStatus(testResults, maxExecutionTime, maxMemoryUsed, s.cfg.MaxExecutionTime, s.cfg.MaxMemoryUsage)

	return result, nil
}

// handleError handles an error during submission processing
func (s *JudgingService) handleError(submissionID string, err error, producer *kafkalib.Producer) {
	// Create an error result
	result := &model.JudgingResult{
		SubmissionID: submissionID,
		Status:       model.StatusError,
		Error:        err.Error(),
		JudgedAt:     time.Now(),
	}

	// Save the error result
	if dbErr := s.db.SaveJudgingResult(result); dbErr != nil {
		log.Printf("Error saving error result: %v", dbErr)
	}

	// Send the error result to Kafka
	resultBytes, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		log.Printf("Error marshaling error result: %v", marshalErr)
		return
	}

	// Produce the error result message
	if err := producer.Produce(submissionID, resultBytes); err != nil {
		log.Printf("Error producing error result: %v", err)
		return
	}
}

// compareOutput compares the actual output with the expected output
func compareOutput(actual, expected string) bool {
	// Normalize line endings and trim whitespace
	// This is a simple comparison, but could be extended with more sophisticated algorithms
	return normalizeOutput(actual) == normalizeOutput(expected)
}

// normalizeOutput normalizes output by trimming whitespace and normalizing line endings
func normalizeOutput(output string) string {
	// Replace Windows line endings with Unix line endings
	// Trim trailing whitespace
	// This is a simple normalization, but could be extended with more sophisticated algorithms
	return output
}

// determineStatus determines the overall status based on test results
func determineStatus(testResults []model.TestResult, executionTime time.Duration, memoryUsed int64, maxExecutionTime time.Duration, maxMemoryUsage int64) model.Status {
	// Check for time limit exceeded
	if executionTime >= maxExecutionTime {
		return model.StatusTimeLimitExceeded
	}

	// Check for memory limit exceeded
	if memoryUsed >= maxMemoryUsage {
		return model.StatusMemoryLimitExceeded
	}

	// Check for runtime errors
	for _, tr := range testResults {
		if tr.Error != "" {
			return model.StatusRuntimeError
		}
	}

	// Check if all test cases passed
	allPassed := true
	for _, tr := range testResults {
		if !tr.Passed {
			allPassed = false
			break
		}
	}

	if allPassed {
		return model.StatusAccepted
	} else {
		return model.StatusRejected
	}
}
