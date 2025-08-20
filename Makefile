# Devgraph CLI Makefile

# Binary name (can be overridden)
BINARY_NAME ?= devgraph

.PHONY: test test-verbose test-cover build clean lint fmt help

# Build the binary
build:
	go build -o $(BINARY_NAME) .

# Default target
help:
	@echo "Available targets:"
	@echo "  test         - Run all tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  test-cover   - Run tests with coverage report"
	@echo "  build        - Build the binary"
	@echo "  clean        - Clean build artifacts"
	@echo "  lint         - Run linter (requires golangci-lint)"
	@echo "  fmt          - Format code"
	@echo "  security     - Run security scans"
	@echo "  vuln-check   - Check for vulnerabilities"
	@echo "  deps-check   - Check for outdated dependencies"

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage
test-cover:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) coverage.out coverage.html

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod tidy
	go mod download

# Security scanning targets
security: vuln-check
	@echo "Running security scans..."

# Check for known vulnerabilities
vuln-check:
	@echo "Checking for vulnerabilities..."
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# Check for outdated dependencies
deps-check:
	@echo "Checking for outdated dependencies..."
	go list -u -m all
