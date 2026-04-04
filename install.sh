#!/usr/bin/env bash
# install.sh — One-command setup for ShineMonitor MQTT Bridge
# Usage: bash install.sh
#
# What this script does:
#   1. Checks that Go is installed
#   2. Builds the binary
#   3. Copies everything to /opt/shinemonitor-mqtt-bridge
#   4. Installs + enables the systemd service
#
# Author: Piyush Mishra

set -euo pipefail

INSTALL_DIR="/opt/shinemonitor-mqtt-bridge"
SERVICE_NAME="shinemonitor-mqtt-bridge"

echo "========================================="
echo "  ShineMonitor MQTT Bridge — Installer"
echo "========================================="
echo ""

# --- 1. Check Go ---
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed."
    echo "   Install it from https://go.dev/dl/ and re-run this script."
    exit 1
fi
echo "✅ Go found: $(go version)"

# --- 2. Build ---
echo "🔨 Building binary..."
CGO_ENABLED=0 go build -o "$SERVICE_NAME" main.go
echo "✅ Build complete."

# --- 3. Install to /opt ---
echo "📁 Installing to $INSTALL_DIR ..."
sudo mkdir -p "$INSTALL_DIR"
sudo cp "$SERVICE_NAME" "$INSTALL_DIR/"

# Copy .env if it exists, but don't overwrite an existing one
if [ -f .env ]; then
    if [ ! -f "$INSTALL_DIR/.env" ]; then
        sudo cp .env "$INSTALL_DIR/.env"
        echo "   Copied .env to $INSTALL_DIR"
    else
        echo "   .env already exists in $INSTALL_DIR — skipping (won't overwrite your config)"
    fi
else
    echo "⚠️  No .env file found. Copy .env.example to .env, fill it in, then copy to $INSTALL_DIR"
fi

# --- 4. Install systemd service ---
echo "⚙️  Installing systemd service..."
sudo cp "$SERVICE_NAME.service" /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable "$SERVICE_NAME"
sudo systemctl start "$SERVICE_NAME"

echo ""
echo "========================================="
echo "  ✅ Installation complete!"
echo "========================================="
echo ""
echo "  Service status:  sudo systemctl status $SERVICE_NAME"
echo "  View logs:       sudo journalctl -u $SERVICE_NAME -f"
echo "  REST API:        http://localhost:8080/now"
echo ""
