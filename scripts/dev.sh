#!/bin/bash

# Development helper script for Network Recon Toolkit

set -e

echo "=== Network Recon Toolkit - Development Tools ==="

# Function to check dependencies
check_deps() {
    echo "Checking development dependencies..."
    
    # Check if golangci-lint is installed
    if ! command -v golangci-lint &> /dev/null; then
        echo "Installing golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi
    
    echo "✓ Development dependencies ready"
}

# Function to run tests
run_tests() {
    echo "Running tests..."
    go test -v ./...
    echo "✓ Tests completed"
}

# Function to run linting
run_lint() {
    echo "Running linter..."
    golangci-lint run
    echo "✓ Linting completed"
}

# Function to build project
build_project() {
    echo "Building project..."
    mkdir -p bin
    go build -o bin/netrecon ./cmd/netrecon
    echo "✓ Build completed"
}

# Function to run formatting
format_code() {
    echo "Formatting code..."
    go fmt ./...
    echo "✓ Code formatted"
}

# Function to generate docs
generate_docs() {
    echo "Generating documentation..."
    go doc -all ./... > docs/api.md
    echo "✓ Documentation generated"
}

# Function to clean build artifacts
clean() {
    echo "Cleaning build artifacts..."
    rm -rf bin/
    go clean
    echo "✓ Clean completed"
}

# Main script logic
case "${1:-help}" in
    "deps"|"dependencies")
        check_deps
        ;;
    "test")
        run_tests
        ;;
    "lint")
        run_lint
        ;;
    "build")
        build_project
        ;;
    "fmt"|"format")
        format_code
        ;;
    "docs")
        generate_docs
        ;;
    "clean")
        clean
        ;;
    "all")
        check_deps
        format_code
        run_lint
        run_tests
        build_project
        echo "✓ All development tasks completed"
        ;;
    "help"|*)
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  deps      - Install development dependencies"
        echo "  test      - Run tests"
        echo "  lint      - Run linter"
        echo "  build     - Build project"
        echo "  format    - Format code"
        echo "  docs      - Generate documentation"
        echo "  clean     - Clean build artifacts"
        echo "  all       - Run all development tasks"
        echo "  help      - Show this help message"
        ;;
esac