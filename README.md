# Devgraph CLI

> Turn chaos into clarity

A command-line interface for interacting with Devgraph, providing AI-powered chat, authentication, and resource management capabilities.

## Features

- **Interactive AI Chat** - Converse with Devgraph.ai via the command line
- **Devgraph Resource Management**
    - **Token Management** - Create and manage API tokens
    - **Environment Management** - Manage different Devgraph environments
    - **MCP Resources** - Manage Model Context Protocol resources
    - **Model Providers** - Configure and manage AI model providers
    - **Model Management** - Manage AI models and configurations

## Installation

### Download from GitHub Releases

Download the latest release for your platform:

```bash
# Quick install script (downloads latest release for your platform)
curl -fsSL https://raw.githubusercontent.com/arctir/devgraph-cli/main/install.sh | bash

# Or download manually from GitHub releases
# Visit: https://github.com/arctir/devgraph-cli/releases
```

Available platforms:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

### From Source

```bash
git clone https://github.com/arctir/devgraph-cli.git
cd devgraph-cli
make build
```

This will create a binary named `devgraph` in the current directory.

### Custom Binary Name

You can specify a custom binary name during build:

```bash
make build BINARY_NAME=my-devgraph-tool
```

## Usage

### Basic Commands

```bash
# Start an interactive chat with AI
devgraph chat

# Authenticate your client
devgraph auth

# Manage tokens
devgraph token list
devgraph token create

# Manage environments
devgraph env list
devgraph env create

# Manage MCP resources
devgraph mcp list

# Manage model providers
devgraph modelprovider list

# Manage models
devgraph model list
```

### Getting Help

```bash
# Show all available commands
devgraph --help

# Get help for a specific command
devgraph chat --help
devgraph auth --help
```

## Development

### Prerequisites

- Go 1.24 or later

### Building

```bash
# Build the binary
make build

# Run tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage
make test-cover

# Format code
make fmt

# Lint code (requires golangci-lint)
make lint

# Clean build artifacts
make clean
```