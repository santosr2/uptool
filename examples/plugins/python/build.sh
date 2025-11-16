#!/bin/bash
# Build script for uptool Python plugin

set -e

PLUGIN_NAME="python"
VERSION="${VERSION:-dev}"

echo "Building ${PLUGIN_NAME} plugin (version: ${VERSION})..."

# Build plugin
go build -buildmode=plugin -o "${PLUGIN_NAME}.so" .

echo "✓ Built ${PLUGIN_NAME}.so"

# Optionally install to user directory
if [ "$1" == "install" ]; then
    INSTALL_DIR="$HOME/.uptool/plugins"
    mkdir -p "$INSTALL_DIR"
    cp "${PLUGIN_NAME}.so" "$INSTALL_DIR/"
    echo "✓ Installed to $INSTALL_DIR"
elif [ "$1" == "install-system" ]; then
    INSTALL_DIR="/usr/local/lib/uptool/plugins"
    sudo mkdir -p "$INSTALL_DIR"
    sudo cp "${PLUGIN_NAME}.so" "$INSTALL_DIR/"
    echo "✓ Installed to $INSTALL_DIR (system-wide)"
fi
