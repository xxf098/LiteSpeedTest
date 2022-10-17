package outbound

import (
	"context"
	"net"

	"github.com/xxf098/lite-proxy/common"
	C "github.com/xxf098/lite-proxy/constant"
)

type Base struct {
	name string
	addr string
	udp  bool
}

type BasicOption struct {
	Interface   string `proxy:"interface-name,omitempty" group:"interface-name,omitempty"`
	RoutingMark int    `proxy:"routing-mark,omitempty" group:"routing-mark,omitempty"`
}

type ContextDialer interface {
	DialContext(ctx context.Context, m *C.Metadata) (c net.Conn, err error)
}

type Dialer interface {
	ContextDialer
	DialUDP(m *C.Metadata) (net.PacketConn, error)
}

type Creator func(link string) (Dialer, error)

var creators = make(map[string]Creator)

func RegisterDialerCreator(name string, c Creator) {
	creators[name] = c
}

func GetDialerCreator(name string) (Creator, error) {
	if c, ok := creators[name]; ok {
		return c, nil
	}
	return nil, common.NewError("unknown context dialer name " + string(name))
}
