# 🤝 Contributing to DroidTether

We're so glad you're interested in helping DroidTether! This project is built by the community, for the community. Here's a quick guide into how you can contribute:

### 1. Report a Bug 🐛
If you find a bug, please [open an issue](https://github.com/HelloPrincePal/DroidTether/issues) with as much information as possible:
*   Your macOS version and device model.
*   Your Android device model and OS version.
*   Relevant logs from `/var/log/droidtether.log`.

**⚠️ Important**: For security vulnerabilities, please do **NOT** open a public issue. See our [Security Policy](.github/SECURITY.md) for private reporting instructions.

### 2. Suggest a Feature ✨
Have an idea for v1.0? Please open a feature request issue! We're particularly interested in testing with more Android device manufacturers.

### 3. Submit a Pull Request 🛠️
To contribute code:
1.  **Fork** the repository and create your own branch.
2.  **Add tests** for any new features or bug fixes.
3.  **Ensure all tests pass** by running `make test`.
4.  **Format your code** with `make fmt`.
5.  **Submit a Pull Request** with a clear description of your changes.

### 4. Code Standards 📏
*   Follow the standard Go structure.
*   Keep functions small and well-documented.
*   Avoid adding external dependencies unless absolutely necessary.

---

### Verifying Tests Locally:
To run clinical unit tests:
```bash
make test
```

To run a race condition check:
```bash
make test-race
```

---

Thank you for making DroidTether better for everyone! 🚀
