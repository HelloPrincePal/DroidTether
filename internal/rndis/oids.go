package rndis

// RNDIS message types (Host -> Device)
const (
	MsgInit          uint32 = 0x00000002
	MsgQuery         uint32 = 0x00000004
	MsgSet           uint32 = 0x00000005
	MsgReset         uint32 = 0x00000006
	MsgKeepAlive     uint32 = 0x00000008
	MsgPacket        uint32 = 0x00000001
)

// RNDIS message completion types (Device -> Host)
const (
	MsgInitCmplt      uint32 = 0x80000002
	MsgQueryCmplt     uint32 = 0x80000004
	MsgSetCmplt       uint32 = 0x80000005
	MsgResetCmplt     uint32 = 0x80000006
	MsgKeepAliveCmplt uint32 = 0x80000008
)

// RNDIS Object Identifiers (OIDs)
const (
	OID_GEN_SUPPORTED_LIST        uint32 = 0x00010101
	OID_GEN_HARDWARE_STATUS       uint32 = 0x00010102
	OID_GEN_MEDIA_SUPPORTED       uint32 = 0x00010103
	OID_GEN_MEDIA_IN_USE          uint32 = 0x00010104
	OID_GEN_MAXIMUM_LOOKAHEAD     uint32 = 0x00010105
	OID_GEN_MAXIMUM_FRAME_SIZE    uint32 = 0x00010106
	OID_GEN_LINK_SPEED            uint32 = 0x00010107
	OID_GEN_TRANSMIT_BUFFER_SPACE uint32 = 0x00010108
	OID_GEN_RECEIVE_BUFFER_SPACE  uint32 = 0x00010109
	OID_GEN_TRANSMIT_BLOCK_SIZE   uint32 = 0x0001010A
	OID_GEN_RECEIVE_BLOCK_SIZE    uint32 = 0x0001010B
	OID_GEN_VENDOR_ID             uint32 = 0x0001010C
	OID_GEN_VENDOR_DESCRIPTION    uint32 = 0x0001010D
	OID_GEN_CURRENT_PACKET_FILTER uint32 = 0x0001010E
	OID_GEN_MAXIMUM_TOTAL_SIZE    uint32 = 0x00010111

	OID_802_3_PERMANENT_ADDRESS   uint32 = 0x01010101
	OID_802_3_CURRENT_ADDRESS     uint32 = 0x01010102
	OID_802_3_MULTICAST_LIST      uint32 = 0x01010103
	OID_802_3_MAXIMUM_LIST_SIZE   uint32 = 0x01010104
)

// RNDIS Packet Filters
const (
	PacketTypeDirected      uint32 = 0x00000001
	PacketTypeMulticast     uint32 = 0x00000002
	PacketTypeAllMulticast  uint32 = 0x00000004
	PacketTypeBroadcast     uint32 = 0x00000008
	PacketTypeSourceRouting uint32 = 0x00000010
	PacketTypePromiscuous   uint32 = 0x00000020
	PacketTypeSMT           uint32 = 0x00000040
	PacketTypeAllLocal      uint32 = 0x00000080
	PacketTypeAllFunctional uint32 = 0x00000100
	PacketTypeFunctional    uint32 = 0x00000200
	PacketTypeGroup         uint32 = 0x00000400
)

// RNDIS Status Codes
const (
	StatusSuccess          uint32 = 0x00000000
	StatusFailure          uint32 = 0xC0000001
	StatusNotSupported     uint32 = 0xC00000BB
	StatusInvalidData      uint32 = 0xC0010015
	StatusInvalidLength    uint32 = 0xC0010014
	StatusResources        uint32 = 0xC000009A
)
