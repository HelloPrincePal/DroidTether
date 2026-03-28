# internal/tun

Handles macOS `utun` interface creation and bidirectional packet relaying.

## Responsibility

- **utun**: Creates a native macOS `utun` interface (e.g., `utun3`) using `AF_SYSTEM` sockets.
- **Relay**: A high-performance bridge that moves packets between the USB bulk endpoints and the `utun` file descriptor.
- **Packet Transform**:
    - **USB → utun**: Strip RNDIS header (44 bytes) and Ethernet header (14 bytes); prepend 4-byte `utun` family header (e.g., `AF_INET`).
    - **utun → USB**: Strip `utun` header; prepend Ethernet header and RNDIS `PACKET_MSG` header.

## Implementation Details

The relay uses two dedicated goroutines:
1. `USBtoTUN`: Reads from USB bulk IN, transforms, and writes to `utun` FD.
2. `TUNtoUSB`: Reads from `utun` FD, transforms, and writes to USB bulk OUT.
