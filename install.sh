#!/bin/bash

# Devgraph CLI Installation Script
# Downloads and installs the latest release from GitHub

set -e

# Configuration
REPO="arctir/devgraph-cli"
BINARY_NAME="dg"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Detect platform and architecture
detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          error "Unsupported operating system: $(uname -s)" ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac

    echo "${os}_${arch}"
}

# Get latest release tag from GitHub API
get_latest_release() {
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"
    local release_info

    log "Fetching latest release information..." >&2

    if command -v curl >/dev/null 2>&1; then
        release_info=$(curl -s "$api_url")
    elif command -v wget >/dev/null 2>&1; then
        release_info=$(wget -qO- "$api_url")
    else
        error "Either curl or wget is required to download the release"
    fi

    echo "$release_info" | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

# Download and install binary
install_binary() {
    local platform="$1"
    local version="$2"
    local version_number="${version#v}"  # Remove 'v' prefix for filename
    local archive_name="devgraph-cli_${version_number}_${platform}.tar.gz"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"
    local temp_dir="/tmp/devgraph-cli-install"
    local temp_archive="${temp_dir}/${archive_name}"

    log "Downloading ${BINARY_NAME} ${version} for ${platform}..."

    # Create temp directory
    mkdir -p "$temp_dir"

    if command -v curl >/dev/null 2>&1; then
        curl -sL "$download_url" -o "$temp_archive"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$download_url" -O "$temp_archive"
    else
        error "Either curl or wget is required to download the release"
    fi

    if [ ! -f "$temp_archive" ]; then
        error "Failed to download archive from $download_url"
    fi

    log "Extracting archive..."
    tar -xzf "$temp_archive" -C "$temp_dir"

    # Find the binary in the extracted files
    local binary_path
    binary_path=$(find "$temp_dir" -name "$BINARY_NAME" -type f | head -n 1)

    if [ -z "$binary_path" ]; then
        error "Binary '$BINARY_NAME' not found in archive"
    fi

    # Make binary executable
    chmod +x "$binary_path"

    # Check if install directory is writable
    if [ ! -w "$INSTALL_DIR" ]; then
        warn "Install directory $INSTALL_DIR is not writable. Trying with sudo..."
        sudo mv "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    else
        mv "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Clean up
    rm -rf "$temp_dir"

    success "${BINARY_NAME} installed to $INSTALL_DIR/$BINARY_NAME"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version
        version=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        success "Installation verified! Run '$BINARY_NAME --help' to get started."
        log "Installed version: $version"
    else
        warn "Binary installed but not found in PATH. You may need to add $INSTALL_DIR to your PATH."
        log "Add this to your shell profile: export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
}

# Main installation process
main() {
    log "Starting Devgraph CLI installation..."

    # Detect platform
    local platform
    platform=$(detect_platform)
    log "Detected platform: $platform"

    # Get latest release
    local version
    version=$(get_latest_release)
    if [ -z "$version" ]; then
        error "Failed to fetch latest release information"
    fi
    log "Latest version: $version"

    # Install binary
    install_binary "$platform" "$version"

    # Verify installation
    verify_installation

    success "Devgraph CLI installation completed!"
}

# Handle command line arguments
case "${1:-}" in
    -h|--help)
        echo "Devgraph CLI Installation Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  -h, --help     Show this help message"
        echo "  -v, --version  Show version and exit"
        echo ""
        echo "Environment variables:"
        echo "  INSTALL_DIR    Installation directory (default: /usr/local/bin)"
        echo ""
        echo "Examples:"
        echo "  $0                           # Install to /usr/local/bin"
        echo "  INSTALL_DIR=~/.local/bin $0  # Install to ~/.local/bin"
        exit 0
        ;;
    -v|--version)
        echo "Devgraph CLI Installation Script v1.0.0"
        exit 0
        ;;
    "")
        main
        ;;
    *)
        error "Unknown option: $1. Use -h or --help for usage information."
        ;;
esac
