#!/usr/bin/env bash
# scripts/dev-reload.sh
# Hot-reload loop for DroidTether development.
# Kills daemon → rebuilds → restarts in foreground → watches for changes.
# Requires: fswatch (brew install fswatch)

set -e

BINARY="./build/droidtether"
CONFIG="./config/default.toml"
WATCH_DIRS="./cmd ./internal ./config"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_deps() {
  if ! command -v fswatch &> /dev/null; then
    echo -e "${RED}✗ fswatch not found. Install with: brew install fswatch${NC}"
    exit 1
  fi
}

kill_daemon() {
  local pid
  pid=$(pgrep -f "droidtether" 2>/dev/null || true)
  if [ -n "$pid" ]; then
    echo -e "${YELLOW}→ Killing existing droidtether (PID $pid)${NC}"
    kill "$pid" 2>/dev/null || true
    sleep 0.5
  fi
  # Also unload launchd if loaded
  sudo launchctl unload /Library/LaunchDaemons/com.princePal.droidtether.plist 2>/dev/null || true
}

build() {
  echo -e "${YELLOW}→ Building...${NC}"
  if make build 2>&1; then
    echo -e "${GREEN}✓ Build OK${NC}"
    return 0
  else
    echo -e "${RED}✗ Build FAILED${NC}"
    return 1
  fi
}

run_daemon() {
  echo -e "${GREEN}→ Starting droidtether in foreground (Ctrl+C to stop)${NC}"
  echo -e "${YELLOW}   Logs: stdout (dev mode)${NC}"
  sudo "$BINARY" --config "$CONFIG" --log-level debug &
  DAEMON_PID=$!
  echo -e "${GREEN}→ Running as PID $DAEMON_PID${NC}"
}

check_deps
kill_daemon
build && run_daemon

echo ""
echo -e "${GREEN}→ Watching for file changes in: $WATCH_DIRS${NC}"
echo -e "${YELLOW}   Edit any .go or .toml file to trigger rebuild${NC}"
echo ""

fswatch -o $WATCH_DIRS | while read -r; do
  echo ""
  echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo -e "${YELLOW}→ Change detected — rebuilding...${NC}"
  kill_daemon
  if build; then
    run_daemon
  else
    echo -e "${RED}→ Waiting for fix before restarting...${NC}"
  fi
done