# Network Recon Toolkit Makefile

# Version and build information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILT_BY := $(shell whoami)

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
BINARY_NAME := netrecon
MAIN_PATH := ./cmd/netrecon

# Build flags
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -X main.builtBy=$(BUILT_BY)"
BUILD_FLAGS := -trimpath $(LDFLAGS)

# Directories
BUILD_DIR := bin
DIST_DIR := dist
COVERAGE_DIR := coverage

# Docker
DOCKER_IMAGE := ghcr.io/phutran1210dev/network-recon-toolkit
DOCKER_TAG := $(VERSION)

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: all build clean test coverage deps lint security docker docker-build docker-push help

# Default target
all: clean deps test build

## Build targets

# Build the binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME) $(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Build for all platforms
build-all:
	@echo "$(GREEN)Building for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-armv7 $(MAIN_PATH)
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "$(GREEN)Multi-platform build completed$(NC)"

# Quick build for development
dev:
	@echo "$(YELLOW)Building development version...$(NC)"
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

## Test targets

# Run tests
test:
	@echo "$(BLUE)Running tests...$(NC)"
	$(GOTEST) -v -race ./...

# Run tests with coverage
coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_DIR)/coverage.html$(NC)"

# Run benchmarks
bench:
	@echo "$(BLUE)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

## Quality targets

# Install dependencies
deps:
	@echo "$(BLUE)Installing dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy

# Run linting
lint:
	@echo "$(BLUE)Running linters...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(RED)golangci-lint not found. Install it: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)" && exit 1)
	golangci-lint run --timeout=5m

# Format code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOCMD) fmt ./...

