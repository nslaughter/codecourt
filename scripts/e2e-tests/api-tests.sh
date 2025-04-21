#!/bin/bash
set -eo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script variables
NAMESPACE="codecourt"
API_GATEWAY_SERVICE="codecourt-api-gateway"
API_GATEWAY_PORT="8080"

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

# Function to run a test
run_test() {
  local test_name="$1"
  local endpoint="$2"
  local method="${3:-GET}"
  local expected_status="${4:-200}"
  local payload="$5"
  
  log_info "Running test: $test_name"
  
  local curl_cmd="curl -s -o /dev/null -w '%{http_code}' -X $method http://$API_GATEWAY_SERVICE:$API_GATEWAY_PORT$endpoint"
  
  if [ -n "$payload" ]; then
    curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$payload'"
  fi
  
  local status_code=$(kubectl run -n "${NAMESPACE}" curl-test --image=curlimages/curl --restart=Never --rm -it --command -- bash -c "$curl_cmd")
  
  if [ "$status_code" == "$expected_status" ]; then
    log_success "Test passed: $test_name (Status code: $status_code)"
    return 0
  else
    log_error "Test failed: $test_name (Expected: $expected_status, Got: $status_code)"
    return 1
  fi
}

# Main function
main() {
  log_info "Starting API tests..."
  
  # Test 1: Health check
  run_test "Health check" "/health"
  
  # Test 2: Register a new user
  run_test "Register user" "/api/v1/users/register" "POST" "201" '{"username":"testuser","email":"test@example.com","password":"Password123!","firstName":"Test","lastName":"User"}'
  
  # Test 3: Login with the registered user
  run_test "Login user" "/api/v1/users/login" "POST" "200" '{"email":"test@example.com","password":"Password123!"}'
  
  # Test 4: Create a problem (requires authentication)
  # This test will likely fail without a valid JWT token, but we include it for completeness
  run_test "Create problem" "/api/v1/problems" "POST" "401" '{"title":"Test Problem","description":"This is a test problem","difficulty":"medium","timeLimit":1000,"memoryLimit":256,"tags":["test","example"]}'
  
  log_success "API tests completed."
}

# Run the main function
main
