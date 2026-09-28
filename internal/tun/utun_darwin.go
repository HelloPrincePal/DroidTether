package tun

import (
	"fmt"
	"os"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/unix"
)

// utun configuration constants for macOS
const (
	AF_SYSTEM         = 32
	SYSPROTO_CONTROL  = 2
	AF_SYS_CONTROL    = 2
	UTUN_CONTROL_NAME = "com.apple.net.utun_control"
)

// utunInterface implements Interface for Darwin using AF_SYSTEM.
type utunInterface struct {
	f    *os.File
	name string
}

func (i *utunInterface) Read(p []byte) (n int, err error) {
	return i.f.Read(p)
}

func (i *utunInterface) Write(p []byte) (n int, err error) {
	return i.f.Write(p)
}

func (i *utunInterface) Close() error {
	// Cleanup routes we added
	_ = exec.Command("route", "delete", "-net", "0.0.0.0/1").Run()
	_ = exec.Command("route", "delete", "-net", "128.0.0.0/1").Run()

	// Cleanup the SystemConfiguration registration and DNS we added
	cmd := exec.Command("scutil")
	stdin, _ := cmd.StdinPipe()
	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, "open")
		fmt.Fprintln(stdin, "remove State:/Network/Service/droidtether/IPv4")
		fmt.Fprintln(stdin, "remove State:/Network/Service/droidtether/DNS")
		fmt.Fprintln(stdin, "remove State:/Network/Global/DNS")
	}()
	_ = cmd.Run()

	return i.f.Close()
}
func (i *utunInterface) Name() string {
	return i.name
}

// Configure sets the IP addresses and MTU for the utun interface using the 'ifconfig' command.
func (i *utunInterface) Configure(localIP, remoteIP, mtu string) error {
	// Formula: ifconfig <name> <local> <remote> mtu <val> up
	cmd := exec.Command("ifconfig", i.name, localIP, remoteIP, "mtu", mtu, "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ifconfig failed: %w (output: %s)", err, string(out))
	}
	return nil
}

// SetDefaultRoute adds "more specific" default routes (0.0.0.0/1 and
// 128.0.0.0/1) to override the existing default route without deleting it.
// Idempotent: a plain `route add` fails with "File exists" when the route is
// already installed, so delete-then-add and tolerate a failed delete.
func (i *utunInterface) SetDefaultRoute(gateway string) error {
	_ = gateway // routes are injected by interface, not via the gateway

	for _, net := range []string{"0.0.0.0/1", "128.0.0.0/1"} {
		_ = exec.Command("route", "delete", "-net", net).Run()
		cmd := exec.Command("route", "add", "-net", net, "-interface", i.name)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add %s route: %w (output: %s)", net, err, string(out))
		}
	}

	return nil
}

// RegisterService registers DroidTether as a real SystemConfiguration
// service by writing State:/Network/Service/droidtether/IPv4
// (InterfaceName/Addresses/SubnetMasks/Router). configd then builds the
// service's route list itself: a non-zero Router yields a default route,
// which makes the service routable, so the primary election picks it and
// the derived global DNS resolver gets the RequestA bit. Without that bit
// mDNSResponder marks A/AAAA records "unusable" and suppresses every
// getaddrinfo query (the post-macOS-26 tethering breakage). Android
// tethering subnets are /24; utun is point-to-point, so the mask only
// shapes the subnet route and never reachability.
func (i *utunInterface) RegisterService(localIP, gateway string) error {
	cmd := exec.Command("scutil")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, "open")
		fmt.Fprintln(stdin, "d.init")
		fmt.Fprintln(stdin, "d.add InterfaceName "+i.name)
		fmt.Fprintln(stdin, "d.add Addresses * "+localIP)
		fmt.Fprintln(stdin, "d.add SubnetMasks * 255.255.255.0")
		fmt.Fprintln(stdin, "d.add Router "+gateway)
		fmt.Fprintln(stdin, "set State:/Network/Service/droidtether/IPv4")
		fmt.Fprintln(stdin, "quit")
	}()
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("scutil service IPv4 set failed: %w (output: %s)", err, string(out))
	}
	return nil
}

