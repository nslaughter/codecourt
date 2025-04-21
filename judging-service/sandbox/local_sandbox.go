package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/nslaughter/codecourt/judging-service/model"
)

// LocalSandbox implements a sandbox that runs code locally (for development only)
type LocalSandbox struct {
	BaseSandbox
}

// NewLocalSandbox creates a new local sandbox
func NewLocalSandbox(workDir string, maxExecutionTime time.Duration, maxMemoryUsage int64) *LocalSandbox {
	return &LocalSandbox{
		BaseSandbox: NewBaseSandbox(workDir, maxExecutionTime, maxMemoryUsage),
	}
}

// Compile compiles the code if needed
func (s *LocalSandbox) Compile(ctx context.Context, language model.Language, code string) (string, error) {
	// Create workspace
	workspace, err := s.createWorkspace()
	if err != nil {
		return "", err
	}

	// Write code to file
	filePath, err := s.writeCodeToFile(workspace, language, code)
	if err != nil {
		s.cleanup(workspace)
		return "", err
	}

	// Compile the code if needed
	var compileOutput bytes.Buffer
	var compileCmd *exec.Cmd

	switch language {
	case model.LanguageGo:
		// Go compilation check
		compileCmd = exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(workspace, "main"), filePath)
	case model.LanguageC:
		// C compilation
		compileCmd = exec.CommandContext(ctx, "gcc", "-o", filepath.Join(workspace, "main"), filePath)
	case model.LanguageCPP:
		// C++ compilation
		compileCmd = exec.CommandContext(ctx, "g++", "-o", filepath.Join(workspace, "main"), filePath)
	case model.LanguageJava:
		// Java compilation
		compileCmd = exec.CommandContext(ctx, "javac", filePath)
	case model.LanguagePython:
		// Python doesn't need compilation, just syntax check
		compileCmd = exec.CommandContext(ctx, "python3", "-m", "py_compile", filePath)
	default:
		s.cleanup(workspace)
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	compileCmd.Dir = workspace
	compileCmd.Stdout = &compileOutput
	compileCmd.Stderr = &compileOutput

	// Set a timeout for compilation
	_, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	compileCmd.Cancel = func() error {
		return compileCmd.Process.Kill()
	}

	// Run the compilation
	err = compileCmd.Run()
	if err != nil {
		s.cleanup(workspace)
		return compileOutput.String(), fmt.Errorf("compilation failed: %w", err)
	}

	return compileOutput.String(), nil
}

// Execute executes the code with the given input
func (s *LocalSandbox) Execute(ctx context.Context, language model.Language, code string, input string) (string, time.Duration, int64, error) {
	// Create workspace
	workspace, err := s.createWorkspace()
	if err != nil {
		return "", 0, 0, err
	}
	defer s.cleanup(workspace)

	// Write code to file
	filePath, err := s.writeCodeToFile(workspace, language, code)
	if err != nil {
		return "", 0, 0, err
	}

	// Write input to file
	inputPath, err := s.writeInputToFile(workspace, input)
	if err != nil {
		return "", 0, 0, err
	}

	// Compile the code if needed
	// For the test, we'll compile directly here instead of calling s.Compile
	// to avoid workspace cleanup issues
	var compileOutput bytes.Buffer
	var compileCmd *exec.Cmd

	switch language {
	case model.LanguageGo:
		// Go compilation check
		compileCmd = exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(workspace, "main"), filePath)
	case model.LanguageC:
		// C compilation
		compileCmd = exec.CommandContext(ctx, "gcc", "-o", filepath.Join(workspace, "main"), filePath)
	case model.LanguageCPP:
		// C++ compilation
		compileCmd = exec.CommandContext(ctx, "g++", "-o", filepath.Join(workspace, "main"), filePath)
	case model.LanguageJava:
		// Java compilation
		compileCmd = exec.CommandContext(ctx, "javac", filePath)
	case model.LanguagePython:
		// Python doesn't need compilation, just syntax check
		compileCmd = exec.CommandContext(ctx, "python3", "-m", "py_compile", filePath)
	default:
		return "", 0, 0, fmt.Errorf("unsupported language: %s", language)
	}

	compileCmd.Dir = workspace
	compileCmd.Stdout = &compileOutput
	compileCmd.Stderr = &compileOutput

	// Run the compilation
	err = compileCmd.Run()
	if err != nil && language != model.LanguagePython {
		return "", 0, 0, fmt.Errorf("compilation failed: %w", err)
	}

	// Prepare execution command
	var cmd *exec.Cmd
	switch language {
	case model.LanguageGo:
		cmd = exec.CommandContext(ctx, filepath.Join(workspace, "main"))
	case model.LanguageC, model.LanguageCPP:
		cmd = exec.CommandContext(ctx, filepath.Join(workspace, "main"))
	case model.LanguageJava:
		// Extract class name from file path
		className := filepath.Base(filePath)
		className = className[:len(className)-5] // Remove .java extension
		cmd = exec.CommandContext(ctx, "java", "-cp", workspace, className)
	case model.LanguagePython:
		cmd = exec.CommandContext(ctx, "python3", filePath)
	default:
		return "", 0, 0, fmt.Errorf("unsupported language: %s", language)
	}

	// Set up input/output
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	var outputBuffer bytes.Buffer
	cmd.Stdin = inputFile
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &outputBuffer
	cmd.Dir = workspace

	// Set a timeout for execution
	execCtx, cancel := context.WithTimeout(ctx, s.maxExecutionTime)
	defer cancel()

	// Run the command and measure execution time
	startTime := time.Now()
	err = cmd.Start()
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to start execution: %w", err)
	}

	// Wait for completion or timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var execErr error
	select {
	case <-execCtx.Done():
		// Execution timed out
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		execErr = fmt.Errorf("execution timed out after %v", s.maxExecutionTime)
	case err := <-done:
		// Execution completed
		execErr = err
	}

	executionTime := time.Since(startTime)

	// Get memory usage (this is a simplistic approach, in a real system you'd want to use cgroups or similar)
	// For now, we'll just estimate based on output size as a placeholder
	memoryUsed := int64(outputBuffer.Len() * 2) // Simple placeholder

	// Read output
	output := outputBuffer.String()

	return output, executionTime, memoryUsed, execErr
}
