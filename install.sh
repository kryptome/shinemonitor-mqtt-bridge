#!/usr/bin/env bash
# install.sh — One-command setup for ShineMonitor MQTT Bridge
# Usage: bash install.sh
#
# What this script does:
#   1. Checks that Go is installed (offers to install if missing)
#   2. Builds the binary
#   3. Copies everything to /opt/shinemonitor-mqtt-bridge
#   4. Installs logrotate config
#   5. Installs + enables the systemd service
#
# Author: Piyush Mishra

set -euo pipefail

INSTALL_DIR="/opt/shinemonitor-mqtt-bridge"
SERVICE_NAME="shinemonitor-mqtt-bridge"
LOG_FILE="/var/log/shinemonitor-mqtt-bridge.log"
GO_VERSION="1.23.0"

echo "========================================="
echo "  ShineMonitor MQTT Bridge — Installer"
echo "========================================="
echo ""

# --- 1. Check Go ---
if ! command -v go &> /dev/null; then
    echo "⚠️  Go is not installed."
    read -rp "   Do you want me to install it now? (y/N): " install_go
    if [[ "$install_go" =~ ^[Yy]$ ]]; then
        echo "📦 Installing Go $GO_VERSION ..."
        ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m)
        case "$ARCH" in
            amd64|x86_64) ARCH="amd64" ;;
            arm64|aarch64) ARCH="arm64" ;;
            armhf|armv7l) ARCH="armv6l" ;;
            *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
        esac
        curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" -o /tmp/go.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf /tmp/go.tar.gz
        rm /tmp/go.tar.gz
        export PATH="/usr/local/go/bin:$PATH"
        echo 'export PATH="/usr/local/go/bin:$PATH"' | sudo tee /etc/profile.d/go.sh > /dev/null
        echo "✅ Go installed: $(go version)"
    else
        echo "❌ Go is required to build. Install it from https://go.dev/dl/ and re-run this script."
        exit 1
    fi
else
    echo "✅ Go found: $(go version)"
fi

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

# --- 4. Setup log file ---
echo "📝 Setting up log file..."
sudo touch "$LOG_FILE"

# --- 5. Install logrotate config ---
echo "🔄 Installing logrotate config..."
sudo cp "$SERVICE_NAME.logrotate" /etc/logrotate.d/"$SERVICE_NAME"

# --- 6. Install systemd service ---
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
echo "  View logs:       tail -f $LOG_FILE"
echo "  REST API:        http://localhost:8080/now"
echo ""