// SetDNS writes our nameservers into the droidtether service's DNS entity
// (with __FLAGS__ 6 = RequestA|RequestAAAA so configd's derivation keeps the
// RequestA bit) and clears any stale raw override of State:/Network/Global/
// DNS. configd republishes the derived resolver once the service is routable
// (see RegisterService). Global/DNS must never be set directly: a raw write
// is ignored by mDNSResponder and makes configd tear down the primary
// election, re-deriving a resolver without RequestA.
func (i *utunInterface) SetDNS(dnsServers []string) error {
	if len(dnsServers) == 0 {
		return nil
	}

	cmd := exec.Command("scutil")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, "open")
		fmt.Fprintln(stdin, "d.init")
		fmt.Fprint(stdin, "d.add ServerAddresses *")
		for _, s := range dnsServers {
			fmt.Fprintf(stdin, " %s", s)
		}
		fmt.Fprintln(stdin)
		// SupplementalMatchDomains set to empty string tells macOS to use these servers for ALL lookups.
		fmt.Fprintln(stdin, "d.add SupplementalMatchDomains * \"\"")
		fmt.Fprintln(stdin, "d.add SupplementalMatchOrders * 10")
		// RequestA/RequestAAAA (0x2|0x4): without them mDNSResponder treats
		// A/AAAA records as "unusable" and suppresses every getaddrinfo
		// query. configd ORs these into __FLAGS__ (=> 0x4006; it masks the
		// AAAA bit back off when the interface has no IPv6).
		fmt.Fprintln(stdin, "d.add __FLAGS__ # 6")
		fmt.Fprintln(stdin, "set State:/Network/Service/droidtether/DNS")
		fmt.Fprintln(stdin, "remove State:/Network/Global/DNS")
		fmt.Fprintln(stdin, "quit")
	}()

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("scutil DNS set failed: %w (output: %s)", err, string(out))
	}

	return nil
}

// OpenUTUN creates a new utun interface on macOS.
// If index is 0, the system chooses the first available (utun0, utun1, etc.).
func OpenUTUN(index int) (Interface, error) {
	fd, err := unix.Socket(AF_SYSTEM, unix.SOCK_DGRAM, SYSPROTO_CONTROL)
	if err != nil {
		return nil, fmt.Errorf("utun: failed to open system socket: %w", err)
	}

	// 1. Find the control ID for "com.apple.net.utun_control"
	info := struct {
		ctl_id   uint32
		ctl_name [96]byte
	}{}
	copy(info.ctl_name[:], UTUN_CONTROL_NAME)

	// CTLIOCGINFO
	err = ioctl(fd, 0xc0644e03, unsafe.Pointer(&info))
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("utun: failed to get utun control info: %w", err)
	}

	// 2. Connect to the utun control
	sc := struct {
		sc_len      uint8
		sc_family   uint8
		ss_sysaddr  uint16
		sc_id       uint32
		sc_unit     uint32
		sc_reserved [5]uint32
	}{
		sc_len:     32,
		sc_family:  AF_SYSTEM,
		ss_sysaddr: AF_SYS_CONTROL,
		sc_id:      info.ctl_id,
		sc_unit:    uint32(index), // 0 = automatic
	}

	err = connect(fd, unsafe.Pointer(&sc), 32)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("utun: failed to connect to utun control: %w", err)
	}

	// 3. Get the interface name (e.g., utun3)
	nameBuf := make([]byte, 64)
	nameLen := uint32(len(nameBuf))
	// UTUN_OPT_IFNAME (Option 2)
	err = getsockopt(fd, SYSPROTO_CONTROL, 2, unsafe.Pointer(&nameBuf[0]), &nameLen)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("utun: failed to get interface name: %w", err)
	}

	ifname := string(nameBuf[:nameLen-1]) // trim null byte
	return &utunInterface{
		f:    os.NewFile(uintptr(fd), ifname),
		name: ifname,
	}, nil
}

// Wrapper for unix.Ioctl
func ioctl(fd int, request uintptr, argp unsafe.Pointer) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), request, uintptr(argp))
	if errno != 0 {
		return errno
	}
	return nil
}

// Wrapper for unix.Connect
func connect(fd int, addr unsafe.Pointer, len uint32) error {
	_, _, errno := unix.Syscall(unix.SYS_CONNECT, uintptr(fd), uintptr(addr), uintptr(len))
	if errno != 0 {
		return errno
	}
	return nil
}

// Wrapper for unix.Getsockopt
func getsockopt(fd int, level, name int, val unsafe.Pointer, len *uint32) error {
	_, _, errno := unix.Syscall6(unix.SYS_GETSOCKOPT, uintptr(fd), uintptr(level), uintptr(name), uintptr(val), uintptr(unsafe.Pointer(len)), 0)
	if errno != 0 {
		return errno
	}
	return nil
}
