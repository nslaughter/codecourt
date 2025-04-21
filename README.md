# CodeCourt

A modern coding judge system that runs in Kubernetes. CodeCourt is written in Go, uses Kafka for messaging, and PostgreSQL for persistence. It provides a platform for hosting coding competitions, practice sessions, and educational exercises.

> **⚠️ WARNING**
>
> This repository was developed primarily with the objective of getting quick results with AI while a human guided towards architectural ideas, project structure, and testing methods. As a result, the repository has a long way to go in terms of human validation and testing with realistic flows. It should not be considered production-ready without thorough review and additional testing.

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Kubernetes (or Kind for local development)
- Helm 3.x
- Make

### Local Development Setup

```bash
# Clone the repository
git clone https://github.com/nslaughter/codecourt.git
cd codecourt

# Install dependencies
make deps

# Run tests
make test

# Set up local Kubernetes environment with Kind
make e2e-setup

# Run end-to-end tests
make e2e-test

# Clean up
make e2e-teardown
```

For more detailed setup instructions, see [DEVELOPMENT.md](docs/DEVELOPMENT.md).

## Architecture Overview

CodeCourt follows a microservices architecture with six core services:

1. **API Gateway Service**: Handles external requests, authentication, and routing
2. **User Service**: Manages user accounts, authentication, and profiles
3. **Problem Service**: Manages coding challenges, test cases, and categories
4. **Submission Service**: Processes code submissions and tracks their lifecycle
5. **Judging Service**: Executes code in a secure sandbox and evaluates results
6. **Notification Service**: Delivers notifications across multiple channels

The architecture leverages:
- **Go** as the primary programming language
- **Kafka** for reliable message queuing and event streaming
- **PostgreSQL** for persistent data storage
- **Kubernetes** for container orchestration
- **Helm** for deployment management

For a detailed architecture explanation, see [ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Features

- **Secure Code Execution**: Run user-submitted code in isolated containers
- **Real-time Feedback**: Immediate results and performance metrics
- **Problem Management**: Create, organize, and share coding challenges
- **User Management**: Authentication, profiles, and progress tracking
- **Competitions**: Host time-limited coding competitions with leaderboards
- **Notifications**: Multi-channel notifications for system events

## Deployment

CodeCourt can be deployed to any Kubernetes cluster using Helm:

```bash
# Add the CodeCourt Helm repository
helm repo add codecourt https://nslaughter.github.io/codecourt/charts

# Install CodeCourt
helm install codecourt codecourt/codecourt --namespace codecourt --create-namespace
```

For production deployment considerations, see [DEPLOYMENT.md](docs/DEPLOYMENT.md).

## Development Approach

CodeCourt has been developed using a combination of traditional software engineering practices and AI-assisted development:

- **LLM Integration**: Development leveraged AI tools integrated with the IDE (Windsurf) to accelerate development, ensure best practices, and maintain consistency across the codebase
- **Test-Driven Development**: Comprehensive test coverage with unit, integration, and end-to-end tests
- **Conventional Commits**: Structured commit messages for better change tracking
- **Microservices Architecture**: Independent services with clear boundaries and responsibilities

This approach has enabled rapid development while maintaining high code quality and adherence to Go best practices.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## License

CodeCourt is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
