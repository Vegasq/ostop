#!/bin/bash
set -e

# ostop installation script
# Usage: curl -sSL https://ostop.mkla.dev/install.sh | bash

REPO="Vegasq/ostop"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="ostop"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}==>${NC} $1"
}

warn() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

error() {
    echo -e "${RED}Error:${NC} $1"
    exit 1
}

# Detect OS
detect_os() {
    local os
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        MINGW*|MSYS*|CYGWIN*) os="windows";;
        *)          error "Unsupported operating system: $(uname -s)";;
    esac
    echo "$os"
}

# Detect architecture
detect_arch() {
    local arch
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64";;
        aarch64|arm64)  arch="arm64";;
        *)              error "Unsupported architecture: $(uname -m)";;
    esac
    echo "$arch"
}

# Get latest release version
get_latest_version() {
    local version
    version=$(curl -sSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        error "Failed to fetch latest version"
    fi
    echo "$version"
}

# Download and install
install_ostop() {
    local os=$(detect_os)
    local arch=$(detect_arch)
    local version=$(get_latest_version)

    info "Detected OS: $os"
    info "Detected Architecture: $arch"
    info "Latest version: $version"

    # Construct download URL
    local filename extension
    if [ "$os" = "windows" ]; then
        extension="zip"
    else
        extension="tar.gz"
    fi

    filename="${BINARY_NAME}-${version}-${os}-${arch}.${extension}"
    local download_url="https://github.com/$REPO/releases/download/${version}/${filename}"

    info "Downloading from: $download_url"

    # Create temporary directory
    local tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Download
    if ! curl -sSL -o "$tmp_dir/$filename" "$download_url"; then
        error "Failed to download $filename"
    fi

    # Extract
    info "Extracting..."
    cd "$tmp_dir"
    if [ "$extension" = "zip" ]; then
        if command -v unzip >/dev/null 2>&1; then
            unzip -q "$filename"
        else
            error "unzip is required but not installed"
        fi
    else
        tar -xzf "$filename"
    fi

    # Find the binary (it might be in a subdirectory)
    local binary_path
    if [ -f "$BINARY_NAME" ]; then
        binary_path="$BINARY_NAME"
    elif [ -f "$BINARY_NAME.exe" ]; then
        binary_path="$BINARY_NAME.exe"
    else
        # Search for it
        binary_path=$(find . -name "$BINARY_NAME" -o -name "$BINARY_NAME.exe" | head -n 1)
    fi

    if [ -z "$binary_path" ]; then
        error "Binary not found in archive"
    fi

    # Install
    info "Installing to $INSTALL_DIR..."

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        # Try with sudo
        if command -v sudo >/dev/null 2>&1; then
            sudo mv "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
            sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
        else
            error "No write permission to $INSTALL_DIR and sudo not available. Try running as root or set INSTALL_DIR to a writable location."
        fi
    fi

    info "Installation complete! 🎉"
    echo ""
    info "Run '$BINARY_NAME --version' to verify installation"
    info "Run '$BINARY_NAME --help' to see usage instructions"

    # Check if INSTALL_DIR is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "$INSTALL_DIR is not in your PATH"
        echo "  Add it by running: export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
}

# Main
main() {
    info "Installing ostop - OpenSearch Terminal UI"
    echo ""

    # Check for required commands
    for cmd in curl tar; do
        if ! command -v $cmd >/dev/null 2>&1; then
            error "$cmd is required but not installed"
        fi
    done

    install_ostop
}

main
