package net

import "net"

var privateNetworks []*net.IPNet

func init() {
	for _, cidr := range []string{
		// RFC 1918: private IPv4 networks
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		// RFC 4193: IPv6 ULAs
		"fc00::/7",
		// RFC 6598: reserved prefix for CGNAT
		"100.64.0.0/10",
	} {
		_, subnet, _ := net.ParseCIDR(cidr)
		privateNetworks = append(privateNetworks, subnet)
	}
}

func IsPrivateAddress(ip net.IP) bool {
	for _, network := range privateNetworks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
