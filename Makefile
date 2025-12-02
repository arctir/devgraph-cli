# Devgraph CLI Makefile

# Binary name (can be overridden)
BINARY_NAME ?= dg
INSTALL_PATH ?= ~/bin

# Detect shell and OS
# Try multiple methods to detect the actual user shell
USER_SHELL := $(shell basename "$$SHELL" 2>/dev/null || echo "")
ifeq ($(USER_SHELL),)
	USER_SHELL := $(shell getent passwd $$USER 2>/dev/null | cut -d: -f7 | xargs basename)
endif
ifeq ($(USER_SHELL),sh)
	# If detected as 'sh', check if it's actually bash
	USER_SHELL := $(shell if [ -n "$$BASH_VERSION" ]; then echo bash; else echo sh; fi)
endif
SHELL_TYPE := $(USER_SHELL)
UNAME_S := $(shell uname -s)

.PHONY: test test-verbose test-cover build clean lint fmt help install uninstall install-completions

# Build the binary
build:
	go build -o $(BINARY_NAME) .

# Install the binary and completions
install: build install-completions
	install -d $(INSTALL_PATH)
	install -m 755 $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo ""
	@echo "✓ Installation complete!"
	@echo ""
	@echo "Shell completions have been installed."
	@echo "Restart your shell or source the completion file to enable completions."

# Install shell completions
install-completions: build
	@echo "Installing shell completions..."
ifeq ($(SHELL_TYPE),zsh)
	@mkdir -p ~/.zsh/completions
	@./$(BINARY_NAME) completion zsh > ~/.zsh/completions/_$(BINARY_NAME)
	@echo "✓ Zsh completions installed to ~/.zsh/completions/_$(BINARY_NAME)"
	@echo ""
	@echo "  To enable now, run:"
	@echo "    source ~/.zsh/completions/_$(BINARY_NAME)"
	@echo ""
	@echo "  To enable permanently, add to your ~/.zshrc if not present:"
	@echo "    fpath=(~/.zsh/completions \$$fpath) && autoload -Uz compinit && compinit"
else ifeq ($(SHELL_TYPE),fish)
	@mkdir -p ~/.config/fish/completions
	@./$(BINARY_NAME) completion fish > ~/.config/fish/completions/$(BINARY_NAME).fish
	@echo "✓ Fish completions installed to ~/.config/fish/completions/$(BINARY_NAME).fish"
	@echo ""
	@echo "  Completions will be automatically loaded in new fish shells"
else ifeq ($(SHELL_TYPE),bash)
ifeq ($(UNAME_S),Darwin)
	@if [ -d /usr/local/etc/bash_completion.d ]; then \
		./$(BINARY_NAME) completion bash > /usr/local/etc/bash_completion.d/$(BINARY_NAME); \
		echo "✓ Bash completions installed to /usr/local/etc/bash_completion.d/$(BINARY_NAME)"; \
		echo ""; \
		echo "  To enable now, run:"; \
		echo "    source /usr/local/etc/bash_completion.d/$(BINARY_NAME)"; \
		echo ""; \
		echo "  Ensure bash-completion is installed via Homebrew and loaded in ~/.bash_profile"; \
	else \
		mkdir -p ~/.local/share/bash-completion/completions; \
		./$(BINARY_NAME) completion bash > ~/.local/share/bash-completion/completions/$(BINARY_NAME); \
		echo "✓ Bash completions installed to ~/.local/share/bash-completion/completions/$(BINARY_NAME)"; \
		echo ""; \
		echo "  To enable now, run:"; \
		echo "    source ~/.local/share/bash-completion/completions/$(BINARY_NAME)"; \
		echo ""; \
		echo "  To enable permanently, add to your ~/.bashrc if not present:"; \
		echo "    for f in ~/.local/share/bash-completion/completions/*; do source \"\$$f\"; done"; \
	fi
else
	@mkdir -p ~/.local/share/bash-completion/completions
	@./$(BINARY_NAME) completion bash > ~/.local/share/bash-completion/completions/$(BINARY_NAME)
	@echo "✓ Bash completions installed to ~/.local/share/bash-completion/completions/$(BINARY_NAME)"
	@echo ""
	@echo "  To enable now, run:"
	@echo "    source ~/.local/share/bash-completion/completions/$(BINARY_NAME)"
	@echo ""
	@echo "  To enable permanently, add to your ~/.bashrc if not present:"
	@echo "    for f in ~/.local/share/bash-completion/completions/*; do source \"\$$f\"; done"
endif
else
	@echo "⚠ Shell type '$(SHELL_TYPE)' not recognized. Attempting bash installation..."
	@mkdir -p ~/.local/share/bash-completion/completions
	@./$(BINARY_NAME) completion bash > ~/.local/share/bash-completion/completions/$(BINARY_NAME)
	@echo "✓ Bash completions installed to ~/.local/share/bash-completion/completions/$(BINARY_NAME)"
	@echo ""
	@echo "  To enable now, run:"
	@echo "    source ~/.local/share/bash-completion/completions/$(BINARY_NAME)"
	@echo ""
	@echo "  If you use a different shell, run manually:"
	@echo "    ./$(BINARY_NAME) completion <bash|zsh|fish> --install"
endif

# Uninstall the binary and completions
uninstall: uninstall-completions
	rm -f $(INSTALL_PATH)/$(BINARY_NAME)

# Uninstall shell completions
uninstall-completions:
	@echo "Uninstalling shell completions..."
	@rm -f ~/.zsh/completions/_$(BINARY_NAME)
	@rm -f ~/.config/fish/completions/$(BINARY_NAME).fish
	@rm -f ~/.local/share/bash-completion/completions/$(BINARY_NAME)
	@rm -f /usr/local/etc/bash_completion.d/$(BINARY_NAME)
	@echo "✓ Shell completions removed"

# Default target
help:
	@echo "Available targets:"
	@echo "  test                - Run all tests"
	@echo "  test-verbose        - Run tests with verbose output"
	@echo "  test-cover          - Run tests with coverage report"
	@echo "  build               - Build the binary"
	@echo "  install             - Install the binary and shell completions to $(INSTALL_PATH)"
	@echo "  install-completions - Install only shell completions"
	@echo "  uninstall           - Uninstall the binary and completions from $(INSTALL_PATH)"
	@echo "  uninstall-completions - Uninstall only shell completions"
	@echo "  clean               - Clean build artifacts"
	@echo "  lint                - Run linter (requires golangci-lint)"
	@echo "  fmt                 - Format code"
	@echo "  security            - Run security scans"
	@echo "  vuln-check          - Check for vulnerabilities"
	@echo "  deps-check          - Check for outdated dependencies"

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
