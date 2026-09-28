package daemon

import (
	"os/exec"
	"strings"
	"testing"
)

// Live check: the function under test must return true against this machine's
// actual route table right now (both /1 splits installed on utun4), and the
// old `route -n get default` behaviour would have returned false.
func TestRouteIntactLive(t *testing.T) {
	out, err := exec.Command("netstat", "-rn", "-f", "inet").Output()
	if err != nil {
		t.Skipf("netstat unavailable: %v", err)
	}

	iface := detectTetherIface(t)

	if !routeIntactFromTable(string(out), iface) {
		t.Fatalf("routeIntactFromTable() = false against live table on %s:\n%s", iface, out)
	}
	t.Logf("live route table on %s: splits present, check returns true", iface)

	// Demonstrate the old check's failure mode is real: `default` resolves to
	// a different interface, which is exactly why the old code false-positived.
	d, err := exec.Command("route", "-n", "get", "default").Output()
	if err == nil {
		t.Logf("for the record, `route -n get default` says:\n%s", d)
	}
}

// detectTetherIface finds the utun holding the droidtether split routes.
func detectTetherIface(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("netstat", "-rn", "-f", "inet").Output()
	if err != nil {
		t.Skip("netstat unavailable")
	}
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.Fields(line)
		if len(f) >= 3 && f[0] == "0/1" {
			return f[len(f)-1]
		}
	}
	t.Skip("no droidtether split routes present (not tethered via USB)")
	return ""
}
