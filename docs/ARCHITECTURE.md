# DroidTether — Architecture

Deep-dive into data flow, goroutine map, and lifecycle.

---

## Data Flow: USB → Mac Network Stack

```
┌─────────────────────────────────────────────────────────────────┐
│                         Android Phone                           │
│                                                                 │
│   [IP Stack] ──▶ [RNDIS Driver] ──▶ [USB RNDIS Interface]     │
└────────────────────────────┬────────────────────────────────────┘
                             │ USB cable (bulk endpoints)
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                         DroidTether Daemon                        │
│                                                                 │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────────┐  │
│  │  libusb  │    │ RNDIS Engine │    │    Packet Relay       │  │
│  │          │◀──▶│              │    │                       │  │
│  │ bulk IN  │    │ - INIT       │    │  goroutine A:         │  │
│  │ bulk OUT │    │ - QUERY      │    │  USB bulk IN          │  │
│  │ ctrl EP0 │    │ - SET        │    │  → strip RNDIS header │  │
│  └──────────┘    │ - data mode  │    │  → strip eth header   │  │
│                  └──────────────┘    │  → write to utun fd   │  │
│                                      │                       │  │
│                                      │  goroutine B:         │  │
│                                      │  read from utun fd    │  │
│                                      │  → add eth header     │  │
│                                      │  → add RNDIS header   │  │
│                                      │  → USB bulk OUT       │  │
│                                      └──────────────────────┘  │
│                                                │                │
│                                         utun fd (int)          │
└────────────────────────────────────────────────┼────────────────┘
                                                 │
┌────────────────────────────────────────────────▼────────────────┐
│                     macOS Kernel                                 │
│                                                                 │
│  utun3 interface (droidtether0)                                   │
│  IP: 192.168.42.129/24                                          │
│  Default route → utun3                                          │
│                                                                 │
│  [macOS Network Stack] ──▶ [Apps / Safari / curl / etc.]       │
└─────────────────────────────────────────────────────────────────┘
```

---

## Goroutine Map

At steady state (one phone connected), DroidTether runs these goroutines:

```
main goroutine
├── daemon.Run()
│   ├── usb.Watcher.Watch()        [goroutine — USB hotplug poll loop]
│   │
│   └── session.Start()            [one per attached phone]
│       ├── rndis.Handshake()      [blocking, completes before relay starts]
│       ├── dhcp.Acquire()         [blocking, completes before relay starts]
│       │
│       ├── relay.USBtoTUN()       [goroutine — USB bulk IN → utun write]
│       └── relay.TUNtoUSB()       [goroutine — utun read → USB bulk OUT]
│
└── signal handler goroutine       [catches SIGINT/SIGTERM, graceful shutdown]
```

Total goroutines with one phone: ~5.

---

## Session Lifecycle (State Machine)

```
         ┌─────────┐
         │  IDLE   │  Daemon running, no phone attached
         └────┬────┘
              │ USB device attached, RNDIS class matched
              ▼
         ┌─────────┐
         │CLAIMING │  DetachKernelDriver → ClaimInterface
         └────┬────┘
              │ success
              ▼
         ┌─────────────┐
         │ HANDSHAKING │  RNDIS INIT → QUERY (MAC) → SET (filter)
         └──────┬──────┘
                │ handshake complete
                ▼
         ┌──────────────┐
         │ CREATING TUN │  Open AF_SYSTEM socket → create utunN
         └──────┬───────┘
                │
                ▼
         ┌──────────────┐
         │  DHCP DORA   │  DISCOVER → OFFER → REQUEST → ACK
         └──────┬───────┘
                │ IP assigned
                ▼
         ┌──────────────┐
         │    ACTIVE    │  Relay goroutines running, internet flowing
         └──────┬───────┘
                │ USB detach OR signal OR error
                ▼
         ┌──────────────┐
         │  TEARING DOWN│  Stop relay → remove route → destroy utun → release USB
         └──────┬───────┘
                │
                ▼
         ┌─────────┐
         │  IDLE   │  Ready for next attach
         └─────────┘
```

---

## Packet Transform: USB → utun

RNDIS delivers Ethernet frames. utun expects raw IP packets. Transform:

```
USB bulk IN delivers:
┌──────────────────────────────────────────────────────┐
│ RNDIS PACKET_MSG header (44 bytes)                   │
├──────────────────────────────────────────────────────┤
│ Ethernet header (14 bytes: dst MAC + src MAC + type) │
├──────────────────────────────────────────────────────┤
│ IP payload (variable)                                │
└──────────────────────────────────────────────────────┘

After stripping:
┌──────────────────────────────────────────────────────┐
│ utun 4-byte header: [0x00, 0x00, 0x00, 0x02 (IPv4)] │
├──────────────────────────────────────────────────────┤
│ IP payload                                           │
└──────────────────────────────────────────────────────┘
```

utun → USB is the reverse: prepend Ethernet header + RNDIS header, remove utun header.

The 4-byte utun header values:
- `AF_INET`  (IPv4) = `0x00000002`
- `AF_INET6` (IPv6) = `0x0000001E`

Detect which from the IP version byte (first nibble of IP payload).

---

## Config Loading

```
startup
   │
   ├─ look for /etc/droidtether/droidtether.toml
   ├─ look for /usr/local/etc/droidtether/droidtether.toml  (Intel Homebrew)
   ├─ look for /opt/homebrew/etc/droidtether/droidtether.toml  (ARM Homebrew)
   └─ fall back to embedded defaults (config/default.toml compiled in)
```

CLI flag `--config /path/to/config.toml` overrides all.

---

## Logging

DroidTether uses zerolog with structured output. Default: human-readable text to `/var/log/droidtether.log`.

Key log fields:
```
level       — debug/info/warn/error
component   — usb | rndis | tun | dhcp | relay | daemon
session_id  — UUID per phone session
event       — attach | detach | handshake_ok | ip_assigned | relay_error | etc.
```

Set `format = "json"` in config for machine-readable logs (useful for log aggregators).

---

## Error Handling Philosophy

- Every function returns an `error` (no panics in library code)
- Errors wrap with context: `fmt.Errorf("rndis: INIT failed: %w", err)`
- The session state machine logs and retries transient errors (e.g., RNDIS timeout)
- Fatal errors (can't create utun, libusb not available) crash with clear message
- Relay errors cause session teardown and return to IDLE — daemon keeps running