package model

import "errors"

// Error constants
var (
	// ErrProblemNotFound is returned when a problem is not found
	ErrProblemNotFound = errors.New("problem not found")
	
	// ErrTestCaseNotFound is returned when a test case is not found
	ErrTestCaseNotFound = errors.New("test case not found")
	
	// ErrCategoryNotFound is returned when a category is not found
	ErrCategoryNotFound = errors.New("category not found")
	
	// ErrTemplateNotFound is returned when a template is not found
	ErrTemplateNotFound = errors.New("template not found")
	
	// ErrInvalidRequest is returned when a request is invalid
	ErrInvalidRequest = errors.New("invalid request")
	
	// ErrDatabaseError is returned when a database error occurs
	ErrDatabaseError = errors.New("database error")
)
