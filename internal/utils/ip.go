package utils

import (
	"fmt"
	"net/netip"
)

func ParseIPs(values []string) ([]netip.Addr, error) {
	addrs := make([]netip.Addr, 0, len(values))
	for _, value := range values {
		addr, err := netip.ParseAddr(value)
		if err != nil {
			return nil, fmt.Errorf("invalid IP address %q", value)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}
