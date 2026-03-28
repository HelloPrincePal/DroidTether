package rndis

import (
	"encoding/binary"
	"fmt"
)

// CommonHeader is the 8-byte prefix for all RNDIS Control Messages.
type CommonHeader struct {
	MessageType   uint32
	MessageLength uint32
}

// RemoteNdisInitializeMsg is sent by Host -> Device to start the handshake.
type RemoteNdisInitializeMsg struct {
	MessageType    uint32 // MsgInit (0x02)
	MessageLength  uint32 // 24 bytes
	RequestID      uint32
	MajorVersion   uint32 // 1
	MinorVersion   uint32 // 0
	MaxTransferSz  uint32 // Recommended: 16384 (16KB)
}

func (m *RemoteNdisInitializeMsg) Marshal() []byte {
	b := make([]byte, 24)
	binary.LittleEndian.PutUint32(b[0:4], MsgInit)
	binary.LittleEndian.PutUint32(b[4:8], 24)
	binary.LittleEndian.PutUint32(b[8:12], m.RequestID)
	binary.LittleEndian.PutUint32(b[12:16], 1) // Major
	binary.LittleEndian.PutUint32(b[16:20], 0) // Minor
	binary.LittleEndian.PutUint32(b[20:24], m.MaxTransferSz)
	return b
}

// RemoteNdisInitializeCmplt is sent by Device -> Host.
type RemoteNdisInitializeCmplt struct {
	RequestID    uint32
	Status       uint32
	MajorVersion uint32
	MinorVersion uint32
	MaxTransfer  uint32
	MaxPackets   uint32
	PacketFormat uint32 // RNDIS_PACKET_FORMAT_ETH
}

func UnmarshalInitializeCmplt(data []byte) (*RemoteNdisInitializeCmplt, error) {
	if len(data) < 52 {
		return nil, fmt.Errorf("init_cmplt: message too short (%d)", len(data))
	}
	// Verify it's a Completion for Init
	msgType := binary.LittleEndian.Uint32(data[0:4])
	if msgType != MsgInitCmplt {
		return nil, fmt.Errorf("init_cmplt: wrong message type 0x%08X", msgType)
	}

	return &RemoteNdisInitializeCmplt{
		RequestID:    binary.LittleEndian.Uint32(data[8:12]),
		Status:       binary.LittleEndian.Uint32(data[12:16]),
		MajorVersion: binary.LittleEndian.Uint32(data[16:20]),
		MinorVersion: binary.LittleEndian.Uint32(data[20:24]),
		MaxTransfer:  binary.LittleEndian.Uint32(data[32:36]),
		MaxPackets:   binary.LittleEndian.Uint32(data[36:40]),
		PacketFormat: binary.LittleEndian.Uint32(data[44:48]),
	}, nil
}

// RemoteNdisQueryMsg is sent by Host -> Device to query OIDs (e.g., getting MAC address).
type RemoteNdisQueryMsg struct {
	RequestID           uint32
	OID                 uint32
	InformationBufferLength uint32 // for simple queries, usually 0
}

func (m *RemoteNdisQueryMsg) Marshal() []byte {
	b := make([]byte, 28)
	binary.LittleEndian.PutUint32(b[0:4], MsgQuery)
	binary.LittleEndian.PutUint32(b[4:8], 28)
	binary.LittleEndian.PutUint32(b[8:12], m.RequestID)
	binary.LittleEndian.PutUint32(b[12:16], m.OID)
	binary.LittleEndian.PutUint32(b[16:20], 0) // Info Buffer Length
	binary.LittleEndian.PutUint32(b[20:24], 20) // Info Buffer Offset (offset from RequestID field, which is 8 index + 12 = 20)
	binary.LittleEndian.PutUint32(b[24:28], 0) // Device Context
	return b
}

// RemoteNdisQueryCmplt is sent by Device -> Host.
type RemoteNdisQueryCmplt struct {
	RequestID uint32
	Status    uint32
	Payload   []byte
}

func UnmarshalQueryCmplt(data []byte) (*RemoteNdisQueryCmplt, error) {
	if len(data) < 24 {
		return nil, fmt.Errorf("query_cmplt: message too short (%d)", len(data))
	}
	msgType := binary.LittleEndian.Uint32(data[0:4])
	if msgType != MsgQueryCmplt {
		return nil, fmt.Errorf("query_cmplt: wrong message type 0x%08X", msgType)
	}

	infoLen := binary.LittleEndian.Uint32(data[16:20])
	infoOff := binary.LittleEndian.Uint32(data[20:24]) // Offset starts at byte 8 (RequestID)

	start := int(8 + infoOff)
	end := start + int(infoLen)

	if end > len(data) {
		return nil, fmt.Errorf("query_cmplt: payload out of bounds")
	}

	return &RemoteNdisQueryCmplt{
		RequestID: binary.LittleEndian.Uint32(data[8:12]),
		Status:    binary.LittleEndian.Uint32(data[12:16]),
		Payload:   data[start:end],
	}, nil
}

// RemoteNdisSetMsg is sent by Host -> Device to set OIDs (e.g., enable packet processing).
type RemoteNdisSetMsg struct {
	RequestID uint32
	OID       uint32
	Value     uint32
}

func (m *RemoteNdisSetMsg) Marshal() []byte {
	b := make([]byte, 32)
	binary.LittleEndian.PutUint32(b[0:4], MsgSet)
	binary.LittleEndian.PutUint32(b[4:8], 32)
	binary.LittleEndian.PutUint32(b[8:12], m.RequestID)
	binary.LittleEndian.PutUint32(b[12:16], m.OID)
	binary.LittleEndian.PutUint32(b[16:20], 4)  // Value size (uint32)
	binary.LittleEndian.PutUint32(b[20:24], 20) // Value offset (from byte 8)
	binary.LittleEndian.PutUint32(b[24:28], 0)  // Device Context
	binary.LittleEndian.PutUint32(b[28:32], m.Value)
	return b
}

func UnmarshalSetCmplt(data []byte) (uint32, uint32, error) { // returns RequestID, Status
	if len(data) < 16 {
		return 0, 0, fmt.Errorf("set_cmplt: too short")
	}
	msgType := binary.LittleEndian.Uint32(data[0:4])
	if msgType != MsgSetCmplt {
		return 0, 0, fmt.Errorf("set_cmplt: wrong msg type")
	}
	return binary.LittleEndian.Uint32(data[8:12]), binary.LittleEndian.Uint32(data[12:16]), nil
}
