package db

import "github.com/nslaughter/codecourt/submission-service/model"

// Repository defines the interface for database operations
type Repository interface {
	CreateSubmission(submission *model.Submission) error
	GetSubmission(id string) (*model.Submission, error)
	UpdateSubmissionStatus(id string, status string) error
	SaveSubmissionResult(result *model.SubmissionResult) error
	GetSubmissionsByUserID(userID string) ([]*model.Submission, error)
	GetSubmissionsByProblemID(problemID string) ([]*model.Submission, error)
	GetSubmissionResult(submissionID string) (*model.SubmissionResult, error)
	Close() error
}
