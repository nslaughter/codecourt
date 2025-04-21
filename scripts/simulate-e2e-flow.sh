#!/bin/bash
set -eo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script variables
API_URL="http://localhost:8080"
JWT_TOKEN="simulated-jwt-token"

# Log functions
log_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Function to simulate user registration
simulate_registration() {
  log_info "Simulating user registration..."
  echo '{
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2025-04-21T12:45:00Z"
  }'
  log_success "User registered successfully!"
}

# Function to simulate user login
simulate_login() {
  log_info "Simulating user login..."
  echo '{
    "token": "simulated-jwt-token",
    "refresh_token": "simulated-refresh-token",
    "expires_at": "2025-04-22T12:45:00Z"
  }'
  log_success "User logged in successfully!"
}

# Function to simulate getting problem list
simulate_problem_list() {
  log_info "Simulating getting problem list..."
  echo '[
    {
      "id": 1,
      "title": "Two Sum",
      "difficulty": "Easy",
      "category": "Arrays",
      "submission_count": 1024,
      "success_rate": 65.5
    },
    {
      "id": 2,
      "title": "Reverse Linked List",
      "difficulty": "Medium",
      "category": "Linked Lists",
      "submission_count": 768,
      "success_rate": 48.2
    },
    {
      "id": 3,
      "title": "LRU Cache",
      "difficulty": "Hard",
      "category": "Design",
      "submission_count": 512,
      "success_rate": 32.1
    }
  ]'
  log_success "Retrieved problem list successfully!"
}

# Function to simulate getting problem details
simulate_problem_details() {
  log_info "Simulating getting problem details for ID: $1..."
  echo '{
    "id": 1,
    "title": "Two Sum",
    "description": "Given an array of integers nums and an integer target, return indices of the two numbers such that they add up to target. You may assume that each input would have exactly one solution, and you may not use the same element twice.",
    "difficulty": "Easy",
    "category": "Arrays",
    "examples": [
      {
        "input": "nums = [2,7,11,15], target = 9",
        "output": "[0,1]",
        "explanation": "Because nums[0] + nums[1] == 9, we return [0, 1]."
      },
      {
        "input": "nums = [3,2,4], target = 6",
        "output": "[1,2]",
        "explanation": "Because nums[1] + nums[2] == 6, we return [1, 2]."
      }
    ],
    "constraints": [
      "2 <= nums.length <= 10^4",
      "-10^9 <= nums[i] <= 10^9",
      "-10^9 <= target <= 10^9",
      "Only one valid answer exists."
    ],
    "templates": {
      "go": "package main\n\nfunc twoSum(nums []int, target int) []int {\n    // Your code here\n}",
      "python": "class Solution:\n    def twoSum(self, nums: List[int], target: int) -> List[int]:\n        # Your code here\n        pass",
      "java": "class Solution {\n    public int[] twoSum(int[] nums, int target) {\n        // Your code here\n    }\n}"
    }
  }'
  log_success "Retrieved problem details successfully!"
}

# Function to simulate submitting a solution
simulate_submission() {
  log_info "Simulating submitting a solution..."
  echo '{
    "id": 1,
    "problem_id": 1,
    "user_id": 1,
    "language": "go",
    "status": "Pending",
    "created_at": "2025-04-21T12:50:00Z"
  }'
  log_success "Solution submitted successfully!"
}

# Function to simulate getting submission results
simulate_submission_results() {
  log_info "Simulating getting submission results for ID: $1..."
  
  # Wait to simulate processing time
  log_info "Processing submission..."
  sleep 2
  
  echo '{
    "id": 1,
    "problem_id": 1,
    "user_id": 1,
    "language": "go",
    "status": "Accepted",
    "runtime_ms": 4,
    "memory_kb": 3200,
    "test_cases": [
      {
        "input": "nums = [2,7,11,15], target = 9",
        "expected_output": "[0,1]",
        "actual_output": "[0,1]",
        "status": "Passed"
      },
      {
        "input": "nums = [3,2,4], target = 6",
        "expected_output": "[1,2]",
        "actual_output": "[1,2]",
        "status": "Passed"
      },
      {
        "input": "nums = [3,3], target = 6",
        "expected_output": "[0,1]",
        "actual_output": "[0,1]",
        "status": "Passed"
      }
    ],
    "created_at": "2025-04-21T12:50:00Z",
    "completed_at": "2025-04-21T12:50:02Z"
  }'
  log_success "Submission accepted! All test cases passed."
}

