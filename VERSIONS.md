# DroidTether — Version Log

One entry per git push. Semantic versioning (MAJOR.MINOR.PATCH).
- PATCH: bug fix, refactor, docs
- MINOR: new feature, new package
- MAJOR: breaking API or behavior change (v1+ only)

Pre-release: all versions are v0.x.x until `brew install droidtether` works end-to-end.
v1.0.0 = MVP complete and working on M1/M2/M3.

---

## v0.8.1 — 2026-03-28
- Milestone: Legal & Transparency (The "Professional" Release)
- What works: Added `LICENSE` (MIT), `PRIVACY.md`, and `CONTRIBUTING.md`. Verified full compatibility with **Android 16 (One UI 8.0)**. The project is now legally documented and ready for public contributions.
- Next: Finalize Homebrew Formula logic for `brew install` support.

## v0.8.0 — 2026-03-28
- Milestone: One-Line Binary Distribution (Apple Silicon)
- What works: Automated GitHub Release pipeline for `darwin/arm64`. `install.sh` and `uninstall.sh` scripts enable one-liner setup via `curl`. Standardized system paths used for the background daemon (`/usr/local/bin` and `/etc/droidtether`). The project is now effectively "Plug & Play" for any M-series Mac user without requiring a local Go compiler.

## v0.7.2 — 2026-03-28
- Milestone: Release Automation Permissions Fix
- What works: Explicit `contents: write` permissions and `pkg-config` dependency added to GitHub Actions to allow successful CGO compilation and automated asset uploading.

## v0.7.1 — 2026-03-28
- Milestone: "One-Touch" Installation Scripts
- What works: Created `install.sh` at repository root. Added `--config` flag to `droidtether` binary. Standardized CLI to support global system configurations.

## v0.6.0 — 2026-03-28
- Milestone: Full network connectivity — DHCP + ARP + Ping working
- What works: Complete 4-step DHCP handshake (Discover→Offer→Request→ACK) with Android's randomized subnet. Dynamic ARP responder answers phone's ARP queries for our assigned IP. Samsung MAC randomization workaround auto-detects the phone's real current MAC from live traffic. Gratuitous ARP pre-populates phone's ARP cache. Bidirectional ICMP ping confirmed at ~5.9ms avg latency with 0% packet loss.
- What's broken: No default route injection yet (internet doesn't flow through phone). IP is hardcoded to DHCP-assigned values.
- Next: Default route injection, DNS forwarding, internet connectivity through phone.

---

## v0.5.0 — 2026-03-28
- Milestone: Packet Relay Engine
- What works: Bidirectional shuttle service over USB Bulk Endpoints. Reads IP packets from macOS `utun`, synthesizes Ethernet headers, wraps in RNDIS encapsulation, and pushes to Android. Strips RNDIS headers from Android replies and writes to `utun`.
- Next: IP assignment and DHCP handling.

---

## v0.4.0 — 2026-03-28
- Milestone: utun interface creation
- What works: Native macOS virtual interface creation (`utun`). The daemon now spawns a real network interface on the Mac when the phone connects.
- Next: Packet relay engine (Bulk transfer).

---

## v0.3.0 — 2026-03-28
- Milestone: RNDIS handshake working (INIT/QUERY/SET)
- What works: Raw USB Control transfers for RNDIS encapsulated commands; `INIT` handshake; `QUERY MAC` address retrieval; `SET` packet filter to enable promiscuous data mode. **Confirmed on real device.**
- What's broken: No actual data transfer (bulk transfer) implemented yet.
- Next: `internal/tun/utun.go` creation and bulk packet relay.

---

## v0.2.0 — 2026-03-28
- Milestone: USB device detection + hotplug watcher
- What works: `libusb` device monitoring, RNDIS hardware identification, and background daemon. **Validated on real Samsung Galaxy hardware.**
- What's broken: Handshake logic (requires `internal/rndis`); connection drops immediately after detection (expected).
- Next: Build `internal/rndis` handshake sequences (INIT/QUERY/SET).

---

## v0.1.2 — 2026-03-28
- Milestone: Internal structure + Config loader
- What works: Package layout with READMEs, `config/config.go`, Go dependencies in `go.mod`.
- Next: `internal/usb/vidpid.go` and `internal/usb/watcher.go`.

---

## v0.1.1 — 2026-03-28
- Milestone: Repo restructure + Go scaffold
- What works: documented layout (`docs/`, `config/`, `launchd/`, `Formula/`, `scripts/`, `.github/workflows/`), minimal `droidtether` binary (`--version`), `make build` / `go test ./...`
- What's broken: daemon behavior (USB/RNDIS/utun/DHCP not implemented)
- Next: `internal/usb`, `internal/rndis`, config loader in `config/config.go`

---

## v0.1.0 — 2025-03-28
- Milestone: Repository scaffold
- What works: docs, file structure, Makefile skeleton
- What's broken: everything (no code yet)
- Next: Set up go.mod, implement internal/usb/device.go, internal/usb/vidpid.go
- Author: PrincePal

---

## Roadmap Milestones

| Version | Milestone |
|---------|-----------|
| v0.1.0 | Repo scaffold, docs |
| v0.2.0 | USB device detection + hotplug watcher |
| v0.3.0 | RNDIS handshake working (INIT/QUERY/SET) |
| v0.4.0 | utun interface created, packet relay active |
| v0.5.0 | DHCP working, IP assigned |
| v0.6.0 | Default route injected, internet flows |
| v0.7.0 | Daemon (launchd) auto-start working |
| v0.8.0 | Homebrew formula works locally |
| v0.9.0 | Full integration test passes on real Samsung device |
| v1.0.0 | Public release — `brew install droidtether` works |