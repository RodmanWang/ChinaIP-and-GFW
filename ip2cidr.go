package main

import (
	"fmt"
	"net"
	"strings"
)

// parseCIDRs parses a list of CIDR lines. Empty lines and full-line comments
// are ignored; malformed non-empty entries fail the whole operation so that
// an incomplete database is never silently published.
func parseCIDRs(txtList []string) ([]*net.IPNet, error) {
	networks := make([]*net.IPNet, 0, len(txtList))
	for i, raw := range txtList {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		_, network, err := net.ParseCIDR(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid CIDR %q: %w", i+1, line, err)
		}
		networks = append(networks, network)
	}
	return networks, nil
}
