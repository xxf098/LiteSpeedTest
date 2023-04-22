package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/log"
	"github.com/xxf098/lite-proxy/stats"
	"github.com/xxf098/lite-proxy/transport/dialer"
	"github.com/xxf098/lite-proxy/transport/resolver"
	"github.com/xxf098/lite-proxy/transport/socks5"
	"github.com/xxf098/lite-proxy/transport/vmess"
	"github.com/xxf098/lite-proxy/utils"
)

type Vmess struct {
	*Base
	client *vmess.Client
	option *VmessOption
}

type VmessOption struct {
	Name           string            `proxy:"name,omitempty"`
	Server         string            `proxy:"server"`
	Port           uint16            `proxy:"port"`
	UUID           string            `proxy:"uuid,omitempty"`
	Password       string            `proxy:"password,omitempty"`
	AlterID        int               `proxy:"alterId,omitempty"`
	Cipher         string            `proxy:"cipher,omitempty"`
	TLS            bool              `proxy:"tls,omitempty"`
	UDP            bool              `proxy:"udp,omitempty"`
	Network        string            `proxy:"network,omitempty"`
	HTTPOpts       HTTPOptions       `proxy:"http-opts,omitempty"`
	HTTP2Opts      HTTP2Options      `proxy:"h2-opts,omitempty"`
	WSPath         string            `proxy:"ws-path,omitempty"`
	WSHeaders      map[string]string `proxy:"ws-headers,omitempty"`
	SkipCertVerify bool              `proxy:"skip-cert-verify,omitempty"`
	ServerName     string            `proxy:"servername,omitempty"`
	Type           string            `proxy:"type,omitempty"`
	WSOpts         WSOptions         `proxy:"ws-opts,omitempty"`
}

type HTTPOptions struct {
	Method  string              `proxy:"method,omitempty"`
	Path    []string            `proxy:"path,omitempty"`
	Headers map[string][]string `proxy:"headers,omitempty"`
}

type HTTP2Options struct {
	Host []string `proxy:"host,omitempty"`
	Path string   `proxy:"path,omitempty"`
}

type GrpcOptions struct {
	GrpcServiceName string `proxy:"grpc-service-name,omitempty"`
}

type WSOptions struct {
	Path                string            `proxy:"path,omitempty"`
	Headers             map[string]string `proxy:"headers,omitempty"`
	MaxEarlyData        int               `proxy:"max-early-data,omitempty"`
	EarlyDataHeaderName string            `proxy:"early-data-header-name,omitempty"`
}

// https://github.com/Dreamacro/clash/blob/412b44a98185b2a61500628835afcbd2c115b00e/adapter/outbound/vmess.go#L75
func (v *Vmess) StreamConn(c net.Conn, metadata *C.Metadata) (net.Conn, error) {
	var err error
	switch v.option.Network {
	case "ws":
		host, port, _ := net.SplitHostPort(v.addr)
		wsOpts := &vmess.WebsocketConfig{
			Host: host,
			Port: port,
			Path: v.option.WSPath,
		}

		if len(v.option.WSHeaders) != 0 {
			header := http.Header{}
			for key, value := range v.option.WSHeaders {
				header.Add(key, value)
			}
			wsOpts.Headers = header
		}

		if v.option.TLS {
			wsOpts.TLS = true
			wsOpts.SessionCache = getClientSessionCache()
			wsOpts.SkipCertVerify = v.option.SkipCertVerify
			wsOpts.ServerName = v.option.ServerName
		}
		c, err = vmess.StreamWebsocketConn(c, wsOpts)
	case "http":
		// readability first, so just copy default TLS logic
		if v.option.TLS {
			host, _, _ := net.SplitHostPort(v.addr)
			tlsOpts := &vmess.TLSConfig{
				Host:           host,
				SkipCertVerify: v.option.SkipCertVerify,
				SessionCache:   getClientSessionCache(),
			}

			if v.option.ServerName != "" {
				tlsOpts.Host = v.option.ServerName
			}

			c, err = vmess.StreamTLSConn(c, tlsOpts)
			if err != nil {
				return nil, err
			}
		}

		host, _, _ := net.SplitHostPort(v.addr)
		httpOpts := &vmess.HTTPConfig{
			Host:    host,
			Method:  v.option.HTTPOpts.Method,
			Path:    v.option.HTTPOpts.Path,
			Headers: v.option.HTTPOpts.Headers,
		}

		c = vmess.StreamHTTPConn(c, httpOpts)
	case "h2":
		host, _, _ := net.SplitHostPort(v.addr)
		tlsOpts := vmess.TLSConfig{
			Host:           host,
			SkipCertVerify: v.option.SkipCertVerify,
			SessionCache:   getClientSessionCache(),
			NextProtos:     []string{"h2"},
		}

		if v.option.ServerName != "" {
			tlsOpts.Host = v.option.ServerName
		}

		c, err = vmess.StreamTLSConn(c, &tlsOpts)
		if err != nil {
			return nil, err
		}

		h2Opts := &vmess.H2Config{
			Hosts: v.option.HTTP2Opts.Host,
			Path:  v.option.HTTP2Opts.Path,
		}

		c, err = vmess.StreamH2Conn(c, h2Opts)
	default:
		// handle TLS
		if v.option.TLS {
			host, _, _ := net.SplitHostPort(v.addr)
			tlsOpts := &vmess.TLSConfig{
				Host:           host,
				SkipCertVerify: v.option.SkipCertVerify,
				SessionCache:   getClientSessionCache(),
			}

			if v.option.ServerName != "" {
				tlsOpts.Host = v.option.ServerName
			}

			c, err = vmess.StreamTLSConn(c, tlsOpts)
		}
	}

	if err != nil {
		return nil, err
	}

	return v.client.StreamConn(c, parseVmessAddr(metadata))
}

