package config

import (
	"errors"
	"net"

	"github.com/xxf098/lite-proxy/dns"
)

var defaultResolver *dns.Resolver

func resolveIP(host string) (string, error) {
	ipAddr := net.ParseIP(host)
	if ipAddr != nil {
		return host, nil
	}
	if defaultResolver != nil {
		ipAddr, err := defaultResolver.ResolveIP(host)
		if err != nil {
			return "", err
		}
		return ipAddr.String(), nil
	}
	return "", errors.New("resolver not found")
}

func init() {
	c := dns.Config{
		Main: []dns.NameServer{
			{
				Net:  "udp",
				Addr: "223.5.5.5:53",
			},
			{
				Net:  "udp",
				Addr: "8.8.8.8:53",
			},
		},
	}
	defaultResolver = dns.NewResolver(c)
	// defaultResolver = nil
}
