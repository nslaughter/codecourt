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

# Function to create secrets
create_secrets() {
  log_info "Creating secrets for CodeCourt services..."
  
  # API Gateway secrets
  log_info "Creating API Gateway secrets..."
  kubectl create secret generic codecourt-api-gateway-secrets \
    --namespace "${NAMESPACE}" \
    --from-literal=JWT_SECRET=test-jwt-secret \
    --from-literal=JWT_EXPIRY=24h \
    --from-literal=REFRESH_EXPIRY=7d \
    --dry-run=client -o yaml | kubectl apply -f -
  
  # User Service secrets
  log_info "Creating User Service secrets..."
  kubectl create secret generic codecourt-user-service-secrets \
    --namespace "${NAMESPACE}" \
    --from-literal=DB_HOST=codecourt \
    --from-literal=DB_PORT=5432 \
    --from-literal=DB_USER=codecourt \
    --from-literal=DB_PASSWORD=codecourt \
    --from-literal=DB_NAME=codecourt \
    --from-literal=DB_SSLMODE=disable \
    --from-literal=KAFKA_BROKERS=codecourt-kafka-bootstrap:9092 \
    --from-literal=KAFKA_GROUP_ID=user-service \
    --from-literal=KAFKA_TOPICS=user-events \
    --dry-run=client -o yaml | kubectl apply -f -
  
  # Problem Service secrets
  log_info "Creating Problem Service secrets..."
  kubectl create secret generic codecourt-problem-service-secrets \
    --namespace "${NAMESPACE}" \
    --from-literal=DB_HOST=codecourt \
    --from-literal=DB_PORT=5432 \
    --from-literal=DB_USER=codecourt \
    --from-literal=DB_PASSWORD=codecourt \
    --from-literal=DB_NAME=codecourt \
    --from-literal=DB_SSLMODE=disable \
    --from-literal=KAFKA_BROKERS=codecourt-kafka-bootstrap:9092 \
    --from-literal=KAFKA_GROUP_ID=problem-service \
    --from-literal=KAFKA_TOPICS=problem-events \
    --dry-run=client -o yaml | kubectl apply -f -
  
  # Submission Service secrets
  log_info "Creating Submission Service secrets..."
  kubectl create secret generic codecourt-submission-service-secrets \
    --namespace "${NAMESPACE}" \
    --from-literal=DB_HOST=codecourt \
    --from-literal=DB_PORT=5432 \
    --from-literal=DB_USER=codecourt \
    --from-literal=DB_PASSWORD=codecourt \
    --from-literal=DB_NAME=codecourt \
    --from-literal=DB_SSLMODE=disable \
    --from-literal=KAFKA_BROKERS=codecourt-kafka-bootstrap:9092 \
    --from-literal=KAFKA_GROUP_ID=submission-service \
    --from-literal=KAFKA_TOPICS=submission-events \
    --dry-run=client -o yaml | kubectl apply -f -
  
  # Judging Service secrets
  log_info "Creating Judging Service secrets..."
  kubectl create secret generic codecourt-judging-service-secrets \
    --namespace "${NAMESPACE}" \
    --from-literal=KAFKA_BROKERS=codecourt-kafka-bootstrap:9092 \
    --from-literal=KAFKA_GROUP_ID=judging-service \
    --from-literal=KAFKA_TOPICS=submission-events \
    --from-literal=MAX_EXECUTION_TIME=10000 \
    --from-literal=MAX_MEMORY_USAGE=512M \
    --dry-run=client -o yaml | kubectl apply -f -
  
  # Notification Service secrets
  log_info "Creating Notification Service secrets..."
  kubectl create secret generic codecourt-notification-service-secrets \
    --namespace "${NAMESPACE}" \
    --from-literal=DB_HOST=codecourt \
    --from-literal=DB_PORT=5432 \
    --from-literal=DB_USER=codecourt \
    --from-literal=DB_PASSWORD=codecourt \
    --from-literal=DB_NAME=codecourt \
    --from-literal=DB_SSLMODE=disable \
    --from-literal=KAFKA_BROKERS=codecourt-kafka-bootstrap:9092 \
    --from-literal=KAFKA_GROUP_ID=notification-service \
    --from-literal=KAFKA_TOPICS=user-events,submission-events,judging-events \
    --from-literal=SMTP_HOST=smtp.example.com \
    --from-literal=SMTP_PORT=587 \
    --from-literal=SMTP_USERNAME=test \
    --from-literal=SMTP_PASSWORD=test \
    --from-literal=SMTP_FROM=noreply@codecourt.io \
    --dry-run=client -o yaml | kubectl apply -f -
  
  log_success "All secrets created successfully!"
}

# Main function
main() {
  create_secrets
}

# Execute main function
main "$@"
