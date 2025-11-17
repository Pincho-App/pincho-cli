#!/bin/sh
# WirePusher CLI Installer
# Usage: curl -sSL https://gitlab.com/wirepusher/wirepusher-cli/-/raw/main/install.sh | sh

set -e

REPO="wirepusher/wirepusher-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="wirepusher"
PROJECT_NAME="wirepusher-cli"  # Name used in archive files

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

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *)       error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        armv7l|armv7)  echo "armv7" ;;
        *)             error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest version from GitLab API
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -sSL "https://gitlab.com/api/v4/projects/wirepusher%2Fwirepusher-cli/releases" | \
            grep -o '"tag_name":"[^"]*"' | head -1 | cut -d'"' -f4
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://gitlab.com/api/v4/projects/wirepusher%2Fwirepusher-cli/releases" | \
            grep -o '"tag_name":"[^"]*"' | head -1 | cut -d'"' -f4
    else
        error "curl or wget is required"
    fi
}

# Download file
download() {
    url="$1"
    output="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        error "curl or wget is required"
    fi
}

main() {
    info "WirePusher CLI Installer"
    echo ""

    # Detect platform
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected platform: ${OS}/${ARCH}"

    # Get latest version
    info "Fetching latest version..."
    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        error "Failed to fetch latest version. Please check https://gitlab.com/${REPO}/-/releases"
    fi
    info "Latest version: ${VERSION}"

    # Construct download URL
    # Version without 'v' prefix for archive name
    VERSION_NUM=$(echo "$VERSION" | sed 's/^v//')

    if [ "$OS" = "windows" ]; then
        ARCHIVE_NAME="${PROJECT_NAME}_${VERSION_NUM}_${OS}_${ARCH}.zip"
    else
        ARCHIVE_NAME="${PROJECT_NAME}_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
    fi

    DOWNLOAD_URL="https://gitlab.com/${REPO}/-/releases/${VERSION}/downloads/${ARCHIVE_NAME}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    # Download archive
    info "Downloading ${ARCHIVE_NAME}..."
    ARCHIVE_PATH="${TMP_DIR}/${ARCHIVE_NAME}"
    download "$DOWNLOAD_URL" "$ARCHIVE_PATH"

    if [ ! -f "$ARCHIVE_PATH" ]; then
        error "Download failed. Please check https://gitlab.com/${REPO}/-/releases/${VERSION}"
    fi

    # Extract archive
    info "Extracting..."
    cd "$TMP_DIR"
    if [ "$OS" = "windows" ]; then
        unzip -q "$ARCHIVE_NAME"
    else
        tar -xzf "$ARCHIVE_NAME"
    fi

    # Find binary
    if [ ! -f "$BINARY_NAME" ]; then
        error "Binary not found in archive"
    fi

    # Check if install directory is writable
    if [ ! -w "$INSTALL_DIR" ]; then
        warn "Install directory ${INSTALL_DIR} is not writable"
        warn "Attempting to install with sudo..."
        SUDO="sudo"
    else
        SUDO=""
    fi

    # Install binary
    info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
    $SUDO mv "$BINARY_NAME" "${INSTALL_DIR}/${BINARY_NAME}"
    $SUDO chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    # Verify installation
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        INSTALLED_VERSION=$("$BINARY_NAME" --version 2>/dev/null | head -1 || echo "unknown")
        echo ""
        info "Successfully installed WirePusher CLI!"
        info "Version: ${INSTALLED_VERSION}"
        info "Location: ${INSTALL_DIR}/${BINARY_NAME}"
        echo ""
        info "Get started:"
        echo "  1. Get your token: Open WirePusher app → Settings → Help → Copy token"
        echo "  2. Send a test notification:"
        echo "     wirepusher send \"Hello\" \"Test from CLI\" --token YOUR_TOKEN"
        echo ""
        info "For more info: wirepusher --help"
    else
        warn "Installation complete, but ${BINARY_NAME} is not in your PATH"
        warn "You may need to add ${INSTALL_DIR} to your PATH"
        warn "Or run: ${INSTALL_DIR}/${BINARY_NAME}"
    fi
}

main "$@"
