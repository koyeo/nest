#!/bin/bash
set -euo pipefail

# ─── Nest CLI Installer ─────────────────────────────────────
# Auto-detects OS and architecture, downloads the latest binary.
# Usage: curl -fsSL https://raw.githubusercontent.com/koyeo/nest/main/scripts/install.sh | bash

REPO="koyeo/nest"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="nest"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    darwin) OS="darwin" ;;
    linux)  OS="linux" ;;
    *)
        echo "❌ Unsupported OS: $OS"
        echo "   Supported: macOS (darwin), Linux"
        exit 1
        ;;
esac

# Detect Architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        echo "   Supported: x86_64 (amd64), arm64 (aarch64)"
        exit 1
        ;;
esac

BINARY="${BINARY_NAME}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"

echo "🔍 Detected: ${OS}/${ARCH}"
echo "📦 Downloading ${BINARY}..."

# Download
TMP_FILE=$(mktemp)
if ! curl -fsSL "$URL" -o "$TMP_FILE"; then
    rm -f "$TMP_FILE"
    echo "❌ Download failed. Check if the release exists:"
    echo "   https://github.com/${REPO}/releases"
    exit 1
fi

# Install
chmod +x "$TMP_FILE"
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
else
    echo "📝 Requires sudo to install to ${INSTALL_DIR}"
    sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo "✅ Installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"
echo ""
${INSTALL_DIR}/${BINARY_NAME} --help 2>/dev/null | head -2 || true
