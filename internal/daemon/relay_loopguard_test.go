package daemon

import (
	"bytes"
	"testing"
)

func newTestRelay(phoneMAC []byte) *Relay {
	return &Relay{phoneMAC: append([]byte(nil), phoneMAC...)}
}

var (
	usbMAC  = []byte{0x3a, 0x58, 0x68, 0x37, 0xd3, 0xe8}
	wifiMAC = []byte{0xda, 0x7c, 0x6a, 0x44, 0xbe, 0x37}
)

func TestAutoDetectMACIgnoresOwnClientMAC(t *testing.T) {
	r := newTestRelay(usbMAC)

	// A reflected frame carries our own source MAC. Adopting it would make us
	// address every subsequent frame to ourselves.
	for i := 0; i < 50; i++ {
		r.autoDetectMAC(clientMAC)
	}

	if !bytes.Equal(r.phoneMAC, usbMAC) {
		t.Fatalf("phoneMAC was corrupted to our own client MAC: %x", r.phoneMAC)
	}
	if r.macChanges != 0 {
		t.Fatalf("macChanges = %d, want 0", r.macChanges)
	}
}

func TestAutoDetectMACRequiresConsecutiveSightings(t *testing.T) {
	r := newTestRelay(usbMAC)

	// A single stray frame from another interface must not re-address the relay.
	r.autoDetectMAC(wifiMAC)
	if !bytes.Equal(r.phoneMAC, usbMAC) {
		t.Fatalf("single stray sighting changed the MAC to %x", r.phoneMAC)
	}

	// A different candidate resets the streak rather than accumulating. On a
	// fresh relay, feed threshold-1 sightings of A, one of B, then threshold-1
	// more of A. Only threshold-1 *consecutive* A's are ever seen, so no switch.
	r2 := newTestRelay(usbMAC)
	other := []byte{0x02, 0xaa, 0xbb, 0xcc, 0xdd, 0xee}
	for i := 0; i < macAdoptThreshold-1; i++ {
		r2.autoDetectMAC(wifiMAC)
	}
	r2.autoDetectMAC(other)
	for i := 0; i < macAdoptThreshold-1; i++ {
		r2.autoDetectMAC(wifiMAC)
	}
	if !bytes.Equal(r2.phoneMAC, usbMAC) {
		t.Fatalf("non-consecutive sightings changed the MAC to %x", r2.phoneMAC)
	}

	// A genuine sustained run is adopted: one more consecutive sighting
	// completes the threshold-1 streak.
	r2.autoDetectMAC(wifiMAC)
	if !bytes.Equal(r2.phoneMAC, wifiMAC) {
		t.Fatalf("sustained change not adopted: got %x want %x", r2.phoneMAC, wifiMAC)
	}
	if r2.macChanges != 1 {
		t.Fatalf("macChanges = %d, want 1", r2.macChanges)
	}
}

func TestAutoDetectMACCurrentMACCancelsCandidate(t *testing.T) {
	r := newTestRelay(usbMAC)

	// Build a streak toward switching...
	for i := 0; i < macAdoptThreshold-1; i++ {
		r.autoDetectMAC(wifiMAC)
	}
	// ...then see the real USB MAC again. The candidate must be discarded, so
	// a later single stray frame cannot tip it over the threshold.
	r.autoDetectMAC(usbMAC)
	r.autoDetectMAC(wifiMAC)
	if !bytes.Equal(r.phoneMAC, usbMAC) {
		t.Fatalf("interrupted streak still switched the MAC to %x", r.phoneMAC)
	}
}

func TestIsLoopFrame(t *testing.T) {
	frame := func(src []byte) []byte {
		f := make([]byte, 14)
		copy(f[0:6], []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
		copy(f[6:12], src)
		return f
	}

	r := newTestRelay(usbMAC)

	if !r.isLoopFrame(frame(clientMAC)) {
		t.Error("frame with our own source MAC was not detected as a loop")
	}
	if r.isLoopFrame(frame(usbMAC)) {
		t.Error("legitimate phone frame misdetected as a loop")
	}
	if r.isLoopFrame(frame(clientMAC[:4])) {
		t.Error("truncated frame must not be treated as a loop")
	}
}