func (v *Vmess) DialContext(ctx context.Context, metadata *C.Metadata) (net.Conn, error) {
	log.I("start dial from", v.addr, "to", metadata.RemoteAddress())
	c, err := dialer.DialContext(ctx, "tcp", v.addr)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %s", v.addr, err.Error())
	}
	tcpKeepAlive(c)
	if metadata.Type == C.TEST {
		if tcpconn, ok := c.(*net.TCPConn); ok {
			tcpconn.SetLinger(0)
		}
	}

	log.I("start StreamConn from", v.addr, "to", metadata.RemoteAddress())
	sc := stats.NewConn(c)
	return v.StreamConn(sc, metadata)
}

func (v *Vmess) DialUDP(metadata *C.Metadata) (net.PacketConn, error) {
	// vmess use stream-oriented udp, so clash needs a net.UDPAddr
	if !metadata.Resolved() {
		ip, err := resolver.ResolveIP(metadata.Host)
		if err != nil {
			return nil, errors.New("can't resolve ip")
		}
		metadata.DstIP = ip
	}

	ctx, cancel := context.WithTimeout(context.Background(), tcpTimeout)
	defer cancel()
	c, err := dialer.DialContext(ctx, "tcp", v.addr)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %s", v.addr, err.Error())
	}
	tcpKeepAlive(c)
	sc := stats.NewConn(c)
	c, err = v.StreamConn(sc, metadata)
	if err != nil {
		return nil, fmt.Errorf("new vmess client error: %v", err)
	}
	return &vmessPacketConn{Conn: c, rAddr: metadata.UDPAddr()}, nil
}

func (v *Vmess) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type": "Trojan",
	})
}

func NewVmess(option *VmessOption) (*Vmess, error) {
	security := strings.ToLower(option.Cipher)
	client, err := vmess.NewClient(vmess.Config{
		UUID:     option.UUID,
		AlterID:  uint16(option.AlterID),
		Security: security,
		HostName: option.Server,
		Port:     option.Port,
		IsAead:   option.AlterID == 0, // VMess AEAD will be used when alterId is 0
	})
	if err != nil {
		return nil, err
	}
	if option.Network == "h2" && !option.TLS {
		return nil, fmt.Errorf("TLS must be true with h2 network")
	}

	return &Vmess{
		Base: &Base{
			name: option.Name,
			addr: net.JoinHostPort(option.Server, utils.U16toa(option.Port)),
			udp:  option.UDP,
		},
		client: client,
		option: option,
	}, nil
}

func parseVmessAddr(metadata *C.Metadata) *vmess.DstAddr {
	var addrType byte
	var addr []byte
	switch metadata.AddrType() {
	case socks5.AtypIPv4:
		addrType = byte(vmess.AtypIPv4)
		addr = make([]byte, net.IPv4len)
		copy(addr[:], metadata.DstIP.To4())
	case socks5.AtypIPv6:
		addrType = byte(vmess.AtypIPv6)
		addr = make([]byte, net.IPv6len)
		copy(addr[:], metadata.DstIP.To16())
	case socks5.AtypDomainName:
		addrType = byte(vmess.AtypDomainName)
		addr = make([]byte, len(metadata.Host)+1)
		addr[0] = byte(len(metadata.Host))
		copy(addr[1:], []byte(metadata.Host))
	}

	port, _ := strconv.ParseUint(metadata.DstPort, 10, 16)
	return &vmess.DstAddr{
		UDP:      metadata.NetWork == C.UDP,
		AddrType: addrType,
		Addr:     addr,
		Port:     uint(port),
	}
}

type vmessPacketConn struct {
	net.Conn
	rAddr net.Addr
}

func (uc *vmessPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	return uc.Conn.Write(b)
}

func (uc *vmessPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, err := uc.Conn.Read(b)
	return n, uc.rAddr, err
}
