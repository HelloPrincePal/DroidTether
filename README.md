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

## 🛠️ Verified Test Environment
This project has been rigorously tested and confirmed working on:
- **Phone**: Samsung Galaxy A55 📱
- **Host**: MacBook Air M4 (Apple Silicon) 💻
- **OS**: macOS Tahoe 26.3.1(a) ⛰️
- **Android Support**: Android 16 (One UI 8.0) Verified ✅
- **Connectivity**: Full bidirectional traffic + Global DNS resolution 🌐

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

## 📜 License
MIT — © PrincePal

## 🤝 Contributing
Found a bug? Have a feature request for v1.0? Please open an issue or submit a PR!