package vmess

import (
	"context"
	"fmt"
	"net"

	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/tunnel"
)

type Client struct {
	ctx   context.Context
	vmess *outbound.Vmess
}

func (c Client) DialConn(addr *tunnel.Address, _ tunnel.Tunnel) (net.Conn, error) {
	meta := &C.Metadata{
		NetWork:  0,
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
	return c.vmess.DialContext(c.ctx, meta)
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

func NewClient(ctx context.Context, vmess *outbound.Vmess) Client {
	return Client{
		ctx:   ctx,
		vmess: vmess,
	}
}
