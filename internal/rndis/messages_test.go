package rndis

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestEncapsulatePacketLayout locks the wire format to the canonical
// REMOTE_NDIS_PACKET_MSG layout used by the Linux host driver: a 44-byte
// header, DataOffset 36 (relative to the start of RequestId) and the
// Ethernet frame immediately after the header.
func TestEncapsulatePacketLayout(t *testing.T) {
	frame := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x02, 0x00, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00}
	pkt := EncapsulatePacket(frame)

	if len(pkt) != 44+len(frame) {
		t.Fatalf("packet length = %d, want %d", len(pkt), 44+len(frame))
	}
	if got := binary.LittleEndian.Uint32(pkt[0:4]); got != MsgPacket {
		t.Errorf("MessageType = 0x%08X, want 0x%08X", got, MsgPacket)
	}
	if got := binary.LittleEndian.Uint32(pkt[4:8]); got != uint32(len(pkt)) {
		t.Errorf("MessageLength = %d, want %d", got, len(pkt))
	}
	if got := binary.LittleEndian.Uint32(pkt[8:12]); got != 36 {
		t.Errorf("DataOffset = %d, want 36", got)
	}
	if got := binary.LittleEndian.Uint32(pkt[12:16]); got != uint32(len(frame)) {
		t.Errorf("DataLength = %d, want %d", got, len(frame))
	}
	if !bytes.Equal(pkt[44:], frame) {
		t.Errorf("payload = %x, want %x", pkt[44:], frame)
	}
	for i := 16; i < 44; i++ {
		if pkt[i] != 0 {
			t.Errorf("reserved byte %d = 0x%02X, want 0", i, pkt[i])
		}
	}
}

// TestDecapsulatePacket checks that the canonical layout round-trips through
// the decoder and that truncated input is rejected.
func TestDecapsulatePacket(t *testing.T) {
	frame := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x01, 0x42, 0x86, 0x6b, 0x49, 0x67, 0x24, 0x08, 0x06, 0x01, 0x02}
	pkt := EncapsulatePacket(frame)

	got, err := DecapsulatePacket(pkt)
	if err != nil {
		t.Fatalf("DecapsulatePacket: %v", err)
	}
	if !bytes.Equal(got, frame) {
		t.Errorf("round-trip payload = %x, want %x", got, frame)
	}

	if _, err := DecapsulatePacket(pkt[:20]); err == nil {
		t.Error("expected error for truncated packet")
	}
}
