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

## 🛡️ Transparency & Privacy
DroidTether is built on a "local-only" model.
- 📂 **100% Open Source**: Every line of code is available for audit in this repository.
- 🚫 **No Telemetry**: No tracking, no analytics, and no "call-home" features. 
- 🔒 **Local Connectivity**: All networking happens strictly between your Mac and your Android device. No external servers are involved in the packet relay process.
- 🕵️ **Log Privacy**: Logs reside only on your local machine at `/var/log/droidtether.log` for debugging purposes.

> 📝 **Audit Note**: The entire core logic of DroidTether is contained in less than **2,000 lines of Go code**, making it exceptionally easy to audit for security and transparency. We believe in simplicity and clear source code as the ultimate form of trust.

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

## 🔑 Why `sudo` is Required?
Because DroidTether operates at the system network level, it requires elevated privileges for specific operations:
1. **Network Interface Management**: Creating and configuring the virtual `utun` interface on macOS is a kernel-restricted task.
2. **Routing Table Injection**: Updating your Mac's routing table to prioritize the phone's internet connection requires superuser permissions.
3. **Log Management**: Writing operational logs to `/var/log/droidtether.log` for system-wide transparency.

*DroidTether performs these tasks purely in userspace—no persistent kernel extensions are installed.*

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