#!/bin/bash
set -eo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CLUSTER_NAME="codecourt"
NAMESPACE="codecourt"
TIMEOUT="300s"

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

# Function to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Function to check prerequisites
check_prerequisites() {
  log_info "Checking prerequisites..."
  
  # Check if kind is installed
  if ! command_exists kind; then
    log_error "kind is not installed. Please install it first: https://kind.sigs.k8s.io/docs/user/quick-start/"
    exit 1
  fi
  
  # Check if kubectl is installed
  if ! command_exists kubectl; then
    log_error "kubectl is not installed. Please install it first: https://kubernetes.io/docs/tasks/tools/install-kubectl/"
    exit 1
  fi
  
  # Check if helm is installed
  if ! command_exists helm; then
    log_error "helm is not installed. Please install it first: https://helm.sh/docs/intro/install/"
    exit 1
  fi
  
  log_success "All prerequisites are satisfied."
}

# Function to create a kind cluster
create_cluster() {
  log_info "Creating Kind cluster '${CLUSTER_NAME}'..."
  
  if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    log_warning "Cluster '${CLUSTER_NAME}' already exists. Deleting it..."
    kind delete cluster --name "${CLUSTER_NAME}"
  fi
  
  kind create cluster --config "${SCRIPT_DIR}/kind-config.yaml"
  log_success "Kind cluster '${CLUSTER_NAME}' created."
}

# Function to install NGINX Ingress Controller
install_nginx_ingress() {
  log_info "Installing NGINX Ingress Controller..."
  
  kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
  
  # Wait for ingress controller to be ready
  kubectl wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=${TIMEOUT}
  
  log_success "NGINX Ingress Controller installed."
}

# Function to create namespace
create_namespace() {
  log_info "Creating namespace '${NAMESPACE}'..."
  
  kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
  
  log_success "Namespace '${NAMESPACE}' created."
}

# Function to add Helm repositories
add_helm_repos() {
  log_info "Adding Helm repositories..."
  
  # Update Helm repositories
  helm repo update
  
  log_success "Helm repositories updated."
}

# Function to install CodeCourt Helm chart
install_codecourt() {
  log_info "Installing CodeCourt Helm chart..."
  
  # First, update dependencies to make sure we have the latest versions
  log_info "Updating Helm chart dependencies..."
  helm dependency update "${PROJECT_ROOT}/helm/codecourt"
  
  # Create service account for the application
  log_info "Creating service account..."
  kubectl create serviceaccount codecourt -n "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
  
  # Install the chart with operators enabled
  log_info "Installing Helm chart..."
  helm install codecourt "${PROJECT_ROOT}/helm/codecourt" \
    --namespace "${NAMESPACE}" \
    --create-namespace \
    --set global.storageClass=standard \
    --set postgresql.enabled=true \
    --set kafka.enabled=true
  
  # Wait for operators to be ready before proceeding
  log_info "Waiting for PostgreSQL Operator to be ready..."
  kubectl wait --namespace "${NAMESPACE}" \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/name=postgres-operator \
    --timeout=${TIMEOUT} || true
  
  log_info "Waiting for Strimzi Kafka Operator to be ready..."
  kubectl wait --namespace "${NAMESPACE}" \
    --for=condition=ready pod \
    --selector=name=strimzi-cluster-operator \
    --timeout=${TIMEOUT} || true
  
  log_success "CodeCourt Helm chart installed with operators."
}

# Function to wait for all deployments to be ready
wait_for_deployments() {
  log_info "Waiting for all deployments to be ready..."
  
  kubectl wait --namespace "${NAMESPACE}" \
    --for=condition=available deployment \
    --selector=app.kubernetes.io/instance=codecourt \
    --timeout=${TIMEOUT}
  
  log_success "All deployments are ready."
}

