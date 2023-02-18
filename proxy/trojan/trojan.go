package trojan

import (
	"context"
	"fmt"
	"net"

	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/tunnel"
)

type Client struct {
	ctx    context.Context
	trojan *outbound.Trojan
}

func (c Client) DialConn(addr *tunnel.Address, _ tunnel.Tunnel) (net.Conn, error) {
	meta := &C.Metadata{
		NetWork: 0,
		Type:    0,
		SrcPort: "",
		DstPort: fmt.Sprintf("%d", addr.Port),
	}
	switch addr.AddressType {
	case tunnel.IPv4:
	case tunnel.IPv6:
		meta.DstIP = addr.IP
	case tunnel.DomainName:
		meta.Host = addr.DomainName
	}
	return c.trojan.DialContext(c.ctx, meta)
}

func (c Client) Close() error {
	return nil
}

func NewClient(ctx context.Context, trojan *outbound.Trojan) Client {
	return Client{
		ctx:    ctx,
		trojan: trojan,
	}
}
