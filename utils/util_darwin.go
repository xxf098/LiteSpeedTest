package utils

import "net"

func Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}
