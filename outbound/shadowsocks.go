package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/xxf098/lite-proxy/stats"
	"github.com/xxf098/lite-proxy/transport/dialer"
	"github.com/xxf098/lite-proxy/transport/socks5"

	"github.com/Dreamacro/go-shadowsocks2/core"
	"github.com/xxf098/lite-proxy/common/structure"
	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/log"
)

type ShadowSocks struct {
	*Base
	cipher core.Cipher

	// obfs
	obfsMode   string
	obfsOption *simpleObfsOption
	// v2rayOption *v2rayObfs.Option
}

type ShadowSocksOption struct {
	Name       string                 `proxy:"name,omitempty"`
	Server     string                 `proxy:"server"`
	Port       int                    `proxy:"port"`
	Password   string                 `proxy:"password"`
	Cipher     string                 `proxy:"cipher"`
	UDP        bool                   `proxy:"udp,omitempty"`
	Plugin     string                 `proxy:"plugin,omitempty"`
	PluginOpts map[string]interface{} `proxy:"plugin-opts,omitempty"`
	Remarks    string                 `proxy:"remarks,omitempty"`
}

type simpleObfsOption struct {
	Mode string `obfs:"mode"`
	Host string `obfs:"host,omitempty"`
}

type v2rayObfsOption struct {
	Mode           string            `obfs:"mode"`
	Host           string            `obfs:"host,omitempty"`
	Path           string            `obfs:"path,omitempty"`
	TLS            bool              `obfs:"tls,omitempty"`
	Headers        map[string]string `obfs:"headers,omitempty"`
	SkipCertVerify bool              `obfs:"skip-cert-verify,omitempty"`
	Mux            bool              `obfs:"mux,omitempty"`
}

func (ss *ShadowSocks) StreamConn(c net.Conn, metadata *C.Metadata) (net.Conn, error) {
	// switch ss.obfsMode {
	// case "tls":
	// 	c = obfs.NewTLSObfs(c, ss.obfsOption.Host)
	// case "http":
	// 	_, port, _ := net.SplitHostPort(ss.addr)
	// 	c = obfs.NewHTTPObfs(c, ss.obfsOption.Host, port)
	// case "websocket":
	// 	var err error
	// 	c, err = v2rayObfs.NewV2rayObfs(c, ss.v2rayOption)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("%s connect error: %w", ss.addr, err)
	// 	}
	// }
	c = ss.cipher.StreamConn(c)
	_, err := c.Write(serializesSocksAddr(metadata))
	return c, err
}

func (ss *ShadowSocks) DialContext(ctx context.Context, metadata *C.Metadata) (net.Conn, error) {
	log.I("start dial from", ss.addr, "to", metadata.RemoteAddress())
	c, err := dialer.DialContext(ctx, "tcp", ss.addr)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", ss.addr, err)
	}
	tcpKeepAlive(c)
	log.I("start StreamConn from", ss.addr, "to", metadata.RemoteAddress())
	sc := stats.NewStatsConn(c)
	return ss.StreamConn(sc, metadata)
}

func (ss *ShadowSocks) DialUDP(metadata *C.Metadata) (net.PacketConn, error) {
	pc, err := dialer.ListenPacket("udp", "")
	if err != nil {
		return nil, err
	}

	addr, err := resolveUDPAddr("udp", ss.addr)
	if err != nil {
		return nil, err
	}

	spc := stats.NewStatsPacketConn(pc)
	pc = ss.cipher.PacketConn(spc)
	return &ssPacketConn{PacketConn: pc, rAddr: addr}, nil
}

func (ss *ShadowSocks) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type": "shadowsocks",
	})
}

func NewShadowSocks(option *ShadowSocksOption) (*ShadowSocks, error) {
	addr := net.JoinHostPort(option.Server, strconv.Itoa(option.Port))
	cipher := option.Cipher
	password := option.Password
	ciph, err := core.PickCipher(cipher, nil, password)
	if err != nil {
		return nil, fmt.Errorf("ss %s initialize error: %w", addr, err)
	}

	// var v2rayOption *v2rayObfs.Option
	var obfsOption *simpleObfsOption
	obfsMode := ""

	decoder := structure.NewDecoder(structure.Option{TagName: "obfs", WeaklyTypedInput: true})
	if option.Plugin == "obfs" {
		opts := simpleObfsOption{Host: "bing.com"}
		if err := decoder.Decode(option.PluginOpts, &opts); err != nil {
			return nil, fmt.Errorf("ss %s initialize obfs error: %w", addr, err)
		}

		if opts.Mode != "tls" && opts.Mode != "http" {
			return nil, fmt.Errorf("ss %s obfs mode error: %s", addr, opts.Mode)
		}
		obfsMode = opts.Mode
		obfsOption = &opts
	} else if option.Plugin == "v2ray-plugin" {
		opts := v2rayObfsOption{Host: "bing.com", Mux: true}
		if err := decoder.Decode(option.PluginOpts, &opts); err != nil {
			return nil, fmt.Errorf("ss %s initialize v2ray-plugin error: %w", addr, err)
		}

		if opts.Mode != "websocket" {
			return nil, fmt.Errorf("ss %s obfs mode error: %s", addr, opts.Mode)
		}
		obfsMode = opts.Mode
		// v2rayOption = &v2rayObfs.Option{
		// 	Host:    opts.Host,
		// 	Path:    opts.Path,
		// 	Headers: opts.Headers,
		// 	Mux:     opts.Mux,
		// }

		// if opts.TLS {
		// 	v2rayOption.TLS = true
		// 	v2rayOption.SkipCertVerify = opts.SkipCertVerify
		// 	v2rayOption.SessionCache = getClientSessionCache()
		// }
	}

	return &ShadowSocks{
		Base: &Base{
			name: option.Name,
			addr: addr,
			udp:  option.UDP,
		},
		cipher: ciph,

		obfsMode: obfsMode,
		// v2rayOption: v2rayOption,
		obfsOption: obfsOption,
	}, nil
}

type ssPacketConn struct {
	net.PacketConn
	rAddr net.Addr
}

func (spc *ssPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	packet, err := socks5.EncodeUDPPacket(socks5.ParseAddrToSocksAddr(addr), b)
	if err != nil {
		return
	}
	return spc.PacketConn.WriteTo(packet[3:], spc.rAddr)
}

func (spc *ssPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, _, e := spc.PacketConn.ReadFrom(b)
	if e != nil {
		return 0, nil, e
	}

	addr := socks5.SplitAddr(b[:n])
	if addr == nil {
		return 0, nil, errors.New("parse addr error")
	}

	udpAddr := addr.UDPAddr()
	if udpAddr == nil {
		return 0, nil, errors.New("parse addr error")
	}

	copy(b, b[len(addr):])
	return n - len(addr), udpAddr, e
}
