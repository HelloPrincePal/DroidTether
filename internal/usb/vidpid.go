package usb

import "strings"

// Standard RNDIS USB Interface Class, SubClass, and Protocol
const (
	RNDISClass    uint8 = 0xE0 // Wireless Controller
	RNDISSubClass uint8 = 0x01 // Radio Frequency
	RNDISProtocol uint8 = 0x03 // RNDIS
)

// Known Android Manufacturer VIDs (Fallback for misbehaving devices that don't advertise RNDIS class properly)
var knownAndroidVIDs = []uint16{
	0x04e8, // Samsung
	0x18d1, // Google (Pixel / Nexus)
	0x2717, // Xiaomi
	0x22d9, // OPPO / OnePlus
	0x2a70, // OnePlus
	0x0fce, // Sony Ericsson
	0x05c6, // Qualcomm
	0x12d1, // Huawei
	0x2b0e, // Nothing
	0x17ef, // Lenovo / Motorola
	0x22b8, // Motorola
}

// ManufacturerNameKeywords matches known Android device manufacturers by string (if VID is unknown).
var knownAndroidManufacturers = []string{
	"android",
	"samsung",
	"google",
	"oneplus",
	"xiaomi",
	"motorola",
	"huawei",
	"sony",
	"oppo",
	"vivo",
	"nothing",
	"lg",
	"nokia",
}

// MatchRNDIS checks whether a given USB Interface matches the signature of an Android RNDIS tethering interface.
// It returns true if the class/subclass/protocol match the RNDIS spec, or as a fallback, if the device matches
// a known Android manufacturer VID.
func MatchRNDIS(vid, pid uint16, class, subClass, proto uint8) bool {
	// Primary check: Does the interface explicitly declare itself as RNDIS?
	if class == RNDISClass && subClass == RNDISSubClass && proto == RNDISProtocol {
		return true
	}

	// 2. Vendor-specific fallback (0xFF or 0xEF)
	// Many vendors (Xiaomi, Samsung) use 0xFF or 0xEF class for various USB modes.
	if class == 0xFF || class == 0xEF {
		// Protocol 0x03 often identifies RNDIS even in vendor-specific class
		if proto == RNDISProtocol {
			return true
		}
		// Fallback for known Android manufacturers (Samsung, Xiaomi, etc.)
		for _, knownVID := range knownAndroidVIDs {
			if vid == knownVID {
				return true
			}
		}
	}

	return false
}

// IsKnownAndroidManufacturer checks if the manufacturer string matches a known Android manufacturer.
func IsKnownAndroidManufacturer(manufacturer string) bool {
	lower := strings.ToLower(manufacturer)
	for _, kw := range knownAndroidManufacturers {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
