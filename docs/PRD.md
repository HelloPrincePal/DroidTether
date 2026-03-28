# DroidTether — Product Requirements Document

**Project:** DroidTether  
**Author:** PrincePal  
**License:** MIT (Open Source)  
**Approach:** Option B — Userspace libusb + utun bridge  
**Target:** macOS 13+ on Apple Silicon (M1/M2/M3/M4)  
**Install:** `brew install princePal/tap/droidtether`

---

## 1. Problem Statement

Android USB tethering on Apple Silicon Macs is broken. The only working driver (HoRNDIS) is a kernel extension (kext) that:

- Was last released in 2018 (jwise/HoRNDIS)
- Requires disabling SIP (System Integrity Protection)
- Does not work on macOS Ventura+ with Apple Silicon
- Apple's DriverKit replacement requires paid entitlements that take months to get

There is no zero-friction, open-source, Homebrew-installable solution today.

---

## 2. Goal

A single `brew install` command that makes Android USB tethering work on Apple Silicon Macs, permanently, with zero SIP changes, no reboots, and no kernel extensions.

---

## 3. Non-Goals

- iOS tethering (handled natively by macOS)
- Wi-Fi tethering (already works natively)
- Windows or Linux support
- Bluetooth tethering
- A GUI app

---

## 4. Technical Approach — Userspace RNDIS Bridge

### Why this works without a kernel driver

Android exposes a USB interface using Microsoft's **RNDIS protocol** (Remote NDIS) — USB class `0xE0`, subclass `0x01`, protocol `0x03`. macOS doesn't have a native RNDIS driver, so it ignores this interface entirely.

DroidTether bypasses the kernel entirely using:

1. **libusb** — opens the raw USB device from userspace (no kernel driver needed)
2. **RNDIS protocol** — implemented fully in Go; handles the initialize/query/set/data message exchange
3. **utun** — macOS userspace TUN interface (built into macOS, no extra drivers); creates a virtual `utunN` network interface
4. **Packet relay** — goroutine pair bridges USB↔utun bidirectionally at full speed
5. **DHCP client** — requests IP from the phone's built-in DHCP server
6. **Route injection** — sets default route through `utunN` so all traffic flows via phone

### Protocol Stack

```
[Android Phone]
      │  USB (RNDIS over USB CDC)
      ▼
[libusb — userspace USB I/O]
      │  Raw RNDIS frames
      ▼
[RNDIS Engine — Go]
      │  Ethernet frames (stripped of RNDIS header)
      ▼
[utun interface — macOS kernel]
      │  IP packets
      ▼
[macOS Network Stack]
      │  Default route via utunN
      ▼
[Internet]
```

### Why utun instead of tun/tap

- `tun/tap` requires a kernel extension (TunTap driver) — defeats the purpose
- `utun` is built into macOS since 10.10, used by VPNs like WireGuard
- Available via `AF_SYSTEM / SYSPROTO_CONTROL` socket — no installation needed
- Works on all Apple Silicon Macs without any permissions beyond the daemon running as root

---

## 5. Architecture

### Component Map

```
droidtether (binary)
├── cmd/droidtether/main.go         — CLI entry point, arg parsing
├── internal/daemon/daemon.go     — Main loop, USB hotplug watcher
├── internal/usb/device.go        — libusb device open/close, interface claim
├── internal/rndis/rndis.go       — RNDIS protocol state machine
├── internal/rndis/messages.go    — RNDIS message structs (binary encoding)
├── internal/tun/utun.go          — utun interface create/destroy
├── internal/tun/relay.go         — Bidirectional packet relay goroutines
└── internal/dhcp/client.go       — Minimal DHCP client (DORA sequence)
```

### Daemon Lifecycle

```
start
  │
  ├─ watch for USB devices matching RNDIS VID/PID list
  │
  ├─ [device attached]
  │     ├─ claim USB interface
  │     ├─ RNDIS handshake (INIT → QUERY → SET → data mode)
  │     ├─ create utunN interface
  │     ├─ start relay goroutines (usb→tun, tun→usb)
  │     ├─ DHCP (get IP from phone)
  │     └─ inject default route
  │
  └─ [device detached]
        ├─ stop relay goroutines
        ├─ remove route
        └─ destroy utunN interface
```

---

## 6. Technology Stack

| Component | Technology | Reason |
|-----------|-----------|--------|
| Language | Go 1.22+ | Single binary, fast goroutines, arm64 native, easy libusb bindings |
| USB I/O | `google/gousb` (wraps libusb) | Battle-tested, CGo-based, works on macOS arm64 |
| TUN interface | Raw `AF_SYSTEM` syscalls | Built into macOS, no deps |
| DHCP | Custom minimal implementation | Only need DORA sequence; full DHCP lib is overkill |
| Logging | `rs/zerolog` | Structured, zero-alloc, easy to parse in tests |
| Config | TOML file | Human-readable, LLM-friendly (see §10) |
| Daemon mgmt | launchd plist | macOS native, auto-restart on crash |
| Distribution | Homebrew tap | Standard open-source macOS tooling |
| CI | GitHub Actions | Free for open source |

