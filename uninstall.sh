#!/bin/bash

# fdrop uninstallation script
BIN_DIR="/usr/local/bin"
FILENAME="fdrop"
INSTALL_PATH="$BIN_DIR/$FILENAME"

echo "Uninstalling $FILENAME from $INSTALL_PATH..."

# Check if the binary exists
if [ -f "$INSTALL_PATH" ]; then
    sudo rm -f "$INSTALL_PATH"
    
    # Confirm removal
    if [ ! -f "$INSTALL_PATH" ]; then
        echo "✅ $FILENAME has been uninstalled successfully."
    else
        echo "❌ Failed to remove $FILENAME."
        exit 1
    fi
else
    echo "ℹ️  $FILENAME is not installed in $INSTALL_PATH."
fi

