# DroidTether — Changelog

Every small change is logged here. AI reads this file first to understand recent context.
One entry per logical change. Keep entries concise — 3 lines max each.

Format:
```
## YYYY-MM-DD HH:MM — <short description>
- What: one sentence
- Why: one sentence
- Files: list touched files
- Breaking: yes/no
```

---

---

## v0.8.1 — 2026-03-28

### 2026-03-28 20:04 — ⚖️ Legal and Transparency Documentation
- What: Added `LICENSE` (MIT), `PRIVACY.md`, and `CONTRIBUTING.md` to the root directory.
- Why: Establish trust by explicitly stating our "zero-data" privacy policy and defining open-source contribution guidelines.
- Files: `LICENSE`, `PRIVACY.md`, `CONTRIBUTING.md`
- Breaking: no ✅

## v0.7.2 — 2026-03-28

### 2026-03-28 19:24 — 🟢 Automated Distribution Support (Apple Silicon)
- What: Updated `release.yml` with `pkg-config` and `contents: write` permissions; verified successful GitHub Actions build and binary artifact generation.
- Why: Enable "one-liner" installation where GitHub automatically compiles and hosts the DroidTether binary for users.
- Files: `.github/workflows/release.yml`
- Breaking: no ✅

## v0.7.1 — 2026-03-28

### 2026-03-28 19:12 — 📦 One-Line Installer and CLI Improvements
- What: Created `install.sh` and `uninstall.sh` for `curl | bash` setup; added `--config` flag to `main.go`; standardized `launchd` path to `/usr/local/bin` and `/etc/droidtether`.
- Why: Transition from a "developer-only" build tool to a consumer-ready utility with frictionless setup/teardown.
- Files: `install.sh`, `uninstall.sh`, `cmd/droidtether/main.go`, `launchd/com.princePal.droidtether.plist`, `README.md`
- Breaking: yes (standardized system paths) 🚀

## v0.7.0 — 2026-03-28

### 2026-03-28 18:35 — 🌐 DNS Routing and MTU Stability Fixes
- What: Updated `scutil` script to use `SupplementalMatchDomains` for forced DNS resolution; lowered default `TUN.MTU` to 1400; added support for configuring MTU and filtered MacOS IPv6 broadcasts.
- Why: Pinging worked but DNS failed because macOS was skipping our `utun` for lookups. Web traffic failed due to USB fragmentation limits on Android exceeding 1400 MTU.
- Files: `internal/tun/utun_darwin.go`, `internal/tun/utun.go`, `internal/daemon/daemon.go`, `internal/daemon/relay.go`, `config/default.toml`
- Breaking: no ✅

### 2026-03-28 16:45 — Instant Teardown and Panic Fix
- What: Resolved a shutdown deadlock by force-closing USB handles when the daemon stops; added nil-pointer checks to the USB watcher to prevent exit panics.
- Why: Ensure the daemon exits immediately and cleanly on `Ctrl+C` even while waiting for network packets.
- Files: `internal/daemon/relay.go`, `internal/usb/watcher.go`, `internal/daemon/daemon.go`
- Breaking: no 🚀

### 2026-03-28 16:35 — 🌐 DNS Auto-Configuration and Graceful Shutdown
- What: Implemented `SetDNS` using `scutil` to automatically set phone gateway as system DNS; Added `sync.WaitGroup` to `Daemon` for robust session lifecycle management.
- Why: Provide a true "plug-and-play" experience with seamless internet and clean interface/route teardown on exit. Tested ON Samsung A55.
- Files: `internal/tun/utun.go`, `internal/tun/utun_darwin.go`, `internal/daemon/daemon.go`
- Breaking: no ✅

### 2026-03-28 16:25 — Default Route Injection on macOS
- What: Implemented `SetDefaultRoute` via override routes (0/1 & 128/1) in `internal/tun/utun_darwin.go`; auto-injects on successful DHCP if `set_default_route` is true; added cleanup for these routes on daemon exit.
- Why: Allow all system traffic to automatically flow through the tethered phone without manual routing commands.
- Files: `internal/tun/utun.go`, `internal/tun/utun_darwin.go`, `internal/daemon/daemon.go`
- Breaking: no

