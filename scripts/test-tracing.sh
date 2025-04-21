#!/bin/bash
# Test script for verifying the distributed tracing setup in CodeCourt

set -e

echo "Testing distributed tracing setup in CodeCourt..."

# Check if the Helm chart is valid
echo "Checking Helm chart validity..."
helm lint helm/codecourt || { echo "Helm chart validation failed"; exit 1; }
echo "✅ Helm chart is valid"

# Check if the Jaeger deployment is correctly configured
echo "Checking Jaeger deployment configuration..."
if ! grep -A 2 "jaeger:" helm/codecourt/values.yaml | grep -q "enabled: true"; then
  echo "Jaeger is not enabled in values.yaml"
  exit 1
fi
grep -q "name: {{ include \"codecourt.fullname\" . }}-jaeger" helm/codecourt/templates/jaeger-deployment.yaml || { echo "Jaeger deployment template not found"; exit 1; }
echo "✅ Jaeger deployment is correctly configured"

# Check if the OpenTelemetry Collector deployment is correctly configured
echo "Checking OpenTelemetry Collector configuration..."
if ! grep -A 2 "otelCollector:" helm/codecourt/values.yaml | grep -q "enabled: true"; then
  echo "OpenTelemetry Collector is not enabled in values.yaml"
  exit 1
fi
grep -q "name: {{ include \"codecourt.fullname\" . }}-otel-collector" helm/codecourt/templates/otel-collector-deployment.yaml || { echo "OpenTelemetry Collector deployment template not found"; exit 1; }
echo "✅ OpenTelemetry Collector is correctly configured"

# Check if the service monitors for Prometheus are set up properly
echo "Checking service monitors for Prometheus..."
grep -q "ServiceMonitor" helm/codecourt/templates/otel-jaeger-service-monitors.yaml || { echo "Service monitors for OpenTelemetry and Jaeger not found"; exit 1; }
echo "✅ Service monitors for Prometheus are correctly configured"

# Check if the Problem Service is configured to send traces to the OpenTelemetry Collector
echo "Checking Problem Service configuration for tracing..."
grep -q "OTEL_SERVICE_NAME" helm/codecourt/templates/problem-service-deployment.yaml || { echo "OpenTelemetry configuration not found in Problem Service"; exit 1; }
grep -q "OTEL_EXPORTER_OTLP_ENDPOINT" helm/codecourt/templates/problem-service-deployment.yaml || { echo "OpenTelemetry endpoint configuration not found in Problem Service"; exit 1; }
echo "✅ Problem Service is correctly configured for tracing"

echo "All tests passed! The distributed tracing setup is correctly configured."
echo ""
echo "To deploy the application with tracing enabled:"
echo "1. Run: helm install codecourt helm/codecourt"
echo "2. Wait for all pods to be ready"
echo "3. Access the Jaeger UI at http://jaeger.codecourt.local or via port-forwarding:"
echo "   kubectl port-forward svc/codecourt-jaeger 16686:16686"
echo ""
echo "To generate traces, make requests to your services and then view them in the Jaeger UI."
