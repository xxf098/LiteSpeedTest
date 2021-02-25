package dialer

import (
	"context"
	"net"
	"syscall"

	"github.com/xxf098/lite-proxy/common"
	"github.com/xxf098/lite-proxy/log"
)

var (
	effectiveListener = DefaultListener{}
)

type DefaultListener struct {
	controllers []controller
}

func getControlFunc(ctx context.Context, controllers []controller) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		return c.Control(func(fd uintptr) {

			setReusePort(fd) // nolint: staticcheck

			for _, controller := range controllers {
				if err := controller(network, address, fd); err != nil {
					log.E("failed to apply external controller")
					continue
				}
			}
		})
	}
}

// func setReusePort(fd uintptr) error {
// 	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
// 		return common.NewError("failed to set SO_REUSEPORT")
// 	}
// 	return nil
// }

func (dl *DefaultListener) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	var lc net.ListenConfig

	lc.Control = getControlFunc(ctx, dl.controllers)

	return lc.ListenPacket(ctx, network, address)
}

func RegisterListenerController(controller func(network, address string, fd uintptr) error) error {
	if controller == nil {
		return common.NewError("nil listener controller")
	}

	effectiveListener.controllers = append(effectiveListener.controllers, controller)
	return nil
}
