package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/nslaughter/codecourt/judging-service/model"
)

// Sandbox defines the interface for code execution sandboxes
type Sandbox interface {
	// Compile compiles the code if needed and returns any compilation output or error
	Compile(ctx context.Context, language model.Language, code string) (string, error)
	
	// Execute executes the code with the given input and returns the output, execution time, memory usage, and any error
	Execute(ctx context.Context, language model.Language, code string, input string) (string, time.Duration, int64, error)
}

// BaseSandbox provides common functionality for sandbox implementations
type BaseSandbox struct {
	workDir         string
	maxExecutionTime time.Duration
	maxMemoryUsage   int64
}

// NewBaseSandbox creates a new base sandbox
func NewBaseSandbox(workDir string, maxExecutionTime time.Duration, maxMemoryUsage int64) BaseSandbox {
	return BaseSandbox{
		workDir:         workDir,
		maxExecutionTime: maxExecutionTime,
		maxMemoryUsage:   maxMemoryUsage,
	}
}

// createWorkspace creates a temporary workspace for code execution
func (s *BaseSandbox) createWorkspace() (string, error) {
	// Create a unique directory for this execution
	workspaceID := uuid.New().String()
	workspacePath := filepath.Join(s.workDir, workspaceID)
	
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %w", err)
	}
	
	return workspacePath, nil
}

// writeCodeToFile writes code to a file in the workspace
func (s *BaseSandbox) writeCodeToFile(workspace string, language model.Language, code string) (string, error) {
	// Determine file extension based on language
	var extension string
	switch language {
	case model.LanguageGo:
		extension = ".go"
	case model.LanguagePython:
		extension = ".py"
	case model.LanguageJava:
		extension = ".java"
	case model.LanguageC:
		extension = ".c"
	case model.LanguageCPP:
		extension = ".cpp"
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}
	
	// Create the file
	filename := "main" + extension
	filePath := filepath.Join(workspace, filename)
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write code to file: %w", err)
	}
	
	return filePath, nil
}

// writeInputToFile writes input to a file in the workspace
func (s *BaseSandbox) writeInputToFile(workspace string, input string) (string, error) {
	// Create the input file
	inputPath := filepath.Join(workspace, "input.txt")
	
	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		return "", fmt.Errorf("failed to write input to file: %w", err)
	}
	
	return inputPath, nil
}

// cleanup removes the workspace directory
func (s *BaseSandbox) cleanup(workspace string) {
	os.RemoveAll(workspace)
}
