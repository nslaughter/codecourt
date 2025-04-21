# Distributed Tracing in CodeCourt

This document describes the distributed tracing setup in the CodeCourt system using OpenTelemetry and Jaeger.

## Architecture

The distributed tracing architecture in CodeCourt consists of the following components:

1. **OpenTelemetry Instrumentation**: Each service is instrumented with OpenTelemetry to generate traces.
2. **OpenTelemetry Collector**: Collects traces from all services and forwards them to Jaeger.
3. **Jaeger**: Stores and visualizes the traces.

```ascii
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Service   │     │    OTEL     │     │   Jaeger    │
│             │────▶│  Collector  │────▶│             │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Components

### OpenTelemetry Collector

The OpenTelemetry Collector is deployed as a separate service in the Kubernetes cluster. It receives traces from all services and forwards them to Jaeger.

Configuration:

- **Receivers**: OTLP (gRPC and HTTP)
- **Processors**: Batch, Memory Limiter, Resource
- **Exporters**: Jaeger, Prometheus, Logging

### Jaeger

Jaeger is deployed as an all-in-one service in the Kubernetes cluster. It receives traces from the OpenTelemetry Collector and provides a UI for visualization.

- **Storage**: In-memory (default), can be configured to use Elasticsearch for production
- **UI**: Available at `http://jaeger.codecourt.local` (when configured with ingress)

## Service Instrumentation

Each service in the CodeCourt system is instrumented with OpenTelemetry. The instrumentation is configured using environment variables:

```yaml
- name: OTEL_SERVICE_NAME
  value: "service-name"
- name: OTEL_RESOURCE_ATTRIBUTES
  value: "service.namespace=codecourt,service.name=service-name"
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: "http://codecourt-otel-collector:4317"
- name: OTEL_TRACES_SAMPLER
  value: "parentbased_traceidratio"
- name: OTEL_TRACES_SAMPLER_ARG
  value: "1.0"
- name: OTEL_PROPAGATORS
  value: "tracecontext,baggage,b3"
```

## Go Service Implementation

To implement OpenTelemetry in your Go services, add the following dependencies to your `go.mod` file:

```go
go.opentelemetry.io/otel v1.19.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0
go.opentelemetry.io/otel/sdk v1.19.0
go.opentelemetry.io/otel/trace v1.19.0
```

### Initialization Code

Add the following code to initialize OpenTelemetry in your service:

```go
package tracing

import (
    "context"
    "log"
    "time"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// InitTracer initializes the OpenTelemetry tracer
func InitTracer(serviceName string, otelEndpoint string) func() {
    ctx := context.Background()

    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(serviceName),
        ),
    )
    if err != nil {
        log.Fatalf("Failed to create resource: %v", err)
    }

    // Set up a connection to the collector
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    conn, err := grpc.DialContext(ctx, otelEndpoint, 
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
    )
    if err != nil {
        log.Fatalf("Failed to create gRPC connection to collector: %v", err)
    }

    // Set up a trace exporter
    traceExporter, err := otlptrace.New(ctx, 
        otlptracegrpc.NewClient(
            otlptracegrpc.WithGRPCConn(conn),
        ),
    )
    if err != nil {
        log.Fatalf("Failed to create trace exporter: %v", err)
    }

    // Register the trace exporter with a TracerProvider
    bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
    tracerProvider := sdktrace.NewTracerProvider(
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
        sdktrace.WithResource(res),
        sdktrace.WithSpanProcessor(bsp),
    )
    otel.SetTracerProvider(tracerProvider)

    // Set global propagator to tracecontext (the default is no-op)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return func() {
        // Shutdown will flush any remaining spans and shut down the exporter
        if err := tracerProvider.Shutdown(ctx); err != nil {
            log.Fatalf("Failed to shutdown TracerProvider: %v", err)
        }
    }
}
```

### Usage in Main Function

In your main function, initialize the tracer:

```go
package main

import (
    "context"
    "os"

    "github.com/nslaughter/codecourt/pkg/tracing"
)

func main() {
    serviceName := os.Getenv("OTEL_SERVICE_NAME")
    otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

    // Initialize tracer
    cleanup := tracing.InitTracer(serviceName, otelEndpoint)
    defer cleanup()

    // Rest of your application code
}
```

### Creating Spans

To create spans in your code:

```go
import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

func someFunction(ctx context.Context) {
    // Get a tracer
    tracer := otel.Tracer("github.com/nslaughter/codecourt")

    // Create a span
    ctx, span := tracer.Start(ctx, "someFunction")
    defer span.End()

    // Add attributes to the span
    span.SetAttributes(attribute.String("key", "value"))

    // Create child spans for sub-operations
    ctx, childSpan := tracer.Start(ctx, "childOperation")
    // Do something
    childSpan.End()
}
```

## Viewing Traces

To view traces, access the Jaeger UI at `http://jaeger.codecourt.local` (or the appropriate URL based on your ingress configuration).

## Production Considerations

For production environments, consider the following:

1. **Storage**: Use Elasticsearch or Cassandra for persistent storage instead of in-memory storage.
2. **Sampling**: Adjust sampling rates based on traffic volume.
3. **Resource Allocation**: Allocate appropriate resources to Jaeger and OpenTelemetry Collector.
4. **Security**: Implement proper authentication and authorization for Jaeger UI.
5. **Backup**: Implement a backup strategy for trace data if needed.

## Integration with Existing Monitoring

The distributed tracing setup integrates with the existing monitoring infrastructure:

1. **Prometheus**: The OpenTelemetry Collector exposes metrics that can be scraped by Prometheus.
2. **Grafana**: Create dashboards to visualize trace metrics alongside other system metrics.
