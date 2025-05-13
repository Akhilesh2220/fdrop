#!/bin/bash

# fdrop uninstallation script
BIN_DIR="/usr/local/bin"
FILENAME="fdrop"
INSTALL_PATH="$BIN_DIR/$FILENAME"

# Check if the binary exists in the globally accessible location
if [ -f "$INSTALL_PATH" ]; then
    sudo rm -f $INSTALL_PATH
    echo "$FILENAME has been uninstalled."
else
    echo "$FILENAME is not installed in $INSTALL_PATH."
fi

