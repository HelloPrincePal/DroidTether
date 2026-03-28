#!/usr/bin/env bash
# Install launchd plist for local development (not Homebrew).
# Uses HOMEBREW_PREFIX when set, otherwise prefers $(brew --prefix) if brew exists.

set -euo pipefail

PLIST_NAME="com.princePal.droidtether.plist"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SRC_PLIST="$REPO_ROOT/launchd/$PLIST_NAME"
DEST="/Library/LaunchDaemons/$PLIST_NAME"

if [[ ! -f "$SRC_PLIST" ]]; then
  echo "error: missing $SRC_PLIST" >&2
  exit 1
fi

if [[ -n "${HOMEBREW_PREFIX:-}" ]]; then
  PREFIX="$HOMEBREW_PREFIX"
elif command -v brew >/dev/null 2>&1; then
  PREFIX="$(brew --prefix)"
else
  PREFIX="/opt/homebrew"
  echo "warning: brew not found; using PREFIX=$PREFIX (set HOMEBREW_PREFIX to override)" >&2
fi

TMP="$(mktemp)"
sed -e "s|/opt/homebrew|$PREFIX|g" "$SRC_PLIST" >"$TMP"

echo "→ Installing $DEST (PREFIX=$PREFIX)"
sudo cp "$TMP" "$DEST"
sudo chown root:wheel "$DEST"
sudo chmod 644 "$DEST"
rm -f "$TMP"

echo "→ Loading daemon"
sudo launchctl bootout system "$DEST" 2>/dev/null || true
sudo launchctl bootstrap system "$DEST"

echo "✓ launchd installed and loaded"
