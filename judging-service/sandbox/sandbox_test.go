package sandbox

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/nslaughter/codecourt/judging-service/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalSandbox(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sandbox-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a local sandbox
	sandbox := NewLocalSandbox(tempDir, 5*time.Second, 100*1024*1024)

	// Define test cases
	tests := []struct {
		name           string
		language       model.Language
		code           string
		input          string
		expectedOutput string
		shouldPass     bool
	}{
		{
			name:     "Go Hello World",
			language: model.LanguageGo,
			code: `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`,
			input:          "",
			expectedOutput: "Hello, World!",
			shouldPass:     true,
		},
		{
			name:     "Go Echo Input",
			language: model.LanguageGo,
			code: `package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}`,
			input:          "Echo this",
			expectedOutput: "Echo this",
			shouldPass:     true,
		},
		{
			name:     "Go Compilation Error",
			language: model.LanguageGo,
			code: `package main

func main() {
	fmt.Println("Hello, World!")
}`,
			input:          "",
			expectedOutput: "",
			shouldPass:     false,
		},
		{
			name:     "Python Hello World",
			language: model.LanguagePython,
			code:     `print("Hello, World!")`,
			input:          "",
			expectedOutput: "Hello, World!",
			shouldPass:     true,
		},
		{
			name:     "Python Echo Input",
			language: model.LanguagePython,
			code:     `print(input())`,
			input:          "Echo this",
			expectedOutput: "Echo this",
			shouldPass:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Skip if the language is not available on the test machine
			if tc.language == model.LanguageGo && !isCommandAvailable("go") {
				t.Skip("Go is not available")
			}
			if tc.language == model.LanguagePython && !isCommandAvailable("python3") {
				t.Skip("Python is not available")
			}

			// Compile the code
			compileOutput, err := sandbox.Compile(context.Background(), tc.language, tc.code)
			if !tc.shouldPass {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err, "Compilation failed: %s", compileOutput)

			// Execute the code
			output, executionTime, memoryUsed, err := sandbox.Execute(context.Background(), tc.language, tc.code, tc.input)
			require.NoError(t, err)

			// Check the output
			assert.Contains(t, output, tc.expectedOutput)
			
			// Check that execution time and memory usage are reasonable
			assert.Greater(t, executionTime.Nanoseconds(), int64(0))
			assert.Less(t, executionTime, 5*time.Second)
			assert.Greater(t, memoryUsed, int64(0))
		})
	}
}

// TestBaseSandbox tests the base sandbox functionality
func TestBaseSandbox(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sandbox-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a base sandbox
	sandbox := NewBaseSandbox(tempDir, 5*time.Second, 100*1024*1024)

	// Test createWorkspace
	workspace, err := sandbox.createWorkspace()
	require.NoError(t, err)
	defer os.RemoveAll(workspace)

	// Check that the workspace exists
	_, err = os.Stat(workspace)
	assert.NoError(t, err)

	// Test writeCodeToFile
	code := "package main\n\nfunc main() {}"
	filePath, err := sandbox.writeCodeToFile(workspace, model.LanguageGo, code)
	require.NoError(t, err)

	// Check that the file exists and contains the code
	fileContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, code, string(fileContent))
	
	// Verify the file path is correct
	expectedFilePath := filepath.Join(workspace, "main.go")
	assert.Equal(t, expectedFilePath, filePath)

	// Test writeInputToFile
	input := "test input"
	inputPath, err := sandbox.writeInputToFile(workspace, input)
	require.NoError(t, err)

	// Check that the file exists and contains the input
	inputContent, err := os.ReadFile(inputPath)
	require.NoError(t, err)
	assert.Equal(t, input, string(inputContent))

	// Test cleanup
	sandbox.cleanup(workspace)

	// Check that the workspace no longer exists
	_, err = os.Stat(workspace)
	assert.True(t, os.IsNotExist(err))
}

// Helper function to check if a command is available
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// Integration test for SecureSandbox - only run if Docker is available and test is explicitly enabled
func TestSecureSandbox(t *testing.T) {
	// Skip by default unless explicitly enabled with environment variable
	if os.Getenv("ENABLE_DOCKER_TESTS") != "true" {
		t.Skip("Docker tests are disabled by default. Set ENABLE_DOCKER_TESTS=true to enable")
	}
	
	// Skip if Docker is not available
	if !isCommandAvailable("docker") {
		t.Skip("Docker is not available")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sandbox-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a secure sandbox
	sandbox := NewSecureSandbox(tempDir, 5*time.Second, 100*1024*1024)

	// Test with a simple Go program
	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello from Docker!")
}
`
	// Compile the code
	compileOutput, err := sandbox.Compile(context.Background(), model.LanguageGo, code)
	require.NoError(t, err, "Compilation failed: %s", compileOutput)

	// Execute the code
	output, executionTime, memoryUsed, err := sandbox.Execute(context.Background(), model.LanguageGo, code, "")
	require.NoError(t, err)

	// Check the output
	assert.Contains(t, output, "Hello from Docker!")
	
	// Check that execution time and memory usage are reasonable
	assert.Greater(t, executionTime.Nanoseconds(), int64(0))
	assert.Less(t, executionTime, 5*time.Second)
	assert.Greater(t, memoryUsed, int64(0))
}
