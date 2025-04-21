# CodeCourt End-to-End Testing Guide

This document provides instructions for running end-to-end tests and simulations to validate the complete workflow of the CodeCourt system.

## Prerequisites

Before running the end-to-end tests, ensure you have:

- A running Kubernetes cluster (Kind is recommended for local development)
- The CodeCourt Helm chart installed
- All services up and running
- kubectl configured to access your cluster

## Running the End-to-End Test Script

The repository includes a script for running comprehensive end-to-end tests:

```bash
# Run the end-to-end tests
make e2e-test
```

This script:
1. Creates a Kind cluster if one doesn't exist
2. Installs the NGINX Ingress Controller
3. Installs the CodeCourt Helm chart with all dependencies
4. Creates necessary secrets for all services
5. Waits for all deployments to be ready
6. Runs a series of tests to validate the system functionality
7. Cleans up resources (optional)

## Simulated End-to-End Workflow

For demonstration purposes, we also provide a simulation script that shows the expected API interactions and responses:

```bash
# Run the simulation script
./scripts/simulate-e2e-flow.sh
```

This script simulates a complete user journey through the CodeCourt system:

### 1. User Registration

```
POST /api/v1/users/register
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "Password123!"
}
```

Response:
```json
{
  "id": 1,
  "username": "testuser",
  "email": "test@example.com",
  "created_at": "2025-04-21T12:45:00Z"
}
```

### 2. User Authentication

```
POST /api/v1/users/login
{
  "username": "testuser",
  "password": "Password123!"
}
```

Response:
```json
{
  "token": "jwt-token",
  "refresh_token": "refresh-token",
  "expires_at": "2025-04-22T12:45:00Z"
}
```

### 3. Browsing Problems

```
GET /api/v1/problems
Authorization: Bearer jwt-token
```

Response:
```json
[
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
]
```

### 4. Viewing Problem Details

```
GET /api/v1/problems/1
Authorization: Bearer jwt-token
```

Response:
```json
{
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
}
```

### 5. Submitting a Solution

```
POST /api/v1/submissions
Authorization: Bearer jwt-token
{
  "problem_id": 1,
  "language": "go",
  "code": "package main\n\nfunc twoSum(nums []int, target int) []int {\n    numMap := make(map[int]int)\n    for i, num := range nums {\n        complement := target - num\n        if idx, found := numMap[complement]; found {\n            return []int{idx, i}\n        }\n        numMap[num] = i\n    }\n    return nil\n}"
}
```

Response:
```json
{
  "id": 1,
  "problem_id": 1,
  "user_id": 1,
  "language": "go",
  "status": "Pending",
  "created_at": "2025-04-21T12:50:00Z"
}
```

### 6. Checking Submission Results

```
GET /api/v1/submissions/1
Authorization: Bearer jwt-token
```

Response:
```json
{
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
}
```

### 7. Viewing Submission History

```
GET /api/v1/users/me/submissions
Authorization: Bearer jwt-token
```

Response:
```json
[
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
]
```

## Behind the Scenes: System Workflow

When a user submits a solution, the following happens behind the scenes:

1. **API Gateway** receives the submission request and forwards it to the Submission Service
2. **Submission Service** stores the submission in the PostgreSQL database
3. **Submission Service** publishes a message to Kafka on the `submission-events` topic
4. **Judging Service** consumes the message from Kafka
5. **Judging Service** executes the code in a secure container with resource limits
6. **Judging Service** compares the output with the expected results for each test case
7. **Judging Service** publishes the results back to Kafka
8. **Submission Service** updates the submission status in the database
9. **Notification Service** consumes the event and notifies the user of the results

This event-driven architecture allows for asynchronous processing and scalability, as each service can be scaled independently based on load.

## Testing with Real API Endpoints

When the system is fully implemented, you can test the real API endpoints using curl or any API client:

```bash
# Set up port forwarding to the API Gateway
kubectl port-forward -n codecourt svc/codecourt-api-gateway 8080:8080

# Register a new user
curl -X POST -H "Content-Type: application/json" \
  -d '{"username": "testuser", "email": "test@example.com", "password": "Password123!"}' \
  http://localhost:8080/api/v1/users/register

# Login to get a JWT token
curl -X POST -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "Password123!"}' \
  http://localhost:8080/api/v1/users/login
```

## Troubleshooting

If you encounter issues during end-to-end testing:

1. **Check pod status**:
   ```bash
   kubectl get pods -n codecourt
   ```

2. **View logs for specific services**:
   ```bash
   kubectl logs -n codecourt deployment/codecourt-api-gateway
   kubectl logs -n codecourt deployment/codecourt-submission-service
   kubectl logs -n codecourt deployment/codecourt-judging-service
   ```

3. **Check Kafka topics**:
   ```bash
   kubectl get kafkatopics -n codecourt
   ```

4. **Verify secrets are properly created**:
   ```bash
   kubectl get secrets -n codecourt
   ```

5. **Restart a service if needed**:
   ```bash
   kubectl rollout restart deployment/codecourt-api-gateway -n codecourt
   ```

## Conclusion

The end-to-end testing and simulation capabilities demonstrate the complete workflow of the CodeCourt system, from user registration to submission evaluation. These tools help validate that all components are working together correctly and provide a reference for expected API interactions.
