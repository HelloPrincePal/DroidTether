# DroidTether — Quick Reference

One-page cheat sheet. Load this whenever working on RNDIS, utun, or USB code.

---

## RNDIS Protocol

### Message Type Constants

```go
REMOTE_NDIS_INITIALIZE_MSG     = 0x00000002
REMOTE_NDIS_HALT_MSG           = 0x00000003
REMOTE_NDIS_QUERY_MSG          = 0x00000004
REMOTE_NDIS_SET_MSG            = 0x00000005
REMOTE_NDIS_RESET_MSG          = 0x00000006
REMOTE_NDIS_INDICATE_STATUS_MSG = 0x00000007
REMOTE_NDIS_KEEPALIVE_MSG      = 0x00000008
REMOTE_NDIS_INITIALIZE_CMPLT   = 0x80000002
REMOTE_NDIS_QUERY_CMPLT        = 0x80000004
REMOTE_NDIS_SET_CMPLT          = 0x80000005
REMOTE_NDIS_RESET_CMPLT        = 0x80000006
REMOTE_NDIS_KEEPALIVE_CMPLT    = 0x80000008
REMOTE_NDIS_PACKET_MSG         = 0x00000001
```

### INITIALIZE_MSG Layout (host → device)

```
Offset  Size  Field
0       4     MessageType = 0x00000002
4       4     MessageLength = 24
8       4     RequestID (increment each message)
12      4     MajorVersion = 1
16      4     MinorVersion = 0
20      4     MaxTransferSize = 16384
```

### INITIALIZE_CMPLT Layout (device → host)

```
Offset  Size  Field
0       4     MessageType = 0x80000002
4       4     MessageLength
8       4     RequestID (must match request)
12      4     Status (0 = success)
16      4     MajorVersion
20      4     MinorVersion
24      4     DeviceFlags
28      4     Medium = 0 (802.3)
32      4     MaxPacketsPerTransfer
36      4     MaxTransferSize
40      4     PacketAlignmentFactor
44      4     AFListOffset
48      4     AFListSize
```

### QUERY_MSG Layout

```
Offset  Size  Field
0       4     MessageType = 0x00000004
4       4     MessageLength
8       4     RequestID
12      4     OID (what you're querying)
16      4     InformationBufferLength
20      4     InformationBufferOffset
24      4     DeviceVcHandle = 0
```

### PACKET_MSG Layout (data phase)

```
Offset  Size  Field
0       4     MessageType = 0x00000001
4       4     MessageLength
8       4     DataOffset (from byte 8, typically 36)
12      4     DataLength (Ethernet frame size)
16      4     OOBDataOffset = 0
20      4     OOBDataLength = 0
24      4     NumOOBDataElements = 0
28      4     PerPacketInfoOffset = 0
32      4     PerPacketInfoLength = 0
36      4     VcHandle = 0
40      4     Reserved = 0
44      N     Ethernet frame (DataLength bytes)
```

---

## Key OIDs

```go
OID_GEN_SUPPORTED_LIST          = 0x00010101
OID_GEN_HARDWARE_STATUS         = 0x00010102
OID_GEN_MEDIA_SUPPORTED         = 0x00010103
OID_GEN_MEDIA_IN_USE            = 0x00010104
OID_GEN_MAXIMUM_LOOKAHEAD       = 0x00010105
OID_GEN_MAXIMUM_FRAME_SIZE      = 0x00010106
OID_GEN_LINK_SPEED              = 0x00010107
OID_GEN_CURRENT_PACKET_FILTER   = 0x0001010E
OID_GEN_MAXIMUM_TOTAL_SIZE      = 0x00010111
OID_802_3_PERMANENT_ADDRESS     = 0x01010101
OID_802_3_CURRENT_ADDRESS       = 0x01010102
OID_802_3_MULTICAST_LIST        = 0x01010103

// Packet filter flags (OR these together for SET)
NDIS_PACKET_TYPE_DIRECTED       = 0x0001
NDIS_PACKET_TYPE_MULTICAST      = 0x0002
NDIS_PACKET_TYPE_BROADCAST      = 0x0008
NDIS_PACKET_TYPE_PROMISCUOUS    = 0x0020
// Use: DIRECTED | MULTICAST | BROADCAST = 0x000B
```

---

## USB Interface Details

### Android RNDIS USB Class
```
bInterfaceClass    = 0xE0  (Wireless Controller)
bInterfaceSubClass = 0x01  (Bluetooth Programming Interface)
bInterfaceProtocol = 0x03  (RNDIS)
```

### Endpoint Pairs
```
Control:   EP0 (standard USB control endpoint, all devices have this)
Interrupt: EP IN  — 0x89 typical  (device→host notifications)
Bulk IN:   EP IN  — 0x81 typical  (device→host data)
Bulk OUT:  EP OUT — 0x02 typical  (host→device data)
```

