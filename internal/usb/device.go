package usb

import (
	"fmt"
	"github.com/google/gousb"
)

// Device wraps a gousb.Device with RNDIS specific fields and capabilities.
type Device struct {
	usbd         *gousb.Device
	usbc         *gousb.Config
	usbIntf      *gousb.Interface
	InterfaceNum int
}

// NewDevice opens and claims the RNDIS interface on the provided gousb.Device.
func NewDevice(dev *gousb.Device) (*Device, error) {
	// 1. Detach kernel driver if necessary.
	// On macOS, claiming the interface detaches the native driver if we have proper config.
	// We'll set the device config to the first one.
    // NOTE: libusb handles detaching the kernel driver automatically on claim if configured.
	usbc, err := dev.Config(1)
	if err != nil {
		return nil, fmt.Errorf("failed to open device config 1: %w", err)
	}

	// 2. Iterate through interfaces to find the RNDIS one.
	var rndisInterfaceNum, rndisAltSettingNum int = -1, -1
	for _, intfDesc := range usbc.Desc.Interfaces {
		for _, altDesc := range intfDesc.AltSettings {
			// vid and pid are available at the device level, we pass dummy values here as we check class directly
			if MatchRNDIS(uint16(dev.Desc.Vendor), uint16(dev.Desc.Product), uint8(altDesc.Class), uint8(altDesc.SubClass), uint8(altDesc.Protocol)) {
				rndisInterfaceNum = intfDesc.Number
				rndisAltSettingNum = altDesc.Number
				break
			}
		}
		if rndisInterfaceNum != -1 {
			break
		}
	}

	if rndisInterfaceNum == -1 {
		usbc.Close()
		return nil, fmt.Errorf("no RNDIS interface found on device")
	}

	// 3. Set AutoDetach (detaches macOS built-in driver automatically).
	dev.SetAutoDetach(true)

	// 4. Claim the interface.
	intf, err := usbc.Interface(rndisInterfaceNum, rndisAltSettingNum)
	if err != nil {
		usbc.Close()
		return nil, fmt.Errorf("failed to claim RNDIS interface %d: %w", rndisInterfaceNum, err)
	}

	return &Device{
		usbd:         dev,
		usbc:         usbc,
		usbIntf:      intf,
		InterfaceNum: rndisInterfaceNum,
	}, nil
}

// Close releases the claimed USB interfaces and config configuration.
func (d *Device) Close() error {
	if d.usbIntf != nil {
		d.usbIntf.Close()
		d.usbIntf = nil
	}
	if d.usbc != nil {
		d.usbc.Close()
		d.usbc = nil
	}
	// Note: We don't close d.usbd here, because the Watcher handles raw gousb.Device lifecycle.
	return nil
}
// Control performs a vendor-specific control transfer.
func (d *Device) Control(rType, request uint8, val, idx uint16, data []byte) (int, error) {
	return d.usbd.Control(rType, request, val, idx, data)
}

// OpenBulkEndpoints returns the bulk IN and OUT endpoints for data transfer.
// index is the bulk endpoint pair index.
func (d *Device) OpenBulkEndpoints() (in *gousb.InEndpoint, out *gousb.OutEndpoint, err error) {
	// Find bulk endpoints in the claimed interface
	for _, ep := range d.usbIntf.Setting.Endpoints {
		if ep.TransferType == gousb.TransferTypeBulk {
			if ep.Direction == gousb.EndpointDirectionIn {
				in, err = d.usbIntf.InEndpoint(ep.Number)
			} else {
				out, err = d.usbIntf.OutEndpoint(ep.Number)
			}
		}
	}
	if in == nil || out == nil {
		return nil, nil, fmt.Errorf("could not find bulk endpoint pair")
	}
	return in, out, nil
}
