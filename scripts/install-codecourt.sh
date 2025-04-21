#!/bin/bash
# Script to install CodeCourt with proper CRD handling

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Installing CodeCourt with proper CRD handling..."

# Add required Helm repositories
echo "Adding required Helm repositories..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add strimzi https://strimzi.io/charts/
helm repo add postgres-operator https://opensource.zalando.com/postgres-operator/charts/postgres-operator
helm repo update

# Install Prometheus Operator CRDs using the chart's CRDs
echo "Installing Prometheus Operator CRDs..."

# Create a temporary directory for CRDs
TMP_DIR=$(mktemp -d)
echo "Created temporary directory: $TMP_DIR"

# Extract CRDs from the chart
echo "Extracting CRDs from kube-prometheus-stack chart..."
helm template --include-crds --output-dir "$TMP_DIR" prometheus-operator prometheus-community/kube-prometheus-stack > /dev/null

# Apply the CRDs
echo "Applying CRDs..."
kubectl apply -f "$TMP_DIR/kube-prometheus-stack/charts/crds/crds/" --server-side

# Wait for CRDs to be ready
echo "Waiting for CRDs to be established..."
kubectl wait --for condition=established --timeout=60s crd/prometheuses.monitoring.coreos.com crd/alertmanagers.monitoring.coreos.com crd/servicemonitors.monitoring.coreos.com crd/podmonitors.monitoring.coreos.com crd/prometheusrules.monitoring.coreos.com

# Clean up temporary directory
echo "Cleaning up temporary directory..."
rm -rf "$TMP_DIR"

# Build dependencies
echo "Building Helm dependencies..."
cd "$REPO_ROOT/helm/codecourt"
helm dependency build

# Clean up any existing resources that might conflict
echo "Cleaning up any existing resources that might conflict..."

# Check for existing Helm release
if helm status codecourt &>/dev/null; then
  echo "Found existing 'codecourt' Helm release. Uninstalling..."
  helm uninstall codecourt
  # Wait a moment for resources to be cleaned up
  sleep 5
fi

# Clean up specific cluster-scoped resources that might cause conflicts
echo "Cleaning up specific cluster-scoped resources..."
kubectl delete clusterrole postgres-pod --ignore-not-found
kubectl delete clusterrole codecourt-postgresql-operator --ignore-not-found

# Look for other potential conflicting resources
echo "Looking for other potential conflicting resources..."

# Clean up codecourt-prefixed resources
echo "Cleaning up codecourt-prefixed resources..."
for resource in clusterrole clusterrolebinding; do
  kubectl get ${resource} | grep codecourt | awk '{print $1}' | xargs -r kubectl delete ${resource} --ignore-not-found
done

# Clean up strimzi-related resources
echo "Cleaning up strimzi-related resources..."
for resource in clusterrole clusterrolebinding; do
  kubectl get ${resource} | grep strimzi | awk '{print $1}' | xargs -r kubectl delete ${resource} --ignore-not-found
done

# Install the Helm chart
echo "Installing CodeCourt Helm chart..."
# Try to install with regular approach first
if ! helm install codecourt "$REPO_ROOT/helm/codecourt" "$@"; then
  echo "\nRegular installation failed. Trying with --force flag..."
  # If regular install fails, try with --force
  if ! helm install codecourt "$REPO_ROOT/helm/codecourt" --force "$@"; then
    echo "\nInstallation failed even with --force flag. You might need to manually clean up resources:"
    echo "1. Check for any remaining resources: kubectl get all,pvc,configmap,secret,ingress,serviceaccount -l app.kubernetes.io/instance=codecourt"
    echo "2. Delete any remaining resources manually"
    echo "3. Try installation again: ./scripts/install-codecourt.sh"
    exit 1
  fi
fi

echo "CodeCourt installation complete!"
echo ""
echo "To access the Jaeger UI, run:"
echo "kubectl port-forward svc/codecourt-jaeger 16686:16686"
echo ""
echo "Then open http://localhost:16686 in your browser."
