# 🛡️ Security Policy

## Security Philosophy
DroidTether operates natively in userspace and touches sensitive system network configuration aspects (specifically `utun` virtualization and OS-level routing tables). Because of this necessary privilege escalation (`sudo`), security and data sovereignty are our absolute top priorities.

We rely on a minimalist, auditable, Go-based codebase. We treat all security issues, especially those concerning the parsing of USB protocol packets (RNDIS) and arbitrary code execution vectors, with the highest urgency.

## Threat Model & Scope

We encourage security researchers to focus on the following high-priority vectors:
- **Packet Parsing Engine**: Buffer overflows, memory leaks, or execution vectors resulting from malicious RNDIS sequences passed back from the USB device.
- **Privilege Escalation**: Mechanisms that trick the daemon into writing arbitrary data outside of `/var/log/droidtether.log` or altering system states beyond the `utun` interface routing table.
- **Data Leakage**: Scenarios where local packets are improperly broadcast, logged in plaintext, or forwarded inappropriately.

### 🚫 Out of Scope
The following are universally out of scope for our security reporting:
- **Physical Device Compromise**: If an attacker already has physical access to your unlocked Mac or Android device.
- **Upstream Toolchain Vulnerabilities**: Issues stemming exclusively from core `libusb` or the Go compiler (these should be reported to their respective security teams).
- **Social Engineering / Phishing**: Tricking a user into downloading a malicious, unsigned version of DroidTether.
- **Denial of Service (DoS)**: Overloading the daemon by repeatedly plugging/unplugging devices is considered a stability issue, not a highly critical security failure unless it results in persistent kernel lockups.

## Supported Versions

DroidTether currently supports the following versions for security updates. We strongly recommend always running the latest signed version available on the Releases page.

| Version | Supported          |
| ------- | ------------------ |
| [Latest Release](https://github.com/HelloPrincePal/DroidTether/releases/latest) | :white_check_mark: |
| v0.8.x  | :white_check_mark: |
| < v0.8.0| :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.** 

If you discover a potential security vulnerability in DroidTether, please report it through one of the following channels:

### 1. GitHub Private Vulnerability Reporting (Preferred)
You can report vulnerabilities privately directly through the GitHub repository:
1. Go to the [Security tab](https://github.com/HelloPrincePal/DroidTether/security/advisories) of the repository.
2. Click on **Report a vulnerability**.
3. Provide a detailed summary of the vulnerability, including steps to reproduce.

### 2. Direct Contact
If you prefer, or do not have a GitHub account, you can reach out to the author privately via:
- 🔗 **LinkedIn**: [Prince Pal](https://www.linkedin.com/in/theprincepal/)
- 🤖 **Reddit**: [u/PrincePal_](https://www.reddit.com/user/PrincePal_/)

---

### What to Include in Your Report
To help us triage and respond to your report as quickly as possible, please include:
- A clear description of the vulnerability and its potential systemic impact.
- Step-by-step instructions to reproduce the issue (including Proof-of-Concept code or malicious RNDIS packet captures, if possible).
- The exact version of DroidTether, macOS build, and Android device being used.
- Any relevant crash logs or stack traces (e.g., from `/var/log/droidtether.log` or `/Library/Logs/DiagnosticReports/`).

### Our Commitment
- **Response Time**: We will acknowledge the receipt of your report within **48 hours**.
- **Transparency**: We will keep you informed of our progress as we investigate, patch, and deploy a fix.
- **Recognition**: We will gladly provide public credit/attribution in our Release Notes and Security Advisories for your discovery (if you wish) once the vulnerability has been patched.
*(Note: As an open-source hobby project, we currently do not offer financial bug bounties).*

---

Thank you for helping keep DroidTether and our users safe! 🚀
