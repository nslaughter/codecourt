package service

import "github.com/nslaughter/codecourt/submission-service/model"

// SubmissionServiceInterface defines the interface for submission service operations
type SubmissionServiceInterface interface {
	CreateSubmission(submission *model.Submission) error
	GetSubmission(id string) (*model.Submission, error)
	GetSubmissionResult(submissionID string) (*model.SubmissionResult, error)
	GetSubmissionsByUserID(userID string) ([]*model.Submission, error)
	GetSubmissionsByProblemID(problemID string) ([]*model.Submission, error)
}
