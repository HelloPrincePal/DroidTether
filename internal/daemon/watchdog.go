package daemon

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/princePal/droidtether/internal/tun"
)

// Config watchdog: macOS standby can silently remove the injected split
// default routes (0.0.0.0/1, 128.0.0.0/1), the dynamic-store DNS entries, and
// even the utun IP configuration — while the USB/RNDIS session survives the
// wake, so no new DHCP event fires and the original injection in OnDHCP never
// runs again. The result is a tethered-but-dead network: interface up, pipe
// to the phone alive, but no route past it and no name resolution.
//
// This watchdog verifies every few seconds that the injected config is still
// in place and re-applies whatever is missing. All checks are local
// syscalls/commands (route/ifconfig/scutil) — zero carrier data.

// routeIntact reports whether the split default routes (0.0.0.0/1 and
// 128.0.0.0/1) are still installed on the given utun interface.
//
// It must NOT check `route -n get default`: DroidTether deliberately injects
// the two /1 split routes instead of a real default route, so a `default`
// query resolves to whatever real interface owns it (en0/Wi-Fi, en1/Ethernet)
// and the check fails forever — a permanent false positive that re-injected
// the routes every tick and spammed route(8) errors. Verify the routes we
// actually install.
func routeIntact(ifaceName string) bool {
	out, err := exec.Command("netstat", "-rn", "-f", "inet").Output()
	if err != nil {
		return false
	}
	return routeIntactFromTable(string(out), ifaceName)
}

// routeIntactFromTable holds the parsing logic, separated from the exec so it
// is unit-testable.
func routeIntactFromTable(table, ifaceName string) bool {
	// Both halves must be present on our interface; one alone leaves a
	// black hole for the other half of the IPv4 space.
	for _, prefix := range []string{"0/1", "128.0/1"} {
		found := false
		for _, line := range strings.Split(table, "\n") {
			fields := strings.Fields(line)
			if len(fields) < 3 || fields[0] != prefix {
				continue
			}
			if fields[len(fields)-1] == ifaceName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// dnsIntact reports whether DroidTether's DNS is healthy end-to-end:
// configd still considers the droidtether service the primary (Global/IPv4
// PrimaryService) and mDNSResponder's first resolver carries both our
// nameserver and the RequestA bit. Checking the derived resolver via
// `scutil --dns` instead of the raw store dict matters: a raw
// State:/Network/Global/DNS write can show __FLAGS__ with the A bit set
// while mDNSResponder still only has a Supplemental-only resolver and
// suppresses every A/AAAA query (getaddrinfo fails despite a "good" store).
func dnsIntact() bool {
	cmd := exec.Command("scutil")
	cmd.Stdin = strings.NewReader("show State:/Network/Global/IPv4\nquit\n")
	out, err := cmd.Output()
	if err != nil || !strings.Contains(string(out), "PrimaryService : droidtether") {
		return false
	}

	dns, err := exec.Command("scutil", "--dns").Output()
	if err != nil {
		return false
	}
	s := string(dns)
	start := strings.Index(s, "resolver #1")
	if start < 0 {
		return false
	}
	block := s[start:]
	if end := strings.Index(block, "resolver #2"); end >= 0 {
		block = block[:end]
	}
	return strings.Contains(block, "8.8.8.8") && strings.Contains(block, "Request A records")
}

// ifaceConfigIntact reports whether the utun interface still holds the DHCP
// lease address (ifconfig <name> <local> <remote> mtu <val>).
func ifaceConfigIntact(ifaceName, localIP string) bool {
	out, err := exec.Command("ifconfig", ifaceName).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "inet "+localIP)
}

// startConfigWatchdog launches the watchdog goroutine for an active session.
// It is bound to ctx, so it stops together with the relay session.
func (d *Daemon) startConfigWatchdog(ctx context.Context, iface tun.Interface, localIP, gateway, mtu string) {
	interval := time.Duration(d.cfg.Route.WatchdogIntervalS) * time.Second
	if interval <= 0 {
		interval = 10 * time.Second
	}

	log.Info().
		Str("component", "watchdog").
		Dur("interval", interval).
		Msg("Config watchdog started — will re-apply route/DNS/interface if macOS removes them (e.g. standby)")

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}

			name := iface.Name()

			// Interface config first: route/DNS injection and the relay
			// itself are useless without the lease address.
			if !ifaceConfigIntact(name, localIP) {
				log.Warn().
					Str("component", "watchdog").
					Str("interface", name).
					Msg("Interface IP config was removed — re-applying ifconfig...")
				if err := iface.Configure(localIP, gateway, mtu); err != nil {
					log.Error().Str("component", "watchdog").Err(err).Msg("Failed to re-apply interface config")
					continue
				}
			}

			if !d.cfg.Route.SetDefaultRoute {
				continue
			}

			if !routeIntact(name) {
				log.Warn().
					Str("component", "watchdog").
					Str("interface", name).
					Msg("Default route no longer resolves through DroidTether — re-injecting split default routes...")
				if err := iface.SetDefaultRoute(gateway); err != nil {
					log.Error().Str("component", "watchdog").Err(err).Msg("Failed to re-inject default route")
				}
			}

			if !dnsIntact() {
				log.Warn().
					Str("component", "watchdog").
					Msg("DroidTether DNS resolver broken (primary election or RequestA lost) — re-registering service and re-injecting DNS...")
				if err := iface.RegisterService(localIP, gateway); err != nil {
					log.Error().Str("component", "watchdog").Err(err).Msg("Failed to re-register service with configd")
				}
				if err := iface.SetDNS([]string{"8.8.8.8", gateway}); err != nil {
					log.Error().Str("component", "watchdog").Err(err).Msg("Failed to re-inject DNS")
				}
			}
		}
	}()
}

// startLinkWatchdog probes the phone's gateway through the tunnel. When
// Android disables USB tethering (e.g. the phone screen locks), the phone
// re-enumerates without the RNDIS config and the relay's writes to the
// dead handle keep "succeeding" — the session becomes a zombie that holds
// the watcher's active-device slot, so a later re-enable never re-attaches.
// Three consecutive failed probes end the session; the watcher then sees
// the device again (or waits for the user to re-enable tethering, covered
// by the handshake-failure notification).
func (d *Daemon) startLinkWatchdog(ctx context.Context, relay *Relay, gateway, client string) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		fails := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := exec.Command("ping", "-c", "1", "-t", "3", "-S", client, gateway).Run()
				if err == nil {
					fails = 0
					continue
				}
				fails++
				if fails < 3 {
					log.Warn().Str("component", "watchdog").
						Int("fails", fails).
						Msg("Gateway ping failed")
					continue
				}
				log.Error().Str("component", "watchdog").
					Msg("Phone left tethering state; ending zombie session")
				relay.Stop()
				return
			}
		}
	}()
}
