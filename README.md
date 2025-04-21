# codecourt

A coding judge system that runs in Kubernetes, it's written in Go, uses Kafka for messaging, and PostgreSQL for persistence.

## Architecture Overview

The system follows a microservices architecture with six core services:

1. **API Gateway Service**: Handles external requests, authentication, and routing
2. **Submission Service**: Manages code submissions lifecycle
3. **Judging Service**: Executes and evaluates submissions in a secure sandbox
4. **Problem Service**: Manages coding challenges and test cases
5. **User Service**: Handles user management and profiles
6. **Notification Service**: Manages user notifications

The architecture leverages:
- **Go** as the primary programming language
- **Kafka** for reliable message queuing and event streaming
- **PostgreSQL** for persistent
