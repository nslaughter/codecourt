# CodeCourt Development Guide

This document provides detailed instructions for setting up a development environment and contributing to the CodeCourt project.

## Development Environment Setup

### Prerequisites

Before you begin, ensure you have the following tools installed:

- **Go 1.21+**: Required for building and testing the services
- **Docker**: Used for containerization and local development
- **Docker Compose**: Used for running dependent services locally
- **Kubernetes**: A local cluster for development (Kind recommended)
- **Helm 3.x**: For deploying the application to Kubernetes
- **Make**: Used for running development commands
- **golangci-lint**: For code quality checks

### Initial Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/nslaughter/codecourt.git
   cd codecourt
   ```

2. **Install Go dependencies**:
   ```bash
   make deps
   ```

3. **Set up local development database**:
   ```bash
   docker-compose up -d postgresql kafka
   ```

4. **Run the tests to verify your setup**:
   ```bash
   make test
   ```

## Project Structure

The project follows a standard Go project layout with microservices:

```
codecourt/
├── cmd/                # Main applications for each service
│   ├── api-gateway/
│   ├── judging-service/
│   ├── notification-service/
│   ├── problem-service/
│   ├── submission-service/
│   └── user-service/
├── internal/           # Private application and library code
│   ├── api/            # API definitions
│   ├── auth/           # Authentication utilities
│   ├── config/         # Configuration handling
│   ├── database/       # Database utilities
│   ├── kafka/          # Kafka utilities
│   ├── models/         # Data models
│   └── services/       # Service implementations
├── pkg/                # Public libraries that can be used by external applications
├── scripts/            # Scripts for development, CI/CD, etc.
├── helm/               # Helm charts for Kubernetes deployment
├── docs/               # Documentation
└── Makefile            # Development commands
```

## Development Workflow

### Running Services Locally

Each service can be run independently for development:

```bash
# Run the API Gateway
go run ./cmd/api-gateway

# Run the User Service
go run ./cmd/user-service

# Run other services similarly
```

For a complete local environment, use Docker Compose:

```bash
docker-compose up
```

### Testing

CodeCourt follows test-driven development practices with comprehensive test coverage:

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run end-to-end tests (requires Kubernetes)
make e2e-test
```

#### Test Types

1. **Unit Tests**: Test individual functions and methods in isolation
2. **Integration Tests**: Test interactions between components
3. **End-to-End Tests**: Test the complete system in a Kubernetes environment

All Go tests follow the table-driven testing pattern as per Go best practices.

### Code Quality

Maintain code quality with linting and formatting:

```bash
# Run linter
make lint

# Format code
make fmt

# Vet code for potential issues
make vet
```

## Kubernetes Development

### Setting Up a Local Cluster

We recommend using Kind for local Kubernetes development:

```bash
# Create a Kind cluster
make kind-create

# Set up the development environment
make e2e-setup

# Run end-to-end tests
make e2e-test

# Clean up
make e2e-teardown
```

### Deploying to Kubernetes

Deploy the application to your Kubernetes cluster:

```bash
# Install the Helm chart
make helm-install

# Upgrade an existing installation
make helm-upgrade

# Uninstall the application
make helm-uninstall
```

## Working with Helm Charts

The Helm charts are located in the `helm/` directory:

```bash
# Update Helm dependencies
make helm-deps

# Lint the Helm chart
make helm-lint

# Generate templates for inspection
make helm-template
```

## Database Migrations

Each service manages its own database schema using migrations:

```bash
# Run migrations for a specific service
go run ./cmd/user-service migrate up
```

## Kafka Event Management

Services communicate through Kafka events. The event schemas are defined in the `internal/api` package.

To monitor Kafka topics during development:

```bash
# List Kafka topics
docker-compose exec kafka kafka-topics.sh --list --bootstrap-server localhost:9092

# Consume messages from a topic
docker-compose exec kafka kafka-console-consumer.sh --topic user-events --bootstrap-server localhost:9092
```

## AI-Assisted Development

CodeCourt development leverages AI tools integrated with the IDE (Windsurf) to enhance productivity:

- **Code Generation**: AI assists in generating boilerplate code and implementing patterns
- **Documentation**: AI helps create and maintain comprehensive documentation
- **Testing**: AI suggests test cases and helps implement table-driven tests
- **Refactoring**: AI identifies improvement opportunities and assists with refactoring

While AI tools accelerate development, all code undergoes human review to ensure quality and adherence to project standards.

## Debugging

### Local Debugging

For local debugging, you can use standard Go debugging tools:

```bash
# Run a service with verbose logging
go run ./cmd/user-service -v

# Use Delve for debugging
dlv debug ./cmd/user-service
```

### Kubernetes Debugging

For debugging in Kubernetes:

```bash
# View logs for a specific pod
kubectl logs -f -n codecourt <pod-name>

# Port-forward to access a service locally
kubectl port-forward -n codecourt svc/codecourt-api-gateway 8080:8080

# Get a shell in a running container
kubectl exec -it -n codecourt <pod-name> -- /bin/sh
```

## Common Issues and Solutions

### Database Connection Issues

If you encounter database connection issues:

1. Ensure PostgreSQL is running: `docker-compose ps`
2. Check connection parameters in your configuration
3. Verify network connectivity: `telnet localhost 5432`

### Kafka Connection Issues

If services can't connect to Kafka:

1. Ensure Kafka is running: `docker-compose ps`
2. Check Kafka logs: `docker-compose logs kafka`
3. Verify topics exist: `docker-compose exec kafka kafka-topics.sh --list --bootstrap-server localhost:9092`

### Kubernetes Deployment Issues

If the Helm deployment fails:

1. Check the Helm release: `helm list -n codecourt`
2. View Kubernetes events: `kubectl get events -n codecourt`
3. Check pod status: `kubectl get pods -n codecourt`
4. View pod logs: `kubectl logs -n codecourt <pod-name>`

## Conclusion

This development guide should help you get started with CodeCourt development. If you encounter any issues not covered here, please check the project issues or create a new one.
