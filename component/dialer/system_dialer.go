package dialer

import (
	"errors"
	"net"
	"syscall"
	"time"

	"github.com/xxf098/lite-proxy/log"
)

type controller func(network, address string, fd uintptr) error

var controllers []controller

func Dialer() (*net.Dialer, error) {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}
	if len(controllers) > 0 {
		dialer.Control = func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {

				for _, ctl := range controllers {
					if err := ctl(network, address, fd); err != nil {
						// errors.New("failed to apply external controller").Base(err).WriteToLog(session.ExportIDToError(ctx))
						log.E("failed to apply external controller")
						continue
					}
				}
			})
		}
	}
	return dialer, nil
}

func RegisterDialerController(ctl func(network, address string, fd uintptr) error) error {
	if ctl == nil {
		return errors.New("nil listener controller")
	}

	controllers = append(controllers, ctl)
	return nil
}