**RNDIS control messages** go over USB control (EP0) via vendor requests:
```
bmRequestType = 0x21 (host→device, class, interface)
bRequest      = 0x00 (SEND_ENCAPSULATED_COMMAND) for sending
bRequest      = 0x01 (GET_ENCAPSULATED_RESPONSE) for receiving
```

**RNDIS data packets** go over bulk endpoints (IN/OUT).

---

## utun Interface — macOS Syscall API

### Create a utun interface

```go
import "golang.org/x/sys/unix"

// 1. Open control socket
fd, err := unix.Socket(unix.AF_SYSTEM, unix.SOCK_DGRAM, unix.SYSPROTO_CONTROL)

// 2. Get control ID for utun
ctlInfo := &unix.CtlInfo{}
copy(ctlInfo.Name[:], "com.apple.net.utun_control")
err = unix.IoctlCtlInfo(fd, ctlInfo)

// 3. Connect (this creates the utunN interface)
// Set sc_unit = 0 to let kernel pick the next available index
addr := unix.SockaddrCtl{
    ID:   ctlInfo.Id,
    Unit: 0,  // kernel assigns; read back to get N in utunN
}
err = unix.Connect(fd, &addr)

// 4. fd is now the utun file descriptor — read/write IP packets directly
```

### Read/Write packets

```go
// Write IP packet to utun (4-byte header required on macOS)
header := []byte{0x00, 0x00, 0x00, unix.AF_INET}  // IPv4
packet := append(header, ipPayload...)
n, err := unix.Write(fd, packet)

// Read IP packet from utun
buf := make([]byte, 1500+4)
n, err := unix.Read(fd, buf)
ipPacket := buf[4:n]  // skip 4-byte header
```

### Set IP address and MTU

```go
// Use exec.Command to call ifconfig — simpler than raw ioctls
exec.Command("ifconfig", "utun3", "192.168.42.2", "192.168.42.1", "mtu", "1500").Run()
```

---

## DHCP DORA Sequence

```
Host (DroidTether)                    Phone (DHCP server)
      │                                    │
      │──── DHCPDISCOVER ─────────────────▶│
      │     src: 0.0.0.0:68               │
      │     dst: 255.255.255.255:67        │
      │                                    │
      │◀─── DHCPOFFER ────────────────────│
      │     offered IP: 192.168.42.x      │
      │     server IP:  192.168.42.1      │
      │                                    │
      │──── DHCPREQUEST ──────────────────▶│
      │     request offered IP             │
      │                                    │
      │◀─── DHCPACK ──────────────────────│
      │     confirm IP, lease time        │
      │                                    │
      ✓  Configure utunN with assigned IP
```

Key DHCP option codes:
```
53 = DHCP Message Type (1=DISCOVER, 2=OFFER, 3=REQUEST, 5=ACK)
54 = Server Identifier (server IP)
50 = Requested IP Address
51 = IP Address Lease Time
1  = Subnet Mask
3  = Router (default gateway)
6  = DNS Servers
```

---

## Common Samsung/Android VID/PID Pairs

```go
// Format: {VendorID, ProductID, "Description"}
var KnownRNDIS = []DeviceID{
    {0x04E8, 0x6863, "Samsung Android RNDIS"},
    {0x04E8, 0x685D, "Samsung Android RNDIS"},
    {0x04E8, 0x6881, "Samsung Android RNDIS"},
    {0x18D1, 0x4EE3, "Google Pixel RNDIS"},
    {0x18D1, 0x4EE4, "Google Pixel RNDIS"},
    {0x05C6, 0x9024, "Qualcomm/OnePlus RNDIS"},
    {0x2717, 0xFF80, "Xiaomi RNDIS"},
    {0x0BB4, 0x0EE6, "HTC RNDIS"},
    // Fallback: match by RNDIS class/subclass/protocol
}
```

---

## Default Route Injection

```bash
# Add default route via utun interface
route add default -interface utun3

# Remove it on teardown
route delete default -interface utun3
```

Via Go:
```go
exec.Command("route", "add", "default", "-interface", ifName).Run()
exec.Command("route", "delete", "default", "-interface", ifName).Run()
```

---

## Key Go Dependencies

```go
// go.mod additions needed:
require (
    github.com/google/gousb v1.1.3     // libusb bindings
    github.com/rs/zerolog v1.32.0      // structured logging
    github.com/BurntSushi/toml v1.3.3  // TOML config
    golang.org/x/sys v0.18.0           // unix syscalls for utun
)
```

Build command for Apple Silicon:
```bash
GOARCH=arm64 GOOS=darwin CGO_ENABLED=1 go build -o droidtether ./cmd/droidtether
```