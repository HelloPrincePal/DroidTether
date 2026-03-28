# DroidTether — LLM Coding Guide

This file tells you exactly what to load into your AI context window depending on what you're building.
Always read CHANGELOG.md first — it tells you what changed recently.

---

## Context Loading Recipes

### Starting a fresh session (any feature)
Load in this order:
1. `CHANGELOG.md` — recent changes (always first)
2. `VERSIONS.md` — where we are in development
3. `docs/PRD.md` — full product context
4. The relevant internal package `README.md`

### Working on RNDIS protocol
```
CHANGELOG.md
docs/QUICK_REF.md          ← RNDIS message structs and OID table
internal/rndis/README.md
internal/rndis/messages.go
internal/rndis/oids.go
```

### Working on USB hotplug / device detection
```
CHANGELOG.md
internal/usb/README.md
internal/usb/vidpid.go
internal/usb/watcher.go
```

### Working on utun / packet relay
```
CHANGELOG.md
docs/QUICK_REF.md          ← utun syscall reference
internal/tun/README.md
internal/tun/utun.go
internal/tun/relay.go
```

### Working on DHCP
```
CHANGELOG.md
internal/dhcp/README.md
internal/dhcp/packets.go
```

### Writing tests
```
CHANGELOG.md
docs/TESTING.md
internal/<package>/mock.go (if exists)
test/fixtures/ (binary captures)
```

### Debugging a live issue
```
CHANGELOG.md
VERSIONS.md
internal/daemon/daemon.go
internal/daemon/session.go
config/default.toml
```

---

## Prompting Conventions

When asking AI to write code for DroidTether, always include:

> "This is DroidTether — an Android USB tethering daemon for Apple Silicon Macs using libusb + utun. Read CHANGELOG.md first. Follow existing error handling patterns (explicit `error` returns, zerolog logging). Do not add new dependencies without noting it."

### Scope control phrases
- "Only modify `internal/rndis/`" — keeps AI from refactoring unrelated code
- "Match the style of `internal/usb/device.go`" — enforces consistency
- "Write tests using the mock in `internal/usb/mock.go`" — prevents real USB calls in tests

---

## Config Format: TOML (not JSON)

Config is TOML. Reasons:
- Comments are supported (JSON doesn't allow them)
- Keys don't need quotes — fewer tokens in LLM context
- Cleaner diffs in git

Example:
```toml
# droidtether default config
[usb]
poll_interval_ms = 500
claim_timeout_ms = 2000

[rndis]
max_transfer_size = 16384
init_timeout_ms = 3000

[tun]
interface_name = "droidtether0"
mtu = 1500

[dhcp]
timeout_ms = 5000
retry_count = 3

[logging]
level = "info"   # debug | info | warn | error
format = "text"  # text | json
```

---

## State Files: TOML (not JSON)

Any runtime state that gets persisted (e.g., session info, device registry) uses TOML.

Reason: When AI reads a state dump for debugging, TOML is ~30% fewer tokens than equivalent JSON and is self-commenting.

Example session state file (`/tmp/droidtether-session.toml`):
```toml
[session]
started_at = "2025-03-28T10:00:00Z"
phone_mac = "aa:bb:cc:dd:ee:ff"
utun_interface = "droidtether0"
assigned_ip = "192.168.42.129"
gateway_ip = "192.168.42.1"
status = "active"
```

---

## Hallucination Guard Rules

Things AI commonly gets wrong in this codebase — always verify:

1. **utun index** — macOS utun interfaces are `utun0`, `utun1`, etc. The index is returned by the kernel when you open the control socket. Do NOT hardcode.

2. **RNDIS message RequestID** — must be unique per message and echoed back in the completion. AI often forgets to increment or match it.

3. **USB bulk endpoint direction** — `0x81` is IN (device→host), `0x01` is OUT (host→device). AI often swaps these.

4. **RNDIS data messages are NOT the same as control messages** — data flows on a different endpoint pair than control. Read `docs/QUICK_REF.md`.

5. **libusb interface claiming requires detaching the kernel driver first** — call `DetachKernelDriver()` before `ClaimInterface()`. AI often omits this step.

6. **launchd plists require absolute paths** — no `~` expansion, no relative paths. Always use `/opt/homebrew/bin/droidtether`.

7. **utun is layer 3, not layer 2** — it handles IP packets, not Ethernet frames. RNDIS gives us Ethernet frames, so we must strip the Ethernet header before writing to utun.

---

## CHANGELOG Entry Format

When adding a CHANGELOG entry (always do this after any code change):

```markdown
## [Unreleased]

### YYYY-MM-DD HH:MM — <short description>
- What changed (1 sentence)
- Why it changed (1 sentence)  
- Files touched: `internal/rndis/rndis.go`, `internal/rndis/messages.go`
- Breaking: no | yes (describe what breaks)
```

Keep entries small. One logical change = one entry. AI uses these to understand recent context.

---

## VERSIONS Entry Format

On every `git push` (or `git tag`), add to VERSIONS.md:

```markdown
## v0.3.1 — 2025-03-28
- Milestone: RNDIS handshake working end-to-end
- What works: USB detect, INIT/QUERY/SET, data mode active
- What's broken: DHCP not yet started
- Next: implement DHCP DORA in internal/dhcp
```

Semantic versioning:
- `v0.x.x` — pre-release development
- `v1.0.0` — first `brew install` works end-to-end
- `v1.x.x` — stable with features
- Breaking changes bump major only after v1