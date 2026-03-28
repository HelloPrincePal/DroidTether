# 📱 DroidTether

**Seamless Android RNDIS USB tethering for Apple Silicon Macs.**
*No Kernel Extensions. No SIP Changes. No Reboots.* 🚀

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
- **Host**: MacBook AI M4 (Apple Silicon) 💻
- **OS**: macOS Tahoe 26.3.2(a) ⛰️
- **Connectivity**: Full bidirectional traffic + DNS resolution 🌐

---

## 🏗️ Installation (Developer Build)

Since this project is in active development, please build from source:

### 1. Prerequisites
Ensure you have the following installed:
- [Go](https://go.dev/dl/) (1.21+)
- `libusb` (Install via `brew install libusb` if needed)
- `pkg-config`

### 2. Build from Source
```bash
git clone https://github.com/princePal/droidtether
cd droidtether
make build
```

### 3. Run the Daemon
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

While the daemon is running, you can verify everything is working in another terminal:

### 1. Check the Route 🛣️
```bash
route -n get google.com | grep interface
# Should return: interface: utunX
```

### 2. Test DNS 📡
```bash
nslookup google.com
# Should resolve via your phone's gateway (e.g., 10.215.9.141)
```

### 3. Clean Exit 🛑
Simply press **`Ctrl+C`** in the daemon window. DroidTether will instantly clean up your routes, reset your DNS, and close the virtual interface.

---

## 📜 License
MIT — © PrincePal

## 🤝 Contributing
Found a bug? Have a feature request for v1.0? Please open an issue or submit a PR!