# Run security checks
security:
	@echo "$(BLUE)Running security checks...$(NC)"
	@which gosec > /dev/null || (echo "$(YELLOW)Installing gosec...$(NC)" && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...
	@which govulncheck > /dev/null || (echo "$(YELLOW)Installing govulncheck...$(NC)" && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

## Docker targets

# Build Docker image
docker-build:
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(DATE) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		.
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

# Push Docker image
docker-push: docker-build
	@echo "$(BLUE)Pushing Docker image...$(NC)"
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker image pushed$(NC)"

# Run Docker container
docker-run: docker-build
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run --rm -it $(DOCKER_IMAGE):$(DOCKER_TAG) --help

# Start development environment with Docker Compose
docker-dev:
	@echo "$(BLUE)Starting development environment...$(NC)"
	docker-compose up -d postgres redis
	@echo "$(GREEN)Development environment started$(NC)"
	@echo "$(YELLOW)PostgreSQL: localhost:5432$(NC)"
	@echo "$(YELLOW)Redis: localhost:6379$(NC)"

# Stop development environment
docker-dev-stop:
	@echo "$(BLUE)Stopping development environment...$(NC)"
	docker-compose down
	@echo "$(GREEN)Development environment stopped$(NC)"

## Release targets

# Create a new release
release: clean deps test lint security build-all
	@echo "$(GREEN)Creating release $(VERSION)...$(NC)"
	@which goreleaser > /dev/null || (echo "$(RED)goreleaser not found. Install it: go install github.com/goreleaser/goreleaser@latest$(NC)" && exit 1)
	goreleaser release --clean
	@echo "$(GREEN)Release $(VERSION) created$(NC)"

# Create a snapshot release (no Git tag required)
snapshot: clean deps test build-all
	@echo "$(GREEN)Creating snapshot release...$(NC)"
	@which goreleaser > /dev/null || (echo "$(RED)goreleaser not found. Install it: go install github.com/goreleaser/goreleaser@latest$(NC)" && exit 1)
	goreleaser release --snapshot --clean
	@echo "$(GREEN)Snapshot release created in $(DIST_DIR)$(NC)"

## Installation targets

# Install binary to system
install: build
	@echo "$(BLUE)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)$(BINARY_NAME) installed successfully$(NC)"

# Install binary to user directory
install-user: build
	@echo "$(BLUE)Installing $(BINARY_NAME) to ~/.local/bin...$(NC)"
	@mkdir -p ~/.local/bin
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	@echo "$(GREEN)$(BINARY_NAME) installed to ~/.local/bin$(NC)"
	@echo "$(YELLOW)Make sure ~/.local/bin is in your PATH$(NC)"

# Uninstall binary
uninstall:
	@echo "$(BLUE)Uninstalling $(BINARY_NAME)...$(NC)"
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	rm -f ~/.local/bin/$(BINARY_NAME)
	@echo "$(GREEN)$(BINARY_NAME) uninstalled$(NC)"

## Utility targets

# Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@echo "$(GREEN)Clean completed$(NC)"

# Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"
	@echo "Built by: $(BUILT_BY)"

# Show project statistics
stats:
	@echo "$(BLUE)Project Statistics:$(NC)"
	@echo "Lines of code:"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
	@echo ""
	@echo "Go files:"
	@find . -name "*.go" -not -path "./vendor/*" | wc -l
	@echo ""
	@echo "Packages:"
	@go list ./... | wc -l
	@echo ""
	@echo "Dependencies:"
	@go list -m all | wc -l

# Generate documentation
docs:
	@echo "$(BLUE)Generating documentation...$(NC)"
	@mkdir -p docs/api
	godoc -http=:6060 &
	@echo "$(GREEN)Documentation server started at http://localhost:6060$(NC)"

# Run all quality checks
check: deps fmt lint security test coverage
	@echo "$(GREEN)All quality checks completed$(NC)"

# Show help
help:
	@echo "$(BLUE)Network Recon Toolkit - Makefile Help$(NC)"
	@echo ""
	@echo "$(YELLOW)Build targets:$(NC)"
	@echo "  build         Build the binary"
	@echo "  build-all     Build for all platforms"
	@echo "  dev           Quick development build"
	@echo ""
	@echo "$(YELLOW)Test targets:$(NC)"
	@echo "  test          Run tests"
	@echo "  coverage      Run tests with coverage"
	@echo "  bench         Run benchmarks"
	@echo ""
	@echo "$(YELLOW)Quality targets:$(NC)"
	@echo "  deps          Install dependencies"
	@echo "  lint          Run linters"
	@echo "  fmt           Format code"
	@echo "  security      Run security checks"
	@echo "  check         Run all quality checks"
	@echo ""
	@echo "$(YELLOW)Docker targets:$(NC)"
	@echo "  docker-build  Build Docker image"
	@echo "  docker-push   Push Docker image"
	@echo "  docker-run    Run Docker container"
	@echo "  docker-dev    Start development environment"
	@echo ""
	@echo "$(YELLOW)Release targets:$(NC)"
	@echo "  release       Create a new release"
	@echo "  snapshot      Create a snapshot release"
	@echo ""
	@echo "$(YELLOW)Installation targets:$(NC)"
	@echo "  install       Install to /usr/local/bin"
	@echo "  install-user  Install to ~/.local/bin"
	@echo "  uninstall     Remove installed binary"
	@echo ""
	@echo "$(YELLOW)Utility targets:$(NC)"
	@echo "  clean         Clean build artifacts"
	@echo "  version       Show version information"
	@echo "  stats         Show project statistics"
	@echo "  docs          Generate documentation"
	@echo "  help          Show this help message"
	@echo ""
	@echo "$(YELLOW)Environment Variables:$(NC)"
	@echo "  VERSION       Override version (default: git describe)"
	@echo "  DOCKER_IMAGE  Override Docker image name"
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make build                    # Build for current platform"
	@echo "  make test coverage            # Run tests with coverage"
	@echo "  make check                    # Run all quality checks"
	@echo "  make docker-build DOCKER_TAG=v1.0.0  # Build Docker image with custom tag"