#!/usr/bin/env bash
set -e

# qurl installer script for Linux and macOS
# Usage:
#   curl -sS https://raw.githubusercontent.com/bkeane/qurl/main/install.sh | bash
#   curl -sS https://raw.githubusercontent.com/bkeane/qurl/main/install.sh | bash -s v0.1.0

GITHUB_REPO="bkeane/qurl"
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="qurl"
VERSION="${1:-latest}"  # Use first argument or default to latest

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)

    # Normalize OS name
    case "$os" in
        darwin) os="Darwin" ;;
        linux) os="Linux" ;;
        *) echo -e "${RED}Unsupported OS: $os${NC}" >&2; exit 1 ;;
    esac

    # Normalize architecture
    case "$arch" in
        x86_64) arch="x86_64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) echo -e "${RED}Unsupported architecture: $arch${NC}" >&2; exit 1 ;;
    esac

    echo "${os}_${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    curl -sS "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | \
        grep '"tag_name":' | \
        sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install qurl
install_qurl() {
    local platform=$(detect_platform)
    local version_to_install="$VERSION"

    # If version is "latest", fetch the actual latest version
    if [ "$version_to_install" = "latest" ]; then
        version_to_install=$(get_latest_version)
        if [ -z "$version_to_install" ]; then
            echo -e "${RED}Failed to get latest version${NC}" >&2
            exit 1
        fi
    fi

    echo "Installing qurl $version_to_install for $platform..."

    local download_url="https://github.com/$GITHUB_REPO/releases/download/$version_to_install/qurl_${platform}.tar.gz"
    local temp_dir=$(mktemp -d)

    # Download and extract
    echo "Downloading from $download_url..."
    curl -sL "$download_url" | tar xz -C "$temp_dir"

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Install binary
    mv "$temp_dir/$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    # Clean up
    rm -rf "$temp_dir"

    echo -e "${GREEN}✓ qurl installed successfully to $INSTALL_DIR/$BINARY_NAME${NC}"
}

# Check if directory is in PATH
check_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "${YELLOW}⚠ Warning: $INSTALL_DIR is not in your PATH${NC}"
        echo ""
        echo "Add the following to your shell configuration file:"

        if [ -n "$BASH_VERSION" ]; then
            echo "  # Add to ~/.bashrc or ~/.bash_profile:"
            echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        elif [ -n "$ZSH_VERSION" ]; then
            echo "  # Add to ~/.zshrc:"
            echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        else
            echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        fi
        echo ""
    fi
}

# Show completion setup instructions
show_completion_instructions() {
    echo "To enable shell completions:"
    echo ""

    if [ -n "$BASH_VERSION" ] || [ "$SHELL" = "/bin/bash" ] || [ "$SHELL" = "/usr/bin/bash" ]; then
        echo "  # Bash (add to ~/.bashrc):"
        echo "  source <(qurl completion bash)"
        echo ""
    fi

    if [ -n "$ZSH_VERSION" ] || [ "$SHELL" = "/bin/zsh" ] || [ "$SHELL" = "/usr/bin/zsh" ]; then
        echo "  # Zsh (add to ~/.zshrc):"
        echo "  source <(qurl completion zsh)"
        echo ""
    fi

    if [ "$SHELL" = "/bin/fish" ] || [ "$SHELL" = "/usr/bin/fish" ]; then
        echo "  # Fish (add to ~/.config/fish/config.fish):"
        echo "  qurl completion fish | source"
        echo ""
    fi
}

# Main installation
main() {
    echo "Installing qurl..."
    echo ""

    # Check for required tools
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}Error: curl is required but not installed${NC}" >&2
        exit 1
    fi

    # Install qurl
    install_qurl

    # Check PATH
    check_path

    # Show completion instructions
    show_completion_instructions

    echo "Get started with:"
    echo "  export QURL_OPENAPI='https://your-api.com/openapi.json'"
    echo "  qurl /api/endpoint"
    echo ""
    echo "For more information, visit: https://github.com/$GITHUB_REPO"
}

main "$@"