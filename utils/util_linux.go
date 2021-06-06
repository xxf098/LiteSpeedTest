// +build linux

package utils

import (
	"context"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func Listen(ctx context.Context, network, address string) (net.Listener, error) {
	lc := &net.ListenConfig{
		Control: func(network, address string, conn syscall.RawConn) error {
			return conn.Control(func(fd uintptr) {
				_ = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEADDR|unix.SO_REUSEPORT, 1)
			})
		},
	}

	conn, err := lc.Listen(ctx, network, address)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
