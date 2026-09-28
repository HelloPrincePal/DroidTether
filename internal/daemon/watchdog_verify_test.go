package daemon

import "testing"

// Guards the exact regression that caused the 10s re-injection loop: the check
// must validate the split /1 routes droidtether actually installs, and must not
// depend on `route -n get default` (which resolves to en0/Wi-Fi on macOS 27).
func TestRouteIntactParsesNetstatColumns(t *testing.T) {
	const live = `Routing tables

Internet:
Destination        Gateway            Flags       Netif Expire
0/1                utun4              UScg        utun4
default            192.168.35.25      UGScg       en0
127                127.0.0.1          UCS            lo0
128.0/1            utun4              USc         utun4
169.254            link#14            UCS            en0
192.168.60.85      192.168.60.3       UH           utun4
`

	cases := []struct {
		name  string
		table string
		want  bool
	}{
		{"both splits on utun4", live, true},
		{"only 0/1 present is a black hole", `0/1                utun4              UScg        utun4
default            192.168.35.25      UGScg       en0
`, false},
		{"only 128.0/1 present is a black hole", `default            192.168.35.25      UGScg       en0
128.0/1            utun4              USc         utun4
`, false},
		{"splits on the wrong interface", `0/1                en0                UScg        en0
128.0/1            en0                USc         en0
`, false},
		{"no splits at all", `default            192.168.35.25      UGScg       en0
`, false},
		{"empty table", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := routeIntactFromTable(tc.table, "utun4"); got != tc.want {
				t.Fatalf("routeIntactFromTable() = %v, want %v", got, tc.want)
			}
		})
	}
}
