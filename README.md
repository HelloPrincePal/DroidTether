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

## 🛡️ Security & Privacy Posture

DroidTether is built on a strict **"local-only"** and **"least-privilege"** security model. We understand that system-level network applications require a high degree of trust, which is why we enforce extreme transparency.

### 🚫 Zero Telemetry & Data Sovereignty
- **No Data Inspection**: DroidTether simply routes encrypted and unencrypted packets between the macOS kernel (`utun`) and the Android USB interface (`libusb`). It **does not** read, inspect, or modify the contents of your web traffic.
- **No Analytics**: There is absolutely zero telemetry, tracking, or "call-home" functionality built into this daemon.
- **Local Logs Only**: Operational logs reside strictly on your local machine at `/var/log/droidtether.log` for debugging purposes and are never transmitted anywhere.

### 🔑 Why `sudo` (Root) is Required
To function without relying on deprecated Kernel Extensions, DroidTether operates natively in userspace but requires elevated OS privileges to bind to the network stack:
1. **Virtual Interface Creation**: Creating the `utun` network interface requires macOS kernel routing permissions.
2. **Routing Table Modification**: Injecting routes to prioritize your Android phone's internet connection requires superuser access.
3. **Hardware USB Binding**: Opening raw protocol communication via `libusb` requires device-level access.

*Note: DroidTether performs these tasks purely in userspace without modifying System Integrity Protection (SIP) or demanding reduced security boot modes.*

### 📂 100% Auditable Core
The entire core routing logic is written in modern Go and consists of fewer than **2,000 lines of code**. We believe in simplicity and auditable code as the ultimate form of security. Review our [Security Policy](.github/SECURITY.md) for vulnerability reporting.

---

## 🛠️ Verified Test Environment

| Phone Name | Android Version | Host Name | OS Version | Results |
| :--- | :--- | :--- | :--- | :--- |
| Samsung Galaxy S24 | 16 (One UI 8.0) | MacBook Air M4 | macOS Tahoe | **290 Mbps** 🚀 |
| Samsung Galaxy A55 | 16 (One UI 8.0) | MacBook Air M4 | macOS Tahoe | Stable ✅ |

*Verified with full bidirectional traffic and global DNS resolution.*

---

## 🚀 Quick Install (Apple Silicon only)

Open your terminal and paste this one-liner to install DroidTether and start the background service:

```bash
curl -sL https://raw.githubusercontent.com/HelloPrincePal/DroidTether/main/install.sh | bash
```

### 🗑️ Uninstall
To completely remove DroidTether, its configuration, and the background service:

```bash
curl -sL https://raw.githubusercontent.com/HelloPrincePal/DroidTether/main/uninstall.sh | bash
```

---

## 🏗️ Developer Build (Source)

If you prefer to build from source:

### 1. Prerequisites
Ensure you have the following installed:
- [Go](https://go.dev/dl/) (1.21+)
- `libusb` and `pkg-config` (Install via `brew install libusb pkg-config`)

### 2. Build from Source
```bash
git clone https://github.com/HelloPrincePal/DroidTether
cd DroidTether
make build
```

### 3. Run Manually
```bash
sudo ./build/droidtether
```

---

## 📖 How to Use

1. **Connect** your Android phone to your Mac via a USB-C cable.
2. **Enable Tethering** on your phone:
   - Go to **Settings** ⚙️
   - Search for **Tethering**
   - Toggle **USB Tethering** to **ON** ✅
3. **Enjoy!** DroidTether will log `✨ Network auto-configured!`. Your Mac is now using your phone's internet.

---

## 🔍 Verifying Connectivity

Once DroidTether is installed and your phone is connected, you can run these commands to verify the bridge is active:

### 1. Check the Service Status ⚙️
```bash
sudo launchctl list | grep princePal
# Expected: A process ID (number) should appear, e.g., "67337  0  com.princePal.droidtether"
```

### 2. Check the Network Interface 📡
```bash
ifconfig | grep -A 5 utun
# Expected: You should see a 'utun' interface with an 'inet' address (e.g., 10.x.x.x)
```

### 3. Verify the Routing 🛣️
```bash
route -n get google.com | grep interface
# Expected: interface: utunX (where X is your DroidTether interface number)
```

### 4. Monitor Live Traffic 📜
```bash
tail -f /var/log/droidtether.log
# Expected: "🚀 Traffic Monitor" logs showing received/sent data totals.
```

### 5. Performance & Quality Test ⚡
```bash
ping -c 10 8.8.8.8
# Expected: 0% packet loss and stable round-trip times.
# Note: If latency is high, try switching your phone from 5G to 4G/LTE for better stability.
```

---

## 🛑 Stopping & Uninstalling

### Stop the Background Service
If you want to stop the service momentarily without uninstalling:
```bash
sudo launchctl bootout system /Library/LaunchDaemons/com.princePal.droidtether.plist
```

### Complete Uninstall
To completely remove DroidTether, its configuration, and the background service:
```bash
curl -sL https://raw.githubusercontent.com/HelloPrincePal/DroidTether/main/uninstall.sh | bash
```

---

## ⚠️ A Note on Apple's `networkQuality`

If you try to run the `networkQuality` command while using DroidTether, you may encounter an "offline" error. This is a known macOS behavior where high-level system utilities sometimes only bind to physical hardware services (WiFi/Ethernet).

**Don't worry!** Real-world high-performance tasks like **Gaming, 4K Streaming, and Video Calling** use the underlying data plane and are **completely unaffected**. For accurate benchmarks, we recommend using `ping 8.8.8.8` or [fast.com](https://fast.com) in your browser.

---

## 👤 Connect with the Author

Feel free to reach out or follow the project’s journey! 🚀

🔗 **LinkedIn**: [Prince Pal](https://www.linkedin.com/in/theprincepal/)  

---

## 📜 License
MIT — © PrincePal

## 🤝 Contributing
Found a bug? Have a feature request for v1.0? Please open an issue or submit a PR! 

**⚠️ Security**: For security vulnerabilities, please refer to our [Security Policy](.github/SECURITY.md) for private reporting instructions.