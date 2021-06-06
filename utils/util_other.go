// +build !linux

package utils

import (
	"context"
	"net"
)

func Listen(ctx context.Context, network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}
