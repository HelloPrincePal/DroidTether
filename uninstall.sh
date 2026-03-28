#!/usr/bin/env bash
# DroidTether — One-line uninstaller for Apple Silicon macOS
# Usage: curl -sL https://raw.githubusercontent.com/HelloPrincePal/DroidTether/main/uninstall.sh | bash

set -euo pipefail

echo "🗑 Uninstalling DroidTether..."

# 1. Background Service Cleanup
PLIST_NAME="com.princePal.droidtether.plist"
PLIST_DEST="/Library/LaunchDaemons/$PLIST_NAME"

if [[ -f "$PLIST_DEST" ]]; then
  echo "→ Stopping and removing background service"
  sudo launchctl bootout system "$PLIST_DEST" 2>/dev/null || true
  sudo rm -f "$PLIST_DEST"
fi

# 2. Binary Cleanup
if [[ -f "/usr/local/bin/droidtether" ]]; then
  echo "→ Removing binary from /usr/local/bin"
  sudo rm -f "/usr/local/bin/droidtether"
fi

# 3. Log Cleanup (Optional)
LOG_FILE="/var/log/droidtether.log"
if [[ -f "$LOG_FILE" ]]; then
  echo "→ Removing logs"
  sudo rm -f "$LOG_FILE"
fi

# 4. Config Cleanup (Ask for confirmation if needed, but defaults to keep for state?)
# Actually, uninstaller should usually clean up.
CONFIG_DIR="/etc/droidtether"
if [[ -d "$CONFIG_DIR" ]]; then
  echo "→ Removing configuration files"
  sudo rm -rf "$CONFIG_DIR"
fi

echo ""
echo "✓ DroidTether completely uninstalled."
echo "🔄 Wi-Fi behavior should return to normal."
