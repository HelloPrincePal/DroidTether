package tun

import (
	"io"
)

// Interface represents a virtual network interface (TUN).
type Interface interface {
	io.ReadWriteCloser
	Name() string
}
