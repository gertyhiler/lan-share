package netutil

import (
	"net"
	"os"
	"slices"
	"strings"
)

// LANIPv4 returns sorted non-loopback IPv4 addresses for the host name (same idea as Python _get_local_ips).
func LANIPv4() []string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return nil
	}
	addrs, err := net.LookupIP(name)
	if err != nil {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	for _, ip := range addrs {
		v4 := ip.To4()
		if v4 == nil {
			continue
		}
		s := v4.String()
		if strings.HasPrefix(s, "127.") {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	slices.Sort(out)
	return out
}
