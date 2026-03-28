# internal/usb

This package handles USB device discovery, hotplug events, and low-level communication with Android devices via `libusb`.

## Responsibility

- **Watcher**: Polls for USB hotplug events and identifies Android devices matching the RNDIS class.
- **Device**: Handles opening/closing devices, detaching kernel drivers, and claiming the RNDIS interface.
- **VID/PID**: Maintains a list of known RNDIS-supported devices as a fallback to class-based matching.

## Exported Symbols

- `Watcher`: Struct to monitor USB state changes.
- `Device`: Wrapper around a `gousb.Device` with RNDIS-specific capabilities.
- `MatchRNDIS(v, p, class, subClass, proto uint8) bool`: Checks if a device is an Android RNDIS interface.

## Usage

```go
watcher := usb.NewWatcher()
watcher.OnAttach(func(dev *usb.Device) {
    // handle new device
})
watcher.Start()
```
