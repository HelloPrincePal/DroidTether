#!/usr/bin/env bash
# Unload and remove launchd plist installed by install-launchd.sh or manual copy.

set -euo pipefail

PLIST_NAME="com.princePal.droidtether.plist"
DEST="/Library/LaunchDaemons/$PLIST_NAME"

if [[ -f "$DEST" ]]; then
  echo "→ Unloading $DEST"
  sudo launchctl bootout system "$DEST" 2>/dev/null || sudo launchctl unload "$DEST" 2>/dev/null || true
  sudo rm -f "$DEST"
  echo "✓ Removed $DEST"
else
  echo "→ No plist at $DEST (nothing to do)"
fi
