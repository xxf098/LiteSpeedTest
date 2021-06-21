package outbound

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/stats"
	"github.com/xxf098/lite-proxy/transport/dialer"
	"github.com/xxf098/lite-proxy/transport/trojan"
)

type Trojan struct {
	*Base
	instance *trojan.Trojan
}

type TrojanOption struct {
	Name           string   `proxy:"name"`
	Server         string   `proxy:"server"`
	Port           int      `proxy:"port"`
	Password       string   `proxy:"password"`
	ALPN           []string `proxy:"alpn,omitempty"`
	SNI            string   `proxy:"sni,omitempty"`
	SkipCertVerify bool     `proxy:"skip-cert-verify,omitempty"`
	UDP            bool     `proxy:"udp,omitempty"`
	Remarks        string   `proxy:"remarks,omitempty"`
}

func (t *Trojan) StreamConn(c net.Conn, metadata *C.Metadata) (net.Conn, error) {
	c, err := t.instance.StreamConn(c)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", t.addr, err)
	}

	err = t.instance.WriteHeader(c, trojan.CommandTCP, serializesSocksAddr(metadata))
	return c, err
}

func (t *Trojan) DialContext(ctx context.Context, metadata *C.Metadata) (net.Conn, error) {
	c, err := dialer.DialContext(ctx, "tcp", t.addr)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", t.addr, err)
	}
	tcpKeepAlive(c)
	sc := stats.NewStatsConn(c)
	return t.StreamConn(sc, metadata)
}

func (t *Trojan) DialUDP(metadata *C.Metadata) (net.PacketConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), tcpTimeout)
	defer cancel()
	c, err := dialer.DialContext(ctx, "tcp", t.addr)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", t.addr, err)
	}
	tcpKeepAlive(c)
	c, err = t.instance.StreamConn(c)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", t.addr, err)
	}

	err = t.instance.WriteHeader(c, trojan.CommandUDP, serializesSocksAddr(metadata))
	if err != nil {
		return nil, err
	}

	pc := t.instance.PacketConn(c)
	return pc, err
}

func (t *Trojan) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type": "Trojan",
	})
}

func NewTrojan(option *TrojanOption) (*Trojan, error) {
	addr := net.JoinHostPort(option.Server, strconv.Itoa(option.Port))

	tOption := &trojan.Option{
		Password:           option.Password,
		ALPN:               option.ALPN,
		ServerName:         option.Server,
		SkipCertVerify:     option.SkipCertVerify,
		ClientSessionCache: getClientSessionCache(),
	}

	if option.SNI != "" {
		tOption.ServerName = option.SNI
	}

	return &Trojan{
		Base: &Base{
			name: option.Name,
			addr: addr,
			udp:  option.UDP,
		},
		instance: trojan.New(tOption),
	}, nil
}
