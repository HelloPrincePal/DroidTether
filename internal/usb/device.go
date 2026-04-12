package usb

import (
	"fmt"
	"github.com/google/gousb"
)

// Device wraps a gousb.Device with RNDIS specific fields and capabilities.
type Device struct {
	usbd         *gousb.Device
	usbc         *gousb.Config
	controlIntf  *gousb.Interface
	dataIntf     *gousb.Interface
	InterfaceNum int
}

// NewDevice opens and claims the RNDIS interface on the provided gousb.Device.
func NewDevice(dev *gousb.Device) (*Device, error) {
	var usbc *gousb.Config
	var rndisInterfaceNum, rndisAltSettingNum int = -1, -1

	// 1. Iterate through all configurations to find the RNDIS one.
	for _, cfgDesc := range dev.Desc.Configs {
		c, err := dev.Config(cfgDesc.Number)
		if err != nil {
			fmt.Printf("USB DEBUG: failed to open config %d: %v\n", cfgDesc.Number, err)
			continue
		}

		found := false
		for _, intfDesc := range c.Desc.Interfaces {
			for _, altDesc := range intfDesc.AltSettings {
				// Debug log all interfaces
				fmt.Printf("USB DEBUG: Config %d, Interface %d, Alt %d, Class 0x%02X, SubClass 0x%02X, Protocol 0x%02X, Endpoints %d\n",
					cfgDesc.Number, intfDesc.Number, altDesc.Number, int(altDesc.Class), int(altDesc.SubClass), int(altDesc.Protocol), len(altDesc.Endpoints))

				if MatchRNDIS(uint16(dev.Desc.Vendor), uint16(dev.Desc.Product), uint8(altDesc.Class), uint8(altDesc.SubClass), uint8(altDesc.Protocol)) {
					if rndisInterfaceNum == -1 {
						rndisInterfaceNum = intfDesc.Number
						rndisAltSettingNum = altDesc.Number
						usbc = c
						found = true
						break
					}
				}
			}
			if found {
				break
			}
		}

		if !found {
			c.Close()
		} else {
			break
		}
	}

	if usbc == nil {
		return nil, fmt.Errorf("no RNDIS interface found on any device configuration")
	}

	// 2. Set AutoDetach (detaches macOS built-in driver automatically).
	dev.SetAutoDetach(true)

	// 3. Claim the control interface.
	intf, err := usbc.Interface(rndisInterfaceNum, rndisAltSettingNum)
	if err != nil {
		usbc.Close()
		return nil, fmt.Errorf("failed to claim RNDIS interface %d: %w", rndisInterfaceNum, err)
	}

	return &Device{
		usbd:         dev,
		usbc:         usbc,
		controlIntf:  intf,
		InterfaceNum: rndisInterfaceNum,
	}, nil
}

// Close releases the claimed USB interfaces and config configuration.
func (d *Device) Close() error {
	if d.dataIntf != nil {
		d.dataIntf.Close()
		d.dataIntf = nil
	}
	if d.controlIntf != nil {
		d.controlIntf.Close()
		d.controlIntf = nil
	}
	if d.usbc != nil {
		d.usbc.Close()
		d.usbc = nil
	}
	return nil
}

// Control performs a vendor-specific control transfer.
func (d *Device) Control(rType, request uint8, val, idx uint16, data []byte) (int, error) {
	return d.usbd.Control(rType, request, val, idx, data)
}

// OpenBulkEndpoints returns the bulk IN and OUT endpoints for data transfer.
func (d *Device) OpenBulkEndpoints() (in *gousb.InEndpoint, out *gousb.OutEndpoint, err error) {
	// 1. Try finding endpoints on the current (Control) interface first.
	in, out, _ = d.findEndpoints(d.controlIntf)
	if in != nil && out != nil {
		return in, out, nil
	}

	// 2. Scan for a Data interface if not already claimed
	if d.dataIntf == nil {
		for _, intfDesc := range d.usbc.Desc.Interfaces {
			if intfDesc.Number == d.InterfaceNum {
				continue
			}
			for _, alt := range intfDesc.AltSettings {
				if len(alt.Endpoints) >= 2 {
					fmt.Printf("USB DEBUG: Try claim intf %d alt %d\n", intfDesc.Number, alt.Number)
					intf, err := d.usbc.Interface(intfDesc.Number, alt.Number)
					if err != nil {
						fmt.Printf("USB DEBUG: failed to claim intf %d alt %d: %v\n", intfDesc.Number, alt.Number, err)
						if alt.Number != 0 {
							fmt.Printf("USB DEBUG: trying fallback alt 0 for intf %d\n", intfDesc.Number)
							intf, err = d.usbc.Interface(intfDesc.Number, 0)
							if err != nil {
								fmt.Printf("USB DEBUG: fallback failed: %v\n", err)
								continue
							}
						} else {
							continue
						}
					}
					
					in, out, _ = d.findEndpoints(intf)
					if in != nil && out != nil {
						fmt.Printf("USB DEBUG: Found bulk endpoints on intf %d alt %d\n", intfDesc.Number, alt.Number)
						d.dataIntf = intf
						return in, out, nil
					} else {
						fmt.Printf("USB DEBUG: Endpoints not found after claiming.\n")
					}
					intf.Close()
				}
			}
		}
	} else {
		return d.findEndpoints(d.dataIntf)
	}

	return nil, nil, fmt.Errorf("could not find bulk endpoint pair")
}

func (d *Device) findEndpoints(intf *gousb.Interface) (*gousb.InEndpoint, *gousb.OutEndpoint, error) {
	var in *gousb.InEndpoint
	var out *gousb.OutEndpoint

	for _, ep := range intf.Setting.Endpoints {
		if ep.TransferType == gousb.TransferTypeBulk {
			if ep.Direction == gousb.EndpointDirectionIn {
				in, _ = intf.InEndpoint(ep.Number)
			} else {
				out, _ = intf.OutEndpoint(ep.Number)
			}
		}
	}
	if in == nil || out == nil {
		return nil, nil, fmt.Errorf("missing endpoints")
	}
	return in, out, nil
}
