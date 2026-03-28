# DroidTether

**Android USB tethering for Apple Silicon Macs — no kernel extension, no SIP changes.**

Works on macOS Ventura, Sonoma, and Sequoia on M1/M2/M3/M4 Macs.

---

## Install

```bash
brew tap princePal/droidtether
brew install droidtether
```

That's it. The daemon starts automatically at boot.

## Usage

1. Connect your Android phone via USB
2. On your phone: **Settings → Network → Hotspot & Tethering → USB Tethering → ON**
3. Internet routes through your phone. Done.

## How it works

DroidTether is a userspace daemon that:
1. Detects Android RNDIS USB interfaces via **libusb** (no kernel driver required)
2. Performs the RNDIS handshake to put the device into data mode
3. Creates a macOS **utun** interface (the same mechanism VPNs use — built into macOS)
4. Bridges packets between USB and utun in both directions
5. Requests an IP via DHCP from the phone's built-in DHCP server
6. Injects a default route so traffic flows through the phone

No kernel extensions. No SIP changes. No reboots. Works on every M-series Mac.

## Uninstall

```bash
brew uninstall droidtether
```

## Config

```bash
# Edit config
nano /opt/homebrew/etc/droidtether/droidtether.toml

# View logs
tail -f /var/log/droidtether.log
```

## Supported Devices

Works with any Android phone that supports USB Tethering, including:
- Samsung Galaxy (all models)
- Google Pixel
- OnePlus
- Xiaomi
- Any device matching USB RNDIS class (0xE0/0x01/0x03)

## Background: Why does this exist?

The classic solution was [HoRNDIS](https://github.com/jwise/HoRNDIS) — a kernel extension that hasn't been updated since 2018 and doesn't work on Apple Silicon. Apple's DriverKit replacement requires special entitlements that take months to obtain.

DroidTether takes a different path: implement the RNDIS protocol entirely in userspace using libusb and route packets through macOS's built-in utun interface. No kernel involvement, no entitlements needed.

## License

MIT — © PrincePal

## Contributing

PRs welcome. Read [docs/PRD.md](docs/PRD.md) for the full product spec and [docs/LLM_GUIDE.md](docs/LLM_GUIDE.md) for the development guide.

```bash
git clone https://github.com/princePal/droidtether
cd droidtether
make build
make test
```