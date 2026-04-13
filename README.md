# 📱 DroidTether

**Seamless Android RNDIS USB tethering for Apple Silicon Macs.**
*No Kernel Extensions. No SIP Changes. No Reboots.* 🚀

![Downloads](https://img.shields.io/github/downloads/HelloPrincePal/DroidTether/total?style=for-the-badge&color=green)
![Status](https://img.shields.io/github/v/release/HelloPrincePal/DroidTether?style=for-the-badge&color=blue)

DroidTether is a lightweight userspace daemon that brings high-performance USB tethering to macOS by implementing the RNDIS protocol via `libusb` and routing traffic through the native `utun` interface.

---

## ✨ Why DroidTether?

- 🔒 **Zero System Security Changes**: Unlike HoRNDIS, DroidTether runs entirely in userspace. You don't need to disable System Integrity Protection (SIP) or allow reduced security mode.
- ⚡ **Apple Silicon Native**: Built from the ground up for M1, M2, M3, and M4 Macs.
- 🤖 **Samsung Friendly**: Includes a specialized workaround for Samsung's dynamic MAC address randomization on tethering interfaces.
- 🔌 **Plug & Play**: Automatically detects your phone, performs the handshake, and configures your Mac's routing/DNS instantly.

---

## 🚀 macOS 15+ (Tahoe) Compatibility Guide

macOS 15 introduces a strict "System Trust" model for network interfaces. DroidTether operates within these boundaries, leading to a split-networking experience:

### ✅ Works Out-of-the-Box (Independent Apps)
Apps that use their own internal network or DNS libraries (DNS-over-HTTPS) bypass the OS "reachability" checks and work at full speed instantly:
*   **Browsers**: Chrome, Firefox, Brave, Microsoft Edge.
*   **Meetings**: Google Meet, Zoom, Slack, Microsoft Teams.
*   **Streaming**: Netflix, YouTube, Spotify, Twitch.
*   **Developer Tools**: Any connection to a raw IP address.

### ⚠️ CLI & Native System Apps
Native Apple services—such as **Safari**, the **App Store**, **Apple Music**, and native **System Updates**—as well as CLI tools (`curl`, `git`, `brew`), strictly follow the `mDNSResponder` "reachability" flag. 

#### **The Solution (Native DNS Override)**
To enable full system-wide resolution for CLI tools and Safari, run this command while DroidTether is active:
```bash
# Force your hardware adapter to route DNS through the tunnel
sudo networksetup -setdnsservers Wi-Fi 8.8.8.8 8.8.4.4
```
*To revert back to automatic DNS when you stop using DroidTether:*
```bash
sudo networksetup -setdnsservers Wi-Fi empty
```

---

## 🛡️ Security & Privacy Posture

DroidTether is built on a strict **"local-only"** and **"least-privilege"** security model. We enforce extreme transparency because system-level applications require a high degree of trust.

### 🚫 Zero Telemetry & Data Sovereignty
- **No Data Inspection**: DroidTether simply routes encrypted and unencrypted packets between the macOS kernel (`utun`) and the Android USB interface (`libusb`). It **does not** read, inspect, or modify the contents of your web traffic.
- **No Analytics**: There is absolutely zero telemetry, tracking, or "call-home" functionality.
- **Local Logs Only**: Operational logs reside strictly on your local machine at `/var/log/droidtether.log`.

### 🔑 Why `sudo` (Root) is Required
To function without relying on Kernel Extensions, DroidTether requires elevated OS privileges:
1. **Virtual Interface Creation**: Creating the `utun` interface requires kernel routing permissions.
2. **Routing Table Modification**: Injecting routes to prioritize the Android connection requires root.
3. **Hardware USB Binding**: Opening raw protocol communication via `libusb` requires device-level access.

### 📂 100% Auditable Core
The core routing logic is written in modern Go and consists of fewer than **2,000 lines of code**, making it trivially auditable. Review our [Architecture Deep-Dive](docs/architecture.md) for more.

---

## 🛠️ Verified Test Environment

| Phone Name | Android Version | Host Name | OS Version | Results |
| :--- | :--- | :--- | :--- | :--- |
| Xiaomi 11 Lite NE | 14 (HyperOS 2.0) | MacBook Air M4 | macOS Tahoe | **260 Mbps** 🚀 |
| Samsung Galaxy S24 | 16 (One UI 8.0) | MacBook Air M4 | macOS Tahoe | **290 Mbps** 🚀 |
| Samsung Galaxy A55 | 16 (One UI 8.0) | MacBook Air M4 | macOS Tahoe | Stable ✅ |

---

## 📦 Installation & Setup

### 1. One-Liner Install
Open your terminal and paste this command to install the binary and start the background service:
```bash
curl -sL https://raw.githubusercontent.com/HelloPrincePal/DroidTether/main/install.sh | bash
```

### 2. Manual Build (From Source)
```bash
# Prerequisites: brew install libusb pkg-config
git clone https://github.com/HelloPrincePal/DroidTether
cd DroidTether
make build
sudo ./build/droidtether
```

---

## 📖 How to Use

1. **Connect** your Android phone to your Mac via a USB-C cable.
2. **Enable Tethering** on your phone:
   - Go to **Settings** ⚙️ → Search for **Tethering**
   - Toggle **USB Tethering** to **ON** ✅
3. **Enjoy!** DroidTether will log `✨ Network auto-configured!`.

---

## 🔍 Verifying Connectivity

### 1. Check the Service Status ⚙️
```bash
sudo launchctl list | grep princePal
# Expected: A process ID (number) should appear.
```

### 2. Check the Network Interface 📡
```bash
ifconfig | grep -A 5 utun
# Expected: A 'utun' interface with an 'inet' address (e.g., 10.x.x.x)
```

### 3. Verify the Routing 🛣️
```bash
route -n get google.com | grep interface
# Expected: interface: utunX
```

### 4. Monitor Live Traffic 📜
```bash
tail -f /var/log/droidtether.log
```

---

## ⚠️ A Note on Apple's `networkQuality`

If you try to run the `networkQuality` command, you may encounter an "offline" error. This is known macOS behavior where high-level system utilities sometimes only bind to physical hardware services (WiFi/Ethernet). Real-world tasks like Gaming and Video Calling are **completely unaffected**.

---

## 🤝 Community
*   **Contributing**: Found a bug? Please open an issue or submit a PR! Review our [Code of Conduct](CODE_OF_CONDUCT.md).
*   **Security**: Please report vulnerabilities via our [Security Policy](.github/SECURITY.md).
*   **License**: MIT — © PrincePal