# DroidTether — File Structure

Complete annotated repository layout. Every file listed here should exist before vibe-coding begins.  
Files marked `[LLM]` are optimized for inclusion in AI context windows.

```
DroidTether/
│
├── README.md                        # Public-facing project README
├── LICENSE                          # MIT
├── PRIVACY.md                      # Privacy policy
├── CONTRIBUTING.md                  # How to contribute
├── install.sh                       # One-line installer
├── uninstall.sh                     # One-line uninstaller
├── Makefile                         # Build/Test commands
├── go.mod                           # Go dependencies
│
├── docs/
│   ├── PRD.md                       # [LLM] Full product requirements (this project's source of truth)
│   ├── FILE_STRUCTURE.md            # [LLM] This file — repo layout reference
│   ├── QUICK_REF.md                 # [LLM] One-page cheat sheet: RNDIS protocol, utun API, key consts
│   ├── LLM_GUIDE.md                 # [LLM] How to load context for AI coding sessions
│   ├── TESTING.md                   # [LLM] Testing strategy and how to run tests
│   └── ARCHITECTURE.md              # [LLM] Deep-dive diagrams: data flow, goroutine map, lifecycle
│
├── CHANGELOG.md                     # [LLM] Every small change — AI reads this for diff context
├── VERSIONS.md                      # Semantic version log — one entry per git push with date
│
├── cmd/
│   └── droidtether/
│       └── main.go                  # Entry point. Parses flags, starts daemon or runs CLI commands.
│
├── internal/
│   │
│   ├── daemon/
│   │   ├── README.md                # [LLM] What this package does, exported symbols, usage
│   │   ├── daemon.go                # Main run loop. Starts USB watcher. Manages session lifecycle.
│   │   ├── session.go               # One Session per attached phone. Owns utun + relay + DHCP.
│   │   └── daemon_test.go           # Unit tests using mock USB
│   │
│   ├── usb/
│   │   ├── README.md                # [LLM] USB package scope, VID/PID matching logic
│   │   ├── device.go                # Open/close libusb device. Claim RNDIS interface.
│   │   ├── vidpid.go                # Known Android VID/PID pairs + RNDIS class matcher
│   │   ├── watcher.go               # Polls libusb hotplug events. Emits attach/detach signals.
│   │   ├── mock.go                  # Mock USB device for unit testing
│   │   └── device_test.go
│   │
│   ├── rndis/
│   │   ├── README.md                # [LLM] RNDIS state machine, message types, OID reference
│   │   ├── rndis.go                 # RNDIS state machine: INIT → QUERY → SET → DATA
│   │   ├── messages.go              # Binary structs for every RNDIS message type
│   │   ├── oids.go                  # OID constants (MAC addr, packet filter, frame size, etc.)
│   │   ├── encode.go                # Marshal/unmarshal RNDIS binary frames
│   │   └── rndis_test.go            # Tests: encode/decode round-trips for every message type
│   │
│   ├── tun/
│   │   ├── README.md                # [LLM] utun creation via AF_SYSTEM, interface naming
│   │   ├── utun.go                  # Create/destroy utun interface via syscall
│   │   ├── relay.go                 # Bidirectional packet relay: USB bulk ↔ utun fd
│   │   └── tun_test.go
│   │
│   ├── dhcp/
│   │   ├── README.md                # [LLM] DORA sequence, what fields we need
│   │   ├── client.go                # DHCP DORA: Discover → Offer → Request → Ack
│   │   ├── packets.go               # DHCP packet encode/decode
│   │   └── dhcp_test.go
│   │
│   └── route/
│       ├── README.md                # [LLM] How we set/remove macOS routes via netlink-style syscalls
│       ├── route.go                 # Add/remove default route through utunN interface
│       └── route_test.go
│
├── config/
│   ├── default.toml                 # [LLM] Default config — commented TOML, human + AI readable
│   └── config.go                    # Loads and validates TOML config
│
├── scripts/
│   ├── test-live.sh                 # Integration test: requires real phone attached
│   ├── install-launchd.sh           # Local dev: install plist (brew prefix); Homebrew uses Formula post_install
│   ├── uninstall-launchd.sh         # Unloads + removes launchd plist
│   └── dev-reload.sh                # Kill daemon → rebuild → restart (hot-reload for dev)
│
├── launchd/
│   └── com.princePal.droidtether.plist  # launchd daemon config (auto-start, crash restart)
│
├── test/
│   ├── fixtures/
│   │   ├── rndis_init_cmplt.bin     # Captured RNDIS INITIALIZE_CMPLT packet for replay tests
│   │   ├── rndis_query_cmplt.bin    # Captured RNDIS QUERY_CMPLT (MAC address response)
│   │   └── dhcp_offer.bin           # Captured DHCP OFFER packet
│   └── integration/
│       └── tether_test.go           # End-to-end test (build tag: //go:build integration)
│
├── Formula/
│   └── droidtether.rb                 # Homebrew formula (references GitHub release tarball)
│
└── .github/
    └── workflows/
        ├── ci.yml                   # Run tests on every PR (go test ./...)
        └── release.yml              # On git tag push: build arm64 binary, create GitHub release
```

---

## Key File Roles (Quick Reference for LLM)

| File | When to Read It |
|------|----------------|
| `docs/PRD.md` | Starting a new feature or need full context |
| `docs/QUICK_REF.md` | Writing RNDIS or utun code |
| `CHANGELOG.md` | Before any code edit — understand what changed recently |
| `VERSIONS.md` | Need to know current version and milestone |
| `internal/rndis/README.md` | Working on RNDIS protocol code |
| `internal/usb/README.md` | Working on USB device detection/hotplug |
| `internal/tun/README.md` | Working on utun interface or packet relay |
| `config/default.toml` | Changing config schema |
| `docs/TESTING.md` | Writing tests or running the test suite |

---

## Package Dependency Graph

```
cmd/droidtether/main.go
        │
        ▼
internal/daemon
    ├──▶ internal/usb       (device open, hotplug)
    ├──▶ internal/rndis     (protocol handshake)
    ├──▶ internal/tun       (utun + relay)
    ├──▶ internal/dhcp      (IP assignment)
    └──▶ internal/route     (default route)

config ◀── daemon (loaded at startup)
```

No circular dependencies. Each internal package is independently testable with mock interfaces.