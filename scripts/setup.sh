#!/bin/bash

# Network Recon Toolkit - Build and Setup Script

set -e

echo "=== Network Recon Toolkit Setup ==="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
REQUIRED_VERSION="1.21"

if ! printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V -C; then
    echo "Error: Go version $GO_VERSION is too old. Requires $REQUIRED_VERSION or later."
    exit 1
fi

echo "✓ Go version: $GO_VERSION"

# Check if nmap is installed
if command -v nmap &> /dev/null; then
    echo "✓ nmap found: $(nmap --version | head -1)"
else
    echo "⚠ nmap not found. Install with:"
    echo "  macOS: brew install nmap"
    echo "  Ubuntu/Debian: sudo apt-get install nmap"
    echo "  CentOS/RHEL: sudo yum install nmap"
fi

# Check if masscan is installed
if command -v masscan &> /dev/null; then
    echo "✓ masscan found: $(masscan --version 2>&1 | head -1)"
else
    echo "⚠ masscan not found. Install with:"
    echo "  macOS: brew install masscan"
    echo "  Ubuntu/Debian: sudo apt-get install masscan"
    echo "  Build from source: https://github.com/robertdavidgraham/masscan"
fi

# Check if Docker is installed
if command -v docker &> /dev/null; then
    echo "✓ Docker found: $(docker --version)"
else
    echo "⚠ Docker not found. Install from https://docker.com"
fi

# Check if Docker Compose is installed
if command -v docker-compose &> /dev/null; then
    echo "✓ Docker Compose found: $(docker-compose --version)"
elif docker compose version &> /dev/null; then
    echo "✓ Docker Compose (plugin) found"
else
    echo "⚠ Docker Compose not found"
fi

echo ""
echo "=== Building Application ==="

# Download dependencies
echo "Downloading Go dependencies..."
go mod download

# Build the application
echo "Building netrecon..."
go build -o bin/netrecon ./cmd/netrecon

echo "✓ Build completed: bin/netrecon"

# Make executable
chmod +x bin/netrecon

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Usage:"
echo "  1. Configure database in configs/config.yaml"
echo "  2. Run with Docker: docker-compose up -d"
echo "  3. Or run directly: ./bin/netrecon --help"
echo ""
echo "Examples:"
echo "  ./bin/netrecon scan 192.168.1.1/24"
echo "  ./bin/netrecon scan --scanner masscan --ports 1-1000 example.com"
echo "  ./bin/netrecon target add 192.168.1.0/24 'Internal network'"
echo "  ./bin/netrecon server"
echo ""