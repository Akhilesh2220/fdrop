#!/bin/bash
go build -o fdrop fdrop.go
# fdrop installation script
BIN_DIR="/usr/local/bin"
FILENAME="fdrop"
INSTALL_PATH="$BIN_DIR/$FILENAME"

# Check if the binary exists in the current directory
if [ ! -f "./$FILENAME" ]; then
    echo "$FILENAME not found in the current directory."
    exit 1
fi

# Make the binary executable
chmod +x ./$FILENAME

# Copy the binary to a globally accessible location
sudo cp ./$FILENAME $INSTALL_PATH

# Verify the installation
if [ -f "$INSTALL_PATH" ]; then
    echo "$FILENAME has been installed successfully at $INSTALL_PATH."
else
    echo "Installation failed."
    exit 1
fi

