# Devgraph CLI

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/arctir/devgraph-cli)](https://goreportcard.com/report/github.com/arctir/devgraph-cli)

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

### Getting Started

```bash
# Authenticate with your Devgraph account (also configures your environment)
dg auth login

# Start an interactive chat with AI
dg chat
```

### Resource Management

```bash
# Environments
dg env list
dg env create

# API tokens
dg token list
dg token create

# Entities
dg entity list
dg entity get <name>

# Entity definitions
dg entitydefinition list

# MCP resources
dg mcp list

# Models
dg model list

# Model providers
dg modelprovider list

# Discovery providers
dg provider list

# OAuth services
dg oauthservice list

# Subscriptions
dg subscription list
```

### Configuration

```bash
# View current context
dg config current-context

# List contexts
dg config get-contexts

# Switch context
dg config use-context <name>
```

### Getting Help

```bash
# Show all available commands
dg --help

# Get help for a specific command
dg chat --help
dg auth --help

# Generate shell completions
dg completion bash
dg completion zsh
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
