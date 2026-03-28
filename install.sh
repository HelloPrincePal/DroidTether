#!/usr/bin/env bash
# DroidTether — One-line installer for Apple Silicon macOS
# Usage: curl -sL https://raw.githubusercontent.com/HelloPrincePal/DroidTether/main/install.sh | bash

set -euo pipefail

# 1. Platform Check
ARCH="$(uname -m)"
OS="$(uname -s)"

if [[ "$OS" != "Darwin" || "$ARCH" != "arm64" ]]; then
  echo "error: DroidTether currently only supports Apple Silicon macOS (arm64)." >&2
  exit 1
fi

echo "🚀 Installing DroidTether for Apple Silicon..."

# 2. Dependency Check (libusb)
if ! command -v brew >/dev/null 2>&1; then
  echo "error: Homebrew not found. Please install Homebrew first: https://brew.sh" >&2
  exit 1
fi

if ! brew list libusb >/dev/null 2>&1; then
  echo "→ Installing dependency: libusb"
  brew install libusb
fi

# 3. Fetch Latest Release from GitHub API
REPO="HelloPrincePal/DroidTether"
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [[ -z "$LATEST_RELEASE" ]]; then
  echo "error: could not find latest release for $REPO" >&2
  exit 1
fi

echo "→ Downloading DroidTether $LATEST_RELEASE"
URL="https://github.com/HelloPrincePal/DroidTether/releases/download/$LATEST_RELEASE/droidtether-darwin-arm64.tar.gz"
TMP_DIR=$(mktemp -d)
curl -sL "$URL" -o "$TMP_DIR/droidtether.tar.gz"

# 4. Extract and Install Binary
tar -xzf "$TMP_DIR/droidtether.tar.gz" -C "$TMP_DIR"

if [[ ! -f "$TMP_DIR/droidtether" ]]; then
  echo "error: Could not find droidtether binary in $LATEST_RELEASE archive." >&2
  ls -R "$TMP_DIR"
  exit 1
fi

echo "→ Installing binary to /usr/local/bin (requires password)"
sudo mkdir -p /usr/local/bin
sudo mv "$TMP_DIR/droidtether" /usr/local/bin/droidtether
sudo chmod +x /usr/local/bin/droidtether

# 5. Config Setup
CONFIG_DIR="/etc/droidtether"
if [[ ! -d "$CONFIG_DIR" ]]; then
  echo "→ Creating config directory at $CONFIG_DIR"
  sudo mkdir -p "$CONFIG_DIR"
fi

# Download default config if not exists
if [[ ! -f "$CONFIG_DIR/droidtether.toml" ]]; then
  echo "→ Installing default configuration"
  CONFIG_URL="https://raw.githubusercontent.com/HelloPrincePal/DroidTether/main/config/default.toml"
  sudo curl -sL "$CONFIG_URL" -o "$CONFIG_DIR/droidtether.toml"
fi

# 6. Launchd Service Setup
PLIST_NAME="com.princePal.droidtether.plist"
PLIST_DEST="/Library/LaunchDaemons/$PLIST_NAME"
PLIST_URL="https://raw.githubusercontent.com/HelloPrincePal/DroidTether/main/launchd/$PLIST_NAME"

echo "→ Installing background service (launchd)"
sudo curl -sL "$PLIST_URL" -o "$PLIST_DEST"
sudo chown root:wheel "$PLIST_DEST"
sudo chmod 644 "$PLIST_DEST"

# 7. Start the Service
echo "→ Starting DroidTether service"
sudo launchctl bootout system "$PLIST_DEST" 2>/dev/null || true
sudo launchctl bootstrap system "$PLIST_DEST"

# Cleanup
rm -rf "$TMP_DIR"

echo ""
echo "✨ DroidTether successfully installed!"
echo "📡 Plug in your Android phone and enable USB Tethering to begin."
echo "📜 Logs: tail -f /var/log/droidtether.log"
echo "❌ To uninstall: curl -sL https://raw.githubusercontent.com/HelloPrincePal/DroidTether/main/uninstall.sh | bash"
