package proxy

import (
	"context"
	"fmt"
	"net"

	"github.com/xxf098/lite-proxy/common"
	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/tunnel"
)

type ContextDialer interface {
	// Dial connects to the given address via the proxy.
	DialContext(ctx context.Context, m *C.Metadata) (c net.Conn, err error)
}

type Creator func(link string) (ContextDialer, error)

var creators = make(map[string]Creator)

func RegisterContextDialerCreator(name string, c Creator) {
	creators[name] = c
}

func GetContextDialerCreator(name string) (Creator, error) {
	if c, ok := creators[name]; ok {
		return c, nil
	}
	return nil, common.NewError("unknown context dialer name " + string(name))
}

type Client struct {
	ctx    context.Context
	dialer ContextDialer
}

func (c Client) DialConn(addr *tunnel.Address, _ tunnel.Tunnel) (net.Conn, error) {
	networkType := C.TCP
	if addr.NetworkType == "udp" {
		networkType = C.UDP
	}
	meta := &C.Metadata{
		NetWork:  networkType,
		Type:     0,
		SrcPort:  "",
		AddrType: int(addr.AddressType),
		DstPort:  fmt.Sprintf("%d", addr.Port),
	}
	switch addr.AddressType {
	case tunnel.IPv4:
	case tunnel.IPv6:
		meta.DstIP = addr.IP
	case tunnel.DomainName:
		meta.Host = addr.DomainName
	}
	return c.dialer.DialContext(c.ctx, meta)
}

func (c Client) Close() error {
	return nil
}

func (c *Client) Dial(network, address string) (net.Conn, error) {
	addr, err := tunnel.NewAddressFromAddr(network, address)
	if err != nil {
		return nil, err
	}
	return c.DialConn(addr, nil)
}

func NewClient(ctx context.Context, dialer ContextDialer) Client {
	return Client{
		ctx:    ctx,
		dialer: dialer,
	}
}