### 2026-03-28 16:24 — Update Config Schema for Route Control
- What: Expanded `Config` and `DHCPConfig` structs in `internal/config/config.go` to match the `default.toml` schema and support new routing features.
- Why: Ensure configuration values can be properly parsed and used by the daemon.
- Files: `internal/config/config.go`
- Breaking: no

*(Add entries here as you work. Move to a version block on each git push.)*

---

## v0.6.0 — 2026-03-28

### 2026-03-28 16:18 — Samsung MAC Randomization Workaround
- What: Auto-detect phone's real current MAC from received Ethernet traffic; also query `OID_802_3_CURRENT_ADDRESS` instead of `PERMANENT_ADDRESS`. Samsung devices use a randomized active MAC on the tethering interface that differs from the permanent address.
- Why: Ethernet frames addressed to the wrong MAC were silently dropped by the phone's NIC, causing 100% packet loss despite correct IP/DHCP configuration.
- Files: `internal/daemon/relay.go`, `internal/rndis/rndis.go`
- Breaking: no

### 2026-03-28 16:07 — Complete DHCP 4-Step Handshake
- What: Implemented full DHCP negotiation (Discover→Offer→Request→ACK) with proper option parsing (TLV format). Dynamic interface configuration from DHCP-assigned IP. Gratuitous ARP announcement after lease confirmation.
- Why: Android's `dnsmasq`/`iptables` requires a completed DHCP lease before accepting traffic from a client. Previous 2-step (Discover→Offer) left us as an unauthorized device.
- Files: `internal/daemon/relay.go`, `internal/daemon/daemon.go`
- Breaking: no

### 2026-03-28 15:35 — DHCP Auto-Discovery for Randomized Subnets
- What: Synthesized raw DHCP Discover packets over RNDIS to probe the phone's actual tethering subnet. Extracted `yiaddr` and server IP from DHCPOFFER to dynamically configure the macOS `utun` interface.
- Why: Android 11+ randomizes the USB tethering IP subnet on every connection. Hardcoded `192.168.42.x` addresses were wrong.
- Files: `internal/daemon/relay.go`, `internal/daemon/daemon.go`, `internal/tun/utun_darwin.go`
- Breaking: yes (removed hardcoded IP configuration)

### 2026-03-28 15:22 — ARP Responder and USB Stability
- What: Built a dynamic ARP responder that answers phone's ARP queries for our DHCP-assigned IP. Removed RNDIS KeepAlive (crashed macOS libusb), replaced with dummy Bulk OUT ARP keepalives. Made USB IN read errors non-fatal.
- Why: macOS `utun` is L3-only and cannot handle ARP. Without ARP responses, the phone couldn't route replies back to us. KeepAlive control transfers conflicted with concurrent Bulk IN reads on macOS.
- Files: `internal/daemon/relay.go`, `internal/daemon/daemon.go`
- Breaking: no

---

## v0.5.0 — 2026-03-28

### 2026-03-28 14:35 — Implement Bidirectional Packet Relay Engine
- What: Added `Relay` struct to bridge macOS `utun` packets and USB Bulk endpoints, complete with Ethernet and RNDIS encapsulation synthesis.
- Why: Milestone v0.5.0; this is the core engine that actually moves data between the phone and the Mac.
- Files: `internal/daemon/relay.go`, `internal/usb/device.go`, `internal/rndis/messages.go`
- Breaking: yes (switched to asynchronous relay loops)

---

## v0.4.0 — 2026-03-28

### 2026-03-28 14:00 — Implement utun creation on macOS
- What: Created `internal/tun/utun_darwin.go` to support spawning virtual network interfaces via `AF_SYSTEM` / `SYSPROTO_CONTROL`; integrated into daemon callback.
- Why: Milestone v0.4.0; enable the OS to talk to our daemon via a standard network handle.
- Files: `internal/tun/utun.go`, `internal/tun/utun_darwin.go`, `internal/daemon/daemon.go`
- Breaking: no

---

## v0.3.0 — 2026-03-28

### 2026-03-28 13:51 — Validated RNDIS Handshake on Live Device
- What: Confirmed RNDIS `INIT`, `QUERY(MAC)`, and `SET` handshake works; retrieved real device MAC address (`ee211e71ab6a`); refined `gousb` field access and control transfer reliability.
- Why: Complete Milestone v0.3.0; the device is now fully initialized and ready to send/receive network packets.
- Files: `internal/rndis/rndis.go`, `internal/usb/device.go`, `internal/daemon/daemon.go`
- Breaking: no

