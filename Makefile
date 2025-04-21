# CodeCourt Makefile
# Provides commands for development, testing, and deployment

# Variables
SHELL := /bin/bash
PROJECT_ROOT := $(shell pwd)
NAMESPACE := codecourt
CLUSTER_NAME := codecourt
HELM_CHART_DIR := $(PROJECT_ROOT)/helm/codecourt
SCRIPTS_DIR := $(PROJECT_ROOT)/scripts
GO_FILES := $(shell find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*")

# Go commands
GO := go
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOMOD := $(GO) mod
GOLINT := golangci-lint

# Kubernetes/Helm commands
KUBECTL := kubectl
HELM := helm
KIND := kind

# Docker commands
DOCKER := docker

# Default target
.PHONY: all
all: lint test

# Development targets
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	gofmt -s -w $(GO_FILES)

.PHONY: vet
vet:
	@echo "Vetting code..."
	$(GO) vet ./...

.PHONY: test
test: test-unit test-integration

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -race -cover ./...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -race -tags=integration ./...

# Build targets
.PHONY: build
build:
	@echo "Building services..."
	$(GOBUILD) -o bin/api-gateway ./cmd/api-gateway
	$(GOBUILD) -o bin/user-service ./cmd/user-service
	$(GOBUILD) -o bin/problem-service ./cmd/problem-service
	$(GOBUILD) -o bin/submission-service ./cmd/submission-service
	$(GOBUILD) -o bin/judging-service ./cmd/judging-service
	$(GOBUILD) -o bin/notification-service ./cmd/notification-service

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker images..."
	$(DOCKER) build -t codecourt/api-gateway:latest -f build/api-gateway/Dockerfile .
	$(DOCKER) build -t codecourt/user-service:latest -f build/user-service/Dockerfile .
	$(DOCKER) build -t codecourt/problem-service:latest -f build/problem-service/Dockerfile .
	$(DOCKER) build -t codecourt/submission-service:latest -f build/submission-service/Dockerfile .
	$(DOCKER) build -t codecourt/judging-service:latest -f build/judging-service/Dockerfile .
	$(DOCKER) build -t codecourt/notification-service:latest -f build/notification-service/Dockerfile .

.PHONY: docker-push
docker-push:
	@echo "Pushing Docker images..."
	$(DOCKER) push codecourt/api-gateway:latest
	$(DOCKER) push codecourt/user-service:latest
	$(DOCKER) push codecourt/problem-service:latest
	$(DOCKER) push codecourt/submission-service:latest
	$(DOCKER) push codecourt/judging-service:latest
	$(DOCKER) push codecourt/notification-service:latest

# Kubernetes/Helm targets
.PHONY: kind-create
kind-create:
	@echo "Creating Kind cluster..."
	$(KIND) create cluster --config $(SCRIPTS_DIR)/kind-config.yaml

.PHONY: kind-delete
kind-delete:
	@echo "Deleting Kind cluster..."
	$(KIND) delete cluster --name $(CLUSTER_NAME)

.PHONY: helm-deps
helm-deps:
	@echo "Updating Helm dependencies..."
	$(HELM) dependency update $(HELM_CHART_DIR)

.PHONY: helm-lint
helm-lint:
	@echo "Linting Helm chart..."
	$(HELM) lint $(HELM_CHART_DIR)

.PHONY: helm-template
helm-template:
	@echo "Templating Helm chart..."
	$(HELM) template $(HELM_CHART_DIR) --output-dir $(PROJECT_ROOT)/helm-output

.PHONY: helm-install
helm-install: helm-deps
	@echo "Installing Helm chart..."
	$(KUBECTL) create namespace $(NAMESPACE) --dry-run=client -o yaml | $(KUBECTL) apply -f -
	$(HELM) install $(NAMESPACE) $(HELM_CHART_DIR) --namespace $(NAMESPACE) --set global.storageClass=standard

.PHONY: helm-upgrade
helm-upgrade: helm-deps
	@echo "Upgrading Helm chart..."
	$(HELM) upgrade $(NAMESPACE) $(HELM_CHART_DIR) --namespace $(NAMESPACE) --set global.storageClass=standard

.PHONY: helm-uninstall
helm-uninstall:
	@echo "Uninstalling Helm chart..."
	$(HELM) uninstall $(NAMESPACE) --namespace $(NAMESPACE)

# End-to-end testing targets
.PHONY: e2e-test
e2e-test:
	@echo "Running end-to-end tests..."
	$(SCRIPTS_DIR)/e2e-test.sh

.PHONY: e2e-setup
e2e-setup: kind-create helm-install
	@echo "Setting up end-to-end test environment..."
	$(KUBECTL) apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	$(KUBECTL) wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=300s

.PHONY: e2e-teardown
e2e-teardown: helm-uninstall kind-delete
	@echo "Tearing down end-to-end test environment..."

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf helm-output/
	rm -rf vendor/

.PHONY: help
help:
	@echo "CodeCourt Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  all               Run lint and tests"
	@echo "  deps              Install dependencies"
	@echo "  lint              Run linter"
	@echo "  fmt               Format code"
	@echo "  vet               Run go vet"
	@echo "  test              Run all tests"
	@echo "  test-unit         Run unit tests"
	@echo "  test-integration  Run integration tests"
	@echo "  build             Build services"
	@echo "  docker-build      Build Docker images"
	@echo "  docker-push       Push Docker images"
	@echo "  kind-create       Create Kind cluster"
	@echo "  kind-delete       Delete Kind cluster"
	@echo "  helm-deps         Update Helm dependencies"
	@echo "  helm-lint         Lint Helm chart"
	@echo "  helm-template     Template Helm chart"
	@echo "  helm-install      Install Helm chart"
	@echo "  helm-upgrade      Upgrade Helm chart"
	@echo "  helm-uninstall    Uninstall Helm chart"
	@echo "  e2e-test          Run end-to-end tests"
	@echo "  e2e-setup         Set up end-to-end test environment"
	@echo "  e2e-teardown      Tear down end-to-end test environment"
	@echo "  clean             Clean up build artifacts"
	@echo "  help              Show this help message"
