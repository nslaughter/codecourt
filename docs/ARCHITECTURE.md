# CodeCourt Architecture

This document provides a detailed overview of the CodeCourt system architecture, explaining the design decisions, component interactions, and technical implementation details.

## System Overview

CodeCourt is built on a microservices architecture, with each service having a specific responsibility and communicating with other services through well-defined interfaces. This approach enables independent development, deployment, and scaling of each component.

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│   API Gateway   │────▶│  User Service   │     │ Problem Service │
│                 │     │                 │     │                 │
└────────┬────────┘     └─────────────────┘     └─────────────────┘
         │                      ▲                        ▲
         │                      │                        │
         ▼                      │                        │
┌─────────────────┐             │                        │
│                 │             │                        │
│   Kafka Event   │─────────────┴────────────────────────┘
│     Broker      │             ▲                        ▲
│                 │             │                        │
└────────┬────────┘             │                        │
         │                      │                        │
         │                      │                        │
         ▼                      │                        │
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│  Submission     │────▶│ Judging Service │────▶│  Notification   │
│    Service      │     │                 │     │    Service      │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

## Core Services

### 1. API Gateway Service

The API Gateway serves as the entry point for all client requests, handling:

- **Request Routing**: Directs requests to appropriate microservices
- **Authentication**: Validates JWT tokens and enforces access control
- **Request/Response Transformation**: Adapts between client and internal formats
- **Rate Limiting**: Prevents abuse of the system
- **Logging and Monitoring**: Tracks request patterns and system health

**Technical Implementation:**
- Written in Go using the standard library's HTTP package
- JWT-based authentication with access and refresh tokens
- Middleware architecture for cross-cutting concerns

### 2. User Service

The User Service manages all user-related operations:

- **User Registration and Authentication**: Account creation and login
- **Profile Management**: User details and preferences
- **Authorization**: Role-based access control
- **Session Management**: Handling user sessions and tokens

**Technical Implementation:**
- RESTful API built with Go
- PostgreSQL database for user data persistence
- bcrypt for password hashing
- JWT for authentication tokens

### 3. Problem Service

The Problem Service handles coding challenges and their metadata:

- **Problem Management**: CRUD operations for coding problems
- **Test Case Management**: Input/output pairs for problem validation
- **Category and Tag Management**: Organization of problems
- **Difficulty Ratings**: Problem complexity classification

**Technical Implementation:**
- RESTful API built with Go
- PostgreSQL database for problem data
- Structured problem format with markdown support
- Version control for problems and test cases

### 4. Submission Service

The Submission Service processes code submissions:

- **Submission Handling**: Receives and queues code submissions
- **Status Tracking**: Monitors the lifecycle of submissions
- **History Management**: Maintains submission records
- **Event Publishing**: Notifies other services of submission events

**Technical Implementation:**
- RESTful API built with Go
- PostgreSQL for submission metadata storage
- Kafka for publishing submission events
- Idempotent processing to handle potential duplicates

### 5. Judging Service

The Judging Service executes and evaluates submitted code:

- **Secure Execution**: Runs code in isolated containers
- **Test Case Validation**: Compares outputs against expected results
- **Performance Measurement**: Tracks execution time and memory usage
- **Result Reporting**: Provides detailed feedback on submissions

**Technical Implementation:**
- Go service with container orchestration
- Kubernetes for container isolation and resource limits
- Strict security policies to prevent malicious code execution
- Time and memory constraints enforcement

### 6. Notification Service

The Notification Service manages communication with users:

- **Event Subscription**: Listens for system events requiring notifications
- **Multi-channel Delivery**: Supports email, in-app, and other notification methods
- **Templating**: Customizable notification content
- **Delivery Status Tracking**: Monitors notification delivery and read status

**Technical Implementation:**
- Go service with Kafka consumer
- PostgreSQL for notification history and preferences
- Email delivery via SMTP
- In-app notifications via WebSockets

## Data Storage

### PostgreSQL Database

Each service maintains its own PostgreSQL database, following the microservices pattern of decentralized data management. This approach:

- Ensures service independence
- Allows for optimized schema design per service
- Enables independent scaling of databases
- Prevents tight coupling between services

### Database Schema Design

Each service follows these database design principles:

- **Normalized Schema**: Reduces data redundancy
- **Appropriate Indexing**: Optimizes query performance
- **Versioned Migrations**: Manages schema evolution
- **Soft Deletion**: Preserves data history where appropriate

## Messaging and Event Flow

### Kafka Event Broker

Kafka serves as the central event bus for asynchronous communication between services:

- **Event-Driven Architecture**: Services publish events and subscribe to relevant topics
- **Reliable Delivery**: Ensures events are processed at least once
- **Scalable Processing**: Allows for parallel event consumption
- **Event Sourcing**: Enables rebuilding state from event history

### Key Event Flows

1. **Submission Processing Flow**:
   - User submits code via API Gateway
   - Submission Service stores submission and publishes event
   - Judging Service consumes event, executes code, and publishes results
   - Notification Service informs user of results

2. **User Activity Flow**:
   - User Service publishes user events (registration, profile updates)
   - Notification Service sends welcome or confirmation messages
   - Problem Service may update recommendations based on user activity

## Deployment Architecture

### Kubernetes Deployment

CodeCourt is designed to run on Kubernetes, with:

- **Service Pods**: Each microservice runs in dedicated pods
- **Horizontal Pod Autoscaling**: Scales based on load
- **ConfigMaps and Secrets**: Manages configuration and sensitive data
- **Persistent Volumes**: Stores persistent data
- **Ingress Controllers**: Manages external access

### Helm Charts

Deployment is managed through Helm charts:

- **Chart Structure**: Organized by service with shared dependencies
- **Values Configuration**: Customizable deployment parameters
- **Dependency Management**: Handles PostgreSQL and Kafka operators
- **Upgrade Strategy**: Supports rolling updates and rollbacks

## Security Considerations

- **Network Security**: Service-to-service communication over TLS
- **Authentication**: JWT-based with short-lived tokens
- **Authorization**: Role-based access control at service level
- **Code Execution**: Isolated containers with resource limits
- **Data Protection**: Encrypted sensitive data at rest and in transit

## Monitoring and Observability

- **Logging**: Structured logs with correlation IDs
- **Metrics**: Prometheus for system and business metrics
- **Tracing**: Distributed tracing for request flows
- **Alerting**: Proactive notification of system issues

## Conclusion

The CodeCourt architecture is designed for scalability, resilience, and maintainability. By following microservices best practices and leveraging modern cloud-native technologies, the system can handle varying loads while maintaining security and performance.

The clear separation of concerns between services allows for independent development and deployment, making the system adaptable to changing requirements and technologies.
