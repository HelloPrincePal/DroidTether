# DroidTether — Testing Guide

How to test DroidTether at every stage of development.

---

## Test Layers

### 1. Unit Tests (no hardware needed)
Run: `make test`  
Location: `internal/*/..._test.go`

Tests run entirely with mocks. No USB device, no utun (we don't create real interfaces in unit tests).

Key test files:
- `internal/rndis/rndis_test.go` — encode/decode every RNDIS message type using captured binary fixtures
- `internal/dhcp/dhcp_test.go` — DORA packet construction and parsing
- `internal/usb/device_test.go` — VID/PID matching logic, interface claiming against mock
- `internal/tun/tun_test.go` — packet framing (Ethernet→IP strip, 4-byte utun header handling)

### 2. Integration Tests (requires Apple Silicon Mac, no phone needed)
Run: `make test-integration`  
Build tag: `//go:build integration`  
Location: `test/integration/`

These tests:
- Create a real utun interface (requires root)
- Send/receive test packets through it
- Verify route injection and cleanup

Run as: `sudo go test -tags integration ./test/integration/`

### 3. Live End-to-End Test (requires phone)
Run: `make test-live`  
Script: `scripts/test-live.sh`

Prerequisites:
- Samsung or Android phone connected via USB
- USB Tethering enabled on the phone
- Running as root (or via sudo)

What the script checks:
1. `droidtether` binary finds the USB device
2. RNDIS handshake completes (check log output)
3. utun interface appears (`ifconfig | grep droidtether`)
4. DHCP assigns an IP
5. `ping -c 3 -I droidtether0 8.8.8.8` succeeds
6. Unplugging phone tears down cleanly

---

## Running Tests

```bash
# All unit tests
make test

# Unit tests with verbose output
make test-v

# With race detector
make test-race

# Integration tests (root required)
make test-integration

# Live end-to-end test (phone required, root required)
make test-live

# Run a single package
go test ./internal/rndis/...

# Run a single test function
go test ./internal/rndis/... -run TestInitializeMessage
```

---

## Development Hot-Reload

During active development, use the dev loop:

```bash
# Terminal 1: watch for changes, rebuild, restart daemon
make dev

# Terminal 2: tail the daemon log
tail -f /var/log/droidtether.log
```

`make dev` runs `scripts/dev-reload.sh` which:
1. Kills the running droidtether daemon (if any)
2. Rebuilds the binary
3. Restarts it in foreground mode (no launchd, logs to stdout)
4. Watches for file changes via `fswatch` and repeats

Install fswatch: `brew install fswatch`

---

## Test Fixtures

Binary packet captures in `test/fixtures/` are used in unit tests for RNDIS decode tests.

To capture new fixtures from a real session:
```bash
# Enable debug logging to capture raw packets
droidtether --log-level debug --dump-packets /tmp/droidtether-capture/

# Fixtures will be written as:
# /tmp/droidtether-capture/rndis_init_cmplt_001.bin
# /tmp/droidtether-capture/rndis_query_cmplt_001.bin
# etc.
```

Copy relevant captures to `test/fixtures/` and reference them in tests:
```go
data, _ := os.ReadFile("../../test/fixtures/rndis_init_cmplt.bin")
msg, err := rndis.DecodeInitCmplt(data)
```

---

## Mock USB Interface

`internal/usb/mock.go` provides `MockDevice` which implements the same `Device` interface as the real USB device. Use it in unit tests:

```go
mock := usb.NewMockDevice()
mock.QueueResponse(rndis.InitCmpltBytes(requestID))  // pre-load responses

engine := rndis.NewEngine(mock)
err := engine.Initialize()
assert.NoError(t, err)
```

---

## Checking the Daemon

```bash
# Is it running?
launchctl list | grep droidtether

# Check logs
tail -f /var/log/droidtether.log

# Manual start/stop (during development)
sudo launchctl load /Library/LaunchDaemons/com.princePal.droidtether.plist
sudo launchctl unload /Library/LaunchDaemons/com.princePal.droidtether.plist

# Force restart
sudo launchctl kickstart -k system/com.princePal.droidtether
```

---

## 🚀 Performance Benchmarks (Real-world)

Use these results as the "Gold Standard" to verify future optimizations or regressions.

| Device | Network | MAC/Host | OS | Result |
|--------|---------|----------|----|--------|
| Galaxy S24 | Airtel 5G | MacBook Air M4 | Android 16 | 290 Mbps (DL) / 41 Mbps (UL) |
| Galaxy A55 | Jio 5G | MacBook Air M4 | Android 16 | 180 Mbps (DL) / 25 Mbps (UL) |

*Tests performed via fast.com and `ping 8.8.8.8`. Zero packet loss verified.*

---

## What to Test When

| Stage | Run |
|-------|-----|
| After changing `internal/rndis/` | `go test ./internal/rndis/...` |
| After changing `internal/usb/` | `go test ./internal/usb/...` |
| Before every commit | `make test` |
| Before every git push | `make test && make test-race` |
| After reaching v0.5.0 | `make test-integration` |
| Before v1.0.0 release | `make test-live` with Samsung + Pixel device |