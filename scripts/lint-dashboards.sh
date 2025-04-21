#!/bin/bash
set -e

# Script to lint Grafana dashboards using dashboard-linter
# https://github.com/grafana/dashboard-linter

# Check if dashboard-linter is installed
if ! command -v dashboard-linter &> /dev/null; then
    echo "Installing dashboard-linter..."
    go install github.com/grafana/dashboard-linter@latest
fi

# Directory containing dashboard files
DASHBOARD_DIR="helm/codecourt/dashboards"

# Create a temporary directory for config
TMP_DIR=$(mktemp -d)
CONFIG_FILE="$TMP_DIR/linter-config.json"

# Create a basic configuration file
cat > "$CONFIG_FILE" << EOF
{
  "rules": {
    "grid-position": "error",
    "panel-title": "error",
    "panel-units": "warning",
    "panel-description": "warning",
    "short-panel-description": "warning",
    "require-panel-description": "warning",
    "panel-datasource": "off",
    "template-datasource": "off",
    "datasource-version": "warning",
    "template-instance-name": "off",
    "template-job-name": "off",
    "template-name": "error",
    "no-duplicate-panels": "error",
    "no-duplicate-targets": "error",
    "no-hidden-panels": "warning",
    "no-disabled-panels": "warning",
    "no-empty-panels": "error"
  }
}
EOF

echo "Linting Grafana dashboards in $DASHBOARD_DIR..."
EXIT_CODE=0

# Find all JSON files in the dashboard directory
for dashboard in $(find "$DASHBOARD_DIR" -name "*.json"); do
    echo "Linting $dashboard..."
    dashboard-linter lint -c "$CONFIG_FILE" "$dashboard"
    LINT_RESULT=$?
    if [ $LINT_RESULT -ne 0 ]; then
        echo "Dashboard linting failed with exit code $LINT_RESULT"
        EXIT_CODE=1
    fi
done

# Clean up
rm -rf "$TMP_DIR"

if [ $EXIT_CODE -eq 0 ]; then
    echo "All dashboards passed linting!"
else
    echo "Some dashboards failed linting. Please fix the issues above."
fi

exit $EXIT_CODE
