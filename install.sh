#!/bin/bash

set -e

BIN_DIR="/usr/local/bin"
INSTALL_NAME="fdrop"
BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN_FOLDER="$BASE_DIR/bin"

# Detect OS
OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Normalize ARCH
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Compose binary filename
BINARY="$BIN_FOLDER/fdrop-${OS}-${ARCH}"

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo "No prebuilt binary found for $OS/$ARCH."
    exit 1
fi

# Install
chmod +x "$BINARY"
sudo cp "$BINARY" "$BIN_DIR/$INSTALL_NAME"

# Confirm
if [ -f "$BIN_DIR/$INSTALL_NAME" ]; then
    echo "✅ fdrop installed successfully at $BIN_DIR/$INSTALL_NAME"
    $BIN_DIR/$INSTALL_NAME --help || true
else
    echo "❌ Installation failed."
    exit 1
fi

