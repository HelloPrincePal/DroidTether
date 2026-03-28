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

## [Unreleased]

*(Add entries here as you work. Move to a version block on each git push.)*

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