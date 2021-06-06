// +build linux

package dialer

import (
	"syscall"

	"github.com/xxf098/lite-proxy/common"
	"golang.org/x/sys/unix"
)

func setReusePort(fd uintptr) error {
	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		return common.NewError("failed to set SO_REUSEPORT").Base(err)
	}
	return nil
}