# Function to simulate getting submission history
simulate_submission_history() {
  log_info "Simulating getting submission history..."
  echo '[
    {
      "id": 1,
      "problem_id": 1,
      "problem_title": "Two Sum",
      "language": "go",
      "status": "Accepted",
      "runtime_ms": 4,
      "memory_kb": 3200,
      "created_at": "2025-04-21T12:50:00Z",
      "completed_at": "2025-04-21T12:50:02Z"
    }
  ]'
  log_success "Retrieved submission history successfully!"
}

# Main function
main() {
  log_info "Starting end-to-end workflow simulation for CodeCourt..."
  
  # Step 1: User Registration
  log_info "Step 1: User Registration"
  echo "Request: POST ${API_URL}/api/v1/users/register"
  echo "Request Body: {\"username\": \"testuser\", \"email\": \"test@example.com\", \"password\": \"Password123!\"}"
  echo "Response:"
  simulate_registration
  echo ""
  
  # Step 2: User Login
  log_info "Step 2: User Login"
  echo "Request: POST ${API_URL}/api/v1/users/login"
  echo "Request Body: {\"username\": \"testuser\", \"password\": \"Password123!\"}"
  echo "Response:"
  simulate_login
  echo ""
  
  # Step 3: Get Problem List
  log_info "Step 3: Get Problem List"
  echo "Request: GET ${API_URL}/api/v1/problems"
  echo "Headers: Authorization: Bearer ${JWT_TOKEN}"
  echo "Response:"
  simulate_problem_list
  echo ""
  
  # Step 4: Get Problem Details
  PROBLEM_ID=1
  log_info "Step 4: Get Problem Details (ID: ${PROBLEM_ID})"
  echo "Request: GET ${API_URL}/api/v1/problems/${PROBLEM_ID}"
  echo "Headers: Authorization: Bearer ${JWT_TOKEN}"
  echo "Response:"
  simulate_problem_details ${PROBLEM_ID}
  echo ""
  
  # Step 5: Submit Solution
  log_info "Step 5: Submit Solution"
  echo "Request: POST ${API_URL}/api/v1/submissions"
  echo "Headers: Authorization: Bearer ${JWT_TOKEN}"
  echo "Request Body: {\"problem_id\": 1, \"language\": \"go\", \"code\": \"package main\\n\\nimport \\\"fmt\\\"\\n\\nfunc twoSum(nums []int, target int) []int {\\n    numMap := make(map[int]int)\\n    for i, num := range nums {\\n        complement := target - num\\n        if idx, found := numMap[complement]; found {\\n            return []int{idx, i}\\n        }\\n        numMap[num] = i\\n    }\\n    return nil\\n}\"}"
  echo "Response:"
  SUBMISSION_ID=$(simulate_submission | grep -o '"id": [0-9]*' | head -1 | cut -d' ' -f2)
  echo ""
  
  # Step 6: Get Submission Results
  log_info "Step 6: Get Submission Results (ID: ${SUBMISSION_ID})"
  echo "Request: GET ${API_URL}/api/v1/submissions/${SUBMISSION_ID}"
  echo "Headers: Authorization: Bearer ${JWT_TOKEN}"
  echo "Response:"
  simulate_submission_results ${SUBMISSION_ID}
  echo ""
  
  # Step 7: Get Submission History
  log_info "Step 7: Get Submission History"
  echo "Request: GET ${API_URL}/api/v1/users/me/submissions"
  echo "Headers: Authorization: Bearer ${JWT_TOKEN}"
  echo "Response:"
  simulate_submission_history
  echo ""
  
  log_success "End-to-end workflow simulation completed successfully!"
}

# Execute main function
main "$@"
