#!/usr/bin/env bash
# scripts/test-live.sh
# End-to-end live test. Requires:
#   - Android phone connected via USB
#   - USB Tethering enabled on phone
#   - Running as root (or via sudo)
#   - droidtether binary built: make build

set -e

BINARY="./build/droidtether"
CONFIG="./config/default.toml"
INTERFACE="droidtether0"
TEST_HOST="8.8.8.8"
PASS=0
FAIL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}✓ PASS${NC}: $1"; ((PASS++)); }
fail() { echo -e "${RED}✗ FAIL${NC}: $1"; ((FAIL++)); }
info() { echo -e "${YELLOW}→${NC} $1"; }

echo ""
echo "DroidTether — Live Integration Test"
echo "=================================="
echo ""

# ── Pre-flight ──────────────────────────────────────────────────

info "Pre-flight checks..."

if [ ! -f "$BINARY" ]; then
  fail "Binary not found at $BINARY — run: make build"
  exit 1
fi
pass "Binary exists"

if [ "$(id -u)" -ne 0 ]; then
  fail "Must run as root (sudo make test-live)"
  exit 1
fi
pass "Running as root"

# ── Start DroidTether ──────────────────────────────────────────────

info "Starting droidtether..."
"$BINARY" --config "$CONFIG" --log-level debug &
PROXPID=$!
info "droidtether PID: $PROXPID"
sleep 3  # Give it time to detect device and complete handshake

# ── Check USB Detection ──────────────────────────────────────────

info "Checking USB device detection..."
if "$BINARY" --config "$CONFIG" list-devices 2>&1 | grep -q "rndis"; then
  pass "RNDIS device detected"
else
  fail "No RNDIS device found — is USB Tethering ON?"
fi

# ── Check utun Interface ─────────────────────────────────────────

info "Checking utun interface creation..."
if ifconfig | grep -q "$INTERFACE"; then
  pass "Interface $INTERFACE exists"
  IP=$(ifconfig "$INTERFACE" | grep "inet " | awk '{print $2}')
  info "  Assigned IP: $IP"
else
  fail "Interface $INTERFACE not found"
fi

# ── Check IP Assignment ──────────────────────────────────────────

info "Checking IP assignment..."
if [ -n "$IP" ]; then
  pass "IP address assigned: $IP"
else
  fail "No IP address on $INTERFACE"
fi

# ── Check Route ──────────────────────────────────────────────────

info "Checking default route..."
ROUTE=$(netstat -rn | grep "^default" | grep "$INTERFACE" || true)
if [ -n "$ROUTE" ]; then
  pass "Default route via $INTERFACE"
else
  fail "Default route not pointing to $INTERFACE"
fi

# ── Ping Test ────────────────────────────────────────────────────

info "Pinging $TEST_HOST via $INTERFACE..."
if ping -c 3 -I "$INTERFACE" "$TEST_HOST" > /dev/null 2>&1; then
  pass "ping $TEST_HOST succeeded"
else
  fail "ping $TEST_HOST failed"
fi

# ── HTTP Test ────────────────────────────────────────────────────

info "HTTP connectivity test..."
if curl -s --interface "$INTERFACE" --max-time 10 https://httpbin.org/ip > /dev/null; then
  pass "HTTP request via $INTERFACE succeeded"
else
  fail "HTTP request failed"
fi

# ── Teardown Test ────────────────────────────────────────────────

info "Testing graceful teardown..."
kill $PROXPID
sleep 2

if ! ifconfig | grep -q "$INTERFACE"; then
  pass "Interface $INTERFACE removed on daemon stop"
else
  fail "Interface $INTERFACE still exists after daemon stop"
fi

# ── Summary ──────────────────────────────────────────────────────

echo ""
echo "=================================="
echo -e "Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}"
echo ""

if [ "$FAIL" -eq 0 ]; then
  echo -e "${GREEN}All tests passed! ✓${NC}"
  exit 0
else
  echo -e "${RED}$FAIL test(s) failed.${NC}"
  exit 1
fi