---

## 7. RNDIS Protocol Implementation

RNDIS messages are little-endian binary structs over USB bulk endpoints.

### Message Types Required

| Message | Direction | Purpose |
|---------|-----------|---------|
| `REMOTE_NDIS_INITIALIZE_MSG` | Host→Device | Negotiate version, max transfer size |
| `REMOTE_NDIS_INITIALIZE_CMPLT` | Device→Host | Confirm init, get device caps |
| `REMOTE_NDIS_QUERY_MSG` | Host→Device | Query OIDs (MAC addr, link speed, etc.) |
| `REMOTE_NDIS_QUERY_CMPLT` | Device→Host | OID response |
| `REMOTE_NDIS_SET_MSG` | Host→Device | Set packet filter (enable data flow) |
| `REMOTE_NDIS_SET_CMPLT` | Device→Host | Confirm set |
| `REMOTE_NDIS_PACKET_MSG` | Both | Actual Ethernet frame payload |

### Key OIDs to Query

- `OID_802_3_PERMANENT_ADDRESS` — device MAC address
- `OID_GEN_MAXIMUM_FRAME_SIZE` — max packet size
- `OID_GEN_CURRENT_PACKET_FILTER` — set to `NDIS_PACKET_TYPE_PROMISCUOUS` to start data flow

### Android RNDIS VID/PID Pairs

DroidTether ships with a curated list of known Android manufacturer VID/PIDs in `internal/usb/vidpid.go`. Unknown devices matching RNDIS class/subclass/protocol are also auto-detected.

---

## 8. Installation Design

### Homebrew Formula Flow

```bash
brew tap princePal/droidtether
brew install droidtether
```

This will:
1. Install the `droidtether` binary to `/usr/local/bin/` (Intel) or `/opt/homebrew/bin/` (Apple Silicon)
2. Install `libusb` as a dependency (Homebrew already has it)
3. Install the launchd plist to `/Library/LaunchDaemons/com.princePal.droidtether.plist`
4. Load the daemon (`sudo launchctl load ...`)
5. Print usage instructions

### Post-Install User Steps

```
1. Connect Android phone via USB
2. On phone: Settings → Network → Hotspot & Tethering → USB Tethering → ON
3. Done. Internet works.
```

### Uninstall

```bash
brew uninstall droidtether
```

Removes binary, plist, and unloads daemon.

---

## 9. File Structure

See `FILE_STRUCTURE.md` for the complete annotated tree.

---

## 10. LLM Optimization Strategy

DroidTether is designed to be vibe-coded efficiently. See `LLM_GUIDE.md` for full context-loading strategy.

Key decisions:
- **TOML** for all config and structured state (not JSON) — fewer tokens, no quotes on keys, comments supported
- **Flat function signatures** with explicit error returns — easier for LLMs to reason about
- **`CHANGELOG.md`** — every small change logged; AI reads this for diff context
- **`VERSIONS.md`** — semantic version + date per git push; AI knows exactly where in development things are
- **`QUICK_REF.md`** — one-page API and protocol cheat sheet; include in every LLM context window
- Each internal package has its own `README.md` — scope is tight, hallucination is reduced

---

## 11. Testing Strategy

See `TESTING.md` for the full testing approach.

Summary:
- **Unit tests** — RNDIS message encoding/decoding, DHCP packet parsing
- **Mock USB** — `internal/usb/mock.go` implements the same interface as the real USB device
- **Integration test script** — `scripts/test-live.sh` runs with a real phone attached
- **`make dev`** — hot-reload loop for development (kills daemon, rebuilds, restarts)

---

## 12. Success Criteria (MVP)

- [ ] `brew install droidtether` completes without errors on M1/M2/M3 Mac
- [ ] Plugging in a Samsung Android phone and enabling USB tethering gives internet in < 5 seconds
- [ ] `ping 8.8.8.8` works through the tethered connection
- [ ] Unplugging the phone gracefully tears down the interface with no kernel panic
- [ ] `brew uninstall droidtether` fully removes all traces
- [ ] Works on macOS Ventura (13), Sonoma (14), and Sequoia (15)

---

## 13. Out of Scope (v1)

- Multiple simultaneous tethered devices
- Traffic metrics / bandwidth display
- Preference for tethered vs Wi-Fi routing
- GUI status bar app (possible v2)
- NCM (USB CDC Network Control Model) support — newer Android versions may prefer this; track in backlog