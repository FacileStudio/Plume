package urlsafe

import (
	"net"
	"testing"
)

// TestPrivateCIDRTableParses checks that every CIDR in the guard table parses
// and has a probe address in the test — a malformed entry is skipped silently
// at check time, so the table itself must be proven valid.
func TestPrivateCIDRTableParses(t *testing.T) {
	probes := probeAddresses()

	for _, cidr := range privateCIDRs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			t.Errorf("guard table entry %q does not parse: %v", cidr, err)
		}
		if _, seen := probes[cidr]; !seen {
			t.Errorf("guard table entry %q has no probe address in the test", cidr)
		}
	}
}

// TestReservedRangesAreRefused checks a representative address from each
// reserved range is refused while public addresses pass.
func TestReservedRangesAreRefused(t *testing.T) {
	for cidr, probe := range probeAddresses() {
		if !isPrivateIP(net.ParseIP(probe)) {
			t.Errorf("isPrivateIP(%s) = false, want true (range %s)", probe, cidr)
		}
	}

	public := []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111"}
	for _, probe := range public {
		if isPrivateIP(net.ParseIP(probe)) {
			t.Errorf("isPrivateIP(%s) = true, want false", probe)
		}
	}
}

// probeAddresses maps each guard-table range to one address inside it.
func probeAddresses() map[string]string {
	return map[string]string{
		"0.0.0.0/8":      "0.0.0.1",
		"127.0.0.0/8":    "127.0.0.1",
		"10.0.0.0/8":     "10.1.2.3",
		"172.16.0.0/12":  "172.16.5.5",
		"192.168.0.0/16": "192.168.1.1",
		"169.254.0.0/16": "169.254.9.9",
		"100.64.0.0/10":  "100.64.0.1",
		"198.18.0.0/15":  "198.19.0.1",
		"224.0.0.0/4":    "224.0.0.1",
		"240.0.0.0/4":    "240.0.0.1",
		"fc00::/7":       "fc00::1",
		"fe80::/10":      "fe80::1",
		"::1/128":        "::1",
	}
}
