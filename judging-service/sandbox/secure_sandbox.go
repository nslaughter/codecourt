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

// SecureSandbox implements a sandbox that runs code in a secure container
type SecureSandbox struct {
	BaseSandbox
}

// NewSecureSandbox creates a new secure sandbox
func NewSecureSandbox(workDir string, maxExecutionTime time.Duration, maxMemoryUsage int64) *SecureSandbox {
	return &SecureSandbox{
		BaseSandbox: NewBaseSandbox(workDir, maxExecutionTime, maxMemoryUsage),
	}
}

// Compile compiles the code if needed
func (s *SecureSandbox) Compile(ctx context.Context, language model.Language, code string) (string, error) {
	// Create workspace
	workspace, err := s.createWorkspace()
	if err != nil {
		return "", err
	}
	defer s.cleanup(workspace)

	// Write code to file
	filePath, err := s.writeCodeToFile(workspace, language, code)
	if err != nil {
		return "", err
	}

	// Prepare Docker command for compilation
	var compileOutput bytes.Buffer
	var compileCmd *exec.Cmd

	// Base Docker command with security constraints
	dockerArgs := []string{
		"run",
		"--rm",                                   // Remove container after execution
		"--network=none",                         // No network access
		"--cpus=1",                               // Limit to 1 CPU
		"--memory=512m",                          // Limit memory to 512MB
		"--memory-swap=512m",                     // Disable swap
		"--pids-limit=50",                        // Limit number of processes
		"--security-opt=no-new-privileges",       // Prevent privilege escalation
		"--cap-drop=ALL",                         // Drop all capabilities
		"--user=nobody",                          // Run as non-root user
		"-v", fmt.Sprintf("%s:/code:ro", workspace), // Mount code directory as read-only
		"-w", "/code",                            // Set working directory
	}

	switch language {
	case model.LanguageGo:
		// Go compilation
		dockerArgs = append(dockerArgs, "golang:1.21-alpine", "go", "build", "-o", "main", filepath.Base(filePath))
	case model.LanguageC:
		// C compilation
		dockerArgs = append(dockerArgs, "gcc:latest", "gcc", "-o", "main", filepath.Base(filePath))
	case model.LanguageCPP:
		// C++ compilation
		dockerArgs = append(dockerArgs, "gcc:latest", "g++", "-o", "main", filepath.Base(filePath))
	case model.LanguageJava:
		// Java compilation
		dockerArgs = append(dockerArgs, "openjdk:17-slim", "javac", filepath.Base(filePath))
	case model.LanguagePython:
		// Python doesn't need compilation, just syntax check
		dockerArgs = append(dockerArgs, "python:3.10-alpine", "python", "-m", "py_compile", filepath.Base(filePath))
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	compileCmd = exec.CommandContext(ctx, "docker", dockerArgs...)
	compileCmd.Stdout = &compileOutput
	compileCmd.Stderr = &compileOutput

	// Set a timeout for compilation
	_, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Run the compilation
	err = compileCmd.Run()
	if err != nil {
		return compileOutput.String(), fmt.Errorf("compilation failed: %w", err)
	}

	return compileOutput.String(), nil
}

// Execute executes the code with the given input
func (s *SecureSandbox) Execute(ctx context.Context, language model.Language, code string, input string) (string, time.Duration, int64, error) {
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
	if _, err := s.Compile(ctx, language, code); err != nil {
		return "", 0, 0, err
	}

	// Create output directory
	outputDir := filepath.Join(workspace, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", 0, 0, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Prepare Docker command for execution
	var outputBuffer bytes.Buffer

	// Base Docker command with security constraints
	dockerArgs := []string{
		"run",
		"--rm",                                   // Remove container after execution
		"--network=none",                         // No network access
		"--cpus=1",                               // Limit to 1 CPU
		fmt.Sprintf("--memory=%dm", s.maxMemoryUsage/(1024*1024)), // Memory limit
		fmt.Sprintf("--memory-swap=%dm", s.maxMemoryUsage/(1024*1024)), // Disable swap
		"--pids-limit=50",                        // Limit number of processes
		"--security-opt=no-new-privileges",       // Prevent privilege escalation
		"--cap-drop=ALL",                         // Drop all capabilities
		"--user=nobody",                          // Run as non-root user
		"-v", fmt.Sprintf("%s:/code:ro", workspace), // Mount code directory as read-only
		"-v", fmt.Sprintf("%s:/input:ro", inputPath), // Mount input file as read-only
		"-v", fmt.Sprintf("%s:/output:rw", outputDir), // Mount output directory as writable
		"-w", "/code",                            // Set working directory
	}

	// Add ulimit for CPU time
	timeoutSecs := int(s.maxExecutionTime.Seconds()) + 1
	dockerArgs = append(dockerArgs, "--ulimit", fmt.Sprintf("cpu=%d:%d", timeoutSecs, timeoutSecs))

	// Add command based on language
	var execCmd []string
	switch language {
	case model.LanguageGo:
		dockerArgs = append(dockerArgs, "golang:1.21-alpine")
		execCmd = []string{"/bin/sh", "-c", "cat /input | ./main > /output/result.txt 2>&1"}
	case model.LanguageC, model.LanguageCPP:
		dockerArgs = append(dockerArgs, "gcc:latest")
		execCmd = []string{"/bin/sh", "-c", "cat /input | ./main > /output/result.txt 2>&1"}
	case model.LanguageJava:
		// Extract class name from file path
		className := filepath.Base(filePath)
		className = className[:len(className)-5] // Remove .java extension
		dockerArgs = append(dockerArgs, "openjdk:17-slim")
		execCmd = []string{"/bin/sh", "-c", fmt.Sprintf("cat /input | java %s > /output/result.txt 2>&1", className)}
	case model.LanguagePython:
		dockerArgs = append(dockerArgs, "python:3.10-alpine")
		execCmd = []string{"/bin/sh", "-c", fmt.Sprintf("cat /input | python %s > /output/result.txt 2>&1", filepath.Base(filePath))}
	default:
		return "", 0, 0, fmt.Errorf("unsupported language: %s", language)
	}

	dockerArgs = append(dockerArgs, execCmd...)
	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &outputBuffer

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

	// Read output file
	outputFile := filepath.Join(outputDir, "result.txt")
	output, err := os.ReadFile(outputFile)
	if err != nil && !os.IsNotExist(err) {
		return "", executionTime, 0, fmt.Errorf("failed to read output file: %w", err)
	}

	// Get memory usage from Docker stats
	// This is a placeholder - in a real implementation, you would parse Docker stats
	// For now, we'll just use a simple estimation
	memoryUsed := int64(len(output) * 10) // Simple placeholder

	// If we got a timeout or other error, but we have some output, return it along with the error
	if execErr != nil && len(output) > 0 {
		return string(output), executionTime, memoryUsed, execErr
	}

	return string(output), executionTime, memoryUsed, execErr
}
