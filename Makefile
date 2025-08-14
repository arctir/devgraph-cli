# Devgraph CLI Makefile

.PHONY: test test-verbose test-cover build clean lint fmt help

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

# Build the binary
build:
	go build -o devgraph-cli .

# Clean build artifacts
clean:
	rm -f devgraph-cli coverage.out coverage.html

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
