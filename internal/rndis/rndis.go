package rndis

import (
	"fmt"
	"time"

	"github.com/princePal/droidtether/internal/usb"
	"github.com/rs/zerolog/log"
)

// Session manages a single RNDIS logical session.
type Session struct {
	dev       *usb.Device
	requestID uint32
}

// NewSession creates an RNDIS session on the provided USB device.
func NewSession(dev *usb.Device) *Session {
	return &Session{
		dev: dev,
	}
}

// Handshake performs the RNDIS INIT -> QUERY(MAC) -> SET(FILTER) sequence.
func (s *Session) Handshake() error {
	log.Info().Str("component", "rndis").Msg("Starting RNDIS handshake...")

	// 1. Send INIT
	s.requestID++
	initMsg := &RemoteNdisInitializeMsg{
		RequestID:     s.requestID,
		MaxTransferSz: 16384,
	}

	if err := s.sendControl(initMsg.Marshal()); err != nil {
		return fmt.Errorf("rndis: failed to send INIT: %w", err)
	}

	initResp, err := s.receiveControl()
	if err != nil {
		return fmt.Errorf("rndis: failed to receive INIT_CMPLT: %w", err)
	}

	initCmplt, err := UnmarshalInitializeCmplt(initResp)
	if err != nil {
		return err
	}
	if initCmplt.Status != StatusSuccess {
		return fmt.Errorf("rndis: INIT failed with status 0x%08X", initCmplt.Status)
	}

	log.Debug().
		Str("component", "rndis").
		Uint32("max_transfer", initCmplt.MaxTransfer).
		Msg("RNDIS initialized")

	// 2. Query MAC Address
	s.requestID++
	queryMac := &RemoteNdisQueryMsg{
		RequestID: s.requestID,
		OID:       OID_802_3_PERMANENT_ADDRESS,
	}

	if err := s.sendControl(queryMac.Marshal()); err != nil {
		return fmt.Errorf("rndis: failed to query MAC: %w", err)
	}

	queryResp, err := s.receiveControl()
	if err != nil {
		return fmt.Errorf("rndis: failed to receive MAC_QUERY_CMPLT: %w", err)
	}

	macCmplt, err := UnmarshalQueryCmplt(queryResp)
	if err != nil {
		return err
	}
	log.Info().
		Str("component", "rndis").
		Hex("mac", macCmplt.Payload).
		Msg("Device MAC address retrieved")

	// 3. Set Packet Filter (Enable data flow)
	s.requestID++
	setFilter := &RemoteNdisSetMsg{
		RequestID: s.requestID,
		OID:       OID_GEN_CURRENT_PACKET_FILTER,
		Value:     PacketTypeDirected | PacketTypeBroadcast | PacketTypeAllMulticast,
	}

	if err := s.sendControl(setFilter.Marshal()); err != nil {
		return fmt.Errorf("rndis: failed to set packet filter: %w", err)
	}

	setResp, err := s.receiveControl()
	if err != nil {
		return fmt.Errorf("rndis: failed to receive SET_CMPLT: %w", err)
	}

	_, setStatus, err := UnmarshalSetCmplt(setResp)
	if err != nil {
		return err
	}
	if setStatus != StatusSuccess {
		return fmt.Errorf("rndis: SET filter failed with status 0x%08X", setStatus)
	}

	log.Info().Str("component", "rndis").Msg("RNDIS handshake complete. Device in DATA mode.")
	return nil
}

// sendControl sends an encapsulated RNDIS command via USB control endpoint.
func (s *Session) sendControl(data []byte) error {
	// bmRequestType = 0x21 (Host-to-Device | Class | Interface)
	// bRequest = 0x00 (SEND_ENCAPSULATED_COMMAND)
	_, err := s.dev.Control(0x21, 0x00, 0, uint16(s.dev.InterfaceNum), data)
	return err
}

// receiveControl retrieves an encapsulated RNDIS response via USB control endpoint.
func (s *Session) receiveControl() ([]byte, error) {
	buf := make([]byte, 1024)
	
	// RNDIS requires polling for the response. Usually we should wait for 
	// a Notification on the interrupt endpoint first, but many devices
	// allow direct polling on EP0 after a short delay.
	for i := 0; i < 5; i++ {
		time.Sleep(20 * time.Millisecond)
		// bmRequestType = 0xA1 (Device-to-Host | Class | Interface)
		// bRequest = 0x01 (GET_ENCAPSULATED_RESPONSE)
		n, err := s.dev.Control(0xA1, 0x01, 0, uint16(s.dev.InterfaceNum), buf)
		if err == nil && n > 0 {
			return buf[:n], nil
		}
		// Code -3 or similar might mean it's busy, so we retry.
	}
	return nil, fmt.Errorf("rndis: timeout waiting for response")
}
