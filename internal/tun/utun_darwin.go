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

	// Cleanup DNS we added
	_ = exec.Command("scutil", "-w", "State:/Network/Service/droidtether/DNS").Run()
	cmd := exec.Command("scutil")
	stdin, _ := cmd.StdinPipe()
	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, "open")
		fmt.Fprintln(stdin, "remove State:/Network/Service/droidtether/DNS")
		fmt.Fprintln(stdin, "quit")
	}()
	_ = cmd.Run()

	return i.f.Close()
}
func (i *utunInterface) Name() string {
	return i.name
}

// Configure sets the IP addresses for the utun interface using the 'ifconfig' command.
func (i *utunInterface) Configure(localIP, remoteIP string) error {
	// Formula: ifconfig <name> <local> <remote> up
	cmd := exec.Command("ifconfig", i.name, localIP, remoteIP, "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ifconfig failed: %w (output: %s)", err, string(out))
	}
	return nil
}

// SetDefaultRoute adds "more specific" default routes (0.0.0.0/1 and 128.0.0.0/1)
// to override the existing default route without deleting it.
func (i *utunInterface) SetDefaultRoute(gateway string) error {
	// 0.0.0.0/1
	cmd1 := exec.Command("route", "add", "-net", "0.0.0.0/1", "-interface", i.name)
	if out, err := cmd1.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add 0/1 route: %w (output: %s)", err, string(out))
	}

	// 128.0.0.0/1
	cmd2 := exec.Command("route", "add", "-net", "128.0.0.0/1", "-interface", i.name)
	if out, err := cmd2.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add 128/1 route: %w (output: %s)", err, string(out))
	}

	return nil
}

// SetDNS sets the system DNS to the phone gateway (or other provided servers)
// using 'scutil' on macOS.
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
		fmt.Fprintln(stdin, "set State:/Network/Service/droidtether/DNS")
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
