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
NOTIFICATION_SERVICE="codecourt-notification-service"
NOTIFICATION_PORT="8085"

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
  
  local curl_cmd="curl -s -o /dev/null -w '%{http_code}' -X $method http://$NOTIFICATION_SERVICE:$NOTIFICATION_PORT$endpoint"
  
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

# Function to test notification creation
test_notification_creation() {
  log_info "Testing notification creation..."
  
  # Create a test notification
  local user_id="00000000-0000-0000-0000-000000000001" # Example UUID
  local notification_payload='{
    "userId": "'"$user_id"'",
    "type": "in-app",
    "subject": "Test Notification",
    "content": "This is a test notification created by the e2e test.",
    "priority": "normal"
  }'
  
  run_test "Create notification" "/api/v1/notifications" "POST" "201" "$notification_payload"
}

# Function to test notification retrieval
test_notification_retrieval() {
  log_info "Testing notification retrieval..."
  
  # Get notifications for a user
  local user_id="00000000-0000-0000-0000-000000000001" # Example UUID
  
  run_test "Get user notifications" "/api/v1/notifications/user/$user_id" "GET" "200"
}

# Function to test notification preference
test_notification_preference() {
  log_info "Testing notification preference..."
  
  # Set notification preference for a user
  local user_id="00000000-0000-0000-0000-000000000001" # Example UUID
  local preference_payload='{
    "userId": "'"$user_id"'",
    "eventType": "submission_judged",
    "channels": ["email", "in-app"],
    "enabled": true
  }'
  
  run_test "Set notification preference" "/api/v1/notifications/preferences" "POST" "201" "$preference_payload"
  
  # Get notification preferences for a user
  run_test "Get notification preferences" "/api/v1/notifications/preferences/user/$user_id" "GET" "200"
}

# Function to test notification template
test_notification_template() {
  log_info "Testing notification template..."
  
  # Create a notification template
  local template_payload='{
    "name": "Test Template",
    "eventType": "submission_judged",
    "type": "email",
    "subject": "Your submission has been judged: {{.status}}",
    "content": "Your submission for {{.problem_name}} has been judged as {{.status}}. Execution time: {{.execution_time}}ms"
  }'
  
  run_test "Create notification template" "/api/v1/notifications/templates" "POST" "201" "$template_payload"
  
  # Get notification templates
  run_test "Get notification templates" "/api/v1/notifications/templates" "GET" "200"
}

# Main function
main() {
  log_info "Starting Notification Service tests..."
  
  # Test 1: Health check
  run_test "Health check" "/health"
  
  # Test 2: Test notification creation
  test_notification_creation
  
  # Test 3: Test notification retrieval
  test_notification_retrieval
  
  # Test 4: Test notification preference
  test_notification_preference
  
  # Test 5: Test notification template
  test_notification_template
  
  log_success "Notification Service tests completed."
}

# Run the main function
main
