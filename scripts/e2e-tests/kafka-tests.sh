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
KAFKA_CLUSTER="codecourt-kafka"

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

# Function to check if Kafka topics are created
check_kafka_topics() {
  log_info "Checking Kafka topics..."
  
  local topics=("user-events" "submission-events" "judging-events" "problem-events")
  local all_topics_exist=true
  
  for topic in "${topics[@]}"; do
    if ! kubectl get kafkatopics -n "${NAMESPACE}" | grep -q "$topic"; then
      log_error "Topic '$topic' does not exist."
      all_topics_exist=false
    else
      log_success "Topic '$topic' exists."
    fi
  done
  
  if [ "$all_topics_exist" = true ]; then
    log_success "All Kafka topics exist."
    return 0
  else
    log_error "Some Kafka topics are missing."
    return 1
  fi
}

# Function to test Kafka producer/consumer
test_kafka_messaging() {
  log_info "Testing Kafka messaging..."
  
  # Create a test message
  local test_message="Test message $(date +%s)"
  local topic="user-events"
  
  # Create a temporary pod for Kafka testing
  cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: kafka-test-client
  namespace: ${NAMESPACE}
spec:
  containers:
  - name: kafka-test-client
    image: quay.io/strimzi/kafka:latest-kafka-3.3.1
    command:
    - sleep
    - "3600"
EOF
  
  # Wait for the pod to be ready
  kubectl wait --namespace "${NAMESPACE}" \
    --for=condition=ready pod \
    --selector=app=kafka-test-client \
    --timeout=60s
  
  # Produce a message
  log_info "Producing a test message to topic '$topic'..."
  kubectl exec -n "${NAMESPACE}" kafka-test-client -- \
    bin/kafka-console-producer.sh \
    --bootstrap-server ${KAFKA_CLUSTER}-kafka-bootstrap:9092 \
    --topic ${topic} \
    --property "parse.key=true" \
    --property "key.separator=:" <<< "test-key:${test_message}"
  
  # Consume the message
  log_info "Consuming messages from topic '$topic'..."
  local consumed_message=$(kubectl exec -n "${NAMESPACE}" kafka-test-client -- \
    bin/kafka-console-consumer.sh \
    --bootstrap-server ${KAFKA_CLUSTER}-kafka-bootstrap:9092 \
    --topic ${topic} \
    --from-beginning \
    --max-messages 1 \
    --property "print.key=true" \
    --property "key.separator=:" \
    --timeout-ms 10000)
  
  # Clean up
  kubectl delete pod -n "${NAMESPACE}" kafka-test-client
  
  # Check if the message was consumed
  if echo "$consumed_message" | grep -q "$test_message"; then
    log_success "Kafka messaging test passed. Message was produced and consumed successfully."
    return 0
  else
    log_error "Kafka messaging test failed. Message was not consumed correctly."
    return 1
  fi
}

# Main function
main() {
  log_info "Starting Kafka tests..."
  
  # Test 1: Check if Kafka topics are created
  check_kafka_topics
  
  # Test 2: Test Kafka messaging
  test_kafka_messaging
  
  log_success "Kafka tests completed."
}

# Run the main function
main
