# CodeCourt Metrics Package

This package provides standardized Prometheus metrics instrumentation for all CodeCourt services. It implements the RED method (Rate, Errors, Duration) for monitoring service health and includes service-specific metrics for each component of the system.

## Overview

The metrics package offers:

- Standardized HTTP metrics collection via middleware
- Service-specific metrics for each CodeCourt service
- Helper functions for recording database and messaging operations
- Consistent naming and labeling conventions
- Prometheus integration with predefined metric types

## Usage

### Basic Setup

To add metrics to a service, include these steps in your `main.go`:

```go
import (
    "net/http"
    "github.com/nslaughter/codecourt/pkg/metrics"
)

func main() {
    // Create router
    mux := http.NewServeMux()
    
    // Register API routes
    // ...
    
    // Register service info metrics
    metrics.RegisterServiceInfo("service-name", "1.0.0", "2025-04-21", "abc123")
    
    // Set up metrics endpoint
    metrics.SetupMetricsEndpoint(mux)
    
    // Apply metrics middleware
    handler := metrics.MetricsMiddleware("service-name")(mux)
    
    // Start server
    http.ListenAndServe(":8080", handler)
}
```

### Recording Database Operations

```go
// Record a database operation
metrics.RecordDatabaseOperation("service-name", "SELECT", "users", "success")

// Measure database operation duration
startTime := time.Now()
// ... perform database operation
duration := time.Since(startTime).Seconds()
metrics.ObserveDatabaseOperationDuration("service-name", "SELECT", "users", duration)
```

### Recording Kafka Operations

```go
// Record a Kafka message
metrics.RecordKafkaMessage("service-name", "topic-name", "produce")

// Measure Kafka operation duration
startTime := time.Now()
// ... perform Kafka operation
duration := time.Since(startTime).Seconds()
metrics.ObserveKafkaOperationDuration("service-name", "topic-name", "produce", duration)
```

## Service-Specific Metrics

### User Service

```go
// Record authentication attempt
metrics.RecordAuthentication("password", "success")

// Record user session duration
metrics.ObserveUserSessionDuration(3600) // 1 hour session
```

### Problem Service

```go
// Record problem access
metrics.RecordProblemAccess("problem-123", "medium")

// Update problem success rate
metrics.SetProblemSuccessRate("problem-123", "medium", 75.5) // 75.5% success rate
```

### Submission Service

```go
// Record submission
metrics.RecordSubmission("go", "pending", "problem-123", "user-456")

// Observe submission processing time
metrics.ObserveSubmissionProcessingTime("go", "problem-123", 1.5) // 1.5 seconds
```

### Judging Service

```go
// Record judging operation
metrics.RecordJudgingOperation("go", "accepted", "problem-123")

// Record test case result
metrics.RecordTestCaseResult("problem-123", "passed")

// Observe code execution memory usage
metrics.ObserveCodeExecutionMemoryUsage("go", "problem-123", 10485760) // 10MB
```

### Notification Service

```go
// Record notification sent
metrics.RecordNotificationSent("email", "submission_result", "success")

// Record event processing
metrics.RecordEventProcessing("submission_completed", "success")
```

## Available Metrics

The package provides the following standard metrics:

- `codecourt_http_requests_total` - Counter for HTTP requests
- `codecourt_http_request_duration_seconds` - Histogram for HTTP request duration
- `codecourt_database_operations_total` - Counter for database operations
- `codecourt_database_operation_duration_seconds` - Histogram for database operation duration
- `codecourt_kafka_messages_total` - Counter for Kafka messages
- `codecourt_kafka_operation_duration_seconds` - Histogram for Kafka operation duration
- `codecourt_service_info` - Gauge for service version information

Plus service-specific metrics for each component of the system.

## Testing

The metrics package includes comprehensive tests following Go's table-driven testing style. Run the tests with:

```bash
go test -v ./pkg/metrics
```

## Integration with Prometheus

These metrics are designed to be scraped by Prometheus via the `/metrics` endpoint that is automatically set up by the `SetupMetricsEndpoint` function. The metrics follow Prometheus naming conventions and are properly labeled for effective querying and alerting.

## Best Practices

1. **Use Consistent Service Names**: Always use the same service name for all metrics from a given service
2. **Add Context with Labels**: Use labels to add context to metrics, but avoid high cardinality
3. **Measure Critical Paths**: Focus on instrumenting critical code paths first
4. **Keep Histograms Focused**: Use appropriate bucket sizes for histograms based on expected values
5. **Test Metrics**: Verify metrics are correctly registered and exposed