### 2026-03-28 13:50 — Implement RNDIS Handshake State Machine
- What: Built binary marshalling for `INIT/QUERY/SET` messages (`internal/rndis/messages.go`); implemented RNDIS state machine for `Handshake()` sequence; added `ControlCall` to `usb.Device`.
- Why: Milestone v0.3.0; allow the daemon to put the phone into data mode and retrieve device MAC address.
- Files: `internal/rndis/oids.go`, `internal/rndis/messages.go`, `internal/rndis/rndis.go`, `internal/usb/device.go`, `internal/daemon/daemon.go`
- Breaking: no

---

## v0.2.0 — 2026-03-28

### 2026-03-28 13:15 — Validated USB RNDIS Detection on Live Device
- What: Confirmed `Watcher` successfully detects Samsung Galaxy device; fixed `gousb` API mismatch in `device.go`; added fallback config loading for local dev; documented system dependencies (`libusb`, `pkg-config`).
- Why: Complete Milestone v0.2.0; ensure stable hardware discovery foundation before protocol implementation.
- Files: `internal/usb/device.go`, `cmd/droidtether/main.go`, `Makefile`
- Breaking: no

### 2026-03-28 13:00 — Implement USB RNDIS Detection Watcher
- What: Implemented `MatchRNDIS` utilizing known vid/class logic, built `Watcher` wrapper around `gousb`, integrated with `daemon.Run()` loop via `Device` struct wrapper.
- Why: Achieve Milestone v0.2.0 to be able to detect explicit Android devices natively via `libusb` and spawn callbacks.
- Files: `internal/usb/vidpid.go`, `internal/usb/watcher.go`, `internal/usb/device.go`, `internal/daemon/daemon.go`, `cmd/droidtether/main.go`
- Breaking: yes (replaced main loop with blocking daemon routine)

---

## v0.1.2 — 2026-03-28

### 2026-03-28 12:50 — Renamed project to DroidTether
- What: Renamed all occurrences of `ProxDroid` and `proxdroid` to `DroidTether` and `droidtether`; updated directory structure (`cmd/droidtether`), module path in `go.mod`, and macOS service definitions.
- Why: `ProxDroid` was already taken; **DroidTether** is a unique and descriptive replacement.
- Files: Global rename across all source, docs, and config files.
- Breaking: yes (package path change, binary rename)

### 2026-03-28 12:45 — Initial internal package structure and config loader
- What: Created READMEs for all `internal/` packages; added core dependencies to `go.mod`; implemented `config/config.go` loader.
- Why: Provide the foundation for technical implementation and enable functional configuration loading.
- Files: `internal/*/README.md`, `go.mod`, `config/config.go`
- Breaking: no

---

## v0.1.1 — 2026-03-28

### 2026-03-28 — Repository layout aligned with FILE_STRUCTURE
- What: Moved docs, `config/`, `launchd/`, `Formula/`, `.github/workflows/`, and `scripts/` into the documented layout; fixed root plist (was duplicate TOML); added local `install-launchd.sh` / `uninstall-launchd.sh`; added minimal `go.mod` and `cmd/droidtether` placeholder so `make build` / CI have an entrypoint.
- Why: Single coherent tree for development, packaging, and future Go packages under `internal/`.
- Files: `Makefile`, `docs/*`, `config/default.toml`, `launchd/`, `Formula/droidtether.rb`, `scripts/*`, `.github/workflows/*`, `go.mod`, `cmd/droidtether/main.go`, `CHANGELOG.md`, `VERSIONS.md`
- Breaking: no (paths only; update any out-of-repo bookmarks to `docs/` paths)

---

## v0.1.0 — 2025-03-28

### 2025-03-28 12:00 — Initial repository scaffold
- What: Created full directory structure, all placeholder files, docs
- Why: Project kickoff by PrincePal
- Files: All files in `docs/`, CHANGELOG.md, VERSIONS.md, Makefile, `docs/FILE_STRUCTURE.md`
- Breaking: no (initial commit)