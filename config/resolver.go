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
	defaultResolver = dns.DefaultResolver()
	// defaultResolver = nil
}
