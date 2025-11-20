# Contributing to Devgraph CLI

Thank you for your interest in contributing to Devgraph CLI! This document provides guidelines and information for contributors.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/devgraph-cli.git`
3. Create a branch for your changes: `git checkout -b feature/your-feature-name`

## Development Setup

### Prerequisites

- Go 1.24 or later
- Make

### Building

```bash
make build
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-cover

# Run tests with verbose output
make test-verbose
```

### Code Quality

```bash
# Format code
make fmt

# Lint code (requires golangci-lint)
make lint
```

## Submitting Changes

### Pull Request Process

1. Ensure your code follows the existing style and passes all tests
2. Update documentation if you're changing functionality
3. Add tests for new features
4. Submit a pull request with a clear description of your changes

### Commit Messages

- Use clear, descriptive commit messages
- Start with a verb in the imperative mood (e.g., "Add", "Fix", "Update")
- Keep the first line under 72 characters

### Code Style

- Follow standard Go conventions and idioms
- Run `make fmt` before committing
- Ensure `make lint` passes without errors

## Reporting Issues

When reporting issues, please include:

- A clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Go version and OS information
- Relevant logs or error messages

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
