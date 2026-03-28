package daemon

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/google/gousb"
	"github.com/princePal/droidtether/internal/rndis"
	"github.com/princePal/droidtether/internal/tun"
	"github.com/princePal/droidtether/internal/usb"
	"github.com/rs/zerolog/log"
)

// Relay handles the packet shuttle between USB Bulk and Tunnel interface.
type Relay struct {
	dev   *usb.Device
	tun   tun.Interface
	ctx   context.Context
	cancel context.CancelFunc

	usbIn  *gousb.InEndpoint
	usbOut *gousb.OutEndpoint

	phoneMAC []byte
}

// NewRelay creates a new bidirectional relay.
func NewRelay(dev *usb.Device, tun tun.Interface, phoneMAC []byte) (*Relay, error) {
	in, out, err := dev.OpenBulkEndpoints()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Relay{
		dev:      dev,
		tun:      tun,
		ctx:      ctx,
		cancel:   cancel,
		usbIn:    in,
		usbOut:   out,
		phoneMAC: phoneMAC,
	}, nil
}

// Start spawns the relay goroutines and blocks until context is cancelled or an error occurs.
func (r *Relay) Start() error {
	errChan := make(chan error, 2)

	// Mac -> Phone (Tunnel -> USB)
	go func() {
		buf := make([]byte, 2048)
		for {
			select {
			case <-r.ctx.Done():
				return
			default:
				n, err := r.tun.Read(buf)
				if err != nil {
					if err != io.EOF {
						errChan <- fmt.Errorf("relay: tun read error: %w", err)
					}
					return
				}

				// macOS utun header is 4 bytes [0 0 0 2] for IPv4. Strip it.
				if n < 4 {
					continue
				}
				rawIP := buf[4:n]

				// Construct Ethernet Header (L2)
				// [DstMAC(6)] [SrcMAC(6)] [Type(2)]
				eth := make([]byte, 14+len(rawIP))
				copy(eth[0:6], r.phoneMAC)
				copy(eth[6:12], []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x01})
				binary.BigEndian.PutUint16(eth[12:14], 0x0800) // IPv4
				copy(eth[14:], rawIP)

				// Wrap in RNDIS (L1-ish)
				pkt := rndis.EncapsulatePacket(eth)
				_, err = r.usbOut.Write(pkt)
				if err != nil {
					errChan <- fmt.Errorf("relay: usb write error: %w", err)
					return
				}
				log.Info().Str("component", "relay").Int("bytes", len(pkt)).Msg("Sent packet to phone")
			}
		}
	}()


	// Phone -> Mac (USB -> Tunnel)
	go func() {
		buf := make([]byte, 16384)
		for {
			select {
			case <-r.ctx.Done():
				return
			default:
				n, err := r.usbIn.Read(buf)
				if err != nil {
					errChan <- fmt.Errorf("relay: usb read error: %w", err)
					return
				}

				offset := 0
				for offset < n {
					msg := buf[offset:]
					if len(msg) < 8 {
						break
					}
					msgType := binary.LittleEndian.Uint32(msg[0:4])
					msgLen := int(binary.LittleEndian.Uint32(msg[4:8]))
					
					if msgType == rndis.MsgPacket && msgLen > 44 {
						ethPkt, err := rndis.DecapsulatePacket(msg[:msgLen])
						if err == nil && len(ethPkt) > 14 {
							// Strip Ethernet header (14 bytes)
							rawIP := ethPkt[14:]
							
							// Prepend macOS utun header [0 0 0 2]
							outBuf := make([]byte, 4+len(rawIP))
							binary.BigEndian.PutUint32(outBuf[0:4], 2)
							copy(outBuf[4:], rawIP)
							
							_, _ = r.tun.Write(outBuf)
							log.Info().Str("component", "relay").Int("bytes", len(outBuf)).Msg("Received packet from phone")
						}
					}
					
					if msgLen == 0 {
						break
					}
					offset += msgLen
				}
			}
		}
	}()


	log.Info().Str("component", "relay").Msg("Bidirectional packet relay started")
	
	// Wait for error or shutdown
	select {
	case err := <-errChan:
		r.cancel()
		return err
	case <-r.ctx.Done():
		return nil
	}
}

// Stop shuts down the relay.
func (r *Relay) Stop() {
	r.cancel()
}
