# 🛡️ Privacy Policy

**DroidTether is built with privacy and security as core principles.**

This project is a userspace network daemon that handles local network traffic between an Android device and a macOS host. 

### 1. No Data Collection
DroidTether **does not collect, transmit, or share** any of your personal data, browser history, or network traffic with any third-party servers. All data processing happens entirely on your local machine.

### 2. Local Logging
DroidTether logs its activity to a local file at `/var/log/droidtether.log` to help with debugging and troubleshooting. These logs remain only on your computer and are never uploaded or shared. 

### 3. Network Traffic Transparency
DroidTether operates as a Layer 3 (IP) relay. It does not inspect the contents of your packets (e.g., your passwords or credit card numbers) and respects any end-to-end encryption (HTTPS, TLS, VPN) that your applications already use.

### 4. Telemetry
DroidTether has **no telemetry, no analytics, and no "call-home" features**. It will never connect to the internet unless you (the user) are using it to route your network traffic through your phone's connection.

---

### Questions?
If you have any questions about this Privacy Policy or how DroidTether handles your data, please open an issue in the [GitHub repository](https://github.com/HelloPrincePal/DroidTether).