# Function to run tests
run_tests() {
  log_info "Running end-to-end tests..."
  
  # Basic infrastructure tests
  log_info "Running infrastructure tests..."
  
  # Test 1: Check if all pods are running
  log_info "Test 1: Checking if all pods are running..."
  if kubectl get pods -n "${NAMESPACE}" | grep -v Running | grep -v Completed | grep -v NAME; then
    log_error "Test 1 failed: Not all pods are running."
    return 1
  fi
  log_success "Test 1 passed: All pods are running."
  
  # Test 2: Check if API Gateway is accessible
  log_info "Test 2: Checking if API Gateway is accessible..."
  if ! kubectl run -n "${NAMESPACE}" curl --image=curlimages/curl --restart=Never --rm -it --command -- curl -s http://codecourt-api-gateway:8080/health; then
    log_error "Test 2 failed: API Gateway is not accessible."
    return 1
  fi
  log_success "Test 2 passed: API Gateway is accessible."
  
  # Test 3: Check if Kafka topics are created
  log_info "Test 3: Checking if Kafka topics are created..."
  if ! kubectl get kafkatopics -n "${NAMESPACE}" | grep -q "user-events"; then
    log_error "Test 3 failed: Kafka topics are not created."
    return 1
  fi
  log_success "Test 3 passed: Kafka topics are created."
  
  # Test 4: Check if PostgreSQL cluster is running
  log_info "Test 4: Checking if PostgreSQL cluster is running..."
  if ! kubectl get postgresql -n "${NAMESPACE}" | grep -q "codecourt"; then
    log_error "Test 4 failed: PostgreSQL cluster is not running."
    return 1
  fi
  log_success "Test 4 passed: PostgreSQL cluster is running."
  
  # Run specialized test suites
  log_info "Running specialized test suites..."
  
  # API tests
  if [ -f "${SCRIPT_DIR}/e2e-tests/api-tests.sh" ]; then
    log_info "Running API tests..."
    if ! "${SCRIPT_DIR}/e2e-tests/api-tests.sh"; then
      log_error "API tests failed."
      return 1
    fi
    log_success "API tests passed."
  else
    log_warning "API tests script not found. Skipping API tests."
  fi
  
  # Kafka tests
  if [ -f "${SCRIPT_DIR}/e2e-tests/kafka-tests.sh" ]; then
    log_info "Running Kafka tests..."
    if ! "${SCRIPT_DIR}/e2e-tests/kafka-tests.sh"; then
      log_error "Kafka tests failed."
      return 1
    fi
    log_success "Kafka tests passed."
  else
    log_warning "Kafka tests script not found. Skipping Kafka tests."
  fi
  
  # Notification tests
  if [ -f "${SCRIPT_DIR}/e2e-tests/notification-tests.sh" ]; then
    log_info "Running Notification tests..."
    if ! "${SCRIPT_DIR}/e2e-tests/notification-tests.sh"; then
      log_error "Notification tests failed."
      return 1
    fi
    log_success "Notification tests passed."
  else
    log_warning "Notification tests script not found. Skipping Notification tests."
  fi
  
  log_success "All tests passed!"
  return 0
}

# Function to clean up resources
cleanup() {
  if [ "$1" == "--no-cleanup" ]; then
    log_info "Skipping cleanup as requested."
    return
  fi
  
  log_info "Cleaning up resources..."
  
  # Delete the Kind cluster
  kind delete cluster --name "${CLUSTER_NAME}"
  
  log_success "Cleanup completed."
}

# Main function
main() {
  log_info "Starting CodeCourt end-to-end test..."
  
  check_prerequisites
  create_cluster
  install_nginx_ingress
  create_namespace
  add_helm_repos
  install_codecourt
  wait_for_deployments
  
  if run_tests; then
    log_success "End-to-end test completed successfully."
    cleanup "$1"
    exit 0
  else
    log_error "End-to-end test failed."
    cleanup "$1"
    exit 1
  fi
}

# Handle script arguments
if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
  echo "Usage: $0 [--no-cleanup]"
  echo ""
  echo "Options:"
  echo "  --no-cleanup  Do not clean up resources after the test"
  echo "  --help, -h    Show this help message"
  exit 0
fi

# Run the main function
main "$1"
