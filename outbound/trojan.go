package outbound

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/stats"
	"github.com/xxf098/lite-proxy/transport/dialer"
	"github.com/xxf098/lite-proxy/transport/gun"
	"github.com/xxf098/lite-proxy/transport/trojan"
	"golang.org/x/net/http2"
)

type Trojan struct {
	*Base
	instance *trojan.Trojan
	option   *TrojanOption

	// for gun mux
	gunTLSConfig *tls.Config
	gunConfig    *gun.Config
	transport    *http2.Transport
}

type TrojanOption struct {
	Name           string      `proxy:"name,omitempty"`
	Server         string      `proxy:"server"`
	Port           int         `proxy:"port"`
	Password       string      `proxy:"password"`
	ALPN           []string    `proxy:"alpn,omitempty"`
	SNI            string      `proxy:"sni,omitempty"`
	SkipCertVerify bool        `proxy:"skip-cert-verify,omitempty"`
	UDP            bool        `proxy:"udp,omitempty"`
	Remarks        string      `proxy:"remarks,omitempty"`
	Network        string      `proxy:"network,omitempty"`
	GrpcOpts       GrpcOptions `proxy:"grpc-opts,omitempty"`
	WSOpts         WSOptions   `proxy:"ws-opts,omitempty"`
}

func (t *Trojan) plainStream(c net.Conn) (net.Conn, error) {
	if t.option.Network == "ws" {
		host, port, _ := net.SplitHostPort(t.addr)
		wsOpts := &trojan.WebsocketOption{
			Host: host,
			Port: port,
			Path: t.option.WSOpts.Path,
		}

		if t.option.SNI != "" {
			wsOpts.Host = t.option.SNI
		}

		if len(t.option.WSOpts.Headers) != 0 {
			header := http.Header{}
			for key, value := range t.option.WSOpts.Headers {
				header.Add(key, value)
			}
			wsOpts.Headers = header
		}

		return t.instance.StreamWebsocketConn(c, wsOpts)
	}

	return t.instance.StreamConn(c)
}

func (t *Trojan) StreamConn(c net.Conn, metadata *C.Metadata) (net.Conn, error) {
	c, err := t.plainStream(c)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", t.addr, err)
	}

	err = t.instance.WriteHeader(c, trojan.CommandTCP, serializesSocksAddr(metadata))
	return c, err
}

func (t *Trojan) DialContext(ctx context.Context, metadata *C.Metadata) (net.Conn, error) {
	// gun transport
	if t.transport != nil {
		c, err := gun.StreamGunWithTransport(t.transport, t.gunConfig)
		if err != nil {
			return nil, err
		}

		if err = t.instance.WriteHeader(c, trojan.CommandTCP, serializesSocksAddr(metadata)); err != nil {
			c.Close()
			return nil, err
		}
		sc := stats.NewStatsConn(c)
		return t.StreamConn(sc, metadata)
	}

	c, err := dialer.DialContext(ctx, "tcp", t.addr)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", t.addr, err)
	}
	tcpKeepAlive(c)
	sc := stats.NewStatsConn(c)
	return t.StreamConn(sc, metadata)
}

// TODO: grpc transport
func (t *Trojan) DialUDP(metadata *C.Metadata) (_ net.PacketConn, err error) {
	var c net.Conn

	// grpc transport
	if t.transport != nil {
		c, err = gun.StreamGunWithTransport(t.transport, t.gunConfig)
		if err != nil {
			return nil, fmt.Errorf("%s connect error: %w", t.addr, err)
		}
	} else {
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

	t := &Trojan{
		Base: &Base{
			name: option.Name,
			addr: addr,
			udp:  option.UDP,
		},
		instance: trojan.New(tOption),
		option:   option,
	}

	// if option.Network == "grpc" {
	// 	dialFn := func(network, addr string) (net.Conn, error) {
	// 		c, err := dialer.DialContext(context.Background(), "tcp", t.addr)
	// 		if err != nil {
	// 			return nil, fmt.Errorf("%s connect error: %v", t.addr, err)
	// 		}
	// 		tcpKeepAlive(c)
	// 		return c, nil
	// 	}

	// 	tlsConfig := &tls.Config{
	// 		NextProtos:         option.ALPN,
	// 		MinVersion:         tls.VersionTLS12,
	// 		InsecureSkipVerify: tOption.SkipCertVerify,
	// 		ServerName:         tOption.ServerName,
	// 	}
	// 	t.transport = gun.NewHTTP2Client(dialFn, tlsConfig)
	// 	t.gunTLSConfig = tlsConfig
	// 	t.gunConfig = &gun.Config{
	// 		ServiceName: option.GrpcOpts.GrpcServiceName,
	// 		Host:        tOption.ServerName,
	// 	}
	// }

	return t, nil
}
