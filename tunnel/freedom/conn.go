package freedom

import (
	"net"

	"github.com/xxf098/lite-proxy/tunnel"
)

type Conn struct {
	net.Conn
}

func (c *Conn) Metadata() *tunnel.Metadata {
	return nil
}
