# internal/dhcp

A minimal DHCP client to request an IP address from the Android phone's tethering DHCP server.

## Responsibility

- **DORA Sequence**: Implements the Discover → Offer → Request → Acknowledge flow.
- **Packet Encoding**: Encodes and decodes DHCP binary packets (Layer 4 UDP inside Layer 3 IP).
- **Lease Management**: Extracts the assigned IP, gateway, and DNS servers from the `DHCPACK`.

## Usage

DHCP packets are sent as raw IP packets through the `utun` interface during the session setup phase, before the general packet relay starts.
