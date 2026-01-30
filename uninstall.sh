#!/bin/sh
# Pincho CLI Uninstaller
# Usage: curl -sSL https://gitlab.com/pincho-app/pincho-cli/-/raw/main/uninstall.sh | sh

set -e

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="pincho"
CONFIG_DIR="${HOME}/.pincho"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    printf "${GREEN}[INFO]${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1"
}

error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1"
    exit 1
}

main() {
    info "Pincho CLI Uninstaller"
    echo ""

    # Check if binary exists
    BINARY_PATH="${INSTALL_DIR}/${BINARY_NAME}"

    if [ ! -f "$BINARY_PATH" ]; then
        # Try to find it in PATH
        if command -v "$BINARY_NAME" >/dev/null 2>&1; then
            BINARY_PATH=$(command -v "$BINARY_NAME")
            info "Found ${BINARY_NAME} at: ${BINARY_PATH}"
        else
            warn "Pincho CLI is not installed at ${INSTALL_DIR}/${BINARY_NAME}"
            warn "Or not found in PATH"
            exit 0
        fi
    fi

    # Check if we need sudo
    BINARY_DIR=$(dirname "$BINARY_PATH")
    if [ ! -w "$BINARY_DIR" ]; then
        warn "Directory ${BINARY_DIR} is not writable"
        warn "Attempting to remove with sudo..."
        SUDO="sudo"
    else
        SUDO=""
    fi

    # Remove binary
    info "Removing ${BINARY_PATH}..."
    $SUDO rm -f "$BINARY_PATH"
    info "Binary removed"

    # Ask about config directory
    if [ -d "$CONFIG_DIR" ]; then
        echo ""
        warn "Configuration directory found: ${CONFIG_DIR}"
        printf "Do you want to remove it? This will delete your saved token. [y/N] "
        read -r response
        case "$response" in
            [yY][eE][sS]|[yY])
                info "Removing ${CONFIG_DIR}..."
                rm -rf "$CONFIG_DIR"
                info "Configuration removed"
                ;;
            *)
                info "Keeping configuration directory"
                ;;
        esac
    fi

    echo ""
    info "Pincho CLI has been uninstalled"
}

main "$@"
